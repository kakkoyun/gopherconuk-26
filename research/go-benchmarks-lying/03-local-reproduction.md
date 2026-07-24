# Local Reproduction: Trusting a Benchmark on Your Own Machine

## The Local-First Principle

Before CI, before nightly suites, before pinned runners: can you trust the number on your own machine? If not, CI will scale the noise, not the signal.

The local workflow is where a developer spends most of their time. A discipline that only works on a dedicated bare-metal CI server is a discipline most developers skip. This section documents every tool and technique for reducing benchmark noise on a developer machine, with honest assessment of what works on macOS vs Linux.

---

## Platform Reality Check (Read This First)

**Linux** — full environment control: SMT via sysfs, CPU frequency governor, core isolation (`isolcpus`), real-time scheduling (`SCHED_FIFO`), `taskset` for affinity, `perflock` for frequency locking. Docker containers on Linux run directly on the host kernel with no virtualisation overhead.

**macOS** — significant limits:
- No sysfs: SMT and Turbo Boost cannot be controlled from user space.
- Docker Desktop runs containers inside a Linux VM (Apple Hypervisor / Virtualization.framework). `--cpuset-cpus` pins vCPUs *inside the VM*, not physical host cores. VM layer adds scheduling jitter.
- No native `taskset`. `cpus` Mach ports are not user-accessible.
- `SCHED_FIFO` requires elevated privileges and special entitlements.

**The honest Mac workflow**: use `perflock`, run high `-count`, accept a higher baseline noise floor, and keep publication-quality numbers for Linux.

---

## 1. Container Isolation (Docker)

### What it does

`--cpuset-cpus` pins container processes to specific CPUs, preventing scheduler migration. `--cpus` and `--memory` cap resource usage from co-running workloads.

### Linux usage

```bash
docker run --rm \
  --cpuset-cpus=0 \
  --cpus=1 \
  --memory=512m \
  --memory-swap=512m \
  -v $(pwd):/app -w /app \
  golang:1.26 \
  go test -bench=. -benchmem -count=10 -benchtime=2s ./...
```

**Verified Docker flags** (Docker resource constraints docs [S-docker]):
- `--cpuset-cpus=0` — restricts to CPU core 0.
- `--cpus=1.0` — CFS quota equivalent to 1 full CPU.
- `--memory=512m` — hard memory cap.
- `--memory-swap=512m` — same as `--memory` = no swap.

### macOS caveat

On macOS, Docker Desktop runs a Linux VM. `--cpuset-cpus` pins vCPUs *inside the VM's virtual CPU set*, not physical host cores. The VM scheduler may migrate the pinned vCPU between physical cores. The VM layer adds its own jitter.

**Practical impact**: containers on macOS reduce noise from co-running containers, but do NOT provide the isolation of native Linux. CV improvement is real but modest.

**Quantified**: FOSDEM 2026 bare-metal Linux with SMT/DFS disabled reaches CV ~0.05% [S8]. A Mac container is unlikely to reach below CV 1–2% due to the VM layer.

### Dropped caches (IO benchmarks)

```bash
# Linux only — requires root
sudo sh -c 'sync && echo 3 > /proc/sys/vm/drop_caches'
```

---

## 2. CPU Affinity (taskset — Linux only)

The OS scheduler migrates processes between cores to balance load. Each migration evicts warm cache lines and adds jitter.

```bash
# Pin go test to core 0
taskset -c 0 go test -bench=. -count=10 -benchtime=2s ./...

# Pin to cores 0-3
taskset -c 0-3 go test -bench=. -count=10 -benchtime=2s ./...
```

Zero cost to apply. Use on Linux whenever running `-count ≥ 10`.

**macOS**: no native equivalent.

---

## 3. Core Isolation (isolcpus / cset shield — Linux only)

More aggressive than affinity: hand specific cores exclusively to the benchmark, removing them from the general scheduler.

### isolcpus (boot-time)

Add to kernel command line:
```
isolcpus=2,3 nohz_full=2,3 rcu_nocbs=2,3
```

After reboot, pin benchmarks to those cores:
```bash
taskset -c 2 go test -bench=. -count=20 -benchtime=5s ./...
```

Appropriate for a dedicated CI runner; requires a reboot.

### cset shield (runtime, no reboot)

```bash
sudo apt-get install cpuset
sudo cset shield --cpu=2,3 --kthread=on
sudo cset shield --exec -- go test -bench=. -count=10 ./...
sudo cset shield --reset
```

`--kthread=on` is critical — without it, kernel threads still run on shielded cores.

---

## 4. Process Priority (nice, chrt — Linux)

```bash
# Higher scheduling priority (default is 0; range is -20 to 19)
nice -n -5 go test -bench=. -count=10 ./...

# SCHED_FIFO real-time (requires root — dedicated machines only)
sudo chrt -f 50 go test -bench=. -count=10 ./...
```

**SCHED_FIFO risk**: a runaway SCHED_FIFO process starves all normal processes including the display server. Only use on dedicated machines or inside a resource-capped container.

---

## 5. CPU Frequency Control

### perflock (recommended — Linux primary, macOS limited)

`perflock` [S14] is a CPU frequency lock daemon by Austin Clements (Go runtime team). It locks the CPU to a stable frequency for the duration of the benchmark run, then releases it.

```bash
go install github.com/aclements/perflock@latest
perflock go test -bench=. -count=10 -benchtime=2s ./...
```

**How it works**: sets the CPU governor to `performance`, pins max/min frequency to base, and disables Turbo Boost via the `no_turbo` sysfs knob on Intel. AMD uses the `boost` knob.

**macOS**: `perflock` installs and runs, but macOS does not expose frequency control via sysfs. Effect on macOS is limited to what the OS will allow. [UNVERIFIED: exact macOS behaviour — check github.com/aclements/perflock README]

**Recommendation**: use `perflock` by default for all Go benchmarking on Linux. Single highest-value local tool after `benchstat`.

### Manual sysfs controls (Linux, requires root)

```bash
# Performance governor (disables dynamic frequency scaling)
echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

# Disable Intel Turbo Boost
echo 1 | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo

# Disable AMD boost
echo 0 | sudo tee /sys/devices/system/cpu/cpufreq/boost

# Disable SMT — CAUTION: halves logical CPU count, degrades interactive perf
echo off | sudo tee /sys/devices/system/cpu/smt/control
# Re-enable after benchmarking
echo on | sudo tee /sys/devices/system/cpu/smt/control
```

Verified variance data with SMT and DFS controls [S8] — see `04-ci-continuous.md` §3 for the full tables.

---

## 6. Thermal Steady-State

CPUs boost when cool, throttle when hot. A benchmark that starts cool and ends hot shows a performance *decrease* unrelated to the code.

**Symptoms**: first few `-count` runs faster than the last few in the raw output.

**Fix**:
1. Run the benchmark once (throwaway) to warm the CPU before capturing.
2. Wait 30–60 seconds between major test runs.
3. Use `perflock` or DFS disable to prevent frequency variation regardless of thermals.

Laptops are particularly affected: ambient temperature and prior workload both matter. This is a real effect.

---

## 7. The Go-Native Local A/B Loop

### Manual workflow

```bash
# Baseline
go test -bench=BenchmarkMyFunc -benchmem -count=10 -benchtime=2s . > old.txt

# Apply change
go test -bench=BenchmarkMyFunc -benchmem -count=10 -benchtime=2s . > new.txt

# Compare
benchstat old.txt new.txt
```

### Automated with benchdiff

`benchdiff` [S15] automates the stash/checkout/run/compare cycle:

```bash
go install github.com/willabides/benchdiff/cmd/benchdiff@latest

# Compare working tree against main
benchdiff --base=main --benchmem --count=10 --benchtime=2s ./...
```

`benchdiff` handles the git operations, runs both sides, pipes to `benchstat`. Recommended local inner loop for Go performance work.

### go test flag reference

| Flag | Effect |
|------|--------|
| `-bench=<regexp>` | Run only matching benchmarks |
| `-benchmem` | Report B/op and allocs/op |
| `-count=N` | Run each benchmark N times |
| `-benchtime=Xs` | Run for X seconds per benchmark |
| `-benchtime=Nx` | Run exactly N iterations |
| `-cpu=1,2,4` | Run with each GOMAXPROCS value |
| `-run=^$` | Skip tests, run only benchmarks |

**Fixed-iteration vs time-based**: use `-benchtime=100x -count=20` for the most reproducible numbers. Time-based calibration introduces inter-run variance from b.N calibration differences.

### Toolchain pinning

Different Go versions produce different benchmark numbers due to compiler improvements. Since Go 1.21, the `go` directive in `go.mod` IS the toolchain pin:

```
go 1.26.5
```

A separate `toolchain` directive is only needed to allow a *newer* toolchain than the minimum.

**PGO as a variance source**: if `default.pgo` exists in the module root, every build uses it. Use `-pgo=off` when benchmarking without an explicit profile:

```bash
go test -bench=. -pgo=off -count=10 .
```

---

## 8. Cheap Background-Noise Wins

Apply before any serious benchmark session:

- **Close browser, IDEs, Slack** — reduces CPU/memory contention.
- **Disable Spotlight indexing** (macOS): `sudo mdutil -a -i off` (undo: `on`).
- **Airplane mode** — eliminates NIC interrupt coalescing and background network traffic.
- **Disable ASLR** (Linux, temporary): `echo 0 | sudo tee /proc/sys/kernel/randomize_va_space` — more reproducible memory layout. Undo with `echo 2`. Never leave disabled permanently.
- **Warm the page cache** (IO benchmarks): run once before capturing.
- **Drop the page cache** (cold-start IO benchmarks): `sudo sh -c 'echo 3 > /proc/sys/vm/drop_caches'` before each run.

---

## 9. The Pragmatic macOS Workflow

On macOS without access to a Linux machine:

1. **Install `perflock`** — prefix every benchmark run with it.
2. **Use Docker `--cpuset-cpus=0 --cpus=1`** — reduces co-running container noise; acknowledge the VM ceiling.
3. **Use `-count=20 -benchtime=2s`** — more samples partially compensate for higher noise.
4. **Watch CV in benchstat output** (`±` column). CV > 5% = unreliable; don't act on the result.
5. **Use `benchdiff`** for automated A/B — ensures same environment for both sides.
6. **Reserve publication-quality numbers for Linux** — CI runner with pinned bare metal, or a dedicated Linux machine.

---

## Summary Decision Table

| Technique | Platform | Setup cost | Noise reduction | Use |
|-----------|----------|------------|-----------------|-----|
| `-count=10+` | All | None | High (more data) | Always |
| `benchstat` A/B | All | One-time install | N/A (analysis) | Always |
| `benchdiff` | All | One-time install | N/A (automation) | Strongly recommended |
| Docker `--cpuset-cpus` | Linux: real / Mac: limited | Low | Medium (Linux), Low (Mac) | Linux yes; Mac marginal |
| `taskset -c` | Linux | None | Medium | Yes (Linux) |
| `perflock` | Linux: full / Mac: limited | One-time install | High (Linux) | Yes |
| `nice -n -5` | Linux | None | Low–medium | Yes (Linux) |
| DFS/Turbo disable | Linux sysfs | Low (root) | High | Dedicated machine |
| SMT disable | Linux sysfs | Low (root) | Very high (~100×) | CI/dedicated only |
| `cset shield` | Linux | Medium (root) | Very high | Dedicated machine |
| `chrt -f` (SCHED_FIFO) | Linux | Low (root) | Very high | Dedicated machine/CI |

---

## Key Takeaways

1. **`perflock` + `benchdiff` + `benchstat`** is the local Go benchmarking trinity. Install once, use always.
2. **macOS containers ≠ bare-metal Linux** — state this in your results when benchmarked on Mac.
3. **CV is the diagnostic**: `±` > 5% in benchstat means fix the environment before interpreting.
4. **Fixed-iteration benchtime** (`-benchtime=100x`) beats time-based for reproducibility.
5. **Cheap wins compound**: close apps + `taskset` + `perflock` takes 5 minutes and can halve CV.

---

## Sources

[S8] FOSDEM 2026 experiments: https://github.com/igoragoli/fosdem-2026-software-performance — accessed 2026-07-22
[S10] Bakhvalov perf-book: https://github.com/dendibakh/perf-book — accessed 2026-07-22
[S14] perflock: https://github.com/aclements/perflock — accessed 2026-07-22
[S15] benchdiff: https://github.com/willabides/benchdiff — accessed 2026-07-22
[S-docker] Docker resource constraints: https://docs.docker.com/engine/containers/resource_constraints/ — accessed 2026-07-22
Linux cpufreq: https://www.kernel.org/doc/html/latest/admin-guide/pm/cpufreq.html — accessed 2026-07-22
Linux isolcpus: https://www.kernel.org/doc/html/latest/admin-guide/kernel-parameters.html — accessed 2026-07-22

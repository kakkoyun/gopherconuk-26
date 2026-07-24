# CI Continuous Benchmarking for Go

## Introduction

Local benchmarks answer: "Can I trust this number on my machine?" CI benchmarks answer: "Did this change regress performance for everyone?" These are different questions requiring different infrastructure. The mistake is conflating them — running local benchmarks in CI without addressing the environment, or skipping local verification and relying only on CI.

The order matters: local first, then CI. A benchmark you can't trust locally will produce noise at scale in CI.

---

## 1. Why Shared CI Runners Lie

GitHub Actions `ubuntu-latest` and similar shared runners are multi-tenant: your benchmark job runs on a physical host shared with other tenants' workloads. Three sources of noise compound:

**Competing workloads** — another tenant's CPU-bound process runs on a sibling core, contending for shared execution units, memory bandwidth, and last-level cache. Your benchmark sees a different effective CPU speed each run.

**Variable CPU frequency** — shared cloud VMs cannot disable dynamic frequency scaling (DFS/Turbo Boost) at the hypervisor level. The CPU clock adjusts based on thermal load, which varies with the whole-host workload mix. A benchmark that ran at 3.5 GHz in one CI run may run at 3.1 GHz in the next.

**Non-dedicated cache** — last-level cache (LLC) is physically shared across cores on the same die. Cache pollution from neighboring processes introduces unpredictable latency in memory-bound benchmarks.

**The result**: a 10% regression can vanish into shared-runner noise. A 10% speedup can appear where none exists. Directional errors — CI reporting a regression that is actually a speedup — happen in practice (see `07-war-stories.md`, Story 1).

**Practical guidance**: treat shared-runner benchmark numbers as indicative only. For regression gates, use dedicated pinned runners (see §6).

---

## 2. Two CI Patterns

### Pattern A: PR Benchmark Gate (fast, strict)

Runs on every PR. Goal: catch clear regressions before merge.

**Constraints:**
- Must complete in ≤5 minutes on a pinned runner to fit PR feedback loops
- Run a curated subset of benchmarks — the hot paths that have regressed before, or that the PR author touched
- Use `-count=6 -benchtime=2s` as a baseline (enough for benchstat to compute a meaningful CI)
- Compare against a stored baseline from `main` (stored via `benchsave` or a CI artifact)
- Verdict: `benchstat` delta with 95% CI. Block if Δ > threshold AND p < 0.05

**Threshold strategy**: percentage-only thresholds are dangerous (a 5% threshold blocks 5.1% improvements; passes 4.9% regressions reliably). Pair with `benchstat` confidence interval: block only when the CI does not contain zero. A regression of "Δ +8% ±12%" at p=0.3 should not block; "Δ +8% ±2%" at p=0.001 should.

### Pattern B: Nightly Full Suite (thorough, pinned)

Runs nightly on dedicated bare-metal runners with full environment controls applied. Goal: historical trend tracking and catching slow-moving regressions.

**Configuration:**
- `-count=20 -benchtime=5s` — enough samples for robust statistics
- Full environment controls applied (see §3)
- Results stored with `benchsave`; compared against rolling 30-day window
- Change-point detection (e-divisive, via Apache Otava) to flag regressions automatically
- Alert on Slack/email; do not block PRs directly

---

## 3. Environment Controls at the Runner Level

These controls require **bare-metal access** (see §4). They cannot be applied from inside a shared VM.

### SMT (Simultaneous Multithreading / HyperThreading)

SMT allows two hardware threads to share one physical core's execution units. For CPU-bound benchmarks, this introduces severe contention and variance.

**Verified data** (source: FOSDEM 2026 experiments, `github.com/igoragoli/fosdem-2026-software-performance`, AWS m5.metal, DFS disabled):

| Configuration | Mean | Coefficient of Variation |
|--------------|------|--------------------------|
| SMT enabled, task 1 | 1537.64 ± 367.29 ms | **23.887%** |
| SMT enabled, task 2 | 1536.88 ± 366.84 ms | **23.869%** |
| SMT disabled, task 1 | 737.37 ± 0.32 ms | **0.044%** |
| SMT disabled, task 2 | 737.93 ± 1.74 ms | **0.235%** |

~100× reduction in CV with SMT disabled.

```bash
# Disable SMT (requires root, bare metal Linux)
echo off > /sys/devices/system/cpu/smt/control
```

### Dynamic Frequency Scaling (DFS / Turbo Boost)

**Verified data** (same source, SMT disabled):

| Configuration | Mean | Coefficient of Variation |
|--------------|------|--------------------------|
| DFS on, 1 task | 533.97 ± 2.046 ms | **0.383%** |
| DFS off, 1 task | 738.18 ± 0.306 ms | **0.041%** |

~10× reduction in CV with DFS disabled. Absolute runtime is higher (base frequency) but stable.

```bash
# Pin to base frequency and disable Turbo Boost
echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
echo 1 > /sys/devices/system/cpu/intel_pstate/no_turbo
# For AMD:
echo 0 > /sys/devices/system/cpu/cpufreq/boost
```

### CPU Affinity

```bash
# Pin benchmark process to core 0, preventing OS scheduler migration
taskset -c 0 go test -bench=. -count=20 -benchtime=5s ./...
```

### Full CI step example

```yaml
- name: Apply benchmark environment controls
  run: |
    echo off > /sys/devices/system/cpu/smt/control
    echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
    echo 1 > /sys/devices/system/cpu/intel_pstate/no_turbo

- name: Run benchmarks
  run: |
    taskset -c 0 go test \
      -bench=. -benchmem -count=20 -benchtime=5s \
      ./... | tee bench-new.txt

- name: Compare with baseline
  run: |
    benchstat bench-baseline.txt bench-new.txt
```

---

## 4. Bare Metal vs VM

Environment controls (SMT, DFS) require writing to sysfs, which is only possible with root access on a machine where those controls are physically meaningful. In a VM:

- The hypervisor manages SMT scheduling; you can't disable it from the guest
- CPU frequency is virtualised; writing to cpufreq sysfs in a guest has no effect on the physical CPU
- `echo off > /sys/devices/system/cpu/smt/control` may succeed (no error) but be silently ignored

**Bare metal options:**
- AWS EC2 bare metal instances (`m5.metal`, `m7i.metal`, etc.) — dedicated physical host, no hypervisor
- Hetzner dedicated servers — cost-effective self-hosted bare metal
- On-premise physical machines — full control, zero cloud cost per run

**Cost trade-off**: a bare-metal EC2 `m5.metal` (48 vCPUs, 192 GB RAM) costs ~$4–5/hour on-demand. Running nightly benchmarks for 2 hours/night costs ~$300/month. For most teams, a single persistent self-hosted bare-metal machine is more cost-effective.

---

## 5. GitHub Actions Setup

### Self-hosted runner registration

```yaml
# .github/workflows/benchmark-nightly.yml
name: Nightly Benchmarks
on:
  schedule:
    - cron: '0 2 * * *'  # 2am UTC nightly

jobs:
  benchmark:
    runs-on: [self-hosted, bare-metal, linux, x86_64]
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Download baseline
        uses: actions/download-artifact@v4
        with:
          name: bench-baseline
          path: .

      - name: Apply environment controls
        run: |
          sudo sh -c 'echo off > /sys/devices/system/cpu/smt/control'
          sudo sh -c 'echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor'
          sudo sh -c 'echo 1 > /sys/devices/system/cpu/intel_pstate/no_turbo'

      - name: Run benchmarks
        run: |
          taskset -c 0 go test \
            -bench=. -benchmem \
            -count=20 -benchtime=5s \
            ./... | tee bench-new.txt

      - name: Compare
        run: |
          go run golang.org/x/perf/cmd/benchstat@latest \
            bench-baseline.txt bench-new.txt

      - name: Upload new baseline
        uses: actions/upload-artifact@v4
        with:
          name: bench-baseline
          path: bench-new.txt
```

**Why `runs-on: [self-hosted, bare-metal, linux, x86_64]`**: the label combination routes only to registered bare-metal runners. `ubuntu-latest` routes to GitHub-hosted shared VMs and must not be used for performance gates.

---

## 6. The golang.org/x/perf Toolchain

The Go team's official performance measurement toolchain. All tools work with the standard `go test -bench` output format.

**`benchstat`** — statistical comparison of two benchmark result sets. The primary tool for human-readable A/B analysis. Computes geomean, delta, confidence interval, and p-value.

```bash
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

**`benchsave`** — uploads benchmark results to a `perfdata` server for historical storage. Used by the Go team's own performance tracking at `perf.golang.org`.

```bash
go install golang.org/x/perf/cmd/benchsave@latest
benchsave -header "key=value" bench-new.txt
```

**`benchstat` for CI gates** — pipe directly and use exit code:

```bash
# Exit non-zero if any benchmark regresses by >5% with p<0.05
benchstat -filter delta:>+5% old.txt new.txt && echo "PASS" || echo "REGRESSION DETECTED"
```

[UNVERIFIED: the `-filter` flag syntax for benchstat — verify against current docs at `pkg.go.dev/golang.org/x/perf/cmd/benchstat`]

---

## 7. What Carries Over from Local

| Technique | Local | CI |
|-----------|-------|----|
| `go test -bench -benchmem -count=N` | ✅ identical | ✅ identical |
| `benchstat` for A/B comparison | ✅ | ✅ |
| `benchdiff` for git-ref comparison | ✅ | ✅ (with checkout) |
| `perflock` for CPU frequency | ✅ Linux | ✅ bare-metal CI |
| SMT disable | ✅ Linux (sysfs) | ✅ bare-metal CI only |
| DFS/Turbo disable | ✅ Linux (sysfs) | ✅ bare-metal CI only |
| CPU affinity (`taskset`) | ✅ Linux | ✅ bare-metal CI |
| Historical trend tracking | ❌ | ✅ (benchsave + storage) |
| PR gate enforcement | ❌ | ✅ (GitHub Actions) |
| Cross-PR comparison | ❌ | ✅ |

The local workflow builds confidence in a single change. CI provides the historical baseline, the automated gate, and the cross-PR comparison that humans cannot maintain manually.

---

## 8. Existing Tools (Summary)

Full evaluation in `05-existing-tools.md`. Quick summary:

| Use case | Recommended tool |
|----------|-----------------|
| Small OSS project, GitHub | `github-action-benchmark` (free, easy setup) |
| Team with dedicated runner | `bencher.dev` (free self-hosted binary, statistical) |
| Large org, change-point detection | Apache Otava / Nyrkiö (e-divisive algorithm) |
| Local A/B, any project | `benchstat` + `benchdiff` |

**Do not use**: `codespeed` (unmaintained since 2019), `chronologer` (not a Go benchmark tool), `cob` (dangerous `git reset` side effect).

---

## Key Takeaways

1. Shared CI runners cannot be trusted for benchmark regression detection — the noise floor is too high.
2. Bare metal with SMT and DFS disabled reduces CV by ~100× vs a shared runner.
3. Two patterns: a fast PR gate (curated benchmarks, strict threshold) and a thorough nightly suite (full suite, change-point detection).
4. The local workflow and the CI workflow share the same tools (`benchstat`, `benchdiff`, `perflock`); only the runner environment and historical storage differ.
5. Block PRs only when the benchstat confidence interval excludes zero — not on raw percentage thresholds.

---

## Sources

1. FOSDEM 2026 experiments (SMT/DFS variance data): https://github.com/igoragoli/fosdem-2026-software-performance — accessed 2026-07-22
2. golang.org/x/perf benchstat: https://pkg.go.dev/golang.org/x/perf/cmd/benchstat — accessed 2026-07-22
3. GitHub Actions self-hosted runners: https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/about-self-hosted-runners — accessed 2026-07-22
4. MongoDB EC2 benchmark variability post: https://www.mongodb.com/company/blog/engineering/reducing-variability-performance-tests-ec2-setup-key-results — accessed 2026-07-22
5. AWS EC2 bare metal instances: https://aws.amazon.com/ec2/instance-types/ — accessed 2026-07-22
6. Existing tools evaluation: `05-existing-tools.md` in this corpus
7. War stories (directional error in practice): `07-war-stories.md` in this corpus

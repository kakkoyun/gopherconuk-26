---
name: diagnose-noisy-bench
description: |
  Diagnoses the current machine's benchmarking environment and reports active noise sources with remedies.
  Wraps the `benchenv` CLI tool from tools/cli/benchenv/.
  USE WHEN: "diagnose noisy benchmark", "why is my benchmark flaky", "check benchmarking environment",
  "benchmark variance too high", "benchmark results unreliable", "set up for benchmarking".
disable-model-invocation: false
---

# diagnose-noisy-bench

Runs `benchenv` to probe the local machine for benchmark noise sources and reports what to fix.

## Build

```bash
cd tools/cli/benchenv
go build -o benchenv .
```

Or install globally:

```bash
go install github.com/kakkoyun/gopherconuk-26/tools/benchenv@latest
```

## Run

```bash
# Human-readable text output
./benchenv

# Machine-readable JSON (for scripting / CI)
./benchenv -json
```

Exit 0 always (diagnostic tool). Exit 2 on internal error (e.g. JSON encode failure).

## Interpreting the output

Each check reports one of three statuses:

| Status | Meaning |
|---|---|
| `[ok]` | No action needed |
| `[warn]` | Active noise source — follow the remedy shown on that line |
| `[unavailable]` | Probe not supported on this OS/hardware — informational only |

On **macOS**, SMT, Turbo Boost, and the CPU frequency governor are all `unavailable` — the OS does not expose these controls. This is expected. Use a Linux machine or CI runner for publication-quality numbers.

On **Linux**, all five platform checks probe live sysfs/procfs. `unavailable` there usually means a VM guest (hypervisor owns the CPU controls).

## Ordered remediation

Apply in this order — each step compounds on the previous:

1. **Install perflock** (`[warn] perflock not installed`):
   ```bash
   go install github.com/aclements/perflock@latest
   ```
   Prefix every benchmark run with `perflock go test -bench=. ...`.
   This is the single highest-value local tool on Linux. On macOS it installs but has limited effect.

2. **CPU affinity** (Linux, no extra tool needed):
   ```bash
   taskset -c 0 go test -bench=. -count=10 -benchtime=2s ./...
   ```
   Prevents OS scheduler migration between runs. Zero-cost once you know the command.

3. **Disable SMT** (`[warn] SMT control`, Linux bare metal only):
   ```bash
   echo off | sudo tee /sys/devices/system/cpu/smt/control
   # Re-enable after benchmarking
   echo on  | sudo tee /sys/devices/system/cpu/smt/control
   ```
   Verified ~100× CV improvement on CPU-bound benchmarks (see research/04-ci-continuous.md §3).

4. **Set performance governor / disable Turbo Boost** (`[warn] CPU frequency governor`, Linux):
   ```bash
   echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
   echo 1           | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo  # Intel
   echo 0           | sudo tee /sys/devices/system/cpu/cpufreq/boost           # AMD
   ```
   Verified ~10× CV improvement on top of SMT disable (see research/04-ci-continuous.md §3).

5. **Reduce background load** (`[warn] load average`):
   Close browser, IDE, Slack, and any other CPU-hungry processes before capturing numbers.

6. **Install benchstat and benchdiff** if warned:
   ```bash
   go install golang.org/x/perf/cmd/benchstat@latest
   go install github.com/willabides/benchdiff/cmd/benchdiff@latest
   ```
   `benchstat` performs statistical A/B comparison. `benchdiff` automates the git stash/run/compare cycle.
   Together with `perflock`, these form the local benchmarking trinity — see research/03-local-reproduction.md §7.

## macOS pragmatic workflow

On macOS without a Linux machine:

1. Install `perflock` and prefix every run with it.
2. Use `-count=20 -benchtime=2s` — more samples partially compensate for the higher noise floor.
3. Watch the `±` column in `benchstat` output. CV > 5% means the environment is too noisy to act on.
4. Use `benchdiff` for A/B comparisons — ensures both sides run in the same environment.
5. Reserve publication-quality numbers for a Linux CI runner.

## References

- Local reproduction techniques: `research/go-benchmarks-lying/03-local-reproduction.md`
- CI environment controls and variance data: `research/go-benchmarks-lying/04-ci-continuous.md`
- Existing tools evaluation: `research/go-benchmarks-lying/05-existing-tools.md`

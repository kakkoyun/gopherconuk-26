---
name: benchstat-gate
description: |
  Runs Go benchmarks N times, computes coefficient of variation (CV) per benchmark, and emits a PASS/FAIL stability verdict against a CV threshold. Optionally diffs against a saved baseline via benchstat.
  USE WHEN: "benchmark stability gate", "check benchmark coefficient of variation", "gate benchmark regression", "is my benchmark stable enough", "cv threshold check", "benchmark noise floor".
disable-model-invocation: false
---

# benchstat-gate

`benchgate` is a Go CLI that wraps `go test -bench` to measure benchmark stability via the **coefficient of variation** (CV = stddev/mean × 100%). A benchmark with CV > threshold is flagged as unstable and the tool exits non-zero — making it safe to use as a CI gate.

## Build

```bash
cd tools/cli/benchgate
go build -o benchgate .
```

Or install directly:

```bash
go install github.com/kakkoyun/gopherconuk-26/tools/benchgate@latest
```

## Stability check (single run)

```bash
./benchgate \
  -pkg ./... \
  -bench 'BenchmarkFoo|BenchmarkBar' \
  -count 10 \
  -benchtime 1s \
  -cv-threshold 5.0
```

Exit code: `0` = PASS, `1` = FAIL (CV exceeded), `2` = error.

### JSON output

```bash
./benchgate -pkg ./... -count 10 -json | jq .
```

Output shape:

```json
{
  "verdict": "PASS",
  "threshold": 5.0,
  "benchmarks": [
    { "name": "BenchmarkFoo", "mean": 123.4, "stddev": 1.8, "cv": 1.46, "pass": true }
  ]
}
```

## A/B comparison with benchstat

Save a baseline, make your change, then compare:

```bash
# Before the change — save baseline
./benchgate -pkg ./... -count 10 -save /tmp/before.txt

# After the change — gate + diff
./benchgate -pkg ./... -count 10 -baseline /tmp/before.txt
```

The tool runs `benchstat <baseline> <newrun>` and prints the delta table below the CV verdict.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-pkg` | `./...` | Package pattern passed to `go test` |
| `-bench` | `.` | Benchmark regexp |
| `-count` | `10` | Number of runs per benchmark |
| `-benchtime` | `1s` | Per-iteration time budget |
| `-cv-threshold` | `5.0` | Max acceptable CV% (fail above this) |
| `-json` | false | Emit JSON instead of human-readable text |
| `-baseline` | — | Path to saved output for `benchstat` comparison |
| `-save` | — | Path to write raw output (to use as a future baseline) |

## Reading the CV verdict

```
  BenchmarkFoo    mean=   123.4 ns/op  cv=  1.8%  ✓
  BenchmarkBar    mean=    99.1 ns/op  cv=  7.2%  ✗ (exceeds 5.0% threshold)

VERDICT: FAIL — 1/2 benchmarks exceed CV threshold 5.0%
```

- **CV < 2%** — excellent; numbers are trustworthy on most hardware.
- **CV 2–5%** — acceptable; typical for user-space code on a quiet machine.
- **CV > 5%** — noisy. Fix the environment before trusting the numbers.

## CV > 5%? Fix the environment first

High CV means the OS scheduler, frequency scaling, or competing processes are
drowning the signal. Before tuning code, stabilise the measurement:

- **Linux:** pin to an isolated core and disable frequency scaling:
  ```bash
  perflock -governor performance taskset -c 2 go test -bench=. -count=20 ./...
  ```
  `perflock` holds a performance-governor lock for the duration of the run.
  `taskset -c 2` pins to CPU 2 (pick an isolated core, not CPU 0).

- **macOS:** close background apps, disable Spotlight indexing, and prefer
  `-benchtime 5s` to amortise timer jitter over longer runs.

- **CI:** run benchmarks on a dedicated bare-metal runner, not a shared VM.
  Use `benchstat` to detect regressions statistically rather than relying on a
  single noisy run.

Only after CV drops below your threshold should you treat benchmark deltas as
signal rather than noise.

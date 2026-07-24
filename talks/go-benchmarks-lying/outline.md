# Talk Outline: Why Your Go Benchmarks Are Lying (And How to Stop Them)
## GopherCon UK 2026 — 60-min session, advanced audience

---

## Beat sheet

| # | Section | Duration | Type |
|---|---------|----------|------|
| 0 | Cold open | 3 min | Story |
| 1 | The thesis | 2 min | Setup |
| 2 | Layer 1 — Compiler honesty | 15 min | Demo |
| 3 | Layer 2 — Statistical interpretation | 13 min | Demo |
| 4 | Layer 3a — Local reproduction (main) | 14 min | Demo |
| 5 | Layer 3b — CI escalation | 7 min | Slides |
| 6 | War story | 3 min | Story |
| 7 | Agent tooling (close) | 2 min | Demo |
| 8 | Three questions / Q&A setup | 1 min | Close |
| — | Buffer | 60 min total | |

---

## 0. Cold open (3 min)

**The OPERA experiment.**
2006: a team builds a 730 km underground tunnel from CERN to Gran Sasso. Five years, €100M, the most rigorous experimental physics on Earth. In 2011: neutrinos appear to travel faster than light. Months of rechecking. Root cause: a single fibre-optic cable not fully plugged in. 73 ns timing error.

**The Go pivot.**
Your benchmark just reported a 12% speedup. Did you ship a faster binary — or did you measure dead-code elimination? The cable isn't loose. The compiler just removed the loop.

**Plant the three questions** (return to them at the close):
1. Is the compiler measuring real work?
2. Is my sample stable enough?
3. Is the difference large relative to the noise?

---

## 1. The thesis (2 min)

Benchmarking is a measurement problem. Every measurement problem has the same structure: signal + noise. The skill is telling them apart.

Go makes it easy to write a benchmark. It does not make it easy to write one that measures what you think it measures. This talk is a three-layer playbook for benchmarks you can trust — **local first**.

Rule: trust the number on your laptop before you push to CI. If you can't trust it locally, CI will just industrialise the lie.

---

## 2. Layer 1 — Making the compiler honest (15 min)

### 2a. Dead-code elimination (5 min)

**LIVE DEMO 1** — `make bench-dce` in the demo repo.

Show `BenchmarkMakeBuffer_DCE`: allocs/op = 0. The allocation never happened.
Show `BenchmarkMakeBuffer_Correct`: allocs/op = 1. Real work, real measurement.

Key point: ns/op can lie (timer floor). allocs/op cannot lie — the allocation either happened or it didn't. Use `-benchmem` always.

The fix: two-variable sink pattern. Local inside the loop, package-level write after.

### 2b. Constant folding (2 min)

`bits.OnesCount(0b10110)` → compiler evaluates to `3` at compile time. Benchmark measures a constant load.
Fix: route inputs through package-level variables.

### 2c. Inlining (2 min)

After inlining, the compiler sees the inlined body and may eliminate it too. Detect with `go build -gcflags='-m'`.

### 2d. ResetTimer / StopTimer (3 min)

**LIVE DEMO 2** — `make bench-timer`.

Show `BenchmarkProcess_StopOnly_BUG`: ns/op ≈ 0 (timer paused for entire run).
Show `BenchmarkProcess_PerIterSetup_Correct`: real measurement.

The misuse pattern: StopTimer without StartTimer. The fix: always bracket setup with Stop/Start pair.

ResetTimer: use after one-time setup before the loop. Matters most with `-benchtime=1x`.

### 2e. testing.B.Loop — Go 1.24 (3 min)

The new `for b.Loop()` form. What footguns it removes:
- Setup before the loop is excluded automatically (no ResetTimer needed)
- Compiler cannot inline-then-DCE the loop body
- No b.N == 0 edge case

Show side-by-side: `BenchmarkHash_BN_WithSetup_Correct` vs `BenchmarkHash_BLoop`. Same result, simpler code.

> **Layer 1 payoff**: ask question 1. "Is the compiler measuring real work?" Now you know how to answer it.

---

## 3. Layer 2 — Statistical interpretation (13 min)

### 3a. A single number is a lie (2 min)

A benchmark that runs once is a point sample from a distribution. Signal looks like noise and noise looks like signal. Run it 10 times. You get 10 different numbers.

### 3b. benchstat (6 min)

**LIVE DEMO 3** — `make bench-dce` with `-count=10`, pipe to `benchstat`.

Walk through the output table:
- `geomean` — why geometric mean, not arithmetic
- `Δ` — the delta (positive = slower, negative = faster)
- `±` — the confidence interval
- `p` — p-value: is the difference statistically distinguishable from noise?

Key read: a result is actionable when Δ is significant AND the confidence interval does not cross zero.

### 3c. Coefficient of variation (2 min)

CV = σ/μ. High CV means noisy environment. Rule of thumb: CV > 5% → don't trust the result; fix the environment first.

Benchstat doesn't show CV directly but you can compute it from the raw `-count` output.

### 3d. How many runs are enough? (2 min)

`-count=10` is the practical minimum for a meaningful distribution. More is better. Time-based (`-benchtime=5s`) vs fixed-iteration (`-benchtime=100x`) — use fixed-iteration for the most reproducible per-commit numbers.

### 3e. The p-hacking trap (1 min)

"Rerun until you get the number you wanted." Each rerun is a new draw from the distribution. With enough draws, any noise pattern can look like a signal. Set your run count before looking at results, not after.

> **Layer 2 payoff**: ask question 2. "Is my sample stable enough?" You have benchstat. Now use it.

---

## 4. Layer 3a — Local reproduction (14 min)

### 4a. The local-first principle (1 min)

Before CI, before nightly, before a pinned runner — can you trust the number on your own machine? If not, CI will scale the noise.

### 4b. The Docker isolation demo (5 min)

**LIVE DEMO 4** — show the same benchmark run:
- (a) bare on a busy dev laptop: capture CV
- (b) inside a container with `--cpuset-cpus=0 --cpus=1 --memory=512m`: capture CV

Watch CV collapse.

**The macOS honesty caveat** (say this explicitly on stage):
On macOS, Docker Desktop runs containers inside a Linux VM. `--cpuset-cpus` pins vCPUs inside that VM — not physical host cores. You cannot disable host SMT or Turbo Boost from inside. Linux host containers are the real story. If you're on a Mac, containers reduce noise from co-running processes but they're not the full picture. For serious numbers, use a Linux machine or a Linux CI runner.

### 4c. CPU affinity and core isolation (2 min)

`taskset -c 0` — pin the benchmark process to one core, stop the OS scheduler from migrating it mid-run and evicting warm cache lines.

Core isolation (`isolcpus`, `nohz_full`) — hand a core exclusively to the benchmark. Higher setup cost; worth it for nightly suites.

### 4d. CPU frequency control (3 min)

**perflock** (`github.com/aclements/perflock`) — the Go-canonical local CPU frequency daemon. Locks the CPU to its base frequency for the duration of the benchmark, then releases it. One `perflock go test` command.

Why it matters: Turbo Boost adjusts the clock based on thermals and neighbouring-core activity. The same benchmark can run measurably faster if the rest of the machine is idle and the chip is cold.

### 4e. The inner loop: benchdiff (2 min)

`benchdiff` — stash current changes, run benchmarks on HEAD, pop the stash, run again, pipe both runs to `benchstat`. The full A/B comparison in one command.

This is the developer-iteration loop: write change → `benchdiff --base=main` → read benchstat output → decide.

### 4f. Cheap wins (1 min)

Close apps, disable Spotlight/indexer, airplane mode, drop page cache for IO benchmarks. Free 1-2% CV improvement with zero setup.

> **Layer 3a payoff**: you now have a local workflow. It's not perfect, but it's honest.

---

## 5. Layer 3b — CI escalation (7 min)

### 5a. Why shared CI runners lie (2 min)

Shared GitHub Actions runners (`ubuntu-latest`): competing workloads, variable CPU frequency, non-dedicated last-level cache. A 10% regression can vanish into runner noise. A 10% speedup can appear where there is none.

### 5b. Pinned runners + environment controls (3 min)

Bare metal instances (AWS m5.metal or equivalent) — the only way to get SMT and DFS control.

The numbers from our FOSDEM experiments (verified):
- SMT enabled → CV ~23% on CPU-bound benchmarks
- SMT disabled → CV ~0.05% — 100× reduction
- DFS (Turbo Boost) → adds ~10× CV on variable-load scenarios

The commands: three sysfs writes + taskset. Takes 5 seconds in a CI step.

### 5c. Two-pattern CI architecture (2 min)

**PR gate**: run the hot benchmarks (the ones that have regressed before), `-count=6`, benchstat with a strict threshold. Fast enough for every PR.

**Nightly suite**: full benchmark suite, pinned runner, all environment controls applied, `-count=20`, results stored via `benchsave` and compared with benchstat over a rolling window.

Existing tools that do this for you: bencher.dev (hosted), github-action-benchmark (Action), gobenchdata, Apache Otava (change-point detection). Recommendation: see `05-existing-tools.md`.

---

## 6. War story (3 min)

**The CI regression that was a speedup.**

A PR to dd-trace-go (a Go observability tracer) triggered a CI benchmark gate reporting a regression: span creation was slower. But locally, the same benchmark was rock-stable at <0.1% CV. Something didn't add up.

Ran both HEAD and the PR branch on the same machine. Result: the PR was actually 5% *faster* than main. CI had inverted the finding — shared runner noise had flipped the sign on a real performance improvement.

Lesson: a benchmark number from a noisy environment isn't just imprecise. It can be *directionally wrong*. The discipline — local verification with perflock and benchstat before pushing — is what catches the lie before it gates a good change.

---

## 7. Agent tooling (2 min)

The three skills in the `tools/` directory of this repo:

- `honest-benchmark` — audits a Go benchmark file: detects missing sink, wrong StopTimer pattern, suggests B.Loop migration.
- `benchstat-gate` — runs a benchmark N times, computes CV, runs benchstat, gives a pass/fail verdict.
- `diagnose-noisy-bench` — checklist for a flaky benchmark: SMT, DFS, affinity, sample count, compiler flags.

An AI agent can run these on your repo in one afternoon. So can you.

---

## 8. Close — the three questions (1 min)

Before you trust any benchmark result, ask three questions:

1. **Is the compiler measuring real work?** — use the sink pattern, use -benchmem, consider B.Loop.
2. **Is my sample stable enough?** — run benchstat, look at CV, run enough iterations.
3. **Is the difference large relative to the noise?** — read the confidence interval, not just the delta.

If the answer to all three is yes, you have a number you can act on.

---

## Demo repo

All live demos run from: `talks/go-benchmarks-lying/demo/`

```bash
make tools          # install benchstat + benchdiff
make bench-dce      # Layer 1 demo: DCE + constant folding
make bench-timer    # Layer 1 demo: timer misuse
make bench-bloop    # Layer 1 demo: B.Loop vs b.N
make bench          # all benchmarks + benchstat summary
```

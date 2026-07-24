# Statistical Interpretation of Go Benchmarks
## Is the sample stable? Is the difference real?

A benchmark that runs once is a point sample from a distribution. Signal looks like noise and noise looks like signal. Without repetition and analysis, you cannot tell them apart. This section covers the statistical tools available to Go developers — primarily `benchstat` — and the discipline needed to use them correctly.

---

## 1. A Single Number Is a Lie

Consider this `go test` output:

```
BenchmarkProcess-16    1000000    1234 ns/op
```

This tells you nothing about reliability. It is one draw from a distribution whose shape, spread, and stability you do not know. The number could be a local minimum due to a warm cache. It could include a GC pause. It could be 20% above the median.

Run the same benchmark ten times:

```bash
go test -bench=BenchmarkProcess -count=10 . 2>&1 | tee results.txt
```

You now have ten draws. Feed them to `benchstat` to summarise the distribution.

---

## 2. Using benchstat

`benchstat` is the standard Go tool for summarising and comparing benchmark results. It is part of `golang.org/x/perf` [S3].

### Installation

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### Summarising a single run set

```bash
go test -bench=. -benchmem -count=10 -benchtime=2s . > results.txt
benchstat results.txt
```

**Example output:**

```
goos: linux
goarch: amd64
pkg: github.com/example/mylib
cpu: Intel Xeon Platinum 8375C @ 2.90GHz
                │  results.txt   │
                │    sec/op      │
BenchmarkSum      1.234µ ± 2%
BenchmarkProcess  4.567µ ± 8%
```

**Reading the output:**
- `sec/op` — geometric mean of the N runs (geomean, not arithmetic mean)
- `± 2%` — the **coefficient of variation** (CV): standard deviation / mean, as a percentage

### Why geomean, not arithmetic mean?

Benchmark distributions are typically right-skewed: occasional slow outliers (GC pauses, OS scheduler preemptions) pull the arithmetic mean upward. The geometric mean is more robust to multiplicative outliers and is the standard for benchmark comparison tools. `benchstat` uses geomean by default.

### Comparing two runs (A/B comparison)

```bash
go test -bench=. -benchmem -count=10 -benchtime=2s . > old.txt
# make your change
go test -bench=. -benchmem -count=10 -benchtime=2s . > new.txt
benchstat old.txt new.txt
```

**Example output:**

```
                │    old.txt     │              new.txt               │
                │    sec/op      │    sec/op      vs base             │
BenchmarkSum      1.234µ ± 2%    1.187µ ± 1%   -3.81% (p=0.001 n=10)
BenchmarkProcess  4.567µ ± 8%    4.523µ ± 9%   -0.96% (p=0.612 n=10)
```

**Reading the comparison:**
- `vs base` — the delta. Negative = faster.
- `p=0.001` — p-value from Mann-Whitney U test. Low p = difference unlikely to be random noise.
- `p=0.612` — NOT statistically significant. The 0.96% improvement is indistinguishable from noise.
- `n=10` — samples used per side.

**The confidence interval:** the delta whose interval contains zero should not be acted on. If benchstat shows `p > 0.05`, treat the result as no change until you collect more samples.

---

## 3. Coefficient of Variation (CV)

CV = σ / μ (standard deviation divided by mean), expressed as a percentage.

| CV | Interpretation |
|----|---------------|
| < 2% | Excellent. Results are reliable. |
| 2–5% | Acceptable for most purposes. |
| 5–10% | Noisy. Investigate environment before acting. |
| > 10% | Do not trust. Fix the environment first. |

**The verified data** (FOSDEM 2026, `github.com/igoragoli/fosdem-2026-software-performance` [S8]):
- SMT enabled: CV ~23.9% on a CPU-bound benchmark on AWS m5.metal
- SMT disabled: CV ~0.05% — ~100× reduction

CV > 23% means you cannot reliably detect regressions smaller than ~50% with any reasonable sample size. Fix the environment; don't throw more samples at it.

---

## 4. How Many Runs Are Enough?

**Minimum: `-count=10`. For reporting: `-count=20`.**

Kalibera & Jones ("Rigorous Benchmarking in Reasonable Time", ECOOP 2013 [S12]) establish that N ≥ 30 is needed for robust inter-run statistics. In practice:

| `-count` | Use case |
|----------|----------|
| 5 | Quick development check only |
| 10 | Reliable local A/B comparison |
| 20 | CI nightly suite, public reporting |
| 50+ | Very noisy environments |

**`-benchtime` options:**
- `-benchtime=2s` — run each benchmark for 2 seconds (more iterations per run)
- `-benchtime=100x` — exactly 100 iterations; most reproducible across runs

**Fixed-iteration beats time-based for reproducibility.** Time-based (`-benchtime=2s`) calibrates `b.N` differently each run; fixed-iteration (`-benchtime=100x -count=20`) eliminates that source of variance. Prefer fixed-iteration for CI baselines.

---

## 5. The P-Hacking Trap

P-hacking is rerunning a benchmark until you see a result that matches expectations, then reporting that result.

**Why it invalidates results:** each run is an independent draw. With 20 runs and α=0.05, you expect at least one false positive by chance alone. The p-value is calibrated for a *pre-specified* number of experiments, not an unbounded search.

**The discipline:**
1. Decide `-count=N` before running.
2. Run once.
3. Report what benchstat says — even if p > 0.05.

If the result is not significant with your chosen N, either accept "no measurable difference" or pre-commit to a larger N and run again once.

**Recognising it in the wild:**
- "I ran it a few times until it stabilised" ← p-hacking
- "benchstat showed improvement on the third attempt" ← p-hacking
- "I closed Chrome and the numbers got better" ← environmental confounding (related problem)

---

## 6. Effect Size vs Statistical Significance

Statistical significance (low p-value) tells you the difference is unlikely to be zero. It does not tell you the difference is meaningful.

**Example — significant but trivial:**
```
BenchmarkHTTPHandler: Δ +0.3% (p=0.0001 n=100)
```
Statistically significant with 100 samples. Practically: 0.3% is well below production noise.

**Example — not significant but potentially important:**
```
BenchmarkCriticalPath: Δ +15% (p=0.12 n=5)
```
Not significant with 5 samples. But 15% would matter — collect more data before dismissing.

**Always report both:**
1. Statistical significance (p-value, whether CI crosses zero)
2. Effect size (the delta in ns/op or B/op)

**Decision table:**

| Δ | p-value | Action |
|---|---------|--------|
| < 2% | any | No action needed |
| 2–10% | > 0.05 | Collect more samples; likely noise |
| 2–10% | < 0.05 | Investigate; may be real |
| > 10% | > 0.05 | Probably noise; collect more |
| > 10% | < 0.05 | Real — act on it |

---

## 7. Change-Point Detection (Continuous Context)

For PR-level A/B comparisons, benchstat is appropriate. For *continuous* benchmarking across many commits, fixed thresholds fail: slow regressions creep in 1–2% per commit and never trip the gate.

**Change-point detection** (ED-PELT algorithm, Andrey Akinshin [S11]) scans a time series and identifies where the underlying distribution shifts. It handles non-normal distributions, multiple change points, and noisy data — all common in benchmark time series.

Netflix Engineering documented their ED-PELT use case at scale [S13]. Apache Otava (Nyrkiö) implements this as a deployable service [S18] — the recommended tool for this layer (see `05-existing-tools.md`).

**When to add it:** after you have a 30+ commit historical baseline and are seeing too many threshold false positives. Don't start there; start with `benchstat`.

---

## 8. Visualisation: Strip Plots Over Boxplots

Boxplots hide distribution shape. For 30–200 benchmark samples, strip plots (one dot per measurement) are better: outliers are immediately visible, bimodality is obvious, gaps and clusters reveal structure.

`benchstat` does not produce plots, but its raw output pipes to R or Python (matplotlib/seaborn `stripplot`). Brendan Gregg's frequency trails [S24] are the alternative for larger sample sets.

---

## Key Takeaways

1. **Always `-count=10` minimum.** One run is not a result.
2. **Read delta + CI + p-value together** — benchstat shows all three.
3. **CV > 5% means fix the environment**, not add more samples.
4. **Pre-specify run count.** Report the first result.
5. **Effect size matters as much as significance.** Report both.
6. **For continuous benchmarking**, add change-point detection after establishing a historical baseline.

---

## Sources

[S3] benchstat: https://pkg.go.dev/golang.org/x/perf/cmd/benchstat — accessed 2026-07-22
[S8] FOSDEM 2026 experiments: https://github.com/igoragoli/fosdem-2026-software-performance — accessed 2026-07-22
[S11] Andrey Akinshin, ED-PELT: https://aakinshin.net/posts/edpelt/ — accessed 2026-07-22
[S12] Kalibera & Jones, "Rigorous Benchmarking in Reasonable Time" (ECOOP 2013): https://dl.acm.org/doi/10.1145/2509136.2509184
[S13] Netflix Engineering: https://netflixtechblog.com/fixing-performance-regressions-before-they-happen-eab2602b86fe — accessed 2026-07-22
[S24] Brendan Gregg, Frequency Trails: https://www.brendangregg.com/FrequencyTrails/outliers.html — accessed 2026-07-22

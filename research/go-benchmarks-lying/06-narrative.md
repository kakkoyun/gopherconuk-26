# Narrative and Framing Guide

**Talk:** "Why Your Go Benchmarks Are Lying (And How to Stop Them)"
**Venue:** GopherCon UK 2026 — 60 min, advanced audience
**Verification date:** 2026-07-22

---

## The Thesis

Three trust questions form the spine of the entire talk. Return to them at each layer
transition so the audience tracks where they are in the argument:

1. **Does the compiler run the code you wrote?** (Layer 1 — compiler honesty)
2. **Does your measurement reflect reality?** (Layer 2 — statistical validity)
3. **Does the result survive tomorrow, on a different machine, in CI?** (Layer 3 — reproducibility)

The answer to all three is "not by default." The talk earns the right to say that,
then fixes each problem in turn.

---

## Cold Open Analysis

### The OPERA Story — Verified Facts

Source: Wikipedia, "OPERA experiment" (fetched 2026-07-22).

- **What happened:** In September 2011, the OPERA collaboration announced muon neutrinos
  appeared to travel faster than the speed of light over the CERN-to-Gran-Sasso baseline.
  The result attracted global attention because it implied a violation of special relativity.
- **Root cause:** Researchers subsequently traced the anomaly to "a loose fibre optic cable
  connecting a GPS receiver to an electronic card in a computer." Once corrected, a July 2012
  re-measurement showed neutrino speed consistent with the speed of light.
- **Distance (730 km):** Wikipedia's text as fetched does not state the kilometre figure
  explicitly; 730 km is the standard published baseline but must be attributed to the OPERA
  collaboration's papers rather than the Wikipedia article alone. [PARTIALLY VERIFIED — confirm
  against OPERA paper JHEP 10 (2012) 093 before using on stage]
- **73 ns timing error:** Wikipedia does not state this figure. It appears in news reports and
  the OPERA erratum but is [UNVERIFIED from the Wikipedia source]. Confirm against the OPERA
  collaboration's erratum (arXiv:1109.4897v4) before quoting.

### Why This Works as a Cold Open

OPERA is structurally identical to a bad Go benchmark: precise-looking instruments, careful
experimenters, a surprising result, and a systematic measurement error so small it was invisible
until someone asked harder questions. The audience of experienced Go engineers will immediately
understand the analogy. It also preempts defensiveness — if CERN physicists ship a loose cable
for months, no one in the room should feel bad about a flawed benchmark.

### Adaptation for a Go Audience

> "OPERA had 15,000 tonnes of instrumentation and a Nobel-Prize-winning team. Their benchmark
> lied because of a cable the width of a headphone jack. Your `go test -bench` has `testing.B`,
> a laptop, and background Chrome tabs. The cable is your compiler, your OS scheduler, and your
> statistics. Let's find it."

Keep the story to 90 seconds. The physics detail is not the point; the epistemological point is:
systematic measurement error can hide in plain sight, look like signal, and survive peer review.

### Recommendation

Use OPERA as the cold open. Do not spend time on the physics. One slide: the headline number
(apparent FTL result), the fix (loose cable), the lesson (small systematic errors invalidate
precise measurements). Then pivot immediately: "Here are Go's cables."

---

## Talk Arc (with Timing)

### 00:00–02:00 — Cold Open: The Neutrino Benchmark

The OPERA story. One slide. Ends with the three trust questions.

### 02:00–05:00 — Stakes

Why benchmarks matter. Performance as a product quality dimension.

- **Google search latency experiment** (see Canonical References): half-second delay, 20% traffic
  drop. Use this to establish that measurement errors in either direction cost real money.
- Brief: "You cannot optimize what you cannot measure. You cannot trust a measurement you do
  not understand."

### 05:00–22:00 — Layer 1: Compiler Honesty (17 min)

**Core problem:** The compiler is not neutral. It observes your benchmark and removes work.

1. Dead-code elimination — `go test -bench` calls the function, but the compiler may eliminate
   the result if it is unused. Demo: benchmark that computes a sum but throws it away; `-gcflags
   -S` shows the elision.
2. Inlining — the function you think you're benchmarking may not be a call at all. Demo: small
   function inlined into the benchmark loop; compare with `//go:noinline`.
3. Constant folding and loop invariant hoisting — the compiler moves work out of loops. Demo:
   hash of a fixed string computed once, not per iteration.
4. The `sink` and `result` patterns — canonical Go idioms for defeating elision. Live coding demo.

**War story slot (2 min):** Pick one real-world case where a benchmark was 10× wrong because
of dead-code elim. The audience should recognise the shape.

**Transition:** "The compiler is now honest. But the number it gives you may still be a lie —
because a single number cannot describe a distribution."

### 22:00–38:00 — Layer 2: Statistics (16 min)

**Core problem:** Benchmarks produce distributions, not scalars. Reporting the mean is lossy
and often misleading.

1. Why `ns/op` is the mean and nothing else.
2. Latency distributions are almost never normal. Bimodal, long-tailed, and multi-modal
   distributions are common.
3. **Brendan Gregg's frequency trails** (verified — see Canonical References): the six-sigma
   outlier test. 100% of sampled production servers showed disk I/O outliers exceeding 6σ; 98%
   of Node.js servers, 96% of MySQL servers. Outliers are the rule, not the exception.
4. **Gil Tene — "How NOT to Measure Latency"** (verified — see Canonical References): coordinated
   omission and the HDR Histogram. Brief explanation: when a system falls behind, naive
   benchmarks stop issuing requests — hiding the very latency they're trying to measure.
5. Go tooling layer: `benchstat` for statistical comparison; `-count=10` for multiple samples;
   reading the Δ and p-value, not just the percentage.
6. Flame graph of a benchmark showing GC interference as a latency outlier. Tool: `pprof` +
   `go test -bench -memprofile`.

**Demo (3 min):** `benchstat` on two implementations where the means are identical but the
99th percentile diverges 4×. The naive comparison would declare them equivalent. `benchstat`
with sufficient samples catches it.

**War story slot (2 min):** A real case where two benchmark runs showed "improvement" in both
directions due to measurement noise, leading to a commit that was later reverted.

**Transition:** "Your numbers are statistically valid. Now: do they mean the same thing on CI?"

### 38:00–52:00 — Layer 3: Reproducibility — Local Then CI (14 min)

**Core problem:** Environmental noise makes benchmarks non-deterministic across runs and
machines. CI makes it worse by design (shared compute, noisy neighbours, power throttling).

**Local reproducibility first:**

1. CPU frequency scaling — `cpupower` / `powermetrics` on macOS. Benchmarks on battery vs AC.
   Demo: same benchmark, 2× variance, no code change.
2. OS scheduler interference — `GOMAXPROCS=1` for single-core baselines; `taskset` on Linux.
3. Address space layout randomisation (ASLR) — minor but measurable effect on tight loops.
4. `perflock` (Linux) — lock CPU frequency before running benchmarks. Mention `benchlock` if
   a Go-native wrapper exists in the project's tools directory.
5. **Bakhvalov "Performance Analysis and Tuning on Modern CPUs"** (verified — see Canonical
   References): Chapter on measurement methodology for CPU performance. Audience can go deeper.

**CI reproducibility:**

6. Why CI benchmarks are worse: shared runners, bursty neighbours, no frequency lock.
7. The right model: CI detects regressions (statistical significance across N runs), not absolute
   performance. Use `benchstat -delta-test=utest` for Mann-Whitney U.
8. `benchdiff` and the GitHub Actions workflow pattern — run benchmark on PR branch and base,
   compare, comment. Show the workflow YAML skeleton.
9. Dedicated bare-metal or pinned-instance benchmark runners: the cost is low, the signal
   improvement is large.

**Demo (3 min):** Same benchmark run 10 times on a noisy shared CI machine vs 10 times with
`perflock` (or equivalent). `benchstat` side by side. The former gives p=0.3; the latter gives
p=0.001.

### 52:00–58:00 — The Close

See "The Close" section below.

### 58:00–60:00 — Q&A seed

Pre-load one question: "What's the single highest-leverage change for someone starting today?"
Answer: `-count=10` + `benchstat`. Everything else is compounding on top of that.

---

## Canonical References

### 1. OPERA Experiment — Loose Cable Invalidates Precise Measurement

- **Verified fact:** Loose fibre-optic cable connecting GPS receiver to electronic card caused
  apparent faster-than-light neutrino result (2011). Corrected result (2012) consistent with
  speed of light.
- **Source:** Wikipedia, "OPERA experiment," https://en.wikipedia.org/wiki/OPERA_experiment
  (fetched 2026-07-22). For the 73 ns figure, cite OPERA erratum arXiv:1109.4897v4 directly.
- **730 km distance:** [PARTIALLY VERIFIED] — confirm from OPERA paper before stating on stage.
- **73 ns timing error:** [UNVERIFIED from Wikipedia] — confirm from OPERA erratum.
- **How to use:** Cold open only. One slide, 90 seconds. The cable = your measurement
  environment. Do not over-explain the physics.

### 2. Google Search Latency — Half-Second Delay, 20% Traffic Drop

- **Verified fact:** "The page with 10 results took 0.4 seconds to generate. The page with 30
  results took 0.9 seconds. Half a second delay caused a 20% drop in traffic."
- **Source:** Cited on Jeff Atwood's Coding Horror blog, "Performance is a Feature"
  (https://blog.codinghorror.com/performance-is-a-feature/), attributed broadly to Google
  with no named individual. Atwood's post is the secondary citation; the primary is a Google
  internal experiment on search result page sizes, often attributed to Marissa Mayer at Web 2.0
  Conference 2006, but [UNVERIFIED — the Mayer attribution could not be confirmed from primary
  sources]. Cite as "Google search team" rather than a named individual.
- **How to use:** Stakes section (05:00). Establishes that latency measurement errors translate
  to real user and revenue impact. Do not over-claim — say "a Google experiment found" rather
  than naming Mayer.

### 3. Yahoo Performance Research — 400ms, 5–9% Traffic Change

- **Status: [UNVERIFIED]** — Could not access primary source (YUI Blog, Stoyan Stefanov,
  2007) during research. The figure is widely cited but the original blog post at
  `yuiblog.com/blog/2007/01/04/performance-research-part-2/` is defunct. Confirm before
  using on stage, or drop in favour of the verified Google stat.
- **Alternative:** Replace with Amazon's verified "every 100ms of latency cost 1% in sales"
  (widely attributed to Amazon engineering, e.g. Greg Linden's 2006 presentation at the O'Reilly
  Emerging Technology Conference — also requires verification from primary source).

### 4. Tobi Lütke — "All World-Class Software Is Fast"

- **Claimed quote:** "All world-class software is fast. There's a strong correlation between
  software quality and speed." Attributed to @tobi on Twitter/X, October 2018, tweet ID
  1052495816956358657.
- **Status: [UNVERIFIED]** — Twitter/X requires authentication; the tweet could not be fetched.
  The quote is widely cited in performance engineering communities but cannot be confirmed as
  exact wording from primary sources accessed.
- **How to use:** If you can independently verify the tweet, use it as the opening epigraph.
  If not, paraphrase as "Tobi Lütke has argued publicly that performance is a proxy for
  software quality" and cite the tweet ID so audience can verify. Do not quote verbatim without
  confirmation.

### 5. Gil Tene — "How NOT to Measure Latency"

- **Verified:** YouTube video exists at https://www.youtube.com/watch?v=lJ8ydIuPFeU, titled
  "How NOT to Measure Latency," speaker Gil Tene. (Fetched 2026-07-22.)
- **Key concept:** Coordinated omission — when a system falls behind, a naive load generator
  stops issuing requests, hiding latency. The benchmark measures the system at its best, not
  under load. HDR Histogram corrects for this.
- **How to use:** Layer 2 (statistics), ~28:00. The concept maps directly to Go benchmarks
  that pause the timer during setup but hide GC pauses inside the timed section. Spend 2 slides
  on coordinated omission, then show the `b.ResetTimer()` and `b.StopTimer()` patterns.

### 6. Brendan Gregg — Frequency Trails and Outlier Detection

- **Verified:** Page exists at https://www.brendangregg.com/FrequencyTrails/outliers.html
  (fetched 2026-07-22).
- **Key facts from the page:**
  - Outlier test: `maxσ = (max(x) − μ) / σ`. If maxσ > 6, outliers are present.
  - "100% of these servers have latency outliers" (disk I/O, six-sigma threshold).
  - Node.js: 98% of sampled servers had outliers. MySQL: 96%.
  - Mean response time example: ~3 ms mean, but most requests ~1 ms, some outliers to 30 seconds.
- **How to use:** Layer 2 (statistics), ~25:00. One slide with the σ formula and the 100%/98%/96%
  figures. Punch line: "If Brendan Gregg's production fleet is all outliers, your laptop
  benchmark is definitely hiding one."

### 7. Denis Bakhvalov — "Performance Analysis and Tuning on Modern CPUs"

- **Verified:** Book title confirmed, author Denis Bakhvalov et al., repository at
  https://github.com/dendibakh/perf-book (fetched 2026-07-22). CC0 licence.
- **How to use:** Layer 3 (reproducibility), ~45:00. Mention as the definitive reference for
  CPU-level measurement methodology (PMU counters, TMA, frequency scaling). The audience can
  read it for free. One slide: "If you want to go deeper than `go test -bench`, this is the
  book. It's free."

---

## The Close

Three questions, delivered as a checklist the audience can act on today:

```
□ 1. Does my benchmark use the result? (sink pattern, -gcflags -S to verify)
□ 2. Do I run it enough times to have a p-value? (-count=10, benchstat)
□ 3. Is my CI environment controlled enough to trust the signal? (bare metal, perflock, N≥10)
```

Call to action — "Wire it up this afternoon":

> "Here is the minimum viable benchmark discipline. One: add `var Sink any` and assign your
> result to it. Two: run `go test -bench=. -count=10 | tee new.txt` and compare with
> `benchstat old.txt new.txt`. Three: if you want CI, copy the GitHub Actions workflow from
> the talk repo and point it at your package. That's it. All three in under an hour. The
> benchmarks you write after today will tell the truth."

Provide a QR code to the talk repo with starter workflow and `sink` template.

---

## What to Cut If Running Long

In priority order — cut later items first:

1. **ASLR discussion** (~1 min, Layer 3) — interesting but low practical impact for most Go code.
2. **Constant folding demo** (~2 min, Layer 1) — the dead-code elim and inlining demos make
   the same point; constant folding is the third example, not the first.
3. **Mann-Whitney U explanation** (~1.5 min, Layer 2) — say "`benchstat` uses a non-parametric
   test; trust the p-value" and move on. The math is in the docs.
4. **Coordinated omission deep dive** (~2 min, Layer 2) — keep the concept, cut the HDR
   Histogram internals. One slide instead of two.
5. **War story #2** (Layer 2, ~2 min) — keep war story #1 (compiler), cut war story #2
   (statistics). Demos carry more weight for this audience.

If cutting to 45 min: drop Layer 3's CI section entirely and close after the local
reproducibility demo. Add a slide: "CI: same ideas, harder environment — see talk repo README."

---

## Blog Series Mapping

The talk structure maps naturally to five posts, each self-contained:

| Post | Title | Talk section | Angle |
|------|-------|-------------|-------|
| 1 | "The Benchmarks That Measure Themselves Away" | Layer 1 | Dead-code elim, inlining, the sink pattern. Hands-on with `-gcflags -S`. |
| 2 | "Your Benchmark's Mean Is a Lie and the 99th Percentile Is the Truth" | Layer 2 statistics | `benchstat`, distributions, Gregg's outlier test applied to `go test` output. |
| 3 | "What Tene's Coordinated Omission Means for `testing.B`" | Layer 2 coordinated omission | Mapping the Gil Tene concept to `b.StopTimer()` misuse; GC pause hiding. |
| 4 | "Making Go Benchmarks Reproducible on Your Laptop" | Layer 3 local | CPU frequency scaling, `perflock`, `GOMAXPROCS`, practical checklist. |
| 5 | "Benchmark CI That Doesn't Lie" | Layer 3 CI | `benchdiff` workflow, statistical significance gates, bare-metal vs shared runners. |

Post 1 goes out the week before the conference as a teaser. Posts 2–5 publish weekly after.

---

## Sources

1. Wikipedia, "OPERA experiment" — https://en.wikipedia.org/wiki/OPERA_experiment
   (fetched 2026-07-22). Confirms loose fibre-optic cable root cause. 73 ns figure not present;
   confirm from OPERA erratum arXiv:1109.4897v4.

2. Jeff Atwood, "Performance is a Feature," Coding Horror blog —
   https://blog.codinghorror.com/performance-is-a-feature/
   (fetched 2026-07-22). Secondary citation for Google 500ms/20% stat. No named individual
   attributed on the page.

3. Gil Tene, "How NOT to Measure Latency" — https://www.youtube.com/watch?v=lJ8ydIuPFeU
   (fetched 2026-07-22). Title and speaker confirmed.

4. Brendan Gregg, "Frequency Trails: Outliers" —
   https://www.brendangregg.com/FrequencyTrails/outliers.html
   (fetched 2026-07-22). Six-sigma outlier formula and 100%/98%/96% server outlier rates
   confirmed from page content.

5. Denis Bakhvalov et al., "Performance Analysis and Tuning on Modern CPUs" —
   https://github.com/dendibakh/perf-book (fetched 2026-07-22). Title, author, CC0 licence
   confirmed.

6. @tobi (Tobi Lütke), tweet ID 1052495816956358657 — https://x.com/tobi/status/1052495816956358657
   [UNVERIFIED — Twitter/X requires authentication; could not access content].

7. Yahoo YUI Blog, Stoyan Stefanov, "Performance Research, Part 2" (2007) —
   http://yuiblog.com/blog/2007/01/04/performance-research-part-2/ [UNVERIFIED — page defunct;
   archive.org blocked during fetch. Confirm via Internet Archive Wayback Machine manually].

8. OPERA collaboration erratum — arXiv:1109.4897v4 [UNCHECKED — cited as primary source for
   73 ns figure; fetch independently before using the number on stage].

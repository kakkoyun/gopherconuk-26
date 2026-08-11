# Speaker notes: Why Your Go Benchmarks Are Lying

GopherCon UK 2026 · 60 minutes including Q&A · advanced audience

Deck: `slides/presentation.md` (76 slides)

## Timing

This is the additive review version. The scripted estimate is about 66 minutes before Q&A, so it does not yet fit the slot. Kemal will choose the pruning pass after reviewing and rehearsing the complete deck.

| Block | Slides | Expanded target | Running total |
| --- | ---: | ---: | ---: |
| Loose-cable story and thesis | 2-8 | 4:00 | 4:00 |
| Personal and Datadog context | 9-10 | 1:30 | 5:30 |
| Why benchmark | 11-14 | 3:30 | 9:00 |
| Before optimizing | 15-18 | 4:00 | 13:00 |
| Micro vs macro | 19-21 | 3:00 | 16:00 |
| Compiler honesty | 22-37 | 14:00 | 30:00 |
| Statistical interpretation | 38-45 | 10:00 | 40:00 |
| Local reproduction | 46-53 | 11:00 | 51:00 |
| CI escalation | 54-58 | 5:00 | 56:00 |
| War story | 59-65 | 5:00 | 61:00 |
| Tools and close | 66-76 | 5:00 | 66:00 |

Do not prune during this pass. Candidate rehearsal cuts include assembly detail, some Linux-control detail, and one of the two `benchenv` output slides. Never cut the DCE evidence, CV experiment, macOS caveat, war story, three questions, or Datadog rationale without Kemal's review.

## Slides 2-8: the loose-cable story

### Slides 2-4: setup and reveal

Open on the story. Do not introduce yourself first.

Ninety seconds is enough for OPERA. Resist the physics details. The audience needs three beats: a careful instrument, a surprising result, and a connector.

The two-fault detail matters. The fibre connector biased the result in one direction; an oscillator defect pushed the other way and partially masked it. A benchmark environment can contain several systematic errors that cancel until one changes.

The 73 ns figure comes from CERN's 22 February 2012 release and Cartlidge, *Science* 335(6072):1027.

### Slides 5-6: pivot to Go

Say: "The point is not that particle physicists were careless. The point is that a precise system can be wrong in a way that survives expert review."

Then make the scale contrast. OPERA had an international collaboration. We have `testing.B`, a laptop, and background processes.

### Slides 7-8: plant the contract

The three questions are the talk's contract. Pause after each one.

The local-first rule is the first answer: if the number is unstable on your machine, CI can only automate that instability.

## Slides 9-10: personal and Datadog context

### Slide 9: personal context

Twenty seconds. Do not read the slide. Say:

"I work where Go meets observability and performance: Prometheus, OpenTelemetry compile-time instrumentation, and, previously, Parca's eBPF profiler. I gave the first version of this material at FOSDEM."

Do not mention contribution counts.

### Slide 10: why Datadog cares

This is not a generic company slide. Say:

"Datadog SDKs execute inside customer processes. Every allocation and nanosecond we add comes out of someone else's workload budget. That is why continuous benchmarking is product correctness for us."

Use the public PGO result: production CPU dropped by about 3.4%. The later dd-trace-go story shows why benchmark gates still need human scrutiny.

## Slides 11-14: the data-backed reason to benchmark

### Slides 12-13: latency, throughput, and user cost

Move briskly but keep the data. Latency is what one operation costs; throughput is system capacity. An optimization can improve one and damage the other.

The Google result is attributed to the search team: adding 500 ms cost about 20% of traffic. Do not attribute it to Marissa Mayer.

### Slide 14: performance as a feature

Let the audience read Tobi Lütke's quote. The point is not that speed is the only feature. It establishes that reliable performance measurement affects product quality.

## Slides 15-18: choose the right target

### Slide 16: production evidence first

Say: "Use pprof or continuous profiling to find the hot path, and production percentiles to identify the user symptom. Then write a benchmark for that path."

The OpenTelemetry eBPF Profiler and Parca provide whole-system CPU evidence without changing application source. The slide intentionally keeps in-process and out-of-process options.

### Slide 17: define success

Without a target, optimization never ends. Use a concrete SLO, check the error budget, and apply Amdahl's Law before polishing cold code.

### Slide 18: prior work

One sentence: "Daniel Martí's GopherCon 2019 talk explains how to find what to optimize; this talk explains how to trust the measurement once you have a target."

## Slides 19-21: microbenchmarks and macrobenchmarks

Microbenchmarks isolate a mechanism and expose compiler tricks. Macrobenchmarks preserve workload shape but make attribution harder. Regression detection usually needs both.

Scope the rest of the talk: Go's `testing.B` microbenchmarks.

## Slides 22-37: making the compiler honest

### Slides 23-27: dead-code elimination

The compiler is doing its job when it removes unused work. The benchmark author failed to create an observable result.

The captured result on slide 25 is the key evidence:

```text
DCE       0.2532 ns/op   0 B/op   0 allocs/op
correct  11.14 ns/op  64 B/op   1 allocs/op
```

`make([]byte, 64)` has no path that avoids allocation. Zero allocations means the call disappeared.

Land this line slowly: "`ns/op` has a timer floor. `allocs/op` tells us whether the allocation happened."

The two-variable sink avoids a global write on every iteration. Keep a local result in the loop and publish once after it.

### Slides 28-30: constant folding and inlining

Constant inputs can turn work into a literal load. The assembly slide is captured evidence, not a terminal switch. One side contains `MOVD $3`; the other contains the real count instruction.

Inlining helps production code and gives DCE more visibility inside a benchmark. A non-constant input and an observable result are both required.

### Slides 31-34: timer ordering

`ResetTimer` clears elapsed time but does not stop the timer. Use it after one-time setup.

For per-iteration setup, stop before fixture creation and restart before the operation under test. The captured output shows the broken benchmark looking 25% faster because it excludes the function it claims to measure.

Slide 34 preserves the non-terminating case. Without a matching `StartTimer`, timed duration never accumulates, so the framework keeps increasing `b.N`. The short line is enough: "We tried it. It hung."

### Slides 35-37: `B.Loop`

Go 1.24's `b.Loop()` excludes setup before the loop, stops timing after the loop, calls the benchmark function once per `-count`, and prevents DCE when the condition appears literally as `b.Loop()`.

Do not read the table. Point to setup exclusion and DCE prevention.

Close Layer 1 by answering the first question: sink pattern, `-benchmem`, `allocs/op`, and `B.Loop`.

## Slides 38-45: statistical interpretation

### Slide 39: one point is not a distribution

The 43% swing is captured data from two runs of the same binary on a deliberately loaded machine. Say that if asked; it strengthens the point.

### Slides 40-42: benchstat and CV

`11.32n` is the median, not the mean. Benchstat's geomean appears only as a summary across benchmarks.

A `~` result means no measurable difference. That is a result, not an invitation to rerun until a delta appears.

Benchstat compares distributions. It does not tell us whether the machine itself is trustworthy. CV, `sigma / mu`, provides that environmental check.

### Slides 43-45: discipline before results

Ten runs are the floor; twenty are better. Fixed-iteration runs can improve per-commit reproducibility. A significant tiny effect and a useful effect are different questions.

Set the run count before looking at the result. Repeating until the desired p-value appears is p-hacking.

Close Layer 2: enough samples, benchstat, and CV before comparison.

## Slides 46-53: local reproduction

### Slides 47-49: measured isolation

The Docker results are committed captured output, not a live run:

| Condition | Mean ns/op | CV |
| --- | ---: | ---: |
| Idle host | 11.46 | 4.75% |
| Host with 16 spinners | 34.97 | 18.88% |
| Container pinned to vCPU 0 | 16.28 | 5.25% |

The loaded machine was three times slower and four times noisier. Pinning restored the idle noise floor while the host stayed saturated.

Then state the ceiling: 5.25% is useful isolation, not publication-quality control. Bare-metal Linux with SMT disabled reached about 0.05% in the referenced experiments.

### Slide 50: the macOS boundary

Do not soften this. Docker Desktop pins a vCPU inside a Linux VM, not a physical host core. The host can still migrate that vCPU and change frequency. Containers on macOS isolate co-running processes; they do not provide controlled hardware.

### Slides 51-53: Linux controls and the inner loop

`taskset`, isolated cores, scheduling priority, and a locked clock attack different noise sources.

The perflock caveat comes from source inspection. On macOS the mutual-exclusion lock works, but the cpufreq and thermal controls are Linux-only. Do not imply identical control on both platforms.

End with the local workflow: stable baseline, isolated change, same conditions, repeated samples, benchstat comparison.

## Slides 54-58: escalation to CI

### Slides 55-56: why shared runners fail

Shared runners change neighbors, CPU frequency, host model, thermal state, and virtualization overhead between runs.

Use the AWS m5.metal result to show why hardware control matters:

- SMT on with dynamic frequency scaling: 3.74% CV.
- SMT off with dynamic frequency scaling: 2.09% CV.
- SMT off with a fixed 2.5 GHz clock: 0.05% CV.

That is roughly a 75-times noise reduction from controlling the machine.

### Slides 57-58: two CI patterns and existing tools

PR-gate CI compares before and after on the same controlled host. Nightly trending detects gradual drift with a rolling baseline. Mature teams often need both.

Preserve the survey. Bencher, Go Benchmarks, BenchDiff, and Gobenchdata provide useful persistence, visualization, or comparison. The gap addressed by this repository is a small Go-native stack that also records environment metadata and applies a CV gate.

## Slides 59-65: the dd-trace-go war story

### Slides 60-61: apparent regression

Set up the surprise: PR #4891 introduced a targeted allocation optimization, but CI reported that `BenchmarkStartSpan/1` was 1.62% slower with p=0.000.

Do not dismiss the CI result. Inspect it.

### Slides 62-63: sign inversion and mechanism

On the same controlled machine, the branch was 0.38% faster. The signs disagree. A real effect should not reverse merely because the host changed.

The changed code was not on that benchmark's path. The plausible mechanism is binary layout: code movement changed instruction-cache or branch-prediction behavior. The measured difference was small, real on a specific binary, and not evidence that the PR slowed its target path.

### Slides 64-65: known phenomenon and lesson

Keep the citations short. Causal profiling, Stabilizer, and LLVM's Machine Function Splitter establish that code layout can move benchmark results.

Then state the lesson separately: `p=0.000` can coexist with a directionally wrong conclusion. Environment matching, path validation, effect size, and stability still matter.

## Slides 66-71: captured tools and results

Do not switch to a terminal. These are captured outputs from tested repository tools.

### Slide 67: the three tools

- `honestbench` statically analyzes benchmark source for discarded results, sink mistakes, timer ordering, and `b.N` loops that can migrate to `B.Loop`.
- `benchgate` runs benchmarks repeatedly and rejects samples above a configurable CV threshold. It can also hand results to benchstat for baseline comparison.
- `benchenv` diagnoses noise controls and missing benchmark tools.

The tools are small and composable. They do not claim to replace hosted benchmark services.

### Slide 68: `honestbench`

Read the findings, not the whole output. The analyzer catches a discarded `makeBuffer` result and a `StartTimer` placed after the work under test. It reports 17 findings across 12 benchmark functions without running a benchmark.

### Slide 69: `benchgate`

The same benchmark exceeds the 5% CV threshold in one capture and passes an 8% threshold in another. The point is not that 8% is a good default; it is that the policy is explicit and machine-enforced.

### Slides 70-71: `benchenv`

Slide 70 shows the diagnostic categories and remedies. Slide 71 is the captured output from this laptop.

The four unavailable checks are deliberate. macOS does not expose Linux sysfs controls for SMT, frequency governors, or Turbo Boost. The honest output is `unavailable`, not a fabricated zero or a green check.

The repository also includes agent skills that wrap all three CLIs.

## Slides 72-76: close

### Slide 72: the honest answer

Pause on the punchline. Missing hardware controls are information. They tell us when a laptop result is useful for iteration but not strong enough for a publication-quality claim.

### Slide 73: minimum discipline

This is the short workflow the audience can apply today: add a sink, run enough samples, compare with benchstat, and inspect the environment with benchenv.

### Slide 74: return to the three questions

Make the structure explicit:

1. Compiler honesty: sink patterns, `-benchmem`, and `B.Loop`.
2. Sample stability: repeated runs, benchstat, and CV.
3. Effect versus noise: read the interval and control the environment.

### Slides 75-76: CTA and final questions

Point at the QR on the CTA. Name what is there: the deck, captured experiments, three CLIs, and the benchmark skills. Mention the prior FOSDEM recording for anyone who wants the earlier version.

End with: "Trust your numbers only after the compiler, the sample, and the machine have each earned that trust."

Stop. Leave the final questions slide up.

## Delivery constraints

- No live or recorded demos. Every output shown is captured from the repository.
- If a command result is challenged, identify its committed result file and explain the environment; do not improvise a rerun on stage.
- Do not treat Docker Desktop on macOS as hardware isolation.
- Do not claim statistical significance proves practical importance.
- Do not imply that existing CI tools are defective; explain the specific environment and gating gap addressed here.

## Public references for questions

| Claim | Public reference |
| --- | --- |
| OPERA connector fault | <https://home.cern/news/press-release/cern/opera-experiment-reports-anomaly-flight-time-neutrinos-cern-gran-sasso> |
| Google 500 ms result | <https://glinden.blogspot.com/2006/11/marissa-mayer-at-web-20.html> |
| Daniel Martí's talk | <https://www.youtube.com/watch?v=BF3qhVmXflw> |
| Go benchmark implementation | <https://github.com/golang/go/blob/go1.26.1/src/testing/benchmark.go> |
| Go `B.Loop` behavior | <https://pkg.go.dev/testing#B.Loop> |
| Benchstat | <https://pkg.go.dev/golang.org/x/perf/cmd/benchstat> |
| Causal profiling | <https://arxiv.org/abs/1608.03676> |
| Stabilizer | <https://people.cs.umass.edu/~emery/pubs/stabilizer-asplos13.pdf> |
| LLVM Machine Function Splitter | <https://llvm.org/docs/MachineFunctionSplitter.html> |
| Datadog PGO result | <https://www.datadoghq.com/blog/datadog-pgo-go/> |
| Prior FOSDEM version | <https://youtu.be/8211fNI_nc4> |

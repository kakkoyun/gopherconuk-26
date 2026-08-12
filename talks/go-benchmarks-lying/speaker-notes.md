# Speaker notes: Why Your Go Benchmarks Are Lying

GopherCon UK 2026 · 60 minutes including Q&A · advanced audience

Deck: `slides/presentation.md` · 109 pages including progressive-reveal steps

## Corrections applied after dry run 2

This file had drifted from the deck. Three things were wrong and are now fixed.
Recorded here so the same errors do not get reintroduced from an old draft.

| Was | Now | Why |
| --- | --- | --- |
| War story described as `BenchmarkStartSpan/1`, 1.62% slower, p=0.000 | `BenchmarkOTLPProtoSize`, **6-9%** slower | The deck and `07-war-stories.md` agree on OTLPProtoSize. The notes were describing a different benchmark. |
| CV figures quoted as 3.74% / 2.09% / 0.05% | **23.9% → 0.044-0.235%** (SMT), **0.383% → 0.041%** (DFS) | Claims ledger rows 1-3 and `04-ci-continuous.md`. The old numbers matched no source. |
| Runtime estimated at 66 minutes | **~38-40 minutes measured**, now built to ~50 | Dry run 2 wall clock: 10:13 start, slide 60 at 10:44. |

## Timing

Target 50 minutes presented, 10 minutes Q&A.

| Part | Target | Cumulative |
| --- | ---: | ---: |
| 0 · Cold open and contract | 5:00 | 5:00 |
| 1 · Who and why | 4:00 | 9:00 |
| 2 · Before you measure | 5:00 | 14:00 |
| 3 · Arc 1, local and micro | 17:00 | 31:00 |
| 4 · Arc 2, CI and macro | 15:00 | 46:00 |
| 5 · Tools and close | 4:00 | 50:00 |

Checkpoints: arc 1 by 14:00, arc 2 by 31:00, close by 46:00. Cut ladder is in
`outline.md`; if arc 2 has not started by 33:00, start applying it.

## Part 0 — cold open and contract

### OPERA

Open on the story. Do not introduce yourself first. Ninety seconds is enough.
Three beats: a careful instrument, a surprising result, a connector.

The two-fault detail is the point worth landing. The fibre connector biased the
result one way; an oscillator defect pushed the other way and partially masked
it. A benchmark environment can hold several systematic errors that cancel until
one of them changes. Do not attach a number to the oscillator (ledger row 9).

The photograph has an arrow on the connector. Let people look at it for a beat
before speaking.

### Pivot and contract

Say: "The point is not that particle physicists were careless. The point is that
a precise system can be wrong in a way that survives expert review."

Then the scale contrast, then the three questions. Pause after each question.

### Roadmap

This slide is new and it is doing real work. State that the same three questions
get asked twice, at two scales, and that the answers are different each time.
Name what they leave with. Do not read the bullets verbatim.

## Part 1 — who and why

### How I got here

This replaced a credentials list. Deliver it as a story with four reveals, and
let each one land as a consequence rather than an achievement:

- `client_golang` — allocations showed up in somebody's scrape budget
- Parca — a profiler costing 5% CPU is not deployable
- Datadog — our SDKs run inside your process
- the punchline: every one of those jobs punished me for trusting a benchmark

Twenty to thirty seconds. Do not list maintainer titles or contribution counts.

### Why Datadog cares

Not a generic company slide. Say: "Every allocation and nanosecond we add comes
out of someone else's workload budget. That is why continuous benchmarking is
product correctness for us." PGO's 3.4% production CPU reduction is public.

## Part 2 — before you measure

Move briskly. Latency and throughput are a two-step reveal; the payoff line is
that an optimization can improve one and damage the other.

The Google result is attributed to the search team, not to Marissa Mayer.

**Representative and repeatable** is the frame for everything after it. Micro
fails the first, macro fails the second. Say it once, clearly, then let the two
arcs demonstrate it.

**Start macro, then drill** is the concrete answer to "what should I actually
do". Walk the four steps on the left, then the failure mode on the right. This
slide exists because the previous version of the talk raised macrobenchmarks and
never told anyone what to do about them.

The scope statement claiming this was a microbenchmarks-only talk is gone. Do not
reintroduce it.

## Part 3 — arc 1, local and micro

### Arc divider and the plant

The divider restates the three questions at micro scale and carries the
local-first rule: if you cannot trust the number on your own machine, CI only
industrialises the lie.

Then plant #4891. The bot says `BenchmarkOTLPProtoSize` is 6-9% slower than main;
nothing in the diff touches OTLP encoding. Say "hold that thought" and move on.
**Do not resolve it here.** The whole point is that the audience needs the
compiler section before they can solve it.

### §1A — making the compiler honest

The framing slide comes first. Everything in the section exists to stop the
compiler deleting the work you are timing, and all four techniques are test
scaffolding. Say that once at the top and the section stops feeling like a
grab-bag of tricks.

The captured DCE result is the key evidence:

```text
DCE       0.2532 ns/op   0 B/op   0 allocs/op
correct  11.14 ns/op  64 B/op   1 allocs/op
```

`make([]byte, 64)` has no path that avoids allocating. Zero allocations means the
call disappeared. Land this slowly: "`ns/op` has a timer floor. `allocs/op` tells
us whether the allocation happened."

The two-variable sink keeps a local in the loop and publishes once after it, so
there is no global write per iteration. The slide says test-only; say it out loud
too, because this was the single most confusing point in the rehearsal.

**Assembly is now two slides.** The first explains how to read the second: `MOVD
$3` loads a literal, `VCNT`/`VUADDLV` are ARM64 population-count instructions.
Most of the room is not fluent in ARM64 assembly, so do the primer before the
side-by-side and keep it to two sentences per column.

Inlining feeds DCE inside a benchmark. `//go:noinline` is a diagnostic, not
production style.

Timer ordering: `ResetTimer` clears elapsed time but does not stop the timer. For
per-iteration setup, stop before the fixture and restart before the operation.
The captured output shows the broken version looking 25% faster because it
excludes the function it claims to measure.

The non-terminating timer case stays, but "We tried it. It hung." is gone. State
the mechanism only: without a matching `StartTimer`, timed duration never
accumulates, so the framework keeps doubling `b.N`.

`B.Loop` excludes setup, stops timing after the loop, runs the function once per
`-count`, and prevents DCE when written literally as `b.Loop()`. Do not read the
table; point at setup exclusion and DCE prevention.

### §1B — the regression that was a speedup

Now pay off the plant. Read the benchmark first: the timed loop is
`proto.Size(tracesData)` on a struct built before `ResetTimer`, and it never
touches code the PR changed. Locally, run-to-run variance was under 0.1%.

Then the same-machine A/B: main 883.3 ns, #4891 840.7 ns. The PR was faster. CI
had inverted the sign. Let that sit.

Mechanism: restructuring `context.go` moved function addresses, which moved the
hot loop relative to cache-line and branch-target-buffer boundaries. At ~390 ns
per iteration that is worth several percent in either direction. Resolution: no
code change.

Berger's *Performance Matters* establishes that code layout alone moves results
by ±10%.

**Causal profiling is the bridge to arc 2.** A flat profile says `encode` is 30%
of CPU; causal profiling says making `encode` 20% faster moves end-to-end by 2%.
Present the numbers as illustrative of the mechanism, not as a measurement
(ledger row 19). Then the consequence: a component speedup does not imply a
system speedup, so a microbenchmark cannot tell you whether the work mattered.
That sentence is the reason the second half of the talk exists.

### §1C — statistics

The 43% swing is two runs of the same binary on a deliberately loaded machine.
`11.32n` is the median. A `~` result means no measurable difference, and that is
a result rather than an invitation to rerun.

Benchstat compares distributions; it does not tell you whether the machine is
trustworthy. CV is that check. Ten runs is the floor, twenty is better, and the
run count is set before looking at results.

### §1D — local reproduction

"So we measured it" is gone; the slide states the gap and moves on.

The Docker table is committed captured output. **State the platform explicitly,
because this was asked directly in the rehearsal:** Apple M4 Max, darwin/arm64,
16 logical CPUs; the container is linux/arm64 under Docker Desktop's Apple
Virtualization.framework VM. No QEMU, no cross-architecture emulation, guest ISA
matches host.

Loaded was 3× slower and 4× noisier. Pinning restored the idle noise floor while
the host stayed saturated. Then the ceiling: 5.25% is useful isolation, not
publication-quality control. Bare-metal Linux with SMT off reaches ~0.05%.

Do not soften the macOS caveat. Pinning a vCPU inside a VM is not pinning a
physical core.

perflock is now framed as the answer to "is there a Go tool that sets this up?"
because that is exactly how it came up. Then the caveat from source inspection:
on macOS the mutual-exclusion lock works, frequency pinning does not, and the
default `-governor 90` errors (ledger row 16).

Close the arc with the recap slide: all three questions answered at micro scale,
plus the layout-noise warning.

## Part 4 — arc 2, CI and macro

### Arc divider and #643

The divider reframes the same three questions: representative workload, stability
across days, and hardware as the variable.

Then the second incident. Local overhead run: `multi` 230%, `largeidle` 212%,
ceiling 150%. Both over. Looks like a serious regression.

Then the turn: `largeidle` shares zero bumped dependencies with the change. It
runs independent code and cannot have been affected, yet it moved by the same
~60%. The machine was the variable — heavy parallel builds had been running all
session. `benchmark/threshold` is CI-only with no local target, which is why the
local number was untrustworthy. Resolution: wait for CI.

**The mirror-image slide is the payoff of having two stories.** #4891: CI said
slower, the laptop was right, cause was code layout. #643: the laptop said
slower, CI was right, cause was machine load. Neither environment is
authoritative by default, and the tell was the same both times — a benchmark
moved that could not have moved.

### §2A — designing a macrobenchmark

Representative workloads as a three-step reveal, then the archetype table.

The archetype tie-back slide is why the taxonomy is here rather than being
trivia: `largeidle` is the Idle archetype, `multi` is Enterprise. Two archetypes
exercising different code moved by the same amount at the same time, which is
shared environment rather than shared cause.

Macro gates compare against a budget, not against a parent commit. An overhead
ceiling answers a product question: how much of the customer's machine may we
consume.

The scale slide is deliberately generic — dedicated hardware, a budget per
component, several archetypes, gating the release. **Claims ledger row 23 is
`pending`:** the Datadog-specific version of this slide is not cleared. Do not
ad-lib internal specifics on stage.

Two macro traps: coordinated omission, where a load generator stops recording the
latency it caused, and non-deterministic inputs.

### §2B — controlling the CI environment

SMT and DFS each get an explainer before their chart, because the rehearsal used
both acronyms without defining them.

SMT: two hardware threads share one core's execution units. Good for throughput,
fatal for repeatability, because your runtime now depends on an invisible
co-tenant. Then the chart: 23.887% CV down to 0.044%, and note it also got twice
as fast because the core stopped sharing.

DFS: the CPU changes its own clock. Run 1 boosts, run 20 is warm and throttles.
The chart shows ~10× less variance with it off — and the mean gets *slower*. Say
the line on the slide: a benchmark's job is to be comparable, not to post the
best number you can reach once.

Three sysfs writes buy ~100×. In a VM the hypervisor owns SMT, frequency is
virtualised, and the write may succeed and be silently ignored.

### §2C — change over time

Benchstat compares two distributions. CI has a time series with hardware changes
and dependency bumps in it. Comparing each commit to its parent re-asks the noise
question every time and misses slow drift.

The ideal-vs-actual pair is the honest version: a step change is obvious in a
diagram and buried in variance in real data.

Change-point detection asks where the distribution shifted and stayed shifted.
ED-PELT, e-divisive means, Apache Otava, Netflix at scale. It needs history, so
it belongs in nightly trending rather than a PR gate.

### §2D — wiring it up

PR gate versus nightly suite have opposing requirements. CI detects; it is not
the primary measurement.

The feedback-loop diagram carries one line: benchmarks have to be locally
reproducible for a developer to act on them, which is why arc 1 came first.

The false-positive ledger slide is the institutional-memory point.
`BenchmarkOTLPProtoSize` fired again eleven days later and took one comment to
dismiss because the first investigation was written down.

Close the arc with its recap slide.

## Part 5 — tools and close

No terminal. Every output shown is captured from the repository.

- `honestbench` — AST analysis, finds discarded results and timer misordering
  without running anything
- `benchgate` — runs N times and fails when CV is too high to trust
- `benchenv` — reports which controls exist on this machine

Pause on the punchline. Four `unavailable` lines are the honest answer: macOS
does not expose Linux sysfs controls, so the output says `unavailable` rather
than inventing a green check.

Then the minimum viable discipline, then the 3×2 recap. The two lines that must
land: a sub-10% micro delta can be layout noise, and a shared runner cannot be
fixed with more samples.

End with: "Trust your numbers only after the compiler, the sample, and the
machine have each earned that trust." Leave the questions slide up.

## Delivery constraints

- No live or recorded demos. Every output is captured from the repository.
- If a result is challenged, name its committed result file and the environment.
  Do not improvise a rerun on stage.
- Do not treat Docker Desktop on macOS as hardware isolation.
- Do not claim statistical significance proves practical importance.
- Do not present ledger row 23 content beyond what is on the slide.
- Present from the bespoke HTML if you want the reveals to animate. The PDF is a
  correct handout but shows every fragment at once.

## Public references for questions

| Claim | Reference |
| --- | --- |
| OPERA connector fault | Cartlidge, *Science* 335(6072):1027, doi:10.1126/science.335.6072.1027 |
| Google 500 ms result | Claims ledger row 10 |
| Daniel Martí's talk | <https://www.youtube.com/watch?v=oE_vm7KeV_E> |
| Go `B.Loop` behaviour | <https://pkg.go.dev/testing#B.Loop> |
| Benchstat | <https://pkg.go.dev/golang.org/x/perf/cmd/benchstat> |
| Causal profiling | <https://arxiv.org/abs/1608.03676> |
| Performance Matters | <https://www.youtube.com/watch?v=r-TLSBdHe1A> |
| ED-PELT | <https://aakinshin.net/posts/edpelt/> |
| Apache Otava | <https://github.com/nyrkio/nyrkio> |
| Netflix change-point detection | <https://netflixtechblog.com/fixing-performance-regressions-before-they-happen-eab2602b86fe> |
| Datadog PGO result | <https://www.datadoghq.com/blog/datadog-pgo-go/> |
| dd-trace-go #4891 | `research/go-benchmarks-lying/07-war-stories.md` story 1 |
| otel-go-compile-instrumentation #643 | `research/go-benchmarks-lying/07-war-stories.md` story 2 |
| Prior FOSDEM version | <https://youtu.be/8211fNI_nc4> |

# Talk outline: Why Your Go Benchmarks Are Lying

> **Format:** 60 minutes including Q&A · advanced Go audience
> **Deck:** `slides/presentation.md` (76 slides)
> **Current state:** additive review version; pruning follows Kemal's deck review and rehearsal
> **Demonstrations:** no live demos; every command and captured result remains available in the repository

## Thesis

A benchmark is a measurement system. It can report a precise number while measuring removed work, an unstable sample, or an uncontrolled machine. Trust a Go benchmark only after answering three questions:

1. Is the compiler measuring real work?
2. Is the sample stable enough?
3. Is the difference large relative to the noise?

## Beat sheet

| Section | Slides | Expanded time | Purpose |
| --- | ---: | ---: | --- |
| Loose-cable story and thesis | 2-8 | 4:00 | Open with OPERA's faster-than-light result, reveal the connector fault, map systematic error to Go benchmarks, and plant the three questions. |
| Personal and Datadog context | 9-10 | 1:30 | Establish relevant Go/observability experience and explain why SDK overhead inside customer workloads makes continuous benchmarking product work. |
| Why benchmark | 11-14 | 3:30 | Preserve the latency, throughput, Google traffic, and performance-as-feature evidence. |
| Before optimizing | 15-18 | 4:00 | Use production profiles, SLOs, error budgets, Amdahl's Law, and Daniel Martí's prior talk to choose the right target. |
| The art of benchmarking | 19-21 | 3:00 | Separate microbenchmarks from macrobenchmarks and state the talk's `testing.B` scope. |
| Layer 1: compiler honesty | 22-37 | 14:00 | DCE, sinks, constant folding, inlining, timer ordering, the non-terminating timer case, and `B.Loop`. |
| Layer 2: statistics | 38-45 | 10:00 | Repeated samples, benchstat, CV, run-count discipline, and p-hacking. |
| Layer 3A: local reproduction | 46-53 | 11:00 | Measured Docker isolation, the macOS VM boundary, Linux controls, perflock, and benchdiff. |
| Layer 3B: CI | 54-58 | 5:00 | Shared-runner failure, AWS m5.metal data, PR/nightly patterns, and existing CI services. |
| War story | 59-65 | 5:00 | Walk through dd-trace-go #4891, show the sign inversion, explain code layout, and state the directionally-wrong result. |
| Tools, results, and CTA | 66-76 | 6:00 | Preserve the three authored tools and captured outputs, return to the three questions, and close with repository/social links and QR. |
| **Expanded delivery** | **76** | **66:00** | **Deliberately over the slot until pruning.** |

The expanded deck is a content-review artifact. Do not hide the timing problem. Choose cuts after Kemal reviews the complete story and evidence.

## Narrative order

### Cold open

Start with OPERA, not a biography or an agenda. The audience sees a careful team, a surprising result, and two faults that partially masked each other. Pivot to a laptop benchmark only after the connector reveal.

### Thesis before credibility

Plant the three trust questions and the local-first rule. Then show the personal ethos and Datadog slides. The company slide is not generic branding: several SDKs execute inside customer workloads, PGO reduced production CPU by 3.4%, and benchmark gates protect tracer hot paths.

### Data-backed reason to benchmark

Keep the existing intro evidence after the story: latency and throughput definitions, the Google 500 ms/20% traffic result, and Tobi Lütke's quote. These slides build the reason to care rather than repeating the story.

### Three-layer body

Each technical section answers one planted question:

- Compiler honesty answers whether work happened.
- Statistical interpretation answers whether the sample is stable.
- Local and CI control answer whether the observed difference is larger than the environment's noise.

### War story and close

The dd-trace-go regression story combines all three layers. The benchmark measured code the PR did not touch, the same-machine result reversed the CI sign, and code layout explained the small shift. The authored tools turn those lessons into repeatable checks.

## Additive-pass rule

Preserve useful existing facts, outputs, references, and cautionary examples. This revision changes framing rather than deleting evidence.

- Captured command outputs replace live-demo labels.
- Existing data slides remain.
- The non-terminating timer example, Daniel Martí reference, and CI-tool survey remain.
- The code-layout lesson gets its own slide so it no longer overflows.
- Pruning happens only after review and rehearsal.

## Public results and sources to preserve

- Google search team latency result: `research/go-benchmarks-lying/claims-ledger.md` and source S26.
- OPERA connector bias and two-fault account: CERN press release and Cartlidge, *Science* 335(6072):1027.
- Local CV experiment: committed files under `talks/go-benchmarks-lying/demo/results/`.
- AWS m5.metal SMT/DFS data: FOSDEM 2026 experiments.
- dd-trace-go #4891: same-machine A/B results and documented resolution.
- PGO production reduction: <https://www.datadoghq.com/blog/datadog-pgo-go/>.
- Prior FOSDEM version: <https://youtu.be/8211fNI_nc4>.

## Audience resources

- Talk repository: <https://github.com/kakkoyun/gopherconuk-26>
- Benchmark tools and skills: <https://github.com/kakkoyun/benchlab>
- Binaries: `go install github.com/kakkoyun/benchlab/cmd/...@latest`
- Agent Skills: `npx skills add kakkoyun/benchlab --all`

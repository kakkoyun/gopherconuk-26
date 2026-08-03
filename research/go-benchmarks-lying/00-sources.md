# Annotated Bibliography
## "Why Your Go Benchmarks Are Lying (And How to Stop Them)"

All sources verified as of 2026-07-22 unless marked [UNVERIFIED].

---

## Go Toolchain & Standard Library

**[S1]** Go 1.24 release notes — `testing.B.Loop` introduction
https://go.dev/doc/go1.24
Verified: B.Loop introduced in Go 1.24 (proposal by Austin Clements, issue #61515).
Used in: `01-compiler-honesty.md`, slides Layer 1.

**[S2]** `testing` package documentation
https://pkg.go.dev/testing
The authoritative reference for `b.N`, `b.ResetTimer`, `b.StopTimer`, `b.StartTimer`, `b.Loop`, `AllocsPerOp`, `-benchmem`.
Used in: `01-compiler-honesty.md`.

**[S3]** `golang.org/x/perf/cmd/benchstat` documentation
https://pkg.go.dev/golang.org/x/perf/cmd/benchstat
Verified: statistical analysis tool; computes geomean, delta, confidence interval, p-value from Go benchmark output.
Used in: `02-statistics.md`, `04-ci-continuous.md`.

**[S4]** golang.org/x/perf repository (benchstat, benchsave, perfdata)
https://pkg.go.dev/golang.org/x/perf
The official Go performance measurement toolchain. Actively maintained by the Go team.
Used in: `04-ci-continuous.md`.

---

## Go Benchmarking Guides

**[S5]** Dave Cheney — "How to write benchmarks in Go" (2013)
https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go
Foundational article. Introduced the sink-variable pattern for defeating DCE. Still accurate; the `testing.B.Loop` addition (Go 1.24) supersedes some b.N guidance.
Used in: `01-compiler-honesty.md`.

**[S6]** Damian Gryski — go-perfbook
https://github.com/dgryski/go-perfbook
Comprehensive Go performance guide. Covers benchmarking discipline, compiler behaviour, profiling. Actively maintained.
Used in: `01-compiler-honesty.md`.

**[S7]** Austin Clements — `testing.B.Loop` proposal (GitHub issue #61515)
https://github.com/golang/go/issues/61515
Original proposal for B.Loop, explaining the b.N footguns it addresses. Authored by Austin Clements (Go runtime team).
Used in: `01-compiler-honesty.md`, slides Layer 1.

---

## Benchmarking Environment & Statistics

**[S8]** FOSDEM 2026 experiments — SMT/DFS variance data
https://github.com/igoragoli/fosdem-2026-software-performance
Jupyter notebooks with reproducible AWS m5.metal experiments. Verified data:
- SMT enabled: CV ~23.9% → SMT disabled: CV ~0.05% (~100× reduction)
- DFS enabled 1-task: CV ~0.383% → DFS disabled: CV ~0.041% (~10× reduction)
Co-authored by Kemal Akkoyun and Augusto de Oliveira. Companion blog post at `/posts/fosdem-2026-measuring-software-performance/`.
Used in: `03-local-reproduction.md`, `04-ci-continuous.md`, slides Layer 3.

**[S9]** MongoDB Engineering — "Reducing Variability in EC2 Performance Tests"
https://www.mongodb.com/company/blog/engineering/reducing-variability-performance-tests-ec2-setup-key-results
Real-world case study on EC2 bare-metal setup for benchmark reproducibility. Corroborates SMT/DFS findings.
Used in: `04-ci-continuous.md`.

**[S10]** Denis Bakhvalov — "Performance Analysis and Tuning on Modern CPUs"
https://github.com/dendibakh/perf-book
Definitive reference on CPU-level performance tuning: SMT, frequency scaling, affinity, cache behaviour. Freely available on GitHub.
Used in: `03-local-reproduction.md`, `04-ci-continuous.md`.

**[S11]** Andrey Akinshin — ED-PELT change-point detection for benchmarks
https://aakinshin.net/posts/edpelt/
Statistical foundation for change-point detection in continuous benchmarking. The algorithm used by Apache Otava (Nyrkiö).
Used in: `02-statistics.md`, `05-existing-tools.md`.

**[S12]** Tomas Kalibera & Richard Jones — "Rigorous Benchmarking in Reasonable Time" (ECOOP 2013)
https://dl.acm.org/doi/10.1145/2509136.2509184
Research paper establishing minimum sample counts for reliable benchmarking. Key finding: N ≥ 30 runs required for meaningful inter-run statistics.
Used in: `02-statistics.md`.

**[S13]** Netflix Engineering — "Fixing Performance Regressions Before They Happen"
https://netflixtechblog.com/fixing-performance-regressions-before-they-happen-eab2602b86fe
Case study on continuous benchmarking and change-point detection at scale. References ED-PELT.
Used in: `02-statistics.md`, `04-ci-continuous.md`.

---

## Local Tooling

**[S14]** `perflock` — CPU frequency lock daemon for Go benchmarks
https://github.com/aclements/perflock
Written by Austin Clements (Go team). Locks CPU frequency for the duration of a benchmark, then releases. The canonical local Go benchmark frequency-stabiliser.
Used in: `03-local-reproduction.md`, `04-ci-continuous.md`, slides Layer 3.

**[S15]** `benchdiff` — automated git-ref A/B benchmark comparison
https://github.com/willabides/benchdiff
Automates: stash changes → run benchmarks on HEAD → pop stash → run again → pipe to benchstat. The ideal local developer-iteration loop.
Used in: `03-local-reproduction.md`, slides Layer 3.

---

## Continuous Benchmark Tools

**[S16]** `benchmark-action/github-action-benchmark`
https://github.com/benchmark-action/github-action-benchmark
GitHub Action for benchmark tracking. Go native (parses `go test -bench` output). Stores history in GitHub Pages. v1.22.1 as of May 2026. Percentage threshold only (no change-point detection). Recommended for small OSS projects.
Used in: `05-existing-tools.md`.

**[S17]** bencher.dev
https://bencher.dev
Continuous benchmarking service. Go support via `go_bench` adapter. Statistical models: t-test, z-score, IQR. Free self-hosted binary. Recommended for teams with a dedicated runner.
Used in: `05-existing-tools.md`.

**[S18]** Apache Otava (formerly Nyrkiö) — change-point detection benchmarking
https://github.com/nyrkio/nyrkio
Apache Incubator project (since Nov 2024). Uses e-divisive means algorithm for change-point detection. Nyrkiö 2.0.0 shipped Feb 2026. The only tool in this survey with real change-point detection. Recommended for large orgs.
Used in: `05-existing-tools.md`.

**[S19]** gobenchdata — [MAINTENANCE MODE]
https://github.com/bobheadxi/gobenchdata
Last substantive commit Jan 2023. Do not adopt for new projects.

**[S20]** cob — [USE WITH CAUTION]
https://github.com/nakabonne/cob
Last release Oct 2023. Known issue: uses `git reset` as part of its workflow, which can lose uncommitted changes. Borderline maintained.

**[S21]** codespeed — [UNMAINTAINED]
https://github.com/tobami/codespeed
Last commit Feb 2019. Do not adopt.

---

## Narrative & Performance Context

**[S22]** OPERA experiment — Wikipedia
https://en.wikipedia.org/wiki/OPERA_experiment
The 730 km CERN-to-Gran-Sasso neutrino detector. 2011: neutrinos appeared faster-than-light. Root cause: a loose fibre-optic cable connector introducing a 73 ns timing error. Cold open for the talk.

**[S23]** Gil Tene — "How NOT to Measure Latency" (2015)
https://www.youtube.com/watch?v=lJ8ydIuPFeU
Defines the coordinated omission problem. Brief mention in Layer 2 for completeness.

**[S24]** Brendan Gregg — Frequency Trails: Outliers
https://www.brendangregg.com/FrequencyTrails/outliers.html
Explains how visualisation choice (strip plots over boxplots) affects ability to detect real patterns in latency data.

**[S25]** Brendan Gregg — Systems Performance: Enterprise and the Cloud (2nd ed.)
https://www.brendangregg.com/blog/2020-07-15/systems-performance-2nd-edition.html
The definitive reference on systems performance methodology.

**[S25b]** Tobi Lütke (@tobi), X, 5 May 2024 — https://x.com/tobi/status/1787139157078188180
"Not all fast software is world-class, but all world-class software is fast. Performance is _the_ killer feature." [VERIFIED 2026-08-03 — the quote is Lütke's, not Gregg's; an earlier draft of this file attributed it to S25 by mistake. The link needs an X login to render.]

**[S26]** Google latency research — "500ms delay costs 20% traffic"
[VERIFIED for use with a constraint: attribute to "the Google search team". The Marissa Mayer attribution could not be confirmed from primary sources — do not name her. Secondary citation: Jeff Atwood, "Performance is a Feature," https://blog.codinghorror.com/performance-is-a-feature/]

**[S27]** Yahoo latency research — "400ms improvement gave 5-9% more traffic"
[DROPPED — the YUI Blog original is defunct and archive retrieval was blocked. Do not use in slides or posts. Use S26 instead.]

**[S28]** CERN press release, 22 February 2012 — https://home.cern/news/press-release/cern/opera-experiment-reports-anomaly-flight-time-neutrinos-cern-gran-sasso
[VERIFIED 2026-08-03] Names OPERA's two candidate faults (fibre-optic connector, oscillator) and their opposite directions. No nanosecond figures.

**[S29]** Edwin Cartlidge, "Loose cable may unravel faster-than-light result," *Science* 335(6072):1027 (2012), doi:10.1126/science.335.6072.1027
[VERIFIED 2026-08-03] Source for the ~73 ns connector bias. Note arXiv:1109.4897v4 is *not* a valid source for this figure — it reports the corrected result only.

---

## War Stories (Primary Sources)

**[S28]** dd-trace-go PR #4891 — "CI regression that was a speedup"
ctx session c8e85c6a (2026-06-12). Full reconstruction in `07-war-stories.md`.
Numbers: main 883.3 ns / PR 840.7 ns (1-span); opposite of CI's reported 6-9% regression.

**[S29]** otel-go-compile-instrumentation PR #643 — "machine running too hot"
ctx session c26c53cb (2026-07-03). Full reconstruction in `07-war-stories.md`.
230% overhead on both touched and untouched scenarios — hot machine, not a regression.

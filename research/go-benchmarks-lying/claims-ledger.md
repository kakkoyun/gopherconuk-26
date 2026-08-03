# Claims Ledger

Every factual claim that will appear in slides or blog posts lives here with a primary source and verification status.

**Status values**: `pending` | `verified` | `disputed` | `dropped`

**Rule**: A claim with any status other than `verified` must NOT appear in slides or published blog posts.

---

| # | Claim | Source | URL | Status | Notes |
|---|-------|--------|-----|--------|-------|
| 1 | SMT enabled → CV ~23% on CPU-bound benchmark on AWS m5.metal (DFS disabled) | FOSDEM 2026 experiments | https://github.com/igoragoli/fosdem-2026-software-performance | verified | Reproduced in experiments; table in FOSDEM blog post |
| 2 | SMT disabled → CV drops to ~0.044–0.235% (~100x reduction) | FOSDEM 2026 experiments | https://github.com/igoragoli/fosdem-2026-software-performance | verified | Same source |
| 3 | DFS enabled 1-task → CV ~0.383%; DFS disabled → CV ~0.041% (~10x reduction) | FOSDEM 2026 experiments | https://github.com/igoragoli/fosdem-2026-software-performance | verified | Same source |
| 4 | `testing.B.Loop` introduced in Go 1.24 | Go 1.24 release notes | https://go.dev/doc/go1.24 | verified | Confirmed in release notes and proposal #61515 |
| 5 | Dead-code elimination can reduce a benchmark hot loop to a no-op | 01-compiler-honesty.md (agent-verified, Go compiler behaviour) | https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go | verified | Demo confirmed: allocs/op=0 vs 1 |
| 6 | `benchdiff` automates git-ref A/B benchmarking piped to benchstat | github.com/willabides/benchdiff | https://github.com/willabides/benchdiff | verified | Tool exists, actively maintained |
| 7 | `perflock` is a CPU-frequency lock daemon for Go benchmarks | github.com/aclements/perflock | https://github.com/aclements/perflock | verified | Tool exists, written by Austin Clements (Go team) |
| 8 | Docker on macOS runs inside a VM; --cpuset-cpus pins vCPUs not physical cores | Docker resource constraints docs | https://docs.docker.com/engine/containers/resource_constraints/ | verified | Confirmed; Mac-VM penalty quantified: ~1-2% CV floor vs 0.05% on bare-metal Linux |
| 9 | OPERA: an improperly seated fibre-optic connector introduced a ~73 ns early-arrival bias | CERN press release (cable root cause) + Cartlidge, *Science* 335(6072):1027 (figure) | https://home.cern/news/press-release/cern/opera-experiment-reports-anomaly-flight-time-neutrinos-cern-gran-sasso ; doi:10.1126/science.335.6072.1027 | verified | Verified 2026-08-03. **Not** in arXiv:1109.4897v4 — that paper reports the corrected result only. Two faults, opposite signs: connector ~73 ns early, oscillator ~37 ns late, net anomaly ~58–62 ns. Approved wording: "An improperly seated fibre-optic connector in OPERA's GPS timing chain introduced a ~73 ns bias — enough to explain the apparent superluminal signal once a partially-offsetting oscillator fault is accounted for." Do not cite a specific oscillator figure. |
| 10 | "500ms delay costs Google 20% of their traffic" | Coding Horror citing Google data | — | verified | Attribute to "Google search team," NOT Marissa Mayer (attribution unconfirmed) |
| 11 | "400ms improvement gave Yahoo 5-9% more traffic" | YUI Blog (defunct) | — | dropped | Source site is defunct; Wayback Machine blocked; do not use on slides |
| 12 | Tobi Lütke: "Not all fast software is world-class, but all world-class software is fast. Performance is _the_ killer feature." | Tobi Lütke (@tobi), X | https://x.com/tobi/status/1787139157078188180 | verified | Verified 2026-08-03, posted 5 May 2024 — **not** the 2018 tweet ID previously recorded. Exact wording corroborated across independent excerpts; the URL itself needs an X login to render, so a reader may not be able to open it. Cite as "Tobi Lütke (@tobi), X, 5 May 2024". |
| 13 | Gil Tene "How NOT to Measure Latency" talk exists on YouTube | YouTube | https://www.youtube.com/watch?v=lJ8ydIuPFeU | verified | Confirmed: title, speaker, URL all valid |
| 14 | Bakhvalov "Performance Analysis and Tuning on Modern CPUs" available on GitHub under CC0 | GitHub | https://github.com/dendibakh/perf-book | verified | Confirmed: title, author, licence |
| 15 | Brendan Gregg frequency trails page documents outlier rates | brendangregg.com | https://www.brendangregg.com/FrequencyTrails/outliers.html | verified | Page content confirmed; 100%/98%/96% server outlier rates quotable |
| 16 | `perflock` builds on macOS, but its frequency pinning is Linux-only: `Domains()` reads `/sys/devices/system/cpu/`, so the default `-governor 90` errors on macOS. Pass `-governor=none` for lock-only behaviour. On Linux it writes `scaling_min_freq`/`scaling_max_freq` via cpufreq sysfs — not `intel_pstate`. | Source inspection: `internal/cpupower/cpupower.go`, `cmd/perflock/main.go` | https://github.com/aclements/perflock | verified | Verified 2026-08-03 by reading the source. The README says nothing about platform support, so cite the source files, not the README. |

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
| 9 | OPERA experiment root cause: a loose fibre-optic cable introduced a 73 ns timing error | Wikipedia OPERA experiment | https://en.wikipedia.org/wiki/OPERA_experiment | pending | Wikipedia text does not mention the 73 ns figure explicitly; verify from arXiv:1109.4897v4 before quoting on stage |
| 10 | "500ms delay costs Google 20% of their traffic" | Coding Horror citing Google data | — | verified | Attribute to "Google search team," NOT Marissa Mayer (attribution unconfirmed) |
| 11 | "400ms improvement gave Yahoo 5-9% more traffic" | YUI Blog (defunct) | — | dropped | Source site is defunct; Wayback Machine blocked; do not use on slides |
| 12 | Tobi Lütke quote: "Not all fast software is world-class, but all world-class software is fast" | X/Twitter (requires auth) | — | pending | Cannot verify exact wording without auth. Paraphrase with tweet-ID citation only after manual verification |
| 13 | Gil Tene "How NOT to Measure Latency" talk exists on YouTube | YouTube | https://www.youtube.com/watch?v=lJ8ydIuPFeU | verified | Confirmed: title, speaker, URL all valid |
| 14 | Bakhvalov "Performance Analysis and Tuning on Modern CPUs" available on GitHub under CC0 | GitHub | https://github.com/dendibakh/perf-book | verified | Confirmed: title, author, licence |
| 15 | Brendan Gregg frequency trails page documents outlier rates | brendangregg.com | https://www.brendangregg.com/FrequencyTrails/outliers.html | verified | Page content confirmed; 100%/98%/96% server outlier rates quotable |

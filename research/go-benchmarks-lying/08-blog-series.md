# Blog Series Map

The corpus is the source; the series is the published surface. This table is how you tell
whether they have drifted.

**Series name** (Hugo `series` taxonomy value, identical in all five posts):
`Why Your Go Benchmarks Are Lying`

**Blog repo**: `github.com/kakkoyun/me` (`~/Vaults/blog`), default branch `master`.
Posts live at `content/posts/<slug>.md`, flat files, `draft: true` until scheduled.

| # | Slug | Corpus source | Centerpiece evidence | Status |
|---|------|---------------|----------------------|--------|
| 1 | `go-benchmarks-lying-compiler-honesty` | `01-compiler-honesty.md` | `allocs/op` 0 vs 1 from `demo/dce_bench_test.go` | draft |
| 2 | `go-benchmarks-lying-statistics` | `02-statistics.md` | benchstat output + CV from `demo/cv.awk` | draft |
| 3 | `go-benchmarks-lying-local-reproduction` | `03-local-reproduction.md` | CV 4.75 / 18.88 / 5.25 from `demo/results/` | draft |
| 4 | `go-benchmarks-lying-ci` | `04-ci-continuous.md`, `05-existing-tools.md` | FOSDEM SMT/DFS tables | draft |
| 5 | `go-benchmarks-lying-three-questions` | `06-narrative.md`, `07-war-stories.md`, `tools/` | dd-trace-go war story; the three CLIs | draft |

## Cadence

Post 1 publishes as a pre-conference teaser, roughly a week before the GopherCon UK slot;
posts 2–5 weekly after. Dates in the frontmatter are placeholders (2026-09-01 onward) until
the slot is confirmed. Every post stays `draft: true` until then.

## Rules

- A post may only carry claims whose `claims-ledger.md` row reads `verified`.
- Numbers and code come from committed artifacts in this repo, never retyped from memory.
  Every figure in a post should be greppable here.
- The corpus is notes. Posts are written fresh against it, not assembled from it.
- `de-slop` then `humanizer` before any post ships.
- Each post cross-links the FOSDEM post as the language-agnostic prerequisite, plus its
  neighbours in the series.

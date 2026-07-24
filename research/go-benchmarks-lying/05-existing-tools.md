# Existing Go Continuous Benchmark Tools

Don't rebuild what already exists. Before wiring your own ad-hoc benchmark
comparison script, it's worth surveying the landscape. Several tools tackle
this problem, with different trade-offs on statistics, hosting, Go-nativeness,
and operational complexity. This document surveys eight candidates, compares
them on the axes that matter for a talk about benchmark correctness, and ends
with concrete recommendations and a wire-up sketch.

---

## Tool Survey

### 1. bencher.dev

**Home:** <https://bencher.dev>  
**Go native:** Yes — `--adapter go_bench` parses `go test -bench` stdout directly; captures `ns/op` as the latency measure (mean only; no lower/upper bound is populated from the raw output).  
**Hosted / self-hosted:** Both. The cloud service at bencher.dev is the primary offering. A self-hosted binary is available for free (equivalent to the Free cloud tier); Enterprise on-prem unlocks Pro features, billed annually.  
**Statistics:** Bencher is the most statistically sophisticated of the threshold-based tools. Supported test models per benchmark series:
- **Percentage** — simple ± percentage from historical mean
- **z-score** — standard deviations from mean (needs ≥ 30 samples)
- **t-test** — Student's t prediction intervals (works with smaller samples)
- **Log Normal** — log-normal distribution likelihood
- **IQR / Delta IQR** — interquartile range multiples from median

None of these is change-point detection (no ED-PELT, no e-divisive). Bencher detects regressions by comparing each new result against a rolling window of past results using the chosen statistical model.

**PR gate:** First-class. `--error-on-alert` exits non-zero on alert; `--github-actions $GITHUB_TOKEN` annotates the PR with inline alerts. The `bencher run` command is a single CLI invocation.  
**Pricing:** Free tier (public projects only, 65,535 metrics/day, community support). Pro starts at $100/month flat for up to 250 active series, scaling to $200/month at 500 series. On-demand bare-metal runners at $1/hr extra. Enterprise is custom.  
**Maintenance:** Actively developed. Docs last updated March 27, 2024 (bencher.dev site); the GitHub Action and CLI see ongoing releases. The quickstart covers Go, Rust, C++, Python, and others.  
**Setup complexity:** Moderate. Requires creating a project and API key on bencher.dev (or self-hosted instance), adding a repository secret, and writing a small workflow step. The `go_bench` adapter requires no code changes — just pipe `go test -bench` output through `bencher run`.

---

### 2. benchmark-action/github-action-benchmark

**Home:** <https://github.com/benchmark-action/github-action-benchmark>  
**Go native:** Yes — one of nine officially supported harnesses; expects `go test -bench` stdout.  
**Hosted / self-hosted:** GitHub-native. Results are committed to a `gh-pages` branch as `data.js` (or stored via `actions/cache`). No external service required.  
**Statistics:** Threshold-based only. A single configurable percentage (default 200%, meaning the benchmark must not get 2× slower). A second `fail-threshold` can be set for hard failures. No statistical significance testing; no change-point detection. Noise-induced false positives are a real risk at tight thresholds.  
**PR gate:** Yes, via GitHub Actions status checks. The action can comment on a PR and set a failing check when the threshold is breached.  
**Maintenance:** Actively maintained. Latest release: **v1.22.1, May 6, 2026**. 52 total releases. 1,200 stars, 183 forks.  
**Setup complexity:** Low-to-moderate. Requires a `gh-pages` branch (orphan) and a few workflow steps. No external accounts or tokens beyond the standard `GITHUB_TOKEN`.

---

### 3. gobenchdata

**Home:** <https://github.com/bobheadxi/gobenchdata>  
**Go native:** Yes — purpose-built for `go test -bench` output, stores results as JSON.  
**Hosted / self-hosted:** GitHub-native. Results pushed to `gh-pages`.  
**Statistics:** None formal. Regression detection is a user-defined arithmetic expression evaluated against two JSON result sets (e.g., `(current.NsPerOp - base.NsPerOp) / base.NsPerOp * 100`). Users set a `max` threshold manually. No p-values, no confidence intervals.  
**PR gate:** Yes — the GitHub Action can fail a PR check when the expression exceeds the threshold.  
**Maintenance:** Low. Latest release: **v1.3.1, January 30, 2023**. 155 stars, 15 forks, 2 watchers. Nine open issues, three open PRs with no recent maintainer activity. Treat as maintenance-mode.  
**Setup complexity:** Low. Docker-based GitHub Action; minimal YAML configuration needed.

---

### 4. cob (knqyf263/cob)

**Home:** <https://github.com/knqyf263/cob>  
**Go native:** Yes — wraps `go test -run '^$' -bench . -benchmem ./...` directly; no adapter layer needed.  
**Hosted / self-hosted:** Local/CI only. No external service; results are ephemeral (printed to stdout for the current run).  
**Statistics:** Simple percentage threshold. Default is 20% (the run fails if any benchmark degrades by more than 20%). Configurable via `--threshold`. No historical tracking, no statistical significance.  
**PR gate:** Yes, via exit code — integrate `cob` as a CI step and it fails the build on regression. Supports `--base origin/main` for PR branch comparisons.  
**Maintenance:** Borderline. Latest release: **v0.0.8, October 27, 2023**. 390 stars, 25 forks. Five open issues, seven open PRs with no recent maintainer response — multiple forks (szuecs/cob, JuanJoseGonGi/cob) exist because issues went unaddressed.  
**Setup complexity:** Very low. Single binary, single command. **Critical caveat:** cob internally runs `git reset` — all changes must be committed before running it. This makes it unsuitable for pre-commit hooks or dirty-working-tree CI steps.

---

### 5. chronologer (dandavison/chronologer)

**Home:** <https://github.com/dandavison/chronologer>  
**Go native:** No. chronologer uses [hyperfine](https://github.com/sharkdp/hyperfine) as its benchmarking engine — it times arbitrary shell commands over git commit history. It does **not** parse `go test -bench` output or the Go benchmark format. Suitable only for coarse end-to-end command timing across commits, not for sub-microsecond benchmark tracking.  
**Hosted / self-hosted:** Local only — generates a static HTML visualization.  
**Statistics:** None documented. No formal statistical methodology.  
**Maintenance:** Minimal. 17 total commits, no releases published, 269 stars. Appears to be a personal experiment.  
**Setup complexity:** Low to run; awkward to use in CI (requires clean git state at every commit; executables must be statically linked).  
**Verdict:** Not a fit for Go continuous benchmarking. Listed here to prevent confusion — the name is plausible but the tool is unrelated to Go benchmark tooling.

---

### 6. Nyrkiö / Apache Otava

**Home:** <https://nyrkio.com> · <https://github.com/nyrkio/nyrkio> · <https://github.com/apache/otava>  
**Go native:** Yes, via the `nyrkio/change-detection` GitHub Action with `tool: 'go'`. The action ingests `go test -bench` stdout, converts it to a common JSON format, and posts results to the Nyrkiö service (or self-hosted instance) for analysis.  
**Hosted / self-hosted:** Both. Nyrkiö.com is the hosted service. A full self-hosted Docker stack is documented. Apache Otava itself is a standalone CLI tool deployable anywhere.  
**Statistics:** This is the standout differentiator. Nyrkiö uses **Apache Otava (incubating)** as its change-detection backend. Otava implements the **e-divisive means** algorithm (Matteson & James 2014), the same lineage as the ED-PELT algorithm referenced in performance engineering literature. The algorithm:
- Finds statistically significant, persistent change-points in the full time-series history
- Adapts automatically to each benchmark's noise level
- Can detect shifts as small as 0.5% in noisy data
- Uses a configurable p-value (default 0.001) and a minimum percent-change threshold (default 5%)

Apache Otava entered the Apache Incubator on **2024-11-27**. Otava v0.8.0-incubating was released in 2025, with its core algorithm rewritten directly from the Matteson & James paper (a latent bug in all previous versions was found in the process). The Nyrkiö platform itself shipped **v2.0.0 on February 23, 2026** (introducing GitHub Runners for repeatable benchmarking). The `nyrkio/change-detection` action last released **v2.0.2 on July 17, 2025**.

**PR gate:** Yes. The action supports `fail-on-alert: true` and `comment-on-alert: true`. Slack and GitHub Issue notifications are also available.  
**Maintenance:** Active. Nyrkiö core repo had activity on **July 22, 2026** (the date of this research). 2,592 commits on the core repo; 68 stars.  
**Setup complexity:** Moderate. Requires a `nyrkio-token` (free account at nyrkio.com), one workflow step, and piping benchmark output to a file. The change-detection action handles the rest.

---

### 7. codespeed

**Home:** <https://github.com/tobami/codespeed>  
**Go native:** No. codespeed is a language-agnostic benchmark dashboard that accepts results via HTTP POST to a REST endpoint. Go integration requires writing a custom uploader script (e.g., parse `go test -bench` output and POST it).  
**Hosted / self-hosted:** Self-hosted only. Requires Django, a database (SQLite, PostgreSQL), and optionally nginx.  
**Statistics:** Basic visualization (time-series charts); no automated alerting or statistical significance. Known users include CPython, PyPy, and Twisted.  
**Maintenance:** **Unmaintained.** Latest release: **v0.13.0, February 23, 2019**. No commits since then. 602 stars, 133 forks — those numbers reflect historical interest, not current activity. Do not adopt for a new project.  
**Setup complexity:** High — requires provisioning a Django stack, which is a significant operational burden for a Go team.

---

### 8. golang.org/x/perf (benchstat / benchsave / perfdata)

**Home:** <https://pkg.go.dev/golang.org/x/perf>  
**Go native:** Yes — this is the official Go performance tooling.  
**Hosted / self-hosted:** The Go team runs <https://perf.golang.org> for internal Go project benchmarking. `benchsave` can upload results there, but it is not a public hosting service for third-party projects.  
**Statistics:** `benchstat` computes **median and bootstrap confidence intervals** across repeated runs, then tests for statistical significance when comparing two files. This is the gold standard for local A/B comparison. The `benchmath` package underlying it handles distributions correctly. However, `benchstat` is a point-in-time comparison tool, not a continuous monitoring tool — it does not maintain a time-series history or alert on regressions.

Key packages and tools:

| Name | Purpose | Status |
|---|---|---|
| `cmd/benchstat` | Statistical A/B comparison of benchmark files | Active |
| `cmd/benchsave` | Upload results to a perfdata server | Active |
| `cmd/benchfilter` | Filter benchmark output by name/tag | Active |
| `benchfmt` | Parse/write Go benchmark format | Active |
| `benchmath` | Statistical distributions over benchmark samples | Active |
| `benchproc` | Filter/group/sort benchmark results | Active |
| `storage` pkg | Client for perfdata server | **Deprecated** (moved to golang.org/x/build) |
| `analysis` pkg | Analysis server for perf.golang.org | **Deprecated** (moved to golang.org/x/build) |

**Maintenance:** Active. The module version as of this research is `v0.0.0-20260709024250-82a0b07e230d` (no stable semver tag; developed as a pseudo-versioned module).  
**Setup complexity:** Very low for `benchstat` as a local tool. Wiring it into CI for continuous tracking requires scripting around it — it has no native CI integration, PR gate, or time-series storage.

---

## Comparison Table

| Tool | Go native | Hosted/Self | Statistics | PR gate | Maintained | Setup complexity |
|---|---|---|---|---|---|---|
| **bencher.dev** | Yes (`go_bench` adapter) | Both | t-test, z-score, IQR, log-normal, percentage (no change-point) | First-class | Yes | Moderate |
| **github-action-benchmark** | Yes | GitHub Pages | Percentage threshold only | Yes | Yes (v1.22.1, May 2026) | Low–moderate |
| **gobenchdata** | Yes | GitHub Pages | User-defined expression threshold | Yes | Maintenance-mode (last: Jan 2023) | Low |
| **cob** | Yes | None (local/CI ephemeral) | Percentage threshold only | Via exit code | Borderline (last: Oct 2023, forks active) | Very low |
| **chronologer** | **No** (hyperfine, not `go test -bench`) | None (local HTML) | None | No | Minimal (no releases) | Low for local |
| **Nyrkiö / Apache Otava** | Yes (`nyrkio/change-detection` action) | Both | **E-divisive change-point detection** (Apache Otava) | Yes (`fail-on-alert`) | Yes (Nyrkiö 2.0.0, Feb 2026; action v2.0.2, Jul 2025) | Moderate |
| **codespeed** | No (requires custom uploader) | Self-hosted only | Visualization only, no alerts | No | **Unmaintained** (last: Feb 2019) | High (Django stack) |
| **golang.org/x/perf benchstat** | Yes (official) | perf.golang.org (Go team only) | Median + bootstrap CI, significance test | None built-in | Yes | Very low (local), high (CI wiring) |

---

## Recommendation

### Small OSS Go project

**Use `benchmark-action/github-action-benchmark`.**

It requires no external accounts, no tokens beyond `GITHUB_TOKEN`, and stores results directly on GitHub Pages. The latest release (v1.22.1, May 2026) is well-maintained. The 200% default threshold is conservative; tighten it to 10–20% for useful signal.

Accept the limitation: the percentage threshold has no concept of statistical significance. To compensate, run benchmarks with `-count=10` and add a manual `benchstat` comparison step in the workflow for human review. This gives the automated gate plus honest numbers in the PR comment.

### Team with a dedicated CI runner

**Use bencher.dev (self-hosted).**

The self-hosted binary is free and avoids the $100/month cloud cost. Running benchmarks on a dedicated, noise-controlled machine eliminates the biggest source of CI benchmark noise. Configure the `t_test` threshold model — it handles small sample sizes correctly and gives a principled false-positive rate. Wire `--error-on-alert` to block PRs.

The `go_bench` adapter requires zero changes to your benchmark code. The main operational cost is running the bencher self-hosted server, which ships as a single binary.

### Large org wanting change-point detection

**Use Nyrkiö with Apache Otava as the detection backend.**

The e-divisive algorithm is the right tool when:
- Benchmarks are noisy and a fixed percentage threshold would produce too many false positives
- You want to detect gradual regressions that accumulate over multiple commits
- You need the detection to adapt automatically to each benchmark's individual noise floor

Nyrkiö 2.0.0 (Feb 2026) adds GitHub Runners for reproducible benchmarking — important for removing environment noise from the measurement. The Apache Otava incubation gives the algorithm long-term governance and community maintenance.

For teams needing purely on-prem: run Apache Otava directly against your own time-series data store (CSV, PostgreSQL, BigQuery, Graphite are all supported). Nyrkiö is the hosted/integrated layer on top.

**Local/CI comparison complement (any scenario):** Always use `benchstat` from `golang.org/x/perf` for local A/B comparisons during development. It is the only tool here that gives you a confidence interval and a p-value on the comparison. Running `go test -bench=. -count=10` + `benchstat old.txt new.txt` before opening a PR catches regressions before they hit CI.

---

## Wire-Up Sketch

### Recommended for small OSS: github-action-benchmark

```yaml
# .github/workflows/benchmark.yml
name: Benchmark

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: write   # push to gh-pages
  checks: write     # set PR check status

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      # Run benchmarks; -count=6 gives benchstat enough samples for a
      # meaningful median even if you run benchstat manually later.
      - name: Run benchmarks
        run: go test -bench=. -benchmem -count=6 ./... | tee output.txt

      # Feed results to the action; compares against the stored baseline
      # on gh-pages and fails the check if any result is >15% slower.
      - name: Store and compare benchmarks
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: go
          output-file-path: output.txt
          github-token: ${{ secrets.GITHUB_TOKEN }}
          auto-push: true              # commit results to gh-pages
          alert-threshold: "115%"      # fail if > 15% slower
          fail-on-alert: true
          comment-on-alert: true
          # The gh-pages branch must exist as an orphan branch beforehand:
          # git checkout --orphan gh-pages && git commit --allow-empty -m "init"
          gh-pages-branch: gh-pages
          benchmark-data-dir-path: dev/bench
```

> **Pre-requisite:** create the `gh-pages` orphan branch once:
> ```bash
> git checkout --orphan gh-pages
> git reset --hard
> git commit --allow-empty -m "chore: init gh-pages for benchmarks"
> git push origin gh-pages
> git checkout main
> ```

---

### Recommended for large orgs: Nyrkiö change-detection action

```yaml
# .github/workflows/benchmark.yml
name: Benchmark

on:
  push:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest   # ideally: a dedicated, noise-controlled runner
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      # Use -count=10 so Nyrkiö / Otava has enough samples per run to
      # compute a meaningful per-run median before the time-series analysis.
      - name: Run benchmarks
        run: go test -bench=. -benchmem -count=10 ./... > bench-output.txt

      - name: Send results to Nyrkiö for change-point analysis
        uses: nyrkio/change-detection@v2
        with:
          tool: go
          output-file-path: bench-output.txt
          nyrkio-token: ${{ secrets.NYRKIO_TOKEN }}
          fail-on-alert: true
          comment-on-alert: true
          # Optional: only alert on changes that persist ≥ 3 commits
          # and exceed 5% (the Otava defaults).
          threshold: 5
```

> **Account setup:** create a free account at <https://nyrkio.com> and add the API token as `NYRKIO_TOKEN` in repository secrets.

---

## Key Takeaways

1. **No tool replaces `-count=10` and `benchstat`.** Every CI tool here compares runs, but if each run is a single sample, the comparison is comparing noise. Run with `-count=6` to `-count=10` minimum.

2. **Percentage thresholds lie on noisy machines.** `github-action-benchmark` and `cob` use simple percentage comparisons. On shared CI runners with variable load, they will generate false positives (flaky benchmark gates) or false negatives (real regressions absorbed by noise). Tighten them only on dedicated, noise-controlled runners.

3. **Change-point detection is the right model for continuous monitoring.** Nyrkiö/Apache Otava's e-divisive algorithm looks at the full history and finds persistent shifts — it ignores one-off noise spikes. This matches how real performance regressions actually appear in software (a commit lands, the metric shifts and stays shifted).

4. **bencher.dev is the best-featured threshold-based tool.** Its t-test and IQR models are statistically sounder than raw percentage comparisons, but they are still window-based: they compare a new result against a rolling window of past results, not against the full history. They do not detect gradual drift.

5. **golang.org/x/perf benchstat is the ground truth for local comparisons.** Use it in development. Use it in PR descriptions. Use it to validate whatever the CI gate is telling you. It is the only tool that shows you a confidence interval and a p-value.

6. **codespeed is dead. chronologer is not a Go tool.** Listed here to save you the detour.

---

## Sources

1. Bencher documentation — thresholds and statistical models: <https://bencher.dev/docs/explanation/thresholds/> (accessed 2026-07-22)
2. Bencher pricing: <https://bencher.dev/pricing/> (accessed 2026-07-22)
3. Bencher quick-start — Go adapter confirmation: <https://bencher.dev/docs/tutorial/quick-start/> (accessed 2026-07-22)
4. benchmark-action/github-action-benchmark README: <https://github.com/benchmark-action/github-action-benchmark> (accessed 2026-07-22)
5. bobheadxi/gobenchdata README: <https://github.com/bobheadxi/gobenchdata> (accessed 2026-07-22)
6. knqyf263/cob README: <https://github.com/knqyf263/cob> (accessed 2026-07-22)
7. dandavison/chronologer README: <https://github.com/dandavison/chronologer> (accessed 2026-07-22)
8. nyrkio/nyrkio GitHub organization: <https://github.com/nyrkio> (accessed 2026-07-22)
9. nyrkio/change-detection README: <https://github.com/nyrkio/change-detection> (accessed 2026-07-22)
10. Apache Otava — project home: <https://otava.apache.org/> (accessed 2026-07-22)
11. Apache Otava — GitHub repository: <https://github.com/apache/otava> (accessed 2026-07-22)
12. "8 Years of Optimizing Apache Otava" (arXiv preprint, 2025): <https://arxiv.org/pdf/2505.06758> (accessed 2026-07-22)
13. "Welcome Apache Otava (incubating project)" — Nyrkiö blog: <https://blog.nyrkio.com/2025/05/08/welcome-apache-otava-incubating-project/> (accessed 2026-07-22)
14. Apache Otava 0.8.0-incubating release announcement: <http://www.mail-archive.com/general@incubator.apache.org/msg86434.html> (accessed 2026-07-22)
15. tobami/codespeed README: <https://github.com/tobami/codespeed> (accessed 2026-07-22)
16. golang.org/x/perf package documentation: <https://pkg.go.dev/golang.org/x/perf> (accessed 2026-07-22)
17. Matteson & James (2014) — e-divisive algorithm paper: <https://www.researchgate.net/publication/239939788_A_Nonparametric_Approach_for_Multiple_Change_Point_Analysis_of_Multivariate_Data> (accessed 2026-07-22)
18. "Continuous benchmarking with Go and GitHub Actions" — DEV Community: <https://dev.to/vearutop/continuous-benchmarking-with-go-and-github-actions-41ok> (accessed 2026-07-22)
19. Apache Incubator — Otava status page: <https://incubator.apache.org/clutch/otava.html> (accessed 2026-07-22)

# About the Speaker — Evidence Ledger

Kemal Akkoyun · GopherCon UK 2026
Verified against public endpoints: 2026-08-10

Every number on the ethos slide resolves to a public URL in this file.
Re-run before each conference; GitHub search results drift as PRs accumulate.

---

## Tier 1 — Governance (lead fact)

**Prometheus Steering Committee member.**
Elected in the 2026 bootstrap election; seated 6 July 2026; term ends July 2027.

| Evidence | URL |
|---|---|
| `GOVERNANCE.md` committee table | `github.com/prometheus/governance/blob/main/GOVERNANCE.md` |
| Election results PR (merged 2026-06-30) | `github.com/prometheus/governance/pull/22` |
| Public Condorcet tally | `civs1.civs.us/cgi-bin/results.pl?id=E_aa350fd8aadf6dd2` |
| Self-nomination statement | `github.com/prometheus/governance/blob/main/elections/2026/candidate-kakkoyun.md` |

Context: all 7 seats were elected simultaneously in the bootstrap election; a company cap of 2
seats per organisation applied. He is one of 7 stewards of a CNCF-graduated project.

```bash
# Re-verify the committee table entry
curl -s https://raw.githubusercontent.com/prometheus/governance/main/GOVERNANCE.md \
  | grep -A3 "Kemal"
```

---

## Tier 2 — Current maintainer roles

| Project | Stars | Role | Evidence file |
|---|---|---|---|
| `prometheus/client_golang` | 6,019 | Maintainer (1 of 3) | `MAINTAINERS.md` |
| `prometheus/promu` | — | **Sole maintainer** | `MAINTAINERS.md` — only entry |
| `prometheus/test-infra` | — | Maintainer (1 of 2) | `MAINTAINERS.md` |
| `open-telemetry/opentelemetry-go-compile-instrumentation` | 401 | Maintainer + approver | GitHub team membership: `go-compile-instrumentation-maintainers` and `-approvers` |

`promu` is the Prometheus build and release tool — directly on-topic for the benchmarking deck.

```bash
# Verify public org membership
curl -s https://api.github.com/users/kakkoyun/orgs | jq '.[].login'
# Expected: prometheus, open-telemetry, thanos-io, prometheus-operator, observatorium, DataDog
```

---

## Tier 3 — Emeritus (state plainly; do not soften)

| Project | Stars | Evidence |
|---|---|---|
| `thanos-io/thanos` | 14,168 | Listed under `## Emeritus Maintainers` |
| `parca-dev/parca` | 4,937 | `### Emeritus Maintainers`; PR #5798 "Stepping down as a maintainer" (2025-07-10) |
| `parca-dev/parca-agent` | 740 | `### Emeritus Maintainers`; PR #3065 (2025-07-18); scope was "Everything (eBPF, profiler, debuginfo, CI/CD)" |

Stepping down cleanly, in public, with a PR, is itself a signal to a room full of maintainers.
Frame as a completed tour of duty, not a lapsed title.

---

## Tier 4 — The Zen of Prometheus arc

A six-year loop that closes cleanly:

1. **2020-09** — talk *The Zen of Prometheus*, PromCon Online 2020
   — `youtube.com/watch?v=Nqp4fjw_omU`
2. **2020-06** — repo `kakkoyun/the-zen-of-prometheus` (35★, his most-starred own repo)
3. **2026-03** — upstreamed into official Prometheus documentation via
   `prometheus/docs` PR #1783, merged by Jan Fajerski; now lives at
   `prometheus.io/docs/practices/the_zen/` as `docs/practices/the_zen.md`
4. **2026-05** — blog post *From Talk to Docs: The Zen of Prometheus* on `kakkoyun.me`

"I don't just give talks about this" — provable.

```bash
# Verify the doc is live
curl -s https://prometheus.io/docs/practices/the_zen/ | grep -c "Zen"
```

---

## Tier 5 — Volume

| Metric | Value | Re-derive |
|---|---|---|
| Merged PRs in Go repos | **963** | `https://github.com/search?q=author%3Akakkoyun+is%3Apr+is%3Amerged+language%3AGo&type=pullrequests` |
| Merged PRs, all repos | 1,456 of 1,707 opened | drop `language:Go` from above |
| PRs reviewed for others | **3,150** | `https://github.com/search?q=reviewed-by%3Akakkoyun+is%3Apr&type=pullrequests` |
| Issues opened | 569 | `q=author:kakkoyun+is:issue` |
| Followers | 727 | `GET /users/kakkoyun` |

Per-org merged PRs: `parca-dev` 474 · `thanos-io` 188 · `DataDog` 126 · `open-telemetry` 70 · `prometheus` 63.

Per-year merged: 2018:1 · 2019:142 · 2020:257 · 2021:109 · 2022:254 · 2023:149 · 2024:86 · 2025:153 · 2026:304.
Eight unbroken years — that consistency is the argument, more than any single total.

The 3,150 review count is the most underrated number. Review volume at 3× authorship is the
signature of a maintainer, not a drive-by contributor, and it cannot be farmed.

**Caveats:**
- `language:Go` filters on each repo's *current* primary language; the 963 figure drifts over time.
  Re-run the search before each conference.
- GitHub search paginates to 1,000 items, so per-repo breakdown is partial.
  Org-level counts are authoritative; per-repo numbers are illustrative.

---

## Tier 6 — The Go + eBPF arc

| Period | Project | Notes |
|---|---|---|
| 2018 | `MeltwaterArchive/drone-cache` (341★, Go) | Earliest public Go artifact; blog: *Making Drone Builds 10 Times Faster* (2019) |
| 2019–2022 | Thanos + Prometheus / `client_golang` | 188 + 63 merged PRs |
| 2021–2025 | Parca + parca-agent | 474 merged PRs; eBPF profiler, symbolization, debuginfo. parca-agent is C + Go — eBPF credibility rests here |
| 2021–2025 | Side projects | `py-perf` (Rust/eBPF Python profiler, 18★), `tiny-profiler` (Go/eBPF, 12★) |
| 2024–today | Datadog | `dd-trace-go` (92 merged), `orchestrion` (27); OTel `go-compile-instrumentation` (64) as maintainer |

**Important:** he has **zero** merged PRs in `opentelemetry-ebpf-profiler` or
`opentelemetry-ebpf-instrumentation`. The eBPF credibility comes from Parca, not from OBI.
The without-a-single-line deck must not imply OBI contributions.

Narrative arc: drone-cache → Thanos → Prometheus → Parca → compile-time instrumentation.
Build tooling → metrics → profiling → instrumentation. On-topic for both decks.

---

## Tier 7 — Talks (14 published, 2020–2026, each with a public recording)

Source: `kakkoyun/me` `content/talks/*.md`

| Talk | Event | Recording |
|---|---|---|
| Unleashing the Go Toolchain | GopherCon UK **2025** | `youtu.be/8Rw-fVEjihw` |
| How to Instrument Go Without Changing a Single Line of Code | FOSDEM 2026 Go devroom | `youtu.be/0TvrSebuDPk` |
| How to Reliably Measure Software Performance | FOSDEM 2026 Software Performance devroom | `youtu.be/8211fNI_nc4` |
| Building a Go Profiler Using Go | GopherCon EU 2022 | `youtu.be/OlHQ6gkwqyA` |
| Profiling Go Applications in the Cloud-Native Era | GopherCon Turkey 2021 | — |
| Building Observable Go Services | GopherCon Turkey 2020 | — |
| The Zen of Prometheus | PromCon Online 2020 | `youtube.com/watch?v=Nqp4fjw_omU` |

Plus: KubeCon NA 2020/2021, KubeCon EU 2022, PromCon EU 2022, FOSDEM 2020, GoDays Berlin 2020,
Cloud-Native eBPF Day EU 2022, PrometheusDay NA 2022.

Both 2026 GopherCon UK decks are the FOSDEM 2026 versions revisited — each deck can point at its
own earlier recording as the strongest "not my first pass at this material" signal.

---

## Tier 8 — Writing (24 posts on kakkoyun.me)

Selected posts:
- *Go Compile-Time Instrumentation v1* (2026-07)
- *Fantastic Symbols and Where to Find Them* parts 1–2 (2022)
- *Profiling Python and Ruby using eBPF* (2023)
- *Making Drone Builds 10 Times Faster* (2019)
- *Mentorship in Open Source* parts 1–3 (2026)
- *From Talk to Docs: The Zen of Prometheus* (2026)

---

## Verification checklist (run before each conference)

```bash
# 1. Steering Committee entry still present
curl -s https://raw.githubusercontent.com/prometheus/governance/main/GOVERNANCE.md \
  | grep "Kemal"

# 2. CIVS tally still public
curl -sI "https://civs1.civs.us/cgi-bin/results.pl?id=E_aa350fd8aadf6dd2" \
  | grep HTTP

# 3. Zen of Prometheus doc is live
curl -sI https://prometheus.io/docs/practices/the_zen/ | grep HTTP

# 4. PR count (re-run; number grows)
# Open in browser — GitHub search doesn't support unauthenticated CLI well:
# https://github.com/search?q=author%3Akakkoyun+is%3Apr+is%3Amerged+language%3AGo

# 5. Maintainer role in client_golang
curl -s https://raw.githubusercontent.com/prometheus/client_golang/main/MAINTAINERS.md \
  | grep -i "kemal"
```

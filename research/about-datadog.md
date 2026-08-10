# About Datadog — Go facts and context

> All facts in this file were verified 2026-08-10. Every row carries a
> primary-source link. Figures marked **⚠ drifting** are counts or stats that
> change over time (repository stars, CL counts, file counts); re-check them
> before quoting on a slide.

---

## 1. Go at Datadog

- Go is the most-used programming language at Datadog by a wide margin.
  _(Speaker statement; no public figure exists to cite.)_
- "Many of the Datadog backend services are written in Go." —
  <https://opensource.datadoghq.com/projects/go/>
- "We use Go, Python, Java, and React" —
  <https://careers.datadoghq.com/engineering/>
- **⚠ drifting** 390 Go-primary repos out of 1,199 public repos on
  github.com/DataDog (counted 2026-08-10).
- Datadog's public Go engineering guidelines:
  <https://datadoghq.dev/datadog-agent/guidelines/languages/go/>

---

## 2. Community and ecosystem

### OpenTelemetry Go compile-time instrumentation SIG (`otelc`)

- Founded by **Alibaba, Datadog, and Quesma** — three organisations, not two.
  <https://opentelemetry.io/blog/2025/go-compile-time-instrumentation/>
  (2025-01-24, OTel Governance Committee)
- Charter leads: @ralf0131 (Alibaba), @dineshg13 (Datadog), @pdelewski (Quesma).
  <https://github.com/open-telemetry/community/blob/main/projects/go-compile-instrumentation.md>
- Origin: donation proposal `open-telemetry/community#2497`, opened 2024-12-19
  by @RomainMuller (Datadog).
  <https://github.com/open-telemetry/community/issues/2497>
- **⚠ drifting** Repo: <https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation>
  — 401 stars, Apache-2.0. Maintainers: 7 total — 2 Datadog, 3 Alibaba, 1 Quesma, 1 Cabify.
- v1 announcement written by Kemal Akkoyun (Datadog):
  <https://opentelemetry.io/blog/2026/go-compile-time-instrumentation-v1/>
  (2026-07-16)

### opentelemetry-ebpf-profiler

- Felix Geisendörfer (Datadog) is 1 of 4 maintainers. The maintainer list is in
  `CONTRIBUTING.md`, not the README:
  <https://github.com/open-telemetry/opentelemetry-ebpf-profiler/blob/main/CONTRIBUTING.md>
- **⚠ drifting** Repo has 3,169 stars.

### Open source in the Go community

| Project | Stars | License | Latest | Notes |
|---|---|---|---|---|
| `go-profiler-notes` | **⚠** 3,666 | CC-BY-SA-4.0 | — | Docs, not code; community reference on Go profiling internals |
| Orchestrion | **⚠** 616 | Apache-2.0 | v1.12.0 (2026-07-30) | Compile-time auto-instrumentation |
| dd-trace-go | **⚠** 850 | Apache-2.0 OR BSD-3-Clause | v2.9.2 (2026-08-07) | Go APM tracer |
| datadog-agent | **⚠** 3,701 | Apache-2.0 | 7.82.0 (2026-08-05) | Main agent |

Sources:
- <https://github.com/DataDog/go-profiler-notes>
- <https://github.com/DataDog/orchestrion>
- <https://github.com/DataDog/dd-trace-go>
- <https://github.com/DataDog/datadog-agent>

**Talks by Datadog engineers** (listed on <https://opensource.datadoghq.com/projects/go/>; the
page gives no talk URLs):

- "Go Profiling and Observability from Scratch" — GopherCon 2021
- "Automatically Instrument Your Go Source Code with Orchestrion" — GopherCon 2023
- "Memory Optimization through Structure Packaging" — GopherCon Europe 2024, Diana Shevchenko

**Contributions beyond Datadog's own repos:**
`prometheus/client_golang#1942`, "feat(collector): add Go 1.26 new runtime metrics", merged
2026-02-11. <https://github.com/prometheus/client_golang/pull/1942>

---

## 3. Upstream contributions to Go

Reproducible query:
<https://go-review.googlesource.com/q/owner:*datadoghq.com+status:merged>

| Metric | Value |
|---|---|
| **⚠ drifting** Merged CLs across Go project repos | 115 |
| **⚠ drifting** …in `golang/go` itself | 108 |
| **⚠ drifting** Open / abandoned | 23 / 26 |
| **⚠ drifting** Distinct Datadog authors | 8 |
| First merged CL | 2021-04-27 (CL 299991, block profile bias) |

Subsystems within those 108: `runtime` 40+, `internal/trace` ~17,
`runtime/metrics` 10, `runtime/pprof` 9+, `runtime/trace` 6.
Top authors: Felix Geisendörfer 57, Nick Ripley 42.

**Re-run the query before quoting any of these counts.**

### Landed features

| Feature | Go version | Evidence |
|---|---|---|
| Execution tracer frame-pointer unwinding — "up to a 10x improvement over the previous release" | 1.21 | <https://go.dev/doc/go1.21> · CL 463835 (Geisendörfer) · CL 490815 (Ripley) |
| Goroutine creator IDs in stack traces | 1.21 | <https://go.dev/doc/go1.21> · CL 435337 (Ripley) |
| `runtime/pprof` label rewrite — 51–71% faster label operations | 1.24 | CL 574516 (Geisendörfer) |
| Clock snapshots in trace generations | 1.25 | CL 653575 (Geisendörfer) |
| Block profile bias fix | 1.17 | <https://github.com/golang/go/issues/44192> · CL 299991 · CL 324471 |
| CPU profiler accuracy + pprof label fixes | 1.18 | CL 351751 · CL 367200 · <https://www.datadoghq.com/blog/engineering/profiling-improvements-in-go-1-18/> |
| `runtime/metrics` additions (`/gc/gogc:percent`, `/gc/gomemlimit:bytes`, `/gc/scan/*`) | 1.21+ | <https://github.com/golang/go/issues/56857> · CLs 497315–497576 |
| `sched/latencies:seconds` fix (Nayef Ghattas) | — | CL 486755 |
| Anonymous memory mapping labels (Lénaïc Huard) | — | <https://github.com/golang/go/issues/71546> · CL 646095 |

CL URLs take the form `https://go-review.googlesource.com/c/go/+/<number>`.

Issues filed in `golang/go`: `felixge` 31, `nsrip-dd` 30, `RomainMuller` 5.

### The strongest single fact

Google's own flight-recording proposal credits the Go 1.21 tracer work as the thing that made
flight recording viable — <https://github.com/golang/go/issues/63185>:

> "This is also enabled by work in the Go 1.21 release to make traces dramatically cheaper. …
> Enabling flight recording across, for example, a small portion of a production fleet, becomes
> much more palatable when the tracing itself isn't too expensive."

Supporting detail from the engineer who did the work: stack unwinding was 94% of
execution-tracer overhead; `BenchmarkPingPongHog` overhead fell from +773.82% to +29.87%.
<https://blog.felixge.de/reducing-gos-execution-tracer-overhead-with-frame-pointer-unwinding/>
(2023-01-31)

### Compiler and build toolchain — the honest boundary

Datadog has landed **no changes in `cmd/compile`**. Toolchain CLs that did land are `cmd/go`
(2), `cmd/link` (2), `cmd/trace` (3), `cmd/asm` (1), `cmd/distpack` (2),
`cmd/internal/goobj` (1). The compile-time work is proposals and tooling built _on_ `-toolexec`,
not changes _to_ the compiler:

- **Proposal #69887**, `cmd/go: compile-time instrumentation and -toolexec`, filed 2024-10-15 by
  Romain Marcadier (Datadog), still open.
  <https://github.com/golang/go/issues/69887>
- **Companion #70046**, `cmd/go: support passing @args-file to -toolexec commands`, open.
- Orchestrion and otelc are how the limits of the existing toolchain get pushed from outside.

---

## 4. Running Go at scale

| Fact | Value | Source |
|---|---|---|
| Profile-guided optimization in production | "a noticeable drop of about 3.4 percent in CPU usage" | <https://www.datadoghq.com/blog/datadog-pgo-go/> |
| PGO benchmark range | −5.4% / −4.4% / −2.0% (max / median / min CPU profile) | same |
| Agent binary growth | 428 MiB → 1,248 MiB uncompressed, v7.16.0 → v7.60.0 (+192% over 5 years) | <https://www.datadoghq.com/blog/engineering/agent-go-binaries/> |
| Agent binary reduction | "reduced the size of our Agent Go binaries by up to 77%"; 20% of it from reducing reflection | same |
| Reflection fixes sent upstream | kubernetes/kubernetes, uber-go/dig, google/go-cmp | same |
| gRPC mesh | "tens of millions of requests per second between tens of thousands of pods" | <https://www.datadoghq.com/blog/grpc-at-datadog/> |
| Hot-path optimization | `NormalizeTag` 25% faster → 0.75% fleet CPU drop | <https://www.datadoghq.com/blog/engineering/self-optimizing-system/> |
| **⚠ drifting** datadog-agent scale | 11,460 `.go` files, ~71 MB of Go source, ~940 MB repo | GitHub trees API, 2026-08-10 |

---

## 5. Claims that do not hold up

These must never appear on a slide.

| Claim | Why not |
|---|---|
| Datadog contributes to `cmd/compile` | Zero merged CLs |
| Datadog filed or drove the flight recorder (#63185) | Filed by mknyszek (Google), implemented by Carlos Amedee (Google). The proposal _credits_ Datadog's tracer work — that is the defensible version |
| #69887 is about goroutine-local storage or `go:linkname` | It is `cmd/go: compile-time instrumentation and -toolexec` |
| "Frame pointer unwinding" appears in the Go 1.21 release notes | It does not. Only the "10x" outcome is stated; cite the CLs for the mechanism |
| Rhys Hiltner is a Datadog engineer | No company on his GitHub profile; he was at Twitch and collaborated with Felix Geisendörfer |
| Datadog is involved in OBI (ex-Beyla) | Zero Datadog people among its maintainers or approvers |
| The otelc SIG is "Datadog and Alibaba" | Quesma is a named founding lead |
| CL 443056 is the Go 1.22 mutex-profile scaling change | Different change, already in go1.20 |
| USDT probes upstream in Go | Not in the Go tree; that work lives in Datadog's agent |
| A lines-of-Go or Go-service count for Datadog | No public figure exists. Do not compute one from the internal monorepo |
| Datadog's numeric contributor rank in OpenTelemetry | devstats denies anonymous API queries; no citable number |
| Datadog sponsored GopherCon 2024 / sponsors GopherCon UK 2026 | First is aggregator-only, second is false — and sponsorship is out of scope anyway |

---

## 6. Sources

All URLs accessed 2026-08-10.

| URL | Used for |
|---|---|
| <https://opensource.datadoghq.com/projects/go/> | Go at Datadog statement, talk list |
| <https://careers.datadoghq.com/engineering/> | Language stack |
| <https://datadoghq.dev/datadog-agent/guidelines/languages/go/> | Go engineering guidelines |
| <https://opentelemetry.io/blog/2025/go-compile-time-instrumentation/> | otelc SIG founding (Alibaba + Datadog + Quesma) |
| <https://github.com/open-telemetry/community/blob/main/projects/go-compile-instrumentation.md> | Charter leads |
| <https://github.com/open-telemetry/community/issues/2497> | Donation proposal origin |
| <https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation> | otelc repo stats |
| <https://opentelemetry.io/blog/2026/go-compile-time-instrumentation-v1/> | v1 announcement |
| <https://github.com/open-telemetry/opentelemetry-ebpf-profiler/blob/main/CONTRIBUTING.md> | ebpf-profiler maintainers |
| <https://github.com/DataDog/go-profiler-notes> | go-profiler-notes repo |
| <https://github.com/DataDog/orchestrion> | Orchestrion repo |
| <https://github.com/DataDog/dd-trace-go> | dd-trace-go repo |
| <https://github.com/DataDog/datadog-agent> | datadog-agent repo |
| <https://github.com/prometheus/client_golang/pull/1942> | Go 1.26 runtime metrics contribution |
| <https://go-review.googlesource.com/q/owner:*datadoghq.com+status:merged> | CL counts (reproducible query) |
| <https://go.dev/doc/go1.21> | Go 1.21 release notes (tracer, goroutine creator IDs) |
| <https://github.com/golang/go/issues/44192> | Block profile bias issue |
| <https://www.datadoghq.com/blog/engineering/profiling-improvements-in-go-1-18/> | Go 1.18 profiling improvements |
| <https://github.com/golang/go/issues/56857> | runtime/metrics additions |
| <https://github.com/golang/go/issues/71546> | Anonymous memory mapping labels issue |
| <https://github.com/golang/go/issues/63185> | Flight recording proposal (credits Go 1.21 tracer) |
| <https://blog.felixge.de/reducing-gos-execution-tracer-overhead-with-frame-pointer-unwinding/> | Tracer overhead reduction detail |
| <https://github.com/golang/go/issues/69887> | cmd/go compile-time instrumentation proposal |
| <https://www.datadoghq.com/blog/datadog-pgo-go/> | PGO in production |
| <https://www.datadoghq.com/blog/engineering/agent-go-binaries/> | Agent binary size work |
| <https://www.datadoghq.com/blog/grpc-at-datadog/> | gRPC at scale |
| <https://www.datadoghq.com/blog/engineering/self-optimizing-system/> | NormalizeTag hot-path optimization |

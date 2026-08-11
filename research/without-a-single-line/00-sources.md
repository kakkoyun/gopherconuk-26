# Sources & Annotated Bibliography

> All load-bearing claims in this research must be pinned here.
> Format: `[S-NNN]` — used as cite keys in all other research docs and in `claims-ledger.md`.

## Format

```
### [S-NNN] Short title
- **URL:** <primary source URL>
- **Type:** official-repo | spec | blog-post | issue | rfc | paper | talk
- **Version/tag/SHA:** <pinned — never "latest">
- **Accessed:** YYYY-MM-DD
- **Notes:** what claim(s) this supports
```

---

## Projects — Official Repos & Docs

### [S-OBI-01] Grafana Beyla project page

- **URL:** <https://grafana.com/oss/beyla-ebpf/>
- **Type:** official-vendor-page
- **Version/tag/SHA:** accessed 2026-07-22
- **Accessed:** 2026-07-22
- **Notes:** Records Beyla's donation to OpenTelemetry as OBI. Supports C-010.

### [S-OBI-02] Grafana Beyla repository and donation notice

- **URL:** <https://github.com/grafana/beyla>
- **Type:** official-repo
- **Version/tag/SHA:** main, accessed 2026-07-22
- **Accessed:** 2026-07-22
- **Notes:** Confirms that Beyla continues as Grafana's distribution of upstream OBI. Supports C-010.

### [S-OBI-03] OBI module releases

- **URL:** <https://pkg.go.dev/go.opentelemetry.io/obi>
- **Type:** official-package-index
- **Version/tag/SHA:** v0.10.0, 2026-06-30
- **Accessed:** 2026-07-22
- **Notes:** Release and module identity. Supports C-013.

### [S-OBI-04] OpenTelemetry eBPF Instrumentation documentation

- **URL:** <https://opentelemetry.io/docs/zero-code/obi/>
- **Type:** official-docs
- **Version/tag/SHA:** accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Platform, architecture, kernel, BTF, signal, and maturity boundaries. Supports C-011 through C-014.

### [S-OBI-05] OBI Kubernetes setup

- **URL:** <https://opentelemetry.io/docs/zero-code/obi/setup/kubernetes/>
- **Type:** official-docs
- **Version/tag/SHA:** accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** DaemonSet requirements, host PID access, and capability examples.

### [S-OBI-06] OBI support matrix

- **URL:** <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/main/SUPPORT_MATRIX.md>
- **Type:** official-repo-doc
- **Version/tag/SHA:** main, accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Thirteen Go library baselines, including the gin v1.7.5 exclusion and go-redis v9 path. Supports C-011 and C-012.

### [S-OBI-07] OBI security, permissions, and capabilities

- **URL:** <https://opentelemetry.io/docs/zero-code/obi/security/>
- **Type:** official-docs
- **Version/tag/SHA:** accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Feature-dependent capability ladder and `perf_event_paranoid` caveats. Supports C-011.

### [S-OBI-08] OBI trace-log correlation

- **URL:** <https://opentelemetry.io/docs/zero-code/obi/trace-log-correlation/>
- **Type:** official-docs
- **Version/tag/SHA:** accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** JSON-only enrichment, trace/span fields, no log export, and feature requirements. Supports C-015.

## Go Runtime Issues & Proposals

<!-- #63185 flight recording, #69887 hook proposal, #67120 runtime/metrics, #75654 #38270 hook-gap bugs -->

## otelc / Orchestrion (compile-time)

### [S-O-01] Datadog open-source page: Orchestrion

- **URL:** <https://opensource.datadoghq.com/projects/orchestrion/>
- **Type:** official-vendor-page
- **Version/tag/SHA:** accessed 2026-07-22
- **Notes:** Confirms SIG co-founding with Alibaba; "built from scratch" language. Supports C-021 refutation.

### [S-O-02] golang/go#69887

- **URL:** <https://github.com/golang/go/issues/69887>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** open issue, accessed 2026-07-22
- **Notes:** Orchestrion maintainer comment explains -toolexec mechanism and AoP framing. Also confirms OTel SIG intent.

### [S-O-03] DataDog/orchestrion repo (HEAD df04ed94b69e, 2026-07-06)

- **URL:** <https://github.com/DataDog/orchestrion>
- **Type:** official-repo
- **Version/tag/SHA:** HEAD df04ed94b69e (2026-07-06); release v1.11.0 (2026-06-25)
- **Accessed:** 2026-07-22
- **Notes:** README vendor-agnostic statement; orchestrion.tool.go pattern; v1.11.0 release.

### [S-O-04] OTel blog: go-compile-time-instrumentation-v1 (2026-07-16)

- **URL:** <https://opentelemetry.io/blog/2026/go-compile-time-instrumentation-v1/>
- **Type:** blog-post (official OTel)
- **Version/tag/SHA:** published 2026-07-16
- **Accessed:** 2026-07-22
- **Notes:** Confirms otelc CLI name, v1.0.1 stable release, -toolexec mechanism, SIG launch.

### [S-O-05] open-telemetry/opentelemetry-go-compile-instrumentation repo

- **URL:** <https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation>
- **Type:** official-repo
- **Version/tag/SHA:** v1.0.1 (2026-07-14) — re-verify before citing
- **Accessed:** 2026-07-22
- **Notes:** otelc README, install command, `-toolexec` usage, and Go 1.25+ baseline.

### [S-O-06] DataDog/dd-trace-go contrib/supported_integrations.md

- **URL:** <https://github.com/DataDog/dd-trace-go>
- **Type:** official-repo
- **Version/tag/SHA:** v2 branch (re-verify SHA before citing count)
- **Accessed:** 2026-07-22
- **Notes:** Confirmed supported integrations list for orchestrion + dd-trace-go v2.

### [S-O-07] otelc supported libraries

- **URL:** <https://opentelemetry.io/docs/zero-code/go/compile-time/supported-libraries/>
- **Type:** official-docs
- **Version/tag/SHA:** accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Documents traces, HTTP/gRPC metrics, runtime metrics, slog/Logrus records, and current integration coverage. Supports C-023.

### [S-O-08] otelc testing strategy

- **URL:** <https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/blob/main/docs/testing.md>
- **Type:** official-repo-doc
- **Version/tag/SHA:** main, accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** CI coverage across Linux amd64/arm64, macOS arm64, and Windows amd64. Supports C-023.

### [S-O-09] OpenTelemetry Go compile-time SIG announcement

- **URL:** <https://opentelemetry.io/blog/2025/go-compile-time-instrumentation/>
- **Type:** official-blog
- **Version/tag/SHA:** published 2025-01-24
- **Accessed:** 2026-08-10
- **Notes:** Names Alibaba, Datadog, and Quesma as the founding organizations. Supports C-025.

### [S-DD-01] Speaker confirmation of public adoption claim

- **URL:** N/A
- **Type:** speaker-attestation
- **Version/tag/SHA:** confirmed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Approves the statement that customer adoption of Orchestrion-based auto-instrumentation grew by about 20%. No denominator or customer names are cleared. Supports C-026.

## OTel Ecosystem

### [S-I-01] OpenTelemetry Injector repository

- **URL:** <https://github.com/open-telemetry/opentelemetry-injector>
- **Type:** official-repo
- **Version/tag/SHA:** main, accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Packages Java, Node.js, .NET, and Python auto-instrumentation. It does not package Go. Supports C-024.

### [S-I-02] OpenTelemetry Operator auto-instrumentation

- **URL:** <https://opentelemetry.io/docs/platforms/kubernetes/operator/automatic/>
- **Type:** official-docs
- **Version/tag/SHA:** accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Operator Go support is a separate eBPF sidecar and is disabled by default behind `enable-go-instrumentation`. Supports C-024.

## opentelemetry-ebpf-profiler

### [S-P-01] OTel blog: Elastic contributes continuous profiling agent

- **URL:** <https://opentelemetry.io/blog/2024/elastic-contributes-continuous-profiling-agent/>
- **Type:** blog-post (official OTel)
- **Version/tag/SHA:** published June 2024
- **Accessed:** 2026-07-22
- **Notes:** Confirms donation completion. Cross-links community issue #1918.

### [S-P-02] CNCF blog: OTel announces support for profiling

- **URL:** <https://www.cncf.io/blog/2024/03/19/opentelemetry-announces-support-for-profiling/>
- **Type:** blog-post (official CNCF)
- **Version/tag/SHA:** published 2024-03-19
- **Accessed:** 2026-07-22
- **Notes:** Donation pledge announcement (forward-looking at time of publish). Supports C-030.

### [S-P-03] OTel blog: profiles-alpha (2026)

- **URL:** <https://opentelemetry.io/blog/2026/profiles-alpha/>
- **Type:** blog-post (official OTel)
- **Version/tag/SHA:** published 2026 (exact date: re-verify)
- **Accessed:** 2026-07-22
- **Notes:** Public Alpha announcement. Confirms otelcol-ebpf-profiler Collector distribution.

### [S-P-04] OTel community issue #1918 (donation tracking)

- **URL:** <https://github.com/open-telemetry/community/issues/1918>
- **Type:** issue (official OTel)
- **Version/tag/SHA:** closed June 2024
- **Accessed:** 2026-07-22
- **Notes:** Primary donation transfer record. Supports C-030.

### [S-P-05] opentelemetry-ebpf-profiler README + internals.md

- **URL:** <https://github.com/open-telemetry/opentelemetry-ebpf-profiler>
- **Type:** official-repo
- **Version/tag/SHA:** v0.0.202627 (re-verify at time of use)
- **Accessed:** 2026-07-22
- **Notes:** Architecture, zero-code-changes claim, .eh_frame unwinding, root/CAP requirements.

### [S-P-06] opentelemetry-ebpf-profiler doc/gopclntab.md

- **URL:** <https://github.com/open-telemetry/opentelemetry-ebpf-profiler/blob/main/doc/gopclntab.md>
- **Type:** official-repo-doc
- **Version/tag/SHA:** main branch (pin SHA before citing in talk/blog)
- **Accessed:** 2026-07-22
- **Notes:** Explains .gopclntab-based Go stack unwinding. Key quote about stripped production binaries.

### [S-P-07] OTel specification: profiles signal

- **URL:** <https://opentelemetry.io/docs/specs/otel/profiles/>
- **Type:** spec (official OTel)
- **Version/tag/SHA:** Status: Alpha at time of access
- **Accessed:** 2026-07-22
- **Notes:** OTel spec stability for profiles = Alpha. Supports C-031.

### [S-P-08] OTLP 1.11.0 specification

- **URL:** <https://opentelemetry.io/docs/specs/otlp/>
- **Type:** spec (official OTel)
- **Version/tag/SHA:** OTLP 1.11.0
- **Accessed:** 2026-07-22
- **Notes:** Profiles signal = Development tier (lowest). Traces/metrics/logs = Stable. Supports C-031.

### [S-P-09] OTEP 4947: thread context sharing

- **URL:** <https://github.com/open-telemetry/opentelemetry-specification/blob/main/oteps/profiles/4947-thread-ctx.md#alternative-for-go-support>
- **Type:** proposal (official OTel)
- **Version/tag/SHA:** main, accessed 2026-08-10
- **Accessed:** 2026-08-10
- **Notes:** Excludes Go from the primary TLSDESC path for the foreseeable future and defines the `go_pprof_labels_v1` alternative. Supports C-037.

## eBPF / USDT / Kernel

### [S-UP-01] Performance and reliability pitfalls of eBPF

- **URL:** <https://fosdem.org/2026/schedule/event/H3LM7G-performance_and_reliability_pitfalls_of_ebpf/>
- **Type:** talk
- **Version/tag/SHA:** FOSDEM 2026
- **Accessed:** 2026-08-10
- **Notes:** Usama Saqib's public discussion of eBPF and probe costs. Supports C-016.

### [S-UP-02] RCU-protected uprobe hot-path optimizations

- **URL:** <https://lists.openwall.net/linux-kernel/2024/08/13/142>
- **Type:** kernel-mailing-list
- **Version/tag/SHA:** patch series v3, 2024-08-12
- **Accessed:** 2026-08-10
- **Notes:** Andrii Nakryiko's lockless SRCU lookup and consumer traversal series. Supports C-016.

### [S-UP-03] OBI Go 1.26 symbol-resolution fix

- **URL:** <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/pull/1851>
- **Type:** official-repo-pull-request
- **Version/tag/SHA:** merged for OBI v0.9.0
- **Accessed:** 2026-08-10
- **Notes:** Go 1.26 removed `pcHeader.textStart`; OBI moved resolution to `runtime.moduledata.text`. Supports C-016.

## Benchmarks & Performance Data

## Thesis / Go Runtime Architecture

### [S-TH-01] runtime/proc.go "hall of shame" comment (Go master)

- **URL:** <https://go.dev/src/runtime/proc.go>
- **Type:** official-repo (Go stdlib)
- **Version/tag/SHA:** Go master (pin SHA before citing: `go version` + `go env GOROOT`)
- **Accessed:** 2026-07-22
- **Notes:** Contains verbatim "hall of shame" comment on go:linkname abuse of gopark/goready. Supports C-001.

### [S-TH-02] internal/runtime/exithook/hooks.go

- **URL:** internal package — <https://cs.opensource.google/go/go/+/main:src/internal/runtime/exithook/hooks.go>
- **Type:** official-repo (Go stdlib)
- **Version/tag/SHA:** Go master (pin at verification time)
- **Accessed:** 2026-07-22
- **Notes:** Confirms exit-time hook scope is termination-only. Supports C-001.

### [S-TH-03] golang/go#28909 (LD_PRELOAD requires external linking)

- **URL:** <https://github.com/golang/go/issues/28909>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** accessed 2026-07-22
- **Notes:** Ian Lance Taylor: "anyone who wants to use LD_PRELOAD must force the use of external linking." Supports C-001.

### [S-TH-04] Dynatrace Go known limitations

- **URL:** <https://docs.dynatrace.com/docs/ingest-from/technology-support/application-software/go/support/go-known-limitations>
- **Type:** vendor-docs
- **Version/tag/SHA:** accessed 2026-07-22
- **Notes:** Static cgo rejection; CGO_ENABLED=0+PIE incompatibility. Supports C-001.

### [S-TH-05a] dd-trace-go internal/orchestrion/gls.orchestrion.yml (Orchestrion aspect — the injector config)

- **URL:** <https://github.com/DataDog/dd-trace-go/blob/main/internal/orchestrion/gls.orchestrion.yml>
- **Type:** official-repo
- **Version/tag/SHA:** commit b97e7cbb4 ("v2.0.0"), c9ff7bbe2 — verified locally
- **Accessed:** 2026-07-22
- **Notes:** YAML aspect targeting `struct-definition: runtime.g`; `add-struct-field` injects `__dd_gls_v2 any`; injects go:linkname get/set via template; patches goexit1. This is the PRIMARY source for C-002.

### [S-TH-05b] dd-trace-go internal/orchestrion/gls.go (the consumer / dd-trace-go side)

- **URL:** <https://github.com/DataDog/dd-trace-go/blob/main/internal/orchestrion/gls.go>
- **Type:** official-repo
- **Version/tag/SHA:** commits b97e7cbb4 ("v2.0.0"), 577c7760f ("ddtrace/tracer: setup GLS-stored context") — verified locally
- **Accessed:** 2026-07-22
- **Notes:** go:linkname variable symbols __dd_orchestrion_gls_get.V2 /__dd_orchestrion_gls_set.V2 with nil defaults; init() guard. Supports C-002.

### [S-TH-06] golang/go#72032 (go:linkname variable fragility)

- **URL:** <https://github.com/golang/go/issues/72032>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** accessed 2026-07-22
- **Notes:** BSS symbol resolution order-dependent; Go 1.22→1.23 breaking change. Supports C-002b.

### [S-TH-07] dd-trace-go contrib/ directory count

- **URL:** <https://github.com/DataDog/dd-trace-go>
- **Type:** official-repo
- **Version/tag/SHA:** v2 branch HEAD (re-count at exact release tag before citing)
- **Accessed:** 2026-07-22
- **Notes:** 51 subdirectories confirmed. Supports C-003.

### [S-TH-08] kakkoyun.me FOSDEM 2026 talk summary

- **URL:** <https://kakkoyun.me/posts/fosdem-2026-auto-instrumenting-go/>
- **Type:** blog-post (speaker's own)
- **Version/tag/SHA:** published 2026-02-xx (verify date)
- **Accessed:** 2026-07-22
- **Notes:** Corroborates "no intermediate bytecode" framing; also references poc_usdt fork.

## Go Runtime Futures

### [S-RF-01] golang/go#63185 (flight recording)

- **URL:** <https://github.com/golang/go/issues/63185>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** CLOSED, milestoned Go 1.25
- **Accessed:** 2026-07-22
- **Notes:** Proposal accepted; JFR-style circular-buffer. Supports C-040.

### [S-RF-02] go.dev blog: flight-recorder (Go 1.25, Sept 2025)

- **URL:** <https://go.dev/blog/flight-recorder>
- **Type:** blog-post (official Go)
- **Version/tag/SHA:** published 26 Sept 2025
- **Accessed:** 2026-07-22
- **Notes:** Confirms Go 1.25 availability; FlightRecorderConfig struct API. Supports C-040.

### [S-RF-03] runtime/trace/flightrecorder.go (Go master)

- **URL:** <https://raw.githubusercontent.com/golang/go/master/src/runtime/trace/flightrecorder.go>
- **Type:** official-repo (Go stdlib)
- **Version/tag/SHA:** master (pin Go 1.25 tag before citing in talk)
- **Accessed:** 2026-07-22
- **Notes:** Source-of-truth for shipped API shape. Supports C-040.

### [S-RF-04] golang/go#69887 (toolexec improvements)

- **URL:** <https://github.com/golang/go/issues/69887>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** OPEN, Proposals Incoming
- **Accessed:** 2026-07-22
- **Notes:** Filed by RomainMuller (DataDog). Supports C-041.

### [S-RF-05] golang/go#75654 (httptrace response-end)

- **URL:** <https://github.com/golang/go/issues/75654>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** OPEN, active
- **Accessed:** 2026-07-22
- **Notes:** Proposes GotResponseEnd hook; HTTP/2 span closure bug. Supports C-042.

### [S-RF-06] opentelemetry-go-contrib#4876 (HTTP/2 span never closes)

- **URL:** <https://github.com/open-telemetry/opentelemetry-go-contrib/issues/4876>
- **Type:** issue (OTel)
- **Version/tag/SHA:** OPEN
- **Accessed:** 2026-07-22
- **Notes:** Live production impact of #75654 gap. Supports C-042.

## USDT

### [S-USDT-01] ebpf.io USDT concepts docs

- **URL:** <https://docs.ebpf.io/linux/concepts/usdt/>
- **Type:** spec/docs
- **Version/tag/SHA:** accessed 2026-07-22
- **Notes:** USDT = NOP → INT3 patch via uprobe path. .note.stapsdt format.

### [S-USDT-02] SystemTap UserSpaceProbeImplementation spec

- **URL:** <https://sourceware.org/systemtap/wiki/UserSpaceProbeImplementation>
- **Type:** spec (normative)
- **Version/tag/SHA:** last edited 2021-10-17
- **Accessed:** 2026-07-22
- **Notes:** .note.stapsdt ELF section structure, link-time PC address, GAS operands.

### [S-USDT-03] bpftime paper (USENIX ATC 2024)

- **URL:** <https://arxiv.org/pdf/2311.07923>
- **Type:** paper (peer-reviewed)
- **Version/tag/SHA:** USENIX ATC 2024
- **Accessed:** 2026-07-22
- **Notes:** 314 ns (bpftime userspace) vs 3224 ns (kernel uprobe) benchmark. ~10x improvement.

### [S-USDT-04] Kernel mailing list PATCHv6 NOP5 uprobe (Andrii Nakryiko, Sept 2025)

- **URL:** <https://www.mail-archive.com/linux-trace-kernel@vger.kernel.org/msg12197.html>
- **Type:** kernel-mailing-list
- **Version/tag/SHA:** PATCHv6, September 2025
- **Accessed:** 2026-07-22
- **Notes:** Proposes NOP5-based uprobe syscall competitive with USDT NOP1 path.

### [S-USDT-05] github.com/linux-usdt/libstapsdt

- **URL:** <https://github.com/linux-usdt/libstapsdt>
- **Type:** official-repo
- **Version/tag/SHA:** checked 2026-07-22 (pin SHA before citing)
- **Accessed:** 2026-07-22
- **Notes:** Runtime .so generation + dlopen for runtime USDT probes. Lists Go wrappers including salp.

### [S-USDT-06] github.com/mmcshane/salp (correct URL — not mmcloughlin)

- **URL:** <https://github.com/mmcshane/salp>
- **Type:** official-repo
- **Version/tag/SHA:** checked 2026-07-22
- **Accessed:** 2026-07-22
- **Notes:** Go CGo binding to libstapsdt. Note: mmcloughlin/salp returns 404.

### [S-USDT-07] github.com/libbpf/usdt

- **URL:** <https://github.com/libbpf/usdt>
- **Type:** official-repo
- **Version/tag/SHA:** checked 2026-07-22
- **Accessed:** 2026-07-22
- **Notes:** C-only single-header USDT macro library. No Go bindings.

### [S-USDT-08] golang/go#57175 (USDT in Go runtime)

- **URL:** <https://github.com/golang/go/issues/57175>
- **Type:** issue (golang/go)
- **Version/tag/SHA:** OPEN, initial-inquiry as of Dec 2024
- **Accessed:** 2026-07-22
- **Notes:** Go team at initial-inquiry stage. No USDT in stdlib.

### [S-USDT-09] kakkoyun.me FOSDEM 2026 talk summary (poc_usdt reference)

- **URL:** <https://kakkoyun.me/posts/fosdem-2026-auto-instrumenting-go/>
- **Type:** blog-post (speaker's own)
- **Version/tag/SHA:** published 2026-02-xx
- **Accessed:** 2026-07-22
- **Notes:** References poc_usdt fork (github.com/kakkoyun/go/tree/poc_usdt) adding USDT to net/http, database/sql, crypto/tls, net via `go tool usdt`.

## Prior Work (leads to re-verify — not authoritative until confirmed)

### [S-LEAD-01] Prior talk proposal (GopherCon EU/US)

- **Source:** personal vault `qmd://personal/creation/proposal/talks/GopherCon-EU-and-US-Go-Auto-Instrumentation.md`
- **Type:** prior-art-lead
- **Notes:** Earlier version of this talk. dd-trace-go ≈54 wrappers claim, `-toolexec` AST rewriting, Orchestrion, flight recording #63185. Re-verify every claim before citing.

### [S-LEAD-02] ctx session dde33599 — thesis/narrative research

- **Source:** ctx session `dde33599` (2026-07-16)
- **Type:** prior-art-lead
- **Notes:** go:linkname g-struct hack in `internal/orchestrion/gls.go`, hook-gap bugs #75654/#38270, proposal #69887. Re-verify every claim before citing.

### [S-LEAD-03] OTel Go compile-time instrumentation v1 blog draft

- **Source:** ctx sessions `f5cf45fa`, `19c1f8e1` (2026-07-03–07-15)
- **Type:** prior-art-lead
- **Notes:** v1 announcement blog; v0.5.0 was latest as of last check; "otelc" framing; OBI framing. Re-verify current release status.

### [S-LEAD-04] FOSDEM-26 Go USDT notes

- **Source:** `qmd://personal/creation/proposal/talks/FOSDEM-26-Go-USDT.md`
- **Type:** prior-art-lead
- **Notes:** Salp (older, Go 1.22 only), libbpf/usdt, libstapsdt, Polar Signals USDT deep-dive claim "uprobes as fast as USDTs". Re-verify all.

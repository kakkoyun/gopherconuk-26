# Claims Ledger

> Every load-bearing claim in this research corpus must have an entry here.
> A claim is "load-bearing" if it would change the talk's argument or a blog post's conclusion if wrong.
>
> **Verdicts:**
>
> - `CONFIRMED` — verified against a primary source (repo, spec, issue, official blog); cite `[S-NNN]`
> - `PLAUSIBLE` — consistent with evidence but not directly confirmed; must be flagged in prose, never asserted
> - `REFUTED` — found to be wrong; do not use; note what is correct instead
> - `PENDING` — not yet researched; placeholder only

## Format

```
### [C-NNN] The claim (exact wording to be used in talk/blog)
- **Verdict:** CONFIRMED | PLAUSIBLE | REFUTED | PENDING
- **Source(s):** [S-NNN], [S-NNN]
- **Notes:** verification method, caveats, version constraints
```

---

## Thesis / Go Runtime Hook Point

### [C-001] Go has no runtime hook point — you cannot monkey-patch a running Go process the way you can with JVM agents or Python's import hooks

- **Verdict:** CONFIRMED
- **Source(s):** S-TH-01 (runtime/proc.go "hall of shame"), S-TH-02 (internal/runtime/exithook), S-TH-03 (golang/go#28909 LD_PRELOAD), S-TH-04 (Dynatrace limitations)
- **Notes:** Native machine code, no classloader, no bytecode, static linking by default. Only internal exit-time hook is scoped to program termination. LD_PRELOAD requires -linkmode=external (non-default) and blocked entirely for CGO_ENABLED=0+PIE. Go team labels go:linkname abusers a "hall of shame" in runtime/proc.go.

### [C-002] Orchestrion injects a field `__dd_gls_v2 any` into `runtime.g` and exposes it via `go:linkname` get/set function variables

- **Verdict:** CONFIRMED ✅
- **Source(s):** S-TH-05a (gls.orchestrion.yml, commit b97e7cbb4), S-TH-05b (gls.go, commits b97e7cbb4 + 577c7760f)
- **Notes:** Complete mechanism (three steps):
  1. **Aspect** `gls.orchestrion.yml` (`internal/orchestrion/gls.orchestrion.yml` in dd-trace-go, commit b97e7cbb4) targets `struct-definition: runtime.g`, uses `add-struct-field` to inject `__dd_gls_v2 any` into the `g` struct.
  2. **Injects** `go:linkname` function-variable symbols: `__dd_orchestrion_gls_get.V2` = `func() any { return getg().m.curg.__dd_gls_v2 }` and `__dd_orchestrion_gls_set.V2` = `func(val any) { getg().m.curg.__dd_gls_v2 = val }`.
  3. **Patches** `runtime.goexit1` with `getg().__dd_gls_v2 = nil` to zero the field when a goroutine exits (prevents memory leaks).
  - dd-trace-go's `gls.go` then links to the symbols via go:linkname with no-op defaults when Orchestrion is absent.
  - The SHA "5fef9c1ba9ac9ff0ac966714d0342f7f85dbc4fa" does NOT exist — use b97e7cbb4 (v2.0.0) for the yml file.

### [C-002b] go:linkname for variables is fragile — BSS resolution is order-dependent; broke between Go 1.22 and 1.23

- **Verdict:** CONFIRMED
- **Source(s):** S-TH-06 (golang/go#72032)
- **Notes:** Unlike function linkname (which supports bodiless forward declarations), variable linkname has no definition/reference separation. If both sides are uninitialized, linker choice is arbitrary. Manifested as breaking change Go 1.22 → 1.23. Go team documented in #72032.

### [C-003] dd-trace-go has 51 integration wrappers in contrib/ (not ~54)

- **Verdict:** CONFIRMED (corrected)
- **Source(s):** S-TH-07 (dd-trace-go contrib/ directory, counted from v2 branch)
- **Notes:** Prior lead said ~54 — wrong. Actual count: 51 subdirs in contrib/. Use "over 50" or link to supported_integrations.md rather than a specific number that will drift.

---

## OBI

### [C-010] OBI (OpenTelemetry eBPF Instrumentation) is the renamed/donated successor to Grafana Beyla

- **Verdict:** CONFIRMED
- **Source(s):** S-OBI-01 (Grafana Beyla OSS page), S-OBI-02 (Beyla README + community issue #2406)
- **Notes:** Grafana Labs donated Beyla to CNCF/OTel in 2025 via community issue #2406. Beyla continues as Grafana's distribution. New repo: github.com/open-telemetry/opentelemetry-ebpf-instrumentation, module: go.opentelemetry.io/obi.

### [C-011] OBI requires supported Linux kernels with BTF; the required capabilities depend on the enabled features

- **Verdict:** CONFIRMED
- **Source(s):** S-OBI-04, S-OBI-06, S-OBI-07
- **Notes:** Baseline platform support is Linux 5.8+ with BTF, or documented RHEL-family 4.18+ backports. Network flow capture, application observability, context propagation, and Go library propagation require different capabilities. Do not publish one fixed list as universal. `CAP_SYS_ADMIN` may be required for Go propagation or restrictive `perf_event_paranoid` settings.

### [C-012] OBI is language-agnostic at network-protocol level; also provides Go-specific library uprobes for 13 libraries

- **Verdict:** CONFIRMED
- **Source(s):** S-OBI-04, S-OBI-06
- **Notes:** Language-agnostic: Go, Java (JDK 8+), .NET, Node.js, Python, Ruby, C, C++, Rust, GenAI SDKs. Go-specific uprobe support for 13 named libraries (net/http, gin, gRPC, gorilla/mux, go-redis, Kafka, database/sql, etc.). See SUPPORT_MATRIX.md v0.10.0 for complete list.

### [C-013] OBI current release: v0.10.0 (2026-06-30), Development status (breaking changes expected)

- **Verdict:** CONFIRMED (at time of research, 2026-07-22)
- **Source(s):** S-OBI-03 (pkg.go.dev), S-OBI-04 (README)
- **Notes:** README: "OBI is currently in Development. Users should expect breaking changes between minor releases while the project remains in v0." No v1.x release exists. Re-verify before talk/blog.

### [C-014] "Zero code changes" holds for HTTP/gRPC RED metrics and 13 Go library spans; does NOT hold for custom spans, business-logic events, SQL query details

- **Verdict:** CONFIRMED
- **Source(s):** S-OBI-04 (official docs caveat)
- **Notes:** Official docs explicitly: "Use language agents or manual instrumentation when you need custom spans, application-specific attributes, business events, or other in-process telemetry." Be precise in the talk: "zero changes for standard observability."

### [C-015] OBI can correlate JSON logs with traces by enriching writes with trace and span IDs, but it does not export logs

- **Verdict:** CONFIRMED
- **Source(s):** S-OBI-08
- **Notes:** Enrichment applies to selected JSON logs while a span is active. Plain text is unchanged. Existing log forwarding remains responsible for export. Current docs require Linux 6.0+ for this feature.

### [C-016] Uprobe instrumentation has infrastructure, BPF-program, and compatibility costs even though recent kernels improved hot-path scalability

- **Verdict:** CONFIRMED
- **Source(s):** S-UP-01, S-UP-02, S-UP-03
- **Notes:** Attached uprobes cross into the kernel. RCU-protected lookup and consumer traversal removed major scalability bottlenecks in newer kernels. A BPF program can still serialize a concurrent path through shared state. Go 1.26's removal of `pcHeader.textStart` required OBI to change symbol resolution; that issue is fixed and should be presented as maintenance evidence, not a current defect.

---

## otelc (OTel Go compile-time instrumentation)

### [C-020] otelc uses `-toolexec` to rewrite Go AST at compile time (not at link time, not at runtime)

- **Verdict:** CONFIRMED
- **Source(s):** S-O-03 (orchestrion repo + golang/go#69887), S-O-04 (OTel blog), S-O-05 (otelc repo)
- **Notes:** Both DataDog/orchestrion and otelc intercept `go tool compile` via `-toolexec`, rewrite .go source ASTs (using github.com/dave/dst in Orchestrion) before the compiler sees them. Orchestrion maintainer quote in #69887 confirms mechanism verbatim.

### [C-021] Orchestrion was donated to OTel as otelc

- **Verdict:** REFUTED — do not use
- **Source(s):** S-O-01, S-O-04
- **Notes:** Orchestrion was NOT donated. DataDog + Alibaba co-founded OTel Go Compile-Time Instrumentation SIG and built `otelc` (open-telemetry/opentelemetry-go-compile-instrumentation) from scratch, inspired by Orchestrion. Two separate tools with same mechanism, different codebases and default tracers.

### [C-021b] DataDog/orchestrion (CLI: `orchestrion`) and OTel SIG's otelc (CLI: `otelc`) are two distinct tools using the same -toolexec + AST rewriting mechanism

- **Verdict:** CONFIRMED
- **Source(s):** S-O-01, S-O-03, S-O-04, S-O-05
- **Notes:** orchestrion: GA v1.11.0 (2026-06-25), defaults to dd-trace-go/v2 but vendor-agnostic. otelc: stable v1.0.1 (2026-07-14), defaults to OTel SDK. Both intercept go tool compile via -toolexec.

### [C-022] otelc current release version: v1.0.1 (2026-07-14), first stable non-retracted release

- **Verdict:** CONFIRMED (at time of research, 2026-07-22)
- **Source(s):** S-O-04 (OTel blog), S-O-05 (repo)
- **Notes:** Context lead said v0.5.0 — this was outdated. v1.0.1 shipped 2026-07-14 (barely a week before this research). Re-verify at talk/blog publish time. orchestrion is separately at v1.11.0 (2026-06-25).

### [C-023] otelc is a production-capable, cross-platform build-time path rather than a local-development-only tool

- **Verdict:** CONFIRMED
- **Source(s):** S-O-04, S-O-07, S-O-08
- **Notes:** Official docs describe fleet-wide build-pipeline use. The repository builds/tests across Linux, macOS, and Windows. Current supported integrations produce traces, HTTP/gRPC metrics, runtime metrics, and supported slog/Logrus records. Runtime behavior still depends on the selected SDK and integration bundle.

### [C-024] The OpenTelemetry host injector does not package Go; the Operator's Go option is a separate feature-gated eBPF sidecar

- **Verdict:** CONFIRMED
- **Source(s):** S-I-01, S-I-02
- **Notes:** Keep the host injector and Kubernetes Operator mechanisms distinct in the talk.

### [C-025] Alibaba, Datadog, and Quesma founded the OpenTelemetry Go compile-time instrumentation SIG

- **Verdict:** CONFIRMED
- **Source(s):** S-O-09
- **Notes:** The three organizations contributed independent experience. Do not describe Orchestrion as renamed or donated wholesale into otelc.

### [C-026] Customer adoption of Orchestrion-based Go auto-instrumentation grew by about 20%

- **Verdict:** CONFIRMED by the speaker and approved for public use
- **Source(s):** S-DD-01
- **Notes:** Use only the approved percentage. Do not publish a denominator, customer names, or the earlier "mostly new users" interpretation without separate clearance.

---

## opentelemetry-ebpf-profiler

### [C-030] opentelemetry-ebpf-profiler was donated to OTel by Elastic (formerly Elastic Universal Profiling)

- **Verdict:** CONFIRMED
- **Source(s):** S-P-01 (OTel blog June 2024), S-P-02 (CNCF blog March 2024), S-P-04 (community issue #1918)
- **Notes:** Pledged March 2024, transfer completed June 2024 via community issue #1918.

### [C-031] The OTel profiling signal uses OTLP profiles (not a vendor format)

- **Verdict:** CONFIRMED (with important nuance)
- **Source(s):** S-P-07 (OTel spec), S-P-08 (OTLP 1.11.0)
- **Notes:** OTLP profiles exist but are NOT stable — Alpha in OTel spec, Development tier in OTLP 1.11.0. Do NOT say "stable" in talk/blog. The profiler ships as an OTel Collector receiver (otelcol-ebpf-profiler), emitting natively via existing pipelines — no separate converter.

### [C-032] opentelemetry-ebpf-profiler supports whole-system continuous profiling without code changes

- **Verdict:** CONFIRMED (with deployment caveats)
- **Source(s):** S-P-05 (README + internals.md)
- **Notes:** "No code changes" holds for profiled apps (no recompilation, no agent injection, no restarts). BUT: requires the otelcol-ebpf-profiler DaemonSet deployed, root/CAP_BPF+CAP_PERFMON on the agent, minimum Linux kernel version (verify exact version). macOS dev requires Linux VM/container.

### [C-033] opentelemetry-ebpf-profiler unwinds Go stacks via .gopclntab (works on stripped production binaries)

- **Verdict:** CONFIRMED
- **Source(s):** S-P-05, S-P-06 (doc/gopclntab.md)
- **Notes:** Go executables lack .eh_frame (unless CGo), so profiler uses .gopclntab. Works on fully static stripped binaries. Covers OS-thread CPU stacks — goroutine-level profiling NOT confirmed.

### [C-034] CPU profiling is the confirmed profiling type; off-CPU is not yet in current release

- **Verdict:** CONFIRMED (CPU) / PLAUSIBLE NOT YET (off-CPU)
- **Source(s):** S-P-03 (profiles-alpha blog 2026)
- **Notes:** The data model supports off-CPU but the current profiler implementation has off-CPU as future work. Do not claim off-CPU support without verification against current CHANGELOG.

### [C-035] Current version is v0.0.202627 (calendar-week tags, no formal numbered releases)

- **Verdict:** CONFIRMED (at time of research, 2026-07-22)
- **Source(s):** S-P-05 (GitHub repo tags)
- **Notes:** Versioning scheme is ISO week-based (e.g. v0.0.202627 = week 27, 2026). Re-verify at time of publication.

### [C-036] Profiles are OpenTelemetry's fourth observability signal; the Profiles specification remains Alpha

- **Verdict:** CONFIRMED
- **Source(s):** S-P-07, S-P-08
- **Notes:** Logs, metrics, traces, and profiles are the four signal categories used in the talk. Do not call profiling the third signal or imply a stable profile data model.

### [C-037] OTEP 4947 proposes a Go-specific pprof-label path for sharing request context with profile readers

- **Verdict:** CONFIRMED as a proposal, not as shipped behavior
- **Source(s):** S-P-09
- **Notes:** Go is outside the primary TLSDESC mechanism for the foreseeable future because of goroutine concurrency and FFI cost. Go SDKs identify `go_pprof_labels_v1`; readers consume pprof labels directly. Compile-time instrumentation can arrange those labels, but the talk must describe this as proposed work.

---

## Go Runtime Futures

### [C-040] golang/go#63185 — flight recording — SHIPPED in Go 1.25

- **Verdict:** CONFIRMED (with API correction)
- **Source(s):** S-RF-01 (issue), S-RF-02 (go.dev blog Sept 2025), S-RF-03 (flightrecorder.go source)
- **Notes:** CLOSED Proposal-Accepted, milestoned Go 1.25. JFR-style circular-buffer tracer. API: `trace.NewFlightRecorder(cfg FlightRecorderConfig)` with struct fields `MinAge time.Duration, MaxBytes uint64`. Methods: `Start()`, `Stop()`, `Enabled()`, `WriteTo(w io.Writer)`. The prior lead mentioned `SetMinAge` as a method — WRONG. Shipped API uses struct fields. Do not use old API in talk or blog.

### [C-041] golang/go#69887 proposes -toolexec/compile-time instrumentation improvements (OPEN)

- **Verdict:** CONFIRMED (scope correction)
- **Source(s):** S-RF-04 (issue)
- **Notes:** Filed 2024-10-15 by Romain Marcadier (DataDog/Orchestrion). OPEN, Proposals Incoming. Does NOT propose general runtime hooks — proposes 5 improvements to -toolexec tooling: per-package build-ID influence, build-graph edges, go/ast API, source mapping, -p parallelism. Root gaps: cache invalidation, no visibility into build flags without crawling process tree.

### [C-042] golang/go#75654 and #38270 describe httptrace hook gaps (different classes)

- **Verdict:** CONFIRMED
- **Source(s):** S-RF-05 (#75654), S-RF-04 (research context for #38270)
- **Notes:** #75654 (2025-09-29, OPEN active): client-side — no hook for response body completion; OTel Go uses PutIdleConn as workaround which never fires for HTTP/2, causing client spans to never close (otel-contrib#4876, still open). #38270 (2020, OPEN Proposal-Hold Unplanned): server-side request lifecycle hooks. DIFFERENT classes — not the same problem. Prior lead described both as "hook-gap bugs" which is accurate but imprecise.

### [C-043] golang/go#67120 is the runtime/metrics recommended set proposal (OPEN, Incoming)

- **Verdict:** PLAUSIBLE — consistent with research but not directly confirmed against issue page
- **Source(s):** S-RF-04
- **Notes:** Research summary mentions it as "still incoming." Verify exact title and current status from <https://github.com/golang/go/issues/67120> before citing.

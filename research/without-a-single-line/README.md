# Research: Without a Single Line

> **Talk:** "How to Instrument Go Without Changing a Single Line of Code"
> GopherCon UK 2026 — 30-min keynote (advanced)

## Reading order

| Doc | Topic | Blog post |
|-----|-------|-----------|
| `01-thesis.md` | Why Go is hard to instrument; no runtime hook; the landscape | Post 1: "Why Go can't be monkey-patched" |
| `02-obi.md` | OBI (OpenTelemetry eBPF Instrumentation) — architecture, coverage, prod story | Post 2: "OBI: eBPF auto-instrumentation for Go in production" |
| `03-otelc.md` | otelc (OTel Go compile-time) — `-toolexec`, AST aspects, gls hack, local dev | Post 3: "otelc: zero-touch Go traces at compile time" |
| `04-ebpf-profiler.md` | opentelemetry-ebpf-profiler — profiling signal, OTLP profiles, unwinding | Post 4: "The third signal: continuous profiling without code changes" |
| `05-benchmarking.md` | Benchmark methodology — overhead/latency/alloc/CPU (numbers in demo/) | Post 5: "Benchmark: overhead of zero-touch instrumentation in Go" |
| `06-runtime-futures.md` | Go runtime futures — #63185, #69887, #67120, hook-gap bugs | Post 6: "Go runtime futures: flight recording, USDT, and hook proposals" |
| `07-usdt.md` | USDT for Go — Salp, libbpf/usdt, libstapsdt, Polar Signals | Post 6 (continued) |

## Key files

- `00-sources.md` — annotated bibliography; every URL + pinned version/SHA + access date
- `claims-ledger.md` — every load-bearing claim → source → verdict (CONFIRMED/PLAUSIBLE/REFUTED/PENDING)

## Research status

| Doc | Status |
|-----|--------|
| `01-thesis.md` | ✅ Complete — C-001 CONFIRMED, C-002 CONFIRMED (g-struct injection + gls.orchestrion.yml), C-003 corrected (51 not 54) |
| `02-obi.md` | ✅ Complete — C-010–C-014 CONFIRMED; open: full SUPPORT_MATRIX.md list, overhead benchmarks |
| `03-otelc.md` | ✅ Complete — C-020 CONFIRMED, C-021 REFUTED (no donation), C-022 corrected (v1.0.1 not v0.5.0) |
| `04-ebpf-profiler.md` | ✅ Complete — C-030–C-035 CONFIRMED; open: kernel min version, goroutine-level profiling |
| `05-benchmarking.md` | ⬜ Methodology outlined — numbers are TODO (experiments phase) |
| `06-runtime-futures.md` | ✅ Complete — C-040 CONFIRMED (SHIPPED Go 1.25!), C-041–C-042 CONFIRMED |
| `07-usdt.md` | ✅ Complete — USDT mechanics confirmed; speaker's poc_usdt fork identified |
| `claims-ledger.md` | ✅ All core claims resolved — no PENDING entries for primary content |

## Status: Research phase complete

All 7 primary research docs written. All load-bearing claims in `claims-ledger.md` are either
CONFIRMED (primary source), PLAUSIBLE (flagged), or REFUTED. No open PENDING items block the
talk or blog series.

**Remaining open items (low priority — verify before publishing):**
- `02-obi.md`: complete Go library list from SUPPORT_MATRIX.md; exact kernel min version
- `03-otelc.md`: otelc supported framework list; Go version minimum
- `04-ebpf-profiler.md`: goroutine-level profiling support; off-CPU status; kernel min version
- `06-runtime-futures.md`: confirm #67120 title/status from issue page
- `07-usdt.md`: Salp Go version compatibility; poc_usdt current build status

## Open questions (pre-research)

1. What is the official name after donation — "OBI" or something else? Is "otelc" the official name?
2. What is otelc's current release status — v0.5.0 was latest as of 2026-07-15; has v1 shipped?
3. Does opentelemetry-ebpf-profiler produce OTLP profiles natively, or does it need a converter?
4. Which Go versions does each approach support? Does otelc support Go 1.24+?
5. What is the current status of Go proposal #69887?

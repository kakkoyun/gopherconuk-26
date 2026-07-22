# Benchmarking Methodology

> **Status:** ⬜ Pending deep-research (methodology) + ⬜ Pending experiments (actual numbers)
> **Blog post:** "Benchmark: overhead of zero-touch instrumentation in Go"
> **Note:** This doc covers methodology only. Actual numbers come from `talks/without-a-single-line/demo/` experiments and are marked TODO until produced.

## What to measure

| Axis | Metric | Tool |
|------|--------|------|
| Latency | p50 / p95 / p99 request latency | hey / k6 / wrk2 |
| CPU overhead | % CPU increase vs uninstrumented baseline | top / perf stat / pprof |
| Memory / allocations | heap alloc rate, GC pressure | `go test -benchmem`, runtime/metrics |
| Binary size | binary size delta (compile-time approaches only) | `ls -lh`, `size` |
| Compile time | wall-clock build time delta (compile-time approaches only) | `time go build` |

## Fairness controls

- Same demo service, same workload, same hardware for all comparisons
- Warm-up period before measurement (avoid cold-start bias)
- Run at realistic concurrency (not single-threaded microbenchmarks)
- Measure at the service level, not just Go-level benchmarks (eBPF overhead is host-level)
- Report hardware: CPU model, core count, RAM, OS, kernel version
- Report software: Go version, OBI version, otelc version, profiler version
- Each run repeated N times; report median + stddev, not cherry-picked best

## Benchmark design questions (to research)

- [ ] What demo service is realistic? HTTP+DB (net/http + database/sql) covers the most common integrations
- [ ] What workload generator is appropriate for a 30-min talk demo? (k6 is scriptable, good for demos)
- [ ] Are there existing official benchmarks from OBI / otelc / profiler projects? (cite those too)
- [ ] How to isolate eBPF overhead from application overhead at the system level?
- [ ] What kernel tuning (if any) affects eBPF profiler overhead?

## Numbers (TODO — produced in experiments phase)

> Do not fabricate numbers. All figures here are placeholders until experiments are run.

| Approach | p99 latency vs baseline | CPU overhead | Notes |
|----------|------------------------|--------------|-------|
| Baseline | — | — | `talks/without-a-single-line/demo/baseline/` |
| OBI | TODO | TODO | kernel X.X, OBI vX.X.X |
| otelc | TODO | TODO | Go X.X, otelc vX.X.X |
| ebpf-profiler | TODO | TODO | kernel X.X, profiler vX.X.X |

## Content (to be filled by deep-research — methodology section)

<!-- Deep-research output goes here -->

# Benchmarking methodology

> **Status:** Deferred from the keynote
> **Decision:** The talk makes no comparative overhead claim for OBI, otelc, or the eBPF Profiler.

## Why the shootout was removed

A fair comparison needs a dedicated experiment rather than a placeholder table:

- otelc adds SDK work inside the application process.
- OBI and the eBPF Profiler consume host CPU outside the application.
- Kernel version, enabled probes, workload shape, concurrency, and signal configuration change the result.
- Application p99 latency alone cannot account for node-level observer cost.

The keynote has a 30-minute limit and prioritizes mechanism, constraints, and composition. Any future benchmark belongs in a separate experiment and must not be reconstructed from rehearsal notes.

## Requirements for a future experiment

| Axis | Metric | Example tool |
| --- | --- | --- |
| Request latency | p50, p95, p99 | k6 or wrk2 |
| Application CPU | CPU time per request | `perf stat`, runtime metrics |
| Observer CPU | OBI/profiler CPU by node | cgroup or container metrics |
| Memory | Heap and resident-set delta | runtime metrics, cgroup metrics |
| Binary size | Compile-time binary delta | `size`, file metadata |
| Build cost | Clean and cached build duration | hyperfine or controlled shell timing |

A valid run must pin hardware, kernel, Go version, tool versions, integration configuration, warm-up, workload, and sample count. It must report distributions rather than a single best result.

Until those measurements exist, the public deck should make no overhead ranking.

# opentelemetry-ebpf-profiler — The Profiling Signal

> **Status:** ✅ Research complete (deep-research workflow, 2026-07-22)
> **Blog post:** "The fourth signal: continuous profiling without code changes"
> **Key claims:** C-030 ✅ CONFIRMED · C-031 ✅ CONFIRMED · C-032 ✅ CONFIRMED

---

## 1. Provenance & Project Status

**opentelemetry-ebpf-profiler** originated as **Elastic Universal Profiling Agent**. Elastic pledged
to donate it to OpenTelemetry in **March 2024** (CNCF blog announcement) and completed the transfer
in **June 2024** via OTel community issue [#1918](https://github.com/open-telemetry/community/issues/1918).

- **GitHub repo:** `github.com/open-telemetry/opentelemetry-ebpf-profiler`
- **Current version:** `v0.0.202627` — the project uses **calendar-week git tags** (ISO week 27 of 2026);
  there are no formal numbered GitHub releases.
- **CNCF status:** Part of the OpenTelemetry project (CNCF Graduated).
- **Deployment today:** Ships as an **official OpenTelemetry Collector receiver** (`otelcol-ebpf-profiler`
  distribution), not a standalone binary. The eBPF agent works as a Collector receiver, leveraging
  existing OTel processing pipelines.

**Sources:** [S-P-01], [S-P-02], [S-P-03], [S-P-04]

---

## 2. How It Works: Architecture

The profiler is **100% non-intrusive** — the README states verbatim:

> "No need to load agents or libraries into the processes that are being profiled. No need for any
> reconfiguration, instrumentation or restarts of HLL interpreters and VMs."

**Mechanism:**

1. An eBPF program fires on CPU sample events at the **kernel level**.
2. It reads process-internal data structures *from outside* the target process (via BPF maps).
3. Stack frames are reconstructed using language-specific unwinding strategies (see below).
4. Symbolization maps raw addresses to function names/file/line.

The profiler agent itself requires **root/sudo** (or CAP_BPF + CAP_PERFMON). This is an agent-side
requirement only — the profiled applications need zero changes.

No LD_PRELOAD, no ptrace — purely kernel-level eBPF attachment. [S-P-05]

---

## 3. Stack Unwinding: Native Code

For C/C++ and system libraries, the profiler unwinds stacks using **.eh_frame data** (an exception
handling table present in most ELF binaries), protected by US patent US11604718B1.

- **No DWARF debug symbols required** on the host.
- **No frame pointers required.**
- Works on stripped production binaries.

[S-P-05]

---

## 4. Stack Unwinding: Go

For Go binaries, the profiler reads the **`.gopclntab` section** — Go's internal program counter to
line number table — rather than .eh_frame or DWARF.

Key point from `doc/gopclntab.md`:

> "The information remains present even for fully static and stripped executables and is thus very
> valuable because it allows us to symbolize and unwind production executables."

Go executables lack .eh_frame (unless built with CGo), so the profiler falls back to .gopclntab for
stack deltas. This means **stripped, statically-linked production Go binaries are profiled correctly
without any build-time changes.**

**Important caveat:** This covers OS-thread CPU stacks. Goroutine-level profiling (as distinct from
OS threads) and GC pressure profiling are **not confirmed** as current capabilities — they are not
addressed in the primary sources reviewed. Do not claim these without further verification.

[S-P-05], [S-P-06]

---

## 5. The OTel Profiling Signal: Status

The profiling signal occupies **two distinct stability axes**:

| Axis | Status | Source |
| ------ | -------- | -------- |
| OTel specification (`/docs/specs/otel/profiles/`) | **Alpha** | [S-P-07] |
| OTLP wire format (OTLP 1.11.0) | **Development** (lowest tier) | [S-P-08] |
| Traces / metrics / logs (for comparison) | Stable / Stable / Stable | [S-P-08] |

The profiler README self-describes: *"Implements the Alpha OTel Profiles signal."*

The signal entered **public Alpha** per the 2026 OTel blog post ("profiles-alpha"). [S-P-03]

**For the talk:** Do NOT say the profiling signal is "stable" or "production-ready" in the OTel
spec sense. It is Alpha/Development — evolving, not guaranteed backward-compatible.

---

## 6. Supported Profiling Types

**Confirmed in current release:**

- **CPU profiling** (on-CPU): core capability, fully supported.

**Status uncertain (PLAUSIBLE but not CONFIRMED):**

- **Off-CPU / blocking profiling:** The OTel profiles data format is *designed* to support off-CPU
  event capture (timestamped per-event data), but off-CPU collection by the current profiler
  implementation appears to be **future work**, not a current release capability. Do not assert this
  is supported without verification.
- **Memory / allocation profiling:** Not addressed in confirmed primary sources.

**Go-specific:** CPU profiling works at OS-thread level, correctly attributing Go function names
(via .gopclntab). Goroutine-level breakdown is not confirmed.

---

## 7. Language Support

The profiler supports multiple language runtimes via per-language unwinders. Confirmed support
includes:

- **Go** (.gopclntab-based)
- **C/C++** (.eh_frame-based)
- **System libraries** (stripped, no frame pointers)

Additional runtimes (JVM, Python, Ruby, Node.js, PHP, .NET) are mentioned in the README as
supported. **Verify the complete supported language list from the current repo README before
asserting in the talk.**

---

## 8. Kubernetes Deployment

The profiler deploys as a **DaemonSet** (one instance per node, profiles all workloads on the node).
This is the standard whole-system profiler model.

Exact Kubernetes YAML / Helm chart location: see the repo's `deploy/` directory or the Collector
distribution docs. **Verify the exact kernel minimum version from the current README before citing.**

---

## 9. "Zero Code Changes" — Caveats

The "zero code changes" claim holds for the **profiled applications**: no recompilation, no agent
injection, no restarts.

Deployment does require:

- The `otelcol-ebpf-profiler` Collector distribution deployed (as a DaemonSet or sidecar to the Collector).
- Root/CAP_BPF + CAP_PERFMON on the profiler agent node.
- A minimum Linux kernel version (verify exact version from repo).

For a macOS development machine, this means **a Linux VM or container runtime** — eBPF requires Linux.

---

## 10. Fit With the Talk

The profiler supplies the **fourth observability signal**: profiles, alongside logs, metrics, and traces.

The complementarity story:

- OBI: zero-touch **traces + metrics** in production.
- otelc: portable, production-capable **build-time instrumentation** with supported Go semantics.
- ebpf-profiler: zero-touch **continuous CPU profiles**, always-on, whole-system.

Together they cover complementary signals and layers without requiring application source edits.

---

## Open Questions (before talk/blog)

- [ ] Exact minimum Linux kernel version — check current README.
- [ ] Complete supported language list — check current README.
- [ ] Is off-CPU profiling in the current release? Check CHANGELOG.
- [ ] Does goroutine-level profiling (distinct from OS threads) work? Check docs.
- [ ] Exact Helm/DaemonSet deployment steps — check `deploy/` in repo.

---

## Sources Used

| Key | Description | URL |
| ----- | ------------- | ----- |
| S-P-01 | OTel blog: Elastic contributes continuous profiling agent (June 2024) | <https://opentelemetry.io/blog/2024/elastic-contributes-continuous-profiling-agent/> |
| S-P-02 | CNCF blog: OTel announces support for profiling (March 2024) | <https://www.cncf.io/blog/2024/03/19/opentelemetry-announces-support-for-profiling/> |
| S-P-03 | OTel blog: profiles-alpha (2026) | <https://opentelemetry.io/blog/2026/profiles-alpha/> |
| S-P-04 | OTel community issue #1918 (donation) | <https://github.com/open-telemetry/community/issues/1918> |
| S-P-05 | opentelemetry-ebpf-profiler README + internals.md | <https://github.com/open-telemetry/opentelemetry-ebpf-profiler> |
| S-P-06 | opentelemetry-ebpf-profiler doc/gopclntab.md | <https://github.com/open-telemetry/opentelemetry-ebpf-profiler/blob/main/doc/gopclntab.md> |
| S-P-07 | OTel specification: profiles signal | <https://opentelemetry.io/docs/specs/otel/profiles/> |
| S-P-08 | OTLP 1.11.0 specification | <https://opentelemetry.io/docs/specs/otlp/> |

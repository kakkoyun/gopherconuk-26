# OBI — OpenTelemetry eBPF Instrumentation

> **Status:** ✅ Research complete (deep-research workflow, 2026-07-22)
> **Blog post:** "OBI: eBPF auto-instrumentation for Go in production"
> **Key claims:** C-010 ✅ CONFIRMED · C-011 ✅ CONFIRMED · C-012 ✅ CONFIRMED

---

## 1. Provenance & Project Status

**OBI (OpenTelemetry eBPF Instrumentation)** is the direct successor to **Grafana Beyla**.
Grafana Labs donated Beyla to the CNCF OpenTelemetry project in **2025** (community issue
[#2406](https://github.com/open-telemetry/community/issues/2406)), renamed it OBI, and moved
all active development to the new repo.

From the Grafana Beyla README:
> "Beyla has been donated to the CNCF OpenTelemetry Project, under the project name OpenTelemetry
> eBPF Instrumentation or OBI for short... All Beyla current maintainers work full time on the
> upstream repository."

Grafana Beyla continues to exist as Grafana Labs' distribution of the upstream OBI project.

- **GitHub repo:** `github.com/open-telemetry/opentelemetry-ebpf-instrumentation`
- **Go module:** `go.opentelemetry.io/obi`
- **License:** Apache-2.0
- **Current release:** **v0.10.0 (2026-06-30)**
  - Still in **Development** status. From the README: *"OBI is currently in Development. Users
    should expect breaking changes between minor releases while the project remains in `v0`."*
- **Previous release:** v0.9.0 (2026-05-11). No v1.x release exists yet.
- **Official docs:** <https://opentelemetry.io/docs/zero-code/obi/>

**Sources:** [S-OBI-01], [S-OBI-02], [S-OBI-03]

---

## 2. How It Works: Architecture

OBI places **eBPF uprobes and kprobes** at the kernel/binary level. eBPF programs are
**JIT-compiled to the host native architecture** (x86-64, ARM64, etc.) by the Linux kernel.

For the standard case (HTTP/gRPC RED metrics — Rate, Errors, Duration):

- **Zero source code changes** required for the profiled applications.
- **Zero recompilation, zero restarts, zero agent injection** into process memory.

Important caveat from official docs:
> "Use language agents or manual instrumentation when you need custom spans, application-specific
> attributes, business events, or other in-process telemetry."

**The "zero code changes" claim holds for standard RED metrics and library-level spans.
It does NOT hold for custom spans, business-logic events, or SQL query details.** Those still
require manual instrumentation or compile-time tools like otelc.

**Sources:** [S-OBI-04], [S-OBI-05]

---

## 3. Linux Kernel Requirements

| Requirement | Detail |
| ------------- | -------- |
| **Kernel version** | Linux **5.8+** |
| **RHEL exception** | Linux **4.18+** for RHEL 8, CentOS 8, Rocky Linux 8, AlmaLinux 8, and compatible derivatives with required eBPF backports |
| **BTF required** | Yes — BPF Type Format must be enabled. BTF became default on most Linux distros with kernel **5.14+** |

From the OBI docs (last modified 2026-07-20, compatibility table):
> "Linux kernel: 5.8+, or RHEL-family Linux 4.18+ with the required eBPF backports"

**For macOS development:** eBPF requires Linux. Demos need a Linux VM or Docker on Linux.

**Sources:** [S-OBI-04], [S-OBI-06]

---

## 4. Linux Capabilities Vary by Feature

OBI does not have one universal unprivileged capability set. The required access grows with the enabled mode:

| Mode | Required access |
| ------------ | --------- |
| Network flow capture | `CAP_BPF`, `CAP_NET_RAW` |
| Application observability | Process and ELF access, `CAP_PERFMON`, uprobes |
| Context propagation | `CAP_NET_ADMIN` |
| Go library propagation | May require `CAP_SYS_ADMIN` for `bpf_probe_write_user` |

`CAP_SYS_PTRACE` allows access to process information under `/proc`; OBI does not use it to call `PTRACE_ATTACH`. Restrictive `kernel.perf_event_paranoid` settings may force `CAP_SYS_ADMIN` instead of `CAP_PERFMON`. Secure Boot and kernel lockdown can disable context-propagation helpers.

Alternatively, `privileged: true` can be used in Kubernetes, but it grants much broader access.

**Sources:** [S-OBI-05], [S-OBI-07]

---

## 5. Go Library-Level Support

OBI provides Go-specific **library-level uprobe instrumentation** (distinct from generic network-level interception) for **13 documented baselines**:

| Library | Baseline |
| --------- | --------------- |
| `net/http` | Go 1.17+ |
| `golang.org/x/net/http2` | >= 0.12.0 |
| `github.com/gorilla/mux` | >= v1.5.0 |
| `github.com/gin-gonic/gin` | >= v1.6.0, except v1.7.5 |
| `google.golang.org/grpc` | >= 1.40 |
| `net/rpc/jsonrpc` | Go 1.17+ |
| `database/sql` | Go 1.17+ |
| `github.com/go-sql-driver/mysql` | >= v1.5.0 |
| `github.com/lib/pq` | All versions |
| `github.com/redis/go-redis/v9` | >= v9.0.0 |
| `github.com/segmentio/kafka-go` | >= v0.4.11 |
| `github.com/IBM/sarama` | >= 1.37 |
| `go.mongodb.org/mongo-driver` | v1 >= v1.10.1; v2 >= v2.0.1 |

The support matrix is the source of truth. Internal layouts can change even when an import path remains compatible.

**Sources:** [S-OBI-06]

---

## 6. Language-Agnostic Support

At the **network-protocol level**, OBI is language-agnostic. It supports:

- Go (1.17+)
- Java (JDK 8+, with an embedded Java agent extracted at runtime)
- .NET
- Node.js (8.0+, async-hooks context propagation)
- Python, Ruby, C, C++, Rust
- GenAI provider SDKs

This means OBI can instrument a polyglot microservices environment with a single DaemonSet.

**Sources:** [S-OBI-04], [S-OBI-06]

---

## 7. Kubernetes Deployment

Two models:

### DaemonSet (recommended for production)

- One OBI pod per node; instruments all workloads on the node.
- Requires `hostPID: true` (to access all processes on the node).
- Standard Linux capabilities or `privileged: true`.
- No changes to application pods.

### Sidecar

- One OBI container per application pod.
- Requires `shareProcessNamespace: true` and `privileged: true` on the sidecar.
- More granular control; less efficient at scale.

**Sources:** [S-OBI-05]

---

## 8. "Zero Code Changes" — Precise Scope

| Scenario | Zero code changes? |
| ---------- | -------------------- |
| HTTP/gRPC RED metrics (rate, errors, duration) | ✅ Yes |
| Library-level spans (13 supported Go libs) | ✅ Yes |
| Custom spans / business-logic events | ❌ No — requires manual instrumentation |
| SQL query details / parameters | ❌ No — requires manual instrumentation |
| Trace context propagation across services | ✅ Yes (for supported protocols) |

The talk's claim "without a single line of code" holds for the standard observability case.
Acknowledge the boundary: for custom in-process telemetry, code changes are still needed.

---

## 9. Fit With the Talk

**Production story**: OBI is the "attach from outside, zero rebuild" approach. Deploy the DaemonSet
once; all services on the node emit RED metrics and library-level spans. No CI change, no
deployment coordination with app teams.

**Complementarity with otelc:**

- OBI: breadth (13 Go libs + any language on the network level), zero rebuild, production-safe.
- otelc: depth (granular custom spans, stdlib instrumentation, business-logic events), requires rebuild.
- ebpf-profiler: profiles (CPU, system-wide, always-on), separate signal.

**Decision rule** (for the agent skill):

- `"production"` / `"k8s"` / `"runtime"` / `"no rebuild"` → **OBI**
- Build ownership / non-Linux target / richer supported Go semantics / no runtime privileges → **otelc**

---

## Open Questions (before talk/blog)

- [ ] Read `SUPPORT_MATRIX.md` at v0.10.0 tag — get complete Go library list with exact version constraints.
- [ ] Verify exact Helm chart / operator installation command from v0.10.0 docs.
- [ ] Confirm: is Grafana Beyla still in active development, or are all changes going to OBI?
- [ ] Check if OBI has official performance/overhead benchmarks in the repo or docs.
- [ ] Verify: does Java support require OBI to inject a Java agent at runtime — if so, is that "zero code changes"?

---

## Sources Used

| Key | Description | URL |
| ----- | ------------- | ----- |
| S-OBI-01 | Grafana Beyla OSS page (donation announcement) | <https://grafana.com/oss/beyla-ebpf/> |
| S-OBI-02 | Grafana Beyla README (donation reference + #2406) | <https://github.com/grafana/beyla> |
| S-OBI-03 | OBI pkg.go.dev (v0.10.0 release date, module path) | <https://pkg.go.dev/go.opentelemetry.io/obi> |
| S-OBI-04 | OBI official docs (zero-code, language support, kernel) | <https://opentelemetry.io/docs/zero-code/obi/> |
| S-OBI-05 | OBI Kubernetes setup docs | <https://opentelemetry.io/docs/zero-code/obi/setup/kubernetes/> |
| S-OBI-06 | OBI SUPPORT_MATRIX.md (v0.10.0) | <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/main/SUPPORT_MATRIX.md> |

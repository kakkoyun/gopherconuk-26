---
name: collect-go-telemetry
description: >
  Collect telemetry from a Go service without changing its source. Routes to
  OBI for rebuild-free Linux observation, otelc for portable build-time
  instrumentation, or a combined deployment based on platform, privileges,
  signal depth, and build ownership.
use_when:
  - "collect go telemetry"
  - "instrument go service without code changes"
  - "troubleshoot go service production"
  - "attach obi to service"
  - "build with otelc"
  - "zero touch go observability"
disable-model-invocation: false
---

# Collect Go telemetry

Collect telemetry from a Go service without editing application source.

## Step 1: identify the hard constraints

Ask or infer:

1. Is the service running on Linux?
2. Can the team rebuild the binary or change the build pipeline?
3. Can a privileged observer run beside the workload?
4. Does the user need boundary coverage or Go-specific semantic detail?
5. Which signals are required: traces, metrics, logs, profiles?

### Choose OBI when

- The service is already deployed and a rebuild is unavailable.
- The workload runs on supported Linux kernels.
- A privileged DaemonSet, sidecar, or host process is acceptable.
- The fleet uses several languages.
- HTTP, gRPC, database, messaging, RED metrics, or trace-log correlation is sufficient.

OBI is a v0 project. Emitted telemetry and configuration may change between minor releases.

### Choose otelc when

- The team controls the Go build command or CI pipeline.
- The target includes Linux, macOS, or Windows.
- Runtime privileges or eBPF access are unavailable.
- Supported Go libraries need in-process semantic instrumentation.
- The instrumentation bundle should emit traces, metrics, or supported log records.

otelc is a stable production build-time tool. Local development is one use case, not its deployment boundary.

### Combine them when

- A Go service needs rich in-process semantics and also runs in a mixed-language Linux fleet.
- Build-time instrumentation supplies request context that an out-of-process profiler can correlate.
- OBI should cover service boundaries while otelc or Orchestrion covers supported Go internals.

If context is ambiguous, ask: "Can you rebuild the service, what operating system does it run on, and can the runtime host a privileged eBPF observer?"

## Backend A: OBI

### What OBI provides

- OTLP traces and metrics from network protocols and supported library operations.
- Thirteen documented Go library baselines, including `net/http`, gin, gRPC, `database/sql`, go-redis, Kafka, and MongoDB.
- Optional JSON trace-log correlation. OBI enriches logs in place; the existing logging pipeline still exports them.
- No application rebuild or restart for node-level attachment.

### Platform and privilege requirements

- Linux `amd64` or `arm64`.
- Kernel 5.8+ with BTF, or the documented RHEL-family 4.18+ backports.
- Capabilities depend on the enabled mode:
  - Network flow capture uses `CAP_BPF` and `CAP_NET_RAW`.
  - Application observability adds executable/process access and `CAP_PERFMON`.
  - Context propagation adds `CAP_NET_ADMIN`.
  - Go library propagation may require `CAP_SYS_ADMIN`.
- `kernel.perf_event_paranoid`, Secure Boot, and kernel lockdown can restrict features.

### Pull one integration row

```bash
./tools/cli/go-instr-pull/obi-integration.sh net/http
./tools/cli/go-instr-pull/obi-integration.sh gin
./tools/cli/go-instr-pull/obi-integration.sh grpc
```

The script fetches only the matching row from OBI's support matrix.

### Deploy on Kubernetes

Follow the current upstream setup documentation:

<https://opentelemetry.io/docs/zero-code/obi/setup/kubernetes/>

The repository's `tools/cli/kubectl-obi` directory is a prototype. Its command parsing exists, but cluster operations are incomplete. Do not present it as a production installer.

### Verify OBI

```bash
uname -r
ls /sys/kernel/btf/vmlinux
kubectl get pods -n obi-system
kubectl logs -n obi-system -l app.kubernetes.io/instance=obi --tail=50
```

Check that the service uses a supported library and that the OTLP endpoint is reachable.

### OBI boundary

OBI does not infer arbitrary business events, domain attributes, unsupported internal functions, or application-specific sampling decisions. Add manual instrumentation or a compile-time rule when those semantics matter.

## Backend B: otelc

### What otelc provides

- Build-time instrumentation for supported code, dependencies, and standard-library paths.
- Traces, HTTP and gRPC metrics, Go runtime metrics, and supported `slog` or Logrus records.
- No runtime agent, root access, eBPF capability, or Linux-kernel dependency.
- Build and test coverage across Linux, macOS, and Windows.

Current otelc requires Go 1.25 or newer. Coverage is limited to the integrations and versions declared by the project.

Orchestrion with dd-trace-go uses the same build-time family and can also enable Datadog runtime metrics, log correlation, and continuous profiling.

### Pull integration details

```bash
./tools/cli/go-instr-pull/otelc-aspect.sh net/http
./tools/cli/go-instr-pull/otelc-aspect.sh github.com/gin-gonic/gin
./tools/cli/go-instr-pull/otelc-aspect.sh google.golang.org/grpc
```

### Install and build

```bash
go install go.opentelemetry.io/otelc/tool/cmd/otelc@latest

otelc go build -o ./myapp ./...

OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
OTEL_SERVICE_NAME=my-go-service \
./myapp
```

The same wrapper can run tests and other Go commands where supported:

```bash
otelc go test ./...
otelc go run ./...
```

### Verify otelc

```bash
otelc version
otelc go build -v ./... 2>&1 | grep -i instrument
```

Also verify that an OTLP receiver is reachable and that the service uses a supported integration.

## Add profiling

The OpenTelemetry eBPF Profiler provides whole-node CPU profiles without changing the application. It requires Linux and profiler-agent privileges. Use it beside either instrumentation path when CPU evidence matters.

For request-correlated profiles, ensure the in-process instrumentation publishes suitable pprof labels. OTEP 4947 describes the proposed Go-specific context-sharing path:

<https://github.com/open-telemetry/opentelemetry-specification/blob/main/oteps/profiles/4947-thread-ctx.md#alternative-for-go-support>

Treat this OTEP path as proposed work unless the selected SDK documents a shipped implementation.

## Decision summary

| Constraint | First choice |
| --- | --- |
| No rebuild window on supported Linux | OBI |
| Non-Linux target | otelc / Orchestrion |
| Mixed-language boundary coverage | OBI |
| Rich supported Go semantics | otelc / Orchestrion |
| No privileged runtime agent | otelc / Orchestrion |
| Whole-node CPU profiles | OpenTelemetry eBPF Profiler |
| Request-correlated profiles | Build-time context plus profiler |

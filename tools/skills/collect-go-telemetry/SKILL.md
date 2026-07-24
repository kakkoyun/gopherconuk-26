---
name: collect-go-telemetry
description: >
  Collect telemetry from a Go service to troubleshoot or optimize it — without
  changing source code. Routes to OBI (eBPF, production/runtime, zero rebuild)
  or otelc (compile-time, local dev, granular spans) based on context, then
  pulls only the specific integration docs needed.
use_when:
  - "collect go telemetry"
  - "instrument go service without code changes"
  - "troubleshoot go service production"
  - "attach obi to service"
  - "otelc local dev traces"
  - "zero touch go observability"
disable-model-invocation: false
---

# collect-go-telemetry

Collect telemetry from a Go service for troubleshooting or optimization — no source changes required.

## Step 1 — Determine backend

Ask the user (or infer from context) which scenario applies:

**→ OBI** when any of these are true:
- Service is already deployed (production, staging, k8s, docker-compose)
- Rebuild/redeploy is not acceptable
- Service is not Go-only (polyglot, multiple languages)
- Goal is HTTP/gRPC RED metrics (rate, errors, duration) or library-level spans
- User says: "production", "k8s", "pod", "cluster", "running", "no rebuild", "runtime"

**→ otelc** when any of these are true:
- Working locally or in a dev environment
- Need granular spans (custom functions, specific code paths, business logic)
- Willing to rebuild with `otelc go build`
- Debugging a specific slow path with exact trace data
- User says: "local", "dev", "debug", "granular", "custom span", "trace specific function"

If context is ambiguous, ask: *"Is this for a running production service (OBI) or local development where you can rebuild (otelc)?"*

---

## Backend A: OBI — eBPF, production, zero rebuild

### What OBI gives you
- HTTP/gRPC RED metrics and library-level spans: 13 Go libs including net/http, gin, gRPC, gorilla/mux, go-redis, Kafka, database/sql
- No rebuild, no restart, no code changes — attaches from outside the process
- Kernel 5.8+ with BTF required; 6 Linux capabilities (CAP_BPF, CAP_SYS_PTRACE, CAP_NET_RAW, CAP_CHECKPOINT_RESTORE, CAP_DAC_READ_SEARCH, CAP_PERFMON)
- **Limitation:** custom spans, business-logic events, SQL query details still need code changes

### Fetch integration details (token-thrift)

Before attaching, pull docs for only the libraries the service uses:

```bash
# Usage: tools/cli/go-instr-pull/obi-integration.sh <library-name>
# Examples:
./tools/cli/go-instr-pull/obi-integration.sh net/http
./tools/cli/go-instr-pull/obi-integration.sh gin
./tools/cli/go-instr-pull/obi-integration.sh grpc
```

This fetches only the relevant section from OBI's SUPPORT_MATRIX.md — not the full catalog.

### Attach: Kubernetes (DaemonSet)

```bash
# 1. Add OBI DaemonSet to the cluster (one-time, instruments all pods on the node)
kubectl apply -f https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/releases/download/v0.10.0/obi-daemonset.yaml

# 2. Verify OBI is running
kubectl get pods -n obi-system

# 3. Check that spans are flowing (assumes OTel Collector configured)
kubectl logs -n obi-system -l app=obi --tail=50
```

### Attach: Docker Compose

```yaml
# Add to docker-compose.yml alongside your service:
obi:
  image: ghcr.io/open-telemetry/opentelemetry-ebpf-instrumentation:v0.10.0
  pid: host          # required: access host process namespace
  privileged: true   # or use specific capabilities
  environment:
    OTEL_EXPORTER_OTLP_ENDPOINT: http://otel-collector:4317
  volumes:
    - /sys/fs/bpf:/sys/fs/bpf
    - /sys/kernel/debug:/sys/kernel/debug
```

```bash
docker compose up -d obi
```

### Verify output

```bash
# If using Jaeger locally:
open http://localhost:16686

# If tailing OTLP stdout exporter:
docker compose logs obi -f
```

### OBI troubleshooting checklist
- [ ] Kernel ≥ 5.8 with BTF: `uname -r && ls /sys/kernel/btf/vmlinux`
- [ ] Capabilities present: `cat /proc/$(pgrep obi)/status | grep Cap`
- [ ] Service uses a supported library (check SUPPORT_MATRIX.md via script above)
- [ ] OTEL_EXPORTER_OTLP_ENDPOINT points to a live collector

---

## Backend B: otelc — compile-time, local dev, granular spans

### What otelc gives you
- Granular OTel spans for all supported Go frameworks, injected at compile time
- Instruments your code, its dependencies, and parts of the standard library
- No runtime overhead beyond the OTel SDK
- **Requires:** rebuild with `otelc go build`; Linux/macOS dev machine; Go version ≥ (check go.mod)

### Fetch integration details (token-thrift)

Pull docs for only the packages the service uses:

```bash
# Usage: tools/cli/go-instr-pull/otelc-aspect.sh <import-path>
# Examples:
./tools/cli/go-instr-pull/otelc-aspect.sh net/http
./tools/cli/go-instr-pull/otelc-aspect.sh github.com/gin-gonic/gin
./tools/cli/go-instr-pull/otelc-aspect.sh google.golang.org/grpc
```

### Install otelc

```bash
go install go.opentelemetry.io/otelc/tool/cmd/otelc@latest
```

### Build and run with instrumentation

```bash
# Replace your normal go build with:
otelc go build -o ./myapp ./...

# Or for go run:
otelc go run ./...

# With OTel SDK environment:
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
OTEL_SERVICE_NAME=my-go-service \
./myapp
```

### Configure OTel SDK (if not already present)

```bash
# Minimal: stdout exporter for local dev
OTEL_EXPORTER_OTLP_PROTOCOL=grpc \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
./myapp
```

Or use Jaeger all-in-one for local trace UI:

```bash
docker run -d --name jaeger \
  -p 16686:16686 -p 4317:4317 \
  jaegertracing/all-in-one:latest
open http://localhost:16686
```

### otelc troubleshooting checklist
- [ ] `otelc version` — confirm binary is on PATH
- [ ] `go.mod` uses a Go version otelc supports
- [ ] OTel Collector or Jaeger running at `OTEL_EXPORTER_OTLP_ENDPOINT`
- [ ] Run `otelc go build -v` to see which packages are being instrumented
- [ ] If zero spans: check `OTEL_SDK_DISABLED` is not set

---

## Decision summary

| Need | Use |
|------|-----|
| Production service, no rebuild | **OBI** |
| Any language in the fleet | **OBI** |
| HTTP/gRPC RED metrics only | **OBI** |
| Local dev, granular spans | **otelc** |
| Custom business-logic traces | **otelc** |
| Specific slow function traced exactly | **otelc** |
| Always-on CPU profiling (no code changes) | `opentelemetry-ebpf-profiler` (separate skill) |

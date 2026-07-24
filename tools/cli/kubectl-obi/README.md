# kubectl-obi

kubectl plugin for [OBI (OpenTelemetry eBPF Instrumentation)](https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation) — zero-touch distributed tracing for Kubernetes workloads.

OBI uses eBPF to instrument Go, Java, .NET, Python, Node.js, and more without code changes or restarts.

## Requirements

- Linux kernel 5.8+ with BTF (or RHEL 4.18+ with eBPF backports)
- Capabilities: `CAP_BPF`, `CAP_SYS_PTRACE`, `CAP_NET_RAW`, `CAP_CHECKPOINT_RESTORE`, `CAP_DAC_READ_SEARCH`, `CAP_PERFMON`
- An OTel Collector (or Jaeger) reachable from the cluster

> **macOS:** eBPF requires Linux. Use a Linux VM, kind/k3d with Docker on Linux, or a remote cluster.

## Install

```bash
# Via krew (once published to krew index):
kubectl krew install obi

# Or build from source:
cd tools/cli/kubectl-obi
go build -o kubectl-obi .
mv kubectl-obi $(go env GOPATH)/bin/
```

## Commands

### `kubectl obi attach`

Attach OBI to your cluster. DaemonSet mode (default) instruments every pod on every node — no per-service configuration.

```bash
# DaemonSet (recommended — instruments everything, no pod changes)
kubectl obi attach

# Sidecar (one specific deployment, requires rollout restart)
kubectl obi attach my-service --mode=sidecar --namespace=production
```

### `kubectl obi status`

Show which workloads OBI is currently instrumenting.

```bash
kubectl obi status
kubectl obi status --all-namespaces
```

### `kubectl obi traces`

Pull recent spans for a deployment (requires OTel Collector or Jaeger).

```bash
kubectl obi traces my-service --tail=50
kubectl obi traces my-service --follow
```

### `kubectl obi detach`

Remove OBI instrumentation.

```bash
# Remove DaemonSet (stops all instrumentation)
kubectl obi detach

# Remove sidecar from one deployment
kubectl obi detach my-service --mode=sidecar
```

## Status

**Implementation status: skeleton.** Commands are defined and flag parsing is complete. The k8s API calls (DaemonSet apply/delete, deployment patching, trace querying) are stubbed with TODO markers — they require a cluster to test and implement. See `main.go` for the TODOs.

**To complete the implementation:**
1. Wire up `client-go` for the cluster API calls (or shell out to `kubectl apply`)
2. Add a trace query backend (Jaeger HTTP API or OTLP query protocol)
3. Build release binaries and populate `kubectl-obi.yaml` with real SHAs
4. Submit to the krew index: https://krew.sigs.k8s.io/docs/developer-guide/release/

## Reference

- [OBI repo](https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation)
- [OBI Kubernetes setup](https://opentelemetry.io/docs/zero-code/obi/setup/kubernetes/)
- [krew plugin guide](https://krew.sigs.k8s.io/docs/developer-guide/)

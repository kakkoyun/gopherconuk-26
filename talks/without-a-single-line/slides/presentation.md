---
marp: true
theme: otel
math: mathjax
html: true
paginate: true
style: |
  .columns { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1.5rem; }
  .columns3 { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 1rem; }
  .big    { font-size: 1.4em; }
  .medium { font-size: 1.0em; }
  .small  { font-size: 0.8em; }
  .tiny   { font-size: 0.6em; }
  .center { text-align: center; }
  .tag    { display: inline-block; padding: 0.15em 0.6em; border-radius: 4px; font-size: 0.75em; font-weight: bold; }
  .prod   { background: #1e6845; color: #c3e8d4; }
  .dev    { background: #1a3a6e; color: #c3d4f0; }
  .always { background: #5a1e6e; color: #e4c3f0; }
  table, td, tr, th { background-color: transparent !important; }
---

<!-- _class: lead -->
<!-- _paginate: false -->

# How to Instrument Go Without Changing a Single Line of Code

**Kemal Akkoyun** · Datadog
GopherCon UK 2026 · Keynote

---

<!-- _paginate: false -->
<!-- _class: vcenter -->

## Quick question

<br>

### Who here has added instrumentation by hand — wrapping every HTTP client, every DB call, every function?

<br>

### 🙋

---

<!-- _paginate: false -->
<!-- _class: vcenter -->

## What if you didn't have to?

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 1
## Why Go Resists Instrumentation

---

## The easy routes — closed for Go

| Language | Zero-touch mechanism |
|----------|---------------------|
| **JVM** | `-javaagent` JVMTI hook, bytecode rewriting |
| **Python** | `importlib` hooks, `sys.settrace` |
| **Ruby** | `TracePoint`, dynamic method wrapping |
| **Go** | ❌ None of the above |

Go compiles to **native machine code**. No bytecode. No classloader. No dynamic dispatch.

---

## What Go actually has

```go
// runtime/proc.go — the Go team's comment on go:linkname abuse:

// gopark should be an internal detail,
// but widely used packages access it using linkname.
// Notable members of the hall of shame include:
// ...
```

The only "hook" is reluctant accommodation of abuse.

---

## LD_PRELOAD? Blocked.

```bash
# Works on dynamic binaries:
LD_PRELOAD=./tracer.so ./python-app    ✅

# Go uses its internal linker by default:
LD_PRELOAD=./tracer.so ./go-app        ❌ (silently ignored)

# Requires forcing external linking:
go build -ldflags="-linkmode=external" ⚠️  (non-default, not always possible)
```

Static cgo binaries: **rejected by injection vendors outright.**
`CGO_ENABLED=0 + -buildmode=pie`: no libc dependency → no injection.

---

## Three workarounds (today)

<div class="columns3">

<div>

### eBPF
Probe from the **kernel**

No rebuild, no restart
Language-agnostic at network level
Requires Linux 5.8+ + BTF

</div>

<div>

### Compile-time
Rewrite **before** the compiler

Rebuild required
Granular, in-process spans
Works on any OS

</div>

<div>

### Runtime injection
Patch **running** binary

Needs external linker
Limited support
Dynatrace OneAgent

</div>

</div>

<br>

Today: three open-source projects, one for each signal.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 2
## OBI — eBPF from the Outside
<br>
<span class="tag prod">production · zero rebuild</span>

---

## What is OBI?

**OpenTelemetry eBPF Instrumentation** — donated to OTel by Grafana Labs in 2025.
Previously known as Grafana Beyla.

- Repo: `github.com/open-telemetry/opentelemetry-ebpf-instrumentation`
- Current: **v0.10.0** (Development — breaking changes in v0.x)
- Beyla continues as Grafana's distribution of upstream OBI

---

## How it works

```
Your Go service (running)
        │
        │  eBPF uprobes on function entry/exit
        ▼
  Linux Kernel (5.8+ with BTF)
        │
        │  JIT-compiled eBPF programs
        ▼
  OBI DaemonSet pod
        │
        │  OTLP spans + metrics
        ▼
  OpenTelemetry Collector → Jaeger / Prometheus / ...
```

**Zero source changes. Zero restarts. Zero recompilation.**

---

## What OBI instruments (Go)

13 Go libraries via dedicated uprobes:

| Library | Min version |
|---------|------------|
| `net/http` | Go 1.17+ |
| `google.golang.org/grpc` | ≥ 1.40 |
| `github.com/gin-gonic/gin` | ≥ v1.6.0 |
| `gorilla/mux` | ≥ v1.5.0 |
| `go-redis/redis` v8/v9 | (added v0.7.1) |
| `database/sql` | (fixed v0.7.0) |
| + 7 more | see SUPPORT_MATRIX.md |

Plus: Java, .NET, Node.js, Python, Ruby, C, C++, Rust at network level.

---

## Deploy: DaemonSet (one command)

```bash
kubectl apply -f \
  https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/\
releases/download/v0.10.0/obi-daemonset.yaml
```

One pod per node. Instruments **every workload** on the node automatically.
No changes to application deployments.

```bash
# Or with the kubectl-obi plugin:
kubectl obi attach
kubectl obi status
```

---

## Requirements

| Requirement | Detail |
|-------------|--------|
| **Kernel** | Linux 5.8+ with BTF (BTF default since 5.14+) |
| **RHEL exception** | 4.18+ with eBPF backports |
| **Capabilities** | `CAP_BPF`, `CAP_SYS_PTRACE`, `CAP_NET_RAW`, `CAP_CHECKPOINT_RESTORE`, `CAP_DAC_READ_SEARCH`, `CAP_PERFMON` |
| **macOS dev** | Linux VM or remote cluster required |

---

## The honest scope

✅ **Zero code changes** for:
- HTTP/gRPC RED metrics (rate, errors, duration)
- Library-level spans (13 Go libs)

❌ **Still needs code changes** for:
- Custom spans with business-logic attributes
- SQL query details / parameters
- In-process context propagation beyond library boundaries

---

<!-- _class: vcenter center -->

## 🔴 Live Demo
### Attach OBI to a running HTTP service

```bash
kubectl obi attach --mode=daemonset
# → traces appear in Jaeger within seconds
```

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 3
## otelc — Compile-Time, From the Inside
<br>
<span class="tag dev">local dev · granular spans</span>

---

## Two tools, one idea

<div class="columns">

<div>

### `orchestrion` (Datadog)
```bash
orchestrion go build .
```
GA v1.11.0 · defaults to dd-trace-go/v2
Vendor-agnostic · battle-tested

</div>

<div>

### `otelc` (OTel SIG)
```bash
otelc go build ./...
```
Stable v1.0.1 (2026-07-14) · OTel SDK
Datadog + Alibaba co-founded SIG

</div>

</div>

Same core mechanism. Both use Go's `-toolexec` flag.

---

## How `-toolexec` works

```
go build ./...
    │
    │  -toolexec otelc
    ▼
otelc intercepts EVERY call to go tool compile
    │
    │  Parse .go files → Decorated Syntax Tree (dst)
    │  Apply instrumentation aspects
    │  Write rewritten files
    ▼
go tool compile (sees modified source)
    │
    ▼
Binary with spans baked in
```

Orchestrion calls this "compile-time-woven Aspect-Oriented Programming."

---

## The goroutine-local storage hack

Go has no goroutine-local storage. The workaround:

**Step 1** — Orchestrion's aspect patches `runtime.g`:

```yaml
# internal/orchestrion/gls.orchestrion.yml (in dd-trace-go, commit b97e7cbb4)
join-point:
  struct-definition: runtime.g
advice:
  - add-struct-field:
      name: __dd_gls_v2
      type: any
```

---

## The goroutine-local storage hack (cont.)

**Step 2** — Inject typed accessors via `go:linkname`:

```go
// Injected into the runtime package:
//go:linkname __dd_orchestrion_gls_get __dd_orchestrion_gls_get.V2
var __dd_orchestrion_gls_get = func() any {
    return getg().m.curg.__dd_gls_v2  // THIS goroutine's field
}
```

**Step 3** — Clean up on goroutine exit:

```go
// Injected into runtime.goexit1:
getg().__dd_gls_v2 = nil   // prevent memory leaks
```

The Go team's own comment calls `go:linkname` abusers a **"hall of shame."**
This is how deep you have to go.

---

## go:linkname for variables: fragile

```go
// dd-trace-go/internal/orchestrion/gls.go
//go:linkname __dd_orchestrion_gls_get __dd_orchestrion_gls_get.V2
var __dd_orchestrion_gls_get func() any   // nil when Orchestrion absent
```

Unlike function `go:linkname`, variable linkname has **no definition/reference separation**.
When both sides are BSS symbols — the linker picks arbitrarily.

> *"the choice is arbitrary. In the implementation it depends on the symbol loading order"*
> — Ian Lance Taylor, golang/go#72032

This **broke between Go 1.22 and 1.23.**

---

## Requirements & limits

| | |
|--|--|
| **Go version** | **1.25+** (otelc); check go.mod for orchestrion |
| **Platform** | Any OS (no eBPF required) |
| **What it gives you** | Granular spans, stdlib instrumentation, custom span injection |
| **Requires** | `otelc go build` instead of `go build` (CI/CD change) |
| **Runtime overhead** | OTel SDK cost only — no eBPF, no kernel context switches |

---

<!-- _class: vcenter center -->

## 🔴 Live Demo
### Build with otelc, see granular spans

```bash
otelc go build -o ./myapp ./...
OTEL_SERVICE_NAME=demo ./myapp
# → per-function spans in Jaeger
```

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 4
## opentelemetry-ebpf-profiler — The Third Signal
<br>
<span class="tag always">always-on · whole-system profiles</span>

---

## The missing signal

OBI gives you **traces + metrics**.
otelc gives you **granular spans**.
But what about **profiles**?

> Where is the CPU time actually going?
> Which functions allocate the most?

---

## opentelemetry-ebpf-profiler

Donated by **Elastic** (formerly Elastic Universal Profiling) to OTel in June 2024.

- Repo: `github.com/open-telemetry/opentelemetry-ebpf-profiler`
- Version: `v0.0.202627` (calendar-week tags — no formal releases yet)
- Ships as the **`otelcol-ebpf-profiler`** OTel Collector distribution
- OTLP profiles signal: **Alpha** (OTel spec) / **Development** (OTLP 1.11.0)

---

## How it unwinds Go stacks

The profiler reads **`.gopclntab`** — Go's internal PC-to-line table.

```
Go binary (stripped, static, production)
    │
    │  .gopclntab survives stripping ✅
    ▼
opentelemetry-ebpf-profiler
    │  reads process memory via eBPF
    │  symbolizes using .gopclntab
    ▼
Named function stacks → OTLP profiles
```

From `doc/gopclntab.md`:
> *"The information remains present even for fully static and stripped executables."*

---

## What zero-touch means here

✅ No code changes to profiled apps
✅ No recompilation — not even for stripped production binaries
✅ No agent injection into process memory

Deploy one DaemonSet. Profile **every** service on every node simultaneously.

Profiler agent itself needs `root` / `CAP_BPF + CAP_PERFMON`.
That's the agent's requirement, not each application's.

---

## The three signals together

<div class="columns3">

<div class="center">

### OBI
<span class="tag prod">Traces + Metrics</span>
<br><br>
HTTP/gRPC spans
RED metrics
Library-level

</div>

<div class="center">

### otelc
<span class="tag dev">Granular Spans</span>
<br><br>
Function-level
Custom attributes
In-process context

</div>

<div class="center">

### ebpf-profiler
<span class="tag always">Profiles</span>
<br><br>
CPU flame graphs
Allocation stacks
Whole-system

</div>

</div>

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 5
## Benchmark Shootout

---

## Methodology

Same demo service (HTTP + database/sql), same workload (k6, realistic concurrency).
Measured at service level — not microbenchmarks.

Hardware/versions pinned — see `talks/without-a-single-line/demo/bench/README.md`.

> ⚠️ Numbers below are placeholders — run the harness in `demo/` for real data.

---

## Overhead (placeholder — run demo/)

| Approach | p99 latency vs baseline | CPU overhead | Binary size delta |
|----------|------------------------|--------------|------------------|
| **Baseline** | — | — | — |
| **OBI** | TODO | TODO | n/a (no recompile) |
| **otelc** | TODO | TODO | TODO |
| **ebpf-profiler** | TODO (DaemonSet, host-level) | TODO | n/a |

**Key insight:** OBI and ebpf-profiler overhead is *host-level* — it appears as node CPU,
not application latency. Benchmark design matters.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 6
## Decision Framework

---

## Which tool for which context?

<div class="columns">

<div>

### Use OBI when:
<br>

<span class="tag prod">production</span>
<span class="tag prod">k8s / deployed</span>
<span class="tag prod">no rebuild</span>
<span class="tag prod">polyglot fleet</span>

<br><br>

HTTP/gRPC RED metrics + library spans.
Attach in seconds, zero rollout.

</div>

<div>

### Use otelc when:
<br>

<span class="tag dev">local dev</span>
<span class="tag dev">debugging</span>
<span class="tag dev">granular spans</span>
<span class="tag dev">Go 1.25+</span>

<br><br>

Per-function traces, custom attributes,
business-logic visibility.

</div>

</div>

**Always:** `ebpf-profiler` for CPU profiling — it's orthogonal to both.

---

<!-- _class: vcenter center -->

## 🔴 Agent Demo

```
User: "my-service is slow in production, help me debug it"

→ Skill detects "production" context
→ Routes to OBI
→ Runs: obi-integration.sh net/http
→ Fetches only the net/http section of SUPPORT_MATRIX.md
→ Provides: kubectl obi attach command

User: "now I want granular spans locally"

→ Routes to otelc
→ Runs: otelc-aspect.sh net/http
→ Provides: otelc go build command
```

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Part 7
## Runtime Futures

---

## Flight recording — **shipped** in Go 1.25

This is no longer a future. It's available today.

```go
fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
    MinAge:   2 * time.Second,  // ~2x your event window
    MaxBytes: 64 << 20,
})
fr.Start()

// On latency spike:
var buf bytes.Buffer
fr.WriteTo(&buf)
// → buf contains the last 2s of execution trace
```

JFR-style circular buffer. On-demand snapshot. No continuous I/O overhead.

---

## Still missing (active proposals)

**golang/go#75654** — `httptrace.ClientTrace.GotResponseEnd`

> OTel Go currently uses `PutIdleConn` as a workaround.
> `PutIdleConn` is **never called for HTTP/2** — so client spans never close.
> Filed: 2025-09-29. Open.

**golang/go#69887** — `-toolexec` build graph improvements
Orchestrion maintainer (Datadog) filed. Makes otelc/orchestrion less fragile.

---

## USDT: what could be

USDT probes = NOP sled → `INT3` patch at attach time. Near-zero unattached overhead.

Go runtime ships **no built-in USDT probes** (as of Go 1.25, #57175 at initial-inquiry stage).

But a proof-of-concept fork adds them to `net/http`, `database/sql`, `crypto/tls`:

```bash
# github.com/kakkoyun/go/tree/poc_usdt
go tool usdt list ./...
# → probe: net/http.(*conn).serve enter
# → probe: database/sql.(*DB).QueryContext enter
```

If flight recording is Go 1.25 — USDT could be a future Go 1.2X.

---

<!-- _class: vcenter -->

## The mental model

<br>

| When | Tool | Signal |
|------|------|--------|
| **Always, in prod** | OBI (eBPF) | Traces + metrics |
| **Debugging locally** | otelc | Granular spans |
| **Always, CPU** | ebpf-profiler | Profiles |
| **On-demand in prod** | Flight recorder (Go 1.25) | Execution trace |

**One service. Three zero-touch signals. No source changes.**

---

<!-- _class: lead -->
<!-- _paginate: false -->

## Thank you

<br>

**Kemal Akkoyun** · `@kakkoyun`
Datadog · Go & Observability

<br>

**Resources:**
- OBI: `github.com/open-telemetry/opentelemetry-ebpf-instrumentation`
- otelc: `github.com/open-telemetry/opentelemetry-go-compile-instrumentation`
- ebpf-profiler: `github.com/open-telemetry/opentelemetry-ebpf-profiler`
- This talk: `github.com/kakkoyun/gopherconuk-26/tree/main/talks/without-a-single-line`
- Agent skill: `tools/skills/collect-go-telemetry/`

---

<!-- _paginate: false -->
<!-- _class: vcenter center -->

## Questions?

`@kakkoyun` · GitHub · Mastodon · LinkedIn

# otelc — OTel Go Compile-Time Instrumentation

> **Status:** ✅ Research complete (deep-research workflow, 2026-07-22)
> **Blog post:** "otelc: zero-touch Go traces at compile time"
> **Key claims:** C-020 ✅ CONFIRMED · C-021 ⚠️ CORRECTED · C-022 ✅ CONFIRMED

---

## ⚠️ CRITICAL CORRECTION — Read First

**Orchestrion was NOT donated to OpenTelemetry.** Prior research assumed a donation; this is wrong.

There are **two distinct but related tools**:

| | DataDog/orchestrion | open-telemetry/opentelemetry-go-compile-instrumentation |
|---|---|---|
| **CLI binary** | `orchestrion` | `otelc` |
| **Repo** | github.com/DataDog/orchestrion | github.com/open-telemetry/opentelemetry-go-compile-instrumentation |
| **Owner** | Datadog (proprietary vendor tool) | OTel SIG (Datadog + Alibaba co-founded) |
| **Status** | GA v1.11.0 (2026-06-25) | Stable v1.0.1 (2026-07-14) 🔥 **NEW** |
| **Default tracer** | dd-trace-go/v2 (but vendor-agnostic) | OpenTelemetry SDK |
| **Relationship** | Inspiration + SIG co-founder | New tool built from scratch |

The talk uses `otelc` (the OTel SIG tool). Orchestrion is relevant as the inspiration and for the
`dd-trace-go` integration count. Both use the same core mechanism.

---

## 1. Provenance & Project Status

Datadog and Alibaba were independently building compile-time Go instrumentation tools and converged.
They co-founded the **OTel Go Compile-Time Instrumentation SIG** and built `otelc` from scratch —
*"a new next generation of an OpenTelemetry-standard tool inspired by Orchestrion."*

- **DataDog/orchestrion**: Datadog's own tool, vendor-agnostic but defaults to dd-trace-go/v2.
  GA since v1.0.0 (2024-11-26). Current: **v1.11.0 (2026-06-25)**. Repo: `github.com/DataDog/orchestrion`.
- **open-telemetry/opentelemetry-go-compile-instrumentation** (`otelc`): OTel SIG tool.
  First non-retracted stable: **v1.0.1 (2026-07-14)**. Repo: `github.com/open-telemetry/opentelemetry-go-compile-instrumentation`.
  Install: `go install go.opentelemetry.io/otelc/tool/cmd/otelc@latest`

**Sources:** [S-O-01], [S-O-02], [S-O-03], [S-O-04]

---

## 2. How It Works: The Core Mechanism (both tools)

Both tools use the same mechanism, which Orchestrion's maintainers describe as **"compile-time-woven
Aspect-Oriented Programming (AoP)"**:

1. **`-toolexec` proxy**: The tool registers as a Go toolchain proxy via Go's `-toolexec` flag.
   Every invocation of `go tool compile` is intercepted before it runs.
2. **AST rewriting**: Each `.go` source file is parsed into a decorated syntax tree
   (Orchestrion uses [github.com/dave/dst](https://github.com/dave/dst)) and instrumentation is
   woven in at the AST level — *before* the compiler sees the file.
3. **Aspects/rules**: What gets instrumented is controlled by an aspects/rules system.
   For Orchestrion: `orchestrion.tool.go` (imports) and/or `orchestrion.yml` (aspects file).
4. **No runtime agent**: The instrumentation becomes part of the compiled binary. At runtime,
   the instrumented code calls the tracer SDK (dd-trace-go for Orchestrion, OTel SDK for otelc).

From golang/go#69887 (Orchestrion maintainer):
> "It uses -toolexec to intercept all invocations to go tool compile, re-writing all the .go files
> to add instrumentation everywhere possible. Orchestrion can in some ways be seen as a
> compile-time-woven Aspect-Oriented Programming (AoP) framework."

From the OTel blog on otelc:
> "hooks into the standard Go toolchain during the build (through its `-toolexec` mechanism) and
> injects OpenTelemetry instrumentation into your code, its dependencies, and the standard library
> as they are compiled."

**Sources:** [S-O-03], [S-O-04], [S-O-05]

---

## 3. The Goroutine-Local Storage Hack (⚠️ STILL PENDING)

The prior research lead claims Orchestrion fakes goroutine-local storage via `go:linkname`,
injecting a synthetic field into the runtime `g` struct at `internal/orchestrion/gls.go`.

**Verdict: PENDING — not confirmed by deep-research web search.** The deep-research agents stalled
on this question. This requires a direct read of the DataDog/orchestrion repo source.

**Action required:** Read `internal/orchestrion/gls.go` (or equivalent) directly from the repo to
verify the exact `go:linkname` usage and what it injects. Pin commit SHA.

This is the "killer proof-point" for the thesis — the most technically striking detail in the talk.
Do not assert it until confirmed with a primary source read.

*See `claims-ledger.md` entry C-002 — remains PENDING.*

---

## 4. What It Instruments

### otelc (OTel SIG) — current supported list
**TODO:** Verify the complete supported frameworks list from the `opentelemetry-go-compile-instrumentation`
repo's current release. The deep-research web search did not return a definitive list for otelc specifically.
Check: `github.com/open-telemetry/opentelemetry-go-compile-instrumentation` — look at the
instrumentation packages or the README's supported integrations table.

### DataDog/orchestrion + dd-trace-go v2 — confirmed list
Verified from `contrib/supported_integrations.md` in DataDog/dd-trace-go v2 repo:

**HTTP/Frameworks:**
- Gin, Gorilla Mux, chi, echo v4, Fiber, net/http

**RPC:**
- gRPC

**Databases:**
- database/sql, sqlx, MongoDB (mongo-driver v1+v2), Gorm v2
- Redis: go-redis v6-v9, redigo, rueidis, valkey-go
- Cassandra: gocql
- Memcache

**GraphQL:**
- graph-gophers, graphql-go, gqlgen

**Cloud / Infrastructure:**
- AWS SDK v1 (deprecated) + v2
- Kafka: confluent-kafka-go v1+v2, IBM/sarama (Shopify/sarama deprecated)
- Vault, Consul

**Kubernetes:**
- client-go

**Sources:** [S-O-06]

---

## 5. CI / Build Integration

Usage pattern for both tools is the same: prefix `go build` with the tool binary.

```bash
# otelc
otelc go build -o myapp .

# orchestrion
orchestrion go build .
```

This works with any Go build system that supports `GOFLAGS` or custom build commands. It does not
require `go generate`, build tags, or code changes in the project.

CI pipeline integration: set `GOTOOLCHAIN` or add the tool to `$PATH` and prefix the build command.

**Sources:** [S-O-03], [S-O-04]

---

## 6. Version & Go Compatibility

| Tool | Current version | Go min version |
|------|----------------|----------------|
| DataDog/orchestrion | v1.11.0 (2026-06-25) | Verify from go.mod |
| otelc | v1.0.1 (2026-07-14) | **Go 1.25+** (confirmed from README badge) |

**otelc requires Go 1.25+** — confirmed from the README badge `Go-1.25%2B`.
This is a significant constraint for the talk: services on Go 1.23/1.24 cannot use otelc.
For those, OBI or Orchestrion (check its go.mod) are the alternatives.

**TODO:** Check DataDog/orchestrion `go.mod` for minimum Go version.

---

## 7. Orchestrion is Vendor-Agnostic

From the DataDog/orchestrion README (HEAD df04ed94b69e, 2026-07-06):
> "Orchestrion is a vendor-agnostic tool. By default, `orchestrion pin` enables Datadog's tracer
> integrations by importing `github.com/DataDog/dd-trace-go/v2` in `orchestrion.tool.go`, but other
> vendors (such as OpenTelemetry) may provide alternate integrations that can be used instead."

This is relevant for the talk: Orchestrion can wire in OTel SDK too — it's not locked to Datadog.

---

## 8. Fit With the Talk

**Talk angle**: otelc (OTel SIG) is the zero-touch compile-time story. Orchestrion is the
"inspiration / predecessor" that shows the same idea at production scale (Datadog uses it internally).

**Decision framework**:
- otelc → local dev, granular OTel spans, requires `go build` with `otelc` wrapper, no runtime overhead
- Orchestrion → same mechanism, battle-tested, defaults to dd-trace-go, vendor-agnostic

**Relationship to the thesis**: The `-toolexec` mechanism is a *compile-time workaround* for Go
having no runtime hook point. The hack (go:linkname / g-struct) is the evidence that even compile-time
approaches must reach into Go internals to support goroutine-aware tracing.

---

## Open Questions (before talk/blog)

- [ ] Read `internal/orchestrion/gls.go` directly in DataDog/orchestrion repo — confirm go:linkname + g-struct claim. Pin SHA. Update C-002.
- [ ] Get complete otelc supported frameworks list from current release.
- [ ] Verify minimum Go version for both tools (from go.mod).
- [ ] Does otelc instrument the standard library too? (OTel blog says "your code, its dependencies, and the standard library" — verify what stdlib packages are covered.)
- [ ] What is otelc's supported integrations list vs Orchestrion's? Are they converging?

---

## Sources Used

| Key | Description | URL |
|-----|-------------|-----|
| S-O-01 | Datadog open-source page: Orchestrion | https://opensource.datadoghq.com/projects/orchestrion/ |
| S-O-02 | golang/go#69887 (OTel SIG context from Orchestrion maintainer) | https://github.com/golang/go/issues/69887 |
| S-O-03 | DataDog/orchestrion repo (HEAD df04ed94b69e, 2026-07-06) | https://github.com/DataDog/orchestrion |
| S-O-04 | OTel blog: go-compile-time-instrumentation-v1 (2026-07-16) | https://opentelemetry.io/blog/2026/go-compile-time-instrumentation-v1/ |
| S-O-05 | open-telemetry/opentelemetry-go-compile-instrumentation | https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation |
| S-O-06 | DataDog/dd-trace-go contrib/supported_integrations.md | https://github.com/DataDog/dd-trace-go |

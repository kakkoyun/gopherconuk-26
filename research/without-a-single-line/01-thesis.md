# The Thesis: Why Go Is Hard to Instrument Without Source Changes

> **Status:** ✅ Research complete (deep-research workflow, 2026-07-22)
> **Blog post:** "Why Go can't be monkey-patched (and what people do about it)"
> **Key claims:** C-001 ✅ CONFIRMED · C-002 ⚠️ PARTIALLY CONFIRMED · C-003 ✅ CONFIRMED (corrected count)

---

## 1. Go Has No Runtime Hook Point

**Verdict: CONFIRMED (C-001)**

Go is structurally hostile to zero-touch instrumentation:
- Compiles to **native machine code** — no intermediate bytecode layer to rewrite at load time.
- **No classloader** — no analog to Java's `ClassLoader.defineClass` where you can intercept and rewrite classes.
- **Links statically by default** — no shared library resolution phase where you could intercept symbols.
- **No general runtime hook API** — the only internal exit-time hook (`internal/runtime/exithook`) is scoped strictly to program termination.

The `go:linkname` stubs in `gopark` and `goready` in `runtime/proc.go` exist only because third-party packages abused them. The Go team's own comment on those stubs:

> "gopark should be an internal detail, but widely used packages access it using linkname.
> Notable members of the hall of shame include…"

This is not a supported hook — it's a reluctant accommodation of prior abuse.

**Source:** [S-TH-01] (`runtime/proc.go` in Go master), [S-TH-02] (`internal/runtime/exithook/hooks.go`)

---

## 2. LD_PRELOAD Is Blocked for Go Binaries

**Verdict: CONFIRMED (part of C-001)**

LD_PRELOAD is the classic escape hatch on Linux for injecting into processes. For Go:

- **Internal linker (default):** Go's internal linker does not invoke the dynamic linker's
  LD_PRELOAD mechanism at all. LD_PRELOAD has zero effect on internally-linked Go binaries.
- **External linker (`-linkmode=external`):** LD_PRELOAD can work, but requires forcing external
  linking (Go core contributor Ian Lance Taylor: *"anyone who wants to use LD_PRELOAD must force
  the use of external linking"* — golang/go#28909).
- **Static cgo binaries:** Runtime-injection vendors (e.g., Dynatrace OneAgent) explicitly reject
  monitoring of static Go binaries that use cgo.
- **`CGO_ENABLED=0` + `-buildmode=pie`:** Produces a binary with no libc dependency. libc is
  required by injection agents — this combination is structurally incompatible with injection.

**Source:** [S-TH-03] (golang/go#28909, Ian Lance Taylor), [S-TH-04] (Dynatrace Go known limitations)

---

## 3. The GLS Hack: How Orchestrion Fakes Goroutine-Local Storage

**Verdict: CONFIRMED ✅ (C-002)**

The prior lead was correct — Orchestrion does inject a synthetic field into `runtime.g`. The
complete mechanism is defined by a YAML aspect file in dd-trace-go:
**`internal/orchestrion/gls.orchestrion.yml`** (commit `b97e7cbb4`, v2.0.0).

### The three-step mechanism

**Step 1 — Inject a field into `runtime.g`:**

```yaml
# internal/orchestrion/gls.orchestrion.yml
join-point:
  struct-definition: runtime.g
advice:
  - add-struct-field:
      name: __dd_gls_v2
      type: any
```

Orchestrion rewrites the Go runtime source to add `__dd_gls_v2 any` as a field of the internal
`runtime.g` struct. Every goroutine gets its own copy — this is what makes it goroutine-local.

**Step 2 — Inject `go:linkname` get/set symbols that access the field:**

```go
// Injected by Orchestrion into the runtime package at compile time:
//go:linkname __dd_orchestrion_gls_get __dd_orchestrion_gls_get.V2
var __dd_orchestrion_gls_get = func() any {
    return getg().m.curg.__dd_gls_v2   // getg() → current g → M → curg field
}

//go:linkname __dd_orchestrion_gls_set __dd_orchestrion_gls_set.V2
var __dd_orchestrion_gls_set = func(val any) {
    getg().m.curg.__dd_gls_v2 = val
}
```

`getg().m.curg` navigates: current goroutine `g` → OS thread M → currently running goroutine.

**Step 3 — Clean up on goroutine exit** (prevents memory leaks):

```go
// Injected into runtime.goexit1:
getg().__dd_gls_v2 = nil
```

**dd-trace-go's consumer side** (`internal/orchestrion/gls.go`, commits `b97e7cbb4` + `577c7760f`):

```go
//go:linkname __dd_orchestrion_gls_get __dd_orchestrion_gls_get.V2
var __dd_orchestrion_gls_get func() any   // nil when Orchestrion absent

//go:linkname __dd_orchestrion_gls_set __dd_orchestrion_gls_set.V2
var __dd_orchestrion_gls_set func(any)

func init() {
    if __dd_orchestrion_gls_get != nil && __dd_orchestrion_gls_set != nil {
        getDDGLS = __dd_orchestrion_gls_get
        setDDGLS = __dd_orchestrion_gls_set
    }
}
```

When Orchestrion is absent, both stay nil and the tracer silently degrades to no-op GLS.

### Why this is fragile

`go:linkname` for **variables** has no definition/reference separation. When both sides are
uninitialized BSS symbols, the linker's choice is **arbitrary and order-dependent**. This broke
between Go 1.22 and 1.23 (golang/go#72032). The aspect YAML itself references the Go runtime
`HACKING.md` — the Go team considers this access pattern deeply unofficial.

**Files (both in DataDog/dd-trace-go):**
- `internal/orchestrion/gls.orchestrion.yml` — commit `b97e7cbb4` (v2.0.0), `c9ff7bbe2`
- `internal/orchestrion/gls.go` — commits `b97e7cbb4` (v2.0.0), `577c7760f`

**Source:** [S-TH-05a] (`gls.orchestrion.yml`), [S-TH-05b] (`gls.go`), [S-TH-06] (golang/go#72032)

---

## 4. dd-trace-go Integration Count

**Verdict: CONFIRMED (C-003) — prior count was wrong**

The `contrib/` directory in DataDog/dd-trace-go v2 has **51 subdirectories** (not ~54 as the
prior lead stated). Do not use 54; use the actual current count or link to the supported
integrations list.

**Source:** [S-TH-07] (dd-trace-go `contrib/` directory, counted by deep-research agent)

---

## 5. The Three Families of Zero-Touch Approaches

**Verdict: CONFIRMED (three families exist; production-readiness assessed)**

| Family | Mechanism | Production-ready? | Examples |
|--------|-----------|-------------------|----------|
| **eBPF uprobes** | Kernel-level probes on function entry/exit | ✅ Yes (with caveats) | OBI |
| **Compile-time AST rewriting** | `-toolexec` intercepts `go tool compile`, rewrites AST | ✅ Yes (most broadly) | Orchestrion (v1.11.0), otelc (v1.0.1) |
| **Runtime injection** | LD_PRELOAD / binary trampolines / shared lib injection | ⚠️ Limited (external linking only) | Dynatrace OneAgent |

Compile-time rewriting is currently the **most broadly production-ready** for Go:
- Works across Go versions (with caveats around go:linkname stability).
- No kernel version requirement.
- Instruments arbitrary framework code, not just network-level calls.
- Dynatrace's approach requires `-linkmode=external` and has static binary limitations.

**Source:** [S-TH-03], [S-TH-04], [S-TH-08]

---

## 6. Talk Spine

The thesis is the **opening argument**: before showing OBI, otelc, and ebpf-profiler, the audience
needs to understand *why* zero-touch Go is hard. The structure:

1. Open: "Go can't be monkey-patched." (15 seconds)
2. Why: native code, no classloader, no bytecode, no LD_PRELOAD (without tricks). (2 min)
3. The hall of shame: even the Go team reluctantly accommodates go:linkname abuse. (30 seconds)
4. Three workaround families: eBPF from outside, compile-time from the build system, injection with caveats. (1 min)
5. Transition: "Let's look at what's production-ready today…" → OBI.

---

## Open Questions (before talk/blog)

- [ ] **CRITICAL:** Read Orchestrion repo to find where `__dd_orchestrion_gls_get.V2` / `__dd_orchestrion_gls_set.V2` are defined. What is the actual implementation? Does it involve the `g` struct? Pin file path and commit SHA.
- [ ] Verify dd-trace-go `contrib/` count at the exact v2 release tag (not HEAD). The count of 51 is from the deep-research agent's local read — pin the tag.
- [ ] Check golang/go#72032 status: has a fix landed? Which Go version?
- [ ] Is Dynatrace OneAgent the only meaningful runtime-injection player for Go, or are there others?

---

## Sources Used

| Key | Description | URL / Location |
|-----|-------------|----------------|
| S-TH-01 | runtime/proc.go "hall of shame" comment (Go master) | https://go.dev/src/runtime/proc.go |
| S-TH-02 | internal/runtime/exithook/hooks.go (Go master) | internal package — verify via Go source |
| S-TH-03 | golang/go#28909 (Ian Lance Taylor on LD_PRELOAD) | https://github.com/golang/go/issues/28909 |
| S-TH-04 | Dynatrace Go known limitations | https://docs.dynatrace.com/docs/ingest-from/technology-support/application-software/go/support/go-known-limitations |
| S-TH-05 | dd-trace-go internal/orchestrion/gls.go (locally verified) | https://github.com/DataDog/dd-trace-go — commits b97e7cbb4, 577c7760f |
| S-TH-06 | golang/go#72032 (go:linkname variable fragility) | https://github.com/golang/go/issues/72032 |
| S-TH-07 | dd-trace-go contrib/ count (51 subdirs) | https://github.com/DataDog/dd-trace-go |
| S-TH-08 | FOSDEM 2026 talk summary (kakkoyun.me) | https://kakkoyun.me/posts/fosdem-2026-auto-instrumenting-go/ |

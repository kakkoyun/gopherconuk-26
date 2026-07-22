# Go Runtime Futures

> **Status:** ✅ Research complete (deep-research workflow, 2026-07-22)
> **Blog post:** "Go runtime futures: flight recording, USDT, and hook proposals"
> **Key claims:** C-040 ✅ CONFIRMED (SHIPPED) · C-041 ✅ CONFIRMED · C-042 ✅ CONFIRMED · C-043 PLAUSIBLE

---

## ⚠️ MAJOR UPDATE — Flight Recording Shipped in Go 1.25

**golang/go#63185 is not a future — it shipped in Go 1.25.** The API also differs from the prior
lead. Do not use `SetMinAge` in the talk; use the correct struct-based API shown below.

---

## 1. golang/go#63185 — Flight Recording (SHIPPED, Go 1.25)

- **Title:** proposal: runtime/trace: flight recording
- **Status:** CLOSED — Proposal-Accepted, milestoned Go 1.25
- **Proposer:** mknyszek (Michael Knyszek, Go runtime team)
- **Blog post:** https://go.dev/blog/flight-recorder (published 26 Sept 2025)

### What it is

A JFR-style (Java Flight Recorder-inspired) **circular-buffer tracer**. Instead of streaming
trace data to a file, it buffers the last few seconds of execution trace data in memory. On
demand, `WriteTo` performs a point-in-time snapshot.

From the issue body: *"trace data is kept in a conceptual circular buffer, flushed upon request."*
From the blog: *"instead of writing it out to a socket or a file, it buffers the last few seconds
of the trace in memory."*

### Shipped API (Go 1.25) — NOT the proposal draft

```go
// Constructor — takes a config struct, NOT setter methods
fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
    MinAge:   2 * time.Second, // ~2x the event window being debugged
    MaxBytes: 64 * 1024 * 1024,
})

// Methods on *FlightRecorder
fr.Start() error
fr.Stop()
fr.Enabled() bool
fr.WriteTo(w io.Writer) (n int64, err error)
```

**Critical:** The prior lead mentioned `SetMinAge` and `SetMaxBytes` as methods — this was the
original *proposal* shape. The shipped implementation uses `FlightRecorderConfig` struct fields.
Do not use the old API in slides or blog posts.

**Constraints:**
- Only one goroutine may call `WriteTo` at a time (returns error if concurrent).
- Returns error if the recorder is inactive.
- `MinAge` should be ~2x the time window of the event being debugged.

### Talk angle

This is no longer a "future" — demo it with Go 1.25! Live flight recorder dump showing a
latency spike is a compelling demo. Pair with OBI for production: "OBI tells you *what* is slow;
flight recording tells you *why* at the trace level, on demand."

**Sources:** [S-RF-01], [S-RF-02], [S-RF-03]

---

## 2. golang/go#69887 — Compile-Time Toolexec Improvements (OPEN)

- **Title:** proposal: cmd/go: compile-time instrumentation and -toolexec
- **Status:** OPEN — project: Proposals Incoming, no acceptance verdict
- **Filed:** 2024-10-15 by Romain Marcadier (RomainMuller, DataDog/Orchestrion)

### What it proposes

Five specific improvements to make `-toolexec` tooling (Orchestrion, otelc) less fragile:

1. **Per-package build-ID influence** — allow -toolexec tools to inject a cache key per package
2. **New build-graph edges** — let -toolexec tools declare inter-package dependencies
3. **Improved `go/ast` comment/directive API**
4. **Improved source mapping for stacked transforms**
5. **`-toolexec` respecting `-p n` parallelism**

### The two root gaps documented

From the issue body (verbatim):
> "The only hook point available is intercepting the -V=full invocation of all tools... The
> drawback of being able to do this only at the complete build level is that it'll result in
> excessively frequent cache invalidation for untransformed packages."

> "Today, the toolchain does not provide any visibility on the full build arguments and Orchestrion
> has to crawl the process tree in search for a `go build` invocation, gather its arguments list,
> and do its best at parsing it."

Go team member aclements engaged substantively, confirming the gaps are real.

### Talk angle

This explains *why* Orchestrion and otelc have rough edges — the `-toolexec` hook was not
designed for instrumentation. #69887 is what the community is pushing to fix it.

**Sources:** [S-RF-04]

---

## 3. golang/go#75654 — httptrace Response-End Hook (OPEN — ACTIVE)

- **Title:** proposal: net/http/httptrace: hook for response completion (connection released)
- **Status:** OPEN — only `Proposal` label, no acceptance decision
- **Filed:** 2025-09-29 by starkross (Rost Khaniukov)

### The gap

No reliable, protocol-agnostic hook exists for when an HTTP client response body is fully read.

- `PutIdleConn` is the current OTel workaround — but **PutIdleConn is never called for HTTP/2 or
  HTTP/3**, causing OTel Go client spans to **never finish** on HTTP/2 (tracked as
  [opentelemetry-go-contrib#4876](https://github.com/open-telemetry/opentelemetry-go-contrib/issues/4876),
  still open).
- Related: closed issue #16400 (2016) sought the same hook — gap unresolved for 9+ years.

### Proposed API

```go
// Add to httptrace.ClientTrace:
GotResponseEnd func(err error)
// Fires exactly once per request when resp.Body.Read returns io.EOF,
// a non-nil error, or the body is closed early.
```

### Talk angle

Perfect hook-gap example for the "Go still needs better hooks" slide. Live production impact:
OTel spans silently never close on HTTP/2 today.

**Sources:** [S-RF-05], [S-RF-06]

---

## 4. golang/go#38270 — Server-Side httptrace (OPEN — STALLED)

- **Title:** proposal: net/http, net/http/httptrace: add mechanism for tracing request serving
- **Status:** OPEN — Proposal-Hold, milestone: Unplanned
- **Filed:** 2020-04-06 by CAFxX (Carlo Alberto Ferraris)

### The gap

No hook for **server-side** HTTP request lifecycle tracing (as opposed to #75654's client-side gap).
The issue has had no substantial Go team engagement and is milestoned Unplanned.

**Different class from #75654:** #75654 = client response-end; #38270 = server-side request
lifecycle. Both are httptrace gaps, but for opposite sides of the HTTP connection.

**Talk angle:** Mention briefly — shows the pattern: instrumentation hooks are missing on both sides.

**Sources:** [S-RF-04] (mentioned in research; verify URL directly before citing)

---

## 5. golang/go#67120 — runtime/metrics Recommended Set (OPEN — INCOMING)

- **Status:** OPEN — project: Proposals Incoming, not yet accepted
- **Notes:** Proposes a curated "recommended set" of runtime/metrics metrics for observability tools
  to surface by default. Relevant to zero-touch observability: a standard recommended set would make
  it easier for OBI and otelc to know which runtime metrics to expose without per-tool decisions.

**Action required:** Verify exact title and status from https://github.com/golang/go/issues/67120
before citing in talk/blog.

**Sources:** [S-RF-04] (verify against primary)

---

## 6. golang/go#73798 — Goroutine-Start Hook (CLOSED)

- **Status:** CLOSED — due to inactivity
- **Notes:** Proposed a hook for goroutine start/end events. Closed without acceptance — not in Go's
  roadmap. Do not cite as an active future.

---

## 7. Summary: The Horizon

| Issue | Status | In Go? | Talk relevance |
|-------|--------|--------|----------------|
| #63185 flight recording | ✅ SHIPPED Go 1.25 | Yes — demo it | High |
| #69887 -toolexec improvements | OPEN, Incoming | No | Medium (explains rough edges) |
| #75654 httptrace response-end | OPEN, active | No | High (live impact: HTTP/2 spans never close) |
| #38270 server-side httptrace | OPEN, Unplanned | No | Low (mention pattern) |
| #67120 runtime/metrics set | OPEN, Incoming | No | Low |
| #73798 goroutine hook | CLOSED | No | Skip |

---

## Sources Used

| Key | Description | URL |
|-----|-------------|-----|
| S-RF-01 | golang/go#63185 (flight recording issue) | https://github.com/golang/go/issues/63185 |
| S-RF-02 | go.dev blog: flight-recorder (Go 1.25, Sept 2025) | https://go.dev/blog/flight-recorder |
| S-RF-03 | runtime/trace/flightrecorder.go (Go master) | https://raw.githubusercontent.com/golang/go/master/src/runtime/trace/flightrecorder.go |
| S-RF-04 | golang/go#69887 (toolexec improvements) | https://github.com/golang/go/issues/69887 |
| S-RF-05 | golang/go#75654 (httptrace response-end) | https://github.com/golang/go/issues/75654 |
| S-RF-06 | opentelemetry-go-contrib#4876 (HTTP/2 span never closes) | https://github.com/open-telemetry/opentelemetry-go-contrib/issues/4876 |

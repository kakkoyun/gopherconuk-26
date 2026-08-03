# Speaker Notes: Without a Single Line

## 1. Hook (2 min)

Ask the question out loud, then wait — raise your own hand first, and hold the silence for two or three full seconds while the room settles. Let them feel the weight of it: every raised hand is a person who has spent real hours wrapping HTTP clients and DB calls in boilerplate they didn't want to write. Then drop the turn: "What if you didn't have to?" — and move on without explaining it yet.

## 2. The Problem (3 min)

Read the `go:linkname` hall-of-shame comment verbatim from the slide, slowly — "the Go team's comment on abuse" lands harder when the audience hears the actual words rather than a paraphrase. The point to leave them with: this comment exists because Go has no designed hook point at all; the "hall of shame" is the language's reluctant acknowledgment that people need one. Every other runtime gives you bytecode rewriting, a classloader, a `sys.settrace` — Go gives you nothing, and that's not an accident, it's a design choice we have to work around.

## 3. OBI (7 min)

When the demo terminal opens, narrate what's happening in real time: "OBI is now attaching uprobes to the running process — no restart, no recompile, nothing touched in the application repository." The key correction to make verbally: "zero code changes" has a precise scope — HTTP and gRPC RED metrics plus spans for 13 specific Go libraries, full stop; if you need a custom span carrying a user ID or a SQL query parameter, OBI cannot give you that, and I'll show you what can in a moment.

## 4. otelc (7 min)

When you reach the goroutine-local storage YAML slide, slow down — this is the technical peak of the talk. Say out loud: "What you're looking at is an aspect that patches the `runtime.g` struct at compile time to add a field that Go itself doesn't expose. This is not a hack around the language; it is the only way to get per-goroutine context without the language providing it." Before leaving Part 3, make the separation explicit: orchestrion is Datadog's production-grade tool, otelc is the OTel SIG's vendor-neutral implementation — they share the same `-toolexec` mechanism, but they are two separate projects with different release cadences, and otelc requires Go 1.25 or later.

## 5. ebpf-profiler (5 min)

After the `.gopclntab` slide appears, pause and give the audience a moment before continuing — the claim that symbolization works on a fully stripped static production binary surprises most people. Explain why it's true: Go's own runtime needs the PC-to-function table to unwind goroutine stacks and format panics, so the linker keeps `.gopclntab` even when you strip every other debug section, and the ebpf-profiler reads it from process memory via eBPF without ever touching the binary on disk. That's not a quirk to work around — it's a structural guarantee you can rely on in production.

## 6. Benchmark (3 min)

Flag the numbers immediately: "These are placeholders from a controlled lab environment — run the harness in `demo/bench/` against your own service before making architectural decisions." The framing point that matters more than any specific number: OBI and ebpf-profiler overhead shows up at the node level, not in your application's p99 latency — you'll see it in your Kubernetes node CPU budget, not in your SLO dashboard, and that distinction changes how you think about the cost.

## 7. Decision Framework + Agent Demo (2 min)

Narrate the routing decision the skill is making as it runs: "It sees 'production service, no rebuild window' and routes to OBI — attach in seconds, zero rollout. When I say 'now I want granular spans locally', it switches to otelc — different tool, different context, same zero-source-change principle." The point is not the demo itself; it's that the decision is deterministic and teachable: production with no rebuild → OBI; local dev needing business-logic visibility → otelc; always-on CPU profiling with no overhead budget → ebpf-profiler.

## 8. Runtime Futures (1 min)

Do not end on "someday maybe" — end on "this is now." Flight recording shipped in Go 1.25: `trace.NewFlightRecorder`, JFR-style circular buffer, on-demand snapshot, no continuous I/O. The USDT proof-of-concept is the speaker's own fork, so you can say directly: "I've run this, it works, the question is whether the Go team wants to ship it." Leave the room with the mental model on the final slide: one service, three zero-touch signals, no source changes — and then take questions.

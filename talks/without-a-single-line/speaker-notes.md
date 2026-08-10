# Speaker Notes: Without a Single Line

## 0. Ethos slide (< 1 min)

Twenty seconds. Do not read the slide aloud. Land one sentence: "I'm on the Prometheus
Steering Committee — which is to say, someone other than me decided I know something about
Go observability." Then move directly to the show-of-hands question.

Credibility must land before the audience ask, or the ask falls flat.

---

## 1. Hook (2 min)

Ask the question out loud, then wait — raise your own hand first, and hold the silence for two or three full seconds while the room settles. Let them feel the weight of it: every raised hand is a person who has spent real hours wrapping HTTP clients and DB calls in boilerplate they didn't want to write. Then drop the turn: "What if you didn't have to?" — and move on without explaining it yet.

## 2. The Problem (3 min)

Read the `go:linkname` hall-of-shame comment verbatim from the slide, slowly — "the Go team's comment on abuse" lands harder when the audience hears the actual words rather than a paraphrase. The point to leave them with: this comment exists because Go has no designed hook point at all; the "hall of shame" is the language's reluctant acknowledgment that people need one. Every other runtime gives you bytecode rewriting, a classloader, a `sys.settrace` — Go gives you nothing, and that's not an accident, it's a design choice we have to work around.

## 3. OBI (7 min)

When the demo terminal opens, narrate what's happening in real time: "OBI is now attaching uprobes to the running process — no restart, no recompile, nothing touched in the application repository." The key correction to make verbally: "zero code changes" has a precise scope — HTTP and gRPC RED metrics plus spans for 13 specific Go libraries, full stop; if you need a custom span carrying a user ID or a SQL query parameter, OBI cannot give you that, and I'll show you what can in a moment.

## 4. otelc (7 min)

When you reach the goroutine-local storage YAML slide, slow down — this is the technical peak of the talk. Say out loud: "What you're looking at is an aspect that patches the `runtime.g` struct at compile time to add a field that Go itself doesn't expose. This is not a hack around the language; it is the only way to get per-goroutine context without the language providing it." On the convergence slide, land the key point: Orchestrion and the Alibaba tool didn't compete — they both proposed donation to OTel, and the SIG built otelc as a unified, vendor-neutral codebase taking the best of both. Orchestrion remains active as a standalone Datadog project. otelc is v1 stable; it requires Go 1.25 or later — a hard gate worth stating explicitly.

## 5. ebpf-profiler (5 min)

After the `.gopclntab` slide appears, pause and give the audience a moment before continuing — the claim that symbolization works on a fully stripped static production binary surprises most people. Explain why it's true: Go's own runtime needs the PC-to-function table to unwind goroutine stacks and format panics, so the linker keeps `.gopclntab` even when you strip every other debug section, and the ebpf-profiler reads it from process memory via eBPF without ever touching the binary on disk. That's not a quirk to work around — it's a structural guarantee you can rely on in production.

## 6. Benchmark (3 min)

Flag the numbers immediately: "These are from a controlled environment — your production numbers will differ based on workload shape, kernel version, and concurrency." The framing that matters more than any specific number: OBI and ebpf-profiler overhead appears at the node level, not in your application's p99 latency — it shows up in your Kubernetes node CPU budget, not your SLO dashboard. That distinction changes how you think about the cost.

## 7. Decision Framework + Agent Demo (2 min)

On the "Which tool for which context?" slide, call out the maturity difference explicitly before going to the demos: OBI is v0 — emitted telemetry fields can change between minor releases; otelc is v1 stable. That matters if someone is building dashboards they want to last. On the "They work at different layers" slide, the key phrase is "complementary, not competing" — OBI sees what crosses the boundary, otelc sees what happens inside. If you have a Go service in a mixed-language fleet, use both. Narrate the agent demo: "It sees 'production service, no rebuild window' → routes to OBI. 'Granular spans locally' → routes to otelc. The routing is deterministic and teachable."

## 8. The Horizon (1 min)

Two concrete things to land. First, USDT: Go ships no built-in USDT probes today, but the proof-of-concept is the speaker's own fork — "I've run this, it works." The argument for USDT is architectural: stable, named hook points mean OBI, ebpf-profiler, debuggers, and injectors all get a reliable contract instead of chasing symbol addresses that change with every build. Second, Live Debugger: this exists today, it uses eBPF via Datadog's system-probe, and it adds log lines to production Go services without a restart. Bits Live Debugger is in preview and lets you describe a bug — the AI places logpoints on the running service and reads real production data. End on that image: one service, one description of a bug, real data from production, no code change.

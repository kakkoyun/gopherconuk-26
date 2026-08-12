# Speaker notes: Zero-Touch Go Instrumentation

GopherCon UK 2026 · 30 minutes · advanced audience

Deck: `slides/presentation.md` (52 slides)

## Timing

This is the additive review version. Its scripted estimate is 30:30, so it remains slightly over the slot until Kemal chooses the pruning pass.

| Block | Slides | Target | Running total |
| --- | ---: | ---: | ---: |
| Story and reason to care | 2-5 | 2:15 | 2:15 |
| Personal and Datadog context | 6-7 | 1:15 | 3:30 |
| Why Go resists instrumentation | 8-13 | 4:00 | 7:30 |
| OBI | 14-24 | 6:30 | 14:00 |
| otelc and Orchestrion | 25-35 | 6:45 | 20:45 |
| Profiles and request correlation | 36-43 | 4:45 | 25:30 |
| Decision framework | 44-46 | 2:00 | 27:30 |
| USDT, Live Debugger, results, and CTA | 47-52 | 3:00 | 30:30 |

Checkpoints: open OBI by 7:30, otelc by 14:00, profiles by 20:45, and the decision section by 25:30.

Do not delete slides during this additive pass. The likely rehearsal cuts are spoken detail on the timeline, the full library enumeration, and the second `.gopclntab` proof slide. Never cut the uprobe caveat, OTEP correlation argument, or final decision model without Kemal's review.

## Slides 2-5: story and reason to care

### Slide 2: show of hands

Ask the question and raise your own hand. Wait for two or three seconds. Do not rescue the silence.

Say: "Every hand is time spent wrapping code that already knew how to make an HTTP request or a database call."

### Slide 3: the turn

Say only: "What if you did not have to?" Let the question stand and advance.

### Slide 4: how coverage decays

Keep this concrete. The first service has an owner who remembers every wrapper. A growing fleet adds clients, queues, framework upgrades, and new teams. Missing instrumentation becomes normal even when every individual decision was reasonable.

The line to land: "At fleet scale, instrumentation coverage is a platform problem."

### Slide 5: why Go is the exception

Do not teach every runtime mechanism. Point out that each has a supported place to intervene after source code was written. Go compiles the place away.

Transition: "Before I show the workarounds, you should know why I care about this problem."

## Slides 6-7: personal and Datadog context

### Slide 6: personal context

Twenty seconds. Consequences, not achievements. The slide now carries the
consequences, so you can say them rather than translate a credentials list.

"I maintained `client_golang`, which means I wrapped handlers by hand for years.
Then Parca, where the whole point was a profiler nobody has to install. Now
OpenTelemetry compile-time instrumentation, which ships inside your binary."

"Every one of those jobs taught me the same thing: hand-written instrumentation
decays. That is what this talk is about."

The standing line is small print at the bottom deliberately. Do not read it out.
Do not mention contribution counts. Do not imply work on OBI.

### Slide 7: Datadog context

This slide is about the company's reason to care, not a second biography.

Say: "Datadog builds at all three intervention points. We have compile-time instrumentation, startup injection, and kernel-level instrumentation and profiling. I do not need one approach to win. I need the right combinations to work in customer services."

## Slides 8-13: why Go resists instrumentation

### Slides 9-10: native code and the hall of shame

The lack of bytecode and a classloader is the main point. The `go:linkname` comment is evidence that users need hooks Go did not design.

Read only the phrase "hall of shame." Do not read the whole code block.

### Slide 11: `LD_PRELOAD`

Be precise. Do not say `LD_PRELOAD` never works with Go. It works when the build uses the external linker and a compatible dynamic path. It does not reach the default internally linked binary. Static cgo and pure-Go PIE builds narrow the surface further.

### Slide 12: lifecycle map

This is the map for the rest of the talk. Point to the three positions:

- Build time preserves language semantics but requires a rebuild.
- Process-start injection avoids source edits but depends on binary and linker surfaces.
- Kernel observation avoids the rebuild but inherits Linux, privilege, and compatibility constraints.

### Slide 13: injector distinction

Name the products clearly:

- Dynatrace supports eligible dynamically linked Go binaries.
- The OpenTelemetry host injector packages Java, .NET, Node.js, and Python. It does not package a Go agent.
- The OpenTelemetry Operator's Go option is different: an alpha eBPF sidecar behind a feature gate.
- Datadog is working on a Go Single-Step Instrumentation path.

Do not promise a public release date.

## Slides 14-24: OBI

### Slides 15-17: what OBI changes operationally

OBI is the upstream successor to Beyla. The application does not rebuild or restart when a node-level OBI deployment attaches. The operational change is the privileged observer running beside the workload.

On the punchline slide, pause after each line. Then add: "The DaemonSet still has a deployment and security cost. Zero-touch describes the application, not the platform."

### Slide 18: supported Go libraries

Say "thirteen documented baselines" and point to a few categories. Do not read all thirteen. Call out the corrected names only if useful:

- `github.com/redis/go-redis/v9`, not the old v8 path.
- gin excludes v1.7.5.

The support matrix is the contract. Library internals can change faster than a slide deck.

### Slide 19: logs

OBI emits traces and metrics. Its log feature enriches selected JSON writes with trace and span IDs while a span is active. Existing log shipping moves those bytes to a backend.

Say explicitly: "Correlation is log support, but OBI is not a log exporter."

Trace-log correlation has stricter requirements than baseline OBI, including Linux 6.0 or newer in current documentation. Keep that detail in reserve unless asked.

### Slides 20-21: platform and privileges

Do not read the capability table as a universal list. Explain the ladder:

1. Network capture needs BPF and raw-socket access.
2. Application observability adds process and executable inspection.
3. Propagation adds network administration.
4. Go library propagation may need `CAP_SYS_ADMIN` because it uses `bpf_probe_write_user`.

`CAP_SYS_PTRACE` is used to inspect `/proc`, not to attach with `PTRACE_ATTACH`. `perf_event_paranoid`, Secure Boot, and kernel lockdown can remove capabilities even when a manifest looks correct.

### Slide 22: honest scope

This is the boundary slide. OBI can see standardized library and protocol operations. It cannot infer a tenant ID, checkout stage, or domain event unless those semantics already cross an instrumented boundary.

### Slide 23: uprobe costs

Credit Usama Saqib verbally for the two-category framing and point to his public FOSDEM talk in the footer.

Infrastructure cost: an attached uprobe crosses into the kernel. Andrii Nakryiko's RCU-protected hot-path work removed major scalability bottlenecks in newer kernels, including the 6.12 line. Do not describe old scalability results as current on every kernel.

Program cost: the BPF program itself can serialize a hot path if it touches shared state. Instrumenting a thread-safe function does not guarantee that the observer preserves its concurrency.

Compatibility cost: uprobes target compiled locations and inferred layouts. Go 1.26 removed `pcHeader.textStart`; OBI changed its resolver in PR #1851. The static-PIE crash reported in issue #2104 was also closed as fixed by that work. Present these as evidence of maintenance cost, not as current unfixed defects.

### Slide 24: no demo

Say: "I am not going to bet the keynote on conference Wi-Fi or a cluster context. zeroins has a release-pinned offline catalog and a skill that chooses the path from your constraints."

The exact commands on screen are:

```bash
go run github.com/kakkoyun/zeroins/cmd/obi-integration@latest net/http
npx skills add kakkoyun/zeroins --all
```

`kubectl-obi` and `kubectl-profiler` are experimental privileged wrappers, not production installers. Before an agent runs `attach` or `detach`, it must confirm the Kubernetes context, namespace, telemetry endpoint, transport, and privilege impact. Both wrappers require an explicit destination.

## Slides 25-35: otelc and Orchestrion

### Slide 26: convergence

Avoid the inaccurate donation shorthand. Orchestrion was not renamed to otelc.

Datadog, Alibaba, and Quesma formed the SIG after working on independent approaches. The SIG built a vendor-neutral tool. Orchestrion remains a Datadog distribution and production proof of the mechanism.

### Slide 27: timeline

Move quickly. The point is that profiling, eBPF instrumentation, and compile-time instrumentation matured in parallel. They did not merge into one runtime agent.

### Slide 28: `-toolexec`

`otelc` wraps the Go command and becomes a proxy for tool invocations. It rewrites supported syntax before `go tool compile` sees it. The resulting binary contains normal SDK calls. There is no process to attach at startup.

### Slide 29: production evidence

The 20% figure is approved for public use. Say exactly:

"Customer adoption of Orchestrion-based Go auto-instrumentation grew by about twenty percent. This is a production path, not a local debugging trick."

Do not add a customer denominator or claim that most users were new unless you have a separate cleared source.

### Slide 30: portability

The upstream project builds and tests on Linux, macOS, and Windows. Its runtime contract is Go code and an SDK, not Linux BPF APIs. Individual cgo dependencies may still have their own platform limits.

### Slide 31: four signals

Keep the distinction between engine and integration bundle:

- Upstream otelc currently instruments traces, HTTP/gRPC metrics, Go runtime metrics, and supported log records.
- Orchestrion with dd-trace-go provides traces, runtime metrics, correlated logs, and continuous profiles.

Do not say otelc alone automatically enables every profile type. Do not imply every Datadog integration has moved upstream.

### Slide 32: the GLS example

The aspect file belongs in dd-trace-go because it defines tracer-specific runtime context. The rewriting engine applies the aspect. That is why the path says `dd-trace-go`, not `otelc`.

The mechanism adds a field to `runtime.g`, injects accessors, and clears the field on goroutine exit.

### Slide 33: `go:linkname` fragility

Keep the distinction between function and variable linknames. Variable symbols have no definition/reference separation when both sides are uninitialized. The linker can choose by load order, which broke this mechanism between Go 1.22 and 1.23.

This slide preserves the failure detail from the original deck. Do not over-explain BSS layout unless asked.

### Slide 34: build-path boundaries

This is the compile-time caveat. Build ownership replaces root access as the hard constraint. Current otelc requires Go 1.25+, supported integrations define coverage, and Go toolchain internals can change.

### Slide 35: attendee path

No live build. Show the lookup and build commands, then move on. The catalog command can also be installed with the other three zeroins commands:

```bash
go install github.com/kakkoyun/zeroins/cmd/...@latest
otelc-aspect net/http
```

The direct `go run github.com/kakkoyun/zeroins/cmd/otelc-aspect@latest net/http` form on the slide avoids making installation a prerequisite.

Optional speaker-only detail about Orchestrion v2:

"Our current plan for Orchestrion v2 is to make it a downstream wrapper around otelc. otelc remains the rewriting engine, while the wrapper injects dd-trace-go instead of the OpenTelemetry Go SDK. That preserves Datadog's integration bundle without forking the core engine."

Keep this detail in spoken notes. Do not add it to the slide.

## Slides 36-43: profiles and correlation

### Slides 37-41: profiles as the fourth signal

Use the pair of questions on slide 37. A trace identifies the slow request. A CPU profile identifies where execution time went.

The OpenTelemetry Profiles specification is Alpha. The profiler itself can still be useful, but do not imply a stable wire or data-model contract.

On `.gopclntab`, explain that Go keeps the table because the runtime needs PC-to-function information for stack traces. This is why stripped static binaries remain symbolizable.

Slide 40 repeats the primary-source quote as a deliberate proof slide. Let the audience read it, then move on.

### Slide 42: OTEP 4947

This is proposed work. Say so before explaining it.

The primary proposal publishes thread context through native thread-local storage. Go is explicitly outside that primary mechanism for the foreseeable future because goroutines move across threads and FFI on each event is too expensive. The Go-specific alternative tells readers to consume pprof labels and identifies the scheme as `go_pprof_labels_v1`.

Credit Scott Gerring and Ivo Anjo verbally for connecting compile-time instrumentation to this profiling path. Link only to the public OTEP.

### Slide 43: synthesis

Slow down. The claim is:

"Compile-time instrumentation can arrange request context inside the Go process. An out-of-process profiler can read the resulting labels. One approach supplies semantics; the other supplies continuous evidence."

Do not claim the complete OTEP path is shipping today.

## Slides 44-46: decision framework

### Slide 45: first hard constraint

Read examples across the table, not every row. The first constraint usually decides the first tool:

- No rebuild window selects OBI.
- Non-Linux selects a build-time path.
- No privileged agent selects a build-time path.
- Whole-node CPU selects the profiler.

### Slide 46: practical combination

The default combination is intentionally not universal. Use build-time instrumentation where you own the Go build, OBI for rebuild-free and mixed-language boundaries, and the profiler for whole-node CPU. Add each layer only when the signal pays for its cost.

## Slides 47-52: stable hooks, Live Debugger, results, and CTA

### Slide 48: User Statically-Defined Tracing

Expand the acronym before using USDT. Stable named probe points reduce dependence on symbol addresses. Go does not ship these probes. The fork is your proof of concept and has not been proposed as an accepted upstream feature.

### Slide 49: Live Debugger

This is an existing practical application of eBPF to Go debugging. Datadog Live Debugger adds expiring logpoints and snapshots without a redeploy. Bits Live Debugger is in preview and uses those snapshots to form hypotheses. Keep product claims bounded to what the slide says.

### Slide 50: combine the layers

Do not read the columns. Land the result: Go now has credible zero-code paths at build time and runtime, and adoption grew 20% because teams want this choice.

### Slides 51-52: CTA and questions

Point at the QR and say that it still opens the talk repository: both decks, research, and the earlier FOSDEM recording. Do not imply that the QR opens zeroins.

Then point beside it to <https://github.com/kakkoyun/zeroins>. Name the four commands: `obi-integration`, `otelc-aspect`, `kubectl-obi`, and `kubectl-profiler`. The full install command is:

```bash
go install github.com/kakkoyun/zeroins/cmd/...@latest
```

The `collect-go-telemetry` Agent Skill installs with:

```bash
npx skills add kakkoyun/zeroins --all
```

Remind the room that catalog lookups are offline and read-only, while the Kubernetes wrappers are experimental, privileged, and require explicit telemetry endpoints. Mention the earlier FOSDEM recording for anyone who wants the previous version.

End with: "Pick the constraint you cannot change. Then choose the layer that can still reach your service."

Stop. Let the questions slide stay up.

## Public references for questions

| Claim | Public reference |
| --- | --- |
| zeroins toolkit and Agent Skill | <https://github.com/kakkoyun/zeroins> |
| OBI requirements and signals | <https://opentelemetry.io/docs/zero-code/obi/> |
| OBI capability ladder | <https://opentelemetry.io/docs/zero-code/obi/security/> |
| OBI trace-log correlation | <https://opentelemetry.io/docs/zero-code/obi/trace-log-correlation/> |
| OBI Go support matrix | <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/blob/main/SUPPORT_MATRIX.md> |
| uprobe performance pitfalls | <https://fosdem.org/2026/schedule/event/H3LM7G-performance_and_reliability_pitfalls_of_ebpf/> |
| RCU-protected uprobe hot path | <https://lists.openwall.net/linux-kernel/2024/08/13/142> |
| Go 1.26 OBI resolver change | <https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation/pull/1851> |
| otelc v1 and mechanism | <https://opentelemetry.io/blog/2026/go-compile-time-instrumentation-v1/> |
| otelc supported libraries | <https://opentelemetry.io/docs/zero-code/go/compile-time/supported-libraries/> |
| OpenTelemetry Injector languages | <https://github.com/open-telemetry/opentelemetry-injector> |
| OpenTelemetry Operator Go sidecar | <https://opentelemetry.io/docs/platforms/kubernetes/operator/automatic/> |
| Go pprof-label context proposal | <https://github.com/open-telemetry/opentelemetry-specification/blob/main/oteps/profiles/4947-thread-ctx.md#alternative-for-go-support> |

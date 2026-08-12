# Outline: Zero-Touch Go Instrumentation

> **Format:** 30-minute keynote, advanced Go audience
> **Deck:** `slides/presentation.md` (52 slides)
> **Current state:** additive review version; pruning follows Kemal's deck review and rehearsal
> **Scripted estimate:** 30:30
> **Demonstrations:** none; attendees get release-pinned zeroins commands and an Agent Skill instead

## Thesis

Go has no universal startup hook for instrumentation. The available approaches act at different points in the software lifecycle:

- OBI observes running processes from the Linux kernel without rebuilding them.
- otelc and Orchestrion rewrite Go code during the build and preserve Go-level semantics across platforms.
- The OpenTelemetry eBPF Profiler samples whole-node CPU activity without changing applications.

The approaches complement one another. Build-time instrumentation can publish request context that makes out-of-process profiles more useful.

## Beat sheet

| Block | Slides | Time | Purpose |
| --- | ---: | ---: | --- |
| Story and reason to care | 2-5 | 2:15 | Ask who instruments by hand, pose the zero-touch alternative, show how coverage decays, and establish why Go is different. |
| Personal and Datadog context | 6-7 | 1:15 | Establish relevant Go/observability experience, then explain why Datadog works across build-time, startup, and kernel instrumentation. |
| Why Go resists instrumentation | 8-13 | 4:00 | Native code, static linking, movable goroutine stacks, `go:linkname`, and the limited runtime-injection surface. Introduce the lifecycle map. |
| OBI | 14-24 | 6:30 | Explain deployment, signal coverage, the 13 documented Go baselines, platform and privilege constraints, and honest uprobe costs. Replace the demo with zeroins catalog tooling. |
| otelc and Orchestrion | 25-35 | 6:45 | Explain the three-organization convergence, `-toolexec`, production adoption, portability, signal depth, GLS, variable-linkname fragility, and build-path limits. |
| Profiles and request correlation | 36-43 | 4:45 | Establish profiles as the fourth signal, preserve the `.gopclntab` proof, then connect OTEP 4947's Go pprof-label path to compile-time instrumentation. |
| Decision framework | 44-46 | 2:00 | Choose by constraints, then show a practical combined deployment. |
| Stable hooks, Live Debugger, results, CTA | 47-52 | 3:00 | Keep the USDT proof of concept and Live Debugger example, restate the complementary portfolio, and send attendees to the repository and prior talk. |
| **Expanded total** | **52** | **30:30** | **Pruning is deliberately deferred.** |

## Additive-pass rule

Preserve useful technical evidence and examples so Kemal can review the complete argument. Remove only:

- Unsupported or corrected claims.
- The fabricated benchmark placeholder table and dangling harness path.
- Live-demo framing.
- Internal research links and contribution vanity metrics.

Potential timing cuts belong in the later rehearsal pass, not this revision.

## Narrative transitions

### Story to ethos

After the show of hands, ask what would happen if source changes were unnecessary. Then state the failure mode plainly: instrumentation is often careful in the first service and incomplete in the fiftieth. The problem has become a build and platform problem. Only then establish why the speaker and Datadog have relevant experience.

### Go constraints to OBI

The lifecycle map replaces a winner-versus-loser comparison. Go offers three practical intervention points: build, process start, and kernel. OBI is the kernel route and therefore inherits Linux's strengths and operational constraints.

### OBI to otelc

OBI reaches deployed workloads without rebuilding them, but it cannot infer arbitrary business semantics. otelc moves the intervention point earlier, where source structure and dependency information are still available.

### otelc to profiles

Tracing follows requests; profiling finds CPU cost. The two signals become much stronger when they share request context. Go cannot use the native thread-local mechanism directly, so pprof labels become the bridge.

### Decision to close

Return to the lifecycle map. Start with the constraint the audience cannot change, then add the next layer only when its signal pays for its operational cost. USDT shows a future stable hook contract; Live Debugger shows a practical eBPF debugging path today.

## Facts that must remain precise

- OBI supports Linux `amd64` and `arm64`, kernel 5.8+ with BTF, plus documented RHEL-family backports.
- OBI trace-log correlation enriches selected JSON logs and does not export them.
- OBI's current support matrix documents 13 Go library baselines, including `gin >= v1.6.0, != v1.7.5` and `github.com/redis/go-redis/v9 >= v9.0.0`.
- Privileges vary by feature. Do not present one capability list as universal.
- Recent kernels improved uprobe scalability, but attached probes still cost context switches and BPF programs can introduce contention.
- otelc is a stable production build-time path, not a local-development tool.
- otelc's CI covers Linux, macOS, and Windows builds. It has no eBPF or Linux-kernel runtime dependency.
- Upstream otelc currently produces traces, HTTP/gRPC metrics, runtime metrics, and supported log records. Orchestrion with dd-trace-go can also enable continuous profiling and Datadog log correlation.
- Not every Datadog integration has moved to otelc.
- Profiles are the fourth observability signal. The OpenTelemetry Profiles specification remains Alpha.
- OTEP 4947 proposes a Go-specific pprof-label path. Present it as proposed work, not a shipped guarantee.
- The OpenTelemetry host injector does not inject Go. The OpenTelemetry Operator has a separate feature-gated Go eBPF sidecar.
- Do not imply that Kemal contributes to OBI. His eBPF experience comes from Parca and parca-agent.
- The zeroins catalog commands are offline and read-only. Its Kubernetes wrappers are experimental and require explicit telemetry endpoints plus confirmation of context, namespace, transport, and privilege impact.
- Do not discuss an Orchestrion sunset.
- Keep the 20% adoption statement within its approved wording. Do not add a denominator or the "mostly new users" interpretation.

## Public audience resources

- Talk repository: <https://github.com/kakkoyun/gopherconuk-26>
- zeroins toolkit: <https://github.com/kakkoyun/zeroins>
- Prior FOSDEM version: <https://youtu.be/0TvrSebuDPk>
- OBI documentation: <https://opentelemetry.io/docs/zero-code/obi/>
- Go compile-time instrumentation: <https://opentelemetry.io/docs/zero-code/go/compile-time/>
- OTEP 4947: <https://github.com/open-telemetry/opentelemetry-specification/blob/main/oteps/profiles/4947-thread-ctx.md#alternative-for-go-support>
- Usama Saqib's public eBPF pitfalls talk: <https://fosdem.org/2026/schedule/event/H3LM7G-performance_and_reliability_pitfalls_of_ebpf/>

# Outline: Zero-Touch Go Instrumentation

> **Format:** 30-minute keynote, advanced Go audience
> **Deck:** `slides/presentation.md` (~60 slides)
> **Script:** `speaker-notes.md` (SAY/DO format)

## Thesis

The debugging loop is slow. An agent alone does not fix it. Three approaches take
the rebuild out of it: build, process start, kernel. Then the agent, holding the
framework, closes it.

## Beat sheet

| Section | Slides | Est. | Running | Purpose |
| --- | ---: | ---: | ---: | --- |
| Open | 1-10 | 2:35 | 2:35 | The debugging loop, told straight. The agent joke. Why Go is different. |
| 01 Why Go resists | 11-15 | 2:15 | 4:50 | Native code, static linking, movable stacks, go:linkname, the lifecycle map. |
| 02 OBI | 16-28 | 4:45 | 9:35 | Kernel observation. Deployment, signals, 13 baselines, platform and privilege constraints, honest uprobe costs. Usama reference. Meme A. |
| 03 otelc | 29-39 | 4:10 | 13:45 | Three-org convergence, -toolexec, production adoption, portability, signal depth, GLS, variable-linkname fragility. Meme B. Build-path limits. |
| 04 Injector | 40-44 | 1:50 | 15:35 | The third leg. What the injector is, the LD_PRELOAD problem, the teaser, what it means. |
| 05 Profiles | 45-53 | 2:40 | 18:15 | The fourth signal. .gopclntab proof, deployment cost, OTEP 4947 correlation bridge. |
| 06 When to reach for what | 54-56 | 1:15 | 19:30 | Decision table, practical combination, 20% adoption. |
| 07 What comes next | 57-63 | 3:15 | 22:45 | USDT, the agentic close, takeaways, take-home kit, SIG call to action, closing material. |

Timing is a **word-count construction**, not a measurement. The SAY content is
~1950 words. At 130 wpm with 1.5x overhead for pauses, transitions, and visual
beats, that is ~22:30. The actual read-through against a clock is still required
before delivery. On talk one the construction was wrong three times running, so
treat this as a direction, not a guarantee.

Checkpoints: open OBI by ~4:50, otelc by ~9:35, injector by ~13:45, profiles by
~15:35, the decision section by ~18:15.

## Cut ladder

Nothing is cut up front. In order, with why each is safe. Total recovery is about
4:30 of pages plus roughly a minute of trimmed spoken detail.

1. **Convergence timeline** (-0:40). The three-organisations slide carries the
   argument; the dates corroborate.
2. **Second .gopclntab proof slide** (-0:30). Repeats the quote from the slide
   before it.
3. **Full 13-library enumeration** (-0:40). "Thirteen documented baselines"
   and the support-matrix link survive on the OBI intro slide.
4. **Production-evidence 20% slide** (-0:30). The number now also sits on the
   practical-combination slide.
5. **The static-PIE anecdote** (-0:30). Script only, costs no page.
6. **How OBI works diagram** (-0:35). The punchline slide immediately after
   makes the same point harder.
7. **Usama reference slide** (-0:25). Generous, not load-bearing; the footer
   link and the verbal credit remain.
8. **Zero-touch profiling deployment-cost table** (-0:40). The OBI
   platform-contract table already taught the room to expect a privilege and
   platform bill.

## Never-cut list

Do not cut these to buy time. Trimming the jokes recreates the density problem
they exist to fix.

- The agent joke (both slides)
- Both memes
- The loop open
- The injector teaser
- The agentic close
- The uprobe-cost slide
- The OTEP correlation argument
- The decision table
- The lifecycle map
- The take-home kit

## Narrative transitions

### Open to 01

The loop is slow. The agent joke says the agent alone does not fix it. Why?
Because Go compiled away the attachment point.

### 01 to 02 (OBI)

The lifecycle map shows three intervention points. OBI is the kernel route. It
takes the rebuild out of the loop. The cost is Linux, privileges, and kernel
contracts.

### 02 to 03 (otelc)

OBI reaches deployed workloads without rebuilding, but cannot infer arbitrary
business semantics. otelc moves the intervention earlier, where source
structure is still available. The cost is owning the build.

### 03 to 04 (Injector)

Build and kernel are two legs. The third is process start. The injector loads at
startup. No source change, no rebuild. The binary is the constraint.

### 04 to 05 (Profiles)

Tracing follows requests. Profiling finds CPU cost. The two signals become
stronger when they share request context.

### 05 to 06 (Decision)

Return to the lifecycle map. Start with the constraint you cannot change. Add
the next layer only when its signal pays for its operational cost.

### 06 to 07 (Close)

The loop from the open, now short. The agent places the probe and reads
production evidence. No redeploy. Same move as the joke, except now the agent
has the framework.

## Facts that must remain precise

- OBI supports Linux amd64 and arm64, kernel 5.8+ with BTF, plus documented
  RHEL-family backports.
- OBI trace-log correlation enriches selected JSON logs and does not export
  them.
- OBI's support matrix documents 13 Go library baselines, including
  gin >= v1.6.0, != v1.7.5 and github.com/redis/go-redis/v9 >= v9.0.0.
- Privileges vary by feature. Do not present one capability list as universal.
- Recent kernels improved uprobe scalability, but attached probes still cost
  context switches and BPF programs can introduce contention.
- otelc is a stable production build-time path, not a local-development tool.
- otelc's CI covers Linux, macOS, and Windows builds. It has no eBPF or
  Linux-kernel runtime dependency.
- Upstream otelc currently produces traces, HTTP/gRPC metrics, runtime metrics,
  and supported log records. Orchestrion with dd-trace-go can also enable
  continuous profiling and Datadog log correlation.
- Not every Datadog integration has moved to otelc.
- Profiles are the fourth observability signal. The OpenTelemetry Profiles
  specification remains Alpha.
- OTEP 4947 proposes a Go-specific pprof-label path. Present it as proposed
  work, not a shipped guarantee.
- The OpenTelemetry host injector does not inject Go. The OpenTelemetry
  Operator has a separate feature-gated Go eBPF sidecar.
- Do not imply that Kemal contributes to OBI. His eBPF experience comes from
  Parca and parca-agent.
- The zeroins catalog commands are offline and read-only. Its Kubernetes
  wrappers are experimental and require explicit telemetry endpoints plus
  confirmation of context, namespace, transport, and privilege impact.
- Do not discuss an Orchestrion sunset.
- Keep the 20% adoption statement within its approved wording. Do not add a
  denominator or the "mostly new users" interpretation.
- The injector mechanism is deliberately undisclosed. Never add linking model,
  entry path, or binary-compatibility detail to any slide or script line. If
  asked, defer to next year's talk.

## Public audience resources

- Talk repository: <https://github.com/kakkoyun/gopherconuk-26>
- zeroins toolkit: <https://github.com/kakkoyun/zeroins>
- opentelemetry-agent-skills: <https://github.com/ollygarden/opentelemetry-agent-skills>
- Prior FOSDEM version: <https://youtu.be/0TvrSebuDPk>
- OBI documentation: <https://opentelemetry.io/docs/zero-code/obi/>
- Go compile-time instrumentation: <https://opentelemetry.io/docs/zero-code/go/compile-time/>
- OTEP 4947: <https://github.com/open-telemetry/opentelemetry-specification/blob/main/oteps/profiles/4947-thread-ctx.md#alternative-for-go-support>
- Usama Saqib's public eBPF pitfalls talk: <https://fosdem.org/2026/schedule/event/H3LM7G-performance_and_reliability_pitfalls_of_ebpf/>
- OpenTelemetry community SIGs: <https://github.com/open-telemetry/community>

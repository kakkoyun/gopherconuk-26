# Speaker notes: Zero-Touch Go Instrumentation

GopherCon UK 2026. 30-minute keynote. Advanced Go audience.

Deck: `slides/presentation.md`. Section headings below match slide titles
verbatim, so you can find your place by what is on screen.

## Contract

**SAY** lines are the words, written to be spoken. Short sentences, one idea
each. **DO** lines are delivery cues, kept separate so the prose stays sayable.
Memorise the beat (the bold line before each block), then let the words follow.

## Beat table

| Section | One thing |
| --- | --- |
| Open | The debugging loop is slow. An agent alone does not fix it. |
| 01 Why Go resists | Go compiled away the attachment point other runtimes have. |
| 02 OBI | Observe from the kernel. No rebuild. Linux and privileges are the cost. |
| 03 otelc | Rewrite at build time. Rich semantics. The rebuild is the cost. |
| 04 Injector | Load at process start. No source change. The binary is the constraint. |
| 05 Profiles | The fourth signal, from the kernel. Correlation is the bridge. |
| 06 When to reach for what | Start from the constraint you cannot change. |
| 07 What comes next | The loop, without the redeploy. The agent now has the framework. |

## Timing

Measured read-through pending (P4.1). The deck is ~60 pages against a 52-page
baseline. The cut ladder in `outline.md` recovers about four and a half minutes
of pages plus roughly a minute of spoken detail if the slot demands it.

Checkpoints: open OBI by ~7:30, otelc by ~14:00, injector by ~18:00,
profiles by ~20:45, the decision section by ~25:30.

---

## The bug you cannot reproduce locally

**SAY**

There is a bug. It happens in production. You cannot reproduce it locally.

So you add a log line. You redeploy.

Wrong place. You add another log line. You redeploy again.

**DO**

Let it sit. Two beats of silence after "again." Do not rescue the silence.
The discomfort is the point.

---

## What if you did not have to?

**SAY**

What if you did not have to?

**DO**

Say only this. Let the question stand. Advance.

---

## Every new dependency creates another place to forget

**SAY**

You wrapped every HTTP client and database call by hand.

The first service gets careful instrumentation. The fiftieth inherits gaps.

At fleet scale, instrumentation coverage is a platform problem.

**DO**

Keep it concrete. Do not abstract. The line to land is "a platform problem."

---

## Other runtimes have an attachment point

**SAY**

Every other runtime has a place to intervene after the source was written.

The JVM has javaagent. Python has import hooks. .NET has profiling APIs.

Go compiles that place away. It made deployment simple. It made late
instrumentation hard.

**DO**

Do not teach every mechanism. Point out that each has one. Go does not.

---

## How I got here

**SAY**

I maintained client_golang. I wrapped handlers by hand, for years.

Then Parca, where the whole point was a profiler nobody has to install.

Now OpenTelemetry compile-time instrumentation, which ships inside your binary.

Every one of those jobs taught me the same thing: hand-written
instrumentation decays. That is what this talk is about.

**DO**

Twenty seconds. Consequences, not achievements. The standing line at the
bottom is small print. Do not read it out. Do not mention contribution counts.
Do not imply work on OBI.

---

## Why Datadog cares

**SAY**

Datadog builds at all three intervention points. Compile-time instrumentation.
Startup injection. Kernel-level instrumentation and profiling.

I do not need one approach to win. I need the right combinations to work in
customer services.

**DO**

This is about the company's reason to care, not a second biography.

---

## Agent joke (slide 1)

**SAY**

Right. That is the talk. Thanks, everyone.

**DO**

Deadpan. Beat. Two seconds of silence. Do not explain the joke. Do not milk
the fake-out.

---

## Agent joke (slide 2)

**SAY**

Not really. The agent still has to pick a mechanism, and it has to know what
that mechanism costs. That is the next twenty-five minutes.

**DO**

The turn. This is the sentence that earns the rest of the talk. Ten seconds
total for both slides. Do not explain the joke.

This is a plant. It pays off twice: at the agentic close, and at the take-home
skill, which is the framework the agent was missing.

---

## 01 Why Go resists instrumentation

**SAY**

So why is Go the hard one?

**DO**

Section bridge. One sentence tying back: this is why the loop is slow, and the
three approaches that follow each take the rebuild out of it in a different way.

---

## What makes Go different

**SAY**

The compiler emits native machine code. The linker produces static binaries by
default. There is no classloader, no general runtime hook API.

And goroutines move between OS threads on movable stacks.

There is no single point where an agent can safely rewrite every function at
startup.

**DO**

Reveal each line. Do not rush. The last line is the one that matters.

---

## What Go actually has

**SAY**

Go does have linkname, a way to reach into runtime internals from your own
package.

The Go team calls the people who do this a hall of shame. Not an
instrumentation API.

**DO**

Read only the phrase "hall of shame." Do not read the whole code block.

---

## Choose where to intervene

**SAY**

There are three practical places to intervene. Build time. Process start. The
kernel.

Build time preserves Go semantics but requires a rebuild. Process-start
injection avoids source edits but depends on binary and linker surfaces.
Kernel observation avoids the rebuild but inherits Linux, privilege, and
compatibility constraints.

This is the map for the rest of the talk.

**DO**

Point to the three positions on the lifecycle map. This slide recurs. The
audience should see it as the spine.

---

## Runtime injection is a narrower route

**SAY**

Runtime injection exists for Go, but it is narrow.

Dynatrace supports eligible dynamically linked binaries. The OpenTelemetry
host injector ships Java, .NET, Node.js, and Python. It does not inject Go.
The OpenTelemetry Operator has a separate, feature-gated Go eBPF sidecar.

Datadog is working on a Go path for Single-Step Instrumentation.

**DO**

Do not promise a public release date. This slide sets up the injector section
later.

---

## 02 OBI: observe from the kernel

**SAY**

The first approach takes the rebuild out of the loop by observing from the
kernel. No source change, no rebuild, no application rollout. The cost is
Linux, privileges, and kernel contracts.

**DO**

Section bridge. Name the trade-off in one sentence.

---

## OpenTelemetry eBPF Instrumentation

**SAY**

OBI is the upstream successor to Grafana Beyla, donated to OpenTelemetry in
2025.

One DaemonSet can observe a mixed-language node. Network protocols give broad
coverage. Go-specific uprobes add library-level detail.

It is still v0. Breaking changes are possible.

**DO**

Reveal each line. The chip at the bottom is amber: do not ship this as stable.

---

## How OBI works

**SAY**

OBI sits beside your service. Uprobes and network probes go through the kernel.
Events come back to the DaemonSet, which emits OTLP traces and metrics to your
collector.

**DO**

Move quickly. The diagram is the point, not the prose.

---

## No source change / No rebuild / No application rollout

**SAY**

No source change. No rebuild. No application rollout.

**DO**

Pause after each line. Then add: the DaemonSet still has a deployment and
security cost. Zero-touch describes the application, not the platform.

---

## Go library coverage: 13 documented baselines

**SAY**

Thirteen documented baselines. HTTP, gRPC, databases, messaging.

The support matrix is the contract. Library internals can change faster than a
slide deck.

**DO**

Do not read all thirteen. Point to a few categories. Call out the corrected names
only if useful: github.com/redis/go-redis/v9, not the old v8 path. Gin excludes
v1.7.5.

---

## What OBI emits

**SAY**

OBI emits traces and metrics. HTTP, gRPC, database, messaging. RED and runtime
metrics.

Its log feature enriches selected JSON writes with trace and span IDs while a
span is active. It leaves log shipping to your existing pipeline. It is not a
log exporter.

**DO**

Say explicitly: correlation is log support, but OBI is not a log exporter.

---

## The platform contract

**SAY**

Linux only. amd64 and arm64. Kernel 5.8 or newer, with BTF. Deploy as a host
process, sidecar, or DaemonSet. macOS and Windows need a Linux VM or remote
cluster.

**DO**

Read across the table, not every row. This is the constraint, not the detail.

---

## Privileges depend on the mode

**SAY**

Privileges are a ladder, not one list.

Network capture needs BPF and raw-socket access. Application observability
adds process and executable inspection, plus perf monitoring and uprobes.
Context propagation adds network administration. Go library propagation may
need SYS_ADMIN because it uses bpf_probe_write_user.

**DO**

Speak the ladder as four rungs, never as a table. CAP_SYS_PTRACE inspects
/proc. It does not PTRACE_ATTACH. perf_event_paranoid, Secure Boot, and kernel
lockdown can remove capabilities even when a manifest looks correct.

---

## What actually works without source changes

**SAY**

OBI can see service boundaries, HTTP and gRPC spans, RED and runtime metrics,
supported library operations, and JSON log correlation.

It cannot infer a tenant ID, a checkout stage, or a domain event unless those
semantics already cross an instrumented boundary.

**DO**

This is the boundary slide. The honesty is the point.

---

## Uprobes have real costs

**SAY**

Uprobes have real costs.

Infrastructure: an attached uprobe crosses into the kernel. Heavy contention
once hurt scalability across many CPUs. Kernel 6.12 includes major
RCU-protected hot-path improvements. Do not describe old scalability results
as current on every kernel.

Program: the BPF program itself can serialize a hot path if it touches shared
state. Instrumenting a thread-safe function does not guarantee the observer
preserves its concurrency.

Compatibility: uprobes target compiled locations and inferred layouts. Go
1.26 removed pcHeader.textStart. OBI changed its resolver in PR 1851.

**DO**

Credit Usama Saqib verbally for the two-category framing. Point to his public
FOSDEM talk in the footer.

**Static-PIE anecdote (cuttable, about 30 seconds):**

Here is what that compatibility cost looks like in practice. A Kubernetes
controller changed one build flag to static-PIE. No application code changed.
On nodes running OBI, its pods exited 139. kubectl logs previous was empty.
GOTRACEBACK equals crash was silent, because the SIGSEGV beat the Go
runtime's signal handler.

Static-PIE is ET_DYN with no PT_INTERP. The external linker restructured
segments. The runtimeText offset OBI derived from .gopclntab stopped matching
prog.Vaddr, and the computed offset put the uprobe in the wrong place.

Fixed by PR 1851, which derives runtime.text from runtime.moduledata after Go
1.26 removed pcHeader.textStart.

The line it earns: that is the maintenance cost of inferring a layout you did
not compile.

**DO**

This anecdote is marked cuttable. If running long, skip it and keep only the
PR 1851 line on the slide. Present these as evidence of maintenance cost, not
as current unfixed defects. The issue was fixed. The team changed build flags,
not application code.

---

## Usama Saqib: pitfalls of production eBPF

**SAY**

Usama Saqib covers this in depth at FOSDEM 2026. The pitfalls of building
production eBPF at Datadog: kprobe performance across kernel versions, a fentry
kernel bug, and the pain of scaling uprobes.

Twenty-eight minutes, public, and it goes deeper than this section can.

**DO**

Credit him verbally. Point to the slide link. Do not read the screenshot.

---

## somehow, it works

**SAY**

Somehow, it works.

**DO**

Ten seconds. Do not explain the joke. Explaining the joke spends the time it
bought. The caption is the punchline to the uprobe section. Advance.

---

## 03 otelc: instrument during the build

**SAY**

The second approach takes the rebuild out of the debugging loop by moving the
intervention earlier, where source structure and dependency information are
still available. The cost is owning the build.

**DO**

Section bridge. Name the trade-off in one sentence.

---

## Three organizations converged upstream

**SAY**

Three organizations converged. Datadog's Orchestrion, Alibaba's rules engine,
Quesma's instrgen. They formed the OpenTelemetry Go compile-time SIG and
built otelc, a vendor-neutral tool.

Orchestrion remains a production Datadog distribution of the same core idea.

**DO**

Do not say Orchestrion was donated or renamed to otelc. It was not. Two separate
tools, same mechanism, different codebases and default tracers.

---

## A short convergence timeline

**SAY**

The projects matured in parallel. Profiling, eBPF instrumentation, and
compile-time instrumentation did not collapse into one agent. They learned to
compose.

**DO**

Move quickly. The dates corroborate; they are not the argument.

---

## `-toolexec` puts otelc before the compiler

**SAY**

otelc wraps the Go command. It becomes a proxy for tool invocations. It
rewrites supported syntax before go tool compile sees it. The resulting binary
contains normal SDK calls. There is no process to attach at startup.

It covers your code, dependencies, and supported standard-library paths.

**DO**

The diagram is the point. Do not over-explain the proxy mechanism.

---

## Production evidence

**SAY**

Customer adoption of Orchestrion-based Go auto-instrumentation grew by about
twenty percent. This is a production path, not a local debugging trick.

**DO**

Say exactly this. Do not add a customer denominator. Do not claim that most
users were new. The approved wording is the percentage alone.

---

## Build-time instrumentation is portable

**SAY**

No runtime privilege. No Linux kernel dependency. One build-command change.
Tested on Linux, macOS, and Windows. Go 1.25 or newer.

The trade-off is build ownership, not operating-system access.

**DO**

Read across the table, not every row.

---

## Signal depth depends on the integration bundle

**SAY**

Upstream otelc today instruments traces, HTTP and gRPC metrics, runtime
metrics, and supported log records.

Orchestrion with dd-trace-go provides traces, runtime metrics, correlated
logs, and continuous profiles.

Not every Datadog integration has moved to otelc yet.

**DO**

Keep the distinction between engine and integration bundle. Do not say otelc
alone automatically enables every profile type.

---

## Compile time can reach Go internals

**SAY**

Compile-time instrumentation can reach into Go internals. This aspect adds a
field to runtime.g for goroutine-local storage, typed accessors, and exit
cleanup.

dd-trace-go owns this tracer-context integration. otelc supplies the rewriting
engine.

**DO**

The aspect file belongs in dd-trace-go because it defines tracer-specific
runtime context. That is why the path says dd-trace-go, not otelc.

---

## `go:linkname` variables are fragile

**SAY**

Variable linknames are fragile. Unlike function linknames, variables have no
definition/reference separation. When both sides are uninitialized BSS symbols,
the linker chooses by symbol loading order.

Ian Lance Taylor said it in golang/go 72032: the choice is arbitrary. In the
implementation it depends on the symbol loading order.

This broke between Go 1.22 and 1.23.

**DO**

Keep the distinction between function and variable linknames. Do not
over-explain BSS layout unless asked.

---

## the linker picks by symbol loading order

**SAY**

The linker picks by symbol loading order.

**DO**

Ten seconds. Do not explain the joke. The caption is the punchline to the
linkname slide. Advance.

---

## The build path has boundaries too

**SAY**

The build path has boundaries. You need control of the Go build pipeline. You
need Go 1.25 or newer. Coverage stops at the supported integration set.
Toolchain internals like -toolexec and go:linkname can change.

No root access and no kernel version gate, but the binary must be rebuilt.

**DO**

This is the compile-time caveat. Build ownership replaces root access as the
hard constraint.

---

## 04 Process start: the injector

**SAY**

The third approach takes the rebuild out of the loop by loading at process
start. No source change, no rebuild. The cost is the binary itself.

**DO**

Section bridge. Name the trade-off in one sentence.

---

## What the injector is

**SAY**

An injector is a shared library loaded at process start. No source change, no
rebuild.

Datadog Single-Step Instrumentation and the OpenTelemetry host injector both
live here. Today they ship Java, .NET, Node.js, and Python.

**DO**

The category is public. The Go capability is the teaser.

---

## `LD_PRELOAD` reaches only some Go binaries

**SAY**

Two slides ago I told you why this is supposed to be impossible.

LD_PRELOAD works when the build uses the external linker and a compatible
dynamic path. It does not reach the default internally linked binary. Static
cgo and pure-Go PIE builds narrow the surface further.

**DO**

Be precise. Do not say LD_PRELOAD never works with Go. It works for some builds.
This slide is the problem statement for the section.

---

## The teaser

**SAY**

We made it work.

I am not going to tell you how today. That is next year's talk.

What you see is the invocation and the result. LD_PRELOAD, a service name,
debug mode on. Span creation and trace-id extraction on stderr.

**DO**

The contradiction is the device. The room will connect this slide to the
LD_PRELOAD slide. Get there first.

Do not be drawn on mechanism. Not the linking model, not the entry path, not
.gopclntab. Not on a slide, not in the script, not if asked. Defer.

The chip is amber: in development, not released.

---

## What it means

**SAY**

No rebuild. No source change. No kernel version gate. One line.

**DO**

One line. Advance.

---

## 05 Profiles complete the four signals

**SAY**

The fourth signal is profiles. Traces tell you which request was slow.
Profiles tell you where the CPU went.

**DO**

Section bridge. The two questions are the point.

---

## Traces tell you which request was slow

**SAY**

Traces tell you which request was slow.

**DO**

Pause. Let the question sit.

---

## Profiles tell you where the CPU went

**SAY**

Profiles tell you where the CPU went.

**DO**

Advance immediately after.

---

## OpenTelemetry eBPF Profiler

**SAY**

The OpenTelemetry eBPF Profiler originated as Elastic Universal Profiling and
joined OpenTelemetry in June 2024.

It samples CPU from the kernel. It profiles every process on a node. No
application rebuild, no injected library. It emits the OpenTelemetry Profiles
signal.

The Profiles specification is Alpha.

**DO**

The profiler itself can still be useful, but do not imply a stable wire or
data-model contract.

---

## Go stacks survive stripped binaries

**SAY**

Go binaries keep .gopclntab even when fully static and stripped, because the Go
runtime needs it for stack traces. The profiler uses it to unwind and
symbolize.

**DO**

Explain that the runtime needs the table. That is why stripped static binaries
remain symbolizable.

---

## Stripping does not erase `.gopclntab`

**SAY**

The information remains present even for fully static and stripped executables.

**DO**

Let the audience read the quote. Then move on. This is a deliberate proof slide.

---

## Zero-touch profiling still has a deployment cost

**SAY**

No source changes, no rebuild, no in-process agent. Stripped binaries are
supported.

But the agent needs Linux node access, BPF and perf monitoring or root, and a
DaemonSet or Collector distribution. CPU profiling is the confirmed core
signal.

**DO**

Read across the table, not every row.

---

## Correlation is the missing bridge

**SAY**

Correlation is the missing bridge.

Native runtimes can publish request context through thread-local storage. Go
cannot rely on that model. Goroutines move between OS threads, and crossing
FFI on every event is too costly.

The proposed Go path uses pprof labels. The scheme is go_pprof_labels_v1.

**DO**

Say "proposed" before explaining it. This is OTEP 4947. Credit Scott Gerring
and Ivo Anjo verbally for connecting compile-time instrumentation to this
profiling path. Link only to the public OTEP. Do not claim the complete OTEP
path ships today.

---

## Compile time sets context. eBPF reads evidence

**SAY**

Compile-time instrumentation can arrange request context inside the Go process.
An out-of-process profiler can read the resulting labels. One approach supplies
semantics. The other supplies continuous evidence.

Together, request context reaches profiles.

**DO**

Slow down. This is the synthesis. Do not claim the complete OTEP path is
shipping today.

---

## 06 When to reach for what

**SAY**

Start from the constraint you cannot change. Add a layer only when its signal
pays for its operational cost.

**DO**

Section bridge. The decision table is the point.

---

## When to reach for what

**SAY**

No rebuild window: OBI. Non-Linux target: otelc or Orchestrion.
Mixed-language boundary coverage: OBI. Rich Go library semantics: otelc or
Orchestrion. No privileged runtime agent: otelc or Orchestrion. Whole-node
CPU profiles: the eBPF Profiler. Request-correlated profiles: compile-time
context plus the profiler.

**DO**

Read examples across the table, not every row. The first constraint usually
decides the first tool.

---

## A practical production combination

**SAY**

Use build-time instrumentation where you own the Go build. OBI for
rebuild-free and mixed-language boundaries. The profiler for whole-node CPU.

Add each layer only when the signal pays for its cost.

Customer adoption grew twenty percent. Go users want the build-time path too.

**DO**

The default combination is intentionally not universal. The 20 percent line is
in its approved wording: no denominator, no "mostly new users."

---

## 07 What comes next

**SAY**

So what comes next?

**DO**

Section bridge. This is where the arc lands.

---

## User Statically-Defined Tracing (USDT)

**SAY**

USDT probes put stable, named hook points in the binary instead of forcing
tools to chase function addresses.

Go ships no built-in USDT probes today. This proof of concept shows what a
stable contract could look like.

**DO**

Expand the acronym before using USDT. The fork is a proof of concept, not an
accepted upstream feature. It is food for thought for the future. It could work
with both eBPF-based systems and injectors.

---

## The loop, without the redeploy

**SAY**

Remember the loop from the opening. The bug you cannot reproduce. The log line.
The redeploy. The wrong place.

Now: describe the bug in words. The agent places the probe and reads production
evidence. No source edit, no restart, no redeploy.

Bits Live Debugger does this today, in preview.

Same move as the opening joke. Except now the agent has the framework.

**DO**

This is the one place enthusiasm is the point. Still bounded. Name the product
once. The specifics: system-probe, the kernel floor at 5.17, logpoints that
expire automatically, the preview tier. These go here in the script because the
slide carries the idea, not the product spec.

This is the callback to the joke. The agent was missing the framework. Now it
has it.

---

## Takeaways

**SAY**

Start from the constraint you cannot change.

Add a layer only when its signal pays for its operational cost.

**DO**

Slow down. These two lines are the whole talk. Let them land.

---

## Take it home

**SAY**

Take it home.

go install the zeroins toolkit. Two offline catalog lookups: obi-integration
for OBI, otelc-aspect for otelc. Both are pinned to the versions on screen.

And the skill line: npx skills add kakkoyun/zeroins --all.

That is the joke's payoff. The framework the agent was missing is the thing you
take home.

Catalog lookups are offline and read-only. The Kubernetes wrappers are
experimental, privileged, and need an explicit telemetry endpoint.

For more agent skills, see ollygarden's opentelemetry-agent-skills.

**DO**

Point at the QR and say it opens the talk repository: both decks, research, and
the earlier FOSDEM recording. Then point to zeroins. Name the four commands.
Do not imply the QR opens zeroins.

---

## Get involved in the SIGs

**SAY**

If you want to help build this, get involved in the SIGs.
github.com/open-telemetry/community has every SIG's calendar, notes, and
channels.

Go Compile-Time Instrumentation. eBPF Instrumentation. The Injector. Profiling.

**DO**

Point to the community repo. All channels are listed there.

---

## Take it with you

**SAY**

The talk repository has both decks, the research, and the earlier FOSDEM
recording. The zeroins toolkit has the four commands. The agent skills repo
has more.

Pick the constraint you cannot change. Then choose the layer that can still
reach your service.

**DO**

End with this line. Stop. Let the questions slide stay up.

---

## Reserve answers

Answers to likely questions. Keep them bounded.

**OBI roadmap.** The v0.11.0 target and the proposed path to v1.0 are SIG
proposals, not commitments. v0.11.0 is targeted mid-August, gated on Config v2.
A proposed path: v0.12 in September, rc1 in October, v1.0.0 target late
October, before KubeCon NA. All proposed, not committed. Never state the
October date as a plan of record.

**Go-Auto sunset.** The slide's claim is about the OpenTelemetry Operator's
feature gate, not about a Go-Auto project. The claim holds either way. Do not
discuss an Orchestrion sunset.

**Telemetry cost.** Real question. Answer with open-source backends and short
retention. You can always use an open-source observability backend and keep
signals for a shorter time in your local environment.

**OBI custom spans.** RFC work is active. The boundary slide is accurate today:
zero code changes for standard observability, but use language agents or manual
instrumentation when you need custom spans, application-specific attributes, or
business events.

**Injector mechanism.** Deferred to next year. The mechanism is deliberately
undisclosed. Do not add linking model, entry path, or binary-compatibility
detail. If asked, say: that is next year's talk.

## Guard rails that survive the rewrite

Never say Orchestrion was donated or renamed to otelc. C-021 is REFUTED and
both projects' people are reachable.

Never present one capability list as universal. Privileges vary by feature.

Never claim the OTEP path ships today. It is proposed work.

The standing line on the ethos slide is small print and is not read aloud.

## Public references for questions

| Claim | Public reference |
| --- | --- |
| zeroins toolkit and Agent Skill | <https://github.com/kakkoyun/zeroins> |
| opentelemetry-agent-skills | <https://github.com/ollygarden/opentelemetry-agent-skills> |
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
| OpenTelemetry community SIGs | <https://github.com/open-telemetry/community> |

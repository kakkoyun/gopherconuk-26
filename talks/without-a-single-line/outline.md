# Outline: Without a Single Line

> **Status:** ⬜ Draft — to be refined after research is complete
> **Format:** 30-min keynote, advanced audience

## Beat sheet

| # | Section | Time | Key point |
|---|---------|------|-----------|
| 0 | Why listen to me | < 1 min | Ethos slide — 20 seconds, do not read aloud; land the steering-committee line. |
| 1 | Hook | 2 min | "Go can't be monkey-patched — or can it?" Live question to audience. |
| 2 | The problem | 3 min | No runtime hook point. Static binary. The three workaround families. |
| 3 | OBI | 7 min | eBPF from the outside. Production-safe. Live demo: attach to a running service. |
| 4 | otelc | 7 min | Compile-time from the inside. Granular. Live demo: `go build -toolexec otelc`. |
| 5 | ebpf-profiler | 5 min | The third signal: profiling. No code changes. Whole-system. |
| 6 | Benchmark shootout | 3 min | Overhead table. When each approach costs you something. |
| 7 | Decision framework + agent demo | 2 min | OBI=prod, otelc=local-dev, profiler=always-on. Skill demo. |
| 8 | Runtime futures | 1 min | Flight recording, USDT, proposal #69887 — the horizon. |

## Total: 30 min

## Live demos (to be built in `demo/`)

1. **OBI demo**: realistic HTTP+DB service; attach OBI via docker-compose; show traces in backend
2. **otelc demo**: same service; `go build -toolexec otelc`; show compile output + traces
3. **Agent skill demo**: Claude Code skill choosing OBI vs otelc based on context

## Slides (to be built after research)

See `slides/` — Marp deck using otel theme.

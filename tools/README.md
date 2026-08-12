# Tools

## Benchmark tools

The benchmark commands and their Agent Skills now live in [`github.com/kakkoyun/benchlab`](https://github.com/kakkoyun/benchlab).

```bash
# Install honestbench, benchgate, and benchenv
go install github.com/kakkoyun/benchlab/cmd/...@latest

# Install all three skills for supported coding agents
npx skills add kakkoyun/benchlab --all
```

## Observability tools

The remaining tools support the *How to Instrument Go Without Changing a Single Line of Code* talk.

| Path | Purpose |
|---|---|
| `cli/kubectl-obi/` | Kubernetes helper for deploying and inspecting OpenTelemetry eBPF Instrumentation (OBI) |
| `cli/kubectl-profiler/` | Kubernetes helper for profiler workflows |
| `cli/go-instr-pull/obi-integration.sh` | Queries OBI integration support for a Go package |
| `cli/go-instr-pull/otelc-aspect.sh` | Queries `otelc` aspect support for a Go package |
| `skills/collect-go-telemetry/SKILL.md` | Agent workflow for choosing and applying Go telemetry backends |

## Repo checks

`check_slide_footer.py` verifies both generated PDFs during `make check`.

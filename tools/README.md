# Tools: agent-actionable benchmark discipline

Three CLIs and the Claude Code skills that wrap them. Architecture: **skills wrap CLIs** — the Go binary is the single source of truth; the skill is the agent-facing entry point. A human runs the CLI directly; an agent invokes the skill.

Each CLI is its own Go module (`go 1.24`), stdlib-only, with table-driven tests.

| CLI (`tools/cli/`) | Skill (`tools/skills/`) | Answers |
|--------------------|-------------------------|---------|
| `honestbench` | `honest-benchmark` | Is the compiler measuring real work? (static analysis) |
| `benchgate` | `benchstat-gate` | Is my sample stable enough? (CV gate + benchstat) |
| `benchenv` | `diagnose-noisy-bench` | Why is my benchmark noisy? (environment diagnosis) |

These map directly to the talk's three trust questions.

## honestbench — benchmark correctness analyzer

Parses `*_test.go` with `go/ast` and flags: discarded results (dead-code elimination), missing sink patterns, `StopTimer`/`StartTimer` misordering, and `b.N` loops that should migrate to `testing.B.Loop` (Go 1.24). Exit 1 on findings — usable as a CI gate.

```bash
cd tools/cli/honestbench && go build -o honestbench .
./honestbench ../../../talks/go-benchmarks-lying/demo/
```

## benchgate — benchmark stability gate

Runs benchmarks N times, computes coefficient of variation (CV) per benchmark, and fails if any exceeds a threshold. Optionally diffs against a saved baseline via `benchstat`.

```bash
cd tools/cli/benchgate && go build -o benchgate .
./benchgate -pkg ./... -count 10 -cv-threshold 5.0
```

## benchenv — environment diagnosis

Reports active noise sources (SMT, frequency governor, Turbo Boost, load average, missing `perflock`/`benchstat`/`benchdiff`) with per-item remedies. Cross-platform; degrades gracefully on macOS where sysfs controls are unavailable.

```bash
cd tools/cli/benchenv && go build -o benchenv .
./benchenv
```

## Verification

Each module passes `go build ./...`, `go vet ./...`, `go test ./...`, and `gofmt -l .` clean. `benchenv` additionally cross-compiles to `linux/amd64`.

Research backing these tools lives in `../research/go-benchmarks-lying/`.

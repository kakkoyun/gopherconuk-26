---
name: honest-benchmark
description: |
  Static analyzer for Go benchmark correctness. Flags dead-code elimination
  traps, timer misuse, and missing sink patterns in *_test.go files via go/ast.
  USE WHEN: "audit go benchmark", "check benchmark correctness", "benchmark dead
  code elimination", "is my benchmark measuring real work", "benchmark timer order".
disable-model-invocation: false
---

# honest-benchmark

Runs `honestbench` — a static analyzer that reads Go benchmark functions from
`*_test.go` files and flags the four most common correctness mistakes.

## Build the CLI

From the repo root:

```bash
cd tools/cli/honestbench
go build -o honestbench .
```

Or install it to your PATH:

```bash
go install github.com/kakkoyun/gopherconuk-26/tools/honestbench@latest
```

## Run

```bash
# Single file
honestbench path/to/foo_test.go

# Directory (non-recursive)
honestbench ./mypkg/

# Directory tree (recursive)
honestbench -r ./...

# Machine-readable output for CI or agents
honestbench -json ./mypkg/ | jq .

# Findings only, no summary (useful in scripts)
honestbench -q ./mypkg/
```

**Exit codes:** `0` = no findings, `1` = findings present (CI gate), `2` = error.

## Output format

```
path/file_test.go:42:3: high: discarded-result: call to compute() result is discarded; compiler may eliminate this call via DCE
```

Fields: `file:line:col: severity: rule: message`

JSON (`-json`) adds `func`, `suggestion` fields, and is an array suitable for
programmatic consumption.

## Rules and how to fix them

### `discarded-result` (high)

A function call inside the benchmark loop whose return value is discarded.
The Go compiler is free to remove the entire call via dead-code elimination —
the loop runs `b.N` times but does no real work.

```go
// BAD — DCE removes makeBuffer; allocs/op = 0
for range b.N {
    makeBuffer(64)
}

// GOOD — result escapes through package-level sink
var sink []byte
for b.Loop() {
    sink = makeBuffer(64)
}
```

**Rule:** assign the result to a local accumulator inside the loop, then write
the accumulator to a package-level variable after the loop. One global write
per benchmark run keeps allocation overhead negligible while guaranteeing
the compiler retains the computation.

### `missing-sink` (medium)

The benchmark accumulates results into a local variable but then discards it
with `_ = v`. The compiler sees through `_ = v` and may still eliminate the
computation.

```go
// BAD — _ = s does NOT prevent DCE
var s [32]byte
for range b.N {
    s = sha256.Sum256(payload)
}
_ = s

// GOOD — package-level sink defeats DCE
var globalSink [32]byte
var s [32]byte
for b.Loop() {
    s = sha256.Sum256(payload)
}
globalSink = s
```

### `stoptimer-without-starttimer` (high)

Two sub-patterns are flagged:

1. `b.StopTimer()` inside the loop with **no** `b.StartTimer()` — the timer
   never restarts; `b.N` keeps doubling and the benchmark never exits.

2. `b.StartTimer()` is the **last** statement in the loop body — work runs
   while the timer is stopped, so `ns/op` measures fixture cost, not work cost.

```go
// BAD — StartTimer after work: measures fixture, not processString
for range b.N {
    b.StopTimer()
    input := buildFixture(n)
    result = processString(input) // timer is OFF here
    b.StartTimer()                // too late
}

// GOOD — StartTimer before work
for b.Loop() {
    b.StopTimer()
    input := buildFixture(n)
    b.StartTimer()                // timer ON before measured work
    result = processString(input)
}
globalSink = result
```

### `suggest-bloop` (info)

The benchmark uses `for range b.N` or `for i := 0; i < b.N; i++`. Go 1.24
introduced `b.Loop()`, which solves three footguns in one construct:

- Automatic timer reset (no need for `b.ResetTimer()` after setup).
- Suppresses inlining of the loop body, reducing DCE risk.
- Handles the `b.N == 0` edge case (benchmarks that pre-allocate a
  result slice of length `b.N` would panic).

```go
// Before
for range b.N {
    s = sha256.Sum256(data)
}

// After (Go 1.24+)
for b.Loop() {
    s = sha256.Sum256(data)
}
```

## Interpreting the summary line

```
17 findings (2 high, 4 medium, 11 info) across 12 functions
```

- **High** findings are correctness bugs — the benchmark is likely measuring
  nothing useful. Fix before trusting any numbers.
- **Medium** findings are likely to produce incorrect results under optimising
  compilers. Fix before publishing numbers.
- **Info** findings are migration suggestions; the benchmark is correct today
  but could be improved.

## Using in CI

```yaml
- name: Audit benchmarks
  run: |
    cd tools/cli/honestbench && go build -o /tmp/honestbench .
    /tmp/honestbench -r ./...   # exits 1 if any findings
```

Gate on exit code `1` to block PRs that introduce discarded-result or
stoptimer bugs. Consider allowing info findings through with `-q` and a
severity filter in scripts.

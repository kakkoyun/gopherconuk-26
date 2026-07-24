# War Stories: First-Hand Benchmark Failures and Fixes

These are real incidents from production Go codebases, not constructed hypotheticals. Both happened
during active development sessions and are reconstructed from session transcripts. The numbers are
exact; the reasoning is verbatim.

---

## Story 1: The CI Regression That Was a Speedup

**Codebase:** DataDog/dd-trace-go  
**Date:** June 12, 2026  
**PR:** #4891 — restructure of `context.go` in the `ddtrace/tracer` package (compile-time
instrumentation plumbing; larger footprint than a sibling PR #4808, which was a 2-line change)

### What CI said

After pushing #4891, the `pr-commenter` benchmark bot flagged `BenchmarkOTLPProtoSize` as
**6–9% slower than main**. A bot comment on the PR pointed to the regression.

First instinct: something in the change is hurting the OTLP encoding path. But before touching
anything, the right move is to read what the benchmark actually measures.

### Investigation

Step 1: read the benchmark. `BenchmarkOTLPProtoSize`'s timed loop is:

```go
proto.Size(tracesData)
```

That's a protobuf-library size computation on a struct that is fully assembled *before*
`b.ResetTimer()`. It never calls `ContextWithSpan`, `SpanFromContext`, or anything else the PR
touches. The change provably cannot affect `proto.Size` execution time.

Step 2: check local variance. Running the benchmark repeatedly on the development machine gave
**<0.1% run-to-run variance** — the machine itself is stable, so the CI signal of 6–9% is not
generic runner noise on this box.

Step 3: the decisive test. Build `main` and `#4891` on the *same machine* and compare directly
with `benchstat`:

| Build   | 1 span    | 10 spans  |
|---------|-----------|-----------|
| main    | 883.3 ns  | 7115 ns   |
| #4891   | 840.7 ns  | 6775 ns   |

**#4891 was faster than main — the opposite direction from what CI reported.**

### Root cause

Restructuring `context.go` shifted function addresses across the `ddtrace/tracer` package. That
moved the hot `proto.Size` loop's instruction fetch window relative to cache-line and branch-target
buffer boundaries. The benchmark is measuring ~390 ns per iteration; small alignment shifts at that
scale produce several-percent swings in either direction.

The tell: a performance delta that **flips sign** between the CI runner and a developer machine is
the signature of code-layout noise, not a real cost change.

Comparison with #4808: that PR is a 2-line change, so it doesn't perturb the package's codegen.
#4891's larger `context.go` restructure does — which is enough to make alignment-sensitive
microbenchmarks flap in CI.

### Resolution

No code fix. Nothing was wrong. Pushing a speculative "fix" would have been chasing noise. The PR
remained as drafted.

### Postscript

The same `BenchmarkOTLPProtoSize` false positive appeared again on a subsequent PR eleven days
later (June 23). At that point it was immediately recognized: "This benchmark bot comment is the
known `BenchmarkOTLPProtoSize` false positive — documented for #4891 as a same-package
code-layout/alignment artifact (the benchmark touches no code I changed; local A/B was ~+0.3%).
No action."

The false positive had become known-noise within two occurrences.

### Lesson

**Code-layout changes in the same package are enough to produce 6–9% benchmark noise in either
direction.** If a CI regression names a benchmark whose timed loop doesn't call any code your PR
changed, read the benchmark before fixing the code. An A/B on the same machine — not a re-run on
the CI runner — is the decisive test.

---

## Story 2: The Machine That Was Running Too Hot

**Codebase:** open-telemetry/opentelemetry-go-compile-instrumentation  
**Date:** July 3, 2026  
**PR:** #643 — dependency pinning fix (grpc/redis pinned independently in
`test/bench/scenarios/multi`, same root cause as #644)

### What local measurement said

After landing the fix, the overhead benchmark was run locally to confirm nothing regressed. The
output showed:

- `multi` scenario: **230% overhead** (ceiling: 150%)
- `largeidle` scenario: **212% overhead** (ceiling: 150%)

Both scenarios were blowing through the threshold. Instinct: the fix introduced a real regression.

### Why that was wrong

`largeidle` shares **zero** bumped dependencies with the fix. The PR touched dependency pinning in
`multi`. If the PR caused a real performance regression, `largeidle` — which runs independent code
— could not be affected.

Yet `largeidle` showed the same ~60% overhead inflation as `multi`.

The common factor is not the code. It's the machine.

At the time the benchmark was run, the development machine had been running heavy parallel builds
and integration tests throughout the session. Background CPU and memory pressure was the actual
variable; the benchmark was measuring that, not the PR's change.

### Relevant structure

The `benchmark/threshold` job is **CI-only** — there is no local make target that runs it. When
overhead benchmarks are run locally, they run on whatever state the developer machine happens to be
in. The CI job runs on a dedicated, idle runner.

The decision: "CI runs on a dedicated runner; letting it re-run is the reliable signal here."

### Resolution

Waited for CI. The `Overhead Threshold Check` job fired later with a fresh run on a clean runner,
producing numbers that reflected only the code, not the developer machine's load.

### Lesson

**A benchmark that reports the same regression in scenarios your change doesn't touch is measuring
machine state, not code behavior.** When `largeidle` and `multi` both show the same anomaly and
only `multi` was touched, the machine is the variable. On a loaded developer machine, overhead
benchmarks are not trustworthy. The authoritative number comes from a dedicated, idle runner — and
sometimes the right call is simply to wait for CI.

---

## Story 3: The Recurring False Positive (Pattern Recognition)

This is not a separate incident but a pattern that emerged across Stories 1 and 2 and deserves its
own framing for a talk.

`BenchmarkOTLPProtoSize` is a nanosecond-range microbenchmark that measures a function
(`proto.Size`) in the same package as dd-trace-go's tracer core. Any PR that changes *any* file in
`ddtrace/tracer` — even one that doesn't touch the OTLP encoding path at all — can shift function
addresses enough to perturb this benchmark's cache alignment and produce a multi-percent signal in
either direction.

This benchmark became a known false positive within two occurrences. The pattern it illustrates:

1. CI reports a regression.
2. Investigation shows the benchmark's hot loop doesn't call any code the PR changed.
3. Local A/B shows the delta is in the opposite direction or is within noise.
4. The benchmark is filed as a known layout artifact.
5. The next PR that touches the same package encounters the same bot comment and can immediately
   dismiss it.

**The infrastructure cost:** without the investigation in Story 1, a developer encountering this
bot comment on a later PR would have no context. They might spend hours looking for a real
regression that doesn't exist, or worse, push a speculative "optimization" that makes the benchmark
quieter by accident while introducing actual regressions.

Documented false positives are a form of benchmark hygiene. They save time on every subsequent PR
that triggers the same noise.

### Lesson

**Keep a ledger of known false positives.** When a CI benchmark regularly fires on PRs that
demonstrably don't touch its code path, document it — what benchmark, why it's a layout artifact,
what the A/B on same-machine showed. That documentation pays for itself on the very next PR.

---

## How These Stories Map to the Talk

| Story | Section | Claim it illustrates |
|---|---|---|
| 1 — The CI Regression That Was a Speedup | Layer 1: What Go benchmarks measure (compiler section) / Layer 2: Environment | Nanosecond microbenchmarks are sensitive to code layout; package-level restructuring shifts alignment enough to produce spurious multi-percent signals |
| 2 — The Machine That Was Running Too Hot | Layer 2: Environment | A benchmark run on a loaded developer machine is not a benchmark of the code; the CI dedicated runner is the authoritative signal |
| 3 — The Recurring False Positive | Tooling / CI integration section | Known false positives must be documented; undocumented noise taxes every future PR |

Story 1 is the strongest stage anecdote because it has a dramatic inversion (CI says slower, same-machine A/B says faster), concrete numbers (883 ns vs 841 ns), a clear mechanical explanation (alignment shift), and a resolution that required no code change. Lead with it.

Story 2 is the more subtle lesson: the tell is not the absolute number but the *scope* — when
a scenario that can't be affected shows the same anomaly, the machine is the problem, not the code.
It fits naturally in the "what makes a benchmark environment trustworthy" section.

Story 3 is the meta-point for the "practical recommendations" close: building institutional memory
around benchmark noise is as important as fixing the code.

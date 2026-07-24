# Compiler Honesty in Go Benchmarks
## Is the compiler measuring real work?

A Go benchmark makes an implicit contract with the developer: "the number reported is the time to execute the code between the curly braces." The contract breaks the moment the compiler notices it can produce a semantically equivalent program that does less work — and it is aggressively correct to do so. Dead-code elimination, constant folding, and inlining are not bugs; they are the optimizer doing its job. But that job makes microbenchmarks fundamentally adversarial: the same transformations that make production code fast will silently gut a benchmark loop to a no-op, leaving a result that looks plausible but measures nothing.

This section catalogues every compiler transformation that invalidates Go benchmarks, shows how to detect each one, and gives the canonical fix. It closes with `testing.B.Loop` (Go 1.24), which resolves several of these issues at the language level, and with a cross-language survey confirming the problem is not Go-specific.

---

## 1. Dead-Code Elimination

### What the compiler does

The Go compiler performs dead-code elimination (DCE): if a computation produces a value that nothing ever reads, the compiler is free to remove the computation entirely. In a benchmark, the function under test is typically called inside a loop and its return value is not used for anything — the benchmark just wants to measure the time. The compiler sees an unused result and eliminates the call.

The effect is a benchmark that reports sub-nanosecond timings: the loop still executes `b.N` times, but each iteration is an empty shell.

### Minimal example

```go
package bench_test

import "testing"

func sum(a, b int) int {
    return a + b
}

// WRONG: result is unused → compiler eliminates the call to sum
func BenchmarkSumDCE(b *testing.B) {
    for range b.N {
        sum(1, 2) // optimized away entirely
    }
}
```

Running this benchmark will report something like `0.30 ns/op` — not because `sum` is that fast, but because it is not running at all.

A second, subtler case arises when a function is inlined (see section 3). After inlining, the compiler can see that the computation inside the function body is also unused, and eliminates the entire inlined body. This can produce plausible — but wrong — timings, because a partially eliminated body may still do some work.

### The sink-variable fix

The canonical fix uses a package-level variable, conventionally named `sink`, to create an observable side effect that forces the compiler to retain the computation:

```go
package bench_test

import "testing"

var sink int // package-level sink defeats DCE

func sum(a, b int) int {
    return a + b
}

func BenchmarkSumCorrect(b *testing.B) {
    var s int
    for range b.N {
        s = sum(1, 2) // result assigned → not dead
    }
    sink = s // write to package-level var after loop
}
```

Two variables are used deliberately:

- The local `s` accumulates inside the loop without touching a global on every iteration (one global write per benchmark run, not per iteration), keeping the benchmark overhead low.
- The assignment `sink = s` after the loop is what defeats DCE. Because `sink` is a package-level variable, the compiler cannot prove it is unread — some other package, some test runner, or the linker might observe it — so it must retain the computation that produces the value.

This pattern is documented in Dave Cheney's foundational article on Go benchmarking [1] and is the de facto standard in the Go standard library's own benchmarks.

**Rule of thumb:** If a benchmark produces a result (almost any non-void function does), assign it to a local inside the loop and write the local to a package-level `var` after the loop.

---

## 2. Constant Folding

### What the compiler does

Constant folding is a compile-time optimization: if all inputs to an expression are compile-time constants, the compiler evaluates the expression at compile time and replaces it with the result literal. The benchmark loop then iterates over a constant — zero work, zero execution time attributable to the function under test.

### Example

```go
package bench_test

import (
    "testing"
    "math/bits"
)

// WRONG: bits.OnesCount(0b10110) has a constant argument.
// The compiler evaluates it to 3 at compile time.
func BenchmarkOnesCountConstant(b *testing.B) {
    var sink int
    for range b.N {
        sink = bits.OnesCount(0b10110) // constant-folded to: sink = 3
    }
    _ = sink
}
```

The reported time measures the cost of assigning a literal integer to a variable, not the cost of `bits.OnesCount`.

### Fix

Use a variable input that the compiler cannot prove is constant at compile time. A package-level variable works because the compiler cannot inline across package boundaries for the purpose of constant propagation:

```go
var input uint = 0b10110

func BenchmarkOnesCountVariable(b *testing.B) {
    var sink int
    for range b.N {
        sink = bits.OnesCount(input)
    }
    _ = sink
}
```

For table-driven benchmarks, the subtlety is that a loop variable captured by value inside a sub-benchmark can still be constant-folded if the compiler can prove its value statically. Prefer passing inputs through a `[]struct{ name string; input T }` slice loaded from a package variable.

---

## 3. Inlining Changing What Is Timed

### What the compiler does

The Go compiler inlines small functions: it replaces a call site with a copy of the callee's body. Inlining is almost always beneficial in production — it eliminates call overhead and enables further optimizations on the combined code. In a benchmark, inlining interacts badly with DCE: once the callee's body is inlined into the benchmark loop, the compiler can see that the result is unused and eliminate the entire body.

The dangerous case is a benchmark that *appears* to measure function X but actually measures the inlined body of Y, which X calls. If Y is trivial enough after inlining, the entire chain can be eliminated.

### Example

```go
package bench_test

import "testing"

// isCond has a complex-looking condition, but with a constant argument
// the compiler inlines it and then constant-folds or eliminates it.
func isCond(b byte) bool {
    if b%3 == 1 && b%7 == 2 && b%17 == 11 && b%31 == 9 {
        return true
    }
    return false
}

// WRONG: isCond is inlined; with constant argument 201, the result is
// known at compile time and the call is eliminated entirely.
func BenchmarkIsCondWrong(b *testing.B) {
    for range b.N {
        isCond(201) // inlined, constant-folded, DCE'd → ~0 ns/op
    }
}
```

This benchmark was used as the motivating example in the Go blog post announcing `testing.B.Loop` [2] and in the proposal discussion [3]. It produces results well under 1 ns/op — not because the condition is cheap, but because the compiler proves the result is `false` at compile time and removes the code.

### Fix

Apply both remedies: use a non-constant input *and* capture the result through a sink:

```go
var condInput byte = 201
var sink bool

func BenchmarkIsCondCorrect(b *testing.B) {
    var s bool
    for range b.N {
        s = isCond(condInput)
    }
    sink = s
}
```

**Detecting the problem:** A result under 1 ns/op for anything non-trivial is the strongest signal that DCE or constant folding has struck. Any result that does not scale linearly with the computational complexity of the function under test is suspect.

**Checking inlineability:** Run `go build -gcflags='-m'` to see which functions the compiler decides to inline. A function annotated `can inline X` will be inlined at every call site.

---

## 4. AllocsPerOp and -benchmem

### What it measures

The `-benchmem` flag (or the per-benchmark `b.ReportAllocs()` call) adds two columns to benchmark output:

```
BenchmarkFoo-8   1000000   523 ns/op   128 B/op   2 allocs/op
```

- **`B/op`** — bytes allocated per operation: `r.MemBytes / r.N` [4].
- **`allocs/op`** — heap allocations per operation: `r.MemAllocs / r.N` [4].

These are measured by capturing `runtime.ReadMemStats` at the start and end of the benchmark run. The deltas are divided by the final `b.N` value.

### Why it matters

Allocation count is often more actionable than timing. Two implementations can have similar `ns/op` on a warm benchmark but very different GC pressure in production. The `allocs/op` column surfaces:

- Unexpected escapes to heap (interface boxing, closure captures, `fmt.Sprintf` allocations).
- Missed opportunities for sync.Pool or pre-allocated buffers.
- Regressions introduced by a refactor that changes an escape decision.

A value of `0 allocs/op` is the target for any hot-path function that should not allocate.

### Limitations

`-benchmem` counts *heap* allocations only. Stack allocations are free and invisible to this counter. The Go compiler's escape analysis decides whether a value lives on the stack or escapes to the heap; `go build -gcflags='-m'` reveals escape decisions. A function that previously stack-allocated a value and now heap-allocates it after a code change will show a regression in `allocs/op` — which is the signal you want.

The counter also does not distinguish between allocations in the function under test and allocations in the benchmark harness itself. Keep benchmark setup outside the timed loop (using `b.ResetTimer` or the automatic reset in `b.Loop`) to avoid counting setup allocations in the per-op figures.

---

## 5. ResetTimer and StopTimer

### The contract

The `testing.B` timer measures elapsed wall time and allocation counts. It starts automatically when the benchmark function is called. Anything that runs before the first iteration of the measurement loop is included in the measurement unless the benchmark explicitly calls `b.ResetTimer()`.

### Correct ResetTimer usage

Use `b.ResetTimer()` when setup before the benchmark loop is non-trivial (network connections, large allocations, file I/O, complex data structure construction):

```go
func BenchmarkSortLarge(b *testing.B) {
    data := generateLargeSlice(100_000) // expensive setup
    b.ResetTimer()                       // zero the timer AFTER setup
    for range b.N {
        sorted := make([]int, len(data))
        copy(sorted, data)
        slices.Sort(sorted)
    }
}
```

`ResetTimer` zeroes both the elapsed time and the allocation counters. It does *not* stop and restart the timer; if the timer was running, it continues running after the reset [4].

### Classic misuse #1 — forgetting ResetTimer

```go
// WRONG: setup time is included in the reported ns/op
func BenchmarkWithExpensiveSetupWrong(b *testing.B) {
    conn := openDatabaseConnection() // might take 50ms
    defer conn.Close()
    // no b.ResetTimer() — the 50ms shows up in ns/op
    for range b.N {
        conn.Query("SELECT 1")
    }
}
```

The reported `ns/op` includes the connection setup time divided across `b.N` iterations. For large `b.N` the distortion is small, but for small `b.N` (slow operations) it is significant.

### Classic misuse #2 — ResetTimer called repeatedly or inside the loop

```go
// WRONG: calling ResetTimer inside the loop discards all prior measurements
func BenchmarkResetInsideLoopWrong(b *testing.B) {
    for range b.N {
        doSetup()
        b.ResetTimer() // resets on every iteration — measures nothing useful
        doWork()
    }
}
```

`ResetTimer` inside the loop zeroes accumulated time and allocation counts on every iteration, leaving only the last iteration's data. Use `b.StopTimer` / `b.StartTimer` instead when per-iteration setup is needed.

### StopTimer / StartTimer for per-iteration setup

Use `b.StopTimer()` and `b.StartTimer()` when each iteration requires its own setup that should not be timed:

```go
func BenchmarkSortInts(b *testing.B) {
    ints := make([]int, 1000)
    for b.Loop() {          // b.Loop handles the outer timer automatically
        b.StopTimer()
        fillRandom(ints)    // per-iteration setup, not timed
        b.StartTimer()
        slices.Sort(ints)   // only this is timed
    }
}
```

**Warning:** `StopTimer` and `StartTimer` inside a `RunParallel` body have global effect and will corrupt measurements across goroutines [4]. Never use them inside `b.RunParallel`.

### Classic misuse #3 — calling StopTimer without StartTimer

```go
// WRONG: timer is stopped; everything from here to end of benchmark is untimed,
// but the benchmark still reports a (wrong) ns/op.
func BenchmarkStopWithoutRestart(b *testing.B) {
    for range b.N {
        doWork()
        b.StopTimer() // pauses timer after first iteration
        // forgot b.StartTimer() — subsequent iterations are not timed at all
    }
}
```

The reported time will be the time of exactly one iteration, divided by `b.N`. The result is correct for one iteration but reported as if it is the average of all iterations.

---

## 6. testing.B.Loop (Go 1.24)

### Background and proposal

`testing.B.Loop` was proposed by Austin Clements (GitHub: `@aclements`) in [Go issue #61515](https://github.com/golang/go/issues/61515), opened July 21, 2023 [3]. The proposal was accepted and shipped in Go 1.24 [5].

The motivating observation was that the `b.N` pattern has a cluster of recurring failure modes that are hard to catch statically, produce silent wrong results, and trip up experienced Go developers:

1. **Forgotten loops** — benchmarks silently pass with a `b.N`-loop omitted.
2. **`b.N` misused as an input size** — `Fib(b.N)` instead of `for range b.N { Fib(10) }` [1].
3. **`ResetTimer` confusion** — unclear when setup is "expensive enough" to warrant it; vet cannot reliably detect the omission.
4. **Expensive ramp-up** — the framework calls the benchmark function multiple times with increasing `b.N` until the run is long enough to time reliably, meaning any setup code in the function body (outside the loop) runs several times.
5. **Dead-code elimination** — the compiler eliminates the loop body when results are unused; the existing `var sink T` workaround is easy to forget and not enforced.

### The API

```go
// Go 1.24+
func Benchmark(b *testing.B) {
    // setup here — runs exactly once
    for b.Loop() {
        // code to measure
    }
    // cleanup here — runs exactly once
}
```

The `for b.Loop()` form mirrors `testing.PB.Next()` from parallel benchmarks and was chosen over a callback form (`b.Loop(func() {...})`) to preserve normal control flow (returns, defer, variable scoping) inside the loop body.

### Mechanical differences from b.N

| Behaviour | `for range b.N` | `for b.Loop()` |
|---|---|---|
| Automatic `ResetTimer` at loop start | No | Yes [4] |
| Automatic `StopTimer` at loop end | No | Yes [4] |
| Benchmark function called per measurement | Multiple times | Exactly once per `-count` [5] |
| Setup before loop re-executes on ramp-up | Yes | No [3] |
| DCE of loop body | Possible | Prevented by compiler [2] |
| Compatible with `b.StopTimer`/`b.StartTimer` inside loop | Yes | Yes [4] |

### How the compiler prevents DCE

In Go 1.24, the compiler detects loops whose condition is syntactically `b.Loop()` and **disables inlining into the loop body** [2]. This prevents the chain of inlining-then-DCE described in section 3. The implementation uses `runtime.KeepAlive` semantics applied to variables used within the loop body [4].

The limitation: this optimization applies only when the condition is written *exactly* as `b.Loop()`. Assigning `b.Loop` to a variable and using the variable as the condition does not trigger the compiler transformation [4]. Future improvement of this mechanism is tracked at [issue #73137](https://github.com/golang/go/issues/73137) [2].

### What b.Loop does not change

In-loop per-iteration setup still requires manual `b.StopTimer` / `b.StartTimer`. There must be exactly one benchmark loop per benchmark function — `b.N`-style and `b.Loop`-style cannot coexist in the same benchmark. Every iteration of the loop must do the same work (the same invariant that applied to `b.N`) [2].

### Adoption guidance

For new benchmarks, prefer `b.Loop`. For existing benchmarks, migration is mechanical: replace `for n := 0; n < b.N; n++` (or `for range b.N`) with `for b.Loop()` and delete any `b.ResetTimer()` calls that existed solely to exclude setup before the loop.

---

## 7. Cross-Language Context (JMH, DoNotOptimize, black_box)

The dead-code elimination problem is not a Go quirk. Every compiled or JIT-compiled language with an optimizing compiler faces the same adversarial relationship between benchmark loops and the optimizer. Solutions have converged on a pattern: inject an artificial observable side effect that the compiler cannot prove is a no-op.

### Java / JVM — JMH Blackhole

The [Java Microbenchmark Harness (JMH)](https://openjdk.org/projects/code-tools/jmh/) uses a `Blackhole` object injected into every `@Benchmark`-annotated method. Calling `blackhole.consume(value)` creates an artificial observer for the value:

```java
@Benchmark
public void measureSum(Blackhole bh) {
    bh.consume(sum(1, 2)); // forces the compiler to retain the computation
}
```

The `consume()` implementation uses volatile reads of two distinct tombstone values and XORs the argument against them [6]. Because the tombstones are volatile and initialized to different values, the compiler cannot prove the conditional `(v ^ t1) == (v ^ t2)` is always false, so it cannot eliminate the branch — and therefore cannot eliminate the computation that produces `v`. The JMH documentation warns: "Implementing an efficient and correct blackhole is not a simple task... it requires significant JVM/compiler/performance expertise." [6]

### C++ — Google Benchmark DoNotOptimize / ClobberMemory

[Google Benchmark](https://github.com/google/benchmark) provides two primitives:

```cpp
// Forces the result of expr into a register or memory location,
// acting as a read/write barrier.
benchmark::DoNotOptimize(result);

// Forces the compiler to flush all pending writes to global memory.
benchmark::ClobberMemory();
```

`DoNotOptimize` prevents elimination of the value it wraps; `ClobberMemory` prevents deferral of writes to memory that the compiler might otherwise hoist out of the loop [7]. Critically, the Google Benchmark docs note that `DoNotOptimize(expr)` applied to a temporary does *not* always prevent optimization of the expression — the expression can still be constant-folded before `DoNotOptimize` sees it. The recommended pattern materializes the result into a local variable first:

```cpp
auto result = foo(0);
benchmark::DoNotOptimize(result); // pass lvalue, not temporary
```

This maps directly to Go's two-variable sink pattern.

### Rust — std::hint::black_box

Rust's standard library provides `std::hint::black_box(val)`, which takes a value and returns it unchanged but inhibits compiler optimization of the value across the call site. The recommended idiom is:

```rust
use std::hint::black_box;

fn benchmark_sum(n: u64) {
    for _ in 0..n {
        black_box(sum(black_box(1), black_box(2)));
    }
}
```

`black_box` on the inputs prevents constant folding; `black_box` on the output prevents DCE.

### The universal pattern

All three languages converge on the same insight: defeating compiler optimizations in a benchmark requires constructing artificial observable side effects. The compiler is correct to optimize; the benchmark must opt out of specific optimizations at specific points. No language has found a way to make benchmarks immune to these issues without explicit annotation or framework support — which is exactly what `testing.B.Loop` provides at the Go 1.24 language level.

---

## Key Takeaways

- **Sub-nanosecond timings are a red flag.** Anything under 1 ns/op for a non-trivial computation almost certainly means the compiler has eliminated the work. Verify with `-gcflags='-m'` and the sink pattern.
- **Always sink the result.** Assign function results to a local inside the loop and write the local to a package-level `var` after the loop. This is the canonical defense against DCE and is the de facto standard in the Go standard library.
- **`-benchmem` should be default.** Allocation counts often reveal more actionable regressions than timing alone. Make `go test -bench=. -benchmem` the baseline invocation.
- **Prefer `b.Loop` for new benchmarks (Go 1.24+).** It automates timer management, eliminates setup ramp-up cost, and enlists compiler support to prevent DCE — removing the most common `b.N` footguns at the language level.
- **Constant inputs are silent killers.** Any benchmark that calls a function with literal constant arguments risks constant folding. Use package-level variables for all benchmark inputs.

---

## Sources

1. Dave Cheney, "How to write benchmarks in Go," June 30, 2013. https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go — accessed 2026-07-22.

2. Go Team, "Evolving the Go benchmark API" (testing.B.Loop blog post). https://go.dev/blog/testing-b-loop — accessed 2026-07-22.

3. Austin Clements, Go proposal #61515: `testing: add B.Loop`. https://github.com/golang/go/issues/61515 — opened July 21, 2023; accepted; shipped in Go 1.24. Accessed 2026-07-22.

4. Go standard library, `testing` package documentation. https://pkg.go.dev/testing — accessed 2026-07-22. (Source for `B.Loop`, `B.N`, `B.ResetTimer`, `B.StopTimer`, `B.StartTimer`, `AllocsPerOp`, `AllocedBytesPerOp` documentation.)

5. Go 1.24 Release Notes. https://go.dev/doc/go1.24 — accessed 2026-07-22. (Source for `testing.B.Loop` section: "The benchmark function will execute exactly once per -count... Function call parameters and results are kept alive, preventing the compiler from fully optimizing away the loop body.")

6. OpenJDK JMH, `Blackhole.java` source. https://github.com/openjdk/jmh — accessed 2026-07-22. (Source for volatile tombstone mechanism and the JMH warning about implementation difficulty.)

7. Google Benchmark, User Guide — DoNotOptimize and ClobberMemory. https://github.com/google/benchmark/blob/main/docs/user_guide.md — accessed 2026-07-22.

8. Rust standard library, `std::hint::black_box`. https://doc.rust-lang.org/std/hint/fn.black_box.html — [UNVERIFIED: URL not fetched; existence confirmed by general knowledge of Rust std as of Go 1.21 era, and cross-referenced in Google Benchmark user guide cross-language comparison.]

9. Go issue #73137: future improvement to `B.Loop` compiler optimization beyond disabling inlining. https://github.com/golang/go/issues/73137 — referenced in Go blog post [2]; not separately fetched.

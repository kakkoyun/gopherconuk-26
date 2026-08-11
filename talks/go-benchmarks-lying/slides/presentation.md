---
marp: true
theme: gophercon-datadog
math: mathjax
html: true
paginate: true
header: "Why Your Go Benchmarks Are Lying · GopherCon UK 2026"
footer: " "
style: |
  .columns { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  .big    { font-size: 1.4em; }
  .medium { font-size: 1.0em; }
  .small  { font-size: 0.8em; }
  .tiny   { font-size: 0.6em; }
---

<!-- _class: title gopher-sage -->
<!-- _paginate: false -->
<!-- _header: "" -->

##### GopherCon UK · 2026

# Why Your Go Benchmarks Are Lying

### Kemal Akkoyun · Datadog

And How to Stop Them

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 01

# A Loose Cable

A systematic measurement error can hide in plain sight.

---

## September 2011

The OPERA collaboration announces muon neutrinos arriving **faster than light**.

Months of rechecking. The maths. The sensors. The calibration.

---

## The cause

An improperly seated **fibre-optic connector** in the GPS timing chain.

A **~73 ns** bias made neutrinos appear early.

<br>

A second fault, an oscillator defect, pushed the other way and *partially masked the first*.

<div class="small">

CERN press release, 22 Feb 2012 · Cartlidge, *Science* 335(6072):1027

</div>

---

## The point is not physics

A systematic measurement error can hide in plain sight,

look exactly like signal,

and survive review by people far more careful than you.

---

## Your setup

OPERA had an international collaboration of particle physicists.

You have `testing.B`, a laptop, and background Chrome tabs.

<br>

**The cables are your compiler, OS scheduler, and statistics.**

---

## Three questions

<div class="big">

1. Is the compiler measuring **real work**?
2. Is my sample **stable enough**?
3. Is the difference **large relative to the noise**?

</div>

---

# Local first

## Trust the number on your laptop

## before you push to CI

<br>

If you can't trust it locally, CI will industrialise the lie.

---

<!-- _class: vcenter -->

## Why listen to me

<div class="columns">
<div>

**Go and observability**

Prometheus Steering Committee

Maintainer of `client_golang`, `promu`, and OpenTelemetry Go compile-time instrumentation

</div>
<div>

**Built across the stack**

Former maintainer of Thanos, Parca, and parca-agent

Talks on Go tooling, instrumentation, benchmarking, and profiling at GopherCon, FOSDEM, KubeCon, and PromCon

</div>
</div>

---

<!-- _class: vcenter -->

## Why Datadog cares

<div class="columns">
<div>

### Inside customer workloads

Datadog ships SDKs for several languages.

Those SDKs consume part of a customer's CPU, memory, and latency budget.

</div>
<div>

### Continuous evidence

PGO reduced production CPU usage by **3.4%**.

Benchmark gates protect tracer and instrumentation hot paths.

</div>
</div>

<br>

### Instrumentation overhead is product correctness

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 02

# Why Benchmark?

Performance is a feature.

---

## Latency and throughput

<div class="columns">
<div>

### Latency

Time for a single operation to complete.

Low latency means a fast response.

*Users feel this directly.*

</div>
<div>

### Throughput

Operations completed per unit of time.

High throughput means more capacity.

*Your system's ceiling.*

</div>
</div>

---

## The cost of slowness

| Response time | User perception |
| --- | --- |
| 100-200 ms | Minimally noticeable |
| 300-500 ms | Quick but slightly slow |
| 1-3 s | Amount of work noticeable |
| 5-10 s+ | User switches away |

<br>

A 500 ms delay cost Google 20% of search traffic.

<div class="tiny">

Google search team data, cited by [Coding Horror](https://blog.codinghorror.com/performance-is-a-feature/)

</div>

---

<!-- _class: vcenter -->

> "Not all fast software is world-class,
> but all world-class software is fast."

Tobi Lütke · [X, 5 May 2024](https://x.com/tobi/status/1787139157078188180)

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 03

# Before You Optimize

Measure symptoms. Set targets.

---

## Is it actually slow?

Measure in production before writing a benchmark.

<div class="columns">
<div>

### Inside the process

- `pprof`: CPU, heap, goroutine, block
- Datadog Continuous Profiler

</div>
<div>

### Across the system

- [OpenTelemetry eBPF Profiler](https://github.com/open-telemetry/opentelemetry-ebpf-profiler)
- [Parca](https://parca.dev)
- p50 / p95 / p99 from production

</div>
</div>

<br>

**Benchmark the path your production evidence says is hot.**

---

## Is it worth optimizing?

Define a target before you start. Without one, you never finish.

- **SLOs:** "p99 < 200 ms" is an objective, not a wish
- **Error budgets:** within budget, optimization is optional
- **Amdahl's Law:** only the hottest path produces meaningful wins

*Benchmark what your profiler tells you is hot, not what looks interesting.*

---

## Further reading: finding what to optimize

[**Optimizing Go Code Without a Blindfold**](https://www.youtube.com/watch?v=oE_vm7KeV_E)
Daniel Martí · GopherCon 2019

pprof, benchmarks, and data-driven optimization in Go.
Martí covers *how to find what to optimize*;
this talk covers *how to trust the measurement once you have a target*.

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 04

# The Art of Benchmarking

Two kinds. Know which one you need.

---

## Microbenchmarks vs macrobenchmarks

<div class="columns">
<div>

### Microbenchmarks

- Isolated functions or operations
- Nanosecond-level precision
- Prone to compiler tricks
- Risk: **not representative**

`testing.B` is the focus of this talk.

</div>
<div>

### Macrobenchmarks

- End-to-end workflows
- Realistic production workloads
- Higher variance, harder to isolate
- Risk: **slow feedback loop**

Load testing tools

</div>
</div>

---

## When to use which

| Use case | Benchmark type |
| --- | --- |
| Comparing algorithms | Micro |
| Validating a specific optimization | Micro |
| Regression detection | Both |
| Capacity planning | Macro |
| User-facing latency targets | Macro |

<br>

*This talk focuses entirely on microbenchmarks with Go's `testing.B`.*

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 05

# Making the Compiler Honest

Layer 1

---

## The compiler is not a neutral observer

It reads your benchmark.

It notices the result is unused.

It removes the work **correctly**, according to the language specification.

<br>

Your loop still runs `b.N` times. It just runs empty.

---

## Dead-code elimination (DCE)

```go
func makeBuffer(n int) []byte {
    return make([]byte, n) // heap-escaping allocation
}

// Result unused → DCE fires.
func BenchmarkMakeBuffer_DCE(b *testing.B) {
    for range b.N {
        makeBuffer(64) // result discarded → call removed
    }
}
```

---

## Captured result: `make bench-dce`

```text
DCE       0.2532 ns/op   0 B/op   0 allocs/op
correct  11.14 ns/op  64 B/op   1 allocs/op
```

<br>

`make([]byte, 64)` is **unconditional**. There is no path that skips it.

Yet: **0 allocs/op**.

---

## `ns/op` can lie

## `allocs/op` cannot

<br>

A timer has a floor (~0.25 ns). An empty loop and a fast function look alike.

An allocation either **happened** or it **did not**.

<br>

**Always `-benchmem`.**

---

## The fix: the two-variable sink

```go
var sink []byte // package-level: compiler can't prove it is never read

func BenchmarkMakeBuffer_Correct(b *testing.B) {
    var s []byte
    for range b.N {
        s = makeBuffer(64)
    }
    sink = s // one global write per run, not per iteration
}
```

Writing to `sink` *inside* the loop adds a memory write per iteration.

---

## Constant folding

```go
s = bits.OnesCount(0b10110)   // every input is a constant
```

The compiler evaluates it at compile time. You benchmark a **constant load**.

<br>

Both versions time the same on Apple Silicon, near the timer floor.

The assembly tells the truth.

---

## Assembly check: `make asm-dce`

<div class="columns">
<div>

**Constant-folded**

```asm
MOVD  $3, R2
```

The literal `3`. No popcount.

</div>
<div>

**Correct**

```asm
MOVD    onesInput(SB), R3
VCNT    V0.B8, V0.B8
VUADDLV V0.B8, V0
```

An actual instruction.

</div>
</div>

<br>

Route inputs through a **package-level variable**.

---

## Inlining

Inlining is good in production. In a benchmark it *feeds* DCE.

Once the body is inlined into the loop, the compiler can see the unused result
and eliminate the inlined body.

<br>

```bash
go build -gcflags='-m'   # "can inline X" = candidate
```

Fix: non-constant input **and** capture the result. Either alone is not enough.

---

## Timer traps: one-time setup

```go
func BenchmarkHash_BN_WithSetup_Correct(b *testing.B) {
    data := make([]byte, 1024)
    copy(data, payload)
    b.ResetTimer()          // exclude setup
    var s [32]byte
    for range b.N {
        s = sha256.Sum256(data)
    }
    _ = s
}
```

`ResetTimer` zeroes elapsed time. It does **not** stop the timer.

---

## Timer traps: per-iteration setup

```go
for range b.N {
    b.StopTimer()
    input := buildFixture(fixtureSize) // not timed
    b.StartTimer()                     // restart BEFORE the work
    s = processString(input)           // only this is measured
}
```

Get the order wrong and you time the fixture, not the function.

---

## Captured result: `make bench-timer`

```text
buggy    415.8 ns/op  128 B/op  1 allocs/op
correct  550.6 ns/op  144 B/op  1 allocs/op
```

<br>

The **buggy** one looks 25% faster.

It is timing `buildFixture` and excluding the function under test entirely.

<br>

<div class="small">

A benchmark that measures the wrong thing does not look broken. It looks like good news.

</div>

---

## What if the timer never restarts?

`StopTimer` with **no** matching `StartTimer`.

<br>

The framework waits for enough timed duration.
The timer never runs, so duration never accumulates.

It doubles `b.N` and tries again. Forever.

<br>

<div class="small">

We tried it. It hung.

</div>

---

## `testing.B.Loop` in Go 1.24

```go
func BenchmarkHash_BLoop(b *testing.B) {
    data := make([]byte, 1024)   // setup excluded automatically
    copy(data, payload)

    var s [32]byte
    for b.Loop() {
        s = sha256.Sum256(data)
    }
    _ = s
}
```

<div class="small">

Proposal #61515, Austin Clements

</div>

---

## What `B.Loop` removes

| Behaviour | `for range b.N` | `for b.Loop()` |
| --- | --- | --- |
| Automatic ResetTimer at loop start | No | **Yes** |
| Automatic StopTimer at loop end | No | **Yes** |
| Benchmark fn called per ramp-up | Multiple times | **Once per `-count`** |
| Setup re-executes on ramp-up | Yes | **No** |
| DCE of loop body prevented | No | **Yes** |

The DCE prevention needs the condition written **literally** as `b.Loop()`.

---

## Question 1 answered

**Is the compiler measuring real work?**

<br>

Sink pattern. `-benchmem`. Check `allocs/op`. Prefer `b.Loop`.

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 06

# Statistical Interpretation

Layer 2

---

## A single number is a point sample

Two runs of the **same binary**, eight runs apart, on a loaded machine:

```text
BenchmarkMakeBuffer_Correct-16    41877204    39.39 ns/op
...
BenchmarkMakeBuffer_Correct-16    52521198    27.54 ns/op
```

<br>

**A 43% swing.** Run it once, file the PR, and you can be 43% off in either direction.

---

## `benchstat`

```bash
go test -bench=. -benchmem -count=20 -benchtime=1s . | tee new.txt
benchstat old.txt new.txt  # golang.org/x/perf/cmd/benchstat
```

| Environment | sec/op | Difference |
| --- | --- | --- |
| idle | `11.32n ± 5%` | n/a |
| noisy | `37.40n ± 25%` | `+230.34% (p=0.000 n=20)` |

---

## Read the output

- `11.32n`: the **median**, not the mean
- `± 5%`: spread of the distribution
- `p=0.000`: distinguishable from noise
- `~` instead of a delta means **no measurable difference**. That is a result.

---

## What `benchstat` won't tell you

It answers: *is A different from B?*

It does not answer: *is this machine a trustworthy place to ask?*

<br>

Coefficient of Variation (CV):

$$ CV = \frac{\sigma}{\mu} $$

<br>

Benchstat deliberately does not report CV. It compares distributions rather than
characterising the environment producing them. A separate pass takes about 20 lines of awk.

---

## Rules of thumb

| Question | Answer |
| --- | --- |
| How many runs? | `-count=10` is the floor. 20 is better. |
| Time or iterations? | `-benchtime=100x` for the most reproducible per-commit numbers |
| When is CV too high? | **Above ~5%, fix the environment before comparing anything** |
| Significant but tiny? | Effect size and significance are different questions |

---

## The p-hacking trap

"Rerun until you get the number you wanted."

<br>

Every rerun is a fresh draw from the distribution.

With enough draws, any noise pattern looks like signal.

<br>

**Set your run count before you look at results.**

---

## Question 2 answered

**Is my sample stable enough?**

<br>

`-count=10` minimum. Read the p-value. Check CV before you trust the comparison.

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 07

# Local Reproduction

Layer 3A

---

## The question nobody measures

Container isolation is standard benchmarking advice.

Almost nobody publishes **what it actually buys you** on a working laptop.

<br>

So we measured it.

---

## Isolation experiment: `make bench-docker`

Same benchmark. `-count=20 -benchtime=1s`. Apple M4 Max, 16 logical CPUs.

| Condition | mean ns/op | stddev | **CV%** |
| --- | --- | --- | --- |
| idle host | 11.46 | 0.54 | **4.75** |
| host with 16 background spinners | 34.97 | 6.60 | **18.88** |
| container pinned to core 0, same load | 16.28 | 0.85 | **5.25** |

<br>

Loaded: **3× slower and 4× noisier**. Pinned: back to the idle noise floor
while the host is still fully saturated.

---

## 5.25% is not a triumph

## It is a ceiling

<br>

Bare-metal Linux, Simultaneous Multi-Threading (SMT) off: **~0.05%**. A hundred times tighter.

---

## The macOS caveat

Docker Desktop on macOS runs containers **inside a Linux VM**.

- `--cpuset-cpus=0` pins **vCPU 0 inside that VM**, not a physical core
- The VM scheduler can still migrate that vCPU across physical cores
- Nothing inside the container can disable host SMT or pin the host clock

<br>

**What you get:** isolation from co-running processes. Real, and worth having.
**What you don't:** controlled hardware.

---

## The Linux toolbox

| Control | Command | What it buys |
| --- | --- | --- |
| CPU affinity | `taskset -c 0` | No scheduler migration, warm cache |
| Core isolation | `isolcpus`, `cset shield` | Exclusive cores; needs reboot or root |
| Priority | `nice -n -5`, `chrt -f` | Helps under load; `chrt` can starve your display |
| Frequency lock | `perflock` | Stable clock for the run |

---

## `perflock`: read the source

```bash
perflock go test -bench=. -count=10 -benchtime=2s ./...
```

- Writes `scaling_min_freq` / `scaling_max_freq` via **cpufreq sysfs**
- Not `intel_pstate`. The README says nothing about any of this.

<br>

**On macOS:** it builds, and the mutual-exclusion lock works. Frequency pinning does not.
The default `-governor 90` reads Linux sysfs and errors. Pass `-governor=none`
and you get serialisation between runs. Nothing more.

---

## The inner loop

```bash
benchdiff --base-ref=main ./...
```

Stash, run on the base ref, restore, run again, pipe both to `benchstat`.

<br>

Write change → `benchdiff` → read the interval → decide.

<br>

**Cheap wins:** close the indexer, airplane mode, let the machine reach thermal
steady state. Free variance reduction, zero setup.

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 08

# Escalating to CI

Layer 3B

---

## Why shared runners lie

Competing workloads. Variable CPU frequency. Non-dedicated last-level cache.

<br>

A real 10% regression **vanishes** into runner noise.

A phantom 10% regression **appears** where there is none.

<br>

This is an **environment** problem, not a statistics problem.

---

## AWS m5.metal results

| Configuration | Runtime | **CV** |
| --- | --- | --- |
| SMT enabled, CPU-bound | n/a | **~23%** |
| SMT disabled, task 1 | 737.37 ± 0.32 ms | **0.044%** |
| SMT disabled, task 2 | 737.93 ± 1.74 ms | **0.235%** |
| Dynamic Frequency Scaling (DFS) on, 1 task | 533.97 ± 2.046 ms | **0.383%** |
| DFS off, 1 task | 738.18 ± 0.306 ms | **0.041%** |

<br>

SMT off: **~100× less variance**. Turbo off: **~10×**. Three sysfs writes.

---

## Two patterns, opposing requirements

<div class="columns">
<div>

### PR gate

Fast. Strict. Every PR.

The benchmarks that have
regressed before.

`-count=6`, tight threshold.

</div>
<div>

### Nightly suite

Complete. Pinned runner.

All environment controls
applied.

`-count=20`, rolling window.

</div>
</div>

<br>

CI is for **detecting** regressions. It is not your primary measurement.

---

## Existing CI tools

| Tool | Fit |
| --- | --- |
| **bencher.dev** | Hosted, Go-native, PR comments; default recommendation |
| **github-action-benchmark** | Self-hosted, simple first gate |
| **Apache Otava** | Change-point detection over a rolling window |
| **gobenchdata** | Go-specific GitHub Pages dashboards |

<br>

The talk repository contains the full survey and setup notes.

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 09

# The CI Regression That Was a Speedup

---

## dd-trace-go #4891

A change to tracer and orchestrion internals across `ddtrace/tracer`. June 2026.

The benchmark bot comments: **`BenchmarkOTLPProtoSize` 6-9% slower than main.**

---

## Read the benchmark first

```go
// The entire timed loop:
proto.Size(tracesData)
```

A protobuf size computation on a struct built **before** `b.ResetTimer()`.

It never calls `ContextWithSpan`, `SpanFromContext`, or **any code the PR touched**.

<br>

Locally: **<0.1% run-to-run variance.** The 6-9% was specific to CI.

---

## Same machine, both branches

| Build | 1 span | 10 spans |
| --- | --- | --- |
| `main` | 883.3 ns/op | 7115 ns/op |
| `#4891` | **840.7 ns/op** | **6775 ns/op** |

<br>

<div class="big">

**The PR was faster.** CI had inverted the sign.

</div>

---

## The mechanism

Restructuring `context.go` shifted **function addresses** across the package.

That moved the hot `proto.Size` loop relative to cache-line and
branch-target-buffer boundaries.

<br>

At ~390 ns per iteration, small alignment shifts produce several-percent swings
**in either direction**, enough to flip the verdict on a runner that cannot
lock CPU frequency.

<br>

**Resolution: no code change.** The PR shipped as written.

---

## This is a known phenomenon

[**Performance Matters**](https://www.youtube.com/watch?v=r-TLSBdHe1A), Emery Berger · Strange Loop 2019

Code layout, meaning which symbol lands at which address, can swing
performance by ±10% or more. Berger demonstrated this by randomising
object placement and measuring the effect.

The dd-trace-go case is a live instance: the diff touched zero
performance-critical code, yet the linker placed one function differently,
and the benchmark flipped sign.

**Implication:** micro-benchmark deltas smaller than ~10% need statistical
confidence. This is precisely what benchstat gives you.

---

## The lesson

A noisy result can be **directionally wrong**.

<br>

<div class="big">

It can block a good change

and wave a regression through.

</div>

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 10

# Wire It Up

This Afternoon

---

## Three tools

```bash
honestbench ./...                                    # static analysis
benchgate   -pkg=./... -count=10 -cv-threshold=5    # CV gate
benchenv                                             # env diagnosis
```

Stdlib-only Go modules, also wrapped as Claude Code skills.

---

## `honestbench`: static analysis for benchmarks

Go Abstract Syntax Tree (AST) analysis. No benchmark needs to run. Finds problems before you measure.

```text
dce_bench_test.go:46:3: high: discarded-result:
    call to makeBuffer() result is discarded;
    compiler may eliminate this call via DCE

timer_bench_test.go:80:3: high: stoptimer-without-starttimer:
    b.StartTimer() called after the work under test;
    measured code runs while the timer is stopped

17 findings (2 high, 4 medium, 11 info) across 12 functions
```

---

## `benchgate`: CV gate

Runs the benchmark N times and fails if variance is too high to trust.

```text
# threshold 5%          → FAIL
benchgate: BenchmarkMakeBuffer_Correct  mean=11.8 ns/op  cv=5.2%  ✗
VERDICT: FAIL — 1/1 benchmarks exceed CV threshold 5.0%

# threshold 8%          → PASS
benchgate: BenchmarkMakeBuffer_Correct  mean=11.1 ns/op  cv=2.2%  ✓
VERDICT: PASS — all 1 benchmarks within CV threshold 8.0%
```

Catches the case where your environment is too noisy before the numbers go into a PR comment.

---

## `benchenv`: diagnose your environment

No flags needed. Run it before any serious benchmarking session.

<div class="columns">
<div>

### Hardware and scheduler

- SMT control
- CPU frequency governor
- Turbo Boost
- Load average

</div>
<div>

### Benchmark toolchain

- `perflock`
- `benchstat`
- `benchdiff`
- `GOMAXPROCS` versus `NumCPU`

</div>
</div>

Unsupported controls are reported as **unavailable**, not guessed.

---

## `benchenv` on this laptop

```text
benchenv: benchmarking environment diagnosis (darwin/arm64, 16 CPUs)

  [unavailable]  SMT control — no sysfs controls on macOS
  [unavailable]  CPU frequency — not exposed on macOS
  [unavailable]  Turbo Boost — no user-space control
  [ok]           load average — 1-min load 6.15 ≤ 8.0
  [warn]         perflock not installed
  [ok]           benchstat installed
  [warn]         benchdiff not installed

Summary: 3 ok, 2 warn, 4 unavailable.
```

---

<!-- _class: punchline dark -->
<!-- _paginate: false -->

# Four *unavailable* lines are the honest answer

---

## The minimum viable discipline

```bash
# 1. Add a sink. Assign your result to a package-level var.
# 2. Run enough samples.
go test -bench=. -benchmem -count=10 | tee new.txt
benchstat old.txt new.txt
# 3. Check the environment before you believe the comparison.
benchenv
```

<br>

Under an hour. The benchmarks you write after today will tell the truth.

---

# Three questions

1. Is the compiler measuring **real work**?
   <div class="small">sink pattern · <code>-benchmem</code> · <code>allocs/op</code> · <code>b.Loop</code></div>

2. Is my sample **stable enough**?
   <div class="small"><code>-count=10</code> · <code>benchstat</code> · CV &lt; 5%</div>

3. Is the difference **large relative to the noise**?
   <div class="small">read the interval · control the environment first</div>

---

<!-- _class: dark -->

# Take it with you

<div class="columns">
<div>

### Benchmarks you can audit

[github.com/kakkoyun/gopherconuk-26](https://github.com/kakkoyun/gopherconuk-26)

- `honestbench`: static benchmark checks
- `benchgate`: coefficient-of-variation gate
- `benchenv`: environment diagnosis
- Captured benchmark outputs and both decks

[Earlier FOSDEM version](https://youtu.be/8211fNI_nc4)

</div>
<div class="center">

![w:250](../../assets/gopherconuk-26-repo-qr.png)

**Scan for the repository**

</div>
</div>

---

<!-- _class: end gopher-rocket -->
<!-- _paginate: false -->

# Questions?

[Website](https://kakkoyun.me) · [LinkedIn](https://www.linkedin.com/in/kakkoyun/) · [Bluesky](https://bsky.app/profile/kakkoyun.me) · [X](https://x.com/kakkoyun_me) · [GitHub](https://github.com/kakkoyun)

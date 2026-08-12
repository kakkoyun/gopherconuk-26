---
marp: true
theme: gophercon-datadog-minimal
# fragment-floor: 27 — `*` bullets are Marp progressive-reveal fragments and are
# load-bearing; `make check/fragments` guards them against prettier.
math: mathjax
html: true
paginate: true
header: "Why Your Go Benchmarks Are Lying · GopherCon UK 2026"
footer: " "
style: |
  .columns { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  .columns3 { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 1rem; }
  .big    { font-size: 1.4em; }
  .medium { font-size: 1.0em; }
  .small  { font-size: 0.8em; }
  .tiny   { font-size: 0.6em; }
  .center { text-align: center; }
  /* Helpers ported from the FOSDEM deck for the arc-2 charts and tables.
     Kept deck-local rather than in the shared theme so the sibling talk
     is unaffected. */
  .hidden { visibility: hidden; }
  .hl-blue   { color: #00ADD8; font-weight: 700; }
  .hl-orange { color: #E8833A; font-weight: 700; }
  .centered-table table { margin-left: auto; margin-right: auto; }
  .comment { display: block; opacity: 0.45; font-size: 0.75em; line-height: 1.5; }
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

<div class="center">

![width:620](../assets/particles-break-light-speed-headline.png)

</div>

Months of rechecking. The maths. The sensors. The calibration.

---

## The cause

<div class="columns">
<div>

An improperly seated **fibre-optic connector** in the GPS timing chain.

A **~73 ns** bias made neutrinos appear early.

A second fault, an oscillator defect, pushed the other way and
*partially masked the first*.

</div>
<div class="center">

![width:380](../assets/opera-loose-cable.png)

</div>
</div>

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

1. Is the benchmark measuring **real work**?
2. Is my sample **stable enough**?
3. Is the difference **large relative to the noise**?

</div>

---

## Two scales, same three questions

<div class="columns">
<div>

### Arc 1 — local and micro

`testing.B` on your laptop.

Compiler behaviour, code layout,
statistics, machine control.

</div>
<div>

### Arc 2 — CI and macro

Whole-workload benchmarks over time.

Workload design, hardware control,
change detection.

</div>
</div>

<br>

**You will leave with:** the failure modes at each scale, the checks that catch
them, and three small tools that automate the checks.

---

<!-- _class: vcenter -->

## How I got here

I did not set out to care about measurement.

* I maintained `client_golang`, and every allocation showed up in someone's scrape budget.
* I worked on Parca, where a profiler that costs 5% CPU is not a profiler anyone deploys.
* Now I work on Go instrumentation at Datadog, where our SDKs run inside your process.
* Every one of those jobs punished me for trusting a benchmark I had not questioned.

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
<div class="hidden">

### Throughput

Operations completed per unit of time.

High throughput means more capacity.

*Your system's ceiling.*

</div>
</div>

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

<br>

*An optimization can improve one and damage the other.*

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

Measure symptoms. Set targets. Pick a scale.

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

* **SLOs:** "p99 < 200 ms" is an objective, not a wish
* **Error budgets:** within budget, optimization is optional
* **Amdahl's Law:** only the hottest path produces meaningful wins

*Benchmark what your profiler tells you is hot, not what looks interesting.*

---

## Further reading: finding what to optimize

[**Optimizing Go Code Without a Blindfold**](https://www.youtube.com/watch?v=oE_vm7KeV_E)
Daniel Martí · GopherCon 2019

pprof, benchmarks, and data-driven optimization in Go.
Martí covers *how to find what to optimize*;
this talk covers *how to trust the measurement once you have a target*.

---

<!-- _class: vcenter -->

## Every benchmark needs two properties

<div class="big center">

**representative** and **repeatable**

</div>

<br>

Microbenchmarks usually fail the first.

Macrobenchmarks usually fail the second.

---

## Microbenchmarks vs macrobenchmarks

<div class="columns">
<div>

### Microbenchmarks

- Isolated functions or operations
- Nanosecond-level precision
- Prone to compiler tricks
- Risk: **not representative**

</div>
<div class="hidden">

### Macrobenchmarks

- End-to-end workflows
- Realistic production workloads
- Higher variance, harder to isolate
- Risk: **hard to attribute a cause**

</div>
</div>

---

## Microbenchmarks vs macrobenchmarks

<div class="columns">
<div>

### Microbenchmarks

- Isolated functions or operations
- Nanosecond-level precision
- Prone to compiler tricks
- Risk: **not representative**

</div>
<div>

### Macrobenchmarks

- End-to-end workflows
- Realistic production workloads
- Higher variance, harder to isolate
- Risk: **hard to attribute a cause**

</div>
</div>

<br>

*This talk covers both. They fail differently, so they need different checks.*

---

## When to use which

<div class="centered-table">

| Use case | Benchmark type |
| --- | --- |
| Comparing algorithms | Micro |
| Validating a specific optimization | Micro |
| Regression detection | Both |
| Capacity planning | Macro |
| User-facing latency targets | Macro |

</div>

---

## Start macro, then drill

A microbenchmark can only tell you a function got faster.
It cannot tell you the *system* got faster.

<div class="columns">
<div>

### The order that works

1. Macro: does the workload move?
2. Profile: which component?
3. Micro: fix that component
4. Macro again: did it hold?

</div>
<div>

### The order that wastes weeks

1. Micro: make a function 40% faster
2. Ship it
3. Nothing measurable changes
4. Nobody knows why

</div>
</div>

---

<!-- _class: section -->
<!-- _paginate: false -->

###### ARC 1

# Local and Micro

Trust the number on your laptop before you push to CI.

---

## Arc 1: the three questions at micro scale

<div class="columns">
<div>

1. **Real work?**
   The compiler may delete it.
2. **Stable sample?**
   One run is a point, not a distribution.
3. **Above the noise?**
   Your laptop is not a controlled instrument.

</div>
<div>

### The local-first rule

If you cannot trust the number
on your own machine,

CI will only
**industrialise the lie**.

</div>
</div>

---

## But first: a benchmark bot comment

June 2026. A PR restructures `context.go` in `dd-trace-go`.

The benchmark bot comments:

<div class="big">

**`BenchmarkOTLPProtoSize` is 6-9% slower than main.**

</div>

<br>

Nothing in the diff touches OTLP encoding.

**Hold that thought.** We come back to it once we can read a benchmark properly.

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 1A

# Making the Compiler Honest

Question 1: is it measuring real work?

---

## The basic set of tricks

Everything in this section exists for one reason:

<div class="big">

to stop the compiler from deleting the work you are trying to time.

</div>

<br>

Four techniques. Each one is **test scaffolding**, not production style.

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

<div class="small">

**Test-only.** A package-level sink exists to defeat DCE in a benchmark.
Do not add sinks to production code.

</div>

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

## How to read the next slide

Two builds of the same benchmark, dumped with `go test -gcflags=-S`.

<div class="columns">
<div>

### What to look for

`MOVD $3, R2`
Load the literal `3`
into a register.

The answer is already
in the binary.

</div>
<div>

### Versus

`VCNT`, `VUADDLV`
ARM64 population-count
instructions.

The CPU is actually
counting bits.

</div>
</div>

<br>

**One is a constant. One is work.** That is the whole comparison.

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

<div class="small">

**Test-only.** `//go:noinline` is a diagnostic tool. It does not belong on production functions.

</div>

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

<!-- _class: section -->
<!-- _paginate: false -->

###### 1B

# The Regression That Was a Speedup

Back to that bot comment.

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

<br>

**Implication:** micro-benchmark deltas smaller than ~10% need statistical
confidence before you believe their direction.

---

## Berger's other point: causal profiling

A profiler tells you where time is spent. It does not tell you what happens
if you make that part faster.

<br>

**Causal profiling** answers the second question directly: it applies a
*virtual speedup* to one component and measures the effect on the whole program.

<div class="columns">
<div>

### A flat profile says

`encode` is 30% of CPU.

</div>
<div>

### Causal profiling says

Making `encode` 20% faster
moves end-to-end by 2%.

</div>
</div>

---

## Why that matters for benchmarking

<div class="big">

A component speedup does not imply a system speedup.

</div>

<br>

Contention, queueing, and dependencies between components decide how much of a
local win survives to the top.

<br>

A microbenchmark measures the component. Only a macrobenchmark measures whether
it mattered.

<div class="small">

Which is exactly why the second half of this talk exists.

</div>

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

###### 1C

# Statistical Interpretation

Question 2: is the sample stable enough?

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

* `11.32n`: the **median**, not the mean
* `± 5%`: spread of the distribution
* `p=0.000`: distinguishable from noise
* `~` instead of a delta means **no measurable difference**. That is a result.

---

## What `benchstat` won't tell you

It answers: *is A different from B?*

It does not answer: *is this machine a trustworthy place to ask?*

<br>

Coefficient of Variation (CV):

$$ CV = \frac{\sigma}{\mu} $$

<br>

Benchstat deliberately does not report CV. It compares distributions rather than
characterising the environment producing them.

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

<!-- _class: section -->
<!-- _paginate: false -->

###### 1D

# Local Reproduction

Question 3: is the difference above the noise?

---

## The question nobody measures

Container isolation is standard benchmarking advice.

Almost nobody publishes **what it actually buys you** on a working laptop.

---

## Isolation experiment: `make bench-docker`

Same benchmark. `-count=20 -benchtime=1s`.

| Condition | mean ns/op | stddev | **CV%** |
| --- | --- | --- | --- |
| idle host | 11.46 | 0.54 | **4.75** |
| host with 16 background spinners | 34.97 | 6.60 | **18.88** |
| container pinned to core 0, same load | 16.28 | 0.85 | **5.25** |

<div class="small">

Apple M4 Max, **darwin/arm64**, 16 logical CPUs. Container is **linux/arm64** under
Docker Desktop's Apple Virtualization.framework VM. No QEMU, no cross-architecture
emulation: the guest matches the host ISA.

</div>

---

## 5.25% is not a triumph

## It is a ceiling

<br>

Bare-metal Linux, Simultaneous Multi-Threading (SMT) off: **~0.05%**. A hundred times tighter.

---

## The macOS caveat

Docker Desktop on macOS runs containers **inside a Linux VM** (arm64 guest on an
arm64 host, via Apple's Virtualization.framework).

* `--cpuset-cpus=0` pins **vCPU 0 inside that VM**, not a physical core
* The VM scheduler can still migrate that vCPU across physical cores
* Nothing inside the container can disable host SMT or pin the host clock

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

## "Is there a Go tool that sets this up?"

Yes. `perflock`, by Austin Clements of the Go team.

```bash
perflock go test -bench=. -count=10 -benchtime=2s ./...
```

* Serialises benchmark runs so two never overlap
* Writes `scaling_min_freq` / `scaling_max_freq` via **cpufreq sysfs**

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

## Arc 1 answered

<div class="columns">
<div>

### 1 · Real work

Sink pattern. `-benchmem`.
Check `allocs/op`. Prefer `b.Loop`.

### 2 · Stable sample

`-count=10` minimum.
Read the p-value. Check CV.

</div>
<div>

### 3 · Above the noise

Pin the container.
Know your ceiling: ~5% on a
laptop, ~0.05% on tuned metal.

### And

A sub-10% micro delta can be
**layout noise**, not a change.

</div>
</div>

---

<!-- _class: section -->
<!-- _paginate: false -->

###### ARC 2

# CI and Macro

Same three questions. Different failure modes.

---

## Arc 2: the three questions at macro scale

<div class="columns">
<div>

1. **Real work?**
   Now: is the *workload* representative?
2. **Stable sample?**
   Now: across days, not just runs.
3. **Above the noise?**
   Now: the hardware is the variable.

</div>
<div>

### What changes

Micro fails on
**representative**.

Macro fails on
**repeatable**.

</div>
</div>

---

## A second bot comment, a different lesson

July 2026. A dependency-pinning fix in
`opentelemetry-go-compile-instrumentation`.

The overhead benchmark, run locally:

| Scenario | Overhead | Ceiling |
| --- | --- | --- |
| `multi` | **230%** | 150% |
| `largeidle` | **212%** | 150% |

<br>

Both blowing through the threshold. The fix looks like a serious regression.

---

## Why that reading was wrong

The PR changed dependency pinning in **`multi`** only.

`largeidle` shares **zero** bumped dependencies with the change. It runs
independent code. It *cannot* have been affected.

<br>

Yet it showed the same ~60% inflation.

<div class="big">

The common factor was not the code. It was the machine.

</div>

<div class="small">

Heavy parallel builds and integration tests had been running all session.
`benchmark/threshold` is a CI-only job on a dedicated runner; there is no local target.
Resolution: wait for CI.

</div>

---

## The mirror image

<div class="columns">
<div>

### Arc 1 · `#4891`

CI said **slower**.

The laptop was right.

Cause: code layout.

</div>
<div>

### Arc 2 · `#643`

The laptop said **slower**.

CI was right.

Cause: machine load.

</div>
</div>

<br>

<div class="big center">

Neither environment is authoritative by default.

</div>

<br>

The tell in both cases was the same: **a benchmark moved that could not have moved.**

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 2A

# Designing a Macrobenchmark

Representative is the hard part.

---

## Representative workloads

What does your application actually do?

* **CPU-bound:** number crunching, compression, encryption
* **I/O-bound:** database queries, API calls, file operations
* **Mixed:** most real-world applications

<br>

*Your benchmark workload should match your production workload.*

---

## Workload archetypes

<div class="centered-table">

| Archetype | Pattern | Characteristics |
| --- | --- | --- |
| **Idle** | Background workers, minimal load | Low RPS, minimal CPU, few workers |
| **Latency** | Microservices, APIs | High RPS, low CPU per request |
| **Throughput** | Queue workers, batch processing | Moderate RPS, high CPU, many clients |
| **Enterprise** | Business apps with DB/API calls | Moderate RPS, mixed CPU / I/O |

</div>

<br>

*Pick the archetype that matches your application's behaviour.*

---

## Which is why `largeidle` falsified `multi`

<div class="columns">
<div>

### `largeidle`

The **Idle** archetype.

Background work,
minimal CPU.

</div>
<div>

### `multi`

The **Enterprise** archetype.

gRPC and Redis,
mixed CPU and I/O.

</div>
</div>

<br>

Two archetypes exercising different code, moving by the same amount, in the same
direction, at the same time.

<div class="big">

Shared environment, not shared cause.

</div>

---

## Expressing a macro gate

Micro gates compare two commits. Macro gates compare against a **budget**.

```text
scenario: multi
baseline: uninstrumented process
ceiling:  150% overhead
result:   230%   → FAIL
```

<br>

An overhead ceiling answers a product question, not a statistical one:
*how much of the customer's machine are we allowed to consume?*

---

## What a macro gate needs at scale

One scenario on a laptop is a smoke test. A gate you can block a release on
needs more.

<div class="columns">
<div>

### Shape

* Dedicated hardware, not shared runners
* A budget per component, not one global number
* Several archetypes, not one workload
* Gating the release, not only the PR

</div>
<div>

### Why it is worth it

When your code runs **inside**
someone else's process, an
overhead regression is a
customer-visible defect.

That is the situation every
SDK, agent, and sidecar is in.

</div>
</div>

---

## Two macro traps

<div class="columns">
<div>

### Coordinated omission

When the system falls behind, a naive
load generator stops issuing requests
and **stops recording the latency it
caused**.

Your p99 improves because you
measured less.

</div>
<div>

### Non-deterministic inputs

Random fixtures, live dependencies,
and unpinned datasets change the
workload between runs.

You cannot attribute a delta to code
if the input moved too.

</div>
</div>

<br>

*Run long enough to reach steady state. Pin the inputs. Record what you pinned.*

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 2B

# Controlling the CI Environment

Question 3, at hardware scale.

---

## Why shared runners lie

Competing workloads. Variable CPU frequency. Non-dedicated last-level cache.

<br>

A real 10% regression **vanishes** into runner noise.

A phantom 10% regression **appears** where there is none.

<br>

This is an **environment** problem, not a statistics problem.

---

## What SMT actually is

**Simultaneous Multi-Threading** (Intel calls it Hyper-Threading) lets two
hardware threads share one physical core's execution units.

<div class="columns">
<div>

### The bet

Most code stalls on memory.
A second thread can use the
idle execution slots.

Good for throughput.

</div>
<div>

### The cost

Two CPU-bound threads on one
core **fight for the same units**.

Each one's runtime now depends
on what the other is doing.

</div>
</div>

<br>

For a benchmark, that is a co-tenant you cannot see and did not schedule.

---

## What's the impact of disabling SMT?

<div class="center">

bare metal, dynamic frequency scaling disabled
**2 CPU-bound tasks, <span class="hl-blue">same core</span> vs. <span class="hl-orange">separate cores</span>**

![width:520](../assets/environment-control-smt-experiment.svg)

</div>

---

## What's the impact of disabling SMT?

<div class="columns">
<div>

![width:430](../assets/environment-control-smt-experiment.svg)

</div>
<div>

| Task | mean ± stddev | CV |
| --- | --- | --- |
| smt-1 | 1537.64 ± 367.29 ms | <span class="hl-blue">23.887%</span> |
| smt-2 | 1536.88 ± 366.84 ms | <span class="hl-blue">23.869%</span> |
| no-smt-1 | 737.37 ± 0.32 ms | <span class="hl-orange">0.044%</span> |
| no-smt-2 | 737.93 ± 1.74 ms | <span class="hl-orange">0.235%</span> |

</div>
</div>

<div class="center">

**~100× less variance.** Also twice as fast, because the core stopped sharing.

</div>

---

## What DFS is

**Dynamic Frequency Scaling** lets the CPU change its own clock at runtime:
turbo when there is thermal headroom, throttle when there is not.

<div class="columns">
<div>

### Governors

The kernel picks a frequency
from a policy: `powersave`,
`performance`, `schedutil`.

</div>
<div>

### Why it ruins runs

Run 1 boosts. Run 20 is warm
and throttles.

Same code, different clock,
different answer.

</div>
</div>

---

## What's the impact of disabling DFS?

<div class="columns">
<div>

![width:430](../assets/environment-control-dfs-experiment.svg)

</div>
<div>

| Configuration | mean ± stddev | CV |
| --- | --- | --- |
| DFS on, 1 task | 533.97 ± 2.046 ms | <span class="hl-blue">0.383%</span> |
| DFS off, 1 task | 738.18 ± 0.306 ms | <span class="hl-orange">0.041%</span> |

</div>
</div>

<div class="center">

**~10× less variance.** Note the mean got *slower*: turbo is fast and inconsistent.

<div class="small">

A benchmark's job is to be comparable, not to post the best number you can reach once.

</div>

</div>

---

## Three sysfs writes

```bash
# Disable SMT
echo off | sudo tee /sys/devices/system/cpu/smt/control
# Pin the governor
echo performance | sudo tee /sys/devices/system/cpu/cpufreq/policy*/scaling_governor
# Disable boost
echo 0 | sudo tee /sys/devices/system/cpu/cpufreq/boost
```

<div class="columns">
<div>

### On bare metal

CV ~23% → **~0.05%**.

</div>
<div>

### In a VM

The hypervisor owns SMT.
Frequency is virtualised.
The write may **succeed and
be silently ignored**.

</div>
</div>

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 2C

# Detecting Change Over Time

Question 2, across days.

---

## A/B is the wrong model for CI

`benchstat old.txt new.txt` compares **two** distributions.

Continuous benchmarking has a **time series**: one measurement per commit,
for months, with hardware changes and dependency bumps in the middle.

<br>

Comparing each commit to its parent means you re-ask the noise question every
single time, and you miss slow drift entirely.

---

## What regressions actually look like

<div class="columns">
<div>

### In the slide deck

![width:440](../assets/ideal-performance-regression.png)

</div>
<div>

### In real data

![width:440](../assets/actual-performance-regression.png)

</div>
</div>

<div class="center">

A step change is obvious in a diagram and buried in variance in practice.

</div>

---

## Change-point detection

Instead of asking *"is this commit slower than its parent?"*, ask
*"where in this history did the distribution shift and stay shifted?"*

<div class="columns">
<div>

### Properties

* Handles non-normal data
* Finds multiple change points
* Adapts to each benchmark's noise
* Ignores one-off spikes

</div>
<div>

### Implementations

**ED-PELT** — Andrey Akinshin

**e-divisive means** — Matteson
& James, used by **Apache Otava**
(incubating) and Nyrkiö

Netflix documented this at scale

</div>
</div>

<br>

*This is the right model for nightly trending. It needs history to work.*

---

<!-- _class: section -->
<!-- _paginate: false -->

###### 2D

# Wiring It Into CI

Two patterns. Opposing requirements.

---

## Two patterns

<div class="columns">
<div>

### PR gate

Fast. Strict. Every PR.

The benchmarks that have
regressed before.

`-count=6`, tight threshold,
same-machine A/B.

</div>
<div>

### Nightly suite

Complete. Pinned runner.

All environment controls
applied.

`-count=20`, rolling window,
change-point detection.

</div>
</div>

<br>

CI is for **detecting** regressions. It is not your primary measurement.

---

## The feedback loop

<div class="center">

![width:900](../assets/bp-feedback-flow.png)

</div>

<div class="small center">

Benchmarks have to be locally reproducible for a developer to act on them.
That is why arc 1 came first.

</div>

---

## Existing tools

| Tool | Fit |
| --- | --- |
| **bencher.dev** | Hosted, Go-native, PR comments; default recommendation |
| **github-action-benchmark** | Self-hosted, simple first gate, threshold only |
| **Apache Otava / Nyrkiö** | Change-point detection over a rolling window |
| **gobenchdata** | Go-specific GitHub Pages dashboards |

<br>

The talk repository contains the full survey and setup notes.

---

## Keep a ledger of known false positives

`BenchmarkOTLPProtoSize` fired again eleven days later, on an unrelated PR.

That time it took **one comment** to dismiss:

<div class="small">

> Known `BenchmarkOTLPProtoSize` false positive — documented for #4891 as a
> same-package code-layout artifact. The benchmark touches no code I changed;
> local A/B was ~+0.3%. No action.

</div>

<br>

Without that note, the next engineer re-runs the whole investigation, or worse,
"fixes" it.

<div class="big">

Documented noise is benchmark hygiene.

</div>

---

## Arc 2 answered

<div class="columns">
<div>

### 1 · Real work

Match the archetype.
Pin the inputs.
Avoid coordinated omission.

### 2 · Stable sample

Time series, not pairs.
Change-point detection
over a rolling window.

</div>
<div>

### 3 · Above the noise

Bare metal. SMT off.
Governor pinned. Boost off.

Three sysfs writes buy
**~100×** on variance.

### And

A shared runner cannot be
fixed with more samples.

</div>
</div>

---

<!-- _class: section -->
<!-- _paginate: false -->

###### CLOSE

# Wire It Up

This Afternoon

---

## Three tools

```bash
honestbench ./...                                    # static analysis
benchgate   -pkg=./... -count=10 -cv-threshold=5    # CV gate
benchenv                                             # env diagnosis
```

Ships from [github.com/kakkoyun/benchlab](https://github.com/kakkoyun/benchlab)

```bash
go install github.com/kakkoyun/benchlab/cmd/...@latest
npx skills add kakkoyun/benchlab --all
```

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

## Three questions, two scales

| | Local + micro | CI + macro |
| --- | --- | --- |
| **Real work?** | sink · `-benchmem` · `b.Loop` | archetype · pinned inputs |
| **Stable sample?** | `-count=10` · benchstat · CV | time series · change points |
| **Above noise?** | pin the container · know the ceiling | bare metal · SMT + DFS off |

<br>

<div class="center">

A sub-10% micro delta can be layout noise.
A shared runner cannot be fixed with more samples.

</div>

---

<!-- _class: dark -->

# Take it with you

<div class="columns">
<div>

### Benchmarks you can audit

[github.com/kakkoyun/benchlab](https://github.com/kakkoyun/benchlab)

- `honestbench`: static benchmark checks
- `benchgate`: coefficient-of-variation gate
- `benchenv`: environment diagnosis
- Agent Skills for all three commands

[Talk repo, decks, and captured results](https://github.com/kakkoyun/gopherconuk-26)

[Earlier FOSDEM version](https://youtu.be/8211fNI_nc4)

</div>
<div class="center">

![w:250](../../assets/gopherconuk-26-repo-qr.png)

**Scan for the talk repo**

</div>
</div>

---

<!-- _class: end gopher-rocket -->
<!-- _paginate: false -->

# Questions?

[Website](https://kakkoyun.me) · [LinkedIn](https://www.linkedin.com/in/kakkoyun/) · [Bluesky](https://bsky.app/profile/kakkoyun.me) · [X](https://x.com/kakkoyun_me) · [GitHub](https://github.com/kakkoyun)

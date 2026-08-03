---
marp: true
theme: datadog
math: mathjax
html: true
paginate: true
header: "**Why Your Go Benchmarks Are Lying** · GopherCon UK 2026"
style: |
  /* Two-column layout */
  .columns { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  /* Font-size helpers */
  .big    { font-size: 1.4em; }
  .medium { font-size: 1.0em; }
  .small  { font-size: 0.8em; }
  .tiny   { font-size: 0.6em; }
  /* Table override — Marp injects white backgrounds by default */
  table, td, tr, th { background-color: transparent !important; }
  table { font-size: 0.7em; }
---

<!-- _class: lead -->
<!-- _paginate: false -->
<!-- _header: "" -->

# Why Your Go Benchmarks Are Lying

## (And How to Stop Them)

**Kemal Akkoyun** · Datadog

GopherCon UK 2026

---

<!-- _class: lead -->
<!-- _paginate: false -->

# A Loose Cable

---

## September 2011

The OPERA collaboration announces muon neutrinos arriving **faster than light**.

Months of rechecking. The maths. The sensors. The calibration.

---

## The cause

An improperly seated **fibre-optic connector** in the GPS timing chain.

A **~73 ns** bias, making neutrinos appear early.

<br>

A second fault — an oscillator defect — pushed the other way, *partially masking the first*.

<div class="small">

CERN press release, 22 Feb 2012 · Cartlidge, *Science* 335(6072):1027

</div>

---

<!-- _class: vcenter -->

## The point is not physics

A systematic measurement error can hide in plain sight,

look exactly like signal,

and survive review by people far more careful than you.

---

## Your setup

OPERA had an international collaboration of particle physicists.

You have `testing.B`, a laptop, and background Chrome tabs.

<br>

**The cables are your compiler, your OS scheduler, and your statistics.**

---

<!-- _class: vcenter -->

## Three questions

<div class="big">

1. Is the compiler measuring **real work**?
2. Is my sample **stable enough**?
3. Is the difference **large relative to the noise**?

</div>

---

<!-- _class: vcenter -->

# Local first.

## Trust the number on your laptop
## before you push to CI.

<br>

If you can't trust it locally, CI will industrialise the lie.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Layer 1
## Making the Compiler Honest

---

## The compiler is not a neutral observer

It reads your benchmark.

It notices the result is unused.

It removes the work — **correctly**, per the language spec.

<br>

Your loop still runs `b.N` times. It just runs empty.

---

## Dead-code elimination

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

## DEMO: `make bench-dce`

```text
BenchmarkMakeBuffer_DCE-16          0.2532 ns/op       0 B/op    0 allocs/op
BenchmarkMakeBuffer_Correct-16     11.14   ns/op      64 B/op    1 allocs/op
```

<br>

`make([]byte, 64)` is **unconditional**. There is no path that skips it.

Yet: **0 allocs/op**.

---

<!-- _class: vcenter -->

## `ns/op` can lie.

## `allocs/op` cannot.

<br>

A timer has a floor (~0.25 ns). An empty loop and a fast function look alike.

An allocation either **happened** or it **did not**.

<br>

**Always `-benchmem`.**

---

## The fix: the two-variable sink

```go
var sink []byte // package-level — compiler can't prove it's never read

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

Both versions time the same on Apple Silicon — near the timer floor.

The assembly tells the truth.

---

## DEMO: `make asm-dce`

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

Once the body is inlined into the loop, the compiler can see the result is unused —
and eliminate the now-inlined body.

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

## DEMO: `make bench-timer`

```text
BenchmarkProcess_TimerOrder_BUG-16          415.8 ns/op   128 B/op   1 allocs/op
BenchmarkProcess_PerIterSetup_Correct-16    550.6 ns/op   144 B/op   1 allocs/op
```

<br>

The **buggy** one looks 25% faster.

It is timing `buildFixture` and excluding the function under test entirely.

<br>

<div class="small">

A benchmark that measures the wrong thing does not look broken. It looks like good news.

</div>

---

<!-- _class: vcenter -->

## The one that hangs

`StopTimer` with **no** `StartTimer`.

<br>

The framework accumulates timed duration until it hits the target.
The timer never runs. Duration never accumulates.

It doubles `b.N` and tries again. Forever.

<br>

<div class="small">

We tried to demo it. Don't.

</div>

---

## `testing.B.Loop` — Go 1.24

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
|---|---|---|
| Automatic ResetTimer at loop start | No | **Yes** |
| Automatic StopTimer at loop end | No | **Yes** |
| Benchmark fn called per ramp-up | Multiple times | **Once per `-count`** |
| Setup re-executes on ramp-up | Yes | **No** |
| DCE of loop body prevented | No | **Yes** |

The DCE prevention needs the condition written **literally** as `b.Loop()`.

---

<!-- _class: vcenter -->

## Question 1 answered

**Is the compiler measuring real work?**

<br>

Sink pattern. `-benchmem`. Check `allocs/op`. Prefer `b.Loop`.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Layer 2
## Statistical Interpretation

---

## A single number is a point sample

Two runs of the **same binary**, eight runs apart, on a loaded machine:

```text
BenchmarkMakeBuffer_Correct-16    41877204    39.39 ns/op
...
BenchmarkMakeBuffer_Correct-16    52521198    27.54 ns/op
```

<br>

**A 43% swing.** Run it once, file the PR, and you're up to 43% off — either direction.

---

## `benchstat`

```bash
go test -bench=. -benchmem -count=20 -benchtime=1s . | tee new.txt
benchstat old.txt new.txt
```

```text
                      │ results/idle.txt │           results/noisy.txt           │
                      │      sec/op      │    sec/op     vs base                 │
MakeBuffer_Correct-16        11.32n ± 5%   37.40n ± 25%  +230.34% (p=0.000 n=20)
```

- `11.32n` — the **median**, not the mean
- `± 5%` — spread of the distribution
- `p=0.000` — distinguishable from noise
- `~` instead of a delta means **no measurable difference**. That is a result.

---

## What `benchstat` won't tell you

It answers: *is A different from B?*

It does not answer: *is this machine a trustworthy place to ask?*

<br>

$$ CV = \frac{\sigma}{\mu} $$

<br>

Benchstat deliberately doesn't report CV — it compares distributions rather than
characterising the environment producing them. Separate pass, ~20 lines of awk.

---

## Rules of thumb

| Question | Answer |
|---|---|
| How many runs? | `-count=10` is the floor. 20 is better. |
| Time or iterations? | `-benchtime=100x` for the most reproducible per-commit numbers |
| When is CV too high? | **Above ~5%, fix the environment before comparing anything** |
| Significant but tiny? | Effect size and significance are different questions |

---

<!-- _class: vcenter -->

## The p-hacking trap

"Rerun until you get the number you wanted."

<br>

Every rerun is a fresh draw from the distribution.

With enough draws, any noise pattern looks like signal.

<br>

**Set your run count before you look at results.**

---

<!-- _class: vcenter -->

## Question 2 answered

**Is my sample stable enough?**

<br>

`-count=10` minimum. Read the p-value. Check CV before you trust the comparison.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Layer 3a
## Local Reproduction

---

## The question nobody measures

Container isolation is standard benchmarking advice.

Almost nobody publishes **what it actually buys you** on a working laptop.

<br>

So we measured it.

---

## DEMO: `make bench-docker`

Same benchmark. `-count=20 -benchtime=1s`. Apple M4 Max, 16 logical CPUs.

| Condition | mean ns/op | stddev | **CV%** |
|---|---|---|---|
| idle host | 11.46 | 0.54 | **4.75** |
| host with 16 background spinners | 34.97 | 6.60 | **18.88** |
| container pinned to core 0, same load | 16.28 | 0.85 | **5.25** |

<br>

Loaded: **3× slower and 4× noisier**. Pinned: back to the idle noise floor —
while the host is still fully saturated.

---

<!-- _class: vcenter -->

## 5.25% is not a triumph.

## It is a ceiling.

<br>

Bare-metal Linux, SMT off: **~0.05%**. A hundred times tighter.

---

## The macOS caveat — say it out loud

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
|---|---|---|
| CPU affinity | `taskset -c 0` | No scheduler migration, warm cache |
| Core isolation | `isolcpus`, `cset shield` | Exclusive cores; needs reboot or root |
| Priority | `nice -n -5`, `chrt -f` | Helps under load; `chrt` can starve your display |
| Frequency lock | `perflock` | Stable clock for the run |

---

## `perflock` — read the source, not the README

```bash
perflock go test -bench=. -count=10 -benchtime=2s ./...
```

- Writes `scaling_min_freq` / `scaling_max_freq` via **cpufreq sysfs**
- Not `intel_pstate`. The README says nothing about any of this.

<br>

**On macOS:** it builds, and the mutual-exclusion lock works. Frequency pinning does not —
the default `-governor 90` reads Linux sysfs and errors. Pass `-governor=none`
and you get serialisation between runs. Nothing more.

---

## The inner loop

```bash
benchdiff --base=main ./...
```

Stash, run on the base ref, restore, run again, pipe both to `benchstat`.

<br>

Write change → `benchdiff` → read the interval → decide.

<br>

**Cheap wins:** close the indexer, airplane mode, let the machine reach thermal
steady state. Free variance reduction, zero setup.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Layer 3b
## Escalating to CI

---

## Why shared runners lie

Competing workloads. Variable CPU frequency. Non-dedicated last-level cache.

<br>

A real 10% regression **vanishes** into runner noise.

A phantom 10% regression **appears** where there is none.

<br>

This is an **environment** problem, not a statistics problem.

---

## The numbers — AWS m5.metal

| Configuration | Runtime | **CV** |
|---|---|---|
| SMT enabled, CPU-bound | — | **~23%** |
| SMT disabled, task 1 | 737.37 ± 0.32 ms | **0.044%** |
| SMT disabled, task 2 | 737.93 ± 1.74 ms | **0.235%** |
| DFS on, 1 task | 533.97 ± 2.046 ms | **0.383%** |
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

## Don't build this yourself

| Tool | Verdict |
|---|---|
| **bencher.dev** | Hosted, Go-native, PR comments — the default recommendation |
| **github-action-benchmark** | Self-hosted, simple, good for a first gate |
| **Apache Otava** | Change-point detection over a rolling window |
| **gobenchdata** | Go-specific, GitHub Pages dashboards |

<br>

Full survey and wire-up in the talk repo.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# The CI Regression
## That Was a Speedup

---

## dd-trace-go #4891

A restructure of `context.go` in `ddtrace/tracer`. June 2026.

The benchmark bot comments: **`BenchmarkOTLPProtoSize` 6–9% slower than main.**

---

## Read the benchmark first

```go
// The entire timed loop:
proto.Size(tracesData)
```

A protobuf size computation on a struct built **before** `b.ResetTimer()`.

It never calls `ContextWithSpan`, `SpanFromContext`, or **any code the PR touched**.

<br>

Locally: **<0.1% run-to-run variance.** The 6–9% was specific to CI.

---

## Same machine, both branches

| Build | 1 span | 10 spans |
|---|---|---|
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
**in either direction** — enough to flip the verdict on a runner that cannot
lock CPU frequency.

<br>

**Resolution: no code change.** The PR shipped as written.

---

<!-- _class: vcenter -->

## The lesson

A number from a noisy environment is not merely **imprecise**.

<br>

<div class="big">

It can be **directionally wrong**.

</div>

<br>

A gate that is directionally wrong blocks good changes
and waves bad ones through.

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Wire It Up
## This Afternoon

---

## Three tools

```bash
honestbench ./...     # static analysis: missing sink, DCE risk,
                      # StopTimer misorder, B.Loop migration hints

benchgate -pkg=./... -count=10 -cv-threshold=5
                      # runs N times, fails if CV exceeds the threshold

benchenv              # diagnoses SMT, governor, Turbo, load,
                      # and which tools you're missing
```

<br>

Stdlib-only Go modules. Each also wrapped as a Claude Code skill —
so an agent can run the same discipline.

---

## DEMO: `benchenv` on this laptop

```text
benchenv: benchmarking environment diagnosis (darwin/arm64, 16 CPUs)

  [unavailable]  SMT control — macOS does not expose SMT control via sysfs
  [unavailable]  CPU frequency governor — not exposed on macOS
  [unavailable]  Turbo Boost — not controllable from user space
  [ok]           load average — 1-min load 6.15 ≤ 8.0
  [warn]         perflock not installed
  [ok]           benchstat installed
  [warn]         benchdiff not installed

Summary: 3 ok, 2 warn, 4 unavailable.
```

<br>

Four `unavailable` lines on a Mac. That is the honest answer, printed rather than assumed.

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

<!-- _class: vcenter -->

# Three questions

<div class="big">

1. Is the compiler measuring **real work**?
   <div class="small">sink pattern · <code>-benchmem</code> · <code>allocs/op</code> · <code>b.Loop</code></div>

2. Is my sample **stable enough**?
   <div class="small"><code>-count=10</code> · <code>benchstat</code> · CV &lt; 5%</div>

3. Is the difference **large relative to the noise**?
   <div class="small">read the interval · control the environment first</div>

</div>

---

<!-- _class: lead -->
<!-- _paginate: false -->

# Thank you

**github.com/kakkoyun/gopherconuk-26**

Demo module · three CLIs · the full research corpus

<br>

Kemal Akkoyun · Datadog

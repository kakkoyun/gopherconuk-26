# Script: Why Your Go Benchmarks Are Lying

GopherCon UK 2026 · 50 minutes + 10 Q&A · 110 slides

**How to use this.** `SAY` lines are the words. They are written to be spoken,
not read — short sentences, one idea each. You do not need them verbatim; learn
the **beat** (the bold line before each block) and the numbers, and the words
will come. `DO` lines are delivery, not content.

**Memorise in this order:** the 15 beats → the three questions → the two
stories → the numbers. Everything else is recoverable on stage.

---

## The 15 beats

| # | Section | The one thing |
| --- | --- | --- |
| 01 | Why Benchmark? | Speed is a product decision |
| 02 | Why I Care | Consequences, not credentials |
| 03 | A Loose Cable | Careful people ship broken measurements |
| 04 | Before You Measure | Pick the target before the tool |
| 05 | Local and Micro | Trust the laptop first |
| 06 | Compiler Honest | The compiler deletes your work |
| 07 | Regression That Was a Speedup | Noise can be directionally wrong |
| 08 | Statistics | One number is not a measurement |
| 09 | Local Reproduction | Know your noise ceiling |
| 10 | CI and Macro | Same questions, hardware answers |
| 11 | Macrobenchmark Design | Representative is the hard part |
| 12 | CI Environment | Three sysfs writes buy 100× |
| 13 | Change Over Time | A/B cannot see drift |
| 14 | Wiring Into CI | Detect, don't measure |
| 15 | Wire It Up | Under an hour |

**Checkpoints:** §05 by 14:00 · §10 by 31:00 · §15 by 46:00.
If §10 has not started by 33:00, start cutting (ladder in `outline.md`).

---

## 01 · Why Benchmark? — 4 min

### Title slide

> **SAY:** Every benchmark you have ever written is a measurement system, and
> most of them are lying to you. Not on purpose — by default. I want to give you
> three questions that turn a number into something you can trust.

`DO` Do not introduce yourself yet. The personal context is §02.

### Measure customer happiness, not CPU usage

**Beat: the frame for the whole section.**

> **SAY:** One frame before we start. The number you are chasing is not CPU
> usage, or allocations, or requests per second. It is whether your user is
> happy. Everything else is a proxy. A benchmark measures the proxy. The point
> of the proxy is the user.

`DO` Land it, then straight into "is it actually slow".

### Is it actually slow?

> **SAY:** Before you write a single benchmark: check it is actually slow, in
> production. In-process, that is pprof, or a continuous profiler. Across the
> whole system, the OpenTelemetry eBPF profiler or Parca, and your p99.
>
> Benchmark the path your production evidence says is hot. Not the one that
> looks interesting.

### Could it go faster?

**Beat: slow ≠ improvable.**

> **SAY:** So you found it is slow. Next question: can it actually go faster?
> Slow does not mean improvable. There is a ceiling — memory bandwidth,
> instruction latency, the physics of the machine. Headroom is the gap between
> where you are and that ceiling. No headroom, and the "is it worth it"
> question is moot — you cannot optimize past the physics.

### Is it worth optimizing? (reveal)

> **SAY:** And decide what done looks like, or you will never stop.
>
> An SLO is a target, not a wish. If you are inside your error budget,
> optimizing is optional — go fix something else. And Amdahl: only the hot path
> pays. Ten percent off code that runs one percent of the time is nothing.

### Latency and throughput (two-step reveal)

> **SAY:** Two words, quickly, because people use them interchangeably and they
> are not.
>
> *(reveal 1)* Latency is one operation, start to finish. That is what a user
> feels.
>
> *(reveal 2)* Throughput is operations per second. That is your ceiling.
>
> They are not the same, and this matters: you can improve one and damage the
> other. Batching is the classic — throughput up, latency worse.

### The cost of slowness

**Beat: this is a product decision, not an engineering one.**

> **SAY:** Rough thresholds. Under two hundred milliseconds, nobody notices.
> Half a second, it feels a bit slow. Past a second, people notice they are
> waiting. Past five, they leave.
>
> Google measured this directly. Half a second of extra latency cost them
> twenty percent of search traffic.

`DO` Attribute to "the Google search team". Not Marissa Mayer.

### Lütke quote

`DO` Let them read it. Say one line, then move.

> **SAY:** Not all fast software is world-class. But all world-class software
> is fast.

### Finding what to optimize

> **SAY:** If you want the "what to optimize" half properly, Daniel Martí gave
> it at GopherCon 2019 and it is still the best talk on it. He covers what to
> optimize. I am covering whether you can believe the number when you get there.

### Finding what is worth optimizing

**Beat: plant Berger, paid off at the arc bridge.**

> **SAY:** And the talk that pairs with Martí's is Emery Berger's "Performance
> Matters". Same shelf, different question. Berger's point for us: in a
> concurrent workflow, the component that looks hot on a profile is not
> necessarily the one holding the wall time. We come back to that idea at the
> bridge between the two halves.

---

## 02 · Why I Care — 2 min

### How I got here

`DO` Thirty seconds. Consequences, not achievements. No titles, no counts.

> **SAY:** Quick word on why I care. I maintained `client_golang` — every
> allocation I added landed in somebody's scrape budget. Then Parca, where a
> profiler that costs five percent CPU is a profiler nobody deploys. Now Go
> instrumentation at Datadog, where our SDKs run inside *your* process.
>
> Every one of those jobs punished me for trusting a benchmark I had not
> questioned.

### Why Datadog cares

> **SAY:** That last one is the sharp version. Our code runs in your process.
> Every nanosecond we spend comes out of your budget, not ours. So measurement
> is not an engineering hobby for us, it is product correctness.
>
> One public number: profile-guided optimization took three point four percent
> off production CPU. You only get to claim that if you can measure it.

---

## 03 · A Loose Cable — 5 min

`DO` This is the story beat. It sets up the three questions, so it stays welded
to them — do not let the §02 bio bleed into it.

### September 2011 (headline image)

**Beat: a careful team, an extraordinary result.**

> **SAY:** In 2011 a team of physicists announced they had broken the speed of
> light. The OPERA collaboration fired neutrinos from CERN to a detector in
> Italy, 730 kilometres away. The neutrinos arrived early. Faster than light.
> They did not publish immediately. They spent months rechecking. The maths.
> The sensors. The calibration. Then they published, and asked the world for
> help.

### The cause (connector photo)

**Beat: it was a cable.**

`DO` Let them look at the photo for a beat before you speak.

> **SAY:** It was this. A fibre-optic connector, not quite seated, in the GPS
> timing chain. Seventy-three nanoseconds of bias. That was the whole result.
>
> And here is the part I love. There was a *second* fault — an oscillator —
> pushing the other way. The two errors partially cancelled. The bug was hiding
> behind another bug.

`DO` Do not put a number on the oscillator. Ledger row 9.

### Systematic error hides in plain sight

**Beat: the lesson, stated plainly.**

> **SAY:** The point is not that physicists were careless. They were the
> opposite. The point is that a precise instrument can be confidently wrong,
> and stay wrong, through months of expert review.

### Your cables

> **SAY:** OPERA had an international collaboration and fifteen thousand tonnes
> of detector. You have `testing.B`, a laptop, and thirty Chrome tabs. Your
> cables are the compiler, the scheduler, and your statistics. Let's go find
> them.

### Three questions

**Beat: the contract for the next fifty minutes.**

`DO` Pause after each one. This is the spine — they should hear it four times
today.

> **SAY:** Three questions. Is the benchmark measuring real work? Is my sample
> stable enough? And is the difference big compared to the noise?
>
> If you cannot answer all three, you do not have a measurement. You have a
> number.

### Two scales, same three questions

> **SAY:** We ask them twice. First half: your laptop, `testing.B`,
> microbenchmarks. Second half: CI, whole workloads, over time.
>
> Same questions, different answers, because they fail differently.
> Microbenchmarks usually fail on being *representative*. Macrobenchmarks
> usually fail on being *repeatable*. You will leave with the checks for both,
> and three small tools that automate them.

---

## 04 · Before You Measure — 4 min

### Every benchmark needs

**Beat: the frame for the whole talk.**

> **SAY:** Two properties. Representative — it does what production does.
> Repeatable — run it again, get the same answer.
>
> Hold those two words. Almost every failure in this talk is one of them
> breaking.

### Micro vs macro (two-step reveal)

> **SAY:** *(reveal 1)* Microbenchmarks isolate a function. Nanosecond
> precision. Very easy for the compiler to cheat. Their risk is being *not
> representative* — beautifully precise measurement of something that never
> happens.
>
> *(reveal 2)* Macrobenchmarks run the real workflow. Realistic. Their risk is
> attribution — something moved, good luck finding out what.
>
> This talk covers both, because they fail differently and need different
> checks.

`DO` This replaces the old "this talk is only microbenchmarks" line. Do not
reintroduce it.

### When to use which

`DO` Do not read the table. Point at two rows.

> **SAY:** Comparing two algorithms — micro. Capacity planning — macro.
> Regression detection needs both.

### Start macro, then drill

**Beat: the concrete answer to "so what do I do".**

> **SAY:** Here is the order that works. Macro first: does the workload
> actually move? Then profile to find which component. Then micro on that
> component, where you can iterate in seconds. Then macro again to confirm it
> survived.
>
> And the order that wastes a fortnight: make a function forty percent faster,
> ship it, watch nothing change, and never find out why.

---

## 05 · Local and Micro — 17 min total

### The three questions, micro scale

> **SAY:** First half. Your laptop.
>
> The three questions again, at this scale. Real work — the compiler may have
> deleted it. Stable sample — one run is a single point. Above the noise — your
> laptop is not a controlled instrument.
>
> And the rule for this half: if you cannot trust the number on your own
> machine, pushing to CI does not fix it. CI just industrialises the lie.

### A benchmark bot comment

**Beat: plant the puzzle. Do not solve it.**

> **SAY:** Before that, a real one. June this year, dd-trace-go. I restructure a
> file, push, and the benchmark bot says `BenchmarkOTLPProtoSize` is six to nine
> percent slower than main.
>
> Here is what bothers me. Nothing in my diff goes anywhere near OTLP encoding.
>
> Hold that thought. We come back to it once we can read a benchmark properly.

`DO` Resist solving it. The payoff is §07.

---

## 06 · Making the Compiler Honest — ~10 min

### Stop the compiler deleting your work

> **SAY:** Everything in this section does one job: stop the compiler deleting
> the work you are trying to time. Four techniques. All four are test
> scaffolding — none of this belongs in production code. I will say that again
> when it matters.

### The compiler is not neutral

> **SAY:** Your benchmark is just code. The compiler reads it, notices you never
> use the result, and removes the work. And it is *right* to. The spec allows
> it. Your loop still runs `b.N` times — it just runs empty.

### Dead-code elimination (code)

> **SAY:** `makeBuffer` allocates. The benchmark calls it and throws the result
> away. Watch what happens.

### `make bench-dce`

**Beat: the killer number.**

> **SAY:** Quarter of a nanosecond, versus eleven. Looks like a forty-fold
> speedup. But look at the right-hand column. Zero allocations.
>
> `make([]byte, 64)` has no path that avoids allocating. There is no branch, no
> cache. Zero allocations means the call is gone.

### `ns/op` can lie, `allocs/op` cannot

`DO` Land this slowly. It is the most reusable line in the talk.

> **SAY:** `ns/op` has a floor — a timer costs about a quarter of a nanosecond,
> so an empty loop and a very fast function look identical. But an allocation
> either happened or it did not. `allocs/op` cannot be talked into lying.
>
> Always pass `-benchmem`.

### The two-variable sink

> **SAY:** The fix. A package-level variable the compiler cannot prove is never
> read. Accumulate into a local inside the loop, then publish once after it —
> that way you are not paying for a global write on every iteration.
>
> And to be explicit: this is test-only. Do not put sinks in production code.

### Constant folding

> **SAY:** Second trick. Every input here is a literal, so the compiler just
> computes the answer at build time. You are now benchmarking how fast your CPU
> can load a constant.

### Reading the next slide

`DO` Two sentences per side. Most of the room does not read ARM64.

> **SAY:** We are about to look at assembly, so here is all you need. On the
> left, `MOVD $3` — move the literal three into a register. The answer is
> already baked into the binary. On the right, `VCNT` and `VUADDLV` — the
> actual ARM64 population-count instructions. The CPU is doing the work.
>
> One is a constant. One is work.

### `make asm-dce`

> **SAY:** And there it is. Left, the literal three. Right, real instructions.
> Fix is the same shape as before: route your inputs through a package-level
> variable so the compiler cannot see them.

### Inlining feeds DCE

> **SAY:** Inlining is good — in production. In a benchmark it makes things
> worse, because once your function body is pasted into the loop, the compiler
> can see the result is unused and delete the whole thing.
>
> `-gcflags=-m` tells you what is inlinable. And you need *both* fixes: a
> non-constant input and a captured result. Either one alone is not enough.
>
> `//go:noinline` is a diagnostic tool. Not production style.

### Timer: one-time setup

> **SAY:** Now timers. `ResetTimer` after your setup zeroes the elapsed clock.
> Note what it does *not* do — it does not stop the timer.

### Timer: per-iteration setup

> **SAY:** If you must build a fixture every iteration: stop the timer, build
> it, start the timer, *then* do the work. Get that order wrong and you are
> timing the fixture instead of the function.

### `make bench-timer`

> **SAY:** Here is the wrong order. Four sixteen versus five fifty. The broken
> one looks twenty-five percent faster — because it is timing the fixture and
> excluding the function it claims to measure.
>
> That is the thing about a benchmark measuring the wrong thing. It does not
> look broken. It looks like good news.

### `StopTimer` with no `StartTimer`

> **SAY:** One more. Stop the timer and never restart it. The framework waits
> for enough timed duration. The timer never runs, so duration never
> accumulates. So it doubles `b.N` and tries again. And again. Forever.

`DO` Straight face, then move. Do not editorialise.

### `b.Loop()` — Go 1.24

> **SAY:** Go 1.24 gave us a much better answer. `b.Loop`. Setup before the loop
> is excluded automatically. Timing stops at the end automatically. And it
> prevents dead-code elimination of the loop body.

### What `b.Loop` removes

`DO` Do not read the table. Point at two rows.

> **SAY:** Setup exclusion, and DCE prevention — those are the two that matter.
> One catch: the compiler only does this if you write `b.Loop()` literally in
> the condition. Assign it to a variable first and you lose it.

---

## 07 · The Regression That Was a Speedup — ~5 min

### Read the benchmark first

**Beat: pay off the plant.**

> **SAY:** Back to our bot comment. First thing I did — read the benchmark.
> That is the entire timed loop. One call, `proto.Size`, on a struct that was
> built *before* `ResetTimer`.
>
> It does not call `ContextWithSpan`. It does not call `SpanFromContext`. It
> does not call one line of the code I changed. And locally it was rock solid —
> under nought point one percent variance.

### Same machine, both branches

> **SAY:** So I built both branches on the same machine and compared them
> properly. Main: eight eighty-three. My branch: eight forty.
>
> My branch was *faster*. CI had not just got the size wrong. It had got the
> **sign** wrong.

`DO` Let this sit. It is the biggest surprise in the talk.

### The mechanism

> **SAY:** What actually happened: restructuring the file moved function
> addresses across the package. That moved the hot loop relative to cache lines
> and branch-target-buffer boundaries.
>
> At about three hundred and ninety nanoseconds an iteration, that is worth
> several percent — in either direction. Enough to flip a verdict on a runner
> that cannot pin its clock.
>
> Resolution: no code change. Nothing was wrong.

### A known phenomenon

> **SAY:** This is not folklore. Emery Berger's "Performance Matters" is the
> canonical talk. Code layout alone — which symbol lands at which address —
> swings performance by ten percent or more.

### Causal profiling

> **SAY:** Berger's other idea is worth stealing. A normal profiler tells you
> where time is spent. It does not tell you what happens if you make that part
> faster.
>
> Causal profiling answers the second question directly — it applies a *virtual*
> speedup to one component and measures the effect on the whole program. So
> instead of "encode is thirty percent of CPU", you get "making encode twenty
> percent faster moves end-to-end by two".

`DO` Those numbers are illustrative of the mechanism, not measured. Ledger 19.

### A component speedup is not a system speedup

**Beat: the bridge into the second half.**

> **SAY:** And that is the whole reason the second half of this talk exists.
> Contention, queueing, and dependencies decide how much of a local win
> survives to the top.
>
> A microbenchmark measures the component. Only a macrobenchmark tells you
> whether it mattered.

### A noisy result can be directionally wrong

> **SAY:** So the lesson from this story is not "CI is bad". It is that a noisy
> result is not merely imprecise. It can point the wrong way. It can block a
> good change, and wave a real regression straight through.

---

## 08 · Statistical Interpretation — ~5 min

### One number is a point sample

> **SAY:** Same binary. Same machine. Eight runs apart. Thirty-nine
> nanoseconds, and twenty-seven. A forty-three percent swing, and nothing
> changed but time.
>
> Run it once, file the PR, and you can be forty-three percent wrong in either
> direction.

### `benchstat`

> **SAY:** The tool is `benchstat`, from `x/perf`. Run with `-count`, pipe it in,
> and it compares distributions instead of numbers.

### Read the output (reveal)

> **SAY:** Four things. The number is the *median*, not the mean. The plus-minus
> is spread. `p` tells you whether it is distinguishable from noise.
>
> And a tilde means no measurable difference — which is a *result*. It is not an
> invitation to run it again until you get a delta.

### What benchstat won't tell you

> **SAY:** But there is a question benchstat deliberately does not answer. It
> tells you whether A differs from B. It does not tell you whether this machine
> is a trustworthy place to be asking.
>
> That is coefficient of variation — standard deviation over mean. One number
> for how noisy your environment is. Twenty lines of awk, and it is the check
> most people skip.

### Rules of thumb

> **SAY:** Ten runs is the floor, twenty is better. Above about five percent CV,
> stop comparing and go fix your machine — more samples will not save you. And
> significant and *large* are different questions.

### The p-hacking trap

> **SAY:** Last thing. Do not rerun until you like the answer. Every rerun is a
> fresh draw from the same distribution, and with enough draws any noise looks
> like signal.
>
> Pick your run count before you look at results. Write it down if you have to.

---

## 09 · Local Reproduction — ~5 min

### What does isolation actually buy?

> **SAY:** "Run it in a container" is standard advice. Almost nobody publishes
> what it actually buys you, so we measured it.
>
> Idle laptop, CV under five percent. Now sixteen background spinners: three
> times slower, four times noisier. Now the same load, but the benchmark pinned
> in a container: back to the idle noise floor, while the host is still fully
> saturated.

`DO` If asked about the platform — and you were asked last time — it is on the
slide: Apple M4 Max, arm64 host, arm64 guest, Apple Virtualization.framework.
No QEMU, no cross-architecture emulation.

### 5.25% is not a triumph / It is a ceiling

> **SAY:** But be honest about what that is. Five percent is not a good number.
> It is the *best* number this machine can give you. Bare-metal Linux with SMT
> off reaches nought point nought five. A hundred times tighter.

### The macOS caveat

`DO` Do not soften this.

> **SAY:** And on macOS, Docker runs inside a Linux VM. So `--cpuset-cpus=0`
> pins a *virtual* CPU inside that VM. The VM can still move it across physical
> cores. Nothing inside that container can disable host SMT or pin the clock.
>
> What you get is isolation from your other processes — that is real, and worth
> having. What you do not get is controlled hardware.

### The Linux toolbox

> **SAY:** On Linux you have actual controls. Affinity, core isolation,
> priority, and a frequency lock. Different tools for different noise sources.

### "Is there a Go tool for this?"

> **SAY:** Somebody asked me exactly this last time, so: yes. `perflock`, by
> Austin Clements from the Go team. It serialises benchmark runs so two never
> overlap, and it pins CPU frequency.
>
> One caveat I only found by reading the source: on macOS the mutual-exclusion
> lock works, but frequency pinning does not — the default governor flag reads
> Linux sysfs and errors out. Pass `-governor=none` and you get serialisation
> and nothing else.

### The inner loop

> **SAY:** Day to day, this is the loop. `benchdiff` stashes your change, runs
> the base, restores, runs again, pipes both into benchstat. Make a change, run
> it, read the interval, decide.
>
> And the free wins: close your indexer, go into airplane mode, let the machine
> reach thermal steady state. Zero setup, real variance reduction.

### Local and micro, answered

> **SAY:** So, first half. Real work: sink, `-benchmem`, watch `allocs/op`,
> prefer `b.Loop`. Stable sample: at least ten runs, read the p-value, check CV.
> Above the noise: pin it, and know your ceiling.
>
> And the one to remember: a delta under ten percent might just be code layout.

---

## 10 · CI and Macro — 15 min total

### The three questions, macro scale

> **SAY:** Second half. Now it is CI, whole workloads, and time.
>
> Same three questions, different shape. Real work becomes: is the *workload*
> representative? Stable sample becomes: stable across *days*, not runs. And
> above the noise — now the hardware itself is the variable.

### A second bot comment

**Beat: mirror the first story.**

> **SAY:** Second real one. July, OpenTelemetry Go compile-time
> instrumentation. I land a dependency-pinning fix, run the overhead benchmark
> locally, and: `multi` at two hundred and thirty percent overhead against a
> ceiling of one fifty. `largeidle` at two twelve.
>
> Both blown. Looks like I have broken something badly.

### `largeidle` shares zero changed dependencies

> **SAY:** Except. My change touched dependency pinning in `multi`. `largeidle`
> shares *none* of those dependencies. It runs independent code. It cannot have
> regressed.
>
> And yet it moved by the same sixty percent.
>
> When something that cannot have changed changes anyway, the common factor is
> not your code. It is the machine. I had been running parallel builds and
> integration tests all afternoon. I was benchmarking my own build.

`DO` The fix was to wait for CI. There is no local target for that job — it is
CI-only, on a dedicated runner.

### The mirror image

**Beat: this is why there are two stories.**

> **SAY:** Put the two side by side. First story: CI said slower, the laptop was
> right, and the cause was code layout. Second story: the laptop said slower, CI
> was right, and the cause was machine load.
>
> Neither environment is authoritative by default.
>
> And the tell was identical both times: a benchmark moved that could not have
> moved. That is the single most useful instinct I can give you today.

---

## 11 · Designing a Macrobenchmark — ~5 min

### What does your app actually do? (reveal)

> **SAY:** Representative is the hard part, so start here. Is your workload CPU
> bound, I/O bound, or — like almost everything real — mixed?
>
> Your benchmark workload should look like your production workload. Obvious,
> and almost nobody does it.

### Workload archetypes

> **SAY:** We use four archetypes. Idle — background workers, barely any load.
> Latency — microservices, high request rate, little CPU each. Throughput —
> queues and batch jobs. Enterprise — business apps with databases and API
> calls, mixed.

### Why `largeidle` falsified `multi`

> **SAY:** And now the earlier story gets sharper. `largeidle` is the Idle
> archetype. `multi` is Enterprise. Completely different code, completely
> different shape.
>
> They moved by the same amount, in the same direction, at the same time. That
> is not a shared cause. That is a shared *environment*.

### A macro gate is a budget

> **SAY:** Macro gates work differently from micro gates. Micro compares two
> commits. Macro compares against a *budget* — a ceiling you are not allowed to
> cross.
>
> That is a product question, not a statistical one: how much of the customer's
> machine are we allowed to take?

### What a macro gate needs at scale (reveal)

> **SAY:** To gate a release on it, you need dedicated hardware, a budget per
> component, several archetypes rather than one workload, and gating on the
> release rather than just the PR.
>
> Worth it when your code runs inside someone else's process — then an overhead
> regression is not an internal metric, it is a customer-visible defect. That is
> every SDK, every agent, every sidecar in this room.

`DO` Keep this generic. Ledger row 23 — do not ad-lib Datadog internals.

### Two macro traps

> **SAY:** Two traps worth naming. Coordinated omission: your system falls
> behind, your load generator stops issuing requests, and therefore stops
> recording the latency it caused. Your p99 improves because you measured less.
>
> And non-deterministic inputs. Random fixtures, live dependencies. If the input
> moved too, you cannot attribute the delta to your code.

---

## 12 · Controlling the CI Environment — ~5 min

### Why shared runners lie

> **SAY:** Shared runners: competing workloads, variable frequency, shared
> last-level cache. A real ten percent regression vanishes into the noise. A
> phantom ten percent appears out of it.
>
> That is an environment problem. You cannot fix it with statistics.

### What SMT actually is

> **SAY:** Two acronyms, because I used them last time without explaining them.
>
> SMT — simultaneous multithreading, Hyper-Threading on Intel. Two hardware
> threads share one physical core's execution units. The bet is that most code
> stalls on memory, so a second thread can use the idle slots. Great for
> throughput.
>
> Terrible for benchmarking. Two CPU-bound threads on one core fight for the
> same units, so your runtime now depends on what some other process is doing.
> A co-tenant you cannot see and did not schedule.

### What's the impact of disabling SMT?

> **SAY:** Here is the cost. Two CPU-bound tasks, same core versus separate
> cores.
>
> *(next slide)* Twenty-four percent CV sharing a core. Nought point nought four
> when they are not. About a hundred times less variance. And notice it is also
> twice as fast, because the core stopped being shared.

### What DFS is

> **SAY:** Second one. Dynamic frequency scaling — the CPU changing its own
> clock. Turbo when there is thermal headroom, throttle when there is not, with
> the kernel picking a governor.
>
> Why that ruins a benchmark: run one boosts. Run twenty is warm and throttles.
> Same code, different clock, different answer.

### What's the impact of disabling DFS?

> **SAY:** Ten times less variance with it off. And look — the mean got
> *slower*. Turbo is fast and inconsistent.
>
> Which is the point. A benchmark's job is to be comparable, not to post the
> best number you can hit once.

### Three sysfs writes

> **SAY:** All of that is three writes to sysfs. Disable SMT, pin the governor,
> kill boost. On bare metal that takes you from twenty-three percent CV to about
> nought point nought five.
>
> In a VM, none of it works. The hypervisor owns SMT, frequency is virtualised,
> and the write may succeed and be silently ignored — which is worse than
> failing.

### Noise goes all the way down

**Beat: the deflating punchline. Earn the laugh, then bank the point.**

`DO` Let the picture land before saying anything. It is Brendan Gregg shouting
into a disk array. Do not over-explain it.

> **SAY:** And once you have done all of that — pinned the governor, killed
> turbo, disabled SMT — there is still one more source of noise.
>
> That is Brendan Gregg shouting at a disk array. The vibration alone is enough
> to spike disk latency. Someone filmed it.
>
> So: don't shout in the datacenter.

`DO` Beat. Then the real point, quieter:

> **SAY:** Which is only funny until you realise it generalises. There is always
> another layer of noise underneath the one you just fixed. The job is not to
> eliminate it — you cannot. The job is to know how much is left, and whether
> your effect is bigger than that.

---

## 13 · Detecting Change Over Time — ~3 min

### A/B is the wrong model for CI

> **SAY:** One more thing CI gets wrong. `benchstat` compares two
> distributions. But CI does not have two distributions — it has a time series.
> One measurement per commit, for months, with hardware changes and dependency
> bumps in the middle.
>
> Compare each commit to its parent and you re-ask the noise question every
> single time, and you miss slow drift completely.

### What regressions look like

> **SAY:** This is what we draw when we explain regressions. And this is what
> the data actually looks like. A step change is obvious in a diagram and buried
> in variance in real life.

### Change-point detection

> **SAY:** The right question is not "is this commit slower than its parent". It
> is "where in this history did the distribution shift, and stay shifted".
>
> That is change-point detection. It handles non-normal data, finds multiple
> change points, and ignores one-off spikes. ED-PELT, e-divisive means — Apache
> Otava implements it, Netflix documented doing this at scale.
>
> It needs history, so it belongs in nightly trending, not a PR gate.

---

## 14 · Wiring It Into CI — ~3 min

### Two patterns

> **SAY:** Two patterns, and they want opposite things. A PR gate is fast and
> strict — a curated set of benchmarks that have regressed before, same-machine
> A/B, tight threshold. A nightly suite is complete and slow — pinned runner,
> every control applied, rolling window, change-point detection.
>
> Either way: CI *detects*. It is not your primary measurement. That is still
> your laptop.

### The feedback loop

> **SAY:** And this is why the first half came first. A benchmark has to be
> locally reproducible for a developer to actually act on it. A red check
> nobody can reproduce just gets ignored.

### Existing tools

> **SAY:** You do not have to build this. bencher.dev if you want hosted and
> Go-native. github-action-benchmark for a simple first gate. Apache Otava if
> you want real change-point detection. Full survey is in the repo.

### Keep a ledger of false positives

> **SAY:** Last one, and it is the cheapest thing here. That same benchmark
> fired again eleven days later on an unrelated PR. That time it took one
> comment to dismiss — because the first investigation was written down.
>
> Without that note, the next person re-runs the whole investigation. Or worse,
> "fixes" it. Documented noise is benchmark hygiene.

### CI and macro, answered

> **SAY:** Second half. Real work: match the archetype, pin the inputs. Stable
> sample: think in time series, use change points. Above the noise: bare metal,
> SMT off, governor pinned.
>
> And the one to remember: a shared runner cannot be fixed with more samples.

---

## 15 · Wire It Up — 4 min

### Three tools

> **SAY:** I built three small things so you do not have to remember all of
> this. They ship from a separate repo called benchlab — one `go install` gets
> you all three binaries, and if you use coding agents, one command adds them as
> skills.

`DO` The tools live in `github.com/kakkoyun/benchlab`, **not** the talk repo.
The talk repo has the decks and the captured results. Do not conflate them —
the QR goes to the talk repo.

### `honestbench`

> **SAY:** Static analysis over your benchmark source. It finds discarded
> results, timer ordering mistakes, `b.N` loops that should be `b.Loop`. Nothing
> has to run — it reads the AST. Seventeen findings across twelve functions in
> our demo package.

### `benchgate`

> **SAY:** Runs your benchmark N times and fails if the variance is too high to
> trust. Same benchmark, two thresholds — fails at five percent, passes at
> eight.
>
> The point is not that eight is a good default. The point is that the policy is
> explicit and a machine enforces it, instead of living in your head.

### `benchenv`

> **SAY:** And this one just tells you what your machine can and cannot do. Run
> it before a serious session.

### Four *unavailable* lines

`DO` Pause here. This is the closing idea.

> **SAY:** Look at what it says on my laptop. Four `unavailable`. No SMT
> control. No frequency control. No turbo control.
>
> That is the honest answer. macOS does not expose those, so it says so, instead
> of printing a green tick I would have believed.
>
> Knowing what you cannot control is worth as much as controlling it.

### The minimum viable discipline

> **SAY:** If you take one slide home, take this one. Add a sink. Run enough
> samples. Compare with benchstat. Check the environment before you believe the
> comparison.
>
> That is under an hour of work, and every benchmark you write after today will
> tell you the truth.

### Three questions, two scales

`DO` The final recap. Say the two bold lines slowly.

> **SAY:** So. Three questions, two scales. Real work, stable sample, above the
> noise — asked once on your laptop, once in CI.
>
> A sub-ten-percent micro delta can be layout noise.
> A shared runner cannot be fixed with more samples.

### Take it with you

> **SAY:** Two links. benchlab has the three tools and the agent skills. The
> talk repo — that is the QR — has both decks and every captured result in this
> talk, so you can check my numbers.
>
> And the earlier FOSDEM version is there if you want the language-agnostic take.

### Close

> **SAY:** Trust your numbers only after the compiler, the sample, and the
> machine have each earned it.
>
> Thank you.

`DO` Stop. Leave the questions slide up.

---

## Numbers to memorise

| Number | What |
| --- | --- |
| **~73 ns** | OPERA connector bias |
| **500 ms → −20%** | Google search traffic |
| **−3.4%** | Datadog production CPU, from PGO |
| **0.25 vs 11 ns · 0 vs 1 allocs** | DCE demo |
| **416 vs 551 ns** | timer-order demo, broken looks faster |
| **883 → 841 ns** | #4891, main vs branch |
| **43%** | swing, same binary |
| **4.75 / 18.88 / 5.25** | CV: idle / loaded / pinned |
| **~0.05%** | bare metal, SMT off |
| **230% / 212% vs 150%** | #643 overhead vs ceiling |
| **23.9% → 0.044%** | SMT on → off (~100×) |
| **0.383% → 0.041%** | DFS on → off (~10×) |

## Things to never say

- Do not attribute the Google number to Marissa Mayer — "the Google search team".
- Do not put a figure on the OPERA oscillator fault.
- The datacenter demo is Brendan Gregg's; the video lives on Bryan Cantrill's
  channel. Credit Gregg for the demo, and do not quote a latency number for it.
- Do not call Docker Desktop on macOS hardware isolation.
- Do not claim statistical significance proves practical importance.
- Do not describe Datadog's internal benchmarking setup — ledger row 23 is unresolved.
- Do not present causal profiling's 30%/2% as measured results.

## If challenged

Name the committed result file and the environment. Do not improvise a rerun on
stage. Every number above traces to `claims-ledger.md` or `demo/results/`.

## Setup

Present from the **bespoke HTML** if you want the reveals to animate:
`make serve/go-benchmarks-lying` → `localhost:8000/presentation.html`.
The PDF is a correct handout but shows every fragment at once.

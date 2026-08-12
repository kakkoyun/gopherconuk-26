# Talk outline: Why Your Go Benchmarks Are Lying

> **Format:** 60 minutes presented, Q&A opportunistic · advanced Go audience
> **Deck:** `slides/presentation.md` (129 pages, including progressive-reveal steps)
> **Demonstrations:** no live demos; every command and captured result stays in the repository

## Thesis

A benchmark is a measurement system. It can report a precise number while
measuring removed work, an unstable sample, or an uncontrolled machine. Trust a
Go benchmark only after answering three questions:

1. Is the benchmark measuring real work?
2. Is the sample stable enough?
3. Is the difference large relative to the noise?

**The same three questions apply at two scales, and the answers differ.** That
is the structure of the talk.

| | Local + micro | CI + macro |
| --- | --- | --- |
| Real work? | the compiler deletes it | the workload is unrealistic |
| Stable sample? | one run is a point | pairs cannot see drift |
| Above the noise? | your laptop is uncontrolled | the hardware moves |

Microbenchmarks usually fail **representative**. Macrobenchmarks usually fail
**repeatable**. Neither is a substitute for the other, and the talk covers both.

## Timing baseline

Dry run 2 was measured against wall-clock timestamps: 10:13 start, slide 60 at
10:44. That is ~31 minutes to the war story and ~38-40 minutes for the whole
original 76-slide deck — roughly **2 slides per minute**.

The earlier claim of 66 minutes in this file and in `speaker-notes.md` was an
estimate that the rehearsal disproved. It has been removed. Plan against
measured pace, not the estimate.

Physical page count is no longer a good time proxy, because progressive reveals
now spend several pages on one idea. Budget by section.

> **Note.** The running order changed (WHY first, OPERA third) and a basics
> section was added, so the 60-minute target now covers the expanded content.
> The per-section targets above are a construction from section content, not a
> measurement — a timed read-through is still outstanding (see `TODO.md`).

## Beat sheet

| Part | Section | Purpose | Target | Cumulative |
| --- | --- | --- | ---: | ---: |
| 1 | **01 Why Benchmark?** | Is it slow, could it go faster, is it worth it; latency/throughput, cost of slowness, Lütke; further watching | 6:00 | 6:00 |
| 1 | **02 Why I Care** | How I got here, why Datadog cares, SpeedLab, FOSDEM prior work | 3:00 | 9:00 |
| 1 | **03 A Loose Cable** | OPERA, the three questions, the two-scales roadmap | 4:00 | 13:00 |
| 1 | **04 Before You Measure** | Representative + repeatable, micro vs macro, when to use which, start macro | 4:00 | 17:00 |
| 1 | **05 Benchmarking, Quickly** | Write, run, read, profile, read compiler output — shared vocabulary | 5:00 | 22:00 |
| 2 | **06–10 Arc 1 — local and micro** | bot-comment plant · compiler honesty · timers and `B.Loop` · keepalive mechanism · code layout and Berger · statistics · local environment · recap | 17:30 | 39:30 |
| 3 | **11–15 Arc 2 — CI and macro** | #643 opener · macro design · SMT and DFS · change-point detection · upstream Go dashboard and baseline · CI patterns · presubmit/postsubmit · SlowBots · false-positive ledger · recap | 16:00 | 55:30 |
| 4 | **16 Wire It Up** | Three CLIs, punchline, minimum discipline, 3×2 recap, CTA | 4:00 | 59:30 |

**Rehearsal checkpoints:** §05 by 17:00 · arc 1 (§06) by 22:00 · arc 2
(§11) by 39:30 · close (§16) by 55:30 · finish by 59:30. If arc 2 has not
started by 40:00, apply the cut ladder live.

> **The slack is gone, and delivery was too dense.** The first delivery ran at
> 127 pages and the feedback was that it was hard to digest. The response was a
> deliberate trade rather than more content: three memes in, three slides out.
> The memes are ten-second beats with a `DO` cue saying so; the slides they
> replaced were a minute each, so the deck is page-neutral at 129 and about two
> minutes *cheaper* than before, with three places for the room to breathe.
>
> The cut ladder below is still expected rather than contingency. Do not spend
> the memes to buy time back: they are the pacing, and cutting them recreates the
> problem they were added to fix.

## Narrative order

### Why first, then the story

Open on *why benchmark at all* — the user is the metric, not the CPU. Is it
actually slow? Could it go faster? Is it worth optimizing? Then the vocabulary
(latency vs throughput, cost of slowness, Lütke) and a pointer to where the
"what to optimize" half lives (Martí). This front matter is the WHY; it earns
the right to spend fifty minutes on measurement.

Only then the speaker context — *why I care* — as a short beat about
consequences, not credentials. Datadog's stake lands here because the
"overhead is product correctness" line is the sharp version of the WHY.

### The story, then the contract

OPERA is beat three, not the cold open. A careful team, a surprising result, and
two faults that partially masked each other. Pivot to a laptop benchmark only
after the connector reveal. The three questions and the two-scales roadmap stay
welded to OPERA — they are the contract that sets up the arcs and must
immediately precede them.

### Two arcs, two opening incidents

Each arc opens on a real incident, and the two are deliberate mirror images.

- **Arc 1 — dd-trace-go #4891.** CI reported `BenchmarkOTLPProtoSize` 6-9%
  slower. Planted before the compiler section and left unresolved. Resolved in
  §1B once the audience can read a benchmark: the PR was actually faster, and
  code layout had flipped the sign.
- **Arc 2 — otel-go-compile-instrumentation #643.** A local overhead run
  reported `multi` at 230% and `largeidle` at 212% against a 150% ceiling.
  `largeidle` shares no changed dependencies, so it could not have regressed.
  The machine was the variable; CI on a dedicated runner was right.

The tell is identical in both cases: **a benchmark moved that could not have
moved.** Say that explicitly on the mirror-image slide.

### The bridge between arcs

Berger's causal profiling closes arc 1: a component speedup does not imply a
system speedup, because contention and queueing decide how much of a local win
survives. That is the reason arc 2 exists, and it is the transition line into it.

### Close

Return to the three questions as a 3×2 grid, one row per question and one column
per scale. The two lines that must land: a sub-10% micro delta can be layout
noise, and a shared runner cannot be fixed with more samples.

## Cut ladder

Apply in this order if the timed read-through runs long.

Rungs 1, 4 and 10 of the previous ladder have already been spent: "Upstream
splits it the same way", the feedback-loop diagram, and the "Inlining decisions"
`-m -m` slide were removed to pay for the memes. What is left:

1. The "what a macro gate needs at scale" slide (−1:00)
2. "Go runs this on Go", keeping "Why a baseline, not a pair" (−0:45) — the
   quote is the load-bearing half; the dashboard screenshot is corroboration.
3. "Seeing the hardware" (−0:30) — a pointer to `perfgo`, not an argument the
   talk depends on; the observer-effect slide before it already makes the point
   that `ns/op` cannot see a cache miss.
4. "How the keepalive works" (−1:00) — the `b.Loop` table already carries the
   actionable advice; this slide is the why-it-works footnote.
5. Coordinated omission and deterministic inputs, folded to one line (−1:30)
6. The SMT and DFS *mechanism* explainer slides, keeping the impact data (−1:30)
7. The benchmarking-basics "Reading the output" slide, keeping the rest of §05
   (−1:00) — last resort. Removing an explanation does not fix density, and
   §07 leans on `ns/op` and `allocs/op` being already defined.

**Never cut without review:** either opening incident, the DCE captured
evidence, the local CV experiment, the macOS caveat, the SMT and DFS impact
data, the 3×2 recap, the Datadog rationale, "Why not just add more runners?"
(it answers the objection the CI half always draws, in upstream's own words), the
observer effect (it generalises §07 from "the compiler is not neutral" to "your
harness is not neutral either", which is what earns the second half), or **any of
the three memes** (they are the fix for the density complaint from delivery one,
and they cost ten seconds each).

## Public results and sources to preserve

- OPERA connector bias and the two-fault account: CERN press release and
  Cartlidge, *Science* 335(6072):1027. Claims ledger row 9.
- Google search team latency result: claims ledger row 10.
- Local CV experiment: committed files under `demo/results/`.
- FOSDEM SMT and DFS variance data: claims ledger rows 1-3.
- dd-trace-go #4891: `07-war-stories.md` story 1, sources [S28].
- otel-go-compile-instrumentation #643: `07-war-stories.md` story 2, [S29],
  claims ledger row 18.
- Causal profiling: claims ledger row 19.
- Workload archetypes: claims ledger row 20, from the FOSDEM talk.
- Change-point detection: claims ledger row 21.
- PGO production reduction: <https://www.datadoghq.com/blog/datadog-pgo-go/>.
- Prior FOSDEM version: <https://youtu.be/8211fNI_nc4>.

## Open gate before delivery

Claims ledger **row 23** is `pending`. The "what a macro gate needs at scale"
slide has already been reframed to avoid unpublished statements about Datadog's
internal benchmarking infrastructure. If Kemal wants the Datadog-specific
version back, that row has to move to `verified` first.

## Audience resources

- Talk repository: <https://github.com/kakkoyun/gopherconuk-26>
- Benchmark tools and skills: <https://github.com/kakkoyun/benchlab>
- Binaries: `go install github.com/kakkoyun/benchlab/cmd/...@latest`
- Agent Skills: `npx skills add kakkoyun/benchlab --all`

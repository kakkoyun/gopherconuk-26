# Talk outline: Why Your Go Benchmarks Are Lying

> **Format:** 60 minutes including Q&A · advanced Go audience
> **Deck:** `slides/presentation.md` (109 pages, including progressive-reveal steps)
> **Target:** 50 minutes presented, 10 minutes Q&A
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

## Beat sheet

| Part | Section | Purpose | Target | Cumulative |
| --- | --- | --- | ---: | ---: |
| 0 | Cold open | OPERA, the connector, the three questions, the roadmap | 5:00 | 5:00 |
| 1 | Who and why | How I got here, Datadog's stake, latency/throughput, cost of slowness | 4:00 | 9:00 |
| 2 | Before you measure | Production evidence, SLOs, Amdahl, representative + repeatable, micro vs macro, start-macro | 5:00 | 14:00 |
| 3 | **Arc 1 — local and micro** | #4891 plant · compiler honesty · timers and `B.Loop` · code layout and Berger · statistics · local environment · recap | 17:00 | 31:00 |
| 4 | **Arc 2 — CI and macro** | #643 opener · macro design · SMT and DFS · change-point detection · CI patterns · false-positive ledger · recap | 15:00 | 46:00 |
| 5 | Tools and close | Three CLIs, punchline, minimum discipline, 3×2 recap, CTA | 4:00 | 50:00 |

**Rehearsal checkpoints:** arc 1 begins by 14:00 · arc 2 begins by 31:00 · close
begins by 46:00. If arc 2 has not started by 33:00, apply the cut ladder live.

## Narrative order

### Cold open

Start with OPERA, not a biography or an agenda. A careful team, a surprising
result, and two faults that partially masked each other. Pivot to a laptop
benchmark only after the connector reveal. Ninety seconds.

### Contract before credibility

Plant the three questions, then the roadmap slide that states both arcs and the
takeaways, and only then the personal and Datadog context. The bio is a story
about why measurement started mattering, not a credential list.

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

1. The "what a macro gate needs at scale" slide (−1:00)
2. The feedback-loop diagram (−1:00)
3. Coordinated omission and deterministic inputs, folded to one line (−1:30)
4. The second Berger slide, keeping causal profiling only (−1:00)
5. Assembly back to one slide, dropping the "how to read this" primer (−1:30)

**Never cut without review:** either opening incident, the DCE captured
evidence, the local CV experiment, the macOS caveat, the SMT and DFS impact
data, the 3×2 recap, or the Datadog rationale.

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

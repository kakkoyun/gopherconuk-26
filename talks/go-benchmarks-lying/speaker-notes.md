# Speaker Notes — Why Your Go Benchmarks Are Lying

GopherCon UK 2026 · 60 minutes · advanced audience
Deck: `slides/presentation.md` (54 slides) · Demo: `demo/`

---

## Timing plan

| Block | Slides | Target | Running total |
|---|---|---|---|
| Cold open (OPERA) | 2–7 | 4 min | 0:04 |
| Thesis / local-first | 8 | 1 min | 0:05 |
| Layer 1 — compiler honesty | 9–23 | 15 min | 0:20 |
| Layer 2 — statistics | 24–31 | 12 min | 0:32 |
| Layer 3a — local reproduction | 32–41 | 13 min | 0:45 |
| Layer 3b — CI | 42–47 | 6 min | 0:51 |
| War story | 48–52 | 4 min | 0:55 |
| Tools + close | 53–54 | 3 min | 0:58 |
| Q&A buffer | — | 2 min | 1:00 |

**Checkpoints.** If you are not at Layer 2 by **0:20**, you are behind — the compiler section is
the easiest place to lose five minutes. If you are not at the war story by **0:51**, cut straight
to it; the war story and the three questions are the two things people remember.

---

## Cold open (slides 2–7)

Ninety seconds on OPERA. Resist the physics. The audience does not need the baseline distance or
the detector mass, and every detail invites a question you do not want in the Q&A.

Land three beats: a precise instrument, a surprising result, a cable.

The **two-faults** detail is worth keeping — it is the part that maps onto benchmarking. Two
systematic errors pushing in opposite directions, partially masking each other, is exactly what a
noisy benchmark environment does.

**Do not say the 73 ns figure came from the OPERA erratum.** It did not. The erratum reports the
corrected result and contains no fault analysis. If asked: CERN's press release of 22 Feb 2012
for the cable, Cartlidge in *Science* 335(6072):1027 for the number.

Then the pivot — quick and a little flat, no drum roll:

> "Your benchmark just reported a 12% speedup. Did you ship a faster binary, or did you measure
> dead-code elimination? The cable isn't loose. The compiler removed the loop."

---

## Layer 1 — compiler honesty (slides 9–23)

### DEMO 1 — `make bench-dce`

```bash
cd talks/go-benchmarks-lying/demo && make bench-dce
```

About 20 seconds at `-count=10`. **The output is already on the slide as a fallback** — if the
terminal misbehaves, keep talking and point at the slide.

The line to land, slowly:

> "`make([]byte, 64)` is unconditional. There is no path through that function that skips the
> allocation. And it reports zero allocations per operation."

Let it sit. This is the moment the room gets it.

Then the principle: **ns/op has a floor, allocs/op has no floor.** An empty loop and a genuinely
fast function both sit near 0.25 ns and look identical. An allocation is discrete — it happened or
it did not, on any hardware. That is why this is the demo rather than a timing comparison.

### DEMO 2 — `make asm-dce`

Instant. The slide already has the diff, so the live run is optional. If you are behind schedule,
**this is the first thing to cut** — the DCE demo makes the same point and constant folding is the
third example, not the first.

### The hanging benchmark (slide 21)

Never run this live. Explain why it hangs: the framework accumulates *timed* duration until it
reaches the target, the timer never runs, duration never accumulates, `b.N` doubles forever.

The aside gets a laugh and costs three seconds: "We tried to demo it. That is how we found out."

### `B.Loop` (slides 22–23)

The table is the payload. Do not read it aloud — point at the two rows that matter (setup
exclusion, DCE prevention) and let people read the rest.

Mention the literal-syntax footgun: assigning the method to a variable first defeats the compiler
transformation. This audience appreciates that kind of detail.

---

## Layer 2 — statistics (slides 24–31)

The 43% swing is real captured data from `results/noisy.txt`, two runs of the same binary eight
apart. If someone asks whether the machine was deliberately loaded — yes, sixteen spinners. Say so.
It makes the point stronger, not weaker.

**`11.32n` is the median.** Not the geometric mean. Benchstat's geomean is a summary row across
multiple benchmarks. This is genuinely easy to get wrong on stage — a draft of the companion blog
post got it wrong — so if challenged: the median of those twenty runs is exactly 11.32, the geomean
is 11.444.

The `~` output deserves a beat. A benchmark reporting no measurable difference **has given you a
result**. People treat it as a failed run and re-roll. That is the p-hacking slide.

To save time here, drop the effect-size row from the rules-of-thumb table and move on.

---

## Layer 3a — local reproduction (slides 32–41)

### DEMO 3 — `make bench-docker`

**Do not run this live.** Three conditions at `-count=20 -benchtime=1s` plus container startup
takes several minutes and saturates the machine you are presenting from. The numbers are on the
slide and committed under `demo/results/`.

If you want something live, run `make cv` — it recomputes the table from committed output in under
a second and demonstrates the raw files are real.

### The macOS caveat (slide 38)

**Do not skip this and do not soften it.** Delivering the good result without the ceiling is the
exact failure this talk is about. The sequence:

1. Pinning takes a fully saturated machine from 18.88% CV back to 5.25%. Real and useful.
2. Bare-metal Linux with SMT off reaches 0.05%. A hundred times tighter.
3. Docker Desktop on macOS pins vCPUs inside a VM. You cannot disable host SMT or Turbo from there.

Then the summary: containers on a Mac buy isolation from your other processes. They do not buy
controlled hardware. Good enough to catch an obvious regression while developing; not good enough
to publish.

Expect a question about Colima, OrbStack, or Linux VMs generally. Honest answer: we measured Docker
Desktop, the VM layer is structural to all of them on macOS, and we have not measured the others.

### perflock (slide 40)

The detail worth having: **the README documents none of this.** The findings come from reading
`internal/cpupower/cpupower.go` and `cmd/perflock/main.go`. On macOS it builds and the lock works;
frequency pinning does not, because the default `-governor 90` reads Linux sysfs and errors.

This matters because "just use perflock" is standard Go advice and most of the room is holding a Mac.

---

## Layer 3b — CI (slides 42–47)

Fast section. The SMT/DFS table is the argument; everything else is scaffolding.

The line: **three sysfs writes and a `taskset` buy you two orders of magnitude.** Five seconds in a
CI step.

On the tool table, give the recommendation and move on. A survey that does not end in an answer
wastes the audience's time.

---

## War story (slides 48–52)

Tell it as a detective story, in order. Do not reveal the ending early.

1. The bot says 6–9% slower.
2. First instinct is that the restructure hurt the encoding path. Resist it — read the benchmark.
3. The timed loop is one line: `proto.Size(tracesData)`, on a struct built before `ResetTimer`. It
   never touches the changed code.
4. Locally, under 0.1% variance. So the noise is specific to CI.
5. Same machine, both branches: the PR is *faster*.
6. The mechanism: code layout shifted function addresses, moving a hot loop relative to cache-line
   and branch-target boundaries. At ~390 ns per iteration that is several percent, either direction.
7. Resolution: nothing. No code change. The PR shipped as written.

Then the lesson slide, plainly: not merely imprecise — **directionally wrong**.

Worth adding if time allows: the same benchmark tripped again eleven days later and was dismissed in
under a minute, because the false positive had been documented. Documented false positives pay for
themselves.

---

## Close (slides 53–54)

The three questions are the takeaway. Say them, then stop. Do not summarise the summary.

**Q&A seed** — if the room is quiet, answer the question they are thinking:

> "What is the single highest-leverage change for someone starting on Monday?"
> `-count=10` and `benchstat`. Everything else compounds on top of that.

---

## Contingencies

| Problem | Response |
|---|---|
| No network | Every demo is local. Nothing needs the network. |
| Docker not running | `make bench-docker` was never going to be live. Use the slide. |
| A benchmark hangs | It is the `StopTimer` bug. Ctrl-C, use it as the joke, move on. |
| Projector mangles the code | Full deck as PDF; demo output is on the slides. |
| 10 min behind at 0:35 | Cut Layer 3b to the SMT table alone, go straight to the war story. |
| 5 min short at the end | Take back cut material: constant folding, effect size, the eleven-days-later coda. |

**Cut order if running long** (first to go at the top):

1. `make asm-dce` constant-folding demo
2. Effect size vs significance
3. Core isolation and process priority detail
4. The tool comparison table — name the recommendation only
5. The eleven-days-later coda on the war story

Never cut: the DCE demo, the CV table, the macOS caveat, the war story, the three questions.

---

## Facts you might be challenged on

| Claim | Where it comes from |
|---|---|
| OPERA ~73 ns connector bias | CERN press release 22 Feb 2012; Cartlidge, *Science* 335(6072):1027. **Not** arXiv:1109.4897. |
| SMT 23% → 0.05% CV | FOSDEM 2026 experiments, AWS m5.metal |
| CV 4.75 / 18.88 / 5.25 | Our own run, `demo/results/`, M4 Max, `-count=20` |
| `allocs/op` 0 vs 1 | `demo/dce_bench_test.go`, reproducible via `make bench-dce` |
| dd-trace-go figures | PR #4891, merged 2026-06-12; same-machine A/B |
| perflock macOS behaviour | Source inspection, not the README |
| `B.Loop` in Go 1.24 | Release notes; proposal #61515 (Austin Clements) |

Everything above is a `verified` row in `research/go-benchmarks-lying/claims-ledger.md`.
If a claim is not in that file, do not make it on stage.

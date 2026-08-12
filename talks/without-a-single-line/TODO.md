# TODO — Zero-Touch Go Instrumentation

Open work for this talk. Closed items get deleted, not archived; git history is
the archive.

Companion files: `outline.md` (structure, timing) · `speaker-notes.md` (the spoken
script) · `slides/presentation.md` (the deck) ·
`../../research/without-a-single-line/claims-ledger.md` (every load-bearing claim).

**State:** 53 slides, 30-minute keynote, scripted estimate 30:30. Ledger holds 32
claims: 30 `CONFIRMED`, 1 `PLAUSIBLE`, 1 `REFUTED`. `outline.md` describes itself
as "additive review version; pruning follows Kemal's deck review and rehearsal",
so the deck is known to be un-pruned.

This file was written straight after delivering *Why Your Go Benchmarks Are
Lying*. The second half is the transferable recipe from that talk; the first half
is what that recipe says to do here, in priority order.

---

## Blocking

### Re-verify every version number in the ledger

The ledger was researched **2026-07-22** and six rows carry an explicit
"re-verify before talk/blog" note. This talk is a survey of fast-moving projects,
so version drift is the single largest correctness risk, and it is the kind of
error the audience will contain the maintainers of.

Numbers to re-check against the projects themselves, not against the ledger:

| Claim | Recorded value | Row |
| --- | --- | --- |
| OBI release | v0.10.0 (2026-06-30), still pre-1.0 "Development" | C-013 |
| otelc release | v1.0.1 (2026-07-14) | C-022 |
| orchestrion release | v1.11.0 (2026-06-25) | C-021b |
| ebpf-profiler | v0.0.202627, ISO-week tags | C-035 |
| dd-trace-go contrib count | 51 integrations, say "over 50" | C-003 |

C-013 matters most: if OBI has cut a v1.0 since research, "expect breaking
changes between minor releases" is no longer true and the slide is wrong.

### Resolve the two non-CONFIRMED claims

- **C-043** (`runtime/metrics` recommended set, golang/go#67120) is `PLAUSIBLE`
  and says so: "Verify exact title and current status ... before citing." Either
  confirm it against the issue or keep it off the slides.
- **C-021** is `REFUTED`: Orchestrion was **not** donated to OTel as otelc. They
  are two separate tools, co-founded SIG, different codebases. This is an easy
  thing to say wrong on stage in front of people who worked on both. Check the
  script does not drift back into "donated".

### Timed read-through

Same trap as talk one, and worth stating plainly: **30:30 is a construction from
section content, not a measurement.** On the previous talk that construction was
wrong three times running, and the deck went on stage untimed and came back with
"too dense to digest" as the feedback.

This deck is 53 slides in 30 minutes. That is roughly 1.8 slides per minute,
which is close to the ~2 per minute measured on the other talk, so the estimate
is plausible rather than safe. Time it, and write the real number here.

---

## Should do

### Decide the reveal question deliberately, once

This deck has `fragment-floor: 0`, zero `*` bullets and **55 `-` bullets**. The
other talk is `*`-only with a floor of 60. Neither is wrong, but the difference
should be a decision rather than drift.

If you want reveals, do **not** bulk-convert all 55: that produces 55 reveals and
a deck that advances a line at a time. Pick the lists where staging actually helps
and convert only those, then set the floor to that count so
`make check/fragments` protects them.

If you do not want reveals, leave it and the current frontmatter comment is
already correct.

### Prune, because 53 additive slides is the known state

`outline.md` says pruning is outstanding. Do it *before* rehearsing, not after,
and write a cut ladder while you do: an ordered list of what goes first with the
reason it is safe to lose, plus a never-cut list. On the other talk the ladder was
the single most useful artefact on the day, because it turns a live timing panic
into reading a list.

---

## Nice to have

### A `demo/` with captured evidence

The other talk has `demo/` with a Makefile whose targets regenerate every captured
output on a slide, and ledger rows that name the file each panel came from. This
talk has no equivalent, and it shows: the USDT proof-of-concept points at
`github.com/kakkoyun/go/tree/poc_usdt`, a personal fork, with the probe output on
the slide unreproducible from this repo.

At minimum, capture the probe output to a committed text file and cite it, so the
claim has an artefact behind it.

---

## Content structure

The shape that worked on the other talk, and what it implies here. This is the
half of the recipe that is about craft rather than process, and it is where this
talk currently differs most.

### Open on why, not on the artefact

The other deck used to open on its best story and was deliberately reordered to
why first, then why I care, then the story. The reason: the story sets up a
contract, and a contract means nothing until the audience knows why they should
want it.

This is recorded as contested rather than settled, and honestly so: opening on
framing is a colder open than opening on a story, and that trade was made with
eyes open.

**Here it is already handled and better than talk one's version.** This deck opens
on a direct question, who has wrapped every HTTP client and database call by hand,
then "what if you did not have to?". That is a why built out of the audience's own
pain rather than an abstraction. Do not replace it with a mechanism slide.

### One real incident per major section

The strongest device in the other talk was two war stories, one opening each half,
deliberately built as mirror images with the same tell: a benchmark moved that
could not have moved. Both are on the never-cut list.

Two things made them work beyond being true. The first was **planted and paid off
later**: the incident appears before the section that explains it, sits unresolved
for ten minutes, and is answered only once the audience can read the evidence
themselves. The second was that the two stories **rhymed**, so the second one
landed as recognition rather than as a new fact.

**This is the biggest structural gap here.** This deck contains no incident,
postmortem or war story anywhere: no "we hit this", no "this broke". It is a survey
of three mechanisms, and a survey is exactly the format that goes flat without
someone getting hurt in it. Even one story would change the shape: the time an
instrumentation gap cost you something, the debugging session that ended in "Go has
no hook point", the rollout that needed no rollout.

### Repeat the frame, never the argument

The other talk states its three questions five times, at two scales, and the script
says outright that the audience should hear them four times. That repetition is
load-bearing: it is the spine, and the close returns to it as a grid.

The density complaint after delivery was **not** caused by that. It was caused by
slides that restated an argument already made, and those are exactly what got cut
to pay for the memes. The distinction is the whole lesson:

- Repeating the **frame** orients people. Keep it.
- Repeating the **argument** is what makes a deck feel dense. Cut it.

### Memes are pacing, and they are budgeted

Three memes went in after delivery specifically because the feedback was "too dense
to digest". They are not decoration and they were not free: each is paid for by a
slide that repeated something, so the deck stayed page-neutral and got about two
minutes cheaper.

Rules that made them work rather than pad:

- **Placed where the talk is heavy**, not where a joke fits. Straight after the
  densest stretch, or on the bridge into the next section.
- **Ten seconds, and the script says so.** Every meme carries a `DO` cue with the
  time budget and a warning that explaining the joke spends the time it bought.
- **Tied to the material.** The best of the three is captioned with the exact
  numbers the audience saw two slides earlier, so it reads as a punchline rather
  than as a break.
- **On the never-cut list**, because cutting them to buy time recreates the
  problem they were added to fix.

**Here there are none, and this is a 30-minute keynote.** A keynote needs air more
than a workshop does. Two would be plenty.

### Keep the script in lockstep, and write it to be spoken

The deck and the script are one artefact. Slides carry mottos, data, visuals and
links; prose lives in the script; section headings match slide titles verbatim so
you can find your place by what is on screen.

Two failure modes, both hit on the other talk:

- **Slides shipped without script.** Five went in during a rush and needed a
  separate catch-up pass. Edit both in the same commit or the drift compounds.
- **A script written to be read rather than said.** `SAY` blocks are short
  sentences, one idea each, and `DO` lines are kept separate so the prose stays
  sayable. The instruction that makes it usable on stage is to memorise the beat,
  the bold line before each block, and let the words follow.

**Here the script is roughly 5.5 lines per slide against 9.4 on the other talk.**
That is thin for a keynote, where you have less room to recover if a beat does not
land. The section that most needs words is whichever one you would struggle to
improve without them.

---

## The recipe

What actually worked on *Why Your Go Benchmarks Are Lying*, in rough order of how
much it paid off.

### 1. A claims ledger, and the rule that gates it

Every load-bearing claim gets a row with a primary source and a status, and
**nothing that is not verified may appear on a slide.** This is the highest-value
practice by a wide margin, and it pays off in a way that is easy to miss: most of
its value is in the claims it stops you making.

On the other talk it caught, before delivery: a tool described as a continuous
profiler that is not one, a hosted service's headline feature that does not
support Go, and a compiler change whose direction had been remembered backwards.
Each would have been said confidently to the exact audience that would know.

Two habits make it work:

- **Write the row when you add the claim**, not in a later audit pass.
- **Put the trap in the notes column.** "Direction matters and is easy to get
  backwards", "attribute to the search team, not Mayer", "do not add a figure for
  the second fault". Future-you reads the notes, not the source.

The two talks currently use different ledger formats. Fine, but pick one if you
ever want tooling across both.

### 2. The deck and the script are one artefact

Slides carry mottos, data, visuals and links. Prose lives in the script. Section
headings in the script match slide titles verbatim, so you can find your place by
what is on screen.

The failure mode to avoid: shipping slides and writing the script "later". On the
other talk five slides went in without script entries and it took a separate pass
to fix. Edit both in the same commit.

### 3. Verify by looking, not by asserting

Text-level checks miss clipping, overflow and layout. Render the page and look at
it. Concretely: build the PDF, then `pdftoppm -f N -l N -r 55 -png` and open it.

This caught, in one session: a screenshot that overflowed the page and pushed the
pagination off, a full-page browser grab that was unreadable at slide size, a
caption bar covering a meme's punchline, and a two-column slide that read as one
paragraph from the back of a room. None of those were visible in the markdown.

### 4. Automate the tripwires, then trust them

`make check` is the contract: code-fence headers resolve to real files, external
sources appear in the ledger, the fragment floor holds, no slide text enters the
footer band, both themes build, Go tests pass. Cheap to run, and it catches the
class of error that a human reading the diff will not.

Two lessons about the gates themselves:

- **A gate that fires on you is the gate working.** `benchenv` refused to bless
  the machine for a re-measurement, correctly. Do not override your own tooling
  because you are in a hurry.
- **Read the class before you use it.** The `meme` class is a full-bleed
  background image plus a caption strip, and guessing produced three broken
  slides. `themes/gophercon-datadog/README.md` documents every class.

### 5. Measure on a quiet machine or not at all

If a slide shows a number you produced, the machine state when you produced it is
part of the claim. A CV experiment was run concurrently with a live Marp server
and two PDF builds, which made the "idle" baseline meaningless. The numbers were
discarded rather than shipped, and the discarded values were written into the
record so nobody re-runs it and believes them.

Close everything, or use a machine you are not typing on.

### 6. Cross-repo evidence goes stale silently

Captured output from another repository has no tripwire in this one. On the other
talk, one of three tool panels had drifted: it listed six checks under a summary
that counted nine, arithmetic that fails on screen in a talk about rigour.
Rebuild the binaries and re-run before any delivery, and record the result in the
ledger row so the next person knows when it was last checked.

### 7. Atomic commits, and say why

One logical change per commit, each standing on its own, with a body explaining
the reasoning and what was rejected. This is what made it possible to revert a
single decision without unpicking a day's work, and what makes the diff readable
six months later. Never list files; the diff already does.

### 8. Density is a content problem, not a slide-count problem

The other talk came back from delivery as "too dense to digest" at 127 pages. The
fix was not fewer slides and not more slides: it was three ten-second jokes traded
for three one-minute slides that restated things already said. Page-neutral, two
minutes cheaper, three places for the room to breathe.

Removing an *explanation* does not fix density. Removing *repetition* does.

### 9. Tooling notes, so you do not rediscover these

- **pi-subagents could not start a single child** during that session: 13 of 13
  failed in about 2.8 seconds each with no model activity, including a trivial
  probe. `claude -p` through the cmux shim hung too. Budget for single-threaded
  work, or verify delegation works before planning around it.
- **A formatter that ignores `.markdownlint-cli2.yaml` strips trailing
  punctuation from headings.** It removed a load-bearing comma from
  `## Measure customer happiness,` twice, despite the repo setting
  `MD026: false` exactly to allow it. Check headings after bulk edits.
- **`git worktree remove` needs `--force` when the tree is dirty**, and branches
  that are 0-ahead of main carry nothing worth keeping. Prune deliberately.

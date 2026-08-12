# TODO — Why Your Go Benchmarks Are Lying

Open work for this talk. Closed items get deleted, not archived; git history is
the archive.

Companion files: `outline.md` (structure, timing, cut ladder) ·
`speaker-notes.md` (the spoken script) · `slides/presentation.md` (the deck) ·
`../../research/go-benchmarks-lying/claims-ledger.md` (every factual claim).

---

## Blocking — must resolve before the talk

### Clear or cut claims-ledger row 23

**Status:** `pending`. This is the only ledger row not `verified`.

The row describes Datadog's internal macro benchmarking practice — dedicated
hardware, per-SDK overhead budgets, several archetypes, gating on the release.
No public source exists for any of it.

The slide ("What a macro gate needs at scale") has already been reframed to be
generically true, so nothing unverified is currently on screen. Two ways to
close this:

1. Confirm each line is fine to say publicly → move row 23 to `verified`, cite
   yourself as the source, and optionally restore the Datadog-specific framing.
2. Leave the slide generic → mark row 23 `dropped` with a note.

Either is fine. Leaving it `pending` is not, because the ledger's own rule says
a non-`verified` claim must not appear on a slide.

### Timed read-through

**Now mandatory, not outstanding.** The deck grew from ~110 to ~122 pages and
from 14 to 16 sections. Every prior timing number in the repo is stale twice
over — first by the WHY-first restructure, now by ~22 additional slides. The
checkpoints below are a *construction from section content*, not a
measurement. Treat them as a hypothesis until this runs.

The sparse-slide pacing risk is unchanged: talking *to* sparse slides runs
slower than talking *over* dense ones, and this deck got sparser, not denser.

Run it against the checkpoints in `outline.md`:

| Checkpoint | By |
| --- | --- |
| §05 Benchmarking, Quickly starts | 17:00 |
| §06 Local and Micro starts (arc 1) | 22:00 |
| §11 CI and Macro starts (arc 2) | 39:00 |
| §16 Wire It Up starts (close) | 53:00 |
| Finish | 57:00 |

If §11 has not started by 40:00, apply the cut ladder in `outline.md`.
Update the per-part targets in `speaker-notes.md` with whatever you actually
measure.

---

## Should do

### Fold in Scott's marked-up PDF

He said he would send annotated slides after dry run 2. Not received at the time
of the restructure. When it lands, diff his notes against what is already
addressed — most line-level feedback from the session is done, so only genuinely
new points need action.

### Decide how much of the reveal machinery you actually want

The deck uses Marp fragments (`*` bullets) for 56 reveals. They only animate in
the bespoke HTML, not the PDF.

Decide before the day: presenting from HTML (`make serve/go-benchmarks-lying`)
gets the reveals; presenting from PDF is more robust but shows every bullet at
once. If you choose PDF, the two-step reveal slides (latency/throughput, micro
vs macro) become redundant duplicate pages and should be collapsed.

---

## Nice to have

### perflock PR for Linux/Windows support

Mentioned in dry run 2. The talk cites perflock and documents that its frequency
pinning is Linux-only (ledger row 16). Upstreaming better platform support would
let you drop the caveat. Not blocking — the caveat is accurate and useful as-is.

### Send Scott the podcast link

For marketing to help promote, once published.

### Consider rasterising the DFS chart

`assets/environment-control-dfs-experiment.svg` is 9.5 MB with ~60k `<use>`
elements (a strip plot, one element per point). It renders fine (~3s, compresses
well in the PDF) and is marked `-diff` in `.gitattributes`, so this is purely
about repo weight. Only worth doing if 13 MB of assets starts to annoy.

---

## Cross-repo

The three CLIs and the agent skills now live in
[github.com/kakkoyun/benchlab](https://github.com/kakkoyun/benchlab), not here.

Two consequences worth remembering:

- **Two repos, one QR.** The QR on the closing slide points at the *talk* repo.
  benchlab is a separate link beside it. The script carries a `DO` cue about not
  conflating them on stage.
- **The captured tool outputs in the deck are now cross-repo evidence.** The
  `honestbench` / `benchgate` / `benchenv` outputs on slides came from those
  tools. If benchlab changes their output format, the slides silently go stale
  and nothing in this repo will catch it. Re-check those three slides against
  the current binaries before the talk.

---

## Done — do not redo

Kept short deliberately. These come up repeatedly, so they are recorded to stop
them being "fixed" again.

- **Talk is not microbenchmark-only.** It is two halves, local+micro and
  CI+macro, answering the same three questions at both scales. The old scope
  line was deleted on purpose.
- **No internal vocabulary on slides.** "Arc 1", "CLOSE", "1A" were planning
  scaffolding. Dividers are numbered `01`–`16`. The theme's number block fits
  exactly two characters; longer labels overflow and clip.
- **WHY-first order is deliberate.** The deck used to open on the OPERA cold
  open; `outline.md` even argued for it ("Start with OPERA, not a biography or
  an agenda"). The restructure reverses that on purpose — WHY → why I care →
  the OPERA story — so the audience has the frame before the contract the story
  sets up. Do not "fix" it back to OPERA-first. The craft risk is real (opening
  on SLO framing is a colder open than opening on a story), but the reversal is
  recorded here so it does not get quietly undone.
- **Slides carry mottos, data, visuals, and links only.** Prose belongs in
  `speaker-notes.md`. If you find yourself adding a paragraph to a slide, it
  goes in the script instead.
- **`*` is the only list marker on slides.** Every list is an intentional Marp
  reveal. A `-` bullet in the deck is a bug. `make check/fragments` guards the
  count; `.prettierignore` stops prettier rewriting the markers.

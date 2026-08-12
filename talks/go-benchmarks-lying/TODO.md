# TODO — Why Your Go Benchmarks Are Lying

Open work for this talk. Closed items get deleted, not archived; git history is
the archive.

Companion files: `outline.md` (structure, timing, cut ladder) ·
`speaker-notes.md` (the spoken script) · `slides/presentation.md` (the deck) ·
`../../research/go-benchmarks-lying/claims-ledger.md` (every factual claim).

**Delivered at GopherCon UK 2026** at 127 pages. Now 129 pages, 16 sections,
fragment floor 63, claims ledger 44 `verified` / 1 `dropped` / 0 `pending`. The
full `make check` surface passes. Everything below is what survived the talk or
was created by it.

Two slides were added *after* delivery and have never been presented: the
observer effect and the `perfgo` pointer, both at the end of §07. Rehearse that
section before the next outing.

---

## Open

### Record what the talk actually took

The single most valuable missing number in this repo. Every timing figure here
and in `outline.md` is still a *construction from section content*, never a
measurement, and it has now been wrong three times running: once before the
WHY-first restructure, once after it, and once after five slides were added on
the day-before push.

`outline.md` currently claims 59:30 against a 60-minute slot. Write down what it
really was, section by section if you have it, and rebase `outline.md` and the
per-part targets in `speaker-notes.md` on that. Until someone does, the cut
ladder is being applied against a guess.

If the talk was recorded, the recording is the measurement.

### Memes

Never placed. The theme has a `section.meme` class, and the project's
`memebo.at` MCP server has usable templates: `this-is-fine`, `two-buttons`,
`guy-pointing-at-mirror`, `distracted-boyfriend`. Candidate beats that can carry
a joke without losing the thread: after the OPERA reveal, the macOS caveat, the
p-hacking trap.

Pure jokes make no factual claim and need no ledger row. Attribution only where
the source requires it.

### Fold in Scott's marked-up PDF

Still not received. When it lands, diff his notes against what is already
addressed rather than working through it linearly.

### Send Scott the podcast link

For marketing to help promote, once published.

---

## Blocked

### Re-measure the Docker and macOS CV experiment on an idle machine

**This is the last factual soft spot in the deck.** Ledger row 8 claims roughly
a 1 to 2 percent CV floor for a pinned container on the Mac VM against about
0.05 percent on bare-metal Linux. That number is not confirmed by anything in
this repo.

**Attempt 1** was made and thrown away, for a reason worth keeping: it ran
concurrently with a Marp server and two PDF builds, so the "idle" baseline was
not idle. It produced idle 3.16 percent, noisy 5.82 percent, pinned 2.42 percent,
which contradicts both the committed evidence and row 8, and it also showed
pinning making the mean 2.3x slower. None of that is trustworthy, so
`demo/results/` still holds the older committed run and the deck still shows
idle 4.75, noisy 18.88, pinned 5.25.

**Attempt 2 was refused by our own tooling**, which is the useful part. Load
average was 12.98 / 43.14 / 45.03 with a browser at 42 percent CPU, and `benchenv`
reported `[warn] load average — close background applications before
benchmarking`, dropping its summary from `3 ok, 2 warn` to `2 ok, 3 warn`. The
gate on the closing slides did its job on the person who wrote it. Do not
override it; wait for a quiet machine.

Consequence: the `5.25% is not a triumph / It is a ceiling` slide rests on a run
nobody has reproduced. Redo it on a quiet machine with nothing else running,
then either confirm the slide or change both the slide and row 8.

The irony of contaminating a measurement for this particular talk is noted, and
is the reason it was discarded rather than shipped.

---

## Cross-repo

The three CLIs and the agent skills live in
[github.com/kakkoyun/benchlab](https://github.com/kakkoyun/benchlab), not here.

- **Two repos, one QR.** The QR on the closing slide points at the *talk* repo.
  benchlab is a separate link beside it. The script carries a `DO` cue about not
  conflating them on stage.
- **The captured tool outputs are cross-repo evidence with no tripwire.** If
  benchlab changes an output format, the slides go stale and nothing in this repo
  notices. This bit once already: see the `benchenv` entry below. Re-check all
  three panels against freshly built binaries before any future delivery, and
  record the result in ledger row 25.

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
  an agenda"). The restructure reverses that on purpose: WHY, then why I care,
  then the OPERA story, so the audience has the frame before the contract the
  story sets up. Do not "fix" it back to OPERA-first. The craft risk is real
  (opening on SLO framing is a colder open than opening on a story), but the
  reversal is recorded here so it does not get quietly undone.
- **Slides carry mottos, data, visuals, and links only.** Prose belongs in
  `speaker-notes.md`. If you find yourself adding a paragraph to a slide, it
  goes in the script instead.
- **`*` is the only list marker on slides.** Every list is an intentional Marp
  reveal. A `-` bullet in the deck is a bug. `make check/fragments` guards the
  count; `.prettierignore` stops prettier rewriting the markers.
- **The `b.Loop` removes-table is correct. Do not "fix" it.** All four rows were
  checked against the release notes, including `DCE prevented: Yes`, which
  survives the Go 1.26 change. **Go 1.26 is the fix, not the regression**, and
  the direction is very easy to get backwards. What 1.26 removed is the *inlining
  suppression* that 1.24 and 1.25 used to implement keepalive; the keepalive
  itself remains. Ledger rows 39 to 41 carry the detail, including that the
  #77654 regression did not reproduce on go1.26.5.
- **CodSpeed's Go support is walltime-only.** Its interesting low-variance
  CPU-simulation instrument does not support Go. The tools-table cell says so
  deliberately. Do not "improve" it into claiming simulated execution for Go.
  Ledger row 42.
- **`perfgo` is not a continuous profiler, and the slide must not say it is.** It
  wraps Linux `perf` and CPU PMU counters around a Go benchmark: cache-miss
  hotspots, cache-to-cache transfers as a false-sharing tell, IPC. Calling it
  continuous profiling miscredits it in front of its author, who presented it at
  FOSDEM 2026. It also needs Linux, which the slide and script both say. Ledger
  row 43. The slide has no logo because no asset exists here; do not invent one.
- **The observer-effect numbers are Harsanyi's, not ours.** 15073 / 7358 reused
  against 33547 / 35507 fresh, measured on an Intel i5-7360U. The claim on screen
  is that the fifty percent gap disappears, *not* that the fresh pair are the
  function's true timings, because allocating a matrix per iteration is itself
  being measured. The post also declines to attribute the effect to cache lines
  or prefetching, so neither does the slide. Ledger row 45.
- **Ledger row 23 is cleared, and the slide stays generic anyway.** Row 23 moved
  to `verified` as a scope-limited self citation on the row-30 precedent. The
  macro-gate slide was deliberately *not* reverted to the Datadog-specific
  framing, because nothing on screen depends on the specifics and there is no
  reason to spend the risk. Restoring that framing would be a new decision.
- **The `benchenv` panel was rebuilt against the real binary.** It used to list
  six checks under a summary that counted nine, which does not add up in front of
  an audience. It now carries all nine. The per-check explanations are omitted on
  purpose: they overflow the code panel, which clips horizontally with no
  warning, and paraphrasing them would misquote the tool.
- **The two Go wiki screenshots are cropped on purpose.** Full-page browser grabs
  are unreadable at slide size, and the dashboard grab overflowed the page and
  pushed the pagination off. `golang_perf_dashboard_charts.png` and
  `golang_slowbot_quote.png` are the cropped versions actually used; the
  originals are kept beside them. Do not swap the originals back in.
- **An editor or agent autofix strips the comma from `## Measure customer
  happiness,`.** It happened twice in one session. That slide is one sentence
  split across two headings, so the comma is load-bearing, and the repo sets
  `MD026: false` precisely to allow it. The project's own `make lint/md` is not
  the culprit; a formatter that ignores `.markdownlint-cli2.yaml` is. If the
  comma disappears again, restore it and suspect format-on-save.

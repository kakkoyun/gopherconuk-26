# TODO: Zero-Touch Go Instrumentation

Pre-delivery record for the keynote restructure (2026-08-13).

Companion files: `outline.md` (structure, timing, cut ladder) ·
`speaker-notes.md` (SAY/DO script) · `slides/presentation.md` (the deck) ·
`../../research/without-a-single-line/claims-ledger.md` (every load-bearing claim).

**State:** ~60 pages, 30-minute keynote, 23 fragment reveals, 7 sections.
Timing is a word-count construction (~22:30), not a measurement. The actual
read-through against a clock is still required before delivery.

Count `---` separators minus one to get pages.

---

## TODO

- [ ] **Timed read-through.** Read the SAY content aloud at delivery pace against
  a clock. Record splits at each divider. Write the real number into `outline.md`
  and `speaker-notes.md`. The word-count estimate is ~22:30 but on talk one the
  construction was wrong three times running.
- [ ] **Visual inspection of every changed page.** `make build/pdf` then
  `pdftoppm -f N -l N -r 60 -png` on every changed page. Specifically confirm:
  the prompt line on both joke slides is legible and clear of the footer band;
  the two joke slides are pixel-identical except the prompt; the Usama crop is
  readable at width:480; neither meme's caption strip covers its punchline; the
  injector code panel does not clip horizontally; all seven dividers show two
  digits; the USDT and OTel agent skills screenshots are readable.
- [ ] **Rehearse the cut ladder.** If running long, follow the eight-rung
  ladder in `outline.md`. Never cut anything on the never-cut list.

---

## Done: do not redo

These decisions are settled. They keep getting re-litigated, so the reason is
recorded here.

- **The open is the debugging loop, told straight.** No show of hands, by
  decision. The bug, the log line, the redeploy, the wrong place, the second
  redeploy. Then stop.
- **The agent joke is a plant, not decoration.** It pays off at the agentic
  close and at the take-home skill. Do not move or cut it for time. It is two
  slides, full-bleed, no title, no caption. The beat is spoken: deadpan fake-out,
  then the turn. Ten seconds total.
- **The injector is a teaser.** The mechanism is deliberately undisclosed, and
  the apparent contradiction with the LD_PRELOAD slide is the device, not a bug.
  Never add linking model, entry path, or binary-compatibility detail to any
  slide or script line. If asked, defer to next year's talk.
- **The static-PIE account is script depth.** It lives in the uprobe-cost
  section of `speaker-notes.md`, marked cuttable. Not Kemal's incident, not a
  slide. It is the maintenance cost of inferring a layout you did not compile.
- **Internal claims are speaker-attested ledger rows.** C-050 (Live Debugger),
  C-051 (injector teaser), C-052 (OBI roadmap), C-053 (Usama talk) follow the
  C-026 pattern. Not unsourced slide text.
- **All three screenshots are cropped on purpose.** The two joke screenshots
  are cropped to one identical 16:9 region (image 1 shifted 2px to align
  content). The Usama FOSDEM grab is cropped to title plus poster. The USDT PoC
  and OTel agent skills screenshots are cropped to drop GitHub chrome. Originals
  are kept beside the crops with an `_orig` suffix. Do not swap them back in.
- **Reveals are selective at 23, by decision.** Six lists use `*` fragment
  bullets. Lists inside `.columns` are excluded on purpose. `fragment-floor`
  is 23. This is not drift.
- **Tooling sits at the end because takeaways precede the call to action.** The
  take-home terminal slide and the SIG CTA come after the takeaways, not before.
- **"Combine the layers" was absorbed.** It restated the decision table and the
  practical-combination slide. The 20% adoption line moved to the
  practical-combination slide in its cleared wording. The decision table is
  retitled "When to reach for what."
- **Both mid-deck terminal slides were deleted.** Tooling was stranded
  mid-deck. It now lives at the end as one take-home terminal slide.
- **No em dashes on slides.** Colons and commas are used instead.

# AGENTS.md — gopherconuk-26

Conventions for agents working on the talks in this repository. The user's
global `~/AGENTS.md` still applies; this file adds what is specific to building
conference talks here, and overrides the global file where they disagree.

---

## The deck and the script are one artifact

**Slides carry mottos, data, visuals, and links. Prose lives in the script.**

Text on a slide competes with the speaker. A room that is reading is not
listening. So every slide is at most:

- a claim worth reading aloud (the motto)
- one piece of evidence — code, captured output, a data table, a chart
- a short cue, not a sentence
- links people might photograph

If you are adding a paragraph to a slide, it belongs in `speaker-notes.md`.

**The script is the words, and it is written to be spoken.** `SAY` blocks are
short sentences, one idea each. `DO` lines are delivery cues, kept separate so
the prose stays sayable. It is structured for memorisation — beats first, then
the numbers — not as reference material.

**They must stay in sync.** Any change to a slide changes what gets said. When
you edit one, edit the other in the same commit. Section headings in the script
match slide titles verbatim so the speaker can find their place by what is on
screen — preserve that when renaming.

Verify coverage after restructuring: every content slide should have a scripted
entry.

## Per-talk `TODO.md`

Each talk directory keeps a `TODO.md` for open work, split into blocking,
should-do, and nice-to-have. It ends with a short **Done — do not redo** section
recording decisions that keep getting re-litigated, with the reason.

Closed items are deleted, not archived. Git history is the archive.

## No internal vocabulary on slides

Planning scaffolding — "arc", phase names, section codes like `1A`/`2B` — is for
the outline, never the deck. The audience did not read the plan.

The section divider number block in `gophercon-datadog` is a fixed 230×230px box
with a 128px font: **it fits exactly two characters.** Anything longer clips and
wraps. Use two-digit numbers.

## `*` is the only list marker on slides

Marp's bespoke template turns `*` list items into progressive-reveal fragments,
which makes the marker load-bearing rather than a style choice.

- Every list in a deck is an intentional reveal. A `-` bullet in a deck is a bug.
- Prettier rewrites `*` to `-` and silently destroys every reveal. Decks are in
  `.prettierignore` and `make format/md` is a deliberate no-op on them.
- Some editor and agent format-on-save hooks rewrite `-` to `*`, silently
  turning reference lists into reveals. Keeping decks `*`-only makes that a
  no-op too.
- `make check/fragments` fails when a deck drops below its declared
  `fragment-floor`. Update the floor when you intentionally change reveal count.

Reveals only animate in the bespoke HTML. The PDF is a correct handout that
shows every fragment at once.

## Every number traces to the claims ledger

`research/<talk>/claims-ledger.md` holds every factual claim with a source and a
status. **A claim that is not `verified` must not appear on a slide.**

When you add a number, a quote, or an attribution to a deck, add its ledger row
in the same change. If no public source exists, either reframe the slide so the
claim is not made, or record it `pending` and flag it as blocking — do not ship
it with a warning attached and hope.

This applies especially to statements about an employer's internal
infrastructure.

## Verify before claiming, with evidence

Never say a slide renders, a check passes, or a fix works without having run it.

- `make check/fast` while iterating; `make check` before declaring done
- `make build/pdf/<talk>` and `python3 tools/check_slide_footer.py` — dense
  slides overflow into the footer band and the check catches it
- For visual changes, actually look: render the page to PNG (`pdftoppm -f N -l N
  -r 60 -png`) and inspect it. Text-level checks miss clipping and overflow
- When fixing a guard or tripwire, prove it fails on the broken input before
  proving it passes on the fixed one

## Assets

Imported assets go in `talks/<talk>/assets/`, referenced as `../assets/<file>`.

**Not `slides/assets/`** — `make sync/theme` rsyncs the theme there with
`--delete` and the path is gitignored, so anything you put there is destroyed by
the next build and never committed.

Every committed asset should be referenced by a slide. Generated charts (large
matplotlib SVGs) are marked `-diff` in `.gitattributes` so they do not make
diffs unreadable.

## Working rhythm

**Keep a live task list** and update it as work completes, not at the end. Long
restructures lose their thread otherwise.

**Commit after each meaningful step**, not in one batch at the end. Each commit
should build and pass checks on its own.

**Commit messages explain why.** The body carries the reasoning, the evidence,
and what was rejected. Never list files or functions — the diff already does.

**Report findings that contradict the plan.** Stale docs, wrong estimates, and
latent hazards found along the way are part of the deliverable. Say so plainly
rather than quietly working around them.

**Ask before destructive actions**, including deleting files outside the current
task's scope — even when they look like disposable state, and even when a
scanner is complaining about them.

## Worktrees

Work happens in a linked worktree under `.worktrees/<name>` on a branch of the
same name, never directly on `main`. `plans/` is gitignored and local-only.

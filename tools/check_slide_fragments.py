#!/usr/bin/env python3
"""Fail when a deck loses its Marp progressive-reveal fragment bullets.

Marp's bespoke template turns loose list items written with a `*` marker into
progressive-reveal fragments. That makes `*` load-bearing markup rather than a
style choice.

Prettier's markdown printer normalizes top-level unordered list markers to `-`.
A single `prettier --write` over a deck therefore converts every fragment bullet
into a plain bullet and silently removes every reveal from the talk. The deck is
listed in `.prettierignore` to prevent that; this check is the tripwire for when
the ignore is bypassed, removed, or an editor-integrated formatter runs anyway.

A deck opts in by declaring the number of fragment bullets it expects to keep,
so an accidental partial conversion is caught as well as a total one.
"""

from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass
from pathlib import Path

FRAGMENT_RE = re.compile(r"^\* \S")
FENCE_RE = re.compile(r"^\s*```")
# Decks declare their floor inline so this check needs no separate manifest.
# A YAML frontmatter comment is preferred over an HTML comment because Marp
# renders stray HTML comments as presenter notes, which would put the marker
# on screen in the speaker view during the talk.
DECLARED_RE = re.compile(
    r"(?:<!--|#)\s*fragment-floor:\s*(?P<count>\d{1,4})\s*(?:-->)?",
    re.IGNORECASE,
)


@dataclass(frozen=True)
class Deck:
    """A parsed deck: fragment count and its declared floor."""

    path: Path
    fragments: int
    floor: int | None


def count_fragments(text: str) -> int:
    """Count fragment bullets outside fenced code blocks."""
    total = 0
    in_fence = False
    for line in text.splitlines():
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if FRAGMENT_RE.match(line):
            total += 1
    return total


def declared_floor(text: str) -> int | None:
    """Read the deck's declared minimum fragment count, if it has one.

    Returns None when the marker is absent or unparseable. The caller treats
    both cases as "undeclared", so a malformed marker degrades to the softer
    zero-fragment check rather than crashing the build.
    """
    match = DECLARED_RE.search(text)
    if match is None:
        return None
    try:
        return int(match.group("count"))
    except ValueError:
        return None


def parse(path: Path) -> Deck:
    """Read and parse one deck, surfacing IO problems as a clear error."""
    try:
        text = path.read_text(encoding="utf-8")
    except OSError as err:
        raise SystemExit(f"{path}: cannot read deck: {err}") from err
    except UnicodeDecodeError as err:
        raise SystemExit(f"{path}: deck is not valid UTF-8: {err}") from err

    return Deck(path=path, fragments=count_fragments(text), floor=declared_floor(text))


def problem(deck: Deck) -> str | None:
    """Return an error message, or None when the deck is healthy."""
    if deck.floor is None:
        if deck.fragments == 0:
            return (
                f"{deck.path}: no `* ` fragment bullets and no `fragment-floor` marker.\n"
                "  If this deck intentionally has no reveals, add:\n"
                "    <!-- fragment-floor: 0 -->\n"
                "  Otherwise prettier probably rewrote `*` to `-`; see .prettierignore."
            )
        return None

    if deck.fragments < deck.floor:
        return (
            f"{deck.path}: {deck.fragments} fragment bullets, "
            f"expected at least {deck.floor}.\n"
            "  Progressive reveals have been lost. Most likely cause: prettier\n"
            "  rewrote `*` list markers to `-`. Check .prettierignore, then\n"
            "  restore the bullets. If the reduction is intentional, lower the\n"
            "  `fragment-floor` marker in the deck."
        )
    return None


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("deck", nargs="+", type=Path)
    args = parser.parse_args()

    failures = False
    for path in args.deck:
        if not path.is_file():
            parser.error(f"missing deck: {path}")
        deck = parse(path)
        issue = problem(deck)
        if issue is None:
            print(f"{deck.path}: {deck.fragments} fragment bullets")
            continue
        print(issue, file=sys.stderr)
        failures = True

    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main())

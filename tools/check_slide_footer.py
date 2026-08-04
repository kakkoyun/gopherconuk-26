#!/usr/bin/env python3
"""Fail when slide text enters the shared footer and pagination area."""

from __future__ import annotations

import argparse
import html
import re
import shutil
import subprocess
import sys
from pathlib import Path

PAGE_RE = re.compile(r'<page width="[^"]+" height="([0-9.]+)">')
WORD_RE = re.compile(
    r'<word xMin="[^"]+" yMin="([0-9.]+)" xMax="[^"]+" yMax="[^"]+">(.*?)</word>'
)


def coordinate(value: str) -> float | None:
    """Parse a Poppler coordinate without trusting malformed output."""
    try:
        return float(value)
    except ValueError:
        return None


def footer_overflow(pdf: Path) -> dict[int, list[str]]:
    """Return unexpected words inside the bottom 55 points of each page."""
    result = subprocess.run(
        ["pdftotext", "-bbox-layout", str(pdf), "-"],
        check=True,
        capture_output=True,
        text=True,
    )

    page = 0
    footer_start = 0.0
    overflows: dict[int, list[str]] = {}
    for line in result.stdout.splitlines():
        page_match = PAGE_RE.search(line)
        if page_match:
            page_height = coordinate(page_match.group(1))
            if page_height is None:
                continue
            page += 1
            # ponytail: footer-zone heuristic; use pixel-level checks if layout geometry changes.
            footer_start = page_height - 55
            continue

        word_match = WORD_RE.search(line)
        if not word_match:
            continue
        ymin = coordinate(word_match.group(1))
        if ymin is None or ymin < footer_start:
            continue

        word: str = html.unescape(re.sub(r"<.*?>", "", word_match.group(2)))
        if (
            word == "Kemal"
            or word == "Akkoyun"
            or word == "·"
            or word == "Datadog"
            or word == "/"
            or word.isdecimal()
        ):
            continue
        overflows.setdefault(page, []).append(word)

    return overflows


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("pdf", nargs="+", type=Path)
    args = parser.parse_args()

    if shutil.which("pdftotext") is None:
        parser.error("pdftotext is required (install poppler-utils)")

    failures = False
    for pdf in args.pdf:
        if not pdf.is_file():
            parser.error(f"missing PDF: {pdf}")
        for page, words in footer_overflow(pdf).items():
            print(
                f"{pdf}: page {page}: footer overlap: {' '.join(words)}",
                file=sys.stderr,
            )
            failures = True

    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3
"""Lint code-fence headers and languages in a Marp deck.

Four checks, run on every fenced code block:

1. Kind lint — the fence language is in the known set (the theme's kinds plus
   the kroki diagram languages) or empty. An unknown language fails, so adding
   one is a deliberate edit to both the theme kind map and this list.

2. Header-expected lint — a source fence (``go``, ``yaml``, ``yml``, ``s``) or
   an ``asm`` fence with no ``######`` header and no ``<!-- code-header: none -->``
   opt-out fails. These panels name a file you would open; a bare fence hides
   provenance.

3. Path check — a header that is not a known external source resolves to exactly
   one file by basename under the repo (excluding ``node_modules``, ``.git``,
   and build output). Zero matches fails; more than one fails asking for
   disambiguation, which is what keeps bare filenames honest. Skipped with
   ``--no-path-check`` (used by the theme fixture, whose filenames are invented).

4. Cross-repo ledger check — a header naming an external source must appear in
   that talk's ``research/<talk>/claims-ledger.md``. This converts the TODO's
   silent-staleness risk for cross-repo evidence into a failing check.
"""

from __future__ import annotations

import argparse
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path

# ── known fence languages ──────────────────────────────────────────────

# Kinds the theme styles with a coloured rail. ``text`` is both ASCII diagrams
# and captured output; the opt-in header is the discriminator.
THEME_KINDS = {
    "go",
    "yaml",
    "yml",
    "s",
    "asm",
    "bash",
    "sh",
    "shell",
    "zsh",
    "console",
    "text",
}

# Diagram languages the kroki Marp plugin renders as images. Kept in sync with
# talks/<talk>/slides/marp/kroki-plugin.js.
KROKI_KINDS = {
    "actdiag",
    "blockdiag",
    "bpmn",
    "bytefield",
    "c4plantuml",
    "ditaa",
    "dot",
    "erd",
    "excalidraw",
    "graphviz",
    "mermaid",
    "nomnoml",
    "nwdiag",
    "packetdiag",
    "pikchr",
    "plantuml",
    "rackdiag",
    "seqdiag",
    "svgbob",
    "umlet",
    "vega",
    "vegalite",
    "wavedrom",
}

KNOWN_LANGS = THEME_KINDS | KROKI_KINDS

# Source kinds (and disassembly) that must name their file or opt out.
HEADER_REQUIRED = {"go", "yaml", "yml", "s", "asm"}

# Headers that name an external source, not a file in this repo. These skip the
# path check and instead require a row in the talk's claims ledger. Adding one
# is a deliberate edit here.
EXTERNAL_SOURCES = {
    "benchlab",
    "dd-trace-go",
    "runtime/proc.go",
    "gls.orchestrion.yml",
}

# Directories and suffixes excluded from the basename path check.
PATH_EXCLUDE_DIRS = {"node_modules", ".git", ".worktrees", ".claude", "assets", "fonts"}
PATH_EXCLUDE_SUFFIXES = {".html", ".pdf"}


# ── parsing ────────────────────────────────────────────────────────────

FENCE_OPEN_RE = re.compile(r"^```(\S*)")
FENCE_CLOSE_RE = re.compile(r"^\s*```")
H6_RE = re.compile(r"^######\s+(.*)")
OPTOUT_RE = re.compile(r"^<!--\s*code-header:\s*none\s*-->", re.IGNORECASE)
DIRECTIVE_RE = re.compile(r"^<!--\s*_")  # Marp directive: <!-- _class: ... -->
FRAGMENT_FLOOR_RE = re.compile(r"^(?:<!--\s*)?#\s*fragment-floor", re.IGNORECASE)


@dataclass(frozen=True)
class Panel:
    """One fenced code block and the header context above it."""

    path: Path
    line: int  # 1-based line of the opening fence
    lang: str  # info string, "" when absent
    header: str | None  # h6 text immediately above, or None
    opt_out: bool  # <!-- code-header: none --> above the fence


def parse_deck(path: Path) -> list[Panel]:
    """Return every fenced code block with its header / opt-out context."""
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError as err:
        raise SystemExit(f"{path}: cannot read deck: {err}") from err
    except UnicodeDecodeError as err:
        raise SystemExit(f"{path}: deck is not valid UTF-8: {err}") from err

    panels: list[Panel] = []
    in_fence = False
    for i, line in enumerate(lines):
        if in_fence:
            if FENCE_CLOSE_RE.match(line):
                in_fence = False
            continue
        m = FENCE_OPEN_RE.match(line)
        if not m:
            continue
        lang = m.group(1)
        header, opt_out = _header_context(lines, i)
        panels.append(Panel(path, i + 1, lang, header, opt_out))
        in_fence = True
    return panels


def _header_context(lines: list[str], fence_idx: int) -> tuple[str | None, bool]:
    """Find the h6 header or opt-out comment above a fence.

    Scans backwards past blank lines and Marp directives; the first content
    line is either the ``######`` header, the opt-out comment, or something
    else (a title, body text) which means no header.
    """
    j = fence_idx - 1
    while j >= 0:
        stripped = lines[j].strip()
        if not stripped:
            j -= 1
            continue
        if DIRECTIVE_RE.match(stripped) or FRAGMENT_FLOOR_RE.match(stripped):
            j -= 1
            continue
        if OPTOUT_RE.match(stripped):
            return None, True
        m = H6_RE.match(stripped)
        if m:
            return m.group(1).strip(), False
        return None, False
    return None, False


# ── path check ──────────────────────────────────────────────────────────

def _basename_of(header: str) -> str:
    """Last path component of a header — ``runtime/proc.go`` → ``proc.go``."""
    return header.rsplit("/", 1)[-1]


def _iter_repo_files(repo_root: Path):
    """Yield repo files, excluding build output and vendored trees."""
    for p in repo_root.rglob("*"):
        if not p.is_file():
            continue
        # Exclude by directory name relative to the repo root, so the root's
        # own path components (e.g. running from inside .worktrees/) do not
        # trigger an exclusion.
        try:
            rel_parts = set(p.relative_to(repo_root).parts)
        except ValueError:
            continue
        if rel_parts & PATH_EXCLUDE_DIRS:
            continue
        if p.suffix in PATH_EXCLUDE_SUFFIXES:
            continue
        yield p


@dataclass
class RepoIndex:
    """Basename → repo-relative paths, built once per repo."""

    by_basename: dict[str, list[Path]] = field(default_factory=dict)

    @classmethod
    def build(cls, repo_root: Path) -> RepoIndex:
        idx = cls()
        for p in _iter_repo_files(repo_root):
            idx.by_basename.setdefault(p.name, []).append(p)
        return idx

    def resolve(self, header: str) -> list[Path]:
        return self.by_basename.get(_basename_of(header), [])


# ── ledger check ────────────────────────────────────────────────────────

def _talk_of(deck: Path) -> str | None:
    """Derive the talk name from a deck path: talks/<talk>/slides/... → <talk>."""
    parts = deck.parts
    if "talks" in parts:
        i = parts.index("talks")
        if i + 1 < len(parts):
            return parts[i + 1]
    return None


def _ledger_path(repo_root: Path, deck: Path) -> Path | None:
    talk = _talk_of(deck)
    if talk is None:
        return None
    return repo_root / "research" / talk / "claims-ledger.md"


def ledger_has_source(ledger: Path, source: str) -> bool:
    """True when the external source string appears in the claims ledger."""
    try:
        text = ledger.read_text(encoding="utf-8")
    except OSError:
        return False
    return source in text


# ── checks ─────────────────────────────────────────────────────────────

def lint(deck: Path, repo_root: Path, no_path_check: bool) -> list[str]:
    """Run all four checks on one deck; return a list of error messages."""
    panels = parse_deck(deck)
    errors: list[str] = []
    repo_idx = None if no_path_check else RepoIndex.build(repo_root)
    ledger = _ledger_path(repo_root, deck)

    for panel in panels:
        loc = f"{deck}:{panel.line}"

        # 1. kind lint
        if panel.lang and panel.lang not in KNOWN_LANGS:
            errors.append(
                f"{loc}: unknown fence language `{panel.lang}`. Add it to the "
                f"theme kind map and KNOWN_LANGS in check_code_headers.py."
            )

        # 2. header-expected lint
        if panel.lang in HEADER_REQUIRED and panel.header is None and not panel.opt_out:
            errors.append(
                f"{loc}: `{panel.lang}` fence has no code header. Add a "
                f"###### filename tab, or <!-- code-header: none --> to opt out."
            )

        if panel.header is None:
            continue

        header = panel.header

        # 3/4. external source → ledger check, else path check
        if header in EXTERNAL_SOURCES:
            if ledger is None or not ledger.is_file():
                errors.append(
                    f"{loc}: external source `{header}` needs a row in a "
                    f"claims ledger, but no ledger was found for this deck."
                )
            elif not ledger_has_source(ledger, header):
                errors.append(
                    f"{loc}: external source `{header}` has no row in {ledger}. "
                    f"Add a verified row naming `{header}`."
                )
            continue

        if no_path_check or repo_idx is None:
            continue

        matches = repo_idx.resolve(header)
        if len(matches) == 0:
            errors.append(
                f"{loc}: header `{header}` matches no file in the repo by "
                f"basename. If it is an external source, add it to "
                f"EXTERNAL_SOURCES and the claims ledger."
            )
        elif len(matches) > 1:
            joined = ", ".join(str(m.relative_to(repo_root)) for m in matches)
            errors.append(
                f"{loc}: header `{header}` matches {len(matches)} files by "
                f"basename ({joined}). Use a more specific name."
            )
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("deck", nargs="+", type=Path, help="deck markdown file(s)")
    parser.add_argument(
        "--no-path-check",
        action="store_true",
        help="skip the basename path check (for fixtures with invented filenames)",
    )
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=Path.cwd(),
        help="repo root for the basename search and ledger lookup (default: cwd)",
    )
    args = parser.parse_args()

    failures = False
    for deck in args.deck:
        if not deck.is_file():
            parser.error(f"missing deck: {deck}")
        for err in lint(deck, args.repo_root, args.no_path_check):
            print(err, file=sys.stderr)
            failures = True
        if not failures:
            print(f"{deck}: code headers ok")
    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main())

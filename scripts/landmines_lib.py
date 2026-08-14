"""Shared parsing for LANDMINES.md — the ONE reader, imported by both consumers.

`landmines-sync.py` (markdown -> doc_notes) and `landmines-session-start.py`
(the SessionStart hook) both need to know what an entry is and what its footprints
are. Two copies of that logic is the drift shape this repo keeps paying for — two
hand-maintained council rosters is why 099_SYNC_gate_roster.py had to exist — so
the parser lives here once and both import it.

Owner ruling D10 (2026-07-29): the markdown is the system of record. Everything in
this module reads it; nothing here writes it.
"""

import hashlib
import os
import re

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
LANDMINES = os.path.join(
    REPO, "docs", "agent_docs", "docs024_key_docs_latest", "LANDMINES.md"
)

# Entries live under "# Entries"; the preamble above it contains ### headings of
# its own and must never be parsed as landmines.
ENTRIES_MARKER = "# Entries"

FIELD_RE = re.compile(r"^-\s+\*\*(?P<label>[a-z ]+):\*\*\s*(?P<value>.*)$")
# '###' is the documented entry heading. '##' is ACCEPTED here on purpose, and
# warned about — not rejected, and not ignored. When this matched '###' only, a
# '##' heading was not a heading at all: its lines were absorbed into the
# PRECEDING entry and its '**footprint:**' line overwrote that entry's own, so one
# malformed heading silently cost TWO entries their delivery. Measured 2026-08-02:
# `UpsertPageForRole` had 0 doc_notes rows because a '##' entry appended hours
# after it swallowed it. Nothing errored, and no count looked wrong.
HEADING_RE = re.compile(r"^(?P<hashes>#{2,3})\s+(?P<title>.+?)\s*$")
# A footprint opening with a determiner is a description, not a grep target.
PROSE_FOOTPRINT_RE = re.compile(r"^(the|any|a|an|every|all|some)\s", re.I)


def slugify(title):
    """Stable id for an entry, from the TITLE only.

    Deliberately not derived from the body: editing an entry must UPDATE its row,
    not orphan the old one and insert a new one. A retitled entry is treated as a
    new landmine, which is the honest reading.
    """
    s = re.sub(r"`|\*|_", "", title.lower())
    s = re.sub(r"[^a-z0-9]+", "-", s).strip("-")
    return s[:80] or hashlib.sha1(title.encode()).hexdigest()[:12]


def split_footprints(value):
    """'`cmd/`, `scripts/x.py`' -> ['cmd/', 'scripts/x.py']

    Three separators, three rules (2026-08-14 — before that, only the first
    existed, and the file's own dominant '·' convention silently collapsed
    multi-part footprints into ONE unmatchable string on 63 of 482 entries;
    comma-splitting inside parentheticals had mangled 143 more):
    - ',' / ';' split OUTSIDE backticks and OUTSIDE parentheses. The paren rule
      keeps '`x.go (FuncA, FuncB)`' as one footprint — the trailing-paren strip
      below then removes the qualifier, exactly as it always did for the
      comma-free case.
    - '·' splits UNCONDITIONALLY: it is this file's separator convention and can
      never be part of a real path/table/symbol, while unbalanced backticks in
      the wild make any in-tick state unreliable (12 middots sat inside apparent
      backtick spans when this was measured).
    Then strips EVERY backtick — a stray backtick inside a subject_key makes the
    row unfindable by the exact string a later session would search for, which
    defeats the whole point of a footprint.
    """
    parts = []
    for chunk in value.split("·"):
        buf, in_tick, depth = [], False, 0
        for ch in chunk:
            if ch == "`":
                in_tick = not in_tick
                buf.append(ch)
            elif ch == "(":
                depth += 1
                buf.append(ch)
            elif ch == ")":
                depth = max(0, depth - 1)
                buf.append(ch)
            elif ch in ",;" and not in_tick and depth == 0:
                parts.append("".join(buf))
                buf = []
            else:
                buf.append(ch)
        parts.append("".join(buf))

    out, seen = [], set()
    for p in parts:
        # Backticks off FIRST (2026-08-14) — the qualifier strip below used to run
        # before this, so any `path (Qualifier)` wrapped in backticks ended in a
        # backtick, never matched the $-anchored regex, and kept its qualifier.
        p = p.replace("`", "").strip()
        # Trailing ' (qualifier)' strips; 'name(signature)' is KEPT — the space is
        # the discriminator (a SQL signature like snapshot_agent(text, text) has
        # none, and is_prose's own docstring records it as a correct footprint).
        p = re.sub(r"\s+\(.*?\)\s*$", "", p).strip()
        p = p.strip(",;").strip()
        if not p or p.lower() in seen:
            continue                      # one entry listing a footprint twice wrote two identical rows
        seen.add(p.lower())
        out.append(p)
    return out


def is_prose(footprint):
    """True when a footprint reads as description rather than a grep target.

    Not merely 'contains a space': `git commit`, `git checkout HEAD --` and
    `snapshot_agent(text, text)` all do and are all exactly right. The first
    version of this test flagged 7 of 14 entries, most of them correct — and a
    check that over-reports gets ignored, which is worse than no check.
    """
    return bool(PROSE_FOOTPRINT_RE.match(footprint)) or len(footprint.split()) > 4


def parse(path=None, on_warn=None):
    """LANDMINES.md -> [{title, slug, footprints, body}]"""
    path = path or LANDMINES
    with open(path, encoding="utf-8") as fh:
        text = fh.read()

    if ENTRIES_MARKER not in text:
        raise ValueError(
            f"{path}: no '{ENTRIES_MARKER}' marker — refusing to guess where entries "
            "begin. Parsing the preamble as landmines is the worse failure."
        )
    body_text = text.split(ENTRIES_MARKER, 1)[1]

    entries, cur, open_field = [], None, None
    for line in body_text.splitlines():
        m = HEADING_RE.match(line)
        if m:
            if cur:
                entries.append(cur)
            cur = {"title": m.group("title"), "lines": [], "fields": {}}
            open_field = None
            if m.group("hashes") == "##" and on_warn:
                # Not fatal: the entry parses correctly from here. But a '##' that
                # carries no footprint is a section divider and will be skipped
                # below, so say which one this is rather than guessing.
                on_warn(
                    f"heading uses '##', the format is '###': {m.group('title')[:70]}"
                )
            continue
        if cur is None:
            continue
        cur["lines"].append(line)

        fm = FIELD_RE.match(line)
        if fm:
            open_field = fm.group("label").strip()
            cur["fields"][open_field] = fm.group("value").strip()
            continue

        # A bullet that wraps onto indented continuation lines is ONE value. Taking
        # only the first line silently truncates it — and a truncated instruction
        # reads exactly like a complete one, which is how the hook first printed
        # "a path recalled from memory is an" as if that were the whole check.
        if open_field and line.startswith(("  ", "\t")) and line.strip():
            cur["fields"][open_field] += " " + line.strip()
        elif not line.strip():
            open_field = None
    if cur:
        entries.append(cur)

    out = []
    for e in entries:
        fp = e["fields"].get("footprint", "")
        if not fp:
            if on_warn:
                on_warn(f"skipped (no footprint): {e['title'][:70]}")
            continue
        footprints = split_footprints(fp)
        if on_warn:
            for f in footprints:
                if is_prose(f):
                    on_warn(
                        f"prose footprint {f!r}\n      in: {e['title'][:64]}\n"
                        "      -> subject_key is what a later session greps; prefer "
                        "the literal path/table/symbol"
                    )
        body = "\n".join(e["lines"]).strip()
        out.append(
            {
                "title": e["title"],
                "slug": slugify(e["title"]),
                "footprints": footprints,
                "the_check": e["fields"].get("the check", ""),
                "body": f"{e['title']}\n\n{body}",
            }
        )
    return out


def _self_test():
    """Splitter cases, each naming the wrong answer it guards against."""
    cases = [
        # (input, expected) — the pre-2026-08-14 splitter fails the '·' and paren cases
        ("`cmd/`, `scripts/x.py`", ["cmd/", "scripts/x.py"]),
        ("`a.go` · `b_table` · `some command`", ["a.go", "b_table", "some command"]),
        ("`deployments/kustomize/` · `make deploy-` · `git stash`",
         ["deployments/kustomize/", "make deploy-", "git stash"]),
        # comma inside a parenthetical must NOT split; trailing paren then strips
        ("`x.go (FuncA, FuncB)`", ["x.go"]),
        ("`x.go (FuncA, FuncB)` · `y_table`", ["x.go", "y_table"]),
        # a SIGNATURE paren (no space before it) is part of the key and is KEPT
        ("`snapshot_agent(text, text)`", ["snapshot_agent(text, text)"]),
        # '·' inside a backtick span still splits — unbalanced ticks are common
        ("`a.go · b.go`", ["a.go", "b.go"]),
        # dedupe and empties (pre-existing behaviour, kept)
        ("`a.go`, `a.go`, ,", ["a.go"]),
    ]
    failed = 0
    for value, want in cases:
        got = split_footprints(value)
        ok = got == want
        failed += 0 if ok else 1
        print(f"  {'PASS' if ok else 'FAIL'}  {value!r} -> {got!r}"
              + ("" if ok else f"  (want {want!r})"))
    print(f"{len(cases) - failed}/{len(cases)} passed")
    return 1 if failed else 0


if __name__ == "__main__":
    import sys
    sys.exit(_self_test())

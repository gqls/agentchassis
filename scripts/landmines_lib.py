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
HEADING_RE = re.compile(r"^###\s+(?P<title>.+?)\s*$")
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

    Splits on commas/semicolons OUTSIDE backticks, then strips EVERY backtick —
    entries in the wild carry unbalanced formatting, and a stray backtick inside a
    subject_key makes the row unfindable by the exact string a later session would
    search for, which defeats the whole point of a footprint.
    """
    parts, buf, in_tick = [], [], False
    for ch in value:
        if ch == "`":
            in_tick = not in_tick
            buf.append(ch)
        elif ch in ",;" and not in_tick:
            parts.append("".join(buf))
            buf = []
        else:
            buf.append(ch)
    parts.append("".join(buf))

    out, seen = [], set()
    for p in parts:
        p = re.sub(r"\s*\(.*?\)\s*$", "", p).strip()
        p = p.replace("`", "").strip().strip(",;").strip()
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

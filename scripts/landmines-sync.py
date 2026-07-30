#!/usr/bin/env python3
"""landmines-sync.py — push LANDMINES.md into doc_notes so machines can read it.

Owner ruling D10 (2026-07-29): the MARKDOWN is the system of record. This script
is one-way, markdown -> rows. It never reads a row back into the file, and it
never edits LANDMINES.md.

    ./scripts/landmines-sync.py              # dry run: what would change
    ./scripts/landmines-sync.py --apply      # write to doc_notes
    ./scripts/landmines-sync.py --check      # exit 1 if out of sync (for CI/hooks)

WHY MARKDOWN-FIRST, given the proposal argued rows-first: a markdown append costs
ten seconds, needs no cluster, and works in a fresh clone or with an expired
kubeconfig (which this platform hits every 3 days). An unadopted ledger is the
failure D10 exists to fix, so authoring friction was the deciding cost.

THE ROWS THIS OWNS, AND THE ONES IT MUST NOT TOUCH
  It owns rows whose `source` begins 'LANDMINES.md#'. Those are regenerated.
  It does NOT touch the 7 landmine rows written 2026-07-27/28 by other threads
  (source: 'architecture_review workstream', 'bugs_open/100 + ...'), which are
  subject_type='action' and predate any decision. Deleting another thread's
  records to make this script's output tidy is not a sync, it is a data loss.

CONSUMERS: query `categories ? 'landmine'`, NOT subject_type. The category spans
both conventions; subject_type='landmine' only exists on rows this script writes
(and only after migration 270).
"""

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
LANDMINES = os.path.join(
    REPO, "docs", "agent_docs", "docs024_key_docs_latest", "LANDMINES.md"
)
SOURCE_PREFIX = "LANDMINES.md#"

PSQL = os.environ.get("PSQL_CMD") or (
    "kubectl exec -i -n ai-persona-system postgres-clients-0 -- "
    "psql -U clients_user -d clients_db"
)

# Bullet labels the entry format defines. `footprint` is the load-bearing one —
# it becomes subject_key, which is what makes a row findable by what it guards.
FIELD_RE = re.compile(r"^-\s+\*\*(?P<label>[a-z ]+):\*\*\s*(?P<value>.*)$")
HEADING_RE = re.compile(r"^###\s+(?P<title>.+?)\s*$")
# Entries live under "# Entries"; everything above it is the file's own preamble,
# which contains ### headings too and must not be parsed as landmines.
ENTRIES_MARKER = "# Entries"
# A footprint opening with a determiner is a description, not a grep target.
PROSE_FOOTPRINT_RE = re.compile(r"^(the|any|a|an|every|all|some)\s", re.I)


def slugify(title):
    """Stable id for an entry, used as the `source` suffix.

    Deliberately derived from the TITLE only, not the body: editing an entry's
    text must UPDATE its row, not orphan the old one and insert a new one. A
    retitled entry is treated as a new landmine, which is the honest reading.
    """
    s = re.sub(r"`|\*|_", "", title.lower())
    s = re.sub(r"[^a-z0-9]+", "-", s).strip("-")
    return s[:80] or hashlib.sha1(title.encode()).hexdigest()[:12]


def split_footprints(value):
    """`cmd/`, `scripts/x.py`, the Bash tool -> ['cmd/', 'scripts/x.py', ...]

    Split on commas OUTSIDE backticks, so a footprint containing a comma inside
    code formatting survives. Bare prose footprints (e.g. "the Bash tool itself")
    are kept: a landmine about a tool with no path still needs somewhere to live.
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
        p = re.sub(r"\s*\(.*?\)\s*$", "", p).strip()   # drop trailing asides
        # Strip EVERY backtick, not just the outer pair: entries in the wild carry
        # unbalanced formatting (`site_work_items` where `item_type=...`), and a
        # stray backtick inside subject_key makes the row unfindable by the exact
        # string a later session would search for — which is the whole point of it.
        p = p.replace("`", "").strip().strip(",;").strip()
        if not p or p.lower() in seen:
            continue          # dedupe: one entry listing a footprint twice wrote two identical rows
        seen.add(p.lower())
        out.append(p)
    return out


def parse(path):
    """LANDMINES.md -> [{title, slug, footprints, body}]"""
    with open(path, encoding="utf-8") as fh:
        text = fh.read()

    if ENTRIES_MARKER not in text:
        sys.exit(
            f"{path}: no '{ENTRIES_MARKER}' marker — refusing to guess where entries "
            "begin. Adding one is the fix; parsing the preamble as landmines is not."
        )
    body_text = text.split(ENTRIES_MARKER, 1)[1]

    entries, cur = [], None
    for line in body_text.splitlines():
        m = HEADING_RE.match(line)
        if m:
            if cur:
                entries.append(cur)
            cur = {"title": m.group("title"), "lines": [], "fields": {}}
            continue
        if cur is None:
            continue
        cur["lines"].append(line)
        fm = FIELD_RE.match(line)
        if fm:
            cur["fields"][fm.group("label").strip()] = fm.group("value").strip()
    if cur:
        entries.append(cur)

    out = []
    for e in entries:
        fp = e["fields"].get("footprint", "")
        if not fp:
            print(
                f"  ! skipped (no footprint): {e['title'][:70]}",
                file=sys.stderr,
            )
            continue
        body = "\n".join(e["lines"]).strip()
        footprints = split_footprints(fp)

        # A footprint should be the literal string a later session would grep for.
        # Flag PROSE only — a determiner ("the chassis LLM call layer", "any
        # sql_for_agents/*.sql in the house") or a sentence-length run. NOT merely
        # "contains a space": `git commit`, `git checkout HEAD --` and
        # `snapshot_agent(text, text)` all do, and all are exactly right.
        #
        # The first version of this check flagged every space and fired on 7 of 14
        # entries, most of them correct. A check that over-reports gets ignored, and
        # then it is worse than no check — the same lesson render_audit.py's three
        # false positives taught.
        for f in footprints:
            if PROSE_FOOTPRINT_RE.match(f) or len(f.split()) > 4:
                print(
                    f"  ! prose footprint {f!r}\n"
                    f"      in: {e['title'][:64]}\n"
                    "      -> subject_key is what a later session greps; prefer the "
                    "literal path/table/symbol",
                    file=sys.stderr,
                )

        out.append(
            {
                "title": e["title"],
                "slug": slugify(e["title"]),
                "footprints": footprints,
                "body": f"{e['title']}\n\n{body}",
            }
        )
    return out


def sql_lit(s):
    return "'" + s.replace("'", "''") + "'"


def run_psql(sql, capture=True):
    cmd = PSQL.split()
    proc = subprocess.run(
        cmd + ["-v", "ON_ERROR_STOP=1", "-t", "-A", "-c", sql],
        capture_output=capture,
        text=True,
    )
    if proc.returncode != 0:
        sys.exit(f"psql failed:\n{proc.stderr or proc.stdout}")
    return (proc.stdout or "").strip()


def existing_sources():
    out = run_psql(
        "SELECT source, count(*) FROM doc_notes "
        f"WHERE source LIKE {sql_lit(SOURCE_PREFIX + '%')} GROUP BY 1;"
    )
    got = {}
    for line in out.splitlines():
        if "|" in line:
            src, n = line.rsplit("|", 1)
            got[src.strip()] = int(n)
    return got


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--apply", action="store_true", help="write to doc_notes")
    ap.add_argument(
        "--check",
        action="store_true",
        help="exit 1 if the DB does not match the file (no writes)",
    )
    args = ap.parse_args()

    entries = parse(LANDMINES)
    if not entries:
        sys.exit("no entries parsed — refusing to run (this would delete every owned row)")

    want = {}
    for e in entries:
        want[SOURCE_PREFIX + e["slug"]] = e

    print(f"{len(entries)} entr{'y' if len(entries)==1 else 'ies'} in LANDMINES.md")
    for e in entries:
        print(f"  {e['slug']}  -> {len(e['footprints'])} footprint(s): "
              f"{', '.join(e['footprints'][:4])}")

    have = existing_sources()
    new = [s for s in want if s not in have]
    gone = [s for s in have if s not in want]
    print(f"\ndoc_notes: {sum(have.values())} owned row(s) across {len(have)} entr(ies)")
    print(f"  to insert/refresh: {len(want)}   orphaned (entry retitled/removed): {len(gone)}")
    for s in gone:
        print(f"    orphan: {s}")

    if args.check:
        drift = new or gone
        print("\nOUT OF SYNC — run with --apply" if drift else "\nin sync")
        return 1 if drift else 0

    if not args.apply:
        print("\nDry run. Re-run with --apply to write.")
        print("Note: --apply REPLACES rows whose source starts "
              f"'{SOURCE_PREFIX}'. It never touches landmine rows written by other "
              "threads (different `source`).")
        return 0

    stmts = ["BEGIN;"]
    # Replace-in-place: delete only what this script owns, then reinsert. Scoped by
    # `source`, so another thread's landmine rows are structurally out of reach.
    stmts.append(
        f"DELETE FROM doc_notes WHERE source LIKE {sql_lit(SOURCE_PREFIX + '%')};"
    )
    for src, e in want.items():
        cats = json.dumps(["landmine", "synced-from-markdown"])
        for fp in e["footprints"]:
            stmts.append(
                "INSERT INTO doc_notes (subject_type, subject_key, body, categories, "
                "source, created_by) VALUES ("
                f"'landmine', {sql_lit(fp)}, {sql_lit(e['body'])}, "
                f"{sql_lit(cats)}::jsonb, {sql_lit(src)}, 'landmines-sync');"
            )
    stmts.append("COMMIT;")

    run_psql("\n".join(stmts), capture=True)
    after = existing_sources()
    print(f"\napplied: {sum(after.values())} owned row(s) now present")
    return 0


if __name__ == "__main__":
    sys.exit(main())

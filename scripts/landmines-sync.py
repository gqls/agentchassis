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

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from landmines_lib import LANDMINES, parse, slugify, split_footprints  # noqa: E402

SOURCE_PREFIX = "LANDMINES.md#"

PSQL = os.environ.get("PSQL_CMD") or (
    "kubectl exec -i -n ai-persona-system postgres-clients-0 -- "
    "psql -U clients_user -d clients_db"
)


def sql_lit(s):
    return "'" + s.replace("'", "''") + "'"


def run_psql(sql, capture=True):
    """Send SQL on STDIN, not as `-c`.

    `-c` put the whole statement list in argv, and on 2026-07-30 the corpus grew
    past the kernel's argument limit: --apply died with `OSError: [Errno 7]
    Argument list too long` mid-append, which is a sync that stops working purely
    because the file it syncs got bigger. Nothing about the SQL changes — the apply
    path already wraps itself in explicit BEGIN;/COMMIT;, so atomicity is exactly
    what it was; only the transport moves. ON_ERROR_STOP=1 still aborts the
    transaction on the first failure.
    """
    cmd = PSQL.split()
    proc = subprocess.run(
        cmd + ["-v", "ON_ERROR_STOP=1", "-t", "-A"],
        input=sql,
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


def existing_bodies():
    """{source: body} for whatever landmines-sync currently owns.

    Every footprint row of one entry shares an identical body (see the INSERT
    loop below), so DISTINCT ON (source) returns exactly one row per entry —
    the body as it stood BEFORE this run's delete+reinsert. Used only to
    detect a CONTENT edit under an unchanged slug: existing_sources()'s
    new/gone lists (keyed on source presence alone) cannot see that, and
    landmines-sync has never needed to until a consumer (landmine-verifier,
    RFC_005 3.2) needed to know WHICH entries changed, not just whether the
    set of slugs did.

    JSON, not line-split `-t -A` text like existing_sources() above: `body`
    is genuine multi-line prose (confirmed live — every entry's footprint
    line alone forces a newline), so a naive per-line `|`-split would silently
    truncate every multi-line body to its first line. existing_sources()
    gets away with plain text because `source` and `count(*)` can never
    contain a newline; `body` can and does.
    """
    out = run_psql(
        "SELECT COALESCE(jsonb_object_agg(source, body), '{}'::jsonb) FROM "
        "(SELECT DISTINCT ON (source) source, body FROM doc_notes "
        f"WHERE source LIKE {sql_lit(SOURCE_PREFIX + '%')} "
        "ORDER BY source, created_at DESC) t;"
    )
    return json.loads(out) if out else {}


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
    slug_titles = {}  # slug -> [title, title, ...] seen so far, to catch collisions below
    for e in entries:
        slug_titles.setdefault(e["slug"], []).append(e["title"])
        want[SOURCE_PREFIX + e["slug"]] = e  # last one in file order silently wins

    # slugify() is TITLE-only (by design — see landmines_lib.py), so two entries
    # with the same or near-identical title collide to one dict key and only
    # the LAST survives in `want` — the others are silently dropped from every
    # sync, forever, with no error anywhere. Found live 2026-07-31: a
    # "re-appended after the first copy was lost to a concurrent write" entry
    # collided with the very entry it was re-appending, leaving the original
    # permanently unsynced and invisible to every doc_notes consumer.
    collisions = {slug: titles for slug, titles in slug_titles.items() if len(titles) > 1}
    if collisions:
        print(f"!! {len(collisions)} slug collision(s) — only the LAST entry below survives in doc_notes, the rest are silently dropped:")
        for slug, titles in collisions.items():
            print(f"    {slug}:")
            for t in titles:
                print(f"      - {t}")

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

    # Content changed under an unchanged slug: existing_sources() above only
    # sees whether the SOURCE KEY is present, not whether its body drifted —
    # a hand-edit to an existing entry (typo fix, corrected footprint,
    # tightened "the check") is invisible to new/gone. Needed for
    # landmine-verifier (RFC_005 3.2): a changed entry is exactly as much in
    # need of re-verification as a brand new one, and silently was not
    # getting it before this.
    have_bodies = existing_bodies()
    changed = [s for s in want if s in have and have_bodies.get(s) != want[s]["body"]]
    if changed:
        print(f"  content changed (same slug, edited body): {len(changed)}")
        for s in changed:
            print(f"    changed: {s}")

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

    # Machine-parseable, on their own lines, for landmines-verify-dispatch.sh
    # (RFC_005 3.2) to grep and act on — new or content-changed entries need a
    # fresh landmine-verifier pass; unchanged ones do not.
    needs_verification = new + changed
    if needs_verification:
        print(f"\n{len(needs_verification)} entr{'y' if len(needs_verification)==1 else 'ies'} need verification:")
        for s in needs_verification:
            print(f"NEEDS_VERIFICATION:{s}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

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

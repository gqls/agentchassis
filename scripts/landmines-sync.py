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

    BYTES, not `text=True` (2026-08-12). The same growth curve bit a second time,
    one layer down: at ~2,155 owned rows `--apply` died with
    `UnicodeDecodeError: 'utf-8' codec can't decode byte 0xe2 ... unexpected end
    of data` — a multibyte character (0xe2 = the leading byte of our em-dashes)
    left truncated at the end of the captured stream by the `kubectl exec`
    transport. `text=True` decodes inside subprocess, so this raised BEFORE the
    returncode check and the real psql message was never printed: the sync simply
    crashed, and it crashed for no reason other than the corpus getting bigger.
    Same failure shape as the 2026-07-30 argv-limit note above, same lesson.

    So: capture bytes and decode with errors="replace". A truncated tail can now
    at worst produce a replacement character in a message we are only going to
    print. It cannot mask a real failure — the returncode check below is now
    actually reachable, and the apply path's BEGIN;/COMMIT; with ON_ERROR_STOP=1
    means a truncated *input* fails to commit and is reported rather than
    half-applied.
    """
    cmd = PSQL.split()
    proc = subprocess.run(
        cmd + ["-v", "ON_ERROR_STOP=1", "-t", "-A"],
        input=sql.encode("utf-8"),
        capture_output=capture,
    )

    def _dec(b):
        return (b or b"").decode("utf-8", errors="replace")

    if proc.returncode != 0:
        sys.exit(f"psql failed:\n{_dec(proc.stderr) or _dec(proc.stdout)}")
    return _dec(proc.stdout).strip()


def existing_sources():
    """{source: sorted list of subject_keys} — IDENTITY, not a count.

    This returned {source: row count} until 2026-08-14, and `refootprinted`
    compared counts. The splitter fix that day changed the footprint SET of 185
    entries while leaving the COUNT equal on 6 of them — a count comparison
    would have left those six's stale subject_keys in doc_notes for ever, with
    every sync reporting clean. A count proves the damage, never the no-op.
    """
    out = run_psql(
        "SELECT source, string_agg(subject_key, E'\\x1f' ORDER BY subject_key) "
        "FROM doc_notes "
        f"WHERE source LIKE {sql_lit(SOURCE_PREFIX + '%')} GROUP BY 1;"
    )
    got = {}
    for line in out.splitlines():
        if "|" in line:
            src, keys = line.split("|", 1)
            got[src.strip()] = sorted(keys.strip().split("\x1f"))
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
    ap.add_argument(
        "--verbose-warnings",
        action="store_true",
        help="also print the prose-footprint style advisories (~168 on the current corpus)",
    )
    ap.add_argument(
        "--full",
        action="store_true",
        help="with --apply: rewrite EVERY entry instead of just the delta. The old "
             "behaviour, kept as an escape hatch if the delta logic is ever suspected "
             "of missing a case. Sends the whole corpus, which is what broke the "
             "kubectl exec transport on 2026-08-12 — expect it to fail at scale.",
    )
    args = ap.parse_args()

    # parse() took no on_warn until 2026-08-02, so every warning it raised —
    # "skipped (no footprint)", malformed headings — went to nobody. A parser that
    # reports nothing looks exactly like a file with nothing wrong in it.
    warnings = []
    entries = parse(LANDMINES, on_warn=warnings.append)
    if not entries:
        sys.exit("no entries parsed — refusing to run (this would delete every owned row)")

    # Partition, or the signal drowns. 'prose footprint' fires ~168 times on the
    # current corpus — a standing style advisory, not something this run broke.
    # Printing all of it would bury the two kinds that mean a landmine is NOT
    # BEING DELIVERED, which is the whole reason warnings were wired up.
    blocking = [w for w in warnings if not w.startswith("prose footprint")]
    stylistic = len(warnings) - len(blocking)
    if blocking:
        print(f"!! {len(blocking)} warning(s) that cost DELIVERY:")
        for w in blocking:
            print(f"    {w}")
    if stylistic:
        print(f"(+{stylistic} prose-footprint style advisories — "
              f"run with --verbose-warnings to see them)")
    if args.verbose_warnings:
        for w in warnings:
            if w.startswith("prose footprint"):
                print(f"    {w}")

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
    print(f"\ndoc_notes: {sum(len(v) for v in have.values())} owned row(s) across {len(have)} entr(ies)")
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

    # DELTA apply (2026-08-12). This used to DELETE every owned row and reinsert
    # all of them every run — ~4.6MB of statements once the corpus reached ~2,155
    # rows, and `kubectl exec` stopped carrying it: "error reading from error
    # stream: read message: unexpected EOF", i.e. THE THIRD TIME this sync has
    # broken purely because the file it syncs got bigger (argv limit 2026-07-30,
    # the decode crash above, now the stream). A full replace also had a nastier
    # property: it deleted the whole corpus first, so a transport failure
    # mid-apply is the one case that could leave doc_notes short of landmines.
    #
    # The script already knows exactly what moved, so send only that. Semantics
    # are unchanged — this is still an idempotent upsert scoped by `source`, and
    # another thread's rows (different `source`) remain structurally out of reach.
    #
    # `refootprinted` is the third case and is easy to miss: an entry can keep a
    # byte-identical body while its FOOTPRINT LIST changes, which `changed`
    # (body comparison) cannot see. `have` is source -> sorted subject_keys, one
    # row per footprint, so a SET comparison catches exactly that. (Until
    # 2026-08-14 this compared counts — the splitter fix that day re-keyed 6
    # entries at an unchanged count, which a count comparison misses silently.)
    refootprinted = [
        s for s in want
        if s in have and have[s] != sorted(want[s]["footprints"]) and s not in changed
    ]
    touch = new + changed + refootprinted
    if refootprinted:
        print(f"  footprints changed (body identical): {len(refootprinted)}")
        for s in refootprinted:
            print(f"    refootprinted: {s}")

    if args.full:
        touch = list(want)
        print("  --full: rewriting every entry, not just the delta")

    if not touch and not gone:
        print("\nnothing to apply — already in sync")
        return 0

    stmts = ["BEGIN;"]
    # Scoped by `source`: only this script's own rows, and only for the entries
    # that actually moved.
    for src in gone + touch:
        stmts.append(f"DELETE FROM doc_notes WHERE source = {sql_lit(src)};")
    for src in touch:
        e = want[src]
        cats = json.dumps(["landmine", "synced-from-markdown"])
        for fp in e["footprints"]:
            stmts.append(
                "INSERT INTO doc_notes (subject_type, subject_key, body, categories, "
                "source, created_by) VALUES ("
                f"'landmine', {sql_lit(fp)}, {sql_lit(e['body'])}, "
                f"{sql_lit(cats)}::jsonb, {sql_lit(src)}, 'landmines-sync');"
            )
    stmts.append("COMMIT;")

    payload = "\n".join(stmts)
    print(f"\napplying delta: {len(touch)} entr(ies) rewritten, {len(gone)} orphan(s) "
          f"removed, {len(payload)} bytes on the wire")
    run_psql(payload, capture=True)
    after = existing_sources()
    print(f"\napplied: {sum(len(v) for v in after.values())} owned row(s) now present")

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

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

    RETRY on the transport's mid-stream EOF (2026-08-25) — the FOURTH time this
    sync has broken purely because the corpus got bigger, and the first
    PROBABILISTIC one: at ~846 entries the corpus read-back (existing_state(),
    which then shipped full bodies rather than hashes) returned ~3MB, and the
    `kubectl exec` stream now dies mid-transfer roughly every other call
    ("error reading from error stream: read message: unexpected EOF" — measured
    2026-08-25: 4/4 script runs failed, while the same read standalone went
    rc=1 at 1.9MB then rc=0 at 3.1MB back to back). One failed call killed the
    whole run, so at several large calls per run the sync almost never
    completed — and a sync that fails mid-read-back looks, to its
    operator, exactly like a DB problem. Retrying is safe on every call this
    script makes: reads are pure, and the apply path is a single
    BEGIN/COMMIT + ON_ERROR_STOP idempotent upsert scoped by `source`, so a
    stream that died before, during or after COMMIT re-runs to the same state.
    Only the transport error retries; a real psql error (nonzero rc WITHOUT the
    stream signature) still exits on the first attempt.
    """
    cmd = PSQL.split()

    def _dec(b):
        return (b or b"").decode("utf-8", errors="replace")

    attempts = 3
    for attempt in range(1, attempts + 1):
        proc = subprocess.run(
            cmd + ["-v", "ON_ERROR_STOP=1", "-t", "-A"],
            input=sql.encode("utf-8"),
            capture_output=capture,
        )
        if proc.returncode == 0:
            return _dec(proc.stdout).strip()
        stream_broke = "unexpected EOF" in _dec(proc.stderr)
        if stream_broke and attempt < attempts:
            print(f"  transport EOF mid-stream (attempt {attempt}/{attempts}) — retrying",
                  file=sys.stderr)
            continue
        sys.exit(f"psql failed:\n{_dec(proc.stderr) or _dec(proc.stdout)}")


def md5_hex(s):
    return hashlib.md5(s.encode("utf-8")).hexdigest()


def existing_state():
    """{source: (row_count, footprint_set_hash, body_hash)} for every owned entry.

    IDENTITY, not a count — the footprint comparison keyed on {source: row
    count} until 2026-08-14, and the splitter fix that day changed the
    footprint SET of 185 entries while leaving the COUNT equal on 6 of them:
    a count comparison would have left those six's stale subject_keys in
    doc_notes for ever, with every sync reporting clean.

    HASHES, not the values (2026-08-26, bugs 402). The delta needs only
    equality, but until then this script shipped every body (3.0MB) and every
    subject_key (610KB) through `kubectl exec` to get it — and that stream
    dies mid-transfer roughly every other call at 3MB (the fourth
    corpus-growth failure; see run_psql's history). The retry above makes
    that survivable; hashing server-side makes it improbable: ~130 bytes per
    entry on the wire, so the 856-entry corpus reads back in ~110KB and the
    next scale wall moves out ~25x. A hash mismatch degrades to one spurious
    idempotent rewrite of that entry — never data loss.

    THE SORT MUST MATCH PYTHON'S. The footprint-set hash aggregates
    `ORDER BY subject_key COLLATE "C"` (byte order) because the Python side
    hashes '\\x1f'.join(sorted(fps)), and str sort is code-point order, which
    UTF-8 byte order preserves. The database's DEFAULT collation promises no
    such thing — a locale-ordered agg could hash differently for ever, and
    every run would then rewrite every entry. Verified against the old
    full-payload comparison on all 856 live entries, zero disagreements,
    2026-08-26.

    body is genuine multi-line prose, which is why the old code read it as
    jsonb; an md5 hex digest cannot contain a newline or '|', so plain
    `-t -A` rows are safe for all three columns. Every footprint row of one
    entry carries an identical body (see the INSERT loop), so min(body) is
    THE body.
    """
    out = run_psql(
        "SELECT source, count(*), "
        "md5(string_agg(subject_key, E'\\x1f' ORDER BY subject_key COLLATE \"C\")), "
        "md5(min(body)) "
        "FROM doc_notes "
        f"WHERE source LIKE {sql_lit(SOURCE_PREFIX + '%')} GROUP BY 1;"
    )
    got = {}
    for line in out.splitlines():
        if "|" in line:
            src, n, fp_hash, body_hash = line.split("|", 3)
            got[src.strip()] = (int(n), fp_hash, body_hash)
    return got


def owned_row_count():
    return run_psql(
        f"SELECT count(*) FROM doc_notes WHERE source LIKE {sql_lit(SOURCE_PREFIX + '%')};"
    )


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
        help="also print the style advisories — prose footprints and '##' headings "
             "(~330 on the current corpus)",
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

    # Partition by what each warning actually COSTS, or the signal both drowns
    # and inflates. Only "skipped (no footprint)" loses an entry's delivery.
    # The '##' heading nag does not — the entry parses correctly (landmines_lib
    # says so where it emits the warning; the ##-headed compose-.env entry has
    # 4 delivered rows and two verifier verdicts) — but until 2026-08-26 all
    # 164 of them printed under "!! warning(s) that cost DELIVERY", a false
    # banner that bugs_open/402 repeated as fact. 'prose footprint' fires ~168
    # times. Both are standing style advisories, not something this run broke;
    # printing them all would bury a real loss.
    blocking = [w for w in warnings if w.startswith("skipped")]
    advisory = len(warnings) - len(blocking)
    if blocking:
        print(f"!! {len(blocking)} entr(ies) NOT DELIVERED:")
        for w in blocking:
            print(f"    {w}")
    if advisory:
        print(f"(+{advisory} style advisories ('##' headings, prose footprints) — "
              f"run with --verbose-warnings to see them)")
    if args.verbose_warnings:
        for w in warnings:
            if not w.startswith("skipped"):
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

    have = existing_state()
    new = [s for s in want if s not in have]
    gone = [s for s in have if s not in want]
    # Content changed under an unchanged slug: new/gone (source presence
    # alone) cannot see a hand-edit to an existing entry — typo fix,
    # corrected footprint, tightened "the check". Needed for
    # landmine-verifier (RFC_005 3.2): a changed entry is exactly as much in
    # need of re-verification as a brand new one.
    changed = [s for s in want if s in have and have[s][2] != md5_hex(want[s]["body"])]
    # An entry can also keep a byte-identical body while its FOOTPRINT LIST
    # changes, which the body comparison cannot see; the footprint-set hash
    # catches exactly that.
    refootprinted = [
        s for s in want
        if s in have and s not in changed
        and have[s][1] != md5_hex("\x1f".join(sorted(want[s]["footprints"])))
    ]

    print(f"\ndoc_notes: {sum(v[0] for v in have.values())} owned row(s) across {len(have)} entr(ies)")
    # The DELTA, never len(want): until 2026-08-26 this line printed the whole
    # corpus count labelled "to insert/refresh", and on the night the
    # transport broke, that read as "the delta logic wants to rewrite all 847
    # entries" — bugs 402's title. The numbers printed here must be the ones
    # the apply branch will actually send.
    print(f"  delta vs the file: {len(new)} new, {len(changed)} changed, "
          f"{len(refootprinted)} refootprinted, {len(gone)} orphaned")
    for s in new:
        print(f"    new: {s}")
    for s in changed:
        print(f"    changed: {s}")
    for s in refootprinted:
        print(f"    refootprinted: {s}")
    for s in gone:
        print(f"    orphan: {s}")

    if args.check:
        # changed/refootprinted count as drift too — they were invisible to
        # --check until 2026-08-26, so a hand-edited entry reported "in sync"
        # while doc_notes served the old body.
        drift = new or gone or changed or refootprinted
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
    touch = new + changed + refootprinted

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
    print(f"\napplied: {owned_row_count()} owned row(s) now present")

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

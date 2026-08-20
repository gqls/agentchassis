#!/usr/bin/env python3
"""
Update dartsonline.com's `evidence_base` privacy copy after the owner REMOVED the
last sentence of the "Affiliate links" paragraph (instruction 2026-08-20).

WHY A SECOND SCRIPT AND NOT A FLAG ON THE FIRST ONE.
apply_privacy_evidence_base.py CREATES: it aborts if a current row exists, and that
abort is a real guard, not friction — it is what stops a re-run silently clobbering
bans or facts added since. This script is the opposite job and carries the opposite
precondition (exactly ONE current row must exist), so the two cannot be confused and
neither can do the other's damage.

WHAT IT DOES DIFFERENTLY, and it matters:
  - It DERIVES from the live row (`{**live, ...}`), the noted.co.uk pattern — which was
    inapplicable on 2026-08-16 because no row existed and IS applicable now. Anything
    another session has added to this aspect since (facts, banned_claims, schema notes)
    carries across untouched.
  - It does NOT retype the writer_block from a template. It takes the LIVE writer_block
    and replaces the old body substring with the new one, asserting exactly one
    occurrence first. A template rebuild would silently discard any edit made to that
    block since it was written.

WHY THE SENTENCE WENT (both reasons, because they are independent):
  1. OWNER, 2026-08-20: it leads the reader away from the product and its benefits and
     into legalese and unhelpful detail. His copy, his call — it is REMOVED, not reworded.
  2. The framework had already refused to build the page on it `[MEASURED 2026-08-17]`:
     validate_page_content returned 1 blocker, banned claim "does not appear here"
     (completeness-of-exclusion, short form — claims_global.go:130), leaving the work item
     parked at needs_human_review since 12:24 that day.

The draft file stays the source of truth for the wording: the copy is EXTRACTED from it,
never retyped here, so the document and the database cannot drift.
"""
import json
import re
import subprocess
import sys
from pathlib import Path

DRAFT = Path(__file__).with_name("DRAFT_2026-08-15_privacy_copy_for_owner_approval.md")
DOMAIN = "dartsonline.com"
BANNED = "does not appear here"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]


def psql(sql: str, quiet: bool = False) -> str:
    args = PSQL + (["-t", "-A"] if quiet else [])
    r = subprocess.run(args + ["-c", sql], capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr}")
    return r.stdout.strip()


def extract_copy() -> str:
    """Identical extractor to apply_privacy_evidence_base.py — deliberately, so the two
    scripts cannot disagree about what the approved copy IS."""
    text = DRAFT.read_text(encoding="utf-8")
    m = re.search(r"^## The draft copy\s*\n(.*?)^---", text, re.S | re.M)
    if not m:
        sys.exit("could not locate '## The draft copy' section in the draft file")
    lines = []
    for ln in m.group(1).splitlines():
        if ln.startswith("> "):
            lines.append(ln[2:])
        elif ln.strip() == ">":
            lines.append("")
        elif ln.strip() == "":
            continue
        else:
            sys.exit(f"unexpected non-quoted line inside the copy block: {ln!r}")
    body = "\n".join(lines).strip()
    if "Privacy and Cookies" not in body or "ico.org.uk" not in body:
        sys.exit("extracted copy failed its sanity check (missing title or ICO line)")
    if BANNED in body:
        sys.exit(f"ABORT: the draft still contains {BANNED!r} — edit the draft first; "
                 "this script does not edit the owner's copy, it only carries it")
    return body


def main() -> None:
    body = extract_copy()
    print(f"extracted {len(body)} chars of corrected copy from {DRAFT.name}")

    raw = psql(
        f"SELECT ss.id || E'\\x1f' || ss.data::text FROM site_specs ss JOIN sites s ON s.id=ss.site_id "
        f"WHERE s.domain='{DOMAIN}' AND ss.aspect='evidence_base' AND ss.is_current;",
        quiet=True)
    rows = [ln for ln in raw.splitlines() if ln.strip()]
    if len(rows) != 1:
        sys.exit(f"ABORT: expected exactly 1 current evidence_base row, found {len(rows)}. "
                 "This script UPDATES; creating is apply_privacy_evidence_base.py's job.")
    row_id, data_text = rows[0].split("\x1f", 1)
    live = json.loads(data_text)

    old_body = (live.get("supplied_copy", {}).get("privacy", {}) or {}).get("body_markdown")
    if not old_body:
        sys.exit("ABORT: live row has no supplied_copy.privacy.body_markdown to replace")
    if old_body == body:
        sys.exit("Nothing to do: the live row already carries this exact copy.")

    wb = live.get("writer_block") or ""
    if wb.count(old_body) != 1:
        sys.exit(f"ABORT: the live body appears {wb.count(old_body)} times in writer_block, "
                 "expected exactly 1. Refusing to guess where the copy lives.")

    # Derive, never retype: everything else in the aspect carries across untouched.
    new_privacy = dict(live["supplied_copy"]["privacy"])
    new_privacy["body_markdown"] = body
    new_privacy["approved_on"] = "2026-08-15; affiliate sentence removed by owner 2026-08-20"
    # NB: this note must NOT quote the removed sentence. The aspect is scanned as a whole
    # by the check below, and an earlier draft of this script reproduced the banned phrase
    # verbatim inside its own explanation — which would have re-blocked the page with the
    # copy already fixed. Describe it; do not quote it.
    new_privacy["revision_note"] = (
        "The final sentence of the 'Affiliate links' paragraph — the editorial-independence "
        "promise, asserting that commission never affects what the guides recommend — was "
        "REMOVED at the owner's instruction on 2026-08-20: it read as legalese and led the "
        "reader away from the product. It had also blocked the page build on 2026-08-17 by "
        "matching a fleet-wide completeness-of-exclusion pattern (claims_global.go:130). The "
        "disclosure itself — affiliate links, commission at no extra cost, retailer cookie — "
        "is retained in full.")

    new_data = {**live,
                "writer_block": wb.replace(old_body, body),
                "supplied_copy": {**live["supplied_copy"], "privacy": new_privacy}}

    blob = json.dumps(new_data)
    if BANNED in blob:
        sys.exit(f"ABORT: {BANNED!r} still present somewhere in the new aspect — "
                 "the page would be refused again for the same reason")

    payload = blob.replace("'", "''")
    print(psql(f"""
BEGIN;
UPDATE site_specs SET is_current = false, superseded_at = NOW(), updated_at = NOW()
 WHERE id = '{row_id}';
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes, is_current)
SELECT s.id, 'evidence_base', '{payload}'::jsonb, 'manual', 'dartsonline_traffic',
       'dartsonline_traffic',
       'Owner-approved privacy copy; affiliate sentence removed by owner 2026-08-20.', true
FROM sites s WHERE s.domain='{DOMAIN}'
RETURNING id;
COMMIT;"""))

    # Post-conditions. body_in_writer_block is the load-bearing one (supplied_copy has no
    # Go readers; writer_block is what the writer is actually handed), and banned_absent is
    # the one this revision exists for.
    check = psql(f"""
SELECT format('rows=%s body_len=%s body_in_writer_block=%s managed_flag_absent=%s banned_absent=%s superseded=%s',
  count(*) FILTER (WHERE ss.is_current),
  max(length(ss.data->'supplied_copy'->'privacy'->>'body_markdown')) FILTER (WHERE ss.is_current),
  bool_or(position(ss.data->'supplied_copy'->'privacy'->>'body_markdown' in ss.data->>'writer_block')>0) FILTER (WHERE ss.is_current),
  bool_and(NOT (ss.data ? 'writer_block_managed')) FILTER (WHERE ss.is_current),
  bool_and(position('{BANNED}' in ss.data::text) = 0) FILTER (WHERE ss.is_current),
  count(*) FILTER (WHERE NOT ss.is_current))
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='{DOMAIN}' AND ss.aspect='evidence_base';""", quiet=True)
    print("post-write:", check)
    for must in ("rows=1", f"body_len={len(body)}", "body_in_writer_block=t",
                 "managed_flag_absent=t", "banned_absent=t"):
        if must not in check:
            sys.exit(f"POST-CONDITION FAILED (expected {must!r}) — investigate before "
                     "re-queueing the page build")
    print("OK: one current row carrying the corrected copy, inside the writer_block the "
          "writer reads; banned phrase absent; previous revision superseded, not deleted.")


if __name__ == "__main__":
    main()

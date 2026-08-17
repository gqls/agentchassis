#!/usr/bin/env python3
"""
Create dartsonline.com's `evidence_base` site_specs aspect, carrying the
owner-approved privacy copy so the framework's writer reproduces it VERBATIM.

Owner approved the copy 2026-08-15 and supplied the controller identity
(Fine Tuning, Fleetside, West Molesey, East Surrey). Owner instruction
2026-08-16: "create the evidence base record and add it, then let the
framework write it."

WHY A SCRIPT AND NOT HAND-WRITTEN SQL — the same two guarantees the noted.co.uk
lane's apply_privacy_copy.py was written for:

  1. The copy is EXTRACTED FROM THE DRAFT FILE, so the document and the
     database cannot drift. The draft stays the source of truth for the wording.
  2. The JSON is built by a serialiser, not typed out, so quoting/escaping of
     the owner's prose cannot be silently mangled on its way into jsonb.

DIFFERENCE FROM THE NOTED SCRIPT, and why it is NOT that script parameterised:
noted DERIVED its new row from a live evidence_base row (`data || {...}`) so the
existing banned_claims and facts carried across untouched. **dartsonline has no
evidence_base row at all** (14 live aspects, none of them evidence_base —
measured 2026-08-16), so there is nothing to derive from and that entire safety
argument is inapplicable. This CREATES the aspect.

⚠ KNOWN CONSEQUENCE, MEASURED, AND DELIBERATE — read before running.
`validate_page_content.go:360` gates the unregistered-number scan on
`eb != nil`. dartsonline has no evidence_base today, so that scan is OFF for the
whole site; creating this row turns it ON. Editorial page types are exempt
(`guide`, `blog-post`, `news-index`, `tool`, `game` — claims.go:752), which
covers 14 of the site's 25 active pages. The other 11 (3 content, 3 landing,
3 section-index, 2 entity-page) become subject to it on their NEXT build.
Severity is `error`, never `blocker`, so the effect is mark_needs_review rather
than a failed deploy — but review queues on this site are known not to drain.
`facts` is left EMPTY, matching noted.co.uk's shipped row: registering invented
citation-backed "facts" for a phone number would be fabricating evidence
structure to silence a check, which is worse than the check firing.

Verifies after writing: exactly one current row, the body readable back and
byte-identical to the draft, and the writer_block present.
"""
import json
import re
import subprocess
import sys
from pathlib import Path

DRAFT = Path(__file__).with_name("DRAFT_2026-08-15_privacy_copy_for_owner_approval.md")
DOMAIN = "dartsonline.com"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]


def psql(sql: str, quiet: bool = False) -> str:
    args = PSQL + (["-t", "-A"] if quiet else [])
    r = subprocess.run(args + ["-c", sql], capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr}")
    return r.stdout.strip()


def extract_copy() -> str:
    """Pull the blockquoted copy out of the draft, between '## The draft copy'
    and the next '---' rule. Strips the leading '> ' quote markers only."""
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
    return body


def main() -> None:
    body = extract_copy()
    print(f"extracted {len(body)} chars of approved copy from {DRAFT.name}")

    existing = psql(
        f"SELECT count(*) FROM site_specs ss JOIN sites s ON s.id=ss.site_id "
        f"WHERE s.domain='{DOMAIN}' AND ss.aspect='evidence_base' AND ss.superseded_at IS NULL;",
        quiet=True)
    if existing != "0":
        sys.exit(f"ABORT: {DOMAIN} already has {existing} current evidence_base row(s). "
                 "This script CREATES; deriving from a live row is a different job — "
                 "see the noted.co.uk pattern and do not clobber existing bans/facts.")

    # ⚠ THE COPY MUST BE INLINE HERE, not only in supplied_copy.
    # `supplied_copy` has ZERO Go readers (grep supplied_copy --include=*.go -> nothing);
    # `writer_block` is the key the writer actually consumes
    # (refresh_evidence_base_action.go:16). noted.co.uk's shipped row proves the shape:
    # writer_block 3869 chars CONTAINING its 1582-char body inline (measured 2026-08-16).
    # A row that puts the copy only in supplied_copy hands the writer an instruction to
    # reproduce something it was never shown, and it will improvise instead.
    writer_block = (
        "This is a UK darts publication. It is online only: it does not hold stock, take "
        "payments, or ship anything.\n\n"
        "Do not write or invent privacy, cookie, data-protection or company-identity wording "
        "for this site. That copy is the owner's and IT HAS BEEN SUPPLIED (approved "
        "2026-08-15). It is reproduced IN FULL below. On the privacy page use it WORD FOR "
        "WORD — do not paraphrase it, do not summarise it, do not add reassurance sentences "
        "to it, and do not add sections it does not contain. If it does not answer "
        "something, leave it unanswered rather than inventing an answer.\n\n"
        "This applies to the privacy page specifically and to any privacy/cookie/data claim "
        "anywhere else on the site: elsewhere, say nothing rather than improvise.\n\n"
        "----- BEGIN OWNER-SUPPLIED PRIVACY COPY (verbatim) -----\n\n"
        f"{body}\n\n"
        "----- END OWNER-SUPPLIED PRIVACY COPY -----"
    )

    data = {
        "writer_block": writer_block,
        # `writer_block_managed` deliberately ABSENT: refresh_evidence_base only
        # regenerates the block where that flag is true, so leaving it out is what
        # protects this hand-written block from being overwritten
        # (refresh_evidence_base_action.go:35).
        "facts": [],
        "banned_claims": [],
        "governing_rule": (
            "Privacy, cookie and controller-identity wording is owner-supplied and reproduced "
            "verbatim. Nothing on this site may assert that it holds stock, ships goods, takes "
            "payments, or has a retail premises."),
        "schema_notes": (
            "Created 2026-08-16 by the dartsonline_traffic lane at the owner's instruction, to "
            "carry approved privacy copy to the writer. facts[] and banned_claims[] are "
            "deliberately EMPTY: this row exists for supplied_copy, and populating either "
            "without measuring the 25 live pages first would gate them on unmeasured rules. "
            "See the consequence note in apply_privacy_evidence_base.py before adding to them."),
        "audit_doc": "docs/agent_docs/docs024_key_docs_latest/dartsonline_traffic/DRAFT_2026-08-15_privacy_copy_for_owner_approval.md",
        "supplied_copy": {
            "privacy": {
                "title": "Privacy and Cookies",
                "body_markdown": body,
                "approved_by": "owner",
                "approved_on": "2026-08-15",
                "source_doc": DRAFT.name,
                "checker": "body must appear verbatim on /privacy.html; controller identity "
                           "line must read 'Fine Tuning, of Fleetside, West Molesey, East Surrey'",
                "open_question": "No postcode was supplied for the controller address; an "
                                 "ICO-facing address would normally carry one. Owner's call.",
            }
        },
    }

    payload = json.dumps(data).replace("'", "''")
    sql = f"""
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes, is_current)
SELECT s.id, 'evidence_base', '{payload}'::jsonb, 'manual', 'dartsonline_traffic',
       'dartsonline_traffic', 'Owner-approved privacy copy, 2026-08-15. Created 2026-08-16.', true
FROM sites s WHERE s.domain='{DOMAIN}'
RETURNING id;
"""
    print(psql(sql))

    # Post-conditions. The load-bearing one is body_in_writer_block: without it the
    # writer is told to reproduce copy it was never shown.
    check = psql(f"""
SELECT format('rows=%s body_len=%s wb_len=%s body_in_writer_block=%s managed_flag_absent=%s',
  count(*),
  max(length(ss.data->'supplied_copy'->'privacy'->>'body_markdown')),
  max(length(ss.data->>'writer_block')),
  bool_or(position(ss.data->'supplied_copy'->'privacy'->>'body_markdown' in ss.data->>'writer_block')>0),
  bool_and(NOT (ss.data ? 'writer_block_managed')))
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE s.domain='{DOMAIN}' AND ss.aspect='evidence_base' AND ss.superseded_at IS NULL;""",
                 quiet=True)
    print("post-write:", check)
    for must in (f"rows=1", f"body_len={len(body)}",
                 "body_in_writer_block=t", "managed_flag_absent=t"):
        if must not in check:
            sys.exit(f"POST-CONDITION FAILED (expected {must!r}) — investigate before "
                     f"letting the framework build the page")
    print("OK: one current row; copy byte-length-identical to the draft; copy is INSIDE "
          "the writer_block the writer actually reads; regeneration flag absent.")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
Register the owner-approved privacy copy in noted.co.uk's evidence_base.

Approved by the owner 2026-08-12 (with one edit: "plainly" -> "spell it out").

WHY THIS IS A SCRIPT AND NOT HAND-WRITTEN SQL — two things it guarantees that
typing the blob again would not:

  1. The copy is EXTRACTED FROM THE DRAFT FILE, so the doc and the database
     cannot drift. The file stays the source of truth for the wording.
  2. The new spec row is DERIVED from the live one (`data || {...}`), so the
     seven banned_claims, the facts array, governing_rule, audit_doc and
     schema_notes are carried across untouched. Retyping the blob is how a ban
     silently disappears, and nothing downstream would report it.

It supersedes the current row and inserts a new one, exactly as
SEED_2026-08-10_noted_site_and_specs.sql does. Idempotent-ish: re-running makes
another version, which is the versioning model working, not a fault.

Verifies after writing: 7 bans still present, and the copy readable back.
"""
import json
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).parent
DRAFT = HERE / "COPY_2026-08-12_privacy_DRAFT_for_owner.md"

NEW_CLAUSE = (
    "Do not write privacy or security promises, and do not invent the wording for how this "
    "product describes what it does with people's data. That copy is the owner's, and it HAS "
    "NOW BEEN SUPPLIED (2026-08-12): it is in this spec under `supplied_copy.privacy`. Use it "
    "VERBATIM where a page needs it. Do not paraphrase it, do not extend it, and do not write "
    "a second version of it somewhere else on the site. If a page seems to need a privacy "
    "statement this copy does not cover, that is a question for the owner, not a gap for you "
    "to fill."
)
OLD_CLAUSE_RE = re.compile(
    r"Do not write privacy or security promises.*?will be supplied\.", re.S
)


def psql(sql, tuples_only=True):
    cmd = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
           "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]
    if tuples_only:
        cmd += ["-tA"]
    r = subprocess.run(cmd + ["-c", sql] if len(sql) < 3000 else cmd,
                       input=None if len(sql) < 3000 else sql,
                       capture_output=True, text=True, timeout=120)
    if r.returncode != 0:
        print("psql failed:\n", r.stderr[:1500]); sys.exit(2)
    return r.stdout.strip()


def approved_copy():
    body = DRAFT.read_text(encoding="utf-8")
    section = body.split("## THE DRAFT", 1)[1].split("## Verification", 1)[0]
    # drop the leading heading line, keep the prose
    section = section.split("### Your notes, and what happens to them", 1)[1]
    text = section.strip()
    # Guards run against a WHITESPACE-NORMALISED copy: the markdown is hard-wrapped,
    # so "spell it out" legitimately appears as "spell it\nout" and a literal test
    # fails on a correct file. (It did, first run — the guard was right to fire and
    # the fix is to normalise, not to drop the check.)
    flat = " ".join(text.split())
    if "spell it out" not in flat:
        print("FATAL: the owner's edit ('spell it out') is not in the draft — wrong file?")
        sys.exit(2)
    if "plainly" in flat:
        print("FATAL: 'plainly' still present in the draft copy — the edit did not apply")
        sys.exit(2)
    return text


def main():
    copy = approved_copy()
    print(f"extracted {len(copy.split())} words of approved copy")

    current = psql("""SELECT ss.data->>'writer_block'
                      FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                      WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    if not current:
        print("FATAL: no current evidence_base"); sys.exit(2)
    if not OLD_CLAUSE_RE.search(current):
        print("FATAL: could not find the 'will be supplied' clause to replace.")
        print("The writer_block has changed since this script was written — read it and re-check.")
        sys.exit(2)

    new_block = OLD_CLAUSE_RE.sub(NEW_CLAUSE.replace("\\", "\\\\"), current, count=1)

    payload = json.dumps({
        "writer_block": new_block,
        "supplied_copy": {
            "privacy": {
                "title": "Your notes, and what happens to them",
                "body_markdown": copy,
                "approved_by": "owner",
                "approved_on": "2026-08-12",
                "source_doc": "docs/agent_docs/docs024_key_docs_latest/noted_rebuild/"
                              "COPY_2026-08-12_privacy_DRAFT_for_owner.md",
                "checker": "COPY_2026-08-12_privacy_check.py — clean against all 7 bans, "
                           "positive control fires",
                "open_question": "Deletion vs the 30-day backup object lock is deliberately "
                                 "not addressed in this copy. Owner's call; do not invent a "
                                 "sentence for it.",
            }
        },
    })

    sql = """
BEGIN;
CREATE TEMP TABLE _eb ON COMMIT DROP AS
  SELECT ss.* FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;

DO $do$ DECLARE n int; BEGIN
  SELECT count(*) INTO n FROM _eb;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current evidence_base, found %', n; END IF;
END $do$;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE id IN (SELECT id FROM _eb);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT site_id, aspect,
       data || $payload$__PAYLOAD__$payload$::jsonb,
       'manual',
       'Owner-approved privacy copy registered 2026-08-12. Derived from the previous row so the 7 banned_claims and facts carry across untouched.',
       true, pinned, 'owner-approved (claude session)'
FROM _eb;
COMMIT;
""".replace("__PAYLOAD__", payload)

    cmd = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
           "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]
    r = subprocess.run(cmd, input=sql, capture_output=True, text=True, timeout=120)
    print(r.stdout.strip() or r.stderr.strip()[:800])
    if r.returncode != 0:
        sys.exit(2)

    # ---- verify: the bans MUST have survived, and the copy must read back ----
    print("\nVERIFY")
    bans = psql("""SELECT jsonb_array_length(ss.data->'banned_claims')
                   FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                   WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print(f"  banned_claims still present: {bans} (must be 7)")
    got = psql("""SELECT length(ss.data->'supplied_copy'->'privacy'->>'body_markdown')
                  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print(f"  supplied_copy.privacy body chars: {got} (local draft: {len(copy)})")
    supplied = psql("""SELECT (ss.data->>'writer_block') LIKE '%%HAS NOW BEEN SUPPLIED%%'
                       FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                       WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print(f"  writer_block updated: {supplied}")
    ok = (bans == "7") and (got == str(len(copy))) and (supplied == "t")
    print("\n" + ("REGISTERED OK" if ok else "MISMATCH — investigate before relying on this"))
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())

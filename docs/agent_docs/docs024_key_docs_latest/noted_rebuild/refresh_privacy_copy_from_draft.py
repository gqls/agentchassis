#!/usr/bin/env python3
"""
Refresh noted.co.uk's evidence_base from the privacy DRAFT — RE-RUNNABLE.

Supersedes the two 08-12 one-shots for any FUTURE copy edit:
  - apply_privacy_copy.py     (FATALs now: its replace-target clause is gone)
  - embed_privacy_copy_in_writer_block.py  ("already embedded" checks PRESENCE,
                                            not currency — a changed draft
                                            reads as nothing-to-do)
Both were built for their one-time migration; this one is idempotent against
the draft: it sets supplied_copy.privacy to the draft body and swaps the
verbatim block inside writer_block (between its "## THE APPROVED PRIVACY COPY,
VERBATIM" heading and its "(End of approved copy." marker), keeping both
markers. If the stores already match the draft it says so and writes nothing.

Same discipline as its predecessors: copy extracted FROM THE DRAFT at run
time; new spec row DERIVED from the live one (bans/facts/everything carried);
verifies afterwards that the bans survived and a distinctive new sentence is
actually in the writer_block.

After this: run 074b to put the words on the PAGE (it reads the draft too).
"""
import json
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).parent
DRAFT = HERE / "COPY_2026-08-12_privacy_DRAFT_for_owner.md"
HEAD_MARK = "## THE APPROVED PRIVACY COPY, VERBATIM"
END_MARK = "(End of approved copy."

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]


def run(sql, tuples=True):
    cmd = PSQL + (["-tA"] if tuples else [])
    r = subprocess.run(cmd, input=sql, capture_output=True, text=True, timeout=120)
    if r.returncode != 0:
        print("psql failed:\n", r.stderr[:1500]); sys.exit(2)
    return r.stdout.strip()


def approved_copy():
    body = DRAFT.read_text(encoding="utf-8")
    body = body.split("## THE DRAFT", 1)[1].split("## Verification", 1)[0]
    body = body.split("### Your notes, and what happens to them", 1)[1].strip()
    flat = " ".join(body.split())
    if "spell it out" not in flat or "plainly" in flat:
        print("FATAL: draft does not carry the owner's approved wording"); sys.exit(2)
    return body


def main():
    copy = approved_copy()
    print(f"draft: {len(copy.split())} words")

    row = run("""SELECT ss.data->>'writer_block', ss.data->'supplied_copy'->>'privacy',
                        jsonb_array_length(ss.data->'banned_claims')
                 FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                 WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    wb, sc, bans = row.split("|", 2)[0], row.split("|", 2)[1], row.rsplit("|", 1)[1]
    if HEAD_MARK not in wb or END_MARK not in wb:
        print("FATAL: writer_block no longer carries the verbatim block markers — read it first")
        sys.exit(2)

    flat = " ".join(copy.split())
    if " ".join(sc.split()) == flat and flat in " ".join(wb.split()):
        print("Stores already match the draft — nothing to do.")
        return 0

    head = wb.split(HEAD_MARK, 1)[0]
    tail = END_MARK + wb.split(END_MARK, 1)[1]
    # The heading line itself is regenerated so its date reflects the draft file.
    new_wb = (head + HEAD_MARK
              + " (owner-approved; source file COPY_2026-08-12_privacy_DRAFT_for_owner.md)"
              + " — page title “Your notes, and what happens to them”\n\n"
              + copy + "\n\n" + tail)

    payload = json.dumps({"writer_block": new_wb,
                          "supplied_copy": {"privacy": copy}}, ensure_ascii=False)
    sql = """BEGIN;
CREATE TEMP TABLE _eb AS
  SELECT * FROM site_specs
  WHERE site_id=(SELECT id FROM sites WHERE domain='noted.co.uk')
    AND aspect='evidence_base' AND is_current;
UPDATE site_specs SET is_current=false, superseded_at=now() WHERE id IN (SELECT id FROM _eb);
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT site_id, aspect,
       data || $payload$__PAYLOAD__$payload$::jsonb,
       'manual',
       'Privacy copy refreshed from the draft file (refresh_privacy_copy_from_draft.py). Derived row: bans and facts carried untouched.',
       true, pinned, 'owner-approved (claude session)'
FROM _eb;
COMMIT;
""".replace("__PAYLOAD__", payload)
    r = subprocess.run(PSQL, input=sql, capture_output=True, text=True, timeout=120)
    if r.returncode != 0:
        print("write failed:\n", r.stderr[:1500]); sys.exit(2)

    # Verify: bans intact, and a sentence DISTINCTIVE TO THIS EDIT landed.
    # Whitespace-normalised on BOTH sides: the draft is hard-wrapped, so any
    # multi-word probe can straddle a line break and read as absent when it is
    # there (the estate's false-absence class — this script hit it on its own
    # first run, 2026-08-25).
    probe = run("""SELECT jsonb_array_length(ss.data->'banned_claims'),
                          position('not in those backups' in regexp_replace(ss.data->>'writer_block', '\\s+', ' ', 'g')) > 0,
                          position('not in those backups' in regexp_replace(ss.data->'supplied_copy'->>'privacy', '\\s+', ' ', 'g')) > 0
                   FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                   WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print("verify (bans | wb has new sentence | supplied_copy has it):", probe)
    if probe != f"{bans}|t|t":
        print("FATAL: verification mismatch"); sys.exit(2)
    print("evidence_base refreshed from the draft.")
    return 0


if __name__ == "__main__":
    sys.exit(main())

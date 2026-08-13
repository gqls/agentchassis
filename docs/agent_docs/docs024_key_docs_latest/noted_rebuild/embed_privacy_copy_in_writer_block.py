#!/usr/bin/env python3
"""
Embed the owner-approved privacy copy INLINE in noted.co.uk's writer_block.

WHY. apply_privacy_copy.py (2026-08-12) registered the copy under
evidence_base.supplied_copy.privacy and told the writer, via writer_block, to use
it verbatim. Measured 2026-08-13: 0 of 22 sentences appeared — because the
page-content-writer's prompt template injects ONLY the writer_block STRING
({{.site_specs.specs.evidence_base.writer_block}}); supplied_copy never travels.
The instruction pointed at data outside the reader's context.

FIX. The copy travels inside writer_block itself. supplied_copy stays as the
canonical machine-readable store; writer_block carries the text to the prompt.
This is what makes the copy survive a future REGENERATION of the page (a
rerender merges content_data and is already safe; a regeneration re-writes it —
memory: bugfix 238).

Same discipline as apply_privacy_copy.py: copy extracted FROM THE DRAFT FILE,
new spec row DERIVED from the live one so the 7 bans and facts carry across.
Verifies bans intact and a distinctive copy sentence present in writer_block.
"""
import json
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).parent
DRAFT = HERE / "COPY_2026-08-12_privacy_DRAFT_for_owner.md"

# The exact clause apply_privacy_copy.py wrote on 2026-08-12 — replaced wholesale.
OLD_SENTINEL = "it is in this spec under `supplied_copy.privacy`"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-v", "ON_ERROR_STOP=1"]


def run(sql, tuples=True):
    cmd = PSQL + (["-tA"] if tuples else [])
    r = subprocess.run(cmd, input=sql, capture_output=True, text=True, timeout=120)
    if r.returncode != 0:
        print("psql failed:\n", r.stderr[:1200]); sys.exit(2)
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

    current = run("""SELECT ss.data->>'writer_block'
                     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                     WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    if OLD_SENTINEL not in current:
        if "THE APPROVED PRIVACY COPY, VERBATIM" in current:
            print("Already embedded — nothing to do."); return 0
        print("FATAL: expected clause not found; writer_block has changed — read it first."); sys.exit(2)

    head, tail = current.split(OLD_SENTINEL, 1)
    # drop the rest of the old sentence up to its full stop; keep everything after
    tail = tail.split(".", 1)[1] if "." in tail else ""

    new_block = (
        head
        + "it is reproduced IN FULL below — use it word for word on the privacy page "
        + "(hero subheadline = its first paragraph; body = the rest)."
        + tail.rstrip()
        + "\n\n## THE APPROVED PRIVACY COPY, VERBATIM (owner, 2026-08-12) — page title "
        + "“Your notes, and what happens to them”\n\n"
        + copy
        + "\n\n(End of approved copy. Do not paraphrase it, extend it, or write a rival "
        + "version anywhere on this site. If a page needs a privacy statement this does "
        + "not cover, that is a question for the owner.)"
    )

    payload = json.dumps({"writer_block": new_block}, ensure_ascii=False)
    sql = """
BEGIN;
CREATE TEMP TABLE _eb ON COMMIT DROP AS
  SELECT ss.* FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;
DO $do$ DECLARE n int; BEGIN
  SELECT count(*) INTO n FROM _eb;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base, found %', n; END IF;
END $do$;
UPDATE site_specs SET is_current=false, superseded_at=now() WHERE id IN (SELECT id FROM _eb);
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT site_id, aspect, data || $payload$__PAYLOAD__$payload$::jsonb, 'manual',
       'writer_block now carries the approved privacy copy INLINE (2026-08-13) - supplied_copy never reaches the writer prompt; only the writer_block string travels.',
       true, pinned, 'owner-approved (claude session)'
FROM _eb;
COMMIT;
""".replace("__PAYLOAD__", payload)
    print(run(sql, tuples=False))

    print("\nVERIFY")
    bans = run("""SELECT jsonb_array_length(ss.data->'banned_claims')
                  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print(f"  banned_claims: {bans} (must be 7)")
    probe = run("""SELECT (ss.data->>'writer_block') LIKE '%We will not replace this page quietly%'
                   FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                   WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print(f"  copy travels in writer_block: {probe}")
    kept = run("""SELECT length(ss.data->'supplied_copy'->'privacy'->>'body_markdown')
                  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
                  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
    print(f"  supplied_copy still present: {kept} chars")
    ok = bans == "7" and probe == "t" and kept not in ("", "0")
    print("\n" + ("EMBEDDED OK" if ok else "MISMATCH — investigate"))
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())

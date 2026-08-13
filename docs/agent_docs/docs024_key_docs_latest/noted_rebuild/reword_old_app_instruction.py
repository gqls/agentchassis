#!/usr/bin/env python3
"""
Reword the writer_block clause that MODELS the banned 'no server' phrase.

Owner ruling 2026-08-13: "we don't want to say no server" — the ban stays as
written; pages describe the old app by what it DID, never by what it lacked.

The defect this fixes: the writer's own instructions said "The old site had no
server at all", so a writer describing the old version in the migration guide
echoed its instructions' framing and was blocked by the validator. The gate and
the guidance disagreed; the writer obeyed the guidance.

Same discipline as the other two evidence_base scripts: derived row, bans
verified after, exact-sentinel replace that refuses if the text has moved.
"""
import json, subprocess, sys

OLD = ("Do not carry any wording forward from the old version of this site. "
       "The old site had no server at all and its privacy language was earned by that fact; "
       "the rebuilt product has a server, so the same sentences are now untrue.")
NEW = ("Do not carry any wording forward from the old version of this site: its privacy "
       "language was earned by an architecture this product does not have any more, so those "
       "sentences are untrue here. When a page needs to describe the old version — the "
       "migration guide does — describe it by what it DID: the old Noted kept everything in "
       "the browser you wrote it in, on that one device, and nowhere else. Never describe it "
       "by what it lacked: negative architecture phrasing about servers, cloud or backend is "
       "banned on every page of this site, and it stays banned even in a sentence that is "
       "true of the old version.")

PSQL = ["kubectl","-n","ai-persona-system","exec","-i","postgres-clients-0","--",
        "psql","-U","clients_user","-d","clients_db","-v","ON_ERROR_STOP=1"]

def run(sql, tuples=True):
    r = subprocess.run(PSQL + (["-tA"] if tuples else []), input=sql,
                       capture_output=True, text=True, timeout=120)
    if r.returncode != 0:
        print("psql failed:\n", r.stderr[:1200]); sys.exit(2)
    return r.stdout.strip()

cur = run("""SELECT ss.data->>'writer_block' FROM site_specs ss JOIN sites s ON s.id=ss.site_id
             WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
if OLD not in cur:
    if NEW[:60] in cur:
        print("Already reworded — nothing to do."); sys.exit(0)
    print("FATAL: expected clause not found verbatim — writer_block has moved; read it first."); sys.exit(2)

payload = json.dumps({"writer_block": cur.replace(OLD, NEW, 1)}, ensure_ascii=False)
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
SELECT site_id, aspect, data || $p$__PAYLOAD__$p$::jsonb, 'manual',
       'Owner ruling 2026-08-13: describe the old app by what it DID. The previous clause itself modelled the banned no-server phrase and the writer echoed it.',
       true, pinned, 'owner-approved (claude session)'
FROM _eb;
COMMIT;
""".replace("__PAYLOAD__", payload)
print(run(sql, tuples=False))

print("VERIFY")
print("  bans:", run("""SELECT jsonb_array_length(ss.data->'banned_claims') FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;"""))
print("  approved copy still inline:", run("""SELECT (ss.data->>'writer_block') LIKE '%We will not replace this page quietly%'
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;"""))
import re
new_block = run("""SELECT ss.data->>'writer_block' FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
hits = re.findall(r"(no|zero|without a)[ -]?(server|servers|cloud|backend)", new_block, re.I)
print("  banned-shape phrases still modelled in writer_block:", hits or "none")

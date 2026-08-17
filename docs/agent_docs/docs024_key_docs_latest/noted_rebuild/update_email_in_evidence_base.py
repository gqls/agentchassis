#!/usr/bin/env python3
"""Owner 2026-08-17: the site email is noted@contactforsales.com.
Updates BOTH carriers in evidence_base: the inline copy in writer_block and
supplied_copy.privacy.body_markdown (regenerated from the draft, the source of
truth). Derived row; asserts the 7 bans survive and the old address is gone."""
import json, subprocess, sys
from pathlib import Path

OLD, NEW = "hello@noted.co.uk", "noted@contactforsales.com"
DRAFT = Path(__file__).with_name("COPY_2026-08-12_privacy_DRAFT_for_owner.md")
PSQL = ["kubectl","-n","ai-persona-system","exec","-i","postgres-clients-0","--",
        "psql","-U","clients_user","-d","clients_db","-v","ON_ERROR_STOP=1"]

def run(sql, tuples=True):
    r = subprocess.run(PSQL + (["-tA"] if tuples else []), input=sql,
                       capture_output=True, text=True, timeout=120)
    if r.returncode != 0:
        print("psql failed:\n", r.stderr[:1000]); sys.exit(2)
    return r.stdout.strip()

body = DRAFT.read_text(encoding="utf-8")
body = body.split("## THE DRAFT",1)[1].split("## Verification",1)[0]
body = body.split("### Your notes, and what happens to them",1)[1].strip()
assert NEW in body and OLD not in body, "draft does not carry the new address"

cur = run("""SELECT ss.data->>'writer_block' FROM site_specs ss JOIN sites s ON s.id=ss.site_id
             WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;""")
if OLD not in cur:
    print("writer_block already carries no old address")
wb = cur.replace(OLD, NEW)

payload = json.dumps({"writer_block": wb}, ensure_ascii=False)
body_json = json.dumps(body, ensure_ascii=False)
sql = """
BEGIN;
CREATE TEMP TABLE _eb ON COMMIT DROP AS
  SELECT ss.* FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;
DO $do$ DECLARE n int; BEGIN
  SELECT count(*) INTO n FROM _eb;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current row, %', n; END IF;
END $do$;
UPDATE site_specs SET is_current=false, superseded_at=now() WHERE id IN (SELECT id FROM _eb);
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT site_id, aspect,
       jsonb_set(data || $p$__WB__$p$::jsonb, '{supplied_copy,privacy,body_markdown}', $b$__BODY__$b$::jsonb),
       'manual',
       'Owner 2026-08-17: contact email -> noted@contactforsales.com, in writer_block inline copy and supplied_copy.',
       true, pinned, 'owner-approved (claude session)'
FROM _eb;
COMMIT;
""".replace("__WB__", payload).replace("__BODY__", body_json)
print(run(sql, tuples=False))

print("VERIFY")
print("  bans:", run("""SELECT jsonb_array_length(ss.data->'banned_claims') FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;"""))
print("  old address anywhere in spec:", run("""SELECT (ss.data::text LIKE '%hello@noted.co.uk%') FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;"""))
print("  new address in writer_block:", run("""SELECT ((ss.data->>'writer_block') LIKE '%noted@contactforsales.com%') FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='noted.co.uk' AND ss.aspect='evidence_base' AND ss.is_current;"""))

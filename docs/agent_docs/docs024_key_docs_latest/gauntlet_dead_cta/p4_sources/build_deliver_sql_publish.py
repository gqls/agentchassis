#!/usr/bin/env python3
"""Build the guarded js_content delivery for publish-on-share (2026-07-31).

Sibling of build_deliver_sql.py, which records the EXCHANGE CARD delivery and is
left alone: its hardcoded guard and markers are the record of how that one was
done, and editing it would destroy that.

Two differences worth carrying forward:

  * the concurrency guard is READ LIVE at build time rather than pasted in. A
    hardcoded updated_at goes stale the moment anyone touches the row, and a
    stale guard fails CLOSED (0 rows) — safe, but it wastes a cycle and reads
    like a mystery. Reading it here means the guard is always the value this
    build actually saw.
  * the removed-marker checks name things this change genuinely deletes:
    `buildVerdictCard()` with no argument, and the one-line addEventListener.
    Both are gone by construction, so a survivor means the update did not apply.

base64 carries the payload so no dollar-quoting, backslash or unicode escape in
the JS can interact with SQL parsing.
"""
import base64
import hashlib
import pathlib
import subprocess
import sys

HERE = pathlib.Path(__file__).resolve().parent
JS = HERE / "gauntlet_js_2026-07-31b_publish_on_share.js"
CC = "5da50747-7936-4b8f-a66d-c1ea98919c75"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db", "-qAt", "-c"]

text = JS.read_text(encoding="utf-8")
b64 = base64.b64encode(text.encode("utf-8")).decode("ascii")
md5 = hashlib.md5(text.encode("utf-8")).hexdigest()

r = subprocess.run(PSQL + ["SELECT updated_at, md5(js_content), length(js_content) "
                           "FROM content_components WHERE id = '" + CC + "';"],
                   capture_output=True, text=True)
if r.returncode != 0:
    sys.exit("psql failed: " + r.stderr.strip())
guard, live_md5, live_len = r.stdout.strip().split("|")
print("live row : %s chars, md5 %s, updated %s" % (live_len, live_md5, guard))
print("new file : %d chars, md5 %s" % (len(text), md5))
if live_md5 == md5:
    sys.exit("nothing to do: the row already holds this file")

TEMPLATE = """BEGIN;

UPDATE content_components
   SET js_content = convert_from(decode('@B64@','base64'),'UTF8'),
       updated_at = NOW()
 WHERE id = '@CC@'
   AND updated_at = '@GUARD@';

DO $$
DECLARE v text;
BEGIN
  SELECT js_content INTO v FROM content_components WHERE id = '@CC@';
  IF v IS NULL THEN
    RAISE EXCEPTION 'row vanished';
  END IF;

  -- markers the change ADDED
  IF position('post("publish"' in v) = 0 THEN
    RAISE EXCEPTION 'UPDATE did not apply -- the publish call is absent';
  END IF;
  IF position('data-gi-share-note' in v) = 0 THEN
    RAISE EXCEPTION 'the consent note is absent';
  END IF;
  IF position('function emitCard(' in v) = 0 THEN
    RAISE EXCEPTION 'emitCard is absent';
  END IF;
  IF position('function buildVerdictCard(permalink)' in v) = 0 THEN
    RAISE EXCEPTION 'buildVerdictCard did not gain its permalink argument';
  END IF;

  -- markers the change REMOVED. "new marker absent" and "column blanked" look
  -- identical if you only ever check for what you added.
  IF position('function buildVerdictCard() {' in v) > 0 THEN
    RAISE EXCEPTION 'the no-argument buildVerdictCard survived';
  END IF;
  IF position('if (el.shareCard) el.shareCard.addEventListener' in v) > 0 THEN
    RAISE EXCEPTION 'the old one-line share wiring survived';
  END IF;

  -- an UNTOUCHED control: this delivery must not have disturbed the ledger
  IF position('vonc_gauntlet_ledger_v1' in v) = 0 THEN
    RAISE EXCEPTION 'the opinion ledger vanished -- this is not the right file';
  END IF;

  IF md5(v) <> '@MD5@' THEN
    RAISE EXCEPTION 'stored md5 % is not the verified file', md5(v);
  END IF;
  RAISE NOTICE 'delivered: % chars, md5 %', length(v), md5(v);
END $$;

COMMIT;
"""

sql = (TEMPLATE
       .replace("@B64@", b64)
       .replace("@CC@", CC)
       .replace("@GUARD@", guard)
       .replace("@MD5@", md5))

for tok in ("@B64@", "@CC@", "@GUARD@", "@MD5@"):
    assert tok not in sql, "placeholder %s unresolved" % tok

out = pathlib.Path(sys.argv[1]) if len(sys.argv) > 1 else HERE / "deliver_publish.sql"
out.write_text(sql, encoding="utf-8")
print("wrote %s (%d bytes); all placeholders resolved" % (out, len(sql)))

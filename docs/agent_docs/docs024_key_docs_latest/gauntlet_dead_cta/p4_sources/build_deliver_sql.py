#!/usr/bin/env python3
"""Build the guarded js_content delivery for the gauntlet component.

Two guards, both load-bearing:
  * WHERE updated_at = <the value read before building this> — if another
    session wrote the row in between, 0 rows update instead of clobbering them.
  * a DO block that RAISEs unless the new markers are present, the old ones are
    gone, and the stored md5 equals the file that was actually verified. Without
    it, "UPDATE 0" and a successful write look identical in psql's output.

The template is filled by literal replacement, not Python %-formatting: psql's
RAISE uses % as its own placeholder and the two fight.
"""
import base64
import hashlib
import pathlib

JS = pathlib.Path("/home/ant/projects/agentchassis/docs/agent_docs/"
                  "docs024_key_docs_latest/gauntlet_dead_cta/p4_sources/"
                  "gauntlet_js_2026-07-31_exchange_card.js")
CC = "5da50747-7936-4b8f-a66d-c1ea98919c75"
GUARD = "2026-07-29 08:45:18.450717+00"      # updated_at as read, pre-write

text = JS.read_text(encoding="utf-8")
b64 = base64.b64encode(text.encode("utf-8")).decode("ascii")
md5 = hashlib.md5(text.encode("utf-8")).hexdigest()

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
  IF position('VONC ASKED' in v) = 0 THEN
    RAISE EXCEPTION 'UPDATE did not apply -- another session wrote since my read';
  END IF;
  IF position('A JUDGED VERDICT' in v) > 0 THEN
    RAISE EXCEPTION 'old card header survived';
  END IF;
  IF position('function wrapLines(' in v) = 0 THEN
    RAISE EXCEPTION 'wrapLines missing';
  END IF;
  IF position('wrapText(' in v) > 0 THEN
    RAISE EXCEPTION 'old wrapText survived';
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
       .replace("@GUARD@", GUARD)
       .replace("@MD5@", md5))

out = pathlib.Path("/home/ant/vonc_mocks_tmp/deliver.sql")
out.write_text(sql, encoding="utf-8")
print("wrote %s (%d bytes)" % (out, len(sql)))
print("js chars %d  md5 %s" % (len(text), md5))
assert "@" not in sql.replace("@", "", 0) or True
for tok in ("@B64@", "@CC@", "@GUARD@", "@MD5@"):
    assert tok not in sql, "placeholder %s unresolved" % tok
print("all placeholders resolved")

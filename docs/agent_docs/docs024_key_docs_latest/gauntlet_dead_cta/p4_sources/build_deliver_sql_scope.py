#!/usr/bin/env python3
"""Build the guarded js_content delivery for RFC_020 §5.4 (the verdict-scope line).

Forked from build_deliver_sql.py, which is pinned to the 07-31 exchange-card
change: its source file, its `updated_at` guard and its markers are all that
change's, and re-running it would ship the OLD file over this one.

Three guards, all load-bearing:
  * WHERE updated_at = <read immediately before this runs, not hardcoded> — if
    another session wrote the row in between, 0 rows update instead of
    clobbering them. This tree has many concurrent sessions and the gauntlet
    component has been written by three of them this week.
  * a DO block that RAISEs unless the new markers are present, the old ones are
    gone, and the stored md5 equals the file that was actually verified.
    "UPDATE 0" and a successful write are indistinguishable in psql's output.
  * the NEGATIVE markers are strings this change genuinely deleted from the
    CODE and that do not appear in the file's own prose. deliver_record_component
    was bitten by the other kind: it checked for a colour the change stopped
    using, which still appeared in the correction note explaining why. A string
    can leave the code and stay in the commentary about the code, and then the
    check can never fire.

The template is filled by literal replacement, not Python %-formatting: psql's
RAISE uses % as its own placeholder and the two fight.
"""
import base64
import hashlib
import pathlib
import subprocess
import sys

JS = pathlib.Path("/home/ant/projects/agentchassis/docs/agent_docs/"
                  "docs024_key_docs_latest/gauntlet_dead_cta/p4_sources/"
                  "gauntlet_js_2026-08-10_verdict_scope.js")
CC = "5da50747-7936-4b8f-a66d-c1ea98919c75"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db", "-qAt"]

text = JS.read_text(encoding="utf-8")
b64 = base64.b64encode(text.encode("utf-8")).decode("ascii")
md5 = hashlib.md5(text.encode("utf-8")).hexdigest()

# Read the guard NOW rather than carrying one in the source. A stale hardcoded
# guard fails safe (0 rows) but costs a round trip to discover; worse, a guard
# copied from a previous delivery reads as deliberate and is not.
r = subprocess.run(PSQL + ["-c", "SELECT updated_at, md5(js_content), length(js_content) "
                                 "FROM content_components WHERE id = '" + CC + "';"],
                   capture_output=True, text=True)
if r.returncode != 0 or not r.stdout.strip():
    sys.exit("could not read the baseline: " + (r.stderr or "no row"))
guard, old_md5, old_len = r.stdout.strip().split("|")
print("baseline: %s chars md5 %s updated %s" % (old_len, old_md5, guard))
print("new file: %d chars md5 %s" % (len(text), md5))
if old_md5 == md5:
    sys.exit("nothing to do: the row already holds this file")

# Sanity: the markers must actually discriminate, or the DO block is theatre.
for m in ("var SCOPE =", "FOOT = 172", "#fde68a"):
    assert m in text, "positive marker absent from the new file: " + m
for m in ("var TOP = 112, FOOT = 130, USABLE = H - TOP - FOOT;",
          "x.fillRect(L, H - 112, 120, 6);"):
    assert m not in text, "negative marker still in the new file: " + m

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
  -- positive: the new line exists and is the escape form, not a literal dash
  IF position('var SCOPE =' in v) = 0 THEN
    RAISE EXCEPTION 'UPDATE did not apply -- another session wrote since my read';
  END IF;
  IF position('argued \\u2014 not whether it is true' in v) = 0 THEN
    RAISE EXCEPTION 'the scope string is not the escaped form that was verified';
  END IF;
  IF position('#fde68a' in v) = 0 THEN
    RAISE EXCEPTION 'the measured scope colour is absent';
  END IF;
  -- negative: the pre-5.4 footer geometry is gone from the CODE
  IF position('var TOP = 112, FOOT = 130, USABLE = H - TOP - FOOT;' in v) > 0 THEN
    RAISE EXCEPTION 'the old FOOT reserve survived -- the scope line will collide';
  END IF;
  IF position('x.fillRect(L, H - 112, 120, 6);' in v) > 0 THEN
    RAISE EXCEPTION 'the old rule-bar position survived';
  END IF;
  -- and nothing the 07-31 card change shipped may have been lost on the way
  IF position('VONC ASKED' in v) = 0 OR position('function wrapLines(' in v) = 0 THEN
    RAISE EXCEPTION 'the exchange card is missing -- this is not a descendant of the live file';
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

out = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else "/home/ant/deliver_scope.sql")
out.write_text(sql, encoding="utf-8")
for tok in ("@B64@", "@CC@", "@GUARD@", "@MD5@"):
    assert tok not in sql, "placeholder %s unresolved" % tok
print("wrote %s (%d bytes); all placeholders resolved" % (out, len(sql)))

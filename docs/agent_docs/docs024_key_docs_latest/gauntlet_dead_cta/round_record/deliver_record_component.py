#!/usr/bin/env python3
"""Deliver round_record_component.html to the live component AND the page copy.

    python3 deliver_record_component.py            # print the SQL
    python3 deliver_record_component.py --write /path/out.sql

Writes the SAME bytes to content_components.html_template and to
page_components.rendered_html. That is correct ONLY because this template has
zero variables — with variables, copying the template into the page's column
blanks the page's content (RUNBOOK §11). The DO block re-asserts it anyway.

CONCURRENCY. Both updates carry `WHERE updated_at = <the value just read>`, so
if another session writes the row between the read and the write, the update
matches nothing. `UPDATE 0` and a successful write are indistinguishable in
psql's output, which is why the DO block that follows is not optional: it
asserts the new content is actually in the row and the old content is not.
"""
import hashlib
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
TMPL = HERE / "round_record_component.html"
FN = "gauntlet-round-record"
SITE = "9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
TAG = "grrec"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db", "-qAt"]


def q(sql):
    r = subprocess.run(PSQL + ["-c", sql], capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit("psql failed: " + r.stderr.strip())
    return r.stdout.strip()


def main():
    html = TMPL.read_text()
    assert ("$" + TAG + "$") not in html, "dollar-quote tag collides with the template"
    assert "{{" not in html, "template placeholders present"
    new_md5 = hashlib.md5(html.encode()).hexdigest()

    # ── the three-way baseline. A delivery onto an unknown baseline is not a
    #    delivery: read the row before writing it, and keep updated_at.
    row = q("SELECT updated_at, md5(html_template), length(html_template) FROM "
            "content_components WHERE function = '" + FN + "';")
    if not row:
        sys.exit("no component row for " + FN + " — run 279 first")
    cc_updated, cc_md5, cc_len = row.split("|")

    prow = q("SELECT pc.updated_at, md5(pc.rendered_html), length(pc.rendered_html) "
             "FROM page_components pc JOIN pages p ON p.id = pc.page_id "
             "WHERE p.site_id = '" + SITE + "' AND p.name = '" + FN + "';")
    if not prow:
        sys.exit("no page_components row for " + FN)
    pc_updated, pc_md5, pc_len = prow.split("|")

    sys.stderr.write(
        "baseline  component: %s chars md5 %s  updated %s\n"
        "baseline  page copy: %s chars md5 %s  updated %s\n"
        "new file:            %s chars md5 %s\n"
        % (cc_len, cc_md5, cc_updated, pc_len, pc_md5, pc_updated, len(html), new_md5))

    if cc_md5 == new_md5 and pc_md5 == new_md5:
        sys.stderr.write("nothing to do: both columns already hold this file\n")
        return 0

    # TOKEN ORDER MATTERS. The HTML payload is substituted LAST, so a template
    # that happened to contain one of the short tokens (@FN@, @SITE@) cannot be
    # corrupted by a later pass. Tokens are @-delimited for the same reason:
    # a bare `FN` would also match inside `@NEWMD5@`-adjacent prose.
    sql = """\\set ON_ERROR_STOP on
BEGIN;

UPDATE content_components
   SET html_template = $TAG$@HTML@$TAG$, updated_at = now()
 WHERE function = '@FN@'
   AND updated_at = '@CCU@';

UPDATE page_components pc
   SET rendered_html = $TAG$@HTML@$TAG$, updated_at = now(), build_status = 'pending'
  FROM pages p
 WHERE p.id = pc.page_id
   AND p.site_id = '@SITE@'
   AND p.name = '@FN@'
   AND pc.updated_at = '@PCU@';

DO $assert$
DECLARE
  v_cc text; v_pc text;
BEGIN
  SELECT html_template INTO v_cc FROM content_components WHERE function = '@FN@';
  SELECT pc.rendered_html INTO v_pc
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = '@SITE@' AND p.name = '@FN@';

  IF md5(v_cc) <> '@NEWMD5@' THEN
    RAISE EXCEPTION 'component md5 is %, expected @NEWMD5@ (guard fired? another session wrote the row)', md5(v_cc);
  END IF;
  IF md5(v_pc) <> '@NEWMD5@' THEN
    RAISE EXCEPTION 'page copy md5 is %, expected @NEWMD5@', md5(v_pc);
  END IF;
  -- The md5 equality above already proves the new bytes landed AND that the
  -- old ones are gone — it is strictly stronger than any hand-picked marker.
  -- This second test exists only to tell the two failure modes apart, because
  -- they have opposite fixes: still at the OLD md5 means the concurrency guard
  -- fired and the row was never written; some THIRD md5 means another session
  -- wrote it between the read and the write.
  --
  -- (A literal "removed marker" check was tried here first and was wrong: it
  -- looked for a colour the change stopped USING, which still appears in the
  -- file's own correction note. A string can leave the code and stay in the
  -- prose about the code.)
  IF md5(v_cc) = '@OLDMD5@' THEN
    RAISE EXCEPTION 'row is still at the pre-delivery md5 — the updated_at guard fired, nothing was written';
  END IF;
  IF length(v_cc) < 8000 THEN
    RAISE EXCEPTION 'template is only % chars — a truncated write', length(v_cc);
  END IF;
  IF position('gauntlet-record-section' in v_cc) = 0 THEN
    RAISE EXCEPTION 'section class absent — this is not the record component';
  END IF;
  IF position('{' || '{' in v_cc) > 0 THEN
    RAISE EXCEPTION 'template placeholders present — copy-both-columns is unsafe';
  END IF;
  RAISE NOTICE 'delivered: % chars, md5 %, both columns identical', length(v_cc), md5(v_cc);
END
$assert$;

COMMIT;
""".replace("$TAG$", "$" + TAG + "$") \
   .replace("@NEWMD5@", new_md5) \
   .replace("@OLDMD5@", cc_md5) \
   .replace("@CCU@", cc_updated) \
   .replace("@PCU@", pc_updated) \
   .replace("@SITE@", SITE) \
   .replace("@FN@", FN) \
   .replace("@HTML@", html)          # payload LAST — see the note above

    assert "@" not in sql.replace(html, ""), "unsubstituted token left in the SQL"
    assert sql.count(html) == 2, "payload should appear exactly twice, got " + str(sql.count(html))

    if "--write" in sys.argv:
        out = Path(sys.argv[sys.argv.index("--write") + 1])
        out.write_text(sql)
        sys.stderr.write("wrote " + str(out) + "\n")
    else:
        sys.stdout.write(sql)
    return 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""b2_convert_oldshape.py — convert the last two OLD-SHAPE tool pages to B2, in place.

These two pages were already decomposed (prose-0 / tool-1 / prose-2), so b2_load.py's
route — delete the ported-page row, insert all blocks — structurally does not apply.
This converts the tool-1 row IN PLACE and touches nothing else. Two routes
(HANDOFF_2026-08-15 §3.3):

  loans-consolidation   component-backed: UPDATE component 'loans-consolidation'
                        (function KEPT — it is the acceptance-fence subject) with the
                        parameterised template + schema. The template's embedded
                        tool-doc comment moves to the component DESCRIPTION with its
                        lock paragraph corrected: a {{/* */}} Go comment would render
                        to nothing in Go but survive b2_load.rendered_tool's python
                        substitution, and python==Go render is a lane invariant that
                        b2_verify depends on — so the doc cannot stay in the template.
  mortgages-repayment   the locked row has a NULL component_id and 3,943 chars of
                        rendered_html directly on the instance: CREATE the component
                        (function='mortgages-repayment') and point the row at it.

Both: rendered_html BYTES ARE NOT TOUCHED (md5 asserted unchanged — b2_build proved
render==block with Go's own engine, so the first re-render reproduces them);
content_data becomes fields+provenance; the 08-05 'permanent' lock is CLEARED —
B2 protection is the template (writers reach fields only) + rebuild_policy='owned'
+ the acceptance fences, per the owner's 2026-08-13 editability direction and the
same conversion the other 21 calculator pages have already had.

Backups: page_components_bak_20260815_oldshape / content_components_bak_20260815_oldshape
(the 08-05 backup table is STALE for this era — HANDOFF_2026-08-15 §3.4).

Usage: python3 b2_convert_oldshape.py [--apply]   (default: plan + local checks only)
"""
import hashlib
import json
import os
import subprocess
import sys
import tempfile

S = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, S)
from b2_load import PSQL, SITE, q, rendered_tool, title_of
import b2_build

PC_BAK = "page_components_bak_20260815_oldshape"
CC_BAK = "content_components_bak_20260815_oldshape"

B2_DESCRIPTION = ("Parameterised calculator component (Track B2, bugs_open/263): panel, "
                  "ids and scripts live in this template and no content writer can touch "
                  "them; every clean copy span is a schema field. Page row deliberately "
                  "unlocked — ApplySectionEditAction refuses locked rows.")

LOCK_PARA_OLD = """   ROW IS PERMANENTLY LOCKED. `page_components` slot tool-1 for this page carries
   lock_type='permanent'; a rerender's computed output is discarded in favour of
   the stored bytes. So a rewrite of THIS template does not reach the live page
   until a human unlocks the row. That is intentional: the numbers here are
   consumer-credit figures."""
LOCK_PARA_NEW = """   ROW UNLOCKED 2026-08-15 (Track B2 conversion, bugs_open/263 — this paragraph
   used to record an 08-05 'permanent' lock and was corrected with the conversion).
   Protection now: the machinery lives in the html_template, which content edits
   cannot reach (ApplySectionEditAction writes schema fields only); the page is
   rebuild_policy='owned'; every clean copy span is a schema field. A rerender now
   DOES apply this template — its render was byte-proven equal to the previously
   locked bytes at conversion, with Go's own text/template. The consumer-credit
   arithmetic is exactly as protected as before; only the copy became editable."""

PAGES = ["loans-consolidation", "mortgages-repayment"]


def db(sql):
    r = subprocess.run(PSQL + ["-A", "-t", "-v", "ON_ERROR_STOP=1"], input=sql,
                       capture_output=True, text=True, timeout=180)
    if r.returncode != 0:
        raise SystemExit("db read failed: " + r.stderr.strip()[-300:])
    return r.stdout


def consolidation_description():
    """Current component's tool-doc, lock paragraph corrected, appended to the
    standard B2 description. Idempotent: if the doc is already in the description
    (a re-run after apply), keep what is there."""
    cur_desc = db("SELECT coalesce(description,'') FROM content_components "
                  "WHERE function='loans-consolidation';").rstrip("\n")
    if "=== tool-doc ===" in cur_desc:
        return cur_desc
    tmpl = db("SELECT html_template FROM content_components "
              "WHERE function='loans-consolidation';")
    i = tmpl.index("/* === tool-doc ===")
    j = tmpl.index("=== /tool-doc === */") + len("=== /tool-doc === */")
    doc = tmpl[i:j]
    assert LOCK_PARA_OLD in doc, "lock paragraph not found verbatim in tool-doc"
    return B2_DESCRIPTION + "\n\n" + doc.replace(LOCK_PARA_OLD, LOCK_PARA_NEW)


def page_sql(name, seed, row_md5, extra_desc=None):
    n_fields = len(seed["schema"]["fields"])
    cd = dict(seed["fields"])
    cd["_provenance"] = {
        "source": "trackb2_oldshape_20260815",
        "decomposed_from_pin": seed["pinned_ref"],
        "note": "Track B2 in-place conversion of an old-shape tool row (bugs_open/263): "
                "block bytes are the previously locked instance rendered_html; row "
                "deliberately UNLOCKED; machinery in the component html_template."}
    row_sel = ("(SELECT pc.id FROM page_components pc JOIN pages p ON p.id=pc.page_id "
               "WHERE p.site_id='%s' AND p.name='%s' AND pc.slot_name='tool-1')"
               % (SITE, name))
    lines = ["BEGIN;",
             "CREATE TABLE IF NOT EXISTS %s AS SELECT * FROM page_components WHERE false;" % PC_BAK,
             "CREATE TABLE IF NOT EXISTS %s AS SELECT * FROM content_components WHERE false;" % CC_BAK,
             "INSERT INTO %s SELECT pc.* FROM page_components pc WHERE pc.id=%s "
             "AND NOT EXISTS (SELECT 1 FROM %s b WHERE b.id=pc.id);" % (PC_BAK, row_sel, PC_BAK)]
    if name == "loans-consolidation":
        lines.append("INSERT INTO %s SELECT cc.* FROM content_components cc "
                     "WHERE cc.function='%s' AND NOT EXISTS "
                     "(SELECT 1 FROM %s b WHERE b.id=cc.id);" % (CC_BAK, name, CC_BAK))
        lines.append("""
UPDATE content_components SET html_template=%s, input_schema=%s::jsonb,
       description=%s, updated_at=now()
 WHERE function='%s';""" % (
            q("B2T", seed["template"]),
            q("B2S", json.dumps(seed["schema"], ensure_ascii=False)),
            q("B2D", extra_desc), name))
    else:
        lines.append("""
INSERT INTO content_components (name, description, html_template, input_schema, function, render_mode, display_name, category)
VALUES (%s, %s, %s, %s::jsonb, '%s', 'template', %s, 'calculators');""" % (
            q("B2T", "%s (loanandmortgagecalculator.co.uk)" % title_of(name)),
            q("B2D", B2_DESCRIPTION),
            q("B2T", seed["template"]),
            q("B2S", json.dumps(seed["schema"], ensure_ascii=False)),
            name, q("B2T", title_of(name))))
    lines.append("""
UPDATE page_components SET content_data=%s::jsonb,
       component_id=(SELECT id FROM content_components WHERE function='%s'),
       locked_at=NULL, locked_by=NULL, lock_type=NULL, lock_expires_at=NULL,
       updated_at=now()
 WHERE id=%s;""" % (q("B2C", json.dumps(cd, ensure_ascii=False)), name, row_sel))
    lines.append("""
DO $$
DECLARE bak int; ncc int; r_md5 text; n_locked int; n_fields int; r_cc uuid; cc_id uuid;
BEGIN
  SELECT count(*) INTO bak FROM %(pc_bak)s WHERE id=%(row_sel)s;
  IF bak <> 1 THEN RAISE EXCEPTION '%(name)s: tool row not in backup (%%)', bak; END IF;
  SELECT count(*) INTO ncc FROM content_components WHERE function='%(name)s' AND is_active;
  IF ncc <> 1 THEN RAISE EXCEPTION '%(name)s: component not unique: %%', ncc; END IF;
  SELECT md5(rendered_html), component_id INTO r_md5, r_cc FROM page_components WHERE id=%(row_sel)s;
  IF r_md5 <> '%(row_md5)s' THEN RAISE EXCEPTION '%(name)s: rendered_html MOVED: %%', r_md5; END IF;
  SELECT id INTO cc_id FROM content_components WHERE function='%(name)s';
  IF r_cc IS DISTINCT FROM cc_id THEN RAISE EXCEPTION '%(name)s: row not on component'; END IF;
  SELECT count(*) FILTER (WHERE locked_at IS NOT NULL OR lock_type IS NOT NULL) INTO n_locked
    FROM page_components WHERE page_id=(SELECT id FROM pages WHERE site_id='%(site)s' AND name='%(name)s');
  IF n_locked <> 0 THEN RAISE EXCEPTION '%(name)s: %% row(s) still locked', n_locked; END IF;
  SELECT count(*) INTO n_fields FROM content_components cc, jsonb_object_keys(cc.input_schema->'fields') k
   WHERE cc.function='%(name)s';
  IF n_fields <> %(n_fields)d THEN RAISE EXCEPTION '%(name)s: schema fields %%', n_fields; END IF;
  RAISE NOTICE '%(name)s converted: row bytes unchanged, 0 locked, %(n_fields)d fields';
END $$;
COMMIT;""" % {"pc_bak": PC_BAK, "row_sel": row_sel, "name": name, "site": SITE,
              "row_md5": row_md5, "n_fields": n_fields})
    return "\n".join(lines)


def main():
    apply = "--apply" in sys.argv
    # Go render re-proof at the moment of use, against the LIVE row bytes — the
    # seed's proof was against the same bytes, but re-asserting here means a stale
    # or edited seed cannot seed bytes the template does not reproduce.
    work = tempfile.mkdtemp(prefix="b2conv")
    open(os.path.join(work, "render.go"), "w").write(b2_build.GO_RUNNER)
    open(os.path.join(work, "go.mod"), "w").write("module b2render\n\ngo 1.23\n")
    failures = 0
    for name in PAGES:
        seed = json.load(open(os.path.join(S, "b2_seeds", name + ".json")))
        row_md5 = db("SELECT md5(pc.rendered_html) FROM page_components pc "
                     "JOIN pages p ON p.id=pc.page_id WHERE p.site_id='%s' "
                     "AND p.name='%s' AND pc.slot_name='tool-1';" % (SITE, name)).strip()
        rendered, err = b2_build.go_render(seed["template"], seed["fields"], work)
        if err:
            print("FAIL %-24s go render: %s" % (name, err)); failures += 1; continue
        go_md5 = hashlib.md5(rendered.encode()).hexdigest()
        py_md5 = hashlib.md5(rendered_tool(seed).encode()).hexdigest()
        if not (go_md5 == py_md5 == row_md5):
            print("FAIL %-24s md5 mismatch: go=%s py=%s row=%s"
                  % (name, go_md5[:12], py_md5[:12], row_md5[:12]))
            failures += 1
            continue
        extra = consolidation_description() if name == "loans-consolidation" else None
        sql = page_sql(name, seed, row_md5, extra)
        if not apply:
            print("plan %-24s go==py==row md5 %s  %d fields  %d chars SQL"
                  % (name, row_md5[:12], len(seed["fields"]), len(sql)))
            continue
        r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                           capture_output=True, text=True, timeout=180)
        last = [l for l in (r.stderr + r.stdout).splitlines() if l.strip()][-1:]
        if r.returncode != 0:
            failures += 1
            print("FAIL %-24s %s" % (name, (r.stderr or r.stdout).strip()[-300:]))
        else:
            print("ok   %-24s %s" % (name, last[0] if last else ""))
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())

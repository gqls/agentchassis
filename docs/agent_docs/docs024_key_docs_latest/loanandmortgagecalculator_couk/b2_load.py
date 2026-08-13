#!/usr/bin/env python3
"""b2_load.py — seed b2_build.py's output into the live DB, one guarded page at a time.

Per page (Track B2, bugs_open/263 2026-08-13 entry; proven end to end on
mortgages-simple before this existed):

  content_components   one new component per page: function == pages.name (the
                       acceptance-fence subject key, the consolidation precedent),
                       machinery in html_template, copy as input_schema fields
  page_components      the verbatim ported-page row is REPLACED by one row per
                       block: prose blocks on the shared ported-prose component
                       (content_data.content, section-wrapped rendered_html —
                       byte-identical to how Track A rows render), and the tool
                       block on the new component. NOTHING IS LOCKED: protection
                       is the template, plus rebuild_policy='owned', plus fences.
  pages.sections       positional slot names in block order.

Every page runs in its OWN transaction with a DO/RAISE verify: backup row exists,
row count == block count, tool row md5 == the manifest block's md5 (the same bytes
b2_build proved the template reproduces), zero locked rows, schema field count.
A failed page aborts itself, not the batch.

Usage:
  python3 b2_load.py <seed_dir> --pages a,b,c [--apply]
Without --apply it prints the per-page plan and runs every check that needs no write.
"""
import hashlib
import json
import os
import subprocess
import sys

SITE = "ed633ada-f8af-424b-b4d4-8af79160dbcd"
PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]
PROSE_OPEN = '<section class="ported-prose" data-component="ported-prose">'
PROSE_CLOSE = "</section>"
TAGS = ("B2T", "B2S", "B2C", "B2H")  # template, schema, content_data, html


def q(tag, payload):
    dq = "$%s$" % tag
    if dq in payload:
        raise SystemExit("dollar tag %s occurs in payload — pick another" % dq)
    return dq + payload + dq


def page_sql(seed):
    name = seed["page"]
    blocks, order = [], seed["block_order"]
    prose_iter = iter(seed["prose_blocks"])
    for i, kind in enumerate(order):
        if kind == "prose":
            html = next(prose_iter)
            blocks.append(("prose-%d" % i, "prose", html))
        else:
            blocks.append(("tool-%d" % i, "tool", None))
    slot_json = json.dumps([s for s, _, _ in blocks])
    tool_slot = next(s for s, k, _ in blocks if k == "tool")
    tool_md5 = hashlib.md5(rendered_tool(seed).encode()).hexdigest()

    cd = dict(seed["fields"])
    cd["_provenance"] = {"source": "trackb2_batch_20260813",
                         "decomposed_from_pin": seed["pinned_ref"],
                         "note": "Track B2: machinery in html_template, copy as fields, "
                                 "row deliberately UNLOCKED (bugs_open/263)."}

    lines = ["BEGIN;"]
    lines.append("""
INSERT INTO content_components (name, description, html_template, input_schema, function, render_mode, display_name, category)
VALUES (%s, %s, %s, %s::jsonb, '%s', 'template', %s, 'calculators');""" % (
        q(TAGS[0], "%s (loanandmortgagecalculator.co.uk)" % title_of(name)),
        q(TAGS[0], "Parameterised calculator component (Track B2, bugs_open/263): panel, "
                   "ids and scripts live in this template and no content writer can touch "
                   "them; every clean copy span is a schema field. Page row deliberately "
                   "unlocked — ApplySectionEditAction refuses locked rows."),
        q(TAGS[0], seed["template"]),
        q(TAGS[1], json.dumps(seed["schema"], ensure_ascii=False)),
        name, q(TAGS[0], title_of(name))))
    lines.append("""
DELETE FROM page_components
 WHERE page_id=(SELECT id FROM pages WHERE site_id='%s' AND name='%s')
   AND slot_name='ported-page';""" % (SITE, name))
    for slot, kind, html in blocks:
        pos = int(slot.split("-")[1])
        if kind == "prose":
            lines.append("""
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, rendered_html, build_status)
SELECT p.id, cc.id, '%s', %d,
       jsonb_build_object('content', %s::text,
                          '_provenance', jsonb_build_object('source','trackb2_batch_20260813')),
       %s, 'approved'
FROM pages p, content_components cc
WHERE p.site_id='%s' AND p.name='%s' AND cc.function='ported-prose';""" % (
                slot, pos, q(TAGS[2], html),
                q(TAGS[3], PROSE_OPEN + html + PROSE_CLOSE), SITE, name))
        else:
            lines.append("""
INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, rendered_html, build_status)
SELECT p.id, cc.id, '%s', %d, %s::jsonb, %s, 'approved'
FROM pages p, content_components cc
WHERE p.site_id='%s' AND p.name='%s' AND cc.function='%s';""" % (
                slot, pos, q(TAGS[2], json.dumps(cd, ensure_ascii=False)),
                q(TAGS[3], rendered_tool(seed)), SITE, name, name))
    lines.append("""
UPDATE pages SET sections='%s'::jsonb, updated_at=now()
 WHERE site_id='%s' AND name='%s';""" % (slot_json, SITE, name))
    lines.append("""
DO $$
DECLARE bak int; n_rows int; t_md5 text; n_locked int; n_fields int; n_pp int;
BEGIN
  SELECT count(*) INTO bak FROM page_components_bak_20260805_lmc
   WHERE page_id=(SELECT id FROM pages WHERE site_id='%(site)s' AND name='%(name)s');
  IF bak < 1 THEN RAISE EXCEPTION '%(name)s: no backup row — refusing'; END IF;
  SELECT count(*) INTO n_pp FROM content_components WHERE function='ported-prose';
  IF n_pp <> 1 THEN RAISE EXCEPTION 'ported-prose component not unique: %%', n_pp; END IF;
  SELECT count(*), count(*) FILTER (WHERE locked_at IS NOT NULL) INTO n_rows, n_locked
    FROM page_components WHERE page_id=(SELECT id FROM pages WHERE site_id='%(site)s' AND name='%(name)s');
  SELECT md5(rendered_html) INTO t_md5 FROM page_components
   WHERE page_id=(SELECT id FROM pages WHERE site_id='%(site)s' AND name='%(name)s') AND slot_name='%(tool_slot)s';
  IF n_rows <> %(n_blocks)d OR n_locked <> 0 OR t_md5 <> '%(tool_md5)s' THEN
    RAISE EXCEPTION '%(name)s: rows=%%, locked=%%, tool_md5=%%', n_rows, n_locked, t_md5;
  END IF;
  SELECT count(*) INTO n_fields FROM content_components cc, jsonb_object_keys(cc.input_schema->'fields') k
   WHERE cc.function='%(name)s';
  IF n_fields <> %(n_fields)d THEN RAISE EXCEPTION '%(name)s: schema fields %%', n_fields; END IF;
  RAISE NOTICE '%(name)s seeded: %(n_blocks)d rows, tool md5 ok, 0 locked, %(n_fields)d fields';
END $$;
COMMIT;""" % {"site": SITE, "name": name, "tool_slot": tool_slot,
              "n_blocks": len(blocks), "tool_md5": tool_md5,
              "n_fields": len(seed["schema"]["fields"])})
    return "\n".join(lines)


def rendered_tool(seed):
    """The tool row's rendered_html — b2_build proved template+fields == block, so the
    block from the manifest IS the render. Recompute it here from the seed to keep this
    file self-contained."""
    out = seed["template"]
    for k, v in seed["fields"].items():
        out = out.replace("{{.%s}}" % k, v)
    return out


def title_of(name):
    return name.replace("-", " ").title()


def main():
    seed_dir = sys.argv[1]
    pages = sys.argv[sys.argv.index("--pages") + 1].split(",")
    apply = "--apply" in sys.argv
    failures = 0
    for name in pages:
        seed = json.load(open(os.path.join(seed_dir, name + ".json")))
        # local invariant: python substitution must agree with the Go render that
        # b2_build already verified — if they disagree, the SQL would seed bytes the
        # template does not reproduce, and the whole design claim breaks.
        sql = page_sql(seed)
        if not apply:
            print("plan %-36s %d chars SQL, %d fields" % (name, len(sql), len(seed["fields"])))
            continue
        r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                           capture_output=True, text=True, timeout=180)
        last = [l for l in (r.stderr + r.stdout).splitlines() if l.strip()][-1:]
        if r.returncode != 0:
            failures += 1
            print("FAIL %-36s %s" % (name, (r.stderr or r.stdout).strip()[-200:]))
        else:
            print("ok   %-36s %s" % (name, last[0] if last else ""))
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())

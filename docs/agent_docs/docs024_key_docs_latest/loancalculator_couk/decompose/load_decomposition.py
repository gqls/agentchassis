#!/usr/bin/env python3
"""load_decomposition.py — replace a page's one verbatim row with its decomposed rows.

THIS IS THE STEP THAT CHANGES A LIVE PAGE. Everything before it was inert by
ordering: components with no page row render nowhere, chrome is never read for a
page that ships stored bytes. Writing here flips the page from
`loadVerbatimPageHTML` to `assemblePage` the moment it is next rendered.

THE FLIP IS THE ROW COUNT, NOT A FLAG, and that is worth knowing before you touch
this. A page ships verbatim when rebuild_policy='owned' AND it has EXACTLY ONE
component row AND that row carries content_data.deploy_mode='verbatim'. So
ADDING a row beside the verbatim one does not "add a section" — it silently
switches the page to assembly while leaving the old full document in the mix as
one of the sections, producing a document nested inside a document. The verbatim
row must be REPLACED, in the same transaction, which is what this does.

ORDER OF OPERATIONS, and each one is a precondition for the next:
  1. `ported-prose` exists in content_components (created here if absent).
  2. EVERY page's current rows are backed up — all 27, once, not just the page
     being written. A restore path that only covers the page you thought you
     were changing is not a restore path.
  3. One transaction per page: DELETE the old rows, INSERT the new, update
     pages.sections to the new slot names.
  4. The mirror's PREDICTION for that page is written to <work>/predicted/, so
     the real rendered output can be diffed against it afterwards. That diff is
     the scheduled test of assemble_mirror.py; without it the mirror is an
     unfalsified second implementation.

Usage:
  python3 load_decomposition.py --check  <page-name> [...]
  python3 load_decomposition.py --apply  <page-name> [...]
  python3 load_decomposition.py --apply  --all
  python3 load_decomposition.py --restore <page-name> [...]
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
REWRITE = os.path.join(LANE, "rewrite")
sys.path.insert(0, HERE)
sys.path.insert(0, LANE)
sys.path.insert(0, REWRITE)

from assemble_mirror import assemble_page, join_sections, rows_for_page  # noqa: E402

SITE_ID = "0162cde4-633e-45e9-8ca6-87a6b2fe1d26"
DOMAIN = "loancalculator.co.uk"
CHROME = os.path.join(LANE, "chrome")
BAK = "page_components_bak_20260802_decomp"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]

PROSE_FUNCTION = "ported-prose"
PROSE_TEMPLATE = ('<section class="ported-prose" data-component="ported-prose">'
                  '{{.content}}</section>')
PROSE_SCHEMA = {
    "fields": {
        "content": {
            "type": "html",
            "source": "authored",
            "required": True,
            "fallback": "",
            "llm_guidance": (
                "One block of the original hand-built page, lifted BYTE-FOR-BYTE by "
                "the decomposer. Safe to rewrite as ordinary editable prose. It "
                "deliberately contains no form control and no element addressed by "
                "any script on the page — anything a script touches travels inside "
                "the tool component instead, so rewriting this cannot break a "
                "calculator."),
        }
    }
}


def psql(sql):
    r = subprocess.run(PSQL + ["-tA", "-v", "ON_ERROR_STOP=1", "-c", sql],
                       capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:600])
    return r.stdout.strip()


def psql_stdin(sql):
    r = subprocess.run(PSQL + ["-v", "ON_ERROR_STOP=1"], input=sql,
                       capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:800])
    return r.stdout.strip()


def dollar_tag(*bodies):
    for i in range(2000):
        tag = "$dc%d$" % i
        if all(tag not in b for b in bodies):
            return tag
    raise RuntimeError("no free dollar-quote tag")


def ensure_prose_component():
    got = psql("SELECT id FROM content_components WHERE function='%s';" % PROSE_FUNCTION)
    if got.strip():
        return got.strip().splitlines()[0]
    schema_json = json.dumps(PROSE_SCHEMA, ensure_ascii=False)
    t = dollar_tag(PROSE_TEMPLATE, schema_json)
    psql_stdin(
        "INSERT INTO content_components (function, name, display_name, description, "
        "category, component_level, render_mode, created_from, is_active, "
        "html_template, input_schema) VALUES ("
        "{t}{fn}{t}, {t}Ported Prose Block{t}, {t}Ported Prose Block{t}, "
        "{t}One editable block of a hand-built page, produced by decomposing a "
        "verbatim-adopted document into prose plus tool components. Deliberately "
        "carries no form control and no script-addressed element.{t}, "
        "{t}general{t}, 'section', 'template', 'manual', true, "
        "{t}{h}{t}, {t}{s}{t}::jsonb);".format(
            t=t, fn=PROSE_FUNCTION, h=PROSE_TEMPLATE, s=schema_json))
    cid = psql("SELECT id FROM content_components WHERE function='%s';" % PROSE_FUNCTION)
    print("created component %s (%s)" % (PROSE_FUNCTION, cid))
    return cid


def backup_everything():
    """The PRE-DECOMPOSITION snapshot: all 27 pages, ONE generation each.

    Ported from the LMC lane 2026-08-11 (`bugs_open/250`), which hit both of
    these defects for real. They were present here too, and this lane's backup
    table was already carrying the damage.

    1. `SELECT pc.*` BROKE ON SCHEMA DRIFT. The table is cloned `LIKE
       page_components` on day one; page_components has since gained
       `rendered_html_digest`, so the star had 28 expressions for 27 target
       columns -> "INSERT has more expressions than target columns", and
       --apply could not run at all. Measured here 2026-08-11: 27 backup
       columns against 28 live. We now name the columns the backup table
       actually has (`pc.`-qualified: the source joins `pages`, so a bare `id`
       is ambiguous), so a future added column degrades to "not captured"
       instead of "nothing can be backed up".

       The asymmetry is worth keeping in mind: the RESTORE direction was never
       broken, because fewer expressions than target columns is legal SQL and
       the trailing ones take their defaults. **Drift breaks the backup loudly
       and the restore not at all** — so exercising only --restore finds
       nothing wrong while no new backup has been written since the column
       landed.

    2. IT RE-CAPTURED PAGES IT HAD ALREADY BACKED UP, ONE GENERATION LATER.
       The old guard was per-ROW (`NOT EXISTS ... b.id = pc.id`), so once a page
       was decomposed the NEXT --apply of ANY page swept that page's new prose
       rows in beside its original verbatim row. --restore then replays BOTH ->
       one page carrying a whole verbatim document AND a prose section, which is
       the nested-<html> corruption this lane's own docstring warns about,
       arriving through the rollback that is supposed to be the safety net.
       Measured here before the fix: **28 rows over 27 pages**, one page holding
       two generations. The guard is now per-PAGE: a page with any row here is
       already snapshotted and is never touched again.
    """
    # qualified for the SELECT side; plain for the INSERT target
    cols = psql("SELECT string_agg(quote_ident(column_name), ', ' "
                "ORDER BY ordinal_position) FROM information_schema.columns "
                "WHERE table_name='%s';" % BAK)
    sel = psql("SELECT string_agg('pc.' || quote_ident(column_name), ', ' "
               "ORDER BY ordinal_position) FROM information_schema.columns "
               "WHERE table_name='%s';" % BAK)
    psql_stdin("\n".join([
        "BEGIN;",
        "CREATE TABLE IF NOT EXISTS %s (LIKE page_components INCLUDING ALL);" % BAK,
        "INSERT INTO {bak} ({cols}) SELECT {sel} FROM page_components pc "
        "JOIN pages p ON p.id=pc.page_id WHERE p.site_id='{site}' "
        "AND NOT EXISTS (SELECT 1 FROM {bak} b WHERE b.page_id=pc.page_id);"
        .format(bak=BAK, cols=cols, sel=sel, site=SITE_ID),
        "COMMIT;",
    ]))
    print("backup table %s holds %s row(s) over %s page(s)"
          % (BAK, psql("SELECT count(*) FROM %s;" % BAK),
             psql("SELECT count(DISTINCT page_id) FROM %s;" % BAK)))


def component_ids():
    out = {}
    for line in psql("SELECT function, id FROM content_components "
                     "WHERE component_level IN ('tool','section') AND is_active;").splitlines():
        if "|" in line:
            fn, cid = line.split("|", 1)
            out[fn] = cid
    return out


def main():
    argv = sys.argv[1:]
    apply = "--apply" in argv
    restore = "--restore" in argv
    # --init is separate from --apply so that the two writes that are INERT —
    # creating the prose component and taking the backup — can be done and
    # inspected on their own, before the write that changes a live page. It also
    # unblocks --check, which cannot dry-run a page whose component does not
    # exist yet without weakening the very refusal that makes it useful.
    if "--init" in argv:
        ensure_prose_component()
        backup_everything()
        return 0
    if not (apply or restore or "--check" in argv):
        print(__doc__)
        return 2
    work = os.environ.get("DECOMP_WORK")
    if not work:
        sys.exit("set DECOMP_WORK (see RUNBOOK)")

    manifest = json.load(open(os.path.join(work, "manifest.json"), encoding="utf-8"))
    pages = {}
    for line in open(os.path.join(work, "pages.txt"), encoding="utf-8"):
        n, u = line.rstrip("\n").split("|")
        pages[n] = u
    ids = dict(psql("SELECT name, id FROM pages WHERE site_id='%s';" % SITE_ID
                    ).replace("|", "\t").splitlines() and
               [l.split("|", 1) for l in
                psql("SELECT name, id FROM pages WHERE site_id='%s';" % SITE_ID).splitlines()
                if "|" in l])

    names = [a for a in argv if not a.startswith("--")]
    if "--all" in argv:
        names = sorted(manifest)
    if not names:
        sys.exit("name at least one page, or pass --all")
    for n in names:
        if n not in manifest:
            sys.exit("no manifest entry for %r" % n)

    if restore:
        for n in names:
            psql_stdin("\n".join([
                "BEGIN;",
                "DELETE FROM page_components WHERE page_id='%s';" % ids[n],
                "INSERT INTO page_components SELECT * FROM %s WHERE page_id='%s';"
                % (BAK, ids[n]),
                "UPDATE pages SET sections='[\"ported-page\"]'::jsonb WHERE id='%s';"
                % ids[n],
                "COMMIT;",
            ]))
            print("restored %s from %s" % (n, BAK))
        return 0

    head = open(os.path.join(CHROME, "head.html"), encoding="utf-8").read().rstrip("\n")
    header = open(os.path.join(CHROME, "header.html"), encoding="utf-8").read().rstrip("\n")
    footer = open(os.path.join(CHROME, "footer.html"), encoding="utf-8").read().rstrip("\n")

    if apply:
        ensure_prose_component()
        backup_everything()
    comps = component_ids()
    prose_id = comps.get(PROSE_FUNCTION)

    os.makedirs(os.path.join(work, "predicted"), exist_ok=True)

    for name in names:
        page = manifest[name]
        url = pages[name]
        rows = rows_for_page(page, REWRITE)

        missing = [fn for _s, _h, fn, _cd in rows if fn and fn not in comps]
        if missing:
            print("REFUSE   %s: component(s) not in content_components: %s"
                  % (name, ", ".join(missing)))
            return 1
        if not prose_id and any(fn is None for _s, _h, fn, _cd in rows):
            print("REFUSE   %s: %s not loaded (run --apply, which creates it)"
                  % (name, PROSE_FUNCTION))
            return 1

        sections, dropped = join_sections([(s, h) for s, h, _f, _c in rows])
        if dropped:
            print("REFUSE   %s: %d section(s) would be dropped by assembly: %s"
                  % (name, len(dropped), dropped))
            return 1

        predicted = assemble_page(head, header, footer, sections, DOMAIN, url,
                                  page["title"], page["meta_desc"])
        open(os.path.join(work, "predicted", name + ".html"), "w",
             encoding="utf-8").write(predicted)

        print("%-32s %d row(s) -> %s  (predicted %d bytes)"
              % (name, len(rows), url, len(predicted)))
        for i, (slot, html, fn, _cd) in enumerate(rows):
            print("    %d %-10s %-24s %6d b" % (i, slot, fn or PROSE_FUNCTION, len(html)))
        if not apply:
            continue

        stmts = ["BEGIN;", "DELETE FROM page_components WHERE page_id='%s';" % ids[name]]
        for i, (slot, html, fn, cd) in enumerate(rows):
            cid = comps[fn] if fn else prose_id
            # Provenance is NAMESPACED under _provenance, unlike the ported-page
            # rows it replaces, which put schema/source/generator at the top
            # level. A tool row's content_data IS the render context, so a bare
            # `source` or `schema` key there is one component field name away
            # from silently overriding a real field.
            data = dict(cd)
            data["_provenance"] = {
                "generator": "loancalculator_couk/decompose/1",
                "source_url": "https://%s%s" % (DOMAIN, url),
                "source": ("decomposed from the adopted verbatim document; prose "
                           "lifted byte-for-byte, tool replaced by a component "
                           "proven numerically identical across three vectors"),
                "replaced": "ported-page (deploy_mode=verbatim)",
            }
            cd_json = json.dumps(data, ensure_ascii=False)
            t = dollar_tag(html, cd_json)
            stmts.append(
                "INSERT INTO page_components (page_id, component_id, position, "
                "slot_name, rendered_html, content_data, build_status) VALUES ("
                "'{p}', '{c}', {i}, {t}{s}{t}, {t}{h}{t}, {t}{d}{t}::jsonb, "
                "'approved');".format(p=ids[name], c=cid, i=i, s=slot, h=html,
                                      d=cd_json, t=t))
        slots_json = json.dumps([s for s, _h, _f, _c in rows])
        t = dollar_tag(slots_json)
        stmts.append("UPDATE pages SET sections={t}{s}{t}::jsonb, updated_at=now() "
                     "WHERE id='{p}';".format(t=t, s=slots_json, p=ids[name]))
        stmts.append("COMMIT;")
        psql_stdin("\n".join(stmts))

        back = psql("SELECT count(*), sum(octet_length(rendered_html)) "
                    "FROM page_components WHERE page_id='%s';" % ids[name])
        print("    written: %s (rows|bytes)" % back)

    if apply:
        print("\nThese page(s) now ASSEMBLE rather than shipping stored bytes.")
        print("Nothing is deployed until a page_rerender runs — see the RUNBOOK.")
        print("Diff the real output against %s/predicted/ before moving the rest."
              % work)
    return 0


if __name__ == "__main__":
    sys.exit(main())

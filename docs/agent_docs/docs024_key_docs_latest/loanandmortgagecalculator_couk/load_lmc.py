#!/usr/bin/env python3
"""load_lmc.py — replace a page's one verbatim row with its decomposed rows.

THIS IS THE STEP THAT CHANGES A LIVE PAGE. Adapted from the sibling's
load_decomposition.py — read its docstring; the transaction shape (DELETE the
verbatim row + INSERT the new rows + update pages.sections, one txn per page)
is theirs, and exists because ADDING a row beside a verbatim one silently
flips the page to assembly with the whole stored document as one section
(nested <html>). Differences here:

  * TOOL ROWS HAVE NO COMPONENT. These widgets ship as-is (cards + inline
    scripts + the calculators.js tag), so component_id is NULL — a
    rerender_sections resolve pass finds nothing to re-render and carries the
    row — and every tool row is born LOCKED (permanent). Do NOT fire
    section_data_resolved at a locked row: bugs_open/189 duplicates it on the
    page (measured on the sibling's site).
  * PROSE ROWS reuse the sibling's `ported-prose` content_component
    (function-global). rendered_html is exactly template % content, so a
    future re-render from content_data reproduces the stored bytes.
  * TIGHT PAGES (13 guides + legal): the prose content is wrapped in
    <div class="container container-tight"> INSIDE the section — the head
    chrome's shim gives it the original 740px content box, and the wrapper
    living in content_data (not in the section template) keeps re-renders
    byte-stable.
  * PRE-WRITE GUARD: every targeted page's stored row must still match the
    2026-08-05 baseline md5 (acceptance/BASELINE_2026-08-05_*.txt). A moved
    row means another session wrote the page since — stop and look.

The mirror PREDICTION for each written page goes to <work>/predicted/; the
first real render is diffed against it before anything else moves (the
mirror is a hypothesis with a scheduled test, not an authority — sibling's
words, kept).

⛔ **OWNER RULING 2026-08-06: COPY IS THE FRAMEWORK'S JOB, NOT A CLI SESSION'S.**
This tool ships the DECOMPOSITION — the structural change that turns a frozen
document into rows a writer agent can reach. It must be run against
`manifest.json` (the original copy, byte-for-byte), NOT against a manifest
carrying hand-authored rewrites. `--manifest` therefore defaults to
`manifest.json`. The `voice_overlays/` corpus and `manifest_voiced.json` are
SUPERSEDED by this ruling and kept only as evidence of what the register
looks like; see the PLAN's correction block. Passing
`--manifest manifest_voiced.json` would re-introduce exactly what the ruling
forbids, and there is no reason to.

Usage:
  DECOMP_WORK=<dir> python3 load_lmc.py --check  <name> [...]
  DECOMP_WORK=<dir> python3 load_lmc.py --apply  <name> [...]
  DECOMP_WORK=<dir> python3 load_lmc.py --apply  --all
  DECOMP_WORK=<dir> python3 load_lmc.py --restore <name> [...]
  python3 load_lmc.py --init          # backup + component lookup only, inert
  [--manifest <filename>]             # default manifest.json
"""
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
SIBLING = os.path.join(os.path.dirname(HERE), "loancalculator_couk")
sys.path.insert(0, os.path.join(SIBLING, "decompose"))

from assemble_mirror import assemble_page, join_sections  # noqa: E402


def inject_canonical(page_html, domain, url):
    """Mirror of injectCanonicalLink (rerender_single_page_action.go:973),
    which the sibling's assemble_mirror predates — it shipped with seam 2
    (9c7a8e9e4, live v1.0.1238) and runs AFTER the JSON-LD injection, so the
    tag lands immediately before </head>. Without this the first real render
    would differ from every prediction by exactly one line and the diff would
    blame the decomposition."""
    if (not page_html or 'rel="canonical"' in page_html
            or "rel='canonical'" in page_html):
        return page_html
    if not domain or not url.startswith("/") or set(url) & set("#?"):
        return page_html
    block = '\n<link rel="canonical" href="https://%s%s">\n' % (domain, url)
    idx = page_html.find("</head>")
    if idx >= 0:
        return page_html[:idx] + block + page_html[idx:]
    return page_html

SITE_ID = "ed633ada-f8af-424b-b4d4-8af79160dbcd"
DOMAIN = "loanandmortgagecalculator.co.uk"
CHROME = os.path.join(HERE, "chrome")
BAK = "page_components_bak_20260805_lmc"
# RE-BASELINED 2026-08-09 (bugfix-224 session). The 08-05 file is KEPT, not
# replaced: it is the record of what the pages were before the arithmetic fixes.
# 17 of 41 pages moved between the two, and every one is accounted for — 15 by
# the 0%-rate fix (bugs_open/224), the button ids and the SDLT fix
# (bugs_open/225), and 2 by the voice lane's decomposition (consolidation, now
# 3 rows, and one guide). NOTHING moved unexplained, which is the only thing
# that makes a re-baseline safe: this guard exists to catch another session
# having written a page, so regenerating it blindly would absorb exactly what
# it is meant to surface. If you re-baseline again, diff first and account for
# every moved row.
BASELINE = os.path.join(HERE, "acceptance",
                        "BASELINE_2026-08-09_stored_md5_at_b26fdc81b.txt")
PROSE_FUNCTION = "ported-prose"
PROSE_TEMPLATE = ('<section class="ported-prose" data-component="ported-prose">'
                  '%s</section>')
TIGHT_WRAP = '<div class="container container-tight">\n%s\n</div>'
LOCKED_BY = "lmc_decompose_voice_20260805"

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
        "--", "psql", "-U", "clients_user", "-d", "clients_db"]


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
        tag = "$lm%d$" % i
        if all(tag not in b for b in bodies):
            return tag
    raise RuntimeError("no free dollar-quote tag")


def manifest_name(argv):
    """Default manifest.json — the ORIGINAL copy. See the owner ruling at the
    top of this file: a CLI session ships structure, the framework ships
    words."""
    if "--manifest" in argv:
        return argv[argv.index("--manifest") + 1]
    return "manifest.json"


def backup_everything():
    psql_stdin("\n".join([
        "BEGIN;",
        "CREATE TABLE IF NOT EXISTS %s (LIKE page_components INCLUDING ALL);" % BAK,
        "INSERT INTO %s SELECT pc.* FROM page_components pc "
        "JOIN pages p ON p.id=pc.page_id WHERE p.site_id='%s' "
        "AND NOT EXISTS (SELECT 1 FROM %s b WHERE b.id=pc.id);" % (BAK, SITE_ID, BAK),
        "COMMIT;",
    ]))
    print("backup table %s holds %s row(s)" % (BAK, psql("SELECT count(*) FROM %s;" % BAK)))


def page_ids():
    """Keyed by URL, not name: adoption's verbatimPageIdentity names
    (guide-jargon-buster) differ from this lane's manifest slugs
    (guides-jargon-buster), and the URL is the identity both sides share.
    A bare-name join here was a real defect caught at --check design time.

    TAB field separator, not psql's default pipe: every title on this site
    contains ' | LoanAndMortgageCalculator.co.uk', so the default delimiter
    splits inside the data — the first --check run reported a title
    'truncated by adoption' that was in fact truncated by THIS parser."""
    r = subprocess.run(PSQL + ["-tA", "-F", "\t", "-v", "ON_ERROR_STOP=1", "-c",
                               "SELECT url, id, name, title, meta_description "
                               "FROM pages WHERE site_id='%s';" % SITE_ID],
                       capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError((r.stderr or r.stdout).strip()[:600])
    out = {}
    for line in r.stdout.splitlines():
        parts = line.split("\t")
        if len(parts) >= 5:
            out[parts[0]] = {"id": parts[1], "name": parts[2],
                             "title": parts[3], "meta_desc": parts[4]}
    return out


def baseline_md5s():
    out = {}
    for line in open(BASELINE, encoding="utf-8"):
        url, md5, _len = line.rstrip("\n").split("|")
        out[url] = md5
    return out


def rows_for_page(page):
    """(slot, rendered_html, is_tool, content_data) in position order."""
    rows = []
    for i, b in enumerate(page["blocks"]):
        if b["kind"] == "prose":
            content = b["html"]
            if page["tight"]:
                content = TIGHT_WRAP % content
            rows.append(("prose-%d" % i, PROSE_TEMPLATE % content, False,
                         {"content": content}))
        else:
            html = b["html"]
            if b.get("scripts"):
                html = html + "\n" + b["scripts"]
            rows.append(("tool-%d" % i, html, True, {}))
    return rows


def main():
    argv = sys.argv[1:]
    apply = "--apply" in argv
    restore = "--restore" in argv
    ids = page_ids()

    if "--init" in argv:
        got = psql("SELECT id FROM content_components WHERE function='%s';"
                   % PROSE_FUNCTION)
        if not got:
            raise SystemExit("content_components has no %r — the sibling's "
                             "load_decomposition.py --init creates it; run that "
                             "first (it is function-global)" % PROSE_FUNCTION)
        print("%s component: %s" % (PROSE_FUNCTION, got))
        backup_everything()
        return 0

    if restore:
        names = [a for a in argv if not a.startswith("--")]
        work = os.environ.get("DECOMP_WORK")
        if not work:
            sys.exit("set DECOMP_WORK (restore resolves names via the manifest)")
        manifest = json.load(open(os.path.join(work, manifest_name(argv)),
                                  encoding="utf-8"))["pages"]
        for n in names:
            pid = ids[manifest[n]["url"]]["id"]
            psql_stdin("\n".join([
                "BEGIN;",
                "DELETE FROM page_components WHERE page_id='%s';" % pid,
                "INSERT INTO page_components SELECT * FROM %s WHERE page_id='%s';"
                % (BAK, pid),
                "UPDATE pages SET sections='[\"ported-page\"]'::jsonb, "
                "updated_at=now() WHERE id='%s';" % pid,
                "COMMIT;",
            ]))
            print("restored %s from %s" % (n, BAK))
        return 0

    if not (apply or "--check" in argv):
        print(__doc__)
        return 2

    work = os.environ.get("DECOMP_WORK")
    if not work:
        sys.exit("set DECOMP_WORK to the dir holding the manifest")
    manifest = json.load(open(os.path.join(work, manifest_name(argv)),
                              encoding="utf-8"))["pages"]

    names = [a for a in argv if not a.startswith("--")]
    if "--all" in argv:
        names = sorted(manifest)
    if not names:
        sys.exit("name at least one page, or pass --all")
    for n in names:
        if n not in manifest:
            sys.exit("no manifest entry for %r" % n)
        if manifest[n]["url"] not in ids:
            sys.exit("no pages row with url %r" % manifest[n]["url"])

    base = baseline_md5s()
    prose_id = psql("SELECT id FROM content_components WHERE function='%s';"
                    % PROSE_FUNCTION)
    if not prose_id:
        sys.exit("ported-prose component missing — run --init")

    head = open(os.path.join(CHROME, "head.html"), encoding="utf-8").read().rstrip("\n")
    header = open(os.path.join(CHROME, "header.html"), encoding="utf-8").read().rstrip("\n")
    footer = open(os.path.join(CHROME, "footer.html"), encoding="utf-8").read().rstrip("\n")

    if apply:
        backup_everything()
    os.makedirs(os.path.join(work, "predicted"), exist_ok=True)

    for name in names:
        page = manifest[name]
        db = ids[page["url"]]
        # assembly splices the DB title/desc; the manifest carries the
        # original head's. They must agree or the assembled page silently
        # changes title.
        for k, label in (("title", "title"), ("meta_desc", "meta description")):
            if (page[k] or "") != (db[k] or ""):
                print("REFUSE  %s: %s differs between manifest and pages row:\n"
                      "  manifest: %r\n  db:       %r"
                      % (name, label, page[k], db[k]))
                return 1
        # the stored row must still be the baseline bytes
        cur = psql("SELECT count(*), min(md5(rendered_html)) FROM page_components "
                   "WHERE page_id='%s';" % db["id"])
        cnt, md5 = cur.split("|")
        if cnt != "1" or md5 != base[page["url"]]:
            print("REFUSE  %s: stored state moved since baseline (rows=%s md5=%s)"
                  % (name, cnt, md5))
            return 1

        rows = rows_for_page(page)
        sections, dropped = join_sections([(s, h) for s, h, _t, _c in rows])
        if dropped:
            print("REFUSE  %s: %d section(s) would be dropped by assembly: %s"
                  % (name, len(dropped), dropped))
            return 1

        predicted = assemble_page(head, header, footer, sections, DOMAIN,
                                  page["url"], page["title"], page["meta_desc"])
        predicted = inject_canonical(predicted, DOMAIN, page["url"])
        open(os.path.join(work, "predicted", name + ".html"), "w",
             encoding="utf-8").write(predicted)

        print("%-40s %d row(s) -> %s  (predicted %d bytes)"
              % (name, len(rows), page["url"], len(predicted)))
        for i, (slot, html, is_tool, _cd) in enumerate(rows):
            print("    %d %-8s %s %6d b" % (i, slot,
                                            "LOCKED" if is_tool else "      ",
                                            len(html)))
        if not apply:
            continue

        stmts = ["BEGIN;",
                 "DELETE FROM page_components WHERE page_id='%s';" % db["id"]]
        for i, (slot, html, is_tool, cd) in enumerate(rows):
            data = dict(cd)
            data["_provenance"] = {
                "generator": "loanandmortgagecalculator_couk/decompose/1",
                "source_url": "https://%s%s" % (DOMAIN, page["url"]),
                "source": ("decomposed from the adopted verbatim document at "
                           "sites-repo b318a8fad; prose re-voiced per the "
                           "owner-approved gentle-explanatory register "
                           "(content_direction v2, 2026-08-05); tool markup "
                           "and scripts byte-original"),
                "replaced": "ported-page (deploy_mode=verbatim)",
            }
            cd_json = json.dumps(data, ensure_ascii=False)
            t = dollar_tag(html, cd_json)
            if is_tool:
                stmts.append(
                    "INSERT INTO page_components (page_id, position, slot_name, "
                    "rendered_html, content_data, build_status, locked_at, "
                    "locked_by, lock_type) VALUES ('{p}', {i}, {t}{s}{t}, "
                    "{t}{h}{t}, {t}{d}{t}::jsonb, 'approved', now(), '{lb}', "
                    "'permanent');".format(p=db["id"], i=i, s=slot, h=html,
                                           d=cd_json, t=t, lb=LOCKED_BY))
            else:
                stmts.append(
                    "INSERT INTO page_components (page_id, component_id, "
                    "position, slot_name, rendered_html, content_data, "
                    "build_status) VALUES ('{p}', '{c}', {i}, {t}{s}{t}, "
                    "{t}{h}{t}, {t}{d}{t}::jsonb, 'approved');".format(
                        p=db["id"], c=prose_id, i=i, s=slot, h=html,
                        d=cd_json, t=t))
        slots_json = json.dumps([s for s, _h, _t, _c in rows])
        t = dollar_tag(slots_json)
        stmts.append("UPDATE pages SET sections={t}{s}{t}::jsonb, updated_at=now() "
                     "WHERE id='{p}';".format(t=t, s=slots_json, p=db["id"]))
        stmts.append("COMMIT;")
        psql_stdin("\n".join(stmts))
        back = psql("SELECT count(*), count(*) FILTER (WHERE locked_at IS NOT NULL), "
                    "sum(octet_length(rendered_html)) "
                    "FROM page_components WHERE page_id='%s';" % db["id"])
        print("    written: %s (rows|locked|bytes)" % back)

    if apply:
        print("\nThese page(s) now ASSEMBLE. Nothing deploys until a "
              "page_rerender runs (file it status='triaged', page_id in spec "
              "AND column, NO spec.reason). Diff the real output against "
              "%s/predicted/ before moving the rest." % work)
    return 0


if __name__ == "__main__":
    sys.exit(main())

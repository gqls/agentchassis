#!/usr/bin/env python3
"""toolsource.py — pull a ported tool's section out of the database, and put a
repaired one back.

WHY THIS EXISTS. For a ported page, `page_components.rendered_html` IS the
artefact: content_data holds only the port's provenance, so there is nothing to
re-render from. A repair that only reaches the deployed site repo therefore
looks complete — the live page is fixed and re-probes OK — while the database
still holds the broken copy and any rebuild silently restores it. That happened
to repair #1 of this workstream and was caught by the orphan_element_refs
detector, not by re-reading the work. So: the loop ends at the DATABASE.

    toolsource.py pull <page-name> [outfile]     e.g. pull tool-animated-favicon
    toolsource.py push <page-name> <infile>      writes it back, with assertions
    toolsource.py check <file>                   orphan refs in a file, no DB

`push` REFUSES rather than warns when the payload would leave the page in the
state this workstream exists to fix. It checks, before writing:

  * the section's tags balance and it is a single <section> element;
  * zero orphan element references remain (the same predicate the chassis
    check now applies, so a repair cannot pass here and fail there);
  * the payload is not drastically shorter than what is already stored — a
    truncated LLM completion looks exactly like a successful small edit
    (bugs_open/012), so shrinking by more than a third needs SHRINK_OK=1.

The HTML travels as base64 in both directions. The sections carry quotes,
backslashes and JS template literals, and hand-quoting them into SQL is how a
generator in this workstream corrupted itself once already. Base64 has nothing
in it to escape.
"""
import base64
import os
import re
import subprocess
import sys

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db", "-tAc"]
DOMAIN = "webdesign.co.uk"

REF_BYID = re.compile(r"""getElementById\(\s*['"]([A-Za-z0-9_\-:.]+)['"]\s*\)""")
REF_QSEL = re.compile(r"""querySelector(?:All)?\(\s*['"]#([A-Za-z0-9_\-]+)['"]\s*\)""")
PRESENT = re.compile(r"""\bid\s*=\s*["']([^"']+)["']""")
DYNAMIC = re.compile(r"""(?:\.id\s*=|setAttribute\(\s*["']id["']\s*,)\s*["']([A-Za-z0-9_\-:.]+)["']""")


def orphan_refs(html):
    """Mirror of datahelpers.OrphanElementRefs. Keep the two in step."""
    referenced = set(REF_BYID.findall(html)) | set(REF_QSEL.findall(html))
    if not referenced:
        return []
    present = set(PRESENT.findall(html)) | set(DYNAMIC.findall(html))
    return sorted(referenced - present)


def psql(sql):
    out = subprocess.run(PSQL + [sql], capture_output=True, text=True)
    if out.returncode != 0:
        sys.exit("psql failed: " + out.stderr.strip())
    return out.stdout.strip()


def pull(page, outfile=None):
    b64 = psql(
        "SELECT replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'),E'\\n','') "
        "FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id "
        "WHERE s.domain='%s' AND p.name='%s';" % (DOMAIN, page))
    if not b64:
        sys.exit("no page_component found for %s" % page)
    html = base64.b64decode(b64).decode("utf-8")
    path = outfile or (page + ".html")
    with open(path, "w", encoding="utf-8") as fh:
        fh.write(html)
    missing = orphan_refs(html)
    print("%s -> %s (%d bytes)" % (page, path, len(html)))
    print("orphan refs: %s" % (", ".join(missing) if missing else "none"))
    return html


def push(page, infile):
    with open(infile, encoding="utf-8") as fh:
        html = fh.read()

    if html.count("<section") != html.count("</section>"):
        sys.exit("REFUSED: unbalanced <section> tags")
    if not html.lstrip().startswith("<section"):
        sys.exit("REFUSED: payload does not start with <section> — push the SECTION, not the page")
    missing = orphan_refs(html)
    if missing:
        sys.exit("REFUSED: %d orphan element reference(s) remain: %s"
                 % (len(missing), ", ".join(missing)))

    current = psql(
        "SELECT length(pc.rendered_html) FROM page_components pc JOIN pages p ON p.id=pc.page_id "
        "JOIN sites s ON s.id=p.site_id WHERE s.domain='%s' AND p.name='%s';" % (DOMAIN, page))
    if not current:
        sys.exit("no page_component found for %s" % page)
    if len(html) < int(current) * 2 // 3 and not os.environ.get("SHRINK_OK"):
        sys.exit("REFUSED: payload is %d bytes against %s stored — a truncated completion looks "
                 "exactly like this. Re-check, then SHRINK_OK=1 if it is deliberate."
                 % (len(html), current))

    b64 = base64.b64encode(html.encode("utf-8")).decode("ascii")
    sql = (
        "BEGIN;"
        "UPDATE page_components pc SET rendered_html = convert_from(decode('%s','base64'),'UTF8'),"
        " content_data = pc.content_data || jsonb_build_object('repair', jsonb_build_object("
        "  'at', to_char(NOW(),'YYYY-MM-DD\"T\"HH24:MI:SSZ'),"
        "  'by', 'webdesign tools repair, per-tool loop',"
        "  'diverges_from_source', true)),"
        " updated_at = NOW()"
        " FROM pages p, sites s WHERE pc.page_id=p.id AND p.site_id=s.id"
        "   AND s.domain='%s' AND p.name='%s';"
        "COMMIT;" % (b64, DOMAIN, page))
    psql(sql)

    after = psql(
        "SELECT length(pc.rendered_html) FROM page_components pc JOIN pages p ON p.id=pc.page_id "
        "JOIN sites s ON s.id=p.site_id WHERE s.domain='%s' AND p.name='%s';" % (DOMAIN, page))
    if int(after) != len(html):
        sys.exit("VERIFY FAILED: stored %s bytes, sent %d" % (after, len(html)))
    print("%s: stored %s bytes, 0 orphan refs" % (page, after))


def main():
    if len(sys.argv) < 3:
        sys.exit(__doc__)
    cmd = sys.argv[1]
    if cmd == "pull":
        pull(sys.argv[2], sys.argv[3] if len(sys.argv) > 3 else None)
    elif cmd == "push":
        push(sys.argv[2], sys.argv[3])
    elif cmd == "check":
        with open(sys.argv[2], encoding="utf-8") as fh:
            missing = orphan_refs(fh.read())
        print(", ".join(missing) if missing else "none")
    else:
        sys.exit(__doc__)


main()

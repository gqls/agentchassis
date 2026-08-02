#!/usr/bin/env python3
"""verify_assembled.py — prove the DECOMPOSED site before a single row is written.

WHAT IT ASSERTS, and the failure each assertion exists to catch:

  A. NUMERIC EQUIVALENCE (the 12 interactive pages). Every calculator, driven in
     its new assembled page, must reproduce the golden fingerprint captured from
     the live hand-built page across all three input vectors. This is the same
     comparison verify_rewrite.py runs, against the same golden, using the same
     imported Runner/VECTORS/numeric_diff — but in a DIFFERENT CONTEXT, and the
     context is the point. verify_rewrite spliced each component into its
     original page, so the site stylesheet, the page-local <style> and the
     document order were all still the original's. Assembly changes all three:
     the wrapper .container is gone, chrome comes from site_components, and the
     page-local style now rides in a section instead of <head>. A component
     proven equivalent in the old context is NOT thereby proven in the new one.

  B. NO VISIBLE TEXT LOST (all 27 pages). Every run of visible text in the
     original document must still be present in the assembled one. Compared as
     decoded, whitespace-collapsed text so that &amp; vs & is not reported as a
     loss — see assemble_mirror.visible_text for why that matters.

  C. EVERY SCRIPT TARGET STILL EXISTS (the 12 interactive pages). Each id the
     page's scripts address must be present in the assembled DOM. A calculator
     whose output element was classified as prose and then dropped by
     sectionHasVisibleContent would still compute — into nothing.

  F. NO PROSE TEXT ADDED (all 27 pages). The counterpart to B, and it exists
     because B's absence of a counterpart let a real change through. The
     calculator component is shared by /tools/standard-calc.html and /index.html,
     and it carried standard-calc's risk warning and two market-rate claims onto
     the homepage, which had never shown them. B passed — nothing was lost. The
     numeric fingerprint passed — none of the three has an id. A screenshot
     caught it. So the assembled page's text is now checked in BOTH directions,
     with the tool component's own copy excluded (it is the thing being
     substituted, so its text legitimately differs) and site chrome excluded.

  D. NO INTERNAL LINK GOES NOWHERE (all 27 pages). Checked against the real page
     list, because assembly drops nav.js and emits a server-side header instead:
     if the header's hand-maintained link list has drifted from `pages`, this is
     where it shows up, not in production.

  E. NO SECTION SILENTLY DROPPED (all 27 pages). getPageSections discards any
     row with 10 or fewer visible characters, without failing. A short prose
     block — a lone heading, a one-line caption — therefore vanishes from the
     page and nothing errors. Every drop is reported here as a failure, because
     at decomposition time a dropped block is content loss, not tidying.

WHAT IT STILL CANNOT TELL YOU. That the mirror is right. assemble_mirror.py is a
second implementation of the Go assembler, and the standing objection to second
implementations applies in full: this harness and that mirror agree with each
other by construction. The scheduled test for it is the FIRST PAGE SHIPPED —
diff the real rendered output against the mirror's prediction before the other
26 move. Nothing here substitutes for that.

Usage:
  python3 verify_assembled.py                    # everything
  python3 verify_assembled.py --pages index tool-standard-calc
  python3 verify_assembled.py --keep             # leave the staged site
  python3 verify_assembled.py --no-drive         # skip A (no browser needed)
"""
import json
import os
import re
import shutil
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
REWRITE = os.path.join(LANE, "rewrite")
sys.path.insert(0, LANE)
sys.path.insert(0, HERE)
sys.path.insert(0, REWRITE)

from assemble_mirror import (  # noqa: E402
    assemble_page, join_sections, rows_for_page, section_has_visible_content,
    visible_text,
)
from decompose_prover import raw_blocks, script_ids  # noqa: E402
from toolgolden import Runner, VECTORS, numeric_diff  # noqa: E402
from verify_rewrite import reconcile_renames, serve  # noqa: E402

DOMAIN = "loancalculator.co.uk"
LIVE = "https://" + DOMAIN
GOLDEN = os.path.join(LANE, "acceptance", "GOLDEN_2026-07-31c_tool_values.json")
CHROME = os.path.join(LANE, "chrome")

# Tools whose rewrite legitimately ADDED fingerprint keys (ids on controls that
# had none). Carried over verbatim from verify_rewrite.py's SPECS rather than
# re-derived: relaxing the comparison is a decision, and a decision copied by
# hand from one harness to another is a decision that will drift.
ALLOW_NEW_KEYS = {"tool-consolidation", "tool-credit-health-check",
                  "tool-application-tracker"}

# ── the ONE fingerprint key the decomposition legitimately removes ──────────
#
# The golden was captured from pages that built their nav in the browser: nav.js
# wrote the whole navigation into <div id="nav-placeholder">, so the placeholder
# appeared in the fingerprint carrying the nav's entire text. Assembly emits the
# same nav server-side from site_components.header, and the placeholder div does
# not exist any more. Its disappearance is the intended change.
#
# NARROW ON PURPOSE, and measured rather than assumed. Exactly three ids appear
# on all twelve golden pages — nav-placeholder, mobile-menu-btn, nav-links-menu —
# and they are the only chrome in the fingerprint. The other two are NOT excluded:
# the server-rendered header carries both, and they must still match key for key,
# which is what turns "the header lift is verbatim" from a claim into a check.
# On the first run they did match, and nav-placeholder was the sole divergence on
# all twelve pages.
#
# The lazy fix here would have been "ignore keys that vanished", which passes
# this run and also passes the run where a calculator's output element vanishes.
CHROME_IDS_REMOVED = {"nav-placeholder"}


def build_site(manifest, pages_by_name, staged, assets_dir):
    """Write the assembled site to `staged`, returning per-page diagnostics."""
    head = open(os.path.join(CHROME, "head.html"), encoding="utf-8").read().rstrip("\n")
    header = open(os.path.join(CHROME, "header.html"), encoding="utf-8").read().rstrip("\n")
    footer = open(os.path.join(CHROME, "footer.html"), encoding="utf-8").read().rstrip("\n")

    # Assets first: a component inherits the site stylesheet, and a page served
    # without it is a different page (the lesson verify_rewrite.py records).
    for root, _dirs, files in os.walk(assets_dir):
        for f in files:
            src = os.path.join(root, f)
            rel = os.path.relpath(src, assets_dir)
            dst = os.path.join(staged, rel)
            os.makedirs(os.path.dirname(dst), exist_ok=True)
            shutil.copy2(src, dst)

    built = {}
    for name, page in manifest.items():
        url = pages_by_name[name]["url"]
        rows = rows_for_page(page, REWRITE)
        sections, dropped = join_sections([(s, h) for s, h, _fn, _cd in rows])
        html = assemble_page(head, header, footer, sections, DOMAIN, url,
                             page["title"], page["meta_desc"])
        dest = os.path.join(staged, url.lstrip("/"))
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        open(dest, "w", encoding="utf-8").write(html)
        built[name] = {"url": url, "html": html, "dropped": dropped, "rows": rows}
    return built


RE_BODY = re.compile(r"<body\b[^>]*>(.*?)</body>", re.S | re.I)


def text_nodes(body_html):
    """Every text node in a body fragment, normalised, longest first.

    TEXT NODES, not "runs of the collapsed string". The first version of this
    check collapsed the whole page to one whitespace-normalised string and then
    split it on `\\s{2,}|\\n` — which, AFTER collapsing, matches nothing. Every
    page therefore produced exactly one run: the entire document. The check had
    silently become "is the whole original body one contiguous substring of the
    assembled page?", which is false by construction the moment assembly puts a
    header between two blocks, and it reported all 27 pages as losing text.

    A check that is too strict is not the safe direction. It reported 27
    failures, every one spurious, which is indistinguishable from a check that
    is broken — and the only way to clear it would have been to weaken it until
    it passed, which is how a real loss gets waved through.
    """
    t = re.sub(r"<(script|style)\b.*?</\1>", " ", body_html, flags=re.S | re.I)
    # HTML COMMENTS FIRST, and this was a real bug rather than tidiness. The
    # split below treats "<...>" as a tag, so a comment containing a ">" — every
    # one of the explanatory comments in these components does, they quote CSS
    # and markup — is torn in half and its prose counted as page text. It made
    # check F report a paragraph of a component's own commentary as content
    # added to the application-tracker page.
    t = re.sub(r"<!--.*?-->", " ", t, flags=re.S)
    nodes = []
    for chunk in re.split(r"<[^>]*>", t):
        import html as _h
        s = re.sub(r"\s+", " ", _h.unescape(chunk)).strip()
        if len(s) > 25:
            nodes.append(s)
    return nodes


def squash(s):
    """All whitespace removed, for containment tests across markup boundaries.

    A sentence the original split with inline markup — "…Personal Loans is
    <strong>7.9%</strong>." — collapses to "…Personal Loans is 7.9% ." because
    the tag becomes a space, while the component holds the same sentence as one
    text node with no space before the full stop. Comparing with spaces intact
    reported that as text added and text lost simultaneously, which is not a
    content difference at all. Removing whitespace on both sides answers the
    question actually being asked: are these the same characters?
    """
    return re.sub(r"\s+", "", s)


def check_text_preserved(original, assembled):
    """B — every text node of the original BODY must survive.

    Body only. The prover learned this on P6 and wrote it down: measuring the
    whole document made five pages fail by exactly the length of their <title>,
    which is not lost at all — it is carried in `title` and re-injected into the
    head by assembly. This function repeated that mistake on its first run and
    is scoped the same way for the same reason.
    """
    m = RE_BODY.search(original)
    body = m.group(1) if m else original
    got = squash(visible_text(assembled))
    return [n[:90] for n in text_nodes(body) if squash(n) not in got]


def main():
    argv = sys.argv[1:]
    keep = "--keep" in argv
    drive = "--no-drive" not in argv
    only = []
    if "--pages" in argv:
        only = [a for a in argv[argv.index("--pages") + 1:] if not a.startswith("--")]

    work = os.environ.get("DECOMP_WORK")
    if not work:
        sys.exit("set DECOMP_WORK to the directory holding manifest.json, "
                 "stored/ and assets/ (see RUNBOOK)")
    manifest = json.load(open(os.path.join(work, "manifest.json"), encoding="utf-8"))
    pages_by_name = {}
    for line in open(os.path.join(work, "pages.txt"), encoding="utf-8"):
        n, u = line.rstrip("\n").split("|")
        pages_by_name[n] = {"url": u}
    if only:
        manifest = {k: v for k, v in manifest.items() if k in only}

    staged = tempfile.mkdtemp(prefix="verify-assembled-")
    built = build_site(manifest, pages_by_name, staged, os.path.join(work, "assets"))

    valid_paths = {p["url"] for p in pages_by_name.values()}
    failures = []
    added_by_tool, moved_to_markup = {}, {}

    # ── B, C, D, E: static, no browser ──────────────────────────────────────
    for name, b in sorted(built.items()):
        original = open(os.path.join(work, "stored", name + ".html"),
                        encoding="utf-8").read()

        for slot, why in b["dropped"]:                                    # E
            failures.append("%s: section %s DROPPED (%s) — content loss"
                            % (name, slot, why))

        for run in check_text_preserved(original, b["html"]):             # B
            failures.append("%s: visible text lost: %r" % (name, run))

        # F — nothing ADDED, checked per ROW so the two cases stay separable.
        # A prose row carrying text the original never had is always a defect:
        # prose is lifted byte-for-byte, so it cannot legitimately gain a word.
        # A tool row is a SUBSTITUTION and may legitimately differ, so its
        # additions are listed for sign-off rather than assumed innocent.
        m = RE_BODY.search(original)
        orig_text = squash(visible_text(m.group(1) if m else original))
        # A string the ORIGINAL held inside its own <script> and wrote into the
        # page at runtime is not new copy — it is the same words, relocated. The
        # rewrites moved a lot of copy out of JS string literals and into the
        # markup deliberately (interpolating a quote-bearing sentence into a JS
        # literal is what killed the settlement calculator), so separating
        # "moved" from "added" is the difference between a report worth reading
        # and fifteen lines of noise that trains you to skip it.
        orig_scripts = squash("".join(raw_blocks("script", original)))
        for slot, rhtml, fn, _cd in b["rows"]:
            for node in text_nodes(rhtml):
                sq = squash(node)
                if sq in orig_text:
                    continue
                if fn is None:
                    failures.append("%s: prose row %s ADDS text absent from the "
                                    "original: %r" % (name, slot, node[:90]))
                elif sq in orig_scripts:
                    moved_to_markup.setdefault(name, []).append(node[:110])
                else:
                    added_by_tool.setdefault(name, []).append(node[:110])

        inline = [s for s in raw_blocks("script", b["html"])              # C
                  if not re.search(r"\bsrc\s*=", s, re.I)]
        for i in sorted(script_ids(inline)):
            if i.endswith("-"):
                continue  # a prefix pattern, not a literal id
            if ('id="%s"' % i) not in b["html"] and ("id='%s'" % i) not in b["html"]:
                failures.append("%s: script addresses #%s and no element has it"
                                % (name, i))

        for href in re.findall(r'href="(/[^"#?]*)"', b["html"]):          # D
            if href.startswith("/assets/"):
                continue
            if href not in valid_paths:
                failures.append("%s: internal link %s matches no page" % (name, href))

    # ── A: drive the calculators ────────────────────────────────────────────
    diverged = []
    if drive:
        golden = json.load(open(GOLDEN))["pages"]
        httpd, port = serve(staged)
        r = Runner()
        try:
            for name, b in sorted(built.items()):
                if not manifest[name]["tool_function"]:
                    continue
                local = "http://127.0.0.1:%d%s" % (port, b["url"])
                g = golden.get(LIVE + b["url"])
                if not g:
                    print("NO GOLDEN  %-32s %s" % (name, LIVE + b["url"]))
                    diverged.append(name)
                    continue
                try:
                    got = r.capture(local)
                except Exception as e:  # noqa: BLE001
                    print("CAPTURE-ERR %-31s %s" % (name, str(e)[:70]))
                    diverged.append(name)
                    continue

                diffs, added = [], set()
                for vec, _ in VECTORS:
                    for phase in ("after_input", "after_press"):
                        for fieldname in ("ids", "controls"):
                            before = {k: v for k, v in g[vec][phase].get(fieldname, {}).items()
                                      if k not in CHROME_IDS_REMOVED}
                            after = {k: v for k, v in got[vec][phase].get(fieldname, {}).items()
                                     if k not in CHROME_IDS_REMOVED}
                            if name in ALLOW_NEW_KEYS:
                                before, after, note = reconcile_renames(before, after)
                                added |= {fieldname + ":" + k for k in note}
                            d = numeric_diff(before, after)
                            if d:
                                diffs.append("   %s / %s / %s" % (vec, phase, fieldname))
                                diffs.extend(d)
                if diffs:
                    diverged.append(name)
                    print("\nDIVERGED   %s" % name)
                    print("\n".join(diffs[:30]))
                    if len(diffs) > 30:
                        print("   ... %d more lines" % (len(diffs) - 30))
                else:
                    n = len(got["defaults"]["after_press"].get("ids", {}))
                    extra = (" (+%d new key(s))" % len(added)) if added else ""
                    print("MATCHES    %-32s %d id-fields x %d vectors%s"
                          % (name, n, len(VECTORS), extra))
        finally:
            r.close()
            httpd.shutdown()

    if keep:
        print("\nstaged site kept at %s" % staged)
    else:
        shutil.rmtree(staged, ignore_errors=True)

    print()
    if failures:
        print("STATIC FAILURES (%d):" % len(failures))
        for f in failures[:40]:
            print("  " + f)
        if len(failures) > 40:
            print("  ... %d more" % (len(failures) - 40))
    else:
        print("static checks pass on all %d page(s): no text lost, no prose text "
              "added, no section dropped, no orphaned script target, no dead "
              "internal link" % len(built))
    if moved_to_markup:
        n = sum(len(v) for v in moved_to_markup.values())
        print("\n%d string(s) MOVED from the original's JavaScript into component "
              "markup, on %d page(s) — same words, different home:"
              % (n, len(moved_to_markup)))
        for page in sorted(moved_to_markup):
            print("  %-32s %d" % (page, len(moved_to_markup[page])))
    if added_by_tool:
        print("\nTEXT ADDED BY A TOOL COMPONENT — read it, it is a substitution "
              "and not automatically wrong:")
        for page in sorted(added_by_tool):
            for node in added_by_tool[page][:6]:
                print("  %-32s %r" % (page, node))
            if len(added_by_tool[page]) > 6:
                print("  %-32s ... %d more" % (page, len(added_by_tool[page]) - 6))
    if diverged:
        print("%d calculator(s) DIVERGED from golden" % len(diverged))
    elif drive:
        n = sum(1 for p in manifest.values() if p["tool_function"])
        print("all %d calculator(s) reproduce their golden values in the "
              "ASSEMBLED page" % n)
    return 1 if (failures or diverged) else 0


if __name__ == "__main__":
    sys.exit(main())

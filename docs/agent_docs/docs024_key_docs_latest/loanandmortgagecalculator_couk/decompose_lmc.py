#!/usr/bin/env python3
"""decompose_lmc.py — turn each of this site's 41 verbatim documents into an
ordered, byte-faithful manifest of prose runs and tool runs.

METHOD. The sibling lane's proven descent (decompose_pages.split_ordered) over
the children of the per-page content wrapper `div#content`: a node whose
subtree holds script-addressed elements is walked into, separating prose
siblings from widget machinery; everything is emitted as byte slices of the
source, never reconstructed markup. This matters here because the first draft
of this file assumed whole-child marking and the assertion suite refuted it
on real pages the same hour: mortgages/stamp-duty wraps h1 + intro + widget +
guidance cards in ONE <article> (whole-child marking froze the intro in the
locked tool row), and loans/standard-calc keeps two compliance boxes OUTSIDE
its .container. The descent handles both, and the assertions below stay as
the proof either way.

Consecutive prose blocks merge into one row per run (fewer, more coherent
rows for the voice pass); each contiguous tool run stays one tool block. A
page may legitimately decompose to a single tool block and no prose
(mortgages/simple.html is one card containing everything — its widget-internal
text is out of the voice's scope by the "copy zones only" rule).

GEOMETRY. Assembly wraps sections in <main>, which this lane's head chrome
styles as THE container (the shim neutralises .container inside main). So
dissolved wrappers cost nothing on the 27 normal-width pages; only
`container-tight` (13 guides + legal, 740px content box) carries geometry,
and those pages' prose rows are wrapped in a container-tight div by
load_lmc.py. Tight pages are asserted tool-free here, so the wrapper rule
stays that simple.

SCRIPTS. The calculators keep their logic in body-level inline <script>
blocks after the content wrapper, plus (12 mortgages pages)
<script src="/assets/js/calculators.js"> just before them. Both travel with
the LAST tool block, byte-exact, in source order — the external TAG included,
unlike the sibling (they had replacement components; these widgets ship
as-is). site.js is chrome (mobile menu only; footer component carries it):
its tag is asserted identical per page and dropped. The id set folds
calculators.js's source in (read from the pinned ref) — the sibling's hard
lesson: an id set built from inline scripts alone classifies a shared-js
results box as editable prose. Any OTHER external script is a hard failure.

Scripts moving from "after all prose" to "with the last tool block" is a
deliberate, stated reordering: the inline scripts define functions and
address only id-marked elements (all parsed before the scripts' new
position), and the per-page golden compare catches any behavioural change
regardless.

CHROME ASSERTIONS. Per page, the skip-link, <header>, <footer> and the
site.js tag must equal the chrome this lane installed (header compared with
the per-page aria-current attribute stripped — the one thing that varied).
A page whose chrome differs is a page the shared components would silently
rewrite, so it refuses. The chrome-ids check is by id MEMBERSHIP against the
page's script-id set (never by control-tag marking: the header's mobile-menu
BUTTON is chrome machinery, and marking it tool — the first draft did — flags
all 41 pages at once, the uniform-failure signature of an instrument fault).

ASSERTIONS (all hard, all per page): P-visible (visible text of the blocks
equals the wrapper's, so nothing a reader sees is lost), P-stranded (no
script-addressed id in prose), P-script-bytes (every inline script + the
calculators.js tag ride byte-exact in tool blocks), P-mixed ceiling (no tool
block carries >1500 chars of visible text), tool pages have >=1 tool run,
non-tool pages have none and no inline scripts, tight pages are tool-free.

INPUT is the sites repo at the PINNED ref below — the same bytes as every
stored row (verified 41/41 md5-identical 2026-08-05; load_lmc.py re-verifies
against the live DB before writing). Pinned because a baseline that names a
moving thing stops being a control.

Usage:
  python3 decompose_lmc.py <out-manifest.json> [--pages name1,name2] [--verbose]
"""
import json
import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
SIBLING = os.path.join(os.path.dirname(HERE), "loancalculator_couk")
sys.path.insert(0, SIBLING)

from decompose_prover import (  # noqa: E402
    TreeBuilder, any_marked, raw_blocks, script_ids, visible,
)
from decompose_pages import collapse_runs, split_ordered  # noqa: E402

PINNED_REF = "b318a8fad"
SITE_DIR = "loanandmortgagecalculator.co.uk"
SITES_REPO = os.path.expanduser("~/projects/sites")

CALCULATOR_URLS = {
    "loans/application-tracker", "loans/car-finance-calculator",
    "loans/compare-loans", "loans/consolidation", "loans/credit-health-check",
    "loans/damage-checker", "loans/interest-rate-stress-test",
    "loans/loan-vs-savings", "loans/overpayment-calculator",
    "loans/settlement-calculator", "loans/standard-calc",
    "mortgages/affordability", "mortgages/bridging-loan",
    "mortgages/equity-release", "mortgages/fact-finder",
    "mortgages/fee-analyser", "mortgages/investor", "mortgages/overpayment",
    "mortgages/portfolio", "mortgages/rate-forecaster", "mortgages/repayment",
    "mortgages/simple", "mortgages/stamp-duty",
}

SITE_JS_TAG = '<script src="/assets/js/site.js"></script>'
CALC_JS_TAG = '<script src="/assets/js/calculators.js"></script>'
MIXED_VISIBLE_CEILING = 1500


def repo_file(path):
    r = subprocess.run(["git", "show", "%s:%s/%s" % (PINNED_REF, SITE_DIR, path)],
                       capture_output=True, text=True, cwd=SITES_REPO)
    if r.returncode != 0:
        raise SystemExit("git show %s failed: %s" % (path, r.stderr.strip()[:200]))
    return r.stdout


def strip_aria_current(s):
    return s.replace(' aria-current="page"', "")


def page_paths():
    r = subprocess.run(["git", "ls-tree", "-r", "--name-only", PINNED_REF],
                       capture_output=True, text=True, cwd=SITES_REPO)
    out = []
    for line in r.stdout.splitlines():
        if line.startswith(SITE_DIR + "/") and line.endswith(".html") \
                and not line.endswith("404.html"):
            out.append(line[len(SITE_DIR) + 1:])
    return sorted(out)


def strip_code(s):
    return re.sub(r"<(script|style)\b.*?</\1>", " ", s, flags=re.S | re.I)


def decompose_page(relpath, html, chrome, calc_js_src, verbose=False):
    stem = relpath[:-5]
    is_tool_page = stem in CALCULATOR_URLS

    head_m = re.search(r"<head\b[^>]*>(.*?)</head>", html, re.S | re.I)
    body_m = re.search(r"<body\b[^>]*>(.*?)</body>", html, re.S | re.I)
    if not head_m or not body_m:
        raise SystemExit("%s: no head/body" % relpath)
    head_s, body_s = head_m.group(1), body_m.group(1)

    # The guides + home carry an ld+json data block in <head> (14 site-wide,
    # censused 07-31). It is dropped with the per-page head — an ACCEPTED
    # loss, stated in the PLAN: assembly injects its own WebPage JSON-LD per
    # page. Only an EXECUTABLE head script is a stop-the-line surprise.
    for s in raw_blocks("script", head_s):
        if "application/ld+json" not in s[:80]:
            raise SystemExit("%s: executable <script> in <head> — believed to "
                             "carry none; widen the rule" % relpath)
    if re.search(r"<style\b", body_s, re.I) or re.search(r"<style\b", head_s, re.I):
        raise SystemExit("%s: page-local <style> — censused ZERO site-wide; "
                         "widen the rule" % relpath)

    title_m = re.search(r"<title>(.*?)</title>", head_s, re.S | re.I)
    desc_m = re.search(r'<meta\s+name=["\']description["\']\s+content=["\'](.*?)["\']',
                       head_s, re.S | re.I)

    tb = TreeBuilder(body_s)
    tb.feed(body_s)
    tb.close()

    problems = []
    by_tag = {}
    for c in tb.root.children:
        by_tag.setdefault(c.tag, []).append(c)

    # ── chrome agreement ────────────────────────────────────────────────
    skip = [c for c in by_tag.get("a", []) if "skip-link" in c.starttag]
    if len(skip) != 1 or body_s[skip[0].start:skip[0].end] != chrome["skip"]:
        problems.append("skip-link missing or differs from chrome")
    hdrs = by_tag.get("header", [])
    if len(hdrs) != 1 or \
            strip_aria_current(body_s[hdrs[0].start:hdrs[0].end]) != chrome["header"]:
        problems.append("header differs from chrome (beyond aria-current)")
    foots = by_tag.get("footer", [])
    if len(foots) != 1 or body_s[foots[0].start:foots[0].end] != chrome["footer"]:
        problems.append("footer differs from chrome")

    # ── scripts ─────────────────────────────────────────────────────────
    inline_scripts, ext_tags = [], []
    for c in by_tag.get("script", []):
        span = body_s[c.start:c.end]
        if re.search(r"\bsrc\s*=", c.starttag, re.I):
            ext_tags.append(span)
        else:
            inline_scripts.append(span)
    site_js = [t for t in ext_tags if "/assets/js/site.js" in t]
    calc_js = [t for t in ext_tags if "/assets/js/calculators.js" in t]
    other_js = [t for t in ext_tags if t not in site_js + calc_js]
    if len(site_js) != 1 or site_js[0] != SITE_JS_TAG:
        problems.append("site.js tag missing or differs (chrome carries it)")
    if other_js:
        problems.append("unexpected external script(s): %s" % other_js)
    if calc_js and calc_js != [CALC_JS_TAG]:
        problems.append("calculators.js tag differs from the expected literal")

    inline_bodies = [re.sub(r"^<script\b[^>]*>|</script>$", "", s,
                            flags=re.I | re.S) for s in inline_scripts]
    ids = script_ids(inline_bodies + ([calc_js_src] if calc_js else []))

    # chrome must not carry script-addressed ids — by MEMBERSHIP, see docstring
    chrome_html = "".join(body_s[n[0].start:n[0].end]
                          for n in (hdrs, foots) if n)
    hit = sorted(i for i in ids if 'id="%s"' % i in chrome_html)
    if hit:
        problems.append("chrome carries script-addressed id(s): %s" % ",".join(hit))

    # ── the content wrapper ─────────────────────────────────────────────
    content = [c for c in by_tag.get("div", []) if 'id="content"' in c.starttag]
    if len(content) != 1:
        problems.append("expected exactly one div#content, found %d" % len(content))
        return None, problems
    croot = content[0]
    cls_m = re.search(r'class="([^"]*)"', croot.starttag)
    cclass = (cls_m.group(1) if cls_m else "").split()
    tight = "container-tight" in cclass

    inner_start = croot.inner_start
    inner_end = croot.end - len("</div>")
    inner_html = body_s[inner_start:inner_end]

    # ── descent ─────────────────────────────────────────────────────────
    ordered = []
    if ids or is_tool_page:
        for blk in croot.children:
            if any_marked(blk, ids):
                split_ordered(blk, body_s, ids, ordered)
            else:
                ordered.append(("prose", body_s[blk.start:blk.end]))
        # loose text directly inside #content (between children) would be lost
        # by the child walk; P-visible below catches it, and no page has any
        # today — this keeps the check honest rather than assuming it.
    else:
        ordered = [("prose", inner_html)]

    ordered = [(k, h) for k, h in ordered if h.strip()]
    ordered, tool_runs = collapse_runs(ordered)

    # merge consecutive prose blocks into one row per run
    merged = []
    for kind, h in ordered:
        if kind == "prose" and merged and merged[-1][0] == "prose":
            merged[-1] = ("prose", merged[-1][1] + "\n" + h)
        else:
            merged.append((kind, h))

    blocks = []
    for kind, h in merged:
        if kind == "tool":
            blocks.append({"kind": "tool", "html": h, "scripts": ""})
        else:
            blocks.append({"kind": "prose", "html": h})

    tool_blocks = [b for b in blocks if b["kind"] == "tool"]
    if is_tool_page and not tool_blocks:
        problems.append("calculator page but no tool block found")
    if not is_tool_page and tool_blocks:
        problems.append("tool block on a non-calculator page")
    if not is_tool_page and inline_scripts:
        problems.append("inline script on a non-calculator page")
    if tight and tool_blocks:
        problems.append("container-tight page with a tool block — the wrapper "
                        "rule in load_lmc.py assumes this cannot happen")
    if tool_blocks:
        script_tail = "\n".join((calc_js if calc_js else []) + inline_scripts)
        tool_blocks[-1]["scripts"] = script_tail

    for b in tool_blocks:
        t = visible(strip_code(b["html"]))
        if len(t) > MIXED_VISIBLE_CEILING:
            problems.append("a tool block carries %d chars of visible text "
                            "(ceiling %d) — mixed wrapper suspected"
                            % (len(t), MIXED_VISIBLE_CEILING))

    prose_all = "".join(b["html"] for b in blocks if b["kind"] == "prose")
    stranded = sorted(i for i in ids
                      if ('id="%s"' % i) in prose_all or ("id='%s'" % i) in prose_all)
    if stranded:
        problems.append("P-stranded: script target(s) in prose: %s"
                        % ",".join(stranded))

    orig_vis = visible(strip_code(inner_html))
    got_vis = visible(strip_code("".join(b["html"] for b in blocks)))
    if orig_vis != got_vis:
        problems.append("P-visible: visible text differs after decomposition "
                        "(orig %d chars, got %d)" % (len(orig_vis), len(got_vis)))

    for s in inline_scripts:
        if not any(s in b["scripts"] for b in tool_blocks):
            problems.append("P-script-bytes: an inline script is not in any "
                            "tool block")
    if calc_js and not any(CALC_JS_TAG in b["scripts"] for b in tool_blocks):
        problems.append("P-script-bytes: calculators.js tag lost")

    entry = {
        "name": ("index" if stem == "index" else stem.replace("/", "-")),
        "url": "/" + relpath,
        "title": title_m.group(1).strip() if title_m else "",
        "meta_desc": desc_m.group(1).strip() if desc_m else "",
        "tight": tight,
        "is_tool_page": is_tool_page,
        "uses_calculators_js": bool(calc_js),
        "tool_runs": tool_runs,
        "blocks": blocks,
        "prose_chars": len(visible(strip_code(prose_all))),
    }
    if verbose:
        print("  %-44s %s" % (relpath, " ".join(
            "%s:%d" % (b["kind"], len(b["html"]) + len(b.get("scripts", "")))
            for b in blocks)))
    return entry, problems


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    out_path = sys.argv[1]
    only = None
    if "--pages" in sys.argv:
        only = set(sys.argv[sys.argv.index("--pages") + 1].split(","))
    verbose = "--verbose" in sys.argv

    chrome_dir = os.path.join(HERE, "chrome")
    chrome_header_file = open(os.path.join(chrome_dir, "header.html"),
                              encoding="utf-8").read()
    hm = re.search(r"<header\b.*?</header>", chrome_header_file, re.S)
    sm = re.search(r'<a class="skip-link".*?</a>', chrome_header_file, re.S)
    chrome_footer_file = open(os.path.join(chrome_dir, "footer.html"),
                              encoding="utf-8").read()
    fm = re.search(r"<footer\b.*?</footer>", chrome_footer_file, re.S)
    chrome = {"header": hm.group(0), "skip": sm.group(0), "footer": fm.group(0)}

    calc_js_src = repo_file("assets/js/calculators.js")

    manifest, all_problems = {}, []
    for relpath in page_paths():
        stem = relpath[:-5]
        name = "index" if stem == "index" else stem.replace("/", "-")
        if only and name not in only:
            continue
        html = repo_file(relpath)
        entry, problems = decompose_page(relpath, html, chrome, calc_js_src,
                                         verbose)
        for p in problems:
            all_problems.append("%s: %s" % (relpath, p))
        if entry:
            manifest[entry["name"]] = entry

    ntool = sum(1 for e in manifest.values()
                if any(b["kind"] == "tool" for b in e["blocks"]))
    print("pages: %d   with a tool block: %d   (expected 23)"
          % (len(manifest), ntool))
    if only is None and ntool != 23:
        all_problems.append("expected 23 tool pages, found %d" % ntool)

    if all_problems:
        print("\nREFUSING to write manifest (%d problem(s)):" % len(all_problems))
        for p in all_problems:
            print("  " + p)
        return 1

    with open(out_path, "w", encoding="utf-8") as fh:
        json.dump({"pinned_ref": PINNED_REF, "pages": manifest}, fh, indent=1,
                  ensure_ascii=False)
    print("wrote %s" % out_path)
    return 0


if __name__ == "__main__":
    sys.exit(main())

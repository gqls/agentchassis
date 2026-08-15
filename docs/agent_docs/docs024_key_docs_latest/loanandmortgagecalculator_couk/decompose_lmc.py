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
import hashlib
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

PINNED_REF = "7e6b993ef"  # 2026-08-14: the LAST CLEAN pin. 0a0e89326 captured my own 08-12 restore-deploys, whose backup (20260805) PREDATES the 224 fix — a restore is a time machine
SITE_DIR = "loanandmortgagecalculator.co.uk"
SITES_REPO = os.path.expanduser("~/projects/sites")
# Same value as load_lmc.py's; needed here only by assert_pin_matches_live().
SITE_ID = "ed633ada-f8af-424b-b4d4-8af79160dbcd"

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

# ── the mixed-card five (Track B2 batch 2, 2026-08-15) ──────────────────────
# Each of these pages has EXACTLY ONE `.card` whose children mix copy with
# machinery (a heading, sometimes an advisory <p>, beside the grid). The descent
# therefore walks into the card and dissolves it — `gate_wrapper_parity` refused
# all five on 2026-08-15 with {'card': (1, 0)}, which is the gate working.
#
# Owner ruling 2026-08-13 for this class: take the card WHOLE and let the mixed
# copy become fields. Under B2 that costs nothing a writer needs — the block is a
# parameterised template, every clean copy span is a schema field, and the row is
# unlocked — which is what distinguishes it from the "whole-child marking" that
# was refuted on 2026-08-05, when whole meant FROZEN in a locked verbatim row.
#
# Named per page, not site-wide, because the other 16 calculator pages are
# already converted with the plain `keep_widget_wrapper` rule and re-running them
# through a wider rule would change bytes nobody asked to change.
WHOLE_CARD_PAGES = {
    "loans-damage-checker", "mortgages-bridging-loan", "mortgages-equity-release",
    "mortgages-fee-analyser", "mortgages-rate-forecaster",
}
WHOLE_CARD_CLASSES = {"card"}


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
    page_name = "index" if stem == "index" else stem.replace("/", "-")
    whole_cards = WHOLE_CARD_CLASSES if page_name in WHOLE_CARD_PAGES else ()

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
    # WIDENED 2026-08-12, exactly as the old message asked. The ZERO census was taken
    # 2026-07-31, BEFORE this lane's own decomposition existed. Track A then installed
    # the assembled-layout shim into the shared chrome head (load_chrome.py, reasoning
    # in PLAN_2026-08-05), so every assembled page now carries one page-local <style>
    # and the assertion fired on the lane's own intended CSS —
    # guides/car-finance-and-your-mortgage.html, which blocked Track B until this.
    #
    # A style block in the BODY is still a stop-the-line surprise, and so is any HEAD
    # block that is not the shim: the head is discarded at decomposition, so anything
    # load-bearing in it would be silently lost. Identified by the shim's own comment
    # marker rather than by byte equality, because its --maxw value is site-scoped.
    # Re-censused at pin 7e6b993ef: ZERO of the 22 tool pages in scope carry any
    # page-local <style> at all, so this widening changes nothing for Track B and only
    # unblocks the already-decomposed pages the script walks past.
    if re.search(r"<style\b", body_s, re.I):
        raise SystemExit("%s: page-local <style> in <body> — still censused ZERO; "
                         "a body style survives decomposition unpredictably, so stop "
                         "and look" % relpath)
    for st in raw_blocks("style", head_s):
        if "Assembled-layout shim" not in st:
            raise SystemExit("%s: page-local <style> in <head> that is NOT the "
                             "assembled-layout shim — the head is discarded at "
                             "decomposition, so this would be lost silently. Widen the "
                             "rule only after reading what it styles." % relpath)

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
                # 263: opt in to keeping the widget's own wrapper (.card/.calc-grid/
                # .input-grid) inside the tool block. Default-OFF on the shared
                # helper so the sibling lane's stored rows keep their meaning.
                split_ordered(blk, body_s, ids, ordered,
                              keep_widget_wrapper=True,
                              whole_wrapper_classes=whole_cards)
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
            if whole_cards:
                # EXPECTED on the mixed-card five, and printed rather than
                # silently skipped: this ceiling exists to detect a tool block
                # that swallowed copy, and on these pages it is swallowing copy
                # ON PURPOSE. The protection that replaces it is downstream and
                # stronger — b2_build turns each clean copy span into an unlocked
                # schema field, and the per-page field count is asserted at load.
                # A number here that is WILDLY larger than the card's own copy
                # would still mean the slice took too much, so read it.
                print("  note %-40s tool block carries %d chars of visible text "
                      "(ceiling %d waived: whole-card page, copy becomes fields)"
                      % (relpath, len(t), MIXED_VISIBLE_CEILING))
            else:
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


def assert_pin_matches_live(only, allow_stale=False):
    """REFUSE if the pinned ref no longer holds what the live DB holds.

    Added 2026-08-11 after measuring, for Track B, that PINNED_REF `b318a8fad`
    matched the stored rows of only **6 of the 22** owned+verbatim calculator
    pages. Decomposing from it would have written stale calculator HTML over
    **16 live calculators** — reverting the `bugs_open/224` zero-rate guards and
    the `bugs_open/225` SDLT fix, a tax rule that had been 16 months out of date
    and under-quoting by £5,000.

    Nothing checked this. The pin is deliberately fixed — the docstring's
    "a baseline that names a moving thing stops being a control" is right — but a
    pin that has silently drifted from the artefact it describes is no longer a
    baseline either, it is just old. The two failure modes need different
    treatment and only one of them was guarded.

    This is a HARD REFUSAL rather than a warning because the damage is silent,
    irreversible in practice (it reverts arithmetic fixes on live consumer-finance
    tools) and looks exactly like a successful run. `ALLOW_STALE_PIN=1` overrides
    for the case where you genuinely mean to decompose historical bytes; it prints
    what it is overriding.
    """
    r = subprocess.run(
        ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0",
         "--", "psql", "-U", "clients_user", "-d", "clients_db", "-tA", "-F", "\t",
         "-c", "SELECT p.url, md5(pc.rendered_html) FROM pages p "
               "JOIN page_components pc ON pc.page_id=p.id "
               "WHERE p.site_id='%s' AND p.sections::text='[\"ported-page\"]';" % SITE_ID],
        capture_output=True, text=True)
    if r.returncode != 0:
        print("WARN: could not reach the DB to check the pin (%s) — NOT verified"
              % (r.stderr or r.stdout).strip()[:120])
        return
    stale, checked = [], 0
    for line in r.stdout.strip().splitlines():
        parts = line.split("\t")
        if len(parts) < 2:
            continue
        url, live_md5 = parts[0], parts[1]
        rel = url.lstrip("/")
        name = "index" if rel[:-5] == "index" else rel[:-5].replace("/", "-")
        if only and name not in only:
            continue
        src = subprocess.run(["git", "show", "%s:%s/%s" % (PINNED_REF, SITE_DIR, rel)],
                             capture_output=True, text=True, cwd=SITES_REPO)
        if src.returncode != 0:
            continue
        checked += 1
        if hashlib.md5(src.stdout.encode("utf-8")).hexdigest() != live_md5:
            stale.append(rel)
    if checked == 0:
        # Say so rather than reporting a pass. "matches for all 0 pages" is a
        # vacuous truth and reads exactly like a real check having succeeded —
        # the same empty-set trap that nearly let a bogus calculator/prose
        # partition proof through earlier today (WRONG_CALLS, 2026-08-11).
        print("pin %s NOT CHECKED: no verbatim page in scope (nothing this guard "
              "can protect — every page here is already decomposed)" % PINNED_REF)
        return
    if not stale:
        print("pin %s matches the live stored rows for all %d verbatim page(s) in scope"
              % (PINNED_REF, checked))
        return
    msg = ("PINNED_REF %s is STALE for %d of %d verbatim page(s) in scope:\n  %s\n"
           "Decomposing these would write the PINNED bytes over the LIVE ones.\n"
           "Re-point PINNED_REF to a concrete SHA whose bytes match (origin/master\n"
           "matched 22/22 on 2026-08-11) — never to a branch name, and re-verify at\n"
           "the moment of use: rerenders push to that repo continuously."
           % (PINNED_REF, len(stale), checked, "\n  ".join(sorted(stale))))
    if allow_stale:
        print("WARNING (ALLOW_STALE_PIN=1 — overriding):\n" + msg)
        return
    raise SystemExit("REFUSING: " + msg)


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    out_path = sys.argv[1]
    only = None
    if "--pages" in sys.argv:
        only = set(sys.argv[sys.argv.index("--pages") + 1].split(","))
    verbose = "--verbose" in sys.argv

    assert_pin_matches_live(only, allow_stale=os.environ.get("ALLOW_STALE_PIN") == "1")

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

#!/usr/bin/env python3
"""verify_rewrite.py — prove a rewritten tool computes what the original computed.

THE PROBLEM THIS SOLVES. `toolgolden.py --compare` drives LIVE urls, so it can
only judge a rewrite AFTER it has been deployed. That is the wrong order: on this
platform a commit is a deploy (HEAD is shared, any session's roll ships it), so
"deploy then check" means shipping an unverified calculator to a public money
page and finding out afterwards. This runs the identical comparison against the
rewrite BEFORE it goes anywhere.

HOW. For each tool it takes the REAL page byte-for-byte, removes the original
widget, splices the rendered component in its place, and serves the whole site
locally so stylesheets, nav and every other asset resolve exactly as they do in
production. Only the widget differs. It then drives the local page with
toolgolden's own Runner and the same three vectors, and diffs the fingerprint
against the golden entry recorded for the LIVE url.

WHY IT REUSES toolgolden RATHER THAN REIMPLEMENTING THE COMPARISON. A second
implementation of "did the numbers move" is a second thing that can be subtly
wrong, and the two would agree with each other for exactly the reasons they were
both wrong. Runner, VECTORS, numeric_diff and the settle logic are imported.

WHAT IT CANNOT TELL YOU. That the ORIGINAL was right. The golden is a record of
what the tool did, not of what it should do; a rewrite proven equivalent to a
broken original is a faithful copy of a broken original. That question belongs
to the tool's PLAN and to a human reading the arithmetic.

Usage:
  python3 verify_rewrite.py                 # every tool with a spec below
  python3 verify_rewrite.py settlement-calculator car-finance-calculator
  python3 verify_rewrite.py --keep          # leave the staged site for inspection
"""
import http.server
import json
import os
import re
import shutil
import socketserver
import subprocess
import sys
import tempfile
import threading

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
sys.path.insert(0, LANE)

from toolgolden import Runner, VECTORS, numeric_diff  # noqa: E402

SITE_SRC = os.path.expanduser("~/projects/sites/loancalculator.co.uk")
GOLDEN = os.path.join(LANE, "acceptance", "GOLDEN_2026-07-31c_tool_values.json")
LIVE = "https://loancalculator.co.uk"

# ── the splice specs ───────────────────────────────────────────────────────
#
# One entry per tool: where its page lives, and how to CUT the original widget
# out of it. `cut` is a list of patterns; every one must match exactly once, and
# the harness refuses the tool if any matches zero or more than one time. That
# strictness is not fussiness — a cut pattern that silently matches nothing leaves
# the ORIGINAL widget on the page beside the replacement, and the page then
# passes the comparison while proving nothing about the rewrite. (Two ids, two
# scripts, first one wins; the numbers come out identical.) An over-matching one
# removes surrounding prose and the diff blames the component.
#
# TWO PATTERN KINDS, because a regex cannot count. Every one of these pages uses
# `<div class="card">` for its ARTICLE sections as well as for the tool, so
# `<div class="card">.*?</div></div></div>` matches the wrong block or no block
# depending on how many children the tool happens to have — it silently found
# zero on standard-calc, which is the failure this strictness caught.
#
#   "<regex>"       plain regex, must match exactly once
#   "DIV:<regex>"   the regex locates the OPENING <div ...> of the region; the
#                   harness then BALANCES <div>/</div> to find its real end.
#                   Use this for anything div-delimited — it is correct for a
#                   region with any number of nested children, which is what a
#                   tool widget is.
#
# The component is inserted where the FIRST cut was, so it lands in the page flow
# rather than at the end of the body.
#
# NOTE ON ORDER, learned the hard way on tool 2: the cuts run BEFORE the splice.
# Reversed, a removal regex written in terms of an identifier matches the
# replacement's own code — or worse, the comment in the replacement explaining
# that the identifier is gone.
SPECS = {
    "settlement-calculator": {
        "page": "tools/settlement-calculator.html",
        "component": "tool-early-settlement",
        "cut": ['DIV:<div class="card">', r'<script>\s*\n\s*function estSettle\(\).*?</script>'],
    },
    "interest-rate-stress-test": {
        "page": "tools/interest-rate-stress-test.html",
        "component": "tool-rate-stress-test",
        # Two regions: the input card and the outcomes grid, with prose between
        # neither of them. Anchored on the child that identifies each, because
        # this page has three cards and two of them are results panels.
        "cut": ['DIV:<div class="card">\\s*<div class="input-grid">',
                'DIV:<div class="comparison-grid">',
                r'<script>\s*\n\s*function stressTest\(\).*?</script>'],
    },
    "overpayment-calculator": {
        "page": "tools/overpayment-calculator.html",
        "component": "tool-overpayment-impact",
        "cut": ['DIV:<div class="card">',
                'DIV:<div class="highlight-box">',
                r'<script>\s*\n\s*const fields.*?</script>'],
    },
    "standard-calc": {
        "page": "tools/standard-calc.html",
        "component": "tool-loan-repayment",
        # Anchored on input-grid: this page has FOUR cards and the other three
        # are article sections. A bare card regex matched zero of them, which is
        # what the exactly-once rule caught.
        "cut": ['DIV:<div class="card">\\s*<div class="input-grid">',
                r'<script>\s*\n\s*const inputs.*?</script>'],
    },
    "application-tracker": {
        "page": "tools/application-tracker.html",
        "component": "tool-application-tracker",
        # allow_new_keys: the backup controls were OUTSIDE the page container and
        # their buttons had no ids, so nothing could name them. They are inside
        # the component now and addressable. Re-baseline owed once this ships.
        "allow_new_keys": True,
        "cut": ['DIV:<div class="progress-bar">',
                'DIV:<div class="tracker-container">',
                r'<script>\s*\n\s*const checkboxes.*?</script>',
                # The backup block and its script sit AFTER the container's
                # closing tag in the original — outside the page layout entirely.
                'DIV:<div style="display: flex; gap: 10px; margin-top: 10px;">',
                r'<script>\s*\n\s*function downloadBackup\(\).*?</script>'],
    },
    "consolidation": {
        "page": "tools/consolidation.html",
        "component": "tool-consolidation-risk",
        # allow_new_keys: the debt rows, the add button and the remove buttons had
        # NO ids at all, so the capture emitter refused this tool outright and it
        # had no numeric coverage available to it. Giving every control an id is
        # what makes it coverable, and an id necessarily adds a fingerprint key
        # and renames its `controls` entry. Every pre-existing key is still
        # compared strictly. Re-baseline owed once this ships.
        "allow_new_keys": True,
        "cut": ['DIV:<div class="card" id="debt-manager">',
                'DIV:<div class="card">\\s*<h3>2\\.',
                'DIV:<div class="comparison-grid">',
                'DIV:<div class="fca-warning-box"',
                r'<script>\s*\n\s*function addDebtRow\(\).*?</script>'],
    },
    "car-finance-calculator": {
        "page": "tools/car-finance-calculator.html",
        "component": "tool-car-finance-pcp-hp",
        "cut": ['DIV:<div class="card">\\s*<div class="finance-type-toggle">',
                'DIV:<div class="comparison-grid">',
                r'<script>\s*\n\s*let mode.*?</script>'],
    },
    "compare-loans": {
        "page": "tools/compare-loans.html",
        "component": "tool-compare-loan-offers",
        "cut": ['DIV:<div class="comparison-wrapper">',
                'DIV:<div class="card" style="margin-top: 30px; text-align: center;">',
                r'<script>\s*\n\s*function calc\(p, r, y\).*?</script>'],
    },
    "loan-vs-savings": {
        "page": "tools/loan-vs-savings.html",
        "component": "tool-loan-vs-savings",
        "cut": ['DIV:<div class="card">',
                'DIV:<div class="comparison-card" id="results">',
                r'<script>\s*\n\s*function compare\(\).*?</script>'],
    },
    "damage-checker": {
        "page": "tools/damage-checker.html",
        "component": "tool-return-damage-checker",
        "cut": ['DIV:<div class="card">',
                r'<script>\s*\n\s*function updateDamage\(\).*?</script>'],
    },
    "credit-health-check": {
        "page": "tools/credit-health-check.html",
        "component": "tool-credit-health-check",
        # allow_new_keys: the wizard's answer buttons had no ids, so the capture
        # emitter refused the tool ("pressed button 'Yes' has no id") and it had
        # no numeric coverage available. Every button is addressable now.
        "allow_new_keys": True,
        "cut": [r'<style>.*?</style>',
                'DIV:<div class="card" id="quiz-container">',
                r'<script>.*?</script>'],
    },
}


DIV_TOKEN = re.compile(r'<div\b|</div>', re.I)


def find_region(html, pattern):
    """Locate one cut region. Returns (start, end) or an error string.

    For a DIV: pattern the end is found by BALANCING tags from the opening div,
    which is the only way to cut a region whose children vary. HTML is not a
    regular language and these pages nest three deep inside the tool card.
    """
    if pattern.startswith("DIV:"):
        hits = list(re.finditer(pattern[4:], html, re.S | re.I))
        if len(hits) != 1:
            return None, "opening-div regex matched %d times (need exactly 1)" % len(hits)
        start = hits[0].start()
        depth = 0
        for m in DIV_TOKEN.finditer(html, start):
            depth += 1 if m.group(0).lower().startswith("<div") else -1
            if depth == 0:
                return (start, m.end()), None
        return None, "unbalanced <div> from offset %d — the page itself is malformed" % start

    hits = list(re.finditer(pattern, html, re.S))
    if len(hits) != 1:
        return None, "regex matched %d times (need exactly 1): %s" % (len(hits), pattern[:60])
    return (hits[0].start(), hits[0].end()), None


def reconcile_renames(before, after):
    """Separate RENAMES from real changes, for a tool being given ids.

    Giving a control an id changes its fingerprint key: toolgolden keys `controls`
    by `id || name || tag#index`, so an input that was `input#3` becomes
    `debt-2-bal`. Naively that reads as one control vanishing and another
    appearing, and the second half of that — a key the golden had, now absent —
    is exactly what a genuine regression looks like too.

    So renames are identified BY VALUE and only in matched pairs: a key that
    disappeared is paired with a key that appeared carrying the same value, and
    only then dropped from the comparison. Anything left unpaired stays in the
    diff. A control that actually lost its value therefore still fails, because
    nothing new will carry that value to pair with it.

    Returns (before, after, changed_keys) with the paired keys removed.
    """
    removed = set(before) - set(after)
    fresh = set(after) - set(before)
    if not removed and not fresh:
        return before, after, set()

    by_value = {}
    for k in fresh:
        by_value.setdefault(after[k], []).append(k)

    before, after = dict(before), dict(after)
    noted = set()
    for k in sorted(removed):
        candidates = by_value.get(before[k])
        if not candidates:
            continue  # genuinely gone — leave it in the diff
        new_key = candidates.pop()
        noted.add("%s->%s" % (k, new_key))
        del before[k]
        del after[new_key]
        fresh.discard(new_key)

    for k in sorted(fresh):  # appeared without displacing anything
        noted.add("+" + k)
        del after[k]
    return before, after, noted


def component_for(slug):
    return SPECS[slug].get("component", "tool-" + slug)


def render(slug):
    """Render the component with the REAL Go template engine (see render_tool.go)."""
    comp = component_for(slug)
    tmpl = os.path.join(HERE, comp + ".html.tmpl")
    schema = os.path.join(HERE, comp + ".schema.json")
    if not os.path.exists(tmpl):
        return None, "no template at %s" % os.path.relpath(tmpl, LANE)
    out = os.path.join(tempfile.gettempdir(), comp + ".rendered.html")
    r = subprocess.run(["go", "run", os.path.join(HERE, "render_tool.go"),
                        tmpl, schema, out],
                       capture_output=True, text=True, cwd=HERE)
    if r.returncode != 0:
        return None, (r.stderr or r.stdout).strip()[:300]
    if r.stderr.strip():
        print("      %s" % r.stderr.strip())
    return open(out, encoding="utf-8").read(), None


def splice(slug, staged_root):
    """Cut the original widget out of the real page and put the component in."""
    spec = SPECS[slug]
    src = os.path.join(SITE_SRC, spec["page"])
    html = open(src, encoding="utf-8").read()

    rendered, err = render(slug)
    if err:
        return None, err

    for pattern in spec["cut"]:
        span, err = find_region(html, pattern)
        if err:
            return None, err
        html = html[:span[0]] + "\x00SPLICE\x00" + html[span[1]:]

    # The first marker becomes the component; any others vanish.
    html = html.replace("\x00SPLICE\x00", rendered, 1).replace("\x00SPLICE\x00", "")

    dest = os.path.join(staged_root, spec["page"])
    os.makedirs(os.path.dirname(dest), exist_ok=True)
    open(dest, "w", encoding="utf-8").write(html)
    return spec["page"], None


class QuietHandler(http.server.SimpleHTTPRequestHandler):
    def log_message(self, *a):
        pass


def serve(root):
    handler = lambda *a, **k: QuietHandler(*a, directory=root, **k)  # noqa: E731
    httpd = socketserver.TCPServer(("127.0.0.1", 0), handler)
    threading.Thread(target=httpd.serve_forever, daemon=True).start()
    return httpd, httpd.server_address[1]


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    keep = "--keep" in sys.argv
    slugs = args or [s for s in SPECS
                     if os.path.exists(os.path.join(HERE, component_for(s) + ".html.tmpl"))]
    if not slugs:
        print("no rewritten components found in %s" % HERE)
        return 2

    golden = json.load(open(GOLDEN))["pages"]

    staged = tempfile.mkdtemp(prefix="verify-rewrite-")
    # The WHOLE site is staged, not just the page: the component inherits the
    # site stylesheet and the nav script, and a page served without them is a
    # different page. Symlink-free copy so nothing reaches back into the source.
    shutil.copytree(SITE_SRC, staged, dirs_exist_ok=True)

    prepared, failed = [], []
    for slug in sorted(slugs):
        page, err = splice(slug, staged)
        if err:
            failed.append((slug, err))
            print("PREP-FAIL  %-28s %s" % (slug, err))
        else:
            prepared.append((slug, page))
            print("prepared   %-28s %s" % (slug, page))

    if not prepared:
        print("\nnothing to verify")
        return 1

    httpd, port = serve(staged)
    r = Runner()
    diverged, inert = [], []
    try:
        for slug, page in prepared:
            local = "http://127.0.0.1:%d/%s" % (port, page)
            live = "%s/%s" % (LIVE, page)
            g = golden.get(live)
            if not g:
                print("NO GOLDEN  %-28s %s" % (slug, live))
                diverged.append(slug)
                continue
            try:
                got = r.capture(local)
            except Exception as e:  # noqa: BLE001
                print("CAPTURE-ERR %-27s %s" % (slug, str(e)[:70]))
                diverged.append(slug)
                continue

            allow_new = SPECS[slug].get("allow_new_keys")
            diffs, added = [], set()
            for vec, _ in VECTORS:
                for phase in ("after_input", "after_press"):
                    for fieldname in ("ids", "controls"):
                        before = g[vec][phase].get(fieldname, {})
                        after = got[vec][phase].get(fieldname, {})
                        if allow_new:
                            # NOT a blanket relaxation. Some tools cannot be
                            # covered by the platform's computed_values check at
                            # all until their controls HAVE ids -- consolidation's
                            # debt rows are addressed purely by class -- and
                            # adding an id necessarily adds a fingerprint key and
                            # renames its `controls` entry. Refusing that would
                            # mean those tools stay permanently unverifiable,
                            # which is the worse outcome.
                            #
                            # So new keys are reported and permitted, and EVERY
                            # KEY THE GOLDEN ALREADY HAD IS STILL COMPARED
                            # STRICTLY. The claim this yields is narrower and is
                            # stated as such: no existing output moved; the
                            # fingerprint gained keys. A re-baseline is then owed,
                            # and the spec entry must say why.
                            before, after, note = reconcile_renames(before, after)
                            added |= {fieldname + ":" + k for k in note}
                        d = numeric_diff(before, after)
                        if d:
                            diffs.append("   %s / %s / %s" % (vec, phase, fieldname))
                            diffs.extend(d)
            if added and not diffs:
                print("     +%d new fingerprint key(s), re-baseline owed: %s"
                      % (len(added), ", ".join(sorted(added)[:6])
                         + (" ..." if len(added) > 6 else "")))
            if diffs:
                diverged.append(slug)
                print("\nDIVERGED   %s" % slug)
                print("\n".join(diffs[:30]))
                if len(diffs) > 30:
                    print("   ... %d more lines" % (len(diffs) - 30))
            else:
                n = len(got["defaults"]["after_press"].get("ids", {}))
                print("MATCHES    %-28s %d id-fields x %d vectors"
                      % (slug, n, len(VECTORS)))
    finally:
        r.close()
        httpd.shutdown()
        if keep:
            print("\nstaged site kept at %s" % staged)
        else:
            shutil.rmtree(staged, ignore_errors=True)

    print()
    if failed:
        print("%d tool(s) could not be prepared" % len(failed))
    if diverged:
        print("%d of %d rewritten tool(s) DIVERGED from golden" % (len(diverged), len(prepared)))
        return 1
    print("all %d rewritten tool(s) reproduce their golden values exactly" % len(prepared))
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())

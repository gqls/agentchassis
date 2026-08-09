#!/usr/bin/env python3
"""defect_vectors.py — drive the CONDITIONS the equivalence gate cannot reach.

WHY THIS EXISTS, AND WHY A GREEN verify_rewrite.py IS NOT AN ANSWER.

`toolgolden.py` derives its vectors by scaling each numeric field's OWN default
(x1, x2, x0.5). That is the right default policy — it keeps every value inside
its intended domain for any tool with no per-tool configuration — but it has a
consequence nobody has to notice until it bites: THE GATE ONLY EVER VISITS
NEIGHBOURHOODS OF THE SHIPPED DEFAULTS. Two of the four defects fixed on
2026-08-03 live outside every such neighbourhood:

  - car finance breaks at APR = 0, and the APR default is 8.9. No scaling of 8.9
    is 0, so `verify_rewrite.py` passes identically before and after the fix.
  - consolidation breaks when a debt row's rate is BLANK, and the driver fills
    every numeric field it can find. Every row it builds is complete, so the
    faulty branch is never entered.

Both fixes are therefore INVISIBLE to the gate: it went green before them and
goes green after them, having asserted nothing about either. A test that passes
whether or not the code is there is not evidence, and shipping on one would be
exactly the false-green this lane's harness was built to eliminate.

So this file drives the defect conditions explicitly, with values derived BY HAND
from the arithmetic rather than captured from the tool — a captured expectation
would just re-record whatever the tool does, including the bug.

  ⚠ ITS PASS IS ONLY WORTH SOMETHING BECAUSE THE SAME CASE READS DIFFERENTLY
  WITHOUT THE FIX. `--pre-fix` renders the components from PRE_FIX_REF — a
  pinned sha, see the note on it — instead of the working tree, so every case is
  driven against the code as it was before. `--both` runs the pair and scores
  each case on whether it
  DISCRIMINATES — not on pass/fail, which is a weaker and partly wrong question
  (a case using `prefix_expect_instead` asserts "£448.024 before, £448.02 after",
  and both halves of that are a pass). A defect check nobody has watched read
  differently is a quiet test: it passes when the rule is gone.

      python3 defect_vectors.py --both

  Every case must come back PROVEN or CONTROL. VACUOUS means the case cannot see
  its own defect and needs rewriting before it is worth running again.

WHAT IT REUSES. `splice`/`SPECS`/`find_region` from verify_rewrite.py and
`CDP`/`start_chrome` from toolprobe, for the same reason verify_rewrite reuses
toolgolden: a second implementation of "stage the page and drive it" is a second
thing that can be subtly wrong, and the two would agree with each other for
exactly the reasons they were both wrong.

WHAT IT IS NOT. It is not a replacement for the equivalence gate and it does not
overlap with it. The gate asks "did anything move that should not have"; this
asks "did the one thing that SHOULD move actually move". Run both.

Usage:
  python3 defect_vectors.py             # working tree — every case must PASS
  python3 defect_vectors.py --pre-fix   # PRE_FIX_REF — the defect cases must FAIL
                                        #   here, the same_pre_and_post ones must not
  python3 defect_vectors.py --both      # both, scored as PROVEN / CONTROL / VACUOUS
  python3 defect_vectors.py --ref <sha> # score against some other point in history
  python3 defect_vectors.py --live      # drive the PRODUCTION urls (post-deploy)
  python3 defect_vectors.py --keep      # leave the staged site for inspection
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile
import time

HERE = os.path.dirname(os.path.abspath(__file__))
LANE = os.path.dirname(HERE)
sys.path.insert(0, LANE)
sys.path.insert(0, HERE)
sys.path.insert(0, os.path.abspath(
    os.path.join(LANE, "..", "webdesign_tools_repair")))

from toolprobe import CDP, start_chrome                     # noqa: E402
import verify_rewrite as VR                                 # noqa: E402

REPO = os.path.abspath(os.path.join(LANE, "..", "..", "..", ".."))

# The baseline the defect cases are scored AGAINST — the last commit before the
# 2026-08-03 fixes landed.
#
# ⚠ IT IS AN ABSOLUTE SHA, AND NOT `HEAD`, BECAUSE THIS FILE BROKE ITSELF ONCE.
# The first version read `git show HEAD:`, which was correct for exactly as long
# as the fixes were uncommitted. The moment they were committed, HEAD carried the
# fix, both sides of `--both` rendered the SAME component, every case reported
# VACUOUS — and had the scoring still been the pass/fail version it replaced, they
# would all have reported PROVEN instead, which is worse: a negative control that
# silently stops being one while still printing a reassuring word. A baseline must
# name a commit that cannot move under it. Override with --ref <gitref> to score
# these cases against any other point in history.
PRE_FIX_REF = "6e8098022"

# ── the cases ──────────────────────────────────────────────────────────────
#
# `set` is applied in order; '' means CLEAR THE FIELD, which is the whole point
# for consolidation — a blank rate is not a zero rate and the fix turns on
# telling them apart.
#
# `expect` maps element id -> exact visible text. Exact, not approximate: these
# are money figures and "about right" is how a rounding defect survives a test.
#
# EVERY EXPECTATION IS DERIVED BY HAND, from the arithmetic, and shows its
# working. Capturing them from the tool would only re-record the bug.
CASES = [
    {
        "name": "car-finance 0% APR in HP computes, and charges no interest",
        "slug": "car-finance-calculator",
        "why": (
            "The annuity formula is 0/0 at r=0, so the tool used to skip the whole "
            "calculation and leave the last rate's figures on screen. 0% is a real "
            "advertised product."),
        "set": [("price", "25000"), ("deposit", "5000"),
                ("car-term", "4"), ("car-apr", "0")],
        # principal 20000 over 48 months, no interest: 20000/48 = 416.666...
        # totalRepaid = 416.67*48 + 0 balloon + 5000 deposit = price, so interest 0.
        "expect": {"car-monthly": "£416.67", "car-total-int": "£0.00"},
        # Pre-fix the figures are whatever 8.9% last produced — the defaults.
        "prefix_expect_instead": {"car-monthly": "£496.75"},
    },
    {
        "name": "car-finance non-zero APR is UNCHANGED by the 0% branch",
        "slug": "car-finance-calculator",
        "why": (
            "The control. The fix hoisted totalRepaid/totalInt out of the r>0 "
            "branch, so it has to be shown that the r>0 path still computes what "
            "it always did."),
        "set": [("price", "25000"), ("deposit", "5000"),
                ("car-term", "4"), ("car-apr", "8.9")],
        "expect": {"car-monthly": "£496.75", "car-total-int": "£3844.08"},
        "same_pre_and_post": True,
    },
    {
        "name": "consolidation withholds the comparison when a rate is blank",
        "slug": "consolidation",
        "why": (
            "`parseFloat('') || 0` could not tell a blank rate from a 0% rate, so "
            "a half-typed row added its balance and no interest and the two sides "
            "of the comparison were computed over different debts."),
        "set": [("debt-1-bal", "5000"), ("debt-1-rate", ""), ("debt-1-months", ""),
                ("new-rate", "9.9"), ("new-term", "5")],
        "expect": {"curr-total-bal": "£5,000.00",
                   "old-int": "—", "new-monthly": "—", "new-int": "—"},
        "expect_contains": {"verdict": "Add a rate and a remaining term"},
        # Pre-fix it stated a confident verdict off an understated old side.
        "prefix_expect_instead": {"old-int": "£0.00"},
    },
    {
        "name": "consolidation still scores a COMPLETE set of debts",
        "slug": "consolidation",
        "why": (
            "The control, and the one that would catch the fix over-reaching: a "
            "fully-typed row must be unaffected by the incomplete-row branch."),
        "set": [("debt-1-bal", "5000"), ("debt-1-rate", "20"), ("debt-1-months", "36"),
                ("new-rate", "9.9"), ("new-term", "5")],
        # 5000 at 20%/yr over 36 months: r = 0.2/12, (1+r)^36 = 1.8131314,
        # m = 5000*r*1.8131314/0.8131314 = 185.817917, total 6689.445, int 1689.445.
        # ⚠ Do NOT round m before multiplying by n. Rounding to 185.82 gives
        # £1,689.52 and to 185.80 gives £1,688.80 — this expectation was written
        # from a rounded m first and the harness failed it against the tool,
        # which was right. Derive expectations at full precision and round ONCE,
        # at the end, exactly where Intl.NumberFormat does.
        "expect": {"curr-total-bal": "£5,000.00", "old-int": "£1,689.45"},
        "same_pre_and_post": True,
    },
    {
        "name": "consolidation treats a real 0% debt as complete, not as blank",
        "slug": "consolidation",
        "why": (
            "The boundary the fix turns on. A rate typed as 0 WITH a term is a "
            "genuine interest-free debt: it counts toward the balance, owes "
            "nothing, and must NOT trip the incomplete branch."),
        "set": [("debt-1-bal", "5000"), ("debt-1-rate", "0"), ("debt-1-months", "36"),
                ("new-rate", "9.9"), ("new-term", "5")],
        "expect": {"curr-total-bal": "£5,000.00", "old-int": "£0.00"},
        "same_pre_and_post": True,
    },
    {
        "name": "overpayment prints money to two decimals",
        "slug": "overpayment-calculator",
        "why": (
            "toLocaleString was called with no maximumFractionDigits, whose "
            "default is 3, so the tool's own defaults printed £448.024."),
        "set": [("bal", "15000"), ("rate", "6.5"), ("term", "5"), ("over", "50")],
        "expect": {"save-display": "£448.02", "time-display": "10"},
        "prefix_expect_instead": {"save-display": "£448.024"},
    },
    {
        "name": "loan-vs-savings names its winner in TEXT (loan side)",
        "slug": "loan-vs-savings",
        "why": (
            "The verdict was carried by the .winner colour alone, so a "
            "colour-blind reader and a screen reader both got nothing."),
        "set": [("loan-rate", "7.5"), ("save-rate", "5.0"),
                ("spare-cash", "1000"), ("tax-bracket", "0")],
        "expect": {"loan-badge": "Better option", "save-badge": "",
                   "loan-benefit": "£75.00", "save-benefit": "£50.00"},
        # Pre-fix the elements do not exist at all.
        "prefix_missing": ["loan-badge", "save-badge"],
    },
    {
        "name": "loan-vs-savings moves the badge when the winner changes",
        "slug": "loan-vs-savings",
        "why": (
            "A badge hardcoded into the winning panel would pass the case above "
            "and assert nothing. This is the one that requires it to TRACK."),
        "set": [("loan-rate", "3.0"), ("save-rate", "5.0"),
                ("spare-cash", "1000"), ("tax-bracket", "0")],
        "expect": {"loan-badge": "", "save-badge": "Better option",
                   "loan-benefit": "£30.00", "save-benefit": "£50.00"},
        "prefix_missing": ["loan-badge", "save-badge"],
    },

    # ── 0% RATE, added 2026-08-09 by the bugfix-224 session ──────────────────
    # These six tools had no case here at all, which is why the defect survived
    # a green `toolgolden 11/11` for months: its vectors scale each field's own
    # default by x1/x2/x0.5, and no scaling of 7.9 is 0. The pre-fix readings
    # below are the DEFAULT-vector answers for the two stale tools (the DOM was
    # never written, so the first paint's figures stayed) and literal £NaN for
    # the three ungated ones — all reproducible at PRE_FIX_REF, because these
    # defects predate the 08-03 batch rather than being introduced by it.
    {
        "name": "loan-repayment computes at 0% instead of leaving the last answer",
        # Staged as standard-calc because that is the slug verify_rewrite's SPECS
        # knows; the component is `tool-loan-repayment`, and the LIVE page that
        # carries it is the HOMEPAGE (/tools/standard-calc.html was retired by
        # owner ruling and 404s, though its DB rows remain).
        "slug": "standard-calc",
        "live_page": "index.html",   # standard-calc is retired and 404s; this ships on the homepage
        "why": (
            "Guarded `r > 0` with no else, so a 0% APR returned without touching "
            "the DOM: the same inputs showed two different answers depending on "
            "the route taken to them. This widget is on the HOMEPAGE."),
        "set": [("amount", "10000"), ("interest", "0"), ("years", "5")],
        # 10000 over 60 months, no interest: 10000/60 = 166.666... -> £166.67
        "expect": {"monthly-display": "£166.67", "total-interest": "£0.00",
                   "total-cost": "£10,000.00"},
        "prefix_expect_instead": {"monthly-display": "£202.29"},
    },
    {
        "name": "loan-repayment is UNCHANGED at a non-zero rate",
        "slug": "standard-calc",
        "live_page": "index.html",
        "why": "The control for the case above: the r>0 path must still compute.",
        "set": [("amount", "10000"), ("interest", "7.9"), ("years", "5")],
        "expect": {"monthly-display": "£202.29", "total-interest": "£2,137.40",
                   "total-cost": "£12,137.40"},
        "same_pre_and_post": True,
    },
    {
        "name": "compare-loans prices a 0% offer and does not invert the verdict",
        "slug": "compare-loans",
        "why": (
            "Ungated, so 0/0 printed £NaN — and `NaN < x` is false, so the "
            "verdict fell through and declared the OTHER option cheaper. A 0% "
            "loan in slot A was ALWAYS reported as the more expensive one."),
        "set": [("amt-a", "5000"), ("apr-a", "0"), ("term-a", "3"),
                ("amt-b", "5000"), ("apr-b", "10"), ("term-b", "3")],
        # 5000/36 = 138.888... -> £138.89, and no interest on an interest-free loan
        "expect": {"res-m-a": "£138.89", "res-i-a": "£0.00"},
        "expect_contains": {"verdict": "A"},
        "prefix_expect_instead": {"res-m-a": "£NaN", "res-i-a": "£NaN"},
    },
    {
        "name": "rate stress test prices the current payment at 0%",
        "slug": "interest-rate-stress-test",
        "why": (
            "Only P and n were guarded, never the rate, so #curr-pay read £NaN "
            "at 0% while #new-pay (at apr + the stress delta) stayed correct — "
            "which is why only one selector was ever visibly wrong."),
        "set": [("stress-bal", "10000"), ("stress-apr", "0"), ("stress-term", "3")],
        "expect": {"curr-pay": "£277.78"},          # 10000/36
        "prefix_expect_instead": {"curr-pay": "£NaN"},
    },
    {
        "name": "overpayment at 0% saves nothing and says so in months too",
        "slug": "overpayment-calculator",
        "why": (
            "M was NaN, so `NaN > balance` was false, the else branch ran, the "
            "balance became NaN, `NaN > 0` was false and the loop exited after "
            "ONE iteration — which is where '59 months saved' came from."),
        "set": [("bal", "15000"), ("rate", "0"), ("term", "5"), ("over", "50")],
        # 15000/60 = 250; at +50 the debt clears in 50 months, so 10 saved and
        # there is no interest to save on an interest-free debt.
        "expect": {"save-display": "£0.00", "time-display": "10"},
        "prefix_expect_instead": {"save-display": "£NaN", "time-display": "59"},
    },
    {
        "name": "early settlement at 0% is the balance, not the last rate's figure",
        "slug": "settlement-calculator",
        "why": (
            "Guarded `apr > 0` with no else. This estimate is LINEAR in the rate, "
            "so zero is simply 'no 58-day interest' — the one tool here where "
            "widening the guard is the right fix rather than the wrong one."),
        "set": [("settle-bal", "5000"), ("settle-apr", "0")],
        "expect": {"settle-result": "£5,000.00"},
        "prefix_expect_instead": {"settle-result": "£5,078.66"},
    },
    {
        "name": "consolidation prices a 0% new loan instead of quoting £0.00 a month",
        "slug": "consolidation",
        "why": (
            "newMonthly was initialised to 0 and the guard was `newR > 0`, so an "
            "interest-free consolidation was quoted at £0.00 a month — a "
            "DETERMINISTIC wrong answer, which is why the sweep that caught the "
            "other five is blind to it. And the verdict tested newN but not newR, "
            "so newTotalInterest = 0 fed 'this will SAVE you' on the page whose "
            "job is to warn about term extension."),
        "set": [("debt-1-bal", "5000"), ("debt-1-rate", "10"), ("debt-1-months", "24"),
                ("new-rate", "0"), ("new-term", "5")],
        # 5000 over 60 months at 0%: 83.33/month, no interest.
        "expect": {"new-monthly": "£83.33", "new-int": "£0.00"},
        "prefix_expect_instead": {"new-monthly": "£0.00"},
    },
    {
        "name": "consolidation still refuses a BLANK new-loan rate after the 0% fix",
        "slug": "consolidation",
        "why": (
            "The regression the 0% fix could most easily have introduced: once a "
            "zero rate prices a real loan, `parseFloat('') || 0` would price an "
            "UNFILLED form as interest-free. Blank stays distinct from zero, the "
            "same distinction the debt loop already makes."),
        "set": [("debt-1-bal", "5000"), ("debt-1-rate", "10"), ("debt-1-months", "24"),
                ("new-rate", ""), ("new-term", "5")],
        "expect": {"verdict": ""},
    },
]

DRIVE = r"""
(() => {
  const set = (id, v) => {
    const e = document.getElementById(id);
    if (!e) return 'MISSING:' + id;
    e.value = v;
    // bubbles matters: consolidation listens on the SECTION, not the input.
    e.dispatchEvent(new Event('input',  {bubbles: true}));
    e.dispatchEvent(new Event('change', {bubbles: true}));
    return null;
  };
  const missing = %s.map(p => set(p[0], p[1])).filter(Boolean);
  const read = {};
  %s.forEach(id => {
    const e = document.getElementById(id);
    read[id] = e === null ? null : (e.textContent || '').replace(/\s+/g, ' ').trim();
  });
  return JSON.stringify({missing, read});
})()
"""


def render_from(tmpl_path, schema_path, out_name):
    """Render a component with the REAL Go template engine."""
    out = os.path.join(tempfile.gettempdir(), out_name)
    r = subprocess.run(["go", "run", os.path.join(HERE, "render_tool.go"),
                        tmpl_path, schema_path, out],
                       capture_output=True, text=True, cwd=HERE)
    if r.returncode != 0:
        return None, (r.stderr or r.stdout).strip()[:400]
    return open(out, encoding="utf-8").read(), None


def ref_sources(tmpdir, ref):
    """Extract every component template+schema as committed at `ref`.

    THE NEGATIVE CONTROL HAS TO COME FROM SOMEWHERE THAT IS NOT THE WORKING TREE.
    Reverting the files to run the check and reverting them back is the obvious
    alternative and it is a bad one: it mutates a tree several other sessions are
    also committing from, and a crash in between leaves the fix silently undone.
    `git show` reads the object store and touches nothing.
    """
    got = {}
    for spec in VR.SPECS.values():
        comp = spec.get("component")
        for ext in (".html.tmpl", ".schema.json"):
            rel = os.path.relpath(os.path.join(HERE, comp + ext), REPO)
            r = subprocess.run(["git", "show", "%s:%s" % (ref, rel)],
                               capture_output=True, text=True, cwd=REPO)
            if r.returncode != 0:
                return None, "git show %s:%s failed: %s" % (ref, rel, r.stderr.strip())
            dest = os.path.join(tmpdir, comp + ext)
            open(dest, "w", encoding="utf-8").write(r.stdout)
        got[comp] = True
    return got, None


def stage(slugs, ref):
    """Copy the whole site and splice each component into its real page.

    `ref` is None for the working tree, or a git ref to render the baseline from.
    """
    staged = tempfile.mkdtemp(prefix="defect-vectors-")
    shutil.copytree(VR.SITE_SRC, staged, dirs_exist_ok=True)

    srcdir = HERE
    if ref:
        srcdir = tempfile.mkdtemp(prefix="defect-vectors-baseline-")
        _, err = ref_sources(srcdir, ref)
        if err:
            return None, None, err

    pages = {}
    for slug in sorted(set(slugs)):
        spec = VR.SPECS[slug]
        comp = spec["component"]
        rendered, err = render_from(os.path.join(srcdir, comp + ".html.tmpl"),
                                    os.path.join(srcdir, comp + ".schema.json"),
                                    comp + ".defect.html")
        if err:
            return None, None, "%s: %s" % (slug, err)

        html = open(os.path.join(VR.SITE_SRC, spec["page"]), encoding="utf-8").read()
        for pattern in spec["cut"]:
            span, err = VR.find_region(html, pattern)
            if err:
                return None, None, "%s: %s" % (slug, err)
            html = html[:span[0]] + "\x00SPLICE\x00" + html[span[1]:]
        html = html.replace("\x00SPLICE\x00", rendered, 1).replace("\x00SPLICE\x00", "")

        dest = os.path.join(staged, spec["page"])
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        open(dest, "w", encoding="utf-8").write(html)
        pages[slug] = spec["page"]
    return staged, pages, None


class Driver:
    def __init__(self):
        import socket
        s = socket.socket(); s.bind(("127.0.0.1", 0))
        self.port = s.getsockname()[1]; s.close()
        self.chrome = start_chrome(self.port, tempfile.mkdtemp(prefix="defect-vectors-"))

    def run(self, url, sets, reads):
        import urllib.request
        req = urllib.request.Request(
            "http://127.0.0.1:%d/json/new?about:blank" % self.port, method="PUT")
        tab = json.loads(urllib.request.urlopen(req, timeout=10).read())
        cdp = CDP(tab["webSocketDebuggerUrl"])
        cdp.call("Runtime.enable"); cdp.call("Page.enable")

        def ev(expr, timeout=20):
            r = cdp.call("Runtime.evaluate",
                         {"expression": expr, "returnByValue": True,
                          "awaitPromise": True}, timeout=timeout)
            return r.get("result", {}).get("result", {}).get("value")

        cdp.call("Page.navigate", {"url": url})
        # Same settle discipline as toolgolden: readyState is not enough, a
        # half-parsed document drives cleanly and answers £0.00 for everything.
        for _ in range(80):
            if ev("document.readyState", timeout=10) == "complete":
                break
            time.sleep(0.1)
        last = None
        for _ in range(40):
            cur = ev("document.querySelectorAll('script').length + ':' + "
                     "document.documentElement.outerHTML.length", timeout=10)
            if cur is not None and cur == last:
                break
            last = cur
            time.sleep(0.15)

        out = ev(DRIVE % (json.dumps(sets), json.dumps(reads)))
        cdp.close() if hasattr(cdp, "close") else None
        return json.loads(out) if out else {"missing": ["<no result>"], "read": {}}


def check(case, got, pre_fix):
    """Compare one case's readings against its expectations. Returns a list of
    failure strings — empty means the case passed."""
    bad = []
    exact = dict(case.get("expect", {}))
    if pre_fix:
        for k, v in case.get("prefix_expect_instead", {}).items():
            exact[k] = v
    for eid, want in sorted(exact.items()):
        have = got["read"].get(eid)
        if have != want:
            bad.append("%s = %r, want %r" % (eid, have, want))
    for eid, frag in sorted(case.get("expect_contains", {}).items()):
        have = got["read"].get(eid) or ""
        if frag not in have:
            bad.append("%s = %r, want it to contain %r" % (eid, have, frag))
    return bad


def run_side(ref, keep, live=False):
    """Drive every case once.

    `live=True` drives the PRODUCTION urls instead of a staged copy. That is the
    only run that can state anything about what the public is served: every other
    mode splices a locally-rendered component into a local copy of the page, which
    proves the component and says nothing about whether it reached the wire.
    `ref` is None for the working tree.
    """
    pre_fix = ref is not None and not live
    label = ("LIVE %s" % VR.LIVE if live else
             "%s (baseline)" % ref if pre_fix else "working tree")

    httpd = None
    if live:
        pages = {c["slug"]: VR.SPECS[c["slug"]]["page"] for c in CASES}
        port = None
    else:
        slugs = [c["slug"] for c in CASES]
        staged, pages, err = stage(slugs, ref)
        if err:
            print("STAGE FAILED  %s: %s" % (label, err))
            return None
        httpd, port = VR.serve(staged)
    drv = Driver()

    print("\n== %s ==" % label)
    results = []
    for case in CASES:
        # `live_page` exists because a component's staging page and its LIVE page
        # are not always the same file. `tool-loan-repayment` is staged as
        # standard-calc (the slug verify_rewrite's SPECS knows) but that page was
        # retired by owner ruling and now 404s, while the component itself ships
        # on the HOMEPAGE. Without the override, --live drove the dead URL and
        # reported MISSING on every element — the defect case AND its control
        # failing identically, which is the tell for a harness pointed at the
        # wrong page rather than a broken tool.
        page_path = case.get("live_page", pages[case["slug"]]) if live \
            else pages[case["slug"]]
        url = ("%s/%s" % (VR.LIVE, page_path) if live
               else "http://127.0.0.1:%d/%s" % (port, page_path))
        reads = sorted(set(list(case.get("expect", {}))
                           + list(case.get("expect_contains", {}))
                           + list(case.get("prefix_expect_instead", {}))
                           + case.get("prefix_missing", [])))
        got = drv.run(url, case["set"], reads)
        bad = check(case, got, pre_fix)

        if pre_fix and case.get("prefix_missing"):
            absent = [i for i in case["prefix_missing"] if got["read"].get(i) is None]
            if absent:
                bad = bad or ["elements absent at %s: %s" % (ref, ", ".join(absent))]

        ok = not bad
        results.append((case["name"], ok, bad, got["read"]))
        print("  %-4s %s" % ("PASS" if ok else "FAIL", case["name"]))
        for b in bad:
            print("         %s" % b)
        if got["missing"]:
            print("         could not set: %s" % ", ".join(got["missing"]))

    if httpd is not None:
        httpd.shutdown()
    try:
        drv.chrome.terminate()
    except Exception:
        pass
    if not live:
        if keep:
            print("  staged site kept at %s" % staged)
        else:
            shutil.rmtree(staged, ignore_errors=True)
    return results


def main():
    keep = "--keep" in sys.argv
    both = "--both" in sys.argv
    pre = "--pre-fix" in sys.argv
    live = "--live" in sys.argv
    ref = PRE_FIX_REF
    if "--ref" in sys.argv:
        ref = sys.argv[sys.argv.index("--ref") + 1]

    if both:
        post = run_side(None, keep)
        prev = run_side(ref, keep)
        if post is None or prev is None:
            return 2
        print("\n== the pair ==")
        # THE VERDICT IS A DIFF OF READINGS, NOT OF PASS/FAIL, and the first
        # version of this got it wrong. A case may legitimately PASS on both
        # sides — `prefix_expect_instead` exists precisely so a case can assert
        # "£448.024 before, £448.02 after", and both halves of that are a pass.
        # Scoring on pass/fail called that VACUOUS when it was the most exactly
        # specified case in the file. The question a defect case has to answer is
        # whether it DISCRIMINATES: does any asserted element read differently
        # with the fix and without it? Nothing else establishes that a green run
        # is evidence rather than a formality.
        rc = 0
        for (name, ok_post, _, read_post), (_, ok_pre, _, read_pre) in zip(post, prev):
            case = next(c for c in CASES if c["name"] == name)
            moved = sorted(k for k in read_post
                           if read_post.get(k) != read_pre.get(k))

            if case.get("same_pre_and_post"):
                held = ok_post and ok_pre and not moved
                verdict, detail = ("CONTROL", "unmoved") if held else (
                    "CONTROL BROKEN", "moved: " + ", ".join(moved) if moved
                    else "does not pass both sides")
                rc |= 0 if held else 1
            elif ok_post and moved:
                verdict, detail = "PROVEN", "discriminates on " + ", ".join(moved)
            elif ok_post:
                verdict, detail = "VACUOUS", "reads identically with and without the fix"
                rc |= 1
            else:
                verdict, detail = "FAILING", "does not pass against the working tree"
                rc |= 1
            print("  %-15s %s\n%18s%s" % (verdict, name, "", detail))
        print("\n%s" % ("every defect case DISCRIMINATES (it reads differently "
                        "without the fix) and every control is unmoved" if rc == 0
                        else "NOT PROVEN — read the pair above"))
        return rc

    if live:
        results = run_side(None, keep, live=True)
        if results is None:
            return 2
        failed = [n for n, ok, _, _ in results if not ok]
        if failed:
            print("\n%d of %d case(s) FAILED ON THE LIVE SITE" % (len(failed), len(CASES)))
            return 1
        print("\nall %d case(s) pass against the SERVED pages" % len(CASES))
        return 0

    results = run_side(ref if pre else None, keep)
    if results is None:
        return 2
    failed = [n for n, ok, _, _ in results if not ok]
    if pre:
        expected_fail = [c["name"] for c in CASES if not c.get("same_pre_and_post")]
        still_passing = [n for n in expected_fail if n not in failed]
        if still_passing:
            print("\n%d defect case(s) PASS against %s — they assert nothing:"
                  % (len(still_passing), ref))
            for n in still_passing:
                print("  %s" % n)
            return 1
        print("\nevery defect case fails at %s, as it must for its pass to mean "
              "anything" % ref)
        return 0
    if failed:
        print("\n%d of %d case(s) FAILED" % (len(failed), len(CASES)))
        return 1
    print("\nall %d case(s) pass" % len(CASES))
    return 0


if __name__ == "__main__":
    sys.exit(main())

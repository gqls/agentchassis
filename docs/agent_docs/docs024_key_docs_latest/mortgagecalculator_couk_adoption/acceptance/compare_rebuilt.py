#!/usr/bin/env python3
"""compare_rebuilt.py — REPLAY the golden's recorded inputs into each REBUILT
mortgagecalculator tool, and diff the answers on IDENTICAL inputs.

WHY THIS FILE WAS REWRITTEN (2026-08-08 night). The previous version reused
toolgolden's capture on the rebuilt page, and toolgolden derives every driven
value by scaling the page's OWN markup `value` attributes (fixed 1000 into a
field with none). It goldens a page against ITSELF. So the old compare fed the
original ITS defaults and the rebuild ITS OWN different defaults, and reported
two correct answers to two different questions as six broken calculators
(WRONG_CALLS.md 2026-08-08; LANDMINES "toolgolden … cross-page compare").

THE FIX, per the owner ruling of 2026-08-08 §0.2 ("the checker's job is to
prove results don't differ on identical inputs, and to catch wrong results"):
the golden already records the fill plan the capture actually followed — per
control `sel`/`action`/`value`, recorded for --emit-criteria replay. This
version drives the REBUILT page with those recorded LITERAL values, by id:
  fill    -> set the exact recorded value, dispatch input+change, read it back;
  select  -> select BY VALUE (never by index — the rebuilds reorder options);
  click   -> ensure-checked for checkbox/radio (absolute state, not a toggle).
The press still uses toolgolden's button heuristic, because none of the
original press buttons carries an id (golden `pressed.sel` is null on all 12) —
the press is not an input value; the fills/selects are what must be absolute.

HOW TO READ THE REPORT. Only ids present on BOTH sides are judged, on
`after_press`, numerically:
  VERIFIED         every shared numeric output equal on identical inputs.
                   Rounding-equal counts as equal: |a-b| within half a unit of
                   the COARSER side's displayed precision (original repayment
                   prints £1,390 where the rebuild prints £1,389.58 — same
                   answer, one shows pence).
  DIVERGED         a shared id holds a genuinely different number -> compute
                   independently and judge WHICH side is right (ruling §0.1:
                   correctness, not fidelity; both-right-differently -> supply
                   both, ruling §0.5).
  NEEDS-JUDGEMENT  a shared id holds a different COUNT of numbers (the rebuilt
                   copy may phrase a breakdown with more figures) — read it,
                   don't trust an automatic verdict either way.
  REPLAY-FAIL      an input could not be driven to the recorded value (missing
                   id, missing option value, value refused). The tool was NOT
                   judged — fix the replay first; a partial drive would compare
                   answers to different questions again.
One-sided ids are reported as a count only: a renamed id is a fence-authoring
problem, not arithmetic.

Usage:
  python3 compare_rebuilt.py                  # the 9 id-aligned tools
  python3 compare_rebuilt.py repayment simple # by original slug (stragglers
                                              # affordability/fact-finder/
                                              # portfolio only run when named)
"""
import json
import os
import re
import sys
import time

HERE = os.path.dirname(os.path.abspath(__file__))
LC_LANE = os.path.join(HERE, "..", "..", "loancalculator_couk")
sys.path.insert(0, os.path.abspath(LC_LANE))
from toolgolden import Runner, SNAP_JS, PRESS_JS  # noqa: E402
from toolprobe import CDP  # noqa: E402  (on sys.path via toolgolden's import)

GOLDEN = os.path.join(HERE, "GOLDEN_2026-08-05_original_tools.json")
BASE = "https://mortgagecalculator.co.uk"

# original file-form URL -> rebuilt directory-form URL (pages.url per pages table)
MAPPING = {
    "/repayment.html":       "/tools/repayment/index.html",
    "/affordability.html":   "/tools/affordability/index.html",
    "/simple.html":          "/tools/simple/index.html",
    "/overpayment.html":     "/tools/overpayment/index.html",
    "/stamp-duty.html":      "/tools/stamp-duty/index.html",
    "/bridging-loan.html":   "/tools/bridging-loan/index.html",
    "/equity-release.html":  "/tools/equity-release/index.html",
    "/fee-analyser.html":    "/tools/fee-analyser/index.html",
    "/rate-forecaster.html": "/tools/rate-forecaster/index.html",
    "/portfolio.html":       "/tools/portfolio/index.html",
    "/investor.html":        "/investor/index.html",
    "/fact-finder.html":     "/games/fact-finder/index.html",
}

# Not yet id-aligned (handoff 08-08b §2) — a replay against renamed ids would
# report REPLAY-FAIL noise, not information. Run them by naming them.
STRAGGLERS = {"affordability", "fact-finder", "portfolio"}

# Absolute-input driver. Differences from toolgolden's DRIVE_JS are the point:
# values come from the recorded plan, selects go BY VALUE, checkables are set
# (not toggled), and every step reads back what the control now holds so a
# refused value cannot silently become "the rebuild answered differently".
REPLAY_JS = r"""
(steps) => {
  const out = [];
  const fire = e => ['input', 'change'].forEach(
      ev => e.dispatchEvent(new Event(ev, {bubbles: true})));
  for (const s of steps) {
    if (!s.sel) { out.push({sel: null, action: s.action, err: 'null-sel'}); continue; }
    const e = document.querySelector(s.sel);
    if (!e) { out.push({sel: s.sel, action: s.action, err: 'missing'}); continue; }
    const t = (e.type || e.tagName).toLowerCase();
    if (s.action === 'fill') {
      e.value = s.value; fire(e);
      out.push({sel: s.sel, action: 'fill', want: s.value, got: String(e.value),
                ok: String(e.value) === String(s.value)});
    } else if (s.action === 'select') {
      const opts = Array.from(e.options || []).map(o => String(o.value));
      if (!opts.includes(String(s.value))) {
        out.push({sel: s.sel, action: 'select', want: s.value, err: 'no-option',
                  have: opts.slice(0, 12)});
        continue;
      }
      e.value = s.value; fire(e);
      out.push({sel: s.sel, action: 'select', want: s.value, got: String(e.value),
                ok: String(e.value) === String(s.value)});
    } else if (s.action === 'click') {
      if (t === 'checkbox' || t === 'radio') {
        if (!e.checked) e.click();
        out.push({sel: s.sel, action: 'click', ok: e.checked === true});
      } else {
        e.click();
        out.push({sel: s.sel, action: 'click', ok: true});
      }
    }
  }
  return JSON.stringify(out);
}
"""

# The stubs/settle below are duplicated from toolgolden.Runner.capture (they are
# closures there, not importable). Same reasons apply verbatim: a modal blocks
# the renderer silently; storage must be cleared BEFORE the scripts read it; a
# readyState poll alone can hand back a DOM whose inline script has not parsed.
DIALOG_STUB = ("window.confirm=()=>true;window.alert=()=>undefined;"
               "window.prompt=()=>'';'stubbed'")
CLEAR_STATE = ("try{localStorage.clear();sessionStorage.clear()}catch(e){};"
               "'cleared'")


def replay_capture(runner, url, gpage):
    """Drive `url` with gpage's recorded plan; return {vec: {...}} like capture."""
    cdp = CDP(runner._tab()["webSocketDebuggerUrl"])
    cdp.call("Runtime.enable"); cdp.call("Page.enable")

    def ev(expr, timeout=20):
        r = cdp.call("Runtime.evaluate",
                     {"expression": expr, "returnByValue": True,
                      "awaitPromise": True}, timeout=timeout)
        return r.get("result", {}).get("result", {}).get("value")

    def settle():
        for _ in range(80):
            if ev("document.readyState", timeout=10) == "complete":
                break
            time.sleep(0.1)
        last = None
        for _ in range(40):
            cur = ev("document.querySelectorAll('script').length + ':' + "
                     "document.documentElement.outerHTML.length", timeout=10)
            if cur is not None and cur == last:
                return cur
            last = cur
            time.sleep(0.15)
        return last

    page = {}
    for vec, g in sorted(gpage.items()):
        cdp.call("Page.navigate", {"url": url})
        settle()
        ev(CLEAR_STATE, timeout=10)
        cdp.call("Page.navigate", {"url": url})
        shape = settle()
        ev(DIALOG_STUB, timeout=10)
        # Setup rows first, exactly as emit_criteria replays them: twice per
        # recorded add-button (SETUP_JS clicked twice).
        setup_fails = []
        for item in g.get("added", []):
            if not item.get("sel"):
                setup_fails.append("setup button %r has no id" % item.get("label"))
                continue
            ev("(() => { const b = document.querySelector(%s); "
               "if (b) { b.click(); b.click(); } return 'ok' })()"
               % json.dumps(item["sel"]), timeout=10)
        steps = ev("(%s)(%s)" % (REPLAY_JS.strip(), json.dumps(g["drove"])),
                   timeout=20)
        steps = json.loads(steps) if steps else []
        pressed = {"label": "none", "sel": None}
        if (g.get("pressed") or {}).get("label") not in (None, "none"):
            p = ev(PRESS_JS, timeout=20)
            pressed = json.loads(p) if p else pressed
        snap = ev(SNAP_JS)
        page[vec] = {
            "dom_shape": shape,
            "setup_fails": setup_fails,
            "steps": steps,
            "pressed": pressed,
            "after_press": json.loads(snap) if snap else {},
        }
    return page


# ── numeric judgement ───────────────────────────────────────────────────────
NUM = re.compile(r"-?\d[\d,]*\.?\d*")


def _text(composite):
    """Strip the trailing |display (and |canvas) the fingerprint appends."""
    if composite is None:
        return ""
    if composite.endswith("|canvas"):
        composite = composite[: -len("|canvas")]
    return composite.rsplit("|", 1)[0] if "|" in composite else composite


def _parse(tok):
    t = tok.replace(",", "")
    dec = len(t.split(".")[1]) if "." in t else 0
    return float(t), dec


def _pair_ok(a, b):
    """Equal within half a unit of the COARSER displayed precision."""
    fa, da = _parse(a)
    fb, db = _parse(b)
    d = min(da, db)
    return abs(fa - fb) <= 0.5 * 10 ** (-d) + 1e-9


def judge_vector(g_ids, r_ids):
    # Categories, in verdict weight order:
    #   diverged    both sides answered, numbers genuinely differ  -> conviction
    #   structural  same-count assumption broke (different number
    #               counts, both sides non-empty)                  -> eyeball
    #   truncated   the 200-char fingerprint slice cut the text —
    #               the two sides show different WINDOWS of (e.g.)
    #               an amortization table, so token counts can
    #               never align; unjudgeable by machine            -> eyeball
    #   no_answer   one side has numbers, the other has none (the
    #               rebuild validated/refused an input the original
    #               accepts, or vice versa) — a DOMAIN difference,
    #               not wrong arithmetic                           -> domain
    out = {"diverged": [], "structural": [], "text": [], "rounding": [],
           "truncated": [], "no_answer": []}
    for k in sorted(set(g_ids) & set(r_ids)):
        gv, rv = g_ids[k], r_ids[k]
        if gv == rv:
            continue
        gt, rt = _text(gv), _text(rv)
        ga, ra = NUM.findall(gt), NUM.findall(rt)
        if ga == ra:
            out["text"].append((k, gv, rv))
        elif len(gt) >= 200 or len(rt) >= 200:
            out["truncated"].append((k, gv, rv))
        elif bool(ga) != bool(ra):
            out["no_answer"].append((k, gv, rv))
        elif len(ga) != len(ra):
            out["structural"].append((k, gv, rv))
        elif all(_pair_ok(a, b) for a, b in zip(ga, ra)):
            out["rounding"].append((k, gv, rv))
        else:
            out["diverged"].append((k, gv, rv))
    out["only_golden"] = sorted(set(g_ids) - set(r_ids))
    out["only_rebuilt"] = sorted(set(r_ids) - set(g_ids))
    return out


def _show(rows, cap=8):
    for k, gv, rv in rows[:cap]:
        print("         %-26s golden  %r" % (k, gv))
        print("         %-26s rebuilt %r" % ("", rv))
    if len(rows) > cap:
        print("         ... and %d more" % (len(rows) - cap))


def main():
    args = sys.argv[1:]
    verbose = "-v" in args
    only = set(a for a in args if a != "-v")
    golden = json.load(open(GOLDEN))["pages"]
    r = Runner()
    verdicts = {}
    try:
        for old, new in sorted(MAPPING.items()):
            slug = old.strip("/").replace(".html", "")
            if only and slug not in only:
                continue
            if not only and slug in STRAGGLERS:
                print("SKIP      %-16s not id-aligned yet — run by name" % slug)
                continue
            g = golden.get(BASE + old)
            if not g:
                verdicts[slug] = "NO-GOLDEN"
                print("NO-GOLDEN %s" % old)
                continue
            try:
                rp = replay_capture(r, BASE + new, g)
            except Exception as e:
                verdicts[slug] = "CAPTURE-ERROR"
                print("CAPTURE-ERROR %-16s %s" % (slug, str(e)[:90]))
                continue

            fails = []
            rows = {"diverged": [], "structural": [], "no_answer": [],
                    "truncated": [], "rounding": [], "text": []}
            notes = {"rounding": 0, "text": 0, "one_sided": 0}
            for vec in sorted(g):
                rv = rp[vec]
                fails += ["%s: %s" % (vec, f) for f in rv["setup_fails"]]
                fails += ["%s: %s %s -> %s" % (vec, s.get("action"), s.get("sel"),
                                               s.get("err") or ("got %r" % s.get("got")))
                          for s in rv["steps"] if s.get("err") or not s.get("ok")]
                res = judge_vector(g[vec]["after_press"].get("ids", {}),
                                   rv["after_press"].get("ids", {}))
                for kind in rows:
                    rows[kind] += [(vec, row) for row in res[kind]]
                notes["rounding"] += len(res["rounding"])
                notes["text"] += len(res["text"])
                notes["one_sided"] = max(
                    notes["one_sided"],
                    len(res["only_golden"]) + len(res["only_rebuilt"]))

            def dump(kind, cap=8):
                for vec, row in rows[kind][:cap]:
                    print("      vector %s:" % vec)
                    _show([row], cap=1)

            if fails:
                verdicts[slug] = "REPLAY-FAIL"
                print("\nREPLAY-FAIL %-16s NOT judged — %d step(s) did not take:"
                      % (slug, len(fails)))
                for f in fails[:8]:
                    print("         %s" % f)
            elif rows["diverged"]:
                verdicts[slug] = "DIVERGED"
                print("\nDIVERGED  %-16s identical inputs, different numbers:" % slug)
                dump("diverged", cap=12)
            elif rows["structural"]:
                verdicts[slug] = "NEEDS-JUDGEMENT"
                print("\nNEEDS-JUDGEMENT %-16s number COUNTS differ — read, don't assume:"
                      % slug)
                dump("structural")
            elif rows["no_answer"]:
                verdicts[slug] = "DOMAIN-DIFF"
                print("\nDOMAIN-DIFF %-16s one side answered, the other refused "
                      "(input-domain difference, not arithmetic):" % slug)
                dump("no_answer")
            else:
                verdicts[slug] = "VERIFIED"
                print("VERIFIED  %-16s identical inputs -> same answers "
                      "(%d rounding-equal, %d wording-only, ≤%d one-sided ids)"
                      % (slug, notes["rounding"], notes["text"], notes["one_sided"]))
            if rows["truncated"]:
                print("      note: %d truncated id(s) (200-char fingerprint slice) "
                      "not machine-judged — eyeball:" % len(rows["truncated"]))
                dump("truncated", cap=2)
            if verbose and rows["rounding"]:
                print("      rounding-equal detail (-v):")
                dump("rounding", cap=20)
    finally:
        r.close()

    print("\n== summary ==")
    for slug, v in sorted(verdicts.items()):
        print("   %-16s %s" % (slug, v))
    bad = [s for s, v in verdicts.items() if v not in ("VERIFIED",)]
    if bad:
        print("\n%d tool(s) not verified: %s" % (len(bad), ", ".join(sorted(bad))))
        sys.exit(1)
    print("\nall replayed tools reproduce the original answers on identical inputs")


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""invariants.py — the class C tools: what CAN be checked when nothing is right.

FIVE OF THE 23 TOOLS HAVE NO EXTERNAL RIGHT ANSWER. A credit "health" score out
of 100, a damage-charge verdict from four checkboxes, a mortgage "approval
score" out of 420 — these are the site's own scoring models. There is no HMRC
table and no annuity formula to check them against, and inventing one would be
worse than not checking: it would dress a preference up as arithmetic and then
defend it. So this file does NOT claim to validate their answers, and the report
must say so in those words.

WHAT IT DOES INSTEAD. It checks properties that must hold whatever the model is:

  MONOTONICITY   answering strictly worse must never score strictly better.
                 This is the one property every scoring tool on this site
                 asserts by construction — the buttons are literally labelled
                 with the direction — and it is violable by a sign error, a
                 mis-wired handler or an off-by-one in a weights array.
  BOUNDS         a percentage lands in [0, 100]; a meter fill does not exceed
                 its track; a verdict is one of the strings the tool can emit.
  DETERMINISM    the same answers give the same result.
  ROUND-TRIP     state that survives a reload comes back as what went in.
  AGGREGATION    where a "scoring" tool also does plain arithmetic — portfolio
                 sums values, subtracts mortgages, divides for LTV — that part
                 IS checkable, and it is checked here against the definitions,
                 not against the page.

A FAILING INVARIANT IS A REAL DEFECT; a passing one is much weaker evidence than
a passing oracle, and the report labels it that way rather than letting five
green ticks read like fifteen.

Usage:
  python3 invariants.py [--json out.json] [--tools damage,portfolio]
"""
import argparse
import json
import sys
import time

import oracles as O
from oracle_driver import BASE, Driver, DriveError, PageLoadError, parse_money

RESULTS = []

# Sentinel for "the element exists but the user cannot see it" — distinct from
# None (no such element) and from '' (present, displayed, and genuinely empty).
HIDDEN_SENTINEL = "__HIDDEN__"


def record(tool, prop, name, status, detail=""):
    RESULTS.append({"tool": tool, "property": prop, "name": name,
                    "status": status, "detail": detail})
    icon = {"PASS": "  ok  ", "FAIL": " FAIL ", "N/A": " n/a  "}[status]
    print("  [%s] %-11s %-46s %s" % (icon, prop, name[:46], detail[:110]))


def visible_text(d, sel):
    """Text of an element ONLY IF it is displayed.

    A verdict box on these pages is populated up front and revealed later, so
    `textContent` alone reports a verdict the user cannot see — damage-checker's
    #damage-verdict already holds "Potential Charges Detected" on a page where
    nothing has been ticked. Asserting on hidden text would convict a tool of
    saying something it never said.
    """
    js = """
    (() => {
      const e = document.querySelector(%s);
      if (!e) return null;
      const cs = getComputedStyle(e);
      if (cs.display === 'none' || cs.visibility === 'hidden') return %s;
      return (e.textContent || '').replace(/\\s+/g, ' ').trim();
    })()
    """ % (json.dumps(sel), json.dumps(HIDDEN_SENTINEL))
    v = d.ev(js)
    return None if v == HIDDEN_SENTINEL else v


# ---------------------------------------------------------------------------


def check_damage_checker(d):
    """4 checkboxes -> a charge verdict. More damage must never read better."""
    tool = "loans/damage-checker.html"
    url = BASE + tool
    boxes = ["#dmg-1", "#dmg-2", "#dmg-3", "#dmg-4"]
    print("\n=== %s   [class C — checkbox verdict, no external right answer]" % tool)

    d.goto(url)
    none_ticked = visible_text(d, "#damage-verdict")
    record(tool, "BOUNDS", "verdict hidden until something is ticked",
           "PASS" if none_ticked is None else "FAIL",
           "nothing ticked, verdict %s" %
           ("is hidden" if none_ticked is None else "READS %r" % none_ticked[:60]))

    seen = []
    for k in range(len(boxes) + 1):
        d.goto(url)
        for b in boxes[:k]:
            d.set(b, True)
        seen.append((k, visible_text(d, "#damage-verdict")))

    # Monotone in the ONLY direction a checkbox verdict can be monotone: once
    # the tool has started warning, adding more damage must not stop it.
    warned = [k for k, t in seen if t and "no charge" not in (t or "").lower()]
    ok = warned == list(range(min(warned), len(boxes) + 1)) if warned else False
    record(tool, "MONOTONE", "warning never withdrawn as damage is added",
           "PASS" if ok else "FAIL",
           "ticked->warned: %s" % ", ".join("%d:%s" % (k, "warn" if t else "-")
                                            for k, t in seen))

    d.goto(url)
    for b in boxes:
        d.set(b, True)
    a = visible_text(d, "#damage-verdict")
    d.goto(url)
    for b in reversed(boxes):
        d.set(b, True)
    b_ = visible_text(d, "#damage-verdict")
    record(tool, "DETERMINISM", "same boxes ticked in the reverse order",
           "PASS" if a == b_ else "FAIL",
           "identical" if a == b_ else "%r vs %r" % ((a or "")[:40], (b_ or "")[:40]))


def check_credit_health(d):
    """A 5-step wizard; each button carries its own score contribution."""
    tool = "loans/credit-health-check.html"
    url = BASE + tool
    print("\n=== %s   [class C — wizard scoring, no external right answer]" % tool)

    # Walk the wizard by always taking the FIRST / LAST offered answer at each
    # step. The button labels ("Yes" / "No or Not Sure", "Zero" / "Two or more")
    # order the options best-to-worst, so "always first" is the strong path and
    # "always last" the weak one. That ordering is the page's own claim, read
    # from the visible labels rather than from the handlers' arguments.
    # RESET is excluded from the clickable set, and this is a correction, not a
    # precaution. The first version walked steps 1..5 and clicked the first
    # button it found in each: step-5 is the RESULT panel, whose only button is
    # "Start Over" -> location.reload(). So the walk answered four questions,
    # reached the verdict, and immediately destroyed it — the check reported
    # 'Calculating...' and a non-deterministic replay, and I nearly filed the
    # tool's determinism as a defect. `toolgolden.PRESS_JS` already excludes
    # reset-ish buttons and names this exact failure in its comment; reusing the
    # browser harness while leaving its hard-won exclusion behind is how a
    # lesson gets re-learned at full price.
    RESET = "/reset|clear|start over|start again|cancel/i"

    def walk(pick):
        d.goto(url)
        for step in range(1, 6):
            js = """
            (() => {
              const s = document.querySelector('#step-%d');
              if (!s || getComputedStyle(s).display === 'none') return 'GONE';
              const b = Array.from(s.querySelectorAll('button'))
                .filter(e => !e.disabled)
                .filter(e => !%s.test((e.textContent || '') + ' ' + (e.id || '')));
              if (!b.length) return 'NOBUTTONS';
              const t = %s;
              t.click();
              return t.textContent.trim().slice(0, 24);
            })()
            """ % (step, RESET, "b[0]" if pick == "first" else "b[b.length-1]")
            if d.ev(js) in ("GONE", "NOBUTTONS"):
                break
        txt = visible_text(d, "#rating-text")
        fill = d.ev("(()=>{const e=document.querySelector('#meter-fill');"
                    "return e?e.style.width||getComputedStyle(e).width:null})()")
        return txt, fill

    strong = walk("first")
    weak = walk("last")
    record(tool, "DETERMINISM", "wizard reaches a verdict at all", "PASS"
           if strong[0] else "FAIL", "strong path -> %r" % (strong[0],))

    def pct(v):
        try:
            return float(str(v).replace("%", "").replace("px", ""))
        except Exception:
            return None

    ps, pw = pct(strong[1]), pct(weak[1])
    if ps is None or pw is None:
        record(tool, "MONOTONE", "best answers score at least as well as worst",
               "N/A", "meter width unreadable (%r / %r)" % (strong[1], weak[1]))
    else:
        record(tool, "MONOTONE", "best answers score at least as well as worst",
               "PASS" if ps >= pw else "FAIL",
               "strong %.1f vs weak %.1f (%r vs %r)"
               % (ps, pw, strong[0], weak[0]))
        record(tool, "BOUNDS", "meter fill within 0-100%",
               "PASS" if 0 <= ps <= 100 and 0 <= pw <= 100 else "FAIL",
               "strong %.1f, weak %.1f" % (ps, pw))

    again = walk("first")
    record(tool, "DETERMINISM", "identical path replayed",
           "PASS" if again == strong else "FAIL",
           "identical" if again == strong else "%r vs %r" % (strong, again))


def check_fact_finder(d):
    """11 scored selects + 2 weight sliders -> an 'Approval Score'."""
    tool = "mortgages/fact-finder.html"
    url = BASE + tool
    print("\n=== %s   [class C — scorecard, no external right answer]" % tool)

    sels = ["#q_address", "#q_vote", "#q_empType", "#q_probation", "#q_jobTime",
            "#q_accounts", "#q_deposit", "#q_overdraft", "#q_gambling",
            "#q_missed", "#q_ccj"]

    def drive(which):
        """Set every select to its highest- or lowest-VALUE option.

        The option VALUES are the score contributions and they are visible in
        the DOM, so 'best' and 'worst' are determined without reading the
        scoring function — which is the only part of this tool that could be
        called its model.
        """
        d.goto(url)
        js = """
        (() => {
          const out = [];
          %s.forEach(s => {
            const e = document.querySelector(s);
            if (!e) return;
            const opts = Array.from(e.options).map(o => parseFloat(o.value))
                              .map(v => isFinite(v) ? v : null);
            if (opts.some(v => v === null)) { out.push(s + ':nonnumeric'); return; }
            let idx = 0;
            opts.forEach((v, i) => {
              if (%s) idx = i;
            });
            e.selectedIndex = idx;
            ['input','change'].forEach(ev => e.dispatchEvent(new Event(ev, {bubbles:true})));
            out.push(s + '=' + e.value);
          });
          return out.join(' ');
        })()
        """ % (json.dumps(sels),
               "v > opts[idx]" if which == "best" else "v < opts[idx]")
        driven = d.ev(js)
        return driven, visible_text(d, "#displayScore"), visible_text(d, "#scoreBadge")

    db, best, badge_b = drive("best")
    dw, worst, badge_w = drive("worst")
    try:
        vb, vw = parse_money(best), parse_money(worst)
        record(tool, "MONOTONE", "all-best scores above all-worst",
               "PASS" if vb > vw else "FAIL",
               "best %.0f (%s) vs worst %.0f (%s)" % (vb, badge_b, vw, badge_w))
        record(tool, "BOUNDS", "score does not go negative on the worst answers",
               "PASS" if vw >= 0 else "FAIL", "worst score = %.0f" % vw)
    except Exception as e:                                  # noqa: BLE001
        record(tool, "MONOTONE", "all-best scores above all-worst", "N/A", str(e))

    db2, best2, _ = drive("best")
    record(tool, "DETERMINISM", "same answers replayed",
           "PASS" if best2 == best else "FAIL",
           "identical (%s)" % best if best2 == best else "%r vs %r" % (best, best2))


def check_application_tracker(d):
    """Checklist + notes persisted to localStorage."""
    tool = "loans/application-tracker.html"
    url = BASE + tool
    print("\n=== %s   [class C — checklist + localStorage, no arithmetic]" % tool)
    boxes = ["#doc-id", "#doc-address", "#doc-income", "#doc-bank", "#doc-credit"]

    def fill(n):
        js = """
        (() => {
          const e = document.querySelector('#progress-fill');
          return e ? (e.style.width || getComputedStyle(e).width) : null;
        })()
        """
        d.ev(js)
        return d.ev(js)

    d.goto(url)
    widths = []
    for k in range(len(boxes) + 1):
        if k:
            d.set(boxes[k - 1], True)
        widths.append(fill(k))

    def pct(v):
        try:
            return float(str(v).replace("%", "").replace("px", ""))
        except Exception:
            return None

    nums = [pct(w) for w in widths]
    if any(n is None for n in nums):
        record(tool, "MONOTONE", "progress rises with each item ticked", "N/A",
               "progress width unreadable: %r" % (widths,))
    else:
        ok = all(nums[i] <= nums[i + 1] + 1e-9 for i in range(len(nums) - 1))
        record(tool, "MONOTONE", "progress rises with each item ticked",
               "PASS" if ok else "FAIL", "widths %s" % nums)
        record(tool, "BOUNDS", "progress ends at 100% with everything ticked",
               "PASS" if abs(nums[-1] - 100.0) < 0.5 else "FAIL",
               "final %.1f%%" % nums[-1])

    # Round trip. Note the driver clears storage on goto() by design, so the
    # reload here is done WITHOUT that clear — otherwise the test would be
    # measuring the harness, not the tool.
    note = "oracle round-trip probe 2026-08-08"
    d.set("#user-notes", note)
    # WAIT FOR THE TOOL'S OWN CONFIRMATION before reloading. The notes field
    # saves on a 1-second debounce and reports its state in #save-status
    # ("Typing..." -> "Saved to browser memory"). The first version reloaded
    # immediately, the debounce never fired, and the check reported that the
    # notes did not persist — a defect in the harness wearing the tool's
    # clothes. Polling the tool's own status text is better than sleeping a
    # fixed 1.2s, because it asserts the contract the tool advertises rather
    # than a number I picked.
    saved = False
    for _ in range(40):
        st = (d.text("#save-status") or "")
        if "saved" in st.lower():
            saved = True
            break
        time.sleep(0.1)
    record(tool, "ROUNDTRIP", "tool confirms the note was saved",
           "PASS" if saved else "FAIL",
           "#save-status reads %r" % (d.text("#save-status") or "")[:50])
    d.goto(url, clear_state=False)
    back_note = d.ev("(()=>{const e=document.querySelector('#user-notes');"
                     "return e?e.value:null})()")
    ticked = d.ev("(()=>%s.filter(s=>{const e=document.querySelector(s);"
                  "return e&&e.checked}).length)()" % json.dumps(boxes))
    record(tool, "ROUNDTRIP", "notes survive a reload",
           "PASS" if back_note == note else "FAIL",
           "wrote %r, read back %r" % (note[:28], (back_note or "")[:28]))
    record(tool, "ROUNDTRIP", "ticked boxes survive a reload",
           "PASS" if ticked == len(boxes) else "FAIL",
           "%s of %d still ticked" % (ticked, len(boxes)))


def check_portfolio(d):
    """Aggregation IS arithmetic and is checked as such; the rest is not."""
    tool = "mortgages/portfolio.html"
    url = BASE + tool
    print("\n=== %s   [class C tool with a class A core — the AGGREGATE is "
          "checkable]" % tool)

    props = [
        {"name": "Alpha", "value": 300000, "mortgage": 210000, "rent": 1400,
         "repayment": 900},
        {"name": "Beta", "value": 180000, "mortgage": 45000, "rent": 850,
         "repayment": 300},
        {"name": "Gamma", "value": 520000, "mortgage": 390000, "rent": 2100,
         "repayment": 1650},
    ]
    tot_val = sum(p["value"] for p in props)
    tot_mort = sum(p["mortgage"] for p in props)
    tot_rent = sum(p["rent"] for p in props)
    tot_pay = sum(p["repayment"] for p in props)

    # Seeded deterministically through the tool's OWN storage key, then reloaded
    # so the page rebuilds from it — the brief names this as the only way to get
    # a controlled baseline out of a localStorage tool.
    d.goto(url)
    d.seed_storage("uk_mortgage_portfolio_v1", props)
    d.goto(url, clear_state=False)

    def read(sel):
        return parse_money(d.text(sel))

    for sel, want, what in [
            ("#d_totalVal", tot_val, "total value = Σ values"),
            ("#d_equity", tot_val - tot_mort, "net equity = Σ(value − mortgage)"),
            ("#d_ltv", O.ltv_pct(tot_mort, tot_val), "portfolio LTV = Σmortgage/Σvalue"),
            ("#d_yield", O.gross_yield_pct(tot_rent, tot_val),
             "gross yield = 12·Σrent/Σvalue"),
            ("#d_rent", tot_rent, "monthly rent roll = Σ rents"),
            ("#d_cashflow", tot_pay and tot_rent - tot_pay, "cashflow = Σrent − Σpayment"),
            ("#d_count", len(props), "property count")]:
        try:
            got = read(sel)
            tol = 0.51 if abs(got - round(got)) < 1e-9 else 0.011
            record(tool, "AGGREGATE", what, "PASS" if abs(got - want) <= tol
                   else "FAIL", "shown %.2f, definition %.2f" % (got, want))
        except Exception as e:                              # noqa: BLE001
            record(tool, "AGGREGATE", what, "N/A", str(e))

    back = d.ev("(()=>localStorage.getItem('uk_mortgage_portfolio_v1'))()")
    same = False
    try:
        same = json.loads(back) == props
    except Exception:                                       # noqa: BLE001
        pass
    record(tool, "ROUNDTRIP", "seeded portfolio survives the render unchanged",
           "PASS" if same else "FAIL",
           "identical" if same else "storage now %r" % (str(back)[:70],))


CHECKS = {
    "damage": check_damage_checker,
    "credit": check_credit_health,
    "factfinder": check_fact_finder,
    "tracker": check_application_tracker,
    "portfolio": check_portfolio,
}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--json")
    ap.add_argument("--tools", help="comma-separated: %s" % ",".join(CHECKS))
    a = ap.parse_args()

    want = [t.strip() for t in a.tools.split(",")] if a.tools else list(CHECKS)
    d = Driver()
    try:
        for k in want:
            if k not in CHECKS:
                print("no such check: %s" % k)
                continue
            try:
                CHECKS[k](d)
            except (PageLoadError, DriveError) as e:
                record(k, "DRIVE", "could not drive the tool", "N/A", str(e))
    finally:
        d.close()

    counts = {}
    for r in RESULTS:
        counts[r["status"]] = counts.get(r["status"], 0) + 1
    print("\n%s" % ("-" * 72))
    print("INVARIANTS — PASS %d   FAIL %d   N/A %d   "
          "(these are NOT arithmetic checks; see the report)"
          % (counts.get("PASS", 0), counts.get("FAIL", 0), counts.get("N/A", 0)))
    if a.json:
        with open(a.json, "w") as f:
            json.dump({"counts": counts, "results": RESULTS}, f, indent=1)
        print("wrote %s" % a.json)
    return 1 if counts.get("FAIL") else 0


if __name__ == "__main__":
    sys.exit(main())

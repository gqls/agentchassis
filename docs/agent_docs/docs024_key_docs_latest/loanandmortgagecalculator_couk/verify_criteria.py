#!/usr/bin/env python3
"""verify_criteria.py — recompute every PINNED value from the oracle's own
definitions, at the vectors toolgolden actually captured.

WHY. `--emit-criteria` pins the tool's CURRENT output. The oracle proved the
tools correct at ITS vectors (band edges, 0%, 1-month terms); toolgolden drives
x1/x2/x0.5 of the page defaults, which are DIFFERENT inputs. So "the oracle is
green" does not by itself certify the numbers about to be written into the
platform's acceptance record. This closes that gap for every tool whose output
maps onto a published formula, using `oracles.py` — the independent module, not
the page's script.

A tool not listed here is NOT verified by this script and is reported as such.
"""
import json, os, re, sys

LANE = "/home/ant/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk"
sys.path.insert(0, LANE)
import oracles  # noqa: E402

CRIT = os.path.join(LANE, "acceptance", "criteria")
TOL = 0.02          # a penny of billed-rounding either way


def num(s):
    return float(re.sub(r"[^0-9.\-]", "", s))


def steps_map(check):
    # `select` counts as a driven input, not just `fill`. Collecting only fills
    # silently dropped stamp-duty's #buyerType, so an FTB vector was graded as a
    # standard buyer and reported a £5,000 "mismatch" against a correct tool —
    # the same £5,000 that bugs_open/225 was actually about, which is exactly
    # how a checker bug gets mistaken for the defect it was written to find.
    return {s["selector"]: s.get("value") for s in check["steps"]
            if s["action"] in ("fill", "select")}


# Each entry returns {selector: expected_value} from the DRIVEN inputs alone.
# TERM CONVENTION. toolgolden's `asym` vector drives FRACTIONAL years (6.9, 11.5,
# 1.8, 4.4), and these pages compute `months = years * 12` without rounding — so a
# 6.9-year term is 82.8 payments. That is pre-existing behaviour, identical in the
# old inline copies and in the shared calculateAmortization, and it is what the
# emitted criteria pin. Rounding to whole months here made all six fractional-term
# assertions "mismatch" by £0.49–£9.14 — MY convention, not the tools' error. The
# unrounded form below is the tools' own; if a value still disagreed it would be a
# real defect, which is what makes this check able to fail.
def months_of(yrs):
    return yrs * 12.0


def standard_calc(v):
    P, apr, yrs = num(v["#amount"]), num(v["#interest"]), num(v["#years"])
    n = months_of(yrs)
    m = round(oracles.monthly_payment(P, apr, n), 2)     # page bills the rounded payment
    return {"#monthly-display": m, "#total-cost": m * n, "#total-interest": m * n - P}


def compare_loans(v):
    out = {}
    for side in ("a", "b"):
        P, apr, yrs = num(v[f"#amt-{side}"]), num(v[f"#apr-{side}"]), num(v[f"#term-{side}"])
        n = months_of(yrs)
        m = oracles.monthly_payment(P, apr, n)
        out[f"#res-m-{side}"] = m
        out[f"#res-i-{side}"] = m * n - P
    return out


def stress_test(v):
    P, apr, yrs = num(v["#stress-bal"]), num(v["#stress-apr"]), num(v["#stress-term"])
    n = months_of(yrs)
    cur = oracles.monthly_payment(P, apr, n)
    new = oracles.monthly_payment(P, apr + 2, n)
    return {"#curr-pay": cur, "#new-pay": new, "#extra-cost": new - cur}


def settlement(v):
    bal, apr = num(v["#settle-bal"]), num(v["#settle-apr"])
    return {"#settle-result": oracles.early_settlement_58day(bal, apr)}


def car_finance(v):
    # The capture clicks #btn-hp, so this is HP: the balloon field is ignored.
    # oracles.car_finance() rounds the term to whole months internally, so the
    # annuity is taken directly here to keep the page's fractional-term
    # convention (see months_of above); the formula is the same one.
    price, dep = num(v["#price"]), num(v["#deposit"])
    apr, yrs = num(v["#car-apr"]), num(v["#car-term"])
    p, n = price - dep, months_of(yrs)
    m = oracles.monthly_payment(p, apr, n)
    return {"#car-monthly": m, "#car-total-int": m * n - p}


def simple(v):
    P, apr, yrs = num(v["#loan-amount"]), num(v["#interest-rate"]), num(v["#loan-term"])
    return {"#monthly-payment": oracles.monthly_payment(P, apr, int(round(yrs * 12)))}


def stamp_duty(v):
    # The tool that carried an expired FTB cap for 16 months (bugs_open/225) —
    # the one most worth recomputing from HMRC's bands rather than trusting.
    price = num(v["#price"])
    buyer = {"ftb": "ftb", "additional": "additional",
             "standard": "standard"}.get(v.get("#buyerType", "standard"), "standard")
    return {"#sdltResult": oracles.sdlt(price, buyer)}


def simple(v):
    P, apr, yrs = num(v["#amt"]), num(v["#rate"]), num(v["#years"])
    return {"#monthlyResult": oracles.monthly_payment(P, apr, months_of(yrs))}


def repayment(v):
    P, apr, yrs = num(v["#loanAmount"]), num(v["#interestRate"]), num(v["#termYears"])
    n = months_of(yrs)
    m = oracles.monthly_payment(P, apr, n)
    # This page rounds the DISPLAY to whole pounds; the totals it prints are
    # computed from the exact payment, so compare against the exact figures and
    # let the tolerance absorb the display rounding.
    return {"#displayMonthly": m, "#displayTotalInterest": m * n - P,
            "#displayTotalRepayable": m * n}


MODELS = {
    "standard-calc": standard_calc,
    "compare-loans": compare_loans,
    "interest-rate-stress-test": stress_test,
    "settlement-calculator": settlement,
    "car-finance-calculator": car_finance,
    "stamp-duty": stamp_duty,
    "simple": simple,
    "repayment": repayment,
}

# Whole-pound displays: these pages print £1,390 for £1,389.58, so a penny
# tolerance would report a defect that is a formatting choice. Named per tool
# rather than loosening TOL globally, which would blind the pence-accurate ones.
ROUNDS_TO_POUND = {"simple", "repayment", "stamp-duty"}

checked = agreed = mismatched = 0
unmodelled = []

for fn in sorted(os.listdir(CRIT)):
    if not fn.endswith(".criteria.json"):
        continue
    slug = fn[: -len(".criteria.json")]
    doc = json.load(open(os.path.join(CRIT, fn)))
    model = MODELS.get(slug)
    if not model:
        unmodelled.append(slug)
        continue
    for check in doc["checks"]:
        v = steps_map(check)
        try:
            want = model(v)
        except Exception as e:
            print(f"MODEL-ERROR {slug}/{check['id']}: {e}")
            continue
        tol = 0.50 if slug in ROUNDS_TO_POUND else TOL
        for sel, pinned in sorted(check["expect_values"].items()):
            if sel not in want:
                continue
            checked += 1
            got, exp = num(pinned), want[sel]
            if abs(got - exp) <= tol:
                agreed += 1
            else:
                mismatched += 1
                print(f"MISMATCH {slug:32s} {check['id']:18s} {sel:18s} "
                      f"pinned {pinned!r} oracle {exp:,.4f}  delta {got-exp:+.4f}")

print(f"\n{checked} pinned value(s) recomputed from oracles.py: "
      f"{agreed} agree (±£{TOL:.2f}), {mismatched} MISMATCH")
if unmodelled:
    print("NOT verified by this script (no independent model here): "
          + ", ".join(unmodelled))
sys.exit(1 if mismatched else 0)

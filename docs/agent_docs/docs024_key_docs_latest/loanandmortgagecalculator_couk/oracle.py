#!/usr/bin/env python3
"""oracle.py — does each calculator give the RIGHT answer, not merely a stable one?

THE GAP THIS CLOSES. `toolgolden.py` records what each tool computed on
2026-08-05 and `golden_compare_post.py` re-runs it; together they answer "does
it still do what it did?". Neither can answer "is that the right answer?" — if a
calculator has been wrong since it was written, the golden records the wrong
number faithfully and every later check certifies the bug. This file compares
against an INDEPENDENT ORACLE (`oracles.py`), computed from the annuity formula
and the published HMRC bands, which has never seen these pages' JavaScript.

HOW A CHECK CAN COME OUT, and why there are four outcomes rather than two:

  PASS         the tool's number matches the oracle within the tolerance its own
               display precision permits.
  CONVENTION   it matches a DIFFERENT, named, defensible convention (billed vs
               exact payment; serviced vs retained bridging interest). Reported,
               not convicted — but reported LOUDLY, because two tools on one
               site answering the same question differently is itself a finding.
  FAIL         it matches no stated convention. This is the finding the lane
               exists to produce.
  N/A          the check could not be made — a control that would not accept the
               vector, an output element that does not exist. Printed in the
               report rather than omitted from it, because a silently skipped
               check reads exactly like a passing one.

TOLERANCE IS DERIVED FROM WHAT THE TOOL DISPLAYS, not chosen. A page that
formats to whole pounds (`£1,390`) cannot be checked to the penny by anybody, so
the assertion is ±£0.50 and the report SAYS SO; a page that formats to pence is
held to ±£0.01. Choosing one global tolerance would either convict every
whole-pound tool or excuse every penny-level defect.

VECTORS ARE CHOSEN, NOT SCALED. Every case names its own numbers and most of
them sit on a boundary — band edges (£125k, £250k, £300k, £500k, £625k, £925k,
£1.5m), the £40,000 higher-rate floor, 0% APR, a one-month term, a zero balloon.
Scaling a page's own defaults by x2/x0.5 keeps every vector in the comfortable
middle of the domain, which is precisely where a boundary defect is not.

Usage:
  python3 oracle.py                          # all tools, live site
  python3 oracle.py --tools stamp-duty,simple
  python3 oracle.py --json out.json
  python3 oracle.py --mutate expectation     # CONTROL: must FAIL
  python3 oracle.py --mutate parse           # CONTROL: must FAIL
  python3 oracle.py --mutate crosstool       # CONTROL: must FAIL

Exit 1 if any check FAILs, or if a --mutate control did NOT fail.
"""
import argparse
import json
import re
import sys

import oracles as O
from oracle_driver import (BASE, Driver, DriveError, PageLoadError, ParseError,
                           parse_money)

MONTHS = 12


# ---------------------------------------------------------------------------
# Check construction
# ---------------------------------------------------------------------------


def chk(sel, want, what, alt=None, defect_alt=None, pct=False, raw_contains=None):
    """One assertion.

    `alt`        — DEFENSIBLE conventions, name -> the value that convention
                   gives. Matching one downgrades FAIL to CONVENTION.
    `defect_alt` — NAMED WRONG ANSWERS, name -> value. Matching one is still a
                   FAIL; the name just says precisely which wrong thing the tool
                   is doing.

    The two are separate because the first version had only `alt`, and the very
    first run used it to label `calcSDLT` implementing an EXPIRED tax rule as a
    "convention" — i.e. my own reporting mechanism downgraded the defect the
    lane exists to find, to an advisory, because the machinery for naming a
    cause was the same machinery that excused one. A superseded HMRC rule is not
    a matter of taste.
    """
    return {"sel": sel, "want": want, "what": what, "alt": alt or {},
            "defect_alt": defect_alt or {}, "pct": pct,
            "raw_contains": raw_contains}


def case(name, sets, checks, press=None, setup=None, prime=None):
    """One input vector.

    `prime` is a DIFFERENT vector driven first, whose outputs are recorded
    before the real vector is driven. It exists because of a failure mode that
    no value comparison can name on its own: a calculator whose guard rejects
    the vector (`if (rate > 0) { ... }`) writes nothing to the DOM at all, so
    the page goes on displaying the answer to the PREVIOUS inputs. To a user
    that is indistinguishable from an answer — there is no error, no blank, no
    zero, just a confident wrong number. Comparing against the primed reading
    makes "the tool did not recompute" a mechanically detectable state rather
    than something only a source-reader could assert.
    """
    return {"name": name, "set": sets, "checks": checks, "press": press,
            "setup": setup, "prime": prime}


def display_dp(text):
    """Decimal places the tool actually printed, so tolerance can follow it."""
    m = re.search(r"\d+\.(\d+)", str(text))
    return len(m.group(1)) if m else 0


def tolerance_for(text):
    """±half the last displayed digit, floored at the £0.01 the brief states."""
    dp = display_dp(text)
    return max(0.01, 0.5 * (10.0 ** -dp)) + 1e-9


# ---------------------------------------------------------------------------
# The tools. Selectors and captions come from `inventory.py` (labels, not
# script); expected values come from `oracles.py` (definitions, not script).
# ---------------------------------------------------------------------------


def build_tools():
    T = {}

    # -- loans/standard-calc ------------------------------------------------
    # Labels: "Amount to Borrow (£)", "Interest Rate (APR %)", "Term (Years)";
    # outputs captioned "Monthly Repayment", "Total Interest", "Total Repayable".
    # No button — it recalculates on input.
    def std(p, apr, y, label, prime=None):
        n = y * MONTHS
        m = O.monthly_payment(p, apr, n)
        return case(label,
                    {"#amount": p, "#interest": apr, "#years": y},
                    [chk("#monthly-display", m, "monthly payment"),
                     chk("#total-interest", O.total_interest(p, apr, n),
                         "total interest",
                         alt={"billed (payment rounded to the penny first)":
                              O.total_interest(p, apr, n, round_payment=True)}),
                     chk("#total-cost", p + O.total_interest(p, apr, n),
                         "total repayable",
                         alt={"billed (payment rounded to the penny first)":
                              p + O.total_interest(p, apr, n, round_payment=True)})],
                    prime=prime)

    T["loans/standard-calc.html"] = {
        "cls": "A", "oracle": "annuity M = P·r/(1−(1+r)^−n)",
        "cases": [
            std(10000, 7.9, 5, "site defaults"),
            std(10000, 0, 5, "BOUNDARY 0% APR (r→0 limit: M = P/n)",
                prime={"#amount": 20000, "#interest": 12.0, "#years": 10}),
            std(1000, 12.0, 1, "BOUNDARY 1-year term"),
            std(250000, 4.5, 25, "a mortgage-sized principal"),
        ],
        "determinism": [
            {"name": "0% APR reached from two different prior rates",
             "prime_a": {"#amount": 20000, "#interest": 12.0, "#years": 10},
             "prime_b": {"#amount": 3000, "#interest": 3.0, "#years": 2},
             "vector": {"#amount": 10000, "#interest": 0, "#years": 5},
             "sels": ["#monthly-display", "#total-interest", "#total-cost"]},
        ]}

    # -- mortgages/simple ---------------------------------------------------
    def simple(p, r, y, label):
        n = y * MONTHS
        return case(label, {"#amt": p, "#rate": r, "#years": y},
                    [chk("#monthlyResult", O.monthly_payment(p, r, n),
                         "monthly payment")],
                    press="button[onclick='doSimpleCalc()']")

    T["mortgages/simple.html"] = {
        "cls": "A", "oracle": "annuity",
        "cases": [simple(200000, 4.5, 25, "site defaults"),
                  simple(200000, 0, 25, "BOUNDARY 0% rate"),
                  simple(500000, 6.0, 40, "BOUNDARY 40-year term"),
                  simple(75000, 9.25, 10, "off-default vector")]}

    # -- mortgages/repayment ------------------------------------------------
    def rep(p, r, y, label):
        n = y * MONTHS
        m = O.monthly_payment(p, r, n)
        ti = O.total_interest(p, r, n)
        return case(label, {"#loanAmount": p, "#interestRate": r, "#termYears": y},
                    [chk("#displayMonthly", m, "monthly payment"),
                     chk("#displayTotalInterest", ti, "total interest",
                         alt={"billed (payment rounded first)":
                              O.total_interest(p, r, n, round_payment=True)}),
                     chk("#displayTotalRepayable", p + ti, "total repayable",
                         alt={"billed (payment rounded first)":
                              p + O.total_interest(p, r, n, round_payment=True)})],
                    press="button[onclick='runCalc()']")

    T["mortgages/repayment.html"] = {
        "cls": "A", "oracle": "annuity + total cost",
        "cases": [rep(250000, 4.5, 25, "site defaults"),
                  rep(250000, 0, 25, "BOUNDARY 0% rate"),
                  rep(100000, 3.0, 5, "short term"),
                  rep(1000000, 5.75, 30, "large principal")]}

    # -- loans/compare-loans ------------------------------------------------
    # Two independent annuities plus a verdict. Driven ASYMMETRICALLY: identical
    # options make the verdict unfalsifiable and the two interest figures agree
    # by construction, which would look like corroboration and be none.
    def cmpl(a, b, label):
        na, nb = a[2] * MONTHS, b[2] * MONTHS
        ia = O.total_interest(a[0], a[1], na)
        ib = O.total_interest(b[0], b[1], nb)
        return case(label,
                    {"#amt-a": a[0], "#apr-a": a[1], "#term-a": a[2],
                     "#amt-b": b[0], "#apr-b": b[1], "#term-b": b[2]},
                    [chk("#res-m-a", O.monthly_payment(a[0], a[1], na), "option A monthly"),
                     chk("#res-i-a", ia, "option A total interest",
                         alt={"billed": O.total_interest(a[0], a[1], na, True)}),
                     chk("#res-m-b", O.monthly_payment(b[0], b[1], nb), "option B monthly"),
                     chk("#res-i-b", ib, "option B total interest",
                         alt={"billed": O.total_interest(b[0], b[1], nb, True)}),
                     chk("#verdict", None,
                         "verdict names the cheaper option (oracle: %s, "
                         "£%.2f vs £%.2f)"
                         % ("A" if ia < ib else "B", ia, ib),
                         raw_contains="Option A" if ia < ib else "Option B")])

    T["loans/compare-loans.html"] = {
        "cls": "A", "oracle": "annuity ×2, total cost = M·n",
        "cases": [cmpl((10000, 7.9, 5), (10000, 9.9, 4), "site defaults"),
                  cmpl((10000, 9.9, 4), (10000, 7.9, 5), "swapped — verdict must follow"),
                  # A 0% option is unambiguously the cheaper one, so the verdict
                  # here is decidable without any convention argument. Driven in
                  # BOTH slots: if only slot A were driven, a verdict that always
                  # names B would look like a slot-A arithmetic fault rather than
                  # a comparison made on a value that is not a number.
                  cmpl((5000, 0, 3), (5000, 5.0, 3),
                       "BOUNDARY 0% APR on option A — A must win"),
                  cmpl((5000, 5.0, 3), (5000, 0, 3),
                       "BOUNDARY 0% APR on option B — B must win")]}

    # -- loans/interest-rate-stress-test ------------------------------------
    # Caption states the shock: "If Rates Rise by 2%".
    def stress(bal, apr, y, label, prime=None):
        n = y * MONTHS
        cur = O.monthly_payment(bal, apr, n)
        new = O.monthly_payment(bal, apr + 2.0, n)
        return case(label, {"#stress-bal": bal, "#stress-apr": apr, "#stress-term": y},
                    [chk("#curr-pay", cur, "current payment"),
                     chk("#new-pay", new, "payment at +2%"),
                     chk("#extra-cost", new - cur, "extra per month",
                         alt={"extra over the whole term": (new - cur) * n})],
                    prime=prime)

    T["loans/interest-rate-stress-test.html"] = {
        "cls": "A", "oracle": "annuity at r and r+2pp",
        "cases": [stress(10000, 8.5, 3, "site defaults"),
                  stress(10000, 0, 3, "BOUNDARY 0% base rate",
                         prime={"#stress-bal": 25000, "#stress-apr": 14.0,
                                "#stress-term": 7}),
                  stress(200000, 4.5, 25, "mortgage-sized")]}

    # -- mortgages/rate-forecaster ------------------------------------------
    # Three rates over one term. The payment in each window is the payment for
    # the FULL term at that rate — the standard remortgage sketch.
    def fore(amt, term, r1, r2, r3, label):
        n = term * MONTHS
        # Windows named by the page's own captions: "Years 1-2", "Years 3-5",
        # "Years 6+" — 24 months, 36 months, the rest.
        p1, p2, p3 = O.reprice_schedule(amt, n, [(r1, 24), (r2, 36), (r3, n - 60)])
        naive = [O.monthly_payment(amt, r, n) for r in (r1, r2, r3)]
        return case(label,
                    {"#fcAmount": amt, "#fcTerm": term,
                     "#rate1": r1, "#rate2": r2, "#rate3": r3},
                    [chk("#pay1", p1, "years 1-2 payment"),
                     chk("#pay2", p2, "years 3-5 payment",
                         defect_alt={"the naive model: the FULL original "
                                     "principal re-amortised over the FULL "
                                     "original term": naive[1]}),
                     chk("#pay3", p3, "years 6+ payment",
                         defect_alt={"the naive model: the FULL original "
                                     "principal re-amortised over the FULL "
                                     "original term": naive[2]})],
                    press="button[onclick='calcForecast()']")

    T["mortgages/rate-forecaster.html"] = {
        "cls": "A", "oracle": "each window's payment amortises the balance "
                              "REMAINING at its start over the term REMAINING",
        "cases": [fore(250000, 25, 4.5, 5.5, 3.5, "site defaults"),
                  fore(250000, 25, 4.5, 4.5, 4.5, "CONTROL all three rates equal "
                       "— the three payments must be equal"),
                  fore(120000, 15, 0, 6.0, 2.25, "BOUNDARY 0% initial rate")]}

    # -- loans/overpayment-calculator + mortgages/overpayment ---------------
    # 12 mortgages pages share /assets/js/calculators.js, so these two are ONE
    # unit of evidence, not two (brief §7). Both are driven anyway — agreement
    # between them corroborates nothing, but DISagreement would be a finding.
    def over_loans(bal, rate, y, extra, label):
        n = y * MONTHS
        saved, months_saved = O.overpayment_saving(bal, rate, n, extra)
        return case(label, {"#bal": bal, "#rate": rate, "#term": y, "#over": extra},
                    [chk("#save-display", saved, "interest saved"),
                     chk("#time-display", months_saved, "months saved",
                         alt={"years saved": months_saved / 12.0})])

    T["loans/overpayment-calculator.html"] = {
        "cls": "A", "oracle": "amortisation schedule with and without the extra",
        "shared_js": "/assets/js/calculators.js",
        "cases": [over_loans(15000, 6.5, 5, 50, "site defaults"),
                  over_loans(15000, 6.5, 5, 0, "BOUNDARY zero overpayment "
                             "— saving must be £0 and 0 months"),
                  # This page carries its OWN inline annuity as well as loading
                  # the shared helper, so the shared file's explicit `rate === 0`
                  # branch does not necessarily protect it. Drive the case.
                  over_loans(15000, 0, 5, 50, "BOUNDARY 0% interest rate — the "
                             "overpayment saves TIME but no interest"),
                  over_loans(200000, 4.5, 25, 250, "mortgage-sized")]}

    def over_mort(bal, rate, y, extra, label):
        n = y * MONTHS
        saved, months_saved = O.overpayment_saving(bal, rate, n, extra)
        return case(label,
                    {"#opBalance": bal, "#opRate": rate, "#opYears": y,
                     "#opAmount": extra},
                    [chk("#saveInterest", saved, "interest saved"),
                     chk("#saveTime", None, "time saved (prose)",
                         raw_contains="%d Year" % (months_saved // 12)),
                     chk("#dispYearsEarlier", months_saved // 12, "whole years earlier")],
                    press="button[onclick='runOverpayment()']")

    T["mortgages/overpayment.html"] = {
        "cls": "A", "oracle": "amortisation schedule with and without the extra",
        "shared_js": "/assets/js/calculators.js",
        "cases": [over_mort(200000, 4.5, 20, 100, "site defaults"),
                  over_mort(200000, 4.5, 20, 0, "BOUNDARY zero overpayment"),
                  over_mort(90000, 7.0, 10, 300, "large overpayment")]}

    # -- loans/loan-vs-savings ----------------------------------------------
    # Labels: loan APR %, savings AER %, spare cash £, higher-rate taxpayer
    # (select values "0" and "0.4" — i.e. the marginal tax rate on interest).
    # One year's comparison: interest avoided vs interest earned net of tax.
    def lvs(loan_r, save_r, cash, tax, label):
        return case(label,
                    {"#loan-rate": loan_r, "#save-rate": save_r,
                     "#spare-cash": cash, "#tax-bracket": str(tax)},
                    [chk("#loan-benefit", cash * loan_r / 100.0,
                         "interest avoided by repaying"),
                     chk("#save-benefit", cash * save_r / 100.0 * (1 - tax),
                         "interest earned, net of tax",
                         alt={"gross of tax": cash * save_r / 100.0})])

    T["loans/loan-vs-savings.html"] = {
        "cls": "A", "oracle": "one year of simple interest each way, net of the "
                              "marginal rate the control declares",
        "cases": [lvs(7.5, 5.0, 1000, 0, "site defaults, basic rate"),
                  lvs(7.5, 5.0, 1000, 0.4, "higher-rate taxpayer — savings side "
                      "must fall by 40%"),
                  lvs(0, 5.0, 1000, 0, "BOUNDARY 0% loan rate")]}

    # -- loans/settlement-calculator ----------------------------------------
    # The page's own breakdown names the rule: 'approx. £x in "58-day" interest'.
    # That is the Consumer Credit (Early Settlement) Regulations 2004 figure
    # (28 days + up to 30 more on an agreement over a year), NOT the Rule of 78.
    def settle(bal, apr, label, prime=None):
        # At 0% APR the correct settlement figure IS the balance, so "balance
        # only" cannot be listed as a defensible alternative there — it would
        # excuse the very case the vector exists to test.
        alt = {"30 days": O.early_settlement_58day(bal, apr, 30),
               "28 days": O.early_settlement_58day(bal, apr, 28)}
        if apr != 0:
            alt["balance only, no deferment interest"] = float(bal)
        return case(label, {"#settle-bal": bal, "#settle-apr": apr},
                    [chk("#settle-result", O.early_settlement_58day(bal, apr),
                         "settlement figure (balance + 58 days' interest)",
                         alt=alt)],
                    prime=prime)

    T["loans/settlement-calculator.html"] = {
        "cls": "A", "oracle": "CCA 2004 early-settlement deferment: balance + "
                              "58 days' simple interest",
        "cases": [settle(5000, 9.9, "site defaults"),
                  settle(5000, 0, "BOUNDARY 0% APR — settlement must equal the balance",
                         prime={"#settle-bal": 12000, "#settle-apr": 19.9}),
                  settle(20000, 24.9, "high APR")],
        "determinism": [
            {"name": "0% APR reached from two different prior rates",
             "prime_a": {"#settle-bal": 12000, "#settle-apr": 19.9},
             "prime_b": {"#settle-bal": 800, "#settle-apr": 3.0},
             "vector": {"#settle-bal": 5000, "#settle-apr": 0},
             "sels": ["#settle-result", "#settle-breakdown"]},
        ]}

    # -- mortgages/investor (yield AND LTV — two tools in one page) ----------
    def inv(price, rent, val, loan, label):
        return case(label,
                    {"#btlPrice": price, "#btlRent": rent,
                     "#ltvPrice": val, "#ltvLoan": loan},
                    [chk("#yieldResult", O.gross_yield_pct(rent, price),
                         "gross annual yield", pct=True),
                     chk("#ltvResult", O.ltv_pct(loan, val), "LTV", pct=True)],
                    press="button[onclick='calcYield()']")

    T["mortgages/investor.html"] = {
        "cls": "A", "oracle": "yield = 12·rent/price; LTV = loan/value",
        "second_press": "button[onclick='calcLTV()']",
        "cases": [inv(250000, 1200, 300000, 225000, "site defaults"),
                  # Asymmetric: a ratio is scale-invariant, so scaling both
                  # sides cannot move it — that is what defeated toolgolden here.
                  inv(400000, 1200, 300000, 300000, "asymmetric; LTV exactly 100%"),
                  inv(100000, 1000, 500000, 50000, "12% yield, 10% LTV")]}

    # -- mortgages/bridging-loan --------------------------------------------
    def bridge(net, rate, months, fee, label):
        g = O.bridging_gross(net, rate, months, fee)
        serviced = net / (1.0 - fee / 100.0)
        return case(label,
                    {"#brLoan": net, "#brRate": rate, "#brTerm": months, "#brFee": fee},
                    [chk("#resGross", g["gross"], "gross facility",
                         alt={"serviced (interest paid monthly, not retained)": serviced}),
                     chk("#resInterest", g["interest"], "retained interest",
                         alt={"interest on the NET advance":
                              net * rate / 100.0 * months}),
                     chk("#resFee", g["fee"], "arrangement fee",
                         alt={"fee on the NET advance": net * fee / 100.0}),
                     chk("#dispNet", net, "net advance echoed back")],
                    press="button[onclick='calcBridge()']")

    T["mortgages/bridging-loan.html"] = {
        "cls": "A", "oracle": "retained-interest gross-up: gross·(1−m·r−fee) = net",
        "cases": [bridge(200000, 0.75, 12, 2.0, "site defaults"),
                  bridge(200000, 0.75, 1, 0, "BOUNDARY 1 month, no fee"),
                  bridge(50000, 1.5, 18, 3.0, "expensive short-term")]}

    # -- mortgages/equity-release (roll-up; the LTV limit is a house table) --
    def equity(value, age, loan, rate, label):
        return case(label,
                    {"#erValue": value, "#erAge": age, "#erLoan": loan, "#erRate": rate},
                    [chk("#debt10", O.compound(loan, rate, 10), "debt after 10 years"),
                     chk("#debt20", O.compound(loan, rate, 20), "debt after 20 years"),
                     chk("#debt30", O.compound(loan, rate, 30), "debt after 30 years")],
                    press="button[onclick='calcCompound()']")

    T["mortgages/equity-release.html"] = {
        "cls": "A", "oracle": "compound roll-up P·(1+r)^n, annual compounding",
        "conventions_alt": True,
        "cases": [equity(400000, 65, 100000, 6.5, "site defaults"),
                  equity(400000, 65, 100000, 0, "BOUNDARY 0% — debt must not grow"),
                  equity(750000, 80, 250000, 8.25, "older borrower, higher rate")]}

    # -- loans/car-finance-calculator ---------------------------------------
    # Reclassified from the brief's class C to class A; see REPORT §3. It has a
    # closed form under BOTH conventions and the tool ships an HP/PCP switch, so
    # which convention it implements is answerable rather than a matter of taste.
    def car(price, dep, apr, years, balloon, mode, label, prime=None):
        c = O.car_finance(price, dep, apr, years, balloon if mode == "PCP" else 0.0)
        key = "pcp" if mode == "PCP" else "hp"
        return case(label,
                    {"#price": price, "#deposit": dep, "#car-apr": apr,
                     "#car-term": years, "#balloon": balloon},
                    [chk("#car-monthly", c[key + "_monthly"],
                         "%s monthly payment" % mode,
                         alt={"balloon subtracted UNdiscounted":
                              O.monthly_payment(price - dep - (balloon if mode == "PCP" else 0),
                                                apr, int(years * MONTHS)),
                              "balloon ignored entirely":
                              O.monthly_payment(price - dep, apr, int(years * MONTHS))}),
                     chk("#car-total-int", c[key + "_interest"],
                         "%s total interest" % mode)],
                    press="button[onclick=\"setType('%s')\"]" % mode,
                    prime=prime)

    T["loans/car-finance-calculator.html"] = {
        "cls": "A", "oracle": "HP = annuity on (price−deposit); PCP = annuity on "
                              "(price−deposit−discounted balloon)",
        "cases": [car(25000, 5000, 8.9, 4, 10000, "HP", "HP, site defaults"),
                  car(25000, 5000, 8.9, 4, 10000, "PCP", "PCP, site defaults"),
                  car(25000, 5000, 8.9, 4, 0, "PCP",
                      "BOUNDARY zero balloon — PCP must equal HP"),
                  car(30000, 0, 0, 3, 12000, "PCP", "BOUNDARY 0% APR, no deposit",
                      prime={"#price": 30000, "#deposit": 0, "#car-apr": 8.9,
                             "#car-term": 4, "#balloon": 12000})],
        "determinism": [
            {"name": "0% APR (a real product: manufacturer 0% finance) from "
                     "two different prior rates",
             "prime_a": {"#price": 40000, "#deposit": 8000, "#car-apr": 12.9,
                         "#car-term": 5, "#balloon": 15000},
             "prime_b": {"#price": 9000, "#deposit": 500, "#car-apr": 4.0,
                         "#car-term": 2, "#balloon": 2000},
             "vector": {"#price": 30000, "#deposit": 0, "#car-apr": 0,
                        "#car-term": 3, "#balloon": 12000},
             "press": "button[onclick=\"setType('PCP')\"]",
             "sels": ["#car-monthly", "#car-total-int"]},
        ]}

    # -- loans/consolidation -------------------------------------------------
    # Rows are built by "+ Add Another Debt" and carry classes, not ids.
    def consol(debts, new_rate, new_years, label):
        total_bal = sum(d[0] for d in debts)
        old_int = sum(O.total_interest(d[0], d[1], d[2]) for d in debts)
        n = new_years * MONTHS
        new_m = O.monthly_payment(total_bal, new_rate, n)
        new_int = O.total_interest(total_bal, new_rate, n)
        setup = [{"rows": len(debts), "debts": debts}]
        return case(label, {"#new-rate": new_rate, "#new-term": new_years},
                    [chk("#curr-total-bal", total_bal, "total debt to clear"),
                     chk("#old-int", old_int, "remaining interest on the old debts",
                         alt={"billed": sum(O.total_interest(d[0], d[1], d[2], True)
                                            for d in debts)}),
                     chk("#new-monthly", new_m, "new monthly payment"),
                     chk("#new-int", new_int, "new total interest",
                         alt={"billed": O.total_interest(total_bal, new_rate, n, True)})],
                    setup=setup)

    T["loans/consolidation.html"] = {
        "cls": "A", "oracle": "per-debt annuity interest summed, vs one "
                              "consolidated annuity",
        "cases": [consol([(5000, 21.9, 36)], 9.9, 5, "one debt"),
                  consol([(5000, 21.9, 36), (3000, 29.9, 24)], 9.9, 5, "two debts"),
                  consol([(2000, 0, 12)], 9.9, 5, "BOUNDARY 0% APR debt — "
                         "old interest must be £0"),
                  # The 0%-APR DEBT case passes because a guard returning zero
                  # happens to be the right answer for "interest on a 0% debt".
                  # A 0% NEW LOAN is where the same guard would be wrong: the
                  # monthly payment on an interest-free consolidation is
                  # balance/n, not zero and not blank. Checking only the case
                  # where a broken guard looks correct is how a defect survives
                  # a boundary suite.
                  consol([(5000, 21.9, 36)], 0, 5, "BOUNDARY 0% APR on the NEW "
                         "consolidation loan — payment must be balance/n")]}

    # -- mortgages/stamp-duty (class B — the answer is external and dated) ---
    def duty(price, buyer, label):
        want = O.sdlt(price, buyer)
        bad = {}
        if buyer == "ftb":
            sup = O.sdlt_superseded_ftb(price)
            if abs(sup - want) > 0.005:
                bad["the SUPERSEDED 2022-09-23..2025-03-31 FTB rule "
                    "(£625,000 cap), which expired 16 months ago"] = sup
        if buyer == "additional" and price < O.SURCHARGE_FLOOR:
            bad["the 5% higher rate on a purchase below the £40,000 floor, "
                "where the higher rates do not apply at all"] = \
                O._banded(price, O.SURCHARGE_ADDITIONAL)
        return case(label, {"#price": price, "#buyerType": buyer},
                    [chk("#sdltResult", want, "SDLT payable", defect_alt=bad)],
                    press="button[onclick='calcSDLT()']")

    T["mortgages/stamp-duty.html"] = {
        "cls": "B",
        "oracle": "HMRC SDLT bands in force from 1 April 2025 — gov.uk "
                  "residential-property-rates + higher-rates guidance + SDLTM29805",
        "cases": [
            duty(350000, "next", "site defaults, standard buyer"),
            duty(125000, "next", "BOUNDARY top of the nil band"),
            duty(125001, "next", "BOUNDARY £1 into the 2% band"),
            duty(250000, "next", "BOUNDARY top of the 2% band"),
            duty(925000, "next", "BOUNDARY top of the 5% band"),
            duty(1500000, "next", "BOUNDARY top of the 10% band"),
            duty(1500001, "next", "BOUNDARY £1 into the 12% band"),
            duty(300000, "ftb", "BOUNDARY FTB nil-band ceiling"),
            duty(300001, "ftb", "BOUNDARY £1 over the FTB nil band"),
            duty(500000, "ftb", "BOUNDARY FTB relief cap — relief still available"),
            duty(500001, "ftb", "BOUNDARY £1 over the cap — relief WITHDRAWN, "
                 "standard rates on the whole price"),
            duty(600000, "ftb", "inside the withdrawn-relief band"),
            duty(625000, "ftb", "BOUNDARY the SUPERSEDED cap"),
            duty(625001, "ftb", "BOUNDARY £1 over the superseded cap"),
            duty(39999, "additional", "BOUNDARY £1 under the £40,000 "
                 "higher-rate floor — surcharge must NOT apply"),
            duty(40000, "additional", "BOUNDARY exactly at the floor"),
            duty(350000, "additional", "additional property, mid-range"),
        ]}

    # -- mortgages/affordability (class B — the model is the page's, the
    #    4.5× LTI framing is the page's own caption and is checkable) --------
    def afford(i1, i2, exp, label):
        return case(label, {"#income1": i1, "#income2": i2, "#expenses": exp},
                    [chk("#resHigh", O.income_multiple(i1 + i2, exp, 4.5),
                         "upper estimate at 4.5× LTI"),
                     chk("#resLow", O.income_multiple(i1 + i2, exp, 4.0),
                         "lower estimate at 4.0× LTI",
                         alt={"4.25×": O.income_multiple(i1 + i2, exp, 4.25)})],
                    press="button[onclick='calcAffordability()']")

    T["mortgages/affordability.html"] = {
        "cls": "B", "oracle": "(income − 12·commitments) × multiple; the 4.5× "
                              "figure is the page's own caption and the FPC's LTI limit",
        "cases": [afford(40000, 0, 250, "site defaults"),
                  afford(40000, 30000, 0, "BOUNDARY joint income, no commitments"),
                  afford(30000, 0, 2500, "BOUNDARY commitments exceed the whole "
                         "income at the multiple — result must not go negative "
                         "without saying so")]}

    # -- mortgages/fee-analyser (class B — arithmetic closed-form, fee
    #    TREATMENT is a policy choice, so both treatments are named) ---------
    def fees(amount, rate, deal, fee, other, label):
        d = O.deal_period_cost(amount, rate, deal, fee, other)
        d_short = O.deal_period_cost(amount, rate, deal, fee, other,
                                     term_years=deal)
        return case(label,
                    {"#tcAmount": amount, "#tcRate": rate, "#tcDeal": deal,
                     "#tcFee": fee, "#tcOther": other},
                    [chk("#tcTotal", d["total"], "total cost over the deal period",
                         alt={"payments only, fees excluded": d["payments"],
                              "term treated as the DEAL length, not the mortgage "
                              "term": d_short["total"]})],
                    press="button[onclick='calcTrueCost()']")

    T["mortgages/fee-analyser.html"] = {
        "cls": "B", "oracle": "payments over the deal window at the deal rate "
                              "(amortised over the full term) + fees",
        "cases": [fees(200000, 4.19, 2, 999, 0, "site defaults"),
                  fees(200000, 4.19, 2, 0, 0, "BOUNDARY no fees — must equal "
                       "payments alone"),
                  fees(150000, 3.5, 5, 1495, 500, "five-year fix with broker fee")]}

    return T


# ---------------------------------------------------------------------------
# Execution
# ---------------------------------------------------------------------------


def run_determinism(d, url, spec):
    """Is the tool's output a FUNCTION OF ITS INPUTS, or of how you got there?

    Drive the SAME final vector twice, reached by two different routes, and
    compare the readings. This needs no oracle, no formula and no sight of the
    page's source — and it states the defect at exactly the altitude that
    matters to a user: type these numbers and the answer depends on what you
    typed before them.

    It exists because the `prime` diagnosis was not enough. A guarded
    calculator (`if (rate > 0) { ...write the DOM... }`) leaves the previous
    answer on screen, but "the previous answer" is whatever the LAST ACCEPTED
    vector produced — including an intermediate state created halfway through
    typing the new one. Comparing against a single primed reading therefore
    misses it: on `standard-calc` the stale figure was £143.47, the answer to
    (£10,000, 12%, 10y), a combination the user never deliberately entered and
    the harness never recorded. Two routes to one destination has no such blind
    spot, because it compares two readings that MUST agree whatever the tool
    computes.
    """
    out = []
    readings = []
    for route in ("A", "B"):
        d.goto(url)
        for sel, val in spec["prime_" + route.lower()].items():
            d.set(sel, val)
        if spec.get("press"):
            d.click(spec["press"])
        for sel, val in spec["vector"].items():
            d.set(sel, val)
        if spec.get("press"):
            d.click(spec["press"])
        readings.append({s: d.text(s) for s in spec["sels"]})

    for s in spec["sels"]:
        a, b = readings[0][s], readings[1][s]
        rec = {"vector": spec["name"], "what": "same inputs by two routes: %s" % s,
               "sel": s, "shown": "%r vs %r" % (a, b)}
        if a == b:
            rec["status"] = "PASS"
        else:
            rec.update(status="FAIL",
                       note="the SAME final inputs give %r by one route and %r "
                            "by another — the output is not a function of the "
                            "inputs on screen" % (a, b))
        out.append(rec)
    return out


def apply_setup(d, setup):
    """Tool-specific state that must exist before the vector can be driven."""
    for s in setup:
        if "rows" in s:
            # consolidation: press "+ Add Another Debt" until there are as many
            # rows as the case needs, then fill by class (the row inputs carry
            # no ids — `addDebtRow` builds them with .d-bal/.d-rate/.d-months).
            for _ in range(s["rows"] - 1):
                d.click("button[onclick='addDebtRow()']")
            for i, (bal, apr, months) in enumerate(s["debts"], start=1):
                base = "#debt-list .debt-row:nth-of-type(%d) " % i
                d.set(base + ".d-bal", bal)
                d.set(base + ".d-rate", apr)
                d.set(base + ".d-months", months)


def run_case(d, url, c, tool, mutate=None):
    """Drive one vector; return a list of result records."""
    out = []
    d.goto(url)
    if c.get("setup"):
        apply_setup(d, c["setup"])

    primed = {}
    if c.get("prime"):
        for sel, val in c["prime"].items():
            d.set(sel, val)
        if c.get("press"):
            d.click(c["press"])
        for a in c["checks"]:
            primed[a["sel"]] = d.text(a["sel"])

    for sel, val in c["set"].items():
        if mutate == "parse":
            # CONTROL: hand the field a formatted string instead of a number.
            # A tool that silently reads it as 0 (or a harness that silently
            # parses '£0.00' back as a pass) must be caught here.
            val = "£%s" % format(val, ",") if isinstance(val, (int, float)) else val
        d.set(sel, val)
    if c.get("press"):
        d.click(c["press"])
    if tool.get("second_press"):
        d.click(tool["second_press"])

    for a in c["checks"]:
        rec = {"vector": c["name"], "what": a["what"], "sel": a["sel"]}
        try:
            raw = d.text(a["sel"])
            if raw is None:
                rec.update(status="N/A", note="no element %s on the page" % a["sel"])
                out.append(rec)
                continue
            rec["shown"] = raw

            if a["raw_contains"] is not None:
                ok = a["raw_contains"] in raw
                rec.update(status="PASS" if ok else "FAIL",
                           want="text containing %r" % a["raw_contains"])
                out.append(rec)
                continue

            got = parse_money(raw)
            want = a["want"]
            if mutate == "expectation":
                want = 100.0        # CONTROL: assert a number nothing computes
            tol = tolerance_for(raw)
            rec.update(got=got, want=want, tol=tol,
                       resolution="±£%.2f (tool displays %d dp)"
                                  % (tol, display_dp(raw)))
            tw = a.get("_true_want")
            if (mutate == "crosstool" and isinstance(tw, float)
                    and isinstance(want, float) and abs(tw - want) <= tol):
                rec.update(status="N/A",
                           note="NON-TEST: the borrowed expectation (%.2f) "
                                "equals this vector's own (%.2f) within ±%.2f, "
                                "so no comparator could fail it — excluded from "
                                "the control rather than counted as a miss"
                                % (want, tw, tol))
                out.append(rec)
                continue
            if abs(got - want) <= tol:
                rec["status"] = "PASS"
                if mutate == "crosstool" and isinstance(tw, float):
                    # The comparator agreed with a BORROWED expectation. That is
                    # only innocent if the tool's own answer here is wrong and
                    # happens to equal the neighbour's right one — which is
                    # exactly the £39,999 SDLT case: the page charges the 5%
                    # surcharge below the £40,000 floor and prints £2,000, and
                    # £2,000 is what the £40,000 vector correctly expects. The
                    # mutation did not bite; the comparator was handed a true
                    # statement. Anything else is comparator blindness and must
                    # stay a red control.
                    if abs(got - tw) > tol:
                        rec.update(status="N/A",
                                   note="MUTATION DID NOT BITE: the tool's own "
                                        "answer here (%.2f) is WRONG — its true "
                                        "expectation is %.2f — and coincides "
                                        "with the borrowed one (%.2f). Verify "
                                        "this vector FAILs in the unmutated run."
                                        % (got, tw, want))
            else:
                rec.update(status="FAIL", delta=got - want)
                if primed.get(a["sel"]) is not None and primed[a["sel"]] == raw:
                    rec["diagnosis"] = ("a STALE answer — the reading is "
                                        "byte-identical to the one produced by "
                                        "the previous, different vector (%r), so "
                                        "the tool did not recompute and wrote "
                                        "nothing" % primed[a["sel"]][:32])
                # A named wrong answer is still wrong — check these FIRST so a
                # defect can never be excused by an overlapping convention.
                for name, v in (a["defect_alt"] or {}).items():
                    if v is not None and abs(got - v) <= tol:
                        # Both diagnoses can be true at once (a stale reading
                        # that also happens to equal a named wrong rule); keep
                        # both rather than letting whichever ran last win.
                        rec["diagnosis"] = (rec["diagnosis"] + "; also equals "
                                            if "diagnosis" in rec else "") + name
                        break
                if "diagnosis" not in rec and mutate != "expectation":
                    for name, v in (a["alt"] or {}).items():
                        if v is not None and abs(got - v) <= tol:
                            rec.update(status="CONVENTION", convention=name,
                                       convention_value=v)
                            break
        except (ParseError, DriveError) as e:
            rec.update(status="FAIL", note=str(e))
        except Exception as e:                      # noqa: BLE001
            rec.update(status="N/A", note="harness error: %s" % e)
        out.append(rec)
    return out


ICON = {"PASS": "  ok  ", "FAIL": " FAIL ", "CONVENTION": " CONV ", "N/A": " n/a  "}


def selftest_parse():
    """CONTROL, output side: `parse_money` must REFUSE, never quietly return 0.

    `--mutate parse` exercises the INPUT half — an unparseable value typed into
    a control is rejected by the element and refused by the driver. This is the
    OUTPUT half, and it is the half that matters most: every comparison in this
    file runs through `parse_money`, so a version that answered 0.0 for an empty
    or garbage reading would turn "the output element was never written" into
    "the tool answered zero" — a number the oracle can then be confidently
    wrong about. The £NaN readings this suite found on four live pages are the
    reason it is not hypothetical: had the parser zeroed them, three of them
    would have compared equal to a correct £0.00 expectation.
    """
    must_raise = ["", "   ", "£NaN", "NaN", "Infinity", "£", "-", "n/a",
                  "Calculating...", "£1,234.56 and £99.00", "1,2,3.4.5"]
    must_parse = {"£202.29": 202.29, "£1,390": 1390.0, "1,000": 1000.0,
                  "-£50.25": None, "9.35": 9.35, "0.0%": 0.0,
                  "£12,137.40": 12137.40, "£448.024": 448.024,
                  "£0 / mo Rent": 0.0, "-12.5": -12.5}
    bad = 0
    print("CONTROL — parse_money must refuse, not return 0:")
    for s in must_raise:
        try:
            v = parse_money(s)
            print("  BROKEN  %-22r parsed as %r (should have raised)" % (s, v))
            bad += 1
        except ParseError as e:
            print("  ok      %-22r refused: %s" % (s, str(e)[:44]))
    print("\nCONTROL — and must still read the real formats:")
    for s, want in must_parse.items():
        try:
            v = parse_money(s)
            if want is None:
                print("  note    %-22r parsed as %r" % (s, v))
            elif abs(v - want) < 1e-9:
                print("  ok      %-22r -> %s" % (s, v))
            else:
                print("  BROKEN  %-22r -> %r, wanted %r" % (s, v, want))
                bad += 1
        except ParseError as e:
            if want is None:
                print("  ok      %-22r refused: %s" % (s, str(e)[:44]))
            else:
                print("  BROKEN  %-22r refused but should parse to %r" % (s, want))
                bad += 1
    print("\n%s" % ("PARSE CONTROL OK" if not bad
                    else "PARSE CONTROL FAILED: %d cases" % bad))
    return 1 if bad else 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tools", help="comma-separated substrings of page paths")
    ap.add_argument("--json", help="write the full result record here")
    ap.add_argument("--mutate",
                    choices=["expectation", "parse", "crosstool", "wrongpage"],
                    help="run a CONTROL that must not come out green")
    ap.add_argument("--base", default=BASE)
    ap.add_argument("--selftest-parse", action="store_true",
                    help="CONTROL: prove parse_money refuses rather than "
                         "silently returning 0. No browser needed.")
    a = ap.parse_args()

    if a.selftest_parse:
        return selftest_parse()

    tools = build_tools()
    if a.tools:
        want = [t.strip() for t in a.tools.split(",") if t.strip()]
        tools = {k: v for k, v in tools.items() if any(w in k for w in want)}
        if not tools:
            print("no tool matched %r" % a.tools)
            return 2

    # CONTROL `wrongpage`: drive one tool's vectors against a DIFFERENT tool's
    # page. Useful but WEAK on its own, and worth saying why: these pages do not
    # share control ids, so it is refused at the first `set()` and never reaches
    # a comparison at all. It proves the driver notices a missing selector; it
    # proves nothing about whether the comparator can detect a wrong ANSWER.
    crosstool = None
    if a.mutate == "wrongpage":
        crosstool = "mortgages/simple.html"
        tools = {k: v for k, v in tools.items() if k == "loans/standard-calc.html"} \
            or {"loans/standard-calc.html": build_tools()["loans/standard-calc.html"]}

    # CONTROL `crosstool`: the strong version. Drive each vector correctly, on
    # the right page, with every selector present and every value landing — then
    # compare the readings against the expectations belonging to the NEXT vector
    # of the same tool. The page is real and working; only the expectation is
    # mismatched. Every comparison must therefore fail, and a green run means
    # the comparator is not looking at the numbers.
    if a.mutate == "crosstool":
        for tool in tools.values():
            cs = tool["cases"]
            rotated = []
            for i, c in enumerate(cs):
                donor = cs[(i + 1) % len(cs)]["checks"]
                # Carry each check's TRUE expectation alongside the borrowed
                # one. Some rotations are not tests at all: this suite deliberately
                # packs vectors onto adjacent boundaries, and £1,500,000 and
                # £1,500,001 of SDLT differ by 12p — inside any tolerance. A
                # borrowed expectation that equals the true one cannot fail, and
                # counting those as "the comparator missed it" would push me to
                # loosen the control until it went green, which is the trap.
                # So they are marked NON-TESTS and excluded by name.
                swapped = []
                for j, chk_new in enumerate(donor):
                    true_want = (c["checks"][j]["want"]
                                 if j < len(c["checks"]) else None)
                    swapped.append(dict(chk_new, _true_want=true_want))
                rotated.append(dict(c, checks=swapped))
            tool["cases"] = rotated
            tool["determinism"] = []      # a determinism check has no expectation
                                          # to mis-pair, so rotating is a no-op
                                          # for it — exclude rather than let it
                                          # pass and dilute the control.

    d = Driver()
    results = {}
    counts = {"PASS": 0, "FAIL": 0, "CONVENTION": 0, "N/A": 0}
    try:
        for page, tool in sorted(tools.items()):
            url = a.base + (crosstool or page)
            print("\n=== %s   [class %s]" % (page, tool["cls"]))
            print("    oracle: %s" % tool["oracle"])
            if crosstool:
                print("    CONTROL: driving these vectors against %s" % crosstool)
            recs = []
            for c in tool["cases"]:
                try:
                    recs += run_case(d, url, c, tool, mutate=a.mutate)
                except (PageLoadError, DriveError) as e:
                    recs.append({"vector": c["name"], "what": "(drive)",
                                 "status": "N/A", "note": str(e)})
            for spec in tool.get("determinism", []):
                try:
                    recs += run_determinism(d, url, spec)
                except (PageLoadError, DriveError) as e:
                    recs.append({"vector": spec["name"], "what": "(determinism)",
                                 "status": "N/A", "note": str(e)})
            results[page] = recs
            for r in recs:
                counts[r["status"]] = counts.get(r["status"], 0) + 1
                line = "  [%s] %-58s %s" % (ICON[r["status"]], r["vector"][:58],
                                            r["what"])
                print(line)
                if r["status"] in ("FAIL", "CONVENTION"):
                    if "got" in r:
                        print("           shown %-14s  oracle %-14s  delta %+.4f  %s"
                              % (r.get("shown", "")[:14],
                                 "%.4f" % r["want"] if isinstance(r.get("want"), float)
                                 else r.get("want"),
                                 r.get("delta", 0.0), r.get("resolution", "")))
                    if r.get("diagnosis"):
                        print("           DIAGNOSIS: the tool is computing %s"
                              % r["diagnosis"])
                    if r.get("convention"):
                        print("           matches convention: %s (= %.4f)"
                              % (r["convention"], r["convention_value"]))
                    if r.get("note"):
                        print("           %s" % r["note"])
                elif r["status"] == "N/A" and r.get("note"):
                    print("           %s" % r["note"])
    finally:
        d.close()

    print("\n%s" % ("-" * 72))
    print("PASS %d   FAIL %d   CONVENTION %d   N/A %d"
          % (counts["PASS"], counts["FAIL"], counts["CONVENTION"], counts["N/A"]))

    if a.json:
        with open(a.json, "w") as f:
            json.dump({"counts": counts, "mutate": a.mutate,
                       "results": results}, f, indent=1)
        print("wrote %s" % a.json)

    if a.mutate:
        # WHAT A CONTROL MUST PROVE, stated carefully — the first version got
        # this wrong and would have reported a working control as inert.
        #
        # The property under test is "a corrupted run cannot come out green".
        # FAIL satisfies it. So does a LOUD REFUSAL: `--mutate parse` types
        # '£250,000' into an <input type=number>, the element rejects it and
        # holds '', and `set()` raises rather than driving on with an empty
        # field — which is the very "silent 0" the control exists to forbid.
        # Requiring specifically a FAIL would have marked that correct refusal
        # as a broken checker. What must NEVER happen is a PASS.
        if counts["PASS"] > 0:
            print("\nCONTROL DID NOT FAIL — %d checks PASSED under a mutation "
                  "that makes every answer wrong. The checker is inert. This is "
                  "the failure the control exists to detect." % counts["PASS"])
            return 1
        print("\nCONTROL OK: no check passed under mutation "
              "(%d FAIL, %d refused before comparison)."
              % (counts["FAIL"], counts["N/A"]))
        return 0

    return 1 if counts["FAIL"] else 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""add_tool_doc_headers.py — give each rewritten component its tool-doc header.

WHY. 019 §Tool Doc Header: every tool's `html_template` must open its `<script>`
with one sentinel-delimited block carrying purpose and behavioural invariants.
`check_tool_health` tests for the exact sentinels and raises an `improve_tool`
work item per tool without one, so loading eleven components bare would have
manufactured eleven warnings and pointed `tool-improver` at eleven tools that
are, in fact, the most thoroughly verified on the fleet.

OUTPUT-NEUTRAL BY CONTRACT, not by hope: the block is STRIPPED at deploy
assembly (`StripToolDocHeader`), so it never reaches a public page, and it is a
JS comment in any case — it cannot enter textContent and cannot move a number.
Proven rather than argued: `verify_rewrite.py` is re-run over all eleven after
this, against the same golden.

THE `invariants:` LINE IS THE POINT. The rest of the header restates what the
code says. That line says what a future editor must not break and what will
catch them if they do — which is the one thing the code cannot say about itself.

Idempotent: a template that already has the sentinels is left alone.
"""
import io
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
OPEN, CLOSE = "/* === tool-doc ===", "=== /tool-doc === */"

# name · page url · description · behaviour lines · invariants
DOCS = {
    "tool-loan-repayment": (
        "Standard Loan Repayment Calculator", "/tools/standard-calc.html",
        "Monthly repayment, total interest and total repayable for a fixed-rate personal loan, "
        "using the standard annuity formula with monthly compounding.",
        ["User sets amount, APR and term in years; every field recalculates on input.",
         "Monthly payment is the annuity formula on a monthly rate of (APR/100)/12.",
         "Total repayable is monthly x months; total interest is that less the principal.",
         "Runs once on load, so the tool is never blank."],
        "£10,000 at 7.9% over 5 years MUST read £202.29 / £2,137.40 / £12,137.40. A single "
        "divisor error (/12 -> /11) is caught in all three vectors by the computed_values "
        "fence. This component also serves /index.html — both pages share one calculator.",
    ),
    "tool-credit-health-check": (
        "Credit Health Check", "/tools/credit-health-check.html",
        "A five-step self-assessment wizard scoring indicative creditworthiness from four "
        "questions, showing a banded verdict with a meter.",
        ["Each answer adds its data-chc-points to a running total and advances one step.",
         "Step 5 clamps the meter to 10..95% and picks a band from two thresholds.",
         "One delegated listener on the container; no globals, no inline onclick."],
        "The two CSS rules .chc-step{display:none} / .chc-step.chc-active{display:block} are "
        "the whole of the one-step-at-a-time behaviour — without them every step renders at "
        "once and the tool scores correctly while looking broken. display is fingerprinted: "
        "step-1 block and steps 2-5 none on first paint.",
    ),
    "tool-rate-stress-test": (
        "Variable Rate Stress Test", "/tools/interest-rate-stress-test.html",
        "Shows what a variable-rate loan repayment becomes if the rate rises by a configurable "
        "number of percentage points, and the extra monthly cost.",
        ["Annuity formula applied twice: at the entered APR, and at APR + stress_delta_points.",
         "The difference is reported per month.",
         "stress_delta_points is percentage POINTS, not a multiplier."],
        "£10,000 at 8.5% over 3 years MUST read £315.68 / £325.02 / 9.35. Formatting is "
        "toFixed(2), NOT Intl.NumberFormat — switching would add thousands separators and "
        "change every displayed figure. stressed_heading must agree with stress_delta_points; "
        "nothing enforces it.",
    ),
    "tool-early-settlement": (
        "Early Settlement Estimator", "/tools/settlement-calculator.html",
        "Estimates a UK early-settlement figure: outstanding balance plus simple daily interest "
        "over a statutory notice window (the 58-day convention).",
        ["Daily rate is APR/days_in_year; extra interest is balance x daily x settlement_days.",
         "Settlement figure is balance + that interest.",
         "The breakdown sentence lives in the MARKUP; the script writes only the number."],
        "£5,000 at 9.9% MUST read £5,078.66 with £78.66 of charges. NEVER interpolate the "
        "breakdown copy into a JS string literal — the fallback contains the phrase 58-day in "
        "double quotes and doing so produced a syntax error that killed the whole tool while it "
        "still passed every structural check. 58 days is a DEFAULT, not a universal rule.",
    ),
    "tool-overpayment-impact": (
        "Overpayment Impact Tool", "/tools/overpayment-calculator.html",
        "Shows the total interest saved and months cut from a loan by paying a fixed extra "
        "amount each month, by amortising month by month.",
        ["Scheduled payment from the annuity formula; then a month-by-month loop at payment + overpayment.",
         "Saving is original total interest less the amortised total interest.",
         "Both outputs are floored at zero."],
        "£15,000 at 6.5% over 5 years with £50/month MUST read £448.024 saved and 10 months "
        "earlier. THE max_months LOOP CAP IS LOAD-BEARING — without it a negative overpayment "
        "never terminates and the tab hangs. KNOWN DEFECT, preserved deliberately: "
        "toLocaleString omits maximumFractionDigits so it prints THREE decimal places on money. "
        "Fix it as its own change with a re-baselined golden.",
    ),
    "tool-loan-vs-savings": (
        "Pay Off Loan or Save?", "/tools/loan-vs-savings.html",
        "Compares one year of interest avoided by clearing debt against one year of net interest "
        "earned by saving the same sum, adjusted for the saver's tax band.",
        ["Loan side: amount x rate. Savings side: amount x rate x (1 - tax fraction).",
         "The larger side gets the .winner class.",
         "higher_rate_fraction is a FRACTION (0.4), not a percentage."],
        "£1,000 at 7.5% against 5.0% MUST read £75.00 versus £50.00. On an exact tie the "
        "SAVINGS panel wins (the comparison is strict) — recorded behaviour, not a decision. "
        "KNOWN DEFECT: the verdict is signalled by COLOUR ALONE, invisible to a colour-blind "
        "reader; a badge was written and reverted because it changes the panels' text.",
    ),
    "tool-return-damage-checker": (
        "Car Return Damage Checker", "/tools/damage-checker.html",
        "A pre-return inspection checklist for a PCP or HP car handover; ticking any item "
        "reveals a warning about likely reconditioning charges.",
        ["Four checkboxes; the verdict box is shown when any is ticked and hidden when none is.",
         "The checkbox query is SCOPED TO THIS TOOL, not the page."],
        "NO ARITHMETIC AT ALL, so computed_values does not apply and the capture emitter "
        "correctly refuses it. Its contract is a visibility change: assert it with `interaction` "
        "plus `has_visible_area`, and do NOT manufacture a number so a check type fits. The row "
        "must NOT be display:flex — that blockifies the checkboxes and computed display is "
        "fingerprinted.",
    ),
    "tool-compare-loan-offers": (
        "Compare Loan Offers", "/tools/compare-loans.html",
        "Side-by-side comparison of two loan offers, ranked on TOTAL INTEREST over the life of "
        "each loan rather than on monthly payment.",
        ["Annuity formula per column; the lower total interest wins and gets .winner.",
         "The verdict names the winner and the amount saved.",
         "Every runtime-chosen string is read from a hidden markup copy bank."],
        "£10,000 at 7.9%/5y against £10,000 at 9.9%/4y MUST read £202.29 / £2137.14 against "
        "£253.15 / £2151.00, A cheaper by £13.86. RANKING ON TOTAL INTEREST IS THE WHOLE POINT — "
        "the cheaper option here has the HIGHER monthly payment, and a tool ranking by monthly "
        "payment would tell people the opposite of the truth.",
    ),
    "tool-car-finance-pcp-hp": (
        "Car Finance: PCP vs HP", "/tools/car-finance-calculator.html",
        "Monthly payment and total interest for a car finance agreement, in Hire Purchase or "
        "Personal Contract Purchase mode with an optional final balloon payment.",
        ["One annuity formula in balloon form serves both modes; HP is simply balloon = 0.",
         "The balloon field is hidden in HP and revealed in PCP.",
         "Total interest is measured against the CAR PRICE, including the deposit."],
        "£25,000 less £5,000 deposit at 8.9% over 4 years in HP MUST read £496.75 and £3,844.08. "
        "The .cf-balloon display:none / .cf-show pair is BEHAVIOUR — without it a field the "
        "arithmetic ignores is visible in HP mode. #balloon-box display is fingerprinted: none "
        "on first paint. KNOWN DEFECT: 0% APR computes nothing (the formula divides by zero) and "
        "0% is a real advertised product.",
    ),
    "tool-consolidation-risk": (
        "Debt Consolidation Risk Checker", "/tools/consolidation.html",
        "Compares the total interest still owed across several existing debts against the total "
        "interest of one consolidation loan, warning when consolidating costs more.",
        ["Each debt row is amortised separately; the interest still to pay is summed.",
         "The consolidated loan is amortised once over the new term.",
         "Rows can be added and removed; every control has a stable id."],
        "COMPARES ON TOTAL INTEREST, never on monthly payment — consolidation almost always "
        "lowers the monthly payment, so a monthly comparison would recommend it unconditionally, "
        "which is the mis-selling risk the page exists to warn about. Existing debts are in "
        "MONTHS, the new loan in YEARS — do not harmonise. KNOWN DEFECT: a debt with no rate "
        "counts toward the balance but contributes no interest, flattering consolidation.",
    ),
    "tool-application-tracker": (
        "Loan Application Tracker", "/tools/application-tracker.html",
        "A browser-local checklist of the documents a UK lender asks for, with a progress bar, "
        "free-text notes, and JSON backup and restore.",
        ["Checkbox state and notes persist to localStorage under namespaced keys.",
         "Notes save on a debounce; the status line distinguishes pending from saved.",
         "Clear removes only THIS tool's keys, never the whole origin."],
        "The row MUST be display:flex — this page's original stylesheet made it so and computed "
        "display is fingerprinted (the sibling damage-checker needs the opposite; do not "
        "harmonise them). Changing storage_check_prefix or storage_notes_key ORPHANS every "
        "existing user's saved data silently. 'Securely tracked in your browser' means "
        "unencrypted localStorage — do not strengthen that copy.",
    ),
}


def header_for(fn):
    name, page, desc, behaviour, invariants = DOCS[fn]
    lines = [OPEN,
             "  name: %s" % name,
             "  function: %s" % fn,
             "  page: %s (%s)" % (fn, page),
             "  description: %s" % desc,
             "  behaviour: |"]
    lines += ["    - %s" % b for b in behaviour]
    lines += ["  invariants: |",
              "    %s" % invariants,
              "  verified: computed_values fence emitted from a captured golden; the rewrite is",
              "    proven identical to the hand-built original across three input vectors",
              "    (see loancalculator_couk/rewrite/verify_rewrite.py).",
              "  " + CLOSE]
    return "\n".join(lines)


def main():
    changed = skipped = 0
    for fn in sorted(DOCS):
        path = os.path.join(HERE, fn + ".html.tmpl")
        if not os.path.exists(path):
            print("MISSING  %s" % os.path.basename(path))
            return 1
        src = io.open(path, encoding="utf-8").read()
        if OPEN in src:
            skipped += 1
            print("has one  %s" % fn)
            continue
        # Insert immediately after the FIRST <script> tag — the contract says the
        # header opens the script block, and StripToolDocHeader scans for the
        # sentinels wherever they are, so position matters for humans not code.
        m = re.search(r"<script>\s*\n", src)
        if not m:
            print("NO <script> in %s" % fn)
            return 1
        src = src[:m.end()] + header_for(fn) + "\n" + src[m.end():]
        io.open(path, "w", encoding="utf-8").write(src)
        changed += 1
        print("header   %s" % fn)
    print("\n%d header(s) added, %d already present" % (changed, skipped))
    print("NOW RE-RUN verify_rewrite.py — the header must not have moved a single value.")
    return 0


if __name__ == "__main__":
    sys.exit(main())

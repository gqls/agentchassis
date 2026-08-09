#!/usr/bin/env python3
"""oracles.py — the expected answers, computed from PUBLISHED DEFINITIONS.

INDEPENDENCE IS THE ENTIRE VALUE OF THIS FILE, so it is worth being exact about
what that means here. Nothing below was transcribed from, or checked against,
any calculator on loanandmortgagecalculator.co.uk. Every function is written
from a source that exists outside this estate:

  * the standard annuity (amortising loan) formula, M = P·r / (1 − (1+r)^−n);
  * straight-line amortisation month by month, which is the definition of what
    that payment does to a balance;
  * the SDLT bands and reliefs as published by HMRC on gov.uk (cited inline,
    with the date they took effect);
  * arithmetic identities — a yield is annual rent over price, an LTV is loan
    over value, a compound roll-up is P·(1+r)^n.

An oracle transcribed from the page's own <script> is not an oracle; it is the
same claim written twice, and it will agree with the bug. That failure mode is
not hypothetical on this site — see `bugs_open/` for `calcSDLT`, which a
transcribed oracle would have certified.

WHERE A CONVENTION EXISTS, THIS FILE STATES IT RATHER THAN PICKING ONE SILENTLY.
"Total interest" has two defensible readings: n·M − P using the exact payment,
or n·round(M) − P using the payment the borrower is actually billed. They differ
by up to n·£0.005 (26p over five years) and BOTH appear on this site. So the
functions return both and the comparator reports which convention a tool
matches, rather than convicting it of a rounding choice.
"""
import math

# ---------------------------------------------------------------------------
# Amortising credit — the annuity formula
# ---------------------------------------------------------------------------


def monthly_payment(principal, annual_rate_pct, months):
    """M = P·r/(1−(1+r)^−n), r the monthly rate. r = 0 is the straight division.

    The zero-rate limb is not an edge case to be tidied away: it is the correct
    limit (as r→0, M→P/n) and it is a boundary vector, because the general form
    divides by zero there. A calculator that returns NaN, Infinity or £0.00 on a
    0% APR is wrong, and only a vector that lands exactly on 0 can show it.
    """
    if months <= 0:
        raise ValueError("months must be positive")
    r = (annual_rate_pct / 100.0) / 12.0
    if r == 0:
        return principal / months
    return principal * r / (1.0 - (1.0 + r) ** (-months))


def total_interest(principal, annual_rate_pct, months, round_payment=False):
    """n·M − P. `round_payment` selects the billed-payment convention."""
    m = monthly_payment(principal, annual_rate_pct, months)
    if round_payment:
        m = round(m, 2)
    return m * months - principal


def amortise(principal, annual_rate_pct, months, extra=0.0):
    """Run the schedule month by month; return (months_taken, interest_paid).

    Interest is charged on the opening balance, then the payment (contractual +
    extra) is applied; the final month pays only what is left. This is the
    definition of amortisation, not a model of it — which is why it can check a
    closed-form overpayment claim without sharing any code with one.
    """
    m = monthly_payment(principal, annual_rate_pct, months) + extra
    r = (annual_rate_pct / 100.0) / 12.0
    bal = float(principal)
    interest = 0.0
    n = 0
    # A payment that does not cover the monthly interest never clears the debt;
    # bound the loop and say so rather than spinning.
    while bal > 1e-9:
        i = bal * r
        if m <= i + 1e-9 and r > 0:
            raise ValueError("payment %.2f does not cover interest %.2f — "
                             "balance never clears" % (m, i))
        interest += i
        bal = bal + i - m
        n += 1
        if bal < 0:
            interest += 0.0        # overshoot is refunded, not charged
            bal = 0.0
        if n > 12000:
            raise ValueError("schedule did not terminate in 1000 years")
    return n, interest


def overpayment_saving(principal, annual_rate_pct, months, extra):
    """(interest_saved, months_saved) from adding `extra` to every payment."""
    base_n, base_i = amortise(principal, annual_rate_pct, months, 0.0)
    over_n, over_i = amortise(principal, annual_rate_pct, months, extra)
    return base_i - over_i, base_n - over_n


def balance_after(principal, annual_rate_pct, months_total, months_elapsed):
    """Balance remaining after `months_elapsed` contractual payments."""
    m = monthly_payment(principal, annual_rate_pct, months_total)
    r = (annual_rate_pct / 100.0) / 12.0
    bal = float(principal)
    for _ in range(months_elapsed):
        bal = bal * (1.0 + r) - m
    return bal


def reprice_schedule(principal, term_months, windows):
    """Payments through a sequence of rate windows on ONE mortgage.

    `windows` is [(rate_pct, months_in_window), ...]. Each window's payment
    amortises the balance REMAINING at its start over the term REMAINING, which
    is what a remortgage onto a new rate actually does: the debt already repaid
    does not come back, and the clock does not restart.

    > **CORRECTED 2026-08-08.** This function did not exist in the first version
    > of the oracle, which asserted the naive model — the payment on the FULL
    > original principal over the FULL original term at each new rate — and
    > therefore reported 4 FAILs against `mortgages/rate-forecaster`. The page
    > was right and the oracle was wrong: re-running the definition above
    > reproduces the page's £1,526 and £1,286 to the pound. The naive model is
    > kept as a NAMED WRONG ANSWER in the checks (`defect_alt`), because it is
    > the error a future rewrite is most likely to introduce, and because a
    > refuted hypothesis is worth more written down than deleted.
    """
    out = []
    bal = float(principal)
    left = term_months
    for rate, span in windows:
        out.append(monthly_payment(bal, rate, left))
        step = min(span, left)
        if step <= 0:
            break
        bal = balance_after(bal, rate, left, step)
        left -= step
    return out


def compound(principal, annual_rate_pct, years, per_year=1):
    """P·(1+r/k)^(k·y) — the roll-up. k=1 is annual compounding."""
    r = (annual_rate_pct / 100.0) / per_year
    return principal * (1.0 + r) ** (per_year * years)


# ---------------------------------------------------------------------------
# SDLT — England & Northern Ireland, in force from 1 April 2025
# ---------------------------------------------------------------------------
#
# SOURCES (fetched 2026-08-08, quoted in REPORT §2):
#   https://www.gov.uk/stamp-duty-land-tax/residential-property-rates
#   https://www.gov.uk/guidance/stamp-duty-land-tax-buying-an-additional-residential-property
#   https://www.gov.uk/hmrc-internal-manuals/stamp-duty-land-tax-manual/sdltm29805
#
# Standard residential (single dwelling, main residence):
#   up to £125,000 .................. 0%
#   £125,001 – £250,000 ............. 2%
#   £250,001 – £925,000 ............. 5%
#   £925,001 – £1,500,000 ........... 10%
#   above £1,500,000 ................ 12%
#
# First Time Buyers' Relief, from 1 April 2025:
#   0% to £300,000; 5% on £300,001–£500,000;
#   "If the purchase price is more than £500,000, you cannot claim the relief
#    and you must pay the standard rates on the total purchase price."
#   (Between 23 Sep 2022 and 31 Mar 2025 the cap was £625,000 — that is the
#    SUPERSEDED rule, and the reason this constant is named for its date.)
#
# Higher rates for additional dwellings: standard + 5 percentage points on every
# band, and they apply only where the chargeable consideration is £40,000 or
# more — "the property is worth less than £40,000" is excluded entirely.

SDLT_BANDS = [(125000, 0.00), (250000, 0.02), (925000, 0.05),
              (1500000, 0.10), (float("inf"), 0.12)]

FTB_NIL_BAND = 300000          # 0% up to here
FTB_RELIEF_CAP = 500000        # relief unavailable ABOVE this (from 2025-04-01)
FTB_RELIEF_CAP_SUPERSEDED = 625000   # the 2022-09-23 .. 2025-03-31 cap
SURCHARGE_ADDITIONAL = 0.05
SURCHARGE_FLOOR = 40000        # higher rates apply at £40,000 or more


def _banded(price, surcharge=0.0):
    tax = 0.0
    lower = 0.0
    for upper, rate in SDLT_BANDS:
        if price <= lower:
            break
        slice_ = min(price, upper) - lower
        tax += slice_ * (rate + surcharge)
        lower = upper
    return tax


def sdlt(price, buyer):
    """buyer in {'standard', 'ftb', 'additional'}. Returns tax in £."""
    if price <= 0:
        return 0.0
    if buyer == "ftb":
        if price > FTB_RELIEF_CAP:
            # Relief withdrawn entirely — standard rates on the WHOLE price.
            return _banded(price, 0.0)
        if price <= FTB_NIL_BAND:
            return 0.0
        return (price - FTB_NIL_BAND) * 0.05
    if buyer == "additional":
        if price < SURCHARGE_FLOOR:
            # Higher rates do not apply; standard rates do, and below £125,000
            # the standard charge is nil.
            return _banded(price, 0.0)
        return _banded(price, SURCHARGE_ADDITIONAL)
    return _banded(price, 0.0)


def sdlt_superseded_ftb(price):
    """What the 2022–2025 temporary rule gave. Used ONLY to demonstrate that a
    tool's answer matches the expired rule rather than the current one — i.e. to
    name the defect precisely instead of just reporting a mismatch."""
    if price > FTB_RELIEF_CAP_SUPERSEDED:
        return _banded(price, 0.0)
    if price <= FTB_NIL_BAND:
        return 0.0
    return (price - FTB_NIL_BAND) * 0.05


# ---------------------------------------------------------------------------
# Ratios and identities
# ---------------------------------------------------------------------------


def gross_yield_pct(monthly_rent, price):
    """Gross annual yield = 12·rent / price, as a percentage."""
    return (monthly_rent * 12.0) / price * 100.0


def ltv_pct(loan, value):
    return loan / value * 100.0


def bridging_gross(net, monthly_rate_pct, months, fee_pct):
    """Retained-interest bridging: the GROSS facility from which the interest
    and the arrangement fee are deducted, leaving `net` in hand.

        gross · (1 − months·rate − fee) = net

    This is the standard retained/rolled-up construction; the alternative
    ("serviced") model charges interest monthly and grosses up only by the fee.
    Both are returned so the comparator can say which one a tool implements
    instead of asserting a house style.
    """
    deduction = (monthly_rate_pct / 100.0) * months + (fee_pct / 100.0)
    if deduction >= 1.0:
        raise ValueError("interest + fee consume the whole facility")
    gross = net / (1.0 - deduction)
    interest = gross * (monthly_rate_pct / 100.0) * months
    fee = gross * (fee_pct / 100.0)
    return {"gross": gross, "interest": interest, "fee": fee}


def early_settlement_58day(balance, annual_rate_pct, days=58):
    """Consumer Credit (Early Settlement) Regulations 2004: the settlement date
    is deferred by 28 days, plus up to a further 30 days where the agreement
    runs more than 12 months — the "58 days' interest" a lender may keep.
    Simple interest on the outstanding balance for that period."""
    return balance + balance * (annual_rate_pct / 100.0) * (days / 365.0)


def income_multiple(income_total, monthly_commitments, multiple):
    """Lender affordability sketch: annualise commitments and deduct at the same
    multiple. (income − 12·commitments) · multiple."""
    return (income_total - monthly_commitments * 12.0) * multiple


def deal_period_cost(amount, rate_pct, deal_years, fee, other_fees, term_years=25):
    """Total cost over an initial deal period = payments made + fees.

    The payment is the one for the FULL mortgage term at the deal rate — that is
    what a lender bills during a 2-year fix on a 25-year mortgage. Computing it
    over the deal period alone would treat a 2-year fix as a 2-year mortgage and
    inflate the monthly payment several-fold; that is a distinct, checkable
    error, so `term_years` is explicit rather than assumed.
    """
    n_deal = int(round(deal_years * 12))
    m = monthly_payment(amount, rate_pct, int(round(term_years * 12)))
    return {"payments": m * n_deal, "total": m * n_deal + fee + other_fees,
            "monthly": m}


def car_finance(price, deposit, apr, years, balloon=0.0):
    """HP is an annuity on (price − deposit). PCP defers a balloon to the end:
    the monthly payment amortises the balance down to the balloon, not to zero.

        M = (P − B·(1+r)^−n) · r / (1 − (1+r)^−n)

    i.e. the annuity on the balance net of the DISCOUNTED balloon. Both are
    returned; which one a page implements is a finding, not an assumption.
    """
    p = price - deposit
    n = int(round(years * 12))
    r = (apr / 100.0) / 12.0
    hp = monthly_payment(p, apr, n)
    if r == 0:
        pcp = (p - balloon) / n
    else:
        pcp = (p - balloon * (1.0 + r) ** (-n)) * r / (1.0 - (1.0 + r) ** (-n))
    return {"hp_monthly": hp, "hp_interest": hp * n - p,
            "pcp_monthly": pcp, "pcp_interest": pcp * n + balloon - p}

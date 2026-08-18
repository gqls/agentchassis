# First 50 — build order — **APPROVED by the owner 2026-08-18** ("I approve your order of the first 50, do as you direct")

> **ONE CHANGE MADE UNDER THAT AUTHORITY, and it is a reorder not a substitution.**
> Wave 1 now leads with **M5 `adversecreditmortgage.co.uk`** (dispatched 2026-08-18,
> corr `4a128539-8a51-4e4c-a410-15dffa3f6946`); **M3 `mortgage-rates.co.uk` moves to the
> end of Wave 1, pending an owner decision** — see the box below. Nothing was dropped.

Requested by the owner 2026-08-18. Ordering follows his standing **M → B → I** ruling.
**Nothing here is dispatched until he approves this list.**

## The one thing that should change the order, or the schedule

**The mortgage-lender register has 2 lenders. Savings has 13, health-insurer has 10.**
(Measured 2026-08-18: `directory_entities` active, with current found claims — mortgage 2
entities/3 claims, savings-provider 13/15, health-insurer 10/15.)

M is the family the owner ruled FIRST and it is the family with the LEAST directory data. The
pilot's live directory page shows exactly two building societies. Eleven mortgage sites built
this week would each carry that same two-row directory.

**Recommendation, cheap and already proven:** force-trigger the finance researcher on
`mortgage-lender` a handful of times before Wave 2, exactly as B4 did (`last_triggered_at=NULL`
on the scheduled task; supervised, HITL queue worked per run). It is the difference between a
2-row and a 15-row directory on every mortgage site in the fleet. **It does not block Wave 1.**

## Excluded, and why

| | |
|---|---|
| **M1** `loanandmortgagecalculator.co.uk`, **L1** `loancalculator.co.uk` | LIVE + adopted |
| **M2** `mortgagecalculator.co.uk` | handed off for adoption; site exists |
| **M4** `remortgagecalculator.uk` | **the pilot — already built** (3 of 6 pages deployed) |
| **L10** `loancash.co.uk` | BUILDING (owner direction, other lane) |
| **B8** savingsapp, **B9** bankingequipment, **I10** insurance brandables | **HOLD** — see the decisions below |
| **L9** loan brandables | **reassigned 2026-08-18: `loanzy.uk` is now an EXAMPLE SITE** — no register entry, built from the webdesign.uk prompt, separate thread |

## ⚠ M3 `mortgage-rates.co.uk` — a conflict I should not resolve alone

M3's whole proposition is **showing mortgage rates** ("the market, today", stance: data
journalist). Two things in force collide with that:

1. **The finance-kinds compliance ruling (owner, 2026-08-12/13): NON-PRICE facts only** for
   named regulated firms — never APR/rates/premiums, because a stale price under a lender's
   name is a financial-promotion exposure.
2. **We have no verified-facts pipeline for rates.** The pilot shipped with an EMPTY facts
   roster precisely because no rate could be verified against a live source. M3 seeded the
   same way would be **a rate site with no rates**.

There is a defensible reading — the *directory* stays non-price while the site's own editorial
carries **dated, sourced, market-level** figures (base rate, published averages) rather than
per-lender live prices. That is ordinary data journalism and not a promotion. But it needs a
source of truth we do not currently have, and choosing to publish rates at all is the owner's
call, not mine. **So M3 is held at the end of Wave 1 rather than built empty.**

**The decision needed:** (a) M3 becomes market-level/dated with a named data source we commit
to maintaining; (b) M3 is reshaped away from live rates (e.g. "how mortgage pricing works");
or (c) M3 waits until a rates-facts pipeline exists.

## Wave 1 — supervised, ONE AT A TIME (5 domains)

The owner asked to see each site built directly through the framework. These five stress
different parts deliberately: a rate-table site, an eligibility/decision-tree site, an
age-gated calculator, a B2B-ish audience, and a deadline/urgency site.

| # | domain | prop | why this one, this early |
|---|---|---|---|
| 1 | `adversecreditmortgage.co.uk` | M5 | **DISPATCHED 2026-08-18.** Eligibility decision-trees + the register's own hard compliance rule ("no guaranteed acceptance, ever") — so build #1 stress-tests the claims guard on the audience where getting it wrong does real harm |
| 5b | `mortgage-rates.co.uk` | M3 | **HELD** — see the conflict box above |
| 3 | `equityreleasecalculator.co.uk` | M6 | age-gated (55+) audience; long-form + calculator |
| 4 | `buytoletcalculator.uk` | M9 | landlord//investor audience — different register entirely |
| 5 | `mortgageextension.co.uk` | M7 | payment-pressure/urgency; nearest to the pilot, so it is the comparison case |

## Wave 2 — remaining mortgage + the cross-cutting pair (8)

Batch in pairs once Wave 1 reads clean. **Run the mortgage-lender researcher before this wave.**

`consolidatemortgage.co.uk` (M8) · `offset-mortgage.co.uk` (M10) · `bespokemortgages.uk` (M11) ·
`smbmortgages.co.uk` (M12) · `mortgagerepaymentsinsurance.co.uk` (X1) · `private-finance.uk` (X2) ·
`mortgageinterestcalculator.co.uk` (M2-also: interest mechanics) ·
`mortgage-refinance.co.uk` (M4-also: the "refinance" searcher)

> ⚠ `mortgage-refinance.co.uk` is the domain that exposed **bugs_closed/292** (it contains both
> "mortgage" and — inside "refinance" — "finance", which flipped its directory decision per run).
> Fixed and live in v1.0.1307. It is safe to build; watch that it gets the lender directory.

## Wave 3 — savings / rates (8) — the RICHEST directory data

`savingsrates.co.uk` (B1) · `highestinterest.co.uk` (B2) · `buildingsocietyrates.co.uk` (B3) ·
`savingsaccountrates.co.uk` (B4) · `interestrates.co.uk` (B5) · `saving-rate.co.uk` (B6) ·
`rateforecast.co.uk` (B7) · `highestrate.co.uk` (B2-also: cross-product)

## Wave 4 — insurance (11)

`privatehealthinsurancequote.co.uk` (I1) · `comparefamilyhealthinsurance.co.uk` (I2) ·
`corporatehealthinsurance.uk` (I3) · `incomeprotectioncover.co.uk` (I4) ·
`landlordinsurancerates.co.uk` (I5) · `lifeinsurancequotation.co.uk` (I6) ·
`homeinsurancerates.co.uk` (I7) · `keypersoncover.co.uk` (I8) · `indemnitycover.co.uk` (I9)
— plus two seconds to be picked from I1/I6 once the primaries read clean.

## Wave 5 — loans/debt (10)

Later than M/B/I per the owner's order, and partly adjacent to another lane's live sites
(L1/L10) — coordinate before dispatch.

`whichloan.co.uk` (L2) · `consolidateloans.co.uk` (L3) · `longtermloan.co.uk` (L4) ·
`financeforcompanies.co.uk` (L5) · `unsecuredpersonalloans.co.uk` (L6) · `investmentloan.uk` (L7) ·
`fleetfinancing.co.uk` (L8) · `borrowing.co.uk` (L2-also: umbrella authority) ·
`business-lenders.co.uk` (L5-also: lender directory — **a natural second directory consumer**) ·
`companyloan.uk` (L5-also: explainer + calculator)

## Count

5 + 8 + 8 + 11 + 10 = **42 fresh builds**. With the pilot and the eight remaining
angle-carrying seconds held in reserve, that reaches 50 without inventing work. **If the owner
wants exactly 50 dispatched, the eight reserves are named in the register's `also:` lines and
I will list them on request.**

## Cost at this scale (from the measured baseline)

42 sites × $3.81 text = **~$160** today, or **~$203** after 2026-09-01. Imagery is the variable:
at 30 images/site it adds **$50–$315** depending on the per-image rate we still have not pinned
(see NOTES 2026-08-18). **So the whole first 50 is roughly $210–$520** — the uncertainty is
almost entirely imagery, not text.

# 225 — the SDLT calculator applies the EXPIRED £625,000 First Time Buyer cap, under-quoting a real tax bill by £5,000; and charges the 5% surcharge below the £40,000 floor

**Filed 2026-08-08 by the `loanandmortgagecalculator_couk` lane**, from the
owner-requested arithmetic-validation work
(`docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/HANDOFF_2026-08-08_arithmetic_validation.md`).

Live page: <https://loanandmortgagecalculator.co.uk/mortgages/stamp-duty.html>
Source: `sites` repo, `loanandmortgagecalculator.co.uk/mortgages/stamp-duty.html`,
inline `calcSDLT()`.

**Two defects, one page, opposite signs. Defect A under-quotes a tax the user
will actually pay. Defect B over-quotes one.** Both are boundary-band defects
and neither is reachable from the page's default inputs.

---

## Why no existing check could ever have caught this

The lane's `GOLDEN_2026-08-05_prechange.json` records what every calculator
answered on 2026-08-05 and `golden_compare_post.py` re-runs it. Those prove
CONSISTENCY. `calcSDLT` has been wrong since it was written, so the golden
recorded the wrong number faithfully and every comparison since has certified
it — [[a-pass-from-a-blind-check-outlives-the-blindness]]. Tier-2
`tool_acceptance` says in its own header that its static checks "CONFIRM, never
refute". Nothing in the estate was asking whether the answer was RIGHT.

It was found by an independent oracle — the HMRC bands recomputed in Python from
gov.uk, never from this page — driven at explicit band edges.

---

## Defect A — First Time Buyers' Relief uses the cap that expired 16 months ago

**The rule, from HMRC** (fetched 2026-08-08):

- <https://www.gov.uk/stamp-duty-land-tax/residential-property-rates> — FTB
  relief threshold £300,000; "5% SDLT on the portion from £300,001 to £500,000";
  above £500,000 buyers "cannot claim the relief" and must "follow the rules for
  people who've bought a home before".
- <https://www.gov.uk/hmrc-internal-manuals/stamp-duty-land-tax-manual/sdltm29805>
  — "From 1 April 2025 the relief applies to purchases of residential property
  for £500,000 or less… If the purchase price is more than £500,000, you cannot
  claim the relief and you must pay the standard rates on the total purchase
  price." **"Between 23 September 2022 and 31 March 2025 the relief applied to
  purchases of residential property for £625,000 or less."**

**What the page does.** The FTB branch is gated on `price <= 625000` — the
temporary cap — and for £500k–£625k charges £10,000 plus 5% of the excess over
£500,000:

```js
if (type === 'ftb' && price <= 625000) {
    ...
    } else {
        // Between 500k and 625k - Rules vary, but often relief is capped or removed.
        // Standard practice: if > 500k, standard rates often apply on the whole or the relief is tapered.
        // Let's use Standard Rates for safety if > £625k, but calculate 5% on the chunk above 300k if <625k.
        tax = (200000 * 0.05) + ((price - 500000) * 0.05);
    }
```

**The page contradicts itself in its own prose.** Immediately above the
calculator it says: *"Following the end of the temporary relief period in March
2025, thresholds have reverted to standard levels."* The standard band table
beside it is correct and current. Only the FTB branch is still on the expired
rule — so the copy is right, the bands are right, and the arithmetic is 16
months stale.

**Measured, live, 2026-08-08** (`oracle.py --tools stamp-duty`):

| price, FTB | page shows | correct (HMRC) | error |
|---|---|---|---|
| £500,000 | £10,000 | £10,000 | — (relief still available AT the cap) |
| £500,001 | £10,000 | £15,000.05 | **−£5,000** |
| £600,000 | £15,000 | £20,000 | **−£5,000** |
| £625,000 | £16,250 | £21,250 | **−£5,000** |
| £625,001 | £21,250 | £21,250 | — (falls through to standard) |

**The error is a flat £5,000 across the whole band** `£500,000 < price ≤
£625,000` — algebraically, the page charges `0.05P − 15,000` where the correct
standard charge is `0.05P − 10,000`. It is the exact value of the relief the
buyer is no longer entitled to.

**Why this one matters most.** It is a false number about a tax a user will
actually pay, in the band where a first-time buyer is most likely to be
stretching, and it errs in the direction that under-prepares them: a buyer
budgeting £15,000 will be asked for £20,000 at completion.

## Defect B — the 5% surcharge is charged below the £40,000 higher-rate floor

**The rule**: <https://www.gov.uk/guidance/stamp-duty-land-tax-buying-an-additional-residential-property>
— "You must pay the higher Stamp Duty Land Tax (SDLT) rates when you buy a
residential property (or a part of one) for £40,000 or more"; property "worth
less than £40,000" is excluded from the higher rates entirely.

The page's `else` branch applies `+ surcharge` to band 1 unconditionally:

```js
let band1 = Math.min(remaining, 125000);
tax += band1 * (0 + surcharge);
```

| price, additional property | page shows | correct | error |
|---|---|---|---|
| £39,999 | £2,000 | £0 | **+£2,000** |
| £40,000 | £2,000 | £2,000 | — |

Below £40,000 the higher rates do not apply, standard rates do, and standard
rates below £125,000 are nil. The page quotes £2,000 of tax on a transaction
that attracts none. Lower impact than A (few residential purchases sit under
£40,000) but the same class and a one-line fix.

## What is CORRECT on this page

Worth stating, because "the SDLT calculator is broken" would be wrong. 13 of 17
boundary vectors pass: every standard-rate band edge (£125,000 / £125,001 /
£250,000 / £925,000 / £1,500,000 / £1,500,001), the FTB nil band and its
ceiling, and the higher rates at and above the floor. The band table and the
5-percentage-point surcharge construction are right and current.

## Fix candidates, ordered by what closes the door

1. **Replace `calcSDLT`'s hand-rolled branches with one banded function plus a
   relief predicate**, and put the thresholds in NAMED, DATED constants
   (`FTB_RELIEF_CAP_2025_04_01 = 500000`). The present code cannot be read for
   correctness because the rule is spread across a gate, three branches and a
   comment that admits uncertainty. This makes the expired-rule state
   unrepresentable rather than merely fixed — you cannot leave a stale cap in
   place if the cap is a dated constant that a reviewer reads at a glance.
2. **Add the `price < 40000` early return** for `additional`.
3. **Regression-lock it**: `oracle.py`'s 17 stamp-duty vectors already encode
   every band edge and both defects. Run it after the fix; it must go to 17/17.

## How to verify

```bash
cd docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
python3 oracle.py --tools stamp-duty
```

Today: `PASS 13   FAIL 4`. After a correct fix: `PASS 17   FAIL 0`.
The oracle's own controls (`--mutate expectation`, `--mutate crosstool`,
`--selftest-parse`) must be re-run alongside, or a green result proves nothing.

## What NOT to do

Do not change the number without the owner seeing this file first. These are
consumer tax figures and a changed answer is a changed claim
(brief §9). The correction is well-evidenced and cited above, but the decision
to publish a different tax number belongs to the owner.

## Related

> **Numbered 225, not 223.** This file was written as 223 and renumbered
> before it was committed: a concurrent session committed a different
> `bugs_open/223` (the landmine verifier's non-Go footprints) in the same hour.
> Theirs was already at HEAD, so it keeps the number.

- `bugs_open/224` — the zero-rate defect family on the same site. Different
  mechanism (duplicated formula), same root discovery method.
- Report with the full method, controls and refutations:
  `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/REPORT_2026-08-08_arithmetic_validation.md`

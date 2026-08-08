# HANDOFF — arithmetic validation for loanandmortgagecalculator.co.uk's 23 calculators

**Owner-requested, 2026-08-07/08:** *"build the arithmetic validation and 'is
the tool doing the right thing' checks."* This is that brief.

Read first: this lane's `NOTES` tail (2026-08-06 entries), `RUNBOOK` §12,
and `acceptance/GOLDEN_2026-08-05_prechange.json` (what exists today).

---

## 1. The gap, stated exactly

We have **behavioural goldens**: `toolgolden.py` drives each calculator in a
real browser and records every id-bearing element's text after each input
vector. `golden_compare_post.py` re-runs them and diffs.

**Those prove CONSISTENCY, not CORRECTNESS.** They answer *"does it still do
what it did on 2026-08-05?"*. They cannot answer *"is that the right answer?"*.
If a calculator has been wrong since the day it was written, the golden records
the wrong answer faithfully, and every future check certifies the bug. This is
the [[a-pass-from-a-blind-check-outlives-the-blindness]] shape, and this site
has already produced one instance: `loans/credit-health-check` was broken on
the SOURCE site by CSS that was never written (07-31 session 1), and nothing
computational noticed, because nothing computational was asking.

**What closes it is an INDEPENDENT ORACLE**: recompute the expected answer from
the definition — the standard amortisation formula, the published SDLT bands,
the definition of loan-to-value — in code that has never seen the page's
JavaScript, and compare. Independence is the whole value. An oracle
transcribed from the page's own `<script>` is not an oracle; it is the same
claim written twice, and it will agree with the bug.

---

## 2. Classification of the 23 — do this honestly, it is part of the deliverable

**A. Closed-form analytic (14).** An external formula gives one right answer.
These are the deliverable's core.

| page | function | oracle |
|---|---|---|
| `loans/standard-calc` | `calculateLoan` | annuity: `M = P·r/(1−(1+r)^−n)`, r = APR/12 |
| `mortgages/repayment` | `runCalc` | same annuity + amortisation schedule |
| `mortgages/simple` | `doSimpleCalc` | same annuity |
| `loans/overpayment-calculator` | `calc` | annuity, then months-to-zero with extra payment |
| `mortgages/overpayment` | `runOverpayment` | as above (shares `/assets/js/calculators.js`) |
| `loans/consolidation` | `calcRisk` | per-debt annuity total interest vs consolidated annuity |
| `loans/compare-loans` | `calc`,`compare` | total cost per option = M·n |
| `loans/interest-rate-stress-test` | `stressTest` | annuity at base rate vs base+shock |
| `mortgages/rate-forecaster` | `calcForecast` | annuity at new rate; delta vs current |
| `loans/loan-vs-savings` | `compare` | loan interest saved vs savings interest forgone |
| `loans/settlement-calculator` | `estSettle` | early settlement — **name the rule** (Rule of 78 vs actuarial); they differ materially |
| `mortgages/investor` | `calcYield`,`calcLTV` | yield = rent·12/price; LTV = loan/price |
| `mortgages/bridging-loan` | `calcBridge` | rolled/retained monthly interest + fees |
| `mortgages/equity-release` | `calcCompound` | compound roll-up `P·(1+r)^n` |

**B. Regulatory — the right answer is external and DATED (3).** The oracle must
cite a source, not the page.

| page | function | oracle source |
|---|---|---|
| `mortgages/stamp-duty` | `calcSDLT` | HMRC SDLT bands + FTB relief + additional-property surcharge, 2025/26 |
| `mortgages/affordability` | `calcAffordability` | income multiple + commitment deduction — the page's own model, but the FPC 4.5× LTI framing is checkable |
| `mortgages/fee-analyser` | `calcTrueCost` | total cost = payments + fees; arithmetic is closed-form, fee TREATMENT is a policy choice — state it |

**C. Heuristic / stateful — NO external right answer exists (6).** Do not
invent one. Validate INVARIANTS instead (monotonicity, bounds, state
round-trip), and say plainly in the report that these are not arithmetic
checks.

`loans/credit-health-check` (wizard scoring) · `loans/damage-checker`
(checkbox verdict) · `loans/application-tracker` (checklist + localStorage) ·
`mortgages/fact-finder` (`calcScore`) · `mortgages/portfolio` (aggregates —
the AGGREGATION is checkable, the scoring is not) · `loans/car-finance-calculator`
(`calcCar` — PCP balloon handling is closed-form, but check which convention
the page uses before classifying it A).

Sum: 14 + 3 + 6 = 23. **If your classification disagrees with this one, that
disagreement is a finding — record it, do not silently re-bucket.**

---

## 3. CANDIDATE DEFECT ALREADY IN HAND — start here, it validates the method

`mortgages/stamp-duty`'s `calcSDLT` **contains its author's own admission of
uncertainty, in comments, in the shipped file** (sites-repo `b318a8fad`):

```
// Between 500k and 625k - Rules vary, but often relief is capped or removed.
// Standard practice: if > 500k, standard rates often apply on the whole or the relief is tapered.
// Let's use Standard Rates for safety if > £625k, but calculate 5% on the chunk above 300k if <625k.
tax = (200000 * 0.05) + ((price - 500000) * 0.05); // 5% on amount over 300k
```

The FTB branch is gated on `price <= 625000`, and for £500k–£625k it charges
£10,000 plus 5% of the excess over £500k. **[UNVERIFIED — verify against HMRC,
not against this code]** the actual rule is that First Time Buyers' Relief is
withdrawn entirely above a purchase-price cap, in which case standard rates
apply to the whole price and this branch under-quotes the tax.

This is the ideal first oracle because:
- it is a **false number shown to a user about a tax they will actually pay**;
- **no golden could ever catch it** — the golden records whatever this returns;
- the page's own comments are evidence the author was guessing, so it is a
  live hypothesis rather than a fishing expedition.

Verify the rule from an authoritative source and cite it inline. Do NOT
transcribe the fix from the page.

---

## 4. Method

1. **Drive with EXPLICIT vectors, not scaled defaults.** `toolgolden`'s
   x1/x2/x0.5 scaling exists to stay in-domain for any tool with no
   per-tool config; an oracle knows its tool, so choose vectors that land on
   and around the **boundaries** — band edges (£125k, £250k, £300k, £500k,
   £625k, £925k), 0% APR, term = 1 month, a balloon of zero. The golden's
   neighbourhood-of-defaults vectors are exactly where a boundary defect
   hides (sibling lane learned this: two of four real fixes were invisible to
   its golden).
2. **Compute expected in Python from the definition.** Independent module, no
   reference to the page's script.
3. **Compare with a stated tolerance** — these pages format to pence via
   `Intl.NumberFormat`, so compare to ±£0.01 after parsing, and state the
   parse (strip `£`, `,`, `%`).
4. **Report PASS / FAIL / NOT-APPLICABLE per tool per vector**, and make
   NOT-APPLICABLE loud rather than absent.

## 5. Prove the checker can fail — non-negotiable

A validator that agrees with everything is worse than none. Before trusting a
green run:
- **Mutate the expectation** (assert £100 where the tool says £202.29) and
  require FAIL.
- **Mutate the input parse** (feed the raw `£1,234.56` string unparsed) and
  require FAIL, not a silent 0.
- **Cross-tool control**: run `standard-calc`'s oracle against `simple`'s page
  with different inputs and require FAIL.
This lane has already had FIVE adverse verdicts that were the instrument, not
the site (07-31 sessions 1–2; the `investor` refusal on 08-05). **On this site,
the prior probability that a red result is your harness is high — but so is the
probability that a green one is blind.** Both need a control.

## 6. Reuse, do not rebuild

- `loancalculator_couk/toolgolden.py` — `Runner`, and especially `settle()`:
  it waits for a FULLY PARSED document. A bare `readyState=='complete'` poll
  once certified a golden where every answer was £0.00 because the inline
  script had not parsed. Reuse it verbatim.
- Storage clearing + reload between vectors — `application-tracker` and
  `portfolio` persist to `localStorage`; a contaminated baseline is perfectly
  self-consistent.
- `window.confirm`/`alert` stubs — a modal blocks the renderer and the next
  evaluate simply times out with no indication why.
- `loanandmortgagecalculator_couk/investor_golden.py` — the staggered-vector
  pattern for ratio tools.
- `golden_compare_post.py` — the `content`-field shape assertion, needed on
  any decomposed page (see §7).

## 7. Traps specific to this site

- **A ratio tool is invariant under uniform scaling.** `mortgages/investor`
  cannot be certified by `toolgolden` at all; this is in `LANDMINES.md`.
- **`id="content"` changes meaning after decomposition** — it was the page
  wrapper (holding all text), it is now an empty span in the shared header.
  Any fingerprint keyed on ids must assert the new shape, not ignore the field.
- **12 mortgages pages share `/assets/js/calculators.js`** — a defect there
  hits all 12 at once, and a per-page oracle that agrees with itself 12 times
  has NOT corroborated anything. Treat shared-helper tools as ONE unit of
  evidence, not twelve.
- **Two tools compute in the same page** (`investor`: yield AND LTV;
  `equity-release`: `calcEquityRelease` AND `calcCompound`). Drive and assert
  both; a per-page pass/fail hides one of them.
- **`mortgages/portfolio` needs seeded state** — it renders from
  `localStorage` (`PORTFOLIO_KEY = 'uk_mortgage_portfolio_v1'`). Seed
  deterministically, then assert the aggregate.

## 8. Acceptance criteria

- Every tool in class A and B has an oracle, or a written reason it does not.
- Every oracle has run its mutation control, and the control's failure output
  is in the report.
- Boundary vectors, not just defaults.
- The SDLT FTB question in §3 is answered against a cited external source, and
  the answer is recorded whichever way it comes out — **a REFUTED hypothesis is
  a success and costs one run**.
- Findings that are real defects → `bugs_open/` with the usual evidence, and
  the diagnosis loop (`090`) if the cause looks structural rather than local.
- Class C tools have INVARIANT checks, clearly labelled as not arithmetic.

## 9. What NOT to do

- Do not "fix" a calculator to agree with the oracle without the owner seeing
  the finding first: several of these are consumer-credit figures, and a
  changed answer is a changed claim.
- Do not transcribe expected values from the page's script.
- Do not emit criteria into the platform's acceptance record from a tool you
  have not validated — that pins the wrong answer into the record and defends
  it (`toolgolden`'s own `--emit-criteria` refuses for this reason).
- Do not run against a page mid-deploy: B2 serves a `NoSuchKey` blob at HTTP
  200 and every check silently reads zero.

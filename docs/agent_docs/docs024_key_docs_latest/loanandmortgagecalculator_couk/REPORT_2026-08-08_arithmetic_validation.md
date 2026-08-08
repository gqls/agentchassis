# REPORT — arithmetic validation of the 23 calculators, 2026-08-08

Answers the brief in `HANDOFF_2026-08-08_arithmetic_validation.md`, which
answers the owner's *"build the arithmetic validation and 'is the tool doing the
right thing' checks."*

Everything below was measured against the **live site** on 2026-08-08, in a real
headless browser, at explicitly chosen boundary vectors.

---

## 1. Headline

| | |
|---|---|
| tools with an independent oracle | **18 of 23** (all of class A and B) |
| oracle checks run | **176** — 143 PASS, 27 FAIL, 6 CONVENTION |
| class C invariant checks | **21** — 21 PASS (after two harness faults were fixed) |
| **real defects found** | **2 bug files: `bugs_open/225`, `bugs_open/224`** |
| tools affected | **8 of 23** |
| controls run | 4, all red-on-demand; one of them found a bug in the harness |
| oracle claims REFUTED by the site | **1** — recorded, not deleted |

**Two defect families, both invisible to every check that existed before today.**

- **`bugs_open/225` — `mortgages/stamp-duty`** applies the First Time Buyer
  relief cap that **expired on 31 March 2025**, under-quoting SDLT by a flat
  **£5,000** for any FTB purchase between £500,001 and £625,000; and charges the
  5% additional-property surcharge below the £40,000 floor where no higher rate
  applies (+£2,000 at £39,999).
- **`bugs_open/224` — a zero interest rate** breaks **six of the seven**
  `loans/*` calculators, because each re-implements the annuity formula inline
  and only the SHARED `assets/js/calculators.js` copy has a zero-rate branch.
  Three print `£NaN`; three silently leave the previous answer on screen;
  `compare-loans` additionally **declares a 0% loan the more expensive option**.

## 2. The gap this closes, and why nothing caught these before

The lane already had **behavioural goldens** — `toolgolden.py` records what each
calculator answered on 2026-08-05, `golden_compare_post.py` re-runs and diffs.
Those prove **consistency**: *does it still do what it did?* They cannot prove
**correctness**. `calcSDLT` has been wrong since it was written, so the golden
recorded the wrong number faithfully and every comparison since certified it —
the [[a-pass-from-a-blind-check-outlives-the-blindness]] shape. Tier-2
`tool_acceptance` states in its own header that its static checks "CONFIRM,
never refute".

What closes it is an **independent oracle**: the expected answer recomputed from
the definition, in code that has never seen the page's JavaScript.

**How independence was actually maintained, concretely.** An oracle still has to
know which box takes the principal and which element holds the answer, and the
obvious way to learn that — read the page's `<script>` — is one line away from
reading the arithmetic. So `inventory.py` reads each page the way a USER does:
the visible `<label>` bound to each control, the button's visible text, and the
caption above each result box. `oracles.py` then computes from the annuity
formula, the published HMRC bands, and arithmetic identities. **The pages'
calculation bodies were not read until a check had already FAILED**, at which
point reading them is diagnosis, not authorship. Both SDLT defects and all seven
zero-rate defects were found before any of that source was opened.

### The oracle sources, cited

| oracle | source |
|---|---|
| annuity `M = P·r/(1−(1+r)^−n)`, zero-rate limit `P/n` | standard |
| amortisation with overpayment | month-by-month schedule, by definition |
| SDLT standard bands, FTB relief, higher rates, £40,000 floor | [gov.uk residential rates](https://www.gov.uk/stamp-duty-land-tax/residential-property-rates) · [higher rates guidance](https://www.gov.uk/guidance/stamp-duty-land-tax-buying-an-additional-residential-property) · [SDLTM29805](https://www.gov.uk/hmrc-internal-manuals/stamp-duty-land-tax-manual/sdltm29805) |
| early settlement | Consumer Credit (Early Settlement) Regulations 2004 — 28 days + up to 30 more = the "58 days" the page itself names |
| yield, LTV, compound roll-up, retained-interest gross-up | arithmetic identities |

## 3. Classification — and where it DISAGREES with the brief

The brief said 14 A / 3 B / 6 C and asked for disagreement to be recorded rather
than silently re-bucketed. **Mine is 15 A / 3 B / 5 C.** Two disagreements:

1. **`loans/car-finance-calculator` moves C → A.** The brief flagged it "PCP
   balloon handling is closed-form, but check which convention the page uses
   before classifying it A". Checked, by measurement rather than assertion: the
   page amortises `price − deposit` down to a **discounted** balloon, i.e.
   `M = (P − B(1+r)^−n)·r/(1−(1+r)^−n)`. That is the standard PCP convention,
   the oracle matches it at three vectors, and the HP mode matches a plain
   annuity. It has one right answer, so it is class A. (The two alternative
   conventions — undiscounted balloon, balloon ignored — are kept in the checks
   as named wrong answers so a future rewrite that drifts is diagnosed, not just
   failed.)
2. **`mortgages/portfolio` stays C but its AGGREGATE is checked as class A.**
   The brief already said the aggregation is checkable; this makes it explicit —
   7 arithmetic assertions (Σvalue, Σ(value−mortgage), LTV, gross yield, rent
   roll, cashflow, count) against the definitions, all passing, plus a
   localStorage round-trip. The scoring parts remain unvalidatable and are
   labelled so.

A third, smaller correction to the brief: **`loans/overpayment-calculator` does
NOT share `/assets/js/calculators.js`.** The brief lists it as sharing. It does
not load that file at all — it carries its own inline copy of the formula, which
is exactly why it is one of the zero-rate casualties while `mortgages/overpayment`
(which does share) is not. The distinction is the whole of `bugs_open/224`.

**Class A (15)** standard-calc · repayment · simple · overpayment-calculator ·
mortgages/overpayment · consolidation · compare-loans · interest-rate-stress-test
· rate-forecaster · loan-vs-savings · settlement-calculator · investor ·
bridging-loan · equity-release · car-finance-calculator

**Class B (3)** stamp-duty · affordability · fee-analyser

**Class C (5)** credit-health-check · damage-checker · application-tracker ·
fact-finder · portfolio

## 4. Results, per tool

| tool | class | checks | result |
|---|---|---|---|
| `mortgages/repayment` | A | 12 | **12 pass** |
| `mortgages/simple` | A | 4 | **4 pass** |
| `mortgages/overpayment` | A | 9 | **9 pass** |
| `mortgages/rate-forecaster` | A | 9 | **9 pass** — and it refuted my oracle, §6 |
| `mortgages/bridging-loan` | A | 12 | **12 pass** |
| `mortgages/equity-release` | A | 9 | **9 pass** |
| `mortgages/investor` | A | 6 | **6 pass** (yield AND LTV, driven asymmetrically) |
| `loans/loan-vs-savings` | A | 6 | **6 pass** |
| `mortgages/affordability` | B | 6 | **6 pass** |
| `mortgages/fee-analyser` | B | 3 | **3 pass** |
| `loans/consolidation` | A | 16 | 15 pass, **1 FAIL** (0% new loan → £0.00/mo) |
| `loans/interest-rate-stress-test` | A | 9 | 7 pass, **2 FAIL** (`£NaN`) |
| `loans/overpayment-calculator` | A | 8 | 6 pass, **2 FAIL** (`£NaN`, wrong months) |
| `loans/settlement-calculator` | A | 5 | 2 pass, **3 FAIL** (stale) |
| `loans/car-finance-calculator` | A | 10 | 6 pass, **4 FAIL** (stale) |
| `loans/compare-loans` | A | 20 | 15 pass, **5 FAIL** (`£NaN` + inverted verdict) |
| `loans/standard-calc` | A | 15 | 3 pass, 6 CONVENTION, **6 FAIL** (stale) |
| `mortgages/stamp-duty` | B | 17 | 13 pass, **4 FAIL** (expired FTB cap ×3, £40k floor ×1) |

### The 6 CONVENTION results are not defects, and not nothing

`standard-calc` rounds the monthly payment to the penny and THEN multiplies for
total interest (`£2,137.40`); `compare-loans` multiplies the exact payment
(`£2,137.14`). Both are defensible — one is what you will be billed, the other
is the arithmetic total — and they differ by 26p over five years. The comparator
reports these as CONVENTION rather than convicting a rounding choice. **But two
tools on one site answering the identical question differently IS worth the
owner's attention**, and it is why the distinction is reported loudly rather
than tolerated silently.

### Class C — 21 invariant checks, all passing. These are NOT arithmetic checks

Stated plainly as the brief requires: a credit "health" score out of 100, a
damage-charge verdict, a 420-point mortgage scorecard have **no external right
answer**, and inventing one would dress a preference up as arithmetic. What was
checked instead: monotonicity (worse answers never score better), bounds,
determinism, localStorage round-trip, and — for `portfolio` — the aggregation,
which genuinely is arithmetic. All pass. **A passing invariant is much weaker
evidence than a passing oracle** and should not be read as "these five tools are
validated".

## 5. Proving the checker can fail — four controls

A validator that agrees with everything is worse than none. On this site the
prior probability that a red result is the harness is high — but so is the
probability that a green one is blind.

| control | what it does | result |
|---|---|---|
| `--mutate expectation` | asserts £100 where the tool says £1,112 | **16 FAIL / 0 pass** ✔ |
| `--mutate parse` | types `£200,000` into a `type=number` field | **refused before comparison, 0 pass** ✔ |
| `--selftest-parse` | 11 unparseable strings must raise, 10 real formats must parse | ✔ — **and it found a bug in my parser**, below |
| `--mutate crosstool` | drives every vector correctly, then judges it against the NEXT vector's expectations | **28 FAIL / 0 pass** ✔ |
| `--mutate wrongpage` | one tool's vectors against another tool's page | refused at the first selector — kept, but labelled WEAK: it never reaches a comparison, so it says nothing about the comparator |

Sample failure output, `--mutate expectation`:

```
  [ FAIL ] site defaults                          monthly payment
           shown £1,112     oracle 100.0000    delta +1012.0000  ±£0.50 (tool displays 0 dp)
```

Sample, `--mutate crosstool` (the strong one — real page, real drive, borrowed expectation):

```
  [ FAIL ] BOUNDARY top of the 2% band            SDLT payable
           shown £2,500     oracle 36250.0000  delta -33750.0000  ±£0.50 (tool displays 0 dp)
```

**Two of these controls needed their own criterion fixed before they were
honest**, and both fixes are the interesting part:

- `--mutate parse` produced **N/A, not FAIL**, and my first exit rule called
  that "the checker is inert". It was wrong: `<input type=number>` rejects
  `£200,000`, the field holds `''`, and `set()` refuses to drive on — which is
  precisely the "silent 0" the control exists to forbid. The criterion is now
  **"no check may PASS under mutation"**, which FAIL and a loud refusal both
  satisfy.
- `--mutate crosstool` left 4 PASSes that were **non-tests**: this suite packs
  vectors onto adjacent boundaries, and £1,500,000 and £1,500,001 of SDLT differ
  by 12p, inside any tolerance — a borrowed expectation equal to the true one
  cannot fail. Those are now excluded **by name and printed**, not by loosening
  the bar. A fifth PASS was different again and is worth reading: at £39,999 the
  borrowed expectation (£2,000, correct for £40,000) equals what the **buggy**
  page prints, so the mutation did not bite. That is reported as
  `MUTATION DID NOT BITE`, with the true expectation alongside — an excuse that
  only applies where the tool is independently known to be wrong.

**Tolerance is derived, not chosen.** A page formatting to whole pounds
(`£1,390`) cannot be checked to the penny by anyone, so it is asserted at ±£0.50
and every line PRINTS its resolution; pages formatting to pence are held to
±£0.01. One global tolerance would either convict every whole-pound tool or
excuse every penny-level defect.

## 6. What the harness got wrong — four corrections, recorded not deleted

This is the section that would be missing if the report were written at the end
instead of during. Also in `WRONG_CALLS.md`.

1. **My oracle was wrong about `mortgages/rate-forecaster`; the page was right.**
   I asserted the naive model — each window's payment on the FULL original
   principal over the FULL original term — and reported 4 FAILs. The page
   amortises the **balance remaining** at each window's start over the **term
   remaining**, which is what a remortgage actually does. Recomputing from that
   definition reproduces its £1,526 and £1,286 to the pound. **A refuted oracle
   claim costs one run; a wrong root cause in a handoff costs every thread that
   believes it.** The naive model is retained in the checks as a *named wrong
   answer* so a future rewrite that regresses to it is diagnosed rather than
   merely failed.
2. **My own reporting mechanism downgraded the biggest finding to an advisory.**
   The first version had one bucket for "matches a different but defensible
   convention", and the first stamp-duty run used it to label `calcSDLT`
   implementing an **expired tax rule** as a CONVENTION. The machinery for
   naming a cause was the same machinery for excusing one. Split into `alt`
   (defensible → CONVENTION) and `defect_alt` (named wrong answer → still FAIL,
   with a DIAGNOSIS line).
3. **`--selftest-parse` found a real gap in my parser.** `£0 / mo Rent` — a live
   reading from `portfolio`'s `#d_rent` — was refused, because the unit-phrase
   suffix had to start with a letter. That check would have come back N/A: a
   check quietly not made, which reads exactly like a check made.
4. **Two "class C defects" were my harness.** The credit-health-check walk
   clicked the result panel's **"Start Over"** button (`location.reload()`),
   destroying the verdict it had just produced and reporting a non-deterministic
   wizard. The application-tracker round-trip reloaded immediately, before the
   notes field's **1-second debounce** fired, and reported that notes do not
   persist. `toolgolden.PRESS_JS` already excludes reset-ish buttons and says
   why in its comment — **reusing a harness while leaving its hard-won
   exclusions behind is how a lesson gets re-learned at full price.** The
   tracker check now waits for the tool's own "Saved" status rather than
   sleeping a number I picked.

## 7. What is NOT covered — said loudly rather than left absent

- **Ten of the eighteen oracles have never seen a failing tool.** They pass, and
  a suite that has only ever agreed is weakly evidenced no matter how many green
  ticks it prints. The controls address the comparator; they do not prove that
  these particular expectations are the right ones.
- **`mortgages/affordability` and `mortgages/fee-analyser` are validated against
  their own stated models**, not an external authority. The 4.5× figure is the
  FPC's LTI limit and the page's own caption, and the fee arithmetic is
  closed-form — but the CHOICE to deduct commitments at the multiple, and to
  treat fees as a flat addition, are the site's policy and are not checkable.
  This is stated in the oracle's own `oracle` line for each.
- **`settlement-calculator`'s model is thinner than its answer looks.** The
  58-day deferment arithmetic is right, and the page has no term input, so it
  cannot compute the actuarial rebate a real settlement figure needs. The
  arithmetic passes; the model is an estimate and the page says so.
- **Nothing here validates a tool against another lender's published figure.**
  Each oracle checks the definition, not the market.
- **No criteria were emitted into the platform's acceptance record** (brief §9):
  emitting from a tool not yet proven correct pins the wrong answer into the
  record and then defends it. That step is available once 224 and 225 are fixed.

## 7b. The diagnosis loop ran on the structural claim and returned NO VERDICT

CLAUDE.md requires a `bugs_open/` file asserting a cross-cutting root cause to
go through `090` first. It was filed (intake `fe69a7b8-…`, run
`3e18a949-8732-4603-b19b-f0c159860fa5`), it was claimed in under a minute, and
it finished in ~9 minutes with **five `bundle` artifacts and no verdict** —
while reporting `COMPLETED` and `status='complete'`.

Measured cause, not inferred: `code_symbols` holds **one repo
(`gqls/agentchassis`) and one extension (`.go`), 5,755 symbols**; zero rows
match `calculators.js`, `standard-calc` or `loanandmortgage`. The files the
symptom names are `.html` and `.js` in the **`sites`** repo. The agent fetched
the DB half (`page_sections` rows) and could not reach the JavaScript the claim
is about.

**The shape worth carrying: a `090` run on a non-Go artefact terminates as a
success.** Nothing distinguishes "looked and found nothing" from "structurally
unable to look". Same Go-only index that makes the landmine verifier report
unindexed footprints as non-existent (`bugs_open/223`, filed by another lane the
same day) — one index, two consumers, and in both the silence reads as a
finding. Recorded in `LANDMINES.md`; `bugs_open/224` states the substitution
explicitly rather than omitting it, which is the escape hatch the norm provides.

The substituted verification does not lean on my having read the right file:
every failure was reproduced in a real browser against the live site, and the
determinism result — same on-screen inputs, two different answers — was
established with no source reading at all.

## 8. Files

| file | what |
|---|---|
| `oracles.py` | the expected answers, from published definitions. No page knowledge. |
| `oracle_driver.py` | CDP driver; subclasses `toolgolden.Runner`; refuses a mid-deploy NoSuchKey blob, verifies every input landed, refuses to parse garbage as a number |
| `oracle.py` | the 18 tool specs, ~55 boundary vectors, the comparator, and the four controls |
| `invariants.py` | class C: monotonicity, bounds, determinism, round-trip, portfolio aggregation |
| `inventory.py` | dumps each tool's user-facing interface from LABELS, so specs can be authored without reading the arithmetic |

## 9. What to do next

1. **Owner decision on `bugs_open/225`.** These are consumer tax figures; a
   changed answer is a changed claim, so the number should not be changed until
   the owner has seen the finding (brief §9). The correction is cited to HMRC
   and the direction is unambiguous.
2. **Fix `bugs_open/224` by deletion, not by patching seven copies** — the
   shared `calculateAmortization` is already correct and already loaded by 11
   pages.
3. **Re-run `oracle.py` after each fix, with its controls in the same session.**
4. **Then, and only then, `--emit-criteria`** so the platform's own Tier-4
   runner keeps checking these answers unprompted.

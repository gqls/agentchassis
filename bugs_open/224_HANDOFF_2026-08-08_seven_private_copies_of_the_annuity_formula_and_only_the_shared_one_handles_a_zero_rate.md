# 224 — seven private copies of the annuity formula on one site, and only the SHARED copy handles a 0% rate: six calculators print `£NaN` or a stale answer

**Filed 2026-08-08 by the `loanandmortgagecalculator_couk` lane**, from the
owner-requested arithmetic-validation work
(`.../loanandmortgagecalculator_couk/HANDOFF_2026-08-08_arithmetic_validation.md`).
Diagnosis loop filed the same day: intake correlation
`fe69a7b8-d364-4e12-8039-f93f42a4170c`, run correlation
`3e18a949-8732-4603-b19b-f0c159860fa5` — see "Diagnosis-loop status" below.

Site: loanandmortgagecalculator.co.uk. Source: `sites` repo.

---

## The finding, in one line

`/assets/js/calculators.js` contains `calculateAmortization`, which has an
explicit `// Handle 0% interest edge case` branch. **Every `mortgages/*` tool
that computes a payment calls it. Every `loans/*` tool re-implements the formula
inline, and not one of the private copies has that branch.** The split is exact,
and it runs along a directory boundary rather than anything to do with the
tools' difficulty.

| implementation | 0% handled? | what a 0% rate produces |
|---|---|---|
| `assets/js/calculators.js` `calculateAmortization` | **yes**, `if (rate === 0)` | correct: `P/n` |
| `loans/standard-calc.html` `calculateLoan` | no — gated `r > 0` | **stale**: previous answer left on screen |
| `loans/settlement-calculator.html` `estSettle` | no — gated `apr > 0` | **stale** |
| `loans/car-finance-calculator.html` `calcCar` | no — gated `apr > 0` | **stale** |
| `loans/consolidation.html` `calcRisk` | no — gated `r > 0` | **£0.00 monthly payment** |
| `loans/compare-loans.html` `calc` | no — ungated | **`£NaN`**, and an inverted verdict |
| `loans/interest-rate-stress-test.html` `stressTest` | no — ungated | **`£NaN`** |
| `loans/overpayment-calculator.html` `calc` | no — ungated | **`£NaN`** + wrong months |

Five `mortgages/*` tools driven at 0% pass every vector. That is the control: it
is not that a zero rate is hard, it is that the shared function was written with
it in mind and six private copies were not.

**A 0% rate is not a synthetic input on a UK consumer-credit site.** 0% purchase
finance, 0% balance transfers, interest-free employer loans and manufacturer 0%
car finance are ordinary products, and `car-finance-calculator` is precisely
where a user would type 0.

## Two failure modes, and the silent one is worse

**Mode 1 — `£NaN`.** `(P*r*Math.pow(1+r,n)) / (Math.pow(1+r,n)-1)` is `0/0` at
`r = 0`. Ugly, obviously broken, harmless: nobody acts on `£NaN`.

**Mode 2 — a STALE answer, with no error and no blank.** Where the author added
`if (rate > 0) { …write the DOM… }`, a zero rate makes the function return
without touching the page, so the previous answer stays on screen looking
exactly like a fresh one. Measured live:

```
loans/standard-calc.html — 0% APR entered:
  the SAME final inputs give '£143.47' by one route and '£429.81' by another
  — the output is not a function of the inputs on screen
loans/car-finance-calculator.html — 0% APR entered:
  '£501.78' by one route and '£1222.56' by another
loans/settlement-calculator.html — 0% APR entered:
  '£5,158.11' by one route and '£5,023.84' by another
```

**That is the sharpest statement of the defect and it needs no reference to the
source at all**: type the same numbers into the same boxes, arrive by a
different path, get a different answer. It was produced by driving each vector
twice from two different priming vectors (`oracle.py`'s `determinism` check).

## The cases that are wrong, measured

| tool | vector | shows | correct |
|---|---|---|---|
| `standard-calc` | £10,000, 0%, 5y | £143.47 monthly, £7,216.40 interest | £166.67 monthly, £0 interest |
| `compare-loans` | A: £5,000, 0%, 3y | `£NaN`, **and the verdict names option B** | A is cheaper |
| `interest-rate-stress-test` | £10,000, 0%, 3y | `£NaN` | £277.78 now, £286.53 at +2% |
| `overpayment-calculator` | £15,000, 0%, 5y, +£50 | `£NaN` saved, **59 months saved** | £0 saved, 10 months saved |
| `settlement-calculator` | £5,000, 0% | £5,158.11 | £5,000 |
| `car-finance-calculator` | £30,000, 0%, 3y, £12,000 balloon | £536.08 | £500.00 |
| `consolidation` | new loan at 0%, £5,000, 5y | **£0.00 per month** | £83.33 per month |

**`compare-loans`' verdict is the one to look at twice.** `NaN < x` is `false`,
so the comparison falls to the `else` branch and the tool declares "Option B is
Cheaper" — meaning **a 0% loan entered in slot A is always declared the more
expensive option**. Driven in both slots to confirm it is the comparison and not
slot A's arithmetic: put the 0% loan in slot B and B correctly wins. It is a
confident, plausible, inverted recommendation on the site's own comparison tool.

`consolidation`'s `£0.00` is the same shape: an interest-free consolidation loan
is quoted as costing nothing per month.

## Why a boundary suite that tested 0% still missed one of these

Worth recording, because it nearly happened here. `consolidation` was first
driven with a **0% APR DEBT**, and it passed — the guarded branch returns 0, and
0 is the right answer for "interest remaining on a 0% debt". The defect only
shows when the **new consolidation loan** is at 0%, where returning 0 means a
£0.00 monthly payment. Testing the case where a broken guard's output coincides
with the correct one produces a green tick and no information — the no-op case
([[check-the-no-op-case-not-only-the-damage-case]] inverted).

## Fix candidates, ordered by what closes the door

1. **Delete the six private copies; call `calculateAmortization`.** The shared
   function is already correct, already loaded by 11 pages, and already carries
   the branch. This makes the defect class unrepresentable rather than fixing
   seven instances of it — there would be one implementation to be right about.
   The `loans/*` pages do not currently load `/assets/js/calculators.js` at all;
   adding the `<script>` tag is part of the change.
2. If (1) is too large for one pass, **fix the ungated three first** (`compare-loans`,
   `interest-rate-stress-test`, `overpayment-calculator`) — they print `£NaN`,
   which is visible, and `compare-loans` additionally inverts its verdict.
3. **Never fix a guarded one by widening the guard alone.** `if (r >= 0)` on
   `standard-calc` turns a stale answer into `£NaN`; the zero branch has to be
   written, not the condition relaxed.
4. **Regression-lock**: `oracle.py` carries 0%-rate vectors for all seven and
   `determinism` probes for the three stale ones.

## How to verify

```bash
cd docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
python3 oracle.py --tools standard-calc,compare-loans,stress,settlement,car-finance,overpayment-calculator,consolidation
```

Today: 23 FAILs across the seven. After a correct fix: 0.
Re-run the controls in the same session (`--mutate expectation`,
`--mutate crosstool`, `--selftest-parse`) or a green run is not evidence.

## Diagnosis-loop status — RAN, COMPLETED, and returned NO VERDICT. That is itself a finding

CLAUDE.md requires a `bugs_open/` file asserting a cross-cutting or structural
root cause to have been through `090` before it counts as filed. **It was filed**
— intake `fe69a7b8-d364-4e12-8039-f93f42a4170c`, claimed within 60s, run
correlation `3e18a949-8732-4603-b19b-f0c159860fa5` — and it produced **five
`bundle` artifacts and no verdict artifact, no `doc_notes` row, in ~9 minutes**,
while the orchestration reported `COMPLETED` and the work item reported
`status='complete'`.

**Why, measured rather than guessed** (2026-08-08):

```sql
SELECT DISTINCT repo FROM code_symbols;                        -- gqls/agentchassis  (one row)
SELECT DISTINCT substring(path from '\.[a-zA-Z0-9]+$') FROM code_symbols;   -- .go     (one row)
SELECT count(*) FROM code_symbols;                             -- 5755
SELECT count(*) FROM code_symbols WHERE path LIKE '%calculators.js%' OR path LIKE '%standard-calc%';   -- 0
SELECT count(*) FROM code_symbols WHERE path LIKE '%loanandmortgage%';                                 -- 0
```

**The diagnosis agent could not read a single file this symptom names.** They
live in the `sites` repo as `.html` and `.js`; the code index holds one repo and
one extension. What its five bundles actually fetched was `page_sections` rows —
the DB half of the symptom — so it looked, found the ported page records, and
never reached the JavaScript the claim is about.

⚠ **The shape to carry: a `090` run on a non-Go artefact terminates as a
SUCCESS.** `COMPLETED` orchestration, `complete` work item, artifacts present.
Nothing distinguishes "diagnosed and found nothing" from "structurally unable to
look", and the second is what happened here. Same root cause as
`bugs_open/223` (the landmine verifier narrating an unindexed footprint as
non-existent) — one Go-only index, two consumers, and in both the silence reads
as a finding. Recorded in `LANDMINES.md`.

**So this file rests on the norm's stated escape hatch, declared rather than
silently taken.** The substituted first-hand verification: every failure
reproduced in a real headless browser against the live site at named vectors;
all eight implementations read directly and quoted above; and the determinism
result — the same on-screen inputs giving two different answers — established
**without reading any source at all**, so it does not depend on my having read
the right file. The one durable claim that the loop would have been most useful
for is the structural one ("only the shared copy has the branch"), and that is
the `grep -l` over eight files reproduced in the table at the top; re-run it
before trusting it.

## Related

- `bugs_open/225` — SDLT on the same site, found by the same oracle, different
  mechanism (a stale legal rule, not a duplicated formula).
- Report, method, controls and the three things the harness got wrong first:
  `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/REPORT_2026-08-08_arithmetic_validation.md`

# 224 — seven private copies of the annuity formula on one site, and only the SHARED copy handles a 0% rate: six calculators print `£NaN` or a stale answer

## STATE 2026-08-10 — A SEVENTH INSTANCE, found by the unattended monitoring on its FIRST run. Mode 2 was NOT dead as a class

The 2026-08-09 block below says the stale-answer mode is "dead as a class". That
is true of the six tools converted to the shared engine and **false as a
statement about the site**, which is how it reads. Correcting it here rather
than editing it away.

**`mortgages/equity-release.html` had the same defect all along.**
`calcEquityRelease` ran `if(age < 55) { alert(...); return; }` — a bare return
that never touches the DOM. Measured live before any change:

```
age 65 (valid)                dispAge=65   erMaxCash=£124,000
age 32.5 after a valid calc   dispAge=65   erMaxCash=£124,000   <- STALE
```

Enter an eligible age, read £124,000; change the age to 32, still read £124,000
"for age 65". **The 0% sweep never found it because this guard is on AGE, not on
rate** — the detector and the fix pass were both scoped to rate guards, and this
is the same mechanism wearing a different trigger. Any `return` inside a
calculator's handler that does not write the DOM is a candidate; that is the
grep worth running, not `r > 0`.

FIXED and LIVE (sites `34239c7e4`): validates, then always writes — £0 against
the age actually entered, bands 55/65/75/85 re-verified unchanged.

**Then I swept the site for the shape rather than waiting for the weekly run to
surface them one at a time, and found THREE MORE** (sites `19543d40f`):

- `mortgages/bridging-loan.html` `calcBridge` — an unviable deal (interest +
  fees ≥ 100% of the loan) alerted and returned, leaving the previous
  **viable-looking gross loan** on screen. The user reads a facility size for a
  deal the tool has just decided cannot be done.
- `mortgages/investor.html` `calcLTV` and `calcYield` — clearing a field left
  the previous ratio standing beside inputs that no longer produced it.

All three now write on every path. Pre-flighted headless 6/6 on the transitions
that ARE the defect (viable → unviable, computed → cleared), with LTV 75.0% and
yield 5.76% unchanged. `investor`'s two buttons also got ids — the last unnamed
action buttons on the site.

**Monitoring is now 17 of 17** (2026-08-10): `loans-consolidation` joined the
unattended sweep via a tool-level component whose `function` equals `pages.name`,
so the fence installed for it still applies. Its row stays permanently locked and
the page stays `rebuild_policy='owned'`, which is what keeps an automated
rewriter off it — `check_tool_health` audits tool-level components and does NOT
read `no_auto_fix`. Run after the change: 4 passed / 0 failed.

**So the count is ten, not six**: six rate-guarded (the original), one
age-guarded, one viability-guarded, two blank-input-guarded. The common shape is
not "0%" and not "a rate" — it is **a guard that leaves a handler without
writing the DOM**. That is the grep.

**It was found by the unattended acceptance sweep the same night it was switched
on**, and the run's own note records that no automated rewriter was dispatched:
*"NOT auto-fixed — this fence declares no_auto_fix"*. 0 `improve_tool` items.

Second finding from the same failure: **the fence had pinned the page's INITIAL
MARKUP as an expected answer.** `--emit-criteria` records what the tool
displays, and for the two ineligible-age vectors the tool displayed nothing, so
the capture stored the untouched DOM. Re-emitted against the fixed tool; the
expectations are now real answers. **A captured expectation from a tool that did
not write is not an expectation** — the emitter gates on a wholly inert tool but
has no gate for a tool inert on ONE vector.

## STATE 2026-08-09 — FIXED, LIVE, VERIFIED BY THE FILING LANE'S OWN ORACLE. Fix candidate 1 executed: the private copies are GONE

Owner directed a session at this bug ("bugfix 224", 2026-08-08 evening) —
which is the owner seeing the finding, satisfying the arithmetic-validation
handoff's §9 rule for changed consumer-credit figures. The file stays in
`bugs_open/` per the owner ruling of 2026-08-06 (finished bugs stay put).

**What shipped** (sites repo `ea72609d6` + `Rerender: loans/consolidation.html`
`5b55a1ca4`; DB `page_components` synced by `gate_component_bytes.py --repair`,
6 rows, none lock-suppressed):

- The six verbatim `loans/*` pages now load `/assets/js/calculators.js` and
  call the shared engine: `calculateAmortization` (standard-calc,
  compare-loans, stress-test), `calculateOverpayment` (overpayment), and a NEW
  additive `calculateBalloonAmortization` (car-finance — PCP is the annuity on
  the balloon-discounted principal, so the 0% limit stays in exactly one
  place). The inline formula copies are deleted, not patched.
- `settlement-calculator` is not an annuity (linear 58-day estimate, correct
  at 0 by its own formula): guard became `apr >= 0` + always-write.
- `consolidation`'s locked `tool-1` row: both inline copies in `calcRisk`
  replaced with shared calls by deliberate operator SQL through the permanent
  lock (lock retained, provenance in `content_data._provenance.bugfix_224`);
  assemble-only rerender deployed it, served bytes == prediction.
- **Every submit now writes the DOM** (answer or cleared state) — mode 2, the
  stale answer, is dead as a class, not per-page.
- Non-zero-rate outputs preserved exactly: pre-flighted headless against the
  old formulas before deploy; live-vs-local matched byte-for-byte on default
  vectors; display conventions kept (standard-calc still bills-rounded — the
  oracle's CONV entries — but at 0% shows exact £0 interest).

**Verification (2026-08-09, live, controls in-session)** — the bug's own
"How to verify" plus the full control set:

```
oracle.py --tools <the seven>   →  PASS 77  FAIL 0  CONV 6  N/A 0   (was 23 FAIL)
--selftest-parse OK · --mutate expectation: 16 FAIL 0 passed ·
--mutate crosstool: 28 FAIL 0 passed · --mutate parse: 0 passed, 4 refused
full sweep: PASS 166 FAIL 4 — all four are mortgages/stamp-duty = bugs_open/225
golden: consolidation MATCHES (arithmetic exact); the six verbatim pages show
only the comparator's decomposed-shape content assertion, reproduced
identically on an UNTOUCHED verbatim control page (loan-vs-savings)
```

Determinism (mode 2's sharpest statement) is asserted by the oracle's
same-inputs-two-routes probes on the three formerly-stale tools — all pass.

Incidental hardening in the same session, in the lane's tooling:
`gate_component_bytes.py --repair` would have overwritten a decomposed page's
SECTION rows with the whole repo file (caught before running it; now skips
non-verbatim rows loudly), and `deploy_pages.py` died polling its own INSERT
(the psql command tag rode along in the returned id; fixed). Both in the lane
NOTES 2026-08-08/09 entries.

### STATE 2026-08-09 (evening) — loancalculator.co.uk FIXED IN THE DATABASE, six components, awaiting six queued rerenders

Taken by the bugfix-224 session on the owner's direction ("fix it in all the
calculators"); the loancalculator lane was told first, in
`loancalculator_couk/CONTRIB_2026-08-09b_bugfix224_session_taking_the_zero_rate_fix.md`.
Their two live threads are copy/voice and `bugs_open/227`; neither is
arithmetic, and no work item mentioned this defect.

**Six components fixed** (templates committed `767681e0d` BEFORE apply, per that
lane's own landmine), through their pipeline — `update_component --apply` →
`render_tool_row --apply` — so `page_components.rendered_html` now holds the
fix on all seven rows (`tool-loan-repayment` writes two: the homepage and the
retired `/tools/standard-calc.html`).

- `tool-loan-repayment` (HOMEPAGE widget), `tool-early-settlement` — stale-answer
  guards → zero branch / `apr >= 0`, plus always-write.
- `tool-compare-loan-offers` — zero branch, **and the verdict is now withheld
  unless BOTH offers compute**, which is what let `NaN < x` recommend the wrong
  loan.
- `tool-rate-stress-test`, `tool-overpayment-impact` — zero limit in the payment
  (also kills the "59 months saved": NaN exited the loop after one iteration).
- `tool-consolidation-risk` — the detector-blind one: a deterministic £0.00/month
  **with a "this is better" verdict**. Fixed, and a **blank** rate is kept
  distinct from a zero rate (mirroring the debt loop) — without that, this fix
  would have priced an unfilled form as interest-free.

**Verified before shipping, on the exact bytes that will ship**: the rewritten
rows were pulled from the DB and driven headless — 18/18
(`loancalculator_couk/rewrite/probe_zero_rate_rows.py`). **8 new
`defect_vectors.py` cases** now cover the 0% rate on all six; that lane had none
for these tools, which is precisely why a green `toolgolden 11/11` coexisted
with five broken pages.

**The new cases are not quiet tests**: `defect_vectors.py --both` scores every
one of the eight as **PROVEN** — it reads differently against the pre-fix
component — and every control as **unmoved**. Nothing VACUOUS. That matters
because the defect they cover survived a green `toolgolden 11/11` for months.

✅ **LIVE AND VERIFIED 2026-08-09.** All six assemble-only rerenders completed
and committed (`895e3bcff`, `1aaf7d1c4`, `94b4aa266`, `15a9ed066`, `ecc65cc1e`,
`07144b905`). On the SERVED pages:

| check | result |
|---|---|
| `zero_rate_sweep.py` live, the 5 detectable pages | **0 of 5 affected** (was 5 of 8) — 0 NaN, 0 history-dependent |
| `defect_vectors.py --live` | **all 16 pass against the served pages** |
| `check_site_serving.sh` | 26/26 (200 + ≥2000 B + DOCTYPE, so no B2 error blob was graded) |
| `--both` scoring | all 8 new cases **PROVEN**, controls unmoved, none VACUOUS |

Consolidation is covered by the `defect_vectors` cases, not by the sweep, which
is structurally blind to its deterministic zero — a clean sweep line for that
page would have meant *unmeasured*.

> **One harness fix on the way through, and the tell is worth keeping.** The two
> `tool-loan-repayment` cases failed `--live` with every element MISSING — the
> defect case AND its control failing identically, which is the signature of a
> harness pointed at the wrong page rather than a broken tool.
> `/tools/standard-calc.html` is retired and 404s, while the component ships on
> the HOMEPAGE. `defect_vectors.py` now takes a per-case `live_page` override.

## 2026-08-09 — THE SAME DEFECT IS LIVE ON loancalculator.co.uk (5 pages). NOT FIXED

> **CLAIMED — DO NOT START A SECOND FIX. Corrected 2026-08-09, ~1 hour after
> this section was written.** The **bugfix-224 session already owns this work**
> and published its claim before I swept:
> `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/CONTRIB_2026-08-09_*`
> (commit `0e4e810f4`, "the bugfix-224 session is taking the 0% fix on your
> site"), with five `rewrite/tool-*.html.tmpl` edits live in the working tree at
> the time of writing. **My closing line below — "awaiting a scope decision
> before touching another lane's live site" — is therefore withdrawn**: the
> decision was not mine to take and had already been taken.
>
> **How I nearly duplicated it, which is the reusable part.** `who-owns.py`
> pointed at MY lane, because it reads COMMITS and the other session's work was
> uncommitted. The doc-level check I ran (the sibling lane's own handoff) said
> its threads were copy/voice and `bugs_open/227` — true, and irrelevant, because
> the claiming session is not in that lane. **The claim was in a file I had not
> thought to look at, in a lane I had already cleared.** What would have caught
> it in one command is a `git log --since` over BOTH the target lane and the
> whole tree, plus `git status` for uncommitted work — a session mid-fix is
> invisible to every ownership check that reads history.
>
> **Their measurement is better than mine and supersedes it on one row.** They
> read the six components' `rendered_html` directly and found a **sixth**
> affected component, `tool-consolidation-risk`, which my detector reported
> clean — confirming the blind spot I flagged below was real and load-bearing,
> not a hedge. Worse than I guessed, too: `newMonthly` initialises to `0`, and
> the verdict branch tests `totalBal > 0 && newN > 0` but **not** `newR`, so a 0%
> consolidation is presented as £0.00/month **with a "this will save you"
> verdict attached** — on the page whose purpose is to warn about term
> extension. Read their contrib note, not this section, for the component list.
>
> What stands from this section: the detector, its two controls, and its two
> stated blind spots. Live re-measure at the time of writing this correction
> still showed 5 of 6 affected, i.e. their fix had not yet deployed.

Found by sweeping the sibling sites after the owner asked for the 0% bug fixed
"in all the calculators". Measured, not grepped, with a new general detector
(`zero_rate_sweep.py`, same lane) that needs no per-tool oracle: it drives every
rate-labelled field to 0 and reports (a) any output containing `NaN`/`Infinity`
and (b) any output that differs when the SAME final inputs are reached by a
different route.

| site | result |
|---|---|
| loanandmortgagecalculator.co.uk | **clean** — 0 of 6 (this bug's fix, verified) |
| **loancalculator.co.uk** | **5 of 8 pages AFFECTED** |
| mortgagecalculator.co.uk | clean — it uses a shared engine; see the false-positive note below |

Affected on **loancalculator.co.uk**, live 2026-08-09:

```
index.html                            HIST  #monthly-display  '£202.29' vs '£251.22'
                                      HIST  #total-interest   '£2,137.40' vs '£5,073.20'
tools/compare-loans.html              NaN   #res-m-a #res-i-a #res-m-b #res-i-b  (£NaN)
tools/interest-rate-stress-test.html  NaN   #curr-pay (£NaN)
tools/overpayment-calculator.html     NaN   #save-display (£NaN)
tools/settlement-calculator.html      HIST  #settle-result '£5,078.66' vs '£5,139.04'
```

Note the first line: **the site's HOMEPAGE carries the standard-calc widget**,
so the stale-answer mode is on the front door.

**`tools/consolidation.html` reads clean and probably is not** — the detector is
blind to this site's third failure mode (a guard that deterministically returns
£0.00 is neither `NaN` nor history-dependent), which is exactly how the same
page nearly escaped here. Treat it as unmeasured, not as passing.

⚠ **The fix is NOT the same shape as the one above.** Two differences:
1. **loancalculator.co.uk has no `assets/js/calculators.js`** — there is no
   shared engine to delete the private copies in favour of. One would have to be
   created, or the copies fixed in place.
2. **All five pages are DECOMPOSED**: the calculator lives in a `tool-N`
   `page_components` row and the DB is the render source, so editing the repo
   file changes nothing and would be overwritten by the next rerender
   (RUNBOOK §12). The route is operator SQL through the lock plus an
   assemble-only rerender, as used for `consolidation` here.

```
/index.html                            prose-0,prose-1,prose-2,prose-4,tool-3
/tools/compare-loans.html              prose-0,prose-1,tool-2
/tools/consolidation.html              prose-0,prose-1,prose-3,tool-2
/tools/interest-rate-stress-test.html  prose-0,prose-1,tool-2
/tools/overpayment-calculator.html     prose-0,prose-1,prose-2,prose-4,tool-3
/tools/settlement-calculator.html      prose-0,prose-1,prose-3,tool-2
```

**Why the owning lane has not seen it**: its handoff records `toolgolden 11/11
exact` — the consistency check that, by this bug's whole argument, cannot refute
a defect present when the golden was taken. That is the finding restated on a
second site, and it is the reason not to read that 11/11 as reassurance.

**Detector's known blind spots, so a clean sweep is not over-read:**
- deterministic-zero mode (see consolidation above) is invisible to it;
- accumulator tools produce FALSE POSITIVES — `mortgagecalculator.co.uk`'s two
  portfolio pages flagged as history-dependent because route B presses the
  first non-reset button ("Add property") twice, adding a second row. Verified
  as harness artefact, not a defect: `#portfolioCount` read '1 property' vs
  '2 properties'.
- Controls run 2026-08-09: the pre-fix pages of THIS site, served from git at
  `2bbe50582`, flagged 6 of 7 (consolidation being the known blind spot) and
  cleared the known-good `mortgages/simple.html` — so the detector fires and can
  also stay silent.

Original handoff below, unchanged.

---

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

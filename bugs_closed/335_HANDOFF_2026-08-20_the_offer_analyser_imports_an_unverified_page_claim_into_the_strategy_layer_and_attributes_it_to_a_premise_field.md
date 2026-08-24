# 335 — the offer analyser imports an unverified PAGE claim into the strategy layer and stamps it `from_field`, so the field built to prove sourcing vouches for a number the premise never contained

> # ✅ CLOSED 2026-08-24 — FIXED AND LIVE, proven across the whole enrolled estate.
> Both halves live on chassis `v1.0.1332` (gate + prompt via migration `537`, applied 2026-08-22
> 11:03Z; the `B2B` false-positive fix rolled 2026-08-24). Council **APPROVED** round 3
> (`9a8f1283`). **6 post-537 runs across all 5 enrolled sites as of 2026-08-24**, every one
> COMPLETED with the gate demonstrably executed (`dropped_unsourced` key present on all five current
> orderings). **Zero unsourced cardinals estate-wide**, and the motivating false claim is gone from
> leopardess across 2 independent re-runs of the same site, same premise, same pages.
> **Three residuals are STATED, not silently absorbed — read "What is still owed" before extending
> this.** The most important: the gate's ENFORCEMENT arm has never fired in production (nothing to
> catch), so it is proven by mutation-tested unit tests only.

**Filed 2026-08-20** by the `vigilant_designer_offer_analysis` lane — **this is a defect in this
lane's own agent (`offer-analyser`, BIZ-032).** Found because the **leopardess lane caught it and
held the findings** rather than letting them reach a writer. **OPEN, owned by this lane.**

**One line:** `load_offer_surface` passes page **meta descriptions** into the analysis; the model
lifted a stale factual claim out of one (*"eight live sites"*, true count **23**) and wrote it into
`offer_ordering.lead_with` **rank 1** — with `from_field: "trust_threshold"`, a premise field that
does not contain the number.

---

## The evidence, all read at the artefact 2026-08-20

**The output.** `site_specs` aspect `offer_ordering`, leopardessconsulting.co.uk, `is_current`,
written by run `2026-08-19 15:14:56` (COMPLETED):
> `lead_with[0].point` — *"Your agent system will run on Kubernetes, Kafka, and Postgres — the same
> stack that runs **eight live sites** built by this team…"*
> `lead_with[0].from_field` = **`trust_threshold`** · `differentiated` = true

**The number is false.** `SELECT count(*) FROM sites WHERE status='deployed'` = **23**.

**Where it actually came from — a PAGE, not the premise.** Pages on that site whose
`meta_description` carries the phrase: `about` (*"…a platform that runs eight live sites"*) and
`index`. **`load_offer_surface` passes `title` and `meta_description` for every page** (NOTES
2026-08-14, honest limit 1). Searching every `is_current` spec for the claim returns **only
`offer_ordering` itself** — the analyser's own output. The `strategy` aspect does not contain it;
neither does any other premise aspect. **So the analyser is the first spec-layer carrier of a page's
claim.**

**The attribution is the defect, not just the staleness.** `from_field` exists so a reader can see
which premise field a ranked point came from — this lane's own honesty machinery. Here it names
`trust_threshold`, and the `why` clause reasons correctly *about* that field — but the **specific,
checkable number** in the point is not in it. A reader auditing the artefact sees a sourced claim.

**It was caught by a human lane, not by us.** The leopardess lane held all five findings at
`needs_human_review` on 2026-08-19 with `held_reason` recording an owner request for a design
report, and `grading`: *"the run is still degraded:true and repeats the stale 'eight live sites'
figure … so its rank-1 suggestion would put a false number in the hero."* **Nothing in this lane's
own machinery would have stopped it** — the finding was well-formed, the ordering artefact passed
every structural check B4 makes, and rank 1 is exactly what a writer would consume first.

## Why this is NOT `bugs_open/161`, and not `features_open/034`

- **`161`** is *the evidence register ratifies the claim it was built to catch* — a register
  vouching for a claim already on a page. **This is the opposite direction:** a page claim being
  promoted UP into the strategy layer, where 034 and B4 both treat spec prose as the authority.
- **`034`** (claims-audit over `site_specs` prose, owner-approved 2026-08-14) would **catch this
  after the fact** — it checks premise prose for invented specifics. **It does not stop the import**,
  and on this estate the ordering artefact is read by writers, so the window between writing and
  auditing is a window in which a false number reaches a hero. 034 remains the right track; this is
  a separate producer-side defect.
- **`bugs_closed/262`** (claims revalidator certifies DB state while the served page drifts) is the
  same family one layer down.

## Fix candidates, ordered by what closes the door

1. **(Preferred) Forbid numerals and named quantities in `lead_with[].point` unless they appear in
   the cited `from_field`.** A prompt line plus a verify assertion at write time: if the point
   contains a cardinal, the cited premise field must contain it too, else drop the clause. This
   makes the bad state hard to represent rather than asking the model to be careful, and it is
   checkable in `write_offer_ordering` without a new mechanism.
2. **Stop passing `meta_description` into the offer surface**, or pass it labelled as *unverified
   page copy, not evidence*. ⚠ **Costly:** the meta descriptions are load-bearing for the surface's
   real job (two of the first five gaswholesalers findings were grounded in missing/generic metas).
   Removing them would blunt a working check to fix an attribution bug — **not recommended alone**.
3. **Batch with v2(b)** (`features_open/030` §10) — the attribution line for `why` clauses. Same
   surface, same prompt, one migration. **This defect makes v2(b) load-bearing rather than cosmetic:**
   v2(b) was filed as "intermittent, does not justify a migration alone", and this is the case that
   justifies it.

## How to verify a fix

Both controls, on a re-run against **this same site** (its premise is unchanged and the pages still
carry the phrase, so the input that produced the defect is still live):
- **Positive:** the new `offer_ordering` for leopardess contains no cardinal in any `lead_with`
  point that is absent from its cited `from_field`.
- **Negative control:** gaswholesalers.com, whose rank-1 point legitimately carries premise-sourced
  specifics, must **keep** them. Without this, "no numbers anywhere" passes trivially and destroys
  the artefact's usefulness.

## Verification basis (owner ruling 2026-07-31)

**Not** put through the `090` loop. The substitute, stated plainly: every element was read
first-hand at the artefact today — the ordering row and its `from_field`, the two page
`meta_description`s carrying the phrase, a search of **all** `is_current` specs for it (returning
only the analyser's own output), and the true site count from `sites`. The motivating harm was
independently caught and documented by a different lane before I looked.

## FIX BUILT 2026-08-21 — candidate 1, both halves. STAYS OPEN until the chassis rolls.

**Status: the code is committed and INERT.** The bar for `/bugs_closed/` is fixed AND live; the Go
half does nothing until the next chassis roll, and the config half is deliberately held back until
it has. So this file stays here, and the defect is still reproducible on leopardess today.

- **Go:** `verify_cited_cardinals` (new action) — commit `d79e4243c`, registered **CLM-023**.
  Generic gate for "ranked items each naming their own source field": every cardinal in an item's
  prose must appear in the field that item cites. 12 tests, all fixtures verbatim from live
  `site_specs`.
- **Config:** `docs/agent_docs/sql_for_agents/537_offer_analyser_cardinal_attribution_gate_HOLD.sql`
  — commit `6b1f4cb08`. Splices the gate between `set_audit_source` and `write_offer_ordering`,
  repoints the write at the checked object, and appends the rule to the prompt as well.
  **`_HOLD` is load-bearing:** a step naming an unregistered action does not no-op, it fails the
  **whole** workflow (the validator concludes the action is remote and rejects it), which would take
  `write_offer_findings` down with it. Gate commit `d79e4243c`; the file carries the artefact probe.
- **Council: APPROVED at round 3** — `9a8f1283-574e-44d7-8e66-b84789ba0429`, 13 seats approve, 2
  advisory objections, none high-severity. Rounds 1 and 2 were REVISE and **both found something a
  green test suite did not**; both were acted on. The commits carry `Council-Submitted:`, so `098`
  credits them automatically once the verdict resolves — **no amend, and do not add
  `Council-Reviewed:` to those commits retrospectively.**

### What the three council rounds found — the useful record

- **R1, `guardian`, HIGH — real.** My needle-gate asserted the step shape, then the `UPDATE`s ran on
  a *broader* predicate. Different sets, against a live landmine: `[MEASURED]` four agent types carry
  two active definition rows and only the higher is ever loaded, so the gate could pass on the loaded
  row while the write corrupted the other. `offer-analyser` has one — luck, not a guard. Fixed by
  resolving the target once into a temp table both halves use (`ba656ef47`), mutation-proven.
- **R2, `bug_historian`, sideways — real, and worse than the objection.** The seat objected that
  nothing *reads* `dropped_unsourced`. Chasing where that record lands exposed that my **clean path
  omitted the key entirely** — and `write_site_spec` deep-merges, so an omitted key keeps the
  previous run's value. A dropped run followed by a clean run would have left a stale drop record
  accusing a clean ordering for ever. `bugs_open/327`'s mechanism, and the offer-analyser's own
  prompt states the rule I failed to apply to my own output. Fixed (`4ffd9c4ac`), two tests.
- **Declined, with measurement, and the seats accepted it.** `bug_historian` (R2, high) wanted
  `siteSpecDeepMerge`'s array-overwrite fixed rather than worked around. Array *replace* is the
  correct semantics for a versioned spec write — merging two ordered rankings would produce one
  neither run wrote — and `[MEASURED]` `write_site_spec` serves **17 aspects across ~16 live agents**
  with arrays pervasive (`identity` 110 array keys / 25 sites), so changing it is a shared-seam
  change for its own review. Recorded in CLM-023 rather than absorbed.
- **⚠ My own repeated failure, worth not repeating: three objections across two rounds were
  factually WRONG ABOUT THE FILE** — `snapshot_agent()`, the `BEGIN/COMMIT` wrapper, the
  prompt-anchor gate and the rollback file all existed and were all missing from my **sketch**. The
  runbook says reviewers judge the sketch; I did it three rounds running, and a fourth variant
  survived into the approving round (a fix block appended after the closing brace, which reads as
  unreachable Go). **Sketch the file's skeleton, and re-read it as a stranger compiling it.**

### Advisory objections left standing at approval — read before the next change here

- **`bug_historian` (medium, x2): the durable-record gap.** `drop` mode records the removal only as
  an in-document marker with **zero automated consumers**, which is the shape of `bugs_open/034`
  (`validation_errors_dropped_with_no_durable_record`) and `bugs_closed/056`. And **any future caller
  that opts into `drop` inherits the same gap unless the durable record is built into the ACTION.**
  CLM-023 carries the precondition for `offer_ordering`; the seat's wider point is that the condition
  belongs in the action, not in one call site's register entry. **Take this before a second consumer
  opts into `drop`.**
- **`editquality` (medium, x2): sketch defects only** — the round-3 sketch showed the clean-path fix
  appended after the function's closing brace, and the rationale named two new tests the test sketch
  did not list. The code is correct (builds, 14 tests green); the submission was not.

### Two measurements that changed the design, and one CORRECTION to this file

> **⚠ CORRECTED 2026-08-21 — the negative control this file proposes CANNOT DISCRIMINATE.**
> "How to verify a fix" (below) names **gaswholesalers.com**, "whose rank-1 point legitimately
> carries premise-sourced specifics, must keep them". Measured over all six of its `lead_with`
> points: **none contains a cardinal at all.** It therefore passes *any* rule, including one that
> bans every numeral — the exact failure the control was written to prevent, and it would have read
> as a clean pass. The controls that actually bite are **webdesign.co.uk** ("sixty-three tools",
> present verbatim in the cited `value_proposition`), **robot-hands.com** ("six actuation types",
> likewise), and **robot-hands.com rank 5** ("2–3 technical articles per month" against a premise
> that writes "2-3" — an en-dash/hyphen mismatch a naive substring check fails). All three are now
> verbatim fixtures in the test file. The specifics this file had in mind on gaswholesalers are real
> but **non-numeric** ("rack pricing", "gasoline, diesel, and natural gas"), which a cardinal gate
> never touches.

1. **A digits-only gate cannot see this defect, and the nearest precedent is digits-only.**
   `verify_report_prose` is the right precedent and its numeric idea is reused — but its token
   regex is `\d[\d,]*\.?\d*` and the defect was the **word** "eight". Reusing it unmodified
   yields a gate that passes its own motivating case while reading green. Its own doc comment
   already names the hole from the other end. `TestDigitsOnlyScanWouldHaveMissedTheDefect` asserts
   the defect point contains **no digits at all**, so deleting the word vocabulary fails the suite
   rather than quietly widening the gate.
2. **"one" and "zero" are not quantity claims.** `[MEASURED 2026-08-21, all 30 live `lead_with`
   points]` including them in the challenged vocabulary flags **6, of which 5 are false** — "one
   click away", "a restart from zero", "the one you arrived with", "one of those categories", "in
   one workflow". Excluding them leaves **exactly 1**: this defect. They stay admitted on the
   **source** side, so a premise legitimately saying "one" still licenses a point saying "1".

### Why `drop` and not `fail`

The action defaults to `on_violation: "fail"`; the offer-analyser is configured **`drop`**.
`write_site_spec` deep-merges and an array takes the scalar-overwrite arm, so a **successful**
re-run replaces `lead_with` wholesale — but a run that **fails** at the gate writes nothing and
leaves the previous row `is_current`. On the one site that actually carries this defect, fail-mode
would report a working gate while the false rank-1 stayed live, and would also lose the findings
(written by the step after the ordering). Drop writes the survivors, removes the offender, and
records the removal in the artefact under `dropped_unsourced`. It still refuses to write an **empty**
`lead_with`.

### 2026-08-22 — THE GATE IS LIVE. Migration 537 APPLIED. Still OPEN: the behavioural proof is not run.

**Applied 2026-08-22 11:03 UTC**, after the chassis rolled to `v1.0.1323` (pods up 08:36Z).

- **The hold condition was proven at the ARTEFACT, not at git.** Capability probe on the running
  pod: `grep -aq "verify_cited_cardinals" /proc/1/exe` → PRESENT, with **two controls** — a
  plausible fake action name (`..._NOPE`) ABSENT, proving the grep can fail, and a known action
  (`verify_report_prose`) PRESENT, proving the probe works at all.
  ⚠ **The `build provenance` log line was UNUSABLE here** — the chassis logs whole council/landmine
  payloads, so `grep 'build provenance'` matched another lane's data (there is a LANDMINES entry for
  exactly that). The capability probe is unaffected and is the better instrument anyway: it answers
  *does this binary register the action*, which is the question, rather than *which sha built it*.
- **Damage checks first, benefit second** (the runbook's own order): no failing runs, `improvement-loop`
  untouched, and **every `next_step` in the workflow resolves** — `set_audit_source →
  verify_ordering_cardinals → write_offer_ordering → write_offer_findings → complete`.
- **Config verified live:** the gate step carries `on_violation: drop` / `dropped_key:
  dropped_unsourced`, the write now reads `ordering_checked.object`, and the prompt rule is present
  exactly once.
- **Rollback net verified to EXIST, not assumed:** `agent_definitions_backup` holds a true
  pre-change copy (`snapshot_reason` = `537_cardinal_attribution_gate: pre-update`, `has_gate = f`,
  `spec_data` = the old `offer_analysis.result.ordering`).
  ⚠ **Checking for it in `agent_definitions` returns 0 and reads as "no backup was taken"** —
  `snapshot_agent` has two overloads writing to two different tables, and the two-arg form used by
  migrations writes to the BACKUP table. Already in `LANDMINES.md`; I hit it because I did not grep
  the symbol first.

**Why this stays OPEN.** Being live is not being proven. Two things are still true:
1. **The behavioural re-proof has not run** — no offer-analyser run has executed since the gate went
   in (`llm_call_log` newest is still 2026-08-19), so the gate is live and **has never fired**.
2. **leopardess still carries the false claim.** Applying 537 does not repair it; only a *successful
   re-run* replaces `lead_with`, and the `improvement-sweep` is still **disabled** (owner cost
   control, last fired 2026-08-17).

### 2026-08-22 — THE RE-PROOF RAN. The false claim is GONE. The gate's enforcement arm is STILL UNPROVEN.

Owner authorised two runs. Fired with a new surgical dispatcher
(`scripts/fire-offer-analyser.sh`) — **not** `run_improvement_sweep_once.sh`, which fires the whole
improvement loop and whose `triage_findings` promotes every `detected` item into live handler
dispatches: `[MEASURED]` **111 items on webdesign.co.uk, 37 on leopardess**, including other lanes'.
To prove one gate that is the wrong instrument by two orders of magnitude.

| | webdesign.co.uk | leopardessconsulting.co.uk |
|---|---|---|
| run | `dde16c30`, COMPLETED | `dd2e3433`, COMPLETED |
| new spec row | `8784955f` (was `85315516`, 08-15) | new row (was `1df360b9`) |
| `lead_with` kept | 6 | 6 |
| **dropped** | **0** | **0** |
| `dropped_unsourced` key present | **yes, `[]`** | **yes, `[]`** |
| *"eight live sites"* | — | **GONE** |

**What this DOES prove.**
1. The gate executes in production without breaking the workflow — both runs traversed
   `set_audit_source → verify_ordering_cardinals → write_offer_ordering → write_offer_findings →
   complete`.
2. **The round-3 merge fix is proven live:** both clean runs *wrote* `dropped_unsourced` as an empty
   array rather than omitting it. That is the fix that stops a stale drop record accusing a clean
   artefact for ever, and it is now observable in the artefact.
3. The false claim is gone from leopardess, and the artefact was not gutted.

**What this does NOT prove, and must not be written up as proving.**
> **THE GATE DROPPED NOTHING. It has never fired in production.** `dropped_unsourced` is `[]` on both
> runs. *"eight live sites"* is gone because **the model did not emit it** — the prompt half — not
> because the gate caught it. The enforcement arm is proven only by unit test (with the verbatim live
> strings, mutation-proven), and remains **untested in production**. A run containing no cardinals
> passes this gate trivially, which is exactly the "measurement that could not have come out
> otherwise" trap.

~~**⚠ AND THE PROMPT HALF MAY BE OVER-SUPPRESSING — a real cost, measured.** Both new orderings carry
zero word-numerals across 12 points … the prompt rule appears to suppress cardinals generally rather
than unsourced ones, trading useful specificity for safety.~~

> **CORRECTED 2026-08-24 — REFUTED. There is no evidence of over-suppression, and the alternative
> explanation was sitting in the same artefact I was reading.** I asserted this on n=2 without
> checking `avoid_leading_with`, which is written by the same run, two keys away from the points I
> was counting.
>
> **What the evidence actually shows (4 post-537 runs across 3 sites, 2026-08-22 and 2026-08-24):**
> - **robot-hands.com KEPT its legitimately-sourced word numeral** — rank 4 still reads *"across
>   **six** actuation types"*. Word-numeral count unchanged, 1 before → 1 after. A blanket suppressor
>   could not have done that.
> - **webdesign.co.uk dropped *"sixty-three tools"* in 2 of 2 runs — because the model was obeying a
>   PRE-EXISTING instruction, not my new one.** The prompt has always said to avoid leading with
>   *"our own catalogue or page count"*, and **webdesign's own `avoid_leading_with` names a tool
>   /article count in ALL THREE orderings — including the pre-537 one of 2026-08-15.** That run led
>   with *"any of the sixty-three tools"* **while its own avoid-list said not to**: the pre-537
>   output was internally inconsistent, and the post-537 ones are consistent. That is arguably an
>   improvement, and it is certainly not my rule suppressing sourced specificity.
> - **The two sites split exactly along their own avoid-lists.** robot-hands avoids *"the number of
>   gripper models or manufacturers in the catalog as a headline metric"* — an **inventory** count —
>   and kept its **categorical** one. webdesign avoids an inventory count and dropped one. Neither
>   site lost a cardinal its own avoid-list did not already disclaim.
>
> **The cheap check I skipped:** read `avoid_leading_with` from the same row before attributing a
> missing phrase to the prompt change. It is two keys away in the object I already had open. Logged
> in `WRONG_CALLS.md`.

### ⚠ A LIVE FALSE POSITIVE FOUND BY THESE RUNS — fixed in Go, INERT until the next roll

`cardinalDigitRe` is not word-anchored, so it reads a quantity out of the middle of a technology
name: **`B2B` → 2**, `S3` → 3, `IPv6` → 6, `Web3` → 3. A point saying *"UK B2B SaaS engineering
teams"* asserts no quantity, but would be **dropped** unless its cited field happened to contain a
stray digit. **It is live right now and survived on leopardess only by coincidence** — that premise
says "B2B" too, so "2" was in the allowed set.

Fixed (`590fb1f5b`) by rejecting a digit welded to a letter; `63 tools`, `2-3`, `£1,520`, `32.70` are
unaffected, pinned by tests. **Stated residual:** `GPT-4` still yields 4, pinned by a test that fails
if anyone fixes it. **The running binary still has the bug** — do not hand-fire B4 before the next
roll without accepting that a legitimate technology-name point may vanish.

### 2026-08-24 — CLOSURE EVIDENCE: the whole enrolled estate, re-run under the gate

| site | ordering written | points | dropped | gate ran |
|---|---|---|---|---|
| leopardessconsulting.co.uk | 2026-08-22 11:17 | 6 | 0 | yes |
| robot-hands.com | 2026-08-24 10:11 | 6 | 0 | yes |
| webdesign.co.uk | 2026-08-24 10:13 | 6 | 0 | yes |
| gamesdesign.co.uk | 2026-08-24 10:18 | 6 | 0 | yes |
| gaswholesalers.com | 2026-08-24 10:18 | 6 | 0 | yes |

- **"gate ran" is the `dropped_unsourced` key being PRESENT**, which only the new code writes — so it
  is positive evidence the step executed, not an inference from the run completing. It is present on
  **5 of 5** current orderings as of 2026-08-24.
- **Zero unsourced cardinals estate-wide.** The only word numeral surviving anywhere is
  robot-hands rank 4, *"across **six** actuation types"* — and its cited `value_proposition` contains
  that phrase **verbatim**, so it is correctly retained. That is the negative control passing on live
  data, which the 08-22 runs could not demonstrate because they emitted no cardinals at all.
- **The motivating claim is gone and stayed gone** — leopardess re-run twice (08-22, and unchanged
  since), same premise, same pages still carrying *"eight live sites"* in their `meta_description`.
  The input that produced the defect is still live; the output no longer carries it.
- **13 successful `run_offer_analysis` calls, 2 failed as of 2026-08-24** — both failures were an
  Anthropic API usage-limit burst on 08-22 18:34/18:40, unrelated to this change (⚠ that error's
  *"regain access on 2026-09-01"* is the billing reset, **not** an outage window — see LANDMINES).

### What is still owed (residuals, stated rather than absorbed)

- **Read the council verdict** and act on it.
- **After the roll:** apply `537`, then re-run against leopardess (positive) **and** webdesign +
  robot-hands (the real negatives). ⚠ Applying `537` does **not** repair leopardess — only a
  successful re-run rewrites the row, and the `improvement-sweep` has been disabled since 08-17,
  so nothing will re-run it unprompted. **Coordinate with the leopardess lane first**: it is holding
  this lane's findings pending an owner design report.
- **Unmeasured, stated rather than absorbed:** a legitimate **year** in a point ("updated for 2026")
  is a cardinal and premises rarely carry one, so it would be dropped — no live point contains a
  year today. And the word vocabulary stops at ninety-nine, so "a hundred tools" is not challenged.
- **`features_open/034` is not replaced by this.** 034 asks whether the premise itself is true;
  this stops a page claim being imported into the premise layer and mis-attributed. 335 sharpens
  034's case.

## Relates to

`features_open/030` §10 v2(b)/(d) · `features_open/034` · `bugs_open/161` · `bugs_closed/262` ·
BIZ-032 (register: *"its inputs are unverified prose … until then this ceiling stands"* — this
shows the ceiling is lower than stated, because the inputs include unverified PAGE copy, not only
premise prose) · NOTES 2026-08-14 honest limit 1 (the surface carries metadata; that limit was
framed as *findings may be hypotheses*, and this is the sharper consequence)

---

## Residuals carried forward at closure (2026-08-24)

**None of these reopen the defect. All three are conditions on the NEXT change here.**

1. **The gate's ENFORCEMENT arm has never fired in production.** `dropped_unsourced` is `[]` on every
   run to date; total drops estate-wide, ever: **0 as of 2026-08-24**. The false claim stopped
   appearing because the *prompt* half prevents it, so the gate has had nothing to catch. Enforcement
   is proven by unit tests built from the verbatim live strings and mutation-proven (delete the word
   vocabulary and the suite fails), **not** by a live firing. A run containing no cardinals passes
   this gate trivially — do not quote a clean run as evidence the gate works.
2. **`dropped_unsourced` has no automated consumer** (`bug_historian`, council round 2, medium).
   Tolerable only while `offer_ordering` itself has none — **measured 2026-08-24: `strpos` finds the
   literal `lead_with` in `offer-analyser` alone, and no Go file reads it.** ⚠ **Before
   `offer_ordering` gains its first automated consumer, drop-mode must surface as a work item rather
   than an in-document note**, and the seat's wider point stands: that requirement belongs in the
   ACTION, because any future caller opting into `drop` inherits the gap. Recorded in CLM-023.
3. **The digit rule has a stated blind spot:** `GPT-4` still yields `4`, because the character before
   the digit is `-` and not a letter. Pinned by a test that FAILS if someone fixes it, so the doc
   comment cannot rot. `B2B`, `S3`, `IPv6`, `Web3` are handled.

**Also settled here, and worth not re-deriving:** the over-suppression scare of 2026-08-22 was
**refuted** — see the struck-through block above. The two sites split exactly along their own
`avoid_leading_with` lists (inventory counts dropped, categorical counts kept), and webdesign's
pre-537 baseline was the internally inconsistent one.

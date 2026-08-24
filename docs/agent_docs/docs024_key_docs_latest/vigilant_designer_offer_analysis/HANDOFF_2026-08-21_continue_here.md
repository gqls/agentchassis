# HANDOFF — vigilant designer + offer analyser (2026-08-21)

**COLD-START = this file + `bugs_open/335` (fix built today, still OPEN) + `features_open/030` §10
(the v2 backlog) + `features_open/034`.**
**This supersedes `HANDOFF_2026-08-20_continue_here.md`.**

> **Re-run every liveness claim here before acting.** ~570 commits/day land on this branch. In one
> session today the working tree would not compile (another lane mid-edit), and a baseline archived
> minutes apart came off two different HEADs. Verify against `git archive <resolved-sha>`, never the
> tree and never the moving name `HEAD`.

## The one-line state

> # ✅ 2026-08-24 — `335` IS CLOSED (`bugs_closed/335`). FIXED AND LIVE, proven estate-wide.
> Chassis `v1.0.1332` carries both the gate and the `B2B` false-positive fix (capability-probed with
> two controls). **6 post-537 runs across all 5 enrolled sites as of 2026-08-24**, gate executed on
> every one, **zero unsourced cardinals estate-wide**, false claim gone from leopardess and stayed
> gone. Negative control finally passed on live data — robot-hands KEPT *"six actuation types"*.
>
> **THE OVER-SUPPRESSION SCARE IS REFUTED** (see `WRONG_CALLS.md` 2026-08-24). The two sites split
> along their own `avoid_leading_with` lists; webdesign's pre-537 baseline was the inconsistent one.
>
> **THREE RESIDUALS, all conditions on the next change, none reopening the defect** — full text in
> `bugs_closed/335`: (1) the gate's **enforcement arm has never fired** (0 drops ever), so it is
> unit-proven only — never quote a clean run as evidence it works; (2) `dropped_unsourced` has **no
> automated consumer**, so drop-mode must surface as a work item *in the ACTION* before a second
> caller opts in; (3) `GPT-4` still yields `4`, pinned by a test that fails if fixed.
>
> **NEXT WORK for this lane is the v2 batch** — `features_open/030` §10, plus **v2(d)**, which ⚠ is
> NOT in §10 (it lives in `NOTES` ~line 2214 and the 08-17/08-18 handoffs).


> **UPDATE 2026-08-22 ~11:20Z — RE-PROOF RUN. The false claim is GONE; the gate itself has NEVER
> FIRED.** Both authorised runs COMPLETED (`dde16c30` webdesign, `dd2e3433` leopardess), both kept 6
> points and dropped **0**. *"eight live sites"* is gone from leopardess — but because the **prompt**
> half stopped the model emitting it, not because the gate caught it. **Proven live:** the gate runs
> without breaking the workflow, and the round-3 merge fix (a clean run WRITES `dropped_unsourced` as
> `[]`). **Not proven:** enforcement, which is unit-tested only.
> **⚠ TWO THINGS TO CARRY FORWARD.** (1) A **live false positive** — `B2B` is read as the quantity 2
> (`S3`→3, `IPv6`→6). Fixed in Go (`590fb1f5b`) but **the running binary still has it**, so a
> hand-fired B4 can silently drop a legitimate technology-name point until the next roll.
> (2) ~~The prompt may be over-suppressing~~ — **REFUTED 2026-08-24.** robot-hands KEPT its sourced
> word numeral (*"six actuation types"*); webdesign's loss of *"sixty-three tools"* is the model
> obeying the **pre-existing** *"avoid our own catalogue or page count"* rule, which its own
> `avoid_leading_with` names in **all three** orderings including the pre-537 one. No suppression.
> **⚠ FIRE B4 WITH `scripts/fire-offer-analyser.sh`, NOT `run_improvement_sweep_once.sh`** — the
> latter fires the whole loop and promotes every `detected` item into live handler dispatches
> ([MEASURED] 111 on webdesign, 37 on leopardess, including other lanes').
>
> **UPDATE 2026-08-22 11:03Z — THE GATE IS NOW LIVE, AND `bugs_open/335` IS STILL OPEN.**
> Chassis rolled to `v1.0.1323`; migration `537` applied after a capability probe (with two controls)
> confirmed the action is registered in the running binary. Council APPROVED at round 3.
> **Live is not proven:** the gate has **never fired** — no offer-analyser run since 2026-08-19 — and
> leopardess still carries the false claim, because only a successful *re-run* replaces `lead_with`
> and the sweep is still disabled.
> **The one thing owed is the behavioural re-proof, and it spends an LLM run against the owner's
> standing cost control, so it is the owner's call rather than a session's.** If authorised:
> POSITIVE control = leopardess (must lose *"eight live sites"*; ⚠ coordinate with the leopardess
> lane first — it is holding our findings pending an owner design report). NEGATIVE controls =
> webdesign.co.uk (*"sixty-three tools"*) and robot-hands.com (*"six actuation types"*, *"2–3
> articles"*) must **KEEP** their specifics. ⚠ **NOT gaswholesalers** — it carries no cardinal at all
> and cannot discriminate.
>
> ~~The Go action is committed and inert until the next chassis roll; the migration that switches it
> on is deliberately held. The only thing owed before the roll is reading the council verdict.~~
> *(Superseded 2026-08-22: the roll happened, the verdict was read and APPROVED, 537 is applied.)*

## What happened today

Took `bugs_open/335` — this lane's own agent writing a false figure into rank 1 and stamping it
`from_field`. Candidate 1, both halves.

| piece | where | state |
|---|---|---|
| `verify_cited_cardinals` (new action) | commit `d79e4243c`, register **CLM-023** | built, 12 tests green, **INERT until the roll** |
| `537_offer_analyser_cardinal_attribution_gate_HOLD.sql` | commit `6b1f4cb08` | **HELD** — do not apply before the roll |
| council submission | `9a8f1283-574e-44d7-8e66-b84789ba0429` | **APPROVED r3** (13 seats, 2 advisory, none high) |
| docs + correction + landmine + wrong-call | commit `4d68303f8` | done |
| r1 fix: bind UPDATEs to the resolved row | commit `ba656ef47` | done, mutation-proven |
| r2 fix: clean path clears the drop record | commit `4ffd9c4ac` | done — a REAL defect |
| r2 fix: capability probe + CLM-023 conditions | commit `3b3941abb` | done |

## What the next session should do

1. **The council is DONE — nothing is owed there.** `9a8f1283-574e-44d7-8e66-b84789ba0429`
   **APPROVED at round 3** (13 seats approve, 2 advisory, none high-severity). The commits carry
   `Council-Submitted:`, so `098` credits them automatically — **no amend, and do not add
   `Council-Reviewed:` to those commits retrospectively.**

   **⚠ ONE ADVISORY IS A CONDITION ON THE NEXT CHANGE, not a nicety.** `bug_historian` left standing
   that `drop` mode records a removal only as an in-document marker with **zero automated
   consumers** (`bugs_open/034`'s shape), and that **any future caller opting into `drop` inherits
   the same gap unless the durable record is built into the ACTION itself.** CLM-023 carries the
   precondition for `offer_ordering`; the seat's wider point is that a register entry is the wrong
   home for it. **Do not add a second `drop` consumer before that is built.**

   **What the three rounds found is in `bugs_open/335`, and it is the useful part** — both REVISE
   rounds found something a green test suite did not (R1: the needle-gate and the `UPDATE`s were
   different sets; R2, sideways: my clean path omitted the drop key under a deep-merging writer, so
   a stale record would have accused a clean artefact for ever). Also recorded there: three
   objections across two rounds were factually wrong about the FILE because they described my
   **sketch**, and a fourth variant survived into the approving round. **Sketch the skeleton.**

2. **After the next chassis roll: apply `537`, then re-prove.** The file carries the full runbook.
   The order that matters: confirm the binary registers the action (ask the artefact, per SERVICE,
   with a control capable of being absent), apply, then ask **"what did I break?"** before "did it
   work?" — an unregistered action name fails the *whole* offer-analyser workflow.
   - **Positive:** re-run leopardess; no cardinal in a surviving point that is absent from its cited
     field; the removed one recorded under `data->'dropped_unsourced'`.
   - **Negative controls — and use THESE, not the one the bug file originally named:**
     webdesign.co.uk ("sixty-three tools"), robot-hands.com ("six actuation types", "2–3 articles")
     must all **keep** their specifics. ⚠ gaswholesalers **cannot discriminate** — no cardinal in any
     of its six points, so it passes a rule banning every numeral. Corrected in the bug file.
   - ⚠ **Applying `537` does not repair leopardess.** The false rank-1 is already persisted and only
     a *successful* re-run replaces it (`write_site_spec` merges; an array overwrites wholesale, but
     a failed run writes nothing). The sweep has been off since 08-17, so nothing re-runs it
     unprompted — and **the leopardess lane is holding this lane's findings pending an owner design
     report, so coordinate before firing B4 there.**
3. **Then the v2 batch** (`features_open/030` §10) — one migration, one re-proof. **v2(b) is now
   partly done**: `537` adds the attribution rule to the prompt for `lead_with` points, but **not**
   for `why` clauses, which is what v2(b) actually asked for. Remaining: **(d)** machine-checkable
   acceptance predicates (strongest; ~8 of 22 tests expressible, incl. the one that failed) ·
   **(b)** the `why`-clause half · **(a)** head-of-hero excerpt ⚠ *invalidates v1's truncation
   baseline — re-run that check on webdesign.co.uk after it* · **(c)** `primary_model` in the
   degraded arm ⚠ *latent, no live instance — must not motivate the batch*.
   ⚠ **v2(d) is NOT written down in `features_open/030` §10** — §10 holds only (a)/(b)/(c). It lives
   in `NOTES` ("v2(d) CENSUS", ~line 2214) and in the 08-17/08-18 handoffs. Someone should fold it
   into the feature file; until then, do not conclude from §10 that (d) does not exist.
4. **`features_open/034`** — claims audit over `site_specs` prose, owner-approved 2026-08-14, still
   not designed. **335 sharpens its case and does not replace it:** 034 asks whether the premise
   itself is true; 335's fix stops a page claim being imported into the premise layer.
5. **Sweep window** — 18 sites still lack a ranked record, ~4.5 hours to finish the estate. Owner's
   call. Enable by direct `UPDATE`, never a migration, and **disable in the same session.**
   ⚠ Worth doing *after* `537` is live, so new orderings are gated on the way in.

## Watch-outs (today's additions first)

- **⚠ A baseline is a claim about a COMMIT, not a moment.** `git archive HEAD` twice, minutes apart,
  gave two different codebases and a confident false conclusion that my change had broken three
  tests. Resolve the sha and archive *that*. `WRONG_CALLS.md`, 2026-08-21.
- **⚠ A numeric gate reused from `verify_report_prose` is DIGITS-ONLY** and silently passes every
  spelled-out numeral — which was this very bug. `LANDMINES.md`, entry armed today.
- **⚠ The column is `site_specs.data`.** `spec_data` is the *config key* on the `write_site_spec`
  step; typing it into SQL gets `column does not exist` (loudly, at least).
- **⚠ psql prints UTC, your shell prints BST.** Always toward alarm. Make the DATABASE subtract.
- **⚠ `count(*) = count(DISTINCT item_key)` is the WRONG dedup test** — `idx_swi_dedup` is
  `UNIQUE (site_id, item_key)`, per site.
- **⚠ A new producer on a deduped item type reads ZERO while working** — suppression is per SUBJECT,
  so a non-zero count "proves it works" while saying nothing about the repeat cases.
- **⚠ A roll is not a deploy.** Same-tag rebuilds serve the cached image. Probe `/proc/1/exe` with a
  negative control **capable of being absent** (a plausible fake sha, never 40 zeros).
- **⚠ A site with `created_at` today, 0 pages, or `status='active'` is UNDER CONSTRUCTION** and is
  not a fact about the estate.

## Who owns what nearby

`bugs_open/333` and `bugs_closed/301` belong to the 301 lane — contribute, do not compete.
**`bugs_open/335` is ours.** The **leopardess lane** is actively working that site and is holding our
findings pending an owner design report — **coordinate before firing B4 there again.**
`copy_quality_two_stage` + the LMC lane still work loanandmortgagecalculator.co.uk.

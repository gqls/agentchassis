# HANDOFF — vigilant designer + offer analyser (2026-08-21)

**COLD-START = this file + `bugs_open/335` (fix built today, still OPEN) + `features_open/030` §10
(the v2 backlog) + `features_open/034`.**
**This supersedes `HANDOFF_2026-08-20_continue_here.md`.**

> **Re-run every liveness claim here before acting.** ~570 commits/day land on this branch. In one
> session today the working tree would not compile (another lane mid-edit), and a baseline archived
> minutes apart came off two different HEADs. Verify against `git archive <resolved-sha>`, never the
> tree and never the moving name `HEAD`.

## The one-line state

> **`bugs_open/335` is FIXED IN BOTH HALVES AND STILL OPEN, because fixed is not live.**
> The Go action is committed and inert until the next chassis roll; the migration that switches it on
> is deliberately held. The only thing owed before the roll is **reading the council verdict**.

## What happened today

Took `bugs_open/335` — this lane's own agent writing a false figure into rank 1 and stamping it
`from_field`. Candidate 1, both halves.

| piece | where | state |
|---|---|---|
| `verify_cited_cardinals` (new action) | commit `d79e4243c`, register **CLM-023** | built, 12 tests green, **INERT until the roll** |
| `537_offer_analyser_cardinal_attribution_gate_HOLD.sql` | commit `6b1f4cb08` | **HELD** — do not apply before the roll |
| council submission | `9a8f1283-574e-44d7-8e66-b84789ba0429` | **submitted, VERDICT NOT READ** |
| docs + correction + landmine + wrong-call | commit `4d68303f8` | done |

## What the next session should do

1. **READ THE COUNCIL VERDICT — this is the one outstanding obligation.** The code is already on the
   shared branch, so a REVISE or REJECTED needs acting on, not filing.
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
    WHERE correlation_id='9a8f1283-574e-44d7-8e66-b84789ba0429' AND kind='council_report'
    ORDER BY created_at;
   SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
   ```
   If it approved, the migration commit already carries `Council-Submitted:` and `098` credits it
   automatically — **no amend, and do not write `Council-Reviewed:` retrospectively.**
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

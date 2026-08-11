# 255 — `missing_conversion_path` is routed at a handler that cannot read its spec, so it is refused and released back to the detector

**Filed 2026-08-11 by the vigilant_designer_offer_analysis lane, which OWNS the defect**
(`check_revenue_shape` is ours, BIZ-031/B3).

> **DIAGNOSIS LOOP: CONFIRMED, first iteration** (`stopped_by: confirmed`), run
> `RUN_CORRELATION_ID=64e5ab04-2a3d-4c5c-8d05-aa875ba211f5`, 2026-08-11 19:52Z. All five symptom
> clauses returned `[explained]` with citations. **And it produced evidence this file did not
> have:** it read the LIVE index definition out of `pg_indexes` rather than trusting the Go
> list —
> `CREATE UNIQUE INDEX idx_swi_dedup ON public.site_work_items USING btree (site_id, item_key) WHERE ((item_key IS NOT NULL) AND (status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved','cancelled'])))`
> — so the dedup-slot release is now measured at the DATABASE, not inferred from
> `workItemTerminalStatuses` agreeing with it. That upgrade is exactly why the run was worth its
> minutes: my version of the claim rested on the two staying in lockstep, and lockstep is an
> assumption; the index predicate is a fact.
>
> ⚠ **The run read 8 of 12 in-scope symbols. `ReadSymbolBody` could not read
> `(*RevenueShapeCheck).Run`, `(*RevenueShapeCheck).runConversionPath`, `workItemTerminalStatuses`
> or `workItemDispatchableStatuses`** — the two pointer-receiver methods and the two package-level
> `var`s. That is `bugs_closed/145` (fixed, council-approved, committed — **but not yet live**), so
> the tooling failure is expected and is NOT evidence about the code; the bundle says so itself
> ("Absence of a body here is never evidence that a symbol is irrelevant"). **Both facts it could
> not read are verified first-hand above with quoted evidence** — the live spec row, and the const
> list read from the file. Do not treat the CONFIRMED as independent corroboration of those two;
> treat it as independent corroboration of the handler-refusal and dedup-release links, which are
> the ones that were actually in question.

## Symptom, observed live

`check_revenue_shape` filed `missing_conversion_path:62b5978e-4271-4589-8e00-4baebfc0447c`
(mortgagecalculator.co.uk) on 2026-08-09. The finding is **true and hand-verified**: the site
records `lead_generation`, `contact-index` is `planned` and never shipped, the shipped
contactish candidate fell back to `index` (landing) which carries no form, and the only
`<form>`s on the site are calculator inputs on tool pages.

The improvement loop promoted it at 17:43:51Z on 2026-08-11 and `build-dispatch-loop` handed it
to its routed handler, `content-gap-planner`. At 19:01:49Z the item came back **`wont_fix`**,
with this in `error` — the handler's own words:

> "The content gap description and original category are both blank. There is no gap to
> evaluate. Please resubmit with a specific description of the missing content, the audience it
> serves, and any relevant search intent or user need it should address."

**The handler did not disagree with the finding. It could not see one.**

## Root cause — a producer/handler CONTRACT mismatch, not a judgement

[VERIFIED — read from the live row] The spec `runConversionPath` writes carries exactly:
`{check, primary_model, missing, rule, adopted_branch}` (+ `original_pipeline`, added by the
promoter). **There is no `description` and no `category`** — the two fields
`content-gap-planner`'s planner step reads. So the item is unreadable to the only agent it is
ever routed to.

[VERIFIED — `platform/orchestration/actions/apply_gap_plan_action.go`] The refusal path is
`applyNotActionable` (:1109-1123): it sets the originating item
`status='wont_fix', error=<reason>, completed_at=NOW()`.

[VERIFIED — `platform/orchestration/actions/work_items_common.go:40-55`, **and now at the live
index itself**, `pg_indexes`, via the diagnosis run] **`wont_fix` is a member of
`workItemTerminalStatuses`, AND `idx_swi_dedup`'s own `WHERE` clause excludes it**
(`status <> ALL (ARRAY[…,'wont_fix',…])`). So a `wont_fix` row **releases the dedup slot its
`item_key` was holding** — measured at the database, not inferred from the Go list agreeing with
the index.

[PREDICTED, NOT YET OBSERVED — one refusal has happened, and the next rotation is ~7 days out.
The dedup slot being FREE is now verified; that the detector will actually re-file into it is
not, because it needs the rotation to reach the site while the defect persists]
The consequence is a closed loop with no exit: the defect on the site is untouched, so
`check_revenue_shape` re-files the same `item_key` on its next rotation; the promoter promotes
every `detected` row on a site it reaches; the handler refuses it again for the same reason;
`wont_fix` releases the slot again. **One LLM call per rotation, for ever, with the site never
fixed and nothing anywhere reading as broken.** Marked as a prediction deliberately: the
mechanism is verified link by link, the recurrence is not, and it becomes checkable on the
next rotation tick (see "How to verify").

## Why this is a class, and why it is worse than `bugs_open/077`

`077` was *detector predicate wider than the handler's **remit***: the handler ran and its
transform could not touch part of the population. This is one level earlier — *detector's item
**shape** is not what the handler **reads***. The handler never gets as far as having a remit.

**The seam that would have prevented it already exists, in the file this lane extended today.**
`remit.go`'s `HandlerStepConfig(ctx, db, agentType, action)` exists precisely so a check can read
its handler's **LIVE step config** before filing, rather than a plausible reading of the Go
source — its own doc comment says nav-link-fixer's live row carries three patterns where the Go
default has four. `check_revenue_shape` routed at an agent without ever consulting it.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **A producer↔handler spec-shape lockstep (the class fix).** A check that names a
   `handler_agent` should have to declare the spec keys that handler reads, asserted by a test
   in the same package — the `RegisteredVerifiersMatchClaimTimeoutExclusion` pattern, applied to
   the routing edge instead of the verification edge. Closes the door for **every** check, and
   this defect is evidence the door is open in general, not just here. Most expensive; needs a
   home (WII-01x, and it is arguably the same family as WII-013: `item_type` as the join between
   producer and *predicate*, this as the join between producer and *handler*).
2. **Stop routing what nothing can handle: file the conversion-path finding as a
   `capability_gap` instead** (`GapRuleMissing`'s sibling, `GapHandlerRemit` — the handler exists
   and cannot act on this shape). A `deferred` + empty-`handler_agent` row **cannot be claimed at
   all**, so the churn is not merely unlikely, it is unrepresentable. Loses the automated repair
   this lane wanted, and honestly: there is no automated repair today, only a refusal.
3. **Give the spec what the handler reads** — add `description` (the `missing` sentence, plus the
   audience/intent the planner asks for) and `category` to `runConversionPath`'s SpecJSON, read
   from `HandlerStepConfig` rather than guessed. Cheapest, one edit, and it is the only candidate
   that could actually get a contact page built. ⚠ **But it hands a live site's page plan to a
   planner on the strength of a spec we hand-shaped** — and what it then builds has never been
   seen. Do not ship this one without watching the first plan it produces.
4. **Route somewhere else** (`needs_human_review`-shaped). Blocked today for the reason already
   on record in `check_revenue_shape.go`'s header: `sites` has no adopted/managed column to
   predicate that routing on (checked `information_schema` 2026-08-09).

**Owner context, 2026-08-11:** the owner chose to *roadmap* this finding rather than build the
page, then chose "let it plan, decide before it builds". Both are now moot for this round — the
machinery planned nothing and refused. **Candidate 3 is the only one that makes his second
answer meaningful**, and it is the one that needs watching, so it should not be shipped
un-witnessed.

## How to verify (and the control that makes it honest)

The churn prediction becomes checkable on the next rotation tick for mortgagecalculator.co.uk:

```sql
-- expect: a SECOND row on the same item_key, born after the wont_fix, refused the same way
SELECT id, status, created_at, updated_at, left(COALESCE(error,''),80)
FROM site_work_items WHERE item_type='missing_conversion_path' ORDER BY created_at;
```

**Positive control in the same query:** the existing `wont_fix` row must still be there with its
`error` intact — if it has vanished, something is reaping these and the churn theory is measuring
a different mechanism. **Do NOT grade the fix by re-running the detector**: it will file the item
again correctly, which is not the question. The question is whether the handler can read it.

## Ownership

`check_revenue_shape` and this item type: **vigilant_designer_offer_analysis** (this lane).
`content-gap-planner` / `apply_gap_plan_action.go`: not ours — candidate 3 touches only our spec,
candidate 1 touches a shared contract and would need the council gate and probably an RFC.

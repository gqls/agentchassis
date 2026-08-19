# 323 — `cta_improvement`: the handler says in its own payload that it cannot do this work, and 468 items completed anyway

> ## STATUS 2026-08-19 evening — OWNED by `bugfix_323_cta_improvement_refusal`; half LIVE+PROVEN, half committed and inert until the next chassis roll
>
> Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_323_cta_improvement_refusal/` (PLAN, RUNBOOK,
> NOTES, README_where_we_are, council submission). Diagnosis run `b218f39d` **CONFIRMED** the mechanism.
>
> **Correction to this file's premise, measured:** the handler's vocabulary ALREADY separates refusal from
> no-op — every refusal arm in `fix_component_template_action.go` carries `action:"needs_review"`, every
> idempotent no-op carries no `action` key (470 vs 299 rows lifetime, zero overlap). The "same field, same
> value, opposite meanings; only the `reason` separates them" claim below is wrong; only a READER was missing.
> The comment at `fix_component_template_action.go:58` claiming the flag stopped the loop was false.
>
> - **Layer 1 — LIVE + PROVEN (config, migration `495`, 20:02Z).** The fixer's own workflow now branches on
>   that key: refusals park at `needs_human_review` via `fail_work_item` (the page-build-handler pattern);
>   no-ops unchanged. Proven by dispatching the real fixer at a probe item (`claimed → needs_human_review`,
>   note `## refused: cta_improvement`), probe torn down. Composes with the 283 lane's `486_HOLD`.
> - **Layer 2 — committed `0e4622bab`, INERT until roll.** `classifyFinding` Rule 3 files `cta` /
>   `nav_restructure` as `capability_gap` (`handler_missing`, deferred, no handler, detail preserved, one
>   open row per site per category — `bugs_closed/077`) instead of dispatching at the fixer.
> - **Layer 3 — same commit.** `fixTypesRefusedByDesign` + `TestAuditRoutingNeverTargetsAFixerRefusalArm`:
>   the router can no longer name the fixer with a fix_type the fixer refuses (mutation-proven).
> - Council: round 1 **APPROVED** 20:31Z (corr `92829711`, 14 seats, 4 advisories none high — triage in the lane NOTES; two acted on: the shared `fallbackFixType` ladder, the 486/495 coordination note).
> - **Named residual:** the capability_gap row keeps the FIRST finding's suggestion/acceptance_test per site per category; later ones are counted as dedup skips (077's one-row-per-site shape, deliberate) — repointing `noHandlerCategories` at a real handler restores per-finding flow at the next audit.
>
> **The open question below is answered, by class:** DESTINATION defects are repaired by the resolver /
> `cta_links_stale` recompute regardless of the item (robot-hands.com/index, ~2h later, graded at
> `page_component_history`); LABEL/COPY defects have no handler at all — that handler ("one named
> component, one named defect → `field_updates` for section-editor") is the same missing piece the `277`
> and `301/083` lanes asked the `copy_quality_two_stage` lane for on 2026-08-19; this lane is its third
> customer and does NOT build it. **Candidate 3 (gate 1b boolean) is not needed.** Stays OPEN until the
> Go half is live (CLAUDE.md bar: fixed AND live).

**Filed 2026-08-19** by the `bugfix_302_design_repair_verification` lane (briefly numbered 318 for ~4 minutes until another session's 318 was found — renumbered rather than adding a seventh double-used number to the list CLAUDE.md already calls out), found while measuring
`spacing_fix` for an owner decision. **OPEN, UNOWNED.** Same class as `bugs_closed/302` /
`bugs_closed/213` D1 — a handler reporting it did nothing while its item closes green — at
**18× the scale** of the case that lane was filed on.

## The measurement

[MEASURED 2026-08-19, archive-inclusive — `site_work_items UNION site_work_items_archive`, because
the live table is only a ~7-day window]

```sql
WITH a AS (SELECT item_type, status, result, handler_agent, site_id FROM site_work_items
           UNION ALL SELECT item_type, status, result, handler_agent, site_id FROM site_work_items_archive)
SELECT handler_agent, status, count(*) n, count(DISTINCT site_id) sites,
       count(*) FILTER (WHERE result #>> '{response,fix_result,fixed}'='false') AS reported_not_fixed
FROM a WHERE item_type='cta_improvement' GROUP BY 1,2;
```

| handler | status | items | sites | reported **not** fixed |
|---|---|---|---|---|
| `component-template-fixer` | `complete` | **993** | 22 | **468** |
| `component-template-fixer` | `wont_fix` | 6 | 1 | 0 |

And the reason those 468 give, verbatim and identical across all of them:

> `fixed: false`, `reason: "fix_type requires LLM-driven changes, not programmatic"`

**That is not "already correct".** It is the handler stating that the work this item asks for is
outside what it does. The distinction matters and is the whole reason this is filed: the sibling
types measured in the same pass — `spacing_fix` (226 × *"already has flex CSS"*) and
`responsive_fix` (72 × *"already has responsive CSS"*) — report the same `fixed: false` shape and
mean the opposite thing, a repair finding its work already done. Same field, same value, opposite
meanings; only the `reason` separates them.

## What is NOT established, and must be before anyone calls this damage

**Whether the LLM work happens by another route.** There is no second handler for this item_type
(993 of 999 rows are `component-template-fixer`), but a *differently-typed* item could carry the
copy change — `content_rewrite`, `unresolved_cta` and `cta_names_unknown_destination` all exist and
all touch CTAs. Nobody has checked. **Until that is checked, the honest claim is "468 items closed
with their handler saying it did not do the work", not "468 CTAs are unimproved".**

⚠ Do not grade this by re-running the detector: `bugs_closed/213`'s landmine applies — a verifier or
check re-run will answer its own question correctly and tell you nothing about whether the item's
own request was met. Grade at the artefact, or against `spec.acceptance_test`.

## Why it is not simply "add it to gate 1b's roster"

Tempting, and wrong for two reasons this estate has already paid for:

1. **`fixed` is a BOOLEAN.** `noChangeGates` reads numeric `CounterPaths` via `lookupNumericPath`,
   which treats a bool as not-present — so every row would read as *unreadable*, not as *no change*.
   Gate 1b cannot express this handler's report at all without a contract change.
2. **The roster's bar is a measurement that a zero-change run cannot be a repair FOR THIS TYPE**, and
   for `cta_improvement` that has not been established — see the section above. An entry written
   without it is precisely the guess about somebody else's handler that roster forbids
   (`complete_work_item_no_change.go`, header).

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Route the type at something that can do LLM-driven copy work** — the structural fix, and the
   same shape as the owner's 2026-08-19 ruling on `dark_section` (route it at a handler that can
   actually make the change, rather than gating the one that cannot). `designRouting`'s `cta`
   category currently maps to `component-template-fixer` via `categoryToFixType`.
2. **Split the population at the producer, per `bugs_closed/077`'s owner ruling**: the fix_types this
   handler CAN do stay; the rest become one `capability_gap` — the estate's existing durable record
   for "found work I have no handler for", already read as a roadmap by `diagnose_triage_action` and
   `fixloop_digest_action`. No new item type.
3. **Teach gate 1b to read a boolean "did you change anything" flag**, then opt this type in. Widens
   a shared contract for one consumer; do it only if 1 and 2 are both refused.
4. **Not a fix:** leaving it. The items close green and the dedup key then suppresses re-detection,
   which is `bugs_closed/017`'s mechanism.

## Relations

`bugs_closed/302` (the same class, and the lane that found this) · `bugs_closed/213` D1 (gate 1b,
WII-017, and the do-not-re-run-the-detector landmine) · `bugs_closed/077` (the owner-ruled
partition-and-queue pattern candidate 2 borrows) · `bugs_closed/017` (green-close suppressing
re-detection) · `write_audit_findings_action.go` (`categoryToFixType`, `designRouting`) ·
LANDMINES *`site_work_items` is a ~7-DAY WINDOW* (why every count here is archive-inclusive).

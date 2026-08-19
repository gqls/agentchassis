# 323 — `cta_improvement`: the handler says in its own payload that it cannot do this work, and 468 items completed anyway

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

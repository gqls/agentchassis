# 204 — `plan_sections` resolves a section by NAME/FUNCTION only, so a decomposed page can never be rebuilt — and the build path asks the fleet to manufacture junk components

**Filed 2026-08-05** from the `loancalculator_couk` lane, while carrying out the
owner's instruction to rerun a site's copy through the framework rather than by hand.

**This is `bugs_closed/182` in the sibling call site.** 182 fixed exactly this
blindness in the RE-RENDER path (`rerender_page_sections`) by resolving
`page_components.component_id` first and falling back to name. The BUILD path
(`plan_sections` → `page-build-handler` → `page-content-writer`) never got the same
fix — **and 182's own commit edited this very file.**

## The defect

`pages.sections` on a decomposed site is a list of **positional slot names**:

```
loancalculator.co.uk / guide-how-loans-are-calculated:  ["prose-0", "prose-1"]
```

`plan_sections` resolves those against component **name and function only**:

- `loadComponentSchemas` (`plan_sections_action.go:1144`) — its own comment says it
  builds `componentInfo` records *"keyed by both name and function (the lookup pattern
  planSection expects)"*.
- `:918` — `comp, ok := components[sectionName]`.

`prose-0` is neither a component name nor a function (the function is `ported-prose`,
attached via `page_components.component_id`). So the lookup misses, control falls to
the selector at `:937`, and the section is deferred.

## Measured

**0 of 57** section names on loancalculator.co.uk resolve to any component by name or
function. Fleet-wide, **86 unresolvable across 5 sites**:

```sql
WITH s AS (
  SELECT si.domain, jsonb_array_elements_text(p.sections) AS sec
  FROM pages p JOIN sites si ON si.id=p.site_id
  WHERE p.status='active' AND jsonb_array_length(COALESCE(p.sections,'[]'::jsonb))>0)
SELECT domain, count(*) AS names,
       count(*) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM content_components cc WHERE cc.function=s.sec OR cc.name=s.sec)) AS unresolvable
FROM s GROUP BY domain ORDER BY 3 DESC;
```

```
loancalculator.co.uk          57    57   (100% — fully blocked)
gaswholesalers.com           122    11
finetuning.uk                152    10
leopardessconsulting.co.uk   106     6
oufe.com                      20     2
```

## Induced end to end, not argued

One real `content_rewrite` work item (`created_by='voiceh-canary'`) at
`guide-how-loans-are-calculated` — 2 prose blocks, no calculator, so a failure could
not touch arithmetic:

```
status: needs_human_review
error:  page-build-handler no-op: no sections ready to build (empty spec sections,
        or all sections deferred for missing data) — the target section was NOT rebuilt
```

**It refuses LOUDLY, which is correct and worth preserving** (the shape
`bugs_open/194`'s framework half shipped for). The defect is that it can never
succeed, not that it lies.

## ⚠ The second-order damage: it asks the fleet to build junk

The selector reads the unresolvable name as an unknown **component type** and files
work items to create it. My single canary produced:

```
needs_new_component  "Need component template for section type: prose-0"
needs_new_component  "Need component template for section type: prose-1"
needs_section_data   "Section 'prose-0' on guide-how-loans-are-calculated needs: "
needs_section_data   "Section 'prose-1' on guide-how-loans-are-calculated needs: "
```

All four **cancelled** with an explanatory note before a component-creator could act.
A full-site attempt on loancalculator would have filed **114** of these (57 × 2), for
components that already exist. **Anyone attempting a build-path run on a decomposed
site must sweep for these afterwards** — see the query in the lane handoff.

## Root cause, and why it survived 182

`a43be1e70` (182's fix) **modified `plan_sections_action.go`** — 199 lines — to factor
`componentInfoFromRaw` so the template-truncation guard *"can't drift across the three
now-shared conversion sites"*. It refactored around this lookup and left it keyed by
name/function, adding `component_id`-first resolution only to the re-render path.

This is the documented shape in 016b §9: *one call site of a shared judgement gets the
rigorous fix, the sibling stays heuristic.* Same family as `bugs_closed/041` (section
lookup never normalises) and `bugs_closed/095` (wrong slot name renders nothing and
reports complete) — this is the third or fourth appearance of "the section→component
lookup has one more spelling than anyone checked".

Re-verified at chassis **v1.0.1254**: `plan_sections_action.go` untouched since
2026-08-04. Open and unowned.

## Fix candidates, ordered by what closes the door

1. **Resolve by `page_components.component_id` first, fall back to name/function** —
   exactly what 182 did one function over, and its `loadComponentSchemasByID` /
   `loadContentComponentsByID` already exist. **Do not write a third resolver.**
   ⚠ **The real design question:** `plan_sections` has no `pageID` at that point in
   the workflow — its own comment says so (`:1140`, *"plan_sections doesn't have a
   pageID at this point"*). Work out where the page id comes from before writing code;
   that, not the lookup, is the hard part.
2. **Make the miss fail loudly at the lookup**, naming the unresolved names, instead of
   silently routing to the selector. Weaker (it does not enable the rebuild) but it
   would have turned this into a one-line diagnosis, and it stops the junk work items.
3. Re-point `pages.sections` at component functions. **Rejected:** slot names are
   positional and a page with three prose blocks would collide on one function name —
   the positional naming exists precisely to disambiguate them.

Candidate 1 makes the bad state unrepresentable; candidate 2 only makes it visible.
Both are worth having, and 2 is cheap.

## How to verify a fix

- The census above returns **0 unresolvable** for loancalculator, or the lookup
  resolves them by id.
- Re-fire the canary and assert the prose actually changes:
  ```sql
  SELECT pc.slot_name, left(regexp_replace(pc.content_data->>'content','<[^>]+>',' ','g'),200)
  FROM pages p JOIN page_components pc ON pc.page_id=p.id
  WHERE p.name='guide-how-loans-are-calculated' ORDER BY pc.position;
  ```
  Pre-fix baseline: `prose-0` opens *"How Your Monthly Repayment is Actually
  Calculated / Demystifying the 'Amortisation' formula… Most people see a monthly loan
  repayment as a flat fee. In reality…"* (1993 b), `prose-1` 192 b.
- **Zero `needs_new_component` items filed** for the run.
- ⚠ The 12 tool rows on that site are `lock_type='permanent'` — a fix must leave them
  untouched. Backups: `page_components_bak_20260805_framework_rewrite` (63 rows).

## Filing basis

**CLAUDE.md requires a cross-cutting structural claim to go through the `090`
diagnosis loop, or the filing session to state plainly why it substituted equivalent
first-hand verification. This is that statement.** Substituted, because all three
links were read directly and the failure was then INDUCED on a live page rather than
predicted: the live `pages.sections` value, the live `content_components` schema, the
source of `loadComponentSchemas`/`planSection`, the live `page-content-writer` config,
a fleet-wide census with a non-zero result on 5 sites, and a real work item whose
error message names the failure. The one thing a `090` run would add that this does
not is an independent reader — worth having if the fixing thread wants it, and the
symptom to file would be *"plan_sections resolves pages.sections against
content_components.name/function; sites whose sections are positional slot names
cannot be rebuilt"*.

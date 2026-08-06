# 189 — resolving a previously-unresolvable LOCKED section duplicates it on the page

**Filed 2026-08-03**, discovered while inducing the live verification for
`bugs_closed/182` (component_id-first resolution) on loancalculator.co.uk's
`tool-loan-vs-savings` page. **This is a real consequence of 182's fix, not a
flaw in it** — 182's resolution is correct and desired; this bug is a
pre-existing defect in `save_page_sections_action.go` that 182's fix newly
*reaches*, because before 182 these sections never resolved at all.

## What happened, measured

Firing the documented `section_data_resolved` re-render on
`tool-loan-vs-savings` (work item `b46a134b-c17c-4f6b-8aa1-23fa942ec354`,
2026-08-03 11:12 UTC, chassis v1.0.1240 — 182's fix) produced **five**
`page_components` rows where there should be four, and the served page
(`https://loancalculator.co.uk/tools/loan-vs-savings.html`) rendered the
loan-vs-savings calculator **twice**:

```
 slot_name             | position | locked | html_len
 ported-prose          |    1     |  f     |   728
 ported-prose          |    2     |  f     |   200
 tool-loan-vs-savings  |    3     |  f     |  11844   <- NEW, fresh render
 ported-prose          |    4     |  f     |   685
 tool-2                |    5     |  t     |  11845   <- OLD, locked 2026-08-02
```

Both rows shared `component_id='448422ce-fbf0-4e3d-98a1-fab0a6e856ed'` and
**identical `content_data`** (same md5) — the fresh render reproduced the
locked, proven content almost exactly (1-byte difference, incidental). The
served page had 5 top-level `<section>` blocks and `data-component="ported-prose"`
×3 where there should be 3 distinct positional names.

**Remediated live, same session**: deleted the duplicate (position 3, unlocked),
repositioned the locked row back to position 3, restored the three prose rows'
slot_names from the generic `ported-prose` back to their original `prose-0`/
`prose-1`/`prose-3`, and fired an assemble-only redeploy (no `reason`) to ship
the correction. Verify this landed before reading anything else in this file as
current: `curl -s https://loancalculator.co.uk/tools/loan-vs-savings.html | grep -c '<section'` should read 4.

## The mechanism — two pre-existing defects that compound

**1. `extractSectionsFromMetadata` prefers `component_function` over
`component_name`** (`save_page_sections_action.go:896-902`):

```go
componentName := "section"
if fn, ok := m["component_function"].(string); ok && fn != "" {
    componentName = fn
} else if name, ok := m["component_name"].(string); ok && name != "" {
    componentName = name
}
```

`RerenderPageSectionsAction`'s successful-render entry sets BOTH fields —
`component_name: s.slotName` (the stored positional name, e.g. `tool-2`) and
`component_function: comp.Function` (the component's own identity, e.g.
`tool-loan-vs-savings`) — and has done so unchanged since before 182. Before
182, a positionally-named slot NEVER reached this success path (nothing
resolved, so `carryStoredSection` ran instead, which sets only
`component_name`, no `component_function` key at all) — so this precedence
rule was dormant for exactly the population 182 fixes. 182 makes resolution
succeed, `component_function` gets populated, and the persisted `slot_name`
silently becomes the generic component identity — **destroying the
deliberate positional naming** the loancalculator decomposition chose
specifically "so that a dropped-section warning names which paragraph
vanished" (`bugs_closed/182`'s own rationale).

**2. `matchLockedRow` matches by `section.ComponentName`**
(`save_page_sections_action.go:586`):

```go
if lr := matchLockedRow(lockedRows, section.ComponentName); lr != nil {
    lr.consumed = true
    // reposition the locked row, discard the fresh copy, continue
}
```

The locked-row guard (`bugs_closed/058`) is supposed to make exactly ONE of
{locked row, fresh copy} survive. It works by matching the INCOMING section's
name against the locked rows' `slot_name`. Once defect #1 renames the incoming
section from `tool-2` to `tool-loan-vs-savings`, the match against the locked
row (still named `tool-2`) **fails silently** — `matchLockedRow` finds
nothing, the guard never fires, the fresh copy is inserted as a **new** row,
and the locked row (excluded from the DELETE-all by its own protection) also
survives. Both defects individually look like "reasonable behaviour"; together
they mean **a locked, positionally-named section is duplicated, not
protected**, the first time it becomes resolvable.

## Blast radius — measured 2026-08-03, re-run before trusting

Sections that are BOTH positionally-named (unresolvable by name/function,
resolvable only by `component_id` — 182's repair population) AND currently
locked:

```sql
SELECT s.domain, count(*) FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
LEFT JOIN content_components cc  ON cc.function = pc.slot_name
LEFT JOIN content_components cc2 ON cc2.name    = pc.slot_name
WHERE cc.function IS NULL AND cc2.id IS NULL
  AND pc.component_id IS NOT NULL AND pc.locked_at IS NOT NULL
GROUP BY s.domain;
--  loancalculator.co.uk | 12
--  oufe.com             |  2
```

**14 sections across 2 sites are armed**: any `section_data_resolved` (or
`image_landed`) re-render that reaches one of these will duplicate it exactly
as above, on the next roll or the next manual fire — whichever comes first.
This is a LATENT trap `182`'s fix newly created reachability for; it did not
exist before `182` because these sections could never resolve, so the locked
guard's silent mismatch never had a fresh copy to fail to catch.

## Fix candidates, ordered by what closes the door

1. **Match `matchLockedRow` by the STORED slot_name / position, not the
   post-resolution `ComponentName`.** `RerenderPageSectionsAction` (and any
   other producer of `sections_metadata`) already knows the ORIGINAL
   `page_components.slot_name` for every section it re-renders — thread it
   through as a stable key (e.g. a `stored_slot_name` field on the metadata
   entry, separate from the display/diagnostic `component_name`) and match
   locks against THAT, never against a name that can change between the read
   and the write.
2. **Stop letting `component_function` silently overwrite the stored
   `slot_name` in `extractSectionsFromMetadata`.** The positional name is
   information (which section this is), the function is a different piece of
   information (what renders it) — collapsing them into one field is the
   underlying defect. Candidate 1 is required regardless of whether this one
   is taken; this one prevents the *diagnostic* regression even outside the
   locked case (any positionally-named, unlocked section loses its identity
   the first time it resolves, fleet-wide, once 182 is live — 65 sections,
   not just the 14 locked ones).
3. Weakest: detect the resulting duplicate after the fact (same `component_id`
   twice on one page) and raise a work item. Does not prevent the live
   duplication, only shortens how long it's visible.

**1 and 2 are complementary, matching 182's own candidate reasoning**: 1 stops
the duplication (the acute harm), 2 stops the silent rename (the property this
site's design depends on) even where nothing is locked.

## How to verify a fix

Do NOT re-induce on `loancalculator.co.uk` again casually — every fire of
`section_data_resolved` on a page with one of the 14 armed sections currently
reproduces this defect until it's fixed. To verify a fix without touching
production content: fire the same re-render on `tool-loan-vs-savings` (page_id
`558f9f3f-ebac-4e4a-8265-30721054f351`, site_id
`0162cde4-633e-45e9-8ca6-87a6b2fe1d26`) and confirm the result is **exactly 4**
rows afterward, `tool-2` still locked at position 3, `locked_at`/`locked_by`/
`id` unchanged (058's own invariant) — not 5.

## Related

- `bugs_closed/182` — the fix whose success exposed this; not a defect in 182
  itself. Cross-linked both directions.
- `bugs_closed/058` — the lock-preservation guard this defeats; its own
  verification ("locked row's id/md5/updated_at unchanged, unlocked sibling
  rebuilt") was true for every case it was tested against because none of
  those cases involved a NAME CHANGE between the locked row and its own
  incoming re-render.

## Diagnosis loop

Not run through `090` — filed from direct, first-hand verification: the exact
live rows (before/after, md5-compared), the exact two functions read and
quoted above with line numbers, and the mechanism reproduced once, deliberately,
in a controlled test that was then remediated in the same session. Per the
2026-07-31 owner ruling's stated escape hatch: this substitutes for the loop
because the causal chain is fully read, not inferred, and re-running it live to
generate a second data point would recreate the very duplication being
reported.

---

## §Blast radius extension 2026-08-06 — the BUILD path becomes a second armed route (added by the 204-fixing session, 7fffb7ef)

`13252f714` (the `bugs_open/204` fix, committed 2026-08-06, inert until an image
rolls) gives the BUILD path (`plan_sections` → `page-content-writer` →
`compile_page_sections` → `save_page_sections`) the same component_id-first
resolution 182 gave the re-render path. **That makes this bug's trap reachable
from a second direction, and the build path's version is WORSE for the rename
half**, because unlike the re-render path it never carries the stored slot name
at all:

- `RenderComponentAction` (`v3_site_actions.go:1899-1903`) outputs
  `component_name: comp.Name` and `component_function: comp.Function` — the
  COMPONENT's identities. The planned section's positional name
  (`sectionPlanItem.Name`, e.g. `prose-0`) is on `current_section.name` and is
  copied into NEITHER.
- `extractSectionFromMap` (`v3_site_actions.go:2142+`) forwards only
  `component_id`/`component_name`/`component_function`/`content_data` into
  `sections_metadata`.
- So `extractSectionsFromMetadata`'s function-first preference (§mechanism
  defect 1 above) persists the slot as the component function — on the 204
  canary page BOTH prose slots would come back named `ported-prose`, the
  positional naming destroyed with no field anywhere still holding it. On the
  re-render path at least `component_name: s.slotName` preserved it.
- Defect 2 (matchLockedRow by post-resolution name) then fires identically:
  the 14 armed locked sections (12 loancalculator, 2 oufe) duplicate on any
  build-path run that reaches them, once 13252f714 is live.

**Consequences, until this bug is fixed:**
- The 204 closure canary must run on an UNLOCKED page (or a throwaway page)
  and must expect + restore the slot rename, exactly as this file's
  remediation did; it must NOT run on any page holding one of the 14 armed
  locked rows.
- Fix candidate 1 (thread the stored slot name through the metadata as a
  stable key) now needs the producer fixed on BOTH paths:
  `RerenderPageSectionsAction`'s entries AND the build path's
  (`RenderComponentAction` output or `extractSectionFromMap` — the loop's
  `current_section.name` is available in collected_data at compile time).
- 204's fix is NOT the defect here (same verdict as 182 in §header: resolution
  is correct and desired); this save-path defect predates both and is the
  remaining half of the decomposed-site rebuild story.

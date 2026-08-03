# 182 — a re-render that carried EVERY section is indistinguishable from one that worked

**Filed 2026-08-03 from the loancalculator.co.uk lane**, after shipping four
calculator fixes into `content_components`, firing the documented re-render, and
getting back a `complete` work item, an unchanged page, and no signal anywhere
that the fixes had not been applied.

## The mechanism

`RerenderPageSectionsAction` resolves each section's component like this:

```go
// rerender_page_sections_action.go:227-232
names := make([]string, 0, len(stored))
for _, s := range stored {
    names = append(names, s.slotName)          // <- slot_name, not function
}
schemas := loadComponentSchemas(ctx, params.DB, names, logger)
```

`loadComponentSchemas` → `loadSectionComponents` matches `content_components` by
**name or function** (`v3_site_actions.go:3595`, raw + kebab-normalised since
`bugs_closed/041`). So the lookup key is a slot name and the corpus is keyed by
component identity. Those coincide only when a site happens to name its slots
after its components.

When they do not coincide, every section takes this branch:

```go
// rerender_page_sections_action.go:309-317
comp, haveComp := schemas[s.slotName]
if !haveComp {
    logger.Warn("rerender_page_sections: component not found, carrying stored HTML", ...)
    sectionsMetadata = append(sectionsMetadata, carryStoredSection(s))
    carried++
    continue
}
```

`carryStoredSection` re-emits the stored `rendered_html` unchanged. The action then
returns normally, `save_sections` writes the carried bytes back, `render_page`
assembles them, `deploy_page` ships them, and the work item goes `complete`.

**A re-render in which NOTHING was re-rendered completes successfully.**

## Why this is the bug, and not just a naming mismatch

The naming mismatch is arguably the caller's problem. **The defect is that the
outcome is unobservable.** Compare `bugs_closed/095`, which fixed exactly this
shape one layer down: the *assembler* used to return empty and report COMPLETED,
and 095 split that into "rows exist and contribute nothing" (fail, naming the
slots) versus "no rows at all" (legitimate skip). The re-renderer has the same
ambiguity and did not get the same treatment:

| outcome | `rerendered` | `carried` | step result | work item |
|---|---|---|---|---|
| every section re-rendered | N | 0 | ok | `complete` |
| some sections legitimately carried | N−k | k | ok | `complete` |
| **NOTHING resolved, everything carried** | **0** | **N** | **ok** | **`complete`** |

The third row is a no-op dressed as a success. `carried == section_count` while
`rerendered == 0` is not a degraded render — it is **the action failing to do the
one thing it exists to do**, and it is the only case where a template change
silently does not reach the page.

The `Warn` is emitted per section, so the information exists — but it is a warning
among thousands, on a run whose own status says success, on a page that still
looks right because the carried bytes ARE the previous good render.

## Evidence (measured, not inferred)

Work item `b0c2265d-a49a-442f-8a8f-15a919c8badd`, orchestration
`439489b6-73fa-4755-be37-2f3982a9cef9`, page `tool-loan-vs-savings` on site
`0162cde4-633e-45e9-8ca6-87a6b2fe1d26`, 2026-08-02 22:51 UTC:

```sql
SELECT collected_data->'rerender_sections'->>'rerendered' AS rerendered,
       collected_data->'rerender_sections'->>'carried'    AS carried,
       collected_data->'rerender_sections'->>'skipped'    AS skipped
FROM orchestration_states
WHERE correlation_id='439489b6-73fa-4755-be37-2f3982a9cef9';
--  rerendered | carried | skipped
--  0          | 4       | false
```

At that moment `content_components.html_template` for `tool-loan-vs-savings` had
already been updated (md5 `6cbb6e090419`, 11989 bytes, 22:39:35Z) and the
`page_components` row pointed at that exact component row. The page came back with
the OLD content — identifiable because the rendered output still contained the old
template's CSS comment (`OPEN DEFECT`) and none of the new markup (`lvs-badge`,
`lvs-copy`, `winner-badge`).

The site's slots are `prose-0`, `prose-1`, `tool-2`, `prose-3` — positional and
kind-bearing, chosen deliberately so that a dropped-section warning names which
paragraph vanished (`decompose/assemble_mirror.py:270-273`). No component is called
`tool-2`.

## Blast radius — how many sites cannot be re-rendered

Any site whose `slot_name`s are not component names/functions. Count it before
quoting a number; this is the query, not the answer:

```sql
SELECT s.domain,
       count(*) FILTER (WHERE cc.function IS NULL AND cc2.id IS NULL) AS unresolvable_slots,
       count(*) AS total_slots
FROM page_components pc
JOIN pages p  ON p.id = pc.page_id
JOIN sites s  ON s.id = p.site_id
LEFT JOIN content_components cc  ON cc.function = pc.slot_name
LEFT JOIN content_components cc2 ON cc2.name    = pc.slot_name
GROUP BY s.domain
HAVING count(*) FILTER (WHERE cc.function IS NULL AND cc2.id IS NULL) > 0
ORDER BY 2 DESC;
```

Measured 2026-08-03 (re-run it; slot names change when pages are rebuilt):

```
domain                 unresolvable   total slots
loancalculator.co.uk        63            63     <- 100%, every slot on every page
finetuning.uk                7           215
lendzy.co.uk                 3            66
gaswholesalers.com           2           153
oufe.com                     2            22
gamesdesign.co.uk            1           105
```

**78 slots across 6 sites**, and the two shapes fail differently:

- **loancalculator.co.uk is total** — nothing resolves, so a re-render is a pure
  no-op. That is how this was found: four fixes had nowhere to go.
- **The other five are PARTIAL, which is worse to detect.** Those pages
  re-render most sections and silently carry one or two. `rerendered` is non-zero,
  so even fix candidate 1's `rerendered == 0` test would NOT catch them — the run
  looks entirely normal and one section is quietly frozen at whatever it last
  rendered to. Any candidate should therefore key on `carried > 0` with an
  unresolved-component reason, not on the all-or-nothing case that happened to be
  the one this lane hit.

## Fix candidates, ordered by what closes the door

1. **Fail the step on ANY section carried for `component not found`**, naming the
   unresolved slots — the `bugs_closed/095` treatment, applied to the re-renderer.
   Note the predicate: **not** `rerendered == 0`. Five of the six affected sites
   are partial (see above), so an all-or-nothing test would clear exactly the cases
   that are hardest to spot by eye. Distinguish this from the *legitimate* carries
   — "section not ready" and "empty html_template" — which are deliberate
   fallbacks, not failures to resolve. Makes the silent no-op unrepresentable
   without fixing resolution, and would have saved this lane a whole cycle.
2. **Resolve by `component_id` first, falling back to `slot_name`.** The
   `page_components` row already carries `component_id` and `loadStoredSections`
   already selects it (`s.componentID`) — it is read and then not used for the
   lookup. This is the actual repair: the row knows exactly which component it is,
   and a name-based search for something the row can point at directly is the
   defect underneath the defect. Slot naming becomes free again.
3. Emit a `lock_blocked_change`-style review item on a fully-carried run.
   Weakest — it adds a signal rather than removing the ambiguity.

**1 and 2 are complementary, not alternatives:** 2 makes it work, 1 makes the next
failure visible. 2 alone would leave a differently-shaped no-op just as silent.

## How to verify a fix

Do NOT verify on a healthy page — a green re-render of a page that would have been
carried anyway proves the deploy, not the fix. Induce it:

1. Point a page's `slot_name` at something no component is called.
2. Change that component's `html_template`.
3. Re-render with `spec.reason='section_data_resolved'`.
4. **Before:** `complete`, page unchanged, `rerendered=0/carried=N`.
   **After (candidate 1):** the step fails and names the slot.
   **After (candidate 2):** the page picks up the template change via `component_id`.

## Related, and why this is none of them

- `bugs_closed/095` — same *shape* (empty result reports COMPLETED) one layer down,
  in the assembler. Its correction explicitly records that slot mismatch was NOT
  its mechanism. This is the re-renderer's version, still open.
- `bugs_closed/041` — section lookup not normalising, so `snake_case` vanished.
  Fixed by matching raw + kebab-normalised. Does not help here: `tool-2` is already
  kebab-case and is not a stylistic variant of any component name.
- `bugs_closed/024` — a durable tool fix could never re-render, via the
  `</section>` truncation guard rejecting tools that end `</script>`. Different
  gate, same consequence, and its fix is why `componentTemplateValid` takes
  `component_level`. **Closed and live**, so it is not this.

  ⚠ **But it names a SECOND route to the same empty `schemas` map, and that
  matters for the fix.** Its council submission (corr `7ef4de4e`, edit 5) records
  that `sectionTemplateValid` "dropped self-contained tool templates from the
  schemas map, which would have made the exemption a no-op on the very component
  it was written for". So a component can be absent from the map because its NAME
  did not match (this bug) **or** because a validity check discarded it (024's
  edit 5) — and `!haveComp` cannot tell them apart. That is an argument for fix
  candidate 1 over 2: candidate 2 repairs one route in, candidate 1 makes *every*
  route in visible. Take both, and take 1 first.

## Diagnosis loop — **CONFIRMED**, first iteration

Filed to `090` per the owner ruling of 2026-07-31 (a `bugs_open/` file asserting a
cross-cutting root cause goes through the loop).
`RUN_CORRELATION_ID=834a24b0-4d3e-4ce7-a17c-6b270493bfd6`, `outcome: CONFIRMED`.

It reached the mechanism independently and cited the same five code sites plus
live state — worth reading, because its chain is tighter than the one above and
names a step this file glossed:

> stored `slot_name` values (`prose-0`, `prose-1`, `prose-3`, `tool-2`) are
> positional and never equal the matching `content_components.name`/`function`
> (`Ported Prose Block`/`ported-prose`, `Pay Off Loan or Save?`/`tool-loan-vs-savings`);
> `loadSectionComponents` matches only on `name`/`function IN(...)`, so the query
> returns nothing for these slot_names, **the resulting stub carries no
> `component_id` and is dropped by `loadComponentSchemas`**, so `schemas[s.slotName]`
> is absent and the rerender action's `haveComp` check forces every section into
> `carryStoredSection`.

The bolded step is the part this file did not have: `loadSectionComponents` does
not return nothing, it returns a **name-stub**, and `loadComponentSchemas` then
drops it precisely because it has no `component_id`
(`plan_sections_action.go`: `if _, hasID := comp["component_id"]; !hasID { continue }`).
That matters for fix candidate 2 — the `component_id` the row already carries is
being used as a *filter* two layers up while the lookup that needed it went by name.

Citations: `names = append(names, s.slotName)`,
`comp, haveComp := schemas[s.slotName]`,
`"rerender_page_sections: component not found, carrying stored HTML"`,
`FROM content_components WHERE name IN (%s)`, the `hasID` drop, the two live
`page_components` rows, and the orchestration output
(`"carried": 4, "rerendered": 0, "section_count": 4, "escalated": false`) against
the work item's `complete`.

## What the lane did meanwhile

Not blocked on this. `decompose/render_tool_row.py` renders the component offline
with the same Go engine, writes `rendered_html` directly, and lets the
**assemble-only** branch (`render_page`, no `spec.reason`) stitch it — the route all
27 pages were originally shipped through, proven byte-exact. Its `--check` runs a
control that re-renders from a baseline ref and requires the CURRENT stored bytes
back, so it refuses to write when the offline renderer and the live row disagree.
That is a workaround for this site, not a fix for the platform.

## FIXED 2026-08-03, LIVE v1.0.1240, pod-verified both replicas

Took both fix candidates, in the order the file recommended (2 makes it work, 1
makes the next failure visible):

- **Candidate 2**: `page_components.component_id` — already read by
  `loadStoredSections`, never used for the lookup — is now resolved FIRST via a
  new `loadComponentSchemasByID`/`loadContentComponentsByID`, falling back to
  `schemas[s.slotName]`. The raw-map→`componentInfo` conversion (incl. the
  `componentTemplateValid` guard) was factored into `componentInfoFromRaw` and
  shared across all three now-existing lookup paths, closing the drift pair
  `loadComponentSchemas`/`loadSingleComponentSchema` already had.
- **Candidate 1**: a `rerenderResolution` struct (mirroring
  `bugs_closed/095`'s `pageAssembly`) accumulates named `UnresolvedSlots` and
  `InvalidTemplateSlots` through the render loop — sections still carry, so the
  run completes its pass — then the step FAILS after the loop if either list is
  non-empty, naming every slot. Predicate is `carried-for-unresolved > 0`, NOT
  `rerendered == 0`, per the file's own warning (5 of 6 affected sites are
  partial). `NotReadySlots`/`EmptyTemplateSlots` stay non-fatal, matching
  `bugs_closed/095`'s council-corrected predicate rather than its rejected
  first draft — and are now surfaced in the output alongside `rerendered`/
  `carried`.

**Pre-ship measurements** (in-package scratch test calling the real functions,
deleted after use): 0 of the 65 newly-resolvable sections trip
`missingRequiredLLMFields`; 0 of 110 referenced components fail
`componentTemplateValid`. **Also measured, and logged observe-only rather than
switched silently**: 13 sections fleet-wide where `component_id` and a
coincidental name/function match both resolve and DISAGREE — id now wins,
confirmed against substantively different templates by byte length.

**Files**: `plan_sections_action.go`, `v3_site_actions.go`,
`rerender_page_sections_action.go`, `rerender_page_sections_resolve_test.go`
(8 new tests). Commit `a43be1e70`. Council submission `80fbbe7d-9b79-4dbf-be6a-286d3fe084a4`
(verdict pending at close — see `Council-Submitted:` trailer; `098` resolves it
automatically).

**Pod-verified**, both replicas, positive + negative string controls:
```
strings /app/agent-chassis | grep -c 'could not resolve a component and were carried unrendered instead'  # 1
strings /app/agent-chassis | grep -c 'bugs_open/182'                                                       # 1
strings /app/agent-chassis | grep -c 'template truncated, rejecting'   # the OLD, now-replaced wording — 0
```

**Induced live verification**, `tool-loan-vs-savings` (site
`0162cde4-633e-45e9-8ca6-87a6b2fe1d26`, page `558f9f3f-ebac-4e4a-8265-30721054f351`):

- Positive: fired `section_data_resolved` unmodified → all 4 sections
  resolved via `component_id` (0 unresolved) where before the fix this page's
  own recorded evidence was `rerendered: 0, carried: 4`.
- Negative: temporarily broke `prose-1`'s `slot_name`/`component_id` → the
  step FAILED with `page "tool-loan-vs-savings": 1 of 4 section(s) could not
  resolve a component and were carried unrendered instead — unresolved
  component [zzz-induced-test-182 (pos 2)]; invalid template []; not ready
  (legitimate) []; empty template (legitimate) [] (bugs_open/182)` — restored
  immediately after, work item cancelled.
- Re-ran the blast-radius query: `loancalculator.co.uk` and `oufe.com` now
  show **0** name-AND-id-unresolvable slots.

**A second, distinct defect was found and handled while inducing the positive
case**: resolving a LOCKED, positionally-named section duplicates it on the
page instead of the lock guard protecting it (a pre-existing interaction in
`save_page_sections_action.go` that 182's fix newly makes reachable). Filed
separately as `bugs_open/189` (NOT folded into this fix — different file,
different root cause), remediated live in the same session (4 rows restored,
no content lost), and documented as a landmine + in the loancalculator lane's
own NOTES/README so nobody fires the documented re-render at the lane's other
12 locked sections (or oufe.com's 2) before 189 is fixed.

**Not touched by this fix, noted for whoever picks it up**: 13 of the original
78 slots have neither a resolvable name NOR a `component_id` (9 empty-content
stubs that already escalate to the writer today, unchanged; 4 live tool pages
on lendzy.co.uk/gamesdesign.co.uk whose remediation is a `component_id`
backfill — data repair, not an action-code change).

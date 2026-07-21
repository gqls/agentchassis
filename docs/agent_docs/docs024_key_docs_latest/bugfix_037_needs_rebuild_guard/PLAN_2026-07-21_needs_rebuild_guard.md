# PLAN — bugs_open/037: protect a `needs_rebuild` page's composition from a re-plan

**Started & shipped 2026-07-21.** Filed by the owner as a *deliberate decision* left open by
`/bugs_closed/001` ("it may well be the wanted behaviour"). The point of this workstream was to make
that choice on evidence, then implement it.

## The question

`/bugs_closed/001` widened the re-plan preservation guard from adoption-locked to
`adoption_locked OR build_status='deployed'`. A `needs_rebuild` page falls **outside** the preserved
set, so a re-plan takes the LLM's composition for it — even when the page has a full previously-built
composition in `pages.sections`. Is that the wanted behaviour (`needs_rebuild` = "recompose me") or a
defect (silent composition loss)?

## Decision: it is a defect. Ship candidate 2.

**Refuted the "leave as-is" reading against the code.** Every writer of `needs_rebuild` preserves
`pages.sections` and means *"re-render as planned"*, never *"recompose from scratch"*:

| writer | intent | sections |
|---|---|---|
| `UpdatePageStatusAction` (`v3_site_actions.go:644`) | refused 0-component / partial deploy | **kept**, `built_from_plan_version` cleared |
| `flagPagesForRebuild` (`maintenance_actions.go`) | image / maintenance rebuild | **kept** |
| `markPagesForRebuild` (`store_generated_component_action.go`) | a now-available component the sections **already name** | **kept** — recomposition would *defeat* it |
| `check_unresolved_sections.go` | same: sections reference a now-available component | **kept** — recomposition would *defeat* it |

The last two make the "explicit redesign intent" reading impossible: the whole reason the page is
`needs_rebuild` is to render the components its **existing** sections name. So a re-plan replacing
those sections is silent loss.

## The fix (composes with the in-flight 050 work)

`v3_site_actions.go` was, at the time, mid-flight under another session's uncommitted work for
`/bugs_open/040` (partial-build guard), `/041` (kebab section lookup) and **`/050`** (the
deployed-empty gate). 050 had overloaded `realisedPageIsBuilt` for *two* jobs: preservation-set
**membership** and the empty-sections **classification**. So the fix could not simply widen
`realisedPageIsBuilt`.

- Add a **separate membership predicate** `realisedPageCompositionIsPreserved(rm)` =
  `deployed OR needs_rebuild`. Use it at the two membership sites (preserved-set filter; truncation
  must-keep).
- **Leave** `realisedPageIsBuilt` (= `deployed`) driving 050's empty-gate. A `needs_rebuild` page
  with empty sections may be *awaiting composition* (dartsonline `brands-index`, 0 components) rather
  than *rendered-elsewhere* (robot-hands tools, 1 component) — Pass B2's non-empty gate already routes
  both correctly. Widening the empty-gate predicate would **force-empty** the awaiting-composition
  ones (a real regression; test-guarded).

## What is explicitly NOT in scope

- **Candidate 1 (explicit redesign intent).** A `rebuild:true` / `needs_replan` spec field (001's fix
  step 4) is a deferred FEATURE and a policy call, not required to fix this defect. The redesign route
  after this fix is "empty the composition, then re-plan". Left for the owner.
- **`/bugs_open/050` itself** (deployed-empty classification) — another session's, in flight.
- **`/bugs_open/038`** (every deployed page still rebuilt / content regenerated) — the other half.

## Status

Fix **landed & LIVE on v1.0.1146** (swept into the fleet build by the owner; symbol verified in the
running pod). Discriminating tests committed (`9864fab37`). Open: candidate-1 decision; optional live
re-plan verification. See `README_where_we_are.md` and `SUMMARY_2026-07-21_*`.

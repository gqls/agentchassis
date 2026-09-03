# PLAN 2026-09-03 — bugs_open/437: teach the writer prompt the nested element shape

## The problem, and the correction to the originating brief

`bugs_open/437` was filed 2026-09-02 by the loanzy lane: 119 failed page builds in 14 days
across six sites, all `mechanism-flow: steps[N].branches declared array (items: object),
got string`, pages `active` and never deployed for weeks while live pages link them.

> **CORRECTION to the brief, 2026-09-03 — the filed mechanism section offered two
> candidates and the true cause was neither.** It read: *"Either the schema changed under
> the writer, or the writer's prompt never learned the nested shape."* The schema never
> moved, and the prompt did not merely fail to learn the shape — **it actively stated the
> wrong one.** page-content-writer's Output Format exemplar is GENERATED from
> `llm_field_specs`, which carried element field NAMES only, so a nested array-of-objects
> flattened to a scalar and the prompt rendered `"branches": "..."`. The writer copied the
> demonstration; the type gate refused it, correctly, every time. Proven at
> `llm_call_log` `34f25815-42d3-4057-b42a-b8b42189ae7e` (2026-09-02 19:07Z), which holds
> the instruction and the obedient reply in one row. The bug file now carries this
> correction in place.

This also corrects the natural reading of `bugs_closed/260` §5 candidate 4 ("ask the
writer to obey the schema" — dismissed as weakest). That dismissal is about *asking a
model to be careful with no check*. This is *fixing a false statement in a generated
artefact*, with the mechanical check unchanged and still in place.

## Design, and why each choice went the way it did

**1. A second, additive projection — not a widening of `extractArrayItemFields`.**
That function still answers the flat question for the render-time key reconciler
(PBP-005); changing its `[]string` return would change what that reconciler reads for no
gain here. `StructuredItemShape` is new and additive.

**2. It lives in `datahelpers`, beside the gate that judges its output.**
`itemsDeclareObject` / `nestedFieldDecls` / `IsEmptyContentValue` are already there for
`ContentTypeViolations`. Putting the renderer in `actions` would have made a third reader
of the items-dialect split in a second package — the drift class this estate keeps paying
for. Now the prompt promises exactly what the gate enforces, from one package, and a test
asserts precisely that (feed the demonstrated shape to `ContentTypeViolations`; it must
pass, with the old string shape as an in-query control that must still fail).

**3. Emit NOTHING unless an element property is itself a collection.**
This is the whole blast-radius story: `[MEASURED 2026-09-03]` exactly **1** live component
qualifies across all three element dialects, so every other component's prompt is
byte-identical. Not an optimisation — it is what makes the change safe to ship without a
fleet-wide prompt churn nobody asked for.

**4. `omitempty` is the deploy-safety mechanism, not a style choice.**
An un-upgraded chassis emits neither key; both new template directives sit inside `{{if}}`
guards; so the Go and DB halves deploy in **either order**. Proven through the real
`RenderPromptTemplate` rather than argued. ⚠ That path runs under text/template's DEFAULT
missingkey (`invalid`), not `missingkey=zero` — same `{{if}}` behaviour, but a bare print
of an absent key would emit a literal `<no value>` into the prompt, so both keys appear
only inside their guards and a test asserts on that string.

**5. Empty stays legal, deliberately.** The notes tell the writer it may omit an optional
structured property or send `[]`, suppressed when the item schema marks it required. A
live page has served five steps with `branches: ""` since 2026-08-15 and
`IsEmptyContentValue` keeps that legal at every depth. A fix that pushed writers off
empties would damage correct pages.

**6. The live contract is DECLARED, not just migrated.** `livespec` key
`workflow.page-content-writer.prompt_item_shape` holds both template sites Min 1/Max 1 and
the pre-437 spelling as **Forbidden**, so a revert is caught by the daily drift auditor
rather than by six sites' builds failing again. Its fragments are the migration's literals
and the Go test's fixture — one spelling, three readers. `PhaseGoSide`, because a Go test
reads it.

## Phasing, and where it actually got to

| phase | state |
|---|---|
| Go helper + spec wiring + tests | **DONE**, `a0044e73b`; INERT until the next chassis roll |
| livespec declaration + waiver | **DONE**, same commit |
| Migration 724 (+ `_ROLLBACK`) | **DONE and APPLIED** 2026-09-03 09:44:42Z; verified at the live row |
| Council round | **SUBMITTED** `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`, verdict pending |
| Records (bug file, PBP-052, LANDMINES, 016b §9, WRONG_CALLS, CONTRIB) | **DONE** |
| Post-roll verification at the artefact | **OWED** — see below |
| 437 candidates 2 and 3 | **OPEN, not this work** |

## What is deliberately NOT in scope

- **Candidate 2**, a repair path for type-mismatch refusals: today's only outcome is fail →
  strike → terminal, and nothing re-plans a section or regenerates a field with the error
  in hand. Closest existing design is `copy_quality_two_stage/DESIGN_2026-08-20`'s narrow
  sibling, still unbuilt and unclaimed; if someone takes it, these should be one piece of
  work.
- **Candidate 3**, the escalation gap: an ACTIVE page, linked from live pages, unbuilt for
  weeks, surfaces nowhere. This is what let advertise.co.uk's page sit quietly.
- **Rebuilding the six sites' stuck pages.** This fix stops new occurrences and rebuilds
  nothing. A `failed` item re-mints on the next reconcile sweep by itself; an
  `[unresolved after 2 attempts]` one is held in the open set and blocks re-minting for
  ever until someone closes it. That is a deliberate, separate state-changing step, and it
  should follow a verified build on one page rather than precede it.

## Verification owed after the roll

1. `git merge-base --is-ancestor a0044e73b <the agent-chassis build-provenance stamp>` —
   per SERVICE, not per fleet.
2. At the prompt: a fresh writer run's `prompt_rendered` shows
   `"branches": [{ "body": "...", "label": "..." }]`.
3. At the page: a previously-failing page builds with `branches` stored as an ARRAY in
   `page_components.content_data` — read the artefact, not the work-item status.
4. Census re-run **with a demand control** (writer runs on mechanism-flow pages in the same
   window), or a zero is indistinguishable from an idle pipeline.
5. **Over-production watch:** exemplars govern, so a writer shown a filled `branches` may
   fill it more often than the source warrants. Census the fill-rate on new mechanism-flow
   sections a few days post-roll. Accepted risk, named in the council submission.

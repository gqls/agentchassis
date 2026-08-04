# 192 — `page-content-writer`'s `select_sections` step is failing broadly since ~2026-08-03 21:00, unrelated to bugs_open/087

**Filed 2026-08-04**, discovered while live-verifying `bugs_open/178`'s fix
(`bugfix_154_work_item_routing_columns` lane). **Not yet diagnosed** — this is
a filing with the evidence gathered incidentally, not a completed root-cause.
**A `090` diagnosis run is owed**; not run yet, flagged here so it isn't lost.

## Symptom

`page-content-writer` orchestrations are failing at `process_sections_loop`
with:

```
step process_sections_loop failed: failed to execute action loop: failed to
get collection at 'sections_for_render.sections_ready': key 'sections_ready'
not found at position 1 in path 'sections_for_render.sections_ready'
```

This is the **exact same error string** as `bugs_open/087`, but NOT the same
cause: 087 is specific to the `page-rebuild` agent, which supplies no
`section_plan` at all, and its own text states plainly *"page-build-handler
... its writer children always carry a real section_plan (26 of 26 recent
runs, all COMPLETED)"* — i.e. the build-handler path was the known-good
control. **This instance is ON the build-handler path.**

## Evidence gathered (incidental, while verifying 178 — not a full diagnosis)

Two orchestrations hit it live on 2026-08-04 ~08:26, dispatched via
`build-dispatch-loop → page-build-handler → page-content-writer` for two
`content_rewrite` work items on `vetcomparison.uk`
(`0883b1aa-d5d6-45ad-a596-df0cc06744ec`, page `guide-cma-compliance`;
`df69efd6-19b7-4788-8fe1-668ea769f3fc`, an unrelated tool page
`tool-gripper-payload-calculator-guide` on a different site — confirming this
is not scoped to one site or one item type).

`select_sections`'s live config (`page-content-writer`, confirmed unchanged
from `bugs_open/087`'s own quote of it):

```json
"select_sections": {
  "action": "extract_fields",
  "config": {"fields": {"sections_ready": [
      "resolved_links.response.link_resolution.sections_ready",
      "input_data.section_plan.sections_ready"]}},
  "next_step": "process_sections_loop",
  "output_field": "sections_for_render"
}
```

On `0883b1aa`, `collected_data.resolved_links` at the time of failure:

```json
{"response": {"link_resolution": {"unresolved": null, "sections_ready": null},
              "resolve_links": {"unresolved": null, "sections_ready": null},
              "input_data": {"site_id": "...", "page_name": "guide-cma-compliance", "page_type": "guide"}},
 "response_status": "complete"}
```

i.e. path 1 (`resolved_links.response.link_resolution.sections_ready`) is
present but explicitly `null`. `ExtractFieldsAction`
(`v3_site_actions.go:4232+`) DOES null-check (`value != nil`) before
accepting a candidate, so on its face this should fall through to path 2 —
and `collected_data.input_data.section_plan.sections_ready` genuinely holds a
full, correct 1-element array at that same point in the SAME row (confirmed
directly: `SELECT collected_data->'input_data'->'section_plan'->'sections_ready'`
on this orchestration returns the real ready section, complete with
`existing_content_html` from the 178 fix, proving THAT part of the pipeline
is healthy). **Why the fallback still doesn't populate `sections_for_render`
is the open question** — `[UNVERIFIED]` whether it's an ordering issue
(select_sections running before `input_data.section_plan` is actually merged
into collected_data, vs. only appearing there later), a second code path,
or something about the map-of-arrays extraction loop not behaving as its
source reads.

**Timing, `[MEASURED]` from `orchestration_states`:** hourly counts of
`process_sections_loop` COMPLETED vs FAILED over the last 3 days show every
run COMPLETED through 2026-08-03 20:00, then FAILED spikes at 21:00 (11),
22:00 (14), 23:00 (12 fail / 1 complete), tapering to 00:00-01:00, quiet
overnight (likely just low traffic, not resolved), then my 2 today at 08:00.
**This predates this session's own chassis roll by many hours** — the fix
this session shipped (178) was pushed as an image at ~20:2x on 08-03 and
only actually deployed by the owner's whole-fleet release the following
morning (pods ~11 min old when checked at 08:2x on 08-04) — so **178's code
cannot be the cause**, and the failure was already live before this session's
image ever ran anywhere. The tree had heavy concurrent activity in the
21:00-23:00 window (many sessions, several image rolls) — no single
suspect commit identified; genuinely not run down.

## Why this matters for other lanes

Roughly half of `page-content-writer` invocations failed outright for a ~4
hour window and it is unclear whether the failure rate has actually dropped
or the quiet overnight period is just low traffic. **Anyone whose work
depends on a content build completing should check this before trusting a
`complete` on `page-content-writer`, and should not assume a `page-build-handler`
COMPLETED status without checking its writer child's own status.**

## Effect on bugs_open/178

**No content was lost** — the failure is upstream of `save_page_sections`
entirely (the workflow never reaches `compile_page`/save), so 178's own
guard-rail concern (silent content loss) does not apply here; this is a loud
failure, not a silent one. But it DOES mean 178's live end-to-end
verification (before/after `content_data` length on a real dispatched item)
could not be completed this session. **What WAS verified**: the
`load_current_section_content` step itself worked exactly as designed —
`section_plan.sections_ready[0].existing_content_html` was populated with
the page's exact current rendered content, matched by slot name, byte for
byte the same live prose. The remaining check (does the writer actually
preserve it end to end) is blocked on this bug, not on 178's own code.

Both parked items (`9e9ec430-ff92-4264-83cc-6072840faad8` still `claimed`,
`18bc832c-c937-4608-9a05-718772d44c88` now `failed` attempt_count=1) are
in a safe, non-terminal-for-long state — `failed` with attempt_count=1 is
not `unresolved`/terminal, so they can retry once this is fixed. Do not
re-dispatch them until 192 is understood; a retry would hit the same wall.

## Fix candidates

Not analysed — this filing is evidence, not a diagnosis. First step for
whoever picks this up: run the `090` trigger
(`./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh`)
naming the mechanism above, since the cause is non-obvious and clearly
cross-cutting (hit two unrelated sites, two unrelated item types, in the
same short window).

## How to verify a fix

Re-dispatch either parked item (or any `content_rewrite`/`needs_content_page`
item) via `build-dispatch-loop` for its site, and confirm the
`page-content-writer` child reaches `compile_page`/`COMPLETED` rather than
failing at `process_sections_loop`. Then, separately, re-run 178's own
verification: assert `page_components.content_data` length for the target
slot grows only by the inserted link anchor, not a wholesale replacement.

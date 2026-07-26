# 087 — `page-rebuild` gives its writer no section plan, and nothing in the writer builds one

**Filed:** 2026-07-26 · **By:** the `bugs_open/068` thread, from 068's own verification run
**Status:** OPEN — diagnosed from a live run, **not fixed**. Config-only fix candidates below.
**Severity:** Medium. `page-rebuild` is the documented single-page rebuild entry point and it cannot
complete a page — but it is dormant (0 runs in the 13 days `orchestration_states` retains, other
than this test). Routine rebuilds go `build-dispatch-loop → page-build-handler`, which works.

## Symptom

Fixing `bugs_open/068` let the rebuild-path writer past `resolve_links`. It now dies one step later:

```
step process_sections_loop failed: failed to execute action loop:
failed to get collection at 'sections_for_render.sections_ready':
key 'sections_ready' not found at position 1 in path 'sections_for_render.sections_ready'
```

Live run, 2026-07-26: correlation `7becd532-6c2b-43ec-b4d4-08a4727ddeb7`
(page-rebuild `fd5854e7-f017-4b28-a6f8-ec5b09d8ef3f`, writer child
`bbcc1186-381c-42ea-8a09-0a35da69bac6`, target finetuning.uk `/about.html`).
`agent_error_log` row at 17:56:02Z, `severity='fatal'`.

## Root cause

`page-content-writer.select_sections` is not a chooser — it is an `extract_fields` over two sources:

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

On the rebuild path **both sources are empty**:

* source 1 is the link resolver's output. `page-rebuild` supplies no `section_plan`, so the resolver
  is called with no `sections` and — by the nil-safe design that `068` fix B deliberately relies on —
  returns an empty `sections_ready`;
* source 2 is the caller's plan, and `page-rebuild`'s dispatch does not map one. Its
  `write_page_content` substep maps exactly: `db_sync, hero_url, logo_url, site_plan, site_record,
  current_page, reviewed_brief, style_collection`.

`extract_fields` then writes `sections_for_render` **without** the key, and the loop over
`sections_for_render.sections_ready` fails.

> **This corrects `068`.** That case rejected fix candidate A ("supply section_plan from
> page-rebuild's dispatch") on the premise that *"this generation of the writer selects its own
> sections … so the rebuild caller CANNOT sensibly supply one"*. The premise is false: `select_sections`
> selects nothing, and the writer has no path to a section plan of its own. Candidate A was the right
> shape all along — it was rejected on an unchecked reading of a step name.

A `section_plan.sections_ready` entry is **not** a section name. It is the rich object
`plan_sections` produces — `{name, status, function, component:{…full definition, html_template,
input_schema…}, llm_fields, llm_field_specs, resolved_data, component_id}`. `pages.sections` (and
`current_page.sections`) is only `["hero-about","content-block-about",…]`. So no input-mapping
one-liner can conjure the plan: something must **run** `plan_sections`.

## Why the routine path is unaffected

`page-build-handler` runs `load_spec_sections → plan_sections → … → call_content_writer`, so its
writer children always carry a real `section_plan` (26 of 26 recent runs, all `COMPLETED`). That is
also how finetuning.uk's about page was actually rebuilt on 2026-07-24 16:33 — all 7 of its
`page_components` rows were written in the same second by the build-handler path, an hour after the
`page-rebuild` attempt had failed.

## Fix candidates

- **A — give `page-rebuild` a `plan_sections` step (config-only, live immediately).** Insert it into
  `build_pages_loop`'s sub-workflow before `write_page_content`, output `section_plan`, and add
  `"section_plan": "section_plan"` to the writer's `input_mapping`. Open questions to settle first,
  both cheap: `plan_sections` config on the handler reads `sections` from `spec_sections.sections`
  (the site_plan spec) and takes a `work_item_id` — the rebuild flow has `current_page.sections` and
  **no work item**, so check whether the action tolerates an absent `work_item_id` before wiring it.
- **B — have `page-rebuild` call `page-build-handler` per page** instead of the writer directly.
  Maximum reuse, no duplicated planning machinery to drift, but the handler owns its own
  deploy/status steps, which overlap the rebuild loop's `assemble_page → deploy_page →
  save_sections → update_page_status`; those would have to go.
- **C — retire `page-rebuild`** and route single-page rebuilds through the build pipeline
  (`needs_rebuild` + dispatcher) or the assemble-only `049b` page-rerender. It has not run once in
  the retention window; the two live paths already cover the need. Cheapest, and it removes a
  duplicated pipeline — but it deletes a documented entry point other threads have scripted against
  (`about_page_commercial/p1_trigger_rebuild.sh`).

**A is the smallest change that makes the documented path work; C is the one that removes the class.**
Either wants an owner call, because B and C change what other lanes' scripts do.

## Verification when fixed

Re-arm one page (`UPDATE pages SET build_status='needs_rebuild' WHERE id=…`) and dispatch
`page-rebuild` for its site; the writer child must reach `compile_page`, and the page must deploy
with its components rewritten. Assert the branch, not the happy path: the child's
`initial_request_data->'input_data'` must be the rebuild shape and the run must not be `FAILED`.
Restore `build_status` afterwards if the run does not complete — this thread left finetuning.uk
`/about` back on `deployed`, its 07-24 content and the `about-commercial-block` intact.

**Related:** `bugs_closed/068` (the contract fix that exposed this — same rebuild path, one step
earlier), `bugs_open/086` (the dropped `error_step`; had it been carried, this failure would still be
fatal — `process_sections_loop` declares no handler).

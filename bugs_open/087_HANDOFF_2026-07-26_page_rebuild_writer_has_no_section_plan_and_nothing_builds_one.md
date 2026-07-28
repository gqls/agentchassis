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

---

## 2026-07-27 — both blocking questions ANSWERED, and candidate A is APPLIED

Picked up by the bug-sweep thread. `who-owns.py 087` reports no owning
workstream, and no session had touched the file since 07-26 19:25.

### The two "cheap open questions" the case named as blockers — neither blocks

Both settled by reading `platform/orchestration/actions/plan_sections_action.go`,
not by trying it:

**Q1 — does `plan_sections` tolerate an absent `work_item_id`? YES.**
`:50-52` — `Required` is `["site_id"]` **alone**; `work_item_id` sits in
`Optional`. `createDeferredItems` (`:1824-1830`) guards
`if parentWorkItemID != ""`, so an absent one leaves `parentID` nil and deferred
items simply get no parent. The rebuild flow has no work item and does not need
one.

**Q2 — where does `sections` come from on the rebuild path? `current_page.sections`,
which is the shape the action documents for ITSELF.** The case treated this as a
possible mismatch because the handler passes `spec_sections.sections`. But the
action's own header example (`:22-25`) is:

```json
"sections":  "page_record.sections",
"page_name": "page_record.name"
```

— a **page record's** section list, which is exactly what `current_page.sections`
is. The parser (`:649-664`) accepts `[]interface{}` of strings, `[]string`, or a
JSON string; `["hero-about","content-block-about",…]` is the first case.
`filterSiteLevelSections` then strips any header/footer names, which is the one
real hazard in feeding it a raw page section list and is already handled.

**And the failure mode improves even in the worst case.** If
`current_page.sections` were empty, `plan_sections` returns
`{"sections_ready": [], …, "reason": "no sections to plan"}` (`:673-681`) — the
**key exists**. So `extract_fields` finds it and the loop iterates zero times
instead of dying on `key 'sections_ready' not found`. The fatal error this case
reports cannot recur in that shape.

### Applied — `docs/agent_docs/sql_for_agents/246_page_rebuild_plans_its_sections.sql`

Config-only, live on apply, snapshot taken first. Three targeted `jsonb_set`
operations (an added key and two scalars — never a literal-object write at a
parent path, which would have destroyed the eight-entry `input_mapping`):

1. new `plan_sections` step in `build_pages_loop`'s sub-workflow,
   `next_step: write_page_content`, `output_field: section_plan`;
2. `start_step` moved from `write_page_content` to `plan_sections`;
3. `"section_plan": "section_plan"` added to the writer's `input_mapping`.

Verified inside the transaction before COMMIT:

```
start_step            -> plan_sections
plan_sections         -> write_page_content     (heads the graph, 9 steps total)
input_mapping         -> 9 keys, has_section_plan = t   (the original 8 survived)
```

That third check is the load-bearing one: the writer's mapping had eight entries
and a careless write at the `input_mapping` path would have replaced all of them,
breaking the rebuild path in a new and less obvious way.

### ⚠️ NOT YET VERIFIED LIVE — the acceptance test needs a real dispatch

Nothing above proves the fix WORKS; it proves the config is shaped correctly. The
case's own test still stands and is outward-facing, so it has not been run:
re-arm one page (`UPDATE pages SET build_status='needs_rebuild'`), dispatch
`page-rebuild` for its site, and require the writer child to reach `compile_page`
with the page deployed and its components rewritten. **Assert the branch, not the
happy path** — the child's `initial_request_data->'input_data'` must be the
rebuild shape. Restore `build_status` if the run does not complete.

Until that runs, this is a **config change that should work**, not a fixed bug —
so 087 stays OPEN.

**Candidates B and C are untouched** and still want an owner call: B routes
rebuilds through `page-build-handler`, C retires `page-rebuild` altogether. C is
the one that removes the class, and the case's argument for it is strong — the
path has 0 runs in the 13-day retention window and two live paths already cover
the need.

---

## ✅ CANDIDATE A VERIFIED LIVE 2026-07-28 — and it exposed a worse defect behind it

Acceptance test run on the owner's chosen target: `finetuning.uk`,
`ai-agent-roi-estimator` (the site's only armed page, so blast radius was exactly
one). Correlation `298b5543-267f-464d-b8ad-d208a5a0f0d0`, 07:03–07:10 UTC,
chassis `v1.0.1180`.

**The fix works.** The writer child received a fully populated plan:

```
5de26d35  (page-content-writer, child of the rebuild)   sections_ready = 4
```

Four sections planned, matching the page's four `sections`. The writer then ran
`process_sections_loop_iter_0 … iter_1 …` and **COMPLETED** — the step that
previously died with `key 'sections_ready' not found` now iterates normally. The
run continued through `review_page_content` and reached `deploy_page`.

Migration 246 also survived the `v1.0.1180` roll (`start_step=plan_sections`,
step present, `input_mapping` still carries `section_plan`), so the config is not
re-seed-fragile.

### The run still FAILED — two steps later, and correctly

```
step build_pages_loop_iter_0_save_sections failed: failed to execute action
save_page_sections: page ai-agent-roi-estimator is rebuild_policy=owned
(tool/widget-owned): a generic section save would clobber it. Use
apply_section_edit for targeted edits or the tool pipeline for rebuilds.
Refusing to o[verwrite]
```

**That is a guard working, not a regression.** The page is `rebuild_policy=owned`
and `save_page_sections` refused to overwrite it. 087's own fix candidate A is
unaffected — the failure is downstream of everything this case is about, and the
writer had already done its job.

It does mean this target could never have completed end to end, so
`update_page_status` never ran and the page is still `needs_rebuild`. **A cleaner
future test picks a page that is NOT `rebuild_policy=owned`** — check that column
before choosing.

### ⚠️ It also created a live orphaned page — filed as `bugs_open/125`

`deploy_page` ran BEFORE the guard fired, and wrote to the wrong path:

| URL | before (07:00) | after (07:12) |
|---|---|---|
| `/ai-agent-roi-estimator.html` | 404 | **200, 29,521 b — new orphan** |
| `/tools/ai-agent-roi-estimator.html` | 200, 35,129 b | 200, 35,129 b, byte-identical |

`resolveFilePath` (`git_deployer_actions.go:414-445`) derives the file path from
`slug`/`name`/`page_name`/`filename`/`id` and **never consults `url`** — which
the page object carries, and which says `/tools/ai-agent-roi-estimator.html`.
**280 of 431 pages (65%) would deploy to the wrong path.**

**This is the interesting part of the whole exercise.** 087 was masking 125: the
rebuild never reached `deploy_page`, so a resolver that has been wrong all along
could not show it. Fixing one defect made the estate exactly one working rebuild
away from publishing duplicates at scale — which is a strong argument for the
case's own **candidate C (retire `page-rebuild`)**, and worth weighing before
anyone routes more traffic down this path.

**Cleanup owed:** the orphan at `/ai-agent-roi-estimator.html` needs removing and
this thread could not do it — no credentialed access to `github.com/gqls/sites`,
and the git adapter's only deletion verb is the unimplemented `delete_repo`. See
`bugs_open/125` §"Damage done".

### Status

**087's fix is verified and stays applied.** The case stays **OPEN** only because
its own stated acceptance bar ("the page must deploy with its components
rewritten") cannot be met on this target — blocked by the `rebuild_policy=owned`
guard, and now by 125. Re-test on a non-owned page once 125 is fixed and the two
together will close it.

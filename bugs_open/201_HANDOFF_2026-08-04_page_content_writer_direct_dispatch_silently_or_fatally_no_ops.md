# 201 — `page-content-writer` dispatched directly by `build-dispatch-loop` either hard-fails or silently no-ops on an already-built page

**Filed 2026-08-04** by the `bugfix_091_workitem_conflict_refresh`/`184` lane,
found while verifying `bugs_open/184`'s auto-repair step. **OPEN, unowned.**

## Why this is not just 184's problem

`page-content-writer` is the declared `HandlerAgent` for **three** discovery
checks that route work items at it *directly* (not via `page-build-handler`,
`page-rebuild`, `pageflow-builder`, or `site-work-orchestrator` — the only four
callers `bugs_closed/087` audited and fixed):
`check_literal_markdown.go:245`, `check_placeholder_contact.go:125`,
`check_component_standards.go:477`. All three are dispatched the same way —
`build-dispatch-loop`'s generic `call_handler` step, whose `input_mapping`
(`site_id, domain, work_item_id, item_type, spec, current_page,
refresh_site_components?`) **never includes `section_plan`.**

`bugs_closed/087` (closed **today**, 2026-08-04, migration
`309_page_content_writer_plans_its_own_sections.sql`) gave the writer a
self-plan fallback for exactly this "no plan supplied" case — and its own
closing note says so explicitly:

> "Not proven here, and not claimed: the *falsy* branch. The rebuild path
> supplies a plan, so it exercises the truthy side. The self-plan branch needs
> its own dispatch via `pageflow-builder` / `site-work-orchestrator`."

**This bug is that untested branch, hit for the first time — by a fifth
caller 087 never enumerated: `build-dispatch-loop` calling
`page-content-writer` directly.**

## Symptom 1 — hard fail, 11 of 11 attempts, same page each time

Dispatched `184`'s `literal_markdown` items (12 rows, 3 sites) through the
normal queue path (`claim_work_item` → `spawn_agent`/`call_agent` with
`handler_agent='page-content-writer'`). 11 of 12 ended `status='failed'`
(`attempt_count=3`, retries exhausted), all with the identical error:

```
step fail_no_ready_sections failed: failed to execute action fail_workflow:
workflow failed: page-content-writer planned its own sections and none are
ready — no page can be written. Check the page's sections list and the
components' input_schema data requirements; see section_plan.reason and
bugs_open/087. (code: CHILD_ORCHESTRATION_FAILED)
```

Root cause, read from `bugs_closed/087`'s own description of candidate D: no
`section_plan` in the call → `check_section_plan` falsy →
`plan_sections` runs → returns `ready_count: 0` for a page whose sections are
already built (there is nothing left to *plan*, only to *edit*) →
`check_planned_sections` falsy → `fail_no_ready_sections` → hard fail. This is
not item-type-specific and not retry-recoverable — the 12th item
(mortgagecalculator, `attempt_count=1`) will fail identically on its remaining
two attempts; it is on the same code path as the 11 that already exhausted
theirs.

Failing correlation ids (site → dispatch → child writer failure), from
`orchestration_states.processing_history`/`error`: webdesign.co.uk items
`efaa39a2…`, `c2f516c9…`, `e1e488ea…`, `c61389d6…`, `dd902041…`, `c8461dd2…`,
`6252cb77…`, `6361240f…`, `99214c3f…`, `cd99a656…` (all `site_work_items.id`,
`item_type='literal_markdown'`) — dispatch orchestration
`eb3732c9-7663-47b3-8526-3ae0643548fd`. mortgagecalculator item
`dad119c9-de2c-456d-9177-455a38df0ce4`, dispatch
`446fe9bb-401d-4cf1-a838-d749fde11af3`.

## Symptom 2 — the one item that reached `complete` silently wrote nothing

gaswholesalers.com's one item (`d14f66b1-cced-425e-b18c-bb6e3c4e36a9`,
dispatch `96c323a0-a198-4e96-8b11-2a5f43a0fc4b`) ended `status='complete'`,
`result` has no `error` key — looks like a genuine success. **It is not.**

```sql
SELECT slot_name, updated_at, length(content_data::text)
FROM page_components WHERE page_id='9eed738b-30da-47a3-ae39-6b82299a7098';
-- pricing | 2026-08-03 22:35:17.349534+00 | 1450   ← yesterday, before this dispatch
```

The `pricing` slot's `content_data` **still contains** the exact string the
work item was filed for (`**Recommended next steps:**`), and its
`updated_at` predates today's dispatch entirely — nothing wrote to it. Every
component on the page is stamped `2026-08-03 22:35:17`, i.e. this dispatch
touched **none** of them. The writer's workflow reached `complete` without
performing (or at least without persisting) the edit the work item asked for.
This is the `a-complete-work-item-is-not-a-repaired-artefact` pattern,
confirmed live, not inferred.

**Net result: 0 of 12 real findings were repaired by this dispatch.** 11
hard-failed for a stated, code-cited reason; 1 silently no-op'd for a reason
not yet diagnosed (worth its own look: does `build_render_context`'s
self-plan branch route a *different* handler path than a `spec`-only surgical
edit needs — the writer may have "succeeded" at planning and regenerating
sections it chose itself, unrelated to the `pricing` slot the check meant to
target, and never touched that slot at all).

## Why this also concerns `check_placeholder_contact` and `check_component_standards`

`check_placeholder_contact` has **never had an item reach `complete` or
`failed`** — `SELECT status, count(*) FROM site_work_items WHERE
item_type='placeholder_contact' GROUP BY 1` → 2 rows, both still `triaged`,
2026-08-04. The "precedent" `184`'s progress notes cited ("the
`check_placeholder_contact` precedent") for trusting `page-content-writer` as
an auto-repair handler was **never actually exercised in production** — it is
a routing configuration that has sat untested. `check_component_standards`'s
`page-content-writer` route (`:477`) is equally unverified. Symptom 1's
hard-fail is not `literal_markdown`-specific; any of these will hit it the
moment a real item is dispatched, on any page whose sections are already
built (i.e. essentially every existing page — freshly-planned "ready"
sections are for pages under construction).

## Fix candidates, not yet evaluated

1. **Give `build-dispatch-loop`'s handler dispatch a `section_plan` for
   already-built pages**, e.g. load the page's existing rendered section list
   and shape it the way `page-build-handler`'s `plan_sections` output is
   shaped, so `check_section_plan` takes the truthy branch. Non-trivial: the
   object 087 describes (`{name, status, function, component:{…},
   llm_fields, …}`) is richer than a bare section-name list, so this is not a
   one-line input_mapping change.
2. **Route these three checks through a different handler** built for scoped
   single-slot edits — `check_component_standards.go:477`'s neighbour steps
   already use `component-template-fixer` for narrower jobs; whether one of
   those (or `page-build-handler`, which 087 confirms handles the self-plan
   case correctly via its own gate on `section_plan.ready_count > 0` *before*
   spawning the writer) fits needs checking against what these three checks'
   `spec` shape actually carries.
3. **Diagnose symptom 2 separately** — a silent no-op is worse than a loud
   fail (it retracts nothing, and a check that trusts `complete` would mark
   the finding resolved without repair). Needs its own read of what
   `build_render_context`'s self-plan branch actually persists and why.

## How to verify a fix

Re-arm one of the 11 failed `literal_markdown` items (reset `status`,
`attempt_count`, `claimed_by`, `error`) or dispatch fresh via `184`'s own
verification query, and require the child writer to reach `compile_page` /
`save_sections` **and** the specific slot's `content_data` to actually change
(`updated_at` moves, string is gone) — not just the work item to reach
`complete`. Artefact-level check after: curl the page, per `bugs_open/097`
(`content_data` ≠ `rendered_html`).

## Related

`bugs_closed/087` (the closed bug whose self-plan fix this is the untested
branch of — this is not a regression in 087, it is 087's own stated gap,
now confirmed hit). `bugs_open/184` (the case that surfaced this — 184 cannot
close via its planned auto-repair route until this is fixed or 184 is
re-routed to a different handler).

---

## HANDOFF 2026-08-05 — start here, in this order. Nothing below needs re-deriving.

Written because this session's context is large; picking this up fresh should
not require re-reading the investigation above line-by-line. State is stable —
chassis rolled to `v1.0.1252` on 2026-08-05 09:10 in the interim, but `git log`
since this bug's filing commit (`49e8e3048..HEAD`) touches nothing in
`page-content-writer`, `section_plan`, `build-dispatch-loop`, or
`plan_sections_action.go` — this finding is unaffected and unaddressed.

### 1. The lead that should shape the fix: `check_sectionless_pages.go`'s own header already names this exact mistake

```
// HandlerAgent is page-build-handler (NOT page-content-writer): the build
// handler is the workflow that runs load_spec_sections -> plan_sections,
// which is where the sibling fallback lives. Routing straight to the writer
// would bypass it.
```

That comment is about a different check, but it states the general rule our
three checks (`check_literal_markdown`, `check_placeholder_contact`,
`check_component_standards:477`) violate: **`page-content-writer` is not
meant to be dispatched directly; `page-build-handler` is the wrapper that
plans first.** `page-build-handler`'s own step list (confirmed live,
`agent_definitions.type='page-build-handler'`) includes
`load_spec_sections`, `plan_sections`, `check_has_ready_sections`, **and**
`load_existing_content` / `load_current_section_content` /
`check_content_produced` — names that suggest it already has a path for
*editing* an existing page's content, not only building a fresh one. This is
the single most promising unopened lead and the first thing to read:
`platform/orchestration/actions/` — find `page-build-handler`'s
`load_spec_sections` and `check_has_ready_sections` actions and work out
whether, for a page whose sections are already built, it returns
`ready_count > 0` (unlike `page-content-writer`'s self-plan branch, which
returned 0 for exactly this case on 2026-08-04's live dispatch). If yes,
**candidate 2 (re-route these three checks' `HandlerAgent` to
`page-build-handler`) is very likely correct** and candidate 1 (teach
`build-dispatch-loop` to synthesise a `section_plan`) is probably the wrong
layer to fix this at.

### 2. The other structural fact worth carrying forward: `mark_complete` trusts the handler blindly

Live `build-dispatch-loop` config (`process_item` step, confirmed
2026-08-05):

```
"mark_complete": {
  "action": "complete_work_item",
  "config": { "result": "handler_result", "work_item_id": "current_item.id" },
  ...
}
```

No check that `handler_result` reflects an actual write — this is the
mechanical reason Symptom 2 (the gaswholesalers item marked `complete` with
`content_data` untouched) was possible, structurally, for ANY handler this
loop calls, not just `page-content-writer`. Worth asking whether
`complete_work_item` (or its caller) should require some positive signal
(e.g. a `sections_written` count) before trusting a handler's silence-shaped
success — but do not fix this before fix-1, or a test against the current
broken routing will look like it "worked" (reached `complete`) when it still
didn't write anything.

### 3. What NOT to redo

- Do not re-dispatch any of the 12 `literal_markdown` items to test a theory
  — 11 are `status='failed'` with `attempt_count=3` (exhausted); the 12th
  (`dad119c9…`, mortgagecalculator) is `triaged` with 2 attempts left and
  will fail identically if dispatched through the *same* broken path. Reset
  `status`/`attempt_count`/`claimed_by`/`error` only once a real fix is ready
  to test, and test against ONE item first, not all 12.
- Do not re-verify the failure mechanism — the exact error text, the
  `bugs_closed/087` self-plan-branch citation, and the artefact-level proof
  that the "complete" item wrote nothing are all confirmed live and cited
  above with correlation ids, page ids, and timestamps. Re-running the same
  checks would reproduce the same numbers.
- Do not touch `184`'s detection half — `check_literal_markdown` itself is
  correct, live, and finding real defects (10 genuine findings on
  webdesign.co.uk alone, not a false-positive artefact of the check). Only
  the repair leg is broken.

### 4. Decision this needs, once the read in step 1 is done

If `page-build-handler` already handles the "already-built page, one slot
needs a text edit" case correctly: re-route the three checks'
`HandlerAgent` and re-verify with one item, artefact-level (per
`bugs_open/097` — check `content_data`/`rendered_html`, not `status`). If it
does NOT (e.g. it also assumes it's building a page from a site-plan spec
the check's `spec` shape doesn't carry), this needs an actual new/adapted
handler or a `section_plan`-synthesis step at the dispatch seam (candidate
1) — and that is a bigger, owner-worthy design call, similar in shape to how
`bugs_closed/087` itself flagged candidates B/C as wanting one.

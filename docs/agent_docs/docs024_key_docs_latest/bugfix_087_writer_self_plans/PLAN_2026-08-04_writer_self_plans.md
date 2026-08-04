# PLAN — `bugs_open/087`: the writer builds its own section plan when a caller supplies none

**Started** 2026-08-04 · **Bug** `bugs_open/087_HANDOFF_2026-07-26_page_rebuild_writer_has_no_section_plan_and_nothing_builds_one.md`

## What 087 is, and what was already done to it

`page-content-writer.select_sections` is an `extract_fields` that reads `sections_ready`
from a fallback list of paths, all rooted in either the link resolver's output or
`input_data.section_plan`. A caller that dispatches the writer **without** a
`section_plan` produces no `sections_ready`, and the next step —
`process_sections_loop`, a `loop` over `sections_for_render.sections_ready` — dies:

```
key 'sections_ready' not found at position 1 in path 'sections_for_render.sections_ready'
```

A `section_plan` entry is not a section *name*; it is the rich object `plan_sections`
produces. No input-mapping one-liner can conjure one — something must **run**
`plan_sections`.

- **2026-07-27** — the bug-sweep thread applied **candidate A** as
  `sql_for_agents/246_page_rebuild_plans_its_sections.sql`: a `plan_sections` step
  inside `page-rebuild`'s `build_pages_loop`, `start_step` moved onto it, and
  `"section_plan": "section_plan"` added to the writer's `input_mapping`.
- **2026-07-28** — verified live. The writer child got `sections_ready = 4` and
  `process_sections_loop` iterated. The run still failed two steps later at
  `save_page_sections` because the chosen target was `rebuild_policy=owned` — a guard
  working, not a regression — and it exposed `bugs_open/125` (deploy path derived from
  `name`, ignoring `pages.url`), which is now **CLOSED, fixed and live on v1.0.1217**.
- 087 stayed **OPEN** because its own acceptance bar ("the page must deploy with its
  components rewritten") had never been met on a page it could legitimately rewrite.

## The measurement that reframes the bug (2026-08-04, this lane)

Candidate A fixed **one caller**. There are four live agent definitions that reference
`page-content-writer`, and I read each one's live config today:

| caller | runs `plan_sections`? | maps `section_plan` to the writer? | state |
|---|---|---|---|
| `page-build-handler` | yes | yes | OK — and it gates on `section_plan.ready_count > 0` before it will even spawn the writer |
| `page-rebuild` | yes (mig 246) | yes | OK — fix still live, re-verified today |
| **`pageflow-builder`** | **no** | **no** | **broken — `build_pages_loop.start_step` is `write_page_content`, no planning step anywhere** |
| **`site-work-orchestrator`** | **no** | **no** | **broken — same shape, `current_page` comes from `current_item.spec`** |

Both are reachable: `scripts/initial_messages/170_work_item_flow_build/075d_simple_maintain_trigger.sh`
dispatches `site-work-orchestrator` directly, and `pageflow-builder` is the only value
ever recorded in `site_specs.recommended_builder` (1,216 rows as of 2026-08-02).

So **087's title is literally true for half the callers**: the writer has no section
plan and nothing builds one. Repeating candidate A twice more would mean **four**
hand-maintained copies of the same planning step in four workflows — the drift class
this estate keeps filing bugs about (five disagreeing deploy-path derivations in
`bugs_closed/125`; two hand-maintained council rosters in CLAUDE.md).

## The fix — candidate D: the writer self-plans

Make the bad state unrepresentable at the only place it manifests. The writer already
has everything `plan_sections` needs in scope (`input_data.site_record.site_id`,
`input_data.current_page.sections`, `input_data.current_page.name`), so it can build
the plan its caller did not supply. **No caller, present or future, can get this
wrong.**

New chain (four new steps, one rewired scalar, one appended fallback path):

```
build_render_context
      ↓ next_step  (was: resolve_links)
check_section_plan            conditional_branch  "input_data.section_plan.sections_ready"
      ├── truthy → resolve_links                  caller supplied a usable plan: kept verbatim
      └── falsy  → plan_sections                  output_field: section_plan
                        ↓
                  check_planned_sections   conditional_branch  "section_plan.ready_count > 0"
                        ├── truthy → resolve_links
                        └── falsy  → fail_no_ready_sections   fail_workflow
```

plus `select_sections` gains a **fourth** fallback path, `section_plan.sections_ready`
(unprefixed), which is the only one that can see a writer-local plan.

### Why each piece, and what would break without it

- **The conditional, not an unconditional plan.** `plan_sections` calls
  `createDeferredItems`; running it when a caller already planned would double-file
  deferred work items and discard the caller's (possibly richer, possibly
  content-attached) plan. The conditional means the 100% of live traffic that comes
  from `page-build-handler` takes the truthy branch and its behaviour is **byte
  identical** — one extra ~ms condition evaluation.
- **`check_planned_sections`, the anti-empty-page guard.** With 0 ready sections,
  `plan_sections` returns `{"sections_ready": [], "ready_count": 0, "reason": "no
  sections to plan"}` — the **key exists**, so `select_sections` succeeds, the loop
  iterates zero times, and `compile_page` produces an **empty page** which the caller
  then deploys over a real one. `page-build-handler` already refuses this case
  (`check_has_ready_sections` → `mark_no_ready_sections`); the writer had no
  equivalent. Failing loudly is the same judgement, made in the one place every caller
  passes through.
- **The fourth `select_sections` path.** `ExtractFieldsAction` tries each configured
  path directly and then with `input_data.` **prefixed** — never stripped
  (`v3_site_actions.go:4284-4305`). Paths 2 and 3 are already written with the prefix,
  so neither can reach a top-level `section_plan`. Appended, not inserted, so it
  composes with the 192 lane's owed post-roll cleanup of their path 3.

### Verified in code before writing a line of SQL

Every one of these was read, not assumed:

- `conditional_branch_action.go:305-315` — a bare dotted path with no operator **is** a
  truthiness check; `valueIsTruthy` (`:527-551`) is false for nil, empty string, empty
  array, empty map, 0 and false. So "absent or empty" needs no new grammar and no Go
  change.
- `resolveFieldValue` Strategy 5 (`:396-411`) recursively searches under the base path,
  so a **192-wrapped** caller plan still reads as truthy — and it cannot fire when no
  plan was supplied at all, because the base object does not exist.
- `datahelpers.ExtractNestedField` (`data_helpers.go:1199-1234`) is strict traversal
  plus a `.response` unwrap, so the new unprefixed path resolves a top-level
  `section_plan` exactly and nothing else.
- `plan_sections_action.go:49-51` — `Required` is `["site_id"]` alone; `work_item_id`
  is optional and `createDeferredItems` guards on it. **Zero LLM calls in the file** —
  it is DB-only, so the self-plan branch costs no credits.
- `registry.go:65/71` — `conditional_branch` is the canonical name; `conditional` is a
  deprecated alias to the same handler. `fail_workflow` is registered at `:35` and is
  in live use (`report-builder.fail_out`).

## Deliberately NOT done, and why

- **Not repointing `resolve_links`' `sections?` mapping.** On the self-plan branch the
  link resolver is handed nothing, because that mapping is
  `input_data.section_plan.sections_ready` and **both** of `FindByPath`'s prefix
  fallbacks guard `i == 0` (`content_search.go:70-95`) — `input_data` resolves at
  position 0, so the miss at position 1 gets no fallback. Repointing it to the
  unprefixed `section_plan.sections_ready` would fix that. It is **owed, not done**,
  for three reasons: seed 308 pins that mapping deliberately and the 192 lane is
  mid-flight on the Go half; the resulting degradation (no internal CTA resolution) is
  exactly the cost the estate has **already accepted fleet-wide** for the pre-roll
  window, so the self-plan branch would be no worse than every build running today; and
  the repoint trades an **exact hit** for a three-deep `FindByPath` fallback on the one
  branch that currently works — not a trade to make inside the change I am using to
  prove 087. Follow-up recorded in NOTES with the exact one-line edit and its post-roll
  test.
- **Not patching `pageflow-builder` / `site-work-orchestrator` individually** (A′) —
  that is the four-copies outcome D exists to avoid.
- **Not candidates B or C.** B routes rebuilds through `page-build-handler`; C retires
  `page-rebuild` altogether. Both change what other lanes' scripts do and the case says
  they want an **owner call**. D decouples them: the writer is now safe regardless, so C
  can proceed on its own merits without blocking 087.
- **Not removing mig 246's now-redundant rebuild-side plan step.** It is live-verified
  and harmless; removing it is pure risk for no behaviour change.

## Scope and gates

- **Config-only. Live on apply, no image roll.** Both actions it uses are registered in
  the running binary and exercised live today. That matters right now: the fleet is
  mid-flight on the 192 lane's Go fix, and this change does not queue behind it.
- **Council gate does not apply.** Its scope is `platform/`, `internal/`, `pkg/`
  (owner ruling 2026-07-17); `097_TRIGGER_council_review_v1.sh:127` refuses a submission
  touching none of them, client-side, before spending credits. This change is a SQL seed
  plus docs. Stated in the commit rather than left implicit.
- **Not architecture-scope** under the 2026-07-29 narrowing: no new key, namespace or
  behaviour on any shared *action* — two already-registered actions rewired inside one
  agent's workflow. What changes is `page-content-writer`'s own input contract
  (`section_plan`: de-facto-required → optional-with-derivation), its consumers are
  enumerable (four, all measured above) and all four are named in the seed header.

## Acceptance — what closes 087, and what would be over-claiming

**Closes it:** dispatch `page-rebuild` at a `rebuild_policy='generic'` page and assert
(1) the writer child's `initial_request_data->'input_data'` is the rebuild shape and the
run is not FAILED; (2) it passes `process_sections_loop` and reaches `compile_page`;
(3) `save_sections` and `update_page_status` complete — `page_components` rewritten,
`build_status` → `deployed`; (4) **the deploy path** matches `pages.url` and no new file
appears at the name-derived path — the 125 interaction that blocked closure.

**Would be over-claiming:** that the whole no-plan class is proven from that one run —
the rebuild path supplies a plan, so it exercises the **truthy** branch. The falsy
branch needs its own dispatch and is recorded separately. Also over-claiming: any fleet
failure-*rate* figure (traffic is far too low), and anything about 192's `required`
opt-in, which is **inert until the next roll**.

# NOTES — bug 025 content_direction wire-up

Append-only, newest at the bottom. Missteps are the point.

## 2026-07-21 — verification of the handoff diagnosis (all held)

- `\d`-style check: `pages.content_direction` jsonb exists, comment is the misleading
  one; **0 of 301 pages** have any value (`count(content_direction)=0`). So deletion
  would have been lossless — but owner chose to wire it up.
- Both loaders confirmed by reading: `get_pages_to_build_actions.go` and
  `load_page_record_action.go` did not SELECT the column.
- All Go `content_direction` refs are the site-level `site_specs` aspect (grep).
- Live `page-content-writer` prompt: 3 `content_direction` refs, all
  `.site_specs.specs.content_direction.*`; the per-section `content_brief` block is
  present and dereferences `.purpose/.tone_direction/.section_guidance`.

## 2026-07-21 — the misstep I nearly made, and what caught it

The handoff's model (and seed `055_page_build_handler.sql:85`) says the writer's
`current_page` = `input_data.spec`. If I'd trusted that, I'd have "verified" the fix
by only wiring `get_pages_to_build` (which fills the spec) and skipped
`load_page_record`. **The LIVE page-build-handler maps `current_page = page_record`**
(fresh `load_page_record` output) — so on the actual rebuild path the spec's
content_direction is never read into current_page. Caught by dumping the live
`page-build-handler` default_config instead of reading the seed. Recorded as a
correction in the bug file and PLAN. Lesson (already in 016b): seeds are migration
history, not live truth — this bit in the *inverse* direction from usual.

Path table (live agent_definitions):
- page-build-handler → `page_record` (load_page_record)   ← rebuild acceptance test
- pageflow-builder / page-rebuild → loop var over get_pages_to_build
- site-work-orchestrator → `current_item.spec` (= json.Marshal(page) from get_pages_to_build)
So **both** loaders genuinely need the column. Wired both.

## 2026-07-21 — template correctness

Renderer is std `text/template` (`datahelpers.RenderPromptTemplate`). Pulled the live
template (post-migration) and parsed+executed it standalone with the same funcMap for
three cases:
- NO_DIRECTION: block absent; `<no value>` count = 8
- WITH_DIRECTION: block present, all 4 sub-fields; `<no value>` count = 8
- PARTIAL (instruction only): only Instruction line; `<no value>` count = 8
The count is CONSTANT across cases → my block contributes zero `<no value>`. The 8 are
from the stub's missing `render_context` fields, not the block. Guards work.

[UNMEASURED so far] end-to-end on a real page after the image roll — the acceptance
test. Go half is inert until then.

## 2026-07-21 — Go half already LIVE (rode a sweep), acceptance test PASSED

- My two Go edits were swept into commit `fe2ba5e52` ("v1.0.1146 - sweep") by a
  concurrent thread, which built+rolled v1.0.1146 fleet-wide. Running pod binary (Jul
  21 12:07) pod-grep: `content_direction::text` ×3, `addContentDirection` ×2, control
  `queryPagesForBuild` ×6. So the Go half is LIVE — no build needed from me. (bugfix-047
  lesson confirmed again: check the pod tag before rebuilding.)
- **Acceptance test on vonc.com/about — PASS.** Set content_direction to weave in the
  verbatim off-theme marker "Quite simply, it began with one stubborn question"
  (baseline 0). Rebuilt via a hand-queued triaged `needs_content_page` work item
  (`61de62d8`) → build-pipeline-trigger → dispatch loop → page-build-handler →
  load_page_record → writer. Post-rebuild: the verbatim sentence in **4 sections**
  (hero-about, content-block-about, game-master-explanation, differentiators),
  page redeployed. Verified against `page_components.rendered_html`, not a status.
- Trigger gotcha worth keeping: `build_status='needs_rebuild'` alone does NOT start a
  build — `build-pipeline-trigger`'s pre_query only fires a site that already has a
  **triaged pipeline=build** work item. Create the work item (or let the
  site-work-orchestrator run); the flag by itself sits inert.
- Work-item lag noticed: the item stayed `claimed` while the page reached
  build_status=`deployed` and sections were saved — the completion stamp lagged the
  actual work. Not this bug's concern (separate completion-integrity matter); the page
  did deploy. [OBSERVED, not chased]
- Cleanup: content_direction reset to NULL; restoration rebuild `f92fe45e` queued as a
  live negative control (marker must clear).

# PLAN — bugs_open/029: tool-suggester writes phantom tool links

**Started 2026-07-21.** Branch `085_debug_and_feature_loops`. Bug file:
`/bugs_open/029_HANDOFF_2026-07-19_tool_suggester_writes_phantom_tool_links.md`
(the "VERIFIED + SHARPENED 2026-07-21" section is the current diagnosis of record).

## The defect, in one line

The URL of a tool page **cannot be constructed** from the tool's function name — it must be
looked up from `pages.url`, and that lookup is only meaningful once the tool page row exists.
`create_tool_cross_link_items.go:142` constructs `/tools/{function}.html` at *suggestion* time
and bakes it into a rewrite instruction + acceptance test, so the woven link 404s.

## Evidence (primary, 2026-07-21 — all re-derivable, see RUNBOOK)

- **0 of 24** `source='tool-suggester'` `content_rewrite` items (leopardess 10, gamesdesign 13,
  ai-agent-orchestration 1) have a constructed URL that resolves to a real `pages.url`.
- Including `tool-process-automation-scorer`, which **was built** — deployed at
  `/tools/process-automation-scorer/index.html`, item points at
  `/tools/tool-process-automation-scorer.html`.
- Deployed-tool URL shape is non-deterministic across build paths (strip-prefix `.html`,
  keep-prefix `.html`, `CanonicalisePage` `/index.html`). The emitter is wrong on all three.
- The live `page-build-handler` maps `rewrite_guidance? → input_data.spec.suggestion`, so the
  writer obeys the fabricated URL; the item's own `acceptance_test` requires it. Not an LLM
  confabulation — deterministic, and the emitter's.

## Decisions & their reasons

- **CORRECTION to the original handoff framing** ("the writer invents a plausible URL"): the
  *emitter* fabricates the URL; the writer obeys it. Recorded in the bug file. This matters
  because it changes the fix layer — you cannot fix this by "teaching the writer", it is not
  the writer's choice.
- **Fix layer = the tool-BUILD success path (candidate 1 made concrete), NOT the emitter and
  NOT the consumer.** Reasons:
  - The emitter runs at suggestion time; a same-orchestration URL lookup finds nothing because
    builds are async — so an emitter-only fix would emit nothing and silently kill cross-linking.
  - A consumption-gate (page-build-handler resolves the tool at claim time) works but needs a
    deferral/retry mechanism for "tool not built yet" and leaves the fabricated URL in the queue.
  - The build path already (a) has the real `pageURL` it just created, and (b) receives
    `related_pages` in the `add_tool` spec (`spec_data: current_suggestion`, verified live), and
    (c) already emits follow-on items (`needs_content_page`, companion guide). Emitting the
    cross-link there removes the race **by construction** and uses the real URL.

## Fix shape (to implement)

1. Extract the `content_rewrite` spec-builder from `create_tool_cross_link_items.go` into a
   shared helper taking `realURL string` (delete the `:142` fabrication).
2. Call it from `deploy_tool_action.go` and `create_tool_component_action.go` after the page
   INSERT, iterating `related_pages` from the incoming `add_tool` spec, resolving each to an
   active non-`tool-` page, using the just-created `pageURL`.
3. Remove the suggester's `create_cross_links` workflow step (migration; 098 added it) so
   nothing emits at suggestion time.
4. Keep the dedup `item_key = tool_crosslink:{function}:{page}:{site}` and the
   terminal-status-derived ON CONFLICT clause (do NOT hardcode the status list — see the
   dedup-index lockstep landmine).

## Residuals (deliberately out of scope)

- **Tool page created (`planned`) but content build never deploys → link still 404s.** That is
  `049` mechanism 2 (planned-but-unbuilt page linked), a broader class. Note it; coordinate.
- **`validate_page_content.go` warns-but-does-not-block on an in-body phantom link** (`:571`).
  Defence-in-depth belongs to the 023/033/049 "detected-but-not-delivered" family, not here.
- **Existing damage** (24 items + woven links on live leopardess/gamesdesign pages) is not
  cleaned by the emitter fix — separate sweep, coordinated with `049`.

## Process

Platform change under `platform/orchestration/actions/` + a config migration → cross-cutting,
changes fleet behaviour. Route through the **council gate** (advisory) before committing, then
build + roll + verify against a live tool suggestion. Commit per task, narrow pathspec.

## Phasing

- **P0** — diagnosis recorded (bug file + these docs). ← done 2026-07-21
- **P1** — implement the shared helper + build-path emit + remove `create_cross_links` step.
- **P2** — council gate submission (rationale + plan), await verdict.
- **P3** — build chassis image, bump tag, roll, verify against a fresh tool suggestion:
  the emitted `content_rewrite` carries a URL equal to the tool page's real `pages.url`.
- **P4** — existing-damage sweep decision (coordinate with 049 owner).

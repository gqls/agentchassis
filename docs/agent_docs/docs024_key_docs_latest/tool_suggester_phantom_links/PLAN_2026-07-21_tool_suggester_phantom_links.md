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

---

## Update 2026-07-26 — P1 done, P2 submitted, config half live

Phasing now reads:

- **P0** — diagnosis recorded. Done 2026-07-21.
- **P1** — shared helper + build-path emit + `create_cross_links` removed. **Done 2026-07-26**
  (Go: `create_tool_cross_link_items.go`, `deploy_tool_action.go`,
  `create_tool_component_action.go`; config: `211_tool_crosslink_emit_at_build.sql`, applied and
  recorded — renumbered from 210 after another thread took that number).
- **P2** — council gate. Submitted 2026-07-26, `SUBMISSION_CORR=745f9dfd-0a08-415b-a0a2-92c96bd30260`.
- **P3** — build + roll + verify against a fresh tool build. **Outstanding.** The Go half is
  inert until then; the config half is live now.
- **P4** — existing-damage sweep (27 items). Unchanged: out of scope here, coordinate with 049.

> **CORRECTION 2026-07-26 to §Residuals — "tool page created but never deployed" is NOT deferred
> to 049 after all; it is gated here.** The original reasoning was that a planned-but-unbuilt page
> is 049's broad class. That holds while the emitter runs at suggestion time, when it has no
> relationship to any build. Once the emitter moves INTO the build path, the same residual becomes
> this emitter's own remaining failure mode, and it reproduces the exact damage this bug is about:
> a live page referencing a tool page that never goes live. The fleet makes it likely rather than
> theoretical — 19 of 33 live `needs_content_page` items are parked in `needs_human_review`.
>
> So `emitToolCrossLinkItems` gates: emit immediately if the tool page is already
> `deployed`/`needs_rebuild`; otherwise attach `depends_on` = the open `needs_content_page` item
> for that page; if there is no open item (or it has failed terminally), emit nothing. This is the
> loader's existing mechanism (`load_work_item_actions.go:562-571`), not new machinery.
>
> **What this costs, stated plainly:** a tool page whose content build never completes now leaves
> cross-link items parked in `triaged` forever rather than writing a dead link. Parked items age
> and may be picked up by the stale-item reaper (`bugs_open/070`). That is the intended direction —
> the alternative is the bug.

Two smaller decisions taken while implementing:

- **The suggestion-time action is kept registered, not deleted.** A workflow naming an
  unregistered action is invalid at runtime (`bugs_closed/017`), and config can be restored from a
  stale backup (this thread already caught `k8s/bk_agent_definitions_backup.sql` being stale on
  07-21). So the action survives as a fail-safe: it resolves the tool to a real page and emits
  nothing when there is none. It can never fabricate again, wherever it is invoked from.
- **`deploy_tool_to_site` emits on its already-deployed early return too**, resolving the URL from
  the page row. That makes re-running the deployer the supported way to backfill cross-links for a
  tool deployed before this fix — useful for P4 — and dedup makes the repeat harmless.

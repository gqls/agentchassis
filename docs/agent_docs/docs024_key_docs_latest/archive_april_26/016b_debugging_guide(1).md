# 016b — Debugging Guide (Vol. 2)

Continues `016_debugging_guide` (closed at v2_56, ~2500 lines). 016 holds the full
back-catalogue of failure patterns and their fixes; consult it for anything not covered
here. New debugging entries go in THIS file from 2026-06-22 onward.

*v1 (2026-06-22): opened as the continuation of 016. Carries forward the orientation, the
durable invariants/heuristics, and the open threads (Parts 2/3/4 of the page
build/re-render work). No new failure patterns yet — the most recent one (interactive page
rebuilt as content) is the last §9 entry in 016 v2_56.*

---

## Orientation (read this first)

The platform builds and maintains multi-page websites with autonomous **agents** — rows
in `agent_definitions` that run as Kubernetes pods (namespace `ai-persona-system`),
communicate over Kafka, and persist run state in Postgres `clients_db`. Each unit of work
(build a page, swap a hero image, re-resolve links) is a row in `site_work_items`, claimed
off `build-dispatch-loop` and routed to a handler agent; the handler usually spawns child
agents (e.g. `page-content-writer`, `page-rerender`) whose state lives in
`orchestration_states` (linked by `parent_orchestration_id`). Deployment is image-tag
based: code ships GitHub → Actions → Backblaze into a new chassis image, then each agent's
`image_tag` is bumped to adopt it; workflow (`default_config`) changes are DB-only and take
effect immediately. The operator runs all SQL / `kubectl` / `git`; the agents do the building.

DB access: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## Durable invariants and heuristics (the lessons that keep paying off)

These are distilled from 016. When an investigation stalls, it is almost always because one
of these was skipped.

**Trust the rendered artefact, not the status.** A work item `complete` — or a green git
commit — is NOT proof the work happened. The silent-completion family (016 §9) recurs:
a result-contract drop replaced output with a stub that still said `status: completed`;
a content-regression guard error was masked by `complete_error` as success; a pod dying
mid-flight left `complete` with a non-empty `error`. Verify against `page_components`
(timestamps + `rendered_html`) and the live page, not the work item.

**`completed_at` is the orchestration END, not the write instant.** The actual page write
happens mid-run inside a child orchestration. Never exclude a candidate work item on a
near-miss timestamp. To find who wrote a page/row, trace `orchestration_states` filtered by
the target id in `collected_data` over the window — the writer is the child whose
`owner_agent_type` + mid-run timing line up. (016: "recurring debugging trap" parts 1–3.)

**A config key read on a different path than it's set is a silent no-op, not an error.**
`output_field` vs `output_fields`; `max_tokens`/`ai_service` shadowing; dispatch
`input_mapping` path mismatches. When a handoff "completes" but the payload isn't there,
compare the producer's stored output to the consumer's received field by exact path before
suspecting generation, compile, or a guard.

**Know who writes `page_components`.** Only `save_page_sections` writes it (DELETE+INSERT),
recording a row in `page_component_history` with `source` set but `source_item_id`
currently NULL on the overwrite path (a traceability gap — see Part 4). `page-content-writer`
regenerates a page from `plan_sections` and does NOT preserve interactive sections absent
from the plan. Interactive game/tool pages store the whole tool as a section's
`rendered_html` (often the hero slot, ~18 KB of `tool-page` markup), not as a separate
library component — so a writer run on such a page silently drops the game.

**0 rows is not decisive.** Check the query/path before concluding absence — a parent-side
JSON path queried against a child row, a name that doesn't match, a wrong timestamp window,
or a column in another schema all return 0 rows. (016 Schema reminders; assumption checklist.)

**Reuse before rebuild; check the schema before SQL.** Look for an existing function/struct
(sibling functions in the same file are the canonical pattern) before writing new code, and
read the table's columns (`information_schema.columns` / `\d`) before composing a query —
tables reported "not found" by `\d` may live in another schema.

**Key schema gotchas (full list in 016):** `site_work_items.error` (not `error_message`);
`site_work_items.pipeline` (not `domain`); `orchestration_states` PK is `orchestration_id`
(no `id` column); back up an agent with `snapshot_agent('<type>','<reason>')` /
`revert_agent('<type>')`, not a hand-rolled `agent_definitions_backup`.

---

## Open threads (carried forward 2026-06-22)

The page build/re-render work is a connected series of fixes to the same failure shape —
work that reports success but doesn't happen, or work dropped/duplicated/clobbered.
Operational detail (apply/deploy/verify, with SQL) is in
`RUNBOOK_gamesdesign_index_rebuild.md`; the running investigation log is
`NOTES_gamesdesign_silent_norebuild.md`.

**Part 1 — result-contract drop (DONE).** The chassis coordinator discarded a child's
workflow result (singular `output_field`, or oversize) and substituted a stub that reported
success → no-op save → work item `complete`. Shipped 2026-06-18 (`result_spec.go` +
`coordinator.go`); verified — gamesdesign `index` rebuilt+deployed 06-19 and re-rendered
cleanly 06-22. This also resolved the long-open "index returns thin content" question (it
was the stub/fallback, not thin generation). Full write-up: 016 §9 "Child workflow result
silently replaced by a stub."

**Part 2 — no-LLM re-render path (VERIFIED for image_landed; remainder pending).** Added
`rerender_page_sections` + a gated pre-pass on `page-rerender`, and repointed
`flag_page_image_rebuild` / `reconcile_section_data` to emit a `page_rerender`-type item
(was `needs_page`). Index Test 1 (06-22) passed: all 5 sections re-rendered, copy preserved,
no LLM, P2.1 wiring confirmed. **Remaining:** P2.4 (real image-landed flow), P2.5
(section_data_resolved), P2.6 (NULL content_data escalation → needs_page), P2.7 (plain
page_rerender backward-compat). Also confirm the live `/index.html` now shows all 5 sections
(system-stats). Runbook Part 2.

**Part 3 / Bug 2 — item_key canonicalization (CODE PREPARED; NOT APPLIED).** `item_key`
prefixes drifted from `item_type` across creators. Confirmed live: the adoption creator keys
BOTH `needs_content_page` and `needs_tool_recreation` as `needs_page:<name>`, so a tool and a
content page of the same name collide on `idx_swi_dedup` and one is dropped. Prepared:
`workItemKey(itemType, target)` builder in `work_items_common.go`; adoption tool item →
`workItemKey("needs_tool_recreation", page.Name)`; content item stays
`workItemKey("needs_page", page.Name)` (Option B — preserves the deliberate doc-029 co-dedup
with planner `needs_page` builds; decision recorded in the runbook P3.1). **Apply after Part 2
verifies** (don't stack unverified changes), then run the P3.4 tests. Runbook Part 3.

**Part 4 — interactive page rebuilt as content (CAUSE CONFIRMED; FIX PENDING).** A
`link_resolution_rebuild` (spec: "preserve the copy; re-resolve links") is handled by
`page-build-handler`, which runs the full `page-content-writer` → regenerates from
`plan_sections` (no knowledge of the tool) → the interactive game is discarded, falling back
to `generic-text-block`. Blast radius: one page (`game-pathfinding`). **Step-config result
(2026-06-22):** `page-build-handler` has NO `item_type` branch — `link_resolution_rebuild`
runs the full writer path, and the tool dies because it isn't in the page spec the plan is
built from (`load_existing_content` runs but unconditionally proceeds to the writer, so it's
not the lever). So the fix is a routing change, not a conditional fix. **Next diagnostic:**
determine how internal links are represented (resolved-at-render → route to the Part 2
`page_rerender` path; baked into prose → a links-only rewrite, noting an `internal-link-resolver`
already exists) — inspect stored `rendered_html`/`content_data` and read
`resolve_internal_links_action.go` + the §5 batch creator. **Fixes:** the chosen route; stamp
`source_item_id` into `page_component_history`; add an interactivity-aware save guard. Then
re-run the tool-recreation for `game-pathfinding`. Full write-up: 016 §9 (last entry) "Interactive
game/tool page rebuilt as plain content"; runbook Part 4.

---

## 9. Specific Failure Patterns (new)

*(New patterns from 2026-06-22 onward go here, in the 016 §9 style: `### <descriptive
title>`, then Symptom / Diagnose / Root cause / Fix, with the SQL and cross-references. The
immediately-prior pattern — the interactive-page clobber — is the final §9 entry in 016
v2_56; start the next one below.)*

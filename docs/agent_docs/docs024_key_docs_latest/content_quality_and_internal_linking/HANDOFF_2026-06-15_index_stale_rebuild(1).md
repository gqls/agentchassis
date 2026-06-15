# HANDOFF — gamesdesign.co.uk linking batch close-out + the `index` stale-rebuild defect (2026-06-15)

This hands off a near-complete internal-linking fix batch plus one unresolved defect (`index` page won't update on rebuild). Written to be self-contained for a fresh chat. Read this top-to-bottom; the live thread is §3 (the `index` defect) — everything before it is context/done-state.

---

## 0. Standing working agreement (carry into the new chat)
- Site: **gamesdesign.co.uk**, site_id **e33263f4-74f8-494f-b191-546845dbbddf** (re-resolve after any teardown: `SELECT id FROM sites WHERE domain='gamesdesign.co.uk'`). Postgres `clients_db`, Go agent-chassis, Kafka saga agents, deploy git→Cloudflare (repo `gqls/sites`).
- Division of labour: the assistant reads files + writes Go/SQL/doc deliverables to `/mnt/user-data/outputs/`; the **user runs all SQL / kubectl / builds / deploys** (assistant has no cluster or DB access, no Go toolchain — validate Go by brace/paren balance only).
- Rules: snapshot before any DB change (`snapshot_agent('<type>','reason')` / `revert_agent`); reuse existing functions/patterns before creating; complexity in Go, workflows thin; **read the actual deployed code before editing** (this caught multiple wrong assumptions this session); don't conclude from partial signals; no "final/perfect/critical"; **trust rendered HTML / DB state over work-item status** (silent-completion is rife here).
- k8s namespaces: `-n ai-persona-system` (chassis, agents), `-n kafka` (cluster `personae-kafka-cluster-...`).
- psql vs shell gotcha (bit us 3×): inside psql set vars with `\set X 'v'` and use `:'X'`; a bare `X='v'` is shell syntax and errors. `psql` is NOT on the desktop — can't resolve ids there. Simplest: inline the literal site_id. No leading/trailing space in shell `SITE_ID="..."` (a space = silent no-op kcat run).

## 1. What the linking batch was (DONE except index)
Goal: eliminate phantom internal links (hero/CTA pointing at non-existent /contact.html, /services.html; "Browse All X" buttons with href=""). Layers, all DEPLOYED + APPLIED + VERIFIED:
- **Step 1 / Layer 1a (SQL + Go):** hero & call-to-action schemas set `on_missing: skip_field`, phantom `fallback`s removed; templates gate each button on `{{if and .cta_text .cta_url}}`. Go: `plan_sections_action.go` `resolve` `case "pages"` no longer fabricates `/<path>.html` (returns nil,false). Verified: both components skip_field/fallbacks_gone/has_and_gate all true.
- **Layer 1b (Go):** header/footer phantoms (built fresh in Go at render). Verified gone (§4 audit: 0 site_component findings).
- **B4/B5 (SQL):** the 3 list components' Browse-All cta_url → `query.section_index_for:<type>` (a queryresolve verb resolving the hub URL from real page relationships); templates gate the anchor. Verified.
- **Step 3 (Go + agent + wiring):** an `internal-link-resolver` sub-agent spawned by `page-content-writer`, augmenting hero/CTA sections' resolved data with real top content-hub URLs at build time. Wiring confirmed LIVE (query D earlier: for_render == plan_count on completed rebuilds).
- **Post-deploy audit check** `check_phantom_internal_links.go`: routing confirmed correct as-is (per-finding pipeline/handler survive `insertWorkItem`); home is **completeness-discovery-agent** (content domain). NOT YET ENABLED (deliberate, observe-only later — see §6).

Chassis image deployed: **v1.0.1062** (internal-link-resolver, page-content-writer, research-agent, page-build-handler all aligned).

## 2. §5 rebuild batch result (21/22 good; games-index fixed; index is the holdout)
- Queued 22 page rebuilds via `site_work_items` (handler_agent=page-build-handler, pipeline=build, status=triaged, spec{page_id,page_name,mode:recreate,suggestion}). 21 completed correctly; verified in deployed HTML: heroes gated (no phantom buttons), Browse-All → real hubs, header/footer clean, unresolved_cta=0 throughout.
- **games-index:** initially failed (claim timeout), re-queued, NOW DONE (rendered games/index.html: 5 cards, Browse All Games → /games/index.html, clean hero).
- **index: NOT FIXED — the live defect. See §3.**
- The §5 batch covered the 21 CONTENT pages (index, the *-index hubs, guide-* detail pages). It did NOT include the tool-* / game-* DETAIL pages — those remain at their 06-06 adoption state (expected; not failures).

## 3. ⭐ THE LIVE DEFECT: `index` rebuild completes but deploys stale content

### Symptom
Every rebuild of `index` (3 production attempts 06-13/14/15 + 1 diagnostic `_diag1` 06-15) shows work-item `status=complete`, the handler orchestration reaches `COMPLETED` (~610–770s), git commits `/index.html` — **but the committed HTML is the OLD 06-06 content** (hero still has phantom `/contact.html` + `/services.html` with button text; list Browse-All buttons still `href=""`). The page's 5 `page_components` rows are **all frozen at created_at=updated_at=2026-06-06 16:59:53** — nothing written despite the "successful" rebuild.

### Theories RAISED and ELIMINATED (do not re-litigate these)
1. ~~Load / dispatcher starvation~~ — user confirms ample capacity, minimal load.
2. ~~Concurrent production deploy cycling pods~~ — user deployed nothing at the retry windows; `kubectl get pods` shows agent-chassis up 2d2h, 0 restarts.
3. ~~page-build-handler claim-lease / duration~~ — index runs ~610-770s and reaches COMPLETED, well inside timeout_seconds=1800; doesn't die partway.
4. ~~Caller-side call_agent timeout~~ — real as a STATUS artifact (the work item's "Claim timed out — handler pod likely died" comes from a caller giving up while the child finished) but NOT the defect: the child finishes and deploys.
5. ~~page_components lock~~ — index's components have locked_at/locked_by/lock_type/lock_expires_at ALL NULL (despite a `trigger_auto_lock_on_deploy` existing). Not locked; save not blocked by a lock.

### LEADING mechanism (coherent, NOT yet 100% confirmed — needs ONE number)
`save_page_sections` (deployed `save_page_sections_action.go`, lines ~223-259) has a **content-regression guard**: it sums stripped-text of existing `build_status='deployed'` components; if that >200 AND new content's stripped text < existing/4, it `return nil, fmt.Errorf("content regression blocked...")` — an ERROR, refusing to overwrite a rich page with a thin one. In `page-build-handler` the `save_sections` step has `error_step: complete_error`, and `complete_error` is a `complete_workflow` (SUCCESS exit). So the guard's error is **laundered into a `complete` work item** (silent-completion, same family as v2_33's `complete_error==complete_workflow`); the deploy then re-renders the unchanged 06-06 components and git commits stale HTML.

**>>> UPDATE 2026-06-15 — THIS GUARD HYPOTHESIS IS NOW FALSIFIED. <<<**
Measured the writer's actual output for `_diag1`'s writer child (472eed7d): the five per-section outputs (`section_output_0..4`) sum to **33,335 stripped chars** (3032+7955+8101+7449+6798) — far ABOVE the guard's 5760 threshold (and above the 23041 existing). The guard CANNOT have fired; the content is abundant, not thin. **The writer worked fine; the defect is downstream of the writer, in how save receives sections.**

### SURVIVING / NOW-LEADING mechanism: metadata-path mismatch (writer output ≠ what save reads)
`save_page_sections` reads its sections from `sections_metadata_field: page_content.response.sections_metadata` (primary path) or an HTML fallback. The writer emits content in `section_output_0..4` / `sections_for_render` — and the writer child's `page_content.response` had **NO keys** (the `jsonb_object_keys(... ->'page_content'->'response')` query returned 0 rows). So save's configured metadata path is very likely **empty/missing**, save finds no sections via its path, and writes nothing — while the writer DID produce 5 full sections elsewhere. This is a WIRING/PATH MISMATCH between writer output and save input, not a content problem.

### A CONTRADICTION to resolve FIRST (honest flag)
The guard's `return nil,err` yields NO page_id; save's "no sections found" early-return (lines 211-220) INCLUDES page_id. `_diag1` showed EMPTY `result_page_id` — which is what originally pointed at the guard. But the guard is now ruled out. So either (a) my mapping of which return yields empty page_id is wrong, or (b) `site_work_items.result` is populated by the handler/dispatch, not the action's return map, so empty result_page_id doesn't imply the guard at all. **Resolve this before trusting any page_id-based signature** — read how `site_work_items.result` is set for a page-build-handler run.

### DECISIVE NEXT READS (new chat, replaces the old new_text_chars step which is now DONE = 33k)
1. **What did save actually receive?** Inspect index-run collected_data for `page_content.response.sections_metadata` — present? empty? The writer child's `page_content.response` had no keys, so likely save's path is empty.
2. **Where's the bridge?** 472eed7d has a `compile_page` key — inspect it. Is `compile_page` (or an assembly step) supposed to populate `page_content.response.sections_metadata` from `section_output_*`? Did it run/produce? The mismatch is likely here: writer emits per-section, save expects compiled sections_metadata, and the bridge is missing/empty on this path.
3. **Resolve the page_id contradiction:** read how `site_work_items.result` / result_page_id is populated (action return vs handler-set) so the empty-page_id clue is interpreted correctly.
4. Also: the HTML fallback in save (`saveSectionsExtractFromHTML`) — does save even reach it? It needs `html_field` (default `assembled_page.html`) populated; if neither metadata nor assembled HTML is present on index's run, save hits "no sections found" and writes nothing. Check whether `assembled_page.html` exists in the run.

### Two BUGS already identified (regardless of which mechanism; the guard ITSELF is correct to refuse a wipe)
1. **error_step: complete_error launders a genuine save failure into `complete`.** A blocked/failed save must surface as a NON-terminal/needs-review work-item status, never `complete`. **FIX WRITTEN:** `/mnt/user-data/outputs/page_build_handler_save_failure_visible.sql` — adds a `mark_save_failed` step (fail_work_item → needs_human_review) and repoints `save_sections.error_step` to it. **PREREQUISITE before applying:** `save_sections` has error_step in TWO places (step-level + config-level), both `complete_error`; UNKNOWN which the engine reads for routing — must read the chassis engine's error_step resolution (step.ErrorStep vs config["error_step"]) and set the right one (safest: both). Snapshot/revert included.
2. **Deploy proceeds after a no-write save** → re-renders + commits stale components (the "git committed ≠ new content" trap). FIX direction (not written): gate `deploy_page` on `sections_saved > 0`.

### SEPARATE ROOT QUESTION (the content one, still open)
If the guard is correctly catching a genuinely-thin regeneration: WHY does index's `mode:recreate` + "preserve existing copy" writer produce thin content? Suspicion (TO TEST): the recreate/preserve path may not load the existing rich content_data into the writer prompt, so it regenerates sparse content. index's stored hero content_data is rich (headline/subheadline/CTAs/bg image) — if output is thin, "preserve" isn't preserving. Needs the per-section measurement above + a read of how the writer's recreate path sources existing content.

### Schema corrections (carry forward — I had these wrong)
- **There is NO `page_components.resolved_data` column.** Section data is in `page_components.content_data` (jsonb). Read CTA via `content_data->>'cta_url'`.
- `page_components` has a locking subsystem (locked_at/locked_by/lock_type[permanent|timed|review]/lock_expires_at) + `trigger_auto_lock_on_deploy` — NOT the cause here (all NULL on index) but relevant to know.
- `site_work_items`: category col is `pipeline` (not domain); error col is `error` (not error_message). Dedup is partial-unique `(site_id,item_key)` over NON-terminal statuses — terminal rows (complete/failed/...) are excluded, so a fresh item_key re-queues cleanly.

## 4. The SECOND defect (separate, lower priority): pathfinding game has no widget
`games/pathfinding/index.html` renders hero + "How A* Actually Works" prose but NO interactive game (no canvas/script). The OTHER 4 games DO have their interactive surface (user-confirmed) → isolated to pathfinding, NOT the systemic all-games recreation-loss. PLAN written: `/mnt/user-data/outputs/PLAN_pathfinding_missing_game.md` — diagnosis-first, scoped queries (pathfinding vs working sibling), candidates: clobber (pathfinding WAS a §5 rebuild target; save_page_sections delete-and-reinsert can drop a widget row not render_action-preserved), never-persisted, mis-target, the `<div>`-not-`<section>` parser miss (v2_49 §A1). CONFIRMED same problem class as the May–June "adopted interactive page no widget" work (PLAN_tool_widget_clobber.md / HANDOFF_2026-05-26_tool_routing_fix_deployed.md). Do NOT bulk-re-trigger before pinning the candidate.

## 5. Speed-up TODO (user wants soon)
Rebuild pipeline takes MANY HOURS; should be 1-2h even with slow dispatcher tick + container boot. NOT investigated. Angles: dispatcher tick interval; per-item container cold-boot vs warm/long-lived handler; sequential vs parallel claim; whether items serialise. Note index's ~610-770s per single rebuild — if per-section work is serial, that's both why index is slow AND part of the batch-hours problem (same root).

## 6. Parked / later (deliberate, gated)
- Enable `phantom_internal_links` in completeness-discovery-agent (observe-only): snapshot + jsonb_set append to its run_checks.config.checks, run once via kcat; findings land status='detected' (unclaimable, sweep stays off). Code confirmed correct, no change needed. Gate: do after index is fixed.
- Re-enable improvement-sweep — only after observe-only is clean.
- Readopt gamedesign.uk → gamesdesign.co.uk as from-scratch acceptance + content-quality baseline.

## 7. Content-quality defects observed (adopt-path, parked — next package, EXPECTED to recur on readopt)
- Title + card titles carry brand suffix "- GameDesign.uk" / "| GameDesign.uk" — and it's the OLD-domain casing (GameDesign.uk, not gamesdesign.co.uk) → brand-name source is stale/wrong-domain.
- Footer brand-tagline empty (`<p></p>`); footer contact empty (`<a href="mailto:"></a>`, empty phone) — identity.contact null. **LEAD: finetune.uk has WORKING contact details** — inspect its site_specs identity/contact + content_data shape, compare to gamesdesign's nulls.
- games hub + index meta description empty.

## 8. Deliverables produced this session (all in /mnt/user-data/outputs/)
- `page_build_handler_save_failure_visible.sql` — the visibility fix (see §3 bug 1); has an unmet prerequisite (which error_step the engine reads).
- `PLAN_pathfinding_missing_game.md` — the pathfinding plan (§4).
- `016_debugging_guide_v2_49.md` — debugging guide (from uploaded v2_48) with new entries: two rebuild routes (page-build-handler vs page-rebuild section_plan); complete-but-stale / pod-died (now split into concurrent-deploy vs reproducible-same-page sub-cases); "git committed ≠ new content"; save_page_sections regression-guard → complete_error masking. **NOTE: parts of the v2_49 index narrative are now SUPERSEDED by this handoff** (the duration/lease sub-case was wrong for index; the confirmed picture is the regression-guard/laundering above). The new chat should reconcile v2_49 with §3 here once new_text_chars confirms the mechanism — do NOT treat v2_49's index entry as final.
- `running_notes_17_internal_linking_phantom_fixes.md` — full chronological reasoning trail (every theory raised/falsified, with the confounds). The authoritative detailed record.
- (earlier, still valid) `RUNBOOK_linking_phantom_fixes.md`, the various FOCUS_/PLAN_ docs, `check_linking_sql_applied.sql`.

---

## What to put in the NEW CHAT's project files (user asked)
Upload these so the new chat has ground truth without re-deriving:
1. **This handoff** (`HANDOFF_2026-06-15_index_stale_rebuild.md`) — the entry point.
2. **`save_page_sections_action.go`** (deployed) — the regression guard is the crux; the new chat must read it directly.
3. **The `page-build-handler` agent_definitions row** — `SELECT * FROM agent_definitions WHERE type='page-build-handler';` (its `default_config.workflow` has the save_sections step + complete_error; needed for the error_step prerequisite).
4. **`page-content-writer` agent_definitions row** — the writer whose recreate/preserve path may under-produce (the content root question); the new chat will need its workflow to see how recreate sources existing content.
5. **The chassis workflow-engine code that resolves `error_step`** — whatever file handles a step error and reads error_step (step.ErrorStep vs config["error_step"]). This is the unmet prerequisite for the visibility fix; without it the new chat is blocked on the same question.
6. **`running_notes_17_...md`** and **`016_debugging_guide_v2_49.md`** — the reasoning trail + the guide to update.
7. **`page_build_handler_save_failure_visible.sql`** — the pending fix.
8. Optionally: `save_page_sections_patch.go` (the prior findPreservedComponentIDs clobber fix, referenced by the pathfinding plan) and `PLAN_pathfinding_missing_game.md` if continuing that thread too.

Also useful to state up front in the new chat: site_id e33263f4-...; the writer child orch id 472eed7d-... (for the new_text_chars measurement); and that `_diag1` (item_key linking_rebuild_index_diag1) is the reproduction — its components are still 06-06, so it can be re-measured anytime.

## FIRST ACTIONS in the new chat (ordered) — UPDATED 2026-06-15 (guard ruled out)
The new_text_chars measurement is DONE: writer produced 33,335 chars (section_output_0..4) >> 5760 threshold → **regression guard FALSIFIED**. The writer is fine; defect is in how save receives sections. So:
1. **Inspect what save received on index's run:** is `page_content.response.sections_metadata` present/empty in collected_data? (Writer child's `page_content.response` had no keys → likely empty.) And is `assembled_page.html` (save's HTML-fallback field) present? If both empty → save hits "no sections found" / writes nothing despite the writer's 5 sections.
2. **Find the missing bridge:** inspect 472eed7d's `compile_page` key — is compile_page meant to turn `section_output_*` into the `sections_metadata` (or assembled HTML) that save reads? Did it run and produce? The wiring mismatch (writer emits per-section; save reads compiled metadata) is the leading defect; the bridge is the suspect.
3. **Resolve the page_id contradiction:** empty `result_page_id` originally implied the guard's `return nil,err`, but the guard is ruled out. Read how `site_work_items.result` is populated (action return vs handler/dispatch) so the signature is interpreted correctly.
4. Read the engine's error_step resolution → unblock `page_build_handler_save_failure_visible.sql` → apply (snapshot first). This visibility fix is STILL valuable regardless of the root cause (a zero-section save should surface, not complete silently).
5. Once the save-input mismatch is pinned: fix the bridge (make the writer's per-section output reach save as sections_metadata, or point save at where the writer actually puts it), gate deploy on sections_saved>0, re-run index, confirm components flip to current date + committed HTML has real hub CTAs + no phantoms. Then the §6 whole-site audit closes the linking batch.

## (historical) original first-actions before guard was ruled out
1. ~~Measure new_text_chars on 472eed7d vs 5760~~ DONE = 33335, guard falsified.
2. Confirm the save_sections metadata input (now step 1 above — promoted).
3. Engine error_step resolution → visibility fix (now step 4).
4. Root fix per measurement (now step 5).
5. Re-run index + §6 audit.

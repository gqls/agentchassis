# Register — work-item-integrity

7 concepts, consolidated from 9 raw extractions across units U05, U09.

### WII-001 — Silent-completion failure family ("work reports success but doesn't happen")
- **status:** partial
- **status-evidence:** "modes 1-3 resolved; one residual gap" (running_notes_15(12) Part 11); "2026-06-09 update: modes 1-3 are now resolved in current code; one residual gap remains (`complete_work_item` clobber)" — independently confirmed by a second unit the same week. Fix B (complete_error laundering) is "deferred, low urgency given monitor=0."
- **what:** The unifying defect class across two separate investigations: work items reach `complete` while the artifact was never produced. Members catalogued: the result-stub drop; `complete_error` being a SUCCESS-labelled `complete_workflow` for sectionless/skip paths; error_step laundering a genuine save failure into complete; the old reaper auto-complete on lost handler responses; claim-timeout marked complete instead of reset; `complete_work_item` clobbering deliberate flags (needs_human_review); deploy re-committing stale components ("git committed ≠ new content"). Doctrine that emerged: trust rendered HTML / DB state over work-item status; a blocked/failed step must surface as a non-terminal status, never `complete`. The gamesdesign homepage (deployed+stamped in DB, no file in repo) and a guide page stuck in the same state were direct consequences.
- **sources:** running_notes_15(12).md#part-10-12; HANDOFF_2026-06-09(2).md#RESOLVED; FOCUS_page_build_handler_silent_completion(1).md; CATALOGUE(9)#A4
- **relations:** evidence-gated claimed-item-timeout reaper (WII-002); complete_work_item flag-preservation guard (WII-003); complete_error family (page-build-pipeline register, PBP-016, PBP-020); silent completion pathology (work-dispatch register, WDS-004)
- **verify-later:** FOCUS_page_build_handler_silent_completion.md; positive-evidence monitor query results; load_work_item_actions.go CompleteWorkItemAction guard

### WII-002 — Evidence-gated claimed-item-timeout reaper (positive-evidence completion + stale-claim reset)
- **status:** deployed
- **status-evidence:** "Option A APPLIED + verified live (enabled=t, interval 120s, new pre_query 14:09)"; re-confirmed a second time: "the v2 migration is in place — claimed-item-timeout shows enabled=t, interval_seconds=120, and the new 'provably done on the specific targeted artifact' pre_query" (2026-06-04); "Mode 1… RESOLVED… Mode 3… RESOLVED" (2026-06-09).
- **what:** The `claimed-item-timeout` scheduled task's SQL pre_query auto-completes a stuck claim ONLY with positive artifact evidence specific to the item type — `page_components` updated after claim for needs_content_page (a v2 migration decoupled it from the untrustworthy `build_status='deployed'` flag), `deployed_at` for page_rerender, head-slot update for needs_design — else resets: attempt_count+1, back to triaged (or failed at max_attempts, threshold ~40 min). Replaced the loose "any page updated since claim" auto-complete that had falsely completed a homepage build. The reset CTE IS the claim-watchdog a separate design doc proposed building — explicitly cancelled as duplication once this was found ("REVISED — DO NOT BUILD THIS. The reset already exists.").
- **sources:** running_notes_14(26).md#part-11-12; running_notes_15(12).md#part-11; FOCUS_dispatch_throughput_and_claim_watchdog(3).md#decision; FOCUS_page_build_handler_silent_completion(1).md#update
- **relations:** silent-completion failure family (WII-001); positive-evidence deploy guard (WII-007); dispatch chain + NOT EXISTS blocker (work-dispatch register, WDS-002)
- **verify-later:** scheduled_tasks claimed-item-timeout pre_query; migration_claimed_item_timeout_evidence_v2.sql; the 40-min threshold tuning note (~25 min floor above the 1200s call_handler)

### WII-003 — complete_work_item flag-preservation guard (Fix A)
- **status:** deployed
- **status-evidence:** "Fix A marked applied" in the FOCUS doc; listed in the deploy batch ("Fix A must ship with S2").
- **what:** `CompleteWorkItemAction` did an unconditional UPDATE to status='complete', so the dispatch loop's mark_complete clobbered deliberate handler-set flags (needs_human_review from mark_needs_review / mark_no_sections). The guard adds `AND status NOT IN (<flagged/terminal set>)` and returns completed=rows>0. Confirmed necessary by inference: a sectionless-page retry proved `complete_error → dispatch mark_complete` fires. Prerequisite for the sectionless-page durability stack's S2 flag (page-build-pipeline register, PBP-017) and for needs_human_review to be effective at all.
- **sources:** running_notes_15(12).md#part-11-12; HANDOFF_2026-06-09(2).md
- **relations:** silent-completion failure family (WII-001); sectionless-page durability stack (page-build-pipeline register, PBP-017); workItemTerminalStatuses (needs_human_review deliberately NON-terminal)
- **verify-later:** load_work_item_actions.go CompleteWorkItemAction WHERE clause

### WII-004 — item_key canonicalization + dedup namespace decisions
- **status:** partial
- **status-evidence:** "Part 3 — CODE PREPARED, not applied — apply after Part 2 verifies."
- **what:** item_key prefixes drifted from item_type across creators, causing two confirmed bugs: (1) content rebuilds not co-deduping (`needs_page:<name>` vs `page_rerender:<name>` for the same work → double builds); (2) adoption keying BOTH `needs_tool_recreation` and `needs_content_page` as `needs_page:<name>` → unique-index collision silently drops one (observed live: a tool-recreation item was mis-keyed). Fix: a plain `workItemKey(itemType, target)` builder in work_items_common.go, with the tool branch getting its own namespace; the content branch was DECIDED (Option B) to stay in the needs_page namespace, preserving a deliberate planner co-dedup design — the prefix==item_type invariant carries one documented exception. Doctrine until shipped: route/diagnose by item_type → handler_agent, never by item_key.
- **sources:** NOTES(44)#item_key-contract; RUNBOOK_gamesdesign_index_rebuild(29).md#part-3; HANDOFF_page_pipeline(11).md#6
- **relations:** work-item dedup index (WII-005); interactive-page clobber (page-build-pipeline register, PBP-012); adoption apply_adoption_plan
- **verify-later:** work_items_common.go workJobKey; apply_adoption_plan_action.go lines ~627-655

### WII-005 — Work-item dedup index + two-strike anti-churn rule
- **status:** deployed
- **status-evidence:** running_notes_15(12) Part 8: "Key enabling discovery (insertWorkItem …): a built-in two-strike rule."
- **what:** `idx_swi_dedup` is a partial UNIQUE(site_id,item_key) over non-terminal statuses only — terminal rows are excluded so a completed key can be requeued cleanly; ON CONFLICT DO NOTHING is the safe insert idiom. `insertWorkItem` adds a two-strike rule (an item_key with ≥2 terminal attempts in 7 days inserts as `unresolved`; <3h after a terminal item is suppressed), so discovery checks need no anti-churn logic of their own. A non-terminal flag (needs_human_review) deliberately holds the dedup slot, preventing re-trigger loops. Same rule as the work-dispatch register's WDS-007, described here from the dedup-index angle.
- **sources:** running_notes_15(12).md#part-8; HANDOFF_page_pipeline(11).md#schema-gotchas; RUNBOOK_linking_phantom_fixes(7).md#5
- **relations:** item_key canonicalization (WII-004); two-strike rule (work-dispatch register, WDS-007); sectionless durability stack (page-build-pipeline register, PBP-017)
- **verify-later:** insertWorkItem two-strike logic; idx_swi_dedup definition

### WII-006 — Work-item routing map (item_type → handler agent)
- **status:** deployed
- **status-evidence:** NOTES(44) 2026-06-19 "Corrected trigger (confirmed from live config, not inferred)."
- **what:** `needs_page`/`needs_content_page`/`link_resolution_rebuild` → page-build-handler (full content path through the writer); `page_rerender` → page-rerender (assemble-from-DB + deploy, plus the no-LLM re-render pre-pass); `needs_rerender` → rerender-pages (site loop that mints per-page page_rerender items via create_rerender_items); `needs_tool_recreation` → tool-recreation-handler. build-dispatch-loop claims status triaged/approved only; discovery findings land 'detected' (unclaimable). page-build-handler does NOT branch on item_type — dispatch metadata only; it reads spec.page_id/page_name/mode/suggestion.
- **sources:** NOTES(44) 2026-06-19; RUNBOOK_linking_phantom_fixes(7).md#5; running_notes_17(21).md#page-build-handler-contract
- **relations:** work-item routing (work-dispatch register, WDS-011); re-render vs rebuild distinction (page-build-pipeline register, PBP-010); dispatch throughput (work-dispatch register, WDS-002)
- **verify-later:** build-dispatch-loop claim SQL; agent workflow defs for the four handlers

### WII-007 — Positive-evidence deploy guard (0-component page never marked deployed)
- **status:** deployed
- **status-evidence:** "(B) DONE — v3_site_actions.go patch delivered… UpdatePageStatusAction now calls pageHasComponents(pageID) before marking deployed; if 0 rendered components it refuses and flips to needs_rebuild (clearing the stamp)" (applied 2026-06-04).
- **what:** `pageHasComponents` (EXISTS on page_components with non-null component_id + non-empty rendered_html) gates the `deployed` status write; a 0-component page flips to `needs_rebuild` with its plan stamp cleared so the reconciler rebuilds instead of skipping. Fail-open on check error so a transient failure can't halt legitimate deploys. Makes `build_status='deployed'` trustworthy for downstream evidence checks such as the claimed-item-timeout reaper (WII-002). Same underlying code documented from the page-build angle in the page-build-pipeline register (PBP-023).
- **sources:** CATALOGUE(9)#A4; running_notes_14(25)#part-12-addendum-2
- **relations:** UpdatePageStatusAction zero-component deploy guard (page-build-pipeline register, PBP-023, same mechanism); evidence-gated claimed-item-timeout reaper (WII-002); silent-completion failure family (WII-001)
- **verify-later:** v3_site_actions.go pageHasComponents + UpdatePageStatusAction deployed branch

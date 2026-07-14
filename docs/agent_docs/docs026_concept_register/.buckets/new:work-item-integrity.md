
<!-- SOURCE: U05_content_quality_linking.md -->
### Silent-completion failure family ("work reports success but doesn't happen")
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** running_notes_15(12) Part 11 "modes 1–3 resolved; one residual gap"; Fix B "deferred, low urgency (monitor=0)"; NOTES(44) documents further members through June.
- **what:** The unit's unifying defect class: work items reach `complete` while the artifact was never produced. Members catalogued: the result-stub drop (Part 1); complete_error being a SUCCESS-labelled complete_workflow for sectionless/skip paths; error_step laundering a genuine save failure into complete; the old reaper auto-complete on lost response; claim-timeout marked complete; complete_work_item clobbering deliberate flags; deploy re-committing stale components ("git committed ≠ new content"). Doctrine that emerged: trust rendered HTML / DB state over work-item status; a blocked/failed step must surface as a non-terminal status, never `complete`.
- **sources:** running_notes_15(12).md#part-10-12; HANDOFF_2026-06-09(2).md#RESOLVED; NOTES(44) passim; page_build_handler_save_failure_visible.sql (header)
- **relations:** every fix below: evidence-gated reaper, Fix A, mark_no_sections, mark_save_failed, positive-evidence monitor. FOCUS_page_build_handler_silent_completion.md is the home doc.
- **verify-later:** FOCUS_page_build_handler_silent_completion.md; positive-evidence monitor query results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Evidence-gated claimed-item-timeout reaper (positive-evidence completion + reset)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2: "Option A APPLIED + verified live (enabled=t, interval 120s, new pre_query 14:09)"; running_notes_15(12) Part 11 re-confirms.
- **what:** The claimed-item-timeout scheduled task's SQL pre_query auto-completes a stuck claim ONLY with positive artifact evidence specific to the item type (page_components updated after claim for needs_content_page — the v2 migration decoupled it from the untrustworthy build_status='deployed' flag; deployed_at for page_rerender), else resets: attempt_count+1, back to triaged (or failed at max). Replaced the loose "any page updated since claim" auto-complete that falsely completed the gamesdesign homepage build. The reset branch made a separately-planned stale-claim watchdog redundant (reuse-not-build).
- **sources:** running_notes_14(26).md#part-11-12; running_notes_15(12).md#part-11; HANDOFF_2026-06-09(2).md#FOCUS-modes
- **relations:** silent-completion family; UpdatePageStatusAction 0-component guard (keeps the evidence honest).
- **verify-later:** scheduled_tasks claimed-item-timeout pre_query; migration_claimed_item_timeout_evidence_v2.sql.

<!-- SOURCE: U05_content_quality_linking.md -->
### complete_work_item flag-preservation guard (Fix A)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-09(2): "Fix A marked applied" in the FOCUS doc; listed in the deploy batch ("Fix A must ship with S2").
- **what:** CompleteWorkItemAction did an unconditional UPDATE to status='complete', so the dispatch loop's mark_complete clobbered deliberate handler-set flags (needs_human_review from mark_needs_review / mark_no_sections). The guard adds `AND status NOT IN (<flagged/terminal set>)` and returns completed=rows>0. Confirmed necessary by inference: the skinner-box sectionless retry proved complete_error → dispatch mark_complete fires. Prerequisite for S2 and for the existing HITL flag to be effective.
- **sources:** running_notes_15(12).md#part-11-12; HANDOFF_2026-06-09(2).md
- **relations:** silent-completion family; sectionless durability stack; workItemTerminalStatuses (needs_human_review deliberately NON-terminal).
- **verify-later:** load_work_item_actions.go CompleteWorkItemAction WHERE clause.

<!-- SOURCE: U05_content_quality_linking.md -->
### item_key canonicalization (Part 3/B) + dedup namespace decisions
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 3 — CODE PREPARED, not applied — apply after Part 2 verifies".
- **what:** item_key prefixes drifted from item_type across creators, causing two confirmed bugs: (1) content rebuilds not co-deduping (needs_page:<name> vs page_rerender:<name> for the same work → double builds); (2) adoption keying BOTH needs_tool_recreation and needs_content_page as needs_page:<name> → unique-index collision silently drops one (observed live: the pathfinding tool-recreation item mis-keyed). Fix: a plain `workItemKey(itemType, target)` builder in work_items_common.go, tool branch → its own namespace; content branch DECIDED (Option B) to stay in the needs_page namespace, preserving the deliberate doc-029 planner co-dedup — the prefix==item_type invariant carries one documented exception. Doctrine until shipped: route/diagnose by item_type → handler_agent, never by item_key.
- **sources:** NOTES(44)#item_key-contract + Part B sections; RUNBOOK_gamesdesign_index_rebuild(29).md#part-3; HANDOFF_page_pipeline(11).md#6
- **relations:** dedup index; work-item routing; adoption apply_adoption_plan.
- **verify-later:** work_items_common.go workItemKey; apply_adoption_plan_action.go lines ~627–655; P3.2 survey results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item dedup index + two-strike anti-churn rule
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 8 "Key enabling discovery (insertWorkItem …): a built-in two-strike rule".
- **what:** idx_swi_dedup is a partial UNIQUE(site_id,item_key) over non-terminal statuses only — terminal rows are excluded so a completed key can be requeued cleanly; ON CONFLICT DO NOTHING is the safe insert idiom. insertWorkItem adds a two-strike rule (an item_key with ≥2 terminal attempts in 7 days inserts as `unresolved`; <3h after a terminal item is suppressed), so discovery checks need no anti-churn logic of their own. A non-terminal flag (needs_human_review) deliberately holds the dedup slot, preventing re-trigger loops.
- **sources:** running_notes_15(12).md#part-8; HANDOFF_page_pipeline(11).md#schema-gotchas; RUNBOOK_linking_phantom_fixes(7).md#5
- **relations:** item_key canonicalization; sectionless durability stack.
- **verify-later:** insertWorkItem two-strike logic; idx_swi_dedup definition.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item routing map (item_type → handler agent)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19 "Corrected trigger (confirmed from live config, not inferred)".
- **what:** needs_page / needs_content_page / link_resolution_rebuild → page-build-handler (full content path through the writer); page_rerender → page-rerender (assemble-from-DB + deploy; after Part 2 also the no-LLM re-render pre-pass); needs_rerender → rerender-pages (site loop that mints per-page page_rerender items via create_rerender_items); needs_tool_recreation → tool-recreation-handler. build-dispatch-loop claims status triaged/approved only; discovery findings land 'detected' (unclaimable). page-build-handler does NOT branch on item_type — dispatch metadata only; it reads spec.page_id/page_name/mode/suggestion.
- **sources:** NOTES(44) 2026-06-19; RUNBOOK_linking_phantom_fixes(7).md#5 handler facts; running_notes_17(21).md#page-build-handler-contract
- **relations:** re-render vs rebuild; interactive clobber (link_resolution_rebuild routed to the full builder by design); dispatch throughput.
- **verify-later:** build-dispatch-loop claim SQL; agent workflow defs for the four handlers.

<!-- SOURCE: U09_adoption.md -->
### Silent-completion family: "complete" means "we stopped", not "the work succeeded"
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** "2026-06-09 update: modes 1–3 are now resolved in current code; one residual gap remains (`complete_work_item` clobber)… Fix A (applied 2026-06-09)… Fix B (deferred, low urgency given monitor=0)."
- **what:** The architectural flaw that a work item reaches `status='complete'` without the work succeeding, in several modes: (1) reaper auto-complete on lost handler responses ("Auto-completed: work verified done despite lost response"), (2) validate_content failures routed to complete, (3) claim-timeout marked complete instead of reset, plus the dispatch-level variants — the unguarded `CompleteWorkItemAction` clobbering handler-set `needs_human_review` flags (Fix A: status guard applied) and `complete_error` being a SUCCESS-labelled `complete_workflow` on genuine-failure paths (Fix B, deferred). Modes 1–3 resolved via the evidence-gated reaper; the rule is: complete only on explicit handler success OR positive DB evidence. The gamesdesign homepage (deployed+stamped in DB, no file in repo) and guide-skinner-box were direct consequences.
- **sources:** FOCUS_page_build_handler_silent_completion(1).md, HANDOFF_2026-06-09, running_notes_15(10)#part-10–12, CATALOGUE(9)#A4
- **relations:** claimed-item-timeout reaper; positive-evidence completion; sectionless-page durability (S2 depends on Fix A); work-item lifecycle (001)
- **verify-later:** `load_work_item_actions.go` CompleteWorkItemAction guard; page-build-handler `complete_error` semantics; monitor query results

<!-- SOURCE: U09_adoption.md -->
### claimed-item-timeout scheduled task: evidence-gated completion + stale-claim reset
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "Option A APPLIED + verified live: the v2 migration is in place — claimed-item-timeout shows enabled=t, interval_seconds=120, and the new 'provably done on the specific targeted artifact' pre_query" (2026-06-04); "Mode 1… RESOLVED… Mode 3… RESOLVED" (2026-06-09 re-verification).
- **what:** A scheduled task whose SQL pre_query both (a) auto-completes a stuck claimed item only with positive artifact evidence — `page_components` with `component_id` + non-empty `rendered_html` + `updated_at > claimed_at` for needs_content_page, `deployed_at > claimed_at` for page_rerender, head-slot update for needs_design — and (b) resets stale claims (>40 min, no evidence) to `triaged` (or `failed` at max_attempts) with attempt_count+1. The reset CTE IS the Lever-C claim watchdog the dispatch doc designed — building a separate watchdog was explicitly cancelled as duplication ("REVISED 2026-06-04 — DO NOT BUILD THIS. The reset already exists."). Evidence deliberately prefers ground-truth artifacts over the untrustworthy `build_status='deployed'` flag.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#decision, running_notes_14(25)#part-12, FOCUS_page_build_handler_silent_completion(1).md#update
- **relations:** silent-completion family; Option B deployed-guard keeps the flag honest; dispatch NOT-EXISTS deadlock (the reset unfreezes the site)
- **verify-later:** scheduled_tasks row `claimed-item-timeout` pre_query (v2: page_components evidence for needs_content_page); 40-min threshold tuning note (~25 min floor above the 1200s call_handler)

<!-- SOURCE: U09_adoption.md -->
### Positive-evidence deploy guard (0-component page never marked deployed)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "(B) DONE — v3_site_actions.go patch delivered… UpdatePageStatusAction now calls pageHasComponents(pageID) before marking deployed; if 0 rendered components it refuses and flips to needs_rebuild (clearing the stamp)" (CATALOGUE A4, applied 2026-06-04).
- **what:** `pageHasComponents` (EXISTS on page_components with non-null component_id + non-empty rendered_html) gates the `deployed` status write; a 0-component page flips to `needs_rebuild` with stamp cleared so the reconciler rebuilds instead of skipping. Fail-open on check error so a transient failure can't halt legitimate deploys. Makes `build_status='deployed'` trustworthy for downstream evidence checks.
- **sources:** CATALOGUE(9)#A4, running_notes_14(25)#part-12-addendum-2
- **relations:** silent-completion family; claimed-item-timeout evidence
- **verify-later:** v3_site_actions.go pageHasComponents + UpdatePageStatusAction deployed branch

<!-- SOURCE: U05_content_quality_linking.md -->
### Silent-completion failure family ("work reports success but doesn't happen")
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** running_notes_15(12) Part 11 "modes 1–3 resolved; one residual gap"; Fix B "deferred, low urgency (monitor=0)"; NOTES(44) documents further members through June.
- **what:** The unit's unifying defect class: work items reach `complete` while the artifact was never produced. Members catalogued: the result-stub drop (Part 1); complete_error being a SUCCESS-labelled complete_workflow for sectionless/skip paths; error_step laundering a genuine save failure into complete; the old reaper auto-complete on lost response; claim-timeout marked complete; complete_work_item clobbering deliberate flags; deploy re-committing stale components ("git committed ≠ new content"). Doctrine that emerged: trust rendered HTML / DB state over work-item status; a blocked/failed step must surface as a non-terminal status, never `complete`.
- **sources:** running_notes_15(12).md#part-10-12; HANDOFF_2026-06-09(2).md#RESOLVED; NOTES(44) passim; page_build_handler_save_failure_visible.sql (header)
- **relations:** every fix below: evidence-gated reaper, Fix A, mark_no_sections, mark_save_failed, positive-evidence monitor. FOCUS_page_build_handler_silent_completion.md is the home doc.
- **verify-later:** FOCUS_page_build_handler_silent_completion.md; positive-evidence monitor query results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Evidence-gated claimed-item-timeout reaper (positive-evidence completion + reset)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2: "Option A APPLIED + verified live (enabled=t, interval 120s, new pre_query 14:09)"; running_notes_15(12) Part 11 re-confirms.
- **what:** The claimed-item-timeout scheduled task's SQL pre_query auto-completes a stuck claim ONLY with positive artifact evidence specific to the item type (page_components updated after claim for needs_content_page — the v2 migration decoupled it from the untrustworthy build_status='deployed' flag; deployed_at for page_rerender), else resets: attempt_count+1, back to triaged (or failed at max). Replaced the loose "any page updated since claim" auto-complete that falsely completed the gamesdesign homepage build. The reset branch made a separately-planned stale-claim watchdog redundant (reuse-not-build).
- **sources:** running_notes_14(26).md#part-11-12; running_notes_15(12).md#part-11; HANDOFF_2026-06-09(2).md#FOCUS-modes
- **relations:** silent-completion family; UpdatePageStatusAction 0-component guard (keeps the evidence honest).
- **verify-later:** scheduled_tasks claimed-item-timeout pre_query; migration_claimed_item_timeout_evidence_v2.sql.

<!-- SOURCE: U05_content_quality_linking.md -->
### complete_work_item flag-preservation guard (Fix A)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-09(2): "Fix A marked applied" in the FOCUS doc; listed in the deploy batch ("Fix A must ship with S2").
- **what:** CompleteWorkItemAction did an unconditional UPDATE to status='complete', so the dispatch loop's mark_complete clobbered deliberate handler-set flags (needs_human_review from mark_needs_review / mark_no_sections). The guard adds `AND status NOT IN (<flagged/terminal set>)` and returns completed=rows>0. Confirmed necessary by inference: the skinner-box sectionless retry proved complete_error → dispatch mark_complete fires. Prerequisite for S2 and for the existing HITL flag to be effective.
- **sources:** running_notes_15(12).md#part-11-12; HANDOFF_2026-06-09(2).md
- **relations:** silent-completion family; sectionless durability stack; workItemTerminalStatuses (needs_human_review deliberately NON-terminal).
- **verify-later:** load_work_item_actions.go CompleteWorkItemAction WHERE clause.

<!-- SOURCE: U05_content_quality_linking.md -->
### item_key canonicalization (Part 3/B) + dedup namespace decisions
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 3 — CODE PREPARED, not applied — apply after Part 2 verifies".
- **what:** item_key prefixes drifted from item_type across creators, causing two confirmed bugs: (1) content rebuilds not co-deduping (needs_page:<name> vs page_rerender:<name> for the same work → double builds); (2) adoption keying BOTH needs_tool_recreation and needs_content_page as needs_page:<name> → unique-index collision silently drops one (observed live: the pathfinding tool-recreation item mis-keyed). Fix: a plain `workItemKey(itemType, target)` builder in work_items_common.go, tool branch → its own namespace; content branch DECIDED (Option B) to stay in the needs_page namespace, preserving the deliberate doc-029 planner co-dedup — the prefix==item_type invariant carries one documented exception. Doctrine until shipped: route/diagnose by item_type → handler_agent, never by item_key.
- **sources:** NOTES(44)#item_key-contract + Part B sections; RUNBOOK_gamesdesign_index_rebuild(29).md#part-3; HANDOFF_page_pipeline(11).md#6
- **relations:** dedup index; work-item routing; adoption apply_adoption_plan.
- **verify-later:** work_items_common.go workItemKey; apply_adoption_plan_action.go lines ~627–655; P3.2 survey results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item dedup index + two-strike anti-churn rule
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 8 "Key enabling discovery (insertWorkItem …): a built-in two-strike rule".
- **what:** idx_swi_dedup is a partial UNIQUE(site_id,item_key) over non-terminal statuses only — terminal rows are excluded so a completed key can be requeued cleanly; ON CONFLICT DO NOTHING is the safe insert idiom. insertWorkItem adds a two-strike rule (an item_key with ≥2 terminal attempts in 7 days inserts as `unresolved`; <3h after a terminal item is suppressed), so discovery checks need no anti-churn logic of their own. A non-terminal flag (needs_human_review) deliberately holds the dedup slot, preventing re-trigger loops.
- **sources:** running_notes_15(12).md#part-8; HANDOFF_page_pipeline(11).md#schema-gotchas; RUNBOOK_linking_phantom_fixes(7).md#5
- **relations:** item_key canonicalization; sectionless durability stack.
- **verify-later:** insertWorkItem two-strike logic; idx_swi_dedup definition.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item routing map (item_type → handler agent)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19 "Corrected trigger (confirmed from live config, not inferred)".
- **what:** needs_page / needs_content_page / link_resolution_rebuild → page-build-handler (full content path through the writer); page_rerender → page-rerender (assemble-from-DB + deploy; after Part 2 also the no-LLM re-render pre-pass); needs_rerender → rerender-pages (site loop that mints per-page page_rerender items via create_rerender_items); needs_tool_recreation → tool-recreation-handler. build-dispatch-loop claims status triaged/approved only; discovery findings land 'detected' (unclaimable). page-build-handler does NOT branch on item_type — dispatch metadata only; it reads spec.page_id/page_name/mode/suggestion.
- **sources:** NOTES(44) 2026-06-19; RUNBOOK_linking_phantom_fixes(7).md#5 handler facts; running_notes_17(21).md#page-build-handler-contract
- **relations:** re-render vs rebuild; interactive clobber (link_resolution_rebuild routed to the full builder by design); dispatch throughput.
- **verify-later:** build-dispatch-loop claim SQL; agent workflow defs for the four handlers.

<!-- SOURCE: U09_adoption.md -->
### Silent-completion family: "complete" means "we stopped", not "the work succeeded"
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** "2026-06-09 update: modes 1–3 are now resolved in current code; one residual gap remains (`complete_work_item` clobber)… Fix A (applied 2026-06-09)… Fix B (deferred, low urgency given monitor=0)."
- **what:** The architectural flaw that a work item reaches `status='complete'` without the work succeeding, in several modes: (1) reaper auto-complete on lost handler responses ("Auto-completed: work verified done despite lost response"), (2) validate_content failures routed to complete, (3) claim-timeout marked complete instead of reset, plus the dispatch-level variants — the unguarded `CompleteWorkItemAction` clobbering handler-set `needs_human_review` flags (Fix A: status guard applied) and `complete_error` being a SUCCESS-labelled `complete_workflow` on genuine-failure paths (Fix B, deferred). Modes 1–3 resolved via the evidence-gated reaper; the rule is: complete only on explicit handler success OR positive DB evidence. The gamesdesign homepage (deployed+stamped in DB, no file in repo) and guide-skinner-box were direct consequences.
- **sources:** FOCUS_page_build_handler_silent_completion(1).md, HANDOFF_2026-06-09, running_notes_15(10)#part-10–12, CATALOGUE(9)#A4
- **relations:** claimed-item-timeout reaper; positive-evidence completion; sectionless-page durability (S2 depends on Fix A); work-item lifecycle (001)
- **verify-later:** `load_work_item_actions.go` CompleteWorkItemAction guard; page-build-handler `complete_error` semantics; monitor query results

<!-- SOURCE: U09_adoption.md -->
### claimed-item-timeout scheduled task: evidence-gated completion + stale-claim reset
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "Option A APPLIED + verified live: the v2 migration is in place — claimed-item-timeout shows enabled=t, interval_seconds=120, and the new 'provably done on the specific targeted artifact' pre_query" (2026-06-04); "Mode 1… RESOLVED… Mode 3… RESOLVED" (2026-06-09 re-verification).
- **what:** A scheduled task whose SQL pre_query both (a) auto-completes a stuck claimed item only with positive artifact evidence — `page_components` with `component_id` + non-empty `rendered_html` + `updated_at > claimed_at` for needs_content_page, `deployed_at > claimed_at` for page_rerender, head-slot update for needs_design — and (b) resets stale claims (>40 min, no evidence) to `triaged` (or `failed` at max_attempts) with attempt_count+1. The reset CTE IS the Lever-C claim watchdog the dispatch doc designed — building a separate watchdog was explicitly cancelled as duplication ("REVISED 2026-06-04 — DO NOT BUILD THIS. The reset already exists."). Evidence deliberately prefers ground-truth artifacts over the untrustworthy `build_status='deployed'` flag.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#decision, running_notes_14(25)#part-12, FOCUS_page_build_handler_silent_completion(1).md#update
- **relations:** silent-completion family; Option B deployed-guard keeps the flag honest; dispatch NOT-EXISTS deadlock (the reset unfreezes the site)
- **verify-later:** scheduled_tasks row `claimed-item-timeout` pre_query (v2: page_components evidence for needs_content_page); 40-min threshold tuning note (~25 min floor above the 1200s call_handler)

<!-- SOURCE: U09_adoption.md -->
### Positive-evidence deploy guard (0-component page never marked deployed)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "(B) DONE — v3_site_actions.go patch delivered… UpdatePageStatusAction now calls pageHasComponents(pageID) before marking deployed; if 0 rendered components it refuses and flips to needs_rebuild (clearing the stamp)" (CATALOGUE A4, applied 2026-06-04).
- **what:** `pageHasComponents` (EXISTS on page_components with non-null component_id + non-empty rendered_html) gates the `deployed` status write; a 0-component page flips to `needs_rebuild` with stamp cleared so the reconciler rebuilds instead of skipping. Fail-open on check error so a transient failure can't halt legitimate deploys. Makes `build_status='deployed'` trustworthy for downstream evidence checks.
- **sources:** CATALOGUE(9)#A4, running_notes_14(25)#part-12-addendum-2
- **relations:** silent-completion family; claimed-item-timeout evidence
- **verify-later:** v3_site_actions.go pageHasComponents + UpdatePageStatusAction deployed branch

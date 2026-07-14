
<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Work-item state machine (detected → triaged → claimed → complete/failed)
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Phase 2G + 2H verified end-to-end at scale. Seven hero items … processed through the full chain … without manual intervention" (2026-05-15)
- **what:** `detected` is a valid intermediate state: discovery emits at detected; design-audit-agent's `triage_detected_items` step (registry.go:722) promotes to triaged; dispatch claims only triaged/approved (partial indexes idx_swi_handler / idx_swi_site_pending); handlers mark complete/failed (mark_work_item_complete / mark_work_item_failed steps). There is NO automated coupling between discovery and audit — items sit in detected until an audit runs. Admin-created items insert directly at triaged.
- **sources:** FOCUS_dispatch_diagnostic(4).md#TL;DR, #Evidence-trail; HANDOFF-pipeline-triage-april-2026.md
- **relations:** dispatch chain; auto-triage open question; two-strike rule; silent completion
- **verify-later:** registry.go triage_detected_items; site_work_items partial indexes; design-audit-agent workflow

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatch chain: build-pipeline-trigger → find_dispatchable_site → build-dispatch-loop
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "researched in depth this session" (2026-05-15) with the actual selection SQL quoted; scheduled_tasks row build-pipeline-trigger every 30s
- **what:** The scheduler fires build-pipeline-trigger, whose find_dispatchable_site step picks ONE site per tick (DISTINCT ON with no outer ORDER BY — effectively arbitrary among eligible sites) and spawns build-dispatch-loop scoped to it, which loads up to 5 items (pipeline='build') and claims/spawns handlers. Throughput cap ~5 items per site per 30s, one site at a time. build-pipeline-trigger doesn't write orchestration_states, making its decisions untraceable.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** NOT EXISTS blocker; Bug 3 site-targeting; fairness ORDER BY improvement
- **verify-later:** scheduled_tasks 'build-pipeline-trigger' row; build-pipeline-trigger / build-dispatch-loop agent definitions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### NOT EXISTS whole-site claim blocker
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "A single stuck claimed item on a site excludes the entire site from dispatch consideration until it clears … by design … but it makes stuck claims a system-stopping condition" (2026-05-15)
- **what:** find_dispatchable_site's NOT EXISTS clause excludes any site with ANY item in status='claimed' — an absolute blocker, not a deprioritiser. Prevents racing claims mid-execution but converts one dead handler into a site-wide stall. Proposed (cheap, high-leverage, not built): watchdog that resets claims older than ~15 min.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3
- **relations:** claim-timeout sweeper absence (Bug 2); dispatcher stall (Bug 1)
- **verify-later:** whether an auto-reset sweeper now exists

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### pipeline column as soft routing label
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "Decision reached (2026-05-15, with user): leave the field as a soft, currently-unused routing label … Not implemented yet."
- **what:** `site_work_items.pipeline` (renamed from `domain`; default 'build') is a coarse label allowing pipeline-specific dispatchers, but only build-dispatch-loop exists — 'design' and 'maintenance' items sit dormant. It duplicates what handler_agent already implies, with nothing keeping them in sync (the unfulfilled_imagery_plan check emitted pipeline='design' and stalled). Decided: discovery checks write 'build'; loosen the dispatcher to accept any value. Stale `target_domain` config keyword survives the rename.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q4
- **relations:** unfulfilled_imagery_plan check; dispatch chain
- **verify-later:** current unfulfilled_imagery_plan INSERT; build-dispatch-loop load_items item_pipeline config

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Silent completion pathology and the positive-evidence rule
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "observed and characterised; not yet fixed" (captured 2026-04-19); mode 2 later "already confirmed fixed" per FOCUS_content_quality (2026-06-09); modes 1/3 ran at 66×/47× per week (2026-04-20)
- **what:** Three modes mark work complete that isn't: reaper auto-completion on lost responses; validate_content failures inconsistently routed to complete instead of needs_human_review; 40-minute blind reaper marking claim-timeouts complete instead of resetting to triaged. Root flaw: "we're done trying" treated as "the work is done". Fix rule: complete only on explicit success response OR positive DB evidence (page_components rows, build_status='deployed', git commit). Symptoms attempt_count=0-on-success and updated_at<claimed_at belong to the same semantic muddle.
- **sources:** FOCUS_page_build_handler_silent_completion.md (whole); HANDOFF_2026-04-20_error_investigations.md#2, #3; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#C
- **relations:** claim-timeout mechanism; validate_page_content gate; two-strike rule
- **verify-later:** reaper code paths setting status='complete'; whether modes 1 and 3 were fixed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatcher response-stall and missing claim/orchestration timeout cleanup
- **category:** NEW:work-dispatch
- **status-signal:** unknown
- **status-evidence:** "Bug 1 … Blocker for autonomous cascade completion" and "Bug 2 — No claim-timeout / orchestration-timeout cleanup" (2026-04-23); every cascade needed "manual dispatcher pokes"
- **what:** build-dispatch-loop orchestrations stall at process_item_iter_N_call_handler even when the handler response arrived (suspects: Kafka consumer reconnect failure; mark_complete not firing); with no sweeper, claimed items and AWAITING_RESPONSES orchestrations accumulate forever and block sites (compounding the NOT EXISTS blocker). Fix shapes: consumer reconnect detection, periodic claim-release sweeper, force-fail of timed-out orchestrations.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md Bugs 1-2
- **relations:** NOT EXISTS blocker; consumer group race; silent completion
- **verify-later:** any sweeper added after 2026-04-23; kafka client reconnect handling

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### build-pipeline-trigger site targeting via pre_query
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Bug 3 … Scheduler-driven dispatcher invocations all default to system.internal site_id … Fix shape: Add a pre_query" (2026-04-23)
- **what:** The scheduled dispatcher fires with no site targeting so it lands on system.internal and no-ops while real sites wait. Proposed pre_query on the scheduled_tasks row selecting sites with open build items so one dispatcher fires per site.
- **sources:** HANDOFF_2026-04-23(1).md Bug 3
- **relations:** dispatch chain; find_dispatchable_site arbitrariness
- **verify-later:** scheduled_tasks.build-pipeline-trigger pre_query column value

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two-strike rule for work items
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Two-strike rule — FINAL DECISION … Decided NOT to weaken" (2026-04-23); born-unresolved pile-up pattern noted 2026-04-17
- **what:** insertWorkItem marks a new item `unresolved` when 2 prior items with the same item_key ended (complete + failed both count), breaking discover↔fix loops. Cost: items born unresolved accumulate; re-cascades hit strikes from a previous run's completes. The sanctioned fix is item_key cascade_run_id scoping (deferred), not weakening the rule. Centralised `workItemTerminalStatuses` const (work_items_common.go) keeps the dedup index and ON CONFLICT predicates from drifting.
- **sources:** HANDOFF_2026-04-23(1).md #Two-strike, deploy table; HANDOFF-pipeline-triage-april-2026.md#patterns
- **relations:** idx_swi_dedup migration 012; discovery noise on dead sites
- **verify-later:** work_items_common.go; whether cascade_run_id scoping landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery auto-triage and scheduled-audit open questions
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Q1 — discovery emissions auto-triage (still open); Q2 — scheduled audit runs (still open)" (2026-05-15)
- **what:** Should low-risk discovery emissions (e.g. needs_imagery) auto-triage via a per-check `auto_triage_emissions` flag rather than waiting for an audit run? And is design-audit-agent scheduled anywhere, or is triage operator-driven? Both parked; determine before more discovery checks ship.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q1, #Q2
- **relations:** work-item state machine
- **verify-later:** scheduled_tasks rows for design-audit-agent; DiscoveryCheck interface

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Work-item state machine (detected → triaged → claimed → complete/failed)
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Phase 2G + 2H verified end-to-end at scale. Seven hero items … processed through the full chain … without manual intervention" (2026-05-15)
- **what:** `detected` is a valid intermediate state: discovery emits at detected; design-audit-agent's `triage_detected_items` step (registry.go:722) promotes to triaged; dispatch claims only triaged/approved (partial indexes idx_swi_handler / idx_swi_site_pending); handlers mark complete/failed (mark_work_item_complete / mark_work_item_failed steps). There is NO automated coupling between discovery and audit — items sit in detected until an audit runs. Admin-created items insert directly at triaged.
- **sources:** FOCUS_dispatch_diagnostic(4).md#TL;DR, #Evidence-trail; HANDOFF-pipeline-triage-april-2026.md
- **relations:** dispatch chain; auto-triage open question; two-strike rule; silent completion
- **verify-later:** registry.go triage_detected_items; site_work_items partial indexes; design-audit-agent workflow

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatch chain: build-pipeline-trigger → find_dispatchable_site → build-dispatch-loop
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "researched in depth this session" (2026-05-15) with the actual selection SQL quoted; scheduled_tasks row build-pipeline-trigger every 30s
- **what:** The scheduler fires build-pipeline-trigger, whose find_dispatchable_site step picks ONE site per tick (DISTINCT ON with no outer ORDER BY — effectively arbitrary among eligible sites) and spawns build-dispatch-loop scoped to it, which loads up to 5 items (pipeline='build') and claims/spawns handlers. Throughput cap ~5 items per site per 30s, one site at a time. build-pipeline-trigger doesn't write orchestration_states, making its decisions untraceable.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** NOT EXISTS blocker; Bug 3 site-targeting; fairness ORDER BY improvement
- **verify-later:** scheduled_tasks 'build-pipeline-trigger' row; build-pipeline-trigger / build-dispatch-loop agent definitions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### NOT EXISTS whole-site claim blocker
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "A single stuck claimed item on a site excludes the entire site from dispatch consideration until it clears … by design … but it makes stuck claims a system-stopping condition" (2026-05-15)
- **what:** find_dispatchable_site's NOT EXISTS clause excludes any site with ANY item in status='claimed' — an absolute blocker, not a deprioritiser. Prevents racing claims mid-execution but converts one dead handler into a site-wide stall. Proposed (cheap, high-leverage, not built): watchdog that resets claims older than ~15 min.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3
- **relations:** claim-timeout sweeper absence (Bug 2); dispatcher stall (Bug 1)
- **verify-later:** whether an auto-reset sweeper now exists

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### pipeline column as soft routing label
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "Decision reached (2026-05-15, with user): leave the field as a soft, currently-unused routing label … Not implemented yet."
- **what:** `site_work_items.pipeline` (renamed from `domain`; default 'build') is a coarse label allowing pipeline-specific dispatchers, but only build-dispatch-loop exists — 'design' and 'maintenance' items sit dormant. It duplicates what handler_agent already implies, with nothing keeping them in sync (the unfulfilled_imagery_plan check emitted pipeline='design' and stalled). Decided: discovery checks write 'build'; loosen the dispatcher to accept any value. Stale `target_domain` config keyword survives the rename.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q4
- **relations:** unfulfilled_imagery_plan check; dispatch chain
- **verify-later:** current unfulfilled_imagery_plan INSERT; build-dispatch-loop load_items item_pipeline config

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Silent completion pathology and the positive-evidence rule
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "observed and characterised; not yet fixed" (captured 2026-04-19); mode 2 later "already confirmed fixed" per FOCUS_content_quality (2026-06-09); modes 1/3 ran at 66×/47× per week (2026-04-20)
- **what:** Three modes mark work complete that isn't: reaper auto-completion on lost responses; validate_content failures inconsistently routed to complete instead of needs_human_review; 40-minute blind reaper marking claim-timeouts complete instead of resetting to triaged. Root flaw: "we're done trying" treated as "the work is done". Fix rule: complete only on explicit success response OR positive DB evidence (page_components rows, build_status='deployed', git commit). Symptoms attempt_count=0-on-success and updated_at<claimed_at belong to the same semantic muddle.
- **sources:** FOCUS_page_build_handler_silent_completion.md (whole); HANDOFF_2026-04-20_error_investigations.md#2, #3; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#C
- **relations:** claim-timeout mechanism; validate_page_content gate; two-strike rule
- **verify-later:** reaper code paths setting status='complete'; whether modes 1 and 3 were fixed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatcher response-stall and missing claim/orchestration timeout cleanup
- **category:** NEW:work-dispatch
- **status-signal:** unknown
- **status-evidence:** "Bug 1 … Blocker for autonomous cascade completion" and "Bug 2 — No claim-timeout / orchestration-timeout cleanup" (2026-04-23); every cascade needed "manual dispatcher pokes"
- **what:** build-dispatch-loop orchestrations stall at process_item_iter_N_call_handler even when the handler response arrived (suspects: Kafka consumer reconnect failure; mark_complete not firing); with no sweeper, claimed items and AWAITING_RESPONSES orchestrations accumulate forever and block sites (compounding the NOT EXISTS blocker). Fix shapes: consumer reconnect detection, periodic claim-release sweeper, force-fail of timed-out orchestrations.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md Bugs 1-2
- **relations:** NOT EXISTS blocker; consumer group race; silent completion
- **verify-later:** any sweeper added after 2026-04-23; kafka client reconnect handling

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### build-pipeline-trigger site targeting via pre_query
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Bug 3 … Scheduler-driven dispatcher invocations all default to system.internal site_id … Fix shape: Add a pre_query" (2026-04-23)
- **what:** The scheduled dispatcher fires with no site targeting so it lands on system.internal and no-ops while real sites wait. Proposed pre_query on the scheduled_tasks row selecting sites with open build items so one dispatcher fires per site.
- **sources:** HANDOFF_2026-04-23(1).md Bug 3
- **relations:** dispatch chain; find_dispatchable_site arbitrariness
- **verify-later:** scheduled_tasks.build-pipeline-trigger pre_query column value

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two-strike rule for work items
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Two-strike rule — FINAL DECISION … Decided NOT to weaken" (2026-04-23); born-unresolved pile-up pattern noted 2026-04-17
- **what:** insertWorkItem marks a new item `unresolved` when 2 prior items with the same item_key ended (complete + failed both count), breaking discover↔fix loops. Cost: items born unresolved accumulate; re-cascades hit strikes from a previous run's completes. The sanctioned fix is item_key cascade_run_id scoping (deferred), not weakening the rule. Centralised `workItemTerminalStatuses` const (work_items_common.go) keeps the dedup index and ON CONFLICT predicates from drifting.
- **sources:** HANDOFF_2026-04-23(1).md #Two-strike, deploy table; HANDOFF-pipeline-triage-april-2026.md#patterns
- **relations:** idx_swi_dedup migration 012; discovery noise on dead sites
- **verify-later:** work_items_common.go; whether cascade_run_id scoping landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery auto-triage and scheduled-audit open questions
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Q1 — discovery emissions auto-triage (still open); Q2 — scheduled audit runs (still open)" (2026-05-15)
- **what:** Should low-risk discovery emissions (e.g. needs_imagery) auto-triage via a per-check `auto_triage_emissions` flag rather than waiting for an audit run? And is design-audit-agent scheduled anywhere, or is triage operator-driven? Both parked; determine before more discovery checks ship.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q1, #Q2
- **relations:** work-item state machine
- **verify-later:** scheduled_tasks rows for design-audit-agent; DiscoveryCheck interface

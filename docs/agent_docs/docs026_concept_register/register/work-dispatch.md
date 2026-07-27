# Register — work-dispatch

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

13 concepts, consolidated from 18 raw extractions across units U01, U02, U09.
Absorbed categories: new:dispatch-pipeline, new:work-item-system (their raw material
was largely re-derivations of the same dispatch-chain/state-machine/two-strike facts
documented independently by different units; folded in below rather than kept as
separate register files).

### WDS-001 — Work-item state machine (detected → triaged → claimed → complete/failed) + site-exclusion by stuck claim
- **status:** deployed
- **status-evidence:** "Phase 2G + 2H verified end-to-end at scale. Seven hero items … processed through the full chain … without manual intervention" (2026-05-15); re-confirmed with an ownership table and a 2026-05-14 mis-diagnosis lesson in a separate unit's numbered core doc.
- **what:** `detected` is a valid intermediate state: discovery emits at detected; design-audit-agent's `triage_detected_items` step promotes to triaged; dispatch claims only triaged/approved (partial indexes idx_swi_handler/idx_swi_site_pending); handlers mark complete/failed. There is NO automated coupling between discovery and audit — items sit in detected until an audit runs. Admin-created items insert directly at triaged. The most common "won't dispatch" symptom is a single stuck `claimed` item excluding the ENTIRE site from dispatch consideration (see WDS-002's NOT EXISTS clause) — a repeatedly-relearned debugging trap: don't infer writers from readers (indexes show the read path; grep the verb for the writer).
- **sources:** FOCUS_dispatch_diagnostic(4).md#TL;DR, #Evidence-trail; HANDOFF-pipeline-triage-april-2026.md; 016#Work-item-lifecycle, #Site-excluded-from-dispatch
- **relations:** dispatch chain + NOT EXISTS blocker (WDS-002); two-strike rule (WDS-007); silent completion pathology (WDS-004)
- **verify-later:** registry.go triage_detected_items; site_work_items partial indexes; design-audit-agent workflow

### WDS-002 — Dispatch chain (build-pipeline-trigger → find_dispatchable_site → build-dispatch-loop) + NOT-EXISTS whole-site claim blocker + one-site-per-tick throughput
- **status:** deployed
- **status-evidence:** "researched in depth this session" (2026-05-15) with the actual selection SQL quoted; independently revisited 2026-06-04: "Lever C [claim watchdog]… effectively already implemented; the remaining actions are… the guardrails (timeout alignment, fairness ORDER BY, git-adapter retry), which are the real remaining dispatch work."
- **what:** The scheduler fires `build-pipeline-trigger`, whose `find_dispatchable_site` step picks ONE site per tick (`DISTINCT ON` with no outer ORDER BY — effectively arbitrary among eligible sites, and lowest-UUID sites can starve others) and spawns `build-dispatch-loop` scoped to it, which loads up to 5 items (pipeline='build') and claims/spawns handlers — throughput cap ~5 items per site per 30s, one site at a time (tools dispatch serially within a site; ~5-min Opus builds mean 11 tools ≈ an hour minimum). `find_dispatchable_site`'s `NOT EXISTS` clause excludes ANY site with an item in status='claimed' — an absolute blocker, not a deprioritiser, by design (prevents racing claims mid-execution) but converting one dead handler into a site-wide stall (observed 47-67 min gaps until the reaper resets the claim). `build-pipeline-trigger` doesn't write orchestration_states, making its decisions untraceable. Levers considered: multi-site decoupling (lower priority — cadence already 30s/8); bounded per-site concurrency (K=2-3, gated on OOM guardrails); claim watchdog (already exists — see the claimed-item-timeout entry in the work-item-integrity register). A scheduler/handler timeout mismatch (300s task vs 900/1200/4200s work) is a latent over-spawn/OOM risk.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3; HANDOFF_2026-04-23(1).md Bug 3; FOCUS_dispatch_throughput_and_claim_watchdog(3).md; CATALOGUE(9)#family-j
- **relations:** work-item state machine (WDS-001); build pump and queue immune system (build-pipeline register, BLD-003); evidence-gated claimed-item-timeout reaper (work-item-integrity register)
- **verify-later:** find_dispatchable_site SQL (NOT EXISTS clause); scheduled_tasks build-pipeline-trigger row; handler pod requests/limits

### WDS-003 — pipeline column as soft routing namespace/label
- **status:** partial
- **status-evidence:** "Decision reached (2026-05-15, with user): leave the field as a soft, currently-unused routing label … Not implemented yet"; an earlier doc frames the same column as an already-settled dispatch namespace ("Everything in the initial build must be pipeline='build'").
- **what:** `site_work_items.pipeline` (renamed from `domain` — clash with the website's own domain concept; default 'build') is a coarse label allowing pipeline-specific dispatchers, but only `build-dispatch-loop` exists — 'design' and 'maintenance' items sit dormant. It duplicates what `handler_agent` already implies, with nothing keeping them in sync (a discovery check once emitted pipeline='design' and stalled). Decided: discovery checks write 'build'; loosen the dispatcher to accept any value. Historical trap: dispatch once passed the namespace to handlers as the site domain; a stale `target_domain` config keyword survives the rename.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q4; 001(5)#Work-item-pipeline-must-be-"build"; 016 Schema reminders
- **relations:** unfulfilled_imagery_plan check; dispatch chain (WDS-002)
- **verify-later:** current unfulfilled_imagery_plan INSERT; build-dispatch-loop load_items item_pipeline config

### WDS-004 — Silent completion pathology and the positive-evidence rule
- **status:** deployed
- **status-evidence:** "observed and characterised; not yet fixed" (captured 2026-04-19); one mode later "already confirmed fixed" (2026-06-09); the other two modes ran at 66x/47x per week as of 2026-04-20.
- **stage2-verified (2026-07-14):** partial → deployed — All 3 modes now confirmed fixed in code: (1) validate_page_content.go:71 routes failures to needs_human_review (not silent complete); (2) scheduled_tasks 'claimed-item-timeout' pre_query (docs/agent_docs/sql_for_agents/052_build_pipeline_trigger.sql:560-586) resets stale claims to triaged / fails at max_attempts ins...
- **what:** Three modes mark work complete that isn't: reaper auto-completion on lost responses; validate_content failures inconsistently routed to complete instead of needs_human_review; a 40-minute blind reaper marking claim-timeouts complete instead of resetting to triaged. Root flaw: "we're done trying" treated as "the work is done." Fix rule: complete only on explicit success response OR positive DB evidence (page_components rows, build_status='deployed', git commit). Symptoms attempt_count=0-on-success and updated_at<claimed_at belong to the same semantic muddle. This is the work-dispatch-side statement of the same defect family documented in full in the work-item-integrity register's silent-completion entries.
- **sources:** FOCUS_page_build_handler_silent_completion.md; HANDOFF_2026-04-20_error_investigations.md#2, #3; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#C
- **relations:** silent-completion failure family (work-item-integrity register); claim-timeout mechanism (work-item-integrity register); two-strike rule (WDS-007)
- **verify-later:** reaper code paths setting status='complete'; whether the two unresolved modes were fixed

### WDS-005 — Dispatcher response-stall and missing claim/orchestration timeout cleanup
- **status:** partial
- **status-evidence:** "Bug 1 … Blocker for autonomous cascade completion" and "Bug 2 — No claim-timeout / orchestration-timeout cleanup" (2026-04-23); every cascade needed "manual dispatcher pokes."
- **stage2-verified (2026-07-14):** unknown → partial — Bug 2 (no claim/orchestration-timeout cleanup) resolved: coordinator.go:777-800 force-fails orchestrations exceeding 3x workflow timeout / 60min fallback (git blame 8c60e8f46 'timeout rationalisation'), plus the claimed-item-timeout reaper (WII-002) resets stuck claims. Bug 1 root cause (Kafka consumer reconnect fai...
- **what:** `build-dispatch-loop` orchestrations stall at `process_item_iter_N_call_handler` even when the handler response arrived (suspects: Kafka consumer reconnect failure; mark_complete not firing); with no sweeper, claimed items and AWAITING_RESPONSES orchestrations accumulate forever and block sites (compounding the NOT EXISTS blocker, WDS-002). Fix shapes: consumer reconnect detection, periodic claim-release sweeper, force-fail of timed-out orchestrations.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md Bugs 1-2
- **relations:** NOT EXISTS blocker (WDS-002); consumer group race; silent completion (WDS-004)
- **verify-later:** any sweeper added after 2026-04-23; kafka client reconnect handling

### WDS-006 — build-pipeline-trigger site targeting via pre_query
- **status:** aspirational
- **status-evidence:** "Bug 3 … Scheduler-driven dispatcher invocations all default to system.internal site_id … Fix shape: Add a pre_query" (2026-04-23).
- **what:** The scheduled dispatcher fires with no site targeting so it lands on `system.internal` and no-ops while real sites wait. Proposed pre_query on the scheduled_tasks row selecting sites with open build items so one dispatcher fires per site.
- **sources:** HANDOFF_2026-04-23(1).md Bug 3
- **relations:** dispatch chain (WDS-002); find_dispatchable_site arbitrariness
- **verify-later:** scheduled_tasks.build-pipeline-trigger pre_query column value

### WDS-007 — Two-strike rule for work items (dedup/anti-churn)
- **status:** deployed
- **status-evidence:** "Two-strike rule — FINAL DECISION … Decided NOT to weaken" (2026-04-23); born-unresolved pile-up pattern noted 2026-04-17; migration 006 "Two-strike dedup [check]. Deployed."
- **what:** `insertWorkItem` marks a new item `unresolved` when 2 prior items with the same item_key ended (complete + failed both count), breaking discover↔fix loops; items born unresolved accumulate, and re-cascades hit strikes from a previous run's completes. The sanctioned fix is item_key cascade_run_id scoping (deferred), not weakening the rule. A centralised `workItemTerminalStatuses` const keeps the dedup index and ON CONFLICT predicates from drifting. `wont_fix` + "superseded by active duplicate" is the dedup system working as designed, not a bug.
- **sources:** HANDOFF_2026-04-23(1).md #Two-strike; 006 issues table; 001(5)#Work-Item-Lifecycle; 016 §9 wont_fix entry
- **relations:** idx_swi_dedup (work-item-integrity register); feasibility-recheck; silent completion pathology (WDS-004)
- **verify-later:** work_items_common.go; whether cascade_run_id scoping landed

### WDS-008 — Discovery auto-triage and scheduled-audit open questions
- **status:** aspirational
- **status-evidence:** "Q1 — discovery emissions auto-triage (still open); Q2 — scheduled audit runs (still open)" (2026-05-15).
- **what:** Should low-risk discovery emissions (e.g. needs_imagery) auto-triage via a per-check `auto_triage_emissions` flag rather than waiting for an audit run? And is design-audit-agent scheduled anywhere, or is triage operator-driven? Both parked; determine before more discovery checks ship.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q1, #Q2
- **relations:** work-item state machine (WDS-001)
- **verify-later:** scheduled_tasks rows for design-audit-agent; DiscoveryCheck interface

### WDS-009 — Terminal work items: every pipeline ends with assembly + deployment
- **status:** deployed
- **status-evidence:** 001(5) rule with the minimum brochure chain; 004 step 9 sets needs_rerender priority 99.
- **what:** The planner/WriteBuildItems must create terminal items (needs_rerender, priority ~20/99) that re-render site components, assemble pages from page_components + site_components, git commit, and trigger deploy — otherwise the pipeline produces data but no website.
- **sources:** 001(5)#Every-pipeline-must-end-with-assembly; 004 step 9
- **relations:** commit-is-deploy; rerender agents (page-build-pipeline register)
- **verify-later:** WriteBuildItemsAction

### WDS-010 — Unified build & maintenance via site_work_items (single queue, same code)
- **status:** deployed
- **status-evidence:** 002(4): "Build and maintenance are the same process"; live lifecycle steps 3/7/11 "same code."
- **what:** Every piece of work is a `site_work_items` row (source, pipeline, item_type, severity, spec jsonb, handler_agent, status enum incl. needs_human_review, priority, depends_on uuid[], item_key dedup, result with commit_sha). New site = planner-written items; maintenance = discovery-written items; same orchestrator/dispatch/handlers. Cross-domain coordination happens only through the table (side_effect items with parent_item_id), never agent-to-agent calls.
- **sources:** 002(4)#Unified-Build-and-Maintenance, #site_work_items; 003(8)#Cross-Domain-Coordination
- **relations:** dispatch chain (WDS-002); site-work-orchestrator (build-pipeline register, BLD-009); side-effect rules engine (WDS-013)
- **verify-later:** site_work_items schema incl. depends_on, item_key

### WDS-011 — Work-item routing: content rebuild vs re-render (needs_page / page_rerender / needs_rerender / link_resolution_rebuild)
- **status:** deployed
- **status-evidence:** 002(4) routing table, dated with a 2026-06-22 hazard confirmation.
- **what:** Route by item_type, never item_key; the distinction is whether copy is regenerated. `needs_page`/`needs_content_page` → full LLM rebuild via page-build-handler; `page_rerender` (reason image_landed/section_data_resolved) → no-LLM re-resolve + re-render from stored content_data; `needs_rerender` → batch reassembly; `link_resolution_rebuild` is INTENDED links-only but runs the full writer (a documented hazard — see interactive-page clobber in the page-build-pipeline register). `page_rerender` on a NULL-content_data section escalates the page to `needs_page` (backfills content_data).
- **sources:** 002(4)#Work-item-routing; 003(8)#Source-of-truth-principle
- **relations:** work-item routing map (work-item-integrity register); interactive-page clobber (page-build-pipeline register, PBP-012)
- **verify-later:** rerender_page_sections action; flag_page_image_rebuild/reconcile_section_data emitters

### WDS-012 — needs_section_data: resolvable-by-query vs genuinely-human classification
- **status:** partial
- **status-evidence:** 016 §9 (2026-05-27 direction): pages_where_type implemented, pages_under_section named-but-unimplemented at the time; a reconciler direction recorded in FUTURE_section_data_handler.
- **what:** Some `needs_section_data` items are human-only (team, pricing); list/grid sections source from `query.*` and resolve mechanically once pages exist — but an unimplemented query name or pre-active timing defers them anyway. Read `spec.missing[].source` to classify. Direction: implement `pages_under_section` + a lightweight resolver (not an LLM agent) that re-attempts open items via queryresolve, closes via `closeResolvedDataRequest`, and flags re-render. This is the work-item-classification angle of the same mechanism documented architecturally in the site-plan-and-reconciler register (PLAN-009).
- **sources:** 016 §9 needs_section_data entry; 002(4) reconcile_section_data → page_rerender emitters
- **relations:** section-data deferral + reconciler (site-plan-and-reconciler register, PLAN-009); input schema v2
- **verify-later:** queryresolve vocabulary today

### WDS-013 — Side-effect rules engine (deterministic follow-on items)
- **status:** partial
- **status-evidence:** A Phase-1 table of triggers exists; some pieces are live (side_effect source appears in the 002/003 contracts), but a full Go rules engine is unconfirmed.
- **what:** After each handler completes, deterministic Go rules (not LLM) are meant to emit follow-ons: new page → needs_nav_update + needs_sitemap; deletion → redirects; CSS change → needs_rerender; milestone item types → needs_snapshot.
- **sources:** P1#Side-effects; 003(8)#Cross-Domain-Coordination
- **relations:** unified build & maintenance (WDS-010); cross-domain coordination; snapshot triggers
- **verify-later:** rules engine implementation

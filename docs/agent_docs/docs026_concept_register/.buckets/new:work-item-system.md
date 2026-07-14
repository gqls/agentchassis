
<!-- SOURCE: U01_docs024_numbered_core.md -->
### `pipeline` column as work-item routing namespace
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5): renamed from `domain` (clash with site domain); dispatch loop filters pipeline='build'
- **what:** site_work_items.pipeline ('build','design','maintenance'…) is the dispatch routing namespace, never the website domain. Everything in the initial build must be pipeline='build'; items emitted straight-to-triaged must set it at emission (triage rewrites it for detected items). Historical trap: dispatch once passed the namespace to handlers as the site domain.
- **sources:** 001(5)#Work item pipeline must be "build"; 016 Schema reminders
- **relations:** dispatch loop; schema-drift rule
- **verify-later:** find_dispatchable_site / LoadWorkItemsAction filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Terminal work items: every pipeline ends with assembly + deployment
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5) rule with the minimum brochure chain
- **what:** The planner/WriteBuildItems must create terminal items (needs_rerender, priority ~20/99) that re-render site components, assemble pages from page_components + site_components, git commit, and trigger deploy — otherwise the pipeline produces data but no website.
- **sources:** 001(5)#Every pipeline must end with assembly; 004 step 9 (needs_rerender priority 99)
- **relations:** commit-is-deploy; rerender agents
- **verify-later:** WriteBuildItemsAction

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item blocking/unblocking and the `unresolved` two-strike mechanism
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 006 "Two-strike dedup ✅ Deployed"; 001(5) describes insertWorkItem mechanics
- **what:** Three block causes (missing handler → feasibility-recheck auto-promotes; spec blocked → manual; manual). insertWorkItem suppresses re-detection within 3h of a terminal duplicate and, after 2+ terminal attempts in 7 days, creates new items as status `unresolved` — visible, not dispatched. `wont_fix` + "superseded by active duplicate" is the dedup system working, not a bug.
- **sources:** 001(5)#Work Item Lifecycle; 006 issues table (12k duplicate cleanup); 016 §9 wont_fix entry
- **relations:** idx_swi_dedup; feasibility-recheck
- **verify-later:** load_work_item_actions.go insertWorkItem

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Unified build & maintenance via site_work_items (single queue, same code)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4): "Build and maintenance are the same process"; live lifecycle steps 3/7/11 "same code"
- **what:** Every piece of work is a site_work_items row (source, pipeline, item_type, severity, spec jsonb, handler_agent, status enum incl. needs_human_review, priority, depends_on uuid[], item_key dedup, result with commit_sha). New site = planner-written items; maintenance = discovery-written items; same orchestrator/dispatch/handlers. Cross-domain coordination happens only through the table (side_effect items with parent_item_id), never agent-to-agent calls.
- **sources:** 002(4)#Unified Build and Maintenance, #site_work_items; 003(8)#Cross-Domain Coordination
- **relations:** dispatch loop; work-item state machine (016); P1 expansion sources table
- **verify-later:** site_work_items schema incl. depends_on, item_key

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item routing: content rebuild vs re-render (needs_page / page_rerender / needs_rerender / link_resolution_rebuild)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4) routing table dated with 2026-06-22 hazard confirmation
- **what:** Route by item_type never item_key; the distinction is whether copy is regenerated. needs_page/needs_content_page → full LLM rebuild via page-build-handler; page_rerender (reason image_landed/section_data_resolved) → no-LLM re-resolve + re-render from stored content_data; needs_rerender → batch reassembly; link_resolution_rebuild is INTENDED links-only but runs the full writer (hazard). page_rerender on a NULL-content_data section escalates the page to needs_page (backfills content_data).
- **sources:** 002(4)#Work-item routing; 003(8)#Source of truth principle (two re-render paths)
- **relations:** interactive-page de-tool hazard; content_data source of truth
- **verify-later:** rerender_page_sections action; flag_page_image_rebuild/reconcile_section_data emitters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item state machine, transition ownership, and site-exclusion by stuck claim
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 016 dedicated section with ownership table and the 2026-05-14 mis-diagnosis lesson
- **what:** detected→(design-audit-agent triage_detected_items)→triaged→(build-dispatch-loop claim)→claimed→(handler)→complete/failed; admin inserts at triaged. Most common "won't dispatch" cause: one stuck claimed item excludes the ENTIRE site via find_dispatchable_site's NOT EXISTS. Debugging trap #1: don't infer writers from readers (indexes show the read path; grep the verb for the writer).
- **sources:** 016#Work item lifecycle; #Site excluded from dispatch
- **relations:** two-strike; claimed-item-timeout; operator reset SQL
- **verify-later:** triage_detected_items registration (registry.go:722)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### needs_section_data: resolvable-by-query vs genuinely-human, and the section-data reconciler direction
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** 016 §9 (2026-05-27 direction): pages_where_type implemented, pages_under_section named-but-unimplemented; reconciler recorded in FUTURE_section_data_handler
- **what:** Some needs_section_data items are human-only (team, pricing); list/grid sections source from query.* and resolve mechanically once pages exist — but the unimplemented query name or pre-active timing defers them anyway. Read spec.missing[].source to classify. Direction: implement pages_under_section + a lightweight resolver (not an LLM agent) that re-attempts open items via queryresolve, closes via closeResolvedDataRequest, and flags re-render.
- **sources:** 016 §9 needs_section_data entry; 002(4) reconcile_section_data → page_rerender emitters
- **relations:** input schema v2; deferral-drop bug
- **verify-later:** queryresolve vocabulary today

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Side-effect rules engine (deterministic follow-on items)
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** P1 table of triggers; some live (side_effect source in 002/003 contracts), full Go rules engine unconfirmed
- **what:** After each handler completes, deterministic Go rules (not LLM) emit follow-ons: new page → needs_nav_update + needs_sitemap; deletion → redirects; CSS change → needs_rerender; milestone item types → needs_snapshot.
- **sources:** P1#Side effects; 003(8)#Cross-Domain Coordination
- **relations:** cross-domain coordination; snapshots triggers
- **verify-later:** rules engine implementation

<!-- SOURCE: U01_docs024_numbered_core.md -->
### `pipeline` column as work-item routing namespace
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5): renamed from `domain` (clash with site domain); dispatch loop filters pipeline='build'
- **what:** site_work_items.pipeline ('build','design','maintenance'…) is the dispatch routing namespace, never the website domain. Everything in the initial build must be pipeline='build'; items emitted straight-to-triaged must set it at emission (triage rewrites it for detected items). Historical trap: dispatch once passed the namespace to handlers as the site domain.
- **sources:** 001(5)#Work item pipeline must be "build"; 016 Schema reminders
- **relations:** dispatch loop; schema-drift rule
- **verify-later:** find_dispatchable_site / LoadWorkItemsAction filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Terminal work items: every pipeline ends with assembly + deployment
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5) rule with the minimum brochure chain
- **what:** The planner/WriteBuildItems must create terminal items (needs_rerender, priority ~20/99) that re-render site components, assemble pages from page_components + site_components, git commit, and trigger deploy — otherwise the pipeline produces data but no website.
- **sources:** 001(5)#Every pipeline must end with assembly; 004 step 9 (needs_rerender priority 99)
- **relations:** commit-is-deploy; rerender agents
- **verify-later:** WriteBuildItemsAction

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item blocking/unblocking and the `unresolved` two-strike mechanism
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 006 "Two-strike dedup ✅ Deployed"; 001(5) describes insertWorkItem mechanics
- **what:** Three block causes (missing handler → feasibility-recheck auto-promotes; spec blocked → manual; manual). insertWorkItem suppresses re-detection within 3h of a terminal duplicate and, after 2+ terminal attempts in 7 days, creates new items as status `unresolved` — visible, not dispatched. `wont_fix` + "superseded by active duplicate" is the dedup system working, not a bug.
- **sources:** 001(5)#Work Item Lifecycle; 006 issues table (12k duplicate cleanup); 016 §9 wont_fix entry
- **relations:** idx_swi_dedup; feasibility-recheck
- **verify-later:** load_work_item_actions.go insertWorkItem

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Unified build & maintenance via site_work_items (single queue, same code)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4): "Build and maintenance are the same process"; live lifecycle steps 3/7/11 "same code"
- **what:** Every piece of work is a site_work_items row (source, pipeline, item_type, severity, spec jsonb, handler_agent, status enum incl. needs_human_review, priority, depends_on uuid[], item_key dedup, result with commit_sha). New site = planner-written items; maintenance = discovery-written items; same orchestrator/dispatch/handlers. Cross-domain coordination happens only through the table (side_effect items with parent_item_id), never agent-to-agent calls.
- **sources:** 002(4)#Unified Build and Maintenance, #site_work_items; 003(8)#Cross-Domain Coordination
- **relations:** dispatch loop; work-item state machine (016); P1 expansion sources table
- **verify-later:** site_work_items schema incl. depends_on, item_key

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item routing: content rebuild vs re-render (needs_page / page_rerender / needs_rerender / link_resolution_rebuild)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4) routing table dated with 2026-06-22 hazard confirmation
- **what:** Route by item_type never item_key; the distinction is whether copy is regenerated. needs_page/needs_content_page → full LLM rebuild via page-build-handler; page_rerender (reason image_landed/section_data_resolved) → no-LLM re-resolve + re-render from stored content_data; needs_rerender → batch reassembly; link_resolution_rebuild is INTENDED links-only but runs the full writer (hazard). page_rerender on a NULL-content_data section escalates the page to needs_page (backfills content_data).
- **sources:** 002(4)#Work-item routing; 003(8)#Source of truth principle (two re-render paths)
- **relations:** interactive-page de-tool hazard; content_data source of truth
- **verify-later:** rerender_page_sections action; flag_page_image_rebuild/reconcile_section_data emitters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item state machine, transition ownership, and site-exclusion by stuck claim
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 016 dedicated section with ownership table and the 2026-05-14 mis-diagnosis lesson
- **what:** detected→(design-audit-agent triage_detected_items)→triaged→(build-dispatch-loop claim)→claimed→(handler)→complete/failed; admin inserts at triaged. Most common "won't dispatch" cause: one stuck claimed item excludes the ENTIRE site via find_dispatchable_site's NOT EXISTS. Debugging trap #1: don't infer writers from readers (indexes show the read path; grep the verb for the writer).
- **sources:** 016#Work item lifecycle; #Site excluded from dispatch
- **relations:** two-strike; claimed-item-timeout; operator reset SQL
- **verify-later:** triage_detected_items registration (registry.go:722)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### needs_section_data: resolvable-by-query vs genuinely-human, and the section-data reconciler direction
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** 016 §9 (2026-05-27 direction): pages_where_type implemented, pages_under_section named-but-unimplemented; reconciler recorded in FUTURE_section_data_handler
- **what:** Some needs_section_data items are human-only (team, pricing); list/grid sections source from query.* and resolve mechanically once pages exist — but the unimplemented query name or pre-active timing defers them anyway. Read spec.missing[].source to classify. Direction: implement pages_under_section + a lightweight resolver (not an LLM agent) that re-attempts open items via queryresolve, closes via closeResolvedDataRequest, and flags re-render.
- **sources:** 016 §9 needs_section_data entry; 002(4) reconcile_section_data → page_rerender emitters
- **relations:** input schema v2; deferral-drop bug
- **verify-later:** queryresolve vocabulary today

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Side-effect rules engine (deterministic follow-on items)
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** P1 table of triggers; some live (side_effect source in 002/003 contracts), full Go rules engine unconfirmed
- **what:** After each handler completes, deterministic Go rules (not LLM) emit follow-ons: new page → needs_nav_update + needs_sitemap; deletion → redirects; CSS change → needs_rerender; milestone item types → needs_snapshot.
- **sources:** P1#Side effects; 003(8)#Cross-Domain Coordination
- **relations:** cross-domain coordination; snapshots triggers
- **verify-later:** rules engine implementation

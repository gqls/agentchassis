
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Interactive-page de-tool hazard (content rebuild silently drops a tool/game)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** CONFIRMED 2026-06-22 (gamesdesign game-pathfinding, 18KB A* canvas overwritten 06-14); "fix pending" in 002/005/016; 016b v2 update: two-layer save_page_sections fix WRITTEN, un-deployed
- **what:** A tool lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page or link_resolution_rebuild) regenerates from plan_sections and replaces it with generic-text-block; the prose-based content-regression guard doesn't catch markup/JS loss. Fix layers: interactivity-aware save guard + carry-forward of interactive sections in save_page_sections (written), source_item_id stamping into page_component_history, and routing link maintenance through a preserve-sections path (page_rerender ruled out for CTA re-resolution — it doesn't re-run link logic).
- **sources:** 005(1) hazard block; 002(4)#Interactive-page hazard; 016 §9 final entry; 016b Part 4
- **relations:** phantom-CTA resolution bug (separate); tool recreation mis-key (Part 3)
- **verify-later:** save_page_sections_action.go patched version deployed?; page_component_history.source_item_id population

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool doc header (sentinel comment; stripped at deploy)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 019 full lifecycle table incl. StripToolDocHeader call sites and tool_health checks
- **what:** Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never ids/dates; no */). It never ships: StripToolDocHeader runs at the three outbound assembly points; DB rendered_html retains it for audit parity. Creation gate validates presence; improver preserves/updates it; auditor audits code AGAINST its stated behaviour; malformed (opener without closer) is left in and flagged by tool_health.
- **sources:** 019#Tool Doc Header; 020 tool_health tier-1 checks
- **relations:** per-tool travelling docs (037); tool-auditor
- **verify-later:** platform/content/tool_doc_header.go; prompt migration applied

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool quality three tiers (structural / LLM audit / headless-browser future)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 020: tiers 1–2 automated live, tier 3 "planned"
- **what:** tool_health tier 1 (Go, free): deploy status, HTML/template present, script/style/@media, hex/external-dep warnings, doc-header checks — blockers create improve_tool. Tier 2: audit_tool queued (30-day/tool cooldown) → tool-auditor Sonnet code review across six categories, findings by confidence (certain/likely → improve_tool, possible → needs_human_review), quality_score 1-10 tracked. Tool removal is a human decision via dashboard.
- **sources:** 020#tool_health, #tool-auditor; 019#Tool Quality Standards
- **relations:** tool-improver; component-quality-auditor (sections)
- **verify-later:** check_tool_health.go cooldowns

<!-- SOURCE: U04_idea_uk.md -->
### Content-rebuild de-tools tool pages (confirmed hazard)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "confirmed hazard, fix pending" (TODO P3, 2026-06-26).
- **what:** A needs_page / link_resolution_rebuild on a tool or game page regenerates the page from plan_sections, and the plan knows nothing about the interactive tool living in a section's rendered_html — so the tool is silently replaced with generated prose. Fix direction: route link maintenance through a preserve-sections re-render path, stamp source_item_id, add an interactivity-aware save guard. Flagged as a direct risk to idea.uk's post-P0 rebuild if tools land first.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md#P3; idea.uk/running_notes_2(6).md (backlog)
- **relations:** tool pipeline (005/016b/020/026 cross-refs); page-rerender vs page-build-handler distinction.
- **verify-later:** whether the preserve-sections path landed.

<!-- SOURCE: U05_content_quality_linking.md -->
### tool-recreation-handler (interactive rebuild path)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-26: "WI 6a28c4b3 completed … the playable A* tool is BACK".
- **what:** Recreates interactive tool/game pages from the adoption crawl: recreate_tool (Opus 64k, timeout 2400s) → check_tool_completeness → validate_tool → save_page_sections → update_page_status → spawn_rerender. Mode must be `recreate_tool` (not `recreate` — load_existing_content skips unless mode matches, a prior gotcha). Interactive pages are routed here from adoption via buildPageFeatureMap (T1 routing fix). One of the three save_page_sections callers, so it carries the Part-4 guards; re-creating a tool doesn't trip Layer 1 (new content IS interactive).
- **sources:** PLAN_pathfinding_missing_game.md#2; NOTES(44) 2026-06-25/26; HANDOFF_page_pipeline(11).md#4
- **relations:** interactive clobber; adoption pipeline T1 routing; item_key mis-key.
- **verify-later:** tool-recreation-handler default_config; apply_adoption_plan buildPageFeatureMap.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECISION: acceptance criteria live in the tool's doc_plans PLAN" (2026-07-04 rev 4); consumed live by Tier 2 and Tier 4 checkers.
- **what:** The per-tool definition of *working*. Candidates judged on key/lifecycle/owner: `site_specs` — right machinery, wrong key (site-scoped; per-site copies drift); `site_plans`/directives — wrong lifecycle and owner (churniest artifact, planner-owned; "never store the bar in the artifact that regenerates most"); findings' `acceptance_test` — right pattern, wrong duration (dies with the work item; the standing criteria SEED it). The PLAN wins on all three axes. Format: a machine-extractable fenced ```criteria JSON block (tool-doc-header precedent), extracted by `load_doc_context` as `criteria_json`; lifts to a column only on volume. Per-site parametrisation goes to `direction.must_have`, not the PLAN.
- **sources:** PLAN_travelling_docs(6).md#where-acceptance-criteria-live; 001_README_acceptance_criteria.md; RUNNING_NOTES_travelling_docs(39).md#rev4
- **relations:** verification ladder; findings acceptance_test/max_fix_attempts (improvement-loop 004); direction.must_have.
- **verify-later:** criteria fence extraction in load_doc_context; `has_fence` on live PLANs.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria describe DELIVERED reality, not aspiration (Option B)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECIDED 2026-07-10: Option B — inline reality (user: 'I choose option B and surrender')"; migrations 143 (PLANs superseded) + 144 (composer fixed) applied and verified.
- **what:** The composer had asserted a designed-but-never-built JS extraction (`asset_loads /tools/assets/<fn>.js`) in every PLAN; Tier-2's first sweep failed every tool on it by construction. Principle on record: criteria must describe what the system delivers; aspirations live in roadmaps. If extraction ever ships, PLANs supersede forward again. Corollary: the composer's standard checks became boots/console/status/mobile-fit + optional interaction from real selectors.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5 (Option B block); HANDOFF_2026-07-10…md#§2; PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** js-not-extracted delivery gap; Tier-2 first sweep; PLAN supersede versioning.
- **verify-later:** current PLAN fences have no asset check; compose_plan prompt (four checks, inline delivery line).

<!-- SOURCE: U08_travelling_docs.md -->
### The tool verification ladder (Tiers 0–4)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "The verification ladder is whole (Tier 0/1/2/4) and closes on both outcomes" — RUNBOOK position 2026-07-12; Tier 3 remains Phase B.
- **what:** Cheap-to-expensive tiers, each catching a different class: Tier 0 generation-time output integrity (`HasToolDocHeader` gate + `check_tool_completeness`, deliberately flags-but-passes); Tier 1 structural post-deploy (`check_tool_health`); Tier 2 static contract-presence against deployed HTML (anchor rule); Tier 3 acceptance audit (`tool-auditor` vs criteria — Phase B, unbuilt extension); Tier 4 behavioural — drive the deployed tool in headless Chromium until criteria pass. Standing rule: never read a Tier-2 pass as "the tool works" — that claim belongs to Tier 4.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_travelling_docs(6).md#tool-assurance; OVERVIEW_self_verifying_tools.md#mechanism-2
- **relations:** every tier concept below; "passed checks ≠ working".
- **verify-later:** check_tool_completeness + check_tool_health + discovery_checks/check_tool_acceptance.go + browser-runner adapter, all in the chassis repo.

<!-- SOURCE: U08_travelling_docs.md -->
### "Completeness + validation passed" ≠ working — twice demonstrated
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN(6) rollout outcomes: "The June recreation introduced the economy-simulator's two bugs and passed; run 2 of the repair faithfully recreated them and passed."
- **what:** The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity, not behaviour. The same game shipped broken twice while passing every existing check — the June 2026-06-05 recreation introduced the bugs (proven from tool_recreation_training rows and the origin game.js which has neither bug), and repair run 2 recreated them while its own note truthfully said "completeness + validation passed".
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45; OVERVIEW_self_verifying_tools.md#problem
- **relations:** Tier 4; seam rule; economy-simulator case.
- **verify-later:** tool_recreation_training rows for page d9a8e6e8 dated 2026-06-05.

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-2 static acceptance checker (discovery check `tool_acceptance`)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Stage 5 — LIVE 2026-07-10 ✓ (first sweep proven)" — run cd0d9731 on v1.0.1107 produced exactly the pre-verified findings (2 improve_tool items + 2 acceptance-fail notes, check-level precision confirmed).
- **what:** A browserless discovery check (sibling of `tool_health` in `discovery_checks/`): loads the current PLAN's criteria fence, fetches the deployed page (bounded 12s/2MB, cached per run), and evaluates the statically-visible subset under the anchor rule, plus shell checks (tool-doc header not leaked, no `<no value>` residue). No criteria → a `needs_criteria` note (30-day cooldown), never a fake pass. Failures → one improve_tool item (criteria embedded as `acceptance_test`, 7-day cooldown, cancelled items excluded since migration 146's correct-while-touching) + an acceptance-fail note. Scope limit by construction: only generator-created tools have content_components rows; adopted/recreated page-section tools are invisible to Tier 2.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5; RUNNING_NOTES_travelling_docs(39).md#stage-5-built,#stage-5-live; HANDOFF_2026-07-10…md#§1,§2
- **relations:** anchor rule; migration 142; Tier 4 (reaches page-section tools via pages).
- **verify-later:** `discovery_checks/check_tool_acceptance.go`; design-discovery-agent run_checks list.

<!-- SOURCE: U08_travelling_docs.md -->
### The anchor rule — static checks confirm, never refute
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "STAGE-5 RULE SETTLED" 2026-07-09 after the #tableWrap inspection (empty div filled by JS); implemented + unit-tested incl. the founding cases.
- **what:** Validate only a criteria selector's ANCHOR (leftmost id/class token) against `html_template`: `#tableWrap` exists ⇒ `#tableWrap tr` passes (rows are JS-built; Tier 4 asserts them for real); `#xpTableBody` exists nowhere ⇒ fails ⇒ drop or -EDIT. Static validation can confirm a selector but never refute one — never delete a check merely because the DOM is constructed at runtime. Motivated by the composer inventing selectors it ASSERTS on while copying real ones it ACTS on; the remedy is a check made by the system on itself, not a sterner prompt. Implementation detail banked: CSS class tokens are whitespace-delimited (Go regexp `\b` wrongly splits on hyphens).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5-rule; RUNNING_NOTES_travelling_docs(39).md#rev39,#rev40; OVERVIEW_self_verifying_tools.md#tier-2
- **relations:** composer selector-invention incident; Tier 4 runtime assertions; tool-auditor (same logic belongs there — unbuilt).
- **verify-later:** anchor extraction + class-token comparison in check_tool_acceptance.go tests.

<!-- SOURCE: U08_travelling_docs.md -->
### Composer selector invention — caught twice, machine-corrected
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** CONFIRMED 2026-07-09 (`#xpTableBody`/`#statsStrip` token_anywhere=f); second sighting caught by Tier 2 itself 2026-07-10 (kebab `#drop-chance` vs real camelCase `#dropChance`).
- **what:** The PLAN-composer LLM invented DOM ids for assertion targets despite an explicit "never invent a selector" instruction — the rule held for controls it acts on and failed for things it asserts on. First instance corrected by a guarded supersede migration that itself initially refused a valid runtime selector (leading to the anchor rule); second caught automatically by the live Tier-2 sweep and corrected by migration 143. Demonstrates the design stance: hallucination is countered by verification, not prompt escalation.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#stage-5-live; 0NN_supersede_xp_curve_plan_selectors(2).sql; RUNBOOK_travelling_docs(38).md#task-3-proven
- **relations:** anchor rule; supersede versioning (correction recorded as a NOTES entry — "the travelling-docs loop applied to itself").
- **verify-later:** xp-curve PLAN v1→v2 chain + its correction note in doc_notes.

<!-- SOURCE: U08_travelling_docs.md -->
### tool-acceptance-agent — Tier 4 self-driving orchestrator
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** First machine acceptance-run note (run bf330ac6, 2026-07-12, "Tier-4 acceptance PASSED — all 3 evaluated checks"); fail path proven live via a controlled reverted test (failed=1, improve_tool_created=true, full teardown verified).
- **what:** An agent (migration 145) closing the loop with zero humans: `ensure_site_record → load_docs → request_browser_run (Kafka await; resolves the tool's deployed URL from pages itself; NO-OP skips without awaiting when the PLAN has no criteria) → judge_acceptance_results → complete`. Judge recomputes the verdict from results: all pass → acceptance-run note; any fail → acceptance-fail note + ONE improve_tool item (criteria embedded as acceptance_test, handler tool-improver); component-less recreated/adopted tools get the note but no item — logged honestly for manual routing. Trigger 087 (dry-run default).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tool-acceptance-agent-built,#tier-4-self-driving,#fail-path-proven; README_summary_paragraph2_for_discussion.md; 087_TRIGGER_tool_acceptance.sh (header)
- **relations:** browser-runner adapter; acceptance iteration loop; continuous sweep.
- **verify-later:** `platform/orchestration/actions/tool_acceptance_actions.go`; agent_definitions row tool-acceptance-agent; migration 145.

<!-- SOURCE: U08_travelling_docs.md -->
### Continuous acceptance — the `tool_acceptance_due` periodic sweep
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Built + migration 146 applied 2026-07-12, but "v1.0.1111 … the continuous sweep is NOT in the binary" (untracked-file trap); "GATE: continuous acceptance activates on the next image built from 83ba9bd4+" (T11, 2026-07-13 — state at unit close).
- **what:** A discovery check that emits one `acceptance_run` work item per active tool with a deployed page and current PLAN criteria, unless a verdict landed within 7 days or a run is open. Design calls: post-creation/post-improve hooks deliberately NOT used (they'd fire before the page redeploys — creation ends at 'planned', improve merely queues a rerender; the sweep only ever sees deployed pages); items emitted straight to `triaged` (acceptance needs no human judgment; `detected` items were observed sitting unswept); priority 90 so acceptance tests the NEW page after builds/rerenders.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tier-4-continuous,#v1.0.1111; HANDOFF_2026-07-10…md#T10,T11; README_summary_paragraph2_for_discussion.md
- **relations:** tool-acceptance-agent; untracked-file deploy trap; improve_tool cooldown (cancelled items excluded).
- **verify-later:** `discovery_checks/check_tool_acceptance_due.go` in the deployed image; first unattended acceptance-run note.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance iteration loop — iterate until criteria pass
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Both halves proven separately (fail path via controlled test 2026-07-12; fix agents write notes); "let a REAL failure flow through to tool-improver and back" still open at unit close.
- **what:** deploy → acceptance run → failing criterion → `improve_tool` item (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer loads PLAN+NOTES first → fix → append note → redeploy → re-run. Criteria hold the bar still across iterations; NOTES stop iterations fighting each other. *Working* = criteria pass. The one link proven only with a synthetic input is a real failure flowing through tool-improver and back.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_tool_acceptance_runner(2).md#flow; OVERVIEW_self_verifying_tools.md#autonomous-loop
- **relations:** findings acceptance_test pattern (improvement-loop); tool-improver; continuous sweep.
- **verify-later:** an improve_tool item with source 'acceptance' processed end-to-end by tool-improver.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria contract v0 (check-type vocabulary + profiles)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** P0 implements 3 of 7 check types; "the composer emitted "action":"select" … a verb the Tier-4 criteria vocabulary must now define" (open).
- **what:** The machine-readable criteria schema: `profiles: [desktop, mobile]`; check types `selector_exists`, `selector_count`, `no_console_errors`, `asset_loads`, `interaction` (fill/click/select steps + expect), `no_horizontal_overflow`, `page_status_ok`. Deterministic only in v0 — no LLM drives the browser. Desktop = Chromium 1366×900; mobile = one stable Playwright device descriptor (emulation first; real devices out of scope). Phasing P0 boot checks → P1 interpreter+mobile → P2 interactions → P3 screenshots (via the existing Backblaze deploy path) → P4 optional LLM-exploratory mode.
- **sources:** PLAN_tool_acceptance_runner(2).md#criteria-contract,#profiles,#phasing; RUNBOOK_travelling_docs(38).md#stage-6
- **relations:** browser-runner adapter (P0); multi-page tool criteria (open question — url_role field).
- **verify-later:** criteria interpreter coverage in run_checks_action.go; whether "select" verb was added.

<!-- SOURCE: U08_travelling_docs.md -->
### Multi-page tool documentation prerequisites
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK §5.4 "Multi-page prerequisite: preserve-sections re-render + interactivity-aware save guard (pending) before scaling page counts."
- **what:** Multi-page tools add a "Page set & inter-page contract" PLAN section (URLs, shared state keys, data feeds) and may need per-page checks (a `url_role` field). Scaling page counts is explicitly gated on the pending preserve-sections re-render and interactivity-aware save guard.
- **sources:** RUNBOOK_travelling_docs(38).md#§2,#§5; PLAN_travelling_docs(6).md#tool-assurance; PLAN_tool_acceptance_runner(2).md#open-questions
- **relations:** interactive-section clobber (Part 4) below; criteria contract.
- **verify-later:** save_page_sections interactivity guard deployment status.

<!-- SOURCE: U08_travelling_docs.md -->
### Recreation writes page sections — component-less tools and their visibility gap
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Established by query 2026-07-09 ("pages.sections is EMPTY … the 32 KB game body exists only as deployed HTML in the sites repo"); Tier-2 scope note 2026-07-10.
- **what:** `tool-recreation-handler` ends save_page_sections → update_status → deploy_page and never creates a `content_components` row — adopted/recreated tools exist only as page sections + deployed HTML (source in adoption-crawl research_results: adoption_crawl full markdown+rawHTML, adoption_page per-page; `spec.mode="recreate"` is the handshake set by apply_adoption_plan). Consequences: no component address for tool-improver; invisible to Tier 2 by construction (Tier 4 reaches them via pages); NOTES subject must be pipeline-scoped. `site_plan_sections` is site-plan STRUCTURE, not HTML.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev41,#rev42,#rev43; HANDOFF_2026-07-09…_1_.md#§4; RUNBOOK_travelling_docs(38).md#task-5-record
- **relations:** dangling-doc rule; adoption pipeline (007); Tier-2 scope limit.
- **verify-later:** tool-recreation-handler workflow steps; research_results result_types.

<!-- SOURCE: U09_adoption.md -->
### Tools/games behavioural QA loop (planned)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "PLAN_tools_games_behavioral_qa_loop.md (this session) — a standalone QA/maintenance loop that builds out the planned-but-unbuilt Tier 3 (headless behavioral testing) and adds a games lifecycle… Phased; first cut Phase 0+1" (HANDOFF_2026-06-06).
- **what:** A standalone QA loop for deployed interactive tools/games, motivated by real defects (Jelly Invaders degrading over time, P2P host replies not reaching mobile clients, untested cross-browser/mobile variants). Referenced from the adoption thread as FUTURE work; the plan doc itself lives elsewhere.
- **sources:** HANDOFF_2026-06-06#future, HANDOFF_2026-06-09#later-parked
- **relations:** tool-recreation-handler output quality (Family I1); 019/020 tool library/lifecycle Tier 3
- **verify-later:** PLAN_tools_games_behavioral_qa_loop.md (outside this unit)

<!-- SOURCE: U09_adoption.md -->
### Validation observability: structured rejection logging (recordValidationRejection)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Validation observability deployed: store_generated_component_action.go recordValidationRejection writes a structured agent_error_log row on every pre-store rejection… guide-list's attempt-1 failure was captured exactly — one SQL row, no pod-log forensics" (Session 5, 2026-05-11).
- **what:** Every component pre-store validation rejection writes a structured agent_error_log row (severity warning for bookkeeping vs error for structural; orphan/unknown field names as typed JSONB arrays), replacing pod-log forensics. Companion pattern: the retry budget of 3 is calibrated for the single-bookkeeping-orphan failure class seen in Tier-D regens (tool-list missed card_link_label, guide-list read_guide_label); a central label registry would prevent the class entirely (idea, not built).
- **sources:** FOCUS_directory_builder_and_list_components.md#tier-d-converge
- **relations:** component-creator; chrome-template gate (would reuse the same gate/log shape)
- **verify-later:** store_generated_component_action.go recordValidationRejection

<!-- SOURCE: U12_docs024_archives.md -->
### Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option)
- **category:** tool-lifecycle
- **status-signal:** superseded
- **status-evidence:** Archive: "It returns 2–5 suggestions." Live: "It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate."
- **what:** The earliest `tool-suggester` design forced the LLM to always propose at least two tools per site. Replaced by an explicit zero-is-valid design, directly tied to the same failure class as `matchToolToSite` (irrelevant tools forced onto sites).
- **sources:** old/older1/012_tool_lifecycle_guide.md#"Agent: tool-suggester"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Agent: tool-suggester"
- **relations:** tag-based deterministic tool-to-site matching (above)
- **verify-later:** check tool-suggester's current prompt for the zero-suggestions instruction.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool acceptance runner (headless-browser acceptance testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: initial plan (P0 not started)." (tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md, header)
- **what:** Tier-4 rung of the tool verification ladder: a `browser-runner-adapter` (Playwright+Chromium, mirroring analyser-adapter) drives a deployed tool page under desktop+mobile profiles, judges declared criteria (selector_exists, no_console_errors, asset_loads, interaction, no_horizontal_overflow, page_status_ok) pass/fail, feeding failures back as `improve_tool` work items until criteria pass. Criteria live in the tool's travelling PLAN as a criteria block.
- **sources:** tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md#Aim, #Criteria-contract, #Phasing
- **relations:** Behavioral QA loop for tools & games (this is the deterministic v0 layer); tool-lifecycle (020); a recent repo commit ("browser-runner-adapter: commit the full Tier-4 adapter") suggests adapter code may already exist — verify
- **verify-later:** `browser-runner-adapter` deployment; `tool-acceptance-agent` orchestrator; `max_fix_attempts` convention

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool widget clobber mechanism (M1: DELETE+INSERT rebuild wipes side-written components)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "M1 (clobber) is a confirmed latent defect that does not explain these pages but would bite once M2 is fixed. Fixes drafted, not implemented." (PLAN_tool_widget_clobber(9).md)
- **what:** `save_page_sections_action.go` rebuilds a page's `page_components` by `DELETE FROM page_components WHERE page_id=$1` then re-INSERTs only the sections the content writer supplied. Any side-written row not in that list — including a tool/game widget inserted by `create_tool_component`/`deploy_tool` at position 2 — is destroyed on the next `needs_content_page` build. A content-regression guard exists but compares only visible text length after stripping tags, so it is structurally blind to script-heavy widgets. Old content is snapshotted to `page_component_history` before delete, so wipes are recoverable/detectable.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.1-2.2,#3,#7, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-1
- **relations:** Two divergent tool-creation paths; site_plan as authoritative build source; Canonical tool-page section-shape design question; Recreation-loss defect
- **verify-later:** `save_page_sections_action.go` regression guard/delete lines; `page_component_history` rows with `source='save_page_sections_overwrite'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two divergent tool-creation paths (novel vs fork)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** table in §2.3 of PLAN_tool_widget_clobber(9).md documents both paths as existing, currently-running code
- **what:** `create_tool_component_action.go` (the "novel" path) never sets `pages.sections`, leaving it default `[]`; `deploy_tool_action.go` (the "fork" path) sets `pages.sections` to `["hero-tool","tool-guide-intro","<toolFunction>","tool-cta"]`. Both side-write the widget into `page_components` at position 2 and queue `needs_content_page`. The novel path is more exposed to the clobber mechanism since the widget is a member of no section list anywhere.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.3,#9
- **relations:** Tool widget clobber mechanism (M1); Canonicalise tool page identity (T3)
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`; `idx_cc_tool_function_unique` partial index behaviour

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-adoption detection check (T2 — check_tool_recreation_needed.go)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2: "T2 — check_tool_recreation_needed.go ... Deployed. Backfills automatically on next discovery run, if recreation works."
- **what:** A new `discovery_checks` package check: finds `page_type IN ('tool','game')`, `status='active'` pages with no widget, sources `interactive_features` from adoption findings via the same canonical-name transform as T1, and emits `needs_tool_recreation` (7-day per-page cooldown). Pages with no captured features are surfaced but deferred to the tool-suggester/generation path. Doubles as the backfill mechanism for pre-existing widget-less pages.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2
- **relations:** Adoption interactivity misroute (T1); check_tool_health blind spot
- **verify-later:** `check_tool_recreation_needed.go`; `idx_swi_dedup`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Recreation-loss defect (correctly-routed recreation still produces no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "Not confirmed: that widgets now actually deploy... correct routing → completed recreation → no widget... Hold the trigger." (HANDOFF §1,§3)
- **what:** Query K showed all five games on gamesdesign.co.uk — which had routed correctly to `tool-recreation-handler` all along and whose recreation work items completed — had no deployed widget component and no inline `<script>` section. So the routing fix (T1) is necessary but not sufficient; something downstream prevents the widget from landing. Diagnosis was interrupted mid-investigation when a parallel adoption chat reset the underlying state, so the exact mechanism remains open.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§2.9,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§5
- **relations:** Tool widget clobber mechanism (M1); tool-game-* duplicate pages; Post-adoption detection check (T2)
- **verify-later:** re-run queries R1-R3/L/M/N1/N2 against current gamesdesign.co.uk state; check `page_component_history` for a clobber snapshot on a game page

<!-- SOURCE: U13_docs024_small_dirs.md -->
### tool-game-* duplicate pages (T5)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "T5 — tool-game-* duplicate pages (step 8) ... Pending re-observe ... May have been wiped by step-9 reset" (PLAN_tool_widget_clobber(9).md §5b)
- **what:** Five `page_type=tool`, `build_status=planned` pages surfaced that duplicate the five existing games by name (`tool-game-<name>`). Candidate mechanisms: `tool-recreation-handler` building a separate page instead of populating the original interactive page, or a planner/reconciler role-divergence in the `029` canonicalisation family. The duplicates vanished in the step-9 state reset before their origin could be confirmed.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§6
- **relations:** Recreation-loss defect; Canonicalise tool page identity (T3)
- **verify-later:** query M (who created tool-game-* pages) re-pointed at current state

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonicalise tool page identity across surfaces (T3)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "T3 ... Open, independent ... Low risk; can land at any time." (HANDOFF §6)
- **what:** `create_tool_component` and `deploy_tool` build page name/url/page_type ad hoc, diverging from the canonical `datahelpers.CanonicalisePage` helper that adoption and the planner already use. Proposed fix: route both tool actions through the same canonical helper. Flagged as a gap in `029`'s Phase-0 deliverable list, which covered only two other files.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2,§6,§8
- **relations:** Two divergent tool-creation paths; tool-game-* duplicate pages
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`, `datahelpers/page_canonical.go`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonical tool-page section-shape design question and fix options
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Fix options (structural-first; not yet implemented)" (PLAN_tool_widget_clobber(9).md §5)
- **what:** Raises and answers (as a design decision, not yet built) whether a tool page even wants generic hero/guide-intro/CTA sections, or just the widget. Three options recorded: (1) make the widget a first-class section in whichever authority the build reads; (2) right-size the tool page's canonical section list; (3) make `save_page_sections` structure-aware as a safety net. Recommended: 1+2 together with 3 as a guard. Notes `content_guidance` already instructs the writer not to regenerate the widget, but the writer has no mechanical way to honour that.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §4,§5
- **relations:** Tool widget clobber mechanism (M1); site_plan as build authority; content-governance
- **verify-later:** whether `plan_sections_action.go` now emits a tool/embed section for `page_type='tool'` pages

<!-- SOURCE: U13_docs024_small_dirs.md -->
### check_tool_health INNER JOIN blind spot
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "check_tool_health blind spot. Its INNER JOIN content_components → page_components means a tool with no linked page_components row ... is invisible" (PLAN_tool_widget_clobber(9).md §8)
- **what:** The Tier-1 tool health check joins `content_components` to `page_components` with an INNER JOIN, so a `page_type='tool'` page with zero linked components (post-clobber, or never-generated) is invisible to it and the check silently reports "no tools" as a pass. T2 partially closes this by detecting the same condition independently, but the original check itself was not corrected.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §9
- **relations:** Post-adoption detection check (T2); Recreation-loss defect
- **verify-later:** `check_tool_health.go` join logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### forked_from NULL collision risk on novel tools
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "forked_from NULL on novel tools ... Two sites generating the same function would collide. Latent; not today's bug." (PLAN_tool_widget_clobber(9).md §8)
- **what:** `create_tool_component` omits `forked_from`, so novel/generated tools are classified as library tools by the partial unique index `idx_cc_tool_function_unique (function) WHERE component_level='tool' AND forked_from IS NULL AND is_active`. Two different sites independently generating a tool with the same function name would collide.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8
- **relations:** Two divergent tool-creation paths
- **verify-later:** `idx_cc_tool_function_unique` definition; whether any collision has actually occurred

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Behavioral QA loop for tools & games (Tier 3+ headless-browser testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: proposed (2026-06-06)." (PLAN_tools_games_behavioral_qa_loop.md, header)
- **what:** A standalone, slower-cadence QA loop that runs generated tools/games in an isolated multi-engine Playwright pod over time under synthetic drive, to catch defect classes invisible to a single render/screenshot: temporal degradation, cross-browser divergence, mobile-specific layout/touch bugs, and multi-context networked/relay failures. Correctness judged via a three-layer oracle: generic deterministic invariants, type-specific assertions, and LLM-as-judge over a screenshot/video series — with auto-fix gated to high-confidence deterministic findings only. Reuses the existing check→work-item→improver pipeline.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md#1-Why,#4-The-headless-pod,#5-The-oracle-problem,#10-Phasing
- **relations:** Tool acceptance runner (this loop is the heavier behavioral/temporal successor); Games quality lifecycle parity; tool-lifecycle (020)
- **verify-later:** whether any phase has been built; `qa_runs`/`last_qa_at` storage location

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tool-doc header rollout (provenance + stripped headers)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** thin_slice(27) "Tool-doc header rollout (2026-06-11) — apply order is load-bearing. … Three stages; do not reorder — the gate without the prompt fails every generation, and the stamps without the columns fail every insert." No completion claim in this unit's files.
- **what:** Rollout procedure for tool documentation headers: (1) provenance columns on content_components (source_agent_type, source_orchestration_id), (2) anchored idempotent prompt updates adding the `=== tool-doc ===` header requirement (abort if prompts drifted), (3) one binary release (tool_doc_header.go + five action edits) so headers are stamped in the DB template but STRIPPED from shipped pages/CDN assets, with a tool_health no_doc_header WARNING converging old tools on the normal sweep — no retrofit campaign.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#tool-doc-header-rollout
- **relations:** doc_plans/doc_notes (the tools thread's later system); tiered tool acceptance
- **verify-later:** content_components source_% columns; '=== tool-doc ===' in html_template rows; tool_health sweep items

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tiered tool acceptance (static contract check + browser-runner)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Their Stages 5–6 define a TIERED ACCEPTANCE system for tools: a static Tier-2 contract-presence check and a Tier-4 browser-runner adapter (Chromium+Playwright, Kafka request/response per the 035 Adapter Guide) — their 'loop for complicated tools' is acceptance/verification + docs, NOT a rival diagnosis loop."
- **what:** The tools thread's acceptance ladder for generated tools, recorded here as a shared component: Tier-2 static contract-presence verification and a Tier-4 browser-runner adapter executing tools in real Chromium — also earmarked as a future verification service for fix-loop F1 fixes touching pages and as a council reviewer's instrument.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** council of reviewers; tool pipeline; adapters (035 guide)
- **verify-later:** browser-runner adapter existence; tool acceptance stages in the tools thread docs

<!-- SOURCE: U15_docs019_running_notes.md -->
### Tool-doc header system
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: 019/020 CONSOLIDATED — splice files retired" (principles(59)); status marked apply-ready and untouched in v2(36) small-pending list.
- **what:** A standardised 6-12 line sentinel-delimited header block written into every generated tool's script (purpose, behavioural invariants, no-external-calls, version marker) at creation time, stripped at deploy-assembly (three call sites: single-page rerender, `collectJSAssets`, bulk rerender) so it never ships to visitors but is retained in the DB `html_template` for audit/parse parity. Enforced via a hard `HasToolDocHeader` gate in `create_tool_component`, tool-generator/tool-improver prompt edits, and two new `tool_health` tier-1 checks (`no_doc_header` warning, `malformed_doc_header` error). Paired with new `source_agent_type`/`source_orchestration_id` provenance columns on `content_components`, mirroring `knowledge_base`'s existing provenance pair.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tool-doc entries (multiple DONE items).
- **relations:** JS content separation contract; doc claim-verification convention; canonical-doc-home discipline.
- **verify-later:** Whether the rollout (provenance migration → prompts SQL → binary release) was ever applied — repeatedly flagged as "apply-ready, not yet applied" across all later notes files through 2026-07-06.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Fork-divergence detection for library tools
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "IMMEDIATE WIN INSTEAD: FORK-DIVERGENCE detection — pure SQL discovery check (tier-1, zero cost)" (principles(59)).
- **what:** A proposed zero-cost SQL discovery check comparing a deployed fork's `html_template` hash against its `forked_from` library original to answer "which forks are unmodified / safe to bulk-push a library change" — deliberately deferred building full code-symbol indexing of tools (each tool is one IIFE, thin symbol pickings; tool discovery already solved via `semantic_tags`) until a concrete consumer needs it.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tools/provenance/docs design entry.
- **relations:** Tool-doc header system; JS content separation contract.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Tool page missing widget (M1 clobber vs M2 misroute)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 016 addendum(4) "RESOLVED 2026-05-26 → b1 … key the feature map by the canonical name in buildPageFeatureMap"; companion PLAN_tool_widget_clobber.md
- **what:** A `page_type='tool'` page rendering a description but no widget has two causes needing different fixes: M1 clobber (`SavePageSectionsAction` deletes page_components and its content-regression guard can't see a script-heavy widget) vs M2 never-generated (adoption recreate has no parse stage). For gamesdesign, root cause was a misroute: `buildPageFeatureMap` keys by raw page name while the route looks up canonicalised (`tool-`-prefixed) names.
- **sources:** WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#root-cause-m2-corrected-after-verification, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#diagnostic-recipe-read-only-30-seconds, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** CanonicalisePage; adoption parse-stage; site plan reconciler; interactive content
- **verify-later:** buildPageFeatureMap; tool-recreation-handler; SavePageSectionsAction; PLAN_tool_widget_clobber.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool quality tiers: tool-auditor (Tier 2 LLM review), tool-improver, acceptance checks (Tier 2 static) and tool-acceptance-agent (Tier 4 browser runs)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 088 (tool-auditor); 142 enables tool_acceptance check (2026-07-10, doc_notes entry "unit tests... green"); 145 inserts tool-acceptance-agent; 146 makes Tier 4 continuous via tool_acceptance_due sweep.
- **what:** Layered tool verification. Tier 1: check_tool_health structural checks. Tier 2 (LLM): tool-auditor reads full HTML/CSS/JS and reasons through logic/mobile/UX/accessibility, creating improve_tool or needs_human_review items. Tier 2 (static): check_tool_acceptance asserts the PLAN's criteria fence against the deployed page under the ANCHOR RULE ("validate a selector's leftmost id/class token, never the whole path; confirm, never refute; -EDIT ids skipped"). Tier 4: tool-acceptance-agent drives the deployed tool in headless Chromium via the browser-runner adapter against PLAN criteria — "the tier that turns 'deployed' into 'works'" — pass → acceptance-run note; fail → acceptance-fail note + one improve_tool item carrying criteria as acceptance_test. tool-improver executes improve_tool fixes. 7-day cooldowns; cancelled items excluded from cooldown (146).
- **sources:** 088_tool_auditor_agent.sql; 142_enable_tool_acceptance_check.sql; 145_tool_acceptance_agent.sql; 146_enable_tool_acceptance_due.sql; 062_tool_suggester_and_improver.sql
- **relations:** travelling PLAN criteria fences; design-discovery-agent hosts the checks; browser-runner adapter
- **verify-later:** request_browser_run / judge_acceptance_results actions; check_tool_acceptance.go anchor rule; browser-runner adapter deployment

<!-- SOURCE: U18_sql_for_agents.md -->
### Acceptance-criteria honesty: invented selectors and inline-delivery decisions
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 136 (2026-07-09) repairs the first machine-written PLAN's invented ids (#xpTableBody→#tableWrap tbody, #statsStrip→#statRow); 143/144 (2026-07-10) "PLANs surrender to delivered reality" — asset extraction "was designed but never built", so criteria drop asset_loads and the composer prompt is corrected.
- **what:** Two recurring failure classes in machine-written acceptance criteria, and their durable remedies: (1) composers invent selectors they ASSERT on even while obeying never-invent for controls they ACT on — remedy is Tier-2 static validation of criteria selectors against html_template (anchor rule), not sterner prompts; (2) criteria must describe what the system DELIVERS, not aspirations — the /tools/assets/<fn>.js extraction path was never built, all JS ships inline, so PLANs and the composer prompt were superseded to inline delivery ("born honest"). Also note the abandoned mechanism: Path-1 tool asset extraction on rerender.
- **sources:** 136_supersede_xp_curve_plan_selectors.sql; 143_supersede_plans_inline_delivery.sql; 144_composer_inline_delivery.sql; 113_site_asset_renderer.sql (the extraction design it contradicts)
- **relations:** travelling docs supersede pattern; tool acceptance tiers
- **verify-later:** whether asset extraction ever ships (would trigger forward supersede)

<!-- SOURCE: U19_sql_tables_components.md -->
### Component quality tracking (0–100 score)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "None of these fields are required by the existing pipeline — they are additive... selector will use them when present and ignore when NULL" (005 ~9848).
- **what:** Additive quality fields on content_components computed by a compute_component_quality action, with indexes for auditor queries (below threshold OR unscored) and planner preference (higher quality per function). Distinct from avg_quality_score in the selector metadata set.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-quality-tracking
- **relations:** component selector metadata; improvement loop auditors.
- **verify-later:** compute_component_quality action in registry; populated quality_score values.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component versioning (component_versions)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** Table created in schema-mode migration (008 PART 3); page_components.component_version_id exists in live dump with comment "if versioning enabled".
- **what:** Versioned snapshots of component templates (html_template, css_template, input_schema per version_number) so strict-mode pages could pin a specific template version. Referenced as an optional backup target in later template-fix migrations; unclear whether any writer maintains it.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#PART3; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** schema-mode subsystem (abandoned); site_plan_sections.component_version_id (planner provenance).
- **verify-later:** row count in component_versions; writers in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Tool library fork-on-deploy model
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** forked_from column, partial unique index on function scoped to canonical tools, and the later constraint amendment "Forks (forked_from IS NOT NULL) are excluded from the uniqueness check" fixing the add_tool failure on gamedesign.uk.
- **what:** Library tools are canonical rows (component_level='tool', forked_from IS NULL); deploying to a site copies the row as a fork (forked_from = library id) referenced by page_components. Library changes never cascade to forks; fleet updates go through per-site work items. Uniqueness of `function` applies only to active canonical tools so many site forks can share a function; forks are only ever addressed by component_id.
- **sources:** docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#fork-constraint-fix; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#idx_cc_tool_function_unique
- **relations:** component library; seeded tool library; improvement-loop fleet updates.
- **verify-later:** deployer fork-copy code; fork counts per library tool.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component regeneration in place (store_generated_component mechanics)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 083 result: brief-explanation updated in place (same id, created_at unchanged, status 'regenerated', component_versions snapshot, needs_rerender raised) — "matches the documented behaviour (003 §348)".
- **what:** store_generated_component looks up an existing component by the LLM's EMITTED `function` (forked_from IS NULL); if found, it snapshots the current row to component_versions (MAX+1), UPDATEs in place (component_id preserved → all page/site FKs keep resolving), sets template/schema/js_content/render_mode/is_active, then markPagesPendingRebuild raises ONE needs_rerender per affected site. Determinism hazard: regeneration keys on the emitted function name — an unpinned LLM can emit a different name and INSERT a stray duplicate (the 081 'general-hero' incident); pin the function in the description. Pre-store validation rejects `<no value>` templates and checks placeholder/schema parity.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-06-30-~18:35 + #2026-07-01-~12:46; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e
- **relations:** shared library guard; component-quality-auditor; call_agent contract validation (the trigger saga)
- **verify-later:** store_generated_component_action.go lookup + snapshot + markPagesPendingRebuild; component_versions rows

<!-- SOURCE: U23_docs_root_vonc.md -->
### component-quality-auditor auto-regeneration threshold
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Read from its default_config 2026-06-29: creates needs_component_regeneration items only for quality_score < 50, handler component-creator, spec {function, component_id, quality_score, quality_issues}.
- **what:** The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant the three vonc shells scoring EXACTLY 50 were never auto-picked-up (explaining zero queued items and requiring manual triggers). Its item shape confirms the designed regen path keys on function and routes to component-creator. Boundary-condition gap worth a rule review; also the future home of the autonomy plan's maintenance detections.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~21:00; docs/PLAN_dynamic_sections_and_loaders(4).md#maintenance
- **relations:** component regeneration in place; autonomous section composition (auditor rules gap 4)
- **verify-later:** component-quality-auditor default_config condition; quality_score distribution at exactly 50

<!-- SOURCE: U23_docs_root_vonc.md -->
### Store-path template validation (+ pending <script>-balance hardening)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Existing checks confirmed in code 2026-06-29 (`<no value>` rejection, placeholder/schema parity, unclosed `<style>`, section/div presence); the `<script>` balance check + separateInlineJS truncation warning remain "STILL MISSING" / backlog item 2 on 2026-07-09.
- **what:** store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT an unclosed `<script>` — the gap that let provocation-card ship a truncated inline script that swallowed the page footer at render. Hardening definition: add a `<script>` open/close balance check (reject or flag-for-regeneration) plus a truncation warning in separateInlineJS. Prevents the class "truncated template ships and breaks the page".
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-07-03-~13:25 (hardening def); docs/RUNBOOK_phase2_provocation_js(29).md#appendix-g; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** legacy un-extracted shells (the truncation instance); Mode A/B taxonomy
- **verify-later:** store_generated_component_action.go validation block for script balance

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Tool widget clobber (save_page_sections DELETE+INSERT destroys widgets)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.2: `SavePageSectionsAction` does `DELETE FROM page_components WHERE page_id=$1` then re-inserts only writer sections; content-regression guard "compares visible text length … structurally blind to tools"; M1 confirmed latent, "Fixes drafted, not implemented".
- **what:** Two writers collide: create_tool_component/deploy_tool side-writes a tool widget into `page_components`, but the authoritative build rebuilds page_components by DELETE+INSERT from the section list (whose authority is `site_plan`, synced into `pages.sections`). The widget isn't in that list, so the first `needs_content_page` build deletes it (snapshotted to `page_component_history` with `source='save_page_sections_overwrite'`). Fix options: make the widget a first-class site_plan section (preferred), right-size the tool page, or make save_page_sections structure-aware.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2-confirmed-findings, #5-fix-options; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md
- **relations:** load_page_sections_from_spec authority; adoption interactivity misroute; live tools/tool_widget_clobber/ set
- **verify-later:** save_page_sections_action.go; load_page_sections_from_spec_action.go; create_tool_component_action.go vs deploy_tool_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adoption interactivity misroute (canonical prefix desync)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.7 "b1 confirmed"; HANDOFF T1 "Deployed": `buildPageFeatureMap` keyed the feature map by raw `fm["page"]` while the routing loop looked up `CanonicalisePage`-canonicalised names (tool branch adds `tool-` prefix), so tool lookups always missed → empty `Features` → static `needs_content_page` route; games (already `game-` prefixed) matched.
- **what:** Adopted tool pages rendered as static description pages because the feature-map key (bare slug) never matched the canonicalised lookup key (with `tool-` prefix). Fixed (T1) by keying `buildPageFeatureMap` on the canonical name resolved from `plan["pages"]`, so routing and content attachment both land in the same namespace. Self-contained one-function change.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.7-resolved, #5b-tasks; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t1
- **relations:** CanonicalisePage (029); tool-recreation-handler; T3 canonicalise create_tool_component/deploy_tool
- **verify-later:** apply_adoption_plan_action.go buildPageFeatureMap; datahelpers/page_canonical.go CanonicalisePage

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### tool-recreation-handler + recreation discovery check (T2)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF T2 "Deployed": `check_tool_recreation_needed.go` (discovery_checks) detects `page_type IN ('tool','game')` active pages with no tool/game component and no inline `<script>`, sources `interactive_features` from adoption findings by canonical name, emits `needs_tool_recreation:<page>` (7-day cooldown). tool-recreation-handler workflow: `recreate_tool`(execute_llm_prompt)→`check_tool_completeness`→`spawn_rerender`→page-rerender.
- **what:** A registered agent that LLM-recreates interactive widgets for pages adoption captured as text-only, plus a discovery check that backfills widget-less interactive pages automatically on the next scheduled run. Item key deliberately distinct from adoption's `needs_page:<name>` to avoid collision.
- **sources:** tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t2; tools/PLAN_tool_widget_clobber(11).md#5b-tasks
- **relations:** adoption interactivity misroute (T1); recreation-loss defect (T4); check_tool_health blind spot
- **verify-later:** check_tool_recreation_needed.go; tool-recreation-handler agent_definition; check_tool_completeness_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Recreation-loss defect (correct routing yields no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** HANDOFF §3 / PLAN(11) §2.8 (step 8): five games routed `needs_tool_recreation → tool-recreation-handler` and completed (query I), yet all five are widget-less (`has_widget_component=f, has_script_section=f`), plus five new `tool-game-*` duplicate planned pages; step 9 state reset (L/M/N returned 0 rows) left diagnosis incomplete.
- **what:** Even correctly-routed, completed recreation didn't land a deployed widget — a second active defect downstream of routing (T4), blocking. Candidate mechanisms: recreation mis-targeted a parallel `tool-game-*` page, M1 clobber, handler completed without persisting, or a snippets false-negative (inline `<script>` extracted to `/assets/js/snippets.js`). Must be diagnosed before bulk-triggering backfill.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.8-blocking, #2.9-state-reset; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#3
- **relations:** tool widget clobber (M1); tool-recreation-handler; snippets extraction mechanism
- **verify-later:** tool-recreation-handler recreate_tool→check_tool_completeness→spawn_rerender; page_component_history source values

<!-- SOURCE: U25_leopardess_social.md -->
### Mode-B rendered-artifact templates (components stored as rendered output)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** VERDICT §2 (2026-07-09): "rendered_html == html_template with all '<no value>' removed — which the byte counts confirm exactly"; "they are rendered outputs stored as source templates".
- **what:** A component corruption class: html_template full of bare `<no value>`, zero {{.}} slots, empty input_schema — the stored template IS a rendered artifact. Consequences: render is a pure function of the template (predictable to the byte — used twice as an acceptance test); content_data is dead weight; repair_template_slots cannot repair them (zero `</no>` tags → needs_regeneration); for runtime-fill shells the emptiness is accidentally exactly what the loader needs, so regeneration must consciously re-establish the empty-shell contract or sections ship with baked copy.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#2, #8.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-24
- **relations:** runtime-fill guards; component selector/creator (regeneration path); problem-category taxonomy (empty-shell/mode-b-template)
- **verify-later:** components with `<no value>` and 0 schema fields fleet-wide

<!-- SOURCE: U25_leopardess_social.md -->
### Component-creator invocation contract (dual placement + quote-free description)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** NOTES_brief-explanation 083 (2026-07-01) "SUCCEEDED (in-place UPDATE)" via dual placement; framework fix "PATCH_validate_input_contract.go — drafted, not deployed" (HANDOFF §9.3).
- **what:** Manually invoking component-creator (spawn+call) must satisfy BOTH the input_contract (top-level required fields — call_agent validates against top-level extracted fields) AND the workflow's field paths (input_data.spec.*): the working pattern places section_type both top-level and inside spec, pins the function name in the description so the store lands as an in-place UPDATE (else a stray component INSERTs), and keeps the description quote-free to survive the kcat/JSON pipeline. The generic build-dispatch-loop cannot satisfy top-level-required contracts (same class); the durable fix — contract validator accepting top-level OR input_data.spec.{field} — is drafted, not deployed. Regeneration semantics: UPDATE-in-place keyed by function, component_versions snapshot, auto needs_rerender per affected site, store validation rejects `<no value>` templates.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (081/082/083 arc); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Component-creator-input
- **relations:** component selector/creator; shared component library semantics
- **verify-later:** call_agent contract validation code; PATCH_validate_input_contract.go status

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Interactive-page de-tool hazard (content rebuild silently drops a tool/game)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** CONFIRMED 2026-06-22 (gamesdesign game-pathfinding, 18KB A* canvas overwritten 06-14); "fix pending" in 002/005/016; 016b v2 update: two-layer save_page_sections fix WRITTEN, un-deployed
- **what:** A tool lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page or link_resolution_rebuild) regenerates from plan_sections and replaces it with generic-text-block; the prose-based content-regression guard doesn't catch markup/JS loss. Fix layers: interactivity-aware save guard + carry-forward of interactive sections in save_page_sections (written), source_item_id stamping into page_component_history, and routing link maintenance through a preserve-sections path (page_rerender ruled out for CTA re-resolution — it doesn't re-run link logic).
- **sources:** 005(1) hazard block; 002(4)#Interactive-page hazard; 016 §9 final entry; 016b Part 4
- **relations:** phantom-CTA resolution bug (separate); tool recreation mis-key (Part 3)
- **verify-later:** save_page_sections_action.go patched version deployed?; page_component_history.source_item_id population

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool doc header (sentinel comment; stripped at deploy)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 019 full lifecycle table incl. StripToolDocHeader call sites and tool_health checks
- **what:** Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never ids/dates; no */). It never ships: StripToolDocHeader runs at the three outbound assembly points; DB rendered_html retains it for audit parity. Creation gate validates presence; improver preserves/updates it; auditor audits code AGAINST its stated behaviour; malformed (opener without closer) is left in and flagged by tool_health.
- **sources:** 019#Tool Doc Header; 020 tool_health tier-1 checks
- **relations:** per-tool travelling docs (037); tool-auditor
- **verify-later:** platform/content/tool_doc_header.go; prompt migration applied

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool quality three tiers (structural / LLM audit / headless-browser future)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 020: tiers 1–2 automated live, tier 3 "planned"
- **what:** tool_health tier 1 (Go, free): deploy status, HTML/template present, script/style/@media, hex/external-dep warnings, doc-header checks — blockers create improve_tool. Tier 2: audit_tool queued (30-day/tool cooldown) → tool-auditor Sonnet code review across six categories, findings by confidence (certain/likely → improve_tool, possible → needs_human_review), quality_score 1-10 tracked. Tool removal is a human decision via dashboard.
- **sources:** 020#tool_health, #tool-auditor; 019#Tool Quality Standards
- **relations:** tool-improver; component-quality-auditor (sections)
- **verify-later:** check_tool_health.go cooldowns

<!-- SOURCE: U04_idea_uk.md -->
### Content-rebuild de-tools tool pages (confirmed hazard)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "confirmed hazard, fix pending" (TODO P3, 2026-06-26).
- **what:** A needs_page / link_resolution_rebuild on a tool or game page regenerates the page from plan_sections, and the plan knows nothing about the interactive tool living in a section's rendered_html — so the tool is silently replaced with generated prose. Fix direction: route link maintenance through a preserve-sections re-render path, stamp source_item_id, add an interactivity-aware save guard. Flagged as a direct risk to idea.uk's post-P0 rebuild if tools land first.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md#P3; idea.uk/running_notes_2(6).md (backlog)
- **relations:** tool pipeline (005/016b/020/026 cross-refs); page-rerender vs page-build-handler distinction.
- **verify-later:** whether the preserve-sections path landed.

<!-- SOURCE: U05_content_quality_linking.md -->
### tool-recreation-handler (interactive rebuild path)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-26: "WI 6a28c4b3 completed … the playable A* tool is BACK".
- **what:** Recreates interactive tool/game pages from the adoption crawl: recreate_tool (Opus 64k, timeout 2400s) → check_tool_completeness → validate_tool → save_page_sections → update_page_status → spawn_rerender. Mode must be `recreate_tool` (not `recreate` — load_existing_content skips unless mode matches, a prior gotcha). Interactive pages are routed here from adoption via buildPageFeatureMap (T1 routing fix). One of the three save_page_sections callers, so it carries the Part-4 guards; re-creating a tool doesn't trip Layer 1 (new content IS interactive).
- **sources:** PLAN_pathfinding_missing_game.md#2; NOTES(44) 2026-06-25/26; HANDOFF_page_pipeline(11).md#4
- **relations:** interactive clobber; adoption pipeline T1 routing; item_key mis-key.
- **verify-later:** tool-recreation-handler default_config; apply_adoption_plan buildPageFeatureMap.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECISION: acceptance criteria live in the tool's doc_plans PLAN" (2026-07-04 rev 4); consumed live by Tier 2 and Tier 4 checkers.
- **what:** The per-tool definition of *working*. Candidates judged on key/lifecycle/owner: `site_specs` — right machinery, wrong key (site-scoped; per-site copies drift); `site_plans`/directives — wrong lifecycle and owner (churniest artifact, planner-owned; "never store the bar in the artifact that regenerates most"); findings' `acceptance_test` — right pattern, wrong duration (dies with the work item; the standing criteria SEED it). The PLAN wins on all three axes. Format: a machine-extractable fenced ```criteria JSON block (tool-doc-header precedent), extracted by `load_doc_context` as `criteria_json`; lifts to a column only on volume. Per-site parametrisation goes to `direction.must_have`, not the PLAN.
- **sources:** PLAN_travelling_docs(6).md#where-acceptance-criteria-live; 001_README_acceptance_criteria.md; RUNNING_NOTES_travelling_docs(39).md#rev4
- **relations:** verification ladder; findings acceptance_test/max_fix_attempts (improvement-loop 004); direction.must_have.
- **verify-later:** criteria fence extraction in load_doc_context; `has_fence` on live PLANs.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria describe DELIVERED reality, not aspiration (Option B)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECIDED 2026-07-10: Option B — inline reality (user: 'I choose option B and surrender')"; migrations 143 (PLANs superseded) + 144 (composer fixed) applied and verified.
- **what:** The composer had asserted a designed-but-never-built JS extraction (`asset_loads /tools/assets/<fn>.js`) in every PLAN; Tier-2's first sweep failed every tool on it by construction. Principle on record: criteria must describe what the system delivers; aspirations live in roadmaps. If extraction ever ships, PLANs supersede forward again. Corollary: the composer's standard checks became boots/console/status/mobile-fit + optional interaction from real selectors.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5 (Option B block); HANDOFF_2026-07-10…md#§2; PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** js-not-extracted delivery gap; Tier-2 first sweep; PLAN supersede versioning.
- **verify-later:** current PLAN fences have no asset check; compose_plan prompt (four checks, inline delivery line).

<!-- SOURCE: U08_travelling_docs.md -->
### The tool verification ladder (Tiers 0–4)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "The verification ladder is whole (Tier 0/1/2/4) and closes on both outcomes" — RUNBOOK position 2026-07-12; Tier 3 remains Phase B.
- **what:** Cheap-to-expensive tiers, each catching a different class: Tier 0 generation-time output integrity (`HasToolDocHeader` gate + `check_tool_completeness`, deliberately flags-but-passes); Tier 1 structural post-deploy (`check_tool_health`); Tier 2 static contract-presence against deployed HTML (anchor rule); Tier 3 acceptance audit (`tool-auditor` vs criteria — Phase B, unbuilt extension); Tier 4 behavioural — drive the deployed tool in headless Chromium until criteria pass. Standing rule: never read a Tier-2 pass as "the tool works" — that claim belongs to Tier 4.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_travelling_docs(6).md#tool-assurance; OVERVIEW_self_verifying_tools.md#mechanism-2
- **relations:** every tier concept below; "passed checks ≠ working".
- **verify-later:** check_tool_completeness + check_tool_health + discovery_checks/check_tool_acceptance.go + browser-runner adapter, all in the chassis repo.

<!-- SOURCE: U08_travelling_docs.md -->
### "Completeness + validation passed" ≠ working — twice demonstrated
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN(6) rollout outcomes: "The June recreation introduced the economy-simulator's two bugs and passed; run 2 of the repair faithfully recreated them and passed."
- **what:** The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity, not behaviour. The same game shipped broken twice while passing every existing check — the June 2026-06-05 recreation introduced the bugs (proven from tool_recreation_training rows and the origin game.js which has neither bug), and repair run 2 recreated them while its own note truthfully said "completeness + validation passed".
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45; OVERVIEW_self_verifying_tools.md#problem
- **relations:** Tier 4; seam rule; economy-simulator case.
- **verify-later:** tool_recreation_training rows for page d9a8e6e8 dated 2026-06-05.

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-2 static acceptance checker (discovery check `tool_acceptance`)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Stage 5 — LIVE 2026-07-10 ✓ (first sweep proven)" — run cd0d9731 on v1.0.1107 produced exactly the pre-verified findings (2 improve_tool items + 2 acceptance-fail notes, check-level precision confirmed).
- **what:** A browserless discovery check (sibling of `tool_health` in `discovery_checks/`): loads the current PLAN's criteria fence, fetches the deployed page (bounded 12s/2MB, cached per run), and evaluates the statically-visible subset under the anchor rule, plus shell checks (tool-doc header not leaked, no `<no value>` residue). No criteria → a `needs_criteria` note (30-day cooldown), never a fake pass. Failures → one improve_tool item (criteria embedded as `acceptance_test`, 7-day cooldown, cancelled items excluded since migration 146's correct-while-touching) + an acceptance-fail note. Scope limit by construction: only generator-created tools have content_components rows; adopted/recreated page-section tools are invisible to Tier 2.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5; RUNNING_NOTES_travelling_docs(39).md#stage-5-built,#stage-5-live; HANDOFF_2026-07-10…md#§1,§2
- **relations:** anchor rule; migration 142; Tier 4 (reaches page-section tools via pages).
- **verify-later:** `discovery_checks/check_tool_acceptance.go`; design-discovery-agent run_checks list.

<!-- SOURCE: U08_travelling_docs.md -->
### The anchor rule — static checks confirm, never refute
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "STAGE-5 RULE SETTLED" 2026-07-09 after the #tableWrap inspection (empty div filled by JS); implemented + unit-tested incl. the founding cases.
- **what:** Validate only a criteria selector's ANCHOR (leftmost id/class token) against `html_template`: `#tableWrap` exists ⇒ `#tableWrap tr` passes (rows are JS-built; Tier 4 asserts them for real); `#xpTableBody` exists nowhere ⇒ fails ⇒ drop or -EDIT. Static validation can confirm a selector but never refute one — never delete a check merely because the DOM is constructed at runtime. Motivated by the composer inventing selectors it ASSERTS on while copying real ones it ACTS on; the remedy is a check made by the system on itself, not a sterner prompt. Implementation detail banked: CSS class tokens are whitespace-delimited (Go regexp `\b` wrongly splits on hyphens).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5-rule; RUNNING_NOTES_travelling_docs(39).md#rev39,#rev40; OVERVIEW_self_verifying_tools.md#tier-2
- **relations:** composer selector-invention incident; Tier 4 runtime assertions; tool-auditor (same logic belongs there — unbuilt).
- **verify-later:** anchor extraction + class-token comparison in check_tool_acceptance.go tests.

<!-- SOURCE: U08_travelling_docs.md -->
### Composer selector invention — caught twice, machine-corrected
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** CONFIRMED 2026-07-09 (`#xpTableBody`/`#statsStrip` token_anywhere=f); second sighting caught by Tier 2 itself 2026-07-10 (kebab `#drop-chance` vs real camelCase `#dropChance`).
- **what:** The PLAN-composer LLM invented DOM ids for assertion targets despite an explicit "never invent a selector" instruction — the rule held for controls it acts on and failed for things it asserts on. First instance corrected by a guarded supersede migration that itself initially refused a valid runtime selector (leading to the anchor rule); second caught automatically by the live Tier-2 sweep and corrected by migration 143. Demonstrates the design stance: hallucination is countered by verification, not prompt escalation.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#stage-5-live; 0NN_supersede_xp_curve_plan_selectors(2).sql; RUNBOOK_travelling_docs(38).md#task-3-proven
- **relations:** anchor rule; supersede versioning (correction recorded as a NOTES entry — "the travelling-docs loop applied to itself").
- **verify-later:** xp-curve PLAN v1→v2 chain + its correction note in doc_notes.

<!-- SOURCE: U08_travelling_docs.md -->
### tool-acceptance-agent — Tier 4 self-driving orchestrator
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** First machine acceptance-run note (run bf330ac6, 2026-07-12, "Tier-4 acceptance PASSED — all 3 evaluated checks"); fail path proven live via a controlled reverted test (failed=1, improve_tool_created=true, full teardown verified).
- **what:** An agent (migration 145) closing the loop with zero humans: `ensure_site_record → load_docs → request_browser_run (Kafka await; resolves the tool's deployed URL from pages itself; NO-OP skips without awaiting when the PLAN has no criteria) → judge_acceptance_results → complete`. Judge recomputes the verdict from results: all pass → acceptance-run note; any fail → acceptance-fail note + ONE improve_tool item (criteria embedded as acceptance_test, handler tool-improver); component-less recreated/adopted tools get the note but no item — logged honestly for manual routing. Trigger 087 (dry-run default).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tool-acceptance-agent-built,#tier-4-self-driving,#fail-path-proven; README_summary_paragraph2_for_discussion.md; 087_TRIGGER_tool_acceptance.sh (header)
- **relations:** browser-runner adapter; acceptance iteration loop; continuous sweep.
- **verify-later:** `platform/orchestration/actions/tool_acceptance_actions.go`; agent_definitions row tool-acceptance-agent; migration 145.

<!-- SOURCE: U08_travelling_docs.md -->
### Continuous acceptance — the `tool_acceptance_due` periodic sweep
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Built + migration 146 applied 2026-07-12, but "v1.0.1111 … the continuous sweep is NOT in the binary" (untracked-file trap); "GATE: continuous acceptance activates on the next image built from 83ba9bd4+" (T11, 2026-07-13 — state at unit close).
- **what:** A discovery check that emits one `acceptance_run` work item per active tool with a deployed page and current PLAN criteria, unless a verdict landed within 7 days or a run is open. Design calls: post-creation/post-improve hooks deliberately NOT used (they'd fire before the page redeploys — creation ends at 'planned', improve merely queues a rerender; the sweep only ever sees deployed pages); items emitted straight to `triaged` (acceptance needs no human judgment; `detected` items were observed sitting unswept); priority 90 so acceptance tests the NEW page after builds/rerenders.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tier-4-continuous,#v1.0.1111; HANDOFF_2026-07-10…md#T10,T11; README_summary_paragraph2_for_discussion.md
- **relations:** tool-acceptance-agent; untracked-file deploy trap; improve_tool cooldown (cancelled items excluded).
- **verify-later:** `discovery_checks/check_tool_acceptance_due.go` in the deployed image; first unattended acceptance-run note.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance iteration loop — iterate until criteria pass
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Both halves proven separately (fail path via controlled test 2026-07-12; fix agents write notes); "let a REAL failure flow through to tool-improver and back" still open at unit close.
- **what:** deploy → acceptance run → failing criterion → `improve_tool` item (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer loads PLAN+NOTES first → fix → append note → redeploy → re-run. Criteria hold the bar still across iterations; NOTES stop iterations fighting each other. *Working* = criteria pass. The one link proven only with a synthetic input is a real failure flowing through tool-improver and back.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_tool_acceptance_runner(2).md#flow; OVERVIEW_self_verifying_tools.md#autonomous-loop
- **relations:** findings acceptance_test pattern (improvement-loop); tool-improver; continuous sweep.
- **verify-later:** an improve_tool item with source 'acceptance' processed end-to-end by tool-improver.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria contract v0 (check-type vocabulary + profiles)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** P0 implements 3 of 7 check types; "the composer emitted "action":"select" … a verb the Tier-4 criteria vocabulary must now define" (open).
- **what:** The machine-readable criteria schema: `profiles: [desktop, mobile]`; check types `selector_exists`, `selector_count`, `no_console_errors`, `asset_loads`, `interaction` (fill/click/select steps + expect), `no_horizontal_overflow`, `page_status_ok`. Deterministic only in v0 — no LLM drives the browser. Desktop = Chromium 1366×900; mobile = one stable Playwright device descriptor (emulation first; real devices out of scope). Phasing P0 boot checks → P1 interpreter+mobile → P2 interactions → P3 screenshots (via the existing Backblaze deploy path) → P4 optional LLM-exploratory mode.
- **sources:** PLAN_tool_acceptance_runner(2).md#criteria-contract,#profiles,#phasing; RUNBOOK_travelling_docs(38).md#stage-6
- **relations:** browser-runner adapter (P0); multi-page tool criteria (open question — url_role field).
- **verify-later:** criteria interpreter coverage in run_checks_action.go; whether "select" verb was added.

<!-- SOURCE: U08_travelling_docs.md -->
### Multi-page tool documentation prerequisites
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK §5.4 "Multi-page prerequisite: preserve-sections re-render + interactivity-aware save guard (pending) before scaling page counts."
- **what:** Multi-page tools add a "Page set & inter-page contract" PLAN section (URLs, shared state keys, data feeds) and may need per-page checks (a `url_role` field). Scaling page counts is explicitly gated on the pending preserve-sections re-render and interactivity-aware save guard.
- **sources:** RUNBOOK_travelling_docs(38).md#§2,#§5; PLAN_travelling_docs(6).md#tool-assurance; PLAN_tool_acceptance_runner(2).md#open-questions
- **relations:** interactive-section clobber (Part 4) below; criteria contract.
- **verify-later:** save_page_sections interactivity guard deployment status.

<!-- SOURCE: U08_travelling_docs.md -->
### Recreation writes page sections — component-less tools and their visibility gap
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Established by query 2026-07-09 ("pages.sections is EMPTY … the 32 KB game body exists only as deployed HTML in the sites repo"); Tier-2 scope note 2026-07-10.
- **what:** `tool-recreation-handler` ends save_page_sections → update_status → deploy_page and never creates a `content_components` row — adopted/recreated tools exist only as page sections + deployed HTML (source in adoption-crawl research_results: adoption_crawl full markdown+rawHTML, adoption_page per-page; `spec.mode="recreate"` is the handshake set by apply_adoption_plan). Consequences: no component address for tool-improver; invisible to Tier 2 by construction (Tier 4 reaches them via pages); NOTES subject must be pipeline-scoped. `site_plan_sections` is site-plan STRUCTURE, not HTML.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev41,#rev42,#rev43; HANDOFF_2026-07-09…_1_.md#§4; RUNBOOK_travelling_docs(38).md#task-5-record
- **relations:** dangling-doc rule; adoption pipeline (007); Tier-2 scope limit.
- **verify-later:** tool-recreation-handler workflow steps; research_results result_types.

<!-- SOURCE: U09_adoption.md -->
### Tools/games behavioural QA loop (planned)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "PLAN_tools_games_behavioral_qa_loop.md (this session) — a standalone QA/maintenance loop that builds out the planned-but-unbuilt Tier 3 (headless behavioral testing) and adds a games lifecycle… Phased; first cut Phase 0+1" (HANDOFF_2026-06-06).
- **what:** A standalone QA loop for deployed interactive tools/games, motivated by real defects (Jelly Invaders degrading over time, P2P host replies not reaching mobile clients, untested cross-browser/mobile variants). Referenced from the adoption thread as FUTURE work; the plan doc itself lives elsewhere.
- **sources:** HANDOFF_2026-06-06#future, HANDOFF_2026-06-09#later-parked
- **relations:** tool-recreation-handler output quality (Family I1); 019/020 tool library/lifecycle Tier 3
- **verify-later:** PLAN_tools_games_behavioral_qa_loop.md (outside this unit)

<!-- SOURCE: U09_adoption.md -->
### Validation observability: structured rejection logging (recordValidationRejection)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Validation observability deployed: store_generated_component_action.go recordValidationRejection writes a structured agent_error_log row on every pre-store rejection… guide-list's attempt-1 failure was captured exactly — one SQL row, no pod-log forensics" (Session 5, 2026-05-11).
- **what:** Every component pre-store validation rejection writes a structured agent_error_log row (severity warning for bookkeeping vs error for structural; orphan/unknown field names as typed JSONB arrays), replacing pod-log forensics. Companion pattern: the retry budget of 3 is calibrated for the single-bookkeeping-orphan failure class seen in Tier-D regens (tool-list missed card_link_label, guide-list read_guide_label); a central label registry would prevent the class entirely (idea, not built).
- **sources:** FOCUS_directory_builder_and_list_components.md#tier-d-converge
- **relations:** component-creator; chrome-template gate (would reuse the same gate/log shape)
- **verify-later:** store_generated_component_action.go recordValidationRejection

<!-- SOURCE: U12_docs024_archives.md -->
### Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option)
- **category:** tool-lifecycle
- **status-signal:** superseded
- **status-evidence:** Archive: "It returns 2–5 suggestions." Live: "It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate."
- **what:** The earliest `tool-suggester` design forced the LLM to always propose at least two tools per site. Replaced by an explicit zero-is-valid design, directly tied to the same failure class as `matchToolToSite` (irrelevant tools forced onto sites).
- **sources:** old/older1/012_tool_lifecycle_guide.md#"Agent: tool-suggester"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Agent: tool-suggester"
- **relations:** tag-based deterministic tool-to-site matching (above)
- **verify-later:** check tool-suggester's current prompt for the zero-suggestions instruction.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool acceptance runner (headless-browser acceptance testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: initial plan (P0 not started)." (tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md, header)
- **what:** Tier-4 rung of the tool verification ladder: a `browser-runner-adapter` (Playwright+Chromium, mirroring analyser-adapter) drives a deployed tool page under desktop+mobile profiles, judges declared criteria (selector_exists, no_console_errors, asset_loads, interaction, no_horizontal_overflow, page_status_ok) pass/fail, feeding failures back as `improve_tool` work items until criteria pass. Criteria live in the tool's travelling PLAN as a criteria block.
- **sources:** tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md#Aim, #Criteria-contract, #Phasing
- **relations:** Behavioral QA loop for tools & games (this is the deterministic v0 layer); tool-lifecycle (020); a recent repo commit ("browser-runner-adapter: commit the full Tier-4 adapter") suggests adapter code may already exist — verify
- **verify-later:** `browser-runner-adapter` deployment; `tool-acceptance-agent` orchestrator; `max_fix_attempts` convention

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool widget clobber mechanism (M1: DELETE+INSERT rebuild wipes side-written components)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "M1 (clobber) is a confirmed latent defect that does not explain these pages but would bite once M2 is fixed. Fixes drafted, not implemented." (PLAN_tool_widget_clobber(9).md)
- **what:** `save_page_sections_action.go` rebuilds a page's `page_components` by `DELETE FROM page_components WHERE page_id=$1` then re-INSERTs only the sections the content writer supplied. Any side-written row not in that list — including a tool/game widget inserted by `create_tool_component`/`deploy_tool` at position 2 — is destroyed on the next `needs_content_page` build. A content-regression guard exists but compares only visible text length after stripping tags, so it is structurally blind to script-heavy widgets. Old content is snapshotted to `page_component_history` before delete, so wipes are recoverable/detectable.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.1-2.2,#3,#7, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-1
- **relations:** Two divergent tool-creation paths; site_plan as authoritative build source; Canonical tool-page section-shape design question; Recreation-loss defect
- **verify-later:** `save_page_sections_action.go` regression guard/delete lines; `page_component_history` rows with `source='save_page_sections_overwrite'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two divergent tool-creation paths (novel vs fork)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** table in §2.3 of PLAN_tool_widget_clobber(9).md documents both paths as existing, currently-running code
- **what:** `create_tool_component_action.go` (the "novel" path) never sets `pages.sections`, leaving it default `[]`; `deploy_tool_action.go` (the "fork" path) sets `pages.sections` to `["hero-tool","tool-guide-intro","<toolFunction>","tool-cta"]`. Both side-write the widget into `page_components` at position 2 and queue `needs_content_page`. The novel path is more exposed to the clobber mechanism since the widget is a member of no section list anywhere.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.3,#9
- **relations:** Tool widget clobber mechanism (M1); Canonicalise tool page identity (T3)
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`; `idx_cc_tool_function_unique` partial index behaviour

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-adoption detection check (T2 — check_tool_recreation_needed.go)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2: "T2 — check_tool_recreation_needed.go ... Deployed. Backfills automatically on next discovery run, if recreation works."
- **what:** A new `discovery_checks` package check: finds `page_type IN ('tool','game')`, `status='active'` pages with no widget, sources `interactive_features` from adoption findings via the same canonical-name transform as T1, and emits `needs_tool_recreation` (7-day per-page cooldown). Pages with no captured features are surfaced but deferred to the tool-suggester/generation path. Doubles as the backfill mechanism for pre-existing widget-less pages.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2
- **relations:** Adoption interactivity misroute (T1); check_tool_health blind spot
- **verify-later:** `check_tool_recreation_needed.go`; `idx_swi_dedup`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Recreation-loss defect (correctly-routed recreation still produces no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "Not confirmed: that widgets now actually deploy... correct routing → completed recreation → no widget... Hold the trigger." (HANDOFF §1,§3)
- **what:** Query K showed all five games on gamesdesign.co.uk — which had routed correctly to `tool-recreation-handler` all along and whose recreation work items completed — had no deployed widget component and no inline `<script>` section. So the routing fix (T1) is necessary but not sufficient; something downstream prevents the widget from landing. Diagnosis was interrupted mid-investigation when a parallel adoption chat reset the underlying state, so the exact mechanism remains open.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§2.9,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§5
- **relations:** Tool widget clobber mechanism (M1); tool-game-* duplicate pages; Post-adoption detection check (T2)
- **verify-later:** re-run queries R1-R3/L/M/N1/N2 against current gamesdesign.co.uk state; check `page_component_history` for a clobber snapshot on a game page

<!-- SOURCE: U13_docs024_small_dirs.md -->
### tool-game-* duplicate pages (T5)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "T5 — tool-game-* duplicate pages (step 8) ... Pending re-observe ... May have been wiped by step-9 reset" (PLAN_tool_widget_clobber(9).md §5b)
- **what:** Five `page_type=tool`, `build_status=planned` pages surfaced that duplicate the five existing games by name (`tool-game-<name>`). Candidate mechanisms: `tool-recreation-handler` building a separate page instead of populating the original interactive page, or a planner/reconciler role-divergence in the `029` canonicalisation family. The duplicates vanished in the step-9 state reset before their origin could be confirmed.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§6
- **relations:** Recreation-loss defect; Canonicalise tool page identity (T3)
- **verify-later:** query M (who created tool-game-* pages) re-pointed at current state

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonicalise tool page identity across surfaces (T3)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "T3 ... Open, independent ... Low risk; can land at any time." (HANDOFF §6)
- **what:** `create_tool_component` and `deploy_tool` build page name/url/page_type ad hoc, diverging from the canonical `datahelpers.CanonicalisePage` helper that adoption and the planner already use. Proposed fix: route both tool actions through the same canonical helper. Flagged as a gap in `029`'s Phase-0 deliverable list, which covered only two other files.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2,§6,§8
- **relations:** Two divergent tool-creation paths; tool-game-* duplicate pages
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`, `datahelpers/page_canonical.go`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonical tool-page section-shape design question and fix options
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Fix options (structural-first; not yet implemented)" (PLAN_tool_widget_clobber(9).md §5)
- **what:** Raises and answers (as a design decision, not yet built) whether a tool page even wants generic hero/guide-intro/CTA sections, or just the widget. Three options recorded: (1) make the widget a first-class section in whichever authority the build reads; (2) right-size the tool page's canonical section list; (3) make `save_page_sections` structure-aware as a safety net. Recommended: 1+2 together with 3 as a guard. Notes `content_guidance` already instructs the writer not to regenerate the widget, but the writer has no mechanical way to honour that.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §4,§5
- **relations:** Tool widget clobber mechanism (M1); site_plan as build authority; content-governance
- **verify-later:** whether `plan_sections_action.go` now emits a tool/embed section for `page_type='tool'` pages

<!-- SOURCE: U13_docs024_small_dirs.md -->
### check_tool_health INNER JOIN blind spot
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "check_tool_health blind spot. Its INNER JOIN content_components → page_components means a tool with no linked page_components row ... is invisible" (PLAN_tool_widget_clobber(9).md §8)
- **what:** The Tier-1 tool health check joins `content_components` to `page_components` with an INNER JOIN, so a `page_type='tool'` page with zero linked components (post-clobber, or never-generated) is invisible to it and the check silently reports "no tools" as a pass. T2 partially closes this by detecting the same condition independently, but the original check itself was not corrected.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §9
- **relations:** Post-adoption detection check (T2); Recreation-loss defect
- **verify-later:** `check_tool_health.go` join logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### forked_from NULL collision risk on novel tools
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "forked_from NULL on novel tools ... Two sites generating the same function would collide. Latent; not today's bug." (PLAN_tool_widget_clobber(9).md §8)
- **what:** `create_tool_component` omits `forked_from`, so novel/generated tools are classified as library tools by the partial unique index `idx_cc_tool_function_unique (function) WHERE component_level='tool' AND forked_from IS NULL AND is_active`. Two different sites independently generating a tool with the same function name would collide.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8
- **relations:** Two divergent tool-creation paths
- **verify-later:** `idx_cc_tool_function_unique` definition; whether any collision has actually occurred

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Behavioral QA loop for tools & games (Tier 3+ headless-browser testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: proposed (2026-06-06)." (PLAN_tools_games_behavioral_qa_loop.md, header)
- **what:** A standalone, slower-cadence QA loop that runs generated tools/games in an isolated multi-engine Playwright pod over time under synthetic drive, to catch defect classes invisible to a single render/screenshot: temporal degradation, cross-browser divergence, mobile-specific layout/touch bugs, and multi-context networked/relay failures. Correctness judged via a three-layer oracle: generic deterministic invariants, type-specific assertions, and LLM-as-judge over a screenshot/video series — with auto-fix gated to high-confidence deterministic findings only. Reuses the existing check→work-item→improver pipeline.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md#1-Why,#4-The-headless-pod,#5-The-oracle-problem,#10-Phasing
- **relations:** Tool acceptance runner (this loop is the heavier behavioral/temporal successor); Games quality lifecycle parity; tool-lifecycle (020)
- **verify-later:** whether any phase has been built; `qa_runs`/`last_qa_at` storage location

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tool-doc header rollout (provenance + stripped headers)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** thin_slice(27) "Tool-doc header rollout (2026-06-11) — apply order is load-bearing. … Three stages; do not reorder — the gate without the prompt fails every generation, and the stamps without the columns fail every insert." No completion claim in this unit's files.
- **what:** Rollout procedure for tool documentation headers: (1) provenance columns on content_components (source_agent_type, source_orchestration_id), (2) anchored idempotent prompt updates adding the `=== tool-doc ===` header requirement (abort if prompts drifted), (3) one binary release (tool_doc_header.go + five action edits) so headers are stamped in the DB template but STRIPPED from shipped pages/CDN assets, with a tool_health no_doc_header WARNING converging old tools on the normal sweep — no retrofit campaign.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#tool-doc-header-rollout
- **relations:** doc_plans/doc_notes (the tools thread's later system); tiered tool acceptance
- **verify-later:** content_components source_% columns; '=== tool-doc ===' in html_template rows; tool_health sweep items

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tiered tool acceptance (static contract check + browser-runner)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Their Stages 5–6 define a TIERED ACCEPTANCE system for tools: a static Tier-2 contract-presence check and a Tier-4 browser-runner adapter (Chromium+Playwright, Kafka request/response per the 035 Adapter Guide) — their 'loop for complicated tools' is acceptance/verification + docs, NOT a rival diagnosis loop."
- **what:** The tools thread's acceptance ladder for generated tools, recorded here as a shared component: Tier-2 static contract-presence verification and a Tier-4 browser-runner adapter executing tools in real Chromium — also earmarked as a future verification service for fix-loop F1 fixes touching pages and as a council reviewer's instrument.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** council of reviewers; tool pipeline; adapters (035 guide)
- **verify-later:** browser-runner adapter existence; tool acceptance stages in the tools thread docs

<!-- SOURCE: U15_docs019_running_notes.md -->
### Tool-doc header system
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: 019/020 CONSOLIDATED — splice files retired" (principles(59)); status marked apply-ready and untouched in v2(36) small-pending list.
- **what:** A standardised 6-12 line sentinel-delimited header block written into every generated tool's script (purpose, behavioural invariants, no-external-calls, version marker) at creation time, stripped at deploy-assembly (three call sites: single-page rerender, `collectJSAssets`, bulk rerender) so it never ships to visitors but is retained in the DB `html_template` for audit/parse parity. Enforced via a hard `HasToolDocHeader` gate in `create_tool_component`, tool-generator/tool-improver prompt edits, and two new `tool_health` tier-1 checks (`no_doc_header` warning, `malformed_doc_header` error). Paired with new `source_agent_type`/`source_orchestration_id` provenance columns on `content_components`, mirroring `knowledge_base`'s existing provenance pair.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tool-doc entries (multiple DONE items).
- **relations:** JS content separation contract; doc claim-verification convention; canonical-doc-home discipline.
- **verify-later:** Whether the rollout (provenance migration → prompts SQL → binary release) was ever applied — repeatedly flagged as "apply-ready, not yet applied" across all later notes files through 2026-07-06.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Fork-divergence detection for library tools
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "IMMEDIATE WIN INSTEAD: FORK-DIVERGENCE detection — pure SQL discovery check (tier-1, zero cost)" (principles(59)).
- **what:** A proposed zero-cost SQL discovery check comparing a deployed fork's `html_template` hash against its `forked_from` library original to answer "which forks are unmodified / safe to bulk-push a library change" — deliberately deferred building full code-symbol indexing of tools (each tool is one IIFE, thin symbol pickings; tool discovery already solved via `semantic_tags`) until a concrete consumer needs it.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tools/provenance/docs design entry.
- **relations:** Tool-doc header system; JS content separation contract.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Tool page missing widget (M1 clobber vs M2 misroute)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 016 addendum(4) "RESOLVED 2026-05-26 → b1 … key the feature map by the canonical name in buildPageFeatureMap"; companion PLAN_tool_widget_clobber.md
- **what:** A `page_type='tool'` page rendering a description but no widget has two causes needing different fixes: M1 clobber (`SavePageSectionsAction` deletes page_components and its content-regression guard can't see a script-heavy widget) vs M2 never-generated (adoption recreate has no parse stage). For gamesdesign, root cause was a misroute: `buildPageFeatureMap` keys by raw page name while the route looks up canonicalised (`tool-`-prefixed) names.
- **sources:** WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#root-cause-m2-corrected-after-verification, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#diagnostic-recipe-read-only-30-seconds, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** CanonicalisePage; adoption parse-stage; site plan reconciler; interactive content
- **verify-later:** buildPageFeatureMap; tool-recreation-handler; SavePageSectionsAction; PLAN_tool_widget_clobber.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool quality tiers: tool-auditor (Tier 2 LLM review), tool-improver, acceptance checks (Tier 2 static) and tool-acceptance-agent (Tier 4 browser runs)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 088 (tool-auditor); 142 enables tool_acceptance check (2026-07-10, doc_notes entry "unit tests... green"); 145 inserts tool-acceptance-agent; 146 makes Tier 4 continuous via tool_acceptance_due sweep.
- **what:** Layered tool verification. Tier 1: check_tool_health structural checks. Tier 2 (LLM): tool-auditor reads full HTML/CSS/JS and reasons through logic/mobile/UX/accessibility, creating improve_tool or needs_human_review items. Tier 2 (static): check_tool_acceptance asserts the PLAN's criteria fence against the deployed page under the ANCHOR RULE ("validate a selector's leftmost id/class token, never the whole path; confirm, never refute; -EDIT ids skipped"). Tier 4: tool-acceptance-agent drives the deployed tool in headless Chromium via the browser-runner adapter against PLAN criteria — "the tier that turns 'deployed' into 'works'" — pass → acceptance-run note; fail → acceptance-fail note + one improve_tool item carrying criteria as acceptance_test. tool-improver executes improve_tool fixes. 7-day cooldowns; cancelled items excluded from cooldown (146).
- **sources:** 088_tool_auditor_agent.sql; 142_enable_tool_acceptance_check.sql; 145_tool_acceptance_agent.sql; 146_enable_tool_acceptance_due.sql; 062_tool_suggester_and_improver.sql
- **relations:** travelling PLAN criteria fences; design-discovery-agent hosts the checks; browser-runner adapter
- **verify-later:** request_browser_run / judge_acceptance_results actions; check_tool_acceptance.go anchor rule; browser-runner adapter deployment

<!-- SOURCE: U18_sql_for_agents.md -->
### Acceptance-criteria honesty: invented selectors and inline-delivery decisions
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 136 (2026-07-09) repairs the first machine-written PLAN's invented ids (#xpTableBody→#tableWrap tbody, #statsStrip→#statRow); 143/144 (2026-07-10) "PLANs surrender to delivered reality" — asset extraction "was designed but never built", so criteria drop asset_loads and the composer prompt is corrected.
- **what:** Two recurring failure classes in machine-written acceptance criteria, and their durable remedies: (1) composers invent selectors they ASSERT on even while obeying never-invent for controls they ACT on — remedy is Tier-2 static validation of criteria selectors against html_template (anchor rule), not sterner prompts; (2) criteria must describe what the system DELIVERS, not aspirations — the /tools/assets/<fn>.js extraction path was never built, all JS ships inline, so PLANs and the composer prompt were superseded to inline delivery ("born honest"). Also note the abandoned mechanism: Path-1 tool asset extraction on rerender.
- **sources:** 136_supersede_xp_curve_plan_selectors.sql; 143_supersede_plans_inline_delivery.sql; 144_composer_inline_delivery.sql; 113_site_asset_renderer.sql (the extraction design it contradicts)
- **relations:** travelling docs supersede pattern; tool acceptance tiers
- **verify-later:** whether asset extraction ever ships (would trigger forward supersede)

<!-- SOURCE: U19_sql_tables_components.md -->
### Component quality tracking (0–100 score)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "None of these fields are required by the existing pipeline — they are additive... selector will use them when present and ignore when NULL" (005 ~9848).
- **what:** Additive quality fields on content_components computed by a compute_component_quality action, with indexes for auditor queries (below threshold OR unscored) and planner preference (higher quality per function). Distinct from avg_quality_score in the selector metadata set.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-quality-tracking
- **relations:** component selector metadata; improvement loop auditors.
- **verify-later:** compute_component_quality action in registry; populated quality_score values.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component versioning (component_versions)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** Table created in schema-mode migration (008 PART 3); page_components.component_version_id exists in live dump with comment "if versioning enabled".
- **what:** Versioned snapshots of component templates (html_template, css_template, input_schema per version_number) so strict-mode pages could pin a specific template version. Referenced as an optional backup target in later template-fix migrations; unclear whether any writer maintains it.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#PART3; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** schema-mode subsystem (abandoned); site_plan_sections.component_version_id (planner provenance).
- **verify-later:** row count in component_versions; writers in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Tool library fork-on-deploy model
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** forked_from column, partial unique index on function scoped to canonical tools, and the later constraint amendment "Forks (forked_from IS NOT NULL) are excluded from the uniqueness check" fixing the add_tool failure on gamedesign.uk.
- **what:** Library tools are canonical rows (component_level='tool', forked_from IS NULL); deploying to a site copies the row as a fork (forked_from = library id) referenced by page_components. Library changes never cascade to forks; fleet updates go through per-site work items. Uniqueness of `function` applies only to active canonical tools so many site forks can share a function; forks are only ever addressed by component_id.
- **sources:** docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#fork-constraint-fix; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#idx_cc_tool_function_unique
- **relations:** component library; seeded tool library; improvement-loop fleet updates.
- **verify-later:** deployer fork-copy code; fork counts per library tool.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component regeneration in place (store_generated_component mechanics)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 083 result: brief-explanation updated in place (same id, created_at unchanged, status 'regenerated', component_versions snapshot, needs_rerender raised) — "matches the documented behaviour (003 §348)".
- **what:** store_generated_component looks up an existing component by the LLM's EMITTED `function` (forked_from IS NULL); if found, it snapshots the current row to component_versions (MAX+1), UPDATEs in place (component_id preserved → all page/site FKs keep resolving), sets template/schema/js_content/render_mode/is_active, then markPagesPendingRebuild raises ONE needs_rerender per affected site. Determinism hazard: regeneration keys on the emitted function name — an unpinned LLM can emit a different name and INSERT a stray duplicate (the 081 'general-hero' incident); pin the function in the description. Pre-store validation rejects `<no value>` templates and checks placeholder/schema parity.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-06-30-~18:35 + #2026-07-01-~12:46; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e
- **relations:** shared library guard; component-quality-auditor; call_agent contract validation (the trigger saga)
- **verify-later:** store_generated_component_action.go lookup + snapshot + markPagesPendingRebuild; component_versions rows

<!-- SOURCE: U23_docs_root_vonc.md -->
### component-quality-auditor auto-regeneration threshold
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Read from its default_config 2026-06-29: creates needs_component_regeneration items only for quality_score < 50, handler component-creator, spec {function, component_id, quality_score, quality_issues}.
- **what:** The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant the three vonc shells scoring EXACTLY 50 were never auto-picked-up (explaining zero queued items and requiring manual triggers). Its item shape confirms the designed regen path keys on function and routes to component-creator. Boundary-condition gap worth a rule review; also the future home of the autonomy plan's maintenance detections.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~21:00; docs/PLAN_dynamic_sections_and_loaders(4).md#maintenance
- **relations:** component regeneration in place; autonomous section composition (auditor rules gap 4)
- **verify-later:** component-quality-auditor default_config condition; quality_score distribution at exactly 50

<!-- SOURCE: U23_docs_root_vonc.md -->
### Store-path template validation (+ pending <script>-balance hardening)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Existing checks confirmed in code 2026-06-29 (`<no value>` rejection, placeholder/schema parity, unclosed `<style>`, section/div presence); the `<script>` balance check + separateInlineJS truncation warning remain "STILL MISSING" / backlog item 2 on 2026-07-09.
- **what:** store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT an unclosed `<script>` — the gap that let provocation-card ship a truncated inline script that swallowed the page footer at render. Hardening definition: add a `<script>` open/close balance check (reject or flag-for-regeneration) plus a truncation warning in separateInlineJS. Prevents the class "truncated template ships and breaks the page".
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-07-03-~13:25 (hardening def); docs/RUNBOOK_phase2_provocation_js(29).md#appendix-g; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** legacy un-extracted shells (the truncation instance); Mode A/B taxonomy
- **verify-later:** store_generated_component_action.go validation block for script balance

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Tool widget clobber (save_page_sections DELETE+INSERT destroys widgets)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.2: `SavePageSectionsAction` does `DELETE FROM page_components WHERE page_id=$1` then re-inserts only writer sections; content-regression guard "compares visible text length … structurally blind to tools"; M1 confirmed latent, "Fixes drafted, not implemented".
- **what:** Two writers collide: create_tool_component/deploy_tool side-writes a tool widget into `page_components`, but the authoritative build rebuilds page_components by DELETE+INSERT from the section list (whose authority is `site_plan`, synced into `pages.sections`). The widget isn't in that list, so the first `needs_content_page` build deletes it (snapshotted to `page_component_history` with `source='save_page_sections_overwrite'`). Fix options: make the widget a first-class site_plan section (preferred), right-size the tool page, or make save_page_sections structure-aware.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2-confirmed-findings, #5-fix-options; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md
- **relations:** load_page_sections_from_spec authority; adoption interactivity misroute; live tools/tool_widget_clobber/ set
- **verify-later:** save_page_sections_action.go; load_page_sections_from_spec_action.go; create_tool_component_action.go vs deploy_tool_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adoption interactivity misroute (canonical prefix desync)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.7 "b1 confirmed"; HANDOFF T1 "Deployed": `buildPageFeatureMap` keyed the feature map by raw `fm["page"]` while the routing loop looked up `CanonicalisePage`-canonicalised names (tool branch adds `tool-` prefix), so tool lookups always missed → empty `Features` → static `needs_content_page` route; games (already `game-` prefixed) matched.
- **what:** Adopted tool pages rendered as static description pages because the feature-map key (bare slug) never matched the canonicalised lookup key (with `tool-` prefix). Fixed (T1) by keying `buildPageFeatureMap` on the canonical name resolved from `plan["pages"]`, so routing and content attachment both land in the same namespace. Self-contained one-function change.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.7-resolved, #5b-tasks; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t1
- **relations:** CanonicalisePage (029); tool-recreation-handler; T3 canonicalise create_tool_component/deploy_tool
- **verify-later:** apply_adoption_plan_action.go buildPageFeatureMap; datahelpers/page_canonical.go CanonicalisePage

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### tool-recreation-handler + recreation discovery check (T2)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF T2 "Deployed": `check_tool_recreation_needed.go` (discovery_checks) detects `page_type IN ('tool','game')` active pages with no tool/game component and no inline `<script>`, sources `interactive_features` from adoption findings by canonical name, emits `needs_tool_recreation:<page>` (7-day cooldown). tool-recreation-handler workflow: `recreate_tool`(execute_llm_prompt)→`check_tool_completeness`→`spawn_rerender`→page-rerender.
- **what:** A registered agent that LLM-recreates interactive widgets for pages adoption captured as text-only, plus a discovery check that backfills widget-less interactive pages automatically on the next scheduled run. Item key deliberately distinct from adoption's `needs_page:<name>` to avoid collision.
- **sources:** tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t2; tools/PLAN_tool_widget_clobber(11).md#5b-tasks
- **relations:** adoption interactivity misroute (T1); recreation-loss defect (T4); check_tool_health blind spot
- **verify-later:** check_tool_recreation_needed.go; tool-recreation-handler agent_definition; check_tool_completeness_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Recreation-loss defect (correct routing yields no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** HANDOFF §3 / PLAN(11) §2.8 (step 8): five games routed `needs_tool_recreation → tool-recreation-handler` and completed (query I), yet all five are widget-less (`has_widget_component=f, has_script_section=f`), plus five new `tool-game-*` duplicate planned pages; step 9 state reset (L/M/N returned 0 rows) left diagnosis incomplete.
- **what:** Even correctly-routed, completed recreation didn't land a deployed widget — a second active defect downstream of routing (T4), blocking. Candidate mechanisms: recreation mis-targeted a parallel `tool-game-*` page, M1 clobber, handler completed without persisting, or a snippets false-negative (inline `<script>` extracted to `/assets/js/snippets.js`). Must be diagnosed before bulk-triggering backfill.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.8-blocking, #2.9-state-reset; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#3
- **relations:** tool widget clobber (M1); tool-recreation-handler; snippets extraction mechanism
- **verify-later:** tool-recreation-handler recreate_tool→check_tool_completeness→spawn_rerender; page_component_history source values

<!-- SOURCE: U25_leopardess_social.md -->
### Mode-B rendered-artifact templates (components stored as rendered output)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** VERDICT §2 (2026-07-09): "rendered_html == html_template with all '<no value>' removed — which the byte counts confirm exactly"; "they are rendered outputs stored as source templates".
- **what:** A component corruption class: html_template full of bare `<no value>`, zero {{.}} slots, empty input_schema — the stored template IS a rendered artifact. Consequences: render is a pure function of the template (predictable to the byte — used twice as an acceptance test); content_data is dead weight; repair_template_slots cannot repair them (zero `</no>` tags → needs_regeneration); for runtime-fill shells the emptiness is accidentally exactly what the loader needs, so regeneration must consciously re-establish the empty-shell contract or sections ship with baked copy.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#2, #8.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-24
- **relations:** runtime-fill guards; component selector/creator (regeneration path); problem-category taxonomy (empty-shell/mode-b-template)
- **verify-later:** components with `<no value>` and 0 schema fields fleet-wide

<!-- SOURCE: U25_leopardess_social.md -->
### Component-creator invocation contract (dual placement + quote-free description)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** NOTES_brief-explanation 083 (2026-07-01) "SUCCEEDED (in-place UPDATE)" via dual placement; framework fix "PATCH_validate_input_contract.go — drafted, not deployed" (HANDOFF §9.3).
- **what:** Manually invoking component-creator (spawn+call) must satisfy BOTH the input_contract (top-level required fields — call_agent validates against top-level extracted fields) AND the workflow's field paths (input_data.spec.*): the working pattern places section_type both top-level and inside spec, pins the function name in the description so the store lands as an in-place UPDATE (else a stray component INSERTs), and keeps the description quote-free to survive the kcat/JSON pipeline. The generic build-dispatch-loop cannot satisfy top-level-required contracts (same class); the durable fix — contract validator accepting top-level OR input_data.spec.{field} — is drafted, not deployed. Regeneration semantics: UPDATE-in-place keyed by function, component_versions snapshot, auto needs_rerender per affected site, store validation rejects `<no value>` templates.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (081/082/083 arc); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Component-creator-input
- **relations:** component selector/creator; shared component library semantics
- **verify-later:** call_agent contract validation code; PATCH_validate_input_contract.go status

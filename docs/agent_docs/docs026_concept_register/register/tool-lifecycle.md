# Register — tool-lifecycle

30 concepts, consolidated from 51 raw extractions across units U01, U04, U05, U08, U09, U12, U13, U14, U15, U17a, U18, U19, U23, U24a, U25. One concept (fork-on-deploy tool ownership model) was cross-merged into register/tool-library.md's equivalent entry (TLIB-001) since it is the same mechanism described from a different angle — noted here and there.

### TL-001 — Tool widget clobber hazard (interactive content silently destroyed by content rebuild)
- **status:** partial
- **status-evidence:** CONFIRMED 2026-06-22 (gamesdesign game-pathfinding, 18KB A* canvas overwritten 06-14) and again via code-level investigation naming it "M1"; "Fixes drafted, not implemented" as of the most recent evidence.
- **what:** A tool/game lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page, link_resolution_rebuild, or the ordinary needs_content_page path) regenerates page_components from plan_sections and replaces the widget with a generic-text-block; the prose-based content-regression guard compares only visible text length after stripping tags, so it is structurally blind to script-heavy widgets. Mechanistically: `save_page_sections_action.go` does `DELETE FROM page_components WHERE page_id=$1` then re-inserts only the sections the content writer supplied — any side-written row not in that list (including a tool/game widget inserted by create_tool_component/deploy_tool at position 2) is destroyed on the next build. Old content is snapshotted to page_component_history (source='save_page_sections_overwrite') before delete, so wipes are recoverable/detectable. Fix layers drafted: an interactivity-aware save guard + carry-forward of interactive sections in save_page_sections, source_item_id stamping into page_component_history, and routing link maintenance through a preserve-sections path (page_rerender was ruled out for CTA re-resolution since it doesn't re-run link logic).
- **sources:** 005(1) hazard block; 002(4)#Interactive-page hazard; 016 §9; 016b Part 4; idea.uk/TODO_chassis_and_idea_uk(1).md#P3; tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.1-2.2,#3,#7; tools/PLAN_tool_widget_clobber(11).md#2-confirmed-findings,#5-fix-options
- **relations:** Adoption interactivity misroute (TL-002, the M2 alternate cause for the same symptom); Two divergent tool-creation paths (TL-003); phantom-CTA resolution bug (separate); Canonical tool-page section-shape design question (TL-009)
- **verify-later:** save_page_sections_action.go patched version deployed?; page_component_history.source_item_id population; page_component_history rows with source='save_page_sections_overwrite'

### TL-002 — Adoption interactivity misroute (canonical prefix desync, "M2")
- **status:** deployed
- **status-evidence:** "b1 confirmed... RESOLVED 2026-05-26 → keyed the feature map by the canonical name in buildPageFeatureMap."
- **what:** A page_type='tool' page rendering description-only prose but no widget can have two distinct causes needing different fixes: M1 is the clobber mechanism (TL-001); M2 is that `buildPageFeatureMap` keyed the feature map by the raw page name while the routing loop looked up `CanonicalisePage`-canonicalised names (the tool branch adds a `tool-` prefix), so tool lookups always missed → empty Features → static needs_content_page route (games, already `game-`prefixed, matched correctly). For gamesdesign, M2 was the confirmed root cause. Fixed by keying buildPageFeatureMap on the canonical name resolved from plan["pages"], landing routing and content attachment in the same namespace — a self-contained one-function change.
- **sources:** WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#root-cause-m2-corrected-after-verification,#diagnostic-recipe,#potential-solutions; tools/PLAN_tool_widget_clobber(11).md#2.7-resolved,#5b-tasks; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t1
- **relations:** CanonicalisePage (029); tool-recreation-handler; Tool widget clobber (TL-001); Canonicalise tool page identity across surfaces (TL-010)
- **verify-later:** apply_adoption_plan_action.go buildPageFeatureMap; datahelpers/page_canonical.go CanonicalisePage

### TL-003 — Two divergent tool-creation paths (novel vs fork)
- **status:** deployed
- **status-evidence:** Documented as both existing, currently-running code paths.
- **what:** `create_tool_component_action.go` (the "novel" path) never sets pages.sections, leaving it default `[]`; `deploy_tool_action.go` (the "fork" path) sets pages.sections to `["hero-tool","tool-guide-intro","<toolFunction>","tool-cta"]`. Both side-write the widget into page_components at position 2 and queue needs_content_page. The novel path is more exposed to the clobber mechanism (TL-001) since the widget is a member of no section list anywhere.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.3,#9
- **relations:** Tool widget clobber mechanism (TL-001); Canonicalise tool page identity (TL-010); Fork-on-deploy tool ownership model (tool-library TLIB-001)
- **verify-later:** create_tool_component_action.go, deploy_tool_action.go; idx_cc_tool_function_unique partial index behaviour

### TL-004 — Post-adoption tool-recreation detection check (T2) + tool-recreation-handler workflow
- **status:** deployed
- **status-evidence:** "T2 — check_tool_recreation_needed.go... Deployed. Backfills automatically on next discovery run, if recreation works" — confirmed independently by two units.
- **what:** A discovery_checks package check finds page_type IN ('tool','game'), status='active' pages with no widget, sources interactive_features from adoption findings via the same canonical-name transform as the misroute fix (TL-002), and emits needs_tool_recreation (7-day per-page cooldown, item key deliberately distinct from adoption's needs_page:<name> to avoid collision). Pages with no captured features are surfaced but deferred to the tool-suggester/generation path. Doubles as the backfill mechanism for pre-existing widget-less pages. The handler workflow: recreate_tool (Opus 64k, timeout 2400s, execute_llm_prompt) → check_tool_completeness → validate_tool → save_page_sections → update_page_status → spawn_rerender. Mode must be `recreate_tool` (not `recreate` — load_existing_content skips unless mode matches, a documented gotcha). Interactive pages are routed here from adoption via buildPageFeatureMap (the T1 routing fix, TL-002).
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b; tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2; tools/PLAN_tool_widget_clobber(11).md#5b-tasks; PLAN_pathfinding_missing_game.md#2; NOTES(44) 2026-06-25/26; HANDOFF_page_pipeline(11).md#4
- **relations:** Adoption interactivity misroute (TL-002); check_tool_health blind spot (TL-011); item_key mis-key
- **verify-later:** check_tool_recreation_needed.go; tool-recreation-handler default_config/agent_definition; check_tool_completeness_action.go; idx_swi_dedup

### TL-005 — Recreation-loss defect (correctly-routed recreation still produces no deployed widget)
- **status:** unknown
- **status-evidence:** "Not confirmed: that widgets now actually deploy... correct routing → completed recreation → no widget. Hold the trigger." Diagnosis was interrupted mid-investigation when a parallel adoption chat reset the underlying state.
- **what:** Query evidence showed all five games on gamesdesign.co.uk had routed correctly to tool-recreation-handler and whose recreation work items completed, yet all five had no deployed widget component and no inline `<script>` section. So the routing fix (TL-002) is necessary but not sufficient — something downstream prevents the widget from landing. Candidate mechanisms considered: recreation mis-targeted a parallel tool-game-* page (TL-006), M1 clobber (TL-001), handler completing without persisting, or a snippets false-negative (inline `<script>` extracted to /assets/js/snippets.js).
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§2.9,§5b; tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§5; tools/PLAN_tool_widget_clobber(11).md §2.8,§5b
- **relations:** Tool widget clobber mechanism (TL-001); tool-game-* duplicate pages (TL-006); Post-adoption detection check (TL-004)
- **verify-later:** re-run queries R1-R3/L/M/N1/N2 against current gamesdesign.co.uk state; check page_component_history for a clobber snapshot on a game page

### TL-006 — tool-game-* duplicate pages (T5)
- **status:** unknown
- **status-evidence:** "T5 — tool-game-* duplicate pages (step 8)... Pending re-observe... May have been wiped by step-9 reset."
- **what:** Five page_type=tool, build_status=planned pages surfaced that duplicate the five existing games by name (tool-game-<name>). Candidate mechanisms: tool-recreation-handler building a separate page instead of populating the original interactive page, or a planner/reconciler role-divergence in the canonicalisation family. The duplicates vanished in a step-9 state reset before their origin could be confirmed.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§5b; tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§6
- **relations:** Recreation-loss defect (TL-005); Canonicalise tool page identity (TL-010)
- **verify-later:** query M (who created tool-game-* pages) re-pointed at current state

### TL-007 — Tool doc header system (sentinel header, stripped at deploy, provenance)
- **status:** deployed
- **status-evidence:** "DONE 2026-06-11: 019/020 CONSOLIDATED — splice files retired"; 019 full lifecycle table incl. StripToolDocHeader call sites and tool_health checks — but a rollout runbook flags "apply order is load-bearing" and no completion claim was found for that specific rollout track, so applied-state may lag the design.
- **what:** Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never ids/dates; no closing `*/`). It never ships: StripToolDocHeader runs at three outbound assembly points (single-page rerender, collectJSAssets, bulk rerender); DB rendered_html retains it for audit parity. Creation gate (`HasToolDocHeader`) validates presence in create_tool_component; tool-generator/tool-improver prompts updated to emit it; auditor audits code AGAINST its stated behaviour; malformed (opener without closer) is left in and flagged by two new tool_health tier-1 checks (no_doc_header warning, malformed_doc_header error). Paired with new source_agent_type/source_orchestration_id provenance columns on content_components, mirroring knowledge_base's existing provenance pair.
- **sources:** 019#Tool Doc Header; 020 tool_health tier-1 checks; docs019/RUNBOOK_thin_slice(27).md#tool-doc-header-rollout; NOTES_running_synthesis_principles(59) 2026-06-11 tool-doc entries
- **relations:** per-tool travelling docs (037); tool-auditor; doc_plans/doc_notes (the tools thread's later system)
- **verify-later:** platform/content/tool_doc_header.go; content_components source_% columns; '=== tool-doc ===' in html_template rows; whether the provenance-columns → prompts → binary rollout sequence was ever fully applied (repeatedly flagged "apply-ready, not yet applied" across notes through 2026-07-06)

### TL-008 — The tool verification ladder (Tiers 0–4)
- **status:** deployed
- **status-evidence:** "The verification ladder is whole (Tier 0/1/2/4) and closes on both outcomes" — RUNBOOK position 2026-07-12; earlier snapshots (2026-06) described only 3 tiers with the LLM-audit tier "planned." Tier numbering has drifted across docs: an earlier framing called the LLM code-review "Tier 2"; the mature framing reassigns LLM audit to Tier 3 (still Phase B/unbuilt as a criteria-driven check) and gives Tier 2 to the static contract-presence checker.
- **what:** Cheap-to-expensive tiers, each catching a different class. Tier 0: generation-time output integrity (HasToolDocHeader gate + check_tool_completeness, deliberately flags-but-passes). Tier 1: structural post-deploy (check_tool_health — deploy status, HTML/template present, script/style/@media, hex/external-dep warnings, doc-header checks; blockers create improve_tool; 30-day/tool cooldown). Tier 2: static contract-presence against deployed HTML under the anchor rule (check_tool_acceptance, TL-013). Tier 3: acceptance audit via tool-auditor's Sonnet code review across six categories, findings routed by confidence (certain/likely → improve_tool, possible → needs_human_review), quality_score 1-10 tracked — separately also framed as "Phase B, unbuilt" when judged against the travelling-PLAN criteria specifically. Tier 4: behavioural — drive the deployed tool in headless Chromium until criteria pass (TL-014/TL-015). Standing rule: never read a Tier-2 pass as "the tool works" — that claim belongs to Tier 4. Tool removal remains a human decision via dashboard.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_travelling_docs(6).md#tool-assurance; OVERVIEW_self_verifying_tools.md#mechanism-2; 020#tool_health,#tool-auditor; 019#Tool Quality Standards; docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; 088_tool_auditor_agent.sql; 142_enable_tool_acceptance_check.sql; 145_tool_acceptance_agent.sql; 146_enable_tool_acceptance_due.sql
- **relations:** every tier concept below; "passed checks ≠ working" (TL-012); check_tool_health cooldowns
- **verify-later:** check_tool_completeness + check_tool_health + discovery_checks/check_tool_acceptance.go + browser-runner adapter, all in the chassis repo

### TL-009 — Canonical tool-page section-shape design question and fix options
- **status:** aspirational
- **status-evidence:** "Fix options (structural-first; not yet implemented)."
- **what:** Raises and answers (as a design decision, not yet built) whether a tool page even wants generic hero/guide-intro/CTA sections, or just the widget. Three options recorded: (1) make the widget a first-class section in whichever authority the build reads (preferred); (2) right-size the tool page's canonical section list (preferred, paired with 1); (3) make save_page_sections structure-aware as a safety-net guard. Notes content_guidance already instructs the writer not to regenerate the widget, but the writer has no mechanical way to honour that.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §4,§5; tools/PLAN_tool_widget_clobber(11).md#2-confirmed-findings,#5-fix-options
- **relations:** Tool widget clobber mechanism (TL-001); site_plan as build authority; content-governance
- **verify-later:** whether plan_sections_action.go now emits a tool/embed section for page_type='tool' pages

### TL-010 — Canonicalise tool page identity across surfaces (T3)
- **status:** aspirational
- **status-evidence:** "T3... Open, independent... Low risk; can land at any time."
- **what:** create_tool_component and deploy_tool build page name/url/page_type ad hoc, diverging from the canonical datahelpers.CanonicalisePage helper that adoption and the planner already use. Proposed fix: route both tool actions through the same canonical helper. Flagged as a gap in an earlier canonicalisation-family deliverable that covered only two other files.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b; tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2,§6,§8
- **relations:** Two divergent tool-creation paths (TL-003); tool-game-* duplicate pages (TL-006)
- **verify-later:** create_tool_component_action.go, deploy_tool_action.go, datahelpers/page_canonical.go

### TL-011 — check_tool_health INNER JOIN blind spot
- **status:** partial
- **status-evidence:** "check_tool_health blind spot. Its INNER JOIN content_components → page_components means a tool with no linked page_components row... is invisible."
- **what:** The Tier-1 tool health check joins content_components to page_components with an INNER JOIN, so a page_type='tool' page with zero linked components (post-clobber, or never-generated) is invisible to it and the check silently reports "no tools" as a pass. T2 (TL-004) partially closes this by detecting the same condition independently, but the original check itself was not corrected.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8; tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §9
- **relations:** Post-adoption detection check (TL-004); Recreation-loss defect (TL-005)
- **verify-later:** check_tool_health.go join logic

### TL-012 — "Completeness + validation passed" ≠ working — twice demonstrated
- **status:** deployed
- **status-evidence:** "The June recreation introduced the economy-simulator's two bugs and passed; run 2 of the repair faithfully recreated them and passed."
- **what:** The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity, not behaviour. The same game shipped broken twice while passing every existing check — the June 2026-06-05 recreation introduced the bugs (proven from tool_recreation_training rows and the origin game.js, which has neither bug), and repair run 2 recreated them while its own note truthfully said "completeness + validation passed."
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45; OVERVIEW_self_verifying_tools.md#problem
- **relations:** Tier 4 (TL-008); seam rule; economy-simulator case
- **verify-later:** tool_recreation_training rows for page d9a8e6e8 dated 2026-06-05

### TL-013 — Tier-2 static acceptance checker + the anchor rule (discovery check `tool_acceptance`)
- **status:** deployed
- **status-evidence:** "Stage 5 — LIVE 2026-07-10 (first sweep proven)" — run cd0d9731 on v1.0.1107 produced exactly the pre-verified findings (2 improve_tool items + 2 acceptance-fail notes, check-level precision confirmed); anchor rule "SETTLED" 2026-07-09 after the #tableWrap inspection.
- **what:** A browserless discovery check (sibling of tool_health): loads the current travelling PLAN's criteria fence, fetches the deployed page (bounded 12s/2MB, cached per run), and evaluates the statically-visible subset under the ANCHOR RULE — validate only a criteria selector's ANCHOR (leftmost id/class token) against html_template (`#tableWrap` exists ⇒ `#tableWrap tr` passes since rows are JS-built and Tier 4 asserts them for real; `#xpTableBody` exists nowhere ⇒ fails ⇒ drop). Static validation can confirm a selector but never refute one — the DOM is constructed at runtime, so a check must never delete a rule merely because the DOM doesn't statically contain it. Plus shell checks (tool-doc header not leaked, no `<no value>` residue). No criteria → a needs_criteria note (30-day cooldown), never a fake pass. Failures → one improve_tool item (criteria embedded as acceptance_test, 7-day cooldown, cancelled items excluded per migration 146) + an acceptance-fail note. Scope limit by construction: only generator-created tools have content_components rows; adopted/recreated page-section tools are invisible to Tier 2 (reached instead via Tier 4, which operates on pages). Implementation detail: CSS class tokens are whitespace-delimited (Go regexp `\b` wrongly splits on hyphens).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5,#stage-5-rule; RUNNING_NOTES_travelling_docs(39).md#stage-5-built,#stage-5-live,#rev39,#rev40; HANDOFF_2026-07-10…md#§1,§2; OVERVIEW_self_verifying_tools.md#tier-2
- **relations:** migration 142; Tier 4 (TL-014); Composer selector invention (TL-016)
- **verify-later:** discovery_checks/check_tool_acceptance.go; design-discovery-agent run_checks list; anchor extraction + class-token comparison tests

### TL-014 — tool-acceptance-agent — Tier 4 self-driving orchestrator + continuous sweep + iteration loop
- **status:** partial
- **status-evidence:** First machine acceptance-run note (run bf330ac6, 2026-07-12, "Tier-4 acceptance PASSED — all 3 evaluated checks"); fail path proven live via a controlled reverted test; continuous sweep migration 146 applied 2026-07-12 but "the continuous sweep is NOT in the binary" as of v1.0.1111 (untracked-file trap) — "GATE: continuous acceptance activates on the next image built from 83ba9bd4+."
- **what:** An agent (migration 145) closing the loop with zero humans: ensure_site_record → load_docs → request_browser_run (Kafka await; resolves the tool's deployed URL from pages itself; NO-OP skips without awaiting when the PLAN has no criteria) → judge_acceptance_results → complete. Judge recomputes the verdict from results: all pass → acceptance-run note; any fail → acceptance-fail note + ONE improve_tool item (criteria embedded as acceptance_test, handler tool-improver); component-less recreated/adopted tools get the note but no item — logged honestly for manual routing. A companion discovery check (tool_acceptance_due) emits one acceptance_run item per active tool with a deployed page and current criteria unless a verdict landed within 7 days or a run is open; deliberately not triggered by post-creation/post-improve hooks (they'd fire before the page redeploys); items land straight at 'triaged' (acceptance needs no human judgment); priority 90 so acceptance tests the NEW page after builds/rerenders. Full iteration loop: deploy → acceptance run → failing criterion → improve_tool item (bounded by max_fix_attempts) → fixer loads PLAN+NOTES first → fix → append note → redeploy → re-run; the one link proven only with a synthetic input at unit close was a real failure flowing through tool-improver and back.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tool-acceptance-agent-built,#tier-4-self-driving,#fail-path-proven,#tier-4-continuous,#v1.0.1111; README_summary_paragraph2_for_discussion.md; 087_TRIGGER_tool_acceptance.sh; HANDOFF_2026-07-10…md#T10,T11; RUNBOOK_travelling_docs(38).md#§5; PLAN_tool_acceptance_runner(2).md#flow
- **relations:** browser-runner adapter; Criteria contract v0 (TL-015); Tier-2 static checker (TL-013); tool-improver
- **verify-later:** platform/orchestration/actions/tool_acceptance_actions.go; agent_definitions row tool-acceptance-agent; migration 145; discovery_checks/check_tool_acceptance_due.go in the deployed image; an improve_tool item with source 'acceptance' processed end-to-end by tool-improver

### TL-015 — Criteria contract v0 (check-type vocabulary + profiles) + browser-runner-adapter design
- **status:** partial
- **status-evidence:** "Status: initial plan (P0 not started)" in the earliest planning doc; by 2026-07 "P0 implements 3 of 7 check types," with "the composer emitted 'action':'select'... a verb the Tier-4 criteria vocabulary must now define" left open; a later repo commit ("browser-runner-adapter: commit the full Tier-4 adapter") suggests the adapter code itself may now exist.
- **what:** The machine-readable criteria schema consumed by Tier 4: `profiles: [desktop, mobile]`; check types selector_exists, selector_count, no_console_errors, asset_loads, interaction (fill/click/select steps + expect), no_horizontal_overflow, page_status_ok. Deterministic only in v0 — no LLM drives the browser. Desktop = Chromium 1366×900; mobile = one stable Playwright device descriptor (emulation first; real devices out of scope). Phasing: P0 boot checks → P1 interpreter+mobile → P2 interactions → P3 screenshots (via the existing Backblaze deploy path) → P4 optional LLM-exploratory mode. The executing component is a browser-runner-adapter (Playwright+Chromium, mirroring analyser-adapter) driving a deployed tool page under desktop+mobile profiles, judging criteria pass/fail and feeding failures back as improve_tool items.
- **sources:** PLAN_tool_acceptance_runner(2).md#criteria-contract,#profiles,#phasing,#Aim,#Criteria-contract; RUNBOOK_travelling_docs(38).md#stage-6; tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md
- **relations:** browser-runner adapter (P0); multi-page tool criteria (TL-017, open question — url_role field); tool-acceptance-agent (TL-014, the orchestrator that consumes this contract)
- **verify-later:** criteria interpreter coverage in run_checks_action.go; whether the "select" verb was added; browser-runner-adapter deployment state; max_fix_attempts convention

### TL-016 — Composer selector invention & the delivered-reality principle (Option B)
- **status:** deployed
- **status-evidence:** CONFIRMED 2026-07-09 (#xpTableBody/#statsStrip token_anywhere=f); second sighting caught by Tier 2 itself 2026-07-10 (kebab #drop-chance vs real camelCase #dropChance); "DECIDED 2026-07-10: Option B — inline reality (user: 'I choose option B and surrender')"; migrations 143 (PLANs superseded) + 144 (composer fixed) applied and verified.
- **what:** Two recurring failure classes in machine-written acceptance criteria, and their durable remedies. (1) The PLAN-composer LLM invented DOM ids for assertion targets despite an explicit "never invent a selector" instruction — the rule held for controls it acts on and failed for things it asserts on; first instance corrected by a guarded supersede migration (which itself initially refused a valid runtime selector, leading to the anchor rule, TL-013); second caught automatically by the live Tier-2 sweep and corrected by migration 143. Remedy is Tier-2 static validation, not sterner prompts. (2) The composer had asserted a designed-but-never-built JS extraction (asset_loads /tools/assets/<fn>.js) in every PLAN, so Tier-2's first sweep failed every tool on it by construction; principle settled on record: criteria must describe what the system DELIVERS (all JS in fact ships inline), aspirations live in roadmaps — if extraction ever ships, PLANs supersede forward again. Corollary: the composer's standard checks became boots/console/status/mobile-fit + optional interaction from real selectors.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#stage-5-live,#rev39,#rev40; 0NN_supersede_xp_curve_plan_selectors(2).sql; RUNBOOK_travelling_docs(38).md#task-3-proven,#stage-5; RUNBOOK_travelling_docs(38).md#stage-5 (Option B block); HANDOFF_2026-07-10…md#§2; PLAN_travelling_docs(6).md#rollout-outcomes; 136_supersede_xp_curve_plan_selectors.sql; 143_supersede_plans_inline_delivery.sql; 144_composer_inline_delivery.sql; 113_site_asset_renderer.sql (the extraction design it contradicts)
- **relations:** anchor rule (TL-013); supersede versioning; Inline-JS extraction Path 1 (tool-pipeline TP-003, the abandoned mechanism this superseded)
- **verify-later:** xp-curve PLAN v1→v2 chain + its correction note in doc_notes; current PLAN fences have no asset check; compose_plan prompt (four checks, inline delivery line); whether asset extraction ever ships (would trigger forward supersede)

### TL-017 — Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block)
- **status:** deployed
- **status-evidence:** "DECISION: acceptance criteria live in the tool's doc_plans PLAN" (2026-07-04 rev 4); consumed live by Tier 2 and Tier 4 checkers.
- **what:** The per-tool definition of *working*. Candidates judged on key/lifecycle/owner and rejected: site_specs (right machinery, wrong key — site-scoped, per-site copies drift); site_plans/directives (wrong lifecycle and owner — churniest artifact, planner-owned; "never store the bar in the artifact that regenerates most"); findings' acceptance_test (right pattern, wrong duration — dies with the work item; the standing criteria SEED it). The PLAN wins on all three axes. Format: a machine-extractable fenced ```criteria JSON block (the tool-doc-header precedent), extracted by load_doc_context as criteria_json; lifts to a column only on volume. Per-site parametrisation goes to direction.must_have, not the PLAN. Multi-page tools additionally add a "Page set & inter-page contract" PLAN section and may need per-page checks (a url_role field) — explicitly gated on the pending preserve-sections re-render/interactivity-aware save guard before scaling page counts.
- **sources:** PLAN_travelling_docs(6).md#where-acceptance-criteria-live,#tool-assurance; 001_README_acceptance_criteria.md; RUNNING_NOTES_travelling_docs(39).md#rev4; RUNBOOK_travelling_docs(38).md#§2,#§5; PLAN_tool_acceptance_runner(2).md#open-questions
- **relations:** verification ladder (TL-008); findings acceptance_test/max_fix_attempts (improvement-loop IMP-006/IMP-027); direction.must_have; interactive-section clobber (TL-001)
- **verify-later:** criteria fence extraction in load_doc_context; has_fence on live PLANs; save_page_sections interactivity guard deployment status

### TL-018 — Recreation writes page sections — component-less tools and their visibility gap
- **status:** deployed
- **status-evidence:** Established by query 2026-07-09 ("pages.sections is EMPTY... the 32 KB game body exists only as deployed HTML in the sites repo").
- **what:** tool-recreation-handler ends save_page_sections → update_status → deploy_page and never creates a content_components row — adopted/recreated tools exist only as page sections + deployed HTML (source in adoption-crawl research_results; spec.mode="recreate" is the handshake set by apply_adoption_plan). Consequences: no component address for tool-improver; invisible to Tier 2 by construction (Tier 4 reaches them via pages instead); NOTES subject must be pipeline-scoped. site_plan_sections is site-plan STRUCTURE, not HTML.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev41,#rev42,#rev43; HANDOFF_2026-07-09…_1_.md#§4; RUNBOOK_travelling_docs(38).md#task-5-record
- **relations:** dangling-doc rule; adoption pipeline; Tier-2 scope limit (TL-013)
- **verify-later:** tool-recreation-handler workflow steps; research_results result_types

### TL-019 — Behavioural QA loop for tools & games (planned Tier 3+ headless-browser testing)
- **status:** aspirational
- **status-evidence:** "Status: proposed (2026-06-06)"; also referenced from the adoption thread as FUTURE work.
- **what:** A standalone, slower-cadence QA loop motivated by real defects (Jelly Invaders degrading over time, P2P host replies not reaching mobile clients, untested cross-browser/mobile variants) that runs generated tools/games in an isolated multi-engine Playwright pod over time under synthetic drive, to catch defect classes invisible to a single render/screenshot: temporal degradation, cross-browser divergence, mobile-specific layout/touch bugs, and multi-context networked/relay failures. Correctness judged via a three-layer oracle: generic deterministic invariants, type-specific assertions, and LLM-as-judge over a screenshot/video series — with auto-fix gated to high-confidence deterministic findings only. Reuses the existing check→work-item→improver pipeline. This is the heavier, temporal/cross-browser successor to the Tier-4 acceptance runner (TL-014/TL-015), which is deterministic and single-pass.
- **sources:** HANDOFF_2026-06-06#future; HANDOFF_2026-06-09#later-parked; tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md#1-Why,#4-The-headless-pod,#5-The-oracle-problem,#10-Phasing
- **relations:** Tool acceptance runner (TL-015, the deterministic v0 layer); Games quality lifecycle parity (NEW:games-lifecycle GML-001); tool-lifecycle overall
- **verify-later:** whether any phase has been built; qa_runs/last_qa_at storage location; PLAN_tools_games_behavioral_qa_loop.md

### TL-020 — Validation observability: structured rejection logging (recordValidationRejection)
- **status:** deployed
- **status-evidence:** "Validation observability deployed: store_generated_component_action.go recordValidationRejection writes a structured agent_error_log row on every pre-store rejection... guide-list's attempt-1 failure was captured exactly — one SQL row, no pod-log forensics" (2026-05-11).
- **what:** Every component pre-store validation rejection writes a structured agent_error_log row (severity warning for bookkeeping vs error for structural; orphan/unknown field names as typed JSONB arrays), replacing pod-log forensics. Companion pattern: the retry budget of 3 is calibrated for the single-bookkeeping-orphan failure class seen in Tier-D regens (tool-list missed card_link_label, guide-list read_guide_label); a central label registry would prevent the class entirely (idea, not built).
- **sources:** FOCUS_directory_builder_and_list_components.md#tier-d-converge
- **relations:** component-creator; chrome-template gate (would reuse the same gate/log shape); F1 field-contract guard (component-lifecycle CLC-003)
- **verify-later:** store_generated_component_action.go recordValidationRejection

### TL-021 — Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option) — superseded
- **status:** superseded
- **status-evidence:** Archive: "It returns 2–5 suggestions." Live: "It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate."
- **what:** The earliest tool-suggester design forced the LLM to always propose at least two tools per site. Replaced by an explicit zero-is-valid design, directly tied to the same failure class as the tag-based matchToolToSite function (tool-library TLIB-009).
- **sources:** old/older1/012_tool_lifecycle_guide.md#"Agent: tool-suggester"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Agent: tool-suggester"
- **relations:** Tag-based deterministic tool-to-site matching (tool-library TLIB-009)
- **verify-later:** check tool-suggester's current prompt for the zero-suggestions instruction

### TL-022 — forked_from NULL collision risk on novel tools
- **status:** unknown
- **status-evidence:** "forked_from NULL on novel tools... Two sites generating the same function would collide. Latent; not today's bug."
- **what:** create_tool_component omits forked_from, so novel/generated tools are classified as library tools by the partial unique index idx_cc_tool_function_unique (function) WHERE component_level='tool' AND forked_from IS NULL AND is_active. Two different sites independently generating a tool with the same function name would collide.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8
- **relations:** Two divergent tool-creation paths (TL-003)
- **verify-later:** idx_cc_tool_function_unique definition; whether any collision has actually occurred

### TL-023 — Fork-divergence detection for library tools (proposed)
- **status:** aspirational
- **status-evidence:** "IMMEDIATE WIN INSTEAD: FORK-DIVERGENCE detection — pure SQL discovery check (tier-1, zero cost)."
- **what:** A proposed zero-cost SQL discovery check comparing a deployed fork's html_template hash against its forked_from library original to answer "which forks are unmodified / safe to bulk-push a library change" — deliberately deferred building full code-symbol indexing of tools (each tool is one IIFE, thin symbol pickings; tool discovery already solved via semantic_tags) until a concrete consumer needs it.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tools/provenance/docs design entry
- **relations:** Tool doc header system (TL-007); Fork-on-deploy tool ownership model (tool-library TLIB-001)
- **verify-later:** whether this check was ever built

### TL-024 — Component quality tracking (0–100 score)
- **status:** deployed (reconciled — see verify-later)
- **status-evidence:** "None of these fields are required by the existing pipeline — they are additive... selector will use them when present and ignore when NULL."
- **what:** Additive quality fields on content_components computed by a compute_component_quality action, with indexes for auditor queries (below threshold OR unscored) and planner preference (higher quality per function). Distinct from avg_quality_score in the selector metadata set (tool-library TLIB-016). Cross-reference: component-quality-auditor (tool-library TLIB-015) actively creates needs_component_regeneration items from this score in production, and a specific boundary bug (auto-regen threshold, TL-026) exists in that consuming code — so despite this entry's older "additive, optional" framing, the field is in active use.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-quality-tracking
- **relations:** component selector metadata (tool-library TLIB-016); component-quality-auditor (tool-library TLIB-015); component-quality-auditor auto-regeneration threshold (TL-026)
- **verify-later:** compute_component_quality action in registry; populated quality_score values

### TL-025 — Component versioning (component_versions table) — schema-mode origin
- **status:** deployed (reconciled — see relations)
- **status-evidence:** Table created in schema-mode migration (008 PART 3); page_components.component_version_id exists in live dump with comment "if versioning enabled" — this unit's own evidence says "unclear whether any writer maintains it," but a separate unit (component-lifecycle) documents active writers and real incidents against this same table.
- **what:** Versioned snapshots of component templates (html_template, css_template, input_schema per version_number), originally designed so strict-mode pages could pin a specific template version. Referenced as an optional backup target in later template-fix migrations. This entry's own evidence base predates confirmation of active use; see NEW:component-lifecycle's "Component versioning via component_versions" (CLC-009) for the fuller, more current picture — StoreGeneratedComponentAction and update_component_html both write to this table in production, with known coverage gaps (some write paths historically bypassed it).
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#PART3; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** schema-mode subsystem (abandoned); site_plan_sections.component_version_id (planner provenance); Component versioning via component_versions (component-lifecycle CLC-009, fuller/current picture)
- **verify-later:** row count in component_versions; writers in Go

### TL-026 — Component regeneration in place (store_generated_component mechanics) + a naming-collision incident
- **status:** deployed
- **status-evidence:** 083 result: brief-explanation updated in place (same id, created_at unchanged, status 'regenerated', component_versions snapshot, needs_rerender raised) — "matches the documented behaviour (003 §348)."
- **what:** store_generated_component looks up an existing component by the LLM's EMITTED function (forked_from IS NULL); if found, it snapshots the current row to component_versions (MAX+1), UPDATEs in place (component_id preserved → all page/site FKs keep resolving), sets template/schema/js_content/render_mode/is_active, then markPagesPendingRebuild raises ONE needs_rerender per affected site. Determinism hazard: regeneration keys on the emitted function name — an unpinned LLM can emit a different name and INSERT a stray duplicate (the 081 'general-hero' incident); the mitigation is to pin the function name in the description. Pre-store validation rejects `<no value>` templates and checks placeholder/schema parity.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30,#2026-06-30-~18:35,#2026-07-01-~12:46; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e
- **relations:** shared library guard; component-quality-auditor; StoreGeneratedComponentAction regeneration branch (component-lifecycle CLC-002, the same mechanism documented via a different incident — the F1-F8 saga); F4 regen-vs-create keyed on LLM-chosen function (component-lifecycle CLC-006, the same determinism hazard)
- **verify-later:** store_generated_component_action.go lookup + snapshot + markPagesPendingRebuild; component_versions rows

### TL-027 — component-quality-auditor auto-regeneration threshold (boundary bug)
- **status:** deployed
- **status-evidence:** Read from its default_config 2026-06-29: creates needs_component_regeneration items only for quality_score < 50, handler component-creator.
- **what:** The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant three vonc shells scoring EXACTLY 50 were never auto-picked-up (explaining zero queued items and requiring manual triggers). Its item shape confirms the designed regen path keys on function and routes to component-creator. Boundary-condition gap worth a rule review; also the future home of the autonomy plan's maintenance detections.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~21:00; docs/PLAN_dynamic_sections_and_loaders(4).md#maintenance
- **relations:** Component regeneration in place (TL-026); component-quality-auditor (tool-library TLIB-015)
- **verify-later:** component-quality-auditor default_config condition; quality_score distribution at exactly 50

### TL-028 — Store-path template validation (+ pending `<script>`-balance hardening)
- **status:** partial
- **status-evidence:** Existing checks confirmed in code 2026-06-29 (`<no value>` rejection, placeholder/schema parity, unclosed `<style>`, section/div presence); the `<script>` balance check + separateInlineJS truncation warning remain "STILL MISSING" as of 2026-07-09.
- **what:** store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT an unclosed `<script>` — the gap that let a provocation-card component ship a truncated inline script that swallowed the page footer at render. Hardening definition: add a `<script>` open/close balance check (reject or flag-for-regeneration) plus a truncation warning in separateInlineJS. Prevents the class "truncated template ships and breaks the page."
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30,#2026-07-03-~13:25; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-g; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** legacy un-extracted shells (the truncation instance); Mode A/B taxonomy; Mode-B rendered-artifact templates (TL-030)
- **verify-later:** store_generated_component_action.go validation block for script balance

### TL-029 — Component-creator invocation contract (dual placement + quote-free description)
- **status:** partial
- **status-evidence:** NOTES_brief-explanation 083 (2026-07-01) "SUCCEEDED (in-place UPDATE)" via dual placement; framework fix "PATCH_validate_input_contract.go — drafted, not deployed."
- **what:** Manually invoking component-creator (spawn+call) must satisfy BOTH the input_contract (top-level required fields — call_agent validates against top-level extracted fields) AND the workflow's field paths (input_data.spec.*): the working pattern places section_type both top-level and inside spec, pins the function name in the description so the store lands as an in-place UPDATE (else a stray component INSERTs), and keeps the description quote-free to survive the kcat/JSON pipeline. The generic build-dispatch-loop cannot satisfy top-level-required contracts (same class); the durable fix — a contract validator accepting top-level OR input_data.spec.{field} — is drafted, not deployed.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (081/082/083 arc); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Component-creator-input
- **relations:** Component regeneration in place (TL-026); shared component library semantics (tool-library TLIB-023)
- **verify-later:** call_agent contract validation code; PATCH_validate_input_contract.go status

### TL-030 — Mode-B rendered-artifact templates (components stored as rendered output)
- **status:** deployed
- **status-evidence:** VERDICT §2 (2026-07-09): "rendered_html == html_template with all '<no value>' removed — which the byte counts confirm exactly."
- **what:** A component corruption class: html_template full of bare `<no value>`, zero `{{.}}` slots, empty input_schema — the stored template IS a rendered artifact. Consequences: render is a pure function of the template (predictable to the byte — used twice as an acceptance test); content_data is dead weight; repair_template_slots cannot repair them (zero `</no>` tags → needs_regeneration); for runtime-fill shells the emptiness is accidentally exactly what the loader needs, so regeneration must consciously re-establish the empty-shell contract or sections ship with baked copy.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#2,#8.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-24
- **relations:** Runtime-fill guards (improvement-loop IMP-047); component selector/creator regeneration path; Corrupted component templates bridge (improvement-loop IMP-018); Store-path template validation (TL-028)
- **verify-later:** components with `<no value>` and 0 schema fields fleet-wide


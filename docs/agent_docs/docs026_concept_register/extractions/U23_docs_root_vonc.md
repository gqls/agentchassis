# EXTRACTION U23 — docs/ root files (vonc/Spark session corpus)
Extracted 2026-07-13. Files in scope: 168. Concepts found: 66.

Unit character: the docs/ root is dominated by the vonc.com / "Spark" build-and-fix
campaign (2026-06-22 → 2026-07-09) in heavily versioned families. The RUNNING_NOTES
families are append-only (byte-prefix verified), so family-latest reads cover them;
the RUNBOOK families were edited in place and were diffed for dropped concepts.
Proposed NEW category: `NEW:page-build-pipeline` (page-build-handler →
page-content-writer → save_page_sections → deploy; section sources; rerender paths;
silent-success semantics) — enough distinct, load-bearing concepts to back a council
agent, and no existing slug owns them.

## Coverage
| file | treatment |
|---|---|
| docs/016b_debugging_guide_6_(1).md | family-delta |
| docs/016b_debugging_guide_merged(2).md | family-delta |
| docs/016b_debugging_guide_merged(3).md | family-latest |
| docs/082_regenerate_brief-explanation_vonc.sh | header-scan |
| docs/083_regenerate_brief-explanation_vonc.sh | header-scan |
| docs/084_create_provocations-archive-list_vonc.sh | header-scan |
| docs/API_DOCUMENTATION.md | full |
| docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(1).md | family-delta |
| docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md | family-latest |
| docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim.md | family-delta |
| docs/HANDOFF_vonc_write_site_spec_spec_data.md | full |
| docs/PATCH_section_visible_content(1).go.txt | header-scan |
| docs/PATCH_validate_input_contract.go.txt | header-scan |
| docs/PLAN_brief-explanation(1).md | family-latest |
| docs/PLAN_brief-explanation.md | family-delta |
| docs/PLAN_dynamic_sections_and_loaders(1).md | family-delta |
| docs/PLAN_dynamic_sections_and_loaders(2).md | family-delta |
| docs/PLAN_dynamic_sections_and_loaders(3).md | family-delta |
| docs/PLAN_dynamic_sections_and_loaders(4).md | family-latest |
| docs/PLAN_dynamic_sections_and_loaders.md | family-delta |
| docs/PLAN_lobby-grid(1).md | family-delta |
| docs/PLAN_lobby-grid(2).md | family-latest |
| docs/PLAN_lobby-grid.md | family-delta |
| docs/PLAN_provocation-card(2).md | family-delta |
| docs/PLAN_provocation-card(3).md | family-latest |
| docs/PLAN_spark_provocation_pipeline.md | full |
| docs/PLAN_vonc_next_steps(1).md | family-latest |
| docs/PLAN_vonc_next_steps.md | family-delta |
| docs/PROPOSED_MOVES.tsv | full |
| docs/README_summary_paragraph_for_handoff.md | full |
| docs/RUNBOOK_phase2_provocation_js(1).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(10).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(11).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(12).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(13).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(14).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(15).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(16).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(17).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(18).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(19).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(2).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(20).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(21).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(22).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(23).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(24).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(25).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(26).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(27).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(28).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(29).md | family-latest |
| docs/RUNBOOK_phase2_provocation_js(3).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(4).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(5).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(7).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(8).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js(9).md | family-delta |
| docs/RUNBOOK_phase2_provocation_js.md | family-delta |
| docs/RUNBOOK_section_assembly_drop(1).md | family-delta |
| docs/RUNBOOK_section_assembly_drop(2).md | family-delta |
| docs/RUNBOOK_section_assembly_drop(3).md | family-latest |
| docs/RUNBOOK_vonc_migrations(1).md | family-delta |
| docs/RUNBOOK_vonc_migrations(10).md | family-delta |
| docs/RUNBOOK_vonc_migrations(11).md | family-delta |
| docs/RUNBOOK_vonc_migrations(12).md | family-delta |
| docs/RUNBOOK_vonc_migrations(13).md | family-delta |
| docs/RUNBOOK_vonc_migrations(14).md | family-latest |
| docs/RUNBOOK_vonc_migrations(2).md | family-delta |
| docs/RUNBOOK_vonc_migrations(3).md | family-delta |
| docs/RUNBOOK_vonc_migrations(4).md | family-delta |
| docs/RUNBOOK_vonc_migrations(5).md | family-delta |
| docs/RUNBOOK_vonc_migrations(6).md | family-delta |
| docs/RUNBOOK_vonc_migrations(7).md | family-delta |
| docs/RUNBOOK_vonc_migrations(8).md | family-delta |
| docs/RUNBOOK_vonc_migrations(9).md | family-delta |
| docs/RUNBOOK_vonc_migrations.md | family-delta |
| docs/RUNBOOK_vonc_session(1).md | family-latest |
| docs/RUNBOOK_vonc_session.md | family-delta |
| docs/RUNNING_NOTES_vonc(1).md | family-delta |
| docs/RUNNING_NOTES_vonc(10).md | family-delta |
| docs/RUNNING_NOTES_vonc(11).md | family-delta |
| docs/RUNNING_NOTES_vonc(12).md | family-delta |
| docs/RUNNING_NOTES_vonc(13).md | family-delta |
| docs/RUNNING_NOTES_vonc(14).md | family-delta |
| docs/RUNNING_NOTES_vonc(15).md | family-delta |
| docs/RUNNING_NOTES_vonc(16).md | family-delta |
| docs/RUNNING_NOTES_vonc(17).md | family-delta |
| docs/RUNNING_NOTES_vonc(18).md | family-delta |
| docs/RUNNING_NOTES_vonc(19).md | family-delta |
| docs/RUNNING_NOTES_vonc(2).md | family-delta |
| docs/RUNNING_NOTES_vonc(21).md | family-delta |
| docs/RUNNING_NOTES_vonc(22).md | family-delta |
| docs/RUNNING_NOTES_vonc(23).md | family-delta |
| docs/RUNNING_NOTES_vonc(24).md | family-delta |
| docs/RUNNING_NOTES_vonc(25).md | family-delta |
| docs/RUNNING_NOTES_vonc(26).md | family-delta |
| docs/RUNNING_NOTES_vonc(27).md | family-delta |
| docs/RUNNING_NOTES_vonc(28).md | family-delta |
| docs/RUNNING_NOTES_vonc(29).md | family-delta |
| docs/RUNNING_NOTES_vonc(3).md | family-delta |
| docs/RUNNING_NOTES_vonc(30).md | family-delta |
| docs/RUNNING_NOTES_vonc(31).md | family-delta |
| docs/RUNNING_NOTES_vonc(32).md | family-delta |
| docs/RUNNING_NOTES_vonc(33).md | family-delta |
| docs/RUNNING_NOTES_vonc(34).md | family-delta |
| docs/RUNNING_NOTES_vonc(35).md | family-delta |
| docs/RUNNING_NOTES_vonc(36).md | family-latest |
| docs/RUNNING_NOTES_vonc(4).md | family-delta |
| docs/RUNNING_NOTES_vonc(5).md | family-delta |
| docs/RUNNING_NOTES_vonc(6).md | family-delta |
| docs/RUNNING_NOTES_vonc(7).md | family-delta |
| docs/RUNNING_NOTES_vonc(8).md | family-delta |
| docs/RUNNING_NOTES_vonc(9).md | family-delta |
| docs/RUNNING_NOTES_vonc.md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(1).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(10).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(11).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(12).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(13).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(14).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(15).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(16).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(17).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(18).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(19).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(2).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(20).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(21).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(22).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(23).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(24).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(25).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(26).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(27).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(28).md | family-latest |
| docs/RUNNING_NOTES_vonc_v2(3).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(4).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(5).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(6).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(7).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(8).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2(9).md | family-delta |
| docs/RUNNING_NOTES_vonc_v2.md | family-delta |
| docs/SPEC_provocations-archive-list.md | full |
| docs/bundle_minilobby_trim(1).sh | family-delta |
| docs/bundle_minilobby_trim(2).sh | family-delta |
| docs/bundle_minilobby_trim(3).sh | family-delta |
| docs/bundle_minilobby_trim(4).sh | header-scan |
| docs/bundle_minilobby_trim.sh | family-delta |
| docs/dedup-manifest.tsv | header-scan |
| docs/fix_archive_template_display(1).sql | header-scan |
| docs/fix_archive_template_display.sql | family-delta |
| docs/fix_marker_selector.sql | header-scan |
| docs/lobby_grid_install.sql | header-scan |
| docs/lobby_grid_loader.js | header-scan |
| docs/main_docs_directory_tree.txt | skipped-generated |
| docs/make_085_rerender_provocations.sh | header-scan |
| docs/phase2_step3_insert_snippet.sql | header-scan |
| docs/provocation_card_loader.js | header-scan |
| docs/provocations.sample(1).json | family-delta |
| docs/provocations.sample(2).json | family-delta |
| docs/provocations.sample(3).json | full |
| docs/provocations.sample.json | family-delta |
| docs/provocations_archive_install.sql | header-scan |
| docs/provocations_archive_loader.js | header-scan |
| docs/summary.txt | header-scan |
| docs/thin-manifest.tsv | header-scan |

Family notes (audit trail): RUNNING_NOTES_vonc (1)–(35) and RUNNING_NOTES_vonc_v2 base–(27)
are byte-prefixes of their family-latest (verified with `cmp -n`), so they contain no
dropped content; RUNNING_NOTES_vonc.md (base) differs from (1) only in the migration-002/003
wording (the dropped "render_mode sweep" concept, captured below). RUNBOOK_phase2,
RUNBOOK_vonc_migrations, PLAN_* and HANDOFF families were edited in place; dropped-line
diffs were computed and the only substantive dropped concepts are captured below
(render_mode sweep; Option-1 static bake; Gap-2 sub-options (a)/(b); the pre-correction
"hand-patch rendered_html" trim method; `site_plan_directives` table mention). There is no
RUNBOOK_phase2(6) or RUNNING_NOTES_vonc(20) — numbering gaps, not missing content.

---

## Concepts

### Spark daily-provocation product (vonc.com)
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-09 §0: index + arena + archive live and browser-verified; but "the data file is currently hand-committed... a Phase-3 pipeline will emit it"; v1 roadmap features (daily_provocation_generation_from_scraping) not built.
- **what:** vonc.com / "Spark" — an AI daily-provocation platform: one charged provocation per day, users file a position, "the Gauntlet" scores the room, users get an Archetype. "The product IS the landing page": a single provocation card fills the screen; daily static regeneration; AI as producer (frames/scores/curates), not performer. v1 = daily provocations + Gauntlet; v3 concept = live challenge rooms. Serves as the platform's live test bed for the runtime-fill mechanism.
- **sources:** docs/PLAN_provocation-card(3).md#source-spec; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§0/§2; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~17:15; docs/RUNNING_NOTES_vonc_v2(28).md#carried-forward-state
- **relations:** runtime-fill mechanism; Phase-3 provocation data pipeline; provocation-card/lobby-grid/provocations-archive-list components
- **verify-later:** live vonc.com; sites row 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74; site_specs aspects (mission, roadmap, cta)

### Runtime-fill mechanism (data-runtime-fill shells + client loaders + JSON feed)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** HANDOFF §0: "That mechanism is now proven three times over" (provocation-card 2026-06-29, lobby-grid 2026-07-04, archive 2026-07-08, all browser-verified).
- **what:** Sections ship deliberately EMPTY at build time; the component `<section>` carries `data-runtime-fill="true"` so the assembler keeps the shell; an IIFE loader stored in `js_snippets` and bundled into `/assets/js/snippets.js` fetches `/data/provocations.json` in the visitor's browser and fills the shell's selectors, failing gracefully. Explicitly on-doctrine per doc 022 Tier 1 ("dynamic content injection... the dynamic part runs in the browser"; backend complexity lives in agents).
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2; docs/RUNBOOK_phase2_provocation_js(29).md#what-phase-2-is + #on-doctrine-check; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35; docs/README_summary_paragraph_for_handoff.md
- **relations:** visible-content filter exemption; js_snippets library; two JS delivery paths; static-vs-dynamic section distinction
- **verify-later:** rerender_single_page_action.go (reRuntimeFill regexp); js_snippets rows provocation-card-loader/lobby-grid-loader/provocations-archive-loader; /assets/js/snippets.js on vonc.com

### Two JS delivery paths (Path 1 component js_content vs Path 2 js_snippets bundle)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~17:20 (PD-1 answered): latest-news fetches via its extracted component JS (Path 1); Path-2 loader proven live the same day; 2026-07-07 side-evidence: extraction pattern live on three tool components.
- **what:** Two separate JS delivery mechanisms exist. Path 1: a component's inline `<script>` is extracted to `content_components.js_content` and deployed as `/tools/assets/{function}.js` automatically on every page rerender (how gauntlet-interface, latest-news, archetype-quiz ship JS — including news's data fetch). Path 2: library `js_snippets` rows are bundled to `/assets/js/snippets.js` by site-asset-renderer — NOT part of the normal build/rerender flow. The vonc daily-feed shells "fell between" the two paths. PD-3 decided Path 1 is the durable home for fetch-and-fill loaders; the Path-2 snippets remain the live working interim.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00 + #2026-06-29-~17:20; docs/RUNBOOK_phase2_provocation_js(29).md#path-decision + #framework-fix; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed
- **relations:** separateInlineJS extraction; js_snippets library; runtime-fill mechanism; js-bundle-stale gap
- **verify-later:** rerender_single_page_action.go collectJSAssets; content_components.js_content for gauntlet-interface/latest-news; /tools/assets/*.js in the sites repo

### js_snippets library + render_js_snippets_for_site + site-asset-renderer
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~17:20: renderer ran clean (commit eb7f2ac), snippet bundled; bundle header "3 active snippet(s)" 2026-07-07.
- **what:** `js_snippets` is a LIBRARY-WIDE table (no site_id) of JS behaviours keyed by `applies_to` (jsonb array of component functions). `render_js_snippets_for_site` selects active snippets whose applies_to overlaps the site's component functions, concatenates them (ordered by name, header comments, empty bundle still written so the head `<script src>` never 404s) into `/assets/js/snippets.js`, committed by the `site-asset-renderer` agent via git_commit. Loaders self-check for their section so a global snippet is inert on other sites. Pre-existing snippets were all small generic behaviours (accordion, scroll-reveal...); the fetch-and-fill loaders are a new, heavier use of the table.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 + #2026-06-25-~19:30 (inventory); docs/RUNBOOK_phase2_provocation_js(29).md#mechanism-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** two JS delivery paths; site-asset-renderer triggering gap; runtime-fill mechanism
- **verify-later:** render_js_snippets_for_site_action.go; site_asset_actions.go; js_snippets table contents

### js-bundle-stale gap (site-asset-renderer not wired into the build)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** FX-6 checkbox never ticked; RUNNING_NOTES 2026-06-29: "Gap 3 is NOT on the critical path... still a real latent issue for genuinely-generic snippets, but lower priority."
- **what:** Only `rerender-site` and `webdesign-agent` reference site-asset-renderer, so `/assets/js/snippets.js` is rebuilt at initial design and full site rerender but nothing re-runs it when a js_snippets row is added/changed later — the direct cause of the first loader never reaching vonc. Proposed fix: a design-discovery-agent check ("site has an applicable active snippet newer than its deployed bundle" → spawn site-asset-renderer). Deprioritised after PD-3 chose Path 1, never built; manual trigger scripts are the working practice.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#gap-3; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00 (GAP 3 CONFIRMED) + #2026-06-29-~17:20
- **relations:** design-discovery-agent named checks; two JS delivery paths
- **verify-later:** design-discovery-agent run_discovery_checks array; agent_definitions referencing render_js_snippets

### separateInlineJS inline-script extraction (+ collectJSAssets reader)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~19:30: "CODE IS CORRECT (not the bug)"; 2026-07-07: extraction pattern confirmed live on gauntlet-interface/latest-news/archetype-quiz (js_content + `<script src=` refs, no raw inline).
- **what:** On component store, `separateInlineJS` extracts bare `<script>` blocks (regex requires a closing tag; deliberately skips attributed tags — `src`, `type="application/ld+json"`, `type="module"` must stay inline) into `content_components.js_content`, replacing them with a `<script src="/tools/assets/{function}.js">` ref; multiple blocks are lazily matched and joined. `collectJSAssets` at rerender emits the per-component JS files. Known soft gaps: silent empty return on an unterminated `<script>` (warning proposed) and no log when an attributed script is left inline.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-06-29-~20:00; docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug
- **relations:** two JS delivery paths; store-path validation hardening; legacy un-extracted components
- **verify-later:** store_generated_component_action.go separateInlineJS (~line 105); rerender_single_page_action.go collectJSAssets

### Legacy un-extracted Mode-B shells (js-not-extracted class)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** Root-caused 2026-06-29 ("stored through a path that did NOT run separateInlineJS — most likely they predate its addition"); cosmetic-script extraction for provocation-card/lobby-grid still on the backlog 2026-07-09.
- **what:** provocation-card, lobby-grid (and brief-explanation) were stored via a pre-separateInlineJS path: raw inline script still in html_template, empty js_content, empty schema, `<no value>` placeholders — so `/tools/assets/{fn}.js` was never produced and their built-in interactivity never deployed. provocation-card's stored script was additionally truncated at generation (no `</script>`), which once shipped and swallowed the page footer. One creation-era bug with several surface symptoms (`js-not-extracted`, `mode-b-template`, section drops). Fix direction: regenerate through the current store path.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#extraction-bug-findings; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~19:30 + #2026-07-02-~19:35; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed (side-evidence)
- **relations:** Mode A/Mode B taxonomy; store-path validation hardening; separateInlineJS
- **verify-later:** content_components js_content/html_template for provocation-card 6163ff14 and lobby-grid 9304f14d (still raw inline?)

### Mode A / Mode B broken-template taxonomy + repair/regeneration routing
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session "Structural findings to carry forward"; code delivered 2026-06-22/23 (checkBrokenTemplateSlots, repair_template_slots); gauntlet-interface Mode-A repaired, archetype-result-card Mode-B regenerated to q100.
- **what:** Two distinct broken-template failure modes in the component library. Mode A: `<no value>FIELD</no>` — a render output stored as source with field names surviving as fallback text; repairable by string substitution (`repair_template_slots`). Mode B: bare `<no value>` — template rendered against an empty context and the cleaned output stored back; field names irretrievably lost; requires `needs_component_regeneration` → component-creator. `repair_template_slots` detects Mode B (no `</no>` tags) and returns needs_regeneration instead of attempting repair; `checkBrokenTemplateSlots` discovery check surfaces both.
- **sources:** docs/RUNBOOK_vonc_session(1).md#structural-findings; docs/RUNNING_NOTES_vonc(36).md#two-broken-template-failure-modes; docs/RUNBOOK_vonc_migrations(14).md#step-1
- **relations:** legacy un-extracted shells; store-path validation (rejects `<no value>` at the gate); component regeneration in place
- **verify-later:** check_component_standards.go; fix_component_template_action.go repairNoValueSlots

### render_mode derivation + LLM routing condition (migration 002)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Migration table 2026-06-24: "002 DONE — check_render_mode condition fixed"; deriveRenderMode code deployed 2026-06-24 ~15:00; hero LLM content confirmed on the rebuilt index.
- **what:** `StoreGeneratedComponentAction` originally hardcoded `render_mode='template'` on every component, making the LLM content path permanently unreachable; `deriveRenderMode(inputSchemaJSON)` now returns 'agent' iff any schema field has source='llm'. Separately, page-content-writer's `check_render_mode` condition was reading a never-populated field; migration 002 changed it to `current_section.llm_field_specs != null` (populated by plan_sections from the schema), routing any section with LLM fields to content generation for all sites. Note: render_mode is NOT what routes sections (a later red herring) — llm_field_specs is.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#4-render_mode-hardcoded + #migration-002-outcome + #2026-07-02-~19:20; docs/RUNBOOK_vonc_migrations(14).md#background-migration-002
- **relations:** render_mode sweep (dropped); plan_sections deferral (render_mode red herring)
- **verify-later:** store_generated_component_action.go deriveRenderMode; page-content-writer agent_definition check_render_mode condition

### Component-table render_mode sweep (65 components) — dropped migration
- **category:** NEW:page-build-pipeline
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_vonc.md base vs (1) diff: "Migration 002 (render_mode sweep across 65 components) is DROPPED"; PLAN_vonc_next_steps(1): "The 65-component render_mode update is DROPPED; existing components are fine as-is."
- **what:** The first plan for fixing LLM routing was a DB sweep updating `render_mode` on 65 existing library components. Dropped once it was established that workflow routing reads `llm_field_specs` (set by plan_sections from the schema), not the stored render_mode — so only the agent_definition condition needed fixing and existing component rows were fine as-is. Captures the earliest documented shape of the fix; useful provenance for why component rows still carry historical render_mode values.
- **sources:** docs/RUNNING_NOTES_vonc.md#4 (pre-edit base); docs/PLAN_vonc_next_steps(1).md#p1; docs/RUNBOOK_vonc_migrations(1).md (earlier "Fix render_mode on components" migration heading, dropped from later versions)
- **relations:** render_mode derivation + routing condition (the replacement)
- **verify-later:** none (historical)

### write_site_spec spec_data string coercion
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session: "FIXED, deployed"; migration table: code deployed 2026-06-24 ~15:00.
- **what:** `WriteSiteSpecAction` hard type-asserted `spec_data` to a map, rejecting the plain-string `mission_brief`/`roadmap_brief` the domain-submitter workflow resolves ("spec_data must be a JSON object, got string"). Fix: a coercion block — JSON string → parse; plain string → wrap as `{"text": value}` (matching the classifier prompt's `{{.site_specs.specs.mission_brief.text}}` read); objects pass through. The HANDOFF doc for this bug is also a worked example of the evidence-only handoff pattern (symptom carried, cause left to be read from code).
- **sources:** docs/RUNBOOK_vonc_session(1).md#1; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc(36).md#1
- **relations:** handoff document convention; data-shape/contract-drift debugging family
- **verify-later:** platform/orchestration/actions/site_spec_actions.go WriteSiteSpecAction coercion block

### Visible-content filter + data-runtime-fill assembler exemption
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_section_assembly_drop(3) RESULT 2026-07-03: "FIX VERIFIED... provocation-card RESTORED to the live page; lobby-grid correctly ABSENT"; carried-forward state: "DEPLOYED + verified".
- **what:** `rerender_single_page`'s getPageSections drops any section whose rendered_html has ≤10 chars of visible text after stripping style/script/tags/entities — correct for genuinely empty shells, wrong for intentionally-empty runtime-filled ones. PATCH_section_visible_content adds one regexp + early return: a section carrying `data-runtime-fill` is kept regardless of build-time text; unmarked sections filter exactly as before (so unbuilt shells like the then-empty lobby-grid stay correctly dropped). The investigation is also a model correction arc: the raw-inline-script hypothesis was proven WRONG by reading the action (script is stripped before measuring).
- **sources:** docs/RUNBOOK_section_assembly_drop(3).md#d4-result + #result-fix-verified; docs/PATCH_section_visible_content(1).go.txt; docs/RUNNING_NOTES_vonc(36).md#2026-07-03-~14:15
- **relations:** runtime-fill mechanism; assemble-only rerender; marker REPLACE anchoring
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent + reRuntimeFill

### plan_sections readiness triage and deferral semantics
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Confirmed in code 2026-07-02 (planSection read); fix validated end-to-end 2026-07-03 (index went 3 → 6 sections after populating cta spec + relaxing illustration field); 016b §9 entry.
- **what:** plan_sections classifies each planned section by resolving its schema fields: source=llm always available; query.*/renderer/static resolve at render time or fall back; any other source runs the resolver. A REQUIRED field whose source doesn't resolve hits the on_missing switch, whose `default:` case DEFERS the section ("default to defer for safety") — and empty on_missing defaults to skip_field, which is not a case in the required switch, so it defers. save_page_sections then persists only the ready set, dropping deferred sections' page instances. Authoring rule: never `required=true` + `on_missing=skip_field`; fix by populating the site data source or degrading the field.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:20 + #2026-07-02-~19:35; docs/016b_debugging_guide_merged(3).md#regenerated-content-section-is-deferred; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f
- **relations:** site_specs cta aspect; resolver asset kinds gap; plan-driven rebuild + clobber
- **verify-later:** plan_sections_action.go planSection on_missing switch; save_page_sections_action.go

### Plan-driven rebuild + interactive/deferred-section clobber (carry-forward fix)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** 016b: "Part 4... fix WRITTEN (un-deployed)" — Layer 1 interactivity guard + Layer 2 carry-forward in patched save_page_sections_action.go; the 2026-07-02 vonc rebuild demonstrated the drop live (6 planned → 3 saved, brief-explanation instance gone).
- **what:** A needs_page rebuild is PLAN-driven, not pending-driven: load sections from the plan → triage → the writer renders ALL ready planned sections → save_page_sections DELETE+INSERTs the page's components. Sections present in page_components but absent from the plan (interactive tools stored only as rendered_html) or deferred by triage get silently dropped. Fix (written, not deployed): interactivity-aware guard blocking a non-interactive set replacing a deployed interactive one, plus carry-forward of existing interactive sections; three callers to bump (page-build-handler, page-rerender, tool-recreation-handler); plus source_item_id stamping for traceability.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + 2026-06-24 update); docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:40 + #2026-07-02-~19:00; docs/PLAN_spark_provocation_pipeline.md#standing-constraints
- **relations:** plan_sections deferral; page_components single-writer (save_page_sections); interactive tool pages stored as rendered_html
- **verify-later:** save_page_sections_action.go (is the guard/carry-forward deployed?); page_component_history.source_item_id

### Three section sources for a page build (aspect → pages.sections → plan tables)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Workflow dump + code read 2026-07-06: "load_spec_sections... reads site_specs aspect site_plan (AUTHORITATIVE) → fallback page_record.sections. The site_plan_sections TABLE is NOT read by this path."
- **what:** Page builds resolve their section list from, in order: the `site_specs` aspect `site_plan` (legacy blob, 5 sites carry one; vonc has none), `pages.sections` (jsonb fallback — what actually serves vonc; the newer planner dual-writes plan tables → pages.sections), and same-role sibling synthesis; the `site_plan_sections` table is written by the vonc-generation planner but not read by the build path. Three peer stores with unclear precedence caused ten silent no-op builds and two fixes landing in the wrong store (a plan-table row, then the pages.sections UPDATE that finally unblocked).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3; docs/RUNBOOK_phase2_provocation_js(29).md#update-2026-07-06
- **relations:** plan storage authority (029 Q1); complete_error silent no-ops; load_page_record lookup semantics
- **verify-later:** load_page_sections_from_spec_action.go source order; SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'

### complete_error silent-success family (page build completes having built nothing)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Mechanism fully confirmed 2026-07-06 (workflow dump); the defect is live ("an error path implemented as a SUCCESSFUL completion"); preventions listed as backlog item 1 in the HANDOFF, not built.
- **what:** page-build-handler routes zero-ready-sections to a step literally named `complete_error` — a complete_workflow with success_message "Content writer skipped — page has no sections defined" — so builds against a section-less page complete cleanly having done nothing. Ten silent no-ops (33–65s completes) hid a 404 CTA destination for two weeks; a work-item result carrying ONLY `site_record` (healthy: `[sections_saved, deploy_result]`) is the diagnostic signature. Variants: plan row naming a nonexistent component also passes silently. Preventions (aspirational): complete_error fails loudly or raises needs_plan_sections; auditor linked+planned+URL-presence rules; `last_built_at` is never written by build or rerender (dead column — write it or drop it).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed + #2026-07-08; docs/016b_debugging_guide_merged(3).md#page-build-completes-having-built-nothing; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§9
- **relations:** three section sources; planner ≥1-section invariant; trust-the-artifact doctrine
- **verify-later:** page-build-handler default_config complete_error step; pages.last_built_at writes anywhere in the codebase

### load_page_record lookup semantics (name-first, page_id fallback)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** load_page_record_action.go read in full 2026-07-06: "Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames)... returns sections PARSED FROM pages.sections + section_count."
- **what:** The build's page lookup resolves by page_name against `pages.name` first, falling back to page_id only for empty/bogus names, and returns the page's own `sections` jsonb with a count — which is what gates the zero-sections branch. Schema gotcha bundled with it: `pages` has `name` not `page_name`; work-item specs use domain/page_id/filename/page_name.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-load_page_record-read; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** complete_error family; three section sources
- **verify-later:** load_page_record_action.go nonPageNames list

### Plan storage authority — 029 Q1 and the withdrawn table-first alteration
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** PLAN_dynamic_sections(4): "SUPERSEDED (2026-07-06, same day) — decision deferred to 029 Q1; alteration WITHDRAWN"; "Decision closed (2026-07-07): the user chose REVERT."
- **what:** After the silent no-ops, a decision was made (then withdrawn the same day) to make the `site_plans` family the authoritative plan store and alter `load_page_sections_from_spec` to read site_plan_sections first. Reading design doc 029 showed plan storage is its OPEN Q1 ("site_specs aspects vs new table", lean = partitioned site_plan_* aspects + a reconcile_site_plan action); three shapes coexist in production (legacy site_plan blob aspect ×5 sites; 029 partitioned aspects apparently unimplemented; the vonc-generation tables with pages.sections dual-write). The alteration was withdrawn and the repo file reverted (ORIGINAL.go; cluster reverts on next chassis push); evidence contributed to Q1: the table path now exists in production post-dating the lean. Store-agnostic preventions retained. Earlier draft (v2 of the plan) also named a `site_plan_directives` child table not mentioned in the final version.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#decision + #superseded; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-alteration-withdrawn + #2026-07-07-revert-decision; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3
- **relations:** three section sources; reconcile_site_plan (029); planner ≥1-section invariant
- **verify-later:** git history of load_page_sections_from_spec_action.go (reverted?); repo grep reconcile_site_plan; docs024 029 doc Q1 status

### Planner role-aware ≥1-section invariant + role→pipeline mapping
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** Backlog item 1 in HANDOFF §9; "Invariant refined: every planned page whose ROLE is built by page-build-handler must have ≥1 section" (Gate B, 2026-07-06) — nowhere claimed built.
- **what:** The June planner emitted all 8 vonc pages but skipped SECTIONS for exactly the two non-standard roles — blog-post (legitimate: the blog pipeline builds those) and section-index (the defect that caused the archive 404). Prevention: at plan-store time, every planned page whose role page-build-handler owns must have ≥1 section, with the role→pipeline mapping made explicit; plus auditor drift rule (pages.sections vs current plan) and post-deploy URL-presence checks per active page.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#gate-results; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-i-gate-results
- **relations:** complete_error family; section descriptor design; quality-auditor rules
- **verify-later:** site-planner agent_definition; site_plan_pages roles for recent sites

### Autonomous section composition — per-section descriptor {role, kind, data_feed}
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** PLAN_dynamic_sections_and_loaders(4) status "DESIGN"; gaps list "(1) Section descriptor... Without this the framework can't tell static from dynamic" — none of gaps 1–5 marked built.
- **what:** The framework (not a human) should decide, from the domain/site-spec, which sections a page has, each section's role (to prevent overlaps like provocation-card's mini-lobby vs lobby-grid), whether it is static (build-time content) or dynamic (runtime-filled from a feed), and which named feed — encoded as a per-section descriptor `{component_name, role, kind, data_feed}` on the plan, written by the site-planner, consumed by build AND maintenance flows. The plan not carrying `kind` is why the assembler dropped the runtime-filled shells. Includes a spec-level feed catalogue and quality-auditor maintenance detections (dropped-dynamic, overlap, deferral, empty-dynamic). The root design point: a data-driven component should DECLARE its runtime data dependency so the pipeline provisions feed + loader automatically.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#the-question + #structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed; docs/RUNBOOK_phase2_provocation_js(29).md#how-a-component-should-declare
- **relations:** Tier E runtime-feed tier; loader-builder agent; static-vs-dynamic distinction; plan storage authority (where the descriptor lives follows 029 Q1)
- **verify-later:** site_plan_sections columns (kind/data_feed/role exist?); site-planner prompt/workflow

### Component field-source tiers (A/B/C/D + renderer) and proposed Tier E runtime-feed
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** Tiers A–D + renderer confirmed from the component-creator prompt at v1.0.1080 (deployed); Tier E is "proposed, pending decision" (2026-06-29) and gap 2 of the autonomy plan (not built).
- **what:** component-creator's schema contract sources each field from Tier A (voice/llm), B (tunable labels/static+fallback), C (site data — site_specs/site_assets), D (derived lists, query.* resolved at plan time), plus a "renderer" source (JS-filled single value with fallback). There is NO tier for content fetched client-side from a JSON feed at runtime — so regenerating a daily-feed component as-is would wrongly bake a build-time provocation into the template. Proposed Tier E ("feed.{name}"): emit a stable-selector DOM shell + declared DOM contract + (originally) an inline loader following a canonical pattern; the archive build refined this to marker-at-generation + external loader.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 (GAP 2 CONFIRMED); docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/RUNBOOK_phase2_provocation_js(29).md#gap-2
- **relations:** section descriptor; generation-time guards; loader-builder agent. Dropped earlier framing: Gap-2 sub-options "(a) component-creator emits a companion js_snippet" vs "(b) loader snippets are library fixtures" (early runbook versions) — superseded by the Tier-E + loader-builder design.
- **verify-later:** component-creator prompt_template tier section; any feed.* source in input_schemas

### loader-builder agent (fetch-and-fill sibling of tool-generator)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** PLAN_dynamic_sections gap 3; "the missing piece (provocation-card/lobby-grid loaders were hand-built)" (2026-07-04); the two hand-built loaders named as its reference implementations.
- **what:** A proposed agent that LLM-generates client-side fetch-and-fill loaders for dynamic sections: input = the section's DOM contract + feed shape; output = a graceful IIFE installed as a js_snippet and bundled by site-asset-renderer. Modelled on tool-generator (which LLM-generates, saves and wires SELF-CONTAINED tools) but necessarily a SIBLING because tool-generator explicitly forbids fetch. The framework currently has component-creator (section templates) and tool-generator (tools) but no runtime-fill loader builder.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** tool-generator (tool-pipeline); Tier E; provocation_card_loader.js / lobby_grid_loader.js / provocations_archive_loader.js as references
- **verify-later:** agent_definitions for tool-generator (no-fetch rule); absence of any loader-builder agent

### Phase-3 provocation data pipeline (provocation-generator + orchestrator + render action + daily schedule)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** Phase-1 diagnostics confirmed "a clean slate — nothing exists yet" (2026-06-25); FX-4 checkbox never ticked; provocations.json still hand-committed as of 2026-07-09.
- **what:** The pipeline that would generate `/data/provocations.json` daily: clone the news pipeline — seed content_sources (trending-topic scraping targets) → reuse feed-ingester → NEW provocation-generator agent (LLM: raw topics → provocations + AI takes; generative analogue of feed-triage) → NEW render_provocations_section Go action (mirror of render_news_section; Go struct defines the JSON shape; returns a files map for git_commit) → provocation-orchestrator (clone of content-feed-orchestrator) → scheduled_tasks row `provocation-refresh` (daily; the column is `name`, not task_name). Open questions recorded: sources, volume per day, archive-page reads.
- **sources:** docs/PLAN_spark_provocation_pipeline.md; docs/RUNBOOK_phase2_provocation_js(29).md#data-deploy + #gap-1; docs/RUNBOOK_vonc_migrations(14).md#step-8
- **relations:** news feed pipeline (the template); provocations.json contract (the target shape); Spark product
- **verify-later:** absence of provocation-* agent_definitions/scheduled_tasks/content_sources for vonc; render_news_section_action.go as the model

### News feed pipeline as the proven data-layer template
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-25 ~18:00: "all confirmed active v1.0.1078" — read from agent definitions and action source.
- **what:** content-feed-trigger (scheduled heartbeat 6h via scheduled_tasks name='content-feed-refresh'; finds news-recommended sites, spawns content-feed-orchestrator per site, max 5) → orchestrator (seed_content_sources → dispatch_feed_sources → feed-ingester per due source [rss/scrape/news_search/api_news] → feed-triage LLM relevance+credibility scoring) → render_news_section (loads items, expires stale, builds JSON from a Go struct, produces an archive JSON if a news-index page exists) → git_commit `/data/latest-news.json`. The latest-news component fetches the JSON via its own extracted component JS (Path 1); the news-date-formatter snippet is only a helper. This is the platform's model for any static-site runtime data feed.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 + #2026-06-29-~17:20 (PD-1); docs/PLAN_spark_provocation_pipeline.md#architecture
- **relations:** Phase-3 provocation pipeline (clone); scheduler-and-tasks (scheduled_tasks.name)
- **verify-later:** render_news_section_action.go; content-feed-* agent_definitions; scheduled_tasks row content-feed-refresh

### provocations.json data contract (today / lobby / arena / archive)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** v3 served live (curl grep '"archive"' = 1, 2026-07-07); all three loaders verified filling from it in a browser.
- **what:** The versioned feed contract for Spark's runtime-fill sections: `generated_at`; `today` {eyebrow, headline (may carry `<em>`), body, primary_cta/secondary_cta {label,url}, stats ×3}; `lobby` [4 × {icon,title,desc,url}] (becomes dead after the mini-lobby trim); `arena` — an OBJECT {eyebrow,title,subtitle,cta_label,cta,cards[≤6]} because the grid's header + CTA need data too; `archive` {entries[≤24] {date,title,teaser,stat,url}, newest-first}. Evolved v1→v3 in provocations.sample.json; hand-committed interim, the fixed generation target for Phase 3.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35 (confirmed-good shape); docs/PLAN_lobby-grid(2).md#build-progress; docs/SPEC_provocations-archive-list.md#data-contract; docs/provocations.sample(3).json
- **relations:** the three loaders; Phase-3 pipeline; runtime-fill mechanism
- **verify-later:** live vonc.com/data/provocations.json keys

### provocation-card component (daily hero card) + mini-lobby trim
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** "Live and working via Path-2 loader" (PLAN status); trim CONFIRMED 2026-07-04, drafted 2026-07-09, blocked on the bundle verdict — not executed within this corpus.
- **what:** The Spark centrepiece: single daily contested claim + AI take + 3 stats + 2 CTAs + (currently) a 4-card mini-lobby, filled at runtime from `today`/`lobby` by provocation-card-loader against the `.pc-*` DOM contract. JS-required by design — do NOT "fix" by baking content. Known limitation: the underlying template is Mode-B broken (loader masks it; JS-off shows `<no value>`). NEXT TASK: trim the mini-lobby (template pc-card-grid block, loader lobby fill, the orphaned 1fr-1fr media query, the dead hover script) because lobby-grid owns the arena role — with the method itself under a bundle verdict since HTML patching is the rejected mechanism.
- **sources:** docs/PLAN_provocation-card(3).md; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:20; docs/provocation_card_loader.js (header)
- **relations:** lobby-grid overlap decision; sanctioned edit paths; runtime-fill mechanism
- **verify-later:** content_components 6163ff14 html_template (pc-card-grid still present?); js_snippets provocation-card-loader lobby block

### lobby-grid arena component (six-room grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "lobby-grid DONE (browser-verified)" 2026-07-04 — six arena cards + pulsing stat dots + "Enter the Arena" live; PLAN_lobby-grid marked DELIVERED 2026-07-09.
- **what:** The Arena lobby: 6-card grid (1 featured spanning 2 cols, 4 standard, 1 wide), each card icon (SVG inner markup with emoji fallback)/tag/title/desc/stat + pulsing dot, plus header and CTA — filled at runtime from `arena` by lobby-grid-loader. Honest v1 semantics decision: "live rooms" is a v3 concept, so in v1 the grid shows TODAY'S PROVOCATIONS as enterable cards. Confirmed decisions: lobby-grid is the primary "today's provocations grid" (D-A) with the `arena` object as feed (D-B). Its build was deliberately the reference implementation for the loader-builder design.
- **sources:** docs/PLAN_lobby-grid(2).md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-lobby-grid-verified; docs/lobby_grid_loader.js (header); docs/lobby_grid_install.sql (header)
- **relations:** provocation-card mini-lobby trim; loader-builder reference; marker REPLACE anchoring incident
- **verify-later:** js_snippets lobby-grid-loader; live index data-component="lobby-grid"

### brief-explanation static explainer (regeneration, not a loader)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083 succeeded 2026-07-01 (in-place update, quality 50→100, 0→20 fields); rendered with real copy on the live index 2026-07-03.
- **what:** The "what is Spark / how it works" index explainer — STABLE brand content (eyebrow, heading with `<mark>`, description, exactly 3 numbered steps, 3 stats, 2 CTAs, illustration+badge) that belongs in build-time HTML for SEO and no-JS robustness. Establishes the key distinction: Option-2 runtime loaders are ONLY for daily-changing data shells; static shells that happen to be empty are fixed by REGENERATION with a real schema — two different resolutions for the same empty-shell symptom. Its stat fields were later re-sourced static→llm to stop generic SaaS fallbacks leaking.
- **sources:** docs/PLAN_brief-explanation(1).md; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~18:00 + #2026-07-01-~12:46; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-static-field-discrepancy
- **relations:** static-vs-dynamic distinction; shared component library (58363894 shared ×3 sites); component regeneration in place
- **verify-later:** content_components 58363894 field sources; idea.uk/robot-hands pending instances

### provocations-archive-list component + provocations archive page
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "PROVOCATIONS-INDEX THREAD DONE" 2026-07-08: page live, 8 rows fill, ghost row eliminated; live confirm grep = 2 on 2026-07-09.
- **what:** The Provocations Archive at /provocations/index.html — destination of every primary CTA — as a single self-contained runtime-fill section: llm header fields (nothing can defer), a hidden clone-template row the loader clones per `archive.entries[]` (variable-length list vs lobby-grid's fixed six), a visible empty state so the page ships before data lands, CTA back to today. Built via the full arc: component (70d6662a, 084 trigger) → plan row → pages.sections unblock → first real build (~5 min after ten 33–65s no-ops) → loader + data → ghost-row CSS fix.
- **sources:** docs/SPEC_provocations-archive-list.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/RUNBOOK_phase2_provocation_js(29).md#you-are-here; docs/provocations_archive_loader.js (header)
- **relations:** complete_error family (its 404 was the trigger); generation-time guards (first live validation); CTA graph
- **verify-later:** pages e4b3b195 build_status; live /provocations/index.html

### Generation-time guards for dynamic components
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** 084 result 2026-07-06: "FIRST LIVE VALIDATION of baking the guards in at generation" — has_marker=t, has_inline_script=f on the created component; guards held through the real pipeline end to end.
- **what:** Lessons from the whole thread baked into component GENERATION instead of post-hoc surgery: emit `data-runtime-fill` in the template's section tag at generation (no string-REPLACE marker step); forbid inline `<script>` entirely in dynamic components (extraction-bug class becomes impossible; behaviour lives in the external loader); make header copy llm-sourced (no deferral risk); list entries pure markup (nothing for the resolver to fail on); hidden clone-template item plus a `[data-…-template]{display:none}` author rule (hidden alone loses to author CSS); visible empty state. Declared "the pattern for all future dynamic sections".
- **sources:** docs/SPEC_provocations-archive-list.md#design-decisions; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-084-succeeded; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** Tier E proposal; marker anchoring lesson; hidden-vs-author-CSS lesson; store-path validation
- **verify-later:** component-creator output for any newer dynamic component (marker present at generation?)

### Shared component library semantics + field-set guard + neutral-base/fork rule
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Guard read in code (store_generated_component L315-335, referencing the fdd92ad4 incident); shared-component clobber check VERDICT 2026-07-04: "no contamination; brief-explanation base is neutral"; rule recorded as standing.
- **what:** Components with `forked_from IS NULL` are a cross-site SHARED library keyed by `function` (brief-explanation is shared by vonc + idea.uk + robot-hands). A deliberate guard blocks regenerations that DROP or RENAME existing fields on a shared component (the fdd92ad4 system-stats incident: an in-place field rename silently emptied every dependent); pure field ADDITION passes. Standing rule derived: regenerate a shared base only for neutral, purely-additive improvements; site-specific voice must FORK (`forked_from = base_id`) — the "deliberate migration" the code prescribes; direct SQL UPDATEs bypass both the guard and component_versions snapshots. An optional multi-site regen gate (`allow_shared_base_regen`) was considered and HELD.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-shared-component-clobber-check + #2026-07-04-verdict + #2026-07-04-lobby-grid-verified (store analysis); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** component regeneration in place; content-governance (voice leakage); section descriptor (fork-vs-base per site should live on the plan)
- **verify-later:** store_generated_component_action.go field-set guard; content_components forked_from usage

### Component regeneration in place (store_generated_component mechanics)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 083 result: brief-explanation updated in place (same id, created_at unchanged, status 'regenerated', component_versions snapshot, needs_rerender raised) — "matches the documented behaviour (003 §348)".
- **what:** store_generated_component looks up an existing component by the LLM's EMITTED `function` (forked_from IS NULL); if found, it snapshots the current row to component_versions (MAX+1), UPDATEs in place (component_id preserved → all page/site FKs keep resolving), sets template/schema/js_content/render_mode/is_active, then markPagesPendingRebuild raises ONE needs_rerender per affected site. Determinism hazard: regeneration keys on the emitted function name — an unpinned LLM can emit a different name and INSERT a stray duplicate (the 081 'general-hero' incident); pin the function in the description. Pre-store validation rejects `<no value>` templates and checks placeholder/schema parity.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-06-30-~18:35 + #2026-07-01-~12:46; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e
- **relations:** shared library guard; component-quality-auditor; call_agent contract validation (the trigger saga)
- **verify-later:** store_generated_component_action.go lookup + snapshot + markPagesPendingRebuild; component_versions rows

### component-quality-auditor auto-regeneration threshold
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Read from its default_config 2026-06-29: creates needs_component_regeneration items only for quality_score < 50, handler component-creator, spec {function, component_id, quality_score, quality_issues}.
- **what:** The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant the three vonc shells scoring EXACTLY 50 were never auto-picked-up (explaining zero queued items and requiring manual triggers). Its item shape confirms the designed regen path keys on function and routes to component-creator. Boundary-condition gap worth a rule review; also the future home of the autonomy plan's maintenance detections.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~21:00; docs/PLAN_dynamic_sections_and_loaders(4).md#maintenance
- **relations:** component regeneration in place; autonomous section composition (auditor rules gap 4)
- **verify-later:** component-quality-auditor default_config condition; quality_score distribution at exactly 50

### call_agent contract validation vs input_data.spec convention (dual placement; validator patch)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** Mechanism confirmed in code 2026-07-01 (ValidateInputContract checks only top level); PATCH_validate_input_contract.go "WRITTEN, deploy PENDING" (carried-forward state; still backlog item 3 on 2026-07-09).
- **what:** call_agent resolves input_mapping then validates the target's input_contract.required against TOP-LEVEL keys, while handler workflows read spec fields at `input_data.spec.*` (the work-item convention). The two read different places, so component-creator (required: section_type) can be satisfied neither by pure-top-level (empty-context generic generation — the 081 stray) nor pure-nested (contract violation — 082); the working manual shape (083) provides section_type BOTH top-level AND inside spec. build-dispatch-loop's generic mapping flattens no section_type, so the designed work-item path would hit the same violation (predicted, unconfirmed). Framework fix: the validator accepts a required field top-level OR at input_data.spec.X — not per-handler loop mappings, not enshrining the duplication.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~10:10 + #2026-07-01-~12:46 + #2026-07-01-~13:10; docs/PATCH_validate_input_contract.go.txt; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e; docs/083_regenerate_brief-explanation_vonc.sh (header)
- **relations:** manual agent trigger pattern; build-dispatch-loop genericity (002 §414); component regeneration
- **verify-later:** input_mapping.go ValidateInputContract (patched?); a needs_component_regeneration item dispatched through the loop

### Manual agent trigger via the generic entry point (spawn+call pattern)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Proven repeatedly: 083 (2026-07-01), 084 ("the dual-placement trigger pattern worked again", 2026-07-06), trigger-asset-renderer and rerender scripts.
- **what:** One-off manual agent runs post a spawn_agent+call_agent message to `system.agent.generic.requests` (kcat with correlation/request/client headers), with input_mapping delivering the payload. Hard-won sub-rules: dual placement of contract-required fields; a QUOTE-FREE description (name attribute values in prose) to survive the kcat/JSON escaping pipeline; JSON embedded literally (no jq dependency); watch via orchestration_states by correlation_id. The numbered trigger-script series (080–085) in scripts/initial_messages/210_vonc_trigger/ is the operational library, including make_085 which sed-copies a proven trigger for a new page (reuse-first).
- **sources:** docs/084_create_provocations-archive-list_vonc.sh (header); docs/make_085_rerender_provocations.sh; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6 + #§8; docs/summary.txt (kcat basics)
- **relations:** call_agent contract validation; work-item conventions
- **verify-later:** scripts/initial_messages/210_vonc_trigger/ contents

### Store-path template validation (+ pending <script>-balance hardening)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Existing checks confirmed in code 2026-06-29 (`<no value>` rejection, placeholder/schema parity, unclosed `<style>`, section/div presence); the `<script>` balance check + separateInlineJS truncation warning remain "STILL MISSING" / backlog item 2 on 2026-07-09.
- **what:** store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT an unclosed `<script>` — the gap that let provocation-card ship a truncated inline script that swallowed the page footer at render. Hardening definition: add a `<script>` open/close balance check (reject or flag-for-regeneration) plus a truncation warning in separateInlineJS. Prevents the class "truncated template ships and breaks the page".
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-07-03-~13:25 (hardening def); docs/RUNBOOK_phase2_provocation_js(29).md#appendix-g; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** legacy un-extracted shells (the truncation instance); Mode A/B taxonomy
- **verify-later:** store_generated_component_action.go validation block for script balance

### CSS variable naming convention (--color-*) + creator prompt STRICT RULE
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Hero template fixed and verified (magenta CTA, dark bg) 2026-06-24/25; component-creator prompt patched with "USE ONLY THESE NAMES" + STRICT RULE, UPDATE confirmed in DB 2026-06-24 ~16:50; library-wide audit complete.
- **what:** System CSS custom properties follow `--color-primary/-secondary/-accent/-background/...` naming; LLM-generated components had emitted `--primary-color`-style names that don't exist in styles.css, so fallback hexes fired (the "brochure-blue" index). Fix: template REPLACE on hero + a component-creator prompt section explicitly prohibiting the wrong names and separating Palette from Layout tokens. Documented exception: `--archetype-color` is intentional per-card tinting with `--color-accent` fallback. All new components inherit the correct names.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:30 + #2026-06-24-~16:50; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** post-025 CSS theme flow; legacy-variable "fossilised render" tell (016b chrome entry)
- **verify-later:** component-creator prompt_template section 7; grep new templates for --primary-color

### Post-025 CSS theme flow (empty css_content by design; composition via FK chain)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-24 ~16:00 citing doc 027 and install_site_composition_action.go L210-212: css_content "intentionally empty — post-025 renderer reads composition via FK chain at render time"; styles.css deployed by webdesign-agent.
- **what:** The design pipeline runs needs_composition (site-design-planner) → gated needs_design (webdesign-agent: analyze_design → update_site → generate_css via render_css_from_spec reading composition FKs → deploy_css writes assets/css/styles.css → optional fork_theme). `css_themes.css_content` is intentionally empty post-025; the empty "Theme-specific styles injected here" head block is expected, not a bug. webdesign-agent is not deprecated. Key debugging consequence: a wrong colour on a page is more likely a component variable-name mismatch than a theme-injection failure.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:00; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** CSS variable naming; two chrome assembly paths (stale renders)
- **verify-later:** install_site_composition_action.go; render_css_from_spec

### Resolver asset-kind surfacing gap (hero/logo only; illustrations unreachable)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Confirmed 2026-07-02: illustration assets EXIST (illustration_game_master, illustration_gauntlet_cta, purpose=illustration, active, files deployed) but "resolver ensureAssets only surfaces hero/logo, so site_assets.illustration can't reach them"; workaround applied (field made optional); extension still backlog item 4 on 2026-07-09.
- **what:** The plan_sections resolver's ensureAssets populates only `hero` and `logo` asset keys (from site_plan_imagery kinds hero/logo), so any schema field sourced `site_assets.illustration` can never resolve even when illustration assets exist — deferring sections. Interim: make such fields optional (text-only render). Structural options: extend ensureAssets to surface kind=illustration from site_plan_imagery+assets (benefits all sites), or per-field fallback URLs. Related mismatch: gauntlet-cta has no illustration field despite an illustration_gauntlet_cta asset existing.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:50 + #2026-07-03-~13:18; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; site_plan_imagery (imagery pipeline)
- **verify-later:** plan_sections resolver ensureAssets; site_plan_imagery kinds in use

### site_specs `cta` aspect + CTA graph audit (parked)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** cta aspect inserted 2026-07-02 (primary_url/secondary_url) — un-deferred two sections; CTA-map pass explicitly PARKED (user chose Option B 2026-07-07: leave the circular graph until the real arena exists).
- **what:** A per-site `site_specs` aspect `cta` supplies shared CTA URLs (`cta.primary_url`, `cta.secondary_url`) resolved into component fields (gauntlet-cta.cta_primary_url, system-stats.cta_url) — one populated source fixes all dependants. The vonc CTA graph was then found CIRCULAR (hero→archive, archive→home, gauntlet-cta→archive; only nav/footer reach the Gauntlet tool, and no arena page exists); a deliberate CTA-map pass is queued because CTA URLs are baked into rendered sections, so a proper refresh is a section rebuild, not string surgery.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:35 + #2026-07-02-~19:50; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-step-4-done + #2026-07-07-~16:30; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; phantom CTA bug; unresolved_cta work items (self-resolve when hubs exist)
- **verify-later:** site_specs aspect='cta' rows; retarget SQL parked in notes

### Phantom CTA resolution bug (fabricated /{area}.html hero CTAs)
- **category:** link-management
- **status-signal:** partial
- **status-evidence:** 016b Part 4 (confirmed 2026-06-22 in deployed HTML, gamesdesign): hero carries two phantom CTAs from schema sources pages.contact/pages.services; "workflow-only fix staged" (select_sections reading resolved_links at the wrong path).
- **what:** Hero CTA resolution can produce constructed/fabricated URLs (`/contact.html`, `/services.html`) while the real hubs live elsewhere, because `select_sections` reads `resolved_links.sections_ready` (null) instead of `resolved_links.response.link_resolution.sections_ready`, falling back to the un-augmented plan; `resolve_internal_links` is a build-time augmenter (writes cta_url into resolved_data for the writer), explicitly not a rendered-HTML patcher, and `check_phantom_internal_links` routes page-link fixes to page-build-handler by design. Distinct from the interactive clobber; `page_rerender` does not re-resolve schema-sourced CTAs (ruled out as a link fix).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + update); docs/RUNBOOK_vonc_session(1).md#remaining-steps (unresolved_cta parking)
- **relations:** site_specs cta aspect; internal link management (024)
- **verify-later:** select_sections workflow path fix; resolve_internal_links action

### Trust-the-artifact debugging doctrine (silent-success family + verification discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b "Durable invariants" section; proven repeatedly (ten complete-with-nothing items; "after ten silent no-ops... NOT trusting 'complete' without artifacts").
- **what:** The unit's core debugging doctrine: a `complete` work item or green commit proves nothing — verify by artifact (DB row, curl, browser); completed_at is orchestration END, not the write instant (trace child orchestrations); a config key read on a different path than it is set is a silent no-op (compare producer output to consumer read by exact path); 0 rows is not decisive until the query is cleared (wrong column/id/schema/window); a negative inference from an artifact's shape needs the mechanism checked in all cases (the separateInlineJS attribute-skip example); pod logs are ephemeral across rollouts (grep zap by message + JSON field, never 'field=value'; agent_error_log outlives pods); copy full UUIDs, never hand-type; ±6-byte js_len paste drift is cosmetic — bundle and browser are ground truth; dated backup tables per change (never reuse an IF-NOT-EXISTS backup name); only save_page_sections writes page_components.
- **sources:** docs/016b_debugging_guide_merged(3).md#durable-invariants; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§8; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-03-section-drop-closed
- **relations:** complete_error family; zap/pod-log entry; SQL surgery pattern
- **verify-later:** n/a (doctrine); stage 2 can test individual heuristics against code

### Travelling per-tool documentation convention (PLAN_/NOTES_ per component)
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** Convention written 2026-06-29 (TOOL_DOCS_convention.md); instantiated for provocation-card, brief-explanation, lobby-grid, provocations-index, provocations-archive-list; DB layer explicitly deferred ("files now, hybrid later"); pipeline integration is a spec'd future feature.
- **what:** Every tool/component carries its own reasoning history: PLAN_<function>.md (aim, source spec, behaviour + data/DOM contract, delivery mechanism, dependencies, deliberate decisions) and NOTES_<function>.md (timestamped entries with `Categories:` tags: choices, bugs symptom→cause→fix→verify, dead ends). Problem-category taxonomy (css-variable-mismatch, mode-b-template, js-not-extracted, js-bundle-stale, schema-template-drift...) rolls up into the global debugging guide. Storage decision: git now, HYBRID later (NOTES → a tool_doc_notes table when agents start writing them; PLAN stays in git — never an unversioned DB text column). Future: tool-generator writes the PLAN at creation, capturing LLM design reasoning currently discarded; bug entry-points load PLAN+NOTES first. Global guide = cross-tool patterns; site runbook = one site; per-tool docs = one tool across sites.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~18:10 + #2026-06-29-~18:45; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-tool-docs; docs/PLAN_provocation-card(3).md (worked exemplar)
- **relations:** debugging-guide fork/merge; handoff convention; docs026 itself is a consumer
- **verify-later:** TOOL_DOCS_convention.md location; existence of tool_docs/tool_doc_notes tables (expect none)

### Debugging-guide fork-and-merge maintenance (cumulative 016b copy)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** merged v5 changelog line 2026-07-04: "Guide had forked across chats; this is the cumulative version"; HANDOFF §9 item 10: "Apply 016b_debugging_guide_merged.md to the project."
- **what:** The 016b debugging guide is maintained across parallel chat threads and FORKS: the project copy and thread working copies diverge (each gaining entries the other lacks). Practice: merge into a cumulative copy (v5 folded three vonc-thread entries + the silent-noop entry into the parallel-chat version), version-stamp the changelog, and apply the merged copy back to the project. The docs/ root copies (guide_6_, merged(2)/(3)) are these thread artifacts; their unique-to-thread entries are the deferral drop, artifact-not-pod-logs, marker anchoring, silent no-op, hidden-vs-author-CSS, plus parallel-chat entries (two chrome paths, SQL pitfalls, sites.status).
- **sources:** docs/016b_debugging_guide_merged(3).md#v5-changelog; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-guide-merged; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** travelling docs convention; category tag roll-up
- **verify-later:** which 016b copy the docs024 consolidated guide corresponds to; whether the merge was applied

### Handoff document convention (stand-alone dated brief for a fresh chat)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-07-09 written to the convention and used (§1 first actions ... §10 file inventory; `Categories:` tags); the write_site_spec handoff is the evidence-only variant.
- **what:** A dated, self-contained handoff document lets a fresh chat (memory off) start work: orientation, verified DONE state with copy-paste ids, the next task's full scope/method/acceptance, data to collect first, commands/triggers, schema notes, gotchas (each "paid for" in the thread), backlog, file inventory. The diagnostic variant carries EVIDENCE and context but deliberately NO diagnosis ("the cause is still to be read from the real code"). Related authoring hygiene rules: no bare angle-bracket tags in markdown prose (breaks readers — the same failure mode as the live page bug); quote heredoc delimiters; /home/claude resets between sessions while outputs persists (re-seed working copies before appends — the fragment-clobber incident).
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-handoff + #2026-07-07-incidents
- **relations:** travelling docs; README_summary_paragraph (the §0 orientation paragraph reused)
- **verify-later:** n/a (convention)

### Doc-consolidation manifests (dedup / thin / proposed-moves)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** dedup-manifest.tsv and thin-manifest.tsv record executed moves (action=move rows with moved_from → _archive/... destinations); PROPOSED_MOVES.tsv has an ACTION(edit:keep|move|archive|skip) decision column awaiting fill.
- **what:** Evidence of the documentation-consolidation system operating on the docs tree: a dedup pass (exact-duplicate groups with a chosen canonical, duplicates moved to _archive mirrors), a thinning pass (versioned running-notes families archived, e.g. idea.uk running_notes(NN)), and a proposal file for unclassified root files (API_DOCUMENTATION, summary.txt, tree, thin-manifest itself) awaiting keep/move/archive decisions. The docs/ root vonc families postdate or escaped these passes — relevant input for any future consolidation round.
- **sources:** docs/dedup-manifest.tsv; docs/thin-manifest.tsv; docs/PROPOSED_MOVES.tsv; docs/main_docs_directory_tree.txt
- **relations:** documentation-system (037 conventions); this concept register (stage-1 consumer)
- **verify-later:** docs/_archive contents match the manifests

### cmd/bundle context-assembly harness (contextkit)
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** "bundle_minilobby_trim.sh (v4) completed: bundle written to /tmp/bundle_minilobby_trim.md" (2026-07-09) after four documented failures.
- **what:** A read-only Go tool (in the contextkit tree, a SEPARATE Go module) that assembles a decision bundle for an LLM verdict: required -analysis/-root/-constitution/-task, repeatable -scope path[:Symbol]/-include/-doc, DB gathers via -psql (-schema-tables, -runtime-site/-page), -dry-run. Operational lessons made durable: resolve an action's file from the REGISTRY (key → Handler: symbol → function definition), never from header-comment conventions; scope a dedicated <key>_action.go file WHOLE but a shared file BY SYMBOL (attention dilution); run from inside contextkit with absolute -analysis/-constitution/-doc/-out and root-relative -scope; prefer the authored runbook's invocation over an example's shorthand. Used here to settle the sanctioned template-edit path before touching anything.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-bundle-v1→v4 (four entries); docs/bundle_minilobby_trim(4).sh (header); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4.0
- **relations:** cmd/diagnose harness; sanctioned edit paths (the question it settles)
- **verify-later:** docs/agent_docs/docs019.../go_files/contextkit/cmd/bundle; RUNBOOK_thin_slice invocation form

### cmd/diagnose read-only diagnosis harness
- **category:** diagnosis-loop
- **status-signal:** deployed
- **status-evidence:** HANDOFF_vonc_write_site_spec: runtime evidence "read read-only via the diagnosis harness's runtime gather"; full re-run command given with -seed-hypothesis/-dry-bundle.
- **what:** The diagnosis loop entry point: `go run ./cmd/diagnose` with -analysis (callgraph json), -constitution, -psql (read-only runtime gather against the cluster DB: agent_error_log, site_work_items), -seed-hypothesis/-seed-scope, -runtime-site/-page, producing per-iteration bundles (/tmp/diag_bundle_N.md, bundle-<id>/runtime.md); a -verdict-script drives the loop, the stub abstains without a model. The write_site_spec handoff shows the intended usage pattern: harness gathers evidence, a fresh session re-scopes and reads the real code.
- **sources:** docs/HANDOFF_vonc_write_site_spec_spec_data.md#how-to-get-the-evidence
- **relations:** cmd/bundle; fix-loop council (later consumers)
- **verify-later:** cmd/diagnose flags vs docs019 contextkit docs

### Sanctioned content-edit paths (content_data is truth; HTML patching rejected)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Doc 003 quoted verbatim in the HANDOFF: "this is why HTML patching was rejected as an edit mechanism"; fix_component_template's remove_element fix type and its header deferral to the section-editor confirmed from file headers; section_editor_actions.go confirmed to exist as Go code.
- **what:** `content_data` is the source of truth for section content; patching `page_components.rendered_html` is a bridge at best (lost on the next re-render) and was explicitly rejected as an edit mechanism. Template changes have designated routes: `fix_component_template_action` fix types (including `remove_element` — "removes HTML elements matching a pattern"), with page-component content changes deferred to the section-editor workflow. The mini-lobby trim's method question — which action edits a template, which re-render propagates it, what a NULL-content_data section does — was deliberately settled by bundle verdict rather than guessed. Fallback when no supported path exists: full-text template UPDATE (never multi-line REPLACE of nested markup), verified by length delta, propagated by a page_rerender item.
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/PLAN_provocation-card(3).md#method-corrected; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-09-bundle-primed; docs/bundle_minilobby_trim(4).sh (header)
- **relations:** two re-render paths; per-tool docs (method correction recorded); section-editor
- **verify-later:** fix_component_template_action.go fix types; section_editor_actions.go component_swap

### Two re-render paths + assemble-only rerender distinction
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Doc-003-derived and header-confirmed 2026-07-09: rerender_single_page "confirmed ASSEMBLE-ONLY" from its own header; light-path escalation rule quoted from 003.
- **what:** Three distinct "rerender" operations that must not be conflated: (1) FULL rebuild — needs_page → page-build-handler → page-content-writer (LLM regenerates copy); (2) LIGHT re-render — rerender_page_sections behind a page_rerender item: re-renders every section from EXISTING content_data via RenderComponentAction, no LLM, and escalates the whole page to a full rebuild when content_data is NULL; (3) ASSEMBLE-ONLY — rerender_single_page (the habitual rerender-*.sh trigger): reassembles stored page_components.rendered_html + stored site_components chrome and deploys; template-only edits will NOT appear through it. Mode-B sections likely have NULL content_data, making the light path escalate — the deciding probe for edit sequencing.
- **sources:** docs/PLAN_provocation-card(3).md#method-corrected; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/016b_debugging_guide_merged(3).md#open-threads (Part 2)
- **relations:** sanctioned edit paths; assemble-time visible-content filter; two chrome assembly paths
- **verify-later:** rerender_page_sections_action.go escalation branch; page_rerender item routing

### Site snapshots + dated-backup reversibility discipline
- **category:** site-snapshots-and-revert
- **status-signal:** deployed
- **status-evidence:** Pre-migration snapshot 044a0b57 taken 2026-06-23; take_site_snapshot call pattern in the migrations runbook; dated backup tables (_vonc_pc_backup_20260704/09 etc.) created before every risky UPDATE.
- **what:** Every significant change is preceded by reversibility: `take_site_snapshot(site_id, name, ..., 'manual')` for site state; `snapshot_agent('<type>','<reason>')`/`revert_agent` for agent definitions (never a hand-rolled agent_definitions_backup); ad-hoc dated `CREATE TABLE _<site>_<what>_backup_<date> AS SELECT ...` before direct row edits, with the explicit rule never to reuse an old backup name (CREATE TABLE IF NOT EXISTS silently no-ops while looking fresh); restore is UPDATE-in-place keyed on id, not delete+insert.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-snapshot; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-dated-backups; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§5-T0; docs/016b_debugging_guide_merged(3).md#key-schema-gotchas
- **relations:** debugging doctrine; direct-SQL-bypasses-guards caveat
- **verify-later:** take_site_snapshot / snapshot_agent SQL functions (doc 014)

### Work-item conventions and manual spec shapes
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01 (spec jsonb, item_key dedup via idx_swi_dedup, handler_agent, status flow, pipeline default 'build'); manual rerender/needs_page recipes proven repeatedly.
- **what:** site_work_items is the unit of work: `spec` jsonb (not spec_data at this layer), `item_key` (dedup), `handler_agent`, status detected→triaged→claimed→complete (dispatch picks up triaged/approved), `pipeline`. Manual page items require the FULL spec — page_id (real UUID inline), domain, filename, page_name; placeholder strings get claimed and fail ("invalid UUID length: 18"), and fixing them must filter on the PLACEHOLDER string, not the intended value (the wrong-WHERE no-op lesson). Duplicate insertions are cleaned by grouping on spec->>'page_name' and deleting the older of each pair. Fresh gen_random_uuid item_keys make re-fires safe.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#reference-manual-rerender + #duplicate-work-item-cleanup; docs/RUNNING_NOTES_vonc(36).md#work-item-fix-2026-06-24 + #2026-07-01-~13:40; docs/RUNBOOK_vonc_session(1).md#correct-spec-shape
- **relations:** item_key canonicalization; build-dispatch-loop; complete_error family
- **verify-later:** \d site_work_items; idx_swi_dedup definition

### item_key canonicalization (workItemKey builder)
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** 016b Part 3: "CODE PREPARED; NOT APPLIED" — workItemKey(itemType, target) builder in work_items_common.go; apply gated behind Part-2 verification.
- **what:** item_key prefixes drifted from item_type across creators: the adoption creator keyed BOTH needs_content_page and needs_tool_recreation as `needs_page:<name>`, so a tool and a content page of the same name collide on the dedup index and one is silently dropped. Fix: a shared workItemKey builder; the tool item moves to its own prefix while the content item deliberately keeps `needs_page:` co-dedup with planner builds (Option B, decision recorded).
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 3)
- **relations:** work-item conventions; dedup index
- **verify-later:** work_items_common.go workItemKey applied?; adoption creator key prefixes

### system-stats key-contract mismatch (content_data ↔ template key sets)
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** 016b Part 5 "TRIAGED, 2026-06-24"; remedy un-applied at that entry ("full content rebuild... then re-check"); the component itself later regenerated to q100 in the vonc arc.
- **what:** A populated-but-blank section is a content↔template KEY-CONTRACT problem, not a generation failure: system-stats' stored content_data keys (eyebrow/heading/stat_1_number...) shared ZERO keys with its template placeholders (eyebrow_label/section_headline/stat1_value...) after component-creator rewrote the component mid-flight, so every placeholder rendered empty and the (correct) visible-content filter dropped the band fleet-wide (usage_count 22). Durable heuristic: diff the two key sets directly; and a component schema change should trigger dependent rebuilds.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 5) + #wrong-turns (#4)
- **relations:** shared library field guard (the same incident class the guard now blocks); visible-content filter
- **verify-later:** whether schema-change→dependent-rebuild triggering exists (markPagesPendingRebuild covers regen; mid-build rewrites?)

### Two page-assembly paths with different chrome sources (stale site_components)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** 016b merged entry (from the parallel scheme-to-components thread): mechanism confirmed with provenance greps; paired-variable fix + workstreams referenced in SPEC_scheme_to_components (fix not claimed deployed in this corpus).
- **what:** Build path (page-build-handler → CompilePageSections → InjectHeader/Footer → RenderHeader/Footer) renders chrome FRESH — via style_collections.header_component_id (a dead, never-written column) falling through to GetComponentByFunction('site-header') or a dark RenderFallbackHeader. Rerender path reassembles stored page_components.rendered_html and injects STORED site_components.rendered_html, which can carry long-deactivated dark chrome — nothing refreshes site_components on deactivation, and stored section renders "fossilise" old templates (legacy `--accent-color` vars are the tell; needs_rerender re-fossilises; only a full rebuild re-renders templates). Provenance greps distinguish the three header origins; InjectHeader skips when a site-header class already exists.
- **sources:** docs/016b_debugging_guide_merged(3).md#light-site-renders-dark-chrome
- **relations:** two re-render paths; CSS variable naming (legacy-tell); post-025 theme flow
- **verify-later:** site_components refresh logic; style_collections.header_component_id writes; RUNBOOK/SPEC_scheme_to_components outcome

### sites.status vocabulary and the blast-radius filter trap
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 016b merged entry: UpdateSiteStatusAction (v3_site_actions.go:323) validated vocabulary read from code; the wrong 'active' filter incident recorded.
- **what:** `sites.status` vocabulary is draft/building/review/published/deployed/archived/error ('active' is a legacy hand-written value); no code filters on it — it is an informational lifecycle label, and build dispatch keys on site_work_items (a deployed site is still rebuildable). Heuristic: never scope blast-radius or "live sites" queries with status='active' (it silently dropped the site under investigation); enumerate GROUP BY status first. Companion reuse-gate lesson: a shared set_updated_at() trigger function already exists — check pg_proc/pg_trigger before creating.
- **sources:** docs/016b_debugging_guide_merged(3).md#sites.status
- **relations:** debugging doctrine (0-rows discipline); shared library blast-radius checks
- **verify-later:** UpdateSiteStatusAction vocabulary; pg_trigger set_updated_at users

### SQL template-surgery pattern (needle-gate discipline)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** 016b merged entry: "Template-surgery pattern that held up" with the 2026-07-02 false-alarm refinement; practised across the marker/ghost-row/truncation fixes.
- **what:** Safe in-DB template edits: (1) needle-gate read — every needle as a LIKE boolean PLUS occurrence counts so partial coverage is visible BEFORE mutating (counts must be counted from the dump, not recalled); (2) shell backup of the full column; (3) guarded idempotent UPDATE (exact-string nested replace or anchored regexp_replace with backreference, plus NOT LIKE pre-state guard); (4) RETURNING boolean checks; (5) rollback file. Postgres pitfalls: regex quantifier bounds cap at 255; substring-with-parens returns the capture group; gradient-embedded hexes escape naive background regexes; needles containing literal % can't be LIKE-gated (use position()). Anchor REPLACEs on the opening tag (see marker lesson); dump→edit-offline→full-text UPDATE for multi-line blocks.
- **sources:** docs/016b_debugging_guide_merged(3).md#sql-verification-pitfalls; docs/fix_archive_template_display(1).sql (header); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§8
- **relations:** marker anchoring; hidden-vs-author-CSS fix; sanctioned edit paths (this is the fallback)
- **verify-later:** n/a (practice)

### Marker/attribute REPLACE anchoring lesson (fix_marker_selector)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Bug introduced twice (provocation-card, lobby-grid), fixed via fix_marker_selector.sql with RETURNING checks (still_broken=f ×4), corrected HTML redeployed 2026-07-04; guide entry added.
- **what:** Adding an attribute by replacing the bare string `data-component="X"` also hits the section's own inline `querySelector('[data-component="X"]')`, producing a malformed two-attribute selector → SyntaxError → the cosmetic IIFE dies (loaders unaffected). Rule: anchor marker REPLACEs on the OPENING TAG (the copy followed by more attributes), revert only the in-selector copy (the one followed by `]`); better still, emit markers at generation.
- **sources:** docs/fix_marker_selector.sql (header); docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-marker-replace-broke; docs/016b_debugging_guide_merged(3).md#data-runtime-fill-marker-anchoring
- **relations:** generation-time guards (the prevention); SQL surgery pattern
- **verify-later:** n/a (lesson; instance fixed)

### `hidden` attribute vs author CSS (clone-template ghost rows)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Ghost-row fix verified end-to-end 2026-07-08 (rendered_len 7455→7671; live grep 2); prevention added to guide + component-creator requirement.
- **what:** The `hidden` attribute maps to UA-stylesheet `display:none`, which loses to ANY author `display` rule on the same element — so a hidden clone-template item inside a `display:grid` item class renders as a ghost row. Fix: a more specific author rule `[data-…-template] { display:none; }` in template AND instance (the REPLACE correctly fired twice — base selector + its mobile media-query copy). Prevention: component-creator must emit the hiding rule alongside `hidden` for clone templates.
- **sources:** docs/fix_archive_template_display(1).sql (header); docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/016b_debugging_guide_merged(3).md#hidden-attribute-loses
- **relations:** generation-time guards; clone-template list pattern
- **verify-later:** component-creator prompt includes the hiding-rule requirement?

### Auto-lock on deploy (page_components lock trigger)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01: "trigger_auto_lock_on_deploy auto-locks on deploy... lock_type permanent|timed|review"; lock check run pre-rebuild (all 4 index rows unlocked).
- **what:** page_components carries locked_at/lock_type/locked_by with a trigger that auto-locks components on deploy (fires on UPDATE). Operational consequence observed: deployed components MAY be locked, so rebuilds/re-renders must check lock state (a lock could block re-render of a target or protect neighbours); on the vonc index all rows were NULL-locked so the behaviour never actually bit in this corpus.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:25 + #2026-07-01-~13:55; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** locks category (031); save_page_sections (does it honour locks? open question in 016b Part 4)
- **verify-later:** trigger_auto_lock_on_deploy definition; save_page_sections lock handling

### Standing engineering rules (the session working constitution)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** Restated as "Standing rules (brief)" in the v2 carried-forward state and as "Standing instruction from the user, in force" in the HANDOFF.
- **what:** The recurring project rules this corpus operates under: schema-before-SQL (\d first); reuse/alter before create (STEP ZERO); structural over quick fixes; workflows THIN with logic in Go actions; no sub-workflows in SQL — spawn sub-agents; every agent is an orchestrator; agents respond to the CALLER's responses topic; no logger.Debug (invisible in cluster logs); British English; flag variable/signature changes; never treat 0 rows as decisive; verify against deployed artifacts not pod logs; no summary docs unless asked; work in reasonable step sizes.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#standing-rules; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§1; docs/HANDOFF_vonc_write_site_spec_spec_data.md#standing-rules
- **relations:** debugging doctrine; development-guide (001) anchors
- **verify-later:** n/a (convention; verify against 001/002/003 docs)

### Sites deployment chain (git → GitHub Actions → Backblaze B2) + image-tag chassis deploys
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** Used throughout ("after Actions propagates (a few minutes)"); 016b orientation: "Deployment is image-tag based... each agent's image_tag is bumped to adopt it; workflow (default_config) changes are DB-only and take effect immediately."
- **what:** Everything site-facing reaches production via git_commit to the 'sites' repo (files map keyed by repo-relative path — pages, tools/assets/*.js, assets/js/snippets.js, data/*.json) → GitHub Actions → Backblaze B2 (public), with the long-running git-adapter handling commits. Platform code ships as a chassis image (GitHub → Actions → image) adopted by bumping per-agent image_tag — so a source revert only reaches the cluster on the next build/push, while agent workflow changes (agent_definitions.default_config) are DB-only and instant. B2 404 NoSuchKey is the "page never deployed" signature.
- **sources:** docs/016b_debugging_guide_merged(3).md#orientation; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 (git_commit pattern); docs/RUNBOOK_phase2_provocation_js(29).md#4.2-4.3; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2
- **relations:** system-architecture; plan-storage revert note (pods keep old behaviour until push)
- **verify-later:** git-adapter; sites repo Actions workflow

### Option 1 — build-time static content for the daily shells (rejected alternative)
- **category:** vonc
- **status-signal:** abandoned
- **status-evidence:** Early migrations-runbook versions carried "Recommendation: Option 1 for the first deployable version — get real content" (dropped line); final: "DECISION MADE: Option 2... Option 1 would freeze a single set of provocations permanently, defeating the daily-content product."
- **what:** The rejected fix for the empty index shells: regenerate them WITH proper input_schemas so the content writer fills them at build time. Briefly the recommended first-version route in early runbook versions, then dropped when the original Spark roadmap (daily provocations via client-side JS) was recovered — build-time content would bake one day's provocations permanently. Survives only in its correct form: genuinely static shells (brief-explanation) ARE fixed by regeneration.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#step-7 (decision); early-version dropped lines (family diff); docs/PLAN_spark_provocation_pipeline.md#why-option-2
- **relations:** static-vs-dynamic distinction; brief-explanation (where Option-1 logic is right)
- **verify-later:** none (historical)

### API documentation system (OpenAPI external + per-service internal API.md)
- **category:** documentation-system
- **status-signal:** unknown
- **status-evidence:** File dated Aug 2025 (per the directory-tree listing) and flagged "unclassified" in PROPOSED_MOVES.tsv; no corroborating recent activity in this unit.
- **what:** A two-tier API documentation practice for the platform: customer-facing APIs documented as OpenAPI 3.0 (internal/auth-service/api/openapi.yaml + swagger annotations in *_swagger.go, `make swagger`/`make validate-openapi`, Swagger UI/Redoc/Editor via docker-compose) and internal service communication documented as per-service API.md files (auth-service, core-manager, agents/*, adapters/*) covering Kafka topics, message formats, DB schemas, env vars; CI validation workflow proposed. Predates the vonc corpus; whether the practice is followed is unverified.
- **sources:** docs/API_DOCUMENTATION.md; docs/PROPOSED_MOVES.tsv
- **relations:** admin-dashboard-and-api (the gateway/API surface it documents); documentation-system
- **verify-later:** internal/auth-service/api/openapi.yaml exists?; make targets swagger/validate-openapi; API.md coverage

### Basic operations reference (kcat spawn, scale, monitoring)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** summary.txt is a concatenation of basic_usage docs actively describing the current operating procedure (spawn_group website-builder message shape, headers, monitoring queries).
- **what:** The operator's basic-usage layer: scale the deployment set up/down (agent-chassis, auth-service, content-creator-agent, core-manager, image-generator-adapter, reasoning-agent, web-search-adapter); post spawn_group/orchestrate messages via kcat from a test pod to the cross-namespace Kafka bootstrap with required headers (correlation_id, request_id, client_id, agent_instance_id, fuel_budget); monitor via orchestrator_state/orchestration_states by correlation_id. The fuel_budget header and the fixed header set are part of the platform's message contract.
- **sources:** docs/summary.txt; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§6
- **relations:** manual agent trigger pattern; system-architecture (topics)
- **verify-later:** docs/basic_usage originals; current deployment list

### Result-contract drop fix (child workflow result replaced by a stub)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** 016b Part 1: "DONE... Shipped 2026-06-18 (result_spec.go + coordinator.go); verified — gamesdesign index rebuilt+deployed 06-19."
- **what:** The chassis coordinator used to discard a child workflow's result (singular output_field, or oversize) and substitute a stub that still reported success — producing no-op saves under `complete` status (a root member of the silent-success family, and the resolution of the long-open "index returns thin content" question: it was the stub, not thin generation). Fixed in result_spec.go + coordinator.go. Carried here because the 016b copies in this unit are the guide's cumulative record; the docs024 consolidated guide is the canonical home.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 1)
- **relations:** complete_error family (sibling); trust-the-artifact doctrine
- **verify-later:** result_spec.go / coordinator.go in platform

---

## Dropped-concept notes from family deltas (audit)

- RUNNING_NOTES_vonc.md (base): migration renumbering — "DB migration: Migration 003" became
  "Migration 002 = agent_definition condition fix; render_mode sweep DROPPED" (captured as a
  concept above).
- RUNBOOK_vonc_migrations early versions: original "Migration 002 — Fix render_mode on
  components" heading and the "Recommendation: Option 1 for the first deployable version"
  line (both captured as abandoned concepts above); "Two fix options — choose one" framing.
- RUNBOOK_phase2 early versions: Gap-2 sub-options (a) creator-emits-companion-snippet vs
  (b) loader-snippets-as-library-fixtures (superseded by Tier E + loader-builder); the
  step-checklist framing (P2-1..P2-6, FX-1..FX-6, PD-1..PD-3) whose IDs the later docs still
  reference; all content otherwise carried forward into (29).
- PLAN_provocation-card(2)→(3): the pre-correction trim method (hand-UPDATE the live
  instance) — replaced by the bundle-verdict method; captured under "sanctioned edit paths".
- PLAN_dynamic_sections(2)→(4): `site_plan_directives` named as a site_plans child table in
  the pre-supersession decision text; not mentioned in the final doc (noted in the plan-storage
  concept).
- HANDOFF base/(1)→(2): same method correction (rejected mechanism → bundle §4.0).
- bundle_minilobby_trim base..(3)→(4): resolver evolution v1→v4 (registry-first resolver,
  module boundary, scope-by-symbol) — captured under the cmd/bundle concept.
- provocations.sample.json v1→v3: additive key evolution today/lobby → +arena → +archive —
  captured in the data-contract concept.

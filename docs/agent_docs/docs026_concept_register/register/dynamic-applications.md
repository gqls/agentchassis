# Register — dynamic-applications

12 concepts, consolidated from 36 raw extractions (the cluster input file contains
the entire dynamic-applications block set duplicated exactly twice — 18 unique raw
blocks appearing twice each — plus further cross-unit duplication of the same
concepts under different titles) across units U01, U02, U17a, U21, U23, U25.

### DYN-001 — Dynamic applications direction (three tiers; framework specs; thin generated backends)
- **status:** aspirational
- **status-evidence:** doc 022 tier 1 "now → near term", tiers 2–3 medium/longer term.
- **what:** Tier 1 is static+dynamic components (forms via external services, client search, client-side A/B); Tier 2 is agent-powered per-site backends (workers/lightweight services fed by agents — business logic stays in agents, the backend is a thin render layer); Tier 3 is full application generation (admin panels, SaaS prototypes). Governing principles: framework specs stored for each target stack, one site/one repo/one deployment, generated-vs-human content marked with human edits taking precedent, and incremental complexity (mailto → Formspree → Worker → CRM).
- **sources:** 022 full
- **relations:** infrastructure three layers (adoption-pipeline ADO-001, shared tier language); CSS variable contract
- **verify-later:** none built beyond tier 1 basics

### DYN-002 — Interactive fingerprint parse stage (C1–C6) — adoption's missing interactive-logic capture
- **status:** partial
- **status-evidence:** "C1 — extract_interactive_fingerprint (new action). Status: in progress, 2026-05-15"; C2–C6 planned; a later doc calls it "Smaller piece of work than I first thought — closer to a couple of days".
- **what:** A planned Go extractor over crawled rawHtml capturing canvas elements, inline/external scripts, event handlers, forms, and library signals (rAF, canvas contexts, jQuery/Three/Phaser/React/Vue), producing a per-page type_hint heuristic (calculator/game_or_animation/interactive_widget/static). Follow-on stages: external-JS fetch loop, enrichment, an LLM interactive_intent brief with feasibility markers, and new `interactive_reference`/`interactive_intent` site_specs aspects — reusing the proven design-fingerprint shape (goquery selectors + firecrawl_scrape fetch + LLM synthesis) but deliberately as a new file, not an extension of the design extractor. AST parsing is out of scope.
- **sources:** FOCUS_interactive_content_generation(4).md#Path-C; WM/FOCUS_interactive_content_generation(3).md#where-the-parsing-capability-work-would-slot-in
- **relations:** design fingerprint pattern (adoption-pipeline ADO-003); Firecrawl escalation ladder (adoption-pipeline ADO-008); interactive fingerprint gap named from the adoption side (adoption-pipeline ADO-016)
- **verify-later:** extract_interactive_fingerprint_action.go existence; interactive_reference aspects in site_specs

### DYN-003 — Four-stage interactive-content pattern (parse / assess / generate / integrate)
- **status:** aspirational
- **status-evidence:** "Not a roadmap — a map of the territory" (family doc, updated through 2026-05-15); sequencing "locked in 2026-05-14"; tools implement it "mostly", minus the parse stage.
- **what:** The reference shape for handling any interactive content type encountered on adopted sites: parse the source, assess producibility (producible_now / producible_simpler / blocked, per the doc-028 spec-status model with feasibility-recheck promotion), generate the artefact, integrate into the build pipeline. Agreed sequencing: Path C (parse stage) → Path D (tool reliability) → Path A (games) → B (news publishing) / E (numbered-component cleanup).
- **sources:** FOCUS_interactive_content_generation(4).md#four-stage, #Sequencing; WM/FOCUS_interactive_content_generation(3).md#the-four-stage-pattern, #capability-assessment
- **relations:** tools pipeline; games gap (DYN-004); interactive fingerprint parse stage (DYN-002); spec-has-status (site-spec-and-classifier SPEC-001)
- **verify-later:** feasibility-recheck task existence; tool-suggester/deployer/generator/improver/auditor agent set

### DYN-004 — Games as a content type (largest pipeline gap)
- **status:** aspirational
- **status-evidence:** "Games — nothing yet … page_type='game' doesn't exist in the classifier vocabulary" (2026-05); the vocabulary absence later caused the 2026-05-26 duplication bug.
- **what:** No game-suggester/generator/improver/auditor, no game template library, no game_health check, no spec aspect exists — game-list components force fabrication. The plan is to copy the tools pattern wholesale and add `game` to the page_type vocabulary. The gap was not cosmetic: without it, the planner re-typed adopted game pages to `tool`, driving rename and duplication.
- **sources:** FOCUS_interactive_content_generation(4).md#Games, #classification-vocabulary; HANDOFF_2026-05-26…md#diagnosis
- **relations:** page_type vocabulary gap; four-stage pattern (DYN-003); library model (DYN-005)
- **verify-later:** plan_site Canonical Page Types list today

### DYN-005 — Generator architecture convergence (shared interactive-artefact-generator)
- **status:** aspirational
- **status-evidence:** "Worth considering once two more generators exist; one isn't enough to abstract from."
- **what:** Every content-type generator (tools, games, news articles, dashboards) needs a brief contract, prompt template, persistence action, page-creation step, tiered quality checks, and a companion-content step; a shared base with per-type specialisation is anticipated once games exist. The library model (canonical templates, forked_from IS NULL, per-site forks) is the copyable storage shape underneath it.
- **sources:** FOCUS_interactive_content_generation(4).md#Generator-architecture, #Library-model, #Quality-model
- **relations:** tools pipeline; games gap (DYN-004)
- **verify-later:** n/a (design idea, not built)

### DYN-006 — Tool builder tiers (static / dynamic / application)
- **status:** partial
- **status-evidence:** docs017/019b tier table ("Static: component library; Dynamic: self-contained JS, LLM-generated or pre-built; Application: engineer-built only") and platform stress test ("Agent-as-API pattern. User IS the HITL"); mortgagecalculator + website-design.com cited as early instances.
- **what:** Interactive functionality classified by creation risk: static HTML components from the library, dynamic self-contained JS applications (calculators, visualisations) that LLMs may generate, and full applications with API integration reserved for engineers. The agent-as-API pattern for platform sites treats the end user as the HITL. This tiering matured into the current tool-pipeline/tool-library/tool-lifecycle systems.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#6-Tool-Builder-Agent; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** tool-pipeline (its successor); JavaScript management; finance/tools stress test
- **verify-later:** current tool generation pipeline lineage

### DYN-007 — Runtime-fill mechanism (data-runtime-fill shells + client loaders + JSON feed)
- **status:** deployed
- **status-evidence:** "That mechanism is now proven three times over" — the daily provocation card, the six-room lobby grid, and the Provocations Archive, all browser-verified (HANDOFF §0, 2026-07-09).
- **what:** vonc/Spark's central delivery mechanism for daily-changing content on a static site (doc 022 Tier 1): sections ship as deliberately empty shells at build time; the component's `<section>` carries `data-runtime-fill="true"` so the assembler's visible-content filter exempts it and keeps the shell; an IIFE loader (stored as a js_snippet, bundled into `/assets/js/snippets.js`) fetches a JSON feed (e.g. `/data/provocations.json`) in the visitor's browser and fills the shell's selectors, failing gracefully to an empty state. Explicitly on-doctrine per doc 022 Tier 1 ("the dynamic part runs in the browser; backend complexity lives in agents"). Distinguished from build-time content: only daily-dynamic shells get loaders; static explainers get regenerated schemas and content-writer fills.
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2; docs/RUNBOOK_phase2_provocation_js(29).md#what-phase-2-is, #on-doctrine-check; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md; docs/social001_vonc_tiktok_social/tool_docs/PLAN_lobby-grid(3).md
- **relations:** visible-content filter exemption; two JS delivery paths (DYN-008); js_snippets library (DYN-009); Phase-3 pipeline (the missing data producer)
- **verify-later:** rerender_single_page_action.go reRuntimeFill regexp; js_snippets rows provocation-card-loader/lobby-grid-loader/provocations-archive-loader; /assets/js/snippets.js and /data/provocations.json on vonc.com

### DYN-008 — Two JS delivery paths (component js_content vs js_snippets bundle) and the inline-script truncation bug class
- **status:** deployed (mechanism); partial (extraction coverage / hardening)
- **status-evidence:** RUNNING_NOTES 2026-06-29 (PD-1 answered): latest-news fetches via extracted component JS (Path 1); Path-2 loader proven live same day; 2026-07-09 NOTES_lobby-grid(6): extraction pattern already live for three components, but provocation-card/lobby-grid inline scripts remained unextracted.
- **what:** Two separate JS delivery mechanisms coexist. Path 1: a component's inline `<script>` is extracted (`separateInlineJS`) to `content_components.js_content` and deployed as `/tools/assets/{function}.js` automatically on every page rerender (how gauntlet-interface, latest-news, and archetype-quiz ship JS). Path 2: library `js_snippets` rows are bundled to `/assets/js/snippets.js` by site-asset-renderer, outside the normal build/rerender flow. PD-3 decided Path 1 is the durable home for fetch-and-fill loaders going forward; Path 2 remains the working interim. A bug class was discovered alongside it: components stored via paths predating extraction keep raw inline scripts with empty js_content, and one template was truncated mid-script at generation (token limit) because store validation checks unclosed `<style>` but not `<script>` — hardening items proposed: script-balance check at store time, warn on unterminated scripts.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00, #2026-06-29-~17:20; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-09; docs/RUNBOOK_phase2_provocation_js(29).md#path-decision, #framework-fix
- **relations:** runtime-fill mechanism (DYN-007); js_snippets library (DYN-009); generation-time guards (DYN-012, the no-inline-script rule that makes this bug class impossible going forward)
- **verify-later:** rerender_single_page_action.go collectJSAssets; content_components.js_content coverage across components; separateInlineJS in store_generated_component

### DYN-009 — js_snippets library + render_js_snippets_for_site + site-asset-renderer bundling
- **status:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29: renderer ran clean (commit eb7f2ac), snippet bundled; live bundle header "3 active snippet(s)" 2026-07-07/09; VERDICT Q5 (2026-07-09): "Direct SQL is the only writer … render_js_snippets_for_site is a pure reader."
- **what:** `js_snippets` is a library-wide table (no site_id) of JS behaviours keyed by `applies_to` (a jsonb array of component functions), written only by direct SQL. `render_js_snippets_for_site` selects active snippets whose `applies_to` overlaps the site's component functions, concatenates them (ordered by name, header comments, empty bundle still written so the head `<script src>` never 404s) into `/assets/js/snippets.js`, committed by the site-asset-renderer agent via git_commit. Loaders self-check for their section so a global snippet is inert on other sites; the table has no updated_at column. Pre-existing snippets were small generic behaviours (accordion, scroll-reveal); the fetch-and-fill loaders are a new, heavier use of the table.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00, #2026-06-25-~19:30; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#1; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** two JS delivery paths (DYN-008); runtime-fill mechanism (DYN-007); js-bundle-stale gap (DYN-010)
- **verify-later:** render_js_snippets_for_site_action.go; site_asset_actions.go; js_snippets table contents

### DYN-010 — js-bundle-stale gap (site-asset-renderer not wired into ongoing builds)
- **status:** aspirational
- **status-evidence:** FX-6 checkbox never ticked; "Gap 3 is NOT on the critical path... still a real latent issue for genuinely-generic snippets, but lower priority" (RUNNING_NOTES 2026-06-29).
- **what:** Only `rerender-site` and `webdesign-agent` reference site-asset-renderer, so `/assets/js/snippets.js` is rebuilt at initial design and full site rerender but nothing re-runs it when a js_snippets row is added or changed later — the direct cause of the first fetch-and-fill loader never reaching vonc initially. The proposed fix (a design-discovery-agent check: "site has an applicable active snippet newer than its deployed bundle" → spawn site-asset-renderer) was deprioritised after Path 1 was chosen as the durable home for loaders; manual trigger scripts remain the working practice.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#gap-3; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00, #2026-06-29-~17:20
- **relations:** js_snippets library (DYN-009); design-discovery-agent named checks; two JS delivery paths (DYN-008)
- **verify-later:** design-discovery-agent run_discovery_checks array; agent_definitions referencing render_js_snippets

### DYN-011 — loader-builder agent + section descriptor + Tier E runtime-feed source (autonomy design)
- **status:** aspirational
- **status-evidence:** PLAN_dynamic_sections gap 3; "the missing piece (provocation-card/lobby-grid loaders were hand-built)" (2026-07-04); HANDOFF §9.6: "Autonomy design... The two hand-built loaders are its reference implementations."
- **what:** The designed path from hand-built runtime-fill to autonomous generation: plans would carry a section descriptor declaring role/kind/data_feed; component-creator would gain a Tier E schema source (`source: "feed.{name}"`) emitting stable-selector shells plus a loader; a proposed loader-builder agent — necessarily a sibling of tool-generator (which explicitly forbids fetch), not a variant of it — would LLM-generate the fetch-and-fill IIFE from the section's DOM contract and feed shape, installed as a js_snippet and bundled by site-asset-renderer. The framework currently has component-creator (section templates) and tool-generator (tools) but no runtime-fill loader builder; the lobby-grid and provocations-archive hand builds are its explicit reference implementations.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9, #9.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04
- **relations:** tool-generator (tool-pipeline); runtime-fill mechanism (DYN-007); component creation contract
- **verify-later:** agent_definitions for tool-generator (no-fetch rule); absence of any loader-builder agent; PLAN_dynamic_sections_and_loaders.md

### DYN-012 — Generation-time guards for dynamic components (the archive-list reference build)
- **status:** deployed
- **status-evidence:** SPEC_provocations-archive-list AS BUILT (2026-07-09): "Both generation-time guards held on the first attempt: has_marker = t, has_inline_script = f … first live validation."
- **what:** Bakes the lessons of DYN-007/DYN-008 into generation rather than repairing after: instruct component-creator to emit `data-runtime-fill` in the section tag directly (no post-hoc marker SQL), forbid `<script>` elements entirely (making the extraction/truncation bug class impossible), make header copy llm-sourced so nothing can defer, use a single hidden clone-template item for variable-length lists with an explicit `[data-…-template] { display:none; }` rule (the `hidden` attribute loses to author display rules), and include a visible empty state so the page ships before data lands. `provocations-archive-list` (component 70d6662a) is the canonical reference build validating all of these on the first attempt.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3
- **relations:** component creation contract; runtime-fill mechanism (DYN-007); editing-stored-HTML landmines
- **verify-later:** component 70d6662a row; component-creator description patterns

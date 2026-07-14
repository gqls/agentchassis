
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dynamic applications direction (three tiers; framework specs; thin generated backends)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** 022 tier 1 "now → near term", tiers 2–3 medium/longer term
- **what:** Tier 1 static+dynamic components (forms via external services, client search, client-side A/B); Tier 2 agent-powered per-site backends (workers/lightweight services fed by agents — business logic stays in agents, backend is a thin render layer); Tier 3 full application generation (admin panels, SaaS prototypes). Principles: framework specs stored for each target stack; one site one repo one deployment; generated-vs-human content marked, human edits precedent; incremental complexity (mailto → Formspree → Worker → CRM).
- **sources:** 022 full
- **relations:** infrastructure layers (007); CSS variable contract (shared)
- **verify-later:** none built beyond tier 1 basics

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Interactive fingerprint parse stage (C1–C6)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** "C1 — extract_interactive_fingerprint (new action). Status: in progress, 2026-05-15"; C2–C6 planned
- **what:** New Go extractor over crawled rawHtml capturing canvas elements, inline/external scripts, event handlers, forms, library signals (rAF, canvas contexts, jQuery/Three/Phaser/React/Vue) and a per-page type_hint heuristic (calculator/game_or_animation/interactive_widget/static); then external-JS fetch loop (C3), enrich (C4), LLM interactive_intent brief with feasibility markers (C5), written to new interactive_reference/interactive_intent spec aspects (C6). Deliberately a new file, not an extension of the design extractor; AST parsing out of scope.
- **sources:** FOCUS_interactive_content_generation(4).md#Path-C
- **relations:** design fingerprint pattern; capability markers; Firecrawl executeJavascript escalation
- **verify-later:** extract_interactive_fingerprint_action.go existence; interactive_reference aspects in site_specs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Four-stage interactive-content pattern (parse / assess / generate / integrate)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Not a roadmap — a map of the territory" (family v4, updated through 2026-05-15); tools implement it "mostly", minus the parse stage
- **what:** The reference shape for handling any interactive content type encountered on adopted sites: parse the source, assess producibility (producible_now / producible_simpler / blocked per the 028 spec model with feasibility-recheck promotion), generate the artefact, integrate into the build pipeline. Agreed sequencing: Path C (parse stage) → Path D (tool reliability — tools "currently don't work") → Path A (games) → B (news publishing) / E (numbered-component cleanup).
- **sources:** FOCUS_interactive_content_generation(4).md#four-stage, #Sequencing
- **relations:** tools pipeline; games gap; news publishing gap; capability assessment
- **verify-later:** feasibility-recheck task existence; state of tool reliability work

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Games as a content type (largest pipeline gap)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Games — nothing yet … page_type='game' doesn't exist in the classifier vocabulary" (2026-05); the vocabulary absence later CAUSED the 05-26 duplication bug
- **what:** No game-suggester/generator/improver/auditor, no game template library, no game_health check, no spec aspect; game-list components force fabrication. Plan: copy the tools pattern wholesale; add `game` to the page_type vocabulary (Option 1 hardcode now, Option 4 page_types table later — canonicalise kebab/snake first). The missing `game` type is not cosmetic: the planner re-typed adopted game pages to `tool`, driving rename + duplication.
- **sources:** FOCUS_interactive_content_generation(4).md#Games, #classification-vocabulary; HANDOFF_2026-05-26…md#diagnosis
- **relations:** page_type vocabulary gap; four-stage pattern; library model
- **verify-later:** plan_site Canonical Page Types list today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Generator architecture convergence (shared interactive-artefact-generator)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Worth considering once two more generators exist; one isn't enough to abstract from"
- **what:** Every content-type generator (tools, games, news articles, dashboards) needs a brief contract, prompt template, persistence action, page-creation step, tiered quality checks and companion-content step; a shared base with per-type specialisation is anticipated once games exist. The library model (canonical templates, forked_from IS NULL, per-site forks) is the copyable storage shape.
- **sources:** FOCUS_interactive_content_generation(4).md#Generator-architecture, #Library-model, #Quality-model
- **relations:** tools pipeline; games gap
- **verify-later:** n/a (design idea)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Interactive content generation (four-stage parse/assess/generate/integrate)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** FOCUS_interactive_content_generation(3) "Tools — most mature … Games — nothing yet"; sequencing "locked in 2026-05-14"
- **what:** A map for building tools/games/news/other interactive types via a four-stage pattern (parse the source, assess what's producible, generate the artefact, integrate into a page). Tools are most mature but missing a parse stage; games have nothing. Capability assessment is a spec-lifecycle property marking each element `producible_now`/`producible_simpler`/`blocked`.
- **sources:** WM/FOCUS_interactive_content_generation(3).md#the-four-stage-pattern, WM/FOCUS_interactive_content_generation(3).md#whats-working-today, WM/FOCUS_interactive_content_generation(3).md#capability-assessment
- **relations:** tool pipeline; component creator; spec-has-status; adoption parse-stage
- **verify-later:** tool-suggester/deployer/generator/improver/auditor; page_types vocabulary

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption parse-stage for interactive logic (interactive_reference/intent)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** FOCUS_interactive_content_generation(3) "1. Path C — Parse stage in adoption … Smaller piece of work than I first thought — closer to a couple of days"
- **what:** The prioritised gap: adoption captures markdown/design but not interactive JS. Closing it reuses the proven design-extraction shape: add `<script>`/`<canvas>` selectors to goquery, fetch `<script src>` via existing firecrawl_scrape, and add an LLM step producing `interactive_reference`/`interactive_intent` site_specs aspects.
- **sources:** WM/FOCUS_interactive_content_generation(3).md#where-the-parsing-capability-work-would-slot-in, WM/FOCUS_interactive_content_generation(3).md#sequencing-agreed-order, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** design fingerprint; interactive content generation; tool-recreation-handler misroute
- **verify-later:** extract_design_fingerprint_action.go; firecrawl_scrape; proposed extract_interactive_fingerprint

<!-- SOURCE: U21_legacy_docs_b.md -->
### Tool builder tiers (static / dynamic / application)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** docs017/019b tier table ("Static: component library; Dynamic: self-contained JS, LLM-generated or pre-built; Application: engineer-built only") and platform stress test ("Agent-as-API pattern. User IS the HITL"); mortgagecalculator + website-design.com cited as early instances in docs018/008b.
- **what:** Interactive functionality classified by creation risk: static HTML components from the library; dynamic self-contained JS applications (calculators, visualisations) that LLMs may generate; full applications with API integration reserved for engineers. The agent-as-API pattern for platform sites treats the end user as the HITL. Matured into the tool-pipeline/tool-library/tool-lifecycle systems.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#6-Tool-Builder-Agent; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** tool-pipeline (successor); JavaScript management (docs017/023); finance/tools stress test.
- **verify-later:** current tool generation pipeline lineage.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Runtime-fill mechanism (data-runtime-fill shells + client loaders + JSON feed)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** HANDOFF §0: "That mechanism is now proven three times over" (provocation-card 2026-06-29, lobby-grid 2026-07-04, archive 2026-07-08, all browser-verified).
- **what:** Sections ship deliberately EMPTY at build time; the component `<section>` carries `data-runtime-fill="true"` so the assembler keeps the shell; an IIFE loader stored in `js_snippets` and bundled into `/assets/js/snippets.js` fetches `/data/provocations.json` in the visitor's browser and fills the shell's selectors, failing gracefully. Explicitly on-doctrine per doc 022 Tier 1 ("dynamic content injection... the dynamic part runs in the browser"; backend complexity lives in agents).
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2; docs/RUNBOOK_phase2_provocation_js(29).md#what-phase-2-is + #on-doctrine-check; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35; docs/README_summary_paragraph_for_handoff.md
- **relations:** visible-content filter exemption; js_snippets library; two JS delivery paths; static-vs-dynamic section distinction
- **verify-later:** rerender_single_page_action.go (reRuntimeFill regexp); js_snippets rows provocation-card-loader/lobby-grid-loader/provocations-archive-loader; /assets/js/snippets.js on vonc.com

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two JS delivery paths (Path 1 component js_content vs Path 2 js_snippets bundle)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~17:20 (PD-1 answered): latest-news fetches via its extracted component JS (Path 1); Path-2 loader proven live the same day; 2026-07-07 side-evidence: extraction pattern live on three tool components.
- **what:** Two separate JS delivery mechanisms exist. Path 1: a component's inline `<script>` is extracted to `content_components.js_content` and deployed as `/tools/assets/{function}.js` automatically on every page rerender (how gauntlet-interface, latest-news, archetype-quiz ship JS — including news's data fetch). Path 2: library `js_snippets` rows are bundled to `/assets/js/snippets.js` by site-asset-renderer — NOT part of the normal build/rerender flow. The vonc daily-feed shells "fell between" the two paths. PD-3 decided Path 1 is the durable home for fetch-and-fill loaders; the Path-2 snippets remain the live working interim.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00 + #2026-06-29-~17:20; docs/RUNBOOK_phase2_provocation_js(29).md#path-decision + #framework-fix; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed
- **relations:** separateInlineJS extraction; js_snippets library; runtime-fill mechanism; js-bundle-stale gap
- **verify-later:** rerender_single_page_action.go collectJSAssets; content_components.js_content for gauntlet-interface/latest-news; /tools/assets/*.js in the sites repo

<!-- SOURCE: U23_docs_root_vonc.md -->
### js_snippets library + render_js_snippets_for_site + site-asset-renderer
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~17:20: renderer ran clean (commit eb7f2ac), snippet bundled; bundle header "3 active snippet(s)" 2026-07-07.
- **what:** `js_snippets` is a LIBRARY-WIDE table (no site_id) of JS behaviours keyed by `applies_to` (jsonb array of component functions). `render_js_snippets_for_site` selects active snippets whose applies_to overlaps the site's component functions, concatenates them (ordered by name, header comments, empty bundle still written so the head `<script src>` never 404s) into `/assets/js/snippets.js`, committed by the `site-asset-renderer` agent via git_commit. Loaders self-check for their section so a global snippet is inert on other sites. Pre-existing snippets were all small generic behaviours (accordion, scroll-reveal...); the fetch-and-fill loaders are a new, heavier use of the table.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 + #2026-06-25-~19:30 (inventory); docs/RUNBOOK_phase2_provocation_js(29).md#mechanism-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** two JS delivery paths; site-asset-renderer triggering gap; runtime-fill mechanism
- **verify-later:** render_js_snippets_for_site_action.go; site_asset_actions.go; js_snippets table contents

<!-- SOURCE: U23_docs_root_vonc.md -->
### js-bundle-stale gap (site-asset-renderer not wired into the build)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** FX-6 checkbox never ticked; RUNNING_NOTES 2026-06-29: "Gap 3 is NOT on the critical path... still a real latent issue for genuinely-generic snippets, but lower priority."
- **what:** Only `rerender-site` and `webdesign-agent` reference site-asset-renderer, so `/assets/js/snippets.js` is rebuilt at initial design and full site rerender but nothing re-runs it when a js_snippets row is added/changed later — the direct cause of the first loader never reaching vonc. Proposed fix: a design-discovery-agent check ("site has an applicable active snippet newer than its deployed bundle" → spawn site-asset-renderer). Deprioritised after PD-3 chose Path 1, never built; manual trigger scripts are the working practice.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#gap-3; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00 (GAP 3 CONFIRMED) + #2026-06-29-~17:20
- **relations:** design-discovery-agent named checks; two JS delivery paths
- **verify-later:** design-discovery-agent run_discovery_checks array; agent_definitions referencing render_js_snippets

<!-- SOURCE: U23_docs_root_vonc.md -->
### loader-builder agent (fetch-and-fill sibling of tool-generator)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** PLAN_dynamic_sections gap 3; "the missing piece (provocation-card/lobby-grid loaders were hand-built)" (2026-07-04); the two hand-built loaders named as its reference implementations.
- **what:** A proposed agent that LLM-generates client-side fetch-and-fill loaders for dynamic sections: input = the section's DOM contract + feed shape; output = a graceful IIFE installed as a js_snippet and bundled by site-asset-renderer. Modelled on tool-generator (which LLM-generates, saves and wires SELF-CONTAINED tools) but necessarily a SIBLING because tool-generator explicitly forbids fetch. The framework currently has component-creator (section templates) and tool-generator (tools) but no runtime-fill loader builder.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** tool-generator (tool-pipeline); Tier E; provocation_card_loader.js / lobby_grid_loader.js / provocations_archive_loader.js as references
- **verify-later:** agent_definitions for tool-generator (no-fetch rule); absence of any loader-builder agent

<!-- SOURCE: U25_leopardess_social.md -->
### Runtime-fill mechanism (data-runtime-fill shells + client loaders)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** HANDOFF §0 (2026-07-09): "That mechanism is now proven three times over: the daily provocation card, the six-room arena grid, and … the Provocations Archive."
- **what:** vonc/Spark's central delivery mechanism for daily-changing content on a static site (doc 022 Tier 1): sections ship as deliberately empty shells whose `<section>` carries data-runtime-fill="true"; the page assembler's visible-content filter exempts marked sections; an IIFE loader fetches /data/provocations.json in the visitor's browser and fills the DOM contract, failing gracefully (shell + empty state remain). Distinction enforced against build-time content: static explainers get regenerated schemas and content-writer fills; only daily-dynamic shells get loaders.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#2; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md; docs/social001_vonc_tiktok_social/tool_docs/PLAN_lobby-grid(3).md
- **relations:** assembler visible-content filter; js_snippets bundling; generation-time guards; runtime-fill guards in discovery checks; Phase-3 pipeline (the missing data producer)
- **verify-later:** vonc.com index/archive shells + /data/provocations.json; rerender_single_page_action.go exemption

<!-- SOURCE: U25_leopardess_social.md -->
### js_snippets library + site-asset-renderer bundling (Path 2 JS delivery)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** VERDICT Q5 (2026-07-09): "Direct SQL is the only writer … render_js_snippets_for_site is a pure reader"; live bundle header "3 active snippet(s)".
- **what:** Library JS ships as js_snippets rows (name, js_content, applies_to array); site-asset-renderer selects active snippets whose applies_to overlaps the site's component functions, concatenates into /assets/js/snippets.js and git-commits. Direct SQL is the sanctioned writer (the generated banner says so); the table has no site_id (snippets are global; each loader self-checks for its section) and no updated_at column. Known gap: the bundle is NOT auto-re-rendered when a snippet changes (only at initial design/full rerender) — manual trigger required (js-bundle-stale, FX-6).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#1; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-29; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#7
- **relations:** runtime-fill mechanism; Path-1 extraction; tool-pipeline JS conventions
- **verify-later:** render_js_snippets_for_site_action.go; js_snippets schema; snippet-change triggers (expect none)

<!-- SOURCE: U25_leopardess_social.md -->
### Path-1 inline-JS extraction (component js_content) and the truncation bug class
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** NOTES_lobby-grid(6) 2026-07-09: "the extraction pattern … is already live for gauntlet-interface, latest-news and tool-archetype-taster-quiz"; provocation-card/lobby-grid inline scripts still unextracted.
- **what:** The architecturally-preferred JS home: a component's inline `<script>` is extracted (separateInlineJS) to content_components.js_content and served as /tools/assets/{function}.js, auto-deploying on rerender (rerender_single_page injects js_content at assembly). Bug class discovered: components stored via paths predating extraction keep raw inline scripts with empty js_content; one template was truncated mid-script at generation (token limit) because store validation checks unclosed `<style>` but not `<script>` — hardening items: script-balance check at store time, warn on unterminated scripts.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-29 entries; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.2
- **relations:** js_snippets Path 2; Mode-B templates; generation-time guards (no-inline-script rule makes the class impossible)
- **verify-later:** separateInlineJS in store_generated_component; js_content coverage across components

<!-- SOURCE: U25_leopardess_social.md -->
### Section descriptor + loader-builder agent + Tier E runtime-feed source (autonomy design)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** HANDOFF §9.6: "Autonomy design (PLAN_dynamic_sections_and_loaders.md): section descriptor {role, kind, data_feed} … a loader-builder agent … The two hand-built loaders are its reference implementations." (Referenced plan lives outside this unit.)
- **what:** The path from hand-built runtime-fill to autonomous: plans carry a section descriptor declaring role/kind/data_feed; component-creator gains a Tier E schema source (`source: "feed.{name}"`) that emits stable-selector shells plus a loader; a loader-builder agent (sibling of tool-generator, which forbids fetch) writes the fetch-and-fill IIFE from the DOM contract. The lobby-grid and archive-list hand builds are explicitly the reference implementations.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#Gap-2
- **relations:** runtime-fill mechanism; component creation contract; Phase-3 pipeline
- **verify-later:** PLAN_dynamic_sections_and_loaders.md (other unit); tool-generator fetch prohibition

<!-- SOURCE: U25_leopardess_social.md -->
### Generation-time guards for dynamic components (the archive-list reference build)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** SPEC_provocations-archive-list AS BUILT (2026-07-09): "Both generation-time guards held on the first attempt: has_marker = t, has_inline_script = f … first live validation."
- **what:** Bake the lessons into generation instead of repairing after: instruct component-creator to emit data-runtime-fill in the section tag (no post-hoc marker SQL), forbid `<script>` elements entirely (extraction/truncation class impossible), make header copy llm-sourced (nothing can defer), use a single hidden clone-template item for variable-length lists with an explicit `[data-…-template] { display:none; }` rule (the `hidden` attribute loses to author display rules), and include a visible empty state so the page ships before data lands. provocations-archive-list (70d6662a) is the canonical reference build.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3
- **relations:** component creation contract; runtime-fill mechanism; editing-stored-HTML landmines
- **verify-later:** component 70d6662a row; component-creator description patterns

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dynamic applications direction (three tiers; framework specs; thin generated backends)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** 022 tier 1 "now → near term", tiers 2–3 medium/longer term
- **what:** Tier 1 static+dynamic components (forms via external services, client search, client-side A/B); Tier 2 agent-powered per-site backends (workers/lightweight services fed by agents — business logic stays in agents, backend is a thin render layer); Tier 3 full application generation (admin panels, SaaS prototypes). Principles: framework specs stored for each target stack; one site one repo one deployment; generated-vs-human content marked, human edits precedent; incremental complexity (mailto → Formspree → Worker → CRM).
- **sources:** 022 full
- **relations:** infrastructure layers (007); CSS variable contract (shared)
- **verify-later:** none built beyond tier 1 basics

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Interactive fingerprint parse stage (C1–C6)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** "C1 — extract_interactive_fingerprint (new action). Status: in progress, 2026-05-15"; C2–C6 planned
- **what:** New Go extractor over crawled rawHtml capturing canvas elements, inline/external scripts, event handlers, forms, library signals (rAF, canvas contexts, jQuery/Three/Phaser/React/Vue) and a per-page type_hint heuristic (calculator/game_or_animation/interactive_widget/static); then external-JS fetch loop (C3), enrich (C4), LLM interactive_intent brief with feasibility markers (C5), written to new interactive_reference/interactive_intent spec aspects (C6). Deliberately a new file, not an extension of the design extractor; AST parsing out of scope.
- **sources:** FOCUS_interactive_content_generation(4).md#Path-C
- **relations:** design fingerprint pattern; capability markers; Firecrawl executeJavascript escalation
- **verify-later:** extract_interactive_fingerprint_action.go existence; interactive_reference aspects in site_specs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Four-stage interactive-content pattern (parse / assess / generate / integrate)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Not a roadmap — a map of the territory" (family v4, updated through 2026-05-15); tools implement it "mostly", minus the parse stage
- **what:** The reference shape for handling any interactive content type encountered on adopted sites: parse the source, assess producibility (producible_now / producible_simpler / blocked per the 028 spec model with feasibility-recheck promotion), generate the artefact, integrate into the build pipeline. Agreed sequencing: Path C (parse stage) → Path D (tool reliability — tools "currently don't work") → Path A (games) → B (news publishing) / E (numbered-component cleanup).
- **sources:** FOCUS_interactive_content_generation(4).md#four-stage, #Sequencing
- **relations:** tools pipeline; games gap; news publishing gap; capability assessment
- **verify-later:** feasibility-recheck task existence; state of tool reliability work

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Games as a content type (largest pipeline gap)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Games — nothing yet … page_type='game' doesn't exist in the classifier vocabulary" (2026-05); the vocabulary absence later CAUSED the 05-26 duplication bug
- **what:** No game-suggester/generator/improver/auditor, no game template library, no game_health check, no spec aspect; game-list components force fabrication. Plan: copy the tools pattern wholesale; add `game` to the page_type vocabulary (Option 1 hardcode now, Option 4 page_types table later — canonicalise kebab/snake first). The missing `game` type is not cosmetic: the planner re-typed adopted game pages to `tool`, driving rename + duplication.
- **sources:** FOCUS_interactive_content_generation(4).md#Games, #classification-vocabulary; HANDOFF_2026-05-26…md#diagnosis
- **relations:** page_type vocabulary gap; four-stage pattern; library model
- **verify-later:** plan_site Canonical Page Types list today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Generator architecture convergence (shared interactive-artefact-generator)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** "Worth considering once two more generators exist; one isn't enough to abstract from"
- **what:** Every content-type generator (tools, games, news articles, dashboards) needs a brief contract, prompt template, persistence action, page-creation step, tiered quality checks and companion-content step; a shared base with per-type specialisation is anticipated once games exist. The library model (canonical templates, forked_from IS NULL, per-site forks) is the copyable storage shape.
- **sources:** FOCUS_interactive_content_generation(4).md#Generator-architecture, #Library-model, #Quality-model
- **relations:** tools pipeline; games gap
- **verify-later:** n/a (design idea)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Interactive content generation (four-stage parse/assess/generate/integrate)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** FOCUS_interactive_content_generation(3) "Tools — most mature … Games — nothing yet"; sequencing "locked in 2026-05-14"
- **what:** A map for building tools/games/news/other interactive types via a four-stage pattern (parse the source, assess what's producible, generate the artefact, integrate into a page). Tools are most mature but missing a parse stage; games have nothing. Capability assessment is a spec-lifecycle property marking each element `producible_now`/`producible_simpler`/`blocked`.
- **sources:** WM/FOCUS_interactive_content_generation(3).md#the-four-stage-pattern, WM/FOCUS_interactive_content_generation(3).md#whats-working-today, WM/FOCUS_interactive_content_generation(3).md#capability-assessment
- **relations:** tool pipeline; component creator; spec-has-status; adoption parse-stage
- **verify-later:** tool-suggester/deployer/generator/improver/auditor; page_types vocabulary

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption parse-stage for interactive logic (interactive_reference/intent)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** FOCUS_interactive_content_generation(3) "1. Path C — Parse stage in adoption … Smaller piece of work than I first thought — closer to a couple of days"
- **what:** The prioritised gap: adoption captures markdown/design but not interactive JS. Closing it reuses the proven design-extraction shape: add `<script>`/`<canvas>` selectors to goquery, fetch `<script src>` via existing firecrawl_scrape, and add an LLM step producing `interactive_reference`/`interactive_intent` site_specs aspects.
- **sources:** WM/FOCUS_interactive_content_generation(3).md#where-the-parsing-capability-work-would-slot-in, WM/FOCUS_interactive_content_generation(3).md#sequencing-agreed-order, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** design fingerprint; interactive content generation; tool-recreation-handler misroute
- **verify-later:** extract_design_fingerprint_action.go; firecrawl_scrape; proposed extract_interactive_fingerprint

<!-- SOURCE: U21_legacy_docs_b.md -->
### Tool builder tiers (static / dynamic / application)
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** docs017/019b tier table ("Static: component library; Dynamic: self-contained JS, LLM-generated or pre-built; Application: engineer-built only") and platform stress test ("Agent-as-API pattern. User IS the HITL"); mortgagecalculator + website-design.com cited as early instances in docs018/008b.
- **what:** Interactive functionality classified by creation risk: static HTML components from the library; dynamic self-contained JS applications (calculators, visualisations) that LLMs may generate; full applications with API integration reserved for engineers. The agent-as-API pattern for platform sites treats the end user as the HITL. Matured into the tool-pipeline/tool-library/tool-lifecycle systems.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#6-Tool-Builder-Agent; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** tool-pipeline (successor); JavaScript management (docs017/023); finance/tools stress test.
- **verify-later:** current tool generation pipeline lineage.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Runtime-fill mechanism (data-runtime-fill shells + client loaders + JSON feed)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** HANDOFF §0: "That mechanism is now proven three times over" (provocation-card 2026-06-29, lobby-grid 2026-07-04, archive 2026-07-08, all browser-verified).
- **what:** Sections ship deliberately EMPTY at build time; the component `<section>` carries `data-runtime-fill="true"` so the assembler keeps the shell; an IIFE loader stored in `js_snippets` and bundled into `/assets/js/snippets.js` fetches `/data/provocations.json` in the visitor's browser and fills the shell's selectors, failing gracefully. Explicitly on-doctrine per doc 022 Tier 1 ("dynamic content injection... the dynamic part runs in the browser"; backend complexity lives in agents).
- **sources:** docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§2; docs/RUNBOOK_phase2_provocation_js(29).md#what-phase-2-is + #on-doctrine-check; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35; docs/README_summary_paragraph_for_handoff.md
- **relations:** visible-content filter exemption; js_snippets library; two JS delivery paths; static-vs-dynamic section distinction
- **verify-later:** rerender_single_page_action.go (reRuntimeFill regexp); js_snippets rows provocation-card-loader/lobby-grid-loader/provocations-archive-loader; /assets/js/snippets.js on vonc.com

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two JS delivery paths (Path 1 component js_content vs Path 2 js_snippets bundle)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~17:20 (PD-1 answered): latest-news fetches via its extracted component JS (Path 1); Path-2 loader proven live the same day; 2026-07-07 side-evidence: extraction pattern live on three tool components.
- **what:** Two separate JS delivery mechanisms exist. Path 1: a component's inline `<script>` is extracted to `content_components.js_content` and deployed as `/tools/assets/{function}.js` automatically on every page rerender (how gauntlet-interface, latest-news, archetype-quiz ship JS — including news's data fetch). Path 2: library `js_snippets` rows are bundled to `/assets/js/snippets.js` by site-asset-renderer — NOT part of the normal build/rerender flow. The vonc daily-feed shells "fell between" the two paths. PD-3 decided Path 1 is the durable home for fetch-and-fill loaders; the Path-2 snippets remain the live working interim.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00 + #2026-06-29-~17:20; docs/RUNBOOK_phase2_provocation_js(29).md#path-decision + #framework-fix; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-gate-passed
- **relations:** separateInlineJS extraction; js_snippets library; runtime-fill mechanism; js-bundle-stale gap
- **verify-later:** rerender_single_page_action.go collectJSAssets; content_components.js_content for gauntlet-interface/latest-news; /tools/assets/*.js in the sites repo

<!-- SOURCE: U23_docs_root_vonc.md -->
### js_snippets library + render_js_snippets_for_site + site-asset-renderer
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-29 ~17:20: renderer ran clean (commit eb7f2ac), snippet bundled; bundle header "3 active snippet(s)" 2026-07-07.
- **what:** `js_snippets` is a LIBRARY-WIDE table (no site_id) of JS behaviours keyed by `applies_to` (jsonb array of component functions). `render_js_snippets_for_site` selects active snippets whose applies_to overlaps the site's component functions, concatenates them (ordered by name, header comments, empty bundle still written so the head `<script src>` never 404s) into `/assets/js/snippets.js`, committed by the `site-asset-renderer` agent via git_commit. Loaders self-check for their section so a global snippet is inert on other sites. Pre-existing snippets were all small generic behaviours (accordion, scroll-reveal...); the fetch-and-fill loaders are a new, heavier use of the table.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~18:00 + #2026-06-25-~19:30 (inventory); docs/RUNBOOK_phase2_provocation_js(29).md#mechanism-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** two JS delivery paths; site-asset-renderer triggering gap; runtime-fill mechanism
- **verify-later:** render_js_snippets_for_site_action.go; site_asset_actions.go; js_snippets table contents

<!-- SOURCE: U23_docs_root_vonc.md -->
### js-bundle-stale gap (site-asset-renderer not wired into the build)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** FX-6 checkbox never ticked; RUNNING_NOTES 2026-06-29: "Gap 3 is NOT on the critical path... still a real latent issue for genuinely-generic snippets, but lower priority."
- **what:** Only `rerender-site` and `webdesign-agent` reference site-asset-renderer, so `/assets/js/snippets.js` is rebuilt at initial design and full site rerender but nothing re-runs it when a js_snippets row is added/changed later — the direct cause of the first loader never reaching vonc. Proposed fix: a design-discovery-agent check ("site has an applicable active snippet newer than its deployed bundle" → spawn site-asset-renderer). Deprioritised after PD-3 chose Path 1, never built; manual trigger scripts are the working practice.
- **sources:** docs/RUNBOOK_phase2_provocation_js(29).md#gap-3; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~19:00 (GAP 3 CONFIRMED) + #2026-06-29-~17:20
- **relations:** design-discovery-agent named checks; two JS delivery paths
- **verify-later:** design-discovery-agent run_discovery_checks array; agent_definitions referencing render_js_snippets

<!-- SOURCE: U23_docs_root_vonc.md -->
### loader-builder agent (fetch-and-fill sibling of tool-generator)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** PLAN_dynamic_sections gap 3; "the missing piece (provocation-card/lobby-grid loaders were hand-built)" (2026-07-04); the two hand-built loaders named as its reference implementations.
- **what:** A proposed agent that LLM-generates client-side fetch-and-fill loaders for dynamic sections: input = the section's DOM contract + feed shape; output = a graceful IIFE installed as a js_snippet and bundled by site-asset-renderer. Modelled on tool-generator (which LLM-generates, saves and wires SELF-CONTAINED tools) but necessarily a SIBLING because tool-generator explicitly forbids fetch. The framework currently has component-creator (section templates) and tool-generator (tools) but no runtime-fill loader builder.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** tool-generator (tool-pipeline); Tier E; provocation_card_loader.js / lobby_grid_loader.js / provocations_archive_loader.js as references
- **verify-later:** agent_definitions for tool-generator (no-fetch rule); absence of any loader-builder agent

<!-- SOURCE: U25_leopardess_social.md -->
### Runtime-fill mechanism (data-runtime-fill shells + client loaders)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** HANDOFF §0 (2026-07-09): "That mechanism is now proven three times over: the daily provocation card, the six-room arena grid, and … the Provocations Archive."
- **what:** vonc/Spark's central delivery mechanism for daily-changing content on a static site (doc 022 Tier 1): sections ship as deliberately empty shells whose `<section>` carries data-runtime-fill="true"; the page assembler's visible-content filter exempts marked sections; an IIFE loader fetches /data/provocations.json in the visitor's browser and fills the DOM contract, failing gracefully (shell + empty state remain). Distinction enforced against build-time content: static explainers get regenerated schemas and content-writer fills; only daily-dynamic shells get loaders.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#2; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md; docs/social001_vonc_tiktok_social/tool_docs/PLAN_lobby-grid(3).md
- **relations:** assembler visible-content filter; js_snippets bundling; generation-time guards; runtime-fill guards in discovery checks; Phase-3 pipeline (the missing data producer)
- **verify-later:** vonc.com index/archive shells + /data/provocations.json; rerender_single_page_action.go exemption

<!-- SOURCE: U25_leopardess_social.md -->
### js_snippets library + site-asset-renderer bundling (Path 2 JS delivery)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** VERDICT Q5 (2026-07-09): "Direct SQL is the only writer … render_js_snippets_for_site is a pure reader"; live bundle header "3 active snippet(s)".
- **what:** Library JS ships as js_snippets rows (name, js_content, applies_to array); site-asset-renderer selects active snippets whose applies_to overlaps the site's component functions, concatenates into /assets/js/snippets.js and git-commits. Direct SQL is the sanctioned writer (the generated banner says so); the table has no site_id (snippets are global; each loader self-checks for its section) and no updated_at column. Known gap: the bundle is NOT auto-re-rendered when a snippet changes (only at initial design/full rerender) — manual trigger required (js-bundle-stale, FX-6).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#1; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-29; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#7
- **relations:** runtime-fill mechanism; Path-1 extraction; tool-pipeline JS conventions
- **verify-later:** render_js_snippets_for_site_action.go; js_snippets schema; snippet-change triggers (expect none)

<!-- SOURCE: U25_leopardess_social.md -->
### Path-1 inline-JS extraction (component js_content) and the truncation bug class
- **category:** dynamic-applications
- **status-signal:** partial
- **status-evidence:** NOTES_lobby-grid(6) 2026-07-09: "the extraction pattern … is already live for gauntlet-interface, latest-news and tool-archetype-taster-quiz"; provocation-card/lobby-grid inline scripts still unextracted.
- **what:** The architecturally-preferred JS home: a component's inline `<script>` is extracted (separateInlineJS) to content_components.js_content and served as /tools/assets/{function}.js, auto-deploying on rerender (rerender_single_page injects js_content at assembly). Bug class discovered: components stored via paths predating extraction keep raw inline scripts with empty js_content; one template was truncated mid-script at generation (token limit) because store validation checks unclosed `<style>` but not `<script>` — hardening items: script-balance check at store time, warn on unterminated scripts.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-29 entries; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.2
- **relations:** js_snippets Path 2; Mode-B templates; generation-time guards (no-inline-script rule makes the class impossible)
- **verify-later:** separateInlineJS in store_generated_component; js_content coverage across components

<!-- SOURCE: U25_leopardess_social.md -->
### Section descriptor + loader-builder agent + Tier E runtime-feed source (autonomy design)
- **category:** dynamic-applications
- **status-signal:** aspirational
- **status-evidence:** HANDOFF §9.6: "Autonomy design (PLAN_dynamic_sections_and_loaders.md): section descriptor {role, kind, data_feed} … a loader-builder agent … The two hand-built loaders are its reference implementations." (Referenced plan lives outside this unit.)
- **what:** The path from hand-built runtime-fill to autonomous: plans carry a section descriptor declaring role/kind/data_feed; component-creator gains a Tier E schema source (`source: "feed.{name}"`) that emits stable-selector shells plus a loader; a loader-builder agent (sibling of tool-generator, which forbids fetch) writes the fetch-and-fill IIFE from the DOM contract. The lobby-grid and archive-list hand builds are explicitly the reference implementations.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#9.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_lobby-grid(6).md#2026-07-04; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#Gap-2
- **relations:** runtime-fill mechanism; component creation contract; Phase-3 pipeline
- **verify-later:** PLAN_dynamic_sections_and_loaders.md (other unit); tool-generator fetch prohibition

<!-- SOURCE: U25_leopardess_social.md -->
### Generation-time guards for dynamic components (the archive-list reference build)
- **category:** dynamic-applications
- **status-signal:** deployed
- **status-evidence:** SPEC_provocations-archive-list AS BUILT (2026-07-09): "Both generation-time guards held on the first attempt: has_marker = t, has_inline_script = f … first live validation."
- **what:** Bake the lessons into generation instead of repairing after: instruct component-creator to emit data-runtime-fill in the section tag (no post-hoc marker SQL), forbid `<script>` elements entirely (extraction/truncation class impossible), make header copy llm-sourced (nothing can defer), use a single hidden clone-template item for variable-length lists with an explicit `[data-…-template] { display:none; }` rule (the `hidden` attribute loses to author display rules), and include a visible empty state so the page ships before data lands. provocations-archive-list (70d6662a) is the canonical reference build.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3
- **relations:** component creation contract; runtime-fill mechanism; editing-stored-HTML landmines
- **verify-later:** component 70d6662a row; component-creator description patterns

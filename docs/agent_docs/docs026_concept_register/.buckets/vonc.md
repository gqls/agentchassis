
<!-- SOURCE: U23_docs_root_vonc.md -->
### Spark daily-provocation product (vonc.com)
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-09 §0: index + arena + archive live and browser-verified; but "the data file is currently hand-committed... a Phase-3 pipeline will emit it"; v1 roadmap features (daily_provocation_generation_from_scraping) not built.
- **what:** vonc.com / "Spark" — an AI daily-provocation platform: one charged provocation per day, users file a position, "the Gauntlet" scores the room, users get an Archetype. "The product IS the landing page": a single provocation card fills the screen; daily static regeneration; AI as producer (frames/scores/curates), not performer. v1 = daily provocations + Gauntlet; v3 concept = live challenge rooms. Serves as the platform's live test bed for the runtime-fill mechanism.
- **sources:** docs/PLAN_provocation-card(3).md#source-spec; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§0/§2; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~17:15; docs/RUNNING_NOTES_vonc_v2(28).md#carried-forward-state
- **relations:** runtime-fill mechanism; Phase-3 provocation data pipeline; provocation-card/lobby-grid/provocations-archive-list components
- **verify-later:** live vonc.com; sites row 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74; site_specs aspects (mission, roadmap, cta)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phase-3 provocation data pipeline (provocation-generator + orchestrator + render action + daily schedule)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** Phase-1 diagnostics confirmed "a clean slate — nothing exists yet" (2026-06-25); FX-4 checkbox never ticked; provocations.json still hand-committed as of 2026-07-09.
- **what:** The pipeline that would generate `/data/provocations.json` daily: clone the news pipeline — seed content_sources (trending-topic scraping targets) → reuse feed-ingester → NEW provocation-generator agent (LLM: raw topics → provocations + AI takes; generative analogue of feed-triage) → NEW render_provocations_section Go action (mirror of render_news_section; Go struct defines the JSON shape; returns a files map for git_commit) → provocation-orchestrator (clone of content-feed-orchestrator) → scheduled_tasks row `provocation-refresh` (daily; the column is `name`, not task_name). Open questions recorded: sources, volume per day, archive-page reads.
- **sources:** docs/PLAN_spark_provocation_pipeline.md; docs/RUNBOOK_phase2_provocation_js(29).md#data-deploy + #gap-1; docs/RUNBOOK_vonc_migrations(14).md#step-8
- **relations:** news feed pipeline (the template); provocations.json contract (the target shape); Spark product
- **verify-later:** absence of provocation-* agent_definitions/scheduled_tasks/content_sources for vonc; render_news_section_action.go as the model

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations.json data contract (today / lobby / arena / archive)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** v3 served live (curl grep '"archive"' = 1, 2026-07-07); all three loaders verified filling from it in a browser.
- **what:** The versioned feed contract for Spark's runtime-fill sections: `generated_at`; `today` {eyebrow, headline (may carry `<em>`), body, primary_cta/secondary_cta {label,url}, stats ×3}; `lobby` [4 × {icon,title,desc,url}] (becomes dead after the mini-lobby trim); `arena` — an OBJECT {eyebrow,title,subtitle,cta_label,cta,cards[≤6]} because the grid's header + CTA need data too; `archive` {entries[≤24] {date,title,teaser,stat,url}, newest-first}. Evolved v1→v3 in provocations.sample.json; hand-committed interim, the fixed generation target for Phase 3.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35 (confirmed-good shape); docs/PLAN_lobby-grid(2).md#build-progress; docs/SPEC_provocations-archive-list.md#data-contract; docs/provocations.sample(3).json
- **relations:** the three loaders; Phase-3 pipeline; runtime-fill mechanism
- **verify-later:** live vonc.com/data/provocations.json keys

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocation-card component (daily hero card) + mini-lobby trim
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** "Live and working via Path-2 loader" (PLAN status); trim CONFIRMED 2026-07-04, drafted 2026-07-09, blocked on the bundle verdict — not executed within this corpus.
- **what:** The Spark centrepiece: single daily contested claim + AI take + 3 stats + 2 CTAs + (currently) a 4-card mini-lobby, filled at runtime from `today`/`lobby` by provocation-card-loader against the `.pc-*` DOM contract. JS-required by design — do NOT "fix" by baking content. Known limitation: the underlying template is Mode-B broken (loader masks it; JS-off shows `<no value>`). NEXT TASK: trim the mini-lobby (template pc-card-grid block, loader lobby fill, the orphaned 1fr-1fr media query, the dead hover script) because lobby-grid owns the arena role — with the method itself under a bundle verdict since HTML patching is the rejected mechanism.
- **sources:** docs/PLAN_provocation-card(3).md; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:20; docs/provocation_card_loader.js (header)
- **relations:** lobby-grid overlap decision; sanctioned edit paths; runtime-fill mechanism
- **verify-later:** content_components 6163ff14 html_template (pc-card-grid still present?); js_snippets provocation-card-loader lobby block

<!-- SOURCE: U23_docs_root_vonc.md -->
### lobby-grid arena component (six-room grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "lobby-grid DONE (browser-verified)" 2026-07-04 — six arena cards + pulsing stat dots + "Enter the Arena" live; PLAN_lobby-grid marked DELIVERED 2026-07-09.
- **what:** The Arena lobby: 6-card grid (1 featured spanning 2 cols, 4 standard, 1 wide), each card icon (SVG inner markup with emoji fallback)/tag/title/desc/stat + pulsing dot, plus header and CTA — filled at runtime from `arena` by lobby-grid-loader. Honest v1 semantics decision: "live rooms" is a v3 concept, so in v1 the grid shows TODAY'S PROVOCATIONS as enterable cards. Confirmed decisions: lobby-grid is the primary "today's provocations grid" (D-A) with the `arena` object as feed (D-B). Its build was deliberately the reference implementation for the loader-builder design.
- **sources:** docs/PLAN_lobby-grid(2).md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-lobby-grid-verified; docs/lobby_grid_loader.js (header); docs/lobby_grid_install.sql (header)
- **relations:** provocation-card mini-lobby trim; loader-builder reference; marker REPLACE anchoring incident
- **verify-later:** js_snippets lobby-grid-loader; live index data-component="lobby-grid"

<!-- SOURCE: U23_docs_root_vonc.md -->
### brief-explanation static explainer (regeneration, not a loader)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083 succeeded 2026-07-01 (in-place update, quality 50→100, 0→20 fields); rendered with real copy on the live index 2026-07-03.
- **what:** The "what is Spark / how it works" index explainer — STABLE brand content (eyebrow, heading with `<mark>`, description, exactly 3 numbered steps, 3 stats, 2 CTAs, illustration+badge) that belongs in build-time HTML for SEO and no-JS robustness. Establishes the key distinction: Option-2 runtime loaders are ONLY for daily-changing data shells; static shells that happen to be empty are fixed by REGENERATION with a real schema — two different resolutions for the same empty-shell symptom. Its stat fields were later re-sourced static→llm to stop generic SaaS fallbacks leaking.
- **sources:** docs/PLAN_brief-explanation(1).md; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~18:00 + #2026-07-01-~12:46; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-static-field-discrepancy
- **relations:** static-vs-dynamic distinction; shared component library (58363894 shared ×3 sites); component regeneration in place
- **verify-later:** content_components 58363894 field sources; idea.uk/robot-hands pending instances

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations-archive-list component + provocations archive page
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "PROVOCATIONS-INDEX THREAD DONE" 2026-07-08: page live, 8 rows fill, ghost row eliminated; live confirm grep = 2 on 2026-07-09.
- **what:** The Provocations Archive at /provocations/index.html — destination of every primary CTA — as a single self-contained runtime-fill section: llm header fields (nothing can defer), a hidden clone-template row the loader clones per `archive.entries[]` (variable-length list vs lobby-grid's fixed six), a visible empty state so the page ships before data lands, CTA back to today. Built via the full arc: component (70d6662a, 084 trigger) → plan row → pages.sections unblock → first real build (~5 min after ten 33–65s no-ops) → loader + data → ghost-row CSS fix.
- **sources:** docs/SPEC_provocations-archive-list.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/RUNBOOK_phase2_provocation_js(29).md#you-are-here; docs/provocations_archive_loader.js (header)
- **relations:** complete_error family (its 404 was the trigger); generation-time guards (first live validation); CTA graph
- **verify-later:** pages e4b3b195 build_status; live /provocations/index.html

<!-- SOURCE: U23_docs_root_vonc.md -->
### Option 1 — build-time static content for the daily shells (rejected alternative)
- **category:** vonc
- **status-signal:** abandoned
- **status-evidence:** Early migrations-runbook versions carried "Recommendation: Option 1 for the first deployable version — get real content" (dropped line); final: "DECISION MADE: Option 2... Option 1 would freeze a single set of provocations permanently, defeating the daily-content product."
- **what:** The rejected fix for the empty index shells: regenerate them WITH proper input_schemas so the content writer fills them at build time. Briefly the recommended first-version route in early runbook versions, then dropped when the original Spark roadmap (daily provocations via client-side JS) was recovered — build-time content would bake one day's provocations permanently. Survives only in its correct form: genuinely static shells (brief-explanation) ARE fixed by regeneration.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#step-7 (decision); early-version dropped lines (family diff); docs/PLAN_spark_provocation_pipeline.md#why-option-2
- **relations:** static-vs-dynamic distinction; brief-explanation (where Option-1 logic is right)
- **verify-later:** none (historical)

<!-- SOURCE: U25_leopardess_social.md -->
### Phase-3 provocation pipeline (automated provocations.json emission)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-11: "There is no Phase-3 emitter yet; all prior commits to the file were hand-made."
- **what:** The missing producer for the runtime-fill economy: a provocation-orchestrator + scheduled refresh generating /data/provocations.json ({generated_at, today, arena, archive}) daily from the scraping/framing engine, replacing hand-committed sample data. The dead `lobby` key was dropped 2026-07-11 (commit c244ddc) after the mini-lobby trim. Until it exists, vonc's "daily" provocation is static.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/PLAN_provocation-card(4).md#Data-contract; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Data-contract; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11
- **relations:** provocation engine (the design it implements); runtime-fill mechanism; scheduler-and-tasks
- **verify-later:** agent_definitions for any provocation-orchestrator; scheduled_tasks; sites repo /data/provocations.json history

<!-- SOURCE: U25_leopardess_social.md -->
### vonc.com Spark v1 site (the live testbed)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083_update_work_items.sh lists the 8 live pages; HANDOFF §3 "Index page — live, six sections … Provocations archive — CLOSED 2026-07-08"; archetype hub live 2026-07-12.
- **what:** The built v1: index (hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid, system-stats), /provocations/index.html archive, about, contact, archetypes hub + 8 entity pages, blog/provocation, and two tools (gauntlet, archetype-taster-quiz). Serves as the platform's live test bed for runtime-fill, component generation, discovery checks and the section-editor; "the landing page IS the product — a provocation card, not a marketing page".
- **sources:** docs/social001_vonc_tiktok_social/trigger_script/083_update_work_items.sh; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0
- **relations:** content-first launch; runtime-fill; archetype hub; site id 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74
- **verify-later:** live vonc.com pages; pages table for the site

<!-- SOURCE: U25_leopardess_social.md -->
### Archetype hub built with existing machinery (entity pages + query-resolved grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-12: "End state, live-verified: archetypes.html shows 8 cards … all 8 detail pages HTTP 200, each with its icon."
- **what:** Fix for a page that had "rendered zero archetypes": archetype-grid is build-time query-resolved (items source query.pages_where_type) — a third content mode beside static and runtime-fill — and its page_type value was kebab-forbidden (chk_page_type_kebab_case) with zero matching pages. Approach A created 8 site_plan_pages (role entity-page), 24 plan sections, 8 page-scope site_plan_imagery hero rows consuming the 8 orphaned icon assets via kind-alias resolution, plus 8 pages rows (page-build-handler loads pages, never creates them), then reconcile_site_plan emitted the builds. 089 re-authored generic writer copy from the spec's archetype canon via content_data (light no-LLM rerender).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header); 089_archetype_page_copy.sql (header); docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12
- **relations:** site-plan-and-reconciler; behavioural archetype system; illustration/section-imagery resolution
- **verify-later:** pages page_type='entity-page' rows; archetype-grid input_schema source; chk_page_type_kebab_case

<!-- SOURCE: U23_docs_root_vonc.md -->
### Spark daily-provocation product (vonc.com)
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-09 §0: index + arena + archive live and browser-verified; but "the data file is currently hand-committed... a Phase-3 pipeline will emit it"; v1 roadmap features (daily_provocation_generation_from_scraping) not built.
- **what:** vonc.com / "Spark" — an AI daily-provocation platform: one charged provocation per day, users file a position, "the Gauntlet" scores the room, users get an Archetype. "The product IS the landing page": a single provocation card fills the screen; daily static regeneration; AI as producer (frames/scores/curates), not performer. v1 = daily provocations + Gauntlet; v3 concept = live challenge rooms. Serves as the platform's live test bed for the runtime-fill mechanism.
- **sources:** docs/PLAN_provocation-card(3).md#source-spec; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§0/§2; docs/RUNNING_NOTES_vonc(36).md#2026-06-25-~17:15; docs/RUNNING_NOTES_vonc_v2(28).md#carried-forward-state
- **relations:** runtime-fill mechanism; Phase-3 provocation data pipeline; provocation-card/lobby-grid/provocations-archive-list components
- **verify-later:** live vonc.com; sites row 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74; site_specs aspects (mission, roadmap, cta)

<!-- SOURCE: U23_docs_root_vonc.md -->
### Phase-3 provocation data pipeline (provocation-generator + orchestrator + render action + daily schedule)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** Phase-1 diagnostics confirmed "a clean slate — nothing exists yet" (2026-06-25); FX-4 checkbox never ticked; provocations.json still hand-committed as of 2026-07-09.
- **what:** The pipeline that would generate `/data/provocations.json` daily: clone the news pipeline — seed content_sources (trending-topic scraping targets) → reuse feed-ingester → NEW provocation-generator agent (LLM: raw topics → provocations + AI takes; generative analogue of feed-triage) → NEW render_provocations_section Go action (mirror of render_news_section; Go struct defines the JSON shape; returns a files map for git_commit) → provocation-orchestrator (clone of content-feed-orchestrator) → scheduled_tasks row `provocation-refresh` (daily; the column is `name`, not task_name). Open questions recorded: sources, volume per day, archive-page reads.
- **sources:** docs/PLAN_spark_provocation_pipeline.md; docs/RUNBOOK_phase2_provocation_js(29).md#data-deploy + #gap-1; docs/RUNBOOK_vonc_migrations(14).md#step-8
- **relations:** news feed pipeline (the template); provocations.json contract (the target shape); Spark product
- **verify-later:** absence of provocation-* agent_definitions/scheduled_tasks/content_sources for vonc; render_news_section_action.go as the model

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations.json data contract (today / lobby / arena / archive)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** v3 served live (curl grep '"archive"' = 1, 2026-07-07); all three loaders verified filling from it in a browser.
- **what:** The versioned feed contract for Spark's runtime-fill sections: `generated_at`; `today` {eyebrow, headline (may carry `<em>`), body, primary_cta/secondary_cta {label,url}, stats ×3}; `lobby` [4 × {icon,title,desc,url}] (becomes dead after the mini-lobby trim); `arena` — an OBJECT {eyebrow,title,subtitle,cta_label,cta,cards[≤6]} because the grid's header + CTA need data too; `archive` {entries[≤24] {date,title,teaser,stat,url}, newest-first}. Evolved v1→v3 in provocations.sample.json; hand-committed interim, the fixed generation target for Phase 3.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:35 (confirmed-good shape); docs/PLAN_lobby-grid(2).md#build-progress; docs/SPEC_provocations-archive-list.md#data-contract; docs/provocations.sample(3).json
- **relations:** the three loaders; Phase-3 pipeline; runtime-fill mechanism
- **verify-later:** live vonc.com/data/provocations.json keys

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocation-card component (daily hero card) + mini-lobby trim
- **category:** vonc
- **status-signal:** partial
- **status-evidence:** "Live and working via Path-2 loader" (PLAN status); trim CONFIRMED 2026-07-04, drafted 2026-07-09, blocked on the bundle verdict — not executed within this corpus.
- **what:** The Spark centrepiece: single daily contested claim + AI take + 3 stats + 2 CTAs + (currently) a 4-card mini-lobby, filled at runtime from `today`/`lobby` by provocation-card-loader against the `.pc-*` DOM contract. JS-required by design — do NOT "fix" by baking content. Known limitation: the underlying template is Mode-B broken (loader masks it; JS-off shows `<no value>`). NEXT TASK: trim the mini-lobby (template pc-card-grid block, loader lobby fill, the orphaned 1fr-1fr media query, the dead hover script) because lobby-grid owns the arena role — with the method itself under a bundle verdict since HTML patching is the rejected mechanism.
- **sources:** docs/PLAN_provocation-card(3).md; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~17:20; docs/provocation_card_loader.js (header)
- **relations:** lobby-grid overlap decision; sanctioned edit paths; runtime-fill mechanism
- **verify-later:** content_components 6163ff14 html_template (pc-card-grid still present?); js_snippets provocation-card-loader lobby block

<!-- SOURCE: U23_docs_root_vonc.md -->
### lobby-grid arena component (six-room grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "lobby-grid DONE (browser-verified)" 2026-07-04 — six arena cards + pulsing stat dots + "Enter the Arena" live; PLAN_lobby-grid marked DELIVERED 2026-07-09.
- **what:** The Arena lobby: 6-card grid (1 featured spanning 2 cols, 4 standard, 1 wide), each card icon (SVG inner markup with emoji fallback)/tag/title/desc/stat + pulsing dot, plus header and CTA — filled at runtime from `arena` by lobby-grid-loader. Honest v1 semantics decision: "live rooms" is a v3 concept, so in v1 the grid shows TODAY'S PROVOCATIONS as enterable cards. Confirmed decisions: lobby-grid is the primary "today's provocations grid" (D-A) with the `arena` object as feed (D-B). Its build was deliberately the reference implementation for the loader-builder design.
- **sources:** docs/PLAN_lobby-grid(2).md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-lobby-grid-verified; docs/lobby_grid_loader.js (header); docs/lobby_grid_install.sql (header)
- **relations:** provocation-card mini-lobby trim; loader-builder reference; marker REPLACE anchoring incident
- **verify-later:** js_snippets lobby-grid-loader; live index data-component="lobby-grid"

<!-- SOURCE: U23_docs_root_vonc.md -->
### brief-explanation static explainer (regeneration, not a loader)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083 succeeded 2026-07-01 (in-place update, quality 50→100, 0→20 fields); rendered with real copy on the live index 2026-07-03.
- **what:** The "what is Spark / how it works" index explainer — STABLE brand content (eyebrow, heading with `<mark>`, description, exactly 3 numbered steps, 3 stats, 2 CTAs, illustration+badge) that belongs in build-time HTML for SEO and no-JS robustness. Establishes the key distinction: Option-2 runtime loaders are ONLY for daily-changing data shells; static shells that happen to be empty are fixed by REGENERATION with a real schema — two different resolutions for the same empty-shell symptom. Its stat fields were later re-sourced static→llm to stop generic SaaS fallbacks leaking.
- **sources:** docs/PLAN_brief-explanation(1).md; docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~18:00 + #2026-07-01-~12:46; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-static-field-discrepancy
- **relations:** static-vs-dynamic distinction; shared component library (58363894 shared ×3 sites); component regeneration in place
- **verify-later:** content_components 58363894 field sources; idea.uk/robot-hands pending instances

<!-- SOURCE: U23_docs_root_vonc.md -->
### provocations-archive-list component + provocations archive page
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** "PROVOCATIONS-INDEX THREAD DONE" 2026-07-08: page live, 8 rows fill, ghost row eliminated; live confirm grep = 2 on 2026-07-09.
- **what:** The Provocations Archive at /provocations/index.html — destination of every primary CTA — as a single self-contained runtime-fill section: llm header fields (nothing can defer), a hidden clone-template row the loader clones per `archive.entries[]` (variable-length list vs lobby-grid's fixed six), a visible empty state so the page ships before data lands, CTA back to today. Built via the full arc: component (70d6662a, 084 trigger) → plan row → pages.sections unblock → first real build (~5 min after ten 33–65s no-ops) → loader + data → ghost-row CSS fix.
- **sources:** docs/SPEC_provocations-archive-list.md; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-08; docs/RUNBOOK_phase2_provocation_js(29).md#you-are-here; docs/provocations_archive_loader.js (header)
- **relations:** complete_error family (its 404 was the trigger); generation-time guards (first live validation); CTA graph
- **verify-later:** pages e4b3b195 build_status; live /provocations/index.html

<!-- SOURCE: U23_docs_root_vonc.md -->
### Option 1 — build-time static content for the daily shells (rejected alternative)
- **category:** vonc
- **status-signal:** abandoned
- **status-evidence:** Early migrations-runbook versions carried "Recommendation: Option 1 for the first deployable version — get real content" (dropped line); final: "DECISION MADE: Option 2... Option 1 would freeze a single set of provocations permanently, defeating the daily-content product."
- **what:** The rejected fix for the empty index shells: regenerate them WITH proper input_schemas so the content writer fills them at build time. Briefly the recommended first-version route in early runbook versions, then dropped when the original Spark roadmap (daily provocations via client-side JS) was recovered — build-time content would bake one day's provocations permanently. Survives only in its correct form: genuinely static shells (brief-explanation) ARE fixed by regeneration.
- **sources:** docs/RUNBOOK_vonc_migrations(14).md#step-7 (decision); early-version dropped lines (family diff); docs/PLAN_spark_provocation_pipeline.md#why-option-2
- **relations:** static-vs-dynamic distinction; brief-explanation (where Option-1 logic is right)
- **verify-later:** none (historical)

<!-- SOURCE: U25_leopardess_social.md -->
### Phase-3 provocation pipeline (automated provocations.json emission)
- **category:** vonc
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-11: "There is no Phase-3 emitter yet; all prior commits to the file were hand-made."
- **what:** The missing producer for the runtime-fill economy: a provocation-orchestrator + scheduled refresh generating /data/provocations.json ({generated_at, today, arena, archive}) daily from the scraping/framing engine, replacing hand-committed sample data. The dead `lobby` key was dropped 2026-07-11 (commit c244ddc) after the mini-lobby trim. Until it exists, vonc's "daily" provocation is static.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/PLAN_provocation-card(4).md#Data-contract; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Data-contract; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11
- **relations:** provocation engine (the design it implements); runtime-fill mechanism; scheduler-and-tasks
- **verify-later:** agent_definitions for any provocation-orchestrator; scheduled_tasks; sites repo /data/provocations.json history

<!-- SOURCE: U25_leopardess_social.md -->
### vonc.com Spark v1 site (the live testbed)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** 083_update_work_items.sh lists the 8 live pages; HANDOFF §3 "Index page — live, six sections … Provocations archive — CLOSED 2026-07-08"; archetype hub live 2026-07-12.
- **what:** The built v1: index (hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid, system-stats), /provocations/index.html archive, about, contact, archetypes hub + 8 entity pages, blog/provocation, and two tools (gauntlet, archetype-taster-quiz). Serves as the platform's live test bed for runtime-fill, component generation, discovery checks and the section-editor; "the landing page IS the product — a provocation card, not a marketing page".
- **sources:** docs/social001_vonc_tiktok_social/trigger_script/083_update_work_items.sh; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0
- **relations:** content-first launch; runtime-fill; archetype hub; site id 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74
- **verify-later:** live vonc.com pages; pages table for the site

<!-- SOURCE: U25_leopardess_social.md -->
### Archetype hub built with existing machinery (entity pages + query-resolved grid)
- **category:** vonc
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-12: "End state, live-verified: archetypes.html shows 8 cards … all 8 detail pages HTTP 200, each with its icon."
- **what:** Fix for a page that had "rendered zero archetypes": archetype-grid is build-time query-resolved (items source query.pages_where_type) — a third content mode beside static and runtime-fill — and its page_type value was kebab-forbidden (chk_page_type_kebab_case) with zero matching pages. Approach A created 8 site_plan_pages (role entity-page), 24 plan sections, 8 page-scope site_plan_imagery hero rows consuming the 8 orphaned icon assets via kind-alias resolution, plus 8 pages rows (page-build-handler loads pages, never creates them), then reconcile_site_plan emitted the builds. 089 re-authored generic writer copy from the spec's archetype canon via content_data (light no-LLM rerender).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql (header); 089_archetype_page_copy.sql (header); docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12
- **relations:** site-plan-and-reconciler; behavioural archetype system; illustration/section-imagery resolution
- **verify-later:** pages page_type='entity-page' rows; archetype-grid input_schema source; chk_page_type_kebab_case

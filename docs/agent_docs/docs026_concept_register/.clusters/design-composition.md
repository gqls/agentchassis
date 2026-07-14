# Cluster: design-composition
Categories included: design-composition


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Design system three layers (content_components / css_themes / style_collections)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4) opening section; live tables
- **what:** Layer 1 HTML components (self-contained blocks, inline style, CSS variables with fallbacks, never hardcode brand colours); Layer 2 CSS theme (one styles.css per site rendered from installed composition); Layer 3 style_collections bundling header/footer components + theme + palette/typography. Sites reference via sites.style_collection_id.
- **sources:** 002(4)#Design System Layers; 003(8) contracts
- **relations:** palette/layout/typography migration (025); composition (027)
- **verify-later:** content_components, css_themes, style_collections schemas

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Composition three stages: direction → composition → execution
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4) and 027 describe the deployed reorder ("Applied"); install_theme removed from webdesign-agent
- **what:** (1) domain-research-classifier writes design_intent (structured palette/typography reference_values + style_direction scheme). (2) site-design-planner (deterministic, no LLM, `needs_composition` item) resolves layout (weighted scheme-aware tag match), typography (match-or-insert), palette (always new site-specific row) via signal cascades, then install_site_composition atomically writes css_themes+style_collections+sites pointer+resolved_composition spec (a decision record, not CSS; refuses overwrite). (3) webdesign-agent (needs_design, depends_on composition) renders and commits styles.css — sole writer.
- **sources:** 002(4)#Composition; 027 full; 025 (schema underneath)
- **relations:** scheme-aware matcher; renderer cascade; superseding-spec-doesn't-undo-install failure mode (028)
- **verify-later:** fork_theme_composition.go resolvers; install_site_composition; needs_composition/needs_design ordering

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Scheme-aware weighted layout matcher + needs_new_layout_candidate HITL signal
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4); 016 §9 v2_55 records its half-merged-deploy failure and fix shipped as one merged file
- **what:** Layout matching weights tags by rarity with category/description bonuses; the site's scheme (from design_intent.style_direction) is a near-hard constraint (light site never placed on dark layout while any non-dark fits). On fallback it queues `needs_new_layout_candidate` (status needs_human_review, skipped by dispatch) — the honest "library is missing a layout" signal. layouts.scheme nullable → incremental curation.
- **sources:** 002(4)#Composition; 027 §2; 016 §9 scheme-matcher entry
- **relations:** library growth; section-contrast open question (036)
- **verify-later:** resolveLayoutByTagsWeighted; layouts.scheme population

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Theme/layout library growth and the fork-with-review gate
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 003(8) forking rules deployed; 002(3)'s auto "search→reuse→generate→store" loop dropped from 002(4) (superseded by curated-library stance); LLM layout matching "a future step"
- **what:** Layouts are a curated shared grammar (no auto-generated bespoke layout per site). Growth: hand-added variants, or HITL route — ForkThemeFromSiteAction promotes a rendered design into css_themes+style_collections with needs_review=true and a needs_theme_review item; selectors must exclude needs_review rows; rejection only affects future sites. Lineage columns (origin/forked_from/source_site/forked_at) on themes and collections.
- **sources:** 002(4)#Library growth; 003(8)#CSS Theme Template Contract (lineage, review gate, forking rules)
- **relations:** 025 migration lineage model on palettes/layouts/typography_sets
- **verify-later:** fork_theme_from_site_action.go; needs_review filtering in selectors

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Superseded: single webdesign-agent brand+CSS role and the brand-designer/layout-architect/style-generator split
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** dropped between 002(3) and 002(4); 002(4): "The earlier 'one agent generates brand + CSS' shape is superseded by the composition/execution split"
- **what:** Earlier architecture had webdesign-agent doing brand analysis + CSS with a planned future split into brand-designer/layout-architect/style-generator, and an auto theme-library reuse loop. Replaced by site-design-planner (composition) + webdesign-agent (render); a finer split deferred until search-and-adapt clearly beats render-from-composition.
- **sources:** 002_system_architecture(3).md (family-delta); 002(4)#Design Agent Family
- **relations:** composition three stages
- **verify-later:** no brand-designer/layout-architect agent_definitions rows expected

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Palette/layout/typography composable-theme migration (025)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 025(3) is the migration plan; 027/036 confirm palettes/layouts/typography_sets live and read by renderer; legacy css_themes columns retained "until Phase 7"; 036 notes Phase 4.5 coupling still present
- **what:** Split css_themes into palettes (colours jsonb open slot map), layouts (css_template + structure_tokens + default header/footer FKs + scheme), typography_sets (fonts+scale), each with the origin/needs_review/fork lineage model; css_themes becomes a composition of three FKs. Motivation: the old library was one layout with 14 palette skins behind a silent standard-brochure fallback. Template data moves to map-based Palette/Typography/Structure (no Go change per new slot). Direct cutover, no shadow mode; selector unchanged in this phase.
- **sources:** 025(3) full; 036 §3
- **relations:** composition stages; layout scheme matcher
- **verify-later:** legacy columns still read anywhere; Phase 4.5/7 progress

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Renderer theme-resolution cascade and the emergency fallback
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 027 §4: theme_name literal cleared "as the cutover moment"; emergency fallback + logger.Error monitoring rule
- **what:** render_css_from_spec resolves theme by config.theme_id → config.theme_name → sites.style_collection_id join (production path); all-miss falls to standard-brochure WITH logger.Error — any emergency-fallback line is a pipeline bug. resolveThemeIDFromSiteContext never errors, warns with a distinguishing reason.
- **sources:** 027#Renderer Changes
- **relations:** install-before-render ordering; B-029-3 theme-vars-not-deployed bug
- **verify-later:** emergency fallback frequency in logs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Build-time design/imagery trigger emission (Gap A)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "deployed step order in v1.0.1047 is read_specs → … → reconcile_site_plan → emit_design → emit_imagery → complete … So Gap A is closed on both fresh-build and adoption paths" (2026-05-26, verified on gamesdesign cascade)
- **what:** `emit_design_items` and `emit_imagery_items` (shared `imageryplan` package) wired as plan-time steps in build-site-planner, closing the long-standing gap where composition and imagery items were never emitted after the Phase-1 refactor moved the terminal step away from WriteBuildItemsAction. Nine needs_imagery items at documented priority bands (65 index-hero, 70 site-logo, 75/80 others, 98 clamped section-scope) observed live.
- **sources:** HANDOFF_2026-05-26_design_imagery_triggers_and_adoption_diagnosis.md#What-deployed
- **relations:** site-design-planner; imagery loop closure; site_plan_imagery
- **verify-later:** build-site-planner workflow steps in agent_definitions; imageryplan package

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site-design-planner composition resolver (composition before render)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "First successful composition run … completed cleanly" (2026-04-20 gamedesign.uk with all five IDs); "needs_composition ran via site-design-planner and install_site_composition populated sites.style_collection_id" (2026-05-26)
- **what:** A handler agent (needs_composition work item) that resolves layout (deterministic tag-overlap against layouts.industry_tags), typography (font-family/character match) and palette (fingerprint → mission → design_intent priority) BEFORE webdesign-agent renders, installing css_themes + style_collections rows transactionally and hard-failing when classification is missing. Fixes the fork+install conflation that produced first-render-with-wrong-layout (two commits, first knowingly wrong). Scope decisions: brave backfill, hard-fail loud logging, adoption and new builds unified, re-resolution deferred to HITL, fork-to-library gated behind two flags.
- **sources:** HANDOFF_2026-04-18_design_and_styling…md#3-6; HANDOFF_2026-04-20_composition_deployed…md; FOCUS_navigation_HANDOFF_navigation_fix.md#Architectural-Gap (origin: navigation/layout spec idea)
- **relations:** composable theme migration; navigation/layout specs; classification tags mismatch
- **verify-later:** site-design-planner agent definition; resolve_composition_*.go, install_site_composition.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Composable theme migration 025 (palette + layout + typography decomposition)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phases 1-3 (data model, layouts, seeding) are deployed and verified. Phases 4-5 (renderer cutover, fork action rewrite) were deployed but not end-to-end verified" (2026-04-18); renderer subsequently exercised in later cascades
- **what:** Themes decomposed into palettes, layouts (15 seeded CSS templates each passing a 7-point contract audit), typography_sets (6 seeded), FK-linked from css_themes/style_collections; renderer cutover to a single JOIN loader + FuncMap (palette/typo/token) with hard error on NULL FKs; fork action resolves the three pieces before insert.
- **sources:** HANDOFF_2026-04-18_design_and_styling…md#2
- **relations:** CSS assembly pipeline; site-design-planner
- **verify-later:** layouts/palettes/typography_sets row counts; render_css_composition_loader.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### webdesign-agent post-merge loop bug and generate_css stuck mystery
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "This is a loop bug in my migration … Fix proposal (NOT YET APPLIED)"; "Even with the loop fixed, we STILL don't know why generate_css didn't execute" (2026-04-20); a later cascade (04-23) "proceeded through generate_css and deploy_css to check_should_fork" suggesting recovery
- **what:** The 010 migration left every path out of deploy_css looping back to generate_css (update_site.next_step and check_update_db.else_step should point at check_should_fork); separately, one run sat at generate_css (deterministic action) producing no log line, no heartbeat, evidence lost to pod rotation. Instrumentation runbook written for reproduction.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#A
- **relations:** silent completion; consumer-group race (candidate explanation)
- **verify-later:** current webdesign-agent next_step wiring; whether the loop fix SQL was applied

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Composition classification-tags mismatch (industry_tags empty)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "composition_layout.reason: 'fallback — no classification tags' … layout resolver fell back to brochure-formal for a site that clearly wants something dashboard/application-like" (2026-04-20); migration 008 (dynamic taxonomy, industry_tags array from classifier) validated 2026-04-23 likely addresses it
- **what:** The layout resolver read a nonexistent tags array while classification stored industry/sub_industry strings, so every site fell back to the generic layout and style_collections.industry_tags was written empty, breaking future library matching. Migration 008 made the classifier emit an industry_tags array against a dynamic taxonomy read from the layouts table (read_layout_taxonomy action), validated end-to-end with tool-portal-dark selected via library_match.
- **sources:** HANDOFF_2026-04-20_composition_deployed…md#B; HANDOFF_2026-04-23(1).md#deployed, #validated
- **relations:** site-design-planner; dynamic taxonomy classifier
- **verify-later:** readClassificationFromContext in resolve_composition_helpers.go; classifier output shape post-008

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Planner palette prose vs structured reference_values (Gap C)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "OPEN" (2026-05-26): planner emits colour decision as design_intent.colour_mood prose; composition palette cascade misses the design_intent slot and falls to layout-seed default
- **what:** Planned colours reach the render only via the webdesign-agent overlay, not the base composition. Fix options: planner emits a structured palette.reference_values block (primary/secondary/accent/background/surface/text/text_muted/border) or site-design-planner consumes colour_mood directly.
- **sources:** HANDOFF_2026-05-26…md#gaps, #Where-to-resume
- **relations:** palette cascade; site-design-planner
- **verify-later:** plan_site output schema; palette resolver slots

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Library-row cleanup pattern for failed cascades
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** executed 2026-04-23 with counts (4 css_themes, 7 palettes, 4 style_collections cleared for gamesdesign) and a NOT IN guard protecting seeded layouts
- **what:** Bad cascades leave one set of library rows (css_themes/palettes/style_collections/typography_sets) per resolve attempt; if left, the matcher can pick wrong-decision artefacts for future sites. Reverse-FK-order delete by source_domain is the recovery pattern. Related open item: site deletion should clean up unreferenced library rows (FKs are SET NULL, leaving orphans).
- **sources:** HANDOFF_2026-04-23(1).md#cleanup, item 18
- **relations:** site-design-planner re-resolution ambiguity; duplicate sites-row question (item 20)
- **verify-later:** any delete-site action's library handling

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme derivation and drop at render
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** PLAN §Confirmed-at-code-level: "`deriveSchemeFromDesignIntent(style_direction, suggested_style)` returns light/dark/''; … `buildResolvedCompositionSpec` records the layout/palette ids … but not the scheme value"; notes (Sb) traced it end-to-end 2026-06-30.
- **what:** Scheme (light/dark) is derived at composition from `design_intent.style_direction` by substring matching, used by `resolveLayoutByTags` as a near-hard constraint to pick the layout, then dropped: neither the CSS loader SELECT nor the component `RenderContext` reads `layouts.scheme` (check-constrained to light/dark/neutral). It survives only as the layout's curated property, recoverable via `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme`. Light/dark variety is handled by paired layouts (tool-portal-light/-dark), not runtime component flipping. Corollary data fact: only 3 of 18 active layouts have `scheme` set — scheme metadata is sparsely curated.
- **sources:** PLAN_scheme_to_components(1).md#Confirmed-at-code-level; running_notes_scheme_to_components(55).md#Sb #Sf; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3e)
- **relations:** three-part styles.css assembly; paired-variable standard; explicit RenderContext.Scheme (abandoned).
- **verify-later:** `platform/orchestration/actions/` deriveSchemeFromDesignIntent, resolveLayoutByTags, buildResolvedCompositionSpec; `layouts.scheme` column + values.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Layout CTA pair curation with WCAG contrast gates
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Notes (Sq) "W1 complete + verified"; (Su) "W1b COMPLETE: five layouts curated"; w1b comments record the five hex swaps and expected values.
- **what:** W1 added the missing CTA pair to tool-portal-light via anchored `regexp_replace` on the verified `--color-footer-text` line (`{{palette "cta_bg" "#e9e2d3"}}` + `cta_text "#1a1a1a"`, contrast ≈13.5, mirroring tool-portal-dark's neutral elevated band; accent alternative offered). A sweep computed every layout's cta pair contrast; five seed layouts failed 4.5 with white text and got same-hue darker fallbacks (W1b, zero live impact — no site uses them). Pair values are deliberate per-layout design: several light layouts curate DARK footer bands — "light site, dark band by choice" is already a curated model in the library. Requirement carried into the contract: pair contrast ≥ 4.5.
- **sources:** w1_01_add_cta_pair.sql; w1b_01_contrast_batch.sql; RUNBOOK_scheme_to_components(50).md#W1-RESULTS #CHECK-4-RESULTS (4b); SPEC_scheme_to_components.md#W1
- **relations:** paired-variable standard; three-part styles.css assembly.
- **verify-later:** layouts css_template cta pair values for the six touched layouts.

<!-- SOURCE: U04_idea_uk.md -->
### Two-stage base+override design pipeline (site-design-planner + webdesign-agent)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Investigation 2026-06-20 "confirmed against agent_definitions… and the deployed render code"; routing table shows needs_composition→site-design-planner, needs_design→webdesign-agent with an explicit depends_on.
- **what:** Design is deliberately split (027 §2 — ordering was the reason): Stage 1 `site-design-planner` (deterministic, no LLM) resolves layout/typography/palette and installs them (css_themes 3-FK row + style_collections + sites.style_collection_id + a resolved_composition spec) — renders nothing; Stage 2 `webdesign-agent` produces an LLM design overlay, renders the layout template over the installed composition base per a fixed merge-authority rule (LLM wins core palette slots + typography; composition wins layout/structure tokens/specialised slots), and is the sole styles.css deployer (git_commit → Actions → B2). `emit_design_items` queues both from one step with needs_design gated on composition. The 2026-06-20 correction: this is NOT a shared-responsibility bug — the overlay was designed as *optional and partial* (025 §5).
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md; idea.uk/HANDOFF(13).md (three-layer model); idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md
- **relations:** mandatory-full overlay bug; resolved_composition pointer; palette cascade; 002 architecture doc (rewritten 2026-06-22 to match).
- **verify-later:** render_css_composition_helpers.go buildPaletteMap/buildTypographyMap; emit_design_items_action.go.

<!-- SOURCE: U04_idea_uk.md -->
### resolved_composition pointer spec + install_site_composition semantics
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Verified live on idea.uk twice (dark install, then re-resolve); "resolved_composition is a *pointer* — it carries palette_id/name/source, not the colour values".
- **what:** The composition install contract: css_themes row created with all three FKs but empty css_content (webdesign-agent fills at render); style_collections points at the theme; `sites.style_collection_id` set only if NULL — install **errors rather than overwrites** an existing composition ("re-resolve not supported; clear it manually"), which is why re-resolving requires an explicit detach; the old resolved_composition spec is superseded and a new one inserted as the lineage/decision record (`lineage.{palette_source, typography_source, layout_source}`). Renderer resolution is strict: missing/NULL composition parts hard-error ("migration gaps are audit events, not silent fallbacks"), with a loud emergency fallback to standard-brochure.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (install + loader sections); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A: "a bare needs_composition requeue no-ops")
- **relations:** composition re-resolve procedure; two-stage pipeline.
- **verify-later:** install_site_composition_action.go; render_css_composition_loader.go.

<!-- SOURCE: U04_idea_uk.md -->
### Palette/typography resolution cascade + the dead-slot bug and fingerprint fallback hardening
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Cascade live and proven; dead slot "CONFIRMED why (2026-06-19, from resolve_composition_palette.go)"; hardening "DELIVERED" as code 2026-06-19/20 but "READY… NOT YET APPLIED" in the backlog (needs image rebuild + roll).
- **what:** Palette source cascade: design_reference → mission → design_intent.palette.reference_values → layout seed → archetype default (typography analogous; palettes always site-specific, layouts a shared curated library). The bug: cascade slot 1 reads `design_reference.palette.reference_values`, a key the adoption fingerprint never writes (it stores suggested_mapping/css_variables/colors) — so slot 1 was dead and adopted references never drove the palette; the delivered hardening points slot 1 at the fingerprint's real keys as a fallback after design_intent. Under the current LLM-wins merge the composition palette mostly doesn't paint anyway (it fixes lineage + rare-gap fallback) — the painting lever is the classifier fix feeding the LLM.
- **sources:** idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md#3; idea.uk/HANDOFF(13).md (cascade + backlog item 3); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (problem 1)
- **relations:** structured design_intent migration; two-stage pipeline; adoption generate_design_intent.
- **verify-later:** resolve_composition_reference_helpers.go deployed or not; extractPaletteSignal/extractTypographySignal.

<!-- SOURCE: U04_idea_uk.md -->
### The mandatory-full overlay bug + improver-not-rewriter fix (and the superseded rewrite options)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** "direction settled 2026-06-20, not started" (runbook open items); superseded options (a)/(b)/(c) "kept for history" in the investigation doc.
- **what:** The merge rule assumed the LLM overlay would be *optional and partial* (asserting only genuine brand identity), but the analyze_design prompt mandates a full 8-slot color_scheme with "be distinctive" framing — so the LLM repaints every fresh build, the 028-forbidden silent override. Fix v1 (no contract change): show the LLM the established palette, require it to keep it as the foundation and change slots only with a reason, diff the result against the composition base and write an audit record (slot, old→new, reason) per build; v2 (deferred, evidence-driven): cap core-slot changes per refine + optional denylist. Explicitly supersedes the earlier single-owner options: (a) LLM-owns-core / slim the planner, (b) flip the merge so structured composition wins, (c) collapse to one design agent — rejected because the base+partial-overlay split is intentional.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (CORRECTION + fix sections; superseded options); idea.uk/DESIGN_PIPELINE_two_track_investigation(1).md (the pre-correction "decision options" — family-delta); idea.uk/HANDOFF(13).md (backlog item 4)
- **relations:** structured design_intent (precondition); design docs 025/027/028.
- **verify-later:** analyze_design prompt in webdesign-agent def; any design-audit table.

<!-- SOURCE: U04_idea_uk.md -->
### Scheme-aware weighted layout matcher + layouts.scheme + tool-portal-light
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Matcher code LIVE — merged … built into the chassis image and site-design-planner rolled. (Confirmed live 2026-06-25.)"; migration applied; re-resolve proved tool-portal-light selection end-to-end 2026-06-25.
- **what:** Replaces the tags-only, scheme-blind `resolveLayoutByTags` (exact-overlap count, alphabetical ties) that put light-editorial idea.uk on tool-portal-dark. New matcher (all in Go, ~17-row library fetched and scored transparently): scheme as a **near-hard constraint** (a light site won't land dark while any non-dark fits; mismatch queues the existing needs_new_layout_candidate HITL item), IDF-weighted tag rarity (specific beats generic), synonym normalisation to a controlled vocabulary, category + description keyword bonuses. Paired migration adds nullable `layouts.scheme` (light/dark/neutral; NULL degrades gracefully) and a new `tool-portal-light` layout — same structural class contract as its dark twin, light fallbacks, reads palette vars. Decision history: NO auto-layout-generation — a curated, varied library + scheme-aware matching is the lever; LLM-judge/pgvector deferred.
- **sources:** idea.uk/resolveLayoutByTags_weighted.go.patch.txt (header); idea.uk/migration_layouts_scheme_and_light_tool_portal.sql; idea.uk/HANDOFF(13).md (matcher rewrite)
- **relations:** deriveSchemeFromDesignIntent; composition re-resolve; scheme-to-components gap (the next layer down).
- **verify-later:** fork_theme_composition.go current resolveLayoutByTags; remaining NULL layouts.scheme rows (backlog).

<!-- SOURCE: U04_idea_uk.md -->
### Composition re-resolve procedure (gated, file-based, backup-first)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Steps 1–6 all marked DONE with results (2026-06-22→25); "RE-RESOLVE SUCCEEDED: idea.uk now on tool-portal-light (scheme fix proven end-to-end)".
- **what:** The safe pattern for re-running composition on an already-built site (install refuses overwrites): ordered SQL FILES — backup+inspect (with four uniqueness checks that must all be 0), gated detach+clear (NULL style_collection_id; delete the site's own collection→theme→palette→typography chain only where source_site_id matches; supersede the old spec), state-check, kcat re-trigger of site-design-planner (`domain` required by ensure_site_record), verify. Two learned caveats now doctrine: run SQL as files never pasted (paste mangled \set/blank lines and left an open transaction); a standalone-orchestrated planner ends at install and emits NO needs_design — the styles.css render is a separate explicit webdesign-agent orchestration. Distinct from the adoption teardown (bulk delete by source_domain), which must NOT be used on a fresh site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (re-resolve section); idea.uk/reresolve_idea_uk_01_backup_and_inspect.sql (+02/02b/03/04/05 series); idea.uk/running_notes(63).md (xxx–jjj checkpoints)
- **relations:** install semantics; launch idioms; scheme-aware matcher validation.
- **verify-later:** bak_*_idea_20260625 tables; orchestration_states rows for the re-resolve correlations.

<!-- SOURCE: U04_idea_uk.md -->
### The scheme-does-not-reach-components gap (P0 framework fix)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Investigation complete + report written 2026-06-26 ("Step 7 DONE — conclusion: do NOT rebuild yet; structural gap"); the fix itself deferred to a dedicated thread; composition+stylesheet layers verified correct, page layer still dark.
- **what:** The central framework finding of the thread: a site's scheme (light/dark) is decided at composition and reaches styles.css `:root`, but never reaches the components that render sections/header/footer. Components are drawn from a **dark-oriented library** by a one-active-component-per-function lookup (nothing light exists for hero/CTA/footer), self-style via inline CSS with their own class vocabulary (the layout's section rules don't apply — class-name mismatch), and hardcode dark treatments or set dark `--section-*` themselves — so a light-resolved site renders dark chrome over light content (only var-reading sections went parchment). Supporting facts: `is_dark_section` is loaded but never used in selection, unreliable, and conflates "intrinsically dark" with "should contrast the page"; no layout declares default header/footer and the planner never runs update_site_defaults; the hero navy-button bug (--accent-color vs --color-accent) was already fixed in the library — deployed pages were stale. Modelled as: scheme was treated as colour-only; the structural half was never plumbed.
- **sources:** idea.uk/REPORT_scheme_does_not_reach_components.md; idea.uk/HANDOFF_scheme_to_components(1).md; idea.uk/running_notes_2(6).md (lll–ooo); idea.uk/one_sentence_description.md
- **relations:** scheme-as-override thesis (the fix shape); section-contrast model; header/footer wiring; component class-contract question; scheme-aware matcher (upstream, done).
- **verify-later:** whether the dedicated thread landed (component templates de-hardcoded; light footer exists; update_site_defaults in build path).

<!-- SOURCE: U04_idea_uk.md -->
### Scheme-as-override thesis + section-contrast model (base scheme + per-section contrast intent)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** REPORT §4–5: "the likely shape (to be validated by the investigations, not assumed)"; eight design questions "must be answered before any code".
- **what:** The design steer for the P0 fix: scheme is a **set of variable values** — an override layer supplied by composition/renderer and consumed by de-hardcoded components — never a duplication of components into *-light/*-dark variants (new functions only for genuine structural divergence). The model separates **site scheme (base)** from **per-section contrast intent** (a dark hero on a light site is legitimate, intentional contrast — "make everything light" is wrong), both applied as data at render time through the existing `--section-*` mechanism, making the renderer the single adaptation point. Eight design questions scoped (where scheme lives at render; who owns section darkness; the override mechanism; the gating class-vocabulary question Q4; is_dark_section's fate; header/footer; migration without breaking dark sites; an auditor guard), with nine investigations (A–I) and a provisional fix shape stated as hypothesis. Definition of done includes a scheme-coherence audit so it can't silently regress.
- **sources:** idea.uk/REPORT_scheme_does_not_reach_components.md#4-8; idea.uk/HANDOFF_scheme_to_components(1).md; idea.uk/TODO_chassis_and_idea_uk(1).md#P0
- **relations:** CSS colour-inheritance model (the vehicle); improvement-loop (the guard); scheme gap (the problem).
- **verify-later:** stage-2 code check of component templates for --section-* consumption.

<!-- SOURCE: U04_idea_uk.md -->
### Header/footer chrome wiring chain (and its live gaps)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Data findings 2026-06-26: all inspected layouts have NULL default header/footer ids; idea.uk's new style_collection inherited NULLs; site_components still point at the original dark, now-inactive site-header/site-footer.
- **what:** Site-level chrome flows down a chain: `layouts.default_{header,footer}_component_id` → install_site_composition copies onto style_collections → `update_site_defaults` copies onto site_components → renderAndStoreSiteComponent renders into site_components.rendered_html, with a hardcoded RenderFallbackHeader when unlinked. Live gaps: no layout declares defaults, site-design-planner never runs update_site_defaults, and header/footer are therefore never scheme-derived — a re-resolve leaves the old chrome in place. The library has light headers but NO light footer. Fix direction: layouts declare scheme-appropriate defaults + the build runs update_site_defaults, one adaptive header/footer per the override thesis.
- **sources:** idea.uk/running_notes_2(6).md (mmm data findings); idea.uk/REPORT_scheme_does_not_reach_components.md (facts + Q6/investigation F); idea.uk/001_component_flow.md
- **relations:** scheme gap; contracts-and-standards Site Component Linkage Contract (003).
- **verify-later:** update_site_defaults_action.go and its call sites; how the original build chose idea.uk's header.

<!-- SOURCE: U04_idea_uk.md -->
### Section→component resolution: direct-function Path 1 vs scoring selector Path 2
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** plan_sections code read 2026-06-26: "Path 1 = components[sectionName] direct lookup… All current sites hit this path"; component_selector "SELECTs is_dark_section into the struct but NEVER uses it in scoring".
- **what:** How a planned section becomes a component: Path 1 matches the section name directly against `content_components.function` (one active component per function — uniqueness index), which all current sites hit; Path 2, the scoring `component_selector` (suitable_site_types 0.35 + page_types 0.15 + quality 0.3 + specificity + usage), only runs for section_type names that aren't functions — and is scheme-blind. Consequences: there is no place to pick a scheme-appropriate variant for current sites (making a scheme-aware selector necessary-but-insufficient), and layout-aware section selection is explicitly documented future work (027 §10). page-rerender re-assembles stored HTML without re-selecting; only page-build-handler re-runs plan_sections.
- **sources:** idea.uk/running_notes_2(6).md (mmm/nnn corrections); idea.uk/REPORT_scheme_does_not_reach_components.md#2; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Step 7 findings)
- **relations:** scheme gap; tool-library (component registry/matching); flag_page_image_rebuild as the rebuild trigger.
- **verify-later:** plan_sections_action.go Path-1 comment; component_selector.go scoring.

<!-- SOURCE: U05_content_quality_linking.md -->
### Design fingerprint pipeline (design_reference / design_intent / three-way priority)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** WORK_PLAN_v2 status tables: Phases 1–2 largely "Applied to DB"/code-ready, Phase 3 (site-design-planner) and Phase 4 (requirement-driven components) "Not started".
- **what:** The adoption design-fidelity subsystem: extract_design_fingerprint parses crawled rawHTML for concrete tokens (CSS vars, hex, fonts) into a design_reference spec; an LLM derives the semantic design_intent from it; the webdesign-agent applies a three-way priority (design_intent → design_reference reproduce-faithfully → generate-from-industry); audit loop proposes rather than applies design changes. The companion problems doc catalogues the failure it answers (LLM guessed colours/fonts from markdown summaries, layout flattened to generic components, header/footer genericised). Phase 3 (navigation/layout specs + site-design-planner agent) and Phase 4 (section recipes, requirement-driven component selection, "every build conceptually an adoption") remain unbuilt. Carried in this unit as packaged context (stale-risk flagged; the composition refactor postdates them).
- **sources:** package_module/output_contexts/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md; FOCUS_design_and_styling_adoption_problems.md; HANDOFF_2026-06-09(2).md#design-fidelity-background
- **relations:** adoption pipeline workflow (extract_fingerprint/generate_design_intent steps); design-composition unit docs 025–027.
- **verify-later:** extract_design_fingerprint_action.go deployed?; site_specs design_reference/design_intent aspects; site-design-planner existence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Deterministic styles.css rendering (webdesign-agent: LLM spec → Go template → git commit)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "render_css_from_spec — 'Render CSS from design spec using Go template (deterministic, no LLM)'" (NOTES §9bi, from webdesign-agent config); full chain observed live §9bn.
- **what:** The webdesign flow is analyze_design (LLM → design-spec JSON: color_scheme/typography/spacing) → render_css_from_spec (deterministic Go template over DB layout templates — `comp.LayoutTemplate` — merged with palette/typography; forkable themes) → git_commit styles.css → site-asset-renderer. The defined CSS vocabulary therefore lives in one Go-owned render path (the single home for generic fixes); storage_actions.go's styles.css writes belong to the OLD builder extract paths and must not be patched for this flow. Caution: re-running analyze_design mints a fresh LLM spec — palettes can shift unless pinned — hence the manual bridge-commit option for palette-preserving fixes.
- **sources:** NOTES(43).md §9bi, §9bj, §9bm; RUNBOOK(49).md Part D
- **relations:** D2a (lives inside it); R6f; layout curation.
- **verify-later:** render_css_from_spec_action.go; webdesign-agent default_config; needs_design production (build-site-planner).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Scheme resolution pipeline and where the signal stops
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Scheme→variable pipeline verified correct; all 18 layouts carry the four chrome vars" (RUNBOOK_scheme(18) CURRENT POSITION, 2026-07-02); RenderContext gap traced in running_notes Sb.
- **what:** A site's light/dark scheme derives from design intent (`deriveSchemeFromDesignIntent`), constrains layout matching (`resolveLayoutByTags`; `layouts.scheme` light/dark/neutral), and reaches styles.css via palette :root + luminance defaults — but is never recorded in the composition spec and never reaches the component render context (`RenderContext` has palette colours, no scheme field). The corrected understanding: the scheme reaches components IMPLICITLY via variables; components defeat it by hardcoding dark assumptions — so the core fix is de-hardcoding, not new plumbing.
- **sources:** running_notes_scheme_to_components(22).md Sb, Sc, Sf, Sk; RUNBOOK_scheme_to_components(18).md header + CHECK 1
- **relations:** paired-variable direction; buildSectionDefaults; R6f (later vocabulary-level echo).
- **verify-later:** deriveSchemeFromDesignIntent; RenderContext struct; layouts.scheme population (only 3 of 18 curated).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Paired-variable design direction (Alt C: curated bg+text pairs; completion of the existing standard)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "pair convention is ALREADY the standard — 18/18 --color-primary-text, 17/18 --color-cta-text" (Check 4c); W1–W3e template work executed 2026-07-02/03 but "inert until re-render/rebuild"; Go batch + tail unshipped.
- **what:** The user requirement "a light scheme must be able to render fully light, and may carry dark hero bands" selects layout-curated background+text variable pairs (chrome pattern generalised: --color-cta-bg/--color-cta-text etc.), palette-overridable per site (specialised slots theme-wins), per-instance later via plan directives — components consume pairs and never declare `--section-*`; renderer luminance defaults remain the base; dark-band-by-choice stays curated per layout. Judged a COMPLETION of existing architecture, not a restructure (one layout to patch, components to bring in line). Execution: ten templates fixed + seven verified clean (footer, CTA via inverse-pair buttons, hero, five hero-* variants, about-content, brief-explanation), idea.uk chrome repointed; full rebuild + Go batch (scheme-aware fallbacks, creator prompt, fixer re-aim) pending at capture.
- **sources:** running_notes(22).md Sn, So; RUNBOOK_scheme_to_components(18).md CHECK 4 RESULTS + WHERE WE ARE; SPEC referenced therein
- **relations:** hazard/band split; hero ink model; is_dark_section demotion; Phase 4.5 (deferred); chrome linkage repair.
- **verify-later:** SPEC_scheme_to_components.md (outside unit); layouts cta pair coverage; whether W6 rebuild shipped.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Hazard-class vs band-class self-declarer split; is_dark_section demoted to metadata
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "the 37 self-declarers split into two classes" with named components (Check 3c, run 2026-07-02); "6 declarers have is_dark_section=f… never key styling on the LLM-authored flag".
- **what:** Library-wide diagnosis: of 84 active section components, 37 self-declare `--section-*` — ~18 hazard-class (declare dark context while painting surface vars or nothing → white-on-light bugs today) vs ~19 band-class (paint palette bands + white text — coherent but block "fully light"); 15 carry hex backgrounds. `is_dark_section` is an LLM-authored component bool contradicted by 6 of its own declarers and consumed by nothing that styles — demoted to selection/imagery metadata; styling must never key on it. This classification sized every subsequent fix batch.
- **sources:** RUNBOOK_scheme_to_components(18).md CHECK 2/3 RESULTS; running_notes(22).md Sn, Sh (E findings)
- **relations:** paired-variable direction; improvement-loop fixers (key on the flag — part of why they're wrong); component-creator prompt drift.
- **verify-later:** content_components is_dark_section values vs template styling; remaining unconverted declarers (~10 hazard + ~17 band).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Chrome linkage tangle: four overlapping header/footer default stores and the dark fallback
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "header_component_id is effectively a DEAD column — nothing populates it" (running_notes Sl); repoint executed (Ta); scheme-aware fallback Go batch still pending in WHERE WE ARE.
- **what:** Four coexisting default stores for site chrome: style_collections.header/footer_component_id (the store RenderHeader reads first — installed NULL and never written), site_components slots (render cache, can pin inactive components indefinitely), sites.default_components jsonb (UpdateSiteDefaultsAction target, unread on the render path), layouts.default_*_component_id (all NULL). RenderHeader's chain is collection-id → GetComponentByFunction("site-header") → RenderFallbackHeader, and the fallback hardcodes dark (PrimaryColor bg + white text) — so any linkage break yields dark chrome regardless of scheme. Fix shape: de-hardcoded active chrome components (header already model), repoint stale pins, scheme-aware fallbacks consuming the chrome var pairs (all 18 layouts already define them).
- **sources:** running_notes(22).md Sg, Sh, Sl; RUNBOOK_scheme_to_components(18).md CHECK 3b, HEAD-SLOT RESOLUTION, W4b
- **relations:** chrome refresh gating; rerender fossilisation (stale pinned renders reached deploys); paired-variable direction.
- **verify-later:** style_collections.*_component_id population; RenderFallbackHeader/Footer/Head current CSS; whether the Go batch shipped.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Layout curation: CTA pair completion, WCAG contrast batch, updated_at trigger
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "W1 COMPLETE… W1b: five × UPDATE 1… trigger observed working in anger" (RUNBOOK_scheme(18) W1/W1b/W2b RESULTS, 2026-07-02).
- **what:** Seed-layout curation as part of the theming fix: tool-portal-light gained the missing --color-cta-bg/--color-cta-text pair (#e9e2d3/#1a1a1a, ≈13.5 contrast); a five-layout batch nudged failing cta_bg fallbacks to same-hue passes (all ≥4.5); layouts.updated_at gained a BEFORE UPDATE trigger via the shared set_updated_at function (reuse-gate fired as designed when CREATE FUNCTION collided). Several light layouts deliberately curate dark footer bands — "light site, dark band by choice" is an existing curated model, not a bug.
- **sources:** RUNBOOK_scheme_to_components(18).md W1/W1b/W2b RESULTS, CHECK 4b; running_notes(22).md Sq–Ss
- **relations:** paired-variable direction; deterministic styles.css rendering.
- **verify-later:** layouts cta pair values; trg_layouts_updated_at.

<!-- SOURCE: U09_adoption.md -->
### Design/composition flow gaps A–B–C and the plan-time triggers
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "UPDATE 2026-05-26: gap (A) is being deployed to production now — emit_design_items + emit_imagery_items are wired into build-site-planner (agent-chassis v1.0.1047)… Gaps (B) and (C) remain open."
- **what:** Three stacked gaps behind themeless/off-palette built sites: (A) composition never triggered — the Phase-1 refactor lost the needs_composition/needs_design emission; restored as plan-time steps `emit_design` (guarded on style_collection_id IS NULL) and `emit_imagery` (priority-banded so imagery lands in the first deploy); (B) planner design drift — the adopted `design`/`design_reference` aspect is never rendered into the plan_site prompt, and `design_intent.style_direction` is a fixed 3-value enum (professional-dark|modern-light|bold-creative) that cannot express e.g. "cyberpunk terminal", forcing collapse; (C) colour reaches the resolver only as prose `colour_mood` flattened into directives — the palette cascade's design_intent slot needs structured `palette.reference_values` (core keys primary…border). Whether colours are preserved vs re-chosen is governed by fidelity+locks, not the shape. REUSE NOTE outstanding: extract shared emitInitialCompositionAndDesign so emit_design's insert block can't drift from WriteBuildItemsAction.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md, README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** theme-layer resolution; imagery loop closure; fidelity dial; spec-retire investigation (dead `structure` aspect, single-reader `design` aspect)
- **verify-later:** build-site-planner v1.0.1047 workflow (reconcile_site_plan → emit_design → emit_imagery → complete); style_direction enum; createPalette core keys

<!-- SOURCE: U09_adoption.md -->
### Spec ownership / silent-override failure-mode principle
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Restated as settled doctrine from doc 028 in FOCUS_design_composition §5 (2026-05-26).
- **what:** An agent that changes behaviour on information it did not put in the spec is a bug; an agent that overwrites a spec aspect another agent owns is a category error; an agent that produces the right output but doesn't write it to the spec is not helpful. Applied concretely: `design_reference` is owned by site-adoption-agent and records the source site's design — writing our chosen palette into it would misrepresent the source (corrected an earlier proposal). Producer/consumer schema drift (colour_mood prose vs reference_values; features {title,description} vs {icon,name,description}) is the same failure shape.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#5, #2
- **relations:** design gaps B/C; site_specs aspect ownership model
- **verify-later:** doc 028 statement; aspect writer inventory

<!-- SOURCE: U10_imagery.md -->
### No runtime re-compose path — layout change via the 025 FK-swap pattern
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "B7 COMPLETED 2026-07-10 evening — via the 025 FK-swap pattern… there is no runtime re-compose path (deliberate deferral). NEW OPEN ITEM: build a proper runtime re-compose mode."
- **what:** Changing an existing site's layout is deliberately unsupported at runtime: install_site_composition refuses when a style_collection exists, and fork_theme_from_site's install mode was removed 2026-04-19. The sanctioned workaround is a targeted `css_themes.layout_id` FK swap (backup + verify) followed by a webdesign-agent CSS re-render + page rerenders. Root cause of the B7 brochure fallback: robot-hands' old-format classification lacked `industry_tags`, so the scheme-aware matcher had nothing to score — while the layout library already held the right answer (tool-portal-dark, itself grown from a prior instance of the same gap: the library learns).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-18–22, SQL_2026-07-10_b7_layout_fix.sql, SQL_2026-07-10_b7_layout_swap.sql, PLAN_imagery_best_in_class.md#B7
- **relations:** design-composition doc 027 matcher; needs_new_layout_candidate → library-growth loop; classification format drift (also caused the missing news flag).
- **verify-later:** install_site_composition refusal; robot-hands css_themes.layout_id = tool-portal-dark.

<!-- SOURCE: U12_docs024_archives.md -->
### design_reference / design_intent spec-aspect split
*(merged from 2 independent findings)*
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** `FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md`: "Replaces the old vague `design` spec"; `027_design_and_site_planner_v2.md`'s "Related Specs" table (an independent, later live doc) confirms both `design_reference` and `design_intent` as live, read spec aspects with defined priority cascades.
- **what:** `design_reference` holds concrete values (hex colours, font families, CSS variables, spacing) extracted mechanically (no LLM) from an adopted site's crawled HTML/CSS — a historical, immutable record. `design_intent` holds semantic creative direction (e.g. "dark IDE aesthetic... start here"), deliberately non-prescriptive so the improvement loop and webdesign-agent retain creative room; it may be auto-generated at adoption time or written later by a strategist/human. Together they replace a single, vague, LLM-guessed `design` spec aspect that conflated historical fact with creative direction (see the separate "Unified design spec aspect for adopted sites" concept below for that earlier, superseded state). Guiding principle: "design reference is history, design intent is direction" / "every build is conceptually an adoption."
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made"; old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"Principles Restated"; old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"Related Specs"
- **relations:** Unified design spec aspect for adopted sites (the superseded precursor, below); webdesign-agent three-way design priority; palette-locked-until-design_intent policy; design agent write-back resolution
- **verify-later:** confirm `design` spec aspect is no longer written anywhere; check `site_specs` for population rate of `design_reference`/`design_intent` across adopted sites.

---
*(remaining concepts below are as extracted by the 7 sub-slices, grouped by their originating file cluster)*

<!-- SOURCE: U12_docs024_archives.md -->
### Design agent responsibility split — site-design-planner (composition) vs webdesign-agent (execution)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Live 002_system_architecture(4).md line 596: "The earlier 'one agent generates brand + CSS' shape is **superseded** by the composition/execution split above."
- **what:** All archived versions of the system-architecture doc (baseline through the april26 near-final draft) model design as owned by a single agent, `webdesign-agent`, which does brand analysis, colour/typography/spacing decisions, AND CSS generation — with `brand-designer`, `layout-architect`, and `style-generator` listed as "Future split"/"Planned" agents that never materialized under those names. The live doc replaces this with a Composition/Execution/Maintenance split: `site-design-planner` (new agent) deterministically resolves layout (weighted, scheme-aware match against a shared `layouts` library), typography (match-or-new against `typography_sets`), and a site-specific palette via signal cascades, then installs `css_themes` + `style_collections` + a `resolved_composition` decision-record; `webdesign-agent` is narrowed to rendering/committing `/assets/css/styles.css` from that installed composition — "the only writer of styles.css." The `Design | webdesign-agent | Colour palette, typography, spacing, CSS` row in the Responsibility Boundaries table is likewise split into separate "Composition" (site-design-planner) and "Render" (webdesign-agent) rows, with a `needs_new_layout_candidate` HITL escalation replacing the old simple "search → maybe reuse → maybe generate" theme-growth description.
- **sources:** old/older1/002c_system_architecture_v3.md#"Design Agent Family", #"Theme library growth"; old/older1/002d_quality_assurance_architecture.md#"Classifier → Planner → Design Agent → Audit Agent"; docs024_key_docs_latest/002_system_architecture(4).md#"Composition: how a site's design is resolved and installed", #"Classifier → Planner → Design Agent → Audit Agent"
- **relations:** superseded planned agents brand-designer / layout-architect / style-generator (never built under those names); fork_theme_composition.go resolvers; QA architecture's "Responsibility Boundaries" chain
- **verify-later:** confirm `site-design-planner` agent_definitions row and `resolve_composition_layout/typography/palette` actions exist and are active.

<!-- SOURCE: U12_docs024_archives.md -->
### Early "visual identity poles" layout taxonomy (dropped)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Diff-confirmed only in the earliest of four palette/layout/typography-migration drafts; final 15-layout table uses different (hyphenated) names though keeps several "pole" nicknames.
- **what:** The very first migration draft described layout diversity as nine named "poles" tied to specific sites (Brochure/corporate, Magazine/editorial, Portfolio/kinetic "vonc", Commerce/grid, Utility/tool "thunder compute", Media/streaming "youtube", Documentation/reference, High-energy/bold "boxing", Soft/editorial). Dropped in favour of vaguer prose, then crystallised differently as the final 15-layout table (adding six layouts absent from the original nine-pole list).
- **sources:** old/older1/025_palette_layout_typography_migration.md#"2. Scope Decisions"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md#"7. The Layouts to Build"
- **relations:** composable theme migration; site-design-planner
- **verify-later:** final `layouts` table row count/names vs. the 15-layout plan.

<!-- SOURCE: U12_docs024_archives.md -->
### Phased belt-and-braces removal plan for webdesign-agent install_theme (abandoned same-day)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** v1 doc's changelog (2026-04-19) states the belt-and-braces step "remains... pending the two-phase removal plan"; live v2 doc's changelog the same day: "Merge applied" (direct removal).
- **what:** `026_design_and_site_planner_v1.md` proposed a cautious two-phase removal of webdesign-agent's defensive `install_theme`/`check_should_install` steps (diagnostic no-op first, delete only after a week of zero firings). Abandoned within hours: live v2 shows the two steps deleted outright the same day, routing rewired directly to `generate_css`, relying instead on the renderer's emergency-fallback logging as the sole safety net.
- **sources:** old/older1/026_design_and_site_planner_v1.md#"6. Removing install_theme From Webdesign-Agent (Planned)"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"6... (Applied)", #"12. Change Log"
- **relations:** site-design-planner composition-install path; renderer emergency-fallback logging
- **verify-later:** confirm `install_theme`/`check_should_install` are absent from webdesign-agent's agent_definitions.

<!-- SOURCE: U12_docs024_archives.md -->
### Visual identity library and effects library (composable design assets)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)" in the 2026-04-11 plan; not confirmed built.
- **what:** Longer-term plan for two accumulating libraries: a visual identity library of palettes/typography/effects searchable by purpose/audience, and an effects library treating elevation/corner radius/animation/density as composable modifiers independent of layout. Likely precursor idea to the `palettes`/`typography_sets`/`layouts` table split actually implemented.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4: Requirement-Driven Components (longer term)"
- **relations:** composable theme migration; component selector by functional requirement
- **verify-later:** whether structure_tokens/effects concepts in the live composable-theme schema fulfil this idea.

<!-- SOURCE: U12_docs024_archives.md -->
### Three-layer design system (content_components / css_themes / style_collections)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Live: "Today css_themes.css_template mixes three distinct concerns into one row: palette, typography, and layout" — the migration splits this monolith.
- **what:** Early architecture with three independently-varying layers: HTML components, a monolithic CSS theme (one row = whole stylesheet), and a `style_collections` bridge table. Superseded internally when `css_themes` was split into three composable entities; `style_collections` survives as the outer bundle.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"1. The Design System: Three Independent Layers"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md#"Splitting css_themes"
- **relations:** composable theme system; style_collections bundle
- **verify-later:** confirm `css_themes` legacy columns actually dropped (Phase 7).

<!-- SOURCE: U12_docs024_archives.md -->
### Design fingerprint extraction pipeline
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Not started" → "✅ Deployed, works" (2026-04-14) → "Victory: Design Fingerprint Now Correct" (2026-04-16).
- **what:** Pipeline step parsing a crawled site's CSS into a colour/font/layout "fingerprint" so adoption rebuilds match the original. Went from unstarted idea to working end-to-end (gamedesign.uk) across several debugging sessions.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"4. The Adoption Design Gap"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Victory"
- **relations:** design_reference/design_intent split; computed styles extraction; fpExtractCSSVars fix
- **verify-later:** `site_specs` rows with aspect='design_reference' for adopted sites.

<!-- SOURCE: U12_docs024_archives.md -->
### Webdesign-agent three-way design priority
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "✅ Applied" (2026-04-14 handoff).
- **what:** `analyze_design` step branches on which specs exist: design_intent present → creative freedom around described character; only design_reference → faithful reproduction, no invented palette; neither → generate from industry/audience/identity.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-14_v2.md#"Webdesign-Agent Prompt (deployed)"
- **relations:** design_reference/design_intent spec-aspect split
- **verify-later:** current webdesign-agent agent_definitions prompt text.

<!-- SOURCE: U12_docs024_archives.md -->
### Palette-locked-until-design_intent policy
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Palette is locked until design_intent exists."
- **what:** First adoption build reproduces the original palette exactly (locked); once design_intent is written, webdesign-agent gains creative freedom within the described character, letting the improvement loop evolve the palette over time.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made"
- **relations:** design_reference/design_intent split; audit loop "propose not apply"
- **verify-later:** improvement-loop audit code for propose-vs-enforce mode switch.

<!-- SOURCE: U12_docs024_archives.md -->
### Legacy monolithic CSS renderer internals (removed)
- **category:** design-composition
- **status-signal:** abandoned
- **status-evidence:** "Phase 4.3 already removed... cssTemplateData struct (and its 16 hardcoded fields)... Compile-clean."
- **what:** The original renderer held a flat struct populated by `extractDesignColors`/`designColorMaps`, loading one Go template per theme. Deleted wholesale in Phase 4.3 when the renderer switched to composable palette/layout/typography_set rows via FK.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"What Phase 4.3 already removed"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Template variable system"
- **relations:** three-layer design system (superseded); layout archetype library
- **verify-later:** grep codebase for `loadCSSGoTemplate`/`extractDesignColors`/`designColorMaps` — should be absent.

<!-- SOURCE: U12_docs024_archives.md -->
### fpExtractCSSVars regex-based CSS variable extraction (superseded)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** BEM selectors like `.btn--primary:hover` captured as fake variables; replacement uses `:root` block targeting with semicolon-splitting.
- **what:** Original extractor used one whole-stylesheet regex, producing false positives on BEM class names. Replaced with a multi-strategy extractor isolating `:root`/body/`[data-theme]` blocks, with fallback frequency analysis for utility-CSS sites.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"P6 — fpExtractCSSVars BEM False Positives"; old_design_and_styling/FOCUS_design_and_styling_fp_extract_css_vars_integration.md
- **relations:** design fingerprint extraction pipeline; computed styles extraction
- **verify-later:** `extract_design_fingerprint_action.go` — confirm regex-based extractor removed.

<!-- SOURCE: U12_docs024_archives.md -->
### css_templating.go theme-forking bridge (known-broken, scheduled rewrite)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "fork_theme_from_site produces rows with NULL palette_id, NULL layout_id, NULL typography_set_id... Adoption-forked themes are unusable by the render path."
- **what:** `TemplateCSSFromSpec` converts a rendered CSS snapshot into old flat-field-name placeholders and writes it to the legacy `css_themes.css_template` column, which the post-Phase-4.3 renderer never reads — silently producing unusable NULL-FK theme rows. Flagged for a Phase 5 rewrite.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"1. css_templating.go"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Ready for deployment"
- **relations:** fork_theme_from_site rewrite (Phase 5); parallel legacy HTML-assembly render path
- **verify-later:** confirm `fork_theme_from_site_action.go` now produces palette/typography_set rows.

<!-- SOURCE: U12_docs024_archives.md -->
### Parallel legacy HTML-assembly render path (getThemeByID/GetThemeByName)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "css_content is populated for 13 of the 14 themes. standard-brochure has empty css_content... falls through to GetThemeByName('default')."
- **what:** A second, older render path reads `css_themes.css_content` directly into assembled HTML, independent of the spec-driven render path. Left untouched by Phase 4, own known gap flagged for resolution when Phase 7 drops legacy columns.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"2. getThemeByID / GetThemeByName"
- **relations:** css_templating.go bridge; legacy css_themes columns drop (Phase 7)
- **verify-later:** grep for `getThemeByID`/`GetThemeByName` call sites.

<!-- SOURCE: U12_docs024_archives.md -->
### Component-creation via HITL work-item triage (superseded)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** "migration_025_component_triage.sql — an earlier work-item-based approach that was superseded by the direct insert approach... Do not run this file."
- **what:** Earlier plan for seeding new library components via work items routed through HITL triage. Superseded by a direct SQL insert once components were designed and reviewed.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"2. What's Been Completed"
- **relations:** layout archetype library
- **verify-later:** none — historical, file explicitly marked do-not-run.

<!-- SOURCE: U12_docs024_archives.md -->
### Computed-styles extraction via browser JS injection
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Computed styles (Phase 2) deferred... Spec written but not implemented" vs. a complete Go action + workflow SQL in the Phase 2 doc.
- **what:** Supplementary fingerprint step scraping a homepage with injected JS calling `getComputedStyle()`, writing resolved values for a Go action to parse and merge — "ground truth" overriding source-CSS guesses. Fully spec'd but recorded elsewhere as deferred/not implemented.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_computed_styles_extraction_phase2.md; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Fixes Ready But Not Deployed"
- **relations:** design fingerprint extraction pipeline; fpExtractCSSVars fix
- **verify-later:** registry.go for `extract_computed_styles`; site-adoption-agent workflow steps.

<!-- SOURCE: U12_docs024_archives.md -->
### Layout archetype library (15 named layouts)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Phase 1 is next: designing and writing the 15 layout CSS templates" → "Phase 1 — Layouts seeded (15 rows in layouts table)... deployed."
- **what:** Taxonomy of 15 named structural/visual archetypes (brochure-formal, portfolio-kinetic, utility-tool, media-grid, etc.), each with character/structural-trait descriptions, default header/footer/typography, and legacy-theme mappings — the target library for the composable-theme migration's `layouts` table.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"4. The 15 Layouts — Detailed Descriptions"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md
- **relations:** composable theme system; site-design-planner layout resolver
- **verify-later:** `layouts` table rows in DB — confirm 15 rows.

<!-- SOURCE: U12_docs024_archives.md -->
### Palette merge rule (core slots vs specialised slots)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Core slots (spec wins where present)... Specialised slots (theme wins)."
- **what:** When a site composes a theme, core palette slots let the site's own spec win when present; specialised slots (primary_hover, hero_title, cta_bg, etc.) always take the theme's value.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Palette merge rule"
- **relations:** layout archetype library; site-design-planner palette resolver
- **verify-later:** `resolve_composition_palette_action.go` merge logic.

<!-- SOURCE: U12_docs024_archives.md -->
### site-design-planner "Choice B" scope (composition-only)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Choice B adopted. The agent's exclusive responsibility is composition resolution... It does NOT write navigation or layout specs."
- **what:** Decision narrowing site-design-planner to write exactly one spec, `resolved_composition` (palette_id/layout_id/typography_set_id + lineage + reasoning), deferring `navigation`/`layout` spec ownership to future specialist agents — justified by "slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"3. Scope Refinement"
- **relations:** composition resolution architecture; design pipeline guiding principles
- **verify-later:** agent_definitions row for site-design-planner — confirm workflow only writes `resolved_composition`.

<!-- SOURCE: U12_docs024_archives.md -->
### Composition resolution architecture (3 resolvers + install action)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "validate_composition_inputs_action.go — DONE... install_site_composition_action.go — DONE (~562 lines)."
- **what:** site-design-planner pipeline: `validate_composition_inputs` → three resolvers (`resolve_composition_layout` tag-overlap match, `resolve_composition_typography`, `resolve_composition_palette` fingerprint→mission→design_intent→layout-inherit→default cascade) → `install_site_composition` (one transaction: css_themes+style_collections insert, sites update, resolved_composition spec write).
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Deliverable 4"
- **relations:** site-design-planner Choice B scope; composition resolver orphan-rows policy; fork_theme_from_site
- **verify-later:** confirm `resolve_composition_*.go`/`install_site_composition_action.go` in current codebase.

<!-- SOURCE: U12_docs024_archives.md -->
### webdesign-agent install/render ordering bug ("first render wrong layout")
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "9.1 Known ordering issue in webdesign-agent... This is the exact 'first render wrong layout' bug site-design-planner was built to eliminate."
- **what:** webdesign-agent ran `generate_css → deploy_css → ... → install_theme`, so any site without a pre-installed composition hit the emergency fallback and committed it to git before the correct composition was installed a step later. Documented, deferred fix: reorder `install_theme` before `generate_css`.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"9.1 Known ordering issue"
- **relations:** composition resolution architecture; render_css_from_spec_action emergency fallback
- **verify-later:** webdesign-agent workflow step order.

<!-- SOURCE: U12_docs024_archives.md -->
### Fork_theme step double-creation guard
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** "Guard: require both should_fork_theme AND should_promote_to_library flags. Implementation deferred to Deliverable 6."
- **what:** Once site-design-planner runs, the pre-existing `fork_theme` step in webdesign-agent risks creating duplicate theme/collection rows. Documented mitigation requires two flags both true before forking proceeds.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"6. Risks Still Live"
- **relations:** composition resolution architecture; fork_theme_from_site rewrite
- **verify-later:** `fork_theme` step config in webdesign-agent — confirm both flags gate execution.

<!-- SOURCE: U12_docs024_archives.md -->
### Design pipeline guiding principles (mottos)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "Principles Restated" section repeated verbatim across 2026-04-19 handoffs, sourced from `007_adoption_pipeline_v2.md` and a FOCUS work-plan doc.
- **what:** A shared decision-shorthand invoked to settle scope questions: "Every build conceptually an adoption," "Design reference is history, design intent is direction," "Adoption is a starting point, not a ceiling," "LLM for reasoning, Go for extraction," "Handlers are self-contained," "Slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"7. Principles Restated"
- **relations:** site-design-planner Choice B scope; design_reference/design_intent split
- **verify-later:** none — a documentation/culture artifact, not directly code-verifiable.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Colour Inheritance Model
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** every layout_NN header lists this as CONTRACT CHECK #1 and the CSS body implements it identically (e.g. layout_01 lines ~160-182: `h1..h6 { color: var(--section-heading, var(--color-primary)); }`)
- **what:** The rule that element-level colour rules (headings, body text, links) resolve via a two-tier CSS custom-property fallback chain: `var(--section-*, var(--color-*))`. This lets a "dark section" component override just the `--section-*` variable on its own container without any layout needing to restate rules. Applied identically across all 17 layout templates as CONTRACT CHECK #1.
- **sources:** layouts/layout_01_brochure-formal.sql#header+L160-182, layouts/layout_02_brochure-bold.sql#header, layouts/layout_16_17_vonc_gamesdesign.sql#L832, layouts/layout_10_high-energy.sql#header
- **relations:** Dark Section Variable Contract; template helper system; renderer-managed surface sections
- **verify-later:** render_css_from_spec_action.go; every layout's `:root` and base element rules

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dark Section Variable Contract / buildSectionDefaults renderer behaviour
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** layout_01 lines 268-289: "TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync with buildSectionDefaults in render_css_from_spec_action.go... Tracked as Phase 4.5 in 025_palette_layout_typography_migration."
- **what:** Layout templates must NOT declare `--section-*` defaults on section containers; a Go renderer function `buildSectionDefaults` appends `--section-*` overrides after rendering, chosen by palette luminance. Five renderer-managed surface classes (`.features-section`, `.services-section`, `.differentiators-section`, `.about-section`, `.faq-section`) are hardcoded on both sides and must be kept in sync; hero/CTA/testimonials/contact are excluded as component-owned. One documented exception: a palette-declared `heading` slot emits a root-level `--section-heading`.
- **sources:** layouts/layout_01_brochure-formal.sql#L14-32,L268-289, layouts/layout_02_brochure-bold.sql#header, layouts/layout_16_17_vonc_gamesdesign.sql#L85-93
- **relations:** Colour Inheritance Model; template helper system; layout archetype concepts (all 17)
- **verify-later:** render_css_from_spec_action.go buildSectionDefaults; docs025/026/027 Phase 4.5 status

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Template helper system ({{palette}}/{{typo}}/{{token}} with fallback)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** every layout file's CONTRACT CHECK #4; 003_palettes_seed.sql: "Key naming... matches the {{palette \"primary_hover\" \"...\"}} template helpers"
- **what:** A Go-template-style substitution convention embedded in the `css_template` CSS text: `{{palette "key" "fallback"}}`, `{{typo "key" "fallback"}}`, `{{token "key" "fallback"}}`, each resolving a JSONB slot lookup with a mandatory literal fallback. A `{{with palette "heading" ""}}...{{end}}` conditional-block variant is also used.
- **sources:** layouts/layout_01_brochure-formal.sql#L33-34,L89-138, layouts/003_palettes_seed.sql#L14-19, layouts/layout_16_17_vonc_gamesdesign.sql#L96-145
- **relations:** Colour Inheritance Model; structure_tokens JSONB convention; palettes table; typography_sets table
- **verify-later:** the Go renderer executing these templates; helper lookup precedence (site-adopted vs seed palette)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### structure_tokens JSONB convention
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** present as a populated JSONB literal column in the INSERT of all 17 layout seed files
- **what:** Each layout row carries a `structure_tokens` JSONB column holding non-colour design tokens — spacing, radii, shadows, transitions, and layout-specific one-offs (e.g. `diagonal_slope_top` for high-energy, `split_pane_left/right` for tool-first-landing). The layout-level counterpart to the palette/typography tables; explicitly excluded from palette extraction.
- **sources:** layouts/layout_01_brochure-formal.sql#L55-69, layouts/layout_10_high-energy.sql#L38-53, layouts/003_palettes_seed.sql#L39-41
- **relations:** template helper system; palettes table; per-layout archetype concepts
- **verify-later:** `layouts` table column `structure_tokens` DDL/constraints in Phase 2 migration

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Seed-driver transactional load pattern
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 003_layouts_seed_driver.sql: `BEGIN;` ... 15 `\ir` includes ... verification block asserting `actual_count >= 15` ... `COMMIT;`
- **what:** A psql driver script wrapping all 15 numbered layout `\ir` includes in a single transaction with `\set ON_ERROR_STOP on`, so any single layout's error rolls back the entire batch. Ends with a `DO $verify$` block raising an exception if the seeded row count is below expected. Each INSERT is itself idempotent (`ON CONFLICT (name) DO UPDATE`).
- **sources:** layouts/003_layouts_seed_driver.sql (full file), layouts/003_palettes_seed.sql#verify block, layouts/003_typography_sets_seed.sql#verify block
- **relations:** palettes table/seed; typography_sets table/seed; all 15 numbered layout archetypes
- **verify-later:** Phase 2 migration creating palettes/layouts/typography_sets tables; confirm this driver actually ran against the live DB

<!-- SOURCE: U13_docs024_small_dirs.md -->
### palettes table / seed (CSS-theme-extracted colour slots)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Diagnostic run (Phase 3 preflight) confirmed... 13 rows have palette data in css_content"; verify block asserts `actual >= 13`
- **what:** `palettes` table stores one row per design palette (`name`, `display_name`, `colours` JSONB slot map, `category`, `industry_tags`, `origin`, `is_active`). The seed migrates 13 legacy `css_themes` rows via a PL/pgSQL helper `_extract_css_palette` that regex-parses `--color-KEY: VALUE;` declarations. Non-colour vars are deliberately excluded (belong to structure_tokens). One theme (`standard-brochure`) has no palette of its own, mapped to `default` in a later step.
- **sources:** layouts/003_palettes_seed.sql#header,#_extract_css_palette function,#insert+select,#verify+report
- **relations:** template helper system; structure_tokens JSONB convention; typography_sets table; css_themes (legacy source table)
- **verify-later:** `css_themes` table; confirm the "Phase 3 Step 3" theme-mapping UPDATE actually ran

<!-- SOURCE: U13_docs024_small_dirs.md -->
### typography_sets table / seed (6 named font/scale bundles)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Seeds the 6 typography sets described in the migration plan section 8"; verify block asserts `actual_count >= 6`
- **what:** `typography_sets` table stores 6 named typography bundles — sans-modern, serif-editorial, display-bold, mono-technical, serif-classical, sans-friendly — each with `fonts` JSONB and `scale` JSONB, plus `category`/`industry_tags`. Layouts reference these via `{{typo "key" "fallback"}}`. Each set's description names which layout archetypes it pairs with (documented convention, not FK-enforced).
- **sources:** layouts/003_typography_sets_seed.sql#header,#sans-modern,#display-bold,#mono-technical,#serif-classical/sans-friendly,#verify
- **relations:** palettes table; template helper system; structure_tokens JSONB convention
- **verify-later:** confirm each layout's declared "Default typography" matches typography_sets.name at composition time

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout-resolution-by-tags gap (resolveLayoutByTags fallback problem)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** layout_16_17_vonc_gamesdesign.sql header: "Root cause reminder: resolveLayoutByTags matches tagSet against layout.industry_tags. The classifier doesn't currently emit those two fields, so tagSet is empty → fallback path 'no classification tags'... Neither migration alone is sufficient."
- **what:** The site-design-planner's layout picker (`resolveLayoutByTags`) intersects a site's classification tags against each layout row's `industry_tags`; when the classifier doesn't emit those fields, every site falls back to `brochure-formal` regardless of fit — exactly what happened to gamesdesign.co.uk. Fixing it requires two coordinated migrations: this file (007) seeding the missing layouts, and a separate 008 migration (not in scope) updating the classifier prompt to emit category/industry_tags.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L1-38
- **relations:** tool-portal-dark; social-lobby; all 15 numbered layouts (as matching candidates)
- **verify-later:** migration "008" classifier prompt; `resolveLayoutByTags` function location; live check whether sites still fall back to brochure-formal

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: brochure-formal
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phase 1 of 025_palette_layout_typography_migration"; satisfies all 5 numbered CONTRACT CHECKS
- **what:** Structured, understated, CTA-driven brochure layout with corporate restraint. Mapped to themes `default`, `standard-brochure`, `professional-dark`. Suits consultancies, law, finance, B2B. Serves as the canonical reference implementation of all 5 contract checks and as the fallback layout when tag resolution fails.
- **sources:** layouts/layout_01_brochure-formal.sql#L1-37,L50-69
- **relations:** Colour Inheritance Model; Dark Section Variable Contract; brochure-bold; Layout-resolution-by-tags gap
- **verify-later:** `layouts` row name='brochure-formal'; confirm still the de facto fallback

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: brochure-bold
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phase 3 of 025_palette_layout_typography_migration"
- **what:** High-energy conversion variant of brochure-formal — tall hero, gradient accents, display-bold typography, strong CTAs. Suits tech startups, SaaS, fitness brands.
- **sources:** layouts/layout_02_brochure-bold.sql#L1-30,L43-65
- **relations:** brochure-formal; Dark Section Variable Contract; typography_sets(display-bold)
- **verify-later:** `layouts` row name='brochure-bold'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: portfolio-kinetic
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header explicit "STRUCTURAL DIVERGENCE from brochure-* layouts" list; "Mapped themes: none currently"
- **what:** Asymmetric, motion-forward, display-type-led layout for creative-studio energy — animated underline text-links instead of hero/CTA buttons, 40/60 asymmetric columns, dense-packed work showcase, narrower 1140px container. Suits design studios, creative agencies, photography portfolios.
- **sources:** layouts/layout_03_portfolio-kinetic.sql#L1-33,L46-66
- **relations:** brochure-formal (contrast); typography_sets(serif-classical alt); Colour Inheritance Model
- **verify-later:** `layouts` row name='portfolio-kinetic'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: magazine-grid
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: content-modern"
- **what:** Publication-feel layout: top-level 2/3 main + 1/3 sidebar grid, article cards, featured-article variant, sidebar widgets, serif-editorial typography. Suits news, opinion, long-form blogs.
- **sources:** layouts/layout_04_magazine-grid.sql#L1-35,L37-70
- **relations:** typography_sets(serif-editorial); soft-editorial; industry-hub
- **verify-later:** `layouts` row name='magazine-grid'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: utility-tool
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: none — exists for selector/adoption matching"
- **what:** Minimal-chrome layout where "the tool is the reason" — narrowest container (800px), compact header, single tool card with output region, no card-grids, larger form controls. Suits online calculators, converters, developer utilities.
- **sources:** layouts/layout_05_utility-tool.sql#L1-25,L27-59
- **relations:** tool-first-landing (explicit divergence); typography_sets(sans-modern)
- **verify-later:** `layouts` row name='utility-tool'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: media-grid
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: none"; dark-mode-by-default palette
- **what:** Thumbnail-dominant, continuous-scroll discovery layout — auto-fill fluid grid, optional featured/pinned item, scrollable chip filter bar, "featured row"/horizontal-scroll shelf variants, fixed aspect-ratio tokens. Suits video platforms, audio libraries, image galleries. Dark theme by default.
- **sources:** layouts/layout_06_media-grid.sql#L1-24,L26-58,L67-90
- **relations:** high-energy; tool-portal-dark; Colour Inheritance Model
- **verify-later:** `layouts` row name='media-grid'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: docs-sidebar
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Default typography: mono-technical" matches typography_sets seed row's own note
- **what:** Reference-grade documentation layout — 3-zone CSS grid (fixed sidebar nav, main reading column, collapsing table-of-contents). Code blocks get accent-border + copy-button; admonitions use `.callout` variants. Suits developer docs, API references, knowledge bases.
- **sources:** layouts/layout_07_docs-sidebar.sql#L1-25,L27-58
- **relations:** typography_sets(mono-technical); tool-portal-dark
- **verify-later:** `layouts` row name='docs-sidebar'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: soft-editorial
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: bakery, warm-friendly, calm-minimal, soft-editorial" — the only numbered layout with 4 named theme mappings
- **what:** Warm, reading-first, organic layout — tinted background, pill-shaped buttons, barely-there card borders, serif display headings, transparent floating header, 1.75 line-height. Suits wellness blogs, lifestyle sites, personal essays, bakeries.
- **sources:** layouts/layout_08_soft-editorial.sql#L1-23,L25-57
- **relations:** typography_sets(serif-editorial, sans-friendly); magazine-grid; industry-hub
- **verify-later:** `layouts` row name='soft-editorial'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: technical-precise
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: premium-elegant (with serif override), modern-engineering-clean"
- **what:** "Engineered" layout — glass-effect header (backdrop-filter blur) as its signature moment, tight border-radius, bordered/low-shadow cards, flat solid CTAs, light (not dark) footer contrasted against brochure-*'s dark footers. Suits SaaS platforms, infrastructure products, engineering consultancies.
- **sources:** layouts/layout_09_technical-precise.sql#L1-25,L27-58
- **relations:** typography_sets(sans-modern default, serif-classical override); brochure-formal (footer contrast)
- **verify-later:** `layouts` row name='technical-precise'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: high-energy
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: boxing" (narrowest mapping of all 15)
- **what:** Aggressive, kinetic layout — uppercase headings, 80vh dark hero, diagonal clip-path section separators, zero border-radius, hard offset shadows, numeral-prefixed feature cards. Suits boxing gyms, combat sports, fitness events. Uses display-bold typography.
- **sources:** layouts/layout_10_high-energy.sql#L1-20,L22-53
- **relations:** typography_sets(display-bold); media-grid
- **verify-later:** `layouts` row name='high-energy'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: comparison-aggregator
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header distinguishes itself from 3 sibling commerce-adjacent layouts by its defining primitive `.result-card`
- **what:** Search-first, data-dense, trust-oriented layout — hero IS a search input, sticky filter bar, dense horizontal result-card rows, regulatory info banners, heavy disclaimer footer. First of four deliberately-differentiated "commerce-adjacent" layouts. Suits price/insurance/broadband comparison, trade directories.
- **sources:** layouts/layout_11_comparison-aggregator.sql#L1-24,L26-60
- **relations:** affiliate-hub; ecommerce-storefront; industry-hub; tool-first-landing
- **verify-later:** `layouts` row name='comparison-aggregator'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: affiliate-hub
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's explicit divergence table against comparison-aggregator and ecommerce-storefront
- **what:** Product-review/buyer-guide layout — persistent disclosure strip, vertical product "picks" cards, pros/cons review blocks, horizontally-scrolling comparison tables, optional sticky "Top Picks" sidebar. Suits product review sites, "best X for Y" guides, deal aggregators.
- **sources:** layouts/layout_12_affiliate-hub.sql#L1-21,L23-56
- **relations:** comparison-aggregator; ecommerce-storefront; industry-hub
- **verify-later:** `layouts` row name='affiliate-hub'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: ecommerce-storefront
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's divergence note vs affiliate-hub (cover-fit lifestyle photography vs contain-fit product-on-white)
- **what:** Retail-clean, product-forward storefront — promo hero, image-overlay category tiles, product grid, add-to-cart CTAs, strike-through sale pricing, CSS-only mini-cart dropdown structure, trust-bar strip. Suits independent shops, small-catalogue retailers.
- **sources:** layouts/layout_13_ecommerce-storefront.sql#L1-24,L26-60,L94-97
- **relations:** affiliate-hub; comparison-aggregator
- **verify-later:** `layouts` row name='ecommerce-storefront'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: tool-first-landing
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's explicit divergence from utility-tool (full-container vs 800px narrow column)
- **what:** Full-container (up to 1400px) tool-dominated landing page where "the tool IS the page" — defining primitive `.split-pane` (50/50 default), dark-mode-friendly, optional tabbed interface. The "loud" counterpart to utility-tool's contained/quiet version. Suits calculators, API playgrounds, demo tools.
- **sources:** layouts/layout_14_tool-first-landing.sql#L1-22,L24-56
- **relations:** utility-tool; tool-portal-dark
- **verify-later:** `layouts` row name='tool-first-landing'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: industry-hub
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's 4-way divergence table naming this the only non-commercial member of the "commerce-adjacent" family
- **what:** Vertical information-authority layout — "About this site" independence-claim banner, `.directory-card`/`.guide-card`/`.news-card`/`.glossary-list` primitives, ordered directory→guides→news→reference, serif-editorial typography for "authority without being corporate." Suits regulatory information hubs, industry explainer sites.
- **sources:** layouts/layout_15_industry-hub.sql#L1-28,L30-61
- **relations:** comparison-aggregator; affiliate-hub; ecommerce-storefront; magazine-grid/soft-editorial
- **verify-later:** `layouts` row name='industry-hub'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: tool-portal-dark
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** seeded by migration "007_seed_layouts_tool_portal_and_social_lobby.sql" explicitly framed as necessary-but-not-sufficient; `needs_review` column present on the INSERT
- **what:** Dark developer-utility portal layout supporting three page shapes in one template — portal/index, tool pages, article/guide pages (narrow reading column). Dark-mode-first, flat technical aesthetic. Built specifically to close the layout-library gap that caused gamesdesign.co.uk to fall back to brochure-formal.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L1-38,L55-145,L71-94
- **relations:** Layout-resolution-by-tags gap; social-lobby; docs-sidebar; media-grid
- **verify-later:** `layouts` row name='tool-portal-dark', `needs_review` flag value; migration "008"

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: social-lobby
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** same migration-007 framing as tool-portal-dark; `needs_review` column present
- **what:** Light, colour-forward social-platform layout built around a room/lobby metaphor. Primary UI unit is the "provocation card"; Arena (competitive) and Stage (creative) rooms differentiated via dedicated palette slots (`arena`, `stage`) rather than component variants. Four page shapes: lobby/homepage, room/topic index, provocation detail, archetype/profile. Reaction-colour slots (`reaction_positive`/`reaction_negative`/`reaction_meta`) are a distinctive palette extension. Named target: vonc.com.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L21-23,L713-757,L759-810
- **relations:** Layout-resolution-by-tags gap; tool-portal-dark; vonc workstream (site-case-studies); palettes table (arena/stage/reaction slots)
- **verify-later:** `layouts` row name='social-lobby'; live check against vonc.com

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Design fingerprint & design_reference vs design_intent
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 007 "Design Fingerprint Pipeline (added 2026-04-12)"; principle 7 "Design reference is history, design intent is direction"
- **what:** A Go extractor (`extract_design_fingerprint`, goquery) parses crawled rawHTML/external CSS into a fingerprint with a `suggested_mapping`; an LLM (`generate_design_intent`) turns it into a semantic brief. `design_reference` is an immutable historical record; `design_intent` is forward-looking direction — evolution happens by updating intent, never reference.
- **sources:** WM/007_adoption_pipeline_v3.md#design-fingerprint-pipeline-added-2026-04-12, WM/007_adoption_pipeline_v3.md#design-evolution-lifecycle, WM/FOCUS_interactive_content_generation(3).md#adoption-captures-content-and-extracts-structured-design-data
- **relations:** site adoption agent; interactive parse-stage; webdesign-agent three-way priority
- **verify-later:** enrich_fingerprint_with_css_action.go; site_specs design_reference/design_intent

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Component selector + creator (section_type vs function)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 007 "Phase 3 — Component selector, patterns, and research" (planned); "The separation: Planner decides WHAT section types … Component selector decides WHICH template"
- **what:** Splits the planner's conflated role: the planner picks section_types, a Go component-selector scores templates by metadata with a fallback to `needs_new_component`, and a `component-creator` agent LLM-generates a template from the full component contract when none fits. `function` currently does two jobs (page-role identifier + template choice); `section_type` separates them.
- **sources:** WM/007_adoption_pipeline_v3.md#component-selector-and-creator, WM/007_adoption_pipeline_v3.md#component-creation-contracts, WM/FOCUS_interactive_content_generation(3).md#components-more-broadly
- **relations:** interactive content generators; site plan sections; tool/game library model
- **verify-later:** content_components metadata columns; component-creator agent; plan_sections

<!-- SOURCE: U18_sql_for_agents.md -->
### chief-strategist (build-plan LLM) + component placement dedup rules
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** 040 still upgrades its model (haiku→sonnet) but the work-item pipeline planner (build-site-planner, 053) owns planning thereafter; 019 patch injects "COMPONENT PLACEMENT RULES" into its prompt.
- **what:** The v1/v2 planning agent producing sections/component_details build plans. Its lasting contribution is the component placement rule-set injected by 019: testimonials/team-grid/faq/contact-form on ONE page only, per-page hero variants, no duplicated services content, merge similar pages — an early anti-repetition contract for planners.
- **sources:** 019_chief_strategist.sql; sql_for_agents_v1/019_chief_strategist.sql; sql_for_agents_v2/019_chief_strategist.sql; 040_optimise_which_llms.sql
- **relations:** site-planner, build-site-planner inherit the planning role; parse_json_field/unwrapDeep pattern (v1/019)
- **verify-later:** is chief-strategist still active or deleted

<!-- SOURCE: U18_sql_for_agents.md -->
### webdesign-agent (full CSS stylesheet generation)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 4,683-line definition file; referenced as the "full theme" path by 076 ("Unlike webdesign-agent (which regenerates everything from scratch)"); idle timeout in 075.
- **what:** Generates production CSS for a site. Accepts a provided site_context or loads context from DB (conditional first step), analyzes design requirements, writes stylesheet via git_commit with file_path config. It is the heavyweight regeneration path, contrasted with css-patch-agent for targeted fixes.
- **sources:** 031_webdesign_agent.sql; 076_css_patch_agent.sql; 103_site_design_planner.sql (resolved_composition reader list)
- **relations:** site-scraper feeds it site_context; css themes/style_collections; site-design-planner
- **verify-later:** current webdesign-agent workflow vs 031 copy; patch_01_git_commit_file_path.go

<!-- SOURCE: U18_sql_for_agents.md -->
### site-design-planner spec aspects (navigation / layout / resolved_composition)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 103 is "Deliverable 2: Spec schemas... pre validation" — documents shapes and creates best-effort validation functions, explicitly not table constraints; reader lists name live actions.
- **what:** Defines three site_specs aspects the site-design-planner writes, separated by reader: navigation (nav architecture, items, CTA, mobile pattern → populate_nav_tables, InjectHeader, GetNavItems), layout (page-level layout, header/footer style → AssembleMultipageSiteAction, templates), resolved_composition (machine-readable pointers to palette/layout/typography rows + reasoning → render_css_from_spec, webdesign-agent, audit agents). Validation functions run at write time; site_specs stays open JSONB.
- **sources:** 103_site_design_planner.sql
- **relations:** design-composition docs 025/026/027; webdesign-agent; nav-updater
- **verify-later:** site-design-planner agent existence and writers of these aspects

<!-- SOURCE: U19_sql_tables_components.md -->
### Style collections
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Two generations of the migration in 001 (initial + 030_style_collections), sites.style_collection_id FK, seed collections professional-dark / minimal-light / bold-gradient with linked header/footer components.
- **what:** A style collection bundles the components and tokens defining a site's visual identity: header/header-home/footer component ids, css_theme_id, color_palette and typography JSONB, category and industry_tags. Sites link to one collection and may override via sites.style_overrides without forking the collection. Original motivation: replace inconsistent LLM-generated headers with tested templates.
- **sources:** docs/agent_docs/sql_for_components/001_style_collections.sql; docs/agent_docs/sql_for_components/003_styles_implementation.md; docs/agent_docs/sql_for_components/002_styles_documentation.md
- **relations:** component-based headers; palette/layout/typography decomposition; design lineage columns.
- **verify-later:** style_collections rows; assignment logic in EnsureSiteRecordAction / classification.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component-based headers replacing LLM-generated chrome
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002/003 md docs lay out the plan (store tested header templates, render with site data, inject replacing LLM header); 012 executes population and SQL-side rendering of site_components for header/footer/head.
- **what:** The founding decision that page chrome (header/footer/head) is never LLM-generated per page: tested templates render with a site-derived context (logo from domain, nav from pages/nav tables, colours from collection+overrides) and are injected at assembly. Benefits table: consistency, instant DB-side updates, A/B-able collections.
- **sources:** docs/agent_docs/sql_for_components/002_styles_documentation.md; docs/agent_docs/sql_for_components/003_styles_implementation.md; docs/agent_docs/sql_for_tables/012_site_components.sql
- **relations:** style collections; site/area/page component hierarchy; template syntax unification.
- **verify-later:** RenderHeaderForSite / render_site_components action.

<!-- SOURCE: U19_sql_tables_components.md -->
### Palette / layout / typography decomposition (migration 025 phase 2)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Three new tables, empty after this migration... new columns are read only once Phase 4 ships. Phase 3 seeds... Phase 7 drops the legacy columns" (038 header).
- **what:** Splits css_themes.css_template's conflated concerns into three independently versioned tables: palettes (free-shape colours JSONB consumed via {{palette "key" "fallback"}}), layouts (Go CSS template + structure_tokens + default header/footer component ids), typography_sets (fonts + scale via {{typo}}). css_themes becomes a composition row via nullable FKs; renderer migrates in later phases. Also created 10 library layout components (header-with-categories, header-docs, directory-listing, product-grid, etc.).
- **sources:** docs/agent_docs/sql_for_tables/038_style_collections.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components
- **relations:** style collections; design lineage; site_plan_sections resolved palette/layout/typography ids.
- **verify-later:** palettes/layouts/typography_sets row counts; renderer read path (phase 4); legacy column drops (phase 7).

<!-- SOURCE: U19_sql_tables_components.md -->
### Design-asset fork lineage (origin / needs_review / source_site)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 038 Part 2(b): lineage columns "required by the Phase 5 fork_theme_from_site action. A prior session reported them as already added but the current schema shows them absent... nothing needs review (fork action hasn't shipped yet)".
- **what:** Uniform provenance on palettes, layouts, typography_sets, css_themes and style_collections: origin ('seed' default), needs_review, forked_from_<entity>_id, source_site_id, source_domain, forked_at. Enables adopting a live site's design into the library as a reviewed fork.
- **sources:** docs/agent_docs/sql_for_tables/038_style_collections.sql#PART2
- **relations:** adoption-pipeline (design adoption); tool fork model (same pattern for tools).
- **verify-later:** fork_theme_from_site action existence; any rows with origin != 'seed'.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Brand designer agent (theme selection)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Agent SQL + mvp-site-builder workflow insertion (spawn/call_brand_designer feeding brand_theme to the architect); superseded by content-creator's theme recommendation + semantic tag matching in 006semantic_themes, and later by the design-composition system.
- **what:** An LLM agent that analyses domain + objective and picks a CSS theme from the named library (boxing, bakery, tech, professional-dark, default) with reasoning — the first brand/design decision point in the pipeline.
- **sources:** docs004_website_capture_project/website_analysis/README.018.brand_designer_agent.md
- **relations:** semantic CSS theme system; successor: site-design-planner / palette resolution (design-composition docs 025-027).
- **verify-later:** brand-designer agent_definitions row.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Semantic CSS theme and snippet system (theme_tags, css_themes, css_snippets, js_snippets)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Full DDL + seed data in two iterations (020 text[]; 027 jsonb) with helper matching functions; themes are complete `:root` CSS-variable palettes. The design-composition palette/typography system is the taxonomy-named successor.
- **what:** A semantic tagging vocabulary (mood/style/industry/audience/functional/colour tags with related_tags pairing) applied to: css_themes (full CSS-variable palettes: calm-minimal, bold-conversion, warm-friendly, dark-modern, premium-elegant…), css_snippets (hover/animation/effect/pattern/utility fragments), and js_snippets (nav, scroll animations, accordion, clipboard, form interactions with trigger metadata). Content-creator recommends theme + theme_tags; assembler matches snippets by tags. All theming via CSS variables — the ancestor of the platform's CSS-variable contract.
- **sources:** docs004_website_capture_project/006semantic_themes/README.020.brand_theme_preparation.md; docs004_website_capture_project/007different_types_of_site/027_css_js_schema.sql; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md
- **relations:** successors: contracts-and-standards CSS variables; design-composition palette resolution; styling-render-pipeline.
- **verify-later:** css_themes/css_snippets/js_snippets/theme_tags tables today.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Style collections as the design bridge
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** docs017/017 three-layer model (HTML components / CSS theme / style collection "the bridge"); docs012/009 migration 030_style_collections_migration.sql; per docs015/004 "load style collections" is a standard planner step.
- **what:** Layer 3 of the design system: a style_collection binds a site to specific header/footer/head component choices plus a CSS theme (colors, typography), selected per site (stored on sites, or chosen by domain keywords as fallback). Enables mix-and-match of structure and appearance and consistent chrome across the multipage path. Ancestor of the current palette/typography/layout resolution system.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/017_agent_architecture_v2.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Design-System-Layers; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** css_themes; webdesign-agent; design agent family split; current design-composition docs 025-027.
- **verify-later:** style_collections table shape and GetStyleCollectionForSite.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Design agent family split (brand-designer / style-generator / layout-architect)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** docs017/019b: webdesign-agent "Exists, prompt updated for colour inheritance"; brand-designer/style-generator "Future split"; layout-architect/nav-layout-agent "New" (never appear later); "There's no rush on this split."
- **what:** Decompose the monolithic webdesign-agent (analyse_design → generate_css → update_site, deploying /assets/css/styles.css) into: brand-designer producing a rarely-changing brand_spec (palette, type scale, spacing, tone, image direction) in sites.content_data; style-generator producing CSS with theme-library search-and-adapt before generating fresh (feeding css_themes for reuse); layout-architect producing per-page-type layout definitions (nav placement, content zones, max components) with rendering fallbacks. Direct ancestor of the current site-design-planner / design composition system.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#3-Design-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/003_design.md
- **relations:** style collections; colour inheritance; current design-composition docs 025-027 (successor).
- **verify-later:** brand_spec/layout_definitions keys in sites.content_data; whether split agents exist.

<!-- SOURCE: U22_recent_small_docs.md -->
### Vertical-specific planner variants
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** Phase 3.5 todo: "Create veterinary/energy/mortgage/seasonal site planner prompt variant" — all unchecked.
- **what:** Separate agent definitions using the same planner Go code but vertical-tuned prompt templates, so a well-established vertical produces better plans than a generic planner with config injected. Each knows its vertical's page types, conversion funnel, and per-page guidance (e.g. every breed-health page links to "find a vet for this breed"; every mortgage calculator has lead capture below results).
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#3.5
- **relations:** site-planner, vertical knowledge architecture, unified site spec
- **verify-later:** agent_definitions for veterinary/energy/mortgage/seasonal site-planner variants

<!-- SOURCE: U23_docs_root_vonc.md -->
### Post-025 CSS theme flow (empty css_content by design; composition via FK chain)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-24 ~16:00 citing doc 027 and install_site_composition_action.go L210-212: css_content "intentionally empty — post-025 renderer reads composition via FK chain at render time"; styles.css deployed by webdesign-agent.
- **what:** The design pipeline runs needs_composition (site-design-planner) → gated needs_design (webdesign-agent: analyze_design → update_site → generate_css via render_css_from_spec reading composition FKs → deploy_css writes assets/css/styles.css → optional fork_theme). `css_themes.css_content` is intentionally empty post-025; the empty "Theme-specific styles injected here" head block is expected, not a bug. webdesign-agent is not deprecated. Key debugging consequence: a wrong colour on a page is more likely a component variable-name mismatch than a theme-injection failure.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:00; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** CSS variable naming; two chrome assembly paths (stale renders)
- **verify-later:** install_site_composition_action.go; render_css_from_spec

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Design fingerprint extraction (adoption fidelity)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** FOCUS_design_and_styling_adoption_problems (gamedesign.uk case): "rawHTML is captured but never parsed for design data"; WORK_PLAN_v2 (2026-04-11) Phase 1 `extract_design_fingerprint` Go action "Code ready — needs deploying", `design_reference` spec replacing vague `design`; live successors 025/026/027 design docs exist.
- **what:** Mechanism to parse crawled rawHTML `<style>` blocks, CSS variables, Google-Fonts links, and layout into a concrete `design_reference` spec aspect (hex values, font stacks, `suggested_mapping` from source→our CSS variable names), replacing the LLM's guessed `design` spec so adopted sites reproduce the original's colours/fonts/layout instead of generic component defaults.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_problems.md#a-fingerprint-extraction; FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#phase-1
- **relations:** superseded by 025_palette_layout_typography_migration / 027_design_and_site_planner_v2; design_intent; adoption pipeline
- **verify-later:** extract_design_fingerprint_action.go; site_specs aspect design_reference

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Three-way webdesign priority (design_intent → design_reference → industry)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** WORK_PLAN_v2 Decisions #3: "design_intent (creative freedom) → design_reference (reproduce faithfully) → generate from industry (new builds)"; palette "locked until design_intent is written"; Phase 2b "Applied to DB".
- **what:** The webdesign-agent prompt resolves imagery/palette by priority: honour `design_intent` (semantic creative brief, auto-generated from `design_reference` via LLM in adoption Phase 2e) first, else reproduce `design_reference` faithfully, else generate from industry. The palette can only change once design_intent exists.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#decisions-made, #phase-2
- **relations:** superseded by live design-composition docs 025/026/027; design fingerprint; imagery_direction
- **verify-later:** webdesign-agent analyze_design prompt in agent_definitions

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Site-design-planner agent (structure × identity × effects)
- **category:** design-composition
- **status-signal:** abandoned
- **status-evidence:** WORK_PLAN_v2 Phase 3 "Site-Design-Planner Agent (not started)" — all of 3a-3g "Not started"; Phase 4 requirement-driven components also "Not started"; superseded by live 027_design_and_site_planner_v2.md.
- **what:** Proposed dedicated `site-design-planner` agent (Option B) decomposing site design into structure × identity × effects, owning navigation/layout spec schemas, wired into the build pipeline to drive header/footer selection and hero/nav merging — plus requirement-driven component selection generating custom components when the library has no match. Never built as specified; replaced by the v2 design/site-planner architecture.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#phase-3, #phase-4
- **relations:** replacement = 027_design_and_site_planner_v2.md; site_plan tables; component library
- **verify-later:** agent_definitions for site-design-planner (likely absent); 027 live doc

<!-- SOURCE: U25_leopardess_social.md -->
### Per-site style fork chain (palette → css_theme → style_collection)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** HANDOFF §3: "Palette — forked … seed 3196d966 untouched, still dresses 3 other sites. Deployed styles.css matches the validated palette exactly" (2026-07-10/12).
- **what:** Safe per-site restyling: clone palettes + css_themes + style_collections rows (reusing seed layout/typography/header/footer), repoint sites.style_collection_id, never edit the shared seed collection that dresses multiple sites. The leopardess fork carries the two-tone gold system (A10): bright #C8A951 only on dark chrome (8.56:1), bronze #836E32 for links on light (bright gold fails AA at 2.1:1 on light). Header component forked too (header-professional-dark hardcodes navy with zero CSS variables across 4 sites) — a site_components/collection-wired fork sticks where a section fork does not.
- **sources:** docs/leopardessconsulting/scripts/L3_fork_palette.sql (header); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10, #Turn-12; docs/leopardessconsulting/RUNBOOK.md#O10
- **relations:** core-vs-specialised slot merge semantics; specialised-slot contrast gap; section resolver override behaviour
- **verify-later:** style_collections/palettes rows for leopardess; fork_theme_composition.go / install_site_composition

<!-- SOURCE: U25_leopardess_social.md -->
### Deterministic contrast gate missing on specialised palette slots
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES turn 10: "nothing stops a fork shipping an inaccessible palette — the WCAG primitives exist but aren't called at generation/fork/install/render for specialised slots."
- **what:** color_util.go has correct WCAG code (relativeLuminance, wcagContrastRatio, pickReadableOnBackground) but it is wired only to section-text defaults (loose 3.0/2.0) and forced-text-colour stripping (AA 4.5); the specialised slots (card_bg, header_bg, cta_bg/cta_text) — the exact slots that leaked white cards/navy chrome/blue CTA — are never contrast-gated. Adding the gate is small; validation is currently done by hand (all 15 reader-experienced pairs checked with the platform's own formula).
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; docs/leopardessconsulting/HANDOFF.md#8; docs/leopardessconsulting/RUNBOOK.md#O10
- **relations:** per-site style fork; styling-render-pipeline slot merge
- **verify-later:** color_util.go call sites; whether any generation/fork path calls wcagContrastRatio on specialised slots

<!-- SOURCE: U25_leopardess_social.md -->
### `analyze_design` requires palette.reference_values (else the LLM invents a palette)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 12: "the analyze_design LLM step INVENTED a dark core … Fix: restructured design_intent into palette.reference_values + prescriptive guidance … Re-rendered → all slots now exactly match."
- **what:** The webdesign-agent's analyze_design LLM reads colours only from design_intent.palette.reference_values (not color_scheme); without prescriptive values there ("these eight values are FIXED, output verbatim") it improvises from the mood text under explicit creative freedom. Same pattern applied for typography reference_values. The leopardess design_intent JSON is the worked example of the contract.
- **sources:** docs/leopardessconsulting/specs/design_intent.json#palette; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12; docs/leopardessconsulting/HANDOFF.md#4.6
- **relations:** per-site style fork; core-vs-specialised slot merge
- **verify-later:** webdesign-agent workflow analyze_design step; render determinism with empty design_spec

<!-- SOURCE: U25_leopardess_social.md -->
### Three-per-row no-orphan grid rule as a content fix
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** L5_homepage.sql header (2026-07-10): "card grids are 3-up (no orphan row), per the brief. That is a CONTENT fix, not a CSS one — the grid components are shared across 5 sites."
- **what:** Neither global `repeat(3,1fr)` nor per-component `auto-fit,minmax()` avoids orphan/stretched last cards; the durable rule is card counts divisible by three (which also forces cutting panels that repeat each other) because grid component CSS is shared and untouchable. case-studies-grid is hard-wired to five cards and cannot be 3-up. Encoded in the design_intent layout_preference ("if the content does not divide into threes … two of the cards are saying the same thing").
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L4; docs/leopardessconsulting/scripts/L5_homepage.sql, L5_pages.sql (headers); docs/leopardessconsulting/specs/design_intent.json#layout_preference
- **relations:** shared component library semantics; anti-hype voice (repetition cut)
- **verify-later:** grid component CSS and shared usage counts

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Design system three layers (content_components / css_themes / style_collections)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4) opening section; live tables
- **what:** Layer 1 HTML components (self-contained blocks, inline style, CSS variables with fallbacks, never hardcode brand colours); Layer 2 CSS theme (one styles.css per site rendered from installed composition); Layer 3 style_collections bundling header/footer components + theme + palette/typography. Sites reference via sites.style_collection_id.
- **sources:** 002(4)#Design System Layers; 003(8) contracts
- **relations:** palette/layout/typography migration (025); composition (027)
- **verify-later:** content_components, css_themes, style_collections schemas

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Composition three stages: direction → composition → execution
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4) and 027 describe the deployed reorder ("Applied"); install_theme removed from webdesign-agent
- **what:** (1) domain-research-classifier writes design_intent (structured palette/typography reference_values + style_direction scheme). (2) site-design-planner (deterministic, no LLM, `needs_composition` item) resolves layout (weighted scheme-aware tag match), typography (match-or-insert), palette (always new site-specific row) via signal cascades, then install_site_composition atomically writes css_themes+style_collections+sites pointer+resolved_composition spec (a decision record, not CSS; refuses overwrite). (3) webdesign-agent (needs_design, depends_on composition) renders and commits styles.css — sole writer.
- **sources:** 002(4)#Composition; 027 full; 025 (schema underneath)
- **relations:** scheme-aware matcher; renderer cascade; superseding-spec-doesn't-undo-install failure mode (028)
- **verify-later:** fork_theme_composition.go resolvers; install_site_composition; needs_composition/needs_design ordering

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Scheme-aware weighted layout matcher + needs_new_layout_candidate HITL signal
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002(4); 016 §9 v2_55 records its half-merged-deploy failure and fix shipped as one merged file
- **what:** Layout matching weights tags by rarity with category/description bonuses; the site's scheme (from design_intent.style_direction) is a near-hard constraint (light site never placed on dark layout while any non-dark fits). On fallback it queues `needs_new_layout_candidate` (status needs_human_review, skipped by dispatch) — the honest "library is missing a layout" signal. layouts.scheme nullable → incremental curation.
- **sources:** 002(4)#Composition; 027 §2; 016 §9 scheme-matcher entry
- **relations:** library growth; section-contrast open question (036)
- **verify-later:** resolveLayoutByTagsWeighted; layouts.scheme population

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Theme/layout library growth and the fork-with-review gate
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 003(8) forking rules deployed; 002(3)'s auto "search→reuse→generate→store" loop dropped from 002(4) (superseded by curated-library stance); LLM layout matching "a future step"
- **what:** Layouts are a curated shared grammar (no auto-generated bespoke layout per site). Growth: hand-added variants, or HITL route — ForkThemeFromSiteAction promotes a rendered design into css_themes+style_collections with needs_review=true and a needs_theme_review item; selectors must exclude needs_review rows; rejection only affects future sites. Lineage columns (origin/forked_from/source_site/forked_at) on themes and collections.
- **sources:** 002(4)#Library growth; 003(8)#CSS Theme Template Contract (lineage, review gate, forking rules)
- **relations:** 025 migration lineage model on palettes/layouts/typography_sets
- **verify-later:** fork_theme_from_site_action.go; needs_review filtering in selectors

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Superseded: single webdesign-agent brand+CSS role and the brand-designer/layout-architect/style-generator split
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** dropped between 002(3) and 002(4); 002(4): "The earlier 'one agent generates brand + CSS' shape is superseded by the composition/execution split"
- **what:** Earlier architecture had webdesign-agent doing brand analysis + CSS with a planned future split into brand-designer/layout-architect/style-generator, and an auto theme-library reuse loop. Replaced by site-design-planner (composition) + webdesign-agent (render); a finer split deferred until search-and-adapt clearly beats render-from-composition.
- **sources:** 002_system_architecture(3).md (family-delta); 002(4)#Design Agent Family
- **relations:** composition three stages
- **verify-later:** no brand-designer/layout-architect agent_definitions rows expected

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Palette/layout/typography composable-theme migration (025)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 025(3) is the migration plan; 027/036 confirm palettes/layouts/typography_sets live and read by renderer; legacy css_themes columns retained "until Phase 7"; 036 notes Phase 4.5 coupling still present
- **what:** Split css_themes into palettes (colours jsonb open slot map), layouts (css_template + structure_tokens + default header/footer FKs + scheme), typography_sets (fonts+scale), each with the origin/needs_review/fork lineage model; css_themes becomes a composition of three FKs. Motivation: the old library was one layout with 14 palette skins behind a silent standard-brochure fallback. Template data moves to map-based Palette/Typography/Structure (no Go change per new slot). Direct cutover, no shadow mode; selector unchanged in this phase.
- **sources:** 025(3) full; 036 §3
- **relations:** composition stages; layout scheme matcher
- **verify-later:** legacy columns still read anywhere; Phase 4.5/7 progress

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Renderer theme-resolution cascade and the emergency fallback
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 027 §4: theme_name literal cleared "as the cutover moment"; emergency fallback + logger.Error monitoring rule
- **what:** render_css_from_spec resolves theme by config.theme_id → config.theme_name → sites.style_collection_id join (production path); all-miss falls to standard-brochure WITH logger.Error — any emergency-fallback line is a pipeline bug. resolveThemeIDFromSiteContext never errors, warns with a distinguishing reason.
- **sources:** 027#Renderer Changes
- **relations:** install-before-render ordering; B-029-3 theme-vars-not-deployed bug
- **verify-later:** emergency fallback frequency in logs

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Build-time design/imagery trigger emission (Gap A)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "deployed step order in v1.0.1047 is read_specs → … → reconcile_site_plan → emit_design → emit_imagery → complete … So Gap A is closed on both fresh-build and adoption paths" (2026-05-26, verified on gamesdesign cascade)
- **what:** `emit_design_items` and `emit_imagery_items` (shared `imageryplan` package) wired as plan-time steps in build-site-planner, closing the long-standing gap where composition and imagery items were never emitted after the Phase-1 refactor moved the terminal step away from WriteBuildItemsAction. Nine needs_imagery items at documented priority bands (65 index-hero, 70 site-logo, 75/80 others, 98 clamped section-scope) observed live.
- **sources:** HANDOFF_2026-05-26_design_imagery_triggers_and_adoption_diagnosis.md#What-deployed
- **relations:** site-design-planner; imagery loop closure; site_plan_imagery
- **verify-later:** build-site-planner workflow steps in agent_definitions; imageryplan package

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site-design-planner composition resolver (composition before render)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "First successful composition run … completed cleanly" (2026-04-20 gamedesign.uk with all five IDs); "needs_composition ran via site-design-planner and install_site_composition populated sites.style_collection_id" (2026-05-26)
- **what:** A handler agent (needs_composition work item) that resolves layout (deterministic tag-overlap against layouts.industry_tags), typography (font-family/character match) and palette (fingerprint → mission → design_intent priority) BEFORE webdesign-agent renders, installing css_themes + style_collections rows transactionally and hard-failing when classification is missing. Fixes the fork+install conflation that produced first-render-with-wrong-layout (two commits, first knowingly wrong). Scope decisions: brave backfill, hard-fail loud logging, adoption and new builds unified, re-resolution deferred to HITL, fork-to-library gated behind two flags.
- **sources:** HANDOFF_2026-04-18_design_and_styling…md#3-6; HANDOFF_2026-04-20_composition_deployed…md; FOCUS_navigation_HANDOFF_navigation_fix.md#Architectural-Gap (origin: navigation/layout spec idea)
- **relations:** composable theme migration; navigation/layout specs; classification tags mismatch
- **verify-later:** site-design-planner agent definition; resolve_composition_*.go, install_site_composition.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Composable theme migration 025 (palette + layout + typography decomposition)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phases 1-3 (data model, layouts, seeding) are deployed and verified. Phases 4-5 (renderer cutover, fork action rewrite) were deployed but not end-to-end verified" (2026-04-18); renderer subsequently exercised in later cascades
- **what:** Themes decomposed into palettes, layouts (15 seeded CSS templates each passing a 7-point contract audit), typography_sets (6 seeded), FK-linked from css_themes/style_collections; renderer cutover to a single JOIN loader + FuncMap (palette/typo/token) with hard error on NULL FKs; fork action resolves the three pieces before insert.
- **sources:** HANDOFF_2026-04-18_design_and_styling…md#2
- **relations:** CSS assembly pipeline; site-design-planner
- **verify-later:** layouts/palettes/typography_sets row counts; render_css_composition_loader.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### webdesign-agent post-merge loop bug and generate_css stuck mystery
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "This is a loop bug in my migration … Fix proposal (NOT YET APPLIED)"; "Even with the loop fixed, we STILL don't know why generate_css didn't execute" (2026-04-20); a later cascade (04-23) "proceeded through generate_css and deploy_css to check_should_fork" suggesting recovery
- **what:** The 010 migration left every path out of deploy_css looping back to generate_css (update_site.next_step and check_update_db.else_step should point at check_should_fork); separately, one run sat at generate_css (deterministic action) producing no log line, no heartbeat, evidence lost to pod rotation. Instrumentation runbook written for reproduction.
- **sources:** HANDOFF_2026-04-20_composition_deployed_design_stuck.md#A
- **relations:** silent completion; consumer-group race (candidate explanation)
- **verify-later:** current webdesign-agent next_step wiring; whether the loop fix SQL was applied

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Composition classification-tags mismatch (industry_tags empty)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "composition_layout.reason: 'fallback — no classification tags' … layout resolver fell back to brochure-formal for a site that clearly wants something dashboard/application-like" (2026-04-20); migration 008 (dynamic taxonomy, industry_tags array from classifier) validated 2026-04-23 likely addresses it
- **what:** The layout resolver read a nonexistent tags array while classification stored industry/sub_industry strings, so every site fell back to the generic layout and style_collections.industry_tags was written empty, breaking future library matching. Migration 008 made the classifier emit an industry_tags array against a dynamic taxonomy read from the layouts table (read_layout_taxonomy action), validated end-to-end with tool-portal-dark selected via library_match.
- **sources:** HANDOFF_2026-04-20_composition_deployed…md#B; HANDOFF_2026-04-23(1).md#deployed, #validated
- **relations:** site-design-planner; dynamic taxonomy classifier
- **verify-later:** readClassificationFromContext in resolve_composition_helpers.go; classifier output shape post-008

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Planner palette prose vs structured reference_values (Gap C)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "OPEN" (2026-05-26): planner emits colour decision as design_intent.colour_mood prose; composition palette cascade misses the design_intent slot and falls to layout-seed default
- **what:** Planned colours reach the render only via the webdesign-agent overlay, not the base composition. Fix options: planner emits a structured palette.reference_values block (primary/secondary/accent/background/surface/text/text_muted/border) or site-design-planner consumes colour_mood directly.
- **sources:** HANDOFF_2026-05-26…md#gaps, #Where-to-resume
- **relations:** palette cascade; site-design-planner
- **verify-later:** plan_site output schema; palette resolver slots

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Library-row cleanup pattern for failed cascades
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** executed 2026-04-23 with counts (4 css_themes, 7 palettes, 4 style_collections cleared for gamesdesign) and a NOT IN guard protecting seeded layouts
- **what:** Bad cascades leave one set of library rows (css_themes/palettes/style_collections/typography_sets) per resolve attempt; if left, the matcher can pick wrong-decision artefacts for future sites. Reverse-FK-order delete by source_domain is the recovery pattern. Related open item: site deletion should clean up unreferenced library rows (FKs are SET NULL, leaving orphans).
- **sources:** HANDOFF_2026-04-23(1).md#cleanup, item 18
- **relations:** site-design-planner re-resolution ambiguity; duplicate sites-row question (item 20)
- **verify-later:** any delete-site action's library handling

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme derivation and drop at render
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** PLAN §Confirmed-at-code-level: "`deriveSchemeFromDesignIntent(style_direction, suggested_style)` returns light/dark/''; … `buildResolvedCompositionSpec` records the layout/palette ids … but not the scheme value"; notes (Sb) traced it end-to-end 2026-06-30.
- **what:** Scheme (light/dark) is derived at composition from `design_intent.style_direction` by substring matching, used by `resolveLayoutByTags` as a near-hard constraint to pick the layout, then dropped: neither the CSS loader SELECT nor the component `RenderContext` reads `layouts.scheme` (check-constrained to light/dark/neutral). It survives only as the layout's curated property, recoverable via `sites.style_collection_id → style_collections.css_theme_id → css_themes.layout_id → layouts.scheme`. Light/dark variety is handled by paired layouts (tool-portal-light/-dark), not runtime component flipping. Corollary data fact: only 3 of 18 active layouts have `scheme` set — scheme metadata is sparsely curated.
- **sources:** PLAN_scheme_to_components(1).md#Confirmed-at-code-level; running_notes_scheme_to_components(55).md#Sb #Sf; RUNBOOK_scheme_to_components(50).md#CHECK-3-RESULTS (3e)
- **relations:** three-part styles.css assembly; paired-variable standard; explicit RenderContext.Scheme (abandoned).
- **verify-later:** `platform/orchestration/actions/` deriveSchemeFromDesignIntent, resolveLayoutByTags, buildResolvedCompositionSpec; `layouts.scheme` column + values.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Layout CTA pair curation with WCAG contrast gates
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Notes (Sq) "W1 complete + verified"; (Su) "W1b COMPLETE: five layouts curated"; w1b comments record the five hex swaps and expected values.
- **what:** W1 added the missing CTA pair to tool-portal-light via anchored `regexp_replace` on the verified `--color-footer-text` line (`{{palette "cta_bg" "#e9e2d3"}}` + `cta_text "#1a1a1a"`, contrast ≈13.5, mirroring tool-portal-dark's neutral elevated band; accent alternative offered). A sweep computed every layout's cta pair contrast; five seed layouts failed 4.5 with white text and got same-hue darker fallbacks (W1b, zero live impact — no site uses them). Pair values are deliberate per-layout design: several light layouts curate DARK footer bands — "light site, dark band by choice" is already a curated model in the library. Requirement carried into the contract: pair contrast ≥ 4.5.
- **sources:** w1_01_add_cta_pair.sql; w1b_01_contrast_batch.sql; RUNBOOK_scheme_to_components(50).md#W1-RESULTS #CHECK-4-RESULTS (4b); SPEC_scheme_to_components.md#W1
- **relations:** paired-variable standard; three-part styles.css assembly.
- **verify-later:** layouts css_template cta pair values for the six touched layouts.

<!-- SOURCE: U04_idea_uk.md -->
### Two-stage base+override design pipeline (site-design-planner + webdesign-agent)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Investigation 2026-06-20 "confirmed against agent_definitions… and the deployed render code"; routing table shows needs_composition→site-design-planner, needs_design→webdesign-agent with an explicit depends_on.
- **what:** Design is deliberately split (027 §2 — ordering was the reason): Stage 1 `site-design-planner` (deterministic, no LLM) resolves layout/typography/palette and installs them (css_themes 3-FK row + style_collections + sites.style_collection_id + a resolved_composition spec) — renders nothing; Stage 2 `webdesign-agent` produces an LLM design overlay, renders the layout template over the installed composition base per a fixed merge-authority rule (LLM wins core palette slots + typography; composition wins layout/structure tokens/specialised slots), and is the sole styles.css deployer (git_commit → Actions → B2). `emit_design_items` queues both from one step with needs_design gated on composition. The 2026-06-20 correction: this is NOT a shared-responsibility bug — the overlay was designed as *optional and partial* (025 §5).
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md; idea.uk/HANDOFF(13).md (three-layer model); idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md
- **relations:** mandatory-full overlay bug; resolved_composition pointer; palette cascade; 002 architecture doc (rewritten 2026-06-22 to match).
- **verify-later:** render_css_composition_helpers.go buildPaletteMap/buildTypographyMap; emit_design_items_action.go.

<!-- SOURCE: U04_idea_uk.md -->
### resolved_composition pointer spec + install_site_composition semantics
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Verified live on idea.uk twice (dark install, then re-resolve); "resolved_composition is a *pointer* — it carries palette_id/name/source, not the colour values".
- **what:** The composition install contract: css_themes row created with all three FKs but empty css_content (webdesign-agent fills at render); style_collections points at the theme; `sites.style_collection_id` set only if NULL — install **errors rather than overwrites** an existing composition ("re-resolve not supported; clear it manually"), which is why re-resolving requires an explicit detach; the old resolved_composition spec is superseded and a new one inserted as the lineage/decision record (`lineage.{palette_source, typography_source, layout_source}`). Renderer resolution is strict: missing/NULL composition parts hard-error ("migration gaps are audit events, not silent fallbacks"), with a loud emergency fallback to standard-brochure.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (install + loader sections); idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A: "a bare needs_composition requeue no-ops")
- **relations:** composition re-resolve procedure; two-stage pipeline.
- **verify-later:** install_site_composition_action.go; render_css_composition_loader.go.

<!-- SOURCE: U04_idea_uk.md -->
### Palette/typography resolution cascade + the dead-slot bug and fingerprint fallback hardening
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Cascade live and proven; dead slot "CONFIRMED why (2026-06-19, from resolve_composition_palette.go)"; hardening "DELIVERED" as code 2026-06-19/20 but "READY… NOT YET APPLIED" in the backlog (needs image rebuild + roll).
- **what:** Palette source cascade: design_reference → mission → design_intent.palette.reference_values → layout seed → archetype default (typography analogous; palettes always site-specific, layouts a shared curated library). The bug: cascade slot 1 reads `design_reference.palette.reference_values`, a key the adoption fingerprint never writes (it stores suggested_mapping/css_variables/colors) — so slot 1 was dead and adopted references never drove the palette; the delivered hardening points slot 1 at the fingerprint's real keys as a fallback after design_intent. Under the current LLM-wins merge the composition palette mostly doesn't paint anyway (it fixes lineage + rare-gap fallback) — the painting lever is the classifier fix feeding the LLM.
- **sources:** idea.uk/UPDATE_FOCUS_design_adoption_workplan_2026-06-19(1).md#3; idea.uk/HANDOFF(13).md (cascade + backlog item 3); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (problem 1)
- **relations:** structured design_intent migration; two-stage pipeline; adoption generate_design_intent.
- **verify-later:** resolve_composition_reference_helpers.go deployed or not; extractPaletteSignal/extractTypographySignal.

<!-- SOURCE: U04_idea_uk.md -->
### The mandatory-full overlay bug + improver-not-rewriter fix (and the superseded rewrite options)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** "direction settled 2026-06-20, not started" (runbook open items); superseded options (a)/(b)/(c) "kept for history" in the investigation doc.
- **what:** The merge rule assumed the LLM overlay would be *optional and partial* (asserting only genuine brand identity), but the analyze_design prompt mandates a full 8-slot color_scheme with "be distinctive" framing — so the LLM repaints every fresh build, the 028-forbidden silent override. Fix v1 (no contract change): show the LLM the established palette, require it to keep it as the foundation and change slots only with a reason, diff the result against the composition base and write an audit record (slot, old→new, reason) per build; v2 (deferred, evidence-driven): cap core-slot changes per refine + optional denylist. Explicitly supersedes the earlier single-owner options: (a) LLM-owns-core / slim the planner, (b) flip the merge so structured composition wins, (c) collapse to one design agent — rejected because the base+partial-overlay split is intentional.
- **sources:** idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md (CORRECTION + fix sections; superseded options); idea.uk/DESIGN_PIPELINE_two_track_investigation(1).md (the pre-correction "decision options" — family-delta); idea.uk/HANDOFF(13).md (backlog item 4)
- **relations:** structured design_intent (precondition); design docs 025/027/028.
- **verify-later:** analyze_design prompt in webdesign-agent def; any design-audit table.

<!-- SOURCE: U04_idea_uk.md -->
### Scheme-aware weighted layout matcher + layouts.scheme + tool-portal-light
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Matcher code LIVE — merged … built into the chassis image and site-design-planner rolled. (Confirmed live 2026-06-25.)"; migration applied; re-resolve proved tool-portal-light selection end-to-end 2026-06-25.
- **what:** Replaces the tags-only, scheme-blind `resolveLayoutByTags` (exact-overlap count, alphabetical ties) that put light-editorial idea.uk on tool-portal-dark. New matcher (all in Go, ~17-row library fetched and scored transparently): scheme as a **near-hard constraint** (a light site won't land dark while any non-dark fits; mismatch queues the existing needs_new_layout_candidate HITL item), IDF-weighted tag rarity (specific beats generic), synonym normalisation to a controlled vocabulary, category + description keyword bonuses. Paired migration adds nullable `layouts.scheme` (light/dark/neutral; NULL degrades gracefully) and a new `tool-portal-light` layout — same structural class contract as its dark twin, light fallbacks, reads palette vars. Decision history: NO auto-layout-generation — a curated, varied library + scheme-aware matching is the lever; LLM-judge/pgvector deferred.
- **sources:** idea.uk/resolveLayoutByTags_weighted.go.patch.txt (header); idea.uk/migration_layouts_scheme_and_light_tool_portal.sql; idea.uk/HANDOFF(13).md (matcher rewrite)
- **relations:** deriveSchemeFromDesignIntent; composition re-resolve; scheme-to-components gap (the next layer down).
- **verify-later:** fork_theme_composition.go current resolveLayoutByTags; remaining NULL layouts.scheme rows (backlog).

<!-- SOURCE: U04_idea_uk.md -->
### Composition re-resolve procedure (gated, file-based, backup-first)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Steps 1–6 all marked DONE with results (2026-06-22→25); "RE-RESOLVE SUCCEEDED: idea.uk now on tool-portal-light (scheme fix proven end-to-end)".
- **what:** The safe pattern for re-running composition on an already-built site (install refuses overwrites): ordered SQL FILES — backup+inspect (with four uniqueness checks that must all be 0), gated detach+clear (NULL style_collection_id; delete the site's own collection→theme→palette→typography chain only where source_site_id matches; supersede the old spec), state-check, kcat re-trigger of site-design-planner (`domain` required by ensure_site_record), verify. Two learned caveats now doctrine: run SQL as files never pasted (paste mangled \set/blank lines and left an open transaction); a standalone-orchestrated planner ends at install and emits NO needs_design — the styles.css render is a separate explicit webdesign-agent orchestration. Distinct from the adoption teardown (bulk delete by source_domain), which must NOT be used on a fresh site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (re-resolve section); idea.uk/reresolve_idea_uk_01_backup_and_inspect.sql (+02/02b/03/04/05 series); idea.uk/running_notes(63).md (xxx–jjj checkpoints)
- **relations:** install semantics; launch idioms; scheme-aware matcher validation.
- **verify-later:** bak_*_idea_20260625 tables; orchestration_states rows for the re-resolve correlations.

<!-- SOURCE: U04_idea_uk.md -->
### The scheme-does-not-reach-components gap (P0 framework fix)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Investigation complete + report written 2026-06-26 ("Step 7 DONE — conclusion: do NOT rebuild yet; structural gap"); the fix itself deferred to a dedicated thread; composition+stylesheet layers verified correct, page layer still dark.
- **what:** The central framework finding of the thread: a site's scheme (light/dark) is decided at composition and reaches styles.css `:root`, but never reaches the components that render sections/header/footer. Components are drawn from a **dark-oriented library** by a one-active-component-per-function lookup (nothing light exists for hero/CTA/footer), self-style via inline CSS with their own class vocabulary (the layout's section rules don't apply — class-name mismatch), and hardcode dark treatments or set dark `--section-*` themselves — so a light-resolved site renders dark chrome over light content (only var-reading sections went parchment). Supporting facts: `is_dark_section` is loaded but never used in selection, unreliable, and conflates "intrinsically dark" with "should contrast the page"; no layout declares default header/footer and the planner never runs update_site_defaults; the hero navy-button bug (--accent-color vs --color-accent) was already fixed in the library — deployed pages were stale. Modelled as: scheme was treated as colour-only; the structural half was never plumbed.
- **sources:** idea.uk/REPORT_scheme_does_not_reach_components.md; idea.uk/HANDOFF_scheme_to_components(1).md; idea.uk/running_notes_2(6).md (lll–ooo); idea.uk/one_sentence_description.md
- **relations:** scheme-as-override thesis (the fix shape); section-contrast model; header/footer wiring; component class-contract question; scheme-aware matcher (upstream, done).
- **verify-later:** whether the dedicated thread landed (component templates de-hardcoded; light footer exists; update_site_defaults in build path).

<!-- SOURCE: U04_idea_uk.md -->
### Scheme-as-override thesis + section-contrast model (base scheme + per-section contrast intent)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** REPORT §4–5: "the likely shape (to be validated by the investigations, not assumed)"; eight design questions "must be answered before any code".
- **what:** The design steer for the P0 fix: scheme is a **set of variable values** — an override layer supplied by composition/renderer and consumed by de-hardcoded components — never a duplication of components into *-light/*-dark variants (new functions only for genuine structural divergence). The model separates **site scheme (base)** from **per-section contrast intent** (a dark hero on a light site is legitimate, intentional contrast — "make everything light" is wrong), both applied as data at render time through the existing `--section-*` mechanism, making the renderer the single adaptation point. Eight design questions scoped (where scheme lives at render; who owns section darkness; the override mechanism; the gating class-vocabulary question Q4; is_dark_section's fate; header/footer; migration without breaking dark sites; an auditor guard), with nine investigations (A–I) and a provisional fix shape stated as hypothesis. Definition of done includes a scheme-coherence audit so it can't silently regress.
- **sources:** idea.uk/REPORT_scheme_does_not_reach_components.md#4-8; idea.uk/HANDOFF_scheme_to_components(1).md; idea.uk/TODO_chassis_and_idea_uk(1).md#P0
- **relations:** CSS colour-inheritance model (the vehicle); improvement-loop (the guard); scheme gap (the problem).
- **verify-later:** stage-2 code check of component templates for --section-* consumption.

<!-- SOURCE: U04_idea_uk.md -->
### Header/footer chrome wiring chain (and its live gaps)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** Data findings 2026-06-26: all inspected layouts have NULL default header/footer ids; idea.uk's new style_collection inherited NULLs; site_components still point at the original dark, now-inactive site-header/site-footer.
- **what:** Site-level chrome flows down a chain: `layouts.default_{header,footer}_component_id` → install_site_composition copies onto style_collections → `update_site_defaults` copies onto site_components → renderAndStoreSiteComponent renders into site_components.rendered_html, with a hardcoded RenderFallbackHeader when unlinked. Live gaps: no layout declares defaults, site-design-planner never runs update_site_defaults, and header/footer are therefore never scheme-derived — a re-resolve leaves the old chrome in place. The library has light headers but NO light footer. Fix direction: layouts declare scheme-appropriate defaults + the build runs update_site_defaults, one adaptive header/footer per the override thesis.
- **sources:** idea.uk/running_notes_2(6).md (mmm data findings); idea.uk/REPORT_scheme_does_not_reach_components.md (facts + Q6/investigation F); idea.uk/001_component_flow.md
- **relations:** scheme gap; contracts-and-standards Site Component Linkage Contract (003).
- **verify-later:** update_site_defaults_action.go and its call sites; how the original build chose idea.uk's header.

<!-- SOURCE: U04_idea_uk.md -->
### Section→component resolution: direct-function Path 1 vs scoring selector Path 2
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** plan_sections code read 2026-06-26: "Path 1 = components[sectionName] direct lookup… All current sites hit this path"; component_selector "SELECTs is_dark_section into the struct but NEVER uses it in scoring".
- **what:** How a planned section becomes a component: Path 1 matches the section name directly against `content_components.function` (one active component per function — uniqueness index), which all current sites hit; Path 2, the scoring `component_selector` (suitable_site_types 0.35 + page_types 0.15 + quality 0.3 + specificity + usage), only runs for section_type names that aren't functions — and is scheme-blind. Consequences: there is no place to pick a scheme-appropriate variant for current sites (making a scheme-aware selector necessary-but-insufficient), and layout-aware section selection is explicitly documented future work (027 §10). page-rerender re-assembles stored HTML without re-selecting; only page-build-handler re-runs plan_sections.
- **sources:** idea.uk/running_notes_2(6).md (mmm/nnn corrections); idea.uk/REPORT_scheme_does_not_reach_components.md#2; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Step 7 findings)
- **relations:** scheme gap; tool-library (component registry/matching); flag_page_image_rebuild as the rebuild trigger.
- **verify-later:** plan_sections_action.go Path-1 comment; component_selector.go scoring.

<!-- SOURCE: U05_content_quality_linking.md -->
### Design fingerprint pipeline (design_reference / design_intent / three-way priority)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** WORK_PLAN_v2 status tables: Phases 1–2 largely "Applied to DB"/code-ready, Phase 3 (site-design-planner) and Phase 4 (requirement-driven components) "Not started".
- **what:** The adoption design-fidelity subsystem: extract_design_fingerprint parses crawled rawHTML for concrete tokens (CSS vars, hex, fonts) into a design_reference spec; an LLM derives the semantic design_intent from it; the webdesign-agent applies a three-way priority (design_intent → design_reference reproduce-faithfully → generate-from-industry); audit loop proposes rather than applies design changes. The companion problems doc catalogues the failure it answers (LLM guessed colours/fonts from markdown summaries, layout flattened to generic components, header/footer genericised). Phase 3 (navigation/layout specs + site-design-planner agent) and Phase 4 (section recipes, requirement-driven component selection, "every build conceptually an adoption") remain unbuilt. Carried in this unit as packaged context (stale-risk flagged; the composition refactor postdates them).
- **sources:** package_module/output_contexts/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md; FOCUS_design_and_styling_adoption_problems.md; HANDOFF_2026-06-09(2).md#design-fidelity-background
- **relations:** adoption pipeline workflow (extract_fingerprint/generate_design_intent steps); design-composition unit docs 025–027.
- **verify-later:** extract_design_fingerprint_action.go deployed?; site_specs design_reference/design_intent aspects; site-design-planner existence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Deterministic styles.css rendering (webdesign-agent: LLM spec → Go template → git commit)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "render_css_from_spec — 'Render CSS from design spec using Go template (deterministic, no LLM)'" (NOTES §9bi, from webdesign-agent config); full chain observed live §9bn.
- **what:** The webdesign flow is analyze_design (LLM → design-spec JSON: color_scheme/typography/spacing) → render_css_from_spec (deterministic Go template over DB layout templates — `comp.LayoutTemplate` — merged with palette/typography; forkable themes) → git_commit styles.css → site-asset-renderer. The defined CSS vocabulary therefore lives in one Go-owned render path (the single home for generic fixes); storage_actions.go's styles.css writes belong to the OLD builder extract paths and must not be patched for this flow. Caution: re-running analyze_design mints a fresh LLM spec — palettes can shift unless pinned — hence the manual bridge-commit option for palette-preserving fixes.
- **sources:** NOTES(43).md §9bi, §9bj, §9bm; RUNBOOK(49).md Part D
- **relations:** D2a (lives inside it); R6f; layout curation.
- **verify-later:** render_css_from_spec_action.go; webdesign-agent default_config; needs_design production (build-site-planner).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Scheme resolution pipeline and where the signal stops
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Scheme→variable pipeline verified correct; all 18 layouts carry the four chrome vars" (RUNBOOK_scheme(18) CURRENT POSITION, 2026-07-02); RenderContext gap traced in running_notes Sb.
- **what:** A site's light/dark scheme derives from design intent (`deriveSchemeFromDesignIntent`), constrains layout matching (`resolveLayoutByTags`; `layouts.scheme` light/dark/neutral), and reaches styles.css via palette :root + luminance defaults — but is never recorded in the composition spec and never reaches the component render context (`RenderContext` has palette colours, no scheme field). The corrected understanding: the scheme reaches components IMPLICITLY via variables; components defeat it by hardcoding dark assumptions — so the core fix is de-hardcoding, not new plumbing.
- **sources:** running_notes_scheme_to_components(22).md Sb, Sc, Sf, Sk; RUNBOOK_scheme_to_components(18).md header + CHECK 1
- **relations:** paired-variable direction; buildSectionDefaults; R6f (later vocabulary-level echo).
- **verify-later:** deriveSchemeFromDesignIntent; RenderContext struct; layouts.scheme population (only 3 of 18 curated).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Paired-variable design direction (Alt C: curated bg+text pairs; completion of the existing standard)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "pair convention is ALREADY the standard — 18/18 --color-primary-text, 17/18 --color-cta-text" (Check 4c); W1–W3e template work executed 2026-07-02/03 but "inert until re-render/rebuild"; Go batch + tail unshipped.
- **what:** The user requirement "a light scheme must be able to render fully light, and may carry dark hero bands" selects layout-curated background+text variable pairs (chrome pattern generalised: --color-cta-bg/--color-cta-text etc.), palette-overridable per site (specialised slots theme-wins), per-instance later via plan directives — components consume pairs and never declare `--section-*`; renderer luminance defaults remain the base; dark-band-by-choice stays curated per layout. Judged a COMPLETION of existing architecture, not a restructure (one layout to patch, components to bring in line). Execution: ten templates fixed + seven verified clean (footer, CTA via inverse-pair buttons, hero, five hero-* variants, about-content, brief-explanation), idea.uk chrome repointed; full rebuild + Go batch (scheme-aware fallbacks, creator prompt, fixer re-aim) pending at capture.
- **sources:** running_notes(22).md Sn, So; RUNBOOK_scheme_to_components(18).md CHECK 4 RESULTS + WHERE WE ARE; SPEC referenced therein
- **relations:** hazard/band split; hero ink model; is_dark_section demotion; Phase 4.5 (deferred); chrome linkage repair.
- **verify-later:** SPEC_scheme_to_components.md (outside unit); layouts cta pair coverage; whether W6 rebuild shipped.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Hazard-class vs band-class self-declarer split; is_dark_section demoted to metadata
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "the 37 self-declarers split into two classes" with named components (Check 3c, run 2026-07-02); "6 declarers have is_dark_section=f… never key styling on the LLM-authored flag".
- **what:** Library-wide diagnosis: of 84 active section components, 37 self-declare `--section-*` — ~18 hazard-class (declare dark context while painting surface vars or nothing → white-on-light bugs today) vs ~19 band-class (paint palette bands + white text — coherent but block "fully light"); 15 carry hex backgrounds. `is_dark_section` is an LLM-authored component bool contradicted by 6 of its own declarers and consumed by nothing that styles — demoted to selection/imagery metadata; styling must never key on it. This classification sized every subsequent fix batch.
- **sources:** RUNBOOK_scheme_to_components(18).md CHECK 2/3 RESULTS; running_notes(22).md Sn, Sh (E findings)
- **relations:** paired-variable direction; improvement-loop fixers (key on the flag — part of why they're wrong); component-creator prompt drift.
- **verify-later:** content_components is_dark_section values vs template styling; remaining unconverted declarers (~10 hazard + ~17 band).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Chrome linkage tangle: four overlapping header/footer default stores and the dark fallback
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "header_component_id is effectively a DEAD column — nothing populates it" (running_notes Sl); repoint executed (Ta); scheme-aware fallback Go batch still pending in WHERE WE ARE.
- **what:** Four coexisting default stores for site chrome: style_collections.header/footer_component_id (the store RenderHeader reads first — installed NULL and never written), site_components slots (render cache, can pin inactive components indefinitely), sites.default_components jsonb (UpdateSiteDefaultsAction target, unread on the render path), layouts.default_*_component_id (all NULL). RenderHeader's chain is collection-id → GetComponentByFunction("site-header") → RenderFallbackHeader, and the fallback hardcodes dark (PrimaryColor bg + white text) — so any linkage break yields dark chrome regardless of scheme. Fix shape: de-hardcoded active chrome components (header already model), repoint stale pins, scheme-aware fallbacks consuming the chrome var pairs (all 18 layouts already define them).
- **sources:** running_notes(22).md Sg, Sh, Sl; RUNBOOK_scheme_to_components(18).md CHECK 3b, HEAD-SLOT RESOLUTION, W4b
- **relations:** chrome refresh gating; rerender fossilisation (stale pinned renders reached deploys); paired-variable direction.
- **verify-later:** style_collections.*_component_id population; RenderFallbackHeader/Footer/Head current CSS; whether the Go batch shipped.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Layout curation: CTA pair completion, WCAG contrast batch, updated_at trigger
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "W1 COMPLETE… W1b: five × UPDATE 1… trigger observed working in anger" (RUNBOOK_scheme(18) W1/W1b/W2b RESULTS, 2026-07-02).
- **what:** Seed-layout curation as part of the theming fix: tool-portal-light gained the missing --color-cta-bg/--color-cta-text pair (#e9e2d3/#1a1a1a, ≈13.5 contrast); a five-layout batch nudged failing cta_bg fallbacks to same-hue passes (all ≥4.5); layouts.updated_at gained a BEFORE UPDATE trigger via the shared set_updated_at function (reuse-gate fired as designed when CREATE FUNCTION collided). Several light layouts deliberately curate dark footer bands — "light site, dark band by choice" is an existing curated model, not a bug.
- **sources:** RUNBOOK_scheme_to_components(18).md W1/W1b/W2b RESULTS, CHECK 4b; running_notes(22).md Sq–Ss
- **relations:** paired-variable direction; deterministic styles.css rendering.
- **verify-later:** layouts cta pair values; trg_layouts_updated_at.

<!-- SOURCE: U09_adoption.md -->
### Design/composition flow gaps A–B–C and the plan-time triggers
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "UPDATE 2026-05-26: gap (A) is being deployed to production now — emit_design_items + emit_imagery_items are wired into build-site-planner (agent-chassis v1.0.1047)… Gaps (B) and (C) remain open."
- **what:** Three stacked gaps behind themeless/off-palette built sites: (A) composition never triggered — the Phase-1 refactor lost the needs_composition/needs_design emission; restored as plan-time steps `emit_design` (guarded on style_collection_id IS NULL) and `emit_imagery` (priority-banded so imagery lands in the first deploy); (B) planner design drift — the adopted `design`/`design_reference` aspect is never rendered into the plan_site prompt, and `design_intent.style_direction` is a fixed 3-value enum (professional-dark|modern-light|bold-creative) that cannot express e.g. "cyberpunk terminal", forcing collapse; (C) colour reaches the resolver only as prose `colour_mood` flattened into directives — the palette cascade's design_intent slot needs structured `palette.reference_values` (core keys primary…border). Whether colours are preserved vs re-chosen is governed by fidelity+locks, not the shape. REUSE NOTE outstanding: extract shared emitInitialCompositionAndDesign so emit_design's insert block can't drift from WriteBuildItemsAction.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md, README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** theme-layer resolution; imagery loop closure; fidelity dial; spec-retire investigation (dead `structure` aspect, single-reader `design` aspect)
- **verify-later:** build-site-planner v1.0.1047 workflow (reconcile_site_plan → emit_design → emit_imagery → complete); style_direction enum; createPalette core keys

<!-- SOURCE: U09_adoption.md -->
### Spec ownership / silent-override failure-mode principle
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Restated as settled doctrine from doc 028 in FOCUS_design_composition §5 (2026-05-26).
- **what:** An agent that changes behaviour on information it did not put in the spec is a bug; an agent that overwrites a spec aspect another agent owns is a category error; an agent that produces the right output but doesn't write it to the spec is not helpful. Applied concretely: `design_reference` is owned by site-adoption-agent and records the source site's design — writing our chosen palette into it would misrepresent the source (corrected an earlier proposal). Producer/consumer schema drift (colour_mood prose vs reference_values; features {title,description} vs {icon,name,description}) is the same failure shape.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#5, #2
- **relations:** design gaps B/C; site_specs aspect ownership model
- **verify-later:** doc 028 statement; aspect writer inventory

<!-- SOURCE: U10_imagery.md -->
### No runtime re-compose path — layout change via the 025 FK-swap pattern
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "B7 COMPLETED 2026-07-10 evening — via the 025 FK-swap pattern… there is no runtime re-compose path (deliberate deferral). NEW OPEN ITEM: build a proper runtime re-compose mode."
- **what:** Changing an existing site's layout is deliberately unsupported at runtime: install_site_composition refuses when a style_collection exists, and fork_theme_from_site's install mode was removed 2026-04-19. The sanctioned workaround is a targeted `css_themes.layout_id` FK swap (backup + verify) followed by a webdesign-agent CSS re-render + page rerenders. Root cause of the B7 brochure fallback: robot-hands' old-format classification lacked `industry_tags`, so the scheme-aware matcher had nothing to score — while the layout library already held the right answer (tool-portal-dark, itself grown from a prior instance of the same gap: the library learns).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-18–22, SQL_2026-07-10_b7_layout_fix.sql, SQL_2026-07-10_b7_layout_swap.sql, PLAN_imagery_best_in_class.md#B7
- **relations:** design-composition doc 027 matcher; needs_new_layout_candidate → library-growth loop; classification format drift (also caused the missing news flag).
- **verify-later:** install_site_composition refusal; robot-hands css_themes.layout_id = tool-portal-dark.

<!-- SOURCE: U12_docs024_archives.md -->
### design_reference / design_intent spec-aspect split
*(merged from 2 independent findings)*
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** `FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md`: "Replaces the old vague `design` spec"; `027_design_and_site_planner_v2.md`'s "Related Specs" table (an independent, later live doc) confirms both `design_reference` and `design_intent` as live, read spec aspects with defined priority cascades.
- **what:** `design_reference` holds concrete values (hex colours, font families, CSS variables, spacing) extracted mechanically (no LLM) from an adopted site's crawled HTML/CSS — a historical, immutable record. `design_intent` holds semantic creative direction (e.g. "dark IDE aesthetic... start here"), deliberately non-prescriptive so the improvement loop and webdesign-agent retain creative room; it may be auto-generated at adoption time or written later by a strategist/human. Together they replace a single, vague, LLM-guessed `design` spec aspect that conflated historical fact with creative direction (see the separate "Unified design spec aspect for adopted sites" concept below for that earlier, superseded state). Guiding principle: "design reference is history, design intent is direction" / "every build is conceptually an adoption."
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made"; old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"Principles Restated"; old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"Related Specs"
- **relations:** Unified design spec aspect for adopted sites (the superseded precursor, below); webdesign-agent three-way design priority; palette-locked-until-design_intent policy; design agent write-back resolution
- **verify-later:** confirm `design` spec aspect is no longer written anywhere; check `site_specs` for population rate of `design_reference`/`design_intent` across adopted sites.

---
*(remaining concepts below are as extracted by the 7 sub-slices, grouped by their originating file cluster)*

<!-- SOURCE: U12_docs024_archives.md -->
### Design agent responsibility split — site-design-planner (composition) vs webdesign-agent (execution)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Live 002_system_architecture(4).md line 596: "The earlier 'one agent generates brand + CSS' shape is **superseded** by the composition/execution split above."
- **what:** All archived versions of the system-architecture doc (baseline through the april26 near-final draft) model design as owned by a single agent, `webdesign-agent`, which does brand analysis, colour/typography/spacing decisions, AND CSS generation — with `brand-designer`, `layout-architect`, and `style-generator` listed as "Future split"/"Planned" agents that never materialized under those names. The live doc replaces this with a Composition/Execution/Maintenance split: `site-design-planner` (new agent) deterministically resolves layout (weighted, scheme-aware match against a shared `layouts` library), typography (match-or-new against `typography_sets`), and a site-specific palette via signal cascades, then installs `css_themes` + `style_collections` + a `resolved_composition` decision-record; `webdesign-agent` is narrowed to rendering/committing `/assets/css/styles.css` from that installed composition — "the only writer of styles.css." The `Design | webdesign-agent | Colour palette, typography, spacing, CSS` row in the Responsibility Boundaries table is likewise split into separate "Composition" (site-design-planner) and "Render" (webdesign-agent) rows, with a `needs_new_layout_candidate` HITL escalation replacing the old simple "search → maybe reuse → maybe generate" theme-growth description.
- **sources:** old/older1/002c_system_architecture_v3.md#"Design Agent Family", #"Theme library growth"; old/older1/002d_quality_assurance_architecture.md#"Classifier → Planner → Design Agent → Audit Agent"; docs024_key_docs_latest/002_system_architecture(4).md#"Composition: how a site's design is resolved and installed", #"Classifier → Planner → Design Agent → Audit Agent"
- **relations:** superseded planned agents brand-designer / layout-architect / style-generator (never built under those names); fork_theme_composition.go resolvers; QA architecture's "Responsibility Boundaries" chain
- **verify-later:** confirm `site-design-planner` agent_definitions row and `resolve_composition_layout/typography/palette` actions exist and are active.

<!-- SOURCE: U12_docs024_archives.md -->
### Early "visual identity poles" layout taxonomy (dropped)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Diff-confirmed only in the earliest of four palette/layout/typography-migration drafts; final 15-layout table uses different (hyphenated) names though keeps several "pole" nicknames.
- **what:** The very first migration draft described layout diversity as nine named "poles" tied to specific sites (Brochure/corporate, Magazine/editorial, Portfolio/kinetic "vonc", Commerce/grid, Utility/tool "thunder compute", Media/streaming "youtube", Documentation/reference, High-energy/bold "boxing", Soft/editorial). Dropped in favour of vaguer prose, then crystallised differently as the final 15-layout table (adding six layouts absent from the original nine-pole list).
- **sources:** old/older1/025_palette_layout_typography_migration.md#"2. Scope Decisions"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md#"7. The Layouts to Build"
- **relations:** composable theme migration; site-design-planner
- **verify-later:** final `layouts` table row count/names vs. the 15-layout plan.

<!-- SOURCE: U12_docs024_archives.md -->
### Phased belt-and-braces removal plan for webdesign-agent install_theme (abandoned same-day)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** v1 doc's changelog (2026-04-19) states the belt-and-braces step "remains... pending the two-phase removal plan"; live v2 doc's changelog the same day: "Merge applied" (direct removal).
- **what:** `026_design_and_site_planner_v1.md` proposed a cautious two-phase removal of webdesign-agent's defensive `install_theme`/`check_should_install` steps (diagnostic no-op first, delete only after a week of zero firings). Abandoned within hours: live v2 shows the two steps deleted outright the same day, routing rewired directly to `generate_css`, relying instead on the renderer's emergency-fallback logging as the sole safety net.
- **sources:** old/older1/026_design_and_site_planner_v1.md#"6. Removing install_theme From Webdesign-Agent (Planned)"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"6... (Applied)", #"12. Change Log"
- **relations:** site-design-planner composition-install path; renderer emergency-fallback logging
- **verify-later:** confirm `install_theme`/`check_should_install` are absent from webdesign-agent's agent_definitions.

<!-- SOURCE: U12_docs024_archives.md -->
### Visual identity library and effects library (composable design assets)
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)" in the 2026-04-11 plan; not confirmed built.
- **what:** Longer-term plan for two accumulating libraries: a visual identity library of palettes/typography/effects searchable by purpose/audience, and an effects library treating elevation/corner radius/animation/density as composable modifiers independent of layout. Likely precursor idea to the `palettes`/`typography_sets`/`layouts` table split actually implemented.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4: Requirement-Driven Components (longer term)"
- **relations:** composable theme migration; component selector by functional requirement
- **verify-later:** whether structure_tokens/effects concepts in the live composable-theme schema fulfil this idea.

<!-- SOURCE: U12_docs024_archives.md -->
### Three-layer design system (content_components / css_themes / style_collections)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Live: "Today css_themes.css_template mixes three distinct concerns into one row: palette, typography, and layout" — the migration splits this monolith.
- **what:** Early architecture with three independently-varying layers: HTML components, a monolithic CSS theme (one row = whole stylesheet), and a `style_collections` bridge table. Superseded internally when `css_themes` was split into three composable entities; `style_collections` survives as the outer bundle.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"1. The Design System: Three Independent Layers"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md#"Splitting css_themes"
- **relations:** composable theme system; style_collections bundle
- **verify-later:** confirm `css_themes` legacy columns actually dropped (Phase 7).

<!-- SOURCE: U12_docs024_archives.md -->
### Design fingerprint extraction pipeline
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Not started" → "✅ Deployed, works" (2026-04-14) → "Victory: Design Fingerprint Now Correct" (2026-04-16).
- **what:** Pipeline step parsing a crawled site's CSS into a colour/font/layout "fingerprint" so adoption rebuilds match the original. Went from unstarted idea to working end-to-end (gamedesign.uk) across several debugging sessions.
- **sources:** old_design_and_styling/FOCUS_design_and_styling.md#"4. The Adoption Design Gap"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Victory"
- **relations:** design_reference/design_intent split; computed styles extraction; fpExtractCSSVars fix
- **verify-later:** `site_specs` rows with aspect='design_reference' for adopted sites.

<!-- SOURCE: U12_docs024_archives.md -->
### Webdesign-agent three-way design priority
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "✅ Applied" (2026-04-14 handoff).
- **what:** `analyze_design` step branches on which specs exist: design_intent present → creative freedom around described character; only design_reference → faithful reproduction, no invented palette; neither → generate from industry/audience/identity.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-14_v2.md#"Webdesign-Agent Prompt (deployed)"
- **relations:** design_reference/design_intent spec-aspect split
- **verify-later:** current webdesign-agent agent_definitions prompt text.

<!-- SOURCE: U12_docs024_archives.md -->
### Palette-locked-until-design_intent policy
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Palette is locked until design_intent exists."
- **what:** First adoption build reproduces the original palette exactly (locked); once design_intent is written, webdesign-agent gains creative freedom within the described character, letting the improvement loop evolve the palette over time.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_design_fingerprint_pipeline.md#"Key Decisions Made"
- **relations:** design_reference/design_intent split; audit loop "propose not apply"
- **verify-later:** improvement-loop audit code for propose-vs-enforce mode switch.

<!-- SOURCE: U12_docs024_archives.md -->
### Legacy monolithic CSS renderer internals (removed)
- **category:** design-composition
- **status-signal:** abandoned
- **status-evidence:** "Phase 4.3 already removed... cssTemplateData struct (and its 16 hardcoded fields)... Compile-clean."
- **what:** The original renderer held a flat struct populated by `extractDesignColors`/`designColorMaps`, loading one Go template per theme. Deleted wholesale in Phase 4.3 when the renderer switched to composable palette/layout/typography_set rows via FK.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"What Phase 4.3 already removed"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Template variable system"
- **relations:** three-layer design system (superseded); layout archetype library
- **verify-later:** grep codebase for `loadCSSGoTemplate`/`extractDesignColors`/`designColorMaps` — should be absent.

<!-- SOURCE: U12_docs024_archives.md -->
### fpExtractCSSVars regex-based CSS variable extraction (superseded)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** BEM selectors like `.btn--primary:hover` captured as fake variables; replacement uses `:root` block targeting with semicolon-splitting.
- **what:** Original extractor used one whole-stylesheet regex, producing false positives on BEM class names. Replaced with a multi-strategy extractor isolating `:root`/body/`[data-theme]` blocks, with fallback frequency analysis for utility-CSS sites.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"P6 — fpExtractCSSVars BEM False Positives"; old_design_and_styling/FOCUS_design_and_styling_fp_extract_css_vars_integration.md
- **relations:** design fingerprint extraction pipeline; computed styles extraction
- **verify-later:** `extract_design_fingerprint_action.go` — confirm regex-based extractor removed.

<!-- SOURCE: U12_docs024_archives.md -->
### css_templating.go theme-forking bridge (known-broken, scheduled rewrite)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "fork_theme_from_site produces rows with NULL palette_id, NULL layout_id, NULL typography_set_id... Adoption-forked themes are unusable by the render path."
- **what:** `TemplateCSSFromSpec` converts a rendered CSS snapshot into old flat-field-name placeholders and writes it to the legacy `css_themes.css_template` column, which the post-Phase-4.3 renderer never reads — silently producing unusable NULL-FK theme rows. Flagged for a Phase 5 rewrite.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"1. css_templating.go"; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Ready for deployment"
- **relations:** fork_theme_from_site rewrite (Phase 5); parallel legacy HTML-assembly render path
- **verify-later:** confirm `fork_theme_from_site_action.go` now produces palette/typography_set rows.

<!-- SOURCE: U12_docs024_archives.md -->
### Parallel legacy HTML-assembly render path (getThemeByID/GetThemeByName)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "css_content is populated for 13 of the 14 themes. standard-brochure has empty css_content... falls through to GetThemeByName('default')."
- **what:** A second, older render path reads `css_themes.css_content` directly into assembled HTML, independent of the spec-driven render path. Left untouched by Phase 4, own known gap flagged for resolution when Phase 7 drops legacy columns.
- **sources:** old_design_and_styling/PHASE_4_4_cleanup_summary.md#"2. getThemeByID / GetThemeByName"
- **relations:** css_templating.go bridge; legacy css_themes columns drop (Phase 7)
- **verify-later:** grep for `getThemeByID`/`GetThemeByName` call sites.

<!-- SOURCE: U12_docs024_archives.md -->
### Component-creation via HITL work-item triage (superseded)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** "migration_025_component_triage.sql — an earlier work-item-based approach that was superseded by the direct insert approach... Do not run this file."
- **what:** Earlier plan for seeding new library components via work items routed through HITL triage. Superseded by a direct SQL insert once components were designed and reviewed.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"2. What's Been Completed"
- **relations:** layout archetype library
- **verify-later:** none — historical, file explicitly marked do-not-run.

<!-- SOURCE: U12_docs024_archives.md -->
### Computed-styles extraction via browser JS injection
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Computed styles (Phase 2) deferred... Spec written but not implemented" vs. a complete Go action + workflow SQL in the Phase 2 doc.
- **what:** Supplementary fingerprint step scraping a homepage with injected JS calling `getComputedStyle()`, writing resolved values for a Go action to parse and merge — "ground truth" overriding source-CSS guesses. Fully spec'd but recorded elsewhere as deferred/not implemented.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_computed_styles_extraction_phase2.md; old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Fixes Ready But Not Deployed"
- **relations:** design fingerprint extraction pipeline; fpExtractCSSVars fix
- **verify-later:** registry.go for `extract_computed_styles`; site-adoption-agent workflow steps.

<!-- SOURCE: U12_docs024_archives.md -->
### Layout archetype library (15 named layouts)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Phase 1 is next: designing and writing the 15 layout CSS templates" → "Phase 1 — Layouts seeded (15 rows in layouts table)... deployed."
- **what:** Taxonomy of 15 named structural/visual archetypes (brochure-formal, portfolio-kinetic, utility-tool, media-grid, etc.), each with character/structural-trait descriptions, default header/footer/typography, and legacy-theme mappings — the target library for the composable-theme migration's `layouts` table.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"4. The 15 Layouts — Detailed Descriptions"; docs024_key_docs_latest/025_palette_layout_typography_migration(3).md
- **relations:** composable theme system; site-design-planner layout resolver
- **verify-later:** `layouts` table rows in DB — confirm 15 rows.

<!-- SOURCE: U12_docs024_archives.md -->
### Palette merge rule (core slots vs specialised slots)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Core slots (spec wins where present)... Specialised slots (theme wins)."
- **what:** When a site composes a theme, core palette slots let the site's own spec win when present; specialised slots (primary_hover, hero_title, cta_bg, etc.) always take the theme's value.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md#"Palette merge rule"
- **relations:** layout archetype library; site-design-planner palette resolver
- **verify-later:** `resolve_composition_palette_action.go` merge logic.

<!-- SOURCE: U12_docs024_archives.md -->
### site-design-planner "Choice B" scope (composition-only)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Choice B adopted. The agent's exclusive responsibility is composition resolution... It does NOT write navigation or layout specs."
- **what:** Decision narrowing site-design-planner to write exactly one spec, `resolved_composition` (palette_id/layout_id/typography_set_id + lineage + reasoning), deferring `navigation`/`layout` spec ownership to future specialist agents — justified by "slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"3. Scope Refinement"
- **relations:** composition resolution architecture; design pipeline guiding principles
- **verify-later:** agent_definitions row for site-design-planner — confirm workflow only writes `resolved_composition`.

<!-- SOURCE: U12_docs024_archives.md -->
### Composition resolution architecture (3 resolvers + install action)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "validate_composition_inputs_action.go — DONE... install_site_composition_action.go — DONE (~562 lines)."
- **what:** site-design-planner pipeline: `validate_composition_inputs` → three resolvers (`resolve_composition_layout` tag-overlap match, `resolve_composition_typography`, `resolve_composition_palette` fingerprint→mission→design_intent→layout-inherit→default cascade) → `install_site_composition` (one transaction: css_themes+style_collections insert, sites update, resolved_composition spec write).
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"4. Work Plan — Deliverable 4"
- **relations:** site-design-planner Choice B scope; composition resolver orphan-rows policy; fork_theme_from_site
- **verify-later:** confirm `resolve_composition_*.go`/`install_site_composition_action.go` in current codebase.

<!-- SOURCE: U12_docs024_archives.md -->
### webdesign-agent install/render ordering bug ("first render wrong layout")
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "9.1 Known ordering issue in webdesign-agent... This is the exact 'first render wrong layout' bug site-design-planner was built to eliminate."
- **what:** webdesign-agent ran `generate_css → deploy_css → ... → install_theme`, so any site without a pre-installed composition hit the emergency fallback and committed it to git before the correct composition was installed a step later. Documented, deferred fix: reorder `install_theme` before `generate_css`.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"9.1 Known ordering issue"
- **relations:** composition resolution architecture; render_css_from_spec_action emergency fallback
- **verify-later:** webdesign-agent workflow step order.

<!-- SOURCE: U12_docs024_archives.md -->
### Fork_theme step double-creation guard
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** "Guard: require both should_fork_theme AND should_promote_to_library flags. Implementation deferred to Deliverable 6."
- **what:** Once site-design-planner runs, the pre-existing `fork_theme` step in webdesign-agent risks creating duplicate theme/collection rows. Documented mitigation requires two flags both true before forking proceeds.
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"6. Risks Still Live"
- **relations:** composition resolution architecture; fork_theme_from_site rewrite
- **verify-later:** `fork_theme` step config in webdesign-agent — confirm both flags gate execution.

<!-- SOURCE: U12_docs024_archives.md -->
### Design pipeline guiding principles (mottos)
- **category:** design-composition
- **status-signal:** unknown
- **status-evidence:** "Principles Restated" section repeated verbatim across 2026-04-19 handoffs, sourced from `007_adoption_pipeline_v2.md` and a FOCUS work-plan doc.
- **what:** A shared decision-shorthand invoked to settle scope questions: "Every build conceptually an adoption," "Design reference is history, design intent is direction," "Adoption is a starting point, not a ceiling," "LLM for reasoning, Go for extraction," "Handlers are self-contained," "Slim strict responsibilities."
- **sources:** old_design_and_styling/HANDOFF_2026-04-19_design_and_styling_composable_theme_and_site_design_planner_update4(3).md#"7. Principles Restated"
- **relations:** site-design-planner Choice B scope; design_reference/design_intent split
- **verify-later:** none — a documentation/culture artifact, not directly code-verifiable.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Colour Inheritance Model
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** every layout_NN header lists this as CONTRACT CHECK #1 and the CSS body implements it identically (e.g. layout_01 lines ~160-182: `h1..h6 { color: var(--section-heading, var(--color-primary)); }`)
- **what:** The rule that element-level colour rules (headings, body text, links) resolve via a two-tier CSS custom-property fallback chain: `var(--section-*, var(--color-*))`. This lets a "dark section" component override just the `--section-*` variable on its own container without any layout needing to restate rules. Applied identically across all 17 layout templates as CONTRACT CHECK #1.
- **sources:** layouts/layout_01_brochure-formal.sql#header+L160-182, layouts/layout_02_brochure-bold.sql#header, layouts/layout_16_17_vonc_gamesdesign.sql#L832, layouts/layout_10_high-energy.sql#header
- **relations:** Dark Section Variable Contract; template helper system; renderer-managed surface sections
- **verify-later:** render_css_from_spec_action.go; every layout's `:root` and base element rules

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dark Section Variable Contract / buildSectionDefaults renderer behaviour
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** layout_01 lines 268-289: "TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync with buildSectionDefaults in render_css_from_spec_action.go... Tracked as Phase 4.5 in 025_palette_layout_typography_migration."
- **what:** Layout templates must NOT declare `--section-*` defaults on section containers; a Go renderer function `buildSectionDefaults` appends `--section-*` overrides after rendering, chosen by palette luminance. Five renderer-managed surface classes (`.features-section`, `.services-section`, `.differentiators-section`, `.about-section`, `.faq-section`) are hardcoded on both sides and must be kept in sync; hero/CTA/testimonials/contact are excluded as component-owned. One documented exception: a palette-declared `heading` slot emits a root-level `--section-heading`.
- **sources:** layouts/layout_01_brochure-formal.sql#L14-32,L268-289, layouts/layout_02_brochure-bold.sql#header, layouts/layout_16_17_vonc_gamesdesign.sql#L85-93
- **relations:** Colour Inheritance Model; template helper system; layout archetype concepts (all 17)
- **verify-later:** render_css_from_spec_action.go buildSectionDefaults; docs025/026/027 Phase 4.5 status

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Template helper system ({{palette}}/{{typo}}/{{token}} with fallback)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** every layout file's CONTRACT CHECK #4; 003_palettes_seed.sql: "Key naming... matches the {{palette \"primary_hover\" \"...\"}} template helpers"
- **what:** A Go-template-style substitution convention embedded in the `css_template` CSS text: `{{palette "key" "fallback"}}`, `{{typo "key" "fallback"}}`, `{{token "key" "fallback"}}`, each resolving a JSONB slot lookup with a mandatory literal fallback. A `{{with palette "heading" ""}}...{{end}}` conditional-block variant is also used.
- **sources:** layouts/layout_01_brochure-formal.sql#L33-34,L89-138, layouts/003_palettes_seed.sql#L14-19, layouts/layout_16_17_vonc_gamesdesign.sql#L96-145
- **relations:** Colour Inheritance Model; structure_tokens JSONB convention; palettes table; typography_sets table
- **verify-later:** the Go renderer executing these templates; helper lookup precedence (site-adopted vs seed palette)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### structure_tokens JSONB convention
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** present as a populated JSONB literal column in the INSERT of all 17 layout seed files
- **what:** Each layout row carries a `structure_tokens` JSONB column holding non-colour design tokens — spacing, radii, shadows, transitions, and layout-specific one-offs (e.g. `diagonal_slope_top` for high-energy, `split_pane_left/right` for tool-first-landing). The layout-level counterpart to the palette/typography tables; explicitly excluded from palette extraction.
- **sources:** layouts/layout_01_brochure-formal.sql#L55-69, layouts/layout_10_high-energy.sql#L38-53, layouts/003_palettes_seed.sql#L39-41
- **relations:** template helper system; palettes table; per-layout archetype concepts
- **verify-later:** `layouts` table column `structure_tokens` DDL/constraints in Phase 2 migration

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Seed-driver transactional load pattern
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 003_layouts_seed_driver.sql: `BEGIN;` ... 15 `\ir` includes ... verification block asserting `actual_count >= 15` ... `COMMIT;`
- **what:** A psql driver script wrapping all 15 numbered layout `\ir` includes in a single transaction with `\set ON_ERROR_STOP on`, so any single layout's error rolls back the entire batch. Ends with a `DO $verify$` block raising an exception if the seeded row count is below expected. Each INSERT is itself idempotent (`ON CONFLICT (name) DO UPDATE`).
- **sources:** layouts/003_layouts_seed_driver.sql (full file), layouts/003_palettes_seed.sql#verify block, layouts/003_typography_sets_seed.sql#verify block
- **relations:** palettes table/seed; typography_sets table/seed; all 15 numbered layout archetypes
- **verify-later:** Phase 2 migration creating palettes/layouts/typography_sets tables; confirm this driver actually ran against the live DB

<!-- SOURCE: U13_docs024_small_dirs.md -->
### palettes table / seed (CSS-theme-extracted colour slots)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Diagnostic run (Phase 3 preflight) confirmed... 13 rows have palette data in css_content"; verify block asserts `actual >= 13`
- **what:** `palettes` table stores one row per design palette (`name`, `display_name`, `colours` JSONB slot map, `category`, `industry_tags`, `origin`, `is_active`). The seed migrates 13 legacy `css_themes` rows via a PL/pgSQL helper `_extract_css_palette` that regex-parses `--color-KEY: VALUE;` declarations. Non-colour vars are deliberately excluded (belong to structure_tokens). One theme (`standard-brochure`) has no palette of its own, mapped to `default` in a later step.
- **sources:** layouts/003_palettes_seed.sql#header,#_extract_css_palette function,#insert+select,#verify+report
- **relations:** template helper system; structure_tokens JSONB convention; typography_sets table; css_themes (legacy source table)
- **verify-later:** `css_themes` table; confirm the "Phase 3 Step 3" theme-mapping UPDATE actually ran

<!-- SOURCE: U13_docs024_small_dirs.md -->
### typography_sets table / seed (6 named font/scale bundles)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Seeds the 6 typography sets described in the migration plan section 8"; verify block asserts `actual_count >= 6`
- **what:** `typography_sets` table stores 6 named typography bundles — sans-modern, serif-editorial, display-bold, mono-technical, serif-classical, sans-friendly — each with `fonts` JSONB and `scale` JSONB, plus `category`/`industry_tags`. Layouts reference these via `{{typo "key" "fallback"}}`. Each set's description names which layout archetypes it pairs with (documented convention, not FK-enforced).
- **sources:** layouts/003_typography_sets_seed.sql#header,#sans-modern,#display-bold,#mono-technical,#serif-classical/sans-friendly,#verify
- **relations:** palettes table; template helper system; structure_tokens JSONB convention
- **verify-later:** confirm each layout's declared "Default typography" matches typography_sets.name at composition time

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout-resolution-by-tags gap (resolveLayoutByTags fallback problem)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** layout_16_17_vonc_gamesdesign.sql header: "Root cause reminder: resolveLayoutByTags matches tagSet against layout.industry_tags. The classifier doesn't currently emit those two fields, so tagSet is empty → fallback path 'no classification tags'... Neither migration alone is sufficient."
- **what:** The site-design-planner's layout picker (`resolveLayoutByTags`) intersects a site's classification tags against each layout row's `industry_tags`; when the classifier doesn't emit those fields, every site falls back to `brochure-formal` regardless of fit — exactly what happened to gamesdesign.co.uk. Fixing it requires two coordinated migrations: this file (007) seeding the missing layouts, and a separate 008 migration (not in scope) updating the classifier prompt to emit category/industry_tags.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L1-38
- **relations:** tool-portal-dark; social-lobby; all 15 numbered layouts (as matching candidates)
- **verify-later:** migration "008" classifier prompt; `resolveLayoutByTags` function location; live check whether sites still fall back to brochure-formal

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: brochure-formal
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phase 1 of 025_palette_layout_typography_migration"; satisfies all 5 numbered CONTRACT CHECKS
- **what:** Structured, understated, CTA-driven brochure layout with corporate restraint. Mapped to themes `default`, `standard-brochure`, `professional-dark`. Suits consultancies, law, finance, B2B. Serves as the canonical reference implementation of all 5 contract checks and as the fallback layout when tag resolution fails.
- **sources:** layouts/layout_01_brochure-formal.sql#L1-37,L50-69
- **relations:** Colour Inheritance Model; Dark Section Variable Contract; brochure-bold; Layout-resolution-by-tags gap
- **verify-later:** `layouts` row name='brochure-formal'; confirm still the de facto fallback

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: brochure-bold
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Phase 3 of 025_palette_layout_typography_migration"
- **what:** High-energy conversion variant of brochure-formal — tall hero, gradient accents, display-bold typography, strong CTAs. Suits tech startups, SaaS, fitness brands.
- **sources:** layouts/layout_02_brochure-bold.sql#L1-30,L43-65
- **relations:** brochure-formal; Dark Section Variable Contract; typography_sets(display-bold)
- **verify-later:** `layouts` row name='brochure-bold'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: portfolio-kinetic
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header explicit "STRUCTURAL DIVERGENCE from brochure-* layouts" list; "Mapped themes: none currently"
- **what:** Asymmetric, motion-forward, display-type-led layout for creative-studio energy — animated underline text-links instead of hero/CTA buttons, 40/60 asymmetric columns, dense-packed work showcase, narrower 1140px container. Suits design studios, creative agencies, photography portfolios.
- **sources:** layouts/layout_03_portfolio-kinetic.sql#L1-33,L46-66
- **relations:** brochure-formal (contrast); typography_sets(serif-classical alt); Colour Inheritance Model
- **verify-later:** `layouts` row name='portfolio-kinetic'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: magazine-grid
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: content-modern"
- **what:** Publication-feel layout: top-level 2/3 main + 1/3 sidebar grid, article cards, featured-article variant, sidebar widgets, serif-editorial typography. Suits news, opinion, long-form blogs.
- **sources:** layouts/layout_04_magazine-grid.sql#L1-35,L37-70
- **relations:** typography_sets(serif-editorial); soft-editorial; industry-hub
- **verify-later:** `layouts` row name='magazine-grid'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: utility-tool
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: none — exists for selector/adoption matching"
- **what:** Minimal-chrome layout where "the tool is the reason" — narrowest container (800px), compact header, single tool card with output region, no card-grids, larger form controls. Suits online calculators, converters, developer utilities.
- **sources:** layouts/layout_05_utility-tool.sql#L1-25,L27-59
- **relations:** tool-first-landing (explicit divergence); typography_sets(sans-modern)
- **verify-later:** `layouts` row name='utility-tool'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: media-grid
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: none"; dark-mode-by-default palette
- **what:** Thumbnail-dominant, continuous-scroll discovery layout — auto-fill fluid grid, optional featured/pinned item, scrollable chip filter bar, "featured row"/horizontal-scroll shelf variants, fixed aspect-ratio tokens. Suits video platforms, audio libraries, image galleries. Dark theme by default.
- **sources:** layouts/layout_06_media-grid.sql#L1-24,L26-58,L67-90
- **relations:** high-energy; tool-portal-dark; Colour Inheritance Model
- **verify-later:** `layouts` row name='media-grid'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: docs-sidebar
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Default typography: mono-technical" matches typography_sets seed row's own note
- **what:** Reference-grade documentation layout — 3-zone CSS grid (fixed sidebar nav, main reading column, collapsing table-of-contents). Code blocks get accent-border + copy-button; admonitions use `.callout` variants. Suits developer docs, API references, knowledge bases.
- **sources:** layouts/layout_07_docs-sidebar.sql#L1-25,L27-58
- **relations:** typography_sets(mono-technical); tool-portal-dark
- **verify-later:** `layouts` row name='docs-sidebar'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: soft-editorial
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: bakery, warm-friendly, calm-minimal, soft-editorial" — the only numbered layout with 4 named theme mappings
- **what:** Warm, reading-first, organic layout — tinted background, pill-shaped buttons, barely-there card borders, serif display headings, transparent floating header, 1.75 line-height. Suits wellness blogs, lifestyle sites, personal essays, bakeries.
- **sources:** layouts/layout_08_soft-editorial.sql#L1-23,L25-57
- **relations:** typography_sets(serif-editorial, sans-friendly); magazine-grid; industry-hub
- **verify-later:** `layouts` row name='soft-editorial'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: technical-precise
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: premium-elegant (with serif override), modern-engineering-clean"
- **what:** "Engineered" layout — glass-effect header (backdrop-filter blur) as its signature moment, tight border-radius, bordered/low-shadow cards, flat solid CTAs, light (not dark) footer contrasted against brochure-*'s dark footers. Suits SaaS platforms, infrastructure products, engineering consultancies.
- **sources:** layouts/layout_09_technical-precise.sql#L1-25,L27-58
- **relations:** typography_sets(sans-modern default, serif-classical override); brochure-formal (footer contrast)
- **verify-later:** `layouts` row name='technical-precise'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: high-energy
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Mapped themes: boxing" (narrowest mapping of all 15)
- **what:** Aggressive, kinetic layout — uppercase headings, 80vh dark hero, diagonal clip-path section separators, zero border-radius, hard offset shadows, numeral-prefixed feature cards. Suits boxing gyms, combat sports, fitness events. Uses display-bold typography.
- **sources:** layouts/layout_10_high-energy.sql#L1-20,L22-53
- **relations:** typography_sets(display-bold); media-grid
- **verify-later:** `layouts` row name='high-energy'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: comparison-aggregator
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header distinguishes itself from 3 sibling commerce-adjacent layouts by its defining primitive `.result-card`
- **what:** Search-first, data-dense, trust-oriented layout — hero IS a search input, sticky filter bar, dense horizontal result-card rows, regulatory info banners, heavy disclaimer footer. First of four deliberately-differentiated "commerce-adjacent" layouts. Suits price/insurance/broadband comparison, trade directories.
- **sources:** layouts/layout_11_comparison-aggregator.sql#L1-24,L26-60
- **relations:** affiliate-hub; ecommerce-storefront; industry-hub; tool-first-landing
- **verify-later:** `layouts` row name='comparison-aggregator'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: affiliate-hub
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's explicit divergence table against comparison-aggregator and ecommerce-storefront
- **what:** Product-review/buyer-guide layout — persistent disclosure strip, vertical product "picks" cards, pros/cons review blocks, horizontally-scrolling comparison tables, optional sticky "Top Picks" sidebar. Suits product review sites, "best X for Y" guides, deal aggregators.
- **sources:** layouts/layout_12_affiliate-hub.sql#L1-21,L23-56
- **relations:** comparison-aggregator; ecommerce-storefront; industry-hub
- **verify-later:** `layouts` row name='affiliate-hub'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: ecommerce-storefront
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's divergence note vs affiliate-hub (cover-fit lifestyle photography vs contain-fit product-on-white)
- **what:** Retail-clean, product-forward storefront — promo hero, image-overlay category tiles, product grid, add-to-cart CTAs, strike-through sale pricing, CSS-only mini-cart dropdown structure, trust-bar strip. Suits independent shops, small-catalogue retailers.
- **sources:** layouts/layout_13_ecommerce-storefront.sql#L1-24,L26-60,L94-97
- **relations:** affiliate-hub; comparison-aggregator
- **verify-later:** `layouts` row name='ecommerce-storefront'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: tool-first-landing
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's explicit divergence from utility-tool (full-container vs 800px narrow column)
- **what:** Full-container (up to 1400px) tool-dominated landing page where "the tool IS the page" — defining primitive `.split-pane` (50/50 default), dark-mode-friendly, optional tabbed interface. The "loud" counterpart to utility-tool's contained/quiet version. Suits calculators, API playgrounds, demo tools.
- **sources:** layouts/layout_14_tool-first-landing.sql#L1-22,L24-56
- **relations:** utility-tool; tool-portal-dark
- **verify-later:** `layouts` row name='tool-first-landing'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: industry-hub
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** header's 4-way divergence table naming this the only non-commercial member of the "commerce-adjacent" family
- **what:** Vertical information-authority layout — "About this site" independence-claim banner, `.directory-card`/`.guide-card`/`.news-card`/`.glossary-list` primitives, ordered directory→guides→news→reference, serif-editorial typography for "authority without being corporate." Suits regulatory information hubs, industry explainer sites.
- **sources:** layouts/layout_15_industry-hub.sql#L1-28,L30-61
- **relations:** comparison-aggregator; affiliate-hub; ecommerce-storefront; magazine-grid/soft-editorial
- **verify-later:** `layouts` row name='industry-hub'

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: tool-portal-dark
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** seeded by migration "007_seed_layouts_tool_portal_and_social_lobby.sql" explicitly framed as necessary-but-not-sufficient; `needs_review` column present on the INSERT
- **what:** Dark developer-utility portal layout supporting three page shapes in one template — portal/index, tool pages, article/guide pages (narrow reading column). Dark-mode-first, flat technical aesthetic. Built specifically to close the layout-library gap that caused gamesdesign.co.uk to fall back to brochure-formal.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L1-38,L55-145,L71-94
- **relations:** Layout-resolution-by-tags gap; social-lobby; docs-sidebar; media-grid
- **verify-later:** `layouts` row name='tool-portal-dark', `needs_review` flag value; migration "008"

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Layout: social-lobby
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** same migration-007 framing as tool-portal-dark; `needs_review` column present
- **what:** Light, colour-forward social-platform layout built around a room/lobby metaphor. Primary UI unit is the "provocation card"; Arena (competitive) and Stage (creative) rooms differentiated via dedicated palette slots (`arena`, `stage`) rather than component variants. Four page shapes: lobby/homepage, room/topic index, provocation detail, archetype/profile. Reaction-colour slots (`reaction_positive`/`reaction_negative`/`reaction_meta`) are a distinctive palette extension. Named target: vonc.com.
- **sources:** layouts/layout_16_17_vonc_gamesdesign.sql#L21-23,L713-757,L759-810
- **relations:** Layout-resolution-by-tags gap; tool-portal-dark; vonc workstream (site-case-studies); palettes table (arena/stage/reaction slots)
- **verify-later:** `layouts` row name='social-lobby'; live check against vonc.com

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Design fingerprint & design_reference vs design_intent
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 007 "Design Fingerprint Pipeline (added 2026-04-12)"; principle 7 "Design reference is history, design intent is direction"
- **what:** A Go extractor (`extract_design_fingerprint`, goquery) parses crawled rawHTML/external CSS into a fingerprint with a `suggested_mapping`; an LLM (`generate_design_intent`) turns it into a semantic brief. `design_reference` is an immutable historical record; `design_intent` is forward-looking direction — evolution happens by updating intent, never reference.
- **sources:** WM/007_adoption_pipeline_v3.md#design-fingerprint-pipeline-added-2026-04-12, WM/007_adoption_pipeline_v3.md#design-evolution-lifecycle, WM/FOCUS_interactive_content_generation(3).md#adoption-captures-content-and-extracts-structured-design-data
- **relations:** site adoption agent; interactive parse-stage; webdesign-agent three-way priority
- **verify-later:** enrich_fingerprint_with_css_action.go; site_specs design_reference/design_intent

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Component selector + creator (section_type vs function)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 007 "Phase 3 — Component selector, patterns, and research" (planned); "The separation: Planner decides WHAT section types … Component selector decides WHICH template"
- **what:** Splits the planner's conflated role: the planner picks section_types, a Go component-selector scores templates by metadata with a fallback to `needs_new_component`, and a `component-creator` agent LLM-generates a template from the full component contract when none fits. `function` currently does two jobs (page-role identifier + template choice); `section_type` separates them.
- **sources:** WM/007_adoption_pipeline_v3.md#component-selector-and-creator, WM/007_adoption_pipeline_v3.md#component-creation-contracts, WM/FOCUS_interactive_content_generation(3).md#components-more-broadly
- **relations:** interactive content generators; site plan sections; tool/game library model
- **verify-later:** content_components metadata columns; component-creator agent; plan_sections

<!-- SOURCE: U18_sql_for_agents.md -->
### chief-strategist (build-plan LLM) + component placement dedup rules
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** 040 still upgrades its model (haiku→sonnet) but the work-item pipeline planner (build-site-planner, 053) owns planning thereafter; 019 patch injects "COMPONENT PLACEMENT RULES" into its prompt.
- **what:** The v1/v2 planning agent producing sections/component_details build plans. Its lasting contribution is the component placement rule-set injected by 019: testimonials/team-grid/faq/contact-form on ONE page only, per-page hero variants, no duplicated services content, merge similar pages — an early anti-repetition contract for planners.
- **sources:** 019_chief_strategist.sql; sql_for_agents_v1/019_chief_strategist.sql; sql_for_agents_v2/019_chief_strategist.sql; 040_optimise_which_llms.sql
- **relations:** site-planner, build-site-planner inherit the planning role; parse_json_field/unwrapDeep pattern (v1/019)
- **verify-later:** is chief-strategist still active or deleted

<!-- SOURCE: U18_sql_for_agents.md -->
### webdesign-agent (full CSS stylesheet generation)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 4,683-line definition file; referenced as the "full theme" path by 076 ("Unlike webdesign-agent (which regenerates everything from scratch)"); idle timeout in 075.
- **what:** Generates production CSS for a site. Accepts a provided site_context or loads context from DB (conditional first step), analyzes design requirements, writes stylesheet via git_commit with file_path config. It is the heavyweight regeneration path, contrasted with css-patch-agent for targeted fixes.
- **sources:** 031_webdesign_agent.sql; 076_css_patch_agent.sql; 103_site_design_planner.sql (resolved_composition reader list)
- **relations:** site-scraper feeds it site_context; css themes/style_collections; site-design-planner
- **verify-later:** current webdesign-agent workflow vs 031 copy; patch_01_git_commit_file_path.go

<!-- SOURCE: U18_sql_for_agents.md -->
### site-design-planner spec aspects (navigation / layout / resolved_composition)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 103 is "Deliverable 2: Spec schemas... pre validation" — documents shapes and creates best-effort validation functions, explicitly not table constraints; reader lists name live actions.
- **what:** Defines three site_specs aspects the site-design-planner writes, separated by reader: navigation (nav architecture, items, CTA, mobile pattern → populate_nav_tables, InjectHeader, GetNavItems), layout (page-level layout, header/footer style → AssembleMultipageSiteAction, templates), resolved_composition (machine-readable pointers to palette/layout/typography rows + reasoning → render_css_from_spec, webdesign-agent, audit agents). Validation functions run at write time; site_specs stays open JSONB.
- **sources:** 103_site_design_planner.sql
- **relations:** design-composition docs 025/026/027; webdesign-agent; nav-updater
- **verify-later:** site-design-planner agent existence and writers of these aspects

<!-- SOURCE: U19_sql_tables_components.md -->
### Style collections
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** Two generations of the migration in 001 (initial + 030_style_collections), sites.style_collection_id FK, seed collections professional-dark / minimal-light / bold-gradient with linked header/footer components.
- **what:** A style collection bundles the components and tokens defining a site's visual identity: header/header-home/footer component ids, css_theme_id, color_palette and typography JSONB, category and industry_tags. Sites link to one collection and may override via sites.style_overrides without forking the collection. Original motivation: replace inconsistent LLM-generated headers with tested templates.
- **sources:** docs/agent_docs/sql_for_components/001_style_collections.sql; docs/agent_docs/sql_for_components/003_styles_implementation.md; docs/agent_docs/sql_for_components/002_styles_documentation.md
- **relations:** component-based headers; palette/layout/typography decomposition; design lineage columns.
- **verify-later:** style_collections rows; assignment logic in EnsureSiteRecordAction / classification.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component-based headers replacing LLM-generated chrome
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** 002/003 md docs lay out the plan (store tested header templates, render with site data, inject replacing LLM header); 012 executes population and SQL-side rendering of site_components for header/footer/head.
- **what:** The founding decision that page chrome (header/footer/head) is never LLM-generated per page: tested templates render with a site-derived context (logo from domain, nav from pages/nav tables, colours from collection+overrides) and are injected at assembly. Benefits table: consistency, instant DB-side updates, A/B-able collections.
- **sources:** docs/agent_docs/sql_for_components/002_styles_documentation.md; docs/agent_docs/sql_for_components/003_styles_implementation.md; docs/agent_docs/sql_for_tables/012_site_components.sql
- **relations:** style collections; site/area/page component hierarchy; template syntax unification.
- **verify-later:** RenderHeaderForSite / render_site_components action.

<!-- SOURCE: U19_sql_tables_components.md -->
### Palette / layout / typography decomposition (migration 025 phase 2)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "Three new tables, empty after this migration... new columns are read only once Phase 4 ships. Phase 3 seeds... Phase 7 drops the legacy columns" (038 header).
- **what:** Splits css_themes.css_template's conflated concerns into three independently versioned tables: palettes (free-shape colours JSONB consumed via {{palette "key" "fallback"}}), layouts (Go CSS template + structure_tokens + default header/footer component ids), typography_sets (fonts + scale via {{typo}}). css_themes becomes a composition row via nullable FKs; renderer migrates in later phases. Also created 10 library layout components (header-with-categories, header-docs, directory-listing, product-grid, etc.).
- **sources:** docs/agent_docs/sql_for_tables/038_style_collections.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components
- **relations:** style collections; design lineage; site_plan_sections resolved palette/layout/typography ids.
- **verify-later:** palettes/layouts/typography_sets row counts; renderer read path (phase 4); legacy column drops (phase 7).

<!-- SOURCE: U19_sql_tables_components.md -->
### Design-asset fork lineage (origin / needs_review / source_site)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** 038 Part 2(b): lineage columns "required by the Phase 5 fork_theme_from_site action. A prior session reported them as already added but the current schema shows them absent... nothing needs review (fork action hasn't shipped yet)".
- **what:** Uniform provenance on palettes, layouts, typography_sets, css_themes and style_collections: origin ('seed' default), needs_review, forked_from_<entity>_id, source_site_id, source_domain, forked_at. Enables adopting a live site's design into the library as a reviewed fork.
- **sources:** docs/agent_docs/sql_for_tables/038_style_collections.sql#PART2
- **relations:** adoption-pipeline (design adoption); tool fork model (same pattern for tools).
- **verify-later:** fork_theme_from_site action existence; any rows with origin != 'seed'.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Brand designer agent (theme selection)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Agent SQL + mvp-site-builder workflow insertion (spawn/call_brand_designer feeding brand_theme to the architect); superseded by content-creator's theme recommendation + semantic tag matching in 006semantic_themes, and later by the design-composition system.
- **what:** An LLM agent that analyses domain + objective and picks a CSS theme from the named library (boxing, bakery, tech, professional-dark, default) with reasoning — the first brand/design decision point in the pipeline.
- **sources:** docs004_website_capture_project/website_analysis/README.018.brand_designer_agent.md
- **relations:** semantic CSS theme system; successor: site-design-planner / palette resolution (design-composition docs 025-027).
- **verify-later:** brand-designer agent_definitions row.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Semantic CSS theme and snippet system (theme_tags, css_themes, css_snippets, js_snippets)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** Full DDL + seed data in two iterations (020 text[]; 027 jsonb) with helper matching functions; themes are complete `:root` CSS-variable palettes. The design-composition palette/typography system is the taxonomy-named successor.
- **what:** A semantic tagging vocabulary (mood/style/industry/audience/functional/colour tags with related_tags pairing) applied to: css_themes (full CSS-variable palettes: calm-minimal, bold-conversion, warm-friendly, dark-modern, premium-elegant…), css_snippets (hover/animation/effect/pattern/utility fragments), and js_snippets (nav, scroll animations, accordion, clipboard, form interactions with trigger metadata). Content-creator recommends theme + theme_tags; assembler matches snippets by tags. All theming via CSS variables — the ancestor of the platform's CSS-variable contract.
- **sources:** docs004_website_capture_project/006semantic_themes/README.020.brand_theme_preparation.md; docs004_website_capture_project/007different_types_of_site/027_css_js_schema.sql; docs004_website_capture_project/006semantic_themes/README.021.semantic_themes_agent_definitions.md
- **relations:** successors: contracts-and-standards CSS variables; design-composition palette resolution; styling-render-pipeline.
- **verify-later:** css_themes/css_snippets/js_snippets/theme_tags tables today.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Style collections as the design bridge
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** docs017/017 three-layer model (HTML components / CSS theme / style collection "the bridge"); docs012/009 migration 030_style_collections_migration.sql; per docs015/004 "load style collections" is a standard planner step.
- **what:** Layer 3 of the design system: a style_collection binds a site to specific header/footer/head component choices plus a CSS theme (colors, typography), selected per site (stored on sites, or chosen by domain keywords as fallback). Enables mix-and-match of structure and appearance and consistent chrome across the multipage path. Ancestor of the current palette/typography/layout resolution system.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/017_agent_architecture_v2.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Design-System-Layers; docs012_site_maps_and_components/009_assemble_from_library_vs_component_library.md
- **relations:** css_themes; webdesign-agent; design agent family split; current design-composition docs 025-027.
- **verify-later:** style_collections table shape and GetStyleCollectionForSite.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Design agent family split (brand-designer / style-generator / layout-architect)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** docs017/019b: webdesign-agent "Exists, prompt updated for colour inheritance"; brand-designer/style-generator "Future split"; layout-architect/nav-layout-agent "New" (never appear later); "There's no rush on this split."
- **what:** Decompose the monolithic webdesign-agent (analyse_design → generate_css → update_site, deploying /assets/css/styles.css) into: brand-designer producing a rarely-changing brand_spec (palette, type scale, spacing, tone, image direction) in sites.content_data; style-generator producing CSS with theme-library search-and-adapt before generating fresh (feeding css_themes for reuse); layout-architect producing per-page-type layout definitions (nav placement, content zones, max components) with rendering fallbacks. Direct ancestor of the current site-design-planner / design composition system.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#3-Design-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/003_design.md
- **relations:** style collections; colour inheritance; current design-composition docs 025-027 (successor).
- **verify-later:** brand_spec/layout_definitions keys in sites.content_data; whether split agents exist.

<!-- SOURCE: U22_recent_small_docs.md -->
### Vertical-specific planner variants
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** Phase 3.5 todo: "Create veterinary/energy/mortgage/seasonal site planner prompt variant" — all unchecked.
- **what:** Separate agent definitions using the same planner Go code but vertical-tuned prompt templates, so a well-established vertical produces better plans than a generic planner with config injected. Each knows its vertical's page types, conversion funnel, and per-page guidance (e.g. every breed-health page links to "find a vet for this breed"; every mortgage calculator has lead capture below results).
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#3.5
- **relations:** site-planner, vertical knowledge architecture, unified site spec
- **verify-later:** agent_definitions for veterinary/energy/mortgage/seasonal site-planner variants

<!-- SOURCE: U23_docs_root_vonc.md -->
### Post-025 CSS theme flow (empty css_content by design; composition via FK chain)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES 2026-06-24 ~16:00 citing doc 027 and install_site_composition_action.go L210-212: css_content "intentionally empty — post-025 renderer reads composition via FK chain at render time"; styles.css deployed by webdesign-agent.
- **what:** The design pipeline runs needs_composition (site-design-planner) → gated needs_design (webdesign-agent: analyze_design → update_site → generate_css via render_css_from_spec reading composition FKs → deploy_css writes assets/css/styles.css → optional fork_theme). `css_themes.css_content` is intentionally empty post-025; the empty "Theme-specific styles injected here" head block is expected, not a bug. webdesign-agent is not deprecated. Key debugging consequence: a wrong colour on a page is more likely a component variable-name mismatch than a theme-injection failure.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-24-~16:00; docs/RUNBOOK_vonc_migrations(14).md#step-6
- **relations:** CSS variable naming; two chrome assembly paths (stale renders)
- **verify-later:** install_site_composition_action.go; render_css_from_spec

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Design fingerprint extraction (adoption fidelity)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** FOCUS_design_and_styling_adoption_problems (gamedesign.uk case): "rawHTML is captured but never parsed for design data"; WORK_PLAN_v2 (2026-04-11) Phase 1 `extract_design_fingerprint` Go action "Code ready — needs deploying", `design_reference` spec replacing vague `design`; live successors 025/026/027 design docs exist.
- **what:** Mechanism to parse crawled rawHTML `<style>` blocks, CSS variables, Google-Fonts links, and layout into a concrete `design_reference` spec aspect (hex values, font stacks, `suggested_mapping` from source→our CSS variable names), replacing the LLM's guessed `design` spec so adopted sites reproduce the original's colours/fonts/layout instead of generic component defaults.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_problems.md#a-fingerprint-extraction; FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#phase-1
- **relations:** superseded by 025_palette_layout_typography_migration / 027_design_and_site_planner_v2; design_intent; adoption pipeline
- **verify-later:** extract_design_fingerprint_action.go; site_specs aspect design_reference

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Three-way webdesign priority (design_intent → design_reference → industry)
- **category:** design-composition
- **status-signal:** superseded
- **status-evidence:** WORK_PLAN_v2 Decisions #3: "design_intent (creative freedom) → design_reference (reproduce faithfully) → generate from industry (new builds)"; palette "locked until design_intent is written"; Phase 2b "Applied to DB".
- **what:** The webdesign-agent prompt resolves imagery/palette by priority: honour `design_intent` (semantic creative brief, auto-generated from `design_reference` via LLM in adoption Phase 2e) first, else reproduce `design_reference` faithfully, else generate from industry. The palette can only change once design_intent exists.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#decisions-made, #phase-2
- **relations:** superseded by live design-composition docs 025/026/027; design fingerprint; imagery_direction
- **verify-later:** webdesign-agent analyze_design prompt in agent_definitions

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Site-design-planner agent (structure × identity × effects)
- **category:** design-composition
- **status-signal:** abandoned
- **status-evidence:** WORK_PLAN_v2 Phase 3 "Site-Design-Planner Agent (not started)" — all of 3a-3g "Not started"; Phase 4 requirement-driven components also "Not started"; superseded by live 027_design_and_site_planner_v2.md.
- **what:** Proposed dedicated `site-design-planner` agent (Option B) decomposing site design into structure × identity × effects, owning navigation/layout spec schemas, wired into the build pipeline to drive header/footer selection and hero/nav merging — plus requirement-driven component selection generating custom components when the library has no match. Never built as specified; replaced by the v2 design/site-planner architecture.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md#phase-3, #phase-4
- **relations:** replacement = 027_design_and_site_planner_v2.md; site_plan tables; component library
- **verify-later:** agent_definitions for site-design-planner (likely absent); 027 live doc

<!-- SOURCE: U25_leopardess_social.md -->
### Per-site style fork chain (palette → css_theme → style_collection)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** HANDOFF §3: "Palette — forked … seed 3196d966 untouched, still dresses 3 other sites. Deployed styles.css matches the validated palette exactly" (2026-07-10/12).
- **what:** Safe per-site restyling: clone palettes + css_themes + style_collections rows (reusing seed layout/typography/header/footer), repoint sites.style_collection_id, never edit the shared seed collection that dresses multiple sites. The leopardess fork carries the two-tone gold system (A10): bright #C8A951 only on dark chrome (8.56:1), bronze #836E32 for links on light (bright gold fails AA at 2.1:1 on light). Header component forked too (header-professional-dark hardcodes navy with zero CSS variables across 4 sites) — a site_components/collection-wired fork sticks where a section fork does not.
- **sources:** docs/leopardessconsulting/scripts/L3_fork_palette.sql (header); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10, #Turn-12; docs/leopardessconsulting/RUNBOOK.md#O10
- **relations:** core-vs-specialised slot merge semantics; specialised-slot contrast gap; section resolver override behaviour
- **verify-later:** style_collections/palettes rows for leopardess; fork_theme_composition.go / install_site_composition

<!-- SOURCE: U25_leopardess_social.md -->
### Deterministic contrast gate missing on specialised palette slots
- **category:** design-composition
- **status-signal:** aspirational
- **status-evidence:** RUNNING_NOTES turn 10: "nothing stops a fork shipping an inaccessible palette — the WCAG primitives exist but aren't called at generation/fork/install/render for specialised slots."
- **what:** color_util.go has correct WCAG code (relativeLuminance, wcagContrastRatio, pickReadableOnBackground) but it is wired only to section-text defaults (loose 3.0/2.0) and forced-text-colour stripping (AA 4.5); the specialised slots (card_bg, header_bg, cta_bg/cta_text) — the exact slots that leaked white cards/navy chrome/blue CTA — are never contrast-gated. Adding the gate is small; validation is currently done by hand (all 15 reader-experienced pairs checked with the platform's own formula).
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; docs/leopardessconsulting/HANDOFF.md#8; docs/leopardessconsulting/RUNBOOK.md#O10
- **relations:** per-site style fork; styling-render-pipeline slot merge
- **verify-later:** color_util.go call sites; whether any generation/fork path calls wcagContrastRatio on specialised slots

<!-- SOURCE: U25_leopardess_social.md -->
### `analyze_design` requires palette.reference_values (else the LLM invents a palette)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 12: "the analyze_design LLM step INVENTED a dark core … Fix: restructured design_intent into palette.reference_values + prescriptive guidance … Re-rendered → all slots now exactly match."
- **what:** The webdesign-agent's analyze_design LLM reads colours only from design_intent.palette.reference_values (not color_scheme); without prescriptive values there ("these eight values are FIXED, output verbatim") it improvises from the mood text under explicit creative freedom. Same pattern applied for typography reference_values. The leopardess design_intent JSON is the worked example of the contract.
- **sources:** docs/leopardessconsulting/specs/design_intent.json#palette; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-12; docs/leopardessconsulting/HANDOFF.md#4.6
- **relations:** per-site style fork; core-vs-specialised slot merge
- **verify-later:** webdesign-agent workflow analyze_design step; render determinism with empty design_spec

<!-- SOURCE: U25_leopardess_social.md -->
### Three-per-row no-orphan grid rule as a content fix
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** L5_homepage.sql header (2026-07-10): "card grids are 3-up (no orphan row), per the brief. That is a CONTENT fix, not a CSS one — the grid components are shared across 5 sites."
- **what:** Neither global `repeat(3,1fr)` nor per-component `auto-fit,minmax()` avoids orphan/stretched last cards; the durable rule is card counts divisible by three (which also forces cutting panels that repeat each other) because grid component CSS is shared and untouchable. case-studies-grid is hard-wired to five cards and cannot be 3-up. Encoded in the design_intent layout_preference ("if the content does not divide into threes … two of the cards are saying the same thing").
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L4; docs/leopardessconsulting/scripts/L5_homepage.sql, L5_pages.sql (headers); docs/leopardessconsulting/specs/design_intent.json#layout_preference
- **relations:** shared component library semantics; anti-hype voice (repetition cut)
- **verify-later:** grid component CSS and shared usage counts

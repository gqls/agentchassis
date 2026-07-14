
<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Content quality defect catalogue (gamesdesign) and work order
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** "current as of 2026-06-09. Source of record: CATALOGUE_gamesdesign_post_sync_fix_defects(9).md"
- **what:** Five live defect classes on built pages: hero CTA text↔destination mismatch site-wide (lead item, spans content+linking); guide copy tool-flavoured; brand suffix leaking into card titles; empty footer brand/contact; empty tool descriptions. Work order: settle CTA field-vs-template, reuse component-template-fixer's CTA handling; then footer/titles/descriptions batch; then guide re-flavouring. Routing reality check flagged: the three-way finding classification and specialist agents are PROPOSED, not confirmed built.
- **sources:** FOCUS_content_quality.md (whole)
- **relations:** recommendation specialist architecture; internal linking; validate_page_content
- **verify-later:** whether identity-advisor/component-template-fixer CTA handling/sites.approval_mode exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### validate_page_content gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "the content validator and the gate … routes validate_content error_step → mark_needs_review → needs_human_review. (This is Mode 2 of the silent-completion work, already confirmed fixed.)" (2026-06-09)
- **what:** Blocker-detecting validator (placeholder text, unrendered templates, empty required sections, cross-site contamination) that any content fix must pass; failures now route consistently to human review. Known false-positive class: adopted content referencing the source domain trips the contamination heuristic (Bug 7 — needs an adopted-from whitelist for mode=recreate); legitimate emails (contactforsales.com) also flagged.
- **sources:** FOCUS_content_quality.md#machinery; HANDOFF_2026-04-23(1).md Bug 7; HANDOFF-pipeline-triage-april-2026.md#queue
- **relations:** silent completion mode 2; phantom-link check hook candidate
- **verify-later:** validate_page_content.go blocker classes; adopted-domain whitelist

<!-- SOURCE: U05_content_quality_linking.md -->
### Content-quality defect catalogue (gamesdesign.co.uk)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** FOCUS_content_quality(2) 2026-06-10 status table: defect 1 "addressed (Step 1)", defects 2–5 "open"; HANDOFF_2026-06-15 §7 lists the parked items "next package, EXPECTED to recur on readopt".
- **what:** A maintained catalogue of content-correctness defects on built pages (CATALOGUE_gamesdesign_post_sync_fix_defects as source of record): hero CTA text↔destination mismatch, tool-flavoured guide copy, brand-suffix in card titles, empty footer brand-tagline/contact, empty tool descriptions, empty meta descriptions. Content quality is explicitly scoped as "words and per-component data" — distinct from design fidelity and from link destinations. Defects are worked as separate threads, each read-the-code-first.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality(2).md; HANDOFF_2026-06-15_index_stale_rebuild(2).md#7; running_notes_17_internal_linking_phantom_fixes(21).md#content-quality-observations
- **relations:** internal-linking through-line; brand-suffix leakage; site metadata fixer; guide copy re-flavouring; readopt as acceptance test.
- **verify-later:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md; live gamesdesign.co.uk pages; site_work_items for content_rewrite items.

<!-- SOURCE: U05_content_quality_linking.md -->
### validate_page_content deploy gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** FOCUS_content_quality(2) machinery §1 "verified this round"; running_notes_15(12) Part 10 "Mode 2 … ALREADY FIXED".
- **what:** The pre-deploy content validator: placeholder/template/contamination/email checks remain blockers (error → mark_needs_review → needs_human_review), while `validateInternalLinks` (now on datahelpers) flags `phantom_link` and `empty_internal_href` as non-blocking warnings, tolerating planned-but-unbuilt pages. Known gap: it does not flag brand-suffix titles, empty contact, or empty descriptions (content/spec issues, not link/placeholder issues).
- **sources:** FOCUS_content_quality(2).md#the-machinery; FOCUS_internal_linking(1).md#shared-machinery; running_notes_15(12).md#part-10
- **relations:** phantom policy; mark_needs_review; content-quality catalogue gaps.
- **verify-later:** validate_page_content.go; page-build-handler validate_content error_step.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared-component regen clobber failure mode (silent overwrite of dependent pages)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "RESOLVED + RECOVERED (verified)" (HANDOFF(7), 2026-07-04); "R6b pass 2026-07-03: distinct md5s, needles true"; root cause section dated + confirmed in NOTES §4.
- **what:** Regenerating a shared component overwrites its `input_schema`/`html_template` field contract in place without migrating dependent pages' `content_data` to the new field names; rendering binds by exact field name and silently empties misses, so every dependent renders a content-free shell that the assembler silently drops — fanning out across every page/site sharing the component. Confirmed on `system-stats` (`fdd92ad4`, regen 2026-06-24 15:06): 24 old keys vs 22 new, five live pages on three sites byte-identical empty. `content_data` stayed intact and per-page, so the breakage was recoverable without an LLM.
- **sources:** NOTES_component_regen_clobber(43).md §1, §4, §8; HANDOFF_component_regen_clobber(7).md §Incident 1; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F1 field-contract guard (fix); F3 scoped rerender (repair path); F5/F8 (sibling facets); RenderTemplate silent-empty mechanism; visible-content filter.
- **verify-later:** platform/orchestration/actions/store_generated_component_action.go (regen branch ~L354–432); content_components/component_versions/page_components tables; component `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Recovery playbook for stranded dependents (Route A rebuild vs Route B re-key + scoped re-render)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "R5b step 8 PASS — all five RECOVERED and verified" (NOTES §9ad, 2026-07-03); leopardess confirmed live by screenshot (§9ae).
- **what:** Two recovery routes for pages stranded by a shared-contract change: Route A — full `needs_page` writer rebuild (regenerates content_data under the new schema; simplest, costs LLM); Route B — re-key each page's `content_data` old→new (explicit reviewable jsonb_build_object mapping, dry-run first, CTAS backup, non-1:1 fields handled explicitly) then trigger the F3-scoped section re-render (no LLM, preserves per-page values). Route B executed for the five, doubling as F3's end-to-end proof; gated on fleet image, freshness check, and a cta-schema decision.
- **sources:** NOTES(43).md §6, §9q, §9s–§9t, §9ad; RUNBOOK(49).md Part A; PLAN(1).md Phase 4
- **relations:** F3; section readiness model (the cta_url blocker it hit); optimistic-lock co-management; snapshot-before-change.
- **verify-later:** page_components content_data keys for the five; backup tables page_components_bak_sysstats_20260702 / _briefexp_20260703 (may be dropped).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F8 — shared-component contamination: site-specific copy baked into shared machinery (three carriers)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "STEP-12 SWEEP PASSES — contamination cleared board-wide" (NOTES §9bg, 2026-07-06) — incident remediated; "F8 mitigation… WHAT: a guard/lint so shared-component fallbacks and llm_guidance must be site-neutral" still an open Part E flag.
- **what:** A pre-guard regen (2026-07-01) baked vonc's product pitch ("Spark", the daily Gauntlet) into the shared `brief-explanation` component via three carriers invisible to the name-only F1 guard: (1) static-field fallback values; (2) those values merged into dependents' content_data by the stored⊕resolved merge; (3) per-field `llm_guidance` — the strongest, actively instructing every future writer pass on any site to write vonc's product (reproduced verbatim on robot-hands and idea.uk; contamination also migrated into generated LLM copy on pages built pre-fix — the knock-on). Remediation playbook executed: snapshot v2/v3 → neutralize fallbacks (stats→llm optional; CTAs→neutral statics) → strip merged keys with CTAS backup → scoped F3 re-renders → writer rebuilds under cleaned guidance → board-wide strpos sweep (clean except vonc's own legitimate copy). Falsified along the way: field-description carrier, content_brief column, restore-v1 option (old-architecture contract). Proposed structural mitigation (unbuilt): store-time site-neutrality lint over fallbacks + llm_guidance.
- **sources:** NOTES(43).md §9an–§9bb, §9bg; RUNBOOK(49).md Part C + Part E F8; HANDOFF(7).md §Incident 2
- **relations:** F1 guard (its blind spot); llm_guidance surface; stored⊕resolved merge; neutralize-in-place remediation; D2b lint (same detection-net shape).
- **verify-later:** brief-explanation input_schema (neutral guidance ×11, stats source=llm no fallback); component_versions v1–v3; store-side lint absence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Neutralize-in-place remediation pattern for contaminated shared components
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "v1 not a restore candidate… NEUTRALIZE-IN-PLACE chosen" (NOTES §9ao); Steps 1–2 landed with optimistic lock held (§9ap); Steps 6–7 landed §9bd.
- **what:** When a shared component's history offers no clean restore point (v1 predated the current field contract; restoring would regress dependents on the new architecture), the fix is surgical in-place neutralization: manual snapshot first, then targeted jsonb patches replacing only the offending attributes (fallbacks, guidance) under an optimistic lock, preserving names/types/structure — followed by per-dependent cleanup (strip merged keys, scoped re-renders, writer rebuilds) mapped per consumer (vonc's own copy untouched; robot-hands stripped; old-architecture pages escalated to rebuilds).
- **sources:** NOTES(43).md §9ao–§9aq, §9bb, §9bd; RUNBOOK(49).md Part C Steps 1–9
- **relations:** F8; optimistic-lock co-management; component versioning; recovery playbook.
- **verify-later:** the CTE jsonb_object_agg patch shape in RUNBOOK(49) Step 7 as reusable SQL.

<!-- SOURCE: U09_adoption.md -->
### Adoption content-quality defect families (polish batch)
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** Open Groups 3 items as of HANDOFF_2026-06-09: "- GameDesign.uk brand-suffix in card titles; footer mailto/tagline empty; one empty tool description; guide tables render poorly; guides should cross-link to tools"; hero H1 reuse and empty meta descriptions from the catalogue remain untracked-as-fixed.
- **what:** The residual content-quality class after build mechanics were fixed: source-brand `<title>` suffixes used as display names (preserving the source brand, not the destination), empty footer contact/tagline (no graceful no-data path — components render empty structure instead of hiding), hero H1 duplicated across hubs, meta_description populated in DB but emitted empty, tool-flavoured guide copy (user open to real embedded interactive demos in guides), guide→tool cross-linking as enhancement.
- **sources:** CATALOGUE(9)#family-e, HANDOFF_2026-06-09#next-task, running_notes_14(25)#part-14n
- **relations:** silent-fallback link family; page-content-writer prompts; internal linking (024)
- **verify-later:** current gamesdesign deployed HTML; hero/footer component schemas' no-data paths

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-build validation of structured components (Fix D, unimplemented)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "Post-build validation (Fix D): assert a component whose input_schema declares a required structured field ... actually has it populated before deploy" — listed under open/structural, not done
- **what:** Proposed check that runs after a build, asserting that any component whose `input_schema` declares a required structured field actually has that field populated in `content_data`; if empty, flags the page instead of deploying a silently-empty component. Catches the bug class regardless of which planner or writer path produced the empty result.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#D-Post-build-validation, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** grep/inspect `input_schema`; `content_data`

<!-- SOURCE: U15_docs019_running_notes.md -->
### Site-quality programme handoff
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "site-quality programme HANDED OFF to its own runbook... 0 nav / 0 img / 0 svg / 0 script on ALL pages" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** Following the platform's first recorded domain→deployed-site milestone (dartsonline.com), a measured baseline (four rendered pages, all missing nav/images/svg/script, thin CSS variable usage, near-zero internal links) triggered a dedicated handoff (`RUNBOOK_site_quality.md`) splitting remaining work into stuck-dispatch (chrome/design/imagery), delivered-but-poor (content depth, links), and never-in-scope (feeds/RSS/graphics/games, disabled improvement-sweep) — with a live hypothesis that the relay path lacks the monolith's `render_site_components` chrome step, explaining nav-zero across every page.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-06 "MILESTONE recorded" and "site-quality programme HANDED OFF" entries.
- **relations:** Work-item relay / builder-generations architecture; diagnosis→fix loop workstream founding (the same "unresolved_cta" defect class recurs across both threads).

<!-- SOURCE: U19_sql_tables_components.md -->
### Placeholder-content suppression sweep
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** Executed SQL in 018: find deployed sections containing 'NEEDS HUMAN REVIEW'/'Lorem ipsum'/'[INSERT'/'<no value>', replace with hidden comment, create per-page placeholder_content items (handler 'human-review', status needs_human_review) plus per-site needs_rerender items.
- **what:** A validation pattern: placeholder or unreviewed text must never stay live — offending sections are hidden behind an HTML comment, a needs_human_review work item requests the real data (team names, photos...), and a rerender item republishes. Companion flows later resolve needs_section_data items as wont_fix when data arrives via site_specs (team, departments) or the section is dropped (pricing → engagement process).
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#placeholder-sweep and #075b-075e
- **relations:** work-item queue; site_specs identity enrichment; hitl approval.
- **verify-later:** validation agent producing these; recurrence of placeholder text.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Audited content pipeline (persona → research → draft → veracity/copyright audits)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** "Content Pipeline cannot be a single agent… Step 4 (Audit - Veracity)… Step 5 (Audit - Copyright)" (001); "Purifier Agent" and "Copywriters with Character" in the phase summary (014); site_persona step defined in 011.
- **what:** Content generation as an orchestrated sub-system: define a site persona/style guide, research via search/scrape adapters, persona-driven drafting, fact-check against research (separate agent, possible HITL), plagiarism/copyright audit (images only from licensed/free sources), then inject into template slots found by parsing data-function attributes. Motivated by veracity/copyright being "mission-critical legal and reputational risks".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#content-bottleneck; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** live successors: content-quality docs (content_quality_and_internal_linking), research-agents; persona idea → persona architecture across the platform.
- **verify-later:** whether any veracity/plagiarism audit step exists in the current content pipeline.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Content validation before review (validate_page_content)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** docs018/003: "validate_page_content action runs before review mode determination... Validation errors force HITL review, blocking auto-approval."
- **what:** Deterministic pre-review checks on generated pages: extract all hrefs and verify internal links against the pages table, verify emails against site contact data; errors (broken links) force human review while warnings flow through. Companion mechanisms: prepare_link_context injects an only-link-to-these-pages allowlist into writer prompts, and rerender-time contact injection replaces hallucinated phone/email with DB truth.
- **sources:** docs018_rerendering/003_website_builder_architecture_status_report.md#3; docs018_rerendering/002_summary_link_constraints.md; docs018_rerendering/003_website_builder_architecture_status_report.md#6
- **relations:** content-reviewer workflow; link_registry; content-quality internal linking (successor).
- **verify-later:** validate_page_content + prepare_link_context in registry; prompt inclusion of link_constraint_text ("Not Yet Done" at the time).

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Prompt composition asymmetry (text cascade vs image)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** loop_closure(9) decisions: "Step 5 image prompt cascade — Defer. Keep the single-prepend `imagery_direction` cascade for v1"; references `FOCUS_prompt_composition_pattern.md` — "the text pattern itself is fragile and shouldn't be copied wholesale … a composer step that produces a parameter envelope (for both text and images) is the strongest candidate."
- **what:** Deliberate design opinion that image prompts use only a single-prepend `imagery_direction` cascade, not the richer page-content-writer text composition — because the text cascade is considered fragile and a better target is a unified composer producing a parameter envelope for both text and images (likely landing in 2H, not a step-5 extension).
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#decisions-taken, #image-prompt-cascade-deferred
- **relations:** live FOCUS_prompt_composition_pattern.md; image request shape (2H); directive cascade
- **verify-later:** composeImagePromptWithDirection; getImageryDirectionForSite

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Sanitisation v2 … now strips Cc AND Cf … Real bug found by the new tests: checking IsControl before IsSpace silently JOINED words".
- **what:** The engine's `sanitizeValue()` strips control (Cc) and format (Cf: zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen) chars, collapses whitespace runs (IsSpace checked FIRST to avoid joining words like `gmt\t\tmaster`→`gmtmaster`), and caps by RUNES not bytes (MaxValueLen semantic changed). NFD combining marks deliberately survive; NFC normalisation + lowercasing are deferred to the P4 collector (needs x/text; engine stays stdlib-only).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_plan(11).md#p4
- **relations:** pairs with P4 ingest validation contract (NFC there)
- **verify-later:** service.go sanitizeValue; MAX_VALUE_LEN handling

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `component-template-fixer` CTA-handling reuse assumption — corrected, replaced by dedicated agent
- **category:** content-quality
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09): "The plan notes `component-template-fixer` 'already handles CTA fixes' — verify and extend rather than build new." Live `FOCUS_content_quality(2).md` (2026-06-10): "`component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:'needs_review'`)... So the PLAN's 'already handles CTA fixes' was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3)."
- **what:** `PLAN_design-note-recommendation-specialists.md` asserted `component-template-fixer` already had CTA-fix handling that could be reused/extended for the hero-CTA phantom-link defect. Verification against the live agent's actual routing table found it explicitly declines CTA improvements, routing them to `needs_review` instead of fixing them. This wrong assumption, once corrected, directly motivated building a new dedicated agent (`internal-link-resolver`, see below) rather than extending the wrong one.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived); live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** internal-link-resolver agent (below); identity-advisor/sites.approval_mode (below)
- **verify-later:** `component-template-fixer`'s current action set.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `identity-advisor` agent and `sites.approval_mode` gate — proposed, confirmed never built
- **category:** content-quality
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09) lists them as PROPOSED pieces needing verification ("Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists"). Live `FOCUS_content_quality(2).md`/`FOCUS_internal_linking(1).md` (2026-06-10) confirm: "`identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built."
- **what:** `PLAN_design-note-recommendation-specialists.md`'s three-way finding-routing design (bug / gap / recommendation) named `identity-advisor` as the specialist for contact/email findings and `sites.approval_mode` as the gate for whether recommendation-type findings auto-apply. Neither was ever implemented — a clean case of a documented plan whose specific pieces were checked against the live schema/agent_definitions and found absent.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived) and live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** component-template-fixer CTA-reuse assumption (above)
- **verify-later:** re-check `agent_definitions` and the `sites` table for these names in case they were built later.

<!-- SOURCE: U25_leopardess_social.md -->
### LLM fabrication classes in self-built site content
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** AUDIT U1–U11 with removal dates ("Section DELETED 2026-07-10", "FAQ replaced 2026-07-10"); "Fabrication sweep, 2026-07-10 … CLEAN".
- **what:** Catalogue of what unconstrained content agents invented on a live site: fictional staff ("Peter Grenfell"), a nonexistent "8 departments" taxonomy, platform subsystems dressed as client case studies, AI agents listed as human team members with 404 portraits, capabilities that don't exist (Playwright scraping, proxy pools, circuit breakers, Helm/IAM), and misaligned stat suffixes ("99.9x uptime"). Removal required both spec rewrites and component deletion because some copy was baked into rendered_html.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5, #Turn-11; docs/leopardessconsulting/scripts/L5_pages.sql (header)
- **relations:** claim-evidence audit rule; site_specs pinned gap (fabrications regenerate while specs are wrong)
- **verify-later:** page_components.content_data pattern sweep on all sites; content-gap-planner rewrite history in site_specs

<!-- SOURCE: U25_leopardess_social.md -->
### Anti-hype voice and claim-discipline spec
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** HANDOFF §3 "Specs — identity/voice/design_intent/portfolio rewritten (source_agent operator-rebuild, pinned)".
- **what:** A reusable voice contract for LLM-written site copy: positive framing (no strawmen, no competitor swipes), prefer the smaller exactly-true claim, plain language over compression, a banned-language list ("digital transformation", "leverage", "seamless"…), an LLM-tells-to-avoid list (reflexive triads, em-dash rhythm, summarising flourishes), CTA governance (name the next thing that happens; vary per page — repetition "signals content shallowness"), and honest uncertainty framing ("we have not done that one yet").
- **sources:** docs/leopardessconsulting/specs/voice.json; docs/leopardessconsulting/specs/identity.json#content_posture; docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header)
- **relations:** LLM fabrication classes; portfolio honest-labelling pattern ("Not yet done for a client")
- **verify-later:** site_specs aspect 'voice' for leopardess; whether content writers consume voice spec

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Content quality defect catalogue (gamesdesign) and work order
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** "current as of 2026-06-09. Source of record: CATALOGUE_gamesdesign_post_sync_fix_defects(9).md"
- **what:** Five live defect classes on built pages: hero CTA text↔destination mismatch site-wide (lead item, spans content+linking); guide copy tool-flavoured; brand suffix leaking into card titles; empty footer brand/contact; empty tool descriptions. Work order: settle CTA field-vs-template, reuse component-template-fixer's CTA handling; then footer/titles/descriptions batch; then guide re-flavouring. Routing reality check flagged: the three-way finding classification and specialist agents are PROPOSED, not confirmed built.
- **sources:** FOCUS_content_quality.md (whole)
- **relations:** recommendation specialist architecture; internal linking; validate_page_content
- **verify-later:** whether identity-advisor/component-template-fixer CTA handling/sites.approval_mode exist

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### validate_page_content gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "the content validator and the gate … routes validate_content error_step → mark_needs_review → needs_human_review. (This is Mode 2 of the silent-completion work, already confirmed fixed.)" (2026-06-09)
- **what:** Blocker-detecting validator (placeholder text, unrendered templates, empty required sections, cross-site contamination) that any content fix must pass; failures now route consistently to human review. Known false-positive class: adopted content referencing the source domain trips the contamination heuristic (Bug 7 — needs an adopted-from whitelist for mode=recreate); legitimate emails (contactforsales.com) also flagged.
- **sources:** FOCUS_content_quality.md#machinery; HANDOFF_2026-04-23(1).md Bug 7; HANDOFF-pipeline-triage-april-2026.md#queue
- **relations:** silent completion mode 2; phantom-link check hook candidate
- **verify-later:** validate_page_content.go blocker classes; adopted-domain whitelist

<!-- SOURCE: U05_content_quality_linking.md -->
### Content-quality defect catalogue (gamesdesign.co.uk)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** FOCUS_content_quality(2) 2026-06-10 status table: defect 1 "addressed (Step 1)", defects 2–5 "open"; HANDOFF_2026-06-15 §7 lists the parked items "next package, EXPECTED to recur on readopt".
- **what:** A maintained catalogue of content-correctness defects on built pages (CATALOGUE_gamesdesign_post_sync_fix_defects as source of record): hero CTA text↔destination mismatch, tool-flavoured guide copy, brand-suffix in card titles, empty footer brand-tagline/contact, empty tool descriptions, empty meta descriptions. Content quality is explicitly scoped as "words and per-component data" — distinct from design fidelity and from link destinations. Defects are worked as separate threads, each read-the-code-first.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality(2).md; HANDOFF_2026-06-15_index_stale_rebuild(2).md#7; running_notes_17_internal_linking_phantom_fixes(21).md#content-quality-observations
- **relations:** internal-linking through-line; brand-suffix leakage; site metadata fixer; guide copy re-flavouring; readopt as acceptance test.
- **verify-later:** CATALOGUE_gamesdesign_post_sync_fix_defects(9).md; live gamesdesign.co.uk pages; site_work_items for content_rewrite items.

<!-- SOURCE: U05_content_quality_linking.md -->
### validate_page_content deploy gate
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** FOCUS_content_quality(2) machinery §1 "verified this round"; running_notes_15(12) Part 10 "Mode 2 … ALREADY FIXED".
- **what:** The pre-deploy content validator: placeholder/template/contamination/email checks remain blockers (error → mark_needs_review → needs_human_review), while `validateInternalLinks` (now on datahelpers) flags `phantom_link` and `empty_internal_href` as non-blocking warnings, tolerating planned-but-unbuilt pages. Known gap: it does not flag brand-suffix titles, empty contact, or empty descriptions (content/spec issues, not link/placeholder issues).
- **sources:** FOCUS_content_quality(2).md#the-machinery; FOCUS_internal_linking(1).md#shared-machinery; running_notes_15(12).md#part-10
- **relations:** phantom policy; mark_needs_review; content-quality catalogue gaps.
- **verify-later:** validate_page_content.go; page-build-handler validate_content error_step.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared-component regen clobber failure mode (silent overwrite of dependent pages)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "RESOLVED + RECOVERED (verified)" (HANDOFF(7), 2026-07-04); "R6b pass 2026-07-03: distinct md5s, needles true"; root cause section dated + confirmed in NOTES §4.
- **what:** Regenerating a shared component overwrites its `input_schema`/`html_template` field contract in place without migrating dependent pages' `content_data` to the new field names; rendering binds by exact field name and silently empties misses, so every dependent renders a content-free shell that the assembler silently drops — fanning out across every page/site sharing the component. Confirmed on `system-stats` (`fdd92ad4`, regen 2026-06-24 15:06): 24 old keys vs 22 new, five live pages on three sites byte-identical empty. `content_data` stayed intact and per-page, so the breakage was recoverable without an LLM.
- **sources:** NOTES_component_regen_clobber(43).md §1, §4, §8; HANDOFF_component_regen_clobber(7).md §Incident 1; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F1 field-contract guard (fix); F3 scoped rerender (repair path); F5/F8 (sibling facets); RenderTemplate silent-empty mechanism; visible-content filter.
- **verify-later:** platform/orchestration/actions/store_generated_component_action.go (regen branch ~L354–432); content_components/component_versions/page_components tables; component `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Recovery playbook for stranded dependents (Route A rebuild vs Route B re-key + scoped re-render)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "R5b step 8 PASS — all five RECOVERED and verified" (NOTES §9ad, 2026-07-03); leopardess confirmed live by screenshot (§9ae).
- **what:** Two recovery routes for pages stranded by a shared-contract change: Route A — full `needs_page` writer rebuild (regenerates content_data under the new schema; simplest, costs LLM); Route B — re-key each page's `content_data` old→new (explicit reviewable jsonb_build_object mapping, dry-run first, CTAS backup, non-1:1 fields handled explicitly) then trigger the F3-scoped section re-render (no LLM, preserves per-page values). Route B executed for the five, doubling as F3's end-to-end proof; gated on fleet image, freshness check, and a cta-schema decision.
- **sources:** NOTES(43).md §6, §9q, §9s–§9t, §9ad; RUNBOOK(49).md Part A; PLAN(1).md Phase 4
- **relations:** F3; section readiness model (the cta_url blocker it hit); optimistic-lock co-management; snapshot-before-change.
- **verify-later:** page_components content_data keys for the five; backup tables page_components_bak_sysstats_20260702 / _briefexp_20260703 (may be dropped).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F8 — shared-component contamination: site-specific copy baked into shared machinery (three carriers)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "STEP-12 SWEEP PASSES — contamination cleared board-wide" (NOTES §9bg, 2026-07-06) — incident remediated; "F8 mitigation… WHAT: a guard/lint so shared-component fallbacks and llm_guidance must be site-neutral" still an open Part E flag.
- **what:** A pre-guard regen (2026-07-01) baked vonc's product pitch ("Spark", the daily Gauntlet) into the shared `brief-explanation` component via three carriers invisible to the name-only F1 guard: (1) static-field fallback values; (2) those values merged into dependents' content_data by the stored⊕resolved merge; (3) per-field `llm_guidance` — the strongest, actively instructing every future writer pass on any site to write vonc's product (reproduced verbatim on robot-hands and idea.uk; contamination also migrated into generated LLM copy on pages built pre-fix — the knock-on). Remediation playbook executed: snapshot v2/v3 → neutralize fallbacks (stats→llm optional; CTAs→neutral statics) → strip merged keys with CTAS backup → scoped F3 re-renders → writer rebuilds under cleaned guidance → board-wide strpos sweep (clean except vonc's own legitimate copy). Falsified along the way: field-description carrier, content_brief column, restore-v1 option (old-architecture contract). Proposed structural mitigation (unbuilt): store-time site-neutrality lint over fallbacks + llm_guidance.
- **sources:** NOTES(43).md §9an–§9bb, §9bg; RUNBOOK(49).md Part C + Part E F8; HANDOFF(7).md §Incident 2
- **relations:** F1 guard (its blind spot); llm_guidance surface; stored⊕resolved merge; neutralize-in-place remediation; D2b lint (same detection-net shape).
- **verify-later:** brief-explanation input_schema (neutral guidance ×11, stats source=llm no fallback); component_versions v1–v3; store-side lint absence.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Neutralize-in-place remediation pattern for contaminated shared components
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "v1 not a restore candidate… NEUTRALIZE-IN-PLACE chosen" (NOTES §9ao); Steps 1–2 landed with optimistic lock held (§9ap); Steps 6–7 landed §9bd.
- **what:** When a shared component's history offers no clean restore point (v1 predated the current field contract; restoring would regress dependents on the new architecture), the fix is surgical in-place neutralization: manual snapshot first, then targeted jsonb patches replacing only the offending attributes (fallbacks, guidance) under an optimistic lock, preserving names/types/structure — followed by per-dependent cleanup (strip merged keys, scoped re-renders, writer rebuilds) mapped per consumer (vonc's own copy untouched; robot-hands stripped; old-architecture pages escalated to rebuilds).
- **sources:** NOTES(43).md §9ao–§9aq, §9bb, §9bd; RUNBOOK(49).md Part C Steps 1–9
- **relations:** F8; optimistic-lock co-management; component versioning; recovery playbook.
- **verify-later:** the CTE jsonb_object_agg patch shape in RUNBOOK(49) Step 7 as reusable SQL.

<!-- SOURCE: U09_adoption.md -->
### Adoption content-quality defect families (polish batch)
- **category:** content-quality
- **status-signal:** unknown
- **status-evidence:** Open Groups 3 items as of HANDOFF_2026-06-09: "- GameDesign.uk brand-suffix in card titles; footer mailto/tagline empty; one empty tool description; guide tables render poorly; guides should cross-link to tools"; hero H1 reuse and empty meta descriptions from the catalogue remain untracked-as-fixed.
- **what:** The residual content-quality class after build mechanics were fixed: source-brand `<title>` suffixes used as display names (preserving the source brand, not the destination), empty footer contact/tagline (no graceful no-data path — components render empty structure instead of hiding), hero H1 duplicated across hubs, meta_description populated in DB but emitted empty, tool-flavoured guide copy (user open to real embedded interactive demos in guides), guide→tool cross-linking as enhancement.
- **sources:** CATALOGUE(9)#family-e, HANDOFF_2026-06-09#next-task, running_notes_14(25)#part-14n
- **relations:** silent-fallback link family; page-content-writer prompts; internal linking (024)
- **verify-later:** current gamesdesign deployed HTML; hero/footer component schemas' no-data paths

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-build validation of structured components (Fix D, unimplemented)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "Post-build validation (Fix D): assert a component whose input_schema declares a required structured field ... actually has it populated before deploy" — listed under open/structural, not done
- **what:** Proposed check that runs after a build, asserting that any component whose `input_schema` declares a required structured field actually has that field populated in `content_data`; if empty, flags the page instead of deploying a silently-empty component. Catches the bug class regardless of which planner or writer path produced the empty result.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#D-Post-build-validation, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** grep/inspect `input_schema`; `content_data`

<!-- SOURCE: U15_docs019_running_notes.md -->
### Site-quality programme handoff
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "site-quality programme HANDED OFF to its own runbook... 0 nav / 0 img / 0 svg / 0 script on ALL pages" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** Following the platform's first recorded domain→deployed-site milestone (dartsonline.com), a measured baseline (four rendered pages, all missing nav/images/svg/script, thin CSS variable usage, near-zero internal links) triggered a dedicated handoff (`RUNBOOK_site_quality.md`) splitting remaining work into stuck-dispatch (chrome/design/imagery), delivered-but-poor (content depth, links), and never-in-scope (feeds/RSS/graphics/games, disabled improvement-sweep) — with a live hypothesis that the relay path lacks the monolith's `render_site_components` chrome step, explaining nav-zero across every page.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-06 "MILESTONE recorded" and "site-quality programme HANDED OFF" entries.
- **relations:** Work-item relay / builder-generations architecture; diagnosis→fix loop workstream founding (the same "unresolved_cta" defect class recurs across both threads).

<!-- SOURCE: U19_sql_tables_components.md -->
### Placeholder-content suppression sweep
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** Executed SQL in 018: find deployed sections containing 'NEEDS HUMAN REVIEW'/'Lorem ipsum'/'[INSERT'/'<no value>', replace with hidden comment, create per-page placeholder_content items (handler 'human-review', status needs_human_review) plus per-site needs_rerender items.
- **what:** A validation pattern: placeholder or unreviewed text must never stay live — offending sections are hidden behind an HTML comment, a needs_human_review work item requests the real data (team names, photos...), and a rerender item republishes. Companion flows later resolve needs_section_data items as wont_fix when data arrives via site_specs (team, departments) or the section is dropped (pricing → engagement process).
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#placeholder-sweep and #075b-075e
- **relations:** work-item queue; site_specs identity enrichment; hitl approval.
- **verify-later:** validation agent producing these; recurrence of placeholder text.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Audited content pipeline (persona → research → draft → veracity/copyright audits)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** "Content Pipeline cannot be a single agent… Step 4 (Audit - Veracity)… Step 5 (Audit - Copyright)" (001); "Purifier Agent" and "Copywriters with Character" in the phase summary (014); site_persona step defined in 011.
- **what:** Content generation as an orchestrated sub-system: define a site persona/style guide, research via search/scrape adapters, persona-driven drafting, fact-check against research (separate agent, possible HITL), plagiarism/copyright audit (images only from licensed/free sources), then inject into template slots found by parsing data-function attributes. Motivated by veracity/copyright being "mission-critical legal and reputational risks".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#content-bottleneck; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** live successors: content-quality docs (content_quality_and_internal_linking), research-agents; persona idea → persona architecture across the platform.
- **verify-later:** whether any veracity/plagiarism audit step exists in the current content pipeline.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Content validation before review (validate_page_content)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** docs018/003: "validate_page_content action runs before review mode determination... Validation errors force HITL review, blocking auto-approval."
- **what:** Deterministic pre-review checks on generated pages: extract all hrefs and verify internal links against the pages table, verify emails against site contact data; errors (broken links) force human review while warnings flow through. Companion mechanisms: prepare_link_context injects an only-link-to-these-pages allowlist into writer prompts, and rerender-time contact injection replaces hallucinated phone/email with DB truth.
- **sources:** docs018_rerendering/003_website_builder_architecture_status_report.md#3; docs018_rerendering/002_summary_link_constraints.md; docs018_rerendering/003_website_builder_architecture_status_report.md#6
- **relations:** content-reviewer workflow; link_registry; content-quality internal linking (successor).
- **verify-later:** validate_page_content + prepare_link_context in registry; prompt inclusion of link_constraint_text ("Not Yet Done" at the time).

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Prompt composition asymmetry (text cascade vs image)
- **category:** content-quality
- **status-signal:** aspirational
- **status-evidence:** loop_closure(9) decisions: "Step 5 image prompt cascade — Defer. Keep the single-prepend `imagery_direction` cascade for v1"; references `FOCUS_prompt_composition_pattern.md` — "the text pattern itself is fragile and shouldn't be copied wholesale … a composer step that produces a parameter envelope (for both text and images) is the strongest candidate."
- **what:** Deliberate design opinion that image prompts use only a single-prepend `imagery_direction` cascade, not the richer page-content-writer text composition — because the text cascade is considered fragile and a better target is a unified composer producing a parameter envelope for both text and images (likely landing in 2H, not a step-5 extension).
- **sources:** imagery/old/PLAN_imagery_loop_closure(9).md#decisions-taken, #image-prompt-cascade-deferred
- **relations:** live FOCUS_prompt_composition_pattern.md; image request shape (2H); directive cascade
- **verify-later:** composeImagePromptWithDirection; getImageryDirectionForSite

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Sanitisation v2 … now strips Cc AND Cf … Real bug found by the new tests: checking IsControl before IsSpace silently JOINED words".
- **what:** The engine's `sanitizeValue()` strips control (Cc) and format (Cf: zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen) chars, collapses whitespace runs (IsSpace checked FIRST to avoid joining words like `gmt\t\tmaster`→`gmtmaster`), and caps by RUNES not bytes (MaxValueLen semantic changed). NFD combining marks deliberately survive; NFC normalisation + lowercasing are deferred to the P4 collector (needs x/text; engine stays stdlib-only).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_plan(11).md#p4
- **relations:** pairs with P4 ingest validation contract (NFC there)
- **verify-later:** service.go sanitizeValue; MAX_VALUE_LEN handling

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `component-template-fixer` CTA-handling reuse assumption — corrected, replaced by dedicated agent
- **category:** content-quality
- **status-signal:** superseded
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09): "The plan notes `component-template-fixer` 'already handles CTA fixes' — verify and extend rather than build new." Live `FOCUS_content_quality(2).md` (2026-06-10): "`component-template-fixer` exists but **explicitly punts on CTAs** (`cta_improvement`/`cta` → `fixed:false, action:'needs_review'`)... So the PLAN's 'already handles CTA fixes' was wrong; there was no CTA resolver to reuse — hence the dedicated `internal-link-resolver` (Step 3)."
- **what:** `PLAN_design-note-recommendation-specialists.md` asserted `component-template-fixer` already had CTA-fix handling that could be reused/extended for the hero-CTA phantom-link defect. Verification against the live agent's actual routing table found it explicitly declines CTA improvements, routing them to `needs_review` instead of fixing them. This wrong assumption, once corrected, directly motivated building a new dedicated agent (`internal-link-resolver`, see below) rather than extending the wrong one.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived); live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** internal-link-resolver agent (below); identity-advisor/sites.approval_mode (below)
- **verify-later:** `component-template-fixer`'s current action set.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `identity-advisor` agent and `sites.approval_mode` gate — proposed, confirmed never built
- **category:** content-quality
- **status-signal:** abandoned
- **status-evidence:** Archived `FOCUS_content_quality.md` (2026-06-09) lists them as PROPOSED pieces needing verification ("Before relying on `identity-advisor` / `component-template-fixer` / `sites.approval_mode`, confirm each exists"). Live `FOCUS_content_quality(2).md`/`FOCUS_internal_linking(1).md` (2026-06-10) confirm: "`identity-advisor` does **not** exist. `sites.approval_mode` does **not** exist. The three-way `finding_type` classification and those specialists are PROPOSED, not built."
- **what:** `PLAN_design-note-recommendation-specialists.md`'s three-way finding-routing design (bug / gap / recommendation) named `identity-advisor` as the specialist for contact/email findings and `sites.approval_mode` as the gate for whether recommendation-type findings auto-apply. Neither was ever implemented — a clean case of a documented plan whose specific pieces were checked against the live schema/agent_definitions and found absent.
- **sources:** content_quality_and_internal_linking/FOCUS_content_quality.md (archived) and live FOCUS_content_quality(2).md; running_notes_17(16)
- **relations:** component-template-fixer CTA-reuse assumption (above)
- **verify-later:** re-check `agent_definitions` and the `sites` table for these names in case they were built later.

<!-- SOURCE: U25_leopardess_social.md -->
### LLM fabrication classes in self-built site content
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** AUDIT U1–U11 with removal dates ("Section DELETED 2026-07-10", "FAQ replaced 2026-07-10"); "Fabrication sweep, 2026-07-10 … CLEAN".
- **what:** Catalogue of what unconstrained content agents invented on a live site: fictional staff ("Peter Grenfell"), a nonexistent "8 departments" taxonomy, platform subsystems dressed as client case studies, AI agents listed as human team members with 404 portraits, capabilities that don't exist (Playwright scraping, proxy pools, circuit breakers, Helm/IAM), and misaligned stat suffixes ("99.9x uptime"). Removal required both spec rewrites and component deletion because some copy was baked into rendered_html.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#2; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-5, #Turn-11; docs/leopardessconsulting/scripts/L5_pages.sql (header)
- **relations:** claim-evidence audit rule; site_specs pinned gap (fabrications regenerate while specs are wrong)
- **verify-later:** page_components.content_data pattern sweep on all sites; content-gap-planner rewrite history in site_specs

<!-- SOURCE: U25_leopardess_social.md -->
### Anti-hype voice and claim-discipline spec
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** HANDOFF §3 "Specs — identity/voice/design_intent/portfolio rewritten (source_agent operator-rebuild, pinned)".
- **what:** A reusable voice contract for LLM-written site copy: positive framing (no strawmen, no competitor swipes), prefer the smaller exactly-true claim, plain language over compression, a banned-language list ("digital transformation", "leverage", "seamless"…), an LLM-tells-to-avoid list (reflexive triads, em-dash rhythm, summarising flourishes), CTA governance (name the next thing that happens; vary per page — repetition "signals content shallowness"), and honest uncertainty framing ("we have not done that one yet").
- **sources:** docs/leopardessconsulting/specs/voice.json; docs/leopardessconsulting/specs/identity.json#content_posture; docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql (header)
- **relations:** LLM fabrication classes; portfolio honest-labelling pattern ("Not yet done for a client")
- **verify-later:** site_specs aspect 'voice' for leopardess; whether content writers consume voice spec

# Register — adoption-pipeline

36 concepts, consolidated from 106 raw extractions (the cluster input file contains
the entire adoption-pipeline block set duplicated exactly twice — 53 unique raw
blocks appearing twice each — plus further cross-unit duplication of the same
concepts under different titles) across units U01, U02, U04, U05, U09, U12, U13,
U14, U15, U16, U17a, U18, U20, U21, U24d, U24f.

### ADO-001 — Infrastructure three layers (core platform / client delivery / framework builder)
- **status:** partial
- **status-evidence:** 007/007v3: "Layer 1 — Adoption pipeline (current)"; Layer 2 client delivery beyond static S3 serving is "planned"; Layer 3 framework provisioning is "future".
- **what:** The platform runs in three layers: Layer 1 (K8s factory) never serves external traffic — it only produces artefacts and pushes them; Layer 2 is client delivery (S3 static now, site-api-router + client Postgres on OVH VMs planned, config-driven routes reusing the action library); Layer 3 is provisioning agent frameworks for clients. Five backend capability tiers run static-first (vetcomparison's JSON-on-S3 pattern handles ~50k items, Pagefind extends to ~500k).
- **sources:** 007#Infrastructure Separation, #Backend Capability Tiers, #Site API Router; WM/007_adoption_pipeline_v3.md#infrastructure-separation, #backend-capability-tiers
- **relations:** dynamic-applications direction (DYN-001, shared tier language); site-api-router
- **verify-later:** any site-api-router code; vetcomparison export path

### ADO-002 — Adoption is a one-off capture, not a ceiling (specs separation)
- **status:** deployed
- **status-evidence:** 007 v4 principles + what-gets-stored table; patch updates to Phase-1 split.
- **what:** Crawl data lands in research_results (adoption_crawl/adoption_page), never in site_specs. Spec aspects split into identity (with adopted_from provenance), design_reference (concrete extracted values, historical, never modified), design_intent (semantic brand-level brief, survives plan rebuilds), site_archetype (character + inviolable constraints), content_direction (brand voice), and structure. Webdesign reads intent not reference; evolution means updating intent, and the strategist then writes aspiration beyond the adopted baseline.
- **sources:** 007#Site Adoption, #What gets stored where; 004#Adopted Sites
- **relations:** site_specs aspect-versioned store (site-spec-and-classifier SPEC-002); strategic-vs-plan-time split (doc 030)
- **verify-later:** site_specs aspects on an adopted site

### ADO-003 — Site-adoption pipeline: wrapper orchestrator, Go design-fingerprint, LLM classify, apply_adoption_plan
- **status:** deployed
- **status-evidence:** 007's 16-step workflow with runtime expectations; repeated verified gamesdesign adoption runs through 2026-06-05; "the proven template" (FOCUS_interactive_content_generation).
- **what:** A thin `site-adoption-orchestrator` (spawn→call→complete) spawns `site-adoption-agent` as its own K8s Job. The agent crawls via Firecrawl (markdown + rawHtml, limit 30), extracts a design fingerprint Go-side with goquery (hex colours, fonts, CSS vars, layout patterns, Google Fonts; external CSS fetched via firecrawl_scrape and merged by EnrichFingerprintWithCSSAction), then an LLM stage runs analyze_site (per-page page_type + url), classify_archetype, derive_content_direction, and generate_design_intent, and finally Go's `apply_adoption_plan` writes page records, per-page markdown to research_results, the design_reference spec, and work items. Governing principle: LLM for reasoning, Go for extraction — never pay an LLM to transcribe hex values a regex can read.
- **sources:** 007#The adoption agent, #Three-stage processing, #Design Fingerprint Pipeline; FOCUS_component_schema_patterns.md#appendix; WM/007_adoption_pipeline_v3.md#the-adoption-agent; 091_site_adoption_agent.sql; 104_site_adoption_orchestrator.sql
- **relations:** wrapper-orchestrator pattern; classifier handoff (ADO-006); design_reference/design_intent split superseding the old unified `design` aspect (ADO-019); interactive-content parse-stage gap (ADO-016, DYN-002)
- **verify-later:** extract_design_fingerprint_action.go; apply_adoption_plan_action.go; agent_definitions rows for site-adoption-agent (4e2d8e8e…) and site-adoption-orchestrator

### ADO-004 — Source vs destination separation (target_url / destination_domain)
- **status:** deployed
- **status-evidence:** "Phase 1 plumbing (target_url/destination_domain separation) is deployed... source/destination separation deployed and working" (FOCUS_adoption_fidelity_and_variants).
- **what:** Adoption separates the crawled reference site from the site being built: `target_url` is what gets crawled, `destination_domain` is what gets built (legacy single url/domain shape still accepted). `ensure_site_record` uses the destination via a domain-override field; provenance records the source host, which also keys the crawl-content lookup — a source/destination mismatch here silently drops all adopted content, which is what the `sourceDomain` vs `domain` fix in `apply_adoption_plan_action.go` addressed.
- **sources:** 007#Source vs destination, #Adoption modes; FOCUS_adoption_fidelity_and_variants.md#what-phase-1-deployed; old2/HANDOFF_2026-04-22
- **relations:** adoption variants axis (ADO-005); single-agent trigger it replaced (ADO-018)
- **verify-later:** EnsureSiteRecordAction; apply_adoption_plan_action.go ~52169-52176

### ADO-005 — Adoption variants A–D (reference / structure / clone / analysis) and the unwired selector
- **status:** partial
- **status-evidence:** "the variant selector was never wired — everything defaults to the current behaviour, which sits roughly between A and C and commits to neither... Variant C is what's needed and does not yet exist in a meaningful sense" (FOCUS_adoption_fidelity_and_variants).
- **what:** Four adoption operations were designed: A reference-only (design inspiration), B design+structure (same pages, your content), C full clone (copy everything, rename), D multi-source analysis. The parameter plumbing (target_url/destination_domain, an `adoption_variant` field) exists but no selector reads it, so the live pipeline produces "a site-planner brief plus specs, not a deterministic copy" — sitting between A and C without committing to either.
- **sources:** FUTURE_adoption_source_destination_separation.md; FOCUS_adoption_fidelity_and_variants.md#the-adoption-variants; FOCUS_design_composition_flow_and_adoption_fidelity(1).md#4
- **relations:** fidelity dial (ADO-011, orthogonal axis: variant = what the operation is, dial = how much aspiration); source/destination separation (ADO-004)
- **verify-later:** adoption workflow input schema; any variant/clone parameter in site-adoption-orchestrator config

### ADO-006 — Adoption → classifier handoff: adoption writes specs first, classifier consumes under fidelity rules
- **status:** deployed
- **status-evidence:** Independently re-derived and confirmed by five different units across two months: apply_adoption_plan was flagged "NOT YET DEPLOYED" 2026-04-23, then verified live 2026-05-26 ("Adoption does not bypass the planner — it routes through it via the strategy→briefing→site_plan chain"), and re-confirmed 2026-07-04/2026-07-09 (README_flows: "your instinct was half right, inverted"). Most recent and most specific evidence is taken as authoritative.
- **what:** Adoption never calls the classifier directly. `apply_adoption_plan` writes site_archetype/design_reference/design_intent/content_direction/identity specs, pages, and work items itself, then emits exactly one strategic item, needs_domain_research — never needs_composition/needs_design directly. When the relay later reaches the site, the domain-research-classifier treats the adopted identity/archetype/content_direction/design_intent as ground truth outranking its own search+scrape, reads-and-extends rather than overwrites, always runs its full vertical research, and queues needs_strategy so strategist→briefing→planner run for adopted sites exactly as for fresh builds.
- **sources:** 007#Handoff to the classifier, #Post-adoption; HANDOFF_2026-04-23(1).md#not-deployed; HANDOFF_2026-05-26…md#verified; docs019/RUNBOOK_builder_route(21).md#B2, #B3; NOTES_running_synthesis_v4(39).md 2026-07-04; README_flows.md
- **relations:** classifier-as-strategic-brain (site-spec-and-classifier SPEC-011); work-item relay spine; reconciler (post-030 planner writes plan-domain tables, reconciler emits page items)
- **verify-later:** apply_adoption_plan_action.go current emissions; site_archetype writer; check_adoption_skip_scrape branch in the classifier workflow

### ADO-007 — Pattern extraction, code-as-reference, and RAG-fed generation
- **status:** aspirational
- **status-evidence:** 007 Phase 3 items, described entirely in future tense ("Runs as a side effect… patterns accumulate").
- **what:** A planned pattern-extraction-agent would mine research into reusable tool specs, layout/content patterns, and good/bad UX examples. Complex tool builds would include reference code in the prompt with an explicit original-implementation instruction (never deployed directly, on a copyright stance), and prompt+output pairs would feed RAG so future generations retrieve both abstract specs and concrete prior successes.
- **sources:** 007#Research, Patterns, and the Component Library
- **relations:** knowledge_base; tool-recreation-handler (ADO-027)
- **verify-later:** none built

### ADO-008 — Firecrawl capability escalation ladder (executeJavascript, waitFor, structured json)
- **status:** aspirational
- **status-evidence:** "These are upgrades, not prerequisites" (interactive FOCUS doc).
- **what:** When plain rawHtml plus external-CSS-fetch parsing misses dynamically-injected scripts or bundled logic, Firecrawl's executeJavascript actions (script inventory via querySelectorAll), waitFor, and schema-driven json extraction are the identified escalation path for the adoption parse stage.
- **sources:** FOCUS_interactive_content_generation(4).md#Firecrawl-features
- **relations:** interactive fingerprint C1–C6 (DYN-002)
- **verify-later:** firecrawl adapter capabilities actually used today

### ADO-009 — Duplicate sites-row on re-adoption (open investigation)
- **status:** deployed
- **status-evidence:** "Couldn't confirm … worth checking on next adoption run" (HANDOFF_2026-04-23, item 20; never revisited in later units seen).
- **stage2-verified (2026-07-14):** unknown → deployed — EnsureSiteRecordAction (site_db_actions.go:102) calls upsertSite() which does 'INSERT INTO sites ... ON CONFLICT (domain) DO UPDATE SET updated_at=NOW()' (site_db_actions.go:1011-1021); sites.domain has UNIQUE constraint 'sites_domain_key' (docs/agent_docs/sql_for_components/006_old_summary_table_descriptions.sql:49...
- **what:** Suspicion that adopting a destination_domain that already has a sites row creates a second row, leaving orphan work items pointing at the stale row while a new cascade runs against the other. The needed decision — refuse when destination exists vs reuse as refresh — was never made; duplicate-creation is agreed to be the worst outcome.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 20
- **relations:** source/destination separation (ADO-004); library-row cleanup
- **verify-later:** ensure_site_record behaviour on an existing domain

### ADO-010 — Fresh vs adoption entry paths converge on one cascade
- **status:** deployed
- **status-evidence:** Capability map with per-row verdicts ("already in fresh"); resolved empirically 2026-06-14 — a fresh submit flowed end-to-end through dispatch without manual triage.
- **what:** Two entry agents — domain-submitter (fresh: {domain,email,mission_brief}) and site-adoption-orchestrator (adopt: crawl→fingerprint→archetype→seeds) — converge on needs_domain_research and share the whole downstream cascade. The only adoption capabilities fresh lacks (CSS fingerprint, interactive-feature detection, full archetype) are inherently crawl-products, so a separate "fresh-build" single-agent copy was rejected as premature. The unified trigger `082_submit_domain_unified.sh` picks the entry path (--from ⇒ adopt) and gained `--mission-file`.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (capability map); idea.uk/HANDOFF(13).md (pipeline graph)
- **relations:** classifier handoff (ADO-006); fidelity dial (ADO-011)
- **verify-later:** 082 script; site-adoption-orchestrator definition

### ADO-011 — Adoption fidelity dial (locked/high/medium/low; phases 1–4)
- **status:** partial
- **status-evidence:** "Implementation status (the catch). Only Phase 1 exists: an adoption-aware classifier prompt giving implicit `high` fidelity at the prompt level… today fidelity is coarse prompt behaviour, not the deployed-vs-planned model" (FOCUS_design_composition_flow, 2026-05-26); the unified trigger records `--fidelity` but nothing reads it.
- **what:** Unifying idea: every input (bare domain, questionnaire, scraped live site) is the same thing at different fidelity, with adoption as the high-fidelity end of one pipeline. The planned dial (locked/absolute, high, medium, low; re-purposed as research-confidence for blank sites) would govern how much aspiration reaches the first build and how fast the improvement loop narrows the gap, flowing into a `build_policy`/`adoption_meta` spec aspect with per-item status. Only Phase 1 (implicit `high`) exists; Phases 2–4 (per-item spec status, explicit input, status-marked aspiration) are unbuilt.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#4; FOCUS_adoption_fidelity_and_variants.md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md
- **relations:** site-spec-and-classifier's platform-level fidelity dial (SPEC-003, same concept from the doc-028 side); variant axis (ADO-005); timed locks as interim enforcement
- **verify-later:** classifier prompt for adoption-aware fidelity; any fidelity/build_policy aspect in prod

### ADO-012 — Readopt-as-acceptance-test pattern
- **status:** aspirational
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §7c "only after §6 passes"; FOCUS_content_quality(2) work order 3 — planned, not yet run in these docs.
- **what:** After a fix batch verifies on the existing site, tear down and re-adopt the source (gamedesign.uk → gamesdesign.co.uk) as the from-scratch acceptance test and fresh content-quality baseline, so any failure is attributable to the virgin path. Expected recurrences (adopt-path defects untouched by the linking work) are pre-declared as the next package's input, not regressions; a corollary discipline is that site_id changes on every teardown, so always resolve via domain.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7c; running_notes_17(21).md#readopt-decision; FOCUS_content_quality(2).md#work-order
- **relations:** content-quality catalogue; debugging heuristics (site_id resolution)
- **verify-later:** whether the readopt ran post-2026-06-26; new site_id baseline audit results

### ADO-013 — Tool/game pages never deployed (A1): section-only parser + upsert flip churn
- **status:** deployed
- **status-evidence:** "A1 VERIFIED CLOSED… all five games committed… tools deploy… The three-file fix (parser fallback + deploy-time stamp + flip removal) is confirmed in production" (2026-06-03).
- **what:** `saveSectionsExtractFromHTML` only extracted `<section>…</section>` blocks, but tool-recreation-handler emits `<div class="tool-page">…`, so zero blocks matched, zero page_components were created, and rerender's `getPageSections` returned empty — the page was skipped with no git commit despite a `complete` work item. Root causes were pinned as (1) the section-only parser and (2) `upsertPage`'s ON CONFLICT flipping `deployed → needs_rebuild`. Fix: when zero section blocks match but the HTML is non-empty, store the whole fragment as one section (guarded against full documents), plus removing the flip. Key mechanism fact this surfaced: `build_status='deployed'` and file presence are independent — the deploy path depends on `page_components.rendered_html`, not `pages.sections`.
- **sources:** CATALOGUE_gamesdesign_post_sync_fix_defects.md#A1, CATALOGUE_gamesdesign_post_sync_fix_defects(4).md; running_notes_14(20) Parts 7–10
- **relations:** built_from_plan_version stamp; tool-recreation-handler (ADO-027); dispatch throughput bottleneck (masked the remaining 7 tools)
- **verify-later:** save_page_sections_action.go current parser logic; rerender_single_page_action.go assemblePage/getPageSections

### ADO-014 — Sectionless-page durability stack (sibling fallback + discovery check + no-sibling flag)
- **status:** partial
- **status-evidence:** "Durability code WRITTEN this session (NOT yet deployed) — see runbook" (HANDOFF_2026-06-09); the underlying skinner-box instance is "built and deployed… verified page_components=2".
- **what:** A page can sit in the plan with zero sections (residue from a killed build falsely marked complete), and every rebuild completes empty in ~90s because `check_has_ready_sections`'s ELSE branch routes to a SUCCESS-labelled `complete_error`. The fix stack: (2b) `load_page_sections_from_spec` gains a final fallback synthesising the layout from a same-role sibling (WARN-logged, layout skeleton only); (S1) a new self-registering `check_sectionless_pages` discovery check flags plan pages a sibling can repair and re-issues needs_content_page, closing a self-healing loop (relying on the built-in two-strike rule for churn control); (S2) a workflow-def change routes the no-sibling case to `mark_no_sections` → needs_human_review instead of silent success. A prerequisite `complete_work_item` guard stops the dispatch loop clobbering the flag.
- **sources:** RUNBOOK_section_sectionless_durability(2).md; running_notes_15(10); HANDOFF_2026-06-09; check_sectionless_pages(1).go header
- **relations:** silent-completion family; pages.sections as the build-read field
- **verify-later:** chassis image containing load_page_sections_from_spec_action.go + check_sectionless_pages.go; completeness-discovery-agent checks array

### ADO-015 — guide as a first-class page_type (adoption classifier vocabulary + retype + URL canonicalisation)
- **status:** deployed
- **status-evidence:** "guides typed page_type=guide directly… migration_adoption_add_guide_page_type.sql worked — adoption classifier emits guide, NO post-hoc re-typing needed" (2026-06-05 re-adoption); retype + URL migrations applied 2026-06-04.
- **what:** Adoption's `analyze_site` enum folded guides into blog-post (source guides lived flat at /blog/guide-*.html), so `query.pages_where_type:guide` returned zero. A first pass explicitly rejected retyping as `guide` as a quick fix because it would misplace the URL relative to the untouched source, leaving it an open product decision. The later, deliberate structural fix added `guide` to the classifier enum + guidance, re-typed the 5 content-bearing pages, and migrated URLs to the canonical `/guides/<slug>/index.html` (page_canonical.go already had the guide case) — the exact "less faithful" move earlier rejected, now chosen once `guide` was a proper page_type.
- **sources:** migration_adoption_add_guide_page_type.sql; migration_retype_guides_to_guide.sql; migration_guides_url_to_canonical.sql; running_notes_14(20) Parts 2, 13, 14a–14h
- **relations:** guide-list Tier-D resolution; bare-guide-duplicate defect (ADO-034); parallel classifier-vocabulary framing of the same fix (site-spec-and-classifier SPEC-009)
- **verify-later:** site-adoption-agent analyze_site enum in the live definition; pages.page_type values on a fresh adoption; whether build-site-planner's vocabulary was ever updated

### ADO-016 — Interactive fingerprint extraction gap (Path C: tools rebuilt as prose)
- **status:** aspirational
- **status-evidence:** Planned workflow insertion sketched ("extract_interactive_fingerprint (NEW)… between extract_fingerprint and check_has_external_css"); no deployment claim anywhere.
- **what:** Adoption pulls markdown but not `<script>`/`<canvas>` interactive machinery, so crawled calculator pages rebuild as paragraphs describing the calculator instead of working widgets. The planned second fingerprint pass over the same crawl_result, capturing interactive elements for tool recreation, was never built — in practice tool-recreation-handler plus the A1 fixes got real tools deploying from recreate prompts without it.
- **sources:** FOCUS_component_schema_patterns.md#appendix; old2/README.md; FOCUS_adoption_fidelity_and_variants.md#problems-ranked
- **relations:** tool-recreation-handler (ADO-027); interactive fingerprint parse stage C1–C6 (DYN-002, the fuller design of this same gap)
- **verify-later:** site-adoption-agent workflow for any extract_interactive_fingerprint step (expected absent)

### ADO-017 — Adoption resume logic (never built)
- **status:** abandoned
- **status-evidence:** "orchestration_states.collected_data already persists per-step output (378KB survived a failed run), but ResumeWorkflowTopic has no subscriber — resume was anticipated, never built. User: 'a new crawl is fine.'"
- **what:** Mid-workflow resume of a failed adoption from persisted collected_data was designed but never consumed — the plumbing half-exists (state persistence, a resume topic) with no subscriber, and the accepted operational answer became re-crawl/re-adopt rather than resume.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#deferred; old2/HANDOFF_2026-04-22#resume-logic
- **relations:** error_step fix (reduced the need — CSS timeout no longer fatal); teardown+re-adopt operational pattern
- **verify-later:** ResumeWorkflowTopic subscribers (expected none)

### ADO-018 — Single-agent adoption trigger, superseded by thin wrapper orchestrator
- **status:** superseded
- **status-evidence:** v2 doc: a dedicated `site-adoption-agent` triggered directly via `./trigger-adopt-site.sh gamedesign.uk`; the patch rewrites this into "Two agents, one thin wrapper," fully documented in the live v4 doc.
- **what:** Adoption originally ran as one agent invoked directly by a shell script with a positional domain argument, mixing "site being crawled" and "site being built" into a single identifier. Replaced by the thin `site-adoption-orchestrator` plus a JSON trigger payload separating `target_url` from `destination_domain`, while keeping the old `url`/`domain` shape as legacy-compatible input.
- **sources:** old/older1/007_adoption_pipeline_v2_april26.md#"The adoption agent"; old/older1/007_adoption_pipeline_v2.patch.diff; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Source vs destination"
- **relations:** "every pod-running agent needs a parent that spawned it" (development-guide); source/destination separation (ADO-004)
- **verify-later:** confirm site-adoption-orchestrator agent_definitions row exists and trigger-adopt-site.sh uses the JSON payload shape today

### ADO-019 — Unified `design` spec aspect, superseded by design_reference/design_intent split
- **status:** superseded
- **status-evidence:** v1 doc: single `design` spec aspect; live v4 doc has a dated addendum, "Design Fingerprint Pipeline (added 2026-04-12)," documenting the two-aspect replacement.
- **what:** The earliest adoption design captured only one blended `design` spec aspect — palette and typography guessed by the LLM alongside identity/structure classification, with no separation between what the source site actually used and what the new site should aim for. Replaced by the design_reference (concrete, historical)/design_intent (semantic, evolvable) split.
- **sources:** old/older1/007_adoption_pipeline.md#"What gets stored where"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Design Fingerprint Pipeline"
- **relations:** three-stage adoption pipeline (ADO-003, current successor)
- **verify-later:** check site_specs for any legacy rows with aspect='design' from pre-2026-04-12 adoptions never migrated

### ADO-020 — Two-stage → three-stage adoption processing (historical evolution)
- **status:** superseded
- **status-evidence:** Archive header: "Two-stage processing: LLM classifies, Go extracts." Live header: "Three-stage processing: Go extracts design, LLM classifies, Go extracts content."
- **what:** Early adoption split work into just two stages — lightweight LLM classification from page summaries, then Go-only content extraction. The later design inserts a Go-only design-fingerprint stage ahead of LLM classification, on the principle "don't ask an LLM to read hex values when a regex can do it."
- **sources:** old/older1/007_adoption_pipeline.md#"Two-stage processing"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Three-stage processing"
- **relations:** current three-stage pipeline (ADO-003); unified design aspect (ADO-019)
- **verify-later:** extract_design_fingerprint_action.go/enrich_fingerprint_with_css_action.go existence and wiring

### ADO-021 — Section recipes for adoption (purpose + structure + reference implementation)
- **status:** aspirational
- **status-evidence:** Listed under "Phase 4: Requirement-Driven Components (longer term)" in the 2026-04-11 plan; no confirmation of shipping in any later doc reviewed.
- **what:** When adopting a site, each section would be captured as a "recipe" — purpose, structure, reference implementation (as a guide, not a spec), and a component match — and recipes without a good match would generate `needs_new_component` work items where the recipe becomes the build brief.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale", #"Phase 4"
- **relations:** component selector by functional requirement; needs_new_component work items
- **verify-later:** whether any adoption workflow step produces structured "recipes" today

### ADO-022 — Adopt-from vs deploy-to separation (unbuilt staging area)
- **status:** aspirational
- **status-evidence:** "discussed but not implemented. Options: snapshot to S3, stage to subdomain, or store crawl artifacts."
- **what:** Unbuilt idea for a staging area distinct from the live deploy target, so a freshly-adopted rebuild could be reviewed before overwriting production. The workaround at time of writing was manual: pause work items, verify specs, unpause.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Architecture Decisions Made"
- **relations:** site snapshots and revert (doc 014); design fingerprint extraction pipeline
- **verify-later:** whether any staging/subdomain mechanism exists for adoption today

### ADO-023 — Adoption interactivity misroute — canonical-prefix key desync in buildPageFeatureMap
- **status:** deployed
- **status-evidence:** "Deployed: T1 (apply_adoption_plan_action.go — canonical-keyed buildPageFeatureMap)... Both are in production" (HANDOFF_2026-05-26); root cause independently confirmed via query G in a separate postmortem doc.
- **what:** `apply_adoption_plan_action.go` routes adopted pages by interactivity (`len(page.Features) > 0` → needs_tool_recreation; else → needs_content_page). `buildPageFeatureMap` keyed its feature map by the raw adoption-LLM page key, but the routing loop looked up the canonicalised name via `CanonicalisePage`, whose `tool` branch adds a `tool-` prefix while its `game` branch preserves an already-present `game-` prefix — so every tool page's feature lookup missed even when the LLM correctly detected interactivity (silently misrouted to the static page-build-handler path), while games matched only by coincidence. Fix: key buildPageFeatureMap by the canonical name.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.6,§2.7,§5b; tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2; docs/_archive/.../016_debugging_guide_addendum_adopted_tools_no_widget(3).md
- **relations:** tool routing fix deployment status (ADO-024); component-regeneration-flow clobber path; doc 029 CanonicalisePage
- **verify-later:** confirm in production that buildPageFeatureMap still contains the canonical-keyed version — a parallel adoption chat may have re-edited the file since merge

### ADO-024 — Tool routing fix deployment status: T1+T2 in production, symptom fix unconfirmed
- **status:** partial
- **status-evidence:** "Deployed: T1... and T2... Both are in production." / "Not confirmed: that widgets now actually deploy... Hold the trigger. Do not mass-emit needs_tool_recreation yet" (HANDOFF_2026-05-26_tool_routing_fix_deployed.md).
- **what:** The authoritative status record for the tool-widget-clobber investigation as of 2026-05-26: the routing fix (T1) and a detection check (T2) are both confirmed deployed, with defined acceptance criteria for calling the fix complete (every tool/game page has_widget=t; a deployed tool page renders an interactive widget in-browser; T2 finds nothing new steady-state; no duplicate pages). None of those criteria were confirmed met at time of writing — the recreation-loss defect remained open and blocking.
- **sources:** tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1, §6, §7
- **relations:** canonical-prefix key desync (ADO-023); recreation-loss defect
- **verify-later:** re-run the exact acceptance-criteria queries against current gamesdesign.co.uk state

### ADO-025 — Adoption faithfulness — WriteSitePlanAction identity strip
- **status:** partial
- **status-evidence:** "The corruption is in WriteSitePlanAction, not the LLM… Fix direction (not yet applied)" (016 debugging guide v2_44 §9).
- **what:** Even after a faithful adoption, `WriteSitePlanAction`'s `ValidateRoles`+`CanonicalisePage` interaction permanently strips identity for `content`/`blog_post` page_types: `ValidateRoles` derives a slug that strips `tool-/guide-/game-`/`-index`, and `CanonicalisePage` only re-adds prefixes for tool/game/guide roles, so mistyped section-index hubs flatten. Root cause is the wrong `page_type`; the clean fix is upstream at adoption time, not applied as of this record.
- **sources:** WM/016_debugging_guide_v2_44.md#adoption-faithfulness-llm-convergence; WM/ARCHITECTURAL_TENSIONS(2).md#tension-2
- **relations:** CanonicalisePage; adoption_locked locks; FOCUS_adoption_faithfulness_via_locks
- **verify-later:** WriteSitePlanAction; datahelpers/page_canonical.go ValidateRoles/normaliseSlug; analyze_site page_type

### ADO-026 — site-scraper (Firecrawl scrape → site_context), ancestor of full-crawl adoption
- **status:** deployed
- **status-evidence:** 032 definition ("Uses firecrawl_scrape action"); docs017/003 describes the matured standardized site_context schema (source: database|scrape|manual).
- **what:** Scrapes a live site's homepage via the webscrape adapter (Firecrawl), then an LLM step transforms results into the standardized site_context format (domain, company, industry, palette, typography, component functions, source) that webdesign-agent consumes from any source — enabling "scrape competitor → feed to design agent → apply to your site" flows. The original design-transfer mechanism and direct ancestor of site-adoption-agent's full multi-page crawl.
- **sources:** 032_site_scraper_agent.sql; docs017_legacy_agent_rules_images_design_keydocs/003_design.md#Architecture, #008_checklist_for_new_specialist_agents_v5.md#Standardized-Interface-Schemas
- **relations:** webdesign-agent; standardized interface schemas doctrine; three-stage adoption pipeline (ADO-003, its successor)
- **verify-later:** whether site-scraper is still used vs site-adoption-agent; load_site_for_design action

### ADO-027 — tool-recreation-handler (recreate interactive tools from crawled source)
- **status:** deployed
- **status-evidence:** 099 definition; live-run evidence in 138 (run e1018366 recreated bugs faithfully → prompt fix) and 132/137 (note-writing wired, subject corrected).
- **what:** Two-stage recreation of JS-heavy pages during site adoption: analyze_tool (LLM functional spec from source + context), then recreate_tool (Opus generates working replacement HTML/CSS/JS), with completeness/truncation checks, validation, and save/deploy. 138 added a "Mandatory Behaviour Requirements" prompt section rendered from spec.interactive_features that OVERRIDES the original source, fixing an observed failure where explicit spec fixes were buried in analysis JSON and Opus faithfully recreated the original bugs.
- **sources:** 099_tool_recreation_handler.sql; 138_recreate_tool_carries_spec_features.sql; 137_recreation_spec_and_note_subject.sql
- **relations:** site-adoption-agent creates its work items (ADO-003); tool acceptance verification; interactive fingerprint gap (ADO-016)
- **verify-later:** current recreate_tool prompt; spec.interactive_features producers

### ADO-028 — 11-agent website analysis framework (four agent groups) — legacy predecessor
- **status:** superseded
- **status-evidence:** Whole docs003 set is planning ("Here is the detailed, point-by-point analysis…"); docs004 explicitly reframes it ("The old numbers are meaningless now… rename them") into the Learn/Execute playbook model.
- **what:** The original web-capture master plan grouped agents into Strategy & Content (Strategist, Content Infuser), Library & Storage (Librarian, S3+Postgres/pgvector), Design Ingestion (Prospector, Site Profiler, Capture Bot/Playwright, Layout & Labeling XY-Cut+LLaVA, Component Generator VLM screenshot-to-code, Style Extractor getComputedStyle — later eliminated for Firecrawl branding data, Behavior Extractor CodeLlama), and Generation (Publisher showcase site, Architect template builder via CLIP embedding). All were planned as agent_definitions rows, not new binaries.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** successor chain: playbook model → MVP site builder → current adoption-pipeline and site-spec-and-classifier; Publisher's public design-library site was abandoned
- **verify-later:** which of the 11 agent types ever got agent_definitions rows

### ADO-029 — website-analyzer conditional scraping group
- **status:** deployed
- **status-evidence:** Tested with kcat messages (boxing-tickets.com, basic and structured+crawl variants) and a live UPDATE of its orchestration_workflow.
- **what:** An agent group taking target_url + flags (extract_structured, crawl_pages, crawl_limit/depth) and conditionally routing between basic scrape, structured extraction, and multi-page crawl using evaluate_condition — the first "smart" capture entry point, since superseded in practice by the adoption pipeline's crawl/classify flow.
- **sources:** docs003_firecrawl/README.0129.testing_webscrape_message.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** firecrawl adapter; successor: adoption pipeline (ADO-003)
- **verify-later:** agent_group_definitions row group_type='website-analyzer'

### ADO-030 — Playwright capture adapter + website-capture agent, superseded by Firecrawl
- **status:** superseded
- **status-evidence:** Complete deliverables existed (adapter, capture_actions.go, agent SQL) but docs004/002 records "Agent 5 eliminated… use Firecrawl branding data instead."
- **what:** Deep browser-based capture: desktop/mobile screenshots, DOM, computed styles, interaction states (hover/focus for up to 50 selectors), scroll-position screenshots with parallax/sticky detection, asset extraction, and organised S3 upload. Deferred in favour of the managed Firecrawl service for MVP; the deeper capture ideas never resurfaced.
- **sources:** docs004_website_capture_project/playwright/website_capture_agent.sql, playwright_adapter.py, implementation_roadmap.md
- **relations:** firecrawl adapter (chosen replacement); behaviour capture (rrweb) idea also abandoned
- **verify-later:** playwright-adapter deployment existence

### ADO-031 — Website-builder orchestrator (capture → vision → code → synthesis → content → library) — abandoned maximal vision
- **status:** abandoned
- **status-evidence:** Orchestrator SQL references agent types (website-vision, website-code-analyzer, website-synthesis, content-strategist) and actions never defined or mentioned again; the MVP builder took a different shape.
- **what:** A master workflow to rebuild a site from a captured one — capture, visual analysis, code cleaning, synthesis into a template, content planning, parallel section generation, aggregation, and component storage with embeddings. The maximal "clone-and-improve" vision that was never realised.
- **sources:** docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql, website_builder_integration_guide.md
- **relations:** successor in spirit: adoption-pipeline content recreation (ADO-003); vision analysis resurfaces in current image-analysis tooling
- **verify-later:** confirm none of the four sub-agent types exist in agent_definitions

### ADO-032 — Adopting existing external sites ("Adopt" workflow) — legacy precursor
- **status:** superseded
- **status-evidence:** Designed in doc 004 ("Adopt workflow… status: 'adopted_partial'… match_confidence"); the current adoption-pipeline (site crawling, classification, content recreation) is the named live successor.
- **what:** Run the Learn loop against an existing site the platform didn't build: scrape, deconstruct layout, match found blocks to the in-house component library with confidence scores, and generate a manifest marking it adopted_partial, making external sites partially manageable by agent edit workflows.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md
- **relations:** successor: current adoption-pipeline (ADO-003)
- **verify-later:** compare with current adoption pipeline design

### ADO-033 — Site interrogation & pattern library
- **status:** aspirational
- **status-evidence:** "'Interrogate' successful sites to extract... Store extracted patterns" restated across docs009, docs010 (Phase 4), and docs012's 5-phase pipeline with pattern_sources marked "(future)".
- **what:** Learning from successful sites without copying: capture HTML+screenshot, LLM-analyse section types/visual hierarchy/content strategy/psychological principles, extract reusable patterns tagged by industry/funnel-stage/audience with "why it works" notes, and mint content_components (origin_type='extracted') from them. The most persistent unfulfilled idea of this era, restated across three separate roadmap eras without ever shipping.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-3; docs009_site_interrogation_and_solutions/003_claude_save_point.md#2; docs010_multitrack_flows_persona_architecture/018_priority_matrix.md
- **relations:** Pragmatic Evolution Engine phase 2; adoption-pipeline site crawling as the current descendant; component library
- **verify-later:** pattern_sources table; origin_type/industry_tags/funnel_stages columns on content_components

### ADO-034 — Bare-guide / spurious duplicate pages from "planner ignores adopted state"
- **status:** deployed
- **status-evidence:** "DECISIVE (llm_call_log plan_site)... the planner WAS given the adopted guides and emitted `economy-basics` anyway → PROMPT-RULE gap... NOT a wiring/status gap"; cleanup migration applied and confirmed durable ("current-plan bare-name query returns 0 rows").
- **what:** `build-site-planner`/`blog-content-planner` invents a differently-slugged sibling page (e.g. `economy-basics`) for a topic already adopted under a prefixed name (`guide-economy-basics`), because its "never duplicate an existing page" prompt rule only named games/tools examples and didn't generalise to the `guide-` prefix pattern — "a second surface invents parallel pages after adoption." A durable Go-level topic-stem collision guard (reusing CanonicalisePage's prefix-stripping) was recommended but not shipped; only the data cleanup migration (removing bare pages table rows, site_plan_pages/sections, and terminalising dangling work items) and an optional prompt-rule stopgap were delivered.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 14c–14g; docs/_archive/agent_docs/sql_for_tables/040b_migration_cleanup_bare_guide_duplicates(1).sql
- **relations:** "type guides as guide" structural fix (ADO-015); tool-page canonicalisation misroute (ADO-023, same "adoption vs. a second surface" bug family); FOCUS_planner_ignores_adopted_state.md; site-plan-and-reconciler
- **verify-later:** FOCUS_planner_ignores_adopted_state.md; whether the upstream planner prompt/logic has since been tightened to check adopted state first

### ADO-035 — Doc-tree adoption plan (constitution + tag/embedding retrieval) [category mismatch: this is a documentation-retrieval design, not a site-adoption concept — filed here only because the raw block was labelled adoption-pipeline]
- **status:** aspirational
- **status-evidence:** FOCUS_doc_tree_adoption.md header: "actionable plan … without committing to the atomic rewrite, the mediator, or the routing build"; "the corpus does not fit in context (~200 files, ~6.7MB, ~1.0–1.7M tokens)".
- **what:** First path to value from a doc-tree design against the current setup: Phase 1 write a tiny constitution, Phase 2 tag existing docs by concern/applies_to into a manifest, Phase 3 make the retrieval split real (tag-based deterministic selection for rules; existing nomic/pgvector/ollama RAG for the broad corpus), Phase 4 atomic extraction deferred/evidence-driven.
- **sources:** ED/FOCUS_doc_tree_adoption.md#2, #4, #5
- **relations:** atomic standard; mediator routing; RAG actions (existing stack)
- **verify-later:** rag_actions/nomic prefixes; proposed doc_index/standards table

### ADO-036 — Vertical-slice dogfooding of the automation ratchet [category mismatch: about the self-development/ratchet methodology, not site adoption — "adoption" here means adopting a way of working]
- **status:** aspirational
- **status-evidence:** MASTER(4) §8.1 "Vertical slice, not horizontal layer … Dogfood the ratchet … First capability = writing Go actions".
- **what:** Walk one capability (writing Go actions) end-to-end through route→produce→verify→gate→feedback before generalising; each new machinery piece starts at `confirm-every` and graduates on evidence. Phases 1–2 double as "improve my current chat workflow"; phases 3–6 are the leap to autonomy.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#8.1, #8.2, #8.5
- **relations:** automation ratchet; self-development coding pipeline
- **verify-later:** none

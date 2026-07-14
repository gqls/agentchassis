# Cluster: adoption-pipeline-and-cases
Categories included: adoption-pipeline, site-spec-and-classifier, dynamic-applications, site-case-studies


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Infrastructure three layers (core platform / client delivery / framework builder)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 007: Layer 1 exists; Layer 2 "planned" beyond static serving; Layer 3 "future"
- **what:** Layer 1 (K8s factory) never serves external traffic — it produces artifacts and pushes them. Layer 2 = client delivery (S3 static now; site-api-router + client Postgres on OVH VMs planned, config-driven routes reusing the action library). Layer 3 = provisioning agent frameworks for clients. Backend capability tiers 1–5 (static→full platform) with static-first principle; vetcomparison JSON-on-S3 pattern handles up to ~50k items (Pagefind extends to ~500k).
- **sources:** 007#Infrastructure Separation, #Backend Capability Tiers, #Site API Router
- **relations:** dynamic applications tiers (022); P1 marketing/OpenClaw
- **verify-later:** any site-api-router code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption is a one-off capture, not a ceiling (specs separation)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 v4 principles + what-gets-stored table; patch updates to Phase-1 split
- **what:** Crawl data → research_results (adoption_crawl/adoption_page), never site_specs. Spec aspects: identity (with adopted_from provenance), design_reference (concrete extracted values, historical, never modified), design_intent (semantic brand-level brief, survives plan rebuilds post-030), site_archetype (character + inviolable constraints), content_direction (brand voice), structure. Webdesign reads intent not reference; evolution = update intent. The strategist then writes aspiration beyond the adopted baseline.
- **sources:** 007#Site Adoption, #What gets stored where (incl. patch revision); 004#Adopted Sites
- **relations:** 028 fidelity; 030 strategic-vs-plan-time split
- **verify-later:** site_specs aspects on an adopted site

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption agent three-stage processing (Go fingerprint / LLM classify / Go content+plan)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 16-step workflow with runtime expectations and success signals
- **what:** site-adoption-orchestrator (wrapper) spawns site-adoption-agent: firecrawl crawl → Go extract_design_fingerprint (goquery over rawHTML: hexes, fonts, CSS vars, layout patterns, Google Fonts; external CSS fetched and merged; suggested_mapping to our var names) → LLM analyze_site on ~500-char page summaries + classify_archetype + derive_content_direction + generate_design_intent → Go apply_adoption_plan (buildCrawlPageIndex, page records, per-page markdown to research_results, design_reference spec, work items). Principle: LLM for reasoning, Go for extraction — never pay an LLM to transcribe.
- **sources:** 007#The adoption agent, #Three-stage processing, #Design Fingerprint Pipeline
- **relations:** wrapper pattern; classifier handoff
- **verify-later:** extract_design_fingerprint_action.go; adoption agent definition

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Source vs destination separation in adoption (target_url / destination_domain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 modes table + operational signals ("ApplyAdoptionPlanAction: source differs from destination")
- **what:** Adoption separates the crawled reference site from the site being built; ensure_site_record uses destination via domain_override_field, crawl hits target_url, provenance records source host — which also keys the crawl-content lookup (mismatch silently drops all adopted content). Legacy single-domain shape still accepted.
- **sources:** 007#Source vs destination, #Adoption modes
- **relations:** build-inspired-by mode (goes via classifier instead)
- **verify-later:** apply_adoption_plan source/destination handling

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption → classifier handoff (needs_domain_research; no shortcuts)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 section + patch; grounded in 028's "classifier always runs in full"
- **what:** Adoption queues exactly one strategic item, needs_domain_research; it does NOT queue needs_composition/needs_design directly — the cascade produces them naturally. The classifier reads site_archetype/design_reference as ground truth, read-and-extends identity/content_direction/design_intent, always runs vertical research, emits classification, queues needs_strategy. Post-030: planner writes to plan-domain tables and the reconciler emits page items; the planner's job ends at "the new plan is durably current".
- **sources:** 007#Handoff to the classifier, #Post-adoption (patched version); 028#The classifier is the strategic brain
- **relations:** unified pipeline; reconciler
- **verify-later:** adoption work-item emissions in apply_adoption_plan

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Pattern extraction, code-as-reference, and RAG-fed generation
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** 007 Phase 3 items; "Runs as a side effect… patterns accumulate" future tense
- **what:** A pattern-extraction-agent mines research into reusable tool specs, layout/content patterns, and good/bad UX examples; complex tool builds include reference code in the prompt with explicit original-implementation instruction (never deployed directly — copyright stance); prompt+output pairs feed RAG so future generations retrieve both abstract specs and concrete prior successes.
- **sources:** 007#Research, Patterns, and the Component Library
- **relations:** knowledge_base; tool-recreation-handler
- **verify-later:** none built

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Design fingerprint extraction pattern (adoption parse stage for design)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** described as "the proven template" (interactive FOCUS, sequencing locked 2026-05-14); exercised live in 04-23 cascade (crawl → fingerprint → fetch CSS → enrich → analyze)
- **what:** Firecrawl rawHtml parsed Go-side (goquery) by extract_design_fingerprint for colours/fonts/CSS vars/layout/dark sections; external CSS fetched via firecrawl_scrape and merged (EnrichFingerprintWithCSSAction); an LLM step (generate_design_intent) produces the semantic brief; stored as design_reference (concrete) + design_intent (semantic) spec aspects. The template any other parse-stage extractor copies.
- **sources:** FOCUS_interactive_content_generation(4).md#Adoption; HANDOFF_2026-04-23(1).md#validated
- **relations:** interactive fingerprint (clone of this pattern); site-design-planner palette source
- **verify-later:** extract_design_fingerprint action; design_reference/design_intent aspects on adopted sites

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption source/destination separation and variant axis
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** FUTURE doc (2026-04-20) "Status: Future work … Option 1 sketched"; but 2026-04-23 handoff triggers "the standard adopt-separated kcat command SOURCE_URL=… DEST_DOMAIN=…" — Option 1 evidently landed
- **what:** Decouple the crawled target from the built destination (target_url + destination_domain inputs; ensure_site_record override) and gate spec-writing on an adoption_variant: reference (design only), structure (+archetype/pages), clone (+content_direction — old behaviour), analysis (aggregate competitor_landscape). Phase 2: sites.source_site_id provenance; Phase 3: adoption_references library. Risks: variant-gated data bleed, typo'd destination domains creating junk sites.
- **sources:** FUTURE_adoption_source_destination_separation.md (whole); HANDOFF_2026-04-23(1).md#priority-4
- **relations:** adoption faithfulness (fidelity axis); duplicate-sites-row question
- **verify-later:** apply_adoption_plan variant gating; extractDestinationDomain in ensure_site_record_action.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption → classifier handoff (classifier as strategic brain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** apply_adoption_plan rewrite "NOT YET DEPLOYED" 2026-04-23, but 2026-05-26 verifies "Adoption does not bypass the planner — it routes through it via the strategy→briefing→site_plan chain, as 007_adoption_pipeline_v4.md intended"
- **what:** Adoption stops queueing needs_composition/needs_design directly; it writes its specs and emits a single needs_domain_research item so the classifier (with dynamic taxonomy) then strategist → briefing → planner run for adopted sites exactly as for fresh builds — doc 028's ownership model applied to adoption.
- **sources:** HANDOFF_2026-04-23(1).md#not-deployed; HANDOFF_2026-05-26…md#verified
- **relations:** dynamic taxonomy classifier; spec ownership contract; pipeline cascade
- **verify-later:** apply_adoption_plan_action.go current emissions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Firecrawl capability escalation ladder (executeJavascript, waitFor, structured json)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "These are upgrades, not prerequisites" (interactive FOCUS)
- **what:** When plain rawHtml + external-fetch parsing misses dynamically-injected scripts or bundled logic, Firecrawl's executeJavascript actions (script inventory via querySelectorAll), waitFor, and schema-driven json extraction are the escalation path for the parse stage.
- **sources:** FOCUS_interactive_content_generation(4).md#Firecrawl-features
- **relations:** interactive fingerprint C1-C6
- **verify-later:** firecrawl adapter capabilities used today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Duplicate sites-row on re-adoption (open investigation)
- **category:** adoption-pipeline
- **status-signal:** unknown
- **status-evidence:** item 20 (2026-04-23, version (1) only): "Couldn't confirm … worth checking on next adoption run"
- **what:** Suspicion that adopting a destination_domain that already has a sites row creates a second row, leaving orphan work items pointing at the stale row while a new cascade runs against the other. Decision needed: refuse when destination exists vs reuse as refresh; duplicate-creation is the worst option.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 20
- **relations:** source/destination separation; library-row cleanup
- **verify-later:** ensure_site_record behaviour on existing domain

---

## Proposed NEW categories

| slug | why |
|---|---|
| NEW:work-dispatch | The detected→triaged→claimed state machine, dispatch chain, claim blockers/timeouts, pipeline label, two-strike rule and silent-completion semantics form a coherent expert domain that spans (and is not owned by) improvement-loop or scheduler-and-tasks. 10 concepts landed here. |
| NEW:prompt-composition | Prompt architecture (mega-prompt fragility, envelope/tool-call/validation patterns, parameter-shaping for images) is design-of-prompts, distinct from llm-quality-testing's evaluation focus. |
| NEW:language-i18n | Language/i18n surfaces (implicit language mechanism, hardcoded-English map, lang attribute, soft statics) have no home in the seed taxonomy. |

<!-- SOURCE: U04_idea_uk.md -->
### Fresh vs adoption entry paths converge on one cascade (fresh = adoption minus the crawl)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Capability map table with per-row verdicts ("already in fresh"); resolved empirically 2026-06-14 — a fresh submit flowed end-to-end through dispatch without manual triage.
- **what:** Two entry agents — domain-submitter (fresh: {domain,email,mission_brief}) and site-adoption-orchestrator (adopt: crawl→fingerprint→archetype→seeds) — converge on needs_domain_research and share the whole cascade (classifier read-and-extends adopted seeds → strategist → briefing → planner → composition → design → pages → rerender). The capability map shows the only adoption capabilities fresh lacks (CSS fingerprint, interactive-feature detection, full archetype) are inherently crawl-products; a new "fresh-build" single-agent copy was rejected as premature — reuse the existing path. The unified trigger `082_submit_domain_unified.sh` picks the entry (--from ⇒ adopt) and gained `--mission-file` (used to ship idea.uk's mission). The richest "seed with the existing setup" is adoption pointed at the live site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (capability map + submission entry points); idea.uk/HANDOFF(13).md (pipeline graph)
- **relations:** Phase 0 read; fidelity dial; adoption teardown vs fresh detach.
- **verify-later:** 082 script; site-adoption-orchestrator definition.

<!-- SOURCE: U04_idea_uk.md -->
### Fidelity dial — documented but not wired
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "The trigger records --fidelity but it doesn't yet modulate the build… doc 028's explicit fidelity input and its build_policy/adoption_meta aspect and per-item status are not yet wired (fidelity is currently implicit high)."
- **what:** A planned locked/high/medium/low fidelity input governing how faithfully a build reproduces its source/spec, flowing into a build_policy aspect with per-item planned/deployed status. Today the unified trigger records the flag and nothing reads it — a clean doc-vs-reality gap flagged repeatedly in the handoffs.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (launch idioms); idea.uk/HANDOFF(13).md
- **relations:** doc 028 (design/adoption); fresh vs adoption convergence.
- **verify-later:** any consumer of the fidelity field.

<!-- SOURCE: U05_content_quality_linking.md -->
### Readopt-as-acceptance-test pattern
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §7c "only after §6 passes"; FOCUS_content_quality(2) work order 3 — planned, not yet run in these docs.
- **what:** After a fix batch verifies on the existing site, tear down and re-adopt the source (gamedesign.uk → gamesdesign.co.uk) as the from-scratch acceptance test and the fresh content-quality baseline — any failure then attributable to the virgin path. Expected recurrences are pre-declared (adopt-path defects untouched by the linking work: brand-suffix titles, tool-flavoured guide copy, empty descriptions, footer metadata) and are the next package's input, not regressions. Corollary discipline: site_id changes on every teardown — always resolve via domain.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7c; running_notes_17(21).md#readopt-decision; FOCUS_content_quality(2).md#work-order
- **relations:** content-quality catalogue; adoption pipeline; debugging heuristics (site_id).
- **verify-later:** whether the readopt ran post-2026-06-26; new site_id baseline audit results.

<!-- SOURCE: U09_adoption.md -->
### Site-adoption agent pipeline (crawl → fingerprint → analyze → classify → apply)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Workflow traced in FOCUS_component_schema_patterns appendix ("crawl_site (firecrawl_crawl…) → format_crawl → check_crawl_content → extract_fingerprint → check_has_external_css → fetch_primary_css → enrich_fingerprint → analyze_site"); repeated verified gamesdesign adoption runs through 2026-06-05.
- **what:** The `site-adoption-agent` crawls a source site (Firecrawl, markdown + rawHtml, limit 30), extracts a design fingerprint (CSS variables, typography, palette), enriches with fetched primary CSS, classifies pages via `analyze_site` (LLM emits per-page page_type + url), derives content direction and archetype, generates a `design_intent` spec, and `apply_adoption_plan` writes pages + specs. It writes pages and specs but no site_plan — the build cascade (build-site-planner etc.) runs later.
- **sources:** FOCUS_component_schema_patterns.md#appendix, FOCUS_adoption_fidelity_and_variants.md#the-core-gap, HANDOFF_2026-05-25#standing-context
- **relations:** adoption variants, adoption fidelity dial, wrapper-orchestrator pattern, guide/game page_type vocabulary
- **verify-later:** agent_definitions row `site-adoption-agent` (id 4e2d8e8e…), `apply_adoption_plan_action.go`, `extract_design_fingerprint_action.go`, `enrich_fingerprint_with_css_action.go`

<!-- SOURCE: U09_adoption.md -->
### Adoption source/destination separation (target_url vs destination_domain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** "Phase 1 plumbing (`target_url` / `destination_domain` separation) is deployed" (FOCUS_adoption_fidelity_and_variants); "source/destination separation deployed and working".
- **what:** Adoption previously conflated source and destination into one site_id. Phase 1 parameterised it: `target_url` = what to crawl, `destination_domain` = what to build (legacy url/domain still accepted), via migrations 001–004 plus the `sourceDomain` vs `domain` fix in `apply_adoption_plan_action.go` that was silently dropping page content when source ≠ destination.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#what-phase-1-deployed, old2/HANDOFF_2026-04-22
- **relations:** adoption variants (the selector this separation was built for)
- **verify-later:** `EnsureSiteRecordAction`, `apply_adoption_plan_action.go` ~52169-52176

<!-- SOURCE: U09_adoption.md -->
### Adoption variants A–D and the unwired variant selector
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "the variant selector was never wired — everything defaults to the current behaviour, which sits roughly between A and C and commits to neither… Variant C is what's needed and does not yet exist in a meaningful sense" (FOCUS_adoption_fidelity_and_variants).
- **what:** Four adoption operations defined in FUTURE_adoption_source_destination_separation: A reference-only (design inspiration), B design+structure (same pages, your content), C full clone (copy everything, rename), D multi-source analysis. Plumbing exists but no selector; the current pipeline produces "a site-planner brief plus specs, not a deterministic copy" — the gap between specs+LLM interpretation and the actual source site.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#the-adoption-variants, FOCUS_design_composition_flow_and_adoption_fidelity(1).md#4
- **relations:** fidelity dial (orthogonal axis: variant = what the operation is, dial = how much aspiration); faithful-first-pass locks
- **verify-later:** adoption workflow input schema; any `variant`/`clone` parameter in site-adoption-orchestrator config

<!-- SOURCE: U09_adoption.md -->
### Adoption fidelity dial (locked/high/medium/low; phases 1–4)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Implementation status (the catch). Only Phase 1 exists: an adoption-aware classifier prompt giving implicit `high` fidelity at the prompt level… today fidelity is coarse prompt behaviour, not the deployed-vs-planned model" (FOCUS_design_composition_flow §4, 2026-05-26).
- **what:** Unifying idea: every input (bare domain, questionnaire, scraped live site) is the same thing at different fidelity — adoption is the high-fidelity end of one pipeline. The dial (locked/absolute, high, medium, low; re-purposed as research-confidence for blank sites) governs how much aspiration reaches the first build and how fast the improvement loop narrows the gap. Real dial needs Phase 2 (per-item status on specs), Phase 3 (explicit `build_policy`/`adoption_meta` input), Phase 4 (classifier produces status-marked aspiration alongside faithful baseline).
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#4, FOCUS_adoption_fidelity_and_variants.md
- **relations:** doc 028 platform mission; timed locks are the enforcement; variant axis
- **verify-later:** classifier prompt for adoption-aware fidelity; site_specs per-item status columns (Phase 2 — expected absent)

<!-- SOURCE: U09_adoption.md -->
### Tool/game pages never deployed (A1): save_page_sections `<section>`-only parser
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** "A1 VERIFIED CLOSED… all five games committed… tools deploy… The three-file fix (parser fallback + deploy-time stamp + flip removal) is confirmed in production" (2026-06-03).
- **what:** `saveSectionsExtractFromHTML` extracted only `<section>…</section>` blocks, but tool-recreation-handler emits `<div class="tool-page">…` — zero matches → zero page_components → rerender's `getPageSections` returns empty → page skipped, no git commit. Fix: when zero section blocks match but HTML is non-empty, store the whole fragment as one section (guarded against full documents). Key mechanism fact: the deploy path depends on `page_components.rendered_html`, not `pages.sections`; `build_status='deployed'` and file presence are independent.
- **sources:** CATALOGUE(9)#A1, running_notes_14(25)#part-7–9
- **relations:** built_from_plan_version stamp; tool-recreation-handler; dispatch throughput (masked the remaining 7 tools)
- **verify-later:** `save_page_sections_action.go` fallback; `rerender_single_page_action.go` assemblePage/getPageSections

<!-- SOURCE: U09_adoption.md -->
### Sectionless-page durability stack (2b sibling fallback + S1 discovery check + S2 flag)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Durability code WRITTEN this session (NOT yet deployed) — see runbook" (HANDOFF_2026-06-09); RUNBOOK(2) gives the deploy/verify sequence; the underlying skinner-box instance is "built and deployed… verified page_components=2".
- **what:** A page can sit in the plan with zero sections (residue from a killed build falsely marked complete) and every rebuild completes empty in ~90s because `check_has_ready_sections` ELSE routes to a SUCCESS-labelled `complete_error`. The stack: 2b — `load_page_sections_from_spec` gains a final fallback synthesising the layout from a same-role sibling (modal layout, WARN-logged, writes pages.sections; layout skeleton only, content still written from the page's own crawl); S1 — new self-registering `check_sectionless_pages` discovery check flags plan pages with empty sections that a sibling can repair and re-issues needs_content_page to page-build-handler (closed self-healing loop; relies on insertWorkItem's built-in two-strike rule for churn control); S2 — workflow-def change routing the no-sibling case to `mark_no_sections` (`fail_work_item` → needs_human_review) instead of silent success; Fix A (complete_work_item guard) is the prerequisite that stops the dispatch loop clobbering the flag.
- **sources:** RUNBOOK_section_sectionless_durability(2).md, running_notes_15(10), HANDOFF_2026-06-09, check_sectionless_pages(1).go header
- **relations:** silent-completion family; pages.sections as the build-read field; dormant checkEmptyPageSections
- **verify-later:** chassis image containing load_page_sections_from_spec_action.go + check_sectionless_pages.go; completeness-discovery-agent checks array contains "sectionless_pages"; page-build-handler else_step = mark_no_sections

<!-- SOURCE: U09_adoption.md -->
### guide as a first-class page_type (classifier vocabulary + retype + URL canonicalisation)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** "guides typed page_type=guide directly with sections=['generic-text-block']… migration_adoption_add_guide_page_type.sql worked — adoption classifier emits guide, NO post-hoc re-typing needed this run" (2026-06-05 re-adoption); retype + URL migrations APPLIED 2026-06-04.
- **what:** Guides were folded into blog-post by the analyze_site enum (source guides lived flat at /blog/guide-*.html), so `query.pages_where_type:guide` returned zero. Structural route chosen over the band-aid (query blog_post): add `guide` to the classifier enum + guidance (quote-free replace() edits on default_config), re-type the 5 content-bearing guide-* pages, and move URLs to the canonical `/guides/<slug>/index.html` (peer of tools/games; page_canonical.go already had the guide case). Earlier session had *rejected* typing guides as `guide` when the plan was flat-blog faithfulness — the later product decision flipped to canonical nesting. `pages.page_type` has only a kebab-case CHECK, no value allowlist. The `game` page_type gap (flagged in FOCUS_component_schema_patterns) was closed the same way earlier — a doc-vs-live-DB staleness lesson.
- **sources:** migration_adoption_add_guide_page_type.sql, migration_retype_guides_to_guide.sql, migration_guides_url_to_canonical.sql, running_notes_14(25)#part-14–14b, FOCUS_component_schema_patterns.md#missing-page_type-game
- **relations:** guide-list Tier-D resolution; guides faithfulness question (F1 duplicate rows); build-site-planner's staler vocabulary (preserves existing verbatim, so adopted types survive)
- **verify-later:** site-adoption-agent analyze_site enum in live def; pages.page_type values on a fresh adoption

<!-- SOURCE: U09_adoption.md -->
### Interactive fingerprint extraction (Path C: tools rebuilt as prose)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** Planned workflow insertion sketched ("extract_interactive_fingerprint (NEW)… between extract_fingerprint and check_has_external_css") with collision check, in the FOCUS_component_schema_patterns appendix / old2 README; no deployment claim anywhere.
- **what:** Adoption pulls markdown but not `<script>`/`<canvas>` interactive machinery, so crawled calculator pages rebuild as paragraphs describing the calculator. Path C plans a second fingerprint pass over the same crawl_result capturing interactive elements, feeding tool recreation. (In practice tool-recreation-handler + A1 fixes got real tools deploying from recreate prompts; the fingerprint step itself was never built.)
- **sources:** FOCUS_component_schema_patterns.md#appendix, old2/README.md, FOCUS_adoption_fidelity_and_variants.md#problems-ranked (#3)
- **relations:** tool-recreation-handler; adoption fidelity problems ranked
- **verify-later:** site-adoption-agent workflow for any extract_interactive_fingerprint step (expected absent)

<!-- SOURCE: U09_adoption.md -->
### Adoption resume logic (never built)
- **category:** adoption-pipeline
- **status-signal:** abandoned
- **status-evidence:** "orchestration_states.collected_data already persists per-step output (378KB survived a failed run), but ResumeWorkflowTopic has no subscriber — resume was anticipated, never built. User: 'a new crawl is fine.'" (FOCUS_adoption_fidelity_and_variants, deferred list).
- **what:** Mid-workflow resume of a failed adoption from persisted collected_data. The plumbing half-exists (state persistence, a resume topic) but no consumer; the accepted operational answer is re-crawl/re-adopt. Interacts with the fetch_primary_css hard-fail trade-off (a surviving state isn't reusable without resume).
- **sources:** FOCUS_adoption_fidelity_and_variants.md#deferred, old2/HANDOFF_2026-04-22#resume-logic
- **relations:** error_step fix reduced the need (CSS timeout no longer fatal); teardown+re-adopt operational pattern
- **verify-later:** ResumeWorkflowTopic subscribers (expected none)

<!-- SOURCE: U12_docs024_archives.md -->
### Single-agent adoption trigger (positional domain, no orchestrator wrapper)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** `007_adoption_pipeline_v2_april26.md`: a dedicated `site-adoption-agent` triggered directly via `./trigger-adopt-site.sh gamedesign.uk`; the patch rewrites this into "Two agents, one thin wrapper," documented fully in live `007_adoption_pipeline_v4.md`.
- **what:** Adoption originally ran as one agent invoked directly by a shell script with a positional domain argument, mixing "site being crawled" and "site being built" into a single identifier. Replaced by a thin `site-adoption-orchestrator` (spawn → call → complete) that spawns `site-adoption-agent` as its own K8s Job, and a JSON trigger payload separating `target_url` (crawl source) from `destination_domain` (site being built) — while keeping the old `url`/`domain` shape as legacy-compatible input.
- **sources:** old/older1/007_adoption_pipeline_v2_april26.md#"The adoption agent", #"Adoption modes"; old/older1/007_adoption_pipeline_v2.patch.diff; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"The adoption agent", #"Source vs destination"
- **relations:** "every pod-running agent needs a parent that spawned it" (development-guide)
- **verify-later:** confirm `site-adoption-orchestrator` agent_definitions row exists and `trigger-adopt-site.sh` uses the JSON payload shape today.

<!-- SOURCE: U12_docs024_archives.md -->
### Unified `design` spec aspect for adopted sites (superseded precursor)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** `007_adoption_pipeline.md` (v1): single `design` spec aspect; live `007_adoption_pipeline_v4.md` has a dated addendum, "Design Fingerprint Pipeline (added 2026-04-12)," documenting the two-aspect replacement.
- **what:** The earliest adoption design captured only one `design` spec aspect, generated by the LLM alongside identity/structure classification — a single blended palette-and-typography guess with no separation between what the source site actually used and what the new site should aim for. Replaced by the `design_reference`/`design_intent` split (see merged entry above).
- **sources:** old/older1/007_adoption_pipeline.md#"What gets stored where"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Design Fingerprint Pipeline (added 2026-04-12)"
- **relations:** design_reference/design_intent spec-aspect split (its replacement)
- **verify-later:** check `site_specs` for any legacy rows with `aspect='design'` from pre-2026-04-12 adoptions never migrated.

<!-- SOURCE: U12_docs024_archives.md -->
### Two-stage adoption processing (LLM classifies, Go extracts) → three-stage
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Archive header: "Two-stage processing: LLM classifies, Go extracts." Live header: "Three-stage processing: Go extracts design, LLM classifies, Go extracts content."
- **what:** Early adoption split work into just two stages — lightweight LLM classification from page summaries, then Go-only content extraction. The later design inserts a Go-only design-fingerprint extraction stage (colours/fonts/CSS vars/layout via goquery) ahead of LLM classification, on the principle "don't ask an LLM to read hex values when a regex can do it."
- **sources:** old/older1/007_adoption_pipeline.md#"Two-stage processing"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Three-stage processing"
- **relations:** unified design spec aspect (above), design_reference/design_intent split
- **verify-later:** `extract_design_fingerprint_action.go`/`enrich_fingerprint_with_css_action.go` existence and wiring.

<!-- SOURCE: U12_docs024_archives.md -->
### Section recipes for adoption (purpose + structure + reference implementation)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4: Requirement-Driven Components (longer term)" in the 2026-04-11 plan; no confirmation of shipping in any later doc reviewed.
- **what:** When adopting a site, each section would be captured as a "recipe": purpose, structure, reference implementation (guide not spec), and component match. Recipes without a good match would generate `needs_new_component` work items where the recipe becomes the build brief.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale (4)", #"Phase 4"
- **relations:** component selector by functional requirement; needs_new_component work items
- **verify-later:** whether any adoption workflow step produces structured "recipes" today.

<!-- SOURCE: U12_docs024_archives.md -->
### Adopt-from vs deploy-to separation (unbuilt)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "discussed but not implemented. Options: snapshot to S3, stage to subdomain, or store crawl artifacts."
- **what:** Unbuilt idea for a staging area distinct from the live deploy target, so a freshly-adopted rebuild could be reviewed before overwriting production. Workaround at time of writing was manual: pause work items, verify specs, unpause.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Architecture Decisions Made (item 6)"
- **relations:** site snapshots and revert (014); design fingerprint extraction pipeline
- **verify-later:** whether any staging/subdomain mechanism exists for adoption today.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Adoption interactivity misroute — canonical-prefix key desync (M2 root cause, T1)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1/§2: "Deployed: T1 (apply_adoption_plan_action.go — canonical-keyed buildPageFeatureMap) ... Both are in production."
- **what:** `apply_adoption_plan_action.go` routes adopted pages by interactivity (`len(page.Features) > 0` → `needs_tool_recreation`; else → `needs_content_page`). `buildPageFeatureMap` keyed its feature map by the raw adoption-LLM page key, but the routing loop looked up the canonicalised name via `datahelpers.CanonicalisePage`, whose `tool` branch adds a `tool-` prefix while its `game` branch preserves an already-present `game-` prefix. Result: every tool page missed the lookup (empty Features → static content route); games matched only by coincidence. Fix: key `buildPageFeatureMap` by the canonical name.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.6,§2.7,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2
- **relations:** Abandoned "no owner" claim; Post-adoption detection check (T2); Canonicalise tool page identity (T3); Recreation-loss defect
- **verify-later:** confirm in production that `buildPageFeatureMap` still contains the canonical-keyed version — HANDOFF flags a parallel adoption chat may have re-edited the file since merge

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool routing fix deployment status (T1 + T2 in production; symptom fix unconfirmed)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Deployed: T1 ... and T2 ... Both are in production." / "Not confirmed: that widgets now actually deploy ... Hold the trigger. Do not mass-emit needs_tool_recreation yet" (HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1)
- **what:** The authoritative status record for the tool-widget-clobber investigation as of 2026-05-26: T1 (routing fix) and T2 (detection check) are both confirmed deployed to production, with defined acceptance criteria for calling the deploy complete (every tool/game page has_widget=t; a deployed tool page renders an interactive widget in-browser; T2 finds nothing new on a steady-state run; no duplicate pages). None of those criteria were confirmed met at time of writing — the recreation-loss defect remains open and blocking.
- **sources:** tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1,§6,§7
- **relations:** Adoption interactivity misroute (T1); Post-adoption detection check (T2); Recreation-loss defect
- **verify-later:** re-run the exact acceptance-criteria queries against current gamesdesign.co.uk state

<!-- SOURCE: U14_docs019_runbooks.md -->
### Adoption-first fidelity inversion
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "Q4 ANSWERED — adoption does NOT call the classifier; the lean is inverted … adoption writes the specs FIRST; the classifier, when the relay later reaches it, CONSUMES them under its fidelity rules"; Q7 answered in §B3.
- **what:** How site adoption meets the relay: site-adoption-agent does the heavy work (firecrawl 30 pages, no-LLM design/interactive fingerprints, three LLM analyses — site analysis, archetype snapshot with improvement-loop constraints, content-direction guide), writes specs + pages + work items, then hands off needs_domain_research into the relay; the classifier's adoption-fidelity block treats adopted identity/archetype/content_direction/design_intent as ground truth outranking its own search+scrape. apply_adoption_plan writes site_archetype reading from collected data regardless of declared input_fields.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2; docs019/RUNBOOK_builder_route(21).md#B3 (Q7)
- **relations:** work-item relay spine; vertical-exemplar researcher (adopted sites run the hop too)
- **verify-later:** site-adoption-agent workflow; check_adoption_skip_scrape branch; classifier fidelity prompt block

<!-- SOURCE: U15_docs019_running_notes.md -->
### Adoption pipeline consumption of vertical/exemplar research
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Adoption (user orientation): classifier CONSUMES adoption specs is CONFIRMED from its own workflow (skip-scrape conditional on site_archetype...)" (v4(39), 2026-07-04); "Q4: adoption never calls the classifier... Lean is INVERTED: adoption writes first, classifier consumes under fidelity rules later." (v4(39)).
- **what:** Clarifies the actual (initially misunderstood) relationship between the site-adoption pipeline and the domain-research-classifier: adoption crawls and fingerprints the target site first, writes specs/pages/work items via `apply_adoption_plan`, then hands off to the relay (`needs_domain_research`) — the classifier consumes adoption's output under fidelity rules, rather than adoption calling the classifier directly as first assumed.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B2 read" entry.
- **relations:** Work-item relay / builder-generations architecture; vertical-exemplar-researcher hop.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Adoption writes first; classifier consumes (the corrected lean)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** README_flows: "your instinct was half right, inverted … apply_adoption_plan writes the specs, pages, and work items itself — it never calls domain-research-classifier."
- **what:** The adoption orchestrator is a thin spawn→call wrapper; the agent crawls via firecrawl, extracts design/interactive fingerprints without an LLM, runs three LLM analyses (site structure; archetype snapshot with improvement-loop constraints; content-direction style guide), and apply_adoption_plan writes specs/pages/work items directly. The classifier later consumes adopted specs under fidelity rules when the relay reaches the site. Parked question: does apply_adoption_plan write site_archetype (classify_archetype's output isn't in its declared inputs).
- **sources:** README_flows.md
- **relations:** relay spine; classifier consolidation queue
- **verify-later:** apply_adoption_plan_action.go; site_archetype writer

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption path — vertical-slice dogfooding of the ratchet
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §8.1 "Vertical slice, not horizontal layer … Dogfood the ratchet … First capability = writing Go actions"
- **what:** Walk one capability (writing Go actions) end-to-end through route→produce→verify→gate→feedback before generalising; each new machinery piece starts at `confirm-every` and graduates on evidence. Phases 1–2 double as "improve my current chat workflow"; 3–6 are the leap to autonomy.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#8.1, ED/MASTER_autonomous_build_and_operate(4).md#8.2, ED/MASTER_autonomous_build_and_operate(4).md#8.5
- **relations:** automation ratchet; self-development coding pipeline
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Doc-tree adoption plan (constitution + tag/embedding retrieval)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_doc_tree_adoption.md header "actionable plan … without committing to the atomic rewrite, the mediator, or the routing build"; §1 "the corpus does not fit in context (~200 files, ~6.7MB, ~1.0–1.7M tokens)"
- **what:** First path to value from the doc-tree design against the current setup: Phase 1 write a tiny constitution, Phase 2 tag existing docs by concern/`applies_to` into a manifest, Phase 3 make the retrieval split real (tag-based deterministic selection for rules; existing nomic/pgvector/ollama RAG for the broad corpus), Phase 4 atomic extraction deferred/evidence-driven.
- **sources:** ED/FOCUS_doc_tree_adoption.md#4, ED/FOCUS_doc_tree_adoption.md#2, ED/FOCUS_doc_tree_adoption.md#5
- **relations:** atomic standard; mediator routing; RAG actions (existing stack)
- **verify-later:** rag_actions/nomic prefixes; proposed doc_index/standards table

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption pipeline & backend capability tiers (three-layer infra)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 007 "Phase 1 — Adoption pipeline (current)"; Phases 4–5 marked planned/future
- **what:** The platform runs in three layers (Layer 1 core factory; Layer 2 client delivery via S3 + config-driven site-api-router + client Postgres; Layer 3 framework builder), with five backend capability tiers from static+JS up to full platform. Adoption is a one-off capture, not a permanent state.
- **sources:** WM/007_adoption_pipeline_v3.md#infrastructure-separation, WM/007_adoption_pipeline_v3.md#backend-capability-tiers, WM/007_adoption_pipeline_v3.md#site-adoption, WM/007_adoption_pipeline_v3.md#principles
- **relations:** site adoption agent; design fingerprint; component selector/creator; site-api-router
- **verify-later:** site-adoption-orchestrator; site-api-router; vetcomparison export path

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Site adoption agent (crawl → fingerprint → classify → apply plan)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 "site-adoption-agent workflow (runs in the spawned pod): 16 steps … apply_adoption_plan → complete"
- **what:** A thin `site-adoption-orchestrator` wrapper spawns `site-adoption-agent` to run a 16-step workflow: firecrawl crawl, Go design-fingerprint extraction, LLM classification/archetype/content-direction/design-intent, and `apply_adoption_plan` writing specs, pages, and work items. Separates `target_url` (crawled) from `destination_domain` (built).
- **sources:** WM/007_adoption_pipeline_v3.md#the-adoption-agent, WM/007_adoption_pipeline_v3.md#three-stage-processing-go-extracts-design-llm-classifies-go-extracts-content, WM/007_adoption_pipeline_v3.md#running-an-adoption-what-to-expect-and-what-to-watch
- **relations:** wrapper-orchestrator; design fingerprint; canonicalisation (page identity); interactive parse-stage gap
- **verify-later:** apply_adoption_plan_action.go; extract_design_fingerprint_action.go; firecrawl_crawl

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption faithfulness — WriteSitePlanAction identity strip
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §9 "The corruption is in WriteSitePlanAction, not the LLM … Fix direction (not yet applied)"
- **what:** Even after a faithful adoption, `WriteSitePlanAction`'s `ValidateRoles`+`CanonicalisePage` interaction permanently strips identity for `content`/`blog_post` page_types: `ValidateRoles` derives a slug that strips `tool-/guide-/game-`/`-index`, and `CanonicalisePage` only re-adds prefixes for tool/game/guide roles — so mistyped section-index hubs flatten. Root cause is the wrong `page_type`; clean fix is upstream at adoption time.
- **sources:** WM/016_debugging_guide_v2_44.md#adoption-faithfulness-llm-convergence-are-faithful-writesiteplanaction-strips-identity-for-content-blog_post-types, WM/016_debugging_guide_v2_44.md#0, WM/ARCHITECTURAL_TENSIONS(2).md#tension-2-page-identity-is-derived-in-multiple-places-that-can-undo-each-other
- **relations:** CanonicalisePage; architectural tension #1/#2; locks (adoption_locked); FOCUS_adoption_faithfulness_via_locks
- **verify-later:** WriteSitePlanAction; datahelpers/page_canonical.go ValidateRoles/normaliseSlug; analyze_site page_type

<!-- SOURCE: U18_sql_for_agents.md -->
### site-scraper (Firecrawl scrape → site_context)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 032 definition ("Uses firecrawl_scrape action (requires patch_02_webscrape_url_field.go)").
- **what:** Scrapes a live site's homepage via the webscrape adapter (Firecrawl), then an LLM step transforms results into the site_context format webdesign-agent consumes — the original design-transfer mechanism.
- **sources:** 032_site_scraper_agent.sql
- **relations:** webdesign-agent; ancestor of site-adoption-agent's full crawl
- **verify-later:** whether site-scraper is still used vs site-adoption-agent

<!-- SOURCE: U18_sql_for_agents.md -->
### tool-recreation-handler (recreate interactive tools from crawled source)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 099 definition; live-run evidence in 138 (run e1018366 recreated bugs faithfully → prompt fix) and 132/137 (note-writing wired, subject corrected).
- **what:** Two-stage recreation of JS-heavy pages during site adoption: analyze_tool (LLM functional spec from source + context) then recreate_tool (Opus generates working replacement HTML/CSS/JS), with completeness/truncation checks, validation, save/deploy. 138 adds the "Mandatory Behaviour Requirements" prompt section rendered from spec.interactive_features which OVERRIDES the original source — fixing the observed failure where explicit spec fixes were buried in analysis JSON and Opus faithfully recreated the original bugs.
- **sources:** 099_tool_recreation_handler.sql; 138_recreate_tool_carries_spec_features.sql; 137_recreation_spec_and_note_subject.sql
- **relations:** site-adoption-agent creates its items; tool acceptance verifies results
- **verify-later:** current recreate_tool prompt; spec.interactive_features producers

<!-- SOURCE: U18_sql_for_agents.md -->
### Site adoption pipeline (site-adoption-agent + wrapper orchestrator)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 091 definition; 104 adds the wrapper (2026, "Pattern copied verbatim from med-export-orchestrator"); 115 adds the 'adoption' lock source for "faithful first pass".
- **what:** Adopts an existing live site: firecrawl_crawl via the webscrape adapter returns per-page markdown; an LLM analyze step classifies pages and extracts identity/design/content structure into a JSON plan; apply_adoption_plan creates site_specs, page records, and work items to recreate the site in-platform. 104 wraps it in a spawn→call orchestrator so the long crawl runs in its own Job pod with clean correlation logs.
- **sources:** 091_site_adoption_agent.sql; 104_site_adoption_orchestrator.sql; 115_locks.sql
- **relations:** page-content-writer Recreate Mode; tool-recreation-handler; adoption locks
- **verify-later:** apply_adoption_plan action; adoption directive writer (pending per 115)

<!-- SOURCE: U20_legacy_docs_a.md -->
### 11-agent website analysis framework (four agent groups)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Whole docs003 set is planning ("Here is the detailed, point-by-point analysis…"); docs004 explicitly reframes it ("The old numbers are meaningless now… rename them") into the Learn/Execute playbook model.
- **what:** The original web-capture master plan: Strategy & Content group (Strategist A10, Content Infuser A11), Library & Storage (Librarian A7, S3+Postgres/pgvector), Design Ingestion (Prospector A0, Site Profiler A1, Capture Bot A2/Playwright, Layout & Labeling A3 XY-Cut+LLaVA, Component Generator A4 VLM screenshot-to-code, Style Extractor A5 getComputedStyle — later eliminated in favour of Firecrawl branding data, Behavior Extractor A6 CodeLlama), Generation (Publisher A8 "Dribbble-like" showcase site, Architect A9 template builder querying by CLIP embedding). All implemented as agent_definitions rows + new action adapters, not new binaries.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md; docs003_firecrawl/README.0124.11_agent_summary.md; docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** successor chain: playbook model (docs004) → MVP site builder → current adoption-pipeline (docs 007) and site-spec-and-classifier (docs 021). Publisher A8's public design-library site was abandoned.
- **verify-later:** which of the 11 agent types ever got agent_definitions rows.

<!-- SOURCE: U20_legacy_docs_a.md -->
### website-analyzer conditional scraping group
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Tested with kcat messages (boxing-tickets.com, both basic and structured+crawl variants) and a live UPDATE of its orchestration_workflow.
- **what:** An agent group that takes target_url + flags (extract_structured, crawl_pages, crawl_limit/depth) and conditionally routes between basic scrape, structured extraction, and multi-page crawl using evaluate_condition — the first "smart" capture entry point.
- **sources:** docs003_firecrawl/README.0129.testing_webscrape_message.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** firecrawl adapter; successor: adoption pipeline crawl/classify flow.
- **verify-later:** agent_group_definitions row group_type='website-analyzer'.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Playwright capture adapter + website-capture agent
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Complete deliverables (adapter py, capture_actions.go, agent SQL: desktop/mobile viewports, hover/focus states, scroll intervals with parallax/sticky detection, asset extraction, S3 upload) — but docs004/website_analysis 002 records "Agent 5 eliminated… use Firecrawl branding data instead" and firecrawl/001 adapts the MVP away from Playwright.
- **what:** Deep browser-based capture: Playwright adapter on system.adapter.playwright.requests capturing full-page desktop + mobile screenshots, DOM, computed styles, interaction states (hover/focus for up to 50 selectors), scroll-position screenshots (0/25/50/75/100%) with parallax/sticky detection, asset extraction, and organised S3 upload with manifest. Deferred in favour of the managed Firecrawl service for MVP; the deeper capture ideas (interaction/scroll states) never resurfaced.
- **sources:** docs004_website_capture_project/playwright/website_capture_agent.sql; docs004_website_capture_project/playwright/playwright_adapter.py; docs004_website_capture_project/playwright/implementation_roadmap.md; docs004_website_capture_project/firecrawl/001claude_initial.md
- **relations:** firecrawl adapter (chosen replacement); adoption-pipeline crawling successor; behaviour capture (rrweb) idea from docs003 also abandoned.
- **verify-later:** playwright-adapter deployment existence.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Website-builder orchestrator (capture → vision → code → synthesis → content → library)
- **category:** adoption-pipeline
- **status-signal:** abandoned
- **status-evidence:** Orchestrator SQL references agent types (website-vision, website-code-analyzer, website-synthesis, content-strategist) and actions (analyze_input_type, parallel_section_generation, store_component) that are never defined or mentioned again; the MVP builder took a different shape.
- **what:** A master workflow to rebuild a site from a captured one: capture data → visual analysis (layout/palette from screenshots) → code cleaning/analysis → synthesis correlating visual+code into a template → content planning → parallel section generation → aggregate → store components with embeddings in the library. The maximal "clone-and-improve" vision.
- **sources:** docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql; docs004_website_capture_project/playwright/website_builder_integration_guide.md; docs003_firecrawl/README.0125.claude_11_agent_summary.md
- **relations:** successor in spirit: adoption-pipeline content recreation (docs 007); vision analysis resurfaces in the current image-analysis tooling.
- **verify-later:** confirm none of the four sub-agent types exist in agent_definitions.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Adopting existing external sites ("Adopt" workflow)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Designed in 004 ("Adopt workflow… status: 'adopted_partial'… match_confidence"); the taxonomy's adoption-pipeline (docs 007: site crawling, classification, content recreation) is the named live successor.
- **what:** Run the Learn loop against an existing site the platform didn't build: scrape, deconstruct layout, match found blocks to the in-house component library with confidence scores, generate a manifest marking it adopted_partial — making external sites partially manageable by agent edit workflows.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md
- **relations:** successor: adoption-pipeline (docs 007).
- **verify-later:** compare with current adoption pipeline design.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Site interrogation & pattern library
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** docs009/003: "'Interrogate' successful sites to extract... Store extracted patterns"; docs012/012 Part 3 details the 5-phase pipeline (discover → firecrawl capture → LLM structure analysis → pattern extraction → component creation) with pattern_sources table marked "(future)".
- **what:** Learning from successful sites without copying: capture HTML+screenshot, LLM-analyse section types, visual hierarchy, content strategy and psychological principles, extract reusable patterns tagged by industry/funnel-stage/audience with "why it works" notes, and mint content_components (origin_type='extracted') from them. Patterns become queryable ("for finance trust-building, use X") and feed component selection. The most persistent unfulfilled idea of this era — restated in docs009, docs010 roadmaps (Phase 4), and docs012.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-3; docs009_site_interrogation_and_solutions/003_claude_save_point.md#2; docs010_multitrack_flows_persona_architecture/018_priority_matrix.md
- **relations:** Pragmatic Evolution Engine phase 2; adoption-pipeline site crawling (current descendant); pattern_sources/captured_sites tables; component library.
- **verify-later:** pattern_sources table; origin_type/industry_tags/funnel_stages columns on content_components; website-capture-firecrawl agent.

<!-- SOURCE: U21_legacy_docs_b.md -->
### site-scraper companion agent (design context from live URLs)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** docs017/003: site-scraper "firecrawl_scrape → analyze_design → returns site_context for webdesign-agent"; standardized site_context schema with source: "database|scrape|manual".
- **what:** A standardized site_context interface (domain, company, industry, palette, typography, component functions, source) produced by either DB load, live-site scraping, or manual input, so the webdesign-agent can restyle from any source — enabling "scrape competitor → feed to design agent → apply to your site" pipelines. The schema-standardization idea matured; the scraper flow folded into the adoption pipeline.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/003_design.md#Architecture; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Standardized-Interface-Schemas
- **relations:** adoption-pipeline capture; webdesign-agent; standardized interface schemas doctrine.
- **verify-later:** site-scraper agent definition; load_site_for_design action.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### "Type guides as `guide`" — falsified as a quick companion fix, later built properly as a structural fix
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 2: "**Falsified the 'type guides as `guide`' companion.** ... source guides live flat at `/blog/guide-rng-design.html`, while a `guide` role would *nest* them... typing `guide` would be *less* faithful... Left as an open product decision; did NOT ship the wrong patch." Then Part 14 (2026-06-04), same session-log lineage: `migration_retype_guides_to_guide.sql` + `migration_guides_url_to_canonical.sql` were written, applied, and guides were deliberately moved to `/guides/<slug>/index.html` — the exact "less faithful" move earlier rejected, now chosen deliberately once `guide` became a first-class page_type with its own canonicalisation rule.
- **what:** Two-stage decision on how adopted "guide" content should be typed/URL'd. First pass: rejected retyping guides as `guide` as a *quick fix* for a de-prefixing side-effect, because it would misplace the URL relative to the untouched source. Second pass, as a *deliberate structural project*: added `guide` to the page_type enum, re-typed the 5 real guide-* pages, added the classifier's default_config guidance, and migrated their URLs from `/blog/guide-*.html` to `/guides/<slug>/index.html` — closing the exact gap the earlier rejection had flagged as "an open product decision."
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 2, 13, 14, 14a–14h
- **relations:** SyncPagesToDBAction canonicalisation divergence; bare-guide-duplicate defect (below); adoption-faithfulness-via-locks (below)
- **verify-later:** `pages.page_type` enum and `page_canonical.go`'s `guide` case in current code; whether `build-site-planner`'s vocabulary was ever updated to include `guide` for new adoptions.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Bare-guide duplicate pages — root cause: planner ignores adopted state (prompt-rule gap, not wiring)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 14e: "DECISIVE (llm_call_log plan_site)... `saw_guide_pages=t`, `prompt_says_no_existing=f`, `planned_bare_in_response=t`... So the planner WAS given the adopted guides and emitted `economy-basics` anyway → PROMPT-RULE gap... NOT a wiring/status gap." Cleanup migration (`migration_cleanup_bare_guide_duplicates.sql`) applied and confirmed durable (Part 14f: "current-plan bare-name query returns 0 rows").
- **what:** `build-site-planner` re-invents a differently-slugged sibling page (`economy-basics`) for a topic already adopted under a prefixed name (`guide-economy-basics`), because its "never duplicate an existing page" prompt rule only named games/tools examples and didn't generalise to the `guide-` prefix pattern. This is a fresh, concretely-diagnosed instance of the previously-documented `FOCUS_planner_ignores_adopted_state.md` mechanism (2026-05-19). A durable Go-level fix (deterministic topic-stem collision guard in `validate_site_plan`/`write_site_plan`, reusing `CanonicalisePage`'s prefix-stripping) was recommended but not shipped in this arc; only the data cleanup + an optional prompt-rule stopgap were delivered.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 14c–14g
- **relations:** adoption-faithfulness-via-locks convergence (below); "type guides as guide" (above)
- **verify-later:** `FOCUS_planner_ignores_adopted_state.md`; whether the Go-level topic-stem guard was ever built.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### A1 — tool/game pages never deployed: root-cause hypothesis evolution
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Original catalogue (base, 2026-05-28): "*Tentative area:* the gap is between 'page row built/complete in DB' and 'file committed to git'. Could be the rerender/deploy path skipping nested child pages, the tool-recreation handler not producing a deployable artefact, or a git-path mismatch." Catalogue `(4)` (2026-06-04): "*Cause pinned (two coordinated root causes):* 1. Parser... `saveSectionsExtractFromHTML` extracts only `<section>` blocks, but `tool-recreation-handler`'s prompt emits `<div class="tool-page">` (no `<section>`)... 2. Flip churn: `upsertPage`'s ON CONFLICT flipped `deployed → needs_rebuild`."
- **what:** The site's actual interactive product (tools/games) never deployed a file despite `pages` rows and `complete` work items. The three original hypotheses (deploy-path bug, handler artefact-production bug, git-path mismatch) were all superseded by two pinned, source-confirmed causes: an HTML-fragment parser that only recognises `<section>` blocks (tool output uses `<div class="tool-page">`, so it silently extracted zero sections), plus the ON CONFLICT flip churn (above). Fix: single-fragment fallback in the parser + the Option B stamp/flip removal. Verified end-to-end on a subsequent adoption run (all 5 games + tools deployed with working links).
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md; adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md; running_notes_14(20) Parts 7–10
- **relations:** deployed→needs_rebuild flip (above); dispatch throughput bottleneck (Family J, below)
- **verify-later:** `save_page_sections_action.go` current parser logic.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### tool-page canonicalisation misroute (adoption Features key desync)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "RESOLVED 2026-05-26 → b1" (root cause confirmed via query G) but the numbered "Potential solutions" section proposes the actual code fix as still to be applied — diagnosis complete, fix landing unconfirmed from the doc alone.
- **what:** A specific, resolved-diagnosis bug postmortem: an adopted `page_type='tool'` page deploys with prose describing the tool but no interactive widget. Two distinct causes were disambiguated (M1 — a widget existed and a later `SavePageSectionsAction` rebuild deleted it via a text-only regression guard blind to script-heavy content; M2 — no widget was ever generated because adoption captures text but has no JS-parse stage). Root cause for the gamesdesign.co.uk case was M2, but *not* because generation is unowned — `tool-recreation-handler` exists and should have run. The actual fault: `apply_adoption_plan` routes by `len(page.Features)`, but `buildPageFeatureMap` keys its map by the **raw** page name the adoption LLM wrote, while the routing lookup uses the **canonicalised** name (`CanonicalisePage` prepends `tool-` for tools) — so every tool page's feature lookup misses even when the LLM correctly detected interactivity, silently misrouting it to the static `page-build-handler` path instead of `tool-recreation-handler`. Games pages don't hit this because their canonical prefix (`game-`) already matches the raw key.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/016_debugging_guide_addendum_adopted_tools_no_widget(3).md
- **relations:** doc 029 (CanonicalisePage, Phase-0 helper); component-regeneration-flow (SavePageSectionsAction clobber path); spurious-duplicate-pages pattern (below, same "adoption vs. a second surface" family)
- **verify-later:** buildPageFeatureMap in the adoption/orchestration action code — confirm whether the canonical-key fix was applied

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### spurious duplicate pages from "planner ignores adopted state"
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** The migration is a real, executed cleanup with before/after verification queries and an explicit rollback path via `_bak_del_bare` snapshot tables, dated against a specific incident (gamesdesign.co.uk, created 2026-06-03 20:25:30, cleaned up in this migration).
- **what:** A confirmed, named failure pattern: a post-adoption planner pass (`build-site-planner`/`blog-content-planner`) invents new `page_type='blog-post'` pages (`sections=[]`, `build_status='planned'`, never rendered) that duplicate content already faithfully recreated by adoption as `page_type='guide'` pages at a different URL — "a second surface invents parallel pages after adoption" because the planner doesn't check adopted state before generating its own content plan. The cleanup migration is durable — it removes the bare pages from the pages table, the *current* `site_plan_pages`/`site_plan_sections` (so the reconciler won't recreate them), and terminalises the dangling `site_work_items` rows (which have no FK to pages and would otherwise linger holding a dedup slot) — but explicitly does not fix the upstream planner logic that would reintroduce the same duplicates on a future `plan_site` run.
- **sources:** docs/_archive/agent_docs/sql_for_tables/040b_migration_cleanup_bare_guide_duplicates(1).sql
- **relations:** tool-page canonicalisation misroute (above, same "adoption vs. second surface" bug family); FOCUS_planner_ignores_adopted_state.md; doc 029; site-plan-and-reconciler
- **verify-later:** FOCUS_planner_ignores_adopted_state.md, whether the upstream planner prompt/logic has since been tightened to check adopted state first

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Infrastructure three layers (core platform / client delivery / framework builder)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 007: Layer 1 exists; Layer 2 "planned" beyond static serving; Layer 3 "future"
- **what:** Layer 1 (K8s factory) never serves external traffic — it produces artifacts and pushes them. Layer 2 = client delivery (S3 static now; site-api-router + client Postgres on OVH VMs planned, config-driven routes reusing the action library). Layer 3 = provisioning agent frameworks for clients. Backend capability tiers 1–5 (static→full platform) with static-first principle; vetcomparison JSON-on-S3 pattern handles up to ~50k items (Pagefind extends to ~500k).
- **sources:** 007#Infrastructure Separation, #Backend Capability Tiers, #Site API Router
- **relations:** dynamic applications tiers (022); P1 marketing/OpenClaw
- **verify-later:** any site-api-router code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption is a one-off capture, not a ceiling (specs separation)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 v4 principles + what-gets-stored table; patch updates to Phase-1 split
- **what:** Crawl data → research_results (adoption_crawl/adoption_page), never site_specs. Spec aspects: identity (with adopted_from provenance), design_reference (concrete extracted values, historical, never modified), design_intent (semantic brand-level brief, survives plan rebuilds post-030), site_archetype (character + inviolable constraints), content_direction (brand voice), structure. Webdesign reads intent not reference; evolution = update intent. The strategist then writes aspiration beyond the adopted baseline.
- **sources:** 007#Site Adoption, #What gets stored where (incl. patch revision); 004#Adopted Sites
- **relations:** 028 fidelity; 030 strategic-vs-plan-time split
- **verify-later:** site_specs aspects on an adopted site

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption agent three-stage processing (Go fingerprint / LLM classify / Go content+plan)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 16-step workflow with runtime expectations and success signals
- **what:** site-adoption-orchestrator (wrapper) spawns site-adoption-agent: firecrawl crawl → Go extract_design_fingerprint (goquery over rawHTML: hexes, fonts, CSS vars, layout patterns, Google Fonts; external CSS fetched and merged; suggested_mapping to our var names) → LLM analyze_site on ~500-char page summaries + classify_archetype + derive_content_direction + generate_design_intent → Go apply_adoption_plan (buildCrawlPageIndex, page records, per-page markdown to research_results, design_reference spec, work items). Principle: LLM for reasoning, Go for extraction — never pay an LLM to transcribe.
- **sources:** 007#The adoption agent, #Three-stage processing, #Design Fingerprint Pipeline
- **relations:** wrapper pattern; classifier handoff
- **verify-later:** extract_design_fingerprint_action.go; adoption agent definition

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Source vs destination separation in adoption (target_url / destination_domain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 modes table + operational signals ("ApplyAdoptionPlanAction: source differs from destination")
- **what:** Adoption separates the crawled reference site from the site being built; ensure_site_record uses destination via domain_override_field, crawl hits target_url, provenance records source host — which also keys the crawl-content lookup (mismatch silently drops all adopted content). Legacy single-domain shape still accepted.
- **sources:** 007#Source vs destination, #Adoption modes
- **relations:** build-inspired-by mode (goes via classifier instead)
- **verify-later:** apply_adoption_plan source/destination handling

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption → classifier handoff (needs_domain_research; no shortcuts)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 section + patch; grounded in 028's "classifier always runs in full"
- **what:** Adoption queues exactly one strategic item, needs_domain_research; it does NOT queue needs_composition/needs_design directly — the cascade produces them naturally. The classifier reads site_archetype/design_reference as ground truth, read-and-extends identity/content_direction/design_intent, always runs vertical research, emits classification, queues needs_strategy. Post-030: planner writes to plan-domain tables and the reconciler emits page items; the planner's job ends at "the new plan is durably current".
- **sources:** 007#Handoff to the classifier, #Post-adoption (patched version); 028#The classifier is the strategic brain
- **relations:** unified pipeline; reconciler
- **verify-later:** adoption work-item emissions in apply_adoption_plan

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Pattern extraction, code-as-reference, and RAG-fed generation
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** 007 Phase 3 items; "Runs as a side effect… patterns accumulate" future tense
- **what:** A pattern-extraction-agent mines research into reusable tool specs, layout/content patterns, and good/bad UX examples; complex tool builds include reference code in the prompt with explicit original-implementation instruction (never deployed directly — copyright stance); prompt+output pairs feed RAG so future generations retrieve both abstract specs and concrete prior successes.
- **sources:** 007#Research, Patterns, and the Component Library
- **relations:** knowledge_base; tool-recreation-handler
- **verify-later:** none built

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Design fingerprint extraction pattern (adoption parse stage for design)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** described as "the proven template" (interactive FOCUS, sequencing locked 2026-05-14); exercised live in 04-23 cascade (crawl → fingerprint → fetch CSS → enrich → analyze)
- **what:** Firecrawl rawHtml parsed Go-side (goquery) by extract_design_fingerprint for colours/fonts/CSS vars/layout/dark sections; external CSS fetched via firecrawl_scrape and merged (EnrichFingerprintWithCSSAction); an LLM step (generate_design_intent) produces the semantic brief; stored as design_reference (concrete) + design_intent (semantic) spec aspects. The template any other parse-stage extractor copies.
- **sources:** FOCUS_interactive_content_generation(4).md#Adoption; HANDOFF_2026-04-23(1).md#validated
- **relations:** interactive fingerprint (clone of this pattern); site-design-planner palette source
- **verify-later:** extract_design_fingerprint action; design_reference/design_intent aspects on adopted sites

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption source/destination separation and variant axis
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** FUTURE doc (2026-04-20) "Status: Future work … Option 1 sketched"; but 2026-04-23 handoff triggers "the standard adopt-separated kcat command SOURCE_URL=… DEST_DOMAIN=…" — Option 1 evidently landed
- **what:** Decouple the crawled target from the built destination (target_url + destination_domain inputs; ensure_site_record override) and gate spec-writing on an adoption_variant: reference (design only), structure (+archetype/pages), clone (+content_direction — old behaviour), analysis (aggregate competitor_landscape). Phase 2: sites.source_site_id provenance; Phase 3: adoption_references library. Risks: variant-gated data bleed, typo'd destination domains creating junk sites.
- **sources:** FUTURE_adoption_source_destination_separation.md (whole); HANDOFF_2026-04-23(1).md#priority-4
- **relations:** adoption faithfulness (fidelity axis); duplicate-sites-row question
- **verify-later:** apply_adoption_plan variant gating; extractDestinationDomain in ensure_site_record_action.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption → classifier handoff (classifier as strategic brain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** apply_adoption_plan rewrite "NOT YET DEPLOYED" 2026-04-23, but 2026-05-26 verifies "Adoption does not bypass the planner — it routes through it via the strategy→briefing→site_plan chain, as 007_adoption_pipeline_v4.md intended"
- **what:** Adoption stops queueing needs_composition/needs_design directly; it writes its specs and emits a single needs_domain_research item so the classifier (with dynamic taxonomy) then strategist → briefing → planner run for adopted sites exactly as for fresh builds — doc 028's ownership model applied to adoption.
- **sources:** HANDOFF_2026-04-23(1).md#not-deployed; HANDOFF_2026-05-26…md#verified
- **relations:** dynamic taxonomy classifier; spec ownership contract; pipeline cascade
- **verify-later:** apply_adoption_plan_action.go current emissions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Firecrawl capability escalation ladder (executeJavascript, waitFor, structured json)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "These are upgrades, not prerequisites" (interactive FOCUS)
- **what:** When plain rawHtml + external-fetch parsing misses dynamically-injected scripts or bundled logic, Firecrawl's executeJavascript actions (script inventory via querySelectorAll), waitFor, and schema-driven json extraction are the escalation path for the parse stage.
- **sources:** FOCUS_interactive_content_generation(4).md#Firecrawl-features
- **relations:** interactive fingerprint C1-C6
- **verify-later:** firecrawl adapter capabilities used today

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Duplicate sites-row on re-adoption (open investigation)
- **category:** adoption-pipeline
- **status-signal:** unknown
- **status-evidence:** item 20 (2026-04-23, version (1) only): "Couldn't confirm … worth checking on next adoption run"
- **what:** Suspicion that adopting a destination_domain that already has a sites row creates a second row, leaving orphan work items pointing at the stale row while a new cascade runs against the other. Decision needed: refuse when destination exists vs reuse as refresh; duplicate-creation is the worst option.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 20
- **relations:** source/destination separation; library-row cleanup
- **verify-later:** ensure_site_record behaviour on existing domain

---

## Proposed NEW categories

| slug | why |
|---|---|
| NEW:work-dispatch | The detected→triaged→claimed state machine, dispatch chain, claim blockers/timeouts, pipeline label, two-strike rule and silent-completion semantics form a coherent expert domain that spans (and is not owned by) improvement-loop or scheduler-and-tasks. 10 concepts landed here. |
| NEW:prompt-composition | Prompt architecture (mega-prompt fragility, envelope/tool-call/validation patterns, parameter-shaping for images) is design-of-prompts, distinct from llm-quality-testing's evaluation focus. |
| NEW:language-i18n | Language/i18n surfaces (implicit language mechanism, hardcoded-English map, lang attribute, soft statics) have no home in the seed taxonomy. |

<!-- SOURCE: U04_idea_uk.md -->
### Fresh vs adoption entry paths converge on one cascade (fresh = adoption minus the crawl)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Capability map table with per-row verdicts ("already in fresh"); resolved empirically 2026-06-14 — a fresh submit flowed end-to-end through dispatch without manual triage.
- **what:** Two entry agents — domain-submitter (fresh: {domain,email,mission_brief}) and site-adoption-orchestrator (adopt: crawl→fingerprint→archetype→seeds) — converge on needs_domain_research and share the whole cascade (classifier read-and-extends adopted seeds → strategist → briefing → planner → composition → design → pages → rerender). The capability map shows the only adoption capabilities fresh lacks (CSS fingerprint, interactive-feature detection, full archetype) are inherently crawl-products; a new "fresh-build" single-agent copy was rejected as premature — reuse the existing path. The unified trigger `082_submit_domain_unified.sh` picks the entry (--from ⇒ adopt) and gained `--mission-file` (used to ship idea.uk's mission). The richest "seed with the existing setup" is adoption pointed at the live site.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (capability map + submission entry points); idea.uk/HANDOFF(13).md (pipeline graph)
- **relations:** Phase 0 read; fidelity dial; adoption teardown vs fresh detach.
- **verify-later:** 082 script; site-adoption-orchestrator definition.

<!-- SOURCE: U04_idea_uk.md -->
### Fidelity dial — documented but not wired
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "The trigger records --fidelity but it doesn't yet modulate the build… doc 028's explicit fidelity input and its build_policy/adoption_meta aspect and per-item status are not yet wired (fidelity is currently implicit high)."
- **what:** A planned locked/high/medium/low fidelity input governing how faithfully a build reproduces its source/spec, flowing into a build_policy aspect with per-item planned/deployed status. Today the unified trigger records the flag and nothing reads it — a clean doc-vs-reality gap flagged repeatedly in the handoffs.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (launch idioms); idea.uk/HANDOFF(13).md
- **relations:** doc 028 (design/adoption); fresh vs adoption convergence.
- **verify-later:** any consumer of the fidelity field.

<!-- SOURCE: U05_content_quality_linking.md -->
### Readopt-as-acceptance-test pattern
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §7c "only after §6 passes"; FOCUS_content_quality(2) work order 3 — planned, not yet run in these docs.
- **what:** After a fix batch verifies on the existing site, tear down and re-adopt the source (gamedesign.uk → gamesdesign.co.uk) as the from-scratch acceptance test and the fresh content-quality baseline — any failure then attributable to the virgin path. Expected recurrences are pre-declared (adopt-path defects untouched by the linking work: brand-suffix titles, tool-flavoured guide copy, empty descriptions, footer metadata) and are the next package's input, not regressions. Corollary discipline: site_id changes on every teardown — always resolve via domain.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#7c; running_notes_17(21).md#readopt-decision; FOCUS_content_quality(2).md#work-order
- **relations:** content-quality catalogue; adoption pipeline; debugging heuristics (site_id).
- **verify-later:** whether the readopt ran post-2026-06-26; new site_id baseline audit results.

<!-- SOURCE: U09_adoption.md -->
### Site-adoption agent pipeline (crawl → fingerprint → analyze → classify → apply)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Workflow traced in FOCUS_component_schema_patterns appendix ("crawl_site (firecrawl_crawl…) → format_crawl → check_crawl_content → extract_fingerprint → check_has_external_css → fetch_primary_css → enrich_fingerprint → analyze_site"); repeated verified gamesdesign adoption runs through 2026-06-05.
- **what:** The `site-adoption-agent` crawls a source site (Firecrawl, markdown + rawHtml, limit 30), extracts a design fingerprint (CSS variables, typography, palette), enriches with fetched primary CSS, classifies pages via `analyze_site` (LLM emits per-page page_type + url), derives content direction and archetype, generates a `design_intent` spec, and `apply_adoption_plan` writes pages + specs. It writes pages and specs but no site_plan — the build cascade (build-site-planner etc.) runs later.
- **sources:** FOCUS_component_schema_patterns.md#appendix, FOCUS_adoption_fidelity_and_variants.md#the-core-gap, HANDOFF_2026-05-25#standing-context
- **relations:** adoption variants, adoption fidelity dial, wrapper-orchestrator pattern, guide/game page_type vocabulary
- **verify-later:** agent_definitions row `site-adoption-agent` (id 4e2d8e8e…), `apply_adoption_plan_action.go`, `extract_design_fingerprint_action.go`, `enrich_fingerprint_with_css_action.go`

<!-- SOURCE: U09_adoption.md -->
### Adoption source/destination separation (target_url vs destination_domain)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** "Phase 1 plumbing (`target_url` / `destination_domain` separation) is deployed" (FOCUS_adoption_fidelity_and_variants); "source/destination separation deployed and working".
- **what:** Adoption previously conflated source and destination into one site_id. Phase 1 parameterised it: `target_url` = what to crawl, `destination_domain` = what to build (legacy url/domain still accepted), via migrations 001–004 plus the `sourceDomain` vs `domain` fix in `apply_adoption_plan_action.go` that was silently dropping page content when source ≠ destination.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#what-phase-1-deployed, old2/HANDOFF_2026-04-22
- **relations:** adoption variants (the selector this separation was built for)
- **verify-later:** `EnsureSiteRecordAction`, `apply_adoption_plan_action.go` ~52169-52176

<!-- SOURCE: U09_adoption.md -->
### Adoption variants A–D and the unwired variant selector
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "the variant selector was never wired — everything defaults to the current behaviour, which sits roughly between A and C and commits to neither… Variant C is what's needed and does not yet exist in a meaningful sense" (FOCUS_adoption_fidelity_and_variants).
- **what:** Four adoption operations defined in FUTURE_adoption_source_destination_separation: A reference-only (design inspiration), B design+structure (same pages, your content), C full clone (copy everything, rename), D multi-source analysis. Plumbing exists but no selector; the current pipeline produces "a site-planner brief plus specs, not a deterministic copy" — the gap between specs+LLM interpretation and the actual source site.
- **sources:** FOCUS_adoption_fidelity_and_variants.md#the-adoption-variants, FOCUS_design_composition_flow_and_adoption_fidelity(1).md#4
- **relations:** fidelity dial (orthogonal axis: variant = what the operation is, dial = how much aspiration); faithful-first-pass locks
- **verify-later:** adoption workflow input schema; any `variant`/`clone` parameter in site-adoption-orchestrator config

<!-- SOURCE: U09_adoption.md -->
### Adoption fidelity dial (locked/high/medium/low; phases 1–4)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Implementation status (the catch). Only Phase 1 exists: an adoption-aware classifier prompt giving implicit `high` fidelity at the prompt level… today fidelity is coarse prompt behaviour, not the deployed-vs-planned model" (FOCUS_design_composition_flow §4, 2026-05-26).
- **what:** Unifying idea: every input (bare domain, questionnaire, scraped live site) is the same thing at different fidelity — adoption is the high-fidelity end of one pipeline. The dial (locked/absolute, high, medium, low; re-purposed as research-confidence for blank sites) governs how much aspiration reaches the first build and how fast the improvement loop narrows the gap. Real dial needs Phase 2 (per-item status on specs), Phase 3 (explicit `build_policy`/`adoption_meta` input), Phase 4 (classifier produces status-marked aspiration alongside faithful baseline).
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#4, FOCUS_adoption_fidelity_and_variants.md
- **relations:** doc 028 platform mission; timed locks are the enforcement; variant axis
- **verify-later:** classifier prompt for adoption-aware fidelity; site_specs per-item status columns (Phase 2 — expected absent)

<!-- SOURCE: U09_adoption.md -->
### Tool/game pages never deployed (A1): save_page_sections `<section>`-only parser
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** "A1 VERIFIED CLOSED… all five games committed… tools deploy… The three-file fix (parser fallback + deploy-time stamp + flip removal) is confirmed in production" (2026-06-03).
- **what:** `saveSectionsExtractFromHTML` extracted only `<section>…</section>` blocks, but tool-recreation-handler emits `<div class="tool-page">…` — zero matches → zero page_components → rerender's `getPageSections` returns empty → page skipped, no git commit. Fix: when zero section blocks match but HTML is non-empty, store the whole fragment as one section (guarded against full documents). Key mechanism fact: the deploy path depends on `page_components.rendered_html`, not `pages.sections`; `build_status='deployed'` and file presence are independent.
- **sources:** CATALOGUE(9)#A1, running_notes_14(25)#part-7–9
- **relations:** built_from_plan_version stamp; tool-recreation-handler; dispatch throughput (masked the remaining 7 tools)
- **verify-later:** `save_page_sections_action.go` fallback; `rerender_single_page_action.go` assemblePage/getPageSections

<!-- SOURCE: U09_adoption.md -->
### Sectionless-page durability stack (2b sibling fallback + S1 discovery check + S2 flag)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Durability code WRITTEN this session (NOT yet deployed) — see runbook" (HANDOFF_2026-06-09); RUNBOOK(2) gives the deploy/verify sequence; the underlying skinner-box instance is "built and deployed… verified page_components=2".
- **what:** A page can sit in the plan with zero sections (residue from a killed build falsely marked complete) and every rebuild completes empty in ~90s because `check_has_ready_sections` ELSE routes to a SUCCESS-labelled `complete_error`. The stack: 2b — `load_page_sections_from_spec` gains a final fallback synthesising the layout from a same-role sibling (modal layout, WARN-logged, writes pages.sections; layout skeleton only, content still written from the page's own crawl); S1 — new self-registering `check_sectionless_pages` discovery check flags plan pages with empty sections that a sibling can repair and re-issues needs_content_page to page-build-handler (closed self-healing loop; relies on insertWorkItem's built-in two-strike rule for churn control); S2 — workflow-def change routing the no-sibling case to `mark_no_sections` (`fail_work_item` → needs_human_review) instead of silent success; Fix A (complete_work_item guard) is the prerequisite that stops the dispatch loop clobbering the flag.
- **sources:** RUNBOOK_section_sectionless_durability(2).md, running_notes_15(10), HANDOFF_2026-06-09, check_sectionless_pages(1).go header
- **relations:** silent-completion family; pages.sections as the build-read field; dormant checkEmptyPageSections
- **verify-later:** chassis image containing load_page_sections_from_spec_action.go + check_sectionless_pages.go; completeness-discovery-agent checks array contains "sectionless_pages"; page-build-handler else_step = mark_no_sections

<!-- SOURCE: U09_adoption.md -->
### guide as a first-class page_type (classifier vocabulary + retype + URL canonicalisation)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** "guides typed page_type=guide directly with sections=['generic-text-block']… migration_adoption_add_guide_page_type.sql worked — adoption classifier emits guide, NO post-hoc re-typing needed this run" (2026-06-05 re-adoption); retype + URL migrations APPLIED 2026-06-04.
- **what:** Guides were folded into blog-post by the analyze_site enum (source guides lived flat at /blog/guide-*.html), so `query.pages_where_type:guide` returned zero. Structural route chosen over the band-aid (query blog_post): add `guide` to the classifier enum + guidance (quote-free replace() edits on default_config), re-type the 5 content-bearing guide-* pages, and move URLs to the canonical `/guides/<slug>/index.html` (peer of tools/games; page_canonical.go already had the guide case). Earlier session had *rejected* typing guides as `guide` when the plan was flat-blog faithfulness — the later product decision flipped to canonical nesting. `pages.page_type` has only a kebab-case CHECK, no value allowlist. The `game` page_type gap (flagged in FOCUS_component_schema_patterns) was closed the same way earlier — a doc-vs-live-DB staleness lesson.
- **sources:** migration_adoption_add_guide_page_type.sql, migration_retype_guides_to_guide.sql, migration_guides_url_to_canonical.sql, running_notes_14(25)#part-14–14b, FOCUS_component_schema_patterns.md#missing-page_type-game
- **relations:** guide-list Tier-D resolution; guides faithfulness question (F1 duplicate rows); build-site-planner's staler vocabulary (preserves existing verbatim, so adopted types survive)
- **verify-later:** site-adoption-agent analyze_site enum in live def; pages.page_type values on a fresh adoption

<!-- SOURCE: U09_adoption.md -->
### Interactive fingerprint extraction (Path C: tools rebuilt as prose)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** Planned workflow insertion sketched ("extract_interactive_fingerprint (NEW)… between extract_fingerprint and check_has_external_css") with collision check, in the FOCUS_component_schema_patterns appendix / old2 README; no deployment claim anywhere.
- **what:** Adoption pulls markdown but not `<script>`/`<canvas>` interactive machinery, so crawled calculator pages rebuild as paragraphs describing the calculator. Path C plans a second fingerprint pass over the same crawl_result capturing interactive elements, feeding tool recreation. (In practice tool-recreation-handler + A1 fixes got real tools deploying from recreate prompts; the fingerprint step itself was never built.)
- **sources:** FOCUS_component_schema_patterns.md#appendix, old2/README.md, FOCUS_adoption_fidelity_and_variants.md#problems-ranked (#3)
- **relations:** tool-recreation-handler; adoption fidelity problems ranked
- **verify-later:** site-adoption-agent workflow for any extract_interactive_fingerprint step (expected absent)

<!-- SOURCE: U09_adoption.md -->
### Adoption resume logic (never built)
- **category:** adoption-pipeline
- **status-signal:** abandoned
- **status-evidence:** "orchestration_states.collected_data already persists per-step output (378KB survived a failed run), but ResumeWorkflowTopic has no subscriber — resume was anticipated, never built. User: 'a new crawl is fine.'" (FOCUS_adoption_fidelity_and_variants, deferred list).
- **what:** Mid-workflow resume of a failed adoption from persisted collected_data. The plumbing half-exists (state persistence, a resume topic) but no consumer; the accepted operational answer is re-crawl/re-adopt. Interacts with the fetch_primary_css hard-fail trade-off (a surviving state isn't reusable without resume).
- **sources:** FOCUS_adoption_fidelity_and_variants.md#deferred, old2/HANDOFF_2026-04-22#resume-logic
- **relations:** error_step fix reduced the need (CSS timeout no longer fatal); teardown+re-adopt operational pattern
- **verify-later:** ResumeWorkflowTopic subscribers (expected none)

<!-- SOURCE: U12_docs024_archives.md -->
### Single-agent adoption trigger (positional domain, no orchestrator wrapper)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** `007_adoption_pipeline_v2_april26.md`: a dedicated `site-adoption-agent` triggered directly via `./trigger-adopt-site.sh gamedesign.uk`; the patch rewrites this into "Two agents, one thin wrapper," documented fully in live `007_adoption_pipeline_v4.md`.
- **what:** Adoption originally ran as one agent invoked directly by a shell script with a positional domain argument, mixing "site being crawled" and "site being built" into a single identifier. Replaced by a thin `site-adoption-orchestrator` (spawn → call → complete) that spawns `site-adoption-agent` as its own K8s Job, and a JSON trigger payload separating `target_url` (crawl source) from `destination_domain` (site being built) — while keeping the old `url`/`domain` shape as legacy-compatible input.
- **sources:** old/older1/007_adoption_pipeline_v2_april26.md#"The adoption agent", #"Adoption modes"; old/older1/007_adoption_pipeline_v2.patch.diff; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"The adoption agent", #"Source vs destination"
- **relations:** "every pod-running agent needs a parent that spawned it" (development-guide)
- **verify-later:** confirm `site-adoption-orchestrator` agent_definitions row exists and `trigger-adopt-site.sh` uses the JSON payload shape today.

<!-- SOURCE: U12_docs024_archives.md -->
### Unified `design` spec aspect for adopted sites (superseded precursor)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** `007_adoption_pipeline.md` (v1): single `design` spec aspect; live `007_adoption_pipeline_v4.md` has a dated addendum, "Design Fingerprint Pipeline (added 2026-04-12)," documenting the two-aspect replacement.
- **what:** The earliest adoption design captured only one `design` spec aspect, generated by the LLM alongside identity/structure classification — a single blended palette-and-typography guess with no separation between what the source site actually used and what the new site should aim for. Replaced by the `design_reference`/`design_intent` split (see merged entry above).
- **sources:** old/older1/007_adoption_pipeline.md#"What gets stored where"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Design Fingerprint Pipeline (added 2026-04-12)"
- **relations:** design_reference/design_intent spec-aspect split (its replacement)
- **verify-later:** check `site_specs` for any legacy rows with `aspect='design'` from pre-2026-04-12 adoptions never migrated.

<!-- SOURCE: U12_docs024_archives.md -->
### Two-stage adoption processing (LLM classifies, Go extracts) → three-stage
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Archive header: "Two-stage processing: LLM classifies, Go extracts." Live header: "Three-stage processing: Go extracts design, LLM classifies, Go extracts content."
- **what:** Early adoption split work into just two stages — lightweight LLM classification from page summaries, then Go-only content extraction. The later design inserts a Go-only design-fingerprint extraction stage (colours/fonts/CSS vars/layout via goquery) ahead of LLM classification, on the principle "don't ask an LLM to read hex values when a regex can do it."
- **sources:** old/older1/007_adoption_pipeline.md#"Two-stage processing"; docs024_key_docs_latest/007_adoption_pipeline_v4.md#"Three-stage processing"
- **relations:** unified design spec aspect (above), design_reference/design_intent split
- **verify-later:** `extract_design_fingerprint_action.go`/`enrich_fingerprint_with_css_action.go` existence and wiring.

<!-- SOURCE: U12_docs024_archives.md -->
### Section recipes for adoption (purpose + structure + reference implementation)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4: Requirement-Driven Components (longer term)" in the 2026-04-11 plan; no confirmation of shipping in any later doc reviewed.
- **what:** When adopting a site, each section would be captured as a "recipe": purpose, structure, reference implementation (guide not spec), and component match. Recipes without a good match would generate `needs_new_component` work items where the recipe becomes the build brief.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Decisions Made & Rationale (4)", #"Phase 4"
- **relations:** component selector by functional requirement; needs_new_component work items
- **verify-later:** whether any adoption workflow step produces structured "recipes" today.

<!-- SOURCE: U12_docs024_archives.md -->
### Adopt-from vs deploy-to separation (unbuilt)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** "discussed but not implemented. Options: snapshot to S3, stage to subdomain, or store crawl artifacts."
- **what:** Unbuilt idea for a staging area distinct from the live deploy target, so a freshly-adopted rebuild could be reviewed before overwriting production. Workaround at time of writing was manual: pause work items, verify specs, unpause.
- **sources:** old_design_and_styling/FOCUS_design_and_styling_adoption_HANDOFF_2026-04-16_v2(1).md#"Architecture Decisions Made (item 6)"
- **relations:** site snapshots and revert (014); design fingerprint extraction pipeline
- **verify-later:** whether any staging/subdomain mechanism exists for adoption today.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Adoption interactivity misroute — canonical-prefix key desync (M2 root cause, T1)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1/§2: "Deployed: T1 (apply_adoption_plan_action.go — canonical-keyed buildPageFeatureMap) ... Both are in production."
- **what:** `apply_adoption_plan_action.go` routes adopted pages by interactivity (`len(page.Features) > 0` → `needs_tool_recreation`; else → `needs_content_page`). `buildPageFeatureMap` keyed its feature map by the raw adoption-LLM page key, but the routing loop looked up the canonicalised name via `datahelpers.CanonicalisePage`, whose `tool` branch adds a `tool-` prefix while its `game` branch preserves an already-present `game-` prefix. Result: every tool page missed the lookup (empty Features → static content route); games matched only by coincidence. Fix: key `buildPageFeatureMap` by the canonical name.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.6,§2.7,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2
- **relations:** Abandoned "no owner" claim; Post-adoption detection check (T2); Canonicalise tool page identity (T3); Recreation-loss defect
- **verify-later:** confirm in production that `buildPageFeatureMap` still contains the canonical-keyed version — HANDOFF flags a parallel adoption chat may have re-edited the file since merge

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool routing fix deployment status (T1 + T2 in production; symptom fix unconfirmed)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Deployed: T1 ... and T2 ... Both are in production." / "Not confirmed: that widgets now actually deploy ... Hold the trigger. Do not mass-emit needs_tool_recreation yet" (HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1)
- **what:** The authoritative status record for the tool-widget-clobber investigation as of 2026-05-26: T1 (routing fix) and T2 (detection check) are both confirmed deployed to production, with defined acceptance criteria for calling the deploy complete (every tool/game page has_widget=t; a deployed tool page renders an interactive widget in-browser; T2 finds nothing new on a steady-state run; no duplicate pages). None of those criteria were confirmed met at time of writing — the recreation-loss defect remains open and blocking.
- **sources:** tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §1,§6,§7
- **relations:** Adoption interactivity misroute (T1); Post-adoption detection check (T2); Recreation-loss defect
- **verify-later:** re-run the exact acceptance-criteria queries against current gamesdesign.co.uk state

<!-- SOURCE: U14_docs019_runbooks.md -->
### Adoption-first fidelity inversion
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "Q4 ANSWERED — adoption does NOT call the classifier; the lean is inverted … adoption writes the specs FIRST; the classifier, when the relay later reaches it, CONSUMES them under its fidelity rules"; Q7 answered in §B3.
- **what:** How site adoption meets the relay: site-adoption-agent does the heavy work (firecrawl 30 pages, no-LLM design/interactive fingerprints, three LLM analyses — site analysis, archetype snapshot with improvement-loop constraints, content-direction guide), writes specs + pages + work items, then hands off needs_domain_research into the relay; the classifier's adoption-fidelity block treats adopted identity/archetype/content_direction/design_intent as ground truth outranking its own search+scrape. apply_adoption_plan writes site_archetype reading from collected data regardless of declared input_fields.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2; docs019/RUNBOOK_builder_route(21).md#B3 (Q7)
- **relations:** work-item relay spine; vertical-exemplar researcher (adopted sites run the hop too)
- **verify-later:** site-adoption-agent workflow; check_adoption_skip_scrape branch; classifier fidelity prompt block

<!-- SOURCE: U15_docs019_running_notes.md -->
### Adoption pipeline consumption of vertical/exemplar research
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "Adoption (user orientation): classifier CONSUMES adoption specs is CONFIRMED from its own workflow (skip-scrape conditional on site_archetype...)" (v4(39), 2026-07-04); "Q4: adoption never calls the classifier... Lean is INVERTED: adoption writes first, classifier consumes under fidelity rules later." (v4(39)).
- **what:** Clarifies the actual (initially misunderstood) relationship between the site-adoption pipeline and the domain-research-classifier: adoption crawls and fingerprints the target site first, writes specs/pages/work items via `apply_adoption_plan`, then hands off to the relay (`needs_domain_research`) — the classifier consumes adoption's output under fidelity rules, rather than adoption calling the classifier directly as first assumed.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B2 read" entry.
- **relations:** Work-item relay / builder-generations architecture; vertical-exemplar-researcher hop.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Adoption writes first; classifier consumes (the corrected lean)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** README_flows: "your instinct was half right, inverted … apply_adoption_plan writes the specs, pages, and work items itself — it never calls domain-research-classifier."
- **what:** The adoption orchestrator is a thin spawn→call wrapper; the agent crawls via firecrawl, extracts design/interactive fingerprints without an LLM, runs three LLM analyses (site structure; archetype snapshot with improvement-loop constraints; content-direction style guide), and apply_adoption_plan writes specs/pages/work items directly. The classifier later consumes adopted specs under fidelity rules when the relay reaches the site. Parked question: does apply_adoption_plan write site_archetype (classify_archetype's output isn't in its declared inputs).
- **sources:** README_flows.md
- **relations:** relay spine; classifier consolidation queue
- **verify-later:** apply_adoption_plan_action.go; site_archetype writer

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption path — vertical-slice dogfooding of the ratchet
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §8.1 "Vertical slice, not horizontal layer … Dogfood the ratchet … First capability = writing Go actions"
- **what:** Walk one capability (writing Go actions) end-to-end through route→produce→verify→gate→feedback before generalising; each new machinery piece starts at `confirm-every` and graduates on evidence. Phases 1–2 double as "improve my current chat workflow"; 3–6 are the leap to autonomy.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#8.1, ED/MASTER_autonomous_build_and_operate(4).md#8.2, ED/MASTER_autonomous_build_and_operate(4).md#8.5
- **relations:** automation ratchet; self-development coding pipeline
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Doc-tree adoption plan (constitution + tag/embedding retrieval)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_doc_tree_adoption.md header "actionable plan … without committing to the atomic rewrite, the mediator, or the routing build"; §1 "the corpus does not fit in context (~200 files, ~6.7MB, ~1.0–1.7M tokens)"
- **what:** First path to value from the doc-tree design against the current setup: Phase 1 write a tiny constitution, Phase 2 tag existing docs by concern/`applies_to` into a manifest, Phase 3 make the retrieval split real (tag-based deterministic selection for rules; existing nomic/pgvector/ollama RAG for the broad corpus), Phase 4 atomic extraction deferred/evidence-driven.
- **sources:** ED/FOCUS_doc_tree_adoption.md#4, ED/FOCUS_doc_tree_adoption.md#2, ED/FOCUS_doc_tree_adoption.md#5
- **relations:** atomic standard; mediator routing; RAG actions (existing stack)
- **verify-later:** rag_actions/nomic prefixes; proposed doc_index/standards table

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption pipeline & backend capability tiers (three-layer infra)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 007 "Phase 1 — Adoption pipeline (current)"; Phases 4–5 marked planned/future
- **what:** The platform runs in three layers (Layer 1 core factory; Layer 2 client delivery via S3 + config-driven site-api-router + client Postgres; Layer 3 framework builder), with five backend capability tiers from static+JS up to full platform. Adoption is a one-off capture, not a permanent state.
- **sources:** WM/007_adoption_pipeline_v3.md#infrastructure-separation, WM/007_adoption_pipeline_v3.md#backend-capability-tiers, WM/007_adoption_pipeline_v3.md#site-adoption, WM/007_adoption_pipeline_v3.md#principles
- **relations:** site adoption agent; design fingerprint; component selector/creator; site-api-router
- **verify-later:** site-adoption-orchestrator; site-api-router; vetcomparison export path

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Site adoption agent (crawl → fingerprint → classify → apply plan)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 007 "site-adoption-agent workflow (runs in the spawned pod): 16 steps … apply_adoption_plan → complete"
- **what:** A thin `site-adoption-orchestrator` wrapper spawns `site-adoption-agent` to run a 16-step workflow: firecrawl crawl, Go design-fingerprint extraction, LLM classification/archetype/content-direction/design-intent, and `apply_adoption_plan` writing specs, pages, and work items. Separates `target_url` (crawled) from `destination_domain` (built).
- **sources:** WM/007_adoption_pipeline_v3.md#the-adoption-agent, WM/007_adoption_pipeline_v3.md#three-stage-processing-go-extracts-design-llm-classifies-go-extracts-content, WM/007_adoption_pipeline_v3.md#running-an-adoption-what-to-expect-and-what-to-watch
- **relations:** wrapper-orchestrator; design fingerprint; canonicalisation (page identity); interactive parse-stage gap
- **verify-later:** apply_adoption_plan_action.go; extract_design_fingerprint_action.go; firecrawl_crawl

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Adoption faithfulness — WriteSitePlanAction identity strip
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §9 "The corruption is in WriteSitePlanAction, not the LLM … Fix direction (not yet applied)"
- **what:** Even after a faithful adoption, `WriteSitePlanAction`'s `ValidateRoles`+`CanonicalisePage` interaction permanently strips identity for `content`/`blog_post` page_types: `ValidateRoles` derives a slug that strips `tool-/guide-/game-`/`-index`, and `CanonicalisePage` only re-adds prefixes for tool/game/guide roles — so mistyped section-index hubs flatten. Root cause is the wrong `page_type`; clean fix is upstream at adoption time.
- **sources:** WM/016_debugging_guide_v2_44.md#adoption-faithfulness-llm-convergence-are-faithful-writesiteplanaction-strips-identity-for-content-blog_post-types, WM/016_debugging_guide_v2_44.md#0, WM/ARCHITECTURAL_TENSIONS(2).md#tension-2-page-identity-is-derived-in-multiple-places-that-can-undo-each-other
- **relations:** CanonicalisePage; architectural tension #1/#2; locks (adoption_locked); FOCUS_adoption_faithfulness_via_locks
- **verify-later:** WriteSitePlanAction; datahelpers/page_canonical.go ValidateRoles/normaliseSlug; analyze_site page_type

<!-- SOURCE: U18_sql_for_agents.md -->
### site-scraper (Firecrawl scrape → site_context)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 032 definition ("Uses firecrawl_scrape action (requires patch_02_webscrape_url_field.go)").
- **what:** Scrapes a live site's homepage via the webscrape adapter (Firecrawl), then an LLM step transforms results into the site_context format webdesign-agent consumes — the original design-transfer mechanism.
- **sources:** 032_site_scraper_agent.sql
- **relations:** webdesign-agent; ancestor of site-adoption-agent's full crawl
- **verify-later:** whether site-scraper is still used vs site-adoption-agent

<!-- SOURCE: U18_sql_for_agents.md -->
### tool-recreation-handler (recreate interactive tools from crawled source)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 099 definition; live-run evidence in 138 (run e1018366 recreated bugs faithfully → prompt fix) and 132/137 (note-writing wired, subject corrected).
- **what:** Two-stage recreation of JS-heavy pages during site adoption: analyze_tool (LLM functional spec from source + context) then recreate_tool (Opus generates working replacement HTML/CSS/JS), with completeness/truncation checks, validation, save/deploy. 138 adds the "Mandatory Behaviour Requirements" prompt section rendered from spec.interactive_features which OVERRIDES the original source — fixing the observed failure where explicit spec fixes were buried in analysis JSON and Opus faithfully recreated the original bugs.
- **sources:** 099_tool_recreation_handler.sql; 138_recreate_tool_carries_spec_features.sql; 137_recreation_spec_and_note_subject.sql
- **relations:** site-adoption-agent creates its items; tool acceptance verifies results
- **verify-later:** current recreate_tool prompt; spec.interactive_features producers

<!-- SOURCE: U18_sql_for_agents.md -->
### Site adoption pipeline (site-adoption-agent + wrapper orchestrator)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** 091 definition; 104 adds the wrapper (2026, "Pattern copied verbatim from med-export-orchestrator"); 115 adds the 'adoption' lock source for "faithful first pass".
- **what:** Adopts an existing live site: firecrawl_crawl via the webscrape adapter returns per-page markdown; an LLM analyze step classifies pages and extracts identity/design/content structure into a JSON plan; apply_adoption_plan creates site_specs, page records, and work items to recreate the site in-platform. 104 wraps it in a spawn→call orchestrator so the long crawl runs in its own Job pod with clean correlation logs.
- **sources:** 091_site_adoption_agent.sql; 104_site_adoption_orchestrator.sql; 115_locks.sql
- **relations:** page-content-writer Recreate Mode; tool-recreation-handler; adoption locks
- **verify-later:** apply_adoption_plan action; adoption directive writer (pending per 115)

<!-- SOURCE: U20_legacy_docs_a.md -->
### 11-agent website analysis framework (four agent groups)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Whole docs003 set is planning ("Here is the detailed, point-by-point analysis…"); docs004 explicitly reframes it ("The old numbers are meaningless now… rename them") into the Learn/Execute playbook model.
- **what:** The original web-capture master plan: Strategy & Content group (Strategist A10, Content Infuser A11), Library & Storage (Librarian A7, S3+Postgres/pgvector), Design Ingestion (Prospector A0, Site Profiler A1, Capture Bot A2/Playwright, Layout & Labeling A3 XY-Cut+LLaVA, Component Generator A4 VLM screenshot-to-code, Style Extractor A5 getComputedStyle — later eliminated in favour of Firecrawl branding data, Behavior Extractor A6 CodeLlama), Generation (Publisher A8 "Dribbble-like" showcase site, Architect A9 template builder querying by CLIP embedding). All implemented as agent_definitions rows + new action adapters, not new binaries.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md; docs003_firecrawl/README.0124.11_agent_summary.md; docs003_firecrawl/README.0121.good_gemini_summary_of_architecture.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** successor chain: playbook model (docs004) → MVP site builder → current adoption-pipeline (docs 007) and site-spec-and-classifier (docs 021). Publisher A8's public design-library site was abandoned.
- **verify-later:** which of the 11 agent types ever got agent_definitions rows.

<!-- SOURCE: U20_legacy_docs_a.md -->
### website-analyzer conditional scraping group
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** Tested with kcat messages (boxing-tickets.com, both basic and structured+crawl variants) and a live UPDATE of its orchestration_workflow.
- **what:** An agent group that takes target_url + flags (extract_structured, crawl_pages, crawl_limit/depth) and conditionally routes between basic scrape, structured extraction, and multi-page crawl using evaluate_condition — the first "smart" capture entry point.
- **sources:** docs003_firecrawl/README.0129.testing_webscrape_message.md; docs003_firecrawl/README.0128.go_text_template.md
- **relations:** firecrawl adapter; successor: adoption pipeline crawl/classify flow.
- **verify-later:** agent_group_definitions row group_type='website-analyzer'.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Playwright capture adapter + website-capture agent
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Complete deliverables (adapter py, capture_actions.go, agent SQL: desktop/mobile viewports, hover/focus states, scroll intervals with parallax/sticky detection, asset extraction, S3 upload) — but docs004/website_analysis 002 records "Agent 5 eliminated… use Firecrawl branding data instead" and firecrawl/001 adapts the MVP away from Playwright.
- **what:** Deep browser-based capture: Playwright adapter on system.adapter.playwright.requests capturing full-page desktop + mobile screenshots, DOM, computed styles, interaction states (hover/focus for up to 50 selectors), scroll-position screenshots (0/25/50/75/100%) with parallax/sticky detection, asset extraction, and organised S3 upload with manifest. Deferred in favour of the managed Firecrawl service for MVP; the deeper capture ideas (interaction/scroll states) never resurfaced.
- **sources:** docs004_website_capture_project/playwright/website_capture_agent.sql; docs004_website_capture_project/playwright/playwright_adapter.py; docs004_website_capture_project/playwright/implementation_roadmap.md; docs004_website_capture_project/firecrawl/001claude_initial.md
- **relations:** firecrawl adapter (chosen replacement); adoption-pipeline crawling successor; behaviour capture (rrweb) idea from docs003 also abandoned.
- **verify-later:** playwright-adapter deployment existence.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Website-builder orchestrator (capture → vision → code → synthesis → content → library)
- **category:** adoption-pipeline
- **status-signal:** abandoned
- **status-evidence:** Orchestrator SQL references agent types (website-vision, website-code-analyzer, website-synthesis, content-strategist) and actions (analyze_input_type, parallel_section_generation, store_component) that are never defined or mentioned again; the MVP builder took a different shape.
- **what:** A master workflow to rebuild a site from a captured one: capture data → visual analysis (layout/palette from screenshots) → code cleaning/analysis → synthesis correlating visual+code into a template → content planning → parallel section generation → aggregate → store components with embeddings in the library. The maximal "clone-and-improve" vision.
- **sources:** docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql; docs004_website_capture_project/playwright/website_builder_integration_guide.md; docs003_firecrawl/README.0125.claude_11_agent_summary.md
- **relations:** successor in spirit: adoption-pipeline content recreation (docs 007); vision analysis resurfaces in the current image-analysis tooling.
- **verify-later:** confirm none of the four sub-agent types exist in agent_definitions.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Adopting existing external sites ("Adopt" workflow)
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Designed in 004 ("Adopt workflow… status: 'adopted_partial'… match_confidence"); the taxonomy's adoption-pipeline (docs 007: site crawling, classification, content recreation) is the named live successor.
- **what:** Run the Learn loop against an existing site the platform didn't build: scrape, deconstruct layout, match found blocks to the in-house component library with confidence scores, generate a manifest marking it adopted_partial — making external sites partially manageable by agent edit workflows.
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md
- **relations:** successor: adoption-pipeline (docs 007).
- **verify-later:** compare with current adoption pipeline design.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Site interrogation & pattern library
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** docs009/003: "'Interrogate' successful sites to extract... Store extracted patterns"; docs012/012 Part 3 details the 5-phase pipeline (discover → firecrawl capture → LLM structure analysis → pattern extraction → component creation) with pattern_sources table marked "(future)".
- **what:** Learning from successful sites without copying: capture HTML+screenshot, LLM-analyse section types, visual hierarchy, content strategy and psychological principles, extract reusable patterns tagged by industry/funnel-stage/audience with "why it works" notes, and mint content_components (origin_type='extracted') from them. Patterns become queryable ("for finance trust-building, use X") and feed component selection. The most persistent unfulfilled idea of this era — restated in docs009, docs010 roadmaps (Phase 4), and docs012.
- **sources:** docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#Part-3; docs009_site_interrogation_and_solutions/003_claude_save_point.md#2; docs010_multitrack_flows_persona_architecture/018_priority_matrix.md
- **relations:** Pragmatic Evolution Engine phase 2; adoption-pipeline site crawling (current descendant); pattern_sources/captured_sites tables; component library.
- **verify-later:** pattern_sources table; origin_type/industry_tags/funnel_stages columns on content_components; website-capture-firecrawl agent.

<!-- SOURCE: U21_legacy_docs_b.md -->
### site-scraper companion agent (design context from live URLs)
- **category:** adoption-pipeline
- **status-signal:** aspirational
- **status-evidence:** docs017/003: site-scraper "firecrawl_scrape → analyze_design → returns site_context for webdesign-agent"; standardized site_context schema with source: "database|scrape|manual".
- **what:** A standardized site_context interface (domain, company, industry, palette, typography, component functions, source) produced by either DB load, live-site scraping, or manual input, so the webdesign-agent can restyle from any source — enabling "scrape competitor → feed to design agent → apply to your site" pipelines. The schema-standardization idea matured; the scraper flow folded into the adoption pipeline.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/003_design.md#Architecture; docs017_legacy_agent_rules_images_design_keydocs/008_checklist_for_new_specialist_agents_v5.md#Standardized-Interface-Schemas
- **relations:** adoption-pipeline capture; webdesign-agent; standardized interface schemas doctrine.
- **verify-later:** site-scraper agent definition; load_site_for_design action.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### "Type guides as `guide`" — falsified as a quick companion fix, later built properly as a structural fix
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 2: "**Falsified the 'type guides as `guide`' companion.** ... source guides live flat at `/blog/guide-rng-design.html`, while a `guide` role would *nest* them... typing `guide` would be *less* faithful... Left as an open product decision; did NOT ship the wrong patch." Then Part 14 (2026-06-04), same session-log lineage: `migration_retype_guides_to_guide.sql` + `migration_guides_url_to_canonical.sql` were written, applied, and guides were deliberately moved to `/guides/<slug>/index.html` — the exact "less faithful" move earlier rejected, now chosen deliberately once `guide` became a first-class page_type with its own canonicalisation rule.
- **what:** Two-stage decision on how adopted "guide" content should be typed/URL'd. First pass: rejected retyping guides as `guide` as a *quick fix* for a de-prefixing side-effect, because it would misplace the URL relative to the untouched source. Second pass, as a *deliberate structural project*: added `guide` to the page_type enum, re-typed the 5 real guide-* pages, added the classifier's default_config guidance, and migrated their URLs from `/blog/guide-*.html` to `/guides/<slug>/index.html` — closing the exact gap the earlier rejection had flagged as "an open product decision."
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 2, 13, 14, 14a–14h
- **relations:** SyncPagesToDBAction canonicalisation divergence; bare-guide-duplicate defect (below); adoption-faithfulness-via-locks (below)
- **verify-later:** `pages.page_type` enum and `page_canonical.go`'s `guide` case in current code; whether `build-site-planner`'s vocabulary was ever updated to include `guide` for new adoptions.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Bare-guide duplicate pages — root cause: planner ignores adopted state (prompt-rule gap, not wiring)
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 14e: "DECISIVE (llm_call_log plan_site)... `saw_guide_pages=t`, `prompt_says_no_existing=f`, `planned_bare_in_response=t`... So the planner WAS given the adopted guides and emitted `economy-basics` anyway → PROMPT-RULE gap... NOT a wiring/status gap." Cleanup migration (`migration_cleanup_bare_guide_duplicates.sql`) applied and confirmed durable (Part 14f: "current-plan bare-name query returns 0 rows").
- **what:** `build-site-planner` re-invents a differently-slugged sibling page (`economy-basics`) for a topic already adopted under a prefixed name (`guide-economy-basics`), because its "never duplicate an existing page" prompt rule only named games/tools examples and didn't generalise to the `guide-` prefix pattern. This is a fresh, concretely-diagnosed instance of the previously-documented `FOCUS_planner_ignores_adopted_state.md` mechanism (2026-05-19). A durable Go-level fix (deterministic topic-stem collision guard in `validate_site_plan`/`write_site_plan`, reusing `CanonicalisePage`'s prefix-stripping) was recommended but not shipped in this arc; only the data cleanup + an optional prompt-rule stopgap were delivered.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 14c–14g
- **relations:** adoption-faithfulness-via-locks convergence (below); "type guides as guide" (above)
- **verify-later:** `FOCUS_planner_ignores_adopted_state.md`; whether the Go-level topic-stem guard was ever built.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### A1 — tool/game pages never deployed: root-cause hypothesis evolution
- **category:** adoption-pipeline
- **status-signal:** superseded
- **status-evidence:** Original catalogue (base, 2026-05-28): "*Tentative area:* the gap is between 'page row built/complete in DB' and 'file committed to git'. Could be the rerender/deploy path skipping nested child pages, the tool-recreation handler not producing a deployable artefact, or a git-path mismatch." Catalogue `(4)` (2026-06-04): "*Cause pinned (two coordinated root causes):* 1. Parser... `saveSectionsExtractFromHTML` extracts only `<section>` blocks, but `tool-recreation-handler`'s prompt emits `<div class="tool-page">` (no `<section>`)... 2. Flip churn: `upsertPage`'s ON CONFLICT flipped `deployed → needs_rebuild`."
- **what:** The site's actual interactive product (tools/games) never deployed a file despite `pages` rows and `complete` work items. The three original hypotheses (deploy-path bug, handler artefact-production bug, git-path mismatch) were all superseded by two pinned, source-confirmed causes: an HTML-fragment parser that only recognises `<section>` blocks (tool output uses `<div class="tool-page">`, so it silently extracted zero sections), plus the ON CONFLICT flip churn (above). Fix: single-fragment fallback in the parser + the Option B stamp/flip removal. Verified end-to-end on a subsequent adoption run (all 5 games + tools deployed with working links).
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md; adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md; running_notes_14(20) Parts 7–10
- **relations:** deployed→needs_rebuild flip (above); dispatch throughput bottleneck (Family J, below)
- **verify-later:** `save_page_sections_action.go` current parser logic.

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### tool-page canonicalisation misroute (adoption Features key desync)
- **category:** adoption-pipeline
- **status-signal:** partial
- **status-evidence:** "RESOLVED 2026-05-26 → b1" (root cause confirmed via query G) but the numbered "Potential solutions" section proposes the actual code fix as still to be applied — diagnosis complete, fix landing unconfirmed from the doc alone.
- **what:** A specific, resolved-diagnosis bug postmortem: an adopted `page_type='tool'` page deploys with prose describing the tool but no interactive widget. Two distinct causes were disambiguated (M1 — a widget existed and a later `SavePageSectionsAction` rebuild deleted it via a text-only regression guard blind to script-heavy content; M2 — no widget was ever generated because adoption captures text but has no JS-parse stage). Root cause for the gamesdesign.co.uk case was M2, but *not* because generation is unowned — `tool-recreation-handler` exists and should have run. The actual fault: `apply_adoption_plan` routes by `len(page.Features)`, but `buildPageFeatureMap` keys its map by the **raw** page name the adoption LLM wrote, while the routing lookup uses the **canonicalised** name (`CanonicalisePage` prepends `tool-` for tools) — so every tool page's feature lookup misses even when the LLM correctly detected interactivity, silently misrouting it to the static `page-build-handler` path instead of `tool-recreation-handler`. Games pages don't hit this because their canonical prefix (`game-`) already matches the raw key.
- **sources:** docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/working/main/016_debugging_guide_addendum_adopted_tools_no_widget(3).md
- **relations:** doc 029 (CanonicalisePage, Phase-0 helper); component-regeneration-flow (SavePageSectionsAction clobber path); spurious-duplicate-pages pattern (below, same "adoption vs. a second surface" family)
- **verify-later:** buildPageFeatureMap in the adoption/orchestration action code — confirm whether the canonical-key fix was applied

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### spurious duplicate pages from "planner ignores adopted state"
- **category:** adoption-pipeline
- **status-signal:** deployed
- **status-evidence:** The migration is a real, executed cleanup with before/after verification queries and an explicit rollback path via `_bak_del_bare` snapshot tables, dated against a specific incident (gamesdesign.co.uk, created 2026-06-03 20:25:30, cleaned up in this migration).
- **what:** A confirmed, named failure pattern: a post-adoption planner pass (`build-site-planner`/`blog-content-planner`) invents new `page_type='blog-post'` pages (`sections=[]`, `build_status='planned'`, never rendered) that duplicate content already faithfully recreated by adoption as `page_type='guide'` pages at a different URL — "a second surface invents parallel pages after adoption" because the planner doesn't check adopted state before generating its own content plan. The cleanup migration is durable — it removes the bare pages from the pages table, the *current* `site_plan_pages`/`site_plan_sections` (so the reconciler won't recreate them), and terminalises the dangling `site_work_items` rows (which have no FK to pages and would otherwise linger holding a dedup slot) — but explicitly does not fix the upstream planner logic that would reintroduce the same duplicates on a future `plan_site` run.
- **sources:** docs/_archive/agent_docs/sql_for_tables/040b_migration_cleanup_bare_guide_duplicates(1).sql
- **relations:** tool-page canonicalisation misroute (above, same "adoption vs. second surface" bug family); FOCUS_planner_ignores_adopted_state.md; doc 029; site-plan-and-reconciler
- **verify-later:** FOCUS_planner_ignores_adopted_state.md, whether the upstream planner prompt/logic has since been tightened to check adopted state first

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dream spec / gap analysis / feasibility (one spec with status, not two documents)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 021 resolved decision 24 "One spec, not two"; 028: per-item status "not fully implemented yet — Phase 2"
- **what:** The full spec IS the dream; items carry status deployed/planned/blocked; gap analysis = blocked/planned subset; feasibility-recheck promotes blocked→planned when capability arrives; feasibility annotations prevent impossible work items. Older 002d dream_spec-in-content_data shape superseded by this. Phase 2 of 028 makes it mechanical.
- **sources:** 021#One Spec, Not Two; 028#The spec has status; 002d#Gap Analysis
- **relations:** fidelity dial; feasibility-recheck task
- **verify-later:** does feasibility-recheck scheduled task exist; per-item status columns

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Site spec unification (site_specs aspects as the one authoritative spec)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 021 target-state doc with immediate/short/medium ordering; backfill + fallback both recommended; newer pipeline already uses aspects
- **what:** One versioned spec per site as independent aspect rows (classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance), is_current/superseded_at per-aspect versioning with source/source_agent/source_item_id provenance; write_site_spec deep-merges so every row is a complete self-contained record (pruning-safe). content_data is legacy; read_site_spec falls back to it. Classifier writes intent, planner implements, design agent executes, audits enforce. The 15 content-strategy questions map onto aspects; deep research is a future classifier enrichment.
- **sources:** 021 full; P1#Site Specification System
- **relations:** 028 ownership; 030 strategic-vs-plan-time; dream spec
- **verify-later:** backfill done for legacy sites; read_site_spec fallback

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fidelity dial (locked/high/medium/low + no-adoption confidence mode)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 028 Phases: Phase 1 current (implicit high); explicit fidelity input is Phase 3; depends on per-item spec status (Phase 2)
- **what:** Controls how much aspiration reaches first deployment vs is deferred to the improvement loop, and the loop's promotion rate. locked = adopted-exact, no promotions; high (default with adoption) = faithful launch, ~one substantive change per audit cycle; medium = modest extensions; low = adoption as inspiration; blank domains reinterpret it as research-confidence tolerance. Lives on the trigger input + a build_policy/adoption_meta aspect.
- **sources:** 028#Fidelity, #Phased implementation
- **relations:** spec status; improvement loop rate
- **verify-later:** any fidelity/build_policy aspect in prod

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec aspect ownership and read-and-extend (anti-silent-overwrite)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028 ownership list; adoption-aware classifier prompt (migration 006) live; open question notes planner still writing design_intent pre-030
- **what:** Classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planners own site_plan. Classifier over adoption output is read-and-extend (preserve adopted dimensions, add strategy), never overwrite. Named failure modes: silent overwrite, confident fabrication on thin signal, default-to-brochure, reflexive upstream re-runs, schema-level commercial bias, adoption without strategic analysis.
- **sources:** 028#Who writes what, #Failure modes
- **relations:** 030 planner redirect to directives; composition self-heal
- **verify-later:** build-site-planner no longer writes design_intent/content_direction post-030

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Superseding a spec doesn't undo installed artefacts (re-queue rule)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028: gamesdesign remediation hit it (fresh specs + stale installed theme at sites.style_collection_id)
- **what:** Agents with install side-effects (composition, nav, pages, assets) write beyond site_specs and leave live pointers from long-lived tables; invalidating their spec must also queue the re-run work item (needs_composition etc.) — long-term the supersession itself should emit the recovery item. Test: does the agent write other tables AND does a live pointer reference the write?
- **sources:** 028#Failure modes (last)
- **relations:** install_site_composition; composition trigger matrix (027)
- **verify-later:** whether supersession-emits-item was ever built

<!-- SOURCE: U04_idea_uk.md -->
### Structured design_intent from the classifier (palette + typography reference_values)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Migration applied 2026-06-20; "Palette migration proven on a real build" 2026-06-21 — fresh idea.uk resolved `palette_source=design_intent_values`, parchment, "no invented blue".
- **what:** Root cause of generic-looking fresh builds: the classifier wrote design_intent colours as **prose** (hex buried in colour_mood sentences) while every consumer — the composition cascade and the analyze_design prompt — reads **structured** `design_intent.palette.reference_values` (8 slots) + `typography.reference_values`. The migration edits the classifier's classify_and_extract schema and adds a MANDATORY-fields bullet (all 8 slots as hex; style_direction must agree with the palette; never default to blue-and-grey), applied via snapshot_agent backup + exact-anchor replace() with a RAISE self-check. This single change is what makes both design stages agree (base = parchment, overlay starts from parchment).
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A + checklist); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md ("Direct consequence" section)
- **relations:** two-stage pipeline; mandatory-full overlay bug (this is precondition for its fix); prompt-migration discipline.
- **verify-later:** live classifier prompt contains the palette block; site 1244516d design_intent spec.

<!-- SOURCE: U04_idea_uk.md -->
### Phase 0 classifier-only positioning read
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** "Phase 0 result (2026-06-14 — ran, with the live site up)": faithful identity/classification/content_direction/design_intent specs, interactive-platform at ~0.91 confidence.
- **what:** Running just domain-research-classifier on a domain as a near-zero-cost positioning brief before committing to a build — its four spec aspects ARE the answer to "what does this site do for a stranger?". Caveats codified: a fresh read is NOT blank-slate (the classifier scrapes the live site up to 3 pages unless an adoption already ran); a generic name yields a generic name-only read, so hiding the live site removes signal not bias; a safe suppression trick exists (temporary blank nginx `location = /` — never touch DNS/nginx wholesale with a live Stripe webhook); the classifier's terminal needs_strategy item will flow into a full build if dispatch is running. Decision: leave the live site up.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 0); idea.uk/HANDOFF_chassis_site.md
- **relations:** standing-ambition mission (the read was faithful-but-backward-looking, which motivated it); fresh vs adoption.
- **verify-later:** idea.uk site 97ed2f64 spec rows (incl. the duplicate-spec tidy-up).

<!-- SOURCE: U04_idea_uk.md -->
### Build-standard classifier migration (best-in-class quality/fit, not scope)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "READY, PROVEN-CORRECT, NOT YET APPLIED… replace() simulated against the live prompt → lands cleanly" (2026-06-21); still unapplied in TODO P2 (2026-06-26).
- **what:** A migration prepending a "Build standard" block to the classifier prompt: every build should aim at best-in-class **quality and fit** for its vertical — explicitly not scope inflation. Its first draft failed on a multi-line anchor (whitespace mangling) and was fixed to a single-line anchor with a rollback proven clean — feeding the prompt-migration discipline. Test plan: fresh build first; confirm an adopted rebuild stays faithful rather than drifting.
- **sources:** idea.uk/HANDOFF(13).md (Migrations status); idea.uk/TODO_chassis_and_idea_uk(1).md#P2; idea.uk/running_notes(63).md (lll/mmm 06-20/21)
- **relations:** standing-ambition default; prompt-migration discipline.
- **verify-later:** whether migration_classifier_build_standard.sql was later applied (file itself lives outside this unit).

<!-- SOURCE: U05_content_quality_linking.md -->
### Guide as first-class page_type (classifier vocabulary + canonical URLs)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14j: "guides typed page_type=guide directly … migration_adoption_add_guide_page_type.sql worked".
- **what:** Guides were folded into blog-post by the site-adoption-agent's analyze_site enum, defeating query.pages_where_type:guide list resolution. Structural fix over band-aid: `guide` added to the adoption classifier enum + guidance, existing rows re-typed, URLs migrated /blog/guide-<slug>.html → /guides/<slug>/index.html (page_canonical.go already had the guide case). Classification geography documented: analyze_site LLM emits per-page page_type+url; site-classifier is site-type only; build-site-planner has a staler vocabulary but preserves existing pages verbatim; pages.page_type has a kebab-format check, no value allowlist.
- **sources:** running_notes_14(26).md#part-13-14c
- **relations:** Tier-D lists; adoption pipeline; bare-duplicate cleanup.
- **verify-later:** site-adoption-agent analyze_site prompt enum; pages.page_type values on a fresh adoption.

<!-- SOURCE: U14_docs019_runbooks.md -->
### site_type taxonomy drift between classifier and strategist (Q8)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B2 "vocabulary drift between hops — the classifier's site_type set (brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social) vs the strategist's canonical set (brochure|authority-portal|local-directory|review-site|content-hub|landing-page|portfolio). Two taxonomies for the same concept, one spec chain." Queued item 3.
- **what:** Two adjacent relay hops use different canonical vocabularies for the same site_type concept flowing through one spec chain — a contract-drift hygiene defect awaiting a one-canonical-set decision plus two snapshot migrations.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2 (Q8); docs019/RUNBOOK_builder_route(21).md#queue (item 3)
- **relations:** two front doors Q5; workflow result contract (drift class)
- **verify-later:** classifier and strategist prompt enumerations

<!-- SOURCE: U16_docs019_design_plans.md -->
### Classifier consolidation + site_type taxonomy alignment (queued work)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_builder_thread Q5/Q8 — a queued brief with a read-first plan, not executed in these files.
- **what:** Two classifiers overlap: domain-research-classifier (newer, relay-native) and site-classifier (pageflow/intake path, which does NOT use work-item triage — its differences may be load-bearing: intake's hitl_confirm_type keys off confirmed_type.recommended_builder). Brief: diff both rows, map dependency points, check hard before changing; merge additions both ways with snapshot migrations; deprecate only at zero usage. Behind it: the classifier and strategist use different canonical site_type vocabularies in the same spec chain — one decision, two snapshot prompt migrations.
- **sources:** HANDOFF_builder_thread.md#3; README_flows.md (Q8 note)
- **relations:** adoption-writes-first; vertical-exemplar hop
- **verify-later:** both classifier rows; intake usage query on orchestration_states

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Classifier as strategic brain (always runs full)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "The classifier … runs on every site entering the pipeline, and it always does its full job … Adoption does not shortcut it"; Phase 1 current, Phases 2–5 not implemented
- **what:** The `domain-research-classifier` decides what a site *should be* on every site; adoption/operator-mission are weighted inputs, not bypasses. It is not constrained to current capability — best-version items it can't build yet are marked `blocked` for `feasibility-recheck`. Silent override is the failure mode.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-classifier-is-the-strategic-brain-it-always-runs-in-full, WM/028_platform_mission_and_pipeline_direction.md#input-sources-and-their-weight, WM/028_platform_mission_and_pipeline_direction.md#phased-implementation
- **relations:** website mission; fidelity dial; spec-has-status; adoption pipeline
- **verify-later:** domain-research-classifier agent_definition; migration 006

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Fidelity dial (locked/high/medium/low)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "Fidelity … Five values, with high as the default when adoption evidence is present"; "depends on per-item status on specs … Phase 3"
- **what:** A dial controlling how much aspirational extension reaches the first build vs how faithfully it matches the strongest evidence (usually adoption): `locked` (exact, no promotion), `high`, `medium`, `low`; no-adoption reinterprets it as a confidence tolerance. Currently only implicit `high` (Phase 1).
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#fidelity-controlling-how-much-aspiration-reaches-the-first-build
- **relations:** spec-has-status; classifier strategic brain; adoption faithfulness locks
- **verify-later:** proposed adoption_meta/build_policy spec aspect

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Spec has per-item status — one spec, not two
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "one spec, not two. Items have status (deployed / planned / blocked) … It is not fully implemented yet … planned to be implemented in Phase 2"
- **what:** The dream is the full spec; the build is its non-blocked subset. Per-item status makes the dream-vs-build distinction mechanical — the build pipeline builds only `deployed`, `feasibility-recheck` promotes `blocked→planned`. Each spec row records source/source_agent/source_item_id for provenance; agents read-and-extend, never silently overwrite.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-spec-has-status-deployed-planned-blocked, WM/028_platform_mission_and_pipeline_direction.md#who-writes-what-who-doesnt-override
- **relations:** references doc 021; fidelity dial; feasibility-recheck
- **verify-later:** site_specs is_current/superseded_at; feasibility-recheck task

<!-- SOURCE: U18_sql_for_agents.md -->
### site-classifier → research-backed classification with domain_profile
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** 003 header: "Changes site-classifier from a single Haiku LLM guess into a research-backed orchestrator"; later 049 creates domain-research-classifier for the work-item pipeline, which takes over first-stage classification.
- **what:** Evolution of classification: v1 was one Haiku call mapping domain+objective to {landing, content, portfolio, brochure} + recommended_builder. v2 (file 003) made it an orchestrator: Haiku research brief → research-agent web investigation → Sonnet synthesis producing backward-compatible site_type plus a rich domain_profile (business identity, tone, visual_direction, image_guidance, strategic analysis). Explicit responsibility fences: does NOT pick pages or style_collection (planner's job) but DOES provide design inputs consumed by planner, image-generator, webdesign-agent, page-content-writer.
- **sources:** 003_site_classifier.sql; sql_for_agents_v1/003_site_classifier.sql; sql_for_agents_v2/000_backup_agents.sql
- **relations:** succeeded by domain-research-classifier (work-item pipeline); domain_profile is ancestor of site_specs aspects
- **verify-later:** live site-classifier definition vs domain-research-classifier; who consumes domain_profile today

<!-- SOURCE: U18_sql_for_agents.md -->
### domain-research-classifier (work-item first stage)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 049 header documents pipeline position "first agent after seed_build_queue"; 067 adds extended-thinking budget to its classify_and_extract step (conditional on patch deploy).
- **what:** Handler for needs_domain_research: researches a domain via web search and scrape, classifies site type, extracts identity signals, writes site_specs aspects "identity" and "classification", creates the next work item (needs_briefing; later needs_strategy per 060).
- **sources:** 049_domain_research_classifier.sql; 060_domain_strategist.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** successor of site-classifier v2; site_specs aspect model
- **verify-later:** current next-item wiring (strategy vs briefing)

<!-- SOURCE: U18_sql_for_agents.md -->
### spec-updater (mechanical site_specs merge from findings)
- **category:** site-spec-and-classifier
- **status-signal:** unknown
- **status-evidence:** 072 definition; no later patches in this unit.
- **what:** Handler for needs_spec_update items: applies {aspect, field, suggested_value} to site_specs with the WriteSiteSpecAction versioning pattern. No LLM. Description-only items complete as "needs human review". "The complexity is in the Go action, not in the workflow."
- **sources:** 072_spec_updater_agent.sql
- **relations:** content-gap-planner and audits emit its items; site_specs supersede-versioning
- **verify-later:** update_site_spec_from_item action

<!-- SOURCE: U19_sql_tables_components.md -->
### site_specs aspect-versioned specification store
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Table created in Phase-0 migration (019); live pg_dump backup (bk_site_specs.sql) shows the production shape including pinned; extensive backfills for real sites.
- **what:** All strategic site specification lives as (site_id, aspect, data JSONB) rows — identity, strategy, tone, design_intent, content_direction, growth_config, adoption_source — with provenance (source enum: classifier/adoption/hitl/planner/improvement/seed/manual/rollback/fork/recovery; source_agent; source_item_id) and history via is_current + superseded_at (unique current per site+aspect). write_site_spec deep-merges partials so each row is self-contained. `pinned` (Phase 4) prevents agents overriding human-set specs.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql; docs/agent_docs/sql_for_tables/bk_site_specs.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a-team-data
- **relations:** site_plans (operational counterpart); site snapshots capture current specs; identity enrichment (departments/team).
- **verify-later:** write_site_spec action; pinned enforcement in writers.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site classifier and site_type taxonomy
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** site-classifier agent SQL with the landing/content/portfolio/brochure taxonomy and recommended_group mapping, plus a template-flattening prompt fix (evidence it was actually run); taxonomy names the live classifier architecture (docs 021).
- **what:** A lightweight LLM agent classifying a project into site types — landing (conversion single-CTA), content (publishing/ads/SEO), portfolio (showcase), brochure/directory (multi-page business / listings) — with confidence, reasoning, detected signals, and a recommended builder group. The direct ancestor of the platform's archetype/classification system.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** successor: site-spec-and-classifier (classification architecture, archetype); briefing agent; intake orchestrator.
- **verify-later:** site-classifier agent row; current archetype enum vs this 4-type taxonomy.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Specialist architects per site type
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 023 SQL: landing-page-architect created (renamed copy of site-component-architect), content-site-architect created with content-site components (article grid, sidebar, ad zones, category nav), portfolio-architect created "for future use".
- **what:** One architect agent per site type, each with its own default sections and component_category filter into the library; the alternative "one architect, differentiated by build plan" was debated (025) and the group-per-project-type model won conceptually.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** groups as project recipes; conditional_call_agent routing.
- **verify-later:** the three architect rows; content-site component rows.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Site classifier agent
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/003 SQL defines 'site-classifier' with a Haiku prompt classifying landing/content/portfolio/brochure; docs015/004 confirms "Single LLM call → outputs ONE site_type... ONE recommended_builder"; current system uses the multi-aspect site-spec classifier (docs024 021).
- **what:** A lightweight LLM agent that classifies a domain+objective into a site type (landing/content/portfolio/brochure) with confidence, reasoning, detected industry and signals, and recommends the corresponding builder group. Its single-label output was later superseded by the richer site-spec aspect classification.
- **sources:** docs006_workflow_builder/003_current_state_of_agents.sql#2-SITE-CLASSIFIER; docs007_brochure_builder/001_brochure_builder_plan.md#Classification-Signals; docs015_data_flow_verification/004_builder_flow.md
- **relations:** intake orchestrator; HITL type confirmation; successor: site-spec-and-classifier architecture.
- **verify-later:** agent_definitions 'site-classifier' vs current classifier agents.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Intake orchestrator workflow (classify → brief → spawn builder)
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/011: "✅ WORKING SYSTEM: HITL Orchestration with Multi-Agent Workflow" listing the 11-step intake workflow with two HITL pauses; docs017/023 later introduces intake-orchestrator-v2 routing.
- **what:** The entry-point orchestration: spawn/call site-classifier → fetch_available_builders from DB → HITL confirm site type (human can override classifier and builder choice) → fetch builder questionnaire → briefing agent fills it → HITL review brief → spawn and call the chosen builder. Established the pattern of human quality-gates before expensive generation.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#Working-Agents; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Naming-and-Coexistence
- **relations:** site classifier; briefing agent; intake-orchestrator-v2; HITL protocol.
- **verify-later:** intake-orchestrator workflow JSON in agent_definitions; whether v2 routing exists.

<!-- SOURCE: U22_recent_small_docs.md -->
### Unified site spec (status-tagged single document)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** Doc lives under `docs022_domain_authority/old/`; "Extend the current classifier output ... Steps 1-3 can happen incrementally" — proposed, and archived.
- **what:** A proposal for the site-classifier to emit one unified spec covering classification, identity, design_intent, content_direction, pages, features, SEO, and maintenance_profile — every item tagged `status` (deployed/planned/blocked). The "dream" is the whole doc; the "build" is the non-blocked subset. Downstream agents (planner enriches rather than decides pages; design/content agents implement explicit intent; audit agents treat the spec as ground truth; HITL edits it).
- **sources:** docs022.../old/004_classifier_notes.md
- **relations:** site-classifier vertical/disposition output, feasibility/blocked-handler pattern, design_intent, HITL
- **verify-later:** site_specs.spec_type='unified_spec'; classifier identity/design_intent/content_direction fields

<!-- SOURCE: U22_recent_small_docs.md -->
### Feasibility / blocked-handler pattern
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "the claim action catches it ... Item marked 'blocked', error='Handler agent not registered' ... weeks later ... Feasibility-recheck task finds it ... promoted to 'triaged'." Describes an existing dispatch/claim mechanism.
- **what:** A pattern where planners/discovery emit work items naming an intended handler even if that agent doesn't exist yet. The claim action checks agent_definitions; unknown handlers mark the item `blocked` with an error; a periodic feasibility-recheck task promotes blocked items to `triaged` once the handler is deployed. A Go `check_feasibility` action can also pre-block at spec time based on the agent registry.
- **sources:** docs022.../old/004_classifier_notes.md#feasibility-assessment
- **relations:** unified site spec, work-item lifecycle, tool-lifecycle
- **verify-later:** claim action handler-existence check; feasibility-recheck scheduled task

<!-- SOURCE: U23_docs_root_vonc.md -->
### write_site_spec spec_data string coercion
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session: "FIXED, deployed"; migration table: code deployed 2026-06-24 ~15:00.
- **what:** `WriteSiteSpecAction` hard type-asserted `spec_data` to a map, rejecting the plain-string `mission_brief`/`roadmap_brief` the domain-submitter workflow resolves ("spec_data must be a JSON object, got string"). Fix: a coercion block — JSON string → parse; plain string → wrap as `{"text": value}` (matching the classifier prompt's `{{.site_specs.specs.mission_brief.text}}` read); objects pass through. The HANDOFF doc for this bug is also a worked example of the evidence-only handoff pattern (symptom carried, cause left to be read from code).
- **sources:** docs/RUNBOOK_vonc_session(1).md#1; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc(36).md#1
- **relations:** handoff document convention; data-shape/contract-drift debugging family
- **verify-later:** platform/orchestration/actions/site_spec_actions.go WriteSiteSpecAction coercion block

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-native idea engine (Phase D / Layer 4)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "the chassis version is one idea-orchestrator agent + one workflow reusing these [existing actions], NOT a port of engine.go. Did NOT write the SQL — needs a schema pass first (check-schema-before-SQL)."
- **what:** A mapped-but-unbuilt plan to express the idea-generation method as chassis actions rather than the standalone Go/Python engine: `execute_llm_prompt` for generate/cut/verify/score, `web_search`/`scrape_web`/`firecrawl_*` for verify, and — notably — `request_human_input`/`create_approval_request`/`await_approval`/`process_approval_decision` for the operator confirm+review gate, explicitly identified as "literally HITL." Distinguishes two shapes for applying the method to a domain: Shape A (the site IS the service, like idea.uk) vs Shape B (a static "request a report" page posting to one central service) — because the engine is server-side and minutes-long, it cannot be a forked `content_components` client-side tool the way other tools are.
- **sources:** `running_notes(44).md` ("Wrote the architecture & deployment guide; clarified hosting + OpenAI")
- **relations:** idea generation method; HITL (docs002_hitl_parallel); tool-lifecycle (contrast with deploy_tool_to_site)
- **verify-later:** whether an `idea-orchestrator` agent_definition or workflow exists

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Email identity in site_spec — deterministic address encoding + per-site `email` aspect
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md... Proposed `email` data... so a FUTURE email-provisioner agent (design only; catch-all makes it unnecessary now) can create per-domain forwarders."
- **what:** A proposed platform-wide convention for how any generated site gets an inbound/outbound email identity: a deterministic encoding (lowercase domain, `.`→`-`, `@<operator-domain>`, e.g. `agritec.uk` → `agritec-uk@leopardess.uk`), stored (not derived-on-read) to allow per-site overrides and to handle rare collisions; a new `email` aspect on `site_specs` (joining the existing classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance aspects) carrying status/address/from/reply_to/provider/forwards_to, reusing the spec's existing deployed/planned/blocked + feasibility-recheck state machine.
- **sources:** `running_notes(44).md` ("FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md")
- **relations:** site-spec-and-classifier (021 aspect list); catch-all email routing (superseded sub-concept, below)
- **verify-later:** whether `email` was actually added to the 021 aspect list; `EMAIL_identity_in_site_spec.md` (live doc)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Catch-all email forwarding — abandoned in favour of specific per-site forwarders
- **category:** site-spec-and-classifier
- **status-signal:** abandoned
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching"; and again "inbound still bouncing (No Such User); root cause = Default Address not forwarding" — two consecutive real-world failures of the originally-planned mechanism.
- **what:** The initial plan used a domain-level catch-all (cPanel "Default Address" / "Forward All Email for a Domain") so any `<encoded>@leopardess.uk` address would work without per-site setup. In practice this repeatedly bounced with "No Such User Here" because the mail backend delivers known mailboxes locally and only routes truly-unmatched addresses through the default address, which itself was misconfigured/pointed at the wrong of two confusingly similar domains (`leopardess.uk` vs `leopardess.co.uk`). Design refinement recorded explicitly: "prefer SPECIFIC per-site forwarders (created when a site is published) over a server catch-all — only forward addresses that exist, no backscatter, and it's exactly what the future email-provisioner agent does."
- **sources:** `running_notes(44).md` (two consecutive checkpoints, 2026-06-06)
- **relations:** Email identity in site_spec (the design this discovery feeds back into)
- **verify-later:** current leopardess.uk cPanel Default Address / Forwarders configuration

<!-- SOURCE: U25_leopardess_social.md -->
### Mission + roadmap as site_specs aspects (strategy-driven site intake)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 004_submit_vonc_trigger.sh ("Tier 3 submission: domain + mission + roadmap + briefs") exists and vonc.com was built from it; 003d specifies persist_mission/persist_roadmap via the existing write_site_spec action.
- **what:** Strategic context travels as input_data.mission (positioning, differentiators, tone, target users, core concepts, measurable objectives) and input_data.roadmap (phases with per-page purpose, section_types and content_context), persisted to site_specs aspects 'mission' and 'roadmap'. The classifier is told not to discover business type from the domain for mission-driven sites (site_type "interactive-platform"); the planner builds only the current phase and outputs section_types, not component names; content writers draw voice from mission and per-page content_context. Explicitly requires no new tables, no chassis code, no RAG for v1.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Approach, #What-goes-where, #Pipeline-changes; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh
- **relations:** component selector/creator; phase advancement loop; vonc.com v1 site
- **verify-later:** site_specs aspects mission/roadmap for vonc; intake-orchestrator/domain-submitter workflow steps

<!-- SOURCE: U25_leopardess_social.md -->
### Roadmap phase advancement and automated strategic review
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 003d "Phase advancement (later) … Manual for now"; 003 (earlier version) sketched the full automated loop ("scheduled agent … compare actuals vs targets … propose phase advancement") which later versions dropped to a one-liner.
- **what:** Phases advance by updating the roadmap aspect (current phase → complete, next → active) and re-triggering planning; measurable objectives in the mission aspect (DAU, completion rates, session duration, share rate) tell you when. The fuller vision — a scheduled strategic-review agent closing strategy → build → measure → adjust — was designed in 003 v1 and deferred; the delta is the record that it was consciously parked, not forgotten.
- **sources:** docs/social001_vonc_tiktok_social/003_spark_strategic_planning_architecture.md#Future-automated-strategic-review (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Phase-advancement
- **relations:** mission/roadmap aspects; traffic-analytics (the missing measurement half)
- **verify-later:** any analytics source; scheduler entries for strategic review (expect none)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dream spec / gap analysis / feasibility (one spec with status, not two documents)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 021 resolved decision 24 "One spec, not two"; 028: per-item status "not fully implemented yet — Phase 2"
- **what:** The full spec IS the dream; items carry status deployed/planned/blocked; gap analysis = blocked/planned subset; feasibility-recheck promotes blocked→planned when capability arrives; feasibility annotations prevent impossible work items. Older 002d dream_spec-in-content_data shape superseded by this. Phase 2 of 028 makes it mechanical.
- **sources:** 021#One Spec, Not Two; 028#The spec has status; 002d#Gap Analysis
- **relations:** fidelity dial; feasibility-recheck task
- **verify-later:** does feasibility-recheck scheduled task exist; per-item status columns

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Site spec unification (site_specs aspects as the one authoritative spec)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 021 target-state doc with immediate/short/medium ordering; backfill + fallback both recommended; newer pipeline already uses aspects
- **what:** One versioned spec per site as independent aspect rows (classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance), is_current/superseded_at per-aspect versioning with source/source_agent/source_item_id provenance; write_site_spec deep-merges so every row is a complete self-contained record (pruning-safe). content_data is legacy; read_site_spec falls back to it. Classifier writes intent, planner implements, design agent executes, audits enforce. The 15 content-strategy questions map onto aspects; deep research is a future classifier enrichment.
- **sources:** 021 full; P1#Site Specification System
- **relations:** 028 ownership; 030 strategic-vs-plan-time; dream spec
- **verify-later:** backfill done for legacy sites; read_site_spec fallback

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fidelity dial (locked/high/medium/low + no-adoption confidence mode)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 028 Phases: Phase 1 current (implicit high); explicit fidelity input is Phase 3; depends on per-item spec status (Phase 2)
- **what:** Controls how much aspiration reaches first deployment vs is deferred to the improvement loop, and the loop's promotion rate. locked = adopted-exact, no promotions; high (default with adoption) = faithful launch, ~one substantive change per audit cycle; medium = modest extensions; low = adoption as inspiration; blank domains reinterpret it as research-confidence tolerance. Lives on the trigger input + a build_policy/adoption_meta aspect.
- **sources:** 028#Fidelity, #Phased implementation
- **relations:** spec status; improvement loop rate
- **verify-later:** any fidelity/build_policy aspect in prod

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec aspect ownership and read-and-extend (anti-silent-overwrite)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028 ownership list; adoption-aware classifier prompt (migration 006) live; open question notes planner still writing design_intent pre-030
- **what:** Classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planners own site_plan. Classifier over adoption output is read-and-extend (preserve adopted dimensions, add strategy), never overwrite. Named failure modes: silent overwrite, confident fabrication on thin signal, default-to-brochure, reflexive upstream re-runs, schema-level commercial bias, adoption without strategic analysis.
- **sources:** 028#Who writes what, #Failure modes
- **relations:** 030 planner redirect to directives; composition self-heal
- **verify-later:** build-site-planner no longer writes design_intent/content_direction post-030

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Superseding a spec doesn't undo installed artefacts (re-queue rule)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028: gamesdesign remediation hit it (fresh specs + stale installed theme at sites.style_collection_id)
- **what:** Agents with install side-effects (composition, nav, pages, assets) write beyond site_specs and leave live pointers from long-lived tables; invalidating their spec must also queue the re-run work item (needs_composition etc.) — long-term the supersession itself should emit the recovery item. Test: does the agent write other tables AND does a live pointer reference the write?
- **sources:** 028#Failure modes (last)
- **relations:** install_site_composition; composition trigger matrix (027)
- **verify-later:** whether supersession-emits-item was ever built

<!-- SOURCE: U04_idea_uk.md -->
### Structured design_intent from the classifier (palette + typography reference_values)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Migration applied 2026-06-20; "Palette migration proven on a real build" 2026-06-21 — fresh idea.uk resolved `palette_source=design_intent_values`, parchment, "no invented blue".
- **what:** Root cause of generic-looking fresh builds: the classifier wrote design_intent colours as **prose** (hex buried in colour_mood sentences) while every consumer — the composition cascade and the analyze_design prompt — reads **structured** `design_intent.palette.reference_values` (8 slots) + `typography.reference_values`. The migration edits the classifier's classify_and_extract schema and adds a MANDATORY-fields bullet (all 8 slots as hex; style_direction must agree with the palette; never default to blue-and-grey), applied via snapshot_agent backup + exact-anchor replace() with a RAISE self-check. This single change is what makes both design stages agree (base = parchment, overlay starts from parchment).
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A + checklist); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md ("Direct consequence" section)
- **relations:** two-stage pipeline; mandatory-full overlay bug (this is precondition for its fix); prompt-migration discipline.
- **verify-later:** live classifier prompt contains the palette block; site 1244516d design_intent spec.

<!-- SOURCE: U04_idea_uk.md -->
### Phase 0 classifier-only positioning read
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** "Phase 0 result (2026-06-14 — ran, with the live site up)": faithful identity/classification/content_direction/design_intent specs, interactive-platform at ~0.91 confidence.
- **what:** Running just domain-research-classifier on a domain as a near-zero-cost positioning brief before committing to a build — its four spec aspects ARE the answer to "what does this site do for a stranger?". Caveats codified: a fresh read is NOT blank-slate (the classifier scrapes the live site up to 3 pages unless an adoption already ran); a generic name yields a generic name-only read, so hiding the live site removes signal not bias; a safe suppression trick exists (temporary blank nginx `location = /` — never touch DNS/nginx wholesale with a live Stripe webhook); the classifier's terminal needs_strategy item will flow into a full build if dispatch is running. Decision: leave the live site up.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 0); idea.uk/HANDOFF_chassis_site.md
- **relations:** standing-ambition mission (the read was faithful-but-backward-looking, which motivated it); fresh vs adoption.
- **verify-later:** idea.uk site 97ed2f64 spec rows (incl. the duplicate-spec tidy-up).

<!-- SOURCE: U04_idea_uk.md -->
### Build-standard classifier migration (best-in-class quality/fit, not scope)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "READY, PROVEN-CORRECT, NOT YET APPLIED… replace() simulated against the live prompt → lands cleanly" (2026-06-21); still unapplied in TODO P2 (2026-06-26).
- **what:** A migration prepending a "Build standard" block to the classifier prompt: every build should aim at best-in-class **quality and fit** for its vertical — explicitly not scope inflation. Its first draft failed on a multi-line anchor (whitespace mangling) and was fixed to a single-line anchor with a rollback proven clean — feeding the prompt-migration discipline. Test plan: fresh build first; confirm an adopted rebuild stays faithful rather than drifting.
- **sources:** idea.uk/HANDOFF(13).md (Migrations status); idea.uk/TODO_chassis_and_idea_uk(1).md#P2; idea.uk/running_notes(63).md (lll/mmm 06-20/21)
- **relations:** standing-ambition default; prompt-migration discipline.
- **verify-later:** whether migration_classifier_build_standard.sql was later applied (file itself lives outside this unit).

<!-- SOURCE: U05_content_quality_linking.md -->
### Guide as first-class page_type (classifier vocabulary + canonical URLs)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14j: "guides typed page_type=guide directly … migration_adoption_add_guide_page_type.sql worked".
- **what:** Guides were folded into blog-post by the site-adoption-agent's analyze_site enum, defeating query.pages_where_type:guide list resolution. Structural fix over band-aid: `guide` added to the adoption classifier enum + guidance, existing rows re-typed, URLs migrated /blog/guide-<slug>.html → /guides/<slug>/index.html (page_canonical.go already had the guide case). Classification geography documented: analyze_site LLM emits per-page page_type+url; site-classifier is site-type only; build-site-planner has a staler vocabulary but preserves existing pages verbatim; pages.page_type has a kebab-format check, no value allowlist.
- **sources:** running_notes_14(26).md#part-13-14c
- **relations:** Tier-D lists; adoption pipeline; bare-duplicate cleanup.
- **verify-later:** site-adoption-agent analyze_site prompt enum; pages.page_type values on a fresh adoption.

<!-- SOURCE: U14_docs019_runbooks.md -->
### site_type taxonomy drift between classifier and strategist (Q8)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B2 "vocabulary drift between hops — the classifier's site_type set (brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social) vs the strategist's canonical set (brochure|authority-portal|local-directory|review-site|content-hub|landing-page|portfolio). Two taxonomies for the same concept, one spec chain." Queued item 3.
- **what:** Two adjacent relay hops use different canonical vocabularies for the same site_type concept flowing through one spec chain — a contract-drift hygiene defect awaiting a one-canonical-set decision plus two snapshot migrations.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2 (Q8); docs019/RUNBOOK_builder_route(21).md#queue (item 3)
- **relations:** two front doors Q5; workflow result contract (drift class)
- **verify-later:** classifier and strategist prompt enumerations

<!-- SOURCE: U16_docs019_design_plans.md -->
### Classifier consolidation + site_type taxonomy alignment (queued work)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_builder_thread Q5/Q8 — a queued brief with a read-first plan, not executed in these files.
- **what:** Two classifiers overlap: domain-research-classifier (newer, relay-native) and site-classifier (pageflow/intake path, which does NOT use work-item triage — its differences may be load-bearing: intake's hitl_confirm_type keys off confirmed_type.recommended_builder). Brief: diff both rows, map dependency points, check hard before changing; merge additions both ways with snapshot migrations; deprecate only at zero usage. Behind it: the classifier and strategist use different canonical site_type vocabularies in the same spec chain — one decision, two snapshot prompt migrations.
- **sources:** HANDOFF_builder_thread.md#3; README_flows.md (Q8 note)
- **relations:** adoption-writes-first; vertical-exemplar hop
- **verify-later:** both classifier rows; intake usage query on orchestration_states

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Classifier as strategic brain (always runs full)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "The classifier … runs on every site entering the pipeline, and it always does its full job … Adoption does not shortcut it"; Phase 1 current, Phases 2–5 not implemented
- **what:** The `domain-research-classifier` decides what a site *should be* on every site; adoption/operator-mission are weighted inputs, not bypasses. It is not constrained to current capability — best-version items it can't build yet are marked `blocked` for `feasibility-recheck`. Silent override is the failure mode.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-classifier-is-the-strategic-brain-it-always-runs-in-full, WM/028_platform_mission_and_pipeline_direction.md#input-sources-and-their-weight, WM/028_platform_mission_and_pipeline_direction.md#phased-implementation
- **relations:** website mission; fidelity dial; spec-has-status; adoption pipeline
- **verify-later:** domain-research-classifier agent_definition; migration 006

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Fidelity dial (locked/high/medium/low)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "Fidelity … Five values, with high as the default when adoption evidence is present"; "depends on per-item status on specs … Phase 3"
- **what:** A dial controlling how much aspirational extension reaches the first build vs how faithfully it matches the strongest evidence (usually adoption): `locked` (exact, no promotion), `high`, `medium`, `low`; no-adoption reinterprets it as a confidence tolerance. Currently only implicit `high` (Phase 1).
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#fidelity-controlling-how-much-aspiration-reaches-the-first-build
- **relations:** spec-has-status; classifier strategic brain; adoption faithfulness locks
- **verify-later:** proposed adoption_meta/build_policy spec aspect

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Spec has per-item status — one spec, not two
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "one spec, not two. Items have status (deployed / planned / blocked) … It is not fully implemented yet … planned to be implemented in Phase 2"
- **what:** The dream is the full spec; the build is its non-blocked subset. Per-item status makes the dream-vs-build distinction mechanical — the build pipeline builds only `deployed`, `feasibility-recheck` promotes `blocked→planned`. Each spec row records source/source_agent/source_item_id for provenance; agents read-and-extend, never silently overwrite.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-spec-has-status-deployed-planned-blocked, WM/028_platform_mission_and_pipeline_direction.md#who-writes-what-who-doesnt-override
- **relations:** references doc 021; fidelity dial; feasibility-recheck
- **verify-later:** site_specs is_current/superseded_at; feasibility-recheck task

<!-- SOURCE: U18_sql_for_agents.md -->
### site-classifier → research-backed classification with domain_profile
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** 003 header: "Changes site-classifier from a single Haiku LLM guess into a research-backed orchestrator"; later 049 creates domain-research-classifier for the work-item pipeline, which takes over first-stage classification.
- **what:** Evolution of classification: v1 was one Haiku call mapping domain+objective to {landing, content, portfolio, brochure} + recommended_builder. v2 (file 003) made it an orchestrator: Haiku research brief → research-agent web investigation → Sonnet synthesis producing backward-compatible site_type plus a rich domain_profile (business identity, tone, visual_direction, image_guidance, strategic analysis). Explicit responsibility fences: does NOT pick pages or style_collection (planner's job) but DOES provide design inputs consumed by planner, image-generator, webdesign-agent, page-content-writer.
- **sources:** 003_site_classifier.sql; sql_for_agents_v1/003_site_classifier.sql; sql_for_agents_v2/000_backup_agents.sql
- **relations:** succeeded by domain-research-classifier (work-item pipeline); domain_profile is ancestor of site_specs aspects
- **verify-later:** live site-classifier definition vs domain-research-classifier; who consumes domain_profile today

<!-- SOURCE: U18_sql_for_agents.md -->
### domain-research-classifier (work-item first stage)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 049 header documents pipeline position "first agent after seed_build_queue"; 067 adds extended-thinking budget to its classify_and_extract step (conditional on patch deploy).
- **what:** Handler for needs_domain_research: researches a domain via web search and scrape, classifies site type, extracts identity signals, writes site_specs aspects "identity" and "classification", creates the next work item (needs_briefing; later needs_strategy per 060).
- **sources:** 049_domain_research_classifier.sql; 060_domain_strategist.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** successor of site-classifier v2; site_specs aspect model
- **verify-later:** current next-item wiring (strategy vs briefing)

<!-- SOURCE: U18_sql_for_agents.md -->
### spec-updater (mechanical site_specs merge from findings)
- **category:** site-spec-and-classifier
- **status-signal:** unknown
- **status-evidence:** 072 definition; no later patches in this unit.
- **what:** Handler for needs_spec_update items: applies {aspect, field, suggested_value} to site_specs with the WriteSiteSpecAction versioning pattern. No LLM. Description-only items complete as "needs human review". "The complexity is in the Go action, not in the workflow."
- **sources:** 072_spec_updater_agent.sql
- **relations:** content-gap-planner and audits emit its items; site_specs supersede-versioning
- **verify-later:** update_site_spec_from_item action

<!-- SOURCE: U19_sql_tables_components.md -->
### site_specs aspect-versioned specification store
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Table created in Phase-0 migration (019); live pg_dump backup (bk_site_specs.sql) shows the production shape including pinned; extensive backfills for real sites.
- **what:** All strategic site specification lives as (site_id, aspect, data JSONB) rows — identity, strategy, tone, design_intent, content_direction, growth_config, adoption_source — with provenance (source enum: classifier/adoption/hitl/planner/improvement/seed/manual/rollback/fork/recovery; source_agent; source_item_id) and history via is_current + superseded_at (unique current per site+aspect). write_site_spec deep-merges partials so each row is self-contained. `pinned` (Phase 4) prevents agents overriding human-set specs.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql; docs/agent_docs/sql_for_tables/bk_site_specs.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a-team-data
- **relations:** site_plans (operational counterpart); site snapshots capture current specs; identity enrichment (departments/team).
- **verify-later:** write_site_spec action; pinned enforcement in writers.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site classifier and site_type taxonomy
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** site-classifier agent SQL with the landing/content/portfolio/brochure taxonomy and recommended_group mapping, plus a template-flattening prompt fix (evidence it was actually run); taxonomy names the live classifier architecture (docs 021).
- **what:** A lightweight LLM agent classifying a project into site types — landing (conversion single-CTA), content (publishing/ads/SEO), portfolio (showcase), brochure/directory (multi-page business / listings) — with confidence, reasoning, detected signals, and a recommended builder group. The direct ancestor of the platform's archetype/classification system.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** successor: site-spec-and-classifier (classification architecture, archetype); briefing agent; intake orchestrator.
- **verify-later:** site-classifier agent row; current archetype enum vs this 4-type taxonomy.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Specialist architects per site type
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 023 SQL: landing-page-architect created (renamed copy of site-component-architect), content-site-architect created with content-site components (article grid, sidebar, ad zones, category nav), portfolio-architect created "for future use".
- **what:** One architect agent per site type, each with its own default sections and component_category filter into the library; the alternative "one architect, differentiated by build plan" was debated (025) and the group-per-project-type model won conceptually.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** groups as project recipes; conditional_call_agent routing.
- **verify-later:** the three architect rows; content-site component rows.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Site classifier agent
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/003 SQL defines 'site-classifier' with a Haiku prompt classifying landing/content/portfolio/brochure; docs015/004 confirms "Single LLM call → outputs ONE site_type... ONE recommended_builder"; current system uses the multi-aspect site-spec classifier (docs024 021).
- **what:** A lightweight LLM agent that classifies a domain+objective into a site type (landing/content/portfolio/brochure) with confidence, reasoning, detected industry and signals, and recommends the corresponding builder group. Its single-label output was later superseded by the richer site-spec aspect classification.
- **sources:** docs006_workflow_builder/003_current_state_of_agents.sql#2-SITE-CLASSIFIER; docs007_brochure_builder/001_brochure_builder_plan.md#Classification-Signals; docs015_data_flow_verification/004_builder_flow.md
- **relations:** intake orchestrator; HITL type confirmation; successor: site-spec-and-classifier architecture.
- **verify-later:** agent_definitions 'site-classifier' vs current classifier agents.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Intake orchestrator workflow (classify → brief → spawn builder)
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/011: "✅ WORKING SYSTEM: HITL Orchestration with Multi-Agent Workflow" listing the 11-step intake workflow with two HITL pauses; docs017/023 later introduces intake-orchestrator-v2 routing.
- **what:** The entry-point orchestration: spawn/call site-classifier → fetch_available_builders from DB → HITL confirm site type (human can override classifier and builder choice) → fetch builder questionnaire → briefing agent fills it → HITL review brief → spawn and call the chosen builder. Established the pattern of human quality-gates before expensive generation.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#Working-Agents; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Naming-and-Coexistence
- **relations:** site classifier; briefing agent; intake-orchestrator-v2; HITL protocol.
- **verify-later:** intake-orchestrator workflow JSON in agent_definitions; whether v2 routing exists.

<!-- SOURCE: U22_recent_small_docs.md -->
### Unified site spec (status-tagged single document)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** Doc lives under `docs022_domain_authority/old/`; "Extend the current classifier output ... Steps 1-3 can happen incrementally" — proposed, and archived.
- **what:** A proposal for the site-classifier to emit one unified spec covering classification, identity, design_intent, content_direction, pages, features, SEO, and maintenance_profile — every item tagged `status` (deployed/planned/blocked). The "dream" is the whole doc; the "build" is the non-blocked subset. Downstream agents (planner enriches rather than decides pages; design/content agents implement explicit intent; audit agents treat the spec as ground truth; HITL edits it).
- **sources:** docs022.../old/004_classifier_notes.md
- **relations:** site-classifier vertical/disposition output, feasibility/blocked-handler pattern, design_intent, HITL
- **verify-later:** site_specs.spec_type='unified_spec'; classifier identity/design_intent/content_direction fields

<!-- SOURCE: U22_recent_small_docs.md -->
### Feasibility / blocked-handler pattern
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "the claim action catches it ... Item marked 'blocked', error='Handler agent not registered' ... weeks later ... Feasibility-recheck task finds it ... promoted to 'triaged'." Describes an existing dispatch/claim mechanism.
- **what:** A pattern where planners/discovery emit work items naming an intended handler even if that agent doesn't exist yet. The claim action checks agent_definitions; unknown handlers mark the item `blocked` with an error; a periodic feasibility-recheck task promotes blocked items to `triaged` once the handler is deployed. A Go `check_feasibility` action can also pre-block at spec time based on the agent registry.
- **sources:** docs022.../old/004_classifier_notes.md#feasibility-assessment
- **relations:** unified site spec, work-item lifecycle, tool-lifecycle
- **verify-later:** claim action handler-existence check; feasibility-recheck scheduled task

<!-- SOURCE: U23_docs_root_vonc.md -->
### write_site_spec spec_data string coercion
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session: "FIXED, deployed"; migration table: code deployed 2026-06-24 ~15:00.
- **what:** `WriteSiteSpecAction` hard type-asserted `spec_data` to a map, rejecting the plain-string `mission_brief`/`roadmap_brief` the domain-submitter workflow resolves ("spec_data must be a JSON object, got string"). Fix: a coercion block — JSON string → parse; plain string → wrap as `{"text": value}` (matching the classifier prompt's `{{.site_specs.specs.mission_brief.text}}` read); objects pass through. The HANDOFF doc for this bug is also a worked example of the evidence-only handoff pattern (symptom carried, cause left to be read from code).
- **sources:** docs/RUNBOOK_vonc_session(1).md#1; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc(36).md#1
- **relations:** handoff document convention; data-shape/contract-drift debugging family
- **verify-later:** platform/orchestration/actions/site_spec_actions.go WriteSiteSpecAction coercion block

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-native idea engine (Phase D / Layer 4)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "the chassis version is one idea-orchestrator agent + one workflow reusing these [existing actions], NOT a port of engine.go. Did NOT write the SQL — needs a schema pass first (check-schema-before-SQL)."
- **what:** A mapped-but-unbuilt plan to express the idea-generation method as chassis actions rather than the standalone Go/Python engine: `execute_llm_prompt` for generate/cut/verify/score, `web_search`/`scrape_web`/`firecrawl_*` for verify, and — notably — `request_human_input`/`create_approval_request`/`await_approval`/`process_approval_decision` for the operator confirm+review gate, explicitly identified as "literally HITL." Distinguishes two shapes for applying the method to a domain: Shape A (the site IS the service, like idea.uk) vs Shape B (a static "request a report" page posting to one central service) — because the engine is server-side and minutes-long, it cannot be a forked `content_components` client-side tool the way other tools are.
- **sources:** `running_notes(44).md` ("Wrote the architecture & deployment guide; clarified hosting + OpenAI")
- **relations:** idea generation method; HITL (docs002_hitl_parallel); tool-lifecycle (contrast with deploy_tool_to_site)
- **verify-later:** whether an `idea-orchestrator` agent_definition or workflow exists

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Email identity in site_spec — deterministic address encoding + per-site `email` aspect
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md... Proposed `email` data... so a FUTURE email-provisioner agent (design only; catch-all makes it unnecessary now) can create per-domain forwarders."
- **what:** A proposed platform-wide convention for how any generated site gets an inbound/outbound email identity: a deterministic encoding (lowercase domain, `.`→`-`, `@<operator-domain>`, e.g. `agritec.uk` → `agritec-uk@leopardess.uk`), stored (not derived-on-read) to allow per-site overrides and to handle rare collisions; a new `email` aspect on `site_specs` (joining the existing classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance aspects) carrying status/address/from/reply_to/provider/forwards_to, reusing the spec's existing deployed/planned/blocked + feasibility-recheck state machine.
- **sources:** `running_notes(44).md` ("FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md")
- **relations:** site-spec-and-classifier (021 aspect list); catch-all email routing (superseded sub-concept, below)
- **verify-later:** whether `email` was actually added to the 021 aspect list; `EMAIL_identity_in_site_spec.md` (live doc)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Catch-all email forwarding — abandoned in favour of specific per-site forwarders
- **category:** site-spec-and-classifier
- **status-signal:** abandoned
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching"; and again "inbound still bouncing (No Such User); root cause = Default Address not forwarding" — two consecutive real-world failures of the originally-planned mechanism.
- **what:** The initial plan used a domain-level catch-all (cPanel "Default Address" / "Forward All Email for a Domain") so any `<encoded>@leopardess.uk` address would work without per-site setup. In practice this repeatedly bounced with "No Such User Here" because the mail backend delivers known mailboxes locally and only routes truly-unmatched addresses through the default address, which itself was misconfigured/pointed at the wrong of two confusingly similar domains (`leopardess.uk` vs `leopardess.co.uk`). Design refinement recorded explicitly: "prefer SPECIFIC per-site forwarders (created when a site is published) over a server catch-all — only forward addresses that exist, no backscatter, and it's exactly what the future email-provisioner agent does."
- **sources:** `running_notes(44).md` (two consecutive checkpoints, 2026-06-06)
- **relations:** Email identity in site_spec (the design this discovery feeds back into)
- **verify-later:** current leopardess.uk cPanel Default Address / Forwarders configuration

<!-- SOURCE: U25_leopardess_social.md -->
### Mission + roadmap as site_specs aspects (strategy-driven site intake)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 004_submit_vonc_trigger.sh ("Tier 3 submission: domain + mission + roadmap + briefs") exists and vonc.com was built from it; 003d specifies persist_mission/persist_roadmap via the existing write_site_spec action.
- **what:** Strategic context travels as input_data.mission (positioning, differentiators, tone, target users, core concepts, measurable objectives) and input_data.roadmap (phases with per-page purpose, section_types and content_context), persisted to site_specs aspects 'mission' and 'roadmap'. The classifier is told not to discover business type from the domain for mission-driven sites (site_type "interactive-platform"); the planner builds only the current phase and outputs section_types, not component names; content writers draw voice from mission and per-page content_context. Explicitly requires no new tables, no chassis code, no RAG for v1.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Approach, #What-goes-where, #Pipeline-changes; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh
- **relations:** component selector/creator; phase advancement loop; vonc.com v1 site
- **verify-later:** site_specs aspects mission/roadmap for vonc; intake-orchestrator/domain-submitter workflow steps

<!-- SOURCE: U25_leopardess_social.md -->
### Roadmap phase advancement and automated strategic review
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 003d "Phase advancement (later) … Manual for now"; 003 (earlier version) sketched the full automated loop ("scheduled agent … compare actuals vs targets … propose phase advancement") which later versions dropped to a one-liner.
- **what:** Phases advance by updating the roadmap aspect (current phase → complete, next → active) and re-triggering planning; measurable objectives in the mission aspect (DAU, completion rates, session duration, share rate) tell you when. The fuller vision — a scheduled strategic-review agent closing strategy → build → measure → adjust — was designed in 003 v1 and deferred; the delta is the record that it was consciously parked, not forgotten.
- **sources:** docs/social001_vonc_tiktok_social/003_spark_strategic_planning_architecture.md#Future-automated-strategic-review (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Phase-advancement
- **relations:** mission/roadmap aspects; traffic-analytics (the missing measurement half)
- **verify-later:** any analytics source; scheduler entries for strategic review (expect none)

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

<!-- SOURCE: U03_idea_uk_section_data.md -->
### idea.uk live-VM / chassis-staging duality
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Journal project facts: "idea.uk — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM … DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED." Every checkpoint ends "idea.uk live VM untouched."
- **what:** The genuinely site-specific arrangement underpinning the whole unit's risk model: the revenue-earning idea.uk product (£29 reports, live Stripe webhook, orders in a file, reserved paths /request /confirm /approve /decline /stripe/webhook /internal/* /order/*) is a separate Go binary on a Hetzner VM; the chassis-built idea.uk site deploys to Backblaze B2 while DNS still points at the VM — so all chassis work is invisible staging and the VM cutover is a separate future decision. Two chassis site_ids exist for idea.uk (97ed2f64-… in the June thread, 1244516d-… in the July thread) — treated as separate/earlier rows, confirm before relying on either.
- **sources:** running_notes_scheme_to_components(55).md#Project-facts; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; HANDOFF_idea_uk_differentiators_section_data.md#Key-facts
- **relations:** platform mission; chassis deploy model.
- **verify-later:** sites rows for idea.uk (both ids); DNS state if a cutover is ever planned.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk mission and identity (workshop of tools; never verdicts; warm-paper identity)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "idea.uk mission REFRAMED away from the £29 tool… shipped as a file" and submitted via --mission-file 2026-06-21; classifier "read the mission well" on site 1244516d.
- **what:** The site's genuinely site-specific concept: idea.uk = the place to take an idea seriously — a growing workshop of genuinely good tools (the main event; free + paid), tools labelled **private (in-browser, nothing sent)** vs **AI/hosted** with private leading; cutting-edge succinct research-grounded guides; a news section; **never verdicts** — perspective, evidence, and questions framed as opinion (the legal reframe); the £29 verified report demoted to one specialised flagship tool; later a build-and-bring-to-market service; preserve the warm-paper/ink/single-rust-accent/Fraunces+IBM-Plex editorial identity. Noted honestly: the privacy and latest-research promises are stated intent the chassis can't yet enforce.
- **sources:** idea.uk/KEY_DOC_idea_uk_mission.txt; idea.uk/HANDOFF(13).md (idea.uk current state); idea.uk/running_notes(63).md (nnn/ooo 06-21)
- **relations:** liability framework (never-verdicts); mission-file mechanism; standing-ambition.
- **verify-later:** site 1244516d mission spec content.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk chassis-site build state (two site rows; staging-only; gated go-live)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** Current position 2026-06-26: composition + stylesheet correct and verified (tool-portal-light + parchment, commit 05ef817, no LLM drift); pages still dark; rebuild + review + cutover gated on the P0 scheme thread.
- **what:** The concrete case-study state: first chassis run under site 97ed2f64 (2026-06-14: classifier→…→empty index → coordinator fix validated on it); old chassis site torn down and resubmitted fresh 2026-06-21 as site 1244516d with the mission file; re-resolved onto tool-portal-light 2026-06-25; deployed page defects catalogued (empty differentiators, unresolved CTAs, dead contact form, missing pricing spec, thin nav/footer, empty meta description, dark chrome). The live £29 VM tool is a separate stream, untouched and earning throughout — the safety property the whole thread leans on.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + current position); idea.uk/HANDOFF(13).md
- **relations:** every design/pipeline concept above (this is their test case); VM cutover.
- **verify-later:** current state of site 1244516d; whether the P0 thread completed and cutover happened.

<!-- SOURCE: U10_imagery.md -->
### robot-hands.com rebuild (testbed case study)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "Phase I0 ✅ COMPLETE. 33-page rebuild w/ live news (9 sources); 16 distinct per-page git-path heroes, zero expiring URLs" (2026-07-12).
- **what:** The adopted site's content layer was broken (10 zero-component pages, NULL content, features content in a pre-drift schema) while the imagery pipeline was correct — so it was rebuilt from scratch with news scope: supersede adoption-residue aspects, news-enable classification, add a mission_brief aspect, retire stale items, insert+manually-triage a needs_site_plan trigger, then a fresh 33-page plan (29 imagery rows) built unattended through dispatch. A 2026-05-20 audit first said "PATCH, do not re-plan" (foundation sound, build broken); the 2026-07-08 decision superseded it with full re-plan. Hard requirements: tools must actually work (deployed JS, resolving links) and it is the acceptance surface for all I-phases.
- **sources:** HANDOFF_robot_hands_rebuild.md, SQL_2026-07-08_robothands_rebuild_prep.sql, SQL_2026-07-08_robothands_mission_brief.sql, RUNNING_NOTES_imagery_best_in_class.md#Turns-2–4
- **relations:** news enrichment pattern; schema-contract drift bug (features {title,description} vs {icon,name,description}); orphan pre-rebuild pages cleanup still open.
- **verify-later:** current plan 7a40a0f9; page/component counts; tool pages' build_status.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dartsonline guides defect (benchmark bug, causes A/B/C)
- **category:** site-case-studies
- **status-signal:** deployed (diagnosis confirmed; fix deliberately not applied)
- **status-evidence:** "The benchmark bug (still live, still unfixed — deliberately)" (HANDOFF_CURRENT_fixloop.md)
- **what:** dartsonline.com published a Guides nav link to a blank page. Root mechanism, hand-diagnosed with citations: (A) `build-site-planner` populated `sections` for only 5 of 15 planned pages; (B) `page-build-handler`'s `check_has_ready_sections` routes sectionless pages to `complete_error`, which is `action: complete_workflow` — a success terminal — so the work item is marked complete though the page was never built; (C) `populate_nav_tables_action.go` filters nav candidates on `pages.status` (lifecycle, defaults 'active') instead of `build_status`, publishing links to unbuilt pages. Kept deliberately live and unfixed as a repeatable benchmark; the platform fix is known and can be applied by hand any time.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#benchmark bug
- **relations:** standing hypothesis (refuted); mark_no_sections gap; two intake paths disagreement; platform-not-site-data fix philosophy
- **verify-later:** pages/site_work_items rows for site 5fe8785b-223d-41a3-88ee-c07187622381; page-build-handler workflow JSON

<!-- SOURCE: U20_legacy_docs_a.md -->
### Robot Hands website — first agent-built multi-page site
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Working group definitions (robot-hands-website v1 with usage rows dated 2025-10-27/30), then robot-hands-complete-website (home/about/contact) with full workflow and trigger scripts.
- **what:** The platform's first end-to-end site build: an agent group spawning hero writer, image creator, about writer and contact writer; generating content and a Stability-AI hero image; assembling three HTML pages via aggregate_webpage with embedded CSS/nav; about page explicitly explains the site was built by AI agents (and "may be for sale"). Served as the proving ground for job topics, data helpers and image handling.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** aggregate_webpage; content-creator-about/contact agents; platform concepts evidenced: job topics, group workflows.
- **verify-later:** does robot-hands.com exist/what pipeline now owns it.

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### relojistas.com go-live + bot verdict
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Relojistas VERDICT from the access log: 14,961 reqs, 83% 404s … Human intent ≈ 0. Clean probe result (domain not worth building), not a measurement failure".
- **what:** First live domain: a Spanish watch FORUM (grounded in the Wayback snapshot), hand-made `relojistas-site/` (index + gracias, kind=search, THANKS_PATH=/gracias.html) to unblock go-live. After going live (and later Cloudflare-proxied), the access log showed overwhelmingly bot/crawler traffic (Chrome-spoof crawler, Claude-SearchBot, Semrush, Yandex) with ~0 human intent — a clean negative probe result. Later decided to static-build it anyway (RSS + crawler presence + 404/referer signal are assets).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** hand-instance of intent-probe component; drove the passive access-log harvest decision
- **verify-later:** deploy_setup/relojistas-site/{index,gracias}.html; relojistas_notes/relojistas_golive.md

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### wayfaringlondoner.com page + THANKS_PATH-is-engine-wide
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13 "wayfaringlondoner.com page built … a 2015–16 travel blog … BLOG framing"; "Design point — THANKS_PATH is engine-wide".
- **what:** Second hand-made page: a 2015–16 travel blog (Csilla; London + Bangkok/Transylvania/Jersey), BLOG framing asking for a destination/London spot/story, tagline gained "and under new ownership". Targets the SHARED multi-vhost box. Surfaced the constraint that THANKS_PATH is one engine-wide env var, so domains on a shared box must share a thanks filename (standard `/thanks.html`); relojistas keeps `/gracias.html` on its own box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** shared-box multi-domain onboarding
- **verify-later:** wayfaringlondoner-site/; wayfaringlondoner_notes.md (live)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Original first-domain set (dropped surgerylight + finance/retail)
- **category:** site-case-studies
- **status-signal:** abandoned
- **status-evidence:** runbook base §3 "Suggested first set: relojistas.com, wayfaringlondoner.com, surgerylight.com, plus one finance tool and one clear retail" — absent from runbook(12) §3 which names only relojistas.
- **what:** The earliest runbook proposed a concrete 3–5 domain starter set (relojistas.com, wayfaringlondoner.com, surgerylight.com, plus a finance tool and a clear retail), each grounded via Wayback. Later versions dropped the named list down to relojistas + wayfaringlondoner; surgerylight and the finance/retail candidates silently vanished.
- **sources:** traffic_probe_runbook.md#3, traffic_probe_runbook(2).md#3, traffic_probe_plan(11).md#risks
- **relations:** relates to Wayback grounding method
- **verify-later:** n/a

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk — AI ideation-as-a-service product
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "idea.uk runs as a single Go binary under systemd on a Hetzner box... Billing: Stripe Checkout — a single £29 payment per report, live and earning (proven end-to-end with a real card on 2026-06-14)."
- **what:** A paid tool at idea.uk that runs an internal ideation method (generate → cut → web-verify → score → rank) against a business domain + audience, producing a ranked report of business-idea candidates with citations. Sold as one-off £29 reports (down from an initial £199 concept) via a request-then-confirm flow with a free "audience-check" taster as the hook. Positioned as the dogfood/first customer of the idea-generation method itself.
- **sources:** `RUNBOOK_idea_uk(10).md`, `RUNBOOK_idea_uk(1).md`, `running_notes(44).md` (checkpoint 2026-05-28 pricing section, "Pricing settled for the idea.uk product")
- **relations:** idea generation method; Go engine supersedes Python; REVIEW_BEFORE_PAY billing flow; five-layer consolidation model
- **verify-later:** live idea.uk site; `idea-go/` module if present in the working tree; Stripe dashboard for the two named accounts

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Idea generation method — versioned pipeline (v0 → v3)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md`: v0 → v1 (durability factor + named free substitute) → v2 ("multi-lens generation + richer capability menu... audience-fit challenge... seller-bundles-support-free check") → v3 (Risk column added as a 6th factor, see separate entry). "Method v2 changes derived from the test" and "Method v2 changes — multi-lens generation" sections.
- **what:** The core reusable pipeline: generate (multi-lens: asset×capability, demand, generalist-failure, frontier, outcome) → cut (challenge against the *specific* free substitute + audience-fit challenge + seller-bundles-support-free check) → web-verify → score (Defensibility/Willingness/Buildability/Reuse/Durability, gate Def≥3 AND Will≥3) → rank. Each version fixed a concrete failure found by running the method against real domains (agritec.uk, gaswholesalers.com, robot-hands.com, websitedesign).
- **sources:** `running_notes(44).md` lines ~143-411 (method v1/v2 evolution), `idea_uk_method_v0` family (out of this unit's scope, referenced)
- **relations:** Risk-as-hazard scoring dimension; capability + event watchlists; moat/differentiator framework; cross-vendor critique
- **verify-later:** `idea_method_prompt.md`, `idea_uk_method_v0.md` (live), `idea-go/prompts.go`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Risk-as-hazard scoring dimension
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-05-28 (continued — Risk column added...)": "The rubric had no dimension for the consequence of being wrong. It was caught on operator instinct, which doesn't scale," followed by implementation in `idea-go/engine.go`/`prompts.go` and Python parity, "Built + vetted + tested clean."
- **what:** A 6th scoring factor (1-5, 5=safest) scoring *consequence of being wrong*, deliberately kept separate from the fitness sum (Def+Will+Build+Reuse+Dur) so it can't be gamed by high fitness. Risk=1 auto-drops the candidate into a separate "Dropped for operator risk" section; Risk≤2 still advances but flagged "⚠ needs liability work before building"; Risk is a rank tiebreaker at equal fitness. Triggered by a near-miss: SFI single-farm assessment scored a confident test-now recommendation that could have cost a farmer £5k-50k if wrong.
- **sources:** `running_notes(44).md` (Risk rubric table + rules), `LIABILITY_AND_TERMS.md` (referenced, live)
- **relations:** idea generation method; LIABILITY_AND_TERMS / legal pages
- **verify-later:** `idea-go/engine.go` `scored` struct, `idea_method_runner.py` parity implementation

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Go engine supersedes Python reference implementation
- **category:** site-case-studies
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(1).md` (archive) describes `idea-go/engine.go` + `prompts.go` + `service.go` + `store.go` + `billing.go` + `main.go` as the whole stack; the live `RUNBOOK_idea_uk.md` base file's equivalent table names only `idea_method_runner.py` / `idea_service.py` / `test_idea_flow.py`. `running_notes(44).md` confirms directly: "Ported the idea.uk tooling from Python to Go (platform is Go throughout)... The Python files remain as the reference implementation but Go is now the canonical version, consistent with the rest of the platform."
- **what:** idea.uk's engine + service were first built in Python (FastAPI, `idea_method_runner.py`, `idea_service.py`, sqlite via `test_idea_flow.py`), validated end-to-end (20/20 checks), then rewritten in idiomatic stdlib-only Go (no external deps, `go vet`/`go build`/`go test` all clean, 19/19 checks) to match "the rest of the platform," which is Go throughout. The rewrite preserved the id-based (not title-based) threading bug-fix and the cross-vendor cut design. This is a genuine, confirmed language-migration superseded/replaced-by relationship, one of very few in this corpus with byte-level before/after evidence.
- **sources:** `RUNBOOK_idea_uk(1).md` §pieces table vs live `RUNBOOK_idea_uk.md`; `running_notes(44).md` ("Ported the idea.uk tooling from Python to Go")
- **relations:** idea.uk product; idea.uk deployment topology
- **verify-later:** confirm whether `idea-go/` or the Python files are what's actually running in production today (the archive/live diff plus running_notes both say Go is canonical, but verify on the actual box)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Cross-vendor critique (multi-model critique step)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "Cross-vendor critique implemented (was the one untested multi-model claim). The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set... else falls back to a different Anthropic model." Later ported into the Go engine unchanged.
- **what:** The idea-generation method's "cut" (critique) step deliberately runs on a *different* model vendor than the "generate" step where possible (OpenAI if `OPENAI_API_KEY` is set, else a different Anthropic model as a same-vendor fallback), specifically to avoid "the same model marking its own homework." A stderr log line (`[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic (...)`) was added after user confusion about which vendor actually ran, so every run states its own critique provenance.
- **sources:** `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end", "Added a [cut] vendor log line to engine.go")
- **relations:** idea generation method
- **verify-later:** `idea-go/engine.go` `call_other_vendor` / cross-vendor branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk request-then-confirm intake with capacity throttle
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md` "Flow (request-then-confirm)"; `running_notes(44).md`: "Capacity throttle (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders in flight."
- **what:** The customer-facing order flow deliberately never takes payment until an operator has screened the request: submit (free) → operator `/confirm` (creates Stripe Checkout / or, post-REVIEW_BEFORE_PAY, runs the engine first) → `/decline` available at any point with a polite no-charge email → webhook-driven fulfilment → operator-reviewed delivery (`AUTO_DELIVER` off by default). A `MAX_ACTIVE_ORDERS` throttle returns HTTP 409 "at capacity" once too many orders are in flight, protecting a stated 72-hour turnaround promise; `/capacity` is a public endpoint so the page can show "currently full."
- **sources:** `RUNBOOK_idea_uk(1).md`; `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end")
- **relations:** REVIEW_BEFORE_PAY billing flow; idea.uk product
- **verify-later:** `idea-go/service.go` capacity/throttle logic

<!-- SOURCE: U25_leopardess_social.md -->
### Leopardess rebuild programme (phases L0–L9)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** PLAN phase table (2026-07-12, turn 13): "L0–L4 done, L5 partial, L6–L8 not started, L9 ongoing"; HANDOFF §3 "main pages live and verified".
- **what:** Rebuild of leopardessconsulting.co.uk (a site the platform built for itself) to be honest, well-branded and useful: evidence audit (L0), spec truth pass (L1), brand/logo (L2), palette fork (L3), 3-per-row layout (L4), copy rewrite (L5), explanatory imagery (L6), charts (L7), tools/guides/news build-out (L8), coherent deploy (L9). Includes the audience pivot (A2): sceptical, commercially-sharp, non-specialist buyer, with technical depth one click down.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Phases, docs/leopardessconsulting/HANDOFF.md#0, docs/leopardessconsulting/RUNNING_NOTES.md#Decision-log
- **relations:** claim-evidence audit rule; anti-hype voice spec; per-site style fork; chart component (Go SVG + JS)
- **verify-later:** site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732 pages/build_status; live leopardessconsulting.co.uk pages

<!-- SOURCE: U25_leopardess_social.md -->
### Claim-evidence audit rule ("no claim ships without an audit row")
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** AUDIT header 2026-07-09: "no claim ships unless it has a row in this table"; fabrication sweep 2026-07-10 "Result: CLEAN".
- **what:** A site-truth methodology: every marketing claim is verified against code, live Postgres or an HTTP response before it may appear on the site; unverifiable claims are removed or hedged explicitly (the one allowed unproven claim — "third busiest sports site" — is published labelled as recollection). Produced a verified capability inventory (C1–C6: Companies House cascade with 2,767 verified businesses; news pipeline 5,652 items/4,672 scored; tool-generation agent family; DB-defined hierarchical agents 143 defs/56 active/40 spawners; Banana+SDXL imagery; 8 own sites) and an UNSUPPORTED list (U1–U11).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#1, #2; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Standing-rules
- **relations:** LLM fabrication classes; verify-by-artifact operator discipline; per-category platform pipelines (Companies House, news feed, tool pipeline — this audit is dated deployment evidence for them)
- **verify-later:** business_intel.businesses counts; feed item counts; agent_definitions counts

<!-- SOURCE: U25_leopardess_social.md -->
### Reuse-not-rebuild site build-out with honest "simulation" labelling
- **category:** site-case-studies
- **status-signal:** aspirational
- **status-evidence:** PLAN L8 "not started — reuse existing tool library; surface the live news feed"; HANDOFF §6.5.
- **what:** The leopardess L8 plan: deploy/adopt existing interactive tool components (ROI estimator, quizzes, calculators — deterministic client-side widgets that must be labelled as simulations, not live inference), surface the real news-feed pipeline, pair guides with tools (tool-deployer already creates companion guides + cross-links), and note a "game" has no formal platform existence — it is simply a component_level='tool' component.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L8; docs/leopardessconsulting/HANDOFF.md#6
- **relations:** tool-library; news-feed-pipeline; dynamic-applications
- **verify-later:** tool component inventory; news feed surfacing on any site

<!-- SOURCE: U03_idea_uk_section_data.md -->
### idea.uk live-VM / chassis-staging duality
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Journal project facts: "idea.uk — LIVE Go service selling £29 reports; single binary under systemd on a Hetzner VM … DNS (Cloudflare) → the VM, so chassis B2 deploys are invisible to the live site. UNCHANGED." Every checkpoint ends "idea.uk live VM untouched."
- **what:** The genuinely site-specific arrangement underpinning the whole unit's risk model: the revenue-earning idea.uk product (£29 reports, live Stripe webhook, orders in a file, reserved paths /request /confirm /approve /decline /stripe/webhook /internal/* /order/*) is a separate Go binary on a Hetzner VM; the chassis-built idea.uk site deploys to Backblaze B2 while DNS still points at the VM — so all chassis work is invisible staging and the VM cutover is a separate future decision. Two chassis site_ids exist for idea.uk (97ed2f64-… in the June thread, 1244516d-… in the July thread) — treated as separate/earlier rows, confirm before relying on either.
- **sources:** running_notes_scheme_to_components(55).md#Project-facts; HANDOFF_scheme_to_components_for_claude_code(1).md#Environment; HANDOFF_idea_uk_differentiators_section_data.md#Key-facts
- **relations:** platform mission; chassis deploy model.
- **verify-later:** sites rows for idea.uk (both ids); DNS state if a cutover is ever planned.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk mission and identity (workshop of tools; never verdicts; warm-paper identity)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "idea.uk mission REFRAMED away from the £29 tool… shipped as a file" and submitted via --mission-file 2026-06-21; classifier "read the mission well" on site 1244516d.
- **what:** The site's genuinely site-specific concept: idea.uk = the place to take an idea seriously — a growing workshop of genuinely good tools (the main event; free + paid), tools labelled **private (in-browser, nothing sent)** vs **AI/hosted** with private leading; cutting-edge succinct research-grounded guides; a news section; **never verdicts** — perspective, evidence, and questions framed as opinion (the legal reframe); the £29 verified report demoted to one specialised flagship tool; later a build-and-bring-to-market service; preserve the warm-paper/ink/single-rust-accent/Fraunces+IBM-Plex editorial identity. Noted honestly: the privacy and latest-research promises are stated intent the chassis can't yet enforce.
- **sources:** idea.uk/KEY_DOC_idea_uk_mission.txt; idea.uk/HANDOFF(13).md (idea.uk current state); idea.uk/running_notes(63).md (nnn/ooo 06-21)
- **relations:** liability framework (never-verdicts); mission-file mechanism; standing-ambition.
- **verify-later:** site 1244516d mission spec content.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk chassis-site build state (two site rows; staging-only; gated go-live)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** Current position 2026-06-26: composition + stylesheet correct and verified (tool-portal-light + parchment, commit 05ef817, no LLM drift); pages still dark; rebuild + review + cutover gated on the P0 scheme thread.
- **what:** The concrete case-study state: first chassis run under site 97ed2f64 (2026-06-14: classifier→…→empty index → coordinator fix validated on it); old chassis site torn down and resubmitted fresh 2026-06-21 as site 1244516d with the mission file; re-resolved onto tool-portal-light 2026-06-25; deployed page defects catalogued (empty differentiators, unresolved CTAs, dead contact form, missing pricing spec, thin nav/footer, empty meta description, dark chrome). The live £29 VM tool is a separate stream, untouched and earning throughout — the safety property the whole thread leans on.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + current position); idea.uk/HANDOFF(13).md
- **relations:** every design/pipeline concept above (this is their test case); VM cutover.
- **verify-later:** current state of site 1244516d; whether the P0 thread completed and cutover happened.

<!-- SOURCE: U10_imagery.md -->
### robot-hands.com rebuild (testbed case study)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "Phase I0 ✅ COMPLETE. 33-page rebuild w/ live news (9 sources); 16 distinct per-page git-path heroes, zero expiring URLs" (2026-07-12).
- **what:** The adopted site's content layer was broken (10 zero-component pages, NULL content, features content in a pre-drift schema) while the imagery pipeline was correct — so it was rebuilt from scratch with news scope: supersede adoption-residue aspects, news-enable classification, add a mission_brief aspect, retire stale items, insert+manually-triage a needs_site_plan trigger, then a fresh 33-page plan (29 imagery rows) built unattended through dispatch. A 2026-05-20 audit first said "PATCH, do not re-plan" (foundation sound, build broken); the 2026-07-08 decision superseded it with full re-plan. Hard requirements: tools must actually work (deployed JS, resolving links) and it is the acceptance surface for all I-phases.
- **sources:** HANDOFF_robot_hands_rebuild.md, SQL_2026-07-08_robothands_rebuild_prep.sql, SQL_2026-07-08_robothands_mission_brief.sql, RUNNING_NOTES_imagery_best_in_class.md#Turns-2–4
- **relations:** news enrichment pattern; schema-contract drift bug (features {title,description} vs {icon,name,description}); orphan pre-rebuild pages cleanup still open.
- **verify-later:** current plan 7a40a0f9; page/component counts; tool pages' build_status.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Dartsonline guides defect (benchmark bug, causes A/B/C)
- **category:** site-case-studies
- **status-signal:** deployed (diagnosis confirmed; fix deliberately not applied)
- **status-evidence:** "The benchmark bug (still live, still unfixed — deliberately)" (HANDOFF_CURRENT_fixloop.md)
- **what:** dartsonline.com published a Guides nav link to a blank page. Root mechanism, hand-diagnosed with citations: (A) `build-site-planner` populated `sections` for only 5 of 15 planned pages; (B) `page-build-handler`'s `check_has_ready_sections` routes sectionless pages to `complete_error`, which is `action: complete_workflow` — a success terminal — so the work item is marked complete though the page was never built; (C) `populate_nav_tables_action.go` filters nav candidates on `pages.status` (lifecycle, defaults 'active') instead of `build_status`, publishing links to unbuilt pages. Kept deliberately live and unfixed as a repeatable benchmark; the platform fix is known and can be applied by hand any time.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#★ F0 PILOT, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 1, fixloop_eg_dartsonline/HANDOFF_CURRENT_fixloop.md#benchmark bug
- **relations:** standing hypothesis (refuted); mark_no_sections gap; two intake paths disagreement; platform-not-site-data fix philosophy
- **verify-later:** pages/site_work_items rows for site 5fe8785b-223d-41a3-88ee-c07187622381; page-build-handler workflow JSON

<!-- SOURCE: U20_legacy_docs_a.md -->
### Robot Hands website — first agent-built multi-page site
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** Working group definitions (robot-hands-website v1 with usage rows dated 2025-10-27/30), then robot-hands-complete-website (home/about/contact) with full workflow and trigger scripts.
- **what:** The platform's first end-to-end site build: an agent group spawning hero writer, image creator, about writer and contact writer; generating content and a Stability-AI hero image; assembling three HTML pages via aggregate_webpage with embedded CSS/nav; about page explicitly explains the site was built by AI agents (and "may be for sale"). Served as the proving ground for job topics, data helpers and image handling.
- **sources:** docs002_hitl_parallel/README.0100b.updated_state_of_play_for_creating_website.md; docs002_hitl_parallel/README.0100c.workflow_diagram.md; docs002_hitl_parallel/README.0100d.robot_hands_website_readme.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** aggregate_webpage; content-creator-about/contact agents; platform concepts evidenced: job topics, group workflows.
- **verify-later:** does robot-hands.com exist/what pipeline now owns it.

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### relojistas.com go-live + bot verdict
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Relojistas VERDICT from the access log: 14,961 reqs, 83% 404s … Human intent ≈ 0. Clean probe result (domain not worth building), not a measurement failure".
- **what:** First live domain: a Spanish watch FORUM (grounded in the Wayback snapshot), hand-made `relojistas-site/` (index + gracias, kind=search, THANKS_PATH=/gracias.html) to unblock go-live. After going live (and later Cloudflare-proxied), the access log showed overwhelmingly bot/crawler traffic (Chrome-spoof crawler, Claude-SearchBot, Semrush, Yandex) with ~0 human intent — a clean negative probe result. Later decided to static-build it anyway (RSS + crawler presence + 404/referer signal are assets).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** hand-instance of intent-probe component; drove the passive access-log harvest decision
- **verify-later:** deploy_setup/relojistas-site/{index,gracias}.html; relojistas_notes/relojistas_golive.md

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### wayfaringlondoner.com page + THANKS_PATH-is-engine-wide
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13 "wayfaringlondoner.com page built … a 2015–16 travel blog … BLOG framing"; "Design point — THANKS_PATH is engine-wide".
- **what:** Second hand-made page: a 2015–16 travel blog (Csilla; London + Bangkok/Transylvania/Jersey), BLOG framing asking for a destination/London spot/story, tagline gained "and under new ownership". Targets the SHARED multi-vhost box. Surfaced the constraint that THANKS_PATH is one engine-wide env var, so domains on a shared box must share a thanks filename (standard `/thanks.html`); relojistas keeps `/gracias.html` on its own box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** shared-box multi-domain onboarding
- **verify-later:** wayfaringlondoner-site/; wayfaringlondoner_notes.md (live)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Original first-domain set (dropped surgerylight + finance/retail)
- **category:** site-case-studies
- **status-signal:** abandoned
- **status-evidence:** runbook base §3 "Suggested first set: relojistas.com, wayfaringlondoner.com, surgerylight.com, plus one finance tool and one clear retail" — absent from runbook(12) §3 which names only relojistas.
- **what:** The earliest runbook proposed a concrete 3–5 domain starter set (relojistas.com, wayfaringlondoner.com, surgerylight.com, plus a finance tool and a clear retail), each grounded via Wayback. Later versions dropped the named list down to relojistas + wayfaringlondoner; surgerylight and the finance/retail candidates silently vanished.
- **sources:** traffic_probe_runbook.md#3, traffic_probe_runbook(2).md#3, traffic_probe_plan(11).md#risks
- **relations:** relates to Wayback grounding method
- **verify-later:** n/a

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk — AI ideation-as-a-service product
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & deployment (2026-06-10)": "idea.uk runs as a single Go binary under systemd on a Hetzner box... Billing: Stripe Checkout — a single £29 payment per report, live and earning (proven end-to-end with a real card on 2026-06-14)."
- **what:** A paid tool at idea.uk that runs an internal ideation method (generate → cut → web-verify → score → rank) against a business domain + audience, producing a ranked report of business-idea candidates with citations. Sold as one-off £29 reports (down from an initial £199 concept) via a request-then-confirm flow with a free "audience-check" taster as the hook. Positioned as the dogfood/first customer of the idea-generation method itself.
- **sources:** `RUNBOOK_idea_uk(10).md`, `RUNBOOK_idea_uk(1).md`, `running_notes(44).md` (checkpoint 2026-05-28 pricing section, "Pricing settled for the idea.uk product")
- **relations:** idea generation method; Go engine supersedes Python; REVIEW_BEFORE_PAY billing flow; five-layer consolidation model
- **verify-later:** live idea.uk site; `idea-go/` module if present in the working tree; Stripe dashboard for the two named accounts

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Idea generation method — versioned pipeline (v0 → v3)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md`: v0 → v1 (durability factor + named free substitute) → v2 ("multi-lens generation + richer capability menu... audience-fit challenge... seller-bundles-support-free check") → v3 (Risk column added as a 6th factor, see separate entry). "Method v2 changes derived from the test" and "Method v2 changes — multi-lens generation" sections.
- **what:** The core reusable pipeline: generate (multi-lens: asset×capability, demand, generalist-failure, frontier, outcome) → cut (challenge against the *specific* free substitute + audience-fit challenge + seller-bundles-support-free check) → web-verify → score (Defensibility/Willingness/Buildability/Reuse/Durability, gate Def≥3 AND Will≥3) → rank. Each version fixed a concrete failure found by running the method against real domains (agritec.uk, gaswholesalers.com, robot-hands.com, websitedesign).
- **sources:** `running_notes(44).md` lines ~143-411 (method v1/v2 evolution), `idea_uk_method_v0` family (out of this unit's scope, referenced)
- **relations:** Risk-as-hazard scoring dimension; capability + event watchlists; moat/differentiator framework; cross-vendor critique
- **verify-later:** `idea_method_prompt.md`, `idea_uk_method_v0.md` (live), `idea-go/prompts.go`

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Risk-as-hazard scoring dimension
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-05-28 (continued — Risk column added...)": "The rubric had no dimension for the consequence of being wrong. It was caught on operator instinct, which doesn't scale," followed by implementation in `idea-go/engine.go`/`prompts.go` and Python parity, "Built + vetted + tested clean."
- **what:** A 6th scoring factor (1-5, 5=safest) scoring *consequence of being wrong*, deliberately kept separate from the fitness sum (Def+Will+Build+Reuse+Dur) so it can't be gamed by high fitness. Risk=1 auto-drops the candidate into a separate "Dropped for operator risk" section; Risk≤2 still advances but flagged "⚠ needs liability work before building"; Risk is a rank tiebreaker at equal fitness. Triggered by a near-miss: SFI single-farm assessment scored a confident test-now recommendation that could have cost a farmer £5k-50k if wrong.
- **sources:** `running_notes(44).md` (Risk rubric table + rules), `LIABILITY_AND_TERMS.md` (referenced, live)
- **relations:** idea generation method; LIABILITY_AND_TERMS / legal pages
- **verify-later:** `idea-go/engine.go` `scored` struct, `idea_method_runner.py` parity implementation

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Go engine supersedes Python reference implementation
- **category:** site-case-studies
- **status-signal:** superseded
- **status-evidence:** `RUNBOOK_idea_uk(1).md` (archive) describes `idea-go/engine.go` + `prompts.go` + `service.go` + `store.go` + `billing.go` + `main.go` as the whole stack; the live `RUNBOOK_idea_uk.md` base file's equivalent table names only `idea_method_runner.py` / `idea_service.py` / `test_idea_flow.py`. `running_notes(44).md` confirms directly: "Ported the idea.uk tooling from Python to Go (platform is Go throughout)... The Python files remain as the reference implementation but Go is now the canonical version, consistent with the rest of the platform."
- **what:** idea.uk's engine + service were first built in Python (FastAPI, `idea_method_runner.py`, `idea_service.py`, sqlite via `test_idea_flow.py`), validated end-to-end (20/20 checks), then rewritten in idiomatic stdlib-only Go (no external deps, `go vet`/`go build`/`go test` all clean, 19/19 checks) to match "the rest of the platform," which is Go throughout. The rewrite preserved the id-based (not title-based) threading bug-fix and the cross-vendor cut design. This is a genuine, confirmed language-migration superseded/replaced-by relationship, one of very few in this corpus with byte-level before/after evidence.
- **sources:** `RUNBOOK_idea_uk(1).md` §pieces table vs live `RUNBOOK_idea_uk.md`; `running_notes(44).md` ("Ported the idea.uk tooling from Python to Go")
- **relations:** idea.uk product; idea.uk deployment topology
- **verify-later:** confirm whether `idea-go/` or the Python files are what's actually running in production today (the archive/live diff plus running_notes both say Go is canonical, but verify on the actual box)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Cross-vendor critique (multi-model critique step)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "Cross-vendor critique implemented (was the one untested multi-model claim). The runner's cut step now routes through OpenAI if OPENAI_API_KEY is set... else falls back to a different Anthropic model." Later ported into the Go engine unchanged.
- **what:** The idea-generation method's "cut" (critique) step deliberately runs on a *different* model vendor than the "generate" step where possible (OpenAI if `OPENAI_API_KEY` is set, else a different Anthropic model as a same-vendor fallback), specifically to avoid "the same model marking its own homework." A stderr log line (`[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic (...)`) was added after user confusion about which vendor actually ran, so every run states its own critique provenance.
- **sources:** `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end", "Added a [cut] vendor log line to engine.go")
- **relations:** idea generation method
- **verify-later:** `idea-go/engine.go` `call_other_vendor` / cross-vendor branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### idea.uk request-then-confirm intake with capacity throttle
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(1).md` "Flow (request-then-confirm)"; `running_notes(44).md`: "Capacity throttle (protects the 72h promise): MAX_ACTIVE_ORDERS caps orders in flight."
- **what:** The customer-facing order flow deliberately never takes payment until an operator has screened the request: submit (free) → operator `/confirm` (creates Stripe Checkout / or, post-REVIEW_BEFORE_PAY, runs the engine first) → `/decline` available at any point with a polite no-charge email → webhook-driven fulfilment → operator-reviewed delivery (`AUTO_DELIVER` off by default). A `MAX_ACTIVE_ORDERS` throttle returns HTTP 409 "at capacity" once too many orders are in flight, protecting a stated 72-hour turnaround promise; `/capacity` is a public endpoint so the page can show "currently full."
- **sources:** `RUNBOOK_idea_uk(1).md`; `running_notes(44).md` ("Built out the gaps + ran the flow end-to-end")
- **relations:** REVIEW_BEFORE_PAY billing flow; idea.uk product
- **verify-later:** `idea-go/service.go` capacity/throttle logic

<!-- SOURCE: U25_leopardess_social.md -->
### Leopardess rebuild programme (phases L0–L9)
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** PLAN phase table (2026-07-12, turn 13): "L0–L4 done, L5 partial, L6–L8 not started, L9 ongoing"; HANDOFF §3 "main pages live and verified".
- **what:** Rebuild of leopardessconsulting.co.uk (a site the platform built for itself) to be honest, well-branded and useful: evidence audit (L0), spec truth pass (L1), brand/logo (L2), palette fork (L3), 3-per-row layout (L4), copy rewrite (L5), explanatory imagery (L6), charts (L7), tools/guides/news build-out (L8), coherent deploy (L9). Includes the audience pivot (A2): sceptical, commercially-sharp, non-specialist buyer, with technical depth one click down.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Phases, docs/leopardessconsulting/HANDOFF.md#0, docs/leopardessconsulting/RUNNING_NOTES.md#Decision-log
- **relations:** claim-evidence audit rule; anti-hype voice spec; per-site style fork; chart component (Go SVG + JS)
- **verify-later:** site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732 pages/build_status; live leopardessconsulting.co.uk pages

<!-- SOURCE: U25_leopardess_social.md -->
### Claim-evidence audit rule ("no claim ships without an audit row")
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** AUDIT header 2026-07-09: "no claim ships unless it has a row in this table"; fabrication sweep 2026-07-10 "Result: CLEAN".
- **what:** A site-truth methodology: every marketing claim is verified against code, live Postgres or an HTTP response before it may appear on the site; unverifiable claims are removed or hedged explicitly (the one allowed unproven claim — "third busiest sports site" — is published labelled as recollection). Produced a verified capability inventory (C1–C6: Companies House cascade with 2,767 verified businesses; news pipeline 5,652 items/4,672 scored; tool-generation agent family; DB-defined hierarchical agents 143 defs/56 active/40 spawners; Banana+SDXL imagery; 8 own sites) and an UNSUPPORTED list (U1–U11).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#1, #2; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#Standing-rules
- **relations:** LLM fabrication classes; verify-by-artifact operator discipline; per-category platform pipelines (Companies House, news feed, tool pipeline — this audit is dated deployment evidence for them)
- **verify-later:** business_intel.businesses counts; feed item counts; agent_definitions counts

<!-- SOURCE: U25_leopardess_social.md -->
### Reuse-not-rebuild site build-out with honest "simulation" labelling
- **category:** site-case-studies
- **status-signal:** aspirational
- **status-evidence:** PLAN L8 "not started — reuse existing tool library; surface the live news feed"; HANDOFF §6.5.
- **what:** The leopardess L8 plan: deploy/adopt existing interactive tool components (ROI estimator, quizzes, calculators — deterministic client-side widgets that must be labelled as simulations, not live inference), surface the real news-feed pipeline, pair guides with tools (tool-deployer already creates companion guides + cross-links), and note a "game" has no formal platform existence — it is simply a component_level='tool' component.
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L8; docs/leopardessconsulting/HANDOFF.md#6
- **relations:** tool-library; news-feed-pipeline; dynamic-applications
- **verify-later:** tool component inventory; news feed surfacing on any site

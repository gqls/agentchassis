
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

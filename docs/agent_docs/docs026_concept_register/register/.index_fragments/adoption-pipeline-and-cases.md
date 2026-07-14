| ADO-001 | Infrastructure three layers (core/client delivery/framework builder) | partial | Layer 1 (factory) built, Layer 2 (client delivery) planned, Layer 3 future | adoption-pipeline.md |
| ADO-002 | Adoption is a one-off capture, not a ceiling (specs separation) | deployed | Crawl data goes to research_results, never site_specs; strategist extends beyond baseline | adoption-pipeline.md |
| ADO-003 | Site-adoption pipeline: wrapper, Go fingerprint, LLM classify, apply_adoption_plan | deployed | The core 16-step adoption mechanism: crawl, design fingerprint, LLM analysis, write specs+pages+items | adoption-pipeline.md |
| ADO-004 | Source vs destination separation (target_url / destination_domain) | deployed | Decouples crawled site from built site; mismatch used to silently drop content | adoption-pipeline.md |
| ADO-005 | Adoption variants A-D and the unwired selector | partial | reference/structure/clone/analysis modes defined but selector never wired | adoption-pipeline.md |
| ADO-006 | Adoption -> classifier handoff: adoption writes first, classifier consumes | deployed | Adoption never calls classifier directly; emits needs_domain_research only | adoption-pipeline.md |
| ADO-007 | Pattern extraction, code-as-reference, and RAG-fed generation | aspirational | Future pattern-extraction-agent mining research into reusable specs | adoption-pipeline.md |
| ADO-008 | Firecrawl capability escalation ladder | aspirational | executeJavascript/waitFor/structured json as upgrades to plain rawHtml parsing | adoption-pipeline.md |
| ADO-009 | Duplicate sites-row on re-adoption (open investigation) | unknown | Suspected duplicate site row on re-adopting an existing destination domain | adoption-pipeline.md |
| ADO-010 | Fresh vs adoption entry paths converge on one cascade | deployed | Fresh-build and adoption both converge on needs_domain_research | adoption-pipeline.md |
| ADO-011 | Adoption fidelity dial (locked/high/medium/low; phases 1-4) | partial | Only Phase 1 implicit-high exists; real per-item dial unbuilt | adoption-pipeline.md |
| ADO-012 | Readopt-as-acceptance-test pattern | aspirational | Tear down and re-adopt as the from-scratch acceptance test after a fix batch | adoption-pipeline.md |
| ADO-013 | Tool/game pages never deployed (A1): section-only parser + flip churn | deployed | Section-only HTML parser missed div-based tool output; fix verified in prod | adoption-pipeline.md |
| ADO-014 | Sectionless-page durability stack | partial | Sibling fallback + discovery check + no-sibling flag for zero-section pages | adoption-pipeline.md |
| ADO-015 | guide as a first-class page_type (adoption classifier) | deployed | Guides folded into blog-post then given own page_type + canonical URL | adoption-pipeline.md |
| ADO-016 | Interactive fingerprint extraction gap (Path C) | aspirational | Adoption misses script/canvas machinery; tools rebuild as prose | adoption-pipeline.md |
| ADO-017 | Adoption resume logic (never built) | abandoned | Mid-workflow resume plumbing exists but no subscriber; re-crawl is the answer | adoption-pipeline.md |
| ADO-018 | Single-agent adoption trigger, superseded by wrapper orchestrator | superseded | One-agent positional-domain trigger replaced by spawn->call wrapper | adoption-pipeline.md |
| ADO-019 | Unified design spec aspect, superseded by design_reference/design_intent split | superseded | Single blended design aspect replaced by concrete/semantic split | adoption-pipeline.md |
| ADO-020 | Two-stage -> three-stage adoption processing (historical evolution) | superseded | Go-only design-fingerprint stage inserted ahead of LLM classification | adoption-pipeline.md |
| ADO-021 | Section recipes for adoption | aspirational | Each section captured as purpose+structure+reference implementation recipe | adoption-pipeline.md |
| ADO-022 | Adopt-from vs deploy-to separation (unbuilt staging) | aspirational | No staging area distinct from live deploy target for adopted rebuilds | adoption-pipeline.md |
| ADO-023 | Adoption interactivity misroute - canonical-prefix key desync | deployed | buildPageFeatureMap keyed raw not canonical name, missing tool pages | adoption-pipeline.md |
| ADO-024 | Tool routing fix deployment status (T1+T2 deployed, symptom unconfirmed) | partial | Routing + detection fix live; widget-deploy acceptance criteria unverified | adoption-pipeline.md |
| ADO-025 | Adoption faithfulness - WriteSitePlanAction identity strip | partial | ValidateRoles/CanonicalisePage interaction strips identity on some page types | adoption-pipeline.md |
| ADO-026 | site-scraper (Firecrawl scrape -> site_context) | deployed | Ancestor design-transfer mechanism, standardized site_context schema | adoption-pipeline.md |
| ADO-027 | tool-recreation-handler | deployed | Two-stage analyze_tool/recreate_tool JS-heavy page recreation | adoption-pipeline.md |
| ADO-028 | 11-agent website analysis framework (legacy) | superseded | Original 4-group 11-agent web-capture master plan | adoption-pipeline.md |
| ADO-029 | website-analyzer conditional scraping group | deployed | Early smart capture entry point routing scrape/extract/crawl | adoption-pipeline.md |
| ADO-030 | Playwright capture adapter + website-capture agent | superseded | Deep browser capture deferred in favour of managed Firecrawl | adoption-pipeline.md |
| ADO-031 | Website-builder orchestrator (maximal clone-and-improve vision) | abandoned | Capture->vision->code->synthesis->content->library master workflow never built | adoption-pipeline.md |
| ADO-032 | Adopting existing external sites ("Adopt" workflow), legacy precursor | superseded | Learn loop against existing sites with component-match confidence scores | adoption-pipeline.md |
| ADO-033 | Site interrogation & pattern library | aspirational | Persistent unfulfilled idea: learn patterns from successful sites | adoption-pipeline.md |
| ADO-034 | Bare-guide / spurious duplicate pages from planner ignoring adopted state | deployed | Planner re-invents differently-slugged sibling pages; cleanup applied | adoption-pipeline.md |
| ADO-035 | Doc-tree adoption plan (category mismatch: docs retrieval, not site adoption) | aspirational | Constitution + tag/embedding retrieval plan for the doc corpus | adoption-pipeline.md |
| ADO-036 | Vertical-slice dogfooding of the automation ratchet (category mismatch) | aspirational | Walk one capability end-to-end before generalising the ratchet | adoption-pipeline.md |
| SPEC-001 | Dream spec / gap analysis / feasibility - one spec, not two | aspirational | Full spec is the dream; per-item status makes gap analysis mechanical | site-spec-and-classifier.md |
| SPEC-002 | Site spec unification: site_specs aspect-versioned store | deployed | One versioned spec per site as independent aspect rows with provenance | site-spec-and-classifier.md |
| SPEC-003 | Fidelity dial (locked/high/medium/low + no-adoption confidence mode) | partial | Only Phase 1 implicit-high fidelity exists at the platform level | site-spec-and-classifier.md |
| SPEC-004 | Spec aspect ownership and read-and-extend (anti-silent-overwrite) | deployed | Named owners per aspect; classifier reads-and-extends adoption output | site-spec-and-classifier.md |
| SPEC-005 | Superseding a spec doesn't undo installed artefacts (re-queue rule) | deployed | Invalidating a spec must also queue the re-run work item | site-spec-and-classifier.md |
| SPEC-006 | Structured design_intent from the classifier (palette/typography) | deployed | Fixed prose-buried hex colours so consumers read structured reference_values | site-spec-and-classifier.md |
| SPEC-007 | Phase 0 classifier-only positioning read | deployed | Running just the classifier as a near-zero-cost positioning brief | site-spec-and-classifier.md |
| SPEC-008 | Build-standard classifier migration (best-in-class quality/fit) | partial | Proven-correct prompt migration for quality standard, not yet applied | site-spec-and-classifier.md |
| SPEC-009 | Guide as first-class page_type (classifier vocabulary + canonical URLs) | deployed | Guide added to classifier enum, retyped, URLs migrated to canonical form | site-spec-and-classifier.md |
| SPEC-010 | site_type taxonomy drift between classifier and strategist | partial | Two canonical vocabularies for the same concept in one spec chain | site-spec-and-classifier.md |
| SPEC-011 | Classifier as strategic brain (always runs full) | partial | Classifier decides site destiny on every site; blocked items await feasibility | site-spec-and-classifier.md |
| SPEC-012 | Classifier lineage: v1 Haiku label -> v2 domain_profile -> domain-research-classifier | deployed | Three generations of site classification culminating in current agent | site-spec-and-classifier.md |
| SPEC-013 | spec-updater (mechanical site_specs merge from findings) | unknown | No-LLM handler applying suggested_value patches to site_specs | site-spec-and-classifier.md |
| SPEC-014 | Specialist architects per site type (legacy) | partial | One architect agent per site type with its own component filter | site-spec-and-classifier.md |
| SPEC-015 | Intake orchestrator workflow (classify -> brief -> spawn builder), legacy | superseded | 11-step HITL orchestration ancestor of the work-item relay | site-spec-and-classifier.md |
| SPEC-016 | Feasibility / blocked-handler pattern | partial | Unknown handlers block work items; feasibility-recheck promotes them later | site-spec-and-classifier.md |
| SPEC-017 | write_site_spec spec_data string coercion (bugfix) | deployed | Coercion block accepts plain-string mission/roadmap briefs as JSON objects | site-spec-and-classifier.md |
| SPEC-018 | Chassis-native idea engine (Phase D / Layer 4) | aspirational | Mapped-but-unbuilt plan to express idea-generation as chassis actions | site-spec-and-classifier.md |
| SPEC-019 | Email identity in site_spec (deterministic address encoding + email aspect) | aspirational | Proposed convention for per-domain inbound/outbound email identity | site-spec-and-classifier.md |
| SPEC-020 | Catch-all email forwarding, abandoned for per-site forwarders | abandoned | Domain-level catch-all repeatedly bounced; specific forwarders chosen instead | site-spec-and-classifier.md |
| SPEC-021 | Mission + roadmap as site_specs aspects (strategy-driven intake) | deployed | Strategic context persisted as mission/roadmap aspects, built vonc.com | site-spec-and-classifier.md |
| SPEC-022 | Roadmap phase advancement and automated strategic review | aspirational | Manual phase advancement now; automated strategy-review loop deferred | site-spec-and-classifier.md |
| DYN-001 | Dynamic applications direction (three tiers; thin generated backends) | aspirational | Static/dynamic components -> agent-powered backends -> full app generation | dynamic-applications.md |
| DYN-002 | Interactive fingerprint parse stage (C1-C6) | partial | Planned Go extractor for canvas/script/library signals feeding intent LLM | dynamic-applications.md |
| DYN-003 | Four-stage interactive-content pattern (parse/assess/generate/integrate) | aspirational | Reference shape for building any interactive content type | dynamic-applications.md |
| DYN-004 | Games as a content type (largest pipeline gap) | aspirational | No game generator/library/spec aspect exists; page_type missing | dynamic-applications.md |
| DYN-005 | Generator architecture convergence (shared interactive-artefact-generator) | aspirational | Shared base generator anticipated once games exist alongside tools | dynamic-applications.md |
| DYN-006 | Tool builder tiers (static / dynamic / application) | partial | Interactive functionality classified by creation risk; matured to tool-pipeline | dynamic-applications.md |
| DYN-007 | Runtime-fill mechanism (data-runtime-fill shells + client loaders) | deployed | Empty shells filled client-side from a JSON feed, proven three times over | dynamic-applications.md |
| DYN-008 | Two JS delivery paths + inline-script truncation bug class | deployed | Component js_content extraction vs js_snippets bundle, plus a truncation bug | dynamic-applications.md |
| DYN-009 | js_snippets library + render_js_snippets_for_site + site-asset-renderer | deployed | Library-wide JS behaviours bundled per site by applies_to overlap | dynamic-applications.md |
| DYN-010 | js-bundle-stale gap (site-asset-renderer not wired into ongoing builds) | aspirational | Snippet bundle only rebuilt at initial design/full rerender, not on change | dynamic-applications.md |
| DYN-011 | loader-builder agent + section descriptor + Tier E runtime-feed source | aspirational | Designed autonomy path from hand-built loaders to LLM-generated ones | dynamic-applications.md |
| DYN-012 | Generation-time guards for dynamic components (archive-list reference build) | deployed | Bakes runtime-fill lessons into generation instead of post-hoc repair | dynamic-applications.md |
| CASE-001 | idea.uk live-VM / chassis-staging duality | deployed | Revenue VM untouched while chassis builds deploy invisibly to B2 staging | site-case-studies.md |
| CASE-002 | idea.uk mission and identity (workshop of tools; never verdicts) | deployed | Site-specific brand concept reframing away from the single £29 tool | site-case-studies.md |
| CASE-003 | idea.uk chassis-site build state (two site rows; gated go-live) | partial | Concrete build history across two site_ids with catalogued page defects | site-case-studies.md |
| CASE-004 | robot-hands.com rebuild (testbed case study, 2026-07) | deployed | Broken content layer rebuilt from scratch as the imagery-pipeline acceptance surface | site-case-studies.md |
| CASE-005 | Dartsonline guides defect (benchmark bug, causes A/B/C) | deployed | Nav link to a blank page kept live deliberately as a fixloop benchmark | site-case-studies.md |
| CASE-006 | Robot Hands website - first agent-built multi-page site (2025-10) | deployed | The platform's first end-to-end agent site build, proving ground for job topics | site-case-studies.md |
| CASE-007 | relojistas.com go-live + bot verdict | deployed | Clean negative probe result: access log showed ~0 human intent | site-case-studies.md |
| CASE-008 | wayfaringlondoner.com page + THANKS_PATH-is-engine-wide | partial | Second probe page surfaced a shared-box thanks-filename constraint | site-case-studies.md |
| CASE-009 | Original first-domain set (dropped surgerylight + finance/retail) | abandoned | Early 5-domain starter set silently trimmed to two named domains | site-case-studies.md |
| CASE-010 | idea.uk - AI ideation-as-a-service product | deployed | Paid £29-report tool running the internal ideation method, live and earning | site-case-studies.md |
| CASE-011 | Idea generation method - versioned pipeline (v0 -> v3) | partial | generate->cut->verify->score->rank pipeline refined across four versions | site-case-studies.md |
| CASE-012 | Risk-as-hazard scoring dimension | deployed | 6th scoring factor for consequence-of-being-wrong, kept separate from fitness | site-case-studies.md |
| CASE-013 | Go engine supersedes Python reference implementation | superseded | idea.uk engine ported Python->Go to match the rest of the Go-throughout platform | site-case-studies.md |
| CASE-014 | Cross-vendor critique (multi-model critique step) | deployed | Cut step runs on a different model vendor than generate, logged explicitly | site-case-studies.md |
| CASE-015 | idea.uk request-then-confirm intake with capacity throttle | deployed | No payment until operator screens the request; MAX_ACTIVE_ORDERS throttle | site-case-studies.md |
| CASE-016 | Leopardess rebuild programme (phases L0-L9) | partial | Rebuild of the platform's own consulting site to be honest and useful | site-case-studies.md |
| CASE-017 | Claim-evidence audit rule ("no claim ships without an audit row") | deployed | Every marketing claim verified against code/DB/HTTP before it may ship | site-case-studies.md |
| CASE-018 | Reuse-not-rebuild site build-out with honest "simulation" labelling | aspirational | Deploy existing tool library, label deterministic widgets as simulations | site-case-studies.md |

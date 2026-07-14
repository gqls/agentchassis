# Cluster: imagery
Categories included: imagery, new:data-charts, new:component-asset-pipeline


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Imagery: per-page assets vs site-wide last-write-wins resolution gap
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** 016 §9 deployed-hero-images entry: assets deploy fine; resolution gap diagnosed with fix direction (page-aware ensureAssets + re-render through plan_sections)
- **what:** site_plan_imagery plans per-page keys; store_asset writes content_data[<purpose>_url] keyed by purpose so every page hero overwrites the last (single site-wide hero_url); first render bakes the use_fallback static path; terminal rerender/CSS fixers patch stored HTML without re-resolving. Fix: resolve per-page from site_plan_imagery⋈assets, keep content_data as gap-fill; after an asset lands flag needs_rebuild → needs_page at p99. Logo is chrome (render_site_components) — separate path. imagery kind/scope model + chk_kind constraint implied by site_plan_imagery columns.
- **sources:** 016 §9 hero/logo fallback entry; 002(4) flag_page_image_rebuild → page_rerender
- **relations:** page_rerender image_landed reason; input schema image fields rule
- **verify-later:** ensureAssets page-aware now?; site_plan_imagery schema

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Imagery loop closure plan (Phases 0–6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Decisions table (new imagery-quality-auditor agent; max 2 regen attempts; asset locking mirrors page_components; per-section granularity deferred); later docs show Phase 2G/2H verified 2026-05-15 and asset lock columns landed via migration 053
- **what:** The sequenced plan closing the gap between imagery asked for (specs/plans) and imagery delivered: Phase 0 wire imagery_direction into prompts + populate origin_model; Phase 1 algorithmic discovery checks (unfulfilled_image_prompt, placeholder_image_in_use, image_url_404) routed to the existing image-build-handler; Phase 2 assets locking + asset_key; Phase 3 adoption image mirror; Phase 4 text-only visual-auditor imagery awareness; Phase 5 vision-capable LLM path; Phase 6 imagery-quality-auditor. Explicit non-goals: per-section imagery_plan, icon resolver, infographic generator, provider router, img2img.
- **sources:** PLAN_imagery_loop_closure.md (whole); ASSESSMENT_imagery_phase_0_1…md
- **relations:** dispatch diagnostic (Phase 2G verification); adoption faithfulness locks; site_plan_imagery
- **verify-later:** which discovery checks exist under discovery_checks/; assets.locked_at/lock_type/asset_key columns

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### imagery-quality-auditor agent
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 6 of the plan; not reported deployed anywhere in scope (plan pre-dates 2026-05)
- **what:** A vision-capable sibling of visual-design-auditor: loads assets + imagery_direction (lock-honouring, excluding human uploads), runs a vision LLM audit with imagery-specific categories (direction_mismatch, brand_mismatch, inconsistency, quality, inappropriate), writes findings with max_fix_attempts 2 routing back to image-build-handler; counts toward the existing 3-pass audit cap; gated rollout via design-audit-agent.
- **sources:** PLAN_imagery_loop_closure.md#Phase-6
- **relations:** vision-capable LLM path; asset locking; design-audit-agent
- **verify-later:** agent_definitions for imagery-quality-auditor

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Asset locking mirrors page_components (+ asset_key multi-image readiness)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** planned in Phase 2; FOCUS_adoption_faithfulness (2026-05-19) states 053 migration adds lock_type + lock_expires_at "on page_components, site_components, site_plan_directives, assets … written, ready to apply"
- **what:** assets gains locked_at/lock_type (same vocabulary and exclusion predicate as page_components) so audits/discovery skip locked assets; asset_key column (default = purpose, unique per site) opens multi-image purposes (adoption mirror writes adopted:<filename>) without breaking existing single-purpose upserts; old purpose-unique index dropped only after asset_key bedding-in.
- **sources:** PLAN_imagery_loop_closure.md#Phase-2; FOCUS_adoption_faithfulness_via_locks(2).md#implementation-plan
- **relations:** lock policy table; adoption image mirror; per-page hero resolver (assets.asset_key join)
- **verify-later:** assets table columns and indexes

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Imagery algorithmic discovery checks
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Phase 1 planned three checks; dispatch doc shows `unfulfilled_imagery_plan` check (Phase 2G.4) live and emitting 8 items on robot-hands (2026-05-14) — with the pipeline='design' emission bug
- **what:** No-LLM checks catching spec-to-delivery gaps: image prompt in plan but no asset; hardcoded fallback path in rendered_html with no asset (the silent-failure case); referenced image URL with no matching asset. Emit needs_imagery/needs_hero_image/needs_logo items to image-build-handler via the dispatch loop.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1; FOCUS_dispatch_diagnostic(4).md#why-stuck, #Q4
- **relations:** pipeline soft label bug; baked-fallback problem
- **verify-later:** discovery_checks/check_unfulfilled_image_prompt.go etc.; unfulfilled_imagery_plan pipeline value

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### imagery_direction into image prompts + origin_model provenance (Phase 0/0.1)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "The Phase 0.1 deliverables stand" pending one verification query (assessment, undated ~2026-05); compatible with Phase 1 architecture which stabilises design_intent ownership
- **what:** Image generation reads site_specs design_intent.imagery_direction and prepends "Style direction: <direction>\n\nSubject: <prompt>" to the three-tier prompt; store_asset writes origin_model for provenance. The strategic-only read survives Phase 1 (per-page directives become the successor once site_plan_directives lands). Side benefit: pulls planner-invented hero prompts back toward the adopted look (partial mitigation of Bug 4).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0; ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md (whole)
- **relations:** planner ignores site_archetype imagery; image parameter-shaping
- **verify-later:** generate_image_actions.go composeImagePromptWithDirection; assets.origin_model population

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption image mirror
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 3 of the plan; "Adopted images are persisted but only as historical record for now" — no deployment claim in scope
- **what:** Stop discarding crawled imagery: mirror_adoption_images action downloads source images (capped count/size), uploads to S3, inserts assets rows (origin_type=adopted, asset_key=adopted:<filename>); wired into apply_adoption_plan plus a backfill discovery check and a one-step adoption-image-mirror agent. Future hook for img2img reference generation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3
- **relations:** asset_key; image parameter shaping (reference_image_uri)
- **verify-later:** mirror_adoption_images_action.go existence; assets rows with origin_type='adopted'

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Vision-capable LLM path
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 5 of the plan; no deployment claim in scope
- **what:** Extend aiservice with GenerateTextWithImages (Anthropic image content blocks, URL source), preferring extension of execute_llm_prompt with an image_urls_field config over a new action; presigned-URL freshness required; vision_call tagged in llm_call_log for cost tracking.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** imagery-quality-auditor (consumer); llm_call_log
- **verify-later:** platform/aiservice/anthropic.go vision support

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Image generation as parameter shaping (not prompt blending)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "The composer step is its own work" — recommendation recorded during Phase 2G step 5 design (undated, ~2026-05)
- **what:** Unlike text, images have a 77-token CLIP budget and no "don't" understanding — composition means deriving parameters (subject, negative_prompt from kind, style_preset/LoRA from imagery_direction, reference_image_uri from adopted images, aspect/cfg/steps per kind), not blending prose. A cheap compose_image_request step (Go rules or small LLM) producing a parameter envelope before image-generator is the candidate design; belongs with Phase 2H request-shape work.
- **sources:** FOCUS_prompt_composition_pattern.md#What-this-means-for-images
- **relations:** mega-prompt fragility (envelope pattern B); Phase 2H
- **verify-later:** image-generator request shape; any compose_image_request action

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site_plan_imagery sibling table
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** live JOIN in production code by 2026-06-02 ("site_plan_imagery.key = assets.asset_key"); write_site_plan step description updated to include imagery HITL-lock transfer (2026-05-26)
- **what:** Structured per-image plan rows (kind, key/asset_key, prompt, style hints, scope/scope_ref) mirroring site_plan_directives' scope+locking pattern — the successor to the legacy site_specs.site_plan.image_prompts dictionary; scoped page rows drive per-page heroes.
- **sources:** FOCUS_site_spec_vs_site_plan.md#where-imagery-lives; HANDOFF_2026-06-02…md#fix
- **relations:** per-page hero resolver; lock transfer; directive cascade
- **verify-later:** site_plan_imagery schema; emit_imagery_items writes

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Per-page hero resolver + rebuild-after-asset (baked-fallback fix)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "the fixes are in production" (2026-06-02) — plan_sections page-aware ensureAssets; flag_page_image_rebuild wired as image-build-handler terminal step (UPDATE 1 verified); registry entries "still required" at doc date
- **what:** Root cause triad: site-wide hero_url overwritten last-write-wins; async imagery completing after first render baked the on_missing fallback (/assets/images/hero.jpg) into rendered_html; terminal rerender reassembled stored HTML without re-planning. Fix: ensureAssets resolves this page's hero via site_plan_imagery JOIN assets (page scope) and site logo from site scope; flag_page_image_rebuild flags the page needs_rebuild and emits needs_page at priority 99 (dedup key page_rerender:<page>) so it re-resolves through plan_sections after its asset lands. Logo/header path deliberately out of scope (render_site_components, not plan_sections).
- **sources:** HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md#1
- **relations:** imagery checks; section-data reconciler (same handoff); two image-resolution paths open follow-up
- **verify-later:** registry.go entries for flag_page_image_rebuild/reconcile_section_data; hero component input_schema (field vs hardcoded template — open question)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Planner ignores site_archetype imagery constraints (Bug 4)
- **category:** imagery
- **status-signal:** unknown
- **status-evidence:** "site_archetype.design.imagery … says 'minimal icons/diagrams, no decorative photography'. The planner's site_plan still produced lavish hero prompts" (2026-04-23)
- **what:** The planner invents hero image prompts contradicting the adopted archetype's imagery stance. Fix shape: planner prompt reads site_archetype.design.imagery and sets needs_images=false when it says none/minimal. Phase 0.1's style-direction prepend partially mitigates the symptom.
- **sources:** HANDOFF_2026-04-23(1).md Bug 4; ASSESSMENT_imagery_phase_0_1…md#Bug-4
- **relations:** imagery_direction prompt composition; adoption faithfulness
- **verify-later:** plan_site prompt for archetype imagery constraint

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Section-scope imagery pipeline (plan → emit → generate → deploy → rebuild)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNBOOK §W7: "The flow already exists (05-26 handoff + code): `write_site_plan` → `site_plan_imagery` rows … → `emit_imagery_items` … → `needs_imagery` (priority ≤98) → `image-build-handler` … → `flag_page_image_rebuild` → `needs_page` (99) → rebuild resolves the URL"; W7b exercised it end-to-end (assets active at B2 in ~3 minutes each, 2026-07-03).
- **what:** The dynamic imagery supply chain: the planner writes `site_plan_imagery` rows (scope site/page/section; scope_ref `page:ordinal` for sections; key; kind hero/logo/icon/illustration/infographic; authored prompt — the table has NO description column; ordering; source), the gap-driven `emit_imagery_items` emits `needs_imagery` items only where no asset exists, image-build-handler's 25-step workflow generates, stores, brand-checks, spawns the asset-deployer (download S3 → optimise by purpose → commit to the site git repo, key-named files `_`→`-`.jpg), then `flag_page_image_rebuild` emits `needs_page` so plan_sections re-resolves the now-present asset. For idea.uk it never fired for the brief-explanation illustration simply because the planner never requested one (16 rows: 5 heroes, 10 icons, logo). Ordinal-based scope_refs drift when plans reorder (hygiene note; resolution is by key).
- **sources:** RUNBOOK_scheme_to_components(50).md#W7 #W7-0.3/0.4-RESULTS; w7b_01_imagery.sql; running_notes_scheme_to_components(55).md#Th #Tj #Ty #Tz
- **relations:** ensureAssets resolution gap; flag_page_image_rebuild section scope; presigned-URL expiry.
- **verify-later:** site_plan_imagery schema + emit_imagery_items step in build-site-planner; image-build-handler workflow steps.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### ensureAssets section-scope resolution gap (Edit B)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Tq): "Edit B LIVE (tools t)"; (Tv): "BRIEF-EXPLANATION CLOSED … illustration renders on index AND tools; Edit B fine throughout."
- **what:** The structural gap that made section illustrations unresolvable: `ensureAssets` (plan_sections_action.go) loaded only the page hero and site logo into the resolver's assets map (plus a legacy content_data fallback), so `site_assets.<key>` for section-scope imagery could never resolve — the pipeline's "last inch" was never wired. Edit B adds a third query (spi scope='section', scope_ref LIKE page||':%', joined to active assets), mapping BOTH by key (per-key schema paths like icon sets) AND by kind first-wins alias (generic `site_assets.illustration` paths), modelled on the hero block. The two-day "index miss" after deployment turned out to be a probe artefact (grep for the asset key string, but objects are UUID-named), taught as a debugging lesson.
- **sources:** gobatch_01_plan_sections.md#Edit-B; RUNBOOK_scheme_to_components(50).md#W7-CODE-FINDING; running_notes_scheme_to_components(55).md#Ti #Tt #Tu #Tv
- **relations:** section-scope imagery pipeline; plan_sections field deferral; probe-blindness (SQL pitfalls).
- **verify-later:** plan_sections_action.go ensureAssets section query; rendered brief-explanation `<img>` src on index/tools.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### flag_page_image_rebuild section-scope mapping (Edit H)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** RUNBOOK 07-06 night: "Slice 4c: PENDING APPLY — `gobatch_05_flag_section_scope.md` now carries Edit H … AND Edit H2"; the companion step-description SQL landed (Uj: "4c step description UPDATE 1") but the code edit awaited the next commit/image.
- **what:** `flag_page_image_rebuild` no-ops for non-page scope, so section-scope imagery landings never triggered the page rebuild that would surface them (observed live: zero flag-created needs_page in 30h after the two illustrations landed). Section scope_refs carry the page as a prefix (`index:1`), so the fix is a prefix-split: map scope 'section' to its page and fall through to the existing page path — no new emit code. Edit H2 + slice4c align the file header comment and the agent-definition step description with the new behaviour (cosmetic-drift discipline: descriptions must match deployed behaviour).
- **sources:** gobatch_05_flag_section_scope.md; slice4c_step_description.sql; running_notes_scheme_to_components(55).md#Tp #Ui #Uj
- **relations:** section-scope imagery pipeline; work-item crafting conventions.
- **verify-later:** flag_page_image_rebuild_action.go deployed body; image-build-handler flag_rebuild step description text.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Presigned-URL expiry and deploy-time asset localisation
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Ud) 2026-07-06: "w9_06: t/f/f both pages … localisation verified on content; Edit F deployed → recurrence prevention live → THREAD CLOSED"; w9_04 RETURNING "UPDATE 18, every url now /assets/images/…".
- **what:** A whole failure class: `assets.url` stored the presigned B2/S3 URL from generation (X-Amz-Expires=604800 — dies in seven days), while the asset-deployer had already committed the optimised file into the site repo under a key-derived local name. Renders that resolve from assets.url therefore embed URLs that die; heroes escaped only by being shadowed by a legacy local path. The fix is two-sided: w9_04 backfill flips all 18 idea.uk rows to `/assets/images/<key-hyphenated>.jpg`, preserving the unsigned S3 object path into `storage_path` (+ storage_provider), then a rebuild; Edit F makes `deploy_image_asset` record the committed local URL on the asset row at every future deploy, for ALL kinds (best-effort — a failure must not fail the deploy). Applies platform-wide to any site without the legacy hero_url shadow.
- **sources:** w9_03_assets_schema_and_inventory.sql; w9_04_backfill_flip.sql; gobatch_03_deploy_asset_localise(1).md; running_notes_scheme_to_components(55).md#Tu #Tw #Tz #Ua #Ub #Ud
- **relations:** legacy hero_url shadow; section-scope imagery pipeline; storage-architecture (S3/B2 refs preserved in storage_path).
- **verify-later:** deploy_image_asset_action.go post-commit UPDATE; assets rows url vs storage_path forms across sites.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Legacy site-level hero_url shadow (last-write-wins per purpose)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Tz/Ub): "hero_url lives at SITE level (sites.content_data …), merged beneath section data; template `or` picks it over the presigned background_image … last-write-wins per purpose → site-wide hero currently = hero-about's image everywhere."
- **what:** A legacy mechanism that both saved and distorts heroes: image deploys historically wrote `purpose+"_url"` keys (e.g. hero_url) into site-level `sites.content_data`, which the ContentData-priority merge (component_library.go ~736) supplies to templates ahead of the schema-resolved `site_assets.hero` value. Consequences: hero renders stayed local-path (immune to presigned expiry) but every page shows the same last-written hero image; the per-page hero assets sat unconsumed. Banked as a known quirk (per-page heroes = a later improvement); render-neutral to the localisation flip.
- **sources:** running_notes_scheme_to_components(55).md#Tx #Ty #Tz #Ub; w9_02_deployer_and_shadow.sql; w8_09_hero_exposure.sql
- **relations:** presigned-URL expiry; ensureAssets (content_data fallback is gap-fill, hero/logo only).
- **verify-later:** sites.content_data hero_url keys; component_library.go merge priority; whether per-page heroes were ever wired.

<!-- SOURCE: U05_content_quality_linking.md -->
### Hero image resolver (June-02) — images, not CTAs
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** running_notes_16(1) Part 1: "the 'hero resolver' (June-02) is about hero IMAGES, not CTA URLs".
- **what:** Per-page hero/logo image resolution in plan_sections ensureAssets (site_plan_imagery + assets join) plus flag_page_image_rebuild to re-render when an image lands — previously pages rendered the static /assets/images/hero.jpg fallback. Explicitly disambiguated from the CTA-URL bug (a costly early conflation). After Part 2, image_landed re-renders ride the no-LLM path.
- **sources:** running_notes_16_content_quality_and_internal_linking(1).md#part-1; NOTES(44) Part A; HANDOFF_2026-06-09(2).md#june-02-actions
- **relations:** no-LLM re-render path; site_plan_imagery pipeline (imagery unit).
- **verify-later:** plan_sections ensureAssets; flag_page_image_rebuild_action.go.

<!-- SOURCE: U09_adoption.md -->
### Imagery subsystem assessment (single hardcoded adapter and its gaps)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Part 1 assessment (descriptive) + "Phases 0 and 1 are complete and end-to-end verified… 2 onwards not started" (PLAN_imagery progress, 2026-05-08).
- **what:** One `DynamicImageAdapter` (Stability SDXL hardcoded, request body fixed — no negative prompt, seed, img2img, LoRA, variants); `assets` UNIQUE (site_id, purpose) blocks multi-image purposes; planner asks for exactly `{logo, hero_home}`; components declare only hero_image; features `icon` strings render nowhere; misleading non-consumed image-adapter config files; two image-generator agent rows (one placeholder). Enumerated structural gaps: multi-purpose model (asset_key), provider/model router mirroring the aiservice text pattern, richer request shape, planner imagery_plan, diverse input_schema image needs, icon/SVG path (three approaches), infographic-generator agent, design_intent.imagery_direction wiring, crawled-imagery persistence, per-vertical LoRA (018 plan, blocked on adapter model selection).
- **sources:** PLAN_imagery_loop_closure.md#part-1, old2/PLAN_imagery_loop_closure(1).md
- **relations:** imagery loop-closure phases; emit_imagery build-time trigger; adoption image mirror
- **verify-later:** internal/adapters/imagegenerator/dynamic_adapter.go; assets unique indexes; ImagePurposes map

<!-- SOURCE: U09_adoption.md -->
### Imagery loop closure phases 0–6 (spec-to-delivery audit-and-fix)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Phase 0.1/0.2 ✅ verified; 1.1–1.5 ✅ verified (Phase 1.5 needed an unplanned hotfix); 2 onwards not started" (2026-05-08).
- **what:** Sequenced plan: Phase 0 wire what exists (imagery_direction prepended to prompts; origin_model populated — verified); Phase 1 algorithmic discovery checks (unfulfilled_image_prompt, placeholder_image_in_use, image_url_404) routed to image-build-handler — verified, including the 1.5 hotfix (missing output_mapping made every dispatch-path image silently {stored:false}); Phase 2 asset locking (locked_at/lock_type mirroring page_components) + asset_key multi-image readiness; Phase 3 adoption-image-mirror (persist crawled imagery as origin_type='adopted' — today it is captured then discarded, "throwing away the best reference material we have"); Phase 4 text-only visual-auditor imagery category; Phase 5 vision-capable LLM path (GenerateTextWithImages); Phase 6 dedicated `imagery-quality-auditor` agent (sibling of visual-design-auditor, TOP-5 contract, max_fix_attempts 2, lock/origin_type honouring). Decisions locked: separate auditor agent; 2 regen attempts; mirror page_components locking exactly; per-section granularity deferred. Verification finding parked: dispatch wasn't claiming triaged imagery items while page items queued.
- **sources:** PLAN_imagery_loop_closure.md#part-2, old2/PLAN_imagery_loop_closure(1).md
- **relations:** emit_imagery_items (closed the build-time trigger gap the same shape as Phase-1 checks); C4 blank card images (imagery-to-card linkage open); 053 added lock cols to assets
- **verify-later:** discovery_checks check_unfulfilled_image_prompt.go etc.; assets.locked_at/lock_type/asset_key columns; imagery-quality-auditor existence (expected absent)

<!-- SOURCE: U09_adoption.md -->
### Build-time imagery trigger (emit_imagery_items + imageryplan.go shared selection)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "The imagery trigger had the identical bug and is fixed in the same deploy… deployed 2026-05-26, chassis v1.0.1047" (FOCUS_design_composition §3A).
- **what:** `write_site_plan` records planner image requests in `site_plan_imagery` (flattenImageryBlock) but nothing on the build path acted on them — needs_imagery came only from the loop's `unfulfilled_imagery_plan` check (capped 20/pass). `emit_imagery_items_action.go` emits at plan time; `imageryplan.go` is a shared package holding row selection, priority/severity classification, brand_update rule, item_key and spec body used by both the build emitter (status triaged) and the loop check (status detected) — the anti-drift pattern. Priority bands (index hero 65 … section 98) put imagery before the terminal needs_rerender (99). Known asymmetry: emit_imagery has no site-level no-backfill guard like emit_design's.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#3A, README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** imagery loop closure; work-site-orchestrator monolith mapping (imagery was the "same-shaped gap as design")
- **verify-later:** emit_imagery_items_action.go, imageryplan.go; site_plan_imagery rows on a fresh build

<!-- SOURCE: U10_imagery.md -->
### Imagery loop-closure programme (Phases 0–6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Progress (updated 2026-05-14) — Phase 2G + Phase 2H are operationally verified end-to-end on robot-hands.com … Phases 3, 4, 5, 6 of the outer plan not started"; Phase 3 "⏸ deferred 2026-05-14".
- **what:** The sequenced master plan for closing the gap between what the planner/spec asks for in imagery and what is delivered: Phase 0 (wire unread data), 1 (algorithmic discovery checks), 2A–2H (schema + pipeline refactor + structured plan imagery + request shape), 3 (adoption mirror), 4 (text-only auditor awareness), 5 (vision LLM path), 6 (imagery-quality-auditor). Each phase shippable alone; LLM phases gated on algorithmic checks working.
- **sources:** PLAN_imagery_loop_closure.md#Progress, PLAN_imagery_loop_closure.md#Phase-summary-table, STATUS_imagery_2026-05-12.md#At-a-glance
- **relations:** superseded in spirit by the best-in-class programme (I0–I8) which renumbered to avoid collision; feeds imagery-quality-auditor, adoption image mirror.
- **verify-later:** phases table vs live code: `platform/orchestration/actions/generate_image_actions.go`, `discovery_checks/`, `assets` schema, image-build-handler agent_definitions row.

<!-- SOURCE: U10_imagery.md -->
### imagery_direction prompt prepend (Phase 0.1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "0.1 — read imagery_direction ✅ delivered, verified 2026-05-08"; asset row shows origin_prompt beginning with the direction text.
- **what:** `generate_image` reads `site_specs.design_intent.imagery_direction` and prepends it to the subject prompt ("Style direction: … Subject: …", later unlabeled with a 200-char SDXL-aware sentence-boundary cap). Closed the "webdesign-agent writes imagery taste, image-generator ignores it" gap. Later superseded per-site by the imagery_style_guide when present (one brand voice, no double prepend).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0, old/PHASE_0_BUNDLE_README.md, STATUS_imagery_2026-05-08.md#Today's-verification
- **relations:** imagery_style_guide brand guide; per-kind prompt gating (direction gated OFF icons/logos).
- **verify-later:** `getImageryDirectionForSite` / `composeImagePromptWithDirection` in generate_image_actions.go; assets.origin_prompt on recent generations.

<!-- SOURCE: U10_imagery.md -->
### Asset provenance population (origin_prompt / origin_model)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "origin_model='sdxl' — Phase 0.2 column write happened" (2026-05-08); "origin_model propagation — assets.origin_model carries real provider/model" (2026-05-20 close-out).
- **what:** `store_asset` populates `assets.origin_prompt` (fixing a pre-existing bug where it was silently dropped — every row was NULL) and `assets.origin_model`; later extended so the adapter returns provider/model_id and workflows propagate it (`banana/gemini-3-pro-image-preview` vs `sdxl`) instead of a hardcoded literal. Provenance is the substrate for spec-vs-delivery audits.
- **sources:** old/PHASE_0_BUNDLE_README.md#Phase-0.2, TODO_imagery_followups.md#What-shipped-this-session, STATUS_imagery_2026-05-08.md
- **relations:** imagery discovery checks read it; imagery-quality-auditor (future) compares it to delivered image.
- **verify-later:** StoreAssetAction in v3_site_actions.go; `SELECT origin_model, origin_prompt FROM assets` distribution.

<!-- SOURCE: U10_imagery.md -->
### Algorithmic imagery discovery checks (Phase 1 trio)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Phases 1.1–1.4 all "✅ delivered" 2026-05-08 (1.2/1.3 "partial — check fires; symptom-path site needed").
- **what:** Three Go discovery checks catch spec-to-delivery gaps without LLM cost: `unfulfilled_image_prompt` (planner asked, no asset), `placeholder_image_in_use` (fallback path rendered, no asset), `image_url_404` (HTML references an image no assets row backs, DB-only version). All follow the DiscoveryCheck interface, register via init(), and were appended to design-discovery-agent's run_checks. A longer wishlist of ~12 further checks (alt-text, dimensions, orphans, cross-site contamination, multi-image underfill) was catalogued and mostly remains unbuilt.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1, FOCUS_imagery_assessment_1_.md#13.2, old/phase1/phase_1_register_imagery_checks.sql
- **relations:** check_unfulfilled_imagery_plan (2G successor for the new shape); image_source_unsatisfiable and component_template_corrupted (later siblings).
- **verify-later:** `platform/orchestration/actions/discovery_checks/check_*.go`; design-discovery-agent run_checks array.

<!-- SOURCE: U10_imagery.md -->
### asset_key multi-image model (Phases 2B–2D)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2B ✅ 2026-05-09; 2C ✅ deployed 2026-05-09; 2D ✅ applied 2026-05-10" with migration sanity checks documented.
- **what:** Broke the one-asset-per-purpose-per-site constraint: `assets.asset_key` column (backfilled from purpose), new unique index `(site_id, asset_key) WHERE active`, StoreAssetAction ON CONFLICT switched to asset_key, then the old `(site_id, purpose)` unique index dropped. Enables N heroes/icons/illustrations per site, with `(purpose, asset_key)` split (canonical hero = hero/hero; variant = hero/hero_about). Strict production apply order documented (2A→2B→2C deploy→verify→2D).
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2B/2C/2D, PLAN_imagery_loop_closure.md#Phase-2, old/phase2/2E/phase_2e_store_asset_action.diff
- **relations:** hero-variant routing (2E) consumes it; DeployedWebPath derives filenames from asset_key.
- **verify-later:** `\d assets` indexes; StoreAssetAction ON CONFLICT target.

<!-- SOURCE: U10_imagery.md -->
### Hero-variant routing through image-build-handler (2E)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2E ✅ delivered, verified 2026-05-12" with the full hero_about end-to-end trace.
- **what:** Made `hero_<page>` variants routable: `check_unfulfilled_image_prompt` classifies logo / hero_home / hero_<page> into needs_logo / needs_hero_image / unfulfilled_hero_variant; new `hasActiveAssetForAssetKey` helper (purpose-level check gave false positives); `deploy_image_asset` derives per-variant paths (`assets/images/hero-about.jpg`, `_`→`-`); StoreAssetAction gains `asset_key_field` JSONPath config; a third variant branch added to the image-build-handler workflow (spawn/call/store/deploy) leaving logo/hero branches untouched.
- **sources:** STATUS_imagery_2026-05-12.md#Phase-2E, old/phase2/2E/check_unfulfilled_image_prompt.go, old/phase2/2E/phase_2e_image_build_handler_variant_path.sql
- **relations:** needs_imagery branch (2G.5) later sits in front of it; known gap — variant chain doesn't pass site_id (imagery_direction not prepended for variants).
- **verify-later:** image-build-handler workflow branches in agent_definitions; imagery_helpers.go.

<!-- SOURCE: U10_imagery.md -->
### Spawned asset-deployer deploy pattern / storage-env isolation (2F)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2F ✅ deployed + verified 2026-05-12"; boxed warning 2026-07-10 "Where deploys run (by design, don't 'fix' this)".
- **what:** The chassis pod deliberately carries NO `IMAGE_BUCKET`, so it builds no storage client and inline `deploy_image_asset` fails there ("storage client not available") — by design. Deploys run in a spawned `asset-deployer` child into which `spawn_actions.go` injects S3/B2 env via the `isStorageEnabledAgent` list. 2F replaced three inline deploy step pairs in image-build-handler with spawn+call pairs targeting asset-deployer. Hand-triggering asset-deployer standalone fails because it skipped the injection — a triggering mistake, not a bug.
- **sources:** PLAN_imagery_loop_closure.md#2F, PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS, STATUS_imagery_2026-05-08.md#[BLOCKER]-Storage-architecture-mismatch
- **relations:** brand_head mode rides the same agent; storage-architecture (doc 032).
- **verify-later:** `agentbase/agent.go:294`, `spawn_actions.go` isStorageEnabledAgent, 107_image_build_handler.sql:725 comment.

<!-- SOURCE: U10_imagery.md -->
### site_plan_imagery table (2G.1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.1 — site_plan_imagery table ✅ delivered 2026-05-12"; live `chk_kind` re-verified 2026-07-08.
- **what:** Sibling of site_plan_directives holding structured imagery requirements: scope (site|page|section) + scope_ref, key (→asset_key), kind CHECK enum (logo|hero|illustration|icon|infographic, later +sprite_sheet), required prompt, JSONB style_hints/constraints, ordering, source CHECK (llm|classifier|manual|adoption), lock columns with the same lock-transfer treatment, unique on (plan_id, scope, COALESCE(scope_ref,''), key). `product` deliberately excluded (products come from the affiliate_products resolver, not the planner). Kind enum is mirrored in Go (`validImageryKinds`) — constraint and mirror change together.
- **sources:** PLAN_imagery_phase_2g.md#Schema, old/phase_2g_step1_site_plan_imagery.sql, SQL_2026-07-12_add_sprite_sheet_kind.sql
- **relations:** planner imagery block writes it; check_unfulfilled_imagery_plan reads it; five-place new-kind checklist.
- **verify-later:** `\d site_plan_imagery`; chk_kind vs validImageryKinds in write_site_plan_action.go (~line 183).

<!-- SOURCE: U10_imagery.md -->
### Planner imagery block (2G.3 prompt extension)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.3 ✅ delivered 2026-05-13 (with max_tokens bump + path fix)"; 2026-07-08 ground truth "planner prompt carries the Imagery Block + decomposition rule; max_tokens now 16000".
- **what:** build-site-planner's JSON output gains an `imagery` key (site[] / pages{} / sections{} entries with key, kind, prompt, optional style_hints/constraints) in the same LLM call as pages/design_direction — a single call, no separate imagery planner. Replaces the flat `image_prompts:{logo,hero_home}` contract that had hero/logo-only names baked in. max_tokens raised 4000→8000 (JSON truncation on a 14-page roadmap) and later to 16000. Legacy image_prompts continues to be emitted during transition.
- **sources:** PLAN_imagery_phase_2g.md#Planner-output-shape, PLAN_imagery_loop_closure.md#Application-status, FOCUS_imagery_assessment_1_.md#4.1
- **relations:** one-entry-one-image decomposition rule; planner key stability problem; sprite_sheet planner emission (future).
- **verify-later:** build-site-planner default_config prompt_template "## Imagery Block"; sql_for_agents/053 patches.

<!-- SOURCE: U10_imagery.md -->
### flattenImageryBlock write path + imagery lock transfer (2G.2)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "✅ deployed 2026-05-12; path fix on 2026-05-13 (function looked up data['imagery'] at top level rather than walking wrapper shapes via findDirectiveTree)".
- **what:** `write_site_plan` walks the planner's imagery block and inserts site_plan_imagery rows in the same transaction as pages/sections/directives (`flattenImageryBlock` + `insertImageryRow` enforcing the kind enum), and transfers locks from the previous current plan's locked imagery rows matched on (scope, scope_ref, key) — locked HITL prompt edits survive plan rebuilds.
- **sources:** PLAN_imagery_phase_2g.md#write_site_plan-extension, PLAN_imagery_loop_closure.md#2G
- **relations:** site_plan_imagery table; content-governance lock semantics.
- **verify-later:** write_site_plan_action.go flattenImageryBlock/insertImageryRow; lock-transfer behaviour on replan.

<!-- SOURCE: U10_imagery.md -->
### check_unfulfilled_imagery_plan discovery check (2G.4)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.4 ✅ delivered 2026-05-14 (8 work items emitted on first run; correct priority ordering)"; Pipeline:"build" fix confirmed in code 2026-07-08.
- **what:** Walks the current plan's site_plan_imagery rows, emits one `needs_imagery` work item per row lacking a matching active asset (via hasActiveAssetForAssetKey), capped at 20/pass, priority-banded (site logo 70 → index hero 65 → site other 75 → page hero 80 → page other 90 → section 100) mirroring legacy classifyPromptKey bands. `computeAssetKey` namespaces deeper keys (`page.about.illustration_team_values`, `section.home.2.icon_precision`) while keeping hero/logo names flat for backward-compatible deploy paths. Dedup key `needs_imagery:<scope>:<scope_ref|->:<key>`.
- **sources:** PLAN_imagery_phase_2g.md#Discovery-check-1, PLAN_imagery_loop_closure.md#Decisions/#2G, TODO_imagery_followups.md#7
- **relations:** legacy unfulfilled_image_prompt runs in parallel during transition (both call hasActiveAssetForAssetKey to avoid double work); pipeline-field fix.
- **verify-later:** check_unfulfilled_imagery_plan.go (hardcoded Pipeline "build"); design-discovery-agent run_checks.

<!-- SOURCE: U10_imagery.md -->
### needs_imagery branch in image-build-handler (2G.5)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.5 ✅ delivered 2026-05-14 (with two hotfixes — optional input_mapping + store_asset purpose)"; first asset a12b5d71 through the new path 2026-05-14.
- **what:** A new branch alongside (not extending) the variant chain: check_item_type_imagery → spawn_image_gen_imagery → call_imagery_gen (site_id passed so imagery_direction prepends; kind/style_hints/constraints pass through) → brand-update conditional store → shared spawn_asset_deployer tail. Brand-asset update routed by a `spec.brand_update` boolean computed at discovery (site scope OR index-page hero). Hotfixes established the `?`-suffix optional input_mapping convention and exposed that store_asset lacked `purpose_field` (initially hardcoding purpose:"hero", blocking kind=logo items — later fixed by the purpose_field workflow fix, 2026-05-20). A future refactor option is recorded: collapse the three legacy branches into needs_imagery ("always fix legacy with modern").
- **sources:** PLAN_imagery_loop_closure.md#2G/#Step-5-workflow, PLAN_imagery_phase_2g.md#image-build-handler-extension, TODO_imagery_followups.md#What-shipped-this-session
- **relations:** hero-variant branch (2E); `?` optional-mapping convention; purpose_field fix.
- **verify-later:** image-build-handler workflow JSON; store_imagery_asset purpose_field config.

<!-- SOURCE: U10_imagery.md -->
### Legacy image_prompts age-out check (check_legacy_image_prompts_aspect)
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** Old plan versions: "Migration off legacy image_prompts: Age-out via check_legacy_image_prompts_aspect, registered last"; live plan: "2G.6 ❌ retired as scoped. Reframed 2026-05-13… one string out of a JSON array, no code change."
- **what:** Originally a dedicated discovery check emitting `needs_replan` for sites still on `site_specs.site_plan.image_prompts` (deliberately registered LAST to avoid replan churn before the planner extension shipped). Reframed and retired: "is a site on legacy?" is not a fault signal — the existing checks already detect brokenness on both paths; migration became an operational deregistration decision (pull `unfulfilled_image_prompt` from run_checks once it reliably finds zero gaps).
- **sources:** old/PLAN_imagery_loop_closure(3).md#Decisions, PLAN_imagery_phase_2g.md#Discovery-check-2, PLAN_imagery_loop_closure.md#Decisions (2026-05-13 reframe)
- **relations:** superseded by "operational deregistration, not a dedicated check"; transition dual-check running.
- **verify-later:** confirm no check_legacy_image_prompts_aspect.go exists; whether unfulfilled_image_prompt is still registered.

<!-- SOURCE: U10_imagery.md -->
### pageflow-builder retirement
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** "Decision marker: 2026-05-12 — user agreed pageflow-builder is being left behind"; snapshot saved to pageflow-builder_2026-05-12.txt.
- **what:** The legacy monolithic site builder (inline deploy_image_asset, hardcoded generate_logo/generate_hero_image, sequential 20-iteration page loop, writes site_specs.site_plan directly bypassing the plan-domain tables) is deliberately not extended with the 2G imagery shape. Architecture converges on build-site-planner/plan-builder + triaged work items + page-build-handler + image-build-handler. Sites it built stay on the legacy check path until they age out; a full row snapshot exists as the recovery reference. The classifier's `recommended_builder` default was a noted loose end.
- **sources:** PLAN_imagery_phase_2g.md#On-leaving-pageflow-builder-behind, old/pageflow-builder_2026-05-12_NOTES(1).sql, old/pageflow-builder_2026-05-12.txt
- **relations:** superseded by plan-domain + dispatch architecture; robot-hands rebuild dropped its recommended_builder key.
- **verify-later:** pageflow-builder agent_definitions row status; any remaining live traffic.

<!-- SOURCE: U10_imagery.md -->
### Image-generator request shape + per-kind defaults (Phase 2H)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2H ✅ delivered (action layer) 2026-05-14 partial — chassis confirmed; adapter binary unconfirmed"; later provider work (2026-05-20) shipped the adapter side.
- **what:** Extends the generation request beyond {prompt,width,height}: v1 fields negative_prompt, seed, reference_image_uri (pass-through), cfg_scale, steps; Go-side `kindDefaults` map per kind (logos get people/text/watermark negative prompts; icons tighter aspect; heroes unchanged) with caller spec overriding defaults; style_hints.aspect_ratio drives dimensions and constraints feed the negative prompt. style_preset/samples/safety_mode deferred. Defaults deliberately live in Go, not a config table.
- **sources:** PLAN_imagery_loop_closure.md#2H, STATUS_imagery_2026-05-12.md#Phase-2H-(proposed), TODO_imagery_followups.md#4
- **relations:** provider abstraction; parseAspectRatio whitelist fix; constraints "informational only" decision.
- **verify-later:** kindDefaults/resolveKind/parseAspectRatio in generate_image_actions.go; adapter field mapping in dynamic_adapter.go.

<!-- SOURCE: U10_imagery.md -->
### parseAspectRatio SDXL v1.0 whitelist snap
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Elevated HIGH 2026-05-18 ("16:9 → 1024×576, SDXL rejects"); Turn 24 (2026-07-11) refers to "pre-SDXL-snap-fix residue", implying the fix landed.
- **what:** parseAspectRatio snapped to multiples of 64 rather than SDXL v1.0's strict dimension whitelist (1024×1024, 1152×896, 1344×768, …), so planner-emitted aspect_ratio hints produced rejected sizes and blocked hero generation — a regression enabled by the item-4 prompt patch (heroes previously fell through to valid kindDefaults). Fix: snap to the nearest whitelist pair matching the requested orientation.
- **sources:** TODO_imagery_followups.md#5, RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** Phase 2H request shape; planner prompt patch change 1 (aspect moved into style_hints).
- **verify-later:** whitelist logic in generate_image_actions.go; test 16:9→1344×768.

<!-- SOURCE: U10_imagery.md -->
### Adoption image mirror (Phase 3)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "3 — adoption image mirror ⏸ deferred 2026-05-14. Not cancelled… reference_image_uri plumbing preserved as a forward-compat hook."
- **what:** Stop discarding crawled imagery on adoption: new `mirror_adoption_images` action (download crawl images, upload to S3, insert assets rows with origin_type='adopted', origin_url, attribution/license; caps 50 images/site, 5MB each), wired into apply_adoption_plan; backfill check `check_crawled_images_discarded` routed to a new one-step `adoption-image-mirror` agent. Adopted images become img2img/style references and auditor signals. Deferred because current adopted sites carry minimal imagery.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3, FOCUS_imagery_assessment_1_.md#7/#9-item-9
- **relations:** reference-image style anchoring; adoption-pipeline category (site crawling).
- **verify-later:** existence of mirror_adoption_images_action.go, adoption-image-mirror agent row; assets rows with origin_type='adopted'.

<!-- SOURCE: U10_imagery.md -->
### Visual auditor imagery awareness (Phase 4, text-only)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "4 — visual auditor sees imagery (text-only) — not started".
- **what:** Extend visual-design-auditor's load_design_context SQL to include assets rows (unlocked, generated/adopted), imagery_direction, and site_plan_imagery, and add IMAGERY as a sixth check category with algorithmic-check results passed through to avoid double-flagging; tune on 5–10 sites before enabling fixes (≥80% accuracy target). Today the auditor's context contains zero image data — it cannot notice a missing or off-brief hero.
- **sources:** PLAN_imagery_loop_closure.md#Phase-4, FOCUS_imagery_assessment_1_.md#13.1/#13.3
- **relations:** imagery-quality-auditor (option B chosen as eventual answer); design-composition auditors.
- **verify-later:** visual-design-auditor load_design_context SQL in agent_definitions.

<!-- SOURCE: U10_imagery.md -->
### Vision-capable LLM path (Phase 5)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "5 — vision-capable LLM path — not started".
- **what:** Foundational capability for auditors to actually look at images: extend aiservice.AIService with image inputs, implement Anthropic vision content blocks, prefer extending execute_llm_prompt with an image_urls_field over a new action, refresh presigned URLs immediately before calls, tag vision_call:true in llm_call_log for cost separation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** required by imagery-quality-auditor and sprite-sheet vision auto-verify (I2.4/I8).
- **verify-later:** aiservice interface; anthropic.go vision support.

<!-- SOURCE: U10_imagery.md -->
### imagery-quality-auditor agent (Phase 6 / I8)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "6 — imagery-quality-auditor agent — not started"; I8 in the best-in-class plan also not started.
- **what:** A vision-capable sibling of visual-design-auditor under design-audit-agent, dedicated to imagery: categories direction_mismatch / brand_mismatch / inconsistency / quality / inappropriate; max_fix_attempts 2; findings route to image-build-handler regeneration (different prompt/seed/negative prompt) escalating to needs_human_review; honours locks and origin_type='uploaded'; gated rollout. I8 adds sprite-sheet cell verification and brand-guide reference comparison. Chosen over extending the existing auditor (separate TOP-5 cap; only imagery pays vision cost).
- **sources:** PLAN_imagery_loop_closure.md#Phase-6, FOCUS_imagery_assessment_1_.md#13.4, PLAN_imagery_best_in_class.md#Phase-I8
- **relations:** vision path (Phase 5); imagery_style_guide as the audit standard; improvement-loop pass caps.
- **verify-later:** no imagery-quality-auditor row in agent_definitions (expected absent).

<!-- SOURCE: U10_imagery.md -->
### Image provider abstraction and kind→provider routing
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 2026-05-20 close-out: "Provider abstraction — internal/adapters/imagegenerator/{provider,stability,banana}. dynamic_adapter.go routes by kind: icon → Banana, everything else → Stability. Proven end-to-end."
- **what:** The originally hardcoded Stability-only adapter (env-driven; the image-adapter.yaml/agent-definition config blocks are misleading and unread) was refactored into provider packages with kind-based routing: flat kinds (icon, later logo/illustration/infographic per a committed-but-then-pending routing change, and sprite_sheet) → Google Banana `gemini-3-pro-image-preview`; photographic kinds → Stability SDXL. The provider Request carries ReferenceImageURIs (Banana native reference-image support). Known opens: Stability provider timeout 60s vs old 120s; circuit breaker not threaded into provider clients.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session, FOCUS_imagery_assessment_1_.md#1.1–1.3, PLAN_imagery_best_in_class.md#2
- **relations:** icon model lessons drove it; 2H request shape; multi-provider routing beyond two providers still deferred.
- **verify-later:** internal/adapters/imagegenerator/ package layout; adapter switch cases; timeout value.

<!-- SOURCE: U10_imagery.md -->
### Icon generation lessons and image-model comparison
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Verdict (2026-05-18): Fix A is insufficient… SDXL ignores style instructions"; "Final icon batch state (verified 2026-05-20): all six icons banana/gemini-3-pro-image-preview, visual gate passed."
- **what:** SDXL is the wrong tool for flat-vector icons — strong photorealism bias on concrete subjects, multi-panel drift, no real transparency. A full model comparison (SDXL, SD3.5, FLUX schnell/dev/pro, DALL-E 3, Imagen 3, Nano Banana Pro 2, Midjourney, LLM-SVG) ranked FLUX schnell cheapest-good and Banana best for reference-conditioned sibling consistency; decision: plumb reference images AND switch icon generation to Banana. Related fixes: purpose_field so icons store as purpose=icon (240×240, not hero 1600×900); kindDefaults icon dimensions; jpg-vs-png note for thin line art.
- **sources:** TODO_imagery_followups.md#23, old/001_image_model_comparison.md, TODO_imagery_followups.md#Final-icon-batch-state
- **relations:** provider abstraction; transparency abandonment; LLM-SVG sleeper option; reference-image anchoring.
- **verify-later:** ImagePurposes["icon"]; assets rows origin_model for icon assets.

<!-- SOURCE: U10_imagery.md -->
### LLM-generated SVG icon path (sleeper option)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Sleeper option for later: LLM-generated SVG for icons… Worth a focused experiment once the immediate work is shipped" (2026-05-18).
- **what:** Icons are vector by nature; an LLM (Claude/GPT) writing SVG markup directly bypasses the entire convince-a-diffusion-model problem at ~$0.001–0.005/icon, crisp at any size, no copyright concern. Was the analyst's original recommendation (c2) before the user chose the Banana route; retained as a possible future replacement of the raster icon pipeline. Implies per-kind generation pipeline routing.
- **sources:** TODO_imagery_followups.md#23 (options c1/c2, recommendation), FOCUS_imagery_assessment_1_.md#9-item-6
- **relations:** superseded for now by Banana raster icons; Lucide covers UI chrome regardless.
- **verify-later:** none (idea only).

<!-- SOURCE: U10_imagery.md -->
### Diffusion transparency abandoned → flat-grey chip icons
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Transparency abandoned as too fragile — image models paint a transparency checkerboard into RGB (confirmed: icon_cycle_time mode=RGB has_alpha=False). Decision: option 2, 'embrace the box'" (2026-05-20).
- **what:** Image models cannot produce true alpha; requesting transparent backgrounds yields painted checkerboards. Locked decision: icons generate on a flat selectable grey background (#EEEEEE bg / #4A4A4A line) and are presented inside a styled CSS chip; the planner prompt and all existing icon specs were patched accordingly. The lesson recurs for sprite sheets (flat selectable background, NOT transparent).
- **sources:** TODO_imagery_followups.md#Icon-background-resolution, SCOPE_I2_sprite_sheets.md#3 (planner prompt), CONTEXT_PACK_imagery_sprite_sheet.md#Attach—docs
- **relations:** icon lessons; sprite-sheet prompt rules; CSS chip styling was left as site-template work.
- **verify-later:** planner prompt icon-background wording; icon assets' actual backgrounds.

<!-- SOURCE: U10_imagery.md -->
### Per-kind prompt gating and the five-place new-kind checklist
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "directionAppliesToKind… gates design_intent.imagery_direction OFF icons/logos" (shipped 2026-05-20); "PROVEN on real generations (icons carried palette, not medium)" (2026-07-11); sprite gating fix commit 4629aa17 proven in origin_prompt (Turn 31).
- **what:** Photographic brand direction contaminates non-photographic kinds (prepending it to an icon prompt makes the model paint a photo around the icon), so prompt composition gates per kind: hero/illustration/infographic get medium+mood+palette; icon and sprite_sheet palette only; logo nothing. Two gating functions (`directionAppliesToKind`, `styleGuide.directionForKind`) plus the DB constraint, Go mirror, adapter switch and ImagePurposes form the standing five-place checklist any new imagery kind must touch — the I2.0 lesson.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session, HANDOFF_imagery_best_in_class.md#Mechanisms, RUNNING_NOTES_imagery_best_in_class.md#Turn-29
- **relations:** imagery_style_guide supplies the gated content; sprite_sheet contamination near-miss is the cautionary case.
- **verify-later:** both gating functions list identical kind sets; grep for the five places when a new kind exists.

<!-- SOURCE: U10_imagery.md -->
### One-entry-one-image decomposition rule (planner prompt patch)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** planner_prompt_patch changes applied ("planner prompt carries the Imagery Block + decomposition rule", verified live 2026-07-08).
- **what:** A work item describes one deliverable: prompts must never ask for "a set of six icons" (SDXL renders one six-panel image — unusable but superficially successful). The planner prompt teaches per-entry single-image prompts, bans plural/counting phrasing (RULE 16), biases toward over-decomposition (unused icons are cheap, botched multi-panels expensive), moves aspect ratio to style_hints.aspect_ratio (the key Go reads) and demotes constraints to "informational only, reserved". The icon_cross_technology six-panel artifact and its cleanup SQL are the canonical example.
- **sources:** old/planner_prompt_patch_imagery.md, TODO_imagery_followups.md#25/#4
- **relations:** planner imagery block; SDXL whitelist fix; multi-entry sections remain the canonical way to express multiple images at one scope.
- **verify-later:** RULE 16 in the live planner prompt; absence of "set of" in site_plan_imagery.prompt.

<!-- SOURCE: U10_imagery.md -->
### Planner key stability across replans
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Symptom (2026-05-15): previous plan keyed hero_canonical; new plan called it brand_hero_canonical… discovery emitted a fresh work item"; no fix recorded since.
- **what:** The planner LLM freely chooses imagery `key` values, so replans rename equivalent concepts, discovery sees missing assets, and generations/orphan assets accumulate. Fix options ranked: (a) pass old plan's keys into the prompt with a reuse rule (lowest effort), (b) canonical key dictionary, (c) semantic concept matching at discovery time. Stale keys from this bug were cleaned up during the best-in-class rebuild.
- **sources:** TODO_imagery_followups.md#26, RUNNING_NOTES_imagery_best_in_class.md#Turn-24 (stale failed rows closed)
- **relations:** planner imagery block; replan-driven waste; asset orphan cleanup.
- **verify-later:** whether the planner prompt includes previous-plan keys; duplicate-concept assets on replanned sites.

<!-- SOURCE: U10_imagery.md -->
### Lucide icon strategy and validator wiring
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Lucide validator written (lucide_icons.go) — NOT yet wired" (2026-05-20), re-verified unwired 2026-07-08; still listed as an open I0 close-out item.
- **what:** The features grid renders icons as Lucide webfont glyphs (`<i data-lucide="{{.icon}}">`), not generated raster — the generated icon pipeline was never the right tool for it. Missing icons are LLM-invented Lucide names. Fix design: a single-source allowlist that is both the prompt's choice list and a pre-store `SanitizeFeatureIcons` sweep, plus optional render-time net; the allowlist must be verified against the bundled Lucide version. Blocked on identifying the content-generation step that fills features content_data. Icon strategy stays dual (D6): Lucide for UI chrome, generated sprites/raster for decorative glyphs.
- **sources:** TODO_imagery_followups.md#features-component-icons, old/verify_and_wire_lucide.md, PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** sprite sheets cover decorative glyphs; robot-hands rebuild was to carry the wiring.
- **verify-later:** callers of SanitizeFeatureIcons/ValidateLucideIcon outside lucide_icons.go.

<!-- SOURCE: U10_imagery.md -->
### Data-graph / chart pipeline (code-rendered, never diffusion)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Status: scoping only — not built" (2026-05-20); I4 "not started" as of 2026-07-12; runtime decision confirmed 2026-07-08 (go-echarts in-chassis).
- **what:** Hard constraint: diffusion models cannot plot real data — they fabricate values. Charts are a separate three-stage pipeline: fetch real series (EIA/FRED/per-vertical free-tier sources, stored for reproducibility + attribution) → code-render (go-echarts; static SVG/PNG always exists as fallback) → LLM editorial layer only (titles, callouts, annotations — never data values). Needs a `data-chart-generator`-shaped agent and deliberately does NOT add `chart` to site_plan_imagery kinds (charts are Lane B artefacts); `infographic` stays decorative-Banana and must never carry real numbers.
- **sources:** old/FUTURE_data_graph_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I4/#D1/D3, TODO_imagery_followups.md#Future-workstream
- **relations:** news imagery (I5) consumes it for data-driven stories; RUNBOOK B4 data-source keys.
- **verify-later:** no chart pipeline code expected; go-echarts dependency absence.

<!-- SOURCE: U10_imagery.md -->
### Product illustration pipeline (copyright-safe sketches)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Planned but not built (design work already done — reuse it)" (2026-07-08); I6 not started.
- **what:** Generate stylised product illustrations to avoid copyright/trade-dress exposure from scraped affiliate photos: discovery check `check_product_without_custom_illustration` (per-pass cap ~20), `product-illustration-handler` agent delegating to image-build-handler, `link_asset_to_product` action setting affiliate_products.custom_image_id, renderer precedence custom_image_id → cached_image_url. Stylisation is a hard-coded constraint, not a knob (D7): medium by product category (CAD-like / pencil / watercolour), altered viewpoint, in-context setting, no brand markings; img2img from the cached photo is v2-only under the derivative-work framing.
- **sources:** old/illustration/PLAN_product_illustration.md, PLAN_imagery_best_in_class.md#Phase-I6/#D7, STATUS_imagery_2026-05-12.md#Component-audit-finding
- **relations:** affiliate sites programme (resolver dependency); product components' query.affiliate_products socket; 3D reconstruction explicitly parked.
- **verify-later:** affiliate_products.custom_image_id usage; existence of the handler agent (expected absent).

<!-- SOURCE: U10_imagery.md -->
### Imagery best-in-class programme (G1–G9, D1–D8, phases I0–I8)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-12: "Phase I0 ✅ COMPLETE… Phase I1 ✅ COMPLETE, LIVE-VERIFIED… Phase I2 ⏳ IN PROGRESS… Phases I3–I8 not started."
- **what:** The 2026-07-08 successor programme raising fleet visual quality to best-in-class: nine goals (brand kit/logo permanence, data-accurate infographics, content-linked card imagery, graphic artefacts/sprites, copyright-safe product sketches, news imagery, performance budgets, accessibility/OG surface, quality loop) governed by eight user-confirmed design decisions (D1 code-rendered charts, D2 two lanes, D3 kind batches as text+CHECK, D4 brand guide as data, D5 logo lock, D6 dual icon strategy, D7 sketch constraints, D8 deploy-enforced budgets). Phases I0–I8, each acceptance-gated on robot-hands.com; companion running-notes/runbook/handoff/showcase document set maintained every turn.
- **sources:** PLAN_imagery_best_in_class.md, HANDOFF_imagery_best_in_class.md, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** builds on the loop-closure programme; RUNBOOK human-gate model; showcase docs quote its numbers.
- **verify-later:** phase status blocks vs live DB/site state; open runbook items B4/B5/B9/B10/B11.

<!-- SOURCE: U10_imagery.md -->
### Two lanes of imagery: plan-driven vs content-driven (Lane B)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** D2 "confirmed" 2026-07-08 as a decision; Lane B storage decision "generic entity_type + entity_id columns on assets — confirmed"; I3 not started.
- **what:** Everything built so far is plan-driven (fixed list decided at plan time). Card images, news charts, and product sketches are content-driven — attached to articles/news items/products, arriving continuously after the plan, prompts composed from the content itself plus the brand guide. Lane B generalises the affiliate custom_image_id pattern via entity_type+entity_id columns on assets, per-entity work item types, and content-sweeping discovery checks, sharing all generation/deploy/audit machinery downstream of the work item.
- **sources:** PLAN_imagery_best_in_class.md#3/#8, RUNNING_NOTES_imagery_best_in_class.md#Turn-2
- **relations:** content-linked card imagery (I3), news imagery (I5), product sketches (I6) are its instances.
- **verify-later:** assets table for entity_type/entity_id columns (expected absent yet).

<!-- SOURCE: U10_imagery.md -->
### imagery_style_guide — per-site brand guide as data (I1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "✅ PHASE I1 COMPLETE — LIVE-VERIFIED 2026-07-11… per-site imagery_style_guide driving generation with per-kind gating — PROVEN on real output."
- **what:** A site_specs aspect {palette, medium, mood, avoid, reference_asset_keys} distilled from design_intent, read by generate_image for every generation: photographic kinds get medium+mood+palette prepended, icons/sprite sheets palette only, logos nothing; the guide supersedes free-text imagery_direction when present; `avoid` terms feed the negative prompt (stronger channel than positive pleading); reference_asset_keys resolve to stable s3:// URIs (presigned URLs stripped back to bucket/key so anchors outlive the 7-day signature) and flow to Banana as style anchors. The single biggest lever for consistent professional look, per-site so sites diverge deliberately.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-17/#Turn-24, SQL_2026-07-10_robothands_imagery_style_guide.sql, SHOWCASE_technical_architecture.md#4
- **relations:** per-kind gating; reference-image anchoring; supersedes-at-runtime the Phase 0.1 free-text prepend.
- **verify-later:** imagery_style_guide.go; robot-hands site_specs aspect row; +style_guide log lines.

<!-- SOURCE: U10_imagery.md -->
### Logo permanence: generate → human-approve → lock (D5)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Logo: user-approved, LOCKED (assets.locked_at, lock_type=permanent); store-guard refuses overwrites" (2026-07-11, B6 done).
- **what:** One consistent logo for the life of a site is a policy, not a generation feature: the logo is generated, a human approves it via the runbook (A3 eyeball ritual), `locked_at` is set, and the assets upsert's `WHERE assets.locked_at IS NULL` guard refuses any future overwrite; auditors and regeneration paths must skip locked assets. Favicon and OG card are derived from the approved logo, never independently generated. robot-hands' May-8 logo was approved as-is and locked.
- **sources:** PLAN_imagery_best_in_class.md#D5/#Phase-I1, RUNBOOK_imagery_best_in_class.md#B6, RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** asset locking 2A supplies the columns; brand-head derivation consumes the locked logo.
- **verify-later:** robot-hands logo asset locked_at/lock_type; store guard in StoreAssetAction.

<!-- SOURCE: U10_imagery.md -->
### Brand-head derived assets (favicon + OG card)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "favicon.png/og-card.png serve 200; og:image + twitter:card injected into every head at render time" (I1 verification, 2026-07-11).
- **what:** `derive_brand_head_assets` action deterministically derives favicon (64×64 square resize) and OG card (1200×630, logo centred on a solid brand-palette colour; gradients rejected) from the locked logo bytes — no LLM — commits both to the site repo and records provenance rows (origin_model='derived-from-logo'). `injectBrandHeadTags` in render_site_components injects favicon/OG/Twitter head tags fleet-wide, idempotently. Runs via a `brand_head` mode branch on asset-deployer dispatched by a `needs_brand_head_assets` work item — the reusable pattern for any site (candidate auto-emit after logo lock).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-25/#Turn-26/#Turn-27, SQL_2026-07-11_asset_deployer_brand_head_mode.sql
- **relations:** logo permanence; sprite CSS head-link reuses the same injection + commit shape.
- **verify-later:** derive_brand_head_assets action registration; asset-deployer check_mode branch; live favicon/og-card on robot-hands.

<!-- SOURCE: U10_imagery.md -->
### Header logo resolution from plan imagery
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "header resolves the locked logo from plan imagery (logo-img live in the served header)" (2026-07-11); fix commit b00c150b.
- **what:** The header is a site component rendered by `render_site_components`, untouched by the page-level resolver fixes, and read the never-populated `sites.logo_url` — so sites showed a text mark despite a deployed logo file. Fix: `loadSiteDataFull` resolves the site-scope logo from site_plan_imagery→assets via `storage.DeployedWebPath` (never assets.url), keeping sites.logo_url as legacy fallback. Closed the long-standing "logo-in-header resolution gap" carried since 2026-05-27.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-23, FOCUS_imagery_assessment_1_.md#5.1 (gap origin), PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** image-role resolver (page-side sibling); DeployedWebPath convention.
- **verify-later:** loadSiteDataFull logo resolution; served header `<img>` on fleet sites.

<!-- SOURCE: U10_imagery.md -->
### Sprite-sheet bullets and list treatment (I2)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Phase I2 ⏳ IN PROGRESS — the active phase… I2.0 ✅… I2.1 ⏳ REGEN IN FLIGHT (Turn 31)… I2.2 NEXT BUILDABLE NOW" (2026-07-12/13).
- **what:** One coherent N×M glyph grid per site (3×3 @ 768², 256px cells; Banana — harnessing the model's gridded-image tendency), sliced by CSS `background-position`; bullets/nav via `::before` and `.sprite-<name>` classes — one generation, one asset, one stylesheet, no Go image cropping. Delivery deviation resolved twice: sprite CSS ships as a separate committed `/assets/css/sprites.css` + head `<link>` (css_snippets is a GLOBAL library with no site scoping; the per-site committed bundle is the house pattern). Cell-content alignment is THE risk, mitigated by ordered-grid prompt + human eyeball-and-assign gate (B11, cell_names_verified flag); vision auto-verify deferred to I2.4/I8. First generation was near-perfect (all 9 glyphs in reading order); its deploy failure spawned the ExtractActionInputs lesson.
- **sources:** SCOPE_I2_sprite_sheets.md, CONTEXT_PACK_imagery_sprite_sheet.md, PLAN_imagery_best_in_class.md#Phase-I2, RUNNING_NOTES_imagery_best_in_class.md#Turns-28–32, SQL_2026-07-12_seed_robothands_sprite_sheet.sql
- **relations:** five-place kind checklist (I2.0); brand-head commit pattern reused for sprites.css; referenced PLAN_imagery_sprite_sheet.md lives outside this unit.
- **verify-later:** chk_kind includes sprite_sheet; sprite-sheet-main.png 768×768 on robot-hands; sprites.css emit action existence.

<!-- SOURCE: U10_imagery.md -->
### Content-linked card imagery (I3)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Phases I3–I8: not started" (2026-07-12); card-crop decision confirmed 2026-07-08.
- **what:** Every linking card (blog index, news feed, tool directory) carries an image reflecting the content behind it, sharing a visual family with the content page. Confirmed approach: the card image is the article's asset re-cropped per purpose (one generation yields article hero, card crop ~800×450 WebP, OG crop), not a sibling generation. First real Lane B consumer; also clears the one remaining empty image slot on robot-hands (learning-center-index listing card).
- **sources:** PLAN_imagery_best_in_class.md#Phase-I3, RUNNING_NOTES_imagery_best_in_class.md#Turn-2/#Turn-13
- **relations:** two lanes (Lane B); news imagery reuses its mechanics; performance budgets set the card byte ceiling (≤60KB).
- **verify-later:** card kind/purpose in ImagePurposes (expected absent yet).

<!-- SOURCE: U10_imagery.md -->
### News imagery (I5)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** I5 not started; freshness rule confirmed 2026-07-08 ("no SLA… configurable news_image_grace_interval… working suggestion 6h").
- **what:** News ingestion attaches a per-item imagery decision via a small LLM classification: `chart` (data-driven story → I4 pipeline), illustration/photo (I3 pipeline), or none. Feed cards and article pages share the artefact. No SLA (ingest ~2×/day); after a configurable grace interval an item falls back to a brand-kit-derived default image so the feed never shows an empty slot.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I5, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** data-graph pipeline (I4), card imagery (I3), brand kit (I1); news→infographic backlog enhancement.
- **verify-later:** none yet (design only).

<!-- SOURCE: U10_imagery.md -->
### Image performance budgets (I7 / D8)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** budgets "confirmed as proposed" 2026-07-08 (hero ≤180KB, card ≤60KB, sprite ≤80KB, index above-fold ≤600KB); I7 not started.
- **what:** Per-kind byte and dimension budgets enforced at deploy (extend ImagePurposes with ceilings; WebP for photographic kinds) and policed by a new `image_weight_over_budget` discovery check routed to asset-deployer re-optimisation; responsive srcset + lazy loading in image-bearing templates; alt text required at generation time plus an `image_alt_text_missing` check; sprites amortise small art into one download.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I7/#D8, RUNBOOK_imagery_best_in_class.md#B5
- **relations:** OptimizeImageForWeb/ImagePurposes (existing enforcement point); accessibility goal G8.
- **verify-later:** ImagePurposes byte-ceiling fields (expected absent).

<!-- SOURCE: U10_imagery.md -->
### Image-role alias resolver + authoritative overlay
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "RESOLVED 2026-07-10 — the finding took THREE fixes, all deployed/applied… 16 distinct hero files, each referenced by exactly one page."
- **what:** There was no single contract for how a component gets its image — three incompatible patterns (legacy content_data.hero_url site-wide field; preset `site_assets.background/product_screenshot/...` sources nothing generates; components with no image slot) meant per-page heroes generated correctly but never rendered (same placeholder everywhere / empty src). Fix: `imageryplan.ImageRoleForPath` shared alias table normalising generic image field names to the "hero" role; page-aware `ensureAssets` resolving page hero → site hero fallback; `planSection` injecting the resolved hero under legacy alias keys into resolved_data, which the renderer merges LAST (the designed authoritative overlay), defeating the stale site-wide hero_url. Planner-side key alignment was rejected as structurally impossible (component selected after planning).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1, RUNNING_NOTES_imagery_best_in_class.md#Turns-5–10, PLAN_imagery_best_in_class.md#I0-FINDING
- **relations:** image_source_unsatisfiable check is its guarantee for future domains; DeployedWebPath was fix 2 of the trio; corrupted templates were fix 3.
- **verify-later:** imageryplan package + test; plan_sections_action.go resolve()/ensureAssets/planSection.

<!-- SOURCE: U10_imagery.md -->
### DeployedWebPath committed-path convention (the two-URL serving model)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Verified: zero X-Amz URLs appear in any deployed page's HTML" (2026-07-10 warning box, corrected after a wasted debugging turn).
- **what:** Every generated image has two URLs: `assets.url` is a 7-day presigned S3 URL (SigV4 hard protocol ceiling — a throwaway source handle, never used to render) and the durable committed git path `/assets/images/<asset-key>.<ext>` derived by `storage.DeployedWebPath(asset_key, purpose)` — the single source of truth shared by deployer and resolver so they cannot drift. Pages serve via GitHub Actions → Backblaze B2 → a Cloudflare worker that re-signs each GET server-side. Debugging rule: get the real asset_key/purpose from the DB and curl the derived path; a presigned URL in assets.url is cosmetic staleness, not a broken image.
- **sources:** PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS, RUNNING_NOTES_imagery_best_in_class.md#Turn-8, SHOWCASE_technical_architecture.md#3
- **relations:** storage-architecture (worker.js, buckets); image-role resolver emits these paths; leopardess AUDIT_verified_facts D8 is the full write-up.
- **verify-later:** storage.DeployedWebPath/AssetKeyFilename; deploy_image_asset_action.go url-flip (~line 250).

<!-- SOURCE: U10_imagery.md -->
### flag_page_image_rebuild re-render trigger
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Delivered 2026-05-27 as the edited plan_sections_action.go and new flag_page_image_rebuild_action.go."
- **what:** Because needs_imagery runs after the page first renders, the fallback bakes into rendered_html and terminal rerenders reassemble without re-resolving. A terminal step in image-build-handler flags the page needs_rebuild and emits `needs_page` at priority 99 for page-scoped imagery, so the page re-resolves *through* plan_sections after its asset lands — closing the asset→render timing loop (the general asset→rerender coupling for site-level components remains an open item).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1-Decision, SHOWCASE_technical_architecture.md#3, TODO_imagery_followups.md#12
- **relations:** image-role resolver (same fix bundle); rerender-reassembles-not-resolves root cause 3.
- **verify-later:** flag_page_image_rebuild_action.go; image-build-handler terminal step.

<!-- SOURCE: U10_imagery.md -->
### image_source_unsatisfiable discovery check
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Registered 2026-07-09/10; "live but has produced 0 flags (heroes all resolve now) — expected."
- **what:** Flags component input_schema image fields sourced from a `site_assets.<path>` that no asset key, plan imagery row, or image-role alias can supply — the systematic guarantee that the empty-src class of failure is caught on every future domain instead of eyeballed. Flag-only (needs_human_review, no handler), dedup per site/page/function/path, cap 25/pass. Substituted for the structurally-impossible planner-side guard (component chosen after planning). Shares the alias table with the resolver so the two cannot drift.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-7, SQL_2026-07-09_register_image_source_unsatisfiable.sql
- **relations:** image-role resolver; services-hero orphan case (generated hero no component consumes).
- **verify-later:** check_image_source_unsatisfiable.go; design-discovery-agent run_checks.

<!-- SOURCE: U10_imagery.md -->
### Reference-image style anchoring
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Item 24 decision 2026-05-18 ("plumb reference-image AND switch icon model"); Banana native path live via style guide reference_asset_keys (I1, 2026-07-11); IP-Adapter/img2img paths not built.
- **what:** Conditioning generations on a reference image for sibling consistency. Three techniques ranked (img2img subject-anchor; IP-Adapter style-anchor, not on Stability's standard REST endpoint; LoRA highest); three reference-provenance options (generate-one-then-derive; per-site curated style library; system-wide per-kind house style). What shipped is the Banana-native form: approved reference assets resolved to stable s3:// URIs flow as style anchors for photographic kinds. Schema hooks (reference_image_uri field, origin_asset_id, alterations JSONB) exist ahead of the fuller paths.
- **sources:** TODO_imagery_followups.md#24, PLAN_imagery_best_in_class.md#D4, RUNNING_NOTES_imagery_best_in_class.md#Turn-17
- **relations:** imagery_style_guide; adoption mirror as a reference source; product-sketch img2img v2.
- **verify-later:** Banana provider reference handling; whether any non-Banana reference path exists.

<!-- SOURCE: U10_imagery.md -->
### Per-vertical LoRA fine-tunes
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Status: planned, not started… the fine-tuning plan presupposes an adapter that can take a model field; our adapter cannot" (assessment); best-in-class explicitly keeps it "as a future consistency upgrade once I1's reference-image approach shows its limits".
- **what:** Training per-vertical image LoRAs (60–90 curated images, SDXL/PixArt base, ~£35–95 first pass, per 018_canine_biology) for consistent visual style (vet diagrams, energy infographics). Blocked historically on model-selection plumbing; now deliberately deprioritised behind the cheaper reference-image approach.
- **sources:** FOCUS_imagery_assessment_1_.md#8, PLAN_imagery_loop_closure.md#Open-items, PLAN_imagery_best_in_class.md#6
- **relations:** canine-biology (018); model-infrastructure training; reference-image anchoring as the substitute.
- **verify-later:** none (not started).

<!-- SOURCE: U10_imagery.md -->
### Prompt-composition composer/envelope revisit
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Step 5 — image prompt cascade: Defer… see FOCUS_prompt_composition_pattern.md for why copying the text pattern is the wrong target" (resolved 2026-05-13).
- **what:** The image path keeps a single-prepend cascade rather than matching page-content-writer's richer text composition — a considered asymmetry: the text pattern itself is judged fragile and not worth copying. The strongest future candidate is a composer step producing a parameter envelope for both text and images, likely landing in a 2H-sibling phase. Partially realised since by the style-guide gating (which is composition-by-kind, not a full cascade).
- **sources:** PLAN_imagery_loop_closure.md#Decisions/#Image-prompt-cascade—deferred/#Open-items
- **relations:** FOCUS_prompt_composition_pattern.md (outside unit); imagery_style_guide; 2H request shape.
- **verify-later:** whether any composer/envelope step exists.

<!-- SOURCE: U10_imagery.md -->
### Components declare imagery contracts / many-images-per-page direction
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 3.5 "NEW for the refresh" but Phase 3 deferred; "Future direction — pages with many images (post-Phase-3)".
- **what:** content_components.input_schema v2 already supports typed image fields with arbitrary `source` resolvers, but only hero_image uses it. The direction: components own their imagery contracts (team-grid declares member_avatars arrays; services-grid declares per-service icons), the renderer resolves scoped site_plan_imagery rows by asset_key, discovery walks the declared gaps, and generation scales horizontally (a 30-image page is 30 work items through the unchanged chain). Enables per-image audit and retires silent fallthroughs to /assets/images/hero.jpg. The features `icon` string contract being one-sided (no renderer wiring) is the standing counter-example, resolved separately via Lucide.
- **sources:** PLAN_imagery_loop_closure.md#3.5/#Future-direction, FOCUS_imagery_assessment_1_.md#4.2/#9-item-5
- **relations:** Lane B; card imagery; contracts-and-standards (input_schema slot specs).
- **verify-later:** any component beyond hero_image with resolved image declarations.

<!-- SOURCE: U10_imagery.md -->
### Human taste-gate operating model (runbook rituals)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNBOOK structure in active use: standing rituals A1–A5 + one-off queue B1–B11, most B-items closed with dates; "Humans only at the taste layer" (showcase).
- **what:** The imagery workstream's division of labour: agents do all authoring/migrations/deploy-prep; humans do credentials, backups approval, budget sign-off, and visual approval gates — logo approval (once, then locked), sprite-sheet cell verification (assign true meanings after generation), and sampled page eyeballs per phase acceptance. Gates are deliberately the phases' biggest cost; generation is never trusted to self-judge taste.
- **sources:** RUNBOOK_imagery_best_in_class.md, SHOWCASE_imagery_workstream.md#Why-it's-interesting, SCOPE_I2_sprite_sheets.md#Phasing
- **relations:** logo permanence; sprite eyeball gate B11; hitl category (broader HITL machinery).
- **verify-later:** n/a (process concept); runbook item states.

<!-- SOURCE: U10_imagery.md -->
### Imagery work-item economy end-to-end chain
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** SHOWCASE technical architecture (verified against production 2026-07-10): full diagram planner → site_plan_imagery → needs_imagery → image-build-handler → provider adapter → store_asset → asset-deployer → flag_page_image_rebuild → resolver → rendered page; "~90 s prompt → git commit".
- **what:** The consolidated, operating imagery pipeline as a single nameable chain, including its dedup property (partial unique index lets checks re-run forever) and honest-state property (mark_item_failed). This is the umbrella concept the individual phase concepts compose into, and the shape any new imagery capability (sprites, cards, sketches) must ride.
- **sources:** SHOWCASE_technical_architecture.md#2/#3, SHOWCASE_one_pager.md, PLAN_imagery_best_in_class.md#2
- **relations:** every imagery concept above; development-guide work-item lifecycle.
- **verify-later:** end-to-end trace on a fresh needs_imagery item.

<!-- SOURCE: U10_imagery.md -->
### Rerender reassembles, it does not re-resolve
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** FOCUS §5.1 root cause 3 (2026-05-27); Turn 10 "NEW, SEPARATE ISSUE… rerender-completeness gap" — later narrowed: the specific 2026-07-10 instance was corrupted templates, but the underlying property stands ("needs_rerender and the colour/CSS fixers regex-patch stored rendered_html; they do not re-run plan_sections").
- **what:** The terminal rerender path reassembles existing rendered_html rather than re-running section resolution, so values that later land in content_data (resolved images, alt text) do not reach the HTML without a page rebuild through plan_sections. flag_page_image_rebuild routes page-scoped imagery around this; site-level components (header/footer) and non-hero inline sections remain exposed. Also noted: page_components.rendered_html is a snapshot, not a view — template changes don't reach deployed pages without a rebuild.
- **sources:** FOCUS_imagery_assessment_1_.md#5.1, RUNNING_NOTES_imagery_best_in_class.md#Turn-10/#Turn-11, HANDOFF_robot_hands_rebuild.md#Also-watch
- **relations:** flag_page_image_rebuild; corrupted templates (the misdiagnosis neighbour); styling-render-pipeline.
- **verify-later:** which paths re-run plan_sections vs regex-patch rendered_html.

<!-- SOURCE: U10_imagery.md -->
### Orphaned generated assets (component consumes nothing)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "services-hero template has no `<img>`; the generated hero_services.jpg is never consumed" (2026-07-09); full orphan check "deferred"; services coverage question open since item 21 (2026-05-15).
- **what:** Generation waste where the plan requests imagery no selected component can display (services hero), or assets outlive replans (stale sprite-sheet-main.jpg clutter; superseded May assets). Detection partially covered by undeployed_assets and image_source_unsatisfiable; a dedicated orphan/asset-supersession check and repo cleanup pass remain open, as does the planner-side question of which pages get heroes at all.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-5/#Turn-31, TODO_imagery_followups.md#21, PLAN_imagery_best_in_class.md#I0-finding-(c)
- **relations:** image_source_unsatisfiable (the flag side); planner key stability (another orphan source).
- **verify-later:** assets rows with no referencing rendered_html; site-repo files with no assets row.

## Dropped / superseded ideas ledger (quick index)
- `check_legacy_image_prompts_aspect` — designed, then retired as scoped (see concept above).
- `directory-builder` agent and a dedicated `needs_section_data` handler agent — decided in project docs, never built, superseded by queryresolve + reconciler design.
- Per-section imagery "deferred" (loop-closure v2 decision) — reversed 2026-05-12 by Phase 2G.
- Diffusion transparency for icons — abandoned (flat-grey chip).
- SDXL for flat icons — abandoned (Banana routing).
- Fix-A prompt-only icon rescue — tried, judged insufficient 2026-05-18.
- Planner-side guard "don't plan heroes for hero-less components" — structurally impossible (component chosen post-plan); substituted by image_source_unsatisfiable.
- Hand-rolled spawn+call trigger workflows — anti-pattern, replaced by work-item dispatch routing.
- `system.intake` trigger topic — stale/dead; kcat orchestrate envelope is the working path.
- pageflow-builder — retired from new-feature work, snapshot preserved.
- 3D product reconstruction — parked (sketches cover the need).
- fork_theme_from_site install mode — removed 2026-04-19; FK-swap is the sanctioned workaround.

<!-- SOURCE: U18_sql_for_agents.md -->
### image-generator + image prompt plumbing
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Contract in 029; Phase 0.1 (107) wires site_id through six call sites "so it can read design_intent.imagery_direction from site_specs".
- **what:** AI image generation specialist taking prompt/image_prompts (logo, hero_home) and producing image_url/image_data. Phase 0.1 made it site-aware: callers pass site_id so the generator composes design_intent.imagery_direction into prompts.
- **sources:** 029_image_generator.sql; 107_image_build_handler.sql (Section 1)
- **relations:** image-build-handler, site-work-orchestrator, pageflow-builder call it; asset-deployer deploys results
- **verify-later:** image generation adapter/action; imagery_direction composition code

<!-- SOURCE: U18_sql_for_agents.md -->
### image-build-handler + needs_imagery kind branch (Phase 2G)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 057 defines the handler; 107's "phase_2g_step5" section adds check_item_type_imagery / spawn_image_gen_imagery / call_imagery_gen / check_imagery_brand_update / store_imagery_brand_asset|store_imagery_asset steps ("teach image-build-handler to process needs_imagery work items (emitted by step 4's check_unfulfilled_imagery_plan)").
- **what:** Self-contained dispatch-loop handler for image work items: originally needs_logo/needs_hero_image (branch on spec.purpose → call image-generator → store_asset → deploy_image_asset via S3/optimize/git). Phase 2G extends it to generic needs_imagery items carrying kind-specific behaviour (icon transparency, logo variants), routed by item_type, with a spec.brand_update boolean deciding whether the stored asset also updates site brand assets.
- **sources:** 057_image_build_handler.sql; 107_image_build_handler.sql
- **relations:** build-dispatch-loop (caller); check_unfulfilled_imagery_plan discovery (imagery plan reconciliation); asset-deployer
- **verify-later:** live image-build-handler workflow steps; needs_imagery item emission in discovery checks

<!-- SOURCE: U18_sql_for_agents.md -->
### Imagery provenance: origin_model + origin_prompt on assets
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 107 combined migration Sections 2–3 (2026-05-05): "set origin_model literal on store_asset steps so the assets table records what produced each image"; origin_prompt_field normalised to record "the actual composed prompt sent to the model... not the un-composed plan prompt".
- **what:** Asset provenance discipline: every stored image records the generating model and the exact post-composition prompt. Required coordinated Go+SQL shipping (three concerns in one transaction) across image-build-handler, site-work-orchestrator, pageflow-builder.
- **sources:** 107_image_build_handler.sql (Sections 2–3 + backup preamble)
- **relations:** imagery audit work (file says provenance is "better for future iterations of the imagery audit work")
- **verify-later:** assets table columns origin_model/origin_prompt population

<!-- SOURCE: U19_sql_tables_components.md -->
### Asset key multi-image identity (Phase 2B–2D)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Staged migrations with live psql output pasted (11 rows backfilled; backup table assets_backup_20260508_pre_phase2d; ON_ERROR_STOP guard; old (site_id,purpose) unique index dropped).
- **what:** Replaces one-image-per-purpose with per-row asset_key: unique on (site_id, asset_key) WHERE active, enabling multiple images per logical purpose (e.g. adoption-mirror imports as 'adopted:<filename>'). Four-phase rollout: 2B add+backfill (asset_key=purpose), 2C StoreAssetAction writes asset_key and switches ON CONFLICT, 2D drops old purpose uniqueness after straggler sanity check.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2B and #Phase2D
- **relations:** assets provenance; site_plan_imagery key → namespaced asset_key.
- **verify-later:** idx_assets_site_asset_key_unique; StoreAssetAction ON CONFLICT target.

<!-- SOURCE: U19_sql_tables_components.md -->
### Assets table with full provenance
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 004 PART 6 creates assets with origin tracking; later phases (locks, asset_key) applied to a live table with 11 rows.
- **what:** All binary assets (image/video/document/logo/favicon) with storage location (provider/path/url), file metadata, and provenance: origin_type (generated/uploaded/scraped/stock/affiliate/derived), origin_url/prompt/model, origin_asset_id for derivations, alterations history JSONB, attribution/license. Purpose field ('hero', 'og_image'...) drives placement.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART6; docs/agent_docs/sql_for_tables/041_assets.sql
- **relations:** asset_key identity; image-build-handler work items; storage-architecture (providers).
- **verify-later:** StoreAssetAction; storage_provider values in use.

<!-- SOURCE: U19_sql_tables_components.md -->
### site_plan_imagery structured imagery plan
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "PURE DDL with NO BEHAVIOUR CHANGE. The table is empty until step 2 (write_site_plan extension) and step 3 (planner prompt extension)" with the 5-step Phase 2G sequencing listed (043 header).
- **what:** Sibling of site_plan_directives holding structured imagery requirements at site/page/section scope: key (asset_key stem, namespaced by the discovery check), kind enum via chk_kind (logo, hero, illustration, icon, infographic — product deliberately excluded, it comes from the affiliate_products resolver), required prompt, style_hints/constraints JSONB that cascade ADDITIVELY with directives' imagery_direction, ordering, source enum, and HITL locking with lock-transfer across plan rebuilds. chk_scope_ref_consistency enforces NULL / page_name / 'page:ordering' shapes; unique on (plan, scope, COALESCE(scope_ref,''), key).
- **sources:** docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql
- **relations:** site_plans domain; PLAN_imagery_phase_2g.md; check_unfulfilled_imagery_plan (step 4); image-build-handler (step 5).
- **verify-later:** table population; steps 2–5 delivery status.

<!-- SOURCE: U20_legacy_docs_a.md -->
### generate_image action + image-generator adapter pipeline
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** docs002/0100: "Image creation is now working" (deployed then); architecture: agent → `system.adapter.image-generator.requests` → Stability AI → Backblaze/S3 → reply_to_topic. Taxonomy names site_plan_imagery as the current pipeline.
- **what:** Image generation as a first-class workflow action: GenerateImageAction resolves prompts (template-rendered from CollectedData), sends to a shared adapter topic consumed by 3 load-balanced Python adapter replicas (consumer group), which call Stability AI, upload PNG to S3/Backblaze under `images/{client_id}/{date}/{image_id}`, and respond to the requesting agent's topic. Circuit breaker for API failures. A notable bug/fix: GenerateImageAction originally bypassed the image-generator *agent* and posted straight to the adapter — corrected so the agent orchestrates (parent → agent → adapter → agent → parent).
- **sources:** docs001_flow_general/README.095.a.image_handling.git.057_image.md; docs001_flow_general/README.095c.image_handling_topics.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** successor: docs024 imagery / site_plan_imagery pipeline; adapter microservice pattern; GPU strategy.
- **verify-later:** internal/adapters/imagegenerator; whether Stability AI config survives; current imagery pipeline tables.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Image storage and display URL strategy
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** README.099a recommends and implements dual URIs: "image_uri (s3:// for storage reference), image_url (https:// for web use)"; robot-hands pages embedded presigned URLs.
- **what:** Generated images return both an s3:// URI (storage reference) and a public HTTPS/CDN URL for embedding in HTML; options canvassed were public-bucket/CDN (chosen), presigned URLs (expiry problem for permanent sites), base64 embedding, and an image proxy service. Backblaze B2 public bucket setup documented.
- **sources:** docs001_flow_general/README.099a.image_storage_and_display_urls.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** storage-architecture (S3/B2 credentials); imagery pipeline.
- **verify-later:** ConvertS3URIToPublicURL or equivalent; current image URL scheme on live sites.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Asset & product provenance tables
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** docs012/010 new-table list: "assets — track all images/videos with full provenance; products; product_assets; affiliate_programs; affiliate_products"; docs017/002_pageflow stores hero assets in "assets table (existing)".
- **what:** All media (generated, uploaded, scraped) tracked with provenance in an assets table; product catalog and affiliate product caching designed alongside for e-commerce/review sites. The assets side shipped (used by image generation flow); products/affiliate tables remained design.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#New-Tables; docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md
- **relations:** image generation pipeline; entity-data (products as entities superseded product tables); link_registry affiliate fields.
- **verify-later:** assets table columns; products/affiliate_programs existence.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Image generation in the build pipeline
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** docs017/001_changes + 002_pageflow give exact patches ("generate_hero_image → store_hero_asset → deploy_hero_image (NEW) → templates use {{.hero_url}} → /assets/images/hero.jpg"); the current imagery system (site_plan_imagery, kind enums) replaced this.
- **what:** First-generation site imagery: image-generator agent produces logo/hero via adapter; store_generated_image/StoreAssetAction persists S3 URI into assets table and sites.content_data by purpose; deploy_image_asset downloads from S3, optimizes for web (resize per purpose), base64-commits via git-adapter; hero/logo URLs flow through render context into templates as background images.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md; docs017_legacy_agent_rules_images_design_keydocs/001_changes_needed.md
- **relations:** assets table; imagery pipeline (successor); image-optimiser fix agent.
- **verify-later:** deploy_image_asset action; storage/image_processing.go; current imagery pipeline contrast.

<!-- SOURCE: U22_recent_small_docs.md -->
### Image LoRA — scientific illustration style
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase F todo "Image LoRA fine-tuning (week 7-8)" unchecked; "SDXL recommended for diagrams" over FLUX.
- **what:** A plan to train an image LoRA (SDXL/PixArt preferred over FLUX for clean diagrams) on 60-90 curated, captioned veterinary/biological illustrations so the image-generator produces consistent anatomical cross-sections, pathway diagrams, procedure illustrations, and infographics across a site. Served via serverless (Replicate/RunPod) rather than an in-cluster GPU initially.
- **sources:** docs023.../018_canine_biology.md#7
- **relations:** image-generator adapter, canine biology KB, self-hosted inference
- **verify-later:** image-generator adapter LoRA support; any vet-diagram LoRA weights

<!-- SOURCE: U23_docs_root_vonc.md -->
### Resolver asset-kind surfacing gap (hero/logo only; illustrations unreachable)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Confirmed 2026-07-02: illustration assets EXIST (illustration_game_master, illustration_gauntlet_cta, purpose=illustration, active, files deployed) but "resolver ensureAssets only surfaces hero/logo, so site_assets.illustration can't reach them"; workaround applied (field made optional); extension still backlog item 4 on 2026-07-09.
- **what:** The plan_sections resolver's ensureAssets populates only `hero` and `logo` asset keys (from site_plan_imagery kinds hero/logo), so any schema field sourced `site_assets.illustration` can never resolve even when illustration assets exist — deferring sections. Interim: make such fields optional (text-only render). Structural options: extend ensureAssets to surface kind=illustration from site_plan_imagery+assets (benefits all sites), or per-field fallback URLs. Related mismatch: gauntlet-cta has no illustration field despite an illustration_gauntlet_cta asset existing.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:50 + #2026-07-03-~13:18; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; site_plan_imagery (imagery pipeline)
- **verify-later:** plan_sections resolver ensureAssets; site_plan_imagery kinds in use

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Data-graph / charts future pipeline
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** TODO_imagery_followups(15) "Future workstream: data graphs / charts (added 2026-05-20)"; HANDOFF news-enhancement (b) ties news→infographic to "the future data-graph pipeline (FUTURE_data_graph_pipeline.md) if the infographic is data/chart-driven"; PLAN_imagery_phase_2g: `kind=infographic` "renders as illustration for now."
- **what:** A planned separate generator producing SVG/HTML charts from structured data (distinct from raster image generation), the eventual real backing for `kind=infographic` and the news→infographic enhancement. No scaffolding exists; closest analogues are the dynamic-tool components.
- **sources:** imagery/old/TODO_imagery_followups(15).md#future-workstream-data-graphs; imagery/old/PLAN_imagery_phase_2g(1).md#what-2g-doesnt-include
- **relations:** infographic kind; news→infographic; many-images-per-page
- **verify-later:** FUTURE_data_graph_pipeline.md (live); infographic-generator agent (absent)

<!-- SOURCE: U25_leopardess_social.md -->
### Two-URL asset model (throwaway presigned handle vs durable git path)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** AUDIT D8 correction (2026-07-10): "zero X-Amz presigned URLs appear in … rendered pages"; "the imagery pipeline works. There is no platform-wide emergency."
- **what:** assets.url stores a 7-day presigned S3 URL that is only a source handle — SigV4 caps expiry at 604800s so a permanent presigned URL is impossible. Render never reads it: plan_sections emits storage.DeployedWebPath (the git-committed /assets/images/<key>.<ext> path), served via GitHub Actions → Backblaze B2 → a Cloudflare Worker that re-signs B2 GETs server-side. 83 stale presigned rows are cosmetic; the w9_04 url-flip backfill (done for idea.uk) is the optional cleanup. Includes the D8 self-correction on record: an earlier "asset deploy is broken platform-wide" alarm was withdrawn after re-investigation.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#D8; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; docs/leopardessconsulting/RUNBOOK.md#landmine-6/7
- **relations:** spawn-time storage env injection; storage-architecture (B2, Cloudflare worker)
- **verify-later:** plan_sections_action.go:193/260/290; scripts/cloudflare/worker.js; assets.url vs storage_path across sites

<!-- SOURCE: U25_leopardess_social.md -->
### Spawn-time storage env injection (base chassis carries no IMAGE_BUCKET by design)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** AUDIT D8 finding 1: "spawn_actions.go injects IMAGE_BUCKET + storage creds into spawned storage-enabled agents … 107_image_build_handler.sql:725 documents exactly this."
- **what:** The base agent-chassis pod deliberately lacks IMAGE_BUCKET; deploy_image_asset therefore fails when run inline, but the real pipeline spawns asset-deployer with storage env injected (isStorageEnabledAgent). Hand-triggering a storage agent standalone reproduces a spurious "Storage client not configured" failure. Documented as a "do not add IMAGE_BUCKET to agent-chassis" warning.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#D8; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; docs/leopardessconsulting/scripts/commit_brand_assets.sh (header records the earlier, since-corrected reading)
- **relations:** two-URL asset model; system-architecture (agent spawning)
- **verify-later:** platform/orchestration/actions/spawn_actions.go isStorageEnabledAgent; agentbase/agent.go:294

<!-- SOURCE: U25_leopardess_social.md -->
### Imagery kind→provider routing and reference-image support (A6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** HANDOFF §2 A6: "Committed in dynamic_adapter.go; NOT deployed to cluster (needs a Makefile build-from-local-filesystem)".
- **what:** dynamic_adapter routes kind=="icon" to Banana (Gemini gemini-3-pro-image-preview) and everything else to Stability SDXL. Only Banana honours reference images — the mechanism brand consistency depends on — so A6 extends the Banana route to logo/illustration/infographic, leaving hero/photographic on SDXL. Imagery kinds are constrained twice (chk_kind and validImageryKinds in write_site_plan_action.go): changing the set needs migration + Go edit together.
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4, #Turn-7; docs/leopardessconsulting/RUNBOOK.md#O5, #landmine-5; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L2
- **relations:** logo candidate generation; brand-asset derivation gap; chart concept (charts deliberately NOT an imagery kind)
- **verify-later:** internal/adapters/imagegenerator/dynamic_adapter.go routing on the running pod vs repo

<!-- SOURCE: U25_leopardess_social.md -->
### Brand-asset derivation gap and direct git-adapter brand commit path
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** REPLICATION summary table: favicon + OG-card derivation "[gap] — small; currently manual"; RUNBOOK H4 "resolved 2026-07-10 … deployed live … all verified 200 and byte-identical".
- **what:** Favicon and OG-card generation from a logo is not a pipeline capability (no favicon/OG generator in the codebase); they were manually derived (background knockout to transparency, gold normalisation to brand hex, flood-filled silhouette favicon for 16/32px, multi-size .ico, opaque apple-touch icon, 1200×630 OG card). Delivery used commit_brand_assets.sh: send the same git-adapter commit message deploy_image_asset would send, for pre-approved images where there is no generation step to spawn. deploy_brand_asset.sh notes asset-deployer never rewrites assets.url without asset_id (landmine 6).
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-9; docs/leopardessconsulting/RUNBOOK.md#O7; docs/leopardessconsulting/scripts/commit_brand_assets.sh, deploy_brand_asset.sh (headers)
- **relations:** imagery kind routing (A6); two-URL asset model
- **verify-later:** internal/adapters/git/github_client.go CommitToRepo path prefixing; absence of favicon/OG generator

<!-- SOURCE: U25_leopardess_social.md -->
### Logo candidate generation with small-size survival test and human approval
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 8–9: six candidates generated 2026-07-10, owner chose c2, exact approved PNG deployed live and byte-verified.
- **what:** Brand-mark commissioning practice: generate N candidates through the same model/key the chassis uses, save all prompts for pipeline reproduction, judge by small-size survival (a favicon is 16px — solid-fill marks survive, line-art dissolves), and require human approval for a for-the-life-of-the-site decision (maps to the platform's checkpoint_for_review HITL surface). Key insight: the owner approves a specific image, not a prompt — regenerating "the same" prompt yields a mark they never saw.
- **sources:** docs/leopardessconsulting/logo_candidates/PROMPTS.md; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8, #Turn-9; docs/leopardessconsulting/REPLICATION_in_chassis.md#4
- **relations:** hitl (checkpoint_for_review); imagery kind routing (A6)
- **verify-later:** site_plan_imagery logo row for leopardess; checkpoint_for_review surface

<!-- SOURCE: U25_leopardess_social.md -->
### Illustration/section-imagery asset resolution (kind-alias path)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-11: "the section's render simply predated the section-imagery resolver … schema source site_assets.illustration resolves via site_plan_imagery kind-alias … src filled"; residual: undeployed_assets "infers deployment from rendered_html usage, not from the repo".
- **what:** How section images resolve: a schema field sourced site_assets.<type> resolves through site_plan_imagery kind-alias rows to a deployed asset path; stale renders predating the resolver ship `<img src="">` and are fixed by a light page_rerender (reason image_landed, no LLM). Known gaps: ensureAssets surfaces only hero/logo (named illustrations exist as assets but weren't reachable until aliased); the undeployed_assets check flags committed-but-unreferenced assets (usage-inferred, not repo-inferred). The archetype hub consumed 8 orphaned icons via page-scope hero alias rows.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11, #2026-07-12; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.5a; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (outstanding illustration)
- **relations:** two-URL asset model; plan_sections deferral; archetype hub build
- **verify-later:** ensureAssets; site_plan_imagery kind-alias resolution code; undeployed_assets check logic

---

## Notes for consolidation

- **Family deltas of substance:** the 003 family shows RAG integration and the automated strategic-review agent designed then consciously deferred (extracted above); the 002 spark family shows continuous accretion (Arena/Stage split arrives in 002b, motivation tiers 002c, user journey + cost tables 002d, launch strategy 002e) with only minor drops ("Retention Bridges"/"Crews"/"Workshop Mode" headings folded into later sections, not abandoned). Tool-docs NOTES families are append-only logs; earlier versions contain no concepts absent from the latest.
- **Cross-unit duplicates expected:** verify-by-artifact discipline, the apply_section_edit/approved defect, imagery two-URL model, and 016b debugging entries will also surface from docs024/docs019 units — merge on consolidation; this unit contributes dated deployment evidence.
- **Proposed NEW categories:** `NEW:data-charts` (Go SVG emitter + JS enhancement, D1/D3 doctrine — could back a council agent distinct from diffusion imagery), `NEW:operator-practice` (verify-by-artifact, backups, kcat triggers, in-chassis replicability — cross-cutting discipline no seed category owns), `NEW:rag-knowledge-base` (rag_index/rag_lookup/knowledge_base — referenced as existing infrastructure with no seed home).

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Imagery: per-page assets vs site-wide last-write-wins resolution gap
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** 016 §9 deployed-hero-images entry: assets deploy fine; resolution gap diagnosed with fix direction (page-aware ensureAssets + re-render through plan_sections)
- **what:** site_plan_imagery plans per-page keys; store_asset writes content_data[<purpose>_url] keyed by purpose so every page hero overwrites the last (single site-wide hero_url); first render bakes the use_fallback static path; terminal rerender/CSS fixers patch stored HTML without re-resolving. Fix: resolve per-page from site_plan_imagery⋈assets, keep content_data as gap-fill; after an asset lands flag needs_rebuild → needs_page at p99. Logo is chrome (render_site_components) — separate path. imagery kind/scope model + chk_kind constraint implied by site_plan_imagery columns.
- **sources:** 016 §9 hero/logo fallback entry; 002(4) flag_page_image_rebuild → page_rerender
- **relations:** page_rerender image_landed reason; input schema image fields rule
- **verify-later:** ensureAssets page-aware now?; site_plan_imagery schema

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Imagery loop closure plan (Phases 0–6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Decisions table (new imagery-quality-auditor agent; max 2 regen attempts; asset locking mirrors page_components; per-section granularity deferred); later docs show Phase 2G/2H verified 2026-05-15 and asset lock columns landed via migration 053
- **what:** The sequenced plan closing the gap between imagery asked for (specs/plans) and imagery delivered: Phase 0 wire imagery_direction into prompts + populate origin_model; Phase 1 algorithmic discovery checks (unfulfilled_image_prompt, placeholder_image_in_use, image_url_404) routed to the existing image-build-handler; Phase 2 assets locking + asset_key; Phase 3 adoption image mirror; Phase 4 text-only visual-auditor imagery awareness; Phase 5 vision-capable LLM path; Phase 6 imagery-quality-auditor. Explicit non-goals: per-section imagery_plan, icon resolver, infographic generator, provider router, img2img.
- **sources:** PLAN_imagery_loop_closure.md (whole); ASSESSMENT_imagery_phase_0_1…md
- **relations:** dispatch diagnostic (Phase 2G verification); adoption faithfulness locks; site_plan_imagery
- **verify-later:** which discovery checks exist under discovery_checks/; assets.locked_at/lock_type/asset_key columns

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### imagery-quality-auditor agent
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 6 of the plan; not reported deployed anywhere in scope (plan pre-dates 2026-05)
- **what:** A vision-capable sibling of visual-design-auditor: loads assets + imagery_direction (lock-honouring, excluding human uploads), runs a vision LLM audit with imagery-specific categories (direction_mismatch, brand_mismatch, inconsistency, quality, inappropriate), writes findings with max_fix_attempts 2 routing back to image-build-handler; counts toward the existing 3-pass audit cap; gated rollout via design-audit-agent.
- **sources:** PLAN_imagery_loop_closure.md#Phase-6
- **relations:** vision-capable LLM path; asset locking; design-audit-agent
- **verify-later:** agent_definitions for imagery-quality-auditor

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Asset locking mirrors page_components (+ asset_key multi-image readiness)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** planned in Phase 2; FOCUS_adoption_faithfulness (2026-05-19) states 053 migration adds lock_type + lock_expires_at "on page_components, site_components, site_plan_directives, assets … written, ready to apply"
- **what:** assets gains locked_at/lock_type (same vocabulary and exclusion predicate as page_components) so audits/discovery skip locked assets; asset_key column (default = purpose, unique per site) opens multi-image purposes (adoption mirror writes adopted:<filename>) without breaking existing single-purpose upserts; old purpose-unique index dropped only after asset_key bedding-in.
- **sources:** PLAN_imagery_loop_closure.md#Phase-2; FOCUS_adoption_faithfulness_via_locks(2).md#implementation-plan
- **relations:** lock policy table; adoption image mirror; per-page hero resolver (assets.asset_key join)
- **verify-later:** assets table columns and indexes

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Imagery algorithmic discovery checks
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Phase 1 planned three checks; dispatch doc shows `unfulfilled_imagery_plan` check (Phase 2G.4) live and emitting 8 items on robot-hands (2026-05-14) — with the pipeline='design' emission bug
- **what:** No-LLM checks catching spec-to-delivery gaps: image prompt in plan but no asset; hardcoded fallback path in rendered_html with no asset (the silent-failure case); referenced image URL with no matching asset. Emit needs_imagery/needs_hero_image/needs_logo items to image-build-handler via the dispatch loop.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1; FOCUS_dispatch_diagnostic(4).md#why-stuck, #Q4
- **relations:** pipeline soft label bug; baked-fallback problem
- **verify-later:** discovery_checks/check_unfulfilled_image_prompt.go etc.; unfulfilled_imagery_plan pipeline value

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### imagery_direction into image prompts + origin_model provenance (Phase 0/0.1)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "The Phase 0.1 deliverables stand" pending one verification query (assessment, undated ~2026-05); compatible with Phase 1 architecture which stabilises design_intent ownership
- **what:** Image generation reads site_specs design_intent.imagery_direction and prepends "Style direction: <direction>\n\nSubject: <prompt>" to the three-tier prompt; store_asset writes origin_model for provenance. The strategic-only read survives Phase 1 (per-page directives become the successor once site_plan_directives lands). Side benefit: pulls planner-invented hero prompts back toward the adopted look (partial mitigation of Bug 4).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0; ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md (whole)
- **relations:** planner ignores site_archetype imagery; image parameter-shaping
- **verify-later:** generate_image_actions.go composeImagePromptWithDirection; assets.origin_model population

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption image mirror
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 3 of the plan; "Adopted images are persisted but only as historical record for now" — no deployment claim in scope
- **what:** Stop discarding crawled imagery: mirror_adoption_images action downloads source images (capped count/size), uploads to S3, inserts assets rows (origin_type=adopted, asset_key=adopted:<filename>); wired into apply_adoption_plan plus a backfill discovery check and a one-step adoption-image-mirror agent. Future hook for img2img reference generation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3
- **relations:** asset_key; image parameter shaping (reference_image_uri)
- **verify-later:** mirror_adoption_images_action.go existence; assets rows with origin_type='adopted'

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Vision-capable LLM path
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 5 of the plan; no deployment claim in scope
- **what:** Extend aiservice with GenerateTextWithImages (Anthropic image content blocks, URL source), preferring extension of execute_llm_prompt with an image_urls_field config over a new action; presigned-URL freshness required; vision_call tagged in llm_call_log for cost tracking.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** imagery-quality-auditor (consumer); llm_call_log
- **verify-later:** platform/aiservice/anthropic.go vision support

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Image generation as parameter shaping (not prompt blending)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "The composer step is its own work" — recommendation recorded during Phase 2G step 5 design (undated, ~2026-05)
- **what:** Unlike text, images have a 77-token CLIP budget and no "don't" understanding — composition means deriving parameters (subject, negative_prompt from kind, style_preset/LoRA from imagery_direction, reference_image_uri from adopted images, aspect/cfg/steps per kind), not blending prose. A cheap compose_image_request step (Go rules or small LLM) producing a parameter envelope before image-generator is the candidate design; belongs with Phase 2H request-shape work.
- **sources:** FOCUS_prompt_composition_pattern.md#What-this-means-for-images
- **relations:** mega-prompt fragility (envelope pattern B); Phase 2H
- **verify-later:** image-generator request shape; any compose_image_request action

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site_plan_imagery sibling table
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** live JOIN in production code by 2026-06-02 ("site_plan_imagery.key = assets.asset_key"); write_site_plan step description updated to include imagery HITL-lock transfer (2026-05-26)
- **what:** Structured per-image plan rows (kind, key/asset_key, prompt, style hints, scope/scope_ref) mirroring site_plan_directives' scope+locking pattern — the successor to the legacy site_specs.site_plan.image_prompts dictionary; scoped page rows drive per-page heroes.
- **sources:** FOCUS_site_spec_vs_site_plan.md#where-imagery-lives; HANDOFF_2026-06-02…md#fix
- **relations:** per-page hero resolver; lock transfer; directive cascade
- **verify-later:** site_plan_imagery schema; emit_imagery_items writes

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Per-page hero resolver + rebuild-after-asset (baked-fallback fix)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "the fixes are in production" (2026-06-02) — plan_sections page-aware ensureAssets; flag_page_image_rebuild wired as image-build-handler terminal step (UPDATE 1 verified); registry entries "still required" at doc date
- **what:** Root cause triad: site-wide hero_url overwritten last-write-wins; async imagery completing after first render baked the on_missing fallback (/assets/images/hero.jpg) into rendered_html; terminal rerender reassembled stored HTML without re-planning. Fix: ensureAssets resolves this page's hero via site_plan_imagery JOIN assets (page scope) and site logo from site scope; flag_page_image_rebuild flags the page needs_rebuild and emits needs_page at priority 99 (dedup key page_rerender:<page>) so it re-resolves through plan_sections after its asset lands. Logo/header path deliberately out of scope (render_site_components, not plan_sections).
- **sources:** HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md#1
- **relations:** imagery checks; section-data reconciler (same handoff); two image-resolution paths open follow-up
- **verify-later:** registry.go entries for flag_page_image_rebuild/reconcile_section_data; hero component input_schema (field vs hardcoded template — open question)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Planner ignores site_archetype imagery constraints (Bug 4)
- **category:** imagery
- **status-signal:** unknown
- **status-evidence:** "site_archetype.design.imagery … says 'minimal icons/diagrams, no decorative photography'. The planner's site_plan still produced lavish hero prompts" (2026-04-23)
- **what:** The planner invents hero image prompts contradicting the adopted archetype's imagery stance. Fix shape: planner prompt reads site_archetype.design.imagery and sets needs_images=false when it says none/minimal. Phase 0.1's style-direction prepend partially mitigates the symptom.
- **sources:** HANDOFF_2026-04-23(1).md Bug 4; ASSESSMENT_imagery_phase_0_1…md#Bug-4
- **relations:** imagery_direction prompt composition; adoption faithfulness
- **verify-later:** plan_site prompt for archetype imagery constraint

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Section-scope imagery pipeline (plan → emit → generate → deploy → rebuild)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNBOOK §W7: "The flow already exists (05-26 handoff + code): `write_site_plan` → `site_plan_imagery` rows … → `emit_imagery_items` … → `needs_imagery` (priority ≤98) → `image-build-handler` … → `flag_page_image_rebuild` → `needs_page` (99) → rebuild resolves the URL"; W7b exercised it end-to-end (assets active at B2 in ~3 minutes each, 2026-07-03).
- **what:** The dynamic imagery supply chain: the planner writes `site_plan_imagery` rows (scope site/page/section; scope_ref `page:ordinal` for sections; key; kind hero/logo/icon/illustration/infographic; authored prompt — the table has NO description column; ordering; source), the gap-driven `emit_imagery_items` emits `needs_imagery` items only where no asset exists, image-build-handler's 25-step workflow generates, stores, brand-checks, spawns the asset-deployer (download S3 → optimise by purpose → commit to the site git repo, key-named files `_`→`-`.jpg), then `flag_page_image_rebuild` emits `needs_page` so plan_sections re-resolves the now-present asset. For idea.uk it never fired for the brief-explanation illustration simply because the planner never requested one (16 rows: 5 heroes, 10 icons, logo). Ordinal-based scope_refs drift when plans reorder (hygiene note; resolution is by key).
- **sources:** RUNBOOK_scheme_to_components(50).md#W7 #W7-0.3/0.4-RESULTS; w7b_01_imagery.sql; running_notes_scheme_to_components(55).md#Th #Tj #Ty #Tz
- **relations:** ensureAssets resolution gap; flag_page_image_rebuild section scope; presigned-URL expiry.
- **verify-later:** site_plan_imagery schema + emit_imagery_items step in build-site-planner; image-build-handler workflow steps.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### ensureAssets section-scope resolution gap (Edit B)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Tq): "Edit B LIVE (tools t)"; (Tv): "BRIEF-EXPLANATION CLOSED … illustration renders on index AND tools; Edit B fine throughout."
- **what:** The structural gap that made section illustrations unresolvable: `ensureAssets` (plan_sections_action.go) loaded only the page hero and site logo into the resolver's assets map (plus a legacy content_data fallback), so `site_assets.<key>` for section-scope imagery could never resolve — the pipeline's "last inch" was never wired. Edit B adds a third query (spi scope='section', scope_ref LIKE page||':%', joined to active assets), mapping BOTH by key (per-key schema paths like icon sets) AND by kind first-wins alias (generic `site_assets.illustration` paths), modelled on the hero block. The two-day "index miss" after deployment turned out to be a probe artefact (grep for the asset key string, but objects are UUID-named), taught as a debugging lesson.
- **sources:** gobatch_01_plan_sections.md#Edit-B; RUNBOOK_scheme_to_components(50).md#W7-CODE-FINDING; running_notes_scheme_to_components(55).md#Ti #Tt #Tu #Tv
- **relations:** section-scope imagery pipeline; plan_sections field deferral; probe-blindness (SQL pitfalls).
- **verify-later:** plan_sections_action.go ensureAssets section query; rendered brief-explanation `<img>` src on index/tools.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### flag_page_image_rebuild section-scope mapping (Edit H)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** RUNBOOK 07-06 night: "Slice 4c: PENDING APPLY — `gobatch_05_flag_section_scope.md` now carries Edit H … AND Edit H2"; the companion step-description SQL landed (Uj: "4c step description UPDATE 1") but the code edit awaited the next commit/image.
- **what:** `flag_page_image_rebuild` no-ops for non-page scope, so section-scope imagery landings never triggered the page rebuild that would surface them (observed live: zero flag-created needs_page in 30h after the two illustrations landed). Section scope_refs carry the page as a prefix (`index:1`), so the fix is a prefix-split: map scope 'section' to its page and fall through to the existing page path — no new emit code. Edit H2 + slice4c align the file header comment and the agent-definition step description with the new behaviour (cosmetic-drift discipline: descriptions must match deployed behaviour).
- **sources:** gobatch_05_flag_section_scope.md; slice4c_step_description.sql; running_notes_scheme_to_components(55).md#Tp #Ui #Uj
- **relations:** section-scope imagery pipeline; work-item crafting conventions.
- **verify-later:** flag_page_image_rebuild_action.go deployed body; image-build-handler flag_rebuild step description text.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Presigned-URL expiry and deploy-time asset localisation
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Ud) 2026-07-06: "w9_06: t/f/f both pages … localisation verified on content; Edit F deployed → recurrence prevention live → THREAD CLOSED"; w9_04 RETURNING "UPDATE 18, every url now /assets/images/…".
- **what:** A whole failure class: `assets.url` stored the presigned B2/S3 URL from generation (X-Amz-Expires=604800 — dies in seven days), while the asset-deployer had already committed the optimised file into the site repo under a key-derived local name. Renders that resolve from assets.url therefore embed URLs that die; heroes escaped only by being shadowed by a legacy local path. The fix is two-sided: w9_04 backfill flips all 18 idea.uk rows to `/assets/images/<key-hyphenated>.jpg`, preserving the unsigned S3 object path into `storage_path` (+ storage_provider), then a rebuild; Edit F makes `deploy_image_asset` record the committed local URL on the asset row at every future deploy, for ALL kinds (best-effort — a failure must not fail the deploy). Applies platform-wide to any site without the legacy hero_url shadow.
- **sources:** w9_03_assets_schema_and_inventory.sql; w9_04_backfill_flip.sql; gobatch_03_deploy_asset_localise(1).md; running_notes_scheme_to_components(55).md#Tu #Tw #Tz #Ua #Ub #Ud
- **relations:** legacy hero_url shadow; section-scope imagery pipeline; storage-architecture (S3/B2 refs preserved in storage_path).
- **verify-later:** deploy_image_asset_action.go post-commit UPDATE; assets rows url vs storage_path forms across sites.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Legacy site-level hero_url shadow (last-write-wins per purpose)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Notes (Tz/Ub): "hero_url lives at SITE level (sites.content_data …), merged beneath section data; template `or` picks it over the presigned background_image … last-write-wins per purpose → site-wide hero currently = hero-about's image everywhere."
- **what:** A legacy mechanism that both saved and distorts heroes: image deploys historically wrote `purpose+"_url"` keys (e.g. hero_url) into site-level `sites.content_data`, which the ContentData-priority merge (component_library.go ~736) supplies to templates ahead of the schema-resolved `site_assets.hero` value. Consequences: hero renders stayed local-path (immune to presigned expiry) but every page shows the same last-written hero image; the per-page hero assets sat unconsumed. Banked as a known quirk (per-page heroes = a later improvement); render-neutral to the localisation flip.
- **sources:** running_notes_scheme_to_components(55).md#Tx #Ty #Tz #Ub; w9_02_deployer_and_shadow.sql; w8_09_hero_exposure.sql
- **relations:** presigned-URL expiry; ensureAssets (content_data fallback is gap-fill, hero/logo only).
- **verify-later:** sites.content_data hero_url keys; component_library.go merge priority; whether per-page heroes were ever wired.

<!-- SOURCE: U05_content_quality_linking.md -->
### Hero image resolver (June-02) — images, not CTAs
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** running_notes_16(1) Part 1: "the 'hero resolver' (June-02) is about hero IMAGES, not CTA URLs".
- **what:** Per-page hero/logo image resolution in plan_sections ensureAssets (site_plan_imagery + assets join) plus flag_page_image_rebuild to re-render when an image lands — previously pages rendered the static /assets/images/hero.jpg fallback. Explicitly disambiguated from the CTA-URL bug (a costly early conflation). After Part 2, image_landed re-renders ride the no-LLM path.
- **sources:** running_notes_16_content_quality_and_internal_linking(1).md#part-1; NOTES(44) Part A; HANDOFF_2026-06-09(2).md#june-02-actions
- **relations:** no-LLM re-render path; site_plan_imagery pipeline (imagery unit).
- **verify-later:** plan_sections ensureAssets; flag_page_image_rebuild_action.go.

<!-- SOURCE: U09_adoption.md -->
### Imagery subsystem assessment (single hardcoded adapter and its gaps)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Part 1 assessment (descriptive) + "Phases 0 and 1 are complete and end-to-end verified… 2 onwards not started" (PLAN_imagery progress, 2026-05-08).
- **what:** One `DynamicImageAdapter` (Stability SDXL hardcoded, request body fixed — no negative prompt, seed, img2img, LoRA, variants); `assets` UNIQUE (site_id, purpose) blocks multi-image purposes; planner asks for exactly `{logo, hero_home}`; components declare only hero_image; features `icon` strings render nowhere; misleading non-consumed image-adapter config files; two image-generator agent rows (one placeholder). Enumerated structural gaps: multi-purpose model (asset_key), provider/model router mirroring the aiservice text pattern, richer request shape, planner imagery_plan, diverse input_schema image needs, icon/SVG path (three approaches), infographic-generator agent, design_intent.imagery_direction wiring, crawled-imagery persistence, per-vertical LoRA (018 plan, blocked on adapter model selection).
- **sources:** PLAN_imagery_loop_closure.md#part-1, old2/PLAN_imagery_loop_closure(1).md
- **relations:** imagery loop-closure phases; emit_imagery build-time trigger; adoption image mirror
- **verify-later:** internal/adapters/imagegenerator/dynamic_adapter.go; assets unique indexes; ImagePurposes map

<!-- SOURCE: U09_adoption.md -->
### Imagery loop closure phases 0–6 (spec-to-delivery audit-and-fix)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Phase 0.1/0.2 ✅ verified; 1.1–1.5 ✅ verified (Phase 1.5 needed an unplanned hotfix); 2 onwards not started" (2026-05-08).
- **what:** Sequenced plan: Phase 0 wire what exists (imagery_direction prepended to prompts; origin_model populated — verified); Phase 1 algorithmic discovery checks (unfulfilled_image_prompt, placeholder_image_in_use, image_url_404) routed to image-build-handler — verified, including the 1.5 hotfix (missing output_mapping made every dispatch-path image silently {stored:false}); Phase 2 asset locking (locked_at/lock_type mirroring page_components) + asset_key multi-image readiness; Phase 3 adoption-image-mirror (persist crawled imagery as origin_type='adopted' — today it is captured then discarded, "throwing away the best reference material we have"); Phase 4 text-only visual-auditor imagery category; Phase 5 vision-capable LLM path (GenerateTextWithImages); Phase 6 dedicated `imagery-quality-auditor` agent (sibling of visual-design-auditor, TOP-5 contract, max_fix_attempts 2, lock/origin_type honouring). Decisions locked: separate auditor agent; 2 regen attempts; mirror page_components locking exactly; per-section granularity deferred. Verification finding parked: dispatch wasn't claiming triaged imagery items while page items queued.
- **sources:** PLAN_imagery_loop_closure.md#part-2, old2/PLAN_imagery_loop_closure(1).md
- **relations:** emit_imagery_items (closed the build-time trigger gap the same shape as Phase-1 checks); C4 blank card images (imagery-to-card linkage open); 053 added lock cols to assets
- **verify-later:** discovery_checks check_unfulfilled_image_prompt.go etc.; assets.locked_at/lock_type/asset_key columns; imagery-quality-auditor existence (expected absent)

<!-- SOURCE: U09_adoption.md -->
### Build-time imagery trigger (emit_imagery_items + imageryplan.go shared selection)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "The imagery trigger had the identical bug and is fixed in the same deploy… deployed 2026-05-26, chassis v1.0.1047" (FOCUS_design_composition §3A).
- **what:** `write_site_plan` records planner image requests in `site_plan_imagery` (flattenImageryBlock) but nothing on the build path acted on them — needs_imagery came only from the loop's `unfulfilled_imagery_plan` check (capped 20/pass). `emit_imagery_items_action.go` emits at plan time; `imageryplan.go` is a shared package holding row selection, priority/severity classification, brand_update rule, item_key and spec body used by both the build emitter (status triaged) and the loop check (status detected) — the anti-drift pattern. Priority bands (index hero 65 … section 98) put imagery before the terminal needs_rerender (99). Known asymmetry: emit_imagery has no site-level no-backfill guard like emit_design's.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#3A, README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** imagery loop closure; work-site-orchestrator monolith mapping (imagery was the "same-shaped gap as design")
- **verify-later:** emit_imagery_items_action.go, imageryplan.go; site_plan_imagery rows on a fresh build

<!-- SOURCE: U10_imagery.md -->
### Imagery loop-closure programme (Phases 0–6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Progress (updated 2026-05-14) — Phase 2G + Phase 2H are operationally verified end-to-end on robot-hands.com … Phases 3, 4, 5, 6 of the outer plan not started"; Phase 3 "⏸ deferred 2026-05-14".
- **what:** The sequenced master plan for closing the gap between what the planner/spec asks for in imagery and what is delivered: Phase 0 (wire unread data), 1 (algorithmic discovery checks), 2A–2H (schema + pipeline refactor + structured plan imagery + request shape), 3 (adoption mirror), 4 (text-only auditor awareness), 5 (vision LLM path), 6 (imagery-quality-auditor). Each phase shippable alone; LLM phases gated on algorithmic checks working.
- **sources:** PLAN_imagery_loop_closure.md#Progress, PLAN_imagery_loop_closure.md#Phase-summary-table, STATUS_imagery_2026-05-12.md#At-a-glance
- **relations:** superseded in spirit by the best-in-class programme (I0–I8) which renumbered to avoid collision; feeds imagery-quality-auditor, adoption image mirror.
- **verify-later:** phases table vs live code: `platform/orchestration/actions/generate_image_actions.go`, `discovery_checks/`, `assets` schema, image-build-handler agent_definitions row.

<!-- SOURCE: U10_imagery.md -->
### imagery_direction prompt prepend (Phase 0.1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "0.1 — read imagery_direction ✅ delivered, verified 2026-05-08"; asset row shows origin_prompt beginning with the direction text.
- **what:** `generate_image` reads `site_specs.design_intent.imagery_direction` and prepends it to the subject prompt ("Style direction: … Subject: …", later unlabeled with a 200-char SDXL-aware sentence-boundary cap). Closed the "webdesign-agent writes imagery taste, image-generator ignores it" gap. Later superseded per-site by the imagery_style_guide when present (one brand voice, no double prepend).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0, old/PHASE_0_BUNDLE_README.md, STATUS_imagery_2026-05-08.md#Today's-verification
- **relations:** imagery_style_guide brand guide; per-kind prompt gating (direction gated OFF icons/logos).
- **verify-later:** `getImageryDirectionForSite` / `composeImagePromptWithDirection` in generate_image_actions.go; assets.origin_prompt on recent generations.

<!-- SOURCE: U10_imagery.md -->
### Asset provenance population (origin_prompt / origin_model)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "origin_model='sdxl' — Phase 0.2 column write happened" (2026-05-08); "origin_model propagation — assets.origin_model carries real provider/model" (2026-05-20 close-out).
- **what:** `store_asset` populates `assets.origin_prompt` (fixing a pre-existing bug where it was silently dropped — every row was NULL) and `assets.origin_model`; later extended so the adapter returns provider/model_id and workflows propagate it (`banana/gemini-3-pro-image-preview` vs `sdxl`) instead of a hardcoded literal. Provenance is the substrate for spec-vs-delivery audits.
- **sources:** old/PHASE_0_BUNDLE_README.md#Phase-0.2, TODO_imagery_followups.md#What-shipped-this-session, STATUS_imagery_2026-05-08.md
- **relations:** imagery discovery checks read it; imagery-quality-auditor (future) compares it to delivered image.
- **verify-later:** StoreAssetAction in v3_site_actions.go; `SELECT origin_model, origin_prompt FROM assets` distribution.

<!-- SOURCE: U10_imagery.md -->
### Algorithmic imagery discovery checks (Phase 1 trio)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Phases 1.1–1.4 all "✅ delivered" 2026-05-08 (1.2/1.3 "partial — check fires; symptom-path site needed").
- **what:** Three Go discovery checks catch spec-to-delivery gaps without LLM cost: `unfulfilled_image_prompt` (planner asked, no asset), `placeholder_image_in_use` (fallback path rendered, no asset), `image_url_404` (HTML references an image no assets row backs, DB-only version). All follow the DiscoveryCheck interface, register via init(), and were appended to design-discovery-agent's run_checks. A longer wishlist of ~12 further checks (alt-text, dimensions, orphans, cross-site contamination, multi-image underfill) was catalogued and mostly remains unbuilt.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1, FOCUS_imagery_assessment_1_.md#13.2, old/phase1/phase_1_register_imagery_checks.sql
- **relations:** check_unfulfilled_imagery_plan (2G successor for the new shape); image_source_unsatisfiable and component_template_corrupted (later siblings).
- **verify-later:** `platform/orchestration/actions/discovery_checks/check_*.go`; design-discovery-agent run_checks array.

<!-- SOURCE: U10_imagery.md -->
### asset_key multi-image model (Phases 2B–2D)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2B ✅ 2026-05-09; 2C ✅ deployed 2026-05-09; 2D ✅ applied 2026-05-10" with migration sanity checks documented.
- **what:** Broke the one-asset-per-purpose-per-site constraint: `assets.asset_key` column (backfilled from purpose), new unique index `(site_id, asset_key) WHERE active`, StoreAssetAction ON CONFLICT switched to asset_key, then the old `(site_id, purpose)` unique index dropped. Enables N heroes/icons/illustrations per site, with `(purpose, asset_key)` split (canonical hero = hero/hero; variant = hero/hero_about). Strict production apply order documented (2A→2B→2C deploy→verify→2D).
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2B/2C/2D, PLAN_imagery_loop_closure.md#Phase-2, old/phase2/2E/phase_2e_store_asset_action.diff
- **relations:** hero-variant routing (2E) consumes it; DeployedWebPath derives filenames from asset_key.
- **verify-later:** `\d assets` indexes; StoreAssetAction ON CONFLICT target.

<!-- SOURCE: U10_imagery.md -->
### Hero-variant routing through image-build-handler (2E)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2E ✅ delivered, verified 2026-05-12" with the full hero_about end-to-end trace.
- **what:** Made `hero_<page>` variants routable: `check_unfulfilled_image_prompt` classifies logo / hero_home / hero_<page> into needs_logo / needs_hero_image / unfulfilled_hero_variant; new `hasActiveAssetForAssetKey` helper (purpose-level check gave false positives); `deploy_image_asset` derives per-variant paths (`assets/images/hero-about.jpg`, `_`→`-`); StoreAssetAction gains `asset_key_field` JSONPath config; a third variant branch added to the image-build-handler workflow (spawn/call/store/deploy) leaving logo/hero branches untouched.
- **sources:** STATUS_imagery_2026-05-12.md#Phase-2E, old/phase2/2E/check_unfulfilled_image_prompt.go, old/phase2/2E/phase_2e_image_build_handler_variant_path.sql
- **relations:** needs_imagery branch (2G.5) later sits in front of it; known gap — variant chain doesn't pass site_id (imagery_direction not prepended for variants).
- **verify-later:** image-build-handler workflow branches in agent_definitions; imagery_helpers.go.

<!-- SOURCE: U10_imagery.md -->
### Spawned asset-deployer deploy pattern / storage-env isolation (2F)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2F ✅ deployed + verified 2026-05-12"; boxed warning 2026-07-10 "Where deploys run (by design, don't 'fix' this)".
- **what:** The chassis pod deliberately carries NO `IMAGE_BUCKET`, so it builds no storage client and inline `deploy_image_asset` fails there ("storage client not available") — by design. Deploys run in a spawned `asset-deployer` child into which `spawn_actions.go` injects S3/B2 env via the `isStorageEnabledAgent` list. 2F replaced three inline deploy step pairs in image-build-handler with spawn+call pairs targeting asset-deployer. Hand-triggering asset-deployer standalone fails because it skipped the injection — a triggering mistake, not a bug.
- **sources:** PLAN_imagery_loop_closure.md#2F, PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS, STATUS_imagery_2026-05-08.md#[BLOCKER]-Storage-architecture-mismatch
- **relations:** brand_head mode rides the same agent; storage-architecture (doc 032).
- **verify-later:** `agentbase/agent.go:294`, `spawn_actions.go` isStorageEnabledAgent, 107_image_build_handler.sql:725 comment.

<!-- SOURCE: U10_imagery.md -->
### site_plan_imagery table (2G.1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.1 — site_plan_imagery table ✅ delivered 2026-05-12"; live `chk_kind` re-verified 2026-07-08.
- **what:** Sibling of site_plan_directives holding structured imagery requirements: scope (site|page|section) + scope_ref, key (→asset_key), kind CHECK enum (logo|hero|illustration|icon|infographic, later +sprite_sheet), required prompt, JSONB style_hints/constraints, ordering, source CHECK (llm|classifier|manual|adoption), lock columns with the same lock-transfer treatment, unique on (plan_id, scope, COALESCE(scope_ref,''), key). `product` deliberately excluded (products come from the affiliate_products resolver, not the planner). Kind enum is mirrored in Go (`validImageryKinds`) — constraint and mirror change together.
- **sources:** PLAN_imagery_phase_2g.md#Schema, old/phase_2g_step1_site_plan_imagery.sql, SQL_2026-07-12_add_sprite_sheet_kind.sql
- **relations:** planner imagery block writes it; check_unfulfilled_imagery_plan reads it; five-place new-kind checklist.
- **verify-later:** `\d site_plan_imagery`; chk_kind vs validImageryKinds in write_site_plan_action.go (~line 183).

<!-- SOURCE: U10_imagery.md -->
### Planner imagery block (2G.3 prompt extension)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.3 ✅ delivered 2026-05-13 (with max_tokens bump + path fix)"; 2026-07-08 ground truth "planner prompt carries the Imagery Block + decomposition rule; max_tokens now 16000".
- **what:** build-site-planner's JSON output gains an `imagery` key (site[] / pages{} / sections{} entries with key, kind, prompt, optional style_hints/constraints) in the same LLM call as pages/design_direction — a single call, no separate imagery planner. Replaces the flat `image_prompts:{logo,hero_home}` contract that had hero/logo-only names baked in. max_tokens raised 4000→8000 (JSON truncation on a 14-page roadmap) and later to 16000. Legacy image_prompts continues to be emitted during transition.
- **sources:** PLAN_imagery_phase_2g.md#Planner-output-shape, PLAN_imagery_loop_closure.md#Application-status, FOCUS_imagery_assessment_1_.md#4.1
- **relations:** one-entry-one-image decomposition rule; planner key stability problem; sprite_sheet planner emission (future).
- **verify-later:** build-site-planner default_config prompt_template "## Imagery Block"; sql_for_agents/053 patches.

<!-- SOURCE: U10_imagery.md -->
### flattenImageryBlock write path + imagery lock transfer (2G.2)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "✅ deployed 2026-05-12; path fix on 2026-05-13 (function looked up data['imagery'] at top level rather than walking wrapper shapes via findDirectiveTree)".
- **what:** `write_site_plan` walks the planner's imagery block and inserts site_plan_imagery rows in the same transaction as pages/sections/directives (`flattenImageryBlock` + `insertImageryRow` enforcing the kind enum), and transfers locks from the previous current plan's locked imagery rows matched on (scope, scope_ref, key) — locked HITL prompt edits survive plan rebuilds.
- **sources:** PLAN_imagery_phase_2g.md#write_site_plan-extension, PLAN_imagery_loop_closure.md#2G
- **relations:** site_plan_imagery table; content-governance lock semantics.
- **verify-later:** write_site_plan_action.go flattenImageryBlock/insertImageryRow; lock-transfer behaviour on replan.

<!-- SOURCE: U10_imagery.md -->
### check_unfulfilled_imagery_plan discovery check (2G.4)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.4 ✅ delivered 2026-05-14 (8 work items emitted on first run; correct priority ordering)"; Pipeline:"build" fix confirmed in code 2026-07-08.
- **what:** Walks the current plan's site_plan_imagery rows, emits one `needs_imagery` work item per row lacking a matching active asset (via hasActiveAssetForAssetKey), capped at 20/pass, priority-banded (site logo 70 → index hero 65 → site other 75 → page hero 80 → page other 90 → section 100) mirroring legacy classifyPromptKey bands. `computeAssetKey` namespaces deeper keys (`page.about.illustration_team_values`, `section.home.2.icon_precision`) while keeping hero/logo names flat for backward-compatible deploy paths. Dedup key `needs_imagery:<scope>:<scope_ref|->:<key>`.
- **sources:** PLAN_imagery_phase_2g.md#Discovery-check-1, PLAN_imagery_loop_closure.md#Decisions/#2G, TODO_imagery_followups.md#7
- **relations:** legacy unfulfilled_image_prompt runs in parallel during transition (both call hasActiveAssetForAssetKey to avoid double work); pipeline-field fix.
- **verify-later:** check_unfulfilled_imagery_plan.go (hardcoded Pipeline "build"); design-discovery-agent run_checks.

<!-- SOURCE: U10_imagery.md -->
### needs_imagery branch in image-build-handler (2G.5)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.5 ✅ delivered 2026-05-14 (with two hotfixes — optional input_mapping + store_asset purpose)"; first asset a12b5d71 through the new path 2026-05-14.
- **what:** A new branch alongside (not extending) the variant chain: check_item_type_imagery → spawn_image_gen_imagery → call_imagery_gen (site_id passed so imagery_direction prepends; kind/style_hints/constraints pass through) → brand-update conditional store → shared spawn_asset_deployer tail. Brand-asset update routed by a `spec.brand_update` boolean computed at discovery (site scope OR index-page hero). Hotfixes established the `?`-suffix optional input_mapping convention and exposed that store_asset lacked `purpose_field` (initially hardcoding purpose:"hero", blocking kind=logo items — later fixed by the purpose_field workflow fix, 2026-05-20). A future refactor option is recorded: collapse the three legacy branches into needs_imagery ("always fix legacy with modern").
- **sources:** PLAN_imagery_loop_closure.md#2G/#Step-5-workflow, PLAN_imagery_phase_2g.md#image-build-handler-extension, TODO_imagery_followups.md#What-shipped-this-session
- **relations:** hero-variant branch (2E); `?` optional-mapping convention; purpose_field fix.
- **verify-later:** image-build-handler workflow JSON; store_imagery_asset purpose_field config.

<!-- SOURCE: U10_imagery.md -->
### Legacy image_prompts age-out check (check_legacy_image_prompts_aspect)
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** Old plan versions: "Migration off legacy image_prompts: Age-out via check_legacy_image_prompts_aspect, registered last"; live plan: "2G.6 ❌ retired as scoped. Reframed 2026-05-13… one string out of a JSON array, no code change."
- **what:** Originally a dedicated discovery check emitting `needs_replan` for sites still on `site_specs.site_plan.image_prompts` (deliberately registered LAST to avoid replan churn before the planner extension shipped). Reframed and retired: "is a site on legacy?" is not a fault signal — the existing checks already detect brokenness on both paths; migration became an operational deregistration decision (pull `unfulfilled_image_prompt` from run_checks once it reliably finds zero gaps).
- **sources:** old/PLAN_imagery_loop_closure(3).md#Decisions, PLAN_imagery_phase_2g.md#Discovery-check-2, PLAN_imagery_loop_closure.md#Decisions (2026-05-13 reframe)
- **relations:** superseded by "operational deregistration, not a dedicated check"; transition dual-check running.
- **verify-later:** confirm no check_legacy_image_prompts_aspect.go exists; whether unfulfilled_image_prompt is still registered.

<!-- SOURCE: U10_imagery.md -->
### pageflow-builder retirement
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** "Decision marker: 2026-05-12 — user agreed pageflow-builder is being left behind"; snapshot saved to pageflow-builder_2026-05-12.txt.
- **what:** The legacy monolithic site builder (inline deploy_image_asset, hardcoded generate_logo/generate_hero_image, sequential 20-iteration page loop, writes site_specs.site_plan directly bypassing the plan-domain tables) is deliberately not extended with the 2G imagery shape. Architecture converges on build-site-planner/plan-builder + triaged work items + page-build-handler + image-build-handler. Sites it built stay on the legacy check path until they age out; a full row snapshot exists as the recovery reference. The classifier's `recommended_builder` default was a noted loose end.
- **sources:** PLAN_imagery_phase_2g.md#On-leaving-pageflow-builder-behind, old/pageflow-builder_2026-05-12_NOTES(1).sql, old/pageflow-builder_2026-05-12.txt
- **relations:** superseded by plan-domain + dispatch architecture; robot-hands rebuild dropped its recommended_builder key.
- **verify-later:** pageflow-builder agent_definitions row status; any remaining live traffic.

<!-- SOURCE: U10_imagery.md -->
### Image-generator request shape + per-kind defaults (Phase 2H)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2H ✅ delivered (action layer) 2026-05-14 partial — chassis confirmed; adapter binary unconfirmed"; later provider work (2026-05-20) shipped the adapter side.
- **what:** Extends the generation request beyond {prompt,width,height}: v1 fields negative_prompt, seed, reference_image_uri (pass-through), cfg_scale, steps; Go-side `kindDefaults` map per kind (logos get people/text/watermark negative prompts; icons tighter aspect; heroes unchanged) with caller spec overriding defaults; style_hints.aspect_ratio drives dimensions and constraints feed the negative prompt. style_preset/samples/safety_mode deferred. Defaults deliberately live in Go, not a config table.
- **sources:** PLAN_imagery_loop_closure.md#2H, STATUS_imagery_2026-05-12.md#Phase-2H-(proposed), TODO_imagery_followups.md#4
- **relations:** provider abstraction; parseAspectRatio whitelist fix; constraints "informational only" decision.
- **verify-later:** kindDefaults/resolveKind/parseAspectRatio in generate_image_actions.go; adapter field mapping in dynamic_adapter.go.

<!-- SOURCE: U10_imagery.md -->
### parseAspectRatio SDXL v1.0 whitelist snap
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Elevated HIGH 2026-05-18 ("16:9 → 1024×576, SDXL rejects"); Turn 24 (2026-07-11) refers to "pre-SDXL-snap-fix residue", implying the fix landed.
- **what:** parseAspectRatio snapped to multiples of 64 rather than SDXL v1.0's strict dimension whitelist (1024×1024, 1152×896, 1344×768, …), so planner-emitted aspect_ratio hints produced rejected sizes and blocked hero generation — a regression enabled by the item-4 prompt patch (heroes previously fell through to valid kindDefaults). Fix: snap to the nearest whitelist pair matching the requested orientation.
- **sources:** TODO_imagery_followups.md#5, RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** Phase 2H request shape; planner prompt patch change 1 (aspect moved into style_hints).
- **verify-later:** whitelist logic in generate_image_actions.go; test 16:9→1344×768.

<!-- SOURCE: U10_imagery.md -->
### Adoption image mirror (Phase 3)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "3 — adoption image mirror ⏸ deferred 2026-05-14. Not cancelled… reference_image_uri plumbing preserved as a forward-compat hook."
- **what:** Stop discarding crawled imagery on adoption: new `mirror_adoption_images` action (download crawl images, upload to S3, insert assets rows with origin_type='adopted', origin_url, attribution/license; caps 50 images/site, 5MB each), wired into apply_adoption_plan; backfill check `check_crawled_images_discarded` routed to a new one-step `adoption-image-mirror` agent. Adopted images become img2img/style references and auditor signals. Deferred because current adopted sites carry minimal imagery.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3, FOCUS_imagery_assessment_1_.md#7/#9-item-9
- **relations:** reference-image style anchoring; adoption-pipeline category (site crawling).
- **verify-later:** existence of mirror_adoption_images_action.go, adoption-image-mirror agent row; assets rows with origin_type='adopted'.

<!-- SOURCE: U10_imagery.md -->
### Visual auditor imagery awareness (Phase 4, text-only)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "4 — visual auditor sees imagery (text-only) — not started".
- **what:** Extend visual-design-auditor's load_design_context SQL to include assets rows (unlocked, generated/adopted), imagery_direction, and site_plan_imagery, and add IMAGERY as a sixth check category with algorithmic-check results passed through to avoid double-flagging; tune on 5–10 sites before enabling fixes (≥80% accuracy target). Today the auditor's context contains zero image data — it cannot notice a missing or off-brief hero.
- **sources:** PLAN_imagery_loop_closure.md#Phase-4, FOCUS_imagery_assessment_1_.md#13.1/#13.3
- **relations:** imagery-quality-auditor (option B chosen as eventual answer); design-composition auditors.
- **verify-later:** visual-design-auditor load_design_context SQL in agent_definitions.

<!-- SOURCE: U10_imagery.md -->
### Vision-capable LLM path (Phase 5)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "5 — vision-capable LLM path — not started".
- **what:** Foundational capability for auditors to actually look at images: extend aiservice.AIService with image inputs, implement Anthropic vision content blocks, prefer extending execute_llm_prompt with an image_urls_field over a new action, refresh presigned URLs immediately before calls, tag vision_call:true in llm_call_log for cost separation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** required by imagery-quality-auditor and sprite-sheet vision auto-verify (I2.4/I8).
- **verify-later:** aiservice interface; anthropic.go vision support.

<!-- SOURCE: U10_imagery.md -->
### imagery-quality-auditor agent (Phase 6 / I8)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "6 — imagery-quality-auditor agent — not started"; I8 in the best-in-class plan also not started.
- **what:** A vision-capable sibling of visual-design-auditor under design-audit-agent, dedicated to imagery: categories direction_mismatch / brand_mismatch / inconsistency / quality / inappropriate; max_fix_attempts 2; findings route to image-build-handler regeneration (different prompt/seed/negative prompt) escalating to needs_human_review; honours locks and origin_type='uploaded'; gated rollout. I8 adds sprite-sheet cell verification and brand-guide reference comparison. Chosen over extending the existing auditor (separate TOP-5 cap; only imagery pays vision cost).
- **sources:** PLAN_imagery_loop_closure.md#Phase-6, FOCUS_imagery_assessment_1_.md#13.4, PLAN_imagery_best_in_class.md#Phase-I8
- **relations:** vision path (Phase 5); imagery_style_guide as the audit standard; improvement-loop pass caps.
- **verify-later:** no imagery-quality-auditor row in agent_definitions (expected absent).

<!-- SOURCE: U10_imagery.md -->
### Image provider abstraction and kind→provider routing
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 2026-05-20 close-out: "Provider abstraction — internal/adapters/imagegenerator/{provider,stability,banana}. dynamic_adapter.go routes by kind: icon → Banana, everything else → Stability. Proven end-to-end."
- **what:** The originally hardcoded Stability-only adapter (env-driven; the image-adapter.yaml/agent-definition config blocks are misleading and unread) was refactored into provider packages with kind-based routing: flat kinds (icon, later logo/illustration/infographic per a committed-but-then-pending routing change, and sprite_sheet) → Google Banana `gemini-3-pro-image-preview`; photographic kinds → Stability SDXL. The provider Request carries ReferenceImageURIs (Banana native reference-image support). Known opens: Stability provider timeout 60s vs old 120s; circuit breaker not threaded into provider clients.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session, FOCUS_imagery_assessment_1_.md#1.1–1.3, PLAN_imagery_best_in_class.md#2
- **relations:** icon model lessons drove it; 2H request shape; multi-provider routing beyond two providers still deferred.
- **verify-later:** internal/adapters/imagegenerator/ package layout; adapter switch cases; timeout value.

<!-- SOURCE: U10_imagery.md -->
### Icon generation lessons and image-model comparison
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Verdict (2026-05-18): Fix A is insufficient… SDXL ignores style instructions"; "Final icon batch state (verified 2026-05-20): all six icons banana/gemini-3-pro-image-preview, visual gate passed."
- **what:** SDXL is the wrong tool for flat-vector icons — strong photorealism bias on concrete subjects, multi-panel drift, no real transparency. A full model comparison (SDXL, SD3.5, FLUX schnell/dev/pro, DALL-E 3, Imagen 3, Nano Banana Pro 2, Midjourney, LLM-SVG) ranked FLUX schnell cheapest-good and Banana best for reference-conditioned sibling consistency; decision: plumb reference images AND switch icon generation to Banana. Related fixes: purpose_field so icons store as purpose=icon (240×240, not hero 1600×900); kindDefaults icon dimensions; jpg-vs-png note for thin line art.
- **sources:** TODO_imagery_followups.md#23, old/001_image_model_comparison.md, TODO_imagery_followups.md#Final-icon-batch-state
- **relations:** provider abstraction; transparency abandonment; LLM-SVG sleeper option; reference-image anchoring.
- **verify-later:** ImagePurposes["icon"]; assets rows origin_model for icon assets.

<!-- SOURCE: U10_imagery.md -->
### LLM-generated SVG icon path (sleeper option)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Sleeper option for later: LLM-generated SVG for icons… Worth a focused experiment once the immediate work is shipped" (2026-05-18).
- **what:** Icons are vector by nature; an LLM (Claude/GPT) writing SVG markup directly bypasses the entire convince-a-diffusion-model problem at ~$0.001–0.005/icon, crisp at any size, no copyright concern. Was the analyst's original recommendation (c2) before the user chose the Banana route; retained as a possible future replacement of the raster icon pipeline. Implies per-kind generation pipeline routing.
- **sources:** TODO_imagery_followups.md#23 (options c1/c2, recommendation), FOCUS_imagery_assessment_1_.md#9-item-6
- **relations:** superseded for now by Banana raster icons; Lucide covers UI chrome regardless.
- **verify-later:** none (idea only).

<!-- SOURCE: U10_imagery.md -->
### Diffusion transparency abandoned → flat-grey chip icons
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Transparency abandoned as too fragile — image models paint a transparency checkerboard into RGB (confirmed: icon_cycle_time mode=RGB has_alpha=False). Decision: option 2, 'embrace the box'" (2026-05-20).
- **what:** Image models cannot produce true alpha; requesting transparent backgrounds yields painted checkerboards. Locked decision: icons generate on a flat selectable grey background (#EEEEEE bg / #4A4A4A line) and are presented inside a styled CSS chip; the planner prompt and all existing icon specs were patched accordingly. The lesson recurs for sprite sheets (flat selectable background, NOT transparent).
- **sources:** TODO_imagery_followups.md#Icon-background-resolution, SCOPE_I2_sprite_sheets.md#3 (planner prompt), CONTEXT_PACK_imagery_sprite_sheet.md#Attach—docs
- **relations:** icon lessons; sprite-sheet prompt rules; CSS chip styling was left as site-template work.
- **verify-later:** planner prompt icon-background wording; icon assets' actual backgrounds.

<!-- SOURCE: U10_imagery.md -->
### Per-kind prompt gating and the five-place new-kind checklist
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "directionAppliesToKind… gates design_intent.imagery_direction OFF icons/logos" (shipped 2026-05-20); "PROVEN on real generations (icons carried palette, not medium)" (2026-07-11); sprite gating fix commit 4629aa17 proven in origin_prompt (Turn 31).
- **what:** Photographic brand direction contaminates non-photographic kinds (prepending it to an icon prompt makes the model paint a photo around the icon), so prompt composition gates per kind: hero/illustration/infographic get medium+mood+palette; icon and sprite_sheet palette only; logo nothing. Two gating functions (`directionAppliesToKind`, `styleGuide.directionForKind`) plus the DB constraint, Go mirror, adapter switch and ImagePurposes form the standing five-place checklist any new imagery kind must touch — the I2.0 lesson.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session, HANDOFF_imagery_best_in_class.md#Mechanisms, RUNNING_NOTES_imagery_best_in_class.md#Turn-29
- **relations:** imagery_style_guide supplies the gated content; sprite_sheet contamination near-miss is the cautionary case.
- **verify-later:** both gating functions list identical kind sets; grep for the five places when a new kind exists.

<!-- SOURCE: U10_imagery.md -->
### One-entry-one-image decomposition rule (planner prompt patch)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** planner_prompt_patch changes applied ("planner prompt carries the Imagery Block + decomposition rule", verified live 2026-07-08).
- **what:** A work item describes one deliverable: prompts must never ask for "a set of six icons" (SDXL renders one six-panel image — unusable but superficially successful). The planner prompt teaches per-entry single-image prompts, bans plural/counting phrasing (RULE 16), biases toward over-decomposition (unused icons are cheap, botched multi-panels expensive), moves aspect ratio to style_hints.aspect_ratio (the key Go reads) and demotes constraints to "informational only, reserved". The icon_cross_technology six-panel artifact and its cleanup SQL are the canonical example.
- **sources:** old/planner_prompt_patch_imagery.md, TODO_imagery_followups.md#25/#4
- **relations:** planner imagery block; SDXL whitelist fix; multi-entry sections remain the canonical way to express multiple images at one scope.
- **verify-later:** RULE 16 in the live planner prompt; absence of "set of" in site_plan_imagery.prompt.

<!-- SOURCE: U10_imagery.md -->
### Planner key stability across replans
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Symptom (2026-05-15): previous plan keyed hero_canonical; new plan called it brand_hero_canonical… discovery emitted a fresh work item"; no fix recorded since.
- **what:** The planner LLM freely chooses imagery `key` values, so replans rename equivalent concepts, discovery sees missing assets, and generations/orphan assets accumulate. Fix options ranked: (a) pass old plan's keys into the prompt with a reuse rule (lowest effort), (b) canonical key dictionary, (c) semantic concept matching at discovery time. Stale keys from this bug were cleaned up during the best-in-class rebuild.
- **sources:** TODO_imagery_followups.md#26, RUNNING_NOTES_imagery_best_in_class.md#Turn-24 (stale failed rows closed)
- **relations:** planner imagery block; replan-driven waste; asset orphan cleanup.
- **verify-later:** whether the planner prompt includes previous-plan keys; duplicate-concept assets on replanned sites.

<!-- SOURCE: U10_imagery.md -->
### Lucide icon strategy and validator wiring
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Lucide validator written (lucide_icons.go) — NOT yet wired" (2026-05-20), re-verified unwired 2026-07-08; still listed as an open I0 close-out item.
- **what:** The features grid renders icons as Lucide webfont glyphs (`<i data-lucide="{{.icon}}">`), not generated raster — the generated icon pipeline was never the right tool for it. Missing icons are LLM-invented Lucide names. Fix design: a single-source allowlist that is both the prompt's choice list and a pre-store `SanitizeFeatureIcons` sweep, plus optional render-time net; the allowlist must be verified against the bundled Lucide version. Blocked on identifying the content-generation step that fills features content_data. Icon strategy stays dual (D6): Lucide for UI chrome, generated sprites/raster for decorative glyphs.
- **sources:** TODO_imagery_followups.md#features-component-icons, old/verify_and_wire_lucide.md, PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** sprite sheets cover decorative glyphs; robot-hands rebuild was to carry the wiring.
- **verify-later:** callers of SanitizeFeatureIcons/ValidateLucideIcon outside lucide_icons.go.

<!-- SOURCE: U10_imagery.md -->
### Data-graph / chart pipeline (code-rendered, never diffusion)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Status: scoping only — not built" (2026-05-20); I4 "not started" as of 2026-07-12; runtime decision confirmed 2026-07-08 (go-echarts in-chassis).
- **what:** Hard constraint: diffusion models cannot plot real data — they fabricate values. Charts are a separate three-stage pipeline: fetch real series (EIA/FRED/per-vertical free-tier sources, stored for reproducibility + attribution) → code-render (go-echarts; static SVG/PNG always exists as fallback) → LLM editorial layer only (titles, callouts, annotations — never data values). Needs a `data-chart-generator`-shaped agent and deliberately does NOT add `chart` to site_plan_imagery kinds (charts are Lane B artefacts); `infographic` stays decorative-Banana and must never carry real numbers.
- **sources:** old/FUTURE_data_graph_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I4/#D1/D3, TODO_imagery_followups.md#Future-workstream
- **relations:** news imagery (I5) consumes it for data-driven stories; RUNBOOK B4 data-source keys.
- **verify-later:** no chart pipeline code expected; go-echarts dependency absence.

<!-- SOURCE: U10_imagery.md -->
### Product illustration pipeline (copyright-safe sketches)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Planned but not built (design work already done — reuse it)" (2026-07-08); I6 not started.
- **what:** Generate stylised product illustrations to avoid copyright/trade-dress exposure from scraped affiliate photos: discovery check `check_product_without_custom_illustration` (per-pass cap ~20), `product-illustration-handler` agent delegating to image-build-handler, `link_asset_to_product` action setting affiliate_products.custom_image_id, renderer precedence custom_image_id → cached_image_url. Stylisation is a hard-coded constraint, not a knob (D7): medium by product category (CAD-like / pencil / watercolour), altered viewpoint, in-context setting, no brand markings; img2img from the cached photo is v2-only under the derivative-work framing.
- **sources:** old/illustration/PLAN_product_illustration.md, PLAN_imagery_best_in_class.md#Phase-I6/#D7, STATUS_imagery_2026-05-12.md#Component-audit-finding
- **relations:** affiliate sites programme (resolver dependency); product components' query.affiliate_products socket; 3D reconstruction explicitly parked.
- **verify-later:** affiliate_products.custom_image_id usage; existence of the handler agent (expected absent).

<!-- SOURCE: U10_imagery.md -->
### Imagery best-in-class programme (G1–G9, D1–D8, phases I0–I8)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-12: "Phase I0 ✅ COMPLETE… Phase I1 ✅ COMPLETE, LIVE-VERIFIED… Phase I2 ⏳ IN PROGRESS… Phases I3–I8 not started."
- **what:** The 2026-07-08 successor programme raising fleet visual quality to best-in-class: nine goals (brand kit/logo permanence, data-accurate infographics, content-linked card imagery, graphic artefacts/sprites, copyright-safe product sketches, news imagery, performance budgets, accessibility/OG surface, quality loop) governed by eight user-confirmed design decisions (D1 code-rendered charts, D2 two lanes, D3 kind batches as text+CHECK, D4 brand guide as data, D5 logo lock, D6 dual icon strategy, D7 sketch constraints, D8 deploy-enforced budgets). Phases I0–I8, each acceptance-gated on robot-hands.com; companion running-notes/runbook/handoff/showcase document set maintained every turn.
- **sources:** PLAN_imagery_best_in_class.md, HANDOFF_imagery_best_in_class.md, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** builds on the loop-closure programme; RUNBOOK human-gate model; showcase docs quote its numbers.
- **verify-later:** phase status blocks vs live DB/site state; open runbook items B4/B5/B9/B10/B11.

<!-- SOURCE: U10_imagery.md -->
### Two lanes of imagery: plan-driven vs content-driven (Lane B)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** D2 "confirmed" 2026-07-08 as a decision; Lane B storage decision "generic entity_type + entity_id columns on assets — confirmed"; I3 not started.
- **what:** Everything built so far is plan-driven (fixed list decided at plan time). Card images, news charts, and product sketches are content-driven — attached to articles/news items/products, arriving continuously after the plan, prompts composed from the content itself plus the brand guide. Lane B generalises the affiliate custom_image_id pattern via entity_type+entity_id columns on assets, per-entity work item types, and content-sweeping discovery checks, sharing all generation/deploy/audit machinery downstream of the work item.
- **sources:** PLAN_imagery_best_in_class.md#3/#8, RUNNING_NOTES_imagery_best_in_class.md#Turn-2
- **relations:** content-linked card imagery (I3), news imagery (I5), product sketches (I6) are its instances.
- **verify-later:** assets table for entity_type/entity_id columns (expected absent yet).

<!-- SOURCE: U10_imagery.md -->
### imagery_style_guide — per-site brand guide as data (I1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "✅ PHASE I1 COMPLETE — LIVE-VERIFIED 2026-07-11… per-site imagery_style_guide driving generation with per-kind gating — PROVEN on real output."
- **what:** A site_specs aspect {palette, medium, mood, avoid, reference_asset_keys} distilled from design_intent, read by generate_image for every generation: photographic kinds get medium+mood+palette prepended, icons/sprite sheets palette only, logos nothing; the guide supersedes free-text imagery_direction when present; `avoid` terms feed the negative prompt (stronger channel than positive pleading); reference_asset_keys resolve to stable s3:// URIs (presigned URLs stripped back to bucket/key so anchors outlive the 7-day signature) and flow to Banana as style anchors. The single biggest lever for consistent professional look, per-site so sites diverge deliberately.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-17/#Turn-24, SQL_2026-07-10_robothands_imagery_style_guide.sql, SHOWCASE_technical_architecture.md#4
- **relations:** per-kind gating; reference-image anchoring; supersedes-at-runtime the Phase 0.1 free-text prepend.
- **verify-later:** imagery_style_guide.go; robot-hands site_specs aspect row; +style_guide log lines.

<!-- SOURCE: U10_imagery.md -->
### Logo permanence: generate → human-approve → lock (D5)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Logo: user-approved, LOCKED (assets.locked_at, lock_type=permanent); store-guard refuses overwrites" (2026-07-11, B6 done).
- **what:** One consistent logo for the life of a site is a policy, not a generation feature: the logo is generated, a human approves it via the runbook (A3 eyeball ritual), `locked_at` is set, and the assets upsert's `WHERE assets.locked_at IS NULL` guard refuses any future overwrite; auditors and regeneration paths must skip locked assets. Favicon and OG card are derived from the approved logo, never independently generated. robot-hands' May-8 logo was approved as-is and locked.
- **sources:** PLAN_imagery_best_in_class.md#D5/#Phase-I1, RUNBOOK_imagery_best_in_class.md#B6, RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** asset locking 2A supplies the columns; brand-head derivation consumes the locked logo.
- **verify-later:** robot-hands logo asset locked_at/lock_type; store guard in StoreAssetAction.

<!-- SOURCE: U10_imagery.md -->
### Brand-head derived assets (favicon + OG card)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "favicon.png/og-card.png serve 200; og:image + twitter:card injected into every head at render time" (I1 verification, 2026-07-11).
- **what:** `derive_brand_head_assets` action deterministically derives favicon (64×64 square resize) and OG card (1200×630, logo centred on a solid brand-palette colour; gradients rejected) from the locked logo bytes — no LLM — commits both to the site repo and records provenance rows (origin_model='derived-from-logo'). `injectBrandHeadTags` in render_site_components injects favicon/OG/Twitter head tags fleet-wide, idempotently. Runs via a `brand_head` mode branch on asset-deployer dispatched by a `needs_brand_head_assets` work item — the reusable pattern for any site (candidate auto-emit after logo lock).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-25/#Turn-26/#Turn-27, SQL_2026-07-11_asset_deployer_brand_head_mode.sql
- **relations:** logo permanence; sprite CSS head-link reuses the same injection + commit shape.
- **verify-later:** derive_brand_head_assets action registration; asset-deployer check_mode branch; live favicon/og-card on robot-hands.

<!-- SOURCE: U10_imagery.md -->
### Header logo resolution from plan imagery
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "header resolves the locked logo from plan imagery (logo-img live in the served header)" (2026-07-11); fix commit b00c150b.
- **what:** The header is a site component rendered by `render_site_components`, untouched by the page-level resolver fixes, and read the never-populated `sites.logo_url` — so sites showed a text mark despite a deployed logo file. Fix: `loadSiteDataFull` resolves the site-scope logo from site_plan_imagery→assets via `storage.DeployedWebPath` (never assets.url), keeping sites.logo_url as legacy fallback. Closed the long-standing "logo-in-header resolution gap" carried since 2026-05-27.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-23, FOCUS_imagery_assessment_1_.md#5.1 (gap origin), PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** image-role resolver (page-side sibling); DeployedWebPath convention.
- **verify-later:** loadSiteDataFull logo resolution; served header `<img>` on fleet sites.

<!-- SOURCE: U10_imagery.md -->
### Sprite-sheet bullets and list treatment (I2)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Phase I2 ⏳ IN PROGRESS — the active phase… I2.0 ✅… I2.1 ⏳ REGEN IN FLIGHT (Turn 31)… I2.2 NEXT BUILDABLE NOW" (2026-07-12/13).
- **what:** One coherent N×M glyph grid per site (3×3 @ 768², 256px cells; Banana — harnessing the model's gridded-image tendency), sliced by CSS `background-position`; bullets/nav via `::before` and `.sprite-<name>` classes — one generation, one asset, one stylesheet, no Go image cropping. Delivery deviation resolved twice: sprite CSS ships as a separate committed `/assets/css/sprites.css` + head `<link>` (css_snippets is a GLOBAL library with no site scoping; the per-site committed bundle is the house pattern). Cell-content alignment is THE risk, mitigated by ordered-grid prompt + human eyeball-and-assign gate (B11, cell_names_verified flag); vision auto-verify deferred to I2.4/I8. First generation was near-perfect (all 9 glyphs in reading order); its deploy failure spawned the ExtractActionInputs lesson.
- **sources:** SCOPE_I2_sprite_sheets.md, CONTEXT_PACK_imagery_sprite_sheet.md, PLAN_imagery_best_in_class.md#Phase-I2, RUNNING_NOTES_imagery_best_in_class.md#Turns-28–32, SQL_2026-07-12_seed_robothands_sprite_sheet.sql
- **relations:** five-place kind checklist (I2.0); brand-head commit pattern reused for sprites.css; referenced PLAN_imagery_sprite_sheet.md lives outside this unit.
- **verify-later:** chk_kind includes sprite_sheet; sprite-sheet-main.png 768×768 on robot-hands; sprites.css emit action existence.

<!-- SOURCE: U10_imagery.md -->
### Content-linked card imagery (I3)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Phases I3–I8: not started" (2026-07-12); card-crop decision confirmed 2026-07-08.
- **what:** Every linking card (blog index, news feed, tool directory) carries an image reflecting the content behind it, sharing a visual family with the content page. Confirmed approach: the card image is the article's asset re-cropped per purpose (one generation yields article hero, card crop ~800×450 WebP, OG crop), not a sibling generation. First real Lane B consumer; also clears the one remaining empty image slot on robot-hands (learning-center-index listing card).
- **sources:** PLAN_imagery_best_in_class.md#Phase-I3, RUNNING_NOTES_imagery_best_in_class.md#Turn-2/#Turn-13
- **relations:** two lanes (Lane B); news imagery reuses its mechanics; performance budgets set the card byte ceiling (≤60KB).
- **verify-later:** card kind/purpose in ImagePurposes (expected absent yet).

<!-- SOURCE: U10_imagery.md -->
### News imagery (I5)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** I5 not started; freshness rule confirmed 2026-07-08 ("no SLA… configurable news_image_grace_interval… working suggestion 6h").
- **what:** News ingestion attaches a per-item imagery decision via a small LLM classification: `chart` (data-driven story → I4 pipeline), illustration/photo (I3 pipeline), or none. Feed cards and article pages share the artefact. No SLA (ingest ~2×/day); after a configurable grace interval an item falls back to a brand-kit-derived default image so the feed never shows an empty slot.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I5, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** data-graph pipeline (I4), card imagery (I3), brand kit (I1); news→infographic backlog enhancement.
- **verify-later:** none yet (design only).

<!-- SOURCE: U10_imagery.md -->
### Image performance budgets (I7 / D8)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** budgets "confirmed as proposed" 2026-07-08 (hero ≤180KB, card ≤60KB, sprite ≤80KB, index above-fold ≤600KB); I7 not started.
- **what:** Per-kind byte and dimension budgets enforced at deploy (extend ImagePurposes with ceilings; WebP for photographic kinds) and policed by a new `image_weight_over_budget` discovery check routed to asset-deployer re-optimisation; responsive srcset + lazy loading in image-bearing templates; alt text required at generation time plus an `image_alt_text_missing` check; sprites amortise small art into one download.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I7/#D8, RUNBOOK_imagery_best_in_class.md#B5
- **relations:** OptimizeImageForWeb/ImagePurposes (existing enforcement point); accessibility goal G8.
- **verify-later:** ImagePurposes byte-ceiling fields (expected absent).

<!-- SOURCE: U10_imagery.md -->
### Image-role alias resolver + authoritative overlay
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "RESOLVED 2026-07-10 — the finding took THREE fixes, all deployed/applied… 16 distinct hero files, each referenced by exactly one page."
- **what:** There was no single contract for how a component gets its image — three incompatible patterns (legacy content_data.hero_url site-wide field; preset `site_assets.background/product_screenshot/...` sources nothing generates; components with no image slot) meant per-page heroes generated correctly but never rendered (same placeholder everywhere / empty src). Fix: `imageryplan.ImageRoleForPath` shared alias table normalising generic image field names to the "hero" role; page-aware `ensureAssets` resolving page hero → site hero fallback; `planSection` injecting the resolved hero under legacy alias keys into resolved_data, which the renderer merges LAST (the designed authoritative overlay), defeating the stale site-wide hero_url. Planner-side key alignment was rejected as structurally impossible (component selected after planning).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1, RUNNING_NOTES_imagery_best_in_class.md#Turns-5–10, PLAN_imagery_best_in_class.md#I0-FINDING
- **relations:** image_source_unsatisfiable check is its guarantee for future domains; DeployedWebPath was fix 2 of the trio; corrupted templates were fix 3.
- **verify-later:** imageryplan package + test; plan_sections_action.go resolve()/ensureAssets/planSection.

<!-- SOURCE: U10_imagery.md -->
### DeployedWebPath committed-path convention (the two-URL serving model)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Verified: zero X-Amz URLs appear in any deployed page's HTML" (2026-07-10 warning box, corrected after a wasted debugging turn).
- **what:** Every generated image has two URLs: `assets.url` is a 7-day presigned S3 URL (SigV4 hard protocol ceiling — a throwaway source handle, never used to render) and the durable committed git path `/assets/images/<asset-key>.<ext>` derived by `storage.DeployedWebPath(asset_key, purpose)` — the single source of truth shared by deployer and resolver so they cannot drift. Pages serve via GitHub Actions → Backblaze B2 → a Cloudflare worker that re-signs each GET server-side. Debugging rule: get the real asset_key/purpose from the DB and curl the derived path; a presigned URL in assets.url is cosmetic staleness, not a broken image.
- **sources:** PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS, RUNNING_NOTES_imagery_best_in_class.md#Turn-8, SHOWCASE_technical_architecture.md#3
- **relations:** storage-architecture (worker.js, buckets); image-role resolver emits these paths; leopardess AUDIT_verified_facts D8 is the full write-up.
- **verify-later:** storage.DeployedWebPath/AssetKeyFilename; deploy_image_asset_action.go url-flip (~line 250).

<!-- SOURCE: U10_imagery.md -->
### flag_page_image_rebuild re-render trigger
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Delivered 2026-05-27 as the edited plan_sections_action.go and new flag_page_image_rebuild_action.go."
- **what:** Because needs_imagery runs after the page first renders, the fallback bakes into rendered_html and terminal rerenders reassemble without re-resolving. A terminal step in image-build-handler flags the page needs_rebuild and emits `needs_page` at priority 99 for page-scoped imagery, so the page re-resolves *through* plan_sections after its asset lands — closing the asset→render timing loop (the general asset→rerender coupling for site-level components remains an open item).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1-Decision, SHOWCASE_technical_architecture.md#3, TODO_imagery_followups.md#12
- **relations:** image-role resolver (same fix bundle); rerender-reassembles-not-resolves root cause 3.
- **verify-later:** flag_page_image_rebuild_action.go; image-build-handler terminal step.

<!-- SOURCE: U10_imagery.md -->
### image_source_unsatisfiable discovery check
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Registered 2026-07-09/10; "live but has produced 0 flags (heroes all resolve now) — expected."
- **what:** Flags component input_schema image fields sourced from a `site_assets.<path>` that no asset key, plan imagery row, or image-role alias can supply — the systematic guarantee that the empty-src class of failure is caught on every future domain instead of eyeballed. Flag-only (needs_human_review, no handler), dedup per site/page/function/path, cap 25/pass. Substituted for the structurally-impossible planner-side guard (component chosen after planning). Shares the alias table with the resolver so the two cannot drift.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-7, SQL_2026-07-09_register_image_source_unsatisfiable.sql
- **relations:** image-role resolver; services-hero orphan case (generated hero no component consumes).
- **verify-later:** check_image_source_unsatisfiable.go; design-discovery-agent run_checks.

<!-- SOURCE: U10_imagery.md -->
### Reference-image style anchoring
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Item 24 decision 2026-05-18 ("plumb reference-image AND switch icon model"); Banana native path live via style guide reference_asset_keys (I1, 2026-07-11); IP-Adapter/img2img paths not built.
- **what:** Conditioning generations on a reference image for sibling consistency. Three techniques ranked (img2img subject-anchor; IP-Adapter style-anchor, not on Stability's standard REST endpoint; LoRA highest); three reference-provenance options (generate-one-then-derive; per-site curated style library; system-wide per-kind house style). What shipped is the Banana-native form: approved reference assets resolved to stable s3:// URIs flow as style anchors for photographic kinds. Schema hooks (reference_image_uri field, origin_asset_id, alterations JSONB) exist ahead of the fuller paths.
- **sources:** TODO_imagery_followups.md#24, PLAN_imagery_best_in_class.md#D4, RUNNING_NOTES_imagery_best_in_class.md#Turn-17
- **relations:** imagery_style_guide; adoption mirror as a reference source; product-sketch img2img v2.
- **verify-later:** Banana provider reference handling; whether any non-Banana reference path exists.

<!-- SOURCE: U10_imagery.md -->
### Per-vertical LoRA fine-tunes
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Status: planned, not started… the fine-tuning plan presupposes an adapter that can take a model field; our adapter cannot" (assessment); best-in-class explicitly keeps it "as a future consistency upgrade once I1's reference-image approach shows its limits".
- **what:** Training per-vertical image LoRAs (60–90 curated images, SDXL/PixArt base, ~£35–95 first pass, per 018_canine_biology) for consistent visual style (vet diagrams, energy infographics). Blocked historically on model-selection plumbing; now deliberately deprioritised behind the cheaper reference-image approach.
- **sources:** FOCUS_imagery_assessment_1_.md#8, PLAN_imagery_loop_closure.md#Open-items, PLAN_imagery_best_in_class.md#6
- **relations:** canine-biology (018); model-infrastructure training; reference-image anchoring as the substitute.
- **verify-later:** none (not started).

<!-- SOURCE: U10_imagery.md -->
### Prompt-composition composer/envelope revisit
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Step 5 — image prompt cascade: Defer… see FOCUS_prompt_composition_pattern.md for why copying the text pattern is the wrong target" (resolved 2026-05-13).
- **what:** The image path keeps a single-prepend cascade rather than matching page-content-writer's richer text composition — a considered asymmetry: the text pattern itself is judged fragile and not worth copying. The strongest future candidate is a composer step producing a parameter envelope for both text and images, likely landing in a 2H-sibling phase. Partially realised since by the style-guide gating (which is composition-by-kind, not a full cascade).
- **sources:** PLAN_imagery_loop_closure.md#Decisions/#Image-prompt-cascade—deferred/#Open-items
- **relations:** FOCUS_prompt_composition_pattern.md (outside unit); imagery_style_guide; 2H request shape.
- **verify-later:** whether any composer/envelope step exists.

<!-- SOURCE: U10_imagery.md -->
### Components declare imagery contracts / many-images-per-page direction
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 3.5 "NEW for the refresh" but Phase 3 deferred; "Future direction — pages with many images (post-Phase-3)".
- **what:** content_components.input_schema v2 already supports typed image fields with arbitrary `source` resolvers, but only hero_image uses it. The direction: components own their imagery contracts (team-grid declares member_avatars arrays; services-grid declares per-service icons), the renderer resolves scoped site_plan_imagery rows by asset_key, discovery walks the declared gaps, and generation scales horizontally (a 30-image page is 30 work items through the unchanged chain). Enables per-image audit and retires silent fallthroughs to /assets/images/hero.jpg. The features `icon` string contract being one-sided (no renderer wiring) is the standing counter-example, resolved separately via Lucide.
- **sources:** PLAN_imagery_loop_closure.md#3.5/#Future-direction, FOCUS_imagery_assessment_1_.md#4.2/#9-item-5
- **relations:** Lane B; card imagery; contracts-and-standards (input_schema slot specs).
- **verify-later:** any component beyond hero_image with resolved image declarations.

<!-- SOURCE: U10_imagery.md -->
### Human taste-gate operating model (runbook rituals)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNBOOK structure in active use: standing rituals A1–A5 + one-off queue B1–B11, most B-items closed with dates; "Humans only at the taste layer" (showcase).
- **what:** The imagery workstream's division of labour: agents do all authoring/migrations/deploy-prep; humans do credentials, backups approval, budget sign-off, and visual approval gates — logo approval (once, then locked), sprite-sheet cell verification (assign true meanings after generation), and sampled page eyeballs per phase acceptance. Gates are deliberately the phases' biggest cost; generation is never trusted to self-judge taste.
- **sources:** RUNBOOK_imagery_best_in_class.md, SHOWCASE_imagery_workstream.md#Why-it's-interesting, SCOPE_I2_sprite_sheets.md#Phasing
- **relations:** logo permanence; sprite eyeball gate B11; hitl category (broader HITL machinery).
- **verify-later:** n/a (process concept); runbook item states.

<!-- SOURCE: U10_imagery.md -->
### Imagery work-item economy end-to-end chain
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** SHOWCASE technical architecture (verified against production 2026-07-10): full diagram planner → site_plan_imagery → needs_imagery → image-build-handler → provider adapter → store_asset → asset-deployer → flag_page_image_rebuild → resolver → rendered page; "~90 s prompt → git commit".
- **what:** The consolidated, operating imagery pipeline as a single nameable chain, including its dedup property (partial unique index lets checks re-run forever) and honest-state property (mark_item_failed). This is the umbrella concept the individual phase concepts compose into, and the shape any new imagery capability (sprites, cards, sketches) must ride.
- **sources:** SHOWCASE_technical_architecture.md#2/#3, SHOWCASE_one_pager.md, PLAN_imagery_best_in_class.md#2
- **relations:** every imagery concept above; development-guide work-item lifecycle.
- **verify-later:** end-to-end trace on a fresh needs_imagery item.

<!-- SOURCE: U10_imagery.md -->
### Rerender reassembles, it does not re-resolve
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** FOCUS §5.1 root cause 3 (2026-05-27); Turn 10 "NEW, SEPARATE ISSUE… rerender-completeness gap" — later narrowed: the specific 2026-07-10 instance was corrupted templates, but the underlying property stands ("needs_rerender and the colour/CSS fixers regex-patch stored rendered_html; they do not re-run plan_sections").
- **what:** The terminal rerender path reassembles existing rendered_html rather than re-running section resolution, so values that later land in content_data (resolved images, alt text) do not reach the HTML without a page rebuild through plan_sections. flag_page_image_rebuild routes page-scoped imagery around this; site-level components (header/footer) and non-hero inline sections remain exposed. Also noted: page_components.rendered_html is a snapshot, not a view — template changes don't reach deployed pages without a rebuild.
- **sources:** FOCUS_imagery_assessment_1_.md#5.1, RUNNING_NOTES_imagery_best_in_class.md#Turn-10/#Turn-11, HANDOFF_robot_hands_rebuild.md#Also-watch
- **relations:** flag_page_image_rebuild; corrupted templates (the misdiagnosis neighbour); styling-render-pipeline.
- **verify-later:** which paths re-run plan_sections vs regex-patch rendered_html.

<!-- SOURCE: U10_imagery.md -->
### Orphaned generated assets (component consumes nothing)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "services-hero template has no `<img>`; the generated hero_services.jpg is never consumed" (2026-07-09); full orphan check "deferred"; services coverage question open since item 21 (2026-05-15).
- **what:** Generation waste where the plan requests imagery no selected component can display (services hero), or assets outlive replans (stale sprite-sheet-main.jpg clutter; superseded May assets). Detection partially covered by undeployed_assets and image_source_unsatisfiable; a dedicated orphan/asset-supersession check and repo cleanup pass remain open, as does the planner-side question of which pages get heroes at all.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-5/#Turn-31, TODO_imagery_followups.md#21, PLAN_imagery_best_in_class.md#I0-finding-(c)
- **relations:** image_source_unsatisfiable (the flag side); planner key stability (another orphan source).
- **verify-later:** assets rows with no referencing rendered_html; site-repo files with no assets row.

## Dropped / superseded ideas ledger (quick index)
- `check_legacy_image_prompts_aspect` — designed, then retired as scoped (see concept above).
- `directory-builder` agent and a dedicated `needs_section_data` handler agent — decided in project docs, never built, superseded by queryresolve + reconciler design.
- Per-section imagery "deferred" (loop-closure v2 decision) — reversed 2026-05-12 by Phase 2G.
- Diffusion transparency for icons — abandoned (flat-grey chip).
- SDXL for flat icons — abandoned (Banana routing).
- Fix-A prompt-only icon rescue — tried, judged insufficient 2026-05-18.
- Planner-side guard "don't plan heroes for hero-less components" — structurally impossible (component chosen post-plan); substituted by image_source_unsatisfiable.
- Hand-rolled spawn+call trigger workflows — anti-pattern, replaced by work-item dispatch routing.
- `system.intake` trigger topic — stale/dead; kcat orchestrate envelope is the working path.
- pageflow-builder — retired from new-feature work, snapshot preserved.
- 3D product reconstruction — parked (sketches cover the need).
- fork_theme_from_site install mode — removed 2026-04-19; FK-swap is the sanctioned workaround.

<!-- SOURCE: U18_sql_for_agents.md -->
### image-generator + image prompt plumbing
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Contract in 029; Phase 0.1 (107) wires site_id through six call sites "so it can read design_intent.imagery_direction from site_specs".
- **what:** AI image generation specialist taking prompt/image_prompts (logo, hero_home) and producing image_url/image_data. Phase 0.1 made it site-aware: callers pass site_id so the generator composes design_intent.imagery_direction into prompts.
- **sources:** 029_image_generator.sql; 107_image_build_handler.sql (Section 1)
- **relations:** image-build-handler, site-work-orchestrator, pageflow-builder call it; asset-deployer deploys results
- **verify-later:** image generation adapter/action; imagery_direction composition code

<!-- SOURCE: U18_sql_for_agents.md -->
### image-build-handler + needs_imagery kind branch (Phase 2G)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 057 defines the handler; 107's "phase_2g_step5" section adds check_item_type_imagery / spawn_image_gen_imagery / call_imagery_gen / check_imagery_brand_update / store_imagery_brand_asset|store_imagery_asset steps ("teach image-build-handler to process needs_imagery work items (emitted by step 4's check_unfulfilled_imagery_plan)").
- **what:** Self-contained dispatch-loop handler for image work items: originally needs_logo/needs_hero_image (branch on spec.purpose → call image-generator → store_asset → deploy_image_asset via S3/optimize/git). Phase 2G extends it to generic needs_imagery items carrying kind-specific behaviour (icon transparency, logo variants), routed by item_type, with a spec.brand_update boolean deciding whether the stored asset also updates site brand assets.
- **sources:** 057_image_build_handler.sql; 107_image_build_handler.sql
- **relations:** build-dispatch-loop (caller); check_unfulfilled_imagery_plan discovery (imagery plan reconciliation); asset-deployer
- **verify-later:** live image-build-handler workflow steps; needs_imagery item emission in discovery checks

<!-- SOURCE: U18_sql_for_agents.md -->
### Imagery provenance: origin_model + origin_prompt on assets
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 107 combined migration Sections 2–3 (2026-05-05): "set origin_model literal on store_asset steps so the assets table records what produced each image"; origin_prompt_field normalised to record "the actual composed prompt sent to the model... not the un-composed plan prompt".
- **what:** Asset provenance discipline: every stored image records the generating model and the exact post-composition prompt. Required coordinated Go+SQL shipping (three concerns in one transaction) across image-build-handler, site-work-orchestrator, pageflow-builder.
- **sources:** 107_image_build_handler.sql (Sections 2–3 + backup preamble)
- **relations:** imagery audit work (file says provenance is "better for future iterations of the imagery audit work")
- **verify-later:** assets table columns origin_model/origin_prompt population

<!-- SOURCE: U19_sql_tables_components.md -->
### Asset key multi-image identity (Phase 2B–2D)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Staged migrations with live psql output pasted (11 rows backfilled; backup table assets_backup_20260508_pre_phase2d; ON_ERROR_STOP guard; old (site_id,purpose) unique index dropped).
- **what:** Replaces one-image-per-purpose with per-row asset_key: unique on (site_id, asset_key) WHERE active, enabling multiple images per logical purpose (e.g. adoption-mirror imports as 'adopted:<filename>'). Four-phase rollout: 2B add+backfill (asset_key=purpose), 2C StoreAssetAction writes asset_key and switches ON CONFLICT, 2D drops old purpose uniqueness after straggler sanity check.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2B and #Phase2D
- **relations:** assets provenance; site_plan_imagery key → namespaced asset_key.
- **verify-later:** idx_assets_site_asset_key_unique; StoreAssetAction ON CONFLICT target.

<!-- SOURCE: U19_sql_tables_components.md -->
### Assets table with full provenance
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 004 PART 6 creates assets with origin tracking; later phases (locks, asset_key) applied to a live table with 11 rows.
- **what:** All binary assets (image/video/document/logo/favicon) with storage location (provider/path/url), file metadata, and provenance: origin_type (generated/uploaded/scraped/stock/affiliate/derived), origin_url/prompt/model, origin_asset_id for derivations, alterations history JSONB, attribution/license. Purpose field ('hero', 'og_image'...) drives placement.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART6; docs/agent_docs/sql_for_tables/041_assets.sql
- **relations:** asset_key identity; image-build-handler work items; storage-architecture (providers).
- **verify-later:** StoreAssetAction; storage_provider values in use.

<!-- SOURCE: U19_sql_tables_components.md -->
### site_plan_imagery structured imagery plan
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "PURE DDL with NO BEHAVIOUR CHANGE. The table is empty until step 2 (write_site_plan extension) and step 3 (planner prompt extension)" with the 5-step Phase 2G sequencing listed (043 header).
- **what:** Sibling of site_plan_directives holding structured imagery requirements at site/page/section scope: key (asset_key stem, namespaced by the discovery check), kind enum via chk_kind (logo, hero, illustration, icon, infographic — product deliberately excluded, it comes from the affiliate_products resolver), required prompt, style_hints/constraints JSONB that cascade ADDITIVELY with directives' imagery_direction, ordering, source enum, and HITL locking with lock-transfer across plan rebuilds. chk_scope_ref_consistency enforces NULL / page_name / 'page:ordering' shapes; unique on (plan, scope, COALESCE(scope_ref,''), key).
- **sources:** docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql
- **relations:** site_plans domain; PLAN_imagery_phase_2g.md; check_unfulfilled_imagery_plan (step 4); image-build-handler (step 5).
- **verify-later:** table population; steps 2–5 delivery status.

<!-- SOURCE: U20_legacy_docs_a.md -->
### generate_image action + image-generator adapter pipeline
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** docs002/0100: "Image creation is now working" (deployed then); architecture: agent → `system.adapter.image-generator.requests` → Stability AI → Backblaze/S3 → reply_to_topic. Taxonomy names site_plan_imagery as the current pipeline.
- **what:** Image generation as a first-class workflow action: GenerateImageAction resolves prompts (template-rendered from CollectedData), sends to a shared adapter topic consumed by 3 load-balanced Python adapter replicas (consumer group), which call Stability AI, upload PNG to S3/Backblaze under `images/{client_id}/{date}/{image_id}`, and respond to the requesting agent's topic. Circuit breaker for API failures. A notable bug/fix: GenerateImageAction originally bypassed the image-generator *agent* and posted straight to the adapter — corrected so the agent orchestrates (parent → agent → adapter → agent → parent).
- **sources:** docs001_flow_general/README.095.a.image_handling.git.057_image.md; docs001_flow_general/README.095c.image_handling_topics.md; docs001_flow_general/README.097a.imagecreationandstorageflow.md; docs001_flow_general/README.096b.robothandswebsite.md
- **relations:** successor: docs024 imagery / site_plan_imagery pipeline; adapter microservice pattern; GPU strategy.
- **verify-later:** internal/adapters/imagegenerator; whether Stability AI config survives; current imagery pipeline tables.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Image storage and display URL strategy
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** README.099a recommends and implements dual URIs: "image_uri (s3:// for storage reference), image_url (https:// for web use)"; robot-hands pages embedded presigned URLs.
- **what:** Generated images return both an s3:// URI (storage reference) and a public HTTPS/CDN URL for embedding in HTML; options canvassed were public-bucket/CDN (chosen), presigned URLs (expiry problem for permanent sites), base64 embedding, and an image proxy service. Backblaze B2 public bucket setup documented.
- **sources:** docs001_flow_general/README.099a.image_storage_and_display_urls.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** storage-architecture (S3/B2 credentials); imagery pipeline.
- **verify-later:** ConvertS3URIToPublicURL or equivalent; current image URL scheme on live sites.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Asset & product provenance tables
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** docs012/010 new-table list: "assets — track all images/videos with full provenance; products; product_assets; affiliate_programs; affiliate_products"; docs017/002_pageflow stores hero assets in "assets table (existing)".
- **what:** All media (generated, uploaded, scraped) tracked with provenance in an assets table; product catalog and affiliate product caching designed alongside for e-commerce/review sites. The assets side shipped (used by image generation flow); products/affiliate tables remained design.
- **sources:** docs012_site_maps_and_components/010_component_and_site_architecture.md#New-Tables; docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md
- **relations:** image generation pipeline; entity-data (products as entities superseded product tables); link_registry affiliate fields.
- **verify-later:** assets table columns; products/affiliate_programs existence.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Image generation in the build pipeline
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** docs017/001_changes + 002_pageflow give exact patches ("generate_hero_image → store_hero_asset → deploy_hero_image (NEW) → templates use {{.hero_url}} → /assets/images/hero.jpg"); the current imagery system (site_plan_imagery, kind enums) replaced this.
- **what:** First-generation site imagery: image-generator agent produces logo/hero via adapter; store_generated_image/StoreAssetAction persists S3 URI into assets table and sites.content_data by purpose; deploy_image_asset downloads from S3, optimizes for web (resize per purpose), base64-commits via git-adapter; hero/logo URLs flow through render context into templates as background images.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md; docs017_legacy_agent_rules_images_design_keydocs/001_changes_needed.md
- **relations:** assets table; imagery pipeline (successor); image-optimiser fix agent.
- **verify-later:** deploy_image_asset action; storage/image_processing.go; current imagery pipeline contrast.

<!-- SOURCE: U22_recent_small_docs.md -->
### Image LoRA — scientific illustration style
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase F todo "Image LoRA fine-tuning (week 7-8)" unchecked; "SDXL recommended for diagrams" over FLUX.
- **what:** A plan to train an image LoRA (SDXL/PixArt preferred over FLUX for clean diagrams) on 60-90 curated, captioned veterinary/biological illustrations so the image-generator produces consistent anatomical cross-sections, pathway diagrams, procedure illustrations, and infographics across a site. Served via serverless (Replicate/RunPod) rather than an in-cluster GPU initially.
- **sources:** docs023.../018_canine_biology.md#7
- **relations:** image-generator adapter, canine biology KB, self-hosted inference
- **verify-later:** image-generator adapter LoRA support; any vet-diagram LoRA weights

<!-- SOURCE: U23_docs_root_vonc.md -->
### Resolver asset-kind surfacing gap (hero/logo only; illustrations unreachable)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Confirmed 2026-07-02: illustration assets EXIST (illustration_game_master, illustration_gauntlet_cta, purpose=illustration, active, files deployed) but "resolver ensureAssets only surfaces hero/logo, so site_assets.illustration can't reach them"; workaround applied (field made optional); extension still backlog item 4 on 2026-07-09.
- **what:** The plan_sections resolver's ensureAssets populates only `hero` and `logo` asset keys (from site_plan_imagery kinds hero/logo), so any schema field sourced `site_assets.illustration` can never resolve even when illustration assets exist — deferring sections. Interim: make such fields optional (text-only render). Structural options: extend ensureAssets to surface kind=illustration from site_plan_imagery+assets (benefits all sites), or per-field fallback URLs. Related mismatch: gauntlet-cta has no illustration field despite an illustration_gauntlet_cta asset existing.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:50 + #2026-07-03-~13:18; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** plan_sections deferral; site_plan_imagery (imagery pipeline)
- **verify-later:** plan_sections resolver ensureAssets; site_plan_imagery kinds in use

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Data-graph / charts future pipeline
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** TODO_imagery_followups(15) "Future workstream: data graphs / charts (added 2026-05-20)"; HANDOFF news-enhancement (b) ties news→infographic to "the future data-graph pipeline (FUTURE_data_graph_pipeline.md) if the infographic is data/chart-driven"; PLAN_imagery_phase_2g: `kind=infographic` "renders as illustration for now."
- **what:** A planned separate generator producing SVG/HTML charts from structured data (distinct from raster image generation), the eventual real backing for `kind=infographic` and the news→infographic enhancement. No scaffolding exists; closest analogues are the dynamic-tool components.
- **sources:** imagery/old/TODO_imagery_followups(15).md#future-workstream-data-graphs; imagery/old/PLAN_imagery_phase_2g(1).md#what-2g-doesnt-include
- **relations:** infographic kind; news→infographic; many-images-per-page
- **verify-later:** FUTURE_data_graph_pipeline.md (live); infographic-generator agent (absent)

<!-- SOURCE: U25_leopardess_social.md -->
### Two-URL asset model (throwaway presigned handle vs durable git path)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** AUDIT D8 correction (2026-07-10): "zero X-Amz presigned URLs appear in … rendered pages"; "the imagery pipeline works. There is no platform-wide emergency."
- **what:** assets.url stores a 7-day presigned S3 URL that is only a source handle — SigV4 caps expiry at 604800s so a permanent presigned URL is impossible. Render never reads it: plan_sections emits storage.DeployedWebPath (the git-committed /assets/images/<key>.<ext> path), served via GitHub Actions → Backblaze B2 → a Cloudflare Worker that re-signs B2 GETs server-side. 83 stale presigned rows are cosmetic; the w9_04 url-flip backfill (done for idea.uk) is the optional cleanup. Includes the D8 self-correction on record: an earlier "asset deploy is broken platform-wide" alarm was withdrawn after re-investigation.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#D8; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; docs/leopardessconsulting/RUNBOOK.md#landmine-6/7
- **relations:** spawn-time storage env injection; storage-architecture (B2, Cloudflare worker)
- **verify-later:** plan_sections_action.go:193/260/290; scripts/cloudflare/worker.js; assets.url vs storage_path across sites

<!-- SOURCE: U25_leopardess_social.md -->
### Spawn-time storage env injection (base chassis carries no IMAGE_BUCKET by design)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** AUDIT D8 finding 1: "spawn_actions.go injects IMAGE_BUCKET + storage creds into spawned storage-enabled agents … 107_image_build_handler.sql:725 documents exactly this."
- **what:** The base agent-chassis pod deliberately lacks IMAGE_BUCKET; deploy_image_asset therefore fails when run inline, but the real pipeline spawns asset-deployer with storage env injected (isStorageEnabledAgent). Hand-triggering a storage agent standalone reproduces a spurious "Storage client not configured" failure. Documented as a "do not add IMAGE_BUCKET to agent-chassis" warning.
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#D8; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-10; docs/leopardessconsulting/scripts/commit_brand_assets.sh (header records the earlier, since-corrected reading)
- **relations:** two-URL asset model; system-architecture (agent spawning)
- **verify-later:** platform/orchestration/actions/spawn_actions.go isStorageEnabledAgent; agentbase/agent.go:294

<!-- SOURCE: U25_leopardess_social.md -->
### Imagery kind→provider routing and reference-image support (A6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** HANDOFF §2 A6: "Committed in dynamic_adapter.go; NOT deployed to cluster (needs a Makefile build-from-local-filesystem)".
- **what:** dynamic_adapter routes kind=="icon" to Banana (Gemini gemini-3-pro-image-preview) and everything else to Stability SDXL. Only Banana honours reference images — the mechanism brand consistency depends on — so A6 extends the Banana route to logo/illustration/infographic, leaving hero/photographic on SDXL. Imagery kinds are constrained twice (chk_kind and validImageryKinds in write_site_plan_action.go): changing the set needs migration + Go edit together.
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4, #Turn-7; docs/leopardessconsulting/RUNBOOK.md#O5, #landmine-5; docs/leopardessconsulting/PLAN_leopardess_rebuild.md#L2
- **relations:** logo candidate generation; brand-asset derivation gap; chart concept (charts deliberately NOT an imagery kind)
- **verify-later:** internal/adapters/imagegenerator/dynamic_adapter.go routing on the running pod vs repo

<!-- SOURCE: U25_leopardess_social.md -->
### Brand-asset derivation gap and direct git-adapter brand commit path
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** REPLICATION summary table: favicon + OG-card derivation "[gap] — small; currently manual"; RUNBOOK H4 "resolved 2026-07-10 … deployed live … all verified 200 and byte-identical".
- **what:** Favicon and OG-card generation from a logo is not a pipeline capability (no favicon/OG generator in the codebase); they were manually derived (background knockout to transparency, gold normalisation to brand hex, flood-filled silhouette favicon for 16/32px, multi-size .ico, opaque apple-touch icon, 1200×630 OG card). Delivery used commit_brand_assets.sh: send the same git-adapter commit message deploy_image_asset would send, for pre-approved images where there is no generation step to spawn. deploy_brand_asset.sh notes asset-deployer never rewrites assets.url without asset_id (landmine 6).
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-9; docs/leopardessconsulting/RUNBOOK.md#O7; docs/leopardessconsulting/scripts/commit_brand_assets.sh, deploy_brand_asset.sh (headers)
- **relations:** imagery kind routing (A6); two-URL asset model
- **verify-later:** internal/adapters/git/github_client.go CommitToRepo path prefixing; absence of favicon/OG generator

<!-- SOURCE: U25_leopardess_social.md -->
### Logo candidate generation with small-size survival test and human approval
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES turn 8–9: six candidates generated 2026-07-10, owner chose c2, exact approved PNG deployed live and byte-verified.
- **what:** Brand-mark commissioning practice: generate N candidates through the same model/key the chassis uses, save all prompts for pipeline reproduction, judge by small-size survival (a favicon is 16px — solid-fill marks survive, line-art dissolves), and require human approval for a for-the-life-of-the-site decision (maps to the platform's checkpoint_for_review HITL surface). Key insight: the owner approves a specific image, not a prompt — regenerating "the same" prompt yields a mark they never saw.
- **sources:** docs/leopardessconsulting/logo_candidates/PROMPTS.md; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8, #Turn-9; docs/leopardessconsulting/REPLICATION_in_chassis.md#4
- **relations:** hitl (checkpoint_for_review); imagery kind routing (A6)
- **verify-later:** site_plan_imagery logo row for leopardess; checkpoint_for_review surface

<!-- SOURCE: U25_leopardess_social.md -->
### Illustration/section-imagery asset resolution (kind-alias path)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-11: "the section's render simply predated the section-imagery resolver … schema source site_assets.illustration resolves via site_plan_imagery kind-alias … src filled"; residual: undeployed_assets "infers deployment from rendered_html usage, not from the repo".
- **what:** How section images resolve: a schema field sourced site_assets.<type> resolves through site_plan_imagery kind-alias rows to a deployed asset path; stale renders predating the resolver ship `<img src="">` and are fixed by a light page_rerender (reason image_landed, no LLM). Known gaps: ensureAssets surfaces only hero/logo (named illustrations exist as assets but weren't reachable until aliased); the undeployed_assets check flags committed-but-unreferenced assets (usage-inferred, not repo-inferred). The archetype hub consumed 8 orphaned icons via page-scope hero alias rows.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11, #2026-07-12; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.5a; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (outstanding illustration)
- **relations:** two-URL asset model; plan_sections deferral; archetype hub build
- **verify-later:** ensureAssets; site_plan_imagery kind-alias resolution code; undeployed_assets check logic

---

## Notes for consolidation

- **Family deltas of substance:** the 003 family shows RAG integration and the automated strategic-review agent designed then consciously deferred (extracted above); the 002 spark family shows continuous accretion (Arena/Stage split arrives in 002b, motivation tiers 002c, user journey + cost tables 002d, launch strategy 002e) with only minor drops ("Retention Bridges"/"Crews"/"Workshop Mode" headings folded into later sections, not abandoned). Tool-docs NOTES families are append-only logs; earlier versions contain no concepts absent from the latest.
- **Cross-unit duplicates expected:** verify-by-artifact discipline, the apply_section_edit/approved defect, imagery two-URL model, and 016b debugging entries will also surface from docs024/docs019 units — merge on consolidation; this unit contributes dated deployment evidence.
- **Proposed NEW categories:** `NEW:data-charts` (Go SVG emitter + JS enhancement, D1/D3 doctrine — could back a council agent distinct from diffusion imagery), `NEW:operator-practice` (verify-by-artifact, backups, kcat triggers, in-chassis replicability — cross-cutting discipline no seed category owns), `NEW:rag-knowledge-base` (rag_index/rag_lookup/knowledge_base — referenced as existing infrastructure with no seed home).

<!-- SOURCE: U25_leopardess_social.md -->
### Chart component: Go static-SVG emitter + JS progressive enhancement (A5/A7)
- **category:** NEW:data-charts
- **status-signal:** aspirational
- **status-evidence:** PLAN phase table: "L7 Charts (Go SVG + JS) — not started"; H2 "resolved 2026-07-10 — confirmed, data layer first".
- **what:** The reusable data-chart capability honouring prior imagery decisions: D1 (charts are code-rendered from real data — the LLM proposes the story, code owns the numbers; diffusion never plots data), D3 (chart is a Lane-B asset, deliberately NOT a site_plan_imagery kind). Resolves the recorded conflict between "go-echarts in-chassis" (confirmed 2026-07-08) and "static SVG must always exist": a dependency-free Go SVG emitter produces the accessible static artifact (axes, caption, source line, query date); an inline self-contained JS renderer progressively enhances it (no CDN). First charts use provable in-DB numbers. Explicitly excludes the data-chart-generator agent and external data APIs (deferred to imagery Phase I4).
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#5; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4; docs/leopardessconsulting/RUNBOOK.md#H2
- **relations:** imagery programme I4 / FUTURE_data_graph_pipeline (docs024); imagery kind routing (charts kept out of diffusion)
- **verify-later:** go.mod for charting deps (none expected); any SVG emitter under the chassis

<!-- SOURCE: U25_leopardess_social.md -->
### Chart component: Go static-SVG emitter + JS progressive enhancement (A5/A7)
- **category:** NEW:data-charts
- **status-signal:** aspirational
- **status-evidence:** PLAN phase table: "L7 Charts (Go SVG + JS) — not started"; H2 "resolved 2026-07-10 — confirmed, data layer first".
- **what:** The reusable data-chart capability honouring prior imagery decisions: D1 (charts are code-rendered from real data — the LLM proposes the story, code owns the numbers; diffusion never plots data), D3 (chart is a Lane-B asset, deliberately NOT a site_plan_imagery kind). Resolves the recorded conflict between "go-echarts in-chassis" (confirmed 2026-07-08) and "static SVG must always exist": a dependency-free Go SVG emitter produces the accessible static artifact (axes, caption, source line, query date); an inline self-contained JS renderer progressively enhances it (no CDN). First charts use provable in-DB numbers. Explicitly excludes the data-chart-generator agent and external data APIs (deferred to imagery Phase I4).
- **sources:** docs/leopardessconsulting/PLAN_leopardess_rebuild.md#5; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4; docs/leopardessconsulting/RUNBOOK.md#H2
- **relations:** imagery programme I4 / FUTURE_data_graph_pipeline (docs024); imagery kind routing (charts kept out of diffusion)
- **verify-later:** go.mod for charting deps (none expected); any SVG emitter under the chassis

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Component asset coupling not enforced (external JS/data file existence is convention only)
- **category:** NEW:component-asset-pipeline
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "data_sources enforcement + inline-small-js_content — component data-file paths are convention, not enforced" (open, unresolved as of 2026-05-21)
- **what:** A component template can reference `<script src="/tools/assets/X.js">` or fetch `/data/X.json` with nothing in the pipeline guaranteeing those files exist or get produced. `content_components.data_sources` (text[]) exists to declare the dependency but isn't consistently populated or validated at deploy time. Two proposed fixes: inline js_content <5KB directly vs. enforce/auto-stub for larger payloads; same pattern applies to data files.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Known-gaps, js_snippets_news_gaswholesalers/old/component_asset_pipeline_concerns.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** News rendering three-layer architecture; two news components pattern
- **verify-later:** content_components.data_sources column, git tools/assets/ and data/ paths per site

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Component asset coupling not enforced (external JS/data file existence is convention only)
- **category:** NEW:component-asset-pipeline
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "data_sources enforcement + inline-small-js_content — component data-file paths are convention, not enforced" (open, unresolved as of 2026-05-21)
- **what:** A component template can reference `<script src="/tools/assets/X.js">` or fetch `/data/X.json` with nothing in the pipeline guaranteeing those files exist or get produced. `content_components.data_sources` (text[]) exists to declare the dependency but isn't consistently populated or validated at deploy time. Two proposed fixes: inline js_content <5KB directly vs. enforce/auto-stub for larger payloads; same pattern applies to data files.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Known-gaps, js_snippets_news_gaswholesalers/old/component_asset_pipeline_concerns.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md
- **relations:** News rendering three-layer architecture; two news components pattern
- **verify-later:** content_components.data_sources column, git tools/assets/ and data/ paths per site

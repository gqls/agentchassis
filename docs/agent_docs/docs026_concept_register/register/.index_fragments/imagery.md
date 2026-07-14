| IMG-001 | Imagery loop-closure master plan (Phases 0–6) | partial | Sequenced plan closing spec-to-delivery imagery gaps; Phase 2G/2H verified, 3-6 pending. | imagery.md |
| IMG-002 | Imagery subsystem pre-plan assessment | superseded | As-is baseline (hardcoded Stability adapter, one-purpose assets) that motivated the loop-closure plan. | imagery.md |
| IMG-003 | Imagery best-in-class programme (I0–I8) | partial | 2026-07-08 successor programme; I0/I1 complete, I2 in progress, I3-I8 not started. | imagery.md |
| IMG-004 | Two lanes of imagery: plan-driven vs content-driven (Lane B) | aspirational | Design split between fixed plan-time imagery and continuous content-attached imagery. | imagery.md |
| IMG-005 | imagery_direction prepend + origin_model/origin_prompt provenance | deployed | Prepends site style direction to prompts; fixed provenance columns being silently dropped. | imagery.md |
| IMG-006 | Planner ignores site_archetype imagery constraints (Bug 4) | unknown | Planner produced lavish hero prompts despite archetype saying minimal/no photography. | imagery.md |
| IMG-007 | Algorithmic imagery discovery checks (Phase 1 trio) | deployed | No-LLM checks: unfulfilled_image_prompt, placeholder_image_in_use, image_url_404. | imagery.md |
| IMG-008 | Asset locking mirrors page_components (Phase 2A) | deployed | assets gains locked_at/lock_type so audits/discovery skip locked (e.g. approved) assets. | imagery.md |
| IMG-009 | asset_key multi-image model (Phase 2B–2D) | deployed | Broke one-asset-per-purpose constraint; enables N heroes/icons per site via asset_key. | imagery.md |
| IMG-010 | Hero-variant routing through image-build-handler (Phase 2E) | deployed | Made hero_<page> variants routable via a new classification and deploy-path branch. | imagery.md |
| IMG-011 | Spawned asset-deployer / storage-env isolation (Phase 2F) | deployed | Base chassis pod carries no IMAGE_BUCKET by design; deploys spawn asset-deployer instead. | imagery.md |
| IMG-012 | site_plan_imagery structured plan table (Phase 2G.1) | deployed | Structured per-image plan rows (scope, kind, prompt) succeeding legacy image_prompts dict. | imagery.md |
| IMG-013 | flattenImageryBlock write path + lock transfer (Phase 2G.2) | deployed | write_site_plan inserts site_plan_imagery rows and transfers HITL locks across replans. | imagery.md |
| IMG-014 | Planner imagery block prompt extension (Phase 2G.3) | deployed | Planner JSON output gains a structured imagery key in the same LLM call as pages. | imagery.md |
| IMG-015 | check_unfulfilled_imagery_plan discovery check (Phase 2G.4) | deployed | Emits needs_imagery per unfulfilled site_plan_imagery row, priority-banded by scope. | imagery.md |
| IMG-016 | needs_imagery branch in image-build-handler (Phase 2G.5) | deployed | New generic branch handling structured needs_imagery items alongside legacy branches. | imagery.md |
| IMG-017 | Legacy image_prompts age-out check — retired (Phase 2G.6) | superseded | Dedicated migration check retired; reframed as operational deregistration decision. | imagery.md |
| IMG-018 | Image-generator request shape + per-kind defaults (Phase 2H) | deployed | Extends request with negative_prompt/seed/reference_image_uri and Go kindDefaults map. | imagery.md |
| IMG-019 | parseAspectRatio SDXL v1.0 whitelist snap fix | deployed | Fixed aspect-ratio snapping to SDXL's strict dimension whitelist, unblocking hero gen. | imagery.md |
| IMG-020 | Build-time imagery trigger: emit_imagery_items + imageryplan.go | deployed | Plan-time emission of needs_imagery sharing selection logic with the discovery check. | imagery.md |
| IMG-021 | Adoption image mirror (Phase 3) | aspirational | Would persist crawled imagery as assets instead of discarding it after adoption. | imagery.md |
| IMG-022 | Visual auditor imagery awareness (Phase 4, text-only) | aspirational | Would give visual-design-auditor a sixth IMAGERY check category from text context only. | imagery.md |
| IMG-023 | Vision-capable LLM path (Phase 5) | aspirational | Foundational aiservice image-input support required before any imagery auditor can see images. | imagery.md |
| IMG-024 | imagery-quality-auditor agent (Phase 6 / I8) | aspirational | Planned vision-capable auditor sibling dedicated to imagery direction/brand/quality checks. | imagery.md |
| IMG-025 | Image generation as parameter shaping (deferred composer step) | aspirational | Composition-by-parameter design for images deferred; partially realised by per-kind gating. | imagery.md |
| IMG-026 | Icon generation lessons and image-model comparison | deployed | SDXL judged wrong for flat icons; model comparison drove switch to Banana for icons. | imagery.md |
| IMG-027 | LLM-generated SVG icon path (sleeper option) | aspirational | LLM-written SVG icons retained as a possible future replacement for raster icon gen. | imagery.md |
| IMG-028 | Diffusion transparency abandoned → flat-grey chip icons | deployed | Models can't produce real alpha; icons locked to flat grey background inside a CSS chip. | imagery.md |
| IMG-029 | Lucide icon strategy and validator wiring | partial | Features grid uses Lucide webfont icons; validator written but not yet wired in. | imagery.md |
| IMG-030 | Image provider abstraction and kind→provider routing | partial | Stability/Banana kind-based routing; A6 extension committed but not yet deployed. | imagery.md |
| IMG-031 | Per-kind prompt gating and five-place new-kind checklist | deployed | Photographic brand direction gated off icons/logos; standing checklist for new kinds. | imagery.md |
| IMG-032 | One-entry-one-image decomposition rule (planner prompt patch) | deployed | Planner prompt bans multi-image prompts like "a set of six icons". | imagery.md |
| IMG-033 | Planner key stability across replans | aspirational | Planner freely renames imagery keys across replans, causing spurious regenerations. | imagery.md |
| IMG-034 | generate_image action + image-generator adapter pipeline (legacy) | superseded | First-generation Stability-only image workflow, superseded by site_plan_imagery pipeline. | imagery.md |
| IMG-035 | Image storage and display URL strategy (S3/B2 dual URI) | deployed | Early dual s3://+https:// URI decision; public-bucket/CDN chosen over presigned URLs. | imagery.md |
| IMG-036 | pageflow-builder retirement | superseded | Legacy monolithic site builder deliberately left un-extended; architecture moved on. | imagery.md |
| IMG-037 | Assets table with full provenance (schema) | deployed | All-binary-assets table with origin_type/prompt/model/asset_key; products side stalled. | imagery.md |
| IMG-038 | imagery_style_guide — per-site brand guide as data (Phase I1) | deployed | Per-site palette/medium/mood/avoid/reference guide driving generation with per-kind gating. | imagery.md |
| IMG-039 | Logo permanence: generate → human-approve → lock (D5) | deployed | Logo generated once, human-approved, then locked; favicon/OG derive from it, never regen. | imagery.md |
| IMG-040 | Brand-head derived assets (favicon + OG card) | deployed | derive_brand_head_assets deterministically derives favicon/OG card from locked logo. | imagery.md |
| IMG-041 | Manual brand-asset commit workaround (derivation gap) | partial | Leopardess site hand-derived favicon/OG and committed via a standalone shell script. | imagery.md |
| IMG-042 | Header logo resolution from plan imagery | deployed | Header component fixed to resolve locked logo via site_plan_imagery, not dead sites.logo_url. | imagery.md |
| IMG-043 | Sprite-sheet bullets and list treatment (Phase I2) | partial | One N×M glyph grid per site sliced by CSS background-position; active build phase. | imagery.md |
| IMG-044 | Content-linked card imagery (Phase I3) | aspirational | Would give every linking card an image re-cropped from its article's own generated asset. | imagery.md |
| IMG-045 | News imagery (Phase I5) | aspirational | Per-news-item chart/illustration/none classification with a grace-interval fallback image. | imagery.md |
| IMG-046 | Data-graph / chart pipeline (code-rendered, never diffusion) (I4) | aspirational | Charts must be code-rendered from real data, never diffusion-generated; not built yet. | imagery.md |
| IMG-047 | Product illustration pipeline (copyright-safe sketches) (I6) | aspirational | Stylised product sketches to avoid trade-dress exposure from scraped affiliate photos. | imagery.md |
| IMG-048 | Image performance budgets (Phase I7 / D8) | aspirational | Per-kind byte/dimension ceilings enforced at deploy with a weight-over-budget check. | imagery.md |
| IMG-049 | Reference-image style anchoring | partial | Banana-native reference-image anchoring shipped via style guide; IP-Adapter/LoRA not built. | imagery.md |
| IMG-050 | Per-vertical LoRA fine-tunes | aspirational | Planned per-vertical image LoRA training, deprioritised behind reference-image approach. | imagery.md |
| IMG-051 | Per-page hero resolver + flag_page_image_rebuild trigger (June fix) | deployed | Fixed site-wide hero overwrite + baked fallback + non-reresolving rerender, page scope. | imagery.md |
| IMG-052 | Legacy site-level hero_url shadow (content_data last-write-wins) | deployed | Legacy content_data hero_url still shadows per-page heroes with one site-wide image. | imagery.md |
| IMG-053 | Presigned-URL expiry and deploy-time asset localisation (Edit F) | deployed | assets.url presigned links died after 7 days; deploy now records the durable local path. | imagery.md |
| IMG-054 | DeployedWebPath committed-path convention (two-URL serving model) | deployed | Every asset has a throwaway presigned URL and a durable committed git path; render uses latter. | imagery.md |
| IMG-055 | Section-scope imagery pipeline — idea.uk verification | deployed | End-to-end plan→emit→generate→deploy→rebuild chain exercised live on idea.uk. | imagery.md |
| IMG-056 | ensureAssets scope gap: hero/logo-only surfacing (Edit B / kind-alias) | deployed | ensureAssets only surfaced hero/logo; extended to section/illustration scope across 3 sites. | imagery.md |
| IMG-057 | flag_page_image_rebuild section-scope mapping (Edit H) | partial | Rebuild flag no-op'd for section scope; prefix-split fix pending code apply. | imagery.md |
| IMG-058 | Image-role alias resolver + authoritative overlay (I0 rewrite) | deployed | July rewrite unifying 3 incompatible hero-resolution patterns via a shared alias table. | imagery.md |
| IMG-059 | image_source_unsatisfiable discovery check | deployed | Flags image fields sourced from an asset path nothing can supply; 0 flags = healthy. | imagery.md |
| IMG-060 | Rerender reassembles, it does not re-resolve | partial | Terminal rerender patches stored HTML rather than re-running section resolution. | imagery.md |
| IMG-061 | Orphaned generated assets (component consumes nothing) | partial | Generated imagery with no consuming component slot, or stale post-replan assets. | imagery.md |
| IMG-062 | Components declare imagery contracts / many-images-per-page | aspirational | Direction for components to own typed imagery contracts instead of hero_image-only. | imagery.md |
| IMG-063 | Human taste-gate operating model (runbook rituals) | deployed | Agents author/deploy; humans hold credentials, budget sign-off, and visual approval gates. | imagery.md |
| IMG-064 | Imagery work-item economy end-to-end chain | deployed | The umbrella planner→build→deploy→rebuild chain every imagery phase concept composes into. | imagery.md |
| CHRT-001 | Chart component: Go static-SVG emitter + JS progressive enhancement | aspirational | Code-rendered chart plan: dependency-free Go SVG emitter plus inline JS enhancement, no CDN. | data-charts.md |
| CAP-001 | Component asset coupling not enforced (JS/data file existence) | aspirational | Component templates reference external JS/data files with no pipeline guarantee they exist. | component-asset-pipeline.md |

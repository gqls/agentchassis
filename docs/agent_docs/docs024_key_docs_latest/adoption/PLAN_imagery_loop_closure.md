# FOCUS / PLAN — Imagery: assessment and loop closure

This document combines the imagery **assessment** (where the system is now —
formerly `FOCUS_imagery_assessment.md`) with the **sequenced plan** to close the
spec-to-delivery gap (formerly `PLAN_imagery_loop_closure.md`). Part 1 is the
standing description of the imagery subsystem and its structural gaps; Part 2 is
the phased execution plan (phases 0–1 done, 2+ pending) that acts on them.

---

## Part 1 — Current state (assessment)

_Descriptive: how images are produced, stored, and used today, and the structural gaps. The sequenced plan in Part 2 acts on these gaps._

This is an assessment of how images are currently produced, stored, and used in the build pipeline, and where the gaps are relative to the broader goal (varied imagery — decorative SVG, icons, infographics, illustrations, product approximations — placed intelligently across components, possibly produced by different models for different tasks).

It is descriptive, not prescriptive. Concrete next steps are flagged at the end as candidates for a follow-up plan.

---

## 1. Generation layer

### 1.1 The single adapter

| Item | Value |
|---|---|
| File | `internal/adapters/imagegenerator/dynamic_adapter.go` |
| Type | `DynamicImageAdapter` |
| Topic | `system.adapter.image-generator.requests` |
| Provider | Stability AI (hardcoded) |
| Model | Stable Diffusion XL 1024 (whatever `STABILITY_API_ENDPOINT` resolves to) |
| API key | `STABILITY_API_KEY` env var |
| Storage | Backblaze B2 — bucket `personae-prod-uk001-images` |

The request body is hardcoded in `generateImage()`:

```go
requestBody := map[string]interface{}{
    "text_prompts":         []map[string]interface{}{{"text": prompt, "weight": 1}},
    "cfg_scale":            7,
    "clip_guidance_preset": "FAST_BLUE",
    "height":               height,
    "width":                width,
    "samples":              1,
    "steps":                30,
}
```

Notable gaps in the request:
- No negative prompt
- No style preset (`enhance`, `anime`, `digital-art`, `line-art` etc. that Stability supports)
- No seed
- No reference image or img2img mode
- No LoRA or fine-tune selector
- `samples: 1` — we can't ask for variants in one call

### 1.2 The two `image-adapter` config files are misleading

`configs/image-adapter.yaml` and `agent_definitions` row `1db46661-...` both contain `image_api.url`, `defaults`, and `circuit_breaker` blocks — none of these are actually consumed by the adapter at runtime. The adapter reads `STABILITY_API_ENDPOINT` and `STABILITY_API_KEY` directly from the environment. That's a structural smell — config that looks load-bearing but isn't.

### 1.3 Two `image-generator` agent rows in the DB

| `id` | `is_active` | Notes |
|---|---|---|
| `1db46661-2811-4aca-8a59-5212e36d3b88` | true | Current — adapter category, has env_vars block |
| `59ca0d97-577f-45f8-963a-6c7ba2754da4` | false | Old version, kept inactive |

The active one has `processing_mode: adapter`, but the orchestrator-side code paths (the `generate_image` action in `platform/orchestration/actions/`) still build their own Kafka message and send to `system.adapter.image-generator.requests`. The agent definition is essentially a placeholder — the action does the work.

---

## 2. Action layer

### 2.1 `generate_image` — sends to the adapter

`GenerateImageAction` (in `platform/orchestration/actions/generate_image_actions.go` per the context dumps) builds a request to `system.adapter.image-generator.requests` and uses three-tier prompt priority:

1. Parent's `call_agent` config / `CollectedData["prompt"]` / `input_data.prompt`
2. Agent's `default_config.prompt_template`
3. Workflow step `prompt_template`

The action only ever passes `prompt`, `style`, `width`, `height`. There's no path through this action for: model selection, provider selection, negative prompt, reference images, LoRA, batch/variants, content safety mode, or aspect ratio presets.

### 2.2 `store_asset` — writes to DB

Inserts into `assets` with `asset_type`, `purpose`, `url`, `origin_type`, `origin_prompt`. Also writes `{purpose}_uri` and `{purpose}_url` keys into `sites.content_data` so that downstream renderers can find them.

### 2.3 `deploy_image_asset` — downloads, optimizes, commits

Lives in the `asset-deployer` agent. Downloads from S3, runs `OptimizeImageForWeb` (resize + reencode), then sends a base64-encoded git commit request. Output path follows `BuildAssetPaths(purpose, ext)` — `assets/images/{purpose}.{ext}`, e.g. `assets/images/hero.jpg`, `assets/images/logo.png`.

This is where `ImagePurposes` (in `platform/storage/...`) is consulted:

```go
"hero":          {1600, 900, 85, "jpg"},
"hero_home":     {1600, 900, 85, "jpg"},
"hero_about":    {1600, 900, 85, "jpg"},
"hero_services": {1600, 900, 85, "jpg"},
"logo":          {400, 400, 90, "png"},
"icon":          {240, 240, 85, "jpg"},
"default":       {1200, 800, 85, "jpg"},
```

`icon` is declared but no path in the system actually requests one. There are no entries for: portrait, product, illustration, infographic, decorative-svg, og-card, favicon, etc.

---

## 3. Database

### 3.1 `assets` table

The schema is reasonable:

```
asset_type, purpose, name, url, storage_provider, storage_path, filename,
mime_type, file_size, dimensions (jsonb), duration, origin_type, origin_url,
origin_prompt, origin_model, origin_asset_id, alterations (jsonb),
attribution, license, license_url, status
```

Two important details:

1. **Unique constraint on `(site_id, purpose)`** — `idx_assets_site_purpose_unique` — means **one asset per purpose per site**. This blocks "ten different decorative illustrations" or "five product mockups" unless we encode an index into the purpose string. That has to change before we can support multi-image purposes.
2. `origin_model` exists but the current `store_asset` action doesn't populate it. We're losing the provenance when we eventually have multiple models.
3. `origin_asset_id` exists for derivatives — useful for "alteration of asset X" — also unused.
4. `alterations jsonb` is unused. This is where img2img variants, style transfers, and 3D-from-photos would naturally record their provenance chain.

### 3.2 `product_assets` and `affiliate_products.custom_image_id`

Both FK to `assets`. Wired but no pipeline produces or consumes them. Relevant to the user's "products without copyright-free images" requirement — there's already a place to put a generated product approximation, we just don't generate one.

### 3.3 `site_specs.aspect = 'design_intent'`

Webdesign-agent and the adoption pipeline write a `design_intent` aspect with an `imagery_direction` field (free-text describing what the imagery should convey). **No image action currently reads this.** The webdesign-agent expresses imagery taste, image-generator ignores it.

---

## 4. Planning layer

### 4.1 What the planner asks for

`site-planner` is prompted to produce `image_prompts` with exactly two keys:

```json
"image_prompts": {
  "logo": "...",
  "hero_home": "..."
}
```

The prompt template explicitly says these two keys are required and gives no opening for additional keys. There is no:
- per-page hero prompt (despite `hero_about_url`, `hero_services_url` existing in the render context)
- per-section illustration prompt
- icon set request
- decorative-graphic request
- infographic spec
- product imagery list

Some adoption-pipeline plans do produce richer `image_prompts` objects (the leopardess example showed `about_hero`, `case_studies_hero`, `contact_hero`, `services_hero` keys), but none of those flow through to actually being generated — only `logo` and `hero_home` are consumed downstream.

### 4.2 What components ask for

`content_components.input_schema` v2 supports an `image` field type with a `source: "site_assets.{type}"` resolver:

```json
"hero_image": {
  "type": "image",
  "source": "site_assets.hero",
  "required": false,
  "on_missing": "use_fallback",
  "fallback": "/assets/images/hero.jpg"
}
```

So the schema is ready for components to declare `site_assets.team_member_3` or `site_assets.diagram_pricing` — but the planner doesn't list those, the renderer doesn't resolve them, and no agent generates them.

Features sections have an `icon` field in the LLM JSON output (a string like `"shield"` or `"chart"`). At render time these strings go nowhere — there's no SVG library, no icon component, no resolver. The icon is documented as an LLM output but the rendering side is not wired.

---

## 5. Render integration

`BuildRenderContextAction` in `actions/build_render_context_action.go` looks for image URLs in three places, in order:

1. `hero_deployed.image_url`, `logo_deployed.image_url` (fresh from this build)
2. `data["hero_url"]`, `data["hero_home_url"]`, `data["hero_about_url"]`, `data["hero_services_url"]`, `data["logo_url"]` (from content_data merges)
3. Direct `hero_url` / `logo_url` in collected_data (fallback)

All five hero variants are recognised in the merge but only `hero_home` is ever generated. The discovery check `checkMissingAssetRefs` warns when `logo_url` is set but the header has no `<img>` — only logo coverage is monitored.

---

## 6. Discovery and improvement loop

The work-item infrastructure is already structured for image handling:

| Item type | Handler | Source |
|---|---|---|
| `needs_logo` | `image-build-handler` | site-planner output |
| `needs_hero_image` | `image-build-handler` | site-planner output |
| `undeployed_asset` | `asset-deployer` | discovery scan |
| `missing_logo_in_header` | `site-component-linker` | discovery scan |

Adding new item types like `needs_section_illustration`, `needs_team_portraits`, `needs_product_image`, `needs_infographic` is mechanical — same pattern. The orchestrator passes `asset_id?`, `purpose?`, `check?` through `input_mapping`, and the handler is self-contained.

This is the most reusable thing we have. New imagery work flows naturally as new work-item types pointed at appropriate handler agents — we don't need to invent new orchestration shapes.

---

## 7. Adoption: images crawled but discarded

`apply_adoption_plan` and the Firecrawl scrape adapter both extract images:

- `result["images"]` from Firecrawl's `images[]` field
- Filter on `links[]` for `.jpg/.png/.gif/.webp/.svg/.tif/.bmp` extensions, alt text, link text

This data is captured into the crawl response and **then dropped**. Adopted sites lose their original imagery. We don't:
- Persist the source images to `assets`
- Mirror them to our S3
- Reference them from generated pages
- Use them as img2img references for our own generations

For a site we're adopting, the original imagery is the strongest signal of what imagery should look like on the rebuilt site. We're throwing away the best training/reference material we have.

---

## 8. Per-vertical training infrastructure

`018_canine_biology.md` describes the LoRA fine-tuning plan (Phase F: image LoRA from 60-90 curated training images, SDXL or PixArt base, ~£35-95 first pass). This is the route to consistent visual style per vertical (vet diagrams, energy infographics, mortgage charts).

Status: planned, not started. The infrastructure to *use* a trained LoRA — model selection in the request body, multiple adapter routes — does not exist yet. The fine-tuning plan presupposes an adapter that can take a `model: "vet-diagram-v1"` field; our adapter cannot.

---

## 9. Summary of structural gaps

What's missing, ordered roughly by how foundational each is:

1. **Multiple-purpose-per-site model.** The `assets` UNIQUE constraint and the `{purpose}_uri` / `{purpose}_url` slot pattern in `sites.content_data` both assume one image per purpose. To put fifteen icons on a page or six product illustrations in a section, this has to relax. Options: drop the unique constraint and key on `(site_id, purpose, slot)` or `(site_id, asset_key)`; or make `purpose` itself include an index (`product:001`, `icon:trust-shield`).

2. **Adapter that supports model and provider selection.** The current adapter is "send a hardcoded Stability request." For Banana Pro (Gemini), Midjourney, FLUX, SDXL with LoRA, etc. we need either one adapter that routes by `provider` field in the request, or one adapter per provider with a router — same pattern as `aiservice` already has for text (`anthropic`, `ollama`, `xai`, `openai`, `perplexity` via `resolveLLMNewsProvider`). The text-side has a working pattern we can mirror.

3. **A richer image request shape.** `prompt + width + height` is not enough. We need: `model`, `provider`, `negative_prompt`, `style_preset`, `reference_image_uri` (for img2img / variants), `seed`, `lora`, `aspect_ratio`, `samples`, `safety_mode`. Most of these are no-ops on most providers — they just get passed through if supported, ignored otherwise.

4. **A planner that requests varied imagery.** Today the planner produces `{logo, hero_home}`. It would need to produce something like `imagery_plan: [...]` where each entry is `{purpose_key, kind: hero|illustration|icon|product|infographic|decoration, prompt, placement_hint, style_hint, reference_uri?}`. The component selector or component itself would consume those entries based on its `input_schema`.

5. **Components that declare diverse image needs in `input_schema`.** Today only `hero_image` is declared. A `team-grid` component should declare `member_avatars: array of image with source: site_assets.team_member.{slug}`. A `services-grid` component could declare `service_icons` with `source: generated.icon.{service_name}`. The planner reads schemas, sees these declarations, emits the imagery_plan entries.

6. **Icon/SVG rendering.** This is genuinely a separate path from raster image generation. Three plausible paths, not mutually exclusive:
   - LLM-generated inline SVG (cheap, sometimes ugly, never any copyright concern, fast).
   - A curated icon library (Lucide / Heroicons / Phosphor) referenced by name — a `tool-icon` component or a `{{icon "shield"}}` template helper that swaps in inline SVG. The features `icon` field becomes meaningful.
   - Image-model-generated icons via a tight prompt prefix — works for stylised illustrations but not crisp UI icons. Useful for decorative vignettes, less so for chrome.

7. **Infographics path.** Infographics are SVG (or HTML/CSS) generated from data, not raster images. This needs an entirely separate agent — `infographic-generator` — that takes `{data, chart_type, palette, style}` and returns an SVG string, which the renderer inlines. There's no existing scaffolding for this. The closest analogue is `tool-favicon-generator` and the dynamic-tool components in 019.

8. **Connect `design_intent.imagery_direction` to the prompt builder.** Webdesign-agent already writes character descriptions. The image action should prefix or augment its prompt with that direction. One-line read in `getImagePromptWithPriority` plus a prompt-composition step.

9. **Persist crawled imagery during adoption.** When `apply_adoption_plan` walks the Firecrawl result, mirror images to our S3, insert into `assets` with `origin_type='adopted'` and `origin_url=<source>`, populate `attribution` if we can derive it. The site can then either reuse them as-is (if licensing permits) or use them as img2img references for our own generations — both are dramatically better than starting blank.

10. **Per-model agents vs. per-model adapters.** The user raised the option of separate agents for different image models. The simpler version (closer to existing patterns) is one `image-generator` agent, one router-style adapter, model selection in the request. If a model needs genuinely different orchestration (e.g. Midjourney is async with a webhook callback while Stability is sync), then a separate agent is justified. We'd evaluate this per model rather than blanket-adopting one shape.

---

## 10. What's already in place that we can build on

To keep us grounded — these are all reusable, no new structure needed:

- **`assets` table schema** — already has `origin_model`, `origin_asset_id`, `alterations`, `attribution`, `license`. Drop the unique constraint, populate the unused columns, and the table covers our use cases.
- **`store_asset` and `deploy_image_asset` actions** — generic over `purpose`. Adding new purposes is a Go config change, not new code.
- **Work-item dispatch loop** — `image-build-handler` already shows the pattern: `spawn → call → store_asset → deploy_image_asset`. Adding `illustration-build-handler`, `infographic-build-handler` clones the shape.
- **`image_purpose` config map** — adding new entries is a one-line change per purpose.
- **`aiservice` provider router** — text-side has the multi-provider pattern done. Imagery just needs the same shape applied to a new interface (`ImageService`).
- **Component `input_schema` v2** — already supports declaring `type: image` with arbitrary `source`. We just don't use the flexibility.
- **Render context image extraction** — already merges arbitrary `*_url` fields. Wide-open expansion path.
- **`design_intent` site spec aspect** — already written, just not read.

---

## 11. Candidate next steps (not committed)

These would be the natural follow-ups, in roughly the order I'd consider them:

1. **Read** `design_intent.imagery_direction` in `generate_image` and prepend it to prompts. Smallest possible win, no schema change. Tests that the data we already have is reaching the model.
2. **Loosen the `assets` unique constraint** so multiple assets can share a purpose, keyed by an `asset_key` text column (or an indexed `purpose` that includes a slot identifier).
3. **Persist crawled imagery during adoption** — mirror to S3, write to `assets` with `origin_type='adopted'`. This unlocks reference imagery for adopted sites at zero generation cost.
4. **Provider-router on the adapter** — accept a `provider` and `model` field in the request, dispatch to Stability / FLUX / Gemini Banana Pro / Replicate. Most of the work is in the API request shaping per provider.
5. **Planner produces an `imagery_plan`** — list of entries, varied kinds. Each becomes a work item.
6. **Extend `image-build-handler`** to handle `kind: illustration` and `kind: portrait` purposes (different sizes, different deploy paths).
7. **Icon path** — pick one approach (probably inline SVG library + name lookup) and wire it into the features component. This is mostly a renderer change, not an image-generation change.
8. **Infographic-generator agent** — SVG output, takes data and a style brief. Separate from raster image generation.
9. **Per-vertical LoRA** — only after the rest of the plumbing exists. The current canine_biology plan can be the first one.

10. **Discovery and audit checks for imagery delivery** — closing the loop between spec/plan and what was actually produced. Detailed in section 13.

Each of these is small enough to ship as its own piece without breaking the others.

---

## 12. Open questions for the next round

- Do we want one adapter that routes by provider, or N adapters one per provider? (Slight lean toward router — fewer pods, same operational shape as `aiservice`.)
- Where do model preferences live — site spec, vertical config, agent definition default? (Probably vertical config: "vet sites use vet-diagram-LoRA for illustrations, default for everything else.")
- Are crawled images licensed for our use, or only as references for new generation? (Defaults to "reference only" unless we have a specific license signal — `origin_url` and `attribution` are there to track this.)
- Icon strategy: inline SVG library, LLM-generated SVG, or model-generated raster? (Each has a place — answer is probably "all three, picked per use case.")
- For products without copyright-free imagery: do we generate a single illustration, multiple angles, or attempt 3D reconstruction from multiple source photos? (3D is a much bigger project — likely a Phase 2 ambition.)

---

## 13. Discovery, audit, and the spec-to-delivery gap for imagery

The discovery/audit/fix architecture is the right place to catch mismatches between what the spec, planner, and other agents asked for and what got delivered. The infrastructure works well for CSS, content, and component linkage. For imagery it is almost empty.

### 13.1 What's there for imagery today

Two checks. That's all.

| Check | Where | Detects | Handler |
|---|---|---|---|
| `undeployed_assets` | `design-discovery-agent` | `assets` row exists but no `page_components.rendered_html` references the file path | `asset-deployer` |
| `missing_logo_in_header` (`checkMissingAssetRefs`) | `design-discovery-agent` (sub-check of `validate_component_standards`) | `sites.logo_url` is set but the rendered header has no `<img>` tag | `site-component-linker` |

Both check whether an asset *that already exists* reaches the page. Neither checks whether the asset that was *asked for* actually exists.

The visual-design-auditor (the LLM-based one) loads `style_collection`, `palette`, `typo`, `css_excerpt`, `component_samples`, `index_samples` (excerpts of rendered HTML up to 800 characters per component, 600 per page section). It does **not** load:

- Any `assets` rows
- Any image URLs from `sites.content_data` or render context
- The `site_plan.image_prompts` it was meant to deliver against
- The `design_intent.imagery_direction` it was meant to honour

The auditor's prompt is explicitly about **colour, spacing, typography, dark sections, responsive**. Imagery is not in the checklist. Even if a site shipped with no hero image at all — or with a hero image that completely contradicts the brief — the auditor would not notice.

### 13.2 The discovery checks that don't exist yet

These are the spec→delivery gaps the algorithmic layer could catch (no LLM, no cost). All of them follow the existing `DiscoveryCheck` interface — a Go file in `platform/orchestration/actions/discovery_checks/` with an `init()` registration. Each produces a `WorkItemSpec` pointed at an existing handler agent.

| Check name (proposed) | Detects | Handler | Source signal |
|---|---|---|---|
| `unfulfilled_image_prompt` | `site_plan.image_prompts` (or `site_specs.site_plan.image_prompts`) has a key with no corresponding `assets` row for that purpose | `image-build-handler` | The planner asked for X, no asset exists |
| `unfulfilled_imagery_plan_entry` | (Future, depends on richer planner output.) `imagery_plan[]` entry has no asset matching its `purpose_key` | `image-build-handler` | Per-section/per-component imagery never generated |
| `image_purpose_orphan` | `assets` row exists for a purpose that no component schema or page references | (none yet — flag only) | Wasted generation, or stale spec |
| `image_url_404` | `pages.rendered_html` references `/assets/images/X.jpg` but no `assets` row has that filename and no file is committed in git | `asset-deployer` (regenerate) or `image-build-handler` | Component refers to image we never delivered |
| `image_alt_text_missing` | `<img>` tags in rendered HTML have no `alt` attribute or empty alt | `component-template-fixer` | Accessibility gap |
| `placeholder_image_in_use` | `pages.rendered_html` contains the fallback `/assets/images/hero.jpg` *and* `assets` has no row with `purpose='hero'` | `image-build-handler` | Fallback is showing because nothing real was generated |
| `imagery_direction_unread` | `site_specs.design_intent.imagery_direction` is set but `assets.origin_prompt` for hero/logo doesn't include any tokens from it | `image-build-handler` (regenerate with directive) | Designed direction never reached the model |
| `crawled_images_discarded` | Adoption `research_results` contains image URLs from crawl, but `assets` has no rows with `origin_type='adopted'` for that site | (new) `adoption-image-mirror` | Reference imagery thrown away |
| `cross_site_image_contamination` | `pages.rendered_html` references an image whose `assets.site_id` is a different site | (existing pattern from `cross_site_contamination`) | Wrong site's image rendered (a real risk once asset_key is shared across sites) |
| `image_dimensions_wrong` | Deployed image's actual dimensions don't match the `ImagePurposes` config for its purpose | `asset-deployer` (re-optimise) | Optimisation step skipped or misconfigured |
| `unused_icon_string` | Component `content_data.features[].icon` strings exist but rendered HTML has no SVG or icon for them | `component-template-fixer` (or new `icon-resolver`) | Icon contract is one-sided |
| `multi_image_component_underfilled` | A component declares `team_avatars` (array) in its schema but rendered HTML has fewer `<img>` than the array length | `image-build-handler` (per-entry) | Image-array components only partially populated |

These are all derivable from existing tables: `assets`, `pages`, `page_components`, `site_components`, `site_specs`, `sites.content_data`, `research_results`. None requires new infrastructure.

### 13.3 What the LLM auditor cannot do today (and could)

The visual-design-auditor receives only HTML excerpts. To audit imagery it would need to receive image URLs (and ideally the prompts that produced them) so it can reason about visual fit:

- "The brief asks for a 'professional, calm, mid-century law firm' but the hero is a vibrant high-energy abstract. Mismatch."
- "The hero on `services.html` is the same image as on `index.html`. Variety would help."
- "Three feature cards have icon names but no visual treatment in the rendered HTML."
- "The logo's colour palette doesn't include any of the brand palette colours."

This is a structural gap, not a prompt gap. The auditor's `load_design_context` SQL doesn't pull image data because there's nothing to pull from — image URLs live in `sites.content_data` and `assets`, neither of which the existing query touches.

**To extend it, three changes:**

1. Augment `load_design_context` to also load: `sites.content_data` keys ending in `_url`, `assets` rows with `is_active=true`, and `site_specs.design_intent.imagery_direction`.
2. Update the visual auditor prompt to include an "Imagery" check category alongside Colour/Spacing/Typography/Dark Sections/Responsive.
3. (For real visual analysis, not just URL-presence checks) feed the auditor a vision-capable model and pass image URLs as image content blocks. This is a meaningful upgrade — currently all auditor calls go through text-only paths.

The vision-capable upgrade is the bigger lift but it's the only way to detect "the hero image looks wrong for this brand." Without it the audit is limited to "imagery exists / imagery doesn't exist" — useful but shallow.

### 13.4 A separate "imagery-audit-agent"?

Two patterns are available:

**Option A — extend visual-design-auditor.** Imagery becomes a sixth check category. One LLM call, one set of findings, one work-item write. Lighter operationally. Risk: the prompt becomes long enough that the model loses focus on any single dimension; visual-design-auditor's TOP-5 cap means imagery findings compete with colour/spacing for attention.

**Option B — new `imagery-quality-auditor`.** Sibling of visual-design-auditor under `design-audit-agent`. Separate LLM call dedicated to imagery, separate TOP-5 cap, separate findings. More predictable audit coverage. Cost: one more LLM call per audit pass.

The 015_batch_processing pattern (queue_llm_batch with callback) makes Option B cheaper than it looks — these calls can run in batch alongside other auditors. I'd lean toward B because imagery audit naturally also wants vision-capable inputs that the existing auditor doesn't need; mixing the two means every visual-design-audit pays the vision-model cost for non-imagery findings.

### 13.5 Fix-handler additions

Most imagery findings can be handled by handlers that already exist:

| New work item type | Existing handler | Notes |
|---|---|---|
| `unfulfilled_image_prompt` | `image-build-handler` | Already takes `purpose` and `image_prompts` from `input_data.spec`. Wiring is mechanical |
| `image_url_404` | `image-build-handler` (regenerate) or `asset-deployer` (redeploy if asset exists in S3 but not in git) | Branch on whether `assets` row exists |
| `placeholder_image_in_use` | `image-build-handler` | Same path as `needs_hero_image` |
| `imagery_direction_unread` | `image-build-handler` | Regenerate with directive included in prompt |
| `image_alt_text_missing` | `component-template-fixer` | Existing patterns: this is a template-injection job |
| `image_dimensions_wrong` | `asset-deployer` | Already optimises by purpose; just re-trigger |

New handlers needed:

| Handler | Item type | What it does |
|---|---|---|
| `adoption-image-mirror` | `crawled_images_discarded` | Reads `research_results`, downloads images from source URLs, uploads to our S3, inserts `assets` rows with `origin_type='adopted'` |
| `icon-resolver` | `unused_icon_string` (optional — could be `component-template-fixer`) | Looks up icon name in icon library, injects SVG into template, marks pages for rerender |

Both are small. Neither needs new infrastructure — they fit the existing handler shape (receive site_id + work item spec, do their job, produce `needs_rerender` if anything changed).

### 13.6 The `design_intent` ↔ delivery comparison

The most useful single audit is: **does what was generated reflect what the design intent said it should?**

`site_specs.design_intent.imagery_direction` is a free-text field set by the webdesign-agent or strategist. It says things like "warm, slightly retro, hand-drawn feel, no stock photography, prefer illustration over photo." Without an audit, this is decorative — nobody checks whether the actual hero is "warm, slightly retro, hand-drawn."

A discovery check can do a weak version of this (token overlap between `imagery_direction` and `assets.origin_prompt`). An LLM auditor with vision can do the strong version (look at the actual image, compare to the description).

This is one of the highest-leverage additions because it closes the loop on the work webdesign-agent is already doing — currently writing intent that nobody reads.

### 13.7 Pass-counting and direction respect

The existing improvement loop has guardrails worth preserving for imagery:

- **Audit pass cap** (3 passes per site, in `sites.settings`). Imagery audit findings should count toward this so we don't loop forever on impossible-to-fix imagery (e.g. provider failures).
- **`max_fix_attempts: 2`** is already a structured-finding field. Imagery findings should set this — generation is non-deterministic, so two attempts with different seeds is reasonable; ten is wasteful.
- **Locked components** (`page_components.locked_at`) are excluded from auditor data queries. Same should apply to manually-uploaded assets — once a human uploads a real photo, audit must not flag it as "doesn't match the AI prompt direction" and try to regenerate it.
- **Direction must_have features** are honoured by content-quality-auditor. The same protection applies to imagery: if a human said "must have a photograph of the actual team," the auditor must not propose replacing it with an AI illustration.

The `assets` table doesn't currently have a `locked` or `human_uploaded` flag. `origin_type='uploaded'` is the closest signal — that's probably enough as a first cut, with auditor data queries filtered to `origin_type IN ('generated', 'adopted')` so human uploads are skipped.

### 13.8 What this section adds to the gaps and next steps

Updates to section 9 (gaps): add an item between current 8 and 9.

> **8.5 Audit and discovery layer is blind to imagery.** Two checks exist (asset deployed, logo in header). The visual-design-auditor's data context contains zero image data. The most powerful piece of metadata in the system for guiding imagery — `design_intent.imagery_direction` — is written but never read by anyone, including the auditor that's supposed to enforce it.

Updates to section 11 (next steps): the new item 10 expands to:

> **10a.** Add three algorithmic discovery checks first: `unfulfilled_image_prompt`, `placeholder_image_in_use`, `image_url_404`. Each is ~50-80 lines of Go in `discovery_checks/`. They detect the bulk of "spec asked for it, didn't ship" cases without any LLM cost.
>
> **10b.** Wire `image-build-handler` to handle the new item types. Its workflow already takes `input_data.spec.image_prompts` and `purpose` — should be a config-only change for the new item types pointing at it.
>
> **10c.** Extend `load_design_context` in visual-design-auditor to include `assets` rows and `imagery_direction`. Add an "Imagery" check category to the auditor prompt. Cheapest LLM-side win.
>
> **10d.** New `adoption-image-mirror` handler — small, self-contained, addresses the discarded-imagery loss directly.
>
> **10e.** Vision-capable auditor (probably as `imagery-quality-auditor`, sibling to visual-design-auditor). The bigger lift — needs a vision model on the auditor path. Defer until 10a-10d are working and we have actual delivered imagery to feed it.

Updates to section 12 (open questions):

> - Should `imagery-quality-auditor` be a separate agent or a category within visual-design-auditor? (Leaning separate, per 13.4.)
> - When a generated image is found to mismatch direction, how many regeneration attempts before we mark it `needs_human_review` and stop? (Following the existing `max_fix_attempts: 2` precedent.)
> - Should `assets` get an explicit `locked` boolean (mirroring `page_components.locked_at`), or is `origin_type='uploaded'` enough? (Probably enough initially; revisit if false positives appear.)
> - Do we want the audit to suggest *which* image needs replacing on a per-section basis, or just flag site-level imagery direction mismatch? (Per-section is more useful but requires the richer imagery_plan from section 11.5 to land first.)

---

## Part 2 — The sequenced audit-and-fix plan

Sequenced work plan for closing the gap between what the planner / spec / other agents ask for in imagery, and what is actually delivered. Builds on the assessment in Part 1 above.

---

## Decisions taken

| Question | Decision |
|---|---|
| Imagery audit — extend existing or new agent? | New agent: **`imagery-quality-auditor`**, sibling of `visual-design-auditor` under `design-audit-agent`. |
| Max regeneration attempts per finding | **2** (matches the existing `max_fix_attempts: 2` on structured findings). |
| Asset locking | **Mirror `page_components` exactly**: add `locked_at timestamptz` and `lock_type text` to `assets`, same query exclusion patterns. |
| Per-section / per-component image granularity | **Deferred**. Site-level imagery audit is sufficient for now. Don't introduce a richer `imagery_plan` schema until the basic loop is working. |

These decisions shape the plan: small additive changes, reusing existing handlers, no wholesale planner refactor.

---

## Progress (updated 2026-05-08)

Phases 0 and 1 are complete and end-to-end verified. The current row of an
`assets` insert against `00ff3af5-...` confirms the imagery_direction
prefix lands in `origin_prompt`, `origin_model='sdxl'` populates, and the
URL extraction works. See `STATUS_imagery_2026-05-08.md` for the
verification trail.

| Phase | Status |
|---|---|
| 0.1 | ✅ verified |
| 0.2 | ✅ verified |
| 1.1–1.5 | ✅ verified (Phase 1.5 needed an unplanned hotfix; details below) |
| 2 onwards | not started |

**Unplanned Phase 1.5 hotfix.** During verification we found that
`image-build-handler.call_logo_gen` and `call_hero_gen` were missing the
`output_mapping` block that `pageflow-builder` and `site-work-orchestrator`
already had. Without it, the URL the generator returned lived three levels
deeper than `store_asset` looked for it, so every dispatch-path image asset
silently `{stored: false}`'d. Fix copies the canonical 4-field
output_mapping to the two missing steps. SQL only, no code. Files:
`phase_1_5_pre_migration_backup.sql`, `phase_1_5_image_build_handler_output_mapping.sql`.

**Findings from verification (not blockers, but tracked):**

- Build-dispatch-loop isn't claiming triaged image items when page items
  are also in queue. Imagery items have been sitting in `triaged` for
  hours while page work flows. Worth a separate investigation; doesn't
  block Phase 2.
- The 5 hero variant flag-only items are being claim-attempted by
  dispatch despite empty `handler_agent`. They reach `failed` after 3
  attempts. Phase 2 makes them routable, sidestepping the issue.

---

## Sequencing principles

- Each phase is shippable on its own. No phase depends on a later phase.
- Schema changes are isolated to Phase 2; everything before that runs against today's schema.
- LLM-side changes (Phase 4 onwards) are gated on the algorithmic checks in Phase 1 working — the auditor should not have to re-discover what algorithms already caught.
- Each step within a phase is sized to a single PR / single deploy. No step touches more than 2-3 files.
- Reuse is called out per step. New code is the exception, not the default.

---

## Phase 0 — Wire what's already written but unread (no schema change)

**Goal:** Stop ignoring data we already have. Smallest possible change, demonstrates the path is live.

### 0.1 Read `imagery_direction` into the image prompt

- **File:** `platform/orchestration/actions/generate_image_actions.go` (or wherever `getImagePromptWithPriority` lives — confirmed by grep before editing).
- **Change:** After resolving the prompt via three-tier priority, look up `site_specs.aspect = 'design_intent'` for the current site, extract `data.imagery_direction`. If non-empty, prepend as a directive section: `"Style direction: <direction>\n\nSubject: <prompt>"`.
- **Reuse:** `datahelpers.ExtractNestedFieldString`, the existing site-spec read pattern from webdesign-agent.
- **Acceptance:** Log shows `imagery_direction` text in the prompt sent to the adapter. Site with `design_intent.imagery_direction = "warm hand-drawn illustration"` and `image_prompts.hero_home = "law firm office"` produces a hero whose origin_prompt contains both fragments.
- **Risk:** None significant. Worst case the directive is empty/missing and we fall through to today's behaviour.

### 0.2 Populate `origin_model` in `store_asset`

- **File:** the `store_asset` action (Go).
- **Change:** Read the active model from `agent_config["model"]` or the agent_definition's `default_config.model` for the calling agent. Write to `assets.origin_model` on insert.
- **Reuse:** Existing `loadAgentDefinitionForImageAction` or equivalent — it already returns the agent's `default_config`.
- **Acceptance:** Newly built site has `assets.origin_model = 'sdxl'` (or whatever Stability returns) on the hero/logo rows.
- **Why now:** It's a one-line write that costs nothing and immediately makes provenance visible. When we add provider routing later, the column is already populated.

### 0.3 (Verification only) Confirm the data is reaching where it's supposed to

Quick SQL check after a build:
```sql
SELECT a.purpose, a.origin_model, LEFT(a.origin_prompt, 200) as prompt_preview
FROM assets a WHERE a.site_id = '<site>';
```

`prompt_preview` should include the imagery_direction tokens. If it doesn't, 0.1 isn't fully wired.

---

## Phase 1 — Algorithmic discovery checks for spec-to-delivery gap

**Goal:** Catch the simple, structural mismatches without any LLM cost.

All three checks live in `platform/orchestration/actions/discovery_checks/`, follow the existing `DiscoveryCheck` interface, register via `init()` into the registry. Run from `design-discovery-agent`.

### 1.1 `check_unfulfilled_image_prompt.go`

- **Detects:** `site_specs` aspect `site_plan` has `data.image_prompts.hero_home` (or `data.image_prompts.logo`) but no `assets` row exists for that purpose.
- **Work item:** `item_type = 'needs_hero_image'` or `'needs_logo'`, **routes to existing `image-build-handler`**. No new handler.
- **Spec construction:** Copy `image_prompts` from the site_plan spec into the work item's `spec` field so the handler workflow finds it at `input_data.spec.image_prompts.hero_home`.
- **SQL (rough shape, validate against site_specs schema):**
  ```sql
  SELECT data->'image_prompts' AS prompts
  FROM site_specs
  WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true;
  -- then in Go: for each key in prompts, check if assets row exists
  SELECT COUNT(*) FROM assets
  WHERE site_id = $1 AND purpose = $2 AND status = 'active';
  ```
- **Acceptance:** Site with `image_prompts.logo` set but `assets.purpose='logo'` empty → work item created, `image-build-handler` picks it up via dispatch loop, generates and deploys logo.

### 1.2 `check_placeholder_image_in_use.go`

- **Detects:** Any `page_components.rendered_html` for the site contains the fallback path `/assets/images/hero.jpg` (or `/assets/images/logo.png`) **and** no `assets` row exists with the matching purpose. The asset never got generated; the component schema fell through to its hardcoded fallback.
- **Work item:** Same `needs_hero_image` / `needs_logo` types, same handler.
- **Acceptance:** Site whose hero section HTML contains `/assets/images/hero.jpg` but `assets.purpose='hero'` is empty → work item created, fixed by image-build-handler.
- **Why this matters:** Today this case fails silently. The site looks "fine" because the fallback resolves, but the brief asked for a hero image and one was never made.

### 1.3 `check_image_url_404.go` (lightweight version)

- **Detects:** `page_components.rendered_html` references `/assets/images/X.{jpg,png,webp}` where `X` doesn't match any `assets.purpose` for this site.
- **Work item:** Best-effort — if we can derive intent from context, route to `image-build-handler`; otherwise route to a flag-only finding.
- **Note:** This is the DB-only version. The "is the file actually committed in git?" version would need git-adapter integration — defer that.
- **Acceptance:** Site with manual `<img src="/assets/images/team-photo.jpg">` injected into a component but no `assets` row for `team-photo` → work item flagged.

### 1.4 Wire all three into `design-discovery-agent`

- **File:** `bk_agent_definitions_backup.sql` row for `design-discovery-agent`, `default_config.workflow.steps.run_checks.config.checks` array.
- **Change:** Append `"unfulfilled_image_prompt"`, `"placeholder_image_in_use"`, `"image_url_404"` to the existing checks list.
- **Reuse:** The single-line config update is the only change to the agent itself. The discovery check Go files do all the work.
- **Acceptance:** Trigger `design-discovery-agent` manually for a test site (`./trigger-audit.sh design-discovery-agent <site_id>`); look at `site_work_items` for the new item types.

### 1.5 `image-build-handler` smoke test for new item types

- **No code change** — the handler already handles `needs_logo` and `needs_hero_image`.
- **Verification:** Check that work items created by the new discovery checks flow through `claim → spawn → call → store_asset → deploy_image_asset → complete`, same as items created by the planner.
- **Risk:** The work item spec format from the discovery checks must match what `image-build-handler` reads at `input_data.spec.image_prompts.hero_home`. Confirm with a real test before declaring 1.1 done.

---

## Phase 2 — Schema groundwork: locking and multi-image readiness

**Goal:** Mirror the component locking pattern on assets. Open the door for multi-image purposes without forcing existing callers to change.

### 2.1 Add `locked_at` and `lock_type` to `assets`

- **Migration SQL:**
  ```sql
  ALTER TABLE assets
    ADD COLUMN locked_at timestamptz,
    ADD COLUMN lock_type text;
  CREATE INDEX idx_assets_locked ON assets(site_id, locked_at)
    WHERE locked_at IS NOT NULL;
  ```
- **Schema check:** `\d page_components` — confirm exact column types and pattern. Mirror precisely.
- **Reuse:** Lock-expiry logic in audit data queries — see `004_improvement_loop.md` "data queries EXCLUDE locked components (locked_at IS NULL or expired)".
- **Lock types to define:** `'manual'` (human uploaded a real photo and locked it), `'audit_pending'` (under audit, do not regenerate yet), `'expired_at:<timestamp>'` for time-bounded locks. Match `page_components` lock_type vocabulary exactly — do not invent new values.

### 2.2 Update audit and discovery data queries to honour locks

Anywhere a query reads `assets`, add the lock filter:

```sql
WHERE (locked_at IS NULL OR locked_at < NOW())
```

- **Files to update:**
  - `discovery_checks/check_undeployed_assets.go`
  - `discovery_checks/check_unfulfilled_image_prompt.go` (from Phase 1 — add filter on first write)
  - `discovery_checks/check_placeholder_image_in_use.go`
  - The `load_design_context` query in `visual-design-auditor` (Phase 4)
  - Any future imagery-quality-auditor query

### 2.3 Add `asset_key` column (preparation for multi-image)

- **Migration SQL:**
  ```sql
  ALTER TABLE assets ADD COLUMN asset_key text;
  UPDATE assets SET asset_key = purpose WHERE asset_key IS NULL AND purpose IS NOT NULL;
  CREATE UNIQUE INDEX idx_assets_site_asset_key_unique
    ON assets(site_id, asset_key)
    WHERE asset_key IS NOT NULL AND status = 'active';
  ```
- **Do not drop the existing `idx_assets_site_purpose_unique` yet.** It still applies for `purpose IS NOT NULL`. Keep both for now — they overlap harmlessly while `asset_key = purpose` for everything.
- **Why:** Adoption mirroring (Phase 3) will write multiple `assets` rows for the same logical purpose-area (e.g. five mirrored images all with `purpose='adopted_image'`, distinguished by `asset_key='adopted:<filename>'`). Without the asset_key path, those inserts would fail the existing unique constraint.

### 2.4 Update `store_asset` to write `asset_key`

- **File:** the `store_asset` action.
- **Change:** Accept an optional `asset_key` config field. If absent, default to `purpose`. Write it on insert/upsert.
- **Backwards compatibility:** Existing callers (pageflow-builder, image-build-handler) keep working unchanged because their workflow doesn't pass `asset_key` and the default falls back to `purpose`.
- **Acceptance:** Existing builds produce `assets.asset_key = assets.purpose`. New callers can pass `asset_key='adopted:hero-bg.jpg'` and it lands in the column.

### 2.5 Drop the old purpose-only unique constraint

- **Only after** Phase 3 (adoption mirror) is using `asset_key` and Phase 2.3 has been live for some weeks with `asset_key` populated.
- **Migration SQL:**
  ```sql
  DROP INDEX idx_assets_site_purpose_unique;
  ```
- **Risk:** If any callsite relies on the old constraint for upsert behaviour (`ON CONFLICT (site_id, purpose) DO UPDATE`), those need updating to `ON CONFLICT (site_id, asset_key)` first. Grep before dropping.

---

## Phase 3 — Adoption image mirror

**Goal:** Stop discarding crawled imagery. Every adopted site retains its source images either as direct deploys (where licensing allows) or as references for our own generations.

### 3.1 New action `mirror_adoption_images`

- **Location:** `platform/orchestration/actions/mirror_adoption_images_action.go`.
- **Inputs:** `site_id`, `research_results_id` (or read latest `research_results` row with `result_type = 'adoption_crawl'` for the site).
- **Behaviour:**
  1. Read crawl result, extract image URLs from `images[]` array and link-extension matches.
  2. For each image: download (existing `webscrape` or direct HTTP through circuit breaker), upload to S3 via existing `storage.S3Client.Upload`.
  3. Insert one `assets` row per image:
     - `origin_type = 'adopted'`
     - `origin_url = <source URL>`
     - `purpose = 'adopted_image'`
     - `asset_key = 'adopted:<filename or hash>'`
     - `attribution`, `license` from any signal we can derive (often blank initially).
- **Reuse:** S3 client, HTTP client with circuit breaker (`platform/resilience`), the data extraction pattern from the existing scrape-result handlers.
- **Limits:** Cap per site (e.g. 50 images) and per-image size (e.g. 5MB). Log skipped items, don't fail the action for individual download failures.

### 3.2 Wire into `apply_adoption_plan`

- **File:** `bk_agent_definitions_backup.sql` row for `site-adoption-agent`, workflow steps.
- **Change:** Add a step `mirror_adoption_images` after `apply_adoption_plan`, before `complete`.
- **Reuse:** No new agent. The existing adoption agent is the natural place.
- **Acceptance:** Adopt a site that has 12 source images → 12 `assets` rows with `origin_type='adopted'` after adoption completes.

### 3.3 New discovery check `check_crawled_images_discarded.go` (backfill)

- **Detects:** Site has `research_results` rows with `result_type = 'adoption_crawl'` containing image URLs, but no `assets` rows with `origin_type = 'adopted'` for that site.
- **Work item:** `item_type = 'crawled_images_discarded'`, routes to a new `adoption-image-mirror` agent (3.4).
- **Why a check rather than just a backfill script:** Sites adopted before Phase 3 lands need this. After all old sites are processed, the check becomes a no-op for new sites (because 3.2 prevents the gap).

### 3.4 New agent `adoption-image-mirror`

- **Category:** `specialist`. Orchestrator processing mode.
- **Workflow:**
  ```
  ensure_site_record → mirror_adoption_images → complete
  ```
- **The agent is essentially a one-step wrapper** around the action from 3.1. We make it an agent (not just an action call from the dispatch loop) because that's the project pattern: every dispatched work item type maps to an agent.
- **Definition:** New row in `agent_definitions`. Reuse the resource limits, image_repository, image_tag from sibling specialists.

### 3.5 Reference imagery available to image-generator

- **No code change yet** — once `assets` has `origin_type='adopted'` rows, downstream work (img2img generation, style references, future per-section illustrations) can read them.
- **Future hook (deferred):** When the image-generator gains `reference_image_uri` support, prefer adopted images for the matching purpose / section.

---

## Phase 4 — Visual auditor extension (text-only, cheapest LLM-side win)

**Goal:** Make the existing visual-design-auditor aware that imagery exists, without a vision model.

### 4.1 Extend `load_design_context` SQL

- **File:** `visual-design-auditor` agent definition, `default_config.workflow.steps.load_design_context.config.query`.
- **Change:** Add subqueries to the existing big SELECT:
  - `(SELECT json_agg(json_build_object('purpose', purpose, 'asset_key', asset_key, 'url', url, 'origin_prompt', origin_prompt)) FROM assets WHERE site_id = s.id AND status = 'active' AND (locked_at IS NULL OR locked_at < NOW()) AND origin_type IN ('generated', 'adopted')) as assets`
  - `(SELECT data->>'imagery_direction' FROM site_specs WHERE site_id = s.id AND aspect = 'design_intent' AND is_current = true) as imagery_direction`
- **Schema check:** Confirm `site_specs.aspect = 'design_intent'.data.imagery_direction` is the actual JSON path. Some site_specs rows have nested structure — verify with a sample row.
- **Reuse:** Same query template style as the existing `component_samples` and `index_samples` subqueries.
- **Risk:** Query becomes large. If it breaks the action's prompt template, split into two query steps (`load_design_context` + `load_imagery_context`) and merge in the prompt.

### 4.2 Update auditor LLM prompt

- **File:** Same agent definition, `default_config.workflow.steps.run_visual_llm_audit.config.prompt`.
- **Change:** Add `IMAGERY` as the 6th check category in the existing `Check for:` list. Add a paragraph providing the new context:
  ```
  Imagery context:
  - Imagery direction (designer's intent): {{.design_context.imagery_direction}}
  - Generated assets: {{.design_context.assets}}
  
  6. IMAGERY: missing required images, hero/logo absent, imagery that contradicts the imagery_direction
  ```
- **Acceptance:** Site with no hero asset gets a finding like `{"category": "imagery", "description": "no hero image generated", ...}`.
- **Risk of false positives:** The auditor might double-flag what the algorithmic checks (Phase 1) already caught. Mitigate by adding to the prompt: `"Algorithmic check results (already handled — do not re-report these): missing hero: {{.algorithmic_results.missing_hero}}, missing logo: {{.algorithmic_results.missing_logo}}"`. Same pattern as the existing algorithmic_results passthrough.

### 4.3 Tune for false positives over a few real sites

- Run 4.1 + 4.2 against 5-10 existing sites without enabling automatic fixes.
- Read findings. Any imagery findings that are obviously wrong (e.g. flagging a correctly-generated hero as "doesn't match direction") get triaged against the prompt.
- **Acceptance:** ≥80% of imagery findings on a random sample are accurate (i.e. fixing them would actually improve the site).

---

## Phase 5 — Vision-capable LLM path

**Goal:** Add the foundational capability for an auditor to actually look at images. Required by Phase 6.

### 5.1 Extend `aiservice.AIService` interface

- **File:** `platform/aiservice/interface.go`.
- **Change:** Add a method:
  ```go
  GenerateTextWithImages(ctx context.Context, prompt string, imageURLs []string, options map[string]interface{}) (string, error)
  ```
  Or, equivalently, accept `image_urls` in the existing options map — interface preference depends on local conventions; check what `GenerateText` does today before deciding.
- **Backwards compatibility:** Existing `GenerateText` callers unchanged.

### 5.2 Implement Anthropic vision

- **File:** `platform/aiservice/anthropic.go`.
- **Change:** Build messages with image content blocks per Anthropic's vision API:
  ```json
  {
    "role": "user",
    "content": [
      {"type": "image", "source": {"type": "url", "url": "<image_url>"}},
      {"type": "text", "text": "<prompt>"}
    ]
  }
  ```
- **Note:** Image URLs must be publicly accessible. Our presigned B2 URLs work (valid for 7 days), but expired ones won't. Auditor should check `assets.url` is a fresh presigned URL or fall back to regenerating one before the call.
- **Reuse:** Existing HTTP client, error handling, retry logic.

### 5.3 Workflow action wrapper

- **Decision:** Extend `execute_llm_prompt` action to accept an `image_urls_field` config that resolves to a list of URLs at runtime, OR add a new `execute_llm_prompt_with_vision` action.
- **Lean toward extending** the existing action — fewer registry entries, same workflow shape. Add the new field only if vision is requested.
- **Acceptance:** A test workflow calling `execute_llm_prompt` with `image_urls_field: "design_context.assets[].url"` produces an LLM response that reflects the image content (not just text).

### 5.4 Cost monitoring

- Anthropic vision is more expensive per call than text-only. Add `vision_call: true` to `llm_call_log` metadata so cost analysis can separate them.
- **Reuse:** existing `llm_call_log` writer.

---

## Phase 6 — `imagery-quality-auditor` agent

**Goal:** A separate agent dedicated to imagery audit, vision-capable, sibling of `visual-design-auditor`.

### 6.1 Agent definition

- **Category:** `analyst`. Processing mode: `orchestrator` (every agent is an orchestrator).
- **Resource limits:** Match `visual-design-auditor`.
- **Workflow:**
  ```
  ensure_site_record
    → load_imagery_context        (SQL: assets, imagery_direction, identity)
    → check_has_imagery           (conditional — skip if no assets to audit)
    → run_imagery_audit_vision    (execute_llm_prompt with image_urls)
    → write_audit_findings        (existing action, source = 'imagery-quality-audit')
    → complete
  ```
- **Reuse:** `ensure_site_record`, `query_database`, `write_audit_findings`, the conditional pattern — all existing actions. Only the audit prompt is new.

### 6.2 Audit prompt

The prompt should follow the existing TOP 5 / acceptance_test contract used by `visual-design-auditor`. Categories specific to imagery:

- `direction_mismatch` — image doesn't reflect the imagery_direction
- `brand_mismatch` — colour palette of generated image clashes with site palette
- `inconsistency` — multiple hero variants don't share a visual style
- `quality` — visible artefacts, garbled text in generated image, AI giveaways
- `inappropriate` — generated content unsuitable for the site's audience or tone

Each finding outputs `max_fix_attempts: 2` (Decision 2 above) and a handler routing recommendation:
- For direction/brand/inconsistency mismatches: `image-build-handler` regenerates with an updated prompt
- For quality issues: `image-build-handler` regenerates with a different seed
- For inappropriate content: `image-build-handler` regenerates with stronger negative prompts; if 2 attempts fail, mark as `needs_human_review`

### 6.3 Lock honouring

- The `load_imagery_context` query filters `WHERE (locked_at IS NULL OR locked_at < NOW()) AND origin_type IN ('generated', 'adopted')`. Human uploads (origin_type='uploaded') are excluded — same pattern as content-quality-auditor's locked-component exclusion.

### 6.4 Hook into `design-audit-agent`

- **File:** `design-audit-agent` agent definition.
- **Change:** Add `spawn_imagery_auditor` and `call_imagery_auditor` steps, parallel to the existing visual and content auditor calls. Sequence: visual → content → imagery → triage.
- **Reuse:** Identical pattern to the existing two auditor calls. Copy-modify the step definitions.

### 6.5 Pass count and rate limiting

- Imagery audit findings count toward the existing 3-pass cap on the improvement loop (`sites.settings.audit_pass_count`). No new counter — same mechanism.
- **Per-finding `max_fix_attempts: 2`** caps regeneration thrash.
- **No new infrastructure needed** — both guardrails use existing fields.

### 6.6 Initial deployment as gated rollout

- Deploy with `is_active = false` in the `design-audit-agent` workflow first; verify the agent runs cleanly on a test site (`./trigger-audit.sh imagery-quality-auditor <site_id>`).
- Then flip the call site in design-audit-agent live.
- **Reuse:** Same gradual-rollout pattern from the batch processing migration.

---

## What we're explicitly NOT doing in this plan

The user's deferral on per-section granularity simplifies several things:

- **Planner output stays as-is.** No richer `imagery_plan` schema. `image_prompts.{logo, hero_home}` continues to be the contract.
- **No per-component asset_key enforcement.** `asset_key = purpose` for everything in the build pipeline. Adoption-image-mirror uses `asset_key = adopted:<filename>` because it has multiple per site, but no other code path uses asset_key as a discriminator.
- **No icon resolver.** The `icon` string field in features content_data continues to be unrendered. Defer until per-section imagery is on the table.
- **No infographic generator agent.** Same reason.
- **No image-generator provider router.** Stability remains the only provider. Multi-provider is a separate plan when we want to bring Banana Pro / FLUX / Midjourney in.
- **No img2img / reference-image generation path.** Adopted images are persisted but only as historical record for now. They become useful when reference-image generation is added.

These are valid follow-ups; they're just not in this plan because they don't belong to the spec-to-delivery loop closure.

---

## Phase summary table

| Phase | Goal | New files | Modified files | Schema changes | Agents touched |
|---|---|---|---|---|---|
| 0 | Wire imagery_direction; populate origin_model | none | `generate_image_actions.go`, `store_asset_action.go` | none | none |
| 1 | Algorithmic discovery checks | `check_unfulfilled_image_prompt.go`, `check_placeholder_image_in_use.go`, `check_image_url_404.go` | `agent_definitions` (design-discovery-agent config) | none | design-discovery-agent (config only) |
| 2 | Asset locking + multi-image readiness | none | `assets` table | `locked_at`, `lock_type`, `asset_key` columns; new index; eventually drop old constraint | none |
| 3 | Adoption image mirror | `mirror_adoption_images_action.go`, `check_crawled_images_discarded.go` | `agent_definitions` (site-adoption-agent workflow) | none | site-adoption-agent, new `adoption-image-mirror` |
| 4 | Visual auditor sees imagery (text-only) | none | `agent_definitions` (visual-design-auditor) | none | visual-design-auditor |
| 5 | Vision-capable LLM path | maybe `execute_llm_prompt_with_vision_action.go` (TBD) | `aiservice/interface.go`, `aiservice/anthropic.go`, possibly `execute_llm_prompt_action.go` | none | none directly |
| 6 | imagery-quality-auditor agent | none (workflow lives in agent_definitions) | `agent_definitions` (design-audit-agent) | none | new `imagery-quality-auditor`, design-audit-agent |

---

## Risks and mitigations

| Risk | Phase | Mitigation |
|---|---|---|
| Discovery check spec format mismatch with image-build-handler | 1.5 | Run a real test before declaring 1.1 done. Confirm work item spec lands at `input_data.spec.image_prompts.hero_home`. |
| `assets` UNIQUE constraint blocks adoption mirror | 2.3 → 3.1 | 2.3 must land before 3.1 deploys. The plan sequences this correctly. |
| Visual auditor false positives on imagery findings | 4.3 | Manual sample run on 5-10 sites before enabling fixes. |
| Vision API URL expiry | 5.2 | Refresh presigned URLs immediately before the call. Add to the auditor's `load_imagery_context` step. |
| Audit pass thrash on imagery findings | 6.5 | `max_fix_attempts: 2` per finding; existing 3-pass site cap. Both already in place; no new logic needed. |
| Cost spike from vision calls | 5.4, 6 | Tag `vision_call: true` in llm_call_log. Monitor costs in week 1 of Phase 6 rollout. |

---

## Open items / future phases (not part of this plan)

- Per-section/per-component imagery (richer planner output) — the deferred granularity question. Worth revisiting once Phases 0-6 are stable and we have enough data to know which sections genuinely need their own imagery.
- Multi-provider image generation (Banana Pro, Midjourney, FLUX). Separate plan.
- Icon library and infographic generator — separate plans, may share scaffolding with the imagery_plan work above.
- Per-vertical LoRA fine-tunes — pre-existing plan in `018_canine_biology.md`, unblocked by Phase 5's vision work but not dependent on it.
- Vision-based hero variant deduplication ("the same image is on three pages") — natural extension of Phase 6, low priority until per-section imagery exists.

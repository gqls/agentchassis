# PLAN — Product illustration pipeline

Sibling to `PLAN_imagery_loop_closure.md`. Phase 2G-ish in that doc's terms.

## Why

When we adopt a site or ingest an affiliate feed, products typically arrive with
scraped product photos in `affiliate_products.cached_image_url`. Reusing those
photos on our generated sites carries copyright/trade-dress exposure. The goal:
generate our own stylised illustrations that represent each product accurately
enough to be useful, in a consistent style that's clearly a derivative work
rather than a facsimile of the source photo.

This also handles the case where a product has *no* available image — the
illustration becomes the only image we have.

## What's in place

- `affiliate_products.custom_image_id` is an existing FK column to `assets`.
  Wired but no pipeline populates it.
- `assets.asset_key` (Phase 2B) supports multiple images per site.
- `image-build-handler` (Phase 2E/F) handles the prompt → generate → store →
  deploy chain for variant images.
- Discovery check pattern (Phase 1) is established for emitting work items.
- The "every agent is an orchestrator, delegates via spawn+call" pattern is
  the project standard.

## Shape

Five small pieces:

1. Discovery check that emits `needs_product_illustration` work items.
2. New work item type, routed to a new handler.
3. New agent `product-illustration-handler` (thin wrapper).
4. Renderer change to prefer `custom_image_id` over `cached_image_url`.
5. A prompt-construction step (template assembly initially; LLM-call later
   if needed).

No schema changes. No image-build-handler changes. New code is bounded.

## Component design

### 1. Discovery check `check_product_without_custom_illustration`

A Go file in the discovery-checks directory, ~80 lines following the
existing pattern (`check_unfulfilled_image_prompt.go`, etc.).

Detects affiliate_products rows where:
- `status = 'active'`
- `custom_image_id IS NULL`
- `site_id = $1` (the site being checked)
- Not currently claimed (no in-flight work item with same dedup key)

Emits one work item per missing illustration:

```
item_type: 'needs_product_illustration'
handler_agent: 'product-illustration-handler'
item_key: 'needs_product_illustration:<product_id>'
spec: {
  "product_id": "<uuid>",
  "purpose": "product",
  "asset_key": "product_<external_id>"
}
```

The work-item `spec` is intentionally minimal — the handler reads the
product row fresh rather than relying on snapshotted data in the spec. This
follows the "spec is a pointer; details come from DB" pattern used by
ensure_site_record.

**Per-pass cap.** When affiliate feeds drop in 500 products at once, emitting
500 work items in a single discovery pass would flood dispatch and rate-limit
us at Stability. The discovery check needs a per-pass cap (e.g. 20 items per
site per run); the remaining work picks up in subsequent passes once the
improvement loop is on. **Open question:** should this be a Go-side constant
in the check, or honour the same `cycle_budget` infrastructure mentioned in
doc 029 (if it exists)?

### 2. Work item routing

Mechanical change to the classifier — register `needs_product_illustration`
as routing to `product-illustration-handler`. Same pattern as
`unfulfilled_hero_variant` routing to `image-build-handler` (Phase 2E).

### 3. Agent `product-illustration-handler`

Workflow (8 steps, all reusing existing actions):

```
ensure_site_record
  → load_product
  → check_already_done       (conditional: skip if custom_image_id was just set)
  → construct_prompt
  → spawn_image_builder
  → call_image_builder
  → link_asset_to_product
  → complete
```

Step-by-step:

- **ensure_site_record** — existing action. Loads site, gives us `site_id`
  and `domain`.

- **load_product** — `query_database` action. Read `affiliate_products`
  by id. Output fields: `cached_name`, `cached_description`, `category`,
  `tags`, `cached_image_url`. The cached fields are LLM-readable source
  material for the prompt.

- **check_already_done** — conditional. If `custom_image_id IS NOT NULL`
  (another pass already handled this), skip straight to `complete`.

- **construct_prompt** — initially a template-assembly step (action TBD —
  either a new lightweight Go action, or use the existing `execute_llm_prompt`
  with a "compose a product illustration prompt from these fields" prompt).

  **Decision worth making upfront:** template-assembly or LLM-composition for
  v1?
  - Template-assembly: cheaper, deterministic, no LLM call needed. But less
    sensitive to product specifics ("car polishing tool" produces a generic
    prompt). Probably enough for v1.
  - LLM-composition: better prompts. Adds cost (one LLM call per product).
    Right for v2 once we see what template output looks like.

  I'd start with template-assembly. Shape:

  ```
  Stylised illustration of {cached_name}: {cached_short_description}.
  Style: {imagery_direction from site_specs}.
  No brand markings. No logos. No text on the product itself.
  Soft, clean composition. {category_specific_hint if any}.
  ```

  Reads `imagery_direction` from `site_specs.design_intent` the same way
  `getImageryDirectionForSite` already does in `generate_image_actions.go`.

- **spawn_image_builder** — `spawn_agent` for `image-build-handler`. Same
  pattern as `spawn_image_gen_variant` in image-build-handler's own workflow.

- **call_image_builder** — `call_agent`. Passes:
  ```
  input_mapping: {
    domain: "site_record.domain",
    site_id: "site_record.site_id",
    item_type: "unfulfilled_hero_variant",        # reuse this branch
    "spec.purpose": "product",
    "spec.asset_key": "input_data.spec.asset_key", # product_<external_id>
    "spec.prompt": "construct_prompt.prompt"
  }
  ```

  The image-build-handler variant chain already does what we need: store_asset
  + spawn_asset_deployer + call_asset_deployer with the per-variant deploy
  path derived from asset_key. We're using `item_type=unfulfilled_hero_variant`
  because that's the workflow branch we want, even though semantically this
  isn't a hero. Two options to make this cleaner:

  - **(a) Use `unfulfilled_hero_variant` as-is.** Works today. Slightly
    misleading naming. Practical.
  - **(b) Add a new branch in image-build-handler for `needs_product_illustration`.**
    More work; replicates the variant branch with `purpose=product`. Cleaner
    semantically.

  Option (b) is the "right" answer but option (a) ships first. **Open question
  for the design review:** which?

- **link_asset_to_product** — a small new action (~30 lines of Go). After
  image-build-handler completes, we have the new asset's `id` in the
  collected_data. UPDATE `affiliate_products.custom_image_id` to that id
  WHERE `affiliate_products.id = $1`.

- **complete** — existing action.

### 4. Renderer preference

Wherever product components currently read `cached_image_url`, change the
priority to:
1. `custom_image_id` → resolve via `assets.url` (presigned URL fresh, or
   `/assets/images/product-<key>.jpg` once deployed to git).
2. `cached_image_url` as fallback only.

I don't yet know which component / renderer file handles this — needs a
grep before this part can be written up properly. Worth a separate quick
investigation.

### 5. Stylisation as a constraint, not a parameter

For copyright safety, the prompt template (step `construct_prompt`) hard-codes:

- "Stylised illustration" framing — not "photorealistic"
- "No brand markings, no logos, no text on the product itself"
- A style direction sourced from `site_specs.design_intent.imagery_direction`
  if set, defaulting to something like "warm, hand-drawn vector style"
  if not

These aren't user-configurable knobs in v1. The handler refuses to
generate a photorealistic facsimile. If a user genuinely wants
photorealistic product photos (e.g. they own the product and have rights),
that's a different code path that doesn't go through this handler.

## Discovery rhythm

Once the improvement loop is back on:

- New affiliate_products rows → next discovery pass picks them up
- Discovery pass emits N items (capped per pass)
- Dispatch claims items, routes to product-illustration-handler
- Handler spawns image-build-handler, deploys to git
- `custom_image_id` set on product row
- Renderer now prefers the illustration

A successful pass for one product takes about 30-60s (image generation is
the slow step). At 20 products/pass and one pass per audit cycle, a site
with 500 products takes ~25 audit passes to fully illustrate. That's
roughly a week if the loop runs daily. Acceptable for v1; can be tuned.

## What we're not doing in v1

- **Multi-model selection.** Stability SDXL only. Banana Pro / FLUX /
  Midjourney is the wider "model router" piece that PLAN section 9 of
  the assessment doc calls out.
- **Img2img using the cached photo.** This is technically a stronger
  approach (the cached photo as visual reference, our generation as the
  output) and arguably a *cleaner* derivative-work argument. But it
  requires the adapter to support img2img, which it doesn't today. v2.
- **Regeneration on style change.** If `imagery_direction` is updated
  later, existing custom_images don't refresh. Manual trigger or future
  audit-loop work covers this.
- **Audit of generated illustrations.** Phase 6 of the imagery loop
  (imagery-quality-auditor) covers this once it lands.
- **Anti-copyright vision check.** No "does this illustration look too
  much like the source photo" check. The stylisation constraint is our
  only protection in v1. Worth considering for v2 if we observe
  problems.

## Open design questions for review

1. **Template-assembly or LLM-composition for the prompt?** Above.
2. **New image-build-handler branch (`needs_product_illustration`) or
   reuse `unfulfilled_hero_variant`?** Above. Naming hygiene vs ship-now.
3. **Per-pass cap location: Go constant or `cycle_budget` infrastructure?**
   Depends on whether `cycle_budget` is already in production (doc 029
   reconciler work).
4. **`asset_key` shape for products: `product_<external_id>` or
   `product_<uuid>`?** External_id is human-readable in URLs but not
   guaranteed unique across programs. UUID is unique but ugly in
   filenames. Probably `product_<external_id>` with a uniqueness check.
5. **What happens when cached_image_url updates upstream?** Affiliate
   feeds re-fetch periodically. If the cached image changes, should we
   invalidate the custom_image and regenerate? Probably not in v1 — the
   custom_image is canonical once made — but worth a `last_regenerated_at`
   timestamp somewhere so future invalidation logic has a hook.

## Sequencing within this plan

The five components have these dependencies:

```
discovery check ── emits ──> work item type
                                 │
                                 ▼
                         product-illustration-handler
                              │     │
                              │     └── delegates to ── image-build-handler
                              │                            (no change today)
                              ▼
                     link_asset_to_product action
                              │
                              ▼
                  renderer preference for custom_image_id
```

Order to build:
1. `link_asset_to_product` action (small, isolated).
2. `product-illustration-handler` agent definition (workflow only; depends
   on (1)).
3. End-to-end test on a hand-inserted affiliate_products row, manually
   triggered. Bypasses discovery and dispatch.
4. Discovery check (depends on (2) being live so the work items have
   somewhere to go).
5. Renderer preference. Independent; can be done in parallel with any of
   the above.

Each step is independently testable.

## Acceptance criteria

The minimum end-to-end test once all five components are in place:

1. Manually insert an `affiliate_products` row for robot-hands.com with
   `cached_name='Pneumatic parallel-jaw gripper'`, a description, and
   `custom_image_id=NULL`.
2. Trigger product-illustration-handler directly with `product_id=<that row>`.
3. Observe:
   - product-illustration-handler spawns image-build-handler
   - image-build-handler spawns image-generator (Stability call)
   - Image lands at `assets/images/product-<external_id>.jpg` in the
     robot-hands.com repo
   - `affiliate_products.custom_image_id` is now set
   - All three orchestrations COMPLETED

That's the test. Once it passes, discovery check + dispatch wire it up
to run automatically.

---

## Effort estimate

About the same scale as Phase 2E. Mostly new code (one action, one agent,
one discovery check, one small renderer change). No schema changes. Reuses
the variant pipeline we just built. Probably 2-3 implementation sessions of
focused work, plus design review on the open questions above.

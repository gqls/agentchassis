# Phase 0 Bundle — README

Three deliverables that ship together as the imagery-loop-closure Phase 0:

| File | Purpose |
|---|---|
| `phase_0_revised_generate_image_actions.diff` | **Replaces** the earlier `phase_0_1_generate_image_actions.diff`. Same shape, with the 200-char SDXL-aware cap and unlabeled prompt format. |
| `phase_0_2_store_asset_action.diff` | Fixes a pre-existing bug (origin_prompt silently dropped) and adds origin_model population. |
| `phase_0_combined_migration.sql` | Single transaction covering the agent-definition changes for both phases. Replaces both earlier SQL files. |

## What ships in this bundle

### Phase 0.1 (revised)

- `image-generator` reads `site_specs.design_intent.imagery_direction` and prepends
  it to the prompt before sending to Stability. Direction is enrichment — empty/missing
  is graceful.
- 200-char cap on the direction half, sentence-boundary-preferred truncation,
  no `Style direction:` / `Subject:` labels (those eat ~17 SDXL tokens for nothing).
- The cap is documented as SDXL-specific. Future provider routing will set
  per-provider caps in the same helper.

### Phase 0.2

- **Bug fix.** `StoreAssetAction` now writes `origin_prompt` from the
  `origin_prompt_field` config that workflows have been silently passing for some
  time. Today every row in `assets` has `origin_prompt = NULL` despite the
  workflows declaring otherwise. After this ships, new generations populate it.
- **New.** `assets.origin_model` populated. Accepts `origin_model` (literal — used
  today, value `"sdxl"`) or `origin_model_field` (path — for when provider routing
  is added later). Literal wins if both are set.
- **Data quality refinement.** `site-work-orchestrator` and `pageflow-builder`
  switch their `origin_prompt_field` from the un-composed plan prompt
  (`site_plan.image_prompts.hero_home`) to the composed prompt the model actually
  saw (`hero_result.prompt`). `image-build-handler` was already correct.

### What was found in passing (not addressed in this bundle)

These came up while writing the changes; recording so they don't get lost.

- The `assets` UNIQUE constraint on `(site_id, purpose) WHERE purpose IS NOT NULL`
  remains in place. Multi-image-per-purpose is PLAN section 2.3, scheduled for the
  asset locking phase. Untouched here.
- `origin_url` is unused for generated images. When provider routing lands, it
  becomes useful for tracking which model/provider URL was actually called. Out
  of scope for now.
- `alterations jsonb` is unused. Becomes useful for img2img / variant chains.
  Out of scope for now.
- `attribution`, `license`, `license_url` are populated by neither the generated
  path nor the (planned) adoption-mirror path. PLAN section 3.1 is the right
  place to wire these for adopted imagery.

## Apply order

Strict order across the bundle:

1. **Apply the SQL migration first.** Everything in section 1 (site_id passthrough)
   is dormant until the Go change ships — `inputData["site_id"]` is just an unread
   key. Sections 2 and 3 (origin_model literal, origin_prompt_field rewire)
   similarly do nothing until the Go action reads them.
   
   So the SQL change is independently safe and can land any time before the Go
   changes.

2. **Apply both Go diffs in the same release.** They're independent (different
   files, different actions, no shared call paths) but should ship together so the
   verification queries return useful data on the next build.

3. **Verify.** Two queries:

   ```sql
   -- Confirm Phase 0.1 prompt composition (assumes a recent build with
   -- design_intent.imagery_direction populated):
   SELECT purpose,
          origin_model,
          LEFT(origin_prompt, 280) AS prompt_preview
   FROM assets
   WHERE site_id = '<recently built site>'
   ORDER BY created_at DESC;
   ```
   
   Expectation: `origin_model = 'sdxl'`. `origin_prompt` begins with the imagery
   direction (truncated at first sentence boundary if direction was long), then
   the subject. No `Style direction:` label.

   ```sql
   -- Confirm migration applied to all 16 paths:
   -- (verification CTE block at the bottom of the SQL file, uncomment to run)
   ```

## Rollback

The rollback block at the end of the SQL file reverts all three sections in one
transaction. If only the Go change needs to revert, deploy the prior binary —
the migration's added fields become dead config, no behaviour change.

## What this enables next

After this bundle, downstream work has a clean foundation:

- **Phase 1 of PLAN_imagery_loop_closure** (algorithmic discovery checks) can
  query `assets.origin_prompt` and `assets.origin_model` to answer
  "what was actually generated and how" without depending on logs.
- **Phase 4 of PLAN** (visual auditor extension) can compare the deployed
  imagery against `origin_prompt` to detect mismatches.
- **Phase 1 step 7 of doc 030** (brief renderer) can be wired into
  `getImageryDirectionForSite` to compose strategic + plan-time directives
  once `site_plan_directives` exists. Tracked as a successor work item;
  the current single-source read is correct for today's data.

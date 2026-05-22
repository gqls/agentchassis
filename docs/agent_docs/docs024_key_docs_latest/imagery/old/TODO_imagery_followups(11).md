# TODO — Open items as of 2026-05-15 end of session

Single reference for everything that came out of the Phase 2G+2H verification work. Organised by area, prioritised within each area. Cross-references back to where each item was first discussed.

---

## State of play

**Phase 2G + 2H: pipeline verified end-to-end at scale on 2026-05-15.** Seven hero items processed through the full discovery → audit → triage → dispatch → handler → asset → git deploy chain on robot-hands.com without manual intervention beyond initial unsticking. Each ~25-30s through the handler. All seven assets committed to git, visible at `robot-hands.com/assets/images/hero-*.jpg`.

**Pipeline-correct but output-quality issues remain on robot-hands.com:**
- **Icon completed but is visually wrong (item 23).** The asset shipped but doesn't match the prompt — 6-panel photorealistic 3D robot arms instead of flat single-icon line art. Asset is unshippable as-is.
- **Icon has wrong purpose field (item 22).** `purpose='hero'` rather than `purpose='icon'`. Hygiene bug downstream of store_asset.
- **Services page has no hero image (item 21).** Planner didn't include it in `site_plan_imagery`. By design or gap, unresolved.
- **Page rerender ran successfully** — services.html rendered with its current component (gradient-only `hero-services`, no image).

**Decision: stay on robot-hands.com.** finetuning.uk verification deferred until robot-hands.com output is shippable. Verifying the pipeline at scale against a second site is wasted effort if the first site's output doesn't meet the quality bar.

**Next session entry point:** item 23 (icon visual quality) is the priority — biggest visible gap. Items 22 and 21 can be done in parallel.

---

## Imagery pipeline (Phase 2G/2H follow-ups)

Direct outputs of the verification work. Sized small-to-medium, none are blockers.

### 1. State machine corruption on failed items
**Severity:** medium. Hygiene bug; doesn't block work, but makes diagnosis confusing.
**Symptom:** the icon work item ended up `status=triaged` with `claimed_at=2026-05-15 14:50:47` and `claimed_by=build-dispatch-loop` populated. That's an invalid state combination — `triaged` should mean "not yet claimed" with NULL claim metadata.
**Cause:** some failure-recovery path resets status without clearing the claim metadata. Probably `mark_work_item_failed` or its caller in image-build-handler's error_step.
**Fix:** trace what UPDATE-ed the row when the image-generator failed at 512×512. Ensure claim metadata is cleared on any non-claimed status transition.

### 2. `updated_at` trigger not firing on site_work_items UPDATEs
**Severity:** medium. Makes the table's audit trail unreliable.
**Symptom:** icon row shows `updated_at=2026-05-13 18:54:34` despite multiple recent UPDATEs (`claimed_at` set on 2026-05-15, `result` shows entries from 2026-05-15).
**Cause:** either the trigger doesn't exist, or it's column-filtered to only fire on `status` changes (and our UPDATEs were of other columns).
**Fix:** check for `update_site_work_items_updated_at` trigger; ensure it fires on any UPDATE.

### 3. Icon dimensions — patch verification or second override
**Severity:** medium. Blocking the icon from completing.
**Status:** user says the icon_dimensions_patch is deployed (kindDefaults["icon"] now 1024×1024). But the failure at 14:50 logged 512×512 to Stability. Two possibilities, both worth checking:
  - The chassis pod restart at ~14:30 ran on a pre-patch image (since pod age says 27min but doesn't confirm image tag)
  - There's a second injection path for width/height that overrides kindDefaults
**Fix:** verify the chassis container image tag (`kubectl describe pod ... | grep -i image:`). If correct, reset the icon and trigger; watch the next orchestration's `collected_data->'request_body'` to see what dimensions were actually sent.

### 4. `constraints.aspect` vs `style_hints.aspect_ratio` schema/code mismatch
**Severity:** low. Doesn't break anything today, but creates author confusion.
**Symptom:** the planner LLM was emitting aspect ratio under `spec.constraints.aspect`. The code only reads `spec.style_hints.aspect_ratio`. So constraint hints from the planner get silently ignored for dimension purposes (they ARE used for negative prompt).
**Decision (2026-05-15):** rename in the planner prompt rather than make Go honour `constraints.aspect`. The prompt now teaches `style_hints.aspect_ratio`, matching what the existing Go code reads. See `planner_prompt_patch_imagery.md` for the exact changes.
**Hardening deferred:** later we may want `constraints` to be functionally honoured by the generation pipeline (per-provider validation, content safety modes, transparency flags). For now `constraints` is documented as "informational only, reserved for future use" in the planner prompt. When a real use case arrives — likely alongside multi-provider routing (item 24) — the Go-side change is small: extend `generate_image_actions.go` to read from `constraints` as a fallback to `style_hints`.

### 5. `parseAspectRatio` SDXL v1.0 whitelist validation
**Severity:** HIGH (elevated 2026-05-18 — actively blocking hero generation).
**Symptom (2026-05-18):** with the planner prompt now emitting `style_hints.aspect_ratio: "16:9"` for hero entries, `parseAspectRatio` produces `1024×576`, which SDXL v1.0 rejects (allowed dimensions don't include 1024×576). The function snaps to multiples of 64 (a generic SDXL constraint), not the strict SDXL v1.0 whitelist (1024×1024, 1152×896, 1216×832, 1344×768, 1536×640, 640×1536, 768×1344, 832×1216, 896×1152).

**Regression cause:** previous hero generation worked because the planner did NOT emit aspect_ratio hints — fell through to `kindDefaults["hero"]` which produces a valid SDXL size. Adding the aspect_ratio hint as part of the item 4 prompt patch enabled an existing bug path.
**Fix:** in `generate_image_actions.go`, after computing width/height in `parseAspectRatio`, snap to the nearest dimension pair from the SDXL v1.0 whitelist. Match the requested aspect direction (portrait/landscape/square) when choosing.
**Test case:** input `aspect_ratio="16:9"`, anchor=1024 → currently produces 1024×576 (fails), should produce 1344×768 (nearest valid landscape on whitelist).

---

## Dispatch architecture (FOCUS_dispatch_diagnostic.md Q3/Q4)

Larger improvements to the dispatch chain, surfaced by this session's deep dive.

### 6. Watchdog: auto-reset stuck claimed items
**Severity:** medium (user confirmed 2026-05-18 — existing reaper covers this, just runs infrequently).
**Mechanism:** a single `claimed` item with no progress excludes its entire site from dispatch indefinitely via the `NOT EXISTS` clause in `find_dispatchable_site`.
**Status:** an existing reaper handles this concept already; cadence is the only gap. No need to build a new watchdog — bump the reaper's frequency when convenient.
**Concrete proposal if cadence becomes a problem:** a scheduled task running every 5 minutes:
```sql
UPDATE site_work_items
SET status = 'triaged',
    claimed_at = NULL,
    claimed_by = NULL,
    attempt_count = attempt_count + 1,
    error = COALESCE(error || E'\n', '') || 'Auto-reset by zombie-claim watchdog at ' || now()::text
WHERE status = 'claimed'
  AND claimed_at < now() - interval '15 minutes'
  AND attempt_count < max_attempts;
```
Deferred — existing reaper is sufficient for current load.

### 7. Pipeline field — fix discovery emission + loosen dispatcher
**Severity:** medium. Latent bug that bites whenever a discovery check writes a non-`build` pipeline.
**Decision recorded in Q4 (2026-05-15, with user):** leave the field as soft, currently-unused routing label.
**Two concrete changes:**
  - Fix `unfulfilled_imagery_plan` (and any sibling checks) to NOT explicitly write `pipeline='design'`. Let schema default ('build') apply.
  - Loosen `build-dispatch-loop.load_items.item_pipeline` config to accept any value (or drop the filter).

**Status update (2026-05-17):** confirmed the bug still bites. After applying the planner prompt patch and re-running discovery on robot-hands.com, all 8 emitted `needs_imagery` work items came through with `pipeline='design'`. Required a manual UPDATE to `'build'` before audit/dispatch could process them. **The Go fix must land before declaring 2G+2H done** — otherwise every subsequent site (including finetuning.uk) will hit the same blocker and need the same UPDATE.

Workaround that's been used twice now:
```sql
UPDATE site_work_items
SET pipeline = 'build'
WHERE site_id = '<site>'
  AND item_type = 'needs_imagery'
  AND status = 'detected'
  AND pipeline = 'design';
```

**Two-part resolution scoped (2026-05-17, pending deploy):**
- **Part A — hardcode 'build' in `check_unfulfilled_imagery_plan.go`.** Patch: `check_unfulfilled_imagery_plan_pipeline_patch.go`. The check now writes `Pipeline: "build"` directly rather than reading `dctx.Pipeline` (which resolves to "design" under `design-discovery-agent`). Rationale: pipeline is the destination, not the origin — and the destination handler (`image-build-handler`) is build-side.
- **Part B — remove the `item_pipeline` filter from `build-dispatch-loop.load_items` config.** Migration: `loosen_build_dispatch_pipeline_filter.sql`. The dispatcher now loads triaged items regardless of pipeline value. Decoupling discovery from dispatch routing: if any current or future check writes a non-"build" pipeline value, dispatch still picks it up.

**Belt-and-braces rationale:** Part A fixes today's bug at source. Part B prevents the same bug shape from appearing in any future discovery emission. Cheaper than chasing each new check that gets added.

**Pipeline field's future:** stays as soft routing label. If multi-pipeline dispatchers are introduced later (different cadences, concurrency limits, error tolerances), the filter can be re-introduced selectively on those dispatchers — but the default `build-dispatch-loop` no longer enforces it.

**Open follow-up:** verify after deploy that the `load_items` action treats missing config as "no filter". If dispatch loads zero items consistently after Part B, the action's Go source needs review — set `item_pipeline: ""` explicitly or look at the load_items source.

### 8. Outer ORDER BY in find_dispatchable_site
**Severity:** low. Affects fairness, not correctness.
**Symptom:** when multiple sites have eligible items, the `LIMIT 1` selects arbitrarily (Postgres scan order). Some sites might wait longer than others purely by UUID lexical position.
**Fix:** add an outer `ORDER BY oldest-pending-item ASC` or similar fairness metric.

### 9. build-pipeline-trigger writes orchestration_states
**Severity:** low. Makes debugging dispatch decisions harder.
**Symptom:** the trigger runs every 30s but doesn't appear in `orchestration_states` (not in the 34 known `owner_agent_type` values). Its decisions are invisible to post-hoc auditing.
**Fix:** trigger should write a row per invocation with the selected site_id (or none) and the reason.

### 10. Per-item-type circuit breaker
**Severity:** medium. The `needs_design_review` failures starved all imagery work on robot-hands.com for hours today.
**Mechanism:** when an item type fails N times in a row on a site, exclude it from `find_dispatchable_site` consideration for that site for a cooldown period. Lets other item types make progress.
**Fix:** new column or table tracking per-site-per-type failure rate; SQL clause in `find_dispatchable_site` excluding sites where ALL pending items are of a recently-failing type.

### 11. needs_design_review handler reliability
**Severity:** medium. Currently a recurring source of zombies on robot-hands.com.
**Status:** the audit run created `needs_design_review` items that consistently fail with "Claim timed out — handler pod likely died". Pattern suggests systematic crash, not transient.
**Fix:** separate investigation needed. Not blocking imagery work — can be punted to `failed` manually if it gets in the way again.

---

## Reconciler / coupling gaps

Closing feedback loops in the pipeline.

### 12. Asset → page rerender coupling
**Severity:** medium. Visible to users — pages show old heroes/no heroes even after assets are generated.
**Symptom:** seven new hero assets generated 1 hour ago. Pages on the site last rendered 2-3 days ago. No automatic rerender was triggered.
**Fix:** when `store_asset_action` inserts a new row, find pages referencing that asset_key (via `content_data.*_url` or rendered HTML) and insert a `needs_rerender` work item for each. Closes the loop.

### 13. Phase 3 — adoption image mirror (deferred)
**Severity:** low. Documented in `PLAN_imagery_loop_closure.md`. Adopted sites lose their crawled imagery; the mirror would persist them as reference imagery for img2img generations.
**Status:** deferred — not blocking the build pipeline.

### 14. Variant chain site_id pass-through
**Severity:** low. Edge case, surfaced earlier in the session.
**Symptom:** when variant chains spawn child orchestrations, site_id sometimes doesn't propagate through `input_mapping`, causing `render_js_snippets` to fail with "site_id not found at site_record.site_id".
**Fix:** confirm the chain's input_mapping includes `site_id`; trace any chain that loses it.

---

## Content quality (robot-hands.com observations from rendered HTML)

Not imagery work, but noticed during verification. Worth a separate cleanup pass.

### 15. Duplicate learning pages
Four pages with overlapping names: `learning-center.html`, `learning-center-index.html`, `learning-center-hub.html`, `learning-center-article.html`. Probably accumulated from discovery iterations. Footer lists three of them as "Learning Center" — confusing for users.

### 16. Duplicate selection guide pages
`selection-guide.html` and `gripper-selection-guide.html` — both linked from footer as "Selection Guide".

### 17. Empty `tools.html`
Listed in main nav and footer. Body is empty in the rendered HTML.

### 18. Broken footer contact display
Footer has `<a href="mailto:"></a>` and an empty `<p></p>` where contact details should be. Either `sites.email` / `site_specs.identity.email` is empty, or the renderer isn't reading them.

These are operator-level cleanups, not pipeline work. Worth a separate "site hygiene" pass when convenient.

---

## Findings from rerender pass (2026-05-15)

### 21. Services page hero coverage
**Severity:** low-medium. Quality observation, not a bug.
**Symptom:** Services.html renders with a gradient-only `hero-services` component, no image. This was NOT emitted as a `needs_imagery` work item by Phase 2G discovery — likely because the planner didn't list it in `site_plan_imagery` for the services page. The query confirms: 7 page-scope hero entries exist (about, contact, how-it-works, index, matchmatrix-methodology, selection-guide, tools), but no entry for services.

**Three sub-questions:**
- a) Is the planner's logic for "which pages get hero images" intentional, and if so, should services be on the list?
- b) Does the component database have an image-variant `hero-services` that the planner could have requested?
- c) Should the `unfulfilled_imagery_plan` check be extended to flag pages that have *no* hero imagery declared at all (a sub-class of "missing"), independent of whether the plan asks for it?

Worth investigating after the finetuning.uk verification.

### 22. Icon asset has wrong `purpose` field
**Severity:** medium. Asset is wired up correctly per `asset_key`, but `purpose='hero'` instead of `purpose='icon'`. Any downstream code that filters or routes by `purpose` will treat this icon as a hero.
**Symptom:** icon_cross_technology asset row at 2026-05-15 15:55:36 has `purpose='hero'`. Expected `purpose='icon'`. The `kind` field in the work item spec was 'icon', so this is a translation step that's failing.
**Most likely cause:** the purpose_field Go change (Phase 2H follow-up) reads from `spec.purpose` to populate the assets row. Either:
  - The work item spec's `purpose` field said 'hero' (planner didn't differentiate by kind)
  - The store_asset_action's purpose-resolution logic defaults to 'hero' when not explicitly set
  - The Phase 2H purpose_field migration didn't fully cover this path
**Fix:** trace store_asset_action's purpose resolution; ensure it reads from `spec.kind` as a fallback to `spec.purpose`, or fix the planner to set `purpose` correctly. Sibling consideration: should `purpose` even be a free-text field separate from `kind`, given they appear to encode similar information?

### 23. Icon image visually wrong — not an icon
**Severity:** high. The pipeline produced and shipped an asset, but the asset doesn't match what was requested. Worse than a hygiene bug — the output is unusable.
**Symptom:** `icon_cross_technology` was generated as a 6-panel composition of photorealistic 3D robotic-arm assemblies with shadows on a grey background. The prompt asked for "minimal geometric icons … single-weight line art in cyan on transparent background." Constraints specified `transparent_background: true` and `aspect: 1:1`.
**What went wrong:**
  - **Style:** SDXL doesn't naturally produce flat line-art when the prompt also references concrete subjects ("gripper", "robotic", "industrial"). It defaults to photorealistic.
  - **Composition:** No instruction prevented SDXL from producing a multi-panel grid layout. The prompt described "a set of six … icons" which SDXL interpreted as one image containing six panels rather than six separate images.
  - **Transparency:** `constraints.transparent_background: true` is silently dropped — constraints feed only negative_prompt (per `generate_image_actions.go` line 252), and "transparent_background" doesn't translate to a useful negative phrase.
  - **Subject specificity:** Six distinct gripper types (parallel jaw, servo motor, horseshoe magnet, suction cup, flexible finger, adhesive pad) got merged into "generic robot arm assemblies" — SDXL doesn't have strong concept anchors for these specific mechanisms.

**Several layered fixes possible:**
- **Prompt-side (cheapest):** rewrite icon prompts to emphasise "vector illustration", "flat 2D", "single icon", "no background"; add strong negative prompts for "photorealistic, 3D, multiple, grid, panels, photo"; explicitly ask for ONE icon per generation if multi-element sets are needed downstream.
- **Provider-side:** SDXL isn't the right model for clean icons. FLUX is better at flat illustration; specialist icon generators (or just LLM-generated SVG) are better still. This is the multi-provider routing question from the original imagery assessment.
- **Schema-side:** the `kind='icon'` distinction should imply a different generation pipeline (different model, different default prompt scaffold) rather than just different dimensions. Currently kindDefaults only differentiates by Width/Height/CfgScale/NegativePrompt; the prompt itself flows through unchanged.
- **`transparent_background` constraint plumbing:** if constraints want to influence generation, they need a path beyond negative_prompt. Could become explicit fields on the request (provider-side support varies), or a post-generation alpha-extraction step.

**This is the single biggest quality gap on robot-hands.com.** Heroes look acceptable; the icon is unshippable. Worth focused investigation before declaring robot-hands.com done.

**Status (2026-05-15):** Fix A is scoped and ready. The planner prompt analysis surfaced four concrete changes (rename aspect, add decomposition rule, strengthen prompt construction guidance, replace worked example, add Rule 16). See `planner_prompt_patch_imagery.md` for the patch and application notes. After application, re-plan robot-hands.com and discovery will re-emit per-actuation-type icon work items. If the icons still come out wrong after this, item 24 becomes the next step.

**Verdict (2026-05-18, after running Fix A end-to-end on robot-hands.com):** **Fix A is insufficient.** The planner prompt rewrite produces correct prompts asking for "A single minimalist flat icon ... line illustration, geometric, sharp corners, single dark colour on plain background, no shadows, no photorealism" — and SDXL ignores the style instructions almost entirely. Sample outputs:
- `icon-cycle-time.jpg` → photorealistic 3D rendering of a gauge cluster with shadows on grey background. Not flat. Not line illustration.
- `icon-matchmatrix.jpg` → multi-panel composition of three 3D robotic arm assemblies. Still multi-panel despite the explicit "A single minimalist flat icon" prompt prefix.

**Diagnosis:** SDXL has a strong bias toward photorealistic output when subjects are concrete and identifiable, and a weak ability to follow style-only instructions that contradict that bias. The decomposition worked structurally (one subject per work item, distinct keys) — the style instructions did not.

**This forces a model question.** SDXL is the wrong tool for flat-vector icons. Three real options ranked:
- **(a) Accept SDXL's photorealistic icons as current state.** They're not flat-vector but they're consistent in style and they do represent the right concepts. Ship them; plan migration later. Lowest effort.
- **(b) Plumb reference-image (item 24, IP-Adapter path).** Provides style anchor. Bigger work; requires multi-provider routing since Stability's standard REST endpoint doesn't expose IP-Adapter.
- **(c) Switch icon generation entirely to a different path.** Two sub-options:
  - **c1.** FLUX-schnell via a new generator agent. Better at flat illustration than SDXL; raster output still.
  - **c2.** LLM-generated SVG directly. No raster image generation at all. Crisp at any size. The natural format for icons in any case.

**Recommendation:** **(c2)** — LLM-generated SVG. Icons should be SVG, not JPG. The whole image-generation pipeline is overkill for them. An LLM (Claude, GPT-4) can produce valid SVG markup that's actually flat-vector — because it IS vector. Bypasses the entire "convince SDXL to do flat illustration" problem. Heroes stay on SDXL where it excels.

**Architecturally:** option c2 implies a per-kind generation pipeline routing — `kind='icon'` goes to an LLM-SVG path, `kind='hero'` goes to SDXL, etc. That's the "kind-specific generation pipeline" from item 23's earlier analysis, now with a concrete answer for icons.

**Status: item 23 now blocked on architectural decision between options a/b/c2.** Heroes are unblocked once item 5 lands. Icons can ship today as photorealistic-but-coherent (option a) while the architectural call gets made.

**Decision (2026-05-18):** Path 2 chosen — plumb reference-image (item 24) AND switch icon generation to a different model. SDXL is wrong-tool for flat illustration; an additional provider is required.

**Model comparison summarised:**

| Model | Quality (flat illust.) | Reference image | Cost | Open weights |
|---|---|---|---|---|
| SDXL v1.0 | poor | needs IP-Adapter | $0.02 | yes |
| FLUX.1 schnell | good | via IP-Adapter | $0.003 | yes |
| FLUX.1 dev/pro | very good | mature IP-Adapter support | $0.025-0.05 | yes (dev), no (pro) |
| DALL-E 3 | very good | no native | $0.04-0.08 | no |
| Nano Banana Pro 2 | very good | native multimodal | $0.04 | no |
| LLM-SVG (Claude/GPT) | native vector | n/a | $0.001-0.005 | n/a |

**Recommendation:** FLUX.1 schnell via fal.ai as the icon model. Cheap (10× less than SDXL), open weights, mature IP-Adapter support via fal's endpoint. Alternative if integration speed matters more than cost: Nano Banana Pro 2 (native reference support, no IP-Adapter integration needed).

**Sleeper option for later:** LLM-generated SVG for icons. Icons are vector by nature; raster generation is structurally wrong for them. Worth a focused experiment once the immediate work is shipped — could replace the entire icon raster pipeline with a Claude/GPT SVG-write step. Note as a future architectural review item, not blocking current work.

### 24. Reference-image / style-anchor plumbing for kind='icon' (and beyond)
**Severity:** medium-high. The fallback for item 23 if prompt-rewriting alone doesn't produce consistent flat-vector icons. Also the foundation for consistent imagery across a site more broadly (siblings sharing a visual style).

**What it does:** lets a work item specify a reference image URL that conditions generation toward the reference's aesthetic. Technique varies by provider:
- **img2img** (Stability native): reference becomes a base image, prompt steers the modification. Cheap; reference acts as a *subject* anchor more than a *style* anchor.
- **IP-Adapter** (Replicate/ComfyUI, not on Stability's standard REST endpoint): explicitly separates style of reference from subject of prompt. Right technique for "make 6 icons that all look like THIS."
- **LoRA fine-tune** (already in FOCUS doc as future work): highest quality, needs training corpus.

**What the architecture already has:**
- `reference_image_uri` in the Phase 2H field set — "accept and log only" today; the schema is ready
- `assets.origin_asset_id` — FK for derivative tracking
- `assets.alterations` JSONB — documented as where style-transfer provenance lives
- `origin_url` — for attribution

**What's missing:**
- The image-generator adapter doesn't accept `reference_image_uri` (text-prompts only)
- The `generate_image` action reads it but logs only — doesn't pass through
- No mechanism for "where does the reference come from for this work item"

**Three provenance options for the reference, ranked by setup effort:**
- a) **Generate-one-then-derive.** Get ONE good icon by any means (prompt iteration on SDXL, FLUX, hand drawing, stock library). Use it as the style reference for siblings. Validates the whole architecture against a concrete need.
- b) **Hand-curated style library per site.** `assets` rows tagged `origin_type='style_reference'`, planner emits `reference_image_uri` pointing at these. More structure, more upfront work.
- c) **System-wide style libraries per kind.** A house style for icons across all sites. Best consistency, least flexibility per site.

**Recommendation:** start with (a) for the immediate icon problem. The plumbing built for one case generalises naturally. Validates the multi-provider routing question at the same time, since IP-Adapter may force a switch from Stability's standard endpoint to Replicate or a custom host.

**Fits with FOCUS doc deferred items:** Phase 3 (adoption-image-mirror) was specifically about persisting crawled imagery for use as img2img references. This item is the same capability, different source.

### 25. Planner work-item decomposition — one output per work item
**Severity:** medium. Structural cause of part of item 23's failure mode and a recurring fragility.

**Symptom:** the icon_cross_technology spec asked for "a set of six minimal geometric icons" in a single work item with a single prompt. SDXL produced one image containing six panels — its literal interpretation of "set of six." The pipeline assumed one work item → one image; the prompt asked for one image containing six things.

**Principle:** a work item should describe one deliverable. If the planner wants six icons, it should emit six work items each describing one icon. The generation pipeline can then process them as six separate generations (with shared style anchor per item 24).

**Fix:** at the planner level (the prompt that produces `site_plan_imagery` entries), instruct the LLM to decompose multi-element requests into individual rows. For the icon_cross_technology case, that becomes six rows:
- `key='icon_parallel_jaw'`, kind='icon', prompt='single icon: a parallel-jaw gripper, ...'
- `key='icon_servo_motor'`, kind='icon', prompt='single icon: an electric servo motor, ...'
- ... etc

The constraint is enforceable at the database level if needed: a CHECK constraint on `site_plan_imagery.prompt` rejecting plurals like "set of", "six", "multiple" — but that's heavy-handed. Prompt-level discipline is probably enough.

**This is a prerequisite for item 23's fix A.** Rewriting the icon prompt for clarity doesn't help if the prompt still asks for a set rather than a single image.

**Broader principle worth noting:** this is the same "one job = one work item" pattern that's served the dispatch system well elsewhere. It just hasn't been pushed all the way through to the planner's imagery output. Heroes work because each page gets one hero — naturally one-per-item. Icons need the same decomposition.

**Decisions made (2026-05-15):**
- **Keep the array structure for `sections.<page>:<ord>` and `pages.<page>` entries.** Multi-entry sections ARE the canonical way to express "multiple distinct images at the same scope" — the structure was right, just under-exemplified.
- **Err toward over-decomposition.** The economic argument: N too many icons costs N small generations and N rows-to-delete — bounded, recoverable, visible. N icons crammed into one image is unusable, looks superficially successful, and is expensive to detect and clean up. So the prompt biases toward producing more entries rather than fewer.
- **Concrete fix specified.** See `planner_prompt_patch_imagery.md` for the four changes (entry fields table, what-to-populate guidance, per-row prompt construction, replaced worked example) and added Rule 16.

**Cleanup of the existing icon_cross_technology artifact (do this when starting item 25):**

```sql
-- 1. Delete the asset row (the 6-panel image that shouldn't exist)
DELETE FROM assets
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND asset_key = 'icon_cross_technology';

-- 2. Expire the work item (so it won't claim again)
UPDATE site_work_items
   SET status = 'wont_fix',
       error = 'Superseded by item 25 — replaced with per-actuation-type icons'
 WHERE id = '3cacc0dd-4bb0-44a1-a3ca-87a3911423a8';

-- 3. Delete the site_plan_imagery entry (so re-planning emits decomposed entries)
DELETE FROM site_plan_imagery spi
USING site_plans sp
WHERE spi.plan_id = sp.id
  AND sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND spi.key = 'icon_cross_technology';
```

After cleanup + re-plan, discovery's `unfulfilled_imagery_plan` check will naturally detect the new per-actuation-type entries and emit one work item per icon. Also delete the deployed file from git in the same operation if applicable:

```bash
# git rm sites/robot-hands.com/assets/images/icon-cross-technology.jpg
# (verify path first; commit message: "Remove botched icon — superseded by per-type icons")
```

### 26. Planner key stability across replans
**Severity:** low-medium. Wastes generations and creates orphan assets when a site is re-planned.
**Symptom (2026-05-15):** the previous plan had a site-scope hero keyed `hero_canonical`. The new plan, generated by the same LLM with the same input data but the updated prompt, called the equivalent concept `brand_hero_canonical`. Discovery saw "no asset matches `brand_hero_canonical`" and emitted a fresh work item — even though `hero_canonical` already existed as an asset and was effectively the same image.
**Root cause:** the planner LLM has freedom in choosing `key` values. The Imagery Block guidance says "short identifier, lowercase, underscore-separated" and "Unique within its scope" but doesn't specify reproducibility across runs. Same input, different output — LLMs do this.
**Impact:** every replan of an existing site potentially regenerates imagery for concepts it already has, just because the LLM chose a different key. Orphan assets accumulate in the DB and git.
**Possible fixes, in order of effort:**
- a) **Add a "Reuse existing keys" rule** to the planner prompt. When a site already has a current plan, pass the old plan's imagery keys as a reference and instruct the LLM to reuse them where the concept is unchanged. Soft, prompt-level approach. Requires plumbing the old plan's imagery into the prompt context.
- b) **Canonical key lookup table.** A small dictionary the planner consults: "hero for the site root" → always `hero_canonical`; "icon for cross-technology section" → always `icon_cross_technology`. Removes LLM judgment for well-known slots. Brittle; doesn't help with novel concepts.
- c) **Concept-level matching at discovery time.** Discovery would not just match by exact asset_key, but also by semantic similarity (kind + scope + scope_ref + prompt similarity). Substantially more complex; requires embeddings or LLM-based matching.
**Lowest-effort entry:** (a) — extend `read_specs` step output to include the previous plan's imagery keys, then add a rule to the planner prompt. Worth doing before opening this loop on more sites.

---

## What's verified today (for the record)

- ✅ Discovery emits `needs_imagery` items at `status=detected` ✓ (Phase 2G.4)
- ✅ `design-audit-agent` triage step promotes `detected → triaged` ✓
- ✅ `build-pipeline-trigger` selects site with `find_dispatchable_site` query ✓
- ✅ `build-dispatch-loop` claims up to 5 triaged items per tick per site ✓
- ✅ `image-build-handler` runs handler workflow for `needs_imagery` items ✓
- ✅ `generate_image` produces SDXL-valid 1024×1024 hero images ✓
- ✅ `generate_image` produces SDXL-valid 1024×1024 icon images (after patch propagation) ✓
- ✅ `store_asset` writes `assets` row with correct `asset_key` and `origin_model='sdxl'` ✓
- ✅ `deploy_image_asset` optimises (1600×900 for hero) and commits to git ✓
- ✅ `mark_work_item_complete` transitions items to `status=complete` ✓
- ✅ `rerender-pages` re-renders pages with current component selection ✓
- ⚠️ `store_asset` purpose field set incorrectly for icon (got 'hero', expected 'icon') — see item 22
- ❌ Asset → page rerender coupling — needs reconciler (item 12). Pages don't auto-rerender when new assets land
- ❌ Planner doesn't request hero imagery for services page — by design? gap? — see item 21

---

## Sequence for next session

1. **Stay on robot-hands.com until it's right.** finetuning.uk deferred — no point validating the pipeline against a second site if the output on the first is unshippable.
2. **Item 25 first** — planner decomposition. Without one-output-per-work-item, prompt rewrites can't land cleanly. Fix the structure before the content.
3. **Item 23 fix A** — prompt rewrite for icons. Cheap experiment; if it produces shippable flat-vector icons via SDXL after item 25 is in place, item 24 might not be urgent.
4. **Item 24 if needed** — reference-image plumbing. The fallback if prompt-rewriting alone can't get clean icons out of SDXL. Also opens up consistent multi-site visual identity.
5. **Item 22 in parallel with the above** — small Go fix for purpose field. Doesn't block icon work.
6. **Item 21 last** — services hero coverage. Decide whether services warrants imagery; cheap to address once the icon question is settled.
7. Once robot-hands.com renders correctly end-to-end with appropriate imagery, THEN trigger finetuning.uk for parallel verification.
8. After both sites verified, switch focus to architectural items (6 — watchdog for stuck claims, 12 — asset→page rerender coupling).

---

## Cross-references

- `FOCUS_dispatch_diagnostic.md` — dispatch architecture deep-dive, Q1-Q4
- `016_debugging_guide.md` — operational reference for state-machine debugging
- `PLAN_imagery_loop_closure.md` — broader plan; this session closes Phases 2G+2H
- `icon_dimensions_patch.go` — patch for kindDefaults["icon"] (status: deployed per user, awaiting verification)
- `trigger_rerender_robot_hands.sh` — next immediate action
- `trigger_audit_finetuning.sh` — second immediate action

---

## Future workstream: data graphs / charts (added 2026-05-20)

Separate pipeline, NOT an extension of image generation. Image models cannot
plot real data — they fabricate values. Graphs need: real data fetch (EIA/FRED/
CSV) → code-rendered chart (go-echarts or Plotly) → optional LLM annotation
layer (Claude picks the story, code owns the numbers). Needs a new `chart`/
`graph` kind distinct from `infographic` (which stays Banana/decorative), a new
`data-chart-generator` agent, and a planner-prompt + CHECK-constraint change.

Full scoping in `FUTURE_data_graph_pipeline.md`.

## Icon background resolution (added 2026-05-20)

Transparency abandoned as too fragile — image models paint a transparency
checkerboard into RGB rather than producing alpha (confirmed: icon_cycle_time
mode=RGB has_alpha=False). Decision: option 2, "embrace the box" — icons
generated on a flat selectable grey background, placed in a styled CSS chip.
Planner prompt edited (transparent/plain → flat #EEEEEE bg + #4A4A4A line,
no checkerboard) via `planner_icon_background_fix.sql`. CSS side: icon
containers styled as chips (border-radius + padding + matching/contrasting grey).

## Icon format: jpg vs png (added 2026-05-20)

`ImagePurposes["icon"]` in `url_helpers.go` is currently `{240, 240, 85, "jpg"}`
— icons deploy as 240×240 jpg. With transparency abandoned (flat grey chip),
jpg is acceptable: no alpha to preserve, smaller files.

HOWEVER: if you want png icons — e.g. for crisper line art at small sizes, jpg
artifacts around thin grey lines can look muddy — that's a one-word change in
`ImagePurposes["icon"]` from `"jpg"` to `"png"`. Worth considering for line-art
icons specifically, since jpg compression is unkind to thin lines on flat
backgrounds. But not required.

# Imagery work — STATUS as of 2026-05-12

Successor to `STATUS_imagery_2026-05-08.md`. Records the completion of Phase
2F (variant deploy pipeline) plus the loader-snapshot defect that surfaced
during verification, the adapter timeout/logging fix, and the product-rendering
finding from the component audit.

A sibling doc `STATUS_affiliate_sites_2026-05-12.md` covers affiliate-site
work separately so this stays focused on imagery.

---

## At-a-glance

| Phase | Title | Code | Schema | DB-state | E2E verified |
|---|---|:-:|:-:|:-:|:-:|
| 0.1 | Read `imagery_direction` into image prompt | ✅ | n/a | ✅ | ✅ |
| 0.2 | Populate `origin_prompt` and `origin_model` | ✅ | n/a | ✅ | ✅ |
| 1.1 | Discovery check `unfulfilled_image_prompt` | ✅ | n/a | ✅ | ✅ |
| 1.2 | Discovery check `placeholder_image_in_use` | ✅ | n/a | ✅ | partial |
| 1.3 | Discovery check `image_url_404` | ✅ | n/a | ✅ | partial |
| 1.4 | Register checks in `design-discovery-agent` | n/a | ✅ | ✅ | ✅ |
| 1.5 | Smoke test handler dispatch | ✅ | ✅ (hotfix) | ✅ | ✅ |
| 2A | Asset locking columns | n/a | ✅ applied | n/a | n/a |
| 2B | Add `asset_key` column + new unique index | n/a | ✅ applied | n/a | n/a |
| 2C | `store_asset` writes `asset_key`; switch ON CONFLICT | ✅ deployed | n/a | n/a | n/a |
| 2D | Drop old `(site_id, purpose)` unique constraint | n/a | ✅ applied | n/a | partial |
| 2E | Variant path through image-build-handler | ✅ deployed | ✅ applied | ✅ | ✅ |
| 2F | Refactor `image-build-handler` to spawn `asset-deployer` for deploy | **✅ deployed** | **✅ applied** | **✅** | **✅ 2026-05-12** |
| 2G | (proposed) `imagery_plan` extension to planner | — | — | — | — |
| 2H | (proposed) Image generator request shape | — | — | — | — |
| 3 | Adoption image mirror | — | — | — | — |
| 4 | Visual auditor — text-only imagery awareness | — | — | — | — |
| 5 | Vision-capable LLM path | — | — | — | — |
| 6 | `imagery-quality-auditor` agent | — | — | — | — |

Legend: ✅ done · ⏳ pending · — not started · "partial" = check fires
correctly but no candidate site has exercised the symptom path yet.

---

## End-to-end verification — 2026-05-12

Direct invocation of `image-build-handler` for `hero_about` on
robot-hands.com (`00ff3af5-dad8-4770-9f70-3edc267a3c92`) produced the full
chain working:

1. image-build-handler took the variant branch (Phase 2E)
2. Spawned image-generator
3. Image-generator called Stability SDXL — image matched the prompt
   (industrial robotic gripper, blue-accented lighting, technical scene)
4. `store_variant_asset` wrote the asset row with `purpose=hero,
   asset_key=hero_about`
5. Spawned asset-deployer (Phase 2F)
6. asset-deployer pulled from S3, deployed to git as
   `assets/images/hero-about.jpg`
7. All three orchestrations COMPLETED cleanly

Total time end-to-end: about 90 seconds. The variant pipeline structurally
works end-to-end.

---

## What changed in this round

### Phase 2F — `image-build-handler` spawns `asset-deployer` (Option A delivered)

Resolves the storage architecture mismatch documented in the previous status
doc (chassis pod had no S3 credentials; `deploy_image_asset` ran in-process
and failed with "storage client not available").

The migration replaces three inline `deploy_image_asset` step pairs
(`deploy_logo`, `deploy_hero`, `deploy_variant`) in the `image-build-handler`
workflow with spawn+call patterns targeting `asset-deployer`. Storage env
flows naturally because `asset-deployer` is in the `isStorageEnabledAgent`
list (`spawn_actions.go` line 2972), so the spawn action injects
S3_ENDPOINT, IMAGE_BUCKET, etc. into the child pod's environment.

Files:
- `phase_2f_image_build_handler_spawn_asset_deployer.sql` (migration)
- `phase_2f_pre_migration_backup.sql`

The migration also adds `asset_key` to asset-deployer's
`input_schema.optional` and to its `deploy_asset` step's
`config.input_fields`, so per-variant deploy paths flow through.

### Two patches landed alongside Phase 2F

**Loader-snapshot defect (root cause of "migration's in DB but chassis uses
old workflow")**

Initial wrapper-test trigger of asset-deployer post-Phase-2F succeeded
end-to-end (file landed in robot-hands.com git) but deployed as `hero.jpg`
not `hero-about.jpg`. Master extractor logs showed `requested_fields:
["s3_uri","deploy_path","purpose","domain"]` — no `asset_key` — despite
the migration having added it to the active row's
`default_config.workflow.steps.deploy_asset.config.input_fields`.

Cache hypothesis was wrong (chassis restart didn't help). Snapshot-shadowing
hypothesis confirmed via SQL:

```
asset-deployer: active row version=1, snapshot version=1001
(via the version+1000 offset in 021_model_swap_and_rollback.sql)
```

Loader queries with `ORDER BY version DESC LIMIT 1` and no `is_active=true`
filter were picking the snapshot. Two loaders patched:

- `processor.go::loadAgentDefinition` (line 345) — added
  `is_active = true AND deleted_at IS NULL AND (is_snapshot IS NULL OR
  is_snapshot = false) ORDER BY version DESC LIMIT 1`
- `spawn_actions.go::getAgentDefinition` (line 2120) — added
  `is_active = true AND (is_snapshot IS NULL OR is_snapshot = false)` to
  existing filter

Gold-standard pattern came from `extractAIEndpointFromHandler` (production
context line 16078). Other loaders already correct:
`loadAgentDefinitionForAction`, `loadAgentDefinitionForImageAction`,
`AgentDefinitionDiscovery.FindByType`, `isKnownAgentType`. The two patched
loaders were the only ones using the unguarded `ORDER BY version DESC`
shortcut.

Builds deployed as `docker.io/aqls/agent-chassis:v1.0.1006` (chassis with
loader fix), then `v1.0.1008` (with image-build-handler integration tested).

**Snapshot audit closed clean.** Only `asset-deployer` and
`image-build-handler` had snapshots with material differences from their
active rows. Affected window 2026-05-10 16:04 → 2026-05-12 ~16:00.
Production blast-radius query returned zero rows outside the test site.

Files:
- `processor.go.patch`, `processor.go`
- `spawn_actions.go.patch`, `spawn_actions.go`

**image-generator-adapter HTTP timeout + zap marshal error**

Two defects in `imagegenerator/dynamic_adapter.go`:

1. Line 140: `Timeout: 30 * time.Second` → `Timeout: 120 * time.Second`.
   SDXL response body reads were exceeding 30s, causing periodic
   tear-downs mid-response.
2. Line 329: `zap.Any("...request...", req)` → three safe fields
   (method, url, body_length_bytes). The `*http.Request.GetBody` field
   has type `func() (io.ReadCloser, error)` which zap couldn't marshal,
   producing `json: unsupported type` errors at line 328.

Files:
- `dynamic_adapter.go.patch`, `dynamic_adapter.go`

---

## Component audit finding (2026-05-12) — products are designed to query affiliate_products

The audit on 2026-05-12 surfaced something significant about the rendering
layer that changes how to think about product imagery.

Product-rendering components already exist and are non-trivial:

| Component | Template chars | Schema chars | Quality |
|---|---|---|---|
| `product-card-with-cta` | 4438 | 633 | — |
| `product-grid` | 3888 | 624 | — |
| `product-hero_pre_037` | 7402 | 3470 | 100 |
| `product-details_pre_037` | 9250 | 5991 | 100 |
| `product-specs` | 10862 | 2 (effectively empty schema) | — |

Real templates, real schemas (except `product-specs` whose schema is two
chars and likely needs filling in). And `product-card-with-cta`'s
input_schema declares:

```json
"products": {
  "type": "array",
  "source": "query.affiliate_products",
  "items": {
    "image_url": {"type": "image"},
    "name": {...}, "price": {...}, "rating": {...},
    "tagline": {...}, "cta_text": {...},
    "review_url": {...}, "rating_count": {...},
    "rating_stars": {...}, "affiliate_url": {...}
  }
}
```

So the component is structurally designed to:

1. Query `affiliate_products` directly via a resolver
2. Receive an `image_url` per product (typed `image`)

Today no resolver populates this array — it's a wired socket with no plug.
When a resolver does exist (sibling doc tracks this), the illustration plan
becomes a specific precedence rule inside it:

> if `product.custom_image_id` is set, resolve via that; else use
> `cached_image_url`

This means the **product illustration work isn't blocked by rendering**.
Once the resolver lands, illustrations slot into a precise hook. The
illustration plan drafted earlier
(`/mnt/user-data/outputs/PLAN_product_illustration.md`) plugs in as a
precedence rule, not a separate rendering layer.

The takeaway for the imagery roadmap: **product illustration is one
specific case of "varied imagery for a component", not a special pipeline.**
What matters for general imagery is making the planner emit more than
hero+logo prompts.

---

## What's actually in the imagery pipeline today

```
domain → planner → image_prompts (per variant) → site_specs
                          │
                          ▼
              discovery checks (Phase 1) detect missing
                          │
                          ▼
              site_work_items (unfulfilled_*)
                          │
                          ▼
              dispatch loop claims (currently off)
                          │
                          ▼
              image-build-handler
                ├── conditional: logo / hero / variant
                ├── spawn image-generator
                ├── call image-generator → Stability SDXL
                ├── store_asset (assets table, asset_key keyed)
                ├── spawn asset-deployer (Phase 2F)
                └── call asset-deployer → git commit
                          │
                          ▼
              File in repo at assets/images/{purpose|asset_key}.{ext}
```

What works:

- Multiple asset_keys per site (the unique constraint on (site_id, purpose)
  was dropped in Phase 2D)
- The planner produces rich `image_prompts` with site-specific variant keys
  — robot-hands.com has seven: `logo`, `hero_home`, `hero_about`,
  `hero_tools`, `hero_matchmatrix`, `hero_how_it_works`,
  `hero_selection_guide`
- The variant path through image-build-handler routes correctly post-Phase-2E
- Per-variant deploy filenames work (`hero-about.jpg`, not just `hero.jpg`)
- Storage env is correctly scoped via the spawn pattern (Phase 2F)
- Agent definition loaders correctly bypass snapshots (patched 2026-05-12)

What doesn't:

- The improvement loop / dispatch loop is currently OFF (user request,
  noise reduction). Variant items sit `triaged` waiting. Manual triggering
  works.
- No retry semantics for failed items — they sit `failed` until manually
  cleaned up. The dedup index allows duplicates between `complete` rows
  but not between `triaged`+`failed`, which made the 2026-05-12 reset
  awkward (had to delete duplicates before resetting). Worth understanding
  the index definition exactly before we lean on this much.

---

## Historical phase detail (2A–2E) — preserved from prior status

Detail captured during the original delivery, kept here for context. Skip
if you only need the current state.

### Phase 2E — variant path through image-build-handler (delivered, then verified end-to-end on 2026-05-12)

Makes the 5 hero variant work items routable end-to-end. Four coordinated
changes:

**1. `check_unfulfilled_image_prompt.go` (Go)** — variants now route to
`image-build-handler` (previously flag-only with empty `handler_agent`).
Spec format updated to include `asset_key` and `prompt` directly (avoiding
workflow-engine dynamic key lookup). The `(purpose, asset_key)` split is:
- canonical logo → `purpose=logo, asset_key=logo`
- canonical hero → `purpose=hero, asset_key=hero`
- variant `hero_about` → `purpose=hero, asset_key=hero_about`

`item_key` keyed on `asset_key` so variants get distinct dedup keys.

**2. `imagery_helpers.go` (Go)** — added `hasActiveAssetForAssetKey`. The
classifier switched to this helper for asset-existence checks since
`hasActiveAssetForPurpose` would now report "hero exists" when only
`asset_key=hero` exists — wrong for `asset_key=hero_about` lookup.

**3. `deploy_image_asset_action.go` (Go)** — added `asset_key` and
`asset_key_field` config support. When `asset_key` is provided AND differs
from `purpose`, derives a per-variant deploy path:
`assets/images/<asset_key with _ → ->.{ext}`. Existing logo/hero generations
don't pass asset_key, so the new branch is skipped and their default
purpose-derived paths remain unchanged.

**4. `v3_site_actions.go` `StoreAssetAction` (Go)** — added
`asset_key_field` config support (JSONPath lookup), needed for the variant
workflow which passes asset_key through `input_data.spec.asset_key` rather
than as a literal. Existing literal `asset_key` config still works first.

**5. `image-build-handler` workflow migration (SQL)** — added a third
branch for variant items. The existing logo and hero paths are NOT
modified. New steps:
- `check_item_type` updated to route variants first.
- `check_logo_or_hero` (new) preserves the existing logo/hero split.
- `spawn_image_gen_variant`, `call_variant_gen`, `store_variant_asset`,
  `deploy_variant` (new) — variant generation chain.

### Phase 2D — drop old `(site_id, purpose)` unique constraint (delivered + applied)

Removes `idx_assets_site_purpose_unique`. After this, the only uniqueness
protection on `assets` is `idx_assets_site_asset_key_unique` from Phase 2B.
Multi-image case (multiple rows with same `purpose` but distinct `asset_key`)
becomes possible at the schema level.

Sanity checks in the migration:
1. Asserts no active row has `purpose IS NOT NULL AND asset_key IS NULL`
2. Asserts `idx_assets_site_asset_key_unique` exists

Both checks abort the migration with informative errors. `\set ON_ERROR_STOP
on` ensures psql exits at the first error rather than continuing.

### Phase 2C — `store_asset` writes `asset_key` and switches ON CONFLICT (deployed 2026-05-09)

Go-only change to `StoreAssetAction`:

1. Extract `asset_key` from config, defaulting to `purpose`.
2. Updated INSERT to include the new column.
3. Switched ON CONFLICT target from `(site_id, purpose) WHERE purpose IS
   NOT NULL` to `(site_id, asset_key) WHERE asset_key IS NOT NULL AND
   status = 'active'`.
4. Added `purpose = EXCLUDED.purpose` to SET clause.

No behaviour change for production at deploy time. Every existing caller
passed `purpose` only; default-to-purpose populated `asset_key=purpose`;
the new ON CONFLICT target caught the same conflicts the old one did.
Phase 2C was forward-compatible scaffolding for Phase 2D.

### Phase 2B — `asset_key` column for multi-image readiness (delivered + applied)

Added `asset_key text` (nullable) to `assets`, backfilled from `purpose`
for existing rows, and added new unique index
`idx_assets_site_asset_key_unique (site_id, asset_key) WHERE asset_key IS
NOT NULL AND status = 'active'`. The old constraint stayed in place during
this phase. Both indexes coexisted harmlessly while `asset_key = purpose`
for everything.

### Phase 2A — asset locking columns (delivered + applied)

Added `locked_at timestamp with time zone` and `locked_by text` to `assets`,
plus the partial index `idx_assets_locked btree (locked_at) WHERE locked_at
IS NOT NULL`. Mirrors `page_components` and `site_components` exactly. Pure
DDL — no Go change.

#### Lock semantics — third correction

The docstring went through three iterations as the actual lock model
became clear (v1 wrong on time-based expiry; v2 incomplete on hard-vs-soft;
v3 final, reflecting `check_component_lock.go`):
- Detection: `locked_at IS NULL` vs `IS NOT NULL`.
- Classification: hard (admin / admin-removed / checkpoint) vs soft
  (deploy / manual / auditor names / others) via `locked_by`.
- No time-based expiry today. Timed expiry is documented design intent
  in 004 v4 / 007 v4, paired with the audit-pass-counter reset that IS
  implemented. The lock-expiry half didn't ship.

Approved policy (2026-05-08): implement timed expiry as a focused future
project. Human-set locks stay permanent; only `'deploy'` and auditor
approvals get timed expiry. Reference: `LOCKS_should_locks_expire.md`.

### Phase 1.5 hotfix — `output_mapping` added to image-build-handler

A pre-existing bug surfaced during verification: `image-build-handler`'s
image-generator calls were missing the `output_mapping` block that
`pageflow-builder` and `site-work-orchestrator` had. Without it, the
generator's response dropped into `image_result` wholesale, so the URL
lived at `image_result.response.generate.response.image_url` — three
levels deeper than `store_asset` looks. Result: silent `{stored: false,
reason: "no asset URL found"}`.

This bug pre-dated Phase 0/1 work — silently dropping every image asset
that flowed through `image-build-handler` (the dispatch path). Sites
that built via `pageflow-builder` directly weren't affected.

---

## Findings worth recording (not blockers)

### Build-dispatch-loop is not claiming triaged image items (when loop is on)

Imagery work items emitted at 10:38 sat in `triaged` status for >4 hours
while multiple `build-dispatch-loop` orchestrations ran (14:13, 14:19,
14:40, 14:43, 14:52) and processed page items via `page-build-handler`.
None claimed the imagery items despite their
`handler_agent='image-build-handler'` and `pipeline='build'` matching what
the loop's `load_work_items` should filter on.

Possible causes (not investigated):
- Priority ordering: page items may have higher priority (lower number)
  than the 65-70 we set on imagery items.
- A filter on `LoadWorkItemsAction` we haven't traced.
- Fuel cap exhausted before reaching imagery items.

This didn't block verification — direct invocation of `image-build-handler`
worked fine. But it means imagery items don't self-process through the
normal dispatch path on this site right now. Worth investigating
separately as it's blocking organic image generation. Currently moot
because the loop is off.

### Hero variant items keep getting claimed despite empty handler_agent (historical)

Resolved by Phase 2E (variants now have a real `handler_agent` value, no
longer flag-only). The 10 `failed` items from the pre-Phase-2E period
were cleaned up on 2026-05-12 (5 deleted as duplicates, 5 reset to
`triaged`).

### Dedup index allows duplicate `complete` rows but blocks duplicate `triaged`/`failed`

Surfaced during the 2026-05-12 reset of the failed variant items. The
`idx_swi_dedup` index appears to be partial (likely excludes certain
statuses), allowing two `complete` rows for the same (site_id, item_key)
to coexist while a single `triaged` row blocks a duplicate. This affects
retry semantics when discovery re-runs. Worth understanding the actual
index definition before designing retry logic.

### Image generation produces `demo_client`-pathed S3 URLs

URLs follow `s3://.../images/demo_client/<date>/<uuid>.png`. The path uses
the client_id from the orchestration request, not the site_id. Confirmed
correct — `demo_client` is the actual client. Not a finding.

### `origin_prompt` captures composed prompt only

Today's read in `getImageryDirectionForSite` pulls only the strategic
`site_specs.design_intent.imagery_direction`. Once Phase 1 of doc 030
ships (brief renderer + `site_plan_directives` table), this should be
extended to also pull plan-time directives where `category=design AND
subject LIKE 'imagery%'`. Tracked as a successor.

### Cosmetic — "Claim timed out — handler pod likely died"

Earlier `complete` work items have this stamped in their `error` column
even though their orchestration_states show `COMPLETED`. The reaper writes
the message before the actual completion lands. Cosmetic only.

### `updated_at` on `agent_definitions` rows identical to microsecond after rollouts

Either a bulk touch happens at deploy, or a single migration ran twice.
Not urgent but worth understanding.

---

## Scope: imagery for the broader site (not just heroes/products)

`FOCUS_imagery_assessment.md` section 9 lists structural gaps for "all
kinds of imagery, all kinds of sites":

| Gap | Status |
|---|---|
| Multiple-purpose-per-site model | DONE (Phase 2A–D) |
| Adapter supports model/provider selection | Not started — Stability SDXL only |
| Richer image request shape (model, negative_prompt, style_preset, reference_image_uri, seed, lora) | Not started |
| Planner requesting varied imagery (illustrations, icons, products, infographics) | Not started — planner emits `image_prompts.{logo, hero_*}` only |
| Components declaring diverse image needs in `input_schema` | Partial — product components declare it; most do not |
| Icon/SVG rendering | Not started |

The first item is done as a side effect of variant work. The rest are
proper next phases.

---

## Three proposed phases for imagery, in priority order

Imagery-specific. The product/affiliate work is in the sibling doc.

### Phase 2G (proposed) — `imagery_plan` extension to the planner

**Goal:** Move from "two-or-three hero prompts" to "a list of imagery
requirements" the planner emits per site.

Today the planner emits `image_prompts: {logo, hero_home, hero_about, ...}`.
That's hero-shaped only. To extend to illustrations / decorative graphics /
infographics, the shape needs to broaden:

```json
"imagery_plan": [
  {"key": "hero_home", "kind": "hero", "prompt": "..."},
  {"key": "section_about_illustration", "kind": "illustration",
    "prompt": "...", "style": "..."},
  {"key": "service_icon_consulting", "kind": "icon",
    "prompt": "...", "constraints": {...}},
  ...
]
```

`kind` lets downstream handlers choose generation parameters appropriate
to the kind (illustration → stylised, lower cfg_scale; icon → SVG path;
infographic → different generator entirely, later).

The existing `image_prompts.{purpose}` shape is preserved (components
reading it don't change), but the planner additionally emits
`imagery_plan` for the things components don't ask for by name.
Discovery checks then look for `imagery_plan` entries with no matching
asset — same pattern as today.

This is the "richer planner output" item from doc 030 / FOCUS 9.4.

### Phase 2H (proposed) — Image generator request shape

**Goal:** Pass more than `prompt + width + height` to Stability.

Today's adapter sends a fixed shape. To do illustration well, we need
`style_preset`, `negative_prompt`, `cfg_scale`, `steps` to be passed
through. Some of these are already in the adapter's code as hardcoded
values; they need to be promoted to call-site config.

This unlocks: lower cfg_scale for illustration kind (more creative),
higher for product-shot kind (more literal); negative prompts to exclude
brand markings for product illustration; reproducible seeds when we want
to regenerate variants.

Stays Stability-only. Multi-provider routing (FLUX, Midjourney, Banana
Pro Gemini) is a later phase — separate plan when we decide we want it.

### Phase 3 (from existing PLAN) — Adoption image mirror

**Goal:** When we adopt a site, persist its crawled images to our S3
rather than discarding them.

Already specced in `PLAN_imagery_loop_closure.md` Phase 3. Independent
of 2G and 2H. Knock-on value: mirrored adopted images become historical
record, reference inputs for img2img (future), and a signal for the
auditor on whether our regenerated imagery is on-brief.

User has indicated this is lower urgency. Documented here so we don't
lose track; not on the active path.

### Phases 4–6 (existing PLAN) remain on the back burner

Visual auditor extension (text-only), vision-capable LLM path,
imagery-quality-auditor. The back half of `PLAN_imagery_loop_closure.md`.
Not changed in priority. They're audit infrastructure that matters once
we have meaningfully more imagery running through the system.

---

## What's parked for the deep discussion

Three items that aren't blocking imagery but matter structurally:

1. **Consumer-group race** on chassis pods — chassis runs at replicas=1
   to avoid this. Documented in
   `ANALYSIS_chassis_response_consumer_group_race.md`. Real architectural
   defect. Required before scaling chassis back up.

2. **Snapshot retention policy** — the snapshot mechanism in
   `021_model_swap_and_rollback.sql` accumulates snapshots forever. The
   loader fix prevents shadowing, but the structural risk remains. Worth
   a TTL or "keep last N" policy.

3. **Multiple agent-definition loaders** — three different queries with
   three different filter strictness levels across the chassis Go code.
   A single `AgentDefinitionRepository` would prevent drift but is a
   larger refactor.

None block the imagery work. Worth a sit-down session at some point.

---

## Live data state (robot-hands.com, 2026-05-12)

### `assets` for the site

| asset_key | purpose | origin_model | status | source |
|---|---|---|---|---|
| logo | logo | sdxl | active | 2026-05-08 |
| hero | hero | sdxl | active | from earlier work |
| hero_about | hero | sdxl | active | 2026-05-12 verification |

`image_prompts` in site_specs.site_plan contains: `logo`, `hero_home`,
`hero_about`, `hero_tools`, `hero_matchmatrix`, `hero_how_it_works`,
`hero_selection_guide` (7 variants).

### `site_work_items` for the site (imagery-related)

After the 2026-05-12 reset:

- 5 `unfulfilled_hero_variant` items in `triaged` status (`hero_about`,
  `hero_how_it_works`, `hero_matchmatrix`, `hero_selection_guide`,
  `hero_tools`). Note `hero_about` is now stale (asset exists in git) —
  discovery would mark it `complete` if it ran fresh.
- 2 `complete` items for `unfulfilled_image_prompt:hero` and
  `unfulfilled_image_prompt:logo` (one each from 2026-05-06 and
  2026-05-08, both stamped with reaper or AI-unavailable errors in the
  error column despite COMPLETED orchestrations).
- `affiliate_products` is empty (no live affiliate feeds yet).

The improvement loop is OFF (user request).

---

## Open issues / next steps

In rough priority order:

1. **Decide on Phase 2G shape.** The `imagery_plan` extension is the
   structural change that unblocks anything beyond hero+logo. Worth a
   short discussion before writing code.

2. **Investigate the dedup index** on `site_work_items`. One query to
   see the partial index condition. Affects retry semantics when
   discovery re-runs.

3. **Build-dispatch-loop claim ordering** — when the loop turns back on,
   does it claim imagery items? Pre-2F evidence suggests not. May need
   investigation. Currently moot because the loop is off.

4. **Lock-expiry project (deferred).** Adds `lock_type` +
   `lock_expires_at` columns uniformly across the four Pattern A tables
   (page_components, site_components, site_plan_directives, assets).
   Restores the rhythm doc 004 v4 envisaged. Approved policy:
   human-set locks stay permanent; only `'deploy'` and auditor approvals
   get timed expiry. Sequenced after the imagery loop work completes.
   Reference: `LOCKS_should_locks_expire.md`.

---

## Cumulative file inventory

Phase 0 deliverables (deployed):
- `phase_0_revised_generate_image_actions.diff`
- `phase_0_2_store_asset_action.diff`
- `phase_0_combined_migration.sql`
- `phase_0_pre_migration_backup.sql`
- `generate_image_actions.go` (patched file)
- `v3_site_actions.go` (patched file)

Phase 1 deliverables (deployed):
- `phase_1/check_unfulfilled_image_prompt.go` (with hero variant extension)
- `phase_1/check_placeholder_image_in_use.go`
- `phase_1/check_image_url_404.go`
- `phase_1/imagery_helpers.go`
- `phase_1/phase_1_pre_migration_backup.sql`
- `phase_1/phase_1_register_imagery_checks.sql`

Phase 1.5 hotfix (deployed):
- `phase_1_5_pre_migration_backup.sql`
- `phase_1_5_image_build_handler_output_mapping.sql`

Phase 2A deliverables (applied):
- `phase_2a_pre_migration_backup.sql`
- `phase_2a_assets_locking.sql` (final v3 docstring)

Phase 2B deliverables (applied):
- `phase_2b_pre_migration_backup.sql`
- `phase_2b_add_asset_key.sql`

Phase 2C deliverables (deployed 2026-05-09):
- `phase_2c_store_asset_action.diff`
- `v3_site_actions.go` (full patched file)

Phase 2D deliverables (applied):
- `phase_2d_pre_migration_backup.sql`
- `phase_2d_drop_purpose_constraint.sql`

Phase 2E deliverables (deployed, end-to-end verified 2026-05-12):
- `phase_2e/check_unfulfilled_image_prompt.go`
- `phase_2e/imagery_helpers.go`
- `phase_2e/phase_2e_deploy_image_asset_action.diff`
- `phase_2e/phase_2e_store_asset_action.diff`
- `phase_2e/phase_2e_pre_migration_backup.sql`
- `phase_2e/phase_2e_image_build_handler_variant_path.sql`

Phase 2F deliverables (deployed + verified 2026-05-12):
- `phase_2f_image_build_handler_spawn_asset_deployer.sql`
- `phase_2f_pre_migration_backup.sql`
- `processor.go.patch`, `processor.go` (loader snapshot fix)
- `spawn_actions.go.patch`, `spawn_actions.go` (loader snapshot fix)
- `dynamic_adapter.go.patch`, `dynamic_adapter.go` (timeout + zap fix)
- `ANALYSIS_phase_2f_two_defects.md` (root cause investigation,
  revised 2026-05-11)

Lock investigation:
- `LOCKS_findings_and_proposed_corrections.md`
- `LOCKS_should_locks_expire.md`
- `031_locks_proposed_update.md`
- `PROPOSED_031_locks_addendum.md`

Planning and assessment:
- `FOCUS_imagery_assessment.md`
- `PLAN_imagery_loop_closure.md`
- `PLAN_product_illustration.md` (parked behind resolver work in sibling doc)
- `ASSESSMENT_phase_0_1_vs_phase_1_architecture.md`
- `ADDENDUM_phase_0_1_verification.md`
- `STATUS_imagery_2026-05-06.md` (prior status)
- `STATUS_imagery_2026-05-08.md` (prior status)
- `STATUS_imagery_2026-05-12.md` (this file)
- `STATUS_affiliate_sites_2026-05-12.md` (companion)
- `ANALYSIS_chassis_response_consumer_group_race.md` (parked structural)

Diagnostics from the verification work (kept for future reference):
- `find_image_url_path.sql`
- `phase_1_5_*_diagnostics*.sql`
- `improvement_loop_status_check.sql`
- various others

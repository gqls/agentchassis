# Imagery work — STATUS as of 2026-05-08

Successor to `STATUS_imagery_2026-05-06.md`. Records end-to-end verification
of Phase 0.1, 0.2, and 1, plus the open issues uncovered along the way.

---

## At-a-glance

| Phase | Title | Code | Schema | DB-state | E2E verified |
|---|---|:-:|:-:|:-:|:-:|
| 0.1 | Read `imagery_direction` into image prompt | ✅ | n/a | ✅ | **✅ today** |
| 0.2 | Populate `origin_prompt` and `origin_model` | ✅ | n/a | ✅ | **✅ today** |
| 1.1 | Discovery check `unfulfilled_image_prompt` | ✅ | n/a | ✅ | ✅ |
| 1.2 | Discovery check `placeholder_image_in_use` | ✅ | n/a | ✅ | partial |
| 1.3 | Discovery check `image_url_404` | ✅ | n/a | ✅ | partial |
| 1.4 | Register checks in `design-discovery-agent` | n/a | ✅ | ✅ | ✅ |
| 1.5 | Smoke test handler dispatch | ✅ | ✅ (hotfix) | ✅ | **✅ today** |
| 2A | Asset locking columns | n/a | **✅ applied** | **✅ applied** | **n/a (scaffolding)** |
| 2B | Add `asset_key` column + new unique index | n/a | **✅ applied** | **n/a (forward-compat)** | **n/a (no behaviour change)** |
| 2C | `store_asset` writes `asset_key`; switch ON CONFLICT | **✅ deployed** | n/a | **n/a (forward-compat)** | **n/a (no behaviour change)** |
| 2D | Drop old `(site_id, purpose)` unique constraint | n/a | **✅ applied** | **n/a (constraint removal)** | partial (multi-image enabled at schema level) |
| 2E | Variant path through image-build-handler | **✅ deployed (workflow + action support)** | **✅ applied** | **partial — variant generation verified end-to-end through `store_variant_asset`; `deploy_variant` blocked by storage architecture mismatch (see below)** | partial |
| 2F | **NEW: Refactor `image-build-handler` to spawn `asset-deployer` for deploy steps (Option A)** | — | — | — | — |
| 3 | Adoption image mirror | — | — | — | — |
| 4 | Visual auditor — text-only imagery awareness | — | — | — | — |
| 5 | Vision-capable LLM path | — | — | — | — |
| 6 | `imagery-quality-auditor` agent | — | — | — | — |

Legend: ✅ done · ⏳ pending · — not started · "partial" = check fires
correctly but no candidate site has exercised the symptom path yet.

---

## Today's verification — single asset row tells the whole story

Direct invocation of `image-build-handler` against `00ff3af5-...` produced
this row in `assets`:

```
purpose      : logo
origin_type  : generated
origin_model : sdxl                          ← Phase 0.2 verified
status       : active
origin_prompt: Industrial photography of robotic grippers and end-effectors
               in real manufacturing environments — close-up detail shots
               showing machined surfaces, actuator mechanisms.            ← Phase 0.1
               A precise, technical logomark for Robot-Hands.com — a
               stylised robotic gripper or end-effector silhouette ...    ← Subject
url          : s3://...backblazeb2.com/.../demo_client/20260508/...png
created_at   : 2026-05-08 15:12:17
```

Every previously-unverified piece visible in one row:

- `origin_model='sdxl'` — Phase 0.2 column write happened.
- `origin_prompt` begins with the imagery_direction text from
  `site_specs.design_intent.imagery_direction`, truncated at the
  sentence boundary per the helper's logic — Phase 0.1 composition
  fired and was recorded.
- The composed prompt is followed by the planner's logo prompt with the
  helper's period-and-space separator — format is exactly as designed.
- `url` is a real, fresh, presigned B2 URL.

This is the row that confirms three weeks of plumbing fires correctly.

---

## What changed in this round

### [BLOCKER] Storage architecture mismatch — `deploy_image_asset` runs in chassis pod, can't reach storage

After the Phase 2 series fully deployed, the verification trigger generated `hero_about` for robot-hands.com (asset id `ca7cac72-...`). The variant workflow ran end-to-end through Phase 2E correctly: variant chain spawned, image-generator generated SDXL image (orchestration `fd819e3f-...` COMPLETED at 11:08:30), `store_variant_asset` wrote asset row with `purpose=hero, asset_key=hero_about`. **Phase 2E is verified at every layer up to and including `store_variant_asset`.**

The `deploy_variant` step then failed with `storage client not available` (orchestration `a61b21bc-...` at 11:08:30). This is NOT a Phase 2E bug — `deploy_logo` and `deploy_hero` failures with the same error go back to **2026-05-08 15:11:41**, the day Phase 2C deployed. So the regression onset is May 8 15:11:41 ± chassis pod restart time.

**Root cause** (confirmed via kubectl exec on chassis pod):
- `S3_ENDPOINT=` empty, `IMAGE_BUCKET=` empty in chassis pod env.
- B2_APPLICATION_KEY_ID and B2_APPLICATION_KEY are set, but the chassis init at `cmd/chassis-startup` reads `os.Getenv("IMAGE_BUCKET")` and silently skips storage client init when empty (`a.storageClient = nil`).
- `deploy_image_asset_action.go` reads `params.StorageClient` and errors when nil.
- User design intent (clarified): chassis should NOT carry storage env vars because different operations write to different buckets and may run at different times. Storage belongs in spawned child agents which receive bucket-specific env vars from the `storage-config` ConfigMap via `spawn_actions.go`.

**Architecture per `spawn_actions.go`** (`/mnt/user-data/uploads/spawn_actions.go`):
- When parent spawns child, if child agent_type is in `isStorageEnabledAgent` list (line 2972, includes `asset-deployer`) OR child's category is `orchestrator` / `code-driven`, spawn injects: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, B2_APPLICATION_KEY_ID, B2_APPLICATION_KEY, S3_ENDPOINT, S3_REGION, IMAGE_BUCKET, ASSETS_BUCKET, S3_USE_PATH_STYLE.
- Spawned children get full storage config; chassis pod doesn't.

**Fix path: Option A — spawn `asset-deployer` for deploys**.

Critical discovery: **`asset-deployer` agent definition already exists** in production agent_definitions table:
- id: `e9a9bac9-dfe4-4aca-8f32-19738ac265c6`
- type: `asset-deployer`
- description: "Deploys a single image asset: downloads from S3, optimizes by purpose, commits to git. Reusable for any image deploy task."
- workflow: `deploy_asset → complete`, single step running `deploy_image_asset` with `input_fields: [s3_uri, deploy_path, purpose, domain]`
- input_schema: `required: [domain, s3_uri], optional: [deploy_path, purpose]`
- status: experimental
- already in `isStorageEnabledAgent` list

Comment at production code line 84513 explicitly says: *"New callers (asset-deployer) use input_fields: [s3_uri, deploy_path, purpose, domain]"* — confirms this was the intended canonical caller.

**Phase 2F scope** (next session):
1. Update asset-deployer's `input_schema.optional` to include `asset_key`.
2. Update asset-deployer's `deploy_asset` step's `config.input_fields` to include `asset_key`.
3. Update `image-build-handler` workflow: replace inline `deploy_image_asset` calls in `deploy_logo`, `deploy_hero`, `deploy_variant` steps with spawn_agent + call_agent pattern targeting asset-deployer (similar pattern to existing `spawn_image_gen` + `call_image_gen`).
4. After deploy, trigger variant generation again — variant should land in robot-hands.com git as `assets/images/hero-about.jpg`.

**Open questions for Phase 2F**:
- Pageflow-builder also calls `deploy_image_asset` directly via `deploy_logo_image` and `deploy_hero_image`. Same architecture issue — would also need refactoring. Not blocking variant work but worth aligning.
- Comment at line 84701 says: "This makes the asset-deployer self-contained: callers pass asset_id, ..." — there may be deeper design intent encoded that's worth reading before writing the migrations.
- How did logo.png and hero.jpg actually land in robot-hands.com git on/before May 8? No completed deploy orchestrations exist before that date. Possibilities: pageflow-builder during initial site build (different code path), or an earlier chassis pod that DID have IMAGE_BUCKET set. Worth confirming in Phase 2F before assuming the new design covers all paths.
- User noted: "agent chassis deployment hasn't recently changed but that doesn't mean code hasn't." Open question whether deploy_image_asset code itself recently changed in a way that broke previously-working in-chassis usage.

---

### Phase 2E — variant path through image-build-handler (delivered)

Makes the 5 hero variant work items routable end-to-end. Four
coordinated changes — all delivered, none yet deployed.

**1. `check_unfulfilled_image_prompt.go` (Go)** — variants now route to
`image-build-handler` (previously flag-only with empty `handler_agent`).
Spec format updated to include `asset_key` and `prompt` directly (avoiding
workflow-engine dynamic key lookup). The `(purpose, asset_key)` split is:
- canonical logo → `purpose=logo, asset_key=logo`
- canonical hero → `purpose=hero, asset_key=hero`
- variant `hero_about` → `purpose=hero, asset_key=hero_about`

`item_key` keyed on `asset_key` so variants get distinct dedup keys
(no longer collide as `unfulfilled_image_prompt:hero_about` overlaps with
sibling variants).

**2. `imagery_helpers.go` (Go)** — added `hasActiveAssetForAssetKey`. The
classifier switched to this helper for asset-existence checks since
`hasActiveAssetForPurpose` would now report "hero exists" when only
`asset_key=hero` exists — wrong for `asset_key=hero_about` lookup.
`hasActiveAssetForPurpose` retained for backward reference.

**3. `deploy_image_asset_action.go` (Go)** — added `asset_key` and
`asset_key_field` config support. When `asset_key` is provided AND
differs from `purpose`, derives a per-variant deploy path:
`assets/images/<asset_key with _ → ->.{ext}`. Existing logo/hero
generations don't pass asset_key, so the new branch is skipped and
their default purpose-derived paths remain unchanged. Files: see
`phase_2e_deploy_image_asset_action.diff`.

**4. `v3_site_actions.go` `StoreAssetAction` (Go, Phase 2E extension to
the Phase 2C patch)** — added `asset_key_field` config support
(JSONPath lookup), needed for the variant workflow which passes asset_key
through `input_data.spec.asset_key` rather than as a literal. Existing
literal `asset_key` config still works first. Files: see
`phase_2e_store_asset_action.diff`.

**5. `image-build-handler` workflow migration (SQL)** — adds a third
branch for variant items. The existing logo and hero paths are NOT
modified. New steps:
- `check_item_type` updated to route variants first.
- `check_logo_or_hero` (new) preserves the existing logo/hero split.
- `spawn_image_gen_variant`, `call_variant_gen`, `store_variant_asset`,
  `deploy_variant` (new) — variant generation chain.

The variant chain reads `input_data.spec.prompt` directly, passes
literal `purpose=hero` to store_asset and deploy_image_asset, and
passes `asset_key_field=input_data.spec.asset_key` to both for
per-variant routing. Idempotent — running the migration twice
overwrites with the same content. Files: `phase_2e_pre_migration_backup.sql`,
`phase_2e_image_build_handler_variant_path.sql`.

**Verified end-to-end** against a seeded postgres mirroring the
agent_definitions row's pre-Phase-2E shape:
- All 6 jsonb_set UPDATEs succeed.
- Final workflow has 17 steps (12 original + 4 variant + 1 new
  conditional).
- `check_item_type` routes variants first, falls through to
  `check_logo_or_hero` for canonical items.
- All existing logo/hero steps preserved untouched.
- Re-running migration is idempotent (no errors, no duplicate keys).

**No behaviour change for canonical logo/hero today** — they go through
the same steps with the same configs. Variants that were previously
flag-only now flow through the new chain.

### Phase 2D — drop old `(site_id, purpose)` unique constraint (delivered)

DROP INDEX migration with safety guards. Removes
`idx_assets_site_purpose_unique`. After this, the only uniqueness
protection on `assets` is `idx_assets_site_asset_key_unique` from
Phase 2B. Multi-image case (multiple rows with same `purpose` but
distinct `asset_key`) becomes possible at the schema level.

Files:
- `phase_2d_pre_migration_backup.sql`
- `phase_2d_drop_purpose_constraint.sql` (with `\set ON_ERROR_STOP on`
  and pre-drop sanity checks)

**Sanity checks built into the migration:**
1. Asserts no active row has `purpose IS NOT NULL AND asset_key IS NULL`
   — catches "straggler" rows that might exist if Phase 2C deploy
   happened later than Phase 2B migration. Provides the exact backfill
   SQL in the abort message if it fires.
2. Asserts `idx_assets_site_asset_key_unique` exists — catches the
   case where Phase 2B wasn't applied first.

Both checks abort the migration with informative errors. The
`\set ON_ERROR_STOP on` directive ensures psql exits at the first
error rather than continuing to the DROP INDEX statement (caught
during testing — the original migration would have dropped the index
even after the sanity check failed; fixed before delivery).

**Verified end-to-end** against three seeded postgres scenarios:
- **Happy path**: clean post-Phase-2C state, both indexes present, no
  stragglers. Migration succeeds, only new index remains. Multi-image
  inserts (3 hero variants for same site) succeed afterwards.
- **Sad path 1**: a straggler row with `purpose='logo' AND asset_key=NULL`.
  Migration aborts with the helpful backfill instruction; old index
  preserved.
- **Sad path 2**: new index dropped (simulating Phase 2B not applied).
  Migration aborts with prerequisite error; old index preserved.

**Behaviour change** — this is the first phase in 2A-D that genuinely
changes behaviour: multi-image purposes can now be inserted. The 5 hero
variants currently sitting in `failed` status (from earlier dispatch
attempts that hit the "agent_type is required" error because they have
empty `handler_agent`) become routable in principle. However, they
still won't be picked up by dispatch until Phase 2E parameterises the
hardcoded deploy paths (`assets/images/hero.jpg`, `assets/images/logo.png`)
in `image-build-handler` — currently every hero variant would overwrite
the same `hero.jpg` file even though they'd insert distinct rows.

### Phase 2C — `store_asset` writes `asset_key` and switches ON CONFLICT (deployed 2026-05-09)

Go-only change to `StoreAssetAction` in `platform/orchestration/actions/v3_site_actions.go`:

1. Extract `asset_key` from config, defaulting to `purpose` if not provided.
2. Update the INSERT statement to include the new column.
3. Switch ON CONFLICT target from `(site_id, purpose) WHERE purpose IS NOT NULL` to `(site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active'`. The WHERE clause matches Phase 2B's index exactly (Postgres requires this).
4. Add `purpose = EXCLUDED.purpose` to the SET clause — relevant only for future multi-image callers where purpose might differ between asset_key matches; backward-compat is preserved because existing callers have `asset_key = purpose` and the assignment is a no-op.
5. Update the fallback `simpleQuery` to also include `asset_key`.

Files:
- `phase_2c_store_asset_action.diff` (Phase-2C-only diff for review)
- `v3_site_actions.go` (full patched file, 3791 lines, replaces prior Phase-0.2 version)

Verified end-to-end against a seeded postgres mirroring full post-Phase-2B state:
- **Backward-compat path**: caller passes only `purpose='logo'` (no asset_key). Code defaults `asset_key='logo'`. Existing logo row upserts in place with new name/url/prompt. Single row, no duplicate.
- **Multi-image attempt**: insert hero with `asset_key='hero_about'` — succeeds. Second insert with `asset_key='hero_tools'` (same purpose='hero', different asset_key) **fails on the OLD constraint** as expected — Phase 2D hasn't dropped it yet. This is the intended transitional behaviour.
- **Cross-site**: site B can have its own `logo` independently. No collision.

**No behaviour change for production today.** Every existing caller passes `purpose` only; default-to-purpose populates `asset_key=purpose`; the new ON CONFLICT target catches the same conflicts the old one did. Phase 2C is forward-compatible scaffolding for Phase 2D.

### Phase 2B — `asset_key` column for multi-image readiness (delivered)

Added `asset_key text` (nullable) to `assets`, backfilled from `purpose`
for existing rows, and added new unique index
`idx_assets_site_asset_key_unique (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active'`.

The old `idx_assets_site_purpose_unique` constraint stays in place. Both
indexes coexist harmlessly while `asset_key = purpose` for everything,
which is the case for all rows after this migration. No behaviour change
yet — store_asset still uses the old ON CONFLICT target until Phase 2C.

Files:
- `phase_2b_pre_migration_backup.sql`
- `phase_2b_add_asset_key.sql` (two statements: BEGIN/COMMIT for ALTER +
  backfill, then `CREATE INDEX CONCURRENTLY` outside transaction)

Verified end-to-end against a seeded postgres mirroring Phase 2A's prior
state. Confirmed:
- All existing rows backfilled (4 of 5 in test, the 5th had `purpose IS NULL`).
- Both indexes coexist; neither breaks.
- Old constraint still fires on duplicate-purpose insert (Phase 2D
  behaviour pending).
- New constraint fires on duplicate asset_key when purpose is NULL.
- Phase 2A locking columns and locked rows survive the migration intact.
- Rollback works cleanly.

**What this still doesn't unblock:** the 5 hero variants. They need
Phase 2C+2D plus parameterisation of `image-build-handler`'s deploy
paths (currently hardcoded to `assets/images/hero.jpg`). Phase 2B is
scaffolding for that work; behaviour change comes in 2C+2D.

### Phase 2A — asset locking columns (delivered + applied)

Added `locked_at timestamp with time zone` and `locked_by text` to `assets`,
plus the partial index `idx_assets_locked btree (locked_at) WHERE locked_at IS NOT NULL`.
Mirrors `page_components` and `site_components` exactly. Pure DDL — no Go change.

Files:
- `phase_2a_pre_migration_backup.sql` (full table snapshot)
- `phase_2a_assets_locking.sql` (ALTER TABLE + index)

Verified end-to-end against a seeded postgres (real assets schema mirrored, 3
test rows) before applying. Schema check confirmed columns nullable as
intended; existing rows unaffected; rollback works cleanly.

#### Lock semantics — investigation, third correction

The Phase 2A docstring went through three iterations as the actual lock
model became clear.

**v1 (wrong):** Treated `locked_at` as an expiry timestamp. Came from
guessing how time-bounded locks could work without an extra column.
Doesn't match production code.

**v2 (incomplete):** Rewrote per 031_locks.md's "today's locks are all
effectively `permanent`" wording. Correct on time-based expiry (none
exists) but missed the hard-vs-soft distinction encoded in `locked_by`.

**v3 (final, applied 2026-05-08):** Reflects the production model from
`platform/orchestration/actions/check_component_lock.go`:
- Detection: `locked_at IS NULL` vs `IS NOT NULL`.
- Classification: hard (admin / admin-removed / checkpoint) vs soft
  (deploy / manual / auditor names / others) via `locked_by`.
- No time-based expiry today. Timed expiry is documented design intent
  in 004 v4 / 007 v4, paired with the audit-pass-counter reset that IS
  implemented. The lock-expiry half didn't ship.

**Lock-expiry investigation (separate document
`LOCKS_should_locks_expire.md`)** — investigated whether time-based
expiry should be implemented based on user observation that doc 031's
"permanent today" claim seemed wrong. Findings:

- Production is correctly described as permanent today (no time
  comparison anywhere in code or schema).
- Doc 004 v4 designed `lock_type` + `lock_expires_at` but only the paired
  audit-pass-reset shipped, leaving the rhythm asymmetric.
- The case for implementing timed expiry is forward-looking (no
  observed failures yet — early publishing stage).
- Approved policy (2026-05-08): implement as a future focused project
  after imagery work completes. **Human-set locks (`'admin'`,
  `'admin-removed'`, `'checkpoint'`) stay permanent**; only auto-locks
  (`'deploy'`) and auditor approvals get timed expiry.

**Doc updates landed from the investigation:**
- `phase_2a_assets_locking.sql` — v3 docstring with hard-vs-soft and
  reference to canonical Go helper.
- `031_locks_proposed_update.md` — three additions: hard-vs-soft
  subsection under Pattern A, `assets` row in "Where locks live"
  table, expanded "Lock lifecycle" section that accurately describes
  today's state and the deferred future project.
- `PLAN_imagery_loop_closure.md` — Decisions table and Phase 2.1 block
  updated to reflect the canonical model and the deferred timed-expiry
  project.
- `LOCKS_findings_and_proposed_corrections.md` — initial investigation
  notes, kept for trail.
- `LOCKS_should_locks_expire.md` — substantive analysis, recommendation,
  and approved policy. The reference document for the future project.

### Phase 1.5 hotfix — `output_mapping` added to image-build-handler

A pre-existing bug surfaced during verification: `image-build-handler.call_logo_gen`
and `call_hero_gen` were missing the `output_mapping` block that
`pageflow-builder` and `site-work-orchestrator` already had on their
image-generator calls. Without it, the generator's response dropped into
`image_result` wholesale, so the URL lived at
`image_result.response.generate.response.image_url` — three levels deeper
than `store_asset` looks. Result: silent `{stored: false, reason: "no asset URL found"}`.

Files:
- `phase_1_5_pre_migration_backup.sql`
- `phase_1_5_image_build_handler_output_mapping.sql`

The fix copies the canonical 4-field output_mapping pattern from
`pageflow-builder`. Surgical jsonb_set on two step paths.

**This bug pre-dated Phase 0/1 work** — it's been silently dropping every
image asset that flowed through `image-build-handler` (the dispatch path).
Sites that built via `pageflow-builder` directly weren't affected because
that agent had the right output_mapping. Worth knowing for any historical
audit of "why doesn't this site have a hero/logo."

---

## Findings worth recording (not blockers for Phase 2)

These all surfaced during Phase 1.5 verification and warrant their own
follow-up at some point.

### Build-dispatch-loop is not claiming triaged image items

Imagery work items emitted at 10:38 sat in `triaged` status for >4 hours
while multiple `build-dispatch-loop` orchestrations ran (14:13, 14:19, 14:40,
14:43, 14:52) and processed page items via `page-build-handler`. None
claimed the imagery items despite their `handler_agent='image-build-handler'`
and `pipeline='build'` matching what the loop's `load_work_items` should
filter on.

Possible causes (not investigated):
- Priority ordering: page items may have higher priority (lower number)
  than the 65-70 we set on imagery items.
- A filter on `LoadWorkItemsAction` we haven't traced.
- Fuel cap exhausted before reaching imagery items.

This didn't block today's verification — direct invocation of
`image-build-handler` worked fine. But it does mean imagery items don't
self-process through the normal dispatch path on this site right now.
Worth investigating separately as it's blocking organic image generation
across the system.

### Hero variant items keep getting claimed despite empty handler_agent

The 5 `unfulfilled_hero_variant` items are flag-only by design (`handler_agent=''`),
but build-dispatch-loop claimed them anyway, attempted to spawn their handler,
hit `agent_type is required (provide 'agent_type' or 'agent_type_field')`,
and after `attempt_count=3` they're now in `failed` status.

Phase 2's multi-image work makes these routable (real `handler_agent` value),
which sidesteps the issue. Until then, the noise is bounded — they're in
`failed`, won't be re-claimed, and the dedup logic prevents duplicates
on re-emit. Acceptable.

### Image generation produces `demo_client`-pathed S3 URLs

URLs follow `s3://.../images/demo_client/<date>/<uuid>.png`. The path uses
the client_id from the orchestration request (in our case `demo_client`),
not the site_id. Confirmed correct — `demo_client` is the actual client
that owns the gripper site. Not a finding; recorded for clarity.

### `origin_prompt` captures composed prompt only

Today's read in `getImageryDirectionForSite` pulls only the strategic
`site_specs.design_intent.imagery_direction`. Once Phase 1 of doc 030
ships (brief renderer + `site_plan_directives` table), this should be
extended to also pull plan-time directives where `category=design AND
subject LIKE 'imagery%'`. Tracked as a successor work item; current
single-source read is correct for today's data.

### Cosmetic — "Claim timed out — handler pod likely died"

Earlier `complete` work items have this stamped in their `error` column
even though their orchestration_states show `COMPLETED`. The reaper writes
the message before the actual completion lands. Cosmetic only.

---

## Live data state (gripper site, 2026-05-08)

### `assets` for the site
| purpose | origin_type | origin_model | status | created_at |
|---|---|---|---|---|
| logo | generated | sdxl | active | 2026-05-08 15:12 |

The hero asset doesn't exist yet — only logo was triggered today. Re-running
direct image-build-handler with `purpose=hero` and the planner's `hero_home`
prompt would produce the second row.

### `site_work_items` for the site (imagery-related)

Many items, current state mixed:

- **Recent triaged (10:38 today)**: 7 items — 2 routable (`needs_logo`,
  `needs_hero_image`), 5 flag-only (`unfulfilled_hero_variant`). All
  in `triaged`, none claimed yet by dispatch loop.
- **Older complete (07:28 yesterday)**: 2 items with the broken
  `{stored: false}` payload — historical, won't re-process.
- **Older failed (17:18 yesterday)**: 5 hero variants that fell through
  the spawn_agent error.

The work item state matches the dispatch-loop-not-claiming finding above.

---

## Open issues for tomorrow / next session

In rough priority order:

1. **Phase 2E — parameterise `image-build-handler`'s deploy paths.**
   Currently hardcoded to `assets/images/hero.jpg` and
   `assets/images/logo.png`. With Phase 2D landed, the schema accepts
   multiple rows with same `purpose` but distinct `asset_key`, but if
   five hero variants all deploy to `hero.jpg` they overwrite each other.
   The deploy step needs to derive the output filename from `asset_key`
   (e.g. `hero.jpg` for asset_key=`hero`, `hero-about.jpg` for
   asset_key=`hero_about`). This is the final piece that makes the 5
   hero variants actually routable end-to-end. Once this lands, the
   variants stop being flag-only — they get a `handler_agent` and flow
   through dispatch.

2. **Important apply order for 2A-D in production:**
   - 2A migration first (assets locking columns).
   - 2B migration next (asset_key column + new index).
   - 2C deploy (chassis binary update — store_asset writes asset_key).
   - **Verify** in production that new asset rows have asset_key populated.
     Wait long enough to be confident.
   - 2D migration last (drops old index). Sanity check guards against
     stragglers but it's still cleanest to deploy 2C before applying 2D.

   The order matters: if 2D runs while old StoreAssetAction code is
   live, every asset write fails because the old code's
   `ON CONFLICT (site_id, purpose)` has no matching index after the drop.

3. **Investigate build-dispatch-loop's claim ordering.** Why are imagery
   items in `triaged` status not being claimed when page items are? May
   be a priority issue, may be something else. Standalone investigation.

4. **Trigger the hero asset for the gripper site** to complete the matched
   pair, either via direct image-build-handler call (as we did for logo)
   or by waiting for dispatch to claim once issue 3 is resolved.

5. **Lock-expiry project (deferred).** Adds `lock_type` + `lock_expires_at`
   columns uniformly across all four Pattern A tables (page_components,
   site_components, site_plan_directives, assets). Restores the rhythm
   doc 004 v4 envisaged. Approved policy: human-set locks stay permanent;
   only `'deploy'` and auditor approvals get timed expiry. Sequenced
   after the imagery loop work completes. Reference document:
   `LOCKS_should_locks_expire.md`.

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

Phase 1.5 hotfix (deployed today):
- `phase_1_5_pre_migration_backup.sql`
- `phase_1_5_image_build_handler_output_mapping.sql`

Phase 2A deliverables (delivered today, not yet applied to production):
- `phase_2a_pre_migration_backup.sql`
- `phase_2a_assets_locking.sql` (final v3 docstring)

Phase 2B deliverables (delivered today, not yet applied to production):
- `phase_2b_pre_migration_backup.sql`
- `phase_2b_add_asset_key.sql`

Phase 2C deliverables (deployed to production 2026-05-09):
- `phase_2c_store_asset_action.diff` (Phase-2C-only diff for review)
- `v3_site_actions.go` (full patched file, deployed via chassis pipeline; verified byte-identical to production)

Phase 2D deliverables (delivered today, not yet applied):
- `phase_2d_pre_migration_backup.sql`
- `phase_2d_drop_purpose_constraint.sql` (with sanity checks and `ON_ERROR_STOP`)

Phase 2E deliverables (delivered today, not yet deployed):
- `phase_2e/check_unfulfilled_image_prompt.go` (classifier — variants routable)
- `phase_2e/imagery_helpers.go` (added `hasActiveAssetForAssetKey`)
- `phase_2e/phase_2e_deploy_image_asset_action.diff` (asset_key handling + variant path derivation)
- `phase_2e/phase_2e_store_asset_action.diff` (asset_key_field support)
- `phase_2e/phase_2e_pre_migration_backup.sql`
- `phase_2e/phase_2e_image_build_handler_variant_path.sql` (workflow extension)

Lock investigation (today):
- `LOCKS_findings_and_proposed_corrections.md` (initial notes)
- `LOCKS_should_locks_expire.md` (substantive analysis + approved policy)
- `031_locks_proposed_update.md` (proposed canonical doc updates)
- `PROPOSED_031_locks_addendum.md` (earlier `assets`-row addition, now folded into the proposed update above)

Planning and assessment:
- `FOCUS_imagery_assessment.md`
- `PLAN_imagery_loop_closure.md`
- `ASSESSMENT_phase_0_1_vs_phase_1_architecture.md`
- `ADDENDUM_phase_0_1_verification.md`
- `STATUS_imagery_2026-05-06.md` (prior status)
- `STATUS_imagery_2026-05-08.md` (this file)

Diagnostics from the verification work (kept for future reference):
- `find_image_url_path.sql`
- `phase_1_5_*_diagnostics*.sql`
- `improvement_loop_status_check.sql`
- various others

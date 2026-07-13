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
| 2A | Asset locking columns | n/a | **✅ today** | **✅ today** | **n/a (scaffolding)** |
| 2B | Add `asset_key` column + new unique index | n/a | **✅ today** | **n/a (forward-compat)** | **n/a (no behaviour change)** |
| 2C | `store_asset` writes `asset_key`; switch ON CONFLICT | **✅ today** | n/a | **n/a (forward-compat)** | **n/a (no behaviour change)** |
| 2D | Drop old `(site_id, purpose)` unique constraint | — | — | — | — |
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

### Phase 2C — `store_asset` writes `asset_key` and switches ON CONFLICT (delivered)

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

1. **Phase 2D — drop old `(site_id, purpose)` unique constraint.** After
   Phase 2C is bedded in (a few generations through the loop with no
   issues), drop `idx_assets_site_purpose_unique`. Single statement
   migration. After this, the multi-image case is unblocked at the
   schema level.

2. **Phase 2E — parameterise `image-build-handler`'s deploy paths.**
   Currently hardcoded to `assets/images/hero.jpg` and
   `assets/images/logo.png`. Once 2D is in, multiple hero variants
   would all overwrite `hero.jpg` unless the deploy path is keyed off
   asset_key (e.g. `assets/images/hero-about.jpg` for asset_key='hero_about').
   This is the final piece that makes the 5 hero variants actually
   routable end-to-end.

3. **Investigate build-dispatch-loop's claim ordering.** Why are imagery
   items in `triaged` status not being claimed when page items are? May
   be a priority issue, may be something else. Standalone investigation.

4. **Trigger the hero asset for the gripper site** to complete the matched
   pair, either via direct image-build-handler call (as we did for logo)
   or by waiting for dispatch to claim once issue 3 is resolved.

5. **Consider whether the hero variants should keep firing**. They currently
   produce flag-only items every time discovery runs. Once Phase 2D+2E
   land, they become routable; before that, they're recorded but produce
   no action. Possibly worth gating them behind a feature flag until then,
   but the noise is currently bounded.

6. **Lock-expiry project (deferred).** Adds `lock_type` + `lock_expires_at`
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

Phase 2C deliverables (delivered today, not yet deployed):
- `phase_2c_store_asset_action.diff` (Phase-2C-only diff for review)
- `v3_site_actions.go` (full patched file, replaces post-Phase-0.2 version)

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

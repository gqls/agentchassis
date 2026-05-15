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
| 2 | Asset locking + multi-image readiness | — | — | — | — |
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

1. **Phase 2 — asset locking and multi-image readiness.** The natural
   next step. Adds `locked_at`/`lock_type`/`asset_key` to `assets`,
   relaxes the unique constraint, parameterises `image-build-handler`'s
   deploy paths. Once these are in place, the 5 hero variants become
   routable.

2. **Investigate build-dispatch-loop's claim ordering.** Why are imagery
   items in `triaged` status not being claimed when page items are? May
   be a priority issue, may be something else. Standalone investigation.

3. **Trigger the hero asset for the gripper site** to complete the matched
   pair, either via direct image-build-handler call (as we did for logo)
   or by waiting for dispatch to claim once issue 2 is resolved.

4. **Consider whether the hero variants should keep firing**. They currently
   produce flag-only items every time discovery runs. Once Phase 2 lands,
   they become routable; before that, they're recorded but produce no
   action. Possibly worth gating them behind a feature flag until Phase 2
   is ready, but the noise is currently bounded.

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

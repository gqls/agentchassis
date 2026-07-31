-- 281 — the improvement loop decides "clean" from the SITE's state, not from
--       whichever copy of the promoter happened to run last
--
-- bugs_open/150. The second half of the fix; the first half is Go
-- (triage_detect_items_action.go + work_items_common.go, committed 2026-07-31).
--
-- WHY. `triage_detected_items` is a step in THREE live agents — improvement-loop
-- (`triage_findings`), design-audit-agent (`triage`) and site-review-agent
-- (`triage`) — and the promotion is unconditional over the site (site_id +
-- status, no type filter). The parent calls both children BEFORE running its own
-- copy, so the first copy to run takes every row and the parent's copy honestly
-- reports `promoted: 0`. `check_has_findings` reads that last writer, takes
-- else_step, and lands on `complete_clean` — whose success message is "No issues
-- found — site is clean" — skipping insert_rerender_item → spawn_dispatch →
-- call_dispatch.
--
-- Measured twice, not once:
--   * orchestration 30692439 (gamesdesign.co.uk, 2026-07-29): 67 promoted by the
--     design-audit child, parent `promoted: 0`, terminal step `complete_clean`.
--   * orchestration 911ecdd8 (vetcomparison.uk, 2026-07-31, fired as the control
--     for THIS change): site started at 0 detected; the discovery agents created
--     24; the design-audit child promoted all 24 (`{"promoted": 24, "has_items":
--     true}`); the parent's own copy promoted 0. Same terminal step.
--   The bug file recorded "this happens on every run" as [INFERRED from a single
--   run] because orchestration_states retention had cleared all history. The
--   second observation discharges that marker.
--
-- WHAT THE GO HALF ADDED. Two site-scoped keys beside the existing call-scoped
-- ones: `site_dispatchable` (bool) and `site_dispatchable_count` (int). They
-- count site_work_items in a dispatchable status for the target pipeline —
-- whoever promoted them, in whatever order, including a fourth caller that does
-- not exist yet. `has_items` is deliberately UNCHANGED: it is a fleet-wide
-- convention meaning "my own result set was non-empty", with three other live
-- consumers (build-dispatch-loop.check_has_items, site-work-orchestrator's two)
-- that read it correctly about their own loaders.
--
-- ══════════════════════════════════════════════════════════════════════════
-- ORDER IS LOAD-BEARING. DO NOT APPLY THIS UNTIL THE IMAGE IS LIVE.
-- ══════════════════════════════════════════════════════════════════════════
-- WHICH IS WHY THE FILENAME ENDS `_HOLD`. `run-migrations.sh --apply` takes
-- EVERY pending file in the directory, in order — so a banner asking a human not
-- to apply this yet protects nothing against the next bulk run by another
-- session, and there are ~8 unrecorded files in here right now. The runner's
-- SIDECAR_RE ('_[A-Z][A-Z0-9_]*\.sql$') excludes it from applying while still
-- LISTING it under "Sidecars (hand-run only)", so it is held back visibly rather
-- than hidden.
--
-- TO APPLY, once the pod-grep below passes on every replica:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user \
--     -d clients_db < docs/agent_docs/sql_for_agents/281_..._HOLD.sql
--   ./scripts/migration/run-migrations.sh --record-only \
--     281_improvement_loop_branches_on_site_state.sql --note "bugs_open/150, applied by hand after the image landed"
-- (or `git mv` the `_HOLD` off and let the runner take it — either is fine, but
--  the ledger is only written by --record-only, and recording stays a human act.)
-- The key is emitted by the BINARY; this condition lives in the DB, which is
-- live immediately. Applied against a chassis that predates the Go half,
-- `triage_result.site_dispatchable` resolves to nil → compareValues(nil,"true")
-- is false → else_step → complete_clean, EVERY run, including the runs that
-- promote. That is strictly worse than the bug: today the branch is at least
-- right when the parent's own copy promotes something. Pinned as a unit test —
-- TestConditionOnAPreUpgradeBinaryDoesNotSilentlyInvert — so the ordering rule
-- is a property of the code and not only a sentence in this header.
--
-- Gate on the RUNNING POD, on every replica, never on the tag and never on git
-- (a roll is not evidence your fix shipped — bugs_open/153). The positive
-- control in the same exec is what makes a 0 meaningful:
--
--   for P in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
--                -o jsonpath='{.items[*].metadata.name}'); do
--     echo -n "$P "
--     kubectl -n ai-persona-system exec "$P" -- sh -c \
--       'strings /app/agent-chassis | grep -c "site_dispatchable"; \
--        strings /app/agent-chassis | grep -c "TriageDetectedItemsAction: Starting"'
--   done
--
--   first  = the symbol THIS change added        → must be >= 1 on EVERY replica
--   second = a pre-existing symbol (the control) → if this is 0 the grep itself
--            is broken and the first number means nothing
--
-- WHY NOT FIX CANDIDATE 1 ("one triage, one owner"). The bug file ranks it
-- first and it is the wrong first move: removing the `triage` step from
-- design-audit-agent and site-review-agent requires auditing every OTHER parent
-- that calls those two agents, and leaves the same defect available to the next
-- agent that gains a triage step. A site-scoped signal makes the ordering
-- irrelevant instead of making one ordering mandatory — the bad state stops
-- being representable rather than stopping being reachable by today's callers.
-- Both children keep their triage steps, deliberately.
--
-- SURFACE: this is a CONFIG_CHANGE to the owning pipeline **improvement-loop**,
-- delivered as a migration.

BEGIN;

-- ── STEP 1 — PRE-FLIGHT ASSERTION ─────────────────────────────────────────
-- Derived mechanically HERE, at apply time, not carried in from the
-- investigation. Expect exactly 1. Anything else means the row is not the shape
-- this migration was written against — stop and re-read it.
SELECT count(*) AS rows_to_change_expect_1
FROM agent_definitions
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,check_has_findings,config,condition}'
      = 'triage_result.has_items == true';

-- ── STEP 2 — SNAPSHOT ─────────────────────────────────────────────────────
-- The platform's own versioning mechanism, not a hand-rolled pre-image.
SELECT snapshot_agent('improvement-loop',
                      'bugs_open/150 — check_has_findings branches on site state (281)');

-- ── STEP 3 — THE CHANGE ───────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,check_has_findings,config,condition}',
      '"triage_result.site_dispatchable == true"'::jsonb,
      false   -- create_missing=false: if the key is absent the step is not the
              -- shape this migration was written against; fail loud rather than
              -- inventing a condition next to whatever is really there.
    ),
    updated_at = now()
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,check_has_findings,config,condition}'
      = 'triage_result.has_items == true';   -- idempotent + refuses a drifted row

-- ── STEP 4 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- Expect exactly one row reading site_dispatchable. If it still reads has_items,
-- the WHERE guard refused the row: ROLLBACK and re-read step 1.
SELECT type,
       default_config #>> '{workflow,steps,check_has_findings,config,condition}' AS condition_after,
       default_config #>> '{workflow,steps,check_has_findings,config,then_step}' AS then_step,
       default_config #>> '{workflow,steps,check_has_findings,config,else_step}' AS else_step
FROM agent_definitions
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ── ROLLBACK ──
-- Preferred: restore the snapshot taken in step 2 (that is what it is for).
-- Direct inverse, if the snapshot path is unavailable:
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(default_config,
--       '{workflow,steps,check_has_findings,config,condition}',
--       '"triage_result.has_items == true"'::jsonb, false)
-- WHERE type='improvement-loop'
--   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── VERIFY THE FIX AT THE ARTEFACT, not at this row ──
-- Fire ONE sweep at a site that has work, and read the RUN. A green build proves
-- nothing: the branch under test is the one the build path never reaches.
--
--   ./docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/\
--     run_improvement_sweep_once.sh <site_id> <domain>
--   (read its blast-radius header first — one firing promotes and dispatches
--    every detected item on that site)
--
--   -- the loop must NOT claim clean while the site has work
--   SELECT current_step, status,
--          collected_data->'triage_result'->>'promoted'                AS parent_promoted,
--          collected_data->'triage_result'->>'site_dispatchable'       AS site_dispatchable,
--          collected_data->'triage_result'->>'site_dispatchable_count' AS site_count
--   FROM orchestration_states WHERE owner_agent_type='improvement-loop'
--   ORDER BY created_at DESC LIMIT 1;
--   -- expect current_step='complete', NOT 'complete_clean', with
--   -- parent_promoted='0' and site_dispatchable='true' — that pair IS the bug,
--   -- now routing the other way.
--
--   -- and the closing rerender the clean branch was skipping must exist
--   SELECT id, item_type, priority, created_at FROM site_work_items
--   WHERE site_id='<site>' AND item_key LIKE 'improvement_rerender%'
--     AND created_at > '<run start>';
--
-- KNOWN SIBLING, NOT FIXED HERE (recorded in bugs_open/150): check_audit_pass_limit
-- routes straight to complete_clean when get_audit_pass_count(site) >= 3, so a
-- capped site is also told "site is clean" when what happened is "we skipped
-- auditing" — and its detected pile is never promoted at all. Latent today:
-- 0 of 25 sites are at the limit (measured 2026-07-31). Separate decision,
-- separate change; it needs an honest terminal step, not this condition.

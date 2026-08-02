-- 286 — one owner for the site-wide promoter: remove `triage_detected_items`
--       from design-audit-agent and site-review-agent
--
-- RFC 006, OWNER RULING 2026-08-02: option (a), "one promoter, one owner".
-- docs/agent_docs/docs024_key_docs_latest/architecture_review/
--   RFC_006_one_promoter_one_owner_for_shared_promoting_steps.md
--
-- WHY. `triage_detected_items` promotes EVERY `detected` row on a site —
-- `WHERE site_id = $1 AND status = 'detected'`, no type filter, no ownership
-- filter. It was a step in THREE live agents. The improvement loop calls both
-- children before its own copy, so the first copy took every row and the
-- parent's copy honestly reported `promoted: 0` — which its branch read as
-- "nothing to do", ending the run on "No issues found — site is clean" over a
-- site holding 67 promoted findings (bugs_closed/150, orchestrations 30692439
-- and 911ecdd8).
--
-- 150 fixed the SYMPTOM at the shared action (it now also returns
-- `site_dispatchable`, the site's state, correct whoever promoted). The council's
-- `architecture` and `bug_historian` seats independently said that did not close
-- the CLASS, and were right: it left the next agent to gain a triage step
-- inheriting two candidate signals and a scoping rule living in one file's
-- header. This migration is the owner's answer.
--
-- WHAT CHANGES. Three edits per child, both children:
--   1. repoint the predecessor's next_step from `triage` to `complete`
--   2. delete the `triage` step
--   3. drop `triage_result` from `complete.config.output_fields`
-- (3) is not cosmetic: ApplyResultSpec's ResultModeFields silently skips an
-- absent field, so leaving it would be harmless AND permanently misleading — a
-- published contract naming a field the workflow can no longer produce.
--
-- ══════════════════════════════════════════════════════════════════════════
-- NO ORDERING CONSTRAINT — this one is safe to apply on its own, unlike 281.
-- ══════════════════════════════════════════════════════════════════════════
-- It removes steps; it depends on no new binary affordance. The Go half shipped
-- in the same commit is a DETECTOR (`SingleOwner` + --single-owner-actions),
-- which is offline, read-only, and reports on the fleet rather than running in
-- it. Applying this before or after that image is live changes nothing about
-- behaviour. Hence no `_HOLD` suffix, deliberately — the suffix means "the
-- config half will break production if it lands early", and misusing it for
-- ordinary caution would devalue it where it is load-bearing.
--
-- WHY THIS IS NOT A REGRESSION — the four checks, run 2026-08-02 before writing:
--
--  (1) No other parent calls either child. Scanned every live definition for a
--      step naming them: exactly TWO rows, both improvement-loop
--      (spawn_design_audit, spawn_site_review). There is no second caller whose
--      behaviour changes.
--  (2) Neither child has ever run standalone. agent_run_stats (78 agents
--      tracked, oldest entry 2026-07-26): improvement-loop 3 /
--      design-audit-agent 3 / site-review-agent 3 — identical counts over an
--      identical window; a standalone child run would break the 1:1:1.
--      Corroborated in orchestration_state_audit (3-day retention): 2 / 2 / 2.
--      0 rows in scheduled_tasks name either type; orchestration_requests empty.
--      [CAVEAT, stated because it is the soft edge] only 3 loop runs exist
--      all-history, because improvement-sweep is disabled (bugs_open/083). So
--      this is "no standalone caller today", firmly — not "none has ever
--      existed". Check (1) does not depend on run history at all, and it is
--      check (1) that decides the question.
--  (3) Neither child branches on the result. Both are `triage -> complete`
--      (live config), so removing the step changes no control flow inside them.
--  (4) The parent's own copy runs on the ONLY path the children run on. The
--      loop is linear from spawn_quality_discovery through call_site_review ->
--      triage_findings; the one branch that skips the children
--      (check_audit_pass_limit -> notify_scheduler_clean) skips the children
--      too. So there is no path on which a child promotes and the parent does
--      not.
--
--  Residual, recorded rather than hidden: if the loop DIES between a child
--  running and `triage_findings`, findings that a child would have promoted now
--  wait. Nothing is lost — the promoter is greedy, so the next sweep takes them
--  — it is latency on a failed run, not loss. Accepted as the smaller cost
--  against a false "clean" on every healthy run.
--
-- THE DURABLE HALF IS THE GO COMMIT, NOT THIS FILE. Removing two steps is a
-- one-off that the next agent to gain one silently undoes. `SingleOwner: true`
-- on TriageDetectedItemsInputSpec plus scripts/audit-single-owner-actions.sh
-- makes the second copy loud, fleet-wide, offline. Verified armed against the
-- LIVE fleet before this migration was applied: the detector reported the
-- 3-carrier violation naming all three steps. Re-run it after applying — it
-- must report none.
--
-- SURFACE: CONFIG_CHANGE to design-audit-agent and site-review-agent. The owning
-- pipeline of the effect is improvement-loop, which is untouched by this file.

BEGIN;

-- ── STEP 1 — PRE-FLIGHT ASSERTION ─────────────────────────────────────────
-- Derived mechanically HERE, at apply time. Expect exactly 2 — one per child,
-- each still carrying the promoter behind the predecessor this migration knows
-- about. Anything else means the rows are not the shape this was written
-- against: stop and re-read them.
SELECT count(*) AS rows_to_change_expect_2
FROM agent_definitions ad
WHERE ad.type IN ('design-audit-agent', 'site-review-agent')
  AND ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
  AND ad.default_config #>> '{workflow,steps,triage,action}' = 'triage_detected_items'
  AND ad.default_config #>> '{workflow,steps,triage,next_step}' = 'complete';

-- Second assertion: the improvement loop must still carry its own copy. If this
-- returns 0, applying the file would leave the fleet with NO promoter at all —
-- strictly worse than the fan-out. Expect exactly 1.
SELECT count(*) AS owner_still_has_the_step_expect_1
FROM agent_definitions
WHERE type = 'improvement-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,triage_findings,action}' = 'triage_detected_items';

-- ── STEP 2 — SNAPSHOT ─────────────────────────────────────────────────────
-- The platform's own versioning mechanism. Two-arg form -> a true point-in-time
-- copy in agent_definitions_backup (see 281's header for why the overload
-- matters).
SELECT snapshot_agent('design-audit-agent',
                      'RFC 006 owner ruling — remove the duplicate site-wide promoter (286)');
SELECT snapshot_agent('site-review-agent',
                      'RFC 006 owner ruling — remove the duplicate site-wide promoter (286)');

-- ── STEP 3a — design-audit-agent ──────────────────────────────────────────
-- call_content_auditor -> triage -> complete   becomes
-- call_content_auditor -> complete
UPDATE agent_definitions ad
SET default_config = jsonb_set(
      (jsonb_set(ad.default_config,
                 '{workflow,steps,call_content_auditor,next_step}',
                 '"complete"'::jsonb,
                 false)   -- create_missing=false: fail loud rather than inventing a key
       #- '{workflow,steps,triage}'),
      '{workflow,steps,complete,config,output_fields}',
      (SELECT COALESCE(jsonb_agg(e ORDER BY ord), '[]'::jsonb)
       FROM jsonb_array_elements(ad.default_config #> '{workflow,steps,complete,config,output_fields}')
            WITH ORDINALITY AS t(e, ord)
       WHERE e <> '"triage_result"'::jsonb),
      false),
    updated_at = now()
WHERE ad.type = 'design-audit-agent'
  AND ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
  -- idempotent + refuses a drifted row: on a second run `triage` is gone, so
  -- this matches nothing and the UPDATE is a clean no-op.
  AND ad.default_config #>> '{workflow,steps,triage,action}' = 'triage_detected_items'
  AND ad.default_config #>> '{workflow,steps,call_content_auditor,next_step}' = 'triage';

-- ── STEP 3b — site-review-agent ───────────────────────────────────────────
-- write_strategic_findings -> triage -> complete   becomes
-- write_strategic_findings -> complete
UPDATE agent_definitions ad
SET default_config = jsonb_set(
      (jsonb_set(ad.default_config,
                 '{workflow,steps,write_strategic_findings,next_step}',
                 '"complete"'::jsonb,
                 false)
       #- '{workflow,steps,triage}'),
      '{workflow,steps,complete,config,output_fields}',
      (SELECT COALESCE(jsonb_agg(e ORDER BY ord), '[]'::jsonb)
       FROM jsonb_array_elements(ad.default_config #> '{workflow,steps,complete,config,output_fields}')
            WITH ORDINALITY AS t(e, ord)
       WHERE e <> '"triage_result"'::jsonb),
      false),
    updated_at = now()
WHERE ad.type = 'site-review-agent'
  AND ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
  AND ad.default_config #>> '{workflow,steps,triage,action}' = 'triage_detected_items'
  AND ad.default_config #>> '{workflow,steps,write_strategic_findings,next_step}' = 'triage';

-- ── STEP 4 — VERIFY BEFORE COMMIT ─────────────────────────────────────────
-- (i) Exactly ONE live agent may still carry the promoter, and it must be
--     improvement-loop. This is the whole point of the migration, asserted
--     rather than assumed.
SELECT ad.type AS still_carries_the_promoter, step.key AS step_name
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS step
WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
  AND step.value->>'action' = 'triage_detected_items'
ORDER BY 1;
-- expect exactly one row: improvement-loop | triage_findings

-- (ii) Both children must now run straight through to `complete`, with no
--      dangling next_step pointing at a step that no longer exists — a dangling
--      pointer would strand every run at that step.
SELECT ad.type,
       ad.default_config #>> '{workflow,steps,call_content_auditor,next_step}'     AS design_audit_next,
       ad.default_config #>> '{workflow,steps,write_strategic_findings,next_step}' AS site_review_next,
       ad.default_config #>  '{workflow,steps,complete,config,output_fields}'      AS output_fields,
       (ad.default_config #> '{workflow,steps}') ? 'triage'                        AS triage_still_present
FROM agent_definitions ad
WHERE ad.type IN ('design-audit-agent', 'site-review-agent')
  AND ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
ORDER BY 1;
-- expect: triage_still_present = false on both; the relevant next = 'complete';
-- output_fields no longer containing "triage_result".

-- (iii) Nothing anywhere may still POINT at the deleted step. Catches a
--       predecessor this migration did not know about — the failure mode a
--       targeted repoint cannot see by construction.
SELECT ad.type, step.key AS step_pointing_at_a_deleted_step
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS step
WHERE ad.type IN ('design-audit-agent', 'site-review-agent')
  AND ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
  AND (step.value->>'next_step' = 'triage'
       OR step.value->>'error_step' = 'triage'
       OR step.value->'config'->>'then_step' = 'triage'
       OR step.value->'config'->>'else_step' = 'triage');
-- expect ZERO rows. Any row here means a run will strand: ROLLBACK.

COMMIT;

-- ── ROLLBACK ──
-- Preferred: restore the snapshots taken in step 2 — that is what they are for.
--   SELECT snapshot_taken_at, snapshot_reason FROM agent_definitions_backup
--   WHERE type IN ('design-audit-agent','site-review-agent')
--   ORDER BY snapshot_taken_at DESC LIMIT 2;
-- Direct inverse, if the snapshot path is unavailable (restores BOTH children):
-- UPDATE agent_definitions ad SET default_config = jsonb_set(
--       jsonb_set(ad.default_config, '{workflow,steps,triage}',
--         '{"action":"triage_detected_items","config":{"site_id":"site_record.site_id","target_domain":"build"},"next_step":"complete","output_field":"triage_result"}'::jsonb, true),
--       CASE WHEN ad.type='design-audit-agent'
--            THEN '{workflow,steps,call_content_auditor,next_step}'
--            ELSE '{workflow,steps,write_strategic_findings,next_step}' END,
--       '"triage"'::jsonb, false)
-- WHERE ad.type IN ('design-audit-agent','site-review-agent')
--   AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
-- (output_fields would also need "triage_result" appending back.)
--
-- ── VERIFY AT THE ARTEFACT, not at this row ──
-- The rows above prove the config changed. They do not prove a run still works —
-- the risk of a step deletion is a stranded run, which only a real run shows.
--
--   ./docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/\
--     run_improvement_sweep_once.sh <site_id> <domain>
--   (read its blast-radius header first — one firing promotes and dispatches
--    every detected item on that site)
--
--   -- both children must reach 'complete', not strand mid-workflow
--   SELECT current_step, status,
--          collected_data->'design_audit_result'->>'status' AS design_audit,
--          collected_data->'site_review_result'->>'status'  AS site_review,
--          collected_data->'triage_result'->>'promoted'                AS parent_promoted,
--          collected_data->'triage_result'->>'site_dispatchable_count' AS site_count
--   FROM orchestration_states
--   WHERE collected_data ? 'triage_result' ORDER BY created_at DESC LIMIT 1;
--   -- Now that the parent is the ONLY promoter, `parent_promoted` should be
--   -- NON-ZERO on a site with detected findings — the reverse of bugs_closed/150,
--   -- where it was always 0. That flip IS the proof this migration worked.
--
--   -- and the fleet-wide check must now be silent
--   ./scripts/audit-single-owner-actions.sh    # expect "none", exit 0

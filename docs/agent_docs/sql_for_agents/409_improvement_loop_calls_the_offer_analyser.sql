-- 409_improvement_loop_calls_the_offer_analyser.sql
--
-- APPLIED 2026-08-15 ON THE OWNER'S ENROLMENT CALL. This file was
-- `409_…_HOLD.sql` from 2026-08-14 until then — held per PLAN §B5 ("enrolment
-- order = owner calls at the time") and renamed off the suffix, per its own
-- instructions, the day the owner said go. Trial-run (COMMIT→ROLLBACK) clean
-- against the live row immediately before applying.
--
-- WHY IT WAS HELD, and it was not caution for its own sake:
--   PLAN_2026-08-02 §B5 says in as many words "enrolment order = owner calls at
--   the time", and this migration IS the enrolment — it puts the offer analyser
--   into EVERY improvement sweep, on all 22 sites, at once. Two costs the owner
--   should price rather than discover:
--     1. ONE MORE LLM CALL PER SWEPT SITE. The prompt is 30–47KB (measured
--        2026-08-14) and the completion runs ~3–4k tokens. The fleet hit its
--        Anthropic spend cap at 16:37 on the day this was written — 28 runs
--        failed `invalid_request_error: usage limits` between 15:36 and 16:42,
--        five of them `site-review-agent` — so an extra call per site is a real
--        budget decision, not a rounding error. It IS bounded: the audit-due
--        gate (fingerprint changed OR 14-day cooldown) means an unchanged site
--        in cooldown is skipped entirely.
--     2. FINDINGS ARE NOT PARKABLE. `triage_detect_items_action.go:161-173`
--        promotes every `detected` row on a site the loop reaches — no type
--        filter, no ownership filter. Enrolment therefore means roughly 5 new
--        dispatchable items per swept site, arriving at page-build-handler and
--        component-template-fixer. The item types are all EXISTING ones with
--        live drains (content_rewrite 122 rows, needs_content_page 46,
--        cta_improvement 36, nav_restructure 21 in the last 30 days), so the
--        novelty is volume, not mechanism.
--   The analyser itself is proven: two hand-fired runs, both COMPLETED, 5
--   findings → 5 work items each, degraded arm exercised, protected spec row
--   left byte-identical (BIZ-032, NOTES 2026-08-14 evening). What is NOT yet
--   witnessed is one B4 finding travelling detected → triaged → claimed →
--   complete. IMP-016's order ("one clean cycle, then enable") wants that first,
--   and the five gaswholesalers items are sitting `detected` waiting for it.
--
-- WHAT IT DOES: inserts the analyser between the strategic review and the
-- audit-pass record, on the AUDIT-DUE branch only — which is the "due-gated
-- inside improvement-loop before triage_findings" of PLAN §B4.
--
--   … → spawn_site_review → call_site_review
--        → spawn_offer_analyser → call_offer_analyser
--        → record_audit_pass → triage_findings → …
--
-- `call_offer_analyser` carries `error_step = record_audit_pass`, matching every
-- other audit call in this loop: one auditor's failure must not strand a sweep.
-- That is a deliberate swallow at the loop level, and it is safe here only
-- because the failure is fully visible in `orchestration_states` for the child
-- run — the loop is not the record of whether B4 worked.
--
-- NOT touched: `triage_findings`. Migration 286 made improvement-loop the SINGLE
-- OWNER of promotion (RFC 006 / bugs_closed/150) and this file adds no second
-- triage step. Run `./scripts/audit-single-owner-actions.sh` after applying —
-- it must stay clean.
--
-- SNAPSHOT FIRST (RUNBOOK convention, `bak_ad_<agent>_<date>`): taken inside
-- this transaction, so a rollback of the migration and the snapshot are atomic.
--
-- ROLLBACK: 409_improvement_loop_calls_the_offer_analyser_ROLLBACK.sql
-- restores the two edited pointers and drops the two added steps. (Verified
-- 2026-08-15, pre-apply: BOTH arms of call_site_review read record_audit_pass
-- live, so the rollback's restore of next_step AND error_step is exact.)

BEGIN;

-- 1. Snapshot the live row before editing it.
CREATE TABLE IF NOT EXISTS bak_ad_improvementloop_20260814 AS
SELECT * FROM agent_definitions
WHERE type = 'improvement-loop' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 2. Guard: the shape this migration assumes must actually be there.
DO $guard$
DECLARE
  steps jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO steps
  FROM agent_definitions
  WHERE type = 'improvement-loop' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF steps IS NULL THEN
    RAISE EXCEPTION 'guard: no live improvement-loop row';
  END IF;
  IF NOT (steps ? 'call_site_review') THEN
    RAISE EXCEPTION 'guard: call_site_review is gone — the chain this migration splices into has changed shape, re-read it before applying';
  END IF;
  IF (steps->'call_site_review'->>'next_step') IS DISTINCT FROM 'record_audit_pass' THEN
    RAISE EXCEPTION 'guard: call_site_review.next_step is % , expected record_audit_pass — another session has already edited this seam', (steps->'call_site_review'->>'next_step');
  END IF;
  IF steps ? 'call_offer_analyser' THEN
    RAISE EXCEPTION 'guard: already applied (call_offer_analyser present)';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM agent_definitions
                  WHERE type = 'offer-analyser' AND is_active
                    AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL) THEN
    RAISE EXCEPTION 'guard: offer-analyser agent row absent — migration 408 must be applied first, or the call step targets nothing';
  END IF;
END $guard$;

-- 3. Add the two steps and re-point the seam, in one jsonb write.
UPDATE agent_definitions SET
  default_config = jsonb_set(
    default_config,
    '{workflow,steps}',
    (default_config->'workflow'->'steps')
      || jsonb_build_object(
           'spawn_offer_analyser', jsonb_build_object(
             'action', 'spawn_agent',
             'config', jsonb_build_object(
               'role', 'offer_analyser',
               'agent_type', 'offer-analyser'
             ),
             'next_step', 'call_offer_analyser',
             'description', 'Spawn the offer/benefit analyser (B4)',
             'output_field', 'offer_analyser_spawned'
           ),
           'call_offer_analyser', jsonb_build_object(
             'action', 'call_agent',
             'config', jsonb_build_object(
               'target_role', 'offer_analyser',
               'input_mapping', jsonb_build_object(
                 'domain', 'site_record.domain',
                 'site_id', 'site_record.site_id'
               ),
               'timeout_seconds', 600
             ),
             'next_step', 'record_audit_pass',
             'error_step', 'record_audit_pass',
             'description', 'Offer/benefit analysis: writes the ranked offer_ordering spec and files artefact-vs-premise findings. error_step continues the sweep — one auditor must not strand it, and the child run is the record of whether it worked',
             'output_field', 'offer_analysis_result'
           )
         )
      || jsonb_build_object(
           'call_site_review',
           jsonb_set(
             jsonb_set(default_config->'workflow'->'steps'->'call_site_review',
                       '{next_step}', '"spawn_offer_analyser"'),
             '{error_step}', '"spawn_offer_analyser"')
         )
  ),
  updated_at = now()
WHERE type = 'improvement-loop' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 4. Verify the whole spliced chain, not just the keys.
DO $verify$
DECLARE
  steps jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps' INTO steps
  FROM agent_definitions
  WHERE type = 'improvement-loop' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF (steps->'call_site_review'->>'next_step')  IS DISTINCT FROM 'spawn_offer_analyser'
  OR (steps->'call_site_review'->>'error_step') IS DISTINCT FROM 'spawn_offer_analyser' THEN
    RAISE EXCEPTION 'verify: call_site_review does not hand on to spawn_offer_analyser on BOTH arms';
  END IF;
  IF (steps->'spawn_offer_analyser'->>'next_step') IS DISTINCT FROM 'call_offer_analyser' THEN
    RAISE EXCEPTION 'verify: spawn_offer_analyser does not lead to call_offer_analyser';
  END IF;
  IF (steps->'call_offer_analyser'->>'next_step')  IS DISTINCT FROM 'record_audit_pass'
  OR (steps->'call_offer_analyser'->>'error_step') IS DISTINCT FROM 'record_audit_pass' THEN
    RAISE EXCEPTION 'verify: call_offer_analyser must rejoin record_audit_pass on both arms, or a sweep can be stranded by one auditor';
  END IF;
  IF (steps->'spawn_offer_analyser'->'config'->>'agent_type') IS DISTINCT FROM 'offer-analyser' THEN
    RAISE EXCEPTION 'verify: spawn targets the wrong agent_type';
  END IF;
  -- The role string is the join between spawn and call; a typo here spawns an
  -- agent nothing ever calls, and the loop continues looking healthy.
  IF (steps->'spawn_offer_analyser'->'config'->>'role')
       IS DISTINCT FROM (steps->'call_offer_analyser'->'config'->>'target_role') THEN
    RAISE EXCEPTION 'verify: spawn role and call target_role disagree — the call would wait on an agent that was never spawned under that role';
  END IF;
  -- Promotion ownership is untouched (migration 286 / RFC 006).
  IF (steps->'record_audit_pass'->>'next_step') IS DISTINCT FROM 'triage_findings' THEN
    RAISE EXCEPTION 'verify: the tail of the audit branch no longer reaches triage_findings';
  END IF;
END $verify$;

COMMIT;

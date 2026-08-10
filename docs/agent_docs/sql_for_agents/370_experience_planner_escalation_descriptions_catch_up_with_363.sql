-- 370_experience_planner_escalation_descriptions_catch_up_with_363.sql
--
-- bugs_open/227, second defect, follow-up to migration 363. Config only — live on
-- apply, no image, no roll. No behaviour changes here at all: three prose strings
-- inside the `experience-planner` row still describe the graph 363 replaced, and
-- one of them asserts the exact state 363 made unrepresentable.
--
-- WHY IT MATTERS ENOUGH TO BE ITS OWN FILE. 363 moved the plan write onto the
-- council's approved branch, and it did update `persist_plan` and `check_approved`
-- to say so. It missed three:
--
--   complete_escalated.description
--     "The current (rejected) plan stays is_current but MUST NOT be built until
--      resolved"
--     FALSE since 363: an escalated run persists NOTHING, so there is no row from
--     it to be is_current and nothing to warn a reader off building. A future
--     session reading this would go looking for a rejected plan of record, and
--     the honest conclusion from not finding one — "the demotion already
--     happened" — is wrong in the other direction.
--
--   complete_escalated.config.success_message  ("do NOT build the current plan")
--     Same claim, and this one travels: it is the message the escalation hands
--     back to whoever reads the run.
--
--   recompose.description  ("loops back to persist + review + decide")
--     Names an edge that no longer exists; recompose now hands to review_journeys.
--
-- This is the drift class the concept register's own landmine describes — a
-- status line that outlives its truth and is then read as ground truth. It costs
-- four strings to fix now and an investigation to fix later.
--
-- VERIFIED BEHAVIOURALLY FIRST, 2026-08-10 (corr d81aa5f4-a732-4fb3-b438-4ff496ef7ba2):
-- a real council VETO (feasibility, round 1) produced NO doc_plans row — the count
-- for the probe subject stayed 0 through compose, the veto and the whole reframe
-- round, and reached 1 only after check_approved routed to persist on round 2. So
-- these strings are being corrected against an observation, not against a reading
-- of the graph. See the loancalculator_couk lane's HANDOFF_2026-08-10b.
--
-- ROLLBACK: 370_..._ROLLBACK.sql restores from agent_definitions_backup, picking
-- by snapshot_taken_at DESC with snapshot_reason LIKE 'pre-update: 227 escalation
-- descriptions%'. Every backup row for one agent shares the source row's id and
-- created_at, so ordering by created_at returns an arbitrary snapshot — order by
-- snapshot_taken_at (LANDMINES 2026-07-30).

-- ============================================================================
-- Probe guard: refuse a second application.
-- ============================================================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'experience-planner'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,complete_escalated,description}' NOT LIKE '%stays is_current%'
    ) THEN
        RAISE EXCEPTION '227/370: already applied — complete_escalated.description no longer carries the stale is_current claim';
    END IF;
END $$;

-- ============================================================================
-- Drift guard: 363 must be live, and the three strings must be the ones this
-- file was composed against. A raise means another session edited them — re-read
-- and recompose rather than forcing.
-- ============================================================================
DO $$
DECLARE
    a_then text; p_next text; esc_desc text; esc_msg text; rec_desc text;
BEGIN
    SELECT default_config #>> '{workflow,steps,check_approved,config,then_step}',
           default_config #>> '{workflow,steps,persist_plan,next_step}',
           default_config #>> '{workflow,steps,complete_escalated,description}',
           default_config #>> '{workflow,steps,complete_escalated,config,success_message}',
           default_config #>> '{workflow,steps,recompose,description}'
      INTO a_then, p_next, esc_desc, esc_msg, rec_desc
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF a_then IS DISTINCT FROM 'persist_plan' OR p_next IS DISTINCT FROM 'complete' THEN
        RAISE EXCEPTION '227/370: expected 363 live (check_approved.then=persist_plan, persist.next=complete), found %, %',
                        a_then, p_next;
    END IF;
    IF esc_desc NOT LIKE '%stays is_current%' OR esc_msg NOT LIKE '%do NOT build the current plan%'
       OR rec_desc NOT LIKE '%loops back to persist%' THEN
        RAISE EXCEPTION '227/370 drift: one of the three strings is not what this file was composed against';
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('experience-planner',
    'pre-update: 227 escalation descriptions catch up with 363 (loancalculator_couk lane)');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           jsonb_set(default_config,
             '{workflow,steps,complete_escalated,description}',
             '"ESCALATED: no approvable plan within max_rounds. NOTHING IS PERSISTED on this path (migration 363) — doc_plans holds approved plans only, so there is no row from this run to demote and none to build from. The plan the council refused lives in this run''s collected_data->proposal->>result and in llm_call_log, by correlation_id; the disagreement is the round-boundary decision menu. Read council_report artifacts + doc_notes (experience-council)."'::jsonb, false),
           '{workflow,steps,complete_escalated,config,success_message}',
           '"experience-planner ESCALATED: council did not converge and NO plan was written — doc_plans is unchanged by this run. Surface the disagreement (council_report by correlation_id) as the decision menu; the refused draft is in this run''s collected_data, not in doc_plans."'::jsonb, false),
         '{workflow,steps,recompose,description}',
         '"Revise the plan to address every objection + missing item; loops back to review + decide. Writes the same proposal field compose and reframe write, and persists nothing — only an approved round reaches persist_plan (migration 363)."'::jsonb, false)
 WHERE type = 'experience-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- Verify inside the transaction. Assert the NEW text is in place AND that the
-- old claims are gone from the whole row — a jsonb_set on a wrong path silently
-- ADDS a key rather than failing, so "my string is somewhere in the config" is
-- not the same as "the stale string is gone".
-- ============================================================================
DO $$
DECLARE
    esc_desc text; esc_msg text; rec_desc text; n_stale int;
BEGIN
    SELECT default_config #>> '{workflow,steps,complete_escalated,description}',
           default_config #>> '{workflow,steps,complete_escalated,config,success_message}',
           default_config #>> '{workflow,steps,recompose,description}'
      INTO esc_desc, esc_msg, rec_desc
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF esc_desc NOT LIKE '%NOTHING IS PERSISTED%' OR esc_msg NOT LIKE '%NO plan was written%'
       OR rec_desc NOT LIKE '%loops back to review + decide%' THEN
        RAISE EXCEPTION '227/370 verify: a replacement string did not land (esc_desc=%, esc_msg=%, rec_desc=%)',
                        left(esc_desc, 60), left(esc_msg, 60), left(rec_desc, 60);
    END IF;

    -- Whole-row sweep: none of the retired claims may survive anywhere in the
    -- config, including in a key this file did not think to look at.
    SELECT count(*) INTO n_stale
      FROM agent_definitions
     WHERE type = 'experience-planner'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND (default_config::text LIKE '%stays is_current%'
         OR default_config::text LIKE '%do NOT build the current plan%'
         OR default_config::text LIKE '%loops back to persist%');
    IF n_stale <> 0 THEN
        RAISE EXCEPTION '227/370 verify: a retired claim still appears in default_config';
    END IF;
END $$;

COMMIT;

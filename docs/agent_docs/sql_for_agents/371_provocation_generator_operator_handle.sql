-- 371_provocation_generator_operator_handle.sql
--
-- Seats `provocation-generator-manual`: generate candidate provocations with a
-- real model, then judge them with the real gate, in one dispatch.
--
-- WHY THIS EXISTS
-- `generate_provocations` has been registered, tested, council-approved
-- (`bbbc9fca8`) and live in the fleet since v1.0.1280 — and has **never been run
-- against a real model**. Every provocation in the live pool was written by a
-- person or by an assistant session at the owner's request; `source_ref` on the
-- six LLM-marked rows says so in as many words ("not generator-produced, not
-- gate-judged"). So the generative half of the pipeline is, to date, entirely
-- unevidenced, and the pool runs dry on 2026-08-15.
--
-- The owner's standing constraint is that the framework writes the content, not a
-- session. That constraint is unsatisfiable while the generator has no seat: there
-- is no way to invoke it. This file is that seat.
--
-- WHY OPERATOR-INVOKED FIRST, WITH NO `scheduled_tasks` ROW
-- 321 records that generation is *schedulable* in principle — it can only ever
-- write undated drafts, so a runaway generator produces a pile of drafts and no
-- site change. That remains true and a schedule is the right end state, because
-- the disease this lane is treating is a pool that ran dry for thirteen days, not
-- a single empty shelf.
--
-- It is not the right FIRST state. A model call that has never once been made
-- should not make its debut unattended: the failure modes that matter here are
-- silent (a reply that parses but is off-corpus, a gate that approves nothing, a
-- batch of near-duplicate slugs) and none of them raise an error. Prove the path
-- with one attended run, then schedule it in a follow-up migration with the
-- observed yield as the argument for the cadence. Seating it with a schedule
-- today would fire it before anyone has seen its output.
--
-- WHAT IT CANNOT DO
--   generate -> writes `status='draft'`, `publish_on` NULL, `human_approved_at` NULL
--   gate     -> may set a verdict and `status='approved'`; writes neither date nor stamp
-- The publisher requires `status='approved' AND publish_on IS NOT NULL AND
-- human_approved_at IS NOT NULL` (mig 320). Neither step in this workflow can
-- write the last two, so no path through this agent can put text on the site.
-- The owner's approval and the operator-invoked scheduler (321) remain in front of
-- publication, exactly as owner ruling 2026-08-09 left them.
--
-- MODEL: `claude-sonnet-5` for BOTH steps, deliberately.
-- The gate step must match what the §10.6 calibration measured (mig 319 pins the
-- same string) or the calibration proves nothing about this run. The generator
-- step is uncalibrated whatever it is pinned to — there is no corpus of "good
-- generations" to calibrate against — so it takes the same model rather than an
-- untried one, and the first run's output is the evidence. A stronger model is a
-- reasonable later experiment; it is not a thing to change in the same run that
-- establishes the baseline.
--
-- COUNT: 8. Enough that the gate's rejections still leave a usable shelf, small
-- enough that a bad first run wastes one dispatch rather than a long backlog the
-- owner then has to read through.
--
-- Idempotent; safe to re-run.

BEGIN;

INSERT INTO agent_definitions (type, display_name, category, description, default_config, is_active)
SELECT
    'provocation-generator-manual',
    'Provocation Generator (operator-invoked)',
    'content',
    'Generate candidate provocations with a real model and judge them with the real gate, in one dispatch. Writes drafts only; cannot date or stamp, so it cannot publish. OPERATOR-INVOKED while the generative half is unevidenced — a schedule follows once one attended run has been read.',
    jsonb_build_object(
        'processing_mode', 'task',
        'workflow', jsonb_build_object(
            'start_step', 'generate',
            'steps', jsonb_build_object(
                'generate', jsonb_build_object(
                    'action', 'generate_provocations',
                    'description', 'Write candidate provocations into the pool as drafts',
                    'next_step', 'gate',
                    'config', jsonb_build_object(
                        'domain', 'vonc.com',
                        'site_id', '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74',
                        'count', 8,
                        'ai_service', jsonb_build_object(
                            'provider', 'anthropic',
                            'model', 'claude-sonnet-5',
                            'api_key_env_var', 'ANTHROPIC_API_KEY'
                        )
                    )
                ),
                'gate', jsonb_build_object(
                    'action', 'gate_provocation',
                    'description', 'Judge every ungated draft under vonc.com',
                    'next_step', 'complete',
                    'config', jsonb_build_object(
                        'domain', 'vonc.com',
                        'limit', 40,
                        'ai_service', jsonb_build_object(
                            'provider', 'anthropic',
                            'model', 'claude-sonnet-5',
                            'api_key_env_var', 'ANTHROPIC_API_KEY'
                        )
                    )
                ),
                'complete', jsonb_build_object(
                    'action', 'complete_workflow',
                    'description', 'Generation round finished'
                )
            )
        )
    ),
    true
WHERE NOT EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'provocation-generator-manual'
       AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
);

DO $$
DECLARE n_agent int; n_sched int; n_gate_model text; n_cal_model text;
BEGIN
  SELECT count(*) INTO n_agent FROM agent_definitions
   WHERE type='provocation-generator-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF n_agent <> 1 THEN RAISE EXCEPTION 'expected exactly 1 active generator agent, found %', n_agent; END IF;

  -- Operator-invoked for now, and the absence is the control while the path is
  -- unevidenced. When this assertion is deliberately removed, the commit that
  -- removes it must carry the run whose output justified the schedule.
  SELECT count(*) INTO n_sched FROM scheduled_tasks WHERE target_agent_type='provocation-generator-manual';
  IF n_sched <> 0 THEN
    RAISE EXCEPTION 'a scheduled_tasks row exists for provocation-generator-manual (%). This seat is operator-invoked until one attended run has been read.', n_sched;
  END IF;

  -- The gate step must judge with the model the calibration measured. If 319 is
  -- ever re-pinned and this is not, the calibration silently stops being evidence
  -- about what actually runs — which is the trap mig 319's own header names.
  SELECT default_config->'workflow'->'steps'->'gate'->'config'->'ai_service'->>'model'
    INTO n_gate_model FROM agent_definitions
   WHERE type='provocation-generator-manual' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  SELECT default_config->'workflow'->'steps'->'gate'->'config'->'ai_service'->>'model'
    INTO n_cal_model FROM agent_definitions
   WHERE type='provocation-gate-calibration' AND is_active
     AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF n_cal_model IS NOT NULL AND n_gate_model IS DISTINCT FROM n_cal_model THEN
    RAISE EXCEPTION 'gate model %  does not match the calibrated model % — the calibration is not evidence about this run', n_gate_model, n_cal_model;
  END IF;

  RAISE NOTICE 'generator seated as operator-invoked; 0 scheduled_tasks rows; gate model % matches the calibration', n_gate_model;
END $$;

COMMIT;

-- HOW TO INVOKE: dispatch agent_type 'provocation-generator-manual' on
-- system.agent.generic.requests (envelope: see 097_TRIGGER; one-line payload, since
-- kcat -P splits stdin on newlines into separate messages).
--
-- HOW TO READ THE RESULT — the run's own status is not the answer:
--   SELECT slug, status, gated_at, gate_verdict->'reasons' FROM provocations
--    WHERE domain='vonc.com' AND source='llm' AND source_ref LIKE 'anthropic/%'
--    ORDER BY created_at DESC;
-- `inserted: 0` with `generated: 8` means every slug already existed (the insert is
-- ON CONFLICT DO NOTHING), not that the model failed. An approved row still cannot
-- publish: it needs the owner's `human_approved_at` stamp and then the operator
-- scheduler (321) to date it.

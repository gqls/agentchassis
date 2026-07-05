-- 0NN_wire_persist_diagnosis_note.sql — Stage 3 wiring.
-- DRAFT 2026-07-04. Renumber 0NN. Requires 0NN_doc_plans_and_notes.sql applied
-- and persist_diagnosis_note deployed (it is, per Stage 2).
--
-- Inserts a persist_note step into diagnose-agent's default_config workflow,
-- BETWEEN emit and complete. Verified against the live workflow (this turn):
--   emit:     action diagnose_emit,     output_field "diagnosis", next_step "complete"
--   complete: action complete_workflow, config.result_from "diagnosis"
-- So: emit.next_step -> persist_note; persist_note.next_step -> complete.
-- Because persist runs BEFORE complete, result_from "diagnosis" is unaffected —
-- the caller (diagnose-orchestrator) still receives the same diagnosis result.
--
-- Column: default_config (LIVE). The deprecated orchestrator_workflow copy is
-- left untouched. Targets the CURRENT version row (type + max version,
-- deleted_at IS NULL), matching the versioned agent_definitions pattern —
-- NOT a blind UPDATE by type.
--
-- Step config is empty: persist_diagnosis_note's InputSpec defaults already read
-- diagnosis.status/summary/conclusion/stopped_by/evidence_trail and the
-- input_data.* subject fields. NOTE: diagnose-agent's input contract does not yet
-- carry subject_type/subject_key, so the step will SKIP (persisted:false) until a
-- subject-aware caller supplies them — safe and inert by design (skip-don't-guess).

BEGIN;

WITH current_diag AS (
    SELECT id
    FROM agent_definitions
    WHERE type = 'diagnose-agent'
      AND deleted_at IS NULL
    ORDER BY version DESC
    LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        jsonb_set(
            -- 1) add the new persist_note step
            ad.default_config,
            '{workflow,steps,persist_note}',
            '{
                "action": "persist_diagnosis_note",
                "config": {},
                "next_step": "complete",
                "description": "Persist the diagnosis as a NOTES entry when the run has an explicit subject (skips otherwise). Never a fix.",
                "output_field": "diagnosis_note"
             }'::jsonb,
            true
        ),
        -- 2) redirect emit -> persist_note (was "complete")
        '{workflow,steps,emit,next_step}',
        '"persist_note"'::jsonb,
        true
    ),
    updated_at = now()
FROM current_diag
WHERE ad.id = current_diag.id;

-- Guard: exactly one row updated.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'diagnose-agent' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,emit,next_step}' = 'persist_note'
      AND default_config #>> '{workflow,steps,persist_note,action}' = 'persist_diagnosis_note';
    IF n <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 wired diagnose-agent, found %', n;
    END IF;
END $$;

COMMIT;

-- Verify after apply:
--   SELECT jsonb_pretty(default_config #> '{workflow,steps,emit}'),
--          jsonb_pretty(default_config #> '{workflow,steps,persist_note}')
--   FROM agent_definitions
--   WHERE type='diagnose-agent' AND deleted_at IS NULL
--   ORDER BY version DESC LIMIT 1;
--
-- Rollback (manual):
--   UPDATE agent_definitions
--   SET default_config =
--         (default_config #- '{workflow,steps,persist_note}')
--         || jsonb_set('{}'::jsonb,'{noop}','"noop"')  -- placeholder if needed
--   ... then set emit.next_step back to "complete":
--   UPDATE agent_definitions
--   SET default_config = jsonb_set(default_config #- '{workflow,steps,persist_note}',
--                                  '{workflow,steps,emit,next_step}', '"complete"'::jsonb, true),
--       updated_at = now()
--   WHERE type='diagnose-agent' AND deleted_at IS NULL;

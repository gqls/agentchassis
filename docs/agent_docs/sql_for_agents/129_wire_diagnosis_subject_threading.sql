-- 0NN_wire_diagnosis_subject_threading.sql — Stage 3b: make persist_note write.
-- DRAFT 2026-07-06. Renumber 0NN. Requires: 3a closed (it is); no deploy needed
-- (default_config / input_contract are DB-only, effective immediately).
--
-- Grounded in the 3b.1 paste (2026-07-06):
--   diagnose-orchestrator.call_diagnoser.config.input_mapping has NINE keys
--   (ref?/repo?/owner?/symptom/site_id?/seed_scope?/runtime_page?/
--    runtime_site?/correlation_id?) — no subject fields;
--   BOTH input_contracts are identical: required ["symptom"], eight optionals,
--   no subject fields.
--
-- Three edits (016b §9 spawn+call rule: input_mapping must satisfy the
-- input_contract — so the callee's contract must declare what the mapping
-- sends, and the entry-point contract should declare what the trigger sends):
--   1. orchestrator input_mapping += "subject_type?"/"subject_key?"
--      (the `?` suffix matches the existing optional-mapping convention;
--       object-merge PRESERVES the nine existing keys);
--   2. orchestrator input_contract.optional += subject_type, subject_key;
--   3. diagnose-agent input_contract.optional += subject_type, subject_key.
-- Array appends use a DISTINCT aggregate so re-running cannot duplicate
-- (element order in `optional` is not significant).
--
-- After this, a subject-carrying run (e.g. SUBJECT_TYPE=pipeline
-- SUBJECT_KEY=build via 084) reaches persist_note with input_data.subject_*
-- set and writes the FIRST machine-written doc_notes row. For a smoke symptom
-- the loop's status is expected UNVERIFIABLE, so the row will be tagged
-- unconfirmed-diagnosis — a dead-end entry BY DESIGN, not a fault.

BEGIN;

-- 1+2) diagnose-orchestrator: mapping merge + contract optionals (one row).
WITH cur AS (
    SELECT id
    FROM agent_definitions
    WHERE type = 'diagnose-orchestrator'
      AND deleted_at IS NULL
    ORDER BY version DESC
    LIMIT 1
)
UPDATE agent_definitions ad
SET default_config = jsonb_set(
        ad.default_config,
        '{workflow,steps,call_diagnoser,config,input_mapping}',
        (ad.default_config #> '{workflow,steps,call_diagnoser,config,input_mapping}')
            || '{"subject_type?": "input_data.subject_type",
                 "subject_key?":  "input_data.subject_key"}'::jsonb,
        true
    ),
    input_contract = jsonb_set(
        ad.input_contract,
        '{optional}',
        (SELECT jsonb_agg(DISTINCT v)
         FROM jsonb_array_elements_text(
                ad.input_contract->'optional'
                || '["subject_type","subject_key"]'::jsonb) AS t(v))
    ),
    updated_at = now()
FROM cur
WHERE ad.id = cur.id
  AND ad.default_config #> '{workflow,steps,call_diagnoser,config,input_mapping}' IS NOT NULL;

-- 3) diagnose-agent: contract optionals (one row).
WITH cur AS (
    SELECT id
    FROM agent_definitions
    WHERE type = 'diagnose-agent'
      AND deleted_at IS NULL
    ORDER BY version DESC
    LIMIT 1
)
UPDATE agent_definitions ad
SET input_contract = jsonb_set(
        ad.input_contract,
        '{optional}',
        (SELECT jsonb_agg(DISTINCT v)
         FROM jsonb_array_elements_text(
                ad.input_contract->'optional'
                || '["subject_type","subject_key"]'::jsonb) AS t(v))
    ),
    updated_at = now()
FROM cur
WHERE ad.id = cur.id;

-- Guards: exact mapping paths + both contracts carry both keys.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'diagnose-orchestrator' AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,call_diagnoser,config,input_mapping,subject_type?}' = 'input_data.subject_type'
      AND default_config #>> '{workflow,steps,call_diagnoser,config,input_mapping,subject_key?}'  = 'input_data.subject_key'
      AND input_contract->'optional' ? 'subject_type'
      AND input_contract->'optional' ? 'subject_key';
    IF n <> 1 THEN
        RAISE EXCEPTION 'orchestrator subject threading incomplete (found %)', n;
    END IF;

    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'diagnose-agent' AND deleted_at IS NULL
      AND input_contract->'optional' ? 'subject_type'
      AND input_contract->'optional' ? 'subject_key';
    IF n <> 1 THEN
        RAISE EXCEPTION 'diagnose-agent contract missing subject optionals (found %)', n;
    END IF;
END $$;

COMMIT;

-- Verify after apply (expect the two mapping paths + t/t on both contracts):
--   SELECT default_config #>> '{workflow,steps,call_diagnoser,config,input_mapping,subject_type?}' AS map_type,
--          default_config #>> '{workflow,steps,call_diagnoser,config,input_mapping,subject_key?}'  AS map_key,
--          input_contract->'optional' ? 'subject_type' AS c_type,
--          input_contract->'optional' ? 'subject_key'  AS c_key
--   FROM agent_definitions
--   WHERE type='diagnose-orchestrator' AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
--
--   SELECT input_contract->'optional' ? 'subject_type' AS c_type,
--          input_contract->'optional' ? 'subject_key'  AS c_key
--   FROM agent_definitions
--   WHERE type='diagnose-agent' AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
--
-- Rollback (manual): remove the two mapping keys and filter the two values
-- back out of each contract's optional array:
--   UPDATE agent_definitions SET default_config =
--     (default_config #- '{workflow,steps,call_diagnoser,config,input_mapping,subject_type?}')
--      #- '{workflow,steps,call_diagnoser,config,input_mapping,subject_key?}',
--     updated_at = now()
--   WHERE type='diagnose-orchestrator' AND deleted_at IS NULL;
--   -- and for each of the two types:
--   UPDATE agent_definitions SET input_contract = jsonb_set(input_contract,'{optional}',
--     (SELECT jsonb_agg(v) FROM jsonb_array_elements_text(input_contract->'optional') AS t(v)
--      WHERE v NOT IN ('subject_type','subject_key')), true), updated_at = now()
--   WHERE type IN ('diagnose-agent','diagnose-orchestrator') AND deleted_at IS NULL;

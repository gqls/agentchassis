-- 103_call_data_preparer_optional_inputs.sql
--
-- APPLY TO: templates_db (NOT clients_db). agent_definitions is the source of
-- truth in templates_db (002_system_architecture.md; 011_database_and_infrastructure.md).
--   kubectl -n ai-persona-system exec -it postgres-templates-0 \
--     -- psql -U templates_user -d templates_db -f 103_call_data_preparer_optional_inputs.sql
-- A first apply to clients_db (2026-06-03) had no effect — the chassis reads templates_db.
--
-- WHAT THIS FIXES
-- The model-trainer orchestration chain dies at its first call step,
-- call_data_preparer, with:
--   input_mapping failed: source path 'input_data.orchestration_id' not found
--                          for field 'orchestration_id'
-- (coordinator.go:1534; content_search.go:96 part=orchestration_id
--  available_keys=[hyperparameters, export_id]). Confirmed live in orchestration
-- 23863e2e-3090-4231-980c-6d73746b60e3 (2026-06-02): all three spawns resolved,
-- then call_data_preparer (step 1cfdb238) failed BEFORE dispatching any work
-- request, so the preparer never ran and no training_runs row was inserted.
--
-- WHY
-- call_data_preparer is a call_agent step. call_agent builds the child's
-- input_data via input_contracts.ResolveInputMapping, which HARD-FAILS when a
-- mapped source path is absent UNLESS the destination field name ends in '?'
-- (input_mapping.go L101-128). The step maps four fields:
--     export_id        <- input_data.export_id        (present)
--     hyperparameters  <- input_data.hyperparameters  (present)
--     orchestration_id <- input_data.orchestration_id (ABSENT on manual/most triggers)
--     triggered_by     <- input_data.triggered_by     (ABSENT)
-- with NO '?' markers, so the two absent sources are treated as required and the
-- whole step fails before any work request reaches the data-preparer.
--
-- The preparer ITSELF already treats these two as OPTIONAL
-- (PrepareTrainingDataInputSpec.Optional = [triggered_by, orchestration_id];
-- prepare_training_data_action.go), reads them through NullableString, and
-- training_runs.triggered_by / orchestration_id are nullable UUIDs
-- (019_model_lifecycle_schema.sql). So the workflow contract over-declared
-- required-ness relative to the action contract. This aligns the two.
--
-- THE FIX
-- Mark the two mappings optional with the '?' destination-field suffix — the
-- same convention call_launcher already uses (instance_ip?, ssh_user?,
-- ssh_key_secret_name?). When the source is absent, ResolveInputMapping skips
-- the field; the '?' is stripped (TrimSuffix) before the field is placed in the
-- child input_data, so the preparer still reads them by their bare names and
-- writes NULL when not supplied. export_id + hyperparameters are unchanged and
-- remain required.
--
-- VARIABLE-NAME NOTE (intentional rename): this renames two input_mapping KEYS
-- (orchestration_id -> orchestration_id?, triggered_by -> triggered_by?). The
-- child-visible field names are UNCHANGED — '?' is a resolver-only marker that
-- is trimmed off. No Go variable names change; the preparer keeps reading
-- input_data.export_id / .hyperparameters / .triggered_by / .orchestration_id,
-- and ExtractActionInputs (spec-driven) treats the latter two as optional.
--
-- SCOPE / MECHANISM
-- The chain lives in agent_definitions.default_config -> 'workflow' (verified:
-- call_data_preparer appears only in default_config, column 6). In-place
-- jsonb_set on the exact definition that executed (id 94f5a069-...), no version
-- bump, so a stable definition id is picked up by the next orchestrate.
-- CAVEAT: if the chassis caches agent_definitions in memory, a rollout of
-- agent-chassis may be needed for the change to take effect — confirm on the
-- next run rather than assuming the edit is hot.
--
-- PROVENANCE NOTE: with these optional and not supplied by a manual trigger,
-- training_runs.orchestration_id / triggered_by will be NULL for such runs.
-- Acceptable for iter_0. A follow-up (NOT in this migration) could have the
-- preparer default orchestration_id from its parent_orchestration_id.

BEGIN;

-- Before (expect orchestration_id / triggered_by with no '?'):
SELECT id, type, version, is_active,
       default_config #> '{workflow,steps,call_data_preparer,config,input_mapping}'
         AS before_input_mapping
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_data_preparer,config,input_mapping}',
        jsonb_build_object(
            'export_id',         'input_data.export_id',
            'hyperparameters',   'input_data.hyperparameters',
            'orchestration_id?', 'input_data.orchestration_id',
            'triggered_by?',     'input_data.triggered_by'
        ),
        false  -- do not create the path if missing; fail loud instead
    ),
    updated_at = NOW()
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30'
  AND type = 'model-trainer';

-- After (expect orchestration_id? and triggered_by? keys):
SELECT id, type, version, is_active,
       default_config #> '{workflow,steps,call_data_preparer,config,input_mapping}'
         AS after_input_mapping
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

COMMIT;

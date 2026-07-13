-- 104_provisioner_output_fields_and_launcher_mapping.sql
--
-- APPLY TO: clients_db (NOT templates_db). The flywheel-C agent_definitions
-- (rich schema with version/is_snapshot) are read by the chassis from clients_db;
-- templates_db holds only the old website-builder catalog.
--   <your clients_db psql> -f 104_provisioner_output_fields_and_launcher_mapping.sql
--
-- WHAT / WHY
-- Two in-place edits, no version bump (chassis loads definitions per-orchestrate,
-- so no restart needed):
--
--   1) gpu-provisioner (0bf9fa8a) complete step: replace the NON-STANDARD singular
--      `output_field: provision_response` with the standard `output_fields:
--      ["dispatch_provision"]`. extractWorkflowResult only honours the plural
--      `output_fields`; the singular key was silently ignored, dropping the agent
--      into the fallback dump ({dispatch_provision, input_data, ...}). The provision
--      result lands under the STEP NAME `dispatch_provision` (await storage keys by
--      step name; `provision_response` is never a collected key), so that is the
--      field we surface. This also drops the stray input_data echo.
--
--   2) model-trainer (94f5a069) call_launcher input_mapping: re-point the four
--      provisioning fields from `provisioning_result.<field>` to
--      `provisioning_result.dispatch_provision.<field>`. These resolve via the same
--      `.response` auto-unwrap that already makes `preparation_result.dataset_uri`
--      work (dispatch_provision is the immediate child once provisioning_result.response
--      is unwrapped). dataset_uri / training_run_id / hyperparameters are unchanged.
--
-- NOTE: the gpu-provisioner dispatch step keeps `output_field: provision_response`
-- (cosmetic for await results; the result keys by step name regardless). Left as-is
-- to keep this change minimal.

BEGIN;

-- ── BEFORE ──
SELECT 'gpu-provisioner complete.config BEFORE' AS label,
       default_config #> '{workflow,steps,complete,config}' AS value
FROM public.agent_definitions
WHERE id = '0bf9fa8a-925c-4ab5-9287-2c8e5d7b9451';

SELECT 'model-trainer call_launcher.input_mapping BEFORE' AS label,
       default_config #> '{workflow,steps,call_launcher,config,input_mapping}' AS value
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

-- ── 1) gpu-provisioner: singular output_field → standard output_fields ──
UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config}',
        '{"output_fields": ["dispatch_provision"]}'::jsonb,
        false
    ),
    updated_at = now()
WHERE id = '0bf9fa8a-925c-4ab5-9287-2c8e5d7b9451'
  AND default_config #> '{workflow,steps,complete,config}' IS NOT NULL;

-- ── 2) model-trainer: re-point call_launcher provisioning fields ──
UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_launcher,config,input_mapping}',
        '{
            "dataset_uri": "preparation_result.dataset_uri",
            "training_run_id": "preparation_result.training_run_id",
            "hyperparameters": "input_data.hyperparameters",
            "provisioning_id": "provisioning_result.dispatch_provision.provisioning_id",
            "instance_ip?": "provisioning_result.dispatch_provision.instance_ip",
            "ssh_user?": "provisioning_result.dispatch_provision.ssh_user",
            "ssh_key_secret_name?": "provisioning_result.dispatch_provision.ssh_key_secret_name"
        }'::jsonb,
        false
    ),
    updated_at = now()
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30'
  AND default_config #> '{workflow,steps,call_launcher,config,input_mapping}' IS NOT NULL;

-- ── AFTER ──
SELECT 'gpu-provisioner complete.config AFTER' AS label,
       default_config #> '{workflow,steps,complete,config}' AS value
FROM public.agent_definitions
WHERE id = '0bf9fa8a-925c-4ab5-9287-2c8e5d7b9451';

SELECT 'model-trainer call_launcher.input_mapping AFTER' AS label,
       default_config #> '{workflow,steps,call_launcher,config,input_mapping}' AS value
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

COMMIT;

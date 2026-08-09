-- ============================================================================
-- 356_retire_dead_config_keys_commit_from_and_hitl_output_format_ROLLBACK.sql
--
-- Reverses 356: puts both dead keys back, with their exact prior values, read
-- from the live rows before 356 was written.
--
-- Behaviour-neutral in BOTH directions, which is the whole point of the pair:
-- neither key is read by the action that carries it (UpdatePageStatusAction
-- reads five config keys and commit_from is not one; ProcessApprovalDecisionAction
-- reads only stop_on_reject), so restoring them restores config text and nothing
-- else. The one OBSERVABLE difference is that a chassis carrying the
-- ActionInputSpec opt-in from the same commit will resume logging an
-- unrecognised-config-key warning for these seven steps. That warning is the
-- correct report of a restored dead key, not a fault.
--
-- Hand-run only — the UPPERCASE _ROLLBACK suffix keeps run-migrations.sh from
-- ever applying this file (SIDECAR_RE).
--
-- NOT a substitute for the snapshots. 356 calls two-arg snapshot_agent for all
-- seven agents first, writing agent_definitions_backup rows tagged
-- '356_retire_dead_config_keys: pre-update'. If anything other than these two
-- keys changed, restore from those instead, ordering by snapshot_taken_at DESC
-- (NOT created_at — every backup row for one agent shares the source row's
-- created_at, so ordering by it returns an arbitrary snapshot).
-- ============================================================================

BEGIN;

-- jsonb_set with create_if_missing (the default) puts the key back, but it is a
-- SILENT NO-OP if the PARENT path is absent — so assert the parents first
-- rather than trusting a row count that would look identical either way.
DO $$
DECLARE
  n_bad int;
BEGIN
  SELECT count(*) INTO n_bad FROM (VALUES
      ('pageflow-builder',      '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config}'::text[]),
      ('page-rebuild',          '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config}'::text[]),
      ('page-rerender',         '{workflow,steps,update_status,config}'::text[]),
      ('report-builder',        '{workflow,steps,update_status,config}'::text[]),
      ('section-editor',        '{workflow,steps,update_page_status,config}'::text[]),
      ('site-work-orchestrator','{workflow,steps,build_items_loop,config,sub_workflow,steps,update_page_status,config}'::text[]),
      ('simple-content-writer-with-approval', '{workflow,steps,process_approval,config}'::text[])
    ) AS t(agent_type, cfg_path)
    LEFT JOIN agent_definitions ad
      ON ad.type = t.agent_type AND ad.is_active
     AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
   WHERE ad.id IS NULL
      OR jsonb_typeof(ad.default_config #> t.cfg_path) IS DISTINCT FROM 'object';

  IF n_bad > 0 THEN
    RAISE EXCEPTION '356 ROLLBACK: % target step(s) do not resolve to a config object — the restore would be a silent no-op. Restore from agent_definitions_backup instead.', n_bad;
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config,commit_from}',
         '"page_deployed.commit_sha"'::jsonb),
       updated_at = now()
 WHERE type = 'pageflow-builder' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config,commit_from}',
         '"page_deployed.commit_sha"'::jsonb),
       updated_at = now()
 WHERE type = 'page-rebuild' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,update_status,config,commit_from}',
         '"deploy_result.commit_sha"'::jsonb),
       updated_at = now()
 WHERE type = 'page-rerender' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,update_status,config,commit_from}',
         '"deploy_result.commit_sha"'::jsonb),
       updated_at = now()
 WHERE type = 'report-builder' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,update_page_status,config,commit_from}',
         '"git_result.commit_sha"'::jsonb),
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,build_items_loop,config,sub_workflow,steps,update_page_status,config,commit_from}',
         '"page_deployed.commit_sha"'::jsonb),
       updated_at = now()
 WHERE type = 'site-work-orchestrator' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- The HITL map, verbatim as it stood before 356.
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_approval,config,output_format}',
         '{"content": "{{.generate_draft.result}}",
           "approved_at": "{{.await_human_approval.timestamp}}",
           "approved_by": "{{.await_human_approval.approved_by}}",
           "approval_status": "{{.await_human_approval.approved}}",
           "approval_comments": "{{.await_human_approval.comments}}"}'::jsonb),
       updated_at = now()
 WHERE type = 'simple-content-writer-with-approval' AND is_active
   AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- Verify the restore actually landed, DO/RAISE rather than bare SELECTs.
DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n
    FROM agent_definitions ad
   WHERE ad.is_active AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
     AND ad.default_config::text LIKE '%commit_from%';
  IF n <> 6 THEN
    RAISE EXCEPTION '356 ROLLBACK VERIFY: expected 6 agents carrying commit_from, found %', n;
  END IF;

  SELECT count(*) INTO n
    FROM agent_definitions ad
   WHERE ad.type = 'simple-content-writer-with-approval' AND ad.is_active
     AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
     AND jsonb_typeof(ad.default_config #> '{workflow,steps,process_approval,config,output_format}') = 'object';
  IF n <> 1 THEN
    RAISE EXCEPTION '356 ROLLBACK VERIFY: HITL output_format map not restored as an object (matched % rows)', n;
  END IF;
END $$;

COMMIT;

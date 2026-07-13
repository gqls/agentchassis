-- 0NN_recreation_spec_and_note_subject.sql — two tool-recreation-handler fixes.
-- DRAFT 2026-07-09. Renumber 0NN. DB-only; effective immediately, no deploy.
-- Both faults were found by READING the definition, not by a failed run.
--
-- (i)  spec is UNDECLARED. input_contract = required [site_id, domain],
--      optional [page_name, page_id, sections]. The workflow nonetheless reads
--      input_data.spec.mode, input_data.spec.interactive_features and (via the
--      Task-4 note tail) input_data.spec.function. Same class as the Stage-3b
--      finding: an input the workflow depends on must be declared. FIX: add
--      'spec' to optional (idempotent — skipped if already present).
--
-- (ii) The NOTES subject is WRONG for this agent. tool-recreation-handler ends
--      save_sections -> update_status -> deploy_page and NEVER calls
--      create_tool_component, so ('tool', spec.function) keys a doc to a
--      function no component owns — a dangling doc. Recreation is site-scoped
--      page work, exactly like component-template-fixer. FIX: re-subject
--      append_note to ('pipeline','build') — set subject_type='pipeline', add
--      literal subject_key='build', DROP subject_key_field. note_site_id_field
--      is already 'site_record.site_id' (kept). Mirrors component-template-fixer
--      exactly. Retires the "recreation items must carry spec.function" backlog
--      item (for notes; a future component-creating recreation can revisit).
--      (docResolveSubject: literal subject_key wins; subject_key_field is only a
--      fallback — so dropping it is hygiene, the literal already takes effect.)
--
-- Standing rule: snapshot_agent opens the transaction.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler', '0NN_recreation_spec_and_note_subject.sql: pre-update');

UPDATE agent_definitions
SET
  -- (i) declare spec (idempotent)
  input_contract = CASE
      WHEN input_contract->'optional' ? 'spec' THEN input_contract
      ELSE jsonb_set(input_contract, '{optional}',
             (input_contract->'optional') || '["spec"]'::jsonb)
    END,
  -- (ii) re-subject the note: pipeline/build, drop the tool subject_key_field
  default_config = jsonb_set(
                     jsonb_set(
                       default_config #- '{workflow,steps,append_note,config,subject_key_field}',
                       '{workflow,steps,append_note,config,subject_type}', '"pipeline"'::jsonb),
                     '{workflow,steps,append_note,config,subject_key}', '"build"'::jsonb)
WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL;

-- Guard: assert the exact final shape (one live row, all conditions hold).
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'tool-recreation-handler' AND deleted_at IS NULL
      -- (i) spec now declared
      AND input_contract->'optional' ? 'spec'
      -- (ii) note re-subjected to pipeline/build, no dangling tool key
      AND default_config #>> '{workflow,steps,append_note,config,subject_type}' = 'pipeline'
      AND default_config #>> '{workflow,steps,append_note,config,subject_key}'  = 'build'
      AND NOT (default_config #> '{workflow,steps,append_note,config}' ? 'subject_key_field')
      AND default_config #>> '{workflow,steps,append_note,config,note_site_id_field}' = 'site_record.site_id'
      -- untouched invariants: still the append action, still error-contained,
      -- and the note tail still hangs off deploy_page
      AND default_config #>> '{workflow,steps,append_note,action}' = 'append_doc_note'
      AND default_config #>> '{workflow,steps,append_note,config,error_step}' = 'complete'
      AND default_config #>> '{workflow,steps,deploy_page,next_step}' = 'compose_note'
      -- and we introduced no step-level error_step (the ten were moved to config earlier)
      AND NOT EXISTS (SELECT 1 FROM jsonb_each(default_config #> '{workflow,steps}') t(k,v) WHERE v ? 'error_step');
    IF n <> 1 THEN RAISE EXCEPTION 'tool-recreation-handler fixes incomplete (found %)', n; END IF;
END $$;

COMMIT;

-- Verify after apply (all true, one row):
--   SELECT input_contract->'optional' ? 'spec'                                              AS spec_declared,
--          default_config #>> '{workflow,steps,append_note,config,subject_type}'            AS subj_type,
--          default_config #>> '{workflow,steps,append_note,config,subject_key}'             AS subj_key,
--          default_config #> '{workflow,steps,append_note,config}' ? 'subject_key_field'    AS has_key_field,
--          default_config #>> '{workflow,steps,append_note,config,note_site_id_field}'      AS site_field
--   FROM agent_definitions
--   WHERE type='tool-recreation-handler' AND deleted_at IS NULL;
--   -- expect: t | pipeline | build | f | site_record.site_id
--
-- Rollback: restore from the snapshot taken at the top, or manually:
--   remove 'spec' from input_contract.optional;
--   append_note.config: subject_type -> 'tool', re-add subject_key_field ->
--   'input_data.spec.function', delete subject_key.

-- ============================================================================
-- 356_retire_dead_config_keys_commit_from_and_hitl_output_format.sql
--
-- Two dead step-config keys removed from live agent_definitions, so that opting
-- their actions into unknown-config-key detection (ActionInputSpec.CheckConfig,
-- shipped in the same commit) does not warn about keys already adjudicated.
--
-- NOTE ON NUMBERING: briefed as 352; landed as 356. 352 was already taken when
-- the work started (352_pages_noindex_flag.sql), and while this file was being
-- written and dry-run three more numbers were claimed by other sessions —
-- 353_news_list_tag_ink_fix, 354_doc_notes_decision_subject_type,
-- 355_decision_guards_wiring_and_skip_terminal. Recorded because "next free
-- number" is not a fact you can carry across a session on this tree: it goes
-- stale in minutes. Duplicate numbers are in any case normal here (74 of them
-- in this directory today) and the runner orders by number then filename, so a
-- collision is untidy rather than dangerous — but a fresh number was free, so
-- this file took one rather than adding the 75th.
--
-- ---------------------------------------------------------------------------
-- KEY 1: `commit_from` on update_page_status — 6 live steps.
--
-- UpdatePageStatusAction (v3_site_actions.go:534-841) reads exactly five config
-- keys, and `commit_from` is not among them: status (:550), page_id_field
-- (:558), site_id_field (:584), page_name_field (:585), page_component_id_field
-- (:799). What establishes this is READING the handler end to end, not a grep:
-- the action takes `config := params.StepConfig.Config` at :543 and indexes it
-- directly, with no ResolveConfigSetting / GetStringField / GetIntField
-- indirection anywhere in the body and no ExtractActionInputs call, so there is
-- no second access pattern for a key to hide behind. It also has no
-- ActionInputSpec at all today, so no extractor strategy could resolve it
-- either. The value it names (a git commit sha) is never written to any column:
-- the action's UPDATE statements touch build_status, deployed_at,
-- built_from_plan_version and updated_at only.
--
-- Why it looked live for months, and this is the actual finding:
-- coordinator.go's prefixConfigStepReferences carried `"commit_from", // Used by
-- update_page_status` in its dataRefKeys slice. That comment was FALSE. Three
-- separate readers took the loop-prefixing list as a statement of what the
-- action consumes; it is only a list of keys whose VALUE gets rewritten when a
-- step is expanded inside a loop. The entry is deleted in the same commit — a
-- dead key is cheap, but a dead key with a comment vouching for it is what made
-- this cost repeated re-derivation.
--
-- KEY 2: `output_format` on process_approval_decision — 1 live step.
--
-- simple-content-writer-with-approval.process_approval, which spells the action
-- by its deprecated alias `process_data` (registry.go:1813). Its value is a MAP
-- of four {{.await_human_approval.*}} templates. ProcessApprovalDecisionAction
-- (hitl_actions.go) reads exactly one config key, `stop_on_reject`, and builds
-- its result map from the approval response in CollectedData. Nothing renders
-- the templates. Note the second, independent reason it could never have
-- worked: the only two readers of an `output_format` key anywhere in the tree
-- (ai_actions.go:1195, database_actions.go:26) both type-assert `.(string)`, so
-- even a mis-routed read would have missed a map. The config has been
-- describing an output shape that has never once been produced.
--
-- ---------------------------------------------------------------------------
-- SCOPE. Strictly the 6 update_page_status steps plus the 1 HITL step. The
-- 6 were established by a fleet-wide blindness check rather than by the census
-- that motivated the fix: EVERY agent_definitions row, any is_active /
-- is_snapshot / deleted_at state, across all four workflow columns
-- (default_config, task_workflow, orchestrator_workflow,
-- orchestration_workflow) whose text contains 'commit_from'. Result: 6 rows,
-- all active, all non-snapshot, all non-deleted, all in default_config only.
-- The step paths below are NESTED in three of the six (a loop sub_workflow), so
-- they were taken from a recursive walk, not from `->'workflow'->'steps'` —
-- that top-level-only shape returns a confident wrong number here.
--
-- NOT IN THIS FILE, deliberately: `notes_field` and `validation_issues_field`
-- on content-reviewer.mark_page_needs_attention. Found by the same enumeration
-- and dead the same way (neither string occurs anywhere in the tree outside an
-- unrelated action's `plan_notes_field`). They are NOT removed because, unlike
-- commit_from, they encode an author's intent this action has never had —
-- recording WHY a page was flagged — and pages has no column for it. Deleting
-- them would erase the only record of that intent; implementing them is a
-- behaviour change. They are left standing and will be REPORTED by the new
-- detector. That report is the detector working, not a regression. Same
-- treatment migration 350 gave create_work_item's `spec`.
--
-- SEEDS corrected in the same commit (a future reseed would otherwise replay
-- the key): 026_pageflow_builder.sql, 033_rerender_pages_action.sql,
-- 034_page_rerender_agent.sql, 039_page_rebuild_agent.sql,
-- 043_section_editor.sql, 045_site_work_orchestrator.sql,
-- 209_report_pipeline_agents.sql. That is SEVEN files, not the six live agents:
-- 034 seeds page-rerender (one of the six) and was missing from the briefed
-- list, while 033 seeds `rerender-pages`, a DIFFERENT and still-active agent
-- whose live row has already drifted away from its seed and carries no
-- commit_from today — cleaning it stops a reseed introducing the key into an
-- agent that currently does not have it. The HITL `output_format` has no
-- forward seed at all: it survives only in historical dump files
-- (000_agent_definitions_backup_070_refactor.sql, sql_for_agents_v2/000_backup*),
-- which are history and are deliberately left untouched.
--
-- Live immediately (DB config; no image roll involved). The Go half is inert
-- until the chassis rolls, and the two halves are independent in both
-- directions: this file makes six steps stop warning that the un-rolled binary
-- is not warning about anyway, and the un-rolled binary ignores the key exactly
-- as the rolled one does.
-- ============================================================================

BEGIN;

-- The target list, written ONCE. The guard, the removal and the verification
-- all read this table, so the three cannot drift apart — which is the failure
-- mode when a path is spelled out three times in one file.
CREATE TEMP TABLE _m356_targets (
    agent_type  text NOT NULL,
    step_label  text NOT NULL,
    cfg_path    text[] NOT NULL,   -- path to the step's config OBJECT
    dead_key    text NOT NULL,
    siblings    text[] NOT NULL    -- keys that MUST survive the delete
) ON COMMIT DROP;

INSERT INTO _m356_targets (agent_type, step_label, cfg_path, dead_key, siblings) VALUES
 ('pageflow-builder',      'build_pages_loop/update_page_status',
  '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config}',
  'commit_from',   '{status,page_id_field}'),
 ('page-rebuild',          'build_pages_loop/update_page_status',
  '{workflow,steps,build_pages_loop,config,sub_workflow,steps,update_page_status,config}',
  'commit_from',   '{status,page_id_field}'),
 ('page-rerender',         'update_status',
  '{workflow,steps,update_status,config}',
  'commit_from',   '{status,page_id_field}'),
 ('report-builder',        'update_status',
  '{workflow,steps,update_status,config}',
  'commit_from',   '{status,page_id_field}'),
 ('section-editor',        'update_page_status',
  '{workflow,steps,update_page_status,config}',
  'commit_from',   '{status,page_id_field,page_component_id_field}'),
 ('site-work-orchestrator','build_items_loop/update_page_status',
  '{workflow,steps,build_items_loop,config,sub_workflow,steps,update_page_status,config}',
  'commit_from',   '{status,page_id_field}'),
 ('simple-content-writer-with-approval', 'process_approval',
  '{workflow,steps,process_approval,config}',
  'output_format', '{input_fields}');

-- ---------------------------------------------------------------------------
-- GUARD. Idempotent: re-running after a successful apply RAISEs '... already
-- applied ...', which is what the runner's probe reads to report LIKELY ALREADY
-- APPLIED rather than halting. Any partial or unexpected state is DRIFT and
-- aborts, because a config estate that does not match the measurement this file
-- was written against is one nobody has re-measured.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  n_present  int;
  n_targets  int;
  n_missing_step int;
  detail     text;
BEGIN
  SELECT count(*) INTO n_targets FROM _m356_targets;

  SELECT count(*) INTO n_present
    FROM _m356_targets t
    JOIN agent_definitions ad
      ON ad.type = t.agent_type AND ad.is_active
     AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
   WHERE ad.default_config #> t.cfg_path ? t.dead_key;

  IF n_present = 0 THEN
    RAISE EXCEPTION '356: already applied (none of the % dead keys present)', n_targets;
  END IF;

  -- The step itself must exist on every target, or the path is wrong and a
  -- delete would be a silent no-op that this file would then call success.
  SELECT count(*) INTO n_missing_step
    FROM _m356_targets t
    LEFT JOIN agent_definitions ad
      ON ad.type = t.agent_type AND ad.is_active
     AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
   WHERE ad.id IS NULL
      OR jsonb_typeof(ad.default_config #> t.cfg_path) IS DISTINCT FROM 'object';

  IF n_missing_step > 0 THEN
    SELECT string_agg(t.agent_type || '.' || t.step_label, ', ') INTO detail
      FROM _m356_targets t
      LEFT JOIN agent_definitions ad
        ON ad.type = t.agent_type AND ad.is_active
       AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
     WHERE ad.id IS NULL
        OR jsonb_typeof(ad.default_config #> t.cfg_path) IS DISTINCT FROM 'object';
    RAISE EXCEPTION '356: DRIFT — % target step(s) do not resolve to a config object: %. Re-measure the paths (they are NESTED in three of the six) before applying.',
      n_missing_step, detail;
  END IF;

  IF n_present <> n_targets THEN
    SELECT string_agg(t.agent_type || '.' || t.step_label || ' (' || t.dead_key || ')', ', ') INTO detail
      FROM _m356_targets t
      JOIN agent_definitions ad
        ON ad.type = t.agent_type AND ad.is_active
       AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
     WHERE NOT (ad.default_config #> t.cfg_path ? t.dead_key);
    RAISE EXCEPTION '356: DRIFT — expected % dead keys, found %. Missing on: %. Partial state; re-measure before applying.',
      n_targets, n_present, detail;
  END IF;

  -- Blindness check, re-run at apply time rather than trusted from the file
  -- header: nothing OUTSIDE the target list may carry commit_from, or the
  -- scope stated above has gone stale between writing and applying.
  SELECT count(*) INTO n_present
    FROM agent_definitions ad
   WHERE (ad.default_config::text LIKE '%commit_from%'
          OR COALESCE(ad.task_workflow::text,'')          LIKE '%commit_from%'
          OR COALESCE(ad.orchestrator_workflow::text,'')  LIKE '%commit_from%'
          OR COALESCE(ad.orchestration_workflow::text,'') LIKE '%commit_from%')
     AND ad.type NOT IN (SELECT agent_type FROM _m356_targets WHERE dead_key='commit_from');
  IF n_present > 0 THEN
    RAISE EXCEPTION '356: DRIFT — % agent_definitions row(s) outside the target list now carry commit_from. Scope has changed since measurement; re-census before applying.', n_present;
  END IF;
END $$;

-- House practice: snapshot every agent this file touches, before touching it.
-- Two-arg snapshot_agent writes agent_definitions_backup rows, NOT is_snapshot
-- rows, so it cannot perturb the is_snapshot=false predicates above.
SELECT snapshot_agent('pageflow-builder',                    '356_retire_dead_config_keys: pre-update');
SELECT snapshot_agent('page-rebuild',                        '356_retire_dead_config_keys: pre-update');
SELECT snapshot_agent('page-rerender',                       '356_retire_dead_config_keys: pre-update');
SELECT snapshot_agent('report-builder',                      '356_retire_dead_config_keys: pre-update');
SELECT snapshot_agent('section-editor',                      '356_retire_dead_config_keys: pre-update');
SELECT snapshot_agent('site-work-orchestrator',              '356_retire_dead_config_keys: pre-update');
SELECT snapshot_agent('simple-content-writer-with-approval', '356_retire_dead_config_keys: pre-update');

-- ---------------------------------------------------------------------------
-- REMOVAL. #- takes the full path to the key; the paths are nested for three of
-- the seven, which is why they are data rather than literals in the statement.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions ad
   SET default_config = ad.default_config #- (t.cfg_path || t.dead_key),
       updated_at = now()
  FROM _m356_targets t
 WHERE ad.type = t.agent_type AND ad.is_active
   AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL
   AND ad.default_config #> t.cfg_path ? t.dead_key;

-- ---------------------------------------------------------------------------
-- VERIFY. DO/RAISE, never a block of bare SELECTs: ON_ERROR_STOP ignores a
-- non-empty result, so a SELECT-based "verification" cannot stop the COMMIT and
-- would let a failed migration record itself as applied.
--
-- Asserts BOTH directions. The removal alone is not enough — a whole-config
-- clobber also removes the key, and would pass a removal-only check while
-- destroying the step. So every sibling key must still be present, and the step
-- config must still be an object.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  r        record;
  cfg      jsonb;
  bad      text;
BEGIN
  FOR r IN SELECT * FROM _m356_targets LOOP
    SELECT ad.default_config #> r.cfg_path INTO cfg
      FROM agent_definitions ad
     WHERE ad.type = r.agent_type AND ad.is_active
       AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL;

    IF cfg IS NULL OR jsonb_typeof(cfg) <> 'object' THEN
      RAISE EXCEPTION '356 VERIFY: %.% — step config is no longer an object (got %). The delete took the step with it.',
        r.agent_type, r.step_label, COALESCE(jsonb_typeof(cfg),'NULL');
    END IF;

    IF cfg ? r.dead_key THEN
      RAISE EXCEPTION '356 VERIFY: %.% still carries %',
        r.agent_type, r.step_label, r.dead_key;
    END IF;

    SELECT string_agg(s, ', ') INTO bad
      FROM unnest(r.siblings) s
     WHERE NOT (cfg ? s);
    IF bad IS NOT NULL THEN
      RAISE EXCEPTION '356 VERIFY: %.% lost sibling key(s): %. Config now: %',
        r.agent_type, r.step_label, bad, cfg;
    END IF;
  END LOOP;

  -- Fleet-wide: no commit_from left anywhere, in any workflow column, in any row
  -- state. This is the assertion the per-target loop cannot make.
  SELECT count(*)::text INTO bad
    FROM agent_definitions ad
   WHERE ad.default_config::text LIKE '%commit_from%'
      OR COALESCE(ad.task_workflow::text,'')          LIKE '%commit_from%'
      OR COALESCE(ad.orchestrator_workflow::text,'')  LIKE '%commit_from%'
      OR COALESCE(ad.orchestration_workflow::text,'') LIKE '%commit_from%';
  IF bad <> '0' THEN
    RAISE EXCEPTION '356 VERIFY: % agent_definitions row(s) still carry commit_from somewhere', bad;
  END IF;
END $$;

COMMIT;

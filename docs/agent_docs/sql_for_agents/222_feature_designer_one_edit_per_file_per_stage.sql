-- 222_feature_designer_one_edit_per_file_per_stage.sql
--
-- bugs_open/099 — the feature designer's plans die on a validator rule its prompt
-- never states. Fix candidate 1 (the cheapest of three; the durable one is
-- candidate 2, routing the validation failure into `repropose`, which is NOT done
-- here).
--
-- WHY
-- ---
-- diagnose_persist_fix_plan refuses a staged plan in which one file appears twice
-- in the same stage:
--
--   platform/orchestration/actions/diagnose_persist_fix_plan_action.go:526-532
--     if seenPath[f] {
--         problems = append(problems, fmt.Sprintf(
--             "%s: %s appears in more than one edit of this stage", tag, f))
--     }
--
-- The design step's prompt caps QUANTITIES (rule 4: "at most 6 stages, 8 edits per
-- stage, 24 edits total") and says nothing about uniqueness of `file` within a
-- stage. Measured live 2026-07-26: the string "more than one edit" and the variant
-- "one edit per file" both appear NOWHERE in the feature-designer row. So a
-- designer that splits one file's work into two readable edits — "add the helper",
-- "add the entry point" — emits a plan that is coherent, inside every stated cap,
-- and unpersistable. It cannot know.
--
-- The loss is not small: the run completes, spends the full designer cost, writes
-- NO fix_plan artifact, and lands on complete_refused with
-- orchestration_states.error NULL (the reason lives only in
-- collected_data->>'__step_error'). A good design is destroyed silently.
-- Live instance and the destroyed plan: bugs_open/099.
--
-- SCOPE — design ONLY, deliberately
-- ---------------------------------
-- `reframe` and `repropose` were checked and carry NO copy of the caps rule
-- (their prompt_templates do not contain "CAPS:"), so there is nothing to keep in
-- lockstep and no second edit to make. If that ever changes, this rule has to move
-- with it — two copies of one constraint is the drift bugs_open/099 is about.
--
-- IDEMPOTENT + NON-CLOBBERING
-- ---------------------------
-- Guarded: refuses if the text is already present. Uses replace() on the exact
-- rule-4 sentence rather than rewriting the template wholesale, so a concurrent
-- edit elsewhere in the prompt survives. If rule 4's wording has changed under us
-- the replace is a no-op and the verification block below RAISEs — loudly, rather
-- than reporting a success it did not achieve.
--
-- ROLLBACK
-- --------
--   Restore from the snapshot this file takes:
--     SELECT snapshot_agent('feature-designer', 'manual rollback of 222');
--     -- then copy default_config back from the newest is_snapshot row created by
--     -- '222_feature_designer_one_edit_per_file_per_stage.sql: pre-update'.

BEGIN;

-- ---------------------------------------------------------------------------
-- Guard: refuse a second application.
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'feature-designer'
           AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false
           AND (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template')
               ILIKE '%ONE EDIT PER FILE PER STAGE%'
    ) THEN
        RAISE EXCEPTION '222: already applied — the design prompt already states the one-edit-per-file rule';
    END IF;
END $$;

SELECT snapshot_agent('feature-designer',
    '222_feature_designer_one_edit_per_file_per_stage.sql: pre-update');

-- ---------------------------------------------------------------------------
-- Extend rule 4 in place.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,design,config,prompt_template}',
         to_jsonb(
           replace(
             default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template',
             '4. CAPS: at most 6 stages, 8 edits per stage, 24 edits total.',
             '4. CAPS: at most 6 stages, 8 edits per stage, 24 edits total, and AT MOST '
             || 'ONE EDIT PER FILE PER STAGE. A path repeated inside one stage fails validation '
             || 'and the ENTIRE plan is discarded — combine every change you want to make to a '
             || 'file in that stage into a SINGLE edit whose sketch describes them all, or move '
             || 'the second change to a later stage.'
           )
         )
       ),
       updated_at = now()
 WHERE type = 'feature-designer'
   AND deleted_at IS NULL
   AND COALESCE(is_snapshot, false) = false;

-- ---------------------------------------------------------------------------
-- Verify the replace actually bit. A silent no-op here (rule 4 reworded by
-- someone else) would leave the bug in place while this file reported success.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    ok boolean;
BEGIN
    SELECT (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template')
           ILIKE '%ONE EDIT PER FILE PER STAGE%'
      INTO ok
      FROM agent_definitions
     WHERE type = 'feature-designer'
       AND deleted_at IS NULL AND COALESCE(is_snapshot, false) = false;

    IF NOT COALESCE(ok, false) THEN
        RAISE EXCEPTION '222: replace() did not bite — rule 4 wording has changed; re-read the live prompt_template and update this migration';
    END IF;
    RAISE NOTICE '222: feature-designer design prompt now states the one-edit-per-file-per-stage rule';
END $$;

COMMIT;

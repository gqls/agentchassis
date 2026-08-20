-- 500 — fix 499: a jsonb column arrives as a STRING, so the template had nothing
--       to range over and `write_descriptions` FAILED (bugs_open/320)
--
-- WHAT I BROKE, AND HOW IT PRESENTED. 499 loaded the site's banned phrases with
-- `jsonb_agg(...) AS rules` and the prompt did `{{range .voice_rules.rules}}`. On the
-- first real run the step did not silently skip — it **FAILED**:
--
--     current_step = write_descriptions, status = FAILED
--
-- Diagnosed at the run's own collected_data rather than guessed:
--
--     voice_rules      -> object      (correct, output_format was already "object")
--     voice_rules.rules-> **string**  "[{\"reason\": \"banned_language: ...\"}, ..."
--
-- ⚠ **`query_database` STRINGIFIES a jsonb column.** `database_actions.go` converts a
-- `[]byte` scan target to a Go `string`, and pgx hands back jsonb as `[]byte`. So any
-- column you build with `jsonb_agg`/`jsonb_build_object` reaches the template as a
-- JSON *string*, not a structure — and `{{range}}` over a string is not iteration.
-- The value LOOKS right in every log line and in psql, which is what makes it a trap.
--
-- THE FIX IS THE SHAPE, NOT THE TEMPLATE. Stop aggregating. Return ONE ROW PER RULE
-- with plain text columns, and iterate `output_format: "object"`'s own `rows` array,
-- which the action builds as real structures:
--
--     result := map[string]interface{}{"rows": results, "count": len(results), ...}
--
-- So `{{range .voice_rules.rows}}` iterates row maps and `{{.pattern}}` / `{{.reason}}`
-- resolve. `count` also becomes available, which is what the `{{if}}` should test —
-- a site with no gate returns zero rows and the block is skipped.
--
-- WHY NOT ROLL 499 BACK. Rolling back restores a workflow that works but re-arms the
-- permanent hourly retry 499 exists to stop. The defect here is one wrong column
-- shape, it is understood, and it is verified below against the site that has 14
-- rules. If this file's verification had not passed, the correct move WOULD have been
-- 499's rollback — that is why it exists.
--
-- ROLLBACK: 500_backfiller_voice_rules_row_shape_fix_ROLLBACK.sql
--   (reverts to 499's broken shape; you almost certainly want 499's rollback instead)

BEGIN;

SELECT snapshot_agent('meta-description-backfiller', '500_row_shape_fix: pre-update');

DO $$
DECLARE q text; p text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_voice_rules,config,query}',
         default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}'
    INTO q, p
    FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF q IS NULL THEN
    RAISE EXCEPTION '500: load_voice_rules is missing — 499 is not applied';
  END IF;
  IF position('jsonb_agg' in q) = 0 THEN
    RAISE EXCEPTION '500: the query no longer aggregates — already fixed or changed under me';
  END IF;
  IF position('{{range .voice_rules.rules}}' in p) = 0 THEN
    RAISE EXCEPTION '500: the prompt does not carry 499''s broken range form';
  END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,load_voice_rules,config,query}',
           to_jsonb(
             'SELECT ph->>''pattern'' AS pattern, ph->>''reason'' AS reason ' ||
             'FROM site_specs ss ' ||
             'CROSS JOIN LATERAL jsonb_array_elements(COALESCE(ss.data#>''{voice_gate,banned_phrases}'', ''[]''::jsonb)) ph ' ||
             'WHERE ss.site_id = $1 AND ss.aspect = ''voice'' AND ss.is_current = true ' ||
             '  AND COALESCE((ss.data#>>''{voice_gate,enabled}'')::boolean, false) = true'
           )
         ),
         '{workflow,steps,write_descriptions,config,prompt_template}',
         to_jsonb(
           replace(
             replace(
               default_config#>>'{workflow,steps,write_descriptions,config,prompt_template}',
               '{{if .voice_rules.rules}}', '{{if .voice_rules.count}}'
             ),
             '{{range .voice_rules.rules}}', '{{range .voice_rules.rows}}'
           )
         )
       ),
       updated_at = now()
 WHERE type='meta-description-backfiller' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE cfg jsonb; p text; n int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='meta-description-backfiller' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  p := cfg#>>'{workflow,steps,write_descriptions,config,prompt_template}';
  IF position('{{range .voice_rules.rows}}' in p) = 0 THEN
    RAISE EXCEPTION '500 VERIFY: the prompt does not iterate .rows';
  END IF;
  IF position('.voice_rules.rules' in p) > 0 THEN
    RAISE EXCEPTION '500 VERIFY: the broken .rules reference survives';
  END IF;
  IF position('jsonb_agg' in cfg#>>'{workflow,steps,load_voice_rules,config,query}') > 0 THEN
    RAISE EXCEPTION '500 VERIFY: the query still aggregates';
  END IF;

  -- Run the NEW query and assert it returns ROWS, not one aggregate.
  EXECUTE 'SELECT count(*) FROM ('
       || 'SELECT ph->>''pattern'' AS pattern FROM site_specs ss '
       || 'CROSS JOIN LATERAL jsonb_array_elements(COALESCE(ss.data#>''{voice_gate,banned_phrases}'', ''[]''::jsonb)) ph '
       || 'WHERE ss.site_id = (SELECT id FROM sites WHERE domain=''leopardessconsulting.co.uk'') '
       || '  AND ss.aspect=''voice'' AND ss.is_current) z' INTO n;
  IF n < 10 THEN
    RAISE EXCEPTION '500 VERIFY: expected many rows for the 14-rule check site, got %', n;
  END IF;
  RAISE NOTICE '500 OK: one row per rule (% on the check site), prompt iterates .rows', n;
END $$;

COMMIT;

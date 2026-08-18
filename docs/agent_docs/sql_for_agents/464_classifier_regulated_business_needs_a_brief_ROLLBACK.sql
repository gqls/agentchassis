-- ROLLBACK for 464_classifier_regulated_business_needs_a_brief.sql
--
-- Removes the two static insertions from domain-research-classifier's classify_and_extract
-- prompt. Anchored on the inserted text itself, so it cannot damage anything another
-- migration has added since; it ABORTS rather than guess if the text is not found exactly once.

BEGIN;

SELECT snapshot_agent('domain-research-classifier',
                      '464_ROLLBACK: pre-revert');

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  extra text := ' Subject to the regulated-business rule above — a name tells you the SUBJECT, never that the site may carry on a regulated activity.';
  rule_start int; rule_end int; rule text;
BEGIN
  SELECT default_config #>> '{workflow,steps,classify_and_extract,config,prompt_template}'
    INTO tpl FROM agent_definitions
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  rule_start := position(E'\n\n## Regulated business models are NOT available' in tpl);
  IF rule_start = 0 THEN RAISE EXCEPTION '464 ROLLBACK: rule block not found'; END IF;
  -- the block ends at the paragraph that closes it
  rule_end := position('do not treat a same-named company found elsewhere in the world as evidence that this site is that company.' in tpl);
  IF rule_end = 0 THEN RAISE EXCEPTION '464 ROLLBACK: rule block end not found'; END IF;
  rule_end := rule_end + length('do not treat a same-named company found elsewhere in the world as evidence that this site is that company.');

  rule := substring(tpl from rule_start for rule_end - rule_start);
  newtpl := replace(tpl, rule, '');
  newtpl := replace(newtpl, extra, '');

  IF position('Regulated business models are NOT available' in newtpl) > 0 THEN
    RAISE EXCEPTION '464 ROLLBACK: rule still present after removal';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,classify_and_extract,config,prompt_template}',
           to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='domain-research-classifier' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '464 ROLLBACK: updated % rows', n; END IF;
END
$do$;

COMMIT;

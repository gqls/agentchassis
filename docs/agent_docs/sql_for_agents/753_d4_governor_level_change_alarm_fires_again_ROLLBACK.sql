-- 753_d4_governor_level_change_alarm_fires_again_ROLLBACK.sql — restores 673's stored text
-- VERBATIM. ⚠ 673's text has a DEAD ALARM (bugs_open/459): after this rollback the governor
-- sheds silently again. Roll back only if 753's statement itself misbehaves (e.g. the task stops
-- recomputing); the alarm defect is the known cost, not a surprise.

BEGIN;

DO $$
DECLARE q text;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='spend-governor-state';
  IF q IS NULL THEN RAISE EXCEPTION '753 ROLLBACK REFUSED: task not found.'; END IF;
  IF position('-- 753: alarm reads old_lvl from upd RETURNING' in q) = 0 THEN
    RAISE EXCEPTION '753 ROLLBACK REFUSED: 753 marker absent — not applied (or already rolled back).';
  END IF;
END $$;

UPDATE scheduled_tasks SET pre_query = $PRE$
WITH cfg AS (
  SELECT c.* FROM governor_config c,
       (SELECT pg_advisory_xact_lock(hashtext('spend-governor-state'))) lock_taken
  WHERE c.id = 1),
spend AS (SELECT * FROM governor_spend_mtd),
new AS (
  SELECT CASE
           WHEN cfg.monthly_budget_usd IS NULL THEN 0
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l3_pct/100 THEN 3
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l2_pct/100 THEN 2
           WHEN spend.mtd_usd >= cfg.monthly_budget_usd * cfg.l1_pct/100 THEN 1
           ELSE 0
         END AS lvl,
         spend.mtd_usd, spend.unpriced_io_tokens, spend.month
  FROM cfg, spend),
old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE),
noted AS (
  INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
  SELECT 'pipeline', 'spend-governor',
         format('spend-governor: shed level %s -> %s (month-to-date $%s of budget $%s; unpriced io tokens %s)',
                old.shed_level, new.lvl, new.mtd_usd,
                COALESCE((SELECT monthly_budget_usd::text FROM cfg), 'UNSET'), new.unpriced_io_tokens),
         '["spend-governor"]'::jsonb, 'scheduled_tasks:spend-governor-state'
  FROM old, new WHERE old.shed_level <> new.lvl
  RETURNING 1),
upd AS (
  UPDATE governor_state s
     SET shed_level = new.lvl, month = new.month, mtd_usd = new.mtd_usd,
         unpriced_io_tokens = new.unpriced_io_tokens, computed_at = now()
    FROM new WHERE s.id = 1
  RETURNING s.shed_level)
SELECT (SELECT shed_level FROM upd)  AS shed_level,
       (SELECT count(*) FROM noted)  AS level_changed
$PRE$
WHERE name = 'spend-governor-state';

DO $$
DECLARE q text; lvl int; changed int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='spend-governor-state';
  IF position('old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE),' in q) = 0 THEN
    RAISE EXCEPTION '753 ROLLBACK VERIFY: 673 text not restored';
  END IF;
  EXECUTE q INTO lvl, changed;   -- the restored text must still run and recompute
  RAISE NOTICE '753 ROLLBACK OK: 673 text restored (level now %). REMINDER: this text''s level-change alarm is DEAD (bugs_open/459).', lvl;
END $$;

COMMIT;

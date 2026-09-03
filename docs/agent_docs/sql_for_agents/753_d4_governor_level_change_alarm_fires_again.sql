-- 753_d4_governor_level_change_alarm_fires_again.sql — bugs_open/459: the spend governor's
-- level-change ALARM has never fired since 673 hardened the state task.
--
-- THE DEFECT (A/B-proven 2026-09-03, NOTES + bug 459). The task's single statement held
--   old   AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE)
--   noted AS (INSERT INTO doc_notes … FROM old, new WHERE old.shed_level <> new.lvl …)
--   upd   AS (UPDATE governor_state … WHERE s.id = 1 …)
-- — two CTEs of ONE statement, one row-locking the row the other updates. `noted` selected
-- nothing however far the level moved: two real level changes on 09-03 wrote ZERO notes; delete
-- the one token `FOR UPDATE` and the note lands. 672 installed the alarm and PROVED it (drove
-- 0→3→0, asserted exactly 2 notes); 673 added FOR UPDATE for a real reason (a fire that BLOCKS
-- on the advisory lock keeps its pre-block snapshot, so a plain `old` read could be stale) and
-- its verify checked only that the token was present and the text still ran.
--
-- THE FIX keeps 673's property AND the alarm: the old value is read by the UPDATE ITSELF, from a
-- FOR UPDATE sub-select in its own FROM list, and returned alongside the new value —
--   upd AS (UPDATE governor_state s SET … FROM new, (SELECT shed_level AS prev FROM governor_state
--           WHERE id = 1 FOR UPDATE) o WHERE s.id = 1 RETURNING o.prev AS old_lvl, s.shed_level AS new_lvl …)
--   noted AS (INSERT … FROM upd WHERE upd.old_lvl <> upd.new_lvl …)
-- One data-modifying statement locks, re-reads (EvalPlanQual) and writes the row; `noted` is a
-- dependent CTE reading upd's RETURNING, not a second reader of the row. No CTE races another
-- over the same tuple. The advisory lock stays (belt); the FOR UPDATE inside the UPDATE's own
-- FROM list is the braces, in the one place they can be worn.
--
-- OWNER RULING 2026-09-03 ("first, and loudly"): the note now also SAYS whether shedding
-- INCREASED or EASED, in words a reader will act on, and carries a second category
-- 'level-change' so a session-start banner and the README can find it in one predicate.
--
-- GUARD: refuses unless the stored text carries 673's exact `old AS (… FOR UPDATE),` line once
-- and does not yet carry this migration's marker. VERIFY drives a real level change through the
-- NEW text and asserts the note lands, drives a no-change tick and asserts no note, then deletes
-- exactly the synthetic notes it wrote and restores governor_state EVERY COLUMN as found
-- (752's verify restored shed_level alone and left mtd_usd/computed_at NULL on the real apply —
-- self-healed at the next tick, logged in WRONG_CALLS; not repeated here).
-- Rollback: 753_..._ROLLBACK.sql restores 673's text verbatim (and its dead alarm — say so).

BEGIN;

DO $$
DECLARE q text; n int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='spend-governor-state';
  IF q IS NULL THEN RAISE EXCEPTION '753 REFUSED: task spend-governor-state not found.'; END IF;
  IF position('-- 753: alarm reads old_lvl from upd RETURNING' in q) > 0 THEN
    RAISE EXCEPTION '753 REFUSED: already applied (replay).';
  END IF;
  n := (length(q) - length(replace(q, 'old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE),', '')))
       / length('old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE),');
  IF n <> 1 THEN
    RAISE EXCEPTION '753 REFUSED: expected 673''s FOR UPDATE old-read exactly once in the stored text, found % — drifted, investigate.', n;
  END IF;
  IF position('pg_advisory_xact_lock(hashtext(''spend-governor-state''))' in q) = 0 THEN
    RAISE EXCEPTION '753 REFUSED: 672''s advisory lock is absent from the stored text — not the expected lineage.';
  END IF;
END $$;

UPDATE scheduled_tasks SET pre_query = $PRE$
-- 753: alarm reads old_lvl from upd RETURNING (bugs_open/459); 672 lock kept; 673 FOR UPDATE moved into the UPDATE's own FROM list
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
upd AS (
  UPDATE governor_state s
     SET shed_level = new.lvl, month = new.month, mtd_usd = new.mtd_usd,
         unpriced_io_tokens = new.unpriced_io_tokens, computed_at = now()
    FROM new, (SELECT shed_level AS prev FROM governor_state WHERE id = 1 FOR UPDATE) o
   WHERE s.id = 1
  RETURNING o.prev AS old_lvl, s.shed_level AS new_lvl, s.mtd_usd, s.unpriced_io_tokens),
noted AS (
  INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
  SELECT 'pipeline', 'spend-governor',
         format('spend-governor: shed level %s -> %s (month-to-date $%s of budget $%s; unpriced io tokens %s)%s',
                upd.old_lvl, upd.new_lvl, upd.mtd_usd,
                COALESCE((SELECT monthly_budget_usd::text FROM cfg), 'UNSET'), upd.unpriced_io_tokens,
                CASE WHEN upd.new_lvl > upd.old_lvl THEN
                       ' — SHEDDING INCREASED: work classes at or below this level are now WITHHELD (see governor_work_class_map and governor_agent_class_map; governor_withheld_now lists the items). Owner ruling 2026-09-03: when council reviews are withheld, stop non-essential platform work until the level drops.'
                     ELSE ' — shedding EASED: classes above the new level resume on the next dispatch.' END),
         '["spend-governor","level-change"]'::jsonb, 'scheduled_tasks:spend-governor-state'
  FROM upd WHERE upd.old_lvl <> upd.new_lvl
  RETURNING 1)
SELECT (SELECT new_lvl FROM upd)     AS shed_level,
       (SELECT count(*) FROM noted)  AS level_changed
$PRE$
WHERE name = 'spend-governor-state';

-- ---------------------------------------------------------------- verify (DO/RAISE)
DO $$
DECLARE q text; changed int; lvl int; lvl1 int; before_notes int; after_notes int; id1 uuid; id2 uuid; body1 text;
        s_level int; s_month text; s_mtd numeric; s_tok bigint; s_at timestamptz;
        c_enabled boolean; c_budget numeric;
BEGIN
  -- LOCK ORDER = the live task's (advisory first, then the row). The first dry-run of this file
  -- took the row lock via UPDATE and then waited on the advisory lock inside the EXECUTEd text
  -- while a real tick held the advisory lock and waited on the row: 'deadlock detected'.
  PERFORM pg_advisory_xact_lock(hashtext('spend-governor-state'));
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='spend-governor-state';
  IF position('-- 753: alarm reads old_lvl from upd RETURNING' in q) = 0 THEN
    RAISE EXCEPTION '753 VERIFY: marker absent after UPDATE — the write did not land';
  END IF;
  IF position('old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE),' in q) > 0 THEN
    RAISE EXCEPTION '753 VERIFY: 673''s racing old-read CTE is still in the text';
  END IF;

  -- save EVERY column of the live state and config (752''s lesson: restore the whole row, not one field)
  SELECT shed_level, month, mtd_usd, unpriced_io_tokens, computed_at INTO s_level, s_month, s_mtd, s_tok, s_at FROM governor_state WHERE id=1;
  SELECT enabled, monthly_budget_usd INTO c_enabled, c_budget FROM governor_config WHERE id=1;
  SELECT count(*) INTO before_notes FROM doc_notes WHERE subject_key='spend-governor';

  -- (1) force old <> new and run the LIVE text: exactly one note must land, naming both levels
  UPDATE governor_config SET enabled = true, monthly_budget_usd = 2000 WHERE id=1;
  UPDATE governor_state SET shed_level = 3 WHERE id=1;          -- computed level for $2000 is < 3 today
  EXECUTE q INTO lvl1, changed;
  IF changed <> 1 THEN RAISE EXCEPTION '753 VERIFY: level 3 -> % should have written 1 note, level_changed=%', lvl1, changed; END IF;
  SELECT count(*) INTO after_notes FROM doc_notes WHERE subject_key='spend-governor';
  IF after_notes - before_notes <> 1 THEN RAISE EXCEPTION '753 VERIFY: expected exactly 1 new spend-governor note, got %', after_notes - before_notes; END IF;
  -- read the note by BODY, not by created_at: rows written in one transaction share created_at = transaction start
  SELECT id, body INTO id1, body1 FROM doc_notes
   WHERE subject_key='spend-governor' AND body LIKE 'spend-governor: shed level 3 -> ' || lvl1::text || ' %' AND categories ? 'level-change'
   ORDER BY created_at DESC LIMIT 1;
  IF id1 IS NULL OR position('shedding EASED' in body1) = 0 THEN
    RAISE EXCEPTION '753 VERIFY: no level-change note naming 3 -> % with EASED wording (found: %)', lvl1, COALESCE(left(body1,140),'<none>');
  END IF;

  -- (2) a no-change tick must write NOTHING
  EXECUTE q INTO lvl, changed;
  IF changed <> 0 THEN RAISE EXCEPTION '753 VERIFY: a no-change tick wrote a note (level_changed=%)', changed; END IF;
  SELECT count(*) INTO after_notes FROM doc_notes WHERE subject_key='spend-governor';
  IF after_notes - before_notes <> 1 THEN RAISE EXCEPTION '753 VERIFY: no-change tick changed the note count to +%', after_notes - before_notes; END IF;

  -- (3) the INCREASE wording: drive 0 -> 3 with a tiny budget
  UPDATE governor_config SET monthly_budget_usd = 0.01 WHERE id=1;
  EXECUTE q INTO lvl, changed;
  IF lvl <> 3 OR changed <> 1 THEN RAISE EXCEPTION '753 VERIFY: budget 0.01 should give level 3 and one note; got level % changed %', lvl, changed; END IF;
  SELECT id, body INTO id2, body1 FROM doc_notes
   WHERE subject_key='spend-governor' AND body LIKE 'spend-governor: shed level ' || lvl1::text || ' -> 3 %' AND categories ? 'level-change' AND id <> id1
   ORDER BY created_at DESC LIMIT 1;
  IF id2 IS NULL OR position('SHEDDING INCREASED' in body1) = 0 OR position('stop non-essential platform work' in body1) = 0 THEN
    RAISE EXCEPTION '753 VERIFY: no % -> 3 note with the loud INCREASED wording (found: %)', lvl1, COALESCE(left(body1,160),'<none>');
  END IF;

  -- clean up EXACTLY the two synthetic notes by id, then restore EVERY column as found
  DELETE FROM doc_notes WHERE id IN (id1, id2);
  SELECT count(*) INTO after_notes FROM doc_notes WHERE subject_key='spend-governor';
  IF after_notes <> before_notes THEN RAISE EXCEPTION '753 VERIFY: cleanup left % notes, expected %', after_notes, before_notes; END IF;
  UPDATE governor_config SET enabled = c_enabled, monthly_budget_usd = c_budget WHERE id=1;
  UPDATE governor_state SET shed_level = s_level, month = s_month, mtd_usd = s_mtd, unpriced_io_tokens = s_tok, computed_at = s_at WHERE id=1;

  RAISE NOTICE '753 OK: alarm fires on a level change (3 -> % EASED, then % -> 3 INCREASED with the loud wording), stays silent on a no-change tick, category level-change present; both synthetic notes deleted by id; governor_state and config restored column-for-column.', lvl1, lvl1;
END $$;

COMMIT;

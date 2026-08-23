-- ROLLBACK for 568 — removes the "already committed is normal" amendment from both
-- rosters' editquality seat.
--
-- Restores the prompt from the backup table 568 took, rather than doing string surgery
-- in reverse. A reverse `replace()` would have to match the inserted paragraph byte for
-- byte, and if it missed (a smart quote normalised, a later edit landing in between) it
-- would silently leave a partial prompt and report success. Restoring the whole
-- prompt_template from the backup cannot leave a half-state.
--
-- ⚠ SCOPE: this restores ONLY the review_editquality prompt_template, not the whole
-- default_config. If another lane has edited a DIFFERENT step on these rosters since 568
-- applied, that work is preserved — which is the behaviour you want on a tree this many
-- sessions share.

BEGIN;

DO $r$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_568_editquality_20260823;
  IF n < 2 THEN
    RAISE EXCEPTION '568 ROLLBACK ABORT: backup holds % row(s), expected at least 2 (council-gate + fix-proposer).', n;
  END IF;
END
$r$;

UPDATE agent_definitions a
SET default_config = jsonb_set(
      a.default_config,
      '{workflow,steps,review_editquality,config,prompt_template}',
      b.default_config->'workflow'->'steps'->'review_editquality'->'config'->'prompt_template',
      false)
FROM bak_568_editquality_20260823 b
WHERE a.id = b.id
  AND a.default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template'
      LIKE '%ALREADY COMMITTED IS NORMAL%';

DO $v$
DECLARE
  n_left int;
  n_marked int;
  n_prefix_distinct int;
BEGIN
  SELECT count(*) INTO n_left FROM agent_definitions
  WHERE type IN ('council-gate','fix-proposer') AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND default_config->'workflow'->'steps'->'review_editquality'->'config'->>'prompt_template'
        LIKE '%ALREADY COMMITTED IS NORMAL%';
  IF n_left <> 0 THEN
    RAISE EXCEPTION '568 ROLLBACK VERIFY FAILED: amendment still present in % roster(s).', n_left;
  END IF;

  -- The rollback must not fragment the shared prefix either.
  SELECT count(*), count(DISTINCT split_part(p,'<!--CACHE_BREAKPOINT-->',1))
    INTO n_marked, n_prefix_distinct
  FROM (
    SELECT v->'config'->>'prompt_template' AS p
    FROM agent_definitions, LATERAL jsonb_each(default_config->'workflow'->'steps') AS e(k,v)
    WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false
      AND deleted_at IS NULL AND v->'config'->>'prompt_template' LIKE '%<!--CACHE_BREAKPOINT-->%'
  ) s;
  IF n_marked <> 17 OR n_prefix_distinct <> 1 THEN
    RAISE EXCEPTION '568 ROLLBACK VERIFY FAILED: prefix-cache health is % seats / % distinct prefixes, expected 17 / 1.', n_marked, n_prefix_distinct;
  END IF;

  RAISE NOTICE '568 ROLLBACK OK: both editquality prompts restored; cache prefix intact.';
END
$v$;

COMMIT;

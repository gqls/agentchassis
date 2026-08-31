-- 673_d4_governor_snapshot_close.sql — closes the r2 advisory (corr 80df0963, APPROVED) with
-- the SHARPER form of its own concern. The advisory worried the xact-scoped lock protects only
-- one statement's span — which is in fact the whole race window. The REAL residual hole is the
-- SNAPSHOT: under READ COMMITTED a statement's snapshot is taken at statement start, so a fire
-- that BLOCKS on the advisory lock keeps its pre-block snapshot; on unblock its `old` CTE reads
-- the STALE shed_level and can double-write (or mis-skip) a level-change note. The close:
-- `FOR UPDATE` on the `old` read — EvalPlanQual re-reads the committed row version when the
-- row lock releases, so `old` is always current. The advisory lock stays (belt), FOR UPDATE
-- is the braces. Also folded in: debug_historian's rowcount assertion at the mutation site.
-- Guard: md5 tri-arm, same pattern as 672.

BEGIN;

DO $$
DECLARE m text; n int;
BEGIN
  SELECT md5(pre_query) INTO m FROM scheduled_tasks WHERE name='spend-governor-state';
  IF m IS NULL THEN RAISE EXCEPTION '673 REFUSED: task not found — 671/672 not applied.'; END IF;
  IF m <> '838f8cd1cad9705f9e6651cf04dafab6' THEN
    IF position('FOR UPDATE' in (SELECT pre_query FROM scheduled_tasks WHERE name='spend-governor-state')) > 0 THEN
      RAISE EXCEPTION '673 REFUSED: already applied (replay) — FOR UPDATE present.';
    END IF;
    RAISE EXCEPTION '673 REFUSED: pre_query md5 % is neither 672''s text nor the hardened one — drifted, investigate.', m;
  END IF;

  UPDATE scheduled_tasks SET pre_query = replace(pre_query,
    'old AS (SELECT shed_level FROM governor_state WHERE id = 1),',
    'old AS (SELECT shed_level FROM governor_state WHERE id = 1 FOR UPDATE),')
  WHERE name = 'spend-governor-state';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '673: UPDATE touched % rows, expected exactly 1', n; END IF;
END $$;

DO $$
DECLARE q text; lvl int;
BEGIN
  SELECT pre_query INTO q FROM scheduled_tasks WHERE name='spend-governor-state';
  IF position('FOR UPDATE),' in q) = 0 THEN
    RAISE EXCEPTION '673 VERIFY: FOR UPDATE not present in stored text after replace — anchor missed';
  END IF;
  EXECUTE q;  -- the changed text must still run
  SELECT shed_level INTO lvl FROM governor_state WHERE id=1;
  IF lvl <> 0 THEN RAISE EXCEPTION '673 VERIFY: level % on NULL budget', lvl; END IF;
  RAISE NOTICE '673 OK: old-read now FOR UPDATE (snapshot-safe under overlap), text proven runnable, level 0.';
END $$;

COMMIT;

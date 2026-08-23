-- 570 — `agent_error_log` gets an index on the column every consumer filters by.
--
-- Owner instruction 2026-08-23, from the council's advisory objection on 567
-- (`9dc2e6b4-a8fd-476c-8080-ae23567e25c5`, guardian seat, low): *"agent_error_log is read by
-- diagnostics/monitoring across many pipelines … no index on error_code … an unindexed hot
-- predicate on a shared table is a blast-radius concern for every consumer."*
--
-- ── THE OBJECTION'S PREMISE IS WRONG AND ITS CONCLUSION IS RIGHT ────────────
-- The seat's stated worry was 567's hourly SWEEP, whose new arm filters on
-- `split_part(error_code, ':', 1)`. **Measured 2026-08-23, that is not a problem and an index
-- would not help it:** the sweep is already driven by `idx_error_log_time`, and `split_part` is
-- only a Filter applied to whatever the time index has already narrowed.
--
--   Bitmap Heap Scan … Recheck Cond: (occurred_at < now() - '30 days')
--     Filter: (split_part(error_code, ':', 1) = ANY (…))
--     cost=12.68  Buffers: shared hit=7  actual rows=0
--
-- Seven buffers. Adding an index for the sweep would have been an index added by reflex against a
-- measurement nobody took — which is this lane's own recurring mistake, and the reason the plan
-- was to measure before adding one.
--
-- **The READERS are a different story, and they are why this ships.** Measured on the live table
-- (46,036 rows), before and after, by creating the index inside a rolled-back transaction:
--
--   | query                                                   | before                    | after            |
--   |---------------------------------------------------------|---------------------------|------------------|
--   | strike ladder: error_code = $1 AND occurred_at > 7 days  | Index Scan on TIME index, | Index Scan,      |
--   |   (page_build_failure_guard.go:131)                      | **8,018 buffers**         | **3 buffers**    |
--   | content-loss-check family re-grade: error_code IN (3)    | **Seq Scan, 4,628 buf**   | Bitmap, **76**   |
--   | SELECT DISTINCT error_code (the registry check authority)| **Seq Scan, 4,628 buf**   | Index Only, 3,451|
--
-- The strike ladder is the one that matters: it runs on EVERY refused deploy stamp, and it was
-- reading eight thousand buffers to count rows in a seven-day window because the only usable index
-- was on `occurred_at`, so it scanned the whole window and filtered. 8,018 → 3.
--
-- `SELECT DISTINCT` improves only modestly (3,451 vs 4,628) and is stated rather than dressed up:
-- a full DISTINCT still walks every entry, it just walks a narrower one.
--
-- ── WHY (error_code, occurred_at DESC) AND NOT (error_code) ─────────────────
-- Every consumer that filters by code ALSO bounds or orders by time — the strike ladder's 7-day
-- window, `reconcile_superseded_reviews`'s `ORDER BY occurred_at DESC LIMIT 20`, and 567's own
-- sweep. The composite serves those from the index; a bare `(error_code)` would return the code's
-- whole population and re-filter. The DESC matches the readers' ordering.
--
-- ── LOCKING ────────────────────────────────────────────────────────────────
-- Plain `CREATE INDEX`, not CONCURRENTLY, because CONCURRENTLY cannot run inside a transaction and
-- this migration is a transaction (and the runner expects one). It takes a SHARE lock that blocks
-- writes to `agent_error_log` for the build. Measured build on the rolled-back trial: sub-second on
-- 46,036 rows. Writers here are best-effort error recorders that already tolerate failure, so a
-- sub-second stall is acceptable. **If this table is ever materially larger, revisit — the same
-- statement on millions of rows is a different decision.**
--
-- Rollback sidecar: 570_agent_error_log_indexes_its_own_code_ROLLBACK.sql

BEGIN;

-- ── 1. refuse if the shape is not what this migration expects ───────────────
DO $do$
DECLARE n int;
BEGIN
  IF to_regclass('public.agent_error_log') IS NULL THEN
    RAISE EXCEPTION '570 REFUSED: agent_error_log does not exist';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_code_time') THEN
    RAISE EXCEPTION '570 REFUSED: idx_error_log_code_time already exists — another session applied this';
  END IF;
  SELECT count(*) INTO n FROM pg_indexes WHERE tablename='agent_error_log';
  RAISE NOTICE '570: agent_error_log has % index(es) before this migration.', n;
END
$do$;

-- ── 2. the edit ─────────────────────────────────────────────────────────────
CREATE INDEX idx_error_log_code_time ON agent_error_log (error_code, occurred_at DESC);

-- ── 3. verify (DO/RAISE — ON_ERROR_STOP ignores a non-empty SELECT) ─────────
DO $do$
DECLARE def text; used bool := false; rec record;
BEGIN
  SELECT indexdef INTO def FROM pg_indexes
   WHERE tablename='agent_error_log' AND indexname='idx_error_log_code_time';
  IF def IS NULL THEN
    RAISE EXCEPTION '570: the index was not created';
  END IF;

  -- 3a. it is the index this migration intends, not merely AN index
  IF def NOT LIKE '%(error_code, occurred_at DESC)%' THEN
    RAISE EXCEPTION '570: index exists but its definition is %, not (error_code, occurred_at DESC)', def;
  END IF;

  -- 3b. NEGATIVE CONTROL: the four pre-existing indexes must all survive. A
  --     migration that dropped one while adding another would still pass 3a.
  IF NOT (EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='agent_error_log_pkey')
      AND EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_agent')
      AND EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_site')
      AND EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_time')
      AND EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='agent_error_log' AND indexname='idx_error_log_unresolved')) THEN
    RAISE EXCEPTION '570: a pre-existing agent_error_log index was lost';
  END IF;

  -- 3c. THE BEHAVIOURAL CLAIM, asserted rather than assumed. An index that
  --     exists and is never chosen buys nothing, and "I created it" is not
  --     evidence the planner will use it. Read the PLAN of the query this
  --     migration is FOR (the strike ladder) and require the new index in it.
  ANALYZE agent_error_log;
  FOR rec IN EXECUTE
    'EXPLAIN SELECT count(*) FROM agent_error_log
       WHERE error_code = ''DEPLOY_STAMP_REFUSED_ON_SKIP''
         AND occurred_at > NOW() - INTERVAL ''7 days'''
  LOOP
    IF position('idx_error_log_code_time' in rec."QUERY PLAN") > 0 THEN
      used := true;
    END IF;
  END LOOP;
  IF NOT used THEN
    RAISE EXCEPTION '570: the planner does NOT choose idx_error_log_code_time for the strike-ladder query — the index exists but buys nothing';
  END IF;

  RAISE NOTICE '570: applied. error_code is indexed with occurred_at DESC, and the planner uses it for the strike ladder.';
END
$do$;

COMMIT;

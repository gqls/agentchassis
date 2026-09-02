-- 701_content_sources_daily_fetch_interval_ROLLBACK.sql
--
-- Reverses 701 (news sources 6h -> 24h, owner decision 2026-09-02).
-- Sidecar: the migration runner NEVER runs this (SIDECAR_RE excludes
-- _ROLLBACK). Run it by hand, deliberately.
--
-- ⚠ RESTORES THE ACTUAL PREVIOUS VALUES, not the old default. The pre-701
-- intervals were NOT uniform — 3h x2 (relojistas), 4h x4 (relojistas,
-- dartsonline), 6h x67 — so setting everything back to '06:00:00' would be a
-- silent data change dressed as a rollback. The values come from
-- bak_content_sources_fetch_interval_20260902, which 701 wrote.
--
-- ⚠ AND BE CLEAR WHAT YOU ARE ROLLING BACK INTO. Reverting to 6h restores
-- fetch_interval == the 21600 s trigger cadence, which is the exact
-- precondition of bugs_closed/410's phase lock. That is SURVIVABLE — the
-- half-cadence look-ahead is live in both layers and is what keeps it from
-- locking — but it also restores the cap contention: 14 sites due every pass
-- against a LIMIT of 10, i.e. a ~9 h effective cadence and 4 sites late by
-- rotation. Read bugs_open/316 before deciding that is what you want.

BEGIN;

DO $$
DECLARE n integer;
BEGIN
    SELECT count(*) INTO n FROM information_schema.tables
     WHERE table_name = 'bak_content_sources_fetch_interval_20260902';
    IF n = 0 THEN
        RAISE EXCEPTION 'ROLLBACK 701: backup table bak_content_sources_fetch_interval_20260902 is missing — the previous intervals were not uniform and CANNOT be reconstructed from the old default. Stop.';
    END IF;
END $$;

-- Restore the per-row intervals and due stamps exactly as they were.
UPDATE content_sources cs
   SET fetch_interval = b.fetch_interval,
       next_fetch_at  = b.next_fetch_at
  FROM bak_content_sources_fetch_interval_20260902 b
 WHERE b.id = cs.id;

-- Restore the column default, or every newly seeded source stays at 24h and the
-- rollback is only half done.
ALTER TABLE content_sources ALTER COLUMN fetch_interval SET DEFAULT '06:00:00'::interval;

DO $$
DECLARE bad integer; dflt text;
BEGIN
    SELECT count(*) INTO bad
      FROM content_sources cs
      JOIN bak_content_sources_fetch_interval_20260902 b ON b.id = cs.id
     WHERE cs.fetch_interval IS DISTINCT FROM b.fetch_interval;
    IF bad > 0 THEN
        RAISE EXCEPTION 'ROLLBACK 701 VERIFY: % row(s) do not match the backup.', bad;
    END IF;

    SELECT pg_get_expr(d.adbin, d.adrelid) INTO dflt
      FROM pg_attrdef d
      JOIN pg_attribute a ON a.attrelid = d.adrelid AND a.attnum = d.adnum
     WHERE d.adrelid = 'content_sources'::regclass AND a.attname = 'fetch_interval';
    IF dflt IS NULL OR dflt NOT LIKE '%06:00:00%' THEN
        RAISE EXCEPTION 'ROLLBACK 701 VERIFY: column default is % — not restored to 6h.', COALESCE(dflt,'(none)');
    END IF;

    RAISE NOTICE 'ROLLBACK 701 VERIFY OK: intervals and due stamps match the backup, default restored to 6h.';
END $$;

COMMIT;

-- Leave bak_content_sources_fetch_interval_20260902 in place. It is small and
-- it is the only record of the pre-701 state; drop it deliberately, later, not
-- as part of a rollback.

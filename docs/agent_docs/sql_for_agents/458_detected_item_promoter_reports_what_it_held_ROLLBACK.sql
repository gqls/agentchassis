-- ROLLBACK for 458 — restores the detected-item-promoter pre_query to its
-- pre-458 text (430 + 444 + 454), verbatim.
--
-- The pre-image is md5 1d1efee2913929db7b6b5395d8421ecc, length 2340, captured
-- live 2026-08-17 22:5xZ immediately before 458 was applied.
--
-- WHAT THIS GIVES UP, so the choice is deliberate: the promoter goes back to
-- reporting nothing about what its doors held, and back to returning a row on
-- every tick (so it claims the `maintenance` concurrency slot even when idle).
-- It does NOT change which rows are promotable — 458 was behaviour-identical on
-- that set by assertion.
--
-- This does not un-ship the scheduler logging change (cmd/scheduler/main.go);
-- that is a binary, and after a rollback it will simply log the two-column
-- result instead of four.

BEGIN;

DO $$
DECLARE
    live_md5 text;
BEGIN
    SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF live_md5 IS NULL THEN
        RAISE EXCEPTION '458 ROLLBACK: no detected-item-promoter row';
    END IF;
    -- Refuse if the live text is neither 458's output nor already the pre-image:
    -- that means a THIRD edit landed and blind restoration would revert it.
    IF live_md5 = '1d1efee2913929db7b6b5395d8421ecc' THEN
        RAISE NOTICE '458 ROLLBACK: already at the pre-458 text — nothing to do.';
    ELSIF (SELECT pre_query NOT LIKE '%AS held_detail%' FROM scheduled_tasks WHERE name = 'detected-item-promoter') THEN
        RAISE EXCEPTION '458 ROLLBACK: ABORTING — the live pre_query is neither 458''s output nor the pre-image (md5 %). Another session has edited it since; read the live column before restoring anything.', live_md5;
    END IF;
END $$;

UPDATE scheduled_tasks
SET pre_query = $Q$
    WITH candidates AS (
        SELECT wi.id
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          -- DOOR-CLOSER 1 (444): only pipelines this transition has ever produced.
          AND wi.pipeline IN ('build', 'content', 'design')
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
          -- KNOWN-GOOD (430, corrected by 454): `verified` is a completion that
          -- also passed verification. Counting only 'complete' would hold a pair
          -- whose every success has been verified.
          AND EXISTS (
            SELECT 1 FROM site_work_items done
            WHERE done.item_type = wi.item_type
              AND done.handler_agent = wi.handler_agent
              AND done.status IN ('complete', 'verified')
          )
          -- DOOR-CLOSER 2 (444, corrected by 454): >=5 terminal outcomes => must
          -- still be >=25% good. NB `failed` rows have no completed_at, so this
          -- keys on status only.
          AND (
            SELECT (c + f) < 5 OR c >= 0.25 * (c + f)
            FROM (
              SELECT count(*) FILTER (WHERE h.status IN ('complete', 'verified')) AS c,
                     count(*) FILTER (WHERE h.status = 'failed')                  AS f
              FROM site_work_items h
              WHERE h.item_type = wi.item_type
                AND h.handler_agent = wi.handler_agent
            ) hist
          )
        ORDER BY wi.created_at ASC
        LIMIT 20
    ),
    promoted AS (
        UPDATE site_work_items wi
        SET status = 'triaged',
            triaged_at = now(),
            spec = jsonb_set(COALESCE(wi.spec, '{}'::jsonb), '{original_pipeline}', to_jsonb(wi.pipeline)),
            pipeline = 'build',
            updated_at = now()
        FROM candidates c
        WHERE wi.id = c.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent
    )
    SELECT COUNT(*)::text AS promoted,
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs
    FROM promoted
    WHERE (SELECT COUNT(*) FROM promoted) > 0
$Q$,
    updated_at = now()
WHERE name = 'detected-item-promoter';

DO $$
DECLARE
    live_md5 text;
BEGIN
    SELECT md5(pre_query) INTO live_md5 FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF live_md5 <> '1d1efee2913929db7b6b5395d8421ecc' THEN
        RAISE EXCEPTION '458 ROLLBACK: restored text does not match the captured pre-image (got %, want 1d1efee2913929db7b6b5395d8421ecc). The restoration is NOT byte-exact — do not commit it.', live_md5;
    END IF;
    RAISE NOTICE '458 ROLLBACK: pre_query restored byte-exactly to the pre-458 text.';
END $$;

COMMIT;

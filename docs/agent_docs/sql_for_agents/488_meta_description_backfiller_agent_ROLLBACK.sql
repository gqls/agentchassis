-- 488 ROLLBACK — remove the meta-description-backfiller agent
--
-- Soft-deletes the row rather than hard-deleting it, so any orchestration that
-- already ran against it keeps a resolvable agent_definitions reference.
--
-- ⚠ ROLLING THIS BACK LEAVES `save_page_meta_description` WITH NO CALLER, which
-- is the state bugs_open/320 was filed about: the action exists, the register
-- says so, and nothing drives it. That is worse than not having shipped it,
-- because the register then reads as "this is handled". If you roll this back,
-- update register SEO-004 to say the driver was withdrawn and why.
--
-- Descriptions already written are NOT reverted. They are ordinary page copy at
-- that point, indistinguishable from any other, and blanking them would recreate
-- mechanism M2 by hand.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '488 ROLLBACK: expected exactly 1 live meta-description-backfiller row, found %', n;
  END IF;
END $$;

SELECT snapshot_agent('meta-description-backfiller', '488_ROLLBACK: pre-withdrawal');

UPDATE agent_definitions
   SET is_active = false,
       deleted_at = now(),
       updated_at = now()
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION '488 ROLLBACK VERIFY: % live rows remain', n;
  END IF;
  RAISE NOTICE '488 ROLLBACK OK — the backfiller is withdrawn; SEO-004 now has NO caller, update the register entry to say so';
END $$;

COMMIT;

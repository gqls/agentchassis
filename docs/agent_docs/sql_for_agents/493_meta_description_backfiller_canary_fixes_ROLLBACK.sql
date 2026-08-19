-- 493 ROLLBACK — restore the pre-canary-fix backfiller config
--
-- ⚠ ROLLING THIS BACK RE-ARMS BOTH DEFECTS: the agent goes back to reporting
-- "nothing to do" on every site (bugs_open/313's silent-skip shape) and, if that
-- were then fixed, to showing the writer raw CSS. There is no good reason to run
-- this except to undo a bad regex.
--
-- Restores from the snapshot 493 took, rather than by reversing each edit: the
-- forward migration rewrote a whole query string, and reversing that by string
-- replacement is how a half-restored query happens.

BEGIN;

DO $$
DECLARE n int; snap_id uuid;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '493 ROLLBACK: expected exactly 1 live row, found %', n;
  END IF;

  SELECT id INTO snap_id FROM agent_definitions
   WHERE type = 'meta-description-backfiller' AND COALESCE(is_snapshot,false) = true
     AND description LIKE '%493_canary_fixes: pre-update%'
   ORDER BY created_at DESC LIMIT 1;

  IF snap_id IS NULL THEN
    RAISE EXCEPTION '493 ROLLBACK: no 493 pre-update snapshot found — refusing to guess at the previous config';
  END IF;

  UPDATE agent_definitions live
     SET default_config = snap.default_config, updated_at = now()
    FROM agent_definitions snap
   WHERE snap.id = snap_id
     AND live.type = 'meta-description-backfiller'
     AND live.is_active AND COALESCE(live.is_snapshot,false) = false AND live.deleted_at IS NULL;
END $$;

DO $$
DECLARE fmt text;
BEGIN
  SELECT default_config#>>'{workflow,steps,load_pages_missing_meta,config,output_format}'
    INTO fmt FROM agent_definitions
   WHERE type='meta-description-backfiller'
     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF fmt IS DISTINCT FROM 'array' THEN
    RAISE EXCEPTION '493 ROLLBACK VERIFY: output_format is %, expected the restored pre-fix value array', fmt;
  END IF;
  RAISE NOTICE '493 ROLLBACK OK — pre-canary-fix config restored; BOTH defects are re-armed';
END $$;

COMMIT;

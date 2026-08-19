-- ROLLBACK sidecar for 482. Removes dark_section_audit from the claimed-item-timeout
-- exclusion list, restoring the gate-2-only contract.
-- ⚠ Applying this re-opens bugs_open/317: the sweep can again auto-complete a
-- dark_section_audit item with neither completion gate running.
DO $$
DECLARE n int;
BEGIN
  UPDATE scheduled_tasks
     SET pre_query = replace(pre_query, ', ''dark_section_audit''', '')
   WHERE name = 'claimed-item-timeout';
  SELECT count(*) INTO n FROM scheduled_tasks
   WHERE name = 'claimed-item-timeout' AND pre_query LIKE '%dark_section_audit%';
  IF n <> 0 THEN
    RAISE EXCEPTION 'ABORT: rollback did not remove dark_section_audit (still matched % rows).', n;
  END IF;
END $$;

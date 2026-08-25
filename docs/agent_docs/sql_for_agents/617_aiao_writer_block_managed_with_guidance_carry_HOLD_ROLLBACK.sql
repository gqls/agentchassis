-- 617_aiao_writer_block_managed_with_guidance_carry_HOLD_ROLLBACK.sql
--
-- Restores the exact pre-617 `evidence_base` document for ai-agent-orchestration.com from
-- the migration_backups row 617 wrote: writer_block_managed goes back OFF, the
-- writer_block_guidance key disappears, the `aao-architecture` fact is dropped, and 611's
-- hand-written writer_block returns.
--
-- ⚠ WHAT THIS RESTORES IS THE UNMANAGED STATE — the hand-written block that regeneration
-- cannot touch and that a hand-typed stand-in once reached the public through (bugs_open/387).
-- It is a safe state, not a better one.
-- ⚠ If the daily refresher has run since 617 applied, fact VALUES may have moved; this rollback
-- puts the pre-617 values back until the next ~09:06Z refresh re-reads them. That is one day of
-- a slightly stale register, not a defect — the writer_lines are floors and do not print values.

BEGIN;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT '2a8ebf9c-20a2-4c39-b191-840b012371da', 'evidence_base',
       b.old_value->'data', 'rollback', 'rollback',
       'restored by rollback of 617 — writer_block_managed OFF again, 611''s hand-written block back',
       true, '617_rollback'
FROM migration_backups b
WHERE b.migration_name='617_aiao_writer_block_managed_with_guidance_carry';

DO $$
DECLARE cur jsonb; n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'rollback 617: expected 1 current row, found %', n; END IF;
  SELECT data INTO cur FROM site_specs WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF cur IS DISTINCT FROM (SELECT old_value->'data' FROM migration_backups WHERE migration_name='617_aiao_writer_block_managed_with_guidance_carry') THEN
    RAISE EXCEPTION 'rollback 617: current document is not byte-identical to the backup';
  END IF;
  IF coalesce((cur->>'writer_block_managed')::bool,false) OR cur ? 'writer_block_guidance' THEN
    RAISE EXCEPTION 'rollback 617: managed/guidance still present after restore';
  END IF;
  RAISE NOTICE 'rollback 617 OK: pre-617 document restored exactly; site is UNMANAGED again.';
END $$;

COMMIT;

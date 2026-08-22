-- 557_..._ROLLBACK.sql
--
-- Restores the exact pre-557 `evidence_base` document for ai-agent-orchestration.com
-- from the migration_backups row 557 wrote. Not a reverse-edit, so it cannot drift.
--
-- ⚠ THE STATE THIS RETURNS TO IS THE ONE THAT BLOCKS THE `pricing` REBUILD. The
-- restored `writer_block` instructs the writer to Write "170+ agents", a wording none
-- of the `aao-agent-definitions` context terms match, so the claims checker refuses
-- every page that obeys it. It also re-announces "175 as of 2026-07-26" and "14 live
-- sites" against live values of 199 and 25. Roll back only to isolate a worse
-- regression elsewhere, and say which.
--
-- Supersede semantics, same as the forward file: close the current row, reopen the
-- backed-up document as a new current row. The 557 row is left closed in history —
-- forward-only, nothing is deleted.

BEGIN;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT '2a8ebf9c-20a2-4c39-b191-840b012371da', 'evidence_base',
       b.old_value->'data', 'rollback', 'rollback',
       'restored by rollback of migration 557 — the pricing-blocking writer_block is BACK',
       true, '557_rollback'
FROM migration_backups b
WHERE b.migration_name='557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate';

DO $$
DECLARE cur jsonb; n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN
    RAISE EXCEPTION 'rollback 557: expected 1 current row, found %', n;
  END IF;

  SELECT data INTO cur FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

  IF cur IS DISTINCT FROM (SELECT old_value->'data' FROM migration_backups
        WHERE migration_name='557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate') THEN
    RAISE EXCEPTION 'rollback 557: current document is not byte-identical to the backup';
  END IF;

  RAISE NOTICE 'rollback 557 OK: pre-557 evidence_base restored exactly. The pricing rebuild is blocked again.';
END $$;

COMMIT;

-- 613_..._ROLLBACK.sql
--
-- Restores the exact pre-613 `evidence_base` document (the `611` row) for
-- ai-agent-orchestration.com from the migration_backups row 613 wrote.
--
-- ⚠ WHAT THIS RESTORES IS A LATENT DEFECT, NOT A WORKING STATE. The restored
-- writer_lines carry three frozen dates beside live `{value}` substitutions
-- ("({value} as of 2026-07-26)"), and two that would publish figures the
-- writer_block explicitly forbids (an exact daily orchestration count; a work-item
-- total from a REAPED ledger that falls as well as rises). None of that is visible
-- today because `writer_block_managed` is not set on this site — which is exactly why
-- rolling this back is easy to do and hard to notice.
--
-- ⚠ IT ALSO RESTORES `611` UNCHANGED, because 613 never altered it: `writer_block` and
-- `banned_claims` are byte-identical either side of this rollback. So rolling back 613
-- does NOT reintroduce the `NNN+` stand-in and does NOT undo the live incident fix.
-- If `NNN` is on a page, this file is not the cause and rolling it back will not help.

BEGIN;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT '2a8ebf9c-20a2-4c39-b191-840b012371da', 'evidence_base',
       b.old_value->'data', 'rollback', 'rollback',
       'restored by rollback of 613 — the frozen-date writer_lines are BACK (latent until writer_block_managed is set)',
       true, '613_rollback'
FROM migration_backups b
WHERE b.migration_name='613_aiao_writer_lines_stop_carrying_frozen_dates_and_forbidden_figures';

DO $$
DECLARE cur jsonb; n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'rollback 613: expected 1 current row, found %', n; END IF;

  SELECT data INTO cur FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

  IF cur IS DISTINCT FROM (SELECT old_value->'data' FROM migration_backups
        WHERE migration_name='613_aiao_writer_lines_stop_carrying_frozen_dates_and_forbidden_figures') THEN
    RAISE EXCEPTION 'rollback 613: current document is not byte-identical to the backup';
  END IF;

  RAISE NOTICE 'rollback 613 OK: pre-613 document restored exactly. 611 is unaffected either way.';
END $$;

COMMIT;

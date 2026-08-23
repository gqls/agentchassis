-- ROLLBACK for 569 — restores both input_schemas from the backup 569 took.
--
-- Restores the WHOLE input_schema from the backup rather than reversing the two
-- jsonb_set calls. A reverse edit would have to re-create the exact prior source
-- strings, and if the component has been regenerated in between it would write a
-- phantom source back into a schema that had moved on — re-arming the very defect
-- 569 repaired. Restoring the backed-up document cannot half-apply.
--
-- ⚠ NOTE WHAT ROLLING BACK MEANS HERE: it re-declares `config` and `query.pages`,
-- both of which the CLC-018 birth gate now REFUSES. The components keep working
-- (the gate only fires on generation), but the next regeneration of either would be
-- refused until the source is repointed again — which is the gate doing its job, not
-- a new defect. Roll back only to undo a mistake in 569 itself.

BEGIN;

DO $r$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_569_phantom_source_repair_20260823;
  IF n < 2 THEN
    RAISE EXCEPTION '569 ROLLBACK ABORT: backup holds % row(s), expected at least 2.', n;
  END IF;
END
$r$;

UPDATE content_components c
SET input_schema = b.input_schema
FROM bak_569_phantom_source_repair_20260823 b
WHERE c.id = b.id;

DO $v$
DECLARE n_back int;
BEGIN
  SELECT count(*) INTO n_back FROM content_components
  WHERE is_active AND (
    (name='info-card-grid'   AND input_schema->'fields'->'carousel'->>'source'     = 'config') OR
    (name='Latest News Feed' AND input_schema->'fields'->'insights_url'->>'source' = 'query.pages'));
  IF n_back <> 2 THEN
    RAISE EXCEPTION '569 ROLLBACK VERIFY FAILED: expected both original sources restored, found %.', n_back;
  END IF;
  RAISE NOTICE '569 ROLLBACK OK: both input_schemas restored to their pre-569 state.';
END
$v$;

COMMIT;

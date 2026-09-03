-- 754_d4b_gate_query_types_its_ctx_parameter_ROLLBACK.sql — restores the UNTYPED gate query.
-- ⚠ After this, the gate FAILS at runtime (42P18) and FALLS OPEN on every council run: reviews
-- run, nothing is refused, and nothing says so. Roll back only if the typed text itself
-- misbehaves; the fall-open is the known cost, stated here so it is not a surprise.

BEGIN;

DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF q IS NULL THEN RAISE EXCEPTION '754 ROLLBACK REFUSED: no gate_spend_governor step.'; END IF;
  n := (length(q) - length(replace(q, ', $1::text, COALESCE(gs.shed_level, 0)', ''))) / length(', $1::text, COALESCE(gs.shed_level, 0)');
  IF n <> 1 THEN RAISE EXCEPTION '754 ROLLBACK REFUSED: typed anchor found % times, expected 1 — 754 not applied, or the text moved.', n; END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,gate_spend_governor,config,query}',
  to_jsonb(replace(default_config#>>'{workflow,steps,gate_spend_governor,config,query}', ', $1::text, COALESCE(gs.shed_level, 0)', ', $1, COALESCE(gs.shed_level, 0)')))
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE q text;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('$1::text' in q) > 0 THEN RAISE EXCEPTION '754 ROLLBACK VERIFY: cast still present'; END IF;
  IF md5(q) <> '3b3fa4a4315656960483147f4fa32eec' THEN RAISE EXCEPTION '754 ROLLBACK VERIFY: gate query md5 % is not the pre-754 text', md5(q); END IF;
  RAISE NOTICE '754 ROLLBACK OK: untyped gate query restored (md5 3b3fa4a4…). REMINDER: it fails 42P18 at runtime and the gate FALLS OPEN.';
END $$;

COMMIT;

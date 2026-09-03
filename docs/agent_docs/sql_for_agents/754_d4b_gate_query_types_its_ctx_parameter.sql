-- 754_d4b_gate_query_types_its_ctx_parameter.sql — the council gate (752) FAILED at runtime on
-- its first live run and fell OPEN; this makes it bind.
--
-- THE DEFECT, from the live canary (council run 83186fd9, 2026-09-03 21:34:29Z, __step_error):
--   step gate_spend_governor failed: failed to execute action query_database: query failed:
--   ERROR: could not determine data type of parameter $1 (SQLSTATE 42P18)
-- The gate's query uses $1 (bound from $ctx.correlation_id) ONLY inside format(...), whose
-- arguments are VARIADIC "any" — so when the driver sends the parameter untyped, Postgres
-- cannot infer its type and refuses the statement. error_step = load_schema_hint then did
-- exactly what it was designed to do: the review RAN. Which is why nothing noticed: no output,
-- no refusal, a normal-looking round. Every council run since 752 applied (21:24Z) has fallen
-- open at the gate — the owner's "first" was NOT enforced for that window.
--
-- WHY 752's VERIFY COULD NOT SEE IT. It EXECUTEd the query text with a LITERAL spliced in for
-- $1. A literal has a type; a bound parameter does not until something gives it one. The verify
-- rehearsed the SQL and not the BINDING — the exact runtime piece its own risks section named
-- as unrehearsable. It was: `PREPARE` with unspecified parameter types forces the same server-
-- side inference the driver triggers, so the verify below (and 752's daily VERIFY, arm 5b) now
-- PREPAREs the live text and requires it to succeed — a mutation dropping the cast fails it.
--
-- THE FIX: one token — `$1::text` — in the stored gate query. Nothing else in the row changes.
-- GUARD: refuses unless the live gate query carries the untyped `, $1,` exactly once and no
-- `$1::text`; md5-pins the whole gate step's query text as it stands post-752.
-- Rollback: 754_..._ROLLBACK.sql (restores the untyped text; says the gate will fall open).

BEGIN;

DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF q IS NULL THEN RAISE EXCEPTION '754 REFUSED: council-gate has no gate_spend_governor step — 752 not applied.'; END IF;
  IF position('$1::text' in q) > 0 THEN RAISE EXCEPTION '754 REFUSED: gate query already types $1 — already applied (replay).'; END IF;
  n := (length(q) - length(replace(q, ', $1, COALESCE(gs.shed_level, 0)', ''))) / length(', $1, COALESCE(gs.shed_level, 0)');
  IF n <> 1 THEN RAISE EXCEPTION '754 REFUSED: expected the untyped ", $1, COALESCE(gs.shed_level, 0)" anchor exactly once in the gate query, found % — drifted, investigate.', n; END IF;
  IF md5(q) <> '3b3fa4a4315656960483147f4fa32eec' THEN
    RAISE EXCEPTION '754 REFUSED: gate query md5 % is not the post-752 text this was written against — drifted, investigate before overwriting.', md5(q);
  END IF;
  -- prove the defect is PRESENT before fixing it (a fix for a defect that is not there is a different change)
  BEGIN
    EXECUTE 'PREPARE p754_pre AS ' || q;
    EXECUTE 'DEALLOCATE p754_pre';
    RAISE EXCEPTION '754 REFUSED: the untyped gate query PREPAREs cleanly — the 42P18 defect is not reproducible here; investigate before changing the text.';
  EXCEPTION WHEN indeterminate_datatype THEN
    NULL;   -- 42P18: expected — this is the live failure, reproduced server-side
  END;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,gate_spend_governor,config,query}',
  to_jsonb(replace(default_config#>>'{workflow,steps,gate_spend_governor,config,query}', ', $1, COALESCE(gs.shed_level, 0)', ', $1::text, COALESCE(gs.shed_level, 0)')))
WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE q text; n int; adm boolean; lvl int; body text; saved_level int; saved_enabled boolean;
BEGIN
  SELECT default_config#>>'{workflow,steps,gate_spend_governor,config,query}' INTO q
  FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF position('$1::text' in q) = 0 THEN RAISE EXCEPTION '754 VERIFY: the cast did not land'; END IF;
  n := (length(q) - length(replace(q, '$1', ''))) / 2;
  IF n <> 1 THEN RAISE EXCEPTION '754 VERIFY: $1 appears % times, expected 1', n; END IF;

  -- THE ARM THAT MATTERS: the live text must PREPARE with an UNSPECIFIED parameter type —
  -- the same server-side inference the driver triggers at runtime.
  EXECUTE 'PREPARE p754 AS ' || q;
  EXECUTE 'DEALLOCATE p754';

  -- and still behave: literal-spliced execution at the live level and at a forced L3
  SELECT shed_level INTO saved_level FROM governor_state WHERE id=1;
  SELECT enabled INTO saved_enabled FROM governor_config WHERE id=1;
  EXECUTE 'SELECT admitted FROM (' || replace(q, '$1::text', '''754-verify-corr''') || ') g' INTO adm;
  IF adm IS DISTINCT FROM governor_admits_agent('council-gate') THEN RAISE EXCEPTION '754 VERIFY: gate disagrees with governor_admits_agent at the live level'; END IF;
  UPDATE governor_config SET enabled = true WHERE id=1;
  UPDATE governor_state SET shed_level = 3 WHERE id=1;
  EXECUTE 'SELECT admitted, shed_level, body FROM (' || replace(q, '$1::text', '''754-verify-corr''') || ') g' INTO adm, lvl, body;
  IF adm IS DISTINCT FROM false OR lvl <> 3 OR position('754-verify-corr' in body) = 0 THEN
    RAISE EXCEPTION '754 VERIFY: at L3 the gate should refuse and name the correlation; got admitted=% level=% body=%', adm, lvl, left(body,120);
  END IF;
  UPDATE governor_state SET shed_level = saved_level WHERE id=1;
  UPDATE governor_config SET enabled = saved_enabled WHERE id=1;

  -- nothing else on the row moved
  IF (SELECT default_config#>>'{workflow,start_step}' FROM agent_definitions WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) <> 'gate_spend_governor' THEN
    RAISE EXCEPTION '754 VERIFY: start_step moved';
  END IF;
  SELECT count(*) INTO n FROM agent_definitions, jsonb_object_keys(default_config#>'{workflow,steps}') k
   WHERE type='council-gate' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 48 THEN RAISE EXCEPTION '754 VERIFY: step count % (expected 48)', n; END IF;

  RAISE NOTICE '754 OK: gate query types its $ctx parameter; PREPARE with an unspecified parameter type succeeds (the 42P18 path is closed); flips at L3; row otherwise untouched.';
END $$;

COMMIT;

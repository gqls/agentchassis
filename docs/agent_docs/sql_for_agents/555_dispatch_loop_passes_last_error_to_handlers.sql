-- 555_dispatch_loop_passes_last_error_to_handlers.sql
--
-- bugs_open/345, THE THIRD HALF — without this the other two are inert.
--
-- The Go half (load_work_item_actions.go) puts the previous failure text on the
-- work-item map as `last_error`, and migration 533 makes component-creator's
-- prompt read `{{if .input_data.last_error}}`. Both are live on v1.0.1322+ and
-- CANNOT FIRE, because `build-dispatch-loop`'s
-- `process_item → sub_workflow → call_handler` is a `call_agent` step with an
-- explicit `input_mapping` — an ALLOW-LIST of 14 keys — and `last_error` is not
-- one of them. The handler's `input_data` is built only from that list, so the
-- key never crosses the boundary.
--
-- Found by the council gate's guardian seat (corr 67b07528, round 4), which
-- asked whether an intermediate allow-list gate existed. It did.
-- [MEASURED 2026-08-22] census of every live `call_agent` input_mapping,
-- top-level and sub-workflow: 73 mappings, ZERO pass `last_error`. And
-- `collected_data->'input_data' ? 'last_error'` = 0 across all history — which
-- this lane had wrongly attributed to "no retry has happened yet".
--
-- SCOPE, deliberately narrow: only `build-dispatch-loop`. That is the path
-- `needs_new_component` items travel (pipeline='build', handler_agent=
-- 'component-creator'), which is 345's whole population.
-- `site-work-orchestrator`'s `fix_items_loop → call_handler` is a DIFFERENT
-- loop variable (`current_fix_item`) serving discovery-check fix items, whose
-- handlers have no prompt reading this key — widening a shared seam with no
-- consumer is what this estate's optional-field rulings exist to prevent, so it
-- is NAMED here and left alone.
--
-- `last_error?` is OPTIONAL-marked: an optional path that RESOLVES is forwarded
-- and one that is MISSING is skipped, so an item with no recorded failure
-- produces a byte-identical payload. There is no ordering constraint against
-- the Go half or 533 — with any one of the three absent the block simply never
-- renders — and this file does not claim one.
--
-- ROLLBACK: 555_..._ROLLBACK.sql removes the single key.

BEGIN;

DO $guard$
DECLARE
  m jsonb;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '555: expected exactly 1 live build-dispatch-loop row, found %', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF m IS NULL THEN
    RAISE EXCEPTION '555: call_handler input_mapping not found at the expected path — the workflow has been restructured; re-read it before editing';
  END IF;
  IF m ? 'last_error?' OR m ? 'last_error' THEN
    RAISE EXCEPTION '555: last_error is already mapped — not idempotent by design';
  END IF;
  -- Pin the anchor we are extending, so a restructured mapping cannot be
  -- silently appended to.
  IF NOT (m ? 'work_item_id') THEN
    RAISE EXCEPTION '555: input_mapping does not carry work_item_id — this is not the mapping this migration was written against';
  END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,last_error?}',
      '"current_item.last_error"'::jsonb,
      true
    ),
    updated_at = NOW()
WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $verify$
DECLARE
  m jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF NOT (m ? 'last_error?') THEN
    RAISE EXCEPTION '555 VERIFY: the key was not added';
  END IF;
  IF m->>'last_error?' <> 'current_item.last_error' THEN
    RAISE EXCEPTION '555 VERIFY: the key maps to % , expected current_item.last_error', m->>'last_error?';
  END IF;
  IF NOT (m ? 'work_item_id' AND m ? 'spec' AND m ? 'site_id') THEN
    RAISE EXCEPTION '555 VERIFY: pre-existing keys did not survive the write';
  END IF;
  RAISE NOTICE '555: build-dispatch-loop call_handler now forwards last_error (mapping has % keys)',
    (SELECT count(*) FROM jsonb_object_keys(m));
END
$verify$;

COMMIT;

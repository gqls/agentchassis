-- 564_dispatch_loop_passes_last_error_code.sql
--
-- bugs_open/345 part 1, the WIRE for the code. Without this, migration 563 is
-- INERT — and that is not a guess, it is the failure this bug already paid for
-- once.
--
-- THE PRECEDENT, because it is the whole reason this file exists separately.
-- 345's first fix shipped a Go half (the loader put `last_error` on the item
-- map) and a prompt half (533 read `{{if .input_data.last_error}}`), both live
-- on v1.0.1322+, and they could not fire: `build-dispatch-loop`'s
-- `process_item → sub_workflow → call_handler` is a `call_agent` step whose
-- `input_mapping` is a STRICT ALLOW-LIST — the child's `input_data` is built
-- from that list and nothing else (input_contracts/input_mapping.go
-- ResolveInputMapping builds a fresh map and writes only mapped keys). The key
-- was not in it. The lane had measured `collected_data->'input_data' ?
-- 'last_error'` = 0 across all history and explained the zero away as "no retry
-- has happened yet"; the council gate's guardian seat found the real cause in
-- round 4, and migration 555 wired it. It is demand-proven since: 5
-- orchestrations carry the key, and item ceea0c07 was refused at 12:18:43Z on
-- 2026-08-22, re-dispatched at 12:51 carrying it, and COMPLETED at 12:53:07.
--
-- So `last_error_code` needs the identical wire, for the identical reason. A
-- reader without a wire is a fix connected to nothing.
--
-- SCOPE, deliberately identical to 555's: only `build-dispatch-loop`. That is
-- the path `needs_new_component` items travel (pipeline='build', handler_agent=
-- 'component-creator'), which is 345's whole population.
-- `site-work-orchestrator`'s `fix_items_loop → call_handler` is a DIFFERENT loop
-- variable (`current_fix_item`) serving discovery-check fix items, whose
-- handlers have no prompt reading either key — widening a shared seam with no
-- consumer is what this estate's optional-field rulings exist to prevent, so it
-- is NAMED here and left alone, exactly as 555 named it.
--
-- OPTIONAL-MARKED (`last_error_code?`). An optional path that resolves is
-- forwarded; one that is missing is SILENTLY SKIPPED. That silence is why the
-- verify block below asserts the key positively and asserts the surviving key
-- count, rather than trusting the UPDATE's row count: a mis-typed source path
-- would apply cleanly and forward nothing for ever.
--
-- The RFC_022 optional-key budget does not bind here: `cmd/config-key-audit/
-- optionalexplicitwires.go` states that a `?` key inside `input_mapping` is not
-- reported by the acks audit, which covers the ExtractActionInputs/step-config
-- surface only. No acknowledgement is owed.
--
-- ORDERING: none. The loader sets the code key only when the producer recorded
-- one, so an item without a classified failure produces a byte-identical
-- payload; and 563's block renders nothing when the key is absent.
--
-- ROLLBACK: 564_..._ROLLBACK.sql removes the single key.

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
    RAISE EXCEPTION '564: expected exactly 1 live build-dispatch-loop row, found %', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'
         ->'steps'->'call_handler'->'config'->'input_mapping'
    INTO m
  FROM agent_definitions
   WHERE type='build-dispatch-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF m IS NULL THEN
    RAISE EXCEPTION '564: call_handler input_mapping not found at the expected path — the workflow has been restructured; re-read it before editing';
  END IF;

  -- 555 must already be in place: the code without the message is useless, and
  -- their absence would mean this lane's assumptions about the row are stale.
  IF NOT (m ? 'last_error?') THEN
    RAISE EXCEPTION '564: migration 555''s last_error? key is missing — apply 555 first; a code with no message feeds the prompt nothing';
  END IF;

  IF m ? 'last_error_code?' THEN
    RAISE EXCEPTION '564: last_error_code? is already mapped — another session has applied this';
  END IF;

  -- The anchor keys 555 asserted, so a restructured mapping aborts here.
  IF NOT (m ? 'work_item_id' AND m ? 'spec' AND m ? 'site_id') THEN
    RAISE EXCEPTION '564: the mapping is missing its anchor keys (work_item_id/spec/site_id) — this is not the mapping this migration was written against';
  END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,last_error_code?}',
      '"current_item.last_error_code"'::jsonb,
      true)
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

  IF NOT (m ? 'last_error_code?') THEN
    RAISE EXCEPTION '564 VERIFY: last_error_code? was not written';
  END IF;
  IF m->>'last_error_code?' <> 'current_item.last_error_code' THEN
    RAISE EXCEPTION '564 VERIFY: last_error_code? points at %, not current_item.last_error_code', m->>'last_error_code?';
  END IF;

  -- Nothing pre-existing may have been dropped. An optional key resolves
  -- silently, so a lost key here would never announce itself.
  IF NOT (m ? 'last_error?' AND m ? 'work_item_id' AND m ? 'spec'
          AND m ? 'site_id' AND m ? 'item_type' AND m ? 'domain') THEN
    RAISE EXCEPTION '564 VERIFY: a pre-existing mapping key was lost';
  END IF;

  RAISE NOTICE '564 OK: call_handler input_mapping now has % keys, last_error_code? -> current_item.last_error_code',
    (SELECT count(*) FROM jsonb_object_keys(m));
END
$verify$;

COMMIT;

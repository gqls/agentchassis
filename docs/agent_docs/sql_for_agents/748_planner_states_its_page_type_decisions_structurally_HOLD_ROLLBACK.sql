-- 748_..._ROLLBACK.sql — undo 748 by the exact inverse replacement.
--
-- Byte-exact round trip, rehearsed against the live row before either file was
-- committed: apply → this → md5 of the template unchanged. It removes ONLY the
-- two strings 748 added, so a neighbour lane's later edit to any other part of
-- the prompt survives an unwind of this one.
--
-- ⚠ If a neighbour lane has since edited the JSON output block or 687's
-- sentence, these anchors will not match verbatim and this file REFUSES rather
-- than guessing — that is correct, not broken. Recover from the snapshot
-- snapshot_agent() took at apply time instead.

BEGIN;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  added_a text := $A748$
  "page_type_decisions": [
    {"page_type": "one of the recommended_page_types you did NOT include in pages", "decision": "deferred|rejected|substituted", "reason": "the real, specific reason for THIS type", "deferred_to": "the name of the mechanism you expect to supply it later, or omit this key if there is none"}
  ],$A748$;
  added_b text := $B748$ Record the same decisions in the `page_type_decisions` array as well, one entry per omitted type: prose is for a human, that array is what the build system reads. If your reason is that something else will supply the pages later, name that mechanism in `deferred_to` — it is checked against whether that mechanism is actually running, so a deferral to something dormant is caught rather than believed. Do NOT use these entries to claim a type is present: whether a type is in the plan is read from `pages`, never from what you say about it.$B748$;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '748 ROLLBACK: prompt_template not found'; END IF;

  n := (length(tpl) - length(replace(tpl, added_a, ''))) / length(added_a);
  IF n <> 1 THEN RAISE EXCEPTION '748 ROLLBACK: JSON block addition found % times, expected 1 — refusing rather than guessing', n; END IF;
  n := (length(tpl) - length(replace(tpl, added_b, ''))) / length(added_b);
  IF n <> 1 THEN RAISE EXCEPTION '748 ROLLBACK: rule-text addition found % times, expected 1 — refusing rather than guessing', n; END IF;

  newtpl := replace(replace(tpl, added_a, ''), added_b, '');
  IF length(newtpl) <> length(tpl) - length(added_a) - length(added_b) THEN
    RAISE EXCEPTION '748 ROLLBACK: unexpected length delta';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,plan_site,config,prompt_template}', to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '748 ROLLBACK: updated % rows, expected 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,plan_site,config,prompt_template}' INTO tpl
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('page_type_decisions' in tpl) > 0 THEN
    RAISE EXCEPTION '748 ROLLBACK VERIFY: page_type_decisions still present';
  END IF;
  IF position('name that exact page_type in strategy_notes' in tpl) = 0 THEN
    RAISE EXCEPTION '748 ROLLBACK VERIFY: migration 687 damaged by the unwind';
  END IF;
  IF position('NO LATER EDITORIAL PASS RUNS' in tpl) = 0 THEN
    RAISE EXCEPTION '748 ROLLBACK VERIFY: migrations 730/731 damaged by the unwind';
  END IF;
  RAISE NOTICE '748 ROLLBACK OK: the addition is gone and the neighbours are intact.';
END $$;

COMMIT;

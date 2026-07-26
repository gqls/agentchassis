-- 211_tool_crosslink_emit_at_build_ROLLBACK.sql — run by hand, deliberately.
-- NOT a migration: the runner's SIDECAR_RE excludes UPPERCASE-suffixed files
-- (bugs_open/007), so this is never auto-applied.
--
-- READ THIS BEFORE RUNNING IT.
--
-- Rolling 211 back re-arms bugs_open/029: it restores tool-suggester's
-- create_cross_links step, which — on any image WITHOUT the emitToolCrossLink
-- Items rewrite — constructs `/tools/{function}.html` and hands it to the
-- content writer as both an instruction and an acceptance test. That produced
-- 27 dead links across 4 sites and an owner-visible 404.
--
-- On an image WITH the rewrite the restored step is harmless (the action
-- resolves a real page and emits nothing when there is none), so the only
-- scenario where this rollback is both meaningful and safe is: the new binary
-- is live and the BUILD-path emit itself needs disabling. In that case prefer
-- reverting the Go change — this file cannot do that.
--
-- TWO TRAPS this file is written around (both cost time on 2026-07-26; RUNBOOK R9):
--   1. snapshot_agent() writes to `agent_definitions_backup`, NOT to
--      agent_definitions with is_snapshot=true. Look in the wrong table and it
--      reads as "no safety net was ever taken".
--   2. 211 was applied TWICE, 7 seconds apart, so there are TWO snapshot sets.
--      The EARLIEST is the pre-state. Taking the newest would "roll back" to
--      the migrated state — a no-op that looks like a successful rollback.
--      Hence min(snapshot_taken_at) below, and the create_cross_links guard.

BEGIN;

DO $$
DECLARE
  t text;
  pre_ts timestamptz;
  restored int;
BEGIN
  SELECT min(snapshot_taken_at) INTO pre_ts
  FROM agent_definitions_backup
  WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%';

  IF pre_ts IS NULL THEN
    RAISE EXCEPTION '211 ROLLBACK: no 211 snapshots in agent_definitions_backup — refusing to guess a pre-state';
  END IF;

  -- Needle gate, COUNTED, before touching anything: the pre-state set must be
  -- exactly the three agents, and must actually carry the needle.
  IF (SELECT count(*) FROM agent_definitions_backup
      WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%'
        AND snapshot_taken_at = pre_ts) <> 3 THEN
    RAISE EXCEPTION '211 ROLLBACK: expected 3 rows in the earliest 211 snapshot set, found %',
      (SELECT count(*) FROM agent_definitions_backup
       WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%'
         AND snapshot_taken_at = pre_ts);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM agent_definitions_backup
    WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%'
      AND snapshot_taken_at = pre_ts
      AND type = 'tool-suggester'
      AND default_config #> '{workflow,steps,create_cross_links}' IS NOT NULL
  ) THEN
    RAISE EXCEPTION '211 ROLLBACK: the earliest snapshot set does NOT contain create_cross_links — it is not the pre-state, restoring it would be a no-op dressed as a rollback';
  END IF;

  FOREACH t IN ARRAY ARRAY['tool-suggester','tool-deployer','tool-generator'] LOOP
    UPDATE agent_definitions live
    SET default_config = snap.default_config, updated_at = NOW()
    FROM (
      SELECT default_config FROM agent_definitions_backup
      WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%'
        AND snapshot_taken_at = pre_ts AND type = t
      LIMIT 1
    ) snap
    WHERE live.type = t AND live.is_active AND COALESCE(live.is_snapshot,false) = false;

    GET DIAGNOSTICS restored = ROW_COUNT;
    IF restored <> 1 THEN
      RAISE EXCEPTION '211 ROLLBACK: restored % live rows for % (expected exactly 1)', restored, t;
    END IF;
  END LOOP;
END $$;

-- Post-condition: the pre-state is genuinely back.
DO $$
DECLARE
  sug jsonb;
  dep jsonb;
BEGIN
  SELECT default_config INTO sug FROM agent_definitions
  WHERE type='tool-suggester' AND is_active AND COALESCE(is_snapshot,false)=false;
  SELECT default_config INTO dep FROM agent_definitions
  WHERE type='tool-deployer' AND is_active AND COALESCE(is_snapshot,false)=false;

  IF sug #> '{workflow,steps,create_cross_links}' IS NULL THEN
    RAISE EXCEPTION '211 ROLLBACK GUARD: create_cross_links did not come back';
  END IF;
  IF sug #>> '{workflow,steps,create_items_loop,next_step}' <> 'create_cross_links' THEN
    RAISE EXCEPTION '211 ROLLBACK GUARD: create_items_loop still points at %',
      sug #>> '{workflow,steps,create_items_loop,next_step}';
  END IF;
  IF dep #>> '{workflow,steps,deploy_tool,config,related_pages}' IS NOT NULL THEN
    RAISE EXCEPTION '211 ROLLBACK GUARD: tool-deployer still carries related_pages';
  END IF;
END $$;

COMMIT;

-- The ledger cannot be un-recorded by the runner. If you intend to re-apply
-- 211 later, delete its row by hand:
--   DELETE FROM schema_migrations WHERE filename = '211_tool_crosslink_emit_at_build.sql';

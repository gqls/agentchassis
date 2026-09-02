-- 674_d4_stage_b_enable_selector_and_flags_HOLD.sql — D4 stage B, the CONFIG half.
-- Council corr 8f4bb57d (submitted with the Go half, commit dec5ad61b).
--
-- ⚠ HELD (_HOLD) BEHIND THE IMAGE ROLL — the honour_site_lock pairing discipline
-- (load_work_item_actions.go:716-722): the flags this file sets are read by Go that must be
-- in the running chassis first. Strictly, with governor_config.enabled=false every piece here
-- is inert in ANY order — the hold is one-enable-event discipline, not a live dependency.
-- APPLY PROCEDURE (BY HAND, in this order — the r3 advisories are steps 3–4):
--   1. `git merge-base --is-ancestor dec5ad61b <the pod's build stamp>` on BOTH chassis
--      replicas (the Go call sites must be in the running binary).
--   2. psql -f this file; read the NOTICE.
--   3. ⚠ LOCKSTEP (council 8f4bb57d r3, guardian): the 657 ordering-contract VERIFY pins the
--      selector md5 (d29807313...) — THIS APPLY CHANGES IT. In the same sitting: add the new
--      md5 to 657_selector_ranks_sites_by_loadable_work_VERIFY.sql's accepted list, run it
--      green, commit both files together. Skipping this makes the daily contract check read
--      as selector drift.
--   4. CANARY (r3, editquality): the LIMIT-0 probe proves the text parses, not that the live
--      steps behave — watch the first ~10 fires (loops load, claims proceed, no
--      spend_governor_shed refusals while the governor is disabled) before walking away.
--   5. Drop this suffix and --record-only in the same motion (bugs_closed/150).
--
-- ⚠ SECOND-CONSUMER GATE (r3, architecture — a CONDITION, not a risk note): these flags
-- expose the governor to build-dispatch-loop ONLY. Any other consumer opting in
-- (site-work-orchestrator, diagnose-dispatch-loop, or anything new) OWES ITS OWN
-- ARCHITECTURE-REVIEW ROUND on the shared governor_state/shed_level machine first —
-- treat a second honour_spend_governor flag as architecture-scope, never routine config.
-- Same condition recorded in AGOV-013.
--
-- WHAT IT DOES (two rows, three writes):
--   1. Teaches `find_dispatchable_site` (build-pipeline-trigger, the post-657 text,
--      md5 d29807313a8f6ed543a541c35c1626c4) the spend-governor clause — a ONE-LINE call to
--      `governor_admits(wi.item_type)` (migration 675, the single canonical predicate; the
--      8f4bb57d r1 architecture revision — Go emits the identical call, so the cross-media
--      lockstep is structural, not string-compared). Without this the loader's filter alone
--      would re-create the bugs_closed/413 selection hog under shed.
--   2. Sets honour_spend_governor: true (bare jsonb boolean) on the dispatch loop's
--      load_items step and its process_item sub-workflow claim step.
-- STILL INERT AFTER APPLY: governor_config.enabled=false and monthly_budget_usd NULL remain
-- the master switches (asserted below). Rollback: 674_..._ROLLBACK.sql (flag DELETION is
-- correct HERE — an absent flag reads false in Go's `.(bool)` — unlike 658's knobs, where
-- deletion meant Go defaults 50/20).

BEGIN;

-- Refusal FIRST (the replay-decoy ordering), then snapshots, then writes.
DO $$
DECLARE m text; q text; n int;
BEGIN
  SELECT md5(default_config#>>'{workflow,steps,find_dispatchable_site,config,query}'),
         default_config#>>'{workflow,steps,find_dispatchable_site,config,query}'
    INTO m, q
  FROM agent_definitions
  WHERE type='build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false
    AND deleted_at IS NULL
  ORDER BY version DESC LIMIT 1;
  IF m IS NULL THEN RAISE EXCEPTION '674 REFUSED: no live build-pipeline-trigger row.'; END IF;
  IF position('governor_admits' in q) > 0 THEN
    RAISE EXCEPTION '674 REFUSED: selector already carries the governor clause — already applied (replay).';
  END IF;
  IF m <> 'd29807313a8f6ed543a541c35c1626c4' THEN
    RAISE EXCEPTION '674 REFUSED: selector md5 % is not the post-657 text — drifted, investigate before overwriting.', m;
  END IF;
  -- Belt on the md5's braces (council 8f4bb57d r2, debug_historian): assert the replace()
  -- anchor appears EXACTLY ONCE before replacing. The md5-exact precondition above already
  -- pins every byte of the text, so a failure here means THIS GUARD is stale, not the row —
  -- but a replace() that could silently no-op never gets to rely on that reasoning.
  IF (length(q) - length(replace(q,
        'AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'')', '')))
     / length('AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'')') <> 1 THEN
    RAISE EXCEPTION '674 REFUSED: busy-skip anchor not present exactly once — the replace() would no-op or double-fire.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname='governor_admits') THEN
    RAISE EXCEPTION '674 REFUSED: governor_admits() missing — apply 675 first (the flags call it).';
  END IF;

  SELECT count(*) INTO n FROM agent_definitions WHERE type='build-dispatch-loop';
  IF n <> 1 THEN RAISE EXCEPTION '674 REFUSED: build-dispatch-loop has % rows, expected exactly 1.', n; END IF;
  PERFORM 1 FROM agent_definitions WHERE type='build-dispatch-loop'
    AND (default_config#>'{workflow,steps,load_items,config,honour_spend_governor}') IS NOT NULL;
  IF FOUND THEN RAISE EXCEPTION '674 REFUSED: load_items flag already present (replay/drift).'; END IF;

  PERFORM 1 FROM governor_config WHERE id=1 AND enabled=false;
  IF NOT FOUND THEN
    RAISE EXCEPTION '674 REFUSED: governor_config.enabled is not false — this file must land on a disabled governor (one deliberate enable event, later).';
  END IF;
END $$;

SELECT snapshot_agent('build-pipeline-trigger', '674_d4_stage_b_enable_selector_and_flags_HOLD.sql: pre-update');
SELECT snapshot_agent('build-dispatch-loop', '674_d4_stage_b_enable_selector_and_flags_HOLD.sql: pre-update');

-- 1. The selector clause, inserted before the busy-skip. Token-identical to the Go renderer.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
  '{workflow,steps,find_dispatchable_site,config,query}',
  to_jsonb(replace(
    default_config#>>'{workflow,steps,find_dispatchable_site,config,query}',
    'AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'')',
    'AND governor_admits(wi.item_type) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = ''claimed'')'
  ))
)
WHERE type='build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL;

-- 2. The two step flags, bare jsonb booleans.
UPDATE agent_definitions
SET default_config = jsonb_set(jsonb_set(default_config,
  '{workflow,steps,load_items,config,honour_spend_governor}', 'true'::jsonb),
  '{workflow,steps,process_item,config,sub_workflow,steps,claim,config,honour_spend_governor}', 'true'::jsonb)
WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL;

-- Verify block (DO/RAISE — SELECT-only cannot stop the COMMIT).
DO $$
DECLARE q text; n int;
BEGIN
  SELECT default_config#>>'{workflow,steps,find_dispatchable_site,config,query}' INTO q
  FROM agent_definitions
  WHERE type='build-pipeline-trigger' AND is_active AND COALESCE(is_snapshot,false)=false
    AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;

  -- The clause landed exactly once, and the busy-skip survived it.
  n := (length(q) - length(replace(q, 'governor_admits', ''))) / length('governor_admits');
  IF n <> 1 THEN RAISE EXCEPTION '674 VERIFY: governor clause appears % times in the selector, expected exactly 1', n; END IF;
  IF position('active.status = ''claimed''' in q) = 0 THEN
    RAISE EXCEPTION '674 VERIFY: the busy-skip clause is gone — the replace anchor mis-fired';
  END IF;

  -- The selector must still RUN (constants and joins all resolvable).
  EXECUTE 'SELECT 1 FROM (' || q || ') probe LIMIT 0';

  -- Both flags are bare booleans reading true.
  PERFORM 1 FROM agent_definitions
  WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND jsonb_typeof(default_config#>'{workflow,steps,load_items,config,honour_spend_governor}')='boolean'
    AND (default_config#>>'{workflow,steps,load_items,config,honour_spend_governor}')::bool
    AND jsonb_typeof(default_config#>'{workflow,steps,process_item,config,sub_workflow,steps,claim,config,honour_spend_governor}')='boolean'
    AND (default_config#>>'{workflow,steps,process_item,config,sub_workflow,steps,claim,config,honour_spend_governor}')::bool;
  IF NOT FOUND THEN RAISE EXCEPTION '674 VERIFY: the two step flags are not both bare jsonb true'; END IF;

  -- The master switch is still off: nothing enforces until the owner's deliberate enable.
  PERFORM 1 FROM governor_config WHERE id=1 AND enabled=false;
  IF NOT FOUND THEN RAISE EXCEPTION '674 VERIFY: governor unexpectedly enabled during apply'; END IF;

  -- Whole-fleet negative control: no OTHER live agent row mentions the flag or the map.
  SELECT count(*) INTO n FROM agent_definitions
  WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
    AND type NOT IN ('build-pipeline-trigger','build-dispatch-loop')
    AND default_config::text LIKE '%honour_spend_governor%';
  IF n <> 0 THEN RAISE EXCEPTION '674 VERIFY: % unexpected agent rows carry the flag', n; END IF;

  RAISE NOTICE '674 OK: selector carries the governor clause (token-lockstep with Go), both flags true, governor still DISABLED — enforcement awaits the owner''s budget + enable.';
END $$;

COMMIT;

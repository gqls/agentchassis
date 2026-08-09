-- ============================================================================
-- 349_bugfix_136_residue_repairs.sql
--
-- bugs_open/136 — the deferred data-side residue, cleared on the owner's
-- 2026-08-09 instruction ("we can fix those deferred items now"; handoff
-- HANDOFF_2026-08-09_deferred_items.md items A, B, C).
--
-- The framework fix (SCR-006 alias seam) is live and runtime-witnessed
-- (bug §11, v1.0.1268). This migration clears the data residue it left:
--
--   B. page-build-handler plan_sections.domain — the last UNKNOWN key in
--      ./scripts/audit-config-keys.sh. Value "site_record.domain" was a
--      dot-path aimed at a spec the action does not have: `domain` is not in
--      Required ∪ Optional (spec has `pipeline`, which nothing sets and the
--      action body never reads — there is no inputs.Get("pipeline") in
--      plan_sections_action.go), so Strategy 0 never resolves it. Fetched by
--      nobody, used by nothing (bug §2c, re-confirmed 2026-08-09).
--      DELETED, not renamed: renaming to `pipeline` would wire a dot-path
--      string into a live input nothing consumes — §2's own mistake inverted.
--
--   C. The three deprecated key spellings renamed to what the code reads,
--      in every live active definition that still carries them (19 steps,
--      15 agents, measured 2026-08-09):
--        item_domain   -> item_pipeline    (create_work_item steps, x15)
--        check_domain  -> check_pipeline   (run_discovery_checks steps, x3)
--        target_domain -> target_pipeline  (triage_detected_items steps, x1)
--
--      NINETEEN, not the 13 this lane's handoff and RUNBOOK say. SIX live
--      inside a loop step's `sub_workflow.steps` — component-quality-auditor,
--      internal-linker, tool-auditor (x2), tool-suggester (x2) — and the
--      RUNBOOK's census SQL walks `->'workflow'->'steps'` only, so it cannot
--      see them. That hand-written descent is precisely the shape
--      validation.WalkSteps was extracted to abolish (bugs_open/144: two
--      traversals blind in the same direction, agreeing with each other).
--      The AUDIT is not blind — cmd/config-key-audit walks WalkSteps at every
--      call site — so the acceptance number stays trustworthy; it was the
--      lane's own census query that undercounted, by 32%. The rename below is
--      therefore driven by a depth-recursive walk, not by a hand list, so it
--      cannot inherit that blindness.
--
--      Behaviour-preserving BY CONSTRUCTION: the live binary honours both
--      names via DeprecatedConfigKeys/ResolveConfigSetting and the new name
--      wins when both are set, so every intermediate state is correct — a
--      partially-applied run cannot be wrong, only incomplete. The value
--      transfers verbatim (read from the config, never hardcoded). The
--      aliases stay DECLARED in the specs: they are the net for snapshot
--      replays and stragglers. This drains the audit's DEPRECATED in-use
--      count to 0; it does not remove the net. Seeds carrying the old
--      spellings are corrected in the same commit (the bugs_open/134 lesson)
--      — EXCEPT 051_build_dispatch_loop.sql, deliberately left alone because
--      052 deletes those very keys by name (`? 'item_domain'`) and renaming
--      the seed would break that removal on replay, silently restoring a
--      dispatch filter 052 removed as a defect.
--
--   A. The three mislabelled site_work_items rows repaired design -> content.
--      Filed by completeness-discovery-agent (configured `content`) while the
--      run_discovery_checks shim was missing, so dctx fell back to "design"
--      (bug §2a-update). §6 proved NO live consumer distinguishes design from
--      content today — this is a record correction, keyed on what was
--      measured so a row anything else has since changed is left alone.
--      NOTE: 74bb48ff (capability_gap, `detected`) stays undispatched after
--      the repair — `detected` is a dead queue (bugs_open/083, open); routing
--      its finding is an owner call on 083's queue, not part of this fix.
--      (The bug file says four rows; one capability_gap vanished from the DB
--      before 2026-08-09 — not repaired, simply gone. Not investigated.)
--
-- Live immediately (DB config + data; no image roll involved).
--
-- Verify after apply:
--   ./scripts/audit-config-keys.sh          -- UNKNOWN KEYS: none
--   -- and the DEPRECATED section's keys carried by nothing live:
--   SELECT count(*) FROM agent_definitions
--    WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
--      AND default_config::text ~ '(item|check|target)_domain';   -- 0
--   SELECT pipeline FROM site_work_items
--    WHERE created_by='completeness-discovery-agent';             -- no 'design'
-- ============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- The descent. Temporary (pg_temp) so it cannot outlive this session or
-- collide with another thread's work. Walks objects AND arrays to any depth,
-- returning one row per occurrence of an old-spelling key with the full jsonb
-- path to it and the value to carry across.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION pg_temp.bugfix136_old_key_paths()
RETURNS TABLE(def_id uuid, agent_type text, key_path text[], old_key text,
              new_key text, val jsonb)
LANGUAGE sql STABLE AS $fn$
  WITH RECURSIVE walk(def_id, agent_type, key_path, node) AS (
    SELECT ad.id, ad.type, ARRAY[]::text[], ad.default_config
      FROM agent_definitions ad
     WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false
       AND ad.is_active
       AND ad.default_config::text ~ '(item|check|target)_domain'
    UNION ALL
    SELECT w.def_id, w.agent_type, w.key_path || e.k, e.v
      FROM walk w
      CROSS JOIN LATERAL jsonb_each(
        CASE jsonb_typeof(w.node)
          WHEN 'object' THEN w.node
          WHEN 'array'  THEN COALESCE((SELECT jsonb_object_agg((i-1)::text, v)
                                         FROM jsonb_array_elements(w.node)
                                              WITH ORDINALITY a(v,i)), '{}'::jsonb)
          ELSE '{}'::jsonb
        END) AS e(k,v)
  )
  SELECT def_id, agent_type, key_path,
         key_path[array_length(key_path,1)],
         CASE key_path[array_length(key_path,1)]
           WHEN 'item_domain'   THEN 'item_pipeline'
           WHEN 'check_domain'  THEN 'check_pipeline'
           WHEN 'target_domain' THEN 'target_pipeline'
         END,
         node
    FROM walk
   WHERE key_path[array_length(key_path,1)]
         IN ('item_domain','check_domain','target_domain');
$fn$;

-- Materialise the work list BEFORE mutating anything, so the guard, the
-- rename and the report all describe the same set.
CREATE TEMP TABLE bugfix136_renames ON COMMIT DROP AS
  SELECT * FROM pg_temp.bugfix136_old_key_paths();

-- ---------------------------------------------------------------------------
-- Guard: distinguish "already applied" (probe-detectable) from DRIFT.
-- Composed against live state measured 2026-08-09 ~12:30Z. All three clean
-- => this file has already run. A partial match => another session has moved
-- the config; stop rather than re-shaping their work.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  old_key_steps int;
  plan_domain   int;
  bad_rows      int;
BEGIN
  SELECT count(*) INTO old_key_steps FROM bugfix136_renames;

  SELECT count(*) INTO plan_domain
    FROM agent_definitions
   WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'plan_sections'->'config' ? 'domain';

  SELECT count(*) INTO bad_rows
    FROM site_work_items
   WHERE created_by='completeness-discovery-agent' AND pipeline='design';

  IF old_key_steps = 0 AND plan_domain = 0 AND bad_rows = 0 THEN
    RAISE EXCEPTION '349: already applied (no old-spelling keys, no plan_sections.domain, no mislabelled rows)';
  END IF;

  IF old_key_steps <> 19 OR plan_domain <> 1 OR bad_rows <> 3 THEN
    RAISE EXCEPTION '349: DRIFT — expected 19 old-key steps / 1 plan_sections.domain / 3 mislabelled rows, found % / % / %. Re-measure before applying.',
      old_key_steps, plan_domain, bad_rows;
  END IF;

  -- Every occurrence must map to a known new name; a NULL would mean the
  -- descent matched a key this file does not know how to rename.
  IF EXISTS (SELECT 1 FROM bugfix136_renames WHERE new_key IS NULL) THEN
    RAISE EXCEPTION '349: an old-key occurrence has no mapped new name';
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- B. Delete the dead plan_sections.domain key (surgical; snapshots untouched)
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,plan_sections,config,domain}',
    updated_at = now()
WHERE type='page-build-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'plan_sections'->'config' ? 'domain';

-- ---------------------------------------------------------------------------
-- C. Rename each occurrence at its own path. One UPDATE per occurrence, NOT a
-- set-based join: five definitions carry more than one (improvement-loop x3,
-- tool-auditor x2, tool-suggester x2), and a join would apply only one SET per
-- target row, dropping the rest silently.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  r   record;
  hit int;
  n   int := 0;
BEGIN
  FOR r IN SELECT * FROM bugfix136_renames ORDER BY agent_type, key_path LOOP
    UPDATE agent_definitions
       SET default_config = jsonb_set(
             default_config,
             r.key_path[1:array_length(r.key_path,1)-1] || r.new_key,
             r.val
           ) #- r.key_path,
           updated_at = now()
     WHERE id = r.def_id;
    GET DIAGNOSTICS hit = ROW_COUNT;
    IF hit <> 1 THEN
      RAISE EXCEPTION '349: rename %.% (% -> %) touched % rows, expected 1',
        r.agent_type, array_to_string(r.key_path,'.'), r.old_key, r.new_key, hit;
    END IF;
    n := n + 1;
  END LOOP;
  RAISE NOTICE '349: renamed % deprecated key occurrences', n;
END $$;

-- ---------------------------------------------------------------------------
-- A. Repair the mislabelled rows (keyed on what was measured)
-- ---------------------------------------------------------------------------
UPDATE site_work_items
   SET pipeline='content', updated_at=now()
 WHERE created_by='completeness-discovery-agent' AND pipeline='design';

-- ---------------------------------------------------------------------------
-- Verify: a bare SELECT cannot stop the COMMIT (ON_ERROR_STOP ignores a
-- non-empty result) — DO/RAISE, end-state conditions only.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  leftovers   int;
  new_keys    int;
  plan_domain int;
  bad_rows    int;
BEGIN
  -- re-walk from scratch, not from the materialised list
  SELECT count(*) INTO leftovers FROM pg_temp.bugfix136_old_key_paths();
  IF leftovers <> 0 THEN
    RAISE EXCEPTION '349 VERIFY: % live occurrences still carry an old-spelling key', leftovers;
  END IF;

  -- 19 renamed here + 2 that already used the new spelling (claims-auditor
  -- request_claims_review, site-work-orchestrator load_work_items)
  SELECT count(*) INTO new_keys
    FROM agent_definitions
   WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
     AND default_config::text ~ '(item|check|target)_pipeline';
  IF new_keys < 16 THEN
    RAISE EXCEPTION '349 VERIFY: expected >=16 definitions carrying a new-spelling key, found %', new_keys;
  END IF;

  SELECT count(*) INTO plan_domain
    FROM agent_definitions
   WHERE type='page-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'plan_sections'->'config' ? 'domain';
  IF plan_domain <> 0 THEN
    RAISE EXCEPTION '349 VERIFY: plan_sections still carries the dead domain key';
  END IF;

  SELECT count(*) INTO bad_rows
    FROM site_work_items
   WHERE created_by='completeness-discovery-agent' AND pipeline='design';
  IF bad_rows <> 0 THEN
    RAISE EXCEPTION '349 VERIFY: % mislabelled rows remain', bad_rows;
  END IF;
END $$;

COMMIT;

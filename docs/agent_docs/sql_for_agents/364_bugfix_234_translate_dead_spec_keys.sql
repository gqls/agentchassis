-- ============================================================================
-- 364_bugfix_234_translate_dead_spec_keys.sql
--
-- bugs_open/234, the DATA half. Three live create_work_item steps carry a
-- config key `spec` that NO version of the action has ever read — it builds
-- the item spec from spec_data / spec_paths / spec_literal
-- (create_work_item_action.go:247-296), and no ExtractActionInputs strategy
-- can resolve a key that is in none of Required ∪ Optional, input_fields, or
-- Deprecated. Every item these steps file therefore carries spec = '{}'::jsonb
-- (16/16 improvement_rerender_* rows, measured 2026-08-09; positive control:
-- 5,040 rows fleet-wide with a non-empty spec). bugs_closed/024 introduced the
-- correct spellings (migration 180); these steps never migrated onto them.
--
-- OWNER DECISION 2026-08-09/10 (recorded in the case file): RESTORE the
-- improvement-loop flag rather than delete it. Grounds re-verified at decision
-- time: (a) bugs_open/226's chrome-divergence guard is LIVE in production, so
-- a component rebuild that would destroy hand-patched chrome now files a
-- recoverable record; (b) refresh_site_components=true is filed daily by 8
-- other producers (~5-15 rows/day) — restoring this path returns one lost
-- producer to the fleet norm, it does not switch on a dormant behaviour.
--
-- The three translations — each a RENAME of the key, the VALUE copied verbatim
-- from the live row (config->'spec'), never retyped here, so no transcription
-- of the prose value can drift it:
--
--   1. improvement-loop.insert_rerender_item
--        spec -> spec_literal   {"refresh_site_components": true}
--      BEHAVIOUR-RESTORING: the next improvement_rerender_* item carries the
--      flag, build-dispatch-loop's input_mapping
--      ("pending.first_item.spec.refresh_site_components",
--      051_build_dispatch_loop.sql:823) finally finds it, and rerender-pages'
--      conditional (033_rerender_pages_action.sql:1107) takes the
--      site-component-refresh arm for the first time from this path.
--
--   2. improvement-loop.record_not_converging
--        spec -> spec_literal   {"reason": <prose>, "capability": "audit_not_converging"}
--      Both values are constants. Step has NEVER filed a row
--      (capability_gap_audit% = 0), so nothing live changes shape.
--
--   3. deduplicate-sections.queue_rerender
--        spec -> spec_paths     {"page_id": "input_data.page_id"}
--      The value is a PATH, so spec_paths (per-key resolution) is the correct
--      spelling, NOT spec_literal — a literal would stamp the string
--      "input_data.page_id" into the spec. spec_paths hard-errors if the path
--      does not resolve; that adds no new failure mode, because the SAME step
--      already hard-errors on the SAME path via item_key_suffix_field
--      (create_work_item_action.go:207). Step has never filed a row
--      (dedupe_rerender% = 0).
--
-- NOT touched, deliberately: the 16 historic improvement_rerender_* rows and
-- the currently-triaged improvement_rerender_dartsonline.com row keep their
-- empty specs — they are the damage record, and the triaged row will dispatch
-- with today's (flagless) behaviour, which is what it was filed under.
--
-- Verification is at a FILED ROW, never at the definition (a config that LOOKS
-- right is exactly what this bug is):
--   SELECT item_key, spec, created_at FROM site_work_items
--    WHERE created_by='improvement-loop' AND item_key LIKE 'improvement_rerender%'
--    ORDER BY created_at DESC LIMIT 3;
--   -- the first row CREATED AFTER this migration must carry
--   -- {"refresh_site_components": true}. ~1.8 rows/day natural rate.
--
-- SEEDS corrected in the same commit (the bugs_open/134 lesson — a reseed
-- would replay the dead key): 054_improvement_loop.sql (insert_rerender_item),
-- 291_improvement_loop_convergence_gate_replaces_pass_cap.sql
-- (record_not_converging), 269_deduplicate_sections_handler.sql
-- (queue_rerender). ROLLBACK files are left alone: they reproduce prior state
-- by design.
--
-- The CODE half (bugs_open/234 class fix: ActionInputSpec.RemovedConfigKeys +
-- StrictConfig on create_work_item) ships separately and is inert until an
-- image rolls. THIS FILE MUST BE APPLIED BEFORE THAT CODE IS COMMITTED — on
-- this tree committing is shipping, and the code turns a live `spec` carrier
-- into a hard validation error on every message.
--
-- Live immediately (DB config; no image roll involved).
-- ============================================================================

BEGIN;

-- Pre-guard: exactly the three expected carriers, with exactly the expected
-- values, and none already carrying its target key. Anything else is drift —
-- re-measure before applying. (If the exact-equality checks fail while the
-- key-presence counts pass, suspect the prose value changed under us OR this
-- file's own encoding was mangled in transit; both are reasons to stop.)
DO $$
DECLARE
  il_cfg  jsonb;  -- insert_rerender_item
  nc_cfg  jsonb;  -- record_not_converging
  qr_cfg  jsonb;  -- queue_rerender
BEGIN
  SELECT default_config->'workflow'->'steps'->'insert_rerender_item'->'config',
         default_config->'workflow'->'steps'->'record_not_converging'->'config'
    INTO il_cfg, nc_cfg
    FROM agent_definitions
   WHERE type='improvement-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  SELECT default_config->'workflow'->'steps'->'queue_rerender'->'config'
    INTO qr_cfg
    FROM agent_definitions
   WHERE type='deduplicate-sections' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF il_cfg IS NULL OR nc_cfg IS NULL OR qr_cfg IS NULL THEN
    RAISE EXCEPTION '364: a target step is missing (il=% nc=% qr=%)',
      il_cfg IS NOT NULL, nc_cfg IS NOT NULL, qr_cfg IS NOT NULL;
  END IF;

  IF NOT (il_cfg ? 'spec') AND NOT (nc_cfg ? 'spec') AND NOT (qr_cfg ? 'spec') THEN
    RAISE EXCEPTION '364: already applied (no step carries `spec`)';
  END IF;
  IF NOT (il_cfg ? 'spec' AND nc_cfg ? 'spec' AND qr_cfg ? 'spec') THEN
    RAISE EXCEPTION '364: DRIFT — partial application or concurrent edit (spec on il=% nc=% qr=%). Re-measure.',
      il_cfg ? 'spec', nc_cfg ? 'spec', qr_cfg ? 'spec';
  END IF;
  IF il_cfg ? 'spec_literal' OR nc_cfg ? 'spec_literal' OR qr_cfg ? 'spec_paths' THEN
    RAISE EXCEPTION '364: DRIFT — a target key already exists alongside `spec`. Re-measure.';
  END IF;

  IF il_cfg->'spec' <> '{"refresh_site_components": true}'::jsonb THEN
    RAISE EXCEPTION '364: DRIFT — insert_rerender_item.spec is %, not the expected flag', il_cfg->'spec';
  END IF;
  IF qr_cfg->'spec' <> '{"page_id": "input_data.page_id"}'::jsonb THEN
    RAISE EXCEPTION '364: DRIFT — queue_rerender.spec is %, not the expected path map', qr_cfg->'spec';
  END IF;
  -- record_not_converging: assert the shape and the machine-read value exactly;
  -- assert the prose value by its stable ends rather than retyping it here.
  IF (SELECT count(*) FROM jsonb_object_keys(nc_cfg->'spec')) <> 2
     OR nc_cfg->'spec'->>'capability' IS DISTINCT FROM 'audit_not_converging'
     OR nc_cfg->'spec'->>'reason' NOT LIKE 'three audit passes at an unchanged site fingerprint%'
     OR nc_cfg->'spec'->>'reason' NOT LIKE '%human attention needed' THEN
    RAISE EXCEPTION '364: DRIFT — record_not_converging.spec is %', nc_cfg->'spec';
  END IF;
END $$;

-- The renames. Value copied from the live key; old key removed in the same
-- expression, so there is no intermediate state with both keys.
UPDATE agent_definitions
SET default_config =
      jsonb_set(
        default_config #- '{workflow,steps,insert_rerender_item,config,spec}',
        '{workflow,steps,insert_rerender_item,config,spec_literal}',
        default_config#>'{workflow,steps,insert_rerender_item,config,spec}'),
    updated_at = now()
WHERE type='improvement-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        default_config #- '{workflow,steps,record_not_converging,config,spec}',
        '{workflow,steps,record_not_converging,config,spec_literal}',
        default_config#>'{workflow,steps,record_not_converging,config,spec}'),
    updated_at = now()
WHERE type='improvement-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config =
      jsonb_set(
        default_config #- '{workflow,steps,queue_rerender,config,spec}',
        '{workflow,steps,queue_rerender,config,spec_paths}',
        default_config#>'{workflow,steps,queue_rerender,config,spec}'),
    updated_at = now()
WHERE type='deduplicate-sections' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify: DO/RAISE, not bare SELECTs (ON_ERROR_STOP ignores a non-empty
-- result set). Checks the removal, the arrival WITH the exact value, and the
-- survival of sibling keys — so a whole-config clobber cannot pass as a
-- surgical rename. Proven disconfirmable by mutation before applying (skip one
-- UPDATE in a scratch copy → the matching RAISE fires); the proof is recorded
-- in the lane's NOTES.
DO $$
DECLARE
  il_cfg jsonb;
  nc_cfg jsonb;
  qr_cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'insert_rerender_item'->'config',
         default_config->'workflow'->'steps'->'record_not_converging'->'config'
    INTO il_cfg, nc_cfg
    FROM agent_definitions
   WHERE type='improvement-loop' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  SELECT default_config->'workflow'->'steps'->'queue_rerender'->'config'
    INTO qr_cfg
    FROM agent_definitions
   WHERE type='deduplicate-sections' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF il_cfg ? 'spec' OR nc_cfg ? 'spec' OR qr_cfg ? 'spec' THEN
    RAISE EXCEPTION '364 VERIFY: a `spec` key survived (il=% nc=% qr=%)',
      il_cfg ? 'spec', nc_cfg ? 'spec', qr_cfg ? 'spec';
  END IF;

  IF il_cfg->'spec_literal' IS DISTINCT FROM '{"refresh_site_components": true}'::jsonb THEN
    RAISE EXCEPTION '364 VERIFY: insert_rerender_item.spec_literal is %', il_cfg->'spec_literal';
  END IF;
  IF nc_cfg->'spec_literal'->>'capability' IS DISTINCT FROM 'audit_not_converging'
     OR (SELECT count(*) FROM jsonb_object_keys(nc_cfg->'spec_literal')) <> 2 THEN
    RAISE EXCEPTION '364 VERIFY: record_not_converging.spec_literal is %', nc_cfg->'spec_literal';
  END IF;
  IF qr_cfg->'spec_paths' IS DISTINCT FROM '{"page_id": "input_data.page_id"}'::jsonb THEN
    RAISE EXCEPTION '364 VERIFY: queue_rerender.spec_paths is %', qr_cfg->'spec_paths';
  END IF;

  -- Sibling survival (whole-config-clobber check, per 350's pattern).
  IF NOT (il_cfg ? 'site_id' AND il_cfg ? 'source' AND il_cfg ? 'item_pipeline'
          AND il_cfg ? 'item_type' AND il_cfg ? 'severity' AND il_cfg ? 'summary'
          AND il_cfg ? 'priority' AND il_cfg ? 'handler_agent' AND il_cfg ? 'item_key_prefix') THEN
    RAISE EXCEPTION '364 VERIFY: insert_rerender_item lost a sibling key: %', il_cfg;
  END IF;
  IF NOT (nc_cfg ? 'source' AND nc_cfg ? 'site_id' AND nc_cfg ? 'summary'
          AND nc_cfg ? 'priority' AND nc_cfg ? 'severity' AND nc_cfg ? 'status'
          AND nc_cfg ? 'item_type' AND nc_cfg ? 'item_pipeline' AND nc_cfg ? 'handler_agent'
          AND nc_cfg ? 'item_key_prefix' AND nc_cfg ? 'recurrence_expected') THEN
    RAISE EXCEPTION '364 VERIFY: record_not_converging lost a sibling key: %', nc_cfg;
  END IF;
  IF NOT (qr_cfg ? 'site_id' AND qr_cfg ? 'page_id' AND qr_cfg ? 'source'
          AND qr_cfg ? 'item_type' AND qr_cfg ? 'item_pipeline' AND qr_cfg ? 'handler_agent'
          AND qr_cfg ? 'severity' AND qr_cfg ? 'priority' AND qr_cfg ? 'summary'
          AND qr_cfg ? 'item_key_prefix' AND qr_cfg ? 'item_key_suffix_field') THEN
    RAISE EXCEPTION '364 VERIFY: queue_rerender lost a sibling key: %', qr_cfg;
  END IF;
END $$;

COMMIT;

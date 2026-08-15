-- 418 — content-gap-planner: requires-backend eligibility gate on the SECTION menu
-- (VMB-010's planner half, section side; bugs_open/276 — corrects the bug's own fix
-- candidate 1, which named build-site-planner as the call site to fix)
--
-- WHY. `content_components` rows can carry semantic tag 'requires-backend' — currently
-- exactly one SECTION-level row, `intent-probe`, whose frontend POSTs to a server-side API.
-- The TOOL-level instance of this tag (`chat-input-box`) was gated in tool-suggester by
-- migration 406 (2026-08-14). bugs_open/276 was filed the same day, at the council's own
-- direction, because the section-level instance has NO equivalent gate anywhere.
--
-- bugs_open/276 and the concept register (VMB-010) both name only build-site-planner's
-- load_components step as the fix target. THAT UNDERCOUNTS THE PROBLEM. A fleet-wide sweep
-- of every agent_definitions step whose query touches content_components (14 hits, all read
-- 2026-08-15) found THREE section-candidate "menu" queries, not one. Dispatch volume,
-- orchestration_states.owner_agent_type, last 30 days (measured 2026-08-15 14:03Z):
--   content-gap-planner.load_available_components   131  (most recent: today)  <- THIS FILE
--   build-site-planner.load_components                 2                       <- 419
--   site-planner.load_available_components              0, ever (unbounded)    <- 420
-- content-gap-planner is not named anywhere in the bug file or the register, and is the
-- overwhelmingly dominant real placement path — 65x build-site-planner's volume. Gating only
-- the named call site would leave the actual live risk wide open.
--
-- Current placements of intent-probe: exactly one, relojistas.com/index, which already
-- carries deploy_config->capabilities ? 'backend' (measured 2026-08-15). No live damage —
-- this is a forward-looking gate, not a page repair.
--
-- SHAPE. Same clause as 406 (RFC_022 shape: opt-in tag, unsafe side defaults OFF, consumers
-- enumerated here — no fresh architecture RFC needed per the 2026-08-02 owner rulings §1/§2,
-- both already satisfied: producer/consumer set named, opt-in field defaults OFF).
--
-- $1 BINDING PROOF (this step currently takes no params at all). content-gap-planner's
-- workflow is strictly linear: ensure_site_record (output_field site_record) -> load_specs
-- (binds site_id: site_record.site_id) -> load_existing_pages (same) ->
-- load_available_components (this step) -> plan_gaps -> apply_plan. Two sibling steps
-- immediately upstream of this one already resolve site_record.site_id live. Measured
-- 2026-08-15: 131/131 orchestration_states rows for this agent in the last 30 days carry a
-- non-null collected_data#>>'{site_record,site_id}'.
--
-- ID-SCOPED + PRE-STATE GATED (per 406's addendum: "FUTURE agent_definitions migrations
-- should scope by id + pre-state gate"), not the bare type-scoped UPDATE 406/407's own
-- bodies used — this exact row has been edited twice by a concurrent session in the hours
-- around this migration's authoring (413, 414: model + plan_gaps template, confirmed NOT to
-- touch load_available_components), which is precisely the collision id-scoping guards
-- against. The pre-state check is a byte-exact match on the known-current query text; a
-- mismatch aborts the whole transaction rather than silently double-applying or clobbering a
-- concurrent edit.
--
-- CONSUMERS of load_available_components's output: the plan_gaps LLM step of
-- content-gap-planner, nothing else (output_field available_components, read only inside
-- this workflow).
--
-- Config-only: no image dependency, live on apply.
--
-- ROLLBACK: 418_content_gap_planner_requires_backend_gate_ROLLBACK.sql, or restore the
-- snapshot this file takes (snapshot_agent note
-- '418_content_gap_planner_requires_backend_gate: pre-update').

BEGIN;

SELECT snapshot_agent('content-gap-planner',
  '418_content_gap_planner_requires_backend_gate: pre-update');

DO $$
DECLARE
  target_id uuid;
  cur_query text;
  expected_pre text := 'SELECT name, display_name, "function", category FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true ORDER BY category, name';
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '418: expected exactly 1 live content-gap-planner row, found %', n;
  END IF;

  SELECT id, default_config#>>'{workflow,steps,load_available_components,config,query}'
    INTO target_id, cur_query
    FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur_query IS DISTINCT FROM expected_pre THEN
    RAISE EXCEPTION '418: load_available_components pre-state does not match expected text (concurrent edit?): %', cur_query;
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(
           jsonb_set(
             default_config,
             '{workflow,steps,load_available_components,config,query}',
             to_jsonb('SELECT name, display_name, "function", category FROM content_components WHERE component_level IN (''section'', ''element'') AND is_active = true AND (NOT (COALESCE(semantic_tags, ''[]''::jsonb) ? ''requires-backend'') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->''capabilities'', ''[]''::jsonb) ? ''backend'')) ORDER BY category, name'::text)
           ),
           '{workflow,steps,load_available_components,config,params}',
           '["site_record.site_id"]'::jsonb
         ),
         updated_at = now()
   WHERE id = target_id;
END $$;

DO $$
DECLARE
  q text;
  p jsonb;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '418: post-update, expected exactly 1 live content-gap-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_available_components,config,query}',
         default_config#>'{workflow,steps,load_available_components,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'content-gap-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL OR position('requires-backend' in q) = 0
     OR position('deploy_config' in q) = 0
     OR position('$1' in q) = 0 THEN
    RAISE EXCEPTION '418: load_available_components query does not carry the gate: %', q;
  END IF;
  IF p IS NULL OR jsonb_array_length(p) <> 1
     OR p->>0 <> 'site_record.site_id' THEN
    RAISE EXCEPTION '418: load_available_components params wrong: %', p;
  END IF;
END $$;

COMMIT;

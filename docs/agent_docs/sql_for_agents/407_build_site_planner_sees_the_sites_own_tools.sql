-- 407 — build-site-planner: the component menu includes the site's OWN tool
-- components, per-site opt-in (loancalculator rebuild, owner decision 3 half 2,
-- PLAN_2026-08-12_planner_sees_locked_tools.md).
--
-- WHY. The planner's load_components menu offers section/element components
-- only, so a planner run on a site whose pages carry locked tool components
-- (12 calculator rows on loancalculator.co.uk) cannot NAME them in a plan —
-- the composition omits the slot, and the locked original is exiled to the
-- page foot with a lock_blocked_change item (the decomposition lane's proven
-- failure shape: "seeding a site plan is the dangerous act"). Half 1
-- (matchLockedRow identity arm, f4820a877, LIVE v1.0.1294+, council a625c326
-- APPROVED) pairs a planned tool section with its locked row by component
-- identity — but stays inert until the planner can name a tool at all. This
-- is half 2.
--
-- SHAPE. Self-gating SQL, two gates in series: the site's current structure
-- spec must carry plan_includes_tools='true' (fifth opt-in key on that
-- aspect, after url_shape + the three identity-policy keys — same idiom,
-- unsafe default OFF, absent key = today's behaviour), AND the tool must
-- already be placed on one of the planning site's own pages
-- (content_components is GLOBAL — 81 active tools; without the placement
-- gate a flagged site would be offered other sites' calculators).
--
-- MEASURED before apply (2026-08-14, live DB, control_run.sql in the
-- loancalculator lane's scratch, recorded in NOTES):
--   * old vs new query md5-identical over the full ordered row set
--     (0ceba482d1075409bd2abd8ddebc07b8, 129 rows) for an unflagged site,
--     for loancalculator pre-seed, and for a nonexistent site id — the
--     widening is provably inert everywhere until a site opts in;
--   * executed via PREPARE with an UNTYPED $1 (one arg referenced twice),
--     proving the text-param → uuid coercion the Go driver path relies on;
--   * flag-ON preview for loancalculator: exactly 11 components = the 12
--     locked rows' distinct components (tool-loan-repayment is placed twice).
--   * zero live consumers of 'plan_includes_tools' today:
--     grep over the repo and SELECT over agent_definitions both empty.
--
-- PARAM PATH — the reason this was not a one-line edit (trap 3: a params
-- path resolving nil HARD-FAILS the step, an outage-class failure on the
-- one shared planner row). "site_record.site_id" is settled three ways:
--   1. the planner workflow is strictly linear (start_step read_specs →
--      ensure_site → load_existing_pages → load_components), and
--      EnsureSiteRecordAction returns "site_id" on EVERY path that lets the
--      workflow continue — DB path (site_db_actions.go:194) and DB-nil
--      placeholder (:970); the coordinator stores that map at
--      collected_data["site_record"] (coordinator.go:1877);
--   2. write_site_plan in this same workflow already binds the SAME path
--      ("target_site_id": "site_record.site_id", ExtractActionInputs
--      Strategy 0) — every site_plans row is one live run where it resolved
--      (10 since 2026-08-02, fresh build + repeat replans);
--   3. input_data.site_id was REJECTED: it is only as good as each
--      dispatcher's input_mapping, and input shapes vary by caller (that is
--      why extractDomainFromInput's aggressive search exists). 406 used
--      input_data.site_id for tool-suggester on the adjacent-step-already-
--      binds-it argument; the planner has no such step, and has a stronger
--      key of its own making.
--
-- CONSUMERS. 21 sites already place tool components (78 placements); none
-- is affected until its own structure spec opts in — the gate is per-site
-- and both gates must hold. content-gap-planner has its own
-- load_available_components step with the same section/element menu; it is
-- deliberately NOT widened here (gap-planning a new tool page is a different
-- authority and wants its own decision).
--
-- Config-only: live on apply, no image dependency. Council submission
-- accompanies this migration (advisory gate, RFC_022 counter live via
-- 402-405: this key is site_specs vocabulary, not a RegisterActionInputSpec
-- optional key, so the N=10 action-surface budget does not count it — the
-- structure-aspect key count, now 5, is stated here so the accumulation is
-- on the record rather than unremarked).
--
-- ROLLBACK: 407_build_site_planner_sees_the_sites_own_tools_ROLLBACK.sql,
-- or restore the snapshot this file takes
-- (snapshot_agent note '407_build_site_planner_sees_the_sites_own_tools:
-- pre-update').

BEGIN;

SELECT snapshot_agent('build-site-planner',
  '407_build_site_planner_sees_the_sites_own_tools: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,load_components,config,query}',
           to_jsonb('SELECT name, display_name, "function", category, description FROM content_components WHERE is_active = true AND ( component_level IN (''section'',''element'') OR ( component_level = ''tool'' AND EXISTS (SELECT 1 FROM site_specs ss WHERE ss.site_id = $1 AND ss.aspect = ''structure'' AND ss.is_current AND ss.data->>''plan_includes_tools'' = ''true'') AND id IN (SELECT pc.component_id FROM page_components pc JOIN pages p ON pc.page_id = p.id WHERE p.site_id = $1) ) ) ORDER BY category, name'::text)
         ),
         '{workflow,steps,load_components,config,params}',
         '["site_record.site_id"]'::jsonb
       ),
       updated_at = now()
 WHERE type = 'build-site-planner'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
  q text;
  p jsonb;
  n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '407: expected exactly 1 live build-site-planner row, found %', n;
  END IF;

  SELECT default_config#>>'{workflow,steps,load_components,config,query}',
         default_config#>'{workflow,steps,load_components,config,params}'
    INTO q, p
    FROM agent_definitions
   WHERE type = 'build-site-planner'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF q IS NULL
     OR position('plan_includes_tools' in q) = 0
     OR position('component_level = ''tool''' in q) = 0
     OR position('pc.component_id' in q) = 0 THEN
    RAISE EXCEPTION '407: load_components query does not carry both gates: %', q;
  END IF;
  IF p IS NULL OR jsonb_array_length(p) <> 1
     OR p->>0 <> 'site_record.site_id' THEN
    RAISE EXCEPTION '407: load_components params wrong: %', p;
  END IF;
END $$;

COMMIT;

-- 574 — required-fields-missing-handler: a failed LOOKUP is no longer a terminal CLOSE,
--       and a component that is real but not deployed parks with its facts instead of
--       being reported as gone. CONFIG ONLY — live on apply, no image, no roll.
--
--       Fixes bugs_open/367. Successor to seed 410 (bugs_closed/277), which built this
--       router; 410 is otherwise untouched.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- This agent is a ROUTER. It repairs nothing. For each `required_fields_missing` work
-- item it runs one SQL classification, then converts, parks, or closes the item.
--
-- To classify, it must first find the component the finding is about. It looked for that
-- component with `WHERE pc.build_status = 'deployed'`. If nothing came back, the routing
-- CASE fell through to `stale`, and `stale` CLOSES the item — `status='complete'`, no
-- error, one attempt — saying "the finding cannot be located on the live site".
--
-- That was coherent when this type had ONE producer. 410 mirrored that producer's own
-- filter: `check_required_fields_missing.go` is a POST-DEPLOY check and scans deployed
-- rows, so "no deployed row here any more" really did mean "my producer would no longer
-- file this".
--
-- bugs_closed/342 then added a SECOND producer, at render time
-- (`work_items_common.go:562`), whose entire stated purpose is to reach the population the
-- post-deploy check CANNOT see. Its findings are therefore, by construction, about
-- components that are not deployed. The router closed them as gone.
--
-- Confirmed live 2026-08-23, item 562788c3-c9e9-4e8b-9967-c16dc9b8ed36 (the classify SQL
-- was read out of THIS agent's row and run by hand with its own two params bound):
--
--     route = stale      component_id = ''      html_len = 0
--
-- while the component it "could not locate" was sitting there: page f438eca6
-- (ai-agent-roi-estimator, page_type=tool, rebuild_policy=owned, status=active), component
-- 0a1498b3, slot_name='tool-cta', build_status='pending', NOT locked, 9,220 characters of
-- rendered_html. The finding was true — `headline` and `trust_note` really were empty
-- (n_still_empty 2 of n_named 2).
--
-- THE HARM IS THE SILENCE, NOT THE FAILURE. Before 342's routability fix these items
-- failed three times and parked: ugly, and legible. Now a true finding is recorded
-- `complete` with no error and no trace of the disagreement, so any census asking "did we
-- action our required-fields findings?" scores it as a success.
--
-- ============================================================================
-- WHAT THIS MIGRATION DOES, AND WHY IT IS NOT SIMPLY "WIDEN THE FILTER"
-- ============================================================================
-- Widening the lookup so non-deployed components resolve is necessary but NOT sufficient,
-- and on its own it would be worse than it looks. Two measured reasons:
--
--   1. The repair arm would hard-fail. `partial` files a conversion through `file_rewrite`,
--      which reads `input_data.spec.component_id`, `.page_id`, `.component_function` and
--      `.reason`. The post-deploy producer writes all four (62 of 62 items carry
--      component_id and page_id, 2026-08-23); the render-time producer writes NONE of them
--      (0 of 3). Both an unresolved `spec_paths` entry and an unresolved
--      `item_key_suffix_field` are deliberate HARD ERRORS
--      (create_work_item_action.go:281,294 and :252-256), so the item would route
--      correctly and then die at the next step.
--
--   2. The repair arm is the WRONG PLACE for this population anyway.
--      `content_rewrite`/`edit_live` runs at page-build-handler, whose save step DELETEs
--      every agent-writable row on the page and reinserts at 'deployed'
--      (save_page_sections_action.go:823, :1014). Measured 2026-08-23: 28 of 31
--      `from_rfm` conversions are ALREADY failed at exactly that step on the owned-page
--      refusal — that is bugs_open/333, filed 2026-08-19 and owned by the 277 lane — and
--      the 367 page is itself rebuild_policy=owned. Separately, `pending` and
--      `needs_rebuild` are mid-flight states that fix_component_template_action.go:843
--      refuses to touch on purpose ("honest states whose repair is a rebuild, not a status
--      flip"), and the 17 `approved` rows are verbatim-adopted pages whose content_data is
--      a provenance envelope rather than source (adopt_verbatim.go:514).
--
-- So the rule this migration installs is narrower and stronger than "widen":
--
--     A DISPOSER MAY CLOSE ONLY ON POSITIVE EVIDENCE OF ABSENCE.
--     A FAILED LOOKUP IS NOT EVIDENCE, AND NEITHER IS A NON-DEPLOYED TARGET.
--
-- The estate already rules this way one door over. revalidate_review_queue_action.go:684,
-- on the same class of miss: "That MIGHT mean the finding is moot, but it might equally be
-- a lookup miss — so it is not positive evidence and the item stays queued."
--
-- Five edits, all inside this one agent:
--
--   1. comp CTE: resolve on the LIFECYCLE axis, not the build axis —
--      COALESCE(pc.build_status,'pending') <> 'removed'. Spelled to match the estate's one
--      existing named rendering of this judgement, `pageComponentNotRemovedSQL`
--      (section_editor_actions.go:1537, bugs_open/360). The COALESCE is load-bearing: NULL
--      passes the 049 CHECK constraint, and a bare <> is NULL-unsafe, which would drop such
--      a row silently — the wrong fail direction for something that closes items.
--
--   2. comp CTE also returns the target's state (`bs`), and a new `tomb` CTE asks whether a
--      'removed' row exists at (page_name, slot_name). That is the POSITIVE evidence that a
--      component was retired, as opposed to never having been found.
--
--   3. The `stale` arm now requires positive evidence: the page row is gone, OR the
--      resolved component is locked (277's deliberate accept-as-is resolution), OR nothing
--      resolves AND a removed row is sitting there. Nothing else closes.
--
--   4. A new route `target_not_dispatchable` catches the rest — nothing resolved, or the
--      component resolved but is not deployed — and PARKS at needs_human_review, holding
--      its dedup key exactly like the router's four existing park arms. It is a fifth park,
--      not a fifth close.
--
--   5. A `target_state` output column records WHICH leg fired (page_missing /
--      component_locked / component_retired / lookup_miss / the actual build_status). This
--      has to be an output column rather than close-arm prose, because the dispatch loop's
--      mark_complete OVERWRITES site_work_items.result on completed rows
--      (bugfix_277_required_fields_repair/RUNBOOK_required_fields_repair.md:60), so
--      orchestration_states is the only durable record of a close's cause.
--
--   Plus the two close arms' evidence strings, which are false as written. close_stale
--   asserts the component "no longer exists"; row-absence cannot support that claim when
--   336 of 2,160 slots named in pages.sections ordinarily have no page_components row at
--   all (measured 2026-08-23). close_resolved says "the currently-deployed component",
--   which edit 1 falsifies.
--
-- ============================================================================
-- WHAT THIS DELIBERATELY DOES NOT DO
-- ============================================================================
--   * It does not make the render-time population repairable. It makes it VISIBLE and
--     HONEST. Repair needs bugs_open/333 (owned-page routing) and a producer that writes
--     the convert arm's read-set; both are named, neither is taken here.
--   * It does not re-key file_rewrite to triage.component_id. Mechanically possible, but it
--     re-opens a council-settled key design (register CQ-023) and buys nothing once the
--     non-deployed population parks instead of converting.
--   * It does not register a verifier for this item_type. CQ-023 already warns that one
--     would fail-close this router's own `converted` arm.
--   * It does not touch the 19 hand-typed `pc.build_status = 'deployed'` reads elsewhere.
--     Those are PRODUCERS: their failure mode is under-detection. Only a disposer turns
--     non-detection into an affirmative claim of absence. That asymmetry is the transferable
--     lesson and it goes to 016b §9, not into a helper nobody has measured a wrong answer
--     for.
--
-- ============================================================================
-- MEASURED BEFORE WRITING (read-only, live DB, 2026-08-23)
-- ============================================================================
--   * The real item -> target_not_dispatchable, target_state='pending', component RESOLVES
--     (0a1498b3, html_len 9220), n_still_empty 2.
--   * POSITIVE CONTROL, retired slot (tool-clip-path/ported-page, site 6b49db8e) -> STILL
--     stale, target_state='component_retired'.
--   * POSITIVE CONTROL, page that does not exist -> STILL stale, target_state='page_missing'.
--   * WHOLE POPULATION, all 65 items of this type, old query vs new: EXACTLY ONE route
--     changes. partial 28, no_content_data 21, resolved 11, malformed 2, no_plan_owned 1,
--     no_plan_unbuildable 1 — all identical.
--   * Disconfirmable: 17 (page_id, slot_name) pairs carry more than one non-removed row and
--     0 of them switch which row `ORDER BY pc.updated_at DESC LIMIT 1` picks; and 'removed'
--     is a live, growing terminal (36 of 38 rows written in August 2026, most recent
--     2026-08-23 12:59Z), so the retirement control had something to bite on.
--
-- ROLLBACK: 574_required_fields_router_stops_closing_what_it_cannot_resolve_ROLLBACK.sql
-- ============================================================================

BEGIN;

DO $$
DECLARE
    v_row_count   int;
    v_q           text;
    v_new_q       text;
    v_ev_stale    text;
    v_ev_resolved text;
    v_steps       jsonb;

    -- Verbatim anchors. Every one of these was confirmed to occur EXACTLY ONCE in the
    -- live config on 2026-08-23. If any count is not 1, another lane has edited this
    -- agent since: stop, read what they did, and re-anchor rather than forcing this.
    c_predicate  CONSTANT text := 'pc.build_status = ''deployed''';
    c_comp_cols  CONSTANT text := '(pc.locked_at IS NOT NULL) AS locked, cc.input_schema AS sch';
    c_fx_anchor  CONSTANT text := 'fx AS (SELECT f.name,';
    c_case_arm   CONSTANT text := 'WHEN (SELECT count(*) FROM pg) = 0 OR (SELECT count(*) FROM comp) = 0 OR COALESCE((SELECT locked FROM comp), false) THEN ''stale'' WHEN (SELECT n_still_empty FROM fs) = 0 THEN ''resolved''';
    c_cid_out    CONSTANT text := 'COALESCE((SELECT cid::text FROM comp), '''') AS component_id';

    -- Replacements.
    c_predicate_new CONSTANT text := 'COALESCE(pc.build_status, ''pending'') <> ''removed''';
    c_comp_cols_new CONSTANT text := '(pc.locked_at IS NOT NULL) AS locked, cc.input_schema AS sch, COALESCE(pc.build_status, ''pending'') AS bs';
    c_tomb          CONSTANT text :=
        'tomb AS (SELECT EXISTS (SELECT 1 FROM page_components pc2 JOIN pg ON pc2.page_id = pg.id CROSS JOIN item '
        'WHERE COALESCE(pc2.build_status, ''pending'') = ''removed'' '
        'AND COALESCE(pc2.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''')) AS retired), ';
    c_case_new  CONSTANT text :=
        'WHEN (SELECT count(*) FROM pg) = 0 '
        'OR COALESCE((SELECT locked FROM comp), false) '
        'OR ((SELECT count(*) FROM comp) = 0 AND (SELECT retired FROM tomb)) THEN ''stale'' '
        'WHEN (SELECT n_still_empty FROM fs) = 0 THEN ''resolved'' '
        'WHEN (SELECT count(*) FROM comp) = 0 '
        'OR COALESCE((SELECT bs FROM comp), '''') <> ''deployed'' THEN ''target_not_dispatchable''';
    c_state_out CONSTANT text :=
        'CASE WHEN (SELECT count(*) FROM pg) = 0 THEN ''page_missing'' '
        'WHEN COALESCE((SELECT locked FROM comp), false) THEN ''component_locked'' '
        'WHEN (SELECT count(*) FROM comp) = 0 AND (SELECT retired FROM tomb) THEN ''component_retired'' '
        'WHEN (SELECT count(*) FROM comp) = 0 THEN ''lookup_miss'' '
        'WHEN COALESCE((SELECT bs FROM comp), '''') <> ''deployed'' THEN COALESCE((SELECT bs FROM comp), '''') '
        'ELSE '''' END AS target_state, ';

    -- The evidence strings as they stand today, asserted verbatim so a concurrent edit
    -- cannot be silently overwritten.
    c_ev_stale_old CONSTANT text := 'page or deployed component no longer exists at (page_name, slot_name), or the component is now locked — the producer''s own predicate no longer matches. The dedup key releases; discovery rotation (bugs_open/230, fixed 2026-08-09) re-raises within days if the finding is still real.';
    c_ev_res_old   CONSTANT text := 'every field named in spec.missing_fields is populated on the currently-deployed component — the same predicate the review-queue revalidator closes on';
BEGIN
    -- ------------------------------------------------------------------
    -- PREMISES. Every one aborts the whole transaction.
    -- ------------------------------------------------------------------
    SELECT count(*) INTO v_row_count FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_row_count <> 1 THEN
        RAISE EXCEPTION '574: expected exactly ONE active non-snapshot required-fields-missing-handler row, found % — 410 records this two-active-rows trap; resolve it before patching', v_row_count;
    END IF;

    SELECT default_config -> 'workflow' -> 'steps',
           default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query',
           default_config -> 'workflow' -> 'steps' -> 'close_stale' -> 'config' -> 'result_fields' ->> 'evidence',
           default_config -> 'workflow' -> 'steps' -> 'close_resolved' -> 'config' -> 'result_fields' ->> 'evidence'
      INTO v_steps, v_q, v_ev_stale, v_ev_resolved
      FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_q IS NULL THEN
        RAISE EXCEPTION '574: classify.config.query is absent — refusing to guess at the shape';
    END IF;

    -- Already applied? Say so and stop, rather than double-patching.
    IF position('target_not_dispatchable' in v_q) > 0 THEN
        RAISE EXCEPTION '574: classify already contains target_not_dispatchable — this migration is already applied';
    END IF;

    IF (length(v_q) - length(replace(v_q, c_predicate,  ''))) / length(c_predicate)  <> 1
    OR (length(v_q) - length(replace(v_q, c_comp_cols,  ''))) / length(c_comp_cols)  <> 1
    OR (length(v_q) - length(replace(v_q, c_fx_anchor,  ''))) / length(c_fx_anchor)  <> 1
    OR (length(v_q) - length(replace(v_q, c_case_arm,   ''))) / length(c_case_arm)   <> 1
    OR (length(v_q) - length(replace(v_q, c_cid_out,    ''))) / length(c_cid_out)    <> 1 THEN
        RAISE EXCEPTION '574: one of the five verbatim anchors does not occur exactly once in the live classify query — another lane has edited this agent. Re-read the live row and re-anchor; do NOT force this.';
    END IF;

    IF v_ev_stale IS DISTINCT FROM c_ev_stale_old THEN
        RAISE EXCEPTION '574: close_stale evidence is not the text this migration was written against — another lane got there first';
    END IF;
    IF v_ev_resolved IS DISTINCT FROM c_ev_res_old THEN
        RAISE EXCEPTION '574: close_resolved evidence is not the text this migration was written against — another lane got there first';
    END IF;

    -- Snapshot before writing (410's own idiom).
    PERFORM snapshot_agent('required-fields-missing-handler'::text,
                           '574: before making stale require positive evidence (bugs_open/367)'::text);

    -- ------------------------------------------------------------------
    -- THE FIVE EDITS
    -- ------------------------------------------------------------------
    v_new_q := v_q;
    v_new_q := replace(v_new_q, c_predicate, c_predicate_new);
    v_new_q := replace(v_new_q, c_comp_cols, c_comp_cols_new);
    v_new_q := replace(v_new_q, c_fx_anchor, c_tomb || c_fx_anchor);
    v_new_q := replace(v_new_q, c_case_arm,  c_case_new);
    v_new_q := replace(v_new_q, c_cid_out,   c_state_out || c_cid_out);

    IF v_new_q = v_q THEN
        RAISE EXCEPTION '574: the rewrite was a no-op — refusing to record a change that did not happen';
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(
               jsonb_set(
                   jsonb_set(
                       jsonb_set(default_config,
                           '{workflow,steps,classify,config,query}', to_jsonb(v_new_q), false),
                       '{workflow,steps,close_stale,config,result_fields,evidence}',
                       to_jsonb(
                           'CLOSED ON POSITIVE EVIDENCE, one of three: the page row is gone; or the component at '
                           '(page_name, slot_name) is LOCKED, which bugs_closed/277 treats as the deliberate '
                           'accept-as-is resolution; or nothing resolves there and a build_status=''removed'' row '
                           'is sitting at that slot, i.e. the component was retired. Which leg fired is '
                           'triage.target_state in the orchestration record — read it there, because '
                           'mark_complete overwrites this row''s result at completion. A lookup that simply finds '
                           'nothing is NOT one of these and no longer closes anything (bugs_open/367): 336 of 2,160 '
                           'slots named in pages.sections had no page_components row at all on 2026-08-23, so '
                           'absence of a row was never evidence that one existed. The dedup key releases; discovery '
                           'rotation (bugs_open/230, fixed 2026-08-09) re-raises within days if the finding is still real.'::text
                       ), false),
                   '{workflow,steps,close_resolved,config,result_fields,evidence}',
                   to_jsonb(
                       'every field named in spec.missing_fields is populated on the component currently resolving '
                       'at (page_name, slot_name), whatever its build_status — the same lookup scope the '
                       'review-queue revalidator closes on. Widened from deployed-only by 574 (bugs_open/367).'::text
                   ), false),
               '{workflow,steps,classify,description}',
               to_jsonb(
                   'One deterministic classification per item, resolved by (page_name, slot_name) — the '
                   'revalidator''s own key, never spec.component_id (unstable across rerenders). Resolution is on '
                   'the LIFECYCLE axis (any component not ''removed''), NOT the build axis: this router serves two '
                   'producers and the render-time one files only about components that are not deployed '
                   '(bugs_open/367). triage.target_state names why a stale/park disposition was reached.'::text
               ), false)
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- ------------------------------------------------------------------
    -- THE NEW ROUTE'S WIRING: a fifth PARK, spliced into the else-cascade.
    -- route_stale.else_step currently goes to route_resolved. It still does; the new
    -- branch is inserted after route_resolved's else arm, ahead of route_owned, so the
    -- non-dispatchable case is decided BEFORE any repair arm can claim it.
    -- ------------------------------------------------------------------
    UPDATE agent_definitions
       SET default_config = jsonb_set(
               jsonb_set(
                   jsonb_set(default_config,
                       '{workflow,steps,route_resolved,config,else_step}',
                       to_jsonb('route_not_dispatchable'::text), false),
                   '{workflow,steps,route_not_dispatchable}',
                   jsonb_build_object(
                       'action', 'conditional_branch',
                       'config', jsonb_build_object(
                           'condition', 'triage.route == target_not_dispatchable',
                           'then_step', 'park_not_dispatchable',
                           'else_step', 'route_owned'),
                       'error_step', 'mark_failed',
                       'output_field', 'd1b',
                       'description', 'The finding is real but its target is not in a state this router can safely dispatch a repair for — nothing resolved at (page_name, slot_name), or the component resolved and is not deployed. Park; never close (bugs_open/367).'),
                   true),
               '{workflow,steps,park_not_dispatchable}',
               jsonb_build_object(
                   'action', 'update_work_item_status',
                   'config', jsonb_build_object(
                       'status', 'needs_human_review',
                       'error_message',
                           'ROUTED BY required-fields-missing-handler, NOT DISPATCHABLE HERE: the finding stands, '
                           'but its target is not a deployed component. See the orchestration''s triage.target_state '
                           'for which — a build_status (pending / approved / needs_rebuild), or lookup_miss if '
                           'nothing resolved at (page_name, slot_name) at all. This is NOT evidence the component is '
                           'gone; a retired component has a ''removed'' row and closes as stale instead '
                           '(bugs_open/367). Why this is not auto-repaired: the convert arm files a content_rewrite '
                           'at page-build-handler, whose save step DELETEs every agent-writable row on the page and '
                           'reinserts at deployed (save_page_sections_action.go:823) — 28 of 31 such conversions '
                           'were already failing on the owned-page refusal as of 2026-08-23 (bugs_open/333), and '
                           '''pending''/''needs_rebuild'' are mid-flight states whose repair is a rebuild, not an '
                           'edit (fix_component_template_action.go:843). Repair paths, human''s call: deploy or '
                           'rebuild the component so the normal routes apply; LOCK it, which is the accept-as-is '
                           'resolution and closes the finding; or retire it, which closes it as stale. Parked '
                           'holding its dedup key so the producer cannot churn re-raises.',
                       'result_fields', jsonb_build_object(
                           'route', 'target_not_dispatchable',
                           'triaged_by', 'required-fields-missing-handler'),
                       'skip_if_missing', false,
                       'work_item_id_field', 'input_data.work_item_id'),
                   'next_step', 'done',
                   'description', 'Park-in-place: the finding is true and no safe automated repair exists for a non-deployed target. A close here would be the bugs_open/367 defect.'),
               true)
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
END $$;

-- ============================================================================
-- VERIFY — with negative controls. Any failure aborts the COMMIT.
-- A block of bare SELECTs cannot do this: ON_ERROR_STOP ignores a non-empty result set,
-- so the assertions must RAISE (bugs_closed/RFC_006's trap).
-- ============================================================================
DO $$
DECLARE
    v_q     text;
    v_steps jsonb;
    v_n     int;
BEGIN
    SELECT default_config -> 'workflow' -> 'steps',
           default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO v_steps, v_q
      FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- 1. The old predicate is GONE and the new one is present exactly once.
    IF position('pc.build_status = ''deployed''' in v_q) > 0 THEN
        RAISE EXCEPTION '574 VERIFY: the deployed-only predicate survives in classify';
    END IF;
    v_n := (length(v_q) - length(replace(v_q, 'COALESCE(pc.build_status, ''pending'') <> ''removed''', ''))) / length('COALESCE(pc.build_status, ''pending'') <> ''removed''');
    IF v_n <> 1 THEN
        RAISE EXCEPTION '574 VERIFY: the lifecycle predicate occurs % times, expected 1', v_n;
    END IF;

    -- 2. The new route and its evidence column exist exactly once each.
    v_n := (length(v_q) - length(replace(v_q, 'target_not_dispatchable', ''))) / length('target_not_dispatchable');
    IF v_n <> 1 THEN
        RAISE EXCEPTION '574 VERIFY: target_not_dispatchable occurs % times in classify, expected exactly 1 (the CASE arm)', v_n;
    END IF;
    IF position('AS target_state' in v_q) = 0 OR position('tomb AS (SELECT EXISTS' in v_q) = 0 THEN
        RAISE EXCEPTION '574 VERIFY: target_state output or the tomb CTE is missing';
    END IF;

    -- 3. NEGATIVE CONTROL: the rest of the query is intact. A botched replace() that ate a
    --    CTE would still satisfy the checks above.
    IF position('plan_src AS (' in v_q) = 0
    OR position('has_open_tool_recreation' in v_q) = 0
    OR position('n_asset_empty' in v_q) = 0
    OR position('no_plan_unbuildable' in v_q) = 0 THEN
        RAISE EXCEPTION '574 VERIFY: a pre-existing part of the classify query was lost';
    END IF;

    -- 4. classify still takes exactly (site_id, work_item_id).
    IF jsonb_array_length(v_steps #> '{classify,config,params}') <> 2 THEN
        RAISE EXCEPTION '574 VERIFY: classify no longer takes exactly two params';
    END IF;

    -- 5. Step count: 410 shipped 20; this adds exactly 2.
    SELECT count(*) INTO v_n FROM jsonb_object_keys(v_steps);
    IF v_n <> 22 THEN
        RAISE EXCEPTION '574 VERIFY: expected 22 steps after adding 2, found %', v_n;
    END IF;

    -- 6. Every branch names both arms — 410's own invariant, re-asserted because this
    --    migration rewires one of them. An unnamed arm stops an item mid-route.
    IF v_steps #>> '{route_stale,config,then_step}'            IS DISTINCT FROM 'close_stale'
    OR v_steps #>> '{route_stale,config,else_step}'            IS DISTINCT FROM 'route_resolved'
    OR v_steps #>> '{route_resolved,config,then_step}'         IS DISTINCT FROM 'close_resolved'
    OR v_steps #>> '{route_resolved,config,else_step}'         IS DISTINCT FROM 'route_not_dispatchable'
    OR v_steps #>> '{route_not_dispatchable,config,then_step}' IS DISTINCT FROM 'park_not_dispatchable'
    OR v_steps #>> '{route_not_dispatchable,config,else_step}' IS DISTINCT FROM 'route_owned'
    OR v_steps #>> '{route_owned,config,then_step}'            IS DISTINCT FROM 'park_owned'
    OR v_steps #>> '{route_owned,config,else_step}'            IS DISTINCT FROM 'route_blob'
    OR v_steps #>> '{route_blob,config,then_step}'             IS DISTINCT FROM 'park_blob'
    OR v_steps #>> '{route_blob,config,else_step}'             IS DISTINCT FROM 'route_asset'
    OR v_steps #>> '{route_asset,config,then_step}'            IS DISTINCT FROM 'park_asset'
    OR v_steps #>> '{route_asset,config,else_step}'            IS DISTINCT FROM 'route_noplan'
    OR v_steps #>> '{route_noplan,config,then_step}'           IS DISTINCT FROM 'file_recreate'
    OR v_steps #>> '{route_noplan,config,else_step}'           IS DISTINCT FROM 'route_unbuildable'
    OR v_steps #>> '{route_unbuildable,config,then_step}'      IS DISTINCT FROM 'park_unbuildable'
    OR v_steps #>> '{route_unbuildable,config,else_step}'      IS DISTINCT FROM 'route_partial'
    OR v_steps #>> '{route_partial,config,then_step}'          IS DISTINCT FROM 'file_rewrite'
    OR v_steps #>> '{route_partial,config,else_step}'          IS DISTINCT FROM 'mark_failed' THEN
        RAISE EXCEPTION '574 VERIFY: a conditional_branch arm is missing or mis-wired — an item would stop mid-route';
    END IF;

    -- 7. The park arms must PARK, not complete. The new one included — a close there would
    --    be the very defect this migration fixes, wearing a new name.
    IF v_steps #>> '{park_owned,config,status}'            IS DISTINCT FROM 'needs_human_review'
    OR v_steps #>> '{park_blob,config,status}'             IS DISTINCT FROM 'needs_human_review'
    OR v_steps #>> '{park_asset,config,status}'            IS DISTINCT FROM 'needs_human_review'
    OR v_steps #>> '{park_unbuildable,config,status}'      IS DISTINCT FROM 'needs_human_review'
    OR v_steps #>> '{park_not_dispatchable,config,status}' IS DISTINCT FROM 'needs_human_review' THEN
        RAISE EXCEPTION '574 VERIFY: a park arm does not park at needs_human_review — dedup-key churn returns';
    END IF;

    -- 8. The new park must reach a terminal step, or the item hangs.
    IF v_steps #>> '{park_not_dispatchable,next_step}' IS DISTINCT FROM 'done' THEN
        RAISE EXCEPTION '574 VERIFY: park_not_dispatchable does not reach done';
    END IF;

    RAISE NOTICE '574 OK: lifecycle resolution live, stale requires positive evidence, park_not_dispatchable wired, 22 steps.';
END $$;

COMMIT;

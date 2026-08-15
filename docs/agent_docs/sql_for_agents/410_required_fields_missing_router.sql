-- 410 — required-fields-missing-handler: the repair router for the one item type
--       the owner ruled on this morning
--
-- OWNER RULING 2026-08-15 (bugs_open/277): "we should create a repair handler
-- fleet wide" for required_fields_missing. This is that handler, built as a
-- ROUTER, NOT A REPAIRER, on the 397 pattern (IMG-071): one deterministic
-- classification per item, then convert what has an owning repair path, close
-- what no longer holds, and PARK-IN-PLACE what must not be auto-repaired.
--
-- ============================================================================
-- WHY THE CHECK'S FLAG-ONLY DESIGN IS SUPERSEDED, NOT OVERRULED
-- ============================================================================
-- check_required_fields_missing.go files this type with HandlerAgent "" and
-- status needs_human_review, on the stated ground that "the honest resolutions
-- — give the site a data source, or remove the component — are human
-- decisions." That reasoning holds for TWO of the six classes below, and for
-- exactly those two this router still parks the item for a human — but parked
-- WITH the classification, the evidence and the safe options written on the
-- row, instead of parked as a bare finding nobody can act on. Measured
-- 2026-08-15: exactly ONE required_fields_missing item in platform history had
-- ever reached a terminal status before the 033 drain; the queue the type
-- parks in held 735 items.
--
-- The census over all 44 open items (read-only, run 2026-08-15, saved with its
-- output in docs024_key_docs_latest/bugfix_277_required_fields_repair/):
--
--   no_content_data 35   stale 6   no_plan_generic 1   no_plan_owned 1   partial 1
--
-- ============================================================================
-- THE SIX ROUTES, AND WHY EACH DOES WHAT IT DOES
-- ============================================================================
-- Classification key is (page_name, slot_name) — the review-queue
-- revalidator's own key — NEVER spec.component_id: page_components ids are
-- unstable across rerenders (016b §9; 11/45 items resolved to nothing when
-- keyed on component_id, while the section sat right there under a new row id).
--
--   stale            page gone, no deployed component at (page,slot), or the
--                    component is now locked (the producer skips locked rows,
--                    so its own predicate no longer matches).
--                    -> CLOSE with evidence. The dedup key releases; discovery
--                    rotation (bugs_open/230, FIXED 2026-08-09, ~9 site
--                    examinations/day) re-raises within days if still real.
--   resolved         every field named in spec.missing_fields is populated on
--                    the deployed component — the revalidator's own closure
--                    predicate. -> CLOSE with evidence.
--   no_plan_owned    zero sections from both sources AND the page is
--                    tool/game-typed or rebuild_policy='owned'. The owned-page
--                    guard (reconcile_site_plan decision 3, TP-004/TL-001) is
--                    itself an owner ruling: the generic builder clobbers owned
--                    pages. -> PARK with the tool-pipeline route named.
--   no_content_data  content_data NULL/'{}' while substantial rendered_html
--                    serves (the decomposition-legacy blob class, 35 of 44).
--                    Regenerating through the template would REPLACE the served
--                    HTML with a single template render (bugs_open/263).
--                    -> PARK with the danger and the safe options named:
--                    decompose properly, lock the component (the producer skips
--                    locked), or retire the page.
--   no_plan_generic  sections empty everywhere, page not owned, component not
--                    a blob, AND page_type in ('', 'content', 'landing') — the
--                    home turf of defaultSectionsForPage's generic layout.
--                    -> CONVERT to needs_content_page @ page-build-handler
--                    with spec.mode='recreate' (the check_sectionless_pages
--                    idiom) — but at status 'triaged', NOT 'detected': the
--                    detected->triaged promoter is disabled (bugs_open/083).
--   no_plan_unbuildable  sections empty everywhere, not owned, but the
--                    page_type has NO sane generic archetype (measured
--                    2026-08-15 on the one live row: leopardess blog, a
--                    blog-index whose recreate no-opped at
--                    mark_no_ready_sections, and defaultSectionsForPage would
--                    hand it hero/generic-text/CTA — a listing page rebuilt
--                    as prose). The blog/section-index machinery owns these
--                    (needs_blog_posts -> blog-content-planner; bugs 015/206
--                    family). -> PARK with that routing fact. (Round-2's
--                    "committed fallback" — ensure_page_section_layout — was
--                    investigated and REJECTED on this measurement.)
--   asset_sourced    >=1 still-empty field is declared source site_assets.* in
--                    the component's schema (v2 dialect; legacy defaults to
--                    llm). THE PROSE WRITER MUST NOT FILL THESE — it can only
--                    mint URLs, and the first canary conversion PROVED the
--                    outcome: validate_content refused the writer's output
--                    (2026-08-15, item content_rewrite:from_rfm:_..dda5fbdf,
--                    "0 blockers, 1 errors", artefact untouched). Repair
--                    belongs to the asset/imagery pipeline (bind or deploy the
--                    asset for the declared purpose — bugs 236/238 family), or
--                    the schema is mistyped and needs amending. -> PARK with
--                    that routing fact. (Round-2 council gating objection made
--                    this a route; it began as a canary surprise.)
--   partial          plan exists, content_data populated, >=1 named field
--                    still empty, none asset-sourced. -> CONVERT to
--                    content_rewrite @ page-build-handler with
--                    spec.mode='edit_live' (PBP-028) so the writer edits
--                    current prose rather than fabricating a replacement;
--                    bugs_open/238's resolver-key protection is confirmed in
--                    the running binary (stamp a2a6912...).
--   (anything else)  malformed spec / unknown route -> mark_failed, loudly.
--
-- WHY PARK-IN-PLACE AND NOT checkpoint_for_review: the checkpoint action
-- writes NO item_key (no dedup) and hardcodes handler_agent='human-review' (an
-- unregistered agent), and completing the original releases its dedup key — so
-- the producer re-raises on the next pass and the two-strike rule births
-- endless 'unresolved' rows (5 keys of this very type already sat at 1 strike,
-- measured 2026-08-04, work_items_common.go:123-125). A parked row HOLDS its
-- dedup key: churn is structurally impossible. Parking one's own item is
-- first-class: complete_work_item's guarded UPDATE excludes needs_human_review
-- and returns {completed:false, reason:"already_flagged_or_terminal"} as
-- SUCCESS (load_work_item_actions.go:956-978) — never mark_failed. A dashboard
-- Retry click on a parked row re-dispatches the router, which re-parks:
-- idempotent by design.
--
-- CONVERSION ITEM KEYS: content_rewrite:from_rfm: + spec.component_id (stable
-- across re-raises of the same component, so the two-strike brake works on the
-- conversion namespace) and needs_content_page:from_rfm: + spec.page_name.
--
-- ============================================================================
-- NOTHING IS ASSIGNED BY THIS FILE. THE ROWS ARE INERT UNTIL THE CANARY.
-- ============================================================================
-- The producer still emits HandlerAgent "" until its Go change rides the next
-- chassis roll, and no existing row is touched here (the final DO block ASSERTS
-- 0 assigned). Assignment is a deliberate, canaried, one-off operational step:
--
--   -- CANARY (3-4 representative ids — one stale, the partial, one blob, the
--   -- gas-converter row — taken from the census output):
--   UPDATE site_work_items
--      SET handler_agent = 'required-fields-missing-handler',
--          status = 'triaged', attempt_count = 0, updated_at = NOW()
--    WHERE item_type = 'required_fields_missing'
--      AND status IN ('needs_human_review','unresolved')
--      AND COALESCE(handler_agent,'') = ''
--      AND id IN ('<canary ids>');
--   -- verify each arm against the census prediction (RUNBOOK), then FLEET:
--   -- the same UPDATE without the id filter.
--
-- attempt_count = 0 matters: the claim gate requires attempt_count <
-- max_attempts (claim_work_item_action.go:103), mirroring HandleRetryWorkItem.
--
-- image_tag v1.0.1300 — the fleet tag live at authoring time (re-read the
-- makefile before applying; a spawned handler runs its SPAWNER's image anyway,
-- bugs_open/066). Every referenced action is long-registered (query_database,
-- conditional_branch, create_work_item, update_work_item_status,
-- complete_workflow) — this config waits on no image. create_work_item is
-- StrictConfig with `spec` RETIRED (bugs_open/234, migration 364): specs below
-- are built from spec_paths/spec_literal, and site_id/page_id/parent_item_id/
-- summary are its declared data inputs.
--
-- ORDER: DB config, live the moment it commits — but inert, per above.
-- ROLLBACK: 410_required_fields_missing_router_ROLLBACK.sql (refuses once
-- anything routes here; un-assign first, exactly like 397's).
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 410_….sql
-- ============================================================================

BEGIN;

-- Conditional snapshot: snapshot_agent() RAISEs on a type that does not exist
-- yet, and this file CREATES the type — so snapshot only when re-running (the
-- INSERT below carries ON CONFLICT DO UPDATE and so mutates on re-run). Same
-- reasoning and shape as 397.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM agent_definitions
                WHERE type = 'required-fields-missing-handler' AND is_active
                  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) THEN
        PERFORM snapshot_agent('required-fields-missing-handler', '410_required_fields_missing_router.sql: pre-update');
        RAISE NOTICE '410: snapshotted existing required-fields-missing-handler before update';
    ELSE
        RAISE NOTICE '410: required-fields-missing-handler does not exist yet — nothing to snapshot (first application)';
    END IF;
END $$;

INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, is_active, status,
    default_config, topics, env_vars, idle_timeout_seconds
) VALUES (
    'required-fields-missing-handler',
    'Required Fields Missing Router',
    'Routes a required_fields_missing finding by re-reading live DB state at (page_name, slot_name): closes items whose predicate no longer holds, converts repairable items into content_rewrite (edit_live) or needs_content_page (recreate) for page-build-handler, and parks-in-place (holding the dedup key) the two classes that must not be auto-repaired — blob components whose regeneration would replace served HTML (bugs_open/263) and owned/tool pages the owned-page guard protects. Never edits a page itself. bugs_open/277, seed 410.',
    'site',
    'docker.io/aqls/agent-chassis', 'v1.0.1300',
    true, 'active',
    '{
        "workflow": {
            "start_step": "classify",
            "steps": {
                "classify": {
                    "action": "query_database",
                    "config": {
                        "output_format": "object",
                        "query": "WITH item AS (SELECT spec FROM site_work_items WHERE id = $2::uuid), pg AS (SELECT p.id, COALESCE(p.page_type, '''') AS page_type, COALESCE(p.rebuild_policy, ''generic'') AS rebuild_policy, (p.sections IS NULL OR p.sections = ''[]''::jsonb) AS sections_empty FROM pages p CROSS JOIN item WHERE p.site_id = $1::uuid AND p.name = item.spec->>''page_name'' AND COALESCE(p.status, '''') <> ''deleted'' LIMIT 1), plan_src AS (SELECT EXISTS (SELECT 1 FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id CROSS JOIN item WHERE sp.site_id = $1::uuid AND sp.is_current = true AND sps.page_name = item.spec->>''page_name'') AS has_plan_sections), comp AS (SELECT pc.id AS cid, pc.content_data AS cd, length(COALESCE(pc.rendered_html, '''')) AS html_len, (pc.locked_at IS NOT NULL) AS locked, cc.input_schema AS sch FROM page_components pc JOIN pg ON pc.page_id = pg.id LEFT JOIN content_components cc ON cc.id = pc.component_id CROSS JOIN item WHERE pc.build_status = ''deployed'' AND COALESCE(pc.slot_name, '''') = COALESCE(item.spec->>''slot_name'', '''') ORDER BY pc.updated_at DESC NULLS LAST LIMIT 1), fx AS (SELECT f.name, ((SELECT count(*) FROM comp) = 0 OR (SELECT cd FROM comp) IS NULL OR (SELECT cd->f.name FROM comp) IS NULL OR jsonb_typeof((SELECT cd->f.name FROM comp)) = ''null'' OR (jsonb_typeof((SELECT cd->f.name FROM comp)) = ''string'' AND btrim((SELECT cd->>f.name FROM comp)) = '''') OR (jsonb_typeof((SELECT cd->f.name FROM comp)) = ''array'' AND jsonb_array_length((SELECT cd->f.name FROM comp)) = 0) OR (jsonb_typeof((SELECT cd->f.name FROM comp)) = ''object'' AND (SELECT cd->f.name FROM comp) = ''{}''::jsonb)) AS is_empty, COALESCE((SELECT sch->''fields''->f.name->>''source'' FROM comp), ''llm'') AS src FROM item CROSS JOIN LATERAL jsonb_array_elements_text(COALESCE(item.spec->''missing_fields'', ''[]''::jsonb)) f(name)), fs AS (SELECT count(*) AS n_named, count(*) FILTER (WHERE is_empty) AS n_still_empty, count(*) FILTER (WHERE is_empty AND src LIKE ''site_assets.%'') AS n_asset_empty FROM fx) SELECT CASE WHEN (SELECT count(*) FROM item) = 0 OR (SELECT spec->>''page_name'' FROM item) IS NULL OR jsonb_typeof((SELECT spec->''missing_fields'' FROM item)) IS DISTINCT FROM ''array'' OR jsonb_array_length(COALESCE((SELECT spec->''missing_fields'' FROM item), ''[]''::jsonb)) = 0 THEN ''malformed'' WHEN (SELECT count(*) FROM pg) = 0 OR (SELECT count(*) FROM comp) = 0 OR COALESCE((SELECT locked FROM comp), false) THEN ''stale'' WHEN (SELECT n_still_empty FROM fs) = 0 THEN ''resolved'' WHEN COALESCE((SELECT sections_empty FROM pg), false) AND NOT (SELECT has_plan_sections FROM plan_src) AND ((SELECT page_type FROM pg) IN (''tool'', ''game'') OR (SELECT rebuild_policy FROM pg) = ''owned'') THEN ''no_plan_owned'' WHEN (SELECT cd FROM comp) IS NULL OR (SELECT cd FROM comp) = ''{}''::jsonb THEN ''no_content_data'' WHEN (SELECT n_asset_empty FROM fs) > 0 THEN ''asset_sourced'' WHEN COALESCE((SELECT sections_empty FROM pg), false) AND NOT (SELECT has_plan_sections FROM plan_src) AND (SELECT page_type FROM pg) IN ('''', ''content'', ''landing'') THEN ''no_plan_generic'' WHEN COALESCE((SELECT sections_empty FROM pg), false) AND NOT (SELECT has_plan_sections FROM plan_src) THEN ''no_plan_unbuildable'' ELSE ''partial'' END AS route, COALESCE((SELECT cid::text FROM comp), '''') AS component_id, COALESCE((SELECT html_len FROM comp), 0) AS html_len, (SELECT n_named FROM fs) AS n_named, (SELECT n_still_empty FROM fs) AS n_still_empty, (SELECT n_asset_empty FROM fs) AS n_asset_empty, COALESCE((SELECT page_type FROM pg), '''') AS page_type, COALESCE((SELECT rebuild_policy FROM pg), '''') AS rebuild_policy, COALESCE((SELECT sections_empty FROM pg), false) AS sections_empty, (SELECT has_plan_sections FROM plan_src) AS has_plan_sections, EXISTS (SELECT 1 FROM site_work_items t CROSS JOIN item WHERE t.site_id = $1::uuid AND t.item_type = ''needs_tool_recreation'' AND t.spec->>''page_name'' = item.spec->>''page_name'' AND t.status NOT IN (''complete'', ''verified'', ''rejected'', ''wont_fix'', ''cancelled'', ''failed'', ''unresolved'')) AS has_open_tool_recreation",
                        "params": ["input_data.site_id", "input_data.work_item_id"]
                    },
                    "next_step": "route_stale",
                    "error_step": "mark_failed",
                    "output_field": "triage",
                    "description": "One deterministic classification per item, resolved by (page_name, slot_name) — the revalidator''s own key, never spec.component_id (unstable across rerenders)"
                },
                "route_stale": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == stale", "then_step": "close_stale", "else_step": "route_resolved" },
                    "error_step": "mark_failed", "output_field": "d1",
                    "description": "Page/component gone or locked — the producer''s own predicate no longer matches"
                },
                "route_resolved": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == resolved", "then_step": "close_resolved", "else_step": "route_owned" },
                    "error_step": "mark_failed", "output_field": "d2",
                    "description": "Every named field now populated — the revalidator''s closure predicate"
                },
                "route_owned": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == no_plan_owned", "then_step": "park_owned", "else_step": "route_blob" },
                    "error_step": "mark_failed", "output_field": "d3",
                    "description": "Owned/tool page with no plan — the owned-page guard forbids a generic rebuild"
                },
                "route_blob": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == no_content_data", "then_step": "park_blob", "else_step": "route_asset" },
                    "error_step": "mark_failed", "output_field": "d4",
                    "description": "Blob class: content_data empty while rendered_html serves — regeneration would replace served HTML (bugs_open/263)"
                },
                "route_asset": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == asset_sourced", "then_step": "park_asset", "else_step": "route_noplan" },
                    "error_step": "mark_failed", "output_field": "d7",
                    "description": "Still-empty fields declared source site_assets.* — the prose writer must not fill these (round-2 refinement, proven by the first canary conversion''s validate_content refusal)"
                },
                "route_noplan": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == no_plan_generic", "then_step": "file_recreate", "else_step": "route_unbuildable" },
                    "error_step": "mark_failed", "output_field": "d5",
                    "description": "Sectionless content/landing page — rebuild through the plan/sibling-layout path"
                },
                "route_unbuildable": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == no_plan_unbuildable", "then_step": "park_unbuildable", "else_step": "route_partial" },
                    "error_step": "mark_failed", "output_field": "d8",
                    "description": "Sectionless page whose type has no generic archetype (blog/section-index family) — measured: the recreate no-ops and the default layout would be semantically wrong"
                },
                "route_partial": {
                    "action": "conditional_branch",
                    "config": { "condition": "triage.route == partial", "then_step": "file_rewrite", "else_step": "mark_failed" },
                    "error_step": "mark_failed", "output_field": "d6",
                    "description": "Fallthrough is partial; an unknown or malformed route fails LOUDLY rather than closing an item nothing acted on"
                },
                "close_stale": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "complete",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "result_fields": {
                            "route": "stale",
                            "closed_by": "required-fields-missing-handler",
                            "evidence": "page or deployed component no longer exists at (page_name, slot_name), or the component is now locked — the producer''s own predicate no longer matches. The dedup key releases; discovery rotation (bugs_open/230, fixed 2026-08-09) re-raises within days if the finding is still real."
                        }
                    },
                    "next_step": "done",
                    "description": "Close with evidence — the finding cannot be located on the live site"
                },
                "close_resolved": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "complete",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "result_fields": {
                            "route": "resolved",
                            "closed_by": "required-fields-missing-handler",
                            "evidence": "every field named in spec.missing_fields is populated on the currently-deployed component — the same predicate the review-queue revalidator closes on"
                        }
                    },
                    "next_step": "done",
                    "description": "Close with positive evidence"
                },
                "park_owned": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "needs_human_review",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "error_message": "ROUTED BY required-fields-missing-handler, NOT REPAIRABLE HERE: tool/game/owned page with no section plan from any source. The generic builder is refused for owned pages (reconcile_site_plan decision 3 / TP-004 — it clobbers them). Repair path is the tool pipeline (needs_tool_recreation / tool-improver) or an owner decision. Parked holding its dedup key so the producer cannot churn re-raises; see result.route and the orchestration''s triage facts.",
                        "result_fields": { "route": "no_plan_owned", "triaged_by": "required-fields-missing-handler" }
                    },
                    "next_step": "done",
                    "description": "Park-in-place: the owned-page guard is itself an owner ruling and wins"
                },
                "park_blob": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "needs_human_review",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "error_message": "ROUTED BY required-fields-missing-handler, NOT AUTO-REPAIRABLE: content_data is empty while substantial rendered_html serves (decomposition-legacy blob). Regenerating through the template would REPLACE the served HTML with a single template render (bugs_open/263). Safe options, human''s call: decompose properly (staged_component_build''s domain), lock the component (the producer skips locked rows — this is the accept-as-is resolution), or retire the page. Parked holding its dedup key so the producer cannot churn re-raises.",
                        "result_fields": { "route": "no_content_data", "triaged_by": "required-fields-missing-handler" }
                    },
                    "next_step": "done",
                    "description": "Park-in-place: a wrong repair here deletes served content; the revalidator remains a second drain if content_data is ever populated"
                },
                "park_unbuildable": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "needs_human_review",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "error_message": "ROUTED BY required-fields-missing-handler, NOT GENERICALLY REBUILDABLE: the page has no section plan from any source AND its page_type has no sane generic archetype (defaultSectionsForPage would hand a listing/index page hero + generic prose). Measured 2026-08-15: a needs_content_page recreate for exactly this page no-opped at mark_no_ready_sections. The blog/section-index machinery owns this rebuild (needs_blog_posts -> blog-content-planner; bugs 015/206 family), or an owner decision retires the page. Parked holding its dedup key so the producer cannot churn re-raises.",
                        "result_fields": { "route": "no_plan_unbuildable", "triaged_by": "required-fields-missing-handler" }
                    },
                    "next_step": "done",
                    "description": "Round-2 verification demanded the no-op be settled up front; it no-opped, so this class parks with the owning machinery named"
                },
                "park_asset": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "needs_human_review",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "error_message": "ROUTED BY required-fields-missing-handler, NOT WRITER-REPAIRABLE: at least one still-empty required field is declared source site_assets.* in the component schema (see the orchestration''s triage.n_asset_empty). A prose writer can only MINT a URL for such a field — measured 2026-08-15: the one conversion attempted before this route existed was refused by validate_content and changed nothing. Repair paths, human''s call: bind or deploy the site asset for the declared purpose (asset/imagery pipeline — the bugs 236/238 image-binding family), or amend the component schema if the field should be llm-sourced. Parked holding its dedup key so the producer cannot churn re-raises.",
                        "result_fields": { "route": "asset_sourced", "triaged_by": "required-fields-missing-handler" }
                    },
                    "next_step": "done",
                    "description": "Round-2 refinement: asset-sourced fields are the imagery pipeline''s to fill, never the writer''s"
                },
                "file_recreate": {
                    "action": "create_work_item",
                    "config": {
                        "site_id": "input_data.site_id",
                        "item_type": "needs_content_page",
                        "handler_agent": "page-build-handler",
                        "item_pipeline": "build",
                        "severity": "high",
                        "source": "required-fields-missing-handler",
                        "status": "triaged",
                        "priority": 90,
                        "summary": "Sectionless page rebuild requested by required-fields-missing-handler (converted from required_fields_missing)",
                        "item_key_prefix": "needs_content_page:from_rfm:",
                        "item_key_suffix_field": "input_data.spec.page_name",
                        "parent_item_id": "input_data.work_item_id",
                        "page_id": "input_data.spec.page_id",
                        "spec_paths": {
                            "page_name": "input_data.spec.page_name",
                            "page_id": "input_data.spec.page_id"
                        },
                        "spec_literal": {
                            "check": "required_fields_missing",
                            "mode": "recreate",
                            "reason": "page has zero sections from every source; rebuild through the plan/sibling-layout path (sectionless_pages idiom, routed by seed 410)"
                        }
                    },
                    "next_step": "close_converted",
                    "error_step": "mark_failed",
                    "output_field": "conversion",
                    "description": "Filed at triaged, NOT detected — the detected->triaged promoter is disabled (bugs_open/083) and a detected item is stranded"
                },
                "file_rewrite": {
                    "action": "create_work_item",
                    "config": {
                        "site_id": "input_data.site_id",
                        "item_type": "content_rewrite",
                        "handler_agent": "page-build-handler",
                        "item_pipeline": "build",
                        "severity": "medium",
                        "source": "required-fields-missing-handler",
                        "status": "triaged",
                        "priority": 30,
                        "summary": "Populate schema-required fields left empty on a deployed component (converted from required_fields_missing)",
                        "item_key_prefix": "content_rewrite:from_rfm:",
                        "item_key_suffix_field": "input_data.spec.component_id",
                        "parent_item_id": "input_data.work_item_id",
                        "page_id": "input_data.spec.page_id",
                        "spec_paths": {
                            "page": "input_data.spec.page_name",
                            "page_name": "input_data.spec.page_name",
                            "page_id": "input_data.spec.page_id",
                            "slot_name": "input_data.spec.slot_name",
                            "section_name": "input_data.spec.slot_name",
                            "component_function": "input_data.spec.component_function",
                            "missing_fields": "input_data.spec.missing_fields",
                            "reason": "input_data.spec.reason"
                        },
                        "spec_literal": {
                            "mode": "edit_live",
                            "origin_check": "required_fields_missing",
                            "description": "Populate ONLY the schema-required fields named in missing_fields for the named slot.",
                            "suggestion": "Fill the missing required fields with accurate values consistent with the page. PRESERVE all existing copy — mode=edit_live hands you the current rendered_html to edit, not replace."
                        }
                    },
                    "next_step": "close_converted",
                    "error_step": "mark_failed",
                    "output_field": "conversion",
                    "description": "edit_live (PBP-028): the writer edits current prose while filling the named fields; item_key suffixed by component_id so the two-strike brake works on a stable key"
                },
                "close_converted": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "complete",
                        "work_item_id_field": "input_data.work_item_id",
                        "skip_if_missing": false,
                        "result_fields": {
                            "route": "converted",
                            "closed_by": "required-fields-missing-handler",
                            "note": "repair filed as a follow-on item at page-build-handler; see result.response.conversion for its key"
                        }
                    },
                    "next_step": "done",
                    "description": "Close the original — its repair now lives in an item type page-build-handler already owns"
                },
                "mark_failed": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "failed",
                        "skip_if_missing": true,
                        "work_item_id_field": "input_data.work_item_id"
                    },
                    "next_step": "done",
                    "description": "Fail loudly; the error column picks up the routed step error automatically"
                },
                "done": {
                    "action": "complete_workflow",
                    "config": { "output_fields": ["triage", "conversion"] },
                    "description": "Routing complete"
                }
            }
        },
        "processing_mode": "task"
    }'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '[]'::jsonb, 0
) ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description    = EXCLUDED.description,
    image_tag      = EXCLUDED.image_tag,
    updated_at     = NOW();

-- ----------------------------------------------------------------------------
-- Verification. IS DISTINCT FROM (a jsonb path miss returns NULL and NULL <>
-- never fires); RAISE EXCEPTION (ON_ERROR_STOP ignores a non-empty SELECT).
-- ----------------------------------------------------------------------------
DO $$
DECLARE
    cfg          jsonb;
    assigned     integer;
    active_count integer;
BEGIN
    -- Exactly ONE active definition row for the type (guardian seat, round 2:
    -- four agent types on this estate carry TWO active rows with only the
    -- higher version loaded — a later edit could silently orphan half the
    -- workflow). <> not IS DISTINCT here: count(*) is never NULL.
    SELECT count(*) INTO active_count FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF active_count <> 1 THEN
        RAISE EXCEPTION '410: % active definition rows for required-fields-missing-handler — must be exactly 1 (the two-active-rows landmine)', active_count;
    END IF;

    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '410: required-fields-missing-handler not found after insert';
    END IF;
    IF cfg #>> '{workflow,start_step}' IS DISTINCT FROM 'classify' THEN
        RAISE EXCEPTION '410: start_step is not classify';
    END IF;
    IF jsonb_array_length(cfg #> '{workflow,steps,classify,config,params}') IS DISTINCT FROM 2 THEN
        RAISE EXCEPTION '410: classify must take exactly (site_id, work_item_id)';
    END IF;
    -- Every branch must name both arms, or an item silently stops mid-route.
    IF cfg #>> '{workflow,steps,route_stale,config,then_step}'    IS DISTINCT FROM 'close_stale'
    OR cfg #>> '{workflow,steps,route_stale,config,else_step}'    IS DISTINCT FROM 'route_resolved'
    OR cfg #>> '{workflow,steps,route_resolved,config,then_step}' IS DISTINCT FROM 'close_resolved'
    OR cfg #>> '{workflow,steps,route_resolved,config,else_step}' IS DISTINCT FROM 'route_owned'
    OR cfg #>> '{workflow,steps,route_owned,config,then_step}'    IS DISTINCT FROM 'park_owned'
    OR cfg #>> '{workflow,steps,route_owned,config,else_step}'    IS DISTINCT FROM 'route_blob'
    OR cfg #>> '{workflow,steps,route_blob,config,then_step}'     IS DISTINCT FROM 'park_blob'
    OR cfg #>> '{workflow,steps,route_blob,config,else_step}'     IS DISTINCT FROM 'route_asset'
    OR cfg #>> '{workflow,steps,route_asset,config,then_step}'    IS DISTINCT FROM 'park_asset'
    OR cfg #>> '{workflow,steps,route_asset,config,else_step}'    IS DISTINCT FROM 'route_noplan'
    OR cfg #>> '{workflow,steps,route_noplan,config,then_step}'   IS DISTINCT FROM 'file_recreate'
    OR cfg #>> '{workflow,steps,route_noplan,config,else_step}'   IS DISTINCT FROM 'route_unbuildable'
    OR cfg #>> '{workflow,steps,route_unbuildable,config,then_step}' IS DISTINCT FROM 'park_unbuildable'
    OR cfg #>> '{workflow,steps,route_unbuildable,config,else_step}' IS DISTINCT FROM 'route_partial'
    OR cfg #>> '{workflow,steps,route_partial,config,then_step}'  IS DISTINCT FROM 'file_rewrite'
    OR cfg #>> '{workflow,steps,route_partial,config,else_step}'  IS DISTINCT FROM 'mark_failed' THEN
        RAISE EXCEPTION '410: a conditional_branch arm is missing or mis-wired — an item would stop mid-route';
    END IF;
    -- The park arms must park, not complete — the whole anti-churn design.
    IF cfg #>> '{workflow,steps,park_owned,config,status}' IS DISTINCT FROM 'needs_human_review'
    OR cfg #>> '{workflow,steps,park_blob,config,status}'  IS DISTINCT FROM 'needs_human_review'
    OR cfg #>> '{workflow,steps,park_asset,config,status}' IS DISTINCT FROM 'needs_human_review'
    OR cfg #>> '{workflow,steps,park_unbuildable,config,status}' IS DISTINCT FROM 'needs_human_review' THEN
        RAISE EXCEPTION '410: a park arm does not park at needs_human_review — dedup-key churn returns';
    END IF;
    -- `spec` is RETIRED on create_work_item (bugs_open/234). Assert ABSENCE.
    IF cfg #> '{workflow,steps,file_recreate,config,spec}' IS NOT NULL
    OR cfg #> '{workflow,steps,file_rewrite,config,spec}'  IS NOT NULL THEN
        RAISE EXCEPTION '410: a create step carries the RETIRED create_work_item `spec` key (bugs_open/234) — use spec_paths/spec_literal';
    END IF;
    -- Conversions must be born dispatchable: triaged, never detected (bugs 083).
    IF cfg #>> '{workflow,steps,file_recreate,config,status}' IS DISTINCT FROM 'triaged'
    OR cfg #>> '{workflow,steps,file_rewrite,config,status}'  IS DISTINCT FROM 'triaged' THEN
        RAISE EXCEPTION '410: a conversion item is not born at triaged — it would strand (the detected promoter is off, bugs_open/083)';
    END IF;

    -- INERTNESS is STRUCTURAL: this file contains no UPDATE on site_work_items,
    -- so it cannot assign. (The first application asserted assigned=0; after the
    -- deliberate canary of 2026-08-15 that assert would wrongly abort a
    -- re-apply, so it is now a report.) Assignment remains the separate,
    -- canaried step in ASSIGN_2026-08-15_fleet_assignment.sql.
    SELECT count(*) INTO assigned FROM site_work_items
     WHERE handler_agent = 'required-fields-missing-handler';
    RAISE NOTICE '410: required-fields-missing-handler seeded; % work item(s) currently route to it (assignment happens only via the ASSIGN file)', assigned;
END $$;

COMMIT;

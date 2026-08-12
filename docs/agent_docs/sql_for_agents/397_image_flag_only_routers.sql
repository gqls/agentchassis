-- 397 — handlers for the two image item types that have never had one
--
-- OWNER INSTRUCTION 2026-08-12: "please create the handlers that are missing and
-- leave the assignment until the rung 2 is fixed."
--
-- ============================================================================
-- READ THIS FIRST: THE ABSENCE WAS DELIBERATE, AND THE CODE SAYS SO
-- ============================================================================
-- These two item types are not handler-less by oversight. Both checks set
-- HandlerAgent: "" on purpose, with the reason written beside it:
--
--   check_image_source_unsatisfiable.go:212
--     // Flag-only: no handler, and needs_human_review keeps it out
--     // of the dispatch loop while the dedup key holds it open.
--
--   check_image_url_404.go:274
--     // HandlerAgent intentionally empty — flag-only. Repairing a stale
--     // reference means removing or repointing it, which no image
--     // generator can decide. Generation is check_placeholder_image_in_use's
--     // remit, under its own precondition.
--
-- That reasoning is SOUND and this migration does not overturn it. What it
-- rejects is only the consequence: measured 2026-08-12, the two types have
-- accumulated 93 rows across 16 sites since 2026-07-17 and 2026-07-26
-- respectively, and NOT ONE has ever carried a handler or reached a terminal
-- state. 49 image_source_unsatisfiable sit at needs_human_review; 43 of 44
-- image_url_404 sit at blocked. "A human decides" is only a design if a human
-- is ever shown the decision, and nothing shows them.
--
-- So both agents below are ROUTERS, NOT REPAIRERS. Neither one edits a page,
-- generates an image, or repoints a reference. Each asks ONE deterministic
-- question, and then either:
--   (a) converts the finding into an item type that ALREADY has a working
--       handler (needs_imagery -> image-build-handler), or
--   (b) escalates via checkpoint_for_review with the facts attached, which is
--       the same "a human decides" outcome the checks intended, except that it
--       now actually reaches the human-review surface instead of resting on a
--       status nothing reads.
-- The judgement the check comments protect — remove vs repoint vs re-template —
-- is never made by these agents. Branch (b) exists precisely to refuse it.
--
-- ============================================================================
-- NOTHING IS ASSIGNED. THESE ROWS ARE INERT ON PURPOSE.
-- ============================================================================
-- No work item's handler_agent is set by this migration, and the two checks
-- still emit HandlerAgent: "" (a Go change, not shipped here). So no row routes
-- to either agent and neither will ever be dispatched until someone
-- deliberately assigns it. The final DO block ASSERTS that inertness rather
-- than asserting it in prose.
--
-- WHY ASSIGNMENT WAITS — bugs_open/248_…_under_a_placeholder_name.md ("248" is
-- an ambiguous number; resolve by slug). Both routers' (a) branch ends at
-- image-build-handler, whose deploy step still reaches
-- `asset-deployer.deploy_asset config.asset_key = "input_data.asset_key"` —
-- rung 2, verified STILL LIVE 2026-08-12. Anything routed there today
-- regenerates correctly and deploys to /assets/images/input-data.asset-key.jpg.
-- Assigning these handlers before rung 2 is cut would convert 93 inert findings
-- into 93 placeholder-named files. That is the whole reason the owner deferred
-- assignment, and it is the precondition for lifting it.
--
-- HOW TO ASSIGN, LATER (deliberately not automated here — a one-off operational
-- action on specific rows, not a repeatable config change; same reasoning as
-- 326_directory_build_handler_agent.sql):
--   UPDATE site_work_items SET handler_agent = 'image-url-404-handler',
--          status = 'triaged'
--    WHERE item_type = 'image_url_404' AND status = 'blocked';
-- Do NOT run that until rung 2 is cut AND the fix is pod-verified.
--
-- image_tag v1.0.1291 — the tag both image-build-handler and
-- directory-build-handler carry live today. Every action referenced below is
-- long-registered (query_database, conditional_branch, create_work_item,
-- checkpoint_for_review, update_work_item_status, complete_workflow), so this
-- config is not waiting on an image. NOTE create_work_item is StrictConfig with
-- `spec` RETIRED (bugs_open/234, migration 364): the spec is built from
-- spec_paths/spec_literal below, and a step carrying `spec` now fails
-- validation rather than silently filing '{}'.
--
-- ORDER: DB config, live the moment it commits — but inert, per above.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 397_….sql
-- ============================================================================

BEGIN;

-- ----------------------------------------------------------------------------
-- snapshot_agent — CONDITIONALLY, and here is why it is not the bare opening
-- line the runner's README prescribes.
--
-- `scripts/migration/README.md` says every migration touching agent_definitions
-- opens with `SELECT snapshot_agent('<type>', '<file>: pre-update');`. Both
-- overloads — snapshot_agent(text) and snapshot_agent(text,text) — RAISE
-- 'No active definition found for type %' when the type does not exist yet, so
-- on a migration that CREATES types that opening line aborts the file on its
-- first run, every time. The rule is written for migrations that MUTATE an
-- existing agent, which is the overwhelming majority.
--
-- The rule's intent — never overwrite an agent row without a before-image —
-- still applies here, because the INSERTs below carry ON CONFLICT DO UPDATE and
-- so DO mutate on any re-run. Hence: snapshot when there is something to
-- snapshot, skip when there is not. On first application this is a no-op; on
-- every subsequent run it behaves exactly as the README intends.
-- ----------------------------------------------------------------------------
DO $$
DECLARE
    t text;
BEGIN
    FOREACH t IN ARRAY ARRAY['image-url-404-handler', 'image-source-unsatisfiable-handler'] LOOP
        IF EXISTS (SELECT 1 FROM agent_definitions
                    WHERE type = t AND is_active
                      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) THEN
            PERFORM snapshot_agent(t, '397_image_flag_only_routers.sql: pre-update');
            RAISE NOTICE '397: snapshotted existing % before update', t;
        ELSE
            RAISE NOTICE '397: % does not exist yet — nothing to snapshot (first application)', t;
        END IF;
    END LOOP;
END $$;

-- ----------------------------------------------------------------------------
-- 1. image-url-404-handler  (item_type: image_url_404)
--
-- The one deterministic question: does this site hold an active asset for the
-- purpose the referenced filename implies?
--   no  -> nothing was ever generated for it; that IS a generation request, and
--          needs_imagery/image-build-handler already owns generation.
--   yes -> an asset exists and the page's path still does not resolve, which is
--          a DEPLOY-PATH disagreement between writer and reader, not a missing
--          image. That is bugs_open/248's territory and no generator can decide
--          it — escalate with both paths named.
-- ----------------------------------------------------------------------------
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, is_active, status,
    default_config, topics, env_vars, idle_timeout_seconds
) VALUES (
    'image-url-404-handler',
    'Image URL 404 Router',
    'Routes an image_url_404 finding: files a needs_imagery generation request when the site holds no asset for the implied purpose, and escalates to human review when an asset exists but the referenced path still does not resolve (a deploy-path disagreement — bugs_open/248). Never removes or repoints a reference itself.',
    'site',
    'docker.io/aqls/agent-chassis', 'v1.0.1291',
    true, 'active',
    '{
        "workflow": {
            "start_step": "load_asset_facts",
            "steps": {
                "load_asset_facts": {
                    "action": "query_database",
                    "config": {
                        "output_format": "object",
                        "query": "SELECT CASE WHEN count(*) > 0 THEN ''has_asset'' ELSE ''no_asset'' END AS backing, count(*) AS asset_count FROM assets WHERE site_id = $1::uuid AND COALESCE(status, ''active'') <> ''deleted'' AND lower(replace(COALESCE(purpose, ''''), ''-'', ''_'')) = lower(replace(split_part($2, ''.'', 1), ''-'', ''_''))",
                        "params": ["input_data.site_id", "input_data.spec.filename"]
                    },
                    "next_step": "decide_route",
                    "error_step": "mark_work_item_failed",
                    "output_field": "asset_facts",
                    "description": "Does an active asset exist for the purpose this filename implies? hero.jpg -> hero, og-card.png -> og_card."
                },
                "decide_route": {
                    "action": "conditional_branch",
                    "config": {
                        "condition": "asset_facts.backing == no_asset",
                        "then_step": "file_generation_request",
                        "else_step": "escalate_deploy_path_mismatch"
                    },
                    "error_step": "mark_work_item_failed",
                    "output_field": "route_decision",
                    "description": "No asset at all is a generation request; an asset that exists while the path 404s is a deploy-path disagreement"
                },
                "file_generation_request": {
                    "action": "create_work_item",
                    "config": {
                        "item_type": "needs_imagery",
                        "handler_agent": "image-build-handler",
                        "item_pipeline": "build",
                        "severity": "medium",
                        "source": "image-url-404-handler",
                        "status": "detected",
                        "priority": 65,
                        "summary": "Referenced image path has no backing asset — generation requested by image-url-404-handler",
                        "item_key_prefix": "needs_imagery:from_404:",
                        "item_key_suffix_field": "input_data.spec.filename",
                        "spec_paths": {
                            "referenced_path": "input_data.spec.path",
                            "filename": "input_data.spec.filename",
                            "surface": "input_data.spec.surface"
                        },
                        "spec_literal": {
                            "kind": "hero",
                            "origin_check": "image_url_404",
                            "reason": "no active asset exists for the purpose this filename implies"
                        }
                    },
                    "next_step": "close_complete",
                    "error_step": "mark_work_item_failed",
                    "output_field": "generation_request",
                    "description": "Hand generation to the type that already owns it, rather than inventing a second generation path"
                },
                "escalate_deploy_path_mismatch": {
                    "action": "checkpoint_for_review",
                    "config": {
                        "site_id_from": "input_data.site_id",
                        "item_type": "needs_human_review",
                        "severity": "medium",
                        "summary_from": "input_data.summary",
                        "review_fields_from": "input_data.spec",
                        "save_spec_aspect": "image_url_404_deploy_mismatch"
                    },
                    "next_step": "close_complete",
                    "error_step": "mark_work_item_failed",
                    "output_field": "escalation",
                    "description": "An asset exists and the page still 404s: writer and reader disagree on the path (bugs_open/248). Removing vs repointing is the judgement this agent must NOT make."
                },
                "close_complete": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "complete",
                        "skip_if_missing": true,
                        "work_item_id_field": "input_data.work_item_id"
                    },
                    "next_step": "done",
                    "description": "Close the originating flag-only item — it has now been routed or escalated"
                },
                "mark_work_item_failed": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "failed",
                        "skip_if_missing": true,
                        "work_item_id_field": "input_data.work_item_id"
                    },
                    "next_step": "done",
                    "description": "Fail loudly rather than closing an item nothing acted on"
                },
                "done": {
                    "action": "complete_workflow",
                    "config": { "output_fields": ["asset_facts", "route_decision", "generation_request", "escalation"] },
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
-- 2. image-source-unsatisfiable-handler  (item_type: image_source_unsatisfiable)
--
-- The finding: a component field sources site_assets.<path> and nothing in the
-- estate generates that path. The one deterministic question: does the trailing
-- segment of that source name a purpose the imagery pipeline actually knows?
--   yes -> the source is legitimate and simply unfulfilled; file needs_imagery.
--   no  -> the component names something no purpose can ever supply, so the
--          repair is to change the COMPONENT TEMPLATE. That is a human/template
--          decision by construction — escalate, never guess.
--
-- The purpose vocabulary below is deliberately a literal list rather than a
-- join: platform/orchestration/imageryplan/imageryplan.go holds the aliases in
-- Go, not in a table, so a join would invent an authority that does not exist.
-- If that list moves, this one must be updated with it — stated here because a
-- silently-stale vocabulary would send mappable sources to human review, which
-- is the safe direction but still wrong.
-- ----------------------------------------------------------------------------
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, is_active, status,
    default_config, topics, env_vars, idle_timeout_seconds
) VALUES (
    'image-source-unsatisfiable-handler',
    'Unsatisfiable Image Source Router',
    'Routes an image_source_unsatisfiable finding: files a needs_imagery request when the component''s source names a purpose the imagery pipeline knows, and escalates to human review when it names one nothing can ever supply (the repair being a template change). Never edits a component.',
    'site',
    'docker.io/aqls/agent-chassis', 'v1.0.1291',
    true, 'active',
    '{
        "workflow": {
            "start_step": "classify_source",
            "steps": {
                "classify_source": {
                    "action": "query_database",
                    "config": {
                        "output_format": "object",
                        "query": "SELECT CASE WHEN lower(replace(split_part($1, ''.'', array_length(string_to_array($1, ''.''), 1)), ''-'', ''_'')) IN (''hero'', ''logo'', ''icon'', ''og_card'', ''favicon'', ''illustration'', ''infographic'', ''banner'', ''background'', ''thumbnail'', ''portrait'') THEN ''mappable'' ELSE ''unmappable'' END AS routing, lower(replace(split_part($1, ''.'', array_length(string_to_array($1, ''.''), 1)), ''-'', ''_'')) AS implied_purpose",
                        "params": ["input_data.spec.source"]
                    },
                    "next_step": "decide_route",
                    "error_step": "mark_work_item_failed",
                    "output_field": "source_facts",
                    "description": "Does the trailing segment of the source name a purpose the imagery pipeline knows?"
                },
                "decide_route": {
                    "action": "conditional_branch",
                    "config": {
                        "condition": "source_facts.routing == mappable",
                        "then_step": "file_imagery_request",
                        "else_step": "escalate_unmappable_source"
                    },
                    "error_step": "mark_work_item_failed",
                    "output_field": "route_decision",
                    "description": "A known purpose is merely unfulfilled; an unknown one is a template defect"
                },
                "file_imagery_request": {
                    "action": "create_work_item",
                    "config": {
                        "item_type": "needs_imagery",
                        "handler_agent": "image-build-handler",
                        "item_pipeline": "build",
                        "severity": "medium",
                        "source": "image-source-unsatisfiable-handler",
                        "status": "detected",
                        "priority": 80,
                        "summary": "Component sources an image purpose nothing has generated — generation requested by image-source-unsatisfiable-handler",
                        "item_key_prefix": "needs_imagery:from_unsatisfiable:",
                        "item_key_suffix_field": "input_data.spec.source",
                        "spec_paths": {
                            "page": "input_data.spec.page",
                            "component_function": "input_data.spec.component_function",
                            "field": "input_data.spec.field",
                            "source": "input_data.spec.source",
                            "purpose": "source_facts.implied_purpose"
                        },
                        "spec_literal": {
                            "origin_check": "image_source_unsatisfiable",
                            "reason": "source names a known imagery purpose that nothing has generated yet"
                        }
                    },
                    "next_step": "close_complete",
                    "error_step": "mark_work_item_failed",
                    "output_field": "imagery_request",
                    "description": "Reuse the generation path rather than building a second one"
                },
                "escalate_unmappable_source": {
                    "action": "checkpoint_for_review",
                    "config": {
                        "site_id_from": "input_data.site_id",
                        "item_type": "needs_human_review",
                        "severity": "medium",
                        "summary_from": "input_data.summary",
                        "review_fields_from": "input_data.spec",
                        "save_spec_aspect": "image_source_unsatisfiable_template"
                    },
                    "next_step": "close_complete",
                    "error_step": "mark_work_item_failed",
                    "output_field": "escalation",
                    "description": "No purpose can supply this source, so the component template is what must change — not something an image agent may decide"
                },
                "close_complete": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "complete",
                        "skip_if_missing": true,
                        "work_item_id_field": "input_data.work_item_id"
                    },
                    "next_step": "done",
                    "description": "Close the originating flag-only item — it has now been routed or escalated"
                },
                "mark_work_item_failed": {
                    "action": "update_work_item_status",
                    "config": {
                        "status": "failed",
                        "skip_if_missing": true,
                        "work_item_id_field": "input_data.work_item_id"
                    },
                    "next_step": "done",
                    "description": "Fail loudly rather than closing an item nothing acted on"
                },
                "done": {
                    "action": "complete_workflow",
                    "config": { "output_fields": ["source_facts", "route_decision", "imagery_request", "escalation"] },
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
-- Verification.
--
-- Every check below uses IS DISTINCT FROM, never <>. A `#>>` on a MISSING jsonb
-- path returns NULL, and `NULL <> 'expected'` is NULL rather than TRUE — so an
-- IF on that never fires and a wrong or absent key sits GREEN for ever
-- (LANDMINES: "A migration verify block comparing a jsonb path with `<>` sits
-- GREEN for ever when the key does not exist"). RAISE EXCEPTION, not a bare
-- SELECT: ON_ERROR_STOP ignores a non-empty result set, so a verify block built
-- from SELECTs cannot stop the COMMIT.
-- ----------------------------------------------------------------------------
DO $$
DECLARE
    cfg      jsonb;
    assigned integer;
BEGIN
    -- --- agent 1 -----------------------------------------------------------
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'image-url-404-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '397: image-url-404-handler not found after insert';
    END IF;
    IF cfg #>> '{workflow,start_step}' IS DISTINCT FROM 'load_asset_facts' THEN
        RAISE EXCEPTION '397: image-url-404-handler start_step is not load_asset_facts';
    END IF;
    IF cfg #>> '{workflow,steps,decide_route,config,then_step}' IS DISTINCT FROM 'file_generation_request' THEN
        RAISE EXCEPTION '397: image-url-404-handler then_step missing — the router would not route';
    END IF;
    IF cfg #>> '{workflow,steps,decide_route,config,else_step}' IS DISTINCT FROM 'escalate_deploy_path_mismatch' THEN
        RAISE EXCEPTION '397: image-url-404-handler else_step missing — findings would be silently dropped instead of escalated';
    END IF;
    -- `spec` is RETIRED on create_work_item and fails validation (bugs_open/234).
    -- Assert its ABSENCE, because its presence would file empty specs that read
    -- as success — the exact failure that bug records.
    IF cfg #> '{workflow,steps,file_generation_request,config,spec}' IS NOT NULL THEN
        RAISE EXCEPTION '397: image-url-404-handler carries the RETIRED create_work_item `spec` key (bugs_open/234) — use spec_paths/spec_literal';
    END IF;

    -- --- agent 2 -----------------------------------------------------------
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'image-source-unsatisfiable-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg IS NULL THEN
        RAISE EXCEPTION '397: image-source-unsatisfiable-handler not found after insert';
    END IF;
    IF cfg #>> '{workflow,start_step}' IS DISTINCT FROM 'classify_source' THEN
        RAISE EXCEPTION '397: image-source-unsatisfiable-handler start_step is not classify_source';
    END IF;
    IF cfg #>> '{workflow,steps,decide_route,config,else_step}' IS DISTINCT FROM 'escalate_unmappable_source' THEN
        RAISE EXCEPTION '397: image-source-unsatisfiable-handler else_step missing — an unmappable source would vanish';
    END IF;
    IF cfg #> '{workflow,steps,file_imagery_request,config,spec}' IS NOT NULL THEN
        RAISE EXCEPTION '397: image-source-unsatisfiable-handler carries the RETIRED create_work_item `spec` key (bugs_open/234)';
    END IF;

    -- --- INERTNESS, asserted rather than promised ---------------------------
    -- The owner deferred assignment until bugs_open/248's rung 2 is cut. If this
    -- migration has routed anything to either agent, that instruction has been
    -- broken and the transaction must not commit.
    SELECT count(*) INTO assigned FROM site_work_items
     WHERE handler_agent IN ('image-url-404-handler', 'image-source-unsatisfiable-handler');
    IF assigned <> 0 THEN
        RAISE EXCEPTION '397: % work item(s) already route to the new handlers — assignment was deferred until bugs_open/248 rung 2 is cut', assigned;
    END IF;

    RAISE NOTICE '397: both routers seeded and INERT (0 work items assigned, as instructed)';
END $$;

COMMIT;

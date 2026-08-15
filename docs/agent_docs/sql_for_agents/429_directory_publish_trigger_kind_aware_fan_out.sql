-- 429 - directory publish leg goes kind-aware: the trigger fans out per
-- (site, kind) pair and the publisher publishes exactly the ONE kind it is
-- asked for. Phase B3c of the portfolio_positioning lane
-- (PLAN_2026-08-12_fleet_buildout.md; the defect inventory is
-- FINDING_2026-08-10_the_tracker_publisher_was_reverted_and_never_re_extended.md
-- "latent defect", handed to B3c by that FINDING's 2026-08-15 resolution note).
--
-- WHAT WAS WRONG (all verified on the live rows, 2026-08-15):
--   1. find_directory_sites gated opt-in on content_features.model_directory
--      ONLY - a site opted into any other kind could never be found, and a
--      site opted into model_directory alone would receive every kind the
--      publisher chain hard-coded (adoption + protocol tracker files it
--      never asked for).
--   2. The deployed-component predicate named only the model-directory
--      components - same blindness, second predicate.
--   3. The has-claims gate was EXISTS over directory_claims with no kind at
--      all - any one kind's claims unblocked publishing of every kind.
--   4. ORDER BY s.domain LIMIT 5 - deterministic alphabetical starvation:
--      site #6 would NEVER publish, every cycle, forever, silently.
--   5. The publisher chain was a hard-coded model->company->protocol run,
--      so the three Phase B finance kinds (mortgage-lender,
--      savings-provider, health-insurer - profiles live in the binary since
--      v1.0.1301) were structurally unpublishable.
--
-- THE SHAPE (and why it is this shape):
--   - The trigger's query returns one row per DUE (site, kind): the site
--     opts into that kind's spec key, carries a deployed page with that
--     kind's component, and the kind has publishable claims (is_current,
--     status='found', on status='active' entities - mirrored from
--     QueryDirectoryEntries, queryresolve/directory_items.go, which is what
--     the renderer will actually serve).
--   - The kind->spec_key->components mapping is a VALUES list kept in
--     LOCKSTEP with Go's directoryPublishProfiles
--     (render_directory_action.go) + the SpecKeys in
--     feed_directory_recommendation_action.go. A kind added in Go without a
--     row here silently never publishes - registered as the residual in
--     concept register DIR-001 and LANDMINES.
--   - ORDER BY random() under the LIMIT: every due pair gets a positive
--     probability each cycle, so no pair can be starved DETERMINISTICALLY.
--     True staleness-ordering needs a publish stamp the schema does not
--     have; recorded as future work, not smuggled in here.
--   - The publisher becomes ONE render->commit pair parameterised by
--     input_data.kind. NOT a loop inside the publisher: loop iterations
--     suffix output fields (_0, _1, ...) and the coordinator rewrites only
--     a whitelist of config keys that does NOT include git_commit's
--     files_field (coordinator.go prefixConfigStepReferences dataRefKeys),
--     so a render->commit pair inside a loop would read a field that no
--     longer exists. The fan-out therefore lives in the TRIGGER's existing,
--     proven spawn+call loop - one publisher run per (site, kind).
--   - "kind": "input_data.kind" resolves via ExtractActionInputs Strategy 0
--     (kind is declared Optional in RenderModelDirectoryInputSpec); the
--     closed-set literal override ignores it because the string is not a
--     profile name. Pinned by TestDirectoryKindResolvesFromLiteralStepConfig
--     ("reference that resolved" case). The 2026-07-26 silent-model-default
--     trap is closed UPSTREAM of the action: "kind" joins the publisher's
--     input_contract.required, and call_agent validates required contract
--     fields against the resolved input_mapping and FAILS the call when one
--     is missing (call_agent.go ValidateInputContract) - loud, not model.
--     (The `kind!` strict marker would be tighter still, but the running
--     binary v1.0.1302 was built before RFC_029's strict-marker commit
--     landed 2026-08-15 14:07Z - config must not outrun the binary.)
--   - Per-kind commit messages ride the trigger rows and reach git_commit
--     via commit_message_field (in the binary since 2026-08-05), template
--     fallback preserved. The historical messages ("Update model directory",
--     "Update adoption tracker", "Update protocol tracker") are kept
--     verbatim so the site repos' history stays continuous.
--   - No error_step on the commit step: migration 427's council-reviewed
--     posture (a commit failure fails the run loudly) carries over. Failure
--     isolation now comes from the fan-out itself - one (site, kind) run
--     failing cannot touch any other pair; the trigger loop's
--     continue_on_error covers the rest, and the next 6h cycle self-heals.
--
-- TRANSITION SAFETY: no in-flight orchestrations for either type at
-- authoring time (12x COMPLETED, nothing open; re-check before applying). An
-- old-config trigger run in flight would call the new publisher without
-- "kind": contract validation fails that call loudly, the loop continues,
-- the next cycle runs fully new. In-flight runs carry their own workflow
-- snapshot, so nothing mid-run can be half-updated (LANDMINES: live-config
-- edits cannot reach an in-flight orchestration).
--
-- Config-only: live immediately, no image roll. Both updates in ONE
-- transaction so no cycle can see a kind-aware trigger with a kind-blind
-- publisher or vice versa.

SELECT snapshot_agent('model-directory-publisher', '429_directory_publish_trigger_kind_aware_fan_out.sql: pre-update');
SELECT snapshot_agent('model-directory-trigger',   '429_directory_publish_trigger_kind_aware_fan_out.sql: pre-update');

BEGIN;

-- ── Pre-flight guards ──────────────────────────────────────────────────────
DO $do$
DECLARE
    n int;
    pub_steps jsonb;
    trg_steps jsonb;
    trg_query text;
    im jsonb;
BEGIN
    -- Dual-active-row guards.
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '429: model-directory-publisher does not have exactly one active row (found %)', n;
    END IF;
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'model-directory-trigger' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '429: model-directory-trigger does not have exactly one active row (found %)', n;
    END IF;

    -- Both snapshots from the two-arg snapshot_agent above must exist.
    SELECT count(*) INTO n FROM agent_definitions_backup
    WHERE snapshot_reason = '429_directory_publish_trigger_kind_aware_fan_out.sql: pre-update'
      AND type IN ('model-directory-publisher','model-directory-trigger');
    IF n <> 2 THEN
        RAISE EXCEPTION '429: expected 2 pre-update backup rows, found % - snapshot_agent did not run', n;
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO pub_steps FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- Idempotency: refuse a second application.
    IF pub_steps ? 'render_directory_json' THEN
        RAISE EXCEPTION '429: already applied - publisher carries render_directory_json';
    END IF;

    -- Drift anchors: the publisher must still be exactly the post-411/427
    -- 7-step chain (411 re-extended it; 427 removed the commit error_steps).
    SELECT count(*) INTO n FROM jsonb_object_keys(pub_steps);
    IF n <> 7 THEN
        RAISE EXCEPTION '429: expected the post-411 7-step publisher, found % steps - re-check before applying', n;
    END IF;
    IF NOT (pub_steps ? 'render_model_directory_json' AND pub_steps ? 'render_adoption_json'
            AND pub_steps ? 'render_protocol_json' AND pub_steps ? 'commit_model_directory') THEN
        RAISE EXCEPTION '429: publisher step names do not match the post-411 chain - re-check before applying';
    END IF;
    IF pub_steps#>'{commit_model_directory}' ? 'error_step' THEN
        RAISE EXCEPTION '429: commit_model_directory still carries error_step - 427 not applied? re-check before applying';
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO trg_steps FROM agent_definitions
    WHERE type = 'model-directory-trigger' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    trg_query := trg_steps#>>'{find_directory_sites,config,query}';

    -- Trigger drift anchors: the three kind-blind predicates + LIMIT 5 must
    -- all still be present (they are WHY this migration exists).
    IF trg_query IS NULL THEN
        RAISE EXCEPTION '429: trigger find_directory_sites query not found at the expected path';
    END IF;
    IF position('mortgage_lender_directory' IN trg_query) > 0 THEN
        RAISE EXCEPTION '429: already applied - trigger query already names the finance spec keys';
    END IF;
    IF position($a$'model_directory'->>'recommended'$a$ IN trg_query) = 0 THEN
        RAISE EXCEPTION '429: trigger query opt-in anchor missing - the row has drifted, re-check before applying';
    END IF;
    IF position($a$cc.function IN ('model-directory', 'model-directory-listing')$a$ IN trg_query) = 0 THEN
        RAISE EXCEPTION '429: trigger query component anchor missing - the row has drifted, re-check before applying';
    END IF;
    IF position('LIMIT 5' IN trg_query) = 0 THEN
        RAISE EXCEPTION '429: trigger query LIMIT 5 anchor missing - the row has drifted, re-check before applying';
    END IF;
    IF (trg_steps#>>'{process_sites,config,max_iterations}') IS DISTINCT FROM '5' THEN
        RAISE EXCEPTION '429: trigger loop max_iterations is not 5 - the row has drifted, re-check before applying';
    END IF;
    im := trg_steps#>'{process_sites,config,sub_workflow,steps,call_publisher,config,input_mapping}';
    IF im IS NULL OR (SELECT count(*) FROM jsonb_object_keys(im)) <> 2 THEN
        RAISE EXCEPTION '429: trigger call_publisher input_mapping is not the expected {site_id, domain} - re-check before applying';
    END IF;
END $do$;

-- ── 1. Publisher: one render->commit pair, kind from input, contract-enforced ─
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        $json$
        {
            "start_step": "render_directory_json",
            "processing_mode": "orchestrator",
            "timeout_seconds": 600,
            "steps": {
                "render_directory_json": {
                    "action": "render_directory",
                    "config": {"site_id": "input_data.site_id", "kind": "input_data.kind"},
                    "next_step": "commit_directory",
                    "description": "Build the requested kind's data JSON (+ full listing when a listing page exists) from the global registry; kind arrives from the trigger row and is required by the input contract, so a missing kind fails the call rather than defaulting to model",
                    "output_field": "directory_render_result"
                },
                "commit_directory": {
                    "action": "git_commit",
                    "config": {
                        "files_field": "directory_render_result.files",
                        "domain_field": "directory_render_result.domain",
                        "commit_message_field": "input_data.commit_message",
                        "commit_message": "Update directory data: {{.domain}}"
                    },
                    "next_step": "complete",
                    "description": "Commit the JSON files into the site repo via git-adapter; per-kind message from the trigger via commit_message_field, template fallback if absent. No error_step (427 posture): a commit failure fails this run loudly - isolation now comes from the per-(site,kind) fan-out",
                    "output_field": "directory_commit_result"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {"output_fields": ["directory_render_result", "directory_commit_result"]}
                }
            }
        }
        $json$::jsonb
    ),
    input_contract = '{"required": ["site_id", "domain", "kind"]}'::jsonb,
    description = 'Per-(site, kind) publish leg of the directory pipeline: renders ONE register kind (model, company, protocol, mortgage-lender, savings-provider, health-insurer) to its data/*.json files for one site, commits via git-adapter, and queues scoped page rerenders. The kind is required by the input contract - a call without it fails validation rather than silently defaulting to model. Kind-aware since migration 429 (Phase B3c).',
    updated_at = NOW()
WHERE type = 'model-directory-publisher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 2. Trigger: due (site, kind) pairs, kind-aware on all three predicates ──
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        $json$
        {
            "start_step": "find_directory_sites",
            "processing_mode": "orchestrator",
            "timeout_seconds": 900,
            "steps": {
                "find_directory_sites": {
                    "action": "query_database",
                    "config": {
                        "query": "SELECT site_id, domain, kind, commit_message FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, m.kind, m.commit_message FROM (VALUES ('model','model_directory','model-directory','model-directory-listing','Update model directory'), ('company','adoption_tracker','adoption-tracker','adoption-tracker-listing','Update adoption tracker'), ('protocol','protocol_tracker','protocol-tracker','protocol-tracker-listing','Update protocol tracker'), ('mortgage-lender','mortgage_lender_directory','mortgage-lender-directory','mortgage-lender-directory-listing','Update mortgage-lender directory'), ('savings-provider','savings_provider_directory','savings-provider-directory','savings-provider-directory-listing','Update savings-provider directory'), ('health-insurer','health_insurer_directory','health-insurer-directory','health-insurer-directory-listing','Update health-insurer directory')) AS m(kind, spec_key, snippet_component, listing_component, commit_message) CROSS JOIN sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true WHERE (ss.data->'content_features'->m.spec_key->>'recommended')::boolean = true AND EXISTS (SELECT 1 FROM pages p JOIN page_components pc ON pc.page_id = p.id JOIN content_components cc ON cc.id = pc.component_id WHERE p.site_id = s.id AND p.build_status = 'deployed' AND cc.function IN (m.snippet_component, m.listing_component)) AND EXISTS (SELECT 1 FROM directory_claims dc JOIN directory_entities de ON de.id = dc.entity_id WHERE de.kind = m.kind AND de.status = 'active' AND dc.is_current AND dc.status = 'found')) due ORDER BY random() LIMIT 12",
                        "output_format": "object"
                    },
                    "next_step": "check_has_sites",
                    "description": "Due (site, kind) pairs: the site opts into the kind's spec key, carries a deployed page with that kind's component, and the kind has publishable claims on active entities (mirrors QueryDirectoryEntries). The VALUES mapping is in lockstep with Go's directoryPublishProfiles - a kind added in Go needs a row here or it never publishes. ORDER BY random() under the LIMIT so no pair can be starved deterministically",
                    "output_field": "directory_sites"
                },
                "check_has_sites": {
                    "action": "evaluate_condition",
                    "config": {
                        "condition_field": "directory_sites.count",
                        "conditions": {"0": "notify_scheduler_idle"},
                        "default": "process_sites"
                    },
                    "description": "Skip if no (site, kind) pairs are due"
                },
                "process_sites": {
                    "action": "loop",
                    "config": {
                        "items_field": "directory_sites.rows",
                        "item_variable": "current_site",
                        "max_iterations": 12,
                        "continue_on_error": true,
                        "sub_workflow": {
                            "start_step": "spawn_publisher",
                            "steps": {
                                "spawn_publisher": {
                                    "action": "spawn_agent",
                                    "config": {"role": "directory_publisher", "agent_type": "model-directory-publisher"},
                                    "next_step": "call_publisher",
                                    "output_field": "publisher_spawned",
                                    "description": "Spawn model-directory-publisher for this (site, kind) pair"
                                },
                                "call_publisher": {
                                    "action": "call_agent",
                                    "config": {
                                        "target_role": "directory_publisher",
                                        "input_mapping": {"site_id": "current_site.site_id", "domain": "current_site.domain", "kind": "current_site.kind", "commit_message": "current_site.commit_message"},
                                        "timeout_seconds": 600
                                    },
                                    "next_step": "done",
                                    "error_step": "done",
                                    "output_field": "publish_result",
                                    "description": "Publish this kind for this site; the publisher's input contract requires kind, so a mapping failure is loud"
                                },
                                "done": {"action": "loop_complete", "description": "Next (site, kind) pair"}
                            }
                        }
                    },
                    "next_step": "notify_scheduler",
                    "output_field": "publish_results",
                    "description": "One publisher run per due (site, kind) pair; a failed pair cannot touch any other pair"
                },
                "notify_scheduler": {
                    "action": "query_database",
                    "config": {"query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'model-directory-publish'", "output_format": "object"},
                    "next_step": "complete",
                    "output_field": "scheduler_notified",
                    "description": "Stamp completion"
                },
                "notify_scheduler_idle": {
                    "action": "query_database",
                    "config": {"query": "UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'model-directory-publish'", "output_format": "object"},
                    "next_step": "complete_idle",
                    "output_field": "scheduler_notified",
                    "description": "Stamp completion (idle)"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {"output_fields": ["directory_sites", "publish_results"]}
                },
                "complete_idle": {
                    "action": "complete_workflow",
                    "config": {"output_fields": ["directory_sites"], "success_message": "No (site, kind) pairs due for directory publish"}
                }
            }
        }
        $json$::jsonb
    ),
    description = 'Scheduled fan-out for the directory publish leg: finds due (site, kind) pairs - site opted in via site_specs classification content_features.<spec_key>, a deployed page carries that kind''s component, and the kind has publishable claims on active entities - then spawn+calls model-directory-publisher once per pair. Kind-aware since migration 429 (Phase B3c); mirrors content-feed-trigger.',
    updated_at = NOW()
WHERE type = 'model-directory-trigger' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── Verify with DO/RAISE - a verify block of SELECTs cannot stop the COMMIT ─
DO $do$
DECLARE
    pub agent_definitions%ROWTYPE;
    pub_steps jsonb;
    trg_steps jsonb;
    trg_query text;
    im jsonb;
    n int;
BEGIN
    SELECT * INTO pub FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    pub_steps := pub.default_config#>'{workflow,steps}';

    SELECT count(*) INTO n FROM jsonb_object_keys(pub_steps);
    IF n <> 3 THEN
        RAISE EXCEPTION '429 verify: publisher expected 3 steps, got %', n;
    END IF;
    -- IS DISTINCT FROM throughout: a missing path yields NULL, and
    -- `NULL <> x` is NULL, which IF treats as false - a verify built on
    -- plain <> could never fire on the very corruption it exists to catch.
    IF pub_steps#>>'{render_directory_json,action}' IS DISTINCT FROM 'render_directory' THEN
        RAISE EXCEPTION '429 verify: render_directory_json action wrong';
    END IF;
    IF pub_steps#>>'{render_directory_json,config,kind}' IS DISTINCT FROM 'input_data.kind' THEN
        RAISE EXCEPTION '429 verify: publisher kind is not the input_data.kind reference';
    END IF;
    IF pub_steps#>>'{commit_directory,config,commit_message_field}' IS DISTINCT FROM 'input_data.commit_message' THEN
        RAISE EXCEPTION '429 verify: commit_message_field missing';
    END IF;
    IF pub_steps#>'{commit_directory}' ? 'error_step' THEN
        RAISE EXCEPTION '429 verify: commit_directory must not carry error_step (427 posture)';
    END IF;
    IF NOT (pub.input_contract->'required' @> '"kind"'::jsonb) THEN
        RAISE EXCEPTION '429 verify: publisher input_contract does not require kind';
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO trg_steps FROM agent_definitions
    WHERE type = 'model-directory-trigger' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    trg_query := trg_steps#>>'{find_directory_sites,config,query}';

    IF trg_query IS NULL THEN
        RAISE EXCEPTION '429 verify: trigger query not found at the expected path';
    END IF;
    IF position('mortgage_lender_directory' IN trg_query) = 0
       OR position('savings_provider_directory' IN trg_query) = 0
       OR position('health_insurer_directory' IN trg_query) = 0
       OR position('adoption_tracker' IN trg_query) = 0
       OR position('protocol_tracker' IN trg_query) = 0
       OR position('model_directory' IN trg_query) = 0 THEN
        RAISE EXCEPTION '429 verify: trigger query does not name all six spec keys';
    END IF;
    IF position('de.kind = m.kind' IN trg_query) = 0 THEN
        RAISE EXCEPTION '429 verify: per-kind claims predicate missing';
    END IF;
    IF position('LIMIT 5' IN trg_query) > 0 THEN
        RAISE EXCEPTION '429 verify: LIMIT 5 survived';
    END IF;
    IF (trg_steps#>>'{process_sites,config,max_iterations}') IS DISTINCT FROM '12' THEN
        RAISE EXCEPTION '429 verify: loop max_iterations is not 12';
    END IF;
    im := trg_steps#>'{process_sites,config,sub_workflow,steps,call_publisher,config,input_mapping}';
    IF im->>'kind' IS DISTINCT FROM 'current_site.kind'
       OR im->>'commit_message' IS DISTINCT FROM 'current_site.commit_message' THEN
        RAISE EXCEPTION '429 verify: input_mapping does not carry kind + commit_message';
    END IF;

    -- The query string must be EXECUTABLE SQL, not just present - a JSON
    -- escaping mistake would otherwise fail only at the first live run.
    EXECUTE 'EXPLAIN ' || trg_query;
END $do$;

COMMIT;

-- Post-apply (hand-run, not part of this migration):
--   1. Force-trigger: UPDATE scheduled_tasks SET last_triggered_at = NULL
--      WHERE name = 'model-directory-publish';
--   2. Verify by the FILES per kind, never the statuses - one publisher
--      orchestration per (site, kind), each with its own
--      directory_render_result whose entity_count must DIFFER per kind
--      (identical counts across kinds = the 07-26 defect reproduced):
--      SELECT collected_data->'input_data'->>'kind' AS kind,
--             collected_data->'directory_render_result'->>'entity_count',
--             collected_data->'directory_render_result'->'files'
--      FROM orchestration_states
--      WHERE owner_agent_type = 'model-directory-publisher'
--      ORDER BY created_at DESC LIMIT 6;
--   3. Finance kinds must produce NO rows until a finance site deploys a
--      component (self-gating; the register has claims but no opted-in site).
--
-- ROLLBACK: 429_directory_publish_trigger_kind_aware_fan_out_ROLLBACK.sql
-- (restores default_config, input_contract and description for both rows
-- from the two agent_definitions_backup rows this file took above).

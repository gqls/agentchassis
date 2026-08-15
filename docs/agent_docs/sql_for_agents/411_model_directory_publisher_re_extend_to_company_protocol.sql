-- 411 - model-directory-publisher: re-extend the reverted chain back to
-- three kinds (model, company, protocol) so the adoption/protocol tracker
-- JSON feeds stop 404ing.
--
-- Why: FINDING_2026-08-10_the_tracker_publisher_was_reverted_and_never_re_extended.md
-- (docs/agent_docs/docs024_key_docs_latest/model_directory_pipeline/). On
-- 2026-07-26 the publisher was hand-extended to a 7-step chain, hit the
-- string-config-is-a-reference trap (Go default 'model' won silently on all
-- three "kind" steps, so the chain published the model register three times
-- under adoption/protocol commit messages), and was reverted to the
-- model-only chain from its own v2 snapshot THE SAME HOUR. The Go fix
-- (render_directory_action.go:181-186, commit bb99df77a) — read `kind` from
-- the step's own config when the value is in the closed profile set, so a
-- profile name can never be mistaken for a reference — has been live in the
-- pod since 2026-07-26 and is re-verified below by the descendant-literal
-- check quoted in the FINDING. The re-extension itself never happened; this
-- migration is that re-extension, five weeks late. Config-only, no image
-- roll needed — the FINDING's whole point was that nothing but this UPDATE
-- was ever missing.
--
-- error_step on both new commit steps routes to the NEXT render step (the
-- last one to "complete") so one kind's git-adapter failure cannot abort the
-- other two — the FINDING flagged this as [UNVERIFIED] whether the 07-26
-- config carried it; this migration deliberately does, since a single git
-- failure taking down all three kinds is the more serious failure mode.
--
-- Distinct output_field names per kind (directory_/adoption_/protocol_
-- _render_result, _commit_result) so the three kinds' results cannot clobber
-- each other in collected_data — required for FINDING's own verify query,
-- which reads adoption_render_result and protocol_render_result as separate
-- top-level keys.
--
-- Scope note: this migration does NOT touch the find-sites query's kind-blind
-- gating (FINDING "latent defect", register DIR-001 verify-later) — that is
-- explicitly owned by the portfolio_positioning Phase B3c wiring step, queued
-- behind their image roll. Nor does it add the three Phase B finance kinds
-- (mortgage-lender/savings-provider/health-insurer) to this chain — those
-- have no registered entities yet (research not enabled), so publishing them
-- now would commit empty files; that is Phase B's own remit, in its own order.

SELECT snapshot_agent('model-directory-publisher', '411_model_directory_publisher_re_extend_to_company_protocol.sql: pre-update');

BEGIN;

DO $do$
DECLARE
    n_active int;
    steps jsonb;
    n_steps int;
BEGIN
    -- Dual-active-row guard: refuse unless exactly one active non-snapshot row.
    SELECT count(*) INTO n_active FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n_active <> 1 THEN
        RAISE EXCEPTION '411: model-directory-publisher does not have exactly one active row (found %) - resolve before seeding', n_active;
    END IF;

    SELECT default_config#>'{workflow,steps}' INTO steps FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- Idempotency: already re-extended.
    IF steps ? 'render_adoption_json' THEN
        RAISE EXCEPTION '411: already applied - render_adoption_json step is present';
    END IF;

    -- Drift guard: the row must still be exactly the reverted 3-step,
    -- model-only chain the FINDING inspected on 2026-08-10.
    SELECT count(*) INTO n_steps FROM jsonb_object_keys(steps);
    IF n_steps <> 3 THEN
        RAISE EXCEPTION '411: expected 3 steps on the live row, found % - the row has drifted, re-check before reapplying', n_steps;
    END IF;
    IF NOT (steps ? 'render_model_directory_json' AND steps ? 'commit_model_directory' AND steps ? 'complete') THEN
        RAISE EXCEPTION '411: live row step names do not match the expected reverted chain - re-check before reapplying';
    END IF;
    IF steps #>> '{commit_model_directory,config,commit_message}' <> 'Update model directory' THEN
        RAISE EXCEPTION '411: commit_model_directory.config.commit_message has drifted - re-check before reapplying';
    END IF;
    IF steps #>> '{commit_model_directory,next_step}' <> 'complete' THEN
        RAISE EXCEPTION '411: commit_model_directory.next_step has drifted - re-check before reapplying';
    END IF;
END $do$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow}',
        $json$
        {
            "start_step": "render_model_directory_json",
            "processing_mode": "orchestrator",
            "timeout_seconds": 600,
            "steps": {
                "render_model_directory_json": {
                    "action": "render_model_directory",
                    "config": {"site_id": "input_data.site_id"},
                    "next_step": "commit_model_directory",
                    "description": "Build data/model-directory.json (+ full listing when a listing page exists) from the global registry",
                    "output_field": "directory_render_result"
                },
                "commit_model_directory": {
                    "action": "git_commit",
                    "config": {
                        "files_field": "directory_render_result.files",
                        "domain_field": "directory_render_result.domain",
                        "commit_message": "Update model directory"
                    },
                    "next_step": "render_adoption_json",
                    "error_step": "render_adoption_json",
                    "description": "Commit the JSON files into the site repo via git-adapter",
                    "output_field": "directory_commit_result"
                },
                "render_adoption_json": {
                    "action": "render_directory",
                    "config": {"site_id": "input_data.site_id", "kind": "company"},
                    "next_step": "commit_adoption_directory",
                    "description": "Build data/adoption-tracker.json (+ full listing when a listing page exists) from the global registry, kind=company",
                    "output_field": "adoption_render_result"
                },
                "commit_adoption_directory": {
                    "action": "git_commit",
                    "config": {
                        "files_field": "adoption_render_result.files",
                        "domain_field": "adoption_render_result.domain",
                        "commit_message": "Update adoption tracker"
                    },
                    "next_step": "render_protocol_json",
                    "error_step": "render_protocol_json",
                    "description": "Commit the JSON files into the site repo via git-adapter",
                    "output_field": "adoption_commit_result"
                },
                "render_protocol_json": {
                    "action": "render_directory",
                    "config": {"site_id": "input_data.site_id", "kind": "protocol"},
                    "next_step": "commit_protocol_directory",
                    "description": "Build data/protocol-tracker.json (+ full listing when a listing page exists) from the global registry, kind=protocol",
                    "output_field": "protocol_render_result"
                },
                "commit_protocol_directory": {
                    "action": "git_commit",
                    "config": {
                        "files_field": "protocol_render_result.files",
                        "domain_field": "protocol_render_result.domain",
                        "commit_message": "Update protocol tracker"
                    },
                    "next_step": "complete",
                    "error_step": "complete",
                    "description": "Commit the JSON files into the site repo via git-adapter",
                    "output_field": "protocol_commit_result"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": [
                            "directory_render_result", "directory_commit_result",
                            "adoption_render_result", "adoption_commit_result",
                            "protocol_render_result", "protocol_commit_result"
                        ]
                    }
                }
            }
        }
        $json$::jsonb
    ),
    updated_at = NOW()
WHERE type = 'model-directory-publisher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify with DO/RAISE - a verify block of SELECTs cannot stop the COMMIT.
DO $do$
DECLARE
    steps jsonb;
    n_steps int;
BEGIN
    SELECT default_config#>'{workflow,steps}' INTO steps FROM agent_definitions
    WHERE type = 'model-directory-publisher' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    SELECT count(*) INTO n_steps FROM jsonb_object_keys(steps);
    IF n_steps <> 7 THEN
        RAISE EXCEPTION '411: verify failed - expected 7 steps, got %', n_steps;
    END IF;
    IF steps #>> '{render_adoption_json,config,kind}' <> 'company' THEN
        RAISE EXCEPTION '411: verify failed - render_adoption_json kind is not the literal "company"';
    END IF;
    IF steps #>> '{render_protocol_json,config,kind}' <> 'protocol' THEN
        RAISE EXCEPTION '411: verify failed - render_protocol_json kind is not the literal "protocol"';
    END IF;
    IF steps #>> '{commit_model_directory,error_step}' <> 'render_adoption_json' THEN
        RAISE EXCEPTION '411: verify failed - commit_model_directory.error_step missing';
    END IF;
    IF steps #>> '{commit_adoption_directory,error_step}' <> 'render_protocol_json' THEN
        RAISE EXCEPTION '411: verify failed - commit_adoption_directory.error_step missing';
    END IF;
    IF jsonb_array_length(steps #> '{complete,config,output_fields}') <> 6 THEN
        RAISE EXCEPTION '411: verify failed - complete.config.output_fields does not list all six result fields';
    END IF;
END $do$;

COMMIT;

-- ROLLBACK recipe (hand-run, restores the reverted model-only chain from the
-- agent_definitions_backup row this migration took via snapshot_agent() above):
--   UPDATE agent_definitions live
--   SET default_config = bak.default_config
--   FROM (SELECT default_config FROM agent_definitions_backup
--         WHERE type = 'model-directory-publisher'
--           AND snapshot_reason = '411_model_directory_publisher_re_extend_to_company_protocol.sql: pre-update'
--         ORDER BY snapshot_taken_at DESC LIMIT 1) bak
--   WHERE live.type = 'model-directory-publisher' AND live.is_active
--     AND COALESCE(live.is_snapshot, false) = false AND live.deleted_at IS NULL;

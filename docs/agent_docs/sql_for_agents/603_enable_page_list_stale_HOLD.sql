-- 603_enable_page_list_stale_HOLD.sql
--
-- bugs_open/384, Phase 2 (the sweep). Adds "page_list_stale" to
-- completeness-discovery-agent's run_checks.config.checks array so the check
-- registered by platform/orchestration/actions/discovery_checks/check_page_list_stale.go
-- actually runs. Register entry PBP-048.
--
-- WHY _HOLD. The Go check is INERT until the chassis build that registers the
-- name has rolled; an unregistered name in the checks array is warn-and-skip
-- (not an error), so applying early would not break the agent — it would
-- silently do nothing while reading as enabled, which is the worse failure.
-- Apply by hand AFTER the roll is proven at the binary:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the 384 Phase-2 commit> <that stamp>
-- and confirm the name is in the binary's capability list before applying:
--   SELECT capabilities ? 'page_list_stale' FROM service_binary_capabilities ... (see build-pipeline.md BLD-*)
--
-- WHAT THE CHECK DOES (see the file header): for every page on the site whose
-- component consumes a page-IMAGE query source (blog_posts, pages_where_type:*,
-- pages_under_section), compares the STORED array's per-url image against a
-- fresh queryresolve result and files ONE page_rerender /
-- spec.reason='section_data_resolved' (status detected → the promoter
-- dispatches; key shared with the event emitter, so the two collapse). A source
-- that fails to resolve counts as UNKNOWN, never current. No retraction arm.
--
-- COST: one queryresolve.Resolve per distinct source per site per sweep (a few
-- cheap SELECTs), plus one page_components read per consumer page. The
-- completeness agent visits one site per tick (SCH-025 rotation).
--
-- ROLLBACK: 603_enable_page_list_stale_HOLD_ROLLBACK.sql (removes the name;
-- does not touch filed items — they are ordinary page_rerender requests and
-- complete on their own).

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent', '603_enable_page_list_stale: pre-update');

-- Idempotence guard: refuse a second application rather than double-append.
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'completeness-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["page_list_stale"]'::jsonb;
    IF done > 0 THEN
        RAISE EXCEPTION '603: already applied — completeness-discovery-agent already runs page_list_stale';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') || '["page_list_stale"]'::jsonb),
       updated_at = NOW()
 WHERE type = 'completeness-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- Verify with DO/RAISE: a bare SELECT cannot stop the COMMIT.
DO $$
DECLARE n_rows int; n_checks int;
BEGIN
    SELECT count(*) INTO n_rows FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'completeness-discovery-agent'
       AND default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["page_list_stale"]'::jsonb;
    IF n_rows <> 1 THEN
        RAISE EXCEPTION '603 verify: expected exactly 1 live completeness-discovery-agent carrying page_list_stale, found %', n_rows;
    END IF;

    -- The array must have grown by exactly one and still carry its neighbours:
    -- 44 checks measured 2026-08-24 (contact_form_undeliverable, orphan_pages,
    -- section_source_drift among them). A jsonb_set that replaced the array
    -- would pass the @> test above and take the other 44 checks down with it.
    SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions
     WHERE type = 'completeness-discovery-agent'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_checks < 45 THEN
        RAISE EXCEPTION '603 verify: checks array is % long — expected >= 45 (44 measured 2026-08-24 + page_list_stale); the append replaced the array', n_checks;
    END IF;
    IF NOT (SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' @> '["orphan_pages","contact_form_undeliverable"]'::jsonb
              FROM agent_definitions WHERE type='completeness-discovery-agent'
               AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL) THEN
        RAISE EXCEPTION '603 verify: neighbouring checks missing after the append';
    END IF;
END $$;

COMMIT;

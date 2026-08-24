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
-- and confirm the ROLLED binary registers the name, with a positive AND a
-- negative control in the same query (round 2005a846, debug_historian: never a
-- discovery grep, never `strings`; a control that cannot fail proves nothing).
-- service_binary_capabilities is written by each pod at startup from the
-- registry (kind='discovery_check' rows come from checks.Names()):
--
--   SELECT name, count(DISTINCT pod_name) AS pods, min(left(git_commit,12)) AS commit
--     FROM service_binary_capabilities
--    WHERE service='agent-chassis' AND kind='discovery_check'
--      AND last_seen_at > now() - interval '1 hour'
--      AND name IN ('page_list_stale',      -- must be PRESENT after the roll
--                   'orphan_pages',         -- positive control: registered for months
--                   'no_such_check_xyz')    -- negative control: must be ABSENT
--    GROUP BY 1 ORDER BY 1;
--
-- Expected: two rows (orphan_pages, page_list_stale) on the same commit, no
-- third. Measured 2026-08-24 before the roll: orphan_pages on 594 + 61 pods
-- across two commits, page_list_stale absent, no_such_check_xyz absent.
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
-- EXPOSURE ACKNOWLEDGED, DEFERRED TO THE HUMAN APPLYING THIS (round 2005a846
-- r2, bug_historian): the remedy this check files is the shared page_rerender /
-- rerender_page_sections path, which ESCALATES a whole page to the content
-- writer when any section lacks a required source:"llm" field (STY-048), and
-- the writer path is where whole-section regeneration has shrunk pages before
-- (bugs 238/287 family; the shrink guard now stands in front of it). This sweep
-- adds a fleet-wide, sweep-driven population to that path. No new guard is
-- added here — a pre-check duplicating rerender_page_sections' escalation rule
-- in this package would drift from it. Instead: (1) this migration is HELD, so
-- a human enables the sweep deliberately; (2) before applying, read the
-- escalation rate of section_data_resolved runs (1 of 25 in the 14 days to
-- 2026-08-24) —
--   SELECT coalesce(collected_data->'rerender_sections'->>'escalated','n/a'), count(*)
--     FROM orchestration_states WHERE owner_agent_type='page-rerender'
--      AND created_at > now()-interval '14 days'
--      AND collected_data->'input_data'->'spec'->>'reason'='section_data_resolved' GROUP BY 1;
-- and re-read it a week after; if the sweep's items escalate materially above
-- that baseline, run the ROLLBACK and bring the number to the owner.
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

-- Baseline read AT APPLY TIME, not remembered (round 2005a846 r2, debug_historian:
-- the array was 44 long on 2026-08-24 and will not be on the day this is applied).
CREATE TEMP TABLE _603_before AS
SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS n_checks,
       default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' AS checks
  FROM agent_definitions
 WHERE type = 'completeness-discovery-agent'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM _603_before;
    IF n <> 1 THEN
        RAISE EXCEPTION '603: expected exactly 1 live completeness-discovery-agent row, found %', n;
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

    -- The array must be EXACTLY the pre-apply array plus this one name — the
    -- baseline is the row read a moment ago, not a remembered count. A
    -- jsonb_set that replaced the array would pass the @> test above and take
    -- every other check on this agent down with it.
    SELECT jsonb_array_length(a.default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
      INTO n_checks FROM agent_definitions a
     WHERE a.type = 'completeness-discovery-agent'
       AND a.is_active AND NOT COALESCE(a.is_snapshot,false) AND a.deleted_at IS NULL;
    IF n_checks <> (SELECT b.n_checks FROM _603_before b) + 1 THEN
        RAISE EXCEPTION '603 verify: checks array is % long, expected % (the pre-apply length + 1); the append replaced or duplicated', n_checks, (SELECT b.n_checks FROM _603_before b) + 1;
    END IF;
    IF NOT (SELECT a.default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' @> (SELECT b.checks FROM _603_before b)
              FROM agent_definitions a WHERE a.type='completeness-discovery-agent'
               AND a.is_active AND NOT COALESCE(a.is_snapshot,false) AND a.deleted_at IS NULL) THEN
        RAISE EXCEPTION '603 verify: a pre-apply check name is missing after the append';
    END IF;
END $$;

COMMIT;

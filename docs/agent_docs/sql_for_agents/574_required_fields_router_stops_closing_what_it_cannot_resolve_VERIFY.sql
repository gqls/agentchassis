-- VERIFY for 574 + 576 (bugs_open/367) — re-runnable, read-only, side-effect free.
--
-- WHY THIS EXISTS. 574's own verify block checks the SHAPE of the rewritten JSON: anchors
-- gone, step count, branch arms, park statuses. It never RUNS the classify SQL it just
-- assembled — and a council seat objected on exactly that, citing the estate's landmine
-- "SQL embedded in a step config is DATA to your migration's probe — it parses only when
-- the step RUNS, and your migration's static check cannot see runtime type errors."
--
-- The seat was right, and it was right about the deeper thing too: the behavioural checks
-- WERE run while authoring (against the patched row, inside a rolled-back transaction) but
-- they lived in a scratchpad, so a re-apply or a later edit could not repeat them. This
-- file is those checks, encoded.
--
-- It reads the query out of the LIVE ROW, executes it, and asserts three outcomes that
-- could each have come out otherwise. RUN IT after any re-apply of 574/576, after any
-- re-seed of 410, and any time you are about to trust this router's dispositions.
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < <this file>
--
-- ⚠ THE TWO POSITIVE CONTROLS ARE THE POINT, not the first check. A "fix" that simply
-- stopped this router closing anything would satisfy C1 and break a real route; C2 and C3
-- are what catch that. If C1 passes and either control fails, the router is worse, not
-- better.
--
-- ⚠ C1 is pinned to a real production item (562788c3, ai-agent-roi-estimator/tool-cta on
-- site 4851f6fc). It asserts on the ROUTE, which is bugs_open/367 §6's own bar, and is
-- indifferent to that item's work-item status — deliberately, because status is downstream
-- of the thing under test and is overwritten at completion. If that component is ever
-- deployed, locked or retired, C1's expected value changes: re-pin it to another
-- non-deployed target rather than deleting the check, and say so here.

\pset pager off

DO $$
DECLARE
    q text; r record; spec jsonb; n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION 'VERIFY: expected exactly ONE active non-snapshot row, found % — every check below would be reading an arbitrary one', n;
    END IF;

    SELECT default_config -> 'workflow' -> 'steps' -> 'classify' -> 'config' ->> 'query'
      INTO q FROM agent_definitions
     WHERE type = 'required-fields-missing-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- 0. The static half, so a reverted row fails here with a clear message rather than
    --    producing confusing routes below.
    IF position('pc.build_status = ''deployed''' in q) > 0 THEN
        RAISE EXCEPTION 'VERIFY: the deployed-only predicate is back — 410 has been re-applied over 574. bugs_open/367 is LIVE again and it is SILENT.';
    END IF;
    IF position('tomb AS (SELECT EXISTS' in q) = 0 OR position('target_not_dispatchable' in q) = 0 THEN
        RAISE EXCEPTION 'VERIFY: 574 is not applied to the live row';
    END IF;
    IF position('AND COALESCE(item.spec->>''slot_name'', '''') <> ''''' in q) = 0 THEN
        RAISE EXCEPTION 'VERIFY: 576''s empty-slot guard on the tomb CTE is missing';
    END IF;

    -- 1. THE DEFECT ITSELF. A real, non-deployed target must resolve and must NOT close.
    EXECUTE 'SELECT route, target_state, component_id, html_len, n_still_empty FROM ('||q||') s'
      INTO r USING '4851f6fc-71cf-4160-a270-e03d6d3e0732'::uuid,
                   '562788c3-c9e9-4e8b-9967-c16dc9b8ed36'::uuid;
    RAISE NOTICE 'C1  real non-deployed target -> route=% target_state=% component=% html_len=% still_empty=%',
        r.route, r.target_state, left(r.component_id, 8), r.html_len, r.n_still_empty;
    IF r.route <> 'target_not_dispatchable' THEN
        RAISE EXCEPTION 'C1 FAILED: expected target_not_dispatchable, got % — a real finding is being disposed of again', r.route;
    END IF;
    IF r.component_id = '' THEN
        RAISE EXCEPTION 'C1 FAILED: the component did not resolve — resolution has narrowed back to the build axis';
    END IF;

    -- The remaining checks use a literal spec, so no rows are written and any (page, slot)
    -- can be probed. Swap the item CTE for a third bound parameter.
    q := replace(q,
        'WITH item AS (SELECT spec FROM site_work_items WHERE id = $2::uuid)',
        'WITH item AS (SELECT $3::jsonb AS spec)');
    IF position('$3::jsonb' in q) = 0 THEN
        RAISE EXCEPTION 'VERIFY: could not substitute the item CTE — its text has changed; re-read the live query and re-anchor this file';
    END IF;

    -- 2. POSITIVE CONTROL: a genuinely RETIRED component must STILL close as stale.
    spec := '{"page_name":"tool-clip-path","slot_name":"ported-page","missing_fields":["headline"]}'::jsonb;
    EXECUTE 'SELECT route, target_state FROM ('||q||') s' INTO r
      USING '6b49db8e-d447-4467-8277-4f3018af9897'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid, spec;
    RAISE NOTICE 'C2  retired component        -> route=% target_state=%', r.route, r.target_state;
    IF r.route <> 'stale' OR r.target_state <> 'component_retired' THEN
        RAISE EXCEPTION 'C2 FAILED: expected stale/component_retired, got %/% — a REAL close route has been disabled, which is worse than the bug', r.route, r.target_state;
    END IF;

    -- 3. POSITIVE CONTROL: a page that does not exist must STILL close as stale.
    spec := '{"page_name":"no-such-page-367-control","slot_name":"hero","missing_fields":["headline"]}'::jsonb;
    EXECUTE 'SELECT route, target_state FROM ('||q||') s' INTO r
      USING '6b49db8e-d447-4467-8277-4f3018af9897'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid, spec;
    RAISE NOTICE 'C3  page gone                -> route=% target_state=%', r.route, r.target_state;
    IF r.route <> 'stale' OR r.target_state <> 'page_missing' THEN
        RAISE EXCEPTION 'C3 FAILED: expected stale/page_missing, got %/%', r.route, r.target_state;
    END IF;

    RAISE NOTICE '--';
    RAISE NOTICE 'VERIFY OK — the query PARSES AND RUNS, the defect case parks with its component resolved, and BOTH real close routes still close.';
END $$;

-- 537 — build-dispatch-loop's `mark_complete` step DECLARES where `commit_sha`
--       comes from, with the `?` OPTIONAL-EXPLICIT marker: the handler's own
--       reply or NOTHING, never the whole-tree search. This is the resolver half
--       of `bugs_open/334`. CONFIG ONLY — live on apply.
--
--       This file REFUSES TO APPLY unless every handler that can produce a commit
--       of its own already exposes it at the standard path. That guard is the
--       point of the file, not decoration: see "THE GUARD" below.
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- `complete_work_item` declares `commit_sha` as an Optional input
-- (`load_work_item_actions.go:56`) and writes it into the work item's result
-- (line 937). NO live step config wired it. So the value arrived through the
-- resolver's last-resort arm — `findFieldRecursive`, the whole-tree search,
-- which collects every `commit_sha` anywhere in the tree and picks a winner.
--
-- Inside build-dispatch-loop's multi-iteration loop the tree accumulates one
-- handler result per iteration (`handler_result`, `handler_result_0`,
-- `handler_result_1`, …), each with a DIFFERENT sha. So every completion is a
-- genuine conflict, and the winner is whichever key sorts first.
--
-- ============================================================================
-- IT IS NOT MERELY "RIGHT BY LUCK" — IT IS DEMONSTRABLY WRONG SOMETIMES
-- ============================================================================
-- The earlier read of this class said the unsuffixed alias tracks the current
-- iteration, so the winner is right by accident of sort order. That is true as
-- far as it goes, and it is not the whole story. Worked case, traced by the
-- `bugs_open/306` session and re-verified here:
--
--   item `cc1db035` (tool-generator, an `add_tool`) carries `result.commit_sha`.
--   tool-generator CANNOT produce a commit: 8 orchestrations all-time, ZERO
--   carrying `commit_sha` anywhere in `collected_data`. The value sits at the
--   TOP LEVEL of the dispatch envelope — a sibling of `response`, `retry_payload`,
--   `agent_type` — not inside tool-generator's own `response.create_result`.
--   Its parent (`62df9f7a`) is a multi-iteration loop whose iteration 0 was a
--   `section-editor` item, a real committing handler. The search reached into
--   `handler_result_0` and glued ITERATION 0's COMMIT ONTO ITERATION 1's
--   UNRELATED ITEM.
--
-- So the flip does not just stop a wrong value being CHOSEN; it ends a class
-- where one item's commit is attributed to a different item entirely.
--
-- ⚠ EXPECT `tool-generator` TO SHOW NO `commit_sha` AFTER THIS APPLIES. That is
-- CORRECT — it never had one of its own. Anyone watching a dashboard should not
-- read that drop as a regression.
--
-- ============================================================================
-- WHY `?` AND NOT `!`, AND NOT A PLAIN WIRE
-- ============================================================================
-- `!` (strict) FAILS THE WHOLE EXTRACTION when the reference does not resolve.
-- Many completions legitimately have no commit: a handler that made no change, a
-- skipped rerender, an early exit. Measured on live traffic 2026-08-21 — of 31
-- recent bdl runs carrying a `handler_result`, 29 have the key and 2 do not:
-- one `page-build-handler` early exit whose response holds only `site_record`,
-- and one `page-rerender` whose `deploy_result` produced no commit. Strict would
-- turn both into hard failures. That is not hypothetical: migration 496 wired a
-- DIFFERENT field strictly on 2026-08-21 and every plain `add_tool` died at
-- `save_tool` while its work item read `complete` (hotfixed by 532 — to `?`).
--
-- A PLAIN wire resolves via Strategy 0 when the path exists and FALLS THROUGH TO
-- THE SEARCH when it does not — so on exactly those 2 runs it would go right back
-- to guessing, which is the defect.
--
-- `?` is the only spelling that says "the handler's own reply, or nothing".
--
-- ============================================================================
-- WHY ABSENCE IS SAFE HERE — read at the consumer, first-hand
-- ============================================================================
-- `complete_work_item` writes the field only when it is present:
--     if sha := inputs.Get("commit_sha"); sha != "" { resultData["commit_sha"] = sha }
-- (`load_work_item_actions.go:937`). Absent ⇒ the key is simply not written into
-- the item's result. No empty string is persisted, nothing downstream is handed a
-- blank. And `result!` already maps the WHOLE `handler_result` into `result`, and
-- that envelope still CONTAINS the sha at its own path — so the top-level
-- `result.commit_sha` is a convenience duplicate, not the only copy.
--
-- ============================================================================
-- THE GUARD — why this file can refuse itself, and why it is not a list
-- ============================================================================
-- The regression this wire could cause is precise: a handler that DOES produce a
-- commit but does not expose it at `response.commit_sha` would silently stop
-- recording `result.commit_sha`. Nine handlers were standardised for exactly this
-- (migrations 519/521/522/523/527/528/534/535/536, by the `bugs_open/306` lane).
--
-- A hand-typed list of those nine would be a snapshot of what one census could
-- see — the mistake migration 516 had to be rebuilt to remove. So this file
-- DERIVES the set at apply time, and the derivation is the part to read:
--
--   * "can produce a commit" is a property of the HANDLER — its OWN
--     `orchestration_states` carry a `commit_sha`. NOT "an item of this handler
--     recorded one", which is a property of the RESOLVER WE ARE REMOVING and
--     which counts the contamination above as a requirement. That distinction
--     cost a false positive when this lane first sized the population, and it is
--     the whole reason `tool-generator` is correctly NOT in the set.
--   * "is a live handler" is `site_work_items.handler_agent` — the empirical
--     dispatch population, not a scan for agents that merely own a git step.
--   * "is standardised" is its `complete` step exposing `commit_sha` through
--     `result_mapping` (`output_fields` CANNOT rename a nested path to a
--     top-level key, which is why the 306 lane had to convert rather than add).
--
-- If any handler is in the first two sets and not the third, this migration
-- ABORTS naming it. The failure direction is deliberate: a handler wrongly
-- INCLUDED blocks the apply (loud, cheap); a handler wrongly excluded loses a
-- field silently (quiet, expensive).
--
-- ROLLBACK: 537_bdl_mark_complete_declares_commit_sha_ROLLBACK.sql
-- ============================================================================

BEGIN;

SELECT snapshot_agent('build-dispatch-loop',
                      '537_bdl_mark_complete_declares_commit_sha: pre-update');

DO $$
DECLARE
    step_path text[];
    cfg       jsonb;
    unready   text;
    n_ready   int;
BEGIN
    -- ---------------------------------------------------------------------
    -- THE GUARD, first: never edit before proving the estate is ready.
    -- ---------------------------------------------------------------------
    WITH producers AS (          -- handlers whose OWN runs can produce a commit
        SELECT DISTINCT owner_agent_type AS agent
          FROM orchestration_states
         WHERE created_at > now() - interval '30 days'
           AND collected_data::text LIKE '%commit_sha%'
    ), live_handlers AS (        -- ...and that actually handle dispatched work
        SELECT DISTINCT handler_agent AS agent
          FROM site_work_items
         WHERE handler_agent IS NOT NULL AND handler_agent <> ''
           AND created_at > now() - interval '7 days'
    ), must_expose AS (
        SELECT agent FROM producers INTERSECT SELECT agent FROM live_handlers
    ), exposed AS (              -- ...and that already surface it top-level
        SELECT DISTINCT d.type AS agent
          FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'complete_workflow'
           AND (s.value->'config'->'result_mapping') ? 'commit_sha'
    )
    SELECT string_agg(agent, ', ' ORDER BY agent), count(*)
      INTO unready, n_ready
      FROM (SELECT agent FROM must_expose EXCEPT SELECT agent FROM exposed) x;

    IF n_ready > 0 THEN
        RAISE EXCEPTION '537 REFUSED: these handlers can produce a commit and are live in dispatch, '
            'but do NOT expose commit_sha at response.commit_sha: %. Wiring the loop now would '
            'silently stop recording result.commit_sha for them. Convert their complete step to '
            'result_mapping first (see migrations 519-536), then re-run this.', unready;
    END IF;

    -- The set must not be EMPTY either: zero handlers exposing it means either
    -- the standardisation was rolled back or this query has gone blind, and
    -- applying into that state is how a guard becomes decoration.
    IF (SELECT count(*) FROM (
            SELECT DISTINCT d.type FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
             WHERE d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
               AND s.value->>'action' = 'complete_workflow'
               AND (s.value->'config'->'result_mapping') ? 'commit_sha') y) = 0 THEN
        RAISE EXCEPTION '537 REFUSED: NO live agent exposes commit_sha via result_mapping. The '
            'handler standardisation this wire depends on is absent — refusing to wire a path '
            'nothing produces';
    END IF;

    -- ---------------------------------------------------------------------
    -- Locate the step by DISCOVERY, not by a typed path (516's lesson): the
    -- step is nested inside the loop's sub_workflow, and a top-level
    -- jsonb_each census cannot see it.
    -- ---------------------------------------------------------------------
    SELECT path INTO step_path FROM (
        WITH RECURSIVE walk AS (
            SELECT ARRAY[]::text[] AS path, d.default_config AS node
              FROM agent_definitions d
             WHERE d.type = 'build-dispatch-loop' AND d.is_active
               AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
            UNION ALL
            SELECT w.path || e.key, e.value
              FROM walk w CROSS JOIN LATERAL jsonb_each(w.node) e
             WHERE jsonb_typeof(w.node) = 'object'
        )
        SELECT path FROM walk
         WHERE jsonb_typeof(node) = 'object' AND node->>'action' = 'complete_work_item'
    ) z;

    IF step_path IS NULL THEN
        RAISE EXCEPTION '537: no complete_work_item step found in build-dispatch-loop — refusing to guess';
    END IF;

    cfg := (SELECT default_config #> (step_path || ARRAY['config'])
              FROM agent_definitions WHERE type = 'build-dispatch-loop' AND is_active
                AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL);

    IF cfg ? 'commit_sha?' THEN
        RAISE EXCEPTION '537: commit_sha? is already declared — already applied';
    END IF;
    IF cfg ? 'commit_sha' OR cfg ? 'commit_sha!' THEN
        RAISE EXCEPTION '537: an existing commit_sha wire (%) is present — someone chose that '
            'spelling deliberately; read their reason before converting',
            COALESCE(cfg->>'commit_sha', cfg->>'commit_sha!');
    END IF;

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
               step_path || ARRAY['config','commit_sha?'],
               to_jsonb('handler_result.response.commit_sha'::text), true),
           updated_at = NOW()
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    RAISE NOTICE '537: wired commit_sha? at %', array_to_string(step_path || ARRAY['config'], '.');
END $$;

-- ============================================================================
-- VERIFY — a DO block, not SELECTs (ON_ERROR_STOP does not abort a COMMIT on a
-- non-empty result set).
-- ============================================================================
DO $$
DECLARE
    got text;
BEGIN
    SELECT node->>'commit_sha?' INTO got FROM (
        WITH RECURSIVE walk AS (
            SELECT ARRAY[]::text[] AS path, d.default_config AS node
              FROM agent_definitions d
             WHERE d.type = 'build-dispatch-loop' AND d.is_active
               AND COALESCE(is_snapshot,false) = false AND d.deleted_at IS NULL
            UNION ALL
            SELECT w.path || e.key, e.value FROM walk w CROSS JOIN LATERAL jsonb_each(w.node) e
             WHERE jsonb_typeof(w.node) = 'object'
        )
        SELECT node FROM walk
         WHERE jsonb_typeof(node) = 'object' AND node ? 'commit_sha?'
    ) z;

    IF got IS DISTINCT FROM 'handler_result.response.commit_sha' THEN
        RAISE EXCEPTION '537 VERIFY FAILED: commit_sha? reads %, expected handler_result.response.commit_sha',
            COALESCE(got, '<null>');
    END IF;

    IF (SELECT count(*) FROM agent_definitions_backup
         WHERE type = 'build-dispatch-loop'
           AND snapshot_reason LIKE '537_bdl_mark_complete%') < 1 THEN
        RAISE EXCEPTION '537 VERIFY FAILED: no snapshot row in agent_definitions_backup';
    END IF;

    RAISE NOTICE '537 OK: mark_complete declares commit_sha? = handler_result.response.commit_sha';
END $$;

COMMIT;

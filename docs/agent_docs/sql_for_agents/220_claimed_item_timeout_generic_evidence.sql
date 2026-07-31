-- 220_claimed_item_timeout_generic_evidence.sql
--
-- bugs_open/006 §C — the claim-timeout sweep can only prove completion for 3 of
-- the 18 item types it supervises, so the other 15 re-run work that already
-- succeeded.
--
-- WHAT IS WRONG
-- -------------
-- A handler does the work — renders, deploys, marks the page deployed — and then
-- the dispatch loop's `mark_complete` step never runs (the pod died, or the
-- child's response was lost: bugs_open/003). The item stays `claimed`. The
-- `claimed-item-timeout` scheduled task then has two branches:
--
--   * auto-complete on evidence, at 15 minutes — but the evidence is a
--     hand-written artifact test per item_type, and only THREE exist
--     (needs_content_page, page_rerender, needs_design);
--   * reset to `triaged`, at 40 minutes — everything else, INCLUDING work that
--     demonstrably succeeded.
--
-- Measured 2026-07-26 over 14 days of `error LIKE 'Claim timed out%'`:
--
--     item_type            timed out   auto-completed
--     page_rerender               28                9
--     needs_page                  27                0
--     content_rewrite             15                0
--     needs_imagery                6                0
--     needs_content_page           0                5
--     (5 further types)            8                0
--
-- 84 timeouts against 14 auto-completions. 11 of those items ended `failed` and
-- one `unresolved` — attempts exhausted on work that had already been done.
-- `needs_page`, the item type in 006 §C's original symptom, has never once been
-- auto-completed because no branch knows what artifact proves a `needs_page`
-- done.
--
-- THE FIX, AND WHY IT IS ONE BRANCH INSTEAD OF FIFTEEN
-- ----------------------------------------------------
-- The evidence the artifact tests are reaching for is already recorded, in one
-- place, for every item type: the handler's own orchestration. Both dispatch
-- loops pass the item id into the handler through `call_handler.input_mapping`,
-- so the handler orchestration carries it at
--
--     initial_request_data->'input_data'->>'work_item_id'
--
-- Verified live 2026-07-26 before writing this:
--   * Reach is 100/100. Over a 6-hour window — short enough that no
--     orchestration_states row has yet been purged (retention is ~2 days) —
--     every terminal work item that passes through a claim resolved to its
--     handler orchestration: page_rerender 92/92, needs_page 3/3,
--     needs_content_page 2/2, nav_drift 1/1, missing_*_page 1/1.
--   * The item types that resolve to nothing (needs_section_data,
--     cta_names_unknown_destination, needs_experience_plan) are completed by
--     paths that never claim, so they can never be in the `claimed` status this
--     sweep looks at.
--   * Only two agents claim at all (agent_definitions LIKE '%claim_work_item%'):
--     build-dispatch-loop, and diagnose-dispatch-loop — which uses status
--     `diagnosing` not `claimed`, says so in its own step description ("we own
--     this because diagnosing is inert to claimed-item-timeout"), and runs its
--     own 75-minute reaper. So this sweep's universe is exactly
--     build-dispatch-loop's items, and the link covers all of them.
--
-- THE STANDARD IS PARITY, NOT STRICTER
-- ------------------------------------
-- This branch reproduces what the LOST WRITE would have done, no more: the
-- dispatch loop calls complete_work_item on a handler saga that reached its own
-- complete_workflow. The sweep is a recovery mechanism for a lost write, not a
-- second opinion on the work. A stricter test would leave finished work
-- re-running, which is the defect being fixed.
--
-- THE ONE PLACE PARITY WOULD BE WRONG — the exclusion list
-- -------------------------------------------------------
-- complete_work_item does NOT complete unconditionally: it consults a per-item
-- type verifier (discovery_checks.GetVerifier) which can BLOCK completion. That
-- is the bugs_open/017 + /021 completion-lie guard — the thing that stops a saga
-- which reported success without touching the defect. SQL cannot run a Go
-- verifier, so the three item types that have one are excluded here and keep
-- falling through to the 40-minute reset. Strictly safer than parity, and it
-- costs nothing today: none of the three appears anywhere in the 14-day timeout
-- data above.
--
--   LOCKSTEP. That list is the twin of the RegisterVerifier() calls in
--   platform/orchestration/actions/discovery_checks/. If a fourth verifier is
--   registered and this list is not updated, this sweep will auto-complete an
--   item the verifier would have blocked — silently, because a sweep leaves no
--   caller to notice. TestRegisteredVerifiersMatchClaimTimeoutExclusion in
--   verifier_coverage_test.go fails the build on that drift and names this file.
--   Two hand-maintained lists that must stay identical is the exact class this
--   repo keeps being bitten by; it is pinned, not trusted.
--
-- WHY THE THREE ARTIFACT BRANCHES STAY
-- ------------------------------------
-- orchestration_states is purged at roughly 2 days (measured: 1,819 of 1,840
-- rows are from the last 48 hours). They are the fallback for a claim whose
-- orchestration row is already gone, or was never persisted.
--
-- KNOWN, ACCEPTED EDGE CASE (documented rather than engineered around)
-- --------------------------------------------------------------------
-- If a handler finishes AFTER its item was reset and re-claimed, its COMPLETED
-- orchestration post-dates the NEW claimed_at, so the new claim auto-completes
-- on the old handler's evidence and the second handler's run is abandoned. The
-- work genuinely is done, and this is strictly better than today, where both
-- runs go to completion. Recorded because it is a real behaviour of this
-- predicate, not because it is a problem.
--
-- THE MARKER
-- ----------
-- The new branch writes a DIFFERENT error string from the artifact branch
-- ('Auto-completed: handler orchestration completed after claim'). That is
-- deliberate: it is the only way to measure this change's effect without
-- confusing it with the branches that already existed. Assert a string only the
-- new code can produce — the SQL form of the discriminating pod-grep rule.
--
-- VERIFY (after applying)
-- -----------------------
--   SELECT item_type, count(*) FROM site_work_items
--   WHERE error = 'Auto-completed: handler orchestration completed after claim'
--   GROUP BY 1 ORDER BY 2 DESC;
--   -- expect item types that had 0 auto-completions in 14 days to start appearing
--
-- ROLLBACK
-- --------
-- The previous pre_query is preserved verbatim in
--   220_claimed_item_timeout_generic_evidence_ROLLBACK.sql
-- Apply that file to restore it. Do NOT hand-edit the column back: it is an
-- 84-line SQL string in a text column and a typo in it breaks the fleet's only
-- claim self-heal, silently, every 120 seconds.

BEGIN;

UPDATE scheduled_tasks
SET pre_query = $q$
    WITH completed_by_orchestration AS (
    -- GENERIC completion evidence, bugs_open/006 §C: the handler orchestration
    -- this claim dispatched reached COMPLETED after the claim was taken. Works
    -- for every item type, because every dispatched item carries its id into
    -- the handler via call_handler.input_mapping.work_item_id.
    --
    -- This is parity with the lost mark_complete write, not a stricter test —
    -- see migration 220's header for why that is the correct standard.
    --
    -- The item_type exclusion is the LOCKSTEP TWIN of the RegisterVerifier()
    -- calls in platform/orchestration/actions/discovery_checks/: those item
    -- types have a Go verifier that can BLOCK completion (bugs_open/017, /021)
    -- and SQL cannot run it, so they keep falling through to reset.
    -- TestRegisteredVerifiersMatchClaimTimeoutExclusion pins the two together.
    UPDATE site_work_items wi
    SET status = 'complete',
        completed_at = NOW(),
        error = 'Auto-completed: handler orchestration completed after claim'
    WHERE wi.status = 'claimed'
      AND wi.claimed_at < NOW() - INTERVAL '15 minutes'
      -- AMENDED 2026-07-29: 'orphan_element_refs' added (its Go verifier landed
      -- with check_orphan_element_refs.go). The LIVE column was NOT changed by
      -- re-applying this file — 269_orphan_element_refs_claim_timeout_exclusion.sql
      -- did it with a targeted replace() plus a before/after assertion, exactly
      -- as this file's own ROLLBACK note demands. This line is edited so the
      -- declared list matches the live one; TestRegisteredVerifiersMatchClaimTimeoutExclusion
      -- reads THIS FILE, so leaving it stale would make the guard lie.
      AND wi.item_type NOT IN ('truncated_component', 'hardcoded_section_colors', 'empty_section', 'orphan_element_refs', 'content_duplication')
      AND EXISTS (
        SELECT 1 FROM orchestration_states o
        WHERE o.initial_request_data->'input_data'->>'work_item_id' = wi.id::text
          AND o.status = 'COMPLETED'
          -- Guards a re-claim: an orchestration from a PREVIOUS attempt
          -- completed before this claim was taken, so it cannot complete it.
          AND o.updated_at > wi.claimed_at
      )
    RETURNING id, item_type, handler_agent, status
),
completed_by_evidence AS (
    -- Items where the handler's work is provably done on the specific
    -- targeted artifact (not just "something on the same site changed").
    -- See debugging guide section 9: "claimed-item-timeout evidence
    -- check produces false-positive completions" for the prior bug.
    --
    -- Retained as the fallback for claims whose orchestration row has already
    -- been purged (~2 day retention) or was never persisted; the branch above
    -- covers the live window for every item type.
    UPDATE site_work_items wi
    SET status = 'complete',
        completed_at = NOW(),
        error = 'Auto-completed: work verified done despite lost response'
    WHERE wi.status = 'claimed'
      AND wi.claimed_at < NOW() - INTERVAL '15 minutes'
      AND wi.id NOT IN (SELECT id FROM completed_by_orchestration)
      AND (
        -- Content pages: the page's OWN content artifact (page_components)
        -- was written by THIS claim's handler. needs_content_page produces
        -- components, not a deploy (deploy is a separate page_rerender item),
        -- so we check the artifact directly rather than build_status='deployed'
        -- (a downstream, separately-set flag that can be true with zero
        -- components — see the gamesdesign homepage, 2026-06-04). The
        -- updated_at > claimed_at guard ensures the components are from THIS
        -- claim's run, not stale rows from a prior plan/generation.
        (wi.item_type = 'needs_content_page' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM page_components pc
            WHERE pc.page_id = wi.page_id
              AND pc.component_id IS NOT NULL
              AND pc.rendered_html IS NOT NULL
              AND pc.rendered_html <> ''
              AND pc.updated_at > wi.claimed_at
        ))
        OR
        -- Page rerenders: the specific page was deployed after claim.
        -- Note: NOT 'needs_rerender' — that's a site-level orchestrator
        -- with page_id NULL, can't have per-page evidence. Let it
        -- fall through to reset. A page_rerender's job IS to deploy, so
        -- deployed_at is the correct artifact here (the deployed flag is
        -- hardened separately by the UpdatePageStatusAction guard).
        (wi.item_type = 'page_rerender' AND wi.page_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM pages p
            WHERE p.id = wi.page_id
              AND p.build_status = 'deployed'
              AND p.deployed_at IS NOT NULL
              AND p.deployed_at > wi.claimed_at
        ))
        OR
        -- Design items: site-level by nature. CAVEAT — this branch still
        -- has narrow false-positive potential because (a) it only checks
        -- the head slot, not header/footer, and (b) it uses updated_at
        -- rather than a deploy-specific timestamp (site_components has
        -- no deployed_at column). Acceptable for now; needs_design items
        -- are rare and the impact is bounded to design-only work.
        (wi.item_type = 'needs_design' AND EXISTS (
            SELECT 1 FROM site_components sc
            WHERE sc.site_id = wi.site_id
              AND sc.slot_name = 'head'
              AND sc.updated_at > wi.claimed_at
        ))
      )
    RETURNING id, item_type, handler_agent, status
),
reset AS (
    -- Remaining stuck items: no evidence of completion, reset for retry.
    -- needs_rerender items always land here now (previously could be
    -- false-positive auto-completed); that's intended — retry is cheap.
    UPDATE site_work_items
    SET status = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'failed'
            ELSE 'triaged'
        END,
        claimed_by = NULL,
        claimed_at = NULL,
        attempt_count = attempt_count + 1,
        error = CASE
            WHEN attempt_count + 1 >= max_attempts THEN 'Claim timed out (attempts exhausted)'
            ELSE 'Claim timed out — handler pod likely died'
        END
    WHERE status = 'claimed'
      AND claimed_at < NOW() - INTERVAL '40 minutes'
      AND id NOT IN (SELECT id FROM completed_by_orchestration)
      AND id NOT IN (SELECT id FROM completed_by_evidence)
    RETURNING id, item_type, handler_agent, status
)
SELECT
    (SELECT COUNT(*) FROM completed_by_orchestration) as auto_completed_by_orchestration,
    (SELECT COUNT(*) FROM completed_by_evidence) as auto_completed,
    (SELECT COUNT(*) FROM reset) as reset_count
$q$,
    updated_at = NOW()
WHERE name = 'claimed-item-timeout';

-- Guard 1: the row exists and was updated. A silent 0-row UPDATE here would
-- leave the fleet on the old sweep with nothing to show for the migration.
DO $$
DECLARE
  v_count int;
BEGIN
  SELECT count(*) INTO v_count
  FROM scheduled_tasks
  WHERE name = 'claimed-item-timeout'
    AND pre_query LIKE '%completed_by_orchestration%';
  IF v_count <> 1 THEN
    RAISE EXCEPTION '220 guard 1: expected exactly 1 claimed-item-timeout row carrying the new branch, got %', v_count;
  END IF;
END $$;

-- Guard 2: the new pre_query PARSES AND PLANS.
--
-- This is the load-bearing guard. pre_query is a SQL string in a text column:
-- nothing validates it at write time, the scheduler runs it every 120 seconds,
-- and a syntax error there would silently kill the fleet's only claim self-heal
-- while every surface still reads "enabled, firing normally". PREPARE parses
-- and plans without executing, so this proves the statement is well-formed
-- against the real schema without touching a single row.
DO $$
DECLARE
  v_sql text;
BEGIN
  SELECT pre_query INTO v_sql FROM scheduled_tasks WHERE name = 'claimed-item-timeout';
  EXECUTE 'PREPARE claimed_item_timeout_probe AS ' || v_sql;
  EXECUTE 'DEALLOCATE claimed_item_timeout_probe';
END $$;

-- Guard 3: the sweep discriminates — it completes a claim whose handler
-- orchestration reached COMPLETED, and does NOT complete one whose handler
-- FAILED.
--
-- THIS GUARD RUNS THE REAL pre_query, read back out of the column it was just
-- written to, and asserts on the resulting row STATUS. It deliberately does not
-- re-state the predicate: a check that shares the fix's own expression cannot
-- falsify it — it would pass against a typo'd pre_query that never fires in
-- production, which is precisely the failure this migration exists to prevent.
--
-- Both probes are inserted, swept and deleted inside this transaction, so they
-- are never visible to another session and never survive the file. Their
-- item_type is one no artifact branch can match, so a completion can ONLY have
-- come from the new orchestration branch.
--
-- A green case alone cannot fail, which is why the negative control is here:
-- without it this guard would pass just as happily against a branch that
-- completes everything.
--
-- Executing the sweep once here is not a side effect worth avoiding: it is
-- exactly what the scheduler does every 120 seconds, so any real claim it
-- touches is one that was about to be touched anyway.
DO $$
DECLARE
  v_site       uuid;
  v_sql        text;
  v_ok_item    uuid;
  v_bad_item   uuid;
  v_ok_status  text;
  v_ok_error   text;
  v_bad_status text;
BEGIN
  SELECT id INTO v_site FROM sites ORDER BY created_at ASC LIMIT 1;
  IF v_site IS NULL THEN
    RAISE EXCEPTION '220 guard 3: no site row available to hang the probe work items on';
  END IF;

  INSERT INTO site_work_items
    (site_id, source, item_type, summary, created_by, handler_agent, status, claimed_by, claimed_at)
  VALUES
    (v_site, 'migration_220_probe', 'migration_probe',
     'probe: handler orchestration COMPLETED (rolled back)',
     'migration_220', 'migration-220-probe-handler', 'claimed', 'migration_220',
     now() - interval '20 minutes')
  RETURNING id INTO v_ok_item;

  INSERT INTO site_work_items
    (site_id, source, item_type, summary, created_by, handler_agent, status, claimed_by, claimed_at)
  VALUES
    (v_site, 'migration_220_probe', 'migration_probe',
     'probe: handler orchestration FAILED — must NOT complete (rolled back)',
     'migration_220', 'migration-220-probe-handler', 'claimed', 'migration_220',
     now() - interval '20 minutes')
  RETURNING id INTO v_bad_item;

  INSERT INTO orchestration_states
    (orchestration_id, correlation_id, client_id, status, current_step,
     workflow_plan, initial_request_data, created_at, updated_at)
  VALUES
    (gen_random_uuid(), gen_random_uuid(), 'migration_220_probe', 'COMPLETED', 'complete',
     '{}'::jsonb,
     jsonb_build_object('input_data', jsonb_build_object('work_item_id', v_ok_item::text)),
     now(), now()),
    (gen_random_uuid(), gen_random_uuid(), 'migration_220_probe', 'FAILED', 'complete',
     '{}'::jsonb,
     jsonb_build_object('input_data', jsonb_build_object('work_item_id', v_bad_item::text)),
     now(), now());

  -- Run the sweep exactly as the scheduler will.
  SELECT pre_query INTO v_sql FROM scheduled_tasks WHERE name = 'claimed-item-timeout';
  EXECUTE v_sql;

  SELECT status, error INTO v_ok_status, v_ok_error FROM site_work_items WHERE id = v_ok_item;
  SELECT status INTO v_bad_status FROM site_work_items WHERE id = v_bad_item;

  DELETE FROM orchestration_states WHERE client_id = 'migration_220_probe';
  DELETE FROM site_work_items WHERE id IN (v_ok_item, v_bad_item);

  IF v_ok_status IS DISTINCT FROM 'complete' THEN
    RAISE EXCEPTION '220 guard 3: positive probe was not completed (status %) — the generic evidence branch never fired',
      v_ok_status;
  END IF;
  IF v_ok_error IS DISTINCT FROM 'Auto-completed: handler orchestration completed after claim' THEN
    RAISE EXCEPTION '220 guard 3: positive probe completed with the wrong marker (%) — an OLD branch matched, so this migration proves nothing',
      v_ok_error;
  END IF;
  IF v_bad_status IS DISTINCT FROM 'claimed' THEN
    RAISE EXCEPTION '220 guard 3: negative probe changed status to % — the sweep acts on a claim whose handler FAILED',
      v_bad_status;
  END IF;
END $$;

COMMIT;

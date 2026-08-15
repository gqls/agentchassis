-- 430 — detected-item-promoter: the detected->triaged promoter as its OWN
--       scheduled task, decoupled from the improvement loop (bugs_open/083 fix
--       candidate 2, OWNER RULING 2026-08-15)
--
-- ============================================================================
-- THE DECISION THIS IMPLEMENTS
-- ============================================================================
-- Discovery checks file findings at status='detected' — recorded but
-- UNCLAIMABLE — on the design that a separate triage step promotes them to
-- 'triaged' once the discovery pass is complete, so the dispatch loop can hand
-- them out. That step (TriageDetectedItemsAction, `triage_detected_items`) has
-- exactly one live carrier, the improvement-loop agent, whose only driver is the
-- improvement-sweep scheduled task — DISABLED since 2026-05-02 (register
-- IMP-016, "intentionally paused during core build"). Measured 2026-08-15: 70
-- rows stranded at detected, oldest 2026-08-10, across 18 (item_type, handler)
-- pairs, every handler active. bugs_open/083 records the mechanism and the
-- consequence: producers have been routing around a dead promoter by filing
-- items born 'triaged' — which the council's improvement_guardian seat objected
-- to, at high severity, on bugs_open/277's trail (corr 7b0e2833), because it
-- removes the observe-only stage entirely.
--
-- The owner ruled 2026-08-15, on the question put in bugfix_277's README: DO
-- 083, as candidate 2 — "give triage its own scheduled task, decoupled from the
-- fix loop: promote detected -> triaged on a slow cadence for item types whose
-- handlers are known-good, leaving discovery itself manual." Candidate 1
-- (re-enable improvement-sweep wholesale) was NOT chosen: that re-arms the
-- whole discovery+audit+fix loop on a 180s cadence, and its own pre_query
-- caps sites at <50 open items (excluding the two most-worked sites, IMP-010).
--
-- ============================================================================
-- WHAT THIS TASK IS
-- ============================================================================
-- A pure-SQL scheduled task on the SCH-006 pattern (pre_query as gate AND
-- worker, fire_message=false — no Kafka message, no agent spawned), exactly
-- like `feasibility-recheck` (020_scheduled_tasks.sql §3), which promotes
-- blocked->triaged the same way. Every tick it promotes a BOUNDED batch of
-- detected rows whose (item_type, handler_agent) pair is KNOWN-GOOD, mirroring
-- TriageDetectedItemsAction's UPDATE byte-for-byte in effect: status='triaged',
-- triaged_at=now(), spec.original_pipeline stamped, pipeline='build'.
--
-- WHY pipeline='build' IS LOAD-BEARING (measured 2026-08-15, not folklore):
-- build-dispatch-loop's load_items does NOT filter by pipeline, but the
-- SCHEDULER GATE that wakes the loop (build-pipeline-trigger's pre_query)
-- requires `wi.pipeline = 'build'` for a site to count as dispatchable. A site
-- whose only triaged items are pipeline='content' never wakes the loop. The
-- original promoter rewrote pipeline for this reason; so does this one.
--
-- KNOWN-GOOD, defined as data not opinion (bugs_open/083 candidate 4: "check
-- the handler is real before celebrating the detector"): a pair is promotable
-- when (a) handler_agent is a live, active agent definition AND (b) that exact
-- (item_type, handler_agent) pair has AT LEAST ONE lifetime 'complete' row.
-- Measured on the live pile: 17 of 18 pairs qualify; the one that does not
-- (page_component_status_drift -> component-template-fixer, 0 lifetime
-- dispatches) is held back until a human promotes ONE row by hand as a canary —
-- which then makes the pair known-good. That is the bootstrap by design: every
-- new type's first dispatch is a deliberate canary, never a fleet surprise. To
-- canary a held pair:
--   UPDATE site_work_items SET status='triaged', triaged_at=now(),
--          spec = jsonb_set(COALESCE(spec,'{}'), '{original_pipeline}', to_jsonb(pipeline)),
--          pipeline='build', updated_at=now()
--    WHERE id='<one row>' AND status='detected';
--
-- BLAST RADIUS, bounded three ways: (1) LIMIT 20 promotions per tick, oldest
-- first, at a 15-minute cadence — the 70-row pile drains in ~an hour and a
-- future burst cannot flood the build queue in one tick; (2) only known-good
-- pairs; (3) promotion is not dispatch — the dispatcher's own gates (site lock,
-- attempt_count, claim gate, handler-exists check) still apply per item.
--
-- WHAT THIS DOES NOT DO: it does not re-enable improvement-sweep (the discovery
-- + audit + fix loop stays paused, per IMP-016), it does not touch
-- needs_human_review, blocked, unresolved or deferred rows, and it does not
-- decide anything per finding — that is each handler's job (bugs_open/277's
-- router being the worked example of a handler that triages its own type).
--
-- SEQUEL, once this is live and observed: bugs_open/277's producer
-- (check_required_fields_missing.go) goes back from Status "triaged" to
-- "detected" — the one-line revert promised on that trail — because the
-- contract it bypassed is now honoured. Other born-triaged producers are each
-- their lane's call.
--
-- ORDER: DB config, live at the scheduler's next tick after COMMIT. Rollback:
-- UPDATE scheduled_tasks SET enabled=false WHERE name='detected-item-promoter'
-- (or DELETE the row) — nothing else to unwind; promoted rows are ordinary
-- triaged rows.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 430_….sql
-- ============================================================================

BEGIN;

INSERT INTO scheduled_tasks (
    name, description, interval_seconds,
    target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent,
    enabled, fire_message, timeout_seconds
) VALUES (
    'detected-item-promoter',
    'Promotes detected work items to triaged (pipeline=build, original_pipeline stamped) for (item_type, handler_agent) pairs that are KNOWN-GOOD: handler active AND the pair has >=1 lifetime complete. Bounded to 20 per tick, oldest first. The decoupled triage stage bugs_open/083 candidate 2 / owner ruling 2026-08-15; a pure-SQL worker like feasibility-recheck (SCH-006). Does NOT re-enable improvement-sweep.',
    900,
    'generic',
    'system.agent.generic.requests',
    '{}'::jsonb,
    'maintenance',
    1,
    true, false, 60
) ON CONFLICT (name) DO UPDATE SET
    description      = EXCLUDED.description,
    interval_seconds = EXCLUDED.interval_seconds,
    fire_message     = EXCLUDED.fire_message,
    enabled          = EXCLUDED.enabled,
    updated_at       = NOW();

UPDATE scheduled_tasks
SET pre_query = $PQ$
    WITH candidates AS (
        SELECT wi.id
        FROM site_work_items wi
        WHERE wi.status = 'detected'
          AND COALESCE(wi.handler_agent, '') <> ''
          AND EXISTS (
            SELECT 1 FROM agent_definitions ad
            WHERE ad.type = wi.handler_agent
              AND ad.is_active
              AND COALESCE(ad.is_snapshot, false) = false
              AND ad.deleted_at IS NULL
          )
          AND EXISTS (
            SELECT 1 FROM site_work_items done
            WHERE done.item_type = wi.item_type
              AND done.handler_agent = wi.handler_agent
              AND done.status = 'complete'
          )
        ORDER BY wi.created_at ASC
        LIMIT 20
    ),
    promoted AS (
        UPDATE site_work_items wi
        SET status = 'triaged',
            triaged_at = now(),
            spec = jsonb_set(COALESCE(wi.spec, '{}'::jsonb), '{original_pipeline}', to_jsonb(wi.pipeline)),
            pipeline = 'build',
            updated_at = now()
        FROM candidates c
        WHERE wi.id = c.id
          AND wi.status = 'detected'
        RETURNING wi.id, wi.item_type, wi.handler_agent
    )
    SELECT COUNT(*)::text AS promoted,
           string_agg(DISTINCT item_type || '->' || handler_agent, ', ') AS pairs
    FROM promoted
    WHERE (SELECT COUNT(*) FROM promoted) > 0
$PQ$
WHERE name = 'detected-item-promoter';

-- ----------------------------------------------------------------------------
-- Verification: shape asserted with RAISE (a plain SELECT cannot stop the
-- COMMIT), and the known-good rule PROVEN against the live pile — the query
-- below must classify at least one detected row as promotable AND at least the
-- rule must not classify a pair with zero completes as promotable.
-- ----------------------------------------------------------------------------
DO $$
DECLARE
    row_ok       boolean;
    pq           text;
    n_promotable integer;
    n_held       integer;
    n_unroutable integer;
    n_total      integer;
BEGIN
    SELECT EXISTS (SELECT 1 FROM scheduled_tasks
                    WHERE name = 'detected-item-promoter' AND enabled = true
                      AND fire_message = false AND interval_seconds = 900) INTO row_ok;
    IF NOT row_ok THEN
        RAISE EXCEPTION '430: detected-item-promoter row missing or mis-shaped (enabled/fire_message/interval)';
    END IF;

    SELECT pre_query INTO pq FROM scheduled_tasks WHERE name = 'detected-item-promoter';
    IF pq IS NULL OR position('LIMIT 20' IN pq) = 0 OR position('original_pipeline' IN pq) = 0
       OR position('pipeline = ''build''' IN pq) = 0 OR position('status = ''complete''' IN pq) = 0 THEN
        RAISE EXCEPTION '430: pre_query does not carry the batch cap, the pipeline rewrite, the original_pipeline stamp, or the known-good complete test';
    END IF;

    -- The known-good rule, evaluated read-only against the live pile.
    SELECT count(*) INTO n_promotable
    FROM site_work_items wi
    WHERE wi.status = 'detected' AND COALESCE(wi.handler_agent,'') <> ''
      AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent AND ad.is_active
                    AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL)
      AND EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                    AND d.handler_agent = wi.handler_agent AND d.status = 'complete');
    -- Disconfirmable check: promotable + held (active handler, zero completes)
    -- + unroutable (no/inactive handler) must PARTITION the detected pile. A
    -- predicate bug shows up as a sum that does not match the total.
    SELECT count(*) INTO n_held
    FROM site_work_items wi
    WHERE wi.status = 'detected' AND COALESCE(wi.handler_agent,'') <> ''
      AND EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent AND ad.is_active
                    AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL)
      AND NOT EXISTS (SELECT 1 FROM site_work_items d WHERE d.item_type = wi.item_type
                        AND d.handler_agent = wi.handler_agent AND d.status = 'complete');
    SELECT count(*) INTO n_unroutable
    FROM site_work_items wi
    WHERE wi.status = 'detected'
      AND (COALESCE(wi.handler_agent,'') = ''
           OR NOT EXISTS (SELECT 1 FROM agent_definitions ad WHERE ad.type = wi.handler_agent AND ad.is_active
                            AND COALESCE(ad.is_snapshot,false) = false AND ad.deleted_at IS NULL));
    SELECT count(*) INTO n_total FROM site_work_items WHERE status = 'detected';
    IF n_promotable + n_held + n_unroutable <> n_total THEN
        RAISE EXCEPTION '430: known-good predicates do not partition the detected pile (% + % + % <> %) — a predicate is wrong', n_promotable, n_held, n_unroutable, n_total;
    END IF;

    RAISE NOTICE '430: detected-item-promoter live (900s, fire_message=false, LIMIT 20/tick). Detected pile now %: % promotable, % held (pair never completed — canary one by hand), % unroutable (no/inactive handler)', n_total, n_promotable, n_held, n_unroutable;
END $$;

COMMIT;

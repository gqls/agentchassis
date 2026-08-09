-- ============================================================================
-- 342_thunder_orphan_scan.sql
--
-- thunder-orphan-scan: the reconcile the reaper cannot do.
--
-- The reaper reads only thunder_instances, and provision_action.go INSERTs
-- that row AFTER the vendor instance is up — so a crash in that window, or
-- any hand-created instance, leaves a box billing at Thunder that no
-- automated check can see (the orphan gap; bugs_open/186 "Related" section,
-- finetuning_uk_service RUNBOOK §1b). This scan asks Thunder for its own
-- view of the account and compares:
--
--   scheduled_tasks (thunder-orphan-scan, every 6h, no pre_query = always fires)
--      → thunder-orphan-scan agent_definition (task mode, in-chassis)
--      → step 1 dispatch_thunder_list  → thunder-adapter list_instances
--        (read-only GET /instances/list, awaited)
--      → step 2 reconcile_thunder_instances → classify + file work items
--      → complete
--
-- Mismatches billing at Thunder are filed as site_work_items rows:
--   site   system.internal (the platform-level site needs_diagnosis uses)
--   item_type 'thunder_orphan', item_key 'thunder_orphan:<vendor id>'
--   severity high, status detected, pipeline maintenance
-- Dedup: idx_swi_dedup suppresses refiling while an item is open.
-- Ghost rows (live row, no vendor instance) are reported in the run result
-- and logged, not filed — nothing is billing, and the decommission path
-- self-heals them.
--
-- ⚠ ORDERING: apply ONLY AFTER both images carrying the actions are live —
--   agent-chassis (dispatch_thunder_list, reconcile_thunder_instances) and
--   thunder-adapter (list_instances handler). A seed naming an unregistered
--   action fails at runtime. Verify first:
--     chassis:  kubectl exec <chassis-pod> -- sh -c \
--               'strings /app/agent-chassis | grep -c reconcile_thunder_instances'
--               → non-zero
--     adapter:  fire one manual list_instances request, or check the
--               adapter pod's strings for handleListInstances.
--
-- ⚠ The scan shares concurrency_group 'thunder-lifecycle' with the reaper
--   (max_concurrent 1 each): a scan never runs while a reap is mid-flight,
--   so it cannot snapshot a decommission halfway through.
--
-- ⚠ The scheduler treats NO pre_query as "always fire" (cmd/scheduler/
--   main.go:191 — pre_query is only consulted when non-empty). This task
--   wants that: the whole point is looking when we BELIEVE nothing is there.
--
-- Numbering: sql_for_agents is not schema_migrations-ledgered; the check is
-- the directory listing + git log at commit time. Verified 2026-08-08 evening:
-- 342 carries only this file (a concurrent lane took 343 — no collision; the
-- doubled 340/341 pairs in this directory are exactly the race this check is
-- for). The DO/RAISE block before COMMIT makes a botched apply fail loudly
-- instead of committing a half-seeded pipeline.
--
-- Verification queries + the manual drill are at the bottom.
-- ============================================================================

-- Defensive: clear any sticky aborted transaction left by a previous
-- half-run in the same psql session (harmless "no transaction" warning
-- otherwise). Rollback sidecar for the seeded rows:
-- 342_thunder_orphan_scan_ROLLBACK.sql (never applied by the runner).
ROLLBACK;

BEGIN;

-- ── 1. thunder-orphan-scan agent_definition ──

INSERT INTO agent_definitions (
    type, display_name, description,
    category, agent_category, status,
    default_config,
    capabilities,
    is_active, version
)
VALUES (
           'thunder-orphan-scan',
           'Thunder Orphan Scan',
           'Every 6h, fetches Thunder''s own instance list and reconciles it '
               || 'against thunder_instances. An instance billing at Thunder with '
               || 'no live row here is invisible to the reaper and every other '
               || 'automated check — this scan files it as a thunder_orphan work '
               || 'item on system.internal. Read-and-report only: no remediation.',
           'specialist',
           'executor',
           'experimental',
           jsonb_build_object(
                   'processing_mode', 'task',
                   'workflow', jsonb_build_object(
                           'start_step', 'dispatch_list',
                           'processing_mode', 'task',
                           'timeout_seconds', 120,
                           'steps', jsonb_build_object(
                                   -- ⚠ output_field lives at STEP level, NOT inside config.
                                   -- models.Step parses it from stepMap["output_field"]
                                   -- (processor.go:434); a config-nested copy is INERT — the
                                   -- coordinator stores an awaited response under the STEP
                                   -- NAME plus step-level output_field only. The reaper's
                                   -- own seed (028/114) carries the inert config-nested form,
                                   -- which is how this file's first version inherited the
                                   -- mistake (CORRECTED 2026-08-09 after the first live run
                                   -- failed at reconcile: collected_data had dispatch_list
                                   -- but no thunder_list).
                                   'dispatch_list', jsonb_build_object(
                                           'action', 'dispatch_thunder_list',
                                           'description', 'Ask thunder-adapter for Thunder''s own view of '
                                               || 'the account (GET /instances/list, awaited).',
                                           'output_field', 'thunder_list',
                                           'config', jsonb_build_object(
                                                   'timeout_seconds', 60
                                                     ),
                                           'next_step', 'reconcile'
                                                    ),
                                   'reconcile', jsonb_build_object(
                                           'action', 'reconcile_thunder_instances',
                                           'description', 'Compare the vendor list against thunder_instances; '
                                               || 'file thunder_orphan work items for instances billing '
                                               || 'with no live row; report ghosts.',
                                           'output_field', 'reconcile_result',
                                           'config', jsonb_build_object(
                                                   'list_field', 'thunder_list',
                                                   'grace_minutes', 30
                                                     ),
                                           'next_step', 'complete'
                                                ),
                                   'complete', jsonb_build_object(
                                           'action', 'complete_workflow',
                                           'config', jsonb_build_object(
                                                   'output_field', 'orphan_scan_summary'
                                                     )
                                               )
                                    )
                               )
           ),
           '["lifecycle", "thunder", "reconcile"]'::jsonb,
           true,
           1
       )
    ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                       description = EXCLUDED.description,
                                       category = EXCLUDED.category,
                                       agent_category = EXCLUDED.agent_category,
                                       status = EXCLUDED.status,
                                       default_config = EXCLUDED.default_config,
                                       capabilities = EXCLUDED.capabilities,
                                       is_active = EXCLUDED.is_active,
                                       updated_at = NOW();


-- ── 2. scheduled_tasks row ──
-- No pre_query: the task fires every 6h unconditionally. Gating it on our
-- own tables would rebuild the exact blindness this scan exists to remove.

INSERT INTO scheduled_tasks (
    name, description,
    interval_seconds, target_agent_type, target_topic,
    concurrency_group, max_concurrent, timeout_seconds,
    enabled
)
VALUES (
           'thunder-orphan-scan',
           'Every 6h, reconciles Thunder''s instance list against '
               || 'thunder_instances and files thunder_orphan work items for '
               || 'instances billing at Thunder that our automation cannot see.',
           21600,                                    -- 6 hours
           'thunder-orphan-scan',
           'system.agent.generic.requests',
           'thunder-lifecycle',                      -- shared with thunder-reaper
           1,
           180,
           true
       )
    ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
                              interval_seconds = EXCLUDED.interval_seconds,
                              target_agent_type = EXCLUDED.target_agent_type,
                              target_topic = EXCLUDED.target_topic,
                              concurrency_group = EXCLUDED.concurrency_group,
                              max_concurrent = EXCLUDED.max_concurrent,
                              timeout_seconds = EXCLUDED.timeout_seconds,
                              enabled = EXCLUDED.enabled,
                              updated_at = NOW();

-- ── 3. doc_notes write-back (council round 1, tooling_provenance seat) ──
-- The durable Postgres-side trace of this pipeline, keyed so the next
-- session touching thunder-lifecycle or this scan finds the reasoning
-- without re-deriving it from bug files. Measured before adding: doc_notes
-- carried NO pipeline-subject row for thunder-adapter, thunder-reaper or
-- thunder-lifecycle (only 4 landmine rows) — this is the first.

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
SELECT 'pipeline', 'thunder-orphan-scan',
       'thunder-orphan-scan (FTW-042): 6-hourly vendor-truth reconcile. '
           || 'dispatch_thunder_list asks thunder-adapter for GET /instances/list '
           || '(read-only), reconcile_thunder_instances compares against '
           || 'thunder_instances. Orphans (billing at Thunder, no live row) are '
           || 'filed as thunder_orphan work items on system.internal via '
           || 'insertWorkItem (sole producer: reconcile_thunder_instances; '
           || 'item_key thunder_orphan:<vendor id>; NO automated consumer — the '
           || 'human path is the Error log + RUNBOOK finetuning_uk_service '
           || 'section 1b query). Ghosts (live row, nothing at Thunder) are reported, '
           || 'not filed. Tuning knob: grace_minutes (step config, default 30) '
           || 'absorbs the provision INSERT-after-up window; unknown createdAt '
           || 'cannot hide behind it. SHARES the thunder-lifecycle concurrency '
           || 'group with thunder-reaper (max_concurrent 1 each): a scan never '
           || 'snapshots a reap mid-flight, and can delay a reaper tick by at '
           || 'most its own runtime (one HTTP GET + one small query). '
           || 'Read-and-report only: no DeleteInstance, no thunder_config writes.',
       '["thunder", "reconcile", "work-dispatch"]'::jsonb,
       'sql_for_agents/342_thunder_orphan_scan.sql',
       'finetuning_uk_service lane'
WHERE NOT EXISTS (
    SELECT 1 FROM doc_notes
    WHERE subject_type = 'pipeline' AND subject_key = 'thunder-orphan-scan'
);

-- ── 4. Verify INSIDE the transaction, or the COMMIT proves nothing ──
-- A verify block of SELECTs cannot stop the COMMIT (ON_ERROR_STOP ignores a
-- non-empty result) — DO/RAISE is the form that can.

DO $verify$
DECLARE
    def_start text;
    task_ok   boolean;
    note_ok   boolean;
BEGIN
    -- The filing path resolves its site by domain; without this row every
    -- orphan filing silently degrades to log-and-report (council r2,
    -- editquality). Refuse to seed a scan whose core deliverable cannot fire.
    PERFORM 1 FROM sites WHERE domain = 'system.internal';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'no sites row with domain=system.internal — thunder_orphan work items would have nowhere to go';
    END IF;

    SELECT default_config->'workflow'->>'start_step' INTO def_start
    FROM agent_definitions
    WHERE type = 'thunder-orphan-scan' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF def_start IS DISTINCT FROM 'dispatch_list' THEN
        RAISE EXCEPTION 'thunder-orphan-scan agent_definition wrong or missing (start_step=%)', def_start;
    END IF;

    -- output_field must be STEP-LEVEL (a config-nested copy is inert — the
    -- 2026-08-09 correction above; this assert stops the inert form from
    -- ever re-applying silently).
    PERFORM 1 FROM agent_definitions
    WHERE type = 'thunder-orphan-scan' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config->'workflow'->'steps'->'dispatch_list'->>'output_field' = 'thunder_list'
      AND default_config->'workflow'->'steps'->'reconcile'->>'output_field' = 'reconcile_result';
    IF NOT FOUND THEN
        RAISE EXCEPTION 'thunder-orphan-scan output_field is not step-level — the inert config-nested form is back';
    END IF;

    SELECT enabled AND interval_seconds = 21600
           AND concurrency_group = 'thunder-lifecycle'
           AND pre_query IS NULL
    INTO task_ok
    FROM scheduled_tasks WHERE name = 'thunder-orphan-scan';
    IF task_ok IS DISTINCT FROM true THEN
        RAISE EXCEPTION 'thunder-orphan-scan scheduled_tasks row wrong or missing';
    END IF;

    SELECT EXISTS (
        SELECT 1 FROM doc_notes
        WHERE subject_type = 'pipeline' AND subject_key = 'thunder-orphan-scan'
    ) INTO note_ok;
    IF NOT note_ok THEN
        RAISE EXCEPTION 'thunder-orphan-scan doc_notes row missing';
    END IF;
END
$verify$;

COMMIT;


-- ============================================================================
-- Verification (run after applying)
-- ============================================================================
--
-- SELECT type, status, is_active,
--        default_config->'workflow'->>'start_step' AS start_step
-- FROM agent_definitions WHERE type = 'thunder-orphan-scan';
-- -- Expect: thunder-orphan-scan | experimental | t | dispatch_list
--
-- SELECT name, interval_seconds, enabled, concurrency_group, pre_query IS NULL AS no_pq
-- FROM scheduled_tasks WHERE name = 'thunder-orphan-scan';
-- -- Expect: thunder-orphan-scan | 21600 | t | thunder-lifecycle | t
--
-- ── First-run verification (the scan on a clean account is a NO-OP that
--    must still prove it LOOKED — a 0-findings result has two causes with
--    opposite meanings) ──
--
-- 1. Kick it: UPDATE scheduled_tasks SET last_triggered_at = NULL
--             WHERE name = 'thunder-orphan-scan';
-- 2. SELECT current_step, status,
--           collected_data->'reconcile_result' AS result
--    FROM orchestration_states
--    WHERE owner_agent_type = 'thunder-orphan-scan'      -- owner_agent_type, NOT agent_type
--    ORDER BY created_at DESC LIMIT 1;
--    -- Expect status COMPLETED and a result whose vendor_billing /
--    -- db_rows / matched counts are TRUTHFUL for the account right now —
--    -- check them against RUNBOOK §1b's manual API call, same day.
--
-- ── Orphan drill (proves the FILING path, ~$0 — no vendor call touches
--    the synthetic id) ──
--
-- The scan flags a vendor instance with no row; we cannot fake Thunder's
-- side without a real (billable) instance, but the GHOST direction and the
-- work-item write CAN be drilled by pointing the classifier at a synthetic
-- live row (reported, not filed) — or wait for the first real provision
-- (Phase 0) and check matched=1. The filing INSERT itself is unit-tested
-- at the classification boundary and exercised the first time a real
-- orphan appears; the honest status for the filing path until then is
-- BUILT, NOT YET FIRED IN ANGER (concept register entry says so).
-- ============================================================================

-- 0NN_diagnose_dispatch_loop.sql — automatic dispatch for pipeline='diagnose'.
-- 2026-07-09. Renumber 0NN when filing. Applies to clients_db.
--
-- Creates: agent_definitions row `diagnose-dispatch-loop` (+ image columns
--          copied from a live donor, per the seed gotcha)
--          scheduled_tasks row `diagnose-pipeline-trigger`, ENABLED = FALSE.
--
-- ════════════════════════════════════════════════════════════════════════════
-- WHY THIS LOOP DOES NOT USE load_work_items OR claim_work_item
-- ════════════════════════════════════════════════════════════════════════════
-- Everything below was read from the live definitions on 2026-07-09, not
-- assumed. The `diagnose` namespace cannot ride the standard claim path,
-- because the standard path cannot tell pipelines apart:
--
--  * build-dispatch-loop's `load_items` step is configured with ONLY
--    {site_id, max_items} — NO item_pipeline filter. It claims any item on the
--    site it is handed, whatever the pipeline.
--
--  * build-pipeline-trigger's `find_dispatchable_site` query has NO pipeline
--    filter either, despite its description saying "a site with pending build
--    items". Any site holding a 'triaged'/'approved' item of ANY pipeline is
--    selected. (This is how the `maintenance` pipeline's items get dispatched
--    at all — by accident. So the tempting one-key fix, adding
--    item_pipeline='build' to build-dispatch-loop, would ORPHAN the maintenance
--    pipeline. It is the builder thread's call, not ours. Reported, not fixed.)
--
--  * triage_detect_items promotes `WHERE site_id = $1 AND status = 'detected'`
--    — no item_type, no pipeline filter — and REWRITES pipeline to 'build'.
--    So 'detected' is not a safe parking state either. Its comment claims "the
--    dispatch loop (which filters item_pipeline='build')". The dispatch loop
--    does not. Same comment-vs-code family as the pilot bug.
--
--  * claim_work_item claims only `status IN ('triaged','approved')` — exactly
--    the statuses that expose an item to the two unfiltered readers above.
--
-- CONSEQUENCE: a needs_diagnosis item must never hold 'triaged', 'approved',
-- 'detected' — or 'claimed'. It uses TWO private statuses, and every sweep
-- filters on explicit status values, so both are INERT BY CONSTRUCTION rather
-- than by luck of anchor-site choice:
--
--     queued   → status = 'awaiting_diagnosis'
--     in-flight→ status = 'diagnosing'
--
--     claim_work_item        triaged|approved            → inert (both)
--     load_work_items        triaged|approved            → inert (both)
--     find_dispatchable_site triaged|approved            → inert (both)
--     triage_detect_items    detected                    → inert (both)
--     feasibility-recheck    blocked                     → inert (both)
--     stale-work-item-reaper triaged AND pipeline=build  → inert (both)
--     claimed-item-timeout   claimed                     → inert (both)
--     idx_swi_dedup          non-terminal ⇒ still "open" → idempotent intake ✓
--
-- WHY NOT status='claimed' FOR THE IN-FLIGHT STATE (owner's question, 2026-07-09).
-- The claim itself is NOT optional and is NOT the handler's job: on this platform
-- the DISPATCHER claims. Only build-dispatch-loop (and this loop) reference
-- claim_work_item; page-build-handler neither claims nor completes its own item —
-- the dispatch loop calls complete_work_item on its behalf (see the guard comment
-- at load_work_item_actions.go:750). Without an atomic move out of the queue
-- state, the next 60s tick re-dispatches the same 26-minute LLM run.
-- But 'claimed' is the WRONG VALUE to move it into, because it re-enters the very
-- sweep surface 'awaiting_diagnosis' exists to escape:
--   (a) claimed-item-timeout resets any 40-minute-old claim to 'triaged' —
--       handing a slow diagnosis straight to the build dispatcher;
--   (b) find_dispatchable_site excludes any site holding a 'claimed' item, so a
--       70-minute diagnosis would BLOCK build dispatch for system.internal for
--       its whole duration. Cross-pipeline interference, for free, unasked.
-- (The 15-minute auto-complete branch is safe either way: it is gated on
-- item_type IN (needs_content_page, page_rerender, needs_design).)
-- So: 'diagnosing'. claimed_by/claimed_at are still stamped for audit, and the
-- reap_stuck step below does the cleanup claimed-item-timeout would have done.
--
-- max_attempts = 1 on the intake item is kept, but as SEMANTICS, not as the
-- safety net it was in the first draft: silently auto-retrying a 26-minute LLM
-- loop is not a thing anyone wants.
--
-- KNOWN LIMITATION (documented, not hidden): the auto path does not forward
-- `seed_scope`. query_database flattens every column to text, and
-- ExtractStringListHelper accepts only []interface{}/[]string — a JSON string
-- would silently parse to nil. The manual 090 route passes a real JSON array
-- and works. Without a seed scope, diagnose_assemble_bundle's fallback chain
-- uses lookup_code_symbols' code_results, which is the designed default.

BEGIN;

-- Snapshot before any agent_definitions update (standing gotcha). No-op on the
-- first apply, when the row does not yet exist.
SELECT snapshot_agent('diagnose-dispatch-loop', 'pre-update: in-flight status diagnosing + reap_stuck')
WHERE EXISTS (
    SELECT 1 FROM agent_definitions
    WHERE type = 'diagnose-dispatch-loop'
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
);

-- ── the dispatch loop agent ──────────────────────────────────────────────────
-- Image/infra columns are COPIED FROM A LIVE DONOR (build-dispatch-loop) rather
-- than hand-typed: the seed gotcha. topics use {type} templating, so they are
-- correct verbatim.
INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    is_active, version, capabilities,
    image_repository, image_tag, command, resources, topics, health_config, env_vars,
    default_config
)
SELECT
    'diagnose-dispatch-loop',
    'Diagnose Dispatch Loop',
    'Claims one awaiting_diagnosis work item (pipeline=diagnose) and runs it through diagnose-orchestrator. Uses its own atomic claim because the shared claim path cannot distinguish pipelines.',
    'orchestrator', 'coordinator', 'experimental',
    true, 1, '["dispatch", "orchestration", "work-items", "diagnose"]'::jsonb,
    d.image_repository, d.image_tag, d.command, d.resources, d.topics, d.health_config, d.env_vars,
    jsonb_build_object('workflow', jsonb_build_object(
      'start_step', 'reap_stuck',
      'processing_mode', 'orchestrator',
      'timeout_seconds', 4200,
      'steps', jsonb_build_object(

        -- Because 'diagnosing' is invisible to claimed-item-timeout (by design),
        -- WE must reap our own dead runs. 75 minutes is past this workflow's own
        -- 4200s timeout, so a row still 'diagnosing' by then means the pod died.
        -- attempt_count is bumped so max_attempts=1 keeps it terminal.
        'reap_stuck', jsonb_build_object(
          'action', 'query_database',
          'description', 'Fail any diagnosis whose pod died. We own this because diagnosing is inert to claimed-item-timeout.',
          'output_field', 'reaped',
          'next_step', 'claim_item',
          'config', jsonb_build_object(
            'output_format', 'object',
            'query',
              'UPDATE site_work_items SET status = ''failed'', attempt_count = attempt_count + 1, ' ||
              '       error = ''diagnosis exceeded 75m — handler pod likely died'', claimed_at = NULL ' ||
              'WHERE pipeline = ''diagnose'' AND status = ''diagnosing'' ' ||
              '  AND claimed_at < NOW() - INTERVAL ''75 minutes'' ' ||
              'RETURNING id::text'
          )
        ),

        -- Atomic claim-and-read in ONE statement. FOR UPDATE SKIP LOCKED makes
        -- concurrent ticks safe; the UPDATE...RETURNING hands us the envelope
        -- so no second read is needed. query_database(output_format=object)
        -- flattens the first row to top level, so `claimed.symptom` resolves.
        -- Moves to 'diagnosing', NOT 'claimed' — see the header note on why.
        'claim_item', jsonb_build_object(
          'action', 'query_database',
          'description', 'Atomically take ONE awaiting_diagnosis item into diagnosing. Not claim_work_item: that claims only triaged/approved, the statuses that expose an item to build-dispatch-loop.',
          'output_field', 'claimed',
          'next_step', 'check_claimed',
          'config', jsonb_build_object(
            'output_format', 'object',
            'query',
              'UPDATE site_work_items SET status = ''diagnosing'', claimed_by = ''diagnose-dispatch-loop'', claimed_at = NOW() ' ||
              'WHERE id = (SELECT id FROM site_work_items ' ||
              '            WHERE pipeline = ''diagnose'' AND item_type = ''needs_diagnosis'' ' ||
              '              AND status = ''awaiting_diagnosis'' AND attempt_count < max_attempts ' ||
              '            ORDER BY priority ASC, created_at ASC ' ||
              '            FOR UPDATE SKIP LOCKED LIMIT 1) ' ||
              'RETURNING id::text AS work_item_id, handler_agent, ' ||
              '          spec->>''symptom'' AS symptom, spec->>''owner'' AS owner, ' ||
              '          spec->>''repo'' AS repo, spec->>''ref'' AS ref, ' ||
              '          spec->>''site_id'' AS target_site_id, spec->>''runtime_site'' AS runtime_site, ' ||
              '          spec->>''subject_type'' AS subject_type, spec->>''subject_key'' AS subject_key, ' ||
              '          spec->>''correlation_id'' AS correlation_id'
          )
        ),

        'check_claimed', jsonb_build_object(
          'action', 'conditional',
          'description', 'Nothing queued (count = 0) is the normal case: tell the scheduler and finish.',
          'config', jsonb_build_object(
            'condition', 'claimed.count > 0',
            'then_step', 'spawn_handler',
            'else_step', 'notify_scheduler'
          )
        ),

        -- handler_agent is diagnose-orchestrator: the SAME agent the manual 090
        -- route targets, so manual and automatic dispatch exercise one code
        -- path. It spawns the diagnose-agent pod, which is where the spawn gate
        -- injects GITHUB_READ_TOKEN (isRepoCloningAgent, spawn_actions.go).
        'spawn_handler', jsonb_build_object(
          'action', 'spawn_agent',
          'description', 'Spawn the handler named on the item (diagnose-orchestrator).',
          'output_field', 'handler_spawned',
          'next_step', 'call_handler',
          'config', jsonb_build_object(
            'role', 'handler',
            'agent_type_field', 'claimed.handler_agent',
            'error_step', 'mark_failed'   -- INSIDE config: step-level error_step is silently ignored (001 §16)
          )
        ),

        -- site_id is mapped from claimed.target_site_id (the site UNDER
        -- DIAGNOSIS, out of spec) — never from the item's own site_id column,
        -- which is the system.internal anchor and would be meaningless here.
        -- timeout 2100 > diagnose-orchestrator's internal 1800s call.
        'call_handler', jsonb_build_object(
          'action', 'call_agent',
          'description', 'Run the diagnosis. Envelope comes from spec, not from the item anchor.',
          'output_field', 'handler_result',
          'next_step', 'mark_complete',
          'config', jsonb_build_object(
            'target_role', 'handler',
            'timeout_seconds', 2100,
            'error_step', 'mark_failed',
            'input_mapping', jsonb_build_object(
              'symptom',         'claimed.symptom',
              'owner?',          'claimed.owner',
              'repo?',           'claimed.repo',
              'ref?',            'claimed.ref',
              'site_id?',        'claimed.target_site_id',
              'runtime_site?',   'claimed.runtime_site',
              'subject_type?',   'claimed.subject_type',
              'subject_key?',    'claimed.subject_key',
              'correlation_id?', 'claimed.correlation_id',
              'work_item_id',    'claimed.work_item_id'
            )
          )
        ),

        'mark_complete', jsonb_build_object(
          'action', 'complete_work_item',
          'description', 'Mark the intake item complete, carrying the diagnosis as its result.',
          'output_field', 'item_completed',
          'next_step', 'notify_scheduler',
          'config', jsonb_build_object(
            'work_item_id', 'claimed.work_item_id',
            'result', 'handler_result'
          )
        ),

        'mark_failed', jsonb_build_object(
          'action', 'fail_work_item',
          'description', 'Handler died or timed out. max_attempts=1 means this is terminal.',
          'output_field', 'item_failed',
          'next_step', 'notify_scheduler',
          'config', jsonb_build_object(
            'work_item_id', 'claimed.work_item_id',
            'error_message', 'diagnose handler failed or timed out'
          )
        ),

        'notify_scheduler', jsonb_build_object(
          'action', 'query_database',
          'description', 'Tell the scheduler this execution finished so it can fire again.',
          'output_field', 'scheduler_notified',
          'next_step', 'complete',
          'config', jsonb_build_object(
            'output_format', 'object',
            'query', 'UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = ''diagnose-pipeline-trigger'''
          )
        ),

        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Dispatch tick complete.',
          'config', jsonb_build_object('output_fields', jsonb_build_array('claimed', 'handler_result', 'reaped'))
        )
      )
    ))
FROM agent_definitions d
WHERE d.type = 'build-dispatch-loop'
  AND COALESCE(d.is_snapshot, false) = false
  AND d.deleted_at IS NULL
ON CONFLICT (type, version) DO UPDATE
   SET default_config = EXCLUDED.default_config,
       description    = EXCLUDED.description,
       capabilities   = EXCLUDED.capabilities,
       updated_at     = now();

-- ── the scheduler entry — DISABLED until the chassis image ships and the
--    benchmark's blinding is confirmed. Enable deliberately:
--      UPDATE scheduled_tasks SET enabled = true WHERE name = 'diagnose-pipeline-trigger';
--
-- pre_query gates the tick: fire only when something is actually queued, the
-- same discipline build-pipeline-trigger uses (HAVING COUNT(*) > 0 ⇒ no rows ⇒
-- no dispatch). max_concurrent = 1: one diagnosis at a time; each spawns pods
-- and runs an LLM loop for tens of minutes.
INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds
) VALUES (
    'diagnose-pipeline-trigger',
    'Fires diagnose-dispatch-loop when a needs_diagnosis item is queued (pipeline=diagnose, status=awaiting_diagnosis).',
    60,
    'diagnose-dispatch-loop',
    'system.agent.generic.requests',
    jsonb_build_object(
      'action', 'orchestrate',
      'config', jsonb_build_object('agent_type', 'diagnose-dispatch-loop'),
      'input_data', '{}'::jsonb
    ),
    'diagnose-dispatch',
    1,
    'SELECT COUNT(*)::text AS queued_diagnoses FROM site_work_items ' ||
    'WHERE pipeline = ''diagnose'' AND item_type = ''needs_diagnosis'' ' ||
    '  AND status = ''awaiting_diagnosis'' AND attempt_count < max_attempts ' ||
    'HAVING COUNT(*) > 0',
    false,
    3600
)
ON CONFLICT DO NOTHING;

COMMIT;

-- Rollback (manual):
--   DELETE FROM scheduled_tasks   WHERE name = 'diagnose-pipeline-trigger';
--   DELETE FROM agent_definitions WHERE type = 'diagnose-dispatch-loop' AND version = 1;

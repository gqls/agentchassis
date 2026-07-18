-- SEED — V4 evidence freshness on a cadence (claims-verification layer)
--
-- APPLY ONLY AFTER a chassis image carrying `refresh_evidence_base` is deployed
-- and pod-verified. CLAUDE.md: "Image first, then seeds (a seed naming an
-- unregistered action fails at runtime)." Verify first, against the POD:
--
--   kubectl -n ai-persona-system exec <agent-chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c refresh_evidence_base'
--
-- Expect a non-zero count. If it is 0 the action is not in the image and this
-- seed must NOT be applied.
--
-- Cadence rationale: the facts are counts over slow-moving business data
-- (verified businesses, agent definitions, deployed sites, feed items). Daily
-- is well inside any tolerance and keeps `verified_at` honest without hammering
-- the query layer; the action is idempotent and safely re-runnable, and it
-- writes nothing when nothing moved. 86400s = daily.
--
-- Scope: the action sweeps EVERY site holding a current evidence_base spec when
-- no site_id is supplied — today that is leopardessconsulting.co.uk alone, and
-- each site opted in later is picked up with no seed change.
--
-- Drift terminates at human review (`stale_evidence`, needs_human_review, no
-- automated handler): a number moving is a fact about the world, but changing
-- published copy because of it is a human decision (SPEC §10 open question 3).

INSERT INTO scheduled_tasks (
    name, description, interval_seconds, target_agent_type, target_topic,
    input_data, concurrency_group, max_concurrent, enabled, timeout_seconds
)
SELECT
    'evidence-freshness',
    'V4 of the claims-verification layer: re-run live-verifiable evidence_base facts, re-sync value/verified_at, regenerate the V2 writer whitelist, raise stale_evidence for human ruling when a live value drifts outside the tolerance its published wording allows (including under-claiming).',
    86400,
    'generic',
    'system.agent.generic.requests',
    jsonb_build_object(
        'config', jsonb_build_object(
            'agent_type', 'generic',
            'workflow', jsonb_build_object(
                'start_step', 'refresh_evidence',
                'processing_mode', 'orchestrator',
                'timeout_seconds', 600,
                'steps', jsonb_build_object(
                    'refresh_evidence', jsonb_build_object(
                        'action', 'refresh_evidence_base',
                        'config', jsonb_build_object(),
                        'description', 'Sweep every site with an evidence base; no site_id = all of them',
                        'next_step', 'complete',
                        'output_field', 'refresh_result'
                    ),
                    'complete', jsonb_build_object(
                        'action', 'complete_workflow',
                        'config', jsonb_build_object('output_fields', jsonb_build_array('refresh_result'))
                    )
                )
            )
        )
    ),
    'evidence-freshness',
    1,
    true,
    600
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name = 'evidence-freshness');

-- ── Post-apply verification ────────────────────────────────────────────────
--
-- 1. The row exists and is enabled:
--
--   SELECT name, target_agent_type, interval_seconds, enabled
--   FROM scheduled_tasks WHERE name = 'evidence-freshness';
--
-- 2. After the first fire (within `interval_seconds`, or set
--    last_triggered_at = NULL to bring it forward), the register's dates move:
--
--   SELECT (f->>'id') AS fact, f->>'value' AS value, f->>'verified_at' AS verified
--   FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
--   WHERE ss.aspect = 'evidence_base' AND ss.is_current = true
--     AND f->'source' ? 'sql'
--   ORDER BY 1;
--
-- 3. The spec row records the pass in its notes, and `pinned` survives:
--
--   SELECT created_by, pinned, notes FROM site_specs
--   WHERE aspect = 'evidence_base' AND is_current = true;
--
--   Expect created_by = 'evidence-refresher', pinned = true, and a note of the
--   form 'V4 freshness pass: N live-verifiable fact(s) checked, ...'.
--
-- 4. Any drift raised for a human (empty is the healthy steady state):
--
--   SELECT summary, status, spec->'drifted' FROM site_work_items
--   WHERE item_type = 'stale_evidence' ORDER BY created_at DESC;
--
-- 5. The regenerated whitelist still reads as its human authors wrote it —
--    caveats intact, numbers current:
--
--   SELECT data->>'writer_block' FROM site_specs
--   WHERE aspect = 'evidence_base' AND is_current = true;
--
-- ── Rollback ───────────────────────────────────────────────────────────────
--
--   UPDATE scheduled_tasks SET enabled = false WHERE name = 'evidence-freshness';
--
-- Disabling is sufficient and lossless: the action writes a new spec revision
-- rather than mutating in place, so every pre-refresh evidence base remains in
-- site_specs history (is_current = false) and can be restored by hand.

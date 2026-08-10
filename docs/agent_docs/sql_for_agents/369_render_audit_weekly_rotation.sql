-- ============================================================================
-- 369 — render audit gets a cadence: weekly per site, one site per hourly tick
-- ============================================================================
-- Written 2026-08-10 (bugfix_122_contrast_ink_slots lane). This is "edit 8" of
-- the council-approved plan (PLAN_2026-08-06_contrast_ink_slots.md §3 candidate
-- 2; verdict c4d9c841): the whole audit chain has been live since v1.0.1257
-- (site → request_render_audit on the dedicated pod → write_render_audit_findings
-- → complete, filing contrast_failure items routed to css-patch-agent) and
-- NOTHING dispatches it — the bugs_open/083/093/115 shape: a mechanism made
-- correct and then guarded behind something that never runs. Total
-- contrast_failure items ever raised before this row: 4, all relojistas.com,
-- all 2026-08-04, all from one hand run.
--
-- WHY THIS SHAPE. Clones the proven site-discovery-rotation-* mechanism
-- (hourly tick, pre_query selects ONE due site, stamps site_discovery_rotation
-- keyed (site_id, agent_type), skips sites with claimed build items, 7-day due
-- window => weekly per site). "Weekly not daily" is the plan's own sizing.
-- A pre_query returning zero rows is a stamped no-op (scheduler main.go:198),
-- so a fully-audited fleet costs nothing.
--
-- WHY IT MATTERS, freshly demonstrated: the 2026-08-10 re-audit found the
-- sub-shape A defect re-materialised on dartsonline (info-card-grid, 6 firm
-- failures) within 2 days of the component swap that introduced it. Without a
-- cadence, that class is only ever found when this lane happens to re-run the
-- audit by hand.
--
-- Own concurrency group ('render-audit', cap 1): the audit runs on the
-- DEDICATED render-audit pod, so it must not compete with the discovery
-- agents' shared 'site-discovery' slot, and one in-flight audit at a time is
-- the pod's own sizing. timeout 1800s: 25-page cap x ~60s/page worst case,
-- and safely under the 3600s interval so a hung run cannot stack fires.
--
-- INERT UNTIL: nothing — scheduled_tasks is live config; the scheduler's next
-- tick picks it up. No image roll involved.
--
-- Rollback is at the foot of this file.

BEGIN;

DO $$
DECLARE
    n int;
BEGIN
    -- The target agent must exist, active, live — not a seed-file belief.
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'render-audit-agent' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '369: render-audit-agent live rows = % (expected exactly 1) — fix the agent before scheduling it, STOP', n;
    END IF;

    -- Idempotency: the unique name constraint would refuse anyway; refuse
    -- legibly and before the INSERT.
    PERFORM 1 FROM scheduled_tasks WHERE name = 'site-render-audit-rotation';
    IF FOUND THEN
        RAISE EXCEPTION '369: scheduled_tasks row ''site-render-audit-rotation'' already exists — STOP (already applied?)';
    END IF;

    INSERT INTO scheduled_tasks
        (name, description, interval_seconds, target_agent_type, target_topic,
         input_data, concurrency_group, max_concurrent, pre_query, enabled,
         timeout_seconds, fire_message)
    VALUES (
        'site-render-audit-rotation',
        'Weekly per-site render audit (bug 122 edit 8). Hourly tick; pre_query picks the single most-overdue site not audited in 7 days and not mid-build, stamps site_discovery_rotation, and the render-audit-agent measures every deployed page on the dedicated pod, filing firm contrast failures as contrast_failure -> css-patch-agent. Zero due sites = stamped no-op.',
        3600,
        'render-audit-agent',
        'system.agent.scheduled.requests',
        '{}'::jsonb,
        'render-audit',
        1,
        $preq$WITH due AS (
  SELECT s.id AS sid, s.domain
  FROM sites s
  LEFT JOIN site_discovery_rotation r
    ON r.site_id = s.id AND r.agent_type = 'render-audit-agent'
  WHERE s.status IN ('active', 'deployed')
    AND COALESCE(r.last_selected_at, '-infinity'::timestamptz) < now() - interval '7 days'
    AND NOT EXISTS (
      SELECT 1 FROM site_work_items wi
      WHERE wi.site_id = s.id AND wi.status = 'claimed' AND wi.pipeline = 'build')
  ORDER BY r.last_selected_at ASC NULLS FIRST, s.id
  LIMIT 1
), stamped AS (
  INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
  SELECT sid, 'render-audit-agent', now() FROM due
  ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at = EXCLUDED.last_selected_at
)
SELECT sid::text AS site_id, domain FROM due$preq$,
        true,
        1800,
        true
    );

    -- Post-conditions: exactly one row, enabled, pointing at the right agent,
    -- and it satisfies the no-inline-workflow constraint by construction.
    SELECT count(*) INTO n
      FROM scheduled_tasks
     WHERE name = 'site-render-audit-rotation'
       AND enabled AND target_agent_type = 'render-audit-agent'
       AND interval_seconds = 3600 AND timeout_seconds = 1800;
    IF n <> 1 THEN
        RAISE EXCEPTION '369: post-insert verification found % matching rows (expected 1)', n;
    END IF;

    RAISE NOTICE '369: site-render-audit-rotation scheduled — weekly per site, hourly tick, own concurrency group';
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK
-- ============================================================================
-- DELETE FROM scheduled_tasks WHERE name = 'site-render-audit-rotation';
-- (Optionally: DELETE FROM site_discovery_rotation WHERE agent_type = 'render-audit-agent';
--  — only if you want re-scheduling later to start from a clean rotation queue.)

-- 600_claims_audit_rotation.sql
--
-- bugs_open/380 slice S3: give the claims-auditor a clock. The auditor has existed since
-- 2026-07-17 and has run ONCE (llm_call_log, 2026-07-18, returned []) because nothing
-- dispatches it — no seed, no schedule, no spawner. "A registered-but-uncalled action is
-- not a mechanism" (the council's own REVISE on migration 590, whose rotation shape this
-- copies). The cadence question has been open, owner's-call, since 2026-07-16
-- (claims_verification PLAN/HANDOFF). OWNER DECISION 2026-08-24 (bugfix_380 lane plan D2):
-- 3600s tick, 7-day per-site window, never-audited sites first.
--
-- Shape notes, all inherited from 590 and 346:
--   * ONE site per tick, chosen and STAMPED in the same statement (site_discovery_rotation,
--     agent_type 'claims-auditor' — that table already hosts non-discovery agents; the
--     staleness check iterates its own hardcoded agent list, so a new agent_type cannot
--     create a false finding there).
--   * locked_at IS NULL — an owner HALT must not draw audits (and an audit files work
--     items, which held sites keep queued).
--   * No site is brought into existence: the pre_query yields site_id AND domain for an
--     EXISTING row; the auditor's ensure_site_record upserts by domain, so handing it the
--     canonical stored domain makes that an update of the row we selected.
--   * Only sites with a SHIPPED page: the auditor's own load_page_text reads pages with
--     build_status IN ('deployed','active') and non-empty rendered_html; a site with
--     nothing shipped would burn the tick on an empty audit.
--   * The stamp records SELECTION, not completion — coverage questions join the doc_notes
--     receipts 597 introduced ('pipeline'/'claims-audit'), never this stamp alone
--     (RUNBOOK_claims_fail_open.md carries the query).
--
-- Cost bound: ~28 eligible sites; first sweep completes in ~28 ticks (~28h); steady state
-- ~4 sites/day, ~120 Sonnet calls/month at ~10-25k input tokens each.
--
-- Apply AFTER the 597 hand-dispatch proof at garden-tools.uk (bugfix_380 plan step 8) —
-- the rotation is the amplifier, so the thing it amplifies is proven first.

BEGIN;

INSERT INTO scheduled_tasks
  (name, description, interval_seconds, target_agent_type, target_topic,
   input_data, concurrency_group, max_concurrent, pre_query, enabled, timeout_seconds, fire_message)
SELECT
  'claims-audit-rotation',
  'One claims-auditor pass per due site per tick (7-day window, never-audited first). bugs_open/380: the auditor previously had no caller at all.',
  3600,
  'claims-auditor',
  'system.agent.scheduled.requests',
  '{}'::jsonb,
  'claims-audit',
  1,
  $PQ$
WITH due AS (
  SELECT s.id AS sid, s.domain
  FROM sites s
  LEFT JOIN site_discovery_rotation r
    ON r.site_id = s.id AND r.agent_type = 'claims-auditor'
  WHERE s.status IN ('active', 'deployed')
    AND s.locked_at IS NULL
    AND COALESCE(r.last_selected_at, '-infinity'::timestamptz) < now() - interval '7 days'
    AND EXISTS (
      SELECT 1 FROM pages p
      WHERE p.site_id = s.id
        AND p.build_status IN ('deployed','active'))
    AND NOT EXISTS (
      SELECT 1 FROM site_work_items wi
      WHERE wi.site_id = s.id AND wi.status = 'claimed' AND wi.pipeline = 'build')
  ORDER BY r.last_selected_at ASC NULLS FIRST, s.id
  LIMIT 1
), stamped AS (
  INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
  SELECT sid, 'claims-auditor', now() FROM due
  ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at = EXCLUDED.last_selected_at
)
SELECT sid::text AS site_id, domain FROM due
$PQ$,
  true,
  900,
  true
WHERE NOT EXISTS (SELECT 1 FROM scheduled_tasks WHERE name='claims-audit-rotation');

-- Verify: DO/RAISE (bare SELECTs verify nothing).
DO $$
DECLARE n int; v_target text; v_enabled boolean; v_interval int;
BEGIN
  SELECT count(*), max(target_agent_type), bool_or(enabled), max(interval_seconds)
    INTO n, v_target, v_enabled, v_interval
    FROM scheduled_tasks WHERE name='claims-audit-rotation';
  IF n <> 1 THEN RAISE EXCEPTION '600: expected 1 rotation task, found %', n; END IF;
  IF v_target IS DISTINCT FROM 'claims-auditor' THEN
    RAISE EXCEPTION '600: rotation targets %', v_target;
  END IF;
  IF NOT v_enabled THEN RAISE EXCEPTION '600: rotation task is disabled'; END IF;
  IF v_interval <> 3600 THEN RAISE EXCEPTION '600: interval is %, owner ruled 3600', v_interval; END IF;

  -- The auditor must exist, be active, and no longer carry the opt-in gate (597 first).
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND NOT ((default_config #> '{workflow,steps}') ? 'check_opted_in');
  IF n <> 1 THEN
    RAISE EXCEPTION '600: claims-auditor missing or still gated — apply 597 before 600';
  END IF;

  PERFORM 1 FROM scheduled_tasks
    WHERE name='claims-audit-rotation' AND pre_query LIKE '%locked_at IS NULL%';
  IF NOT FOUND THEN
    RAISE EXCEPTION '600: pre_query does not exclude locked sites — it would audit against an owner halt';
  END IF;

  RAISE NOTICE '600 OK: claims-audit-rotation enabled at 3600s / 7-day window.';
END $$;

COMMIT;

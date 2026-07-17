-- 165_enable_dead_controls_check.sql — enable the dead_controls discovery
-- check on completeness-discovery-agent (Experience Loop guard rail 3,
-- RUNBOOK T2.3).
--
-- The check flags no-op interactive controls (href="#", "#!",
-- javascript:void...) on deployed page content: the control renders, the
-- visitor clicks, nothing happens. The vonc gauntlet's two dead CTAs are the
-- proof. Emits dead_control items, needs_human_review, NO handler
-- (detect-and-surface: wire / build / remove is a judgement call).
-- Runtime-fill shells are exempt; href="" stays phantom_internal_links'
-- class; handler-less <button> is Tier 4's claim.
--
-- ORDER: apply AFTER the chassis image carrying check_dead_controls.go is
-- live (image -> seed). On an older image the unknown name is logged and
-- skipped — harmless, but pointless.
--
-- Applied out of band (psql -f + ledger row same sitting per
-- bugs_open/aaa_fails_to_mend/007).

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent', '165_enable_dead_controls_check: pre-update');

DO $$
DECLARE
  checks jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO checks
  FROM agent_definitions
  WHERE type = 'completeness-discovery-agent' AND is_active;

  IF checks IS NULL THEN
    RAISE EXCEPTION '165: no active completeness-discovery-agent with run_checks.config.checks';
  END IF;
  IF checks ? 'dead_controls' THEN
    RAISE EXCEPTION '165: dead_controls already enabled';
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        checks || '["dead_controls"]'::jsonb)
  WHERE type = 'completeness-discovery-agent' AND is_active;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'discovery',
  '## dead_controls check enabled on completeness discovery
Observed: no-op hrefs (bare #, javascript:void) on deployed pages were invisible — the anchor scope is (correctly) outside phantom_internal_links'' jurisdiction, so the gauntlet''s two dead CTAs shipped and stayed.
Fix: dead_controls discovery check (guard rail 3, experience loop) appended to completeness-discovery-agent run_checks (migration 165, image-first ordering). Items: dead_control, needs_human_review, no handler.
Categories: fix, guard-rail',
  '["fix","guard-rail"]'::jsonb,
  'migration', '165_enable_dead_controls_check'
);

COMMIT;

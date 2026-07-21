-- 186_enable_truncated_component_check.sql — enable the truncated_component
-- discovery check on completeness-discovery-agent (bugs_open/046).
--
-- The check flags active content_components whose html_template is a cut-off
-- generation: an unterminated <script>/<style>/<section>/<div>/<fieldset>. Such a
-- template serves broken markup — an unterminated <script> stops the tool's
-- JavaScript and swallows the page tail as script text. 9 such components (8
-- tool, 1 section) across 6 domains were live and invisible to every check when
-- 046 was filed (rendered_html carried the same cut, so no template-vs-render
-- disagreement; Tier-4 acceptance passes a dead <script>).
--
-- Emits truncated_component items, needs_human_review, NO handler
-- (detect-and-surface: restore-from-intact-version / regenerate / remove is a
-- judgement call, and auto-regeneration is unsafe — tool recreation can fabricate
-- data, bugs_open/020). The spec carries intact_version_available /
-- intact_version_number so triage is immediate. A verifier re-checks the current
-- template at completion time.
--
-- ORDER: apply AFTER the chassis image carrying check_truncated_component.go is
-- live (image -> seed). On an older image the unknown name is logged and skipped
-- — harmless, but pointless. Verify the check compiled into the pod first:
--   kubectl exec -n ai-persona-system <chassis-pod> -- \
--     sh -c 'strings /app/agent-chassis | grep -c truncated_component'
--
-- Apply out of band (psql -f + a migration ledger row the same sitting, per
-- bugs_open/007's --record-only flow).

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent', '186_enable_truncated_component_check: pre-update');

DO $$
DECLARE
  checks jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO checks
  FROM agent_definitions
  WHERE type = 'completeness-discovery-agent' AND is_active;

  IF checks IS NULL THEN
    RAISE EXCEPTION '186: no active completeness-discovery-agent with run_checks.config.checks';
  END IF;
  IF checks ? 'truncated_component' THEN
    RAISE EXCEPTION '186: truncated_component already enabled';
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        checks || '["truncated_component"]'::jsonb)
  WHERE type = 'completeness-discovery-agent' AND is_active;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'discovery',
  '## truncated_component check enabled on completeness discovery
Observed (bugs_open/046): 9 active components (8 tool, 1 section) across 6 domains served unterminated <script> markup to live visitors and were invisible to every check — rendered_html carried the same cut as the template, and Tier-4 acceptance passes a dead <script>.
Fix: truncated_component discovery check (5-pair tag-imbalance predicate, calibrated to catch exactly the census fleet-wide, 0 over-fire) appended to completeness-discovery-agent run_checks (migration 186, image-first ordering). Items: truncated_component, needs_human_review, no handler (restore / regenerate / remove is a judgement call; auto-regeneration is fabrication-risky per bugs_open/020). Verifier re-checks the current template.
Categories: fix, guard-rail',
  '["fix","guard-rail"]'::jsonb,
  'migration', '186_enable_truncated_component_check'
);

COMMIT;

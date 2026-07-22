-- 188_enable_backend_entry_orphaned_check.sql — enable the backend_entry_orphaned
-- discovery check (Finding A: method_mismatch_link) on completeness-discovery-agent.
-- bugs_open/017 (static-cutover orphans backend entry forms).
--
-- The check flags a browser-clickable <a href="/route"> that lands on a backend
-- handler which does NOT accept GET — a GET link to a POST-only endpoint. The
-- visitor clicks and gets "405 Method Not Allowed" ("POST only"), yet every static
-- check passes because the href DOES resolve. This shipped on idea.uk the moment
-- the VM cutover gave "/" to the static site and the tool's landing-page forms were
-- lost, leaving two href="/audience-check" GET links to a POST-only handler. It is
-- an alert (needs_human_review, NO handler): repoint the link at a GET page, or
-- author the entry form — a business decision, not chassis-fixable.
--
-- Un-gated by deploy_config.target='vm' by design: idea.uk's deploy_config is empty
-- {} despite being VM-hosted, so a target gate would NOOP on the very site the bug
-- is about. A live GET probe (bounded: dedupe-per-path, maxProbePaths cap, 6s
-- timeout) sidesteps the unmodelled backend; a probe that cannot RUN is counted and
-- Warn-logged as UNCHECKED, never read as clean.
--
-- ORDER: apply AFTER the chassis image carrying check_backend_entry_orphaned.go is
-- live in the discovery pod (image -> seed). On an older image the unknown name is
-- logged and skipped — harmless, but pointless. Verify first:
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c BackendEntryOrphanedCheck'   # want >=1
-- Then apply out of band (psql -f + ledger row same sitting per bugs_open/007).
-- Council-reviewed (advisory) under corr ed4851c9; two rounds REVISE, detection
-- logic approved by every seat, verdict isolated to the enablement/scope points
-- this seed and the owner scope decision resolve.

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent', '188_enable_backend_entry_orphaned_check: pre-update');

DO $$
DECLARE
  checks jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,run_checks,config,checks}' INTO checks
  FROM agent_definitions
  WHERE type = 'completeness-discovery-agent' AND is_active;

  IF checks IS NULL THEN
    RAISE EXCEPTION '188: no active completeness-discovery-agent with run_checks.config.checks';
  END IF;
  IF checks ? 'backend_entry_orphaned' THEN
    RAISE EXCEPTION '188: backend_entry_orphaned already enabled';
  END IF;

  UPDATE agent_definitions
  SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        checks || '["backend_entry_orphaned"]'::jsonb)
  WHERE type = 'completeness-discovery-agent' AND is_active;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'discovery',
  '## backend_entry_orphaned check enabled on completeness discovery
Observed (bugs_open/017): after the idea.uk VM cutover, GET links to POST-only tool handlers (href="/audience-check") returned "405 Method Not Allowed" — the paid funnel was unreachable, yet every static check passed because the href resolves. No discovery check modelled the backend method.
Fix: backend_entry_orphaned discovery check (Finding A, method_mismatch_link) appended to completeness-discovery-agent run_checks (seed 188, image-first ordering). Live GET probe, un-gated (idea.uk deploy_config is empty {}), flags exactly 405. Items: backend_entry_orphaned, needs_human_review, no handler. Finding B (no_backend_entry) and a pre-cutover content-diff guard are named follow-ons.
Categories: fix, guard-rail',
  '["fix","guard-rail"]'::jsonb,
  'migration', '188_enable_backend_entry_orphaned_check'
);

COMMIT;

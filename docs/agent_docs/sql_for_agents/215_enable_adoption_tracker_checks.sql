-- 215: enable the four Phase E discovery checks — adoption tracker (companies
-- deploying AI agents) and protocol tracker (agent communication protocols)
--
-- Context: model_directory_pipeline Phase E. check_directory.go registers one
-- check pair per register kind from directoryCheckProfiles; the model pair was
-- enabled by 194, this enables the company/protocol pairs:
--   missing_adoption_tracker_section  / missing_adoption_tracker_page
--   missing_protocol_tracker_section  / missing_protocol_tracker_page
-- Both gate on the site's opt-in flag
-- (site_specs.classification.content_features.adoption_tracker /
-- .protocol_tracker) AND on the register holding current found claims OF THAT
-- KIND (the kind-scoped gate the council review specifically endorsed —
-- opting in while only model claims exist must not build an empty section).
--
-- Image precondition SATISFIED at apply time: chassis v1.0.1165 pod-verified
-- 2026-07-26 —
--   strings /app/agent-chassis | grep -c missing_adoption_tracker_section  → 1
--   (positive control missing_model_directory_section → 1; negative → 0)
-- An unregistered check NAME would be skipped, not an error
-- (discovery_checks.go:122-127), but there is nothing to wait for.
--
-- Same statement shape as 150_/188_/189_/190_/194_.
--
-- Verify after applying:
--   SELECT jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
--   FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- expect 25 → 29
--
-- NOTE for the dispatching session: completeness-discovery is per-site on
-- demand (owner ruling 2026-07-24). Findings land status='detected'
-- (unclaimable). Triage the PAGE items only; leave the SECTION items at
-- 'detected' while bugs_open/073 blocks index rebuilds on aao — a triaged
-- section item would burn its three attempts against a deterministic failure.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '["missing_adoption_tracker_section", "missing_adoption_tracker_page", "missing_protocol_tracker_section", "missing_protocol_tracker_page"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND is_active
  AND deleted_at IS NULL
  AND COALESCE(is_snapshot, false) = false   -- never touch a snapshot row
  -- idempotent: skip if already enabled
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
          ? 'missing_adoption_tracker_section';

COMMIT;

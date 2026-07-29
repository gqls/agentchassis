-- 256_render_audit_agent.sql
-- `render-audit-agent` — the callable unit that wires the render audit in.
--
-- This is the last piece of a chain built over 2026-07-28:
--   chassis action `request_render_audit`   (v1.0.1196, pod-verified)
--     -> topic system.adapter.render-audit.requests
--       -> pod render-audit-adapter (own Deployment, own group, own logs)
--         -> Chromium renders every deployed page
--           -> ContrastFinding / BrokenImage / OverflowFinding back on the
--              caller's responses topic
--
-- Everything below that line was proven end to end on 2026-07-28 21:17 with a
-- hand-published message. This seed makes it reachable from an orchestration
-- rather than from a human with kcat, which was the entire point: the Python it
-- replaces found 101 AA failures on one site and 65 across the fleet, but only
-- ever when somebody remembered to run it. An unreadable chart reached a live
-- site and the OWNER found it in a screenshot.
--
-- ORDER MATTERS AND WAS OBSERVED: image first, then seed. A seed naming an
-- action the deployed binary does not register fails at runtime. `request_render_audit`
-- and `system.adapter.render-audit.requests` were both confirmed present in the
-- running chassis binary (v1.0.1196) before this file was applied.
--
-- WHAT THIS DELIBERATELY DOES NOT DO. It does not raise work items from the
-- findings. There is no `contrast_failure` item type yet and inventing one here
-- would be a second, undiscussed decision — `bugs_open/122` candidate 2 describes
-- the triage half and it needs its own design (severity thresholds, dedup across
-- runs, and what an `over_image` approximation should be allowed to raise, which
-- is the question that decides whether the whole thing becomes noise). The audit
-- result lands in the orchestration's collected_data where a human or a later
-- step can read it. That is a smaller claim than "wired into triage", and it is
-- the true one.

BEGIN;

INSERT INTO agent_definitions (
  type, display_name, description, category, agent_category, status,
  is_active, version, default_config, created_at, updated_at
)
VALUES (
  'render-audit-agent',
  'Render audit (what a visitor actually sees)',
  'Renders every deployed page of a site in headless Chromium and reports text contrast against the effective background, images that failed to load, and horizontal overflow. Catches the defect family check_palette_contrast states in its own header it cannot see: a component hard-coding an ink over a themed fill.',
  'quality',
  -- agent_category is CHECK-constrained to strategist|executor|analyst|
  -- integrator|coordinator|specialist. 'quality' is NOT in that set (42P10 on
  -- first try) — the allowed values are not guessable from the column name.
  -- 'analyst' is the honest fit: this measures and reports, it changes nothing.
  'analyst',
  'active',
  true,
  1,
  $cfg${
    "workflow": {
      "steps": {
        "site": {
          "action": "ensure_site_record",
          "config": {
            "error_step": "complete_error"
          },
          "next_step": "audit",
          "description": "Resolve the site row so the audit can enumerate its deployed pages",
          "output_field": "site_record"
        },
        "audit": {
          "action": "request_render_audit",
          "config": {
            "error_step": "complete_error",
            "site_id_field": "site_record.site_id",
            "domain_field": "site_record.domain",
            "max_pages": 25
          },
          "next_step": "complete",
          "description": "Render every deployed page on the DEDICATED render-audit pod and await the findings. max_pages caps the sweep; the action logs a warning and returns truncated=true when the cap bites, because a capped sweep reporting clean is a false green.",
          "output_field": "render_audit"
        },
        "complete": {
          "action": "complete_workflow",
          "config": {
            "multiple_output_fields": ["render_audit"],
            "success_message": "Render audit complete. Findings are in render_audit; an over_image reading is an approximation (unknown backdrop) and is reported, not asserted."
          },
          "description": "Audit complete"
        },
        "complete_error": {
          "action": "complete_workflow",
          "config": {
            "output_fields": ["render_audit"],
            "success_message": "Render audit did not complete. NOTE: this is NOT a clean result — a failed audit and a clean audit must never be read the same way."
          },
          "description": "No verdict: the audit errored. Nothing was measured."
        }
      },
      "initial_step": "site"
    }
  }$cfg$::jsonb,
  now(), now()
)
-- The only unique key is (type, version) — checked with pg_indexes rather than
-- assumed. There is no partial unique on `type` alone, so a filtered ON CONFLICT
-- raises 42P10 (it did, first try).
ON CONFLICT (type, version)
DO UPDATE SET
  default_config = EXCLUDED.default_config,
  description    = EXCLUDED.description,
  is_active      = true,
  updated_at     = now();

COMMIT;

-- VERIFY
SELECT type, is_active,
       default_config->'workflow'->'initial_step' AS initial_step,
       jsonb_object_keys(default_config->'workflow'->'steps') AS step
FROM agent_definitions
WHERE type = 'render-audit-agent' AND is_active;

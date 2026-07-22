-- 190: enable the contact_form_undeliverable discovery check
--
-- Context (bugs_open/006 §B, applied live 2026-07-22): fleet-wide, generated
-- contact forms rendered fine and delivered nothing — form_action came from
-- per-component content_data written by the content LLM ("#contact" on most
-- sites, "" on others), which POSTs to the current URL = 405/404 on a static
-- host, so the message is silently lost. 10 of 11 live sites.
--
-- The render-seam fix (sanitiseFormAction in component_library.go, commits
-- 22678a74b / c419b6f34 / efe634b37) converts a non-delivering action to a
-- mailto wherever the site has a real address in sites.email, and deliberately
-- REFUSES to fabricate one where it has none. This check is the fail-loud
-- backstop for that refused case (and the visibility signal until pages
-- re-render): a contact-form component whose DEPLOYED rendered_html still POSTs
-- to self / an empty / a /contact action. Routes to needs_human_review with NO
-- handler — supplying an address, pointing at a live backend, or dropping the
-- form is a business decision, so it is not auto-fixed (compare dead_controls).
--
-- Runs under completeness-discovery-agent alongside dead_controls. Requires the
-- chassis image that registers "contact_form_undeliverable" (live: pod-grep
-- count 5 on 2026-07-21). The runner skips-not-errors on an unregistered name
-- (discovery_checks.go:122-127), so enabling ahead of an image is safe.
--
-- Applied live 2026-07-22 with the identical statement below (idempotent).
-- Recorded here for audit/reproducibility, matching 150_/188_/189_.
--
-- Verify after applying:
--   SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
--   FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active;
--   -- expect the array to end with "contact_form_undeliverable"
--   SELECT status, count(*) FROM site_work_items
--   WHERE item_type='contact_form_undeliverable' GROUP BY status;
--   -- items appear per site on the next discovery cycle (pipeline-triggered,
--   -- not a timed scheduled_task); expect ~10 until pages re-render.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '"contact_form_undeliverable"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'completeness-discovery-agent'
  AND is_active
  AND deleted_at IS NULL
  AND COALESCE(is_snapshot, false) = false   -- never touch a snapshot row
  -- idempotent: skip if already enabled
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
          ? 'contact_form_undeliverable';

COMMIT;

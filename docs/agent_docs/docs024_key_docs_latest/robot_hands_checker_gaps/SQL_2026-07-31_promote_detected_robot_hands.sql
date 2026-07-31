-- Promote robot-hands.com's detected discovery findings to 'triaged' so the
-- dispatch loop can claim them. Owner-authorised 2026-07-31 ("everything
-- detected on robot-hands"), knowing it rewrites live copy.
--
-- Mirrors TriageDetectedItemsAction (triage_detect_items_action.go:108-120):
-- status → 'triaged', triaged_at → now(), pipeline → 'build', and the original
-- pipeline preserved at spec.original_pipeline for auditing.
--
-- ONE DELIBERATE DEVIATION from that action, stated rather than hidden. The
-- action promotes EVERY detected row for the site, including rows with an empty
-- handler_agent. On this site that is 4 rows (3 image_url_404, 1 capability_gap).
-- An item with no handler cannot be cleared by any handler run: it would be
-- claimed, fail twice, and be relabelled "[unresolved after 2 attempts]" — which
-- asserts a handler FAILED when none could ever have succeeded. That poisons
-- 'unresolved', the fleet-wide needs-investigation signal, and is precisely the
-- defect class filed as bugs_open/077 and guarded by discovery_checks/remit.go.
-- handler_agent = '' is the canonical spelling of "no handler" since migration
-- 217. So: promote only rows that have somewhere to go.
--
-- The 4 excluded rows stay 'detected' and remain visible to reporting. They are a
-- capability gap to be handled, not work to be dispatched.

\set sid '00ff3af5-dad8-4770-9f70-3edc267a3c92'

UPDATE site_work_items
   SET status     = 'triaged',
       triaged_at = now(),
       spec       = jsonb_set(COALESCE(spec, '{}'::jsonb),
                              '{original_pipeline}', to_jsonb(pipeline)),
       pipeline   = 'build'
 WHERE site_id = :'sid'
   AND status  = 'detected'
   AND COALESCE(handler_agent, '') <> '';

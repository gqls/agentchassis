-- Phase 1 tail — set the contact page's business email (owner: idea.uk@contactforsales.com).
--
-- The needs_section_data / needs_human_review item for contact-info wants field `email` from
-- source `site_specs.identity.email` (top-level). idea.uk's identity has the email only NESTED
-- under identity.contact.email, so the resolver found nothing and escalated. Add the top-level
-- key the resolver reads, then re-drive the contact build and close the item.

\set ON_ERROR_STOP on
\set SID '1244516d-014d-421c-88c6-090bb1e9552a'
\set PLAN 'ff03bdef-3bb2-40eb-93ff-efa70f46b6b8'
BEGIN;

-- 1. Set the top-level identity.email (the exact source path the resolver reads).
UPDATE site_specs
SET data = jsonb_set(data, '{email}', '"idea.uk@contactforsales.com"'::jsonb, true),
    updated_at = NOW()
WHERE site_id = :'SID' AND aspect = 'identity' AND is_current = true;

-- 2. Re-drive the contact page build so contact-info renders with the email (this is a page
--    BUILD from the current plan sections, NOT a re-plan — it will not touch the plan).
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key, pipeline, approval_mode)
SELECT :'SID', 'manual-repair', 'needs_page', 'high',
       'Rebuild contact with business email now set (idea.uk@contactforsales.com)',
       jsonb_build_object('reason','rebuild','plan_id', :'PLAN','page_name','contact','page_role','content'),
       52, 'page-build-handler', 'triaged', 'manual-repair',
       'needs_page:contact', 'build', 'auto'
WHERE NOT EXISTS (
  SELECT 1 FROM site_work_items
  WHERE site_id = :'SID' AND item_key = 'needs_page:contact'
    AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved'));

-- 3. Resolve the escalation — the data it was waiting for now exists.
UPDATE site_work_items
SET status = 'complete',
    error = 'resolved: identity.email set to idea.uk@contactforsales.com (2026-07-16)',
    updated_at = NOW()
WHERE site_id = :'SID' AND item_type = 'needs_section_data'
  AND status = 'needs_human_review';

COMMIT;

\echo '=== identity.email now set ==='
SELECT data->>'email' AS top_level_email FROM site_specs
WHERE site_id = :'SID' AND aspect = 'identity' AND is_current = true;
\echo '=== contact rebuild queued + escalation closed ==='
SELECT item_type, status FROM site_work_items
WHERE site_id = :'SID' AND item_type IN ('needs_page','needs_section_data')
  AND (item_key='needs_page:contact' OR item_type='needs_section_data')
ORDER BY item_type, status;

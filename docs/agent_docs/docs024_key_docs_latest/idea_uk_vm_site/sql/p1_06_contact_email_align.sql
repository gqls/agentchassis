-- Phase 1 tail (fix) — align ALL contact-email sources so validate_page_content passes.
--
-- p1_05 set site_specs.identity.email (what the section-data resolver renders into contact-info)
-- to idea.uk@contactforsales.com, but validate_page_content's loadSiteContactEmail
-- (validate_page_content.go:735) resolves the CANONICAL email by COALESCE priority:
--   1. sites.email                                  ← was still idea-uk@leopardess.uk
--   2. sites.content_data->>'contact_email'
--   3. sites.content_data->'reviewed_brief'->>'contact_email'
--   4. site_specs.identity.email                    ← idea.uk@contactforsales.com (p1_05)
-- So contact-info rendered the NEW email but the validator's canonical was the OLD one → the page
-- failed with invalid_email ("doesn't match site contact"). Owner's choice is the new address, so
-- set the highest-priority source (sites.email) to it; now render and canonical agree.
--
-- NOTE: this is the PUBLIC contact email on the static site only. The tool's own operator
-- notification address (OPERATOR_EMAIL in /etc/idea/idea.env, idea-uk@leopardess.uk) is a separate
-- surface and is NOT touched here.

\set ON_ERROR_STOP on
\set SID '1244516d-014d-421c-88c6-090bb1e9552a'
BEGIN;

-- 1. The canonical the validator reads first.
UPDATE sites SET email = 'idea.uk@contactforsales.com', updated_at = NOW() WHERE id = :'SID';

-- 2. Keep the nested identity.contact.email coherent with the top-level identity.email.
UPDATE site_specs
SET data = jsonb_set(data, '{contact,email}', '"idea.uk@contactforsales.com"'::jsonb, true),
    updated_at = NOW()
WHERE site_id = :'SID' AND aspect = 'identity' AND is_current = true;

-- 3. Retry the failed contact build (reset needs_human_review → triaged; attempt_count back to 0).
UPDATE site_work_items
SET status = 'triaged', attempt_count = 0, error = '', updated_at = NOW()
WHERE site_id = :'SID' AND item_key = 'needs_page:contact' AND source = 'manual-repair'
  AND status = 'needs_human_review';

COMMIT;

-- ── VERIFY: what will loadSiteContactEmail now return? (must equal what contact-info renders) ──
\echo '=== canonical email now (all four COALESCE sources) ==='
SELECT COALESCE(
    NULLIF(s.email,''),
    NULLIF(s.content_data->>'contact_email',''),
    NULLIF(s.content_data->'reviewed_brief'->>'contact_email',''),
    (SELECT NULLIF(ss.data->>'email','') FROM site_specs ss
     WHERE ss.site_id=s.id AND ss.aspect='identity' AND ss.is_current LIMIT 1),
    ''
) AS canonical_email
FROM sites s WHERE s.id = :'SID';

\echo '=== contact build re-queued ==='
SELECT item_key, status, attempt_count FROM site_work_items
WHERE site_id = :'SID' AND item_key='needs_page:contact' AND source='manual-repair'
ORDER BY updated_at DESC LIMIT 1;

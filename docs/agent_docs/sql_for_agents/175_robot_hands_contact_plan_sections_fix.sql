-- 175: align robot-hands `contact` in the AUTHORITATIVE site_plan_sections
--      table with the component that is actually built, deployed and live
--      (bugs_open/002 error C)
--
-- Context (2026-07-19). The three stores load_page_sections_from_spec reads
-- (load_page_sections_from_spec_action.go:108-200) disagreed for this page:
--   1. site_plan_sections table   — AUTHORITATIVE — [hero-contact, contact-form, contact-info]
--   2. site_specs.site_plan aspect                 [hero-contact, contact-form, contact-block]
--   3. pages.sections (cache)                      [hero-contact, contact-form, contact-block]
-- and page_components + the live page also hold contact-block.
--
-- Source 1 wins and is synced DOWN over pages.sections, so the next rebuild of
-- `contact` would have replaced the deployed contact-block with contact-info.
-- Verified as a real swap, not a benign alias: resolution Pass 1 is an exact
-- match on content_components.name (v3_site_actions.go:3383-3398), so a plan
-- naming `contact-info` binds the component named `contact-info`; contact-block
-- is never a candidate. This is the same trap that resurrected product-detail's
-- deleted components on 2026-07-15 and was fixed by migration 154.
--
-- WHICH SIDE IS INTENDED: contact-block. Evidence —
--   * it is what is deployed and rendering (https://robot-hands.com/contact.html
--     returns 200 with .contact-block-section / .cb-info);
--   * it is purpose-built for this page (content_components.description =
--     'Component needed for section type "content-block-contact" on page
--     "contact"') and bespoke to this one site (1 site uses it fleet-wide);
--   * it is far richer — 12066-char template / 28 schema fields, against
--     contact-info's 2573 / 6 — and actively maintained (updated 2026-07-14,
--     against contact-info's 2026-03-09);
--   * sources 2 and 3 and page_components all already say contact-block. Only
--     the table, rewritten by the 2026-07-08 replan, says otherwise — the
--     `replan clobbers built pages` class (bugs_open/001).
-- Swapping to contact-info would also render an INCOMPLETE section: contact-info
-- requires a business contact email the site does not supply (that is exactly
-- what the open needs_section_data items on other sites are about).
--
-- SCOPE: source 1 only. Unlike 154, the resurrection has NOT happened yet, so
-- page_components and pages.sections are already correct and are left alone.
-- This migration is preventative.
--
-- Verify after applying (expect: 0 hero-contact | 1 contact-form | 2 contact-block):
--   SELECT sps.ordering, sps.component_name
--   FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id
--   WHERE sp.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND sp.is_current
--     AND sps.page_name='contact' ORDER BY sps.ordering;

BEGIN;

-- Rename the section in place: ordering and the rest of the layout are already
-- correct, so this is a single-field correction, not a delete/insert.
WITH cur AS (
    SELECT id AS plan_id FROM site_plans
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true
)
UPDATE site_plan_sections
SET component_name = 'contact-block'
WHERE plan_id IN (SELECT plan_id FROM cur)
  AND page_name = 'contact'
  AND ordering = 2
  AND component_name = 'contact-info';

COMMIT;

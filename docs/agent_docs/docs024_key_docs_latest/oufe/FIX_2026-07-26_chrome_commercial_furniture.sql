-- ============================================================================
-- oufe.com — remove the commercial furniture from site chrome
--
-- WHY
--   The header carried <a href="/contact.html" class="header-cta">Get Started</a>
--   and the footer a "Our Services" group, on a site that sells nothing and whose
--   roadmap brief said explicitly that nothing may offer or imply a purchase.
--   Neither is a fabrication, but both are wrong for this site: they are default
--   commercial shapes the chrome template supplies, and they set an expectation
--   the site does not meet. ("Get Started" on a research publication invites the
--   reader to begin a transaction that does not exist.)
--
--   Also fixes the dead /cases/index.html targets: that page is being created in
--   the same sitting, so these links resolve rather than 404.
--
-- WHY BOTH TABLES
--   Chrome is rendered once into site_components AND copied into each page's
--   pages.rendered_header / rendered_footer. Patching only site_components
--   changes nothing a visitor sees until every page is re-rendered; patching only
--   the pages leaves the master to be re-copied later. Both, in one transaction.
--
--   Deliberately NOT re-rendering chrome from the template afterwards: a
--   refresh_site_components pass regenerates it and would discard these edits.
--   Assemble-only re-render is the correct follow-up.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- 1. the chrome masters -------------------------------------------------------
UPDATE site_components SET
  rendered_html = replace(
    rendered_html,
    '<a href="/contact.html" class="header-cta">Get Started</a>',
    '<a href="/cases/index.html" class="header-cta">Read the cases</a>'
  ),
  updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND slot_name = 'header';

UPDATE site_components SET
  rendered_html = replace(rendered_html, '<h4>Our Services</h4>', '<h4>Explore</h4>'),
  updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND slot_name = 'footer';

-- 2. the per-page copies ------------------------------------------------------
UPDATE pages SET
  rendered_header = replace(
    rendered_header,
    '<a href="/contact.html" class="header-cta">Get Started</a>',
    '<a href="/cases/index.html" class="header-cta">Read the cases</a>'
  ),
  rendered_footer = replace(rendered_footer, '<h4>Our Services</h4>', '<h4>Explore</h4>'),
  updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND (rendered_header IS NOT NULL OR rendered_footer IS NOT NULL);

COMMIT;

-- Verify: expect 0 rows carrying either string anywhere.
--   SELECT slot_name FROM site_components
--    WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com')
--      AND (rendered_html LIKE '%Get Started%' OR rendered_html LIKE '%Our Services%');
--   SELECT name FROM pages
--    WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com')
--      AND (COALESCE(rendered_header,'') LIKE '%Get Started%'
--        OR COALESCE(rendered_footer,'') LIKE '%Our Services%');

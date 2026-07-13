-- Layer 1b — header/footer phantom link fixes (B2/B3).
--
-- Ships with the render_site_components_action.go change (cta_url resolved from
-- the real contact page; legal_links built from GetNavItems(NavGroupLegal)).
-- After applying, force a re-render of site_components (header, footer) so the
-- corrected templates + data regenerate the rendered HTML, then re-run the
-- phantom-link dry-run to confirm the site_component rows are gone.
--
-- These are SHARED components: the edits benefit every site using them.
--   - footer-4-column: legal links become data-driven, so a site renders only
--     legal links to pages that actually exist (none -> none). A site that truly
--     has privacy/terms pages should have them classified into the legal nav
--     group so they appear; if it relied on the old hardcoded /privacy.html and
--     /terms.html literals without real pages, those were phantoms anyway.
--   - header-bold-gradient: the CTA is gated on cta_url, so a site with no
--     contact page renders no CTA button instead of href="" — previously it
--     always rendered (defaulting text to "Get Started").

-- ---------------------------------------------------------------------------
-- 0. Snapshot
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS content_components_bak_navfix_0610 AS
SELECT * FROM content_components WHERE name IN ('header-bold-gradient', 'footer-4-column');

-- ---------------------------------------------------------------------------
-- 1. header-bold-gradient — gate the CTA on cta_url being present.
--    (Single-line anchor; stable to match.)
-- ---------------------------------------------------------------------------
UPDATE content_components
SET html_template = replace(html_template,
      '<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>',
      '{{if .cta_url}}<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>{{end}}'),
    updated_at = now()
WHERE name = 'header-bold-gradient';

-- ---------------------------------------------------------------------------
-- 2. footer-4-column — replace the two hardcoded legal anchors with a
--    data-driven range over the real legal_links (fed by NavGroupLegal).
--    Two exact-string replaces (no ambiguous whitespace inside the anchors):
--    the first becomes the range, the second is removed.
-- ---------------------------------------------------------------------------
UPDATE content_components
SET html_template = replace(html_template,
      '<a href="/privacy.html">Privacy Policy</a>',
      '{{range .legal_links}}<a href="{{.url}}">{{.name}}</a>{{end}}'),
    updated_at = now()
WHERE name = 'footer-4-column';

UPDATE content_components
SET html_template = replace(html_template,
      '<a href="/terms.html">Terms of Service</a>',
      ''),
    updated_at = now()
WHERE name = 'footer-4-column';

-- ---------------------------------------------------------------------------
-- 3. Verify. Each UPDATE above should report 1 row. If a *_literal flag is
--    still true, the stored text differed from the match and that replace
--    silently no-op'd — adjust the match string, don't assume success.
-- ---------------------------------------------------------------------------
SELECT name,
       (html_template LIKE '%{{if .cta_url}}<a href="{{.cta_url}}"%') AS header_cta_gated,
       (html_template LIKE '%/privacy.html%')                          AS footer_has_privacy_literal,
       (html_template LIKE '%/terms.html%')                            AS footer_has_terms_literal,
       (html_template LIKE '%{{range .legal_links}}%')                 AS footer_legal_data_driven
FROM content_components
WHERE name IN ('header-bold-gradient', 'footer-4-column')
ORDER BY name;

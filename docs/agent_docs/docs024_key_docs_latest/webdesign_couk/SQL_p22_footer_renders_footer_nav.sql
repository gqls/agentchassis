-- SQL_p22_footer_renders_footer_nav.sql — webdesign.co.uk
--
-- OWNER, 2026-07-29: "put about in the bottom nav and not the top."
--
-- The page flags were set (`about`: in_header=false, in_footer=true) and the
-- nav rebuild did exactly the right thing — `site_nav_items` now reads:
--     primary: Tools, Learn, News
--     utility: Home, About
-- and the header lost About as asked. **But About did not appear in the
-- footer**, because this site's footer template renders `{{range .categories}}`
-- and `categories` is built from the PRIMARY group only
-- (`render_site_components_action.go:127` — `categories` is derived from
-- `navItems`, which is `GetNavItems(..., []string{NavGroupPrimary}, ...)`).
--
-- The render context already carries what the footer wants, computed three
-- lines above and unused by this template:
--     footer_nav_items / quick_links  = primary + utility + legal (line 104)
--     quick_links_html                = primary + utility, pre-rendered
-- Same field names (`name`, `slug`, `url`, `label`), so the loop body is
-- unchanged — only the collection it ranges over.
--
-- SCOPE: the component (`043095a1-…`, "webdesign.co.uk Site Footer") is used by
-- ZERO other sites — verified before editing, because a shared component would
-- have made this a fleet change:
--     SELECT count(*) FROM site_components
--      WHERE component_id='043095a1-57de-47c7-8931-b0cb6d65191e'
--        AND site_id <> (SELECT id FROM sites WHERE domain='webdesign.co.uk');
--     -- 0
--
-- The platform-level question — should `categories` mean "primary" in a FOOTER
-- render at all? — is left alone deliberately. It is a shared-mechanism change
-- and belongs in its own review, not inside a site fix.

\set ON_ERROR_STOP on

BEGIN;

UPDATE content_components
   SET html_template = replace(html_template,
         '{{range .categories}}{{if .url}}<li><a href="{{.url}}">{{.name}}</a></li>{{end}}{{end}}',
         '{{range .quick_links}}{{if .url}}<li><a href="{{.url}}">{{.name}}</a></li>{{end}}{{end}}'),
       updated_at = NOW()
 WHERE id = '043095a1-57de-47c7-8931-b0cb6d65191e';

DO $verify$
DECLARE v_tpl text; v_others int;
BEGIN
    SELECT html_template INTO v_tpl FROM content_components
     WHERE id = '043095a1-57de-47c7-8931-b0cb6d65191e';

    IF position('{{range .quick_links}}' in v_tpl) = 0 THEN
        RAISE EXCEPTION 'footer template not switched to quick_links';
    END IF;
    IF position('{{range .categories}}' in v_tpl) > 0 THEN
        RAISE EXCEPTION 'footer template still ranges over categories';
    END IF;

    SELECT count(*) INTO v_others FROM site_components
     WHERE component_id = '043095a1-57de-47c7-8931-b0cb6d65191e'
       AND site_id <> (SELECT id FROM sites WHERE domain = 'webdesign.co.uk');
    IF v_others <> 0 THEN
        RAISE EXCEPTION 'component is shared with % other site(s) — revert', v_others;
    END IF;

    RAISE NOTICE 'footer now renders primary+utility nav; About will appear on the next chrome render';
END $verify$;

COMMIT;

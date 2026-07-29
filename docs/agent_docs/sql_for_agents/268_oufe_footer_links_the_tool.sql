-- 268_oufe_footer_links_the_tool.sql
-- The recovery-waterfall tool is LIVE, Tier-4 verified 13/13, and UNREACHABLE:
-- measured 2026-07-29, no page on oufe.com links to
-- /tools/tool-recovery-waterfall.html, and neither does the chrome. A visitor
-- can only reach it by knowing the URL.
--
-- WHY THE PLATFORM DID NOT CATCH IT. `check_orphan_pages` exists and would flag
-- this page (page_type='tool' is excluded ONLY when it has no nav flag; this one
-- has in_footer=true, so the exclusion does not apply). It runs in exactly one
-- agent, `completeness-discovery-agent`, which has NEVER run on oufe and has not
-- raised an item on any site since 2026-07-17. Fleet measurement and the rest of
-- the evidence: bugs_open/146.
--
-- WHY THIS IS A HAND PATCH AND NOT A CHROME RE-RENDER — READ BEFORE "FIXING"
-- THIS PROPERLY. `buildServicesHTML` (render_site_components_action.go:950)
-- already selects pages with in_header OR in_footer, so a regenerated footer
-- WOULD list the tool without any help. It is still the wrong move here:
--
--   The `<div class="footer-note">` disclosure — "OUFE publishes educational
--   analysis ... some of what is here is assembled with AI assistance that can
--   invent convincing detail. Check anything that matters against the primary
--   source." — is NOT in `footer-theme-chrome`'s template and is built by no Go
--   code. It exists ONLY in this site's stored `site_components.rendered_html`,
--   put there by hand by this workstream. **A `refresh_site_components` run
--   deletes the site's standing fallibility disclosure from every page**, and
--   reports success. Verified 2026-07-29 by grepping the live template for
--   `footer-note` (absent) and for the note's own words (absent).
--
-- So: patch the artefact, preserving the note, then reassemble pages with the
-- STORED chrome (rerender-pages WITHOUT refresh_site_components). The durable
-- fix — making the note survive a regeneration — is a shared-template change
-- and belongs to a council round, not to this file. Filed in bugs_open/146.
--
-- The anchor below is the Explore list. NOTE the Cases <li> appears TWICE in the
-- footer (Quick Links and Explore), so the anchor deliberately includes the
-- surrounding <h4>Explore</h4>/<ul> to hit the right one. The VERIFY block
-- counts occurrences rather than trusting the replace: a non-matching
-- replace() is a silent no-op that returns success (recorded landmine).

BEGIN;

UPDATE site_components
SET rendered_html = replace(rendered_html,
      '<h4>Explore</h4>
            <ul>
                <li><a href="/cases/index.html">Cases</a></li>
            </ul>',
      '<h4>Explore</h4>
            <ul>
                <li><a href="/cases/index.html">Cases</a></li>
                <li><a href="/tools/tool-recovery-waterfall.html">Recovery waterfall tool</a></li>
            </ul>'),
    updated_at = now()
WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39'
  AND slot_name = 'footer'
  -- Replay guard: no-op if the link is already there.
  AND rendered_html NOT LIKE '%/tools/tool-recovery-waterfall.html%';

COMMIT;

-- VERIFY: exactly one tool link, and the honesty note still present.
SELECT
  (length(rendered_html) - length(replace(rendered_html, '/tools/tool-recovery-waterfall.html', ''))) / length('/tools/tool-recovery-waterfall.html') AS tool_links,
  rendered_html LIKE '%OUFE publishes educational analysis%' AS note_preserved,
  rendered_html LIKE '%invent convincing detail%' AS note_body_preserved
FROM site_components
WHERE site_id = 'a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39' AND slot_name = 'footer';

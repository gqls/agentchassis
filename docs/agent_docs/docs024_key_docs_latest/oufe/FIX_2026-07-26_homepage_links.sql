-- ============================================================================
-- oufe.com — stop the homepage advertising pages that do not exist
--
-- WHAT WAS WRONG
--   Every content link on the live homepage 404'd: /restructuring-plan,
--   /creditor-waterfall, /cases/thames-water, /tools, /framework, /cases. Six
--   cards offering six sections, none of which had been built. That is the
--   bugs_open/052 class — a page advertising the site it was planned to become
--   rather than the site that exists.
--
--   Note /cases/thames-water was never going to resolve even once that page is
--   built: its actual url is /blog/thames-water.html. The card was written from
--   the plan's intent, not from the page record.
--
-- HOW THE FIX WORKS — use the templates' own fail-closed behaviour
--   info-card-grid wraps the anchor in {{if .link_url}}; hero and call-to-action
--   use {{if and .cta_text .cta_url}}. So a card with no link_url renders as
--   plain text with no dead anchor, by design. Rather than inventing
--   destinations, the cards that have nowhere real to go lose their link and say
--   plainly that the page is being written.
--
--   The same behaviour explains the two unresolved_cta review items: the hero
--   and closing CTA carried labels with no *_url, so the buttons did not render
--   at all. They now point at pages that exist.
--
-- POSTURE (owner direction, PLAN §7)
--   Say what the site is, and say what is not written yet. A card that promises
--   a page is a claim like any other.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- 1. hero — give the two buttons real destinations ---------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object(
         'cta_text', 'Read the cases',
         'cta_url', '/cases/index.html',
         'secondary_cta', 'How this site works',
         'secondary_cta_url', '/about.html'
       ),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'index' AND pc.position = 1;

-- 2. card grid — links only where a page exists ------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object(
      'section_subtitle',
        'This is the ground the site covers. Most of it is being written now: where a page exists, the card links to it, and where it does not, it says so rather than pretending otherwise.',
      'cards', jsonb_build_array(
        jsonb_build_object(
          'icon', '⚖️',
          'title', 'How restructuring plans bind creditors',
          'body', 'A restructuring plan under Part 26A can bind a class that voted against it. The conditions are specific and they are worth understanding properly. Being written — it comes before the cases, because it is the part that transfers to all of them.'
        ),
        jsonb_build_object(
          'icon', '📋',
          'title', 'The creditor waterfall, ranked',
          'body', 'Value is distributed in strict order of priority, so there is a point in the stack where it runs out and everything below recovers nothing. Being written. The recovery simulator already lets you watch that happen.'
        ),
        jsonb_build_object(
          'icon', '🏗️',
          'title', 'Case analysis: Thames Water',
          'body', 'A restructuring of a regulated utility, and unusually well documented because the regulator and the court both publish. We are working through the primary documents. Nothing is published yet, because we do not have verified figures and will not estimate them.'
        ),
        jsonb_build_object(
          'icon', '🔢',
          'title', 'Scenario tools that show their logic',
          'body', 'Each tool states the assumption you are changing and shows the arithmetic, so you can judge it rather than trust it. Any tool here can give you a wrong answer, and says so. In build.'
        ),
        jsonb_build_object(
          'icon', '📚',
          'title', 'UK statutory framework, precisely stated',
          'body', 'Part 26A, Part 26, special administration, cross-class cramdown. Every term introduced in plain English and then used correctly. Being written, against the statute rather than from memory.'
        ),
        jsonb_build_object(
          'icon', '📂',
          'title', 'Cases, indexed by mechanism',
          'body', 'The index organises coverage by the restructuring tool applied rather than by headline, so you can find the analysis that matches the situation in front of you.',
          'link_url', '/cases/index.html',
          'link_label', 'Go to the index'
        )
      )
    ),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'index' AND pc.position = 3;

-- 3. closing call-to-action — retarget off the unbuilt dossier ----------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object(
         'primary_cta', 'Go to the cases index',
         'primary_cta_url', '/cases/index.html',
         'secondary_cta', 'How this site works',
         'secondary_cta_url', '/about.html'
       ),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = (SELECT id FROM sites WHERE domain='oufe.com')
  AND p.name = 'index' AND pc.position = 4;

COMMIT;

-- Verify no card still points at an unbuilt page:
--   SELECT jsonb_path_query_array(content_data, '$.cards[*].link_url')
--     FROM page_components pc JOIN pages p ON p.id=pc.page_id
--    WHERE p.site_id=(SELECT id FROM sites WHERE domain='oufe.com')
--      AND p.name='index' AND pc.position=3;
--   -- expect exactly ["/cases/index.html"]

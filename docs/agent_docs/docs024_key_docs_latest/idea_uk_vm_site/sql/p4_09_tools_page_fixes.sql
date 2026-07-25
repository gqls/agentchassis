-- p4_09_tools_page_fixes.sql — /tools.html: dead CTAs, missing diagram, fabricated stats,
-- and the paid tool absent from its own tool listing. Owner report 2026-07-25:
--   "The structured diagram is not showing on the tools.html page. The paid-for tool doesn't
--    show on the tool listing. The 'Try the tools' and the 'Browse All Tools' buttons go to
--    the contact form and not a tools listing."
--
-- DIAGNOSES (each verified against the live page + stored content_data + schemas, not assumed):
--
-- 1. "Try the tools" -> /contact.html. The tools.html hero has cta_text but NO cta_url in
--    content_data, so it fell to the render-context default /contact.html — the identical
--    LNK-007 shape p3_05 fixed on the home page; this page's sections were never link-resolved.
--    Its secondary_cta ("See how the tools work") has no URL either, so the hero's second button
--    is gated out entirely.
--
-- 2. "Browse All Tools" -> /contact.html. tool-list.cta_url is query.section_index_for:tool,
--    which resolves nil for idea.uk (no tool section-index) -> default /contact.html. Worse, on
--    /tools.html the button is a self-reference even when "fixed" — pointing "Browse All Tools"
--    anywhere from the tools page is a label/URL mismatch (bugs_open/023). So BOTH label and URL
--    change: -> "Get a verified idea report" / /report.html, matching the guides hub (p4_04).
--    tool-list.cta_label is source=static WITH fallback 'Browse All Tools' — the p4_04
--    unoverridable-fallback defect — so the fallback must be dropped from the shared schema
--    first. Verified no-op: all 6 live instances (4 sites) carry cta_label in content_data with
--    exactly the fallback's value.
--
-- 3. Brief-explanation "Get Started"/"Learn More" -> "#". Labels present, URLs absent; the
--    template p3_05 gated renders '#'. Same mapping the owner chose for the home page: both
--    -> /report.html.
--
-- 4. The structured diagram: the rendered section carries <img src="/assets/images/
--    illustration.jpg"> and that file does not exist (live 404; absent from the vm-sites repo).
--    The site HAS a purpose-built asset that IS live: /assets/images/illustration-tools.jpg
--    (assets row illustration_tools, HTTP 200). Set illustration_url in content_data to it.
--    CAVEAT stated up front: illustration_url is source=site_assets.illustration, and
--    resolved_data merges LAST — if the resolver returns a value at re-render it will beat
--    content_data. Verify the rendered src live; if /assets/images/illustration.jpg comes back,
--    the fix moves to the resolver/asset row, not content_data.
--
-- 5. FABRICATED STAT, not reported by the owner but sitting on the live page: the
--    brief-explanation stats read "8 Tools available free" (there are TWO free tools) and
--    "Data stays on your device — Always" (false: the audience check posts to our server, and
--    the paid report obviously runs on our side). This is the bugs_open/043 class — a required
--    stat field filled with an invented number — on a page whose whole pitch is honesty.
--    Corrected to true values; "2" will need bumping as tools ship, which is the acceptable
--    cost of it being true.
--
-- 6. The paid tool missing from the listing: tool-list items derive from
--    query.pages_where_type:tool, and /report.html is page_type='landing'. Flip it to 'tool' —
--    it IS the site's flagship tool, and every consumer checked tolerates it: nav is untouched
--    (idx_pages_nav keys in_header/nav_order; both unchanged), FetchablePageEligibility passes
--    (deployed), sections-carrying tool pages are legitimate fleet-wide (13/33), and the CTA
--    resolver treating it as interactive is what this site's funnel WANTS (all idea.uk CTAs are
--    explicit + locked anyway). Its meta_description is empty (the listing card would be blank)
--    — set it. Also give the audience-check pointer page a real nav_label + meta_description so
--    its card stops reading "Free Audience Check — idea.uk" twice with no description.
--
-- 7. tool-list section_intro claimed "each one is labelled clearly so you know which is which"
--    (browser-local vs server) — no such labels exist on the cards. cta_supporting_text claimed
--    "Each tool is free" — false once the paid report joins the list. Both reworded to be true.

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 0. Guard for the shared-schema edit (same bar as p4_04): every tool-list
--    instance must already carry cta_label, or dropping the fallback is not a no-op.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  WHERE cc.name = 'tool-list' AND NOT (pc.content_data ? 'cta_label');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % tool-list instance(s) lack cta_label in content_data — fallback drop would not be a no-op.', n;
  END IF;
END
$guard$;

-- Snapshot the schema we edit.
DROP TABLE IF EXISTS bak_toollist_schema_20260725;
CREATE TABLE bak_toollist_schema_20260725 AS
SELECT id, name, input_schema, now() AS snapshotted_at
FROM content_components WHERE name = 'tool-list' AND is_active;

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,cta_label}',
      (input_schema->'fields'->'cta_label') - 'fallback'),
    updated_at = now()
WHERE name = 'tool-list' AND is_active;

-- ---------------------------------------------------------------------------
-- 1. tools.html hero: real targets, honestly labelled.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'cta_text',          'Try the free patent check',
      'cta_url',           '/tools/patent-check/index.html',
      'secondary_cta',     'Get a verified idea report',
      'secondary_cta_url', '/report.html'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.id = '73e8e1bb-f895-4f41-b63f-aca4a69c63d8'
  AND pc.slot_name = 'hero';

-- ---------------------------------------------------------------------------
-- 2. brief-explanation: live URLs, the real illustration, true stats.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'cta_primary_url',   '/report.html',
      'cta_secondary_url', '/report.html',
      'illustration_url',  '/assets/images/illustration-tools.jpg',
      'stat_1_value',      '2',
      'stat_1_label',      'Free tools available today',
      'stat_3_value',      'None',
      'stat_3_label',      'Sign-up needed for the free tools'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.id = '73e8e1bb-f895-4f41-b63f-aca4a69c63d8'
  AND pc.slot_name = 'brief-explanation';

-- ---------------------------------------------------------------------------
-- 3. tool-list instance: funnel CTA + honest copy.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'cta_url',             '/report.html',
      'cta_label',           'Get a verified idea report',
      'section_intro',       'These tools help you pressure-test an idea before you commit serious time or money to it. The free checks take a few minutes and need no account.',
      'cta_supporting_text', 'The checks are free. The Verified Idea Report is the paid option — £29, researched and reviewed before it is sent.'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.id = '73e8e1bb-f895-4f41-b63f-aca4a69c63d8'
  AND pc.slot_name = 'tool-list';

-- ---------------------------------------------------------------------------
-- 4. The paid tool becomes a listed tool; the pointer page gets a real card.
-- ---------------------------------------------------------------------------
UPDATE pages
SET page_type = 'tool',
    meta_description = 'The paid option — £29. A researched, human-reviewed report: ideas worth pursuing for your business, checked against what already exists, each with a cheap first test of real demand.',
    updated_at = now()
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND url = '/report.html';

UPDATE pages
SET nav_label = 'Free Audience Check',
    meta_description = 'A free check on who your idea is really for. Uses AI on our servers; takes about two minutes; no sign-up needed.',
    updated_at = now()
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND url = '/tools.html#audience-check';

COMMIT;

-- Read-backs.
SELECT pc.slot_name,
       pc.content_data->>'cta_url' AS cta_url,
       pc.content_data->>'secondary_cta_url' AS sec_url,
       pc.content_data->>'cta_primary_url' AS brief_pri,
       pc.content_data->>'illustration_url' AS illu,
       pc.content_data->>'stat_1_value' AS stat1
FROM page_components pc
WHERE pc.page_id = '73e8e1bb-f895-4f41-b63f-aca4a69c63d8'
ORDER BY pc.position;

-- What the tool listing will resolve after re-render:
SELECT p.url, p.nav_label, p.nav_order, left(coalesce(p.meta_description,''),50) AS meta
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.page_type = 'tool'
  AND p.status IN ('active','deployed')
  AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')
ORDER BY COALESCE(p.nav_order,100), p.name;

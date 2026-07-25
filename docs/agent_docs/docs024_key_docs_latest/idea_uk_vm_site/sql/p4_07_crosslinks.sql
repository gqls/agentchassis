-- p4_07_crosslinks.sql — wire the three new pages to each other and to the tools hub.
--
-- RUN ONLY AFTER /guides/copyright/index.html and /tools/patent-check/index.html are verified LIVE
-- (the guard enforces it). Until the tool page is fetchable, /tools.html's `tool-list` cannot list
-- it — that listing sources items from `query.pages_where_type:tool` under
-- FetchablePageEligibilitySQL, exactly like the guides hub (p4_02).
--
-- WHAT THIS FIXES. Each page currently works but is a cul-de-sac:
--   * the patents guide's secondary CTA still points at the OLD free taster
--     (/tools.html#audience-check) — it predates the checker, which is now the right next step
--     from that guide;
--   * the copyright guide already links the checker (written that way in p4_05), but the patents
--     guide — the one the checker actually belongs to — does not;
--   * nothing links patents ↔ copyright, though they are the two halves of "how do I protect this".
--
-- The three pages are otherwise only reachable from the hubs, so a reader who lands on one from
-- search sees no route onward. That is the same dead-end shape the CTA work has been chasing all
-- session, just made of absence rather than a wrong href.
--
-- NOTE ON WHAT IS *NOT* CHANGED: /tools.html's own sections are untouched. It only needs a
-- re-render for its derived tool-list to pick up the new page — no content edit at all.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url IN ('/guides/copyright/index.html','/tools/patent-check/index.html')
     AND (deployed_at IS NOT NULL OR build_status = 'deployed');
  IF n <> 2 THEN
    RAISE EXCEPTION 'ABORT: expected both new pages fetchable, found %. Verify them live first (curl, not the job).', n;
  END IF;
END
$guard$;

-- ---------------------------------------------------------------------------
-- 1. Patents guide -> the checker (replaces the pre-checker taster link) and -> copyright.
-- ---------------------------------------------------------------------------
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'secondary_cta',     'Should you patent it? Free check',
      'secondary_cta_url', '/tools/patent-check/index.html'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html'
  AND pc.slot_name = 'call-to-action';

-- Its hero's secondary slot carried a generic "All guides"; point it at the sibling guide instead,
-- which is a more useful next step and still reachable from the hub in the nav.
UPDATE page_components pc
SET content_data = pc.content_data || jsonb_build_object(
      'secondary_cta',     'Copyright: what you already own',
      'secondary_cta_url', '/guides/copyright/index.html'
    ),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html'
  AND pc.slot_name = 'hero';

COMMIT;

SELECT p.url, pc.slot_name,
       pc.content_data->>'secondary_cta'     AS sec_label,
       pc.content_data->>'secondary_cta_url' AS sec_url
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url IN ('/guides/patents/index.html','/guides/copyright/index.html')
  AND pc.slot_name IN ('hero','call-to-action')
ORDER BY p.url, pc.position;

-- What /tools.html will resolve into its tool-list once re-rendered:
SELECT p.url, COALESCE(p.title,p.name) AS title, p.nav_label, p.nav_order
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.page_type = 'tool'
  AND p.status IN ('active','deployed')
  AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')
ORDER BY COALESCE(p.nav_order,100), p.name;

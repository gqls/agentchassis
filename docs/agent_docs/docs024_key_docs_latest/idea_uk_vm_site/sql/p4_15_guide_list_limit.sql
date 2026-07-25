-- p4_15_guide_list_limit.sql — the guides hub silently caps at SIX guides; idea.uk now has nine.
--
-- Observed live: after stages 1–5 shipped, the hub rerender COMPLETED but the page lists 6 cards.
-- Cause: guide-list_pre_037's `items` field declares `"limit": 6`; resolvePagesWhereType honours
-- it (hard cap 24). First six by nav_order win — creating-ideas..feedback-loops + patents — and
-- copyright + both funding guides fall off the very hub that exists to list them. A silent cap:
-- the render is green, the page looks healthy, three guides are just absent (the "no silent caps"
-- pattern from the workflow guidance, here as data).
--
-- Fix: limit 6 → 24 (the resolver's own hard cap). NO-OP VERIFIED for the other instances before
-- writing: gamesdesign has 5 guide pages, relojistas 4 — both under the old cap, so their
-- listings are byte-identical either way. The guard enforces that check rather than trusting it.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  -- No OTHER site using this component may have more than 6 fetchable guides,
  -- or this change alters their live listing and stops being a no-op.
  SELECT count(*) INTO n FROM (
    SELECT p.site_id FROM pages p
    WHERE p.page_type = 'guide'
      AND p.status IN ('active','deployed')
      AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')
      AND p.site_id <> '1244516d-014d-421c-88c6-090bb1e9552a'
      AND p.site_id IN (SELECT DISTINCT p2.site_id FROM page_components pc2
                        JOIN pages p2 ON p2.id = pc2.page_id
                        WHERE pc2.component_id = '9d5e461a-8981-4ecc-b236-05895edfc15d')
    GROUP BY p.site_id HAVING count(*) > 6
  ) x;
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % other site(s) exceed 6 guides — raising the limit would change their live listing; coordinate first.', n;
  END IF;
END
$guard$;

UPDATE content_components
SET input_schema = jsonb_set(input_schema, '{fields,items,limit}', '24'::jsonb),
    updated_at = now()
WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';

DO $guard2$
DECLARE l int;
BEGIN
  SELECT (input_schema->'fields'->'items'->>'limit')::int INTO l
  FROM content_components WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';
  IF l <> 24 THEN RAISE EXCEPTION 'ABORT: limit edit did not take (%).', l; END IF;
END
$guard2$;

COMMIT;

SELECT input_schema->'fields'->'items'->>'limit' AS items_limit
FROM content_components WHERE id = '9d5e461a-8981-4ecc-b236-05895edfc15d';

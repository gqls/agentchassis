-- Curated darts RSS sources. Inserted AFTER the first content-feed-refresh run
-- (13:52Z today) seeded the 5 news_search + 1 api_news sources — deliberately, because
-- seed_content_sources_action.go:92-111 skips seeding entirely if ANY active source
-- exists, so inserting these first would have meant the search sources were never
-- created. Seeding has now happened, so these are additive and safe.
--
-- Each feed was fetched, parsed and recency-checked today before being written here
-- (RUNBOOK §9). Rejected on evidence: live-darts.com 403, dartsnews.com 404,
-- dartsorakel.com 404, and skysports.com/rss/12040 — which returns a healthy-looking
-- 200 with 20 items and is the GENERIC Sky news feed, three occurrences of "darts" in
-- the whole document.
--
-- Names are distinct from the seeder's "News Search: <keyword>" pattern so the
-- (site_id, name) unique constraint cannot collide with a future re-seed.
-- fetch_interval follows the relojistas precedent: 4h for the publications that post
-- daily, 6h for the aggregator.
BEGIN;
INSERT INTO content_sources (site_id, source_type, name, config, fetch_interval)
VALUES
  ('5fe8785b-223d-41a3-88ee-c07187622381', 'rss', 'Darts World',
   '{"feed_url": "https://dartsworld.com/feed/", "max_items": 15}'::jsonb, '04:00:00'),
  ('5fe8785b-223d-41a3-88ee-c07187622381', 'rss', 'PDC (official)',
   '{"feed_url": "https://www.pdc.tv/rss.xml", "max_items": 15}'::jsonb, '04:00:00'),
  ('5fe8785b-223d-41a3-88ee-c07187622381', 'rss', 'Google News — darts',
   '{"feed_url": "https://news.google.com/rss/search?q=darts+PDC&hl=en-GB&gl=GB&ceid=GB:en", "max_items": 10}'::jsonb, '06:00:00')
ON CONFLICT (site_id, name) DO NOTHING;
COMMIT;
SELECT source_type, name, config->>'feed_url' AS url, fetch_interval, error_count
FROM content_sources WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' ORDER BY source_type, name;

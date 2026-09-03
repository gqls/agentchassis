-- 746_advertise_news_feed_enablement_ROLLBACK.sql
--
-- Reverses migration 746 for advertise.co.uk (d991a5b8-428f-44c1-b3eb-e50f44326fd9):
-- removes the six content_sources rows it created (and any content_feed_items they
-- ingested — derived data that exists only because of 746), deletes the authored
-- classification row, and reinstates the superseded classifier row ec005136 as current.
--
-- REFUSES if any ingested item has been published to a page or attached to a work
-- item: at that point the items are provenance for something live and an operator must
-- decide, not a rollback script. Run by hand, never by the runner (UPPERCASE sidecar).
--
-- Run:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--         psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--         -f - < docs/agent_docs/sql_for_agents/746_advertise_news_feed_enablement_ROLLBACK.sql

BEGIN;

DO $pre$
DECLARE
  n_auth      int;
  n_old       int;
  n_src       int;
  n_published int;
BEGIN
  SELECT count(*) INTO n_auth FROM site_specs
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9' AND aspect = 'classification'
     AND is_current = true AND created_by = 'news-feed-ingestion-lane';
  IF n_auth <> 1 THEN
    RAISE EXCEPTION '746 ROLLBACK: expected exactly 1 current classification row authored by 746, found % — nothing to roll back, or a later row superseded it; read the table', n_auth;
  END IF;

  SELECT count(*) INTO n_old FROM site_specs
   WHERE id = 'ec005136-e07d-4d5f-aad4-beef6ec02517' AND is_current = false;
  IF n_old <> 1 THEN
    RAISE EXCEPTION '746 ROLLBACK: the superseded classifier row ec005136 is missing or already current';
  END IF;

  SELECT count(*) INTO n_src FROM content_sources
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND name IN ('WebProNews',
                  'News Search: Advertising Standards Authority rulings',
                  'News Search: CAP Code advertising rules',
                  'News Search: IAB UK digital advertising spend',
                  'News Search: Advertising Association WARC expenditure report',
                  'News Search: UK advertising industry news');
  IF n_src <> 6 THEN
    RAISE EXCEPTION '746 ROLLBACK: expected the 6 sources 746 created, found % — a retune renamed or removed some; roll back by hand', n_src;
  END IF;

  SELECT count(*) INTO n_published FROM content_feed_items cfi
    JOIN content_sources cs ON cs.id = cfi.source_id
   WHERE cs.site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND (cfi.published_page_id IS NOT NULL OR cfi.work_item_id IS NOT NULL);
  IF n_published <> 0 THEN
    RAISE EXCEPTION '746 ROLLBACK REFUSED: % ingested items are attached to a page or work item — deleting them would orphan live provenance; an operator must decide', n_published;
  END IF;
END
$pre$;

-- 1. derived data first (FK content_feed_items.source_id → content_sources)
DO $items$
DECLARE n int;
BEGIN
  -- self-FK duplicate_of: clear it inside the set before deleting
  UPDATE content_feed_items SET duplicate_of = NULL
   WHERE duplicate_of IN (SELECT cfi.id FROM content_feed_items cfi
                           JOIN content_sources cs ON cs.id = cfi.source_id
                          WHERE cs.site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9');
  DELETE FROM content_feed_items
   WHERE source_id IN (SELECT id FROM content_sources WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9');
  GET DIAGNOSTICS n = ROW_COUNT;
  RAISE NOTICE '746 ROLLBACK: deleted % ingested content_feed_items', n;
END
$items$;

-- 2. the six sources
DO $src$
DECLARE n int;
BEGIN
  DELETE FROM content_sources
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9'
     AND name IN ('WebProNews',
                  'News Search: Advertising Standards Authority rulings',
                  'News Search: CAP Code advertising rules',
                  'News Search: IAB UK digital advertising spend',
                  'News Search: Advertising Association WARC expenditure report',
                  'News Search: UK advertising industry news');
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 6 THEN
    RAISE EXCEPTION '746 ROLLBACK: deleted % sources, expected 6', n;
  END IF;
END
$src$;

-- 3. the spec: delete the authored row, reinstate the classifier's
DO $spec$
DECLARE n int;
BEGIN
  DELETE FROM site_specs
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9' AND aspect = 'classification'
     AND is_current = true AND created_by = 'news-feed-ingestion-lane';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '746 ROLLBACK: deleted % authored spec rows, expected 1', n;
  END IF;

  UPDATE site_specs SET is_current = true, superseded_at = NULL
   WHERE id = 'ec005136-e07d-4d5f-aad4-beef6ec02517';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN
    RAISE EXCEPTION '746 ROLLBACK: reinstated % rows, expected 1', n;
  END IF;
END
$spec$;

-- POST-CHECK
DO $post$
DECLARE has_cf bool; n_src int;
BEGIN
  SELECT COALESCE(bool_or(data ? 'content_features'), false) INTO has_cf FROM site_specs
   WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9' AND aspect = 'classification' AND is_current = true;
  IF has_cf THEN
    RAISE EXCEPTION '746 ROLLBACK POST-CHECK FAILED: current classification still carries content_features';
  END IF;
  SELECT count(*) INTO n_src FROM content_sources WHERE site_id = 'd991a5b8-428f-44c1-b3eb-e50f44326fd9';
  IF n_src <> 0 THEN
    RAISE EXCEPTION '746 ROLLBACK POST-CHECK FAILED: % content_sources rows remain for advertise.co.uk', n_src;
  END IF;
  RAISE NOTICE '746 ROLLBACK PASSED: advertise.co.uk back to the 2026-09-02 classifier row and zero sources.';
END
$post$;

COMMIT;

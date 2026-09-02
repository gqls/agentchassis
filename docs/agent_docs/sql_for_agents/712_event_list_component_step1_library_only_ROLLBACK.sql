-- ROLLBACK for 712: removes the event-list component. Safe at any time —
-- this migration never attached it to a page.
BEGIN;
DELETE FROM content_components WHERE name = 'event-list'
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.component_id = content_components.id);
COMMIT;

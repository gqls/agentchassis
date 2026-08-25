-- news_editorial_features lane — 2026-08-25
-- RE-LOCK the two evidence-timeseries instances after the instance-scope
-- conversion has been delivered and VERIFIED at the served artefact.
--
-- Run this as soon as verification passes. An unlocked flagship row is exposed
-- to every sweep on the estate, and this lane has already been hit once by an
-- improvement-loop misfire (2026-08-22).
--
-- Restores exactly what was there before: lock_type='permanent',
-- locked_by='news_editorial_features-lane'.
-- Idempotent; guarded so it can only re-lock rows that are currently unlocked.

BEGIN;

UPDATE page_components
   SET lock_type = 'permanent', locked_by = 'news_editorial_features-lane'
 WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
              'ea6b4ca7-7717-4e29-ae1c-88844040b0d2')
   AND lock_type IS DISTINCT FROM 'permanent'
RETURNING id, slot_name, lock_type, locked_by;

COMMIT;

-- Raw read-back: both rows must show permanent / news_editorial_features-lane.
SELECT id, slot_name, lock_type, locked_by, component_version_id IS NOT NULL AS stamped,
       length(rendered_html) AS html_bytes, updated_at
  FROM page_components
 WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
              'ea6b4ca7-7717-4e29-ae1c-88844040b0d2');

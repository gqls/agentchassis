-- ROLLBACK for 777. Returns /blog's sections to the empty array it held before.
-- ⚠ This restores the SILENT SKIP: the page goes back to being unable to
-- escalate. Only run it if declaring the sections caused a worse outcome than
-- the hole it closed.
BEGIN;
UPDATE pages
SET sections = '[]'::jsonb
WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND name = 'blog' AND status = 'active'
  AND sections::text = '["hero", "blog-listing", "call-to-action"]';
COMMIT;

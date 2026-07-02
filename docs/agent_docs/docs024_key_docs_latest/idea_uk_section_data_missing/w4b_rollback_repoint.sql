-- W4b rollback: restore the old component ids (guarded on the new ones).
-- rendered_html was never touched by the repoint; the pre-repoint dump (.bak) holds it anyway.
UPDATE site_components
SET component_id = '9644c86f-18b0-4f75-b086-5b79a74a48d7', updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND slot_name = 'header'
  AND component_id = 'f420f3fa-43a2-4a2f-b2e1-39770d45b494'
RETURNING slot_name;

UPDATE site_components
SET component_id = '09034086-a581-4bba-a5b4-760d863bb2df', updated_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND slot_name = 'footer'
  AND component_id = '4238e467-25a6-4174-bee0-6fce914398c8'
RETURNING slot_name;

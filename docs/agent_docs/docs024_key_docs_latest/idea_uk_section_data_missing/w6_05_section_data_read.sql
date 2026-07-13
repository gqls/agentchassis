-- W6 step 5 (read-only): what exactly does brief-explanation's illustration_url need?
-- 5.1 The two escalation items in full (spec + the advice columns the schema provides):
SELECT item_key, status, jsonb_pretty(spec) AS spec,
       suggested_action, resolution_path, affected_url
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND item_key LIKE 'section_data_%brief-explanation%'
ORDER BY created_at;

-- 5.2 The component's input_schema — the source declaration for illustration_url
--     (and whether the field is marked required):
SELECT function, jsonb_pretty(input_schema) AS input_schema
FROM content_components
WHERE function = 'brief-explanation' AND is_active = true AND forked_from IS NULL;

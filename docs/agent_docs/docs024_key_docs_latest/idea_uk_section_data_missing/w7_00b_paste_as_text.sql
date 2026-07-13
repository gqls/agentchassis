-- W7 step 0b (read-only): the still-needed reads. PASTE THE OUTPUT AS TEXT — attached
-- documents have arrived unreadable twice; pasted text works.

-- 0.1 brief-explanation: input_schema (the exact field entry for illustration_url —
--     its source path and how optionality is expressed) + the illustration markup:
SELECT function,
       jsonb_pretty(input_schema) AS input_schema,
       substr(html_template, greatest(position('illustration' in html_template) - 120, 1), 480) AS illustration_region
FROM content_components
WHERE function = 'brief-explanation' AND is_active = true AND forked_from IS NULL;

-- 0.2 The two escalations' advice columns:
SELECT item_key, jsonb_pretty(spec) AS spec, suggested_action, resolution_path
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND item_key LIKE 'section_data_%brief-explanation%'
ORDER BY created_at;

-- 0.3 idea.uk's existing plan-imagery rows (schema already read — plan_id confirmed):
SELECT spi.scope, spi.scope_ref, spi.key, spi.kind, left(spi.prompt, 100) AS prompt_head,
       spi.ordering, spi.source
FROM site_plan_imagery spi
JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
ORDER BY spi.scope, spi.scope_ref, spi.ordering;

-- 0.4 needs_imagery shapes: idea.uk's history + three real items from any site:
SELECT item_key, status, left(summary, 80) AS summary, created_at
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk') AND item_type = 'needs_imagery'
ORDER BY created_at DESC LIMIT 10;

SELECT site_id, item_key, status, left(spec::text, 260) AS spec_head, created_at
FROM site_work_items
WHERE item_type = 'needs_imagery'
ORDER BY created_at DESC LIMIT 3;

-- 0.5 the section descriptions (prompt source for the imagery rows). Schema first:
\d site_plan_sections
SELECT sps.page_name, sps.section_name, left(sps::text, 300) AS row_head
FROM site_plan_sections sps
JOIN site_plans sp ON sp.id = sps.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND sps.section_name = 'brief-explanation';
-- (If column names differ per \d, adjust page_name/section_name accordingly.)

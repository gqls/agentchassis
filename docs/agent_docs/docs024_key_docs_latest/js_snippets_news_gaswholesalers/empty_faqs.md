-- What's in the faq page's components — is the FAQ content_data populated? gaswholesalers
SELECT
pc.id,
pc.position,
cc.function,
cc.name AS component_name,
pc.build_status,
-- Is there content_data, and does it have FAQ items?
jsonb_typeof(pc.content_data) AS content_data_type,
CASE
WHEN pc.content_data ? 'faqs' THEN jsonb_array_length(pc.content_data->'faqs')
WHEN pc.content_data ? 'items' THEN jsonb_array_length(pc.content_data->'items')
WHEN pc.content_data ? 'questions' THEN jsonb_array_length(pc.content_data->'questions')
ELSE NULL
END AS item_count,
LEFT(pc.content_data::text, 300) AS content_data_preview
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_components cc ON pc.component_id = cc.id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
AND p.name = 'faq'
ORDER BY pc.position;

-- And what does the FAQ component template expect? What variable does it bind?
SELECT
id, name, function,
LEFT(html_template, 1200) AS template_preview,
input_schema
FROM content_components
WHERE function = 'faq' OR function ILIKE '%faq%' OR name ILIKE '%faq%'
ORDER BY updated_at DESC
LIMIT 3;
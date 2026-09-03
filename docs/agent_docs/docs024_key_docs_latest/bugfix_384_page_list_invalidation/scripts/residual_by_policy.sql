\pset format aligned
\pset pager off
-- Standing blanks in LIVE listing arrays, card-joined (an entry counts only if its target
-- page has an ACTIVE card that already exists). Split by rebuild_policy, because an owned
-- page is structurally out of the seam's reach (save_sections refuses it).
WITH qf AS MATERIALIZED (
  SELECT DISTINCT cc.id AS component_id, f.key AS field
    FROM content_components cc
    CROSS JOIN LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) f
   WHERE f.value->>'source' LIKE 'query.%' AND f.value->>'type' = 'array'
),
live AS MATERIALIZED (
  SELECT pc.page_id, pc.slot_name, qf.field, pc.updated_at,
         CASE WHEN jsonb_typeof(pc.content_data->qf.field)='array' THEN pc.content_data->qf.field END AS arr
    FROM page_components pc JOIN qf ON qf.component_id = pc.component_id
   WHERE pc.build_status <> 'removed'
),
ent AS MATERIALIZED (
  SELECT l.page_id, l.slot_name, l.field, l.updated_at, p.site_id,
         coalesce(p.rebuild_policy,'generic') AS policy, p.name AS page, s.domain,
         e.value->>'url' AS entry_url, coalesce(e.value->>'image','') AS entry_image
    FROM live l
    JOIN pages p ON p.id = l.page_id
    JOIN sites s ON s.id = p.site_id
    CROSS JOIN LATERAL jsonb_array_elements(l.arr) e
   WHERE l.arr IS NOT NULL AND p.status='active'
)
SELECT ent.policy, ent.domain, ent.page, ent.slot_name,
       ent.updated_at::timestamp(0) AS array_written, ent.entry_url,
       round(extract(epoch from (now()-max(ca.created_at)))/3600.0,1) AS card_age_h
  FROM ent
  JOIN pages tp ON tp.site_id = ent.site_id AND tp.url = ent.entry_url
  JOIN assets ca ON ca.site_id=ent.site_id AND ca.entity_type='page'
                AND ca.entity_id=tp.id AND ca.purpose='card' AND ca.status='active'
 WHERE ent.entry_image = ''
 GROUP BY 1,2,3,4,5,6 ORDER BY 1,2,3;

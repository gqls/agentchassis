\pset format aligned
\pset pager off
-- A write is only "repairable" for entries whose TARGET PAGE ALREADY HAD AN ACTIVE CARD
-- when the write ran. Counting bare image:'' conflates "should have an image and hasn't"
-- with "correctly has none" (RUNBOOK: join the card, don't count empties).
-- Restricted to query.*-sourced array fields. MATERIALIZED throughout: AND order is not
-- guaranteed, and an unmaterialised typeof guard gets reordered behind array_length.
WITH qf AS MATERIALIZED (
  SELECT DISTINCT cc.id AS component_id, f.key AS field
    FROM content_components cc
    CROSS JOIN LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) f
   WHERE f.value->>'source' LIKE 'query.%' AND f.value->>'type' = 'array'
),
slots AS MATERIALIZED (            -- (page_id, slot_name) -> the query-sourced field(s)
  SELECT DISTINCT pc.page_id, pc.slot_name, qf.field
    FROM page_components pc JOIN qf ON qf.component_id = pc.component_id
   WHERE pc.build_status <> 'removed'
),
h AS MATERIALIZED (
  SELECT pch.page_id, pch.site_id, pch.slot_name, pch.created_at, pch.application_name,
         s.field, pch.content_data -> s.field AS arr_raw
    FROM page_component_history pch
    JOIN slots s ON s.page_id = pch.page_id AND s.slot_name = pch.slot_name
   WHERE pch.source = 'artefact_archive_trigger'
     AND pch.created_at > now() - interval '10 days'
),
ha AS MATERIALIZED (
  SELECT h.*, CASE WHEN jsonb_typeof(h.arr_raw) = 'array' THEN h.arr_raw END AS arr FROM h
),
seq AS MATERIALIZED (
  SELECT ha.page_id, ha.site_id, ha.slot_name, ha.created_at, ha.application_name, ha.field, ha.arr,
         LEAD(ha.arr) OVER (PARTITION BY ha.page_id, ha.slot_name, ha.field ORDER BY ha.created_at) AS next_arr,
         (SELECT pc.content_data -> ha.field FROM page_components pc
           WHERE pc.page_id = ha.page_id AND pc.slot_name = ha.slot_name
             AND pc.build_status <> 'removed' LIMIT 1) AS live_arr
    FROM ha WHERE ha.arr IS NOT NULL
),
g AS MATERIALIZED (
  SELECT seq.*, CASE WHEN jsonb_typeof(COALESCE(seq.next_arr, seq.live_arr)) = 'array'
                     THEN COALESCE(seq.next_arr, seq.live_arr) END AS produced
    FROM seq
),
-- carded(entry) := the entry's target page had an ACTIVE card created before this write
scored AS MATERIALIZED (
  SELECT g.*,
    (SELECT count(*) FROM jsonb_array_elements(g.arr) e
       JOIN pages tp ON tp.site_id = g.site_id AND tp.url = e->>'url'
      WHERE COALESCE(e->>'image','') = ''
        AND EXISTS (SELECT 1 FROM assets ca WHERE ca.site_id = g.site_id AND ca.entity_type='page'
                      AND ca.entity_id = tp.id AND ca.purpose='card' AND ca.status='active'
                      AND ca.created_at < g.created_at)) AS pre_deficit,
    (SELECT count(*) FROM jsonb_array_elements(COALESCE(g.produced,'[]'::jsonb)) e
       JOIN pages tp ON tp.site_id = g.site_id AND tp.url = e->>'url'
      WHERE COALESCE(e->>'image','') = ''
        AND EXISTS (SELECT 1 FROM assets ca WHERE ca.site_id = g.site_id AND ca.entity_type='page'
                      AND ca.entity_id = tp.id AND ca.purpose='card' AND ca.status='active'
                      AND ca.created_at < g.created_at)) AS post_deficit
    FROM g
)
SELECT s.domain, p.name AS page, sc.slot_name, sc.field, sc.created_at::timestamp(0) AS write_at,
       CASE WHEN sc.created_at < '2026-09-02 11:27:53+00' THEN 'pre ' ELSE 'POST' END AS era,
       CASE WHEN sc.application_name LIKE 'action:%' THEN sc.application_name ELSE 'save_page_sections' END AS writer,
       sc.pre_deficit, sc.post_deficit,
       CASE WHEN sc.produced IS NULL THEN 'unknown'
            WHEN sc.post_deficit = 0 THEN 'REPAIRED' ELSE 'left blank' END AS outcome
  FROM scored sc JOIN pages p ON p.id = sc.page_id JOIN sites s ON s.id = sc.site_id
 WHERE sc.pre_deficit > 0
 ORDER BY sc.created_at;

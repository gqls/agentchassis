-- 090_content_block_stat_labels_llm.sql — fix fleet-wide static stat labels
-- Created 2026-07-12.
--
-- content-block-about (shared component 4e448d51, 13 pages / 5 sites) hardcodes
-- stat_1_label/stat_2_label/stat_3_label/cta_label as source='static' with the
-- business defaults 'Clients Served'/'Satisfaction Rate'/'Awards Won'/'Learn
-- More About Us'. Every render re-applies them, so the writer could only fill
-- the VALUES — producing crossed pairs on ALL sites ('500+ Models / Clients
-- Served', 'Vendor-Neutral / Awards Won', 'Longest / Clients Served' on vonc's
-- new archetype pages). The fields already carry fallback + llm_guidance: they
-- were meant to be authored. This flips their source static->llm (fallback
-- retained as the safety net), then re-authors vonc's 8 archetype pages.
--
-- SAFE for the 4 business sites: all have the labels persisted in stored
-- content_data (verified 2026-07-12), so the light rerender preserves them and
-- a future full rebuild gets BETTER per-site labels via llm_guidance. Their
-- live rendered_html is untouched until they choose to rebuild.
--
-- Reversal: _vonc_cb_backup_20260712_* (component row + vonc content_data).

BEGIN;

CREATE TABLE _vonc_cb_backup_20260712_component AS
  SELECT * FROM content_components WHERE id='4e448d51-c26f-4d9c-ac00-2d0197d3f01e';
CREATE TABLE _vonc_cb_backup_20260712_content AS
  SELECT pc.id, pc.page_id, pc.content_data
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.page_type='entity-page';

-- ── 1. Flip the 4 label fields static -> llm (keep type/fallback/guidance) ──
UPDATE content_components
SET input_schema = jsonb_set(jsonb_set(jsonb_set(jsonb_set(input_schema,
      '{fields,stat_1_label,source}','"llm"'),
      '{fields,stat_2_label,source}','"llm"'),
      '{fields,stat_3_label,source}','"llm"'),
      '{fields,cta_label,source}','"llm"'),
    updated_at = NOW()
WHERE id='4e448d51-c26f-4d9c-ac00-2d0197d3f01e';

UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Sharpest", "stat_1_label": "Structural clarity in the room", "stat_2_value": "Fastest", "stat_2_label": "On logical frameworks", "stat_3_value": "First", "stat_3_label": "To the hidden flaw", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='surgeon'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Highest", "stat_1_label": "Remix rate in the Arena", "stat_2_value": "Freest", "stat_2_label": "Movement across domains", "stat_3_value": "Least", "stat_3_label": "Predictable position filed", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='wildcard'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Best", "stat_1_label": "Prediction accuracy on record", "stat_2_value": "Rarest", "stat_2_label": "Positions filed", "stat_3_value": "Hardest", "stat_3_label": "Landings to ignore", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='oracle'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Longest", "stat_1_label": "Response chains started", "stat_2_value": "Widest", "stat_2_label": "Angles opened per Provocation", "stat_3_value": "Most", "stat_3_label": "Momentum generated", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='catalyst'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Most", "stat_1_label": "Predictive reactions in the Gallery", "stat_2_value": "Sharpest", "stat_2_label": "Pattern recognition", "stat_3_value": "Surest", "stat_3_label": "Taste under pressure", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='judge'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Most", "stat_1_label": "Consistent output on the Stage", "stat_2_value": "Strongest", "stat_2_label": "Skill shown, not claimed", "stat_3_value": "Steadiest", "stat_3_label": "Creative cadence", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='maker'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Earliest", "stat_1_label": "To breakout creators", "stat_2_value": "Keenest", "stat_2_label": "Eye for what others miss", "stat_3_value": "First", "stat_3_label": "To surface the next thing", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='scout'
  AND cc.function='content-block-about';
UPDATE page_components pc
SET content_data = pc.content_data || '{"stat_1_value": "Most", "stat_1_label": "Saved breakdowns on the platform", "stat_2_value": "Clearest", "stat_2_label": "Explanations in the room", "stat_3_value": "Deepest", "stat_3_label": "Knowledge shared forward", "cta_label": "Find Your Archetype"}'::jsonb, updated_at=NOW()
FROM content_components cc, pages p
WHERE cc.id=pc.component_id AND p.id=pc.page_id
  AND p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.name='mentor'
  AND cc.function='content-block-about';

-- ── Verify ──────────────────────────────────────────────────────────────────
DO $$
DECLARE srcs INT; badlabel INT;
BEGIN
  SELECT COUNT(*) INTO srcs
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.id='4e448d51-c26f-4d9c-ac00-2d0197d3f01e'
    AND f.key IN ('stat_1_label','stat_2_label','stat_3_label','cta_label')
    AND f.value->>'source'='llm';
  SELECT COUNT(*) INTO badlabel
  FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
  JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND p.page_type='entity-page'
    AND cc.function='content-block-about'
    AND (pc.content_data->>'stat_1_label'='Clients Served'
      OR pc.content_data->>'cta_label'='Learn More About Us');
  IF srcs<>4 OR badlabel>0 THEN
    RAISE EXCEPTION 'verify failed: srcs=% (want 4), badlabel=% (want 0)', srcs, badlabel;
  END IF;
  RAISE NOTICE 'verified: 4 label fields now llm-sourced; 8 vonc pages re-authored';
END $$;

COMMIT;


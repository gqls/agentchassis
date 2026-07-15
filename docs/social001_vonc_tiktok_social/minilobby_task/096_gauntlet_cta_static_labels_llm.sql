-- 096_gauntlet_cta_static_labels_llm.sql — flip gauntlet-cta's static labels to llm
-- Created 2026-07-15. The 090 pattern, applied to a second component.
--
-- ROOT CAUSE (same landmine as 090's content-block-about): the gauntlet-cta
-- component carries six label fields at source: "static" with generic SaaS
-- defaults — cta_primary_label "Get Started Now", cta_secondary_label
-- "See How It Works", eyebrow_label "Limited Offer", stat_1/2/3_label
-- "Happy Customers"/"Avg. Rating"/"Setup Time". A static field re-applies on
-- every render, so no content_data edit or LLM pass can override it — which is
-- why vonc's about.html and index.html show off-brand business boilerplate on a
-- daily-provocation site. (The paired *_value / *_url fields are already
-- llm / site_specs, so only the labels are stuck.)
--
-- FIX: flip the six label fields' source static -> llm, keeping each fallback as
-- the safety net (identical shape to 090). type / fallback / llm_guidance
-- untouched. After this, the content writer authors Spark-appropriate labels on
-- the next content pass; until then the fallbacks render exactly as today.
--
-- SCOPE: gauntlet-cta is a single-site component (vonc, 2 instances) but the
-- schema lives in shared content_components, so the fix carries if it is ever
-- forked. Behaviour is unchanged until a render re-authors the labels.
--
-- DELIBERATELY NOT DONE HERE: hand-authoring replacement stat labels/values.
-- The stat numbers are marketing claims; fabricating them by hand is off the
-- table (project anti-fabrication rule). A content-writer pass on the two
-- instances is the right way to replace the boilerplate — this migration only
-- removes the schema lock that was preventing it.
--
-- Reversal: _fleet_096_backup_20260715_gauntletcta holds the prior row.

BEGIN;

CREATE TABLE _fleet_096_backup_20260715_gauntletcta AS
  SELECT * FROM content_components WHERE function = 'gauntlet-cta';

UPDATE content_components
SET input_schema = jsonb_set(jsonb_set(jsonb_set(jsonb_set(jsonb_set(jsonb_set(
      input_schema,
      '{fields,cta_primary_label,source}',   '"llm"'),
      '{fields,cta_secondary_label,source}', '"llm"'),
      '{fields,eyebrow_label,source}',       '"llm"'),
      '{fields,stat_1_label,source}',        '"llm"'),
      '{fields,stat_2_label,source}',        '"llm"'),
      '{fields,stat_3_label,source}',        '"llm"'),
    updated_at = NOW()
WHERE function = 'gauntlet-cta';

-- Verify: no gauntlet-cta label field is static any more; ≥1 flipped.
DO $$
DECLARE remaining INT; flipped INT;
BEGIN
  SELECT COUNT(*) INTO remaining
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.function = 'gauntlet-cta'
    AND f.key IN ('cta_primary_label','cta_secondary_label','eyebrow_label',
                  'stat_1_label','stat_2_label','stat_3_label')
    AND f.value->>'source' = 'static';

  SELECT COUNT(*) INTO flipped
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.function = 'gauntlet-cta'
    AND f.key IN ('cta_primary_label','cta_secondary_label','eyebrow_label',
                  'stat_1_label','stat_2_label','stat_3_label')
    AND f.value->>'source' = 'llm';

  IF remaining <> 0 THEN
    RAISE EXCEPTION 'verify failed: % gauntlet-cta label fields still static', remaining;
  END IF;
  IF flipped = 0 THEN
    RAISE EXCEPTION 'verify failed: 0 label fields flipped — wrong function name or schema shape?';
  END IF;
  RAISE NOTICE 'verified: 0 static gauntlet-cta labels remain; % now llm-sourced', flipped;
END $$;

COMMIT;

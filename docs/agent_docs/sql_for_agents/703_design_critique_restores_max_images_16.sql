-- 703_design_critique_restores_max_images_16.sql — undo the 663 cap now the downscale is live.
--
-- WHY: mig 663 capped the 018 critic's vision payload at 8 images because BOTH providers
-- rejected the 16-image payload after the hero batch made full-page captures taller —
-- Gemini per-image (400 INVALID_ARGUMENT, runs 0eff246f/a21e0c3e), Anthropic per-request
-- (413 request_too_large, run 0f686e43) — and no downscaling existed in the pipeline.
-- The real fix has now shipped and is PROVEN LIVE on the fresh chassis (replicaset
-- 744cfb4bf, 2026-09-02 ~15:40Z; pod probe: 'downscaleVisionImage: image scaled for
-- provider limits' present, positive + negative controls passed): execute_vision_prompt
-- scales any image whose long edge exceeds max_image_dimension (default 7900) to fit,
-- JPEG q85. So the cap's premise is gone; restore the seed-645 coverage of 16 images
-- (~8 pages x 2 viewports). Council: submitted alongside (see commit trailer).
-- Reversible: snapshot below + inverse jsonb_set back to '8'::jsonb.

BEGIN;

SELECT snapshot_agent('design-critique-agent', '703_design_critique_restores_max_images_16.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,critique,config,max_images}', '16'::jsonb)
WHERE type = 'design-critique-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE mi int;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'critique'->'config'->>'max_images')::int INTO mi
  FROM agent_definitions WHERE type='design-critique-agent' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF mi IS DISTINCT FROM 16 THEN RAISE EXCEPTION 'max_images is % (want 16)', mi; END IF;
END $$;

COMMIT;

-- ROLLBACK: jsonb_set the same path back to '8'::jsonb, or restore the snapshot taken above.

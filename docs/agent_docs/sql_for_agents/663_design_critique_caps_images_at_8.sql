-- 663_design_critique_caps_images_at_8.sql — halve the critic's vision payload.
-- WHY: after the hero batch made full-page captures taller, BOTH providers reject the
-- 16-image payload — Gemini per-image (400 INVALID_ARGUMENT, runs 0eff246f/a21e0c3e) and
-- Anthropic per-request (413 request_too_large, run 0f686e43, model claude-sonnet-5
-- confirmed in the pod log). No downscaling exists in the pipeline (approved plan §cost
-- envelope; the real fix is a downscale in execute_vision_prompt — 018 follow-up, Go).
-- Until then: cap at 8 images. The action drops excess with a warning, so coverage drops
-- to ~4 pages × 2 viewports (or 8 × 1, per the adapter's ordering) — stated in the report
-- header line the prompt already demands. Reversible: snapshot + inverse jsonb_set.
BEGIN;
SELECT snapshot_agent('design-critique-agent', '663_design_critique_caps_images_at_8.sql: pre-update');
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{workflow,steps,critique,config,max_images}', '8'::jsonb)
WHERE type = 'design-critique-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
DO $$
DECLARE mi int;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'critique'->'config'->>'max_images')::int INTO mi
  FROM agent_definitions WHERE type='design-critique-agent' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF mi <> 8 THEN RAISE EXCEPTION 'max_images is % (want 8)', mi; END IF;
END $$;
COMMIT;
-- ROLLBACK: jsonb_set the same path back to '16'::jsonb, or restore the snapshot.

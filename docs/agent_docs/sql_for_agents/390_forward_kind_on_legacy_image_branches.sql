-- 390 — forward `kind` on image-build-handler's legacy logo/hero branches, so the
-- provider routing table can actually see them.
--
-- THE DEFECT (found 2026-08-11, bugs_open/210 needs_logo slug, owner-rejected hero).
-- The adapter routes hero/logo to Banana (Google Nano Banana family; default model
-- NanoBananaPro) — internal/adapters/imagegenerator/routing.go, live in the running
-- adapter (provenance-checked: commit 6896ce22e is an ancestor of the running
-- c3b424c8e). Yet the mortgagecalculator hero generated on
-- stability/stable-diffusion-xl-1024-v1-0. Why:
--
--   resolveKind (generate_image_actions.go:115-123) reads inputData["kind"], then
--   inputData["default_kind"], else "". call_logo_gen/call_hero_gen forward ONLY
--   prompt/site_id/site_plan; their `default_kind` lives in the parent step CONFIG,
--   which is not input_data for the callee. So kind arrives EMPTY, and the
--   adapter's empty-kind fallback is stability (routing_test.go:33 pins exactly
--   that). The routing table's hero->Banana row is unreachable from these branches.
--
-- Same class as bugs_open/231: a static config value for a field the spec was
-- meant to supply is DEAD. `default_kind` here has never done anything.
--
-- THE FIX: map kind from the item spec's `purpose`, which these items always carry
-- and whose values ("logo","hero") are exactly the routing table's kind names.
-- Optional ("kind?") so a hypothetical purposeless item degrades to today's
-- behaviour rather than dying at the allow-list (input_mapping.go:122-136).
--
-- Blast radius: the two legacy branches only. The Phase-2E branches
-- (call_imagery_gen, call_variant_gen) already forward kind. Effect: legacy
-- logo/hero generations move SDXL -> Banana per the routing table's own policy
-- ("hero — 2026-07-18 (bugs_open/011). The last kind left behind, and left by
-- omission"). A site that wants SDXL heroes sets provider:"stability" in its
-- imagery_style_guide — data, not code — which this preserves.
--
-- Config-only: live at the next dispatch, no image roll.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config,
        '{workflow,steps,call_logo_gen,config,input_mapping,kind?}',
        '"input_data.spec.purpose"'),
      '{workflow,steps,call_hero_gen,config,input_mapping,kind?}',
      '"input_data.spec.purpose"'),
    updated_at = now()
WHERE type='image-build-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify: both mappings must now carry the key; refuse the commit otherwise.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='image-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'call_logo_gen'->'config'->'input_mapping' ? 'kind?'
     AND default_config->'workflow'->'steps'->'call_hero_gen'->'config'->'input_mapping' ? 'kind?';
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 image-build-handler row with kind? on both branches, found %', n;
  END IF;
END $$;

COMMIT;

-- ROLLBACK (manual):
-- UPDATE agent_definitions SET default_config =
--   (default_config #- '{workflow,steps,call_logo_gen,config,input_mapping,kind?}')
--   #- '{workflow,steps,call_hero_gen,config,input_mapping,kind?}'
-- WHERE type='image-build-handler' AND is_active
--   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 360_brand_update_branch_purpose_from_spec.sql
--
-- bugs_open/235 — image-build-handler's brand-update branch stores EVERY asset
-- as purpose "hero", so a logo work item ships as logo.jpg at hero processing.
--
-- THE DEFECT. `store_imagery_brand_asset` carries a STATIC `"purpose": "hero"`,
-- on a branch whose own description says it handles "logo or canonical index
-- hero" (rule b). One static cannot be right for two purposes: it is right for
-- the hero arm and wrong for the logo arm. The item's own `spec.purpose` says
-- exactly which it is, and is not read.
--
-- The damage is downstream and silent: `deploy_image_asset` takes the file
-- EXTENSION and the resize class from purpose and the FILENAME from asset_key
-- (storage.DeployedAssetPath / DownloadOptimizeAndPrepare), so asset_key "logo"
-- + purpose "hero" publishes `logo.jpg` at hero dimensions — JPEG, therefore no
-- alpha channel — instead of `logo.png` at 400x400. Eleven live sites are in
-- that state (census in bugs_open/231's 2026-08-09 contribution).
--
-- THE FIX. Drop the static and read the spec, exactly as this step's own sibling
-- `store_imagery_asset` already does (`purpose_field: input_data.spec.purpose`).
--
-- WHY BOTH KEYS MUST MOVE TOGETHER. StoreAssetAction resolves purpose by
-- literal-first priority (v3_site_actions.go:2662-2670):
--     1. config["purpose"]        — literal
--     2. config["purpose_field"]  — path into collected_data, ONLY if 1 is empty
-- so adding `purpose_field` while leaving `purpose` in place changes NOTHING.
-- Note this is a different mechanism from bugs_open/231's Defaults shadow — here
-- the static resolves fine, it is simply the wrong value. Same artefact, other
-- door; do not conflate them when reasoning about either.
--
-- SAFE FOR BOTH ARMS — measured 2026-08-09, not assumed. Every live
-- needs_imagery item with brand_update=true carries spec.purpose:
--     asset_key=logo      -> spec.purpose=logo   (4 items, 08-02..08-08)
--     asset_key=hero_home -> spec.purpose=hero   (4 items, 08-02..08-09)
-- None is missing it. So the hero arm keeps resolving "hero" and is unchanged;
-- only the logo arm's behaviour moves, which is the bug.
--
-- IF spec.purpose IS EVER ABSENT, StoreAssetAction's purpose ends up "" and the
-- asset row is written with a NULL purpose rather than being mislabelled "hero".
-- That is the honest failure and is deliberately preferred here: a NULL purpose
-- is visible, a wrong one is not.
--
-- NOT FIXED BY THIS MIGRATION: the 11 artefacts already published. They need
-- re-deploying through the repaired branch — bugs_open/235 fix candidate 2, and
-- note the stale `logo.jpg` is not removed by a deploy, so check page references
-- before deleting it (robot-hands' pages reference /assets/images/logo.jpg).

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config #- '{workflow,steps,store_imagery_brand_asset,config,purpose}',
        '{workflow,steps,store_imagery_brand_asset,config,purpose_field}',
        '"input_data.spec.purpose"'::jsonb,
        true
    ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,store_imagery_brand_asset}' IS NOT NULL;

-- Post-verify. A block of SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a
-- non-empty result), so this is DO/RAISE — and it was INDUCED against the
-- unmigrated row first, where it raised "0 of 1", which is what makes a later
-- "1 of 1" mean anything.
DO $$
DECLARE
    ok_count int;
    total    int;
BEGIN
    SELECT count(*) INTO total
    FROM agent_definitions
    WHERE type = 'image-build-handler' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,store_imagery_brand_asset}' IS NOT NULL;

    SELECT count(*) INTO ok_count
    FROM agent_definitions
    WHERE type = 'image-build-handler' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,store_imagery_brand_asset,config,purpose_field}'
          = 'input_data.spec.purpose'
      AND default_config #> '{workflow,steps,store_imagery_brand_asset,config,purpose}' IS NULL;

    IF total = 0 THEN
        RAISE EXCEPTION '360 verify: no live image-build-handler row carries store_imagery_brand_asset — wrong target or the step was renamed';
    END IF;
    IF ok_count <> total THEN
        RAISE EXCEPTION '360 verify FAILED: % of % rows have purpose_field set AND the static purpose removed. Both must move together — a lingering static purpose wins over the field (v3_site_actions.go:2662).', ok_count, total;
    END IF;
    RAISE NOTICE '360 verify OK: % of % image-build-handler rows read purpose from input_data.spec.purpose', ok_count, total;
END $$;

COMMIT;

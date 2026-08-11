-- 380_build_dispatch_loop_maps_purpose_top_level.sql
--
-- bugs_open/231 (2026-08-10 contribution) — the `undeployed_asset` dispatch
-- shape defeats asset-deployer's purpose binding, so every such deploy falls to
-- DeployImageAssetInputSpec's Default {"purpose":"hero"}. Live specimen:
-- relojistas.com, two runs (08-10 17:01/17:03), item spec purpose='logo' AND
-- the asset row purpose='logo', both committed "Deploy hero image".
--
-- THE DEFECT. build-dispatch-loop's call_handler maps the item spec NESTED
-- ("spec": "current_item.spec") with no top-level purpose. asset-deployer's
-- deploy_asset binds "purpose": "input_data.purpose" — a Strategy-0 dotted path
-- that resolves nothing on this shape. Because `purpose` carries a spec Default,
-- every later strategy skips it (action_inputs.go: Defaults first at ~:544, the
-- has-value skips at Strategies 1-4), so the Default wins over the correct value
-- sitting one level down at input_data.spec.purpose. Only fields WITH a Default
-- break this way: s3_uri and asset_key have none, so recursive search finds
-- their spec values — which is why the bad runs fetched the right artwork and
-- still resized it as a hero.
--
-- WHY NOT `purpose_field` ON THE DEPLOY STEP (the earlier fix candidate,
-- HANDOFF_2026-08-10b option 1): the Deprecated *_field bridge is Strategy 3,
-- which skips any field that already has a value — and a defaulted field always
-- has one. Pinned by TestPurposeFieldBridge_DeadForDefaultedField. That
-- candidate is dead config; this migration is the fix that actually moves the
-- value.
--
-- THE FIX. Map purpose top-level in the dispatch mapping, optional:
--     "purpose?": "current_item.spec.purpose"
-- The `?` suffix (input_contracts/input_mapping.go:102) silently skips items
-- whose spec has no purpose — a no-op for them. This is the estate's OWN idiom:
-- site-work-orchestrator's fix_items_loop.call_handler already carries
-- "purpose?": "current_fix_item.spec.purpose"; build-dispatch-loop lacking it
-- is an omission, not a design choice.
--
-- BLAST RADIUS — measured 2026-08-11, not assumed. Live definitions binding
-- input_data.purpose anywhere: exactly two (query over active agent_definitions).
--   1. asset-deployer deploy_asset — the fix target. undeployed_asset,
--      needs_content_image and needs_brand_head_assets items now resolve their
--      real purpose. Brand-head purposes (favicon/og_card, 11 known no-mode
--      items) draw the bugs_open/179-B refusal instead of silently resolving
--      "hero" — the guard fires on the RESOLVED purpose, so this converts a
--      latent mis-deploy into a clean decline.
--   2. image-build-handler check_logo_or_hero — condition
--      "input_data.item_type == 'needs_logo' OR input_data.purpose == 'logo'".
--      The second arm is half-dead today (unresolvable on this shape); it
--      activates, so a needs_imagery item with spec.purpose='logo' (the
--      brand-update shape, bugs_open/235) routes down the LOGO generation
--      branch instead of the hero one. That is the condition's stated intent
--      ("Existing logo-or-hero split preserved") and the 235 family fix.
-- Item types carrying spec.purpose (8, census in NOTES_209 2026-08-11): the
-- remaining handlers (page-build-handler) bind no input_data.purpose — the key
-- appears in their input_data with the same value spec already carries; inert.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,purpose?}',
        '"current_item.spec.purpose"'::jsonb,
        true
    ),
    updated_at = now()
WHERE type = 'build-dispatch-loop'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}' IS NOT NULL;

-- Post-verify. DO/RAISE (a SELECT block cannot stop a COMMIT) — induced against
-- the unmigrated row first, where it raised "0 of 1".
DO $$
DECLARE
    ok_count int;
    total    int;
BEGIN
    SELECT count(*) INTO total
    FROM agent_definitions
    WHERE type = 'build-dispatch-loop' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}' IS NOT NULL;

    SELECT count(*) INTO ok_count
    FROM agent_definitions
    WHERE type = 'build-dispatch-loop' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,purpose?}'
          = 'current_item.spec.purpose';

    IF total = 0 THEN
        RAISE EXCEPTION '380 verify: no live build-dispatch-loop row carries process_item...call_handler.input_mapping — wrong target or the step was renamed';
    END IF;
    IF ok_count <> total THEN
        RAISE EXCEPTION '380 verify FAILED: % of % rows map purpose? -> current_item.spec.purpose', ok_count, total;
    END IF;
    RAISE NOTICE '380 verify OK: % of % build-dispatch-loop rows map purpose top-level for the handler call', ok_count, total;
END $$;

COMMIT;

-- 467_asset_deployer_brand_head_purpose_fallback.sql
--
-- asset-deployer: route a brand-head-PURPOSED item to the deriver when it
-- carries no mode, instead of letting it fall through to a branch that can
-- only ever refuse it.  bugs_open/131 (og-card slug), 2026-08-18.
--
-- THE DEFECT THIS CLOSES (measured 2026-08-18, live DB + live wire)
-- -----------------------------------------------------------------
-- asset-deployer's entry chain keys every conditional on spec.mode/mode:
--   check_mode(brand_head) -> check_sprite_mode -> check_card_mode
--     -> check_ingest_mode -> deploy_asset
-- The discovery producer of needs_brand_head_assets items
-- (check_undeployed_assets.go) filed them with spec.purpose but NO spec.mode,
-- so every one fell through to deploy_asset. deploy_image_asset REFUSES
-- brand-head purposes by design (it is not their writer; the refusal names
-- 're-derive it (mode=brand_head)' as the remedy), and the refusal-as-result
-- completes the workflow — so the items stamped 'complete' with nothing
-- derived. 21 complete items; webdesign.co.uk, cookly.uk, lendzy.co.uk,
-- loancalculator.co.uk serve 404 for both artefacts today. The one hand-filed
-- item carrying mode='brand_head' (idea.uk, 2026-08-17) routed correctly and
-- both its artefacts serve 200.
--
-- THE CHANGE
-- ----------
-- One new conditional, LAST in the chain — between check_ingest_mode and
-- deploy_asset:
--
--   check_brand_head_purpose:
--     spec.purpose == 'favicon' OR spec.purpose == 'og_card'
--       -> derive_head_assets     (the action the refusal text itself names)
--       else -> deploy_asset      (unchanged for everything else)
--
-- WHY LAST AND NOT A WIDENING OF check_mode: an explicit mode must always
-- win. Widening check_mode's condition with purpose would hijack
-- sprite/card/ingest-mode items that happen to carry a brand-head purpose
-- field (ingest_staged_asset legitimately handles staged brand artefacts).
-- Placed after every mode check, the fallback can only catch items that
-- would otherwise reach deploy_asset — and deploy_asset unconditionally
-- refuses brand-head purposes (deploy_image_asset_action.go:229-242), so no
-- reachable behaviour is taken away from any caller: the fallback converts a
-- guaranteed refusal into the action the refusal prescribes.
--
-- CONSUMERS, enumerated not asserted (owner rulings 2026-07-29 #3, 2026-08-11
-- RFC_022): producers of asset-deployer items whose routing this can change =
-- items with a brand-head spec.purpose and no mode. Live census 2026-08-18:
-- one code producer (check_undeployed_assets.go — fixed to emit
-- mode='brand_head' in the same commit that ships this file) and hand-filed
-- redrive items (which carry mode already). Zero items with a brand-head
-- purpose have EVER deployed via deploy_asset — the refusal makes that
-- population empty by construction.
--
-- DRIFT RISK, stated: the two purposes are hardcoded here because the DSL
-- cannot read Go's storage.BrandHeadAssetPaths. If that map ever gains a
-- purpose, items for it fall through to deploy_asset exactly as before —
-- refused, and now BLOCKED at completion by VerifyBrandHeadAssetsResolved
-- (same commit), so the drift surfaces as visible failed items instead of
-- silent completes.
--
-- ORDER: DB config, live the moment it commits. The Go halves (producer mode,
-- verifier) ride the next chassis roll; no ordering constraint exists in
-- either direction — this fallback alone heals the live population on
-- redrive, the producer fix alone routes future items, and neither breaks
-- while the other is absent.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 467_asset_deployer_brand_head_purpose_fallback.sql
--
-- ROLLBACK: restore from the snapshot this file takes (agent_definitions_backup,
-- ORDER BY snapshot_taken_at DESC — the backup keeps the source row's id and
-- created_at, so snapshot_taken_at is the only honest ordering), or:
--   UPDATE agent_definitions SET default_config =
--     jsonb_set(default_config #- '{workflow,steps,check_brand_head_purpose}',
--               '{workflow,steps,check_ingest_mode,config,else_step}', '"deploy_asset"')
--   WHERE type='asset-deployer' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('asset-deployer', '467_asset_deployer_brand_head_purpose_fallback.sql: pre-update');

-- Before-assertions: the chain is the one this file was written against.
DO $pre$
DECLARE v_else text; v_rows int; v_new_step jsonb;
BEGIN
    SELECT count(*) INTO v_rows FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_rows <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 active asset-deployer definition, found %', v_rows;
    END IF;

    SELECT default_config #>> '{workflow,steps,check_ingest_mode,config,else_step}'
      INTO v_else
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_else IS DISTINCT FROM 'deploy_asset' THEN
        RAISE EXCEPTION 'check_ingest_mode.else_step is % (expected deploy_asset) — the chain has drifted; re-read the live config before applying', v_else;
    END IF;

    SELECT default_config #> '{workflow,steps,check_brand_head_purpose}'
      INTO v_new_step
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_new_step IS NOT NULL THEN
        RAISE EXCEPTION 'check_brand_head_purpose already exists — this migration has already been applied, or another session shipped it';
    END IF;
END $pre$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
             default_config,
             '{workflow,steps,check_brand_head_purpose}',
             '{
                "action": "conditional",
                "config": {
                    "condition": "input_data.spec.purpose == \"favicon\" OR input_data.spec.purpose == \"og_card\"",
                    "then_step": "derive_head_assets",
                    "else_step": "deploy_asset"
                },
                "description": "Fallback for brand-head-purposed items filed without a mode: deploy_asset can only ever refuse them (deploy_image_asset brand-head guard), and its refusal names re-derivation as the remedy — this step is that remedy applied at the router. Explicit modes always win; this runs last. bugs_open/131."
             }'::jsonb),
         '{workflow,steps,check_ingest_mode,config,else_step}',
         '"check_brand_head_purpose"'),
       updated_at = NOW()
 WHERE type = 'asset-deployer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- After-assertions: the new step is wired, and nothing else moved.
DO $verify$
DECLARE v_cfg jsonb; v_steps int;
BEGIN
    SELECT default_config INTO v_cfg
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_cfg #>> '{workflow,steps,check_ingest_mode,config,else_step}' <> 'check_brand_head_purpose' THEN
        RAISE EXCEPTION 'check_ingest_mode.else_step was not repointed';
    END IF;
    IF v_cfg #>> '{workflow,steps,check_brand_head_purpose,config,then_step}' <> 'derive_head_assets' THEN
        RAISE EXCEPTION 'check_brand_head_purpose.then_step is not derive_head_assets';
    END IF;
    IF v_cfg #>> '{workflow,steps,check_brand_head_purpose,config,else_step}' <> 'deploy_asset' THEN
        RAISE EXCEPTION 'check_brand_head_purpose.else_step is not deploy_asset';
    END IF;
    IF v_cfg #>> '{workflow,start_step}' <> 'check_mode' THEN
        RAISE EXCEPTION 'start_step moved — it must remain check_mode';
    END IF;

    SELECT count(*) INTO v_steps FROM jsonb_object_keys(v_cfg #> '{workflow,steps}');
    IF v_steps <> 11 THEN
        RAISE EXCEPTION 'expected 11 workflow steps after the change (10 before + 1 new), found %', v_steps;
    END IF;

    RAISE NOTICE 'asset-deployer: brand-head-purposed, mode-less items now route to derive_head_assets (last-in-chain fallback)';
END $verify$;

COMMIT;

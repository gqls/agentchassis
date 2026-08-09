-- 348_pageflow_swo_deploy_steps_resolve_by_identity.sql
--
-- bugs_open/209 "into line" PHASE 1 + the bugs_open/231 repair, in one edit.
-- Owner approval 2026-08-09 ("carry on with the into-line fix ... Phase 1
-- first"), after the ruling that pageflow-builder / site-work-orchestrator are
-- "not dead, but not being worked on".
--
-- THE TWO DEFECTS THIS CLOSES (both measured, both pinned in
-- platform/orchestration/actions/deploy_image_asset_purpose_source_test.go):
--
--   1. bugs_open/231: the deploy steps' static `"purpose": "logo"` is DEAD.
--      ExtractActionInputs applies spec Defaults{purpose:"hero"} first
--      (action_inputs.go:457-460); Strategies 1/2/3 skip already-set fields
--      (:499,:511,:523); Strategy 0 reads only DOTTED config paths (:478).
--      A static value has no strategy that can carry it, so deploy_logo_image
--      resolves purpose="hero" — hero resize class, and the logo's bytes
--      committed to the HERO's deploy path (url_helpers.go:190,
--      filename = purpose + ext). Broken-if-run since 34d2315ce (2026-02-20).
--
--   2. bugs_open/209: with no input_fields, `asset_id` resolves by AGGRESSIVE
--      RECURSIVE SEARCH over collected_data in randomised map order — the logo
--      step picked up the hero's asset_id in 344/400 measured runs. And the
--      purpose-keyed findStorageURI fallback is last-write-wins per purpose.
--
-- THE CHANGE, config only, live on apply — for BOTH definitions' TWO deploy
-- steps (4 step-configs total):
--
--   REMOVE  "uri_field": "{p}_result.image_uri"   (generator URI; identity-blind)
--   REPLACE "purpose":  "hero"|"logo"  ->  "{p}_stored.purpose"     (Strategy 0)
--   ADD     "s3_uri":   "{p}_stored.s3_uri"                          (Strategy 0)
--   ADD     "asset_id": "{p}_stored.asset_id"                        (Strategy 0)
--   ADD     "domain":   "site_record.domain"                         (Strategy 0)
--   ADD     "input_fields": ["purpose","domain","asset_id"]
--
-- Strategy-0 dotted paths are the ONE mechanism that defeats both the Defaults
-- shadow and the recursive search — proven deterministic 100/100
-- (TestStrategy0DottedPaths_DefeatTheDefaultAndTheRecursiveSearch).
-- {p}_stored is each step's OWN store step's output_field (verified live:
-- store_hero_asset -> hero_stored, store_logo_asset -> logo_stored, both
-- definitions), and StoreAssetAction's success return carries purpose, s3_uri
-- and asset_id (v3_site_actions.go). StoreAssetAction reads its purpose from
-- config DIRECTLY (v3_site_actions.go:2603-2611, no ActionInputSpec) so it is
-- NOT shadowed — the dotted paths carry true values.
--
-- WHY input_fields DELIBERATELY EXCLUDES s3_uri: Strategy 0 still resolves it
-- (it iterates SPEC fields, independent of input_fields), but on a store
-- failure ({p}_stored.s3_uri absent) an input_fields entry would send Strategy 1
-- hunting "s3_uri" by aggressive search — which can land on the SIBLING asset's
-- URI (the 209 class through a side door). Excluded, the failure corner
-- resolves s3_uri="" -> asset_id (present even on store-insert failure, but its
-- row does not exist) -> resolveStorageURIFromAsset returns "" -> the action
-- SKIPS with deployed:false — a safe, visible degradation instead of
-- wrong bytes. (Today's behaviour in that corner deploys the generator's URI,
-- bypassing the asset row entirely — the 152/155 identity-bypass pattern. The
-- narrowing is deliberate and desired.)
--
-- CONSUMERS (owner ruling 2026-07-29 #3): the two edited workflows ARE the
-- consumers, and the owner approved this change for them explicitly.
-- asset-deployer and image-build-handler are untouched (their own
-- input_fields/input_mapping already carry identity). No shared Go changes here.
--
-- ORDERING NOTE FOR PHASE 2 (statable, so stated): the Go deletion of
-- findStorageURI must NOT reach a fleet image before this migration is applied
-- and verified — until then the purpose-keyed lookups are these workflows'
-- correct fallback. Sequence inside the 209 lane: apply + row-verify 348, then
-- build.
--
-- VERIFY (behavioural, owed after apply): one sacrificial-domain run of each
-- workflow; assert hero.* and logo.* both committed with DIFFERENT bytes and
-- content_data.logo_url serving 200. Row-level verify is the DO block below.
--
-- ROLLBACK: 348_pageflow_swo_deploy_steps_resolve_by_identity_ROLLBACK.sql
-- (sidecar, excluded from runs) restores the exact prior shape.

BEGIN;

-- ---------------------------------------------------------------------------
-- PRE-GUARDS: abort unless the live shape is exactly the one this file edits.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  n_defs int;
  n_deploy int;
  n_store int;
BEGIN
  SELECT count(*) INTO n_defs
  FROM agent_definitions
  WHERE type IN ('pageflow-builder','site-work-orchestrator')
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n_defs <> 2 THEN
    RAISE EXCEPTION '348: expected exactly 2 live definitions, found % — census drifted, re-read before applying', n_defs;
  END IF;

  -- The four deploy steps must still carry the LEGACY shape (static purpose +
  -- uri_field, no input_fields). If not, another session got here first or the
  -- shape drifted: stop and re-diagnose rather than layering.
  SELECT count(*) INTO n_deploy
  FROM agent_definitions ad,
       LATERAL (VALUES
         ('deploy_hero_image','hero','hero_result.image_uri'),
         ('deploy_logo_image','logo','logo_result.image_uri')
       ) AS want(step_name, want_purpose, want_uri_field),
       LATERAL (SELECT ad.default_config#>ARRAY['workflow','steps',want.step_name] AS step) s
  WHERE ad.type IN ('pageflow-builder','site-work-orchestrator')
    AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.step->>'action' = 'deploy_image_asset'
    AND s.step->'config'->>'purpose'  = want.want_purpose
    AND s.step->'config'->>'uri_field' = want.want_uri_field
    AND s.step->'config'->'input_fields' IS NULL;
  IF n_deploy <> 4 THEN
    RAISE EXCEPTION '348: expected 4 legacy-shaped deploy steps, found % — shape drifted, do not apply blind', n_deploy;
  END IF;

  -- The dotted paths this file installs must point at outputs that exist:
  -- each store step must write the matching {p}_stored output_field.
  SELECT count(*) INTO n_store
  FROM agent_definitions ad,
       LATERAL (VALUES
         ('store_hero_asset','hero_stored'),
         ('store_logo_asset','logo_stored')
       ) AS want(step_name, want_out),
       LATERAL (SELECT ad.default_config#>ARRAY['workflow','steps',want.step_name] AS step) s
  WHERE ad.type IN ('pageflow-builder','site-work-orchestrator')
    AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND s.step->>'action' = 'store_asset'
    AND s.step->>'output_field' = want.want_out;
  IF n_store <> 4 THEN
    RAISE EXCEPTION '348: expected 4 store steps with {p}_stored output_fields, found % — the dotted paths would dangle', n_store;
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- THE EDIT: hero steps, then logo steps, both definitions each.
-- ---------------------------------------------------------------------------
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,deploy_hero_image,config}',
      (default_config#>'{workflow,steps,deploy_hero_image,config}') - 'uri_field'
        || '{"purpose":"hero_stored.purpose","s3_uri":"hero_stored.s3_uri","asset_id":"hero_stored.asset_id","domain":"site_record.domain","input_fields":["purpose","domain","asset_id"]}'::jsonb
    ),
    updated_at = NOW()
WHERE type IN ('pageflow-builder','site-work-orchestrator')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,deploy_logo_image,config}',
      (default_config#>'{workflow,steps,deploy_logo_image,config}') - 'uri_field'
        || '{"purpose":"logo_stored.purpose","s3_uri":"logo_stored.s3_uri","asset_id":"logo_stored.asset_id","domain":"site_record.domain","input_fields":["purpose","domain","asset_id"]}'::jsonb
    ),
    updated_at = NOW()
WHERE type IN ('pageflow-builder','site-work-orchestrator')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- POST-VERIFY: DO/RAISE (a SELECT cannot stop the COMMIT). Induced before
-- apply by running this block standalone against the unmigrated rows — it must
-- RAISE there, proving it can fail.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  n_ok int;
BEGIN
  SELECT count(*) INTO n_ok
  FROM agent_definitions ad,
       LATERAL (VALUES ('deploy_hero_image','hero'), ('deploy_logo_image','logo')) AS want(step_name, p),
       LATERAL (SELECT ad.default_config#>ARRAY['workflow','steps',want.step_name,'config'] AS cfg) c
  WHERE ad.type IN ('pageflow-builder','site-work-orchestrator')
    AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
    AND c.cfg->>'purpose'  = want.p || '_stored.purpose'
    AND c.cfg->>'s3_uri'   = want.p || '_stored.s3_uri'
    AND c.cfg->>'asset_id' = want.p || '_stored.asset_id'
    AND c.cfg->>'domain'   = 'site_record.domain'
    AND c.cfg->'uri_field' IS NULL
    AND c.cfg->'input_fields' = '["purpose","domain","asset_id"]'::jsonb;
  IF n_ok <> 4 THEN
    RAISE EXCEPTION '348 POST-VERIFY FAILED: % of 4 deploy steps carry the identity shape — rolling back', n_ok;
  END IF;
END $$;

COMMIT;

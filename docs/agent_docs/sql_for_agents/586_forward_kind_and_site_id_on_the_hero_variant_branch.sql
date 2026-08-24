-- 586 — forward `kind` AND `site_id` on image-build-handler's hero-VARIANT branch,
-- and delete the three dead `default_kind` keys that made the gap look covered.
--
-- bugs_open/382. Sibling of 390, which fixed this exact class on the two legacy
-- branches on 2026-08-11 and then asserted, in its own header:
--
--     "Blast radius: the two legacy branches only. The Phase-2E branches
--      (call_imagery_gen, call_variant_gen) already forward kind."
--
-- That is FALSE for call_variant_gen and is still false in the live row today.
-- Its input_mapping is {"prompt": "input_data.spec.prompt",
-- "site_plan": "input_data.spec"} — no kind, no site_id — while its config
-- carries "default_kind": "hero", which reads exactly like the thing that would
-- have saved it and is read by NOTHING (390's own header says so: a call_agent
-- step's callee receives input_mapping and nothing else — extractDataForAgent,
-- platform/orchestration/actions/call_agent.go).
--
-- call_variant_gen is the ONLY handler of `unfulfilled_hero_variant` work items
-- (check_item_type routes them to spawn_image_gen_variant), i.e. EVERY per-page
-- hero — hero_about, hero_services, hero_case_studies …
--
-- THE EVIDENCE, and why it settles the ordering question 390 left open. 390 was
-- applied 2026-08-11 13:42 BST (commit 8bb2194d6). LATER THAT DAY, five
-- unfulfilled_hero_variant items completed on ai-agent-orchestration.com at
-- 16:28:33 / 16:37:23 / 16:38:33 / 16:39:58 / 16:42:01, and the five SDXL hero
-- assets on that site are stamped 16:28:04 / 16:36:53 / 16:38:08 / 16:39:30 /
-- 16:41:21 (asset_keys hero_about / hero_services / hero_contact / hero_tools /
-- hero_case_studies, origin_model stability/stable-diffusion-xl-1024-v1-0).
-- Each item completes ~30 s after the asset it produced. Post-390, still SDXL.
--
-- ── EDIT 1: kind? ← spec.purpose ─────────────────────────────────────────────
-- Same shape and same source as 390. classifyPromptKey
-- (discovery_checks/check_unfulfilled_image_prompt.go) returns purpose="hero"
-- for every hero_<page> key, and "hero" is a routing-table kind. Optional
-- ("kind?") for 390's stated reason: a hypothetical purposeless item degrades
-- rather than dying at the allow-list. store_variant_asset in this same
-- workflow already reads input_data.spec.purpose, so the value is PROVEN
-- present on this path, not assumed.
--
-- ── EDIT 2: site_id ← site_record.site_id ────────────────────────────────────
-- A SECOND, SEPARATE DEFECT on the same step, called out rather than smuggled
-- in with the first. With no site_id, generate_image_actions calls
-- getImageryStyleGuideForSite with "" and gets nil, so hero VARIANTS are
-- generated with no style guide, no per-site `provider` preference, no `avoid`
-- terms, no reference-image anchors and no design_intent.imagery_direction —
-- while the canonical hero on the very same site gets all of them.
--
-- REQUIRED, not optional, and the evidence for that choice: ensure_site_record
-- runs before check_item_type on the only path into this branch, and both
-- sibling steps (call_hero_gen, call_imagery_gen) already map site_id as
-- REQUIRED from site_record.site_id and work. So the guarantee is proven on
-- this path by two live consumers. If ensure_site_record ever stopped
-- populating it, this step fails to its configured error_step
-- (mark_work_item_failed) — a recorded failure, which this shop prefers to a
-- silently style-less brand asset (the same ruling that made generate_image
-- REFUSE its generic fallback prompt rather than paint from it).
--
-- ── EDIT 3: delete the dead `default_kind` keys ──────────────────────────────
-- All three, on call_hero_gen / call_logo_gen / call_variant_gen. Nothing reads
-- them (measured: `default_kind` appears on exactly 3 live call_agent steps
-- fleet-wide as of 2026-08-24, all in this agent, and call_agent's config
-- readers are agent_type / agent_type_field / await_response / input_data /
-- input_field / input_fields / target_role plus input_mapping). On
-- call_hero_gen and call_logo_gen they were superseded by 390's kind? mapping;
-- on call_variant_gen the dead key IS the reason the missing mapping looked
-- covered for three months. A config value that reads as a safety net and is
-- not one is worse than no value at all.
--
-- Config-only: live at the next dispatch, no image roll. The CODE half of 382
-- (an absent kind routes to Banana and emits MISSING_IMAGE_KIND) is a separate
-- commit and is inert until the image-generator adapter rolls; the two are
-- independent — this migration is correct with or without it, and vice versa.

BEGIN;

SELECT snapshot_agent('image-build-handler', '586: pre-update (bugs_open/382 hero-variant kind + site_id)');

UPDATE agent_definitions
SET default_config =
      -- edit 3: drop the three dead keys
      (
        (
          (
            -- edits 1 + 2: the two mappings call_variant_gen never had
            jsonb_set(
              jsonb_set(default_config,
                '{workflow,steps,call_variant_gen,config,input_mapping,kind?}',
                '"input_data.spec.purpose"'),
              '{workflow,steps,call_variant_gen,config,input_mapping,site_id}',
              '"site_record.site_id"')
          ) #- '{workflow,steps,call_variant_gen,config,default_kind}'
        ) #- '{workflow,steps,call_hero_gen,config,default_kind}'
      ) #- '{workflow,steps,call_logo_gen,config,default_kind}',
    updated_at = now()
WHERE type='image-build-handler' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify, as a DO block that RAISEs. A verify block made of bare SELECTs
-- CANNOT stop the COMMIT — ON_ERROR_STOP ignores a non-empty result set — so
-- the guard has to be able to throw, and it is written to throw on the exact
-- states this migration could leave behind.
DO $$
DECLARE
  n_mapped   int;
  n_deadkeys int;
BEGIN
  SELECT count(*) INTO n_mapped FROM agent_definitions
   WHERE type='image-build-handler' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'call_variant_gen'->'config'->'input_mapping' ? 'kind?'
     AND default_config->'workflow'->'steps'->'call_variant_gen'->'config'->'input_mapping' ? 'site_id'
     AND default_config #>> '{workflow,steps,call_variant_gen,config,input_mapping,kind?}' = 'input_data.spec.purpose'
     AND default_config #>> '{workflow,steps,call_variant_gen,config,input_mapping,site_id}' = 'site_record.site_id';
  IF n_mapped <> 1 THEN
    RAISE EXCEPTION '586: expected exactly 1 image-build-handler row with kind? AND site_id mapped on call_variant_gen, found %', n_mapped;
  END IF;

  SELECT count(*) INTO n_deadkeys FROM agent_definitions ad,
       jsonb_each(ad.default_config->'workflow'->'steps') s
   WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     AND s.value->>'action'='call_agent'
     AND s.value->'config' ? 'default_kind';
  IF n_deadkeys <> 0 THEN
    RAISE EXCEPTION '586: % live call_agent step(s) still carry the dead default_kind key, expected 0', n_deadkeys;
  END IF;
END $$;

COMMIT;

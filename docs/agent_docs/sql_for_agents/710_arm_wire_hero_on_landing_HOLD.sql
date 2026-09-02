-- 710_arm_wire_hero_on_landing_HOLD.sql
--
-- ############################################################################
-- ##  _HOLD — DO NOT APPLY UNTIL A CHASSIS IMAGE CARRYING                    ##
-- ##  wire_page_hero_on_landing.go HAS ROLLED (gate commit: the one that     ##
-- ##  ships this file with the code).                                        ##
-- ############################################################################
--
-- bugs_open/412 fix candidate 1, armed: sets `wire_hero_on_landing: true` on
-- image-build-handler's `flag_rebuild` step, so a landing content hero's
-- deployed URL is written into the page's stored hero fields at the event, in
-- the same transaction as the re-render emit. Built by the
-- bugfix_114_imagery_wiring lane under 412's recorded handover (§10).
-- Full argument + refusal rules: wire_page_hero_on_landing.go's header.
--
-- WHY THE HOLD. The config key is read by heroWireArmed in the NEW binary
-- only. Applying early is harmless-but-lying (the old binary ignores unknown
-- config), and worse, the unknown-config-key audit would flag it as a stray
-- key on a fleet that cannot honour it — a config asserting behaviour the
-- running code does not have is exactly the `a-config-key-that-nothing-reads`
-- class this bug's GAP 1 was (IMG-072: `update_site_brand_assets`, declared
-- for months, read by nothing). Config after image, always.
--
-- WHY OPT-IN AT ALL (and not just ship it on): owner ruling 2026-08-02 §2 —
-- new authority on a shared seam ships as a field whose unsafe default is OFF.
-- The seam is page_components.content_data, which every composition path
-- writes; this adds a writer, narrowly gated (hero-family, non-fragment,
-- fills-only-fallback), and the flag makes arming a separate, reversible
-- decision a reviewer of THIS file can see. Withdrawal is setting it false —
-- seconds, no roll.
--
-- CONSUMERS, enumerated not asserted [MEASURED 2026-09-02]: exactly ONE live
-- step fleet-wide carries action=flag_page_image_rebuild —
-- image-build-handler's `flag_rebuild`. This file touches only it.
--
-- APPLY BY HAND, in this order:
--   1. Prove the capability, not the commit:
--        SELECT count(DISTINCT pod_name) FILTER (WHERE name='flag_page_image_rebuild') AS have_action
--          FROM service_binary_capabilities
--         WHERE service='agent-chassis' AND last_seen_at > now() - interval '10 minutes';
--      then grep one pod's binary for the new literal with present+absent controls:
--        kubectl -n ai-persona-system exec <pod> -- grep -acF "wire_page_hero_on_landing: wired the landed content hero" /proc/1/exe   (>=1)
--        kubectl -n ai-persona-system exec <pod> -- grep -acF "wire_page_hero_on_landing: XYZZY-not-real" /proc/1/exe                  (0)
--   2. Apply this file.
--   3. Watch the FIRST natural landing:
--        kubectl -n ai-persona-system logs -l app=agent-chassis --tail=5000 | grep 'hero_wire'
--      Every disposition is logged (wired:<n> and each skip). Then verify at
--      the ARTEFACT: the page's stored hero_url carries the content-hero path
--      AND the served page references it (never the item status).
--   ⚠ 4. finetuning.uk is EXCLUDED as acceptance evidence in either direction
--      (migrations 664/649 overlap two defects there — 412 lane, 2026-09-02).
--      Measure on IMG-077 `unwired` rollup sites once 708 is live.
--
-- Rollback: set the key false (or remove it). No hold needed on the rollback —
-- an unread/false key restores today's behaviour exactly.

BEGIN;

-- Idempotency FIRST, snapshot second (648's guardian-caught ordering).
DO $$
DECLARE done int;
BEGIN
    SELECT count(*) INTO done FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'image-build-handler'
       AND (default_config->'workflow'->'steps'->'flag_rebuild'->'config'->>'wire_hero_on_landing')::boolean IS TRUE;
    IF done > 0 THEN
        RAISE EXCEPTION '710: already applied — image-build-handler''s flag_rebuild already arms wire_hero_on_landing';
    END IF;
END $$;

SELECT snapshot_agent('image-build-handler', '710_arm_wire_hero_on_landing: pre-update');

-- COUNTED NEEDLE-GATE: exactly ONE active non-snapshot image-build-handler,
-- whose flag_rebuild step is flag_page_image_rebuild carrying the three
-- mappings this file's consumer census rests on.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'image-build-handler'
       AND default_config->'workflow'->'steps'->'flag_rebuild'->>'action' = 'flag_page_image_rebuild'
       AND default_config->'workflow'->'steps'->'flag_rebuild'->'config' ?& array['site_id','scope','scope_ref'];
    IF n <> 1 THEN
        RAISE EXCEPTION
          '710 needle-gate: expected exactly 1 image-build-handler whose flag_rebuild step is flag_page_image_rebuild with site_id/scope/scope_ref config, found % — re-derive against the live workflow', n;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,flag_rebuild,config,wire_hero_on_landing}',
         'true'::jsonb),
       updated_at = NOW()
 WHERE type = 'image-build-handler'
   AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;

-- VERIFY — DO/RAISE. The key-count arm is the append-vs-replace guard for an
-- OBJECT (648's array-length arm, adapted): 3 config keys before, 4 after.
DO $$
DECLARE ok int; n_keys int;
BEGIN
    SELECT count(*) INTO ok FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'image-build-handler'
       AND (default_config->'workflow'->'steps'->'flag_rebuild'->'config'->>'wire_hero_on_landing')::boolean IS TRUE
       AND default_config->'workflow'->'steps'->'flag_rebuild'->'config' ?& array['site_id','scope','scope_ref'];
    IF ok <> 1 THEN
        RAISE EXCEPTION '710 verify: the armed key and the three original mappings do not coexist on exactly 1 row (matched %)', ok;
    END IF;

    SELECT count(*) INTO n_keys FROM agent_definitions,
         LATERAL jsonb_object_keys(default_config->'workflow'->'steps'->'flag_rebuild'->'config') AS k
     WHERE type='image-build-handler'
       AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL;
    IF n_keys <> 4 THEN
        RAISE EXCEPTION '710 verify: flag_rebuild config has % key(s), expected exactly 4 (3 as of 2026-09-02 + this one) — jsonb_set landed somewhere unexpected, or the step moved under this file', n_keys;
    END IF;
END $$;

COMMIT;

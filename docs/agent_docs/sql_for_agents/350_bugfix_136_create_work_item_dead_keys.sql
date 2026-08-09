-- ============================================================================
-- 350_bugfix_136_create_work_item_dead_keys.sql
--
-- bugs_open/136 item D, the DATA half. Two dead keys removed from live
-- create_work_item steps so that opting the action into unknown-config-key
-- detection (CheckConfig: true) does not immediately warn about keys we have
-- already adjudicated. Data first, then the code — the handoff's ordering.
--
-- Both are dead in the same way and it is worth stating precisely, because
-- "grep found nothing" is NOT what establishes it (bug §6-landmines). What
-- establishes it is reading ExtractActionInputs end to end: every strategy
-- that resolves a config key iterates either spec.Required u spec.Optional
-- (Strategies 0, 2, 4 and the nested-object pass), or config["input_fields"]
-- (Strategy 1), or spec.Deprecated (Strategy 3). A key that is in none of
-- those three sets is resolved by nothing, whatever its value looks like.
--
--   1. grounded-explainer.create_review_item.spec_fields
--      ["draft","grounding_audit","registration","input_data"]. The action
--      builds its spec from spec_data / spec_paths / spec_literal (:208-236);
--      `spec_fields` is not among them and is not in the spec. This is why
--      both grounded_draft_review items have an EMPTY spec, which migration
--      343 already recorded as the reason the repaired captions cannot name
--      their topic. Removing it changes nothing; it stops the config claiming
--      four fields are captured when none is.
--
--   2. claims-auditor.request_claims_review.domain  "site_record.domain"
--      The action DOES call inputs.Get("domain") (:163) when building item_key
--      from item_key_prefix — which is why this key must not be convicted by
--      grep, and nearly was once already (bug §"LANDMINE"). Adjudicated
--      properly: `domain` is not in CreateWorkItemInputSpec's Required u
--      Optional, and this step sets no input_fields, so no strategy resolves
--      it; inputs.Get("domain") returns "" and item_key falls back to
--      siteID[:8] (:163-166). The step has also never executed —
--      `item_key LIKE 'claims_llm%'` is 0 rows, against 4861 rows carrying an
--      underscore item_key as the positive control — so nothing existing
--      depends on either key shape.
--
--      The AUTHOR'S INTENT (an item_key named for the domain) is reachable
--      today, spelled differently: `item_key_suffix_field` is read straight
--      from config and resolved with ExtractNestedFieldString against
--      collected data (:184-196), with no spec involvement at all. If a
--      domain-scoped key is wanted, that is the working mechanism. Declaring
--      `domain` in the spec instead would be a FLEET-WIDE change — the
--      nested-object pass (:544-568) would then resolve site_record.domain for
--      EVERY create_work_item step, silently re-shaping item_key from
--      <prefix>_<siteid8> to <prefix>_<domain> everywhere and breaking dedup
--      continuity with existing rows on idx_swi_dedup. That is not a
--      declaration, it is a behaviour change wearing one.
--
-- Seeds: 224_grounded_explainer_agent.sql is corrected in the same commit
-- (the bugs_open/134 lesson). claims-auditor has NO seed anywhere in the repo
-- — its definition exists only as a live row — so there is nothing to correct
-- and nothing that could replay the key back in. That is itself worth knowing.
--
-- NOT in this file, deliberately: the `spec` key carried by three other live
-- create_work_item steps (improvement-loop x2, deduplicate-sections x1),
-- found by the same enumeration and dead the same way. It has a live
-- consequence and its fix is a behaviour change, so it is being diagnosed
-- separately (090 run be967639-d195-444a-b9c3-ef1445ff7ae1) rather than
-- swept in here.
--
-- Live immediately (DB config; no image roll involved).
-- ============================================================================

BEGIN;

DO $$
DECLARE
  ge_has int;
  ca_has int;
BEGIN
  SELECT count(*) INTO ge_has
    FROM agent_definitions
   WHERE type='grounded-explainer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'create_review_item'->'config' ? 'spec_fields';

  SELECT count(*) INTO ca_has
    FROM agent_definitions
   WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config->'workflow'->'steps'->'request_claims_review'->'config' ? 'domain';

  IF ge_has = 0 AND ca_has = 0 THEN
    RAISE EXCEPTION '350: already applied (neither dead key present)';
  END IF;
  IF ge_has <> 1 OR ca_has <> 1 THEN
    RAISE EXCEPTION '350: DRIFT — expected 1 grounded-explainer spec_fields and 1 claims-auditor domain, found % / %. Re-measure before applying.',
      ge_has, ca_has;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,create_review_item,config,spec_fields}',
    updated_at = now()
WHERE type='grounded-explainer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'create_review_item'->'config' ? 'spec_fields';

UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,request_claims_review,config,domain}',
    updated_at = now()
WHERE type='claims-auditor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'request_claims_review'->'config' ? 'domain';

-- Verify: DO/RAISE, not bare SELECTs (ON_ERROR_STOP ignores a non-empty result).
-- Both the removal AND the survival of the sibling keys, so a whole-config
-- clobber cannot pass as a surgical delete.
DO $$
DECLARE
  ge_cfg jsonb;
  ca_cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'create_review_item'->'config' INTO ge_cfg
    FROM agent_definitions WHERE type='grounded-explainer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF ge_cfg ? 'spec_fields' THEN
    RAISE EXCEPTION '350 VERIFY: grounded-explainer still carries spec_fields';
  END IF;
  IF NOT (ge_cfg ? 'item_type' AND ge_cfg ? 'handler_agent' AND ge_cfg ? 'summary'
          AND ge_cfg ? 'site_id' AND ge_cfg ? 'status' AND ge_cfg ? 'priority'
          AND ge_cfg ? 'severity') THEN
    RAISE EXCEPTION '350 VERIFY: grounded-explainer lost a sibling key: %', ge_cfg;
  END IF;

  SELECT default_config->'workflow'->'steps'->'request_claims_review'->'config' INTO ca_cfg
    FROM agent_definitions WHERE type='claims-auditor' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF ca_cfg ? 'domain' THEN
    RAISE EXCEPTION '350 VERIFY: claims-auditor still carries domain';
  END IF;
  IF NOT (ca_cfg ? 'item_type' AND ca_cfg ? 'handler_agent' AND ca_cfg ? 'summary'
          AND ca_cfg ? 'site_id' AND ca_cfg ? 'status' AND ca_cfg ? 'severity'
          AND ca_cfg ? 'spec_data' AND ca_cfg ? 'item_pipeline'
          AND ca_cfg ? 'item_key_prefix') THEN
    RAISE EXCEPTION '350 VERIFY: claims-auditor lost a sibling key: %', ca_cfg;
  END IF;
END $$;

COMMIT;

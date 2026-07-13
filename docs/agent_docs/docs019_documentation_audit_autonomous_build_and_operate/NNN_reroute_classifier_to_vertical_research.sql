-- NNN_reroute_classifier_to_vertical_research.sql  (§B4 — lengthen the relay)
--
-- domain-research-classifier currently chains needs_strategy →
-- domain-strategist. After this, it chains needs_vertical_research →
-- vertical-exemplar-researcher, and the researcher chains needs_strategy
-- onward (its seed does that). Priority 7 = just below strategy's 8, matching
-- the ascending ladder (8/10/15) so later-stage items still win dispatch when
-- sites compete; within a site, order is enforced by creation anyway.
-- Adopted sites: NOT skipped (user decision) — fidelity rules keep outranking
-- in the classifier and downstream prompts.
--
-- SELF-GUARDING: the UPDATE matches only if the step is named
-- create_next_item AND currently creates needs_strategy (the pattern all
-- three on-disk relay siblings use). UPDATE 0 = assumption wrong, nothing
-- changed → run the inspection SELECT at the bottom and paste it.
-- Byte-exact revert path = the snapshot taken below (the REVERT block is a
-- faithful semantic reconstruction; the snapshot holds the true original).
-- ⚠ APPLY AFTER NNN_seed_vertical_exemplar_researcher.sql (the handler must
-- exist or feasibility-recheck will hold items blocked).

BEGIN;

SELECT snapshot_agent('domain-research-classifier',
  'create_next_item: needs_strategy/domain-strategist -> needs_vertical_research/vertical-exemplar-researcher (§B4 hop insertion)');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,create_next_item,config}',
      '{
        "source": "domain-research-classifier",
        "site_id": "input_data.site_id",
        "summary": "Vertical exemplar research needed after classification",
        "priority": 7,
        "severity": "high",
        "item_type": "needs_vertical_research",
        "item_domain": "build",
        "handler_agent": "vertical-exemplar-researcher",
        "item_key_prefix": "vertical_research"
      }'::jsonb),
    updated_at = now()
WHERE type = 'domain-research-classifier'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_next_item,config,item_type}' = 'needs_strategy';

-- verify — expect needs_vertical_research / vertical-exemplar-researcher
SELECT type,
       default_config #> '{workflow,steps,create_next_item,config}' AS chain_config
FROM agent_definitions
WHERE type = 'domain-research-classifier'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

COMMIT;

-- If UPDATE reported 0 rows: the step name/shape assumption was wrong —
-- inspect and paste:
-- SELECT jsonb_pretty(default_config->'workflow'->'steps')
-- FROM agent_definitions
-- WHERE type='domain-research-classifier'
--   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── REVERT (semantic; snapshot holds the byte-exact original) ───────────────
-- BEGIN;
-- SELECT snapshot_agent('domain-research-classifier','revert §B4 hop insertion');
-- UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--   '{workflow,steps,create_next_item,config}',
--   '{"source":"domain-research-classifier","site_id":"input_data.site_id","summary":"Strategy needed after domain research","priority":8,"severity":"high","item_type":"needs_strategy","item_domain":"build","handler_agent":"domain-strategist","item_key_prefix":"strategy"}'::jsonb),
--   updated_at = now()
-- WHERE type='domain-research-classifier' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- COMMIT;

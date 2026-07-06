-- NNN_strategist_vertical_landscape_nudge.sql  (§B4 — OPTIONAL)
--
-- NOT REQUIRED for the data path: domain-strategist's read_site_spec loads ALL
-- aspects and its prompt injects {{.site_specs}} wholesale, so the new
-- vertical_landscape aspect reaches the strategy LLM with zero changes. This
-- optional nudge tells the prompt to WEIGH it in the sections it answers
-- (Competitive Positioning; Page Type Recommendations). Idempotent (guarded
-- on the marker string). Snapshot first.

BEGIN;

SELECT snapshot_agent('domain-strategist',
  'analyze_strategy prompt: append vertical_landscape weighting nudge (§B4, optional)');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,analyze_strategy,config,prompt_template}',
      to_jsonb(
        (default_config #>> '{workflow,steps,analyze_strategy,config,prompt_template}')
        || E'\n\n## Vertical Landscape\nIf the Research Data includes a `vertical_landscape` aspect (best-of-niche exemplar research), weigh it heavily in section 4 (Competitive Positioning) and section 6 (Page Type Recommendations): build on the success factors and lessons it records — reasons, not copies — and exploit the differentiation_opportunity it names.'
      )),
    updated_at = now()
WHERE type = 'domain-strategist'
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,analyze_strategy,config,prompt_template}' NOT LIKE '%vertical_landscape%';

-- verify — expect the marker present exactly once
SELECT type,
       (default_config #>> '{workflow,steps,analyze_strategy,config,prompt_template}') LIKE '%## Vertical Landscape%' AS nudged
FROM agent_definitions
WHERE type = 'domain-strategist'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

COMMIT;

-- ── REVERT: restore from the snapshot taken above (byte-exact) ──────────────

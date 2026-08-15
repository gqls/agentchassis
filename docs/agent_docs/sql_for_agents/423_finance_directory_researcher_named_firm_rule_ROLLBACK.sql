-- 423 ROLLBACK — remove the named-firm rule from finance-directory-researcher's
-- extract_claims prompt (reverse of 423_finance_directory_researcher_named_firm_rule.sql).
--
-- Reverse replace() of the exact inserted text, guarded on its presence; a second run or
-- a never-applied 423 is a 0-row no-op. Alternative restore: the agent_definitions_backup
-- row 423 took (snapshot_reason '423_finance_directory_researcher_named_firm_rule:
-- pre-update'; order by snapshot_taken_at, and check it holds the PRE-change text).
--
-- Verify after applying (expect f | t on exactly one row):
--   SELECT (default_config::text LIKE '%THE SECOND HARD RULE%') AS rule_still_present,
--          (default_config::text LIKE '%If a source''s best content is pricing, skip it.%') AS anchor_intact
--   FROM agent_definitions WHERE type='finance-directory-researcher' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('finance-directory-researcher',
  '423_ROLLBACK: pre-removal');

UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'If a source''s best content is pricing, skip it.\n\nTHE SECOND HARD RULE — AN ENTITY IS ONE NAMED FIRM. Every entity_slug / entity_name must be a single company or brand that holds (or could hold) its own FCA authorisation. Never register a sector, market segment, product category, regulator, statistics series or any other aggregate as an entity — shapes like ''UK specialist lenders'', ''FCA-regulated mortgage lenders (general)'' or ''the later life mortgage market'' are wrong even when the facts about them are true and citable, and a human reviewer will reject the set for them. A market study, trade-body overview or statistics page may only yield claims about the individual firms it NAMES; if a source discusses the market solely in aggregate, extract nothing from it.',
      'If a source''s best content is pricing, skip it.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%THE SECOND HARD RULE%';

SELECT
  (default_config::text LIKE '%THE SECOND HARD RULE%') AS rule_still_present,
  (default_config::text LIKE '%If a source''s best content is pricing, skip it.%') AS anchor_intact
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- 424 ROLLBACK — remove the quote-shape rule and the ibisworld.com exclusion
-- (reverse of 424_finance_directory_researcher_quote_shape_and_ibisworld.sql).
-- Guarded on presence; a second run or a never-applied 424 is a 0-row no-op.
-- Alternative restore: the agent_definitions_backup row 424 took (snapshot_reason
-- '424_…: pre-update'; order by snapshot_taken_at, check it holds the PRE-change text).
--
-- Verify after applying (expect f | f on exactly one row):
--   SELECT (default_config::text LIKE '%ONE CONTINUOUS passage%') AS quote_rule_present,
--          (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}')
--            ? 'ibisworld.com' AS exclusion_present
--   FROM agent_definitions WHERE type='finance-directory-researcher' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('finance-directory-researcher',
  '424_ROLLBACK: pre-removal');

UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'do not paraphrase, do not stitch sentences, do not normalise numbers or names; quote ONE CONTINUOUS passage of running text — never a bullet list, never text spanning headings or paragraph breaks (layout line breaks and list markers do not survive the verification refetch, so such a quote fails even when every word is on the page);',
      'do not paraphrase, do not stitch sentences, do not normalise numbers or names;'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%ONE CONTINUOUS passage%';

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,prepare_urls,config,exclude_domains}',
      (SELECT COALESCE(jsonb_agg(e), '[]'::jsonb)
         FROM jsonb_array_elements(default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') e
        WHERE e <> to_jsonb('ibisworld.com'::text))
    ),
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'ibisworld.com';

SELECT
  (default_config::text LIKE '%ONE CONTINUOUS passage%') AS quote_rule_present,
  (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'ibisworld.com'
    AS exclusion_present
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

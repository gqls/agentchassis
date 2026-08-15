-- ROLLBACK for 428 — finance-directory-researcher: value-hygiene rules + forbes.com
-- Reverses the four edits by exact string replace / array element removal.
-- Alternative: restore from the agent_definitions_backup snapshot
-- '428_finance_directory_researcher_value_hygiene_and_forbes: pre-update'
-- (verify it holds the PRE-change text before using it).

BEGIN;

SELECT snapshot_agent('finance-directory-researcher',
  '428 ROLLBACK: pre-revert');

-- Reverse Edit A.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'If a source''s best content is pricing, skip it. This extends INSIDE values: no monetary amount may appear anywhere in a value or quote-selected fact (a benefit limit such as ''outpatient cover up to GBP 350 included'' is a price-shaped figure and no allowed field may carry it) — state the cover or product category without the amount, or skip the fact.',
      'If a source''s best content is pricing, skip it.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%no monetary amount may appear anywhere%';

-- Reverse Edit B.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'one per (provider, fact) pair. At most ONE claim per (provider, field) pair — if several passages state the same field, emit only the single most complete enumeration: a later duplicate OVERWRITES the earlier one at registration, so a weaker duplicate destroys a stronger claim.',
      'one per (provider, fact) pair.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('At most ONE claim per (provider, field) pair' in default_config::text) > 0;

-- Reverse Edit C.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'never your synthesis across pages. underwriter is the underwriting FIRM''S NAME alone (e.g. ''AXA PPP Healthcare Limited''), never a description of ownership, structure or ethos.',
      'never your synthesis across pages.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('underwriting FIRM''S NAME alone' in default_config::text) > 0;

-- Reverse Edit D.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,prepare_urls,config,exclude_domains}',
      (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') - 'forbes.com'
    ),
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'forbes.com';

-- Verify the revert (expect f | f | f | f). DO/RAISE so a miss aborts the COMMIT.
DO $$
DECLARE r record;
BEGIN
  SELECT
    (default_config::text LIKE '%no monetary amount may appear anywhere%') AS a,
    position('At most ONE claim per (provider, field) pair' in default_config::text) > 0 AS b,
    position('underwriting FIRM''S NAME alone' in default_config::text) > 0 AS c,
    (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'forbes.com' AS d
  INTO r
  FROM agent_definitions
  WHERE type = 'finance-directory-researcher' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF r.a OR r.b OR r.c OR r.d THEN
    RAISE EXCEPTION '428 rollback verify failed: a=% b=% c=% d=%', r.a, r.b, r.c, r.d;
  END IF;
END $$;

COMMIT;

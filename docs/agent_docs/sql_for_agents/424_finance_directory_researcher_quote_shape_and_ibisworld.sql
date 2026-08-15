-- 424 — finance-directory-researcher: quote-shape rule + ibisworld.com exclusion
--
-- WHY. B4 run 4 (orchestration 26c4f5ac, 2026-08-15 14:48Z, mortgage-lender): the
-- named-firm rule (423) worked — candidates were real firms (Nationwide, Yorkshire BS,
-- Coventry BS, Family BS) — but ALL 8 candidates failed verification, two mechanical ways:
--   1. citation_lost x3: the extractor quoted multi-paragraph BULLET BLOCKS from Family
--      Building Society's page ("...\n\n- Owner Occupier Repayment mortgages – ..."). The
--      verifier refetches and requires the quote verbatim in the page's visible text;
--      layout-derived line breaks and list markers do not survive the refetch's text
--      extraction, so a bullet-block quote fails even though every word is on the page.
--      (Run 1's shorter single-sentence quotes verified fine — the failure is shape, not
--      truth.)
--   2. fetch_error HTTP 405 x4: every claim cited to ibisworld.com — the industry-stats
--      aggregator refuses the verifier's refetch, so facts cited there can NEVER verify.
--      One scrape slot of 4 per run spent on a source whose citations are all dead on
--      arrival.
--
-- EDITS (both this agent's own row; replace()-idiom, snapshot-first, per 206/423):
--   A. extract_claims prompt: extend the verbatim-quote bullet with a continuous-passage
--      requirement.
--   B. prepare_urls exclude_domains: append "ibisworld.com" (evidence names only this
--      domain; the class — refetch-blocked aggregators — is noted in
--      portfolio_positioning NOTES for when a second one appears).
--
-- Config-only: no image dependency, live on apply. Weekly task; force-trigger to exercise.
--
-- ROLLBACK: 424_finance_directory_researcher_quote_shape_and_ibisworld_ROLLBACK.sql, or
-- the agent_definitions_backup row taken below (order by snapshot_taken_at; check it
-- holds the PRE-change text).
--
-- Verify after applying (expect t | t on exactly one row):
--   SELECT (default_config::text LIKE '%ONE CONTINUOUS passage%') AS has_quote_rule,
--          (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}')
--            ? 'ibisworld.com' AS has_exclusion
--   FROM agent_definitions WHERE type='finance-directory-researcher' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('finance-directory-researcher',
  '424_finance_directory_researcher_quote_shape_and_ibisworld: pre-update');

-- Pre-check (expect 1 | f | f): anchor unique, neither edit present yet.
SELECT
  (length(default_config::text) - length(replace(default_config::text,
    'do not paraphrase, do not stitch sentences, do not normalise numbers or names;', '')))
  / length('do not paraphrase, do not stitch sentences, do not normalise numbers or names;')
    AS anchor_count,
  (default_config::text LIKE '%ONE CONTINUOUS passage%') AS quote_rule_present,
  (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'ibisworld.com'
    AS exclusion_present
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Edit A: quote-shape rule, spliced into the verbatim-quote bullet.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'do not paraphrase, do not stitch sentences, do not normalise numbers or names;',
      'do not paraphrase, do not stitch sentences, do not normalise numbers or names; quote ONE CONTINUOUS passage of running text — never a bullet list, never text spanning headings or paragraph breaks (layout line breaks and list markers do not survive the verification refetch, so such a quote fails even when every word is on the page);'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%do not paraphrase, do not stitch sentences, do not normalise numbers or names;%'
  AND default_config::text NOT LIKE '%ONE CONTINUOUS passage%';

-- Edit B: exclude ibisworld.com from scraping (refetch-blocked; citations cannot verify).
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,prepare_urls,config,exclude_domains}',
      (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}')
        || '["ibisworld.com"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'ibisworld.com';

-- Verify (expect t | t on exactly one row).
SELECT
  (default_config::text LIKE '%ONE CONTINUOUS passage%') AS has_quote_rule,
  (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'ibisworld.com'
    AS has_exclusion
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

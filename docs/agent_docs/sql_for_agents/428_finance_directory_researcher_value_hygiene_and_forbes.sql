-- 428 — finance-directory-researcher: value-hygiene rules + forbes.com exclusion
--
-- WHY. B4 run 7 (orchestration 8b6f8e12, 2026-08-15 17:29Z, health-insurer — the kind's
-- first ever run): 13/15 candidates registered, all real named firms, but the supervised
-- review of the registered SET found four defects the existing rules do not cover:
--   1. A monetary amount rode INSIDE an allowed field's value: WPA cover_types =
--      "inpatient, outpatient (£350 included as standard); ...". The registration-time
--      non-price gate blocks price FIELDS, not price CONTENT in allowed fields; a benefit
--      limit is exactly the volatile figure the owner's non-price ruling exists to keep
--      out. (Go-side residual: the mechanical gate could scan value content too — noted
--      in portfolio_positioning NOTES as a fix candidate, needs a roll.)
--   2. Same-run duplicate DESTROYED the better claim: two bupa.cover_types extractions
--      from one page; registration is last-write-wins on (entity, field), and the
--      surviving value was the weaker benefits blurb ("24/7 remote GP access and
--      dental..."), the superseded one the real enumeration ("inpatient, outpatient,
--      mental health...").
--   3. Marketing prose in underwriter: bupa.underwriter = "Bupa (no shareholders;
--      reinvests profits into services)" — ownership/ethos description, not a firm name.
--   4. fetch_error HTTP 403 x2 on forbes.com/advisor — a refetch-blocked aggregator, the
--      exact class 424 named at ibisworld.com with the recorded policy "add members as
--      runs name them". Run 7 named forbes.com.
--
-- EDITS (all this agent's own row; replace()-idiom, snapshot-first, per 206/423/424):
--   A. extract_claims prompt_template: extend the no-prices paragraph — no monetary
--      amount anywhere inside a value or quote-selected fact.
--   B. extract_claims prompt_template: at most one claim per (provider, field) pair;
--      strongest enumeration wins, because a later duplicate overwrites the earlier.
--   C. extract_claims prompt_template: underwriter = the underwriting firm's name alone.
--   D. prepare_urls exclude_domains: append "forbes.com".
--
-- Config-only: no image dependency, live on apply. Weekly task; force-trigger to exercise.
--
-- ROLLBACK: 428_finance_directory_researcher_value_hygiene_and_forbes_ROLLBACK.sql, or
-- the agent_definitions_backup row taken below (order by snapshot_taken_at; check it
-- holds the PRE-change text).
--
-- Verify after applying (expect t | t | t | t on exactly one row):
--   SELECT (default_config::text LIKE '%no monetary amount may appear anywhere%') AS a,
--          (default_config::text LIKE '%At most ONE claim per (provider, field) pair%') AS b,
--          (default_config::text LIKE '%underwriting FIRM''''S NAME alone%') AS c,
--          (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}')
--            ? 'forbes.com' AS d
--   FROM agent_definitions WHERE type='finance-directory-researcher' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

SELECT snapshot_agent('finance-directory-researcher',
  '428_finance_directory_researcher_value_hygiene_and_forbes: pre-update');

-- Pre-check (expect 1 | 1 | 1 | f | f | f | f): anchors unique, no edit present yet.
SELECT
  (length(default_config::text) - length(replace(default_config::text,
    'If a source''s best content is pricing, skip it.', '')))
  / length('If a source''s best content is pricing, skip it.') AS anchor_a,
  (length(default_config::text) - length(replace(default_config::text,
    'one per (provider, fact) pair.', '')))
  / length('one per (provider, fact) pair.') AS anchor_b,
  (length(default_config::text) - length(replace(default_config::text,
    'never your synthesis across pages.', '')))
  / length('never your synthesis across pages.') AS anchor_c,
  (default_config::text LIKE '%no monetary amount may appear anywhere%') AS a_present,
  position('At most ONE claim per (provider, field) pair' in default_config::text) > 0 AS b_present,
  position('underwriting FIRM''S NAME alone' in default_config::text) > 0 AS c_present,
  (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'forbes.com' AS d_present
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Edit A: no monetary amounts inside values, spliced onto the no-prices paragraph.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'If a source''s best content is pricing, skip it.',
      'If a source''s best content is pricing, skip it. This extends INSIDE values: no monetary amount may appear anywhere in a value or quote-selected fact (a benefit limit such as ''outpatient cover up to GBP 350 included'' is a price-shaped figure and no allowed field may carry it) — state the cover or product category without the amount, or skip the fact.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('If a source''s best content is pricing, skip it.' in default_config::text) > 0
  AND NOT (default_config::text LIKE '%no monetary amount may appear anywhere%');

-- Edit B: one claim per (provider, field); strongest enumeration wins.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'one per (provider, fact) pair.',
      'one per (provider, fact) pair. At most ONE claim per (provider, field) pair — if several passages state the same field, emit only the single most complete enumeration: a later duplicate OVERWRITES the earlier one at registration, so a weaker duplicate destroys a stronger claim.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('one per (provider, fact) pair.' in default_config::text) > 0
  AND position('At most ONE claim per (provider, field) pair' in default_config::text) = 0;

-- Edit C: underwriter is a firm name, not a description.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'never your synthesis across pages.',
      'never your synthesis across pages. underwriter is the underwriting FIRM''S NAME alone (e.g. ''AXA PPP Healthcare Limited''), never a description of ownership, structure or ethos.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND position('never your synthesis across pages.' in default_config::text) > 0
  AND position('underwriting FIRM''S NAME alone' in default_config::text) = 0;

-- Edit D: exclude forbes.com from scraping (refetch-blocked; citations cannot verify).
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,prepare_urls,config,exclude_domains}',
      (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}')
        || '["forbes.com"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config#>'{workflow,steps,prepare_urls,config,exclude_domains}') ? 'forbes.com';

-- Verify (expect t | t | t | t on exactly one row). DO/RAISE so a miss aborts the COMMIT.
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
  IF NOT (r.a AND r.b AND r.c AND r.d) THEN
    RAISE EXCEPTION '428 verify failed: a=% b=% c=% d=%', r.a, r.b, r.c, r.d;
  END IF;
END $$;

COMMIT;

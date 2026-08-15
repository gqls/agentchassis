-- 423 — finance-directory-researcher: an entity is ONE NAMED FIRM (extract_claims rule)
--
-- WHY. B4's first supervised run (orchestration c516508b-a1ae-43ad-8f39-63aa07f48b8f,
-- 2026-08-15 14:32Z, kind=mortgage-lender) registered 3 entities of which 2 were
-- CATEGORY-shaped, not firms: 'FCA-regulated mortgage lenders (general)' and
-- 'UK Specialist Lending Sector (nonbank)'. The rejected candidate was category-shaped
-- too ('later-life-mortgage-lenders-uk'). Cause: the scraped set was mostly market-level
-- pages (FCA guidance, KBRA RMBS research, a law firm's market-study commentary, BoE
-- statistics) and the extract_claims prompt says "name the specific provider" but never
-- forbids hanging true, citable facts on a sector/aggregate pseudo-entity. The claims all
-- verified — the citation gate cannot catch this, because the defect is entity SHAPE, not
-- claim truth. A directory listing a sector as a lender fails the owner's bar (a reviewed,
-- non-embarrassing set per kind).
--
-- The sibling directory-researcher (model directory, another lane) has the same latent
-- gap but model sources are model-specific, so it has not fired there. Not edited here —
-- their lane's row; noted in portfolio_positioning NOTES.
--
-- Mechanics: one quote-free edit via replace() on default_config::text, cast back to
-- ::jsonb (the cast validates). \n in the replacement is the two-character JSON escape,
-- correct inside the jsonb text representation (206's idiom). Anchor is the closing
-- sentence of the NO-PRICES paragraph, verified unique (count=1) by the pre-check; the
-- UPDATE's WHERE re-asserts the anchor AND the new rule's absence, so prompt drift or a
-- second apply is a 0-row no-op, never a mis-splice. If the pre-check prints anything but
-- 1|f, STOP and re-derive the anchor from the live row.
--
-- Config-only: no image dependency, live on apply. Weekly task; force-trigger to exercise.
--
-- ROLLBACK: 423_finance_directory_researcher_named_firm_rule_ROLLBACK.sql (reverse
-- replace), or restore the agent_definitions_backup row this file takes (two-arg
-- snapshot_agent writes there, NOT an is_snapshot row; order by snapshot_taken_at and
-- check it holds the PRE-change text — see LANDMINES on both).
--
-- Verify after applying (expect 1 row: has_rule=t, anchor_kept=t, no_doubled_quote=t).
-- NB position(), not LIKE, for the \n check: backslash is LIKE's escape character, so a
-- \n in a LIKE pattern silently matches a literal 'n' instead of the JSON escape.
--   SELECT (default_config::text LIKE '%THE SECOND HARD RULE%') AS has_rule,
--          (position('skip it.\n\nTHE SECOND HARD RULE' in default_config::text) > 0) AS anchor_kept,
--          (default_config::text NOT LIKE '%''''UK specialist lenders%') AS no_doubled_quote
--   FROM agent_definitions WHERE type='finance-directory-researcher' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

BEGIN;

-- SNAPSHOT (rollback safety): two-arg form -> agent_definitions_backup.
SELECT snapshot_agent('finance-directory-researcher',
  '423_finance_directory_researcher_named_firm_rule: pre-update');

-- Pre-check: anchor present exactly once, rule not yet present (expect 1 | f).
SELECT
  (length(default_config::text) - length(replace(default_config::text,
    'If a source''s best content is pricing, skip it.', '')))
  / length('If a source''s best content is pricing, skip it.')
    AS anchor_count,
  (default_config::text LIKE '%THE SECOND HARD RULE%') AS rule_already_present
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Apply: append the named-firm rule as its own paragraph after the price rule.
UPDATE agent_definitions
SET default_config = replace(
      default_config::text,
      'If a source''s best content is pricing, skip it.',
      'If a source''s best content is pricing, skip it.\n\nTHE SECOND HARD RULE — AN ENTITY IS ONE NAMED FIRM. Every entity_slug / entity_name must be a single company or brand that holds (or could hold) its own FCA authorisation. Never register a sector, market segment, product category, regulator, statistics series or any other aggregate as an entity — shapes like ''UK specialist lenders'', ''FCA-regulated mortgage lenders (general)'' or ''the later life mortgage market'' are wrong even when the facts about them are true and citable, and a human reviewer will reject the set for them. A market study, trade-body overview or statistics page may only yield claims about the individual firms it NAMES; if a source discusses the market solely in aggregate, extract nothing from it.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%If a source''s best content is pricing, skip it.%'
  AND default_config::text NOT LIKE '%THE SECOND HARD RULE%';

-- Verify (expect t | t | t on exactly one row). position() not LIKE for the \n check:
-- backslash is LIKE's escape char, so '\n' in a pattern matches a literal 'n'.
SELECT
  (default_config::text LIKE '%THE SECOND HARD RULE%') AS has_rule,
  (position('skip it.\n\nTHE SECOND HARD RULE' in default_config::text) > 0) AS anchor_kept,
  (default_config::text NOT LIKE '%''''UK specialist lenders%') AS no_doubled_quote
FROM agent_definitions
WHERE type = 'finance-directory-researcher' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

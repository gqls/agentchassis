-- 206_planner_news_index_page_type.sql
-- ----------------------------------------------------------------------------
-- bugs_open/015 candidate 2 (LLM half): teach build-site-planner the
-- news-index page_type, so a news listing is emitted on the routing key the
-- news machinery selects on, instead of blog-index (which the canonical-types
-- table currently recommends for "news feed") or section-index (which the Go
-- role validator used to flatten everything to).
--
-- *** DO NOT APPLY until a chassis image containing the Go half is LIVE. ***
-- The Go half (same commit as this file) adds news-index to
-- page_role_validator.go / page_canonical.go. On the OLD binary,
-- ValidateRoles rules 2-4 rewrite an explicit news-index to section-index
-- (a -index page name, a declared parent, or a /<slug>/index.html URL each
-- trigger it), so applying this prompt first just recreates bug 015 with
-- better-looking input. Check the running pod first:
--   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c isTypedIndexRole'
--   (>0 means the new binary; also grep a known-old symbol as positive control)
--
-- Mechanics: two quote-free edits via replace() on default_config::text,
-- cast back to ::jsonb (the cast validates; malformed JSON aborts the txn).
-- \n in the replacement strings is the two-character JSON escape, correct
-- inside the jsonb text representation. Anchors verified unique (count=1
-- each) against the live row on 2026-07-24; the UPDATE's WHERE re-asserts
-- both anchors so prompt drift makes this a 0-row no-op instead of a
-- mis-splice — if the pre-check or update reports 0/false, STOP and
-- re-derive the anchors from the live config.
-- ----------------------------------------------------------------------------

BEGIN;

-- SNAPSHOT (rollback safety): full backup of the agent row before the edit.
CREATE TABLE IF NOT EXISTS agent_definitions_bak_015_planner AS
SELECT * FROM agent_definitions WHERE type = 'build-site-planner';

-- Pre-check: each anchor must be present exactly once (1|1).
SELECT
  (length(default_config::text) - length(replace(default_config::text,
    '| blog-index | Blog/news listing page | Article index, news feed |', '')))
  / length('| blog-index | Blog/news listing page | Article index, news feed |')
    AS blogindex_row_anchor_count,
  (length(default_config::text) - length(replace(default_config::text,
    'Do not invent new page_type values.', '')))
  / length('Do not invent new page_type values.')
    AS default_rule_anchor_count
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Apply. Edit 1 splits the blog-index table row into blog-only + a new
-- news-index row. Edit 2 appends the when-to-use rule after the table's
-- closing default-rule line.
UPDATE agent_definitions
SET default_config = replace(
      replace(
        default_config::text,
        '| blog-index | Blog/news listing page | Article index, news feed |',
        '| blog-index | Blog listing page | Article/blog index (NOT the news feed listing) |\n| news-index | Dedicated news listing page | The separate news page when the news feed recommends separate_page=true |'
      ),
      'Do not invent new page_type values.',
      'Do not invent new page_type values.\n\nNews listing rule: when the classification content_features.news_feed has recommended=true AND separate_page=true, plan exactly ONE news listing page with page_type news-index. Its name may be localised (noticias, nachrichten, news) but the page_type must be exactly news-index — never blog-index or section-index. page_type is a routing key: the news machinery (archive data, build gates, discovery checks) selects on news-index, and any other value orphans the page from all of it.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config::text LIKE '%| blog-index | Blog/news listing page | Article index, news feed |%'
  AND default_config::text LIKE '%Do not invent new page_type values.%';

-- Verify (all three = t).
SELECT
  (default_config::text LIKE '%| news-index | Dedicated news listing page |%') AS has_news_index_row,
  (default_config::text LIKE '%News listing rule:%') AS has_news_rule,
  (default_config::text NOT LIKE '%| blog-index | Blog/news listing page |%') AS old_row_gone
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed): restore the whole config from the snapshot.
--   UPDATE agent_definitions a
--   SET default_config = b.default_config, updated_at = NOW()
--   FROM agent_definitions_bak_015_planner b
--   WHERE a.type = b.type AND a.id = b.id;
-- Drop the snapshot once satisfied: DROP TABLE agent_definitions_bak_015_planner;
-- ----------------------------------------------------------------------------

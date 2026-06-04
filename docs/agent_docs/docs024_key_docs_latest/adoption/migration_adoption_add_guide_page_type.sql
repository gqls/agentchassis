-- migration_adoption_add_guide_page_type.sql
-- ----------------------------------------------------------------------------
-- GOING-FORWARD classifier change (does NOT touch existing rows).
--
-- Teach the site-adoption-agent `analyze_site` step about a first-class `guide`
-- page_type, so future adoptions type instructional/how-to pages as `guide`
-- (which guide-list resolves) instead of folding them into blog-post.
--
-- Two surgical text edits inside default_config, both QUOTE-FREE and
-- NEWLINE-FREE so they are safe to apply via replace() on default_config::text
-- and cast back to jsonb (the ::jsonb cast validates the JSON; a malformed
-- result aborts the txn):
--   1) add `guide` to the page_type enum the LLM chooses from
--   2) narrow the blog-post guidance and route instructional content to `guide`
--
-- Anchors verified unique within the live default_config (the pre-check below
-- returns the match counts so you can confirm each is exactly 1 before trusting
-- the result; replace() changes ALL occurrences).
--
-- NOTE: this does NOT re-type existing pages (see
-- migration_retype_guides_to_guide.sql) and does NOT change the
-- build-site-planner vocabulary. The build-site-planner preserves existing
-- pages' page_type verbatim, so adopted `guide` pages survive it; only NEW
-- guide planning by that agent would need the same addition.
--
-- CAVEAT I could not verify from source: apply_adoption_plan (writes the pages)
-- and validate_site_plan are not on disk. The DB CHECK accepts any kebab value,
-- and `game` already flows through these paths fine, but confirm a future
-- adoption actually persists page_type='guide' (watch one adoption run).
-- ----------------------------------------------------------------------------

BEGIN;

-- SNAPSHOT (rollback safety): full backup of the agent row before the edit.
CREATE TABLE IF NOT EXISTS agent_definitions_bak_adopt_guide AS
SELECT * FROM agent_definitions WHERE type = 'site-adoption-agent';

-- Pre-check: each anchor should be present exactly once (both columns = 1).
-- If either is not 1, STOP — the live prompt has drifted from these anchors.
SELECT
  (length(default_config::text)
     - length(replace(default_config::text,
        'content|tool|game|section-index|blog-index|blog-post|landing', '')))
   / length('content|tool|game|section-index|blog-index|blog-post|landing')
     AS enum_anchor_count,
  (length(default_config::text)
     - length(replace(default_config::text,
        'individual article, guide, or essay page with primarily prose content.', '')))
   / length('individual article, guide, or essay page with primarily prose content.')
     AS blogpost_anchor_count
FROM agent_definitions
WHERE type = 'site-adoption-agent';

-- Apply the two replacements.
UPDATE agent_definitions
SET default_config = replace(
      replace(
        default_config::text,
        'content|tool|game|section-index|blog-index|blog-post|landing',
        'content|tool|game|guide|section-index|blog-index|blog-post|landing'
      ),
      'individual article, guide, or essay page with primarily prose content.',
      'individual news item, opinion piece, or time-bound article with primarily prose content. Use the guide page_type (not blog-post) for instructional how-to, tutorial, or explainer pages that teach the reader to do or understand something; one page per guide.'
    )::jsonb,
    updated_at = NOW()
WHERE type = 'site-adoption-agent';

-- Verify: enum now contains guide and the guidance mentions it (both = t).
SELECT
  (default_config::text LIKE '%content|tool|game|guide|section-index|blog-index|blog-post|landing%') AS enum_has_guide,
  (default_config::text LIKE '%Use the guide page_type (not blog-post)%') AS guidance_has_guide
FROM agent_definitions
WHERE type = 'site-adoption-agent';

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed): restore the whole config from the snapshot.
--   UPDATE agent_definitions a
--   SET default_config = b.default_config, updated_at = NOW()
--   FROM agent_definitions_bak_adopt_guide b WHERE a.type = b.type;
-- Drop the snapshot once satisfied: DROP TABLE agent_definitions_bak_adopt_guide;
-- ----------------------------------------------------------------------------

-- OPTIONAL coherence tweaks (NOT applied here — lower value, and they touch
-- example URLs the LLM only uses as hints):
--   * section-index guidance: "...child tool or game pages..."  ->  add "or guide"
--     and add /guides/index.html to its example list.
--   * blog-index example "(e.g. /guides/index.html)" -> point at /blog/ or /news/
--     since /guides/index.html is a guides section-index, not a blog index.
-- Apply later as further quote-free replace() edits if desired.

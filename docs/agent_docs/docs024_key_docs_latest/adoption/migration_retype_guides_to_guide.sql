-- migration_retype_guides_to_guide.sql
-- ----------------------------------------------------------------------------
-- LIVE-SITE FIX for gamesdesign.co.uk.
--
-- Re-type the 5 content-bearing guide-* pages from blog-post -> guide so the
-- guide-list component (items.source = query.pages_where_type:guide) resolves
-- them. This mirrors the working game-list / page_type=game precedent: `game`
-- pages resolve via pages_where_type:game with no resolver change, so guide
-- pages will resolve the same way once typed `guide`.
--
-- SCOPE — only the 5 guide-* pages (they each have a content section):
--   guide-economy-basics, guide-fairness-in-rng, guide-p2p-architecture,
--   guide-rng-design, guide-skinner-box
-- The 5 EMPTY bare duplicates (economy-basics, fairness-in-rng,
-- p2p-architecture, rng-design, skinner-box; sections = []) are deliberately
-- LEFT as blog-post. They are empty shells and a SEPARATE defect — re-typing
-- them would surface 5 broken/empty cards in the guide list.
--
-- SCHEMA NOTE (checked): pages.page_type is varchar(50) with CHECK
-- chk_page_type_kebab_case — kebab-case FORMAT only, NO value allowlist.
-- 'guide' is valid format, so this UPDATE does not violate the constraint.
--
-- Data-only change. No code deploy. Effective on COMMIT.
-- ----------------------------------------------------------------------------

BEGIN;

-- SNAPSHOT (rollback safety): full backup of the rows being changed, taken
-- inside the txn before the UPDATE so it captures the pre-change page_type.
CREATE TABLE IF NOT EXISTS pages_bak_retype_guides AS
SELECT * FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND name LIKE 'guide-%'
  AND page_type = 'blog-post';

-- Before: the rows we will re-type (expect exactly 5).
SELECT name, page_type, url, sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND name LIKE 'guide-%'
  AND page_type = 'blog-post'
ORDER BY name;

-- Re-type the content-bearing guide pages.
UPDATE pages
SET page_type = 'guide'
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND name LIKE 'guide-%'
  AND page_type = 'blog-post';

-- After: confirm the 5 guide-* rows are now 'guide'; the bare duplicates remain
-- blog-post (they are NOT in this result because they don't match name LIKE).
SELECT name, page_type, url, sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND (name LIKE 'guide-%' OR page_type = 'guide')
ORDER BY name;

COMMIT;

-- ----------------------------------------------------------------------------
-- ROLLBACK (if needed): restore page_type from the snapshot.
--   UPDATE pages p SET page_type = b.page_type
--   FROM pages_bak_retype_guides b WHERE p.id = b.id;
-- Drop the snapshot once satisfied: DROP TABLE pages_bak_retype_guides;
-- ----------------------------------------------------------------------------

-- ============================================================================
-- RUN SEPARATELY, AFTER THE ABOVE COMMITS.
-- Re-render the pages that SHOW a guide-list so plan_sections re-resolves the
-- list against the now-guide-typed pages. `sections` is the page's jsonb array
-- of section functions.
--
-- IMPORTANT — coverage note: guide-list currently appears on the HOMEPAGE
-- (index). The guides hub `guides-index` has sections = [] (no guide-list
-- section at all), so re-typing alone will NOT populate the guides hub — that
-- hub needs a guide-list section added to its plan (separate work, same shape
-- as how the tools hub got its tool-list). The flip below rebuilds whatever
-- pages actually contain guide-list today.
--
-- 1) Verify which pages this targets:
--      SELECT name, url, build_status FROM pages
--      WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--        AND sections @> '["guide-list"]'::jsonb;
--
-- 2) SNAPSHOT before the flip (rollback safety):
--      CREATE TABLE IF NOT EXISTS pages_bak_guidelist_rebuild AS
--      SELECT id, name, url, build_status, built_from_plan_version, NOW() AS snapshot_at
--      FROM pages
--      WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--        AND sections @> '["guide-list"]'::jsonb;
--
-- 3) Flip to rebuild:
--      UPDATE pages
--      SET build_status = 'needs_rebuild', built_from_plan_version = NULL
--      WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--        AND sections @> '["guide-list"]'::jsonb;
--
-- Rollback the flip if needed:
--      UPDATE pages p
--      SET build_status = b.build_status, built_from_plan_version = b.built_from_plan_version
--      FROM pages_bak_guidelist_rebuild b WHERE p.id = b.id;
-- ============================================================================

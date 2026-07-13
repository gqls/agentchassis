-- B4/B5 — resolve "Browse All X" hub links from real pages via the
-- section_index_for queryresolve verb, replacing the unpopulated/inconsistent
-- *_index_url specs.
--
-- Ships with: queryresolve/section_index_for.go (new verb) + the switch case in
-- queryresolve.Resolve. Apply the Go + roll the image BEFORE this SQL, else the
-- query.section_index_for source resolves to "unknown query" and the field
-- falls through to unresolved.
--
-- Why source-only changes: for query.* fields the plan_sections field loop
-- assigns any non-nil result and never consults on_missing (it continues before
-- that switch); a `fallback` would be applied on nil (line ~1074), which we do
-- NOT want — an absent hub should render no button (handled by the template
-- gate, see layer / template edit). So: set source to the verb, and remove the
-- game-list fallback (the only one present). tool-list/guide-list have none.
--
-- For gamesdesign the three hubs exist (tools-index/guides-index/games-index),
-- so cta_url resolves to a real URL and the empty hrefs disappear on re-render.
-- The template gate (separate edit) is the correct-or-absent guarantee for
-- sites that lack a given hub.

-- ---------------------------------------------------------------------------
-- 0. Snapshot
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS content_components_bak_hubfix_0610 AS
SELECT * FROM content_components
WHERE name IN ('tool-list', 'game-list_pre_037', 'guide-list_pre_037');

-- ---------------------------------------------------------------------------
-- 1. tool-list — source identity.tools_index_url -> query.section_index_for:tool
-- ---------------------------------------------------------------------------
UPDATE content_components
SET input_schema = jsonb_set(input_schema,
        '{fields,cta_url,source}', '"query.section_index_for:tool"', true),
    updated_at = now()
WHERE name = 'tool-list';

-- ---------------------------------------------------------------------------
-- 2. game-list_pre_037 — source identity.games_index_url ->
--    query.section_index_for:game, AND drop the dead /games/index.html fallback
-- ---------------------------------------------------------------------------
UPDATE content_components
SET input_schema = jsonb_set(
        (input_schema #- '{fields,cta_url,fallback}'),
        '{fields,cta_url,source}', '"query.section_index_for:game"', true),
    updated_at = now()
WHERE name = 'game-list_pre_037';

-- ---------------------------------------------------------------------------
-- 3. guide-list_pre_037 — source navigation.guides_index_url ->
--    query.section_index_for:guide
-- ---------------------------------------------------------------------------
UPDATE content_components
SET input_schema = jsonb_set(input_schema,
        '{fields,cta_url,source}', '"query.section_index_for:guide"', true),
    updated_at = now()
WHERE name = 'guide-list_pre_037';

-- ---------------------------------------------------------------------------
-- 4. Verify — source repointed, no stray fallback on game-list.
-- ---------------------------------------------------------------------------
SELECT name,
       input_schema #>> '{fields,cta_url,source}'   AS cta_url_source,
       input_schema #>  '{fields,cta_url,fallback}' AS cta_url_fallback
FROM content_components
WHERE name IN ('tool-list', 'game-list_pre_037', 'guide-list_pre_037')
ORDER BY name;

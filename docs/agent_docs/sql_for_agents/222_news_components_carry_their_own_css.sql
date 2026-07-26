-- 222_news_components_carry_their_own_css.sql
--
-- bugs_open/072 — component markup ships without matching CSS on some sites.
--
-- WHAT IS WRONG
-- -------------
-- A site's assets/css/styles.css is a whole-file artefact written ONLY by a
-- webdesign-agent design run. RenderCSSFromSpecAction appends the css_snippets
-- whose applies_to overlaps the site's component list AT THAT INSTANT, and
-- nothing ever re-renders it afterwards. So a site that gains a component after
-- its last design run ships that component's markup with its CSS written
-- nowhere. Measured 2026-07-25 and again 07-26: of the five sites emitting
-- class="news-card", ai-agent-orchestration.com and relojistas.com have ZERO
-- matching rules. ai-agent-orchestration's stylesheet was written 2026-05-02;
-- its markup first carried news-card on 2026-07-21, 80 days later.
--
-- WHAT THIS FILE DOES
-- -------------------
-- Brings latest-news and news-listing into line with the platform's de-facto
-- rule: a component ships its own CSS in html_template. That is not a new
-- convention — 86 of the 94 component functions in use on active pages already
-- do it (hero, call-to-action, model-directory, the adoption/protocol trackers,
-- the report dossier). These two are stragglers, and the reason is recorded in
-- their own seed file: 089_latest_news.sql:11 says "styling comes from the
-- site's CSS theme", which was true when written and stopped being true the
-- first time a site gained news after its design run.
--
-- WHY THE CSS IS COPIED FROM css_snippets RATHER THAN WRITTEN OUT HERE
-- -------------------------------------------------------------------
-- Three of the five sites ALREADY serve these exact rules from styles.css. If
-- this file re-typed them, any drift — a changed colour, a dropped rule — would
-- restyle three live customer sites as a side effect of fixing two. Copying
-- css_content from the css_snippets row makes the two byte-identical by
-- construction, so the three styled sites see NO visual change: identical
-- selectors, identical declarations, later in document order.
--
-- The style block is placed where saveSectionsExtractFromHTML can capture it.
-- Its regex is (<section>...</section>)(<style>*)(<script>*) — style blocks
-- must precede script blocks or they are dropped when the section is saved.
-- latest-news carries its <script> AFTER </section>, so the block is inserted
-- BEFORE that script tag; news-listing carries its script INSIDE the section,
-- so the block is appended at the end. Both land immediately after the closing
-- </section>. Getting this backwards would persist a component whose CSS
-- silently never reaches page_components.rendered_html.
--
-- WHAT THIS FILE DOES NOT DO
-- --------------------------
-- It does not re-render any page. The change is live in the DB immediately and
-- INVISIBLE on every site until a page's components are re-rendered from their
-- templates — rerender_single_page concatenates stored rendered_html and does
-- not re-render templates. "The component was updated" is not evidence of
-- anything; verify on the rendered page (016b, and the bug file's own rule).
--
-- The companion Go change (collectComponentCSS in rerender_single_page_action.go)
-- is what fixes the pages already built and deployed; it injects the same
-- snippets at assembly time and deliberately SKIPS any component whose stored
-- rendered_html already carries its own <style>, so the two never both ship.
--
-- ROLLBACK
-- --------
--   UPDATE content_components cc
--      SET html_template = b.html_template
--     FROM component_news_backup_20260726_222 b
--    WHERE b.id = cc.id;

BEGIN;

-- ---------------------------------------------------------------------------
-- Backup (the 179_news_components_query_sourced_items.sql discipline)
-- ---------------------------------------------------------------------------
DROP TABLE IF EXISTS component_news_backup_20260726_222;
CREATE TABLE component_news_backup_20260726_222 AS
SELECT id, function, name, html_template, now() AS backed_up_at
FROM content_components
WHERE function IN ('latest-news', 'news-listing');

-- ---------------------------------------------------------------------------
-- Needle gates — fail loudly BEFORE writing if any premise has moved.
-- Every one of these is a fact this file depends on, checked rather than
-- assumed.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_grid_len   int;
  v_list_len   int;
  v_ln_tpl     text;
  v_nl_tpl     text;
BEGIN
  -- 1. Both snippets exist, are substantial, and carry the classes the
  --    markup actually emits.
  SELECT length(css_content) INTO v_grid_len
    FROM css_snippets WHERE name = 'Latest News Grid'
      AND css_content LIKE '%news-card%';
  IF COALESCE(v_grid_len, 0) < 1000 THEN
    RAISE EXCEPTION '222: css_snippets "Latest News Grid" missing, too small (%), or no .news-card rule', v_grid_len;
  END IF;

  SELECT length(css_content) INTO v_list_len
    FROM css_snippets WHERE name = 'News Listing Page'
      AND css_content LIKE '%news-list-item%';
  IF COALESCE(v_list_len, 0) < 1000 THEN
    RAISE EXCEPTION '222: css_snippets "News Listing Page" missing, too small (%), or no .news-list-item rule', v_list_len;
  END IF;

  -- 2. Neither snippet contains a </style> that would close the block early.
  IF EXISTS (SELECT 1 FROM css_snippets
              WHERE name IN ('Latest News Grid', 'News Listing Page')
                AND css_content ILIKE '%</style%') THEN
    RAISE EXCEPTION '222: a snippet contains </style> — wrapping it would break out of the block';
  END IF;

  -- 3. The two templates exist, have no <style> already, and have the shape
  --    the placement below depends on.
  SELECT html_template INTO v_ln_tpl FROM content_components WHERE function = 'latest-news';
  SELECT html_template INTO v_nl_tpl FROM content_components WHERE function = 'news-listing';

  IF v_ln_tpl IS NULL OR v_nl_tpl IS NULL THEN
    RAISE EXCEPTION '222: latest-news and/or news-listing component row not found';
  END IF;
  IF v_ln_tpl ILIKE '%<style%' OR v_nl_tpl ILIKE '%<style%' THEN
    RAISE EXCEPTION '222: a news template already carries a <style> block — already applied, or hand-edited; inspect before re-running';
  END IF;
  IF v_ln_tpl NOT LIKE '%<script src="/tools/assets/latest-news.js"></script>%' THEN
    RAISE EXCEPTION '222: latest-news template no longer carries its expected <script src> tag — placement anchor gone';
  END IF;
  IF rtrim(v_nl_tpl) NOT LIKE '%</section>' THEN
    RAISE EXCEPTION '222: news-listing template no longer ends with </section> — appending a style block would not be captured on save';
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- latest-news: insert the block BEFORE the trailing <script src>, so the
-- section-extraction regex (style-blocks-then-script-blocks) captures it.
-- ---------------------------------------------------------------------------
UPDATE content_components cc
SET html_template = replace(
      cc.html_template,
      '<script src="/tools/assets/latest-news.js"></script>',
      E'<style>\n' || s.css_content || E'\n</style>\n' ||
      '<script src="/tools/assets/latest-news.js"></script>'
    ),
    updated_at = NOW()
FROM css_snippets s
WHERE cc.function = 'latest-news'
  AND s.name = 'Latest News Grid'
  AND cc.html_template NOT LIKE '%<style%';

-- ---------------------------------------------------------------------------
-- news-listing: its <script> sits inside the section, so append after the
-- closing </section>.
-- ---------------------------------------------------------------------------
UPDATE content_components cc
SET html_template = cc.html_template ||
      E'\n<style>\n' || s.css_content || E'\n</style>',
    updated_at = NOW()
FROM css_snippets s
WHERE cc.function = 'news-listing'
  AND s.name = 'News Listing Page'
  AND cc.html_template NOT LIKE '%<style%';

-- ---------------------------------------------------------------------------
-- Post-write assertion. A guard that only checks "a <style> is present" would
-- pass on an empty block, so assert the CLASSES the markup emits are the ones
-- now styled, and that the block sits before the script tag on latest-news.
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_ln text;
  v_nl text;
BEGIN
  SELECT html_template INTO v_ln FROM content_components WHERE function = 'latest-news';
  SELECT html_template INTO v_nl FROM content_components WHERE function = 'news-listing';

  IF v_ln NOT LIKE '%<style>%' OR v_ln NOT LIKE '%news-card%' THEN
    RAISE EXCEPTION '222: latest-news did not gain a style block carrying .news-card';
  END IF;
  IF position('<style>' in v_ln) > position('<script src="/tools/assets/latest-news.js">' in v_ln) THEN
    RAISE EXCEPTION '222: latest-news style block landed AFTER its script tag — saveSectionsExtractFromHTML would drop it';
  END IF;
  IF v_nl NOT LIKE '%<style>%' OR v_nl NOT LIKE '%news-list-item%' THEN
    RAISE EXCEPTION '222: news-listing did not gain a style block carrying .news-list-item';
  END IF;

  RAISE NOTICE '222: latest-news template % -> % chars; news-listing % -> % chars',
    (SELECT length(html_template) FROM component_news_backup_20260726_222 WHERE function='latest-news'),
    length(v_ln),
    (SELECT length(html_template) FROM component_news_backup_20260726_222 WHERE function='news-listing'),
    length(v_nl);
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- Verify (run after applying):
--
--   SELECT function,
--          html_template LIKE '%<style%'        AS has_style,
--          html_template LIKE '%news-card%'     AS styles_cards,
--          length(html_template)                AS tpl_len
--   FROM content_components
--   WHERE function IN ('latest-news','news-listing');
--
-- Then, and only then, on the RENDERED page (the bug file's own bar):
--
--   for d in ai-agent-orchestration.com relojistas.com gaswholesalers.com robot-hands.com; do
--     html=$(curl -s "https://$d/"); css=$(curl -s "https://$d/assets/css/styles.css")
--     printf '%-28s uses=%s css_rules=%s inline=%s\n' "$d" \
--       "$(printf '%s' "$html" | grep -c 'class="news-card"')" \
--       "$(printf '%s' "$css"  | grep -c 'news-card')" \
--       "$(printf '%s' "$html" | grep -c 'news-card {')"
--   done
--
-- Pass = every row with uses>0 has css_rules>0 OR inline>0, AND gaswholesalers
-- and robot-hands are UNCHANGED — that control is what proves this added
-- styling rather than restyling live customer sites.
-- ---------------------------------------------------------------------------

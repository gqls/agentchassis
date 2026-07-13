-- migration_news_redesign_temporary.sql
--
-- TEMPORARY tactical fix for the news section visual redesign and
-- date-format expansion. Read the section "Why this is temporary" at
-- the bottom before re-using any of these patterns elsewhere.
--
-- What this does
-- --------------
-- 1. UPDATEs css_snippets rows for "Latest News Grid" and "News Listing
--    Page" with the redesigned CSS. This is the canonical change — every
--    site that uses the latest-news or news-listing component will pick
--    up the new CSS the next time webdesign-agent runs there.
--
-- 2. Surgically inserts formatNewsDate (a small JS function that expands
--    "2d ago" into "2 days ago" etc.) into the news IIFE in two places:
--      a) content_components.html_template for both news components
--         (the canonical template; future rebuilds will include it)
--      b) page_components.rendered_html for gaswholesalers' index and
--         news pages (immediate effect, picked up by next rerender)
--    Plus updates the call site so item.date passes through the new
--    function. Uses REPLACE() with dollar-quoted strings throughout
--    so the JS regex escapes survive intact.
--
-- 3. INSERTs a js_snippets row claiming the name "news-date-formatter"
--    with the function body. THIS ROW IS NOT CURRENTLY LOADED ANYWHERE
--    — the head component template has no snippet-loading mechanism. The
--    row exists so when the loader is built (TODO), the inline copies
--    can be deleted and this becomes the single source of truth.
--
-- Why this is temporary
-- ---------------------
-- - The news component's inline <script> in html_template violates
--   contract 003. Properly extracting it via separateInlineJS() would
--   make js_content the source of truth, with /tools/assets/latest-news.js
--   as the served file. Not done here because of time constraints; tracked
--   as a TODO.
-- - js_snippets has no loader. The contract describes a "head component
--   snippet-loading mechanism" but it doesn't exist in code. A small
--   half-day piece of work to mirror loadComponentCSSSnippets and have
--   RenderHead call it. Tracked as a TODO.
-- - Until the loader exists, formatNewsDate is duplicated inline in the
--   IIFE. When the loader lands, the inline copies come out, the
--   js_snippets row is what's actually loaded.

-- \set ON_ERROR_STOP on

BEGIN;

-- =====================================================================
-- 0. BEFORE: verification SELECTs
-- =====================================================================

-- '--- BEFORE: css_snippets news rows ---'
SELECT name, LENGTH(css_content) AS css_len, applies_to::text AS applies_to
FROM css_snippets
WHERE name IN ('Latest News Grid', 'News Listing Page')
ORDER BY name;

-- '--- BEFORE: content_components news rows ---'
SELECT function, name,
       LENGTH(html_template) AS tpl_len,
       html_template ILIKE '%formatNewsDate%' AS already_has_formatter
FROM content_components
WHERE function IN ('latest-news', 'news-listing')
ORDER BY function;

-- '--- BEFORE: gaswholesalers news page_components ---'
SELECT p.name AS page, pc.slot_name, pc.position,
       LENGTH(pc.rendered_html) AS html_len,
       pc.rendered_html ILIKE '%formatNewsDate%' AS already_has_formatter
FROM page_components pc
         JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name IN ('index', 'news')
  AND (pc.rendered_html ILIKE '%data-component="latest-news"%'
    OR pc.rendered_html ILIKE '%class="news-listing-section"%')
ORDER BY p.name, pc.position;

-- =====================================================================
-- 1. UPDATE css_snippets — Latest News Grid
-- =====================================================================
-- Canonical CSS for the homepage latest-news component. Applied on
-- the next webdesign-agent run for any site that uses this component.

UPDATE css_snippets
SET css_content = $CSS$
/* Latest news section — homepage card grid.
   Uses theme CSS custom properties with fallbacks. */

                      .latest-news-section {
    padding: 5rem 2rem;
background: var(--color-background, #f8fafc);
}
.latest-news-section .container {
  max-width: 1200px;
  margin: 0 auto;
}

.latest-news-section .section-heading {
  font-size: clamp(1.75rem, 3.5vw, 2.5rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.15;
  color: var(--color-heading, #0f172a);
  margin: 0 0 1rem;
  position: relative;
  padding-top: 1.5rem;
}
.latest-news-section .section-heading::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 2.5rem;
  height: 3px;
  background: var(--color-accent, #d97706);
  border-radius: 2px;
}

.latest-news-section .section-subheadline {
  font-size: 1.125rem;
  line-height: 1.6;
  color: var(--color-text-muted, #64748b);
  max-width: 64ch;
  margin: 0 0 3rem;
}

.latest-news-section .news-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.75rem;
}
.latest-news-section .news-empty {
  grid-column: 1 / -1;
  text-align: center;
  padding: 2rem;
  color: var(--color-text-muted, #64748b);
}

.news-card {
  background: var(--color-card-bg, #ffffff);
  border: 1px solid var(--color-border, #e2e8f0);
  border-radius: 0.5rem;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.news-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.08);
  border-color: var(--color-accent, #d97706);
}
.news-card-content {
  padding: 1.5rem 1.5rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  height: 100%;
}

.news-card-title {
  font-size: 1.125rem;
  font-weight: 600;
  line-height: 1.4;
  margin: 0;
}
.news-card-title a {
  color: var(--color-heading, #0f172a);
  text-decoration: none;
  transition: color 0.15s ease;
}
.news-card-title a:hover {
  color: var(--color-accent, #d97706);
}

.news-card-summary {
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--color-text, #475569);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.news-card-meta {
  display: flex;
  align-items: center;
  font-size: 0.8125rem;
  color: var(--color-text-muted, #64748b);
  margin-top: auto;
  padding-top: 0.5rem;
}
.news-card-meta .news-source {
  font-weight: 500;
}
.news-card-meta .news-source::after {
  content: "·";
  display: inline-block;
  margin: 0 0.5rem;
  color: var(--color-border, #cbd5e1);
  font-weight: 400;
}
.news-card-meta .news-date {
  font-variant-numeric: tabular-nums;
}

.news-section-footer {
  margin-top: 3rem;
  text-align: center;
}
.news-more-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-accent, #d97706);
  text-decoration: none;
  border: 1.5px solid var(--color-accent, #d97706);
  border-radius: 0.375rem;
  transition: background 0.15s ease, color 0.15s ease;
}
.news-more-link:hover {
  background: var(--color-accent, #d97706);
  color: #ffffff;
}

@media (max-width: 768px) {
  .latest-news-section { padding: 3.5rem 1.5rem; }
  .latest-news-section .news-grid { grid-template-columns: 1fr; gap: 1.25rem; }
  .latest-news-section .section-subheadline { margin-bottom: 2rem; }
}
$CSS$
WHERE name = 'Latest News Grid';

-- =====================================================================
-- 2. UPDATE css_snippets — News Listing Page
-- =====================================================================

UPDATE css_snippets
SET css_content = $CSS$
/* News listing page — full archive, long-form reading. */

.news-listing-section {
  padding: 5rem 2rem;
  background: var(--color-background, #f8fafc);
}
.news-listing-container {
  max-width: 760px;
  margin: 0 auto;
}

.news-listing-header {
  margin-bottom: 3rem;
  padding-bottom: 2rem;
  border-bottom: 2px solid var(--color-border, #e2e8f0);
}
.news-listing-title {
  font-size: clamp(2rem, 4vw, 2.75rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.15;
  color: var(--color-heading, #0f172a);
  margin: 0 0 1rem;
}
.news-listing-subtitle {
  font-size: 1.125rem;
  line-height: 1.6;
  color: var(--color-text-muted, #64748b);
  margin: 0;
  max-width: 64ch;
}

.news-listing-items {
  display: flex;
  flex-direction: column;
  gap: 0;
}
.news-listing-loading,
.news-listing-empty {
  padding: 3rem 0;
  text-align: center;
  color: var(--color-text-muted, #64748b);
}

.news-list-item {
  padding: 2rem 0;
  border-bottom: 1px solid var(--color-border, #e2e8f0);
}
.news-list-item:last-child { border-bottom: none; }
.news-list-item-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.news-list-item-title {
  font-size: 1.375rem;
  font-weight: 600;
  line-height: 1.35;
  margin: 0;
}
.news-list-item-title a {
  color: var(--color-heading, #0f172a);
  text-decoration: none;
  transition: color 0.15s ease;
}
.news-list-item-title a:hover {
  color: var(--color-accent, #d97706);
}

.news-list-item-summary {
  font-size: 1rem;
  line-height: 1.65;
  color: var(--color-text, #475569);
  margin: 0;
}

.news-list-item-meta {
  display: flex;
  align-items: center;
  font-size: 0.875rem;
  color: var(--color-text-muted, #64748b);
  margin-top: 0.25rem;
}
.news-list-item-source { font-weight: 500; }
.news-list-item-source::after {
  content: "·";
  display: inline-block;
  margin: 0 0.5rem;
  color: var(--color-border, #cbd5e1);
  font-weight: 400;
}
.news-list-item-date {
  font-variant-numeric: tabular-nums;
}

.news-list-item-topics {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.news-list-tag {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-text-muted, #64748b);
  background: var(--color-border, #e2e8f0);
  border-radius: 999px;
  letter-spacing: 0.02em;
}

.news-listing-footer {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 1px solid var(--color-border, #e2e8f0);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.875rem;
  color: var(--color-text-muted, #64748b);
}
.news-listing-count,
.news-listing-updated { margin: 0; }

@media (max-width: 768px) {
  .news-listing-section { padding: 3rem 1.25rem; }
  .news-listing-header { margin-bottom: 2rem; padding-bottom: 1.5rem; }
  .news-list-item { padding: 1.5rem 0; }
  .news-list-item-title { font-size: 1.1875rem; }
}
$CSS$
WHERE name = 'News Listing Page';

-- =====================================================================
-- 3. INSERT js_snippets row claiming "news-date-formatter"
-- =====================================================================
-- This row will NOT load anywhere today — the head component template
-- has no snippet-loading mechanism. The row exists so when the loader
-- is built, this is the canonical source and the inline copies (added
-- below) can be deleted in one pass.

INSERT INTO js_snippets (name, description, js_content, applies_to, semantic_tags)
VALUES (
  'news-date-formatter',
  'Expands abbreviated relative-time strings ("2d ago" -> "2 days ago") for news feeds. NOT YET LOADED — head component lacks snippet loading. Currently duplicated inline in news IIFEs.',
  $JS$function formatNewsDate(s) {
  if (!s) return "";
  s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
  s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
  s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
  s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
  return s;
}
$JS$,
  '["latest-news", "news-listing"]'::jsonb,
  '["news", "presentation", "time", "formatter", "not-yet-loaded"]'::jsonb
)
ON CONFLICT (name) DO NOTHING;

-- =====================================================================
-- 4. content_components.html_template — surgical REPLACE for latest-news
-- =====================================================================
-- Inserts formatNewsDate just before the fetch() call in the IIFE,
-- and updates the call site to wrap item.date in it.

UPDATE content_components
SET html_template = REPLACE(
  REPLACE(
    html_template,
    -- find:
    $OLD$"<time class=\"news-date\">" + item.date + "</time>"$OLD$,
    -- replace with:
    $NEW$"<time class=\"news-date\">" + formatNewsDate(item.date) + "</time>"$NEW$
  ),
  -- find:
  $OLD$fetch("/data/latest-news.json")$OLD$,
  -- replace with (function defined locally inside IIFE, just before fetch):
  $NEW$function formatNewsDate(s) {
    if (!s) return "";
    s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
    s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
    s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
    s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
    return s;
  }
  fetch("/data/latest-news.json")$NEW$
),
    updated_at = NOW()
WHERE function = 'latest-news';

-- =====================================================================
-- 5. content_components.html_template — surgical REPLACE for news-listing
-- =====================================================================

UPDATE content_components
SET html_template = REPLACE(
  REPLACE(
    html_template,
    $OLD$"<span class=\"news-list-item-date\">" + item.date + "</span>"$OLD$,
    $NEW$"<span class=\"news-list-item-date\">" + formatNewsDate(item.date) + "</span>"$NEW$
  ),
  $OLD$fetch("/data/news-archive.json")$OLD$,
  $NEW$function formatNewsDate(s) {
      if (!s) return "";
      s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
      s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
      s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
      s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
      return s;
    }
    fetch("/data/news-archive.json")$NEW$
),
    updated_at = NOW()
WHERE function = 'news-listing';

-- =====================================================================
-- 6. page_components.rendered_html for gaswholesalers index (latest-news)
-- =====================================================================
-- Direct surgery on the already-rendered HTML so the change goes live
-- on next page-rerender, without waiting for a content-writer rebuild.

UPDATE page_components
SET rendered_html = REPLACE(
  REPLACE(
    rendered_html,
    $OLD$"<time class=\"news-date\">" + item.date + "</time>"$OLD$,
    $NEW$"<time class=\"news-date\">" + formatNewsDate(item.date) + "</time>"$NEW$
  ),
  $OLD$fetch("/data/latest-news.json")$OLD$,
  $NEW$function formatNewsDate(s) {
    if (!s) return "";
    s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
    s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
    s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
    s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
    return s;
  }
  fetch("/data/latest-news.json")$NEW$
),
    updated_at = NOW()
FROM pages p
WHERE page_components.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'index'
  AND page_components.rendered_html ILIKE '%data-component="latest-news"%';

-- =====================================================================
-- 7. page_components.rendered_html for gaswholesalers news (news-listing)
-- =====================================================================

UPDATE page_components
SET rendered_html = REPLACE(
  REPLACE(
    rendered_html,
    $OLD$"<span class=\"news-list-item-date\">" + item.date + "</span>"$OLD$,
    $NEW$"<span class=\"news-list-item-date\">" + formatNewsDate(item.date) + "</span>"$NEW$
  ),
  $OLD$fetch("/data/news-archive.json")$OLD$,
  $NEW$function formatNewsDate(s) {
      if (!s) return "";
      s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
      s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
      s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
      s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
      return s;
    }
    fetch("/data/news-archive.json")$NEW$
),
    updated_at = NOW()
FROM pages p
WHERE page_components.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'news'
  AND page_components.rendered_html ILIKE '%class="news-listing-section"%';

-- =====================================================================
-- 8. AFTER: verification SELECTs
-- =====================================================================

-- '--- AFTER: css_snippets news rows ---'
SELECT name, LENGTH(css_content) AS css_len
FROM css_snippets
WHERE name IN ('Latest News Grid', 'News Listing Page')
ORDER BY name;

-- '--- AFTER: content_components news rows (should now contain formatNewsDate) ---'
SELECT function, name,
       LENGTH(html_template) AS tpl_len,
       html_template ILIKE '%formatNewsDate%' AS has_formatter,
       html_template ILIKE '%formatNewsDate(item.date)%' AS has_call_site
FROM content_components
WHERE function IN ('latest-news', 'news-listing')
ORDER BY function;

-- '--- AFTER: gaswholesalers news page_components (should have_formatter = true) ---'
SELECT p.name AS page,
       LENGTH(pc.rendered_html) AS html_len,
       pc.rendered_html ILIKE '%formatNewsDate%' AS has_formatter,
       pc.rendered_html ILIKE '%formatNewsDate(item.date)%' AS has_call_site
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name IN ('index', 'news')
  AND (pc.rendered_html ILIKE '%data-component="latest-news"%'
    OR pc.rendered_html ILIKE '%class="news-listing-section"%')
ORDER BY p.name;

-- '--- AFTER: js_snippets news-date-formatter ---'
SELECT name, LENGTH(js_content) AS js_len, applies_to::text AS applies_to
FROM js_snippets
WHERE name = 'news-date-formatter';

COMMIT;

-- =====================================================================
-- Sanity check: if any of these report has_formatter = false after
-- COMMIT, the REPLACE didn't match and the change didn't take. Most
-- likely cause: whitespace mismatch between expected pattern and actual
-- stored text. Roll back manually if needed and adjust patterns.
--
-- The migration is safe to re-run; the inline-already check
-- (already_has_formatter = true in the BEFORE block) tells you whether
-- a re-run would be a no-op or risk double-insertion. The REPLACE
-- patterns use the OLD form of the call site, so once changed they
-- won't re-match.
-- =====================================================================
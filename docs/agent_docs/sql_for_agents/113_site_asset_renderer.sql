-- migration_b_news_redesign_proper.sql
--
-- The proper news redesign migration. Apply migration_a first.
--
-- Changed from previous draft: steps 4 and 5 (contract-003 migration
-- of the news components) no longer use regex. They use split_part()
-- to extract the IIFE body and position() + substring() to do the
-- surgery on html_template. See "Why no regex" at the bottom.

\set ON_ERROR_STOP on

BEGIN;


-- =====================================================================
-- 1. UPDATE css_snippets — Latest News Grid
-- =====================================================================

UPDATE css_snippets
SET css_content = $CSS$
/* Latest news section — homepage card grid.
   Uses theme CSS custom properties with fallbacks. */

                      .latest-news-section {
    padding: 5rem 2rem;
background: var(--color-background, #f8fafc);
}
.latest-news-section .container { max-width: 1200px; margin: 0 auto; }

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
  top: 0; left: 0;
  width: 2.5rem; height: 3px;
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
  display: flex; flex-direction: column;
  gap: 0.75rem; height: 100%;
}

.news-card-title { font-size: 1.125rem; font-weight: 600; line-height: 1.4; margin: 0; }
.news-card-title a {
  color: var(--color-heading, #0f172a);
  text-decoration: none;
  transition: color 0.15s ease;
}
.news-card-title a:hover { color: var(--color-accent, #d97706); }

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
  display: flex; align-items: center;
  font-size: 0.8125rem;
  color: var(--color-text-muted, #64748b);
  margin-top: auto;
  padding-top: 0.5rem;
}
.news-card-meta .news-source { font-weight: 500; }
.news-card-meta .news-source::after {
  content: "·"; display: inline-block; margin: 0 0.5rem;
  color: var(--color-border, #cbd5e1); font-weight: 400;
}
.news-card-meta .news-date { font-variant-numeric: tabular-nums; }

.news-section-footer { margin-top: 3rem; text-align: center; }
.news-more-link {
  display: inline-flex; align-items: center; gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 1rem; font-weight: 600;
  color: var(--color-accent, #d97706);
  text-decoration: none;
  border: 1.5px solid var(--color-accent, #d97706);
  border-radius: 0.375rem;
  transition: background 0.15s ease, color 0.15s ease;
}
.news-more-link:hover { background: var(--color-accent, #d97706); color: #ffffff; }

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
.news-listing-container { max-width: 760px; margin: 0 auto; }

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

.news-listing-items { display: flex; flex-direction: column; gap: 0; }
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
.news-list-item-content { display: flex; flex-direction: column; gap: 0.75rem; }

.news-list-item-title { font-size: 1.375rem; font-weight: 600; line-height: 1.35; margin: 0; }
.news-list-item-title a {
  color: var(--color-heading, #0f172a);
  text-decoration: none;
  transition: color 0.15s ease;
}
.news-list-item-title a:hover { color: var(--color-accent, #d97706); }

.news-list-item-summary {
  font-size: 1rem;
  line-height: 1.65;
  color: var(--color-text, #475569);
  margin: 0;
}

.news-list-item-meta {
  display: flex; align-items: center;
  font-size: 0.875rem;
  color: var(--color-text-muted, #64748b);
  margin-top: 0.25rem;
}
.news-list-item-source { font-weight: 500; }
.news-list-item-source::after {
  content: "·"; display: inline-block; margin: 0 0.5rem;
  color: var(--color-border, #cbd5e1); font-weight: 400;
}
.news-list-item-date { font-variant-numeric: tabular-nums; }

.news-list-item-topics {
  display: flex; flex-wrap: wrap; gap: 0.5rem;
  margin-top: 0.5rem;
}
.news-list-tag {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem; font-weight: 500;
  color: var(--color-text-muted, #64748b);
  background: var(--color-border, #e2e8f0);
  border-radius: 999px;
  letter-spacing: 0.02em;
}

.news-listing-footer {
  margin-top: 3rem;
  padding-top: 2rem;
  border-top: 1px solid var(--color-border, #e2e8f0);
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 1rem;
  font-size: 0.875rem;
  color: var(--color-text-muted, #64748b);
}
.news-listing-count, .news-listing-updated { margin: 0; }

@media (max-width: 768px) {
  .news-listing-section { padding: 3rem 1.25rem; }
  .news-listing-header { margin-bottom: 2rem; padding-bottom: 1.5rem; }
  .news-list-item { padding: 1.5rem 0; }
  .news-list-item-title { font-size: 1.1875rem; }
}
$CSS$
WHERE name = 'News Listing Page';


-- =====================================================================
-- 3. INSERT js_snippets — news-date-formatter (active)
-- =====================================================================

INSERT INTO js_snippets (name, description, js_content, applies_to, semantic_tags, is_active)
VALUES (
  'news-date-formatter',
  'Expands abbreviated relative-time strings ("2d ago" -> "2 days ago") for news feeds. Loaded via /assets/js/snippets.js, available globally.',
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
  '["news", "presentation", "time", "formatter"]'::jsonb,
  true
)
ON CONFLICT (name) DO UPDATE SET
  description = EXCLUDED.description,
  js_content  = EXCLUDED.js_content,
  applies_to  = EXCLUDED.applies_to,
  is_active   = EXCLUDED.is_active;


-- =====================================================================
-- 4. content_components — migrate news components to contract 003
-- =====================================================================
-- Extract the inline <script>...</script> body into js_content; replace
-- the inline block with a <script src> tag.
--
-- No regex. CTE captures the slice positions and extracted body once.
-- The UPDATE references the CTE so no extraction work is done twice.

WITH news_components AS (
  SELECT
    id, function, html_template,
    -- 1-based position of <script> tag (0 if not found)
    position('<script>' IN html_template)  AS script_start,
    -- 1-based position of </script> tag
    position('</script>' IN html_template) AS script_end,
    -- IIFE body: everything between the FIRST <script> and the NEXT </script>.
    --   split_part(x, '<script>', 2)  → everything after first <script>
    --   split_part(THAT, '</script>', 1) → everything before next </script>
    split_part(
      split_part(html_template, '<script>', 2),
      '</script>',
      1
    ) AS iife_body
  FROM content_components
  WHERE function IN ('latest-news', 'news-listing')
)
UPDATE content_components cc
SET
  -- IIFE body → js_content, with date call site rewritten to use the global
  js_content = TRIM(REPLACE(
    nc.iife_body,
    CASE cc.function
      WHEN 'latest-news'  THEN '"<time class=\"news-date\">" + item.date + "</time>"'
      WHEN 'news-listing' THEN '"<span class=\"news-list-item-date\">" + item.date + "</span>"'
    END,
    CASE cc.function
      WHEN 'latest-news'  THEN '"<time class=\"news-date\">" + formatNewsDate(item.date) + "</time>"'
      WHEN 'news-listing' THEN '"<span class=\"news-list-item-date\">" + formatNewsDate(item.date) + "</span>"'
    END
  )),
  -- html_template = (text before <script>) || (new <script src>) || (text after </script>)
  html_template =
    substring(nc.html_template FROM 1 FOR nc.script_start - 1) ||
    '<script src="/tools/assets/' || cc.function || '.js"></script>' ||
    substring(nc.html_template FROM nc.script_end + length('</script>')),
  updated_at = NOW()
FROM news_components nc
WHERE cc.id = nc.id
  -- Guards (all must be true):
  AND nc.script_start > 0                       -- found a <script> tag
  AND nc.script_end   > nc.script_start         -- found </script> after it
  AND length(TRIM(nc.iife_body)) > 0            -- extraction has content
  AND cc.html_template NOT LIKE '%<script src="/tools/assets/' || cc.function || '.js"></script>%';  -- not already migrated


-- =====================================================================
-- 5. content_components — head templates: add snippets.js loader
-- =====================================================================

UPDATE content_components
SET
  html_template = REPLACE(
    html_template,
    '</head>',
    E'    <script src="/assets/js/snippets.js"></script>\n</head>'
  ),
  updated_at = NOW()
WHERE function = 'head'
  AND name IN ('head-seo-standard', 'Document Head')
  AND html_template LIKE '%</head>%'
  AND html_template NOT LIKE '%/assets/js/snippets.js%';


-- =====================================================================
-- 6. agent_definitions — INSERT site-asset-renderer
-- =====================================================================

INSERT INTO agent_definitions (
  name, display_name, description, role, config, is_active,
  tags, image, version, resources, channels, health, dependencies,
  priority, fallback_config, processing_mode, status, applies_to,
  input_schema, output_schema
)
VALUES (
  'site-asset-renderer',
  'Site Asset Renderer',
  'Renders /assets/js/snippets.js for a site and commits to git. Deterministic — no LLM. Triggered when js_snippets or component set changes, or invoked by webdesign-agent.',
  'specialist',
  $WF${
    "workflow": {
      "start_step": "load_site",
      "steps": {
        "load_site": {
          "action": "ensure_site_record",
          "config": { "site_id_field": "input_data.site_id", "domain_field": "input_data.domain" },
          "next_step": "render_js_snippets",
          "output_field": "site_record",
          "description": "Resolve site_id and domain"
        },
        "render_js_snippets": {
          "action": "render_js_snippets_for_site",
          "config": { "site_id_field": "site_record.site_id" },
          "next_step": "deploy_js_snippets",
          "output_field": "js_snippets_render",
          "description": "Concatenate active js_snippets for this site"
        },
        "deploy_js_snippets": {
          "action": "git_commit",
          "config": {
            "domain_field": "site_record.domain",
            "files_field": "js_snippets_render.files",
            "commit_message": "Update JS snippets bundle"
          },
          "next_step": "complete",
          "output_field": "deploy_result",
          "description": "Commit assets/js/snippets.js"
        },
        "complete": {
          "action": "complete_workflow",
          "config": { "output_fields": ["site_record", "js_snippets_render", "deploy_result"] },
          "description": "Asset render complete"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 120
  }$WF$::jsonb,
  true,
  ARRAY['assets', 'snippets', 'site-level'],
  'docker.io/aqls/agent-chassis',
  'v1.0.1012',
  '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "128Mi"}}'::jsonb,
  '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
  '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb,
  '[]'::jsonb,
  1,
  '{"fallback_to_self": true, "prefer_delegation": true}'::jsonb,
  'executor',
  'active',
  '["website"]'::jsonb,
  '{"required": ["site_id"], "optional": ["domain"], "description": "Provide site_id; domain is loaded from sites table if absent."}'::jsonb,
  '{"produces": {"js_snippets_render": "Bundle metadata", "deploy_result": "git_commit result"}}'::jsonb
)
ON CONFLICT (name) DO UPDATE SET
  config      = EXCLUDED.config,
  description = EXCLUDED.description,
  updated_at  = NOW();


-- =====================================================================
-- 7. Verification SELECTs
-- =====================================================================

\echo '--- AFTER: css_snippets news rows ---'
SELECT name, LENGTH(css_content) AS css_len
FROM css_snippets
WHERE name IN ('Latest News Grid', 'News Listing Page')
ORDER BY name;

\echo '--- AFTER: js_snippets news-date-formatter ---'
SELECT name, is_active, LENGTH(js_content) AS js_len, applies_to::text AS applies_to
FROM js_snippets
WHERE name = 'news-date-formatter';

\echo '--- AFTER: content_components news components ---'
SELECT
  function,
  html_template LIKE '%<script src="/tools/assets/%' AS has_script_src,
  html_template LIKE '%<script>%(function%'          AS still_has_inline_iife,
  LENGTH(COALESCE(js_content, ''))                   AS js_content_len,
  js_content LIKE '%formatNewsDate(item.date)%'      AS js_calls_formatter
FROM content_components
WHERE function IN ('latest-news', 'news-listing')
ORDER BY function;

\echo '--- AFTER: head components have snippets.js loader ---'
SELECT function, name, html_template LIKE '%/assets/js/snippets.js%' AS has_loader
FROM content_components
WHERE function = 'head'
ORDER BY name;

\echo '--- AFTER: site-asset-renderer agent exists ---'
SELECT name, status, is_active, version
FROM agent_definitions
WHERE name = 'site-asset-renderer';

COMMIT;


-- =====================================================================
-- Why no regex
-- =====================================================================
--
-- Previous draft used:
--   regexp_replace(html_template, '^.*<script>\s*(.*?)\s*</script>\s*$', E'\\1', 's')
--
-- Three brittleness sources:
--
-- 1. ^.* and \s*$ anchored the script to the start and end of the
--    template (modulo whitespace). Any trailing comment, blank line,
--    or stray markup after </script> made the match silently fail
--    and the whole UPDATE became a no-op.
--
-- 2. PostgreSQL's regex newline-sensitivity has a non-obvious default
--    (`.` matches newlines without `n` flag), easy to get subtly wrong.
--
-- 3. (.*?) lazy matching combined with anchors interacts oddly with
--    leading/trailing whitespace.
--
-- The new approach (split_part + position + substring):
--
-- - No regex: no engine flags, no anchors, no greedy/lazy concerns.
-- - Tolerant of any whitespace inside the script (we slice between the
--   start of <script> and the start of </script> — no \s* needed).
-- - Tolerant of trailing content after </script> (we preserve it).
-- - Explicit: each piece of the new html_template is a named slice,
--   readable without parsing a regex in your head.
-- - Idempotent: the LIKE guards skip already-migrated rows.
--
-- One restriction: assumes exactly ONE inline <script> block (no
-- attributes) per news component html_template. True today and
-- consistent with contract 003. If a component later has multiple
-- inline scripts, the right fix is splitting it into multiple
-- components, not making the extractor cleverer.
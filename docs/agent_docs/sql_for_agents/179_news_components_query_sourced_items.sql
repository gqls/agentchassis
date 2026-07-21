-- ============================================================================
-- 179 — news components: items as query.* schema fields, rendered server-side
-- (bugs_open/027 rework, part B config. Part A = migration 178, applied.)
--
-- WHAT: converts `news-listing` to the v2 `fields` input_schema (its current
-- JSON-Schema-style shape is INVISIBLE to plan_sections — only "fields" is
-- parsed, plan_sections_action.go:1137, which is also why its required
-- headline could ship empty: no declared fields meant no guidance and no
-- gate; see bugs_open/026 defect B), adds a query-sourced `items` field to
-- BOTH news components, and rewrites both html_templates to render the items
-- server-side. The resolvers (queryresolve: latest_news, news_archive) ship
-- in the image built from the same commit as this file.
--
-- WHY THIS SHAPE: 003's source-of-truth contract — items land in content_data
-- via plan_sections/queryresolve like every other section field, the template
-- renders them, and a scoped rerender REFRESHES them instead of wiping them
-- (the fate of the withdrawn rendered_html-injection approach, council REVISE
-- 4b91237a).
--
-- ORDERING: safe to apply BEFORE the image carrying the resolvers rolls.
-- `items` is required:false with on_missing skip_field, so on the old binary
-- an unknown query name errors the field resolution → field skipped → the
-- template's {{if .items}} takes the else branch → placeholder, exactly
-- today's behaviour. When the new image lands, the next build/rerender fills
-- items. (Old binary v1.0.1140 also still carries the injection writer; its
-- anchors survive these templates, and the same commit that ships the
-- resolvers deletes it.)
--
-- ESCAPING: resolver output is HTML-escaped at projection (templates render
-- via text/template, which does not escape; feed text is third-party). The
-- templates therefore interpolate {{.title}} etc. directly.
--
-- 026 NOTE: the news-listing loading placeholder becomes a translatable
-- llm-sourced field with an English fallback — closes 026 defect A for this
-- component (the hardcoded-English half); defect B's enforcement gap (a
-- required field rendering empty and saving anyway) stays open and is 026's.
-- ============================================================================

BEGIN;

-- Snapshots (no version history exists for these columns).
CREATE TABLE IF NOT EXISTS component_news_backup_20260720_179 AS
SELECT id, function, input_schema, html_template, now() AS backed_up_at
  FROM content_components
 WHERE function IN ('news-listing', 'latest-news');

-- ---------------------------------------------------------------------------
-- Needle-gates: fail loudly if the rows are not in the state this migration
-- was written against (another thread may have edited them).
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  nl_tpl  text; nl_sch  text; ln_tpl  text; ln_sch  text;
BEGIN
  SELECT html_template, input_schema::text INTO nl_tpl, nl_sch
    FROM content_components WHERE function = 'news-listing';
  SELECT html_template, input_schema::text INTO ln_tpl, ln_sch
    FROM content_components WHERE function = 'latest-news';

  IF nl_tpl NOT LIKE '%id="news-listing-items"%' OR
     nl_tpl NOT LIKE '%Loading latest news...%' THEN
    RAISE EXCEPTION 'needle-gate: news-listing template drifted from expected pre-state';
  END IF;
  IF nl_sch NOT LIKE '%"properties"%' OR nl_sch LIKE '%"fields"%' THEN
    RAISE EXCEPTION 'needle-gate: news-listing schema is not the expected properties-shape';
  END IF;
  IF ln_tpl NOT LIKE '%id="news-container"%' OR ln_tpl NOT LIKE '%<noscript>%' THEN
    RAISE EXCEPTION 'needle-gate: latest-news template drifted from expected pre-state';
  END IF;
  IF ln_sch NOT LIKE '%"fields"%' OR ln_sch LIKE '%news_archive%' THEN
    RAISE EXCEPTION 'needle-gate: latest-news schema not in expected pre-state';
  END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 1. news-listing: v2 schema (headline/subheadline preserved as llm fields;
--    items query-sourced; loading_text closes 026-A for this component).
-- ---------------------------------------------------------------------------
UPDATE content_components SET
  input_schema = $sch${
    "fields": {
      "headline": {
        "type": "text", "source": "llm", "required": true,
        "llm_guidance": "Page headline for the news listing, in the site's own language."
      },
      "subheadline": {
        "type": "text", "source": "llm", "required": false,
        "llm_guidance": "Optional one-line context under the headline, in the site's own language."
      },
      "loading_text": {
        "type": "text", "source": "llm", "required": false,
        "llm_guidance": "Short placeholder shown while news loads, in the site's own language (e.g. Spanish 'Cargando noticias…')."
      },
      "items": {
        "type": "array", "source": "query.news_archive", "limit": 20,
        "required": false, "on_missing": "skip_field",
        "items": {
          "title": {"type": "text"}, "summary": {"type": "text"},
          "url": {"type": "url"}, "source": {"type": "text"},
          "date": {"type": "text"}, "topics": {"type": "array"}
        }
      }
    },
    "data_source": "/data/news-archive.json",
    "render_action": "render_news_section",
    "notes": "items are server-rendered from query.news_archive (bugs_open/027); the JSON + news-listing.js client refresh stays as the freshness path between rerenders. Resolver output is pre-escaped."
  }$sch$::jsonb,
  html_template = $tpl$<section data-component="news-listing" class="news-listing-section" id="news-listing">
  <div class="news-listing-container">
    <div class="news-listing-header">
      <h1 class="news-listing-title">{{.headline}}</h1>
      {{if .subheadline}}<p class="news-listing-subtitle">{{.subheadline}}</p>{{end}}
    </div>
    <div class="news-listing-items" id="news-listing-items">
      {{if .items}}{{range .items}}
      <article class="news-list-item">
        <div class="news-list-item-content">
          <h3 class="news-list-item-title"><a href="{{.url}}" target="_blank" rel="noopener noreferrer">{{.title}}</a></h3>
          {{if .summary}}<p class="news-list-item-summary">{{.summary}}</p>{{end}}
          <div class="news-list-item-meta">
            {{if .source}}<span class="news-list-item-source">{{.source}}</span>{{end}}
            {{if .date}}<span class="news-list-item-date">{{.date}}</span>{{end}}
          </div>
          {{if .topics}}<div class="news-list-item-topics">{{range .topics}}<span class="news-list-tag">{{.}}</span>{{end}}</div>{{end}}
        </div>
      </article>
      {{end}}{{else}}<p class="news-listing-loading">{{if .loading_text}}{{.loading_text}}{{else}}Loading latest news...{{end}}</p>{{end}}
    </div>
    <div class="news-listing-footer" id="news-listing-footer" style="display:none;">
      <p class="news-listing-count" id="news-listing-count"></p>
      <p class="news-listing-updated" id="news-listing-updated"></p>
    </div>
  </div>
  <script src="/tools/assets/news-listing.js"></script>
</section>$tpl$,
  updated_at = now()
WHERE function = 'news-listing';

-- ---------------------------------------------------------------------------
-- 2. latest-news: add items field; render cards server-side, keep the
--    noscript fallback for the no-items case.
-- ---------------------------------------------------------------------------
UPDATE content_components SET
  input_schema = jsonb_set(
    input_schema,
    '{fields,items}',
    $itm${
      "type": "array", "source": "query.latest_news", "limit": 6,
      "required": false, "on_missing": "skip_field",
      "items": {
        "title": {"type": "text"}, "summary": {"type": "text"},
        "url": {"type": "url"}, "source": {"type": "text"}, "date": {"type": "text"}
      }
    }$itm$::jsonb
  ) || $nts${"notes": "items are server-rendered from query.latest_news (bugs_open/027); latest-news.js refreshes client-side and leaves server HTML alone when the fetch has nothing. Resolver output is pre-escaped."}$nts$::jsonb,
  html_template = $tpl$<!-- latest-news component -->
<section data-component="latest-news" class="latest-news-section section-padding">
  <div class="container">
    <h2 class="section-heading" id="news-headline">{{.headline}}</h2>
    {{if .subheadline}}<p class="section-subheadline" id="news-subheadline">{{.subheadline}}</p>{{end}}
    <div class="news-grid" id="news-container">
      {{if .items}}{{range .items}}
      <article class="news-card"><div class="news-card-content">
        <h3 class="news-card-title"><a href="{{.url}}" target="_blank" rel="noopener noreferrer">{{.title}}</a></h3>
        {{if .summary}}<p class="news-card-summary">{{.summary}}</p>{{end}}
        <div class="news-card-meta">
          {{if .source}}<span class="news-source">{{.source}}</span>{{end}}
          {{if .date}}<time class="news-date">{{.date}}</time>{{end}}
        </div>
      </div></article>
      {{end}}{{else}}<noscript>
        <p class="news-empty">Enable JavaScript to see the latest news{{if .insights_url}}, or visit
        <a href="{{.insights_url}}">our insights page</a>{{end}}.</p>
      </noscript>{{end}}
    </div>
    <div id="news-footer"></div>
  </div>
</section>
<script src="/tools/assets/latest-news.js"></script>$tpl$,
  updated_at = now()
WHERE function = 'latest-news';

COMMIT;

-- ----------------------------------------------------------------------------
-- Verify (expect: both rows fields-shaped, both templates carry range .items,
-- news-listing keeps its container ids for the JS):
--   SELECT function,
--          input_schema ? 'fields'                          AS v2,
--          input_schema->'fields' ? 'items'                 AS has_items,
--          html_template LIKE '%{{range .items}}%'          AS ranges,
--          html_template LIKE '%id="news-listing-items"%'
--            OR function = 'latest-news'                    AS ids_kept
--     FROM content_components
--    WHERE function IN ('news-listing','latest-news');
--
-- Rollback:
--   UPDATE content_components c
--      SET input_schema = b.input_schema, html_template = b.html_template
--     FROM component_news_backup_20260720_179 b
--    WHERE c.id = b.id;
-- ----------------------------------------------------------------------------

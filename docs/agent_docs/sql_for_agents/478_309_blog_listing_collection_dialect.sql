-- 478_309_blog_listing_collection_dialect.sql
--
-- bugs_open/309, the CASE repair. Owner-chosen fix candidate 1 (2026-08-18),
-- independently endorsed by the bugfix-309 lane in that file's section 8.
--
-- WHAT IS WRONG. `blog-listing_pre_037` is the last list component still on the
-- legacy numbered-flat dialect: six `postN_url` fields sourcing
-- `site_specs.blog.postN_url`. `SELECT count(*) FROM site_specs WHERE
-- aspect='blog'` is 0 with no is_current filter -- that aspect has never existed
-- on any site, ever. So the source resolves nothing, `on_missing` defaults to
-- `skip_field`, the key never reaches `page_components.content_data`, and the
-- template's `{{if .postN_url}}` drops the anchor. Not an empty href -- NO
-- ANCHOR AT ALL, which is why it is invisible to markup-shape checks.
-- `090` verdict CONFIRMED, run correlation 6e578bf5-778a-4e72-aab2-0531e45c07d8.
--
-- WHAT THIS DOES. Moves the component onto the collection dialect its three
-- siblings already use (`tool-list`, `game-list_pre_037`, `guide-list_pre_037`
-- -> `query.pages_where_type:*`; `content-listing` -> `query.blog_posts`).
-- Titles and URLs now come from the same `pages` row and cannot disagree.
--
-- WHAT IS DELIBERATELY LOST, and why that is a fix and not a regression:
--   * per-card date, category, read time and author were `source: llm` -- the
--     model INVENTED them. The live page serves "February 18, 2025" /
--     "March 4, 2025" for articles whose pages were created 2026-07-25..08-07
--     and deployed 2026-08-17. Those are fabricated, wrong by ~18 months, and
--     on a site that sells verifiable governance. The resolver returns no such
--     fields, so they go rather than being re-fabricated.
--   * the category filter nav and /tools/assets/blog-listing.js: they filtered
--     on the fabricated category. With no per-item category there is nothing
--     honest to filter, so the control goes with the data. The `.bl-filters`
--     CSS is left in place deliberately -- dead CSS is inert, and editing the
--     stylesheet widens this change for no gain.
--   * `cta_url` sourced `site_specs.blog.archive_url` -- the SAME dead aspect.
--     It becomes an empty source (exactly `guide-list_pre_037`'s live shape),
--     so the `{{if}}`-gated CTA simply does not render. Leaving it would trip
--     the new source-vocabulary birth gate (0df9f1be9) on the next regeneration.
--
-- BLAST RADIUS. Two live instances fleet-wide: fundamentallyai.com
-- platform-log-index (the broken one) and leopardessconsulting.co.uk blog, whose
-- rendered_html is overwritten by RebuildBlogListingAction and is therefore
-- insulated until something re-renders it from this template.
--
-- INERT UNTIL RE-PLANNED. DB config is live immediately, but the page keeps
-- serving its stored rendered_html until a rerender re-resolves the sources.
-- Verify at the SERVED page, never here.
--
-- ROLLBACK: 478_..._ROLLBACK.sql restores both columns from the backup table
-- this migration creates.

BEGIN;

DO $mig478$
DECLARE n int; phantom int;
BEGIN
    SELECT count(*) INTO n FROM content_components WHERE name = 'blog-listing_pre_037';
    IF n <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 478: expected exactly 1 blog-listing_pre_037 row, found %', n;
    END IF;

    -- Drift guard: refuse if the phantom source is already gone (someone else
    -- migrated it, or this migration already ran). A SELECT here could not stop
    -- the COMMIT; RAISE can.
    SELECT count(*) INTO phantom
      FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
     WHERE cc.name = 'blog-listing_pre_037'
       AND f.value->>'source' LIKE 'site_specs.blog.%';
    IF phantom = 0 THEN
        RAISE EXCEPTION 'MIGRATION 478: blog-listing_pre_037 declares no site_specs.blog.* source -- already migrated, or changed underneath this migration. Refusing.';
    END IF;
    RAISE NOTICE 'migration 478: % phantom site_specs.blog.* field(s) to retire', phantom;
END
$mig478$;

CREATE TABLE IF NOT EXISTS content_components_bak_20260818_309_blog_listing AS
  SELECT * FROM content_components WHERE name = 'blog-listing_pre_037';

UPDATE content_components
   SET input_schema  = $sch478${
  "fields": {
    "articles": {
      "type": "array",
      "source": "query.blog_posts",
      "required": true,
      "on_missing": "skip_section",
      "missing_reason": "No articles published yet"
    },
    "section_heading": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Heading for the article listing section."
    },
    "section_intro": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "One or two sentences introducing the listing."
    },
    "eyebrow_label": {
      "type": "text",
      "source": "static",
      "fallback": "From the Blog",
      "required": false,
      "on_missing": "use_fallback"
    },
    "read_more_label": {
      "type": "text",
      "source": "static",
      "fallback": "Read more",
      "required": false,
      "on_missing": "use_fallback"
    },
    "cta_label": {
      "type": "text",
      "source": "static",
      "fallback": "View all articles",
      "required": false,
      "on_missing": "use_fallback"
    },
    "cta_url": {
      "type": "url",
      "source": "",
      "required": false,
      "on_missing": "skip_field"
    }
  }
}$sch478$::jsonb,
       html_template = $tmpl478$<style>
.blog-listing-section {
  padding: var(--spacing-section, 5rem 2rem);
  background: var(--color-background, #fff);
  color: var(--color-text, #333);
}
.blog-listing-section .bl-inner {
  max-width: var(--container-max-width, 1200px);
  margin: 0 auto;
}
.blog-listing-section .bl-header {
  text-align: center;
  margin-bottom: 3rem;
}
.blog-listing-section .bl-eyebrow {
  display: inline-block;
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-primary, #2563eb);
  margin-bottom: 0.75rem;
}
.blog-listing-section .bl-heading {
  font-size: clamp(1.75rem, 4vw, 2.75rem);
  font-weight: 800;
  color: var(--color-heading, #111);
  margin: 0 0 1rem;
  line-height: 1.2;
}
.blog-listing-section .bl-intro {
  font-size: 1.1rem;
  color: var(--color-text-muted, #666);
  max-width: 640px;
  margin: 0 auto;
  line-height: 1.7;
}
.blog-listing-section .bl-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
  margin-bottom: 2.5rem;
}
.blog-listing-section .bl-filter-btn {
  padding: 0.5rem 1.25rem;
  border: 2px solid var(--color-border, #e5e7eb);
  border-radius: 999px;
  background: transparent;
  color: var(--color-text, #333);
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, color 0.2s, border-color 0.2s;
  min-height: 44px;
}
.blog-listing-section .bl-filter-btn:hover,
.blog-listing-section .bl-filter-btn.active {
  background: var(--color-primary, #2563eb);
  color: var(--color-primary-text, #fff);
  border-color: var(--color-primary, #2563eb);
}
.blog-listing-section .bl-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
}
.blog-listing-section .bl-card {
  background: var(--color-card-bg, #fff);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: var(--border-radius, 12px);
  overflow: hidden;
  box-shadow: var(--shadow, 0 2px 12px rgba(0,0,0,0.07));
  display: flex;
  flex-direction: column;
  transition: transform 0.2s, box-shadow 0.2s;
}
.blog-listing-section .bl-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 28px rgba(0,0,0,0.12);
}
.blog-listing-section .bl-card-img-wrap {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16/9;
}
.blog-listing-section .bl-card-img-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.35s;
}
.blog-listing-section .bl-card:hover .bl-card-img-wrap img {
  transform: scale(1.04);
}
.blog-listing-section .bl-card-category {
  position: absolute;
  top: 0.75rem;
  left: 0.75rem;
  background: var(--color-primary, #2563eb);
  color: var(--color-primary-text, #fff);
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 0.25rem 0.65rem;
  border-radius: 999px;
}
.blog-listing-section .bl-card-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  flex: 1;
}
.blog-listing-section .bl-card-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.8rem;
  color: var(--color-text-muted, #888);
  margin-bottom: 0.75rem;
}
.blog-listing-section .bl-card-date::before {
  content: '';
}
.blog-listing-section .bl-card-read-time {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}
.blog-listing-section .bl-card-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--color-heading, #111);
  margin: 0 0 0.6rem;
  line-height: 1.35;
}
.blog-listing-section .bl-card-excerpt {
  font-size: 0.9rem;
  color: var(--color-text-muted, #666);
  line-height: 1.65;
  flex: 1;
  margin-bottom: 1.25rem;
}
.blog-listing-section .bl-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
}
.blog-listing-section .bl-card-author {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--color-text, #333);
}
.blog-listing-section .bl-card-author img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}
.blog-listing-section .bl-read-link {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--color-primary, #2563eb);
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.3rem;
  transition: gap 0.2s;
}
.blog-listing-section .bl-read-link:hover {
  gap: 0.55rem;
}
.blog-listing-section .bl-read-link::after {
  content: '\2192';
}
.blog-listing-section .bl-cta-wrap {
  text-align: center;
  margin-top: 3rem;
}
.blog-listing-section .bl-cta-btn {
  display: inline-block;
  padding: 0.85rem 2.25rem;
  background: var(--color-primary, #2563eb);
  color: var(--color-primary-text, #fff);
  font-size: 1rem;
  font-weight: 700;
  border-radius: var(--border-radius, 8px);
  text-decoration: none;
  transition: background 0.2s, transform 0.15s;
  min-height: 44px;
  border: none;
  cursor: pointer;
}
.blog-listing-section .bl-cta-btn:hover {
  background: var(--color-primary-hover, #1d4ed8);
  transform: translateY(-2px);
}
@media (max-width: 900px) {
  .blog-listing-section .bl-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
@media (max-width: 768px) {
  .blog-listing-section .bl-grid {
    grid-template-columns: 1fr;
  }
  .blog-listing-section .bl-filters {
    gap: 0.4rem;
  }
}
</style>

<section class="blog-listing-section" data-component="blog-listing">
  <div class="bl-inner">
    <header class="bl-header">
      <span class="bl-eyebrow">{{.eyebrow_label}}</span>
      <h2 class="bl-heading">{{.section_heading}}</h2>
      <p class="bl-intro">{{.section_intro}}</p>
    </header>

    <div class="bl-grid" role="list">
      {{range .articles}}
      <article class="bl-card" role="listitem">
        {{if .image}}<div class="bl-card-img-wrap">
          <img src="{{.image}}" alt="{{.title}}" loading="lazy">
        </div>{{end}}
        <div class="bl-card-body">
          <h3 class="bl-card-title"><a href="{{.url}}">{{.title}}</a></h3>
          <p class="bl-card-excerpt">{{.meta_description}}</p>
          <div class="bl-card-footer">
            <a href="{{.url}}" class="bl-read-link" aria-label="{{$.read_more_label}} {{.title}}">{{$.read_more_label}}</a>
          </div>
        </div>
      </article>
      {{end}}
    </div>

    <div class="bl-cta-wrap">
      {{if .cta_url}}<a href="{{.cta_url}}" class="bl-cta-btn">{{.cta_label}}</a>{{end}}
    </div>
  </div>
</section>
$tmpl478$,
       updated_at    = now()
 WHERE name = 'blog-listing_pre_037';

DO $mig478$
DECLARE phantom int; hasq int; badtmpl int; hasrange int;
BEGIN
    SELECT count(*) INTO phantom
      FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
     WHERE cc.name = 'blog-listing_pre_037'
       AND f.value->>'source' LIKE 'site_specs.%';
    IF phantom <> 0 THEN
        RAISE EXCEPTION 'MIGRATION 478 VERIFY: % site_specs.* source(s) survive', phantom;
    END IF;

    SELECT count(*) INTO hasq FROM content_components
     WHERE name = 'blog-listing_pre_037'
       AND input_schema->'fields'->'articles'->>'source' = 'query.blog_posts';
    IF hasq <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 478 VERIFY: articles does not source query.blog_posts';
    END IF;

    SELECT count(*) INTO badtmpl FROM content_components
     WHERE name = 'blog-listing_pre_037' AND html_template LIKE '%post1_url%';
    IF badtmpl <> 0 THEN
        RAISE EXCEPTION 'MIGRATION 478 VERIFY: template still references post1_url';
    END IF;

    SELECT count(*) INTO hasrange FROM content_components
     WHERE name = 'blog-listing_pre_037' AND html_template LIKE '%range .articles%';
    IF hasrange <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 478 VERIFY: template has no range over .articles';
    END IF;

    RAISE NOTICE 'migration 478: verified -- 0 site_specs sources, articles<-query.blog_posts, template ranges';
END
$mig478$;

COMMIT;

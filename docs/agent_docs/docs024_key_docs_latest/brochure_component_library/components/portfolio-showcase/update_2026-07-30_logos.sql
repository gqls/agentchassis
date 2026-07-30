\set ON_ERROR_STOP on
-- Add an optional per-project logo (logo_url/logo_alt) to portfolio-showcase,
-- rendered above the title in a consistent white chip (object-fit:contain, a
-- fixed height regardless of each partner's native logo aspect ratio -- one
-- wide wordmark and two near-square marks all need to read at the same
-- visual weight). GENERATED from components/portfolio-showcase/
-- {template.html,input_schema.json} -- edit those files and regenerate.
--
-- portfolio-showcase is live on 1 site only (usage_count 0, 1 distinct site
-- via page_components) -- same low-risk profile already established this
-- session for editing a technically-shared-but-currently-single-site
-- template.
BEGIN;

DROP TABLE IF EXISTS bak_cc_portfolio_showcase_pre_logo_update;
CREATE TABLE bak_cc_portfolio_showcase_pre_logo_update AS
SELECT * FROM content_components WHERE function = 'portfolio-showcase';

UPDATE content_components
   SET html_template = $HTML$<section class="portfolio-showcase-section" data-component="portfolio-showcase">
    <div class="portfolio-container">
        <h2>{{.headline}}</h2>
        {{if .intro}}<p class="portfolio-intro">{{.intro}}</p>{{end}}
        <div class="portfolio-grid">
            {{range .projects}}
            <div class="portfolio-item">
                {{if .logo_url}}<div class="portfolio-logo"><img src="{{.logo_url}}" alt="{{.logo_alt}}" loading="lazy"></div>{{end}}
                <div class="portfolio-item-header">
                    <h3>{{.title}}</h3>
                    {{if .live_url}}<a href="{{.live_url}}" class="portfolio-link" target="_blank" rel="noopener">Visit Site &#8594;</a>{{end}}
                </div>
                {{if .domain}}<p class="portfolio-domain">{{.domain}}</p>{{end}}
                <p class="portfolio-description">{{.description}}</p>
                <div class="portfolio-meta">
                    {{if .built_with}}<span class="portfolio-tag">{{.built_with}}</span>{{end}}
                    {{if .build_time}}<span class="portfolio-tag portfolio-tag-time">{{.build_time}}</span>{{end}}
                </div>
            </div>
            {{end}}
        </div>
    </div>
</section>
<style>
/* Dark section to match the visual weight of the social-proof section it replaces.
   Painted with --color-cta-bg (a surface-role token) rather than --color-primary:
   primary is also the library's foreground colour for links and eyebrows, so on a
   dark site it is a LIGHT value and the hard-coded white text below vanished. Every
   ink here now derives from --color-cta-text, which is the pair --color-cta-bg was
   curated with. */
.portfolio-showcase-section {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-cta-bg, var(--color-primary, #1a1a2e));
    color: var(--color-cta-text, var(--color-white, #fff));
    --pf-ink: var(--color-cta-text, var(--color-white, #fff));
    --section-text: var(--pf-ink);
    --section-text-muted: color-mix(in srgb, var(--pf-ink) 78%, transparent);
    --section-heading: var(--pf-ink);
    --section-surface: color-mix(in srgb, var(--pf-ink) 10%, transparent);
    --section-border: color-mix(in srgb, var(--pf-ink) 30%, transparent);
}
.portfolio-container {
    max-width: var(--container-max-width, 1200px);
    margin: 0 auto;
}
.portfolio-showcase-section h2 {
    text-align: center;
    margin-bottom: 1rem;
    color: var(--pf-ink);
}
.portfolio-intro {
    text-align: center;
    max-width: 700px;
    margin: 0 auto 3rem;
    color: color-mix(in srgb, var(--pf-ink) 85%, transparent);
    line-height: 1.7;
}
.portfolio-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 2rem;
}
.portfolio-item {
    padding: 2rem;
    background: color-mix(in srgb, var(--pf-ink) 7%, transparent);
    border-radius: 8px;
    border-left: 3px solid var(--color-accent, #0f3460);
    transition: transform 0.2s, box-shadow 0.2s;
}
.portfolio-item:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0,0,0,0.2);
}
/* Each partner's own logo, at a consistent height regardless of its native
   aspect ratio (a wide wordmark and a near-square mark must sit at the same
   visual weight) -- object-fit:contain, never cover, so nothing crops. A
   white chip behind every logo is deliberate, not a fallback: these are real
   brand marks with their own background colour (cream, white, or none), and
   without a consistent light chip behind all three, one would float free on
   the dark card while the other two sat in a box, reading as inconsistent
   rather than as one honest treatment applied uniformly. */
.portfolio-logo {
    height: 2.5rem;
    margin-bottom: 1.25rem;
    display: flex;
    align-items: center;
}
.portfolio-logo img {
    height: 100%;
    max-width: 11rem;
    width: auto;
    object-fit: contain;
    background: #fff;
    border-radius: 6px;
    padding: 0.35rem 0.6rem;
    box-sizing: content-box;
}
.portfolio-item-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 0.5rem;
}
.portfolio-item h3 {
    margin: 0;
    font-size: 1.2rem;
    color: var(--pf-ink);
}
.portfolio-link {
    color: var(--color-accent, #4da6ff);
    text-decoration: none;
    font-size: 0.85rem;
    font-weight: 500;
    white-space: nowrap;
    transition: color 0.2s;
}
.portfolio-link:hover {
    color: var(--pf-ink);
}
.portfolio-domain {
    font-family: monospace;
    font-size: 0.85rem;
    color: color-mix(in srgb, var(--pf-ink) 72%, transparent);
    margin-bottom: 1rem;
}
.portfolio-description {
    color: color-mix(in srgb, var(--pf-ink) 92%, transparent);
    line-height: 1.7;
    margin-bottom: 1.5rem;
}
.portfolio-meta {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
}
.portfolio-tag {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    background: color-mix(in srgb, var(--pf-ink) 14%, transparent);
    border-radius: 4px;
    font-size: 0.8rem;
    color: color-mix(in srgb, var(--pf-ink) 88%, transparent);
}
.portfolio-tag-time {
    background: color-mix(in srgb, var(--color-accent, #4da6ff) 18%, transparent);
    color: var(--color-accent, #4da6ff);
}
@media (max-width: 768px) {
    .portfolio-showcase-section { padding: 3rem 1.5rem; }
    .portfolio-grid { grid-template-columns: 1fr; }
    .portfolio-item-header { flex-direction: column; gap: 0.5rem; }
}
</style>
$HTML$,
       input_schema  = $SCHEMA${
  "fields": {
    "intro": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field"
    },
    "headline": {
      "type": "text",
      "source": "llm",
      "required": false,
      "on_missing": "skip_field"
    },
    "projects": {
      "type": "array",
      "items": {
        "title": "string",
        "domain": "string",
        "live_url": "string",
        "build_time": "string",
        "built_with": "string",
        "description": "string",
        "logo_url": "string",
        "logo_alt": "string"
      },
      "source": "site_specs.portfolio.projects",
      "required": true,
      "min_items": 1,
      "on_missing": "needs_human_review",
      "missing_reason": "Real project data \u2014 titles, URLs, descriptions"
    }
  }
}$SCHEMA$::jsonb
 WHERE function = 'portfolio-showcase';

COMMIT;

SELECT function, is_active, length(html_template) AS template_bytes,
       html_template LIKE '%portfolio-logo%' AS has_logo_markup
  FROM content_components WHERE function = 'portfolio-showcase';

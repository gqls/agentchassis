-- migration_game_list_tier_d.sql
--
-- Rewrites game-list from the legacy numbered-flat anti-pattern to the Tier-D
-- items-array shape, mirroring the canonical tool-list (migration 041).
--
-- Why: game-list (`game-list_pre_037`) declared `game1_title…game6_*` as
-- `source: "llm"`, so page-content-writer INVENTED the games (approximate names,
-- duplicates — Jelly Invaders appeared twice), `gameN_cta_url`→`site_specs.games.gameN_url`
-- (absent → href="") and `gameN_image_url`→`site_assets.image` (→ src=""). It was
-- never migrated to Tier D. tool-list works because it is Tier D AND tools are
-- page_type=tool; games are now page_type=game on this site (the classifier gained
-- the `game` vocabulary since FOCUS_component_schema_patterns.md was written), so
-- `query.pages_where_type:game` resolves the 5 real game pages.
--
-- Decision (user): simplify to a clean tool-list-parity card list. The `pages`
-- table has only url/title/meta_description, so the old genre/rating/platform/
-- filter were fabricated with no real source; they are dropped. Real games, real
-- /games/<slug>/ links, no invention.
--
-- Field vocabulary is IDENTICAL to tool-list (items, cta_url, cta_label,
-- eyebrow_label, section_intro, card_link_label, section_heading,
-- cta_supporting_text) so the Step-3 content-writer + merge_with path treats it
-- exactly as it treats tool-list. Only the values differ:
--   items.source : query.pages_where_type:game   (tool-list: :tool)
--   cta_url      : site_specs.identity.games_index_url WITH fallback /games/index.html
--                  (tool-list has the site_specs source but NO fallback, which is
--                   why its "Browse All Tools" rendered href="" — the cta_url-dropped
--                   follow-up; the fallback fixes that class for game-list).
--
-- Schema confirmed: content_components(name UNIQUE, function ~ kebab-case,
-- html_template NOT NULL, input_schema jsonb, render_mode default 'template',
-- is_active). Section resolution (loadSectionComponents) is by name then function;
-- there is exactly one game-list row (name='game-list_pre_037', function='game-list'),
-- so we UPDATE it in place and leave the name to avoid touching references.
--
-- This is a DATA change (content_components) — it is live on COMMIT, no code deploy.
-- Pages already built with the old game-list (the games hub, and the homepage once
-- it rebuilds) must be re-rendered to pick it up — see the trailing rebuild step,
-- run AFTER this commits.
--
-- Quality metadata (quality_score, schema_field_count, template_variable_count,
-- schema_template_synced) is intentionally left untouched: render reads only
-- html_template/input_schema/render_mode, so stale metadata is cosmetic; a
-- re-validation pass (or component-creator's validator) refreshes it.

BEGIN;

-- Before: confirm we are targeting the legacy numbered-flat row.
SELECT name, function, is_active, render_mode,
       (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields')) AS field_count,
       input_schema->'fields'->'items'->>'source' AS items_source
FROM content_components
WHERE name = 'game-list_pre_037';

UPDATE content_components
SET html_template = $TMPL$<style>
.game-list-section { padding: var(--spacing-section, 5rem 2rem); background: var(--color-background); color: var(--color-text); }
.game-list-section .gl-inner { max-width: var(--container-max-width, 1200px); margin: 0 auto; }
.game-list-section .gl-header { text-align: center; margin-bottom: 3rem; }
.game-list-section .gl-eyebrow { display: inline-block; font-size: 0.8rem; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; color: var(--color-primary); margin-bottom: 0.75rem; }
.game-list-section .gl-heading { font-size: clamp(1.75rem, 4vw, 2.75rem); font-weight: 800; color: var(--color-heading); margin: 0 0 1rem; line-height: 1.2; }
.game-list-section .gl-intro { font-size: 1.1rem; color: var(--color-text-muted); max-width: 640px; margin: 0 auto; line-height: 1.7; }
.game-list-section .gl-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1.5rem; margin-bottom: 3rem; }
.game-list-section .gl-card { background: var(--color-card-bg, var(--color-surface)); border: 1px solid var(--color-border); border-radius: var(--border-radius, 0.5rem); padding: 1.75rem; display: flex; flex-direction: column; gap: 0.75rem; box-shadow: var(--shadow, 0 2px 8px rgba(0,0,0,0.06)); transition: box-shadow 0.2s ease, transform 0.2s ease; }
.game-list-section .gl-card:hover { box-shadow: 0 6px 24px rgba(0,0,0,0.12); transform: translateY(-2px); }
.game-list-section .gl-card-icon { width: 2.5rem; height: 2.5rem; background: var(--color-primary); border-radius: calc(var(--border-radius, 0.5rem) * 0.75); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.game-list-section .gl-card-icon svg { width: 1.25rem; height: 1.25rem; fill: var(--color-primary-text, #fff); }
.game-list-section .gl-card-title { font-size: 1.1rem; font-weight: 700; color: var(--color-heading); margin: 0; line-height: 1.3; }
.game-list-section .gl-card-desc { font-size: 0.92rem; color: var(--color-text-muted); margin: 0; line-height: 1.6; flex: 1; }
.game-list-section .gl-card-link { display: inline-flex; align-items: center; gap: 0.35rem; font-size: 0.9rem; font-weight: 600; color: var(--color-primary); text-decoration: none; margin-top: 0.5rem; min-height: 44px; padding: 0.5rem 0; transition: color 0.15s ease; }
.game-list-section .gl-card-link:hover { color: var(--color-primary-hover, var(--color-primary)); text-decoration: underline; }
.game-list-section .gl-card-link svg { width: 1rem; height: 1rem; flex-shrink: 0; }
.game-list-section .gl-cta { text-align: center; }
.game-list-section .gl-cta-text { font-size: 1rem; color: var(--color-text-muted); margin: 0 0 1.25rem; }
.game-list-section .gl-cta-btn { display: inline-flex; align-items: center; justify-content: center; min-height: 44px; padding: 0.75rem 2rem; background: var(--color-primary); color: var(--color-primary-text, #fff); font-size: 1rem; font-weight: 700; border-radius: var(--border-radius, 0.5rem); text-decoration: none; transition: background 0.2s ease; }
.game-list-section .gl-cta-btn:hover { background: var(--color-primary-hover, var(--color-primary)); }
@media (max-width: 768px) {
  .game-list-section .gl-grid { grid-template-columns: 1fr; }
  .game-list-section .gl-header { margin-bottom: 2rem; }
}
</style>

<section class="game-list-section" data-component="game-list">
  <div class="gl-inner">
    <header class="gl-header">
      <span class="gl-eyebrow">{{.eyebrow_label}}</span>
      <h2 class="gl-heading">{{.section_heading}}</h2>
      <p class="gl-intro">{{.section_intro}}</p>
    </header>

    <div class="gl-grid">
      {{range .items}}
      <article class="gl-card">
        <div class="gl-card-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M8 5v14l11-7z"/></svg>
        </div>
        <h3 class="gl-card-title">{{.title}}</h3>
        <p class="gl-card-desc">{{.meta_description}}</p>
        <a class="gl-card-link" href="{{.url}}">
          {{$.card_link_label}}
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" xmlns="http://www.w3.org/2000/svg"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </a>
      </article>
      {{end}}
    </div>

    <div class="gl-cta">
      <p class="gl-cta-text">{{.cta_supporting_text}}</p>
      <a class="gl-cta-btn" href="{{.cta_url}}">{{.cta_label}}</a>
    </div>
  </div>
</section>$TMPL$,
    input_schema = $SCHEMA$
{
  "fields": {
    "items": {
      "type": "array",
      "items": {
        "url": { "type": "url" },
        "title": { "type": "text" },
        "nav_label": { "type": "text" },
        "meta_description": { "type": "text" }
      },
      "limit": 6,
      "source": "query.pages_where_type:game",
      "required": true,
      "min_items": 1
    },
    "cta_url": {
      "type": "url",
      "source": "site_specs.identity.games_index_url",
      "fallback": "/games/index.html",
      "required": false
    },
    "cta_label": {
      "type": "text",
      "source": "static",
      "fallback": "Browse All Games",
      "required": false,
      "llm_guidance": "Label for the primary call-to-action button."
    },
    "eyebrow_label": {
      "type": "text",
      "source": "static",
      "fallback": "Our Games",
      "required": false,
      "llm_guidance": "Short uppercase eyebrow label above the section heading."
    },
    "section_intro": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "One to two sentence introduction below the heading. Explain what kinds of games or interactive prototypes are listed and why they are worth playing. 20-40 words."
    },
    "card_link_label": {
      "type": "text",
      "source": "static",
      "fallback": "Play now",
      "required": false,
      "llm_guidance": "Label for the link on each game card. Override if site tone differs."
    },
    "section_heading": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Primary heading for the game list section. Should communicate the value of the games available. 6-12 words."
    },
    "cta_supporting_text": {
      "type": "text",
      "source": "llm",
      "required": true,
      "llm_guidance": "Short sentence below the game grid encouraging visitors to explore all games. 10-20 words."
    }
  }
}
$SCHEMA$::jsonb,
    render_mode = 'template',
    updated_at = NOW()
WHERE name = 'game-list_pre_037';

-- After: confirm the rewrite took (items_source should be query.pages_where_type:game,
-- field_count should drop from 50+ to 8).
SELECT name, function, render_mode, updated_at,
       (SELECT count(*) FROM jsonb_object_keys(input_schema->'fields')) AS field_count,
       input_schema->'fields'->'items'->>'source' AS items_source,
       input_schema->'fields'->'cta_url'->>'fallback' AS cta_fallback
FROM content_components
WHERE name = 'game-list_pre_037';

COMMIT;

-- ============================================================================
-- RUN SEPARATELY, AFTER THE ABOVE COMMITS (the component is live on commit).
-- Re-render pages that use game-list so plan_sections re-resolves with the new
-- Tier-D component. `sections` is the page's jsonb array of section functions.
-- (The homepage may already be rebuilding from the Group 2 unstick; marking it
-- again is idempotent.)
--
-- UPDATE pages
-- SET build_status = 'needs_rebuild', built_from_plan_version = NULL
-- WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--   AND sections @> '["game-list"]'::jsonb;
--
-- Verify which pages that targets first:
-- SELECT name, url, build_status FROM pages
-- WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
--   AND sections @> '["game-list"]'::jsonb;
-- ============================================================================

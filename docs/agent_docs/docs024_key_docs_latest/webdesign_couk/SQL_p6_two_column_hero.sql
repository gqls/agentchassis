-- SQL_p6_two_column_hero.sql — webdesign.co.uk, phase 6
--
-- Replace the home page's hero with the two-column one the brief actually asked
-- for.
--
-- WHAT WAS WRONG. The planner selected the library's generic `hero`, which
-- paints `linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.6))` over a full-bleed
-- background image and sets `--hero-ink: #fff`. That is a dark hero. It
-- contradicts two things at once:
--   * the mission brief — "a short two-column hero (a sentence of copy on the
--     left, an image on the right)";
--   * design_intent.avoid — "Dark backgrounds of any kind — no dark mode, no
--     dark hero sections".
--
-- WHY THE PIN DID NOT PREVENT IT. Worth being precise, because the instinct is
-- to blame the palette pin. The pin governs COLOUR VALUES and it held perfectly:
-- every colour in the committed styles.css is the pinned one. What it cannot
-- govern is which COMPONENT the planner picks, and the darkness here is baked
-- into that component's own template as a literal rgba() overlay, not drawn from
-- the palette at all. A design_intent avoid-list entry is prose; the planner's
-- component choice does not consult it. So: right palette, wrong furniture.
--
-- THE FIX is a per-site component, the same approach as the chrome forks, rather
-- than editing the shared `hero` — six other sites use that row.
--
-- Layout reproduces websitedesign.com's own homepage hero, which is what the
-- owner pointed at: 1fr/1fr grid with 4rem gap, copy left, a 4:3 image right in
-- a 12px-radius container with the large soft shadow, collapsing to one column
-- and 16:9 below 900px.
--
-- Every optional element is gated. An ungated CTA is how a chrome slot ends up
-- rendering href="" (bugs_open/049); here the image, each button and the
-- eyebrow all disappear cleanly if their data is absent.

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO content_components (name, function, section_type, component_level, render_mode, html_template, input_schema, is_active, created_at, updated_at)
VALUES (
  'webdesign.co.uk Two-Column Hero',
  'webdesign-couk-hero',
  'hero',
  'section',
  'template',
  $hero$<style>
.wd-hero {
  background: var(--bg, #f9f8f6);
  padding: 4rem 2rem 3.5rem;
}
.wd-hero-inner {
  max-width: 1200px; margin: 0 auto;
  display: grid; grid-template-columns: 1fr 1fr; gap: 4rem; align-items: center;
}
.wd-hero-copy h1 {
  font-family: 'Inter', system-ui, sans-serif;
  font-size: 3.5rem; font-weight: 800; line-height: 1.1; letter-spacing: -1.5px;
  color: var(--text, #2b2b2b); margin: 0 0 1.25rem;
}
.wd-hero-eyebrow {
  display: inline-block; font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 0.75rem; letter-spacing: 0.08em; text-transform: uppercase;
  color: var(--primary, #5c6b5d); margin-bottom: 0.9rem;
}
.wd-hero-copy p {
  font-size: 1.2rem; line-height: 1.6; color: var(--text-dim, #717171);
  margin: 0 0 2rem; max-width: 46ch;
}
.wd-hero-actions { display: flex; gap: 1rem; flex-wrap: wrap; }
.wd-hero-btn {
  display: inline-block; padding: 0.85rem 1.6rem; border-radius: 8px;
  font-weight: 600; font-size: 0.98rem; text-decoration: none;
  transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease;
}
.wd-hero-btn.primary {
  background: var(--text, #2b2b2b); color: #ffffff; border: 1px solid var(--text, #2b2b2b);
}
.wd-hero-btn.primary:hover { background: var(--primary, #5c6b5d); border-color: var(--primary, #5c6b5d); }
.wd-hero-btn.secondary {
  background: transparent; color: var(--text, #2b2b2b);
  border: 1px solid var(--border-hover, #dbd9d4);
}
.wd-hero-btn.secondary:hover { border-color: var(--primary, #5c6b5d); color: var(--primary, #5c6b5d); }
.wd-hero-media {
  aspect-ratio: 4 / 3; border-radius: 12px; overflow: hidden;
  box-shadow: 0 12px 32px rgba(43,43,43,0.08);
  background: var(--surface, #f3f1ec);
}
.wd-hero-media img { width: 100%; height: 100%; object-fit: cover; display: block; }
@media (max-width: 900px) {
  .wd-hero { padding: 3rem 1.5rem 2.5rem; }
  .wd-hero-inner { grid-template-columns: 1fr; gap: 2.5rem; }
  .wd-hero-copy h1 { font-size: 2.8rem; letter-spacing: -1px; }
  .wd-hero-media { aspect-ratio: 16 / 9; }
}
</style>
<section class="wd-hero" data-component="webdesign-couk-hero">
  <div class="wd-hero-inner">
    <div class="wd-hero-copy">
      {{if .eyebrow}}<span class="wd-hero-eyebrow">{{.eyebrow}}</span>{{end}}
      <h1>{{.headline}}</h1>
      {{if .subheadline}}<p>{{.subheadline}}</p>{{end}}
      {{if or .cta_url .secondary_cta_url}}
      <div class="wd-hero-actions">
        {{if and .cta_url .cta_text}}<a class="wd-hero-btn primary" href="{{.cta_url}}">{{.cta_text}}</a>{{end}}
        {{if and .secondary_cta_url .secondary_cta}}<a class="wd-hero-btn secondary" href="{{.secondary_cta_url}}">{{.secondary_cta}}</a>{{end}}
      </div>
      {{end}}
    </div>
    {{if .hero_url}}<div class="wd-hero-media"><img src="{{.hero_url}}" alt="{{.hero_alt}}" loading="eager"></div>{{end}}
  </div>
</section>$hero$,
  $schema${
    "fields": {
      "headline":         { "type": "string", "source": "llm", "required": true },
      "subheadline":      { "type": "string", "source": "llm", "required": false },
      "eyebrow":          { "type": "string", "source": "llm", "required": false },
      "cta_text":         { "type": "string", "source": "llm", "required": false },
      "cta_url":          { "type": "string", "source": "llm", "required": false },
      "secondary_cta":    { "type": "string", "source": "llm", "required": false },
      "secondary_cta_url":{ "type": "string", "source": "llm", "required": false },
      "hero_url":         { "type": "image",  "source": "llm", "required": false },
      "hero_alt":         { "type": "string", "source": "llm", "required": false }
    }
  }$schema$::jsonb,
  true, NOW(), NOW()
);

-- Repoint the home page's hero instance and give the CTAs real destinations.
-- The old instance had cta_text and secondary_cta but no URLs at all, so the
-- previous template gated both buttons away and the hero rendered with an empty
-- action row. Both targets below are pages that exist and are live.
UPDATE page_components pc
   SET component_id = c.id,
       content_data = pc.content_data
                      || jsonb_build_object(
                           'eyebrow',           'Free · no signup · runs in your browser',
                           'cta_url',           '/tools/index.html',
                           'secondary_cta_url', '/learn/index.html',
                           'hero_alt',          'Warm, minimal workspace'
                         ),
       build_status = 'approved',
       updated_at = NOW()
  FROM pages p, sites s, content_components c
 WHERE pc.page_id = p.id
   AND p.site_id = s.id
   AND s.domain = 'webdesign.co.uk'
   AND p.name = 'index'
   AND pc.slot_name = 'hero'
   AND c.function = 'webdesign-couk-hero'
   AND c.is_active
   AND (pc.locked_at IS NULL
        OR (pc.lock_type = 'timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()));

DO $verify$
DECLARE v_site uuid; v_fn text; v_cta text; v_img text;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';

    SELECT c.function, pc.content_data->>'cta_url', pc.content_data->>'hero_url'
      INTO v_fn, v_cta, v_img
      FROM page_components pc
      JOIN pages p ON p.id = pc.page_id
      JOIN content_components c ON c.id = pc.component_id
     WHERE p.site_id = v_site AND p.name = 'index' AND pc.slot_name = 'hero';

    IF v_fn IS DISTINCT FROM 'webdesign-couk-hero' THEN
        RAISE EXCEPTION 'hero instance still points at % (expected webdesign-couk-hero)', v_fn;
    END IF;
    IF v_cta IS NULL THEN
        RAISE EXCEPTION 'hero cta_url missing — the buttons would gate away again';
    END IF;
    IF v_img IS NULL THEN
        RAISE WARNING 'hero has no image; it will render as a single column';
    END IF;

    -- The whole point: no dark overlay anywhere in the new template.
    IF EXISTS (
        SELECT 1 FROM content_components
         WHERE function = 'webdesign-couk-hero' AND is_active
           AND (html_template ILIKE '%rgba(0,0,0%' OR html_template ILIKE '%hero-ink%')
    ) THEN
        RAISE EXCEPTION 'the new hero still contains a dark overlay';
    END IF;

    RAISE NOTICE 'two-column hero installed and bound to the home page';
END
$verify$;

COMMIT;

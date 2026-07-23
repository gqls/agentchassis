\set ON_ERROR_STOP on
INSERT INTO content_components
  (id, name, function, display_name, description, category, semantic_tags,
   section_type, component_level, render_mode, is_dark_section, is_active,
   suitable_site_types, suitable_page_types, html_template, input_schema)
VALUES (
  gen_random_uuid(),
  'people-feature-block','people-feature-block','People Feature Block',
  'A two-column feature block pairing a line-illustration figure with a statement about an approach or way of working (optional real pull-quote). Reversible layout. Never a named/invented person or fabricated bio. CSS-only.',
  'feature','["feature","people","illustration","brochure"]'::jsonb,
  'people-feature','section','agent',false,true,
  '["brochure","consultancy","professional-services","b2b"]'::jsonb,
  '["index","home","about","capabilities","landing"]'::jsonb,
  $HTML$
<style>
  .people-feature-block {
    padding: var(--spacing-section, 5rem 2rem);
    background: var(--color-background);
    color: var(--color-text);
  }
  .people-feature-block__inner {
    max-width: var(--container-max-width, 1100px);
    margin: 0 auto;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: clamp(2rem, 5vw, 4rem);
    align-items: center;
  }
  .people-feature-block--reverse .people-feature-block__media { order: 2; }

  .people-feature-block__media {
    border-radius: var(--border-radius, 1rem);
    overflow: hidden;
    aspect-ratio: 4 / 3;
    background: var(--color-surface-alt, var(--color-surface));
    border: 1px solid var(--color-hairline, var(--color-border));
  }
  .people-feature-block__img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .people-feature-block__eyebrow {
    display: inline-block;
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--color-primary);
    margin-bottom: 0.75rem;
  }
  .people-feature-block__title {
    font-size: clamp(1.6rem, 3vw, 2.4rem);
    font-weight: 700;
    line-height: 1.2;
    color: var(--color-heading);
    margin: 0 0 1rem;
  }
  .people-feature-block__body {
    font-size: 1.0625rem;
    line-height: 1.7;
    color: var(--color-text);
    margin: 0 0 1.25rem;
  }
  .people-feature-block__quote {
    margin: 0 0 1.25rem;
    padding-inline-start: 1.1rem;
    border-inline-start: 3px solid var(--color-primary);
    font-size: 1.125rem;
    font-style: italic;
    line-height: 1.5;
    color: var(--color-heading);
  }
  .people-feature-block__attribution {
    display: block;
    margin-top: 0.5rem;
    font-size: 0.875rem;
    font-style: normal;
    font-weight: 600;
    color: var(--color-text-muted);
  }
  .people-feature-block__link {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-primary);
    text-decoration: none;
    min-height: 44px;
  }
  .people-feature-block__link:hover { text-decoration: underline; }
  .people-feature-block__link:focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
    border-radius: 2px;
  }

  @media (max-width: 820px) {
    .people-feature-block__inner { grid-template-columns: 1fr; }
    .people-feature-block--reverse .people-feature-block__media { order: 0; }
    .people-feature-block { padding: 3rem 1.25rem; }
  }
</style>

<section class="people-feature-block{{if .reverse}} people-feature-block--reverse{{end}}" data-component="people-feature-block">
  <div class="people-feature-block__inner">
    <div class="people-feature-block__media">
      <img class="people-feature-block__img" src="{{if .image}}{{.image}}{{else}}/assets/images/hero.jpg{{end}}" alt="{{.image_alt}}" loading="lazy" width="720" height="540">
    </div>
    <div class="people-feature-block__content">
      {{if .eyebrow}}<span class="people-feature-block__eyebrow">{{.eyebrow}}</span>{{end}}
      {{if .title}}<h2 class="people-feature-block__title">{{.title}}</h2>{{end}}
      {{if .body}}<p class="people-feature-block__body">{{.body}}</p>{{end}}
      {{if .quote}}<blockquote class="people-feature-block__quote">{{.quote}}{{if .attribution}}<cite class="people-feature-block__attribution">{{.attribution}}</cite>{{end}}</blockquote>{{end}}
      {{if .link_url}}<a class="people-feature-block__link" href="{{.link_url}}">{{if .link_label}}{{.link_label}}{{else}}Read more{{end}}<span aria-hidden="true">&nbsp;&rarr;</span></a>{{end}}
    </div>
  </div>
</section>
$HTML$,
  $SCHEMA${
  "fields": {
    "eyebrow": { "type": "text", "source": "llm", "required": false, "llm_guidance": "Short uppercase eyebrow, e.g. 'How we work'. Optional." },
    "title": { "type": "text", "source": "llm", "required": true, "llm_guidance": "Heading for this feature block. One phrase, under 10 words." },
    "body": { "type": "text", "source": "llm", "required": true, "llm_guidance": "One or two short paragraphs about an approach, value, or way of working — NOT a person's biography. Max 55 words." },
    "quote": { "type": "text", "source": "llm", "required": false, "llm_guidance": "An optional pull-quote. Use ONLY a real, attributable quote (or omit). Never invent a testimonial or attribute words to a person who did not say them." },
    "attribution": { "type": "text", "source": "llm", "required": false, "llm_guidance": "Source of the quote, if any. Never invent a name or a client." },
    "image": { "type": "image" },
    "image_alt": { "type": "text", "source": "llm", "required": true, "llm_guidance": "Describe the line illustration for screen readers, e.g. 'A line illustration of two people reviewing a document'. The figure is a GENERIC line illustration — never a named or specific real/invented individual." },
    "link_url": { "type": "url", "source": "llm", "required": false, "llm_guidance": "Optional link URL." },
    "link_label": { "type": "text", "source": "llm", "required": false, "llm_guidance": "Optional link text." },
    "reverse": { "type": "boolean", "source": "llm", "required": false, "llm_guidance": "Set true to place the illustration on the right instead of the left (for visual rhythm when several blocks stack). Default false." }
  }
}
$SCHEMA$::jsonb
)
RETURNING function, section_type, is_active;

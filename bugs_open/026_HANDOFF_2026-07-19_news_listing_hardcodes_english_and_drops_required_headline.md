# 026 — shared `news-listing` component hardcodes English and renders an empty required `<h1>`

**Filed 2026-07-19** (relojistas thread). **Status: OPEN.** Fleet-wide — one shared library
component, no `site_id`. Two distinct defects in the same template; grouped because they
are one fix pass.

Live evidence throughout is `https://relojistas.com/noticias/`, a Spanish-language site,
which is simply the case that exposes them.

## Defect A — hardcoded English placeholder on non-English sites

`content_components` id `11d4dc21-1ccc-40ef-93bc-b9e26bd95e9f`, `function='news-listing'`,
`render_mode='template'`. The template contains:

```html
<div class="news-listing-items" id="news-listing-items">
  <p class="news-listing-loading">Loading latest news...</p>
</div>
```

That string is **not** a translatable field — it is literal English in the shared template.
It is the first thing every visitor sees, because the items are filled client-side by
`/tools/assets/news-listing.js` (which fetches `/data/news-archive.json`) *after* paint. On
a Spanish site the visible text before fill is English.

It is also what a non-JS client sees **permanently**: crawlers that do not execute JS get
"Loading latest news..." and zero news items on the site's main news page.

Verify:
```sql
SELECT html_template ~ 'Loading latest news' FROM content_components WHERE function='news-listing';
-- t
```

## Defect B — a `required` field renders empty and nothing catches it

Same template: `<h1 class="news-listing-title">{{.headline}}</h1>`, and the component's
`input_schema` declares:

```json
{"type":"object","required":["headline"],
 "properties":{"headline":{"type":"string","source":"llm","description":"Page headline for the news listing"}}}
```

`headline` is **required** and LLM-sourced. On the live page it renders as
`<h1 class="news-listing-title"></h1>` — empty. So a required input was never populated,
the section still saved, deployed, and serves. Whatever enforces `required` either does not
run for this path or does not treat empty-string as missing.

Consequences: the page carries **two** `<h1>`s (the hero's real one plus this empty one),
which is an SEO/a11y defect on its own, and the component's own header block is blank.

Verify on the live page:
```bash
curl -s https://relojistas.com/noticias/ | grep -oE '<h1[^>]*>[^<]*</h1>'
# <h1>Noticias de relojería en español</h1>
# <h1 class="news-listing-title"></h1>
```

## Why these matter beyond one site

The component is in the shared library with no `site_id`, so **every** site using
`news-listing` inherits both. Defect A is latent on English sites and visible on every
non-English one. Defect B is invisible until someone reads the markup — the page looks fine
because the hero supplies a heading.

Note the interaction with discovery: `check_empty_sections` flagged this very section as
`empty_section` on `noticias-index`. That finding is a **false positive for emptiness** —
the section is a runtime-fill template and the data does arrive — but it was the thread that
led here. Treating it as a simple false positive and dismissing it would have buried both
real defects.

## Fix candidates

**Defect A.** Promote the placeholder to a template field with a sensible default, e.g.
`{{if .loading_text}}{{.loading_text}}{{else}}Loading latest news...{{end}}`, and add
`loading_text` to `input_schema` as an optional LLM-sourced string so the writer produces it
in the site's own language. Better still, since the string is only visible pre-fill, have
`news-listing.js` own it and render nothing server-side — but that worsens the no-JS case,
so prefer the field.

Consider the same sweep for sibling components: check any other shared template for literal
English (`grep` `html_template` across `content_components` for common UI strings).

**Defect B.** Two parts, and the second is the important one:
1. Populate `headline` for this page.
2. **Find why a `required` field passed validation empty.** If `validate_page_content` (or
   whatever enforces `input_schema.required`) treats `""` as present, that is the real bug
   and it is not specific to this component. Fixing only (1) papers over it.

## How to verify a fix

- Defect A: the served `/noticias/` markup contains no English placeholder; a Spanish string
  appears instead. Check the **rendered page**, not the component row.
- Defect B: exactly one `<h1>` on the page, non-empty; and — the real check — a component
  with a deliberately blank required field is *rejected* at build rather than saved.

## Related

- `bugs_open/015` — mistyped `page_type` orphans a page from its machinery. Same site, same
  "the page looks fine, the machinery disagrees" family.
- The runtime-fill class is a known landmine: a template whose content arrives client-side
  reads as empty to any server-side check. Recorded in the vonc/link-integrity notes.

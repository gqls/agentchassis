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

---

## Addendum 2026-07-21 (bugfix-026 thread) — DIAGNOSIS complete; Defect A + B-part-1 fixed live; B-part-2 root cause found and is a two-layer fail-open

Re-grounded the whole case against live prod (chassis `v1.0.1144`, pod
`agent-chassis-59c675c4f-pxr9f`) and HEAD. The picture moved a long way since filing.

### Defect A — FIXED LIVE (by the relojistas/027 thread's seed 179, not this thread)
The shared template now server-renders items from `query.news_archive` and the placeholder
was promoted to an optional LLM field:
`{{if .loading_text}}{{.loading_text}}{{else}}Loading latest news...{{end}}`, with
`loading_text` declared `source:"llm"` "in the site's own language". Verified from `curl`
(JS never executing): relojistas/gaswholesalers/robot-hands all serve **0** loading
placeholders and 150+ **server-rendered** `<article>` items; relojistas' visible copy is
Spanish. The English literal survives only as an unreachable `{{else}}` fallback. Nothing
to do here — it is the 027 rework.

### Defect B part 1 (populate the headline) — FIXED LIVE on every actively-served news page
`<h1 class="news-listing-title">` is now non-empty on all three served news-index pages
(relojistas "Últimas noticias de relojería", gaswholesalers, robot-hands). Two residuals
remain in `content_data` but **neither is this component's defect** — both are the
`bugs_open/015` mistyped-`page_type` class:
- `idea.uk` `/news/index.html` — `headline=''`, but `page_type='section-index'` (not
  `news-index`) and the page **404s** (idea.uk is not cut over). The news renderer only
  touches `page_type='news-index'` (`render_news_section_action.go:215,282`), so this page
  was never re-rendered under the fix and keeps stale empty `content_data`.
- `ai-agent-orchestration.com/news.html` — `page_type='content'`, an old (2026-05-01)
  page with 8 static articles and the English placeholder still in its HTML; headline is
  actually non-empty. English site, so Defect A is invisible there.
  → **Handed to `bugs_open/015`'s owner** — re-typing those two pages to `news-index` and
  re-rendering is the fix, and it belongs with 015, not here.

### Defect B part 2 (THE STRUCTURAL ONE) — root cause found; it is a TWO-LAYER fail-open on schema shape
The bug file framed this as "validation treats `''` as present." The real mechanism is
sharper and it is **not** in a validator — **it is that a component whose `input_schema` is
not in the v2 `fields` shape is invisible to BOTH the generation planner and the render-time
required-field gate.** When 026 was filed, `news-listing`'s schema was the old JSON-Schema
shape `{"type":"object","required":["headline"],"properties":{…}}` — which has no `fields`
key. Two independent readers only parse `fields`:

1. **Generation** — `planSection` (`plan_sections_action.go:1182`) does
   `comp.InputSchema["fields"].(map[string]interface{})`. On a miss it falls through
   (:1201-1225); the name-keyword backstop at :1206-1213 only defers functions containing
   `article/content/body/text/blog` — **`news-listing` matches none** — so it returns
   `"no field schema — all fields from LLM"` with an **empty `llmFieldSpecs`**. The writer is
   therefore never told the component even has a `headline` field, let alone a required one.
   The headline is never generated → empty.

2. **Enforcement** — the render gate `missingRequiredLLMFields`
   (`json_envelope.go:192`, called at `v3_site_actions.go:1631` "refusing to render an empty
   section" and `rerender_page_sections_action.go:210` "escalating page to writer") also reads
   only `inputSchema["fields"]`. Old shape → `nil` → **zero enforcement** → the empty headline
   renders, saves, deploys, serves.

So both "why was it empty" and "why did it ship empty" trace to the **same** root: only the
v2 `fields` dialect is understood; the old `properties` dialect is silently unread. The
relojistas thread reached the same conclusion in seed 179's own header comment
(`docs/agent_docs/sql_for_agents/179_news_components_query_sourced_items.sql:5-8`) and
**explicitly left this half open as 026's**: *"the structural half of 026 — a required field
rendering empty and saving anyway — stays open and is 026's."*

**Why it is not currently firing (verified):** seed 179 converted `news-listing` to the v2
`fields` shape, and the old `properties` shape is now **extinct fleet-wide**:
```sql
SELECT count(*) FILTER (WHERE input_schema ? 'fields')                                AS v2,      -- 124
       count(*) FILTER (WHERE input_schema ? 'properties' AND NOT (input_schema ? 'fields')) AS old,  -- 0
       count(*) FILTER (WHERE input_schema::text IN ('{}','null'))                    AS empty,   -- 42
       count(*)                                                                        AS total    -- 173
FROM content_components;
```
The gate is confirmed **live** in the pod (both marker strings grep non-zero in
`/app/agent-chassis` on `v1.0.1144`), so every v2-shape component (the whole fleet) is now
enforced. The 7 non-`fields`/non-`properties` rows are the legacy core sections
(hero/header/footer/cta/features/social-proof) whose schemas are bare example-value maps
declaring **no** requiredness — nothing for the gate to enforce, correctly a no-op.

**The residual = the fail-open itself.** Both readers still fall open on ANY non-v2
non-empty shape, backstopped only by a name-keyword heuristic that misses most functions.
The exact triggering shape is extinct, so this is **defensive**: it re-arms only if the old
shape returns via a config re-seed, a restored snapshot, or a `component-creator` run that
emits JSON-Schema instead of the `fields` dialect — all live risks in this tree.

### Status
- Defect A: **CLOSEABLE** (fixed live via 027/seed 179).
- Defect B part 1: **CLOSEABLE** on served pages; residuals routed to `bugs_open/015`.
- Defect B part 2: **root-caused**, not currently reproducible; the decision is whether to
  build the fail-open hardening (below) before closing, or close on the extinct-shape
  argument and record the pattern only.

### Fix candidate for the fail-open (if built)
Make both readers refuse-or-defer instead of silently proceeding when `input_schema` is
non-empty **and** carries a structured-schema signal it cannot read — i.e. has `properties`
or a top-level `required` array but no `fields`. That predicate excludes the 42 empty and 7
legacy example-value rows (safe) and fires only on the old JSON-Schema dialect (the 026
shape). Cleanest structural form: teach a shared helper to normalise `properties`+`required[]`
→ the `fields` view both callers already consume, so an old-shape component becomes fully
understood (generation asks for its fields; enforcement checks them) rather than merely
rejected. Touches `json_envelope.go` (article-body workstream's file — coordinate) and
`plan_sections_action.go`; Go, so inert until an image roll; worth a council-gate pass.

### Transferable pattern (also filed to 016b §9)
A reader that understands only one schema/format dialect and returns "nothing found" on every
other dialect does not fail safe — it **fails open**: downstream treats "couldn't read the
contract" as "there is no contract." Two such readers over one field (plan + gate) both
fell open here, so a *required* field was neither requested nor enforced. When a shape flips
formats, grep every consumer of the old shape before assuming the flip is inert.

# 041 — A site component's JS is never published: chrome references an asset that always 404s

**Filed:** 2026-07-20 · idea.uk vm site thread · **Status:** ✅ **CLOSED — FIXED & LIVE & VERIFIED 2026-07-21.**
**Severity:** medium-high — silently breaks any interactive behaviour in header/footer fleet-wide.
Today it means **the mobile menu is dead on every page of idea.uk**; the hamburger renders, has
`aria-expanded`, and does nothing.
**Class:** structural — the asset pipeline models pages and does not model chrome.

> **RESOLUTION 2026-07-21 — fix candidate 1 (the `collectJSAssets` UNION), live on v1.0.1146.**
> Shipped in commit `36829b07b` (bundled with the `bugs_open/018` platform fix). `collectJSAssets`
> now UNIONs `site_components`, mirroring `render_js_snippets_for_site_action.go:203-219`, so chrome
> `js_content` publishes. Verified **end-to-end** on idea.uk after a safe assemble-only rerender:
> - `curl https://idea.uk/tools/assets/site-header.js` → **200** (708 B, the real `.hamburger`
>   click handler toggling `aria-expanded`); `…/site-footer.js` → **200** (was 404 for both).
> - The homepage `<script src>`-references both assets, and still renders **0 empty hrefs** (no
>   chrome regression from the rerender).
> Binary confirmed in-pod (v1.0.1146) by discriminating grep alongside the 018 symbols. The mobile
> menu is now functional. Other fleet sites publish their chrome JS on their next rerender (the code
> is live fleet-wide); idea.uk was the only site carrying chrome `js_content`, so it was the test.

---

## Symptom

The chrome templates emit `<script src="/tools/assets/site-header.js">` and
`…/site-footer.js`. Both 404. A *page* component's asset on the same site, same path prefix, serves
fine — which is the whole shape of the bug:

```
$ curl -o /dev/null -w '%{http_code}' https://idea.uk/tools/assets/site-header.js    → 404
$ curl -o /dev/null -w '%{http_code}' https://idea.uk/tools/assets/site-footer.js    → 404
$ curl -o /dev/null -w '%{http_code}' https://idea.uk/tools/assets/latest-news.js    → 200
$ curl -o /dev/null -w '%{http_code}' https://idea.uk/tools/assets/audience-check-form.js → 200
```

The JS is not missing from the database. Both chrome components carry real, non-empty `js_content`
**and** reference the asset from their template:

```
 domain  |    name     |  function   | js_len | refs_asset
---------+-------------+-------------+--------+-----------
 idea.uk | site-footer | site-footer |    628 | t
 idea.uk | site-header | site-header |    708 | t
```

So the content exists, the reference exists, and the file is never written.

## Root cause

`collectJSAssets` (`platform/orchestration/actions/rerender_single_page_action.go:156-176`) is the
only thing that publishes `/tools/assets/{function}.js`, and it reads **page components only**:

```sql
SELECT DISTINCT cc.function, cc.js_content
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = $1
  AND cc.js_content IS NOT NULL
  AND cc.js_content != ''
```
→ `assetPath := fmt.Sprintf("tools/assets/%s.js", function)` (`:174`)

There is no `site_components` branch. A component reached only through `site_components` — i.e. all
chrome — can therefore declare JS that is never published, and no error is raised anywhere: the
render succeeds, the deploy succeeds, the page ships with a `<script>` tag pointing at nothing.

**The platform already knows how to do this correctly, three files away.** The *other* JS path,
which builds the `snippets.js` bundle, unions both tables
(`render_js_snippets_for_site_action.go:203-219`):

```sql
SELECT DISTINCT cc.function FROM page_components pc … WHERE p.site_id = $1
UNION
SELECT DISTINCT cc.function FROM site_components sc … WHERE sc.site_id = $1
```

**So the two JS paths disagree about whether chrome exists.** One models it, one does not. That
disagreement is the bug, and it is the same shape as the chrome/page split behind `bugs_open/018`
(the chrome renderer ignoring `input_schema` while the section renderer honours it) — chrome
repeatedly treated as a second-class citizen by machinery that handles page components properly.

## Blast radius (measured 2026-07-20)

Only **idea.uk** has chrome components carrying `js_content` today, so it is the only site with a
404ing chrome script right now. That is not reassurance:

- The defect is in **shared code**, and it fires the moment any site's header/footer needs
  behaviour. Every site's chrome is one interactive feature away from this.
- The `component-creator` agent is contractually told that template and schema are two views of one
  contract (`093_component_creator.sql:1185`) and that inline JS is separated into `js_content`
  (`store_generated_component_action.go`) — so a *generated* chrome component will naturally land in
  exactly this broken state.
- Nothing detects it. There is no check that a `<script src>` a component emits actually resolves.

## Fix candidates

1. **Add the `site_components` branch to `collectJSAssets`**, mirroring the UNION that
   `loadSiteComponentFunctionsForJS` already uses. Smallest change, removes the disagreement, and
   the correct query is already written three files away — copy it rather than invent one.
2. **Or fold chrome JS into the existing `snippets.js` bundle** (which already covers site
   components) and stop emitting per-component `<script src>` from chrome templates at all. Fewer
   requests, one path instead of two — but it changes load order, so check nothing depends on the
   per-component file.
3. **Detect it**: a post-deploy check that every `<script src="/…">` a rendered page emits resolves
   to a deployed file. Cheap, static, and would have caught this the day it shipped — and would also
   catch the inverse case seen in `bugs_open/024`'s vicinity, where an asset publishes but nothing
   references it.

⚠️ **Do not "fix" this by inlining the JS into the chrome template.** `store_generated_component_action.go`
deliberately separates inline JS into `js_content`, and `sql/p2_02` in the idea.uk workstream
documents the shape that is considered broken (`js_len=0, has_src_ref=f, has_raw_inline=t`).
Inlining trades a 404 for a convention violation and a component the tooling can no longer manage.

## How to verify a fix

```bash
# every chrome-referenced asset must resolve
curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk/tools/assets/site-header.js   # want 200
curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk/tools/assets/site-footer.js   # want 200
```
Then the behavioural test, which is the real one: **the hamburger opens the mobile menu** at ≤768px
on idea.uk. The markup is already correct (`.hamburger[aria-expanded]`, `#mobile-menu`,
`.mobile-menu.is-open` in the header template) — only the script is absent.

## Related

- `bugs_open/018` — chrome renderer ignores `input_schema` while the section path honours it. Same
  chrome-is-second-class shape; the template rewrite that fixed 018 is what surfaced this, because
  the `<script src>` reference it inherited had never resolved.
- `bugs_open/024` — the mirror image on the page side: a template fix that never reaches the page.
  Note the contrast — there the asset path worked and the markup was stale; here the markup is
  correct and the asset is missing.
- `docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §X.3–X.5 — discovery
  context and the live measurements above.

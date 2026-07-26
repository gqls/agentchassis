# 085 — the render data advertises `current_page` and the build path always supplies it empty

**Filed:** 2026-07-26, by the brochure_component_library workstream, which hit it building
the shared `evidence-chart` component. Found by measuring a rendered page, not by reading code.
**Severity:** Low-Medium — no data loss, no wrong figures. It silently removes a capability:
**no section component can know which page it is on**, so nothing can vary per page.
**Class:** structural (a key exists in the contract, is always empty on the main path, and
fails by doing nothing rather than by erroring).
**Status:** OPEN. Cause established and the fix is one line; not attempted here because Go is
inert until an image roll and this workstream's owner ruling was "config now, Go later".

---

## Symptom, measured

`evidence-chart` was placed on `index` and `capabilities`. Its chart definitions declare
which page they belong to (`"pages": ["index"]`), and the template filters on `current_page`,
degrading to "show everything" when that value is empty.

**Every chart rendered on `index`, including the two marked `capabilities`.** So the filter
never ran, which happens only when `current_page` is empty.

```sql
-- the rendered section carries all three charts, not the one assigned to this page
SELECT substring(pc.rendered_html from 'data-chart="[a-z-]*"')
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'fundamentallyai.com' AND p.name = 'index' AND cc.function = 'evidence-chart';
```

## Cause, read from the code

`current_page` **is** in the template data map — twice:
`platform/orchestration/actions/component_library.go:756` and `:873` both emit
`"current_page": ctx.CurrentPage`. So a component author reasonably concludes it is available.

`RenderContext.CurrentPage` (`component_library.go:79`) is set in the multipage path
(`multipage_actions.go:206`) but **never by `BuildRenderContextAction`**
(`v3_site_actions.go:866`), which is what the page-build pipeline uses.

That action merges its configured sources through `mergeIntoRenderContextEnhanced`
(`v3_site_actions.go:1022`). The page-content-writer's step config does pass the page:

```json
"sources": { "page": "input_data.current_page", "site": "input_data.site_record", ... }
```

but the merge extracts only a fixed allowlist — domain, company_name, logo_text, tagline,
email, phone, colours, a handful of image-URL fields — and drops everything else on the
floor. It never assigns `ctx.CurrentPage`, and the page record's `name` reaches no other
key either. **The page's identity is passed in and thrown away.**

## Why it is worth fixing rather than working around

- It fails **silently and plausibly**: a component that branches on `current_page` renders
  the "no information" branch, which looks like a design choice rather than a defect.
- The contract advertises the field, so the next author will make the same assumption.
- It is the difference between a component that can be placed twice with different content
  and one that must be placed once. That is a general capability, not one component's need.

## Fix candidate (one line, plus a test)

In `BuildRenderContextAction`, after the sources are merged, set `CurrentPage` from the
already-configured page source — the value is right there in `params.CollectedData`:

```go
if pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.current_page.name"); pageName != "" {
    renderCtx.CurrentPage = strings.TrimSuffix(pageName, ".html")
}
```

Alternatively assign it inside `mergeIntoRenderContextEnhanced` when `sourceName == "page"`,
which fixes every caller at once but touches a shared function — the more invasive of the two.

**Verification that would settle it** (this is the cheap part): re-render
fundamentallyai.com's `index` and `capabilities` after the roll and check that each page
carries only the charts assigned to it —

```sql
SELECT p.name, substring(pc.rendered_html from 'data-chart="[a-z-]*"')
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'fundamentallyai.com' AND cc.function = 'evidence-chart';
```

The data to prove it with is already seeded: `evidence_base.charts` holds one chart marked
`pages: ["index"]` and two marked `pages: ["capabilities"]`.

## Containment now

The `capabilities` placement was removed and the section left on `index` only, so the site
does not publish the same three charts twice. The `pages` key stays in the data, unused and
correct, so the fix turns it on rather than requiring new data.

## What this is NOT

- Not a claims or evidence defect: every figure rendered is a fact row, and the wrong-page
  charts were accurate, merely misplaced.
- Not `bugs_open/068` (rebuild writer/link-resolver contract) — different path, different
  field, no overlap beyond both touching the rebuild pipeline.

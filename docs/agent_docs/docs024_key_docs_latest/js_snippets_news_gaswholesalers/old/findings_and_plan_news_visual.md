# Findings from gaswholesalers news visual investigation, 2026-05-14

## Two debugging guide entries to add

Both go under `## 9. Specific Failure Patterns` in `016_debugging_guide_v2_4_.md`.
Place after the new "operator does not exist: jsonb && jsonb" entry from
the previous fix.

---

### css_snippets renders correctly but specific snippets missing from deployed styles.css

**Symptom:** Deployed `assets/css/styles.css` contains the
`/* === Component-specific styles === */` block (so loadComponentCSSSnippets
ran successfully — this is the latent bug fix proven working) but a
css_snippet you expected is absent. Verify by greping for one of the
snippet's distinctive selectors in the deployed file.

**Example:** After fixing the jsonb-operator bug, gaswholesalers's
deployed styles.css contains the `fade-in-up` snippet but not `Latest
News Grid` (`.latest-news-section …`) or `News Listing Page`
(`.news-listing-section …`). The news css_snippet rows exist with
`applies_to = ["latest-news"]` and `["news-listing"]`, but their function
names didn't appear in the components list the renderer received.

**Root cause:** `extractCSSComponents` in
`platform/orchestration/actions/render_css_from_spec_action.go` reads
`site_context.all_component_functions` from collected_data. If that
field is missing or empty, the function logs a warning and **falls back
to a hardcoded default list**:

```go
return []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"}
```

The hardcoded fallback does not include `latest-news`, `news-listing`,
`testimonial`, `pricing-table`, `tools`, or anything else that lives on
specific page types rather than across the whole site.

When webdesign-agent's `load_site_context` step fails to populate
`all_component_functions` correctly (which happens if the step uses
narrow criteria — e.g. only site_components, not page_components),
the renderer falls back, and page-level component snippets silently
don't ship.

**Diagnosis:**

```sql
-- 1. What's actually in css_snippets and what their applies_to says
SELECT name, applies_to::text AS applies_to, LENGTH(css_content) AS css_len
FROM css_snippets
ORDER BY name;

-- 2. What component functions does the site actually have across all pages?
SELECT DISTINCT cc.function, COUNT(*) AS uses
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '<site-id>'
  AND cc.function IS NOT NULL
GROUP BY cc.function
ORDER BY cc.function;

-- 3. Cross-reference: which applies_to values would match this site?
SELECT cs.name, cs.applies_to::text
FROM css_snippets cs
WHERE EXISTS (
  SELECT 1
  FROM jsonb_array_elements_text(cs.applies_to) AS a(elem)
  WHERE a.elem IN (
    SELECT DISTINCT cc.function
    FROM page_components pc
    JOIN content_components cc ON cc.id = pc.component_id
    JOIN pages p ON p.id = pc.page_id
    WHERE p.site_id = '<site-id>'
  )
);

-- 4. What was actually in site_context.all_component_functions during the
--    last webdesign-agent run? (collected_data is on the orchestration_states)
SELECT jsonb_pretty(collected_data -> 'site_context' -> 'all_component_functions')
FROM orchestration_states
WHERE owner_agent_type = 'webdesign-agent'
  AND collected_data -> 'site_context' -> 'all_component_functions' IS NOT NULL
ORDER BY created_at DESC
LIMIT 3;
```

Query 3 lists snippets that *should* be applied to the site based on
its components. Query 4 shows what the renderer actually saw. The gap
between them is the bug.

**Fix paths (cheapest first):**

- **A. Bypass via direct CSS append (one site, minutes):** generate the
  missing CSS server-side (concatenate the missing css_snippets rows'
  `css_content`) and git-commit it appended to the site's
  `assets/css/styles.css`. Survives until next webdesign-agent run, by
  which time hopefully (B) has landed.

- **B. Fix all_component_functions upstream (proper, hours):** the
  `load_site_context` action should populate all_component_functions
  from `page_components ∪ site_components` joined to
  `content_components.function`. If it only reads site_components, it
  misses page-level functions. Find and fix it. Webdesign-agent runs
  thereafter will include the full list.

- **C. Loosen the fallback (band-aid):** change the hardcoded default
  in `extractCSSComponents` to include the common page-level functions.
  Doesn't fix the underlying bug but makes the silent-degradation case
  less severe. Not recommended — masks the real problem.

---

### Migration updates content_components but deployed pages still show old content

**Symptom:** A surgical SQL migration to `content_components.html_template`
(e.g. extracting an inline `<script>` block out per contract 003) is
verified working in the DB, but the deployed pages on the affected sites
show the old content. The HTML still contains the inline `<script>`, the
contract-003 `<script src>` tag is absent, and the new `js_content`
asset is never requested by browsers.

**Example:** Migration B from May 2026 moved the news component IIFEs
from `html_template` to `js_content` for both `latest-news` and
`news-listing`. Verified post-migration:
- `content_components.html_template LIKE '%<script src=…>%'` → true
- `content_components.html_template LIKE '%<script>%(function%'` → false
- `content_components.js_content` populated with IIFE body

But gaswholesalers's deployed `news.html` (after page-rerender) still
shows the full inline IIFE in the news-listing section and no
`<script src="/tools/assets/news-listing.js">`.

**Root cause:** `page-rerender` uses `page_components.rendered_html`,
which is the **stored rendered output from when page-content-writer
last ran**. It does NOT re-pull from `content_components.html_template`.
Surgical updates to html_template only affect future content-writer
runs, not existing pages.

The `page_components.rendered_html` column is a snapshot from the
content-writer's LLM render at page-creation time, frozen until a
new content-writer run rebuilds that page_component.

**Diagnosis:**

```sql
-- For a given site + page + slot, compare the live template against
-- the stored rendered HTML:
SELECT
  pc.slot_name,
  pc.position,
  cc.function,
  LENGTH(cc.html_template) AS template_len,
  LENGTH(pc.rendered_html) AS rendered_len,
  cc.html_template LIKE '%<script src=%'      AS template_has_script_src,
  pc.rendered_html LIKE '%<script src=%'      AS rendered_has_script_src,
  cc.html_template LIKE '%<script>%(function%' AS template_has_inline_iife,
  pc.rendered_html LIKE '%<script>%(function%' AS rendered_has_inline_iife
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '<site-id>'
  AND p.name    = '<page-name>'
  AND cc.function IN ('latest-news', 'news-listing');
```

If `template_has_script_src=true` but `rendered_has_script_src=false`,
the migration didn't reach the stored snapshot.

**Fix paths:**

- **A. Surgical UPDATE on page_components.rendered_html (cheapest):**
  apply the same REPLACE / position+substring pattern from migration B
  to `page_components.rendered_html` directly, scoped to the affected
  site + pages. The page-rerender that follows will deploy the corrected
  HTML. Same caveat as migration B's regex section — use position +
  split_part, not regex.

- **B. Run page-content-writer for the affected pages (heavy):** the
  canonical "rebuild this page from scratch" path. Burns LLM tokens and
  may change unrelated content (the LLM regenerates anything ambiguous).
  Use only when (A) is too messy.

- **C. Make page-rerender pull fresh from templates (architectural):**
  changes the contract — page-rerender stops being a fixed-content
  redeploy and becomes a partial rebuild. Not a small change; deferred
  until there's a real need.

**General principle:** in this system, `page_components.rendered_html`
is a *snapshot* not a *view*. Migrations that change `content_components`
need to either also update the snapshots for affected pages, or trigger
a rebuild. The migration framework should ideally either do both or
fail loudly when only the template is touched.

---

## Plan for gaswholesalers news visual

Two independent fixes needed. Each has a quick (surgical) path and a
proper path. Doing the surgical paths today makes the visual correct;
the proper paths fix the underlying gaps for future sites.

### Fix 1 — News CSS not in deployed styles.css

**Surgical (today):** Append the news css_snippet content to
`gaswholesalers.com/assets/css/styles.css` directly via git commit.
The CSS is already in the css_snippets rows; we just bypass the
renderer and write it.

```sql
-- Pull the news CSS out for git append
SELECT css_content
FROM css_snippets
WHERE name IN ('Latest News Grid', 'News Listing Page')
ORDER BY name;
```

Copy each row's css_content into a file, git commit, push:

```bash
cd ~/projects/sites
{
  echo ""
  echo "/* === News component styles (manually appended pending all_component_functions fix) === */"
  # paste Latest News Grid css_content
  # paste News Listing Page css_content
} >> gaswholesalers.com/assets/css/styles.css
git add gaswholesalers.com/assets/css/styles.css
git commit -m "news: append news section CSS manually (pending all_component_functions fix)"
git push
```

Survives until next webdesign-agent run on gaswholesalers, when styles.css
gets regenerated. By then ideally fix 1B has landed and the regeneration
includes news CSS itself.

**Proper (separate session):**

1. Find where `load_site_context` populates `site_context.all_component_functions`.
2. Verify it joins through `page_components` (not just `site_components`).
3. Run diagnostic query 4 from the debugging-guide entry above on a few
   recent webdesign-agent runs to see what's missing.
4. Fix the load_site_context query.
5. Re-run webdesign-agent on affected sites.

### Fix 2 — News IIFE still inline on deployed pages

**Surgical (today):** UPDATE `page_components.rendered_html` for the two
gaswholesalers pages (index for latest-news, news for news-listing) to
swap the inline IIFE for a `<script src>` tag, then page-rerender.

```sql
BEGIN;

-- gaswholesalers index page — latest-news component
UPDATE page_components pc
SET rendered_html =
      substring(pc.rendered_html FROM 1 FOR position('<script>' IN pc.rendered_html) - 1)
   || '<script src="/tools/assets/latest-news.js"></script>'
   || substring(pc.rendered_html FROM position('</script>' IN pc.rendered_html) + length('</script>')),
    updated_at = NOW()
FROM pages p
JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'index'
  AND cc.function = 'latest-news'
  AND pc.rendered_html LIKE '%<script>%(function%'
  AND pc.rendered_html NOT LIKE '%<script src="/tools/assets/latest-news.js"></script>%';

-- gaswholesalers news page — news-listing component
UPDATE page_components pc
SET rendered_html =
      substring(pc.rendered_html FROM 1 FOR position('<script>' IN pc.rendered_html) - 1)
   || '<script src="/tools/assets/news-listing.js"></script>'
   || substring(pc.rendered_html FROM position('</script>' IN pc.rendered_html) + length('</script>')),
    updated_at = NOW()
FROM pages p
JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = p.id
  AND p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND p.name = 'news'
  AND cc.function = 'news-listing'
  AND pc.rendered_html LIKE '%<script>%(function%'
  AND pc.rendered_html NOT LIKE '%<script src="/tools/assets/news-listing.js"></script>%';

-- Verify both rows updated
SELECT p.name AS page, cc.function,
       pc.rendered_html LIKE '%<script src="/tools/assets/%' AS has_script_src,
       pc.rendered_html LIKE '%<script>%(function%'           AS still_has_inline_iife
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
  AND cc.function IN ('latest-news', 'news-listing')
ORDER BY p.name;

COMMIT;
```

Then trigger page-rerender for index and news (same kcat pattern as
before).

**Proper (separate session):**

Build a migration helper action that, when a `content_components.html_template`
column changes in a way that should propagate (script tag substitution,
class name changes, etc.), enqueues page-rerender items for every page
that uses that component. The migration framework should track which
columns are "deploy-affecting" and the system should self-heal.

### Order of operations for today

1. Run the diagnostic queries from both debugging-guide entries to
   confirm the diagnosis on gaswholesalers's actual data.
2. Apply Fix 1 surgical (CSS git append).
3. Apply Fix 2 surgical (page_components UPDATE).
4. Trigger page-rerender for `index` and `news` on gaswholesalers.
5. Open gaswholesalers.com — news cards should have the new visual
   design, dates expand to "2 days ago".

Total work: ~15 minutes. After that the news section is done for this
site. The two "proper" fixes (Fix 1B and Fix 2B) are work for separate
sessions — both useful, neither blocking other sites.

### What about other sites?

Other sites with `latest-news` or `news-listing` components have the same
two issues. The css_snippets fix means their next webdesign-agent run
WILL pick up the snippets — IF their `all_component_functions` includes
those function names. The contract-003 IIFE staleness applies to every
site whose news pages haven't been re-rendered by content-writer since
migration B (which is all of them).

If you want to roll the fix across sites, the same surgical approach
works site-by-site. A small loop over `sites WHERE id IN (...)` could
do the page_components UPDATE in bulk. The CSS append is per-site git
commit but scriptable.

Suggest letting the next natural webdesign-agent run cycle handle other
sites' CSS, and only batch-apply the page_components UPDATE if a specific
site is identified as needing it.

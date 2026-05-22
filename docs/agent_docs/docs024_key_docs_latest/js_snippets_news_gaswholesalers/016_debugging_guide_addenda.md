# 016 Debugging Guide — Addenda (collected)

Diagnostic patterns destined for `016_debugging_guide.md`, collected here
from several 2026-05 debug sessions for a single clean merge into the
canonical guide. Each entry is a Section 9 (Specific Failure Patterns)
addition unless marked Section 0 (Assumption Checklist).

Where an entry has a corresponding fix, the fix lives in a FOCUS doc and is
cross-linked — these entries are the *diagnosis*, not the fix.

Sources folded in: `016_debugging_guide_addendum_faq_diagnosis.md`,
`design_actions_status_filter_fix.md` (the trap lesson),
`css_snippets_matching_known_issue.md` (the diagnosis half),
`findings_and_plan_news_visual.md` (its two debugging entries).

---

## Section 9 — Specific Failure Patterns

### A component renders empty structural shells (FAQ accordion, news grid, any `{{range}}` template)

This is the headline entry: a general *diagnostic method* for "a component
renders empty shells", which generalises to any component-not-populating bug.
Fix and root-cause for the FAQ instance: `FOCUS_faq_empty_items_and_page_content.md`.

**Symptom:** A page renders the *structure* of a component but none of its
repeated content. For the FAQ case: four `<details class="faq-item">` blocks
with empty `<summary>` and `<p>` — the accordion shell is there, every
question/answer blank. The heading renders fine (static template text); only
the `{{range .questions}}` body is empty.

**Why this specific shape matters:** empty *shells* (not a missing section,
not a raw `{{range}}`) tells you the template engine ran and iterated — it
just iterated over items whose fields were empty, or over the right count of
empty objects. That points at the data binding, not the template or the
section list.

**Diagnostic method — narrow from the outside in.** Each step eliminates a
layer; stop when one returns the smoking gun.

```sql
-- STEP 1: List the page's components in order. Look for orphans
-- (component_id IS NULL) and check which slot the content should be in.
SELECT
  pc.id, pc.position,
  cc.function, cc.name AS component_name,
  pc.build_status,
  pc.component_id,                          -- NULL = orphaned, can't resolve template
  pc.content_item_id,                       -- where structured data may live
  jsonb_typeof(pc.content_data) AS cd_type, -- NULL = no inline data
  LEFT(pc.content_data::text, 120) AS cd_preview
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_components cc ON pc.component_id = cc.id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
ORDER BY pc.position;
```

Read off STEP 1: is the target component orphaned (`component_id IS NULL` →
renderer can't map it to a template, renders stale `rendered_html`)? And
where is the data (`content_data` empty AND `content_item_id` NULL → nothing
to bind)?

```sql
-- STEP 2: Confirm what the component template expects to bind.
SELECT id, name, function,
       LEFT(html_template, 1200) AS template_preview,
       input_schema
FROM content_components
WHERE function = '<function>'    -- e.g. 'faq'
ORDER BY updated_at DESC LIMIT 3;
```

The `input_schema` names the data keys; the template shows the `{{range}}`.
For `faq` the key is `questions` (array of `{question, answer}`, `min_items 3`).

```sql
-- STEP 3: Does the structured data exist anywhere — inline content_data OR a
-- linked content_item?
SELECT pc.id, pc.position, pc.content_item_id,
       pc.content_data,
       ci.content_data AS content_item_data
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_items ci ON pc.content_item_id = ci.id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
  AND pc.position = <target-position>;
```

Both empty/NULL → the data was never written to this component.

```sql
-- STEP 4: Count rendered shells — distinguishes "iterated over empty objects"
-- from "iterated over nothing".
SELECT pc.position,
       LENGTH(pc.rendered_html) AS html_len,
       (LENGTH(pc.rendered_html)
        - LENGTH(REPLACE(pc.rendered_html, 'faq-item__question', '')))
        / LENGTH('faq-item__question') AS shell_count
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
  AND pc.position = <target-position>;
```

Non-zero shell count with empty fields = a `questions` array of empty objects
was bound at some past render, frozen into `rendered_html`. Zero shells = the
range got an empty array.

```sql
-- STEP 5: Compare the page's ACTUAL sections against the site_plan.
-- Catches duplicate/competing sections and stale plans.
SELECT name, jsonb_pretty(sections) AS actual_sections
FROM pages
WHERE site_id = '<site-id>' AND name = '<page-name>';

SELECT jsonb_pretty(section) AS planned_sections
FROM site_specs ss,
     jsonb_array_elements(ss.data #> '{pages}') AS section
WHERE ss.site_id = '<site-id>' AND ss.aspect = 'site_plan'
  AND section->>'name' = '<page-name>';
```

For the FAQ page this exposed the root cause: `pages.sections` was
`["hero","generic-text-block","faq","call_to_action"]` (both a freeform block
AND a structured faq), and the `site_plan` query returned zero rows (the page
wasn't in the plan at all).

**General lesson:** for any "renders empty" bug, walk rendered HTML →
page_components (orphan? data?) → content_components (what key?) →
content_data/content_item (is the data there?) → sections vs plan (competing
slot?). The layer that first shows "nothing here" is your cause. Don't trigger
a rebuild before STEP 5 — a rebuild against an unchanged plan reproduces the
same split.

### Page exists with sections but is absent from `site_plan` (stale plan)

Structural prevention: `FOCUS_faq_empty_items_and_page_content.md` (the
gap-planner write-back) and `site_planner_depth` material folded there.

**Symptom:** A page renders and is in nav, but a `site_plan` lookup returns
zero rows for it. For gaswholesalers the `faq` page had `pages.sections`
populated and a live faq.html, but `site_plan` listed only 8 pages, faq not
among them.

**Root cause:** Pages added *after* initial planning (by content-gap-planner
or the improvement loop) get a `pages` row and nav entries but are never
appended to `site_specs.site_plan`. The plan drifts from reality with every
gap-added page.

**Why it matters:** anything reading `site_plan` as the authoritative page
list (audits, regeneration, sitemap) silently misses gap-planned pages;
`load_page_sections_from_spec` falls back to `pages.sections` for them (works,
but the plan can never enrich their briefs); debugging "what should this page
contain" via the plan gives a false negative.

**Diagnosis:**

```sql
-- Pages that exist but are not in the site_plan
SELECT p.name, p.page_type, p.build_status
FROM pages p
WHERE p.site_id = '<site-id>'
  AND p.status IN ('active','deployed')
  AND NOT EXISTS (
    SELECT 1
    FROM site_specs ss,
         jsonb_array_elements(ss.data #> '{pages}') AS pl
    WHERE ss.site_id = p.site_id
      AND ss.aspect = 'site_plan'
      AND pl->>'name' = p.name
  )
ORDER BY p.name;
```

**Fix (structural):** `apply_gap_plan` should append new pages to
`site_specs.site_plan` via deep-merge (mirroring how `enrich_news_feed`
updates the classification aspect); plus a periodic plan-reconciliation check
that diffs `pages` against `site_plan` and back-fills. Patching one site's
plan won't stop the next gap-planned page from drifting.

### Page plan sections are bare component names with no briefs (planner depth gap)

Prevention detail: `FOCUS_faq_empty_items_and_page_content.md` (per-section
briefs).

**Symptom:** Two content surfaces on one page compete for the same content
and the writer fills the wrong one (see the empty-shells entry). Underlying
cause: the plan gives the writer nothing to disambiguate them —
`site_plan.pages[].sections` is an array of bare strings with no per-section
intent. A brief on each section ("faq: 5-7 Q&A pairs addressing procurement
objections; generic-text-block: narrative intro, no Q&A") would route content
correctly, and would also enable plan-time duplicate detection and a
post-build validation surface.

### Assumed-status-values trap

Fix that surfaced it: `FOCUS_visual_pipeline_css_and_component_lists.md`
(the `loadPagesWithComponents` status filter).

When writing or modifying SQL that filters by a status column, ALWAYS query
`SELECT DISTINCT status FROM <table>` first to see what values actually exist.
The `pages.status` column uses `'active'` exclusively; the values `'deployed'`,
`'published'`, `'draft'`, `'planned'` that appear in older queries don't exist
in the data. Pattern-matching status names from other systems or from how
things "should" work leads to silent zero-result queries. The status filter on
`loadPagesWithComponents` excluded every page on every site for as long as it
was in the function; an upstream bug (reading from `pages.sections`) masked it
for months, and when that was fixed this filter became the next layer of the
onion.

### css_snippets exist but never reach the deployed styles.css

Two distinct causes were found; they stack. Fix and fuller analysis:
`FOCUS_visual_pipeline_css_and_component_lists.md`.

**Cause 1 — the component-list fallback fires.** `extractCSSComponents`
(`render_css_from_spec_action.go`) reads `site_context.all_component_functions`
from collected_data; if missing/empty it falls back to a hardcoded 5-item list
(`hero, services-grid, differentiators, social-proof, call-to-action`). That
list excludes `latest-news`, `news-listing`, `testimonial`, `pricing-table`,
etc., so page-level component snippets silently don't ship. The deeper bug:
`load_site_context` populated the field from the fallback (or from too-narrow
criteria), so every site got the same 5 names regardless of its real
components.

**Cause 2 — applies_to granularity mismatch (known issue, not yet fixed).**
Even with the right component list, the matching query does exact-text overlap
between `css_snippets.applies_to` and the site's component functions.
`applies_to` uses generic categorical terms (`card`, `feature`, `button`,
`cta`) while the system reports specific names (`features`, `testimonials`,
`differentiators`, `call-to-action`). No exact overlap → no match. So
`hover-lift` (`["card","feature","testimonial"]`) won't match a site with
`features`/`testimonials`. Singular/plural is the simplest case; lemma families
(`button` vs `cta-button`) are wider. Two fix paths: update `applies_to` to
real names (tight coupling, manual), or make the match lemma/slug-aware (loose
coupling, false-positive risk like `card` matching `scorecard`). Path 2 is the
long-term answer, in `loadComponentCSSSnippets`.

**Diagnosis (which snippets *should* match this site):**

```sql
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

-- And what the renderer actually saw last run:
SELECT jsonb_pretty(collected_data -> 'site_context' -> 'all_component_functions')
FROM orchestration_states
WHERE owner_agent_type = 'webdesign-agent'
  AND collected_data -> 'site_context' -> 'all_component_functions' IS NOT NULL
ORDER BY created_at DESC LIMIT 3;
```

The gap between the two is the bug. A loose-match sizing query (exact /
plural-of-singular / substring) is in
`FOCUS_visual_pipeline_css_and_component_lists.md`.

### Migration updates content_components but deployed pages still show old content

Principle and fix paths: `FOCUS_news_rendering_and_component_assets.md`.

**Symptom:** A surgical SQL migration to `content_components.html_template`
(e.g. extracting an inline `<script>` per contract 003) verifies in the DB,
but deployed pages still show the old content after page-rerender.

**Root cause:** `page-rerender` uses `page_components.rendered_html`, which is
the **stored rendered output** from when page-content-writer last ran — it
does NOT re-pull from `content_components.html_template`. Surgical template
updates only affect future content-writer runs, not existing pages.
`rendered_html` is a *snapshot*, not a *view*.

**Diagnosis:**

```sql
SELECT
  pc.slot_name, pc.position, cc.function,
  cc.html_template LIKE '%<script src=%'       AS template_has_script_src,
  pc.rendered_html LIKE '%<script src=%'       AS rendered_has_script_src,
  cc.html_template LIKE '%<script>%(function%' AS template_has_inline_iife,
  pc.rendered_html LIKE '%<script>%(function%' AS rendered_has_inline_iife
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '<site-id>' AND p.name = '<page-name>'
  AND cc.function IN ('latest-news', 'news-listing');
```

`template_has_script_src=true` but `rendered_has_script_src=false` → the
migration didn't reach the stored snapshot. **General principle:** migrations
that change `content_components` must also update the snapshots for affected
pages, or trigger a rebuild — ideally the migration framework does both or
fails loudly when only the template is touched.

---

## Section 0 — Assumption Checklist additions

**16. "Renders empty" is a data-binding diagnosis, not a template one.** When
a component shows structural shells with no content, resist editing the
template or re-triggering a render first. The template ran — that's why the
shells exist. Walk the data path: is the page_component orphaned
(`component_id IS NULL`)? Is `content_data`/`content_item_id` empty? Is there
a competing section that got the content instead? Only after the data path is
cleared should you suspect the template. A rerender before this diagnosis just
re-emits the same empty shells and costs a cycle.

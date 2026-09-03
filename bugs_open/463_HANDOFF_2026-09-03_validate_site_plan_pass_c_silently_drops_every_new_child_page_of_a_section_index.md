# 463 — `validate_site_plan` Pass C silently drops every NEW child page of a section index, so an empty section index can never be filled

**Filed 2026-09-03 ~15:40Z by the `gamedesign.uk` lane. UNOWNED.**
**Severity: this is why a site with an articles/guides/news hub and no children stays empty for ever, through any number of re-plans.**

## 1. Symptom, at the artefact

gamedesign.uk has served an articles hub with **zero articles since it was built**. Three
rebuilds failed to fill it. The third one (2026-09-03) is the clean case, because by then the
planner was doing everything right:

- `plan_site` emitted **9 pages**, five of them `blog-post` on real subjects.
- `validate_plan` returned **4**. The five articles were gone.
- **Silently**: `capability_gaps_emitted: 0`, no `agent_error_log` row, orchestration `COMPLETED`.

Evidence (orchestration corr `9fe9660e-7272-4f51-b968-2ff769738086`, plan `005fb393`):

```sql
SELECT jsonb_array_length(collected_data->'plan_site'->'result'->'pages'),   -- 9
       jsonb_array_length(collected_data->'validate_plan'->'pages')          -- 4
  FROM orchestration_states WHERE correlation_id='9fe9660e-7272-4f51-b968-2ff769738086';
```

Dropped: `article-sign-off-problem`, `article-gdd-abandonment`,
`article-design-engineering-handoff`, `article-narrative-design-pipeline`,
`article-principal-transition` — all `page_type=blog-post`, urls `/articles/<slug>.html`.
9 − 5 = 4: the arithmetic matches exactly, so nothing else removed pages in this run.

## 2. Root cause — read in the code, not inferred

**Pass C** (`platform/orchestration/actions/v3_site_actions.go:7599`), intended to drop a FLAT
page that collides with a realised section index:

```go
if idxName, isStem := sectionStems[lslug]; isStem &&
    !isSectionIndexType(ltype) && lname != idxName {   // -> dropped
```

Both sides of that comparison are computed as **the first path segment**:

- `slugOf(name, url)` (`:6467`) trims `/`, strips `.html`, strips a trailing `/index`, then
  returns everything before the first `/`. So
  `slugOf("article-sign-off-problem", "/articles/the-sign-off-problem.html")` = **`"articles"`**.
- `sectionStemOf(name, url, pageType)` (`:6447`) does the same for the realised hub:
  `sectionStemOf("articles-index", "/articles/index.html", "section-index")` = **`"articles"`**.

**So a legitimate CHILD of a section index is indistinguishable from a flat page colliding with
it.** Every child the planner proposes under `/<section>/…` is dropped. The URL form does not
save you: `/guides/buy-to-let/index.html` also reduces to `guides`, because `slugOf` strips the
trailing `/index` before taking the first segment.

## 3. Why this has been live since 2026-05-21 and nobody noticed

`Pass A: union — add preserved realised pages not present by name` runs **after** the drop loop
and restores **realised** pages. So:

- an **existing** child is dropped by Pass C and immediately restored by Pass A → invisible;
- a **NEW** child has no realised counterpart → it stays dropped, permanently.

**Net effect: you cannot add a new child page to an existing section index.** An established site
looks perfectly healthy; a hub that is empty *today* can never be filled.

Control, and it is the one that makes the diagnosis falsifiable rather than a story:
mortgagecalculator.co.uk's **9** guide children ARE in its current plan (2026-08-02, i.e. written
after Pass C existed), at `/guides/<slug>/index.html` — restored by Pass A. Introduced in
`f026ad143` (2026-05-21, "site adoption locks").

> ⚠ **A measurement trap I fell into and corrected, recorded because it inverts the finding.**
> My first census counted "children in plan" with `url NOT LIKE '%/index.html'` and returned **0**
> for idea.uk, mortgagecalculator.co.uk and vetcomparison.uk — which reads as fleet-wide plan
> damage. **That zero was my filter**, not the estate: those children use the DIRECTORY form
> `/guides/<slug>/index.html` and my predicate excluded every one. The real discriminator is
> **realised-vs-new**, not URL form. If you re-measure this bug, list the plan rows and LOOK at
> them before filtering.

## 4. The interlock — two guards in series, and this is the part that makes it a trap

`bugs_open/444`'s gate (migration 720, live) holds a listing page whose item source resolves to
zero, filing `capability_gap` `builder_needed=section_children:<page>`. Pass C is what makes that
source resolve to zero. **Pass C drops the children; the gate then holds the childless hub.** Each
guard is defensible alone; together they make an empty section index permanently unfillable, and
each one's evidence looks like a reason for the other.

gamedesign.uk got exactly that pair on 2026-09-03: `capability_gap` `builder_needed=
section_children:articles-index` at 10:40:18Z, and five children dropped at 14:15Z.

## 5. What this is NOT

- **NOT a planner failure, and NOT a rule-20 failure.** `build-site-planner` rule 20 (migrations
  730/731) worked: the planner planned five launch posts on real subjects from the briefing, with
  no deferral language. `bugs_open/428`'s omission-reasoning defect is a DIFFERENT, earlier
  failure on the same site — that one is now fixed and this one was hiding behind it.
- **NOT `bugs_closed/141`**, which touched `sectionStemOf`/Pass C for a different symptom (a
  news index excluded from nav). Same functions, different consequence; read it before editing.
- **NOT about `parent_section`.** The dropped pages carried `parent_section: null`, but Pass C
  never reads that field — setting it correctly would not have saved them.

## 6. Fix candidates, ordered by what closes the door

1. **Make Pass C distinguish a child from a collider — the only fix that makes the bad state
   unrepresentable.** The collider case Pass C exists for is a page whose URL *is* the section
   path (`/articles.html`, or a page named `articles`); a child has a further path segment
   (`/articles/x.html`, `/articles/x/index.html`). Both currently reduce to `articles` because
   `slugOf` returns the first segment. Compare the FULL path against the hub's path instead, and
   Pass C keeps its purpose while children survive. Needs a test per URL form: `/articles.html`
   (drop), `/articles/x.html` (keep), `/articles/x/index.html` (keep), `/articles/index.html`
   (the hub itself, already excluded by `isSectionIndexType`).
2. **Fail loudly instead of silently.** Whatever the drop rule ends up being, a page removed
   between `plan_site` and `validate_plan` should file a durable finding naming the page and the
   pass. Today the only observable is a page count that nothing compares. `bugs_open/428`'s
   in-flight work files record-mode `capability_gap` rows for planner omissions — that machinery
   is the natural home, but it must read the **pre-Pass-C** page list or it will see nothing
   (the type WAS planned, then deleted).
3. Do NOT "fix" this by having the planner emit flat article URLs outside the section prefix.
   That trades a dropped page for an orphaned one and breaks the `section_children` resolver
   that 444's gate uses.

## 7. How to verify a fix

Re-plan a site with a realised section index and no children (gamedesign.uk is the live case) and
assert on the STEP BOUNDARY, not the served page:

```sql
SELECT jsonb_array_length(collected_data->'plan_site'->'result'->'pages') AS proposed,
       jsonb_array_length(collected_data->'validate_plan'->'pages')       AS survived
  FROM orchestration_states WHERE correlation_id='<corr>';
```

`proposed = survived` for a plan with new children is the pass condition. Then confirm the
children reach `site_plan_pages`, and only then look at the served hub. A served-page check alone
cannot distinguish "Pass C dropped them" from "the planner never proposed them" — which is the
whole reason this bug survived two rebuilds and a fleet-wide prompt fix.

## 8. Ownership / routing

**Unowned.** Found by the `gamedesign.uk` lane, which owns the site and is not taking the fix.
The `428` session is mid-build **inside `validate_site_plan`** on the adjacent omission-reasoning
defect and has been told directly — if a fix lands, it should probably land there rather than
have two lanes editing one action. `designblog.co.uk` (owns 730/731) and `bugs_open/427` have been
told, because a zero-article outcome would otherwise read as evidence against rule 20 and it is
not. `scripts/who-owns.py 463` before routing work at it.

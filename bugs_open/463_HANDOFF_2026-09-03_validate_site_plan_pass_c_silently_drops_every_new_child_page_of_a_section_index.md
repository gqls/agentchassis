# 463 — `validate_site_plan` Pass C silently drops every NEW child page of a section index, so an empty section index can never be filled

**Filed 2026-09-03 ~15:40Z by the `gamedesign.uk` lane.**

> ## ✅ FIXED IN CODE 2026-09-03 by session `463` — commit `9b540c2e6`. STILL OPEN: inert until the chassis image rebuilds and rolls, and unverified at the artefact until then.
>
> Both halves are in that commit: Pass C narrowed to a true collider, and the
> `parent_section` derivation §5 below said was not needed (see the correction there —
> **without it the Pass C fix changes nothing at the artefact**). Council submitted,
> corr `9f6c6374-1b76-4094-9b4c-e04808d8428c`, verdict pending.
> Tests: `platform/orchestration/actions/v3_site_reconcile_section_children_test.go`,
> `platform/orchestration/datahelpers/page_parent_section_test.go`.
> **Do not re-plan gamedesign.uk until the roll** — a fourth plan before then is deleted
> by the same pass.

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

> **CORRECTED 2026-09-03 by session `463`, while fixing this bug.** The bullet above is
> **true of Pass C and false of the write path**, and taken as written it would have
> produced a fix that changed nothing on the served page. Pass C indeed never reads
> `parent_section`. But `WriteSitePlanAction` and `SyncPagesToDBAction` both **discard the
> planner's url** and re-derive it from `datahelpers.CanonicalisePage`, whose `blog-post`
> arm is `dir := parent; if dir == "" { dir = "blog" }`. `ValidateRoles` copied
> `ParentSection` verbatim and never derived it, and its rule 5 (`nestedRoleFromURL`)
> rescues only `tools`/`guides`/`games` — which is exactly why §3's control
> (mortgagecalculator's `/guides/` children) survives and `/articles/` cannot.
>
> So a child kept by a fixed Pass C would have been written to `/blog/<slug>.html`,
> `countSectionChildren("/articles/")` would still have counted zero, and 444's gate would
> still have held the hub. Same empty page, different cause.
>
> `[MEASURED 2026-09-03]` at the live `agent_definitions` row: the `plan_site`
> prompt_template is **32,191 chars** and does not contain the string `parent_section`;
> `site_plan_pages` holds **109** `blog-post` rows, **109** with `parent_section` absent and
> **0** set. Every leaf-role plan row fleet-wide sits in its role's DEFAULT directory except
> where a realised identity or an explicit Go producer supplied the value. The planner has
> never once placed a leaf page in a custom section, and cannot: nothing asks it for the
> field and, until this fix, nothing derived it.
>
> What caught it: tracing what happens to a surviving page rather than stopping at the drop.
> The cheap check that would have caught it at filing time is one query —
> `SELECT role, parent_section, url FROM site_plan_pages WHERE role='blog-post' LIMIT 20`
> — which shows every row at `/blog/` with a NULL parent.

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

---

## 9. The fix, and what it deliberately leaves (added 2026-09-03 by session `463`)

### Fleet blast radius, measured before fixing

`[MEASURED 2026-09-03]` **53 of 78** section-index-family hubs at `/<sec>/index.html`, across
**21 sites**, have zero child pages under their prefix. gamedesign.uk is not a special case; it
is the case that got diagnosed. (This is the population that *cannot* be filled while the bug
stands — not a claim that Pass C emptied all 53.)

### What changed (commit `9b540c2e6`)

1. **Pass C compares the full path each side CLAIMS**, through `datahelpers.PagePathKey`, the
   estate's existing collision key. `/articles.html` and `/articles/index.html` both claim
   `/articles` → still dropped. `/articles/x.html` claims `/articles/x` → kept. The name-stem
   map survives only as the fallback for a plan page carrying no url, where behaviour is
   unchanged.
   This makes Pass C and §4's gate complementary **by construction**: "claims the hub's path"
   and "lives under the hub" (`countSectionChildren`'s prefix test) partition the old
   first-segment test, so the two guards in series can no longer disagree about one page.
   `[MEASURED]` **83 of 83** live hubs have a name-derived stem equal to the one their url
   yields, so on today's estate the new rule drops a strict **subset** of what the old one did.
2. **`ValidateRoles` derives `parent_section` from a leaf page's own url** — see the §5
   correction for why this is not optional. Gated to leaf roles, to an absent value, and to an
   entry the reconciler has NOT paired with a realised page.
3. **A drop now names the page, its url and the pass** on `reconcileCounts`.

### Deliberately not done, and by whom

- **The durable finding of candidate 2 is the `428` lane's**, shipped the same day
  (`recommended_type_reconciliation.go`, commits `eee40b554`/`91173c6d7`). It classifies by
  STAGE rather than by pass, which is the vocabulary that survives these passes being
  renumbered. Two lanes filing overlapping findings for one event is the drift this estate
  keeps paying for, so this bug's producer-side record stays a `reconcileCounts` field.
  **Residual that lane named itself:** its check is type-level, so a re-derived url is invisible
  to it — the `blog-post` type IS present in the final set, so it reports no omission while the
  hub is still held. That is exactly the half §5's correction covers.
- **Candidate 3 (flat article URLs outside the section prefix) was NOT taken.** It trades a
  dropped page for an orphaned one and breaks the `section_children` resolver 444's gate
  depends on.

### Residuals, filed or named rather than fixed here

- **`bugs_open/465`** — `truncatePreservingRealised` drops EVERY net-new page once the preserved
  set reaches `max_pages` (20), with one `logger.Warn` and no durable record. Same silent-shrink
  signature, different pass; it means this fix alone will not fill a hub on a site of 20+ pages.
- **`create_blog_posts_action.go`, `apply_gap_plan_action.go`, `apply_adoption_plan_action.go`**
  call `CanonicalisePage` without threading `ParentSection` at all, so they still write
  blog-posts to `/blog/` whatever the plan says. Found by the `feed lane` and verified here.
  Unaffected by this fix because there is no url at those call sites to derive from — a
  different fix, in a producer dormant since 2026-04-24 (`bugs_open/460`).
- **`bugs_open/457`** (`rebuild_blog_listing` appending orphan `page_components` rows) is the
  hub-RENDER path and is owned and in flight. It decides whether a filled hub actually *lists*
  its children, so it gates the end-to-end verification of this fix, not the fix itself.
- **`sectionStems` is last-write-wins with no ambiguity refusal**, unlike every other map in
  `reconcilePlanWithRealised`. Left as-is: the refusal direction is already the safe one and the
  case is unmeasured.

### How to close this bug

Not before the chassis image rebuilds and rolls — Go changes are inert until then. Then §7's
step-boundary check, and one addition to it: confirm the children reach `site_plan_pages` at
**`/articles/<slug>.html`, not `/blog/<slug>.html`**. That second assertion is the one a
Pass-C-only fix would fail, and a served-page check cannot distinguish the two.

# 015 — A mistyped `page_type` orphans a page from every gate that keys on it (OPEN, fleet-relevant)

**Found:** 2026-07-18 on relojistas.com (Spanish news portal). **Status:** worked around on
this site; the underlying planner behaviour is **unfixed and likely affects other sites**,
especially non-English ones. Related family: 016b §9 "Page build 'completes' having built
nothing — zero planned sections treated as success".

## Symptom

The site's news-listing page (`/noticias/index.html`) sat at `build_status='planned'` with
`sections=[]` forever. Its build work item went to `needs_human_review` with
`page-build-handler no-op: no sections ready to build (empty spec sections…)`. Because the
page never deployed, the header nav link to it was a **live 404** — a visible dead end on a
production site. Meanwhile the news pipeline was healthy and the *homepage* news card worked
perfectly, which made the empty page look like an isolated content gap rather than a
routing problem.

## Root cause

The planner created the news listing page as **`page_type='section-index'`** — a generic
"index of a section" — instead of **`page_type='news-index'`**, the type the news machinery
keys on. Nothing validates that a site whose classification says
`content_features.news_feed.separate_page=true` actually *has* a `news-index` page; the
planner is free to satisfy the intent with a differently-typed page, and did.

`page_type` is not a label here, it is a routing key. Being mistyped orphaned the page from
every gate at once:

- **`render_news_section`** only emits `data/news-archive.json` and only looks for a
  listing page `WHERE page_type='news-index'` → no archive data was ever produced for it.
- **`MissingNewsPageCheck`** (`discovery_checks/check_news_feed.go`) fires only when
  `separate_page=true` **and no `news-index` page exists** → it *would* have fired and
  routed to `content-gap-planner` (the proven chain that built gaswholesalers' `/news.html`)
  — but the presence of a page that *looks* like the news page did not suppress it; rather,
  the site never ran the discovery sweep, and had it run it would have created a **second,
  duplicate** page rather than fixing the mistyped one.
- **`page-build-handler`** had nothing to build: a `section-index` gets no news component
  assigned, so `sections` stayed `[]` and the build no-op'd into `needs_human_review`.

So three separate mechanisms each did the right thing for their own key and collectively
left an empty, unreachable page with no error that names the cause.

## Workaround applied (relojistas)

Targeted, reversible, no re-plan (a re-plan is the 001 clobber landmine):
```sql
UPDATE pages SET page_type='news-index',
       sections='["hero","news-listing","call-to-action"]'::jsonb
WHERE site_id=$1 AND name='noticias-index';
UPDATE site_work_items SET status='triaged', error=NULL
WHERE site_id=$1 AND item_type='needs_page' AND summary ILIKE '%noticias%';
```
Section set copied from the working reference (gaswholesalers/robot-hands `/news.html` =
`[hero, news-listing, call-to-action]`). The `news-listing` component already exists.

## Fix candidates (none applied — needs a decision)

1. **Validate intent against realisation.** If `news_feed.separate_page=true`, assert a
   `news-index` page exists; if a page occupies the role under another type, *re-type it*
   rather than mint a duplicate. Cheapest correct fix; belongs next to `MissingNewsPageCheck`.
2. **Teach the planner the typed vocabulary** so a "news"/"noticias" listing page is
   emitted as `news-index` (mirrors the known `page_type='game'` vocabulary gap that caused
   re-typing and duplication — same class of bug).
3. **Make `MissingNewsPageCheck` adopt rather than create** — prefer re-typing an existing
   sectionless listing page over `approach=new_page`, which would otherwise leave a
   duplicate English `/news.html` alongside a Spanish `/noticias`.

## How to find others

```sql
-- sites that want a separate news page but have no news-index page
SELECT s.domain
FROM sites s JOIN site_specs ss ON ss.site_id=s.id
 AND ss.aspect='classification' AND ss.is_current
WHERE (ss.data->'content_features'->'news_feed'->>'separate_page')::bool
  AND NOT EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.page_type='news-index');
-- and the general shape: nav-visible pages that can never build
SELECT s.domain, p.name, p.page_type FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.build_status='planned' AND (p.in_header OR p.in_footer)
  AND jsonb_array_length(COALESCE(p.sections,'[]'::jsonb))=0;
```

## CORRECTION (2026-07-18, after council review) — the link-audit machinery already exists

I drafted a new `unresolvable_internal_links` discovery check for the dead-nav-link half of
this and submitted it to the council gate (`SUBMISSION_CORR=8a1f7b4f`). **It was a duplicate.**
The reuse seat's own suggested check ("search for an existing link-resolution/validation
helper … rather than reimplementing href extraction") surfaced
**`discovery_checks/check_phantom_internal_links.go`**, which already:

- detects internal links in **deployed** rendered HTML whose target is not a real page,
  over both `page_components.rendered_html` and `site_components.rendered_html`;
- reuses the **shared** `ExtractHrefs` / `ClassifyLinkScope` / `PageURLSet` datahelpers —
  the same ones `validate_page_content` (the deploy gate) uses, so gate and audit agree by
  one implementation (my draft would have reimplemented extraction/normalisation);
- **routes by surface**: `site_component` nav literals → `nav-link-fixer`; `page_component`
  hero/CTA/body links → `page-build-handler`, on the stated reasoning that the link resolver
  "is a build-time augmenter, not a rendered-HTML patcher". My draft routed both to
  `page-rerender`, which a reviewer correctly showed would **re-emit an LLM-invented href
  unchanged** and then mark the item resolved — silent-success.

**Decisive for this bug:** that check treats "real page" as *a `pages` row exists* — so
"a planned-but-unbuilt page has a row and is **not** flagged". That is deliberate. Therefore:

| Dead link class | Covered by phantom check? | Correct fix |
|---|---|---|
| `/ferias`, `/archivo`, `/guias/mantenimiento` (LLM-invented; **no** page row) | **Yes** — would flag | run `completeness-discovery-agent` on the site |
| `/noticias`, `/guias`, `/glosario` (page rows exist, `planned`) | **No** — excluded by design | build/deploy the pages (this bug) |

State observed 2026-07-18: `phantom_internal_links` **is** enabled in
`completeness-discovery-agent` (not in design/quality discovery agents) and has produced
**0 work items fleet-wide** — for relojistas simply because no completeness sweep has run
on it yet. No new detector is needed; the gap for *this* bug remains the mistyped
`page_type`, unchanged above.

## Transferable pattern (also filed to 016b §9)

When a value is used as a **routing key by several independent gates**, a wrong value does
not produce one loud failure — it produces silence in every gate at once, and the symptom
surfaces somewhere unrelated (here: a 404 in the nav). Diagnose by asking *which key does
each mechanism select on*, not by following the visible symptom.

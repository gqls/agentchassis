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

---

## PARTIAL FIX APPLIED 2026-07-20 — candidate 3 (adopt, don't duplicate). Candidates 1 and 2 still open.

`MissingNewsPageCheck` no longer assumes "no `news-index` page" means "no news
page". Before emitting the gap item it looks for pages already occupying the
role, and when it finds any it tells the planner to **re-type one of them**
rather than create a second page — the outcome §"Fix candidates" (3) asked for,
and the one that would have prevented an English `/news.html` shipping beside
the Spanish `/noticias`.

**The intervention is the natural-language `description`/`suggestion` fields,
deliberately.** The handler is `content-gap-planner`, an LLM that reads them.
Adding a new structured key like `approach` *alone* would have been the
dead-config shape of `bugs_open/025` and `/042` — a field written by one side
and read by nobody. `approach` is set too, but the words are what carry it.

**Detection is structural, not name-based:** nav-visible, `sections` empty, and
`build_status <> 'deployed'`. A vocabulary match on "news"/"noticias"/
"nachrichten" is the name-heuristic shape `bugs_open/044` was filed against and
would fail worst on exactly the non-English sites this bug hurts most.

> **CORRECTION — my first predicate was wrong, and live data caught it.**
> I first wrote the predicate as nav-visible + sectionless, copying this file's
> own "how to find others" query. Run against production it returned **six pages
> on ai-agent-orchestration.com — all `deployed`, all working**: tool pages and a
> blog index whose `sections` are legitimately empty because their content comes
> from elsewhere. The check would have told the planner those six "can never
> build and are dead links today", which is false, and invited it to re-type a
> live tool page into a news index.
> **An empty `sections` array only means "stranded" when the page never built.**
> Adding `build_status <> 'deployed'` takes the fleet result to zero rows — which
> is correct, because relojistas was already hand-fixed. The clause is pinned by
> a test regex so it cannot be quietly dropped.
> This file's own diagnostic query has the same false-positive and should not be
> pasted into anything that acts on the result.

**Scope, measured not assumed.** Exactly **one** site currently has
`separate_page=true` with no `news-index` page (ai-agent-orchestration.com), and
after the correction it yields zero candidates. So this is a **guard against
recurrence**, not a repair of live damage — the triage note calling it
"fleet-wide, non-English sites worst" was describing the mechanism's reach, not
its current incidence.

**Still open, and still the real fix:** candidate 2 — the planner emits
`section-index` for a news listing in the first place. Everything above is
downstream mitigation. Candidate 1 (assert intent against realisation at
classification time) is also untouched.

**LIVE 2026-07-21 in v1.0.1144** — pod-verified (`stranded nav page list capped`
and `RE-TYPE that page to page_type` both present in the running binary). The
case **stays OPEN** regardless: the partial (candidate 3) is only the
adopt-don't-duplicate mitigation; the defect that makes the planner emit
`section-index` for a news listing (candidate 2) is untouched, so the class is
not closed. Do NOT move this to `/bugs_closed/` on the strength of the partial
being live.

---

## CANDIDATES 1 + 2 BUILT 2026-07-24 — and the root cause was deeper than this file said

> **CORRECTED 2026-07-24 — "the planner created the page as section-index" was
> only half true, and the missing half changes the fix.** The mistype is
> manufactured **deterministically in Go**, not (only) chosen by the LLM.
> `ValidateRoles` (platform/orchestration/datahelpers/page_role_validator.go)
> rewrites ANY non-leaf role to `section-index` when the page name ends in
> `-index` (rule 2), when other pages declare it as parent (rule 3), or when
> the URL is `/<slug>/index.html` (rule 4) — and `news-index` was not a leaf
> role, so even a correct `news-index` emission from the LLM would have been
> flattened. Both persist paths feed the corrected role onward, and
> `CanonicalisePage` had no `news-index` case either. Teaching the prompt the
> vocabulary (candidate 2 as originally written) would have changed nothing on
> its own. Caught by reading the persist path end-to-end instead of stopping
> at the prompt.

**Second finding: candidate 3's advice was unexecutable.** The re-type
suggestion shipped in v1.0.1144 routes to `content-gap-planner`, whose entire
action surface (`apply_gap_plan_action.go`) was `add_to_page | new_page |
update_spec | not_actionable` — `retype_existing` hit the unknown-approach
branch (`applied=false`), or the LLM mapped it onto `new_page` and minted the
exact duplicate the advice warns against. No code path anywhere let the
gap-planner change `pages.page_type`. Third finding, same class:
`defaultSectionsForPage` ignored its `pageType` parameter, so a news page
whose plan omitted sections got `generic-text-block` instead of
`news-listing` — orphaned again, one layer down.

**What shipped (commit carries all of it; council corr `45664479`):**

- `page_role_validator.go` — rule 1b: an explicit flavoured index role
  (`isTypedIndexRole`, currently just `news-index`) is trusted as-is; rules
  2–4 can no longer flatten it. Sloppy/generic roles are corrected exactly as
  before (regression-pinned). `normaliseRole` keeps `news-index` distinct.
- `page_canonical.go` — `news-index` joins the section-index family: planner
  shape `{news-index, noticias}` → `(noticias-index, /noticias/index.html,
  news-index)`, flavour preserved as page_type (the family's stated design).
- `apply_gap_plan_action.go` — new `retype_existing` branch,
  **fail-closed**: the LLM plan only NAMES the page; the authorising facts
  (candidate set, target page_type) come from the original work item's spec
  as written by `findStrandedNavPages` — a page outside the stranded set is
  refused, the target type is never LLM-chosen, `RowsAffected==0` refuses a
  stale candidate. On success: page re-typed + sections set by candidate id,
  `needs_content_page` item filed for page-build-handler, original completed.
  Deliberately no growth-budget check (nothing is created).
  `defaultSectionsForPage` now keys `news-index` → `[hero, news-listing,
  call-to-action]` before the name heuristics (names are localised, types are
  not).
- Tests: `TestValidateRoles_NewsIndexFlavourPreserved` (one subtest per
  flattening rule), `TestCanonicalisePage_NewsIndex`,
  `apply_gap_plan_retype_test.go` (happy path pins UPDATE-by-candidate-id +
  spec-sourced type; three refusal tests pin that nothing beyond the spec
  read touches the DB). All pass on a clean `git archive HEAD` overlay.
  (`discovery_checks` has a PRE-EXISTING verifier-coverage failure on pure
  HEAD — `backend_entry_orphaned`/`contact_form_undeliverable` lack
  verifiers — another thread's, not this change's.)

**ACTIVATION — two steps remain, in this order (Go is inert until rolled):**

1. Roll a chassis image containing this commit; verify against the pod with a
   string this change CREATES plus a positive control:
   `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c isTypedIndexRole'` (>0)
2. Apply BOTH migrations (they are fail-closed — anchor/md5-gated WHERE, so
   drift = 0-row no-op, and each header carries the pod-grep gate):
   - `docs/agent_docs/sql_for_agents/206_planner_news_index_page_type.sql`
   - `docs/agent_docs/sql_for_agents/206_content_gap_planner_retype_approach.sql`
   Applied BEFORE the roll they recreate the bug (planner emission flattened
   by the old binary) or dead-end gap items (approach E with no executor).

Then the branch check (verify-the-failing-branch): confirm a
`retype_existing` plan actually flips `pages.page_type` and files the build
item — ai-agent-orchestration.com is the one `separate_page=true` site with
no `news-index` page (its stranded-candidate set is empty, so it exercises
the `new_page` arm; a synthetic stranded row exercises the retype arm).

**Adjacent gaps observed, deliberately NOT fixed here (evidence first):**
- Rules 2–4 flatten an explicit `entity-directory` the same way (`{name:
  clinics-index, role: entity-directory}` → `section-index`); `blog-index` is
  collapsed even earlier, by `normaliseRole`, deliberately. No observed
  incident for either; `rebuild_blog_listing_action.go:326` even auto-repairs
  the blog case downstream. Widen `isTypedIndexRole` only when a real page
  breaks.
- `check_model_directory.go`'s missing-page check (same shape as news)
  deliberately omits stranded/retype logic; if model-directory ever grows
  legacy mistyped pages, the executor branch is already generic — the check
  need only write `retype_candidates` + `page_type` into its spec.

Case stays OPEN until the image rolls, both migrations apply, and the branch
check above passes.

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

---

## CLOSED 2026-07-26 — all three activation steps done, both branches induced live

The three conditions the section above set are met. Evidence for each, inline.

### 1. Image rolled and pod-verified

Pod `agent-chassis-774877f4c6-zjh4t`, image
`docker.io/aqls/agent-chassis:v1.0.1159`:

| grep | count |
|---|---|
| `isTypedIndexRole` (created by this change) | 1 |
| `applyRetypeExisting` (created by this change) | 4 |
| `stranded nav page list capped` (positive control, v1.0.1144) | 1 |
| `zzz_never_exists` (negative control) | 0 |

### 2. Both migrations applied and recorded

Applied by hand (the runner's `--apply` would have swept four other threads'
pending files), then recorded with `--record-only`:

- `206_planner_news_index_page_type.sql` — anchor pre-check returned `1|1`,
  `UPDATE 1`, post-conditions `has_news_index_row=t, has_news_rule=t,
  old_row_gone=t`.
- `206_content_gap_planner_retype_approach.sql` — the md5 gate
  `c86623bec9455d745e9e6e03119d6ba5` still matched the live template exactly, so
  no thread had touched the prompt since capture; `UPDATE 1`, `has_approach_e=t,
  has_schema_entry=t`.

**Escape-integrity check on the result** (the `\n` in migration 1 are JSON
escapes inside a `::text` splice, a known trap): the live `plan_site`
prompt_template holds **259 real newlines and zero literal backslash
characters**, and `News listing rule:` begins after a genuine blank line. The
gap-planner template likewise holds zero literal backslashes.
> A first attempt at this check used `LIKE '%\\n%'` and returned a meaningless
> `true` — **`LIKE` treats backslash as an escape character**, so that pattern
> matches the ordinary JSON escape. Counting `chr(92)` occurrences in the
> extracted *value* is the check that discriminates.

### 3. Branch check — PASSED, both arms, against the live binary

The retype arm cannot occur naturally: `ai-agent-orchestration.com` is the only
site with `separate_page=true` and no `news-index` page, and its stranded-candidate
set is empty (so it would take the `new_page` arm). The arm was therefore
**induced deliberately**, using the scratch one-step probe harness from
`durable_write_guard/RUNBOOK_durable_write_guard.md`, on
`pool-travel-leisure.internal` — `status='pool'`, 0 pages, never deployed, and a
`.internal` domain, so a build could not reach the public internet. Owner
approved the contained induction.

Fixtures: two identically-stranded pages (`actualidad-index`, `servicios-index`;
nav-visible, `sections=[]`, `build_status='planned'`) and one `missing_news_page`
item whose spec authorised **only** `actualidad-index`.

| arm | probe | result |
|---|---|---|
| **Refusal** (the failing branch) | plan names `servicios-index`, which the spec does NOT list | `applied=false`, reason `page "servicios-index" is not in retype_candidates [actualidad-index] — refusing to re-type a page the check did not identify as stranded`. **Both** pages unchanged (`updated_at` still the fixture timestamp), no work item filed, original item untouched. |
| **Happy path** | plan names `actualidad-index` | `applied=true, from_type=section-index, to_type=news-index`. Page flipped to `news-index` with sections `["hero","news-listing","call-to-action"]` — the type-keyed archetype, not the name heuristic. Control page `servicios-index` still `section-index`/`[]`. `needs_content_page` item filed for `page-build-handler` carrying `retyped_from=section-index`. Original item → `complete`, `handled_by=content-gap-planner`. |

The control page is what makes this a discriminator rather than a green light:
the same run that re-typed one stranded page left an equally-stranded one alone,
because only one was in the authorising set.

> **The completion leg needed a second run, and the reason is worth recording.**
> On the first happy-path probe the originating item stayed `detected`.
> That is not a defect: `markOriginalComplete` updates
> `WHERE status IN ('triaged','claimed')`, and my fixture had bypassed the
> dispatcher that would normally have triaged it. Re-running from `triaged`
> completed it correctly. A fixture in a status the real pipeline never presents
> silently skips the code you meant to test.

All probe artefacts removed afterwards (fixtures, scratch agent, 4
`orchestration_states` + 25 audit rows, error-log rows); leak check returned
**0 on all seven lines**.

### Tests re-run on current clean HEAD

`git archive HEAD` with no overlay (all files committed): `datahelpers`,
`actions`, `discovery_checks`, `queryresolve` all `ok`. The nine 015 tests pass,
including one subtest per flattening rule.
> **CORRECTION to the 2026-07-24 entry above:** it recorded a *pre-existing*
> `discovery_checks` verifier-coverage failure on pure HEAD
> (`backend_entry_orphaned`/`contact_form_undeliverable` lacking verifiers).
> That is **gone** — another thread's `bugs_closed/021` INSTANCE 2 work added the
> missing verifiers. The suite is clean on pure HEAD now.

### Council trail — ADVISORY, and it never reached APPROVED

Two rounds on `SUBMISSION_CORR=45664479-31bd-4065-b89e-a7a09f9cabf1`, both
**REVISE**. No `Council-Reviewed:` trailer is claimed anywhere: per the standing
rule the trailer is earned by an APPROVED verdict only, and the code commit
(`55402bc5e`) shipped before either verdict in any case, so the `098` join is a
miss for this change regardless.

**Round 2 is not a sound verdict: `unreadable: 1`.** The lost seat was
`editquality` — the one seat whose single round-1 objection had been answered.
`reviewers(11) + abstained(4) = 15`, not the 16 a healthy round shows. Per the
gate runbook's own check (`unreadable` MUST be 0) this round lost an opinion we
were owed; it is the `bugs_closed/019` shape.

The round-2 objections are non-blocking in substance and are recorded here
because two of them are **owner decisions, not defects**:

- **`guardian`** — *"much stronger resubmission… not vetoing"*. Asked that the
  adoption-LLM surface be filed as its own follow-up rather than block this edit.
- **`bug_historian`** — *"I don't think this rises to a blocking architectural
  concern"*. **Owner decision requested:** whether to require a generic fail-loud
  *"a role was flattened, and here is what got flattened"* log as a follow-up work
  item, on the argument that without it the next occurrence — under a different
  role name — is again found only after content is silently lost. This is a fair
  point and is NOT addressed by this fix.
- **`prior_art_librarian`** — *"if they come back clean, approve outright on next
  pass"*. Its check: the "no execution path exists" claim had been verified only
  inside `content-gap-planner`'s own file, not platform-wide. **Check run, clean:**
  across `platform/ internal/ pkg/` exactly two writers of `pages.page_type` exist
  — `apply_gap_plan_action.go:592` (this change) and
  `rebuild_blog_listing_action.go:350` (`SET page_type = 'blog-index'`, hardwired
  to the blog case). There was no dormant re-typer to reuse.

### Blast-radius homework (round-1 guardian objection), answered

- `ValidateRoles`: exactly two non-test callers, `site_db_actions.go:276` and
  `write_site_plan_action.go:246` — both planner persist paths, inside the quartet
  the submission named.
- `CanonicalisePage`: six non-test call sites in five files, **three outside** the
  quartet — `create_tool_component_action.go:246` (`Role: "tool"` literal, so the
  new branch is unreachable), `check_tool_recreation_needed.go:251` (result used
  only as a map key), `apply_adoption_plan_action.go:446` and `:759` (`rawType`
  from the adoption LLM).
- Decisive fact: **zero active agent definitions contain the string `news-index`
  or `news_index`** [VERIFIED 2026-07-26], so no LLM surface — adoption included —
  could emit that role before migration 206. And `isSectionIndexRole` already
  contained `blog-index`, so a `news-index` emission yields the **same name and
  URL** a `blog-index` emission already yielded; only `page_type` differs, which
  is the whole fix. **No new name/URL shape enters any pipeline.**

### Residuals — all recorded; ONE of them has a live instance

0. **`bugs_open/081` (filed today) — THE IMPORTANT ONE. A *deployed* mistyped page
   has no repair path, and there is a live instance looping.** This fix covers the
   *stranded* case (never-deployed, sectionless) and the *newly-planned* case. It
   does **not** cover a page that is mistyped and already deployed:
   `findStrandedNavPages` excludes `build_status='deployed'` (correctly — that
   clause is the 2026-07-20 correction above), so no candidate is offered; and the
   `new_page` fallback's `ON CONFLICT (site_id, name) DO UPDATE` sets `title`,
   `sections`, `updated_at` but **not `page_type`**
   (`apply_gap_plan_action.go:394-404`), so it overwrites a live page's content and
   leaves the mistype in place. The check then fires again next sweep.
   **Live:** `ai-agent-orchestration.com` `/news.html` is `page_type='content'`,
   deployed, with a `detected` `missing_news_page` item whose `spec.page_name` is
   `news` — the existing row's name, so the `ON CONFLICT` branch is the one that
   fires. A previous item for the same check went `unresolved` on 2026-05-01, so
   this has been looping ~3 months. `idea.uk` `/news/index.html` is the second
   instance (`section-index`, deployed).
   > **This corrects the scope claim in the 2026-07-20 section above.** That section
   > concluded "zero rows … so this is a guard against recurrence, not a repair of
   > live damage". That was true *of the stranded predicate* and false as a
   > statement about the bug class: the predicate never asked whether a **deployed**
   > page was already occupying the role under the wrong type — and on the one site
   > it measured, one was. The measurement was sound; the question was too narrow.

1. **`bugs_open/080` (filed today)** — `applyNewPage` bypasses `CanonicalisePage`
   (`url := "/" + pageName + ".html"`, `apply_gap_plan_action.go:355`), so the
   gap-planner and planner surfaces disagree on a page's name/URL.
   Pre-existing and identical for `blog-index`; this change neither creates nor
   widens it. Live example: `gaswholesalers.com`/`robot-hands.com` hold
   `news` at `/news.html` where the planner would canonicalise to
   `news-index` at `/news/index.html`.
2. **"At most one `news-index` per site" is assumed but not enforced.**
   `render_news_section_action.go:213-217` selects the listing page with
   `LIMIT 1` and **no `ORDER BY`**, so two would make the choice arbitrary. Rule 1b
   deliberately stops correcting an explicit `news-index`, so the only guard is
   the prompt wording ("plan exactly ONE news listing page"). Live state is one
   per site on all three sites that have one [VERIFIED 2026-07-26]. Cheap
   structural guard if it ever bites: a partial unique index on `(site_id)
   WHERE page_type='news-index'`.
3. **`isTypedIndexRole` stays narrow** (`news-index` only). `entity-directory`
   has the same exposure with no observed incident; `blog-index` cannot join
   because `normaliseRole` collapses it earlier and
   `rebuild_blog_listing_action.go:326` auto-repairs that case downstream.
   Widening is one line per role — the cost of waiting for evidence is small,
   the cost of trusting every flavoured index role today is that sloppy
   emissions stop being corrected.
4. **Owner decision owed** on `bug_historian`'s generic flattening log (above).

**Closing bar met, and what it does and does not mean.** The defect this file
names — `ValidateRoles` deterministically flattening an explicit `news-index` to
`section-index`, orphaning the page from `render_news_section`,
`MissingNewsPageCheck` and `page-build-handler` at once — is fixed in code that is
live, its config half is applied, and both arms of the new executor have been
induced and observed against the running binary. That is the root cause as finally
diagnosed, and it is closed.

**It is not the whole class.** Residual 0 (`bugs_open/081`) is the deployed-mistyped
half, with a live looping instance, and it is a genuinely different mechanism — a
pre-existing wrong row that no repair path covers — rather than unfinished work on
this one. It is filed with full evidence rather than left inside a closed file
where nobody would look. Read `081` before assuming this class is retired.

Moving to `/bugs_closed/`.

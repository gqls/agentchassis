# 080 — `content-gap-planner`'s `new_page` bypasses `CanonicalisePage`, so two creation surfaces disagree on a page's name and URL (FIX LIVE on v1.0.1177 — OPEN for the duplicate rows it does not repair)

> ## STATUS 2026-07-27 — candidate 1 applied, council APPROVED, **LIVE on v1.0.1177**
>
> **Rolled 19:22:02Z.** Verified in the running pod, not the tag:
> `strings /app/agent-chassis | grep -c "failed canonicalisation"` → **3**.
> The creation-side divergence is closed. **This case stays OPEN for its residual:**
> robot-hands.com still serves BOTH `/news.html` and `/news/index.html`, one orphaned
> from the nav, and webdesign.co.uk's `news` row still collides on its next plan.
>
> > **CORRECTION 2026-07-27 (owner): "which row survives" is NOT an open question.**
> > I wrote that it was a decision for the owner. It has already been decided, at
> > length, and I should have found it before saying so. The decided shape is the
> > **section-index family convention**, and it is not a preference — it is shipped,
> > council-approved code:
> >
> > - `page_canonical.go` (doc 029 Phase 0, extended by `bugs_closed/015`):
> >   `news-index` joins the section-index family, so the planner shape
> >   `{news-index, noticias}` canonicalises to
> >   **`(name=<section>-index, url=/<section>/index.html, page_type=news-index)`** —
> >   the flavour preserved as `page_type`, which is "the family's stated design".
> > - Doc 029's own table: `section-index (section="tools")` → `tools-index` at
> >   `/tools/index.html`.
> > - The relojistas precedent, decided in its running notes and now live: *"re-type
> >   `noticias-index` → `news-index` and drive the build onto it (keep the Spanish
> >   `/noticias` URL + nav), rather than let gap-planner mint an English
> >   `/news.html`."*
> >
> > **`relojistas.com` is the only site in the fleet that matches it**, measured
> > 2026-07-27: `noticias-index` | `/noticias/index.html` | `news-index`.
> >
> > | site | name | url | page_type | vs the decision |
> > |---|---|---|---|---|
> > | relojistas.com | `noticias-index` | `/noticias/index.html` | `news-index` | ✅ canonical |
> > | robot-hands.com | `news-index` | `/news/index.html` | `section-index` | name+url ✅, **type wrong** |
> > | robot-hands.com | `news` | `/news.html` | `news-index` | type ✅, **the duplicate** |
> > | idea.uk | `news-index` | `/news/index.html` | `section-index` | name+url ✅, **type wrong** |
> > | gaswholesalers.com | `news` | `/news.html` | `news-index` | type ✅, name+url wrong |
> > | ai-agent-orchestration.com | `news` | `/news.html` | `content` | **both wrong** (`bugs_open/081`'s ~3-month loop) |
> > | webdesign.co.uk | `news` | `/news/index.html` | `news-index` | url+type ✅, name wrong, still `planned` |
> >
> > **So for robot-hands the answer is determined:** keep `news-index` at
> > `/news/index.html` (canonical name and URL, and the row the live nav already
> > points at), **re-type it `section-index` → `news-index`** so `render_news_section`
> > gates on it, and retire `/news.html`. That is not a new judgement — it is the
> > relojistas repair applied to the same shape.
> >
> > **What is genuinely still open is the EXECUTION, not the decision**, and it is
> > `bugs_open/081`'s: a *deployed* mistyped page has no repair path, because 015's
> > retype arm requires `build_status <> 'deployed'`. Both robot-hands rows are
> > deployed. So this is blocked on 081's candidate 2, not on a ruling.
> >
> > *The cheap check I skipped:* grep the docs for the convention before calling a
> > settled convention an open decision. `page_canonical.go`'s own header states it.
>
> **Still owed:** the induced proof — dispatch a gap plan with `page_type: "news-index"`
> and confirm the row lands canonical. The unit tests exercise that branch; production
> has not.
>
> **Fix:** `9759687d1` — `applyNewPage` now takes (name, url, page_type) from
> `datahelpers.CanonicalisePage`, and the hand-rolled `url := "/" + pageName + ".html"`
> is deleted. An identity that cannot be canonicalised is refused instead of writing a
> page named `""` at `/.html`.
>
> **Council:** APPROVED at round 1, correlation `1cc2baab-333f-4169-812b-6161b81a13f0`
> (11 reviewers, 5 abstained, 0 unreadable, 1 medium + 2 low advisory objections, no veto).
>
> ~~This case stays OPEN until a chassis image ships it~~ — **shipped 2026-07-27
> 19:22:02Z on v1.0.1177 and verified in the running pod.** It stays open only for
> the duplicate rows above, which no code change repairs.
>
> ### Two corrections to this file, from doing the work
>
> 1. **"Two surfaces" undercounts.** There are **four** canonicalising CREATION call
>    sites, not one: `write_site_plan_action.go:276` (planner),
>    `site_db_actions.go:281` (reconciler), `apply_adoption_plan_action.go:446`
>    (adoption) and `create_tool_component_action.go:250` (tools) — plus two
>    read-only lookups that are not creation surfaces at all
>    (`apply_adoption_plan_action.go:759`, `check_tool_recreation_needed.go:251`).
>    The gap-planner was the fifth creation surface. My own council submission
>    repeated the "fourth" framing and the guardian seat caught it.
> 2. **The dedup-window risk is negligible, measured.** In-flight
>    `gap_plan_new_%` work items are **5**, all `needs_human_review`, so no item is
>    waiting to be re-keyed by the changed `item_key` shape.
>
> ### What this fix deliberately does NOT do — still an owner call
>
> The existing divergent rows are untouched. Re-measured 2026-07-27:
> robot-hands.com still carries **both** `news` (`/news.html`) and `news-index`
> (`/news/index.html`), both deployed and active, the nav pointing only at the
> latter. **webdesign.co.uk** holds `news` at `/news/index.html` — canonical URL,
> non-canonical name, still `planned`, so it collides on its next plan. Which row
> survives, and whether the other 301s, is a live-site content decision.

**Found:** 2026-07-26, while doing the council gate's blast-radius homework for
`bugs_closed/015`. **Not** caused by 015's fix — it predates it and is identical
for `blog-index`. Filed separately because it is the exact failure class
`CanonicalisePage` was created to end, and 015's closing evidence happens to
contain a clean live example of it.

## Mechanism

The platform has two surfaces that create `pages` rows, and only one of them
canonicalises.

- **Planner path** — `write_site_plan_action.go:276` and `site_db_actions.go:281`
  both run the descriptor through `datahelpers.CanonicalisePage`, which maps the
  section-index family (`section-index`, `blog-index`, `news-index`,
  `entity-directory`) to `name=<section>-index`, `url=/<section>/index.html`.
- **Gap-planner path** — `apply_gap_plan_action.go` `applyNewPage` never calls it.
  It lowercases the LLM's name and synthesises the URL by hand:

```go
// apply_gap_plan_action.go:294
pageName = strings.ToLower(strings.ReplaceAll(pageName, " ", "-"))
// apply_gap_plan_action.go:355
url := "/" + pageName + ".html"
```

So for the same logical page the two surfaces produce different rows.

## Live evidence (queried 2026-07-26)

Three sites hold a `news-index` page. Two were created by the gap-planner, one by
hand:

| domain | name | url | how it was created |
|---|---|---|---|
| gaswholesalers.com | `news` | `/news.html` | gap-planner `new_page` |
| robot-hands.com | `news` | `/news.html` | gap-planner `new_page` |
| relojistas.com | `noticias-index` | `/noticias/index.html` | hand-fixed (015 workaround), i.e. the canonical shape |

A planner-emitted news listing on either of the first two canonicalises to
`news-index` at `/news/index.html`, which does **not** match the existing `news`
at `/news.html`.

## Why it matters

`pages` has `ON CONFLICT (site_id, name)`. A name that disagrees does not
conflict — it **inserts a second row**. That is precisely the duplicate-page
failure `CanonicalisePage`'s own header documents itself as having been written
to prevent:

> *"Without this convergence, both surfaces upserted into pages with disagreeing
> names, the `ON CONFLICT (site_id, name) DO UPDATE` never fired, and we got
> duplicate rows plus parallel work-item streams that `idx_swi_dedup` could not
> collapse."* — `page_canonical.go` header, Phase 0 (doc 029)

The convergence was built for adoption-vs-planner and the gap-planner surface was
never brought into it.

## Scope and severity — measured, not assumed

**No duplicate exists today.** [VERIFIED 2026-07-26] The two `/news.html` sites
are the only exposure, and neither has been re-planned since. This is a latent
trap, not live damage — which is why it is filed rather than hot-fixed.

> ## CORRECTED 2026-07-27 (triage sweep) — **the duplicate already existed when this was filed, and it is live**
>
> `robot-hands.com` carries **both** rows, and has since before this case was
> written:
>
> | name | url | page_type | build_status | status | created_at | nav |
> |---|---|---|---|---|---|---|
> | `news-index` | `/news/index.html` | `section-index` | deployed | active | **2026-07-08** | header+footer |
> | `news` | `/news.html` | `news-index` | deployed | active | **2026-07-10** | header+footer |
>
> Both serve **HTTP 200**. Both were re-deployed within 12 seconds of each other at
> 2026-07-27 08:10. They are near-duplicates of one another — identical `<title>`
> (*"Robotics & Automation News | Gripper Industry Updates | Robot-Hands.com"*) and
> the same listing under two headings (*"Gripper Industry News & Updates"* vs
> *"Gripper Industry News & Automation Updates"*). The live nav points only at
> `/news/index.html`, so **`/news.html` is an orphaned live duplicate** — reachable
> by URL and by search engines, linked from nothing.
>
> **Why the original measurement missed it.** The table above this box lists
> robot-hands.com once, as `news` at `/news.html`, because the survey selected
> *"sites holding a `news-index` page"* by **`page_type`** — and the duplicate row is
> typed `section-index`, so it fell outside the filter. The filter defined the
> conclusion. Re-run it by name/URL shape, never by `page_type`:
>
> ```sql
> SELECT s.domain, p.name, p.url, p.page_type, p.build_status, p.created_at
> FROM pages p JOIN sites s ON s.id=p.site_id
> WHERE p.page_type IN ('news-index','section-index')
>    OR p.name LIKE '%news%' OR p.url LIKE '%news%'
> ORDER BY 1,2;
> ```
>
> **What changes.** Severity moves from *latent trap* to *live damage on one site*,
> and the "reversal trigger" below is spent — it already fired, before filing. The
> fix-candidate ordering is unaffected (candidate 1 is still the honest one), but the
> decision it defers — *what to do with the existing rows* — is no longer
> hypothetical: it is a concrete call on which of robot-hands.com's two news pages
> survives and whether the other 301s.
>
> `webdesign.co.uk` is worth a second look for the same reason: it holds `news` at
> **`/news/index.html`** (canonical URL, non-canonical name, `page_type=news-index`,
> still `planned`). That row would collide with a canonical `news-index` name on the
> next plan — the same trap, one step earlier.

**Reversal trigger:** a re-plan of `gaswholesalers.com` or `robot-hands.com` that
emits a news listing. After migration `206_planner_news_index_page_type.sql`
(applied 2026-07-25) the planner *can* now emit `news-index`, so the trigger is
live where before it was unreachable. Watch for a second news page appearing
beside the existing one.

## Fix candidates (none applied)

1. **Route `applyNewPage` through `CanonicalisePage`** — the structural fix, and
   the one consistent with the helper's stated purpose. Risk: it changes the
   name/URL of *future* gap-created pages, so existing rows must be left alone
   (no data migration) or the two sites above get re-pointed deliberately.
2. **Canonicalise on read** where the two surfaces meet, leaving creation alone.
   Cheaper, but leaves the divergence in the data.
3. **Leave it and add a detector** — a check for "two pages on one site whose
   canonical form collides". Weakest, but it converts a silent duplicate into a
   work item.

Candidate 1 is the honest one; it needs a decision about the two existing rows,
which is why this is filed rather than fixed.

## Related

- `bugs_closed/015` — where this was found; its §"CanonicalisePage callers"
  enumeration is the grep evidence.
- `page_canonical.go` header — the design intent this violates.
- `bugs_open/052`, `bugs_open/049` — listings/links advertising pages that do not
  resolve; a duplicated page row is one way to feed them.

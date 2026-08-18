# 309 — fundamentallyai.com's Platform Log index renders SIX cards with ZERO anchors, so every article it advertises is unreachable from the site

**Filed 2026-08-18** by the 279/284/290 thread, while investigating an owner
observation about the same page. **Status: OPEN. Symptom MEASURED at the served
artefact with a positive control; ROOT CAUSE NOT DIAGNOSED — no `090` run yet, and
this file does not assert a mechanism.** Prior art checked: no bug in `bugs_open/`
or `bugs_closed/` mentions `bl-card`, unlinked cards, or a section index that does
not link its members.

> **STATUS SUPERSEDED 2026-08-18 (later) — read the ADDENDUM at the foot of this
> file before this header.** The mechanism is now measured end to end, the "control"
> below is corrected, the `090` correlation cited further down **names a run that
> does not exist**, and a real diagnosis is in flight under run correlation
> `6e578bf5-778a-4e72-aab2-0531e45c07d8`. Still OPEN, still unfixed.

## The symptom, measured at the served page

`https://fundamentallyai.com/platform-log/index.html` (HTTP 200, 32.6 KB) renders
**6 `<article class="bl-card">` cards, and not one contains an `<a>` element**:

| # | card title | anchors |
|---|---|---|
| 1 | Self-Correction: leopardessconsulting.co.uk | 0 |
| 2 | How to Use the LLM Cost Calculator | 0 |
| 3 | How to Use the Review Council Simulator | 0 |
| 4 | How to Use the AI Readiness Checker | 0 |
| 5 | How to Use the Automation Savings Estimator | 0 |
| 6 | How to Use the Model Approach Selector | 0 |

Each card is otherwise complete — image with alt text, category chip, date, read
time, title, excerpt. The page has 31 anchors in total (chrome, footer) and **2 in
`<main>`, both pointing at `/tools/…`**, neither at any article.

**It is not "empty hrefs" and not "phantom links":** there is no anchor element to
carry an href. So `empty_internal_href` (13 rows, page-build-handler) and
`phantom_internal_link` (51 rows) are both the WRONG item type for it, and filing
either would send it to a remit that does not cover it.

⚠ **A measurement correction inside this filing, because it nearly became the
finding:** a first pass on one truncated card string reported "anchors in card: 1".
Re-running the same regex over ALL SIX cards returned 0 for every one. A regex over
a `[:900]` slice is not a measurement of the card.

## The control — the template CAN do this, so it is not fleet-wide

`[MEASURED 2026-08-18]` another site's section index, same card idiom:

| page | cards | cards containing an anchor |
|---|---|---|
| `mortgagecalculator.co.uk/investor/index.html` | 6 | **6** |
| `fundamentallyai.com/platform-log/index.html` | 6 | **0** |
| `idea.uk/news/index.html` | 0 | – (different markup; inconclusive, claims nothing) |
| `vetcomparison.uk/guides/index.html` | 0 | – (same) |

So the card component renders working links elsewhere. Whatever is wrong is about
this page's data, its template variant, or its render — not the component as such.

## Why it matters

Five of the six cards are the tool guides. The site sells the tools; the guides are
the writing that explains them; and **nothing anywhere on the site links any of
them** — checked at `/`, `/platform-log/index.html`, `/tools.html` and
`/capabilities.html`, none of which contains a single `/blog/…-guide` or
`/guides/…` href (the homepage links exactly one blog post, and it is not a guide).
So the guides are reachable only by direct URL or search, while the index that
exists to route readers to them lists them as inert text.

**Card 4 has a second problem behind the first:** "How to Use the AI Readiness
Checker" corresponds to `pages.name = 'ai-readiness-checker-guide'`, whose status is
**`archived`** (its live sibling is `/guides/tool-ai-readiness-checker-guide.html`).
So the index is advertising an archived page, and simply restoring anchors would
give that card a 404 unless it is repointed.

## What this bug is NOT (an owner hypothesis, measured and corrected)

The owner asked about "duplicate guides", believing three tools had two guides each
and that duplication was unintended. **They are not duplicates.** Each pair is two
DIFFERENT articles about the same tool:

| tool | `/blog/…` article | `/guides/tool-…` article |
|---|---|---|
| Automation Savings Estimator | "How to Use the Automation Savings Estimator" | "How the AI Automation Time Savings Estimator Works" |
| Model Approach Selector | "How the Model Approach Selector weighs fine-tuning, RAG…" | "Prompting, RAG, or fine-tuning: a decision guide…" |

A usage guide and a conceptual/decision guide. **Archiving either would destroy real
content, not remove a copy** — so no de-duplication work was filed. Whether the site
WANTS two pieces per tool is a content-strategy question for the owner, not a defect.
Inventory as at 2026-08-18: 7 active guides (4 under `/blog/`, 3 under `/guides/`)
plus 1 archived; only the `/blog/` "How to Use" set appears on the index.

## Candidate mechanisms — all `[UNVERIFIED]`, which is why this is not a fix

1. The listing template variant used by this page emits a title but no wrapping
   anchor (compare against whatever `mortgagecalculator.co.uk/investor/index.html`
   renders with).
2. The listing data carries no resolvable URL per item, and the template omits the
   anchor rather than emitting an empty one — which would make the archived card 4
   the visible tip of a wider resolution failure.
3. A rebuild regenerated this index from a source that lost the per-item hrefs.
   Every guide page's `updated_at` is 2026-08-17, so something rebuilt them all
   recently; the index's own render date has not been established.

**Next step: a `090` diagnosis run** (symptom: the six-card / zero-anchor
measurement above plus the working control; point it at the platform-log-index
page row, `page_components.rendered_html` for that page, and whatever renders
`bl-card`). Check the `needs_diagnosis` queue first.

## How to verify a fix

At the SERVED page, never the stored HTML (this thread's own landmine): all six
cards contain an anchor whose href resolves to an active page — and card 4 points at
the live `/guides/tool-ai-readiness-checker-guide.html`, not the archived
`/blog/ai-readiness-checker-guide.html`. The one-liner that produced the table above
is in this file's history; it counts `<article class="bl-card">` blocks and the
subset containing `<a`.

---

# ADDENDUM 2026-08-18 (later) — mechanism MEASURED, two corrections to the filing above, and an owner ruling that closes the content question

## 0. Two corrections to this file, before anything else

> **CORRECTED — the `090` run this file points at DOES NOT EXIST.** The section
> "Next step: a `090` diagnosis run" and the note in the lane handoff both cite
> correlation `df8ca3a1-9cca-474a-88fb-19577e088080` and tell the next session a
> verdict should be waiting. There is no run under that id and there never was.
> `[MEASURED 2026-08-18 18:22Z]` all three tables it would have to appear in are
> empty for it — `diagnosis_artifacts` 0 rows, `orchestration_states` 0 rows,
> `site_work_items` 0 rows matching `spec::text LIKE '%df8ca3a1%'` — and **no work
> item of any type was created anywhere in that hour** (`created_at BETWEEN
> 2026-08-18 17:00Z AND 18:10Z` → 0 rows), so it is not a wrong-key lookup either.
> It also cannot be a printed-then-refused id: `090` generates its correlation at
> line 333 and the coverage refusal `exit 1`s at line 320, so a refused invocation
> prints no correlation at all. **What caught it:** querying the id before acting
> on the instruction to "read the verdict first". Logged in `WRONG_CALLS.md`.
>
> A real run is now in flight — see §4. Note the standing trap it sits on, already
> in `LANDMINES.md`: a `needs_diagnosis` item carries **two** correlations, and the
> artefacts are keyed by `spec->>'dispatch_correlation_id'`, not by the intake id
> the script prints first. Querying by the wrong one returns 0 rows with no error,
> which is indistinguishable from "still running".

> **CORRECTED — the control above is weaker than it reads.**
> `mortgagecalculator.co.uk/investor-index` renders 6 cards with 6 anchors, but
> `[MEASURED]` it is built from the **`tool-list`** component, not `blog-listing`.
> It is the same `page_type` (`section-index`), so it does establish that a
> section-index page can carry a linking listing — but it says nothing about the
> component actually on the broken page. The exact control is in §2.

## 1. The mechanism, measured end to end

`[MEASURED 2026-08-18]` The page carries two components; the listing one is
`blog-listing_pre_037` (`content_components.id = 4b097683-…`), instance
`page_components.id = 79d769e4-…`.

**The template does emit a per-card anchor, but gates it:**

```
{{if .post1_url}}<a href="{{.post1_url}}" class="bl-read-link" …>{{.read_more_label}}</a>{{end}}
```

**The gate key is absent from the data.** The instance's `content_data` holds
every key the template names — titles, dates, excerpts, categories, read times,
image urls, author names, all the labels — **except the seven `{{if}}`-gated URL
keys** (`post1_url`…`post6_url`, `cta_url`). That is a clean partition of the
schema by a property the writer never saw, which is the same tell recorded in
`bugs_open/238`.

**Why those seven and only those seven:** they are the only fields sourcing from
`site_specs.blog.*`, and **the `blog` aspect has never existed on any site**
— `SELECT count(*) FROM site_specs WHERE aspect='blog'` → **0**, with no
`is_current` filter, so that is "never, on any site, ever", not "not currently".
Sibling fields resolve fine: `postN_image_url` sources `site_assets.image` and is
present on every card.

So the chain is: declared source resolves nothing → `on_missing` is undeclared and
defaults to `skip_field` (`plan_sections_action.go:2092-2094`) → the key is omitted
from `content_data` → `{{if}}` drops the anchor. **No empty `href`, no gap, nothing
for a markup-shape check to catch** — which is why five cards of orphaned writing
sat there unnoticed.

Note this defeats `bugs_open/238`'s carry fix by construction: `planSection` carries
a non-llm field from the page's own previously deployed `content_data`, and no
deployed row has ever held these keys, so there is nothing to carry from.

## 2. The exact control — same component, working links

`[MEASURED]` `blog-listing_pre_037` has exactly **two** live instances fleet-wide,
and **neither** carries `post1_url`:

| site | page | `page_type` | `bl-card` blocks in stored html | `<a href=` in stored html |
|---|---|---|---|---|
| fundamentallyai.com | `platform-log-index` | `section-index` | 6 | **0** |
| leopardessconsulting.co.uk | `blog` | `blog-index` | **0** | 11 |

The leopardess row has the **same component_id and the same missing keys**, and its
links work — because its `rendered_html` is not the blog-listing template at all. It
is `section--articles` / `article-card` markup with real hrefs
(`<a href="/guides/tool-automation-savings-estimator-guide.html">…`), i.e.
`RebuildBlogListingAction`'s output, which loads the `content-listing` template and
overwrites the slot.

**That action is the second half of the mechanism.** `findBlogPage`
(`rebuild_blog_listing_action.go:421-475`) selects a target by `page_type='blog-index'`
or `name='blog' AND page_type='content'`. `platform-log-index` is **`section-index`**,
so it matches neither and is never rebuilt. The page is therefore left with the raw
template render — the only state in which the missing keys are visible.

Served-page re-measurement `[MEASURED 2026-08-18 18:34Z]`, HTTP 200, 32,594 bytes:
6 cards, **0 anchors in every one**, 31 anchors on the page (all chrome/footer).

`[MEASURED 2026-08-18 18:41Z]` and the "nothing on the site links a guide" claim still
holds — re-checked because other lanes rebuild these pages hourly. `/`,
`/platform-log/index.html`, `/tools.html`, `/capabilities.html`: **0 hrefs containing
"guide" on any of them.** The control that makes those zeros real (this thread's own
earlier trap was six 404-shaped false zeros): all four returned HTTP 200 with bodies of
59,573 / 32,592 / 27,081 / 44,910 bytes, so each zero is a page that loaded and linked
nothing, not a page that was not there.

## 3. This is a MISSED MIGRATION, not a novel defect — and it explains card 4

`[MEASURED]` Every sibling list component was moved to a collection dialect whose
source resolves live. `blog-listing_pre_037` alone still carries the legacy
numbered-flat one:

| component | items source | per-item url |
|---|---|---|
| `content-listing` | `articles` ← `query.blog_posts` | derived from the page row |
| `tool-list` | `items` ← `query.pages_where_type:tool` | derived |
| `game-list_pre_037` | `items` ← `query.pages_where_type:game` | derived |
| `guide-list_pre_037` | `items` ← `query.pages_where_type:guide` | derived |
| **`blog-listing_pre_037`** | **none — `post1_…`…`post6_…` flat** | **`site_specs.blog.postN_url` (dead)** |

**And this dissolves the "card 4 advertises an archived page" finding into the same
root cause.** In the legacy dialect `postN_title` is `source: llm` — the six card
titles are *written by the model*, not read from `pages`. So a card can name
`ai-readiness-checker-guide` (status `archived`) because nothing ever joined the
title to a live page row. In the collection dialect title and URL come from the same
row and **cannot disagree**. Card 4 is not a separate staleness bug; it is the second
visible symptom of the missing join.

## 4. Blast radius — bounded, and measured rather than assumed

`[MEASURED]` Components declaring a `site_specs.<aspect>` source for an aspect no
site has ever carried: **12 aspects, 61 fields, 15 (component, aspect) pairs**. Only
**three** of those components have any live instance on an active page:

| component | dead aspect | dead fields | live instances | sites |
|---|---|---|---|---|
| `blog-listing_pre_037` | `blog` | 7 | 2 | 2 |
| `testimonials` | `social_proof` | 1 | 2 | 1 |
| `social_proof` | `social_proof` | 1 | 1 | 1 |

The other twelve pairs (`nav`, `pricing`, `inventory`, `categories`, `legal`,
`ctas`, `product`, `search`, `structured_data`, `social`) sit on components with
**zero** live instances — dormant legacy rows, worth a sweep but not biting anything
today. **[MEASURED]** the two `social_proof` fields are **not** this shape: both are
`testimonials`, `type: array`, `required: true` — no anchor to gate, and their three
live instances are all on `gaswholesalers.com`, where one page carries the key with
three entries and two rows have **`content_data` NULL entirely**, which is a different
condition (never populated) rather than a field dropped by `skip_field`. Different
lane; noted, not chased.

So exactly **one live page** is broken by this today: the one this bug is about.
The leopardess instance is rescued only incidentally, by being the page type the
rebuild action happens to select.

**`090` diagnosis in flight** (this file must not assert the root cause until it
returns — CLAUDE.md, owner ruling 2026-07-31). Intake `5214334f-4b21-4c37-8886-46d15f28ba7f`;
**run correlation `6e578bf5-778a-4e72-aab2-0531e45c07d8` — query the artefacts by
THIS one.** Dispatched with `FORCE=1` after reading the three covering items: a
`head_essentials_missing` item on the same page (different remit — head/meta, not
card anchors), a `design-audit` item parked in `needs_human_review` since 2026-07-24,
and a **failed** `needs_diagnosis` from 2026-08-14 sharing only the seed file
`plan_sections_action.go` and asking a different question (whether `source=renderer`
fields bypass the carry; these are `site_specs.*` fields, which do reach the
resolver). Seed files verified byte-identical between `origin/087_towards_multiple_domains`
and local HEAD, so the 273-commit gap does not make this a stale-tree diagnosis.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Migrate `blog-listing_pre_037` to the collection dialect** — repoint the listing
   at `query.blog_posts` (what `content-listing` already uses and what
   `RebuildBlogListingAction` already renders) so titles and URLs come from the same
   page rows. Kills the orphaned links AND the fabricated/archived title in one
   change, and it is the migration its three siblings already had. Requires the
   template to move from six numbered blocks to a range.

   `[MEASURED 2026-08-18]` **this candidate is concrete, not hopeful** — the target
   source was checked against this site rather than assumed. `query.blog_posts` is
   `resolvePagesWhereType(site, 'blog-post', listedOnly=true)`
   (`queryresolve.go:175`), whose predicate is `p.status IN ('active','deployed')`
   plus `ListedPageEligibilitySQL` (`deployed_at IS NOT NULL`, `sections` a non-empty
   array). Run against fundamentallyai.com that returns **8 articles**: all nine
   `blog-post` pages are deployed and sectioned, and the one it drops is exactly
   `ai-readiness-checker-guide` (`status='archived'`) — **so this candidate fixes
   card 4 by construction rather than by a second repair step.** Each row carries
   `p.url`, ordered by `COALESCE(nav_order,100), name`.

   ⚠ **And the codebase already believes this is done.** `queryresolve.go:167` reads
   *"Article listings (content-listing, **blog-listing** components declare
   `source: "query.blog_posts"`)"*. The live `content_components` row does not — it
   still declares `site_specs.blog.postN_url`. The comment is not wrong about intent;
   it is wrong about the fleet, and it is the kind of stale comment that makes a
   reader stop looking. Whoever lands this fix should correct that line in the same
   commit.
2. **A daily check that no `content_components.input_schema` field sources a
   `site_specs` aspect that no site carries.** This is live DB config, so a
   pre-commit hook cannot gate it (see the RFC_006 note in memory) — it wants the
   `SingleOwner`/CronJob shape. Turns 61 silently-dead fields loud.
3. **Widen `findBlogPage` to select `section-index` pages carrying a listing slot.**
   Fixes this page without touching the component — but it is new authority on a
   shared seam (every `section-index` page fleet-wide would enter the rebuild), so
   per the owner ruling of 2026-08-02 §2 it ships as an opt-in field defaulting OFF,
   and it leaves the dead schema in place for the next page that lands on it.
4. **Seed a `blog` spec aspect for this site.** Hand-authored URLs that must be
   re-edited whenever an article is added, on an aspect nothing else in the fleet
   uses. Weakest — it makes the page work and leaves the trap fully armed.

## 6. OWNER RULING 2026-08-18 — the content-strategy question is settled

The question this file raised under "What this bug is NOT" — whether the site wants
two articles per tool — is **answered: two or three articles per tool is fine, it is
not a strict rule.** So the pairs stay, no de-duplication work is to be filed, and
the archived `ai-readiness-checker-guide` is not evidence of a duplication policy.
Nothing about the ruling changes §1–§5: the defect is that the index links none of
them, whether there are one, two or three.

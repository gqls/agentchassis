# CONTRIB 2026-08-31 — from the boxingonline.com build: four experience defects that every existing check passed

**To:** experience_loop (you own the layer that asks what a VISITOR actually gets, rather
than whether each artefact is individually well-formed).
**From:** the session reviewing boxingonline.com — `d2aa5206-73bc-4707-a69c-2702c1eb9152`,
order BR-9AUZ59, **the first paid customer build**, built 2026-08-31, owner-reviewed the same
evening. All figures `[MEASURED 2026-08-31]` against the live DB and
`https://boxingonline.ugg2.com`. Pipeline ownership sits with the site_delivery_and_editor
lane; this is analysis, not a dispatch.

---

## Why this lands in your lane specifically

Your SUMMARY's founding observation was *"every check we had was looking at one artifact in
isolation, so all three were invisible."* Here are four more of exactly that shape, on a site
where **every page validated `valid=true, issues=0`** and every work item that was going to
complete, completed.

## Defect 1 — the home page's editorial slot advertises four EXPLAINER GUIDES as news

The `content-listing` component on `/index.html` is titled **"Latest from the ring"**, with the
subtitle *"Fresh news, previews and results from around the boxing world…"*. Its four items are:

| title as displayed | url | image |
|---|---|---|
| Understanding Boxing Quiz — Test Your Knowledge \| Guide | `/guides/tool-boxing-trivia-quiz-guide.html` | `""` |
| Understanding Fight Night Countdown Timer \| Guide | `/guides/tool-fight-countdown-guide.html` | `""` |
| Understanding Fighter Comparison Tool \| Guide | `/guides/tool-fighter-comparator-guide.html` | `""` |
| Understanding Boxing Weight Class Finder \| Guide | `/guides/tool-weight-class-finder-guide.html` | `""` |

The owner's words: the quiz *"doesn't link through to the quiz but instead through to an
unnecessary explanation page — the guide."* He later found the tool and refined it to the
sharper version: **the guide is more prominent than the tool.**

He is right, and it is structural:

```sql
SELECT name, url, in_header, nav_label, nav_order FROM pages
 WHERE site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' ORDER BY in_header DESC, nav_order;
```
The four **tool** pages have `in_header = true`, `nav_order = 200` and **`nav_label` NULL** — so
they cannot render in the nav. The four **guide** pages are `in_header=false, in_footer=false`,
i.e. orphaned — except that the listing resolver put them in the single most prominent editorial
block on the site. Net effect: the thing a reader can USE is reachable only from a card grid
below the fold; the thing that EXPLAINS it leads the page.

**Why it happened is not a mystery and is not the resolver's fault alone:** the plan produced
zero real articles (`bugs_open/419`), so an editorial listing with nothing to list took whatever
pages existed. **A listing that silently substitutes a different content class is the exact
failure your layer exists to catch** — no single artefact is malformed. The guide pages are
valid. The listing is valid. The promise "Latest from the ring" is broken.

Suggested check, in your promise-keeping family: **a listing's items must belong to the content
class its heading promises.** Cheap first cut — a listing whose heading/subtitle says
news/articles/latest must not be populated by pages whose `page_type`/role is `guide` or `tool`.
It would have fired here, and it fires on the same shape anywhere else.

## Defect 2 — an index page with nothing to index writes a manifesto instead of saying so

`/articles/index.html` serves 3,114 characters of body copy and **zero articles**. What it
contains instead: *"What's in the mix"*, *"Every weight, every corner of the world"*, *"Keeping
it accurate"*, *"Where to go next"* — four headed sections describing the editorial standards of
a section that does not exist yet.

The owner: *"filled with explanations about what we're doing … and very little that the user
would find of benefit."*

Note the failure mode: **emptiness did not present as emptiness.** It presented as a
well-written, in-voice, fully-validated page. `bugs_open/419` records the same property on the
CTA side — link repair rewrote the dead editorial CTAs to point at existing pages, so nothing
looked broken. Two independent honesty mechanisms tidied the absence into a clean page.

The copy half is filed with copy_quality_two_stage. **The experience half is yours:** an index
whose listable-item count is zero is a journey that does not reach an end, and it is currently
indistinguishable from a healthy one.

## Defect 3 — the tools ask the reader for the data we were supposed to bring

Owner, on the fighter comparator: *"That fighter comparison tool requires the user to input all
the details. That is not user friendly, we should make the comparisons just from the name and
include all that information from our research instead."*

Measured on the served page: **18 manual inputs** — for each of two fighters, a free-text name
plus wins, losses, draws, KOs, reach, height, age and a five-character form string. Zero fighter
data ships with the page:
```
curl -s .../tools/fighter-comparator/index.html | grep -c -i "usyk\|fury\|joshua\|canelo\|inoue"   -> 0
curl -s .../tools/fighter-comparator/index.html | grep -o "const [A-Za-z_]* *= *\["             -> (none)
```
The weight-class finder has the same shape: *enter a weight*, not *name a fighter*.

The `tools` spec shows the tool-suggester reasoning it through carefully and never once asking
**"do we hold the data that would let this answer the reader's question?"** Its own words:
*"The best opportunities are lightweight, sport-specific tools that complement the editorial
content."* Complementing content is the right instinct; a pure client-side calculator is what
you get when data availability is never a selection criterion.

This is a promise-keeping defect in your vocabulary: the card on the home page says *"Put two
fighters side by side and compare records, KO percentage, reach, height, age and recent form"* —
the tool does not do that; the reader does, having already looked the figures up elsewhere.
**Suggested criterion for tool selection AND for your critics: name the reader input and the
site-supplied data separately. A tool whose site-supplied data set is empty is a form, not a
tool.**

## Defect 4 — the quality auditor that exists never ran

The owner asked directly whether we have a quality auditor that should have caught the padded
guide pages. **We do, and it did not run.**

```sql
SELECT type, is_active, created_at FROM agent_definitions
 WHERE type='content-quality-auditor' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- active since 2026-03-06; "Loads the site brief, page content samples, and target audience.
--  Makes one LLM call to assess tone alignment, content…"
SELECT status, count(*), max(created_at) FROM orchestration_states
 WHERE workflow_plan::text ILIKE '%content-quality-auditor%' GROUP BY 1;
--  COMPLETED 49 | FAILED 25 | latest 2026-08-31 15:19Z
```
So it is alive and running today somewhere on the fleet — and **zero of those runs touched this
site** (no orchestration row joins `d2aa5206`). It is not in the new-build path. That is the
"detection works; schedule and dispatch do not" shape again, and on a paid build it is the
difference between shipping the padding and catching it.

Worth noting the 25 FAILED against 49 COMPLETED — a 34% failure rate is its own question and I
have not looked into it.

## What I would ask of your lane

1. **A listing-class promise check** (defect 1) — the cheapest of the four and it generalises.
2. **Zero-item index as a first-class experience failure** (defect 2), distinct from "page not
   built" — the page IS built, which is why nothing flagged it.
3. **A data-backing criterion for tools** (defect 3), applied at suggestion time and re-asserted
   by a critic: reader-supplied inputs vs site-supplied data, and refuse the tool if the second
   set is empty.
4. **Route content-quality-auditor into the new-build path** (defect 4) — or tell us why it is
   not there, because right now it is a capability we pay for and do not receive.

**Owner's fifth ask, which is yours and the content lane's jointly:** he wants the site to
*reward a visit* — *"bookmarkable facts and editorial and infographics and timeline graphs"*,
plus biographies, editorials, and directories (where to watch / suppliers). Note that our own
`vertical_landscape` research for this site recommended most of that unprompted: fighter entity
pages, event pages as first-class objects, how-to-watch as a repeatable format, and a weekly
"what to watch" curated guide named as the differentiation opportunity. **The research knew.
The plan did not carry it.** That gap is written up for copy_quality_two_stage (their
PLAN_2026-08-25_best_in_class_propagation is the standing fix) — but the experience half, "what
does a visitor leave with", is your seat's question and the strategy spec even states the target:
*"A visitor leaves knowing which fights are happening this weekend, what platform to watch them
on, and having read at least one piece of editorial that gave them genuine context or opinion
they didn't already have."* Today a visitor can leave with none of the three.

Related: `bugs_open/419` (zero-section blog-post page — mechanism deliberately undiagnosed,
090 corr `6ebdaf88` in flight by the delivery lane; do not duplicate).

---

# ADDENDUM 2026-08-31 (evening) — defect 1 resolved at the site, and it is a 20-SITE CLASS

Written after the delivery lane fixed boxingonline and I re-measured. **The proposed check in
defect 1 is now backed by three measurements rather than one, and the class census below is the
reason it belongs in your layer rather than in one site's fix list.** All figures
`[MEASURED 2026-08-31]`.

## 1. The listing did NOT self-correct when the articles arrived

I predicted in defect 1 that the guides squatted the editorial slot *because* zero articles
existed. The delivery lane then built all six articles, deployed them, and re-rendered both
pages. **The listing did not change.** Measured on the served pages after the articles were live:

| | before | after articles built + rerendered |
|---|---|---|
| `/articles/index.html` `/blog/` links | 0 | **0** |
| `/index.html` "Latest from the ring" `/guides/` links | 4 | **4** |

So all six brief-promised articles were **built, serving, and linked from nowhere** — orphans.

**This kills the "it will self-correct once supply exists" reading, including mine.** The check
must run against `content_data` or the served page, and must NOT be assumed satisfied by supply
appearing. A listing is not a view; it is baked data that a rerender faithfully reproduces.

## 2. The mechanism — target resolution, not caching

`rebuild_blog_listing` (`platform/orchestration/actions/rebuild_blog_listing_action.go`) is a
registered LOCAL action, no LLM, and it writes `content_data` alongside `rendered_html` by design.
**It is already a step in the live `rerender-pages` workflow** (steps: get_pages,
check_pages_exist, deploy_js_snippets, mark_site_deployed, render_js_snippets,
**rebuild_blog_listing**, create_rerender_items, render_site_components, check_refresh_components)
and is the only agent fleet-wide carrying it. **So it ran on every rerender that night and did
nothing.**

`findBlogPage` (line 539) resolves its target two ways: `page_type = 'blog-index'`, else
`name = 'blog'`. On boxingonline both were **0** — the listing page is `articles-index`,
`page_type='section-index'`. The rebuilder could not find the page, no-opped, and **reported
success**.

That is the sibling of `bugs_open/220` ("rebuilds the container and reads green") one step
earlier: **cannot find the container at all, and reads green.** Worth citing together.

## 3. Why the guides were selectable as articles in the first place

`tool-*-guide` pages carry **`page_type='blog-post'`**. So to anything asking "what are this
site's blog posts?", the guides ARE posts. At build time they were the only such pages in
existence, so a faithful listing took them. Not a resolver bug — a **typing** one.

## 4. Resolved at the site — verified independently

The delivery lane retyped `articles-index` → `blog-index` (which is `findBlogPage`'s own
canonical state — strategy 2 *stamps* that type when it hits) and the four guides →
`guide`, then rebuilt the home slot separately. I re-measured on the served pages rather than
taking the report:

```
/index.html          blog_links=6  guide_links=0  control=12  email=0
/articles/index.html blog_links=6  guide_links=0  control=6   email=0
all six /blog/ targets -> HTTP 200
```

## 5. THE CLASS — 20 sites can currently list nothing but guides

The site fix is right and the convention it diverges from is fleet-wide, so this is now a
question for your layer, not for one lane. Census over active pages:

```sql
WITH x AS (SELECT s.domain,
  count(*) FILTER (WHERE p.name ILIKE '%guide%' AND p.page_type='blog-post') AS g,
  count(*) FILTER (WHERE p.page_type='blog-post' AND p.name NOT ILIKE '%guide%') AS r
  FROM pages p JOIN sites s ON s.id=p.site_id WHERE p.status='active' GROUP BY 1)
SELECT CASE WHEN g>0 AND r=0 THEN 'ONLY guides are blog-posts'
            WHEN g>0 AND r>0 THEN 'MIXED' ELSE 'clean' END, count(*), sum(g) FROM x GROUP BY 1;
```

| shape | sites | guide pages |
|---|---|---|
| **ONLY guides are blog-posts — the boxingonline shape, any editorial listing can show nothing else** | **20** | **167** |
| MIXED — guides listed alongside real posts | 9 | 65 |
| clean | 6 | 0 |

Worst offenders in the first group: webdesign.co.uk (52 guides, 0 real posts),
gamesdesign.co.uk (13), loanandmortgagecalculator.co.uk (10), gaswholesalers.com (9),
mortgagecalculator.co.uk (9), idea.uk (9).

**And it is visible at the artefact, not just in the typing.** Spot-checked two MIXED siblings'
served home pages: dartsonline.com carries 2 guide links beside 12 article links; agritec.uk
carries 1 and 1. The guides are genuinely occupying editorial slots on live sites today.

## 6. The tension, stated plainly, because it is a decision not a bug

Guides-as-`blog-post` is the **measured fleet convention**. boxingonline now diverges from its
siblings deliberately, on the owner's ruling that guides must not squat editorial slots. Both of
these cannot stay true:

- if the convention is right, boxingonline is an exception that will be "corrected" by the next
  session that notices the divergence;
- if the ruling is right, **29 sites are mistyped** and the retype is a fleet migration.

`[UNVERIFIED]` which the estate wants. **My read is that the ruling is right and the convention
is an accident** — nothing chose to file guides as editorial, it is just the nearest existing
type — but that is a judgement, and it is worth a `guide` page_type being made real fleet-wide
rather than each lane retyping in isolation. **Your durable check from defect 1 survives either
answer**, which is the argument for building it first: *a listing's items must belong to the
content class its heading promises*, asserted against the served page.

# 052 — a listing regenerates from the page set and re-advertises pages that were never built

*Found 2026-07-20 by the robot-hands thread while removing a dead tool card. **Proven on one
site; fleet exposure measured and currently low** — see the scale section, which is
deliberately unflattering to this bug's importance.*

**Family:** the *derived-field* family, with `/bugs_open/023` (CTA urls are recomputed
label-blind on every render, so authored edits cannot hold). Same shape, different field:
a listing's contents are **derived**, not authored, so editing them is undone by the next
render. Read 023 first — its addendum has the fuller write-up of the class.

Adjacent but **not** the same as `/bugs_open/015` (a mistyped `page_type` orphans a page
from its machinery). 015 and this share a *symptom* — a live link to a page that was never
built — but 015's cause is the page failing to build, and this one's is the listing not
caring whether it built. Fixing 015 would not fix this.

## What happened

`robot-hands.com`'s homepage tool list advertised five tools. Two of them,
`tool-matchmatrix` and `tool-robot-payload-budget-calculator`, were `build_status='planned'`
and had never been built — both URLs served **404**, and the cards were live on the homepage.

The dead card was removed from `page_components.content_data->'items'` and the page
re-rendered. **The card came straight back**, all five entries restored, at
`updated_at 17:10:23`:

```sql
SELECT jsonb_array_length(content_data->'items'),
       (SELECT string_agg(i->>'name',', ') FROM jsonb_array_elements(content_data->'items') i)
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.name='index' AND ...;
-- 5 | tool-matchmatrix, tool-gripper-payload-calculator, tool-gripper-cycle-time-estimator,
--     tool-grip-force-friction-calculator, tool-robot-payload-budget-calculator
```

The listing is regenerated from the site's tool pages on render. It included the page
because the page **existed** — nothing in that path consults `build_status`.

## Root cause

A listing component's items are derived from the page set with no build-state filter. A page
that is `planned` (never built, no components, URL 404s) is indistinguishable, to the
listing, from one that is `deployed`.

Note the asymmetry that makes this easy to miss: **`needs_rebuild` pages usually still
serve**. All three of robot-hands' other tool pages are `needs_rebuild` and return 200 —
they were deployed once and the flag was set later. So the naive fix ("only list
`deployed`") would wrongly delist working pages. **`planned` is the state that means
never-built**; that is the distinction a fix has to draw.

## Scale — measured, and smaller than it looks

Fleet-wide there are 18 non-deployed `page_type='tool'` pages across 7 sites, but that
number is misleading for the reason above. Filtering to the ones that actually mean
never-built:

| | count |
|---|---|
| `planned`, non-archived, tool pages | **2** (dartsonline `tool-setup-builder`, vetcomparison `tool-compliance-deadline-calculator`) |
| ...of those, confirmed **404** live | **2** |
| ...of those, **linked from any component or nav** | **0** |

```sql
-- the linkage check (returned zero rows)
SELECT s.domain, p.name, cc.name FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
LEFT JOIN content_components cc ON cc.id=pc.component_id
WHERE pc.rendered_html LIKE '%/tools/setup-builder/%'
   OR pc.rendered_html LIKE '%/tools/compliance-deadline-calculator/%';
```

**So: the defect is real and proven, and it is currently biting exactly one site.** The other
two never-built tool pages are unreferenced, so no user can reach them. This is filed for the
mechanism, not for an outage. `[UNMEASURED]` whether non-tool listings (article grids, news
lists, card grids) share the same derivation and therefore the same gap — I only traced the
tool list.

## Containment applied (robot-hands only)

`pages.status='archived'` on the never-built page, which is the house route R6 established
and which the listings demonstrably respect (R6 archived six dead article rows and the hub
then listed exactly the three real guides). See
`docs024_key_docs_latest/robot_hands/SQL_2026-07-20_r4e_archive_unbuilt_tool_page.sql`.

Archiving per site is containment, not a fix: it requires a human to notice a 404 first,
which is the step that failed here for weeks.

## Fix candidates

1. **Filter the listing derivation on build state** — exclude `build_status='planned'`
   specifically, NOT "include only `deployed`" (that would delist the working
   `needs_rebuild` pages, see above). Smallest correct change; fixes the tool list only.
2. **Make it a property of the page set, not each listing.** If several listings derive from
   pages independently, each needs the same predicate and they will drift — the exact class
   `/bugs_open/023` is in. A shared "listable pages" helper is the structural version.
   Check first whether article/news/card listings derive the same way; if they do, prefer
   this.
3. **Emit a work item when a listing would advertise a never-built page**, rather than
   silently dropping it — the drop hides a planning gap (a page planned and never built is
   usually a symptom, cf. `/bugs_open/028`). Pair with 1 or 2; on its own it routes to
   `needs_human_review`, which nothing consumes (`/bugs_open/033`).

Prefer (1) now and (2) if the survey in `[UNMEASURED]` above comes back positive.

> **CORRECTED 2026-07-26 — candidate 1's predicate is wrong, and the survey came back
> positive, so (2) was the right call, not (1).** Candidate 1 (`exclude
> build_status='planned'`) was already superseded by this file's own ADDENDUM below,
> which measured four `needs_rebuild`-and-never-deployed pages that candidate 1 would
> have kept advertising. The shipped predicate is the addendum's, not candidate 1's.
> The `[UNMEASURED]` survey is now **answered — see the STATUS section at the end.**
> One other listing derives from the page set the same way, and candidate 3 (emit a work
> item) remains unbuilt and still gated on `/bugs_open/033`.

## How to verify a fix

- Create a `planned` tool page on a test site, render a page carrying the tool list, and
  assert the card is absent. Then set it `deployed` and assert it appears.
- Regression, and this is the one that matters: assert the three robot-hands
  `needs_rebuild` tool pages **still appear** in the list and still serve 200. A fix that
  quietly delists working pages is worse than the bug.
- Check the rendered page, not `content_data` — the item array is regenerated on render, so
  a DB read taken before the render tells you nothing (this is how the bug was missed the
  first time).

## Related

- `/bugs_closed/023` + its addendum — the derived-field class this belongs to. (Moved to
  `/bugs_closed/` on 2026-07-25; this file said `/bugs_open/023` until 2026-07-26.)
- `/bugs_open/015` — same symptom, different cause; do not merge.
- `/bugs_open/028` — a page-build no-op reporting `complete` is one way pages end up
  `planned` forever.
- `/bugs_open/033` — why a detection-only fix would rot.

---

## ADDENDUM 2026-07-20 (bugfix-049 session) — the predicate should be `deployed_at IS NULL`, not `build_status='planned'`

This file's asymmetry section is right and I relied on it: `needs_rebuild` pages usually still
serve, so "list only `deployed`" would delist working pages. Measured fleet-wide, that holds
exactly — **34 of 34** `needs_rebuild` pages with a non-null `deployed_at` return **200**.

But the conclusion drawn from it, *"`planned` is the state that means never-built"*, is very
nearly right rather than right, and fix candidate 1 (`exclude build_status='planned'`) inherits
the gap:

| population | count | live |
|---|---|---|
| `planned`, never deployed | 18 | 404 |
| **`needs_rebuild`, never deployed** | **4** | **404** |
| `needs_rebuild`, deployed once | 34 | 200 |
| `deployed`, never stamped | 1 | 200 (`idea.uk`, a `/bugs_open/040` shape) |

**Four pages are `needs_rebuild` AND have never been deployed**, and candidate 1 would treat
all four as listable. One of them is `gaswholesalers.com/fuel-pricing-framework.html` —
`bugs_open/049`'s mechanism 2, live 404, linked from **28 live footers**. So the naive
predicate misses the worst confirmed real-world instance of this very class.

The predicate that tracks fetchability is **`deployed_at IS NULL`** (optionally with
`build_status <> 'deployed'` to exclude the single unstamped-but-deployed row). Discriminating
pair, same `build_status`, opposite live outcome:

```
gaswholesalers /fuel-pricing-framework.html  needs_rebuild  deployed_at NULL       -> 404
aao            /tools.html                   needs_rebuild  deployed_at 2026-05-02  -> 200
```

Validated by fetching every row in both populations, not inferred from the flag.
Full evidence in `/bugs_open/049`'s addendum (Correction 2). `[UNMEASURED]`: one exception,
`gamesdesign.co.uk /games/jelly-invaders/index.html`, is never-stamped yet serves 200 — a
different deploy path is the obvious guess and I did not chase it.

If candidate 2 (a shared "listable pages" helper) is built, this is the predicate it should
carry — and `/bugs_open/053` fix candidate 3 wants the same one for chrome nav, which is a
third derivation of the same question.

---

## STATUS 2026-07-26 (bugfix-052 session) — core FIXED & LIVE; residual committed, INERT

Two halves. The half this file was filed for is **fixed and live**. A second derivation,
found by finally running this file's own `[UNMEASURED]` survey, is **fixed in code and
inert until the next image roll** — so the case stays in `/bugs_open/` per the
`bugs_closed/README.md` bar.

### Half 1 — the tool-list derivation: FIXED, and LIVE since v1.0.1146

**This landed on 2026-07-21 and nobody updated this file, so for five days the case read
as untouched while the fix was already in production.** Commit `fe2ba5e52` (v1.0.1146)
added the shared predicate and wired it into the generic listing path:

- `queryresolve.FetchablePageEligibilitySQL` (`queryresolve.go:234`) —
  `AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')`. This is the ADDENDUM's
  predicate, not fix candidate 1's; its doc comment cites this bug by number.
- `pageListEligibilitySQL(listedOnly)` (`queryresolve.go:245`) — the extraction point that
  makes "the generic listing path always has a build-state floor" unit-testable.
- Applied by `resolvePagesWhereType` (tool-list, game-list, guide-list, archetype-grid)
  and `resolvePagesUnderSection`.
- Pinned by `queryresolve/page_list_eligibility_test.go`.

**Verified live**, not inferred from the tag — pod `agent-chassis-f4d46c88d-p6wqc`,
image `v1.0.1165`, grepping a string the fix CREATED plus controls:

```
strings /app/agent-chassis | grep -c "deployed_at IS NOT NULL OR p.build_status"   -> 2
positive control:  grep -c "AND p.page_type = $2"                                 -> 2
negative control:  grep -c "zzz_not_a_real_symbol_052"                            -> 0
```

So this is **fix candidate 2**, the structural one — a shared "listable pages" helper —
not candidate 1. The file's own recommendation ("prefer (1) now") was overtaken by its
addendum before any code was written.

### Half 2 — the `[UNMEASURED]` survey: ANSWERED, and it found one more path

> The open question was: *"whether non-tool listings (article grids, news lists, card
> grids) share the same derivation and therefore the same gap — I only traced the tool
> list."*

**Answer: yes, exactly one — and it is worse than the original bug in one respect.**
Surveyed every page-set-derived listing in the tree:

| path | floor carried | verdict |
|---|---|---|
| `resolvePagesWhereType` (tool/game/guide/archetype) | `FetchablePageEligibilitySQL` | fixed v1.0.1146 |
| `resolvePagesUnderSection` | `FetchablePageEligibilitySQL` | fixed v1.0.1146 |
| `queryresolve` `blog_posts` source | `ListedPageEligibilitySQL` | already strict |
| `section_index_for.go` (hub "Browse All") | `status` only | resolves a hub URL, not a listing |
| news listings, directory trackers | n/a | derive from `content_feed_items` / `directory_entities`, **not** `pages` |
| **`rebuild_blog_listing_action.go`** | **none** | **the gap** |

`rebuild_blog_listing_action.go` bypasses `queryresolve` entirely and writes
`content_data.articles` + `rendered_html` directly. Its selection was:

```sql
WHERE p.site_id = $1 AND p.page_type = 'blog-post'
  AND p.build_status IN ('deployed', 'needs_rebuild')
```

Two defects:

1. **It is fix candidate 1's shape, the one the addendum measured as wrong.**
   `needs_rebuild` does not mean "still serves": four fleet pages are `needs_rebuild` and
   never deployed, and they 404.
2. **No `status` filter at all** — so `status='archived'` rows were listed anyway. This is
   the worse half, because archiving is the containment route this file recommends
   ("`pages.status='archived'` … which the listings demonstrably respect"). That claim was
   true of every listing path **except this one**. Containment did not hold here.

It is live machinery, not dormant: registered at `registry.go:686` and wired as a step in
the `rerender-pages` workflow — confirmed in `agent_definitions`, not just the seed file.

### Exposure — real defect, measured **zero** live bite today

Fleet-wide, 11 blog-post rows were mis-listed by the old predicate: 5 that would 404 and 6
`archived`. **All 11 are on `robot-hands.com`** — the same site, and the same R6 archiving
episode, that produced this bug in the first place.

But `findBlogPage()` resolves on only three sites fleet-wide, and none of them is
robot-hands:

| site with a resolvable blog page | posts listed | would 404 | archived-but-listed |
|---|---|---|---|
| `ai-agent-orchestration.com` | 14 | 0 | 0 |
| `finetuning.uk` | 16 | 0 | 0 |
| `leopardessconsulting.co.uk` | 6 | 0 | 0 |

robot-hands has no `blog-index` page and no `content` page named `blog`, so the action
skips it entirely. **The gap was latent: reproducible in code, not reachable by any live
site.** It would have become live the moment robot-hands got a blog page — which its own
archived `learning-center-index` (a `blog-listing` slot still holding 4 items) shows was a
plausible next step.

Recording that honestly matters more than the fix: this is the second time this file has
had to say the mechanism is real and the outage is not.

### The fix committed (INERT until the next image roll)

`platform/orchestration/actions/rebuild_blog_listing_action.go`:

- Blog-post selection lifted into a package-level `blogPostsQuery` const that **splices in
  `queryresolve.ListedPageEligibilitySQL`** — the same constant `queryresolve`'s own
  `blog_posts` source uses, plus `status IN ('active','deployed')`. Sharing the constant
  rather than restating it is the point: two derivations of the same article set that
  carry separate literals are the drift class `/bugs_closed/023` documents.
- `findBlogPage()` — both strategies gained `status IN ('active','deployed')`. They had
  none, so an archived `blog-index` page was a valid listing target, and Strategy 2 would
  additionally have **written** `page_type='blog-index'` onto an archived row.
- The empty-result branch now logs a **Warn** (not Info) when a listing component exists
  but no post is eligible: an empty set leaves the existing listing untouched, so it keeps
  advertising its old contents. Blanking it on a transient empty read would be worse, but
  it is not a no-op and it is exactly the stale-listing state this bug is about.

New test `rebuild_blog_listing_eligibility_test.go`, DB-free, mirroring
`queryresolve/page_list_eligibility_test.go`: asserts the fetchability floor, the status
filter, the shared-constant splice, and the `p` alias contract.

**The assertions were checked against the OLD query text, not just the new one** — all
three fire on it. A test that only passes on the fixed code proves nothing about whether
it would catch the regression.

### Regression + drop-set, measured against the live DB

Kept set is **unchanged** on all three real blog sites — nothing delisted:

```sql
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.page_type='blog-post' AND p.status IN ('active','deployed')
  AND p.deployed_at IS NOT NULL
  AND jsonb_typeof(p.sections)='array' AND jsonb_array_length(p.sections)>0
  AND s.domain IN ('ai-agent-orchestration.com','finetuning.uk','leopardessconsulting.co.uk')
GROUP BY 1;
-- ai-agent-orchestration.com | 14
-- finetuning.uk              | 16
-- leopardessconsulting.co.uk |  6      (14/16/6 before AND after — 36 of 36 kept)
```

Drop set is exactly the intended six, all archived, all robot-hands:

```
grip-force-friction-calculator-guide  needs_rebuild archived  deployed_at NULL
gripper-cycle-time-estimator-guide    needs_rebuild archived  deployed_at NULL
gripper-payload-calculator-guide      needs_rebuild archived  deployed_at NULL
learning-center-article               needs_rebuild archived  2026-05-10   <- archived on purpose
learning-center-post                  needs_rebuild archived  deployed_at NULL
news-post                             needs_rebuild archived  deployed_at NULL
```

### Re-grounded 2026-07-26 — the ADDENDUM's table has moved, and I had already repeated it

> **CORRECTION.** The commit message and the first draft of the code comment repeated the
> addendum's *"four pages are `needs_rebuild` and never deployed"* straight from this file.
> It was measured on 2026-07-20 and it is **stale six days later.** Re-measured against the
> live DB:

| population | 2026-07-20 (addendum) | **2026-07-26** |
|---|---|---|
| `planned`, never deployed | 18 | **27** |
| `needs_rebuild`, never deployed | 4 | **10** |
| `needs_rebuild`, deployed once | 34 | **31** |
| `deployed`, never stamped | 1 | **1** (`idea.uk`, unchanged) |

The mechanism is untouched — the predicate is still right, and the disjunct still earns its
place. What moved is the size of the population it protects against, and it **more than
doubled in six days**, which is the opposite of the "small and shrinking" impression the
original scale section gives.

Five of those ten are the robot-hands blog posts this fix drops, so they are the same rows,
counted from the other end. The other five (`dartsonline` ×3, `gaswholesalers`, `oufe`) are
`status='active'` and not blog posts, so they are reachable by other listings and belong to
`/bugs_open/049`'s lane, not this one.

The code comment now states the mechanism and **deliberately carries no count**, because a
figure baked into a Go comment cannot be re-grounded and will be wrong within the week.
Logged as the general lesson: a figure copied from a sibling doc inherits its measurement
date, not today's.

### Council gate

Submitted 2026-07-26, `SUBMISSION_CORR = 37329362-f3bb-4ac2-8c72-fd4e9c80109e`
(commit `fe00304bd`). The verdict **post-dates the commit**, so that commit can never
carry a `Council-Reviewed:` trailer honestly and will show as un-reviewed in the 098
coverage report — a permanent false negative, recorded here rather than papered over.
The trailer is earned by an APPROVED verdict only; a REVISE recorded as reviewed would
be a durable false claim.

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='37329362-f3bb-4ac2-8c72-fd4e9c80109e' AND kind='council_report'
ORDER BY created_at;
```

**That first submission was DROPPED, not queued — and the standing advice says not to conclude
that.** The house rule (and `[[council-queue-latency-trap]]`) is that a missing
`orchestration_states` row means latency, ~16–30 minutes, and that resubmitting on that
evidence costs a duplicate round. Correct as a default, and it did not hold here. What
justified overriding it, after 62 minutes:

- **No trace anywhere** — not in `orchestration_states` by correlation *or* by payload
  substring, not in `diagnosis_artifacts`, not in `doc_notes`. A submission that dies invalid
  still writes a row (`complete_invalid`); this wrote nothing at all.
- **Two later NEW submissions overtook it and finished.** `623d7bce` first ran 15:11 and
  `7abe1a57` first ran 15:19, both after this one was published at ~15:05, and both completed.
  A FIFO queue does not do that.
- **The consumer was live throughout** (`dispatch-queue-depth.sh`: "last orchestration step
  advanced 1s ago"; 18 orchestrations started in the preceding seven minutes).
- **The chassis had rolled at 14:57**, ~8 minutes before the publish. That is outside the
  documented ~300s post-restart drop window, but the window is approximate and a consumer-group
  rebalance is the obvious mechanism. **This is the load-bearing clue, and the one to check
  first next time**: a dispatch that vanishes without trace, near a roll, is the shape.

Watch the BST/UTC gap while doing this arithmetic — `git log` prints `+0100` and the DB prints
UTC, so the commit at "16:03:34 +0100" is 15:03:34 to every query here. Reading them as the
same clock makes a 62-minute-old submission look 2 minutes old.

Resubmitted (see below). The original correlation is kept in the record because a future
`098` run will find neither.

### A roll happened, and it does NOT contain this fix — do not read v1.0.1167 as the gate

The chassis rolled **v1.0.1165 → v1.0.1167 at 14:57 UTC on 2026-07-26**, about six minutes
*before* this fix was committed (`fe00304bd`, 15:03 UTC). So a roll has occurred and the
close-out gate is still **not** met. Confirmed against the running pod
`agent-chassis-5645cb45d6-kpxtq`: the old predicate line
`AND p.build_status IN ('deployed', 'needs_rebuild')` is still present (count 1).

Recording this because "has it rolled yet?" is the obvious way to check the gate, and the
honest answer here is yes-but-not-past-you. v1.0.1167 **does** carry the other session's nav
fix (`a9083d51b`, `NeverDeployedPagePredicate` present), which is why the roll happened.

### Close-out checklist — what has to happen before this moves to `/bugs_closed/`

The Go half is inert until an image roll past commit `fe00304bd` — which v1.0.1167 is not.
Then:

1. Pod-grep — **and not with the marker this checklist first named.**

   > **CORRECTED 2026-07-26.** This step originally said to grep
   > `p.status IN ('active', 'deployed')`. That is a string my change **uses**, not one it
   > **created**: it already occurs **4 times** in `v1.0.1167`, an image built *before* the
   > fix was committed. It would have passed on day one and every day after, proving nothing —
   > the vacuous-verification class logged in `WRONG_CALLS.md` the same day. Caught by running
   > it against the current pod out of habit and getting `4` where `0` was expected.
   >
   > Note also that `strings` splits on newlines, so a multi-line SQL const does **not** appear
   > as one searchable blob — you cannot grep the distinctive *combination* of lines, only an
   > individual line. That is why the discriminating test here has to be the disappearance of
   > the old line.

   The old predicate is a single line unique to this query, so its **absence** is the signal,
   paired with a positive control that fails loudly if the query vanished for any other reason:

   ```sh
   POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
   # must become 0 (it is 1 in v1.0.1167):
   kubectl exec -n ai-persona-system $POD -- sh -c \
     "strings /app/agent-chassis | grep -c \"AND p.build_status IN ('deployed', 'needs_rebuild')\""
   # positive control, must stay >=1 — proves the query is still in the binary at all:
   kubectl exec -n ai-persona-system $POD -- sh -c \
     "strings /app/agent-chassis | grep -c \"AND p.page_type = 'blog-post'\""
   ```
2. Re-run the two queries above; kept set must still be 14/16/6, drop set still those six.
3. Ideally exercise the failing branch rather than the happy path: give a scratch site a
   `blog-index` page plus one archived and one never-deployed `blog-post`, run
   `rebuild_blog_listing`, and assert neither appears. A green run on a site with no
   offending rows proves deployment, not correctness.
4. Then move this file to `/bugs_closed/` keeping its number.

### Notes for adjacent cases

- **`/bugs_open/053` candidate 3 — DONE, by another session, on the same day.** It had
  deferred itself explicitly onto this bug: *"its correct predicate is `deployed_at IS NULL`
  … that predicate fix is `/bugs_open/052`; candidate 3 should ride it, not this."* I left it
  alone deliberately — `who-owns` and a mid-session compile break both showed another thread
  had `nav_tables.go` open — and they landed it as `a9083d51b` a few hours later. Worth
  recording that the hand-off worked: the deferral named the predicate, the predicate got
  built here, and the other lane picked it up without the two of us colliding or duplicating.
  Their change replaced the bool `deployedOnly` with a `NavVisibility` type, moved the filter
  out of SQL into `applyNavVisibility`/`loadFetchablePageSet`, and dropped a `LIMIT` that had
  been capping before filtering. **Do not re-plan candidate 3 from this file.**
- **Candidate 3 of this file** (emit a work item instead of silently dropping) remains
  unbuilt and still gated on `/bugs_open/033` — nothing consumes `needs_human_review`. The
  new Warn above is the cheap stand-in, not a substitute.
- **Ruled out, so nobody re-walks them.** Re-running the survey by the column-sweep method
  (`grep` every Go query against `pages` naming `build_status`/`deployed_at`, then subtract
  the known callers) turns up two more page queries filtered on `build_status = 'deployed'`:
  `render_directory_action.go:344` and `render_news_section_html.go:77`. **Neither is this
  class.** Both are `queue*PageRerenders` helpers answering *"which pages should I queue for
  re-render"*, not *"which pages may I advertise"* — they emit no links, and an over-strict
  filter there costs one skipped re-render, which self-corrects. The predicate is arguably
  still too tight (a `needs_rebuild` page carrying a news listing is skipped), but that is a
  scheduling question, not a 404 question.
- ~~`discovery_checks.neverDeployedPredicate` is an independent, unexported, negative-form
  copy of the same predicate… worth consolidating, not urgent.~~
  > **CORRECTED 2026-07-26, same day, hours later — another session consolidated it while
  > this was being written.** Commit `a9083d51b` ("chrome must not link a page that was
  > never built") promoted that unexported literal to
  > **`datahelpers.NeverDeployedPagePredicate`** (`datahelpers/links.go:210`) and pointed
  > both `check_phantom_internal_links.go` and the new `loadFetchablePageSet` in
  > `nav_tables.go` at it. That commit is also **`/bugs_open/053` fix candidate 3 landing** —
  > and it used the right predicate, not the `build_status = 'deployed'` one that would have
  > delisted 31 working pages.

  So the fleet now has **two exported constants encoding one rule, in opposite polarity**:

  | constant | form | alias |
  |---|---|---|
  | `queryresolve.FetchablePageEligibilitySQL` | `AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')` | requires `p` |
  | `datahelpers.NeverDeployedPagePredicate` | `deployed_at IS NULL AND COALESCE(build_status, '') <> 'deployed'` | unaliased |

  **They are equivalent — genuinely, not accidentally.** My earlier note here guessed they
  matched "only because a NULL `build_status` makes the disjunct falsy either way", which
  implied a latent divergence. Working the NULL case through both forms shows there is none:
  with `build_status` NULL and `deployed_at` NULL the positive form gives `FALSE OR NULL`,
  falsy in a `WHERE`, so the page is excluded, and the negated form gives
  `NOT(TRUE AND TRUE)`, also excluded; with `deployed_at` stamped both include. Same answer
  in every combination.

  Checked against live data too — zero disagreements across all 425 pages:

  ```sql
  SELECT count(*) FROM pages p
  WHERE (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')
     IS DISTINCT FROM
        NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed');
  -- 0
  ```

  **That query is weaker evidence than it looks and should not be the argument**, which is
  why the reasoning above is: `build_status` is never NULL on any live page (only `deployed`
  357, `needs_rebuild` 41, `planned` 27), so the one input that could distinguish the two
  forms does not occur in production. The census confirms they agree on today's population;
  only the case analysis shows they must agree on any population.

  Remaining risk is therefore drift, not a present defect: two constants that must stay
  logical negations, in packages that do not import each other, with nothing asserting the
  relationship. A test that pins `NOT(one) == other` over a table of build_status/deployed_at
  combinations would close it. Not urgent, and no longer two *copies* — that part is fixed.

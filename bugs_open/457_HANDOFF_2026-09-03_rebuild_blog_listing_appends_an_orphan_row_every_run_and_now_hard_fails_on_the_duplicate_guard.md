# 457 — `rebuild_blog_listing` appends an orphan `page_components` row every run, and now hard-fails on the duplicate guard

**Filed 2026-09-03** (components lane). **Live**: it is currently failing a boxingonline chrome
refresh, which stops `create_rerender_items` running, so **zero child rerenders are created for
18 pages** and the retry fails identically.

## The defect, in one INSERT

`rebuild_blog_listing_action.go:403-407`, the "no existing component found" branch:

```sql
INSERT INTO page_components (page_id, slot_name, position, rendered_html,
                             rendered_html_digest, content_data, build_status)
VALUES ($1, $2, 3, $3, md5($3), $4::jsonb, 'deployed')
```

Three faults in six lines:

1. **`position` is hard-coded to `3`** — it does not look at what already occupies position 3.
2. **`component_id` is never set** — every row this creates is an orphan with a NULL component.
3. **It is an unconditional INSERT** — no upsert, no existence check. Whenever
   `findBlogListingSlot` fails to find a listing slot, this runs *again* and appends *another*
   row.

So a page whose slot discovery keeps failing accumulates one orphan row per run, all at position 3,
all NULL-component, alongside whatever legitimately sits there.

## Measured

`[MEASURED 2026-09-03]` boxingonline.com `articles-index`:

| position | component | rows | first | latest |
|---|---|---|---|---|
| 1 | hero | 1 | 08-31 14:30 | 08-31 14:30 |
| 2 | Generic Text Block | 1 | 08-31 14:30 | 08-31 14:30 |
| 3 | call-to-action | 1 | 08-31 14:30 | 08-31 14:30 |
| 3 | **(NULL)** | **6** | **08-31 16:29** | **09-02 16:28** |

Six orphans accumulated over two days beside the legitimate `call-to-action`.

> **CORRECTED 2026-09-03, same day — my first fleet figure was misleading and the
> `site_delivery_and_editor` lane caught the predicate.** I wrote *"12 pages fleet-wide carry more
> than one row at the same position, so the position collision is broader"*. That count groups by
> `(page_id, position)` only, so it sweeps in pages where **different slots legitimately share a
> position** — which `LANDMINES.md` records as normal (generic-text-block used 2–3× on one page).
> It is not a count of this defect.
>
> Re-run grouped by `(page_id, slot_name, position)` — the predicate that isolates *the same slot
> duplicated*:
>
> | page | slot | pos | rows | NULL-component |
> |---|---|---|---|---|
> | **boxingonline.com / articles-index** | generic-text-block | 3 | **6** | **6** |
> | webdesign.co.uk / tool-html-minifier | tool-html-minifier | 2 | 2 | 0 |
> | webdesign.co.uk / tool-svg-optimizer | tool-svg-optimizer | 2 | 2 | 0 |
> | webdesign.co.uk / tool-ab-test-calculator | tool-ab-test-calculator | 2 | 2 | 0 |
>
> **Four pages, and only ONE is an accumulator via this defect.** The three webdesign pages carry
> two rows each with non-NULL components — a different phenomenon, not this one.
>
> ⚠ **AND THE CENSUS HAS A BLIND SPOT BY CONSTRUCTION, which that lane named:** a page whose
> `sections` is `[]` falls through `findBlogListingSlot` to the *default blog-listing* path rather
> than to the strategy-2b branch that produces a position-3 generic-text-block. `ai-agent-orchestration`
> and `leopardess` are in that state, so **no position-or-slot census of this shape can see them.**
> Any fleet number here is a lower bound until that path is enumerated separately.

## Why it only started failing now

Migration **316**'s `uq_page_components_no_byte_identical_duplicate` refuses a byte-identical
duplicate. Once the listing renders to the *same bytes* as an orphan already present, the INSERT
violates the constraint and the whole action errors — so the accumulation was silent for two days
and is now a hard stop. **The constraint is not the bug; it is the thing that finally reported it.**

## Blast radius today

The failure aborts `rebuild_blog_listing`, which runs as a step *before* `create_rerender_items`
in the chrome-refresh chain. So a chrome refresh on boxingonline (`ec92320f`) re-rendered all three
`site_components` at 10:42:35Z and then died, **creating none of its ~18 child page rerenders**.
It retries and fails identically. (Reported by the `site_delivery_and_editor` lane, who hand-filed
the 18 items to work around it.)

## Fix candidates, ordered by what closes the door

1. **Make the INSERT idempotent and position-aware.** It should either upsert on
   `(page_id, slot_name)` or refuse when a row already occupies the slot. Hard-coding `position 3`
   is the root of the collision; the position should come from the discovered slot, not a literal.
2. **Set `component_id`.** An orphan row is invisible to every component-keyed query — including
   the ones this lane has been using all week — and cannot be attributed, re-rendered by component,
   or cleaned up by anything that reasons about components.
3. **Make `findBlogListingSlot`'s failure loud.** The action currently treats "no listing slot
   found" as a reason to *create* one. On a page that has no listing slot because it should not
   have one, that is a silent structural edit, repeated per run.
4. **Clean up the 8 existing orphans** — but only after 1, or they come straight back. Deleting
   them is a per-site data change and belongs to the site owners, not this lane.

## Related

- `bugs_open/451` mentions the same constraint from the two-strike-ladder side.
- The same file's `loadListingTemplate` looks its template up **by name** rather than following the
  `page_components.component_id` it is writing — noted in `bugs_open/425` fix-candidate 5, and the
  two are the same underlying carelessness about component identity in this action.

## CONTRIB from `site_delivery_and_editor`, 2026-09-03 ~10:55Z — WHY the finder misses, and the fleet shape of pages that can hit this

`findBlogListingSlot` (`rebuild_blog_listing_action.go:645–` ) has three strategies and the
boxingonline page falls through all of them to the INSERT branch this file describes:

1. **Strategy 1** — a `page_components` row whose `slot_name` is one of `slotPriority` =
   `blog-listing, article-grid, content-listing, guide-list, featured-article`. articles-index
   has `hero`, `generic-text-block` ×7, `call-to-action` — none match → miss.
2. **Strategy 2a** — a `slotPriority` name inside `pages.sections`. articles-index `sections` =
   `["hero","generic-text-block","call-to-action"]` → miss.
3. **Strategy 2b** — "first content section from the page plan" not in the skip list
   (`hero/header/footer/head/call-to-action/cta`) → returns **`generic-text-block` with
   `uuid.Nil`** → the caller takes the INSERT branch at `:404` with `position` hard-coded 3.

So the accumulation is not random: it is deterministic on any blog-index page whose plan names
no listing-class section, and the row it appends is titled "Latest Articles" in a slot the page
already uses for prose. Migration 316's constraint is what finally reported it (this file says
so; agreed).

**Fleet census `[MEASURED 2026-09-03 10:5xZ]`**, `pages.page_type='blog-index' AND status='active'`
— **4** pages:

| site | url | `sections` | pos-3 `generic-text-block` rows |
|---|---|---|---|
| ai-agent-orchestration.com | /blog.html | `[]` | 0 |
| **boxingonline.com** | /articles/index.html | `[hero, generic-text-block, call-to-action]` | **6** |
| finetuning.uk | /blog.html | `[blog-listing]` | 0 |
| leopardessconsulting.co.uk | /blog.html | `[]` | 0 |

Only boxingonline has BOTH a listing-less plan AND rerender-pages runs (its `stale_chrome` /
chrome-refresh history); the two `[]`-sections pages would take strategy 2b's "no sections →
default `blog-listing`, `uuid.Nil`" path and insert into a `blog-listing` slot instead — so they
accumulate under a different slot name and this census (position-3 `generic-text-block`) is blind
to them by construction. Re-run it grouped by `(page_id, slot_name, position)` before quoting the
fleet number. Disconfirming result for the boxingonline mechanism would have been a listing-class
name in its `sections` or a matching slot row; there is neither.

**Operational consequence on boxingonline today:** every `rerender-pages` run for the site dies at
this step (the operator chrome refresh `ec92320f` at 10:42:34Z, and it will again at 11:12 and on
attempt 3), so chrome NEVER propagates through the workflow on this site; the 10:42Z head/header
(GTM + consent) reached `site_components` because `render_site_components` runs before this step.
Interim taken: the 18 per-page `_assemble` items hand-filed 10:47:56Z (batch `000622a9`), the
shape `create_rerender_items` would have produced. Not touching the six orphan rows (your
candidate 4: futile before the code fix, and a data change on the paid site).

## ⭐ THE ORPHANS ARE SERVED DAMAGE, NOT CLUTTER — and they account for the empty card slots on `/articles/index.html` exactly

`[MEASURED 2026-09-03 15:00Z]` Raised by `site_delivery_and_editor`'s pre-delivery sweep, which
found boxingonline `/articles/index.html` serving **14 empty `article-card__category`** and **2
empty `article-card__excerpt`** elements, and asked whether that was a second listing component,
an unguarded template, or a data gap on `category`. **It is none of those. It is this bug.**

The six orphan rows on page `6bb3b9a6` each carry their own `articles` array **and their own
`rendered_html`**, frozen at whatever the template said when that row was written. The page
assembles all six. Per row:

| instance | written | `category` total / **empty** | `excerpt` total / **empty** | html |
|---|---|---|---|---|
| `cf6f06a9` | 08-31 16:29 | 0 / 0 | 6 / **2** | 4,429 B |
| `dbb4f217` | 08-31 18:14 | 0 / 0 | 6 / 0 | 4,638 B |
| `b14358a8` | 09-01 01:31 | 2 / **2** | 6 / 0 | 5,195 B |
| `99160af4` | 09-01 01:58 | 6 / **6** | 6 / 0 | 6,335 B |
| `b1f9cd4e` | 09-01 02:34 | 6 / **6** | 6 / 0 | 6,334 B |
| `1ef994d5` | 09-02 16:28 | 0 / 0 | 6 / 0 | 6,070 B |
| **sum** | | **14 / 14** | 36 / **2** | |

**14 and 2. Both axes match the served counts exactly**, from two instruments that share no code:
their `curl` of the origin and my `regexp_matches` over `page_components.rendered_html`. That
agreement is the attribution.

**Three consequences that change how this bug should be fixed:**

1. **A re-render cannot repair these rows.** They have `component_id` NULL and no live component to
   resolve, so `resolveComponent` misses and every one of them takes a **carry** branch — their
   stored HTML is re-shipped verbatim, for ever. `bugs_open/454`'s fix, which repaired the deck
   class everywhere else on the estate today, **cannot reach them**. The only remedy is deletion
   (fix candidate 4), which is why candidate 4 should be re-read as a repair rather than a tidy-up.
2. **The empties are a FOSSIL RECORD of migration 682 landing.** Read the table top to bottom: 2
   empty excerpts on 08-31, then `category` appearing empty 2 → 6 → 6 as the pre-682 template
   spread, then 0 on the row written 09-02 16:28 after 682 applied. Each orphan preserves the
   template of its own birthday. So `425`'s class is not "still live on this page" in the sense of
   a template that renders wrong — **the guarded template is fine here too; the page is serving
   archaeology.**
3. **It inflates the page sixfold.** Six copies of the same six-item deck = 36 cards where the
   page should show 6, plus a legitimate `call-to-action` sharing position 3. That is the reader-
   visible defect, and it is larger than the empty slots that led us to it.

**Corrected here, because it was mine:** I told the sweep's owner that `425`'s empty-slot class had
"not reached articles-index". It never will — the class was never on that page. What is on that
page is `457`.

> ⚠ **Do not use this page as a `425` or `682` test target, in either direction.** A guarded
> template will read as unguarded here, and a repair will read as no-op, because the bytes come
> from rows nothing re-renders. This page is a `457` fixture and nothing else.
>
> ⚠ **And do not read the served page as evidence about the live template on ANY page with an
> orphan row.** The check is one query: `SELECT count(*) FROM page_components WHERE page_id=$1
> AND component_id IS NULL` — non-zero means the served bytes are partly unattributable to any
> component, and every class-count you take from them is a mixture of eras. `[MEASURED 2026-09-03]`
> 8 such rows on 3 pages fleet-wide.

> **⚠ CORRECTED 2026-09-03 15:45Z — THIS RULE OVER-FLAGS BY 6×, and the count above was never mine
> to quote.** Checked (not adopted) after the `bugs_open/384` lane challenged it, and their
> correction holds at the code and at the data. `resolveComponent`
> (`rerender_page_sections_action.go:361-393`) does **not** give up on an empty `componentID` — it
> falls through to `schemas[s.slotName]`, and `loadComponentSchemas`
> (`plan_sections_action.go:1981-2002`) indexes **by both `Name` and `Function`**. So a NULL-id row
> resolves whenever its `slot_name` matches either column.
>
> `[MEASURED 2026-09-03 15:45Z]` **14** NULL-`component_id` rows on **7** pages fleet-wide, of which
> **12 RESOLVE** (all by `function`) and only **2 are genuinely stranded** — finetuning.uk `/blog`
> (`article-grid`) and gamesdesign.co.uk `/game-jelly-invaders` (`section`). **Neither is on
> boxingonline.** The six orphan rows on `/articles/index.html` resolve by function, so a re-render
> WOULD refresh their content and clear the empty elements; what it cannot fix is that six of them
> exist, so the page would serve six *fresh* decks instead of six stale ones. **Deletion is still
> the remedy — for the duplication, not for the emptiness.**
>
> The screening predicate that is actually correct:
> ```sql
> pc.component_id IS NULL
>   AND NOT EXISTS (SELECT 1 FROM content_components cc
>                    WHERE (cc.name = pc.slot_name OR cc.function = pc.slot_name) AND cc.is_active)
> ```
> ⚠ **And the trap inside the fix:** slot names match `function`, not `name`. **Zero** of the 14 rows
> match by `name`, so the obvious `WHERE cc.name = pc.slot_name` screen returns **14 of 14
> stranded** — clean, plausible and entirely wrong. The 384 lane caught it only because a
> known-good control (`content-listing`) came back false too. Put a control in the census.
>
> **Two errors of mine, not one.** The rule was over-wide, and the figure "8 such rows on 3 pages"
> was **inherited from `bugs_open/457`'s own earlier census and repeated as though I had measured
> it** — it is 14 on 7 by the plain predicate. `[MEASURED]` next to a number I did not take is the
> worse half of this.
>
> `[INFERRED, untested]` a re-render of that page may still fail at the save: migration 316's
> `uq_page_components_no_byte_identical_duplicate` refuses a row byte-identical to one already
> there, and six rows rendering the same deck from the same data is exactly that shape — which is
> `457`'s own reported failure mode.


**Not fixed by me and not touched:** the six rows stay. The site is paid, the fix is a code fix,
and deleting rows by hand on someone else's live page is the wrong instrument — the `457` fix plus
one rebuild is the right one. Flagged to the owning lane rather than actioned.

## ⚠ TWO PRODUCERS WEAR ONE SYMPTOM — a NULL `component_id` row is usually NOT this bug, and this bug has appended nothing since 09-02 16:28:02

`[MEASURED 2026-09-03 15:31:36Z]` **15** rows with `component_id IS NULL` on **8** pages. Two peer
lanes and I each censused that predicate today and all three of us initially read the growth as
*"457's orphan append is running continuously."* **It is not.** Checked at the code, not inferred
from the shape:

| | **this bug** (`rebuild_blog_listing_action.go:402-407`) | **the ordinary save path** (`save_page_sections_action.go:1124-1127`) |
|---|---|---|
| position | **hard-coded `3`** | `i+1` — the section index |
| `component_id` | **not in the INSERT's column list at all** | written from `componentIDPtr`, NULL when the metadata carries none |
| fires when | a blog-index page's `findBlogListingSlot` misses | any page build where a section's metadata lacks a resolved id |
| arrives | appended alone to an existing page | in the same second as its resolved sibling rows |
| rows today | **6, all on boxingonline `/articles-index`**, newest **09-02 16:28:02** | the rest, including the newest row on the estate |

**This bug's own rows have not grown in 23 hours**, which is exactly what §"Why it only started
failing now" predicts: the action hard-fails on migration 316's duplicate guard *before* reaching
the insert. So a growing NULL-`component_id` count is **not** evidence that this defect is still
accumulating, and a session watching that predicate to decide whether the fix is urgent will read
the wrong signal.

**The row that misled all three of us:** advertise.co.uk `/tool-cpm-cpc-benchmark-comparator`,
created 15:27:34Z, `page_type=tool`, position **5**, slot_name = the page's own tool name — with
**all five rows on that page created in the same second** by one full build (`needs_content_page`
item `74d9bd66`, page-build-handler, claimed 15:21:03Z, complete 15:28:16Z). Four siblings got
component ids; the fifth did not. Two older rows share that shape: idea.uk `/tool-funding-fit`
(09-02 12:27) and loanzy.uk `/tool-loan-vs-savings` (08-28 07:33), both tool pages, both named
after their tool.

⚠ **`[UNVERIFIED]` why those sections' metadata lacked an id is NOT chased here** and is not this
bug. It belongs to whoever owns the tool-page build. Recorded so the next reader does not fold it
into this file the way I nearly did.

**So the census to use for THIS bug is not the plain predicate.** It is the duplicate-position one
that the original §Measured used:

```sql
SELECT page_id, slot_name, position, count(*)
  FROM page_components WHERE component_id IS NULL
 GROUP BY 1,2,3 HAVING count(*) > 1;
```

> **⚠ AND A COUNT ON THIS POPULATION NEEDS THE TIME OF DAY, NOT JUST THE DATE.** Raised by the
> `bugs_open/384` lane and it is a real sharpening of CLAUDE.md's rule. Three censuses inside 45
> minutes: 14 rows / 7 pages at ~14:45Z, 14/7 at 15:15Z, **15/8 at 15:31:36Z**. Each was correct
> when taken and stale within the hour. A reader tomorrow taking "14 rows on 7 pages" as current
> gets neither number.
>
> **The one figure that has not moved, and it is the strongest form of the trap** (384's
> measurement, quoted as theirs): of all 15 rows, **zero** match an active component by `cc.name`.
> Every resolving row resolves by `function`. So a screening query joined on `cc.name` returns
> **100% stranded on this population at any size, on any day** — it cannot come out right, which is
> why the 14-of-14 result looked plausible to two lanes at once.

## THE PRE-FIX BASELINE FOR THE HELD PAGE — recorded here so it outlives the session that took it

Measured at the **served bytes** of boxingonline `/articles/index.html` by
`site_delivery_and_editor` during their pre-delivery sweep, 2026-09-03 (20/20 pages fetched, with
an invented-path 404 control and a `</html>` control per page). **Their measurement, recorded here
rather than left in a session, because whoever implements fix candidate 4 needs a before-state and
the lane that has it will not be running then:**

| | before the fix |
|---|---|
| card links | **36** |
| distinct articles behind them | **6** |
| `Latest Articles` headings | **6** |
| empty `article-card__category` elements | **14** |
| empty `article-card__excerpt` elements | **2** |

**The verification after candidate 4 lands is therefore 36 → 6, six headings → one, and 14/2 → 0.**
The first two are what a reader actually meets; the last is the cosmetic half, and note that a
re-render alone would move **only** the last two and leave 36 and 6 untouched — which is why the
delivery lane declined to re-render the page and why deletion is the remedy for the duplication
rather than for the emptiness.

⚠ **My side of it, from `page_components`, agrees on the two axes it can see** (the per-row sums are
14 and 2, in the section above). It cannot see the 36 or the six headings, because those are
properties of the assembled page rather than of any one row. Two instruments, neither sufficient
alone.

⚠ **The domain does not resolve from every session.** `boxingonline.com` has no A record from this
lane's machine, so `scripts/probe-page-url.sh` fails its own sibling control and correctly refuses
to answer rather than reporting damage. If you cannot fetch it either, the served figures above are
the delivery lane's and the row-level ones are reproducible from the database.

---

## FIX SHIPPED 2026-09-03 (commit `f895616d7`, council `13273c8c`) — and the root cause was one level above the INSERT this file blames

Picked up by the `bugfix_451_457_433_unowned_queue` lane because this bug was unowned for the FIX:
the components lane filed it and wrote *"Not fixed by me and not touched"*, and
`site_delivery_and_editor` declined it as a code fix on a paid site. `scripts/who-owns.py` says
OWNED, but it reads commits — what it sees is two lanes contributing measurements, not working it.

### The root cause is sharper than §"The defect, in one INSERT"

This file blames the INSERT's hard-coded `position 3`, missing `component_id` and lack of an
existence check. All three are real, and none is the cause. **`findBlogListingSlot` has four exit
paths and only strategy 1 ever queried `page_components`.** Strategies 2a, 2b and 3 returned a slot
NAME with `uuid.Nil`, and the caller's `if existingComponentID != uuid.Nil` read that nil as *"this
slot is free"*. It never meant that — it meant *"strategy 1 did not fire"*. That is why the append
was deterministic rather than racy, and it is why fixing the INSERT alone would not have helped.

### ⚠ The trap in this file's own fix candidate 2, which we nearly walked into

Candidate 2 says "set `component_id`". **Doing that ALONE would have made this bug worse and
silent.** `uq_page_components_no_byte_identical_duplicate` is `NULLS NOT DISTINCT`, so the six
orphans collide with each other *because* their `component_id` is NULL. A row carrying a real id
stops colliding — the loud duplicate-key failure becomes a seventh silent append, and the only
instrument currently reporting this defect is removed. **The occupancy check and the id binding must
land in one commit, occupancy first in the code path.** They do, and
`TestBlogListingNeverInsertsWhenTheSlotIsOccupied` is mutation-proven against exactly that.

### ⚠ And a constraint neither this file nor its CONTRIBs states, which shaped the whole design

**A refusal must NOT return an error.** This action is an unconditional step of `rerender-pages`
between `render_site_components` and `get_pages`, and that workflow declares no `error_step` — so an
error return aborts the run before `create_rerender_items` and creates none of the ~18 page
rerenders. That IS the outage in §"Blast radius today". A refusal that errors reproduces it under a
new name. Every refusal now logs at Error and returns `rebuilt:false`.

### What shipped

Every exit path runs an occupancy lookup, and the write decision is a pure function of
`(origin, occupants)` taken in Go before the write: one occupant → UPDATE (whichever strategy named
the slot); plan names the slot and it is empty → INSERT; **guessed (2b) or defaulted (3) → REFUSE**;
several occupants or an unreadable count → REFUSE. A guessed name is by construction not a listing
slot, or strategy 1/2a would have matched — and on this page it named `generic-text-block`, a slot
the page uses for prose. `position` now comes from the plan's section index; `component_id` is bound
from the component that rendered the bytes, NULL on the built-in-default branch where no row exists.

`[MEASURED 2026-09-03 16:2xZ]` occupancy of all four live blog-index pages: aiao `/blog.html`
`blog-listing` pos 2 (1 row), leopardess `/blog.html` `blog-listing` pos 1 (1 row), finetuning.uk
`/blog.html` `article-grid` pos 5 (1 row, strategy 1), boxingonline `generic-text-block` pos 2 (1)
**and pos 3 (6)**. So the INSERT branch is unreachable on every blog-index page on the estate today
— which is why refusing on 2b/3 costs nothing now. Disconfirming result would have been a page with
zero occupants at its resolved slot; there is none.

### A second, LATENT defect found in the same function

`loadContentListingTemplate`'s `function='content-listing' … ORDER BY created_at DESC LIMIT 1` is not
unique. `[MEASURED 2026-09-03 16:1xZ]` **two** active rows: `content-listing` (2025-11-28) and
**`content-listing-guides-boxingonline-com` (2026-09-02)** — so one site's fork was every site's
listing template. Both `md5(html_template) = 1b957ae3…`, byte-identical, so this is latent, not live
damage. Now prefers the canonical row (`name = function`) and **logs** the ambiguity rather than
refusing; refusing would stop every site's listing rebuilding while two forks exist, and RFC_034
makes N forks per function the intended future.

### The orphan rows are NOT touched — handed to `site_delivery_and_editor`

Owner's call, 2026-09-03: the lane that holds the served baseline and owns every dispatch at that
site does the deletion. The exact target, so nobody has to discriminate by count:

| id | position | created | md5(rendered_html) | bytes | component_id |
|---|---|---|---|---|---|
| `3de9aa68` | **2** | 08-31 14:30:43 | *(NULL html)* | — | **set — KEEP** |
| `cf6f06a9` | 3 | 08-31 16:29:47 | `a49768f2` | 4,429 | NULL — remove |
| `dbb4f217` | 3 | 08-31 18:14:32 | `2f84ab40` | 4,638 | NULL — remove |
| `b14358a8` | 3 | 09-01 01:31:57 | `13ea5fd0` | 5,195 | NULL — remove |
| `99160af4` | 3 | 09-01 01:58:36 | `fc99a8b5` | 6,335 | NULL — remove |
| `b1f9cd4e` | 3 | 09-01 02:34:51 | `9352d6a2` | 6,334 | NULL — remove |
| `1ef994d5` | 3 | 09-02 16:28:02 | `5b91bfe5` | 6,070 | NULL — remove |

The legitimate row is unambiguous on three independent axes: position 2, a real `component_id`, and
a NULL `rendered_html` (it renders from `content_data` at assemble time, which is why it is not one
of the six frozen copies). **All six orphan digests are distinct** — which is why 316 did not catch
the accumulation as it happened and only fired when a later render finally matched one.
Prefer `build_status='removed'` (reversible, assembly-excluded, and it drops the rows out of 316's
partial index too) over a DELETE, and check `site_plans` → `site_plan_sections` → `pages.sections`
first per LANDMINES. ⚠ **Deletion is not what fixes the empty slots** — the six rows resolve by
`function`, so a re-render alone moves 14/2 → 0 and leaves 36 cards and 6 headings untouched. If the
empties clear and the card count stays at 36, a re-render happened and the deletion did not.

### Still open after this commit

- The discovery check that should FILE the gap when a blog-index page has no listing slot. Between
  now and then the refusal is legible only in the action's Error log. It needs its own `item_type`,
  a registered handler and a register entry, so it is a separate commit — deliberately not smuggled
  in here.
- **Concurrency:** the occupancy check is check-then-insert. Two simultaneous `rerender-pages` runs
  could both see zero occupants; the second now degrades to a result rather than aborting the chain,
  but a duplicate is still possible if their bytes differ. An advisory lock or a `23505` catch would
  close it.

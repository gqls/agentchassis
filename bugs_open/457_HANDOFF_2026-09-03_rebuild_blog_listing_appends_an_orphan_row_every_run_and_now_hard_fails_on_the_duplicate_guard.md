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

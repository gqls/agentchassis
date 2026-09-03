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

Six orphans accumulated over two days beside the legitimate `call-to-action`. Fleet-wide, NULL-component
rows at position 3: **boxingonline/articles-index 6, gamesdesign.co.uk/game-jelly-invaders 1,
idea.uk/tool-funding-fit 1** — 8 rows, 3 pages. **12 pages fleet-wide carry more than one row at
the same position**, so the position collision is broader than the NULL-component symptom.

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

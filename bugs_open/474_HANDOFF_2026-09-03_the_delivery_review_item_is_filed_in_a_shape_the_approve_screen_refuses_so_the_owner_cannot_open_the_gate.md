# 474 — the delivery review item is filed in a shape the approve screen refuses, so the owner cannot open the delivery gate

Filed 2026-09-03 by `site_delivery_and_editor`, on the **first time `delivery-review-filer` has ever
run** (owner-authorised rehearsal on `idea.uk`, a non-customer site). Found before the owner clicked,
not after.

## What the owner would have seen

He opens the `idea.uk` item in the admin queue. It renders as a checkpoint. The **"Approve &
Continue" button is there.** He presses it and gets **"No review data to approve"**. Nothing happens,
nothing is recorded, and no part of the message tells him what is missing or who should fix it.

The delivery gate stays shut, and every subsequent step of the chain waits on a click that cannot
succeed.

## The mismatch, exactly

`delivery-review-filer` files via `create_work_item`, whose `spec_paths` / `spec_literal` produce:

```json
{"brief": "…", "domain": "idea.uk", "site_url": "https://idea.uk", "checkpoint": true}
```

The admin screen derives everything it will submit from **`spec.review_data`**:

- `App.tsx:519-520` — `editedReviewData` is set only when `item.spec.checkpoint && item.spec.review_data`.
- `App.tsx:762-766` — `handleApprove` opens with `if (!editedReviewData) { setMessage("No review data to approve"); return; }` and posts nothing.
- `App.tsx:1295` — but the button's own visibility condition is only `isCheckpoint && status === 'needs_human_review'`. **So the button shows for an item it cannot submit.** That is what turns a shape mismatch into a dead end rather than an absence.

`HandleApproveWorkItem` would refuse it too, one layer further back: `ReviewData` is bound
`binding:"required"`.

## The two producers disagree, and only one of them is used here

Every other checkpoint item in the estate is filed by `checkpoint_for_review`, which **always** writes
a `review_data` key (`checkpoint_for_review_action.go:157`). `delivery-review-filer` is the one
producer that files a checkpoint through `create_work_item` instead — and `create_work_item` has no
such key.

> **[MEASURED 2026-09-03, all history]** every `site_work_items` row with `spec->>'checkpoint'='true'`,
> grouped by whether it carries `review_data`:
>
> | item_type | has review_data | status | n |
> |---|---|---|---|
> | `copy_edit_proposed` | yes | cancelled / complete / needs_human_review | 3 / 2 / 16 |
> | `needs_human_review` | yes | cancelled / needs_human_review | 5 / 2 |
> | **`needs_delivery_review`** | **NO** | needs_human_review | **1** |
>
> One row in the estate lacks it, and it is the only delivery review ever filed. The measurement
> could have come out otherwise — a second producer filing without `review_data` would show here.

## The line in our own code that predicted this and dismissed it

`platform/delivery/prepare.go`, `ReviewItemRequiredSpec`:

> *"(The approve endpoint additionally requires a non-empty `review_data` JSON body from the CALLER
> of the API — the admin screen's concern, not the producer's, so it is noted here and not encoded.)"*

The observation is right and **the conclusion is wrong**. The screen has no other source for
`review_data` than the item's own spec, so "the caller's concern" and "the producer's concern" are
the same concern. That parenthesis is why the requirement was written down and still not met: it was
recorded as someone else's problem.

The same file's main comment gets the neighbouring case exactly right — it warns that an item missing
`checkpoint: true` produces an error whose advice steers the owner to press RESOLVE, which writes
`resolved_by`, "the key `Reviewed()` deliberately ignores". This is the same failure one notch along,
and it is worse, because there is no error text at all — just a button that does nothing.

## Why this is `bugs_open/466`'s family

466: an approval handed the editor a payload it could not read. 474: a producer hands the approval
*screen* a payload it cannot read. Both are a producer and a consumer disagreeing about where the
content lives; both were invisible until a human pressed a button for the first time; both leave the
person with no way to learn their decision did not land.

## Fix candidates, ordered by what closes the door

1. **Have `delivery-review-filer` file `review_data` in the spec** — `spec_paths` already carries
   `brief`, `domain` and `site_url`, which is precisely what the owner is being asked to review, so
   the content is honest rather than a placeholder to satisfy a check. Pure DB config: live
   immediately, no roll, no image. **Taken — migration `751`.**
2. **Make the button's condition match the button's action** (`App.tsx`). Today visibility is
   `isCheckpoint`, and submittability is `editedReviewData != null`. Those must be one predicate, or
   any future producer reintroduces this. Needs a frontend build; not taken here.
3. **Encode the requirement in `ReviewItemRequiredSpec()`** rather than noting it in a parenthesis —
   the function exists to be the filing contract, and a producer calling it should get a spec that
   can actually be approved. Requires deciding what `review_data` a generic producer should supply.
4. Weakest: document it. Rejected for the same reason 466 rejected it — the parenthesis above IS the
   documentation, and it did not work.

## Interim taken on the affected item

The live `idea.uk` row `e370e0bb` had `review_data` added by hand (migration `751` also patches it),
built from the fields already in its own spec — nothing invented, nothing the filer would not now
produce itself. The owner can approve it.

## How to verify the fix

Dispatch `delivery-review-filer` at any site and confirm the filed item's spec carries `review_data`
with `domain`, `site_url` and `brief`. Then open it in the admin queue: the panel must render editable
fields, and **"Approve & Continue" must produce an approval, not "No review data to approve"**. The
negative control is this file's own evidence: item `e370e0bb` as originally filed.

## Related

- `bugs_open/466` — the same producer/consumer shape one step later in the same chain.
- `platform/delivery/prepare.go` — `ReviewItemFiledStatus`, `ReviewItemRequiredSpec` and the
  parenthesis quoted above.
- `docs/agent_docs/sql_for_agents/651_delivery_review_and_email_agents_HOLD.sql` — the seed, whose
  dispatch recipe was separately found incomplete the same hour.

---

## SECOND DEFECT, found the same hour: the item is also in a pipeline the dashboard does not offer

The owner, told where to click, still could not find it. **He was right and the directions were
wrong** — with the dashboard's default settings the row is never sent to his browser at all.

- `App.tsx:417` — `const [pipelineFilter, setPipelineFilter] = useState("build")`. The Work Items
  view **defaults to the build pipeline**.
- `App.tsx:470` — the filter is applied **server-side**: `/work-items?pipeline=build&…`. So this is
  not a row he could scroll past; it is a row the API is never asked for.
- `App.tsx:1091-1093` — the dropdown's entire option set is **`build` · `content` · `all`**.
- The item is `pipeline = 'delivery'` (`delivery-review-filer`'s step config sets
  `item_pipeline: 'delivery'`, and it is right to).

So the only route to it is the unlabelled catch-all, **"all pipelines"**, which a person has no
reason to select and no hint that they must.

> **[MEASURED 2026-09-03, all history]** `site_work_items` by pipeline:
> `build` 24,887 · `content` 3,204 · **`design` 1,933** · `diagnose` 42 · `experience` 3 ·
> `reports` 3 · **`delivery` 2** · `maintenance` 1.
>
> **Five of the eight pipelines have no dropdown option**, and they hold **1,984** items between
> them. `design` alone has 1,933 rows reachable only by choosing "all". The measurement could have
> come out otherwise — if `delivery` were the sole orphan this would be a one-off rather than a
> class.

**Why this is the same bug as the one above and not a separate inconvenience.** Both are the delivery
pipeline's only item type meeting a consumer that does not know it exists: the approve handler wanted
a `review_data` key the producer never wrote; the list view offers a pipeline vocabulary the producer
is not in. In both cases the machinery reports success — the item is filed, the page renders — and the
person is simply never shown the thing they are being asked to act on. **A filter defaulting to a
subset is an absence-of-evidence generator**, which is the same shape as `bugs_open/033` (a capped
read path whose truncation read as "no items exist") and the reason `HandleListWorkItems` counts
without the status predicate.

### Fix candidates

1. **Derive the pipeline options from the data rather than hardcoding three** — the list endpoint
   already computes `statusCounts` and `typeCounts` server-side; a `pipelineCounts` beside them makes
   the dropdown self-maintaining, and a new pipeline can never again be invisible by omission.
   Strongest, and it closes the door for `design` too.
2. **Default the filter to `all`** rather than `build`. One character, removes the trap for every
   pipeline at once, and costs a larger default result set on a view that already pages server-side.
3. Add a `delivery` option to the hardcoded list. Fixes today's case and leaves the other four
   orphaned pipelines exactly as they are — the shape that produced this.
4. Weakest: document the workaround. Rejected: the person who needs it is the one who does not know
   the pipeline exists.

**None taken yet** — all three real candidates are frontend changes needing a build, and the owner is
unblocked today by selecting "all pipelines". Candidate 1 or 2 is the one worth doing.

### How to verify

File a `needs_delivery_review` item, open Work Items with **default** settings, and confirm it is
visible without touching the pipeline dropdown. Today's negative control: item `e370e0bb`, invisible
under the default and visible the instant the filter is switched to "all pipelines".

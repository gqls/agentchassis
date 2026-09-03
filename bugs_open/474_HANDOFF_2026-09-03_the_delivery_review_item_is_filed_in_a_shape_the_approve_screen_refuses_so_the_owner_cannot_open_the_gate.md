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

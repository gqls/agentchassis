# 466 — an approved review hands the section-editor a payload it cannot read, so the owner's approval fails at the first step

Filed 2026-09-03 by `site_delivery_and_editor`, on the FIRST real use of the approve-to-apply path:
the owner approved a copy edit in the admin queue at 16:21:13Z and the resulting work item failed
17 seconds later, before touching the page.

## What the owner sees

He clicks APPROVE. The review row goes `complete`. A `section_edit` appears. It fails, retries
twice on the ladder, and the page never changes. Nothing tells him. The review row still reads
`complete`, which is true and misleading: it records that HE approved, not that anything applied.

## The failure, verbatim

```
step load_edit_context failed: failed to execute action load_edit_context:
need either page_component_id or both page_name + slot_name
```
`section-editor` orchestration FAILED at `load_edit_context`, 16:22:21 → 16:22:38Z.

## Root cause — two independent defects, either alone is fatal

**1. The approval writes the payload under a key nothing reads.** The review row's `on_approve`
names `include_fields: ["copy_edit", "page_target"]`. The item it produced carries:

```json
{ "copy_edit": null, "page_target": null, "approved_data": { "edits": [ … ] }, … }
```

Both named fields are **null**, and the real content sits under `approved_data`. So the contract
between the checkpoint and the handler is broken at the join: the approval flow and
`include_fields` disagree about where approved content lives.

**2. The shapes do not match even if the fields were plumbed.** `copy-editor` emits **N edits**,
each with its own `page_component_id`, `slot_name` and `field_updates`. `section-editor` applies
**ONE**: `load_edit_context` wants `page_component_id` OR `page_name`+`slot_name` at the top of its
input (`input_fields: [site_id, page_component_id, page_name, slot_name, domain]`), and
`apply_section_edit` takes one `field_updates` (`input_fields: [edit_type, field_updates,
replacement_content_data, new_component_function, page_component_id, transform_name]`). A
two-edit proposal has no single target, so no amount of key-renaming fixes it: **the approval must
fan out to one `section_edit` per edit.**

## Why nobody hit it before

`copy-editor`'s `request_review` → approve → `section_edit` path had **never been exercised**.
Fleet-wide before today: 28 `needs_copy_edit` complete and 58 deferred, but the deferred ones are
RFC_056 verdict rows that are deliberately never dispatched, and no `copy_edit_proposed` review had
been approved. This was the first, on the first paid site.

## Blast radius

Every future copy approval, on every site. The path is the estate's only human-in-the-loop route
for copy changes, and the owner's approval is exactly the signal it exists to carry.

## Fix candidates, ordered by what closes the door

1. **Fan out on approval: one `section_edit` per edit in `approved_data.edits[]`**, each carrying
   `edit_type`, `page_component_id` and that edit's `field_updates` at the top level. This closes
   both defects at once and makes the multi-edit case representable rather than lossy.
2. **Make `include_fields` and the approval writer agree**, with a test that approves a real
   review row and asserts the named fields are non-null. The present pair cannot both be right,
   and nothing failed until a human clicked.
3. **Surface the failure to the approver.** A review row that reads `complete` while its spawned
   work failed is the estate's §8 shape ("a complete work item is not a repaired artefact") aimed
   at a person: he has no way to learn his decision did not land. At minimum, the failing
   `section_edit` should file back against the review row.
4. Weakest: document the shape. Rejected as a fix — it puts the knowledge where the person who
   clicks approve will never look.

## Interim taken on the affected site

Two `section_edit` items hand-filed 16:24:21Z (`fa49b1bc` the listing subtitle, `94835c97` the CTA
line), carrying the OWNER-APPROVED payload byte-for-byte in the shape the handler reads. Nothing
was re-authored: the values are the ones he approved, one of them his own verbatim words. The
failing `5edadfbe` was left alone deliberately, as this file's live evidence.

## How to verify the fix

Approve a two-edit `copy_edit_proposed` review on a test site. Expect **two** `section_edit` items,
each with a non-null `page_component_id` and its own `field_updates`, both reaching `complete`, and
both fields changed at `page_components`. Today's failure is the negative control: one item, both
named fields null, dead at `load_edit_context`.

## Related

- `bugs_open/425` — a complete work item that changed nothing; same family, machine-facing.
- `RFC_056` — the circuit breaker that parks verdict rows, which is why the deferred
  `needs_copy_edit` population is not evidence this path works.
- `bugs_open/457`, `bugs_open/451` — the day's other "built but never exercised" finds.

---

## FIX COMMITTED 2026-09-03 (`33dfeed3a`) — and defect 1 is worse than filed above

Committed, **not yet live**: the Go half needs an image and a roll. Council submission
`d04c1bc1-b9a3-41bb-b144-1d101e68e542`, verdict pending, `Council-Submitted:` trailer on the commit.

### Correction to §"Root cause", defect 1

This file said *"the approval flow and `include_fields` disagree about where approved content
lives"*. That understates it, and the correction came from a census rather than from reading harder.

`checkpoint_for_review` — **the only producer of these items** — builds the review item's spec from a
**fixed key set** (`checkpoint_for_review_action.go:157-180`): `review_data`, `checkpoint`,
`source_agent`, `correlation_id`, then optionally `domain`, `spec_aspect`, `on_approve`. There is no
path by which an arbitrary field named in `include_fields` can be at spec top level. The approve
handler looked its names up in exactly that object.

> **[MEASURED 2026-09-03, all history]** 21 checkpoint items naming **42** `include_fields` entries
> between them since 2026-08-24. **Zero** of the 42 present at spec top level; zero non-null.
> ```sql
> WITH ck AS (SELECT id, spec, jsonb_array_elements_text(spec->'on_approve'->'include_fields') AS fld
>             FROM site_work_items WHERE spec->'on_approve' ? 'include_fields')
> SELECT count(*), count(*) FILTER (WHERE spec ? fld) FROM ck;   -- 42 | 0
> ```
> The measurement could have come out otherwise: a producer writing the named field at spec top
> level would have shown as non-zero.

So it is not a disagreement between two halves. **`include_fields` never copied anything, for any
consumer, from the day the first such item was written** — including the two names in this very
file's own header comment on `checkpoint_for_review` (`reviewed_brief`, `site_record`), which would
have resolved to null the same way. This is the "a dead config key looks like a live one" shape: the
mechanism is named in the config, documented in a header comment, and inoperative.

### A third finding this file did not have: the addresses rot

**[MEASURED 2026-09-03]** of the **31** edits parked in `needs_human_review` (16 items), **3** point
at a `page_components` row that **no longer exists** — a rerender replaces the row with a new id
(`LANDMINES.md`, `copy_quality_two_stage` 2026-08-18). 0 are stale by the `updated_at` test.

That matters *because of* the fix: making the fan-out work unblocks 16 parked proposals, and without
a guard three of their edits would file `section_edit` items that die at `load_edit_context` with
nothing said to the approver — manufacturing more of this bug while fixing it. The fan-out therefore
**refuses and reports** a dead address rather than filing it.

⚠ The staleness test only discriminates for **unapplied** items. In the `complete` bucket it read
4 of 4 "stale", which is artefactual — applying an edit is what bumps `updated_at`. Do not quote
that arm.

### What shipped

| candidate | what was done |
|---|---|
| 1 — fan out | `on_approve.fan_out_from` names an array in the approved data; one follow-on per element, that element's fields merged at the **top** of the child spec. `on_approve.defaults` supplies what the proposal does not carry. Both keys default **absent**, so an `on_approve` naming neither behaves byte-identically to before |
| 2 — include_fields agrees with the writer | resolves `spec` **first** (so nothing that works today can change), then the approved body. A name absent from both is now **absent** rather than an explicit `null` — `copy_edit: null` reads as real, empty content |
| 3 — surface it to the approver | the review row's `result` now carries `follow_on_items[]`, `skipped_edits[]` and `fan_out_note`; `follow_on_item` is kept and names the first. The response always carries `follow_on_item_ids`, so "nothing was filed" is a visible `[]` and not an absent key |
| — | migration **750** wires copy-editor: `fan_out_from: "edits"`, `defaults: {edit_type: "content_edit"}`, `include_fields: ["domain"]`. **[MEASURED 2026-09-03]** of the 41 proposed edits in those items, **41** carry `page_component_id`, **0** carry `edit_type`, **0** carry `page_name` — so `edit_type` must be defaulted and `page_name` must **not** be (a defaulted page_name would be a guess applied to every page; `page_component_id` alone satisfies `load_edit_context`) |

Five tests in `internal/core-manager/admin/approve_fan_out_test.go`. All four load-bearing arms were
**mutation-tested**, not merely observed green: removing the dead-address guard, removing the
approved-body fallback, making fan-out the default, and dropping the element-field merge each kill
exactly the test that names them.

### Prior art this file missed, and it is the strongest evidence for candidate 1

`copy_quality_two_stage` hit the same wall **the day before**. Review `be23d897` (2026-08-31, site
`99cae989`) was approved in chat on 2026-09-02 and released **by hand as two `section_edit` items** —
its own `result` says *"replicating the dashboard `on_approve` contract in the proven `section_edit`
spec shape"*. Two lanes independently hand-built the fan-out before anyone proposed it as a fix. That
is what the handler now does.

Their NOTES (`copy_quality_two_stage/NOTES_two_stage_copy.md:1486`) attribute the blockage to
`bugs_open/033` — "the apply path has never run … because it depends on the dashboard endpoint". True
as far as it goes, and it is why nobody found the defect: the path was never exercised, so the
mechanism's deadness was invisible rather than latent.

### How to verify, revised

The recipe above still stands, plus two arms it did not have:

1. After the roll, approve a **two-edit** proposal: expect **two** `section_edit` items, each with
   its own non-null `page_component_id` and `field_updates`, and the review row's `result` carrying
   `follow_on_items` with both ids.
2. **The opt-in arm, which is the one a later tidy-up will break:** approve a checkpoint whose
   `on_approve` names *neither* new key (any `needs_brief_review` item) and confirm exactly **one**
   follow-on, carrying `approved_data`, as before.
3. **The negative control is on this site and was deliberately preserved:** `5edadfbe`, one item,
   both named fields null, dead at `load_edit_context`.

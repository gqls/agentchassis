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

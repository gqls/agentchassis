# PLAN — bugs_open/367: the required-fields router disposes of findings it cannot resolve

**Started 2026-08-23.** Lane: `bugfix_367_router_remit`. Bug filed by the (now closed)
`bugfix_342_absent_required` lane, which named 367 as its successor and stated
*"Nothing here needs picking up."*

## The problem, in one paragraph

`required-fields-missing-handler` is a **router**: it repairs nothing, it runs one SQL
classification per `required_fields_missing` item and then converts, parks or closes it.
To classify, it must first find the component the finding names, and it looked with
`WHERE pc.build_status = 'deployed'`. When nothing came back the routing `CASE` fell to
`stale`, and `stale` **closes** the item — `complete`, no error, one attempt. That was
coherent while the type had one producer (seed 410 deliberately mirrored the post-deploy
check's own filter). `bugs_closed/342` added a second producer at render time whose entire
purpose is the population that check cannot see, so its true findings were closed as gone.

## The decision, and how it changed

**I started from the bug file's own candidate 1 — "widen the filter" — and narrowed it
after two measurements.** Recording that here because the wrong turn is the useful part.

1. **Widening alone does not repair anything.** The `partial` arm converts via
   `file_rewrite`, which reads four spec fields the render-time producer never writes.
   Unresolved `spec_paths` and `item_key_suffix_field` are both deliberate hard errors, so
   the item would route correctly and die one step later. Widening trades a silent wrong
   close for a loud failure.
2. **The repair arm is the wrong place for this population anyway.**
   `content_rewrite`/`edit_live` runs at `page-build-handler`, whose save step DELETEs
   every agent-writable row on the page. 28 of 31 `from_rfm` conversions were already
   failing there on the owned-page refusal (`bugs_open/333`), and the 367 page is itself
   `rebuild_policy=owned`.

**So the rule installed is narrower and stronger than "widen":**

> A disposer may close only on **positive evidence of absence**. A failed lookup is not
> evidence, and neither is a non-deployed target.

The estate already rules this way one door over —
`revalidate_review_queue_action.go:684`, on the same class of miss: *"That MIGHT mean the
finding is moot, but it might equally be a lookup miss — so it is not positive evidence and
the item stays queued."*

## What shipped — migration 574, config only, live on apply

1. `comp` CTE resolves on the **lifecycle** axis (`COALESCE(build_status,'pending') <>
   'removed'`), not the build axis. Spelled to match `pageComponentNotRemovedSQL`
   (`section_editor_actions.go:1537`). The `COALESCE` is load-bearing — NULL passes the
   `049` CHECK and a bare `<>` is NULL-unsafe.
2. A `tomb` CTE: is there a `'removed'` row at (page, slot)? That is the positive evidence
   of retirement, as opposed to merely not being found.
3. `stale` now requires positive evidence: page gone, **or** locked (277's accept-as-is),
   **or** nothing resolves **and** a removed row is sitting there.
4. New route `target_not_dispatchable` → a fifth **park** at `needs_human_review`, holding
   its dedup key. Not a fifth close.
5. A `target_state` output column so the cause is legible in `orchestration_states` — which
   is the only durable place, because `mark_complete` overwrites `site_work_items.result`.
6. Both close arms' evidence strings corrected; they were false as written.

## Deliberately not done, each with its reason

- **No Go, no chassis build.** With the park route, the render-time population never reaches
  `file_rewrite`, so the spec-contract mismatch never fires.
- **No re-keying of `file_rewrite` to `triage.component_id`** — re-opens a council-settled
  key design (`CQ-023`) and buys nothing once the population parks.
- **No verifier registration** — `CQ-023` already warns one would fail-close this router's
  own `converted` arm.
- **No component-axis predicate family** mirroring `datahelpers/links.go`. The 19 hand-typed
  `pc.build_status='deployed'` reads are all **producers**, whose failure mode is
  under-detection. Only a disposer turns non-detection into an affirmative claim of absence.
  That asymmetry is the transferable lesson and it went to `016b` §9.
- **`bugs_open/333` is OWNED** by the 277 lane — contributed the census, did not compete.

## Open residuals

- Seed 410 is an `ON CONFLICT DO UPDATE` whole-config seed whose verify block never asserts
  the resolution predicate, so a hand re-run silently reverts 574. Mitigated by a header
  pointer only. The real fix belongs to whoever next touches 410.
- The render-time producer still does not write the convert arm's read-set. Harmless today
  because that population parks, but it is a live contract gap.
- `update_work_item_status`'s `complete` arm never consults the verifier framework — filed
  separately; architecture-scope, deliberately not fixed inside a bug patch.

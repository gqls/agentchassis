# 095 — a wrong `slot_name` renders nothing and the run reports COMPLETED

**Filed** 2026-07-26 from the oufe.com workstream.
**Severity** medium-high — silent. The failure is shaped exactly like a success,
and the only symptom is a page that never appears.
**Status** OPEN.

## Symptom

A page with one correctly-populated `page_components` row rendered nothing. The
orchestration reported:

```
COMPLETED | complete_skipped
```

`page_components.rendered_html` stayed NULL, `pages.build_status` never advanced,
nothing was deployed, and **no error was recorded anywhere** — not on the
orchestration, not on the work item, not in the page row.

## Cause

`page_components.slot_name` must equal the component's **function name**, which is
also the entry in `pages.sections`. On every working page in the fleet:

```
pages.sections            = ["hero-about", "about-content", "generic-text-block"]
page_components.slot_name =  hero-about | about-content | generic-text-block
```

The row was inserted with `slot_name = 'main'` — a plausible-looking value, and
the sort of thing anyone hand-authoring a page would reach for. It matches no
entry in `sections`, so the renderer pairs it with nothing.

`assemblePage` then returns empty, and `rerender_single_page` converts that into:

```go
// rerender_single_page_action.go:105-118
return map[string]interface{}{"success": false, "skipped": true,
    "reason": "no components found for page", ...}
```

which `page-rerender`'s `check_skipped` conditional routes to `complete_skipped` —
a terminal step whose name contains "complete" and whose status is COMPLETED.

## Why this matters more than it looks

The adjacent trap is already documented — **a NULL `slot_name` renders nothing
while the job still reports COMPLETED** (idea.uk delivery trap). This is the same
failure with a *wrong* value rather than an absent one, and it is worse in one
respect: a NULL is visibly missing data, whereas `'main'` looks deliberate and
survives review.

The information needed to detect it exists in the same transaction: the renderer
knows the section names it wanted and the slot names it found, and can see that
the intersection is empty.

## Related, found in the same sitting

A page still at `build_status='planned'` is also skipped, with the same
success-shaped outcome. Setting `needs_rebuild` is required before a
never-built page will render.

## Fix candidates, ordered by what closes the door

1. **Make the mismatch unrepresentable**: a CHECK or trigger requiring
   `page_components.slot_name` to match the row's component `function`. The
   value is derivable, so accepting a free-text slot that must equal a known
   value is a schema defect in a documentation costume.
2. **Fail loudly instead of skipping**: when `assemblePage` finds zero components
   *but the page has sections planned*, that is a defect, not a no-op. Return an
   error naming the sections wanted and the slot names present. A page with no
   components at all is a legitimate skip; a page with components that match
   nothing is not.
3. Emit a work item on `skipped` so it lands in the review queue rather than
   vanishing. Weakest of the three — it makes the failure visible without
   preventing it.

## How to verify a fix

Insert a `page_components` row with a slot name that matches no section, then
re-render. Candidate 1: the insert is rejected. Candidate 2: the run fails with a
message naming both lists. In no case should a run report COMPLETED having
rendered nothing while sections were planned.

**Induce the fault to verify** — a green re-render on a correctly-built page
proves only that the happy path works, which was never in doubt.

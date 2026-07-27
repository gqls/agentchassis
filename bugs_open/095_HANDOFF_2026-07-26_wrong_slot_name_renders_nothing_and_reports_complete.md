# 095 — a wrong `slot_name` renders nothing and the run reports COMPLETED (FIXED IN CODE, INERT UNTIL THE NEXT CHASSIS ROLL)

> ## STATUS 2026-07-27 — candidate 2 applied, council APPROVED, **not yet live**
>
> **Fix:** `6579e9ae1`. `getPageSections` now counts what it saw instead of discarding
> it (the SQL pre-filter on `rendered_html` was throwing away exactly the evidence that
> distinguishes the two empty cases), and returns a `pageAssembly` alongside the HTML.
> The re-renderer splits the one ambiguous outcome in two: **component rows that exist
> and contribute nothing fails the step**, naming the planned sections, the unrendered
> slots and the blank slots; **no component rows at all** stays a legitimate skip but
> now names what the page planned instead of "no components found for page".
> `ApplySectionEditAction` uses the same discriminator — it had the same silent shape,
> returning `success=true` with `html=""`.
>
> **Candidate 1 was deliberately NOT taken.** This file's own correction shows slot
> mismatch is not the mechanism, and a CHECK constraint would have to reckon with the
> 70 benign rows across 12 sites that correction identifies.
>
> **Council:** APPROVED, correlation `d7d47150-883a-4991-932a-372d9fe2b4b6`
> (10 reviewers, 0 unreadable, 4 advisory objections, no veto). One was acted on: the
> section-editor arm originally failed on *any* empty reassembly, which two seats
> correctly called asserted rather than evidenced; it now shares the re-renderer's rule.
>
> ### CORRECTION to §"Scale — measured 2026-07-27: zero live instances"
>
> **That is no longer true, and it stopped being true within twenty minutes.** The
> defect shape was 0 fleet-wide at ~18:05 UTC and **1** at ~18:35:
>
> ```
>  domain   |          name           | status | build_status  | comp_rows | usable | planned
> ----------+-------------------------+--------+---------------+-----------+--------+---------
>  oufe.com | tool-recovery-waterfall | active | needs_rebuild |         1 |      0 |       1
> ```
>
> Created/updated 2026-07-27 18:16:53. So there IS a live instance to prove the fix
> against after the roll, and the "no current damage" framing this file rests on has a
> half-life measured in minutes rather than days. Re-run the census before quoting it —
> and note the query in this file is scoped `WHERE p.status='active'`; archived pages
> are also 0 today, but the filter is not load-bearing and should not be copied as if
> it were (a council seat flagged exactly that shape).
>
> ### After the roll
>
> Pod-grep a string this change CREATED, then induce the failing branch:
> ```
> kubectl exec -n ai-persona-system <chassis pod> -- \
>   sh -c 'strings /app/agent-chassis | grep -c "assembled to nothing"'
> ```
> Then re-render `oufe.com/tool-recovery-waterfall` and confirm the run **fails** with
> the two lists named, rather than reporting COMPLETED. A green re-render of a healthy
> page proves the deploy, not the fix.

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

## Scale — measured 2026-07-27 (triage sweep): **zero live instances**

The fault shape is *"page has component rows, but none of them carries usable
`rendered_html`"* — that is what makes `assemblePage` return empty and the run report
`complete_skipped`. Fleet-wide:

```sql
SELECT s.domain, p.name, count(*) AS components,
       count(*) FILTER (WHERE pc.rendered_html IS NOT NULL AND pc.rendered_html <> '') AS usable
FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
GROUP BY 1,2 HAVING count(*) FILTER (WHERE pc.rendered_html IS NOT NULL AND pc.rendered_html <> '') = 0;
-- 0 rows
```

So this is a **defect class with no current damage** — the case for fixing it is the
silence, not a live page. Severity "medium-high" stands on that basis alone.

> **Two corrections to the mechanism as filed, both found by reading the code:**
>
> 1. **`slot_name` not matching `pages.sections` is NOT what makes assembly return
>    empty.** `getPageSections` (`rerender_single_page_action.go:509`) selects *every*
>    `page_components` row for the page and filters only on
>    `rendered_html IS NOT NULL AND rendered_html != ''` — it never consults
>    `pages.sections`. The `'main'` row rendered nothing because **nothing ever
>    populated its `rendered_html`**; the slot mismatch bites earlier, at the render
>    step that pairs planned sections to component rows, not at assembly.
> 2. Consequently a fleet scan for *"slot_name absent from `pages.sections`"* is **not**
>    a detector for this bug. It returns **70 rows across 12 sites** today, and they are
>    overwhelmingly benign — `loadComponentSchemas` keys by both component *name* and
>    *function*, so a slot matches either. Do not use that scan as evidence of damage;
>    use the one above.
>
> This matters for fix candidate 1: a CHECK constraint tying `slot_name` to the
> component `function` would still be correct, but it must be justified as closing the
> *render-pairing* hole, not as preventing an assembly failure — and it would have to
> reckon with those 70 existing rows.

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

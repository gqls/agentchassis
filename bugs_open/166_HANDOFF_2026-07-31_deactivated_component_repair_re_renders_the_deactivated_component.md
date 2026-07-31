# 166 — the `deactivated_component` repair re-renders the deactivated component, so the finding can never be satisfied

**Filed:** 2026-07-31 by the `bugfix_118_chrome_selection` lane, while fixing the
*cause* of the deactivated assignments. This is the *repair* half, and it is a
separate defect: 118's fix stops new bad assignments, it does not clear the 27
existing ones and it does not make these items closable.

**Severity:** medium. Nothing is broken that was not already broken — but a
detector that files an item its handler structurally cannot satisfy is worse
than no detector, because the queue then reads as "known, being handled" for ever.

> ## NARROWED 2026-07-31 (same day), after repointing the fleet by hand
>
> The owner's call was to repoint all 21 deactivated header/footer assignments, and
> doing it exposed **two further gates**, both of which had to be cleared by hand.
> Neither was visible from reading the check; both were found by dispatching
> `rerender-pages` at a real site and watching it report COMPLETED while changing
> nothing. **This is the more precise statement of the defect:**
>
> 1. **`rerender-pages` renders chrome only when `refresh_site_components` is true
>    in `input_data`** — there is a `check_refresh_components` conditional gating the
>    step. The detector DOES set it in its `SpecJSON`, so this half is correct by
>    construction and is not the bug. But dispatch the handler without that key —
>    as any hand-run or any other caller would — and the run completes having
>    skipped chrome entirely, silently.
> 2. **Even with the gate open and the assignment CORRECTED, the slot is skipped**,
>    because `renderAndStoreSiteComponent`'s `!force` idempotence exit tests
>    `rendered_html IS NOT NULL AND != ''` — not whether the component changed. A
>    repointed slot still holds its old HTML, so it looks "already rendered". I had
>    to NULL the stored artefact before the render would run.
>
> ⇒ **The repair needs `force_rerender: true`, and the detector does not set it.**
> That is a one-key fix on the check's `SpecJSON` and it is now the cheapest
> candidate. Without it, even the repoint this bug asks for does not reach the page.
>
> **Fix candidate 1 is DONE for the current fleet** (21 rows repointed, chrome
> re-rendered on 11 sites, 28/28 header+footer slots now active — see
> `bugs_closed/118`). What remains OPEN is the *mechanism*: the next deactivation
> puts the fleet straight back, and the routed repair still cannot clear it.

## The defect

`check_integrity.go:163` (`DeactivatedSiteComponentsCheck`) correctly finds every
`site_components` row pointing at an `is_active=false` component and files:

```go
ItemType:     "deactivated_component",
HandlerAgent: "rerender-pages",
SpecJSON:     {... "refresh_site_components": true},
Summary:      "Site component footer points to deactivated component 'footer-4-column'",
```

`render_site_components` renders **whatever `site_components.component_id` already
points at**. It re-renders the deactivated component, faithfully, and reports
success. The `component_id` is never re-examined, so the condition the item
describes is exactly as true after the repair as before it.

## Evidence

```sql
SELECT status, count(*), min(created_at)::date AS oldest
FROM site_work_items WHERE item_type='deactivated_component'
GROUP BY 1 ORDER BY 2 DESC;
```

Live 2026-07-31: items at `detected` and `unresolved` going back to **2026-07-17**,
two of them stamped `[unresolved after 2 attempts]` — the two-strike machinery
working correctly on a task that cannot succeed.

Meanwhile the state itself:

```sql
SELECT sc.slot_name, cc.name, cc.is_active, count(*) AS sites
FROM site_components sc JOIN content_components cc ON cc.id=sc.component_id
WHERE NOT cc.is_active GROUP BY 1,2,3 ORDER BY 4 DESC;
-- footer | footer-4-column      | f | 11
-- header | header-bold-gradient | f |  7
-- head   | Document Head        | f |  9
```

## Fix candidates, ordered by what closes the door

1. **Make the repair a REPOINT, not a re-render.** `render_site_components` (or a
   new narrow action) should, when the assigned component is ineligible and an
   eligible one exists for the slot's function, repoint `component_id` through the
   lock predicate (`pageComponentAgentWritableSQL`) and then render. 118 shipped
   `ResolveChromeComponent`, so the "which one instead?" half already exists and
   has one answer. **This is fleet-visible** — 11 sites' footers change — so it
   needs an owner call and a before/after on one site per layout first.
2. **If the repoint is not wanted, change the item's handler or retire the
   check.** An item nobody can satisfy should not be filed at `medium` severity
   into a queue that already has no working human surface (`bugs_open/033`).
   Retiring detection is the worse option — the state is real — but it beats a
   permanent false "in progress".
3. **`head` cannot be repaired either way**: there is no active head component at
   all (both candidates `is_active=false`), so the 9 `head` items have no valid
   target until the library gains one. That is a data call which also changes the
   build path's `<head>` fleet-wide (it falls through to `RenderFallbackHead`
   today).

## Verify a fix

Not by the item going `complete` — that is the failure mode. Ask the row and then
the page:

```sql
SELECT cc.name, cc.is_active FROM site_components sc
JOIN content_components cc ON cc.id=sc.component_id
JOIN sites s ON s.id=sc.site_id WHERE s.domain='relojistas.com' AND sc.slot_name='footer';
```
```sh
curl -s https://relojistas.com/index.html | grep -o '<h[34]>[^<]*</h[34]>'
# 'Our Services' = footer-4-column (deactivated) · 'Explore' = footer-theme-chrome (active)
```

## Related

- `bugs_open/118` — the cause (selection ignored `is_active`), fixed 2026-07-31.
  Its `LANDMINE` note in `LANDMINES.md` covers the trap this file describes.
- `bugs_open/083` — detected findings never reach a handler. Distinct: those are
  never picked up; these ARE picked up and the handler cannot help.
- `bugs_open/098` — archiving does not undeploy. Same retirement-flag family.
- `bugs_open/117` — stored chrome is never regenerated by a page re-render.

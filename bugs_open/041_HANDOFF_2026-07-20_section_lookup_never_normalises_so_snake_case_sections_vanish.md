# Handoff — section lookup never normalises, so a `snake_case` section silently vanishes and the platform asks to rebuild a component it already has

**Filed 2026-07-20.** Found while chasing a single missing section on dartsonline; it turned out to
be the smaller half of a two-month-old fleet-wide defect. **This is the upstream cause of most of
the "deployed page is short of its plan" damage catalogued in `/bugs_open/040-partial-build`.**

The platform has a normaliser for exactly this (`NormalizeComponentFunction`: `call_to_action` →
`call-to-action`). The section-lookup path does not call it.

## The mechanism

`loadSectionComponents` (`platform/orchestration/actions/v3_site_actions.go:3353`) resolves section
names to components in two passes, and **both use the raw string**:

```go
// Pass 1: lookup by name
args[i] = name                       // :3374  — raw
...
// Pass 2: lookup by function for anything still missing
funcArgs[i] = name                   // :3446  — still raw
```

`plan_sections_action.go` — its only caller for the build path, via `loadComponentSchemas` — contains
**zero** `Normalize*` calls; it passes `sectionNames` straight from the spec (`:684`).

**A sibling path does normalise**: `v3_site_actions.go:3730` calls `NormalizeSectionNames(...)`. So
whether a `snake_case` section resolves depends on which path ran — the platform's recurring
two-paths-one-guard shape (cf. `/bugs_open/024`, `/bugs_open/044`, and the rerender pair).

When both passes miss, `plan_sections` falls to **Path 3** (`plan_sections_action.go:762`): it
concludes no component exists, raises a `needs_new_component` work item, and marks the section
`deferred`. The build then completes and the page **deploys without that section**.

## Evidence

**1. The requests to build components that already exist.** Ten `needs_new_component` items, four
sites, 2026-05-18 → 2026-07-18, every one for `call_to_action`, every one `failed`, and in every
case the normalised name resolves to a real component:

```sql
SELECT s.domain, w.spec->>'section_type' AS requested, w.status,
       EXISTS (SELECT 1 FROM content_components c
               WHERE lower(c.function)=lower(replace(w.spec->>'section_type','_','-'))
                  OR lower(c.name)=lower(replace(w.spec->>'section_type','_','-'))) AS already_exists
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='needs_new_component' AND w.spec->>'section_type' LIKE '%\_%';
-- leopardess ×5, robot-hands ×3, gaswholesalers ×2 — all call_to_action, all failed, all already_exists=t
```

`content_components` holds `call-to-action` (function `call-to-action`) and `Call to Action` (same
function). There is no literal `call_to_action` row, and there never needed to be.

**2. A user-visible page.** `leopardessconsulting.co.uk/careers.html`:

```
pages.sections = ["hero", "generic-text-block", "call_to_action"]
page_components  = hero, generic-text-block            -- 2 of 3
```

Live: `<main>` has 2 `<section>` elements and **zero** CTA blocks. The page ends on a bare email
address in prose, where a call-to-action component should be. On a consultancy site, the conversion
element is the section being dropped.

**3. Fleet scale.** Of the sections missing from the 25 short deployed pages in
`/bugs_open/040-partial-build`, **`call_to_action` accounts for 14 occurrences and `hero` 6** —
against only 4 explained by legitimate `on_missing=skip_section` data guards. So the naming mismatch,
not the data guards, is the bulk of that sweep.

## What is NOT this bug

- **`on_missing=skip_section` is correct behaviour and must not be "fixed".** dartsonline's `index`
  drops `testimonials` because the component requires
  `site_specs.social_proof.testimonials` with `min_items: 1` and the site has none. The platform is
  refusing to invent customer quotes — exactly right, and recorded explicitly as
  `"reason": "on_missing=skip_section triggered"`. Do not conflate the two: one is a data-integrity
  guard working, the other is a string mismatch.
- The page being marked `deployed` regardless is `/bugs_open/040-partial-build`.
- The empty-schema name heuristic in `planSection` is `/bugs_open/044`.
- The function-vs-name namespace confusion is `/bugs_open/039`.

## Fix candidates

1. **Normalise inside `loadSectionComponents`, before both passes** (preferred). It is the single
   shared chokepoint every caller already goes through, so one change covers the build path, the
   selector path and any future caller. Match on the normalised value; keep the raw string for
   logging so a mismatch is still visible.
2. **Normalise where section names are written**, so `pages.sections` and `site_plan_sections` never
   store `snake_case`. Cleaner data, but does not protect against a caller passing an
   unnormalised name later, and needs a migration for the rows already stored.
3. **Both** — 1 as the guard, 2 as the clean-up. Given this platform's history of one-path fixes,
   1 is the one that must not be skipped.

Whatever is chosen, **Path 3 should not raise `needs_new_component` for a name that resolves under
normalisation** — that check is two lines and would have surfaced this in May.

## How to verify a fix

1. Build a page whose spec section is `call_to_action`. Assert the CTA component resolves, a
   `page_components` row is written, and **no** `needs_new_component` item is raised.
2. Assert `testimonials`-style `skip_section` behaviour is unchanged — a genuinely dataless section
   must still skip, and must still say so.
3. Re-run the 040 fleet sweep; the `call_to_action` and `hero` occurrences should disappear while
   the 4 `skip_section` ones remain.
4. Existing damage: the affected pages need a rebuild after the fix — the fix alone does not
   backfill a section that was never written. Check `/bugs_open/040-partial-build`'s note that a
   `deployed`+stamped page will be skipped by the reconciler, so those pages may need their
   `built_from_plan_version` cleared to be picked up at all.

## Related

- `/bugs_open/040-partial-build` — the page deploying anyway, and the sweep this explains.
- `/bugs_open/039` — the function/name namespace, and the normalisation that DOES exist.
- `/bugs_open/044`, `/bugs_open/024` — the same two-paths-one-guard shape, twice over.
- `/bugs_open/028` — a no-op page build reporting `complete`; adjacent consequence.

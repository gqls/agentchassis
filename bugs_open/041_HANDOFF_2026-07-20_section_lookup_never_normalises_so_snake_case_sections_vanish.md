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

---

## RESOLUTION — fixed AND LIVE on v1.0.1146, 2026-07-21 (behavioural end-to-end still to confirm)

> **Deployment note (the CLAUDE.md sweep-build landmine, in the flesh):** the four
> code files were committed not by this thread's `git commit` but by another
> session's **`v1.0.1146 - sweep`** commit (`fe2ba5e52`) that scooped up the
> working tree before this thread's own pathspec commit ran — the chassis builds
> from `HEAD`, so the sweep carried the complete fix. The running pod
> `agent-chassis-...` is `v1.0.1146` and its binary contains the guard's log
> literal `resolves under kebab-normalisation` (a string this fix introduced;
> pod-grep count 1, positive control 1). So the fix is **already live**, not
> inert. The separate test-file commit is `26ae571ac`.

**Diagnosis re-confirmed against the live DB before touching code.** The evidence
query still returns the same shape — and there is a **fresh** item stamped
`2026-07-21` (gaswholesalers `call_to_action`, `failed`, `already_exists=t`), so the
defect was still firing the day it was fixed. `content_components` holds
`call-to-action` (function `call-to-action`, `section_type` `call-to-action`,
active); there is no `call_to_action` row.

### One correction to fix-candidate 1 — it would have caused a NEW regression as written

Fix-candidate 1 said *"match on the normalised value"*. Taken literally
(normalise the input, then look up **only** the normalised form) it **breaks a
component that resolves today.** Scanning every `snake_case` section name that
appears in a live `pages.sections` against the library:

| raw section       | normalised        | resolves raw | resolves normalised |
|-------------------|-------------------|:---:|:---:|
| `call_to_action`  | `call-to-action`  | ✗ | ✓ (the bug) |
| `social_proof`    | `social-proof`    | ✓ | ✓ |
| `featured_article`| `featured-article`| **✓** | **✗** |
| `article_grid`    | `article-grid`    | ✗ | ✗ (genuinely absent) |
| `category_section`| `category-section`| ✗ | ✗ (genuinely absent) |

`featured_article` is a live component whose **name** is literally snake_case and
whose **function** is a *different* kebab string (`featured-content`). It resolves
**only** by its raw name; normalise-then-look-up would send it to Path 3. So the
fix must match on **raw OR normalised** — a strict superset that can never drop a
currently-resolving section. That is what shipped.

### The fix (4 changes, all strict supersets — nothing that resolved before stops)

1. **`component_validation.go`** — new naming-contract helpers `sectionLookupKeys`
   (raw + normalised, deduped), `sectionResolvedByFound`, `sectionLookupValueSet`.
2. **`loadSectionComponents` (`v3_site_actions.go`)** — the shared chokepoint. Both
   passes now query the raw+normalised value set; the found / missing / still-missing
   / reorder bookkeeping matches under either form. Raw name is kept for stubs and
   logs so a mismatch stays visible. This is fix-candidate 1, done safely.
3. **`loadComponentSchemas` (`plan_sections_action.go`)** — new
   `aliasNormalisedSectionKeys`: the returned map is keyed by the STORED (kebab)
   identifiers, but the build loop and the rerender path look it up by the
   *requested* name. Without an alias `components["call_to_action"]` still misses
   the `call-to-action` entry. The alias only ADDS keys (never rebinds), so a
   snake-named component keyed by its own raw name is untouched.
4. **`CreateNeedsNewComponentItem` (`component_selector.go`)** — the backstop the
   handoff asked for: before raising `needs_new_component`, if the normalised name
   resolves to an active component, log loudly and DON'T raise it. After 1–3 this
   branch should be unreachable for such names; it protects any future caller that
   bypasses them.

Unit tests in `section_normalisation_test.go` pin both properties: `call_to_action`
now resolves to the kebab component, and `featured_article` still resolves by its
raw name. Full `actions` package test green over clean `HEAD` (the shared tree was
independently broken by another session's unused `sort` import in
`component_library.go` — not this change; verified via `git archive HEAD` + these
files overlaid).

### Why this stays OPEN despite being live — and what remains

The code is live and both properties are proven at the unit level + deployment is
pod-verified. What is NOT yet done is the **behavioural end-to-end** confirmation on
the live binary (memory's "verify the failing branch": a deploy-grep proves the code
shipped, not that the running flow resolves the section):

- **Confirm on the next `call_to_action` build.** No `needs_new_component` for
  `call_to_action` has been created since the 12:15Z deploy (latest is 07:48Z,
  pre-deploy) — consistent, but that is absence-of-evidence until a plan_sections
  run with a `call_to_action` section actually executes on ≥ v1.0.1146. Verify per
  steps 1 & 3: CTA resolves → a `page_components` row is written → **no**
  `needs_new_component`. Move to `/bugs_closed/` only after that run is observed.
- **The already-damaged pages are not backfilled by the fix** (step 4). Pages that
  already deployed short of a `call_to_action`/`hero` section (leopardess,
  robot-hands, gaswholesalers) need a rebuild; a `deployed`+stamped page is skipped
  by the reconciler (see `/bugs_open/040`), so those may need
  `built_from_plan_version` cleared to be picked up. **Coordinate the rebuild with
  each site's owning workstream** — do not unilaterally regenerate an owned site's
  content (the leopardess careers page is owned by the leopardess-rebuild thread).
- The `skip_section` data-guard path (`testimonials`) is **untouched** — this change
  only affects component *resolution*, not the `on_missing` data requirement.

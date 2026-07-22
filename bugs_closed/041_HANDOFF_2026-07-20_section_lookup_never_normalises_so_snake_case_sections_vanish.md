# Handoff — section lookup never normalises, so a `snake_case` section silently vanishes and the platform asks to rebuild a component it already has

> **STATUS: CLOSED 2026-07-22 — fixed, LIVE on chassis v1.0.1146, and behaviourally
> verified on the failing branch** (gaswholesalers/index snake `call_to_action`
> resolved to a real `call-to-action` component post-deploy; 0 new
> `needs_new_component` since). See the CLOSED section below. This is the
> section-lookup 041, NOT the chrome-JS 041 (which stays open in `/bugs_open/`).
> Residual = a per-site page backfill, tracked with `/bugs_open/040` and the site
> owners — not this code bug.

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

## CLOSED — fixed, live on v1.0.1146, AND behaviourally verified on the failing branch (2026-07-22)

**Live behavioural proof, on real fleet traffic — no manual trigger needed.**
`gaswholesalers.com/index` carries the snake form in its stored `pages.sections`:
`[hero, features, services-grid, differentiators-section, social_proof,
latest-news, call_to_action]`. The before/after is exact:

| when | binary | what happened |
|---|---|---|
| 2026-05-20 & 2026-07-21 **07:48Z** | pre-fix | `needs_new_component` for `call_to_action` raised, **failed** (the bug) |
| 2026-07-21 **12:15Z** | — | v1.0.1146 (this fix) deploys |
| 2026-07-21 **19:55Z** | post-fix | same page's snake `call_to_action` **resolved to the real `call-to-action` component** — `page_components` row written with `component_id=0197e8d7-1adc-43d6-ab32-d0716e013175` (function `call-to-action`), **not a stub** |

Fleet-wide since the 12:15Z deploy: **8** `call-to-action` `page_components` written,
**0** new `call_to_action` `needs_new_component` items. That is the failing branch
(snake → vanish) resolving, verified on the live binary — the last gate the deploy-
grep alone could not clear (cf. "verify the failing branch").

The other 7 CTA writes were pages whose section name was already kebab
`call-to-action` (would have resolved pre-fix too, via the by-function pass) — only
`gaswholesalers/index` is the discriminating snake_case case. `gaswholesalers/index`
itself is `needs_rebuild` (bugs_open/040's partial-build guard, also live in 1146) —
independent of this bug: the CTA *resolved*, which is all 041 was about.

### RESOLUTION — fixed AND LIVE on v1.0.1146, 2026-07-21

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

### The one residual — NOT a reason to keep 041 open

The defect is fixed, live, and verified non-reproducing, so the CODE bug is CLOSED.
What remains is a **data backfill**, which the fix explicitly never promised (step 4):

- **The already-damaged pages are not backfilled by the fix.** Pages that already
  deployed short of a `call_to_action`/`hero` section (leopardess careers,
  robot-hands, and others) still need a rebuild to restore the dropped section; a
  `deployed`+stamped page is skipped by the reconciler (see `/bugs_open/040`), so
  those may need `built_from_plan_version` cleared to be picked up. **This is a
  per-site cleanup to coordinate with each owning workstream** — regenerating an
  owned site's content is outward-facing and re-runs the LLM copy (claims-audit
  surface on e.g. leopardess), so it is the site owner's call, not a platform sweep.
- The `skip_section` data-guard path (`testimonials`) is **untouched** — this change
  only affects component *resolution*, not the `on_missing` data requirement.

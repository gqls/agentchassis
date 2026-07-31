# PLAN — one predicate for "which library component serves a chrome slot"

**Started 2026-07-31.** Taking `bugs_open/118` (chrome component selection ignores
`is_active` and picks alphabetically). Verified valid first-hand before starting;
the verification changed the shape of the fix twice, so the corrections below are
the load-bearing part of this file.

## The question the platform asks in four different places

"Which library component serves chrome function F?" — asked by:

| # | call site | predicate today | what it picks, live 2026-07-31 |
|---|---|---|---|
| A | `render_site_components_action.go:552` (slot has no assignment) | **none at all**, `ORDER BY name LIMIT 1` | `header-bold-gradient` (INACTIVE), `footer-4-column` (INACTIVE), `Document Head` (INACTIVE) |
| B | `link_site_components_action.go:96,109` | `is_active` only | **`header-leopardess` — an ACTIVE FORK of one client's header**, `footer-theme-chrome` |
| C | `GetComponentByFunction` (`component_library.go:175`) | `is_active AND forked_from IS NULL`, **no `ORDER BY`** | `site-header` / `site-footer` — both `component_level='section'`, i.e. page-section components, not chrome |
| — | `queryCandidates` (section selector) | `is_active AND forked_from IS NULL AND component_level='section'` | the one that is right, and it is the one nobody copied |

Every row of that table was run against the live DB, not read off the code — see
`RUNBOOK_chrome_selection.md`.

## Corrections to the filed bug, both found by measuring

> **CORRECTION 1 (2026-07-31) — candidate 1 does NOT "change the rendered footer on
> every site", and that was the reason it was parked for an owner call.**
> The fallback at call site A fires **only when the slot has no `site_components`
> row**. All 14 real sites already have header/footer/head rows, so their render is
> pinned by `component_id` and is untouched by any change to the selection query.
> Live blast radius of fixing A: **`loancalculator.co.uk`** (created 2026-07-30, zero
> chrome rows) and every site created after it. That is a forward-only fix, and it
> needs no owner call.
> What *does* need an owner call is repointing the 11 sites already assigned to
> `footer-4-column` — which candidate 1 never proposed to do. The bug file collapsed
> the two.

> **CORRECTION 2 (2026-07-31) — "add `AND is_active`" as written would have shipped
> a client's forked header to every new site.**
> The bug file's candidate 1 says the remaining tie-break is between
> `footer-theme-chrome` and `site-footer`. For the *header* the alphabet's answer
> under `AND is_active` is `header-leopardess`, `forked_from IS NOT NULL` — a fork of
> leopardessconsulting.co.uk's header. `GetComponentByFunction`'s own doc comment
> has said since it was written that forks "should only be accessed by
> `component_id`"; two of the three call sites never honoured it. **A doc comment
> enforces nothing.**

> **CORRECTION 3 (2026-07-31) — `is_active` + `forked_from` is still not the chrome
> predicate.** With only those two, the winner for `site-header` is `site-header`
> itself, which is `component_level='section'` — a 6.6KB page-section component, not
> site chrome. `016b` already recorded the principle ("`site-head`
> (`component_level=section`) is unreachable as chrome"); the predicate never
> encoded it. The chrome pool is `component_level IN ('site','header','footer','head')`.

## The fix

**One resolver, `ResolveChromeComponent`, in `component_library.go`** — the same
shape as `assetAgentWritableSQL`/LOCK-007 and `pageComponentAgentWritableSQL`: the
predicate exists as a single named constant, callers use it, and a source-scanning
test fails when someone hand-types a fourth copy.

```
WHERE function = $1
  AND is_active
  AND forked_from IS NULL
  AND component_level IN ('site','header','footer','head')
ORDER BY name LIMIT 1
```

With the level filter there is exactly **one** eligible row per chrome function, so
the tie-break question the bug file raised disappears rather than being answered by
the alphabet twice:

- `site-header` → `header-theme-chrome` (sole eligible)
- `site-footer` → `footer-theme-chrome` (sole eligible)
- `head` → **none** — both candidates are `is_active=false`

The `head` gap is why the resolver returns `(component, eligible bool, error)` rather
than just erroring: it answers with the best available row and tells the caller the
library had nothing legitimate. Callers log that at ERROR and report it. Refusing
outright would take a working (if deactivated) head away from new sites and lose
`injectBrandHeadTags` (favicon + og-card) with it — a regression traded for purity.

### Scope: the two ASSIGNMENT call sites, not the build path

A and B are switched to the resolver. C (`GetComponentByFunction`) is **not** —
switching it changes chrome markup on every page build fleet-wide (from
`site-header`/`site-footer` to the `*-theme-chrome` pair), which is a visible change
nobody asked for and is a different defect from the one 118 filed. It is measured,
written down, and handed to the owner as a question rather than shipped as a side
effect. `GetComponentByFunction` does get `ORDER BY name` so its answer is
guaranteed rather than incidental — measured to change nothing (below).

### Blast radius, measured before submitting

- Functions with >1 eligible row fleet-wide: **exactly 2**, `site-header` and
  `site-footer`. Adding `ORDER BY name` to `GetComponentByFunction` returns the same
  row it returns today for both, so its answer is unchanged and now deterministic.
- Sites affected by A and B: **1 today** (`loancalculator.co.uk`) plus new sites.
- Pages re-rendered by this change: **0**. Nothing repoints an existing assignment.

## Left open deliberately

1. **Repoint the 11 sites on `footer-4-column` / 7 on `header-bold-gradient`** —
   owner call, fleet-visible. The platform already *detects* it
   (`deactivated_site_components` → `deactivated_component` items, raised since
   2026-07-17) and routes them to `rerender-pages`, which re-renders **the same
   deactivated component** — so the repair is structurally incapable of repairing and
   the items age to `unresolved`. That is its own defect; filed separately.
2. **No active `head` component exists.** Activating one changes the build path's
   `<head>` fleet-wide (today it falls through to `RenderFallbackHead`). Data call,
   not code.
3. **The build path renders section-level components as chrome** (correction 3).
   Filed separately.

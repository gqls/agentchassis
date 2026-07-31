# PLAN — 2026-07-31 — bugs_open/167: the page-build path can render a section component as chrome

## The brief

`bugs_open/167`, filed today by the `bugfix_118_chrome_selection` lane as the half
of 118 it deliberately did not fix. Three chrome renderers in
`component_library.go` resolve their component through `GetComponentByFunction`,
which has **no `component_level` filter**, so a `component_level='section'`
component can be selected to serve as site chrome.

118 left it because it believed fixing it would change served markup on every
page build fleet-wide, and it did not want a fleet-visible change hidden inside a
zero-visible-change fix. That reasoning is sound and is why this is a separate
bug.

## CORRECTION to the filed bug — its blast-radius table is STALE

> **CORRECTED 2026-07-31 (this lane):** the bug file's "Live 2026-07-31" table
> says `site-header` resolves to `site-header` (`component_level='section'`,
> 6,614 chars) and `site-footer` to `site-footer` (section, 7,519). **That is no
> longer true, and it stopped being true on the day it was written.**

Measured live, running each function's own query verbatim:

| function | `GetComponentByFunction` (today) | `ResolveChromeComponent` (proposed) |
|---|---|---|
| `site-header` | `header-theme-chrome` **[site]** | `header-theme-chrome` [site] eligible=true |
| `site-footer` | `footer-theme-chrome` **[site]** | `footer-theme-chrome` [site] eligible=true |
| `head` | *no row → `RenderFallbackHead`* | `Document Head` [section] **eligible=false** |

**The two predicates already agree.** What changed: `header-theme-chrome` was
activated at `2026-07-31 12:39:53` (`content_components.updated_at`), in the same
second as three sibling headers were deactivated — i.e. by 118's own fleet
repoint, hours before 167 was filed. With it active, `ORDER BY name` puts
`header-theme-chrome` ahead of `site-header`, and `footer-theme-chrome` ahead of
`site-footer`. Before that repoint the only active unforked row for each function
was the section-level one, which is exactly the table the bug file records.

**So the owner call the bug file asks for has already been answered by the data,
and the answer is that candidate 1 costs nothing today.** The reason to ship it is
not that it changes an answer — it is that today's correct answer is an
**accident of alphabetical order** (`h` < `s` twice over), and one deactivation
or one new component named `a-…` restores the defect silently.

## The decision: candidate 1, plus the door that made it necessary

The bug file offers three. Taking them in the order of *what closes the door*:

- **Candidate 3 (re-point the section-level rows to their own `function`)** is
  data, not code. It fixes today's two rows and leaves the code able to pick a
  section component as chrome for ever. Rejected as the primary fix.
- **Candidate 2 (accept the asymmetry)** is what filing the bug already did.
- **Candidate 1 (route the three renderers through `ResolveChromeComponent`)** is
  the framework fix: one predicate answers "which component serves chrome
  function F", for every caller, and the answer stops depending on `ORDER BY name`
  luck.

Candidate 1 alone is not enough, though, because **the guard 118 shipped cannot
see these three call sites**. `TestNoChromeSelectionHandTypesItsOwnLookup` scans
for hand-typed SQL (`function = 'site-header'`) and **skips `component_library.go`
entirely**. The three defective call sites are in that file and are not SQL — they
are Go calls passing a chrome function name as a string literal. So:

1. **Fix** — the three renderers call `ResolveChromeComponent(ctx, db,
   ChromeSlotFunction(slot), logger)` and use the component **only when
   `eligible`**; otherwise the existing purpose-built fallback renderer runs.
2. **Guard** — a new scan that catches the *Go-level* form: a chrome function name
   passed as a literal to `GetComponentByFunction` / `GetComponentWithFallback`,
   **including inside `component_library.go`**. This is the door; without it the
   next writer reintroduces the bug in the file the existing guard exempts.

### Why the eligibility gate must be a gate, not a default

`head` is the case that decides this. Today it has **no eligible component at
all** (both candidates `is_active=false`) and `GetComponentByFunction` returns no
row, so `RenderFallbackHead` runs. `ResolveChromeComponent` **always answers** —
by design, so the caller can report the gap rather than lose the slot — and its
answer for `head` is `Document Head`, a **`component_level='section'` component**.

A naive swap therefore *introduces* 167's exact defect on the head slot: an 8,523
char section component rendered as `<head>`. The bug file asserts "the resolver
already reports `eligible=false` so `head` keeps falling through to
`RenderFallbackHead` exactly as today" — true **only if the caller reads the
flag**. Gating on `eligible` is what makes the claim true, and it is pinned by a
test.

## Blast radius — measured, not argued

- **Header and footer: no change to any served byte.** Both predicates return the
  same component for both functions (table above), so every page build renders
  what it renders today.
- **Head: no change.** `eligible=false` → `RenderFallbackHead`, which is what runs
  today.
- The by-function branch is reached by **every site whose style collection has no
  chrome pinned** — which is all 10 per-site `collection-*` collections and both
  calculator sites (`header_component_id IS NULL`). So the path is live and
  heavily used; it simply already resolves correctly.
- `GetComponentByFunction` itself is **unchanged**, so its section callers
  (`GetComponentWithFallback` → `generic-text-block`) are untouched.

## Out of scope, and filed rather than fixed

`RenderHeader`/`RenderFooter` try `coll.HeaderComponentID` **before** the
by-function branch, and that branch calls `GetComponentByID`, which applies **no
eligibility predicate at all**. Live:

| domain | pinned header | `is_active` |
|---|---|---|
| ai-agent-orchestration.com | `header-professional-dark` | **false** |
| finetuning.uk | `header-professional-dark` | **false** |
| gaswholesalers.com | `header-professional-dark` | **false** |
| leopardessconsulting.co.uk | `header-leopardess` | true (a fork — correct) |

Three deployed sites render a **deactivated** header on every page build. That is
118's defect class (deactivated component as chrome) surviving on a **fourth**
path 118 did not enumerate — not 167's (section as chrome; all four pins are
`component_level='site'`). Fixing it moves those three sites from
`header-professional-dark` (3,637 chars) to `header-theme-chrome` (2,551) — a
**fleet-visible markup change on live sites**, which is the exact thing 118
refused to hide inside another fix. Filed as its own bug; not touched here.

## Phasing

1. Working docs (this file), then the code fix + guard test in a **new** test file.
2. `go build` + `go test` against `git archive HEAD` (shared tree — a green local
   build is not a green HEAD).
3. Council gate submission (platform code), commit with `Council-Submitted:`.
4. Close 167 → `bugs_closed/`, register the guard, file the collection-pin bug.

## Concurrency

`chrome_selection_test.go` and `render_site_components_action.go` are being edited
**right now** by the session working `bugs_open/166` (transcript `b5017ee5`,
written to seconds before this plan). My tests therefore go in a **new file**,
`chrome_build_path_test.go`, and I touch neither of theirs. A same-file passenger
is the one thing a pathspec commit cannot protect against.

# 167 — the page-build path resolves site chrome to a `component_level='section'` component

---

## RESOLVED 2026-07-31 — fix candidate 1, committed `8b29404d6`

> **⚠ STATUS: FIXED AND COMMITTED, NOT YET LIVE — and this file has been moved to
> `bugs_closed/` anyway, at the owner's explicit instruction.** That is a
> deliberate departure from this repo's stated bar ("fixed AND live — a fix
> committed but inert until the next roll stays OPEN"), recorded here rather than
> left for a reader to discover. **Until it rolls, the defect is still
> reproducible in production.**
>
> Measured at the running binary, 2026-07-31 21:2x UTC, both replicas on
> `v1.0.1223` (started 20:20 UTC, an hour *before* the fix commit):
> ```
> MY NEW STRING (expect 0): 0
> POSITIVE CONTROL ("no component serves chrome function"): 1
> ```
> The control is 118's own string, so the grep is proven to work on this binary —
> the 0 is a real absence, not a broken check (`bugs_open/153`).
>
> **To confirm it has shipped**, after the next chassis roll:
> ```
> kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
>   'strings /app/agent-chassis | grep -c "no eligible header component in the library"'
> ```
> with a positive control in the same exec. Another session had `IMAGE_TAG` at
> `v1.0.1224` uncommitted and was mid-build when this was written; `make build-*`
> builds from committed HEAD, so this fix rides that roll — **but that is an
> expectation, not a verification.**

**What shipped:** all three chrome renderers (`RenderHeader`, `RenderFooter`,
`RenderHead`) now resolve through `ResolveChromeComponent` and use its answer
**only when it reports `eligible`**, otherwise falling through to the existing
fallback renderer. Plus a static guard that fails if a chrome function name is
ever passed to the section-shaped lookup again.

**The owner call this file asks for was already answered by the data — see the
correction below.** Both predicates return the same component today, so the fix
changes no served header or footer byte.

**Council: APPROVED at round 3** — `d73a4b06-a190-426e-bdf7-18d830d06a9d`, 11 of 13
seats approving, 2 advisory objections, none high. The trail is worth reading before
anyone touches this area again:

- **r1 REVISE** (gated by `bug_historian`): shipping with the fourth chrome path
  (`bugs_open/170`) unguarded. Five other seats independently asked whether this fix
  can reach served output at all, given a landmine naming these three functions
  beside `site_components`' `!force` exit. **Answered by measurement:**
  `component_library.go` has **0** references to that table and **0** to `force`; the
  cached path is `renderAndStoreSiteComponent` with **22**. The landmine's *footprint*
  matched, its *mechanism* did not.
- **r2 REVISE** (gated by `reuse_agent`): the per-render diagnostic added in r1 was
  itself the defect — no reader, unbounded noise, a bespoke reporter where a mechanism
  already existed. **Four seats, four angles, all correct.** Removed, not rate-limited.
  Checking the prior art then produced the finding that outlived the code: the
  platform's own `deactivated_site_components` detector **joins `site_components` only
  and is blind to style-collection pins**, which is why three deployed sites have had
  no work item. That is now `bugs_open/170` candidate 1b.
- **r3 APPROVED**, on the removal. `reuse_agent`, `prior_art_librarian` and
  `architecture` all flipped to approve.

**Two advisory objections stand, and they are honest residuals, not oversights.**
`bug_historian` (medium ×2): removing the diagnostic returns the pin path to *zero*
signal — a code comment and a filed bug are all that mark it — and candidate 1b is
specified but not implemented. Both are true. The counter-argument, on the record: a
report nobody reads is not a signal, `discovery_checks/` belongs to the `bugs_open/149`
lane and its test file was dirty in another session's tree, and an item routed to
`rerender-pages` would be unsatisfiable by construction — `bugs_open/166` on a new item
type. **Whoever picks up 170 should treat those two objections as the brief.**

**Spun out, not fixed here:** `bugs_open/170` — the style-collection chrome *pin*
is dereferenced by `GetComponentByID`, which applies **no eligibility predicate at
all**, and three deployed sites are pinned to an `is_active=false` header. That is
118's class on a fourth path, and fixing it *is* a visible markup change.

**Workstream docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_167_chrome_build_path/`.

---

**Filed:** 2026-07-31 by the `bugfix_118_chrome_selection` lane. 118 fixed the two
*assignment* call sites and deliberately left this one, because fixing it changes
chrome markup on every page build fleet-wide. **It is an owner call, not an
oversight**, and it is written down here rather than left as an asymmetry nobody
records.

**Severity:** medium, and entirely latent-looking — nothing errors, the pages
render, the markup is simply produced by a component that was never meant to be
chrome.

## The defect

`GetComponentByFunction` is the library's generic "component serving function F"
lookup, and it is correct for what it is: `is_active = true AND forked_from IS
NULL` (`ORDER BY name` added by 118). It has **no `component_level` filter**,
because its other callers ask it for SECTION functions (`generic-text-block`, via
`GetComponentWithFallback`) — adding one there would break every section lookup.

Five chrome call sites use it:

```
component_library.go  RenderHeader -> GetComponentByFunction(ctx, db, "site-header")
component_library.go  RenderFooter -> GetComponentByFunction(ctx, db, "site-footer")
component_library.go  RenderHead   -> GetComponentByFunction(ctx, db, "head")
```

reached from `InjectHeader`/`InjectFooter`/`InjectHead` in `multipage_actions.go`,
`v3_site_actions.go` and `rerender_pages_actions.go`.

Live 2026-07-31, what those calls resolve to:

| function | resolves to | `component_level` | length |
|---|---|---|---|
| `site-header` | `site-header` | **`section`** | 6,614 |
| `site-footer` | `site-footer` | **`section`** | 7,519 |
| `head` | *(nothing — both candidates `is_active=false`)* | — | — |

> **CORRECTED 2026-07-31 by the `bugfix_167_chrome_build_path` lane — the two rows
> above were true when written and were false within hours, which is why the
> "owner call" this file asks for no longer applied.** Re-running each function's
> own query verbatim:
>
> | function | `GetComponentByFunction` | `ResolveChromeComponent` |
> |---|---|---|
> | `site-header` | `header-theme-chrome` **[site]** | `header-theme-chrome` [site] eligible=true |
> | `site-footer` | `footer-theme-chrome` **[site]** | `footer-theme-chrome` [site] eligible=true |
> | `head` | no row → `RenderFallbackHead` | `Document Head` [section] eligible=**false** |
>
> **The two predicates already agree.** `content_components.updated_at` shows
> `header-theme-chrome` was activated at `2026-07-31 12:39:53` — the same second
> three sibling headers were deactivated, i.e. **118's own fleet repoint**, hours
> before this file was written. With it active, `ORDER BY name` sorts
> `header-theme-chrome` ahead of `site-header` and `footer-theme-chrome` ahead of
> `site-footer`.
>
> So candidate 1 costs **nothing visible today**, and the reason to ship it is not
> that it changes an answer — it is that today's correct answer is an **accident
> of alphabetical order, twice**, on a tie-break that knows nothing about chrome.
> One deactivation, or one new component named `a-…`, restores the defect silently.
> **What caught it:** re-running the measurement instead of quoting it. The figures
> were dated and correct; the tree they described was being changed by another lane
> at the same time.

The intended chrome components — `header-theme-chrome` (2,551) and
`footer-theme-chrome` (1,575), both `component_level='site'` — are not what this
path picks. `016b` §9 already recorded the principle: *"`site-head`
(`component_level=section`) is unreachable as chrome"*. It is reachable.

## Why 118 did not fix it

`ResolveChromeComponent` (shipped by 118) is exactly the right predicate for
these three call sites. Pointing them at it flips every page build's header and
footer from the `site-*` pair to the `*-theme-chrome` pair: different columns,
different headings, different CSS. That is a **visible fleet-wide change**, which
is the one thing 118's own measurement established the assignment fix was NOT.
Shipping both under one bug number would have hidden a fleet-visible change
inside a zero-visible-change fix.

## Fix candidates

1. **Point the three chrome renderers at `ResolveChromeComponent`.** One-line
   each; the resolver already reports `eligible=false` so `head` keeps falling
   through to `RenderFallbackHead` exactly as today. **Needs a before/after on one
   site per layout, and an owner go**, because it changes served markup.

   > **CORRECTED 2026-07-31 (implementing lane) — "one-line each" is a trap, and
   > this is the sentence to read twice.** `ResolveChromeComponent` **always
   > returns a row**, by design. `head` has no eligible component live, and its
   > last-resort answer is `Document Head` — an 8,523-char
   > `component_level='section'` component. A literal one-line swap of the lookup
   > therefore renders a page section as `<head>`, **creating this very bug on the
   > one slot that never had it.** `head` keeps falling through to
   > `RenderFallbackHead` *only if the caller reads the `eligible` flag and treats
   > it as a gate*. Shipped that way, and pinned by a test that fails if the
   > ineligible component's body reaches the output.
2. **Do nothing and accept the asymmetry**, i.e. the assignment path can no longer
   pick a section component as chrome and the build path still can. Defensible
   only while it is written down — which is what this file is for.
3. **Retire the confusion at the data layer**: re-point `site-header`/`site-footer`
   (the section-level rows) to their own `function` so they leave the chrome pool
   entirely. Cheapest, and it makes candidate 1 a no-op change instead of a
   markup change — but check first what plans them as sections
   (`plan_sections_action.go:1864-1866` names both), because that is the reuse
   those rows exist for.

## How to check the current answer before acting

```sql
SELECT function, name, component_level, is_active
FROM content_components
WHERE function IN ('site-header','site-footer','head') AND is_active AND forked_from IS NULL
ORDER BY function, name;
```

Two eligible rows for each of `site-header` and `site-footer`; the `*-theme-chrome`
one is chrome-level, the `site-*` one is section-level. Which one this path
returns was **arbitrary** before 118 (no `ORDER BY` over a two-member pool) and is
now `site-*` by name order — measured to be the same row it was returning
already, so 118 pinned the status quo rather than changing it.

## Related

- `bugs_open/118` — the assignment half, fixed 2026-07-31; concept register CLC-013
  states this asymmetry as its open review question.
- `bugs_open/166` — the repair that cannot repair.
- `bugs_open/117` — stored chrome is never regenerated by a page re-render, which is
  why the build path and the stored artefact can disagree for days.

# 118 — chrome component selection ignores `is_active` and picks alphabetically, so every site's footer renders from a DEACTIVATED component

**Found:** 2026-07-27, on relojistas.com, while trying to work out why an owner-approved
edit to the active footer component changed nothing on any page.
**Severity:** medium-high. It is fleet-wide, it silently overrides every deactivation of a
chrome component, and it makes editing the *active* chrome component a no-op — so the
natural repair for any chrome defect fails in a way that looks like the repair was wrong.
**Status:** **CLOSED 2026-07-31 — FIXED, LIVE AND POD-VERIFIED on `v1.0.1219`, AND the fleet
repaired.** Commits `b052249d8` + `a77034379` + `db0de5656`, council **APPROVED at round 1**
(`5bc232d6-590a-4476-a6b1-4fb6f61751c6`, 5 advisory objections, 3 answered in code), concept
register **CLC-013**, workstream
`docs/agent_docs/docs024_key_docs_latest/bugfix_118_chrome_selection/`.

**Live proof, at the artefact and not at the tag** (both replicas of `v1.0.1219`, positive
control in the same exec): `no eligible component for function` = 1,
`component_level IN ('site','header','footer','head')` = 1, `ineligible_chrome` = 1, control
`RenderSiteComponentsAction` = 6.

**Fleet repaired on the owner's call (2026-07-31, "repoint all eleven now"):** 21 assignments
repointed under the `069` lock predicate — **11 footers** `footer-4-column` →
`footer-theme-chrome`, **10 headers** `header-bold-gradient`/`header-professional-dark` →
`header-theme-chrome`; `leopardessconsulting.co.uk` keeps its own active fork by
`component_id`, which is exactly what a fork is for. Chrome re-rendered on all 11 sites.
**Result: 28 of 28 header/footer slots across all 14 sites now render from an ACTIVE
component; zero deactivated header/footer assignments remain.** Backup of the prior mapping:
`site_components_repoint_backup_20260731`.

**Verified at the artefact on the motivating site:** relojistas' stored footer now emits
`<h4>Explore</h4>` (the documented tell of `footer-theme-chrome`) where it emitted
`<h4>Our Services</h4>` (`footer-4-column`) for the previous eleven days — and the Contact
column is correctly ABSENT, i.e. `bugs_open/111`'s gate finally working on the component that
actually renders, which is the change whose silent failure filed this bug.

**Two residuals, both filed, neither hidden:** `head` still has NO eligible component
fleet-wide (13 assignments, both candidates `is_active=false`) — a library gap, not a
selection bug; and **206 `page_rerender` items are queued at `triaged`** — the stored chrome
is correct on every site but the DEPLOYED pages serve the old footer until those drain
(`bugs_open/117`: chrome is a stored artefact; the queue is `bugs_open/149`'s lane).

~~**FIXED IN CODE 2026-07-31, OPEN ONLY UNTIL THE CHASSIS ROLLS.**~~ Commits
`b052249d8` + `a77034379`, council submission `5bc232d6-590a-4476-a6b1-4fb6f61751c6`,
concept register **CLC-013**, workstream
`docs/agent_docs/docs024_key_docs_latest/bugfix_118_chrome_selection/`.
~~The structural fix changes the rendered footer on every site, so it needs an owner
call, not a site thread's.~~ **That is wrong — see CORRECTION 1 below. It was also the
reason this sat untouched for four days**, so the correction is the most useful line
in the file.

> ## THREE CORRECTIONS, all from measuring rather than reading (2026-07-31)
>
> **CORRECTION 1 — candidate 1 does NOT "change the rendered footer on every site".**
> `renderAndStoreSiteComponent`'s fallback fires **only when the slot has no
> `site_components` row**. All 14 real sites already have `header`/`footer`/`head`
> rows, so their render is pinned by `component_id` and no change to the selection
> query can touch it. Measured live blast radius: **`loancalculator.co.uk`** (created
> 2026-07-30, zero chrome rows) and every site created after it. Pages re-rendered by
> the fix: **zero**.
> What genuinely needs an owner call is **repointing** the 11 sites already assigned
> to `footer-4-column` — which candidate 1 never proposed to do. This file collapsed
> "fix the selection" and "repair the fleet" into one decision; only the second is
> fleet-visible.
>
> **CORRECTION 2 — "add `AND is_active`" as literally written would have shipped a
> client's FORKED header to every new site.** The tie-break note below considers only
> the footer. For the header, `AND is_active ORDER BY name` picks **`header-leopardess`**,
> `forked_from IS NOT NULL` — a fork of leopardessconsulting.co.uk's header. A fork
> carries its parent's `function`, so it sits in the generic pool.
> `GetComponentByFunction`'s own doc comment has said forks "should only be accessed
> by `component_id`" since it was written; two of the three call sites never honoured
> it. **A doc comment enforces nothing.**
>
> **CORRECTION 3 — `is_active` + `forked_from` is still not the chrome predicate.**
> With only those two, the `site-header` winner is **`site-header` itself, which is
> `component_level='section'`** — a 6.6KB page-section component, not chrome. `016b`
> §9 recorded the principle ("`site-head` (`component_level=section`) is unreachable
> as chrome") long before any predicate encoded it. The chrome pool is
> `component_level IN ('site','header','footer','head')` — four values, because the
> vocabulary grew twice. **With the level filter there is exactly ONE eligible row per
> chrome function, so the tie-break question below disappears rather than being
> answered by the alphabet twice.**

## The query

`render_site_components_action.go:545-556` maps each chrome slot to a `function` and then:

```go
"footer": "site-footer",
…
FROM content_components
WHERE function = $1
ORDER BY name LIMIT 1
```

**No `is_active` predicate.** The winner is whichever row sorts first by `name`.

## What that selects today

```sql
SELECT name, is_active FROM content_components WHERE function='site-footer' ORDER BY name;
```

| name | `is_active` | selected |
|---|---|---|
| **`footer-4-column`** | **false** | **← this one** |
| `footer-simple` | false | |
| `footer-standard` | false | |
| `footer-theme-chrome` | **true** | |
| `site-footer` | **true** | |

Three of the five candidates are deactivated, and the alphabetically-first — a deactivated
one — wins on every site. Both *active* components are unreachable through this path.

## Proof, from the live site

relojistas.com serves `<h4>Our Services</h4>` in its footer. That string exists in exactly
one component fleet-wide:

```sql
SELECT name, is_active FROM content_components WHERE html_template ILIKE '%Our Services%';
-- footer-4-column | f
```

The active `footer-theme-chrome` says `<h4>Explore</h4>`. The live site has never shown it.

**Confirmed against a fresh render, not just a stale artefact.** A site-chrome rebuild ran
at 17:18 on 2026-07-27 (`nav_drift` → `nav-updater` → `render_site_components`). The
regenerated `site_components.footer` — written minutes earlier — still contained
`Our Services`. So this is the live selection behaviour, not a leftover.

## How it wasted a change, which is the reason it is filed

An owner-approved fix gated the footer's contact column on `{{if or .email .phone}}`, so
sites with no contact details stop rendering an empty `<h4>Contact</h4>` over blank space
(`bugs_open/111`). It was applied to `footer-theme-chrome` — the **active** component, which
is the obvious and, as it turns out, wrong target. It had no effect, on any page, after a
full chrome rebuild.

The same one-line gate then had to be applied to `footer-4-column`, the deactivated
component that actually renders. **Both now carry it**, which is itself a symptom: a fix has
to be duplicated across an active and a deactivated component to be sure it lands.

`footer-4-column`'s contact block was also strictly worse than the active one's — its email
and phone lines were **completely ungated**:

```
<div class="footer-contact">
    <h4>Contact</h4>
    <p><a href="mailto:{{.email}}">{{.email}}</a></p>   <- no {{if}}
    <p>{{.phone}}</p>                                    <- no {{if}}
</div>
```

which is where the empty `<a href="mailto:"></a>` on relojistas came from (visible only after
decoding Cloudflare's email obfuscation — see `111`).

## Why `is_active = false` is not doing what anyone thinks

Deactivating a chrome component **does not retire it**. `is_active` is respected by whatever
sets it and by other selection paths, but not here — so a component someone deliberately took
out of service has been rendering the footer of every site in the fleet. This is the same
shape as `bugs_open/098` ("archiving does not undeploy") and `bugs_open/117` (stored chrome
is never regenerated): a retirement flag that retracts something from the platform's model
while the thing itself keeps serving.

## Fix candidates, ordered by what closes the door

1. **Add `AND is_active` to the selection, and make the empty result loud.** Makes
   "deactivated component in production" unrepresentable. **This changes the rendered footer
   on every site** (`footer-4-column` → `footer-theme-chrome`: different columns, different
   headings, different CSS), so it is a visible fleet-wide change and needs an owner call plus
   a before/after on at least one site per layout. Note the tie-break still matters afterwards:
   two active candidates remain (`footer-theme-chrome`, `site-footer`) and `ORDER BY name`
   would pick `footer-theme-chrome` — verify that is the intended one rather than accepting
   the alphabet's answer twice.
2. **Make the choice explicit per site** — a `chrome` spec naming the component, defaulting to
   a named constant rather than to sort order. Removes the alphabetical accident permanently
   and gives per-site override, which the estate will want anyway (see the language problem in
   `about_page_commercial`). Bigger change.
3. **Delete or rename the deactivated components.** Cheapest, closes nothing — the next
   component whose name sorts early reintroduces it.

**Do candidate 1 or 2 before anyone else edits chrome**, because until then every chrome edit
must be applied twice to be safe, and that is how the two copies drift.

## How to verify a fix

Not by editing the active component and re-rendering — that is precisely the failing branch,
and it looks identical to success. Verify by which component's *distinctive string* appears:

```sh
curl -s https://relojistas.com/index.html | grep -o '<h4>[^<]*</h4>'
# today:  Quick Links / Our Services / Contact   <- footer-4-column (is_active=false)
# fixed:  Quick Links / Explore / …              <- footer-theme-chrome (is_active=true)
```

Remember `117`: a chrome change is invisible until `render_site_components` runs, so trigger a
`nav_drift` item for the site and re-render its pages before reading the result.

## What actually shipped, 2026-07-31 (candidate 1, in the shape corrections 2 and 3 forced)

**One predicate**, in `component_library.go`, built as a named string so a fourth
hand-typed copy cannot drift back in (the LOCK-007 construction, and a
source-scanning test enforces it):

```
is_active AND forked_from IS NULL AND component_level IN ('site','header','footer','head')
ORDER BY <that> DESC, name LIMIT 1
```

- `ResolveChromeComponent(ctx, db, function, logger) (*Component, eligible bool, error)`
  — returns the row **and** whether the library had anything legitimate. It does not
  error on an ineligible library because `head` has **no eligible component today**
  (both candidates `is_active=false`) and a head slot that goes unrendered also loses
  `injectBrandHeadTags`, i.e. its favicon and og-card. Ineligible logs at **ERROR** and
  is reported per slot as `ineligible_chrome` in the action result — a **library**
  defect, which no site can fix.
- `ChromeSlotFunction(slot)` — the slot→function map, previously inline in one caller
  and hard-coded in the other.
- Callers switched: `renderAndStoreSiteComponent` (the no-predicate one) and both of
  `link_site_components`' by-function fallbacks.
- `GetComponentByFunction` got `ORDER BY name` **and nothing else**. Measured first:
  functions with more than one eligible row fleet-wide = exactly 2 (`site-header`,
  `site-footer`), and the ordered query returns the same row the unordered one does
  for both — so the fleet's answer is unchanged and now guaranteed rather than
  incidental. **An `ORDER BY` added to an existing `LIMIT 1` is a behaviour change
  until you have measured that it is not.**

### Deliberately NOT done, and why — both are owner calls

1. **The 11 sites pinned to `footer-4-column` (and 7 to `header-bold-gradient`, 9 to
   `Document Head`) are not repointed.** Fleet-visible. **And the platform already
   detects them**: `deactivated_site_components` has raised `deactivated_component`
   items since **2026-07-17**, routed to `rerender-pages` — which re-renders **the
   component the row already points at**, i.e. the deactivated one. The routed repair
   is structurally incapable of repairing the finding, which is why two of them read
   `[unresolved after 2 attempts]`. **A `complete` deactivated_component item is not a
   repointed slot.** Own defect; see `bugs_open/166`.
2. **The five build-path callers of `GetComponentByFunction` (`RenderHeader`,
   `RenderFooter`, `RenderHead`) are not switched to the chrome resolver.** Doing so
   changes chrome markup on every page build fleet-wide (`site-header`/`site-footer` →
   the `*-theme-chrome` pair). So today the *assignment* path can no longer pick a
   section component as chrome while the *build* path still can. That asymmetry is
   deliberate and owner-facing. See `bugs_open/167`.

## Related

- `bugs_open/166` — the `deactivated_component` repair loop that cannot repair.
- `bugs_open/167` — the build path renders section-level components as site chrome.
- `bugs_open/111` — the ungated contact column. Its target component was wrong because of
  this; corrected there.
- `bugs_open/117` — stored chrome is never regenerated by a page re-render. These two compound:
  117 means a chrome change does not reach the page, 118 means it was applied to a component
  that would not have rendered anyway.
- `bugs_open/098` — archiving does not undeploy. Same retirement-flag-does-not-retire family.

Evidence and the sequence that found it:
`traffic_probe/relojistas_rebuild_running_notes.md`, 2026-07-27.

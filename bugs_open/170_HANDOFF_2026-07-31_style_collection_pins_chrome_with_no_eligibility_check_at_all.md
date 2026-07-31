# 170 — the style-collection chrome pin applies no eligibility predicate, and three deployed sites are pinned to a deactivated header

**Filed:** 2026-07-31 by the `bugfix_167_chrome_build_path` lane, while fixing
`bugs_open/167`. Found in the same three functions, on the branch **above** the one
167 fixes. Filed rather than fixed because fixing it changes served markup on live
sites — the one thing `bugs_open/118` established must not be smuggled into a
zero-visible-change fix.

**Severity:** medium. Nothing errors. Three deployed sites simply render a header
that the library says is switched off, and have done since it was switched off.

## The defect

`RenderHeader` and `RenderFooter` (`platform/orchestration/actions/component_library.go`)
try the site's style collection **first**, and only fall through to the by-function
lookup if it has no chrome pinned:

```go
if coll != nil && coll.HeaderComponentID != nil {
    comp, err = GetComponentByID(ctx, db, *coll.HeaderComponentID, logger)   // :1743
    ...
}
// only then the by-function branch (this is the half bugs_open/167 fixed)
```

`GetComponentByID` is, in full (`component_library.go:396-411`):

```sql
SELECT id, name, function, COALESCE(category,'') as category,
       html_template, input_schema, COALESCE(is_dark_section,false)
FROM content_components
WHERE id = $1
LIMIT 1
```

**No `is_active`. No `forked_from IS NULL`. No `component_level`.** Not a weaker
predicate than `ResolveChromeComponent` — *no* predicate. So whatever a style
collection points at is rendered as chrome, whatever state it is in.

## What that means live, 2026-07-31

```sql
SELECT s.domain, s.status, sc.name AS collection,
       h.name AS pinned_header, h.is_active, h.component_level,
       h.forked_from IS NOT NULL AS is_fork
FROM sites s
JOIN style_collections sc ON s.style_collection_id = sc.id
JOIN content_components h ON h.id = sc.header_component_id
ORDER BY s.domain;
```

| domain | status | collection | pinned header | `is_active` | level | fork |
|---|---|---|---|---|---|---|
| ai-agent-orchestration.com | deployed | professional-dark | `header-professional-dark` | **false** | site | no |
| finetuning.uk | deployed | professional-dark | `header-professional-dark` | **false** | site | no |
| gaswholesalers.com | deployed | professional-dark | `header-professional-dark` | **false** | site | no |
| leopardessconsulting.co.uk | deployed | leopardess-dark-gold | `header-leopardess` | true | site | **yes** |

Three deployed sites render a **deactivated** component as their header on every
page build. The fourth is correct and must stay: a site rendering its **own** fork
is what a fork is for, and it is the case any fix here has to preserve.

Footers are pinned the same way and are in the same state
(`professional-dark` → `footer-4-column`, `is_active=false`).

## How this relates to 118 and 167 — it is a THIRD thing

- **118** (closed): chrome *assignment* ignored `is_active`. Fixed; both assignment
  call sites now share one predicate, and the fleet was repointed.
- **167** (closed): the chrome *by-function build lookup* had no `component_level`
  filter, so a `component_level='section'` component could serve as chrome. Fixed.
- **170** (this): the chrome *pin* is dereferenced by id with **no predicate at
  all**, so it bypasses both fixes. All four pins are `component_level='site'`, so
  this is **not** 167's defect; it is **118's** class (a deactivated component
  serving as chrome) surviving on a **fourth** path that 118's own enumeration of
  "three places that ask this question" did not include.

118's three-call-site census was of code that *selects from a pool*. A pin is not
a selection, which is exactly why it was not counted — and why it kept the
behaviour the other three had fixed.

## Why it was not fixed with 167

Making the pin honour eligibility moves three deployed sites from
`header-professional-dark` (3,637 chars) to `header-theme-chrome` (2,551) —
different markup, different CSS, on live pages. `bugs_open/167`'s own filing note
says a fleet-visible change must not ride inside a fix measured to have no visible
effect, and 167's fix was measured to have none. So this needs its own before/after
and its own go.

## Fix candidates

1. **Validate the pin at render, fall back to the pool when it is ineligible.**
   Reuse `chromeEligibleSQL` — either a level-filtered `GetChromeComponentByID`, or
   check the returned component and call `ResolveChromeComponent` when it fails.
   Closes the door for every future pin. **Changes markup on 3 live sites.** Must
   keep the fork case working: `forked_from IS NULL` is right for pool *selection*
   and **wrong** for a pin, because pinning a site to its own fork is the intended
   use — so the pin predicate is `is_active AND component_level IN (…)`, *not* a
   copy of the pool predicate. That asymmetry is the whole subtlety here.
2. **Repair the data**: repoint `professional-dark` (and any other collection
   pinning an inactive component) at an active one. Cheapest, visible-change
   equivalent to candidate 1 for today's rows, and leaves the code able to serve a
   deactivated component the next time a component is switched off.
3. **Report without repairing**: have the render log/report an ineligible pin the
   way `render_site_components` reports `ineligible_chrome`. Cheap, no visible
   change — but the resolver's own header already notes that signal has **no
   automated reader** (`bugs_open/166`), so this adds a second unread signal.

Candidate 1 is the only one that makes the bad state unrepresentable. Candidate 2
is worth doing anyway and immediately, because it is data and needs no roll.

## How to check the current answer

The query above. Add `sc.footer_component_id` for the footer half. A collection
with `header_component_id IS NULL` is **not** affected — it takes the by-function
branch, which is the one 167 fixed.

## Verification standard for this filing

Per the owner ruling of 2026-07-31, a `bugs_open/` file asserting a cross-cutting
root cause goes through the `090` diagnosis loop **or** the filing session states
why it substituted equivalent first-hand verification. **Substituted, and here is
what was done instead:** the claim is not an inference about a mechanism I did not
read. `GetComponentByID` is fifteen lines and is quoted above **in full** from
source — the absent predicate is visible in the text, not deduced from a symptom.
The impact claim is a single live query over `sites`/`style_collections`/
`content_components`, printed above with its result, not a count carried from
another document. The two together are what a diagnosis run would have had to
produce. What is **not** claimed: that the deactivated header is *wrong* for those
three sites in a design sense — only that the code cannot tell, which is the
defect. Whether `header-professional-dark` should be reactivated instead of
repointed is an owner question, not a diagnosis one.

## Related

- `bugs_closed/118` — the assignment half, and the predicate this should reuse.
- `bugs_closed/167` — the by-function build half; its fix is the shape this one
  would follow.
- `bugs_open/166` — the repair that cannot repair; the reason candidate 3 is weak.
- `bugs_open/117` — stored chrome is never regenerated by a page re-render, so a
  fix here and the served page can disagree for days.

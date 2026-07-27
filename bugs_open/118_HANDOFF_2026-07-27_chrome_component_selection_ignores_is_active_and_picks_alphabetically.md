# 118 — chrome component selection ignores `is_active` and picks alphabetically, so every site's footer renders from a DEACTIVATED component

**Found:** 2026-07-27, on relojistas.com, while trying to work out why an owner-approved
edit to the active footer component changed nothing on any page.
**Severity:** medium-high. It is fleet-wide, it silently overrides every deactivation of a
chrome component, and it makes editing the *active* chrome component a no-op — so the
natural repair for any chrome defect fails in a way that looks like the repair was wrong.
**Status:** OPEN — diagnosed with evidence, not fixed. The structural fix changes the
rendered footer on every site, so it needs an owner call, not a site thread's.

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

## Related

- `bugs_open/111` — the ungated contact column. Its target component was wrong because of
  this; corrected there.
- `bugs_open/117` — stored chrome is never regenerated by a page re-render. These two compound:
  117 means a chrome change does not reach the page, 118 means it was applied to a component
  that would not have rendered anyway.
- `bugs_open/098` — archiving does not undeploy. Same retirement-flag-does-not-retire family.

Evidence and the sequence that found it:
`traffic_probe/relojistas_rebuild_running_notes.md`, 2026-07-27.

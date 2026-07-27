# 117 — site chrome is a stored artefact that no page re-render regenerates, so header/footer/nav changes are invisible on every existing page

**Found:** 2026-07-27, on relojistas.com, after an owner-approved change to the shared footer
component had no effect on 18 pages that all re-rendered successfully afterwards.
**Severity:** medium, and structural. Every chrome change — component template, nav tables,
site identity — is inert on existing pages until a *site-chrome* rebuild runs, and nothing on
the page path triggers one.
**Status:** OPEN — diagnosed with evidence. Mitigated on this site by queueing the rebuild
that does exist (`nav_drift` → `nav-updater`); the general gap is unfixed.

> **CORRECTED 2026-07-27, before anyone acted on it.** This case was first filed asserting
> the cause was `InjectFooter`'s early return (`component_library.go:1749`, "skip injection if
> page content already contains a site-footer"). **That is the wrong mechanism.** The
> single-page re-render path **never calls `InjectFooter` at all** —
> `grep -n 'InjectFooter\|InjectHeader' rerender_single_page_action.go` returns nothing. I
> reasoned from the first plausible function I found and did not check whether it was on the
> path. The real cause is below. What caught it: the stored footer contains `Our Services`,
> a string that exists in **no active component**, which the skip-guard theory could not
> explain.

## The mechanism

Site chrome is **pre-rendered once and stored**, then reused verbatim.

`assemblePage` (`rerender_single_page_action.go:352`) builds every page as
`head + header + sections + footer`, where three of those four come from a table:

```go
siteComponents, err := getSiteComponents(ctx, db, page.SiteID)
…
head   := resolveComponent(areaComponents, siteComponents, "head")
header := resolveComponent(areaComponents, siteComponents, "header")
footer := resolveComponent(areaComponents, siteComponents, "footer")
```

and `getSiteComponents` is:

```sql
SELECT slot_name, rendered_html
FROM site_components
WHERE site_id = $1 AND rendered_html IS NOT NULL AND rendered_html != ''
```

So the footer served on every page is a **frozen string in `site_components.rendered_html`**.
It is not rendered from `content_components.html_template` at page-render time, and it does
not consult `site_nav_items`. Only `render_site_components` rewrites it.

## Evidence

relojistas' stored chrome, 2026-07-27:

```sql
SELECT slot_name, length(rendered_html), updated_at,
       rendered_html ILIKE '%contacto.html%'  AS has_contacto,
       rendered_html ILIKE '%Our Services%'   AS has_our_services
FROM site_components WHERE site_id='ecf15e75-…';

-- footer | 3991 | 2026-07-16 13:52 | t | t
-- head   | 8635 | 2026-07-16 13:52 | f | f
-- header | 3707 | 2026-07-16 13:52 | f | f
```

**All three frozen since 16 July.** Two independent consequences, both live today:

1. **It renders `<h4>Our Services</h4>`, which exists in no active component.** That string
   is in `footer-4-column` — `is_active = false`. The active chrome component
   (`footer-theme-chrome`) says `<h4>Explore</h4>`. The live site has been serving a
   deactivated component's output for eleven days.
2. **It links `/contacto.html` three times on all 18 pages**, and that URL now **404s** — the
   page was archived and its file deleted on 2026-07-27 (owner ruling), and its
   `site_nav_items` row was set `inactive` at 14:36:27. Neither reached the page.

Measured, so the "maybe it just hasn't re-rendered" explanation is closed:

```sql
SELECT count(*) FILTER (WHERE updated_at > '2026-07-27 14:34:00+00') AS after_nav_change,
       count(*) FILTER (WHERE updated_at > '2026-07-27 16:54:50+00') AS after_template_change,
       count(*) AS total
FROM pages WHERE site_id='ecf15e75-…' AND status='active';
--  18 | 5 | 18
```

18 of 18 pages re-rendered after the nav change and 5 after the template change. Neither
change is in the served HTML, because neither is on the path.

## Why it bites harder than it looks

- **A chrome fix cannot be verified by re-rendering a page.** The natural check — change the
  component, re-render, curl — returns the old output and reads as "my edit was wrong". This
  cost two correct changes on this site before the cause was found.
- **`bugs_open/111` is applied and provably inert.** Its gate is live in the template and has
  never reached a page.
- **The nav half of `bugs_open/098` cannot be repaired by re-rendering either** — fixing the
  stranded `site_nav_items` row is necessary but not sufficient, which that case does not say.
- **Deactivating a chrome component does not retire it.** `is_active = false` stops it being
  *selected*; it does not stop its already-rendered output being served indefinitely. Same
  family as 098's "archiving does not undeploy", one layer up.
- Plausible cause of the residual stale-chrome population handed over by `bugs_closed/049`.
  **[UNVERIFIED]** — I have not checked those sites against this mechanism.

## What does regenerate it

`render_site_components` **is** registered (`registry.go:843`) and is run by
`rerender-site`, `nav-updater`, `site-work-orchestrator`, `pageflow-builder`,
`rerender-pages`, `nav-link-fixer`. The reachable manual route is a `nav_drift` work item
handled by `nav-updater` (3 completed fleet-wide, most recent 2026-07-26) — which is what
was queued for relojistas.

**So this is a coupling gap, not a missing capability**: the rebuild exists and works, and
nothing causes it to run when the thing it renders from changes.

**[UNVERIFIED] whether `RerenderSitePagesAction` matters here.** It is defined
(`rerender_pages_actions.go:32`), strips and re-injects chrome, and is registered nowhere —
`grep -rn 'RerenderSitePagesAction' --include=*.go .` returns only its own file. Noted
because it is dormant machinery in the same area, not because it is implicated.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Invalidate on write.** Whatever writes `site_nav_items`, a chrome `content_components`
   row, or site identity should mark that site's `site_components` stale (or enqueue the
   rebuild). Removes "silently frozen chrome" as a state rather than making it repairable.
   Needs a survey of every writer to those three surfaces.
2. **Stamp provenance and detect drift.** Record which `content_components` version each
   `site_components` row was rendered from, then a cheap sweep reports any site whose stored
   chrome predates its component or nav rows. Would have surfaced this in one query, and
   catches the `is_active=false`-still-serving case for free:
   ```sql
   SELECT s.domain, sc.slot_name, sc.updated_at
   FROM site_components sc JOIN sites s ON s.id = sc.site_id
   WHERE sc.updated_at < (SELECT max(updated_at) FROM content_components WHERE name = 'footer-theme-chrome');
   ```
   **[UNMEASURED] fleet-wide** — run it before designing anything; relojistas is one site and
   this may be near-universal.
3. **Render chrome at assemble time** instead of storing it. Cleanest conceptually, worst
   blast radius: it moves per-page cost and would need the nav/identity resolution that
   `render_site_components` currently does once per site.

## How to verify a fix

Do **not** verify by re-rendering a page and reading it — that is the failing branch and it
looks identical to success. Verify against a word only the current template can produce:

```sh
curl -s https://relojistas.com/index.html | grep -o '<h4>[^<]*</h4>'
# today:  Quick Links / Our Services / Contact     <- footer-4-column, is_active=false
# fixed:  Quick Links / Explore                    <- footer-theme-chrome, and NO Contact
#                                                     block, since 111's gate is already live
#                                                     and relojistas has no email or phone
```

The disappearing `Contact` block is the discriminating half: it is gated by a change that is
already in the template and has never reached a page.

## Related

- `bugs_open/111` — the ungated contact column; applied, and inert because of this.
- `bugs_open/098` — archiving strands the nav row; that repair is also blocked here.
- `bugs_closed/049` — stale chrome fleet-wide; possible shared cause, unverified.

Evidence and the wrong turn: `traffic_probe/relojistas_rebuild_running_notes.md`, 2026-07-27.

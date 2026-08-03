# 191 — the header CTA is validated by a LOOSER predicate than the nav rendered beside it, so chrome ships a 404 button the nav itself would have refused

**Filed 2026-08-03 by the `mortgagecalculator_couk_adoption` lane.** OPEN, UNOWNED.
**LIVE on one site, confirmed at the wire.**
**Severity:** medium. Low blast radius *today* (1 site), but it fires on exactly the
condition every adoption passes through — a site with pages planned and not yet built —
and chrome ships on **every page**, so one bad button scales with the next batch.
**Class:** two call sites answer one question differently (same family as `118`, which
fixed the *component-eligibility* spelling of this and did not touch the *link-target* one).

## The defect

`RenderSiteComponentsAction` builds the header's nav items and the header's CTA button in
the same run, from the same `pages` table, and validates them with **two different
eligibility predicates**:

| element | helper | predicate |
|---|---|---|
| nav items | `loadFetchablePageSet` — `nav_tables.go:258` | `status NOT IN ('deleted','archived')` **AND NOT (`datahelpers.NeverDeployedPagePredicate`)** |
| header CTA | `loadResolverPageSet` — `resolve_internal_links_action.go:486` | `status NOT IN ('deleted','archived')` — **no deployment check** |

The CTA fallback (`render_site_components_action.go:172-187`) runs when there is no
`contact` nav item — which, note, is itself *more likely* on a young site, because the
contact item was already dropped by the stricter nav filter. It then asks
`chooseCTATargets` for an "interactive page" and validates that pick against the **loose**
set:

```go
headerPages := loadResolverPageSet(ctx, params, siteID, params.Logger)   // :171
if ctaURL == "" || !headerPages.Contains(ctaURL) {                        // :172
    …
    primary, _ := chooseCTATargets("", "", interactive, hubs)             // :181
    if primary.URL != "" && headerPages.Contains(primary.URL) {           // :182  ← loose set again
        ctaURL = primary.URL
```

So a page that has never been deployed passes, and its URL is written into
`site_components.header`. The file's own comment at `:149-150` states the opposite
intent — *"so it points at an existing page instead of the hardcoded phantom
/contact.html"* — and `:166-170` says the gated template "renders no CTA button rather
than a phantom". Both are true of the *contact* path and false of the *fallback* path.

## Evidence — the two elements inside ONE rendered header

`mortgagecalculator.co.uk`, `site_components.slot_name='header'`, rendered 2026-08-03 11:01 UTC:

```
<li><a href="/index.html"                                    ← nav: filtered to the 1 deployed page
href="/tools/stamp-duty/index.html" class="header-cta"       ← CTA: never deployed
```

```
$ curl -s -o /dev/null -w '%{http_code}' https://mortgagecalculator.co.uk/tools/stamp-duty/index.html
404
```

`tool-stamp-duty` is `build_status='planned'`, `deployed_at IS NULL`. The nav in the same
component was reduced to a single link precisely *because* the other targets are unbuilt;
the button beside it was not.

## Blast radius — MEASURED, and the DB over-reports it

Extract the CTA from `rendered_html` (it is **not** in `content_data` — see the trap below)
and join to `pages`:

```sql
SELECT s.domain,
       substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"') AS cta_href,
       p.build_status, p.deployed_at IS NOT NULL AS ever_deployed
  FROM site_components sc
  JOIN sites s ON s.id = sc.site_id
  LEFT JOIN pages p ON p.site_id = sc.site_id
       AND p.url = substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
 WHERE sc.slot_name = 'header' AND p.deployed_at IS NULL;
```

→ 2 rows: `lendzy.co.uk` → `/tools/price-cap-checker/index.html`, and ours.

> **The DB answer is wrong by one, and the check that catches it is one `curl`.**
> `lendzy.co.uk/tools/price-cap-checker/index.html` returns **HTTP 200**. `deployed_at IS
> NULL` means "no recorded deploy", **not** "does not serve" — the same conflation as
> `bugs_open/098`'s durable half, in the opposite direction. **12 of 14 sites point their
> CTA at a deployed page; 1 is a false positive; the confirmed live 404 is 1 — ours.**
> Do not repeat the DB-only measurement and report 2.

## Diagnosis-loop verdict

Run `1cdd9556-637d-4dd8-8bc6-c1db3f4196cd` (seeded, 2026-08-03): **UNVERIFIABLE, stopped by
iteration-cap.** Not a refutation — and its trail is worth reading, because it
**independently reproduced the static mechanism**, citing the same three lines quoted
above without being shown them.

What it could not close was *occurrence*, and it diagnosed its own failure correctly:

> "cta_url is not a stored column anywhere in the schema shown (it lives inside
> `site_components.rendered_html` as literal HTML) — that query could not have found a real
> row even if the mismatch exists, so its 0-row result is not evidence of absence."

Its `DataRequests` queried `sc.content_data->>'cta_url'`. **There is no such key.** The
occurrence evidence above is that query re-aimed at `rendered_html`, plus the wire check.
An earlier unseeded run of the same symptom failed outright at `assemble_bundle`
(`no scope`); pass `SEED_SCOPE` — see the correction on `bugs_closed/174`.

## Fix candidates, ordered by what closes the door

1. **Give the CTA the nav's predicate.** Validate `primary.URL` against
   `loadFetchablePageSet` (the set `applyNavVisibility` already built) instead of
   `loadResolverPageSet`. Smallest diff; makes the two elements in one component agree by
   construction. **Caveat that must be handled:** `applyNavVisibility` deliberately
   *disables* the filter when a site has **zero** deployed pages (`nav_tables.go:194-207`,
   so a first build does not freeze an empty nav into chrome). The CTA needs the same
   escape or a first build renders no button — which is the *correct* outcome here, since
   the gated template already omits it on empty `cta_url`.
2. **One named predicate, one resolver** — the shape `118` applied to component
   eligibility (`b052249d8`), applied to link targets. Closes the door for the next
   caller, not just this one. Costlier; the right structural answer.
3. Detect-only: extend the phantom-link checks to `site_components.rendered_html`.
   **Weakest** — `bugs_open/071` is the standing record that detection here is already
   discarded, so this adds a finding nobody consumes.

**Do NOT "fix" it by building the target page.** That clears our symptom and leaves the
predicate wrong for the next site.

## How to verify a fix

```sql
-- must return 0 rows AFTER a nav-updater run on a site with unbuilt pages
SELECT s.domain, substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
  FROM site_components sc JOIN sites s ON s.id=sc.site_id
  LEFT JOIN pages p ON p.site_id=sc.site_id
       AND p.url = substring(sc.rendered_html from 'href="([^"]+)"[^>]*class="header-cta"')
 WHERE sc.slot_name='header' AND p.deployed_at IS NULL;
```

Then **curl each surviving CTA href** — the query alone over-reports (see above).
`mortgagecalculator.co.uk` is the live reproduction: it is locked, has 25 unbuilt pages and
1 deployed one, so re-running `nav-updater` there re-creates the condition on demand.

## Ownership note

`render_site_components_action.go` is actively worked by the chrome lane
(`117`/`118`/`166`/`167`/`170`). This bug is **distinct** from `118` — that one fixed
"which component template serves this slot", this one is "which page may a chrome link
name" — but the fix lands in the same file, so **contribute into that lane rather than
starting a competing change.** Filed here, not routed at them, per `scripts/who-owns.py`.

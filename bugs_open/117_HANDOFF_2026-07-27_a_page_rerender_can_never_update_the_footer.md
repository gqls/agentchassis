# 117 — a page re-render can never update the footer, because injection is skipped whenever a footer already exists

**Found:** 2026-07-27, on relojistas.com, after an owner-approved change to the shared
footer component had no effect on 18 pages that all re-rendered successfully afterwards.
**Severity:** medium, and structural. **No reachable code path refreshes site chrome on
an already-rendered page.** Every fleet-wide footer change is therefore inert for existing
pages and lands only on pages built from scratch.
**Class:** dormant machinery — the code that *would* do it exists and is never registered.

## The one line

`InjectFooter` (`platform/orchestration/actions/component_library.go:1749`) begins:

```go
// Skip injection if page content already contains a site-footer.
if strings.Contains(html, `class="site-footer`) {
    logger.Info("InjectFooter: page already contains site-footer, skipping injection", …)
    return html
}
```

The assembled HTML on a re-render already contains the previous footer, so the guard is
true on every page that has ever been rendered. The footer is returned untouched, and the
freshly-computed `footerNav` on the lines immediately below — including a `GetNavItems`
call made specifically to refresh it — is discarded unused.

## Proof that needs no instrumentation

The live component template and the live page disagree, and the disagreement is a word:

```
$ psql -c "SELECT html_template FROM content_components WHERE name='footer-theme-chrome'" | grep -o '<h4>[^<]*</h4>'
<h4>Quick Links</h4>
<h4>Explore</h4>            <- current template
<h4>Contact</h4>

$ curl -s https://relojistas.com/index.html | grep -o '<h4>[^<]*</h4>'
<h4>Quick Links</h4>
<h4>Our Services</h4>       <- live page, an OLDER version of the same component
<h4>Contact</h4>
```

`Explore` replaced `Our Services` in the template at some earlier date. Every relojistas
page still serves `Our Services`. The footer is frozen at whatever the template said the
last time each page was *built*.

## How it was found, which is the useful part

Two independent, correctly-applied changes both failed to reach the page:

1. **A stranded nav row.** Retiring the `contacto` page left its `site_nav_items` row
   `active` (that half is recorded in `bugs_open/098`). Setting it `inactive` at 14:36:27
   should have dropped three `href="/contacto.html"` links from every page's footer.
2. **An owner-approved gate** on the footer's contact column
   (`{{if or .email .phone}}`, `bugs_open/111`) applied at 16:54:50, so a site with no
   contact details stops rendering an empty `<h4>Contact</h4>`.

Then, measured:

```sql
SELECT count(*) FILTER (WHERE updated_at > '2026-07-27 14:34:00+00') AS after_nav,
       count(*) FILTER (WHERE updated_at > '2026-07-27 16:54:50+00') AS after_gate,
       count(*) AS total
FROM pages WHERE site_id='ecf15e75-…' AND status='active';
--  after_nav | after_gate | total
--         18 |          5 |    18
```

**18 of 18 pages re-rendered after the nav change; 5 after the gate. Neither change is
present in the served HTML.** Both were correct; neither could arrive.

## Why there is no workaround inside the platform today

The code that handles this correctly exists. `RerenderSitePagesAction`
(`rerender_pages_actions.go:32`) **strips the old footer before re-injecting**:

```go
// NEW: Remove injected site-level footers (with associated style blocks)
html = regexp.MustCompile(
    `(?is)(?:<!--\s*FOOTER\s+SOURCE:[^>]*-->\s*)*`+
        `<footer\s[^>]*class="site-footer[^>]*>.*?</footer>`+
        `\s*(?:<style[^>]*>.*?</style>\s*)*`,
).ReplaceAllString(html, "")
```

**It is never registered and no workflow calls it.** Verified:

```
$ grep -rn 'RerenderSitePagesAction' --include=*.go . | grep -v rerender_pages_actions.go
(no output)
$ grep -n 'rerender' platform/orchestration/actions/registry.go
get_pages_for_rerender · rerender_single_page · rerender_page_sections · create_rerender_items
```

`rerender_single_page` is the one that skips. So the fleet has a correct implementation and
an incorrect one, and only the incorrect one is reachable.

## Consequences worth stating plainly

- **Every footer/chrome change is inert on existing pages.** Anyone editing
  `footer-theme-chrome` (or the identity data it renders) and verifying by re-rendering a
  page will observe no change and may reasonably conclude their edit was wrong.
- **`bugs_open/098`'s nav half cannot be repaired by re-rendering** — fixing the nav row is
  necessary but not sufficient, which is not obvious from that case.
- It is a plausible cause of the residual "stale chrome" population handed over by
  `bugs_closed/049`. **[UNVERIFIED]** — I have not checked those pages against this
  mechanism.

## Fix candidates, ordered by what closes the door

1. **Strip-then-inject in `InjectFooter` itself**, using the regex the unregistered bulk
   action already contains. Makes "footer is stale" unrepresentable rather than merely
   repairable, and puts one implementation where both paths already call. Same treatment
   for `InjectHeader` — **[UNVERIFIED] whether it carries the identical guard**; check
   before assuming the footer is the only frozen half.
2. **Register `RerenderSitePagesAction`** and give it a trigger. Fixes the capability gap
   but leaves the broken single-page path in place to be picked next time.
3. **Delete the skip-guard.** Smallest diff, but the guard presumably exists to stop
   double-injection when section content legitimately carries a footer; removing it blind
   risks duplicate footers. Establish why it was added first.

Whichever is taken, **survey first**: any page whose *section content* legitimately contains
`class="site-footer"` would change behaviour under 1 or 3.

## How to verify a fix

Do not verify by re-rendering and reading the page — that is the failing branch, and it
looks identical to success. Verify against the word:

```sh
curl -s https://relojistas.com/index.html | grep -o '<h4>[^<]*</h4>'
# today: Quick Links / Our Services / Contact
# fixed: Quick Links / Explore   / (no Contact block — relojistas has no contact details)
```

The `Contact` block disappearing is the discriminating half: it is gated by `111`'s change,
which is already live in the template and has never reached a page.

## Related

- `bugs_open/111` — the ungated contact column. Its fix is applied and provably inert
  because of this case.
- `bugs_open/098` — archiving strands the nav row. Same site, same day; that repair is also
  blocked here.
- `bugs_closed/049` — stale chrome fleet-wide. Possible shared cause, unverified.
- `bugs_open/071` — the phantom-CTA class on the same site, also rooted in
  `component_library.go` defaults.

Evidence and the full sequence: `traffic_probe/relojistas_rebuild_running_notes.md`,
2026-07-27.

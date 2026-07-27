# 111 — the footer's "Contact" heading is ungated while its contents are gated, so 8 of 14 live sites show a heading over nothing

**Filed:** 2026-07-27, from relojistas.com, while carrying out an owner ruling that
the site should have no contact route at all.
**Severity:** low individually (cosmetic), but it is **fleet-wide and on every page
of every affected site**, and it is the reason a site cannot be given "no contact
surface" by configuration alone.
**Status:** OPEN — diagnosed, not fixed. The fix is one conditional but it edits
shared fleet chrome, so it is not a site thread's call to make unilaterally.

## Symptom

The footer contact column renders its heading unconditionally and each of its
contents conditionally. A site with neither email nor phone therefore gets a
bare English heading over empty space.

Live on relojistas.com, all 19 pages, 2026-07-27:

```html
<div class="footer-contact">
    <h4>Contact</h4>
    <p><a href="/cdn-cgi/l/email-protection#88"></a></p>
    <p></p>
</div>
```

(The empty anchor is Cloudflare's rewrite of an empty `mailto:` — see the
check-methodology note at the end.)

## Root cause: the gate is on the wrong side of the heading

Two independent implementations, same asymmetry.

**1. `footer-theme-chrome`** (DB component, the one actually serving most sites):

```
<div class="footer-contact">
    <h4>Contact</h4>                                    <- unconditional
    {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}
    {{if .phone}}<p>{{.phone}}</p>{{end}}
</div>
```

Contrast the two columns above it, which are gated correctly and disappear
whole when empty:

```
{{if .quick_links_html}}<div class="footer-links">…{{end}}
{{if .services_html}}<div class="footer-services">…{{end}}
```

So the correct pattern is already present in the same template, three lines up.

**2. `RenderFallbackFooter`** (`platform/orchestration/actions/component_library.go:1673`)
— the Go fallback, same shape:

```go
<div class="footer-contact"><h4>Contact</h4><p>%s</p></div>
```

## Measurement

`site_specs.identity.contact` across sites with at least one deployed page,
2026-07-27:

| contact details | sites |
|---|---|
| neither email nor phone → **renders an empty block** | **8** |
| has email (and usually phone) | 6 |

The 8: `gamesdesign.co.uk`, `gaswholesalers.com`, `oufe.com`, `relojistas.com`,
`robot-hands.com`, `vetcomparison.uk`, `vonc.com`, `webdesign.co.uk`.

```sql
SELECT s.domain,
       COALESCE(ss.data->'contact'->>'email','(none)') AS email,
       COALESCE(ss.data->'contact'->>'phone','(none)') AS phone
FROM sites s
LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='identity' AND ss.is_current
WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.deployed_at IS NOT NULL)
ORDER BY 2, s.domain;
```

> ## CORRECTED 2026-07-27 (triage sweep) — the blast radius is **2 sites, not 8**, and the query above cannot measure it
>
> **What caught it:** curling the rendered footers instead of querying keys —
> CLAUDE.md's *"Trust the rendered artefact, not the status"*. The query above reads
> `site_specs.identity.contact`, and **the footer does not read that.**
>
> Measured 2026-07-27 by fetching `/` on all 14 live-page sites and stripping tags
> from the `.footer-contact` block:
>
> | rendered outcome | sites |
> |---|---|
> | **EMPTY — heading over nothing** | **2**: `gamesdesign.co.uk`, `relojistas.com` |
> | populated (email and/or phone) | 10 |
> | no `.footer-contact` block at all | 2: `idea.uk`, `webdesign.co.uk` |
>
> So **five of the eight sites this file names as broken render a populated Contact
> block**: `gaswholesalers.com`, `oufe.com`, `robot-hands.com`, `vetcomparison.uk`,
> `vonc.com`. The file's list was derived from the wrong key.
>
> **The footer's values come from the `sites.email` / `sites.phone` COLUMNS**, not
> from `site_specs.identity` in either its flat or nested form:
>
> ```sql
> SELECT domain, email, phone FROM sites
> WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id=sites.id AND p.deployed_at IS NOT NULL);
> -- sites.email IS NULL on exactly ONE site: gamesdesign.co.uk
> -- sites.phone IS NULL on seven: gamesdesign, oufe, relojistas, robot-hands,
> --                               vetcomparison, vonc, webdesign
> ```
>
> That column set predicts the rendered output on 13 of 14 sites. The exception is
> **`relojistas.com`, which renders an empty anchor despite a populated
> `sites.email`** (`relojistas@contactforsales.com`) — its identity spec was cleared
> at 14:24 and the page re-rendered at 14:29, so a third path is feeding that render.
> **[UNVERIFIED] which one** — worth one query before anyone fixes relojistas
> specifically.
>
> **What this does to the case.** It does not refute the defect — the gate really is
> on the wrong side of the heading, and `gamesdesign.co.uk` is a genuine live
> instance with no contact data anywhere. It refutes the *sizing*: this is a 2-site
> cosmetic issue, not an 8-site fleet-wide one, and one of those two
> (`relojistas.com`) is an owner ruling of *no contact route at all* rather than
> missing data. **Fix candidate 1 is still right and still cheap; it is just not
> urgent, and it should not be justified by the "8 of 14" figure.**
>
> **Dependency (from the main session's 072 work, 2026-07-27, commit `ca53bc19c`):**
> `bugs_open/072` (*contact_info_reads_flat_identity_keys* — resolve by slug, the
> `072` in `bugs_closed/` is the component-CSS one) is the contract defect behind the
> identity-key divergence, and the sequence is **072 then 111**, because fixing 111
> first hides whether 072 is still broken. Note the correction above narrows what 072
> can buy here: the footer is *already* populated on 10 of the 12 sites that carry the
> block, so at most 2 sites are reachable by any contact-data fix, and only
> `gamesdesign.co.uk` by data alone.

## Why it blocks something real, not just tidiness

The owner ruled on 2026-07-27 that relojistas.com carries **no contact route**. Every
other surface was removable as site data — the contacto page archived and its file
deleted, the hero CTAs repointed, the standing request for a business email closed,
the inherited phone number cleared. **The footer heading is the one that cannot be
removed by configuring the site**, because it is not configuration: it is a literal in
shared chrome. There is no per-site footer seam either — `resolved_composition` selects
layout, palette, typography and theme, and nothing else.

So today the platform cannot express "this site has no contact surface", which is a
legitimate and probably common state for a portfolio/for-sale domain.

## Fix candidates, ordered by what closes the door

1. **Gate the container on its contents** — makes the bad state unrepresentable, and
   simply copies the pattern the same template already uses for its other two columns:
   ```
   {{if or .email .phone}}<div class="footer-contact"><h4>Contact</h4>…</div>{{end}}
   ```
   Blast radius: removes an empty heading from 8 sites; changes nothing on the 6 that
   have details, because the guard is true there. **No site loses content.** DB config,
   so live immediately and instantly reversible. Same one-line change needed in
   `RenderFallbackFooter` (Go, needs an image roll) so the two paths do not diverge.
2. **A `contact_visible` flag on the commercial/identity spec** — more expressive, but
   it is an operator-must-remember switch guarding a state the data already implies.
   Worse than 1 for that reason.
3. **Leave it and strip per-site after render** — HTML patching; rejected by doc `003`
   and wiped by the next scoped re-render. Named only to rule it out.

## Landmine for whoever edits `footer-theme-chrome`

`bugs_closed/072`: **a `<style>` block at the END of an `html_template` is silently
deleted** — the extraction regex only takes a style block *ahead of* the script.
`footer-theme-chrome` **has a trailing `<style>`**. Preserve it exactly, and verify the
rendered footer still carries `.footer-legal`/`.footer-brand p` rules after any edit.

## Related, and deliberately kept separate

- The same footer is **hardcoded English on every site whatever its language** —
  `Quick Links`, `Explore`, `Contact`, `All rights reserved` — which on relojistas
  (a wholly Spanish site) is four more English strings. That is the same family as the
  `bugs_open/071` sighting of 2026-07-27 (`component_library.go`'s English defaults
  producing a dead `/contact.html` CTA on a Spanish site) and the `about-commercial-block`
  language gap raised with `about_page_commercial`. **There is no per-site language seam
  anywhere** — no `language` column on `sites`; it appears only inside
  `deploy_config.rss_feed.language`. Fixing the gate does not fix the language, and a
  language seam is a bigger design question than this case.
- `bugs_open/063` — `validateEmails` fails open when a site has no contact email. Same
  underlying "no contact details" state, different consequence.

## Check-methodology note, paid for here

`grep -c 'mailto:'` returns **0 on any Cloudflare-proxied site regardless of the truth**,
because CF rewrites every mailto into `/cdn-cgi/l/email-protection#<hex>`. This is what
made the empty footer anchor invisible for a day. Decode by XOR-ing each byte after the
first against the first byte.

## Unverified observation worth someone's minute — **RESOLVED 2026-07-27, benign at the footer**

> **CHECKED (triage sweep, 2026-07-27).** `dartsonline.com`'s live footer publishes
> `darts@contactforsales.com` / `07934 524 911` — the house convention, taken from
> `sites.email` / `sites.phone`. The research-derived `sales@darts.com` /
> `(800) 526-1920` in its identity spec **are not what ships to the footer**, because
> the footer never reads that spec (see the correction above). So the worry below is
> not live *at this surface*.
> **[UNVERIFIED] whether those spec values surface anywhere else** — the spec is still
> carrying a third party's real address and phone number (`13010 NE David Cir,
> Portland`), which is worth someone's minute for a different reason than the one
> originally filed.

`dartsonline.com` carries `sales@darts.com` / `(800) 526-1920` in its identity contact.
Every other site with details uses the `<name>@contactforsales.com` house convention, and
`darts.com` is a different domain. relojistas had exactly this shape — a real Spanish
phone number inherited from research into the defunct forum, cleared on 2026-07-27 before
any chrome could surface it. **[UNVERIFIED] whether dartsonline's details are the owner's
or were carried in from research**; if the latter, they are being published on a live site.

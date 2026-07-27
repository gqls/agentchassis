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

## Unverified observation worth someone's minute

`dartsonline.com` carries `sales@darts.com` / `(800) 526-1920` in its identity contact.
Every other site with details uses the `<name>@contactforsales.com` house convention, and
`darts.com` is a different domain. relojistas had exactly this shape — a real Spanish
phone number inherited from research into the defunct forum, cleared on 2026-07-27 before
any chrome could surface it. **[UNVERIFIED] whether dartsonline's details are the owner's
or were carried in from research**; if the latter, they are being published on a live site.

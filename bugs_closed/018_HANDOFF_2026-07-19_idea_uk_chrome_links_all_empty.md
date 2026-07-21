# 018 — idea.uk: the site chrome renders with EVERY link empty (`href=""`) — the site is unnavigable

**Found:** 2026-07-19, by the first discovery run ever pointed at idea.uk (see `017` for why it had
never run). **Status:** open, unstarted. **Severity:** high — the site is live, public, and cannot be
navigated. **Scope:** idea.uk confirmed; the *class* is very likely fleet-wide (see "Is this only
idea.uk?").

> **UPDATE 2026-07-21 — ✅ CLOSED: both halves FIXED & LIVE & VERIFIED.**
> - **Site half (idea.uk itself): FIXED & LIVE.** The DB-only template rewrite (`d63e62aad`,
>   `sql/p3_01`/`p3_02`) rewrote site-header/site-footer against the renderer's real vocabulary and
>   gated every anchor. `curl https://idea.uk/` returns **0** `href=""`, a populated logo `src`, and
>   nav/CTA/footer links to real `.html` pages. (The two remaining `href="#"` are the
>   `brief-explanation` "Get Started"/"Learn More" — a separate `dead_control` finding, not this bug.)
> - **Platform half (the fleet-wide mechanism): FIXED & LIVE on v1.0.1146.** `36829b07b` makes
>   `render_site_components` read `input_schema` and gap-fill via the shared `sourceResolver`, and
>   makes `RenderTemplate` name the fields it blanks (Error when a blanked placeholder sits inside
>   `href=`/`src=`). **Binary pod-verified** by discriminating grep: new `RenderTemplateReportingMissing`/
>   `missingBareFields` present, the deleted `Cleaning up <no value>` log absent. Same commit fixed
>   `bugs_open/041` (chrome JS never published) — verified end-to-end (`site-header.js`/`site-footer.js`
>   404→200, real mobile-menu JS, page references it; see 041's resolution block).
> - **Owner ruling 2026-07-21:** ship the **observability** layer (this fix); the **block/escalate**
>   half is the follow-on `bugs_open/054` (OPEN, tracked separately). Council: round 2 REVISE on the
>   scope point the owner overruled — **no `Council-Reviewed` trailer** (trailer discipline).
> Closing this case (both halves live). The *class* of "chrome renderer second-class vs page renderer"
> continues in `054` (escalation) — this case is the observability fix and it is done.

## Symptom — measured, not estimated

Every `<a>` on the deployed homepage, counted:

```
  15  <a href="">                          (bare, several with no link text either)
   8  <a href="" class="nav-link">         ← the ENTIRE primary navigation
   2  <a href="" class="btn-primary">      ← page CTAs
   2  <a href="" class="btn-secondary">
   3  <a href="" aria-label="Twitter|LinkedIn|Facebook">
   1  <a href="#" class="brief-explanation__cta-primary">Get Started
   1  <a href="#" class="brief-explanation__cta-secondary">Learn More
   1  <a href="/" class="header-logo">     ← works
   1  <a href="/" class="footer-logo">     ← works
```

**31 of 33 links are dead.** Plus `<img class="header-logo-img" src="" alt="idea.uk logo">` — the
logo image URL is empty although the asset exists and serves 200 at
`/assets/images/logo.jpg`, so the browser shows the alt text "idea.uk logo". That is what the owner
saw and reported as *"the logo isn't there … a lot of the buttons and text don't do anything."*

Several footer links have **empty text as well as empty href** — `<a href=""></a>` — so they are
invisible rather than merely broken.

## What the auditors found (and one correction)

The `completeness-discovery-agent` run produced 30 findings. The relevant ones:

| item_type | n | note |
|---|---|---|
| `phantom_internal_link` | 9 | `/about`, `/report`, `/tools`, `/how-it-works`, `/method`, `/tools#privacy` — no matching page |
| `cta_names_unknown_destination` | 8 | "How we work →", "Read the method →" etc. |
| `dead_control` | 4 | "Get Started" / "Learn More" → `#`, on index + tools (`brief-explanation`) |
| `empty_internal_href` | 3 | **site_component (header)**, **site_component (footer)**, report:info-card-grid |
| `page_rerender` (misdirected CTA) | 3 | index, about, report |
| `required_fields_missing` | 1 | `info-card-grid` on report missing `section_eyebrow` |
| `empty_section` | 1 | `news-listing` on news-index |
| `needs_rerender` (`missing_structure`) | 1 | claims 9 pages deployed without header/footer |

**Correction on that last one — do not act on it as written.** Its reason is *"Pages deployed without
header/footer"*, and that is **false for the deployed artefact**: every page carries full chrome. It
is easy to mis-verify — the footer is emitted as `<section class="footer-…">`, **not** `<footer>`, so
a `grep '<footer'` returns 0 and appears to confirm the finding. Check for `class="footer-` and the
`site-footer.js` reference instead. So either `missing_structure` inspects DB linkage that has
drifted from the artefact, or it is a false positive. Either way the *remedy* it proposes
(`refresh_site_components: true`) is plausibly right for the wrong stated reason — which is exactly
the situation where someone re-runs a rerender, sees the chrome "return", and records a fix that
fixed nothing.

The two `empty_internal_href` findings against **site_component (header)** and **(footer)** are the
accurate diagnosis: this is a *chrome* defect, not a per-page one. That is why it is uniform across
all nine pages.

## Root cause — ESTABLISHED 2026-07-19

> **The chrome renderer fills templates from a hardcoded key vocabulary and never reads the
> component's `input_schema`. idea.uk's two chrome components declare a different vocabulary, so
> every field they name resolves to the empty string; their templates are ungated, so each empty
> value becomes a visible `href=""` rather than a suppressed element.**

Evidence, in order:

- **Not fleet-wide.** The query below, run across all 11 sites: idea.uk is the only domain with any
  empty-href chrome component (2 of its 3). Every other site: 0.
- **idea.uk is alone on its components.** `site-header` (`f420f3fa…`) and `site-footer`
  (`4238e467…`), both created 2026-05-06, `is_active=t`, each used by **exactly one** site. The
  fleet's nine other sites sit on `header-bold-gradient` / `footer-4-column` (`is_active=f`, used
  by 5 and 9).
- **The renderer's vocabulary is hardcoded.** `render_site_components_action.go:222-262` builds
  `RenderContext.ContentData` as a literal map — `company_name`, `logo_url`, `logo_text`,
  `nav_items_html`, `quick_links_html`, `cta_url`, `cta_text`, `legal_links`, … — and
  `:530` renders the template against it. **`input_schema` appears nowhere in the file**, and
  `site_components.content_data` (`{}` fleet-wide) is not read either.
- **The two vocabularies do not intersect.** `site-header` asks for `nav_link_1_url` …
  `nav_link_4_label`, `cta_primary_url`, `cta_secondary_url`, `nav_aria_label`. `site-footer` asks
  for `col1_link1_url` … `col3_link4_label`, `newsletter_*`, `cookies_url`. None of these is in the
  map.
- **The one field that works proves it.** `company_name` *is* in the map — and is the only value
  that rendered: `<span class="header-logo-name">idea.uk</span>` and
  `aria-label="idea.uk home"`. Everything absent from the map is empty. 
- **Ungated is what makes it visible.** `grep -c '{{if'` → **0** in both idea.uk templates;
  `header-bold-gradient` gates every anchor (`{{if .cta_url}}…{{end}}`) and its `logo_url` has an
  `{{else}}` glyph fallback. vonc's `sites.logo_url` is *also* empty — its gate hides that, idea.uk's
  absence of one emits `src=""`.
- **The declared sources do not exist.** `site-header` sources URLs from
  `site_specs.navigation.link_N_url`; there is **no `navigation` aspect in `site_specs` for any
  site** (checked against the full fleet-wide aspect list). `logo_url` claims `site_assets.logo`.
  These `source:` declarations are decorative — nothing resolves them. Same reason
  `nav_aria_label` (`source:static`, `fallback:"Main navigation"`) rendered empty: **the fallback
  machinery never runs for chrome components at all.** (Note this differs from
  `plan_sections_action.go:1210-1218`, which *does* apply static fallbacks for page sections —
  cited in `023`. Chrome is a separate, thinner path.)

### CORRECTED — the URL-shape theory below is wrong

The original text speculated that empties and phantoms shared a cause: a resolver modelling
`/about` against a site built as `/about.html`. **Not so, for the empties.** idea.uk's nav tables
are in good order — 6 primary + 1 utility + 1 legal `site_nav_items`, all `status='active'`, all
correctly `.html`-suffixed (Home `/index.html`, Tools `/tools.html`, About `/about.html`, Guides
`/guides/index.html`, News `/news/index.html`, Report `/report.html`, Contact `/contact.html`,
Privacy `/privacy.html`). `GetNavItems` would return all of them. The header simply never asks for
`nav_items_html`. The `phantom_internal_link` findings come from **page** content and remain a
separate matter.

**What this means for the fix:** nothing needs rebuilding, re-planning or re-deriving. The data is
present and correct. Rewriting the two templates against the renderer's real vocabulary, with every
anchor gated, is a **DB-only change — live on the next render, no chassis image**.

### Original (superseded) speculation

The shape of the evidence:

- The chrome **templates render** — classes, structure, `site-footer.js` and the nav element are all
  present and correct.
- Only the **values** are missing: `href`, `img src`, and some link text. This is the signature of a
  data-fill step that ran with an empty/unresolved dataset, not of a broken template.
- Both logo *links* (`href="/"`) work — those are literals in the template. Every link whose target
  must be **resolved from site data** is empty.
- The `phantom_internal_link` findings show the resolver's model of the site uses extension-less
  paths (`/about`), while the built site is `.html`-suffixed (`/about.html`). A URL-shape mismatch
  between the nav model and the artefact would explain empties *and* phantoms together.

**Start here:** whatever populates `site_components` (header/footer) nav data at render time, and how
it derives page URLs — compare against `pages.url` (which holds `/about.html`). Related prior art:
the vonc.com link-integrity work (recompute across 6 CTA components, migration 098) and the
fleet-wide note that `source: "pages.*"` CTA fields revert on every render.

## Is this only idea.uk?

Unknown — **check before fixing per-site.** The header/footer are per-site rendered artefacts, but
the code that fills them is shared. A one-line check across the fleet:

```sql
SELECT s.domain, count(*) FILTER (WHERE sc.rendered_html LIKE '%href=""%') AS empty_href_components
FROM site_components sc JOIN sites s ON s.id = sc.site_id
WHERE sc.rendered_html IS NOT NULL
GROUP BY s.domain HAVING count(*) FILTER (WHERE sc.rendered_html LIKE '%href=""%') > 0;
```
If other domains appear, fix the filler, not idea.uk's rows.

## How to verify a fix

Against the **deployed artefact**, never the work-item status (this site has already produced a
`failed` item for a page that rendered and deployed correctly):
```bash
curl -s https://idea.uk/ | grep -oE '<a href="[^"]*"' | sort | uniq -c | sort -rn
#   want: no href="" rows; nav-link entries pointing at real .html pages
curl -s https://idea.uk/ | grep -o 'header-logo-img[^>]*'
#   want: a non-empty src=
```
Then re-run the auditors and confirm `empty_internal_href` and `phantom_internal_link` are gone:
`scripts/initial_messages/060improvement_loop/076_improvement_loop_trigger.sh <site_id> <domain>`
(reuse it — do not write another trigger; see `002 F`).

## Why it survived this long

idea.uk's static build was published to B2, where **nobody ever looked at it** — DNS pointed at the
VM, which served only the tool. The pages have presumably been broken since they were first built;
the VM cutover (2026-07-18) is simply what made them visible. No auditor had ever been pointed at the
site either. Two independent blindnesses, one visible outcome.

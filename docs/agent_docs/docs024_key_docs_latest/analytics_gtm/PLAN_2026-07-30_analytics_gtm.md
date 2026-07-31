# PLAN — Google Tag Manager across the framework's domains

**Started 2026-07-30.** Triggered by the owner asking for GTM container `GTM-PQ3WCTBD`
on every page of idea.uk, then for the best way to track traffic across *all* the
framework's domains.

---

## 1. Where the two tags can physically go

`assemblePage` (`platform/orchestration/actions/rerender_single_page_action.go:574-593`)
emits every chassis-built page as:

```
<!DOCTYPE html>\n<html lang="en">\n  ← Go string literals
head        ← site_components.rendered_html, slot_name='head'
\n<body>\n                            ← Go string literal
header      ← site_components.rendered_html, slot_name='header'
<main>sections</main>
footer
</body></html>
```

Two consequences that set the whole design:

- The `<head>` is the **`head` slot**, so the GTM `<script>` is a chrome edit.
- The `<body>` tag itself is **a Go string literal**. So "immediately after `<body>`"
  is *only* reachable as the **top of the `header` slot** — unless we change the
  chassis and roll an image. The header slot is therefore where the `<noscript>` goes.

## 2. Why the container id is per-site config, not a literal in the template

Chrome components are **shared across sites**. Measured 2026-07-30 over the 14
`status='deployed'` sites:

| slot | distinct components | spread |
|---|---|---|
| `head` | **3** | `Document Head` ×9 · `head-seo-standard` ×4 · `webdesign.co.uk Document Head` ×1 |
| `header` | **6** | `header-bold-gradient` ×7 · `header-professional-dark` ×3 · 4 single-site forks |

Hardcoding idea.uk's container into `Document Head` would have put idea.uk's tag on
**eight other domains**. So:

- the container id is a **per-site value** in `site_specs`, aspect `site_config`,
  at `analytics.gtm_container_id`;
- the template block is **gated** — `{{if .gtm_container_id}}…{{end}}` — so a site
  with no id renders **byte-identically** to before.

**No Go change is needed for this.** The resolution path already exists:
`render_site_components_action.go:585-645` gap-fills a component's `input_schema`
through `newSourceResolver`; `config.*` → `resolveConfigPath`
(`plan_sections_action.go:516-527`) searches `site_specs` aspects
`site_config` / `identity` / `design_intent` for the dotted path.

> ⚠ **Landmine.** `Document Head`'s existing `input_schema` is FLAT and holds bare
> SCALARS (`{"title": "...", "description": "..."}`). The gap-fill loop skips any
> value that is not a map (`:612-615`), so those three were never resolved and never
> could be. A **map-valued** key must be added alongside them for the loop to see it.
> The flat vs wrapped `{"fields":{…}}` shapes are both handled at `:604-607`.

## 3. Why BOTH the template and the rendered artefact must be written

`bugs_open/117`: chrome is a **stored artefact**. Pages assemble from
`site_components.rendered_html`, never from `content_components.html_template`.

- Template only → inert until some future chrome rebuild.
- Artefact only → silently reverted **by** that rebuild.

So every rollout step writes both, and the change is idempotent-guarded.

## 4. idea.uk is TWO applications behind one domain

The discovery that made "every page" bigger than it looked. nginx on the VM proxies
**16 reserved routes** to a Go binary (`/opt/idea/idea`, source in
`docs024_key_docs_latest/idea.uk/golang_files/`); everything else is the static site
synced from the `vm-sites` repo.

Eleven of those routes render HTML through one wrapper, `App.page()`
(`service.go`), and **none of them are in the static build**:

| route / page | why it matters |
|---|---|
| **"Payment received"** (`/order/success`) | **the conversion.** Untagged ⇒ Google cannot attribute a single sale |
| **"Request received"** | the £29 order submission |
| `/privacy` `/terms` `/refund-policy` | static `.html` copies **301 to these**, so the static tag never fires |
| `/order/cancel`, subscribe, audience-check result, operator pages | funnel + ops |

**Tagging the static site alone measures traffic and cannot measure a sale.** This is
the single most important finding in this workstream.

## 5. Phasing

| phase | scope | state |
|---|---|---|
| **A** | idea.uk static site (21 pages) | ✅ **DONE + LIVE 2026-07-30**, 19/19 fetchable URLs verified |
| **B** | idea.uk tool binary (11 pages incl. conversions) | ✅ code written, builds, 5 new tests pass — ⏸ **NOT DEPLOYED, awaiting owner** (live payment service, 1 order active) |
| **C** | remaining 13 domains | ⏸ **BLOCKED on two owner inputs** — see §5a. Not "repeat the recipe 13×" any more |
| **D** | Google-side account structure | ⏸ owner decision — see §6 |

## 5a. Phase C is blocked on two owner inputs, and one design decision (added 2026-07-31)

**Blocker 1 — container ids do not exist yet and cannot be invented.** `GTM-PQ3WCTBD` was
supplied *for idea.uk*. Rolling it to 13 more domains would put 13 businesses in one
container, which is the §6 recommendation but is a **decision, not an implementation
detail**. Either confirm one container estate-wide, or supply a container per domain.
Nothing can be applied until one of those exists.

**Blocker 2 — there is already a SECOND analytics seam, and the two conflict.**
Discovered 2026-07-31 (see NOTES; logged in `WRONG_CALLS.md`):

| seam | component | domains | mechanism | state |
|---|---|---|---|---|
| `config.analytics.gtm_container_id` | `Document Head` | 9 | **GTM** | live on idea.uk only |
| `config.analytics_id` | `head-seo-standard` | 4 | **gtag.js / GA4 direct** | **dormant — 0 sites set it** |
| — | `webdesign.co.uk Document Head` | 1 | neither | — |

`head-seo-standard`'s seam predates this workstream by two months and uses the *same*
pattern. **A site carrying both would load GA4 directly AND through GTM, double-counting
every pageview.** So Phase C must first decide which seam survives.

**Recommendation: GTM only.** Retire `analytics_id` — or make the template refuse to
render both — rather than leave a dormant mechanism that will read as reasonable to the
next session that finds it. Retiring it is safe today precisely *because* it is dormant:
0 sites set the key, so removal changes no rendered byte. That will stop being true the
moment someone populates it.

> ⚠ **Third trap, mechanical.** `webdesign.co.uk Document Head` uses a **lowercase**
> `<meta charset="utf-8">`. `replace()` is case-sensitive, so the idea.uk migration
> applied verbatim updates **0 rows** on that site and still reports success. Guard on a
> per-site anchor count, as `p4_34` does.

## 6. Google-side: how to structure this for many domains

The container is the easy half; the account shape is what decides whether the data is
usable. Recommendation, and the reasoning:

**One GA4 property for the whole estate, not one per domain.**
- Per-domain properties fragment the picture and make cross-site comparison manual.
- One property + a **`hostname` dimension** gives per-domain reporting for free
  (every GA4 event already carries `page_location`), and comparison across the
  estate becomes a filter rather than a spreadsheet.
- Keep the *option* of per-domain properties open by making the container id
  per-site from day one — which the `site_specs` design already does.

**One GTM container for the whole estate, not one per domain.**
- 14 containers means 14 copies of every tag change. The estate's sites are
  structurally identical (same chrome, same renderer), so the tags are identical too.
- Use GTM **Lookup Table** variables keyed on `{{Page Hostname}}` wherever a value
  genuinely differs per domain.
- ⚠ The counter-argument is **access control**: one container = one permission
  boundary. Split only when a domain must be handed to someone who must not see the
  others — which is a *client-handover* question, not a technical one.

**Cross-domain tracking: almost certainly do NOT turn it on.**
- GA4 cross-domain linking exists so ONE user journey spanning two domains is one
  session. These domains are separate businesses, not one funnel — linking them
  would merge unrelated sessions and inflate referral exclusions.
- The exception is any real hand-off (e.g. a site that sends buyers to another
  domain to pay). None is known today; check before enabling.

**Consent.** These are UK-facing sites, so PECR/UK-GDPR applies to analytics cookies.
GTM's Consent Mode v2 is the mechanism, and it needs a banner to feed it. **This is
currently missing estate-wide and is a compliance gap, not a nice-to-have.** Flagged
for the owner as a separate decision; it is not blocked by anything above, and the
`site_specs` key makes it easy to roll a consent tag alongside GTM later.

**Server-side tagging: not yet.** It costs a Cloud Run-ish endpoint per estate and
buys accuracy against ad-blockers. Revisit when a domain's revenue justifies it;
idea.uk is the only one with a money path today.

## 7. Open questions for the owner

1. **Deploy Phase B?** It restarts the live payment service. Rollback is a binary
   swap (`/opt/idea/idea.bak.*` pattern already in use). One order is currently active.
2. **Phase C now, or one domain at a time?** The mechanism is proven on idea.uk; the
   remaining 13 are the same three head templates.
3. **Consent banner** — separate workstream, or fold into C?

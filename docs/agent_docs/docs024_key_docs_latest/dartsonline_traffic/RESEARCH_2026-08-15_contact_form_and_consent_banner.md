# RESEARCH 2026-08-15 — the contact form and the cookie consent banner

Commissioned by the owner 2026-08-15: *"Let's sort out the contact form, possibly using one
of our existing vm boxes and possibly the tools-api or another api"* and *"We can add the
cookie consent banner to both sites."*

**Nothing was changed by this research.** No DB writes, no code edits, no dispatches.

**Headline: both questions turned out to be FLEET-WIDE, not dartsonline-only.** The owner asked about
"both sites"; the measurement says 11. That is the finding, and it changes who should decide.

---

## Q2 first, because it is the bigger one — the consent gap is 11 sites, not 2

`[MEASURED] 2026-08-15`, `curl -sL https://<domain>/` over the live fleet:

| domain | GTM container | consent markers | /privacy.html |
|---|---|---|---|
| dartsonline.com | `GTM-PQ3WCTBD` | **0** | 404 |
| idea.uk | `GTM-PQ3WCTBD` | **0** | 301 |
| finetuning.uk | `GTM-PQ3WCTBD` | **0** | 404 |
| gaswholesalers.com | `GTM-PQ3WCTBD` | **0** | 200 |
| robot-hands.com | `GTM-PQ3WCTBD` | **0** | 404 |
| oufe.com | `GTM-PQ3WCTBD` | **0** | 200 |
| vonc.com | `GTM-PQ3WCTBD` | **0** | 404 |
| relojistas.com | `GTM-PQ3WCTBD` | **0** | 404 |
| webdesign.co.uk | `GTM-PQ3WCTBD` | **0** | 404 |
| vetcomparison.uk | `GTM-PQ3WCTBD` | **0** | 404 |
| ai-agent-orchestration.com | `GTM-PQ3WCTBD` | **0** | 200 |
| noted.co.uk | **none** | 0 | (has a privacy page — `pages` row 2026-08-13) |

Consent markers searched for: `cookieconsent|cookie-consent|cookiebanner|cookie-banner|consent-mode|gtag('consent|onetrust|cookiebot|klaro|osano`.

**Three facts follow, and all three are the owner's to act on rather than a session's:**

1. **Eleven live public sites load a Google tag with no consent mechanism of any kind.** Under
   UK PECR, non-essential cookies (analytics included) require consent *before* they are set.
   This is a live exposure on every one of them, not a dartsonline defect.
2. **They all share ONE container, `GTM-PQ3WCTBD`.** That is the good news: it is plausibly a
   single place to fix — Google Consent Mode configured in that container would apply
   everywhere at once `[INFERRED — not verified inside the container, which I cannot see]`.
   It is also the risk: whoever else relies on that container inherits any change, and I could
   not establish from this repo who owns it or what tags it loads.
3. **Only 3 of the 11 have a privacy page at all** (gaswholesalers, oufe,
   ai-agent-orchestration; idea.uk 301s). So eight sites run analytics with neither consent
   nor a policy. The dartsonline privacy page now in flight fixes one of the eight.

**What I could NOT establish:** what the container actually loads (needs Google account
access, not repo access); whether a banner exists in any site's stored chrome rather than the
served homepage `[UNMEASURED — I checked served HTML only]`; and whether GA4 is even active
behind the container, which decides how urgent this is.

**Recommended route, and its blast radius.** Two layers, and they are separable:
- **The tag layer** — configure Consent Mode in `GTM-PQ3WCTBD` so nothing non-essential fires
  before consent. One change, 11 sites, no code, no deploy. Needs the Google account, so it is
  an owner action.
- **The banner layer** — the UI that captures the choice. This belongs in shared chrome
  (`site_components`, the `footer-theme-chrome` / `header-theme-chrome` rows). ⚠ **Chrome is a
  STORED artefact** (`bugs_open/117`): editing the shared template does not update existing
  sites. Each site's stored chrome must be re-rendered, which is the `nav_drift` → `nav-updater`
  path (confirmed draining: 26 completed items in 14 days, most recent 2026-08-15 20:19
  `[MEASURED]`) — though whether that path is correct for a *non-nav* chrome change is
  `[UNVERIFIED]` and must be checked before relying on it.

**This warrants a `bugs_open/` file** — it is a durable defect biting production on 11 live
sites right now, which is exactly that directory's question. Not filed from here because the
remedy is an owner/legal decision rather than an engineering fix, and filing it as a bug
implies an owner it does not yet have. Recommend the owner says whether to file it.

---

## Q1 — the contact form: the mailto: pattern is fleet-wide too

`[MEASURED]` — every `<form>` with an `action` in any page component, fleet-wide:

```sql
SELECT DISTINCT s.domain, substring(pc.rendered_html from '<form[^>]*action="[^"]*"')
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html ILIKE '%<form%action=%';
```

**Eleven sites carry the identical `contact-form` / `cf-contact-form` component with a
`mailto:` action** — ai-agent-orchestration, cookly, dartsonline, finetuning, fundamentallyai,
gaswholesalers, idea.uk, leopardess (×2 variants), noted, oufe. Two more (`relojistas`,
`pool-ai-agents.internal`) post to `#contact`, i.e. nowhere at all.

So the broken form is **one shared component on eleven sites**, and fixing it once fixes all
of them. That also sets the blast radius: this is a shared-seam change, so per CLAUDE.md's
2026-08-02 §2 ruling it wants an opt-in with the unsafe default OFF, not a flag day.

**The only working form endpoints in the estate are idea.uk's** — `<form class="ac-form"
action="/audience-check">` and `<form class="rr-form" action="/request">` `[MEASURED]`. Those
are the prior art worth reading before designing anything; I did **not** establish what serves
them (`[UNVERIFIED]` — they are not tools-api routes, see below).

**`tools-api` is real, deployed, and does not do this today.** It is a deployed service
(`deployments/kustomize/services/tools-api/`) and a Go/gin app (`cmd/tools-api/main.go`,
30 lines, delegating to `internal/tools-api/api.NewRouter`). Its entire route table
`[MEASURED]`, `internal/tools-api/api/server.go`:

```
GET  /health
POST /api/v1/tools/gauntlet/round      GET /api/v1/tools/gauntlet/round/:slug
POST /api/v1/tools/gauntlet/position
POST /api/v1/tools/gauntlet/defend
POST /api/v1/tools/gauntlet/publish
```

All vonc gauntlet-specific. **There is no enquiry/contact endpoint and nothing generic to
reuse** — but the service is a sound host for one: it already has a database pool, a config
loader, a router, a health check and a deployment. Adding an endpoint is a normal Go change
plus a build and roll, not new infrastructure.

**Recommended route.** Add a minimal enquiry endpoint to `tools-api` (it is the owner's own
suggestion and the evidence supports it), write submissions to a table, and notify by email;
then change the shared contact component to post to it, opt-in per site so the eleven do not
all flip at once. **Before designing it, read whatever serves idea.uk's `/request`** — if that
is already a working enquiry handler, extending it beats writing a second one, and CLAUDE.md
requires reusing existing machinery before building new.

**Risks / what I could not establish:**
- **Spam.** A public POST endpoint needs at least basic abuse protection. I found nothing in
  the estate that already solves this `[MEASURED — no rate-limit/captcha/honeypot found in
  tools-api]`, so it is part of the build, not an afterthought.
- **What serves `/audience-check` and `/request`** — the single most useful unknown here.
- **The "VM boxes"** — I did not establish what they are or what runs on them. There are
  `github-actions-runner-vmsites` and `wireguard` services in the deployment list, and a
  `vm-sites` git repo is referenced elsewhere in the docs, but I did not confirm any of that
  is a host with a public HTTP surface. `[UNVERIFIED]`
- Whether losing enquiries has actually happened. `mailto:` POST failing silently is general
  browser behaviour, not something I measured on these sites — so "enquiries are being lost"
  is `[INFERRED]`, and the honest test is to submit the live form and see what arrives.

# PLAN — about-page commercial elements (for-sale · advertise · built-by)

**Started:** 2026-07-24 · **Owner decision session:** "about page" thread
**Status:** design LOCKED, Phase 1 (DB-only pilot) starting
**Memory pointer:** `~/.claude/.../memory/about-page-commercial-elements.md`

## What we're building and why

Most sites the platform builds are portfolio assets: the domain is for sale, and
ad space can be sold on the site meanwhile. The about page (plus one discreet
footer line) gets a standard commercial block carrying up to three signals:

1. **Domain for-sale** → Afternic (T1 = brokerage, T2/3 = offer flow + minimum-offer floor)
2. **Advertise on this site** → advertise.co.uk (self-serve, flat-rate, being built in another thread)
3. **Built by fundamentallyai.com** → the storefront (doc 015's "store window")

Grounded in `docs/architecture/010-domain-value-maximisation.md` (tier model) and
`015-underserved-niche.md` (storefront + built-by badge). This is the owner's
written strategy, now implemented.

## Decisions and their reasons (all owner-ruled 2026-07-24)

> **⚠ CORRECTED 2026-09-03 — D4's WORKED EXAMPLE IS WRONG ON THE FACTS; THE RULING STANDS.**
> D4 below names *"a paying client's site (leopardess)"* as the relationship-breach case.
> **`leopardessconsulting.co.uk` is the owner's OWN consultancy**, not a client's site. Found by the
> `domain valuation` lane 2026-09-03: the live site reads *"Leopardess Consulting | AI systems that
> do one defined job, and keep doing it"* and its copy says *"We run 22 of our own sites on
> Kubernetes, Kafka and Postgres… before we build anything for you"* — this platform sold as a
> service — and the owner confirmed it directly to them as *"a representation of my own services"*.
> A relationship cannot be breached with oneself, so the example does not illustrate the harm D4
> guards against. **What does NOT change:** fail-closed, confirm-per-site is right regardless of
> the example, and the owner has separately ruled by name (2026-09-03, verbatim: *"no
> leopardessconsulting need not be listed"*) — so the site's exclusion now rests on his word, not
> on D4's reasoning. **What DOES change:** three lanes (`copy_quality_two_stage`, `sedo`,
> `domain valuation`) had recorded leopardess as a paying client on the strength of this row and
> have corrected. **Owner, 2026-09-03, verbatim: "there are no paying clients yet"** — and then,
> minutes later, the RULING that settles it: **"we can work as if leopardess is a paying client
> if that helps, I do pay through the nose for these tokens."** So D4's example stands **by
> ruling rather than by fact**: leopardess is his own consultancy AND is to be treated with
> paying-client care, which gives the relationship-breach reasoning a live case to attach to
> instead of a fictional one. Every lane's operational state (permanently excluded from listing;
> fenced; fence runs read-only) was already correct under this ruling; only the recorded REASON
> changes. The RUNBOOK's "keepers/clients (leopardess)" is therefore right after all, by the same
> ruling. Corrected visibly rather than rewritten: the row below is the ruling as made, this
> note is the fact as found, and the sentence above is the fact as ruled.

| # | Decision | Reason |
|---|---|---|
| D1 | Sell the **domain** (Afternic), not site-as-business | The asset is the name; no revenue/traffic to price a business on. Multi-route (Flippa/Sedo/brokers/direct) **deferred — owner will revisit**; doc 010 posed the question and never picked |
| D2 | **No price on any page**; per-site tier + Afternic **minimum-offer floor** | A public band anchors low and scares real buyers; the floor kills lowball spam mechanically; routing to the marketplace dodges our contact-form deliverability class of bug |
| D3 | Advertising = **self-serve flat-rate sponsored placement**, `rel="sponsored"` nofollow, ASA-labelled units | Near-zero traffic makes CPM/CPC dishonest; flat-rate is both the honest and the automatable model; paid dofollow links = Google penalty for both sides |
| D4 | Commercial signals **fail-closed**: absent config ⇒ keeper ⇒ nothing renders; for-sale **actively confirmed per site** | The one-directional damage: "buy this site"/ads/built-by on a paying client's site (leopardess) is a relationship breach; a portfolio site missing its callout is trivial |
| D5 | **Class** {keeper default / storefront / portfolio} × **mode facts**, not a stored mode | Sites move through phases (developing → advertising → for_sale); facts + render gates can't go stale or fight over write-order |
| D6 | **Passive Afternic listing always live; loud on-page callout only when for_sale_requested ∧ no active ad** | Never closed to a serious offer, never waving "for sale" at an advertiser. Consequence: advertise.co.uk T&Cs need an ad transfer/refund clause from day one |
| D7 | **Tier ≠ design route** | `site_type` (brochure/interactive-platform/…) tracks build effort; tier tracks the NAME's value. A brochure on a premium one-word .co.uk is still five figures. Never derive one from the other |
| D8 | Register: **"available to acquire"**, representation not adjectives; no "premium"/"serious offers"/"for sale" | Domainer clichés cheapen the signal; representation ("domain team") makes lowballers self-filter. "agent" avoided — overloaded in an AI-agents company |
| D9 | T1 destination = **Afternic brokerage** | Keeps a floor AND reads as represented; personal inbox loses the floor |
| D10 | **3 tiers** per doc 010 | Matches the owner's written portfolio strategy |
| D11 | Built-by = **fundamentallyai.com → home** | Site is live; understated line, no unverifiable claims (no "48 hours") |
| D12 | Footer carries **built-by only**; for-sale/advertise live in the about block | A site plastered with for-sale + ads reads as parked junk — hurts both sale and ads. Passive listing means no reach is lost |
| D13 | Flip = **API endpoint**, admin-controlled now, advertise.co.uk later via service credential | Owner: "front end api hook … connect advertise.co.uk later … admin page in control to start" |
| D14 | **Incremental build: Phase 1 DB-only pilot first** | Owner pick 2026-07-24. Most of the design ships without an image roll; prove it rendering on one pilot before the Go layer |

## The locked copy (register: understated, no prices, no hype)

- **T1:** "The [domain] name is available to acquire — acquisition enquiries via our
  domain team." · button **Contact our domain team** → Afternic brokerage
- **T2:** "The [domain] name is available to acquire — register your interest." ·
  button **Register your interest** → Afternic offer + floor
- **T3:** "[domain] is part of our portfolio and may be available to acquire." ·
  button **Make an enquiry** → Afternic offer + floor
- **Advertise:** "Advertise on [domain]. A small number of sponsored placements are
  available on this site — a flat monthly rate, set up in minutes at advertise.co.uk."
  · button **Advertise here** → advertise.co.uk/?site=<domain>. No price/reach/traffic
  claims on the host page; the frank low-traffic pitch lives on advertise.co.uk
- **Built-by:** "Built by fundamentallyai.com. We design and build sites like this
  one — see how it's done →" · footer one-liner: "Built by fundamentallyai.com →"

**Honesty rails:** the represented framing must be TRUE (a real monitored route —
brokerage listing or domains@); built-by carries no unverifiable claim; "Sponsored"
label + rel=sponsored apply to rendered ad UNITS, not the invitation line.

## Architecture (scouted 2026-07-24, greenfield — no existing for-sale/advertise machinery)

- **Config:** new `site_specs` aspect **`commercial`** (versioned, auditable):
  `{class, tier, for_sale_requested, advertising_active, inventory_open, marketplace_url}`.
  Absent ⇒ keeper ⇒ nothing renders. built-by URL is constant; advertise URL derived
  `advertise.co.uk/?site=<domain>`.
- **Gates (render-time):** built-by iff class∈{portfolio,storefront}; for-sale iff
  portfolio ∧ for_sale_requested ∧ ¬advertising_active; advertise iff portfolio ∧
  inventory_open ∧ ¬for_sale_requested.
- **About block:** new `content_components` section (`section_type:
  about-commercial-block`); commercial fields declared `source: site_specs.commercial.*`
  so they re-resolve every render and the copy LLM cannot override them
  (`resolve_internal_links_action.go:92-97` — resolved_data merges last). Eventually in
  the about default set (`apply_gap_plan_action.go:459-472`, Go, Phase 2).
- **Footer built-by:** via the shared footer render (`render_site_components_action.go:197`
  ContentData hook), gated on `commercial.class` — the footer populate is a fleet-wide
  cross-join, so the gate is what keeps keeper/client sites clean (Go, Phase 2).
- **Flip endpoint:** `POST /api/v1/admin/sites/:id/commercial` → new
  `HandleSetSiteCommercial` in `internal/core-manager/admin/site_admin_handlers.go`
  (clone of `HandleUpdateSiteSpec:190`, read-merge-write so advertise.co.uk can later
  set only `advertising_active`), wired in `server.go` siteGroup. Admin JWT now;
  external caller later via the `X-Bootstrap-Key` bootstrap pattern (auth choice
  DEFERRED). `WriteSiteSpecAction` is coordinator-only — not reachable over HTTP;
  write the table directly like the existing handlers do.

## Phasing

- **Phase 1 (DB-only, no image roll) — CURRENT:** `commercial` aspect written via
  existing machinery on ONE pilot site + `about-commercial-block` component row +
  attach to the pilot's about page + single-page render. Prove it serves live.
  Open question in flight: is the single-page attach+render truly DB-feasible
  (scout dispatched — pages.sections vs site_plan_sections, does the render path
  re-resolve site_specs.* sources, page_rerender does NOT re-select components).
- **Phase 2 (Go + council + image roll):** about default-set change; footer built-by
  gate; `HandleSetSiteCommercial`; admin-dashboard controls. Image first, then seeds.
- **Phase 3 (external):** advertise.co.uk integration (service credential, sets
  `advertising_active`/`inventory_open`), ad-unit component with Sponsored label +
  rel=sponsored, T&Cs transfer/refund clause. Owned by the advertise.co.uk thread;
  the endpoint contract here is the interface.

## Deferred / parked

- Multi-route selling (Flippa/Sedo/brokers/direct outreach) — owner will revisit;
  "at the end of the day I want to make money from the domains/sites and will try
  all routes."
- advertise.co.uk external auth choice (service JWT vs bootstrap-key).
- Which sites get which tier (per-site call at rollout, not a platform decision).

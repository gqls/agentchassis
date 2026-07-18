# 017 — Replacing a backend tool's landing page with a static site orphans its entry forms (no auditor catches it)

**Found:** 2026-07-18 on idea.uk, minutes after the VM cutover put the chassis static site at `/`.
**Severity:** the paid funnel is unreachable — on a live earning site. **Status:** open; needs both a
site fix and a new discovery check. **Class:** applies to every future class-B (static + backend on
one origin) site, not just idea.uk.

## Symptom

Owner report: *"https://idea.uk/audience-check produces page: POST only … I can't test the tool
because I can't find it."*

Verified from outside:
```
GET https://idea.uk/audience-check   -> 405   (tool: "POST only")
forms on the whole static site       -> exactly one: <form class="newsletter-form"> (no action)
links to tool endpoints              -> 2 × href="/audience-check"   (plain GET links)
                                        0 × anything posting to /request
```

## Root cause

The tool served its **own landing page at `/`**, and that page carried the actual entry UI — the
audience-check form and the report-request form — which POST to `/audience-check` and `/request`.
Those two routes are **handlers, not pages**: they only accept POST.

The cutover gave `/` to the static site. nginx correctly proxies all 16 tool routes, and the tool is
perfectly healthy — but **the forms that drove it are gone**, and the static site never had any. What
remains are two `href="/audience-check"` links, i.e. GET requests to a POST-only handler, which the
tool answers "POST only".

So the tool is running, reachable, and unusable. Nothing errored; every smoke test passes; the
funnel is simply absent.

**This was half-seen and not followed through.** RUNBOOK §3b's reserved-path list carried the note
*"…and `/` (the landing page it loses)"* — the loss was recorded as a routing fact, and nobody asked
*what was ON that page*. The pointer-page decision (RUNNING_NOTES §K) then pointed the "Free Audience
Check" card at `/audience-check` believing it was a page. It never was.

## Why no auditor caught it

The relevant checks exist and are enabled — `dead_controls`, `phantom_internal_links`,
`misdirected_cta` (completeness-discovery-agent); `content_image_missing`, `image_url_404`
(design-discovery-agent) — but **none of them models the backend**. They ask "does this href go
somewhere in the site?"; `/audience-check` *does* resolve (200-family from nginx), so it looks fine.
No check asks:

1. does a link's target accept the **HTTP method** the link will use (a GET `href` to a POST-only
   handler)?
2. does a site with a declared backend still contain a **form that posts to it** — i.e. is there any
   entry to the funnel at all?

`check_backend_unreachable.go` is the nearest existing check, but it tests reachability of the
backend, not whether the site can *drive* it.

## Proposed new check — `backend_entry_orphaned`

Registered like the others in `platform/orchestration/actions/discovery_checks/`, enabled on
`completeness-discovery-agent`. For sites with a backend (`deploy_config.target='vm'` or a
`requires-backend` component present):

- Build the set of backend routes and their accepted methods. Cheapest reliable source is a declared
  manifest on the site row (e.g. `deploy_config.backend_routes: [{path, methods}]`), because the
  chassis cannot introspect a foreign binary. Fall back to probing `OPTIONS`/`405` responses.
- **Finding A — `method_mismatch_link`:** any `<a href>` in deployed `rendered_html` whose path
  matches a backend route that does not accept GET. Severity high.
- **Finding B — `no_backend_entry`:** the site declares a backend but no deployed page contains a
  `<form>` whose `action` matches a backend route accepting POST. Severity critical — this is the
  "site cannot sell anything" case.

Both are cheap, static-analysis-only, and would have fired the moment the cutover landed.

## Immediate site fix (decide before building the check)

The static site needs real entry points. Options, cheapest first:

1. **Give the tool a page path.** Add a `GET /start` (or `/audience-check-form`) handler to the tool
   that serves its existing landing markup, proxy it in nginx, and point the site's CTAs there. The
   markup already exists in the binary — this is the smallest change and restores both forms at once.
2. **Put the forms in the static build.** Author the audience-check and request forms as chassis
   sections posting to `/audience-check` and `/request` (same origin, so no CORS). More work, but the
   funnel then lives in the site where the rest of the copy is, and the honeypot/timing fields from
   the `/request` hardening can be authored in properly.
3. **Interim stopgap:** repoint the two `/audience-check` links at whatever page ends up hosting the
   form, so no visitor meets "POST only".

Option 1 restores service fastest; option 2 is the better end state. They are not exclusive.

## Also outstanding on this site (separate, existing-check territory)

Direct inspection of the deployed HTML found two more classes that existing checks **would** catch —
they have simply never been run here (see below):
- `<img class="header-logo-img" src="" alt="idea.uk logo" />` — empty `src`, though the asset exists
  (`/assets/images/logo.jpg` → 200). Owner sees the alt text.
- `<a href="" class="btn-primary">`, `<a href="#" class="brief-explanation__cta-primary">` — dead
  controls, several per page.

## The operational gap behind all of it

**No discovery check has ever run against idea.uk.** `SELECT source, count(*) … GROUP BY source` for
the site returns 13 sources across 112 items; `discovery` is not among them (0 rows, query cleared
against the full source list — not a query miss). The auditors are configured fleet-wide on three
agents but nothing had pointed them at this site.

An on-demand trigger now exists: `docs024_key_docs_latest/idea_uk_vm_site/TRIGGER_discovery_agent.sh
<agent-type> <site_id>`. Two snags found while writing it, both worth fixing:
- the workflow opens with `ensure_site_record`, which resolves **by domain** and dies
  `"domain not found in input_data"` if handed only a `site_id` (cost one run). The script now looks
  the domain up and passes both.
- with the domain supplied, later dispatches produced **no `orchestration_states` row at all** — the
  envelope was accepted but nothing ran, and the chassis pod was 6h old so the documented
  300s-post-restart drop does not explain it. On-demand discovery dispatch needs its own diagnosis.

**Wider question worth answering:** should discovery run automatically after a *deploy-target or
origin change*? Every check assumes the site is served the way it was built. A cutover changes the
serving model underneath a site that no check re-examines.

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

---

## STATUS UPDATE 2026-07-21 (session "bugfix 017") — site fix LIVE & holding; two of this file's premises are now stale

**The immediate site fix is DONE, live, and still holding** — verified against the deployed site
today, not against work-item status (idea_uk workstream RUNNING_NOTES §X.5 did the original
end-to-end verify 2026-07-20; re-confirmed 2026-07-21):
```
GET /                -> 200   href="/audience-check" = 0   (was 2 GET links to a POST-only handler)
GET /tools.html      -> 200   <form action="/audience-check" method="POST"> present  (taster)
GET /report.html     -> 200   <form action="/request"> present                        (report funnel)
GET /tools/assets/audience-check-form.js -> 200            (JS interceptor; fetch+inject, no navigation)
GET /audience-check  -> 405   (correct — it is the POST-only handler; no browser route GETs it anymore)
```
So the funnel is reachable and driven again. It was fixed by authoring the two forms as chassis
sections with a JS interceptor (idea_uk `sql/p2_01`, `p3_03`, `p3_04`) — the handoff's option 2
("put the forms in the static build"), not option 1. The `/audience-check` card URLs were retargeted
via the `pages` source of truth (`tool-audience-check.url` → `/tools.html#audience-check`), not the
transient `content_data` copies (see `bugs_open/001`).

**Two premises in the sections above are now stale — corrected against the live DB 2026-07-21:**

1. *"No discovery check has ever run against idea.uk … discovery is not among the sources (0 rows)."*
   **No longer true.** idea.uk now has **30 `source='discovery'` work items** (9 phantom_internal_link,
   8 cta_names_unknown_destination, 4 dead_control, 3 empty_internal_href, 1 required_fields_missing,
   …). The on-demand trigger work closed this operational gap. **Notably, none of the 30 is the 017
   symptom** — no existing check models "a GET `<a href>` to a route that only accepts POST", which is
   exactly the gap this file argues for and it is still open.

2. *Proposed check gates on `deploy_config.target='vm'`.* **idea.uk's `deploy_config` is empty `{}`** —
   no `target`, no backend marker of any kind, despite the site being VM-hosted with 16 proxied tool
   routes. Consequence the handoff did not foresee: **`check_backend_unreachable` (same `target='vm'`
   gate) has never probed idea.uk either**, and `backend_entry_orphaned` as specified would NOOP on the
   very site it was written for. The backend is not modelled in the data — that is the deeper root, and
   any target-gated check inherits it. A probe-based Finding A (GET a linked path, treat a 405 as a
   method-mismatch link) sidesteps the gate entirely and would fire on any site regardless of tagging.

**Remaining open for 017:** the durable auditor gap only — the `backend_entry_orphaned` check (or an
un-gated probe variant of Finding A). The site half is closable. `contact_form_undeliverable`
(landed 2026-07-20, `bugs_open/006 §B`) is a *sibling*, not this: it flags contact-form components
with dead actions and explicitly treats POST handlers like `/request` as valid destinations — it does
not cover the GET-link-to-POST-only-handler case.

### Finding A BUILT (2026-07-21) — `backend_entry_orphaned` / `method_mismatch_link`

`platform/orchestration/actions/discovery_checks/check_backend_entry_orphaned.go` (+ `_test.go`),
committed `7b03f296a`. Follows the owner-chosen **probe-based, un-gated** design (session decision
2026-07-21):

- Reads deployed `page_components.rendered_html` (active + deployed, `page_components` only — chrome
  has its own fixers), extracts anchors via the canonical `datahelpers.ExtractAnchors` /
  `ClassifyLinkScope`, keeps only internal **extensionless** handler-like routes (a cost filter:
  funnel handlers are clean paths like `/audience-check`; `.html`/assets can't 405), dedupes per
  destination path, caps probes at 40 (logged on hit — no silent cap).
- Live-probes `GET https://<domain><path>` (GET, not HEAD — reproduces the click) and files a
  **high** `needs_human_review` item (no handler — repointing the link vs authoring the form is a
  business decision) for **exactly HTTP 405**. 405-only keeps it off `phantom_internal_links`' (404)
  and `backend_unreachable`'s (5xx) turf.
- **Un-gated by `deploy_config.target='vm'`** on purpose — idea.uk's `deploy_config` is empty `{}`,
  so the gate the sibling `check_backend_unreachable` uses would NOOP here.

**Verified live 2026-07-21** (the induced-failing-branch, not just wiring): `GET /audience-check` →
405 and `GET /request` → 405 → both FLAG (the exact symptom); `GET /subscribe` → **400**,
`/tools.html` → 200, `/health` → 200, a bogus path → 404, `/` → 200 → all correctly ignored. Unit
test pins the filter (17 cases); `go vet` clean. On idea.uk *today* it reports clean (the orphaned
links were removed by the site fix), so it won't false-positive on the fixed site.

**Status:** committed, **under council review** (advisory, corr `ed4851c9-e51b-446d-a4b4-bbbf516eaa60`).
**Inert** until (1) an image roll carries the file into the pod, then (2) `backend_entry_orphaned` is
added to a discovery agent's `checks` array (config, live immediately — but image-first, else it
references an unregistered check). **017 stays OPEN** until it is live per the fixed-AND-live bar.

**Still not built (deferred, follow-on):** Finding B `no_backend_entry` — a site declares a backend
but no deployed page carries a `<form>` posting to a backend POST route ("site cannot sell
anything"). Left out because it needs a reliable "does the site have a backend" signal, which the
empty `deploy_config` shows is itself missing — a data-model gap worth its own decision.

#### Council round 1: REVISE (corr `ed4851c9`, 2026-07-21) — 9/12 approve, 3 object, 0 unreadable

Every seat that judged the detection logic **approved** it ("well-targeted", "correct architectural
shape", "unusually well-evidenced"); the reviewers' own read-only checks independently **confirmed**
every evidence claim (deploy_config `{}`, 30 discovery items none matching, `pages.status='active'`
covers the deployed population). Three objections, two of them actionable — both now fixed
(commit `a02e27853`):

- **bug_historian [medium] — FIXED.** `probeGETStatus`'s error was effectively swallowed, so a
  TLS/DNS/timeout blip read as "not a 405" → clean. A check whose job is catching silent failures
  must not go silently blind on the fragile VM backends it watches. Now counts unchecked routes and
  logs `Warn` that they are UNCHECKED (not proven clean).
- **editquality [low] — FIXED.** Test pinned only the filter. `probeGETStatus` now takes a full URL;
  added an httptest test covering 405-flag / 200 / 400 / 404 **and** the failed-probe-must-error case.
- **editquality/guardian [medium] — CLARIFIED (not a code gap).** "Plan omits registration." The
  check **self-registers** via `init(){ Register(...) }` exactly like every sibling — not dead code;
  `RunDiscoveryChecksAction` logs `checks.Names()` to prove it. **Enablement** = add
  `"backend_entry_orphaned"` to the `config.checks` array of the `run_discovery_checks` workflow step
  (`RunDiscoveryChecksAction` reads `config["checks"]`, defaulting to a 4-check list). That is a
  config edit, **image-first** (a named check that isn't registered is skipped, non-fatal), so it is
  a deliberately-sequenced follow-on, not an omission.
- **bug_historian [high] — TRACKED, by design.** This is the *auditor* half; the *cause* — a static
  cutover (a manual nginx op, README §4) that overwrites `/` and silently orphans the tool's forms
  with no check on what it replaces — is untouched. bug_historian's own note calls this check "the
  correct architectural shape" and asks that the cause be tracked separately. **Follow-on:** a
  *pre-cutover* content/route-diff guard (what routes+forms is the old `/` serving that the new `/`
  will drop?) belongs in the idea.uk cutover RUNBOOK §4 and generalises the handoff's "should
  discovery run after a deploy-target/origin change?" It is prevention; this check is detection.

#### Council round 2: REVISE (2026-07-21) — 8/12 approve, and the objections CONVERGED

Two-round advisory verdict, both REVISE, 0 unreadable. Round 2's objections collapsed onto **one
theme, raised by three seats** (editquality/bug_historian/prior_art [all medium]): *the check
self-registers but ENABLEMENT is deferred, so it never actually runs — the edits make the code
available, not live; the dormant-machinery pattern (cf. bugs_open/044).* Answered concretely:

- **Enablement seed WRITTEN — `docs/agent_docs/sql_for_agents/188_enable_backend_entry_orphaned_check.sql`**
  (commit `3d54eb421`), modeled on `165_enable_dead_controls_check.sql`. It appends
  `backend_entry_orphaned` to the live gate — `default_config #> '{workflow,steps,run_checks,config,checks}'`
  on `completeness-discovery-agent` (currently 20 checks). **NOT applied — image-first** (verify
  `strings /app/agent-chassis | grep BackendEntryOrphanedCheck` first; on an older image the name is
  logged and skipped). `contact_form_undeliverable` and `backend_unreachable` are **also**
  built-but-not-enabled on that agent today, so this is the normal image→seed lifecycle, not a defect
  unique to this plan.
- **improvement_guardian [medium] status nit — ANSWERED.** `needs_human_review` at insert is verified
  precedent: `dead_controls` and `contact_form_undeliverable` both do it (no auto-fixer → a human
  decides). `backend_unreachable` uses `detected` only because it **self-clears**; mine does not, so
  `needs_human_review` is the correct, precedented choice.
- **bug_historian [high] symptom-vs-cause — RECURRING → OWNER SCOPE DECISION.** Raised both rounds. Per
  the idea.uk chrome-fix precedent (REVISE twice on a scope question → take it to the owner, don't
  revise past it a third time), the council loop is **stopped at round 2**. The fork for the owner:
  ship the detector alone (this check, ready + enablement seed ready), or also commission the
  *pre-cutover content-diff guard* (prevention) as a sibling piece before this ships.

**Net state:** detection code committed + live-verified + council-approved-on-logic; enablement seed
committed + ready (image-first). 017 stays OPEN until the image roll + seed 188 applied + verified in
pod. The only open judgement is the owner scope call above.

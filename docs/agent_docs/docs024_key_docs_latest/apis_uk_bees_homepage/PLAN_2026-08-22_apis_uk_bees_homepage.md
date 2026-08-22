# PLAN 2026-08-22 — apis.uk bees home page

**Owner ask (this session):** "build a page about bees for the apis.uk home page
but without affecting the dns for the tools-api that runs on that same domain."

**Scope decided with the owner, 2026-08-22:** HOME PAGE ONLY, and the angle is a
personal / enthusiast page — characterful, informative, non-commercial, no offer
and no lead capture. Not a beekeeping how-to, not a conservation campaign.

This is the thread that three other documents have been waiting on. Owner intent
was recorded on 2026-07-25 in `gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md:551`:
*"apex apis.uk will become a BEES homepage (unrelated to the API), built in
another thread"*, and again in `features_open/020_FEATURE_apis_uk_traffic_probe.md`
and `gauntlet_dead_cta/infra/island/RUNBOOK_island.md:76`.

## 1. The DNS constraint — measured, and it turns out to be already satisfied

The owner's constraint is the load-bearing part of the ask, so it was measured
before anything was designed. **The result inverts what the standing docs say.**

All three of the documents above state that serving the bees page will need
"one DNS record swap" on the apex, with the wildcard and probe unaffected. That
was true when written. It is not true now.

**Live zone state, read from the Cloudflare API 2026-08-22** (zone
`a8c1ac6111424c218cb9e9368ed0586f`):

| record | type | content | proxied |
|---|---|---|---|
| `apis.uk` | CNAME | `f917c7c1-….cfargotunnel.com` | yes |
| `*.apis.uk` | CNAME | `f917c7c1-….cfargotunnel.com` | yes |
| `tools.apis.uk` | CNAME | `f917c7c1-….cfargotunnel.com` | yes |
| `www.apis.uk` | A | `192.0.2.1` | yes |

Worker routes on the zone: `apis.uk/*` → `portfolio-sites-router`, and
`www.apis.uk/*` → `portfolio-sites-router`. **There is NO `*.apis.uk/*` route.**

A Cloudflare worker route intercepts at the edge *before* the origin is
consulted, so the apex CNAME to the tunnel is already vestigial: `apis.uk` is
served by `portfolio-sites-router` today, not by the island.

**Verified at the artefact, not inferred from the config** — four hostnames,
three distinct behaviours, each identifying its server exactly:

| hostname | response | therefore served by |
|---|---|---|
| `apis.uk` | 404, body `Not found` | the worker (`scripts/cloudflare/worker.js:91` returns exactly this string) |
| `www.apis.uk` | 301 → `https://apis.uk/` | the worker (its www→apex branch, `worker.js:23`) |
| `zzqq-probe-test.apis.uk` | 404, 0 bytes | island probe vhost :8082 (tunnel) |
| `tools.apis.uk` | 404, 0 bytes | island Caddy :8081 (tunnel) → tools-api |

**CONSEQUENCE — NO DNS CHANGE IS REQUIRED FOR THIS TASK.** The apex already
routes to the portfolio worker; the worker serves B2 object key
`<hostname><path>`; so putting bytes at `b2://portfolio-sites/apis.uk/index.html`
publishes the home page with no zone edit at all. The safest change to
`tools.apis.uk` is the one we are now making: none.

## 2. What could still break tools-api, and the rule that prevents it

The DNS records are not the real hazard — they are per-name and independent.
**The hazard is the worker route pattern.** `apis.uk/*` matches the apex only.
A route `*.apis.uk/*` would match `tools.apis.uk`, intercept it at the edge
ahead of the tunnel, look up a B2 object that does not exist and serve a 404 —
silently killing the live API with no DNS record having changed.

`scripts/cloudflare/add_www_redirect.sh` records that 24 zones already carry
exactly that wildcard route. apis.uk is deliberately not one of them.

> **STANDING RULE FOR THIS DOMAIN: never add a wildcard worker route to the
> apis.uk zone, and never run a "give every zone the standard treatment" sweep
> against it unnamed.** The API lives on a subdomain of a portfolio-served apex,
> which is a shape no other zone in the estate has.

Filed as a landmine, because it fires on touch and the wrong result looks
exactly like the right one (a clean green sweep, and an API that 404s).

## 3. Build route — the framework, per the owner ruling of 2026-08-04

No hand-authored HTML. The site is seeded and then submitted, and the framework
writes the content:

1. **Site row** with an email (`bugs_open/063`: the hallucinated-email check
   FAILS OPEN when a site has no contact email).
2. **`evidence_base` aspect, seeded BEFORE the first page is written.** Grounded
   in code, not in the runbook: `loadEvidenceBase`
   (`platform/orchestration/actions/validate_page_content.go:1272-1290`) returns
   `nil` on `sql.ErrNoRows`, and every claims lane then silently no-ops. A site
   with no evidence base is not "unchecked-but-fine", it is *unchecked and
   reports clean*.
3. **`imagery_style_guide` aspect** (`bugs_closed/027`: `content_hero` generates
   unstyled on a site that has none, and a fresh site has none).
4. **`roadmap_brief` aspect naming ONLY the home page.** This is the mechanism
   that makes "home page only" real. Grounded in the live agent definition, not
   the runbook — `build-site-planner`'s prompt reads
   `{{.site_specs.specs.roadmap_brief.text}}` and says: *"ROADMAP OVERRIDES THE
   COMPONENT LIST. Build ONLY the pages listed in the current phase below. …
   Do NOT invent additional pages. The roadmap is the authority for this site."*
5. **Submit** via the `domain-submitter` envelope with a mission brief.

## 4. Why bees, and the one fact the whole page rests on

*Apis* is the genus of the honey bee (*Apis mellifera*). The domain name is the
joke and the reason the page exists — an API domain whose apex is about the
other kind of apis. The mission brief carries that hook and lets the framework
write from it.

## 5. The fabrication risk, named before it happens

Bees are a subject made almost entirely of famous repeated numbers: the share of
food that depends on pollinators, the number of flowers visited per jar of
honey, the miles flown, the bees per hive, the percentage decline. Every one of
them is quotable everywhere and sourced nowhere we have checked. There is also a
specific well-known misattribution — the "four years left to live" line, which
Einstein did not say.

`facts[]` is therefore deliberately EMPTY and the ban list targets the SHAPES,
not individual figures, exactly as the oufe seed does. The page can be charming
and interesting without asserting a single unverified quantity; if a sentence
seems to need a number, the sentence gets rewritten.

**No figures in the mission brief either** — a number in a spec is a *given* and
outranks every writer-side rule (oufe RUNBOOK §3).

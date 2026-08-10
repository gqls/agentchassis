# 236 — a fully built site can serve **522 to every visitor** indefinitely, and nothing in the platform notices

**Filed:** 2026-08-09 by the `bugfix_209_deploy_purpose_keyed_source` lane, after the
owner asked "can you check lendzy.co.uk" and it turned out to be **down**.
**Status:** OPEN, **OWNED since 2026-08-10** by
`docs024_key_docs_latest/bugfix_236_site_availability/` (this file's candidate 1:
a `site_unreachable` discovery check + a 4-hourly rotation driver; candidate 2
stays with `domains_cloudflare_rollout`). **Severity:** high — this is total,
silent, public unavailability of a finished site, with every internal signal
green.

**Re-verified 2026-08-10 before claiming:** still no availability-shaped
scheduled task, discovery check, or open work item; all 21 deployed sites probed
— 21/21 serve 200 today, so the class is live and currently quiet.

> ### STATUS 2026-08-10 22:15Z — candidate 1 is **LIVE AND RUNNING**; the drill is what remains
>
> **Superseding the banner below**, which was written before the roll.
> `check_site_unreachable` is in chassis **v1.0.1283** (both replicas pod-grepped
> with an invented negative control at 0), **migration 372 is APPLIED**, and the
> 4-hourly rotation completed its first live probe at 22:03:49Z on
> `robot-hands.com` — `checks_run:[site_unreachable]`, no failures, 0 items,
> correct for a serving site. That run also settled the lane's one open
> assumption in a disconfirmable way: blocked pod egress would have FILED an
> item, so silence proves the probe really reached the internet.
>
> **STILL OPEN, and this is the whole reason:** the FILING path has never fired
> in production — the check has never seen a real 522. `0 findings` is also what
> a blinded check reports. **The break-it-on-purpose drill in "How to verify a
> fix" below is owed**, and two costed ways to run it (a pool-site induction and
> the cookly route deletion) are written up in
> `bugfix_236_site_availability/HANDOFF_2026-08-10_continue_here.md`. Also owed:
> a fresh council round once `bugs_open/243` (the Anthropic cap) clears.
>
> ### STATUS 2026-08-10 (evening) — candidate 1 BUILT AND COMMITTED, not yet live
>
> **Committed `4a5d77004`:** `check_site_unreachable` (discovery check, probes
> `https://<domain>/` per site, confirms before filing, one alert-only
> `site_unreachable` item per site via `item_key site_unreachable:<site_id>`,
> self-clears through `Resolved{AllOfType}` when the site serves), its tests
> (8 guards each proven load-bearing by mutation with a NAMED failing test), and
> register entry **IMP-053**.
>
> **NOT live.** The Go half rides the next chassis roll. The driver config —
> `availability-discovery-agent` + `site-discovery-rotation-availability`, a
> 4-hourly fair rotation reusing the 230 lane's `site_discovery_rotation` table —
> is HELD as `sql_for_agents/372_site_availability_driver_HOLD.sql` until a
> pod-grep proves `site_unreachable` is in the running binary, because
> `run_discovery_checks` hard-fails on an unregistered check name (149 B4).
>
> **NOT reviewed.** Submission `7177fb02-51c5-4c2a-bb02-10aa27ae85ca` selected its
> panel and then died at the first seat on an upstream Anthropic 400 (account
> spend cap, "regain access 2026-09-01"; fleet-wide since ~14:51Z — now filed as
> **`bugs_open/243_…_anthropic_account_usage_limit_…`** (note the ambiguous number:
> the OTHER 243 is the tool-acceptance storage client), and independently
> diagnosed the same afternoon from a standalone service outside the cluster by
> the webdesign.uk chat lane in `6a4fbab21`). The run is terminal at
> `complete_invalid` — **which is NOT a rejection of the submission**; treat the
> `Council-Submitted:` trailer on `4a5d77004` as "submitted, never reviewed". A
> fresh submission is owed when LLM access returns.
>
> **Two deliberate non-filings, measured on all 21 live sites, so a later reader
> does not read them as gaps:** an off-domain redirect (`webdesign.uk` → 302 →
> `webdesign.co.uk`) and an on-host 2xx whose body lacks the stored index title
> (`mortgagecalculator.co.uk` serves a divergent render — a staleness defect with
> its own machinery) are named FINDINGS, not work items. Filing on the second was
> 1-in-21 false-positive on day one. **Consequence: a registrar-parked domain
> answering 200 files nothing** — the LANDMINES trap this file's candidate 1
> warned about is only half-closed, deliberately, and is recorded in IMP-053.
>
> **Candidate 2 (zone/route conformance) is NOT this lane's** — left with
> `domains_cloudflare_rollout`, as this file suggests. **Candidate 3 is not taken.**
> **This file stays OPEN** until the check is live and the break-it-on-purpose
> drill in the "How to verify a fix" section below has actually been run: a
> checker that has never fired on a real 522 is not evidence.
>
> Lane docs: `docs024_key_docs_latest/bugfix_236_site_availability/`
> (PLAN / NOTES / RUNBOOK / README_where_we_are / COUNCIL_SUBMISSION).

## What happened

`lendzy.co.uk` returned **HTTP 522** (Cloudflare: connection timed out to origin)
after 19.3s, on every path. Its zone was `active`, its apex `A 199.59.243.228` was
`proxied: true`, and its content was complete and correct — **33 files** in
`gqls/sites` since 2026-08-02, synced to the bucket.

**Cause: the zone had NO worker routes at all.** The estate serves sites by
intercepting proxied requests with the `portfolio-sites-router` worker, which reads
the bucket. `199.59.243.228` is only a **placeholder origin** that accepts no
connections — nothing is meant to reach it. With no route, Cloudflare proxies
straight through to that dead IP: 522.

Fixed by creating the one missing route (`lendzy.co.uk/*` →
`portfolio-sites-router`); the site returned **200, 41,431 B, "Lendzy — Know the
Rules Before You Borrow"** within ~30 seconds.

`[UNVERIFIED]` **How long it was down.** Cloudflare does not expose route creation
times through the API used here, and there is no local record of the route ever
existing. Content has been present since 2026-08-02, so the outage is bounded only
by "at most since the site was built". **Do not state a duration** — but note that
if the route never existed, the site was never once publicly viewable.

## The actual defect: nothing checks whether a site SERVES

The platform has extensive per-page quality checking and a health checker, and
**neither answers "is this site reachable at all"**:

- `endpoint-health-checker` / `check_endpoint_health_action.go` pings **AI
  endpoints** (Ollama `GET {url}/api/tags`) and writes an endpoint health table. It
  has nothing to do with the public sites.
- The discovery-check layer runs against **stored/rendered artefacts**, not the
  served origin. Of all the checks in `discovery_checks/`, exactly **one** makes an
  outbound HTTP request — `check_asset_reference_404.go` — and it probes referenced
  **subresources**, so it only runs after a page has been fetched and parsed.
- Every internal signal for lendzy was green throughout: `sites.status='deployed'`,
  its work items `complete`, its pages `deployed`, the B2 sync workflow `success`,
  the commits present. **The whole pipeline reports success for a site nobody can
  load.**

This is the "trust the rendered artefact, not the status" rule (CLAUDE.md) with the
last link missing: we verify the artefact in the bucket and in git, never that the
public URL returns it.

## How it stayed invisible

- A 522 is generated by **Cloudflare**, not by us — no error reaches any log, table
  or queue we own.
- It is a *routing/config* absence, not a failure: nothing errored, so nothing
  retried and nothing alerted.
- Fleet censuses key on the things that exist (zones, records, pages, work items).
  A missing route is the absence of a row, which no census of rows can see.

## Fix candidates, ordered by what closes the door

1. **A scheduled availability check over every live domain** — one HEAD/GET per
   site, assert 2xx/3xx and a body that is not empty, file a work item on failure.
   Cheap (~40 requests), catches 522/523/526, dead origins, expired certs and
   dangling delegations in one sweep. The probe helper already exists —
   `check_asset_reference_404.go:220-236` is a working timeout-bounded fetch to
   copy. **Assert a BODY property, not just the status** (see LANDMINES 2026-08-09:
   a parked domain answers 200 on every path).
2. **A route/zone conformance check**: for every active zone, assert the estate
   template — proxied apex A, an apex worker route, and the www CNAME + page rule.
   This catches the cause rather than the symptom, and would also have caught
   `loanzy.uk`'s dangling delegation. Census written by hand today (below); it is
   ~15 lines of API calls and belongs in a CronJob.
3. **Make the placeholder origin honest**: `199.59.243.228` accepting nothing is
   what turns a missing route into a 19-second hang. An origin that returned an
   explicit "no route configured" page would make this self-diagnosing. Weakest —
   it improves the message, not the detection — and the IP is not ours.

## Census run today (2026-08-09), so nobody repeats it

All **38 active zones** checked for an apex worker route. Four lacked one:

| zone | verdict |
|---|---|
| `lendzy.co.uk` | **DOWN (522)** — fixed, this bug |
| `idea.uk` | fine — served from its own VM (`idea_uk_vm_site` cutover), 200 |
| `relojistas.com` | fine — VM-served, 200 |
| `webdesign.uk` | fine — deliberate 302 → `webdesign.co.uk`, 200 |

**The three false positives are the point**: "no worker route" is correct for a
VM-served or redirecting domain, so candidate 2 must treat the rollout lane's
skip-list (`relojistas.com`, `finetuning.uk`, `webdesign.uk`, `idea.uk` —
`domains_cloudflare_rollout/PLAN`) as first-class, or it will cry wolf on every
run. I tested all three rather than reporting them, which is the only reason this
file says "one outage" and not "four".

## How to verify a fix

Break it deliberately on a sacrificial domain: delete `cookly.uk/*`'s worker route,
confirm the checker files a work item within its interval, restore the route,
confirm the item closes. A checker that has never fired on a real 522 is not
evidence (LANDMINES: a gate's 0 findings has two causes).

## Related

- `bugs_open/210` (needs_logo file) §6 and `bugs_closed/128` — the same family one
  layer in: checks that read a basename or a stored artefact rather than what is served.
- LANDMINES 2026-08-09 — the parked-domain 200 trap (why a status-only probe is not
  enough) and the dangling-delegation trap (`loanzy.uk`, same session).
- `domains_cloudflare_rollout/` — the lane that owns zone conformance; this bug's
  candidate 2 is arguably theirs.

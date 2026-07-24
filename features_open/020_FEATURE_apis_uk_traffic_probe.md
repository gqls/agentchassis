# 020 — FEATURE: traffic probe on apis.uk (where is existing traffic coming from?)

**Filed:** 2026-07-24, owner request ("bastion host" session). Owner asked for a
"bug report to add a traffic sniffer on apis.uk … like we did with relojistas";
filed as a FEATURE because nothing is broken — this is new observation.
**Precedent:** the domain traffic-probe project,
`docs/agent_docs/docs024_key_docs_latest/traffic_probe/TASK_traffic_probe_brief.md`
(archive.org research first, then a probe that logs what visitors actually ask
for — the relojistas/idea.uk-era model). This filing is the passive first stage
of that method, applied to apis.uk.

## Why now

apis.uk just went live on Cloudflare with one real hostname (`tools.apis.uk` →
the tools-api island via Cloudflare Tunnel `f917c7c1…`). Before we tightened it,
the zone carried a proxied `*` wildcard pointing at a dead origin (every name
525'd) — evidence the domain has SOME resolving/requesting history. We currently
learn nothing from that traffic: the wildcard is gone, so every other name
NXDOMAINs and the requests never reach anything we can see.

## Design (stage 1 — passive, ~30 min on the island, no new infra)

1. **Re-add DNS**: `*` and apex `apis.uk` as proxied CNAMEs →
   `f917c7c1-4dae-446f-a1e0-8f4c636cc345.cfargotunnel.com` (the existing tunnel).
   NOTE this deliberately REVERSES the morning's "delete the dead wildcard"
   advice: same record shape, now pointed at OUR tunnel instead of a dead
   origin. The explicit `tools.apis.uk` record keeps precedence.
2. **Tunnel ingress**: add a catch-all hostname rule (`*.apis.uk` + apex) in
   `/etc/cloudflared/config.yml` → `http://localhost:8082` (before the 404
   fallback; `tools.apis.uk` rule stays first).
3. **Island Caddy probe vhost** on :8082 (loopback-bound like :8081): responds
   404 to everything (no content served — parked stays parked), but logs a
   structured JSON line per request: timestamp, Host, method, path, query,
   Referer, User-Agent, `CF-Connecting-IP`, `CF-IPCountry`. Log to
   `/opt/island/logs/probe/access.log`, logrotate 30 days.
4. **Review** after 2–4 weeks: rank by hostname × path; Cloudflare's free zone
   analytics as the coarse cross-check.

## Stage 2 (separate owner decision, per the precedent brief)

Archive.org CDX on apis.uk (and any hostname stage 1 shows real traffic to) to
learn what it used to serve; only then consider a minimal intent-probe page
(one invited action, stated-intent logging) instead of the bare 404 — the full
relojistas method. Serving ANYTHING beyond 404 is owner-gated.

## Privacy / scope rails

- Passive stage logs only what Cloudflare already forwards; no cookies, no
  fingerprinting, 30-day retention, stays on the island box.
- We observe requests to OUR domain; no recovery of anyone's old gated content
  (rail carried over from the precedent brief).

## Cost / dependencies

- £0 (reuses island + tunnel). Needs: 2 DNS records (owner or an API token),
  ~20 lines of Caddyfile + cloudflared config (this session can apply on
  request), logrotate stanza.
- No platform code, no cluster involvement — entirely island-side.

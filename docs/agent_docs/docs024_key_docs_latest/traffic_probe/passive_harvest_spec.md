# P4 — passive harvest (access log + visit counts): spec & decision

Two numbers the structured `/events` stream does NOT carry, both needed to
complete P4:

1. **Visit counts** (the events-per-1k denominator) — live in the engine's
   `counters.json`, exposed at `/stats`. Cheap to collect: the collector already
   authenticates to the box; add a `/stats` pull alongside `/events`.
2. **Passive signals** — external referer, landing path, the forum-404 intent
   paths, user-agent (for bot classification). These are in nginx's **combined
   access log** on the box (confirmed: setup.sh sets no custom log_format, so
   `$http_referer` + `$http_user_agent` + the request line are all logged). The
   engine does NOT see them on a static page load.

## Decision needed: how does the chassis read the access log?

**Option A — engine reads its own box's nginx log, exposes a key-gated digest.**
Add `GET /access-digest?since=` to site-engine: it tails
`/var/log/nginx/<domain>.access.log`, parses combined-format lines, returns
structured JSON (top referers, top 404 paths, UA buckets: known-good-bot /
known-bad-bot / browser, request counts). setup.sh grants the `site-engine`
user read access to the logs (add it to the `adm` group, or a read ACL on the
log dir). The collector pulls it like `/events`.
- Pro: keeps the "cluster pulls over HTTPS" model; no new infra; one consistent
  auth path; parsing/classification logic lives in Go next to the engine.
- Con: couples the engine to nginx's log path + a permission grant; the engine
  grows a second responsibility (it's no longer purely the capture API).

**Option B — defer to the P5 vmhost adapter (ssh).** The adapter reads the log
over ssh during its lifecycle passes and ships a digest.
- Pro: keeps the engine single-purpose; ssh access already part of P5.
- Con: passive signals wait for P5; needs the adapter built first.

**Option C — Cloudflare analytics (only if a box is proxied).** Use CF's
logs/analytics API for referer/UA/country.
- Pro: no box-side work; CF already classifies bots.
- Con: only for proxied boxes (relojistas is DNS-only today); per-zone API
  wiring; another credential.

**Recommendation:** Option A for the **visit counts now** (trivial — extend the
collector to pull `/stats` into a small `intent_site_stats` table), and Option A
for the **access-log digest** as the next engine increment, because it preserves
the pull model and needs no new infra; fall back to C opportunistically if/when
a box goes proxied. B only if we decide the engine must stay capture-only.

## If Option A (sketch, to build next)
- Engine: `GET /stats` already exists (visits+events per host). Add
  `GET /access-digest?since=` (key-gated, NDJSON or JSON summary). setup.sh:
  `usermod -aG adm site-engine` (or `setfacl` on the log dir) + confirm the
  per-domain `access_log` path.
- New table `intent_site_stats (id, site_id, host, visits, events, observed_at)`
  — the collector upserts the latest `/stats` each run; ranking query 1 then
  joins it for true events-per-1k.
- Collector: one more action step (or extend `collect_intent_events`) to pull
  `/stats` (+ `/access-digest`) and write `intent_site_stats` (+ a referer/404/UA
  rollup table if we structure the digest).
- Discovery: the bot-IP blocklist idea (handoff Thread D) consumes the same UA/
  IP rollup — build the digest with that in mind.

## What's NOT blocked
The ranking queries (`intent_ranking_queries.sql`) work today on absolute
signal (volume, distinct terms, dominant-cluster share, recency, referer,
landing-query). Only the per-1k RATE waits on the visit-count pull above.

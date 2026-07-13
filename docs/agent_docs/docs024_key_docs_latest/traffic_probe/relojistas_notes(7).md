# relojistas.com — domain notes (provenance, decisions, direction)

Living file for this domain. Append; date entries. Sister docs: the project
plan/runbook/running-notes and `relojistas_golive.md`.

## Provenance
- **What it was:** Spanish-language watch FORUM ("Relojistas — Foro de
  relojes", vBulletin-style). Boards from the saved snapshot: Foro general,
  Presentación de nuevos foreros, Ferias de relojería, Marcas de relojes, plus
  a marketplace cluster (Tiendas profesionales, Ventas, Outlet, Classifieds)
  and a project thread ("Proyecto Bauhaus RLJ02"). Login/registration present
  → some content was member-gated; residual visitors plausibly want threads,
  brands, repair advice, and buy/sell.
- **Evidence:** snapshot saved in the project (`Relojistas_-_Foro_de_relojes.html`
  + asset dirs under `docs.../traffic_probe/archive.org.results/relojistas/`).
- **Claimed traffic:** marketplace export estimated ~1,201,799 visits — a big
  outlier; treated as UNVERIFIED until our own logs speak (see Traffic).
- **Status:** parked → being repointed to the probe box (parking revenue stops
  at repoint; accepted deliberately).
- **Coordinates (go-live 2026-06-12):** box = Hetzner CPX22 (x86, 2 vCPU,
  4 GB, 80 GB, 20 TB incl.) #140056673 `ubuntu-4gb-nbg1-1-relojistas`, nbg1
  (Nuremberg, EU), €11.39/mo (cheapest available at order time; CX23 not
  offered) · **IP 167.233.33.159** (IPv6 2a01:4f8:1c1a:150b::/64 — unused v1;
  no AAAA record) · registrar: ___ · DNS: ___ (DNS-only first; CF proxied is
  a later option) · stats key: stored ___ ·
  repos: gqls/vm-sites (content), gqls/site-engine (engine) · webroot:
  `/var/www/vm-sites/relojistas.com` · data: `/var/lib/site-engine/` (counters.json + daily events-YYYYMMDD.jsonl).

## Decisions (dated)
- 2026-06-11 — Probe style: Spanish, `kind=search`, single text input; action
  label spans marca/modelo/reparación/compraventa (matches the forum+market
  heritage). Thanks page `/gracias.html` (engine `THANKS_PATH`).
- 2026-06-11 — **No results page in v1.** The probe performs no search and
  shows nothing back; the submission itself is the product (see "What we
  record"). Revisit if v1 data shows visitors clearly expect an answer
  (e.g. repeated re-submissions of the same term in one visit).
- 2026-06-11 — v1 is apex-only (`relojistas.com`); `www` is a follow-up
  (DNS + server_name/cert extension).
- 2026-06-11 — Store write-amplification fixed BEFORE launch (structural over
  quick): visits debounced via dirty-flag + 5s flusher + SIGTERM flush; events
  persist immediately; compact JSON. Burst-tested 80 visits + 1 event.
- 2026-06-11 — Deploy Actions aligned to the LIVE sites-repo sibling:
  self-hosted runner, dotted-domain regex, full-sync fallback, secret checks.
- 2026-06-11 — **Store v2 for high traffic:** events now append to daily JSONL
  files (one line per submission, O(1) at any volume, nothing held in RAM);
  /stats counters live in a small debounced snapshot. Burst-tested 300 events
  + 100 visits. Env var renamed `ENGINE_DB_PATH` → `ENGINE_DATA_DIR`.
- 2026-06-11 — **Dedicated VM for this domain** (its own box, not the shared
  multi-vhost one): Hetzner CX22-class (x86, 2 vCPU, 4 GB, 40 GB) — sized by
  disk/logs headroom, not CPU; static nginx + O(1) appends leave huge margin
  even if the 1.2M/month claim is real. NOTE: stay on x86 — the engine Action
  builds GOARCH=amd64 (Arm CAX boxes would need a build change).
- 2026-06-11 — **No third "collector" VM.** The serving box buffers (JSONL);
  the cluster PULLS over key-gated HTTPS on a schedule into clients_db
  (P4). Pull keeps all credentials in the cluster — the box never holds DB or
  cluster secrets; a push model or a middle VM would invert that and add a
  hop to secure for no gain. B2 stays available as optional cold backup.
- 2026-06-11 — **Retention enforced on-box:** setup.sh now installs
  `site-engine-prune.timer` (daily; deletes `events-*.jsonl` older than
  `RETENTION_DAYS`, default 90). No logrotate for engine files — daily JSONL IS
  the rotation, and logrotate's move/truncate would race the open handle.
  nginx logs keep their existing size-based logrotate.
- 2026-06-11 — **Imagery:** v1 ships text-only (cleanest signal, fastest).
  Decision posture: do NOT use manufacturer/press photos of current or
  upcoming models (rights, plus a parked probe showing product photos starts
  to imply a shop). v1.1 option: ONE brand-free generated hero (macro watch
  movement / wrist shot) via the chassis image-generator, rsynced alongside.
  v2 idea worth testing: a "novedades" strip rendered as CATEGORY BUTTONS
  (clicking a brand/model IS an intent event — `kind=categories`), which turns
  the "latest models" desire into measurement rather than decoration. Caveat
  logged: any displayed list ANCHORS what visitors search for — run it as an
  A/B against the plain box, not a replacement, and read top-terms before
  choosing the button set.
- 2026-06-11 — **Input hygiene at the source:** engine now strips control
  characters and caps values by RUNES (multibyte-safe); control-only
  submissions are dropped. Ingest-side validation contract lives in the plan
  (P4): parameterised SQL only, shape/enum/length checks, burst dedupe,
  escape-on-display, PII-redaction open choice.
- 2026-06-11 — **Hosting:** dedicated EU box confirmed sensible — Hetzner CX23
  (2 vCPU/4 GB/40 GB, ~€3.49/mo) includes 20 TB traffic in EU; overage €1/TB.
  Even the claimed 1.2M visits/mo ≈ 360 GB at ~300 KB/page — under 2% of the
  allowance. Avoid US/Singapore locations (slashed allowances; SIN overage
  €7.40/TB). Sharing the idea.uk box was considered and declined: capacity
  would be fine, but setup.sh has box-takeover semantics (`ufw --force reset`,
  removes nginx default site) and coupling an unknown-traffic experiment to a
  live product saves ~€3.49/mo — not worth it.

## Open choices
- **Cloudflare proxied mode** (runbook §8): gains `CF-IPCountry` (country field
  currently empty on DNS-only) + instant relocation; costs cache rules +
  real-IP config. Lean: switch after first week's baseline.
- **Categories variant** of the probe (forum boards suggest natural categories:
  marcas / reparación / compraventa / ferias) — needs the {{range}} component
  work; only if search terms cluster poorly.
- **Retention period** for free-text values (suggest 90 days, pruned at
  collection) — confirm before P4 collector lands.
- **Graduation criteria** (probe → real build): proposal — sustained
  events-per-1k ≥ 20 AND a dominant intent cluster covering ≥30% of terms over
  2–4 weeks. Tune once data exists.

## How we see what's submitted (the operator's question)
There is no dashboard yet; three ways now, one planned:
1. **Summary (no terms):** counts + events-per-1k, key-gated:
   `curl -s -H "X-Internal-Key: $KEY" https://relojistas.com/stats`
2. **The terms themselves** live in the on-box store. Last 20 submissions:
   ```bash
   ssh root@BOX "tail -20 /var/lib/site-engine/events-$(date -u +%Y%m%d).jsonl" \
     | jq -r '[.created_at,.kind,.value,.ref_host,.country] | @tsv'
   ```
   Top terms (rough cluster view):
   ```bash
   ssh root@BOX "cat /var/lib/site-engine/events-*.jsonl" \
     | jq -r 'select(.host=="relojistas.com").value' \
     | tr 'A-ZÁÉÍÓÚÜÑ' 'a-záéíóúüñ' | sort | uniq -c | sort -rn | head -25
   ```
3. **Ground-truth traffic** (independent of the beacon): nginx access log —
   `ssh root@BOX "grep -c ' /api/hit' /var/log/nginx/access.log"` and total
   request volume per day.
4. **`GET /events?since=` — BUILT (2026-06-12), live after the next engine
   deploy** (the 3.9 engine-seam push ships it; the WWW_ALIAS setup.sh re-run
   installs its nginx location). Key-gated NDJSON, `_meta` checkpoint line:
   `curl -sS -H "X-Internal-Key: $KEY" "https://relojistas.com/events?since=<RFC3339>"`.
   The P4 collector pulls this into an `intent_events` table on a schedule;
   viewing then becomes SQL / a small report.

## What we record (and the "result" question)
Per submission, one event: `id, host, kind, value (the typed text, ≤500 chars),
ref_host (referer reduced to bare host, blank if same-site), country (coarse
header if a CDN supplies it, else empty), created_at (UTC)`. Plus a per-host
visit counter from the 1×1 beacon (the events-per-1k denominator; the gracias
page deliberately carries no beacon so submissions don't inflate it).
**"Are we recording the result?"** — there is no result: the probe runs no
search and returns nothing; it 303s to /gracias.html. The submission IS the
recorded artifact. Deliberately NOT recorded: IP addresses, user agents,
cookies (none set), full referer URLs, names/emails.

## Traffic handling ("might be substantial — or not")
Posture: measure first, scale on evidence.
- **Day 1–2:** nginx logs + /stats give the real volume; judge the 1.2M claim.
- **Current capacity (post-fix):** nginx static serving is far beyond anything
  this domain can produce on a small VM; `/intent` is rate-limited (10 r/s per
  IP, burst 20); beacon hits are memory-only between 5s flushes; an event
  write rewrites the store file once per event.
- **Mitigation ladder (apply on evidence, in order):**
  1. Events file growing large (>~20 MB) or sustained >~5 events/s → move the
     event store to append-only JSONL + periodic visit snapshot (O(1) writes),
     and add rotation tied to the P4 collector (pull + truncate).
  2. Country/relocation/edge-cache needs → Cloudflare proxied mode (§8).
  3. Box saturation from this one domain → move it to its own box (runbook §4
     relocation; instant if proxied).
- **What we will NOT do under load:** add client-side JS, third-party
  analytics, or IP logging — the privacy posture holds regardless of volume.

## Log
- 2026-06-12 ~14:00 UTC — **3.6 confirmed:** env file holds the real key
  (`819419…6a7`, full value in the box env only) + THANKS_PATH=/gracias.html.
  **/stats verified over HTTPS:** visits 4, events 1, 250/1k (all operator's
  own tests; zero organic so far).
- 2026-06-12 — **Traffic claim assessment (early):** 1.2M/mo ⇒ ~0.46 visits/s
  ⇒ ~1,600/hour expected; observed organic ≈ 0 in the first hour. Signal
  strongly AGAINST the claim, but not yet concluded — three confounds:
  (1) DNS propagation (old-TTL resolvers still feed the parking IP, allow
  24–48h); (2) the beacon counts humans-with-browsers only — the claim, if it
  measured anything, likely counted bot-heavy REQUESTS → nginx access.log is
  the ground truth for that comparison; (3) the www gap — forum-era links
  likely used www., which has NO record → that slice dies invisibly.
  **Verdict criterion:** after 48h, read (a) access.log requests/hour split by
  user-agent, (b) beacon visits/day, (c) with www A record added +
  `WWW_ALIAS=true` re-run, the www share. If access.log < ~100 human-looking
  requests/day, the claim is dead at every level.
  Settle-it commands:
  ```bash
  ssh root@BOX "wc -l /var/log/nginx/access.log; grep -c ' /api/hit' /var/log/nginx/access.log"
  ssh root@BOX "awk '{print \$4}' /var/log/nginx/access.log | cut -d: -f1-2 | sort | uniq -c | tail -12"   # per-hour
  ssh root@BOX "awk -F'\"' '{print \$6}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head"     # UA mix
  dig +short www.relojistas.com   # the invisible-traffic check
  ```
- 2026-06-12 — **WWW_ALIAS added to setup.sh** (opt-in, default false → v1
  apex-only unchanged): serves www.<domain> in every vhost and requests the
  www cert SAN when www DNS resolves (pre-flight `getent`; apex-only with a
  log line otherwise — a later re-run upgrades). Enabling for relojistas =
  add www A record → `WWW_ALIAS=true DOMAINS=… re-run`. Rendered + nginx -t
  in both modes.
- 2026-06-12 13:03:44 UTC — **FIRST LIVE CAPTURE** (one minute after cert
  issuance): `kind=search`, value "correa Omega Seamaster - aa", ref_host
  empty (direct), country empty (expected on DNS-only). Counters: visits 2,
  events 1 → 500/1k. Full HTTPS path proven: page → form → /intent →
  sanitise → JSONL + counters → redirect. OPEN ACTION: key state — a later
  session echoed an empty $KEY (session-scoped variable, not file state);
  confirm with `grep -E '^(INTERNAL_API_KEY|THANKS_PATH)=' /etc/site-engine/
  site-engine.env`, set for real if placeholder, then record the key HERE.
- 2026-06-12 13:02 UTC — **DNS repointed; cert ISSUED on idempotent re-run**
  (validator reached the box; expires 2026-09-10, certbot auto-renew timer in
  place). HTTPS live; engine active. Field finding: nginx 1.28 warned on the
  deprecated `listen ... http2` form — setup.sh generator now emits
  version-neutral `listen 443 ssl;` (http2 opt-in note in the conf). REMINDER:
  run 2 printed no env warning only because run 1 wrote a PLACEHOLDER —
  `grep INTERNAL_API_KEY /etc/site-engine/site-engine.env` and set the real
  key + THANKS_PATH=/gracias.html (3.6) before trusting /stats.
- 2026-06-12 12:32 UTC — **Box provisioned** (setup.sh full run): packages,
  service user, engine installed + **active**, placeholder env written
  (INTERNAL_API_KEY still to set — 3.6), unit + deploy hook + prune timer (90d)
  installed, nginx stage-1/2 OK, ufw/fail2ban/unattended-upgrades/sshd-harden
  applied. **Cert PENDING:** certbot 403 — the ACME validator fetched the
  challenge from **76.223.54.146** (the domain's CURRENT DNS target — parking),
  not our box (167.233.33.159). Root cause: registrar A record not yet
  repointed. setup.sh degraded to HTTP as designed; a re-run after DNS
  propagates upgrades to HTTPS (no binary param needed — installed engine is
  kept). Note: box runs Ubuntu "resolute" (newer than the 24.04 in the docs) —
  no observed issues; nginx 1.28.3, certbot 4.0.0.
- 2026-06-11 — File created. Go-live bundle ready (`relojistas_golive.md`,
  `relojistas-site/`); awaiting box + DNS.

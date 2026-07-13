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
- **Coordinates (fill in at go-live):** registrar: ___ · DNS: ___ (DNS-only
  first; CF proxied is a later option) · box IP: ___ · stats key: stored ___ ·
  repos: gqls/vm-sites (content), gqls/site-engine (engine) · webroot:
  `/var/www/vm-sites/relojistas.com` · store: `/var/lib/site-engine/intent_events.json`.

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
   ssh root@BOX "jq -r '.events[\"relojistas.com\"][-20:][] \
     | [.created_at,.kind,.value,.ref_host,.country] | @tsv' \
     /var/lib/site-engine/intent_events.json"
   ```
   Top terms (rough cluster view):
   ```bash
   ssh root@BOX "jq -r '.events[\"relojistas.com\"][].value' \
     /var/lib/site-engine/intent_events.json" \
     | tr 'A-ZÁÉÍÓÚÜÑ' 'a-záéíóúüñ' | sort | uniq -c | sort -rn | head -25
   ```
3. **Ground-truth traffic** (independent of the beacon): nginx access log —
   `ssh root@BOX "grep -c ' /api/hit' /var/log/nginx/access.log"` and total
   request volume per day.
4. **Planned (P4):** engine gains key-gated `GET /events?since=`; a scheduled
   chassis action pulls into an `intent_events` table; viewing becomes SQL /
   a small report. Until then, 1–3 above are the interface.

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
- 2026-06-11 — File created. Go-live bundle ready (`relojistas_golive.md`,
  `relojistas-site/`); awaiting box + DNS.

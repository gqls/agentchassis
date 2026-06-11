# Traffic-probe — runbook

Operational how-to for putting probe domains live and keeping them running.
Order follows the chassis docs' advice: do the manual VM path first (it teaches
the exact steps), then automate. Items needing a decision are marked **[PENDING]**.

---

## 0. What the probe is (one paragraph)
A probe domain is a normal chassis-built site whose page plausibly reflects the
domain's old vertical and offers ONE invited action (search box / category links /
free-text). The page, privacy line, and the capture form are **chassis build
outputs**. On the VM, nginx serves those static files and proxies the capture API
to a tiny Go engine that records a structured intent event (keyed by host) into a
JSON file. Captured data is collected off-box. Visitors' stated intent ranks
which domains are worth building out.

## 1. Components
- **Static site** (per domain): built by the existing pipeline, committed to git.
- **Engine** (`probe-go`): API-only Go binary. Endpoints `POST /intent`,
  `GET /api/hit` (1×1 visit beacon), `GET /stats` (key-gated), `GET /health`.
  stdlib only, file-based JSON store keyed by canonical host.
- **nginx**: TLS (Let's Encrypt) + serves static files from the web root +
  proxies `/intent`, `/api/hit`, `/stats`, `/health` to the engine.
- **systemd**: keep-alive for the engine.
- **Off-box collection**: periodic push of the JSON store to B2 (reuse the
  documented checkpoint pattern) or pull into `clients_db`.

## 2. Engine configuration (env)
| Env | Purpose | Default |
|---|---|---|
| `PORT` | engine listen port (nginx proxies to it) | `8080` |
| `PROBE_DB_PATH` | JSON store path | `probe_events.json` |
| `INTERNAL_API_KEY` | gates `/stats` (sent as `X-Internal-Key`) | _(unset → /stats 401)_ |
| `ACCEPT_HOSTS` | comma list of canonical hosts this VM serves; empty = accept any | _(empty)_ |
| `THANKS_PATH` | where `/intent` 303-redirects on success (a static page on the site) | `/` |
| `GEO_COUNTRY_HEADER` | coarse country header if a CDN sets one | `CF-IPCountry` |
| `MAX_VALUE_LEN` | cap on a captured value's length | `500` |
| `ALLOWED_ORIGINS` | CORS allow-list (same-origin forms don't need it) | _(empty)_ |

The page must include the visit beacon: `<img src="/api/hit" width="1" height="1" alt="">`.

## 3. Manual go-live for the first domains (Path A)
Start with 3–5 clearly-generic domains you fully control. Suggested first set:
`relojistas.com`, `wayfaringlondoner.com`, `surgerylight.com`, plus one finance
tool and one clear retail.

1. **Confirm the vertical via Wayback** (CDX path list + a snapshot or two) so the
   page is plausible and the invited action fits. Local snapshots already in the
   project for `relojistas.com` and `wayfaringlondoner.com`.
2. **Build the page through the chassis** so it is a first-class site (research →
   plan → write → design → assemble → commit). **[PENDING]** which workflow/entry
   (see plan, Decision 1) and which repo (Decision 2).
3. **Provision a small VM** (1 vCPU, 512MB–1GB is plenty; engine is I/O-bound).
4. **Install the engine + nginx + certbot + systemd** (adapt idea.uk `setup.sh`;
   nginx serves the web root and proxies the four API paths to the engine).
5. **Point DNS** (A record → box). Certbot issues the cert. Repointing DNS stops
   any parking revenue — choose deliberately.
6. **Set engine env** (table above); set `ACCEPT_HOSTS` to the domains on this box.
7. **Walk one capture end-to-end**: load page → submit the invited action → 303 to
   thanks → confirm the event in the JSON store and in `/stats`.

> Capture the exact steps you run here as `setup.sh` + nginx conf + systemd unit.
> That artefact is most of the future automation.

## 4. Multi-domain on one box
One engine binary, nginx `server_name` block per domain (all proxying to the
engine), the page chosen by static web root per host; store keys events by host.
`ACCEPT_HOSTS` lists the hosts this box serves. Moving a busy domain to another
box = move its web root + set DNS + adjust `ACCEPT_HOSTS` on both boxes (and, once
it exists, update the domain→VM registry).

## 5. Deploy on update **[PENDING – Decision 3]**
Intended: keep "commit is deploy" — the site's repo Action ships the rebuilt
static files (and, when changed, the engine binary) to the box and restarts the
service. Mechanism (per-repo Action vs chassis-driven) is under decision.

## 6. Data collection & retention
- Periodic off-box push of `probe_events.json` (B2 checkpoint pattern) or pull
  into `clients_db`.
- Retention: keep free-text only as long as needed; prune on the collection step.
- Metric: intent events per 1,000 visits (visits from the beacon and/or nginx
  access logs), plus the actual terms/categories, coarse referer and country.

## 7. Health / checks
- `GET /health` for liveness; systemd restarts on crash.
- Once sites are live, the chassis discovery/audit agents scan them like any site.

## Changelog
- 2026-06-10: runbook created; engine trimmed to API-only backend; first-domain
  sequence drafted with build-workflow/repo/deploy decisions pending.

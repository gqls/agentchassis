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
- **Engine** (`site-engine`): API-only Go binary. Endpoints `POST /intent`,
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
| `ENGINE_DB_PATH` | JSON store path | `intent_events.json` |
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
   plan → write → design → assemble → commit), via the `build-dispatch-loop`
   pipeline. The site's `github_repo` is set to the **`vm-sites` repo** (separate
   shared repo, domain-subpath layout) so its Action deploys to the VM, not B2.
3. **Provision a small VM** (1 vCPU, 512MB–1GB is plenty; engine is I/O-bound).
4. **Run `setup.sh`** to install engine + nginx + certbot + systemd:
   `DOMAINS="a.com b.com" LETSENCRYPT_EMAIL=you@real.tld DEPLOY_USER=deploy ENGINE_BINARY_PATH=/tmp/site-engine bash setup.sh`
   (scp the engine binary to `/tmp/site-engine` first). `DEPLOY_USER` makes the web
   roots writable by the CI deploy user and installs the engine-update hook.
5. **Point DNS**: either an A record → box (certbot issues the cert), or a
   Cloudflare proxied record → box (see §8; Cloudflare then handles TLS). Either
   way, repointing DNS stops any parking revenue — choose deliberately.
6. **Set engine env** (`/etc/site-engine/site-engine.env`, see table above); set
   `INTERNAL_API_KEY`. `ACCEPT_HOSTS` is optional (nginx already gates by vhost).
7. **Walk one capture end-to-end**: load page → submit the invited action → 303 to
   thanks → confirm the event in the JSON store and in `/stats`.

> Capture the exact steps you run here as `setup.sh` + nginx conf + systemd unit.
> That artefact is most of the future automation.

## 4. Multi-domain on one box, and onboarding a new domain
One engine binary; one nginx `server_name` block per domain (each serves that
domain's static web root and proxies the four API paths to the engine); the store
keys events by host.

**Onboarding a new domain (vhost + cert) — the one-time step the content Action
does NOT do:**
1. **DNS first**: point the domain at the box (A record, or Cloudflare proxied
   record → box) and let it resolve, so the ACME http-01 challenge can succeed.
2. **Add it to `DOMAINS` and re-run `setup.sh` (MODE=full)** — idempotent:
   existing domains are untouched; the new one gets a web root
   (`/var/www/vm-sites/<domain>` + placeholder), an nginx `server_name` block, and a
   webroot certbot cert. If certbot can't issue yet (DNS not propagated), the
   domain stays on HTTP and a later re-run upgrades it to HTTPS.
3. **Deploy its content**: once the domain's folder lands in the `vm-sites` repo,
   the content Action rsyncs it into the web root (or rsync once manually for the
   first cut).
4. **Verify**: `curl -sS https://<domain>/health` and walk one capture.

**Relocating a busy domain to another box**: move its web root + add it to the new
box's `DOMAINS` (re-run `setup.sh`) + repoint DNS (instant if Cloudflare proxied;
TTL-bound on a plain A record), then drop it from the old box's `DOMAINS`. Once a
domain→VM registry exists, update it too.

## 5. Deploy on update
"Commit is deploy", target swapped. Two separate workflows:
- **Content** (`vm-sites` repo — `deploy-to-vm.yml`): on push to `sites/**`, rsync
  each changed `sites/<domain>/` over SSH into `/var/www/vm-sites/<domain>`
  (`rsync -az --delete`, hosted runner, mirrors `deploy-to-b2.yml`). No engine
  restart, no Cloudflare purge. Secrets: `VM_HOST`, `VM_USER`, `VM_SSH_KEY`. The
  deploy user must own the web roots — run `setup.sh` with
  `WEBROOT_OWNER="$VM_USER:www-data"`. Deploys CONTENT for already-provisioned
  domains only.
- **Engine** (site-engine repo — `deploy-engine-to-vm.yml`): on push to `**.go`/
  `go.mod`, build `linux/amd64` (static, stripped) → scp to box → run the narrow
  sudo hook `site-engine-deploy` (installed by `setup.sh` when `DEPLOY_USER` is
  set) which atomically swaps the binary and restarts. Same `VM_*` secrets. The
  swapped binary runs as the unprivileged `site-engine` user.

New-domain provisioning is §4 above (extend `DOMAINS`, re-run `setup.sh`); the
content Action does not provision vhosts. The terminal build item stays
target-agnostic — it commits to the site's repo; the repo's Action does the
VM-specific work.

## 6. Data collection & retention
- Periodic off-box push of `intent_events.json` (B2 checkpoint pattern) or pull
  into `clients_db`.
- Retention: keep free-text only as long as needed; prune on the collection step.
- Metric: intent events per 1,000 visits (visits from the beacon and/or nginx
  access logs), plus the actual terms/categories, coarse referer and country.

## 7. Health / checks
- `GET /health` for liveness; systemd restarts on crash.
- Once sites are live, the chassis discovery/audit agents scan them like any site.

## 8. Optional: Cloudflare in front of the box
Keep DNS on Cloudflare with a **proxied** record (orange cloud) → VM IP. This is a
reverse proxy to the VM as origin — NOT a second Worker, and NOT a second copy of
the content, so there is nothing to keep "in sync"; the VM stays the single source
of truth and Cloudflare just caches with a TTL. (A Worker would only make sense if
it served from a copy, which reintroduces the sync problem — don't.) Adjustments:
- **Cache rule:** bypass cache for `/intent`, `/api/hit`, `/stats`, `/health`;
  cache the static paths. Invalidate on deploy with a per-host purge (the same
  step the B2 Action already uses) if you cache aggressively.
- **Real client IP:** set nginx `set_real_ip_from <CF ranges>` + `real_ip_header
  CF-Connecting-IP`, else the rate-limit zone throttles all of Cloudflare's IPs
  as one. (No IP is stored regardless.)
- **TLS:** Full (strict) — keep certbot on the origin, or use a Cloudflare Origin
  Certificate (which lets you drop per-domain certbot).
- **Bonus:** Cloudflare sets `CF-IPCountry`, which the engine already reads, so
  country data comes for free; and relocating a domain is instant (change the
  proxied record's origin IP) instead of waiting on DNS TTL.

## Changelog
- 2026-06-10: runbook created; engine trimmed to API-only backend; first-domain
  sequence drafted with build-workflow/repo/deploy decisions pending.
- 2026-06-10: decisions resolved — dispatch-loop pipeline, separate `vm-sites` repo
  with its own VM-deploy Action, light per-repo Action, target-agnostic terminal
  item, one-time VM setup separate.
- 2026-06-10: added explicit new-domain onboarding (§4); engine-deploy workflow +
  `setup.sh` `DEPLOY_USER`/engine-update hook; Cloudflare-in-front option (§8).
- 2026-06-11: class-level rename applied (operator-confirmed): repos `vm-sites`
  (content) + `site-engine` (engine); box artifacts neutralised — service/user
  `site-engine`, `/opt/site-engine`, `/var/lib/site-engine`,
  `/etc/site-engine/site-engine.env`, webroots `/var/www/vm-sites/<domain>`,
  nginx conf `vm-sites.conf`, zone `engine_rl`, hook `site-engine-deploy`; env
  var `PROBE_DB_PATH` → `ENGINE_DB_PATH`, store file → `intent_events.json`.
  intent-probe component inserted into the live library (idempotent re-run
  confirmed).

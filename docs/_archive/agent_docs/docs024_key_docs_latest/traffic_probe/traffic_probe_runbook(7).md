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
|  `ENGINE_DATA_DIR` | data dir (counters.json + daily events-YYYYMMDD.jsonl) | `/var/lib/site-engine` (JSONL + counters) |
| `INTERNAL_API_KEY` | gates `/stats` (sent as `X-Internal-Key`) | _(unset → /stats 401)_ |
| `ACCEPT_HOSTS` | comma list of canonical hosts this VM serves; empty = accept any | _(empty)_ |
| `THANKS_PATH` | where `/intent` 303-redirects on success (a static page on the site) | `/` |
| `GEO_COUNTRY_HEADER` | coarse country header if a CDN sets one | `CF-IPCountry` |
| `MAX_VALUE_LEN` | cap on a captured value's length | `500` |
| `ALLOWED_ORIGINS` | CORS allow-list (same-origin forms don't need it) | _(empty)_ |

The page must include the visit beacon: `<img src="/api/hit" width="1" height="1" alt="">`.

## 3. Manual go-live — full command walkthrough (Path A)
Start with 3–5 clearly-generic domains you fully control; for the first
(relojistas.com) the page is hand-made (`relojistas-site/`) so nothing blocks.
Pipeline-built pages (chassis → `vm-sites` repo, root-level domain folders)
take over once the resolveGitRepoName patch lands.

Set these once per box/domain and paste the blocks as-is:
```bash
export OWNER=gqls
export OUTPUTS=~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/traffic_probe/deploy_setup
export BOX=<box-ip>            # from step 3.2
export DOMAIN=relojistas.com
export EMAIL=you@your-real-domain.tld   # real address (Let's Encrypt)
```

**3.1 Repos + secrets (one-time, laptop).**
*Where the repos live:* `site-engine` and `vm-sites` are **standalone repos**
with their own remotes (`$OWNER/site-engine`, `$OWNER/vm-sites`). Working
checkouts sit as **siblings of agentchassis** (`~/projects/site-engine`,
`~/projects/vm-sites`) — never nested inside the agentchassis tree, since a
git repo inside a git repo confuses both and the chassis must not track engine
source. The copy under the docs tree (`$OUTPUTS`) is a **reference snapshot
for the record only**, not the working repo (same pattern as contextkit's
in-repo home). Create the repos BY HAND — the git-adapter auto-creates repos
as PUBLIC.

*File manifest* — the outputs ship the workflow YAMLs **flat** (the delivery
channel cannot include dot-directories), so creating `.github/workflows/` is
part of this step:

| repo | contents |
|---|---|
| `site-engine` | `go.mod`, `env.go`, `main.go`, `service.go`, `store.go`, `.github/workflows/deploy-engine-to-vm.yml` |
| `vm-sites` | `.github/workflows/deploy-to-vm.yml` (+ one folder per domain, added later) |

```bash
cd ~/projects
gh repo create $OWNER/site-engine --private --clone
cd site-engine
cp $OUTPUTS/site-engine/* .                 # the five Go/module files
# (go.mod is authored, not generated — no build step creates it. Copying the
#  delivered one is correct; `go mod init site-engine` would be equivalent.
#  There is deliberately no go.sum: the engine is stdlib-only, zero deps.)
mkdir -p .github/workflows
cp $OUTPUTS/vm-deploy/deploy-engine-to-vm.yml .github/workflows/
go vet ./... && go build -o /dev/null . && echo FILES-OK   # fails if any of the five is missing
git add -A && git commit -m "site-engine v1"
git branch -M main                        # local git may default to master;
git push -u origin main                   # the workflows trigger on main

cd ~/projects
gh repo create $OWNER/vm-sites --private --clone
cd vm-sites
mkdir -p .github/workflows
cp $OUTPUTS/vm-deploy/deploy-to-vm.yml .github/workflows/
git add -A && git commit -m "vm-sites scaffold"
git branch -M main
git push -u origin main

# Deploy key + secrets (VM_HOST is set after 3.2 when $BOX exists):
ssh-keygen -t ed25519 -f ~/projects/deploy_key -N "" -C "vm-deploy"
gh secret set VM_USER    -R $OWNER/site-engine -b "deploy"
gh secret set VM_SSH_KEY -R $OWNER/site-engine < ~/projects/deploy_key
gh secret set VM_USER    -R $OWNER/vm-sites    -b "deploy"
gh secret set VM_SSH_KEY -R $OWNER/vm-sites    < ~/projects/deploy_key
# after 3.2:  gh secret set VM_HOST -R $OWNER/site-engine -b "$BOX"
#             gh secret set VM_HOST -R $OWNER/vm-sites    -b "$BOX"
```
*UI alternative for the secrets:* repo → Settings → Secrets and variables →
Actions → **Secrets** tab → "New repository secret". All three go in as
**Repository secrets** (NOT Variables, NOT environment secrets — the workflows
read `${{ secrets.* }}`): `VM_HOST` = the box IP; `VM_USER` = `deploy`;
`VM_SSH_KEY` = the FULL contents of the private key file `~/projects/deploy_key`,
including the `-----BEGIN/END OPENSSH PRIVATE KEY-----` lines. Repeat on BOTH
repos.

**3.2 Provision the box** — Hetzner **CX23, EU location** (Falkenstein/
Nuremberg/Helsinki; 20 TB included, €1/TB over; avoid US/SIN allowances),
Ubuntu 24.04, your root SSH key. Note the IP → `$BOX`.

**3.3 DNS** — A record `$DOMAIN → $BOX`, DNS-only (grey cloud) so the http-01
challenge works; proxied mode is §8, switchable later. Repointing stops any
parking revenue — deliberate. v1 is **apex-only** (apex = the bare domain,
`relojistas.com` itself, no `www.` or other subdomain): only the apex gets a
DNS record, an nginx vhost, and a cert; `www.relojistas.com` simply won't
resolve until the www follow-up (extra A record + server_name/cert extension).
**Do DNS before 3.5** — certbot's challenge needs the name resolving to the
box; setup.sh degrades gracefully (HTTP only) and is idempotent, so if DNS
lags, just re-run it after propagation.

**3.4 Deploy user (on the box, as root):**
```bash
ssh root@$BOX
adduser --disabled-password --gecos "" deploy
install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
# paste deploy_key.pub then Ctrl-D:
cat > /home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys
exit
```

**3.5 Build engine, ship, provision (laptop → box):**
```bash
cd ~/projects/site-engine
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o site-engine .
scp site-engine root@$BOX:/tmp/site-engine
scp $OUTPUTS/vm-deploy/setup.sh root@$BOX:/tmp/
ssh root@$BOX "DOMAINS=\"$DOMAIN\" LETSENCRYPT_EMAIL=$EMAIL DEPLOY_USER=deploy \
  ENGINE_BINARY_PATH=/tmp/site-engine bash /tmp/setup.sh"
```
The parameters are ENV VARS prefixed to the command — `bash /tmp/setup.sh`
bare will refuse with the DOMAINS guard. If you're already shelled into the
box, the equivalent is:
```bash
DOMAINS="relojistas.com" LETSENCRYPT_EMAIL=you@real.tld DEPLOY_USER=deploy \
  ENGINE_BINARY_PATH=/tmp/site-engine bash /tmp/setup.sh
```
Domains may also be passed positionally (`bash /tmp/setup.sh relojistas.com`),
but the email/deploy-user/binary parameters are env-only.
setup.sh installs nginx vhost(s)+cert, the systemd unit, ufw/fail2ban/
logrotate, the deploy sudo hook, and the `site-engine-prune.timer`
(RETENTION_DAYS=90 default).

**3.6 Engine env (on the box):**
```bash
ssh root@$BOX
KEY=$(openssl rand -hex 24)
sed -i "s|^INTERNAL_API_KEY=.*|INTERNAL_API_KEY=$KEY|"  /etc/site-engine/site-engine.env
sed -i "s|^THANKS_PATH=.*|THANKS_PATH=/gracias.html|"   /etc/site-engine/site-engine.env
systemctl restart site-engine
echo "STATS KEY: $KEY"   # store it (domain notes file)
exit
```

**3.7 First content (laptop; the Action takes over after):**
```bash
rsync -az --delete relojistas-site/ deploy@$BOX:/var/www/vm-sites/$DOMAIN/
```

**3.8 Verify end-to-end:**
```bash
curl -sS https://$DOMAIN/health                       # {"ok":true}
# browser: https://$DOMAIN → submit a term → 303 to /gracias.html
curl -sS -H "X-Internal-Key: $KEY" https://$DOMAIN/stats
ssh root@$BOX "tail -3 /var/lib/site-engine/events-*.jsonl; cat /var/lib/site-engine/counters.json"
ssh root@$BOX "systemctl list-timers site-engine-prune.timer --no-pager"
```

**3.9 Prove the Action seams:**
```bash
cd site-engine && git commit --allow-empty -m "engine deploy test" && git push
# repo → Actions (self-hosted runner) → then:
ssh root@$BOX "journalctl -u site-engine -n 20 --no-pager"
# content seam: edit vm-sites/$DOMAIN/index.html, push, confirm the rsync ran.
```

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
- Periodic off-box push of `/var/lib/site-engine` (JSONL + counters) (B2 checkpoint pattern) or pull
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
  var `PROBE_DB_PATH` → `ENGINE_DATA_DIR`, store file → `/var/lib/site-engine` (JSONL + counters).
  intent-probe component inserted into the live library (idempotent re-run
  confirmed).

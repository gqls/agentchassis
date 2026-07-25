# RUNBOOK — the tools-api island (Route B1, built 2026-07-24)

The VM: Mythic Beasts `vds:toolsapisuk`, **toolsapisuk.vs.mythic-beasts.com**
(176.126.243.183), VPS 2 (1 core/2GB), 20GB SSD, IPv4, Ubuntu 26.04 LTS,
£16.20/mo ex-VAT incl. backup account + 20GB mirrored backup space.
Access: `ssh root@toolsapisuk.vs.mythic-beasts.com` (key-only).

## As built (2026-07-24, "bastion host" session — everything verified live)

- **Hardening**: `/etc/ssh/sshd_config.d/99-island-hardening.conf`
  (PasswordAuthentication no, PermitRootLogin prohibit-password,
  KbdInteractiveAuthentication no; `sshd -t` checked). ufw ACTIVE:
  default deny incoming, allow OpenSSH only; all outbound open (tunnel,
  Anthropic, apt all dial out). unattended-upgrades enabled.
- **Stack** at `/opt/island/` (docker.io 29.1.3 + compose v2.40.3, Ubuntu
  packages): `docker-compose.yml` + `Caddyfile` + `backup_pg.sh` — the copies
  in THIS repo directory are the source of truth; scp them over on change.
  Secrets in `/opt/island/.env` (root-only, mode 600, generated ON the box,
  never in the repo or transcripts): POSTGRES_PASSWORD (set),
  ANTHROPIC_API_KEY (EMPTY until the owner issues a dedicated spend-capped key).
- **Verified behaviour**: `curl 127.0.0.1:8081/` → 404 (allowlist working);
  `/api/v1/tools/ping` → 502 (nothing behind Caddy yet — correct until the
  engine lands); Postgres 16.14 answering as tools_api/tools_api, data in the
  `pgdata` volume; both containers restart unless-stopped.
- **Backups**: `/etc/cron.d/island-backup` → 02:17 nightly
  `/opt/island/backup_pg.sh` (pg_dump | gzip, 14-day local retention, then
  rsync mirror to the MB backup account
  `32950_toolsapisuk@backup-sov-a.mythic-beasts.com:tools-api-backups/`, 20GB,
  MB-mirrored to a second UK site). The backup host is **key-only**.
  **TODO(owner, last step)**: install the island's key in the MB control panel
  (Backup account → SSH keys): `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMhCipVo2VJ32WuzEZ2qjA/QZ3B7brCYiE1OFjF/Tkhm island-backup@toolsapisuk`
  Until then the rsync leg fails loudly (verified 2026-07-25: dump+retention
  fine, rsync `Permission denied (publickey)`) and self-heals once the key is in.
  Test after: `ssh root@island /opt/island/backup_pg.sh` → exit 0.
- **cloudflared** 2026.7.3 installed from Cloudflare's apt repo.
  `cloudflared tunnel login` launched; awaiting the owner clicking the
  authorisation URL (in /root/cf_login.log). Cert lands at
  /root/.cloudflared/cert.pem automatically.

## Tunnel — DONE 2026-07-24 (~16:14Z), verified from the public internet

- The owner's browser auth delivered cert.pem as a DOWNLOAD (the island's
  waiting `tunnel login` had already exited) → moved to island
  /root/.cloudflared/cert.pem (0600), local copy deleted. NOTE: the cert
  transited the session transcript (local-only); revoke+relogin in the
  dashboard if ever concerned.
- Tunnel `tools-api` id **f917c7c1-4dae-446f-a1e0-8f4c636cc345**; credentials
  /etc/cloudflared/tools-api.json (0600); config /etc/cloudflared/config.yml
  (hostname tools.apis.uk → http://localhost:8081, fallback 404);
  CNAME tools.apis.uk → tunnel added by `route dns`; systemd service
  installed, `systemctl is-active` = active.
- VERIFIED from outside: `https://tools.apis.uk/` → **404 from our Caddy**
  (was Cloudflare 525 this morning); `/api/v1/tools/ping` → **502** (no engine
  yet — correct). A random other subdomain now fails to RESOLVE — the dead `*`
  wildcard appears deleted (owner, presumably, while in the dashboard).

Cloudflare dashboard settings **applied by owner 2026-07-25** (SSL/TLS Full
(strict); Always Use HTTPS; rate rule on tools.apis.uk; Free Managed WAF).
Verified from outside: `http://tools.apis.uk/` → 301 to https; https → our 404.

## Traffic probe — features_open/020 stage 1, LIVE 2026-07-25

- DNS: apex `apis.uk` + `*.apis.uk` are proxied CNAMEs → the tunnel, added
  from the island via `cloudflared tunnel route dns tools-api <name>` (the
  zone cert at /root/.cloudflared/cert.pem authorises this — no dashboard).
- cloudflared ingress: apex + `*.apis.uk` → `http://localhost:8082` (rules sit
  between the tools.apis.uk rule and the 404 fallback).
- Caddy `:8082` vhost: responds 404 to everything, logs JSON per request to
  `/var/log/caddy/probe_access.log` = host bind-mount
  `/opt/island/logs/probe/probe_access.log`. Caddy's own rolling: 10MiB ×10,
  kept 720h (=30 days) — no host logrotate. Log includes Host, method, uri and
  all forwarded headers (Cf-Connecting-Ip, Cf-Ipcountry, Referer, User-Agent).
- Verified end-to-end from the public internet: apex/www/random-subdomain →
  404, log line carries CF-Connecting-IP + GB country.
- **Review ~2026-08-08**: rank hostname × path, e.g.
  `jq -r '.request | .host + " " + .uri' probe_access.log | sort | uniq -c | sort -rn | head -30`
  (cross-check against the zone's free Cloudflare analytics).
- Apex note: when the owner's bees homepage (separate thread) exists, repoint
  ONLY the apex record at its hosting; wildcard + probe stay as they are.

## ENGINE LANDED — 2026-07-25 (PR #3 merged 09:19Z, deployed + smoke-verified same day)

As built (supersedes the checklist that stood here):

1. **Image path decided: `docker save | gzip | ssh docker load`** — no registry
   creds on the island, keeping the no-production-credentials rule. Image
   `aqls/tools-api:v1.0.1162` (built from the 086 branch which carries the
   tools-api source + three post-merge fixes; build via `make build-tools-api-ref`).
2. **DB prep applied** (`/opt/island/island_db_prep.sql`, ledgered in the
   island's own `island_migrations` table): minimal `sites` table (id/domain/
   status — the merged CORS lookup + 198's FK need it; island has no platform
   schema) seeded with vonc.com's REAL cluster site id `9ec3b9ee-…`; migration
   198 applied **with a DO-block guard replacing the merged file's top-level
   ASSERT (not valid SQL — psql refuses the file; repo copy corrected too)**.
3. **compose**: tools-api block ACTIVE (PORT 8080 = Caddy upstream;
   DATABASE_URL to the island PG; GAUNTLET_MODEL claude-sonnet-5). The drafted
   `ALLOWED_ORIGINS` env never existed in the merged code — CORS reads the
   island `sites` table.
4. **ANTHROPIC_API_KEY: still the explicit placeholder** `PENDING-OWNER-SPEND-
   CAPPED-KEY` (the service refuses to start on an empty value). /round is fully
   live; /position + /defend return honest JSON 503 until the owner issues the
   dedicated key → put it in `/opt/island/.env`, `docker compose up -d tools-api`.
5. **Three defects found at first smoke, fixed, live** (commits 01e-style on
   branch 086: dockerfile golang 1.23→1.24; GetRound NULL-scan killed both LLM
   endpoints + 404-masked-500; LLM-failure 502→503 because **Cloudflare replaces
   an origin-502 body with its own error page** — JSON only survives as 503).
   Council corr `64e6112c` (advisory, alongside-commit).
6. **Public smoke matrix (2026-07-25, from the internet)**: /round 200 + real
   provocation + persisted row · real-round /position JSON 503 · missing-round
   404 · denied-origin 403 · preflight 204 · non-allowlisted path 404.
7. NEXT: owner key → real /position + /defend round-trip → re-fire the
   experience plan with full liveness evidence (the parked 092).
   **FLAG:** vonc.com/data/provocations.json carries FABRICATED stats
   ("1,284 Positions Filed", "62% Disagree") which /round passes through in
   `provocation.stats` — P4 front-end must not render them and the data file
   needs regenerating (real counts can come from gauntlet_rounds once traffic
   exists).

## Standing facts

- Public path: browser → Cloudflare (tools.apis.uk) → tunnel (outbound-only;
  ufw has NO public inbound except SSH) → Caddy :8081 (path allowlist + 1MB
  cap) → tools-api container → island Postgres. The production cluster appears
  NOWHERE in this path — that isolation is the point of Route B1.
- NOTHING on this box holds any production credential: no kubeconfig, no
  cluster DSN, no platform Anthropic key, no Kafka. Keep it that way.
- Upgrade path (B2, framework island): rent a bigger VM, k3s + 1-broker Kafka +
  core services; move = pg_dump/restore + tunnel credentials file. See
  SUMMARY_2026-07-24c.

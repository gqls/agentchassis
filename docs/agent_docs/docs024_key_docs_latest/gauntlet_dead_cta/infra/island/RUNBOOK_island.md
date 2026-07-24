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
  `/opt/island/backup_pg.sh` (pg_dump | gzip, 14-day local retention).
  First dump verified. **TODO(owner)**: backup-space host + username from the
  MB control panel, install the island's SSH key there, uncomment the rsync
  line in backup_pg.sh — until then dumps are on-box only.
- **cloudflared** 2026.7.3 installed from Cloudflare's apt repo.
  `cloudflared tunnel login` launched; awaiting the owner clicking the
  authorisation URL (in /root/cf_login.log). Cert lands at
  /root/.cloudflared/cert.pem automatically.

## Once the cert lands (next session steps, in order)

```bash
ssh root@toolsapisuk.vs.mythic-beasts.com
cloudflared tunnel create tools-api          # note the UUID
cloudflared tunnel route dns tools-api tools.apis.uk
mkdir -p /etc/cloudflared
cp /root/.cloudflared/<UUID>.json /etc/cloudflared/tools-api.json
# config: ../cloudflared_config.yml from this repo dir → /etc/cloudflared/config.yml
#   (tunnel: tools-api, credentials-file: /etc/cloudflared/tools-api.json,
#    hostname tools.apis.uk → http://localhost:8081)
cloudflared service install
systemctl enable --now cloudflared
curl -s https://tools.apis.uk/            # expect 404 from OUR Caddy (not 525/530)
```
Cloudflare dashboard (apis.uk zone): DELETE the `*` wildcard record (dead-origin
525s); SSL/TLS → Full (strict); Always Use HTTPS; one rate-limiting rule on
`tools.apis.uk/*`; Free Managed WAF ruleset on.

## When the engine lands (feature-builder PR merged, image published)

1. Solve the image path: the island pulls from a registry or `docker save |
   ssh docker load` — decide at PR time ([OPEN] in NOTES).
2. Owner issues a DEDICATED spend-capped Anthropic key → `/opt/island/.env`.
3. Uncomment the tools-api block in docker-compose.yml (fix image + port; the
   Caddyfile's `tools-api:8080` upstream must match the PR's real port).
4. Apply migration 198 to the ISLAND's Postgres (NOT clients_db) and record it
   in a ledger note here.
5. `docker compose up -d` → `/api/v1/tools/*` goes 502 → live. Smoke-POST a
   round from outside; confirm a gauntlet_rounds row.
6. Re-fire the experience plan with this liveness evidence (the parked 092).

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

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
  **DONE (owner, 2026-07-25):** pubkey installed in the MB control panel.
  Verified: `ssh root@island /opt/island/backup_pg.sh` → exit 0 (was
  `Permission denied (publickey)` earlier the same day). Backup leg fully live.
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
4. **ANTHROPIC_API_KEY: LIVE 2026-07-25** — owner created a dedicated key
   (org-level spend limits only on this tier; per-key/Workspace budgets not
   pursued, not blocking) and installed it on the box themselves via SSH (key
   never transited any session transcript). `/opt/island/.env` holds the real
   value.
5. **Four defects found at first smoke, fixed, live** (commits on branch 086):
   dockerfile golang 1.23→1.24; GetRound NULL-scan killed both LLM endpoints +
   404-masked-500; LLM-failure 502→503 because **Cloudflare replaces an
   origin-502 body with its own error page**; and — found only once a REAL key
   was in place — `NewAnthropicClient` requires `config["api_key_env_var"]`
   naming the env var (no default), which both handlers omitted, so client
   creation failed on every call regardless of the key's validity (commit
   `76e9c44d2`, image v1.0.1163). Council corr `64e6112c` (advisory).
6. **Public smoke matrix (2026-07-25, from the internet)**: /round 200 + real
   provocation + persisted row · missing-round 404 · denied-origin 403 ·
   preflight 204 · non-allowlisted path 404.
7. **FULL REAL ROUND-TRIP VERIFIED LIVE 2026-07-25 ~15:00**: /round →
   /position (genuine AI counter-position + challenge) → /defend (genuine AI
   verdict + reasons), all persisted. Two complete rounds in the island DB,
   both `verdict->>'verdict' = 'opponent wins'` — honest judging, not a
   pushover. This is the liveness evidence for the experience re-plan.
8. **Backup pubkey CONFIRMED working** 2026-07-25 (owner pasted it into the MB
   panel): `backup_pg.sh` exits 0.
9. NEXT: carry the liveness evidence into the 197 compose-decisions channel,
   re-fire the experience plan (092).
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

## Verify a rebuild against the RUNNING CONTAINER — never the tag, never the commit

Added 2026-07-27 answering a council objection (debug_historian, corr `e004fd81`)
on the `bugs_open/083` fix. On this host the verify-against-the-pod rule still
applies, and it matters *more* than in the cluster: the island is built with
`docker compose build`, **nothing downstream records which commit the image came
from**, and a `compose up -d` that silently kept the old container looks
identical to a successful deploy.

> **CORRECTED 2026-07-27, before first use — the draft of this section had two
> defects, both of the exact class it warns about.**
> (a) It grepped **`/app/tools-api`**. The dockerfile
> (`build/docker/backend/tools-api.dockerfile`) does `COPY --from=builder
> /tools-api /tools-api` — the binary is at **`/tools-api`**, so every grep would
> have returned 0 and read as "the deploy failed".
> (b) Its negative control grepped for **`JSONError(c, 502`** — that is Go
> *source*, which is not in a compiled binary at all. It returns 0 before AND
> after, so it proved nothing. **A negative control must be a string that really
> was in the OLD binary, or a token that cannot exist anywhere.**

```bash
ISL=root@toolsapisuk.vs.mythic-beasts.com
X="docker compose exec -T tools-api sh -c"

# 1. POSITIVE — a symbol this build CREATED. Verified locally first: this
#    returns 4 on v1.0.1178 and 0 on v1.0.1163, so it genuinely separates them.
ssh $ISL "cd /opt/island && $X 'strings /tools-api | grep -c logInternalFailure'"   # expect > 0

# 2. NEGATIVE CONTROL — a token that exists in NO build. If this returns
#    non-zero the grep is matching everything and check 1 proved nothing.
ssh $ISL "cd /opt/island && $X 'strings /tools-api | grep -c logNeverExisted'"      # expect 0

# 3. BEHAVIOURAL — request logging did not exist in ANY previous build, so a
#    [GIN] line is itself proof the new binary is the one serving traffic.
ssh $ISL 'cd /opt/island && docker compose logs --since 5m tools-api | grep -c "\[GIN\]"'

# 4. Container identity, to catch "up -d quietly kept the old container".
ssh $ISL 'cd /opt/island && docker compose ps --format "{{.Name}}  {{.Image}}  {{.RunningFor}}"'
```

**Verify the image BEFORE shipping it**, too — a 40 MB transfer and a container
swap are a slow way to discover the binary was wrong:

```bash
docker run --rm --entrypoint sh aqls/tools-api:$TAG -c 'strings /tools-api | grep -c <symbol>'
```

**Pick the grep target from what your change CREATED**, not from a type name or a
comment: a comment is not in the binary, and a typed constant may be inlined away
— both make a pod-grep vacuous (this has burned two workstreams already).

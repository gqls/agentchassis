#!/usr/bin/env bash
#
# setup.sh — provision (or rebuild, or update) a VM to host backend sites (site-engine).
#
# Adapted from idea.uk's setup.sh. Same design goals, so it serves the manual
# path NOW and the chassis service-deployer LATER:
#   - NON-INTERACTIVE: every input is an env var with a default; no prompts.
#   - IDEMPOTENT: safe to re-run. Re-running IS the "rebuild" path. Adding a
#     domain = add it to DOMAINS and re-run (full); the conf is regenerated.
#   - PARAMETERISED: nothing host-specific is hard-coded.
#   - SELF-CONTAINED: writes the systemd unit + nginx conf inline.
#
# DIFFERENCE FROM idea.uk (multi-vhost): idea.uk was one domain proxying / to the
# engine. The box runs ONE site-engine (single systemd unit) behind MANY domains.
# Per domain, nginx SERVES the chassis-built static site from a web root and
# PROXIES ONLY the capture API paths (/intent, /api/hit, /stats, /health) to the
# engine. The static files are deployed separately (git -> Action rsync); this
# script only provisions the box, the web roots, nginx, TLS, and the engine.
#
# Run as root on a fresh Ubuntu 22.04/24.04 VM:
#   DOMAINS="relojistas.com surgerylight.com" LETSENCRYPT_EMAIL=you@real.tld \
#   ENGINE_BINARY_PATH=/tmp/site-engine bash setup.sh
#
# Modes:
#   MODE=full   (default) — full provision/rebuild: packages, user, web roots,
#               nginx, TLS, firewall, fail2ban, systemd, start.
#   MODE=update — only replace the engine binary and restart. (Per-domain CONTENT
#               updates come via the deploy Action's rsync, not this script.)
#
# The env file with any secrets (/etc/site-engine/site-engine.env) is NOT overwritten if it
# exists — copy site-engine.env.example there and fill it in (or have the deployer drop
# it over SSH).

set -euo pipefail

# ── parameters ────────────────────────────────────────────────────────────────
# Convenience: positional args become DOMAINS when the env var is unset, so
# both invocation forms work:
#   DOMAINS="a.com b.com" LETSENCRYPT_EMAIL=… bash setup.sh
#   LETSENCRYPT_EMAIL=… bash setup.sh a.com b.com
if [ -z "${DOMAINS:-}" ] && [ "$#" -ge 1 ]; then
  DOMAINS="$*"
fi
DOMAINS="${DOMAINS:?set DOMAINS env var or pass domains as arguments, e.g. bash setup.sh relojistas.com}"
LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:?set LETSENCRYPT_EMAIL for cert registration}"
if [[ "$LETSENCRYPT_EMAIL" == *@example.com || "$LETSENCRYPT_EMAIL" == *@example.org ]]; then
  echo "[setup] ERROR: LETSENCRYPT_EMAIL is a placeholder ($LETSENCRYPT_EMAIL). Use a real address." >&2
  exit 1
fi
MODE="${MODE:-full}"

SERVICE_USER="${SERVICE_USER:-site-engine}"
SERVICE_PORT="${SERVICE_PORT:-8080}"          # must match PORT in the env file
APP_DIR="${APP_DIR:-/opt/site-engine}"
DATA_DIR="${DATA_DIR:-/var/lib/site-engine}"
ENV_DIR="${ENV_DIR:-/etc/site-engine}"
ENV_FILE="${ENV_FILE:-$ENV_DIR/site-engine.env}"
ACME_WEBROOT="${ACME_WEBROOT:-/var/www/letsencrypt}"
WEBROOT_BASE="${WEBROOT_BASE:-/var/www/vm-sites}" # per-domain static roots: $WEBROOT_BASE/<domain>
# Optional deploy user that the GitHub Actions (content rsync + engine update) SSH
# in as. When set: the per-domain web roots default to being owned by it, and a
# narrow privileged engine-update hook + sudoers rule are installed (see below).
# setup.sh does NOT create this user or its SSH key — the operator does that
# during onboarding (see runbook). (NEW param vs the idea.uk original.)
DEPLOY_USER="${DEPLOY_USER:-}"
# Owner of the per-domain web roots. The deploy user must be able to WRITE here;
# nginx (www-data) only needs READ. Defaults to the deploy user when one is set,
# else the old www-data ownership. (NEW param vs the idea.uk original.)
WEBROOT_OWNER="${WEBROOT_OWNER:-${DEPLOY_USER:+$DEPLOY_USER:www-data}}"
WEBROOT_OWNER="${WEBROOT_OWNER:-www-data:www-data}"
WEBROOT_USER="${WEBROOT_OWNER%%:*}"
WEBROOT_GROUP="${WEBROOT_OWNER##*:}"

# Binary source — exactly one applies. URL takes precedence (the chassis path: a
# presigned B2 URL, as the Thunder adapter already does). Otherwise a local path
# already on the box (the manual path: scp the binary up first).
ENGINE_BINARY_URL="${ENGINE_BINARY_URL:-}"
ENGINE_BINARY_PATH="${ENGINE_BINARY_PATH:-/tmp/site-engine}"

HARDEN_SSH="${HARDEN_SSH:-true}"
WHITELIST_IPS="${WHITELIST_IPS:-}"
# Days to keep daily event files (events-YYYYMMDD.jsonl) before the prune
# timer deletes them. Collection (P4) must run well inside this window.
RETENTION_DAYS="${RETENTION_DAYS:-90}"
# WWW_ALIAS=true also serves www.<domain> (server_name on every vhost) and
# requests the www SAN on certs WHEN www DNS resolves. Off by default: v1 is
# apex-only. Enabling later = add the www A record, re-run with WWW_ALIAS=true.
WWW_ALIAS="${WWW_ALIAS:-false}"

log() { echo "[setup] $*"; }
have() { command -v "$1" >/dev/null 2>&1; }

# normalise DOMAINS: trim, lowercase, drop leading www. (canonical host form)
norm_domains() {
  local d out=""
  for d in $DOMAINS; do
    d="$(echo "$d" | tr '[:upper:]' '[:lower:]' | sed 's/^www\.//; s/[[:space:]]//g')"
    [[ -n "$d" ]] && out="$out $d"
  done
  echo "${out# }"
}
DOMAINS="$(norm_domains)"

# ── nginx fragments ─────────────────────────────────────────────────────────────
# http-context rate-limit preamble (canonical geo+map exemption when WHITELIST_IPS set).
rate_limit_preamble() {
  if [[ -n "$WHITELIST_IPS" ]]; then
    echo "geo \$rate_limited {"
    echo "    default 1;"
    local ip
    for ip in $WHITELIST_IPS; do echo "    $ip 0;"; done
    echo "}"
    echo "map \$rate_limited \$limit_key {"
    echo "    1 \$binary_remote_addr;"
    echo '    0 "";'
    echo "}"
    echo "limit_req_zone \$limit_key zone=engine_rl:10m rate=10r/s;"
  else
    echo "limit_req_zone \$binary_remote_addr zone=engine_rl:10m rate=10r/s;"
  fi
}

# The capture API locations proxied to the single engine. Reused in http + https
# server blocks. Exact-match locations so they win over the static `location /`.
api_locations() {
  cat <<EOF
    location = /intent {
        limit_req zone=engine_rl burst=20 nodelay;
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    location = /api/hit {
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
    location = /stats {
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
    }
    location = /events {
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
        proxy_buffering off;   # NDJSON export streams; don't spool large pulls
        proxy_read_timeout 120s;
    }
    location = /health {
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
    }
    # Search-that-answers (relojistas manifest §6 B): server-rendered results
    # page from the engine. Rate-limited like /intent — it is visitor-facing.
    location = /buscar {
        limit_req zone=engine_rl burst=20 nodelay;
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
EOF
}

# Legacy vBulletin feed URLs (relojistas P5.1, reconciled 2026-07-24 from the
# box's SURGICAL edit — this generator now owns it; a re-run must not lose it).
# The old forum served RSS at /external.php?type=RSS2 (+ board/classifieds
# variants); subscribers still pull those URLs, so every variant lands on the
# static /feed.xml the chassis publishes. Case-insensitive type match and the
# /ventas/ + bare-path variants close the 3 residual 404s measured 2026-07-19.
legacy_feed_locations() {
  cat <<'EOF'
    location = /external.php {
        if ($arg_type ~* ^rss2$) { rewrite ^ /feed.xml last; }
        # No type argument: old readers and hand-typed URLs — still the feed.
        rewrite ^ /feed.xml redirect;
    }
    location ~ ^/(ventas|forums)/external\.php$ {
        rewrite ^ /feed.xml redirect;
    }
EOF
}

# Dead vBulletin surface (relojistas, added 2026-07-28 from the first per-source crawl
# analysis — EVIDENCE_2026-07-28_crawl_budget_and_the_dead_forum.md).
#
# WHY 410 AND NOT 404: measured 208,529 requests for /attachment.php in 33 hours from
# 1,409 distinct IPs sweeping 25,030 attachment ids. A 404 means "gone for now, try
# later" — which is why they have been retrying for months. 410 means "gone, permanently"
# and is the only status that makes a crawler drop the URL.
#
# WHY NOT `Disallow:` THESE IN robots.txt INSTEAD: blocking them stops crawlers ever
# FETCHING them, and therefore ever seeing the 410 — so they stay indexed forever. Crawl
# first, be told gone, then de-index.
#
# NOTE /faq.php and /search.php are deliberately ABSENT from this list: both have honest
# live equivalents on the new site and are handled below rather than retired.
dead_forum_locations() {
  cat <<'EOF'
    location = /attachment.php   { return 410; }
    location = /showthread.php   { return 410; }
    location = /newreply.php     { return 410; }
    location = /printthread.php  { return 410; }
    location = /private.php      { return 410; }
    location = /sendmessage.php  { return 410; }
    location = /member.php       { return 410; }
    location = /cron.php         { return 410; }
    location ~ ^/(ventas|forums)/(attachment|showthread|newreply|printthread)\.php$ { return 410; }

    # The old forum search still has a live equivalent: the engine's own results page.
    # Same function, honestly mapped — not a doorway, the user asked to search and gets
    # a search. Query string is carried through by nginx on a rewrite.
    location = /search.php {
        if ($arg_q)          { rewrite ^ /buscar?q=$arg_q redirect; }
        if ($arg_query)      { rewrite ^ /buscar?q=$arg_query redirect; }
        rewrite ^ /buscar redirect;
    }
EOF
}

# Cloudflare real-ip (P0): restore the visitor's address from CF's header so
# access logs count subscribers, not edge nodes. IP list = published CF ranges
# (https://www.cloudflare.com/ips/, snapshot 2026-07-24 — refresh on re-run if
# stale). Emitted once at http{} level via the conf.d include the caller writes.
cloudflare_realip() {
  cat <<'EOF'
# Cloudflare origin-pull ranges — restore true client IP (managed by setup.sh)
set_real_ip_from 173.245.48.0/20;
set_real_ip_from 103.21.244.0/22;
set_real_ip_from 103.22.200.0/22;
set_real_ip_from 103.31.4.0/22;
set_real_ip_from 141.101.64.0/18;
set_real_ip_from 108.162.192.0/18;
set_real_ip_from 190.93.240.0/20;
set_real_ip_from 188.114.96.0/20;
set_real_ip_from 197.234.240.0/22;
set_real_ip_from 198.41.128.0/17;
set_real_ip_from 162.158.0.0/15;
set_real_ip_from 104.16.0.0/13;
set_real_ip_from 104.24.0.0/14;
set_real_ip_from 172.64.0.0/13;
set_real_ip_from 131.0.72.0/22;
set_real_ip_from 2400:cb00::/32;
set_real_ip_from 2606:4700::/32;
set_real_ip_from 2803:f800::/32;
set_real_ip_from 2405:b500::/32;
set_real_ip_from 2405:8100::/32;
set_real_ip_from 2a06:98c0::/29;
set_real_ip_from 2c0f:f248::/32;
real_ip_header CF-Connecting-IP;
EOF
}

# A domain's static-serving body (root + index + the static catch-all). The API
# locations are emitted alongside by the callers.
static_body() {
  local webroot="$1"
  cat <<EOF
    client_max_body_size 1m;
    root $webroot;
    index index.html;
$(api_locations)
$(legacy_feed_locations)
$(dead_forum_locations)
    location / {
        try_files \$uri \$uri/ =404;
    }
EOF
}

server_block_http_only() {
  local domain="$1" webroot="$2"
  cat <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $domain$( [[ "$WWW_ALIAS" == "true" ]] && echo " www.$domain" );

    location ^~ /.well-known/acme-challenge/ {
        root $ACME_WEBROOT;
        default_type "text/plain";
    }
$(static_body "$webroot")
}
EOF
}

server_block_https() {
  local domain="$1" webroot="$2"
  cat <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $domain$( [[ "$WWW_ALIAS" == "true" ]] && echo " www.$domain" );
    location ^~ /.well-known/acme-challenge/ {
        root $ACME_WEBROOT;
        default_type "text/plain";
    }
    location / { return 301 https://\$host\$request_uri; }
}

server {
    # http2 is omitted to stay version-neutral: the "listen ... http2" form is
    # deprecated on nginx >=1.25 and the "http2 on;" directive doesn't exist
    # before 1.25.1. For backend sites that want it (nginx >=1.25.1), add
    # "http2 on;" on the line below.
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name $domain$( [[ "$WWW_ALIAS" == "true" ]] && echo " www.$domain" );

    ssl_certificate     /etc/letsencrypt/live/$domain/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$domain/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
$(static_body "$webroot")
}
EOF
}

NGINX_CONF="/etc/nginx/sites-available/vm-sites.conf"
write_nginx_conf() {
  {
    echo "# managed by setup.sh — do not edit by hand; re-run setup.sh to change."
    rate_limit_preamble
    echo
    local d webroot
    for d in $DOMAINS; do
      webroot="$WEBROOT_BASE/$d"
      if [[ -d "/etc/letsencrypt/live/$d" ]]; then
        server_block_https "$d" "$webroot"
      else
        server_block_http_only "$d" "$webroot"
      fi
      echo
    done
  } >"$NGINX_CONF"
  # P0: real client IPs behind Cloudflare (http{}-level, so a conf.d file).
  cloudflare_realip >/etc/nginx/conf.d/cloudflare-realip.conf
}

# ── binary install (used by both full and update) ──────────────────────────────
install_binary() {
  install -d -o root -g root -m 0755 "$APP_DIR"
  local tmp="/tmp/site-engine.new"
  if [[ -n "$ENGINE_BINARY_URL" ]]; then
    log "fetching engine binary from URL"
    curl -fsSL "$ENGINE_BINARY_URL" -o "$tmp"
  elif [[ -f "$ENGINE_BINARY_PATH" ]]; then
    log "using local engine binary at $ENGINE_BINARY_PATH"
    cp "$ENGINE_BINARY_PATH" "$tmp"
  elif [[ -x "$APP_DIR/site-engine" ]]; then
    log "no new binary provided — keeping the installed engine"
    return 0
  else
    echo "[setup] ERROR: no binary. Set ENGINE_BINARY_URL or place one at $ENGINE_BINARY_PATH" >&2
    exit 1
  fi
  chmod 0755 "$tmp"
  mv -f "$tmp" "$APP_DIR/site-engine"          # atomic swap
  chown root:root "$APP_DIR/site-engine"
  log "engine binary installed at $APP_DIR/site-engine"
}

if [[ "$MODE" == "update" ]]; then
  log "MODE=update — replacing engine binary and restarting only"
  install_binary
  systemctl restart site-engine
  systemctl --no-pager status site-engine | head -n 5 || true
  log "update complete"
  exit 0
fi

# ── full provision / rebuild ────────────────────────────────────────────────────
export DEBIAN_FRONTEND=noninteractive
log "domains: $DOMAINS"

log "installing packages"
apt-get update -y
apt-get install -y --no-install-recommends \
  nginx certbot fail2ban ufw curl ca-certificates unattended-upgrades

# service user (system account, no login)
if ! id "$SERVICE_USER" >/dev/null 2>&1; then
  log "creating service user $SERVICE_USER"
  useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
fi

# directories
install -d -o "$SERVICE_USER" -g "$SERVICE_USER" -m 0750 "$DATA_DIR"
install -d -o root -g "$SERVICE_USER" -m 0750 "$ENV_DIR"
install -d -o www-data -g www-data -m 0755 "$ACME_WEBROOT"
install -d -o "$WEBROOT_USER" -g "$WEBROOT_GROUP" -m 0755 "$WEBROOT_BASE"

# per-domain web roots + a placeholder so nginx serves something before the first
# content deploy. The deploy Action (P2) rsyncs the chassis-built site into these.
for d in $DOMAINS; do
  install -d -o "$WEBROOT_USER" -g "$WEBROOT_GROUP" -m 0755 "$WEBROOT_BASE/$d"
  if [[ ! -e "$WEBROOT_BASE/$d/index.html" ]]; then
    printf '<!doctype html><meta charset="utf-8"><title>%s</title>\n' "$d" >"$WEBROOT_BASE/$d/index.html"
    chown "$WEBROOT_OWNER" "$WEBROOT_BASE/$d/index.html"
  fi
done

install_binary

# env file: do not overwrite if present (it may hold INTERNAL_API_KEY). Placeholder otherwise.
if [[ ! -f "$ENV_FILE" ]]; then
  log "WARNING: $ENV_FILE not found. Writing a placeholder; review before relying on /stats."
  cat >"$ENV_FILE" <<EOF
# site-engine env. See site-engine.env.example.
PORT=$SERVICE_PORT
ENGINE_DATA_DIR=$DATA_DIR
THANKS_PATH=/
GEO_COUNTRY_HEADER=CF-IPCountry
MAX_VALUE_LEN=500
# INTERNAL_API_KEY gates /stats; set a strong value.
INTERNAL_API_KEY=
# ACCEPT_HOSTS is optional: nginx already only routes configured server_names to
# the engine. Leave empty to accept any host that reaches it.
ACCEPT_HOSTS=
EOF
  chown root:"$SERVICE_USER" "$ENV_FILE"
  chmod 0640 "$ENV_FILE"
fi

# ── systemd unit (one engine for all domains) ────────────────────────────────────
log "writing systemd unit"
cat >/etc/systemd/system/site-engine.service <<EOF
[Unit]
Description=site-engine — backend for VM-hosted sites
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
EnvironmentFile=$ENV_FILE
WorkingDirectory=$DATA_DIR
ExecStart=$APP_DIR/site-engine
Restart=on-failure
RestartSec=5

# hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=$DATA_DIR
ProtectKernelTunables=true
ProtectControlGroups=true
RestrictSUIDSGID=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable site-engine >/dev/null 2>&1 || true

# ── engine-update hook (least-privilege path for the engine-deploy Action) ───────
# When DEPLOY_USER is set, install a narrow root-owned script that atomically
# swaps the engine binary and restarts the service, plus a sudoers rule letting
# DEPLOY_USER run ONLY that script. The CI engine-deploy workflow then does:
#   scp site-engine deploy@box:/tmp/site-engine.new
#   ssh deploy@box sudo /usr/local/sbin/site-engine-deploy /tmp/site-engine.new
# The swapped binary runs as the unprivileged $SERVICE_USER (systemd User=), so a
# compromised deploy key cannot escalate to root.
if [[ -n "$DEPLOY_USER" ]]; then
  log "installing engine-update hook for deploy user '$DEPLOY_USER'"
  cat >/usr/local/sbin/site-engine-deploy <<EOF
#!/usr/bin/env bash
set -euo pipefail
src="\${1:?usage: site-engine-deploy /path/to/new/binary}"
[[ -f "\$src" ]] || { echo "no such binary: \$src" >&2; exit 1; }
install -m 0755 -o root -g root "\$src" "$APP_DIR/site-engine.new"
mv -f "$APP_DIR/site-engine.new" "$APP_DIR/site-engine"   # atomic swap
systemctl restart site-engine
systemctl --no-pager status site-engine | head -n 3 || true
EOF
  chmod 0755 /usr/local/sbin/site-engine-deploy
  echo "$DEPLOY_USER ALL=(root) NOPASSWD: /usr/local/sbin/site-engine-deploy" >/etc/sudoers.d/site-engine-deploy
  chmod 0440 /etc/sudoers.d/site-engine-deploy
  visudo -cf /etc/sudoers.d/site-engine-deploy >/dev/null
fi

# ── retention: daily prune of old event files ───────────────────────────────────
# The engine rotates events by writing daily JSONL files; "rotation" needs no
# logrotate (and logrotate's move/truncate would race the open handle). This
# timer enforces RETENTION_DAYS by deleting old daily files. nginx logs are
# handled separately by the size-based logrotate below.
log "installing retention prune timer (${RETENTION_DAYS}d)"
cat >/etc/systemd/system/site-engine-prune.service <<EOF
[Unit]
Description=site-engine — prune event files older than ${RETENTION_DAYS} days

[Service]
Type=oneshot
ExecStart=/usr/bin/find $DATA_DIR -name 'events-*.jsonl' -mtime +${RETENTION_DAYS} -delete
EOF
cat >/etc/systemd/system/site-engine-prune.timer <<EOF
[Unit]
Description=Daily prune of old site-engine event files

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
EOF
systemctl daemon-reload
systemctl enable --now site-engine-prune.timer >/dev/null 2>&1 || true

# ── nginx, stage 1: http (serves static + ACME + API proxy) ──────────────────────
log "writing nginx conf (stage 1: http for all domains)"
write_nginx_conf
ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/vm-sites.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

# ── TLS: per-domain webroot cert (idempotent; graceful on failure) ───────────────
for d in $DOMAINS; do
  if [[ ! -d "/etc/letsencrypt/live/$d" ]]; then
    log "obtaining TLS certificate for $d"
    # NOT `local` — this loop is at top level, and `local` outside a function is
    # a bash error ("can only be used in a function") which, under `set -e`,
    # ABORTS the run. Dormant on a box whose domains all have certs (the branch
    # is skipped); fatal on exactly the path the header advertises — "adding a
    # domain = add it to DOMAINS and re-run". Found + verified 2026-07-25.
    extra_san=""
    if [[ "$WWW_ALIAS" == "true" ]] && getent hosts "www.$d" >/dev/null 2>&1; then
      extra_san="-d www.$d"
      log "including www.$d SAN (DNS resolves)"
    elif [[ "$WWW_ALIAS" == "true" ]]; then
      log "WWW_ALIAS set but www.$d does not resolve — apex-only cert for now; re-run after DNS"
    fi
    certbot certonly --webroot -w "$ACME_WEBROOT" -d "$d" $extra_san \
      --non-interactive --agree-tos -m "$LETSENCRYPT_EMAIL" \
      || log "WARNING: certbot failed for $d — staying on HTTP for it. Fix (DNS/port 80) and re-run; a successful re-run upgrades it to HTTPS."
  else
    log "TLS certificate already present for $d — skipping"
  fi
done

# ── nginx, stage 2: rewrite with https blocks for domains that now have certs ────
log "writing nginx conf (stage 2: https where certs exist)"
write_nginx_conf
nginx -t
systemctl reload nginx

# ── firewall ─────────────────────────────────────────────────────────────────────
log "configuring ufw"
ufw --force reset >/dev/null
ufw default deny incoming
ufw default allow outgoing
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# ── nginx log rotation (size-based) ──────────────────────────────────────────────
log "configuring nginx logrotate"
cat >/etc/logrotate.d/nginx <<'EOF'
/var/log/nginx/*.log {
    size 100M
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 www-data adm
    sharedscripts
    postrotate
        [ -f /run/nginx.pid ] && kill -USR1 $(cat /run/nginx.pid) || true
    endscript
}
EOF

# ── fail2ban (ssh brute-force protection) ────────────────────────────────────────
log "configuring fail2ban"
cat >/etc/fail2ban/jail.local <<'EOF'
[sshd]
enabled = true
maxretry = 5
bantime = 1h
findtime = 10m
EOF
systemctl enable fail2ban >/dev/null 2>&1 || true
systemctl restart fail2ban

# ── unattended security upgrades ──────────────────────────────────────────────────
log "enabling unattended security upgrades"
echo 'Unattended-Upgrade::Automatic-Reboot "false";' >/etc/apt/apt.conf.d/51-site-engine-noreboot
dpkg-reconfigure -f noninteractive unattended-upgrades >/dev/null 2>&1 || true

# ── optional SSH hardening (guarded against lockout) ────────────────────────────
if [[ "$HARDEN_SSH" == "true" ]]; then
  if grep -rqsE '.' /root/.ssh/authorized_keys /home/*/.ssh/authorized_keys 2>/dev/null; then
    log "hardening sshd (key auth detected — disabling password auth)"
    cat >/etc/ssh/sshd_config.d/99-site-engine-hardening.conf <<'EOF'
PasswordAuthentication no
PermitRootLogin prohibit-password
EOF
    systemctl reload ssh 2>/dev/null || systemctl reload sshd 2>/dev/null || true
  else
    log "SKIPPING sshd hardening — no authorized_keys found, would risk lockout"
  fi
fi

# ── start ──────────────────────────────────────────────────────────────────────
log "starting engine"
systemctl restart site-engine
sleep 2
systemctl --no-pager status site-engine | head -n 6 || true

log "done. Checks:"
for d in $DOMAINS; do log "  curl -sS https://$d/health"; done
log "  systemctl status site-engine ; journalctl -u site-engine -f"
log "Per-domain content is deployed separately (git -> Action rsync into $WEBROOT_BASE/<domain>)."

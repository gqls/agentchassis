#!/usr/bin/env bash
#
# setup.sh — provision (or rebuild, or update) a VM to host the traffic-probe.
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
# engine. The probe runs ONE engine (single systemd unit) behind MANY domains.
# Per domain, nginx SERVES the chassis-built static site from a web root and
# PROXIES ONLY the capture API paths (/intent, /api/hit, /stats, /health) to the
# engine. The static files are deployed separately (git -> Action rsync); this
# script only provisions the box, the web roots, nginx, TLS, and the engine.
#
# Run as root on a fresh Ubuntu 22.04/24.04 VM:
#   DOMAINS="relojistas.com surgerylight.com" LETSENCRYPT_EMAIL=you@real.tld \
#   PROBE_BINARY_PATH=/tmp/probe bash setup.sh
#
# Modes:
#   MODE=full   (default) — full provision/rebuild: packages, user, web roots,
#               nginx, TLS, firewall, fail2ban, systemd, start.
#   MODE=update — only replace the engine binary and restart. (Per-domain CONTENT
#               updates come via the deploy Action's rsync, not this script.)
#
# The env file with any secrets (/etc/probe/probe.env) is NOT overwritten if it
# exists — copy probe.env.example there and fill it in (or have the deployer drop
# it over SSH).

set -euo pipefail

# ── parameters ────────────────────────────────────────────────────────────────
DOMAINS="${DOMAINS:?set DOMAINS, space-separated, e.g. \"relojistas.com surgerylight.com\"}"
LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:?set LETSENCRYPT_EMAIL for cert registration}"
if [[ "$LETSENCRYPT_EMAIL" == *@example.com || "$LETSENCRYPT_EMAIL" == *@example.org ]]; then
  echo "[setup] ERROR: LETSENCRYPT_EMAIL is a placeholder ($LETSENCRYPT_EMAIL). Use a real address." >&2
  exit 1
fi
MODE="${MODE:-full}"

SERVICE_USER="${SERVICE_USER:-probe}"
SERVICE_PORT="${SERVICE_PORT:-8080}"          # must match PORT in the env file
APP_DIR="${APP_DIR:-/opt/probe}"
DATA_DIR="${DATA_DIR:-/var/lib/probe}"
ENV_DIR="${ENV_DIR:-/etc/probe}"
ENV_FILE="${ENV_FILE:-$ENV_DIR/probe.env}"
ACME_WEBROOT="${ACME_WEBROOT:-/var/www/letsencrypt}"
WEBROOT_BASE="${WEBROOT_BASE:-/var/www/probe}" # per-domain static roots: $WEBROOT_BASE/<domain>
# Owner of the per-domain web roots. The deploy Action's SSH user (VM_USER) must
# be able to WRITE here; nginx (www-data) only needs READ. Default keeps the old
# www-data ownership; set e.g. "deploy:www-data" to let a `deploy` user rsync in.
# (NEW param vs the idea.uk original.)
WEBROOT_OWNER="${WEBROOT_OWNER:-www-data:www-data}"
WEBROOT_USER="${WEBROOT_OWNER%%:*}"
WEBROOT_GROUP="${WEBROOT_OWNER##*:}"

# Binary source — exactly one applies. URL takes precedence (the chassis path: a
# presigned B2 URL, as the Thunder adapter already does). Otherwise a local path
# already on the box (the manual path: scp the binary up first).
PROBE_BINARY_URL="${PROBE_BINARY_URL:-}"
PROBE_BINARY_PATH="${PROBE_BINARY_PATH:-/tmp/probe}"

HARDEN_SSH="${HARDEN_SSH:-true}"
WHITELIST_IPS="${WHITELIST_IPS:-}"

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
    echo "limit_req_zone \$limit_key zone=probe_rl:10m rate=10r/s;"
  else
    echo "limit_req_zone \$binary_remote_addr zone=probe_rl:10m rate=10r/s;"
  fi
}

# The capture API locations proxied to the single engine. Reused in http + https
# server blocks. Exact-match locations so they win over the static `location /`.
api_locations() {
  cat <<EOF
    location = /intent {
        limit_req zone=probe_rl burst=20 nodelay;
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
    location = /health {
        proxy_pass http://127.0.0.1:$SERVICE_PORT;
        proxy_set_header Host \$host;
    }
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
    server_name $domain;

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
    server_name $domain;
    location ^~ /.well-known/acme-challenge/ {
        root $ACME_WEBROOT;
        default_type "text/plain";
    }
    location / { return 301 https://\$host\$request_uri; }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $domain;

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

NGINX_CONF="/etc/nginx/sites-available/probe.conf"
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
}

# ── binary install (used by both full and update) ──────────────────────────────
install_binary() {
  install -d -o root -g root -m 0755 "$APP_DIR"
  local tmp="/tmp/probe.new"
  if [[ -n "$PROBE_BINARY_URL" ]]; then
    log "fetching engine binary from URL"
    curl -fsSL "$PROBE_BINARY_URL" -o "$tmp"
  elif [[ -f "$PROBE_BINARY_PATH" ]]; then
    log "using local engine binary at $PROBE_BINARY_PATH"
    cp "$PROBE_BINARY_PATH" "$tmp"
  else
    echo "[setup] ERROR: no binary. Set PROBE_BINARY_URL or place one at $PROBE_BINARY_PATH" >&2
    exit 1
  fi
  chmod 0755 "$tmp"
  mv -f "$tmp" "$APP_DIR/probe"          # atomic swap
  chown root:root "$APP_DIR/probe"
  log "engine binary installed at $APP_DIR/probe"
}

if [[ "$MODE" == "update" ]]; then
  log "MODE=update — replacing engine binary and restarting only"
  install_binary
  systemctl restart probe
  systemctl --no-pager status probe | head -n 5 || true
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
# traffic-probe engine env. See probe.env.example.
PORT=$SERVICE_PORT
PROBE_DB_PATH=$DATA_DIR/probe_events.json
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
cat >/etc/systemd/system/probe.service <<EOF
[Unit]
Description=traffic-probe capture engine
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
EnvironmentFile=$ENV_FILE
WorkingDirectory=$DATA_DIR
ExecStart=$APP_DIR/probe
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
systemctl enable probe >/dev/null 2>&1 || true

# ── nginx, stage 1: http (serves static + ACME + API proxy) ──────────────────────
log "writing nginx conf (stage 1: http for all domains)"
write_nginx_conf
ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/probe.conf
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl reload nginx

# ── TLS: per-domain webroot cert (idempotent; graceful on failure) ───────────────
for d in $DOMAINS; do
  if [[ ! -d "/etc/letsencrypt/live/$d" ]]; then
    log "obtaining TLS certificate for $d"
    certbot certonly --webroot -w "$ACME_WEBROOT" -d "$d" \
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
echo 'Unattended-Upgrade::Automatic-Reboot "false";' >/etc/apt/apt.conf.d/51-probe-noreboot
dpkg-reconfigure -f noninteractive unattended-upgrades >/dev/null 2>&1 || true

# ── optional SSH hardening (guarded against lockout) ────────────────────────────
if [[ "$HARDEN_SSH" == "true" ]]; then
  if grep -rqsE '.' /root/.ssh/authorized_keys /home/*/.ssh/authorized_keys 2>/dev/null; then
    log "hardening sshd (key auth detected — disabling password auth)"
    cat >/etc/ssh/sshd_config.d/99-probe-hardening.conf <<'EOF'
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
systemctl restart probe
sleep 2
systemctl --no-pager status probe | head -n 6 || true

log "done. Checks:"
for d in $DOMAINS; do log "  curl -sS https://$d/health"; done
log "  systemctl status probe ; journalctl -u probe -f"
log "Per-domain content is deployed separately (git -> Action rsync into $WEBROOT_BASE/<domain>)."

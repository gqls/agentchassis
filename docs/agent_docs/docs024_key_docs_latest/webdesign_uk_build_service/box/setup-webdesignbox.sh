#!/usr/bin/env bash
# setup-webdesignbox.sh — one-shot provisioning for webdesign.uk's box
# (webdesign.vs.mythic-beasts.com / vds:webdesign, Cambridge). Run as root FROM
# THIS DIRECTORY (it installs the sitesync + nginx files that sit beside it).
# Idempotent: safe to re-run.
#
# Adapted from idea_uk_vm_site/box/provision-pullsync.sh with three deliberate
# differences:
#   1. TUNNEL-ONLY POSTURE: ufw default-deny inbound (SSH excepted), nginx binds
#      127.0.0.1 only. There is no certbot and no public :80/:443 — cloudflared
#      is the only ingress, and it dials OUT.
#   2. The deploy key is added to GitHub via `gh` from the operator's machine
#      (ADMIN on gqls/vm-sites), so the original's interactive PAUSE is replaced
#      by printing the pubkey and exiting if the clone fails.
#   3. sitesync guards on the domain folder existing (box may predate the
#      site's first deploy).
#
# What it does NOT do: create the cloudflared tunnel (needs the owner's
# Cloudflare login — Phase 2 step 5) or install the chat service (Phase 4).
set -euo pipefail
cd "$(dirname "$0")"

SYNC_HOME=/var/lib/sitesync
WEBROOT=/var/www/webdesign.uk
REPO=git@github.com:gqls/vm-sites.git

SSH_ID="$SYNC_HOME/.ssh/id_ed25519"
SSH_KH="$SYNC_HOME/.ssh/known_hosts"
# ssh expands `~` from the passwd entry, NOT $HOME (bugs_open/016) — explicit
# identity + known_hosts, and the same command handed to git.
SSH_CMD="ssh -i $SSH_ID -o IdentitiesOnly=yes -o UserKnownHostsFile=$SSH_KH -o StrictHostKeyChecking=yes"

echo "== packages =="
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq nginx git rsync ufw curl

echo "== firewall: default-deny inbound, SSH only =="
# Fresh box; this is enable, not reset. Allow SSH BEFORE enabling.
ufw allow OpenSSH >/dev/null
ufw default deny incoming >/dev/null
ufw default allow outgoing >/dev/null
ufw --force enable
ufw status verbose | head -8

echo "== dirs =="
install -d -o www-data -g www-data "$WEBROOT" "$SYNC_HOME" "$SYNC_HOME/.ssh"
chmod 700 "$SYNC_HOME/.ssh"

echo "== deploy key =="
if [ ! -f "$SSH_ID" ]; then
  sudo -u www-data ssh-keygen -t ed25519 -N '' -C 'sitesync@webdesignbox1' -f "$SSH_ID"
fi
# ssh-keyscan exits 0 even when it reaches nothing — scan to a temp file,
# assert non-empty, show fingerprints (verify against GitHub's published list).
if [ ! -s "$SSH_KH" ]; then
  T=$(mktemp); ssh-keyscan github.com > "$T" 2>/dev/null
  [ -s "$T" ] || { echo "FATAL: ssh-keyscan returned nothing"; exit 1; }
  ssh-keygen -lf "$T"
  install -o www-data -g www-data -m 644 "$T" "$SSH_KH"; rm -f "$T"
fi
echo "--- public key (add as READ-ONLY deploy key on gqls/vm-sites if not already) ---"
cat "$SSH_ID.pub"

echo "== sparse clone =="
if [ ! -d "$SYNC_HOME/repo/.git" ]; then
  sudo -u www-data env GIT_SSH_COMMAND="$SSH_CMD" \
    git clone --filter=blob:none --sparse "$REPO" "$SYNC_HOME/repo"
  cd "$SYNC_HOME/repo"
  sudo -u www-data git sparse-checkout set webdesign.uk
  cd - >/dev/null
fi

echo "== sitesync script + timer =="
install -m 755 sitesync /usr/local/bin/sitesync
cat > /etc/systemd/system/sitesync.service <<'UNIT'
[Unit]
Description=Pull webdesign.uk built site from vm-sites
[Service]
Type=oneshot
User=www-data
ExecStart=/usr/local/bin/sitesync
UNIT
cat > /etc/systemd/system/sitesync.timer <<'UNIT'
[Unit]
Description=Run sitesync every 5 minutes
[Timer]
OnBootSec=2min
OnUnitActiveSec=5min
[Install]
WantedBy=timers.target
UNIT
systemctl daemon-reload
systemctl enable --now sitesync.timer

echo "== first sync (folder may not exist yet — that is fine) =="
sudo -u www-data /usr/local/bin/sitesync && echo "sync ok"

echo "== nginx: loopback-only vhost =="
rm -f /etc/nginx/sites-enabled/default
install -m 644 webdesign.uk.nginx /etc/nginx/sites-available/webdesign.uk
ln -sf /etc/nginx/sites-available/webdesign.uk /etc/nginx/sites-enabled/webdesign.uk
nginx -t && systemctl reload nginx

echo "== cloudflared (install only — tunnel creation is a separate, owner-authed step) =="
if ! command -v cloudflared >/dev/null; then
  mkdir -p --mode=0755 /usr/share/keyrings
  curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg > /usr/share/keyrings/cloudflare-main.gpg
  echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared noble main" \
    > /etc/apt/sources.list.d/cloudflared.list
  apt-get update -qq && apt-get install -y -qq cloudflared
fi
cloudflared --version

echo "== verify =="
systemctl is-active nginx sitesync.timer
ss -tlnp | grep -E ':8080|:80\b|:443' || true
echo "DONE. Next: cloudflared tunnel (owner-authed), then Phase 3 pages arrive by pull."

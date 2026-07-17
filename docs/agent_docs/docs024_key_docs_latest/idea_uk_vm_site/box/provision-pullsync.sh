#!/usr/bin/env bash
# provision-pullsync.sh — one-shot §3a provisioning for the idea.uk box
# (116.203.204.115). Run as root FROM THIS DIRECTORY (it installs the sitesync
# script + units that sit beside it). Idempotent: safe to re-run.
#
# What it does:  dirs → deploy key (PAUSES for you to add it on GitHub) →
# sparse clone → install script + timer → first sync → verify.
# What it does NOT do: touch nginx. That is §3b–3e, a separate deliberate step.
set -euo pipefail
cd "$(dirname "$0")"

SYNC_HOME=/var/lib/sitesync
WEBROOT=/var/www/idea.uk

echo "== dirs =="
install -d -o www-data -g www-data "$WEBROOT" "$SYNC_HOME" "$SYNC_HOME/.ssh"
chmod 700 "$SYNC_HOME/.ssh"

echo "== deploy key =="
if [ ! -f "$SYNC_HOME/.ssh/id_ed25519" ]; then
  sudo -u www-data ssh-keygen -t ed25519 -N '' -C 'sitesync@idea.uk-box' \
    -f "$SYNC_HOME/.ssh/id_ed25519"
fi
sudo -u www-data sh -c "ssh-keygen -F github.com -f $SYNC_HOME/.ssh/known_hosts >/dev/null 2>&1 \
  || ssh-keyscan github.com >> $SYNC_HOME/.ssh/known_hosts 2>/dev/null"

echo
echo ">>> Add this as a Deploy Key on gqls/vm-sites (Settings → Deploy keys)."
echo ">>> Title: idea.uk-box sitesync — LEAVE 'Allow write access' UNTICKED (read-only)."
echo
cat "$SYNC_HOME/.ssh/id_ed25519.pub"
echo
read -rp "Press Enter once the deploy key is added... "

echo "== sparse clone (this box fetches ONLY idea.uk/) =="
if [ ! -d "$SYNC_HOME/repo/.git" ]; then
  sudo -u www-data env HOME="$SYNC_HOME" \
    git clone --filter=blob:none --no-checkout git@github.com:gqls/vm-sites.git "$SYNC_HOME/repo"
  cd "$SYNC_HOME/repo"
  sudo -u www-data env HOME="$SYNC_HOME" git sparse-checkout set idea.uk
  sudo -u www-data env HOME="$SYNC_HOME" git checkout main
  cd - >/dev/null
fi

echo "== install script + units =="
install -m 755 sitesync /usr/local/bin/sitesync
install -m 644 sitesync.service sitesync.timer /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now sitesync.timer

echo "== first sync + verify =="
systemctl start sitesync.service
echo "-- timer:"; systemctl list-timers sitesync.timer --no-pager | head -3
echo "-- webroot:"
ls "$WEBROOT"
for p in index.html tools.html report.html about.html contact.html privacy.html \
         guides/index.html news/index.html; do
  [ -f "$WEBROOT/$p" ] && echo "OK  $p" || echo "MISSING  $p"
done
echo
echo "Done. nginx is untouched — proceed to §3b only when the 8 pages show OK."

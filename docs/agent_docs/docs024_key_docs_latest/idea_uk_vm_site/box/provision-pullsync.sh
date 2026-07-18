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

# ⚠️ ssh expands `~` from the PASSWD ENTRY (getpwuid), NOT from $HOME. www-data's
# passwd home is /var/www, so `env HOME=/var/lib/sitesync git clone` configures git
# but leaves ssh looking in /var/www/.ssh — which www-data cannot even create. That
# is why the key was never offered ("Permission denied (publickey)") and the host key
# we had written was ignored ("Host key verification failed"). Hit for real on the
# box 2026-07-18; see /bugs_open/016.
# Fix: never rely on HOME for ssh. Name the identity and known_hosts EXPLICITLY, and
# hand git the same command via GIT_SSH_COMMAND.
SSH_ID="$SYNC_HOME/.ssh/id_ed25519"
SSH_KH="$SYNC_HOME/.ssh/known_hosts"
SSH_CMD="ssh -i $SSH_ID -o IdentitiesOnly=yes -o UserKnownHostsFile=$SSH_KH -o StrictHostKeyChecking=yes"

echo "== dirs =="
install -d -o www-data -g www-data "$WEBROOT" "$SYNC_HOME" "$SYNC_HOME/.ssh"
chmod 700 "$SYNC_HOME/.ssh"

echo "== deploy key =="
if [ ! -f "$SYNC_HOME/.ssh/id_ed25519" ]; then
  sudo -u www-data ssh-keygen -t ed25519 -N '' -C 'sitesync@idea.uk-box' \
    -f "$SYNC_HOME/.ssh/id_ed25519"
fi
# GitHub host key → known_hosts. Do NOT trust ssh-keyscan's exit status: it exits 0
# even when it reaches nothing, which silently leaves an EMPTY known_hosts and the
# clone then dies with "Host key verification failed" (hit for real 2026-07-18).
# So: scan to a temp file, assert it is non-empty, and print the fingerprints to
# check against GitHub's published list before installing.
if ! sudo -u www-data ssh-keygen -F github.com -f "$SSH_KH" >/dev/null 2>&1; then
  echo "-- fetching GitHub host keys"
  KS=$(mktemp)
  ssh-keyscan -t ed25519,rsa,ecdsa github.com > "$KS" 2>/dev/null || true
  if [ ! -s "$KS" ]; then
    rm -f "$KS"
    echo "ERROR: could not fetch GitHub's host keys — outbound SSH (port 22) is probably blocked."
    echo "       Test:  ssh -T -p 22 git@github.com"
    echo "       If 22 is blocked, use GitHub's SSH-over-443 endpoint instead:"
    echo "         ssh-keyscan -t ed25519 -p 443 ssh.github.com"
    echo "       ...and add to $SYNC_HOME/.ssh/config:"
    echo "         Host github.com"
    echo "           HostName ssh.github.com"
    echo "           Port 443"
    exit 1
  fi
  echo "-- host key fingerprints (compare with GitHub's published list:"
  echo "   https://docs.github.com/authentication/keeping-your-account-secure/githubs-ssh-key-fingerprints )"
  ssh-keygen -lf "$KS"
  echo "   expected ED25519: SHA256:+DiY3wvvV6TuJJhbpZisF/zLDA0zPMSvHdkr4UvCOqU"
  cat "$KS" >> "$SSH_KH"
  rm -f "$KS"
  chown www-data:www-data "$SSH_KH"
  chmod 644 "$SSH_KH"
fi

echo "== pre-flight: can this box authenticate to GitHub? =="
# Test FIRST, prompt only if it fails: once the key is registered this stage is a
# silent no-op, so re-runs need no interaction at all. (The old order paused
# unconditionally — and `read` cannot work under `ssh host 'bash script'`, which has
# no TTY, so the pause was skipped and the failure came from the clone instead.)
# `ssh -T git@github.com` exits 1 even on SUCCESS (GitHub allows no shell), so match
# the greeting text, not the exit status. A deploy key greets with the repo name.
gh_probe() { sudo -u www-data $SSH_CMD -o BatchMode=yes -T git@github.com 2>&1 || true; }

GH_GREETING=$(gh_probe)
case "$GH_GREETING" in
  *"successfully authenticated"*) : ;;   # already registered — say nothing, carry on
  *)
    echo "$GH_GREETING"
    echo
    echo ">>> Add this as a Deploy Key on gqls/vm-sites (Settings → Deploy keys)."
    echo ">>> Title: idea.uk-box sitesync — LEAVE 'Allow write access' UNTICKED (read-only)."
    echo
    cat "${SSH_ID}.pub"
    echo
    if [ -t 0 ]; then
      read -rp "Press Enter once the deploy key is added... "
      GH_GREETING=$(gh_probe)
    else
      echo "NOTE: no TTY (you ran this via 'ssh host bash script'), so I cannot pause."
      echo "      Add the key above, then re-run — or use: ssh -t root@<box> '…'"
      exit 1
    fi
    ;;
esac

case "$GH_GREETING" in
  *"successfully authenticated"*) echo "-- OK: deploy key accepted." ;;
  *"Host key verification failed"*)
    echo "$GH_GREETING"
    echo "ERROR: host key not trusted — $SSH_KH is empty or unreadable by www-data."; exit 1 ;;
  *"/var/www/.ssh"*)
    echo "$GH_GREETING"
    echo "ERROR: ssh fell back to www-data's passwd home (/var/www) — the explicit"
    echo "       -i/-o UserKnownHostsFile flags did not reach it. See /bugs_open/016."; exit 1 ;;
  *"Permission denied"*)
    echo "$GH_GREETING"
    echo "ERROR: GitHub still refuses the key. Check it was added to gqls/vm-sites"
    echo "       (not a personal account) and re-run."; exit 1 ;;
  *) echo "$GH_GREETING"; echo "ERROR: unexpected SSH result — resolve before cloning."; exit 1 ;;
esac

echo "== sparse clone (this box fetches ONLY idea.uk/) =="
if [ ! -d "$SYNC_HOME/repo/.git" ]; then
  sudo -u www-data env GIT_SSH_COMMAND="$SSH_CMD" \
    git clone --filter=blob:none --no-checkout git@github.com:gqls/vm-sites.git "$SYNC_HOME/repo"
  cd "$SYNC_HOME/repo"
  sudo -u www-data git sparse-checkout set idea.uk
  sudo -u www-data env GIT_SSH_COMMAND="$SSH_CMD" git checkout main
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

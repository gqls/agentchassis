#!/usr/bin/env bash
# Weekly contact-sheet refresh — the cadence the owner approved 2026-08-04.
#
# Ran by cron (crontab -l | grep contact_sheet). Three steps, each of which can
# fail independently and says so:
#   1. auth pre-check   — the kubeconfig token expires every ~3 days and is
#                         refreshed by the owner; a cron run WILL sometimes land
#                         on a dead token. That is a "tell the owner" outcome,
#                         not a retry loop.
#   2. regenerate       — contact_sheet.py does everything deterministic:
#                         queries the notes, fetches the renders, writes the
#                         theme-aware HTML.
#   3. notify           — a push saying the sheet is fresh and where it is.
#                         MEASURED 2026-08-04: headless `claude -p` has NO
#                         Artifact tool (checked the roster, not assumed), so
#                         the claude.ai gallery page cannot be republished from
#                         cron. It refreshes on request instead — say
#                         "republish the contact sheet artifact" in any
#                         interactive session (RUNBOOK §weekly contact sheet).
#
# Logs append to ~/acceptance_renders/refresh.log — check there first when the
# sheet looks stale.
set -u
# /snap/bin matters: kubectl is a snap here, and a cron PATH without it makes
# the auth pre-check report "token expired" for what is actually command-not-
# found — exactly the false alarm this script exists to avoid (bitten on the
# wrapper's very first run, 2026-08-04).
export PATH="/home/ant/.local/bin:/usr/local/bin:/usr/bin:/bin:/snap/bin"
LOG=~/acceptance_renders/refresh.log
SHEET=~/acceptance_renders/contact_sheet.html
# The 08-03 artifact (95bd1577…) was deleted from the gallery; republished
# fresh 2026-08-04 at this URL. If this one is deleted too, any interactive
# session can mint a new one from $SHEET and should update this line.
ARTIFACT_URL="https://claude.ai/code/artifact/14a45889-e1f0-46e9-969a-08295cc36650"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PY=~/.venvs/vonc_pw/bin/python3; [ -x "$PY" ] || PY=python3

mkdir -p ~/acceptance_renders
{
  echo "== $(date -Is) refresh start"

  if ! kubectl -n ai-persona-system exec -i postgres-clients-0 -- true 2>/dev/null; then
    echo "AUTH FAILED — kubeconfig token likely expired (owner refreshes; ~3-day expiry)."
    claude -p "Send a push notification: 'Weekly contact sheet NOT refreshed: cluster auth failed (kubeconfig token likely expired). Run weekly_contact_sheet_refresh.sh after refreshing.' Use the PushNotification tool, then stop." \
      --allowedTools "PushNotification" 2>&1 | tail -2
    echo "== $(date -Is) refresh aborted (auth)"
    exit 1
  fi

  if ! "$PY" "$SCRIPT_DIR/contact_sheet.py" --limit 8 --out "$SHEET"; then
    echo "GENERATION FAILED — see above."
    echo "== $(date -Is) refresh aborted (generate)"
    exit 1
  fi

  claude -p "Send a push notification: 'Contact sheet refreshed $(date +%F) — open ~/acceptance_renders/contact_sheet.html, or ask any session to republish the contact sheet artifact.' Use the PushNotification tool, then stop." \
    --allowedTools "PushNotification" 2>&1 | tail -2

  echo "== $(date -Is) refresh done"
} >> "$LOG" 2>&1

#!/usr/bin/env bash
# swap-chat-api-key.sh — put the webdesign.uk chat on a different Anthropic
# budget, without the key ever touching a session, a transcript, a shell
# history, a process list or a log line.
#
# WHY THIS EXISTS RATHER THAN "ssh in and edit the file"
# -----------------------------------------------------
# The hand recipe (PLAN_2026-08-31 §2) is three steps and each fails QUIETLY:
#
#  1. A mistyped key does NOT stop the service. main.go checks only that the key
#     is non-EMPTY, so systemd reports `active`, /health returns 200, and every
#     visitor silently gets the fail-closed contact line instead of a
#     conversation — the same symptom as the usage-limit outages, from a
#     different cause. So this script PREFLIGHTS the new key against the real
#     API and refuses to write anything on a non-200.
#  2. Editing the file and forgetting the restart changes NOTHING: systemd reads
#     EnvironmentFile once, at start. Worse, every check that reads the FILE
#     then agrees with the file. So this asks the RUNNING PROCESS what it holds.
#  3. A restart that fails leaves the shopfront with no chat at all. So the old
#     file is backed up first and RESTORED automatically if the unit does not
#     come back.
#
# WHAT A FINGERPRINT IS
# ---------------------
# Keys must never be printed, so the currency here is twelve hex characters of
# sha256 over the key. Compute the same number for any key you hold with:
#
#     printf %s "<the key>" | sha256sum | cut -c1-12
#
# Equal digests = the same key. Nothing secret is revealed, so fingerprints are
# safe in chat, docs and commit messages — and the journal, this script and the
# runbook all speak that one currency.
#
# USAGE
#   ./swap-chat-api-key.sh --status   # read-only: which key/budget is live now
#   ./swap-chat-api-key.sh --check    # prompt for a key, TEST it, write nothing
#   ./swap-chat-api-key.sh            # prompt, test, back up, write, restart, verify
#
# The key is typed at a hidden prompt (so it never enters shell history),
# travels to the box on ssh STDIN — never in argv, where /proc/<pid>/cmdline
# would expose it — and is written by awk reading a 0600 temp file deleted on
# exit. This session's standing rule is that a key value never passes through a
# Claude session; that is why the owner runs this, and why it prompts rather
# than taking an argument.

set -euo pipefail

BOX_HOST="${BOX_HOST:-webdesign.vs.mythic-beasts.com}"
BOX_USER="${BOX_USER:-root}"
BOX_KEY="${BOX_KEY:-$HOME/.ssh/webdesign_box_ed25519}"
ENV_FILE="${ENV_FILE:-/etc/webdesign-chat.env}"
UNIT="${UNIT:-webdesign-chat}"
PORT="${PORT:-8081}"

MODE=apply
case "${1:-}" in
  --status)          MODE=status ;;
  --check|--dry-run) MODE=check ;;
  "")                ;;
  *) echo "usage: $0 [--status|--check]" >&2; exit 2 ;;
esac

# The program that runs on the box. Held in a quoted heredoc (nothing expands
# here; every variable in it is the BOX's) and shipped base64-encoded, which
# both survives any quoting and guarantees it runs under bash — the remote
# login shell is not ours to assume, and `set -o pipefail` is not POSIX.
REMOTE_PROG=$(cat <<'REMOTE_BODY_EOF'
#!/usr/bin/env bash
# Runs ON THE BOX. Reads the new key from stdin (apply/check modes only).
# Never prints the key; never puts it in argv.
set -euo pipefail
umask 077

: "${ENV_FILE:=/etc/webdesign-chat.env}"
: "${UNIT:=webdesign-chat}"
: "${PORT:=8081}"
: "${MODE:=status}"

fp() { sha256sum | cut -c1-12; }

env_fp() {
  local v
  v=$(grep -m1 '^ANTHROPIC_API_KEY=' "$ENV_FILE" 2>/dev/null | cut -d= -f2- | tr -d '\r\n' || true)
  if [ -z "$v" ]; then echo "none"; else printf '%s' "$v" | fp; fi
}

# What the RUNNING process holds. The file cannot answer this: systemd reads
# EnvironmentFile once, at start, so an edited-but-not-restarted service
# disagrees with its own config and every file-based check sides with the file.
running_fp() {
  journalctl -u "$UNIT" --no-pager -o cat 2>/dev/null |
    grep 'api key fingerprint' | tail -1 |
    sed -n 's/.*sha256=\([0-9a-f]*\).*/\1/p' || true
}

report() {
  local r e
  e=$(env_fp); r=$(running_fp)
  echo "  unit           : $(systemctl is-active "$UNIT" 2>/dev/null || true)"
  echo "  env file key   : $e   ($ENV_FILE)"
  if [ -n "$r" ]; then
    echo "  RUNNING process: $r   <- the authority; the file is only an intention"
    [ "$r" != "$e" ] && echo "  ** MISMATCH — the file was edited and the service NOT restarted **"
  else
    echo "  RUNNING process: unknown — this binary predates the fingerprint line."
    echo "                   'make box-release' makes the running key checkable."
  fi
  echo "  build          : $(journalctl -u "$UNIT" --no-pager -o cat 2>/dev/null | grep 'build provenance' | tail -1 | sed -n 's/.*git_commit=//p' || true)"
  echo "  /health        : $(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:$PORT/health" || echo unreachable)"
}

if [ "$MODE" = status ]; then
  echo "webdesign.uk chat — live key/budget state:"
  report
  exit 0
fi

D=$(mktemp -d)
trap 'rm -rf "$D"' EXIT
cat > "$D/raw"
tr -d '\r\n' < "$D/raw" > "$D/key"; rm -f "$D/raw"

[ -s "$D/key" ] || { echo "ERROR: empty key received — nothing done." >&2; exit 1; }
if grep -q '[[:space:]]' "$D/key"; then
  echo "ERROR: the key contains whitespace — a paste accident. Nothing done." >&2; exit 1
fi

NEW_FP=$(fp < "$D/key")
OLD_FP=$(env_fp)
echo "  current key    : $OLD_FP"
echo "  new key        : $NEW_FP"

# One token: the cheapest question separating "works" from "rejected" from
# "valid but the account is already capped". The key goes in a curl config
# file, never in argv (/proc/<pid>/cmdline is world-readable).
echo "  preflight      : calling the API with the NEW key (1 token)..."
printf '%s' '{"model":"claude-haiku-4-5","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}' > "$D/body.json"
{
  echo 'url = "https://api.anthropic.com/v1/messages"'
  printf 'header = "x-api-key: %s"\n' "$(cat "$D/key")"
  echo 'header = "anthropic-version: 2023-06-01"'
  echo 'header = "content-type: application/json"'
  printf 'data = "@%s"\n' "$D/body.json"
} > "$D/curl.cfg"
CODE=$(curl -sS --config "$D/curl.cfg" -o "$D/out.json" -w '%{http_code}' --max-time 30 || echo 000)
rm -f "$D/curl.cfg"

if [ "$CODE" != "200" ]; then
  echo
  echo "REFUSED — the new key did not work, so NOTHING was changed."
  echo "  HTTP $CODE"
  echo "  $(head -c 400 "$D/out.json" 2>/dev/null || true)"
  case "$CODE" in
    401) echo "  401 = wrong, revoked, or incompletely pasted key." ;;
    400) echo "  400 naming 'usage limits' = the key is VALID but its account is ALREADY"
         echo "        capped. Swapping to it would buy nothing — raise that account's"
         echo "        limit first, then re-run this." ;;
    000) echo "  000 = the box could not reach api.anthropic.com at all." ;;
  esac
  exit 1
fi
echo "  preflight      : HTTP 200 — the key works and its account is not capped."

if [ "$OLD_FP" = "$NEW_FP" ]; then
  echo; echo "The chat is ALREADY on this key ($NEW_FP). Nothing to do."; report; exit 0
fi
if [ "$MODE" = check ]; then
  echo; echo "--check: the key is good. Nothing written. Re-run without --check to apply."; exit 0
fi

[ "$(grep -c '^ANTHROPIC_API_KEY=' "$ENV_FILE")" = "1" ] || {
  echo "ERROR: $ENV_FILE has no single ANTHROPIC_API_KEY line — refusing to guess." >&2; exit 1; }

BAK="$ENV_FILE.bak-$(date -u +%Y%m%dT%H%M%SZ)"
cp -a "$ENV_FILE" "$BAK"
awk -v kf="$D/key" 'BEGIN{getline k < kf}
  /^ANTHROPIC_API_KEY=/{print "ANTHROPIC_API_KEY=" k; next} {print}' "$ENV_FILE" > "$D/new"

# Exactly one line may differ (one '<' + one '>'). Anything else means awk did
# something unintended to a file that also holds the facts token and the
# contact details the fail-closed path needs.
CHANGED=$(diff "$ENV_FILE" "$D/new" | grep -c '^[<>]' || true)
[ "$CHANGED" = "2" ] || { echo "ERROR: the rewrite would change $((CHANGED/2)) lines, not 1 — refusing." >&2; exit 1; }
grep -q '^ANTHROPIC_API_KEY=' "$D/new" || { echo "ERROR: rewrite lost the key line — refusing." >&2; exit 1; }

chown root:root "$D/new"; chmod 600 "$D/new"; mv "$D/new" "$ENV_FILE"
echo "  wrote          : $ENV_FILE   (backup: $BAK)"

systemctl restart "$UNIT"
sleep 3
if ! systemctl is-active --quiet "$UNIT"; then
  echo "RESTART FAILED — restoring the previous key and restarting." >&2
  cp -a "$BAK" "$ENV_FILE"; systemctl restart "$UNIT" || true; sleep 2
  journalctl -u "$UNIT" -n 15 --no-pager -o cat >&2
  exit 1
fi

echo; echo "DONE. Live state:"; report
R=$(running_fp)
if [ -n "$R" ] && [ "$R" != "$NEW_FP" ]; then
  echo
  echo "** WARNING: the running process reports $R, not the $NEW_FP just written."
  echo "   Do NOT treat this swap as done. **"
  exit 1
fi
REMOTE_BODY_EOF
)

B64=$(printf '%s' "$REMOTE_PROG" | base64 | tr -d '\n')
# `echo | base64 -d` has its own stdin, so the ssh stdin (the key) stays intact
# for `bash "$P"`. Env assignments carry the config; none of them is secret.
RUNNER="P=\$(mktemp) && echo $B64 | base64 -d > \$P && ENV_FILE='$ENV_FILE' UNIT='$UNIT' PORT='$PORT' MODE='$MODE' bash \$P; RC=\$?; rm -f \$P; exit \$RC"

echo "webdesign.uk chat — API key / budget"
echo "  box  : $BOX_USER@$BOX_HOST"
echo "  file : $ENV_FILE    unit: $UNIT"
echo

if [ "$MODE" = status ]; then
  exec ssh -i "$BOX_KEY" "$BOX_USER@$BOX_HOST" "$RUNNER"
fi

echo "Paste the NEW Anthropic API key. It is not echoed, not written to your"
echo "shell history, and not stored on this machine — it goes straight to the box."
read -rsp "  new key: " NEW_KEY; echo
[ -n "$NEW_KEY" ] || { echo "Nothing entered — aborted."; exit 1; }
echo "  fingerprint    : $(printf %s "$NEW_KEY" | sha256sum | cut -c1-12)"
echo

set +e
printf '%s' "$NEW_KEY" | ssh -i "$BOX_KEY" "$BOX_USER@$BOX_HOST" "$RUNNER"
RC=$?
set -e
unset NEW_KEY

if [ $RC -eq 0 ] && [ "$MODE" = apply ]; then
  cat <<'NEXT'

The last step is the only one that proves it end to end, and it is yours:
open https://webdesign.uk, ask the chat one question, and check you get a real
answer rather than the "please reach us directly" contact line. That call will
then appear against the NEW account in the Anthropic Console — which is the
artefact proving the BUDGET moved, not merely the key.
NEXT
fi
exit $RC

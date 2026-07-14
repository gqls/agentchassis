#!/usr/bin/env bash
# 094_pull_digest_to_file.sh — deliver the fix-loop digest as a committed file.
#
# WHY a puller. The digest is COMPOSED inside the chassis (the fixloop-digest
# agent, deterministic, no LLM) and persisted to doc_notes. But a pod cannot
# write to your working tree. This script bridges that gap: it reads the latest
# digest out of doc_notes and writes it here, so it lands in your normal git
# activity and its history becomes a day-by-day diary.
#
# TWO WAYS TO USE IT:
#   1. On demand — run it, review, commit.
#   2. Daily — pair the scheduled fixloop-digest agent (composes to doc_notes)
#      with a local cron running this script (writes + commits the file). That
#      is the "it just arrives" delivery, no image change needed.
#
# Usage:  ./094_pull_digest_to_file.sh [--commit]
#   --commit : also git add + commit the refreshed files (for cron use).
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
PSQL="kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -t -A"

# Latest digest body + the date it was generated (for the archive filename).
BODY="$($PSQL -c "SELECT body FROM doc_notes WHERE categories ? 'digest' ORDER BY created_at DESC LIMIT 1;")"
if [ -z "$BODY" ]; then
  echo "No digest found in doc_notes (has the fixloop-digest agent run?)." >&2
  exit 1
fi
GEN_DATE="$($PSQL -c "SELECT to_char(created_at,'YYYY-MM-DD') FROM doc_notes WHERE categories ? 'digest' ORDER BY created_at DESC LIMIT 1;")"

printf '%s\n' "$BODY" > "$DIR/DIGEST_latest.md"
mkdir -p "$DIR/archive"
printf '%s\n' "$BODY" > "$DIR/archive/DIGEST_${GEN_DATE}.md"
echo "Wrote DIGEST_latest.md and archive/DIGEST_${GEN_DATE}.md"

if [ "${1:-}" = "--commit" ]; then
  cd "$DIR/../.."
  git add docs/fixloop_digests/DIGEST_latest.md "docs/fixloop_digests/archive/DIGEST_${GEN_DATE}.md"
  if git diff --cached --quiet; then
    echo "No digest change to commit."
  else
    git commit -m "chore(digest): fix-loop digest ${GEN_DATE}" >/dev/null
    echo "Committed digest ${GEN_DATE} (not pushed — push per your workflow)."
  fi
fi

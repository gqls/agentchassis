#!/usr/bin/env bash
# rollback_provocation.sh — take a bad provocation off the site.
#
# PLAN §10.4 requires this to EXIST AND BE TESTED before the first automated
# publish: "One command that restores the previous provocation, written and
# tested while nothing is going wrong. Not 'we will write one if we need it' —
# the moment it is needed is the worst moment to write it."
#
# WHY IT WORKS ON THE POOL AND NOT ON THE ARTEFACT — this is the whole point, and
# `publish_feed.sh --rollback` alone is NOT a rollback for an automated pipeline.
# Restoring the previous provocations.json puts the right bytes on the site for
# up to six hours. Then `provocation-feed-refresh` fires, re-derives the feed
# from the pool, finds the bad provocation still `approved` with its date still
# arrived, and republishes it. An artefact rollback against a scheduler that
# re-derives from source is a delay, not a reversal.
#
# So this retires the row. The feed action selects
#   WHERE status='approved' AND publish_on IS NOT NULL
# and takes the LATEST whose date has arrived, so retiring today's entry makes
# the previous one today's automatically — through the same derivation path the
# publisher uses, which means the pool and the feed cannot disagree afterwards.
# `retired` is one of the four statuses the table's CHECK constraint already
# allows; nothing in the estate wrote it before this script.
#
# USAGE
#   ./rollback_provocation.sh <slug>              # DRY RUN: shows the effect, writes nothing
#   ./rollback_provocation.sh <slug> --apply      # retire it
#   ./rollback_provocation.sh <slug> --apply --now # ...and force an immediate republish
#
# The dry run is the default deliberately: the moment you reach for this, you are
# already having a bad day, and a command that acts before it explains itself is
# the wrong shape for that moment.

set -euo pipefail

SLUG="${1:-}"
DOMAIN="${DOMAIN:-vonc.com}"
APPLY=0
FORCE_NOW=0
for arg in "${@:2}"; do
  case "$arg" in
    --apply) APPLY=1 ;;
    --now)   FORCE_NOW=1 ;;
    *) echo "unknown flag: $arg" >&2; exit 2 ;;
  esac
done

if [ -z "$SLUG" ]; then
  echo "usage: $0 <slug> [--apply] [--now]" >&2
  echo "  DOMAIN env var selects the site (default vonc.com)" >&2
  exit 2
fi

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 --
      psql -U clients_user -d clients_db -v ON_ERROR_STOP=1)

echo "=== rollback_provocation: $SLUG on $DOMAIN ==="
echo

# ---------------------------------------------------------------------------
# The whole effect is computed inside ONE transaction that is ROLLED BACK on a
# dry run. That is what makes the preview trustworthy: it is not a description of
# what would happen, it is the actual UPDATE, with the actual selection re-run
# against it, then undone. A preview computed by a different query than the one
# that acts is the drift class this estate keeps rediscovering.
# ---------------------------------------------------------------------------
FINAL="ROLLBACK"
[ "$APPLY" = "1" ] && FINAL="COMMIT"

"${PSQL[@]}" <<SQL
BEGIN;

\echo '-- the row being retired:'
SELECT slug, publish_on, status, source
  FROM provocations WHERE domain = '${DOMAIN}' AND slug = '${SLUG}';

\echo ''
\echo '-- today BEFORE (latest approved whose date has arrived):'
SELECT slug, publish_on FROM provocations
 WHERE domain = '${DOMAIN}' AND status = 'approved'
   AND publish_on IS NOT NULL AND publish_on <= CURRENT_DATE
 ORDER BY publish_on DESC LIMIT 1;

UPDATE provocations
   SET status = 'retired'
 WHERE domain = '${DOMAIN}' AND slug = '${SLUG}' AND status = 'approved';

\echo ''
\echo '-- today AFTER retirement (this is what the next publish will serve):'
SELECT slug, publish_on FROM provocations
 WHERE domain = '${DOMAIN}' AND status = 'approved'
   AND publish_on IS NOT NULL AND publish_on <= CURRENT_DATE
 ORDER BY publish_on DESC LIMIT 1;

${FINAL};
SQL

echo
if [ "$APPLY" != "1" ]; then
  echo "DRY RUN — nothing was written (the transaction was rolled back)."
  echo "Re-run with --apply to retire it."
  exit 0
fi

echo "RETIRED. The next scheduled run will serve the provocation shown as AFTER."
if [ "$FORCE_NOW" = "1" ]; then
  echo
  echo "Forcing an immediate republish..."
  "${PSQL[@]}" -c \
    "UPDATE scheduled_tasks SET last_triggered_at=NULL, last_completed_at=NULL
      WHERE name='provocation-feed-refresh';"
  echo "Done. The publisher picks it up on its next tick."
  echo "VERIFY AT THE ARTEFACT, never at the pool:"
  echo "  curl -s https://${DOMAIN}/data/provocations.json | python3 -c \\"
  echo "    \"import json,sys; d=json.load(sys.stdin); print(d['today']['slug'], d['today']['date'])\""
else
  echo "Not forcing a republish. The site keeps serving the bad provocation until the"
  echo "next 6-hourly tick — pass --now if that is not acceptable."
fi

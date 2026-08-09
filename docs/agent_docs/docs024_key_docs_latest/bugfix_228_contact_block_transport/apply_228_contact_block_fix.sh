#!/usr/bin/env bash
# apply_228_contact_block_fix.sh — the ONLY sanctioned way to run the
# contact-block content_components UPDATE for bugs_open/228.
#
# Do NOT hand-type this SQL. debug_historian's round-2 council objection
# (correlation 46f87e4c-05fc-4a5c-bd6a-93a073b63253) named the exact shape
# this avoids: a "surgical replace()" described only in prose, with no
# needle-occurrence guard, no backup, no RETURNING postcondition, and no
# separate rollback file — the NEEDLE-GATE SQL SURGERY trap, where
# replace() silently no-ops on a missed anchor while still reporting
# UPDATE 1.
#
# PRECONDITION (hard, do not skip): edit 1 (the RenderTemplateReportingMissing
# form_action seeding change, commit 85390ee33) must be pod-verified live on
# EVERY chassis-binary pod fleet-wide first — not just the 2 pods matched by
# `-l app=agent-chassis` (see LANDMINES.md "`-l app=agent-chassis` returns 2
# pods; 41 run that binary"). Enumerate by IMAGE:
#   kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.ownerReferences[0].kind}{"\t"}{.spec.containers[0].image}{"\n"}{end}' | grep agent-chassis
# Every row must be on the tag that was pod-grepped positive for
# "seeded empty form_action for sanitiser". If ANY row (Job or Deployment) is
# on an older tag, that pod CAN still reach RenderTemplateReportingMissing
# for a page-render dispatch — do not proceed until it has cycled or you have
# confirmed it cannot be the one that serves this dispatch.
#
# Usage: ./apply_228_contact_block_fix.sh
# Requires: kubectl context pointed at the cluster, JS_FILE below present.

set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
JS_FILE="$HERE/js_content_after_228_fix.js"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
BACKUP_FILE="$HERE/BACKUP_${STAMP}_contact_block_before_fix.sql"
ROLLBACK_FILE="$HERE/ROLLBACK_${STAMP}_contact_block.sql"

PSQL() {
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db "$@"
}

[ -f "$JS_FILE" ] || { echo "ERROR: $JS_FILE not found" >&2; exit 1; }

OLD_FORM_TAG='<form class="cb-form" id="cb-contact-form" novalidate aria-label="{{.form_heading}}">'
NEW_FORM_TAG='<form class="cb-form" id="cb-contact-form" action="{{.form_action}}" method="POST" novalidate aria-label="{{.form_heading}}">'

echo "== 1/5: needle-count guard on html_template =="
NEEDLE_COUNT=$(PSQL -tA -v ON_ERROR_STOP=1 -c "
  SELECT count(*) FROM content_components
  WHERE function = 'contact-block'
    AND html_template LIKE '%${OLD_FORM_TAG//\%/\\%}%';
")
echo "  occurrences of the exact old form tag: $NEEDLE_COUNT"
if [ "$NEEDLE_COUNT" != "1" ]; then
  echo "ABORT: expected exactly 1 occurrence, got $NEEDLE_COUNT. The stored" >&2
  echo "template no longer matches the sketch verbatim (or matches >1 row) --" >&2
  echo "re-read the live row and re-derive the needle before proceeding." >&2
  exit 1
fi

echo "== 2/5: row identity + setTimeout-still-present check =="
PSQL -v ON_ERROR_STOP=1 -c "
  SELECT id, is_active, forked_from,
         (js_content LIKE '%setTimeout%') AS timer_still_present
  FROM content_components WHERE function = 'contact-block';
"
echo "  Confirm above: exactly one row, is_active=t, forked_from=NULL," \
     "timer_still_present=t. If not, STOP and re-diagnose."
read -r -p "Type EXACT-1-ROW-CONFIRMED to continue, anything else aborts: " CONFIRM1
[ "$CONFIRM1" = "EXACT-1-ROW-CONFIRMED" ] || { echo "aborted"; exit 1; }

echo "== 3/5: backup — capture current row verbatim BEFORE any write =="
PSQL -tA -v ON_ERROR_STOP=1 -c "
  SELECT html_template FROM content_components WHERE function='contact-block';
" > "${BACKUP_FILE}.html_template.txt"
PSQL -tA -v ON_ERROR_STOP=1 -c "
  SELECT js_content FROM content_components WHERE function='contact-block';
" > "${BACKUP_FILE}.js_content.txt"
echo "  backed up to ${BACKUP_FILE}.{html_template,js_content}.txt"

# Auto-generate the rollback file FROM the just-captured backup, not
# hand-typed — this is what makes it trustworthy as a rollback.
{
  echo "-- ROLLBACK for bugs_open/228 contact-block fix, generated $STAMP"
  echo "-- Restores the EXACT pre-fix row captured in ${BACKUP_FILE}.*"
  echo "BEGIN;"
  echo "UPDATE content_components SET html_template = \$RBHTML\$"
  cat "${BACKUP_FILE}.html_template.txt"
  echo "\$RBHTML\$, js_content = \$RBJS\$"
  cat "${BACKUP_FILE}.js_content.txt"
  echo "\$RBJS\$"
  echo "WHERE function = 'contact-block';"
  echo "COMMIT;"
} > "$ROLLBACK_FILE"
echo "  rollback file written: $ROLLBACK_FILE"

echo "== 4/5: apply, with RETURNING as the postcondition =="
{
  echo "BEGIN;"
  echo "UPDATE content_components"
  echo "SET html_template = replace(html_template,"
  echo "      \$OLD\$${OLD_FORM_TAG}\$OLD\$,"
  echo "      \$NEW\$${NEW_FORM_TAG}\$NEW\$"
  echo "    ),"
  echo "    js_content = \$CB228\$"
  cat "$JS_FILE"
  echo "\$CB228\$"
  echo "WHERE function = 'contact-block'"
  echo "RETURNING id,"
  echo "  (html_template LIKE '%action=\"{{.form_action}}\"%') AS has_new_action,"
  echo "  (html_template NOT LIKE '%novalidate aria-label=\"{{.form_heading}}\">%' OR html_template LIKE '%action=%novalidate%') AS form_tag_shape_ok,"
  echo "  length(js_content) AS new_js_len,"
  echo "  (js_content NOT LIKE '%setTimeout%') AS timer_removed,"
  echo "  (js_content LIKE '%Opening your email client%') AS honest_status_present;"
  echo "-- Read the RETURNING row above. All four booleans must be true and"
  echo "-- new_js_len must roughly match $(wc -c < "$JS_FILE") bytes before COMMIT."
  echo "-- This script does NOT auto-commit -- run COMMIT; or ROLLBACK; by hand."
} > "$HERE/APPLIED_${STAMP}_contact_block.sql"
echo "  apply SQL written to $HERE/APPLIED_${STAMP}_contact_block.sql"
echo "  Review it, then run:"
echo "    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db < $HERE/APPLIED_${STAMP}_contact_block.sql"
echo "  Read the RETURNING row, then manually COMMIT or ROLLBACK in a follow-up psql -c."
echo ""
echo "== 5/5: after COMMIT, dispatch the rerenders (separate script) =="
echo "  See dispatch_228_rerenders.sh — do not skip; the DB row changing does"
echo "  NOT propagate to already-rendered pages on its own (render_guardian's"
echo "  round-2 objection)."

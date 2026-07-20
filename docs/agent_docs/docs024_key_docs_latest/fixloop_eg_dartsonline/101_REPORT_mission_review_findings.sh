#!/usr/bin/env bash
# 101_REPORT_mission_review_findings.sh — THE NAMED CONSUMER of the R1
# observe-only mission-review lane (SEED_classifier_mission_R0_R1_2026-07-20.sql).
#
# R1's findings are doc_notes (categories ? 'mission-review'), NOT work items:
# triage_detect_items_action.go promotes ALL detected items site-wide into the
# dispatch pipeline, so a detected work item would be swept toward a nonexistent
# handler — the opposite of observe-only. doc_notes cannot be dispatched.
# bugs_open/023's lesson: a detection lane without a named consumer is a lane
# that ships nothing. THIS script is that consumer — run it before promoting R2
# (enforcement), and weekly alongside 098.
#
# Usage: ./101_REPORT_mission_review_findings.sh [days]   (default 7)
set -euo pipefail
DAYS="${1:-7}"

kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db <<SQL
\\echo '── mission-review findings, last ${DAYS} day(s) ──'
SELECT count(*) AS findings,
       count(DISTINCT site_id) AS sites
FROM doc_notes
WHERE categories ? 'mission-review'
  AND created_at > now() - interval '${DAYS} days';

\\echo ''
\\echo '── the findings (newest first) ──'
SELECT created_at, site_id, left(body, 500) AS finding
FROM doc_notes
WHERE categories ? 'mission-review'
  AND created_at > now() - interval '${DAYS} days'
ORDER BY created_at DESC
LIMIT 50;

\\echo ''
\\echo '── classifier runs in the window (denominator for the objection rate) ──'
SELECT count(*) AS classifier_runs
FROM orchestration_states
WHERE current_step IN ('create_next_item','complete')
  AND collected_data ? 'mission_review'
  AND created_at > now() - interval '${DAYS} days';
SQL

echo ""
echo "R2 promotion gate (owner's call): hand-grade a sample of the findings above"
echo "for false positives first — a consultancy shape is LEGITIMATE when the"
echo "domain's evidence supports it; the mission bans the unargued DEFAULT."

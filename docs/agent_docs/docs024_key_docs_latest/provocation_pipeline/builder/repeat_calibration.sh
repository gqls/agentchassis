#!/usr/bin/env bash
# Run N calibration rounds back to back and print one line per round.
#
# Why: §4a of HANDOFF_2026-08-08 requires more than one round because the judge
# is stochastic, and round 3 on v1.0.1267 proved the point by APPROVING
# cal-bad-insult (a must-reject row) that rounds 1 and 2 both killed as `unsafe`.
# A 1-in-3 observation is not a rate. This exists to turn it into one.
#
# Each round: reset (the gate never re-judges a judged row) -> dispatch -> poll
# -> record the scorecard and every rejection to $OUTDIR/roundN.json.
#
# Usage: ./repeat_calibration.sh 5 [outdir]
set -euo pipefail
N=${1:-3}
OUTDIR=${2:-/home/ant/.claude-scratch/claude-1000/-home-ant-projects-agentchassis/a74963e3-927b-44ef-beb8-5b928730dcd0/scratchpad/calibration}
HERE=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$OUTDIR"
NS=ai-persona-system; PG=postgres-clients-0; DOMAIN=calibration.vonc.com
psql() { kubectl -n "$NS" exec -i "$PG" -- psql -U clients_user -d clients_db "$@"; }

for r in $(seq 1 "$N"); do
  echo "################ round $r of $N ################"
  orch=$("$HERE/run_calibration_round.sh" | sed -n 's/^SAVE: RUN_ORCH_ID=//p')
  [ -n "$orch" ] || { echo "FAILED to dispatch round $r" >&2; exit 1; }
  echo "orchestration: $orch"

  # poll — cover the terminal failure states too, or a crashed round reads as
  # 'still running' and the loop hangs until its iteration cap.
  for i in $(seq 1 30); do
    out=$(psql -tAc "SELECT (SELECT count(*) FROM provocations WHERE domain='$DOMAIN' AND gated_at IS NOT NULL)
        || ' ' || COALESCE((SELECT status FROM orchestration_states WHERE orchestration_id='$orch'),'no-row');" 2>&1 | tr -d '\r')
    case "$out" in
      13\ *)                                  echo "  judged 13"; break;;
      *\ FAILED|*\ ERROR|*\ CANCELLED)        echo "  TERMINAL FAILURE: $out" >&2; break;;
    esac
    sleep 10
  done

  # record the whole round, not just the counts — the reasons are how you tell
  # a stochastic flip from a different defect wearing the same score.
  psql -tAc "SELECT jsonb_pretty(jsonb_build_object(
      'orchestration','$orch',
      'must_approve_approved',(SELECT count(*) FROM provocations WHERE domain='$DOMAIN' AND source_ref LIKE '%must-approve half%' AND status='approved'),
      'must_approve_total',   (SELECT count(*) FROM provocations WHERE domain='$DOMAIN' AND source_ref LIKE '%must-approve half%'),
      'must_reject_rejected', (SELECT count(*) FROM provocations WHERE domain='$DOMAIN' AND source_ref NOT LIKE '%must-approve half%' AND status='rejected'),
      'must_reject_total',    (SELECT count(*) FROM provocations WHERE domain='$DOMAIN' AND source_ref NOT LIKE '%must-approve half%'),
      'unjudged',             (SELECT count(*) FROM provocations WHERE domain='$DOMAIN' AND gated_at IS NULL),
      'rows',(SELECT jsonb_agg(jsonb_build_object('slug',slug,'status',status,
                'judge_ran',gate_verdict->'judge_ran',
                'fatal',(SELECT jsonb_agg(r->>'rule') FROM jsonb_array_elements(gate_verdict->'reasons') r WHERE (r->>'fatal')::bool),
                'note',gate_verdict->'advisory'->>'note') ORDER BY slug)
              FROM provocations WHERE domain='$DOMAIN')));" > "$OUTDIR/round_$r.json"

  python3 - "$OUTDIR/round_$r.json" <<'PY'
import json,sys
d=json.load(open(sys.argv[1]))
print(f"  must-approve {d['must_approve_approved']}/{d['must_approve_total']}   "
      f"must-reject {d['must_reject_rejected']}/{d['must_reject_total']}   "
      f"unjudged {d['unjudged']}")
for r in d['rows']:
    if r['status']!='approved' or not str(r['slug']).startswith('cal-bad'):
        continue
    print(f"    !! LEAK: {r['slug']} approved (judge_ran={r['judge_ran']}) note={r['note']!r}")
PY
done
echo "rounds written to $OUTDIR"

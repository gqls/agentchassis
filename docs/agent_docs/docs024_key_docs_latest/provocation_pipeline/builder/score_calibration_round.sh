#!/usr/bin/env bash
# Score ONE provocation-gate calibration round.
#
# Scores the same way every time, on purpose: a round scored by whichever query
# the session happened to type is how two rounds get compared on different
# bases. Prints, in order:
#   1. completeness  — how many of the 13 rows were actually judged. A round
#      that judged 11 is not a 6/9 result, it is an INCOMPLETE round, and the
#      scorecard alone cannot tell you which (this is the failure §4a warns of).
#   2. the scorecard  — must-approve vs must-reject.
#   3. every rejection with its FATAL reasons — the interesting half.
set -euo pipefail
NS=ai-persona-system; PG=postgres-clients-0; DOMAIN=calibration.vonc.com
psql() { kubectl -n "$NS" exec -i "$PG" -- psql -U clients_user -d clients_db "$@"; }

echo "=== 1. completeness (all 13 judged? unjudged rows make the scorecard a lie) ==="
psql -c "SELECT count(*) AS total,
                count(*) FILTER (WHERE gated_at IS NOT NULL) AS judged,
                count(*) FILTER (WHERE gated_at IS NULL)     AS UNJUDGED,
                max(gated_at)::timestamp(0) AS last_verdict
           FROM provocations WHERE domain='$DOMAIN';"

echo "=== 2. scorecard ==="
psql -c "SELECT CASE WHEN source_ref LIKE '%must-approve half%' THEN 'must APPROVE (9 real)'
                     ELSE 'must REJECT (4 bad)' END AS half,
                status, count(*)
           FROM provocations
          WHERE domain='$DOMAIN' AND gated_at IS NOT NULL
          GROUP BY 1,2 ORDER BY 1,2;"

echo "=== 3. every rejection, with its FATAL reasons ==="
# The reason key is 'rule', NOT 'code' — a scorer reading ->>'code' prints an
# empty column for every rejection and looks like a gate that gave no reasons.
# 'judge_ran' matters when comparing rounds: a deterministic layer (form/safety)
# rejection cannot vary between rounds, so only judge_ran=true rows are subject
# to the stochasticity of §4a.
psql -c "SELECT slug,
                CASE WHEN source_ref LIKE '%must-approve half%' THEN 'REAL(bad)' ELSE 'bad(ok)' END AS half,
                (gate_verdict->>'judge_ran')::bool AS judge_ran,
                (SELECT string_agg(DISTINCT r->>'layer', ',') FROM jsonb_array_elements(gate_verdict->'reasons') r
                  WHERE (r->>'fatal')::bool) AS layers,
                (SELECT string_agg(r->>'rule', ',') FROM jsonb_array_elements(gate_verdict->'reasons') r
                  WHERE (r->>'fatal')::bool) AS fatal_rules,
                (SELECT string_agg(r->>'detail', ' // ') FROM jsonb_array_elements(gate_verdict->'reasons') r
                  WHERE (r->>'fatal')::bool) AS detail
           FROM provocations
          WHERE domain='$DOMAIN' AND status='rejected'
          ORDER BY 2, 1;"

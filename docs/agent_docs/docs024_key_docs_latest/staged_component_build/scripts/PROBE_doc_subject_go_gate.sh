#!/bin/bash
# ============================================================================
# PROBE_doc_subject_go_gate.sh — prove that the RUNNING chassis binary's
# doc-subject vocabulary accepts a given subject_type, and read that vocabulary
# out of the running binary rather than inferring it from a build date.
#
# WHY THIS EXISTS. doc_plans/doc_notes.subject_type has TWO enforcement points:
# the DB CHECK constraints and validDocSubjectTypes in
# platform/orchestration/actions/doc_subjects_common.go. Migrations 163 and 184
# each moved one and not the other (bugs_closed/064), and on 2026-07-31 the
# landmine-verifier lane found a third instance live. `\d doc_plans` proves the
# DB half. Nothing proved the Go half — until this.
#
# WHY IT MUST BE A DISPATCH AND NOT A QUERY OR A GREP.
#  * psql cannot reach the gate. `load_doc_context` takes subject_type from STEP
#    CONFIG only (docResolveSubject in write_doc_plan_action.go:136-145 reads
#    config, never input data), so writing a row by hand exercises the DB CHECK
#    and nothing in Go.
#  * A pod-grep cannot see it. The vocabulary entries are short string literals
#    ('component', 'landmine'); Go compiles those to immediate comparisons that
#    never reach rodata, so `grep -c` on the binary returns a number that means
#    nothing either way. The quoted list in the error message is built at
#    RUNTIME by docSubjectTypesQuoted(), which is precisely why probe 2 below
#    can print it.
#
# HOW IT WORKS — two steps in ONE dispatch, and the second is the control:
#   probe_subject : load_doc_context with subject_type=<the type under test>.
#                   Returns the PLAN => the gate accepted it.
#   probe_vocab   : load_doc_context with a deliberately invalid subject_type.
#                   MUST error, and its message is docSubjectTypesQuoted()'s
#                   rendering of validDocSubjectTypes AS COMPILED INTO THE
#                   RUNNING POD. This is the read-out, and it is also the proof
#                   that probe_subject's route was capable of failing.
#   An all-green run with no probe_vocab error is NOT a pass — it means the
#   probe did not run. See §7 of HANDOFF_2026-07-31b: a check that reports
#   health it never measured is this lane's recurring defect class.
#
# NO agent_definitions ROW IS WRITTEN. The workflow travels inline in the
# message: selectWorkflow's Priority 1 (processor.go:922-928) takes
# config.workflow ahead of any DB lookup. If that override ever stops firing,
# the fallback is agent type `generic`, whose whole workflow is a no-op
# `complete` step — so a misfire is inert AND visible, because current_step
# would read `complete` instead of one of the probe step names.
#
# Usage: ./PROBE_doc_subject_go_gate.sh <subject_type> <subject_key>
#   e.g. ./PROBE_doc_subject_go_gate.sh component teaser-reveal-panel
#
# A PLAN for (subject_type, subject_key) need not exist — has_plan=false with no
# error still proves the gate accepted the type. It proves less, though: with a
# PLAN you also see the body travel and the ```criteria fence extract.
# ============================================================================
set -euo pipefail

SUBJECT_TYPE="${1:?subject_type under test, e.g. component}"
SUBJECT_KEY="${2:?subject_key, e.g. teaser-reveal-panel}"
INVALID_TYPE="zzz-probe-invalid-$$"

NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)

echo "== probing the RUNNING binary's doc-subject vocabulary =="
echo "subject_type under test: $SUBJECT_TYPE"
echo "subject_key:             $SUBJECT_KEY"
echo "invalid control type:    $INVALID_TYPE"
echo "correlation:             $CID"
echo ""

echo "-- the two enforcement points, before the probe --"
echo "DB half (what the database enforces):"
$PSQL -t -A -c "SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname='doc_plans_subject_type_check';"
echo "PLAN rows for this subject: $($PSQL -t -A -c "SELECT count(*) FROM doc_plans WHERE subject_type='$SUBJECT_TYPE' AND subject_key='$SUBJECT_KEY' AND is_current;")"
echo "pod image(s): $(kubectl -n $NS get pods -l app=agent-chassis -o jsonpath='{.items[*].spec.containers[0].image}')"
echo ""

PAYLOAD_B64=$(python3 - "$SUBJECT_TYPE" "$SUBJECT_KEY" "$INVALID_TYPE" <<'PY'
import base64, json, sys
subject_type, subject_key, invalid_type = sys.argv[1:4]
msg = {
    "action": "orchestrate",
    "config": {
        # Deliberately a type that does NOT exist in agent_definitions: if the
        # inline override below ever stopped taking precedence, the Priority 2
        # lookup finds nothing and the run degrades to `generic`'s no-op.
        "agent_type": "doc-subject-gate-probe",
        "workflow": {
            "start_step": "probe_subject",
            "processing_mode": "orchestrator",
            "timeout_seconds": 120,
            "steps": {
                "probe_subject": {
                    "action": "load_doc_context",
                    "description": "THE TEST: does the Go gate accept this subject_type?",
                    "config": {
                        "subject_type": subject_type,
                        "subject_key": subject_key,
                        "notes_limit": 3,
                        "error_step": "probe_vocab",
                    },
                    "next_step": "probe_vocab",
                    "output_field": "doc_subject",
                },
                "probe_vocab": {
                    "action": "load_doc_context",
                    "description": "THE CONTROL: must fail, and prints the running binary's vocabulary",
                    "config": {
                        "subject_type": invalid_type,
                        "subject_key": subject_key,
                        "error_step": "finish",
                    },
                    "next_step": "finish",
                    "output_field": "doc_vocab",
                },
                "finish": {
                    "action": "complete_workflow",
                    "description": "probe complete",
                    "config": {"multiple_output_fields": ["doc_subject", "doc_vocab"]},
                },
            },
        },
    },
    "input_data": {"subject_key": subject_key},
}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "publishing to system.agent.generic.requests ..."
kubectl -n kafka run "kcat-dsp-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.agent.generic.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) \
    -H orchestration_name=doc-subject-probe-$SUBJECT_TYPE \
    -H step_name=start \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was published — kcat exits 0 on a drop."
echo "Waiting for the run ..."

for i in $(seq 1 20); do
  sleep 3
  STATUS=$($PSQL -t -A -c "SELECT status||'|'||current_step FROM orchestration_states WHERE correlation_id='$CID';" 2>/dev/null | tr -d '[:space:]' || true)
  [ -n "$STATUS" ] && echo "  [$i] $STATUS"
  case "$STATUS" in completed*|failed*) break ;; esac
done

echo ""
echo "== VERDICT =="
$PSQL -c "
SELECT
  status,
  current_step,
  collected_data->'doc_subject'->>'has_plan'      AS subject_has_plan,
  length(collected_data->'doc_subject'->>'plan_body')  AS plan_body_chars,
  (collected_data->'doc_subject'->>'criteria_json' <> '') AS criteria_extracted,
  collected_data->'__step_error'->>'failed_step'   AS last_failed_step
FROM orchestration_states WHERE correlation_id='$CID';"

echo "-- the running binary's OWN vocabulary read-out (from the invalid-type control) --"
$PSQL -t -A -c "
SELECT collected_data->'__step_error'->>'message'
FROM orchestration_states WHERE correlation_id='$CID';"

echo ""
echo "-- both probe steps, from processing_history: which errored and with what --"
$PSQL -c "
SELECT h->>'step_name' AS step, h->>'action' AS outcome, left(h->>'details', 200) AS details
FROM orchestration_states s, LATERAL jsonb_array_elements(s.processing_history) h
WHERE s.correlation_id='$CID' AND h->>'step_name' LIKE 'probe%'
ORDER BY h->>'timestamp';"

cat <<EOF

== HOW TO READ IT ==
PASS  subject_has_plan is t (or f with NO error against probe_subject), AND
      last_failed_step = 'probe_vocab', AND the read-out above lists
      '$SUBJECT_TYPE'.
FAIL  last_failed_step = 'probe_subject' — the running binary rejects
      '$SUBJECT_TYPE'. Its message names every type it DOES accept.
VOID  current_step = 'complete' with no probe steps in processing_history: the
      inline workflow override did not take and NOTHING was tested.
correlation: $CID
EOF

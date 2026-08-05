#!/bin/bash
# ============================================================================
# DISPATCH_s6_component_run.sh — fire an S6 acceptance run at ONE component
# placement via request_component_browser_run, with a negative control in the
# SAME dispatch.
#
# This is the parameterised version of the script that closed P2 on 2026-08-02
# (correlation e6a258eb-6ba1-44df-b344-16e42443975f: 15/15 passed, negative
# control confirmed red). The original lived only in a session scratchpad and
# was lost to tmp-cleanup within two days — hence this committed copy.
#
# The proven invocation (teaser-reveal-panel on leopardessconsulting.co.uk):
#   ./DISPATCH_s6_component_run.sh \
#     4851f6fc-71cf-4160-a270-e03d6d3e0732 \
#     leopardessconsulting.co.uk \
#     teaser-reveal-panel \
#     ebc2c413-61e2-465e-b22b-9aab0167abc9 \
#     fc505ab2-a991-4421-85e1-fa856f5b7a39
#
# ARGUMENT 5 (BAD_PAGE_ID) IS NOT OPTIONAL, AND ITS CHOICE IS LOAD-BEARING:
# it must be a REAL, ACTIVE page on the SAME site that does NOT carry the
# component. A nonexistent UUID only proves sql.ErrNoRows works; a real page
# with no placement proves the page_components/content_components JOIN is the
# thing doing the rejecting. Find one:
#   SELECT p.id, p.url FROM pages p
#   WHERE p.site_id='<site>' AND p.status='active'
#     AND p.id NOT IN (SELECT pc.page_id FROM page_components pc
#                      JOIN content_components cc ON cc.id=pc.component_id
#                      WHERE cc.function='<function>') LIMIT 3;
#
# PRECONDITIONS (all verified live before the proven run — do the same):
#  1. doc_plans has a CURRENT PLAN with a ```criteria fence for
#     (subject_type='component', subject_key=<function>). No fence = the run
#     SKIPS with needs_criteria and reads as a clean run that asserted nothing.
#  2. The placement row exists and the page is active (placements MOVE —
#     re-verify, never reuse a page_id from a doc):
#       SELECT pc.page_id, p.site_id, p.url, p.status FROM page_components pc
#       JOIN pages p ON p.id=pc.page_id
#       JOIN content_components cc ON cc.id=pc.component_id
#       WHERE cc.function='<function>';
#  3. The running chassis carries request_component_browser_run — pod-grep a
#     LONG marker, don't trust a roll:
#       kubectl -n ai-persona-system exec <pod> -- sh -c \
#         "strings /app/agent-chassis | grep -c request_component_browser_run"
#  4. No chassis pod (re)started within ~300s (spawn silently dropped).
#
# HOW IT WORKS: mirrors tool-acceptance-agent's LIVE workflow (read from
# agent_definitions, not guessed) with subject_type=component and the sibling
# action. The workflow travels inline (selectWorkflow Priority 1,
# processor.go:922-928) — no agent_definitions row, nothing to clean up.
# agent_type is deliberately nonexistent so a misfire falls through to
# `generic`'s no-op complete (inert AND visible). The neg_control step's
# error_step points at neg_control_confirmed_red — the must-fail arm's error
# IS the pass, same shape as PROBE_doc_subject_go_gate.sh.
#
# READ THE RESULT HONESTLY (RUNBOOK §10): a FAILED run still reports
# status=COMPLETED; the real error is collected_data->'__step_error'. Landing
# on neg_control_confirmed_red is the SUCCESS path. Check the skip reasons,
# never the count: 'not run on profile X' is fence gating; '<type> not
# implemented' is a DEFECT that reads as PASS.
# ============================================================================
set -euo pipefail

SITE_ID="${1:?site_id (uuid)}"
DOMAIN="${2:?domain, e.g. leopardessconsulting.co.uk}"
FUNCTION="${3:?content_components.function == doc_plans.subject_key}"
GOOD_PAGE_ID="${4:?page_id carrying the component (re-verify the placement first)}"
BAD_PAGE_ID="${5:?REAL active page_id on the SAME site WITHOUT the component (see header)}"

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

PAYLOAD_B64=$(python3 - "$SITE_ID" "$DOMAIN" "$FUNCTION" "$GOOD_PAGE_ID" "$BAD_PAGE_ID" <<'PY'
import base64, json, sys
site_id, domain, function, good_page_id, bad_page_id = sys.argv[1:6]

msg = {
    "action": "orchestrate",
    "config": {
        # Deliberately nonexistent: proves the inline workflow override took.
        "agent_type": "component-acceptance-probe",
        "workflow": {
            "start_step": "ensure_site_record",
            "processing_mode": "orchestrator",
            "timeout_seconds": 600,
            "steps": {
                "ensure_site_record": {
                    "action": "ensure_site_record",
                    "config": {"store_brief_in_content_data": False},
                    "next_step": "load_docs",
                    "output_field": "site_record",
                },
                "load_docs": {
                    "action": "load_doc_context",
                    "config": {
                        "subject_type": "component",
                        "subject_key_field": "input_data.spec.function",
                        "error_step": "complete_error",
                    },
                    "next_step": "request_run",
                    "output_field": "doc_context",
                },
                "request_run": {
                    "action": "request_component_browser_run",
                    "config": {
                        "profiles": ["desktop", "mobile"],
                        "function_field": "input_data.spec.function",
                        "criteria_field": "doc_context.criteria_json",
                        "site_id_field": "site_record.site_id",
                        "domain_field": "site_record.domain",
                        "page_id_field": "input_data.spec.page_id",
                        "error_step": "complete_error",
                    },
                    "next_step": "judge",
                    "output_field": "browser_run",
                },
                "judge": {
                    "action": "judge_acceptance_results",
                    "config": {
                        "results_field": "browser_run",
                        "function_field": "input_data.spec.function",
                        "criteria_field": "doc_context.criteria_json",
                        "site_id_field": "site_record.site_id",
                        "error_step": "complete_error",
                    },
                    "next_step": "neg_control",
                    "output_field": "acceptance_verdict",
                },
                # THE CONTROL: page_id deliberately points at a real, active
                # page that does NOT carry this component. MUST error — that
                # is what proves the placement JOIN is actually enforced.
                "neg_control": {
                    "action": "request_component_browser_run",
                    "config": {
                        "function_field": "input_data.spec.function",
                        "criteria_field": "doc_context.criteria_json",
                        "site_id_field": "site_record.site_id",
                        "domain_field": "site_record.domain",
                        "page_id_field": "input_data.spec.bad_page_id",
                        "error_step": "neg_control_confirmed_red",
                    },
                    "next_step": "neg_control_UNEXPECTEDLY_GREEN",
                    "output_field": "neg_control_result",
                },
                "neg_control_confirmed_red": {
                    "action": "complete_workflow",
                    "config": {
                        "success_message": "PASS: negative control correctly rejected a page_id with no placement",
                        "multiple_output_fields": ["acceptance_verdict", "browser_run"],
                    },
                },
                "neg_control_UNEXPECTEDLY_GREEN": {
                    "action": "complete_workflow",
                    "config": {
                        "success_message": "FAIL: negative control did NOT error — placement check is not enforced",
                        "multiple_output_fields": ["acceptance_verdict", "browser_run", "neg_control_result"],
                    },
                },
                "complete_error": {
                    "action": "complete_workflow",
                    "config": {
                        "success_message": "Component acceptance probe completed with errors",
                        "multiple_output_fields": ["acceptance_verdict", "browser_run"],
                    },
                },
            },
        },
    },
    "input_data": {
        "site_id": site_id,
        "domain": domain,
        "spec": {
            "function": function,
            "page_id": good_page_id,
            "bad_page_id": bad_page_id,
        },
    },
}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "function:     $FUNCTION"
echo "good page_id: $GOOD_PAGE_ID"
echo "bad page_id:  $BAD_PAGE_ID"
echo "correlation:  $CID"
echo "publishing to system.agent.generic.requests ..."

kubectl -n kafka run "kcat-cbr-$(date +%s)-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.agent.generic.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) \
    -H orchestration_name=component-browser-run-probe \
    -H step_name=start \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was published — re-run now."
echo "CID=$CID"
echo ""
echo "Follow it (SUCCESS path lands on neg_control_confirmed_red):"
echo "  SELECT current_step, status FROM orchestration_states WHERE correlation_id='$CID';"
echo "Real run summary + judge verdict + the control's error:"
echo "  SELECT jsonb_pretty(collected_data->'browser_run'->'response'->'summary'),"
echo "         jsonb_pretty(collected_data->'acceptance_verdict'),"
echo "         jsonb_pretty(collected_data->'__step_error')"
echo "    FROM orchestration_states WHERE correlation_id='$CID';"
# RUNBOOK — bugfix 247

## Pre-edit go/no-go check (shared tree — re-run immediately before every edit session)

```
grep -rn "\.processRequest(\|\.selectWorkflowOLD(\|\.LoadAgentConfig(" --include=*.go .
grep -rn "configLoader" --include=*.go .
grep -rn "\.determineWorkflowMode(\|\.determineWorkflowModeOLD(\|\.isComplexRequest(\|\.getDefaultTaskWorkflow(" --include=*.go .
```
Expected: no hits outside function definitions and the doomed internal call chain. Any hit
elsewhere = stop, re-scope the deletion.

## Post-edit verification

```
go build ./...
go vet ./platform/messaging/... ./platform/config/... ./internal/agents/contentcreator/...
go test ./platform/messaging/... ./internal/agents/contentcreator/...
go test -race ./platform/messaging/...
grep -rn "processRequest\b\|selectWorkflowOLD\|LoadAgentConfig\|determineWorkflowMode\|isComplexRequest\|getDefaultTaskWorkflow" --include=*.go .
```
Final grep must return only `platform/agentbase/agent.go`'s `processRequests` (plural)
lines — nothing else.

## Council submission

```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
Save the printed `SUBMISSION_CORR`. Budget ~30 min for the round to actually run (queue
latency, not verdict time). Verdict lookup:
```
SELECT current_step, status FROM orchestration_states WHERE
 collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

## Commit

```
git commit platform/messaging/processor.go platform/config/agent_config_loader.go \
  -m "$(cat <<'EOF'
...
Council-Submitted: <corr>
EOF
)"
```

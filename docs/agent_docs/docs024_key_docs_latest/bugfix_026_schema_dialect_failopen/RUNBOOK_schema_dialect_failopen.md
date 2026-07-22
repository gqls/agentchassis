# RUNBOOK — bugs_open/026 schema-dialect fail-open

Commands that were hard to get right, with the gotcha attached. Change them HERE when they
change, not in scrollback.

## Ground the dialect distribution (is the old shape extinct?)
```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT count(*)                                                                        AS total,
       count(*) FILTER (WHERE input_schema ? 'fields')                                 AS v2,
       count(*) FILTER (WHERE input_schema ? 'properties' AND NOT (input_schema ? 'fields')) AS legacy,
       count(*) FILTER (WHERE input_schema::text IN ('{}','null'))                     AS empty
FROM content_components;"
```
2026-07-21: total 173 / v2 124 / legacy 0 / empty 42 (+7 bare example-value rows). **Re-run
before quoting "extinct" — do not carry the figure forward unchecked.**

## Inspect the 7 bare example-value rows (must stay ok=false in the fix)
```
... -c "SELECT function, left(input_schema::text,120) FROM content_components
        WHERE NOT (input_schema ? 'fields') AND NOT (input_schema ? 'properties')
          AND input_schema::text NOT IN ('{}','null') ORDER BY function;"
```
Expect: call-to-action, features, footer, head, header, hero, social-proof — all bare
`{"field":"string"}` maps with no requiredness. `schemaContentFields` returns ok=false for
these; do NOT let a change start projecting them.

## Test the fix locally (shared tree may not compile — scope to the package)
```
gofmt -l platform/orchestration/actions/component_schema_fields.go \
         platform/orchestration/actions/component_schema_fields_test.go \
         platform/orchestration/actions/json_envelope.go \
         platform/orchestration/actions/plan_sections_action.go     # empty = formatted
go build ./platform/orchestration/actions/
go test  ./platform/orchestration/actions/ -run 'SchemaContentFields|MissingRequiredLLMFields' -v
```
The existing `TestMissingRequiredLLMFields` (v2 path) MUST stay green — that is the
no-regression proof.

## Council gate
Submission JSON schema (the trap that cost a retry): `plan` is an **object** with
`plan.summary` (string), `plan.edits` (array ≤8, each needs file/operation/rationale/sketch),
`plan.grounded_in` (array). NOT a top-level `plan` array + top-level `grounded_in`.

**`operation` vocabulary (second trap):** exactly `modify | add | remove | config_change`
(`diagnose_persist_fix_plan_action.go:80`). A NEW file is **`add`**, NOT `create` — a `create`
completes the run at step `complete_invalid` with a persist-time reject *before* any reviewer
runs (no credits, no verdict), so it looks like a silent failure. Check
`collected_data->'__step_error'` on a `complete_invalid` run.
```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
This run: `SUBMISSION_CORR=a85c1220-7174-41fe-8892-64009eadcf47`, orch `27c84581-91ca-4bf7-8c87-11dec63a3b9f`.

Verdict (keyed on SUBMISSION_CORR — always resolves; budget ~30 min, a missing row is queue
latency, not a drop):
```
... -c "SELECT current_step, status FROM orchestration_states
        WHERE collected_data->'input_data'->>'fix_correlation_id' = 'a85c1220-7174-41fe-8892-64009eadcf47';"
... -c "SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
        WHERE correlation_id='a85c1220-7174-41fe-8892-64009eadcf47' AND kind='council_report' ORDER BY created_at;"
... -c "SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;"
```
APPROVED → the follow-on build commit carries trailer `Council-Reviewed: a85c1220-7174-41fe-8892-64009eadcf47`.
REVISE → resubmit with `RESUBMIT_CORR=a85c1220-...` so the trail accumulates.

## Build / deploy / verify (after APPROVED — Go is inert until the roll)
```
# commit first (build is from committed HEAD), bump IMAGE_TAG (~makefile line 16)
make build-agent-chassis
make push-agent-chassis deploy-agent-chassis        # ships whatever is tagged IMAGE_TAG
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "schemaContentFields"'   # symbol won't survive; use a literal instead
```
**Better verify:** the fix is a pure read-path change with no new literal string. Verify
BEHAVIOURALLY, not by pod-grep: seed a scratch component in the legacy dialect with a required
llm field, render it with that field empty, and confirm the gate refuses (log:
"refusing to render an empty section" names the field). A pod-grep of an unchanged marker
proves deployment, not this fix.

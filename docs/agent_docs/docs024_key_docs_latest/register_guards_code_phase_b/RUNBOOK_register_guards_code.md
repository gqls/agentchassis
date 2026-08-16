# RUNBOOK — register↔tool fact drift (Phase B)

Every command here had a gotcha attached the first time. Read the gotcha, not just
the command.

## Build and test on this shared tree

**`go build ./...` at the working tree may fail for reasons that are not yours.** On
2026-08-16 another session's untracked `component_name_resolver_menu.go` did not
compile. Build against a clean HEAD plus only your own files:

```bash
SC=<scratchpad>; rm -rf $SC/headtree; mkdir -p $SC/headtree
git archive HEAD | tar -x -C $SC/headtree
for f in <your changed files>; do cp $f $SC/headtree/$f; done
cd $SC/headtree && go build ./... && go test ./platform/orchestration/... -count=1
```

⚠ `go vet` reports a **pre-existing** `unreachable code` in
`load_component_library_actions.go`. It is at HEAD and not yours — check before
chasing it.

## Prove a guard can actually fail (do not skip this)

A passing test proves nothing about a guard until you have watched it fail. In the
headtree copy, delete the guard, run the one test, expect FAIL, restore:

```bash
cp $F /tmp/orig                      # keep the original
python3 - <<'PY' ... remove one guard ... PY
go test ./platform/orchestration/actions/ -run TestClassifyFactDrift_NonForkRoutesToHuman -count=1   # must FAIL
cp /tmp/orig $F                      # restore, re-run, must pass
```

Six routing guards were proven this way on 2026-08-16 (no_auto_fix, fork,
evidence-vs-value, fetch-error, baseline-present, baseline-precedence), plus P11.

## Measure the fleet — always with a positive control

The `?` jsonb operator silently returns 0 for a key you spelled wrong, which is
indistinguishable from "nothing has one". Run the control in the same breath:

```sql
SELECT count(*) FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
 WHERE ss.is_current AND ss.aspect='evidence_base' AND f->'source' ? 'artifact_check';  -- 0
SELECT count(*) FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
 WHERE ss.is_current AND ss.aspect='evidence_base' AND f->'source' ? 'citation';        -- 61 ← control
```

## Run the ladder's eligibility predicate yourself

Before assuming a tool is visible to any acceptance machinery — it very often is not.
The predicate lives in `discovery_checks/tool_eligibility.go` (`toolEligibilityWhere`);
paste it into a query scoped to the site. On 2026-08-16 it returned neither
stamp-duty tool.

## Does a tool have a PLAN, and what does its fence say?

```sql
SELECT subject_key, is_current, created_by,
       body ~ '```criteria' AS has_fence,
       body ~ 'no_auto_fix"?: ?true' AS no_auto_fix,
       body LIKE '%"facts"%' AS declares_facts
FROM doc_plans WHERE subject_type='tool' AND subject_key IN ('stamp-duty','mortgages-stamp-duty');
```

⚠ **The subject key is not the page name.** mortgagecalculator's page is
`tool-stamp-duty`; its PLAN key is `stamp-duty`. LMC's page and key are both
`mortgages-stamp-duty`.

⚠ **Never hand-edit a `doc_plans` body to add `facts`.** Both lanes' `install_fences.py`
rewrite the whole body on `--apply`, so a hand-added key is lost on the next install.
Add it to the lane's criteria JSON and re-install.

## Fire a one-off dry run of the sweep

Not within 300 s of a chassis pod restart (the spawn is silently dropped). Publish to
`system.agent.generic.requests`, ONE line of JSON, with an inline workflow — `dry_run`
is read from STEP config, not `input_data`:

```json
{"action":"orchestrate","config":{"workflow":{"start_step":"refresh_evidence","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"refresh_evidence":{"action":"refresh_evidence_base","config":{"dry_run":true},"next_step":"complete","output_field":"refresh_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["refresh_result"]}}}}},"input_data":{"site_id":"<site>"}}
```

Read it back by payload, not by a printed id:

```sql
SELECT status, jsonb_pretty(collected_data->'refresh_result')
FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%'
ORDER BY created_at DESC LIMIT 1;
```

A dry run **plans** the fan-out and marks each emission `dry_run`; it writes nothing.

## Induce the fan-out (the only proof that matters)

Supersede the fact, dry-run, restore. `pinned` must be carried forward (CLM-001), and
check `writer_block_managed` first — if true, the daily sweep will regenerate the
writer block with your test number, so keep the window short and do it outside
09:00–09:10 UTC (the sweep's own CAS window).

**Expected:** `fact_drift` names `stamp-duty`, `kind: value_drift`,
`route: fact_drift_review`, `reason: no_auto_fix`. **A dry run that reports nothing
after a real change is the failure.**

## Prove the code is live

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -ac fact_drift_review /proc/1/exe   # expect >0
kubectl -n ai-persona-system exec $POD -- grep -ac stale_attestation /proc/1/exe   # positive control
```

Never `strings` (absent from the image), never a discovery grep for "some 40-hex
string", and always run the control in the same exec.

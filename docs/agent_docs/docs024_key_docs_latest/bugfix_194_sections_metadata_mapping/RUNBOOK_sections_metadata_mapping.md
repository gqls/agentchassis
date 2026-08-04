# RUNBOOK — `bugs_open/194` lane

Every command that was hard to get right, with its gotcha attached. Fix it HERE, not in
scrollback.

## R1 — census every `save_page_sections` caller

**Gotcha:** the step is nested in a loop `sub_workflow` in four of the six callers, so the
usual top-level `jsonb_each(default_config->'workflow'->'steps')` finds only two of them and
reads as "the fleet is fine". Use `jsonb_path_query(..., '$.**.steps')`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT ad.type, s.key, s.value->'config'->>'sections_metadata_field', s.value->'config'->>'html_field'
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '\$.**.steps') AS steps,
LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'save_page_sections'
ORDER BY ad.type;"
```

**Second gotcha (LANDMINE, 016b):** `default_config::text LIKE '%save_page_sections%'` is NOT
a test for this step — `council-gate` and `fix-proposer` both "contain" the string in prompt
text and neither has the step. Match on `value->>'action'`.

## R2 — a caller's step graph (what data is actually in scope at the save)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT s.key, s.value->>'action', COALESCE(s.value->>'output_field','')
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '\$.**.steps') AS steps,
LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.type='<agent>' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;"
```

Read the `output_field` of the writer step: that prefix plus `.response.<key>` is what a
config path must name. `page_content` → `page_content.response.sections_metadata`.

## R3 — is an agent dormant? Do NOT ask `orchestration_states`

**Gotcha:** terminal rows are reaped ~daily. `min(created_at)` over the whole table says
weeks only because unreaped non-terminal statuses set the floor; bound per status and you see
`COMPLETED` goes back one day. Use `agent_run_stats` (no reaper, spans from 2026-07-26):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT agent_type, run_count, first_ran_at::date, last_ran_at FROM agent_run_stats
WHERE agent_type IN ('pageflow-builder','site-work-orchestrator','tool-recreation-handler',
                     'page-build-handler','page-rebuild','page-rerender')
ORDER BY last_ran_at DESC NULLS LAST;"
```

**Sanity check before trusting an absence:** confirm the table tracks agents of the same
SHAPE as the one you are calling dormant (orchestrators, not just leaves) —
`SELECT agent_type, run_count FROM agent_run_stats ORDER BY run_count DESC LIMIT 25;`

## R4 — the state of a page's `content_data` (the bug's own verification query)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT slot_name, length(rendered_html), length(content_data::text), updated_at
FROM page_components WHERE page_id='<uuid>' ORDER BY position;"
```

**Gotcha:** `length(content_data::text)` = 0 is impossible; NULL prints as empty in `-tA`, so
an empty column IS the NULL. Add `content_data IS NULL` explicitly if the output is going
into a claim.

## R5 — offline test of the action

```bash
timeout 400 go test ./platform/orchestration/actions/ -run 'TestSavePageSections' -v 2>&1 | tail -30
gofmt -l platform/orchestration/actions/
```

## R6 — prove the Go half shipped (never trust the tag, the roll or a status)

The council's `debug_historian` seat objected — rightly — that a 24h `agent_error_log`
query tests BEHAVIOUR, not DEPLOYMENT. Both are needed and they are different checks.

```bash
for POD in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name); do
  echo "== $POD"
  # POSITIVE: a LONG literal from the new message. Short strings compile to
  # immediates and grep 0 on a binary that fully supports them.
  kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c \
    "grep -ac 'loses the only thing the' /app/agent-chassis"      # expect >0
  # DISCRIMINATING NEGATIVE: this change REMOVES no string, so there is no natural
  # negative control. A near-miss literal is the substitute — it proves the grep
  # discriminates, where the positive alone only proves the pipeline.
  kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c \
    "grep -ac 'CONTENT_DATA_REGRESSION_V2' /app/agent-chassis"    # expect 0
done
```

`grep -ac` not `strings | grep -c`: some images have no `strings`. Every replica, same
exec — `logs deploy/X` reads one pod of N, and a roll can leave one replica behind.
No orchestration dispatch within ~300s of a pod (re)start; the spawn is silently dropped.

## R7 — the two post-roll checks, with their disconfirming outcomes stated first

**Acceptance.** `site-work-orchestrator` is directly dispatchable —
`scripts/initial_messages/170_work_item_flow_build/075d_simple_maintain_trigger.sh` — so
one of the two dormant callers CAN be proven live. Target a site whose pages are
`rebuild_policy != 'owned'` (check the column first: `087`'s run was blocked by exactly
that guard). Then R4 on the rebuilt page.

- **Pass:** `content_data` non-NULL on every row at the new run's `updated_at`, **AND** the
  save step's result carries `sections_source: 'metadata'`.
- **Why both halves:** `content_data` can also arrive via the interactive carry-forward
  (Layer 2), so the bare column check is a **false pass**. The run must say which route it
  took, and it now does.
- **Disconfirming:** still NULL, or `sections_source: 'html_parse'` — the writer's reply is
  not reaching the save on that path, and the key name is not the fault.

**No-regression, 24h after the roll.**

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT agent_type, count(*), max(created_at) FROM agent_error_log
WHERE error_code='CONTENT_DATA_REGRESSION' GROUP BY 1 ORDER BY 2 DESC;"
```

- **Pass:** zero rows for `page-build-handler` and `page-rerender`.
- **Disconfirming:** any `page-rerender` row — the report's predicate is misconceived
  (~320 runs/day would flood it), and the follow-up `require_sections_metadata` opt-in must
  NOT proceed until that is understood.

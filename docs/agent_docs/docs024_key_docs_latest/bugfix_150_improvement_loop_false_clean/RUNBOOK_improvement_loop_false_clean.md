# RUNBOOK — bugs_open/150

Every command here had to be got right once. The gotcha is attached to the command, not
kept in someone's scrollback.

## Read the live branch (never the seed — the seed is history, the row is fact)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'check_has_findings')
FROM agent_definitions
WHERE type='improvement-loop'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Gotcha:** without all three of `is_active`, `COALESCE(is_snapshot,false)=false` and
`deleted_at IS NULL` you may read a snapshot row and conclude the config is something it is
not. `snapshot_agent()` writes those rows on every migration that follows the house
convention, so there are several per agent.

## Who else runs the promoter (the whole bug in one query)

```sql
SELECT ad.type, step.key AS step_name, step.value->>'output_field' AS out, step.value->>'next_step' AS next
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') step
WHERE step.value->>'action' = 'triage_detected_items'
  AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
ORDER BY 1;
```

Three rows: `improvement-loop / triage_findings`, `design-audit-agent / triage`,
`site-review-agent / triage`.

## Who else branches on `has_items` (before you think about redefining it)

```sql
SELECT ad.type, step.key AS step_name, step.value->'config'->>'condition' AS cond
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') step
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND step.value->'config'->>'condition' LIKE '%has_items%'
ORDER BY 1;
```

Four rows, three actions. Three of them are correct about their own loaders. That result is
the reason the fix adds a key instead of changing one.

## Fire ONE improvement sweep, at a site you choose

```bash
./docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/run_improvement_sweep_once.sh <site_id> <domain>
```

**Read its blast-radius header first.** One firing runs five LLM agents on the site, and
promotes **every** `detected` item of every type — which `build-pipeline-trigger` then
dispatches to real fixers within ~120s, on a live site.

**Gotchas:**
- `PUBLISH_OK` in the output or nothing was sent, whatever the exit code.
- Not within ~300s of an agent-chassis pod restart, or the spawn inside the workflow is
  silently dropped.
- **Find your run by `owner_agent_type`, not by the printed orchestration id and not by
  site** — that site is likely to have several other orchestrations running at once (the
  control run of 2026-07-31 shared its site with three page builds):

```sql
SELECT orchestration_id, current_step, status, created_at
FROM orchestration_states WHERE owner_agent_type='improvement-loop'
ORDER BY created_at DESC LIMIT 5;
```

## Read what the run actually decided

```sql
SELECT current_step, status,
       collected_data->'triage_result'                                  AS parent_triage,
       collected_data->'call_design_audit'->'response'->'triage_result' AS child_audit,
       collected_data->'call_site_review'->'response'->'triage_result'  AS child_review
FROM orchestration_states WHERE orchestration_id='<id>';
```

The bug is `parent_triage.promoted = 0` while a child's is non-zero, with
`current_step = complete_clean`. **Do not read the parent's row alone** — on its own it is
indistinguishable from a genuinely clean site.

## Did the run promote, and did the closing rerender happen?

```sql
SELECT count(*) FROM site_work_items
WHERE site_id='<site>' AND triaged_at > '<run start>';

SELECT count(*) FROM site_work_items
WHERE site_id='<site>' AND item_key LIKE 'improvement_rerender%' AND created_at > '<run start>';
```

Non-zero followed by zero is the defect, measured. `triaged_at` is the right clock here:
`created_at` includes items the discovery half filed in the same run without promoting.

## Site state, for choosing a target

```sql
SELECT s.domain, get_audit_pass_count(s.id) AS passes,
       count(*) FILTER (WHERE w.status='detected') AS detected,
       count(*) FILTER (WHERE w.status IN ('triaged','approved') AND w.pipeline='build') AS actionable
FROM sites s LEFT JOIN site_work_items w ON w.site_id=s.id
GROUP BY 1,2 ORDER BY 3 DESC;
```

**`passes >= 3` short-circuits the whole loop** to `complete_clean` before triage ever runs
(`check_audit_pass_limit`), so such a site cannot exercise the branch under test. All sites
were at 0 on 2026-07-31; check, do not assume.

**These counts move under you.** vetcomparison.uk went 2 → 12 actionable in the forty
minutes between planning the run and firing it.

## Gate the deploy at the running pod, both replicas

```bash
for P in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  echo -n "$P "
  kubectl -n ai-persona-system exec "$P" -- sh -c \
    'strings /app/agent-chassis | grep -c "site_dispatchable"; \
     strings /app/agent-chassis | grep -c "TriageDetectedItemsAction: Starting"'
done
```

First number is the symbol this change added; second is the pre-existing control. **A 0 in
the second means the grep is broken and the first number means nothing.** A roll is not
evidence the fix shipped (`bugs_open/153`).

## Apply the config half — only after the grep above passes on EVERY replica

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/281_improvement_loop_branches_on_site_state.sql
```

The file is a single transaction with a pre-flight assertion, a `snapshot_agent()` call, a
guarded UPDATE that refuses a drifted row, and a verify-before-commit. Expect
`rows_to_change_expect_1 = 1`; anything else means the row is not the shape the migration
was written against.

## Go side

```bash
go build ./platform/...
go test ./platform/orchestration/actions/ -run 'Triage|ImprovementLoopCondition|PreUpgrade' -v
```

**Gotcha:** `-run` with a pattern that matches nothing still prints `ok`. Grep the `--- PASS`
lines for the six test names, or a quiet pass will look like a working guard.

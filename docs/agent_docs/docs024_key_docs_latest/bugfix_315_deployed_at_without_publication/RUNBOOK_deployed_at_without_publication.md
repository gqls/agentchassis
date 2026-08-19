# RUNBOOK — bug 315, publish-seam evidence

Every command here had a gotcha. The gotcha is attached to the command, not kept separately.

## Ordering claims about a workflow — join on `next_step`, never read the key order

`jsonb_each` returns steps in arbitrary order, and reading them by eye gets the sequence WRONG
(`page-build-handler` prints `deploy_page` above `update_status` and actually runs it after).

```sql
WITH steps AS (
  SELECT ad.type AS agent, e.key AS step, e.value->>'action' AS action,
         COALESCE(e.value->>'next_step', e.value->'on_success'->>'next_step') AS next,
         e.value->'config'->>'status' AS status
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') e
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
)
SELECT u.agent, u.step, u.status,
       (SELECT string_agg(p.step||'('||p.action||')', ', ')
          FROM steps p WHERE p.agent=u.agent AND p.next=u.step) AS preceded_by
FROM steps u WHERE u.action='update_page_status' ORDER BY u.agent;
```

⚠ `default_config->'workflow'->'steps'` is an **object keyed by step name**, not an array.
`jsonb_path_query_array(default_config,'$.workflow.steps[*].name')` returns `[]` and looks like
"this agent has no steps" — it does not; you asked an array question of an object.

⚠ Always carry all three of `is_active` / `is_snapshot=false` / `deleted_at IS NULL`. Snapshots are
real rows and will double your counts.

## Is a schema column actually written? (the check that turned this bug around)

Two separate questions; ask both, because either alone misleads.

```sql
SELECT count(*) AS total, count(content_hash) AS populated FROM pages;          -- 786 / 0
SELECT count(*) AS total, count(deploy_commit) AS populated FROM page_components; -- 1775 / 0
```
```bash
# and the writer side — note --include and the ABSOLUTE path (see the pwd trap below)
grep -rn "deploy_commit" --include=*.go /home/ant/projects/agentchassis | grep -v _test
```

⚠ **Run the grep INCLUDING tests first.** "No non-test writer" and "no code at all mentions it" are
different findings; the second is much stronger and is what was true here.

## ⚠ The Bash working directory persists between calls

A `cd` inside one compound command changes the directory for every later call. Three greps returned
empty and were read as absences; they were run from a docs subdirectory. **Use absolute paths, and
never `cd` in a compound command.** Cost: nearly missed `docs026_concept_register/register/
deployment-github.md`, the document that names the whole delivery mechanism.

## Grading at the artefact (the only layer that is evidence)

```bash
curl -sI "https://${domain}${url}?cb=$RANDOM$RANDOM" --max-time 20 | grep -iE '^(HTTP|last-modified|cf-cache-status)'
```

⚠ **Always cache-bust.** `cf-cache-status: DYNAMIC` is the confirmation you read the origin.
⚠ `last-modified` is the **per-object** write time — pages on different domains carry different
values, which is the control proving it is not one global checkout mtime.
⚠ **A batch of pages sharing a `last-modified` to the second is NORMAL**, not a coincidence: the
origin is rewritten per changed domain directory by one `b2 sync`.

## Sizing "deployed_at is stale against the origin" — and why the number lies

```bash
# pages.txt: domain|url|deployed_at  from a psql -At -F'|' query
while IFS='|' read -r domain url dep; do
  lm=$(curl -sI "https://${domain}${url}?cb=$RANDOM$RANDOM" --max-time 25 | grep -i '^last-modified' | sed 's/^[^:]*: //I' | tr -d '\r')
  echo -e "$domain\t$url\t$dep\t$lm\t$(( $(date -u -d "$lm" +%s) - $(date -u -d "$dep" +%s) ))"
done < pages.txt
```

⚠ **This returned 40 of 40 "stale" and that is NOT 40 defects.** The origin lags the commit by tens
of minutes in whole-domain batches, so at any instant most correctly-behaving pages are "stale"
against their own stamp. A 40/40 result is the shape that should send you to the raw values — there,
all 40 shared ONE three-second `last-modified` window, which is the batch, not 40 failures.
**This comparison cannot separate "not synced yet" from "will never sync"; only elapsed time can,
and the known bad case took six hours.**

## Did the commit actually happen, and did the runner deploy it?

```sql
SELECT updated_at,
       collected_data->'deploy_result'->'response'->'data'->>'success'   AS ok,
       collected_data->'deploy_result'->'response'->'data'->>'file_path' AS path,
       collected_data->'deploy_result'->'metadata'->>'status'            AS meta_status
FROM orchestration_states
WHERE collected_data ? 'deploy_result'
  AND collected_data->'deploy_result'->'response'->'data'->>'domain' = '<domain>'
  AND updated_at > NOW() - INTERVAL '70 minutes'
ORDER BY updated_at DESC;
```
```bash
kubectl -n ai-persona-system logs github-actions-runner-54fd5c8547-<pod> --tail=40 | grep -E 'Running job|completed with result'
```

⚠ **`deploy_result` has TWO SHAPES.** Inline deploys sit at `deploy_result.response.data.*`;
deploys done by a called sub-agent are nested one level deeper at
`deploy_result.response.deploy_result.response.data.*`. The query above sees only the first and
reports the other **7.7% of runs (57 of 744 over 7 days)** as having no verdict at all.
⚠ Deploy jobs arrive in **clusters 25–50 minutes apart**, so "no job for 36 minutes" is inside the
normal spacing and is not evidence of a stall.

## Council gate / diagnosis loop refusing with a usage limit

```sql
SELECT date_trunc('minute',occurred_at) m, agent_type, count(*)
FROM agent_error_log
WHERE occurred_at > NOW() - INTERVAL '30 minutes' AND error_message ILIKE '%usage limit%'
GROUP BY 1,2 ORDER BY 1;
```

⚠ The message says *"You will regain access on 2026-09-01"* and reads like a hard lockout. It is
not: the same error appears on five separate days over the past month, and on the day I hit it the
council gate was completing `complete_approved` / `complete_revise` runs **in the same minutes** as
other calls were being refused. **Check whether the fleet is still completing LLM work before
reporting an outage** — `orchestration_states WHERE orchestration_name ILIKE '%council%'` is the
cheap read.

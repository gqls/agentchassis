# RUNBOOK — bugs_open/079 phantom link gate

Every query/command that was hard to get right, with its gotcha attached.

## R1 — The census 079 demanded (which builds would a severity change have blocked)

The bug file assumed `collected_data` is pruned at ~24h. It is not — `orchestration_states`
reaches back 13 days, so the census is possible. **Gotcha:** the table's PK is
`orchestration_id`; there is no `id` column, and `count(DISTINCT o.id)` fails.

```sql
SELECT i->>'type' AS type, i->>'severity' AS sev, count(*) AS issue_instances,
       count(DISTINCT o.orchestration_id) AS runs
FROM orchestration_states o,
     jsonb_array_elements(o.collected_data->'validation_result'->'issues') i
WHERE jsonb_typeof(o.collected_data->'validation_result'->'issues')='array'
GROUP BY 1,2 ORDER BY 3 DESC;
```

Denominator — how many builds were validated at all:

```sql
SELECT count(*) FILTER (WHERE collected_data ? 'validation_result')            AS runs_with_validation,
       count(*) FILTER (WHERE (collected_data->'validation_result'->>'valid')='true')  AS valid_true,
       count(*) FILTER (WHERE jsonb_array_length(
             COALESCE(collected_data->'validation_result'->'issues','[]'::jsonb)) > 0)  AS runs_with_issues
FROM orchestration_states;
```

## R2 — Classify each phantom: repairable rewrite vs pure invention

This is the query that decides between fix candidates. **Gotcha:** the domain is not a
column — it is `collected_data->'site_record'->>'domain'`, and `client_id` is `'system'` on
every one of these rows, so it identifies nothing.

```sql
WITH iss AS (
  SELECT o.collected_data->'site_record'->>'domain' AS domain, i->>'value' AS href
  FROM orchestration_states o,
       jsonb_array_elements(o.collected_data->'validation_result'->'issues') i
)
SELECT DISTINCT iss.domain, iss.href,
  EXISTS (SELECT 1 FROM pages p JOIN sites s ON s.id=p.site_id
          WHERE s.domain=iss.domain AND p.url = iss.href || '.html') AS exists_with_html,
  EXISTS (SELECT 1 FROM pages p JOIN sites s ON s.id=p.site_id
          WHERE s.domain=iss.domain AND p.url = iss.href)            AS exists_exact
FROM iss ORDER BY 1,2;
```

Result 2026-07-26: 15 of 15 false/false — every phantom was a pure invention.

## R3 — Did the writer receive its link constraints? (the upstream finding)

`prepare_link_context` runs inside the **page-content-writer's own** orchestration, not the
page-build-handler's — so it is absent from the parent's `collected_data` and you will
conclude the step never ran. Query the children:

```sql
SELECT count(*) AS runs,
       count(*) FILTER (WHERE (collected_data->'link_context'->>'page_count')::int = 0) AS zero_pages
FROM orchestration_states WHERE collected_data ? 'link_context';
```

2026-07-26: 20 runs, 20 with zero pages.

## R4 — Prove a test can fail before trusting it

The probe that caught a vacuous run. Back the file up **outside** the repo first; the tree
is shared and forward-only, so no stash, no reset.

```bash
SP=<scratchpad>
cp platform/orchestration/datahelpers/link_repair.go $SP/link_repair.go.bak
perl -0pi -e 's/\tmatches := repairAnchorRe\.FindAllStringSubmatchIndex\(html, -1\)/\treturn html, nil \/\/ INDUCED FAULT\n\tmatches := repairAnchorRe.FindAllStringSubmatchIndex(html, -1)/' \
  platform/orchestration/datahelpers/link_repair.go
go test ./platform/orchestration/datahelpers/ -run TestRepairPageLinks -count=1
cp $SP/link_repair.go.bak platform/orchestration/datahelpers/link_repair.go
diff -q $SP/link_repair.go.bak platform/orchestration/datahelpers/link_repair.go   # prove the restore
```

**Gotcha, and it is the whole reason to do this:** `go test` without `-v` prints only
failures, so "4 failures" reads as "4 of 8 tests are vacuous". They were not — an unguarded
`repairs[0]` on an empty slice **panicked the test binary**, and the four tests declared
after it never ran at all. Run with `-v` and read `=== RUN` lines, not the FAIL count. Any
`repairs[i]` must be preceded by a `len()` `t.Fatalf`.

## R5 — Pod-grep that discriminates (post-roll)

Grep a string this change CREATED, plus a positive control that predates it.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "CONTENT_LINK_REPAIR_DETAIL"'   # new: must be >= 1
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "CONTENT_VALIDATION_BLOCKER_DETAIL"'  # control: must be >= 1
```

## R6 — The durable record this fix writes

```sql
SELECT created_at, domain, context->>'rewritten' AS rewritten,
       context->>'unlinked' AS unlinked, context->'repairs' AS repairs
FROM agent_error_log
WHERE error_code = 'CONTENT_LINK_REPAIR_DETAIL'
ORDER BY created_at DESC LIMIT 5;
```

## R7 — Verify the fix on the DEPLOYED artefact, not the log

The log row proves the pass ran. It does **not** prove the dead link is gone. Assert the
OLD state is absent:

```sql
SELECT p.url, pc.slot_name
FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'webdesign.co.uk'
  AND pc.rendered_html ~ 'href="/tools/(typography|colour|css)"';   -- must return 0 rows
```

# RUNBOOK — bugs_open/358 (unread finding codes)

Commands that were hard to get right, with the gotcha attached. Update HERE when one changes.

## The census, re-runnable

```sql
-- totals + resolved usage + retention liveness (oldest row sits on the 30d boundary)
SELECT count(*) AS total, count(*) FILTER (WHERE resolved) AS resolved,
       min(occurred_at) AS oldest FROM agent_error_log;

-- per-code counts, newest first by volume
SELECT error_code, count(*) AS n, min(occurred_at)::date AS first,
       max(occurred_at)::date AS last
FROM agent_error_log GROUP BY 1 ORDER BY 2 DESC LIMIT 45;

-- who is using the resolved workflow (first user ever: content-loss-check, 2026-08-22)
SELECT error_code, count(*), array_agg(DISTINCT resolved_by)
FROM agent_error_log WHERE resolved GROUP BY 1;
```

GOTCHA — **a zero over this table proves 30 days of nothing, never "never"** (358 §8):
retention (mig 466 pre_query) truncates history. All-history claims need git or tests.

GOTCHA — **grep the CONSTANT, not just the literal** (358 §3.2): the one real per-code
reader (`page_build_failure_guard.go:131`) binds a Go const to `$1`; a literal-only grep
verdicts the code unread.

GOTCHA — **`error_code` is free text**: uppercase and lowercase families coexist, and
`create_tool_cross_link_items.go` emits colon-suffixed variants
(`tool_crosslink_not_emitted:*`). Any registry or GROUP BY must state its normalisation
or a family double-counts as compliance.

## Reader census greps

```bash
# all writer literals (struct-field form)
grep -rn 'ErrorCode:' platform/ --include='*.go' | grep -v _test.go
# all direct readers of the table (then check each WHERE for a code filter)
grep -rn 'FROM agent_error_log\|from agent_error_log' platform/ cmd/ scripts/ docs/agent_docs/sql_for_agents/
# for each code carried by a const: grep the const NAME too
grep -rn '<ConstName>' platform/ cmd/ --include='*.go'
```

## Ownership / queue checks (before routing work)

```bash
python3 scripts/who-owns.py 358
```
```sql
SELECT item_type, item_key, status FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected')
  AND (summary ILIKE '%agent_error_log%' OR item_key ILIKE '%error_log%');
```

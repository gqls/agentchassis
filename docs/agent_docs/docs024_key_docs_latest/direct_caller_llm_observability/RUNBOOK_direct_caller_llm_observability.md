# RUNBOOK — direct-caller LLM observability

Commands this lane needs, each with the gotcha attached. Add to it the moment a command was hard to
get right — not later.

## Count truncations — and never the obvious way

```sql
-- CORRECT
SELECT agent_type, step_name, count(*) AS truncated
FROM llm_call_log
WHERE error_message ILIKE '%stop_reason=max_tokens%'
  AND created_at > now() - interval '7 days'
GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠ **NEVER `output_tokens >= max_tokens`.** A truncated call has `output_tokens` **NULL** and states
the cut in `error_message`, so that form can never match one: shipped that way on 2026-07-30 it found
**4** fleet-wide where the real number was **94**. It also biased every headroom figure, because
dropping those rows removes the most extreme calls from exactly the steps that truncate most.
The live checks use three spellings, because there are three live wrapper forms:
`'%response truncated:%'`, `'%stop_reason=max_tokens%'`, `'%TRUNCATED and tolerated%'`.

## The two live consumers of the table

```sql
SELECT name, enabled, interval_seconds, last_completed_at
FROM scheduled_tasks WHERE name LIKE '%token-pressure%';
```
`fleet-step-token-pressure` and `council-seat-token-pressure`, both enabled, both every 21600s.
Read their `pre_query` before changing what rows exist — adding rows changes what they report.
Their output lands in `doc_notes`:
```sql
SELECT created_at, source, left(body, 800) FROM doc_notes
WHERE categories ? 'step-token-pressure' ORDER BY created_at DESC LIMIT 1;
```

## Where the budget came from, for a step

```bash
scripts/audit-budget-placement.sh          # human-readable
scripts/audit-budget-placement.sh --json   # the findings JSON
```
Reports declared-vs-effective per live step by calling production's own ladder
(`actions.ResolveStepBudget`). Exit 0 clean / 1 findings / 2 could-not-determine.
Useful here because it answers "what SHOULD this call have sent" for a call that logged nothing.

## Census the concept, not the interface

```bash
grep -rnE '"(max_tokens|max_output_tokens|maxOutputTokens|num_predict|max_completion_tokens)"' \
  --include=*.go . | grep -v _test.go
```
⚠ A census keyed on `\.GenerateText(` cannot see a provider called over raw HTTP. Four censuses of
bug 257 missed `feed_actions.go` that way over three weeks.

## Is my change live?

```bash
kubectl -n ai-persona-system logs -l app=<service> --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <your-commit> <the sha that line reports> && echo SHIPPED
```
⚠ It is a STARTUP line, so it scrolls. An empty result means "not in range", not "unstamped".

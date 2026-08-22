# RUNBOOK — bugfix 337 (token cap / threshold management)

Every command that was hard to get right, with its gotcha.

## Read the live step config (never the seed)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'generate_template'->'config')
FROM agent_definitions
WHERE type='component-creator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
Gotcha: the budget lives at `config->ai_service->max_tokens`; there is NO
`default_config->prompts` for component-creator (0 rows) — the prompt arrives via
`getPromptWithPriority` (agent `prompt_template` / step `prompt_template`), so do not
conclude "no prompt" from the missing key.

## Census: cap-hit failures per step (all history)

```sql
SELECT agent_type, step_name, count(*) AS truncs,
       min(created_at)::date AS first, max(created_at)::date AS last,
       max(max_tokens) AS cap
FROM llm_call_log
WHERE NOT success AND error_message ILIKE '%reached the configured cap%'
GROUP BY 1,2 ORDER BY truncs DESC;
```
Gotcha: filter failures on `success=false`, never on non-empty `error_message` —
successful RETRY rows carry a marker in `error_message` too (bugs_open/119 precedent).

## Census: headroom on SUCCESSFUL calls (the leading indicator)

```sql
SELECT agent_type, step_name, count(*) AS calls, max(max_tokens) AS cap_sent,
       percentile_cont(0.95) WITHIN GROUP (ORDER BY output_tokens) AS p95_out,
       max(output_tokens) AS max_out,
       count(*) FILTER (WHERE output_tokens >= (max_tokens*9)/10) AS ge90pct
FROM llm_call_log
WHERE created_at > now() - interval '14 days'
  AND output_tokens IS NOT NULL AND max_tokens > 0
GROUP BY 1,2
HAVING count(*) FILTER (WHERE output_tokens >= (max_tokens*9)/10) > 0
ORDER BY ge90pct DESC;
```
Gotcha: truncated FAILED calls can carry NULL `output_tokens` (usage is only sometimes
recovered onto the error path), so this census understates pressure — pair it with the
failure census above; neither alone is the picture.

## The parked items (the live loss)

```sql
SELECT swi.id, s.domain, swi.item_key, swi.status, swi.attempt_count, swi.updated_at::date
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE swi.item_type='needs_new_component' AND swi.error ILIKE '%reached the configured cap%'
ORDER BY swi.updated_at DESC;
```

## Open-work coverage check before routing anything at these targets

```sql
SELECT id, item_type, item_key, status FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected','failed')
  AND (item_key ILIKE '%credit-health%' OR summary ILIKE '%max_tokens%');
```

## Re-drive + verify (from bug file §How to verify; recipe origin RUNBOOK_311_fix.md)

Re-drive `loans-credit-health-check` on loanzy.uk, then assert: item completes with
`attempt_count=0`; `content_components` row for the section has `length(html_template)>0`
and contains `</section>`; and the served page gains real controls —
`curl -s https://loanzy.uk/tools/credit-health-check/index.html | grep -c '<input'`
above a pinned "before" of 0. Verify at the ARTEFACT, not the item status.

# RUNBOOK — council-gate cost

Every query here was hard to get right at least once. Gotchas are attached to
the command, not kept somewhere else.

## Where the fleet's LLM money actually goes (24h)

```sql
SELECT agent_type,
       count(*) AS calls,
       sum(input_tokens)  AS in_tok,
       sum(COALESCE(total_output_tokens, output_tokens)) AS out_tok
FROM llm_call_log
WHERE created_at > now() - interval '24 hours'
GROUP BY 1 ORDER BY in_tok DESC NULLS LAST LIMIT 20;
```

⚠ **Parenthesise any `OR` you add to that `WHERE`.** `AND` binds tighter, so
`WHERE created_at > … AND a LIKE 'x' OR a LIKE 'y'` drops the time filter from
the second branch entirely and returns **all-time** counts that look like 24h
ones. This bit me and nearly became the headline of a wrong answer.

⚠ **Positive control before you believe any filtered count**: run the same query
with the filter removed. If the numbers match, your filter isn't filtering.

⚠ `input_tokens` is the **uncached remainder** once any caller uses the cache
breakpoint (LCO-008). True prompt size is
`input_tokens + cache_creation_input_tokens + cache_read_input_tokens`.
A cost query that reads only `input_tokens` will understate by ~95% **in the
flattering direction** — the direction nobody double-checks.

## Per-seat breakdown for one agent

```sql
SELECT model, model_resolved, step_name, count(*) calls,
       sum(input_tokens) in_tok,
       sum(COALESCE(total_output_tokens,output_tokens)) out_tok,
       max(max_tokens) max_tok_req
FROM llm_call_log
WHERE agent_type='council-gate' AND created_at > now() - interval '24 hours'
GROUP BY 1,2,3 ORDER BY in_tok DESC NULLS LAST;
```

## Where a seat's model is REALLY configured

```sql
SELECT s.key AS seat,
       s.value->'config'->'ai_service'->>'model'      AS model,
       (s.value->'config'->'ai_service'->>'max_tokens')::int AS max_tok
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.type='council-gate' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND s.key LIKE 'review_%'
ORDER BY 1;
```

⚠ **Not `config.model`** — that path is empty and reads as "no model set,
inherits a default". The real path is `config.ai_service.model`, and every seat
sets its own. I reported the wrong thing once from the wrong path; a `jsonb`
path that returns NULL looks identical to a key that genuinely isn't set.

## Is a prompt actually cacheable? (the prefix test)

Two seats from the SAME run — different runs interpolate different plans, so
comparing across runs proves nothing:

```sql
WITH one_run AS (
  SELECT orchestration_id FROM llm_call_log
  WHERE agent_type='council-gate' AND created_at > now() - interval '24 hours'
    AND orchestration_id IS NOT NULL
  GROUP BY 1 HAVING count(*) >= 2 ORDER BY max(created_at) DESC LIMIT 1
), two AS (
  SELECT step_name, prompt_rendered, length(prompt_rendered) len,
         row_number() OVER (ORDER BY step_name) rn
  FROM llm_call_log l JOIN one_run r USING (orchestration_id)
  WHERE prompt_rendered IS NOT NULL LIMIT 2
)
SELECT a.len, b.len,
       (SELECT max(n) FROM generate_series(0, LEAST(a.len,b.len), 500) n
        WHERE left(a.prompt_rendered,n) = left(b.prompt_rendered,n)) AS common_prefix_chars
FROM two a, two b WHERE a.rn=1 AND b.rn=2;
```

⚠ `generate_series` starts at 0, and n=0 always matches (empty string), so a
result of **0 means "no common prefix at all"**, not "the query failed".

**Shared body ≠ shared prefix.** To find out whether reordering would help, test
whether slices of A appear ANYWHERE in B:

```sql
SELECT frac, position(substr(a.p, (a.len*frac/100)::int, 400) in b.p) > 0 AS found
FROM two a, two b, generate_series(10,90,10) frac WHERE a.rn=1 AND b.rn=2;
```
All true + zero common prefix = "the content is shared but misaligned" =
reordering unlocks caching. That was exactly the council's situation.

## Prove the cacheable prefix is byte-identical across seats

The single most important check before trusting any of this — if this is not
`1`, every seat writes its own cache entry and reads none, which costs **more**
than no caching and looks like success:

```sql
SELECT count(DISTINCT md5(split_part(s.value->'config'->>'prompt_template',
                                     '<!--CACHE_BREAKPOINT-->', 1))) AS must_be_1
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.type='council-gate' AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND s.key LIKE 'review\_%';
```

## Is a council run in flight? (check BEFORE editing council-gate config)

```sql
SELECT current_step, status, now()-updated_at AS idle
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

⚠ **Never apply migration 377 (or any seat-template edit) while this returns
`EXECUTING_STEP`.** The chain reloads step config as it advances, so one verdict
would be assembled from two prompt generations and nothing downstream would
record that it happened.

⚠ `orchestration_states` has **no `agent_type` column** — don't filter on one.

## Read the verdict

```sql
SELECT created_at, metadata->>'decision'
FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report'
ORDER BY created_at;

-- human-readable
SELECT body FROM doc_notes WHERE categories ? 'council-gate'
ORDER BY created_at DESC LIMIT 1;
```

## Prove caching actually works after the roll

```sql
SELECT step_name, input_tokens, cache_creation_input_tokens, cache_read_input_tokens
FROM llm_call_log
WHERE agent_type='council-gate' AND orchestration_id='<a run AFTER the roll>'
ORDER BY created_at;
```

Expect: **first** seat `cache_creation > 0`, `cache_read = 0`; **every
subsequent** seat `cache_read > 0` and `input_tokens` collapsed from ~100k to
~5k.

⚠ **A zero in `cache_read` across the whole run is THE failure mode**, not the
absence of one — it means something above the marker varies per seat and every
call is paying the write premium for nothing. Assert a NON-ZERO read on the 2nd+
seat before claiming this works. NULL in these columns means the pod predates
the change (i.e. the roll didn't reach it), which is a different problem again.

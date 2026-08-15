# RUNBOOK — bugfix 271, `content_guidance` has no reader

Every query here was needed to establish a load-bearing fact, and each carries the
gotcha that nearly cost a wrong answer. Run these against the live cluster:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## 1. Which agents mention each guidance spelling

```sql
SELECT type,
       (default_config::text LIKE '%rewrite_guidance%') AS mentions_rewrite_guidance,
       (default_config::text LIKE '%content_guidance%') AS mentions_content_guidance,
       (default_config::text LIKE '%suggestion%')       AS mentions_suggestion
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (default_config::text LIKE '%rewrite_guidance%' OR default_config::text LIKE '%content_guidance%')
ORDER BY 1;
```

Result 2026-08-15: `content-gap-planner` (content_guidance + suggestion),
`page-build-handler` (rewrite_guidance + suggestion), `page-content-writer`
(rewrite_guidance only).

> **GOTCHA.** A `default_config::text LIKE` hit proves the STRING is somewhere in
> the config. It does **not** prove anything reads it — `content-gap-planner`'s
> hit is its own PROMPT asking the LLM for the key, which is the false affordance
> the bug is about. Always follow a hit to the step that consumes it (§3).

## 2. Every live consumer of an item-spec path

Descends only top-level steps — see the gotcha, this is how the dispatcher hid.

```sql
SELECT a.type, s.key AS step, kv.key AS mapped_as, kv.value #>> '{}' AS src
FROM agent_definitions a,
     jsonb_each(a.default_config->'workflow'->'steps') s,
     jsonb_each(s.value->'config'->'input_mapping') kv
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND kv.value #>> '{}' LIKE '%spec.suggestion%'
ORDER BY 1,2;
```

Result 2026-08-15: exactly one row — `page-build-handler | call_content_writer |
rewrite_guidance? | input_data.spec.suggestion`.

> **GOTCHA — this query cannot see a step inside a loop.** `jsonb_each` over
> `workflow->steps` descends ONE level; every dispatcher on this platform nests
> its real work under `config.sub_workflow.steps`, so a loop's mappings are
> invisible here and the query returns a confident, incomplete answer. That is
> why "who maps the whole spec into a child run?" came back EMPTY twice before
> the text-pattern query below found it. For a whole-config search, pattern-match
> the text instead:

```sql
SELECT type, substring(default_config::text from '"spec\??": "[a-z_.]*spec"') AS mapping
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text ~ '"spec\??": "[a-z_.]*spec"'
ORDER BY 1;
```

Result: `build-dispatch-loop | "spec": "current_item.spec"` (plus two
pass-throughs). That is the dispatcher, and it is inside a loop.

## 3. Prove the channel reaches an LLM prompt (not just a mapping)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'process_sections_loop')
FROM agent_definitions
WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Then grep the output for `rewrite_guidance`. The live template carries:

```
{{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT: incorporate this into the content)
{{.rewrite_guidance}}
{{end}}
```

> **GOTCHA.** Do not filter on `s.value->'config'->>'prompt'` — the writer's
> prompt lives under a nested key inside the loop step, so that column is NULL
> and the "is it in a prompt?" test returns blank, which reads as "no". Dump the
> step and search the text.

## 4. The census that sizes the damage

```sql
SELECT item_type,
       count(*) AS rows,
       count(*) FILTER (WHERE COALESCE(spec->>'content_guidance','') <> '') AS has_content_guidance,
       count(*) FILTER (WHERE COALESCE(spec->>'suggestion','') <> '')       AS has_suggestion,
       count(*) FILTER (WHERE COALESCE(spec->>'content_guidance','') <> ''
                          AND COALESCE(spec->>'suggestion','') = '')        AS guidance_only
FROM site_work_items
WHERE spec ? 'content_guidance' OR spec ? 'suggestion'
GROUP BY 1 ORDER BY 2 DESC;
```

`guidance_only` is the number that matters: those rows carry a brief no reader
will ever see. Result 2026-08-15: 56 `content_rewrite` + 34 `needs_content_page`.

> **GOTCHA.** Use `COALESCE(spec->>'k','') <> ''`, never `spec ? 'k'`. A key
> present with an empty string is not a brief, and counting it inflates the
> damage figure — the same shape as the LANDMINE about `domain IS NULL` vs `''`.

## 5. Verification at the artefact (for whoever ships the fix)

Sentinel recipe from `bugs_open/271` §6 — a phrase that appears NOWHERE in the
register or existing copy, so a hit cannot come from anywhere else:

```sql
-- Did the guidance reach the rendered prompt?
SELECT id, created_at, left(prompt_rendered, 200)
FROM llm_call_log
WHERE prompt_rendered LIKE '%<SENTINEL PHRASE>%'
ORDER BY created_at DESC LIMIT 5;
```

Negative control, same run: an item with empty guidance must NOT gain a
`## Rewrite Guidance` heading.

```sql
SELECT count(*) FILTER (WHERE prompt_rendered LIKE '%## Rewrite Guidance%') AS with_heading,
       count(*)                                                             AS total
FROM llm_call_log
WHERE agent_type = 'page-content-writer' AND created_at > '<roll time>';
```

> **GOTCHA.** Both greps return nothing TODAY, before any fix — so a zero is
> only evidence once you have shown the sentinel actually travelled on a control
> item that uses the live `suggestion` key. Run the positive control in the same
> batch, or the post-fix zero is indistinguishable from a blind check.

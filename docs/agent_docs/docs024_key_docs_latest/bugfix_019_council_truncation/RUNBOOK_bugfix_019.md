# RUNBOOK — bugfix 019 (council truncation void)

Commands that were hard to get right, with the gotcha attached.

## Which mechanism voided a round (upstream call vs decider parse)

```sql
SELECT collected_data::jsonb->'__step_error'->>'failed_step' AS failed_step,
  CASE WHEN collected_data::jsonb->'__step_error'->>'message' ILIKE '%execute_llm_prompt%'
         THEN 'UPSTREAM execute_llm_prompt (truncated call)'
       WHEN collected_data::jsonb->'__step_error'->>'message' ILIKE '%invalid JSON%'
         THEN 'DOWNSTREAM council_decide (json.Valid)'
       ELSE 'other' END AS mechanism, count(*)
FROM orchestration_states
WHERE current_step='complete_invalid' AND created_at > now() - interval '10 days'
  AND collected_data::jsonb->'__step_error' IS NOT NULL
GROUP BY 1,2 ORDER BY 3 DESC;
```

Gotchas: fix-proposer/feature-designer use `complete_refused`, not
`complete_invalid` — widen `current_step IN (...)` to cover them.
`orchestration_states` has **no agent_type column**; this query cannot tell you
WHOSE runs these are (that mistake is in NOTES). Attribute via `agent_error_log`.

## Attributing truncations to an agent

```sql
SELECT agent_type, step_name, count(*), min(occurred_at)::date, max(occurred_at)::date
FROM agent_error_log
WHERE error_message ILIKE '%stop_reason=max_tokens%'
  AND occurred_at > now() - interval '10 days'
GROUP BY 1,2 ORDER BY 3 DESC;
```

Gotchas: the time column is **`occurred_at`**, not `created_at`. And the
councils log as **`agent_type='generic'`** — filtering by
`council-gate`/`fix-proposer`/`feature-designer` returns nothing (this blinded
the diagnosis loop into UNVERIFIABLE; name `generic` in any council symptom).

## Finding a config value whose path you don't know

```sql
-- probe for existence FIRST — a NULL from an assumed path is not absence
SELECT default_config::text ~ '8000' FROM agent_definitions
WHERE type='council-gate' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Then walk the tree (dump `default_config::text`, walk every key named
`max_tokens` in Python) rather than probing paths. The real path is
`workflow.steps.<seat>.config.ai_service.max_tokens` — the case file's query
missed `->'config'` and read 13 armed seats as unset.

## Verify migration 177 (the opt-in flag)

```sql
SELECT type, count(*) AS seats,
       count(*) FILTER (WHERE v->'config'->>'tolerate_truncation'='true') AS armed
FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') e(k,v)
WHERE type IN ('council-gate','fix-proposer','feature-designer')
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND k LIKE 'review\_%' GROUP BY type ORDER BY type;
-- 2026-07-20 baseline: council-gate 15/15, fix-proposer 15/15, feature-designer 5/5
```

Gotcha: `LIKE 'review\_%'` — the underscore must be escaped or it matches any
character ("reviewX..." would slip in).

**Rollback** (accepted lore miss: 177 was applied without a pre-write backup —
this strip-key UPDATE is the recovery; the key is additive so removal restores
the exact prior state):

```sql
-- mirror of 177's DO block, with #- (delete key) in place of jsonb_set
DO $$
DECLARE r RECORD; step_key TEXT;
BEGIN
  FOR r IN SELECT id, default_config FROM agent_definitions
    WHERE type IN ('council-gate','fix-proposer','feature-designer')
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  LOOP
    FOR step_key IN SELECT k FROM jsonb_object_keys(r.default_config->'workflow'->'steps') AS k
      WHERE k LIKE 'review\_%'
    LOOP
      UPDATE agent_definitions
      SET default_config = default_config #- ARRAY['workflow','steps',step_key,'config','tolerate_truncation'],
          updated_at = now()
      WHERE id = r.id;
    END LOOP;
  END LOOP;
END $$;
```

## Verify the fix after an image roll

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "tolerate_truncation"'   # >0 = Go half shipped
```

Then a live truncated round should show, instead of `complete_invalid`:
```sql
SELECT metadata FROM diagnosis_artifacts
WHERE kind='council_report' ORDER BY created_at DESC LIMIT 5;
-- look for "unreadable": N > 0, or a review with "degraded": true in body
```
An `approve` with `unreadable > 0` must NEVER appear — that combination is the
safety property; if seen, the downgrade rule is broken.

Cheap reproduction: scratch copy of a council definition with one seat's
`config.ai_service.max_tokens` ~200, submit any valid plan via 097.

## Council submission (097) schema trap

`plan` is an OBJECT (`summary`, `edits[]`, `grounded_in[]` of strings, `risks`),
not an array of edits; each edit wants a `symbol` field. An array-shaped plan
fails client-side with `.plan missing` — misleading wording, the key existed.

## The forensic invariant the fix preserves

A tolerated truncation logs **exactly one** `llm_call_log` row: `success=false`
with the stop_reason in `error_message`. The success-path log is skipped
(`ai_actions.go`, guard on `truncationTolerated`) because a `success=true` row
with `output_tokens == max_tokens` is the pre-008 silent-cut signature and
poisons the headroom queries in the 019 case file.

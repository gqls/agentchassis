# Agent Definitions — Backup, Swap, and Revert Reference (v2)

## Date: 2026-03-25

---

## Emergency Revert (Nuclear Option)

A full backup of all 107 agent definitions was taken on 2026-03-22 before any model swap functions were installed.

**Backup table:** `agent_definitions_backup_20260322`

```sql
-- CHECK FIRST: what's in the backup
SELECT COUNT(*) FROM agent_definitions_backup_20260322;

-- RESTORE
BEGIN;
DELETE FROM agent_definitions;
INSERT INTO agent_definitions SELECT * FROM agent_definitions_backup_20260322;
COMMIT;
```

After restoring, restart chassis pods: `kubectl -n ai-persona-system rollout restart deployment/agent-chassis`

---

## Taking a Fresh Backup

```sql
CREATE TABLE agent_definitions_backup_YYYYMMDD AS SELECT * FROM agent_definitions;

SELECT
    (SELECT COUNT(*) FROM agent_definitions) as live,
    (SELECT COUNT(*) FROM agent_definitions_backup_YYYYMMDD) as backup;
```

---

## Model Swap Functions (migration 083, deployed 2026-03-25)

### swap_agent_model() — the standard way

Takes a snapshot, then swaps the ai_service config in one step:

```sql
SELECT swap_agent_model(
    'briefing-agent',           -- agent type
    'infer_via_llm',            -- step name containing ai_service
    '{"provider": "ollama",
      "model": "mistral-small3.1",
      "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434",
      "max_tokens": 1500}'::jsonb
);
```

Returns the snapshot ID. The snapshot is stored as `is_snapshot=true, is_active=false` in `agent_definitions`.

### revert_agent() — undo a swap

```sql
SELECT revert_agent('briefing-agent');
```

Restores the active definition from the most recent snapshot and deletes the used snapshot.

### snapshot_agent() — manual snapshot without swapping

```sql
SELECT snapshot_agent('page-content-writer');
```

Useful before manual `jsonb_set` changes.

### Viewing snapshots

```sql
SELECT * FROM agent_snapshots;
```

Shows all snapshots with their source agent, model, and provider.

---

## Checking What Models Agents Use

```sql
SELECT
    ad.type,
    s.key as step_name,
    s.value->'config'->'ai_service'->>'provider' as provider,
    s.value->'config'->'ai_service'->>'model' as model
FROM agent_definitions ad,
     jsonb_each(ad.default_config->'workflow'->'steps') s(key, value)
WHERE s.value->'config'->'ai_service' IS NOT NULL
  AND ad.is_active = true
  AND (ad.is_snapshot IS NULL OR ad.is_snapshot = false)
ORDER BY ad.type, s.key;
```

---

## Common Agent Types and Their LLM Steps

| Agent Type | LLM Step Name | Current Model |
|---|---|---|
| briefing-agent | infer_via_llm | claude-haiku-4-5 |
| site-classifier | classify_site | claude-sonnet-4-6 |
| chief-strategist | generate_build_plan | claude-opus-4-6 |
| page-content-writer | generate_content | claude-sonnet-4-6 |
| webdesign-agent | analyze_design | claude-sonnet-4-5 |
| visual-design-auditor | run_visual_llm_audit | claude-sonnet-4-6 |
| content-quality-auditor | run_content_llm_audit | claude-sonnet-4-6 |
| site-review-agent | run_strategic_review | claude-sonnet-4-6 |

---

## Swapping to Ollama

```sql
-- CPU Ollama (Mistral Small 3)
SELECT swap_agent_model(
    'AGENT_TYPE', 'STEP_NAME',
    '{"provider": "ollama", "model": "mistral-small3.1",
      "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434",
      "max_tokens": 1500}'::jsonb
);

-- GPU Ollama (Llama 70B — only when ThunderCompute is running)
SELECT swap_agent_model(
    'AGENT_TYPE', 'STEP_NAME',
    '{"provider": "ollama", "model": "llama3.3:70b",
      "api_url": "http://ollama-gpu.ai-persona-system.svc.cluster.local:11434",
      "max_tokens": 4000}'::jsonb
);
```

### Swapping Back to Anthropic

```sql
SELECT swap_agent_model(
    'AGENT_TYPE', 'STEP_NAME',
    '{"provider": "anthropic", "model": "claude-sonnet-4-6",
      "api_key_env_var": "ANTHROPIC_API_KEY",
      "max_tokens": 2000}'::jsonb
);
```

---

## Evaluating a Swap

```sql
-- After swapping, trigger a build then check llm_call_log:
SELECT agent_type, model, provider, success, latency_ms,
       LEFT(response_text, 100) as preview
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '30 minutes'
ORDER BY created_at DESC LIMIT 20;

-- Compare latency between models
SELECT model, provider,
    COUNT(*) as calls,
    ROUND(AVG(latency_ms)) as avg_ms,
    ROUND(AVG(input_tokens)) as avg_in,
    ROUND(AVG(output_tokens)) as avg_out
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY model, provider
ORDER BY model;
```

---

## Endpoint Health

```sql
-- Check all endpoints
SELECT * FROM ai_endpoint_status;

-- Manually mark an endpoint down/up (for testing)
SELECT update_endpoint_health(
    'http://ollama-gpu.ai-persona-system.svc.cluster.local:11434',
    false,
    'manually marked down for testing'
);
```

---

## Rules

1. Always take a dated backup before batch changes
2. Use `swap_agent_model()` for single-agent swaps — it snapshots automatically
3. Swap one agent at a time, verify via `llm_call_log`, then move to the next
4. Check endpoint health (`ai_endpoint_status`) before swapping to a new endpoint
5. Keep the nuclear revert table until confident in the new setup
6. Don't delete backup tables until the next backup is taken

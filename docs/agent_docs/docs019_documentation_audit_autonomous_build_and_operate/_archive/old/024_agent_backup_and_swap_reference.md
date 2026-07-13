# Agent Definitions — Backup, Swap, and Revert Reference

## Date: 2026-03-22

---

## Emergency Revert (Nuclear Option)

A full backup of all 107 agent definitions was taken on 2026-03-22 before any model swap functions were installed.

**Backup table:** `agent_definitions_backup_20260322`

**To restore everything to the 2026-03-22 state:**

```sql
-- CHECK FIRST: what's in the backup
SELECT COUNT(*) FROM agent_definitions_backup_20260322;
-- Should be 107

-- RESTORE
BEGIN;
DELETE FROM agent_definitions;
INSERT INTO agent_definitions SELECT * FROM agent_definitions_backup_20260322;
COMMIT;

-- VERIFY
SELECT COUNT(*) FROM agent_definitions;
SELECT type, version, LEFT(default_config::text, 80)
FROM agent_definitions
WHERE type IN ('page-content-writer', 'site-classifier', 'webdesign-agent', 'chief-strategist')
ORDER BY type;
```

**After restoring**, restart the chassis pods so they pick up the reverted definitions:

```bash
kubectl -n ai-persona-system rollout restart deployment/agent-chassis
```

Spawned Job pods (content writers, planners, etc.) will use the reverted definitions on their next spawn. Already-running Jobs keep their old config until they complete.

---

## Taking a Fresh Backup

Before making any batch change to agent definitions, take a dated backup:

```sql
-- Replace the date suffix each time
CREATE TABLE agent_definitions_backup_YYYYMMDD AS 
SELECT * FROM agent_definitions;

-- Verify
SELECT 
    (SELECT COUNT(*) FROM agent_definitions) as live,
    (SELECT COUNT(*) FROM agent_definitions_backup_YYYYMMDD) as backup;
```

To see all backups you've taken:

```sql
SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::text))
FROM pg_tables
WHERE tablename LIKE 'agent_definitions_backup%'
ORDER BY tablename;
```

To drop old backups you no longer need:

```sql
DROP TABLE agent_definitions_backup_20260322;
```

---

## Checking What Models Agents Use Now

```sql
-- See every agent step that calls an LLM, with its current model
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

## Swapping a Single Agent's Model (Manual Method)

This is the safest approach — no functions, just SQL you can read and verify.

**1. Check current config:**

```sql
SELECT type, version,
    default_config->'workflow'->'steps'->'generate_content'->'config'->'ai_service' as ai_service
FROM agent_definitions
WHERE type = 'page-content-writer' AND is_active = true;
```

(Replace `generate_content` with whatever step name uses ai_service for that agent.)

**2. Swap the model in that step:**

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,generate_content,config,ai_service}',
    '{
        "provider": "ollama",
        "model": "qwen2.5:14b",
        "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
    }'::jsonb
),
updated_at = NOW()
WHERE type = 'page-content-writer' AND is_active = true;
```

**3. Verify:**

```sql
SELECT default_config->'workflow'->'steps'->'generate_content'->'config'->'ai_service'
FROM agent_definitions
WHERE type = 'page-content-writer' AND is_active = true;
```

**4. To revert just this agent:**

```sql
UPDATE agent_definitions ad
SET default_config = bk.default_config,
    updated_at = NOW()
FROM agent_definitions_backup_20260322 bk
WHERE ad.type = bk.type
  AND ad.type = 'page-content-writer'
  AND ad.is_active = true;
```

---

## Swapping Back to Anthropic

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,STEP_NAME,config,ai_service}',
    '{
        "provider": "anthropic",
        "model": "claude-sonnet-4-6",
        "api_key_env_var": "ANTHROPIC_API_KEY",
        "max_tokens": 2000
    }'::jsonb
),
updated_at = NOW()
WHERE type = 'AGENT_TYPE' AND is_active = true;
```

Replace `STEP_NAME` and `AGENT_TYPE` with the actual values.

---

## Common Agent Types and Their LLM Steps

| Agent Type | LLM Step Name | Current Model | What It Does |
|---|---|---|---|
| site-classifier | classify_site | claude-sonnet-4-6 | Determines site type from domain |
| chief-strategist | generate_build_plan | claude-opus-4-6 | Plans site structure and pages |
| page-content-writer | generate_content (in loop) | claude-sonnet-4-6 | Writes page section content |
| webdesign-agent | analyze_design | claude-sonnet-4-5 | Generates colour/typography spec |
| domain-research-classifier | classify_domain | claude-sonnet-4-6 | Analyses domain for research |
| domain-strategist | generate_strategy | claude-sonnet-4-6 | Revenue model and positioning |
| briefing-agent | infer_via_llm | claude-haiku-4-5 | Answers briefing questionnaire |

(Verify these with the query in "Checking What Models Agents Use Now" — they may have changed since this document was written.)

---

## Evaluating a Model Swap

After swapping a model, trigger a build and check:

```sql
-- Did the LLM calls succeed?
SELECT agent_type, step_name, model, provider, success, latency_ms,
       LEFT(response_text, 100) as preview
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '30 minutes'
ORDER BY created_at DESC
LIMIT 20;

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

If the output quality is poor or errors are high, revert using the per-agent method above or the nuclear option.

---

## Ollama Models Available

These are models you can pull into Ollama and reference in agent configs:

| Model | Size (Q4) | Good For | Pull Command |
|---|---|---|---|
| nomic-embed-text | 274MB | Embeddings only (already loaded) | `ollama pull nomic-embed-text` |
| llama3.1:8b | ~4.7GB | Classification, simple extraction | `ollama pull llama3.1:8b` |
| qwen2.5:14b | ~9GB | Better reasoning, content generation | `ollama pull qwen2.5:14b` |
| qwen2.5:32b | ~20GB | Near-Claude quality for structured tasks | `ollama pull qwen2.5:32b` |

To pull a new model into the running Ollama pods:

```bash
kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- ollama pull qwen2.5:14b
```

Remember: with emptyDir storage, models are lost on pod restart. They'll need to be added to the init container args in the deployment.yaml to persist across restarts.

---

## Rules

1. Always take a dated backup before batch changes
2. Swap one agent at a time, verify, then move to the next
3. Check `llm_call_log` after each swap to confirm calls succeed
4. Keep the nuclear revert table until you're confident in the new setup
5. Don't delete backup tables until the next backup is taken

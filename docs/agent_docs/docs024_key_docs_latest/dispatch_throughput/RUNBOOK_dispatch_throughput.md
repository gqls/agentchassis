# RUNBOOK — dispatch throughput / whole-architecture scale review

Commands and queries this workstream had to get right, with their gotchas. The STARTER's §7
holds the original five recipes; this file carries them forward plus the scale-review set.
DB access: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

## Dispatch meters (from STARTER §7 — gotchas restated)

```sql
-- Concurrency meter: distinct sites claimed PER MINUTE. ⚠ NEVER a 5-minute bucket —
-- 5 ticks × 1 site each reads 2–6 distinct sites and proves nothing.
SELECT date_trunc('minute',claimed_at) m, count(*) claims, count(DISTINCT site_id) sites
FROM site_work_items WHERE claimed_at > now()-interval '6 hours'
GROUP BY 1 ORDER BY 1 DESC LIMIT 30;

-- Throughput: completions/hour over 24h. ⚠ A quiet queue reads as a failed fix —
-- check triaged+approved depth in BOTH windows before concluding (demand control).
SELECT date_trunc('hour', completed_at) hr, count(*) done, count(DISTINCT site_id) sites
FROM site_work_items WHERE completed_at > now()-interval '48 hours' AND status='complete'
GROUP BY 1 ORDER BY 1 DESC;

-- Handler runtime: quote p50/p90, NEVER the mean (one stuck row owns it).
SELECT count(*) n,
  round(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM completed_at-claimed_at))) p50_s,
  round(percentile_cont(0.9) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM completed_at-claimed_at))) p90_s
FROM site_work_items
WHERE status='complete' AND completed_at > now()-interval '7 days' AND claimed_at IS NOT NULL;
```

## Scale baselines (added 2026-08-18/19)

```sql
-- Estate size. ⚠ Do NOT filter sites on status='active' — not in the validated vocabulary
-- (016b §, "enumerate GROUP BY status first"); and the sites table is not the estate
-- (noted.co.uk served live with no row — LANDMINES).
SELECT 'sites_total', count(*)::text FROM sites
UNION ALL SELECT 'work_items_open', count(*)::text FROM site_work_items
  WHERE status NOT IN ('complete','cancelled','rejected')
UNION ALL SELECT 'completions_7d', count(*)::text FROM site_work_items
  WHERE status='complete' AND completed_at > now()-interval '7 days'
UNION ALL SELECT 'llm_calls_24h', count(*)::text FROM llm_call_log
  WHERE created_at > now()-interval '24 hours'
UNION ALL SELECT 'orchestrations_24h', count(*)::text FROM orchestration_states
  WHERE created_at > now()-interval '24 hours'
UNION ALL SELECT 'db_size', pg_size_pretty(pg_database_size('clients_db'));

-- LLM volume by model, 24h. ⚠ input_tokens alone understates cached spend by ~95% since
-- migration 376 — include cache_creation_input_tokens / cache_read_input_tokens for cost.
SELECT model, count(*), sum(input_tokens)::bigint in_tok, sum(output_tokens)::bigint out_tok
FROM llm_call_log WHERE created_at > now()-interval '24 hours'
GROUP BY model ORDER BY 2 DESC;

-- LLM failure / rate-limit shape + latency, 7d.
SELECT count(*) FILTER (WHERE NOT success) fails,
  count(*) FILTER (WHERE error_message ILIKE '%429%' OR error_message ILIKE '%rate%limit%'
                   OR error_message ILIKE '%overloaded%') ratelimited,
  count(*) total FROM llm_call_log WHERE created_at > now()-interval '7 days';
```

```bash
# Kafka topic census (job.* per-spawn topics vs system.*). Broker pods live in ns `kafka`.
# ⚠ The kustomize kafka/ and postgres-clients/ files are ORPHANED — Terraform owns the live
# objects (bugs_open/082 class). Read the cluster, not those files.
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-1 -- \
  bin/kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null | grep -c "^job\."
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-1 -- \
  bin/kafka-topics.sh --bootstrap-server localhost:9092 --list 2>/dev/null | grep -c "^system\."

# Node headroom (the "is compute the constraint" check).
kubectl -n ai-persona-system top nodes
```

## Diagnosis artefacts (090 loop)

```sql
-- Find the run by RUN correlation (the loop mints its own; the intake corr printed first
-- is NOT the artifact key). This workstream's run: a16b82cd-b89a-45d5-b5df-4370c754e2fd
SELECT current_step, status, created_at FROM orchestration_states
WHERE collected_data->'input_data'->>'diagnosis_correlation_id' LIKE 'a16b82cd%'
   OR correlation_id LIKE 'a16b82cd%' ORDER BY created_at DESC;

-- The verdict lands in doc_notes:
SELECT left(body, 2000) FROM doc_notes WHERE body LIKE '%a16b82cd%'
ORDER BY created_at DESC LIMIT 3;
```

## Ownership checks before touching a bug

```bash
scripts/who-owns.py 029          # advisory; reads COMMITS, blind to uncommitted sessions
grep -rn "029" docs/agent_docs/docs024_key_docs_latest/*/HANDOFF*.md | head
```

## Build-cost measurement (the burst-sizing number)

```sql
-- Orchestrations attributable to one site's build window: constrain by TIME window of the
-- build, not by joins you hope exist — orchestration_states carries site linkage only
-- inside collected_data for some agent types. Pick a site with a known build date, then:
SELECT date_trunc('hour', created_at) h, count(*),
       array_agg(DISTINCT agent_type) FILTER (WHERE agent_type IS NOT NULL)
FROM orchestration_states
WHERE created_at BETWEEN '<build_start>' AND '<build_end>'
  AND collected_data::text ILIKE '%<domain>%'
GROUP BY 1 ORDER BY 1;
-- ⚠ collected_data::text ILIKE is expensive and over-matches (mentions ≠ ownership);
-- use it to FIND the window, then attribute by correlation ids found inside it.

-- LLM cost for the same window/domain, split by agent_type:
SELECT agent_type, count(*), sum(input_tokens) in_tok, sum(output_tokens) out_tok,
       sum(cache_creation_input_tokens) cache_w, sum(cache_read_input_tokens) cache_r
FROM llm_call_log
WHERE created_at BETWEEN '<build_start>' AND '<build_end>' AND client_id = '<client>'
GROUP BY 1 ORDER BY 3 DESC;
```

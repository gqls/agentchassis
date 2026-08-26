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

## Concurrency meters that actually measure concurrency (added 2026-08-25 — the §"Dispatch meters" one above does NOT)

⚠ The per-minute `count(DISTINCT site_id)` of CLAIMS at the top of this file is **not a concurrency
meter**: it read ≥2 sites in 27.7% of minutes (max 6) on the nominally-N=1 pre-change window. Claims
are per item, not per turn, and loops outlive minutes. Use `orchestration_states` (retained ~24–27 h —
`min(created_at)` before you trust a window).

```sql
-- SELF-OVERLAP of one scheduled_tasks row (the single-flight refutation, 2026-08-25: 361 pairs).
-- Non-zero on ONE row = the row is not a slot. Swap owner_agent_type for any fire_message task.
WITH t AS (SELECT orchestration_id, collected_data->'input_data'->>'task_name' tn, created_at s, updated_at e
           FROM orchestration_states WHERE owner_agent_type='build-pipeline-trigger'
             AND created_at > now()-interval '24 hours' AND status<>'FAILED')
SELECT a.tn, count(*) self_overlap_pairs, min(b.s-a.s) min_gap
FROM t a JOIN t b ON a.tn=b.tn AND a.orchestration_id<b.orchestration_id AND a.s<b.e AND b.s<a.e AND a.s<b.s
GROUP BY 1;

-- FIRE CADENCE per row (interval 60 + 30 s tick ⇒ p50 90 s, measured).
WITH t AS (SELECT collected_data->'input_data'->>'task_name' tn,
  created_at - lag(created_at) OVER (PARTITION BY collected_data->'input_data'->>'task_name' ORDER BY created_at) gap
  FROM orchestration_states WHERE owner_agent_type='build-pipeline-trigger' AND created_at > now()-interval '24 hours')
SELECT tn, round(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM gap))) p50_s,
       round(percentile_cont(0.9) WITHIN GROUP (ORDER BY EXTRACT(epoch FROM gap))) p90_s FROM t WHERE gap IS NOT NULL GROUP BY 1;

-- LITTLE'S LAW: loops × mean duration ÷ window = mean alive. Run this on ANY "one at a time" claim first.
SELECT count(*) loops, round(avg(EXTRACT(epoch FROM updated_at-created_at))) mean_s,
       round(count(*)*avg(EXTRACT(epoch FROM updated_at-created_at))/EXTRACT(epoch FROM (max(updated_at)-min(created_at))),2) mean_alive
FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop' AND status='COMPLETED' AND created_at > now()-interval '24 hours';

-- ALIVE LOOPS per minute, distribution (2026-08-25: 1–8, mode 1–2).
WITH l AS (SELECT created_at s, updated_at e FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
           AND status='COMPLETED' AND created_at > now()-interval '24 hours'),
mins AS (SELECT generate_series(date_trunc('minute', now()-interval '24 hours'), now()-interval '5 minutes', interval '1 minute') mi),
x AS (SELECT mi, count(*) alive FROM mins LEFT JOIN l ON l.s < mi+interval '1 minute' AND l.e > mi GROUP BY mi)
SELECT alive, count(*) minutes FROM x GROUP BY 1 ORDER BY 1;

-- DOUBLE-HANDLE CENSUS (the induction, on the population): handler orchestrations per work item and
-- overlapping pairs on ONE item. Sequential retries show handlers = attempt_count and overlapped = f.
-- ⚠ A STALE-REAPED handler's updated_at is the REAP stamp, not end-of-life: a stuck handler sits
-- zombie until the reaper fires, and its successor legitimately re-claims the released item inside
-- that window — raw overlap then reads >=1 with the claim invariant intact (first live case
-- 2026-08-26, pair a52ac67f/d0f7ea9e: 2 min of reap lag). Discriminator: status FAILED + error
-- 'Orchestration stale%' on the first-started member, second started minutes-not-seconds later.
-- The VERIFY's 6/7 excludes exactly this shape and reports it as a NOTICE (NOTES 2026-08-26).
WITH loops AS (SELECT orchestration_id FROM orchestration_states WHERE created_at > now()-interval '24 hours' AND owner_agent_type='build-dispatch-loop'),
h AS (SELECT o.orchestration_id, o.collected_data->'input_data'->>'work_item_id' wi, o.created_at s, o.updated_at e
      FROM orchestration_states o JOIN loops l ON l.orchestration_id=o.parent_orchestration_id)
SELECT count(*) handlers, count(DISTINCT wi) items,
  (SELECT count(*) FROM (SELECT wi FROM h WHERE wi IS NOT NULL GROUP BY wi HAVING count(*)>1) q) items_with_2plus_handlers,
  (SELECT count(*) FROM h a JOIN h b ON a.wi=b.wi AND a.orchestration_id<b.orchestration_id AND a.s<b.e AND b.s<a.e) overlapping_pairs_same_item
FROM h;

-- TIME TO FIRST CLAIM per loop (spawn → first handler spawn): decides whether fires spaced N s apart
-- will steer to distinct sites. 2026-08-25: p50 17.7 s, p90 24.2 s, p99 131 s.
-- ⚠ percentile_cont returns double; cast before round() or Postgres refuses round(double, int).
WITH l AS (SELECT orchestration_id, created_at s FROM orchestration_states WHERE created_at > now()-interval '24 hours' AND owner_agent_type='build-dispatch-loop'),
f AS (SELECT l.orchestration_id, EXTRACT(epoch FROM min(h.created_at) - l.s) dt FROM l JOIN orchestration_states h ON h.parent_orchestration_id=l.orchestration_id GROUP BY 1, l.s)
SELECT count(*), round((percentile_cont(0.5) WITHIN GROUP (ORDER BY dt))::numeric,1) p50_s,
       round((percentile_cont(0.9) WITHIN GROUP (ORDER BY dt))::numeric,1) p90_s, count(*) FILTER (WHERE dt > 30) over_30s FROM f;

-- CAUSAL TEST — does a concurrent same-site handler raise the failure rate? (2026-08-25: 1.55% with vs 3.85% without)
WITH loops AS (SELECT orchestration_id, site_id FROM orchestration_states WHERE created_at > now()-interval '24 hours' AND owner_agent_type='build-dispatch-loop'),
h AS (SELECT o.orchestration_id, l.site_id, o.created_at s, o.updated_at e, o.status FROM orchestration_states o JOIN loops l ON l.orchestration_id=o.parent_orchestration_id),
x AS (SELECT h.*, EXISTS (SELECT 1 FROM h b WHERE b.site_id=h.site_id AND b.orchestration_id<>h.orchestration_id AND h.s<b.e AND b.s<h.e) has_partner FROM h)
SELECT has_partner, count(*) handlers, count(*) FILTER (WHERE status='FAILED') failed,
       round(100.0*count(*) FILTER (WHERE status='FAILED')/count(*),2) fail_pct FROM x GROUP BY 1 ORDER BY 1;
```

```bash
# The re-runnable five-assertion check (parity · identity · 0 hardcoded stamps · liveness · 0 double-handles),
# plus the co-pick / lost-claim cost table. Exit 0 = holds; a RAISE names the failing assertion.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql
# Prove it can fire (do this after editing it): invert assertion 1 on a scratch copy and expect a RAISE.
sed 's/IF m >= 2 AND n <> 1 THEN/IF m >= 2 AND n = 1 THEN/' docs/agent_docs/sql_for_agents/584_dispatch_sibling_C_insert_trigger_2_VERIFY.sql > "$SCRATCH/verify_mutant.sql"
```

```sql
-- SIBLING PARITY, one-liner (LANDMINES 2026-08-25): one distinct value per column, or a by-name UPDATE missed a row.
SELECT name, md5(pre_query) pq, interval_seconds, target_topic, timeout_seconds, enabled, input_data
FROM scheduled_tasks WHERE name LIKE 'build-pipeline-trigger%';
```

## Per-site starvation floor (added 2026-08-26 — bugs_open/413; the aggregate meters CANNOT see this)

```sql
-- Hours since last claim for every site with eligible work. The 413 mechanism (a pinned
-- worst-priority old row freezes its site's age and starves younger sites) produces ZERO
-- failures, losses or attempts — only absence — so claims/h and distinct-sites/h read healthy
-- through it. Quote the WORST site, not the mean. ⚠ last_claim reads from site_work_items,
-- a rolling window: a site whose claimed rows have all archived reads NULL — treat NULL as
-- "long ago", not "never", and confirm at orchestration_states loops for the site.
SELECT s.domain,
  (SELECT count(*) FROM site_work_items wi WHERE wi.site_id=s.id
    AND wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
    AND (wi.retry_after IS NULL OR wi.retry_after <= now())) eligible,
  (SELECT min(wi.created_at) FROM site_work_items wi WHERE wi.site_id=s.id
    AND wi.status IN ('triaged','approved')) oldest_eligible,
  (SELECT max(claimed_at) FROM site_work_items c WHERE c.site_id=s.id) last_claim
FROM sites s
WHERE EXISTS (SELECT 1 FROM site_work_items wi WHERE wi.site_id=s.id
   AND wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
   AND (wi.retry_after IS NULL OR wi.retry_after <= now()))
ORDER BY last_claim ASC NULLS FIRST LIMIT 15;

-- PINNED vs VICTIM census (the discriminator is the 391 lane's, 2026-08-26): a site is PINNED
-- when its oldest eligible row falls outside the loader's top-max_items by (priority,
-- created_at) — it wins selection on a row it never loads; a VICTIM's oldest would load fine,
-- the site just never wins. Starvation is POSITIONAL (being behind pins in age order) — a
-- pinned site starves the same way while older pins exist. ⚠ pin status is DYNAMIC (a pin
-- clears when better-priority inflow pauses) — the census is a snapshot, date it.
WITH elig AS (
  SELECT wi.id, wi.site_id, wi.created_at, wi.priority, s.domain
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
  WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[])))
    AND wi.status IN ('triaged','approved') AND wi.attempt_count < wi.max_attempts
    AND (wi.retry_after IS NULL OR wi.retry_after <= now())
    AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')
    AND (wi.depends_on IS NULL OR NOT EXISTS (
          SELECT 1 FROM unnest(wi.depends_on) dep_id
          WHERE dep_id NOT IN (SELECT id FROM site_work_items
                               WHERE site_id=wi.site_id AND status IN ('complete','verified'))))),
ranked AS (
  SELECT e.*,
    row_number() OVER (PARTITION BY site_id ORDER BY created_at, priority, id) age_rank,
    row_number() OVER (PARTITION BY site_id ORDER BY priority, created_at) load_rank
  FROM elig e)
SELECT domain, count(*) eligible, min(created_at) oldest,
  max(load_rank) FILTER (WHERE age_rank=1) oldest_load_rank,
  max(load_rank) FILTER (WHERE age_rank=1) > 5 pinned,
  (SELECT max(claimed_at) FROM site_work_items c WHERE c.site_id=r.site_id) last_claim
FROM ranked r GROUP BY domain, site_id
ORDER BY 3 LIMIT 25;
```

## Council (added 2026-08-25)

```bash
# Resubmit on the SAME correlation so the trail accumulates; DRY_RUN first (validates + scope admission, spends nothing).
DRY_RUN=1 RESUBMIT_CORR=db9b7cbf-7b94-471a-a4cf-26a6679fa47f ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh "$SCRATCH/council_582_584_r3.json"
# Verdict lands in ~30 min; find it by payload, latest first:
#   SELECT created_at, left(body,400) FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%db9b7cbf%' ORDER BY created_at DESC LIMIT 1;
#   SELECT body FROM diagnosis_artifacts WHERE kind='council_report' AND correlation_id LIKE 'db9b7cbf%' ORDER BY created_at DESC LIMIT 1;
# ⚠ the doc_notes header says "(round 1)" on every round — it is a template literal, not the round number; count the reports.
```

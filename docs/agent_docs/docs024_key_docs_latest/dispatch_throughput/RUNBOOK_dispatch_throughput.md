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
-- ⚠ last_claim IS BLIND TO CLAIM-RELEASE CYCLES (added 2026-08-30): a release — deferral OR
-- timeout reset — clears claimed_at, so max(claimed_at) reads STALE on a site that is being
-- claimed and released every few minutes. The 08-30 worked case: mortgagecalculator read
-- "last claim 08-28" while taking 9 loops in 2h on a row deferring on retry_after. The service
-- meter is LOOPS (orchestration_states, within retention), never claimed_at alone — the
-- loops_post_fix column below is load-bearing, not decoration.
-- ⚠ LOCK CONTROL (added 2026-08-27): a LOCKED site (s.locked_at set, no lock-excepted rows)
-- is parked BY DESIGN — the selector must skip it, and it is not starvation. The 08-27 read's
-- worst row (adversecreditmortgage.co.uk, 70 eligible, 27h no claim) was exactly this. The
-- locked_at/except_n columns below are the control — never quote a worst site without them.
-- ⚠ STUCK-CLAIM CONTROL (added 2026-08-27, from bugs_open/414 via 413's addendum; CORRECTED
-- same day): a site with a claimed row whose spawn was DROPPED (zero orchestrations for the
-- id) is EXCLUDED by the busy-skip clause, not out-ordered. ~~unreapable, dark until
-- hand-released~~ — REFUTED at oufe + at the mechanism text: `claimed-item-timeout`
-- (scheduled_tasks, 120s tick) auto-completes evidenced claims >15 min and RESETS the rest at
-- >40 min with backoff, so the dark window is BOUNDED ~40–42 min. A stuck claim OLDER than
-- ~45 min therefore means the timeout task itself is broken/disabled — check the task, don't
-- hand-release first. Same silhouette as 413 from outside. Run this beside every floor read,
-- and BEFORE grading a dark site post-657:
--   SELECT s.domain, wi.item_type, wi.claimed_at FROM site_work_items wi
--   JOIN sites s ON s.id=wi.site_id WHERE wi.status='claimed'
--     AND wi.claimed_at < now() - interval '30 minutes'
--     AND NOT EXISTS (SELECT 1 FROM orchestration_states o
--       WHERE o.collected_data->'input_data'->>'work_item_id' = wi.id::text);
-- (extraction caveat: keys loop-spawned handlers carry; a differently-keyed producer would
--  false-positive — confirm at the site's loop history before acting.)
SELECT s.domain, s.locked_at, COALESCE(array_length(s.lock_except_item_ids,1),0) except_n,
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

## Promotion-layer meters (added 2026-09-02; CORRECTED same day, twice — read the strikes)

> **CORRECTED 2026-09-02 (same day, hours apart — both catches the bugs_open/384 lane's):**
> the census this section FIRST shipped (sites with detected>0 and zero triaged/approved)
> is a WEAK meter that reads ~the whole estate in steady state, twice over: (1) zero
> ELIGIBLE rows at an instant is the NORMAL state of a fast-draining queue (the NOW-census
> trap the 413 file itself warns about — quoted at me an hour after I added the meter);
> (2) `detected` rows with EMPTY handler_agent are the PERMANENT FLAG LAYER by design —
> the promoter's own text: "Flag-only rows (no handler_agent) are NOT here … 'detected' is
> where they belong permanently" (pre_query, scored CTE WHERE + held comment). 1,386 such
> rows / 35 sites on 2026-09-02 are records, not parked work. What caught it: 384's
> every-site control, then reading the promoter's FULL query instead of its first screen.

```sql
-- THE MEANINGFUL promotion meter: handler-BEARING detected rows and their age — the only
-- population the promoter governs. Old rows here mean a door is holding them, and the
-- promoter SAYS WHY on every tick (held_detail in its own output). Interpretation: the
-- doors are deliberate (444/430/454 + bugs_closed/405 lineage) — a held row is a policy
-- outcome to read, not automatically a defect.
SELECT s.domain, wi.item_type, wi.handler_agent, count(*) n, min(wi.created_at)::date oldest
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE wi.status='detected' AND COALESCE(wi.handler_agent,'') <> ''
GROUP BY 1,2,3 ORDER BY min(wi.created_at) LIMIT 15;
SELECT name, enabled, last_triggered_at FROM scheduled_tasks WHERE name='detected-item-promoter';
```

⚠ Adjacent counting trap (384's find): `unresolved` is a TERMINAL status
(workItemTerminalStatuses) that the born-terminal two-strike arm can mint at BIRTH — so the
§"Scale baselines" `work_items_open` figure (status NOT IN complete/cancelled/rejected)
OVERCOUNTS open work by every born-terminal row and by this flag layer. Never read it as
"actionable backlog"; the eligible-count in the floor query is the actionable figure.

## Phase 3 apply + post-checks (added 2026-08-26 — apply ONLY per the HANDOFF gate)

```bash
# Apply (hand-applied _HOLD; refusal-first, snapshotted, guarded; rerun-safe — a replay RAISEs):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db   -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/658_dispatch_phase3_batch_8_HOLD.sql
# Rollback (explicit 5s — NEVER '#-' the keys: Go defaults are 50/20, deletion = 10x batch):
#   ... -f - < docs/agent_docs/sql_for_agents/658_dispatch_phase3_batch_8_ROLLBACK.sql
```

```sql
-- Verify at the artefact, not the migration exit code (and never via updated_at — degenerate):
SELECT default_config#>>'{workflow,steps,load_items,config,max_items}'          AS max_items,
       default_config#>>'{workflow,steps,process_item,config,max_iterations}'  AS max_iterations
FROM agent_definitions WHERE type='build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Probe the CAPABILITY: post-apply loops should show loaded up to 8 and claim_result keys past _4.
SELECT orchestration_id, created_at,
  CASE WHEN jsonb_typeof(collected_data->'pending'->'items')='array'
       THEN jsonb_array_length(collected_data->'pending'->'items') END loaded,
  (SELECT count(*) FROM jsonb_object_keys(collected_data) k WHERE k ~ '^claim_result_[0-9]+$') claims
FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
  AND created_at > '<apply time>' ORDER BY created_at DESC LIMIT 10;

-- collected_data size watch (state.go tripwire: 8 MiB warn / 24 MiB alarm; batch 8 ≈ ×1.6):
SELECT round(avg(pg_column_size(collected_data))/1024) avg_kb,
       round(max(pg_column_size(collected_data))/1024) max_kb, count(*)
FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
  AND created_at > '<apply time>';
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

## 657 — selector↔loader ordering contract (added 2026-08-26, bugs_open/413 fix session)

```bash
# Apply (BY HAND, ≥12:00Z 2026-08-27, after the 24h read + 658 — agreed boundaries; stamp + ping):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work.sql

# Contract check (run after apply, and alongside the daily 584 habit; also re-run after ANY
# edit to load_work_item_actions.go or the trigger/loop rows). ⚠ FAILS BY DESIGN before 657
# is applied (md5 arm) — that failure is its mutation proof, do not widen the md5 list.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work_VERIFY.sql
# Go half of the same lockstep (pins the loader's ORDER BY literal):
go test -run TestLoadWorkItemsOrderingMirrorsTheSelectorWindow ./platform/orchestration/actions/

# Acceptance after apply: §"Per-site starvation floor" above, at +2h and +6h against the
# 09:00Z pre-fix baseline. Disconfirming result: any site with eligible work > ~1h unserved
# while pinned rows exist elsewhere. Quote the WORST site, dated.
```

## Spend governor (D4 / AGOV-013) — the live queries (added 2026-09-03, the day enforcement went on)

Until today this RUNBOOK had **no** governor section: every query lived in NOTES prose or
inside the migrations. These are the ones you actually re-run.

```sql
-- POSTURE, one query: is it on, what level, and is the meter alive.
SELECT gc.enabled, gc.monthly_budget_usd budget, gc.l1_pct, gc.l2_pct, gc.l3_pct,
       gs.shed_level, gs.month, round(gs.mtd_usd,2) mtd_usd, gs.computed_at,
       round(EXTRACT(epoch FROM now()-gs.computed_at)) heartbeat_age_s
FROM governor_config gc, governor_state gs WHERE gc.id=1 AND gs.id=1;
```
⚠ **`computed_at` is the ONLY liveness signal the governor has** — a stale one means the
120 s task is not running, and the level then freezes wherever it was (fail-open: it never
sheds MORE than it was, but it also never sheds when it should). **A single high reading is
not a fault**: interval 120 s + the scheduler's 30 s tick makes ~150 s normal, and one read
of **211 s** on 2026-09-03 self-corrected to 25 s at the next read. Two consecutive reads
over ~300 s is the real signal.

```sql
-- WITHHELD vs STUCK — the one discriminator (bug_historian's r1 gating objection made real).
SELECT class, llm_bearing, count(*) FROM governor_withheld_now GROUP BY 1,2 ORDER BY 1,2;
-- columns: id, site_id, domain, item_type, class, llm_bearing, current_shed_level,
--          created_at, priority, status
```

```sql
-- WIRING CHECK — ⚠ RE-RUN AFTER EVERY RELEASE, not just after touching the governor.
-- A release writes EVERY live agent_definitions row in ONE statement about 70 s before the
-- new pods start (measured 2026-09-03: 208 rows / 203 types, all at 08:56:53.045885Z, with
-- no matching schema_migrations entry). The hand-applied governor clause SURVIVED that write
-- -- but 674 edited the LIVE row and no repo seed carries the clause, so a re-seed that did
-- overwrite it would silently remove the governor's primary gate and nothing else reports it.
SELECT md5(default_config#>>'{workflow,steps,find_dispatchable_site,config,query}') sel_md5,
       (default_config#>>'{workflow,steps,find_dispatchable_site,config,query}' LIKE '%governor_admits%') carries_gov
FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
-- expect: fcbe8821a2a56512911955735796460e / t   (657's VERIFY pins the same md5)
SELECT default_config#>>'{workflow,steps,load_items,config,honour_spend_governor}' load_flag,
       default_config#>>'{workflow,steps,process_item,config,sub_workflow,steps,claim,config,honour_spend_governor}' claim_flag
FROM agent_definitions WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
-- expect: true / true, both jsonb BOOLEAN (a string "true" reads FALSE in Go's .(bool))
```

```bash
# GO HALVES on the CURRENT pods — a roll replaces them, so yesterday's probe proves nothing.
# ⚠ Run the absent-control SEPARATELY: grep cannot stop early on a needle that is not there,
# and combining it with the present-probes times the exec out (seen 2026-09-02 and 09-03).
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  for s in 'governor_admits(' 'spend_governor_shed' 'honour_spend_governor'; do
    kubectl -n ai-persona-system exec "${p#pod/}" -- grep -ac "$s" /proc/1/exe   # expect 1
  done
  kubectl -n ai-persona-system exec "${p#pod/}" -- grep -ac 'governor_forbids(' /proc/1/exe  # expect 0
done
```

```sql
-- THE SHED STAIRCASE, PROVEN WITHOUT WITHHOLDING ANY LIVE WORK. Drives the LIVE selector
-- against synthetic shed levels inside one transaction and ROLLS BACK: readers of
-- governor_state are MVCC-isolated, so live dispatch never sees it. Keep it short — the
-- 120 s state task will block on the row lock until the rollback.
-- ⚠⚠ THE TRAP: the selector ENDS IN `LIMIT 1`, so `SELECT count(*) FROM (<selector>)` is 1
-- at EVERY level and reads as "the governor changes nothing" — a meter that cannot come out
-- otherwise. Strip the trailing LIMIT, and ABORT if the strip does not match.
BEGIN;
DO $$
DECLARE q text; qn text; lvl int; sites int; w int; wclass text;
BEGIN
  SELECT default_config#>>'{workflow,steps,find_dispatchable_site,config,query}' INTO q
  FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY version DESC LIMIT 1;
  qn := regexp_replace(q, 'LIMIT\s+1\s*$', '');
  IF qn = q THEN RAISE EXCEPTION 'ABORT: trailing LIMIT 1 not found — do not trust the count'; END IF;
  IF position('governor_admits' in qn) = 0 THEN RAISE EXCEPTION 'ABORT: no governor clause'; END IF;
  FOR lvl IN 0..3 LOOP
    UPDATE governor_state SET shed_level = lvl WHERE id=1;
    EXECUTE 'SELECT count(*) FROM (' || qn || ') s' INTO sites;
    SELECT count(*) INTO w FROM governor_withheld_now;
    SELECT COALESCE(string_agg(c || ':' || n, ' ' ORDER BY c),'-') INTO wclass FROM (
      SELECT class || CASE WHEN llm_bearing THEN '/llm' ELSE '/free' END c, count(*) n
      FROM governor_withheld_now GROUP BY 1) z;
    RAISE NOTICE 'L% | dispatchable_sites=% | withheld=% [%]', lvl, sites, w, wclass;
  END LOOP;
END $$;
ROLLBACK;
-- Post-rollback control (ALWAYS run it): shed_level back to the real value, withheld back to 0.
SELECT shed_level, enabled, round(mtd_usd,2) mtd, (SELECT count(*) FROM governor_withheld_now) withheld
FROM governor_state gs, governor_config gc WHERE gs.id=1 AND gc.id=1;
-- 2026-09-03 10:47Z result: L0 14 sites/0 withheld · L1 13/51 [maintenance/llm] ·
-- L2 13/112 [+build/llm 61] · L3 13/112 (no research-class item eligible in that window).
-- Sites barely move while items move a lot — correct: shedding is per ITEM_TYPE, so a mixed
-- site stays dispatchable on its llm-free work. That is the council's "withheld, not
-- monopolist" property, measured.
```

**What this canNOT prove** — the Go loader and the claim backstop reading a non-zero level on
live traffic. Only a real shed (at the current burn, ~11 September) or an INDUCED one (lower
the budget briefly, watch L1 fire, restore) does that, and that is option C's gate.

### The induced shed — how today's was run, and the two numbers it produced (added 2026-09-03)

Owner-authorised only: it throttles the live fleet, briefly. Set the budget so MTD crosses a
threshold, poll for the level, hold, restore. **The restore must be on an EXIT trap AND a hard
deadline**, so the window cannot outlive the process that opened it.

- **Size the budget against a read taken MINUTES before, not the daily average.** MTD moved
  $388 → $398.61 in the 20 minutes between sizing and starting (~$35/hour against a daily
  average implying ~$5/hour). A band chosen from a stale MTD can drift into the next band
  mid-window.
- **Pick the level from the DEMAND CONTROL, not from politeness.** Claims by class in the
  preceding hour decide whether an absence can mean anything: at L1 only ~11 claims/hour would
  be silenced, so a zero proves nothing. L2 silenced ~27% of claims and left the llm-free
  `page_rerender` stream running as the positive control that dispatch is alive.
- **⚠ THE GOVERNOR IS ~2× SLOWER THAN ITS INTERVAL, BOTH WAYS.** Measured: onset 156 s, release
  **249 s**, task cadence ~250 s against a stated 120 s (interval 120 + 30 s scheduler tick,
  under load). **The release lag is the surprising half** — the budget was correct again at
  11:29:25Z and 115 items stayed withheld until 11:33:34Z. Do not read "restored" off the
  config row; read it off `shed_level` and `governor_withheld_now`.
- **The measurement that discriminates is the per-loop LOAD census, not claim counts.**
  llm-bearing claims run ~2 per 12 minutes fleet-wide, so before/after claim totals are too thin
  to carry an argument. Every loop in the window is a trial:

```sql
-- What did each dispatch loop actually handle during the window? Expect llm-free ONLY.
WITH l AS (SELECT orchestration_id FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
           AND created_at >= timestamptz '<shed_start>' AND created_at < timestamptz '<shed_end>')
SELECT COALESCE(m.class,'(unmapped)')||CASE WHEN COALESCE(m.llm_bearing,true) THEN '/llm' ELSE '/free' END class,
       count(*) items_handled, count(DISTINCT l.orchestration_id) loops
FROM l JOIN orchestration_states h ON h.parent_orchestration_id = l.orchestration_id
JOIN site_work_items wi ON wi.id = (h.collected_data->'input_data'->>'work_item_id')::uuid
LEFT JOIN governor_work_class_map m ON m.item_type = wi.item_type GROUP BY 1 ORDER BY 2 DESC;
```

- **⚠ AN llm-BEARING CLAIM DURING A SHED IS NOT NECESSARILY A DEFECT — check WHO claimed it.**
  Only `build-dispatch-loop` carries `honour_spend_governor`. `diagnose-dispatch-loop`,
  `report-dispatch-loop` and `zip-deliverable-dispatch` do not, by design, so their claims
  continue at every level. One `needs_diagnosis` claim inside today's L2 window was this, not a
  leak. The census that answers it in one line:
```sql
SELECT type, (default_config::text LIKE '%honour_spend_governor%') has_flag FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND type LIKE '%dispatch%' ORDER BY type;
```

### ⚠ The level-change ALARM does not fire — `bugs_open/459`, open as of 2026-09-03

Do not wait for a `doc_notes` row to tell you the governor moved; there will not be one. Read
`governor_state.shed_level` and `governor_withheld_now`. The A/B that reproduces it in two
minutes (both arms inside `BEGIN … ROLLBACK`, nothing persisted):

```sql
BEGIN;
UPDATE governor_state SET shed_level = 3 WHERE id=1;   -- force old <> new
-- ARM A: paste the LIVE pre_query verbatim        -> level_changed = 0, no note
-- ARM B: the same text with `FOR UPDATE` deleted  -> level_changed = 1, note lands
SELECT count(*) FROM doc_notes WHERE subject_key='spend-governor' AND created_at > now()-interval '2 minutes';
ROLLBACK;
```
Third arm, and it is load-bearing: run the `old` CTE ALONE (no `upd` in the statement) and check
it returns a row whose `shed_level` really does differ from `new.lvl` — otherwise a zero has two
sufficient causes and the A/B cannot tell them apart.

### D4b — the agent-type namespace (mig 751, live 2026-09-03 17:12Z, stage B pending)

```sql
-- Is this agent type governed, and at what level would it shed? Unmapped = ADMITTED, always.
SELECT agent_type, class, llm_bearing, governor_admits_agent(agent_type) admitted_now FROM governor_agent_class_map;
SELECT governor_admits_agent('no-such-agent');   -- must be TRUE at every level: a typo cannot shed the fleet

-- MOVE THE LEVEL (owner decision; one UPDATE). 'maintenance' = L1 (first), 'build' = L2, 'research' = L3 (last).
-- UPDATE governor_agent_class_map SET class='maintenance' WHERE agent_type='council-gate';

-- "My council submission never ran — queued, or withheld?" (meaningful once stage B is live)
SELECT withheld_at, agent_type, shed_level, class FROM governor_withheld_runs_recent
WHERE correlation_id LIKE '<SUBMISSION_CORR>%' ORDER BY withheld_at DESC;
-- 0 rows AND no orchestration row => latency (do not retry; CLAUDE.md). 1 row => withheld, deliberately.

-- All three predicates share ONE comparison — if you ever need to change the ladder, change governor_admits_class only.
SELECT proname FROM pg_proc WHERE proname LIKE 'governor_admits%';   -- expect exactly 3
```
⚠ `--record-only` takes `--note`, not `--notes` (silent "unknown argument", nothing recorded).
⚠ `date -d '2026-09-03 17:12:21'` parses LOCAL time — pass `+00` or `Z`, or your "minutes since" is off by the offset.

### D4b stage B — mig 752 (HELD): apply procedure, the daily check, the induced-L3 probe (added 2026-09-03)

**Preconditions, both owner-shaped:** (1) the owner has confirmed council-gate's shed LEVEL
(`UPDATE governor_agent_class_map SET class='<maintenance|build|research>' WHERE agent_type='council-gate';`
— 'research' = L3 is the seed); (2) council corr `c400d333` has returned APPROVED (or its REVISE
is acted on). Applying IS arming.

```bash
# 1. apply (refusal-first on the live row's whole-workflow md5 8dd74a5b…; snapshot; verify EXECUTEs the gate at L0 and a forced L3)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/752_d4b_governor_council_gate_stage_b_HOLD.sql
# 2. the daily check must now PASS (it FAILS BY DESIGN before apply — do not widen it)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/752_d4b_governor_council_gate_stage_b_VERIFY.sql
# 3. drop the suffix + record in ONE motion (bugs_closed/150): git mv …_HOLD.sql …stage_b.sql; both paths on the commit; then
./scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/752_d4b_governor_council_gate_stage_b.sql --note "…"   # --note, NOT --notes
# 4. CANARY: the next real council submission must route gate -> load_schema_hint and run as before
#    (a $ctx binding failure fails OPEN via error_step — so a wrong key looks like 'ran normally'; check the orchestration's step trail for gate_spend_governor -> route_spend_governor -> load_schema_hint)
# 5. END-TO-END PROOF, exactly as D4's induced shed: drop the budget so shed_level reaches council-gate's class threshold,
#    submit a probe (DRY_RUN=0 097 with a trivial in-scope plan), then read:
#      SELECT current_step, status FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id' = '<probe corr>';   -- expect complete_withheld / COMPLETED
#      SELECT created_at, left(body,160) FROM doc_notes WHERE categories ? 'withheld-run' ORDER BY created_at DESC LIMIT 3;
#    restore the budget; wait a full ~250 s cadence; confirm the next submission runs.
```
**Daily habit gains a fourth line:** 584 VERIFY · 657 VERIFY · governor wiring check · **752 VERIFY**
(6 arms; arm 6 catches a deleted `governor_agent_class_map` row, which would make the live gate a
silent no-op because unmapped = admitted).
**Rollback:** `752_..._HOLD_ROLLBACK.sql` — refuses unless the row is in 752's shape; restores
`load_schema_hint` / 44 steps BYTE-IDENTICAL (md5 asserted) and recreates `governor_withheld_runs`.

### 752 is APPLIED (2026-09-03 21:24Z) and 753 fixed the alarm (21:31Z) — the file names in the sections above have changed
`752_d4b_governor_council_gate_stage_b.sql` (was `_HOLD`), `…_ROLLBACK.sql` (was `_HOLD_ROLLBACK`);
`753_d4_governor_level_change_alarm_fires_again.sql` + `_ROLLBACK`. council-gate class = **maintenance (L1)**.

```sql
-- THE ALARM, now live (753): one row per level change, loud wording, category level-change.
SELECT created_at, left(body,200) FROM doc_notes WHERE subject_key='spend-governor' AND categories ? 'level-change' ORDER BY created_at DESC LIMIT 5;
-- A council submission during a hold: its row ends at complete_withheld (not queued — do not retry)
SELECT status, current_step FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
SELECT created_at, left(body,200) FROM doc_notes WHERE categories ? 'withheld-run' ORDER BY created_at DESC LIMIT 5;
```
```bash
# THE BANNER: what every new session sees (silent at level 0 with council admitted; silent on an expired token — that is NOT level 0)
echo '{}' | python3 scripts/governor-session-start.py
```
⚠ A verify that DELETES a live row to prove fail-open must restore the WHOLE row and assert it — a rolled-back rehearsal cannot exercise a restore (752 committed a half-row; self-healed at the next tick).
⚠ In any verify that EXECUTEs the state task's text: take `pg_advisory_xact_lock(hashtext('spend-governor-state'))` FIRST, or a real tick deadlocks you.

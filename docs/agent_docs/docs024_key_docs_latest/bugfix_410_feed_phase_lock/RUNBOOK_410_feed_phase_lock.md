# RUNBOOK — bugfix_410_feed_phase_lock

DB access (all queries below run through this):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Read the live due machinery (schema first — the column is `interval_seconds`, not `schedule*`)
```sql
\d scheduled_tasks
SELECT name, interval_seconds, enabled, last_triggered_at FROM scheduled_tasks WHERE name='content-feed-refresh';
SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
FROM agent_definitions WHERE type='content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Trigger fire times + per-site orchestrator runs (⚠ orchestration_states prunes at ~2 days)
```sql
SELECT created_at FROM orchestration_states WHERE owner_agent_type='content-feed-trigger' ORDER BY created_at DESC LIMIT 6;
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS') FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > now()-interval '24 hours' ORDER BY o.created_at;
```

## Per-source cadence truth (the artefact, not the status)
```sql
SELECT s.domain, cs.fetch_interval, cs.next_fetch_at, cs.last_fetched_at, cs.error_count
FROM content_sources cs JOIN sites s ON s.id=cs.site_id WHERE cs.is_active ORDER BY s.domain, cs.next_fetch_at;
```

## Deploy sequence (ORDER IS THE POINT — Go first, config second)
1. Commit is in; ride the next chassis roll (or ask the owner — releases are whole-fleet).
2. Prove the roll carries the commit (per SERVICE; startup line scrolls — fall back to binary probe with a known-sha control):
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <this lane's commit sha> <the stamp>   # exit 0 = shipped
```
3. Only then apply, BY HAND (it is `_HOLD` — the runner deliberately skips it):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/653_content_feed_due_lookahead_HOLD.sql
```
   Guards abort it if (a) the cadence is no longer 21600 s (fallback literal stale) or
   (b) the live query is not 556's post-image (someone else changed it — merge, don't force).

## Post-fix acceptance (48 h after both halves live)
```sql
-- every 6h-only site: FOUR distinct run-hours/day (≈02:4x, 08:4x, 14:4x, 20:4x)
SELECT s.domain, string_agg(DISTINCT to_char(o.created_at,'HH24'),',' ORDER BY to_char(o.created_at,'HH24'))
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > now()-interval '24 hours' GROUP BY 1;
-- staleness bound: nothing older than 6h15m while the trigger runs
SELECT s.domain, now()-max(cs.last_fetched_at) AS staleness FROM content_sources cs
JOIN sites s ON s.id=cs.site_id WHERE cs.is_active GROUP BY 1 ORDER BY 2 DESC;
```
Gotchas: exclude remortgagecalculator.uk's 2026-08-26 13:43Z off-cadence run from any
before/after census; cap hits (LCO-009, --capped-schedule-ordering) become NORMAL post-fix
(~12 due vs cap 10) — expected demand, not a regression.

## 090 artifacts for this lane's run
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '15d56c13-2081-431a-ad70-9516c5fcfbc7';
SELECT body FROM doc_notes WHERE body LIKE '%15d56c13%' ORDER BY created_at DESC LIMIT 3;
```

## ⚠ ARMING A TIME-KEYED WATCHER: give `date -d` an EXPLICIT ZONE
> **CORRECTED 2026-08-26** — this section previously said *"local `date -u` and the DB are ~1 h
> apart"*. **FALSE, retracted.** All three clocks agree within 4 s (local `date -u`, the
> postgres container's OS, and `now()` with TimeZone=UTC). The real trap is below; the full
> account is in `WRONG_CALLS.md`.

`date -u -d '<naive timestamp>'` parses the INPUT in **local** time — `-u` only formats OUTPUT.
On a BST box that silently shifts every deadline an hour earlier, and the watcher still fires,
so nothing looks wrong.
```bash
date -ud '2026-08-26 20:53:00' +%s      # 1787773980 = 19:53:00 UTC  ← WRONG, -u did nothing
date -d  '2026-08-26 20:53:00 UTC' +%s  # 1787777580 = 20:53:00 UTC  ← correct
```
**Better still for anything keyed to DB events: don't compute a deadline at all — poll for the
row.** That is what this lane's working watcher does, and it cannot be wrong about time.
```sql
SELECT now();   -- and keep BOTH ends of any age arithmetic on this one clock
```

## Prove the Go half is live WITHOUT the provenance line (it scrolls within hours)
An empty `grep 'build provenance'` means NOT IN RANGE, not unstamped. Probe the capability,
with both controls in the same breath — never `strings`, never a discovery grep:
```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec $POD -- grep -acF 'make_interval(secs => interval_seconds / 2.0)' /proc/1/exe  # expect >=1 (2 today)
kubectl -n ai-persona-system exec $POD -- grep -acF 'interval_seconds / 7.0' /proc/1/exe                         # expect 0
kubectl -n ai-persona-system exec $POD -- grep -acF 'DispatchFeedSourcesAction: dispatched ingester' /proc/1/exe # expect >=1
```

## Confirm the config half is live (independent of the migration's own post-check)
```sql
SELECT CASE WHEN q LIKE '%make_interval(secs => interval_seconds / 2.0)%' THEN 'LOOKAHEAD-PRESENT' ELSE 'ABSENT' END,
       CASE WHEN q LIKE '%ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10' THEN '554+556-INTACT' ELSE 'TAIL-CHANGED' END
FROM (SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query' AS q
      FROM agent_definitions WHERE type='content-feed-trigger' AND is_active
        AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) x;
```
Applied 2026-08-26 ~20:52Z. Pre-change snapshot: `51dd1c59-69e6-4625-baf6-203c35052f18`.

## THE ACCEPTANCE TEST — ⚠ THE FIRST VERSION WAS UNSATISFIABLE. USE THE SECOND.

> **SUPERSEDED 2026-09-02.** The original test read: *"the sites dispatched in tonight's 20:47Z
> pass must ALL reappear at ~02:46Z"* (test set pinned in `HANDOFF` §4). **It cannot be satisfied
> by a working fix**, and the reason was already written four sections further down the same
> document: the look-ahead makes ~all 6h-only sites due at every pass, demand exceeds the query's
> `LIMIT 10`, and the surplus is correctly displaced. Measured 2026-09-02: **4 of 8** discriminating
> sites were capped out of the very pass that proved the fix works. A pass-membership test cannot
> tell "skipped by the phase lock" from "displaced by the cap". `WRONG_CALLS.md` + LANDMINE.
> Also: **the trigger drifts** — ~:46 on 08-26, **~:57** on 09-02, ~11 min in six days. Never
> hardcode a window from a handoff; read the fire times.

**The test that discriminates, and needs no inference.** A site fetched during pass N cannot have
been fetched *before* it was dispatched in pass N, so its next due stamp is **≥ (its pass-N
dispatch time) + fetch_interval**. If pass N+1's trigger fires *before* that bound and the site is
admitted anyway, a bare `next_fetch_at <= NOW()` cannot explain it. Set the two literals from the
fire times you just read; the `interval '6 hours'` must match the site's own interval.

```sql
WITH t AS (SELECT timestamptz '<PASS N+1 FIRE TIME>' AS fired),
prev AS (SELECT s.domain, o.created_at AS d0 FROM orchestration_states o JOIN sites s ON s.id=o.site_id
         WHERE o.owner_agent_type='content-feed-orchestrator'
           AND o.created_at BETWEEN '<PASS N FIRE>' AND '<PASS N FIRE + 1h>'),
cur  AS (SELECT s.domain, o.created_at AS d1 FROM orchestration_states o JOIN sites s ON s.id=o.site_id
         WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > '<PASS N+1 FIRE>'),
six  AS (SELECT s.domain FROM content_sources cs JOIN sites s ON s.id=cs.site_id
         WHERE cs.is_active GROUP BY s.domain
         HAVING bool_and(cs.fetch_interval = interval '6 hours'))
SELECT p.domain, to_char(p.d0,'HH24:MI:SS') AS served_N,
       to_char(p.d0 + interval '6 hours','HH24:MI:SS') AS earliest_possible_due,
       to_char(p.d0 + interval '6 hours' - t.fired,'HH24:MI:SS') AS due_AFTER_trigger_by,
       (p.d0 + interval '6 hours' > t.fired) AS discriminating,
       (c.domain IS NOT NULL) AS served_N_plus_1
FROM prev p CROSS JOIN t JOIN six ON six.domain=p.domain
LEFT JOIN cur c ON c.domain=p.domain ORDER BY p.d0;
```

**Reading it.** `discriminating=false` ⇒ **throw the row away**, it is served under either
predicate (on 09-02 mortgagecalculator's bound fell 4 s the wrong side — the same vacuous shape as
the old prediction (d)). Of the `discriminating=true` rows, **one `served_N_plus_1=true` is a
PASS** — it is arithmetically impossible pre-fix. A `false` proves **nothing on its own**: check
whether the pass was full before concluding anything.

**Close the one gap in the bound** — a straggler source with an older stamp would explain an
admission without the look-ahead. Prefer a single-source site (2026-09-02: `vetcomparison.uk`,
1 source, decisive alone). Otherwise show the sources move together and none is in error backoff:
```sql
SELECT s.domain, count(*) n,
       to_char(max(cs.last_fetched_at)-min(cs.last_fetched_at),'HH24:MI:SS') AS spread,
       count(*) FILTER (WHERE cs.next_fetch_at <> cs.last_fetched_at + cs.fetch_interval) AS off_pattern,
       max(cs.error_count) AS max_err
FROM content_sources cs JOIN sites s ON s.id=cs.site_id
WHERE cs.is_active AND s.domain IN (<the discriminating sites>) GROUP BY s.domain;
```
`off_pattern` and `max_err` must both be **0**, and `spread` seconds, or the bound does not hold.

## Is the cap binding? (run this BEFORE reading any absence as a failure)
```sql
SELECT count(*) FILTER (WHERE six) AS six_only, count(*) AS eligible_total
FROM (SELECT s.id, bool_and(cs.fetch_interval = interval '6 hours') AS six
      FROM content_sources cs JOIN sites s ON s.id=cs.site_id WHERE cs.is_active GROUP BY s.id) x;
```
Compare against the `LIMIT` in the live `find_news_sites` query. `[MEASURED 2026-09-02]` 14
eligible / 12 six-hour-only vs `LIMIT 10` ⇒ 2 slots go to the always-due controls, **8 slots for
12 sites**, so a 6h-only site is served ~2.67 of 4 passes (~9 h effective, not the designed 6 h).
That shortfall is **`bugs_open/316`**, not this bug. ⚠ The `10` is a literal and the site count
grows by addition — "~12" was right on 08-26 and is 14 today.

## Was a pass actually healthy? (a stalled pass mimics the phase lock exactly)
```sql
SELECT to_char(created_at,'MM-DD HH24:MI:SS') fired, status, current_step, left(COALESCE(error,''),200)
FROM orchestration_states WHERE owner_agent_type='content-feed-trigger' ORDER BY created_at;
```
`FAILED` + `current_step=process_sites_iter_1_spawn_orchestrator` + `error` starting `reaper: stale
EXECUTING_STEP` is the spawn→call handshake race (own owner, not 410). Seen 09-01 20:57:41: it
served **1** site and cost the other 13 a whole pass — i.e. **a 12 h gap produced by something that
is not the phase lock**. Exclude such passes from every cadence count, visibly.

# RUNBOOK — bugfix 129, retry replay

Every command here was needed at least once and had a gotcha attached. Newest
gotchas are folded into the command, not appended below it.

---

## 1. Is anyone else on this bug?

```bash
python3 scripts/who-owns.py 129_HANDOFF_2026-07-28_spawned_child_adopts_parents_orchestration_row_and_silently_declines_the_work
```

**Gotcha:** pass the **slug**, not the number. Numbers are ambiguous — several
are reused across two unrelated cases — and `who-owns` matches on the filename.
It reads *commits*, so a session mid-fix with an uncommitted tree is invisible;
check `git status` too.

## 2. The measurements this fix rests on

The split that made the case. **Run it unfiltered first** — the `retry_version > 0`
filter is the question, and without a baseline "430 poisoned" has no denominator:

```sql
SELECT CASE WHEN COALESCE(target_agent_id,'')='' AND requests_topic LIKE 'system.adapter%'
            THEN 'adapter: re-executes step (already correct)'
            ELSE 'agent: synthesised retry (poisoned)' END AS path,
       count(*) AS retried_14d,
       count(*) FILTER (WHERE retry_version=3) AS exhausted
FROM awaited_requests
WHERE sent_at > now() - interval '14 days' AND retry_version > 0
GROUP BY 1;
```

The distribution that shows retries do not recover — **a decaying tail means
retries work; a spike at the cap means they do not**:

```sql
SELECT retry_version, status, count(*) FROM awaited_requests GROUP BY 1,2 ORDER BY 1,3 DESC;
```

**Gotcha — do not read "processed at retry_version ≥ 1" as "the retry worked".**
A late *original* response arriving after the retry was sent produces exactly the
same row. The two are indistinguishable in this table, which is why the claim
carried forward is *"68% exhausted the budget"* and never *"retries never work"*.

Coverage — **which actions can even reach the retry path**, asked of the live
fleet rather than of the code:

```sql
SELECT DISTINCT jsonb_path_query(default_config,'$.workflow.steps.*.action')#>>'{}' AS action
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
ORDER BY 1;
```

**Gotcha:** `#>>'{}'` is load-bearing — `jsonb_path_query` returns `jsonb`, so
without it every action comes back quoted and `WHERE action='call_agent'` matches
nothing while looking like a clean empty result.

## 3. Applying migration 263

**Never `run-migrations.sh --apply` on this tree.** It takes EVERY pending file,
and on 2026-07-28 that was ~20 files belonging to other threads. Dry-run to see
the queue, then apply your own file by hand and record it:

```bash
./scripts/migration/run-migrations.sh            # dry run — READ the list
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/263_awaited_requests_request_payload.sql
./scripts/migration/run-migrations.sh --record-only 263_awaited_requests_request_payload.sql \
  --note "applied by hand; column verified jsonb+nullable via information_schema"
```

Verify against the catalogue, not the psql output — `ALTER TABLE` prints the same
tag whether or not `IF NOT EXISTS` did anything:

```sql
SELECT column_name, data_type, is_nullable FROM information_schema.columns
 WHERE table_name='awaited_requests' AND column_name='request_payload';
```

## 4. The discriminating pod-grep

The obvious positive markers are vacuous on their own. **The best marker is a
string the change DELETED**, with a positive control on the previous image:

| string | v1.0.1192 (before) | v1.0.1193 (after) |
|---|---|---|
| `is_retry` | **1** | **0** |
| `RETRY_PAYLOAD_UNAVAILABLE` | 0 | **1** |
| `RETRY_SELF_ADDRESSED` | 0 | **1** |
| `MISROUTED_REQUEST` | 0 | **1** |

`is_retry` is the stub body the old synthesised retry sent. It is gone from the
binary only if the reconstruction is gone with it.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in is_retry RETRY_PAYLOAD_UNAVAILABLE RETRY_SELF_ADDRESSED MISROUTED_REQUEST; do
  printf '%-28s %s\n' "$s" \
    "$(kubectl exec -n ai-persona-system $POD -- sh -c "strings /app/agent-chassis | grep -c '$s'")"
done
```

**Gotcha — a retag is not a rebuild.** Check the image ID and creation time
before trusting a tag; 1188/1189 once shared one image id, built 56 minutes
*before* the fix they were supposed to carry:

```bash
docker images aqls/agent-chassis --format '{{.Tag}} {{.ID}} {{.CreatedAt}}' | head -3
```

## 5. Proving the fix on the real path

**Induce it. A green happy path proves deployment, not correctness** — the
replay only runs when something times out first.

```bash
./docs/agent_docs/docs024_key_docs_latest/<...>/TRIGGER_code_indexer_v2.sh
```

Then, for the orchestration that results:

```sql
-- the payload is being recorded at all (this is the half that can silently not happen)
SELECT step_name, target_agent_type,
       (request_payload IS NOT NULL) AS payload_recorded,
       request_payload->>'topic' AS replay_topic,
       request_payload->'message'->'headers'->>'orchestration_id' AS child_orch_id,
       request_payload->'message'->'headers'->>'action' AS original_action
FROM awaited_requests
WHERE sent_at > now() - interval '30 minutes'
ORDER BY sent_at DESC LIMIT 20;
```

**The assertion that matters** — `child_orch_id` must NOT equal the awaiting
orchestration's own id. That equality *is* the bug:

```sql
SELECT count(*) AS self_addressed
FROM awaited_requests
WHERE request_payload IS NOT NULL
  AND request_payload->'message'->'headers'->>'orchestration_id' = orchestration_id::text;
-- must be 0, always
```

Fleet-wide coverage of the recording, once a roll has been in for a while — any
step appearing here gets **no retries at all** and needs wiring:

```sql
SELECT step_name, target_agent_type, count(*) AS unrecorded
FROM awaited_requests
WHERE sent_at > now() - interval '2 hours' AND request_payload IS NULL
GROUP BY 1,2 ORDER BY 3 DESC;
```

And the loud failures the refusal emits, which is where an uncovered sender shows
up as an error rather than as silence:

```bash
kubectl logs -n ai-persona-system -l app=agent-chassis --since=1h \
  | grep -E 'RETRY_PAYLOAD_UNAVAILABLE|RETRY_SELF_ADDRESSED|MISROUTED_REQUEST'
```

## 6. Council round

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

**Gotcha — a chassis roll KILLS an in-flight council round**, and
`EXECUTING_STEP` hides it for an hour. On a tree that rolls several times an
evening: build and push freely, but do not `deploy` until your verdict is in.

Find the run by payload, never by the printed id:

```sql
SELECT current_step, status, updated_at FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

**Read `decided_by` and `unreadable`, never `abstained`** — `abstained` counts
seats the relevance filter skipped, which is normal and says nothing about
health; `unreadable` is an opinion that was owed and lost, and ~11% of rounds die
on one seat's unparseable JSON rather than on a judgement.

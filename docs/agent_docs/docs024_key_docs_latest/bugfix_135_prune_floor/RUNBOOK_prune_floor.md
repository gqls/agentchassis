# RUNBOOK — bugs_open/135 prune floor

Every command here had a gotcha attached. The gotcha is the reason the line exists.

## Read the live index (the denominator the guard compares against)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT repo, COALESCE(kind,'(null)') AS kind, commit_sha, count(*) FROM code_symbols GROUP BY 1,2,3 ORDER BY 1,2,4 DESC;"
```

**Gotcha:** `kind` is under a CHECK constraint of eight Go kinds
(`func, method, struct, interface, alias, type, var, const`), so a markdown/doc
row **cannot exist yet** — the D11 markdown-indexing plan has not landed at the
table level. Do not conclude from "0 doc rows" that doc indexing ran and found
nothing; the column would reject it.

## Confirm nothing sets the floor in config (so the code default is what runs)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
 "SELECT jsonb_pretty(default_config->'workflow'->'steps'->'index_symbols'->'config')
    FROM agent_definitions WHERE type='code-indexer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

**Gotcha:** read the LIVE row, never the repo seed — a seed records what the agent
*was*. As of 2026-07-31 the step config carries only `commit_field`,
`analysis_field`, `embedding_service`, so the guard arms at the code default 0.5.
A dead config key looks exactly like a live one; this query is how you tell.

## The prune's own predicate, and the guard's

The DELETE is `commit_sha IS DISTINCT FROM $2`. The guard measures with
`commit_sha IS NOT DISTINCT FROM $2` — the **exact complement**, on purpose.

**Gotcha:** if you re-spell either side (`= $2`, a `COALESCE`, any NULL-safe
rewrite), the guard silently starts judging a population that is not the one being
deleted, and it will still look like it is working. Do not "tidy" one without the
other.

## Verify the deploy at the pod (never at git, never at the tag)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec "$POD" -- sh -c \
 'strings /app/agent-chassis | grep -c "prune_floor_ratio"; echo "--- control:"; strings /app/agent-chassis | grep -c "index_code_symbols: complete"'
```

**Gotcha:** grep a string this change **added** *plus a positive control in the
same exec* (`bugs_open/153`: an image can predate your commit and carries no
provenance; `verify-agent-images` prints all-green on it). A count of 0 for the
new string with a non-zero control means the image is older than the commit — not
that the fix is missing from the source.

## INDUCE THE REFUSAL (the only verification that means anything)

A green run over a healthy repo proves the guard is **inert**, not that it works.
The floor only fires when a cohort falls below it, so the cohort has to be pushed
there deliberately.

```sql
-- 1. Baseline. Note the commit and the interface count.
SELECT commit_sha, count(*) FROM code_symbols WHERE repo='gqls/agentchassis' GROUP BY 1;

-- 2. Push ONE cohort below the floor with synthetic rows at a bogus commit_sha.
--    400 fake interface rows against 33 real ones => that cohort can never reach
--    0.5 even on a perfect run, so the refusal is deterministic, not a race.
INSERT INTO code_symbols (repo, commit_sha, path, symbol, kind, content, content_hash)
SELECT 'gqls/agentchassis', 'deadbeefdeadbeefdeadbeefdeadbeefdeadbeef',
       'zz_synthetic_135/f' || g || '.go', 'Synthetic135_' || g, 'interface',
       'interface Synthetic135_' || g, md5('synthetic135-' || g)
FROM generate_series(1,400) g;

-- 3. Fire the indexer (see below), then read the verdict.
--    EXPECT: prune_status = refused_floor, pruned = 0, all 400 synthetic rows alive.
SELECT collected_data->'index_result'->>'prune_status',
       collected_data->'index_result'->>'pruned',
       collected_data->'index_result'->>'prune_reason'
FROM orchestration_states
WHERE collected_data ? 'index_result' ORDER BY updated_at DESC LIMIT 1;

-- 4. The durable surface must exist.
SELECT created_at, left(body,300) FROM doc_notes
WHERE subject_type='action' AND subject_key='index_code_symbols'
  AND categories ? 'prune-refused' ORDER BY created_at DESC LIMIT 1;

-- 5. CLEAN UP, then re-run: the prune must now proceed normally.
DELETE FROM code_symbols WHERE path LIKE 'zz_synthetic_135/%';
```

**Gotcha (step 2):** `code_symbols` has `NOT NULL` on `path`, `symbol`, `kind`,
`content`, `content_hash` and a unique constraint on `(repo, path, symbol)` — the
synthetic paths must be unique or the INSERT fails half-way. And `kind` must be one
of the eight CHECK values; `'synthetic'` is rejected.

**Gotcha (step 2, the important one):** this induction is **safe even if the guard
does not work**. The DELETE only removes rows whose `commit_sha` differs from the
run's, and today every real row sits at one commit which the run re-confirms — so
the worst case is that the 400 synthetic rows are deleted, which is the cleanup
step anyway. Do not skip the baseline read, though: it is the only way to prove
afterwards that nothing real was lost.

**Gotcha (step 3):** no orchestration dispatch within ~300s of a chassis pod
restart — the spawn is silently dropped. After a roll, wait.

## Fire the code-indexer by hand

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT name, target_agent_type, target_topic, interval_seconds, enabled, last_triggered_at
    FROM scheduled_tasks WHERE target_agent_type IN ('code-indexer','index-orchestrator');"
```

**Gotcha:** if you publish the request yourself, send it to the lane the agent
actually consumes, never `system.agent.generic.requests` by reflex — and
`kubectl run -i … | kcat -P` **sends nothing while exiting 0**, so confirm the row
appeared rather than trusting the exit code.

## Run the tests without another session's WIP in the way

```bash
W=/tmp/head135; rm -rf $W; mkdir -p $W; git archive HEAD | tar -x -C $W
for f in <only your files>; do cp $f $W/$f; done
cd $W && go test ./platform/orchestration/actions/
```

**Gotcha, and it cost real time today:** `go test ./platform/orchestration/actions/`
in the working tree failed to COMPILE, and the cause was another session's
uncommitted refactor sitting in the shared tree, not anything at HEAD. Two
successive breakages (`lockedBrandHeadKeys`'s return type, then a duplicate
`equalStrings` in two test files) both came from untracked/dirty files. `git archive
HEAD` + your own files is the only way to know whose problem it is. See
`WRONG_CALLS.md` for the misdiagnosis this produced.

**Gotcha:** `go build ./...` fails at HEAD for an unrelated reason — two packages
in one directory under `docs/agent_docs/.../traffic_probe/deploy_setup/working_dir`.
Build `./platform/... ./internal/... ./pkg/...` instead. It is not your change.

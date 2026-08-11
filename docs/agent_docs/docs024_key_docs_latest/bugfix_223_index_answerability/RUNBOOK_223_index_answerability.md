# RUNBOOK — bugs_open/223 lane

Every command here had to be got right once. The gotcha is attached to the command.

## Is the bug still valid? (one query, and it is the claim's single point of failure)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT count(*) FILTER (WHERE path LIKE '%.go')     AS go,
       count(*) FILTER (WHERE path NOT LIKE '%.go') AS non_go,
       count(*)                                     AS total,
       count(DISTINCT path)                         AS files
  FROM code_symbols;"
```

**Gotcha:** `code_symbols` has **no `indexed_at` column** — the clocks are `updated_at`
(when the row was written) and `commit_time` (the committer date of the indexed commit).
Asking for `indexed_at` fails the whole statement. Freshness verdicts must key on
`commit_time`; `updated_at` is reset by any refresh (`bugs_closed/108` defect A).

Both counts in one row, with every alternative named, so a future divergence cannot hide
behind a filtered count.

## Which kinds does the index actually hold, versus which are legal?

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;" \
  -c "\d code_symbols"
```

**Gotcha:** the two answers disagree and the disagreement is the finding. The CHECK
constraint permits `type`, `var`, `const`; the table holds none of them. Reading only the
`GROUP BY` tells you what exists; only `\d` tells you what was *intended*.

## Who consumes the seam I am about to change?

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT type, (SELECT string_agg(k,',') FROM jsonb_each(default_config->'workflow'->'steps') e(k,v)
              WHERE v->>'action'='diagnose_code_lookup') AS steps
  FROM agent_definitions
 WHERE default_config::text LIKE '%diagnose_code_lookup%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY type;"
```

**Gotcha:** this finds *config* consumers only. `diagnose_load_runtime_action.go:479`
calls `answerCodeCheck` **in Go**, so it is a fourth consumer that no DB query can see.
Grep the Go call sites as well:

```bash
grep -rn "answerCodeCheck\|diagnose_code_lookup" --include=*.go platform/ internal/ pkg/ | grep -v _test
```

## Read the live verifier definition (never the seed file)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c "
SELECT jsonb_pretty(default_config) FROM agent_definitions
 WHERE type='landmine-verifier' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

**Gotcha:** a repo `SEED_*.sql` records what the agent *was*. Live config drifts. Also
filter `is_snapshot=false` — several agent types carry two active rows and only the
higher version loads.

## Read the verdicts themselves (the artefact, not the status)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT created_at, subject_key, left(body,200) FROM doc_notes
 WHERE categories ? 'landmine-verification' ORDER BY created_at DESC LIMIT 10;"
```

## Which landmine entries have NO verdict at all? (the invisible population)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT DISTINCT n.source FROM doc_notes n WHERE n.categories ? 'landmine'
   AND NOT EXISTS (SELECT 1 FROM doc_notes v
                   WHERE v.categories ? 'landmine-verification' AND v.subject_key = n.source);"
```

**Gotcha (from 223's own recurrence section):** `landmines-sync.py --apply` computes its
`NEEDS_VERIFICATION` list relative to rows *already* in `doc_notes`, so running the sync
directly — which CLAUDE.md instructs — **consumes the signal**, and
`landmines-verify-dispatch.sh` afterwards says "Nothing needs verification" for ever. Run
the dispatch wrapper instead when you want the entry verified.

## Fire the verifier on one entry (acceptance)

```bash
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>' 087_towards_multiple_domains
```

**Gotchas:** (1) the `ref` argument travels into the verdict *body* only — it does not pin
what the lookup reads, which is the index's own snapshot of the last **pushed** tip.
(2) No orchestration dispatch within ~300s of a chassis pod restart — the spawn is
silently dropped. (3) The verdict is async: it lands in `doc_notes`, nothing is awaited.

## Is MY commit in the running binary, when my change has no greppable string?

```bash
kubectl -n ai-persona-system logs <chassis-pod> | grep -m1 'build provenance'
#  → {"msg":"build provenance","git_commit":"55fc8fc3…"}
git merge-base --is-ancestor <my-commit> <that-sha> && echo LIVE || echo NOT LIVE
```

**Gotcha:** this REPLACES the pod-grep for any change that adds no string literal (phase 2
was exactly that — `var`/`const` were already in the binary via `codeKindList`). The stamp
is `bugs_open/153`. Ancestry is the question, not equality: the build sha is whatever HEAD
was at build time, so your commit will usually be an ancestor rather than the stamp itself.
`docker image inspect <img> --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`
answers the same question locally. **Do not try to DISCOVER the sha with a raw
`grep -aoE "[0-9a-f]{40}"`** — unanchored, it matches Go's internal digit table and returns
`000102030405…` on every service with total confidence (153, trap 2).

For a SPAWNED agent, check the pod's own image: `resolveAgentImage` inherits the running
chassis tag unless the row sets `default_config.pin_image_tag`, so verify rather than assume.

```bash
kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].image}{"\n"}{end}' | grep indexer
```

## Re-index, and predict the census BEFORE reading it

```bash
# 1. the ref the indexer will ACTUALLY fetch — the REMOTE tip, not your HEAD
git ls-remote origin 087_towards_multiple_domains
# 2. fire (083c uses index-orchestrator; the in-place code-indexer has no token)
sed "s/^REF='.*'/REF='087_towards_multiple_domains'/" \
  "scripts/initial_messages/310_analysis_adapter/083c_TRIGGER_code_indexer_v2(1).sh" > /tmp/t.sh && bash /tmp/t.sh
# 3. watch by CORRELATION ID (never created_at — shared table, that is a race)
```
```sql
SELECT status, current_step FROM orchestration_states WHERE correlation_id='<corr>'::uuid;
SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 1;
```

**Gotcha 1 — `git rev-parse origin/<branch>` reads the LOCAL remote-tracking cache**, which
is only as fresh as your last fetch. `git ls-remote` asks the remote. On 2026-08-11 the tree
was **228 commits ahead of the pushed tip**, so what gets indexed can be a day behind what
you are reading.

**Gotcha 2 — a census figure measured without `exclude_patterns` is a proxy.** The action
calls `analysis.AnalyseWithExclude(dir, ["docs/"])`. Measuring with the analyser but without
its arguments gave 1,373 where the truth was 1,204 — a 14% phantom shortfall that reads
exactly like an identity collision. To predict exactly, `git archive` the deployed sha into a
temp dir, drop a throwaway `main` in it, and call the analyser the way production does:

```go
out, _ := analysis.AnalyseWithExclude(root, []string{"docs/"})
```

**This is a scratch harness, not a command to add** — it lives and dies in the scratchpad.
Nothing in `cmd/` should grow to hold it: the whole value is that it runs the *deployed*
revision of `internal/analysis`, so it must be built from a `git archive` of the sha you are
verifying, not from the tree.

**Gotcha 3 — `code_symbols` has `line_start`/`line_end`, NOT `start_line`.** The wrong
spelling fails the whole statement. (`indexed_at` does not exist either; the clocks are
`updated_at` and `commit_time`.)

**Gotcha 4 — the analyser's output count is NOT the row count.**
`uq_code_symbols_identity` is `(repo, path, symbol)` and **kind is not in it**. Methods are
stored `(Recv).Name` so they are immune; plain funcs are not, and two `init()` in one file
cost exactly one row each. Reconcile before calling a difference a bug.

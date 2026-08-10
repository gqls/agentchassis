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

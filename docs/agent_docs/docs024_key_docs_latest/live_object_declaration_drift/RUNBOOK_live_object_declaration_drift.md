# RUNBOOK — live-object declaration drift

Every command that was hard to get right, with its gotcha attached. Change it HERE.

---

## Read the live sweep predicate (the artefact, not the file)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT name, enabled, (regexp_match(pre_query, 'NOT IN \(([^)]*)\)'))[1] AS exclusion_list
 FROM scheduled_tasks WHERE name='claimed-item-timeout';"
```

⚠ **Reading the column does NOT advance the task's rotation** — only *executing* the
`pre_query` does. So no `BEGIN/ROLLBACK` wrapper is needed, and adding one is not free
caution: it is noise that suggests the read is dangerous when it is not.

⚠ The `regexp_match` takes the **first** `NOT IN (...)` in the text. On the live column that is
the real clause; **in a repo migration file it may be a comment** (this is the whole bug).

## Census the class

```bash
# scheduled_tasks that carry a live predicate at all
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT count(*) AS total, count(*) FILTER (WHERE enabled) AS enabled,
        count(*) FILTER (WHERE pre_query IS NOT NULL AND pre_query<>'') AS with_pre_query
 FROM scheduled_tasks;"

# ...and which of those embed a literal vocabulary that a Go list could drift from
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c \
"SELECT name, enabled, (pre_query ~* 'IN \s*\(\s*''') AS has_literal_list
 FROM scheduled_tasks WHERE pre_query IS NOT NULL AND pre_query<>'' ORDER BY enabled DESC, name;"
```

## Find every guard in the class

```bash
grep -rln "sql_for_agents" --include=*_test.go .
```

Then, per file, find the read: `os.ReadFile`, `filepath.Join`, `filepath.Glob`.
**Classify each** — a test asserting a convention *about the migration corpus* is fine
(the repo IS its subject); a test asserting what a *live object does* is the defect.

## Enumerate the verifier registry — do NOT grep one call spelling

```bash
# WRONG — misses RegisterVerifierWithPolicy and undercounts by 2
grep -rn 'RegisterVerifier(' platform/orchestration/actions/discovery_checks/

# RIGHT — every registration entry point in one command
grep -rn 'RegisterVerifier\|RegisteredVerifierItemTypes' \
  platform/orchestration/actions/discovery_checks/*.go | grep -v _test.go
```

⚠ This cost me a false finding on 2026-08-22 (NOTES misstep 1). The registry has **two**
public writers, `RegisterVerifier` and `RegisterVerifierWithPolicy`; both land in the same map.
**Read the function the guard itself calls** (`RegisteredVerifierItemTypes()`), which names
every writer at once.

## Run a guard against COMMITTED HEAD, not the shared dirty tree

```bash
SC=<your scratchpad>
rm -rf "$SC/head" && mkdir -p "$SC/head"
git archive HEAD | tar -x -C "$SC/head"
cd "$SC/head" && go test ./platform/orchestration/actions/ \
  -run TestClaimTimeoutExclusionCoversBothCompletionGates -count=1
```

⚠ The shared tree carries ~10 lanes' WIP and frequently will not compile. A red run there tells
you nothing about your change. Record the sha you archived — `git rev-parse --short HEAD` —
because HEAD moves under you (it moved from `0aa8fa611` to `45b728b01` inside this session).

## Confirm the absences (both are load-bearing for the bug)

```bash
# does anything read a live DB object's definition?  Expect ZERO hits.
grep -rn "pg_get_functiondef\|pg_trigger\|information_schema.triggers" \
  --include=*.go --include=*.py --include=*.sh . | grep -v sql_for_agents

# what does schema_migrations actually record?  A checksum of the FILE.
grep -n "checksum" scripts/migration/run-migrations.sh
```

## Ownership and duplication checks, before routing any work

```bash
./scripts/who-owns.py <number|slug>          # advisory, ~0.3s, reads COMMITS only
grep -rn "<mechanism>" bugs_open/ bugs_closed/
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
"SELECT summary, status FROM site_work_items WHERE item_type='needs_diagnosis' AND status='awaiting_diagnosis';"
```

⚠ `who-owns.py` reads commits, so **a session mid-fix is invisible** — check the tree too.

## Read a council verdict — key on YOUR correlation

```sql
SELECT body FROM diagnosis_artifacts
WHERE correlation_id = '<SUBMISSION_CORR>' AND kind = 'council_report';
```

⚠ Never the runbook's `doc_notes ... ORDER BY created_at DESC LIMIT 1` — that returns whoever
finished last, fleet-wide, and several councils run per hour.

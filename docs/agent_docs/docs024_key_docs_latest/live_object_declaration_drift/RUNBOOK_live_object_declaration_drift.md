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

## This lane's live correlations

| what | correlation | read it with |
|---|---|---|
| diagnosis loop (090) intake | `c236fbb4-ca7f-4540-8170-8b806f40fc54` | — (the intake key; artifacts are NOT under it) |
| diagnosis loop RUN | `c8ec6478-5a54-4a16-aaf1-1e3373684ba0` | `SELECT kind, length(body) FROM diagnosis_artifacts WHERE correlation_id='…';` |
| council gate submission | `b3676918-9eee-4b9f-85f3-749e16b3d033` | see below |

```sql
-- the council verdict, keyed on OUR correlation
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='b3676918-9eee-4b9f-85f3-749e16b3d033' AND kind='council_report'
ORDER BY created_at;
```

⚠ **Never** the runbook's `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
created_at DESC LIMIT 1` — it returns whoever finished last, fleet-wide, and several councils run
per hour. A session hit this on 2026-08-22 and read another lane's REVISE as its own verdict.

⚠ The trigger prints the SUBMISSION correlation; the 090 trigger prints an INTAKE correlation and
then the RUN one. **The run correlation is the key the diagnosis artifacts are written under** —
the intake one resolves nothing.

## Council submission schema (the five client-side traps, all hit or avoided on 2026-08-22)

`.plan` is an **object**, not an array: `{summary, edits[<=8], grounded_in[], risks}`.

- `operation` ∈ `modify|add|remove|config_change`. **`create` is NOT valid** — a new file is `add`.
- `grounded_in` must be an array of **strings**, not objects.
- `risks` must be a **string**, not an array — join into one prose block.
- A sketch whose every non-blank line is a comment (`--`, `#`, `//`) is **refused server-side**
  ("a fix plan proposes changes, not observations").
- **One edit = one file.** `.file` must be a single repo-relative path: no whitespace, no leading
  `/`, no `..`. Naming two files in one edit reads fine to a human and is refused.

**Always `DRY_RUN=1` first** — it runs every client-side validation and the scope admission check,
mints no correlation and spends no credits. A server-side refusal instead costs a dispatch and
~30 minutes, and surfaces only as `current_step='complete_invalid'` with no verdict row, which
reads exactly like "still queued".

## Run the phase-2 auditor against the LIVE database from your machine

The CronJob is the production path, but you can run the same binary locally — this
is how phase 2 was proven, and how you re-prove it after changing a declaration.

```bash
PW=$(kubectl -n ai-persona-system get secret personae-platform-secrets \
       -o jsonpath='{.data.CLIENTS_DB_PASSWORD}' | base64 -d)
kubectl -n ai-persona-system port-forward svc/postgres-clients 15432:5432 >/dev/null 2>&1 &
sleep 4
go build -o /tmp/cka ./cmd/config-key-audit/
PG_CLIENTS_HOST=127.0.0.1 PG_CLIENTS_PORT=15432 CLIENTS_DB_PASSWORD="$PW" \
  /tmp/cka --live-declaration-drift ; echo "exit=$?"
```

⚠ **Omit `--report` when running by hand.** With it, the run writes a `doc_notes`
row that looks exactly like the CronJob's, and the row is the fleet's evidence that
the scheduled job ran. A hand-run row is a false positive for "the job is alive".

⚠ **`PG_CLIENTS_PORT` matters** — `dbConn()` defaults to 5432, so a port-forward on
any other port silently connects nowhere useful (or to something else).

### THE DEMAND CONTROL — never trust a clean run on its own

A clean run is the expected result whether the auditor works or not. Induce drift on
the DECLARATION side (never on production), and require it to fire:

```bash
sed -i 's/^\t"dark_section_audit",$//' platform/livespec/livespec.go   # D1
go build -o /tmp/cka ./cmd/config-key-audit/ && <run as above>          # expect exit 1
git checkout platform/livespec/livespec.go
```

Proven 2026-08-22, all four, both outcomes of one instrument in one session:

| | induced | required | got |
|---|---|---|---|
| D1 | drop a type from the declaration | exit 1 naming object + fragment | ✅ |
| D2 | declare 2 trigger bindings against a live 3 | exit 1 `live count is 3, declared 2` | ✅ |
| D3 | point a probe at a nonexistent task | **exit 2, NOT clean** | ✅ |
| D4 | unset `PG_CLIENTS_HOST` | **exit 2, NOT clean** | ✅ |

### Confirm the probes are readable as `clients_user` before adding a declaration

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -t \
  -c "<your ProbeSQL>"
```
Measured 2026-08-22: `pre_query` 6923 chars, `pg_get_functiondef` 819 chars, the
`pg_trigger` count 3 — all readable, no grants needed.

## Verify at HEAD when the shared tree is broken

The tree carries ~10 lanes' WIP and frequently will not compile — on 2026-08-22 it
was `component_instance_conversion.go` (undefined symbols) from another lane, which
broke `go build ./...` while **committed HEAD built clean**. To test YOUR change:

```bash
SC=<scratchpad>
rm -rf "$SC/mine" && mkdir -p "$SC/mine" && git archive HEAD | tar -x -C "$SC/mine"
cp <only your changed files> "$SC/mine/<same paths>"
(cd "$SC/mine" && go build ./... && go test ./cmd/config-key-audit/ ./platform/livespec/ -count=1)
```
⚠ Check the scratchpad is on disk, not the shared 16G `/tmp` tmpfs: `df -h <path>`.

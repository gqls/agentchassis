# RUNBOOK — bugfix_428_planner_deferral

## Checking whether a deferred verdict is RFC_056's kind (before proposing to dispatch it)

```sql
SELECT id, created_by, handler_agent,
       spec->>'filing_mode' AS filing_mode,
       spec->>'routed_handler' AS routed_handler,
       spec->>'routed_status' AS routed_status
FROM site_work_items WHERE id = '<item_id>';
```
`filing_mode='record'` + `handler_agent=''` is the tell. Read
`write_audit_findings_action.go`'s "WHY IT EXISTS" comment in full before touching
anything with this shape — do not infer intent from the row alone.

## Finding the real population, not the keyword-matched one

```sql
-- The narrow, hand-verifiable shape (what this bug actually means by "13 rows"):
SELECT count(*) FROM site_work_items
WHERE item_type IN ('needs_content_page','needs_content_planning') AND status='deferred'
  AND summary LIKE '[verdict, not dispatched]%'
  AND (summary ILIKE '%entity-directory%' OR summary ILIKE '%entity-page%');
-- 16 today; hand-check each — 2 of the 3 above "13" are false positives for a
-- DIFFERENT finding (catalogue depth; imagery on an already-built template).

-- The trap: the bare item_type/status pair is two orders of magnitude bigger and
-- must never be treated as the dispatch target.
SELECT count(*) FROM site_work_items
WHERE item_type IN ('needs_content_page','needs_content_planning') AND status='deferred';
-- 1,284 as of 2026-09-02.
```

## Pulling a live agent's prompt_template without dumping the whole config

```sql
SELECT default_config->'workflow'->'steps'->'<step_name>'->'config'->>'prompt_template'
FROM agent_definitions WHERE type='<agent_type>' AND is_active=true AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot=false);
```
Redirect to a file and grep/Read it — a 27KB+ prompt dumped straight to the
terminal is how a prior session flooded it (noted in bug 428 itself).

## Editing a live agent prompt safely (the pattern, from migration 640 and this bug's 687)

1. `SELECT snapshot_agent('<agent_type>', '<migration file>: pre-update');` — takes
   the point-in-time backup FIRST, outside any transaction (its own function does
   this atomically). Returns the SOURCE row id, not the snapshot's — don't compare
   the live row against it expecting a diff.
2. `CREATE TABLE IF NOT EXISTS agent_definitions_bak_<N> AS SELECT id, type,
   default_config, now() AS backed_up_at FROM agent_definitions WHERE ...` — a
   second, migration-owned backup that the paired `_ROLLBACK.sql` restores from.
3. Inside `BEGIN;`/`COMMIT;`: a pre-flight `DO $$` block that (a) confirms exactly
   one active row exists (`SELECT INTO` silently takes the first of N), (b) checks
   an "already applied" marker and aborts if present, (c) counts occurrences of
   each anchor substring in the live text and aborts unless every count is exactly
   1 (drift detection — never guess which occurrence to touch).
4. `UPDATE ... SET default_config = jsonb_set(default_config, '{workflow,steps,
   <step>,config,prompt_template}', to_jsonb(replace(replace(<live text>, anchor1,
   replacement1), anchor2, replacement2))) WHERE ...` — same predicate as the
   pre-flight check.
5. A post-write `DO $$` block confirming the new text is present and the old text
   is gone, or `RAISE EXCEPTION`.
6. Dry-run BEFORE running for real: copy the file, `s/^COMMIT;/ROLLBACK;/`, pipe
   through psql. A clean `UPDATE 1` with both DO blocks silent (no RAISE) means the
   real run will succeed identically.
7. After the real run, verify AT THE ARTEFACT — re-pull the prompt text (query
   above) and grep for the new strings. Do not trust the migration's own `COMMIT`
   as proof.

## Checking whether `toJSON` (or any template func) is already live before using it in a prompt edit

```bash
grep -n "\"toJSON\":" platform/orchestration/datahelpers/data_helpers.go
```
If it's in the committed, current `funcMap` already, using it in a prompt needs no
Go change and no image roll — the migration can apply immediately, not as a
`_HOLD`.

## Testing a new Gin admin handler without a live server

Pattern used for `HandleReleaseRecordVerdict` (see
`internal/core-manager/admin/release_record_verdict_test.go`):
```go
db, mock, _ := sqlmock.New()
h := NewSiteAdminHandlers(db, zap.NewNop())
gin.SetMode(gin.TestMode)
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)
c.Request = httptest.NewRequest(http.MethodPost, "/work-items/"+id.String()+"/release", bytes.NewReader(body))
c.Request.Header.Set("Content-Type", "application/json")
c.Params = gin.Params{{Key: "item_id", Value: id.String()}}
h.HandleReleaseRecordVerdict(c)
```
For a query-modifying guard, pin EACH WHERE-clause predicate as its own
`mock.ExpectQuery(<regex for one predicate>)` subtest — a single happy-path test
with a loose matcher (e.g. `ExpectQuery("UPDATE site_work_items")`) does not prove
any individual guard clause is load-bearing; mutate the handler (comment out one
predicate) and confirm the corresponding subtest fails before trusting the suite.

## Frontend changes with no node/npm available

This environment has neither — `go build`/`go test` cover the backend, but a
`.tsx` change can only be verified by careful manual read: check every new
`{...}` brace closes, cross-check any new field access (e.g. `selectedItem.spec?.x`)
against how the backend actually serializes that field (grep the Go handler for
`json.Unmarshal`/`json.Marshal` on the same key to confirm the shape), and diff
against an existing sibling pattern in the same file rather than inventing a new
one.

## Council submissions and verdicts from this bug

- `3f9cdfea-7287-4ab3-afad-9c386fbb7365` — migration 687 (prompt formatting +
  omission-reason requirement).
- `38be9226-d5b5-48b7-9b87-20efbaf3dec3` — the release-surface backend + frontend.

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;
```

## Reading the reconciliation's output once a chassis carrying it has rolled

```sql
-- Did it run at all, and what did it find?  The RECONCILED row is the DEMAND
-- CONTROL: it is written on clean runs too, so a zero has only one reading.
SELECT error_code, count(*), max(created_at)
FROM agent_error_log WHERE error_code LIKE 'RECOMMENDED_TYPE_%' GROUP BY 1 ORDER BY 1;
```
⚠ **No rows at ALL** = the check has not run (not rolled, or the kill switch is set).
**`RECOMMENDED_TYPE_RECONCILED` present with no omission codes** = plans are clean.
Do not read a zero as either without looking at both.

⚠⚠ **The `_DROPPED_IN_VALIDATION` bucket has a deadline on it.** The `463` lane is fixing its
producer (Pass C). Once that rolls, this bucket should go to zero for `blog-post` under a section
prefix — which is the fix working and looks identical to this detector having stopped. **Baseline
before the next roll.**

```sql
-- The parked verdicts, per site and type.
SELECT s.domain, wi.item_key, wi.spec->>'omission_class', wi.spec->>'builder_needed',
       wi.spec->>'producer_verdict', wi.created_at
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE wi.item_key LIKE 'recommended_type_gap:%' ORDER BY wi.created_at DESC;
```

## Asking whether a producer is running, WITHOUT overclaiming

```sql
-- Always select the tracking window in the SAME statement: agent_run_stats is
-- FORWARD-ONLY (it began 2026-08-02), so "no row" means "not since we began",
-- never "never ran".
SELECT s.run_count, s.last_ran_at, t.tracking_since
FROM (SELECT min(first_ran_at) AS tracking_since FROM agent_run_stats) t
LEFT JOIN agent_run_stats s ON s.agent_type = 'blog-content-planner';
```
`tracking_since` non-null + `run_count` null = not since tracking began.
Both null = the table is empty, a different and worse problem.
For all-history, `llm_call_log` reaches to 2026-03-25 — **but `cleanup_old_llm_logs()` is live in
`pg_proc`, so that read can silently become a windowed one.** Say which instrument you used.

## Rehearsing a prompt-row migration apply AND its rollback in ONE doomed transaction

Proves the round trip is byte-exact without ever committing. This is the recipe that produced
`369ec3eb… → 581f9766… → 369ec3eb…` for migration 748:

```bash
{ echo "BEGIN;"
  echo "SELECT md5(default_config #>> '{workflow,steps,plan_site,config,prompt_template}') FROM agent_definitions WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
  sed -e 's/^BEGIN;$//' -e 's/^COMMIT;$//' NNN_x.sql | sed 's/PERFORM snapshot_agent(/PERFORM 1; -- snapshot_agent(/'
  echo "SELECT md5(...) ;"          # applied
  sed -e 's/^BEGIN;$//' -e 's/^COMMIT;$//' NNN_x_ROLLBACK.sql
  echo "SELECT md5(...) ;"          # must equal the first
  echo "ROLLBACK;"
} | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -X -v ON_ERROR_STOP=1 -f -
```
⚠ Neuter `snapshot_agent()` in the rehearsal — it writes outside your transaction's intent and you
do not want a snapshot per dry run.
⚠ **A clean rehearsal is not a proven guard.** INDUCE each one: run the apply body TWICE (the second
must raise "already applied"), and damage a neighbour's anchor deliberately (`replace(... ,'NO LATER
EDITORIAL PASS RUNS','XX GONE XX')`) before the apply — it must raise "730/731 damaged". Then re-read
the live md5 to confirm nothing escaped.

## Verifying a frontend change with no node/npm

There is no `node` or `npm` in this environment, but there IS `docker` — so a `.tsx` change can be
compiled and inspected rather than eyeballed:

```bash
make build-dashboard IMAGE_TAG=v0.0.0-<something-throwaway>     # local only, never pushed
docker run --rm --entrypoint sh aqls/admin-dashboard:v0.0.0-<throwaway> -c \
  "grep -c 'my new string' /usr/share/nginx/html/assets/*.js; \
   grep -c 'a string that must ALREADY be there' /usr/share/nginx/html/assets/*.js; \
   grep -c 'zzz-must-be-absent' /usr/share/nginx/html/assets/*.js"
docker rmi aqls/admin-dashboard:v0.0.0-<throwaway>
```
Run all three greps in the same command: the must-be-present control proves the grep works at all,
and the must-be-absent control proves it can return zero. **Use a throwaway `IMAGE_TAG`** so you
cannot collide with the shared makefile default or another lane's staged bump.

## A mutation test that "fails" may not have run at all

```bash
go build ./... && go test ./pkg/ -run 'X' 2>&1 | grep -E '^--- FAIL' || echo "NO TEST FAILED — not a kill"
```
`go test` prints `FAIL` for a package that does not COMPILE, with no `--- FAIL: Test…` line anywhere.
The more thoroughly a mutation guts the code, the more likely it is to stop compiling — so the
mutations that look most decisive are the likeliest to be vacuous.

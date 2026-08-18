# RUNBOOK — bugfix 277 (commands, with their gotchas attached)

## Census / dry run (read-only, idempotent)

```bash
cd docs/agent_docs/docs024_key_docs_latest/bugfix_277_required_fields_repair
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -f - < CENSUS_2026-08-15_predicted_routes.sql
```
Gotcha: the census re-implements the seed's classifier readably. When either changes, change
BOTH, and re-prove the seed's own string (below) — the census passing proves the census, not
the seed.

## Prove the seed's exact embedded SQL against real rows (before apply / after any edit)

Extract the query FROM THE SEED FILE (never retype it), unescape `''`→`'`, substitute
`$1::uuid`/`$2::uuid` with a real (site_id, work_item_id), run via psql. The 2026-08-15 run of
exactly this proved all five canary candidates route as the census predicts (NOTES).

## Apply the seed (DB config, live instantly, INERT — 0 rows assigned)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/410_required_fields_missing_router.sql
```
Gotcha: the verify block RAISEs (and aborts the COMMIT) on any mis-wired branch, a retired
`spec` key, a non-triaged conversion status, or >0 rows already assigned. A re-run snapshots
first (conditional DO block) and is safe.

## Canary assignment (the first execution — treat any surprise as a stop)

```sql
UPDATE site_work_items
   SET handler_agent = 'required-fields-missing-handler',
       status = 'triaged', attempt_count = 0, updated_at = NOW()
 WHERE item_type = 'required_fields_missing'
   AND status IN ('needs_human_review','unresolved')
   AND COALESCE(handler_agent,'') = ''
   AND left(id::text,8) IN ('332bb3f6','4fa5b019','e512af8a','483fb749');
-- expect UPDATE 4
```
Gotchas: `attempt_count = 0` is load-bearing (claim gate requires attempt_count <
max_attempts). Dispatch cadence is 120s; no dispatch within ~300s of a chassis pod restart.

## Verify the canary (per arm)

```sql
-- each canary row: status + recorded route
SELECT left(id::text,8), status, result->'response'->'triage'->>'route' AS route,
       result->>'route' AS route2, left(COALESCE(error,''),60) AS err
FROM site_work_items WHERE left(id::text,8) IN ('332bb3f6','4fa5b019','e512af8a','483fb749');
-- expected: 332bb3f6 complete/stale · 4fa5b019 complete/converted (+ a new content_rewrite
-- row, source='required-fields-missing-handler', spec->>'mode'='edit_live') ·
-- e512af8a needs_human_review with the blob error message · 483fb749 needs_human_review
-- with the owned message
-- conversions:
SELECT item_type, status, item_key FROM site_work_items
WHERE source = 'required-fields-missing-handler';
-- blob dedup key still held (exactly 1 non-terminal row on the key):
SELECT count(*) FROM site_work_items
WHERE item_key = (SELECT item_key FROM site_work_items WHERE left(id::text,8)='e512af8a')
  AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
```
> **CORRECTED 2026-08-15, from the canary itself:** for COMPLETED rows the route is NOT on the
> row at all. The loop's mark_complete overwrites `result` with the SPAWN bookkeeping
> (role/topics/agent_id — measured on `332bb3f6`), replacing both the close arm's
> `result_fields` and any saga response. The audit trail for closed rows lives in
> `orchestration_states`:
> ```sql
> SELECT left(orchestration_id::text,8), status, current_step,
>        collected_data->'triage'->>'route' AS route
> FROM orchestration_states
> WHERE workflow_plan->>'start_step'='classify' ORDER BY created_at;
> ```
> For PARKED rows the loop's complete no-ops (guard excludes needs_human_review), so the route
> IS on the row: `result->>'route'` + the message in `error`. The canary verified all three
> executed arms this way: `0177ce18` stale · `61a71bbd` no_content_data · `8dd51e7e`
> no_plan_owned (the gas converter), each COMPLETED at `done` with correct facts.

## Fleet assignment (after the canary verifies)

Same UPDATE without the id filter. Then the after-state:

```sql
SELECT status, COALESCE(result->'response'->'triage'->>'route', result->>'route') AS route, count(*)
FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1,2 ORDER BY 3 DESC;
```

> **SETTLED 2026-08-15 ~14:00Z — the producer change is ALREADY LIVE.** Another lane's roll to
> `v1.0.1302` carried commit `5ad81182b` (stamp `194907d5b…`, `git merge-base --is-ancestor`
> exit 0; literal probe 1 with negative control 0). **Replica coverage settled by the
> uniform-image observation**: all 25 pods running the agent-chassis image (15 Running, 10
> Succeeded job pods) carry the SAME `v1.0.1302` — one probe speaks for all, and this is the
> honest answer to the "-l app=agent-chassis is not every pod running the binary" landmine
> (enumerate by IMAGE, then check tag uniformity, before trusting any single-pod probe).
> The PBP-028 edit_live channel was probed the same way: `grep -ac 'attached current content
> for edit mode' /proc/1/exe` → 1, negative control 0.

## Post-roll (after the producer Go change ships in a chassis image)

```bash
# whose commit is the service running? (startup line scrolls; binary probe is durable)
kubectl -n ai-persona-system exec <chassis-pod> -- \
  sh -c "grep -oa 'buildinfo.GitCommit=[0-9a-f-]*' /proc/1/exe | head -1"
git merge-base --is-ancestor <producer-commit> <stamp>   # exit 0 = shipped
# belt-and-braces literal probe: 'required-fields-missing-handler' enters the BINARY
# only via the Go const (the seed is DB-side), so on EVERY replica:
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -ac 'required-fields-missing-handler' /proc/1/exe"        # expect >=1
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -ac 'zzq-negative-control-not-a-handler' /proc/1/exe"     # expect 0
```
Then re-run the fleet assignment UPDATE once — items the OLD producer filed between
assignment and roll are born parked/unassigned and need sweeping in.

## Churn guard (+7 days)

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='required_fields_missing' AND status='unresolved'
  AND created_at > '<fleet-assignment-time>';
-- expect ~0; anything else means a close arm is churning against the producer
```

## Council

Submission corr `7b0e2833-715f-4a9a-897b-efd913073582`. Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='7b0e2833-715f-4a9a-897b-efd913073582' AND kind='council_report'
ORDER BY created_at;
```
Gotcha: publish→run start measured at 29 min under normal load — a missing orchestration row
is latency, not a dropped dispatch; find the run by payload, never re-trigger on absence.

## Rollback

`docs/agent_docs/sql_for_agents/410_required_fields_missing_router_ROLLBACK.sql` — refuses
while non-terminal rows still route at the handler (un-assign first; header has the UPDATE).
If the producer Go change has shipped, roll that back too or items go 'blocked' at claim
(bugs_closed/077 shape).

## Apply ONE migration when other lanes have pending files (the only safe route here)

`./scripts/migration/run-migrations.sh --apply` takes **every** pending file in
`docs/agent_docs/sql_for_agents/`, and on a tree this many sessions share that is routinely four or
five other lanes' work (2026-08-18: `462_fixer_rerenders_skip_owned_pages`, `467`, `468`, `470`).
There is no `--only <file>` flag. So apply the single file yourself, then register it:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/<NNN_name>.sql

./scripts/migration/run-migrations.sh --record-only <NNN_name>.sql \
  --note 'applied out of band by single-file psql; <what the controls proved>'
```
`--record-only` takes a **bare filename**, not a path, and is mutually exclusive with `--apply`.
Gotchas: the file must carry its own `BEGIN`/`COMMIT` (psql only wraps `-f` in a transaction if the
file says so), and its guard `DO` block must `RAISE` rather than `SELECT` — `ON_ERROR_STOP` ignores a
non-empty result set, so a verify block made of `SELECT`s cannot stop the `COMMIT`.

## Prove an edit to a live `pre_query` still parses, WITHOUT running it

`pre_query` bodies here are `UPDATE`-in-CTE statements, so "run it to see if it parses" mutates rows.

```sql
EXECUTE 'EXPLAIN ' || new_q;   -- inside the migration's DO block
```
**EXPLAIN plans without executing.** It catches the realistic failure — an apostrophe left undoubled
inside prose that is nested in a SQL string literal — and mutates nothing. Pair it with an occurrence
count on the anchor you are replacing (`(length(q)-length(replace(q,a,'')))/length(a) = 1`) so the
edit cannot silently land on text another session has since changed.

## Apply exactly ONE migration — a better way than the section above (2026-08-18)

> **CORRECTION to "there is no `--only <file>` flag", above.** True as written, and the
> hand-apply-then-`--record-only` recipe still works — but it makes recording a **separate human
> act** that is easy to forget, and an applied-but-unrecorded migration reads as pending to the next
> session's dry run. There is a scoped path that records automatically.

`MIGRATIONS_DIR` is the runner's own env override, and it can point anywhere. Give it a directory
holding **only your file** and `--apply` cannot reach another lane's work, because the runner never
sees it:

```bash
S=$(mktemp -d)                        # or the session scratchpad
cp docs/agent_docs/sql_for_agents/480_owned_page_refusal_is_not_a_handler_failure.sql "$S/"

MIGRATIONS_DIR="$S" ./scripts/migration/run-migrations.sh          # dry run: must list exactly 1
MIGRATIONS_DIR="$S" ./scripts/migration/run-migrations.sh --apply  # applies AND records it
```

You get the probe, the apply, and the `schema_migrations` row in one step, with the ledger keyed on
the real filename.

**Gotchas, all of them load-bearing:**

* **The assignment must be on the SAME line as the command.** `MIGRATIONS_DIR=…` on its own line is
  an ordinary shell assignment that the script's `${MIGRATIONS_DIR:-…}` default may not pick up, and
  the run then covers the whole repo directory — this is a `LANDMINES.md` entry in its own right.
* **Copy, do not move.** The file must stay in `docs/agent_docs/sql_for_agents/` for everyone else.
* **Sidecars are excluded anyway.** `_ROLLBACK.sql` / `_VERIFY.sql` match `SIDECAR_RE` and are never
  run by the runner, so copying only the migration is enough.
* **`--record-only` still takes a bare filename**, if you ever do need the hand-apply route.

Measured 2026-08-18: the unscoped dry run listed **15 pending files** from other lanes, two of them
probing *inconclusive* on live drift (`467`, `468`). Scoped, it listed one.

## Exercise a migration THREE ways before applying it

A dry run proves the SQL runs. It does not prove the guards can fire, and a guard that cannot fire
is not a guard. All three inside transactions that roll back:

```bash
# 1. the whole file, COMMIT swapped for ROLLBACK — expect: guard passes, UPDATE 1, verify NOTICE
sed 's/^COMMIT;$/ROLLBACK;/' <file>.sql | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# 2. PRE-SET the state the guard refuses, then run the file — expect: the guard ABORTS
{ echo "BEGIN;"; echo "<UPDATE that sets the key/state>"; sed 's/^COMMIT;$/ROLLBACK;/' <file>.sql; echo "ROLLBACK;"; } \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# 3. PLANT the leak the negative control looks for, then run the file — expect: VERIFY ABORTS
#    (without this, a positive assertion passes identically on an UPDATE with no WHERE clause)
```

Both probes must print `ERROR:` **naming your own message**. If probe 2 or 3 succeeds, the guard is
decorative. Run them before the apply, not after — after, you cannot roll back.

## Tell an ownership REFUSAL from a genuine save FAILURE, post-roll

The Tier 1 change (`480` + `6aee22b00`) makes them different rows. Both controls, in one query — a
result with only the first line is equally consistent with the status write being broken:

```sql
SELECT status, count(*), bool_or(result ? 'owned_page_refusal') AS stamped
FROM site_work_items
WHERE handler_agent = 'page-build-handler'
  AND updated_at > '<the roll>'
  AND error LIKE '%OWNED_PAGE_GUARD%'
GROUP BY 1
UNION ALL
SELECT 'control: real save failures', count(*), bool_or(result ? 'owned_page_refusal')
FROM site_work_items
WHERE handler_agent = 'page-build-handler'
  AND updated_at > '<the roll>'
  AND status = 'failed' AND COALESCE(error,'') NOT LIKE '%OWNED_PAGE_GUARD%';
```
Expected: refusals `wont_fix` with `stamped = t`; the control non-zero, `failed`, `stamped = f`.
**A zero control means no genuine failures happened in the window, not that the split works** — widen
the window rather than reading it as a pass.

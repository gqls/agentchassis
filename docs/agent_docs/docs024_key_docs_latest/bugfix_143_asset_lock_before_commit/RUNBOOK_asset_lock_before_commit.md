# RUNBOOK — asset locks, and the checks that were hard to get right

Every command here had a gotcha attached. The gotcha is the point of the entry.

## Is the bug live? (measure exposure before claiming severity)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT purpose, count(*) total, count(locked_at) locked, count(lock_type) typed, count(lock_expires_at) timed
        FROM assets GROUP BY purpose ORDER BY 2 DESC;"
```

**Gotcha: `count(locked_at)` counts NON-NULLs, so it is the locked count — but
`count(*)` per purpose is not the whole story.** The unique index that makes an
asset_key addressable is *partial*: `idx_assets_site_asset_key_unique … WHERE
asset_key IS NOT NULL AND status = 'active'`. So there can be several rows per
key, only one of them active, and **any of them can hold the lock**. Group by
purpose alone and you will not see that.

Result 2026-07-31: `card` 13 rows, **0 locked** → 143 is latent, not live. Five
locked asset rows exist fleet-wide (logo, favicon, og_card, 2×hero).

## Who actually holds the locks, and what does `locked_by` look like?

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT asset_key, purpose, status, locked_at, lock_type, lock_expires_at
        FROM assets WHERE locked_at IS NOT NULL ORDER BY locked_at DESC;"
```

**Gotcha: `locked_by` is free text and one live value is a 200-character
sentence** citing `bugs_open/131` and a runbook step. Do NOT classify it, and do
not assume it is an identity like `admin` — 4 of the 5 live locks also have
`lock_type` **NULL**, which the conservative rule reads as permanent. Any refusal
message must report it verbatim or it will lie about who to ask.

## PREPARE the SQL against the live schema — `go build` cannot parse it

The one check that would have caught a wrong column name, a bad cast, or a
`DISTINCT ON`/`ORDER BY` mismatch. Do this before trusting a green `go test`,
because sqlmock matches your query with a *regex* and will happily accept SQL
Postgres would reject.

```bash
SID=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -tAc "SELECT site_id FROM assets WHERE asset_key='favicon' AND locked_at IS NOT NULL LIMIT 1")

kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 <<SQL
PREPARE p (uuid, text[]) AS
    SELECT DISTINCT ON (asset_key)
           asset_key, COALESCE(locked_by,'') AS locked_by, COALESCE(lock_type,'') AS lock_type, locked_at
      FROM assets
     WHERE site_id = \$1 AND asset_key = ANY(\$2::text[]) AND NOT (locked_at IS NULL)
     ORDER BY asset_key, locked_at DESC;
EXECUTE p('$SID', '{"favicon","og_card","card_tool_matchmatrix"}');
DEALLOCATE p;
SQL
```

**Gotchas, both of which cost a round trip:**

1. **`EXECUTE` cannot take a subquery parameter** — `ERROR: cannot use subquery in
   EXECUTE parameter`. Resolve the site id into a shell variable first (above),
   do not inline `(SELECT …)`.
2. **In a `<<SQL` heredoc the `$1`/`$2` placeholders must be escaped** (`\$1`) or
   the shell eats them and Postgres sees `PREPARE p (uuid, text[]) AS SELECT …
   WHERE site_id =` with nothing after it. Use `<<'SQL'` if you have no shell
   variable to interpolate — but then you cannot inject `$SID`, which is why the
   escaped form is used here.

Expected: the two real locks come back, the unlocked card key does not. That is
the positive **and** the negative control in one statement.

## Prove the guard, and prove the test could fail

```bash
go test ./platform/orchestration/actions/ -run 'AssetLock|AssetDeriv|GitCommittingProducer|UnguardedRule|CardAsset|BrandHead|LockedAsset' -count=1 -v
go test ./platform/orchestration/actions/ -count=1     # the whole package, for the siblings
```

**Gotcha: `-count=1` is not optional.** Without it a cached PASS is reported for
a package you just edited, which is indistinguishable from a real run.

**Gotcha: a source-walking test run from the repo root resolves `.` to the repo
root, not the package.** `go test` sets the working directory to the package
directory, so `classifyGitAssetProducers(".")` is correct — but if you factor it
into a helper called from elsewhere, pass the path explicitly.

**Gotcha: the shell's cwd persists between `Bash` calls in this harness.** An
earlier `cd platform/orchestration/actions` made `go test ./platform/…` resolve
to `platform/orchestration/actions/platform/orchestration/actions` and fail with
`directory not found`, which reads exactly like a broken package. Use absolute
paths or re-`cd` to the repo root.

## Is it live? (the only answer that counts)

Go changes are inert until an image is rebuilt and rolled. **The tag proves
nothing** — `bugs_open/153`: an image may predate your commit and carries no
provenance.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "provenance write SUPPRESSED by the asset lock"; \
   strings /app/agent-chassis | grep -c "approved assets are never overwritten"'
```

**Gotcha: grep a string your change ADDED, plus a positive control in the same
exec.** The first symbol is new in this fix; the second predates it (it is
StoreAssetAction's wording) and proves the grep and the binary are fine when the
first returns 0. A single grep returning 0 is ambiguous between "not shipped" and
"my grep is wrong".

**Gotcha: `logs deploy/X` reads one pod of N** — and `-l app=agent-chassis` may
select a pod running a different service's workload. Name the pod.

## Council gate

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh submission.json
```

**Gotcha: `plan` is an OBJECT, not an array.** A top-level array of edits is
refused with `ERROR: .plan missing — see the header for the fix_plan schema`,
which does not say what is actually wrong. The schema is
`{rationale, submitter, plan:{summary, edits:[{file,symbol,operation,rationale,sketch}], grounded_in, risks}}`
and `operation` is `modify|add|remove|config_change` — **`create` is not a
valid operation**, use `add`.

Find the run by payload, never by the printed id:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -c "SELECT current_step, status FROM orchestration_states
        WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';"
```

**Gotcha: do not roll the cluster while your own council is in flight** — a
deploy kills the in-flight run and you pay for the round twice. Wait for the
verdict, then build.

## Landmines sync

```bash
./scripts/landmines-sync.py --apply     # --check exits 1 on drift
```

**Gotcha: append to the markdown, never hand-write a `doc_notes` landmine row**
(owner ruling D10 — the file is the system of record). Append with a real append
(`open(p,"a")`), not a rewrite: several threads add entries to this file and a
read-modify-write can silently drop theirs.

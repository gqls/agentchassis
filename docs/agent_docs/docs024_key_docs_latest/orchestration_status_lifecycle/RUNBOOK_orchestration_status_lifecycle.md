# RUNBOOK — orchestration status lifecycle

Commands that were hard to get right, with the gotcha attached. Change them **here**, not in
your scrollback.

## Read the live reaper — never a migration file's intent

Several migrations have rewritten this text. Only the row is true.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT pre_query FROM scheduled_tasks WHERE name='stale-orchestration-reaper';"
```

## Rewriting a `pre_query` safely — the four things that bite

A `pre_query` is **data** to your migration. `LIKE` assertions prove a needle is present and prove
**nothing** about whether the SQL parses — it parses only when the cron next fires, so a typo
commits happily and takes out **every arm** minutes later.

**1. Build the new text from the LIVE value, never by retyping.** Dump it, transform it in Python,
diff it. Anything but a diff of exactly your intended lines means stop.

**2. The trailing-newline trap.** `psql -At` appends its own record newline. `head -c -1` is right
for a psql dump and **wrong** for an exact python-written value — strip one too many and the
payload fuses onto the `$PQ$,` terminator.

**3. Assert STRUCTURE as well as content.** An md5 check alone is not enough: mine once **passed on
a truncated file**, because the extractor's end-pattern never matched so it ran to EOF and hashed
the right bytes for the wrong reason.

```bash
for p in '^BEGIN;$' '^COMMIT;$' '^\$PQ\$,$' 'SET pre_query = \$PQ\$$'; do
  printf '%-24s %s\n' "$p" "$(grep -cE "$p" "$FILE")"     # each must be exactly 1
done
grep -c '0\$PQ\$,' "$FILE"                                 # fused terminator: must be 0
```

**4. Verify the embedded payload with a robust extractor.** An `awk -v` version of this returned an
empty string and reported the md5 of a bare newline **for both payloads** — which looks exactly
like a corrupted file. The file was fine; the check was broken.

```python
o = "SET pre_query = $%s$" % tag; c = "\n$%s$," % tag
i = s.index(o) + len(o); j = s.index(c, i)
hashlib.md5(s[i:j+1].encode()).hexdigest()   # compare against md5 of the value built from the live row
```

## The two guards every `pre_query` migration wants

House idiom, not invented here: three-way md5 pre-flight ← `458_detected_item_promoter_..._ROLLBACK.sql`;
`EXECUTE`-the-stored-SQL ← `210_report_pipeline_scheduled_tasks.sql`.

```sql
-- GUARD 1: three states, so a benign repeat run does not read as someone else's edit
IF live_md5 = '<new>' THEN RAISE EXCEPTION '... already applied';
ELSIF live_md5 = '<old>' THEN RAISE NOTICE '... applying';
ELSE RAISE EXCEPTION '... REFUSED: a THIRD edit landed (live md5 %)', live_md5; END IF;

-- GUARD 2: it must PARSE AND RUN. The sentinel discards the effects, which 210 does not
-- need because its gates are inert SELECTs — the reaper's pre_query MUTATES.
BEGIN EXECUTE q; RAISE EXCEPTION 'PARSE_CHECK_OK';
EXCEPTION WHEN OTHERS THEN
  IF SQLERRM <> 'PARSE_CHECK_OK' THEN RAISE EXCEPTION '... does NOT execute (%)', SQLERRM; END IF;
END;
```

**Prove the guard by mutation before trusting it.** Corrupt a copy deliberately and confirm it is
caught — a guard only ever observed passing cannot be told from one that cannot fire.

**Always dry-run first**, `COMMIT` swapped for `ROLLBACK`, then re-read the live row to confirm the
dry run left nothing behind.

```bash
sed 's/^COMMIT;$/ROLLBACK;/' "$FILE" | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
```

## Applying — never `--apply`

The runner has no single-file mode and `--apply` takes **every** pending file (~17 other threads').

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/<file>.sql
./scripts/migration/run-migrations.sh --record-only <file>.sql --note "..."
```

## Inducing the reaper — both directions, control in the same tick

Plant rows, note `last_triggered_at`, wait for it to advance, then read. **Foreground `sleep` is
blocked** — use a background `until` loop.

```sql
INSERT INTO orchestration_states
 (orchestration_id, correlation_id, client_id, status, current_step, workflow_plan,
  owner_agent_type, last_activity, awaited_requests)
VALUES ('<uuid>','<uuid>','t','<STATUS>','s','{}','scratch', NOW() - INTERVAL '5 hours','{}');
```

The **negative control** (`last_activity = NOW()`) is not optional: without it, a fix that reaps
*everything* passes. It also retroactively licenses the pre-fix reading — if the same mechanism
flips the stale row a tick later, "stayed put" was a decision, not a dead scheduler.

⚠ **Since migration 466 the status must be in `orchestration_status_vocabulary`** or the insert is
refused by the FK. That is the enforcement working; invent a status and you get a clear error.

## Adding a new orchestration status

One INSERT. Forgetting it is a hard write failure at the first attempt — loud by design.

```sql
INSERT INTO orchestration_status_vocabulary (status, is_terminal, is_pausable, written_by, notes)
VALUES ('MY_STATUS', false, false, 'pkg.Func (file.go:NN)', 'why it exists');
```

`is_pausable = true` means "may legitimately wait for ever" — the reaper's invariant reads that
column, so a pausable status needs **no sweep change at all**.

## Verifying the Go half after a roll

Config is live instantly; Go is inert until the next roll.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -ac "<a LONG marker from your change>" /proc/1/exe
kubectl -n ai-persona-system exec $POD -- grep -ac "Found stuck orchestration, taking over" /proc/1/exe  # control
```

Never `strings` (absent from these images). **Short literals never reach rodata** — pick a long
marker. An absence probe with no positive control proves nothing; for a **deletion**, pair the
removed string with a string that must still be present.

## Counting Kafka topics (for the 240 link) — the trap that returns 0

Piping `--list` down `kubectl exec` **truncates silently**, and a full `/tmp` makes the listing
write zero bytes and still exit 0. List in-pod, count **in-pod**, return only the count.

```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- bash -c \
  "df -k /tmp | tail -1 | awk '{print \$4}'; \
   bin/kafka-topics.sh --bootstrap-server localhost:9092 --list > /tmp/c.txt 2>/dev/null; \
   echo total=\$(wc -l < /tmp/c.txt); rm -f /tmp/c.txt"
```

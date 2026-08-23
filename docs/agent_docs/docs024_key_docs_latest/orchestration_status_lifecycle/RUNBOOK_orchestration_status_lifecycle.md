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

> **CORRECTED 2026-08-23 — it effectively DOES have one: scope the directory.** The runner reads
> `MIGRATIONS_DIR`, so copying your one file into a temp dir gives a single-file apply that still
> runs the probe and still records the row itself — no separate `--record-only`, and no chance of
> sweeping another lane's pending file. Used to apply `566` on 2026-08-23:
>
> ```bash
> SCOPED=$(mktemp -d); cp docs/agent_docs/sql_for_agents/<file>.sql "$SCOPED"/
> MIGRATIONS_DIR="$SCOPED" ./scripts/migration/run-migrations.sh          # dry run: must list exactly 1
> MIGRATIONS_DIR="$SCOPED" ./scripts/migration/run-migrations.sh --apply
> ```
>
> **Why prefer this to the `psql -f` recipe below:** the probe executes your file verbatim in a
> doomed transaction first, so a stale md5 guard, a lost arm or a query that no longer parses is
> caught *before* anything is written — and `ok … ran to its own COMMIT` is a full rehearsal of the
> verify block. Hand-applying with `psql -f` skips that and then needs `--record-only`, which is a
> second chance to forget. The recipe below stays correct for a file the runner cannot contain
> (its own `ROLLBACK`/`ABORT`, psql metas, `setval`), which the runner lists as "not probed".

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

> **CORRECTED 2026-08-23 (migration 566): "loud by design" was true for a NON-terminal status
> and FALSE for a terminal one — and it is only true for both as of today.** The FK makes
> *forgetting the INSERT* loud. It says nothing about *adding* a terminal status, which until
> today was silently unreaped: `database-cleanup`'s arm 3 deleted `WHERE status IN
> ('COMPLETED','FAILED')` — a literal — while arm 4 skipped everything `is_terminal`. A new
> terminal status was named by neither arm, so the INSERT succeeded, nothing failed, and its
> rows accumulated for ever. **This was not hypothetical: `CANCELLED` had been in exactly that
> position since migration 466 — 24 rows, oldest 2026-07-19 (35 days), against this table's
> 24-hour retention norm.** Caught by the `bugs_open/354` lane while costing a new terminal
> status, not by anything in this runbook.
>
> Migration **566** makes arm 3 read the vocabulary, so both arms now ask the same question and
> no status is enumerated anywhere. **After 566 the sentence above is finally true as written**
> — one INSERT is genuinely all a terminal status needs. Nothing here needs changing for the
> next one; the point of recording it is that the guarantee is new, and a reader of the pre-566
> archive should not assume it held earlier.
>
> **⚠ AND THE PRICE OF THAT GUARANTEE, ADDED THE SAME DAY: `is_terminal` is now a DELETION
> predicate.** Arm 3 deletes `status IN (… WHERE is_terminal)`, so marking a status terminal
> also decides its retention — those rows go 24 hours after `updated_at`, fleet-wide, on the
> next hourly sweep. The INSERT above is still one INSERT and still says nothing about this.
> Two things follow for anyone adding a status here:
> 1. **`is_pausable = true` is what spares a row**, and the note below is right that a pausable
>    status needs no sweep change — but that is now because arm 4 checks the column, not because
>    the sweep is indifferent. A status that is neither terminal nor pausable is reaped by arm 4;
>    one that is pausable and not terminal is immortal **by design**.
> 2. **Never set both flags on one row.** Arm 3 would delete it while arm 4 deliberately spares
>    it — the two arms disagree and the destructive one wins. **There is no CHECK constraint
>    stopping you** (verified 2026-08-23). Zero rows are both today; run the two queries in
>    `LANDMINES.md` ("Setting `is_terminal` … now ARMS a 24-hour DELETE") before and after
>    touching this table, with the row-count demand control, because both queries return empty
>    when healthy and an empty table reads the same way.
>
> Raised as a medium advisory objection by the council `guardian` seat on 566's review
> (`9d23ccd9-c16c-422d-8bf9-7b60e8b52795`, APPROVED). Recorded here rather than fixed with a
> constraint: that is a seam change on a shared table and belongs in its own reviewed migration.
> Flagged to this lane because the column gained a consequence it was not asked about.

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

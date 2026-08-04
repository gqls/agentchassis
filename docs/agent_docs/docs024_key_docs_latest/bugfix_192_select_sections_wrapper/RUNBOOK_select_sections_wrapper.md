# RUNBOOK — `bugs_open/192`, the section_plan wrapper

Every command that was hard to get right, with its gotcha attached. Fix them **here**,
not in your scrollback.

---

## Is a value missing, or has its SHAPE changed?

**This is the one that solved the bug.** Do not read the path you expect — a JSON
traversal answers "absent" and "you are one level too high" with the same quiet NULL,
so the obvious query can only confirm what you already believed. **Enumerate the keys,
and put a healthy row beside a failing one.**

```sql
SELECT left(orchestration_id::text,8) AS orch, status, current_step,
       (SELECT string_agg(k, ',' ORDER BY k)
          FROM jsonb_object_keys(collected_data->'input_data'->'section_plan') k) AS section_plan_keys,
       jsonb_array_length(coalesce(collected_data->'input_data'->'section_plan'->'sections_ready','[]'::jsonb)) AS direct_len,
       jsonb_array_length(coalesce(collected_data->'input_data'->'section_plan'->'section_plan'->'sections_ready','[]'::jsonb)) AS nested_len
FROM orchestration_states
WHERE owner_agent_type='page-content-writer'
ORDER BY updated_at DESC LIMIT 14;
```

Two different key sets in that column **is** the answer. `direct_len 0 / nested_len 7`
says the value moved, not that it vanished.

> **Gotcha: `orchestration_states` has no `id` and no `agent_type` column.** It is
> `orchestration_id` and `owner_agent_type`. `\d orchestration_states` first — schema
> before SQL, and this cost a round trip.

## Which failure mode is this, really?

**Never count an agent's failures without splitting on the step.** `192`'s filing
attributed an overnight spike to this bug; that spike is a *different* step, reachable
only **after** this one succeeds. One `CASE` separates them:

```sql
SELECT date_trunc('hour', updated_at) AS hr,
       CASE WHEN current_step='process_sections_loop'        THEN 'select_sections_miss'
            WHEN current_step LIKE 'process_sections_loop_iter%' THEN 'iter_generate_content'
            ELSE current_step END AS mode,
       status, count(*)
FROM orchestration_states WHERE owner_agent_type='page-content-writer'
GROUP BY 1,2,3 ORDER BY 1 DESC, 4 DESC;
```

> **Gotcha: terminal rows are reaped at roughly 24h**, so "it never happened before
> yesterday" may only mean the window is 24h long. Say which window you measured.

## When did the config actually change?

The onset of a config-caused failure is `agent_definitions.updated_at`, not your guess:

```sql
SELECT type, updated_at FROM agent_definitions
WHERE type IN ('page-build-handler','page-content-writer')
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

> **Gotcha: compare a run's `created_at` against the seed time, NEVER `updated_at`.**
> A failed run's `updated_at` is when it *died*. One run here failed **two seconds
> after** my seed committed and looked like a counter-example; it had started two
> seconds *before* it and had already loaded the old config.

## Where does a step's output actually go?

```sql
SELECT ad.type, s.key AS step, s.value->>'action' AS action, s.value->>'output_field' AS output_field
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = '<the action>';
```

If `output_field` names a key an earlier step wrote, the return value **replaces** it
wholesale (`coordinator.go`, `storeActionResult`). Read the action's returns before
assuming it annotates.

## Re-dispatching a failed work item (how to verify a build fix)

The dispatcher claims `status IN ('triaged','approved')` — a `failed` row is **never**
retried on its own, so nothing happens until you put it back:

```sql
UPDATE site_work_items SET status='triaged', updated_at=now()
WHERE id='<uuid>' AND status='failed'
RETURNING left(id::text,8), status;
```

Then watch the child, not the item — the item goes `claimed` long before the truth is in:

```sql
SELECT left(orchestration_id::text,8) AS orch, status, current_step,
       collected_data->'sections_for_render' ? 'sections_ready' AS has_sections, created_at
FROM orchestration_states WHERE owner_agent_type='page-content-writer'
ORDER BY updated_at DESC LIMIT 5;
```

> **Gotchas.** Check `sites.locked_at IS NULL` and `approval_mode='auto'` first, or it
> will sit there. Re-dispatch only items your lane owns or that a bug file has
> explicitly nominated. And no dispatch within ~300s of a chassis pod restart — the
> spawn is silently dropped.

## Applying a seed

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/<file>.sql
```

> **Gotcha 1 — verification must be `DO`/`RAISE`, never bare `SELECT`s.**
> `ON_ERROR_STOP` ignores a non-empty result set, so a `SELECT` below your `UPDATE`
> cannot stop the `COMMIT`. A silent no-op (anchor text drifted, predicate matched
> nothing) commits looking exactly like success.
> **Gotcha 2 — a trailing `ROLLBACK;` protects nothing unless the `BEGIN;` at the head
> is live.** If the first non-comment statement is not `BEGIN;`, every statement has
> already autocommitted.
> **Gotcha 3 — assert the ORDER of a fallback path list, not just its contents.** The
> shim only self-retires because the flat path precedes it.

## Proving a test can fail (mutation)

A passing test proves nothing on its own. Copy first, restore in the **same** command —
the tree is shared and the window must be seconds:

```bash
BAK=$(mktemp) && cp "$F" "$BAK" && test -s "$BAK" || { echo "NO BACKUP - abort"; exit 1; }
# ... re-introduce the defect (separate step) ...
go test ./platform/orchestration/actions/ -run '<Tests>' -vet=off   # must FAIL
cp "$BAK" "$F"
diff "$BAK" "$F" && echo "restore clean"
```

> **CORRECTED 2026-08-04, same day — the version this replaced FAILED OPEN and left a
> shared-tree source file mutated.** It wrote the backup to a fixed `$SCRATCH` path. That
> directory had been cleared, so the `cp` failed, `set -e` did **not** stop the script, the
> mutation applied anyway, and the restore then failed too. Only the trailing `diff` caught
> it. Two rules, both learned the hard way: **(1) assert the backup EXISTS before mutating**
> (`&& test -s "$BAK" ||` above) — a backup step whose failure is survivable is not a backup
> step; **(2) prefer a file whose current state is COMMITTED**, because then
> `git show HEAD:<path> > "$F"` restores it with no external state to lose. Also: never
> nest a heredoc whose terminator appears in the text you are writing. Full account in
> `WRONG_CALLS.md`.

## Running tests / build here

```bash
go build ./platform/...
go test ./platform/orchestration/actions/ -run 'TestX' -vet=off
```

> **Gotcha: `-vet=off` is needed** — `go vet` fails on a pre-existing unreachable-code
> warning in `load_component_library_actions.go:207` that has nothing to do with you.
> **Only `gofmt -w` the files you touched**; the package has many pre-existing
> unformatted files and reformatting them makes you a same-file passenger on another
> session's work.

## Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

> **Gotcha: submit BEFORE you commit.** The `commit-msg` trailer gate rejects
> `Council-Submitted: pending` — a non-UUID resolves to nothing in the `098` join, and
> forward-only forbids fixing it with an amend. The trigger prints the correlation in
> seconds, so the order is: submit → commit with the real id.

## Did the `090` diagnosis run actually produce anything?

Do not cite a run id as corroboration without checking it returned a verdict:

```sql
SELECT kind, count(*), max(created_at) FROM diagnosis_artifacts
WHERE correlation_id='<RUN_CORRELATION_ID>' GROUP BY 1;
```

> **Gotcha:** `kind='bundle'` rows only, with no `council_report`/verdict and an empty
> `final_result`, means the loop **ended without concluding**. Mine did exactly that.
> That is neither a confirmation nor a refutation, and reporting it as either is the
> dishonesty surface.

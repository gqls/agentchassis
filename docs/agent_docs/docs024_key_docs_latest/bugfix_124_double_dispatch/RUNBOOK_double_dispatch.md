# RUNBOOK — bug 124, double dispatch of `needs_diagnosis`

Every command here had a gotcha attached. Change them HERE when they change.

## Is the loop the live dispatcher?

The single fact the whole bug turns on. If this says `t`, the 090 script must
not publish.

```sql
SELECT name, enabled, interval_seconds, max_concurrent, concurrency_group
FROM scheduled_tasks WHERE name = 'diagnose-pipeline-trigger';
```

Gotcha: the concept register (`docs026_concept_register/register/scheduler-and-tasks.md`)
carries `verify-later: should still be false unless deliberately turned on`. It is
**true** in the live DB. Read the row, not the register.

## The failing shape — one item, two chains

```sql
SELECT owner_agent_type, correlation_id, status, created_at
FROM orchestration_states
WHERE created_at > now() - interval '36 hours' AND owner_agent_type LIKE 'diagnose%'
ORDER BY created_at DESC;
```

Read it as pairs of chains:

- a chain **with** a `diagnose-dispatch-loop` row = the loop's dispatch;
- a chain **without** one, whose correlation equals the item's
  `spec->>'correlation_id'` = the 090 script's direct publish.

Both present for one item ⇒ the bug is firing. Gotcha: **do not run this in an
idle window** — no diagnoses running reports zero, which looks like "fixed"
(`029`'s note).

## Join items to their chains

```sql
SELECT w.item_key, w.created_by, w.status, w.created_at,
       w.spec->>'correlation_id'          AS intake_corr,
       w.spec->>'dispatch_correlation_id' AS run_corr,
       (SELECT count(*) FROM orchestration_states o
         WHERE o.correlation_id::text = w.spec->>'correlation_id') AS orch_on_intake_corr
FROM site_work_items w
WHERE w.item_type = 'needs_diagnosis'
ORDER BY w.created_at DESC;
```

- `orch_on_intake_corr = 2` (orchestrator + agent, no dispatch-loop row) is the
  direct publish. **`= 0` is not proof of anything** on rows older than a few
  days: `orchestration_states` is on a retention clock, so an old item shows 0
  whether or not it ever ran. Record a *rate over a window*, never a count over
  all history.
- `run_corr` is written by this fix (P3). Before the fix it is NULL everywhere.

## Resolve an orchestration id cited in a note

The column is `orchestration_id`, **not `id`** — `SELECT id …` fails with
`column "id" does not exist`.

```sql
SELECT orchestration_id, correlation_id, owner_agent_type, status, created_at, updated_at
FROM orchestration_states WHERE orchestration_id::text LIKE '41d64b75%';
```

This one query is what corrected both `029` §6 and `124` — an orchestration
described as a "re-dispatch 43 minutes later" was created 91 seconds after
intake. **Before repeating a claim about *when* something ran, select its
`created_at`.**

## Live dispatch-loop config (never the seed file)

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'claim_item')
FROM agent_definitions
WHERE type = 'diagnose-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
```

Gotcha: `docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql`
is the *seed*, i.e. what the agent was set up as on 2026-07-09. The live row has
drifted. A plan built from the seed ships inert.

## Does the loop close its items?

```sql
SELECT status, claimed_by, count(*) FROM site_work_items
WHERE item_type='needs_diagnosis' GROUP BY 1,2 ORDER BY 3 DESC;
```

It does — `complete` / `failed` with `claimed_by='diagnose-dispatch-loop'`. The
bug file's `[VERIFIED]` claim that nothing closes them was inferred from a print
statement in the shell script. **A comment is not a config row.**

## Apply the migration

```bash
./scripts/migration/run-migrations.sh            # dry run — lists ALL pending
MIGRATIONS_DIR=<dir containing only your file> ./scripts/migration/run-migrations.sh --apply
```

Gotcha: `--apply` takes **every** pending file in the directory, not just yours.
Other sessions leave files there. Dry-run once per session and again after any
roll.

## Verify the Go half on the pod, not on git

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "\$ctx\."'
```

`$ctx.` is a string literal the change **introduces**, so a zero count is a real
negative. Pick a marker your change *creates* or *deletes*, never a Go `const`
or a type name — those are vacuous markers that never appear in the binary.

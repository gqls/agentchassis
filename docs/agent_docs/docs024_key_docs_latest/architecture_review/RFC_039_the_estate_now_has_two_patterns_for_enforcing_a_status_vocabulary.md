# RFC 039 — the estate now has TWO patterns for enforcing a status vocabulary, and nothing says which applies where

**Raised 2026-08-19** by the `orchestration_status_lifecycle` lane, **at the owner's direction** and
at the council gate's. Both the `architecture` and `reuse_agent` seats flagged it on migration
`466`'s round (`f0e95e58`):

> `architecture` [low]: *"The asymmetry between this table+FK pattern and the 8+ other inline-CHECK
> status columns is left unresolved estate-wide. Worth a tracked follow-up so a future author doesn't
> have to rediscover which pattern applies where."*
>
> `reuse_agent` [low]: *"This is the second pattern for 'declared vocabulary enforcing a status
> column' now live in the estate … the plan's own risks section names this asymmetry but no
> follow-up."*

**Status: RAISED, needs an owner ruling. Nothing is broken; this is about which shape the next
author copies.** The owner decided on 2026-08-19 to write it up rather than spread the pattern or
leave it undocumented.

## 1. What exists now, in plain terms

A "status vocabulary" is just the set of values a status column is allowed to hold, plus what each
one means. Two mechanisms now enforce that in this codebase.

**Pattern A — inline `CHECK` constraint.** The allowed values are listed in the constraint itself.
Used by at least eight tables: `agent_definitions`, `agent_group_definitions`, `awaited_requests`,
`asset_ingest_staging`, `chassis_intake_events`, `processed_messages`, `thunder_instances`,
`model_lifecycle.training_runs`.

```sql
CHECK (status = ANY (ARRAY['pending'::text, 'running'::text, 'done'::text, 'failed'::text]))
```

**Pattern B — lookup table + foreign key.** The allowed values are rows, carrying *properties*, and
the constraint is a FK to that table. Used by exactly one table, `orchestration_states`, as of
migration `466`.

```sql
CREATE TABLE orchestration_status_vocabulary (
    status character varying(50) PRIMARY KEY, is_terminal boolean NOT NULL,
    is_pausable boolean NOT NULL DEFAULT false, written_by text, notes text, ...);
ALTER TABLE orchestration_states ADD CONSTRAINT fk_orchestration_states_status
  FOREIGN KEY (status) REFERENCES orchestration_status_vocabulary (status);
```

## 2. Why B was chosen for `orchestration_states`, and why that reason does not obviously generalise

The deciding factor was **who needs to read the vocabulary**, not which is tidier.

`orchestration_states`'s consumers include **DB-resident SQL** — the `pre_query` text of two
`scheduled_tasks` rows (`stale-orchestration-reaper`, `database-cleanup`). Those sweeps must ask
*"which statuses are terminal?"* at query time. A `CHECK` constraint cannot answer a question; only a
table can be joined. The Go-side single-source idiom this estate already has —
`sqlInList(workItemTerminalStatuses)` — renders a Go slice into SQL and therefore reaches **only
Go-built queries**, never a `pre_query`.

**None of the other eight columns has that property.** Their statuses are read by Go, which can hold
its own constant list. So the argument that forced pattern B here is genuinely absent elsewhere, and
"be consistent" is not by itself a reason to convert them.

## 3. What the asymmetry actually costs

- **A future author has to guess.** Adding a new status column, they will copy whichever they saw
  last. Nothing in the tree says the choice turns on "is this read by DB-resident SQL?".
- **Two maintenance shapes.** Pattern A needs a migration to `ALTER` the constraint; pattern B needs
  an `INSERT`. Both are cheap; they are just different, and a reviewer must know which is correct.
- **It is not free to converge either.** Converting eight tables means eight FK constraints on live
  write paths, several with writers this lane has not audited. Migration `466`'s own experience is
  the warning: the single biggest risk there was that seeding from `SELECT DISTINCT status` rather
  than from the *code's writers* would have rejected every new row. That audit would have to be
  repeated per table, correctly, eight times.

## 4. Options

1. **Codify the rule and leave both patterns.** State the test explicitly — *if the vocabulary must
   be read by SQL that lives in the database (a `pre_query`, an agent step), use pattern B; otherwise
   pattern A* — and record it where a new-column author will meet it. Cheapest; keeps both shapes but
   makes the choice mechanical rather than accidental.
2. **Converge on pattern B estate-wide.** One shape everywhere, properties available to any consumer.
   Costs eight per-table writer audits and eight FKs on live paths, for tables where nothing is
   currently bleeding.
3. **Treat B as a one-off and say so.** Record `orchestration_states` as an exception justified by
   its DB-resident consumers, with no general rule implied.
4. **Do nothing.** The status quo, which is what the two seats objected to.

## 5. Recommendation

**Option 1.** The distinguishing property is real, testable and one sentence long, and it explains
both the existing eight and the new one without converting anything. Option 2 spends a large,
error-prone budget to remove a difference that is doing useful work; option 3 understates it, because
the next table with a `pre_query` consumer will face exactly the same choice and deserves the
precedent.

## 6. What is NOT being asked

This RFC does not propose changing `orchestration_states` — `466` is applied, live and
council-approved (round 2 in flight). It asks only where the *rule* should live and whether the other
eight columns are in scope.

## 7. Evidence

- The eight+ pattern-A tables, from `pg_constraint` on 2026-08-19 (`contype='c'`, definition matching
  `%status%`).
- Pattern B's consumers, both DB-resident: `SELECT name FROM scheduled_tasks WHERE pre_query ILIKE
  '%orchestration_status_vocabulary%'` returns `stale-orchestration-reaper` and `database-cleanup`.
- The Go-side idiom that cannot reach them: `sqlInList(workItemTerminalStatuses)`, used at
  `record_vision_finding_action.go:191`, `tool_acceptance_actions.go:959,1241`.
- Precedent for declared-rows-over-literals: `reaper_policies` (migration `335`, RFC_018).
- Full context, decisions and the measurement discipline that produced them:
  `docs024_key_docs_latest/orchestration_status_lifecycle/`.

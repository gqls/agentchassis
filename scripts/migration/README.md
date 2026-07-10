# Migrations (clients_db)

Set up 2026-07-10. Numbered SQL files in `docs/agent_docs/sql_for_agents/`
(`NNN_name.sql`), applied in order by `run-migrations.sh`, recorded in the
`schema_migrations` table in clients_db.

```bash
./scripts/migration/run-migrations.sh          # list pending (dry run)
./scripts/migration/run-migrations.sh --apply  # apply + record, in order
```

## Rules

- **Baseline is 124.** Files `001–123` predate the system: applied ad hoc,
  never auto-applied, kept as history. `124_schema_migrations.sql` bootstraps
  the ledger and backfills the travelling-docs arc (125–139).
- **Next number = highest in the directory + 1.** One change per file.
- **Every migration that touches `agent_definitions` opens with**
  `SELECT snapshot_agent('<type>', '<filename>: pre-update');`
- **Guard your own shape**: end with a `DO $$ ... RAISE EXCEPTION ... $$` block
  asserting the exact post-conditions, inside the same `BEGIN/COMMIT` so a
  failed guard rolls the whole file back. Include a rollback recipe in
  comments.
- **Run files, never paste** (`psql -f` semantics — pasting mangles comments
  and dollar-quoted bodies), and remember an aborted psql session needs
  `ROLLBACK;` before anything else runs.
- A failed file stops the run and is not recorded; fix it (or supersede it
  with the next number) rather than editing an already-recorded file —
  checksums in the ledger are of the file as applied.

Related, pre-existing, unchanged: `snapshot_agent()` (before-images of agent
rows) and `migration_backups` (manual before-values of arbitrary rows). The
ledger records *what ran when*; those record *what it replaced*.

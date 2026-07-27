# Register — agent-definition-registry

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions (1 unique block, present twice
in the source cluster file due to mechanical duplication in the input), from
unit U19.

### ADR-001 — Agent definition snapshot/revert via backup table
- **status:** deployed
- **status-evidence:** Migration "Supersedes 030_snapshot_as_column.sql"; motivated by an audit: 8 Go query sites read agent_definitions unfiltered, 2 picked the wrong row when a version+1000 snapshot existed, and patch UPDATEs overwrote snapshots breaking revert.
- **what:** Agent config snapshots move out of agent_definitions into agent_definitions_backup with snapshot_taken_at/snapshot_reason/restored_at; snapshot_agent(type, reason) copies the live row verbatim, revert_agent(type) restores the most recent unrestored snapshot and marks it restored (audit trail preserved, never deleted); agent_snapshots view exposes per-step model/provider of each snapshot. Structurally eliminates the wrong-row class of bugs since no snapshot rows remain in the live table; contaminated legacy snapshots deleted. Patch contract: snapshot before patch, and bulk ad-hoc backups coexist (NULL snapshot_taken_at).
- **sources:** docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql
- **relations:** model upgrade sweeps; migration discipline; is_snapshot column retained pending Go cleanup; Agent variants + snapshot versioning (agent-memory-and-evolution register — an earlier, abandoned, differently-shaped snapshot design that this superseded/replaced in spirit)
- **verify-later:** snapshot_agent/revert_agent functions live; is_snapshot readers at chassis lines referenced

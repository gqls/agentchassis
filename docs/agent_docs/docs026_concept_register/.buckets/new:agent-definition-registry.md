
<!-- SOURCE: U19_sql_tables_components.md -->
### Agent definition snapshot/revert via backup table
- **category:** NEW:agent-definition-registry
- **status-signal:** deployed
- **status-evidence:** Migration "Supersedes 030_snapshot_as_column.sql"; motivated by an audit: 8 Go query sites read agent_definitions unfiltered, 2 picked the wrong row when a version+1000 snapshot existed, and patch UPDATEs overwrote snapshots breaking revert.
- **what:** Agent config snapshots move out of agent_definitions into agent_definitions_backup with snapshot_taken_at/snapshot_reason/restored_at; snapshot_agent(type, reason) copies the live row verbatim, revert_agent(type) restores the most recent unrestored snapshot and marks it restored (audit trail preserved, never deleted); agent_snapshots view exposes per-step model/provider of each snapshot. Structurally eliminates the wrong-row class of bugs since no snapshot rows remain in the live table; contaminated legacy snapshots deleted. Patch contract: snapshot before patch, and bulk ad-hoc backups coexist (NULL snapshot_taken_at).
- **sources:** docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql
- **relations:** model upgrade sweeps; migration discipline; is_snapshot column retained pending Go cleanup.
- **verify-later:** snapshot_agent/revert_agent functions live; is_snapshot readers at chassis lines referenced.

<!-- SOURCE: U19_sql_tables_components.md -->
### Agent definition snapshot/revert via backup table
- **category:** NEW:agent-definition-registry
- **status-signal:** deployed
- **status-evidence:** Migration "Supersedes 030_snapshot_as_column.sql"; motivated by an audit: 8 Go query sites read agent_definitions unfiltered, 2 picked the wrong row when a version+1000 snapshot existed, and patch UPDATEs overwrote snapshots breaking revert.
- **what:** Agent config snapshots move out of agent_definitions into agent_definitions_backup with snapshot_taken_at/snapshot_reason/restored_at; snapshot_agent(type, reason) copies the live row verbatim, revert_agent(type) restores the most recent unrestored snapshot and marks it restored (audit trail preserved, never deleted); agent_snapshots view exposes per-step model/provider of each snapshot. Structurally eliminates the wrong-row class of bugs since no snapshot rows remain in the live table; contaminated legacy snapshots deleted. Patch contract: snapshot before patch, and bulk ad-hoc backups coexist (NULL snapshot_taken_at).
- **sources:** docs/agent_docs/sql_for_tables/045_agent_definitions_backup.sql
- **relations:** model upgrade sweeps; migration discipline; is_snapshot column retained pending Go cleanup.
- **verify-later:** snapshot_agent/revert_agent functions live; is_snapshot readers at chassis lines referenced.

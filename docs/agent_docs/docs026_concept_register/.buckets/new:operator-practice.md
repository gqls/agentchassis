
<!-- SOURCE: U25_leopardess_social.md -->
### In-chassis replicability requirement for operator work
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** REPLICATION_in_chassis.md (2026-07-10) maps every off-platform action to [chassis]/[human]/[gap]; RUNNING_NOTES turn 8: "everything done in this thread must be replicable inside the chassis … Either document the tool-free path or don't use the tool."
- **what:** Standing owner rule: interactive-agent work on sites must be reproducible by the chassis itself (or documented as a human-judgement or platform-gap item). The audit, spec rewrites, artifact verification and imagery deploys all map to normal platform operations; the genuinely human items (choosing a permanent logo, stating personal-history claims, setting engagement terms) are deliberately not automated; named gaps: pinned enforcement, favicon/OG derivation, chart capability.
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md (whole); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8
- **relations:** hitl; documentation-system; site_specs pinned gap
- **verify-later:** n/a (practice doc); checkpoint_for_review as the in-chassis review surface

<!-- SOURCE: U25_leopardess_social.md -->
### Operator discipline: verify-by-artifact, dated backups, kcat generic trigger
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** PLAN standing rule 2: "Verify by artifact, never by report. (This platform has a long history of builds reporting success while building nothing.)"; RUNBOOK landmines 2/17/18; VERDICT §7 "Verified by artifact, never by item status" with md5 predictions.
- **what:** The cross-workstream operating discipline: (1) never trust a `complete` work item — curl the page, read the DB row, diff the bytes (strongest form: predict output bytes offline and md5-compare); (2) back up before ANY change using dated `bak_*`/`_backup_YYYYMMDD` tables and never reuse a name (CREATE TABLE IF NOT EXISTS silently no-ops); (3) trigger any agent by producing to Kafka system.agent.generic.requests via kcat with the standard header set; (4) kubectl exec heredocs need `-i` or silently run nothing — prefer kubectl cp + psql -f; quote heredoc delimiters.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Landmines; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/leopardessconsulting/HANDOFF.md#5
- **relations:** silent no-op success class; fleet generalisation doctrine
- **verify-later:** n/a (doctrine); bak_* table inventory in clients_db

<!-- SOURCE: U25_leopardess_social.md -->
### In-chassis replicability requirement for operator work
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** REPLICATION_in_chassis.md (2026-07-10) maps every off-platform action to [chassis]/[human]/[gap]; RUNNING_NOTES turn 8: "everything done in this thread must be replicable inside the chassis … Either document the tool-free path or don't use the tool."
- **what:** Standing owner rule: interactive-agent work on sites must be reproducible by the chassis itself (or documented as a human-judgement or platform-gap item). The audit, spec rewrites, artifact verification and imagery deploys all map to normal platform operations; the genuinely human items (choosing a permanent logo, stating personal-history claims, setting engagement terms) are deliberately not automated; named gaps: pinned enforcement, favicon/OG derivation, chart capability.
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md (whole); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8
- **relations:** hitl; documentation-system; site_specs pinned gap
- **verify-later:** n/a (practice doc); checkpoint_for_review as the in-chassis review surface

<!-- SOURCE: U25_leopardess_social.md -->
### Operator discipline: verify-by-artifact, dated backups, kcat generic trigger
- **category:** NEW:operator-practice
- **status-signal:** deployed
- **status-evidence:** PLAN standing rule 2: "Verify by artifact, never by report. (This platform has a long history of builds reporting success while building nothing.)"; RUNBOOK landmines 2/17/18; VERDICT §7 "Verified by artifact, never by item status" with md5 predictions.
- **what:** The cross-workstream operating discipline: (1) never trust a `complete` work item — curl the page, read the DB row, diff the bytes (strongest form: predict output bytes offline and md5-compare); (2) back up before ANY change using dated `bak_*`/`_backup_YYYYMMDD` tables and never reuse a name (CREATE TABLE IF NOT EXISTS silently no-ops); (3) trigger any agent by producing to Kafka system.agent.generic.requests via kcat with the standard header set; (4) kubectl exec heredocs need `-i` or silently run nothing — prefer kubectl cp + psql -f; quote heredoc delimiters.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Landmines; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/leopardessconsulting/HANDOFF.md#5
- **relations:** silent no-op success class; fleet generalisation doctrine
- **verify-later:** n/a (doctrine); bak_* table inventory in clients_db

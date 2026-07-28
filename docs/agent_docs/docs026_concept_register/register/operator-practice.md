# Register — operator-practice

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

2 concepts, consolidated from 4 raw extractions across unit U25 (each of the 2 distinct blocks
appeared byte-identically twice within the cluster input file — treated as duplicate copies of
one extraction, not independent corroboration).

### OPP-001 — In-chassis replicability requirement for operator work
- **status:** deployed
- **status-evidence:** REPLICATION_in_chassis.md (2026-07-10) maps every off-platform action to [chassis]/[human]/[gap]; RUNNING_NOTES turn 8: "everything done in this thread must be replicable inside the chassis … Either document the tool-free path or don't use the tool."
- **what:** Standing owner rule: interactive-agent work on sites must be reproducible by the chassis itself (or documented as a human-judgement or platform-gap item). The audit, spec rewrites, artifact verification and imagery deploys all map to normal platform operations; the genuinely human items (choosing a permanent logo, stating personal-history claims, setting engagement terms) are deliberately not automated; named gaps: pinned enforcement, favicon/OG derivation, chart capability.
- **sources:** docs/leopardessconsulting/REPLICATION_in_chassis.md (whole); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-8
- **relations:** hitl; documentation-system; site_specs pinned gap
- **verify-later:** n/a (practice doc); checkpoint_for_review as the in-chassis review surface

### OPP-002 — Operator discipline: verify-by-artifact, dated backups, kcat generic trigger
- **status:** deployed
- **status-evidence:** PLAN standing rule 2: "Verify by artifact, never by report. (This platform has a long history of builds reporting success while building nothing.)"; RUNBOOK landmines 2/17/18; VERDICT §7 "Verified by artifact, never by item status" with md5 predictions.
- **what:** The cross-workstream operating discipline: (1) never trust a complete work item — curl the page, read the DB row, diff the bytes (strongest form: predict output bytes offline and md5-compare); (2) back up before ANY change using dated bak_*/_backup_YYYYMMDD tables and never reuse a name (CREATE TABLE IF NOT EXISTS silently no-ops); (3) trigger any agent by producing to Kafka system.agent.generic.requests via kcat with the standard header set; (4) kubectl exec heredocs need -i or silently run nothing — prefer kubectl cp + psql -f; quote heredoc delimiters.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Landmines; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/leopardessconsulting/HANDOFF.md#5
- **relations:** silent no-op success class; fleet generalisation doctrine; standing evidence rules (operating-doctrine register, OPD-001)
- **verify-later:** n/a (doctrine); bak_* table inventory in clients_db

### OPP-003 — `check_logged_model_output`: a pre-commit detector for publishing model text
- **status:** deployed (advisory; 8th check in `scripts/pattern-check.py`, run by `.githooks/pre-commit`)
- **what:** Flags a log sink (`log.*`, `logger.*`, `fmt.Print*`) that passes an
  unwrapped payload identifier (`text`, `body`, `completion`, `response`, `raw`,
  …) inside any package that calls `GenerateText`, and points at
  `aiservice.Fingerprint` as the fix. Silent when the value is wrapped in a
  derived-fact helper. Exists because whether to log model output is a *content*
  decision, not a code-review one — it took a full council round to surface once,
  and the reviewer stated it could not be closed by the council alone.
- **sources:** `scripts/pattern-check.py` (`check_logged_model_output`)
- **relations:** LCO-005; `bugs_open/083`
- **note — the check nearly shipped VACUOUS, and this is the transferable part:**
  the first version gated on **the file** containing `GenerateText`. The LLM call
  lives in `defend.go` and the log sink in its sibling `ailog.go`, so it examined
  **zero files and printed a clean result**. A clean result and an unrun check are
  byte-identical output. Caught only by a **positive control** — auditing the
  commit that introduced the defect and *requiring* a finding. Now gated on the
  package and verified three ways: fires on `a37a2037c`, silent on the fix, zero
  findings fleet-wide. **Any new detector needs its failing case demonstrated at
  the moment it is written.**

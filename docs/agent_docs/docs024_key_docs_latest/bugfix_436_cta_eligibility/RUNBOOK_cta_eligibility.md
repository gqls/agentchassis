# RUNBOOK — bugfix_436_cta_eligibility

## Council round
Correlation: `9faa2a23-f3bc-464e-8c3a-9d3d44759cc0` (submitted 2026-09-02 ~19:45 BST).
Find the run by PAYLOAD, not the printed id, and budget ~30 min (dispatch queues behind the fleet):
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '9faa2a23-f3bc-464e-8c3a-9d3d44759cc0';
-- verdict artefacts:
SELECT kind, created_at FROM fix_artifacts WHERE correlation_id='9faa2a23-f3bc-464e-8c3a-9d3d44759cc0' AND kind='council_report' ORDER BY created_at;
-- human-readable note:
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```
REVISE → resubmit with `RESUBMIT_CORR=9faa2a23-f3bc-464e-8c3a-9d3d44759cc0`.

## Apply migration 714 (the column — safe under the old binary, needed by the new one)
⚠ `MIGRATIONS_DIR` assignment ON THE SAME LINE as the command, or the run is UNSCOPED and applies
~100 other threads' pending files (LANDMINES). Dry-run first, per session:
```bash
cd docs/agent_docs/sql_for_agents && MIGRATIONS_DIR=. ./run-migrations.sh 2>&1 | grep 714   # dry run: listed as pending?
# apply just this file: use the runner's scoping mechanism, or by hand:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < 714_pages_eligible_as_cta_target.sql
# verify:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA \
  -c "SELECT count(*) FROM information_schema.columns WHERE table_name='pages' AND column_name='eligible_as_cta_target';"   # 1
```

## Apply 715_HOLD (check enablement) — ⛔ ONLY AFTER the carrying image rolls
The discovery runner FAILS the whole step on an unregistered check name. Prove registration at the
BINARY's own capability record, with both controls (council round 2, debug_historian: state HOW the
roll is confirmed — a log grep is a startup line that scrolls, and an image tag can serve a stale
cached binary):
```sql
-- positive control (must be present), the new check, and a negative control (must be absent):
SELECT name, git_commit, last_seen_at FROM service_binary_capabilities
WHERE kind='discovery_check' AND name IN ('misdirected_cta','cta_rank_anomaly','no_such_check_zz');
-- rolled ⇔ misdirected_cta present AND cta_rank_anomaly present AND no_such_check_zz absent.
-- last_seen_at is refreshed, so this probe has no shelf life.
```
Then apply 715 by hand (same psql pattern as 714). It snapshots the agent row first (two-arg
snapshot_agent → agent_definitions_backup); verify the snapshot holds the PRE-change config, per
the query in the file header. Its DO/RAISE guard aborts if the checks array is not at
`workflow.steps.run_checks.config.checks` (path verified against the live row 2026-09-02).
Undo: `715_enable_cta_rank_anomaly_check_ROLLBACK.sql`, by hand only.

## Induced canary (both directions), after roll + 714
```sql
-- pick a canary site's rank-1 tool (the resolver's own ordering, mirrored by the shared SQL):
SELECT name, COALESCE(nav_order,100) FROM pages
WHERE site_id='<site>' AND page_type IN ('tool','game') AND status IN ('active','deployed')
ORDER BY COALESCE(nav_order,100), name LIMIT 3;
UPDATE pages SET eligible_as_cta_target=false WHERE site_id='<site>' AND name='<rank1>';
```
Dispatch a resolve/rebuild for one page; assert at the STORED field (`page_components.content_data`)
that the CTA url is NOT the opted-out page, AND check the header button at the SERVED bytes
(`scripts/probe-page-url.sh`; the header's pick is never persisted — no DB check can see it).
Then flip back to true, re-run, assert it wins again. Never verify by work-item status
(`complete` is not evidence — 391's own rule).

## The tests, locally
```bash
go test ./platform/orchestration/datahelpers/ -run 'TestRank|OptedOut|Ineligible|CarriesEligibility'
go test ./platform/orchestration/actions/discovery_checks/ -run 'TestCTARankAnomaly'
# actions package: NOT locally testable while another session's untracked half-written test file
# breaks package compile — verify at committed HEAD instead:
scripts/verify-head-builds.sh --test
```

# RUNBOOK — bugfix 444 (listing-page item-source gate)

Commands that were hard to get right, with their gotchas. Plan:
`PLAN_2026-09-02_listing_source_gate.md`. Register: BLD-028.

## Verify the bug / re-verify an instance (judge at ITEMS, never bytes/status)

```bash
# glossary/listing served body — count items, not bytes (016b §9, 444)
curl -s https://<domain>/glossary.html | grep -c '<h3'   # meta-prose headings ≠ terms: read them
curl -s https://<domain>/<listing>/index.html | grep -oE '<h[23][^>]*>[^<]{0,70}' | head
```

```sql
-- the three producer-absence checks, per site (all three ran 2026-09-02; see NOTES)
SELECT count(*) FROM content_sources cs JOIN sites s ON s.id=cs.site_id
 WHERE s.domain='<domain>' AND cs.is_active;                          -- feed half
SELECT (data->'content_features') FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE s.domain='<domain>' AND ss.aspect='classification' AND ss.is_current; -- driver half
SELECT st.input_data->>'vertical' FROM scheduled_tasks st
 JOIN sites s ON s.domain = st.input_data->>'domain'
 WHERE s.domain='<domain>' AND st.target_agent_type='directory-json-exporter'; -- bare directory-listing
```

Gotcha: the build-time `plan_sections` Warn lines do NOT survive pod restarts — the
444 diagnosis lost them (pods rolled 2 hours after the build). The durable evidence
is the stored `section_plan` in `orchestration_states.collected_data` and, post-fix,
the `LISTING_PAGE_HELD_NO_ITEM_SOURCE` rows in `agent_error_log`.

## Apply the migration

```bash
# THIS FILE ONLY — never an unscoped runner --apply (the MIGRATIONS_DIR=… on its
# own line trap is in LANDMINES)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/720_planner_listing_source_gate.sql
```

The file refuses (DO/RAISE) unless exactly one active build-site-planner row exists,
the rule-3 anchor appears exactly once, and template balance is unchanged; it
verifies old-text-gone + new-rule-present + flag=true + 433/718 surfaces intact.

## Prove the gate is LIVE after a chassis roll (owed to council round 1)

```bash
# 1. the binary's own provenance. ⚠ The log line is a STARTUP line and SCROLLS —
#    on agent-chassis it is typically out of --tail range within hours (measured
#    2026-08-11; council r3 debug_historian flagged relying on it). An empty grep
#    means "not in range", never "unstamped". Prefer the binary probe, and ALWAYS
#    run a control in the same breath (a sha that must be absent):
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance' \
  || kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<expected-sha>" /proc/1/exe
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<a-sha-that-must-be-absent>" /proc/1/exe && echo "CONTROL FAILED — probe not discriminating"
git merge-base --is-ancestor 6525b45ae <the stamp>   # exit 0 = the gate shipped

# 2. the config half (live immediately at apply)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c \
 "SELECT default_config #>> '{workflow,steps,validate_plan,config,enforce_listing_sources}'
  FROM agent_definitions WHERE type='build-site-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"   # expect: true

# 3. the CAPABILITY, on the next real plan run (a canary brief or replan):
#    a held page = a capability_gap row + a findings row; both must appear together
```

```sql
-- the enablement receipt (also RUNBOOK_remake_release.md §6, portfolio_positioning)
SELECT item_key, spec->>'builder_needed' AS needs, summary
FROM site_work_items WHERE site_id=:site AND item_type='capability_gap'
AND spec->>'gap_kind'='producer_missing' AND status NOT IN ('complete','cancelled','rejected');
-- the gate's own audit trail
SELECT created_at, error_message FROM agent_error_log
WHERE error_code='LISTING_PAGE_HELD_NO_ITEM_SOURCE' ORDER BY created_at DESC LIMIT 10;
```

## Tests

```bash
go test ./platform/orchestration/actions/ -run 'TestResolveListing|TestEnforceListing|TestPlanSection_RequiredQuery|TestPlanSection_OptionalQuery' -count=1
```

Gotcha (sqlmock): the resolver deliberately queries ONCE per source per page — an
expectation left unmatched usually means a fail-open branch swallowed an unexpected-
call error; assert `ExpectationsWereMet()` on the drop tests, and remember fail-open
means a mocking mistake reads as "kept", never as "dropped".

## Council trail

Corr `c0990eb3-9f50-4e08-b578-a7e05f786945` (round 1 REVISE by bug_historian's
gating objection; round 2 submitted 2026-09-02 late evening with measurements).
Find the run by payload, not printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'c0990eb3-9f50-4e08-b578-a7e05f786945';
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='c0990eb3-9f50-4e08-b578-a7e05f786945' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```

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
# 1. the binary. ⚠ BOTH instruments this section used to prescribe are LANDMINED
#    (corrected 2026-09-02 after running them): the 'build provenance' log line
#    DOES NOT EXIST on backend services (LANDMINES:18299 — zero source hits; the
#    "it scrolled" warning absorbs the real failure), and plain BusyBox
#    `grep -aq` over /proc/1/exe reports FALSE ABSENCES with both controls
#    passing (LANDMINES:16992). The working instrument is the NUL-split probe,
#    with BOTH controls through the SAME pipeline, on SYMBOLS of a called path:
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for lit in enforceListingItemSources "required query source errored" \
           queryListBelowContract zzz_invented_absent_control; do
  printf "%-38s " "$lit"
  kubectl -n ai-persona-system exec "$POD" -- sh -c \
    "tr '\0' '\n' < /proc/1/exe | grep -Fc \"$lit\"" || echo 0
done
# expected once the gate ships: >0 / >0 / >0 / 0. queryListBelowContract is the
# weeks-old present-control; grep -Fc exits 1 on zero matches (that IS "0").
# PROVEN LIVE 2026-09-02 ~21:2x BST roll: enforceListingItemSources=2,
# "required query source errored"=1, controls 1/0 — build ∈ [6525b45ae,
# c610898d1): the GATE is live; the c610898d1/2ac76f11c refinements (derived
# vocabularies, optional-error record, shared-writer receipts) ride the NEXT
# roll (their symbols read 0/0, corroborated pair).

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

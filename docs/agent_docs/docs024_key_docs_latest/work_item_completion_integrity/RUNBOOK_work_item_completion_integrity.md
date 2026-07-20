# RUNBOOK — work-item completion integrity

`PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`

## The defining query — is anything lying about completion?

```sql
SELECT id, site_id, item_type, completed_at FROM site_work_items
WHERE status='complete' AND result->'response'->>'status'='failed';
```
Returned **54** on 2026-07-18 (6 sites, 4 item types, back to May); **0** after the data
correction. **Gotcha:** after the guard ships this should only ever return rows that
predate the deploy. A non-zero result on a fresh row means the guard is not in the
running pod — check the pod, not git.

## Sizing the guard before trusting it (do this before ANY predicate like this)

```sql
-- every value response.status has EVER held. Returned exactly one row: 'failed'.
SELECT DISTINCT result->'response'->>'status' FROM site_work_items
WHERE result->'response'->>'status' IS NOT NULL;

-- blast radius + firing rate over 30 days: 1656 completions / 43 item types,
-- guard would have blocked 6, all genuine failures.
SELECT count(*) FILTER (WHERE result->'response'->>'status' IS NULL)                AS no_status,
       count(*) FILTER (WHERE lower(trim(result->'response'->>'status'))
                              IN ('failed','failure','error'))                       AS would_block
FROM site_work_items
WHERE status='complete' AND completed_at > now() - interval '30 days';
```
**Gotcha:** this is the check that turns "the predicate is safe" from an assertion into a
measurement. A guard with a measured *empty gap* between the success and failure
populations cannot mis-fire; without that gap, don't ship it.

## Verifying the dormantActions claim (guardian's objection — it IS checkable)

```sql
SELECT type, is_active FROM agent_definitions
WHERE default_config::text LIKE '%cleanup_stale_topics%';
SELECT type, is_active FROM agent_definitions
WHERE default_config::text LIKE '%load_pending_verifications%';
```
Both must return **zero rows**. If either returns a row, that action is seeded and must be
registered, not left dormant — otherwise the test is certifying the very bug it exists to
catch. **Gotcha:** `agent_definitions` keys on `type`, not `name` or `identifier`.

## Finding every path that completes a work item (multi-line SQL defeats naive grep)

```bash
for f in $(grep -rl "site_work_items" --include=*.go platform/ internal/ cmd/); do
  python3 - "$f" <<'EOF'
import re,sys
p=sys.argv[1]; s=open(p).read()
for m in re.finditer(r"UPDATE\s+site_work_items(.{0,400}?)(?:`|\";)", s, re.S|re.I):
    if re.search(r"status\s*=\s*'complete'", m.group(1), re.I):
        print(f"{p}:{s[:m.start()].count(chr(10))+1}")
EOF
done
```
Returns 8 paths. **Gotcha — and this one bit me:** the regex tells you WHERE, not WHAT.
You must then OPEN each one. Four loop paths complete on self-gathered evidence; the three
admin paths build their result with `jsonb_build_object` from human input and never touch a
response envelope. I asserted that from filenames before reading them and the council
objected twice, correctly.

## Data correction (reversible)

Full script pattern in `NOTES`. Key points: snapshot targets into a TEMP table first so the
same rows are used throughout; write the prior status to `result._correction.prior_status`
so revert is one UPDATE; verify the defining query returns 0 inside the transaction before
COMMIT.

## Deploy verification — pod binary, NOT git, NOT the tag

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "handlerReportedFailure"'
```
**Gotcha (016b §9, another thread):** a 0 from a pod-grep is not automatically proof the
image is stale — grep an older known symbol as a control, and know that `strings` can miss
symbols depending on build flags. Treat 0 as "investigate", not "rebuild".

## Is my council/diagnosis run dropped, or just queued?

```sql
SELECT orchestration_id, min(changed_at) AS started, max(changed_at) AS last
FROM orchestration_state_audit WHERE changed_at > now() - interval '45 minutes'
GROUP BY 1 ORDER BY started DESC LIMIT 10;
```
**Gotcha — cost me three redundant council runs on 2026-07-18.** Absence of rows for YOUR
orchestration measures queue depth, not delivery. Latency was ~10s quiet, **~16 minutes**
under backlog. Ask when OTHER orchestrations started before concluding anything. Full
pattern: 016b §9 "A queued orchestration is indistinguishable from a dropped one".

## Verifier coverage (bugs_open/021 §INSTANCE 2)

**Refresh the coverage guard's denominator.** `verifier_coverage_test.go` holds a
hand-maintained list of live item types, because they cannot be derived in Go: the
check registry keys on check NAME and item types are string literals inside each
`Run`, and the highest-volume types come from paths with no discovery check at all.

```sql
-- the denominator. Paste into liveItemTypes, update the "refreshed" date.
SELECT DISTINCT item_type FROM site_work_items ORDER BY 1;

-- the shape of the gap: volume + which target ids are available per type
SELECT item_type,
       count(*) FILTER (WHERE status='complete')      AS completed,
       count(*) FILTER (WHERE spec ? 'page_id')       AS has_page_id,
       count(*) FILTER (WHERE spec ? 'component_id')  AS has_component_id
FROM site_work_items GROUP BY 1 ORDER BY 2 DESC;

-- coverage today
SELECT count(*) FILTER (WHERE status='complete')                AS complete_total,
       count(*) FILTER (WHERE status='complete' AND result ? '_verification') AS verified
FROM site_work_items;   -- 2026-07-20: 4,644 / 5
```

**Gotcha — the one that cost this thread a build.** Do NOT conclude an item type is
verifiable from its *detector's* predicate. A verifier asserts the HANDLER did its
job, so read the handler's remit first. `page_rerender` looked ideal (1,849
completions, page_id on 1,914 of 1,929) but its handler only rewrites CTA fields in
six component types; a whole-page predicate would mark correctly-handled items
unresolved and destroy the designed two-strike escalation. See WRONG_CALLS.md
2026-07-20 and the `page_rerender` entry in the guard's gap map.

**Gotcha — `go test` output is the report.** `TestVerifierCoverageIsReported` never
fails; run it with `-v` to see coverage by category:

```bash
go test ./platform/orchestration/actions/discovery_checks/ -run TestVerifierCoverage -v
```


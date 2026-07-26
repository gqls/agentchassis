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


---

## Handler coverage — is every check routing at an agent that exists? (2026-07-26)

```bash
go test ./platform/orchestration/actions/discovery_checks/ -run TestHandlerCoverage -v
```

`TestEveryCheckHandlerAgentExistsOrIsADeclaredGap` fails the build; the `IsReported`
twin never fails and prints the picture. Refresh `knownHandlerAgents` **by UNION,
never replacement** — same rule as `liveItemTypes`, same reason:

```sql
SELECT DISTINCT type FROM agent_definitions WHERE deleted_at IS NULL ORDER BY 1;
```

**Gotcha — a guard nobody has watched fail is not known to work.** The unit probes
exercise the assertion function against a fabricated world; they do NOT exercise
the source scan, which is the half that can silently stop matching. Prove the whole
thing by inducing a real fault and restoring it:

```bash
sed -i 's/HandlerAgent: "webdesign-agent",/HandlerAgent: "induced-fault-fixer",/' \
  platform/orchestration/actions/discovery_checks/check_generic_theme.go
go test ./platform/orchestration/actions/discovery_checks/ -run TestEveryCheckHandlerAgent   # must FAIL, naming the agent
git checkout platform/orchestration/actions/discovery_checks/check_generic_theme.go
go test ./platform/orchestration/actions/discovery_checks/ -run TestEveryCheckHandlerAgent   # must PASS
```

## Does a check file only what its handler can fix? (bugs_open/077)

Measure the remit split for a check before trusting its item counts. Full method in
`durable_write_guard/RUNBOOK_durable_write_guard.md` §"Know the expected verdict"
(dump the population with `row_to_json`, run the shipped Go transform over it).

For `hardcoded_section_colors` there is a SQL shortcut, and its limits matter:

```sql
-- STRICTLY WIDER than ReplaceHardcodedColors on every axis: no <style> boundary,
-- no trailing terminator, no restriction to the detector's own population.
SELECT s.domain, count(*) AS detector_matches,
       count(*) FILTER (
         WHERE pc.rendered_html ~ 'background(-color)?\s*:\s*#[0-4][0-9a-fA-F]{5}'
            OR pc.rendered_html ~ 'linear-gradient\s*\(\s*[0-9]+deg\s*,\s*#[0-9a-fA-F]{3,8}\s*,\s*#[0-9a-fA-F]{3,8}\s*\)'
       ) AS remit_superset
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON s.id = p.site_id
WHERE pc.locked_at IS NULL
  AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
  AND pc.rendered_html LIKE '%<style%'
GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha — a superset proves zero; it can NEVER disprove it.** `remit_superset = 0`
is proof the handler's remit is empty on that site. `remit_superset = 1` proves
nothing at all — the true remit may still be 0, because the superset deliberately
matches things the Go transform will not touch. This thread read a `1` as
contradicting a previous thread's `0` and wrote a "correction" that was itself the
error (`WRONG_CALLS.md`, 2026-07-26). If you need a non-zero answer, run the Go
transform; only use this query to establish zeros.

## Turning a capability gap into queued build work

The residue items are the intake for the feature builder, grouped by the agent
they need:

```sql
SELECT spec->>'builder_needed' AS builder, spec->>'gap_kind' AS kind,
       count(*) AS items, count(DISTINCT site_id) AS sites
FROM site_work_items
WHERE item_type = 'capability_gap' AND status = 'deferred'
GROUP BY 1, 2 ORDER BY 3 DESC;
```

The designer's `check_spec_approved` gate refuses anything without BOTH an
`owner_approval` and `code_pointers` in the spec. The checks already write
`code_pointers`; the approval is a human act and stays one:

```sql
UPDATE site_work_items
   SET spec = spec || '{"owner_approval": {"approved_by": "<name>", "date": "YYYY-MM-DD"}}'
 WHERE id = '<work_item_id>';
```

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh <work_item_id>
# SAVE the printed FEATURE_CORR — it keys the staged plan and the council artifacts.
```

**Gotcha — the cheapest gap is not the one with the most items.**
`forced-text-color-fixer`'s action (`fix_forced_text_colors`) is *already written
and already registered*; only the `agent_definitions` row is missing. But do not
seed it and stop: that action bails out entirely below its WCAG contrast floor and
only rewrites text-element selectors, so seeding it without also partitioning
`check_forced_text_colors` re-creates `bugs_open/077` under a new item type.

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

## Proving the check RAN (do NOT reach for the runner's log line)

The discovery runner records its own arrays in `collected_data.run_checks`. Structured, no shelf
life, names the check individually — strictly better than the log grep this RUNBOOK used to imply:

```sql
SELECT (collected_data->'run_checks'->'checks_run') @> '["cta_rank_anomaly"]'::jsonb AS ran,
       jsonb_array_length(collected_data->'run_checks'->'checks_run')  AS n_run,
       collected_data->'run_checks'->'checks_unregistered' AS unregistered,   -- must be []
       collected_data->'run_checks'->'checks_failed'       AS failed,         -- must be []
       collected_data->'run_checks'->'items_inserted', collected_data->'run_checks'->'items_resolved'
FROM orchestration_states
WHERE owner_agent_type='completeness-discovery-agent'
  AND collected_data->'input_data'->>'domain'='<domain>'
ORDER BY created_at DESC LIMIT 1;
```

⚠ `owner_agent_type`, NOT `agent_type` — `orchestration_states` has no such column.

## The prediction census — what SHOULD fire, so that "0 items" means something

Mirrors `datahelpers/cta_positional.go` exactly: supply predicate + eligibility + excluded areas +
`(nav_order, name)`. Run it BEFORE reading any zero. **Re-run it rather than quoting the 4 sites of
2026-09-03 — a census goes stale by addition.**

```sql
WITH cand AS (
  SELECT p.site_id, s.domain, p.name, COALESCE(p.nav_order,100) AS nav
  FROM pages p JOIN sites s ON s.id=p.site_id
  WHERE p.page_type IN ('tool','game') AND p.status IN ('active','deployed')
    AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'')='planned')   -- PageMayBeLinkedPredicateFor
    AND p.eligible_as_cta_target
    AND lower(regexp_replace(split_part(ltrim(p.url,'/'),'/',1),'\.html$','')) NOT IN
        ('about','contact','privacy','terms','legal')                            -- CTAExcludedAreas
), ranked AS (
  SELECT site_id, domain, name, nav,
         row_number() OVER (PARTITION BY site_id ORDER BY nav, name) AS rk,
         count(*)    OVER (PARTITION BY site_id) AS n
  FROM cand
), top2 AS (
  SELECT site_id, domain, n,
         max(name) FILTER (WHERE rk=1) AS r1, max(nav) FILTER (WHERE rk=1) AS nav1,
         max(name) FILTER (WHERE rk=2) AS r2, max(nav) FILTER (WHERE rk=2) AS nav2
  FROM ranked WHERE rk<=2 GROUP BY site_id, domain, n
)
SELECT domain, n AS candidates, r1, nav1, r2, nav2, (nav2-nav1) AS lead
FROM top2 WHERE n >= 3 AND nav1 < 100 AND nav2 <> nav1 AND (nav2-nav1) >= 50
ORDER BY lead DESC;
```

A site with a below-default rank-1 that is ABSENT from this list is the useful control — check which
arm excluded it (idea.uk: lead 7, the curated ladder).

## Inducing ONE discovery run, with an asserted publish receipt

⛔ **Do NOT run `scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh`.** Its
tail is hard-coded to finetuning.uk and runs
`UPDATE site_work_items SET status='triaged' … WHERE site_id=(… 'finetuning.uk') AND status='detected'`
**regardless of the domain you passed** — so triggering discovery for site A silently triages site B's
queue. It also uses the racing `kubectl run -i … kcat -P` stdin form (LANDMINES: ~4 publishes in 5
lost at exit 0).

Publish through the library instead (OPP-009, `scripts/kafka-publish-lib.sh`) so a silent drop is
distinguishable from queue latency. Payload = the same `spawn_discovery → call_discovery → complete`
workflow the 075 script carries, built with `jq -c` (one message per line, or kcat publishes
fragments); headers `action=process`, `sender_agent_type=cli`. Then `kafka_verify_landing "$CORR" 120`.
Measured 2026-09-03: publish → COMPLETED in **~35–110 s** for a discovery run — this one does not
queue for half an hour the way a build dispatch does.

⚠ Check no chassis pod (re)started within ~300 s first (`kubectl -n ai-persona-system get pods
-l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.startTime}{"\n"}{end}'`)
— a spawn inside that window is silently dropped.

## The induced canary — what it can and cannot show

Two-way at the RANKING (this is the part that verifies the lever, and it needs no render):

```sql
-- 1. before: note the candidate COUNT in the check's detail/reason string
-- 2. UPDATE pages SET eligible_as_cta_target=false WHERE site_id=$1 AND name='<rank1>';
-- 3. re-run discovery -> items_resolved:1, item -> complete, reason quotes the DECREMENTED count
-- 4. UPDATE ... =true;  re-run -> finding detail quotes the original count, items_inserted:1
```

**Assert on the candidate count in the reason/detail string.** It is the only observable that could
only have been produced by `RankCTAPositionalCandidates` reading the live column, and it moves in both
directions. Keep a second fossil site untouched across all four runs as the control that the
resolution was site-specific.

A resolve does NOT poison the item key against a later genuine re-fire (measured: `items_inserted: 1`
on the flip-back, a fresh row beside the `complete` one) — so do not skip direction 2 on dedup grounds.

**What the canary CANNOT show, in this order of difficulty:**
- **The header button at the served bytes** needs `rerender-pages` with
  `refresh_site_components: true` (without it the run reassembles from stored
  `site_components.rendered_html` and the header cannot move) — and the *served* header only exercises
  the ranking fallback on a site with **no footer-group nav item labelled "contact"**
  (`render_site_components_action.go:105,162` — the match is on `site_nav_items.label` over groups
  primary/utility/legal, NOT on `pages.in_header`). On a site that has one, the header CTA is the
  contact URL and the fallback never runs. Query for eligible canaries before picking one; of the 12
  such sites on 2026-09-03, cv1.co.uk was the only one also fossil-shaped.
- **The stored `content_data` CTA fields** cannot move on a rerender at all: `applyCTARecompute`
  KEEP #2 holds any valid stored destination (PLAN, "what this deliberately does NOT do"). That needs
  a full page rebuild, which regenerates copy.

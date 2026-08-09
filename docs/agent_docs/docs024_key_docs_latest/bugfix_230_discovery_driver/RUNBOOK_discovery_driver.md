# RUNBOOK — bugfix 230 discovery driver

## 1. Validity / state queries

```sql
-- the driver gap (bug 230 §2): must filter on target_agent_type, NEVER on the word
-- 'discovery' — the model-directory research tasks match the word and ARE enabled
SELECT name, enabled, last_triggered_at FROM scheduled_tasks
WHERE target_agent_type IN ('quality-discovery-agent','completeness-discovery-agent','design-discovery-agent');

-- the worked case (finetuning.uk, bug 230 §6) — empty until the fix is live and the
-- completeness rotation has reached the site, then gains a featured-content row on its own
SELECT item_key, status, created_at FROM site_work_items
WHERE item_type='empty_section'
  AND page_id IN ('69a50d5d-3732-4efe-9a79-f887b072fa86','8867b4d5-12d1-4ecc-8956-109a80395a18')
ORDER BY created_at DESC;
```

## 2. Apply the migration (after council verdict)

```bash
./scripts/migration/run-migrations.sh            # dry run FIRST — lists pending; check
                                                 # 346 is the ONLY unexpected pending file
./scripts/migration/run-migrations.sh --apply    # applies EVERY pending file in order —
                                                 # if another session has parked one, stop
                                                 # and coordinate before applying
```

Gotcha (migration-runner-practice): `--apply` takes every pending file. Dry-run per
session and after every roll.

## 3. Watch the rotation work

```sql
-- stamps advancing = rotation alive; every deployed site should appear within ~1 day/agent
SELECT r.agent_type, count(*) AS stamped, min(r.last_selected_at), max(r.last_selected_at)
FROM site_discovery_rotation r GROUP BY 1;

-- who's next per agent (the pre_query's own ordering, read-only)
SELECT s.domain, r.last_selected_at FROM sites s
LEFT JOIN site_discovery_rotation r ON r.site_id=s.id AND r.agent_type='completeness-discovery-agent'
WHERE s.status IN ('active','deployed') ORDER BY r.last_selected_at ASC NULLS FIRST, s.id LIMIT 5;
```

Gotchas:
- `last_triggered_at`/`last_completed_at` on the task rows are **fire-and-forget stamps**
  (LANDMINES) — they prove the scheduler fired, never that an orchestration ran. Confirm at
  `orchestration_states` (24h retention) or at the items written.
- A chassis roll makes scheduled tasks look broken for ~5 min, and a dispatch within ~300s
  of a pod restart is silently dropped (LANDMINES / CLAUDE.md). The rotation self-heals
  (the site stays most-due only until its next period), and the watchdog's closer check
  catches a *recurring* drop within a day.

## 4. The knobs (owner-adjustable, live immediately)

```sql
-- pause everything
UPDATE scheduled_tasks SET enabled=false WHERE name LIKE 'site-discovery-rotation-%';
-- change per-site period (edit the interval inside each pre_query, default '7 days')
-- change polling cadence
UPDATE scheduled_tasks SET interval_seconds=7200 WHERE name LIKE 'site-discovery-rotation-%';
```

## 5. Deploy the watchdog

```bash
kubectl apply -k deployments/kustomize/services/site-discovery-staleness-check/base
# fire one run now rather than waiting for 06:35 UTC:
kubectl -n ai-persona-system create job --from=cronjob/site-discovery-staleness-check sdsc-manual-$(date +%s)
# read its report (also written on CLEAN — a missing row means the job did not run):
# SELECT body FROM doc_notes WHERE categories ? 'site-discovery-staleness' ORDER BY created_at DESC LIMIT 1;
```

## 6. Council

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_230_discovery_driver/council_submission.json
# budget ~30 min queue; find the run by PAYLOAD, not the printed id:
# SELECT current_step, status FROM orchestration_states
#  WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

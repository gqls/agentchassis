# RUNBOOK — model directory pipeline

Commands that were hard to get right, with the gotcha attached. Updated the
moment a command changes — not from scrollback later.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Site under pilot

`ai-agent-orchestration.com` = site_id `2a8ebf9c-20a2-4c39-b191-840b012371da`.

## Checking migration numbering before filing a new one

Another session may have claimed the next number since this doc was last
updated — always re-check IMMEDIATELY BEFORE APPLYING, not just when drafting
the file (a collision was caught this way on 2026-07-22: `191` was taken
concurrently by an unrelated session between draft and apply — ours is `192`):

```
ls docs/agent_docs/sql_for_agents/ | grep -oE '^[0-9]+' | sort -n | tail -5
git status --short docs/agent_docs/sql_for_agents/   # catches untracked concurrent files too
git log --oneline -5 -- docs/agent_docs/sql_for_agents/
```

## Applying a migration

```
./scripts/migration/run-migrations.sh            # dry run, lists pending
./scripts/migration/run-migrations.sh --apply    # applies + records, in order
```
This applies ALL pending files >= baseline in order, not just yours — check
the dry-run output first to see whether anything else is pending that isn't
yours (another session's filed-but-unapplied migration).

## Verifying the site_specs opt-in write took

```sql
SELECT aspect, is_current, jsonb_pretty(data->'content_features'->'model_directory')
FROM site_specs
WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect = 'classification' AND is_current;
```

## Verifying a citation is still live (manual spot-check, mirrors verifyCitationLive)

```
curl -sL '<citation.url>' | grep -F '<citation.quote>'
```
Empty output = the quote is gone; the automated freshness sweep should have
flipped that claim's `status` to `citation_lost`, not left it `found`.

## scheduled_tasks target_topic for a custom agent type

A custom `target_agent_type` that runs on the shared agent-chassis (not its
own microservice like business-intel/vet-intel) must still use
`target_topic = 'system.agent.generic.requests'` — NOT
`system.agent.<type>.requests`. The real type travels inside the message
payload (`config.agent_type`), read by whichever pod consumes the generic
topic. Verify the pattern against a known-live example before trusting a new
one:
```sql
SELECT name, target_agent_type, target_topic FROM scheduled_tasks
WHERE name = 'content-feed-refresh';
-- expect target_agent_type='content-feed-trigger', target_topic='system.agent.generic.requests'
```

## Re-firing a discovery run (and watching it without fooling yourself)

```sql
UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name = 'model-directory-discovery';
```
The scheduler picks it up on its next ~30s tick. Gotchas this arc paid for:

- **Find the run by its workflow, not by input filters.** A filter on
  `initial_request_data->>'research_query'` missed a real row and read as
  "nothing dispatched" (WRONG_CALLS 2026-07-24). Use:
  ```sql
  SELECT orchestration_id, status, current_step, error, created_at
  FROM orchestration_states
  WHERE workflow_plan::text ILIKE '%batch_webscrape%'
  ORDER BY created_at DESC LIMIT 3;
  ```
- **Pin which binary processed the run before crediting/blaming a change:**
  compare the run's `created_at` against the pod's start time —
  `kubectl -n ai-persona-system get rs -l app=agent-chassis --sort-by=.metadata.creationTimestamp`
  — a fixed-delay watcher launched "after" a rollout command races the
  rollout itself (run 6 fired against the old pod this way).
- **After a chassis (re)start, wait ≥300s before firing** — the spawn is
  silently dropped inside that window (CLAUDE.md standing rule).
- **Read the ADAPTER's logs inside the failure window, immediately** — fleet
  pod rolls destroy them within hours (lost twice in this arc):
  ```
  kubectl -n ai-persona-system logs <web-scrape-adapter-pod> --since=45m \
    | grep -E "Batch scrape completed|Failed to produce|<request_id>"
  ```
- A caller-side `Request ... timed out after 3 retries` says NOTHING about
  where the loss is — runs 1–4 each failed with that identical error for
  four different mechanisms (transient, oversize reply, oversize reply,
  unparseable reply). The callee's logs discriminate; the caller's error
  never does.

## Getting the page created: discovery does NOT run on a timer

The `missing_model_directory_*` checks only fire when a
completeness-discovery run happens AT the site — and there is no scheduled
task for that: it is pipeline-triggered, and the improvement-sweep that
would drive it is deliberately off fleet-wide (see the
content_quality_and_internal_linking runbooks). On aao, discovery had not
run since **2026-05-02**; the checks were enabled but never got a turn.
Dispatch one manually (proven 2026-07-24, corr `03ee816c`):

```bash
# kcat orchestrate pattern (cloned from cta_link_integrity/scripts/049b_deploy_single_page.sh)
kubectl -n kafka run -i --rm "kcat-disc-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$(cat /proc/sys/kernel/random/uuid) ... <<JSON
{"action":"orchestrate","config":{"agent_type":"completeness-discovery-agent"},"input_data":{"site_id":"<id>","domain":"<domain>"}}
JSON
```

Then: **findings land `status='detected'`, which is UNCLAIMABLE** — the
dispatch loop only picks up `triaged`. With the sweep off, triage the two
model-directory items by hand:

```sql
UPDATE site_work_items SET status='triaged', updated_at=now()
WHERE site_id='<id>' AND item_type IN
  ('missing_model_directory_section','missing_model_directory_page')
AND status='detected';
```

Expect a 3-months-unswept site to also surface a large backlog of OTHER
findings (aao 2026-07-24: 7 phantom links, 20 unknown-destination CTAs, 25
required_fields, 1 truncated_component, ~19 page_rerenders) — most park at
`needs_human_review` by design; leave them for their owning workstreams.

## Publish leg (model-directory-publish, live 2026-07-24)

Self-gating: idles (`complete_idle`) until an opted-in site has a DEPLOYED
page carrying a `model-directory`/`model-directory-listing` component AND
the registry has `found` claims. Verify a cycle ran:
```sql
SELECT name, last_triggered_at, last_completed_at FROM scheduled_tasks
WHERE name = 'model-directory-publish';
```
Once pages exist: expect `data/model-directory.json` committed in the site
repo and `page_rerender:<page>` work items queued.

## Checking directory_claims state

```sql
SELECT de.slug, dc.field, dc.value, dc.status, dc.verified_at
FROM directory_claims dc JOIN directory_entities de ON de.id = dc.entity_id
WHERE dc.is_current
ORDER BY de.slug, dc.field;
```

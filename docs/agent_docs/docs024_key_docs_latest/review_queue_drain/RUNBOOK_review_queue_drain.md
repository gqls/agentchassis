# RUNBOOK — review queue drain (`bugs_open/033`)

Every query and command that was hard to get right, with its gotcha attached.
When one changes, change it HERE.

---

## Queue depth and shape

```sql
SELECT status, count(*) FROM site_work_items GROUP BY 1 ORDER BY 2 DESC;

SELECT item_type, count(*), min(created_at)::date AS oldest, max(created_at)::date AS newest
FROM site_work_items WHERE status='needs_human_review' GROUP BY 1 ORDER BY 2 DESC;
```

## Has anything EVER been actioned through the admin surface?

```sql
SELECT count(*) FILTER (WHERE result->>'resolved_by'='admin') AS via_api_resolve,
       count(*) FILTER (WHERE result ? 'approved_by')         AS via_api_approve,
       count(*) FILTER (WHERE approved_by IS NOT NULL)        AS approved_by_col
FROM site_work_items;
```

**Gotcha:** `approved_by` is a *column* and `resolved_by` lives *inside* `result`
— the API writes the jsonb, never the column. Checking only the column reads as
"never used" even if the API had been used.

## How stale is the queue?

```sql
SELECT w.item_type, count(*) AS n,
  count(*) FILTER (WHERE EXISTS (
    SELECT 1 FROM pages p
    WHERE p.site_id=w.site_id AND p.name = w.spec->>'page_name'
      AND p.deployed_at IS NOT NULL AND p.deployed_at > w.created_at)) AS page_deployed_since
FROM site_work_items w
WHERE w.status='needs_human_review' AND w.spec->>'page_name' IS NOT NULL
GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha:** join on `spec->>'page_name'`, not on `w.page_id` — only 155 of 370
carry `page_id`, so keying on the column undercounts by more than half.

## Is a given parked finding still true? (the manual version of the sweep)

```sql
SELECT p.deployed_at, pc.slot_name,
       pc.content_data->>'cta_url', pc.content_data->>'secondary_cta_url'
FROM pages p JOIN page_components pc ON pc.page_id=p.id JOIN sites s ON s.id=p.site_id
WHERE s.domain='<domain>' AND p.name='<page>'
ORDER BY pc.position;
```

**Gotcha — the one that cost the most time.** Do **not** key on
`spec->>'component_id'`. `page_components.id` is not stable across re-renders, so
a stale id reads as "the component was deleted" when the section is right there
under a new row id. Keyed on `component_id`, 30 of 30 parked `needs_section_data`
items look orphaned; keyed on `(page_name, slot_name)`, none are.

**Gotcha 2.** `content_data IS NULL` on a component does *not* mean the slot is
gone — the row exists and renders from a template / DERIVED source / static
fallback. 31 of 45 `required_fields_missing` items are in that state. `content_data`
cannot answer the question for them, so the sweep returns `unknown`.

## Which item types can the sweep judge?

```sql
-- what the sweep would report as uncovered, without running it
SELECT item_type, count(*) FROM site_work_items
WHERE status='needs_human_review'
  AND item_type NOT IN ('unresolved_cta','required_fields_missing','needs_section_data')
GROUP BY 1 ORDER BY 2 DESC;
```

## Deploy sequence (image before seed, always)

```bash
# 1. commit, then build from committed HEAD
make build-agent-chassis
# 2. bump IMAGE_TAG (makefile ~line 16) — a same-tag rebuild ships the stale cached binary
# 3. push + deploy, then verify against the RUNNING POD, never git, never the tag:
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "auto:revalidated"'     # >= 1
```

**Gotcha:** grep a string the change *created* (`auto:revalidated`), not one it
merely uses (`site_work_items`, `needs_human_review`) — those are in every older
binary and prove nothing. Positive control: the same grep on the *previous* image
must return 0.

## Apply the seed and fire the sweep

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/docs024_key_docs_latest/review_queue_drain/seed_review_queue_revalidator.sql

./docs/agent_docs/docs024_key_docs_latest/review_queue_drain/TRIGGER_revalidate_review_queue_v1.sh
```

**Gotcha:** no orchestration dispatch within ~300s of a chassis pod (re)start —
the spawn is silently dropped. And a missing `orchestration_states` row is almost
always dispatch latency (~30 min under load), **not** a dropped message. Find the
run by the `orchestration_id` the trigger printed, never by `created_at`.

## Read the run

```sql
SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit
WHERE orchestration_id='<RUN_ORCH_ID>' ORDER BY changed_at;

SELECT jsonb_pretty(collected_data->'complete'->'result'->'response')
FROM orchestration_states WHERE orchestration_id='<RUN_ORCH_ID>';
```

## What the sweep did, and undoing it

```sql
-- closed, with the evidence each close rests on
SELECT id, item_type, result->'revalidation'->>'reason'
FROM site_work_items WHERE resolution_path='auto:revalidated' ORDER BY completed_at DESC;

-- survivors + their re-confirmation stamp (the trust signal)
SELECT item_type, result->'revalidation'->>'verdict', result->'revalidation'->>'at', count(*)
FROM site_work_items WHERE status='needs_human_review' AND result ? 'revalidation'
GROUP BY 1,2,3 ORDER BY 4 DESC;

-- full reversal: every close is self-identifying
UPDATE site_work_items SET status='needs_human_review', completed_at=NULL,
       resolution_path=NULL, result = result - 'revalidation'
WHERE resolution_path='auto:revalidated';
```

**Note:** reversal is rarely the right move. Closing an item releases its dedup
key, so if a close was wrong the originating check re-raises the finding fresh on
its next run — usually better than restoring a row with a stale `created_at`.

## Flipping out of dry_run

```sql
-- narrow the first live run to one class
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(default_config, '{workflow,steps,sweep,config,dry_run}', 'false'::jsonb),
      '{workflow,steps,sweep,config,item_type}', '"unresolved_cta"'::jsonb),
    updated_at = now()
WHERE type='diagnosis-review-queue-revalidator' AND is_active AND COALESCE(is_snapshot,false)=false;

-- then widen
UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,sweep,config,item_type}', updated_at=now()
WHERE type='diagnosis-review-queue-revalidator' AND is_active AND COALESCE(is_snapshot,false)=false;
```

**Gotcha:** always add `AND COALESCE(is_snapshot,false)=false` — `snapshot_agent`
leaves historical rows of the same type, and updating those changes nothing live
while looking like it worked.

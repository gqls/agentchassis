# 016 — Debugging Guide

Practical steps for diagnosing and fixing problems in the pipeline. Based on real failure patterns.

**Schema reminder:** The `site_work_items` column for work category is `pipeline` (not `domain`). The site's website domain comes from the `sites` table or `input_data.domain`. See 007_adoption_pipeline schema notes for other renamed columns.

---

## 1. Pod Health Check

Start here. The pod list tells you most of what you need to know.

```bash
kubectl -n ai-persona-system get pods
```

**What to look for:**

- **Pending pods** — cluster resource exhaustion. Check node capacity with `kubectl top nodes`.
- **Many Running agent pods with high ages (hours)** — zombie pods not self-terminating. Check `idle_timeout_seconds` on their agent definitions: `SELECT type, idle_timeout_seconds FROM agent_definitions WHERE type = '<agent-type>'`. If 0, the idle monitor never starts.
- **CrashLoopBackOff** — container startup failure. Check logs: `kubectl -n ai-persona-system logs <pod> --previous`.
- **Completed pods accumulating** — normal, cleaned by TTL (1 hour after finish) and job-cleanup CronJob.

**Pod distribution across nodes:**

```bash
kubectl -n ai-persona-system get pods -o wide --no-headers | awk '{print $7}' | sort | uniq -c | sort -rn
```

If all pods are on one or two nodes with new nodes empty, existing pods won't rebalance — only new spawns land on new nodes. Kill stale jobs to free resources: `kubectl -n ai-persona-system delete jobs -l app=dynamic-agent`.

---

## 2. Work Item Status

The work items table is the dispatch queue. Check what's stuck, failed, or accumulating.

```sql
SELECT wi.item_type, wi.status, s.domain, LEFT(wi.summary, 60)
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.pipeline = 'build' AND wi.status != 'complete'
ORDER BY wi.created_at DESC;
```

**Count by status to spot patterns:**

```sql
SELECT status, COUNT(*) FROM site_work_items
WHERE pipeline = 'build' GROUP BY status ORDER BY COUNT(*) DESC;
```

**For failed items, check the error messages:**

```sql
SELECT wi.item_type, wi.handler_agent, s.domain,
       wi.attempt_count || '/' || wi.max_attempts as attempts,
       LEFT(wi.error, 120) as error
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'failed' AND wi.pipeline = 'build'
ORDER BY wi.created_at DESC;
```

**Common error patterns and what they mean:**

| Error message | Root cause |
|---|---|
| `Claim timed out (attempts exhausted)` | Handler took longer than the claim timeout (30 min). Either the handler is genuinely slow (rerender of large sites) or the pod died mid-work. |
| `Handler failed` | The dispatch loop's `call_handler` step timed out or the spawned handler returned an error. Check `agent_error_log` for details. |
| `Request X timed out after N retries` | The `call_agent` step waited for a child response that never came. The child pod likely died or was never created (resource starvation). |
| `Content validation failed` | page-build-handler's content validator found placeholders, unrendered Go templates, or cross-site company name contamination. Item goes to `needs_human_review`. |
| `query param path 'X' resolved to nil` | The handler workflow references a field (e.g. `input_data.component_id`) that doesn't exist at that path. Usually a mismatch between the dispatch loop's `input_mapping` and the handler's expected paths. The dispatch maps the work item spec as a nested object at `input_data.spec`, but the handler tries to read fields at `input_data.<field>` directly. See section 9. |
| `Handler agent not registered: <agent>` | Work items reference a `handler_agent` that has no matching `agent_definitions` row. Items stay `blocked` forever. Check section 6 to find all missing handlers. |

---

## 3. Scheduled Tasks

Check what's firing, what's stuck, and what's blocked.

```sql
SELECT name, 
       CASE 
         WHEN last_triggered_at IS NOT NULL 
           AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
           AND last_triggered_at + (timeout_seconds || ' seconds')::interval > NOW()
         THEN 'IN-FLIGHT'
         ELSE 'idle'
       END as flight_status,
       concurrency_group,
       last_triggered_at,
       last_completed_at,
       timeout_seconds,
       interval_seconds
FROM scheduled_tasks
WHERE enabled = true
ORDER BY flight_status DESC, name;
```

**Task not firing? Check in order:**

1. Is it enabled? `SELECT enabled FROM scheduled_tasks WHERE name = '...'`
2. Has the interval elapsed since `last_triggered_at`?
3. Is its concurrency group at capacity? (another task in the same group is in-flight)
4. Is the pre_query returning no rows? Run it manually.
5. Is `target_topic` correct? For a task that invokes an agent via the generic entry point, it should be `system.agent.generic.requests`. For a task that talks directly to a long-lived adapter Deployment, it's that adapter's fixed topic. It is never a `job.*` topic — those are created per-spawn and are not reachable from outside a spawning workflow.

**Concurrency group stuck:**

```sql
SELECT name, concurrency_group, last_triggered_at, last_completed_at,
       last_triggered_at + (timeout_seconds || ' seconds')::interval as times_out_at
FROM scheduled_tasks
WHERE concurrency_group = '<group>'
  AND last_triggered_at IS NOT NULL
  AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
ORDER BY last_triggered_at;
```

**Force-unstick:**

```sql
UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = '<blocking-task>';
```

---

## 4. Orchestration States

When agents are running but not completing, check their orchestration state.

```sql
SELECT orchestration_id, owner_agent_type, status, current_step,
       error_message, updated_at, NOW() - updated_at as stale_for
FROM orchestration_states
WHERE status IN ('EXECUTING_STEP', 'WAITING_FOR_RESPONSE')
ORDER BY updated_at DESC
LIMIT 20;
```

**Stale orchestrations (updated > 30 min ago)** are likely orphaned — the pod died but the orchestration wasn't cleaned up. The `stale-orchestration-reaper` scheduled task handles these after 24 hours, but you can fail them manually:

```sql
UPDATE orchestration_states
SET status = 'FAILED', error_message = 'Manual cleanup — stale orchestration'
WHERE orchestration_id = '<id>';
```

---

## 5. Agent Error Log

Persistent error log for handler failures.

```sql
SELECT agent_type, error_type, LEFT(error_message, 100),
       occurred_at, orchestration_id
FROM agent_error_log
ORDER BY occurred_at DESC
LIMIT 20;
```

Filter for a specific site's issues:

```sql
SELECT ael.agent_type, ael.error_type, LEFT(ael.error_message, 100), ael.occurred_at
FROM agent_error_log ael
WHERE ael.context->>'domain' = 'example.com'
   OR ael.context->>'site_id' = '<site-uuid>'
ORDER BY ael.occurred_at DESC
LIMIT 20;
```

---

## 6. Handler Agent Definitions

When a handler fails, check that it exists and is active.

```sql
SELECT type, status, idle_timeout_seconds, 
       default_config->'workflow'->'start_step' as start_step,
       default_config->'workflow'->'timeout_seconds' as workflow_timeout
FROM agent_definitions
WHERE type = '<handler-agent>' AND deleted_at IS NULL;
```

**Check if a work item's handler exists:**

```sql
SELECT DISTINCT wi.handler_agent,
       CASE WHEN ad.type IS NOT NULL THEN 'exists' ELSE 'MISSING' END as agent_status
FROM site_work_items wi
LEFT JOIN agent_definitions ad ON ad.type = wi.handler_agent AND ad.deleted_at IS NULL
WHERE wi.status IN ('triaged', 'failed') AND wi.pipeline = 'build'
ORDER BY agent_status DESC, wi.handler_agent;
```

---

## 7. Timeout Chain

Three timeouts interact and must be ordered correctly.

```
claim_timeout (scheduled task) > call_handler timeout (dispatch loop) > workflow timeout (handler agent)
```

Currently:

| Timeout | Value | Set where |
|---|---|---|
| Claim timeout | 30 min | `claimed-item-timeout` pre_query |
| Dispatch call_handler | 1200s (20 min) | `build-dispatch-loop` workflow config |
| Handler workflow | varies (120-600s) | Each handler's `agent_definitions.default_config` |
| Idle monitor | 3600s default | `spawn_actions.go` fallback when definition has 0 |
| K8s ActiveDeadline | 86400s (24h) | Job spec hard ceiling |

**If claim_timeout < call_handler timeout:** the claim gets reset while the dispatch is still waiting for the handler. The dispatch eventually times out, marks failed, but a different dispatch already picked up the reset item and started a new handler. Two handlers now run for the same item.

**If call_handler timeout < handler workflow timeout:** the dispatch gives up and marks the item failed while the handler is still working. The handler finishes, but nobody is listening for its response.

---

## 8. Cleaning Up Failed Items

Failed items with duplicates need careful handling due to the dedup index.

```sql
BEGIN;

-- Clear FK references to failed items
UPDATE site_work_items SET parent_item_id = NULL
WHERE parent_item_id IN (
    SELECT id FROM site_work_items WHERE status = 'failed' AND pipeline = 'build'
);

-- Delete failed items where a live copy already exists
DELETE FROM site_work_items
WHERE status = 'failed' AND pipeline = 'build' AND item_key IS NOT NULL
  AND EXISTS (
    SELECT 1 FROM site_work_items live
    WHERE live.site_id = site_work_items.site_id
      AND live.item_key = site_work_items.item_key
      AND live.id != site_work_items.id
      AND live.status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
  );

-- Deduplicate within failed rows (keep newest)
DELETE FROM site_work_items
WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (
            PARTITION BY site_id, item_key ORDER BY created_at DESC
        ) as rn
        FROM site_work_items WHERE status = 'failed' AND pipeline = 'build' AND item_key IS NOT NULL
    ) ranked WHERE rn > 1
);

-- Reset remaining failed items
UPDATE site_work_items
SET status = 'triaged', attempt_count = 0, error = NULL,
    claimed_by = NULL, claimed_at = NULL
WHERE status = 'failed' AND pipeline = 'build';

COMMIT;
```

**Why the dedup index causes problems:** `idx_swi_dedup` is a partial unique index on `(site_id, item_key)` that excludes terminal statuses (complete, failed, etc.). Audit sweeps create new items while old failed ones exist with the same key. Resetting the failed one collides with the live one.

---

## 9. Specific Failure Patterns

### Dispatch loop input_mapping path mismatch (most common systematic failure)

The `build-dispatch-loop` maps the work item's `spec` JSONB as a nested object:

```json
"input_mapping": {
    "spec": "current_item.spec",
    "site_id": "current_item.site_id",
    "domain": "input_data.domain",
    ...
}
```

Handlers receive `input_data.spec.component_id`, `input_data.spec.issue`, `input_data.spec.refresh_site_components` etc. But many handler workflows reference these fields at the top level: `input_data.component_id`, `input_data.issue`. The `QueryDatabaseAction` tries a fallback of `input_data.input_data.<field>` which also doesn't match.

**Affected agents and their broken paths:**

| Agent | Broken path | Should be |
|---|---|---|
| `tool-improver` | `input_data.component_id`, `input_data.issue` | `input_data.spec.component_id`, `input_data.spec.issue` |
| `tool-auditor` | `input_data.component_id` | `input_data.spec.component_id` |
| `rerender-pages` | `input_data.refresh_site_components` | `input_data.spec.refresh_site_components` |

**Fix options (pick one):**

Option A — Flatten in the dispatch loop's `input_mapping` (add optional fields):
```json
"component_id?": "current_item.spec.component_id",
"issue?": "current_item.spec.issue",
"refresh_site_components?": "current_item.spec.refresh_site_components"
```

Option B — Update each handler's workflow to reference `input_data.spec.<field>` instead of `input_data.<field>`.

Option A is preferable because it keeps handler workflows clean and follows the pattern already established for `page_name?` and `reviewed_brief?`. But it requires knowing all spec fields handlers might need. Option B is self-documenting per handler.

**Diagnosis query — find all items failing with this pattern:**

```sql
SELECT wi.item_type, wi.handler_agent, s.domain, wi.attempt_count,
       LEFT(wi.error, 150)
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.error LIKE '%resolved to nil%'
  AND wi.pipeline = 'build'
ORDER BY wi.handler_agent, s.domain;
```

### Rerender failing with "pages missing header/footer"

The `rerender-pages` agent checks `input_data.refresh_site_components` but the dispatch loop sends it at `input_data.spec.refresh_site_components`. This is a specific instance of the path mismatch above. Fix: apply the same input_mapping flattening or update the rerender agent's condition.

### Missing handler agents

Work items created with a `handler_agent` that has no `agent_definitions` row stay stuck. The dispatch loop spawns a pod for the agent type, but no image/config exists so the spawn fails. Items go to `blocked` or `failed` with no useful error.

**Find all missing handlers:**

```sql
SELECT DISTINCT wi.handler_agent, wi.status, COUNT(*) as item_count
FROM site_work_items wi
LEFT JOIN agent_definitions ad ON ad.type = wi.handler_agent AND ad.deleted_at IS NULL
WHERE ad.type IS NULL
  AND wi.status NOT IN ('complete', 'verified', 'wont_fix')
  AND wi.pipeline = 'build'
GROUP BY wi.handler_agent, wi.status
ORDER BY item_count DESC;
```

Known missing handlers as of this writing: `internal-linker`, `hitl-review`.

**Resolution:** Either create the agent definition or reclassify the items to an existing handler.

### Content rewrites failing across all sites simultaneously

Usually resource starvation. The content rewrite chain is the deepest: dispatch → page-build-handler → page-content-writer → research-agent. If zombie pods are consuming cluster resources, new pods can't start and the chain times out. Fix: kill stale jobs, apply idle timeouts.

### Audit items accumulating faster than dispatch can process

The improvement loop creates findings every sweep. If the dispatch loop can't keep up (items failing and being re-triaged, or handlers timing out), the backlog grows each cycle. Check if the dispatch concurrency group is stuck, and whether handler pods are completing or dying.

### Claimed items timing out repeatedly

Check the claim timeout interval vs the handler's actual processing time. Large-site rerenders (15+ pages) take 15-20 minutes. If claim timeout is 10 minutes, they'll never complete.

### `wont_fix` with "superseded" accumulating

This is correct behaviour. When the improvement loop detects the same issue again while an older item is stuck in `failed` or `unresolved`, it creates a new item. The old one gets marked `wont_fix` with reason `superseded by active duplicate`. These don't need intervention — they're the dedup system working. To clean the noise:

```sql
-- Count superseded items per site
SELECT s.domain, COUNT(*) as superseded_count
FROM site_work_items wi
JOIN sites s ON s.id = wi.site_id
WHERE wi.status = 'wont_fix' AND wi.error LIKE '%superseded%'
GROUP BY s.domain
ORDER BY superseded_count DESC;
```

### `needs_section_data` items going straight to `wont_fix`

These represent sections that need data the system can't fabricate: leadership team members, pricing tiers, case studies, contact details. The handler correctly refuses to invent this data. Resolution requires human input via the HITL review flow — provide the data through the API, then retry the item.

### `add_tool` constraint violations

`content_components` has a unique constraint that prevents duplicate tool entries. When the tool-generator or tool-deployer runs twice for the same tool (e.g. after a failed first attempt was reset), the insert collides. Clean up the failed items; if the tool already exists as a component, mark the work item complete manually:

```sql
-- Check if the tool already exists
SELECT id, function, display_name FROM content_components
WHERE function = '<tool-function>' AND is_active = true;

-- If it does, mark the work item complete
UPDATE site_work_items SET status = 'complete',
    result = '{"note": "tool already exists, manually resolved"}'::jsonb
WHERE id = '<work-item-id>';
```

### Kcat trigger message silently mis-routed — use here-strings, not heredocs

**Symptom:** Kafka trigger via kcat appears to succeed, orchestration completes cleanly, but `collected_data` shows a generic scheduled-task shape rather than the agent you targeted:

```
"agent_config": {
  "workflow": {
    "steps": {
      "complete": {
        "action": "complete_workflow",
        "description": "No-op — scheduled task pre_query already did the work"
      }
    },
    "start_step": "complete",
    "processing_mode": "task",
    "timeout_seconds": 10
  }
}
```

And `__raw_message__.input_data` is `null`. The orchestration_id your trigger generated is nowhere in the chassis logs. The target agent's action never fired.

**Cause:** Multi-line heredoc JSON in the kcat trigger script. Shell interpolation, quote escaping, or line-ending handling between `<<JSON` and `JSON` can mangle the JSON body before it reaches Kafka. The chassis receives something it doesn't recognise as a valid orchestration request for your agent type and falls through to a default scheduled-task handler that completes with no work done.

**Diagnostic — confirm the routing went sideways:**

```sql
SELECT jsonb_pretty(collected_data->'agent_config') as agent_config,
       jsonb_pretty(collected_data->'__raw_message__'->'input_data') as raw_input
FROM orchestration_states
WHERE correlation_id = '<your-correlation-id>'::uuid;
```

If `agent_config` shows the "No-op" scheduled-task shape and `raw_input` is `null`, the message body did not survive shell processing.

**Fix:** Use a shell here-string (`<<<'...'`) with single quotes and flat one-line JSON, never a multi-line heredoc.

Bad (heredoc — avoid for JSON payloads):

```bash
kcat -P ... -t system.agent.generic.requests <<JSON
{
  "action": "orchestrate",
  "config": {"agent_type": "my-agent"},
  "input_data": {"field": "value"}
}
JSON
```

Good (here-string — always use this):

```bash
kcat -P ... -t system.agent.generic.requests \
  <<<'{"action":"orchestrate","config":{"agent_type":"my-agent"},"input_data":{"field":"value"}}'
```

The single quotes prevent any shell interpolation inside the JSON body — no variable expansion, no command substitution, no glob expansion. The flat JSON has no newlines or indentation that can be mangled.

**If the input_data needs shell variables** (e.g. dynamically built paths), construct the JSON with `jq` first, then pass as a here-string:

```bash
PAYLOAD=$(jq -nc \
    --arg out "/tmp/training_exports/$(date +%Y%m%d).jsonl" \
    '{action:"orchestrate", config:{agent_type:"training-data-exporter"},
      input_data:{output_path:$out, max_rows:100}}')
kcat -P ... <<<"$PAYLOAD"
```

`jq -nc` produces flat one-line JSON with proper escaping. Double quotes around `"$PAYLOAD"` preserve it as a single argument.

---

## 10. Quick Health Dashboard Query

Single query to see system state at a glance:

```sql
SELECT 'work_items' as category, status, COUNT(*) as count
FROM site_work_items WHERE pipeline = 'build'
GROUP BY status
UNION ALL
SELECT 'orchestrations', status, COUNT(*)
FROM orchestration_states
WHERE updated_at > NOW() - INTERVAL '1 hour'
GROUP BY status
UNION ALL
SELECT 'scheduled_tasks',
       CASE WHEN last_triggered_at IS NOT NULL
            AND (last_completed_at IS NULL OR last_completed_at < last_triggered_at)
            AND last_triggered_at + (timeout_seconds || ' seconds')::interval > NOW()
       THEN 'in_flight' ELSE 'idle' END,
       COUNT(*)
FROM scheduled_tasks WHERE enabled = true
GROUP BY 2
ORDER BY category, status;
```
# hunting for logs
Like when looking for SavePageSectionsAction logs that weren't appearing in any logs, look for logs before and after it. e.g.
page-build-handler workflow:

ensure_site_record     →  "EnsureSiteRecordAction: ..." (persistent pod usually)
load_page_record       →  "LoadPageRecordAction: ..." (new action; likely "Starting" / "Complete")
load_existing_content  →  "LoadExistingContent: ..."
load_spec_sections     →  "LoadPageSectionsFromSpecAction: ..."
plan_sections          →  "plan_sections" / "PlanSectionsAction: ..."
check_has_ready_sections → conditional, minimal log
spawn_content_writer   →  "Spawning agent" or "spawn_agent"
call_content_writer    →  "call_agent" lines, then await response
check_content_produced → conditional
validate_content       →  "ValidatePageContentAction: complete"  ← we know this log exists
save_sections          →  "SavePageSectionsAction: Starting" ...  ← what we want
update_status          →  "UpdatePageStatusAction: ..."
spawn_rerender_agent   →  "Spawning agent"
deploy_page            →  "call_agent" → "git_commit" response
complete               →  workflow complete
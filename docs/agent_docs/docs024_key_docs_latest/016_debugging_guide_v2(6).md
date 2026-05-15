# 016 — Debugging Guide

Practical steps for diagnosing and fixing problems in the pipeline. Based on real failure patterns.

**Schema reminder:** The `site_work_items` column for work category is `pipeline` (not `domain`). The site's website domain comes from the `sites` table or `input_data.domain`. See 007_adoption_pipeline schema notes for other renamed columns.

---

## 0. Before You Change Anything — Assumption Checklist

This is process discipline, not a query reference. Most defects in recent sessions did not come from misunderstanding the code — they came from acting on unverified assumptions about it. Before writing a migration, a Go patch, or a trigger script, walk these checks explicitly. Each takes seconds; each is a known repeat-failure category.

**1. Field-naming conventions are per-action, not universal.** The `_field` suffix that lets `store_asset` resolve `asset_key_field: "input_data.spec.asset_key"` does NOT extend to arbitrary keys. Only `asset_key_field`, `site_id_field`, `data_field`, `origin_prompt_field` are attested. Inventing `purpose_field`, `domain_field`, etc. fails silently — the action writes a partial DB row and the downstream step gets a malformed `output_field`. Before assuming a config key works, `grep` the action source for that exact identifier.

**2. `input_mapping` fields are required by default.** Listing a field with no suffix means "this MUST exist in the source data or the workflow fails at extraction time." Use `field?: "path"` (question mark on the destination key) to mark optional. Items emitted from a discovery check whose source row had null columns (e.g. `style_hints`, `constraints`) WILL be missing those spec keys, and a required input_mapping for them will fail every time.

**3. Empty logs do not mean the action didn't run.** When grepping for `SomeActionName` and finding nothing, the action may have never been reached because an earlier step in the workflow failed. Query `orchestration_states.error_preview` FIRST — it shows the actual failure point and the truncated error. Only then grep logs around that step.

**4. Database rows can be partial.** `store_asset` and similar actions write the DB row early and only fail later when emitting the output to `collected_data`. Seeing a row appear in `assets` is NOT proof that the workflow succeeded — the row may have empty `purpose` or a wrong `asset_key` from a config that failed mid-execution. Always cross-check against `orchestration_states.status = COMPLETED`.

**5. SQL is immediate; Go is not.** Migrations apply on COMMIT. Go changes require chassis rebuild AND pod rollout AND, sometimes, image tag bump on `agent_definitions`. After applying a Go change, verify the new behaviour is actually live (look for a new log line you added) before debugging anything else. Several diagnostic sessions have been wasted on "why isn't this working" when the answer was "the new code isn't running."

**6. Sibling functions in the same file are the canonical pattern.** When adding a new function that walks JSON, look for existing walkers in the same file first. If `extractPagesFromPlan` and `flattenSiteScopeDirectives` use `findDirectiveTree` for wrapper-tolerant lookup, your new walker probably should too. Direct `data[key]` access fails as soon as the input grows a wrapper (e.g. `validate_*` inserting `result.` ahead).

**7. Token budgets scale with structured output.** Adding a new required output structure to an LLM prompt (e.g. an `imagery` block with 15 entries on a multi-page site) can blow the existing `max_tokens` cap. Symptom: `validate_*_action` fails with `unexpected end of JSON input`. Before adding structure to a prompt, estimate the output token count and verify it fits.

**8. Every shell variable referenced must be declared.** A trigger script that uses `$MESSAGE_ID` or `$CLIENT_ID` without setting them produces `-H key=` with empty values, which the chassis silently rejects. Put `set -u` at the top of every trigger script. After building any payload, `echo` it before sending.

**9. Intermediate jq state can silently become null.** `jq --slurpfile var file` with a missing/empty file makes `$var[0]` evaluate to `null` without error. Trigger payload becomes `spec: null`. Use `--argjson var "$JSON_STRING"` with a pre-validated shell variable instead. Always `cat | jq .` the constructed payload before sending it.

**10. Manual triggers bypass dispatch.** When work items sit in `detected` and don't move, distinguish "dispatch loop isn't claiming them" from "the handler is broken." Trigger image-build-handler manually via kcat — if the orchestration runs and completes, dispatch is the problem. If it fails, the handler is the problem. The two need different fixes.

**11. Parent and child orchestrations are separate rows.** A `spawn_*` step creates a child orchestration. The child can COMPLETE successfully (asset created in S3, DB row written) while the parent then FAILS on the next step. Always query both: `WHERE orchestration_id = ...` for the parent and `WHERE parent_orchestration_id = ...` for the child.

**12. The `?` suffix matters where it goes.** It belongs on the destination field name in `input_mapping`, not on the source path: `"constraints?": "input_data.spec.constraints"` — not `"constraints": "input_data.spec.constraints?"`. Wrong placement is silently ignored and the field stays required.

**13. Don't refactor on speculation when one fresh failure would name the cause.** A 10-minute timeout in the chassis's response-await path can have multiple causes: `Failed to create ExecutionContext from headers`, `no request ID in headers`, `Awaited request not found`, or `ClaimAwaitedRequest: not claimed`. Each maps to a *different* fix — header-shape, body-shape, table-state, or race-condition respectively. Inferring the cause from architectural reasoning ("response went to the wrong topic") wastes a round of code+deploy+test if the log would have said the cause directly. Before writing any patch to a response-routing or matcher path, refire the failing case with a known timestamp and grep `agent_type=generic` chassis pods for the response-handling log lines. If the response was consumed (`Response consumer received message`), the problem is downstream of receipt; if not, it's upstream. That single fact eliminates half the hypotheses.

**14. Pod rotation eats your logs.** kubectl logs only sees what's in the current pod's stdout buffer. A pod that rotated 1 hour ago has its logs gone (unless you have central log shipping). When debugging a failure that happened more than ~30 minutes ago, your three options are: (a) refire it now and capture live, (b) check `agent_error_log` table if the failure was severe enough to land there, (c) accept that the historic logs are gone and reason from DB state alone (`orchestration_states.error`, `awaited_requests.status`+timestamps, `agent_error_log.context`). Don't chase historic logs that aren't there; refire when the cost is acceptable. Especially important after a chassis deploy — every rollout cycles every pod's logs at once.

When in doubt, write the assumption down and verify it before writing the code. A 30-second grep saves a 30-minute round trip.

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

### Orchestration hung at `*_spawn_handler` step (timeout_at not enforced)

Observed pattern: a `build-dispatch-loop` orchestration sits at `process_item_iter_N_spawn_handler` with `status = 'AWAITING_RESPONSES'` for tens of minutes. The `awaited_requests` JSON shows a single entry with a `timeout_at` value that was 3 minutes after `sent_at` — and that deadline has long since passed without the orchestration noticing. The corresponding `site_work_items` row stays in `claimed` until `claimed-item-timeout` or `stale-orchestration-reaper` papers over it.

Confirmed examples: the gaswholesalers redeploy (May 2026) had 28 `page_rerender` items sit unclaimed for 6-8 days, then a single robot-hands.com `selection-guide` orchestration on May 12 hung for 30+ minutes at iter_3_spawn_handler waiting on a `page-build-handler` response.

Diagnostic:

```sql
-- All orchestrations stuck at a spawn_handler step with an expired timeout
SELECT orchestration_id, owner_agent_type, current_step,
       last_activity, NOW() - last_activity AS idle_for,
       jsonb_path_query(awaited_requests, '$.*.target_agent_type') AS target,
       jsonb_path_query(awaited_requests, '$.*.timeout_at')        AS timeout_at,
       jsonb_path_query(awaited_requests, '$.*.responses_topic')   AS responses_topic
FROM orchestration_states
WHERE status = 'AWAITING_RESPONSES'
  AND current_step LIKE '%spawn_handler%'
  AND last_activity < NOW() - INTERVAL '5 minutes'
ORDER BY last_activity ASC;
```

Possible causes, each needs a different fix:

- **Orchestration tick-loop ignores `timeout_at`.** The engine reads `awaited_requests` looking for matched responses but doesn't check whether any entry's deadline has passed. So unless a response physically arrives, the orchestration never advances. This is the most likely cause given the pattern is uniform.
- **The handler agent died mid-request.** The target agent pod was killed before responding. Look for restarts via `kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[0].restartCount}{"\n"}{end}'`.
- **The ephemeral response topic was never subscribed.** Spawn-handler patterns generate a `responses_topic` per request (e.g. `job.{correlation}-{orch}-build-dispatch-loop-spawn_dispatch.responses`). If the orchestration-engine's response-router only listens to a fixed set of topics, responses on the ephemeral one are dropped silently.
- **The handler responded to a different topic or with a mismatched `in_response_to_request_id`.** The router would never match it back to the orchestration. Check the handler's outgoing logs around the `timeout_at` window.

The `stale-orchestration-reaper` and `claimed-item-timeout` scheduled tasks paper over the symptom (failing the orchestration after 30 minutes, resetting the work item) but don't fix the underlying timeout-enforcement gap. Until the engine honours `timeout_at`, every spawn-handler step has up to a 30-minute tail risk of stalling the per-item path.

### `claimed-item-timeout` evidence check produces false-positive completions

Related to the spawn-handler-hang above. When a claimed work item's response is lost, `claimed-item-timeout`'s pre_query has a "verified done despite lost response" branch that auto-completes items if there's evidence the work was done. The current evidence checks are too loose and fire wrongly when other unrelated work happens on the same site within the claim window.

Specifically, for `page_rerender` and `needs_rerender` items, the check is:

```sql
EXISTS (
    SELECT 1 FROM pages p
    WHERE p.site_id = wi.site_id
      AND p.build_status = 'deployed'
      AND p.updated_at > wi.claimed_at
)
```

Problems:
- It checks **any page on the site**, not the specific page targeted by the work item.
- It uses `pages.updated_at`, which is bumped on every UPDATE for any reason — not `pages.deployed_at`, which is the actual deploy signal.

Result: in any window where multi-page rerenders are running (which is most active sites most of the time), a stuck item on page X gets auto-completed because page Y was successfully rerendered. The work-item record says "complete" but the targeted page may never have been touched.

Confirmed instance: gaswholesalers `fuel-industry-insights` rerender on 2026-05-12 — claimed at 19:28, auto-completed at 19:43 with error `Auto-completed: work verified done despite lost response`, but the actual git commit for that page didn't happen until 20:30 (47 minutes after the auto-complete) via a separate code path. The work-item record is permanently inconsistent with deployment reality.

The `needs_content_page` branch in the same pre_query uses `p.name = wi.spec->>'page_name'` so it IS per-page on the page name. But it still uses `updated_at` rather than `deployed_at`, so it can still false-positive when a page row is touched without an actual deploy.

The `needs_design` branch checks only `site_components` slot `head` — too narrow if the design change was to footer or header, and uses `updated_at` rather than a deploy-specific timestamp.

Diagnostic — find recently auto-completed items where the target page hasn't actually been deployed since the claim:

```sql
SELECT wi.id,
       wi.item_type,
       wi.spec->>'page_name' AS page,
       wi.claimed_at,
       wi.completed_at,
       p.deployed_at AS page_deployed_at,
       p.updated_at  AS page_updated_at,
       (p.deployed_at > wi.claimed_at) AS deploy_after_claim,
       wi.error
FROM site_work_items wi
LEFT JOIN pages p ON p.id = wi.page_id
WHERE wi.status = 'complete'
  AND wi.error = 'Auto-completed: work verified done despite lost response'
  AND wi.completed_at > NOW() - INTERVAL '7 days'
ORDER BY wi.completed_at DESC;
```

Any row where `deploy_after_claim = false` is a false positive.

Item-type-specific guidance for the eventual fix:

| Item type | `page_id` populated? | Correct evidence check |
|---|---|---|
| `page_rerender` | Always | `p.id = wi.page_id AND p.deployed_at > wi.claimed_at` |
| `needs_content_page` | Always | `p.id = wi.page_id AND p.deployed_at > wi.claimed_at` |
| `needs_rerender` | NULL on most rows (site-level orchestrator, fans out) | Don't auto-complete via this path — let it fall through to `reset` and retry |
| `needs_design` | NULL | Needs a `site_components.deployed_at` column, which doesn't exist today, or a different evidence mechanism — leaving as-is keeps the false positives narrow at least |

Until the pre_query is fixed, treat any work-item `result` JSON that is empty/missing alongside `error = 'Auto-completed: ...'` as **untrusted** — the item may not have actually completed. The reaper-fix is a small SQL-only change to the scheduled task's pre_query, no Go side.

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

### LLM output truncated, `validate_*` fails with "unexpected end of JSON input"

The LLM call succeeded HTTP-wise but the response was cut off at `max_tokens`. The downstream validator can't parse a half-written JSON object and the orchestration FAILS with a preview that ends mid-string. Common when adding new required structure to a prompt without raising the cap.

**Diagnosis:**

```sql
SELECT orchestration_id, status, current_step, LEFT(COALESCE(error, ''), 300) AS err
FROM orchestration_states
WHERE site_id = '<site>' AND status = 'FAILED'
ORDER BY created_at DESC LIMIT 5;
```

Look for `unexpected end of JSON input` with a preview ending mid-value. The `current_step` will be `validate_plan` or similar.

**Fix:** Raise `max_tokens` on the LLM step's `ai_service.max_tokens` config. Output token cost is metered per used, not per cap — generous headroom is free.

```sql
UPDATE agent_definitions
   SET default_config = jsonb_set(
       default_config,
       '{workflow,steps,<step_name>,config,ai_service,max_tokens}',
       to_jsonb(8000),
       false
   )
 WHERE type = '<agent>' AND is_active = true;
```


### `operator does not exist: jsonb && jsonb` — silent in CSS path, hard fail in JS

**Symptom:** Workflow step fails with
`ERROR: operator does not exist: jsonb && jsonb (SQLSTATE 42883)`.
First seen on `render_js_snippets_for_site` when site-asset-renderer
ran for the first time on gaswholesalers in May 2026.

**Root cause:** Postgres's `&&` (array overlap) operator only exists
for native Postgres arrays — `text[]`, `int[]`, etc. There is no
`jsonb && jsonb` operator. The `loadComponentCSSSnippets` function
in `render_css_from_spec_action.go` has used `applies_to && $1::jsonb`
since the css_snippets table was added in jsonb form, so the query
has been **silently failing the entire time**. The function's error
handler is `logger.Warn(...); return ""`, so the CSS pipeline
degraded gracefully (theme + section styles still rendered) and
nobody noticed that no css_snippet has ever actually been included
in any deployed `styles.css`.

The JS analog (`loadJSSnippetsForSite` in
`render_js_snippets_for_site_action.go`) treats snippets as its
*entire* output, so the same bug surfaces as a hard workflow
failure instead of silent degradation.

**Diagnosis:**

```sql
-- Confirm the operator doesn't exist
SELECT 'a'::jsonb && 'b'::jsonb;
-- ERROR:  operator does not exist: jsonb && jsonb

-- The working pattern
SELECT EXISTS (
  SELECT 1
  FROM jsonb_array_elements_text('["a","b"]'::jsonb) AS x(elem)
  WHERE x.elem IN (SELECT jsonb_array_elements_text('["b","c"]'::jsonb))
);
-- t

-- For CSS snippets specifically — confirm the silent failure: pick
-- any deployed site and check its styles.css for one of the rule
-- names from a css_snippet row. If none of the rule names are in
-- the file, css_snippets has never reached that site:

SELECT name, css_content FROM css_snippets
WHERE applies_to::text LIKE '%latest-news%';
-- shows snippet rows that should have been included

-- Then read the actual deployed file and grep for unique selectors
-- from those snippets — e.g. `.latest-news-section .news-card-meta`.
-- Absence confirms the silent failure.
```

**Fix pattern:** Replace `applies_to && $1::jsonb` with EXISTS +
`jsonb_array_elements_text`:

```sql
WHERE EXISTS (
  SELECT 1
  FROM jsonb_array_elements_text(applies_to) AS a(elem)
  WHERE a.elem IN (SELECT jsonb_array_elements_text($1::jsonb))
)
```

Pure jsonb on both sides, no driver-side array conversion needed,
no `pq.Array` dependency (the chassis explicitly avoids `lib/pq` —
see comments around line 90168 in the chassis source).

**Where this needs fixing now (May 2026):**

- `platform/orchestration/actions/render_css_from_spec_action.go` —
  function `loadComponentCSSSnippets` (silent failure path)
- `platform/orchestration/actions/render_js_snippets_for_site_action.go` —
  function `loadJSSnippetsForSite` (hard failure path — fixed in the
  same change set)

**Where this might exist elsewhere — audit pattern:**

```bash
# Find any other jsonb && usage in the codebase
grep -rn "&& \$.*::jsonb\|applies_to &&" platform/
grep -rn ":jsonb && \|jsonb && \"" platform/
```

Any match outside test files should be reviewed: if the operator is
between a jsonb column and a jsonb parameter, it's broken; if it's
on a converted array, it's fine.

**Why this stayed hidden for months:** silent-failure error handlers
(`logger.Warn(...); return ""` and similar) plus a graceful
downstream consumer (the CSS theme builder doesn't care if snippets
are empty). When writing similar loaders, prefer surfacing the error
to the caller and letting the orchestration step fail visibly. Hard
failure beats silent degradation when the data is supposed to be
there.

---

### New JSON walker silently returns nothing (canonical resolver bypassed)

A new function reads `data["foo"]` at top level and finds nothing, while sibling functions in the same file use a multi-wrapper resolver like `findDirectiveTree(data, "foo")` and find the same data under `data["site_plan"]["foo"]` or `data["llm_plan"]["result"]["foo"]`.

**Symptom:** Action runs, log line says "no foo block found; skipping", downstream consumers get nothing. But the data is in `collected_data` — just one level deeper than the walker assumes.

**Diagnosis:** Query the actual `collected_data` for the orchestration:

```sql
SELECT
    ARRAY(SELECT jsonb_object_keys(collected_data->'llm_plan'->'result')) AS llm_keys,
    ARRAY(SELECT jsonb_object_keys(collected_data->'site_plan'))          AS site_plan_keys
FROM orchestration_states
WHERE orchestration_id = '<id>';
```

If your target key appears in either `site_plan_keys` or `llm_keys` (but not at the top level) and the walker reports "not found", the walker is looking at the wrong level.

**Fix:** Use the canonical resolver. Read the file and find one. If none exists, write one that mirrors `extractPagesFromPlan`'s pattern.

### `store_asset` writes empty-purpose row when config is invalid

`store_asset` writes the DB row before validating all its config keys. If a config key isn't recognised (e.g. `purpose_field` — which doesn't exist; only `purpose` literal and `asset_key_field`/`site_id_field`/`data_field`/`origin_prompt_field` are attested), the action writes the asset row with the unresolvable field empty, then fails to populate its `output_field`. The next step (typically `call_asset_deployer`) then can't find `asset_stored.image_uri`.

**Diagnosis:**

```sql
-- Recent assets with null purpose or empty purpose
SELECT id, asset_key, purpose, origin_model, created_at
FROM assets
WHERE site_id = '<site>'
  AND (purpose IS NULL OR purpose = '')
  AND created_at > now() - interval '1 hour';

-- Orchestration error
SELECT current_step, LEFT(error, 300) FROM orchestration_states
WHERE orchestration_id = '<id>';
```

If you see a recent asset row with null/empty purpose AND the orchestration FAILED at `call_asset_deployer` with `asset_stored.image_uri not found`, the store step's config is using an unsupported `*_field` key.

**Fix:** Replace the unsupported `*_field` key with a literal value, or with a hardcoded purpose for one path and branch by kind for the others. Mirror the existing `store_variant_asset` config — that's the working reference.

**Cleanup:** Delete the partial asset row before retrying — UPSERT will not overwrite it because the constraint matches:

```sql
DELETE FROM assets WHERE id = '<partial-asset-id>';
UPDATE site_work_items SET status = 'detected', error = NULL, claimed_at = NULL
WHERE id = '<work-item-id>';
```

### `input_mapping failed: source path 'input_data.spec.<field>' not found`

The workflow step's `input_mapping` lists `<field>` as required, but the work item's spec doesn't contain that key. Most commonly: a discovery check emits a field only when the source DB column was non-null, and the workflow's input_mapping doesn't allow it to be absent.

**Diagnosis:** Read the work item's spec to confirm the field is genuinely absent:

```sql
SELECT id, jsonb_pretty(spec::jsonb)
FROM site_work_items WHERE id = '<work-item-id>';
```

**Fix:** Mark the destination field optional with a `?` suffix:

```sql
UPDATE agent_definitions
   SET default_config = jsonb_set(
       default_config,
       '{workflow,steps,<step>,config,input_mapping}',
       (
           (default_config #> '{workflow,steps,<step>,config,input_mapping}')
               - 'old_required_key'
       ) || jsonb_build_object('old_required_key?', 'input_data.spec.<field>'),
       false
   )
 WHERE type = '<agent>' AND is_active = true;
```

The `?` goes on the destination key name, not on the source path.

### kcat trigger doesn't produce an orchestration row

Most likely cause: a header value in the kcat invocation referenced an unset shell variable, producing `-H key=` with an empty string. The chassis rejects messages with empty required headers silently — no log, no orchestration row, no error visible to you.

**Diagnosis:**

```bash
# Check the trigger script for header refs that may be unset
grep -E '^\s*-H \w+=\$' your-trigger.sh
```

Common culprits in simplified trigger scripts: `$MESSAGE_ID`, `$CLIENT_ID`, `$REQUEST_ID`. The original orchestration scripts declare these near the top; simplified copies often drop the declarations.

**Fix:** Put `set -u` at the top of every trigger script — it makes any unset variable an immediate error rather than an empty header. Also echo the payload before sending:

```bash
#!/bin/bash
set -u  # fail on unset variables
set -o pipefail

# all required UUIDs and IDs declared explicitly
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# verify before sending
echo "About to send with:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo "  CLIENT_ID=$CLIENT_ID"

# build payload as a variable, cat | jq . before send
PAYLOAD='{"action":"orchestrate", ...}'
echo "$PAYLOAD" | jq .  # fails loudly if malformed

# only then send
echo "$PAYLOAD" | kubectl -n kafka run -i --rm kcat-... ...
```

### Trigger sent `spec: null` despite the work item having a real spec

The `jq --slurpfile` builder silently produces null when the source file is missing, empty, or contains literal `null`. Trigger lands with `spec: null` and the workflow fails at the first step that reads from spec.

**Fix:** Skip slurpfile. Pull the spec into a shell variable, validate it, then pass via `--argjson`:

```bash
SPEC=$(psql -h $PGHOST -U $PGUSER -d clients_db -t -A -c \
  "SELECT spec::text FROM site_work_items WHERE id = '$WORK_ITEM_ID'")

if [ -z "$SPEC" ] || [ "$SPEC" = "null" ]; then
    echo "ERROR: spec is empty/null for $WORK_ITEM_ID"
    exit 1
fi

jq -nc \
  --arg site_id "$SITE_ID" \
  --argjson spec "$SPEC" \
  '{ action: "orchestrate", input_data: { site_id: $site_id, spec: $spec } }' \
  > /tmp/trigger.json

# Guard the final payload before sending
if ! grep -q '"prompt"' /tmp/trigger.json; then
    echo "ERROR: trigger payload missing spec.prompt"
    exit 1
fi
```

### Parent orchestration FAILED but a child orchestration COMPLETED (and may have done useful work)

When a workflow includes `spawn_agent`, the spawned agent runs as a separate orchestration with its own row in `orchestration_states`. The child can complete successfully (image generated, asset written, file uploaded to S3) while the parent then fails at the very next step (e.g., reading the child's response into a downstream mapping).

**Diagnosis — find the full tree:**

```sql
-- Parent
SELECT orchestration_id, status, current_step, LEFT(error, 200) AS err
FROM orchestration_states
WHERE orchestration_id = '<parent>';

-- Children of that parent
SELECT orchestration_id, owner_agent_type, status, current_step
FROM orchestration_states
WHERE parent_orchestration_id = '<parent>'
ORDER BY created_at;

-- Anything created around the same time on the same site
SELECT orchestration_id, owner_agent_type, status, current_step, created_at
FROM orchestration_states
WHERE site_id = '<site>'
  AND created_at BETWEEN '<parent-start>' AND '<parent-fail-time>'
ORDER BY created_at;
```

**Common downstream artefact:** an asset row in `assets`, or a row in another table, written by the child but disconnected from the parent's workflow state. The cleanup step is to delete the orphan, fix the parent's workflow, retry the work item.

---

## Work item lifecycle and the `detected → triaged → claimed` state machine

Most "work item stuck" symptoms map onto a single underlying question: which agent owns the next transition for this item, and has that agent run? The states are valid intermediate stops — not all of them are bugs.

### The state machine

```
discovery emits  →  detected
                       ↓
                    (design-audit-agent runs visual + content
                     auditors, then calls triage_detected_items)
                       ↓
                    triaged
                       ↓
                    (build-dispatch-loop claims; partial indexes
                     idx_swi_handler and idx_swi_site_pending
                     filter for this status)
                       ↓
                    claimed
                       ↓
                    (handler runs; the mark_work_item_complete
                     step at the end of image-build-handler
                     and similar transitions to complete; the
                     mark_work_item_failed step transitions to failed
                     on error paths)
                       ↓
                    complete  /  failed
```

Other terminal states reachable from elsewhere in the lifecycle: `wont_fix`, `verified`, `rejected`, `unresolved`, `needs_human_review`, `blocked`.

### Who owns each transition

| Transition | Owner | Mechanism |
|---|---|---|
| insert at `detected` | Discovery check (anything in `platform/orchestration/actions/discovery_checks/`) | INSERT in the check's emit logic |
| `detected` → `triaged` | `design-audit-agent` | Calls `triage_detected_items` action at end of its workflow, after the visual and content auditors run |
| insert at `triaged` | Admin-created items (bypass discovery) | `site_admin_handlers.go:455` HTTP POST |
| `triaged` → `claimed` | `build-dispatch-loop` (running every ~60s) | `claim_work_item` action |
| `claimed` → `complete` | Handler agent (image-build-handler, page-build-handler, etc.) | `mark_work_item_complete` step at the end of the handler's workflow |
| `claimed` → `failed` | Handler agent on error | `mark_work_item_failed` step on the error path |
| any → `wont_fix` | Audit reconciler, when the item becomes irrelevant | `closeResolvedDataRequest` and similar |
| any → `needs_human_review` | Handler after reaching `max_attempts` | Per-item-type logic |

### Symptom → cause table

| Symptom | Most likely cause | Fix |
|---|---|---|
| Many items sit in `detected` for site X | Discovery ran but `design-audit-agent` hasn't run since | Trigger `design-audit-agent` for that site |
| Items in `triaged` for hours, never claimed | Dispatch loop is off, OR higher-priority items are crowding the queue | Confirm dispatch is running; investigate priority logic (see open items in `FOCUS_dispatch_diagnostic.md`) |
| Item in `claimed` indefinitely | Handler crashed mid-execution, or chassis died | Check `orchestration_states.error_preview`; reset to `triaged` to retry |
| Item in `failed` with `attempt_count = max_attempts` | Handler genuinely cannot process this item | Investigate the specific failure; manual intervention |

### Operator commands

Bulk-promote detected items to triaged for a specific type (operator override — normally this is done by audit):

```sql
UPDATE site_work_items
   SET status = 'triaged',
       triaged_at = now()
 WHERE site_id = '<site>'
   AND item_type = '<type>'
   AND status = 'detected';
```

Reset a stuck `claimed` item back to `triaged` for another dispatch attempt:

```sql
UPDATE site_work_items
   SET status = 'triaged',
       claimed_at = NULL,
       claimed_by = NULL
 WHERE id = '<item-id>';
```

Reset all `failed` items below their max attempts back to `triaged`:

```sql
UPDATE site_work_items
   SET status = 'triaged',
       claimed_at = NULL,
       claimed_by = NULL,
       error = NULL
 WHERE site_id = '<site>'
   AND status = 'failed'
   AND attempt_count < max_attempts;
```

### The recurring debugging trap — inferring writers from readers

The first hypothesis when items pile up in `detected` is tempting: "the writer of the next state must be missing." That feels especially compelling when you have direct evidence (via partial-index definitions or query filters) of what state dispatch reads from. Both of these queries make `triaged`/`approved` the obvious target of dispatch's claim:

```
idx_swi_handler        WHERE status = ANY (ARRAY['triaged','approved'])
idx_swi_site_pending   WHERE status = ANY (ARRAY['triaged','approved'])
```

But that's evidence of the read path only. It doesn't tell you who writes to `triaged`. In this codebase the writer is `triage_detected_items`, owned by `design-audit-agent`, registered in `registry.go:722`. A 30-second grep for the verb `triage` surfaces both the action and the workflow that calls it. The mistake on 2026-05-14 was concluding "missing transition" without doing that grep.

Generalising: **before changing or adding code to fix a missing transition, grep the full codebase for the verb that performs it.** Search the action registry. Search agent workflows. Search recent session transcripts. The transition probably exists somewhere you haven't looked — owned by an agent you didn't expect to be involved.

This is a specific case of assumption-checklist item 6 ("sibling functions in the same file are the canonical pattern"), but applied across files rather than within them. When you find yourself reasoning about what must be missing, treat that as a flag to first confirm what's actually present.

The earlier version of this doc made the classic mistake of inferring system behaviour from one source (the partial-index definitions) without searching for upstream writers. Indexes told me where dispatch reads from; they didn't tell me where items get written to that state. A 30-second grep for `triage` would have surfaced both the registry entry and the design-audit-agent workflow that calls it.

### Cross-references

- `FOCUS_dispatch_diagnostic.md` — full evidence trail and three open architectural questions (auto-triage emissions, scheduled audit runs, dispatch priority across item types)
- `registry.go:722` — `triage_detected_items` action registration (canonical evidence of the transition)
- `site_admin_handlers.go:455, :749` — admin-driven `triaged` state creation and explicit transitions via the dashboard
- `phase_2g_followup_mark_work_item_complete.sql` and `phase_2g_followup_mark_work_item_failed.sql` — handler-side completion bookkeeping

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

**Before grepping logs, query `orchestration_states` first.** When an expected log line is missing, the most common cause is not "the action ran but didn't log" — it's "the workflow died upstream and the action never ran." `orchestration_states.error_preview` and `current_step` tell you exactly where the workflow stopped. The action you're hunting for may simply be downstream of the failure point.

```sql
SELECT orchestration_id, status, current_step, created_at, updated_at,
       LEFT(COALESCE(error, ''), 300) AS err
FROM orchestration_states
WHERE site_id = '<site>'
  AND created_at > now() - interval '1 hour'
ORDER BY created_at DESC;
```

If `status = FAILED` and `current_step` is earlier than the action you're hunting, that's your answer — no log will exist because no execution occurred. Fix the upstream step first, retry, then grep.

If `status = COMPLETED` and you're missing the log, then it's a real log gap — chassis log levels, log shipper drop, or wrong selector. THEN start grepping.

Once you've confirmed the workflow reached your action: look for logs before and after it. e.g.
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
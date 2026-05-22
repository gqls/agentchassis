# 016 — Debugging Guide

Practical steps for diagnosing and fixing problems in the pipeline. Based on real failure patterns.

**Schema reminders:**
- The `site_work_items` column for work category is `pipeline` (not `domain`). The site's website domain comes from the `sites` table or `input_data.domain`.
- The error column on `site_work_items` is `error` (not `error_message`). The `error_message` name is used on `orchestration_states`, `agent_error_log`, and `llm_call_log` — but not here. Get this wrong and PostgreSQL returns `column "error_message" does not exist`.
- See `007_adoption_pipeline` schema notes for other renamed columns.

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

**13. Don't guess at column names from memory.** Schemas drift — columns get renamed (`domain` → `pipeline` was observed in this codebase on 2026-05-15), added, dropped. Before writing any SQL that references column names, run `\d <table_name>` to confirm the current schema. The cost of a `\d` is 2 seconds; the cost of a wrong-column-name query is rerunning the whole investigation. This applies especially to: nullable columns (the partial-index trap), parent/child relationships (does this table have `site_id` or do you need to JOIN to a parent?), and recently-migrated columns. The audit triage action's stale `target_domain` config keyword (post-rename) is one example of why memory of "what this column was called" is unreliable.

**14. Don't refactor on speculation when one fresh failure would name the cause.** A 10-minute timeout in the chassis's response-await path can have multiple causes: `Failed to create ExecutionContext from headers`, `no request ID in headers`, `Awaited request not found`, or `ClaimAwaitedRequest: not claimed`. Each maps to a *different* fix — header-shape, body-shape, table-state, or race-condition respectively. Inferring the cause from architectural reasoning ("response went to the wrong topic") wastes a round of code+deploy+test if the log would have said the cause directly. Before writing any patch to a response-routing or matcher path, refire the failing case with a known timestamp and grep `agent_type=generic` chassis pods for the response-handling log lines. If the response was consumed (`Response consumer received message`), the problem is downstream of receipt; if not, it's upstream. That single fact eliminates half the hypotheses. Corollary for external APIs: a `400` is undebuggable without seeing the request body you actually sent — add request/response body logging at the HTTP layer before guessing which field is wrong.

**15. Pod rotation eats your logs.** kubectl logs only sees what's in the current pod's stdout buffer. A pod that rotated 1 hour ago has its logs gone (unless you have central log shipping). When debugging a failure that happened more than ~30 minutes ago, your three options are: (a) refire it now and capture live, (b) check `agent_error_log` if the failure was severe enough to land there, (c) accept the historic logs are gone and reason from DB state alone (`orchestration_states.error_preview`, `awaited_requests.status`+timestamps, `agent_error_log.context`). Don't chase historic logs that aren't there; refire when the cost is acceptable. Especially important after a chassis deploy — every rollout cycles every pod's logs at once.

**16. Don't change a value that earlier evidence already proved correct.** When fixing problem A, do not "tidy up" an adjacent value B that has no failing test pointing at it — especially if B was working in an earlier run. Concrete instance: the thunder-adapter defaulted `gpu` to lowercase `"a100"`, which was correct. While fixing the *field names* (`gpu` → `gpu_type`), a `strings.ToUpper` was added on the same line based on an OpenAPI *example* showing `"H100"`, changing a value nothing had complained about. Result: `400 invalid GPU type: A100` — a self-inflicted failure on a value that had been fine. The discipline: a fix should change exactly what the evidence indicts and nothing else. If you find yourself "improving" a neighbouring value, stop — that's a separate change needing its own justification and its own test. When a diff alters something that a previous test exercised successfully, treat that as a red flag and re-confirm the old value was actually wrong before shipping. (See the Thunder API `400` pattern in §9 for the full sequence.)

**17. A deployed code change does not mean the migration ran.** `make build` + `docker push` + `kubectl set image` ships *Go code*. SQL migrations are a *separate* step (`psql`/`\i` against the target DB) — nothing in the build/deploy runs them. This gap has bitten this project twice: migration 029 (gpu-provisioner workflow) and the thunder partial-unique-index migration both sat unapplied while we assumed the deploy had carried them. The tell is unmistakable: **the error names a constraint/column/index that the latest migration was supposed to change** (e.g. `duplicate key ... thunder_instance_id_key` when that constraint was meant to be dropped). Before re-debugging the *code* for such an error, run `\d <table>` and confirm the migration actually applied. A 2-second `\d` versus a wasted deploy cycle. If a row's behaviour contradicts a migration you "know" ran, distrust the assumption, not the code.

**18. Adding a method to a shared interface breaks every implementer and every type-assertion against it — rebuild all importers.** A Go interface is satisfied structurally: add a method to `storage.Client` (or any shared interface) and every concrete type that was assigned to that interface, and every `x.(*ConcreteType)` assertion, stops compiling until the concrete type gains the method. Concrete instance (2026-05-22): adding `GetPresignedPutURL` to the `storage.Client` interface for the thunder-adapter's `prepare_artefact_url` broke the **core-manager** build at `deploy_image_asset_action.go:150` (`*storage.S3Client does not implement storage.Client (missing method GetPresignedPutURL)`) — even though core-manager has nothing to do with thunder. The fix was to add the method to `*S3Client` in `platform/storage/s3.go`. Consequences to remember: (a) the implementation change and the interface change must ship together — shipping the interface alone is the code analogue of an unapplied migration (item 17); (b) **every binary importing `platform/storage` must be rebuilt** (core-manager, chassis, any adapter), not just the one that wanted the new method; (c) watch for test doubles — any mock/fake implementing the interface needs the method too (`grep -rl "storage.Client" --include=*.go`). To limit blast radius, prefer adding such methods to a *narrower* interface the caller actually needs, or as a concrete-type-only method, rather than widening a broadly-implemented interface.

When in doubt, write the assumption down and verify it before writing the code. A 30-second grep — or a 2-second `\d` — saves a 30-minute round trip.

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

### 6.1 Snapshots and revert

Before applying a patch to an agent's `default_config`, the convention is to call `snapshot_agent('<agent-type>')` first. This saves the current state so a later `revert_agent('<agent-type>')` can roll back if the patch breaks something. `snapshot_agent` accepts an optional second arg `p_reason` for human-readable context.

Where the snapshot lives depends on the deployment age:

- **Current (post-migration):** Snapshots are rows in the `agent_definitions_backup` table with the new `snapshot_taken_at` column set. That same table is also used for ad-hoc full-table backups; the two use-cases are distinguished by filtering on `snapshot_taken_at IS NOT NULL` (snapshots) vs `IS NULL` (bulk backups). `revert_agent` finds the most recent **unrestored** snapshot for the agent type (`snapshot_taken_at IS NOT NULL AND restored_at IS NULL`), copies its `default_config` to the live `agent_definitions` row, and marks the snapshot with `restored_at = NOW()`. Snapshots are preserved as an audit trail — never deleted on revert.
- **Pre-migration (legacy):** Snapshots were rows in `agent_definitions` itself, distinguished by `is_snapshot = true`, `is_active = false`, and `version = source_version + 1000`. This pattern caused several footguns documented in section 9 ("Patch UPDATE touched more rows", "Revert doesn't restore", "Agent behaviour mismatches stored prompt"). If you're on an older deployment that still has snapshot rows in `agent_definitions`, plan the migration as a separate task.

**Inspect snapshots for an agent:**

```sql
-- Current architecture (returns both used and unused snapshots — audit trail)
SELECT type, snapshot_taken_at, restored_at, snapshot_reason, version,
       LEFT(default_config->'workflow'->'steps'->'<some_step>'->'config'->>'prompt_template', 100) AS prompt_first_100
FROM agent_definitions_backup
WHERE type = '<agent-type>'
  AND snapshot_taken_at IS NOT NULL
ORDER BY snapshot_taken_at DESC
LIMIT 5;

-- Legacy architecture (pre-migration only)
SELECT id, version, is_active, is_snapshot, created_at,
       LEFT(default_config->'workflow'->'steps'->'<some_step>'->'config'->>'prompt_template', 100) AS prompt_first_100
FROM agent_definitions
WHERE type = '<agent-type>'
  AND deleted_at IS NULL
ORDER BY version DESC;
```

**Patch convention:** every UPDATE that touches `agent_definitions.default_config` should filter to the live row explicitly:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(...)
WHERE type = '<agent-type>'
  AND deleted_at IS NULL
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);  -- defensive even post-migration
```

The `is_snapshot` filter is redundant after the migration (no rows have `is_snapshot = true` anymore) but cheap to keep as a guardrail.

---

## 6.5 page_type Vocabulary and Kebab Constraint

`pages.page_type` is kebab-case and enforced by a CHECK constraint (`chk_page_type_kebab_case`) since migration 051. Snake-form inserts fail at the DB layer rather than landing as a silent inconsistency.

Canonical values:

| page_type | meaning | name shape |
|---|---|---|
| `landing` | Homepage OR conversion-focused flat page. The homepage has `name = 'index'`; other landing pages have any slug | `index` for homepage; otherwise the slug |
| `content` | Generic content page (about, contact, etc.) | the slug |
| `tool` | Interactive utility page | `tool-<slug>` |
| `guide` | Tutorial / explanatory page | `guide-<slug>` |
| `game` | Interactive game page | `game-<slug>` |
| `blog-post` | Individual blog/article page | the slug |
| `blog-index` | Index listing blog posts | `<section>-index` |
| `section-index` | Generic index for a section directory | `<section>-index` |
| `entity-page` | Individual entity within a directory | the slug |
| `entity-directory` | Index listing entities | `<section>-index` |
| `news-index` | News-style index page | `<section>-index` |

**The homepage's page_type is `landing`, not `index`.** "index" is the page's *name* (storage convention for the homepage); "landing" is its *type*. They're distinct facts and a column called `page_type` should hold the type.

**Inspection queries:**

```sql
-- Current distribution across all sites
SELECT page_type, COUNT(*) AS n
FROM pages
GROUP BY page_type
ORDER BY n DESC;

-- Confirm constraint is in place
SELECT pg_get_constraintdef(c.oid)
FROM pg_constraint c
JOIN pg_class t ON c.conrelid = t.oid
WHERE t.relname = 'pages' AND c.conname = 'chk_page_type_kebab_case';
```

If you find a snake-form value in production data, the constraint was either dropped (rare) or the row predates migration 051 and was inserted with the constraint disabled (very rare). Either way, the migration UPDATEs in 051 are idempotent — re-running them is safe.

**Backward compatibility note.** `CanonicalisePage.normalisePageType` and `ValidateRoles.normaliseRole` accept snake-form *inputs* (older planner prompts may emit `blog_post`, `section_index`, etc.) and convert them to kebab on output. This is a one-way safety net for legacy prompts; new prompts should emit kebab directly. The contracts doc's "should not be relied upon" warning applies — if you see snake values flowing through, fix the upstream prompt rather than relying on the normaliser.

---

## 6.6 LLM step config — shadowing bug

`ExecuteLLMPromptAction` resolves `ai_service` for the call in this order: top-level `default_config.ai_service`, then `workflow.steps.<currentStep>.config.ai_service`, then `params.StepConfig.Config["ai_service"]`. Once the first match is found, the others are skipped — even if the matched object is missing the field the call needs. Tracked in `FOCUS_step_level_llm_config_ignored.md`.

This means **a top-level `ai_service` shadows step-level overrides**. If the top-level lacks `max_tokens` (or `model`, or anything), the chassis doesn't fall through to look at the step. `max_tokens` then falls back to the hardcoded 2048 in `AnthropicClient.GenerateText`.

### Where max_tokens actually gets read

| Location | Read? |
|---|---|
| `default_config.max_tokens` (very top) | Always — highest priority |
| `default_config.ai_service.max_tokens` (top-level ai_service) | Yes — when the row above is absent |
| `default_config.workflow.steps.<step>.config.ai_service.max_tokens` (step's ai_service) | Only when there's NO top-level ai_service |
| `default_config.workflow.steps.<step>.config.max_tokens` (step config sibling of ai_service) | NEVER read |

### Symptom

A step that should produce a long output produces a truncated one. `llm_call_log.output_tokens = 2048` (exactly) regardless of step config. JSON outputs cut mid-object and the downstream validator fails or quietly drops the truncated tail.

### Diagnose

```sql
-- What max_tokens reaches the chassis for a specific agent/step
WITH a AS (SELECT default_config AS dc FROM agent_definitions
           WHERE type = '<agent>' AND is_active = true)
SELECT 'top_level (very highest priority)' AS source,
       (dc->>'max_tokens')::int AS value FROM a
UNION ALL SELECT 'top_level_ai_service (used by site-adoption-agent today)',
       (dc->'ai_service'->>'max_tokens')::int FROM a
UNION ALL SELECT 'step_ai_service (shadowed if top-level ai_service exists)',
       (dc->'workflow'->'steps'->'<step>'->'config'->'ai_service'->>'max_tokens')::int FROM a
UNION ALL SELECT 'step_sibling_max_tokens (DEAD — never read)',
       (dc->'workflow'->'steps'->'<step>'->'config'->>'max_tokens')::int FROM a;
```

The first non-null row is the active value. If everything is null, the hardcoded 2048 wins.

```sql
-- Cross-check what actually happened
SELECT created_at, agent_type, step_name, input_tokens, output_tokens, success
FROM llm_call_log
WHERE agent_type = '<agent>' AND step_name = '<step>'
ORDER BY created_at DESC LIMIT 5;
-- output_tokens = 2048 exactly = the hardcoded fallback fired.
```

### Workaround patterns

**Pattern 1 — agent has top-level ai_service missing max_tokens.** Add max_tokens to the top-level ai_service:

```sql
SELECT snapshot_agent('<agent>', 'set top-level ai_service.max_tokens');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{ai_service,max_tokens}',
        to_jsonb(<value>),
        true
    ),
    updated_at = NOW()
WHERE type = '<agent>' AND deleted_at IS NULL AND is_active = true;
```

**Pattern 2 — agent has no top-level ai_service.** Add max_tokens inside the relevant step's ai_service. The chassis falls through to step-level when no top-level exists:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,<step>,config,ai_service,max_tokens}',
        to_jsonb(<value>),
        true
    )
WHERE type = '<agent>' AND deleted_at IS NULL AND is_active = true;
```

**Pattern 3 — wrong location currently.** A `max_tokens` set as `step.config.max_tokens` (sibling of ai_service) is dead. Move it into `step.config.ai_service.max_tokens` OR use Pattern 1.

### Temperature — only one read path

```go
if temp, ok := agentConfig["temperature"].(float64); ok { options["temperature"] = temp }
```

Reads from `default_config.temperature` only (very top of the agent's config). NOT from inside `ai_service`. NOT from any step. Everything below is dead config:

- `default_config.ai_service.temperature` — dead
- `default_config.workflow.steps.<step>.config.temperature` — dead
- `default_config.workflow.steps.<step>.config.ai_service.temperature` — dead

To force a specific temperature today, put it at the very top of `default_config`:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(default_config, '{temperature}', '0.2'::jsonb, true)
WHERE type = '<agent>' AND deleted_at IS NULL AND is_active = true;
```

### Temperature observability gap

`llm_call_log.temperature` is universally NULL across all calls in the last 14 days, even for the 13 agents that have a top-level temperature set. We can't tell from the log whether temperature is reaching the API or not. Two possibilities:

- The chassis sets it but doesn't log it (observability bug)
- The chassis fails to set it (real bug — likely JSONB float type assertion failing)

**Fix planned:** add `options["temperature"]` capture to the `llm_call_log` writer in the next chassis deploy. That makes the two cases distinguishable from the log and gives per-call temperature visibility going forward. Tracked in the focus doc as a TODO. Same change should capture `options["max_tokens"]` for symmetry.

### Same shadowing applies to whole `ai_service` overrides

Doc 023's per-step model swap pattern (`UPDATE ... jsonb_set(... workflow.steps.<step>.config.ai_service)`) only takes effect on agents with no top-level `ai_service`. As of 2026-05-18, 38 of ~60 active agents are in that group; 22 have top-level ai_service and would shadow any step-level override.

Two known agents with shadowed step ai_services right now: `site-adoption-agent` (4 steps), `feed-triage` (1 step). Other agents listed by the query in the focus doc may have shadowed step overrides — verify with:

```sql
SELECT type, step_key.key AS step_name,
       default_config->'ai_service'->>'model' AS top_level_model,
       step_key.value->'config'->'ai_service'->>'model' AS step_model_shadowed
FROM agent_definitions,
     LATERAL jsonb_each(default_config->'workflow'->'steps') AS step_key
WHERE deleted_at IS NULL AND is_active = true
  AND default_config->'ai_service' IS NOT NULL
  AND step_key.value->'config'->'ai_service' IS NOT NULL
ORDER BY type, step_name;
```

### Structural fix path

In `ExecuteLLMPromptAction`, switch from per-object resolution (one `aiServiceConfig` object, all fields read from it) to per-field resolution with explicit fallback chain:

```
max_tokens: step.config → agent.top → step.ai_service → agent.ai_service
temperature: same shape
model: same shape
```

Plus raise the hardcoded `2048` in `AnthropicClient.GenerateText` to `8000` as a safer floor, and capture `options["temperature"]` and `options["max_tokens"]` in the `llm_call_log` insert path. Code locations and detailed rationale in the focus doc.

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

### Patch UPDATE touched more rows than expected (snapshot row contamination)

**Symptom:** A patch SQL that's supposed to update one agent's `default_config` reports `UPDATE 2` or higher. Subsequent `revert_agent` calls have no effect.

**Root cause (pre-migration / legacy deployments):** When snapshots lived in `agent_definitions` as rows with `is_snapshot = true`, a WHERE clause that filtered only on `type` and `deleted_at` matched both the live row AND the snapshot row. The patch wrote the new config to both, overwriting the snapshot's pre-patch state. The snapshot still existed structurally but no longer held the state needed to revert.

**Diagnosis:** count rows that match the patch filter before applying:

```sql
-- Pre-migration (legacy): both live and snapshot rows can match a loose filter
SELECT id, version, is_active,
       COALESCE(is_snapshot, false) AS is_snapshot,
       LEFT(default_config->'workflow'->'steps'->'<step>'->'config'->>'prompt_template', 80) AS prompt_first_80
FROM agent_definitions
WHERE type = '<agent-type>' AND deleted_at IS NULL;
-- If this returns >1 row, your patch UPDATE will touch all of them.
```

**Fix pattern:** every UPDATE on `agent_definitions.default_config` must filter to the live row:

```sql
WHERE type = '<agent-type>'
  AND deleted_at IS NULL
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
```

After migration to the separate snapshot table this filter is technically redundant (no `is_snapshot = true` rows remain in `agent_definitions`), but keep it as a defensive guardrail.

**Why this stayed hidden:** the unique `(type, version)` constraint forces snapshots to use `version + 1000`, which feels safely distinct. Operators reading their own patch SQL see `WHERE type = $X` and assume that's a single-row filter because semantically an agent's type is unique-ish. The is_snapshot column wasn't visible in most operator workflows.

### Revert doesn't restore the previous state

**Symptom:** Patch applied, behaviour observed, `revert_agent('<type>')` called, agent's behaviour is unchanged.

**Causes (try in order):**

1. **Snapshot was contaminated by an over-broad patch UPDATE.** See "Patch UPDATE touched more rows" above. The snapshot row exists but its `default_config` matches the live row's, so revert is a no-op. Pre-migration only.
2. **No snapshot was ever taken.** `snapshot_agent` was skipped before applying the patch. With nothing to restore from, revert raises `No unrestored snapshot found for type <X>`. Check the error message.
3. **All snapshots have been restored already** (post-migration). `revert_agent` only considers snapshots where `restored_at IS NULL`. After one revert, that snapshot is marked restored and a second revert call needs a fresh snapshot. The error message is the same as cause 2; distinguish them by querying `agent_definitions_backup` and looking at `restored_at`.
4. **A buggy runtime query picked up the snapshot at orchestration time** (pre-migration). Two functions known to do this: `loadAgentDefinition` (filters only on `type`) and `getAgentDefinition` (filters on `type AND deleted_at`). Neither filters on `is_active` or `is_snapshot`. With a snapshot row at `version = source + 1000`, `ORDER BY version DESC` picks the snapshot. The orchestrator is then running the snapshot's config, not the live row's. Revert "doesn't work" because it's reverting a row the runtime isn't even reading.
5. **Pod cache.** After `revert_agent` succeeds in the database, in-flight orchestrations on existing pods may still be using cached config. `kubectl rollout restart deployment/agent-chassis` to force a fresh load.

**Diagnosis:**

```sql
-- (post-migration) confirm an unrestored snapshot exists and looks right
SELECT type, snapshot_taken_at, restored_at, snapshot_reason,
       LEFT(default_config->'workflow'->'steps'->'<step>'->'config'->>'prompt_template', 100) AS prompt_first_100
FROM agent_definitions_backup
WHERE type = '<agent-type>'
  AND snapshot_taken_at IS NOT NULL
ORDER BY snapshot_taken_at DESC
LIMIT 1;
-- If restored_at is populated, this snapshot has already been used and
-- revert_agent will skip it. Take a fresh snapshot if you need to revert
-- the current state.

-- (pre-migration) confirm the snapshot still differs from the live row
SELECT id, version, is_active, COALESCE(is_snapshot, false) AS is_snapshot,
       md5(default_config::text) AS config_hash
FROM agent_definitions
WHERE type = '<agent-type>' AND deleted_at IS NULL
ORDER BY version;
-- If both rows have the same config_hash, the snapshot is contaminated.
```

**Fix path:**

- If snapshot is contaminated and you have the pre-patch state recorded elsewhere (git, a previous patch file, an `agent_definitions_backup` snapshot), restore manually by `UPDATE agent_definitions SET default_config = ... WHERE id = <live_id>`.
- If you have no copy of the pre-patch state, revert is no longer possible. Reconstruct from the latest known-good config you can find.

### Agent behaviour doesn't match the prompt stored in the database

**Symptom:** Patched an agent's prompt, restarted pods, but the orchestrator is producing output consistent with the old prompt. The new prompt is visibly present in `agent_definitions.default_config` for the live row.

**Root cause (pre-migration only):** A runtime lookup query is reading from a snapshot row instead of the live row. Both `loadAgentDefinition` (line 6010 of agent-chassis) and `getAgentDefinition` (line 26629) do `ORDER BY version DESC LIMIT 1` without filtering `is_active`. With a snapshot present at version `1000+N`, the query returns the snapshot. The snapshot's `default_config` is whatever it was at the moment `snapshot_agent` was called — which is the pre-patch state — so the orchestrator runs the old prompt.

**Diagnosis:**

```sql
-- Look at what an ORDER-BY-version-DESC query would return without is_active filter
SELECT id, version, is_active, COALESCE(is_snapshot, false) AS is_snapshot,
       LEFT(default_config->'workflow'->'steps'->'<step>'->'config'->>'prompt_template', 120) AS prompt_first_120
FROM agent_definitions
WHERE type = '<agent-type>' AND deleted_at IS NULL
ORDER BY version DESC
LIMIT 3;
-- The first row is what the buggy query path returns. Compare it against
-- the live (is_active = true) row to see if they hold different prompts.
```

**Fix path:** the migration to move snapshots into `agent_definitions_backup` (with `snapshot_taken_at IS NOT NULL` as the discriminator) removes this class of bug entirely (no snapshot rows exist in `agent_definitions` for the buggy queries to pick up). The Go queries that don't filter `is_active` should still be tightened for hygiene — those calls are listed in section 6.1's footguns.

**Workaround on legacy deployments before migration:** delete the snapshot row manually after confirming the live row holds the desired state:

```sql
DELETE FROM agent_definitions
WHERE type = '<agent-type>' AND is_active = false AND is_snapshot = true;
```

This loses revert capability but removes the runtime-misroute risk. Only do this after taking a manual `agent_definitions_backup` copy of the live row's config so you have a fallback.

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

### `pages` insert fails with `chk_page_type_kebab_case` violation

Symptom: a write to `pages` (typically from `apply_adoption_plan` or `write_site_plan`) errors with `new row for relation "pages" violates check constraint "chk_page_type_kebab_case"`. Means a snake-form value (e.g. `blog_post`, `section_index`, `entity_directory`) reached the INSERT despite the canonicaliser supposedly normalising it.

**Diagnosis:**

```sql
-- Look at the most recent attempt (if logged)
SELECT created_at, agent_type, step_name, success,
       LEFT(error_message, 300) AS err
FROM llm_call_log
WHERE success = false
  AND created_at > now() - interval '1 hour'
ORDER BY created_at DESC LIMIT 10;
```

**Likely causes:**

1. **A direct INSERT bypassing CanonicalisePage.** Some action constructs the page row inline and doesn't go through the canonicaliser. Find it: grep the action source for `INSERT INTO pages` and confirm `page_type` is being set from `pageType` (a CanonicalisePage output) rather than from raw LLM input.
2. **An LLM prompt is emitting snake-form values that no normaliser sees.** The site planner and analyze_site prompts should emit kebab. If a different prompt was added recently, it may not. Grep the agent definitions: `SELECT type FROM agent_definitions WHERE default_config::text LIKE '%blog_post%' OR default_config::text LIKE '%section_index%';`
3. **An old migration is being re-run.** Old `005_website_builder_agents.sql` and `080_clientsdb_content_creator_agent_definition_with_memory.sql` reference `"blog_post"` in agent-definition JSON. If you re-applied one of these, the agent now emits snake; either patch the agent JSON or update the migration.

**Workaround / fix:** if you need to push past the constraint to unblock an emergency build, drop the constraint temporarily, capture the offending row, then re-add the constraint after fixing the writer. Better: fix the writer.

### `pages.page_type = 'index'` rows appear after migration 051

Migration 051 renamed all `index` rows to `landing`. If new `index` rows appear:

- `CanonicalisePage` is being called via the old binary (the deploy didn't reach all pods). Verify with `kubectl -n ai-persona-system describe pod <pod> | grep Image` and check the image tag.
- A direct INSERT is bypassing the canonicaliser (see previous entry).

The constraint allows `index` because `index` matches `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` — it's a valid kebab value, just not one we use for page_type. If it's important to forbid, tighten the constraint to an explicit IN-list. Not currently done; we don't want to block legitimate single-word page_types we haven't thought of yet.

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

### Thunder API `400` on `/instances/create` — body shape and enum-value mismatches

A `provision_instance` dispatch reaches the thunder-adapter, the adapter calls `POST /instances/create`, and Thunder returns `400`. The orchestration then sits in `AWAITING_RESPONSES` and retries every 10 min (the adapter sends an `error_unrecoverable` response each time). Two distinct causes, which surface in sequence as each is fixed:

**Cause 1 — wrong field names / missing required fields.** `{"error":"invalid_request","message":"Invalid request body","code":400}`. Thunder requires `gpu_type` (NOT `gpu`), `cpu_cores` (NOT `vcpus`), plus `num_gpus`, `disk_size_gb`, `mode`, `template`. Our internal vocabulary historically used `gpu`/`vcpus`, and the original struct had `omitempty` on fields Thunder requires, so zero-values were dropped. The `invalid_request` message does NOT say which field — it's all-or-nothing.

**Cause 2 — wrong enum value.** Once the body shape is right, an unrecognised value gives a *specific* message: `{"error":"validation_error","message":"validation: invalid template: ubuntu-22.04","code":400}`. The trap: `ubuntu-22.04` is the `template` *example* in Thunder's OpenAPI spec, but it is not a real template. Real templates are `base` (plain GPU instance — use this for fine-tuning) plus AI stacks `ollama`, `comfy-ui`, `forge-neo`, `unsloth`. Enumerate with `GET /thunder-templates`.

**Cause 3 — right field, wrong case.** `{"error":"validation_error","message":"validation: invalid GPU type: A100","code":400}`. Thunder's `gpu_type` enum is lowercase: `t4`, `a100`, `a100xl`, `h100` (per the CLI reference, `tnr create --gpu a100`). The OpenAPI spec's `"H100"` example casing is misleading — same "examples aren't valid values" trap as Cause 2, one field over. Note this one was self-inflicted: the adapter originally defaulted to lowercase `"a100"` (correct), and a `strings.ToUpper` added during the Cause-1 field-name fix broke it. See checklist item 16 — don't change a value that earlier evidence proved correct while fixing something adjacent. The fix: lowercase-normalise `gpu_type` in `provision_action.go` (`strings.ToLower`), constants in `api/types.go` are lowercase.

**Diagnosis — log the request body.** You cannot debug a `400` without seeing what you sent. The adapter's API client (`internal/adapters/thunder/api/client.go`) logs both directions at info level inside the shared `do()` helper:

```
Thunder API request  method=POST path=/instances/create body_bytes=N body_preview={...}
Thunder API response method=POST path=/instances/create status=400 body_preview={"error":...}
```

The request `body_preview` shows exactly what went on the wire; the response `body_preview` shows Thunder's specific complaint. With both side by side, the fix is mechanical. (Bodies are capped at 2KB; `public_key` is a public key and safe to log; the bearer token is in the HTTP header, never the JSON body.)

**Fix.** Field names and the required-vs-omitempty distinction live in `api/types.go` (`CreateInstanceRequest`). Default values for unset fields are applied in `provision_action.go` (`api.DefaultCPUCores`, `DefaultDiskSizeGB`, `DefaultTemplate="base"`, `InstanceModePrototyping`, `GPUTypeA100`). `gpu_type` is uppercase-normalised there too, so legacy callers sending `"a100"` still work. `api/types.go` is the source of truth — change it and the rest follows by compile error.

**General lesson — OpenAPI example values are not valid values.** An `example:` in a schema is illustrative. For any enum-like string field (`template`, `mode`, `gpu_type`), find the real allowed set (a `GET /…` enumeration endpoint, or the product's own CLI/docs showing a working invocation) rather than copying the example. The staged 400s here cost two deploy cycles that the enumeration endpoint would have saved.

### Adapter response silently dropped — awaited_requests stuck `waiting` despite a sent response

**Symptom:** an adapter (thunder, or any) finishes its work, logs `Sent success response`, and the Kafka produce succeeds — but the `awaited_requests` row never leaves `waiting`, and the orchestration retries every ~10 min until it times out. No error in the adapter log; the failure is on the chassis side.

**How to confirm where it dies:** grep the chassis (`agent_type=generic`) pods for the response's `in_response_to_request_id`. The message is consumed (`Response consumer received message`), identified as `message_type: response`, then one of:
- reaches `ClaimAwaitedRequest: successfully claimed` → healthy (the row should be `processed`).
- goes to `BuildCollectedData` / generic `ProcessMessage` → routed as fresh work, not a reply (envelope-recognition problem — see the Adapter Response Envelope contract: `request_id` reuse, `message_id`, `message_type`).
- logs `Failed to unmarshal response message ... cannot unmarshal string into Go struct field ResponseHeaders.headers.is_complete of type bool` → **this entry's cause.**

**Cause:** the adapter built its response `headers` as a `map[string]string`, so `is_complete`/`is_error` went out as JSON *strings* (`"true"`). The chassis unmarshals the reply body's `headers` into `types.ResponseHeaders`, where those fields are Go `bool`. String→bool fails, the response-routing branch returns the error, and the reply is discarded **before** the claim. The awaited row is never touched.

**Fix:** build the response headers from a **typed struct** with real `bool` fields (so they marshal as JSON `true`/`false`), plus a `toKafkaHeaders()` for the Kafka message-header arg. See contracts §"Adapter Response Envelope Contract" → "Headers MUST be a typed struct". Verified end-to-end 2026-05-22: row flips `waiting → processed` ~1s after the adapter's send.

**Note:** the chassis itself emits a string `is_complete` on some internal paths into the same bool-typed struct — a latent inconsistency. Any agent whose response takes a string-bool path can hit this same wall; the durable fix would be making the chassis tolerant (accept string-or-bool) or consistent, but until then adapters must send real bools.

### Provision fails on `duplicate key ... thunder_instance_id_key` every time — recycled provider identifiers

**Symptom:** a Thunder provision reaches RUNNING, then the `thunder_instances` INSERT fails with `duplicate key value violates unique constraint "thunder_instances_thunder_instance_id_key" (SQLSTATE 23505)`; compensating cleanup deletes the instance and the provision fails. Recurs on essentially every provision.

**Cause:** Thunder recycles its numeric identifiers — once an instance is decommissioned, the next provision into the (now-empty) account gets the same low id (almost always `0`). A table-wide `UNIQUE` constraint on `thunder_instance_id` treats that recycled id as a collision with the historical decommissioned row. The constraint encodes the wrong invariant: provider ids are unique only among *live* instances, not across all history.

**Fix:** replace the table-wide unique constraint with a **partial unique index** scoped to live states:
```sql
ALTER TABLE thunder_instances DROP CONSTRAINT thunder_instances_thunder_instance_id_key;
CREATE UNIQUE INDEX thunder_instances_live_identifier_uniq
  ON thunder_instances (thunder_instance_id)
  WHERE status IN ('provisioning','running','decommissioning');
```
Pair it with a lookup change: once historical duplicate ids are allowed, `LookupByThunderIdentifier` must select the live row deterministically (`ORDER BY (status IN (live...)) DESC, provisioned_at DESC LIMIT 1`) rather than relying on the removed global uniqueness — otherwise it can Scan a decommissioned row and decommission/reaper acts on the wrong one. **General principle:** never put a table-wide unique constraint on an external identifier the provider recycles; scope uniqueness to the live subset.

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
| Many items in `detected` for site X | Discovery ran but `design-audit-agent` hasn't run since | Trigger `design-audit-agent` for that site |
| Items in `triaged` for hours, never claimed, **other site's items being dispatched normally** | A stuck `claimed` item on the same site is blocking it via the `NOT EXISTS` clause in `find_dispatchable_site`. ONE stuck claim excludes the **entire site** from dispatch — not just a priority race, an absolute block | See "Site excluded from dispatch by stuck claimed item" below |
| Items in `triaged` for hours, **whole system idle** | Dispatch loop is off, OR the scheduler isn't firing | Check `scheduled_tasks.enabled = true` for `build-pipeline-trigger`; check pod is running |
| Items in `triaged` with `pipeline='design'` or `'maintenance'` | Discovery check emitted at non-`build` pipeline; the dispatcher only loads `pipeline='build'` | UPDATE to `pipeline='build'`; fix the emitting check |
| Item in `claimed` indefinitely | Handler crashed mid-execution, or chassis died | Check `orchestration_states.error_preview`; reset to `triaged` |
| Item in `failed` with `attempt_count = max_attempts` | Handler genuinely cannot process this item | Investigate the specific failure; manual intervention |

### Site excluded from dispatch by stuck claimed item

This is the most common cause of "my items are triaged but won't dispatch when other sites' items are." The dispatcher's site-selection query (`find_dispatchable_site` step inside `build-pipeline-trigger` agent) has a `NOT EXISTS` clause that excludes any site with even one `claimed` item:

```sql
-- Excerpt from build-pipeline-trigger.find_dispatchable_site
WHERE wi.status IN ('triaged', 'approved')
  AND wi.attempt_count < wi.max_attempts
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items active
    WHERE active.site_id = wi.site_id
      AND active.status = 'claimed'
  )
```

A handler that started executing and then crashed (chassis OOM'd, network blip mid-claim) leaves its work item in `claimed` forever. From the dispatcher's perspective, the site is "busy" — even though nothing's actually executing.

**Diagnose:**

```sql
-- Anything claimed on a specific site?
SELECT id, item_type, status, claimed_at, claimed_by,
       EXTRACT(EPOCH FROM (now() - claimed_at))::int AS seconds_claimed
FROM site_work_items
WHERE site_id = '<site>'
  AND status = 'claimed';
```

If `seconds_claimed > 900` (15 min) and the corresponding orchestration_states row is FAILED or stale, it's a zombie. Reset it:

```sql
UPDATE site_work_items
   SET status = 'triaged',
       claimed_at = NULL,
       claimed_by = NULL,
       attempt_count = attempt_count + 1
 WHERE site_id = '<site>'
   AND status = 'claimed'
   AND claimed_at < now() - interval '15 minutes';
```

The site becomes eligible for dispatch on the next 30s trigger tick.

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

## Log tables — try these before tailing pod stdout

Two database tables capture structured logs that survive pod restarts and don't require the log shipper:

### `agent_error_log` — every reported error

Populated whenever an action or workflow step reports an error. Indexed on `(agent_type, occurred_at DESC)`, `(site_id, occurred_at DESC)`, and `(resolved, occurred_at DESC) WHERE resolved = false`, so the common queries are fast.

```sql
-- Recent unresolved errors across the system
SELECT occurred_at, agent_type, step_name, action, severity,
       LEFT(error_message, 200) AS err,
       domain, orchestration_id
FROM agent_error_log
WHERE resolved = false
ORDER BY occurred_at DESC
LIMIT 30;

-- Errors for a specific orchestration
SELECT occurred_at, agent_type, step_name, severity,
       LEFT(error_message, 300) AS err, context
FROM agent_error_log
WHERE orchestration_id = '<orch-id>'
ORDER BY occurred_at;

-- Errors for a specific site in the last day
SELECT occurred_at, agent_type, step_name, severity,
       LEFT(error_message, 200) AS err
FROM agent_error_log
WHERE site_id = '<site-id>'
  AND occurred_at > now() - interval '1 day'
ORDER BY occurred_at DESC;

-- Mark an error as resolved once you've addressed it
UPDATE agent_error_log
SET resolved = true, resolved_at = NOW(), resolved_by = '<your-name>'
WHERE id = '<error-id>';
```

The `context` jsonb column often holds the input that triggered the error — useful for reproducing.

### `llm_call_log` — every LLM call

Populated for each call to an LLM via the chassis (Anthropic, OpenAI, others). Contains the rendered prompt, the response, token counts, latency, and success flag. Indexed for the common access patterns by agent_type/step_name, by orchestration_id, by work_item_id, and a partial index on failures.

```sql
-- All LLM calls for a specific orchestration in order
SELECT created_at, agent_type, step_name, model, success,
       input_tokens, output_tokens, latency_ms,
       LEFT(COALESCE(error_message, ''), 200) AS err
FROM llm_call_log
WHERE orchestration_id = '<orch-id>'
ORDER BY created_at;

-- See what an LLM actually returned for a given step
SELECT created_at, step_name, success,
       LEFT(prompt_rendered, 500) AS prompt_first_500,
       LEFT(response_text, 2000)  AS response_first_2k
FROM llm_call_log
WHERE orchestration_id = '<orch-id>'
  AND step_name = '<step>'
ORDER BY created_at DESC
LIMIT 1;

-- Recent LLM failures
SELECT created_at, agent_type, step_name, model, retry_count,
       LEFT(error_message, 300) AS err
FROM llm_call_log
WHERE success = false
  AND created_at > now() - interval '1 hour'
ORDER BY created_at DESC;

-- Token / latency profile for a specific agent step
SELECT model,
       COUNT(*)              AS calls,
       AVG(input_tokens)::int  AS avg_in,
       AVG(output_tokens)::int AS avg_out,
       AVG(latency_ms)::int    AS avg_ms,
       SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) AS failures
FROM llm_call_log
WHERE agent_type = '<agent>'
  AND step_name = '<step>'
  AND created_at > now() - interval '1 day'
GROUP BY model;
```

This table is especially useful when an LLM step *appears* to have failed silently — the response is often actually present in `llm_call_log.response_text`, and you can see whether the issue was the LLM's output (e.g. malformed JSON) or a downstream parser. It's also the right tool when the workflow log says "got response, parsing failed" but you can't see the raw response anywhere else.

**`prompt_template` vs `prompt_rendered` — useful distinction the schema preserves.** `prompt_template` is the Go-template source (with `{{.field}}` placeholders); `prompt_rendered` is the actual text that was sent to the LLM after substitution. Keep this decomposition in mind when we get to LLM quality work (currently in the focus list but not active) — it makes "find me every call whose template produced malformed output" type analysis straightforward without having to reverse-engineer the substitution. Example:

```sql
-- Templates whose calls have the highest failure rate
SELECT LEFT(prompt_template, 80) AS template_first_80,
       agent_type, step_name,
       COUNT(*)                                          AS calls,
       SUM(CASE WHEN NOT success THEN 1 ELSE 0 END)      AS failures,
       ROUND(100.0 * SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) / COUNT(*), 1) AS failure_pct
FROM llm_call_log
WHERE created_at > now() - interval '7 days'
GROUP BY prompt_template, agent_type, step_name
HAVING COUNT(*) > 5
ORDER BY failure_pct DESC, calls DESC
LIMIT 20;
```

## Step-by-step log trail (chassis stdout)

Once you've confirmed the workflow reached your action and the tables above don't have what you need: look for logs before and after it. e.g.
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
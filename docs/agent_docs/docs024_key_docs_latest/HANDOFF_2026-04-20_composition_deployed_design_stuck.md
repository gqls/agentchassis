# HANDOFF 2026-04-20: Composition merge deployment, design-run stuck investigation

## Context

Continuation of the composable-theme migration (025/026 series). The site-design-planner handler and its six Go actions were deployed to production today. Composition pipeline verified working end-to-end for gamedesign.uk. Design pipeline (webdesign-agent) hit a stuck state at `generate_css` that couldn't be diagnosed before evidence evaporated. Full session transcript is long — this doc is the state to resume from.

---

## What landed successfully today

### 1. Kafka consumer fix

**Problem**: After v1.0.975 deployment, five agent-chassis pods were members of `generic-requests-group` but all showed `#PARTITIONS: 0`. Messages produced to `system.agent.generic.requests` piled up (offset climbed from 143 to 181+) with zero consumption. Work items sat in `triaged` indefinitely.

**Diagnosis**: Group state was `Stable` with 5 members, but partition assignment had landed empty. Likely a rebalance-timing issue on startup — the five pods joined around the same second and the coordinator ended up with the single partition unassigned.

**Fix applied**: Delete one pod to force rebalance. After delete-and-reschedule, the new pod (`ph6m9`) was assigned the partition and consumption resumed. Offset caught up.

**Watch for**: If pods restart at the same moment again (e.g. rolling deploy), this could recur. Worth considering staggered restarts or investigating why the initial assignment went empty. See `kafka-consumer-groups.sh --describe --group generic-requests-group --members` for the assignment state — every new deploy, check at least one member has `#PARTITIONS: 1`.

### 2. Action registry fix

**Problem**: The six composition actions had their `InputSpec` registered in `init()` blocks but had NO entry in `GlobalActionRegistry` in `registry.go`. Workflow engine rejected them as "requires a topic" (treating them as remote actions because they weren't flagged as local).

**Fix applied**: Five entries added to `GlobalActionRegistry`:
- `validate_composition_inputs`
- `resolve_composition_layout`
- `resolve_composition_typography`
- `resolve_composition_palette`
- `install_site_composition`

All with `IsLocal: true, Category: "site"`.

**Doc update**: The 026 doc should have a checklist item that every new action needs BOTH: (a) `init()` call to `RegisterActionInputSpec`, AND (b) a `GlobalActionRegistry` entry with `IsLocal: true`. The two registries are easy to confuse. Not yet added to the doc.

### 3. `industry_tags` type mismatch fix

**Problem**: `install_site_composition_action.go` and `fork_theme_from_site_action.go` both used `json.Marshal(industryTags)` and `$N::jsonb` cast for a column that is `text[]`. Postgres rejected: `column "industry_tags" is of type text[] but expression is of type jsonb (SQLSTATE 42804)`.

**This bug existed in fork_theme long before today** — silently swallowed by `forkSkipped("collection insert failed: " + err.Error())` which downgrades the error to a skip. Never surfaced because forks aren't common. My new install action didn't swallow and surfaced it immediately.

**Fix applied**:
- New helper `datahelpers.PGTextArrayLiteral([]string) string` added to `platform/orchestration/datahelpers/nullable_helpers.go`. Produces PG array literal `{"tag1","tag2"}` suitable for `::text[]` cast.
- `install_site_composition_action.go`: 3-line change (var rename, SQL cast swap jsonb→text[], parameter swap).
- `fork_theme_from_site_action.go`: same 3-line change + upgraded `logger.Warn` on collection insert failure to `logger.Error` so future silent failures are visible. `forkSkipped` still swallows but at least logs Error now.

**Binary verification**: After deploy, `strings /app/agent-chassis | grep -c PGTextArrayLiteral` returned `2` (symbol is in binary). Image tag was bumped.

### 4. First successful composition run

Work item `2b026c18-9921-4398-b71f-c9b557181b1c` on gamedesign.uk completed cleanly. Result:

- `style_collection_id`: `64d3f113-d65e-4ea6-af8b-b5b6f49a6ee2`
- `css_theme_id`: `1fa3d8ad-592e-4ede-92b8-712f9f9756b5`
- `palette_id`: `b8638525-46b9-4f7e-991e-04be4f10c9fd`
- `layout_id`: `a9001f12-df09-4571-b04c-644553fe2c09`
- `typography_set_id`: `04e560c7-fd74-4d97-89da-b3f687fd1387`
- `resolved_composition` spec written with lineage annotations.

---

## What didn't land / is open

### A. Design pipeline stuck at generate_css (webdesign-agent)

**What we queued**: `b8951daa-f882-4bf3-92cd-d5d233b066af` — `needs_design` work item for gamedesign.uk, gated on the completed composition item via `depends_on`.

**What happened**: Dispatch loop claimed and spawned `webdesign-agent` at 10:57:08. Orchestration `7399a901-e05d-42da-a0ce-39cccb4d8669` progressed through `check_site_context → load_site_context → check_has_site_id → read_site_specs → analyze_design (LLM) → update_site → check_update_db` with those keys populated in `collected_data`. Set `current_step = generate_css` at 10:58:00. Never advanced. Pod `47ac75d2-2ms6d` idle-timed-out at 11:30:33 with `awaiting_count: 0` (the 30-minute filter on `hasAwaitingOrchestrations` excluded the now-stale orchestration).

**What we could NOT determine**:
- Why `generate_css` (a deterministic Go action — `render_css_from_spec`) never produced a log line. No evidence of the action starting, no panic, no error.
- Why the orchestration's `updated_at` stopped bumping at 10:58:00 — even a stuck step should heartbeat.
- Whether the pod restart we saw in `kubectl describe` (RESTARTS 1) relates to this run or a later spawn.
- The 36-second gap between `spawn_agent` response (10:57:25) and pod startup logs (10:58:01) — something was writing state before the consumer existed. Possibly the spawn-response comes back before the pod is fully running, and some other path advanced the state. Unresolved.

**Evidence limitations hit**:
- Pod `af96bbea-5gxxv` (the dispatch-loop instance) had already terminated by the time we went to retrieve its full logs.
- Pod logs are rotated by Kubernetes — earliest lines about `7399a901` may already be gone.

**Secondary concern flagged during investigation**:

The 010 SQL migration rewired webdesign-agent's workflow so that `check_update_db.else_step` and `update_site.next_step` both point to `generate_css`. `deploy_css.next_step` is still `check_update_db` (unchanged). This means post-merge the flow is:

```
analyze_design → generate_css → deploy_css → check_update_db
                                                ↓ if site_id != null
                                              update_site
                                                ↓
                                              generate_css ← back to start
```

Every path out of `deploy_css` leads back to `generate_css`. This is a loop bug in my migration. It did NOT cause the specific stuck behaviour today (nothing got past `generate_css` the first time), but it would cause other issues on any successful render-and-deploy.

**Fix proposal (NOT YET APPLIED)**: Change `update_site.next_step` from `generate_css` to `check_should_fork`, and `check_update_db.else_step` from `generate_css` to `check_should_fork`. Original pre-migration flow was: after deploy_css/update_site, check fork, then complete. The merge should have preserved that, not redirected both to generate_css.

SQL to apply (UNTESTED — review before running):

```sql
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,update_site,next_step}',
           '"check_should_fork"'::jsonb
         ),
         '{workflow,steps,check_update_db,config,else_step}',
         '"check_should_fork"'::jsonb
       )
 WHERE type = 'webdesign-agent' AND is_active = true;
```

Verify after:
```sql
SELECT
    default_config #>> '{workflow,steps,update_site,next_step}' AS update_next,
    default_config #>> '{workflow,steps,check_update_db,config,else_step}' AS check_else
  FROM agent_definitions
 WHERE type = 'webdesign-agent' AND is_active = true;
-- Both should be "check_should_fork"
```

Even with the loop fixed, we STILL don't know why `generate_css` didn't execute. Fixing the loop doesn't resolve the stuck state. Needs a fresh session.

### B. `industry_tags = {}` empty — composition resolvers don't read classification as tags

**Problem**: gamedesign.uk classification spec clearly has `industry: "Gaming & Interactive Media"` and `sub_industry: "Game Design Resources, Tools & Education"`. But:
- `composition_layout.site_tags` came out `[]`
- `composition_layout.reason`: `"fallback — no classification tags"`
- `style_collections.industry_tags` written as empty `{}`
- Layout resolver fell back to `brochure-formal` (generic) for a site that clearly wants something dashboard/application-like

**Root cause hypothesis (not verified)**: `readClassificationFromContext` in `resolve_composition_helpers.go` (or the callers in `resolve_composition_layout_action.go` and `install_site_composition_action.go`) is looking for the wrong field. Classification stores `industry` + `sub_industry` as strings, but the resolver is probably looking for an array field like `industry_tags` that doesn't exist at the classification level.

**Impact**:
- `industry_tags` column is empty → future library-matching queries for this collection will miss it
- Layout fallback means brochure-formal CSS renders for gaming/developer-tools site
- Typography resolver did similar: matched existing `Segoe UI, Roboto...` stack when design_intent explicitly called for `Inter`-led stack

**Deferred**: Trace through these three files in next session:
- `platform/orchestration/actions/resolve_composition_helpers.go` → `readClassificationFromContext`
- `platform/orchestration/actions/resolve_composition_layout_action.go`
- `platform/orchestration/actions/install_site_composition_action.go`

Find where classification is read. Fix the field extraction to actually map industry/sub_industry strings into the `industry_tags` text[] for downstream matching.

### C. Other flagged issues (non-blocking)

- **`attempt_count = 0` on successful completions**: The dispatch-loop doesn't appear to increment attempt_count in the happy path. Cosmetic and adds confusion to logs. Already flagged in `FOCUS_page_build_handler_silent_completion.md` — status='complete' used as "terminal" rather than "succeeded."
- **`updated_at` older than `claimed_at` on work items**: The UPDATE that sets `status=claimed` and `claimed_at=NOW()` isn't bumping `updated_at`. Makes diagnostics harder. Minor.
- **Stale content-feed-trigger orchestration `b7750223` from 2026-04-19 18:44**: Sitting in `EXECUTING_STEP spawn_dispatch` for 16+ hours. Reaper hasn't picked it up. Reaper threshold is 24 hours, so it's not late yet, but the orchestration is logically dead.
- **Kafka broker pod-1 unreachable**: `failed to dial ... personae-kafka-cluster-combined-pool-prod-1 ... i/o timeout` in one build-dispatch-loop pod. Other 2 brokers serve the single partition so not blocking. Worth a `kubectl -n kafka get pods -o wide` check.
- **content-feed-trigger workflow**: Its loop step fails with `iterate_over field 'news_sites' is not an array (got <nil>)` when `find_news_sites` returns null. Pre-existing bug, fires every scheduler tick for content-feed, noise in logs.

---

## Files changed in production (recap)

| File | Status |
|------|--------|
| `platform/orchestration/actions/install_site_composition_action.go` | industry_tags fix deployed |
| `platform/orchestration/actions/fork_theme_from_site_action.go` | industry_tags fix + logger.Error deployed |
| `platform/orchestration/datahelpers/nullable_helpers.go` | PGTextArrayLiteral helper deployed |
| `platform/orchestration/actions/registry.go` | 5 composition actions registered |
| 010_remove_webdesign_install_theme.sql | Applied (but has the next_step loop bug — see A above) |
| 010b_fix_fork_theme_config_keys.sql | Applied |

Image tag after today's work: `v1.0.976`.

---

## Runbook for next session

### 1. Verify deployment still healthy

```sql
-- Confirm 010 migration still in place and site-design-planner exists
SELECT
    default_config #>> '{workflow,steps,check_update_db,config,else_step}' AS check_else,
    default_config #>> '{workflow,steps,update_site,next_step}' AS update_next,
    default_config -> 'workflow' -> 'steps' ? 'install_theme' AS install_exists
  FROM agent_definitions
 WHERE type = 'webdesign-agent' AND is_active = true;
-- Expected (current): check_else=generate_css, update_next=generate_css, install_exists=false
-- After fix in A: check_else=check_should_fork, update_next=check_should_fork, install_exists=false

SELECT type, version, is_active
  FROM agent_definitions
 WHERE type = 'site-design-planner' AND is_active = true;
-- Expected: one row, active=true
```

```bash
# Kafka consumer health — make sure we're not stuck at 0-partition assignment again
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group --members
# Expected: at least one member with #PARTITIONS: 1
```

### 2. Decide: fix webdesign loop first, or reproduce stuck state first?

**Option A**: Apply the update_site→check_should_fork fix from Section A FIRST, then rerun design smoke test. If it completes, the stuck state was caused by the loop bug (unlikely — the loop was after generate_css, and generate_css didn't execute). If it still hangs, we have isolated the stuck state from the loop issue.

**Option B**: Rerun design smoke test as-is first. If it completes this time (transient cause), we don't need to investigate. If it hangs again, we have a reproducible bug to dig into with fresh logs.

I'd recommend Option B first (cheap signal: does it happen again) before spending migration effort.

### 3. Design smoke test — fresh work item

```sql
-- Check current state
SELECT s.id, s.domain, s.style_collection_id, s.locked_at
  FROM sites s
 WHERE s.domain = 'gamedesign.uk';
-- style_collection_id should be 64d3f113-d65e-4ea6-af8b-b5b6f49a6ee2 from prior session

-- Any open design items?
SELECT id, status, created_at, item_key
  FROM site_work_items
 WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
   AND item_type = 'needs_design'
   AND status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
 ORDER BY created_at DESC;
```

If there's a stuck `smoke_design_v1_*` item from this session, reset or fail it:

```sql
UPDATE orchestration_states
   SET status = 'FAILED',
       error = 'Manually failed between sessions',
       updated_at = NOW()
 WHERE orchestration_id = '7399a901-e05d-42da-a0ce-39cccb4d8669'
   AND status NOT IN ('COMPLETED', 'FAILED', 'CANCELLED');

-- Cascade fail any dispatch-loop orchs still awaiting on the above
UPDATE orchestration_states
   SET status = 'FAILED',
       error = 'Cascade fail: child stuck',
       updated_at = NOW()
 WHERE status = 'AWAITING_RESPONSES'
   AND updated_at < NOW() - INTERVAL '1 hour';

UPDATE site_work_items
   SET status = 'failed',
       error = 'Deferred to next session'
 WHERE item_key LIKE 'smoke_design_v1_%'
   AND status NOT IN ('complete', 'failed');
```

Then queue fresh:

```sql
WITH target AS (
    SELECT id FROM sites WHERE domain = 'gamedesign.uk' AND locked_at IS NULL
)
INSERT INTO site_work_items (
    site_id, source, pipeline, item_type, severity, summary,
    spec, priority, handler_agent, status, created_by, item_key
)
SELECT id, 'manual', 'build', 'needs_design', 'high',
       'Next-session design smoke test',
       '{}'::jsonb, 8, 'webdesign-agent', 'triaged',
       'next-session-smoke',
       'smoke_design_v2_' || to_char(NOW(), 'YYYYMMDDHH24MI')
  FROM target
  ON CONFLICT DO NOTHING
  RETURNING id, item_key;
```

### 4. Things to instrument this time

When the dispatch-loop spawns webdesign-agent, capture the pod name IMMEDIATELY and start a dedicated log tail on it:

```bash
# Watch for the webdesign-agent pod spawning
kubectl -n ai-persona-system get pods -w \
  | grep webdesign-agent

# As soon as one appears, in a second terminal:
kubectl -n ai-persona-system logs -f <pod-name>
# Keep this tail going — if it stops producing output while orchestration says EXECUTING_STEP, that's the signal
```

Every ~30s while the run is in flight, query:

```sql
SELECT orchestration_id, status, current_step, updated_at,
       NOW() - updated_at AS since_last_update
  FROM orchestration_states
 WHERE owner_agent_type = 'webdesign-agent'
   AND created_at > NOW() - INTERVAL '5 minutes'
 ORDER BY updated_at DESC LIMIT 3;
```

If `current_step = generate_css` AND `since_last_update > 30 seconds`, immediately grab the owner pod's full log:

```bash
kubectl -n ai-persona-system logs <pod-name> > /tmp/webdesign-stuck.log
```

Before it terminates. This is the evidence we lost last session.

### 5. If design run succeeds

Then investigate Option B (industry_tags empty). Files to look at:

- `resolve_composition_helpers.go` — `readClassificationFromContext` function
- `resolve_composition_layout_action.go` — how it calls the above and uses the returned tags
- `validate_composition_inputs_action.go` — what shape of classification it expects
- `install_site_composition_action.go` — how it reads industry_tags for the write

Check against actual classification spec shape:

```sql
SELECT aspect, jsonb_pretty(data)
  FROM site_specs
 WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
   AND aspect = 'classification'
   AND is_current = true;
```

Look for fields: `industry` (string), `sub_industry` (string), `site_type` (string), `tone_suggestion` (string). No `industry_tags` array — that's the mismatch.

---

## Key facts worth preserving

- Composition pipeline IS working. One site (gamedesign.uk) has proof of full cycle: composition item completes, all 5 IDs in sites/css_themes/style_collections/palettes/typography_sets/layouts populated.
- webdesign-agent IS being dispatched and ISN'T receiving Kafka messages as first thought. The issue is in workflow execution, not plumbing.
- Image tag `v1.0.976` is known-good for composition. Any subsequent changes should rebuild and bump tag.
- The 010 SQL migration has an unfixed loop bug (section A). Don't declare it done until update_site.next_step and check_update_db.else_step both point at check_should_fork.
- industry_tags issue is a library-matching bug, not a pipeline-blocking one. Sites can still render without it.
- `FOCUS_page_build_handler_silent_completion.md` covers the `status='complete'` misuse — the `attempt_count=0 on success` and `updated_at<claimed_at` anomalies are symptoms of the same thing.

---

## Files to read at start of next session (in order)

1. This doc.
2. `/mnt/project/025_palette_layout_typography_migration_3_.md` — the core migration context.
3. `/mnt/project/026_design_and_site_planner_v2.md` — site-design-planner architecture.
4. `/mnt/project/010_scheduler_and_tasks.md` — scheduler & reaper behaviour (the 30-minute `hasAwaitingOrchestrations` filter is relevant).
5. `/mnt/project/bk_agent_definitions_backup.sql` — full webdesign-agent and build-dispatch-loop workflow JSON.

Do NOT re-derive the webdesign-agent workflow shape from the parallel chat handoffs — they're pre-merge. Use the live database state via:

```sql
SELECT jsonb_pretty(default_config) FROM agent_definitions
 WHERE type = 'webdesign-agent' AND is_active = true;
```

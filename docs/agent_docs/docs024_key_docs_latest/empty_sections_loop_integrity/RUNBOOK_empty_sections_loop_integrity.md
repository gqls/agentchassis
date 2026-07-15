# RUNBOOK — Empty sections & fix-loop completion integrity

Operator procedures for this workstream. Companion to
`PLAN_empty_sections_loop_integrity.md` (what/why) and
`RUNNING_NOTES_empty_sections_loop_integrity.md` (history).
Testbed: robot-hands.com, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.

## 0. Standing mechanisms

- **DB:** `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U
  clients_user -d clients_db` (read-only SELECT auto-approves; mutations prompt).
- **Page assembly reads `page_components.rendered_html` DIRECTLY** — changing
  `content_data` alone changes nothing on the live page until something
  rewrites `rendered_html`. The `empty_section` verifier therefore checks
  `rendered_html`, which is also what detection checks.
- **Zombie claims:** an item stuck `claimed` >10 min blocks its ENTIRE site
  from dispatch:
  ```sql
  UPDATE site_work_items SET status='triaged', claimed_by=NULL, claimed_at=NULL
  WHERE status='claimed' AND claimed_at < now() - interval '10 minutes';
  ```
- **Manually-inserted work items are NOT auto-triaged** — insert with
  `status='triaged'`, `triaged_at=now()`, `attempt_count=0`.

## 1. Verify deployed code against the pod (never git)

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl get deployment agent-chassis -n ai-persona-system -o jsonpath='{.spec.template.spec.containers[0].image}'

# One distinctive literal per change; expect ≥1 match each:
kubectl exec -n ai-persona-system $POD -- grep -ac "completion blocked: post-fix verification found the defect still present" /app/agent-chassis   # completion gate
kubectl exec -n ai-persona-system $POD -- grep -ac "required_fields_missing: flagged components with missing required fields" /app/agent-chassis    # new discovery check
kubectl exec -n ai-persona-system $POD -- grep -ac "wont_fix, needs_human_review, unresolved" /app/agent-chassis                                     # update_work_item_status extension
kubectl exec -n ai-persona-system $POD -- grep -ac "LLM meta-commentary in content" /app/agent-chassis                                               # meta-commentary guard
```

Spawned handler pods use `agent_definitions.image_tag`, not the standing
deployment — check both:
```sql
SELECT image_tag, count(*) FROM agent_definitions WHERE deleted_at IS NULL GROUP BY 1;
```

**Build-timing trap (bit us on v1.0.1116):** the image is built from the local
filesystem; `COPY . .` snapshots whenever that layer runs. Files saved after
the snapshot silently miss the image — always run the greps above after deploy.

## 2. Apply the workflow SQL (149, 150)

**Status: both confirmed applied 2026-07-14** against the live DB (found
already applied, presumably alongside the v1.0.1117 deploy). Re-run the
verify queries below any time to reconfirm; the apply commands are idempotent
if you ever need to re-run them (e.g. after a snapshot restore).

```bash
kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/149_page_build_handler_noop_flags.sql
kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/150_enable_required_fields_missing_check.sql
```

Verify:
```sql
-- 149: expect mark_no_ready_sections | mark_writer_skipped
SELECT default_config->'workflow'->'steps'->'check_has_ready_sections'->'config'->>'else_step',
       default_config->'workflow'->'steps'->'check_content_produced'->'config'->>'else_step'
FROM agent_definitions WHERE type='page-build-handler' AND is_active;

-- 150: expect the array to contain "required_fields_missing"
SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active;
```

Both files are idempotent-ish (150 guards on already-enabled; 149 re-applies
harmlessly). Current handler config is snapshotted in the session scratchpad
and recoverable from `sql_for_agents/065_page_build_handler_wrapper.sql` history.

## 3. Re-drive an empty_section item (the live integrity test)

ALWAYS reset `attempt_count=0` alongside `status='triaged'` and clear claim
metadata — at `attempt_count >= max_attempts` the item is silently excluded
from dispatch and just sits there.

```sql
-- pick a target (gripper-detail product sections)
SELECT id, summary, status, attempt_count FROM site_work_items
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND item_type='empty_section'
  AND summary LIKE '%product-details%gripper-detail%' ORDER BY created_at DESC LIMIT 3;

-- re-drive
UPDATE site_work_items
SET status='triaged', triaged_at=now(), attempt_count=0,
    claimed_by=NULL, claimed_at=NULL, error=NULL
WHERE id='<ITEM_ID>';
```

Dispatch fires via the build-pipeline-trigger scheduled task; then watch:

```sql
SELECT status, attempt_count, handled_by, error,
       result->'_verification' AS verification, completed_at
FROM site_work_items WHERE id='<ITEM_ID>';
```

**Expected outcomes (post-fix) — anything but `complete`:**

| State | Meaning |
|---|---|
| `needs_human_review`, error "page-build-handler no-op: no sections ready…" | SQL 149 flag fired — handler admitted it can't address the item |
| `triaged`/`failed`, error "completion blocked: post-fix verification…" | Gate blocked a false completion; `result._verification.status = "defect_persists"` |
| `complete` with `result._verification.status = "verified"` | Genuinely fixed (not expected for gripper-detail until it has a data source) |
| `complete` with NO `_verification` key | **Regression** — gate not in the running image; go to §1 |

## 4. Read a gate verdict

Every gated completion embeds evidence in the result:
```sql
SELECT summary, status, result->'_verification' FROM site_work_items
WHERE item_type='empty_section' AND result ? '_verification'
ORDER BY updated_at DESC LIMIT 10;
```
`status` values inside `_verification`: `verified` (defect gone),
`defect_persists` (completion blocked), `error` (verifier couldn't run —
completion allowed, fail-open; investigate if recurring).

## 5. Triggering a discovery pass on one site (PROVEN 2026-07-14)

`improvement-sweep` (the scheduled task that runs discovery fleet-wide) ships
**disabled** — don't flip it on for a single-site test, that's a fleet-wide
decision outside this workstream. `completeness-discovery-agent` is
`processing_mode: task`, `start_step: ensure_site_record`, and accepts
`{site_id, domain}` directly — orchestratable standalone via kcat, same
envelope contract as `idea.uk/reresolve_idea_uk_05_render.sh`:

```bash
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"   # robot-hands
DOMAIN="robot-hands.com"
AGENT="completeness-discovery-agent"
INPUT_DATA="{\"site_id\":\"${SITE_ID}\",\"domain\":\"${DOMAIN}\"}"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

kubectl -n kafka run -i --rm kcat-discovery-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=discovery-$(date +%Y%m%d-%H%M%S) -H step_name=start \
  -H client_id=demo_client -H message_type=request -H action=orchestrate \
  -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"$AGENT"},"input_data":$INPUT_DATA}
JSON
```

Watch: `SELECT status FROM orchestration_states WHERE correlation_id='<CID>'::uuid;`
→ expect `COMPLETED` within ~30s (120s workflow timeout).

Then:
```sql
SELECT summary, status, spec->'missing_fields' FROM site_work_items
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND item_type='required_fields_missing' ORDER BY created_at DESC;
```
**Proven result (2026-07-14): 8 flags**, all `needs_human_review`, no handler
— the 4 expected product-family components plus `tool-guide-intro` and 3
`article-body` instances (missing just `content` each). Cross-check on
dartsonline (site_id `5fe8785b-223d-41a3-88ee-c07187622381`, same trigger
with that site/domain) **proven: 0 flags** — the query-sourced 14-card
product grid correctly does not flag:
```sql
SELECT count(*) FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE s.domain='dartsonline.com' AND swi.item_type='required_fields_missing';
```

## 5b. Triggering a page REBUILD — use the dispatch path, NOT direct kcat

To rebuild a page (e.g. after swapping components), do NOT orchestrate
`page-build-handler` directly via kcat with a `from_agent_type=user`
envelope. It runs, plan_sections succeeds, it spawns the content-writer —
but the internal spawn→call_content_writer handshake never delivers the work
request (the child inits, sends its init response, then idles out after 180s
with `awaiting_count: 0`), and the parent hangs at `spawn_content_writer`
until the 90-min reaper fails it. Proven twice on 2026-07-15; the direct
envelope lacks the parent-orchestration context the sub-spawn needs to route
`call_content_writer` to the child's job topic.

**Use the real dispatch path instead** — re-drive the page's `empty_section`
work item (RUNBOOK §3). `build-dispatch-loop` claims it and calls
`page-build-handler` with the correct envelope; the content-writer then
receives its request and renders. Confirmed working 2026-07-15: same page,
same handler, same content-writer — via dispatch it reached
`process_sections_loop_iter_0_render_section` and rendered; via direct kcat
it hung. The handler is page-scoped, so re-driving ANY one of a page's
`empty_section` items rebuilds the whole page (all sections), which is what
you want after a component swap. Dispatch pickup can lag ~5-7 min (the
trigger loads a small batch per tick); be patient before assuming a stall.

## 5c. Section lists have THREE sources with a priority order — update ALL of them

When you change a page's section layout (a component swap, a removal), the
page-build-handler does NOT read `pages.sections` first. `load_page_sections_
from_spec` (`load_page_sections_from_spec_action.go`) reads in priority order
and SYNCS the winning source DOWN over `pages.sections`:

  1. `site_plan_sections` table (site_plans family) — **AUTHORITATIVE**
  2. `site_specs.site_plan` aspect (older planner generation; ~5 sites)
  3. `pages.sections` (materialised cache / legacy fallback)
  4. same-role sibling layout synthesis (last resort)

So if you edit only `pages.sections` and the page exists in source 1 or 2,
the next rebuild resurrects the OLD layout and overwrites your edit. This bit
the product-detail swap (2026-07-15): migration 153 updated `pages.sections`
+ the `site_specs` aspect, but product-detail was ALSO in the
`site_plan_sections` table with the old components — rebuild served source 1,
re-synced it, and the deleted product-hero/product-specs came back.
gripper-detail was fine only because it isn't in the table (its `site_specs`
aspect edit won). Fix: migration 154 corrected the table. **Before any
section-layout change, check which sources list the page:**

```sql
-- source 1 (authoritative table)
SELECT sps.ordering, sps.component_name FROM site_plan_sections sps
JOIN site_plans sp ON sp.id=sps.plan_id
WHERE sp.site_id=:site AND sp.is_current AND sps.page_name=:page ORDER BY 1;
-- source 2 (aspect JSON)
SELECT elem->'sections' FROM site_specs, jsonb_array_elements(data->'pages') elem
WHERE site_id=:site AND aspect='site_plan' AND is_current AND elem->>'name'=:page;
-- source 3 (cache)
SELECT sections FROM pages WHERE site_id=:site AND name=:page;
```
Update every source that lists the page, plus `page_components`, in one
migration.

## 6. Meta-commentary guard smoke test (deployed in v1.0.1117)

The guard runs inside `validate_page_content` during handler sagas. Cheap
verification: unit test locally
(`go test ./platform/orchestration/actions/ -run TestCheckMetaCommentary`) +
binary grep (§1). Live confirmation comes free the next time any page build
carries refusal prose — it lands as `needs_human_review` with category
`meta_commentary` in `agent_error_log` (error_code
`CONTENT_VALIDATION_BLOCKER_DETAIL`).

## 7. Zombie backlog triage (once the loop is honest)

The ~36 robot-hands `empty_section` items in `unresolved` /
`needs_human_review` / `[stale: triaged 48h+]` predate the fix. After Phase 4
(product-page decision) resolves the 6 product instances, re-drive survivors
per §3 in small batches; anything that ends `needs_human_review` with the
no-op reason needs either a data source, a plan change, or page removal — it
will never self-heal.

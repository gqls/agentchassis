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

## 5. First required_fields_missing pass (after SQL 150 + image with the check)

Trigger or await completeness discovery for robot-hands, then:
```sql
SELECT summary, status, spec->'missing_fields' FROM site_work_items
WHERE site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND item_type='required_fields_missing' ORDER BY created_at DESC;
```
Expect ~6 flags (product components), all `needs_human_review`, no handler.
Cross-check dartsonline produces **zero** (query-sourced product grid must not
flag):
```sql
SELECT count(*) FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE s.domain='dartsonline.com' AND swi.item_type='required_fields_missing';
```

## 6. Meta-commentary guard smoke test (after rebuild)

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

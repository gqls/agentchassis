# RUNBOOK — 333 owned-page door

Every command here had a gotcha attached when it was first got right. Gotchas are inline.

## Is the bug still live? (re-run before quoting any figure)

```sql
-- refusals since the 480 wont_fix terminal went live
SELECT date_trunc('day', w.updated_at)::date AS d, w.item_type, count(*)
FROM site_work_items w JOIN pages p ON p.id = w.page_id
WHERE w.handler_agent = 'page-build-handler' AND p.rebuild_policy = 'owned'
  AND w.status = 'wont_fix'
GROUP BY 1,2 ORDER BY 1, 3 DESC;
```
⚠ `site_work_items` is a ROLLING WINDOW — closing a row archives it. For any HISTORY question union
`site_work_items_archive`; for "what is open right now" the live table alone is correct.

## Which handlers refuse owned pages — the positive control

```sql
SELECT type FROM agent_definitions
WHERE deleted_at IS NULL AND is_active AND COALESCE(is_snapshot,false) = false
  AND jsonb_path_exists(default_config, '$.workflow.steps.*.config.refuse_owned_page ? (@ == true)');
```
⚠ `workflow.steps` is a jsonb **OBJECT** keyed by step name, not an array. `jsonb_array_elements` on it
fails with `cannot extract elements from an object`. Use `jsonb_each` or a jsonpath as above.
⚠ This is the door's own predicate. If it ever returns a handler that does NOT refuse owned pages, the door
will park findings that could have been repaired — that is the failure direction to watch.

## Outcomes per handler on owned pages (what a route is worth)

```sql
WITH u AS (SELECT handler_agent, status, page_id FROM site_work_items
           UNION ALL SELECT handler_agent, status, page_id FROM site_work_items_archive)
SELECT u.handler_agent, u.status, count(*)
FROM u JOIN pages p ON p.id = u.page_id
WHERE p.rebuild_policy = 'owned' AND u.handler_agent <> ''
GROUP BY 1,2 ORDER BY 1, 3 DESC;
```

## After the roll — did the door fire?

```sql
-- POSITIVE: parked rows, per finding, created AFTER the roll
SELECT item_type, count(*), min(created_at), count(DISTINCT page_id) AS pages
FROM site_work_items
WHERE status = 'deferred' AND error LIKE 'OWNED_PAGE_GUARD:%'
GROUP BY 1 ORDER BY 2 DESC;
```
⚠ Split by `created_at` vs the roll time. Legacy `detected` rows filed BEFORE the roll are still promoted and
still refused, so `wont_fix` does not drop to zero on the day — only new filings from seam producers do.
⚠ A count of ZERO is not a pass unless the demand control also ran: if no producer filed anything on an owned
page in the window, zero parked rows measures nothing.

```sql
-- DEMAND CONTROL: was there anything to park?
SELECT created_by, count(*) FROM site_work_items w JOIN pages p ON p.id = w.page_id
WHERE p.rebuild_policy = 'owned' AND w.created_at > '<roll time>' GROUP BY 1;
```

## Prove the binary carries the door (not the tag, not git)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <this commit> <the stamp>   # "did my fix ship" is a query
# no provenance line in range (it is a STARTUP line and scrolls) — probe the binary WITH A CONTROL:
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq DISABLE_OWNED_PAGE_DOOR_DEMOTION /proc/1/exe && echo PRESENT
kubectl -n ai-persona-system exec $POD -- grep -aq DISABLE_OWNED_PAGE_DOOR_ABSENTCTL /proc/1/exe && echo "CONTROL FAILED - matches everything"
```
⚠ Never `strings` (absent from the image) and never a discovery grep for "some 40-hex string" (matches Go's
internal digit table). Always run the must-be-absent control in the same breath.

## Kill switch (redeploy-free rollback)

`DISABLE_OWNED_PAGE_DOOR_DEMOTION=1` on the chassis deployment disarms the door fleet-wide; behaviour reverts
exactly to pre-guard (the handler's own cheap refusal remains). Ships ARMED.

## Tests

```bash
go test ./platform/orchestration/actions/ -run 'OwnedPage|UnregisteredHandler|Recurrence|ConflictRefresh|CreatedHonesty|ToolContent|CrossLink|NavRebuild|RenderAudit' -count=1
scripts/verify-head-builds.sh --with <changed files> --test     # build against HEAD before committing
```
⚠ Do NOT hand-roll `git archive HEAD | tar` — that recipe is why this machine runs out of space.

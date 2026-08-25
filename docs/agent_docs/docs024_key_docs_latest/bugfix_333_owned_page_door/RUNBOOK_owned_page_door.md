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

⚠ **THAT MARKER IS SAFE HERE AND UNSAFE FOR HISTORY, and the difference matters.** Every row the door writes
carries `OWNED_PAGE_GUARD` by construction, so it is the right predicate for *this* query. But the marker was
only added to `SavePageSectionsAction`'s refusal on **2026-08-19** (`bugs_open/301`) — so **any census of
HISTORICAL ownership refusals keyed on it silently drops everything older than that date**, and the dropped half
reads as a different defect. It cost me exactly that mistake on 2026-08-24 (I told the `bugs_open/384` lane 82
real refusals were an unrelated bug; it was 85 of 95, one defect). For history, match the CAUSE wording, which
predates the marker:

```sql
-- HISTORY: ownership refusals however they were worded at the time
... WHERE error LIKE '%rebuild_policy=owned%' OR error LIKE '%OWNED_PAGE_GUARD%'
```
And if a census splits cleanly on one date with zero overlap, suspect the literal rather than the data — the
full trap is in `LANDMINES.md` under "A marker in an error string has a BIRTH DATE".
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

## Monitored literal: `OWNED_PAGE_DOOR_PROBE_FAILED`

The door FAILS OPEN — if either probe errors it logs and carries on, so a finding is never lost to a pod log.
That means policy enforcement can be quietly disabled by a transient fault, and the council's `guardian` seat
asked (round 2, low) that the literal be named here as a **monitored** string rather than left as a grep-only
trail. It is emitted by both stand-down branches.

```bash
# Is the door standing down in production? (nonzero = enforcement is off for those writes)
kubectl -n ai-persona-system logs -l app=agent-chassis --since=24h \
  | grep -c OWNED_PAGE_DOOR_PROBE_FAILED
```
⚠ A count of zero here is only meaningful with a demand control — see the parked-rows query above. And note
this is a LOG line, not a durable row: it is deliberately NOT written to `agent_error_log`, because that write
would ride the same transaction and so could not survive the one failure it exists to report.

## Does any LIVE CONFIG branch on `raiseToolContentItem`'s return value?

Asked by `guardian` (round 2, low): a Go grep proves no Go caller switches on it, but a workflow step's
`condition` could match the string without appearing in Go.

```sql
-- branches naming the output field
SELECT ad.type, s.key AS step, s.value->>'condition'
FROM agent_definitions ad, LATERAL jsonb_each(COALESCE(ad.default_config#>'{workflow,steps}','{}'::jsonb)) s
WHERE ad.deleted_at IS NULL AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false
  AND s.value->>'condition' ILIKE '%content_item%';

-- and any config mentioning the return VALUES at all
SELECT type FROM agent_definitions
WHERE deleted_at IS NULL AND is_active AND COALESCE(is_snapshot,false)=false
  AND (default_config::text ILIKE '%skipped_no_prose_sections%'
    OR default_config::text ILIKE '%deduped_open_item%'
    OR default_config::text ILIKE '%parked_owned_page%');
```
[MEASURED 2026-08-24] Both return **zero live branches**. ⚠ A broader `ILIKE` over the whole config *does* hit
`experience-planner` on the literal `insert_failed` — that is a **false positive**: the string sits inside an
`execute_llm_prompt` prompt body in its `review_contracts` step, not in a condition. **Match on
`s.value->>'condition'`, not on the whole config blob**, or a prompt's prose will read as a branch.

## Consumer notifications sent (owner ruling 2026-07-29 §3)

The artefact the `guardian` seat asked for. Told, with the mechanism and what changes for each — not merely
measured:

| lane | what they were told | what came back |
|---|---|---|
| `bugs_open/326` | collides in `writeWorkItem`; agreed sequencing, sent the exact shape of the hunks | landed on top (`f16c87beb`), pinned both my constraints in their own tests |
| `bugs_open/367` | their `from_rfm` rows now park rather than fail; `row_status` is the honest field | verified my claims first-hand, froze their "28 of 31 failed" figure in five documents, and returned the `close_converted` finding |
| webdesign tool-rebuilds | largest producer; `content_item: parked_owned_page` replaces a false "raised" | — |
| `staged_component_build` / `bugs_open/353` | their backfill's 8 owned-page rows now park; three-bucket warning for the acceptance count | **returned a demand-control warning: cross-link emission is ZERO since 08-21, so an empty parked bucket will not mean the door is inert** |

## Verifying the residual fixes (committed 0ad313f02, 2026-08-25 — INERT until the next chassis roll)

Both markers are compiled string literals (the identifiers-only landmine is why they exist):

```bash
# at the artefact, BOTH replicas, with a must-be-absent control in the same breath
for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  kubectl -n ai-persona-system exec ${pod#pod/} -- sh -c \
    'grep -ac "skipped_owned_page" /proc/1/exe; grep -ac "routed via the work-item seam" /proc/1/exe; grep -ac "OWNED_DOOR_XYZZY_ABSENT" /proc/1/exe'
done   # want: >0, >0, 0 on each
```

Post-roll behaviour, each with its demand control:

```sql
-- (a) write_audit_findings now parks owned-page findings (was: raw INSERT → wont_fix).
--     Demand control: the denominator row — no filings in the window means the zero says nothing.
SELECT count(*) FILTER (WHERE wi.status='deferred' AND wi.error LIKE 'OWNED_PAGE_GUARD%') AS parked,
       count(*) FILTER (WHERE wi.status='wont_fix') AS still_refused,
       count(*) AS all_offer_analysis_owned
FROM site_work_items wi JOIN pages p ON p.id = wi.page_id
WHERE wi.created_by IN ('offer-analysis','design-audit','brief-fidelity-audit')
  AND p.rebuild_policy='owned' AND wi.created_at > '<ROLL TIME>';
-- pass: still_refused = 0 AND (parked > 0 OR all_offer_analysis_owned = 0 — then wait for demand)

-- (b) the page-rerender escalation class stops being FILED (not parked — stopped at source).
SELECT count(*) AS new_escalation_refusals
FROM site_work_items
WHERE created_by='page-rerender' AND item_type='needs_page' AND created_at > '<ROLL TIME>'
  AND (error LIKE '%OWNED_PAGE_GUARD%' OR error LIKE '%rebuild_policy=owned%');
-- pass: 0, with the demand control in the POD LOGS:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --since=24h | grep -c 'skipped_owned_page'
--   (>0 proves escalation attempts on owned pages happened and were skipped; both zero = no demand yet)
```

⚠ (a)'s producers route through `classified.HandlerAgent` — only rows whose handler declares
`refuse_owned_page` park; a producer naming `copy-editor` etc. proceeds untouched, correctly.
⚠ (b) is a stopped-at-source fix: there is no parked row to count. The log literal IS the evidence
of demand; absence of refusals alone is only a pass alongside it.

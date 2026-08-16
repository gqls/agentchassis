# RUNBOOK — bugs_open/284, flag-only work items

Every command here had a gotcha attached. When one changes, change it here.

## Is the class still live?

Ask about the ERROR, not the item_type. Asking about `capability_gap` alone finds
18 of the 60 rows and hides the type with the most damage.

```sql
SELECT item_type, status, left(error,55), count(*), count(DISTINCT site_id)
FROM site_work_items WHERE status='blocked' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

The standing exposure is the queue, not the graveyard — rows that will be blocked
on the next triage of their site:

```sql
SELECT item_type, count(*) FROM site_work_items
WHERE handler_agent='' AND status='detected' GROUP BY 1 ORDER BY 2 DESC;
```

## Which producer wrote a row — and the census that lies

```sql
SELECT item_type, status,
       count(*) FILTER (WHERE spec ? 'original_pipeline') AS via_the_promoter,
       count(*) FILTER (WHERE spec ? 'not_dispatchable')  AS via_CapabilityGapItem,
       count(*) FILTER (WHERE triaged_at IS NOT NULL)     AS was_triaged
FROM site_work_items WHERE item_type='capability_gap' GROUP BY 1,2;
```

⚠ **`spec.original_pipeline` does NOT identify the promoter on its own.** Two other
sites write that key as a hardcoded literal (`site_admin_handlers.go`
`HandleApproveWorkItem`; `tool_acceptance_actions.go` `routeChromeFailures`). The
**value** is the discriminator: the promoter writes `to_jsonb(pipeline)`, so
`design`/`content` can only be its work, while a literal writer always says
`build`.

⚠ **Do not census this class in Go source with `grep 'HandlerAgent: ""'`.** The
producer with the most damage (`check_image_url_404.go`, 40 rows) omits the field
entirely and takes Go's zero value. Ask the DB which item_types actually hold
`handler_agent = ''`, then go and find their producers.

## Reading a `090` diagnosis verdict — three wrong places first

The verdict is NOT in the work item's `result` (that holds the SPAWN record —
`bugs_open/287`'s shape), and NOT in `diagnosis_artifacts` (that holds `bundle`
rows only). It is here, and usually on a LATER orchestration row of the
correlation, not the first:

```sql
SELECT orchestration_id, current_step, status,
       (collected_data ? 'verdict') AS has_verdict
FROM orchestration_states WHERE correlation_id = '<RUN_CORR>' ORDER BY created_at;

SELECT jsonb_pretty(collected_data->'verdict')
FROM orchestration_states WHERE orchestration_id = '<the row with has_verdict = t>';
```

⚠ Query `orchestration_states` by the **`correlation_id` column**, never by a
`collected_data->'input_data'->>'fix_correlation_id'` JSON path — the latter is a
full scan and times out at 120s on this table.

Bundles-but-no-verdict has two causes that look identical. Only this tells them
apart (non-empty ⇒ the budget was exhausted, no verdict is coming, re-file with
one symbol):

```sql
SELECT substring(body from '_\(body omitted[^)]*\)_') FROM diagnosis_artifacts
WHERE correlation_id='<RUN_CORR>' AND kind='bundle' AND body LIKE '%body omitted%';
```

## Building and testing while other sessions hold the package

The `platform/orchestration/actions` package is edited by several sessions at
once, so `go test ./platform/...` in the working tree fails on THEIR half-finished
symbols (`undefined: droppedRemedy`) and tells you nothing about your change.
Test against committed HEAD plus your own files:

```bash
SB=<scratchpad>/headtree
rm -rf $SB && mkdir -p $SB && git archive HEAD | tar -x -C $SB
for f in <your changed files>; do cp "$f" "$SB/$f"; done
(cd $SB && go build ./platform/... && go test ./platform/orchestration/actions/... -count=1)
```

⚠ **Never run a bare `git` command from inside the scratchpad.**
`git rev-parse --show-toplevel` there returns `/home/ant/.claude-scratch/claude-1000`
— ONE repo holding every session's scratchpad (55 files dirty across ~40 sessions,
measured 2026-08-16). `git checkout .` / `git restore .` from that directory
destroys other sessions' working files, and the `git stash` hook ban does not cover
either. Bare `git checkout` is safe only because it takes no action without a
pathspec.

## The mutation proofs (re-run these if you touch the guard)

```bash
# 1. delete `AND %s` from the promoting UPDATE  -> TestPromotionRefusesWhatTheClaimPathWouldBlock fails
# 2. hand-write claim's EXISTS query (add AND is_active) -> TestClaimAndPromoterAskTheSameQuestion fails
# 3. drop the COALESCE half from workItemRoutableSQL -> tests 1 and 3 fail
(cd $SB && go test ./platform/orchestration/actions/ -run 'Triage|Claim|Routab|Promotion' -count=1)
```

## Repair of the stuck rows — ONLY after the guard is live

Verify first that the running binary carries the fix (per SERVICE, not per fleet):

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 7027a2801 <the stamped commit> && echo "guard is live"
```

Then, and not before (they re-block on the next claim otherwise):

```sql
UPDATE site_work_items SET status='deferred', error=NULL, updated_at=now()
WHERE item_type='capability_gap' AND status='blocked' AND handler_agent='';

UPDATE site_work_items SET status='detected', error=NULL, updated_at=now()
WHERE item_type='image_url_404' AND status='blocked' AND handler_agent='';
```

⚠ **`image_url_404` has NO retraction path** — unlike its three flag-only siblings
(`site_unreachable`, `backend_unreachable`, `head_essentials_missing`, which each
populate `CheckResult.Resolved` and close their own rows on re-observation),
`check_image_url_404.go` contains no `result.Resolved` at all. So those 40 rows go
back to `detected` and stay there until a human acts, even after the underlying
reference is repaired. That is the check's pre-existing design, not something this
fix changes — but do not repair them expecting them to clear themselves, and do not
read a persistent count as the guard failing.

The two hand-inserted rows (`page_rerender`, `needs_experience_plan` — no
`spec.original_pipeline`, `created_by` naming a session) are judged individually,
not swept: their sessions may have meant them to be dispatched, in which case the
repair is a handler_agent, not a status.

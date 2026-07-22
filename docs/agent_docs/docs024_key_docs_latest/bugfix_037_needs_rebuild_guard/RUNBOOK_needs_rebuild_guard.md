# RUNBOOK — bugs_open/037 needs_rebuild guard

`PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`

## Enumerate the affected pages (fleet)

```sql
SELECT p.page_type,
       CASE WHEN jsonb_array_length(coalesce(p.sections,'[]'::jsonb))=0 THEN 'empty' ELSE 'has_sections' END,
       count(*)
FROM pages p WHERE p.status='active' AND p.build_status='needs_rebuild'
GROUP BY 1,2 ORDER BY 1,2;
```

Detail with component counts (distinguish *awaiting composition* `n_comp=0` from *rendered elsewhere*
`n_comp>=1`, which is `/bugs_open/050`'s classification):

```sql
SELECT s.domain, p.name, p.page_type,
       jsonb_array_length(coalesce(p.sections,'[]'::jsonb)) AS n_sec,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS n_comp
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND p.build_status='needs_rebuild'
ORDER BY s.domain, p.name;
```

## Find every setter of `needs_rebuild` (the diagnosis)

```bash
grep -rn "SET build_status = 'needs_rebuild'\|\"build_status\": *\"needs_rebuild\"" \
  --include='*.go' platform/ internal/ pkg/
```

## Run the unit tests

```bash
export TMPDIR=<repo>/.gotmp_037 GOTMPDIR=<repo>/.gotmp_037   # /tmp is a 16G tmpfs and fills up
go test ./platform/orchestration/actions/ -count=1 \
  -run 'TestReconcile|TestTruncate|TestRealisedPageCompositionIsPreserved' -v
```

## Prove the tests are discriminating WITHOUT touching the shared (concurrently-edited) tree

`v3_site_actions.go` is under constant concurrent edit — an in-place neutralise WILL be clobbered by
another session's commit (it happened, see NOTES). Use an isolated worktree:

```bash
WT=<repo>/.wt_037
git worktree add --detach "$WT" HEAD
cp platform/orchestration/actions/v3_site_reconcile_test.go "$WT/platform/orchestration/actions/"   # overlay uncommitted tests
# neutralise the fix in the worktree only, then:
( cd "$WT" && go test ./platform/orchestration/actions/ -count=1 -run 'TestReconcile_NeedsRebuild|TestRealisedPageCompositionIsPreserved' -v )
git worktree remove --force "$WT"
```

Two neutralisations to try: (a) drop `needs_rebuild` from `realisedPageCompositionIsPreserved` →
the three membership tests FAIL; (b) instead widen `realisedPageIsBuilt` to include `needs_rebuild`
(the naive fix) → `TestReconcile_NeedsRebuildEmptyPageIsStillComposable` FAILS.

## Verify the fix is LIVE (never trust the tag)

```bash
POD=$(kubectl -n ai-persona-system get pods -o name | grep 'agent-chassis-' | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec "$POD" -- sh -c 'strings /app/agent-chassis | grep -c realisedPageCompositionIsPreserved'  # expect 1
kubectl -n ai-persona-system exec "$POD" -- sh -c 'strings /app/agent-chassis | grep -c reconcilePlanWithRealised'          # positive control, expect 2
```

## features_open/012 — deliberately recompose a specific page (once the chassis rolls)

Emit a `needs_site_plan` with `recompose_pages` in its spec; every page NOT named keeps its layout,
the named ones are redesigned by the LLM:

```sql
INSERT INTO site_work_items (id, site_id, item_type, status, spec, handler_agent, created_at)
SELECT gen_random_uuid(), s.id, 'needs_site_plan', 'detected',
       '{"recompose_pages":["index"]}'::jsonb, 'build-site-planner', NOW()
FROM sites s WHERE s.domain = '<domain>';
```

Verify the feature is live (Go change, inert until the next image roll):

```bash
POD=$(kubectl -n ai-persona-system get pods -o name | grep 'agent-chassis-' | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec "$POD" -- sh -c 'strings /app/agent-chassis | grep -c recomposePagesFromSpec'  # expect 1 once rolled
```

Then live-check behaviour: after the run, the named page's `pages.sections` should be the LLM's new
composition; unnamed peers unchanged. Look for the log lines `recompose_pages requested` and
`recompose — realised pages released from the preserve guard`.

## (Optional) live behavioural verification

Pick a site with a `needs_rebuild` page carrying a real composition (e.g. dartsonline `contact`).
Snapshot `pages.sections`, emit a `needs_site_plan` (see `idea_uk_vm_site/sql/p1_01_replan_emit.sql`
for the shape), let it run (~30 min dispatch), then assert `pages.sections` unchanged for that page.
This mutates a live site + spends build capacity — get owner sign-off first.

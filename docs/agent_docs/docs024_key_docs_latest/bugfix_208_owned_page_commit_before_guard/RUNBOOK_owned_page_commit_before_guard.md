# RUNBOOK — bugs_open/208, owned pages in generic build loops

Every command that was hard to get right, with its gotcha attached. Change it HERE.

## R1 — Find every live consumer of an action (nested walk, NOT top-level)

**Gotcha:** a top-level `jsonb_each` over `{workflow,steps}` finds only the steps at the top
level. Every step inside a loop's `sub_workflow` is invisible to it — that under-reporting is
already recorded at `save_page_sections_action.go:180` (it once turned "6 callers" into "3").
Use the recursive path query:

```sql
SELECT ad.type AS agent, s.key AS step_name,
       s.value->>'next_step'  AS next_step,
       s.value->>'output_field' AS output_field,
       s.value->'config'      AS config
FROM agent_definitions ad,
     LATERAL jsonb_path_query(ad.default_config, '$.**.steps') AS steps,
     LATERAL jsonb_each(steps) AS s(key, value)
WHERE s.value->>'action' = '<action_name>'
  AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
ORDER BY ad.type, s.key;
```

Always carry the three-part liveness filter (`is_active`, not a snapshot, not deleted) or
snapshots inflate the consumer count.

## R2 — Read a whole live step graph (to check ORDER, which no grep can tell you)

```sql
SELECT s.key AS step, s.value->>'action' AS action, s.value->>'next_step' AS next_step
FROM agent_definitions ad,
     LATERAL jsonb_path_query(ad.default_config, '$.**.steps') AS steps,
     LATERAL jsonb_each(steps) AS s(key, value)
WHERE ad.type = 'page-rebuild'
  AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
ORDER BY s.key;
```

**Gotcha:** the rows come back alphabetically, not in execution order — follow `next_step` by
hand. `page-rebuild`'s loop body reads
`plan_sections → write_page_content → review_page_content → check_review_approved →
assemble_page → deploy_page → save_sections → update_page_status → complete_page`.

## R3 — The exposure census (the population a fix must protect)

```sql
SELECT p.name, s.domain, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status='active'
  AND COALESCE(p.rebuild_policy,'generic')='owned'
  AND COALESCE(p.build_status,'planned') IN ('needs_rebuild','planned')
ORDER BY s.domain, p.name;
```

**Gotcha:** `COALESCE` both columns. `rebuild_policy` is `NOT NULL DEFAULT 'generic'` today so
the first is belt-and-braces, but `build_status` is genuinely nullable and a bare
`build_status IN (...)` silently drops those rows — the same shape as the defect being fixed.
Also run it **without** the policy filter to get the denominator; a count of owned pages alone
cannot tell you whether the filter is doing anything.

## R4 — Is another session already on this bug? (both halves)

`scripts/who-owns.py <n>` reads **commits**, so a bug FILED an hour ago reads as OWNED. It
cannot see an uncommitted session at all. Do both:

```bash
./scripts/who-owns.py 208
cd ~/.claude/projects/-home-ant-projects-agentchassis
for f in $(ls -t *.jsonl | head -30); do
  hits=$(tail -c 400000 "$f" | grep -oE "208_HANDOFF|<your symbols>" | sort | uniq -c | tr '\n' ' ')
  [ -n "$hits" ] && echo "$f [$(date -r $f +%H:%M)] $hits"
done
```

**Gotcha:** `tail -c` on a file being appended to by a live session can panic under this
`tail` build (`Os { code: 22 }`) — harmless, but it silently skips that file, so a clean sweep
is not proof. Re-run, or read the file with python if it matters.

## R5 — Verify a fix WITHOUT destroying a live tool page

Never verify this by rebuilding a real owned page. In order:

1. **Unit** — sqlmock, following the idiom in
   `platform/orchestration/actions/save_sections_stored_slot_identity_test.go:324-327`
   (`mock.ExpectQuery("SELECT COALESCE\\(rebuild_policy")`).
2. **Query-level** — run R3, then run the fixed selection SQL against a site that has an owned
   page at `needs_rebuild` and assert the owned page is **absent** from the returned set.
   Assert row **identity**, not a count: a count that drops by one proves something was
   excluded, not that it was the right something.
3. **Live, on a safe target** — dispatch a real rebuild on a site whose `needs_rebuild` set is
   **generic-only**, and assert the owned page's `updated_at` **and** its served HTML are both
   unchanged. Include a **negative control**: a generic page in the same run must still be
   rebuilt, or a fix that excludes everything passes just as well.

## R6 — Verify a change when the shared tree does not compile (it often does not)

Three sessions edit this tree at once, and a build failure is as likely to be someone's
half-written function as your own mistake. On 2026-08-06 the tree failed on
`plan_sections_action.go:1007: undefined: composeScopedWriterBlock` — a symbol present nowhere,
in a file this lane never touched.

**Do not** `git stash` (it takes other sessions' work) and do not conclude anything from a red
tree. Build HEAD plus only your own files:

```bash
SC=<scratchpad>/head208
rm -rf $SC && mkdir -p $SC
git archive HEAD | tar -x -C $SC
for f in <your changed files>; do cp "$f" "$SC/$f"; done
(cd $SC && go build ./platform/... && go test ./platform/orchestration/...)
```

**Gotcha:** this is exactly what `make build-<service>` archives once you commit, so a green
result here is a real prediction about the image. It is also the only way to tell your breakage
from theirs. Diagnose whose it is first: `grep -rn "func <symbol>"` over the tree **and**
`git grep "func <symbol>" HEAD` — absent from both, with the calling file dirty, means
uncommitted WIP that is not yours.

## R7 — Mutation-prove a guard (and the trap that makes a passing test worthless)

A guard with a green test is not a guard with a proven test. Break it and require a **named**
test to fail:

```bash
cp <file>.go /tmp/m.bak
sed -i 's|return policy == ownedRebuildPolicy|return false // MUTANT|' <file>.go
(cd $SC && go test ./platform/orchestration/actions/ -run '<Pattern>' 2>&1 | grep -E "^(ok|--- FAIL)")
cp /tmp/m.bak <file>.go
```

**Gotcha, and it cost a rewrite here:** if the guard's success shape is `{skipped:true}` and the
same function has other paths returning `{skipped:true}`, the mutation **passes** — a guard in
series produces the same answer for a different reason. Before writing the assertion, grep the
function for every other `return` carrying the key you plan to assert on. If there is more than
one, assert the **discriminator** instead (here: `reason`, which only the intended path fills
with `OWNED_PAGE_GUARD`).

## R8 — Tell an action's other consumers (the ruling requires telling, not measuring)

Find who else calls the Go function — not just who uses the action name, which under-reports:

```bash
grep -rn "queryPagesForBuild(" --include=*.go . | grep -v "_test"
```

Then the *live* consumers of the action, with R1's nested walk. On 2026-08-06 the action had 2
consumers and the function had 3 call sites; the third (`WriteBuildItemsAction`) was invisible to
the action-level query and was found only by the compiler complaining about the new parameter.
**A changed signature is a free consumer census — read the compiler errors as a list.**

## R9 — The canary's assertion set (run after the dispatch terminates)

One psql block, six assertions. `<FLAG_TS>` is the timestamp `flagPagesForRebuild` stamped when
it set `needs_rebuild` — the flagging is upstream of the guard and legitimately moves
`updated_at` once; the assertion is that nothing moves it AGAIN.

```sql
-- A2: exactly one review item, and from MY guard, not reconcile
SELECT source, status, spec->>'refused_by', created_at FROM site_work_items
WHERE item_key='owned_page_review:zz-canary-208';
-- A4: row untouched since the flag; still needs_rebuild; never stamped deployed
SELECT name, build_status, updated_at, deployed_at,
       updated_at = '<FLAG_TS>'::timestamptz AS untouched_since_flag
FROM pages WHERE name='zz-canary-208';
-- A1+A5: the run's outcome + which terminal step it took
SELECT status, current_step FROM orchestration_states WHERE orchestration_id='<REBUILD_ORCH>';
-- A3a: no components were written
SELECT count(*) AS must_be_0 FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.name='zz-canary-208';
```

A3b (nothing committed): `curl -sS -o /dev/null -w '%{http_code}' https://oufe.com/zz-canary-208.html`
must stay **404** (it was 404 pre-test — measured, not assumed).
A6 (dedup): re-dispatch, then re-run the first query — still exactly one row.

**Gotcha:** `updated_at` moving at flag time is CORRECT and is not a guard failure —
`flagPagesForRebuild` is a bare `UPDATE pages SET build_status='needs_rebuild'` and sits upstream
of selection. Reading that first bump as "the guard touched my page" would be a false alarm.

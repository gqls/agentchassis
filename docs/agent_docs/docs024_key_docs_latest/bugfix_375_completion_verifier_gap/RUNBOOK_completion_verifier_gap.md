# RUNBOOK — `bugs_open/375`

Every command here was needed to get something right. The gotcha is attached to the command, not
kept in somebody's scrollback.

---

## 1. The census: who completes through the UNGUARDED writer

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

⚠ **Scan RECURSIVELY, not `$.workflow.steps`.** A step can sit inside a nested loop-step config,
and the narrow query silently misses it. This is not hypothetical: the narrow query finds only
**2** of the **4** live agents that name `complete_work_item`, because the dispatch loops carry it
nested. The `update_work_item_status` side happens to be flat today — the recursive query returns
the same 22 steps — but the way you learn that is by running both.

```sql
-- RECURSIVE, the one to trust
WITH live AS (SELECT type, default_config FROM agent_definitions
              WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
     anystep AS (
       SELECT l.type,
              jsonb_path_query(l.default_config,
                'strict $.**{0 to last} ? (@.action == "update_work_item_status")') AS st
       FROM live l)
SELECT type, COALESCE(st->'config'->>'status','(default=complete)') AS sets_status, count(*) AS n
FROM anystep GROUP BY 1,2 ORDER BY 1,2;
```

⚠ **`COALESCE(..., '(default=complete)')` is load-bearing.** `UpdateWorkItemStatusAction` defaults
`status` to `complete` when the key is absent, so a step with no `status` is a `complete` arm that a
`WHERE config->>'status' = 'complete'` query cannot see. All 22 live steps name a status explicitly
as of 2026-08-24 — but the query must be able to tell you when that stops being true.

## 2. The item types those agents handle, with the control

```sql
SELECT handler_agent, item_type, count(*) AS rows,
       count(*) FILTER (WHERE status='complete') AS completed,
       max(created_at)::date AS last_filed
FROM site_work_items
WHERE handler_agent IN ('image-build-handler','image-source-unsatisfiable-handler',
                        'image-url-404-handler','required-fields-missing-handler')
GROUP BY 1,2 ORDER BY 1,2;
```

⚠ **Control the zero.** The headline is "the intersection with registered verifiers is ZERO", and a
zero produced by a mis-spelled item type in an `IN` list is indistinguishable from a real one. Run
the same registered-type list *without* the handler filter and require real rows back:

```sql
SELECT item_type, handler_agent, count(*) AS rows
FROM site_work_items
WHERE item_type IN ('content_duplication','dead_fragment_link','decision_regression','empty_section',
                    'hardcoded_section_colors','literal_markdown','missing_conversion_path',
                    'needs_brand_head_assets','orphan_element_refs','page_canonical_collision',
                    'revenue_shape_cta','truncated_component','unbuilt_internal_link')
GROUP BY 1,2 ORDER BY 1,2;
```
Expected 2026-08-24: **12 rows, 10 distinct types**, and **no handler among the four above**.
⚠ 12 rows is not 12 types — it is (type, handler) pairs; `literal_markdown` has three handlers.
The handoff quoted the row count as a type count; do not repeat that.

## 3. Enumerate the registered verifiers

```bash
grep -rEho 'RegisterVerifier(WithPolicy)?\("[a-z_]+"' platform/ internal/ --include=*.go \
  | grep -v _test | sed -E 's/.*"(.*)"/\1/' | sort -u
```
⚠ **`RegisterVerifier(` does NOT match `RegisterVerifierWithPolicy(`.** The naive grep returns 11
and misses `hardcoded_section_colors` and `needs_brand_head_assets`. **13** as of 2026-08-24.
⚠ Do not count raw grep lines — the registration *functions* and two comments match too (18 lines,
13 types).

## 4. Re-locate the code by SYMBOL, never by line

```bash
grep -n 'func UpdateWorkItemStatusAction' platform/orchestration/actions/v3_site_actions.go
grep -rn 'GetVerifier(' platform/ internal/ --include=*.go
```
⚠ The bug file's line numbers drifted ~30 lines in one day. `:5978` as of 2026-08-24; it will move.
The claim to re-verify is that **no `GetVerifier` call exists anywhere in that function's body** —
the next `func` after it is `containsString`, so the body is the range between the two.

## 5. Build and test just this package

```bash
go build ./platform/orchestration/actions/... ./platform/orchestration/actions/discovery_checks/...
go test ./platform/orchestration/actions/ -run 'Verif|Complete' -count=1
go test ./platform/orchestration/actions/discovery_checks/ -run Verifier -count=1
```

⚠ **Check the change against committed HEAD before committing** — never hand-roll
`git archive HEAD | tar` (that recipe is why this machine runs out of space):
```bash
scripts/verify-head-builds.sh --with platform/orchestration/actions/v3_site_actions.go --test
```

## 6. Prove the guard is load-bearing (mutation, not bookkeeping)

⚠ **A mock's own bookkeeping cannot assert a negative.** A test that says "the verifier was not
called" passes just as happily when the whole arm is unreachable. So:
1. run the test — it passes;
2. **neuter the guard** (delete the opt-in branch / force `mayComplete = true`);
3. **require the test to FAIL.** If it still passes, the test proves nothing.

⚠ And the sibling trap — *"a mutation that PASSES may have hit a guard in SERIES"*. This arm
already carries the terminal-decision guard (`workItemCompletionGuardStatuses`). A fixture whose
row is already `failed`/`wont_fix` is refused by THAT guard, so neutering the verifier consult
changes nothing and the mutation reads as "covered". **Fixture rows must be in a status the
terminal guard lets through** (`detected`, `claimed`, `triaged`) or the mutation test is vacuous.

## 7. Council gate

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh submission.json
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh submission.json
```
⚠ Budget ~30 minutes, not ~2 — the council runs in 2–5 but the dispatch queues behind the fleet.
⚠ Find the run by PAYLOAD, not by the printed id:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

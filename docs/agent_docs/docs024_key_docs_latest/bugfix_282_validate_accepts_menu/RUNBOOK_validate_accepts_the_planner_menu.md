# RUNBOOK — bugfix 282 (validate accepts the planner's menu)

Every command here was needed at least once and had a gotcha attached. Newest
gotchas stay with their command, not in a footnote.

## Prove the bug (or prove it fixed) at the run, not at the plan

`orchestration_states` has **no `agent_type` column** — it is `owner_agent_type`.
(Cost two failed queries in this lane and one in a subagent; the error is
`column "agent_type" does not exist`, which reads like a missing join.)

```sql
-- the three artefacts of one planner run, in the order the drop happens
SELECT jsonb_array_length(collected_data->'available_components')      AS menu_rows
  FROM orchestration_states WHERE correlation_id::text LIKE '2f74a975%';

-- what the planner PROPOSED (raw LLM, one table upstream of the plan)
SELECT p->>'name', p->'sections'
  FROM orchestration_states, jsonb_array_elements(collected_data->'llm_plan'->'result'->'pages') p
 WHERE correlation_id::text LIKE '2f74a975%' AND p->>'name'='index';

-- what validate ACCEPTED (its own output field — the drop is between these two)
SELECT p->>'name', p->'sections'
  FROM orchestration_states, jsonb_array_elements(collected_data->'validate_plan'->'pages') p
 WHERE correlation_id::text LIKE '2f74a975%' AND p->>'name'='index';
```

Which pages carried a tool-level function in the raw plan (the general form —
joins the proposal against the component table by level, so it does not depend
on knowing the function names):

```sql
SELECT p->>'name', p->'sections'
  FROM orchestration_states, jsonb_array_elements(collected_data->'llm_plan'->'result'->'pages') p
 WHERE correlation_id::text LIKE '2f74a975%'
   AND EXISTS (SELECT 1 FROM jsonb_array_elements_text(p->'sections') s
                 JOIN content_components c ON c."function"=s
                WHERE c.component_level='tool' AND c.is_active);
```

## Before believing "the resolver ate this name"

**Check the name's level first.** A name that survived is not evidence of a
second branch — it may simply be a section-level component with a tool-ish name.
This is what corrected the bug file's ADDENDUM:

```sql
SELECT id, "function", component_level, is_active FROM content_components
 WHERE "function" = '<the surviving name>';
-- loans-credit-health-check -> component_level='section' (created 2026-08-13). Arm 1 of resolve() accepts it. No second branch.
```

## Enumerate the consumers (never assert them)

```sql
-- who runs validate_site_plan at all
SELECT type, k FROM agent_definitions, jsonb_object_keys(default_config#>'{workflow,steps}') AS k
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config#>'{workflow,steps}'->k->>'action'='validate_site_plan';
-- -> build-site-planner.validate_plan, site-planner.validate_plan (only these two)

-- every menu-shaped step over content_components, fleet-wide (LANDMINES: the
-- NAMED call site is not the consumer set — bugs_open/276 found the dominant
-- one this way, 65x the named site's traffic)
SELECT type, k, position('component_level = ''tool''' in (default_config#>'{workflow,steps}'->k#>>'{config,query}'))>0 AS has_tool_arm
  FROM agent_definitions, jsonb_object_keys(default_config#>'{workflow,steps}') AS k
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config#>'{workflow,steps}'->k#>>'{config,query}' ILIKE '%content_components%'
   AND default_config#>'{workflow,steps}'->k#>>'{config,query}' ILIKE '%component_level%';
```

Go side, the same question:
`grep -rn "component_level IN ('section'\|component_level = 'section'" platform/ --include=*.go | grep -v _test`
→ 5 hits; only `v3_site_actions.go`'s resolver is on the plan-acceptance path.

## Which sites are opted in to the widened menu

```sql
SELECT s.domain, ss.data->>'plan_includes_tools'
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE ss.aspect='structure' AND ss.is_current AND ss.data ? 'plan_includes_tools';
-- 2026-08-16: loancalculator.co.uk only.
```

## Test + mutation (the tests are worthless unmutated)

```bash
go test ./platform/orchestration/actions/ -run 'AddMenu|MenuRowsFrom|ValidateSitePlan_Menu|ValidateSitePlan_Without' -count=1 -v
# then break each arm in turn and confirm the RIGHT test fails:
#   addMenu -> no-op                  => MenuFieldKeepsAToolSection... fails
#   the menu_field gate always-on     => WithoutMenuFieldTheToolSection... fails
```

`go vet ./platform/orchestration/actions/` reports `load_component_library_actions.go:207: unreachable code` — **pre-existing, not this lane's**.

## Migration

```bash
./scripts/migration/run-migrations.sh            # dry run FIRST, every session
./scripts/migration/run-migrations.sh --apply    # takes EVERY pending file — read the dry run's list before firing
```
The dry run probes each pending file by executing it inside a doomed
transaction, so it is **slow** (minutes, not seconds) — run it in the background
rather than assuming it has hung.

## After the roll (the fix is Go — inert until an image carries it)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <this lane's commit> <the sha in that line>   # exit 0 = shipped
```
An empty grep means "not in range" (it is a startup line and scrolls), **not**
"unstamped".

> **CORRECTED 2026-08-16, same day, after this recipe failed its own control.**
> Two things above do not work on this fleet as written, and I only found out
> because I ran a control:
>
> 1. **Do NOT grep the binary for YOUR commit.** I probed `/proc/1/exe` for this
>    lane's sha (absent, as expected) *and* for `7d9b7334a`, a commit the 285
>    lane had already proven live on the running image — **also absent**. The
>    control failing is the finding: **the binary carries ONE sha, the commit it
>    was BUILT at**, so a grep for any other commit returns absent no matter what
>    the image contains. Without the control I would have written "the fix is not
>    live" from a test that says that about every commit ever made.
> 2. **The chassis log history is ~90 seconds** (a LANDMINES entry says so
>    outright), so `--tail=100000 | grep 'build provenance'` finds nothing on a
>    busy pod — not even from this morning's roll.
>
> **What works, with no stamp at all — prove the ORDERING:**
> ```bash
> kubectl -n ai-persona-system get pods -l app=agent-chassis \
>   -o jsonpath='{.items[*].spec.containers[0].image}' | tr ' ' '\n' | sort -u   # -> v1.0.1304
> git merge-base --is-ancestor <a commit KNOWN live on that image> <your commit> # YES => yours is later
> git merge-base --is-ancestor <your commit> <that known-live commit>            # NO  => yours is not in it
> grep -m1 '^IMAGE_TAG' makefile    # still v1.0.1304 => no build has been cut since
> ```
> Measured this way 2026-08-16: `7d9b7334a` (live on v1.0.1304) IS an ancestor of
> `adb1ee2ad`, and the reverse is false, with the deployed tag and `IMAGE_TAG`
> both still `v1.0.1304` — so this lane's Go half is definitively **not live**.
> **Whoever rolls must bump `IMAGE_TAG`**: a same-tag rebuild ships the node's
> stale cached binary.
>
> **REFINED again after the council's round-3 advisory (`debug_historian`, and it
> is a fair challenge):** "prove a deploy at the ARTEFACT, never at git" is the
> estate's rule, and an ordering proof looks like exactly the anti-pattern it
> bans. The distinction that makes both true is **which direction you are
> proving**, and it belongs in any recipe of this shape:
>
> | claim | what actually proves it |
> |---|---|
> | "my fix is NOT live yet" | the **deployed image tag**, read off the running pod (`kubectl get pod -o jsonpath='{.spec.containers[0].image}'`) — that IS an artefact reading — plus `IMAGE_TAG` showing no newer build exists. Git supplies only the ordering *between commits*, never the deploy fact |
> | "my fix IS live" | **the binary's own stamp**, obtained from the pod, then `git merge-base --is-ancestor <your-commit> <that stamp>`. Nothing in git can establish this, because git does not know what was built |
>
> So: never claim SHIPPED from ordering alone. The negative is safe because it
> rests on a pod-read tag; the positive requires the stamp. Both halves read the
> artefact — the only thing git contributes is the ancestry comparison, which is
> arithmetic on commits, not evidence about deployment.

## Migration 439 — applied ONCE, and a re-run is not free

439 is applied and recorded (`--record-only`, with the reason). It is guarded and
would refuse a wrong-shaped apply, but **it is not free to re-run**: like every
migration in this shop it opens with `snapshot_agent(...)`, so a second run takes
a **second snapshot whose reason still reads "pre-update"** — a documented
landmine, because the snapshot then records the post-change state under a
pre-change label. If you need to reapply, take the rollback first, or expect two
snapshots and know which is which:

```sql
SELECT id, created_at, notes FROM agent_definitions
 WHERE type='build-site-planner' AND COALESCE(is_snapshot,false)
 ORDER BY created_at DESC LIMIT 3;
```

Then re-fire the loancalculator lane's own script — no dispatch within ~300 s of
a chassis pod restart, and `kcat -P` exits 0 having sent nothing, so prove the
dispatch by the orchestration row, not by the exit code:

```bash
./docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/phase2_recompose_26.sh
```

Acceptance — read `site_plan_sections`, NOT `pages.sections` (LOCK-008 changes
the cache on exactly these pages):

```sql
SELECT page_name, ordering, component_name
  FROM site_plan_sections
 WHERE plan_id=(SELECT id FROM site_plans WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND is_current)
   AND component_name LIKE 'tool-%' ORDER BY page_name, ordering;
-- want: the 12 locked tool functions on their own pages. Baseline with no fix: plan dcbae4df = 0 of 12.
```

And the log tell that the new arm actually fired (not merely that the planner
proposed nothing tool-level):
`kubectl -n ai-persona-system logs -l app=agent-chassis --since=1h | grep "resolved via the planner's menu"`

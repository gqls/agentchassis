# RUNBOOK — `bugs_open/201` lane

Every command that was hard to get right, with its gotcha attached. Fix it HERE, not in
scrollback.

## R1 — the cause, in one query: does the item's spec carry `sections`?

This is the whole bug. `page-content-writer`'s self-plan reads
`input_data.current_page.sections`; a discovery spec has no such key.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT wi.item_type, (wi.spec ? 'sections')::text AS has_sections, keys, count(*)::text
FROM (SELECT wi.id, wi.item_type, wi.spec,
             (SELECT string_agg(k,',' ORDER BY k) FROM jsonb_object_keys(wi.spec) k) AS keys
      FROM site_work_items wi WHERE wi.handler_agent='page-content-writer') wi
GROUP BY 1,2,3 ORDER BY 4::int DESC;"
```

**Gotcha:** `jsonb_object_keys` is a set-returning function — it cannot sit bare in a `SELECT`
list beside an aggregate. Wrap it in a scalar sub-select (as above) or it errors, and an
errored check has no result to disagree with.

## R2 — where does an agent get its section list? (the comparison that decides the fix)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c "
SELECT ad.type || ' :: ' || s.key || ' => ' || jsonb_pretty(s.value)
FROM agent_definitions ad,
LATERAL jsonb_path_query(ad.default_config, '\$.**.steps') AS steps, LATERAL jsonb_each(steps) AS s(key,value)
WHERE ad.type IN ('page-build-handler','page-content-writer') AND ad.is_active
  AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' IN ('plan_sections','load_page_sections_from_spec')
ORDER BY 1;"
```

Read `plan_sections.config.sections`:
- `page-build-handler` → `spec_sections.sections` (from `site_specs.site_plan`, **authoritative**)
- `page-content-writer` → `input_data.current_page.sections` (**the caller's blob**)

**Gotcha (carried from the 194 lane):** use `jsonb_path_query(…, '$.**.steps')`. A top-level
`jsonb_each(default_config->'workflow'->'steps')` misses every step nested in a loop's
`sub_workflow`, and reads as "the agent doesn't have that step".

## R3 — does a handler actually cope with ALREADY-BUILT pages?

The question 201 §1 asked. Answer it from outcomes, not from reading the workflow.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT wi.item_type, wi.status,
       (EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=wi.page_id))::text AS page_already_built,
       count(*)::text
FROM site_work_items wi WHERE wi.handler_agent='<handler>' AND wi.page_id IS NOT NULL
GROUP BY 1,2,3 ORDER BY 4::int DESC;"
```

Measured 2026-08-05 — `page-build-handler`: `content_rewrite|complete|true|19`,
`empty_section|complete|true|12`, `empty_internal_href|complete|true|1`.
`page-content-writer`: 12 failed, 1 complete (which wrote nothing — 201 symptom 2), 1 triaged.

**Gotcha:** `page_already_built` must be computed from `page_components`, not from
`pages.build_status` or `deployed_at` — those are HISTORY columns (existing landmine) and a
retired page still reads "deployed".

## R4 — prove the handler guard DISCRIMINATES (do not trust a green test)

`handler_coverage_test.go` scans source literals. A passing test is equally consistent with
"the scanner never saw your line".

```bash
sed -i 's|HandlerAgent: "page-build-handler"|HandlerAgent: "bogus-nonexistent-handler"|' \
  platform/orchestration/actions/discovery_checks/check_literal_markdown.go
go test ./platform/orchestration/actions/discovery_checks/    # MUST FAIL
# then restore, and re-run: MUST pass
```

Expected failure: `TestEveryCheckHandlerAgentExistsOrIsADeclaredGap … routes work at handler
agent "bogus-nonexistent-handler", which is not a known agent and is not a declared gap.`

**Gotcha:** the new value must be inside `knownHandlerAgents` (it is), or the guard passes you
for the wrong reason. Check the map, not just the test result.

## R5 — the shared-tree build check (the one that counts)

The working tree is shared and may not compile because of another session's WIP. Build against
committed HEAD plus only your files.

```bash
SP=<scratchpad>/archcheck; rm -rf $SP && mkdir -p $SP && git archive HEAD | tar -x -C $SP
for f in check_literal_markdown.go check_placeholder_contact.go check_component_standards.go verifier_coverage_test.go; do
  cp platform/orchestration/actions/discovery_checks/$f $SP/platform/orchestration/actions/discovery_checks/$f; done
cd $SP && go build ./platform/... && go test ./platform/orchestration/actions/discovery_checks/
```

## R6 — ⚠ verifying after the roll: THREE ways to get a false result

The fix is inert until a chassis roll (`v1.0.1252` predates it).

1. **The 14 existing rows still carry the OLD `handler_agent`.** The code change only affects
   **newly filed** items. A re-arm that resets `status`/`attempt_count` but not `handler_agent`
   re-runs the broken route and looks like the fix failed.
2. **The one `triaged` item is on a LOCKED site.** `mortgagecalculator.co.uk`,
   `sites.locked_at` 2026-08-03, adoption lane. `load_work_items` returns
   `{items: [], count: 0, skipped_reason: "site_locked"}` — **success with zero items**, which
   looks exactly like an idle site. Check the lock first:
   `SELECT domain, locked_at, locked_by FROM sites WHERE domain='<d>';`
3. **`complete` is not proof.** 201 symptom 2 is an item that reached `complete` having written
   nothing, and `mark_complete` still trusts `handler_result` blindly. **Require the slot's
   `content_data` to change** (`updated_at` moves, the markdown string is gone), per
   `bugs_open/097`: check the artefact, not the status.
4. **⚠ EXPECT THE SECTION'S PRIOR PROSE TO BE GONE — that is not the fix failing.**
   `LANDMINES.md:4433`: `page-build-handler`'s writer sees no stored prose unless
   `spec.mode="recreate"`, which these checks do not set, so it rewrites the slot from scratch.
   A verifier who reads "the heading changed and the copy is shorter" as a regression will
   mis-report a working fix. **Do NOT "fix" this by setting `mode=recreate`** — that sources
   the original adoption crawl, not current content. See PLAN's corrected trade-off.

## R7 — the post-deploy pod-grep (added at the council's `debug_historian` objection)

The council was right that R5/R6 tested BEHAVIOUR and local build, never DEPLOYMENT. A routing
string compiled into the chassis binary must be proven in the running pod — never from git,
never from the image tag (a same-tag rebuild ships the node's stale binary).

```bash
for POD in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name); do
  echo "== $POD"
  # POSITIVE: the new routing comment literal is long enough to survive as a string.
  kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c \
    "grep -ac 'NOT page-content-writer' /app/agent-chassis"        # expect >0
  # DISCRIMINATING NEGATIVE: proves the grep can return 0 on this binary.
  kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c \
    "grep -ac 'NOT page-content-writer-v2' /app/agent-chassis"     # expect 0
done
```

**Gotchas:** `grep -ac`, not `strings | grep` — some images have no `strings`. **Every replica,
same exec** — a roll can leave one behind, and `logs deploy/X` reads one pod of N. A short
literal can compile to an immediate and grep 0 on a binary that fully supports it, which is why
the positive control is a long string. And the positive alone only proves the pipeline; the
near-miss negative is what proves the grep discriminates.

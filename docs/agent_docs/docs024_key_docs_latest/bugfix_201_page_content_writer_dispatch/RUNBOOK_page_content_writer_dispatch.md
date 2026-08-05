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

## R7 — ⚠ THIS CHANGE CANNOT BE PROVEN BY A POD-GREP. Do not try; you will read a false negative.

> **CORRECTED 2026-08-05 20:50Z, by running it on the fresh build and watching it fail.**
> The first version of R7 told you to `grep -ac 'NOT page-content-writer'`. **That string is a
> Go COMMENT. Comments are not compiled into a binary.** Run against `v1.0.1254` — a build that
> post-dates the fix by eight hours — it returned **0** on the first replica. Read as written,
> R7 said "the fix did not ship". It had shipped.

**Why no grep can work here, and this generalises.** The change swaps one *pre-existing* string
literal for another: `HandlerAgent: "page-content-writer"` → `"page-build-handler"`. Both
literals were already in the binary before the change and are both still in it after —
`page-build-handler` from a dozen other call sites, `page-content-writer` because
`page-build-handler` itself spawns it. **The edit introduces no new string and removes none**,
so there is no positive control and no negative control to construct. Go also interns identical
literals, so occurrence *counts* are not a substitute.

**The general rule: a pod-grep verifies a change that ADDS or REMOVES a string. A change that
re-points one existing literal at another existing literal is invisible to it.** Reach for the
behavioural check instead — and say which one you used, because "verified against the pod" is
a claim about a method that did not apply here.

### R7b — the check that DOES prove it: newly-filed items carry the new handler

The routing only affects **items created after the roll**. Existing rows keep the old value
(R6 trap 1). So the proof is the first item each check files on the new binary.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -F'|' -c "
SELECT wi.item_type, wi.handler_agent, count(*), min(wi.created_at), max(wi.created_at)
FROM site_work_items wi
WHERE wi.item_type IN ('literal_markdown','placeholder_contact')
  AND wi.created_at > '<the roll timestamp>'
GROUP BY 1,2 ORDER BY 3 DESC;"
```

- **PASS:** rows exist and `handler_agent = 'page-build-handler'`.
- **FAIL:** rows exist and `handler_agent = 'page-content-writer'` — the binary predates the fix.
- **NOT YET EVIDENCE:** *zero rows*. The checks simply have not fired. **A zero here is not a
  pass**, and this is the same trap the 194 lane hit twice in one day.

**When will they fire?** Both live on `quality-discovery-agent` (its `checks` array:
`broken_nav_links, placeholder_contact, generic_theme, unverified_claims, voice_tells,
literal_markdown`). It has **22 runs** all-time, last at **2026-08-05 12:14Z**, i.e. before the
`v1.0.1254` roll at 20:41Z. So wait for its next run rather than forcing one — and note that
`literal_markdown` only files an item on a page that *has* the defect, so a clean sweep also
produces zero rows. To get proof faster, target a site with a known live instance
(`mortgagecalculator.co.uk` hero, `gaswholesalers.com` pricing, `webdesign.co.uk` news-listing
per `bugs_open/184`) — but read R6 trap 2 first: **mortgagecalculator is LOCKED.**

⚠ **`check_component_standards`' `needs_content_page` sub-check may never produce a row to
check.** All 77 `needs_content_page` items fleet-wide already carry `page-build-handler` —
they come from `write_build_items`, not from this sub-check, which appears never to have fired.
The council's `editquality` seat made exactly this point. That edit is a consistency fix and is
**unverifiable from runtime data**; say so rather than claiming it passed.

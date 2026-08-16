# RUNBOOK — bugfix 285 (section-list assembler is lock-blind)

`PSQL='kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -Atc'`

## C1 — fleet census: locked rows vs the plan list that serves their page
```sql
WITH locked AS (
  SELECT pc.id, pc.page_id, p.site_id, p.name AS page_name, pc.slot_name, pc.position, cc.function, cc.component_level, p.sections::text AS cache
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id
  WHERE pc.lock_type IS NOT NULL AND pc.build_status <> 'removed')
SELECT s.domain||'/'||l.page_name, l.slot_name, l.function, l.component_level, l.position,
  (SELECT string_agg(sps.component_name, ',' ORDER BY sps.ordering) FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.site_id=l.site_id AND sp.is_current AND sps.page_name=l.page_name) AS plan_list,
  (SELECT count(*) FROM site_specs ss WHERE ss.site_id=l.site_id AND ss.aspect='site_plan' AND ss.is_current) AS has_aspect,
  l.cache
FROM locked l JOIN sites s ON s.id=l.site_id ORDER BY 1,5;
```
Gotcha: a bare `plan.component_name = slot_name` join under-matches positional slots (`tool-2`); read the lists.

## C2 — the symptom counter (remove-blocked items)
```sql
SELECT id, item_key, status, created_at, spec->>'locked_by' FROM site_work_items
WHERE item_type='lock_blocked_change' AND spec->>'blocked_action'='remove' ORDER BY created_at DESC;
```
## C3 — which run produced one (page-build-handler, tier that served, list it saw)
```sql
SELECT correlation_id, owner_agent_type, status, created_at,
       collected_data->'spec_sections'->>'source', collected_data->'spec_sections'->>'sections',
       collected_data->'spec_sections'->>'locked_sections_merged'   -- present only after this fix
FROM orchestration_states WHERE collected_data->'page_record'->>'name'='<page>' ORDER BY created_at DESC LIMIT 5;
```
Gotcha: `orchestration_states` has no `id`/`agent_type` columns — `orchestration_id`, `owner_agent_type`.

## C4 — the always-true guard, proven
```sql
SELECT sections::text IS DISTINCT FROM '["hero","contact-info"]', sections IS DISTINCT FROM '["hero","contact-info"]'::jsonb FROM pages WHERE id='4ff10911-ede0-4ba2-943b-547f66859cac';  -- t | f
```
## C5 — post-roll acceptance (owner's five, bug file §How to verify)
1. stamp: `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 | grep -m1 'build provenance'` (may have scrolled → binary probe per CLAUDE.md), then `git merge-base --is-ancestor <fix-commit> <stamp>`.
2. drive ONE page-build-handler pass over webdesign.uk contact: work-item insert recipe in
   `scripts/initial_messages/001_assemble_all_pages_rerender/081c_rerun_by_work_item.md`
   (`pipeline='build'`, `item_type='needs_content_page'`, `handler_agent='page-build-handler'`, `status='triaged'`, spec page_name=contact). NOT `run_improvement_sweep_once.sh` (promotes every detected item on the site).
3. assert: C3 shows `locked_sections_merged=["chat-input-box"]` and sections incl. it; `pages.sections` incl. it; locked row `md5(rendered_html)`,`updated_at` unchanged; hero/contact-info rows re-inserted (new `created_at`); no new `remove` item (C2); `section_source_drift` files nothing for contact.

## C6 — is it LIVE? (the `build provenance` line scrolls; this does not)
```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in locked_sections_merged LOCKED_MERGE_SKIPPED load_page_sections_from_spec deadbeefcontrolstring; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe && echo "PRESENT $s" || echo "absent  $s"
done
```
Expect round 1 → `locked_sections_merged` PRESENT; round 2 → `LOCKED_MERGE_SKIPPED` PRESENT.
**Both controls are the point**: `load_page_sections_from_spec` must be PRESENT (the grep can see
this binary) and `deadbeefcontrolstring` must be absent (it is not matching everything). Never
`strings` (absent from the image), never a discovery grep for "some 40-hex string" (matches Go's
digit table). Pod start vs commit time is the sanity check: `kubectl get pods -l app=agent-chassis
-o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime`.

## C7 — has the merge actually FIRED? (the data-side corroboration, and the demand control)
```sql
SELECT correlation_id, collected_data->'page_record'->>'name' AS page,
       collected_data->'spec_sections'->>'source' AS src,
       collected_data->'spec_sections'->>'locked_sections_merged' AS merged,
       collected_data->'spec_sections'->>'locked_merge_count' AS n, created_at
FROM orchestration_states
WHERE owner_agent_type='page-build-handler' AND created_at > '<roll time>'
  AND collected_data ? 'spec_sections' ORDER BY created_at DESC LIMIT 20;
```
The KEYS being present at all proves the new binary is running (no earlier one emits them).
`n = 0` on every row means the merge has not fired — those pages have no locked rows.
⚠ **Before quoting "0 new `remove` items" (C2) as success, run the demand control:**
```sql
SELECT count(*) FROM orchestration_states WHERE owner_agent_type='page-build-handler'
  AND created_at > '<roll time>' AND (collected_data->'spec_sections'->>'locked_merge_count')::int > 0;
```
0 here means no locked page has rebuilt, so the C2 zero carries no information.

## C8 — who else could re-open the class (the council's census question)
```bash
grep -rn "UPDATE pages" --include=*.go platform/ internal/ cmd/ | grep -i sections | grep -v _test.go
grep -rn "INSERT INTO pages" --include=*.go platform/ internal/ cmd/ | grep -v _test.go
```
```sql
SELECT owner_agent_type, count(*) AS runs_30d,
       count(*) FILTER (WHERE collected_data ? 'spec_sections') AS with_loader
FROM orchestration_states WHERE created_at > now() - interval '30 days'
  AND owner_agent_type IN ('page-build-handler','page-rebuild','pageflow-builder','page-rerender',
                           'site-work-orchestrator','tool-recreation-handler')
GROUP BY 1 ORDER BY 2 DESC;
```
Gotcha: "who calls the ACTION" and "who writes `pages.sections`" are different questions, and the
second is the one the bug_historian seat asks. A caller with `with_loader = 0` is only dangerous
if it composes from the LIST — `page-rerender` composes from the stored ROWS, so it cannot drop a
locked row (read `loadStoredSections`/`carryStoredSection` before assuming otherwise).

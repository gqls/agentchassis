# RUNBOOK — bugfix 311 (commands that were hard to get right, with gotchas)

## Class counts (re-verify the bug's premise)

```sql
SELECT 'null_section_type', count(*) FROM content_components
 WHERE component_level='section' AND is_active AND forked_from IS NULL AND section_type IS NULL
UNION ALL SELECT 'eq_function', count(*) FROM content_components
 WHERE component_level='section' AND is_active AND forked_from IS NULL AND section_type = function
UNION ALL SELECT 'diff_function', count(*) FROM content_components
 WHERE component_level='section' AND is_active AND forked_from IS NULL
   AND section_type IS NOT NULL AND section_type <> function;
```
2026-08-19 baseline: 26 / 89 / 26.

## Is a component in the 311 trap? (three questions, not two)

```sql
-- 1. do the two keys agree, and who is the row?
SELECT id, name, function, section_type, component_level, is_active, created_from
FROM content_components WHERE function='<name>' OR section_type='<name>';
-- 2. who depends on it? (content_components has NO site_id)
SELECT DISTINCT s.domain FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id='<id>';
-- 3. would the LOADER even serve it? (the refinement — this is what hides the incumbents)
SELECT length(html_template) >= 100 AND html_template NOT LIKE '%</section>%' AS guard_dropped
FROM content_components WHERE id='<id>';
```
Gotcha: `sectionTemplateValid` needs `</section>` only for `component_level='section'`;
tool-level rows use the markup-balance check instead (CLC-019).

## The failing work items

```sql
SELECT id, item_type, summary, status, created_at FROM site_work_items
WHERE item_type IN ('needs_new_component','needs_section_data')
  AND status NOT IN ('complete','cancelled','rejected')
ORDER BY created_at DESC LIMIT 20;
```

## Diagnosis intake (refinement run)

Intake `1306e72c-c725-4c3b-b0c3-8a63137f35fb`, run corr `f1433782-6ba7-4304-a7f9-8bd830dfb7c9`.
Failed 2026-08-19 10:32 UTC on the fleet Anthropic API usage cap
("You have reached your specified API usage limits", status 400, `__step_error`);
intake reset to `triaged` — the dispatch loop re-claims it when the API recovers.
Gotcha: the run FAILURE surfaces as a COMPLETED wrapper row plus FAILED step rows —
read `collected_data->>'__step_error'`, the `error` column, never the status alone.

## Council round

Submission corr **`fc3ac5f4-ee3a-4e27-88ab-a8b2536b2c1d`** (2026-08-19). Find the run by payload:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'fc3ac5f4-ee3a-4e27-88ab-a8b2536b2c1d';
-- verdict:
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='fc3ac5f4-ee3a-4e27-88ab-a8b2536b2c1d' AND kind='council_report' ORDER BY created_at;
```
Budget ~30 min publish→run under normal load; while the API cap holds, expect the same
usage-limit failure — resubmit with `RESUBMIT_CORR=fc3ac5f4-…` only if the ROUND ran and
returned REVISE; a cap-failed round re-fires without a resubmit (check before re-firing:
a duplicate round costs credits).

## Tests + the mutation proof

```bash
go test ./platform/orchestration/actions/ -run "ForeignCollision|OwnSiteCollision|UnknownRequester|ResolveStorageIdentity" -count=1
```
Mutation re-run (proves the foreign-collision test still bites): replace the
`if len(foreignIDs) == 0` branch in `resolveStorageIdentity` with `if true` — the test
must FAIL with an uncovered `UPDATE content_components` (2026-08-19: it did). Restore.

## Post-roll verification (BOTH halves, after an image ships this)

```bash
# 1. prove the binary (per SERVICE; provenance log line scrolls — probe with controls):
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<this-commit-sha>" /proc/1/exe && echo present
kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<high-entropy-fake-sha>" /proc/1/exe || echo control-clean
```
```sql
-- 2. baseline the incumbent BEFORE re-driving (pin values, not HEAD refs):
SELECT md5(html_template), md5(input_schema::text) FROM content_components
WHERE id='824e3309-f90c-4aa9-b679-46f4a8722475';
```
Then re-drive one failed loanzy item (fresh needs_new_component via a page rebuild, or
reset item `7a2219bc` per the dispatcher's retry rules) and assert:
```sql
-- new scoped base row exists:
SELECT id, function, section_type, forked_from, is_active FROM content_components
WHERE function='loans-credit-health-check-loanzy-uk';
-- incumbent untouched (md5s equal the baseline);
-- the diversion is recorded:
SELECT created_at, error_message FROM agent_error_log
WHERE error_code='COMPONENT_COLLISION_DIVERTED' ORDER BY created_at DESC LIMIT 3;
-- the page links a component after rebuild:
SELECT pc.component_id, pc.build_status FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='loanzy.uk' AND p.name='tool-credit-health-check';
```
Gotcha: `agent_error_log`'s column names — check `\d agent_error_log` before trusting the
query above; the finding writer is `LogActionFindings` and the code is in `Context` too.

## Post-roll verification — TOOL half (RFC_036 §9.3, commit e24bc9c0f, council ceae30f2)

Same discipline as the section half — per SERVICE, controls both ways:
```bash
# Find the stamp by probing CANDIDATE full shas (commits in the build window) —
# NEVER a discovery-grep for "some 40-hex string" (LANDMINE: matches Go's digit
# table on every service). The stamp is the BUILD commit, not your commit:
for sha in <candidate-full-shas>; do kubectl -n ai-persona-system exec <pod> -- grep -aq "$sha" /proc/1/exe && echo "STAMP=$sha"; done
git merge-base --is-ancestor e24bc9c0f <stamp>   # TRUE = the fork fix is in the binary
kubectl -n ai-persona-system exec <pod> -- grep -aq "library tool claims this function" /proc/1/exe && echo literal-present
kubectl -n ai-persona-system exec <pod> -- grep -aq "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef" /proc/1/exe || echo control-clean
```
Demand test (the webdesign lane's 2 parked tools are the natural cases):
re-drive `tool-ab-test-calculator`'s generation; assert the new row has
`forked_from = <library row id>` and `component_level='tool'`, save_tool COMPLETES (no
SQLSTATE 23505), and the library row is untouched. Baseline the library row's md5 BEFORE,
as with the section half.

## The re-drive that WORKED (2026-08-19, section half, loanzy car-finance) — two items, not one

`needs_rebuild` on the page is NOT enough (no consumer; `page-rebuild` never runs). Two hand-filed
items, both mirroring the code's own shape so dispatch treats them as native:
```sql
-- 1. the component (mirrors CreateNeedsNewComponentItem; idx_swi_dedup ignores 'failed' so a
--    fresh row is allowed; created_by is not a dispatch filter; keep source=component_selector
--    because it becomes component_versions.change_source)
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('<site_id>', 'component_selector', 'build', 'needs_new_component', 'medium',
 'Need component template for section type: <section>',
 '{"site_type":"","description":"Component for section type \"<section>\" on page \"<page>\" ( site)","page_context":"<page>","section_type":"<section>","design_direction":"modern-light"}'::jsonb,
 50, 'component-creator', 'triaged', 'bugfix_311_redrive', 'needs_new_component:<section>') RETURNING id;
-- wait for status=complete, then read result->'response'->'stored_component'->>'diverted_from_component_id'
-- 2. the page (mirrors flag_page_image_rebuild; previous page_rerender:<page> must be terminal)
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key, batch_id)
VALUES ('<site_id>', 'bugfix_311_redrive', 'build', 'needs_page', 'medium', 'Re-render <page> to link the diverted component',
 '{"reason":"bugfix_311_redrive_relink","page_name":"<page>"}'::jsonb, 99, 'page-build-handler', 'triaged', 'bugfix_311_redrive', 'page_rerender:<page>', gen_random_uuid()) RETURNING id;
```
Gotchas: the trigger dispatches ONE unlocked site per 60s tick, oldest item first — check
`find_dispatchable_site`'s order if it does not claim within ~2 min; a site with any `claimed`
item is skipped until it clears. Assert at the artefact: `page_components` row for the new id
with build_status deployed, AND `curl` the tool URL and count `<input` (tags span lines — use
`grep -c '<input'`, not a one-line regex). Pin the served page's md5 and byte count BEFORE.
~~Loanzy's six remaining: settlement / overpayment / standard-calc / compare-loans /
interest-rate-stress-test (collision class) and credit-health-check (fails UPSTREAM on
max_tokens 16000 — a different defect; do not spend attempts on it until that is raised).~~

> **CORRECTED 2026-08-20 08:16Z — the list was three ways wrong, and each error changes what a
> page needs.** Measured by fetching every loanzy tool page and cross-reading the eight `failed`
> items: **FIVE** are collision class (settlement · overpayment · standard-calc → page
> `tool-loan-repayment-calculator` · compare-loans → page `tool-loan-comparison-calculator` ·
> interest-rate-stress-test) — car-finance was already repaired. **TWO** pages die on the
> `max_tokens` cap, not one: `tool-credit-health-check` **and `tool-eligibility-checker`**, which
> planned the same `loans-credit-health-check` section. **ONE** needs only the render leg —
> `tool-loan-vs-savings` has a good component (a plain creation, no collision) and serves 0
> `<input>` purely because `needs_rebuild` has no consumer; a re-drive on it would be wasted
> spend. And `tool-compare-loans` / `tool-is-a-loan-right-for-me` have **zero `page_components`
> rows** — they never planned a section, so no re-drive can help them.

## 2026-08-19 20:35Z — the roll that carried the TOOL half: v1.0.1316, stamp `07eeba4a1eecbe809f518b5d0b7f9fc5f75e71ed`
Both replicas (`agent-chassis-5ddd9744-*`, started 17:13Z). Re-run of the recipe above: stamp
PRESENT, fake sha ABSENT, `e24bc9c0f` and `17d883333` both ancestors, both literals present.
Gotcha that bit this lane: **re-run `kubectl get pods` before repeating a version** — a pod list
is a snapshot; I carried "v1.0.1315" four hours past the roll.

## Filing the re-drive as a BATCH (2026-08-20) — one statement, and the two pre-checks that make it safe

The single-item recipe above scales badly by hand (five near-identical INSERTs is five chances to
mistype a spec). Drive it from a `VALUES` list and let `format()` build the description, so the
shape cannot drift between rows:

```sql
WITH targets(section, page) AS (VALUES
  ('loans-interest-rate-stress-test','tool-interest-rate-stress-test'),
  ('loans-compare-loans','tool-loan-comparison-calculator')      -- section name ≠ page name
)
INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
SELECT '<site_id>', 'component_selector', 'build', 'needs_new_component', 'medium',
       'Need component template for section type: ' || section,
       jsonb_build_object('site_type','', 'description',
         format('Component for section type "%s" on page "%s" ( site)', section, page),
         'page_context', page, 'section_type', section, 'design_direction', 'modern-light'),
       50, 'component-creator', 'triaged', 'bugfix_311_redrive', 'needs_new_component:' || section
FROM targets RETURNING id, spec->>'section_type', spec->>'page_context';
```
Note the description's literal `" ( site)"` — the empty `site_type` leaves that gap in the
production writer's own output. Mirror it; do not tidy it.

**Pre-check 1 — READ the dedup index, do not recall it.** Fresh inserts are only safe because
`idx_swi_dedup` is partial:
```sql
SELECT indexdef FROM pg_indexes WHERE indexname='idx_swi_dedup';
-- excluded: complete, verified, rejected, wont_fix, failed, unresolved, cancelled
```
So a `failed` `needs_new_component:<section>` and a `complete` `page_rerender:<page>` both permit a
new row. If a key is held by anything else (`needs_human_review`, `triaged`, `claimed`) the insert
raises 23505 — check first:
```sql
SELECT item_key, item_type, status FROM site_work_items swi JOIN sites s ON s.id=swi.site_id
WHERE s.domain='<domain>' AND (swi.item_type='needs_new_component' OR swi.item_key LIKE 'page_rerender:%');
```

**Pre-check 2 — re-pin the incumbent md5s IMMEDIATELY before the run, and attribute any drift.**
A day-old pin is not a baseline on a shared row. On 2026-08-20 the `loans-standard-calc` incumbent
`b420389f` had moved overnight (html `85d67379…` → `a9dea7cd…`, 2,469 → 2,852 chars, `updated_at`
07:02:57Z) — **nothing to do with this lane**:
```sql
SELECT version_number, change_source, created_at, md5(html_template)
FROM component_versions WHERE component_id='<id>' ORDER BY version_number DESC;
-- one row: change_source='scope_component_instance_judged', and its bytes ARE the old md5
```
`component_versions` archives the PREVIOUS bytes, so it both dates the write and names the writer.
**Read it before concluding your own run damaged an incumbent** — the "all incumbents byte-identical"
assertion is otherwise a coin flip against every other mechanism that touches shared components.

**Pre-check 3 — a served-page baseline needs TWO reads.** On 2026-08-20 the first fetch of
`/tools/loan-comparison-calculator/index.html` returned **404**; two further reads (same URL from
`pages.url`, nothing deployed in between) returned **200 / 22,600 bytes / 0 `<input>`**. Had the
404 been kept as the "before", the re-render would have been credited with publishing a page it
never touched. Read the before twice; record the reproducible one.

## The RE-RENDER leg is not free — two ways it fails that have nothing to do with 311 (2026-08-20)

**1. The `253` section-component floor can refuse the whole save because of an UNRELATED slot.**
First of the five 09:06Z re-renders came back:
```
step save_sections failed: save_page_sections: SECTION COMPONENT FLOOR REFUSED for page
"tool-interest-rate-stress-test" — hero-tool 12→5 class attributes (42% kept, floor 50%).
A same-named slot may not lose more than 50% of the elements carrying layout classes in one
save, even when its TEXT survives; if this flattening is intended, set section_component_floor
in the step config. Nothing was written (bugs_open/253, framework_rewrite_of_a_prose_block).
```
Read it carefully: the refusal is about **`hero-tool`**, not about the calculator you are trying to
land. A `page_rerender` regenerates *every* section, so any slot the content writer happens to
flatten this time takes the whole page's save down with it — **and nothing is written, so the page
is exactly as it was.** No damage, no partial state; you simply do not get your component.
Do NOT reach for `section_component_floor` in the step config to get past it: that is a shared
step's config on a fleet seam, and the guard exists because a framework rewrite once stripped every
layout component from a prose block (`bugs_closed/253_…framework_rewrite_of_a_prose_block…`).

**2. A refused page build parks the item TERMINALLY at attempt 1 of 3 — it will not retry itself.**
Read, not assumed: `page-build-handler`'s `mark_item_failed` step is
`{"action":"update_work_item_status","config":{"status":"failed",…}}`, so a real step error writes
`status='failed'` outright. `ClaimWorkItemAction` only takes `status IN ('triaged','approved')`
(`claim_work_item_action.go:97-105`), so the item's remaining `max_attempts` are unreachable —
`attempt_count=1, max_attempts=3, status=failed` is a **stopped** item, not a queued one. (Contrast
`FailWorkItemAction`, `load_work_item_actions.go:1147-1158`, which WOULD have returned it to
`triaged`; this handler does not route through that arm.) The step's own description says
"attempt counted", which reads as though a retry follows. It does not.
**So: to try again you file a FRESH `needs_page` item** — the `failed` status is excluded from
`idx_swi_dedup`, so the same `page_rerender:<page>` key is available immediately. Whether a retry
is worth it depends on why it failed. ~~a floor refusal is content-writer variance and may well
pass on a second run, while a deterministic refusal will not.~~
> **CORRECTED 2026-08-20 14:05Z — I retried it and that guess was wrong, in a way my own §9 entry
> written hours earlier would have caught.** The fresh item (`84e586b9`) failed at the same step
> with figures identical to the first: **`hero-tool 12→5 class attributes (42% kept, floor 50%)`**,
> both runs. Identical numbers across independent runs is the deterministic-refusal tell
> (`016b` §9, added this same session from `bugs_open/337`) — so a floor refusal is NOT
> content-writer variance to be retried; it is a stable property of what the writer produces for
> that page from those inputs. **Stop after the first one and read the guard's own remedy** instead
> of spending a second build:
>
> `save_sections_component_floor.go:225-231` states it — *"give it the component vocabulary in
> `content_direction` rather than lowering the floor, which is what fixed the motivating page"* —
> and marks `section_component_floor=0` as "the deliberate escape hatch, not a fix". The floor is a
> **step-config** key (`sectionComponentFloorKey`, default 0.5, slots under 10 class attributes out
> of scope), so there is no per-item override to reach for: the choice is a content-direction fix on
> the site, or leaving the page alone. On another lane's site, that is their call.

## ⚠ PRE-FLIGHT CHECK THE RECIPE ABOVE WAS MISSING: `pages.status`, not just `build_status` (2026-08-20, cost one generation + one page build)

`tool-loan-repayment-calculator` was re-driven and rebuilt end to end — the component diverted, the
page build reported **complete on attempt 0**, four slots rendered including the new calculator at
14,991 chars — and the page still **404s**, because it is **`pages.status='archived'`**. The
archived-page guard did exactly its job and refused the deploy stamp at the very last step
(`agent_error_log`: two `ARCHIVED_PAGE_DEPLOY_REFUSED` rows, 09:21:00–09:21:01Z), so `build_status`
stayed `planned` and `deployed_at` stayed NULL. **The guard is not the problem; my pre-flight was.**
It also demonstrates `bugs_open/266` live: every producer up to the final stamp happily did full
work on an archived page.

**So the pre-flight is:**
```sql
SELECT name, status, build_status, jsonb_array_length(coalesce(sections,'[]'::jsonb)) AS planned,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id) AS slots
FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='<domain>' AND p.name LIKE 'tool-%'
ORDER BY status, name;   -- status='archived' ⇒ do NOT spend a re-drive on it
```
On loanzy that returns **four archived tool pages** — `tool-compare-loans`,
`tool-eligibility-checker`, `tool-is-a-loan-right-for-me`, `tool-loan-repayment-calculator`.

> **This also corrects the corrected list above.** The 08:16Z entry said `tool-compare-loans` and
> `tool-is-a-loan-right-for-me` "never planned a section at all", inferred from their having **zero
> `page_components` rows** and 404ing. Wrong on both counts: they have **5 and 4 planned sections**
> in `pages.sections`, and the 404 is the archival, not an absence of a plan. **A 404 plus zero
> slots does not distinguish "never planned" from "planned and archived" — `pages.status` does, and
> it is one column.** Unarchiving another lane's page is their call, not a repair to take here.

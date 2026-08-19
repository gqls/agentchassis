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

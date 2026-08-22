# RUNBOOK — contrast_ratio check lane

## Measure the live pages (the validity/regression check, re-runnable)

```bash
cd <scratchpad>   # session scratchpad; script also archived in NOTES
/home/ant/.venvs/vonc_pw/bin/python3 check_131_validity.py
```
Gotchas that cost time on 2026-08-22:
- system `python3` has no playwright — the venv is `~/.venvs/vonc_pw` (found via the gauntlet
  RUNBOOK; playwright browsers live in `~/.cache/ms-playwright`).
- the gauntlet page's `.gi-challenge-text` is hidden until the sealed page is revealed — the reveal
  starts a REAL round against the live engine (the gauntlet lane's own verify scripts do the same;
  use a `?cb=` cache-buster).
- the scan composites translucent backgrounds naively — treat sub-2.0 hits as the invisible class
  and verify by screenshot before quoting them (`curl-audit-has-no-opinion-about-rendering`).

## Enumerate consumers of the check type (the RFC_022 query — rerun before any scope claim)

```sql
-- fences using it as a CHECK TYPE (the discriminating query; bare LIKE matches tool prose)
SELECT count(*) FROM doc_plans WHERE body ~ '"type":\s*"contrast_ratio"';
SELECT count(*) FROM agent_definitions WHERE default_config::text LIKE '%contrast_ratio%' AND is_active;
```
```bash
grep -rn "contrast_ratio" docs/agent_docs/sql_for_agents/*.sql | wc -l
```
All three were 0 on 2026-08-22 (before this lane's build).

## Run the new check's tests

```bash
cd /home/ant/projects/agentchassis
go test ./internal/adapters/browserrunner/ -run 'Contrast' -count=1
go test ./internal/adapters/browserrunner/ ./platform/orchestration/actions/ -count=1   # lockstep suites live here
```

## Prove the DEPLOYED adapter knows the type (Phase 2, after the image rolls)

`strings` DOES NOT EXIST in the browser-runner container, and short literals never reach rodata
(LANDMINES:512). Use the long detail sentence the new arm emits, with controls, in ONE exec:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=browser-runner-adapter -o name | head -1)
kubectl -n ai-persona-system exec ${POD#pod/} -- sh -c '
  grep -ac "text is painted below its contrast threshold" /app/browser-runner-adapter;
  grep -ac "but a parent CLIPS it" /app/browser-runner-adapter;
  grep -ac "zzz_not_a_real_marker_zzz" /app/browser-runner-adapter'
```
Expect `1 / 1 / 0` (new marker / positive control / negative control). All three must move together.

## Witness (Phase 2): the check must FAIL a page that IS bad

Known-bad target measured 2026-08-22: `https://vonc.com/tools/gauntlet/index.html` — `div.gi-eyebrow`
1.66:1, `div.gi-rules-label` 1.76:1 (both firm, both screenshot-confirmed). Fire a manual acceptance
work item carrying a `contrast_ratio` check (precedent: 131-B's `manual_131b_witness` item
`4e06c4ab-…`, which reused the 010b shape from `043bfe1d`). What proves it: `pass:false` **with**
`Culprit`/`CulpritSelector` populated and a ratio in the detail — nothing else produces that shape.
Run a clean page in the same batch as the control. Check the queue first:
`SELECT … FROM site_work_items WHERE status NOT IN ('complete','cancelled','rejected') AND <target>`.

## Council

Submission JSON + corr in this dir (`council_submission_2026-08-22.json`, corr in NOTES).
Find the run by payload, not printed id:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```
Budget ~30 min (dispatch queues behind the fleet). Verdict:
```sql
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

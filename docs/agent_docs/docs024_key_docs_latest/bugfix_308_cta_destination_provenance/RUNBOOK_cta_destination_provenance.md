# RUNBOOK — bugs_open/308, CTA destination provenance

Every command here was needed and is written with its gotcha attached. Change it HERE.

## DB access

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## 1. The population — findings naming a target the repairer cannot produce

**Gotcha: the `$` in the regex.** Passed through a double-quoted `-c "…"` the `$)` anchor is
expanded away by the shell before psql sees it, and the query then returns a DIFFERENT
number without erroring. Escape it (`\$`) or use a single-quoted heredoc.

```sql
SELECT count(*)
FROM site_work_items swi, LATERAL jsonb_array_elements(swi.spec->'findings') f
WHERE swi.item_key LIKE 'misdirected_cta:%'
  AND f->>'suggested_target' ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)';
```
2026-08-17 (filing): 149 · 2026-08-22: **200**

## 2. The churn, made visible — split the SAME query by status

This is the one worth quoting: `complete` means the repair ran and reported success.

```sql
SELECT swi.status, count(DISTINCT swi.id) AS items, count(*) AS findings
FROM site_work_items swi, LATERAL jsonb_array_elements(swi.spec->'findings') f
WHERE swi.item_key LIKE 'misdirected_cta:%'
  AND f->>'suggested_target' ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)'
GROUP BY 1 ORDER BY 3 DESC;
```

## 3. DEMAND CONTROL — run this BEFORE believing any count above

A post-fix number here is meaningless unless the detector is actually running. It is not.

```sql
SELECT date_trunc('day', created_at)::date AS day, count(*) AS all_misdirected_items
FROM site_work_items WHERE item_key LIKE 'misdirected_cta:%'
GROUP BY 1 ORDER BY 1 DESC LIMIT 12;
```
Last row 2026-08-19 = 3. **Zero detections for three days.** So: induce a scoped discovery
run first, then measure. A flat 200 after a fix proves nothing.

## 4. Is `suggested_target` read by anything?

```bash
grep -rn "suggested_target" --include=*.go platform/ internal/ pkg/
```
Three hits, all writers/tests. **No consumer.** Re-run after any fix claiming to close the
detector↔repairer gap: a *new* hit outside `discovery_checks/` is the positive signal.

## 5. Metadata-key namespace inside content_data (clean as of 2026-08-22)

```sql
SELECT k, count(*) FROM page_components pc, LATERAL jsonb_object_keys(pc.content_data) k
WHERE pc.content_data IS NOT NULL AND k LIKE '\_\_%' GROUP BY 1 ORDER BY 2 DESC;
```
→ 0 rows. **Gotcha: `_` is a LIKE wildcard** — unescaped, `'__%'` matches every key of two
or more characters and returns the whole schema, which reads as "the namespace is already
crowded". Escape both underscores.

## 6. Ownership / concurrency, re-run at every phase boundary

```bash
python3 scripts/who-owns.py 308           # commits only — blind to in-flight sessions
git status --short -- platform/orchestration/actions/resolve_internal_links_action.go \
  platform/orchestration/actions/rerender_page_sections_action.go \
  platform/orchestration/actions/discovery_checks/ platform/orchestration/datahelpers/
```
Both are LAGGING. The second is the one that sees a session mid-edit; it caught a stale
session-start snapshot today (`section_editor_actions.go` read dirty at session start and
clean twenty minutes later — another session had committed in between).

## 7. Next migration number

```bash
ls docs/agent_docs/sql_for_agents/ | grep -oE '^[0-9]+' | sort -n | tail -1
```
554 as of 2026-08-22. Another session may take the next number between your `ls` and your
write — re-check immediately before creating the file.

## 8. Shell gotcha that cost a file today

`cd <dir> && cat > file <<'EOF'` — the Bash tool's working directory **persists between
calls**. A second `cd <same relative path>` fails, the `&&` short-circuits, the heredoc is
consumed and discarded, and a trailing `echo ok` still prints `ok`. **Use absolute paths for
every write.** The tell is that `ok` appears alongside the `cd` error, which is easy to skim
past.

## 9. Council submission (Phase A)

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/COUNCIL_SUBMISSION_2026-08-22_308_phase_a_provenance.json
```
**Gotcha:** `operation` must be one of `modify|add|remove|config_change`. **`create` is rejected
— a new file is `add`.** The dry run catches it for free; it cost two edits here.

`SUBMISSION_CORR = e4336931-487b-4db3-b4dc-a4b128b3566c` (Phase A, submitted 2026-08-22).

**Read the verdict keyed on the CORRELATION, never `ORDER BY created_at DESC LIMIT 1`** — with
~40 sessions live, the newest `council-gate` note is whoever finished last, not your verdict
(a lane read another lane's REVISE as its own on 2026-08-22; LANDMINES carries it):

```sql
SELECT created_at, metadata->>'decision'
FROM diagnosis_artifacts
WHERE correlation_id='e4336931-487b-4db3-b4dc-a4b128b3566c' AND kind='council_report'
ORDER BY created_at;
```

Budget ~30 minutes, not ~2: the council itself takes 2-5 minutes but the dispatch queues behind
the fleet. A missing row is latency, not a dropped dispatch — do not retry on that evidence.


## 10. The plan-size cap is measured by the SERVER, and `DRY_RUN=1` does not see it

**Round 5 died at `persist_submission` with `plan too large: 65561 bytes (cap 65536)` — 25 bytes
over — after `DRY_RUN=1` had passed on the same file seconds earlier.** The dry run validates
locally with a different size computation; only the server's counts. The failure looks exactly like
the API-cap failure (`complete_invalid`, COMPLETED, no `council_report`), so read `__step_error`
before concluding anything:

```sql
SELECT left(collected_data->>'__step_error', 1200) FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>'
ORDER BY created_at DESC LIMIT 1;
```

**The server measures Go's `json.Marshal` of `plan`** — UTF-8 bytes, with `<`, `>` and `&` escaped
to `\u003c`/`\u003e`/`\u0026` (+5 bytes each). Python's `len(json.dumps(...))` is wrong in both
directions (`ensure_ascii=True` inflates every `—`; compact-vs-indent differs by hundreds). This
predictor was verified exact against the server's own number:

```python
def gosize(plan):
    s = json.dumps(plan, separators=(',',':'), ensure_ascii=False)
    return len(s.encode()) + sum(s.count(c)*5 for c in '<>&')
# 65561 predicted, 65561 reported. Gate on gosize(plan) <= 65536 - 500 before dispatching.
```

**Cost of getting it wrong is low but not zero:** the run fails BEFORE any seat is dispatched, so
no credits are spent — but it consumes a round and ~10 minutes, and the terminal state is
indistinguishable at a glance from the API cap that killed round 4.

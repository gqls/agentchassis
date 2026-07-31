# RUNBOOK — 164 bundle body cap

Every command here had a gotcha attached. The gotcha is the reason the line exists.

## Measure the cap's real rate (this is what the filing asked for first)

```sql
SELECT count(*) AS bundles,
       count(*) FILTER (WHERE (metadata->>'truncated')::bool) AS truncated,
       round(100.0*count(*) FILTER (WHERE (metadata->>'truncated')::bool)/count(*),1) AS pct,
       min(created_at)::date AS first, max(created_at)::date AS last
  FROM diagnosis_artifacts WHERE kind='bundle';
```

⚠ **Report it as a RATE and print the window.** `diagnosis_artifacts` is
retention-clocked (`bundle_retention_days` default **30**, `expires_at` set on
insert), so `count(*)` is "bundles still retained", never a census. Always select
`min/max(created_at)` in the same query — a figure quoted without its window goes
stale silently.

⚠ Do **not** use `length(body) >= 59000` as the truncation proxy (the filing suggests
it as a fallback). It is wrong in both directions: `body` is the WHOLE bundle
including runtime evidence and schema, so a fat untruncated bundle scores as
truncated, and the three worst real cases have `body_chars = 0` for in-scope code.
`metadata->>'truncated'` is the actual flag and was there all along.

## See how much scope each truncation destroyed

```sql
SELECT correlation_id, iteration, created_at::date AS d,
       (metadata->>'symbols_in_scope')::int AS in_scope,
       (metadata->>'symbol_count')::int      AS included,
       (metadata->>'symbols_in_scope')::int - (metadata->>'symbol_count')::int AS dropped,
       (metadata->>'body_chars')::int AS body_chars
  FROM diagnosis_artifacts
 WHERE kind='bundle' AND (metadata->>'truncated')::bool
 ORDER BY dropped DESC;
```

⚠ Pre-fix, `dropped` **conflates two different failures** — bodies that did not fit
and bodies that could not be read. That conflation is why the bug had to be filed
`[UNMEASURED]`. Post-fix the artefact carries them apart:

```sql
SELECT (metadata->>'symbols_omitted_size')::int AS too_big,
       (metadata->>'symbols_unreadable')::int   AS unreadable, count(*)
  FROM diagnosis_artifacts WHERE kind='bundle' GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠ Those two keys **only exist on bundles written after v1.0.12xx** — older rows
return NULL, not 0. Filter on `created_at` or use `COALESCE`, or you will report a
drop to zero that is really an absence of data.

## Read the artefact rather than trusting the counter

```sql
SELECT correlation_id, iteration,
       substring(body from position('## In-scope code' in body) for 220)
  FROM diagnosis_artifacts
 WHERE kind='bundle' AND (metadata->>'truncated')::bool
   AND (metadata->>'symbol_count')::int = 0 ORDER BY created_at LIMIT 3;
```

This is the query that turned "the counter says 0 included" into "the heading is
followed immediately by the next heading" — i.e. into evidence. **Trust the rendered
artefact, not the status.**

## Find every char-budget cap in the repo (the shape audit)

```bash
grep -rn --include=*.go -A4 -E '\+\s*len\([a-zA-Z_.()\[\]]+\)\s*>\s*[a-zA-Z_.]*(cap|Cap|max|Max|budget|limit|Limit)' \
  platform/ internal/ pkg/ | grep -E ':\s*(break|continue)|-\s*(break|continue)'
```

Returns exactly three sites, all in `diagnose_assemble_bundle_action.go` (`:208`,
`:521`, `:605`). ⚠ A narrower first attempt keyed on variable names
(`total|used|size|n|acc|sum`) found the same three but only by luck — **a grep
proves an absence only for the spelling it searches**, so run both spellings before
claiming a population.

## Test it, and prove the test can fail

```bash
gofmt -l platform/orchestration/actions/diagnose_assemble_bodycap_test.go
go test ./platform/orchestration/actions/ -run 'TestBundleBodyCap' -count=1
```

**Induce the failure — a passing test proves nothing until you have seen it fail:**

```bash
SP=<scratchpad>
cp platform/orchestration/actions/diagnose_assemble_bundle_action.go $SP/FIXED.bak
git show HEAD:platform/orchestration/actions/diagnose_assemble_bundle_action.go \
  > platform/orchestration/actions/diagnose_assemble_bundle_action.go
go test ./platform/orchestration/actions/ -run 'TestBundleBodyCap' -count=1   # expect 3 FAIL
cp $SP/FIXED.bak platform/orchestration/actions/diagnose_assemble_bundle_action.go
git diff --stat platform/orchestration/actions/diagnose_assemble_bundle_action.go  # confirm restored
```

⚠ The 4th test (the byte-identity control) **passes against both versions**. That is
correct and intended — it asserts the OLD behaviour is preserved. Do not "fix" it.
⚠ Restore from the backup and re-check `git diff --stat` before doing anything else;
this window is the only moment the tree holds the broken version, and other sessions
build from it.

## Verify against a clean HEAD, not the shared tree

```bash
SP=<scratchpad>; rm -rf $SP/archtest && mkdir -p $SP/archtest
git archive HEAD | tar -x -C $SP/archtest
cp platform/orchestration/actions/diagnose_assemble_bundle_action.go \
   platform/orchestration/actions/diagnose_assemble_bodycap_test.go \
   $SP/archtest/platform/orchestration/actions/
cd $SP/archtest && go test ./platform/orchestration/actions/ -count=1 && go build ./cmd/agent-chassis/
```

⚠ **`go build ./...` fails in the archive for reasons that are not yours**:
`docs/.../traffic_probe/deploy_setup/working_dir` holds two conflicting `package`
declarations, and `cmd/reasoningset/main.go:504` has three declared-and-not-used at
HEAD (`b82b3d8b4`). Build `./platform/... ./internal/... ./pkg/...` and
`./cmd/agent-chassis/` to get a signal about your own change.
⚠ `go vet ./platform/orchestration/actions/` reports
`load_component_library_actions.go:207: unreachable code` — pre-existing at HEAD.
Check `git diff HEAD --name-only -- <file>` before assuming a vet finding is yours.

## Council

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  <submission.json>
```

Find the run by PAYLOAD, not by the printed id, and budget ~30 minutes:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report' ORDER BY created_at;
```

# RUNBOOK — bugfix_427_event_render

## Checking boxingonline's evidence_base history (the fact-count correction)

```sql
SELECT sp.source, sp.created_by, sp.created_at, sp.is_current,
       jsonb_array_length(coalesce(sp.data->'facts','[]'::jsonb)) AS n_facts
FROM site_specs sp JOIN sites s ON s.id = sp.site_id
WHERE s.domain='boxingonline.com' AND sp.aspect='evidence_base'
ORDER BY sp.created_at;
```
Shows the superseded (2 facts) and current (1 fact) rows and when each was
written — the `is_current=false` row plus its `created_by` is how you find WHO
superseded it and get a lead on why (in this case, `site_delivery_and_editor`
acting on the owner's privacy ruling, bugs_open/420).

## Checking a needs_diagnosis run's actual verdict (not its `status`)

`status='complete'` is the item's LIFECYCLE, not its verdict — a run that
exhausted its iteration cap without confirming anything still shows
`complete`. The verdict is in `result`:

```sql
SELECT result->>'response' FROM site_work_items
WHERE summary LIKE '<enough of the summary to be unique>'
ORDER BY created_at DESC LIMIT 1;
```
Pipe to a file and Read it rather than letting a large JSON blob hit the
terminal — this one truncated to a 2KB preview inline and needed the full
`result->>'response'` text to see the per-iteration evidence trail.

## Checking who's live and what they're working on

```
ListAgents                         # every peer session + subagent, by name, busy/idle/waiting
python3 scripts/who-owns.py <N>    # which workstream directory owns bug/thread N, by commit recency
git log --oneline --since="1 hour ago"   # what actually landed recently, fleet-wide
git status --short <dir-or-file>   # what's dirty RIGHT NOW in a specific area — check before touching it
```
Combination that actually caught the collision in this bug: `ListAgents`
showed three sessions started in the last hour with names matching this bug's
territory; grepping docs for a column name (`entity_ids`) surfaced a brand-new
workstream directory nobody had told me about; `git status --short` on the
Go package showed the actual uncommitted files.

## Verifying a resolver's dependency declaration is correct

The lockstep tests do this automatically — they DRIVE every registered
`query.*` handler against a recording sqlmock and check the SQL it actually
issues against what `sourceDependencies` claims:

```
go test ./platform/orchestration/actions/queryresolve/... -run 'TestSourceDependenciesMatchTheResolvers|TestEveryRegisteredBaseDeclaresItsDependencies' -v
```
A new resolver needs, in `page_image_sources_test.go`'s `dependencyNeedles`
map, a SQL fragment that appears in its query and in NO other resolver's
query — verify with a plain grep across the package before trusting it:
```
grep -rn "site_specs\|evidence_base" platform/orchestration/actions/queryresolve/*.go | grep -v _test.go
```

## Mutation-testing a guard (the pattern used three times in this fix)

1. `grep -n "<the guarded line>" <file>` to get the exact line number.
2. `sed -i '<N>s#.*#<broken version, keeping any vars referenced elsewhere still referenced>#' <file>` — e.g. `if false { ... }` breaks a build if a variable used in the real condition becomes unused; `if <cond> && false { ... }` keeps it referenced.
3. Run the specific test; it must FAIL.
4. Restore the exact original line with the same `sed` pattern in reverse; run `go build ./...` and the test again to confirm both are clean.

## Submitting a platform-code change to council review

```
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>   # free admission check first
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>            # real submission, prints SUBMISSION_CORR
```
Gotcha hit this session: `git commit` refuses a `Council-Submitted:`/
`Council-Reviewed:` trailer that isn't a real UUID (or an 8+ char hex prefix
of one) — a placeholder like "pending" is rejected outright by the commit-msg
hook. Submit FIRST, get the real correlation, then commit.

Also hit: the pre-commit pattern-check flags un-gofmt'd files as advisory —
`gofmt -l <files>` to check, `gofmt -w <files>` to fix, before committing.

## Reading a submission's verdict later

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report' ORDER BY created_at;
```
Correlations from this bug's two submissions:
- `d0442d50-e383-477f-9ed8-19eaaeea3d93` — composeWriterBlock event-token fix.
- `08f56b7e-61e4-42d1-a3b6-13d700dd833c` — query.upcoming_events resolver + producer hook.

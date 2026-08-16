# RUNBOOK — bugfix 280

## Build / test

```bash
go build ./platform/orchestration/actions/discovery_checks/... && echo OK
go test ./platform/orchestration/actions/discovery_checks/... -run 'TestDecisionGuards' -v
go test ./platform/orchestration/actions/discovery_checks/...    # whole package suite, should stay green
go build ./...                                                    # whole platform, confirm nothing else broke
```

## Schema check (done once, before writing the SQL)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "\d site_components"
```
Confirms `UNIQUE (site_id, slot_name)` — so a scalar `SELECT sc.rendered_html
... WHERE sc.site_id = $1 AND sc.slot_name = 'header'` is correct (at most
one row) and does not need `string_agg`.

## Ownership / concurrency check (before starting, and again before commit)

```bash
python3 scripts/who-owns.py 280
```
Then, because who-owns is commit-history-only and blind to a live session
mid-fix (see memory `who-owns-is-blind-to-uncommitted-sessions`), grep recent
transcripts for the FIX-SITE symbol (not the bug number, not the mechanism
name — those saturate on incidental mentions):

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
CUT=$(date -u -d '45 minutes ago' +%Y-%m-%dT%H:%M:%SZ)
for f in $(find . -maxdepth 1 -name '*.jsonl' -newermt "$CUT"); do
  n=$(grep -c "storedPageAssemblySQL" "$f" 2>/dev/null)
  [ "$n" != "0" ] && echo "$f: $n"
done
```
`storedPageAssemblySQL` is the load-bearing symbol — the actual fix site.
`check_decision_guards.go` / `check_decision_guards` alone saturates on
package-wide registry listings (confirmed 2026-08-16: 4 sessions hit those,
zero hit the SQL const).

## Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
Submission JSON needs `rationale` + `plan` (≤8 edits) + `grounded_in` quotes.
Save the printed `SUBMISSION_CORR`. Poll:
```sql
SELECT current_step, status FROM orchestration_states WHERE
 collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

## Verify at the artefact, once an image ships (not this session's call — see PLAN)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <fix-commit-sha> <the stamp>
```
If the startup line has scrolled (agent-chassis is busy), fall back to the
binary probe in LANDMINES.md — never `strings`, always run a known-present
AND known-absent sha as controls in the same breath.

## Fleet-level behavioural check, once live

All 5 `doc_notes` rows with `categories ? 'decision-record'` are currently
non-chrome-scoped (280's own census), so there is no existing guard whose
verdict should visibly flip. The fix is confirmed live by the artefact check
above and by the unit tests, not by a change in fleet behaviour — record that
explicitly rather than waiting for a signal that isn't coming.

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
- `ff91e666-608d-4b26-9c41-d97d23a21437` — event-list component (migration 712). REVISE as of
  2026-09-03 (prior_art_librarian, HIGH, gating). Not yet resubmitted — see the 2026-09-03
  NOTES/HANDOFF for the answer to fold in (component_swap, not the full-rebuild path).

## Reading a full council report body (not just the decision)

`diagnosis_artifacts.body` is ONE JSON VALUE, not text — `psql -x -t` on a bare
`SELECT body` wraps it and can read as truncated/two-line. Pipe to a file and Read it:
```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='<CORR>' AND kind='council_report' ORDER BY created_at;" -x -t > /tmp/report.txt
```
Then `Read /tmp/report.txt` — every reviewer's `verdict`/`objections`/`notes` is in there,
not just the top-level `decision`.

## Dispatching a single action directly (bypassing the work-item queue)

`build-dispatch-loop` is a stateless, cron-burst poller — a hand-inserted work item can sit
for hours (LANDMINES/RUNNING_NOTES). To prove or ship something NOW, build the kafka
envelope by hand, fetching the target agent's OWN live workflow from `agent_definitions` so
it carries every guard that agent's normal callers get (do NOT hand-write a workflow).
`scripts/fire-section-edit.sh` is the proven, committed template (content_edit only); this
session's variants (component_swap, and a page-rerender/section_data_resolved dispatcher)
are scratch, not committed, but the pattern is:

```bash
WF=$(mktemp)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT default_config->'workflow' FROM agent_definitions
   WHERE type='<agent-type>' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
# build the envelope: headers (correlation/orchestration/request/message ids, action=process,
# client_id=demo_client — NOT hyphenated, spawn_actions.go interpolates it unquoted into a
# schema name), config.workflow=<contents of $WF>, input_data=<whatever that workflow's first
# step reads>. See fire-section-edit.sh for the exact python3 json.dumps shape.
. scripts/kafka-publish-lib.sh
kafka_publish_checked --topic system.agent.generic.requests --correlation "$CORR" \
  --payload "$(cat "$MSG_F")" --header "orchestration_id=$ORCH" ... # full header list in fire-section-edit.sh
```
Watch it:
```sql
SELECT orchestration_id, status, current_step, collected_data->>'__step_error'
FROM orchestration_states WHERE correlation_id='<CORR>' ORDER BY created_at;
```
**Delivery latency is NOT predictable** — one dispatch this session sat ~6 minutes before its
`orchestration_states` row even appeared; another completed in under 10 seconds. Poll with an
until-loop, don't sleep-and-check-once.

## Catching a fast-completing dispatch's OWN business-logic log lines

`kubectl logs -l app=agent-chassis --since=Nm` after the fact frequently shows NOTHING for a
run that completed in under 10 seconds — the ring buffer / your own multi-minute detour
between dispatch and check rotates it out, and infra-level logs (`coordinator.go`,
`processor.go`) still show while the ACTION's own `logger.Info`/`logger.Warn` calls do not
(seen three times this session, never resolved — may be sampling, may be something else).
The closest to a reliable capture: start `kubectl logs -f` FIRST, in the background, THEN
dispatch:
```bash
(timeout 25 kubectl -n ai-persona-system logs -l app=agent-chassis -f --since=1s > /tmp/live.log 2>&1 &)
./your-dispatch-script.sh   # fires immediately after the follower attaches
until grep -q "$CORR" /tmp/live.log; do sleep 1; done; sleep 3
grep "$CORR" /tmp/live.log | python3 -c 'import json,sys
for l in sys.stdin:
    d=json.loads(l); print(d.get("ts"),"|",d.get("msg"),"| step=",d.get("step_name",""))'
```
Even this did not surface the business-logic lines for `query.upcoming_events` — see NOTES
2026-09-03 for the open finding this was chasing. Prefer checking the DB row's
`content_data`/`rendered_html` directly over trusting log presence/absence either way.

## Checking a page's ACTUAL served origin, not its customer-facing domain

A site pre-handover (`sites.handed_over_at IS NULL`) is not DNS-live at its own domain —
`getent hosts <domain>` returning nothing is normal, not a sandbox restriction. The real
served copy is at `sites.publish_project` (a `.ugg2.com`-style preview subdomain):
```sql
SELECT domain, publish_target, publish_project, handed_over_at FROM sites WHERE domain='<domain>';
```
```bash
curl -s -m 15 "https://<publish_project>/<pages.url>" -o /tmp/page.html   # NOT https://<domain>/...
```

## Tracing a deploy past the DB row, to the actual GitHub Actions run

`pages.deployed_at`/`build_status='deployed'` is not proof the served bytes changed
(LANDMINES has several entries on this). `gh` was authenticated in this session
(`gh auth status`) — use it to find and read the REAL deploy job:
```bash
gh run list --repo gqls/sites --limit 100 | grep "<your commit message text>"
gh run view <run-id> --repo gqls/sites --log | grep -E "upload |delete .*old version"
```
The "Sync to B2" step's own `upload <path>` / `delete <path> (old version)` lines are the
actual evidence; a green "Job deploy completed with result: Succeeded" one-liner from
`kubectl logs -l app=github-actions-runner` is not — that log stream interleaves EVERY
site's deploy jobs fleet-wide with no per-job body, across (in this case) three separate
runner pods (`app=github-actions-runner` ×2, `app=github-actions-runner-vmsites` ×1), so a
`kubectl logs -l app=...` read is systematically the "one pod of N" trap.

## Applying ONE migration by hand without sweeping the pending directory

```bash
cat docs/agent_docs/sql_for_agents/<N>_name.sql | \
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 --single-transaction
# then record it (the file's own BEGIN/COMMIT already applied it — this just updates the ledger):
MIGRATIONS_DIR=/home/ant/projects/agentchassis/docs/agent_docs/sql_for_agents \
  ./scripts/migration/run-migrations.sh --record-only <N>_name.sql --note "applied by hand <date>; <what you verified>"
```
`--single-transaction` plus the file's own explicit `BEGIN`/`COMMIT` prints two harmless
`WARNING: there is already/no transaction in progress` lines — not errors, ignore them.
Never `--apply` on a directory with other sessions' pending files in it (this tree had at
least one concurrently, evidenced by interleaved GH Actions runs from an unrelated site).

## Building and verifying a frontend image actually contains the expected change

`make build-dashboard IMAGE_TAG=<tag>` builds from the WORKING TREE (frontends are exempt
from the git-archive-from-HEAD rule) via Docker — no local node/npm needed. "All layers
CACHED" in the build output does NOT mean the change is missing (a clean working tree
reusing prior layers is expected) — but verify inside the image before pushing, not after:
```bash
docker run --rm --entrypoint sh docker.io/aqls/admin-dashboard:<tag> -c \
  "grep -c '<a string unique to your change>' /usr/share/nginx/html/assets/*.js"
```
`docker push` is a production action and this session's own auto-mode classifier refused it
without confirmation — correct behaviour, not a bug; surface it to the user rather than
finding a workaround.

---

## Prove a struct field is never written (the `bugs_open/454` diagnosis, in one command)

The gotcha this exists for: Go compiles a struct field that is read and never assigned,
and hands you its zero value. `vet` and the linters pass it too. So "the value does not
arrive" has no toolchain signal at all — you have to ask for reads and writes separately.

```bash
grep -n '\.plan\b' platform/orchestration/actions/rerender_page_sections_action.go
# one hit -> that is a READ. Now look for the assignment; if there is no `x.plan =`
# anywhere in the package, the field is a permanent zero value.
```

Generalised, for any struct a refactor introduced to carry state across a new boundary:
for **each** field, one grep for `\.<field>\b` and one for `<field>\s*=`. A field with the
first and not the second is the defect.

## Build and test a change against committed HEAD when the shared tree does not compile

Routine on this tree, not an edge case: `go test ./...` reads the union of every session's
uncommitted work, so a neighbouring lane's half-finished file makes your own result
unreadable in both directions. `scripts/verify-head-builds.sh` extracts committed HEAD and
overlays only the files you name.

```bash
# BEFORE committing — your change, against HEAD, nobody else's WIP
scripts/verify-head-builds.sh \
  --with platform/orchestration/actions/rerender_page_sections_action.go \
  --with platform/orchestration/actions/rerender_page_sections_resolved_data_test.go \
  --test ./platform/orchestration/actions/...

# THE MUTATION PROOF: the test ALONE against unfixed HEAD must FAIL.
# This is the half that is easy to skip, and it is the half that proves the test can fail.
scripts/verify-head-builds.sh \
  --with platform/orchestration/actions/rerender_page_sections_resolved_data_test.go \
  --test ./platform/orchestration/actions/
```

⚠ Never hand-roll `git archive HEAD | tar` for this — that recipe is why the machine keeps
running out of space (CLAUDE.md, and `docs024_key_docs_latest/tmpfs_exhaustion/`).

⚠ Run it AFTER committing too. A pathspec commit takes the named file **from the working
tree**, so if another session had that same file dirty, their half rides along and only a
post-commit HEAD build will tell you. That happened on this lane's own fix commit
(`9831e9ab4` carried the `bugs_open/450` lane's `escalateRerenderToWriter` rework; HEAD
stopped compiling until they committed the closure as `587666be8`). Cheap warning sign
beforehand: `git status --porcelain <the file>` showing it already dirty when your own edit
is one line.

## Re-verify 427 at the artefact once a chassis carrying `9831e9ab4` rolls

```bash
# 1. what is the chassis actually running?
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 9831e9ab4 <the stamp>   # exit 0 = the fix is in that binary
```

The provenance line is a STARTUP line and scrolls out of reach on a busy service; an empty
result means "not in range", not "unstamped". Fall back to the binary probe, and always run
a known-absent sha as a control alongside the known-present one.

```bash
# 2. re-dispatch, then 3. read the ARTEFACT — never the job status
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT pc.content_data->'items' AS items, length(pc.rendered_html)
FROM page_components pc
WHERE pc.page_id='4b74ff1f-455a-4bb2-b81d-e1d0ec824f33' AND pc.slot_name='event-list';"
```

`items` non-empty and `rendered_html` no longer 1,813 bytes. Then curl the served page —
a DB row is not what a customer sees.

⚠ **Do not read a re-render's `escalated`/`rerendered` counts as evidence that data moved.**
That is precisely what 454 proves they cannot tell you: every count was healthy for a
fortnight while nothing was delivered. Note also that `escalateRerenderToWriter` gained a
second refusal disposition (`skipped_tool_pending_page` alongside `skipped_owned_page`,
commit `587666be8`, `bugs_open/450`), so any count of suppressed escalations that reads only
the older value will undercount from the next roll onwards.


---

## Dispatching a page-rerender directly (the working script, 2026-09-03)

The lane's own dispatcher, proven against `d0252fd4d`. Modelled on
`scripts/fire-section-edit.sh` and subject to the same two traps (`client_id` becomes a schema
name, so it must be `demo_client`; the workflow is pulled LIVE from `agent_definitions`, never
hand-written). `page-rerender`'s `rerender_sections` step reads `input_data.spec.reason`,
`input_data.spec.page_name` and `input_data.site_id`; `render_page` reads `page_id`, `site_id`,
`domain` — so the envelope needs all of them.

Scratch copy kept at `<scratchpad>/fire-page-rerender.sh`; the shape is:

```bash
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
# input_data: {site_id, domain, page_id, page_name,
#              spec:{reason:"section_data_resolved", page_name, page_id}}
. scripts/kafka-publish-lib.sh
kafka_publish_checked --topic system.agent.generic.requests --correlation "$CORR" ...
```

**Read the OUTPUT, never the counts.** `rerendered:N carried:0 escalated:false` is exactly what
a run produced for the fortnight `bugs_open/454` was live and delivering nothing. What
discriminates is the per-section metadata:

```sql
SELECT jsonb_pretty(jsonb_build_object(
  'slot', m->>'component_name',
  'content_keys', (SELECT jsonb_agg(k ORDER BY k) FROM jsonb_object_keys(m->'content_data') k),
  'n_items', jsonb_array_length(COALESCE(m->'content_data'->'items','[]'::jsonb)),
  'html_len', length(m->>'rendered_html')))
FROM orchestration_states os, jsonb_array_elements(os.collected_data->'rerender_sections'->'sections_metadata') m
WHERE os.correlation_id='<CORR>';
```

⚠ **Capture a CONTROL before dispatching** — `content_data` keys, item count, `length()` and
`md5()` of `rendered_html` — or "it looks populated" has nothing to be measured against.

## Asking which commit the chassis is running (the trap that fires first)

```sql
-- WRONG: a bare newest-first read. On 2026-09-03 this returned six rows for a SPAWNED
-- agent-image-build-handler pod still on the previous commit, minutes after the roll.
SELECT pod_name, git_commit, last_seen_at FROM service_binary_capabilities
WHERE service='agent-chassis' ORDER BY last_seen_at DESC LIMIT 6;
```

Filter to the STANDING pods and require them to agree:

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,START:.status.startTime,IMAGE:.spec.containers[0].image'
# then, with the replicaset hash from those names:
#   WHERE service='agent-chassis' AND pod_name LIKE 'agent-chassis-<rs>-%'
#   GROUP BY pod_name, git_commit
git merge-base --is-ancestor <your-commit> <the commit they agree on>   # exit 0 = shipped
```

Two standing pods reporting the same commit is the signal; one row from an unfiltered query is
not. `kubectl logs … | grep 'build provenance'` is the other route, but on a busy pod the line
has already scrolled and an empty result means "not in range", not "unstamped".

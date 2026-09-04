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

---

## Recovering from HEAD holding zero copies of a file (the 454-close race)

If `git ls-tree -r --name-only HEAD -- <dir1>/ <dir2>/ | grep <file>` returns **nothing** for a
file you know exists (you were just editing it, or a peer said they were):

1. **Do not re-create the content from memory or a stale local copy.** Find the commit that
   deleted it: `git log --diff-filter=D --oneline -- '<old-path>'`. The content is in its
   **parent**: `git show <that-commit>^:'<old-path>' > /tmp/recovered.md`.
2. **Check whether your OWN index already holds the current version staged** —
   `git status --porcelain -- '<new-path>'`. If it shows `new file:` or `AM`, your working tree
   is more current than the deleted commit's parent; prefer it.
3. **Commit the restoration naming only the path that still matches something git knows.** The
   deleted path errors (`did not match any file(s) known to git`) — that is expected, not a
   retry signal; drop it from the pathspec.
4. **Verify at HEAD, not the tree, before telling anyone it is fixed**:
   `git ls-tree -r --name-only HEAD -- <dir>/ | grep <file>` should return exactly one row.

Full mechanism and the general trap: LANDMINES, "a correct pathspec commit made WRONG by
someone else's concurrent `git mv`".

## Verifying a fix at the SERVED page, not just the DB row

`curl` to a live customer domain may hang or `ETIMEOUT` from this environment even when the
domain is genuinely resolvable elsewhere — not necessarily a sandbox restriction (a
pre-handover site is legitimately not DNS-live at all, see below). `WebFetch` succeeded where
`curl` failed on the same URL this session:

```
WebFetch(url="https://<domain>/<path>", prompt="does section X show <specific content>?")
```

Ask a **specific, falsifiable question** ("does the image src attribute have a real path or is
it empty?"), not "does this page look right?" — a vague prompt gets a vague, unfalsifiable
answer back from the summarising model.

## Checking a site's real served origin before concluding a fix "isn't showing"

```sql
SELECT domain, publish_target, publish_project, handed_over_at FROM sites WHERE domain='<domain>';
```

`handed_over_at IS NULL` means the site is **pre-handover and not DNS-live at its own domain at
all** — `getaddrinfo`/`WebFetch` failing on `https://<domain>/...` is the EXPECTED result, not
evidence the fix failed. The real deploy target is `portfolio-sites/<domain>` in the `sites`
repo (verify via the GitHub Actions "Sync to B2" log, see below), which is what a handover would
eventually point real DNS at. The `.ugg2.com`-style preview subdomain is a THIRD, separate
target with its own reconciliation pipeline (`site-publisher`) and its own tick — checking it and
finding the old content is not evidence either, unless you have separately triggered that
pipeline.

## Confirming a chassis roll actually shipped the fix you need (do this before re-testing anything)

```bash
RS=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}' | sed 's/-[^-]*$//')
kubectl -n ai-persona-system get pods -l app=agent-chassis -o custom-columns='NAME:.metadata.name,START:.status.startTime,IMAGE:.spec.containers[0].image'
```
```sql
SELECT pod_name, git_commit, max(last_seen_at)
FROM service_binary_capabilities
WHERE service='agent-chassis' AND pod_name LIKE '<RS>-%'
GROUP BY pod_name, git_commit ORDER BY 3 DESC;
```
Then `git merge-base --is-ancestor <your fix's commit> <the commit they agree on>` — exit 0 is
the only thing that counts. **A dozen freshly-started pods on the SAME image is not a roll** —
check for spawned agents first (`kubectl get pods --sort-by=.status.startTime` and look at
whether the NAME matches the standing chassis replicaset or a spawned `agent-<type>-<uuid>`
pattern); this session was fooled by exactly that shape once today before catching it. A "fresh
build reported" claim that turns out to be the SAME standing pods, SAME commit, started BEFORE
your fix's own commit timestamp is a real, common failure mode — record it as a dated negative
(with the four checks and their readings) rather than silently re-checking later and hoping.

---

## Correcting a page's composition in the AUTHORITY (added 2026-09-03, migration 750)

**When you need this:** you have changed what sections a page has by editing
`pages.sections`, and you want it to survive. It will not, on its own — the next page
**build** reads `site_plan_sections` (tier 1) and syncs it down over the cache.

### 1. Establish which tier actually serves the page

Do this first; the answer changes what you must write, and a tier-2-served page looks
authority-less if you only query tier 1.

```sql
-- tier 1
SELECT sps.id, sps.ordering, sps.component_name, sps.assigned_fact_ids, sps.subject,
       sps.component_version_id
  FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id
 WHERE sp.site_id = '<site>' AND sp.is_current AND sps.page_name = '<page>'
 ORDER BY sps.ordering;
-- tier 2 — zero rows for most sites, but NOT all; check, do not assume
SELECT count(*) FROM site_specs WHERE site_id = '<site>' AND aspect = 'site_plan';
```

### 2. The four things you must not disturb

`ordering` is a positional join key. Before writing anything, list what is keyed to it:

```sql
-- section imagery binds to the ORDINAL, never the component name
SELECT scope_ref, key, kind FROM site_plan_imagery
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id='<site>' AND is_current)
   AND scope = 'section' AND scope_ref LIKE '<page>:%';
```
`assigned_fact_ids` (`'[]'` ≠ `NULL`), `subject`, `page_components.position`, and
`site_plan_imagery.scope_ref`. **So rename in place at the same `ordering`; never
delete-and-reinsert.** Migration `154` is the delete-renumber-insert shape and is the one
you will find first — it predates three of those four consumers.

### 3. Safety preconditions to assert (not to assume)

```sql
SELECT count(*) FROM site_plans WHERE site_id = '<site>';          -- superseded plans?
SELECT built_from_plan_version FROM pages WHERE site_id='<site>' AND name='<page>';
-- locked rows: if non-zero, a raw list comparison is NOT the drift check's comparison
SELECT count(*) FROM page_components pc JOIN pages p ON p.id = pc.page_id
 WHERE p.site_id='<site>' AND p.name='<page>' AND COALESCE(pc.build_status,'') <> 'removed'
   AND NOT (pc.locked_at IS NULL
            OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()));
```
Only the `is_current` plan may be written. Mutating a **superseded** plan's rows falsifies
build history for every page stamped to it.

### 4. Write the verify block as `DO`/`RAISE`, and induce a failure before shipping

A verify block of bare `SELECT`s **cannot abort the transaction** — `ON_ERROR_STOP` ignores
a non-empty result. And never inspect a variable the block itself set; re-`SELECT` the live
row.

```bash
# 1. copy the migration, point ONE needle at a value no write produces
sed 's/<expected>/<impossible>/' <mig>.sql > "$SCRATCH/induced.sql"
# 2. apply it — it must ABORT and roll back, leaving the pre-state intact
cat "$SCRATCH/induced.sql" | kubectl -n ai-persona-system exec -i postgres-clients-0 \
  -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
# 3. re-query the pre-state to prove nothing moved, THEN apply the real file
```
This proves two things at once: the guard fires, and the transaction is atomic.

### 5. Apply, then record — recording is part of applying

```bash
cat <mig>.sql | kubectl -n ai-persona-system exec -i postgres-clients-0 \
  -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
MIGRATIONS_DIR=docs/agent_docs/sql_for_agents ./scripts/migration/run-migrations.sh \
  --record-only <mig>.sql --note "<what you actually verified>"
```
Never `--apply` the directory: it takes every other session's pending files.

### 6. Prove it, and prove it did NOT change the artefact

An authority correction on an already-correct page must alter **nothing** downstream.

```sql
-- the loader's OWN query (load_page_sections_from_spec_action.go:142-148)
SELECT sps.component_name FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id
 WHERE sp.site_id='<site>' AND sp.is_current AND sps.page_name='<page>' ORDER BY sps.ordering;
-- its sync-down guard (:562) must be a no-op
SELECT sections FROM pages WHERE site_id='<site>' AND name='<page>';
-- and the artefact must be untouched — capture BEFORE, compare AFTER
SELECT pc.position, pc.slot_name, length(pc.rendered_html) FROM page_components pc
  JOIN pages p ON p.id=pc.page_id WHERE p.site_id='<site>' AND p.name='<page>' ORDER BY pc.position;
SELECT updated_at, deployed_at FROM pages WHERE site_id='<site>' AND name='<page>';
```
**`pages.updated_at` moving is a finding, not a success.**

### 7. ⚠ Do NOT dispatch a rebuild to "pick up" the fix on a tool page

`page-build-handler`'s `load_page_record` carries `refuse_owned_page: true` and refuses any
`page_type='tool'` page with zero `component_level='tool'` components. You will get an
`OWNED_PAGE_GUARD` failure and a fresh `owned_page_review` item. Forcing past it produces the
TP-004 clobber — a generic prose page where the tool belongs. There is nothing to pick up
anyway: the artefact is already correct; the migration only removed a latent revert.

**And note the ordering that follows:** that refusal is *derived, not stored*. It
self-clears the moment a real tool component lands. So on a tool page, **correct the
authority BEFORE building the tool**, or the build arms the very revert you removed.

## Triaging the `section_source_drift` backlog (added 2026-09-03, migration 753)

Never read the item's `spec` — it is frozen at filing time and reads as current. Re-derive
both sides live, mirroring the check's precedence exactly:

```sql
WITH tier1 AS (
  SELECT sp.site_id, sps.page_name, jsonb_agg(sps.component_name ORDER BY sps.ordering) AS auth
    FROM site_plans sp JOIN site_plan_sections sps ON sps.plan_id = sp.id
   WHERE sp.is_current GROUP BY sp.site_id, sps.page_name
), tier2 AS (
  SELECT ss.site_id, pg->>'name' AS page_name, pg->'sections' AS auth
    FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'pages') pg
   WHERE ss.aspect='site_plan' AND ss.is_current AND jsonb_typeof(ss.data->'pages')='array'
)
SELECT s.domain, wi.spec->>'page_name',
       COALESCE(t1.auth, t2.auth) AS live_authority, p.sections AS live_cache,
       CASE WHEN COALESCE(t1.auth,t2.auth) IS NOT DISTINCT FROM p.sections THEN
              CASE WHEN p.sections = wi.spec->'pages_sections' THEN 'cache_held'
                   WHEN p.sections = wi.spec->'authoritative'  THEN 'AUTHORITY WON'
                   ELSE 'third_list' END
            ELSE 'LIVE DIVERGENCE' END AS verdict
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
  LEFT JOIN tier1 t1 ON t1.site_id=wi.site_id AND t1.page_name=wi.spec->>'page_name'
  LEFT JOIN tier2 t2 ON t2.site_id=wi.site_id AND t2.page_name=wi.spec->>'page_name'
  LEFT JOIN pages p ON p.site_id=wi.site_id AND p.name=wi.spec->>'page_name'
 WHERE wi.item_type='section_source_drift' AND wi.status NOT IN ('complete','cancelled','rejected');
```

⚠ **`COALESCE(tier1, tier2)`, not `tier1`.** Joining only tier 1 reports a tier-2-served
page as divergent, because its tier-1 authority is NULL. That mistake was made and caught
here on `leopardessconsulting.co.uk/index`.

**`AUTHORITY WON` means a human's edit was destroyed.** Close it so it stops blocking the
dedup key, but record the direction in `result` — closing it as a plain success ratifies
the loss. See `bugs_open/469`.

---

## 2026-09-04 — provenance census recipes

### Is a lane actually dormant? Ask, don't only measure

`scripts/who-owns.py 427` reads COMMITS, so it cannot see a session mid-work, and a
"session has ENDED" claim decays by REVERSAL (`54df41b22`) so re-measuring never catches it.
The check that works is `git log --since` on the **bug file path** to separate *contributions
from other lanes* from *the lane resuming* — today's two commits to `bugs_open/427` were the
`boxingonline.com` and `calendar` lanes, not this one — and then `ListAgents` + a direct
message. All three peers answered within minutes.

### The provenance measurement, both ends

```sql
-- which components were rendered FROM facts, and which carry it to markup
SELECT cc.function,
       count(*)                                                            AS placements,
       count(*) FILTER (WHERE pc.rendered_html LIKE '%data-fact-id%')      AS emits_data_fact_id,
       count(*) FILTER (WHERE pc.rendered_html LIKE '%data-series=%')      AS emits_data_series
FROM page_components pc LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.content_data::text LIKE '%fact_id%'
GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **Run BOTH attribute columns, not just `data-fact-id`.** Querying only the one you expect
returns 0 and reads as "nothing carries provenance". Three placements do — under
`data-series`, a spelling with **one emitter and zero readers** in `platform/`, `internal/`
or `pkg/`. Checking one vocabulary is how a shared-vocabulary defect reads as a missing
feature.

### The tool dataset census — and why NOT to key it on dates

```
kubectl … psql -t -A -c "SELECT jsonb_build_object('name',name,'schema',COALESCE(input_schema,'{}'),'tmpl',html_template)::text
  FROM content_components WHERE component_level='tool' AND is_active;" > tools.jsonl
```
then extract `<script>` bodies, take innermost object literals (`\{[^{}]{10,800}\}`), and score
each on **two independent predicates** — they are NOT nested, which is a correction this
lane had to make after the `482` lane found `ds=18, ea=20` on one row:

- `ds` — has a human-readable string value (>=8 chars containing a space);
- `ea` — has an **identity** key (`name`/`title`/`brand`/`venue`/…) AND an **attribute** key
  (`postcode`/`website`/`price`/`rate`/`year`/…), regardless of string content.

**Use `ea` for a fabrication gate.** `ds` drops single-word entity names ("Secateurs"), and
invented product and practice names are frequently one word. `[MEASURED 2026-09-04]` 335
active tools → 134 with a dataset, **5** with `ea >= 2`.

⚠ **Do NOT key it on date shapes.** The `482` lane's date-shaped census returned **1 of 335**,
which reads as a first occurrence; the entity+attribute framing found five candidates and at
least three real ones. And record count is a bad axis outright: their simulation over this
census gives **89% false positives** at `>=15` records, with the largest legitimate dataset at
**73** and the motivating fabrication at **6**.

### The number that reframes any tool-provenance claim

```sql
SELECT count(*) AS active_tools,
       count(*) FILTER (WHERE input_schema IS NULL)      AS schema_null,
       count(*) FILTER (WHERE input_schema::text = '{}') AS schema_empty
FROM content_components WHERE component_level='tool' AND is_active;
--  335 | 287 | 0
```

**86% of tools declare nothing.** So "declares no fact-bearing field" is the DEFAULT state,
not a signal: sound as an *exculpatory* test, near-inert as an *inculpatory* one. Any claim
about a schema-keyed mechanism's reach must be dated and stated as a fraction of 335.

### Council submission shape — `plan` is an OBJECT, not a list

`097_TRIGGER` refuses with `ERROR: .plan missing` if `plan` is a list. The schema is
`{"rationale":…, "submitter":…, "plan": {"summary":…, "edits":[…<=8], "grounded_in":[…], "risks":…}}`.
`DRY_RUN=1` validates admission for free — use it every time; it also prints scope warnings
(today: unclassified `_ISLAND`/`_RELOCK` migration suffixes, treated as IN scope by default).

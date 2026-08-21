# RUNBOOK — bug 315, publish-seam evidence

Every command here had a gotcha. The gotcha is attached to the command, not kept separately.

## Ordering claims about a workflow — join on `next_step`, never read the key order

`jsonb_each` returns steps in arbitrary order, and reading them by eye gets the sequence WRONG
(`page-build-handler` prints `deploy_page` above `update_status` and actually runs it after).

```sql
WITH steps AS (
  SELECT ad.type AS agent, e.key AS step, e.value->>'action' AS action,
         COALESCE(e.value->>'next_step', e.value->'on_success'->>'next_step') AS next,
         e.value->'config'->>'status' AS status
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') e
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
)
SELECT u.agent, u.step, u.status,
       (SELECT string_agg(p.step||'('||p.action||')', ', ')
          FROM steps p WHERE p.agent=u.agent AND p.next=u.step) AS preceded_by
FROM steps u WHERE u.action='update_page_status' ORDER BY u.agent;
```

⚠ `default_config->'workflow'->'steps'` is an **object keyed by step name**, not an array.
`jsonb_path_query_array(default_config,'$.workflow.steps[*].name')` returns `[]` and looks like
"this agent has no steps" — it does not; you asked an array question of an object.

⚠ Always carry all three of `is_active` / `is_snapshot=false` / `deleted_at IS NULL`. Snapshots are
real rows and will double your counts.

## Is a schema column actually written? (the check that turned this bug around)

Two separate questions; ask both, because either alone misleads.

```sql
SELECT count(*) AS total, count(content_hash) AS populated FROM pages;          -- 786 / 0
SELECT count(*) AS total, count(deploy_commit) AS populated FROM page_components; -- 1775 / 0
```
```bash
# and the writer side — note --include and the ABSOLUTE path (see the pwd trap below)
grep -rn "deploy_commit" --include=*.go /home/ant/projects/agentchassis | grep -v _test
```

⚠ **Run the grep INCLUDING tests first.** "No non-test writer" and "no code at all mentions it" are
different findings; the second is much stronger and is what was true here.

## ⚠ The Bash working directory persists between calls

A `cd` inside one compound command changes the directory for every later call. Three greps returned
empty and were read as absences; they were run from a docs subdirectory. **Use absolute paths, and
never `cd` in a compound command.** Cost: nearly missed `docs026_concept_register/register/
deployment-github.md`, the document that names the whole delivery mechanism.

## Grading at the artefact (the only layer that is evidence)

```bash
curl -sI "https://${domain}${url}?cb=$RANDOM$RANDOM" --max-time 20 | grep -iE '^(HTTP|last-modified|cf-cache-status)'
```

⚠ **Always cache-bust.** `cf-cache-status: DYNAMIC` is the confirmation you read the origin.
⚠ `last-modified` is the **per-object** write time — pages on different domains carry different
values, which is the control proving it is not one global checkout mtime.
⚠ **A batch of pages sharing a `last-modified` to the second is NORMAL**, not a coincidence: the
origin is rewritten per changed domain directory by one `b2 sync`.

## Sizing "deployed_at is stale against the origin" — and why the number lies

```bash
# pages.txt: domain|url|deployed_at  from a psql -At -F'|' query
while IFS='|' read -r domain url dep; do
  lm=$(curl -sI "https://${domain}${url}?cb=$RANDOM$RANDOM" --max-time 25 | grep -i '^last-modified' | sed 's/^[^:]*: //I' | tr -d '\r')
  echo -e "$domain\t$url\t$dep\t$lm\t$(( $(date -u -d "$lm" +%s) - $(date -u -d "$dep" +%s) ))"
done < pages.txt
```

⚠ **This returned 40 of 40 "stale" and that is NOT 40 defects.** The origin lags the commit by tens
of minutes in whole-domain batches, so at any instant most correctly-behaving pages are "stale"
against their own stamp. A 40/40 result is the shape that should send you to the raw values — there,
all 40 shared ONE three-second `last-modified` window, which is the batch, not 40 failures.
**This comparison cannot separate "not synced yet" from "will never sync"; only elapsed time can,
and the known bad case took six hours.**

## Did the commit actually happen, and did the runner deploy it?

```sql
SELECT updated_at,
       collected_data->'deploy_result'->'response'->'data'->>'success'   AS ok,
       collected_data->'deploy_result'->'response'->'data'->>'file_path' AS path,
       collected_data->'deploy_result'->'metadata'->>'status'            AS meta_status
FROM orchestration_states
WHERE collected_data ? 'deploy_result'
  AND collected_data->'deploy_result'->'response'->'data'->>'domain' = '<domain>'
  AND updated_at > NOW() - INTERVAL '70 minutes'
ORDER BY updated_at DESC;
```
```bash
kubectl -n ai-persona-system logs github-actions-runner-54fd5c8547-<pod> --tail=40 | grep -E 'Running job|completed with result'
```

⚠ **`deploy_result` has TWO SHAPES.** Inline deploys sit at `deploy_result.response.data.*`;
deploys done by a called sub-agent are nested one level deeper at
`deploy_result.response.deploy_result.response.data.*`. The query above sees only the first and
reports the other **7.7% of runs (57 of 744 over 7 days)** as having no verdict at all.
⚠ Deploy jobs arrive in **clusters 25–50 minutes apart**, so "no job for 36 minutes" is inside the
normal spacing and is not evidence of a stall.

## Council gate / diagnosis loop refusing with a usage limit

```sql
SELECT date_trunc('minute',occurred_at) m, agent_type, count(*)
FROM agent_error_log
WHERE occurred_at > NOW() - INTERVAL '30 minutes' AND error_message ILIKE '%usage limit%'
GROUP BY 1,2 ORDER BY 1;
```

⚠ The message says *"You will regain access on 2026-09-01"* and reads like a hard lockout. It is
not: the same error appears on five separate days over the past month, and on the day I hit it the
council gate was completing `complete_approved` / `complete_revise` runs **in the same minutes** as
other calls were being refused. **Check whether the fleet is still completing LLM work before
reporting an outage** — `orchestration_states WHERE orchestration_name ILIKE '%council%'` is the
cheap read.

---

# Part 2 — applying and shipping the fix (added 2026-08-19)

## Applying a migration WITHOUT sweeping other threads' pending files

At the time of writing the runner listed **129 pending** files belonging to a dozen lanes. `--apply`
takes every one of them, in order. Scope it with a scratch directory holding only your file:

```bash
SCOPE=<scratch>/mig491; mkdir -p "$SCOPE"
cp docs/agent_docs/sql_for_agents/491_*.sql "$SCOPE/"
MIGRATIONS_DIR="$SCOPE" ./scripts/migration/run-migrations.sh            # expect: Pending (1)
MIGRATIONS_DIR="$SCOPE" ./scripts/migration/run-migrations.sh --apply
```

⚠ **The assignment must be on the SAME LINE as the command.** On its own line it scopes nothing and
the run sweeps the fleet. The scratch dir is the mechanism; the env var alone is not.
⚠ The dry run **executes** each pending file inside a doomed transaction to probe it — brief row
locks, sequences may advance. `--no-probe` if that matters.

## Holding a migration that must NOT be applied yet

Rename it, do not add a banner — **the runner does not read comments.**

```bash
# NNN_name_HOLD.sql — SIDECAR_RE '_[A-Z][A-Z0-9_]*\.sql$' excludes it from --apply
./scripts/migration/run-migrations.sh --no-probe | sed -n '/^Sidecars/,/^$/p' | grep NNN   # must appear
./scripts/migration/run-migrations.sh --no-probe | sed -n '/^Pending/,/^$/p'  | grep NNN   # must NOT
```

Held visibly, not hidden. Put the by-hand apply commands in the file's own header — the person who
runs them will not be you.

⚠ **A `_HOLD` file is a sidecar, so the runner will never apply it even when you want it to.** Pipe
it to psql by hand:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/NNN_..._HOLD.sql
```

⚠ **And do NOT then reach for `--record-only`** — the runner refuses it: *"is an UPPERCASE-suffixed
sidecar … recording one is meaningless"*. Measured 2026-08-19, after this runbook told me to do
exactly that. It is harmless: a sidecar never appears in Pending, so the runner cannot double-apply
it, and the file's own already-applied guard catches a human re-run. **Record the apply in the lane's
NOTES instead** — that is the only place it will be found.

## Snapshotting an agent before a config change

```sql
SELECT snapshot_agent('<type>', '<file>: pre-update');   -- TWO args: writes agent_definitions_backup
```

⚠ **Do not ask whether a snapshot exists — ask whether it holds the PRE-change config.** A row
carrying the post-change value restores nothing:

```sql
SELECT type, snapshot_taken_at, snapshot_reason,
       (default_config->'workflow'->'steps' ? '<the step you removed>') AS holds_pre_change_step
FROM agent_definitions_backup WHERE snapshot_reason LIKE '<your file>%'
ORDER BY snapshot_taken_at DESC;
```

⚠ `ORDER BY created_at` returns an **arbitrary** snapshot — `agent_definitions_backup` copies the
SOURCE row's `id` AND `created_at`, so every backup row for one agent shares them. Order by
`snapshot_taken_at`, and find yours by the distinctive `snapshot_reason` you passed.

## Migration numbering on a shared tree

`ls docs/agent_docs/sql_for_agents/ | grep -oE '^[0-9]+' | sort -n | uniq | tail -4`

⚠ **Re-check at commit time, not at write time.** 492 and 493 were both claimed by other lanes while
this lane's 492 was being drafted; it shipped as 494. Duplicates already exist in the tree (488
twice), so a collision is untidy rather than fatal — but it makes `git log <number>` ambiguous.

## Two prerequisites before the held config can go in

```bash
# 1. Which image is actually running?
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.spec.containers[0].image}{"\n"}{end}'
# 2. Does that image carry the key? Ask the binary's own stamp, not git:
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 086f9b7b7 <the stamped sha> && echo CARRIES-THE-KEY
```

⚠ An **empty** provenance grep means "not in log range", NOT "unstamped" — it is a startup line and
it scrolls on a busy service.

## Proving a guard rather than trusting a green test

Both new guards were mutation-proved, and the mutation is the evidence — a passing test on
unmutated code proves only that the code compiles.

```bash
cp <file> /tmp/f.bak
# break exactly the thing the test claims to catch, e.g. skip the base64 decode
go test ./<pkg>/ -run <Test>      # MUST fail, and on the intended assertion
cp /tmp/f.bak <file>
go test ./<pkg>/ -run <Test>      # MUST pass again
```

⚠ Check *which* test failed. A mutation that fails several tests may have hit a guard in series; one
that fails none means the test was never testing that.

## Reading a council verdict

```sql
SELECT created_at, metadata->>'decision', body
FROM diagnosis_artifacts
WHERE correlation_id='<YOUR SUBMISSION_CORR>' AND kind='council_report'
ORDER BY created_at DESC;
```

⚠ **Never** the `doc_notes` query the trigger prints (`WHERE categories ? 'council-gate' ORDER BY
created_at DESC LIMIT 1`) — the table is fleet-shared and that returned **another lane's APPROVED
verdict** while mine was `revise`. Read by correlation, always.
⚠ On a resubmission the correlation is reused, so filter by `created_at` to get the new round.

---

## Part 3 — the divergence sweep (D5 / DGH-015), added 2026-08-21

### The one query D6 depends on: is every `deployed` stamper still ARMED?

**Re-run this before trusting a `page_content_divergence` finding, and before taking D6.** The check
assumes every path that stamps `deployed` also writes the fingerprint. An UNARMED stamper leaves a
stale hash and the check convicts a healthy page — permanently.

```sql
WITH steps AS (
  SELECT ad.type AS agent, st.key AS step_key,
         st.value->>'action' AS action,
         st.value->'config'->>'status' AS status_cfg,
         st.value->'config'->>'deploy_result_field' AS deploy_field
    FROM agent_definitions ad
    CROSS JOIN LATERAL jsonb_each(ad.default_config) wf(key,value)
    CROSS JOIN LATERAL jsonb_each(CASE WHEN jsonb_typeof(wf.value)='object' AND wf.value ? 'steps'
                                       THEN wf.value->'steps' ELSE '{}'::jsonb END) st(key,value)
   WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
)
SELECT agent, step_key, COALESCE(deploy_field,'*** UNARMED ***') AS deploy_result_field
  FROM steps WHERE action='update_page_status' AND status_cfg='deployed'
 ORDER BY (deploy_field IS NULL) DESC, agent;
```

Expected 2026-08-21: **three rows, all armed** (`page-rerender/update_status`,
`report-builder/update_status`, `section-editor/update_page_status`). **Any row reading UNARMED
invalidates the check's premise** until it is armed or D6 ships.

> ⚠ **THE CONFIG KEY IS `status`, NOT `build_status`.** Writing `build_status` in the predicate
> above returns **zero rows** — cleanly, with no error, looking exactly like "no agent stamps
> deployed". That happened while writing this, and a zero from the wrong column is indistinguishable
> from a zero that means something. If this query returns 0 rows, suspect the query before you
> conclude the fleet has no stampers.

### Measure divergence by hand, fleet-wide

The check's whole comparison, runnable from a terminal. This is what produced the 228-of-228 result.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -A -F'|' -t -c "
SELECT s.domain, p.url, p.content_hash
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE p.content_hash IS NOT NULL AND p.status='active' AND COALESCE(s.domain,'') <> ''
 ORDER BY s.domain, p.url;" > /tmp/hashed_pages.txt

while IFS='|' read -r domain url stored; do
  served=$(curl -s --max-time 20 "https://${domain}${url}?cb=$RANDOM$RANDOM$$" | sha256sum | cut -d' ' -f1)
  [ "$served" = "$stored" ] && echo "MATCH $domain$url" || echo "DIVERGED $domain$url $stored $served"
done < /tmp/hashed_pages.txt
```

> ⚠ **DO NOT PIPE THE psql CAPTURE THROUGH `tee … | head`.** `head` exits after N lines, `tee` takes
> SIGPIPE, and **the file it was writing is silently truncated**. That turned 228 rows into 21 here,
> and 21 was a plausible-looking number — there was no error and nothing to notice. Redirect to the
> file, then read the file.

> ⚠ **Always cache-bust, and never `HEAD`.** The body is the question. A cache-bust query is safe
> because `PageFilePathFromURL` refuses a stored url that already carries one, so it cannot collide
> with a real parameter.

### Measure the delivery lag (what the settle window is sized against)

Re-probe every recently-stamped page every 2 minutes and record age-at-probe against verdict. This
produced the 1,099-reading distribution: 3 DIVERGED at ages 1s/13s/14s, all converged by 140–156s,
0 of 995 readings at age ≥157s diverged.

```bash
# one pass; wrap in a loop with `sleep 120` for a distribution
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -A -F'|' -t -c "
SELECT s.domain, p.url, p.content_hash, round(extract(epoch from (now()-p.deployed_at)))
  FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.content_hash IS NOT NULL AND p.status='active' AND COALESCE(s.domain,'')<>''
   AND p.deployed_at > now() - interval '45 minutes'
 ORDER BY p.deployed_at DESC LIMIT 40;"
```

> ⚠ **`pkill -f lag_watcher.sh` KILLS ITS OWN SHELL.** The pattern matches the command line of the
> very `bash -c` running it, so the kill lands on the killer and the output stops mid-stream, looking
> like the watcher survived. Break the literal (`pkill -f "lag_"'watcher'`) or match on the script's
> unique User-Agent instead, and verify with a separate `pgrep`.

### After applying 526 (enabling the check)

**Read the DAMAGE query first — this is `bugs_open/336`'s lesson, which this lane learned by
breaking every page-publish in the estate for 33 minutes while verifying that its config was right.**
An unregistered check name fails the `run_checks` step for the WHOLE agent, taking `site_unreachable`
down with it — and that damage does not appear anywhere in this check's own output.

```sql
-- 1. WHAT DID I BREAK?
SELECT current_step, status, count(*) FROM orchestration_states
 WHERE agent_type='availability-discovery-agent' AND created_at > now() - interval '30 minutes'
 GROUP BY 1,2;
-- 2. only then: did it find anything? Expect ZERO on day one.
SELECT summary, spec->>'stored_hash', spec->>'served_hash'
  FROM site_work_items WHERE item_type='page_content_divergence' ORDER BY created_at DESC LIMIT 10;
```

A finding on day one is likelier to be a defect in the check than a divergence in the fleet: re-run
the `curl | sha256sum` comparison by hand before believing it. Rollback is
`526_enable_page_content_divergence_HOLD_ROLLBACK.sql`, which removes the name and asserts
`site_unreachable` survives.

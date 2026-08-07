# HANDOFF 2026-08-07 — continue here (`bugs_open/206`, directory-build-handler)

**Read `PLAN_2026-08-06_directory_build_handler.md` first** (design + why),
then `NOTES_directory_build_handler.md` (the trail, including one thing that
went sideways — a same-file passenger — and why it's harmless). This file is
the concrete next-steps checklist.

## State in one line

Code + migrations are written, unit-tested (against `git archive HEAD`), and
**committed** (`f750595dd`). Submitted to the council gate
(`SUBMISSION_CORR=5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`) but the verdict has
**not been read**. Nothing has been applied to the DB, no image has been
built, nothing is live. Re-verify every claim below against the live system
before acting — this tree moves fast (three separate rounds of unrelated
concurrent churn happened *during* this one work session).

## Step 0 — re-orient, don't trust this file's numbers blindly

```bash
git log --oneline -5   # confirm f750595dd is still there, see what's landed since
git status --short     # this tree is heavily concurrent; expect noise from other sessions
grep -n "^IMAGE_TAG" makefile   # re-check — it moved TWICE during the writing session (1259->1260->1261)
```

## Step 1 — read the council verdict

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='5b8e4cf7-31c3-4793-a550-d6b9be1f00e8' AND kind='council_report'
ORDER BY created_at;
```
If empty, it's still queued/running — the dispatch lane was busy with several
other orchestrations when this was submitted; this repo's own norm is budget
~30 minutes, not ~2, before treating "no row yet" as anything other than
latency.

- **APPROVED** → nothing to redo; when you commit the *next* piece of this
  lane's work (the re-triage in Step 5, or any further code), use
  `Council-Reviewed: 5b8e4cf7-31c3-4793-a550-d6b9be1f00e8` on that commit —
  the existing `f750595dd` commit doesn't need amending (forward-only; 098
  credits it automatically at report time via the `Council-Submitted`
  trailer already on it).
- **REVISE** → objections come back with read-only checks already answered.
  Read them, fix in a new commit (never amend), resubmit with
  `RESUBMIT_CORR=5b8e4cf7-31c3-4793-a550-d6b9be1f00e8` so the trail
  accumulates.
- **REJECTED** → a guardian veto with a named contained alternative. Read it
  (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
  created_at DESC LIMIT 1;`) before touching anything else in this lane —
  the code is already on the shared branch either way; a REJECTED verdict
  doesn't un-ship it, it tells you what to do differently next.

## Step 2 — apply the two migrations (only after reading the verdict)

**Do NOT run `--apply`** — it takes EVERY pending file in the directory, and
that directory carries dozens of other threads' half-finished, precondition-
gated migrations (measured mid-session: files requiring pod-greps, files that
error until another migration runs first, files explicitly `REFUSED` pending
a check). Apply just these two, in order, by hand:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/325_directory_listing_binds_to_business_directory_query.sql

kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/326_directory_build_handler_agent.sql
```
Each ends in its own `DO $$ ... RAISE EXCEPTION ...$$` guard — a clean run
with no error means it's genuinely applied. Then register them in the ledger
so a future dry-run doesn't think they're still pending:
```bash
./scripts/migration/run-migrations.sh --record-only \
  325_directory_listing_binds_to_business_directory_query.sql \
  --note "applied by hand 2026-08-07, verified via its own DO block"
./scripts/migration/run-migrations.sh --record-only \
  326_directory_build_handler_agent.sql --note "applied by hand 2026-08-07, verified via its own DO block"
```

**326's `image_tag` says `v1.0.1260`** — re-check the actual tag you build in
Step 3 and `UPDATE agent_definitions SET image_tag='<real tag>' WHERE
type='directory-build-handler';` if it differs (it will; the tag had already
moved to 1261 by the time this was written, before any build in this lane
happened at all).

## Step 3 — build, bump the tag, roll

```bash
git commit  # if step 1 required a fix — do that first, this step assumes clean HEAD
git log --oneline -1        # confirm you're building what you think you're building
# bump IMAGE_TAG in makefile to the NEXT free value — check fresh, don't reuse 1260/1261
make build-agent-chassis
make push-agent-chassis
make deploy-agent-chassis   # or however this fleet's release step is invoked — check makefile targets
```
Per this repo's own release norm: **releases are whole-fleet** (`make
release`, run by the owner) — a one-service apply at its own tag may not be
what's wanted here. Ask if unsure rather than guess.

## Step 4 — verify the image is actually live, on the pod, not the tag

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].image}{"\n"}{end}'
# then, for EACH pod:
kubectl -n ai-persona-system exec <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "ensure_page_section_layout"; \
   strings /app/agent-chassis | grep -c "business_directory"; \
   strings /app/agent-chassis | grep -c "directory-build-handler"'
```
All three must be non-zero on **both** replicas before you go further — a
same-tag rebuild ships the node's stale cached binary, and this exact check
returned all-zero on the pre-existing image when this was last measured
(2026-08-07 pre-work).

Also update `326`'s `image_tag` (or re-run the migration with the corrected
tag) if you didn't already in Step 2.

## Step 5 — re-triage the two named work items (only after Step 4 passes)

**This is the actual "build the pages through the framework" step.** Not a
new dispatch script — these are ordinary `site_work_items` rows; the point
of this whole change is that they become dispatchable the normal way.

```sql
-- directory-index: point it at the new handler and re-queue
UPDATE site_work_items
SET handler_agent = 'directory-build-handler',
    status = 'triaged',
    attempt_count = 0,
    error = NULL,
    updated_at = NOW()
WHERE id = '715ec305-1de1-4901-b988-b4880d58cce9';

-- guides-index: needs NO new handler (page-build-handler already, per
-- load_work_item_actions.go's default) -- it was blocked on the SAME empty-plan
-- problem, and ensure_page_section_layout fixes that generically, not just
-- for entity-directory. Re-queue it the same way, handler unchanged.
UPDATE site_work_items
SET status = 'triaged',
    attempt_count = 0,
    error = NULL,
    updated_at = NOW()
WHERE id = '2f50bfda-0e2f-4a2d-bb14-22f76114f092';
```
**Before running these**, re-check the rows haven't moved (another session
may have touched them in the intervening time):
```sql
SELECT id, status, handler_agent, attempt_count, error
FROM site_work_items WHERE id IN
  ('715ec305-1de1-4901-b988-b4880d58cce9', '2f50bfda-0e2f-4a2d-bb14-22f76114f092');
```

`build-pipeline-trigger` (already live, 120s cadence) will pick these up on
its own — no manual dispatch needed. Watch:
```sql
SELECT id, status, current_step, updated_at FROM orchestration_states
WHERE created_at > NOW() - INTERVAL '10 minutes' ORDER BY created_at DESC LIMIT 10;
```

## Step 6 — verify the DEPLOYED pages, not the status

Per this repo's own recurring lesson (`complete` is not proof the work
happened):
```bash
curl -sI https://vetcomparison.uk/directory/index.html   # last-modified should be ~now
curl -sI https://vetcomparison.uk/guides/index.html
```
Then fetch the body and confirm: `directory/index.html` shows real business
names/postcodes (not fabricated, not empty — if `ensure_page_section_layout`
found no `directory-export-json` config it would resolve to an empty list,
which is a legitimate-but-different failure mode worth distinguishing from
"it worked"); `guides/index.html` lists the three real guide pages
(`guide-cma-compliance`, `guide-cma-market-investigation`,
`guide-independent-strategy`), not fabricated ones.

```sql
SELECT status, claimed_at, completed_at, error_message FROM site_work_items
WHERE id IN ('715ec305-1de1-4901-b988-b4880d58cce9', '2f50bfda-0e2f-4a2d-bb14-22f76114f092');
```

## After a clean live run

- Update this lane's `bugs_open/206` file: close it (→ `bugs_closed/`) once
  BOTH pages are proven live, per this repo's bar ("fixed AND live").
- Update the concept register's BLD-017 entry status line from "built... NOT
  live" to "deployed, proven live \<date\>", with the same evidence style
  BLD-015/016 use (orchestration state, deployed artefact check).
- Tell the owner directly/honestly whether the "alphabetical vet list"
  complaint from the ORIGINAL ask (features_open/021's session) is now
  actually resolved — it should be, once directory-index shows a real,
  claim-status-aware business list instead of nothing.
- `entity-page` (individual practice pages) stays deliberately unbuilt —
  don't build ahead of P1's company-number crawl actually finishing
  (10/~2,109 done as of 2026-08-06; re-check before assuming it's still
  stalled).

# HANDOFF 2026-08-07 — continue here (`bugs_open/206`, directory-build-handler)

**Read `PLAN_2026-08-06_directory_build_handler.md` first** (design + why),
then `NOTES_directory_build_handler.md` (the full trail — a same-file
passenger that landed harmlessly, and a full council round 1 REVISE that got
fixed and resubmitted). This file is the concrete next-steps checklist.

## State in one line

Code is **committed** (`f750595dd`, then `37560f120` fixing council round
1's REVISE) and **confirmed live on the running pods** (`v1.0.1262` — rolled
by another session; pod-grepped both replicas for `ensure_page_section_layout`,
`business_directory`, `directory-build-handler`, all present, negative
control absent). **Council round 2 verdict has not been read yet** — same
correlation, resubmitted after fixing round 1's objections. DB config is
**not applied**: `agent_definitions` has no `directory-build-handler` row,
`directory-listing`'s schema is unchanged. Nothing has been dispatched
against `vetcomparison.uk`. Re-verify every claim below against the live
system before acting — this tree has moved fast enough that IMAGE_TAG alone
shifted three times and the Go code got rolled by a session that wasn't this
one, all *during* the writing of this one fix.

## Step 0 — re-orient, don't trust this file's numbers blindly

```bash
git log --oneline -5   # confirm 37560f120 is still there, see what's landed since
git status --short     # this tree is heavily concurrent; expect noise from other sessions
kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].image}{"\n"}{end}'
# re-pod-grep even if the tag above still says v1.0.1262 -- a same-tag rebuild
# ships a stale cached binary and the tag alone proves nothing:
kubectl -n ai-persona-system exec <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "ensure_page_section_layout"; \
   strings /app/agent-chassis | grep -c "business_directory"; \
   strings /app/agent-chassis | grep -c "directory-build-handler"'
```

## Step 1 — read the council verdict (round 2, same correlation)

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='5b8e4cf7-31c3-4793-a550-d6b9be1f00e8' AND kind='council_report'
ORDER BY created_at;
-- two rows expected once round 2 lands: 2026-08-07 01:31 = revise (round 1,
-- already read and fixed), a LATER one = round 2's actual answer.
-- Full per-reviewer detail (NOT the truncated doc_notes summary):
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='5b8e4cf7-31c3-4793-a550-d6b9be1f00e8' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```
If only the one 01:31 row is there, round 2 is still queued/running — the
dispatch lane was visibly contended both times this was submitted; this
repo's own norm is budget ~30 minutes, not ~2, before treating "no new row
yet" as anything other than latency.

- **APPROVED** → when you commit the *next* piece of this lane's work (the
  re-triage in Step 5, or any further code), use
  `Council-Reviewed: 5b8e4cf7-31c3-4793-a550-d6b9be1f00e8` on that commit —
  neither `f750595dd` nor `37560f120` need amending (forward-only; 098
  credits both automatically at report time via the `Council-Submitted`
  trailer already on them).
- **REVISE again** → read what's still outstanding. Round 2 fixed the
  gating objection (silent empty result) and two other HIGH-severity ones
  (missing spawn before `call_agent`; the shared-writer duplication three
  reviewers flagged). Several MEDIUM/LOW objections were answered with
  evidence rather than code changes (see round 2's `grounded_in` — fleet-wide
  blast radius is exactly 1 pending work item; no pre-existing
  `directory-build-handler` row; no existing `queryresolve` case already
  covers this). If round 2 gates on one of THOSE, that means the evidence
  wasn't convincing enough on its own — read the actual objection text
  before assuming a code change is even needed. Fix in a new commit (never
  amend), resubmit with
  `RESUBMIT_CORR=5b8e4cf7-31c3-4793-a550-d6b9be1f00e8` so the trail
  accumulates.
- **REJECTED** → a guardian veto with a named contained alternative. Read it
  — **filter by THIS correlation, the bare `doc_notes` query returns
  whatever ran most recently fleet-wide, not necessarily this one** (learned
  the hard way this session):
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%5b8e4cf7%' ORDER BY created_at DESC LIMIT 1;`
  Read it before touching anything else in this lane — the code is already
  on the shared branch either way; a REJECTED verdict doesn't un-ship it, it
  tells you what to do differently next.

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

**326's `image_tag` field is stale (`v1.0.1260`) — fix before or right after
applying.** The Go code is ALREADY live: another session built and rolled
`v1.0.1262` mid-session, pod-grepped and confirmed carrying
`ensure_page_section_layout`/`business_directory`/`directory-build-handler`
(2026-08-07). **Re-confirm this is still true** (Step 0's pod-grep) — if the
round-2 fix (spawn step, error-not-empty) isn't ALSO in the running binary
(it was committed in `37560f120`, AFTER whatever got rolled to 1262 — check
which commit that build was cut from before trusting it), you need a fresh
build+roll carrying `37560f120` before Step 5, not just any green pod-grep.
Once you know the real live tag:
```sql
UPDATE agent_definitions SET image_tag='<real tag>' WHERE type='directory-build-handler';
```

## Step 3/4 — build+roll, ONLY if the pod-grep in Step 0 doesn't already show round 2's fix live

If the current pods do NOT carry the round-2 fix (check: the spawn-then-call
shape is DB config in migration 326, not gettable by pod-grep; but
`resolveBusinessDirectory`'s error-not-empty behaviour IS a Go change —
grep for the exact string `"cannot distinguish this from a real zero-business result"`
on both replicas; non-zero means round 2's Go fix is live):

```bash
git log --oneline -1        # confirm you're building what you think you're building
# bump IMAGE_TAG in makefile to the NEXT free value — check fresh
make build-agent-chassis
make push-agent-chassis
make deploy-agent-chassis   # or however this fleet's release step is invoked — check makefile targets
```
Per this repo's own release norm: **releases are whole-fleet** (`make
release`, run by the owner) — a one-service apply at its own tag may not be
what's wanted here. Ask if unsure rather than guess.

Then re-verify on the pod, both replicas, all of: `ensure_page_section_layout`,
`business_directory`, `directory-build-handler`, AND the round-2 marker
string above — plus a negative control. Do not proceed on tag alone.

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

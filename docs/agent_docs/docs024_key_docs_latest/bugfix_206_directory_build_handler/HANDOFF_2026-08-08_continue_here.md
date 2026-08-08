# HANDOFF 2026-08-08 — continue here (`bugs_open/206`, directory-build-handler)

**Supersedes `HANDOFF_2026-08-07_continue_here.md`** (keep it — it's still an
accurate account of rounds 1/2, just not the next-steps list any more).
**Read `NOTES_directory_build_handler.md`'s 2026-08-08 entry first** — it has
the full trail for everything below, cited.

## State in one line

Code is **committed** (`f750595dd`, `37560f120`, `528f545f6`) and **confirmed
live on the running pods** (`v1.0.1263` — pod-grepped both replicas for
`ensure_page_section_layout`/`business_directory`/`directory-build-handler`
AND the round-2 marker string, all present, negative control absent).
**Council round 3 is submitted, verdict not yet read** — same correlation as
rounds 1/2, third resubmission, fixing/evidencing every round-2 objection.
DB config is **still not applied**: `agent_definitions` has no
`directory-build-handler` row, `directory-listing`'s schema is unchanged.
Nothing has been dispatched against `vetcomparison.uk`. **Re-verify
everything below against the live system before acting** — this tree moves
fast; the chassis image tag alone changed under this lane twice across two
sessions without either session triggering the roll itself.

## Step 0 — re-orient, don't trust this file's numbers blindly

```bash
git log --oneline -5   # confirm 528f545f6 (326 guard fix) is still there
git status --short     # this tree is heavily concurrent; expect noise
kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].image}{"\n"}{end}'
# re-pod-grep even if the tag still says v1.0.1263 -- tag alone proves nothing:
kubectl -n ai-persona-system exec <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "ensure_page_section_layout"; \
   strings /app/agent-chassis | grep -c "business_directory"; \
   strings /app/agent-chassis | grep -c "directory-build-handler"; \
   strings /app/agent-chassis | grep -c "zero-business result"'
# expect 5 / 3 / 1 / 1 on BOTH replicas (measured 2026-08-08 on v1.0.1263).
# A different count is not necessarily wrong, but the round-2 marker
# ("zero-business result") going to 0 means round 2's fix has regressed --
# stop and re-diagnose before proceeding.
```

## Step 1 — read the council verdict (round 3, same correlation)

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='5b8e4cf7-31c3-4793-a550-d6b9be1f00e8' AND kind='council_report'
ORDER BY created_at;
-- THREE rows expected once round 3 lands: 01:31 = revise (round 1), 08:28 =
-- revise (round 2), a LATER one (submitted ~2026-08-08 10:11) = round 3.
-- Full per-reviewer detail (NOT the truncated doc_notes summary):
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='5b8e4cf7-31c3-4793-a550-d6b9be1f00e8' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```
Run orchestration to watch directly if the row isn't there yet:
`c7f494a4-de65-43e9-9fec-62d635e871e5`. Queue was clear (LAG 0) at submit
time and a run was already `EXECUTING_STEP` within 11 seconds, so this round
should land faster than rounds 1/2 did — but this repo's own norm is still
budget ~30 minutes, not ~2, before treating "no new row yet" as anything
other than latency.

- **APPROVED** → commit the *next* piece of this lane's work (Step 2 onward)
  with `Council-Reviewed: 5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`. Nothing
  already committed (`f750595dd`/`37560f120`/`528f545f6`) needs amending
  (forward-only; 098 credits all three automatically via their
  `Council-Submitted` trailers).
- **REVISE again** → read what's still outstanding at `body`, not just
  `metadata->>'decision'`. Round 3's submission
  (`/home/ant/.claude-scratch/.../submission_206_round3.json` — **that path
  was this session's own scratchpad and will not exist in a new session**;
  reconstruct from `NOTES_directory_build_handler.md`'s 2026-08-08 entry if
  you need the exact text again) answered every round-2 objection either
  with a code fix (326's `IS DISTINCT FROM` guard) or with grounded evidence
  (the sibling audit, the fresh dispatch-gate check, the processing_mode
  check). If round 3 gates on something NEW, it means a fourth reviewer
  found something rounds 1–3 didn't — read it before assuming a mechanical
  fix. Fix in a new commit (never amend), resubmit with
  `RESUBMIT_CORR=5b8e4cf7-31c3-4793-a550-d6b9be1f00e8`.
- **REJECTED** → a guardian veto. Read it filtered by THIS correlation (the
  bare query returns whatever ran most recently fleet-wide):
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%5b8e4cf7%' ORDER BY created_at DESC LIMIT 1;`

## Step 2 — apply the two migrations (only after reading the round-3 verdict)

**Do NOT run `--apply`** — it takes EVERY pending file in the directory.
Apply just these two, in order, by hand:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/325_directory_listing_binds_to_business_directory_query.sql

kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/326_directory_build_handler_agent.sql
```
326's guard now uses `IS DISTINCT FROM` (fixed 2026-08-08, `528f545f6`) — a
clean run with `326 OK: ...` and no exception means it genuinely verified,
not just that the block didn't error. Then register both so a future
dry-run doesn't think they're still pending:
```bash
./scripts/migration/run-migrations.sh --record-only \
  325_directory_listing_binds_to_business_directory_query.sql \
  --note "applied by hand 2026-08-08, verified via its own DO block"
./scripts/migration/run-migrations.sh --record-only \
  326_directory_build_handler_agent.sql --note "applied by hand 2026-08-08, verified via its own DO block (IS DISTINCT FROM guard)"
```

**326's `image_tag` field is stale (`v1.0.1260`).** The Go code is live on
`v1.0.1263` as of this handoff (Step 0's pod-grep) — **re-confirm this is
still the current tag** before trusting it, then:
```sql
UPDATE agent_definitions SET image_tag='<current live tag>' WHERE type='directory-build-handler';
```

## Step 3/4 — build+roll: almost certainly NOT needed

Step 0's pod-grep already covers this — as of 2026-08-08 the running image
(`v1.0.1263`) carries round 2's fix (the `zero-business result` marker).
Only rebuild if Step 0's fresh check shows that marker has regressed to 0,
or if a commit landed in this lane AFTER whatever tag is currently live and
you need to verify it separately. If you do rebuild: bump `IMAGE_TAG` in the
makefile to the next free value, `make build-agent-chassis` →
`push-agent-chassis` → `deploy-agent-chassis` (or ask about `make release`
per this repo's whole-fleet norm), then repeat Step 0's pod-grep on both
replicas before trusting it.

## Step 5 — re-triage the two named work items (only after Step 2/4 pass)

Ordinary `site_work_items` rows — the point of this whole change is that
they become dispatchable the normal way, no manual dispatch script.

```sql
-- Before running: re-check the rows haven't moved (another session may
-- have touched them since 2026-08-08).
SELECT id, status, handler_agent, attempt_count, error
FROM site_work_items WHERE id IN
  ('715ec305-1de1-4901-b988-b4880d58cce9', '2f50bfda-0e2f-4a2d-bb14-22f76114f092');
-- Expected as of 2026-08-08: both status='needs_human_review', handler_agent
-- (715ec305: page-build-handler; 2f50bfda: page-build-handler),
-- attempt_count=1, max_attempts=3, approval_mode='auto', depends_on NULL.

-- directory-index: point it at the new handler and re-queue
UPDATE site_work_items
SET handler_agent = 'directory-build-handler',
    status = 'triaged',
    attempt_count = 0,
    error = NULL,
    updated_at = NOW()
WHERE id = '715ec305-1de1-4901-b988-b4880d58cce9';

-- guides-index: NO new handler needed (page-build-handler already, per
-- load_work_item_actions.go's default) -- it was blocked on the SAME
-- empty-plan problem, which ensure_page_section_layout fixes generically.
UPDATE site_work_items
SET status = 'triaged',
    attempt_count = 0,
    error = NULL,
    updated_at = NOW()
WHERE id = '2f50bfda-0e2f-4a2d-bb14-22f76114f092';
```

**Fresh-checked 2026-08-08** (this handoff's own diagnosis, see NOTES): the
dispatch selector (`find_dispatchable_site`) and loader (`LoadWorkItemsAction`
via `build-dispatch-loop`'s `load_items` step) carry identical
item-eligibility predicates today, and vetcomparison.uk's site row is
unlocked with no claimed items — both rows will be picked up once re-triaged,
by `build-pipeline-trigger` (already live, 120s cadence), no manual dispatch.
Watch:
```sql
SELECT id, status, current_step, updated_at FROM orchestration_states
WHERE created_at > NOW() - INTERVAL '10 minutes' ORDER BY created_at DESC LIMIT 10;
```

## Step 6 — verify the DEPLOYED pages, not the status

```bash
curl -sI https://vetcomparison.uk/directory/index.html   # last-modified should be ~now
curl -sI https://vetcomparison.uk/guides/index.html
```
Fetch the body and confirm: `directory/index.html` shows real business
names/postcodes (empty is a legitimate-but-different outcome if
`ensure_page_section_layout` found no `directory-export-json` config —
distinguish it from success); `guides/index.html` lists the three real guide
pages (`guide-cma-compliance`, `guide-cma-market-investigation`,
`guide-independent-strategy`), not fabricated ones.

```sql
SELECT status, claimed_at, completed_at, error_message FROM site_work_items
WHERE id IN ('715ec305-1de1-4901-b988-b4880d58cce9', '2f50bfda-0e2f-4a2d-bb14-22f76114f092');
```

## After a clean live run

- Close `bugs_open/206` → `bugs_closed/` once BOTH pages are proven live
  (this repo's bar: "fixed AND live").
- Update the concept register's BLD-017 entry status line to "deployed,
  proven live \<date\>", same evidence style as BLD-015/016.
- Tell the owner directly whether the "alphabetical vet list" complaint from
  the original ask (`features_open/021`) is now actually resolved.
- `entity-page` stays deliberately unbuilt — don't build ahead of P1's
  company-number crawl (10/~2,109 done as of 2026-08-06; re-check before
  assuming still stalled).

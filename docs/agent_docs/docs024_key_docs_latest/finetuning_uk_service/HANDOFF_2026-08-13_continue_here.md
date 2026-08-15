# HANDOFF 2026-08-13 — 259 LIVE, 258 fixed-not-live, and three things blocked on a token

> ## ⛔ SUPERSEDED by `HANDOFF_2026-08-15_continue_here.md`
> All three blocked items are done: migration **400** applied, 258 submitted to the
> council (`d24f9829-…`), landmines synced. 258's fixes are now **LIVE** at
> `v1.0.1301` / `0115f2b45`. One step remains — the unpause + Phase 0 run.

**COLD-START document for the lane.** Supersedes
`HANDOFF_2026-08-12b_continue_here.md` (accurate, but its "next steps" are done).
PLAN has the approved design, RUNBOOK the commands (**§5** claims, **§6** sizing +
the wait deadline), NOTES the evidence trail, README the owner's plain-prose log.

## ⛔ Provisioning is STILL PAUSED, and must stay so

```sql
SELECT is_paused, pause_reason FROM thunder_config;   -- t | '...259... Fix committed 10659b419 ...'
```

`pause_reason` was rewritten 2026-08-12 to name the corrected cause. It stays
paused until **258's** fixes are live too — see "Next steps".

## Where the two bugs stand

| | state |
|---|---|
| **259** — one request, several billable GPUs | **FIXED AND LIVE.** thunder-adapter `v1.0.1295`, stamp `69612d692`; `git merge-base --is-ancestor 10659b419 69612d692` passes. Council **APPROVED** (`20d8b725`). Migration 396 applied. **No live behavioural proof yet** — needs a real provision, so the file stays OPEN. |
| **258** defect 1 (vcpus) + defect 2 (wait deadline) | **FIXED, NOT LIVE** (`236810e4e`). Needs migration 400 applied, a council round, a build and a fleet release. |
| **258** defect 3 (no record of a failed provision) | **FIXED AND LIVE** — a side effect of 259's claims table. |

## Next steps, in order

1. **OWNER: refresh the kubeconfig token.** It expired **2026-08-13 18:05:20Z**
   (decoded from the JWT `exp` in `~/.kube/config_production_uk001`; every
   `kubectl` returns `Unauthorized`). Everything below needs it.
2. **Apply migration 400** (`provision_wait_timeout_seconds`). Not
   ordering-critical — the column is read via `to_jsonb`, so an unmigrated
   database degrades to the compiled-in default rather than failing. Scope the
   runner to the one file:
   ```bash
   mkdir -p /tmp/m400 && cp docs/agent_docs/sql_for_agents/400_thunder_provision_wait_timeout.sql /tmp/m400/
   MIGRATIONS_DIR=/tmp/m400 ./scripts/migration/run-migrations.sh            # dry run first
   MIGRATIONS_DIR=/tmp/m400 ./scripts/migration/run-migrations.sh --apply
   ```
3. **Submit 258 to the council.** The submission is written and committed:
   ```bash
   ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
     docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/council_submission_258_provision_defaults.json
   ```
   Save the `SUBMISSION_CORR`. The code is already on the shared branch, so a
   REVISE needs a follow-up commit, not a hold. **Push back is expected on the
   wait/await coupling being a stated invariant rather than a mechanical one** —
   that is flagged in the submission's own risks, deliberately.
4. **`./scripts/landmines-sync.py --apply`.** Two entries were added on 2026-08-13
   and could not be synced, so they are **file-only**: seats and agents cannot
   read them yet.
5. **`make build-thunder-adapter`** from committed HEAD, then the whole-fleet
   `make release` — **the OWNER runs the release.** Verify per service:
   ```bash
   kubectl -n ai-persona-system logs -l app=thunder-adapter --tail=300 | grep -m1 'build provenance'
   git merge-base --is-ancestor 236810e4e <that sha>    # exit 0 = 258's fixes are in
   ```
6. **Then unpause and run Phase 0.** This is the first run where a provision can
   actually succeed. Everything else is staged: bundle live in B2 (md5
   `a19557ccf61ac951c28e81254a8d76f7`), dataset
   `finetuning/datasets/phase0-2026-08-12/training.jsonl` (300 rows), presign
   proven. No `vcpus` override needed any more.
   **What to check, and it is two things at once:**
   - **258:** the a6000 reaches `running` and is **not** deleted. The adapter logs
     `Resolved vCPU count from Thunder specs spec_key=a6000_x1_prototyping vcpus=6`.
   - **259's live proof:** one claim row per correlation. If the provision
     succeeds first time the retry driver never fires, so 259 gets **no** live
     proof from this run and still needs a deliberate slow case later.
   - **And finally measure how long an a6000 actually takes to boot.** Still
     unknown: both prior attempts were killed at 5 min while `STARTING`, so
     `> 5 min` and nothing more. This run is the first chance to find out.
7. **Phase 1** (offer page + payment link) — BLOCKED on the front-end thread
   (`finetuning_uk_repair`, `7b4e88a8-…`). Owner calls outstanding: final price,
   playground booking shape, sample datasets, Stripe posture.

## The thing most likely to bite the next session

**Raising `provision_wait_timeout_seconds` above `dispatch_provision`'s
`timeout_seconds` (600s) does not give you a longer wait — it gives you a quiet
success.** The await expires first, the retry driver re-dispatches, 259's guard
refuses the duplicate *correctly*, and the workflow reports **FAILED while a real
billed instance runs on with nobody watching it.** Default is 540s for that
reason. **To go higher, raise the STEP first, then the column.** In `LANDMINES.md`,
RUNBOOK §6, migration 400's header, the column `COMMENT`, and a test.

The tell, if it has already happened:
```sql
SELECT c.correlation_id, c.attempts, c.status, c.thunder_instance_id
FROM thunder_provision_claims c
JOIN orchestration_states o ON o.correlation_id = c.correlation_id
WHERE o.status='FAILED' AND c.status IN ('created','succeeded');
-- any row = a box nobody is watching. Check the vendor, not just our tables.
```

## Open questions this lane still cannot answer

- **The real a6000 price.** Advertised $0.35/hr; its minimum is 6 vCPU and the
  page charges +$0.04/vCPU/hr beyond 4. Whether $0.35 already assumes 6 is
  `[UNVERIFIED]` — floor is **$0.35–$0.43/hr** until an invoice settles it.
- **a6000 boot time.** `> 5 min`, nothing more. Step 6 measures it.
- **Does an error response clear an await?** Every await in the 259 evidence
  expired on its own clock, never on the adapter's prompt error response — which
  is *why* the retry loop ran at all. **Undiagnosed; worth a `090`.** Recorded in
  `RFC_026` §6.

## Fleet-wide records from this lane, 2026-08-12/13

`bugs_open/258` · `bugs_open/259_…_billable_gpus` · **`architecture_review/RFC_026`**
(the retry driver re-executes side-effecting actions — 54 live `call_agent` steps
across 33 agents; needs an owner ruling) · register **FTW-043**, **FTW-044** ·
four `LANDMINES.md` entries (two unsynced) · two `WRONG_CALLS.md` entries ·
migrations **396** (applied), **400** (not applied).

## Two pre-existing problems, neither this lane's

- **`internal/adapters/thunder/api/client_test.go` does not compile at HEAD**
  (`unknown field Identifier in struct literal of type Instance`). Untouched.
- **`adapter.go:393` swallows a reply-produce error** (`silent-reply-drop`).
  Adoption of `DeliverReply` beyond webscrape is RFC-gated (`bugs_open/158`
  item 1) — **do not fix casually.**

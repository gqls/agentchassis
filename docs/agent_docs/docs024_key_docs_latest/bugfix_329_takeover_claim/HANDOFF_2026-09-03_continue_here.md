# HANDOFF — `bugs_open/329`, the stale-orchestration takeover claim

**Written 2026-09-03 ~16:1xZ. Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_329_takeover_claim/`.**
Read this first, then `NOTES_takeover_claim.md` (the technical log, newest at the bottom) and
`PLAN_2026-09-03_takeover_claim.md` (the design and why the alternatives lose).

---

## 1. State in one paragraph

`bugs_open/329` — both orchestration takeover arms judged a row "stuck" on a 5-minute clock and
resumed it with **nothing claiming it**, so two actors could run the same step and its external side
effects twice. **The fix is written, council-APPROVED, committed, and LIVE on `v1.0.1359` as of
2026-09-03 ~13:28Z** (proven at both pods' binaries with a three-way control — §4). It has **not yet
been exercised**: no takeover has been claimed since the roll. The bug is **NOT closeable yet** — see
§6 for the exact bar and why.

## 2. What was actually wrong — the filed cause was FALSE, do not rebuild on it

> ⚠ `bugs_open/329` stated `[MEASURED 2026-08-19]` that these writes were unversioned:
> *"ends in `r.UpdateState(...)`, **not** `UpdateStateWithVersion`."* **False.** `state.go:883-885`
> is a one-line delegating wrapper — `UpdateState` **IS** `UpdateStateWithVersion`. Its **fix
> candidate (2) was a no-op.** The premise came from `bugs_closed/294`, where it was used to answer a
> council guardian objection, and travelled from there. **Both files are corrected in place**; the
> trap is in `LANDMINES.md`.

**The real defect is a CHECK-THEN-ACT ACROSS TWO READS.** The arm judged staleness on the *caller's
snapshot*; the write that followed did its **own fresh** `GetState` → mutate → version-CAS which
**never re-tested the predicate**. Every write was properly versioned and two takers still both
proceeded, because each CAS-ed against the version it had just read. The guard was present and
answering a different question — *"has this row changed since I read it?"* rather than *"is this row
still stale?"*.

## 3. The fix

`StateRepository.ClaimStaleOrchestration` re-judges the **fresh** row **inside**
`ExecuteWithOptimisticLocking` and writes only if it is still stale. The write **is** the claim:
`UpdateStateWithVersion` stamps `last_activity = now` (`state.go:1051`) and `version + 1`
(`:1074`), so every later taker's fresh read declines with `ErrTakeoverLost`. Both arms route through
`SagaCoordinator.takeOverStaleOrchestration`, which returns **nil** on a lost claim — the same
disposition as the arms' pre-existing non-stale branch. `ClearExecutingStep` is deleted; its sole
caller was the arm.

No new SQL, no new column, `processing_node` untouched — so the council invariant that
`orchestration_states`' two guarded mechanisms must never govern the same column
(`state_locks_test.go`, corr `4a227ed9`) is preserved **by composition**, not by promise.

**Deliberately NOT built on `TakeOverOrchestration`**, despite both bug files recommending it: its
CAS is `WHERE processing_node = $3` from the **observed** value, so where a row already carries the
acting pod's own name two callers in that pod both match and both report `rowsAffected = 1` — no
exclusion at all. The two methods are now distinguished **in the code, at both sites** (council
`reuse_agent [medium]`).

## 4. Proof it is live — reproduce this before trusting anything below

⚠ The `build provenance` startup line **has already scrolled** on both pods. **An empty result there
means "not in range", not "unstamped".** Use the binary probe, which has no shelf life:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -aq "Orchestration is actively executing"   /proc/1/exe  # CONTROL: must be PRESENT
kubectl -n ai-persona-system exec $POD -- grep -aq "STALE_TAKEOVER_CLAIMED"                /proc/1/exe  # FIX:     PRESENT <=> shipped
kubectl -n ai-persona-system exec $POD -- grep -aq "Found stuck orchestration, taking over" /proc/1/exe # OLD:     PRESENT <=> NOT shipped
```

Result on **both** pods `[MEASURED 2026-09-03 ~16:0xZ]`: **PRESENT / PRESENT / ABSENT.** Run all
three — a broken `grep -aq` returns ABSENT for everything and reads as "the fix is not there". Probe
**both** pods: one release tag can ship several revisions (`bugs_open/249`).

## 5. ⚠ THE OBVIOUS METER IS BROKEN — do not read the pod logs

I grepped the logs for the new needles and got **0 and 0**. It means nothing: **the demand control
was also 0.** `kubectl logs -l app=agent-chassis --since=3h --tail=200000` returns **68 LINES TOTAL**
on a service that created **1,588 orchestration rows in that window** `[MEASURED 2026-09-03 ~16:0xZ]`.
Any zero from that log is uninformative and would read identically before the fix existed.

**The durable needle is the meter, and it exists precisely because of this:**

```sql
SELECT count(*) FROM orchestration_states
WHERE processing_history @> '[{"action":"stale_takeover_claimed"}]';
```
**0 as of 2026-09-03 ~16:0xZ.** That zero **is** trustworthy — `processing_history` is persisted, not
windowed. It means no takeover has been claimed since the roll, which is consistent with everything
else measured. Before this fix, a takeover left **no durable trace at all**, so "have the arms ever
fired?" was unanswerable outside a live log window. It is answerable now.

## 6. THE NEXT ACTION — and why the bug is not closeable yet

The bar for `bugs_closed/` is **fixed AND live**. It is fixed and live. It is **not verified working**,
and I am not closing it on a deploy probe alone (`MEMORY`: prove it at the artefact, and a `complete`
status is not a repaired artefact).

**What is owed, in order:**

1. **Watch the needle.** Re-run the query in §5 daily. The first non-zero row is the first time this
   mechanism has ever demonstrably run. Read that row's `processing_history` entry: it carries
   `pod_name`, `step_id` and a `details` string of the form `idle 6m30s > 5m0s in EXECUTING_STEP`.
2. **Then check it did the right thing**, which is the part a count cannot tell you: for that
   orchestration, did the step complete once, or twice? `execution_path` and any external side effect
   of that step are the evidence.
3. ⚠ **Do NOT try to induce it on the dispatch path.** Three guards sit in series — (i) the chassis
   intake serialisation claim (`ClaimSerialisationKey`, keyed on the orchestration_id, `agent-chassis`
   only), (ii) these arms, (iii) per-path CASes such as the work-item claim in
   `claim_work_item_action.go`. **Guard (iii) absorbs a double-takeover even with the fix reverted**,
   so a green there proves nothing. An induced fault must **bypass** the intake claim, and that
   vehicle needs the owner's sign-off (`bugs_open/075` §2 was refused on exactly this shape).
4. **If the needle is still 0 after a week**, that is itself a finding — say so with the traffic
   control beside it (`orchestration_states` created in the window), never as a bare zero.

## 7. What is deliberately NOT closed — now its own bug

**`bugs_open/461`** (filed from this lane, **OPEN, UNOWNED**): a live driver holds **nothing**, so a
300 s clock can judge a correctly-working orchestration dead. `defaultLocalActionTimeout = 7200 s`
(`coordinator.go:1246`) and nothing refreshes `last_activity` during a local action — a **24× gap**.
Post-329 the bound is **2** concurrent actors (driver + exactly one taker), down from unbounded.
**2 is not 1**, and on a content-writing workflow the two are not equivalent (the `bug_historian`
seat's point, quoted in the file). Widening the threshold is rejected there: no value is both safe
and useful. The heartbeat candidate is flagged **architecture-scope**, not a bug patch.

⚠ 461 says explicitly **not to size it until §5's query has data** — today it necessarily returns 0.

## 8. Council trail

`Council-Reviewed: 3beb3f54-6d51-42fd-969f-78e4ea871659`. **Round 1 REVISE** (gated by
`prior_art_librarian`), **round 2 APPROVED** — *"approved with 2 advisory objection(s) — none
high-severity"*, 9 of 11 seats; `architecture` approved **both** rounds, settling the scope question
(guarantee **narrowed**, not widened, 2026-07-29 §1). Submissions are in the lane
(`council_submission_329_r1.json`, `_r2.json`).

⚠ **Read a verdict by YOUR correlation**, never `doc_notes … ORDER BY created_at DESC LIMIT 1` — that
returns the most recent council note **on the fleet**, and I read another lane's REVISE as mine:

```sql
SELECT body FROM diagnosis_artifacts
WHERE correlation_id LIKE '3beb3f54%' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```
Then confirm its `plan summary` describes **your** change before believing a word of it.

## 9. Traps this lane paid for — all in `LANDMINES.md`, footprinted

1. **`UpdateState` IS a version CAS**, and two bug files say it is not. Believing them costs a no-op fix.
2. **A `sync.WaitGroup` start-line test shows the BROKEN code passing.** Simultaneous takers were
   always absorbed by the version CAS — the loser's CAS fails. The disconfirming case is the
   **SEQUENTIAL** interleaving. This inverts the canonical way to test a race.
3. **The code index cannot find callers.** `code_symbols.content` holds each symbol's **own body**, so
   `content ILIKE '%X%'` returns **1** for everything — measured controls: `NewStateRepository`,
   which **20 files** in the tree reference, also returns **1**, as do `UpdateStateWithVersion` and
   `handleOrchestrationStatus`. For "does anything still call this?", the
   instrument is `go build ./...` with the symbol deleted, plus greps **outside `.go`**.
4. **"No caller" and "no reference" are different questions.** Deleting `ClearExecutingStep` left a
   live `orchestration_status_vocabulary.written_by` row naming it (migration 466) — invisible to the
   compiler and to the index. Migration **736** corrected it (applied, verified, ledger-recorded).

## 10. Files this lane owns

| file | what |
|---|---|
| `platform/orchestration/state.go` | `ClaimStaleOrchestration`; `ClearExecutingStep` deleted; the two-takeover-methods distinction on both doc blocks |
| `platform/orchestration/coordinator.go` | `ErrTakeoverLost`, `takeOverStaleOrchestration`, both arms |
| `platform/orchestration/stale_takeover_claim_test.go` | five arm-level tests, incl. the positive and negative controls |
| `docs/agent_docs/sql_for_agents/736_*.sql` (+ `_ROLLBACK`) | the vocabulary row correction — **applied 2026-09-03** |
| `bugs_open/329_*` | the bug, with the 2026-09-03 update and both corrections |
| `bugs_open/461_*` | the residual, filed |
| register **WFA-025** (`docs026_concept_register/register/workflow-authoring.md`) | the mechanism |
| `016b` §9 | the transferable pattern (a guard present and answering a different question) |

**Commits:** `b55f837ef` (the fix) · `9b7197866` (461 + comment correction) · `593a2c170` (council
advisories actioned) · plus the lane docs, register, landmines and `WRONG_CALLS` rows.

## 11. Other threads

- **`dispatch_throughput`** (session `throughput`) — gave me guard (iii) and the double-handle census;
  I retracted a false exposure claim to them. They declined the 180 s-lease-vs-300 s-timeout finding
  on the correct ground that they have never measured it, so **it lives in `bugs_open/329` §5, cited
  to nobody**. Do not attribute it to them.
- **`chassis_replica_scaling`** (dormant) — their cross-worker safety inventory ("the DB claims,
  `UpdateStateWithVersion`, component locks, CS-1's guard") should now list this claim. Not yet told.
- **`orchestration_status_lifecycle`** (dormant) — filed 329 and disclaimed it; owns the RUNNING arm
  (commit `e34d44f26`, **not** "migration 465" — a migration cannot add Go, and 465 is an ambiguous
  number). Their `bugs_closed/294` carries my correction.

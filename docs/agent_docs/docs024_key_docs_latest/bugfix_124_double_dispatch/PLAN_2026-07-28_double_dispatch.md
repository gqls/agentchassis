# PLAN — `bugs_open/124`: one `needs_diagnosis` item runs twice

**Opened** 2026-07-28. **Bug** `bugs_open/124_HANDOFF_2026-07-28_one_diagnosis_item_runs_twice_under_two_correlations.md`.
**Ownership check before starting:** `scripts/who-owns.py 124` → no owning
workstream, one commit (the filing). The filing thread
(work-item-parallelisation) marked it *"OPEN, unowned. Separable from 029"* and
its own programme completed 2026-07-28. Taken up here.

---

## 1. The root cause — and the correction to the filed one

The bug was filed with two candidate mechanisms, both marked honestly (one
`[VERIFIED]`, one `[UNVERIFIED]`). **Neither is the cause.** The cause is a third
thing, and it is simpler:

> **There are TWO dispatchers for one queue.**
> `090_TRIGGER_needs_diagnosis_v1.sh` writes the intake item at
> `status='awaiting_diagnosis'` **and then publishes the orchestrate envelope
> itself**. Since `diagnose-pipeline-trigger` was enabled, `diagnose-dispatch-loop`
> claims that same `awaiting_diagnosis` item on its next 60s tick and runs a
> **second, independent diagnosis**. Both complete. Neither knows about the other.

The script says so in its own header — it just describes the world before the
loop was switched on:

> *"Automatic dispatch: `diagnose-dispatch-loop` + the `diagnose-pipeline-trigger`
> scheduled task claim awaiting_diagnosis items on a 60s tick. **The task ships
> DISABLED — enable it deliberately. Until then, and for any ad-hoc run, THIS
> SCRIPT is the dispatcher.**"*

Somebody enabled the task. Nobody updated the script. From that moment every
manual diagnosis has been paid for twice.

### Corrections to the filed mechanism — record these, do not silently drop them

> **CORRECTED 2026-07-28 (this thread).** `124` §Mechanism item 1 says *"Nothing
> marks a `needs_diagnosis` item complete on success"*, marked `[VERIFIED]`. It is
> **false**. It was inferred from a *print statement* in the 090 script
> ("closing it by hand until a diagnose dispatch loop exists"), not from the
> loop's config. The live `diagnose-dispatch-loop` row has a `mark_complete` step
> (`complete_work_item`) and it works: every 090-filed item since the loop went
> live sits at `complete` or `failed` with `claimed_by='diagnose-dispatch-loop'`.
> **What caught it:** reading the live `agent_definitions` row instead of the
> script's own prose — the standing "seed is history, live row is fact" rule,
> applied to a *comment* rather than a seed.

> **CORRECTED 2026-07-28 (this thread).** `124` §"What it already cost" and
> `029` §6 both say the loop *"re-dispatched an already-diagnosed item"* whose
> `call_handler` was sent *"43 minutes after that diagnosis finished"*, and
> present that as evidence for the item-closing gap. The orchestration they
> name, `41d64b75`, was **created at 20:08:16 — 91 seconds after the item was
> created**, i.e. it is the *concurrent duplicate*, not a later re-dispatch.
> Its `call_handler` request was still being *retried* at 20:49:31 (that retry
> is `bugs_open/029` proper). The conclusion — that the duplicate wrote `failed`
> over a completed diagnosis and inflates `029`'s rate — **stands and is
> strengthened**; only the "re-dispatch of a closed item" story is wrong.
> **What caught it:** `SELECT created_at FROM orchestration_states WHERE
> orchestration_id::text LIKE '41d64b75%'` — one query against the id the note
> already cited.

Both corrections point the same way: item 2 of the filed fix candidates ("make
the loop reuse `spec.correlation_id`") was ranked as cosmetic. It is not — see §3.

## 2. Evidence

Measured 2026-07-28 against the live cluster. Full queries in the RUNBOOK.

- `scheduled_tasks.enabled = true` for `diagnose-pipeline-trigger`. The concept
  register still carries `verify-later: should still be false unless deliberately
  turned on` — a stale expectation nobody re-checked.
- **Every** `needs_diagnosis` item inside the `orchestration_states` retention
  window shows the two-chain shape: one chain under the item's own
  `spec->>'correlation_id'` with **no** `diagnose-dispatch-loop` row (that is the
  script's direct publish), and one chain under a fresh correlation **with** a
  `diagnose-dispatch-loop` row (that is the loop). 6 of 6 items since 2026-07-27
  19:04. Older items fall outside retention and cannot be checked either way.
- Each duplicate is a full `diagnose-agent` LLM run: **12–14 minutes typical, up
  to 31**. This is the largest single avoidable LLM cost in the diagnosis lane.
- `created_by`: 21 items from `090_TRIGGER_needs_diagnosis`, 2 from
  `diagnosis-triage`. The triage-created ones are dispatched by the loop **only**
  — they were never double-run, and they expose the other half of the bug (§3).

### The throughput objection, checked and dismissed `[VERIFIED]`

Removing the direct publish makes the loop the only dispatcher, and the loop's
task carries `max_concurrent = 1`. That looks like it would serialise the fleet's
diagnoses behind one 13-minute run. It does not: `cmd/scheduler/main.go`
`stampCompleted` advances `last_completed_at` **immediately after publishing**
(fire-and-forget), and `countInFlight` counts *tasks*, not runs — so the group's
slot is free again on the next tick. Observed: dispatch-loop runs `04920015`
(11:25:51→11:38:57) and `2184add8` (11:33:22→11:46:35) **overlapped**. Two items
filed together are dispatched on consecutive ticks, ~60s apart, and run
concurrently. Dispatch latency also *improves*: the loop started the run at
+55s / +40s / +79s from intake, against +4m19s for the direct publish on the
same item, which queues behind the shared generic lane.

## 3. Why the join half must land in the same change

The script prints `SAVE: CORRELATION_ID=…` and the whole loop is documented to
join on that one key. Artifacts are keyed on the **envelope** correlation
(`params.ExecutionContext.CorrelationID`, `diagnose_assemble_bundle_action.go`),
not on the `correlation_id` field the loop passes through `input_mapping`.

So today, for a **loop-dispatched** item, `spec.correlation_id` names **nothing at
all**. It happens to resolve for 090-filed items *only because the script's own
duplicate chain runs under it*. Both `diagnosis-triage` items have a
`spec.correlation_id` that is pure fiction — a uuid minted by
`triageSpecJSON` and never used by any run.

**Therefore: delete the direct publish without fixing the join, and the printed
correlation stops resolving for everyone.** The fix that stops the waste and the
fix that keeps the trail joinable are one change, not two.

## 4. The fix — ordered by what closes the door

**P1 — the claim is the ticket (structural).** Any path that dispatches an intake
item must first take it out of the queue, atomically, with the same
`awaiting_diagnosis → diagnosing` UPDATE the loop uses. Whoever wins the claim
dispatches; whoever loses does not. This makes "two dispatchers, one item"
*unrepresentable* rather than merely unlikely, and it holds even if the two
dispatchers race inside the same second.

**P2 — one dispatcher, chosen by live state, not by memory.** The 090 script asks
the database whether the loop is live (`scheduled_tasks.enabled` **and** the
`diagnose-dispatch-loop` agent row active). Live → record the intake and let the
loop dispatch. Not live → dispatch directly, under P1's claim. No flag for an
operator to remember and no comment to go stale: *"operators must remember X" is
a defect in costume*. If the loop is ever disabled again, manual dispatch resumes
by itself.

**P3 — the run is joinable from the item, on every dispatch path (framework).**
Give `query_database` a generic way to bind its own run identity into SQL — a
reserved `$ctx.` parameter namespace (`$ctx.correlation_id`,
`$ctx.orchestration_id`, …). Then the loop's `claim_item` stamps
`spec.dispatch_correlation_id` in the same atomic UPDATE that claims the item.
One key joins item → `orchestration_states` → `diagnosis_artifacts` → `doc_notes`,
for loop dispatch and triage dispatch alike.

`$ctx.` is deliberately generic rather than a diagnose-specific column: *any*
queue-driven workflow that wants to record which run took a row has this problem,
and the alternative is a bespoke action per lane. Additive by construction — only
paths starting with `$ctx.` take the new branch, every existing param path
resolves exactly as before.

**Not doing:** correlation *override* on `call_agent`/`spawn_agent` (letting a
child chain adopt an upstream correlation). It would give a prettier trail, but
correlation is threaded through `awaited_requests`, response routing and the
Kafka partition key; changing it per-call is a large blast radius for a
cosmetic gain over P3's linkage row. Recorded here as considered and declined.

## 5. Verification

The failing shape is two `diagnose-agent` orchestrations for one item, minutes
apart, under different correlations. After the fix, one 090 run must produce
**exactly one** `diagnose-agent` orchestration, the item must reach a terminal
status with no hand-written `UPDATE`, and `spec->>'dispatch_correlation_id'` must
equal the correlation that chain ran under.

Verify against a **real** run, not an idle window (`029`'s note: the fleet reports
zero when nothing is running), and verify the Go half against the **pod**, not
git.

## 6. Landmine inherited from the bug file

Do **not** cancel or reap duplicate chains to tidy the queue before the evidence
is taken. A previous attempt on the neighbouring bug cancelled the failing
orchestration and reaped its pod, and the diagnosis loop then recorded that it
could find no specimen (`WRONG_CALLS.md`, 2026-07-27). The duplicate chains are
the evidence.

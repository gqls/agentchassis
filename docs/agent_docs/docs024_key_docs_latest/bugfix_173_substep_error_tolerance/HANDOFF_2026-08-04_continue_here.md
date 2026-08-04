# HANDOFF 2026-08-04 — `bugs_open/173`, continue here

**Read this first, then `NOTES_substep_error_tolerance.md` (newest entry at the bottom).**

## State in one paragraph

The fix is **written, committed, council-APPROVED (round 1) and LIVE on chassis
`v1.0.1250`**, pod-verified on both replicas with a positive *and* a negative control.
**One thing is owed before the bug can close: the live induction, both branches.** It is
**in flight** — two induction agents are seeded in the live DB and two runs were
dispatched, but no `orchestration_states` row had appeared when this session ended. That
is probably queue latency; it has **not** been distinguished from a dropped publish yet.
See § "Pick up exactly here".

**`bugs_open/173` is still OPEN and that is correct.** Do not close it on the roll alone —
its own bar is the induction, and the bar exists because a build cannot prove this.

## What shipped

One production file, `platform/orchestration/loop_expansion_handler.go`: the unconditional
stamp at ~line 157 became `resolveSubstepContinueOnError(substep.Config, continueOnError,
substepName, logger)`. A substep declaring `config.continue_on_error` as a bool gets its own
value; silence inherits the loop's; a non-bool falls back **and warns**.

Plus `platform/orchestration/substep_continue_on_error_test.go` (5 functions / 7 subtests,
mutation-proven), concept register **WFA-008**, a `LANDMINES.md` entry, and two
`WRONG_CALLS.md` rows.

**Commits (all mine, this session):**

| commit | what |
|---|---|
| `2e497e846` | the fix + tests + WFA-008 + landmine (`Council-Submitted:` trailer) |
| `a3e1ba182` | ticket update on `bugs_open/173` |
| `f9470fbf8` | council APPROVED r1, objections answered, `bugs_open/193` filed |
| `dc97d1940` | milestone summary + owner log |
| `f56c80944` | wrong-call: a compliance claim whose evidence I left out of the edits array |

**Live proof, measured 2026-08-04 on `v1.0.1250`, BOTH replicas** (`agent-chassis-88cf8787-4dzzx`,
`-5z5sn`):

```
NEW  resolveSubstepContinueOnError                        = 2   (definition + call site)
NEW  "Substep declares continue_on_error with a non-bool" = 1
CTRL "continue_on_error is true for this loop iteration"  = 1   (pre-existing; probe works)
NEG  resolveSubstepContinueOnErrorXYZZY                   = 0   (probe can return 0)
```

## Council

**APPROVED, round 1**, correlation `549e25fb-acc1-4806-a2a7-95bf73cca806`. 8 seats approving,
3 advisory objections, none high. All three answered with checks — full working in NOTES.
Nothing is outstanding from the council except the item below, which is for the owner.

**For the owner, unresolved:** `reuse_agent` and `architecture` both noted that owner ruling
2026-07-29 §1 (the RFC threshold) was **self-applied by the change's author**. Both agreed
with the reading; both flagged that the author is not the ruling's owner. It is in
`README_where_we_are.md` and the `SUMMARY`.

---

## PICK UP EXACTLY HERE — the live induction

### What is seeded in the LIVE database right now (and must be cleaned up)

`SEED_2026-08-04_induction_agents.sql` in this directory inserted **two temporary agent
definitions**. Each sets the loop-level flag to the **opposite** of the substep's, which is
the whole point — if expansion still clobbered the substep value, each run would produce the
other's outcome.

| agent type | loop `continue_on_error` | substep `boom` | expected terminal status |
|---|---|---|---|
| `test-173-tolerant-substep` | **(unset)** — strict loop | **`true`** | **COMPLETED** (both iterations skipped) |
| `test-173-strict-substep` | **`true`** — tolerant loop | **`false`** | **FAILED** |

Both loop over two rows and their first substep `boom` runs `SELECT 1/0` through
`query_database` — deterministic, read-only, cannot touch a production row.

**Verified seeded** (`is_active=t`, settings confirmed by query) before dispatch.

> ### ⚠ CLEANUP IS OWED WHATEVER HAPPENS NEXT
> ```sql
> DELETE FROM agent_definitions
> WHERE type IN ('test-173-tolerant-substep','test-173-strict-substep');
> ```
> Two stray `experimental` agents on a shared fleet is exactly the kind of residue that
> another session later finds and has to reverse-engineer. Do this even if you abandon the
> induction.

### Runs dispatched, outcome not yet observed

| orchestration_id | agent |
|---|---|
| `5f4aaaee-8aeb-46a1-8961-92b694ea1359` | `test-173-tolerant-substep` |
| `d4c7c99c-1247-407f-8ba6-074f2a187032` | `test-173-strict-substep` |

At session end **neither had an `orchestration_states` row**:

```sql
SELECT orchestration_id, orchestration_name, current_step, status, updated_at
FROM orchestration_states
WHERE orchestration_id IN ('5f4aaaee-8aeb-46a1-8961-92b694ea1359',
                           'd4c7c99c-1247-407f-8ba6-074f2a187032');
-- (0 rows)
```

**Do not read that as a dropped dispatch and retry on it** — CLAUDE.md is explicit that a
missing row is almost always latency (publish→start measured at **29 minutes** under normal
load), and the generic lane is the one `bugs_open/096` is about. Runs were dispatched at
roughly **10:45–10:50Z**; the pods rolled at 10:29:20Z/10:29:40Z, so the ~300s post-restart
drop window had passed.

**The one thing genuinely unresolved:** I sent the first two with `kcat`'s output redirected
to `/dev/null`, and **`kcat -P` can send nothing while exiting 0** (a recorded landmine). So
"published successfully" is *not* established for those two. I was re-dispatching the
tolerant variant with output visible when the session ended — that command is in
`RUNBOOK_substep_error_tolerance.md` and is reproduced below. **Never suppress kcat's output
on this estate.**

### Step 1 — decide latency vs drop

Re-check the table first (they may simply have landed):

```sql
SELECT orchestration_id, orchestration_name, current_step, status, updated_at
FROM orchestration_states WHERE orchestration_name LIKE 'ind173-%' ORDER BY updated_at;
```

If still empty after ~30 minutes from dispatch, re-dispatch **with output visible**:

```bash
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
echo "ORCH=$ORCH"
jq -nc '{action:"orchestrate",config:{agent_type:"test-173-tolerant-substep"},
         input_data:{note:"bugs_open/173 live induction"}}' \
| kubectl -n kafka run -i --rm "kcat-173r-$(date +%s | tail -c 5)" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H "correlation_id=$CORR" -H "request_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "message_id=$(cat /proc/sys/kernel/random/uuid)" \
    -H "orchestration_id=$ORCH" -H "orchestration_name=ind173-tolerant-retry" \
    -H "step_name=start" -H "client_id=demo_client" -H "message_type=request" \
    -H "action=orchestrate" -H "from_agent_type=user" -H "from_agent_id=cli" \
    -H "responses_topic=system.agent.generic.responses"
```

(Duplicate runs are harmless here — the agents are throwaway and the fault is read-only.)

### Step 2 — what each run must show, and what would REFUTE the fix

**This is the part to get right. A terminal status alone is not the proof** — the tolerant run
would also read COMPLETED if the loop never reached the failing substep at all. Check the
*mechanism*, not just the outcome:

```sql
SELECT status, current_step,
       collected_data ? 'run_loop_iter_0_error'            AS iter0_error_recorded,
       collected_data #>> '{run_loop_iter_0_error,skipped}' AS iter0_skipped,
       collected_data #>> '{run_loop_error_count}'          AS error_count
FROM orchestration_states WHERE orchestration_id = '<the tolerant run>';
```

| run | expected | what it proves | what would REFUTE |
|---|---|---|---|
| tolerant (`(unset)` loop, `true` substep) | `COMPLETED`, `run_loop_iter_0_error` present with `skipped=true`, `run_loop_error_count = 2` | the substep's `true` **overrode a strict loop** | `FAILED` ⇒ the substep declaration was ignored; or COMPLETED with **no** error record ⇒ the fault never fired, so the run proves nothing |
| strict (`true` loop, `false` substep) | `FAILED` | the substep's `false` **overrode a tolerant loop** — the direction that stops silent drops | `COMPLETED` ⇒ the loop's `true` still won and the fix is not working in the direction that matters |

**Both branches, or the flag is untested in the direction that matters** — 173's own words.
One passing branch is not half a proof; the tolerant branch alone is also consistent with
"the loop-level flag leaked through", which is precisely the bug.

Also worth a look, as corroboration rather than proof — the chassis logs should carry
`"Skipping failed loop iteration, advancing to next"` for the tolerant run, naming
`failed_step=run_loop_iter_0_boom`. Read **both** replicas (`logs deploy/X` reads one pod of N).

### Step 3 — close it

If **both** branches behave as tabled:

1. Append the evidence to `bugs_open/173` (status, the `iter_N_error` record, the error count,
   both orchestration ids, the image tag `v1.0.1250`).
2. Run the CLEANUP delete above and **confirm 0 rows remain**.
3. `git mv bugs_open/173_… bugs_closed/173_…` — and **name BOTH paths on the commit**, because
   a pathspec commit ships a `git mv` as a copy otherwise:
   `git commit bugs_open/173_….md bugs_closed/173_….md -m "close(173): …"`.
   Verify at HEAD, not at the tree:
   `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 173` → exactly one line.
4. Update the WFA-008 register entry's `status` from *built (inert until roll)* to *deployed*,
   and replace its `status-evidence` with the induction result.
5. Add the §10 row / update `MEMORY.md` if the lane warrants it.

If **either** branch misbehaves: **do not close it.** Record what happened in NOTES as a
correction, and treat it as a live defect in the fix — the code is already on the shared
branch, so the repair is forward-only, never an amend.

---

## Also left behind (deliberate, not forgotten)

- **`bugs_open/193`** — filed this session at the `bug_historian` seat's direction. The
  loop-level read (`loop_actions.go:66`) silently ignores a non-bool where its new
  substep-level twin warns, so the mechanism is now loud on one side and silent on the other.
  **Latent** — all 10 declaring loops declare `boolean` [MEASURED 2026-08-04]. Unowned; fix
  candidate 2 (a shared parse helper used by both call sites) is the one that closes the door.
- **A finding nobody has filed, because nobody is misled by it.** `fallback_step` and
  `retry_step` appear exactly twice in the whole Go tree, both in *name-prefixing* lists
  (`loop_expansion_handler.go:528`, `coordinator.go:4243`); there is no `models.Step` field and
  nothing reads either to route. **0 live definitions declare them.** They are prefixed as
  though they were live routing and consumed by nothing. Left unfiled because no author is
  currently misled — but do not "restore" them on the assumption they once worked.

## Two traps this lane paid for, so you do not

1. **Ownership cannot be measured by counting mentions of a bug number.** Every session lists
   `bugs_open/`, so all 54 numbers appear in most transcripts and every candidate scores the
   same. Count sessions that actually **opened the file** (`"file_path":".../bugs_open/NNN"`).
   Full method in `RUNBOOK` §R1; the wrong call is in `WRONG_CALLS.md`.
2. **`who-owns.py`'s verdict is dominated by the FILING lane.** It returned OWNED for 173,
   naming a lane that had merely written the bug file *for someone else* and has since closed.
   Read the named lane's handoff before believing the verdict, or you will decline every bug
   anyone ever wrote down.

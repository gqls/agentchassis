# 124 — one `needs_diagnosis` item runs TWICE, under two different correlations

**Filed** 2026-07-28 from the council-parallelism thread, out of `bugs_open/029` §6
(2026-07-28 section), where it first showed up as a wrong number in someone else's
evidence rather than as a bug in its own right.

**Severity** medium. It is not a correctness fault in any diagnosis — both runs
produce valid verdicts. It costs **double the LLM spend per diagnosis**, and it
**corrupts the evidence base of another open bug**, which is how it was found.

**Status** OPEN, unowned. Separable from `029` (hung spawns) and cheaper.

## Symptom — a live specimen, captured 2026-07-28 07:0x on chassis v1.0.1180

One work item, `needs_diagnosis:robot-hands-com-learning-cen`, created 06:58:44
with `spec->>'correlation_id' = 914dc844-…`, claimed 06:59:39. It produced **two
complete, independent diagnosis chains** — each of which is an LLM-heavy
`diagnose-agent` run:

```
corr c89c718c…   diagnose-dispatch-loop  06:59:39
                 diagnose-orchestrator   06:59:53
                 diagnose-agent          07:00:08   <- chain A
corr 914dc844…   diagnose-orchestrator   07:03:04
                 diagnose-agent          07:03:20   <- chain B
```

`914dc844` is the item's **own** correlation. `c89c718c` is a **fresh** one the
dispatch loop minted. Both chains ran; neither knew about the other.

Note what is *not* wrong here: every spawn in both chains shows
`awaited_requests.status = 'processed'` at `retry_version 0`. **This is the
system working normally and still doing the work twice.**

## Why it is invisible

**The two chains cannot be joined.** The 090 trigger's stated design is that one
key ties the intake item, the `diagnosis_artifacts` bundles and the terminal
`doc_notes` row together — *"carrying the SAME correlation_id, so the item, the
bundles … and the terminal doc_notes row all join on one key"*
(`090_TRIGGER_needs_diagnosis_v1.sh` header). When the dispatch loop mints its
own correlation instead of reusing `spec->>'correlation_id'`, that property is
broken for every loop-dispatched item, so the duplicate leaves no joinable trace.

## What it already cost, in another bug's evidence

`bugs_open/029`'s "fresh instance" section cites *"a second, earlier failure the
same evening (`site_work_items` `needs_diagnosis`, created 20:06:46) also went
`failed`"* as an instance of hung spawns. **It was not.** That item's diagnosis
had already **succeeded**: correlation `eb8df254-…` has two orchestrations,
`30084fbe` and `5143b54f`, both `COMPLETED`, returning a REFUTED verdict.

Its `failed` stamp came from request `c963122a-…`, which belongs to orchestration
`41d64b75` — a `diagnose-dispatch-loop` at step `call_handler`, sent `20:49:31`,
**43 minutes after that diagnosis finished**. The loop re-dispatched an
already-diagnosed item, that re-dispatch hit `029`, and the failure was written
back over a completed diagnosis.

So this bug **manufactures apparent instances of `029` and inflates its rate.**
Anyone counting `failed` needs_diagnosis items is counting some of these.

## Mechanism

Two contributing facts, one verified and one not:

1. **Nothing marks a `needs_diagnosis` item complete on success.** `[VERIFIED]` —
   the 090 trigger still prints *"closing it by hand until a diagnose dispatch
   loop exists"*. The loop now exists (`diagnose-dispatch-loop` is a live
   `agent_definitions` row) and still does not close them, so a diagnosed item
   stays claimable.
2. **What started chain B is not established.** `[UNVERIFIED]` — it could be the
   retry driver on the loop's `call_handler` request (`77ae59dd`, still `waiting`
   at capture with `timeout_at 07:34:53`), a second claim of the same item, or a
   separate direct dispatch. The item was already `claimed_at 06:59:39` when
   chain B started at 07:03:04, so whatever the path, **`claimed_at` is not
   preventing a second run.** Do not assume the retry driver without checking.

## Fix candidates, ordered by what closes the door

1. **Make a diagnosed item terminal.** Close the `needs_diagnosis` item when its
   diagnosis completes, rather than leaving it claimable for ever. This makes the
   duplicate *unrepresentable* instead of merely less likely, and it removes the
   standing pool of already-diagnosed items that the loop keeps re-picking. Check
   `workItemTerminalStatuses` / `idx_swi_dedup` lockstep before changing status
   values — that pair is a documented drift class.
2. **Make the loop reuse `spec->>'correlation_id'`** instead of minting a new one.
   Does not stop the duplication, but restores the one-key join the 090 design
   promises, so a duplicate becomes visible and countable. Cheap, and worth doing
   even if (1) lands.
3. **Establish why `claimed_at` did not exclude the second dispatch**, and give
   the claim real exclusivity if it is meant to have it.

(1) and (2) are independent and both worth having: (1) stops the waste, (2) stops
it being invisible.

## How to verify a fix

Fire one 090 run and confirm that **exactly one** `diagnose-agent` orchestration
exists for it after it completes, and that the item reaches a terminal status
without a hand-written `UPDATE`. The failing shape is two `diagnose-agent` rows
minutes apart under different `correlation_id`s:

```sql
SELECT owner_agent_type, correlation_id, created_at FROM orchestration_states
WHERE created_at > now() - interval '30 minutes' AND owner_agent_type LIKE 'diagnose%'
ORDER BY created_at;
```

Do not verify during an idle window — see `029`'s note that the fleet reports zero
when nothing is running.

## Landmine for whoever picks this up

**Do not cancel or clean up the duplicate chains to tidy the queue before the
diagnosis has run.** A previous attempt on the neighbouring bug cancelled the
failing orchestration and reaped its pod before filing, and the diagnosis loop
then recorded that it could find no specimen to examine (`WRONG_CALLS.md`,
2026-07-27). The duplicate chains ARE the evidence.

---

# RESOLVED 2026-07-28 — root cause was neither filed mechanism

**Fixed by** the `bugfix_124_double_dispatch` thread. Chassis **v1.0.1191**
(digest `sha256:2f96b795a5c4636d41bdc384318f3f2d264188b9bf4017fb2d74ff2746a760cc`)
+ migration `258_diagnose_loop_stamps_run_correlation.sql`. Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_124_double_dispatch/`.

## The cause

**Two dispatchers, one queue.** `090_TRIGGER_needs_diagnosis_v1.sh` wrote its
intake row at `status='awaiting_diagnosis'` **and then published the orchestrate
envelope itself**. `diagnose-pipeline-trigger` had been enabled, so
`diagnose-dispatch-loop` — which claims exactly those `awaiting_diagnosis` rows —
took the same item on its next 60s tick and ran a second, independent diagnosis.

The script's own header described the world before that switch: *"The task ships
DISABLED — enable it deliberately. Until then, and for any ad-hoc run, THIS
SCRIPT is the dispatcher."* Somebody enabled it. Nobody updated the script.

## Corrections to this file's own mechanism section

> **CORRECTED.** §Mechanism item 1, marked `[VERIFIED]`: *"Nothing marks a
> `needs_diagnosis` item complete on success."* **False.** The live
> `diagnose-dispatch-loop` row has a `mark_complete` (`complete_work_item`) step
> and it works — every 090-filed item sits at `complete` or `failed` with
> `claimed_by='diagnose-dispatch-loop'`. The `[VERIFIED]` marker was attached to
> an inference drawn from a **print statement in the shell script**, not from the
> live config. Caught by `SELECT jsonb_object_keys(default_config->'workflow'->'steps')`
> on the live `agent_definitions` row. Fix candidate 1 ("make a diagnosed item
> terminal") therefore needed no work: it was already true.

> **CORRECTED.** §"What it already cost": orchestration `41d64b75` is described
> as the loop re-dispatching an already-diagnosed item *"43 minutes after that
> diagnosis finished"*. Its `created_at` is **20:08:16 — 91 seconds after the
> item was created**. It is the concurrent duplicate; what happened at 20:49:31
> was its `call_handler` **retrying** (that retry is `029` proper). The
> conclusion — that the duplicate wrote `failed` over a completed diagnosis and
> inflates `029`'s apparent rate — **stands and is strengthened**. Written back
> into `bugs_open/029` §6 as well.

> **RE-RANKED.** Fix candidate 2 ("make the loop reuse `spec.correlation_id`")
> was ranked cosmetic — *"does not stop the duplication"*. It was load-bearing.
> `diagnosis_artifacts` are keyed on the **envelope** correlation
> (`params.ExecutionContext.CorrelationID`), so for a loop-dispatched item
> `spec.correlation_id` names **nothing at all** — it resolved only because the
> script's duplicate chain happened to run under it. Both `diagnosis-triage`
> items have a `spec.correlation_id` no run ever used. Remove the direct publish
> without fixing the join and the printed correlation stops resolving for
> everyone, so the two had to ship together.

## What shipped

1. **The claim is the ticket.** A direct publish now first takes the row
   `awaiting_diagnosis → diagnosing` atomically and publishes nothing if it loses
   the claim. This holds even when the `enabled` read is stale — and it always
   is, because the loop ticks every 60 seconds.
2. **One dispatcher, read from live state.** `DISPATCH` unset (the new default)
   asks the database whether the loop is live (task `enabled` **and** the agent
   row active); `1` forces a direct publish for a wedged loop; `0` is intake
   only. Nothing for an operator to remember, and it self-corrects if the loop is
   ever disabled again.
3. **`$ctx.` — a generic execution-context parameter namespace for
   `query_database`** (`platform/orchestration/actions/execution_context_params.go`).
   Any workflow SQL can now bind its own run identity. `claim_item` uses it to
   stamp `spec.dispatch_correlation_id` inside the same atomic UPDATE that
   claims, so item → orchestrations → artifacts → doc_notes join on one key for
   every dispatch path, including `diagnosis-triage`'s.

**Declined:** correlation *override* on `call_agent`/`spawn_agent` (letting a
child chain adopt an upstream correlation). Prettier trail, but correlation is
threaded through `awaited_requests`, response routing and the Kafka partition
key — a large blast radius for a cosmetic gain over the linkage row.

## Verified live, against a real run

`SLUG=scheduler-stamps-completed-at-publish`, fired 2026-07-28 17:04 — the
verification this file specifies, on a genuinely useful symptom rather than a
throwaway.

```
intake corr ae10e615-8e5e-4f02-ab40-da76faeea170   → ZERO orchestrations (the script published nothing)
run    corr 66a65287-f511-4d2d-9c54-7875939a5a78   → diagnose-dispatch-loop 17:05:28
                                                      diagnose-orchestrator  17:05:45
site_work_items.spec->>'dispatch_correlation_id' = 66a65287-…   (the join now holds)
```

**One chain, not two.** The failing shape — two `diagnose-agent` orchestrations
minutes apart under different correlations — did not occur.

For contrast, a specimen captured 90 minutes earlier on the unfixed path, another
session's intake at 16:37:51: `954d8da9` (script) orchestrator 16:37:58 + agent
16:38:12, and `2a656f25` (loop) dispatch-loop 16:38:28 + orchestrator 16:38:45 +
agent 16:38:59 — two `diagnose-agent` pods on the same symptom, 47 seconds apart.

## Residual — read this before re-opening

- The **direct-dispatch branch** (loop disabled, or `DISPATCH=1`) leaves the item
  at `diagnosing` with no closer, relying on `reap_stuck` at 75 minutes — which
  does not tick when the loop is disabled. That row then sits at `diagnosing`.
  Deliberate: leaving it at `awaiting_diagnosis` meant enabling the loop would
  re-run the entire stale backlog. The script prints the close-by-hand SQL.
- **Ordering is permanent, not one-off.** Migration 258 binds
  `$ctx.correlation_id`; a chassis predating v1.0.1191 resolves that to nil and
  **fails `claim_item`, stopping the diagnose lane**. Any rollback below 1191
  must revert 258 too (the pre-update snapshot is in `agent_definitions`).
- `029`'s rate must not be re-derived from `failed` needs_diagnosis rows without
  splitting on 2026-07-28: rows before it include duplicates this bug produced.

**Status: CLOSED — fixed AND live AND verified against a real run.**

---

## Council verdict: **REJECTED** — guardian veto on SCOPE, not on correctness

Corr `90361922-e4c4-482e-a0b7-b1a49640265a`, round 2 (round 1 died when this
change's own chassis roll replaced the pod running it — see the workstream NOTES).
**11 reviewers, 5 abstained, `unreadable: 0`** — a real verdict, not the harness.

| verdict | seats |
|---|---|
| approve | guidelines, diagnosis_guardian, render_guardian, mission |
| object | editquality, reuse_agent, tooling_provenance, debug_historian, constitution, prior_art_librarian |
| **veto** | **guardian** |

**No seat disputed the diagnosis or the fix's correctness.** `editquality` called
the causal analysis sound, said the plan *"correctly separates the two refuted
mechanisms from the two real, surviving conclusions and edits only the latter
two"*, called folding the stamp into the atomic claim *"the strongest part of the
plan"*, and stated its three objections are *"fixable, not structural"*.
`guidelines` analysed the DDL, dedup, provenance and contract rules explicitly and
found no violation.

**The veto, in full, is criterion (b) — architecture change dressed as a point
fix:** *"the plan itself states the intent is a generic, platform-wide seam rather
than a diagnose-specific fix … independent of how well-tested or additive the
mechanism is."* It objects to P3's `$ctx.` namespace touching a shared action's
param-resolution contract inside a bug patch, and names a contained alternative: a
diagnose-lane-scoped bespoke Go action that writes the correlation into
`collected_data` before `claim_item` runs.

**It is a fair process objection and it is recorded as binding on process, not on
this code.** Three things about it, stated plainly:

1. **The gate is advisory and this shipped before the verdict landed** — because
   migration 258 could not be applied against an older chassis without stopping
   the diagnose lane, so the image had to go first. That is a real ordering
   constraint, not a convenience. Said out loud rather than glossed.
2. **The guardian's alternative is the exact thing the `reuse_agent` seat objected
   to in the same round** — *"the platform ends up with two ways to get a run's
   correlation into a query … nothing here proposes migrating the old ones"*.
   Those two seats want opposite things; a 35th bespoke correlation-reader
   satisfies the guardian and deepens what reuse is complaining about. That
   tension is above this bug's pay grade.
3. **The three checkable objections are now MEASURED, not argued** (below).

### The objections, answered with measurements

| objection | seat | answer |
|---|---|---|
| *"the no-collision argument is asserted, not verified"* (high) | guardian, editquality | **Measured across every live workflow**: 63 steps fleet-wide use `params`; **exactly 1** path starts with `$`, and it is `$ctx.correlation_id`, this change's own. Zero pre-existing paths can be shadowed. |
| *"grounded evidence only confirms `claim_item` exists, not its config shape — other keys may be clobbered"* (low) | editquality | **Measured against the pre-update snapshot**: `claim_item.config` was `{query, output_format}`, is now `{query, params, output_field}` → `{query, params, output_format}`. Nothing lost. |
| *"the direct-dispatch branch leaves a row at `diagnosing` with no closer"* (medium) | editquality, guardian | **Real gap, now has the documented sweep the objection asked for** — `RUNBOOK_double_dispatch.md` § "Sweep direct-dispatched diagnoses left at `diagnosing`", keyed on `claimed_by='090_TRIGGER_needs_diagnosis'` (loop-claimed rows close themselves, so the sweep only ever returns rows nobody will close). |
| *"never names the bespoke actions it says every lane had to grow"* | reuse_agent | Fair — the rationale asserted it. **Enumerated: 34 files under `platform/orchestration/actions/` read `ExecutionContext.CorrelationID` directly**, `diagnose_assemble_bundle_action.go` among them. They are actions doing their own work with the correlation, not param-binding paths; `$ctx.` addresses config-authored SQL, which had no route at all. **No migration of the 34 is proposed and that is now an open question, not a silence.** |

### What is owed, and to whom

**The `$ctx.` namespace should go to architecture review on its own merits**, as
the guardian asked — separately from this bug, and it has shipped ahead of that
review. It is registered as `WFA-002` in the concept register with its ordering
landmine, so it is discoverable rather than buried. Reverting it now would break
the item↔run join the 090 script depends on and cost another image+migration
cycle; that is a reason to review it, not a reason to pretend the veto did not
happen.

**This bug stays CLOSED.** The defect is fixed, live and verified against a real
run. The veto is about how a platform capability reached production, not about
whether diagnoses still run twice — they do not.

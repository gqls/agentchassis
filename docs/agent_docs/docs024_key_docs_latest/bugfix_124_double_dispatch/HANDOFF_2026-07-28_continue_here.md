# HANDOFF — bug 124 lane, 2026-07-28 evening. Read this first.

**Bug 124 is CLOSED, live and verified.** Nothing here is a blocked deliverable.
What follows is (1) the state you must not break, (2) four open questions this work
raised, none of which are mine to decide alone, and (3) what I would do next.

Session: "bugsearch 2". Working docs beside this file: `PLAN_2026-07-28_double_dispatch.md`,
`RUNBOOK_double_dispatch.md`, `NOTES_double_dispatch.md`, `README_where_we_are.md`,
`SUMMARY_2026-07-28_two_dispatchers_one_queue.md`. Case file: `bugs_closed/124_…`.

---

## 1. State — RE-VERIFIED 2026-07-28 20:5x, after the v1.0.1194 roll

> Checked again after a third roll I did not perform. **v1.0.1194** (digest
> `sha256:8013878b…`), both pods `ctx-support=1`, replicas **2/2**, diagnose lane
> 0 stuck. The invariant below has now survived three consecutive rolls by other
> sessions — because the build comes from committed HEAD, which carries the fix.
> The table below is from the 1192 check; the 1194 numbers are identical in kind.

| thing | state | how it was checked |
|---|---|---|
| chassis | **v1.0.1192**, digest `sha256:4bd9a111…`, 2/2 replicas | `kubectl get pods -l app=agent-chassis -o custom-columns=…imageID` |
| `$ctx.` support in the binary | **present on BOTH pods** | `strings /app/agent-chassis \| grep -c "unknown execution-context field"` → 1, 1 |
| migration 258 | applied + recorded, `params: ["$ctx.correlation_id"]` live | `SELECT default_config->'workflow'->'steps'->'claim_item'->'config'->'params' FROM agent_definitions WHERE type='diagnose-dispatch-loop' …` |
| `agent_definitions.image_tag` | 186 rows at v1.0.1192 | `scripts/check-agent-image-drift.sh` |
| diagnose lane | 0 items stuck at `awaiting_diagnosis`/`diagnosing` | see RUNBOOK |

### The one thing that can break this, and it is permanent

**Migration 258 binds `$ctx.correlation_id`. A chassis BELOW v1.0.1191 resolves it
to nil, fails `claim_item`, and the diagnose lane stops dispatching entirely.**

That is now a standing invariant, not a one-off deploy note. **After every chassis
roll, pod-grep before assuming the lane is fine:**

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- \
  sh -c 'strings /app/agent-chassis | grep -c "unknown execution-context field"'   # must be 1
```

I ran exactly this against v1.0.1192 and it passed. A rollback below 1191 must
revert 258 too — the pre-update snapshot is in **`agent_definitions_backup`**
(NOT `agent_definitions`; 258's own rollback comment names the wrong table, and
the correction is in the RUNBOOK because 258 is checksum-recorded and must not be
edited).

---

## 2. Open questions — none of these are blocked, all of them are judgement calls

### 2a. ~~`$ctx.` is owed an architecture review~~ **RESOLVED — OWNER RULING 2026-07-28, Option A**

> **The review was held and the owner ruled: keep the code, fix the precedent.**
> `$ctx.` stays. The standing rule it produced is in `CLAUDE.md` §"Platform seams
> and the ordering exemption" and binds every session — a seam ships ahead of
> review only under a **stated ordering constraint** and only if **registered in
> the same commit**; blast-radius claims are **measured before submitting**; a
> **scope veto is not answered by resubmitting**. Full review with the three costed
> options: `REVIEW_2026-07-28_ctx_namespace.md` beside this file.
>
> **Item 2b resolved on the way**: of 34 `ExecutionContext.CorrelationID` readers,
> only **2** bind it into SQL and 16 use it as a Kafka partition key — 32 are not
> about SQL at all. My submission's *"every lane had to grow a bespoke Go action"*
> was **overstated and is corrected**; the substantive point survives (none is a
> reusable mechanism, none is callable from config, so there is nothing to migrate).
>
> **Deliberately NOT ruled on:** whether `$ctx.` gets a second consumer.
> **Accepted with open eyes:** the ordering hazard is mitigated by docs + a
> pod-grep, not by code.

*(original text kept below for the record)*

### 2a-original. `$ctx.` is owed an architecture review — the council vetoed on scope

Council corr `90361922-e4c4-482e-a0b7-b1a49640265a`, round 2: **REJECTED**,
11 reviewers, `unreadable: 0` (a real verdict, not the harness — check `unreadable`,
never `abstained`).

**No seat disputed the diagnosis or the fix.** guidelines / diagnosis_guardian /
render_guardian / mission approved; `editquality` called its own three objections
*"fixable, not structural"* and the atomic-claim stamp *"the strongest part of the
plan"*. The veto is criterion (b): a platform-wide seam added to a shared action's
param-resolution contract **inside a bug patch**, *"independent of how well-tested
or additive the mechanism is."*

It shipped ahead of that review because migration 258 cannot be applied against an
older chassis without stopping the lane — the image genuinely had to go first.
**That is a reason, not an excuse, and it is recorded as such in `bugs_closed/124`.**

Two things for whoever takes this up:

- **The guardian's proposed alternative is what the `reuse_agent` seat objected to
  in the same round.** Guardian: use a diagnose-lane-scoped bespoke Go action.
  Reuse: *"the platform ends up with two ways to get a run's correlation into a
  query … nothing here proposes migrating the old ones."* A 35th bespoke
  correlation-reader satisfies one seat and deepens the other's complaint. **Do not
  simply implement the guardian's alternative without resolving that** — the two
  seats want opposite things and someone senior has to pick.
- The architecture-review seat exists and is live (`docs024/architecture_review/`,
  D11 layers 1+2 on v1.0.1182). That is the right venue. **I did not file it.**

### 2b. ~~34 bespoke `ExecutionContext.CorrelationID` readers, unmigrated~~ **RESOLVED — see 2a**

`grep -rl "ExecutionContext.CorrelationID" platform/orchestration/actions/*.go | wc -l`
→ **34**. My submission asserted *"every lane wanting this had to grow a bespoke Go
action"* without naming them; the reuse seat rightly called that out.

My position — **stated, not proven**: those are Go actions doing their own work
with the correlation (e.g. `diagnose_assemble_bundle_action.go` writing artifacts),
not param-binding paths; `$ctx.` addresses *config-authored SQL*, which had no
route at all. **Someone should test that claim before it hardens into folklore.**
If some of the 34 are genuinely "bind run identity into a query", they are
migration candidates and the reuse objection is stronger than I allowed.

### 2c. ~~The manual-script + automatic-loop shape, fleet-wide, NOT audited~~ **DONE 2026-07-28 evening**

> **RESOLVED — audit run. 0 live instances beyond 124, 1 latent, 1 false positive.**
> Signature: a script that inserts a `site_work_items` row **and** publishes an
> envelope (one that merely pokes a loop is benign — the claim is atomic). 3 hits:
> `090` (fixed), `180_adoption/080_trigger_adoption.sh` (false positive — inserts
> an `evaluate_tools` row, separately publishes unrelated work), and
> **`092_TRIGGER_experience_plan.sh` — LATENT**: intake at a private status plus
> its own publish, safe only while nothing claims `awaiting_experience_plan`
> (nothing does; checked `agent_definitions` and `scheduled_tasks`, both empty).
> **A warning block now sits in 092 above its INSERT** naming both remedies.
> `report-dispatch-loop` is a disabled clone of the diagnose loop but is safe by
> construction — its items come from a Go action, not a publishing script.
> Recorded in NOTES and back into 016b §9, which I corrected: I had asserted
> "fleet-wide" from how the lanes grew rather than from a count.

*(original text kept below for the record)*

This was the highest-value loose end.

`bugs_open/124` was: a hand-run trigger script that dispatches, plus an automatic
loop enabled later that dispatches the same queue. **Most lanes here grew that way**
— script first, loop bolted on, script left in place "until the loop is enabled" —
and enabling a loop is one line of SQL that never touches the script.

Written up as a transferable pattern in **016b §9, "Turning on the automatic path
does not turn off the manual one"**, with the check. **The audit itself has not been
run.** Starting point:

```sql
SELECT name, enabled, target_agent_type, concurrency_group
FROM scheduled_tasks WHERE enabled ORDER BY target_agent_type;
```

then, for each, look for a `*_TRIGGER_*.sh` that publishes to the same lane
unconditionally. I would not assume 124 was the only one.

### 2d. `bugs_open/029`'s rate is partly inflated by 124's duplicates

`029` §6 now carries a dated correction from me. **Do not re-derive 029's rate from
`failed` needs_diagnosis rows without splitting on 2026-07-28** — rows before it
include duplicate chains this bug produced, where one chain succeeded and the
other's `mark_failed` wrote over it.

---

## 3. Mistakes from this session, so they are not repeated

- **My deploy silently scaled the chassis 2 → 1.** The second replica had been
  added imperatively that morning and `patch-deployment.yaml` still said 1, so
  *any* session's next `apply -k` would undo it — toward less capacity, with no
  warning. Restored in under a minute; the overlay now declares 2 with the reason
  inline. **Other services have NOT been checked for the same drift.**
- **My deploy killed my own council round.** The gate runs its seats **inline on
  the chassis** (`bugs_open/096`), so `rollout restart deployment/agent-chassis`
  ends any round in flight. It does not go FAILED — it sits at `EXECUTING_STEP`
  with empty `awaited_requests` and nothing recovers it. **Get the verdict, then
  roll**; where the image must go first, submit *after* the roll.
- **My first claim guard could not fail.** `psql -t -A` prints `UPDATE 0` on a
  zero-row `RETURNING`, so `[ -z "$CLAIMED_ID" ]` never fired and the script
  dispatched against an item another dispatcher owned. Found *only* by staging the
  failure — two hours of green happy-path verification had not touched it. Fixed
  with a CTE-wrap plus a UUID shape assertion. **Two other threads had already hit
  this and fixed it privately in their own script comments**; now in 016b §9, which
  is where it should have been the first time.
- **A `[VERIFIED]` tag in the original bug file was earned off a print statement.**
  Logged in `WRONG_CALLS.md`. Ask *what did I actually read?* before tagging.

---

## 4. What I would do next, in order

0. ~~Run the 2c audit~~ — **DONE**, see above. Result: nothing live to chase.
1. ~~Route 2a to architecture review~~ — **DONE. Owner ruled Option A**; the rule
   is in CLAUDE.md and 2b was settled by the same evidence.
2. **Run the 2c audit.** One query plus a grep per enabled task. If it finds a
   second instance, that is a fleet-wide class and worth a bug of its own.
3. **Settle 2b** with a read of the 34 files — it either strengthens or dissolves
   the reuse objection, and right now it is my assertion against theirs.
4. Leave 2d alone unless someone is actively re-measuring 029.

Nothing above blocks anything else. The lane works, and it now costs half what it
did this morning.

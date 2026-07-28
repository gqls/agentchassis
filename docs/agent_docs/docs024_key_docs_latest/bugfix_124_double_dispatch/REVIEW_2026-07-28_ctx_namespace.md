# Architecture review — the `$ctx.` parameter namespace

**Reviewed** 2026-07-28. **Outcome: OWNER RULING — Option A. Keep the code; fix the
precedent.** The rule that came out of it is now in `CLAUDE.md` §"Platform seams and
the ordering exemption".

**Subject:** `platform/orchestration/actions/execution_context_params.go` plus the
resolution hook in `QueryDatabaseAction` (`database_actions.go`). Live on chassis
≥ v1.0.1191. Concept register **WFA-002**. One consumer: `diagnose-dispatch-loop`'s
`claim_item`, via migration 258.

**Council verdict being reviewed:** REJECTED, corr
`90361922-e4c4-482e-a0b7-b1a49640265a` round 2 — guardian veto, criterion (b),
architecture change inside a point fix. 11 reviewers, `unreadable: 0`.

> **Reviewer's declared bias:** this review was written by the thread that wrote
> the change. Where judgement replaced measurement it is marked. The three
> measurements below were taken *because* the council said they were asserted.

---

## 1. The question

Not whether the mechanism is correct — no seat disputed that. It is:

> **Does a platform-wide capability get to stay when it reached production inside
> a bug patch, ahead of the review it needed?**

## 2. The gap it fills

Workflow steps are database config. A `query_database` step binds SQL parameters
from paths into `collected_data`. **A run's own identity is not in
`collected_data`** — it lives in `ExecutionContext`, reachable only from Go.

So config-authored SQL could not answer *"which run am I?"*. That is exactly what a
dispatch loop must record when it claims a queued row: without it the row says
*that* it was claimed but not *by which run*, and there is no path from the work
item to the artefacts the run produced. In the diagnose lane that made the
documented "one key joins the intake, the bundles and the verdict note" property
false for every loop-dispatched item — **silently**, because `spec.correlation_id`
still looked like a key.

## 3. The objection, at full strength

> *"Adding a new reserved namespace to the param-resolution contract of a shared
> action is an architecture-change signal by definition, regardless of how
> carefully it's tested… `QueryDatabaseAction` is consumed by every pipeline that
> has a `query_database` step."*

And the sharpest part: the plan **asked the reviewer to verify** there was no
collision rather than having verified it. That is a fair hit, and it is the part
that became a standing rule.

The process point stands independently of the technical outcome: a platform seam
should be reviewed on its own timeline, not carried through as a rider. It shipped
first because migration 258 cannot be applied against an older chassis without
stopping the diagnose lane — a genuine ordering constraint, but that **explains the
sequence, it does not license it.**

## 4. Evidence taken after the veto

| question the council raised | measured answer |
|---|---|
| Can `$ctx.` shadow an existing param path? | **No.** 63 `params` entries across every live workflow; exactly **1** begins with `$`, and it is this change's own. |
| Did replacing `claim_item.config` clobber keys? | **No.** Pre-update snapshot `{query, output_format}` → live `{query, params, output_format}`. |
| Direct-dispatch rows left at `diagnosing` with no closer | Real gap. Documented sweep now in `RUNBOOK_double_dispatch.md`, keyed on `claimed_by`. |

### 4a. The 34 correlation-readers — this partly refutes the submission

The `reuse_agent` seat pressed the submission's claim that *"every lane wanting
this had to grow a bespoke Go action"*, which was asserted and not enumerated. Now
classified. Of 34 files reading `ExecutionContext.CorrelationID`:

- **2** bind it into SQL — `diagnose_assemble_bundle_action.go` (keys
  `diagnosis_artifacts`) and `spawn_actions.go` (writes `awaited_requests`);
- **16** use it as a **Kafka partition key**;
- ~28 occurrences put it in an outgoing message field or a log field;
- 2 truncate it for a log prefix.

**The rhetoric overstated it: 32 of the 34 are not about SQL at all.** Recorded as
a correction rather than quietly dropped.

The substantive point survives in better shape. **None of the 34 is a reusable
mechanism and none is callable from workflow config.** The two SQL ones are
hard-coded, single-table, and do considerably more than bind a value. So `$ctx.`
does not duplicate them — it serves a different consumer (config authors, not Go
authors) and removes the reason to write a third. **No migration is proposed
because there is nothing to migrate**, which is the answer the reuse seat asked
for and did not get in the submission.

## 5. The seat conflict

The guardian's contained alternative — *"a diagnose-lane-scoped bespoke Go action
that writes the correlation into `collected_data` before `claim_item` runs"* — is
precisely the shape `reuse_agent` objected to in the same round: *"the platform
ends up with two ways to get a run's correlation into a query… nothing here
proposes migrating the old ones."*

Implementing it would make a **35th** bespoke reader: satisfying the containment
seat by deepening what the reuse seat is complaining about. It is also more code,
and it reintroduces a failure mode the current design eliminates — a separate stamp
step can claim and then die before stamping, leaving the unjoinable row this fixes.
`editquality` called folding the stamp into the atomic claim *"the strongest part
of the plan"*.

**Two seats wanted opposite things. No resubmission resolves that.** This is the
class of verdict that needs a human, and it got one.

## 6. Residual risk, stated plainly

**The ordering hazard is permanent and mitigated by documentation, not by code.**
A config binding `$ctx.` against a chassis predating the action fails that step; on
a *claim* step that stops the lane, silently, with no failed row to find. Every
chassis roll now carries a pod-grep for this lane
(`strings /app/agent-chassis | grep -c "unknown execution-context field"`).
Verified by hand on v1.0.1192. **The next person has to remember** — that is a real
weakness of Option A and it was accepted with open eyes.

**The precedent was the larger risk**, and it is what the ruling addresses.

## 7. Options as put to the owner, and the decision

**A — Keep it; fix the precedent rather than the code.** Technically free; measured
risk near zero. Accept the process breach explicitly and write the rule for next
time. *Recommended by the reviewer, who is also the author.*

**B — Revert; implement the guardian's alternative.** Another image + migration
cycle; breaks the item↔run join until it lands; adds a 35th bespoke reader against
the reuse objection; reintroduces the claim-then-die window. Buys the containment
principle enforced visibly.

**C — Keep but freeze.** No second consumer until a full architecture review runs.
Cheap and honest; leaves the capability discoverable but unspread.

> ### OWNER RULING 2026-07-28 — **Option A.**
> The code stays. The precedent is fixed by rule, now in `CLAUDE.md` §"Platform
> seams and the ordering exemption": a platform seam may ship ahead of review only
> under a **real, stated ordering constraint** AND only if it is **registered in
> the same commit**; blast-radius claims are **measured before submission, not
> handed to the reviewer**; and a **scope veto is not answered by resubmitting** —
> it is recorded, routed, and broken by a human.

**Explicitly NOT ruled on**, and left open for whoever needs it: whether `$ctx.`
should acquire a second consumer. Nothing depends on it today.

## 8. What the council got right, for the record

The veto was correct on its own terms and this review does not overturn it — the
owner overrode the *consequence*, not the *finding*. Both things the council forced
were improvements: the collision claim became a measurement, and the "every lane
grew a bespoke action" rhetoric was corrected against the code. **A REJECTED verdict
that produces two corrections and a standing rule is the gate working**, not the
gate obstructing.

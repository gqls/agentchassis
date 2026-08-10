# RFC 023 — turning a silent success into a loud failure is a DELIVERY-GUARANTEE change, even when it is unambiguously a bug fix

**Filed 2026-08-10 by the `bugfix_239` dispatch lane. For a human to break, not for a
thread to argue.** Filed *after* the change shipped (commit `a097e3e26`), which the
2026-07-29 owner ruling says is by design: HEAD is shared, `make build-*` builds from
committed HEAD, and any other session's roll ships my commit — so I could not have held it
and will not pretend I could. Condition (2) of the ordering exemption **is** met: the seam
is registered in the same commit (concept register **SYS-090**), with its landmine and this
open question written down.

**Council gate: `fca1071b-80ac-40cd-8c6d-d30a735de89b` — REJECTED, hard veto from
`guardian`, on SCOPE. The full verdict and the checked objections are at the foot of this
file; read that before the argument above, which was written before the round returned.**
This RFC is not an appeal — it existed before the verdict, was filed because the pre-commit
hook was right, and the architecture seat's own objection says *"Route it through
architecture_review"*, which is here. The gate reviews the fix I wrote; this asks a question
the gate is the wrong instrument for, and the verdict made that question sharper rather
than answering it.

## Why this is here at all — the pre-commit hook was right

The `.githooks/pre-commit` architecture signal fired on three clauses of the trigger test
at once, and I agree with all three:

- **a shared contract / delivery guarantee.** An `action=orchestrate` message that the
  chassis cannot resolve used to be **COMPLETED**; it is now **FAILED**. For the
  `chassis_intake_events` layer the same message used to be marked `done`; it is now
  `failed` (terminal) or left for re-attempt (transient).
- **added exported symbols other packages depend on** — `errors.ErrDispatchUnresolvable`,
  `errors.ErrDispatchLookupUnavailable` and their predicates, `discovery.IsNotFound` /
  `ErrAgentDefinitionNotFound`, `StateRepository.ReleaseMessageClaim`. ("Adds" counts since
  2026-08-09, RFC_019.)
- **five `platform/*` top-level packages** in one commit (`errors`, `discovery`,
  `messaging`, `agentbase`, `orchestration`) — the rule of thumb is three.

## What changed, in one paragraph

`bugs_open/239`: a dispatch naming `config.agent_type=X` sometimes ran as
`owner_agent_type='generic'` using the generic agent's single no-op step and reported
`COMPLETED` with an empty `execution_path`. Root cause (found from `chassis_intake_events`
payload bytes): `kcat -P` publishes one message per LINE of stdin, so a multi-line envelope
arrived as invalid-JSON fragments — but the durable defect was that `selectWorkflow` had
four fall-throughs to the CONSUMING pod's own workflow, three of them logging nothing.
Those now refuse, with two typed codes carrying opposite dispositions
(`DISPATCH_UNRESOLVABLE` terminal, `DISPATCH_LOOKUP_UNAVAILABLE` transient/retryable), a
FAILED `orchestration_states` row owned by the **requested** type, and the bugs_open/196
failure envelope.

## The question I actually want broken

**Where is the line between "a bug fix that makes a silent failure visible" and "a
delivery-guarantee change that needs architecture review first"?**

Because on the 2026-07-29 narrowing — *"an addition to a shared vocabulary needs an RFC only
when it changes what the shared mechanism GUARANTEES"* — this change is on the RFC side, and
**every fix of this shape will be**. The whole class "X silently succeeded when it should
have failed" changes a guarantee by definition; that is what fixing it means. If that class
routes to architecture review as a matter of course, the seat's signal fills with fixes
nobody disputes. If it never does, then the one property consumers actually depend on — "a
message I send is not going to start failing" — is the one property no review looks at.

I do not think the answer is obvious, and I do not think a thread should decide it. Two
framings, both defensible:

- **(A) It is a point fix.** Nothing that was *working* stops working: the census shows the
  only traffic that begins to fail is malformed. A guarantee that was never true ("your
  dispatch ran") cannot be broken by making it honest.
- **(B) It is a contract change.** Consumers build on observed behaviour, not on stated
  behaviour. Some pipeline somewhere may treat COMPLETED-with-no-steps as "nothing to do,
  carry on", and a FAILED row plus an `error_unrecoverable` response now routes it down an
  error path it has never taken. The census cannot see that: it measures *senders*, not what
  senders do with the answer.

**What I did in the absence of a ruling** — stated plainly so the seat can judge the method,
not just the outcome: I took (A), and bounded (B) by measurement before submitting rather
than leaving it in the risks block. 8 days, 10,851 request messages in
`chassis_intake_events`: 48 unparseable bodies (all this bug's own kcat fragments), 711
no-agent-type orchestrates (all carrying an inline workflow — untouched path), 9,433
scheduler messages all naming `agent_type` explicitly including the `generic` ticks whose
real workflow IS the no-op (they resolve, so they are unaffected), 165 council messages. I
also left the one branch I could not measure — an orchestration action naming nothing, which
is normal on dedicated/spawned pods whose job topics that table cannot see — on its old
behaviour, with a new `DISPATCH_OWN_DEFAULT` warn line so the population becomes countable
before anyone narrows it further.

## The four bars, answered honestly

This is not a "replace battle-tested code" proposal, so bars 1–4 sit oddly; answered anyway
because the process says the burden is on the change.

1. **A defect the current design cannot express a fix for.** *Partly.* The four
   fall-throughs could each have been patched in place. What could NOT be expressed
   contained is the distinction the whole fix turns on: `FindBestGroup` returned an untyped
   error for both "no such agent" (terminal) and "the database faulted" (transient), so no
   caller could route them differently without the typed sentinel this adds. A fifth site
   with the same defect **can** still be created after this fix — `selectWorkflowOLD` is
   already one, dead, recorded in `bugs_open/247`.
2. **Blast radius derived mechanically.** `FindBestGroup`/`FindByType` have exactly **one**
   production caller (`processor.go`, verified by grep over `--include=*.go`); the intake
   disposition has one (`intake_workers.go`); `processMessage`'s new return value is
   discarded by both Kafka consume loops. The message-population census above is the
   data-side blast radius. What is NOT mechanically derived, and I say so rather than
   dressing it up: what downstream consumers *do* with a FAILED row they used to see as
   COMPLETED.
3. **Independently-valuable stages.** Fails, and deliberately. This shipped as one commit
   because its halves are not independently safe: refusing without the typed transient arm
   would drop messages on a database blip, and the claim-release without the refusal is
   dead code. A reviewer who thinks that is wrong is exactly who this RFC is for.
4. **A rollback needing no migration.** Passes. No schema change. The previous binary reads
   every row this one writes (`orchestration_states.status='FAILED'`,
   `chassis_intake_events.status='failed'` are both pre-existing values), so a roll-back is
   a roll-back and nothing is stranded.

## What I am asking for

1. **A ruling on the line** — the question above, in general form, so the next lane fixing a
   silent-success bug knows whether to file here or just fix it. My own guess, offered to be
   argued with: **the trigger should be whether a consumer's SUCCESS path changes, not
   whether a guarantee changes** — making a failure visible is a fix; making something that
   succeeded start failing is a contract change. Under that test this change is a point fix
   and needs no RFC, and RFC_002's worked case (an evaluator gaining the ability to *refute*)
   still needs one, because that changed what a passing page meant.
2. **An objection on record if framing (B) is right**, in which case the remedy is not to
   revert — forward-only, and the silent no-op is worse than either alternative — but to
   name the consumer class that needs telling, and I will tell them.
3. Nothing blocked. The code is live-on-commit and rides the next fleet roll;
   `bugs_open/239` stays OPEN until the post-roll verification passes.

## Sources

- `bugs_open/239` (root cause, census, fix, live verification plan) · commit `a097e3e26`
- Concept register **SYS-090** (the seam, its consumers, its landmine, its open question);
  **SYS-014** (updated — this implements the fix shape it proposed in 2026-04)
- `LANDMINES.md` 2026-08-10 (kcat one-message-per-line) · `WRONG_CALLS.md` 2026-08-10
- Council round `fca1071b-80ac-40cd-8c6d-d30a735de89b`
- Prior rulings this sits under: 2026-07-28 (platform seams are architecture-scope),
  2026-07-29 §1 (narrowed to guarantee-changing) and §2 (the ordering exemption is retired;
  review is after the fact), 2026-08-02 RFC_010 (opt-in field over documented contract —
  considered and not applicable here: there is no "callers must all be X" licence in this
  change, and a default-OFF switch on a *refusal* would leave the silent no-op running,
  which is the bug)

---

# VERDICT: REJECTED — hard veto from `guardian`, on SCOPE (round 1, corr `fca1071b`, 2026-08-10)

11 reviewers, 6 abstained. **Seven approved, one vetoed, three objected.** The veto is on
*how* the change reached production, not on whether it is right — and the seats disagreed
with each other, which CLAUDE.md names as the condition for putting it in front of a human
rather than resubmitting. That is what this RFC now is.

| seat | verdict | the short of it |
|---|---|---|
| **guardian** | **VETO** | four packages, exported signature changes, a shared classifier list — "the MANY packages at once architecture-change signal, not a point fix" |
| architecture | object, `needs_rfc` | "the design direction (fail closed, one parse, typed disposition) is correct… Route it through architecture_review to formalize the rollback path" |
| prior_art_librarian | object (med) | two landmines key on the exact symbols edited and were not cited or reconciled |
| editquality | object (med) | edit 6 is scope creep against minimality; `sendErrorResponse`'s delivery outcome unchecked. Explicitly **"Not a veto: the core fix is on-target"** |
| reuse_agent | approve | "unusually well-grounded on reuse… **No architecture-level reuse concern identified**" |
| constitution | approve | "Root cause is fixed at the mechanism rather than patched at the trigger… the one deliberately left-open path is stated and justified" |
| mission | approve | "a direct instance of the mission's core failure mode — silent override — being eliminated" |
| debug_historian | approve | blast-radius scoping "sound"; mutation verification present; one uncited landmine |
| tooling_provenance, diagnosis_guardian, render_guardian | approve | out of jurisdiction / registry-resolved, no filename-convention resolution |

**Note the direct contradiction the owner should see:** the guardian vetoes on architecture
scope; `reuse_agent` — the seat whose whole remit is whether a change fits the existing
architecture — reports **no architecture-level reuse concern**. This is the same
seats-disagree shape as `bugs_closed/124`.

## The substantive objections, checked rather than argued

Three were checkable. I checked them; two dissolve, one is real and now stated.

**1. guardian edit-6 (HIGH): "FindByType's new guards change which row EVERY consumer
resolves to, fleet-wide."** Measurable, and measured 2026-08-10 — it changes **nothing
today**:

```sql
-- the population the new guards exclude that the old predicate admitted
SELECT count(*) FILTER (WHERE is_active AND COALESCE(is_snapshot,false)) AS active_snapshots,
       count(*) FILTER (WHERE is_active AND deleted_at IS NOT NULL) AS active_but_deleted,
       count(*) FILTER (WHERE is_active) AS loader_visible, count(*) AS total
FROM agent_definitions;                          -- → 0 | 0 | 186 | 203

-- does the SELECTED row differ, old predicate vs new, for ANY type?
WITH old_pick AS (SELECT DISTINCT ON (type) type, id FROM agent_definitions
                  WHERE is_active ORDER BY type, version DESC),
     new_pick AS (SELECT DISTINCT ON (type) type, id FROM agent_definitions
                  WHERE is_active AND deleted_at IS NULL
                    AND (is_snapshot IS NULL OR is_snapshot = false)
                  ORDER BY type, version DESC)
SELECT count(*) FILTER (WHERE o.id IS DISTINCT FROM n.id) AS row_changes,
       count(*) FILTER (WHERE n.id IS NULL)               AS resolves_to_nothing,
       count(*)                                            AS types
FROM old_pick o FULL JOIN new_pick n USING (type);        -- → 0 | 0 | 182
```

**0 of 182 types resolve to a different row; 0 resolve to nothing.** The check is
disconfirmable — a single active snapshot (stored at version+1000) would make the first
count non-zero, and the landmine `prior_art` cited says that population is held empty only
by `snapshot_agent()` writing `is_active=false`, "an invariant nothing states or tests".
So the guards are **inert today and prospective**: they close a door that is currently
unlocked and unused. That does NOT make the guardian's process point wrong — an inert
change is still a change shipped in the wrong commit — but the *blast radius* it objects
to is zero, and I would rather the owner rule on the process point with the number in
front of them.

**2. editquality edit-5 (MED): "sendErrorResponse's outcome via `kafka.DeliverReply` is
unchecked — the refusal may not be loud."** Does not apply on this path, and the design is
better than the objection assumes. `sendWorkflowResponseWithStatus` (`processor.go:866`)
calls `p.producer.Produce` **directly**; it does not route through `DeliverReply` (that is
agentbase's error-forward path). My caller *does* check the returned error and logs it at
Error. Most importantly the ordering is deliberate: **`recordDispatchFailureState` runs
BEFORE `sendErrorResponse`**, so the durable FAILED row does not depend on the wire. If
the response is never delivered, the refusal is still queryable. Worth stating in the code,
and now stated here.

**3. prior_art / debug_historian / constitution: the `extractGroupInfo` landmine
(LANDMINES:5788) is not cited or reconciled with the "Priority 1 untouched" claim.**
**This one is real and I should have cited it — I read it during the prior-art sweep and
did not put it in the submission.** Reconciled now: that landmine is the NESTED-envelope
case — a `call_agent` child whose `config.agent_type` sits under `body`, invisible to
`extractGroupInfo`, which therefore selects nothing and runs the consumer's own no-op.
**My change does not fix it.** Such a message still reaches the own-default branch. What
changes is that it is no longer silent: it now logs `DISPATCH_OWN_DEFAULT` with the pod's
agent type. The census found **7 messages of exactly that shape** in 8 days (6 with keys
`config,headers,input_data`, 1 with `body,headers`). So the honest statement is: *this
change makes that landmine's failure mode detectable and leaves it unfixed* — which is
also why the own-default branch was deliberately not made to refuse.

## The guardian's own contained alternative, recorded verbatim

> "Split into at least two independently-reviewable changes — (1) the platform/errors typed
> codes + platform/discovery FindByType guard/sentinel fix, reviewed on its own for
> fleet-wide blast radius against every FindByType consumer; (2) the
> processor.go/agent.go/intake_workers.go fail-closed refusal behaviour, reviewed once (1)
> is landed and stable… Ship the DISPATCH_OWN_DEFAULT logging first, alone, to get real
> fleet data on the unnamed-agent-type population before deciding whether to also refuse
> it."

**What I think of it, offered as evidence not as an appeal.** The last sentence is good
advice I partly took by accident — `DISPATCH_OWN_DEFAULT` ships as logging only and refuses
nothing, precisely so that population can be measured before anyone narrows it. The split
into (1) and (2) is coherent and I would have taken it had I seen it first. What I do not
think survives contact with this tree is the *sequencing*: "reviewed once (1) is landed and
stable" assumes a thread can hold (2) back, and the 2026-07-29 ruling says it cannot —
HEAD is shared and any session's roll ships whatever is committed. A two-commit split here
buys reviewability, which is real, but not staged exposure, which is what the word "stable"
implies.

## What I am NOT doing, and why

**Not resubmitting.** CLAUDE.md: *"A veto on SCOPE is not answered by resubmitting with
better measurements. It is a judgement about how a capability reached production."* The
measurements above are recorded so the owner can rule, not to re-run the round.

**Not reverting.** Forward-only, and the alternative is worse: the silent no-op is live
right now on every unrolled pod, and the fix is already at HEAD (`a097e3e26`) where any
session's build will ship it. Reverting would restore a defect that has already caused one
production incident.

**Not splitting retroactively.** The commit is at HEAD and forward-only forbids the rewrite
that would separate it. If the owner rules the split should have happened, the precedent is
what gets fixed — the code stays, per the 2026-07-28 ruling on `bugs_closed/124`.

## The question, restated now that the round has run

The original question stands and the verdict sharpens it. Both the guardian (veto) and the
architecture seat (`needs_rfc`) reached "architecture-scope" by counting **packages and
exported symbols**, not by asking whether any consumer's behaviour actually changes. On
that count this change is large. On the count I proposed above — *does a consumer's SUCCESS
path change?* — it is small: 0 of 182 agent-definition resolutions move, and the only
traffic that begins to fail is malformed.

**So: is the trigger a property of the DIFF, or of the BEHAVIOUR?** A fix that touches four
packages to make one seam honest will always fail the first test and always pass the
second. I do not think a thread should pick. Two concrete things a ruling could give the
next lane:

1. whether "four packages" counts when the packages are a leaf error type, its classifier,
   the one caller, and that caller's one caller — i.e. a vertical slice, not four subsystems;
2. whether an *inert* fleet-wide change (measurably 0 rows affected) may ride with the fix
   that motivated it, or must always be its own round.

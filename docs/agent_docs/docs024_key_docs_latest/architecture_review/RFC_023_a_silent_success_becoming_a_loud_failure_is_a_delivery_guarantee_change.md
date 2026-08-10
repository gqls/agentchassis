# RFC 023 — turning a silent success into a loud failure is a DELIVERY-GUARANTEE change, even when it is unambiguously a bug fix

**Filed 2026-08-10 by the `bugfix_239` dispatch lane. For a human to break, not for a
thread to argue.** Filed *after* the change shipped (commit `a097e3e26`), which the
2026-07-29 owner ruling says is by design: HEAD is shared, `make build-*` builds from
committed HEAD, and any other session's roll ships my commit — so I could not have held it
and will not pretend I could. Condition (2) of the ordering exemption **is** met: the seam
is registered in the same commit (concept register **SYS-090**), with its landmine and this
open question written down.

**Council gate: submitted, `fca1071b-80ac-40cd-8c6d-d30a735de89b`, verdict pending at
filing time.** This RFC is not an appeal of anything and does not duplicate that round —
the gate reviews the fix I wrote; this asks a question the gate is the wrong instrument for.

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

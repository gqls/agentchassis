# 259 — three `MessageProcessor` paths guard on `p.sqlDB`, which is nil in production, so none of them has ever run

> **⚠ 259 IS AN AMBIGUOUS NUMBER.** An unrelated `bugs_open/259` was filed the same day —
> `259_HANDOFF_2026-08-12_one_provision_request_builds_several_billable_gpus.md` (GPU
> provisioning). Numbers are never reassigned on this estate and duplicates are tolerated,
> so **resolve this case by its slug** (`three_processor_paths_guard_on_a_handle...`) and
> `git log` the FILE PATH, never the bare number. A commit message saying "259" may well
> mean the other one.
>
> **OWNED BY THE `bugfix_239` LANE** as 239's continuation (accepted 2026-08-12; they
> re-verified all three sites at source before accepting). Filed by the 246 lane, which is
> not working it — contribute into that lane, do not start a parallel fix.

**Filed 2026-08-12** by the `bugfix_246_shared_pool_ownership` lane, found while scoping
246's follow-up D5 ("collapse `p.sqlDB` into `p.db`"). What looked like a cosmetic
two-handles tidy-up is three production-dead code paths sharing one cause.

> ## STATUS 2026-08-14: **FIXED IN THE TREE, STILL OPEN — the fix is a Go change and is inert until a fleet roll.**
>
> Candidate 1 applied at `e37f79b65` (+ `f894b1a38`, gofmt). `p.sqlDB` no longer exists;
> `MessageProcessor` has one handle. Council gate: `Council-Submitted:
> 0ff072ef-ee02-465e-8a70-f5461c585ec9`, verdict pending at the time of writing — **whoever
> picks this up must read it and act on a REVISE/REJECTED, because the code is already on the
> shared branch.**
>
> **It stays OPEN because the bar is fixed AND live**, and the defect is reproducible on every
> chassis pod until the roll ships. **Do not move it to `bugs_closed/` on the strength of the
> commit.** What to do after the roll is in "How to verify any fix" below; the two live paths
> are the honest check and both could come out either way.
>
> **What went, beyond the three sites this file filed** — two more dependents of the same nil
> handle, found while re-locating the line numbers and recorded here because a reader of this
> file would otherwise not know they were in the blast radius:
> - **`stateRepo`** — a `*orchestration.StateRepository` field built in the constructor from
>   the nil `sqlDB` and read **nowhere** in `platform/messaging` (grep returns exactly two
>   hits: the declaration and the assignment; the identifier is unexported, so nothing outside
>   the package could reach it). Left behind, it would have silently become
>   `NewStateRepository(nil, logger)`.
> - **`createSQLDB()`** — a callerless second opener of a second handle to the same database,
>   same class, same file. Not `DATABASE_URL`-based (it reads `CLIENTS_DB_*`), which is why
>   this file never listed it.
>
> **And a third live fallback reader this file did not list:**
> `platform/messaging/validation_drop.go` (`recordDroppedValidationError`) had
> `db := p.sqlDB; if db == nil { db = p.db }` — **the operands the opposite way round** from
> the two `bugs_closed/239` left in `processor.go`. So it reached first for the handle that was
> always nil and only ever wrote its `agent_error_log` row by falling through. Simplified to
> `p.db` like the others.
>
> **One piece of evidence added to C's case**, because the file argued redundancy from the
> `processed_messages` write rate alone: the `bugs_closed/239` **release** half inside C's defer
> is duplicated verbatim in `agentbase` (`agent.go`, the two-phase claim defer, on
> `a.stateRepo`). `ReleaseMessageClaim`'s only two callers fleet-wide were that one and C's. So
> deleting C loses neither the claim, nor the release, nor the completion — not just "the claim
> happens elsewhere".
>
> ### Council verdict: **APPROVED** first round, and the two advisory objections are ANSWERED below
>
> `0ff072ef-ee02-465e-8a70-f5461c585ec9`, 2026-08-14 07:57:26Z. 10 reviewers, 8 approve, 2 object
> (guardian, prior-art-librarian), **none high-severity**, `gated_by_truncation: false`.
> `decided_by`: *"approved with 2 advisory objection(s) — none high-severity"*.
>
> Both medium objections land on the **same** point, and it is a fair one: the plan asserted
> agentbase's gate equivalence in its own risks block rather than measuring it. Guardian: *"The
> safety argument rests entirely on agentbase's equivalent claim always running with an
> equivalent gate (`a.isStateless && a.stateRepo != nil`); the plan's own risks section admits
> this is unverified… confirm the gate equivalence (**or that this site was truly always dead
> regardless of gate state**) before merge, not after."*
>
> **Answer, and note it satisfies BOTH of the guardian's disjuncts — the second one decisively.**
>
> **(1) The deletion is a no-op regardless of agentbase's gate.** Two claims were being run
> together and they need separating: *"deleting C changes nothing"* requires only that **C never
> ran**, which follows from `p.sqlDB` being nil unconditionally — a fact about `DATABASE_URL`,
> nothing to do with agentbase. *"The chassis still dedupes"* is a claim about **agentbase**, and
> it describes the state of the world **before** this change as much as after. If agentbase's
> gate were ever nil, dedupe was **already** absent on that path, because the thing that would
> have had to cover it was dead code. So no deletion here can introduce a dedupe gap. That is
> the guardian's second disjunct and it holds unconditionally.
>
> **(2) The gate equivalence they asked for, now MEASURED — and it is stronger than asserted.**
> Read at `platform/agentbase/agent.go`:
>
> ```go
> if a.config.DatabaseURL != "" {          // :268 — one condition
>     db, err := sql.Open("pgx", a.config.DatabaseURL)
>     if err != nil { return fmt.Errorf(...) }   // so db is valid past here
>     ...
>     a.db        = db                                            // :303
>     a.stateRepo = orchestration.NewStateRepository(db, a.logger) // :304
> }
> ...
> a.processor = messaging.NewMessageProcessor(…, a.db, …)   // :357 — a.db BECOMES p.db
> ```
>
> `a.db` and `a.stateRepo` are assigned on **adjacent lines inside one `if`**, so
> **`a.stateRepo != nil` ⟺ `a.db != nil` ⟺ `p.db != nil`.** The agentbase gate is therefore
> *exactly* as strong as "the processor has a live handle at all" — it cannot be weaker. And
> `isStateless` is declared once (`:84`), assigned `true` once (`:226`) and read once (`:1162`),
> **never false**, so the gate reduces to the `stateRepo` test alone.
> **Conclusion: there is no code path on which site C could have de-duplicated and agentbase
> could not.** The load-bearing invariant the guardian listed under `missing` is now checked
> rather than asserted.
>
> **(3) prior-art's low objection — `DATABASE_URL` on the running pods — re-confirmed rather than
> inherited.** `[MEASURED 2026-08-14]` both live chassis pods (`agent-chassis-64bfd68fd6-7ft4l`,
> `-dc7dm`): `DATABASE_URL` **empty**, `CLIENTS_DATABASE_URL` **set**. The seat was right that a
> deployed-env fact has no check tier and must come from a human; this is that check, taken today
> rather than carried from 08-13.
>
> **Two objections are correctly left OPEN as post-roll work, not answered here:** guardian's low
> on site B (*"zero-caller claim is grep-based against the working tree, not the pushed code
> index — should be independently re-verified post-merge"*) and prior-art's medium on site A
> (*"the code index stores declarations only, never function bodies, so this specific claim is
> unverifiable via any check available to me"*). Both are limits of the reviewers' tooling, not
> gaps in the evidence — the site A proof is exhibited in full above for exactly this reason —
> but both deserve the human re-read at HEAD they ask for. **`debug_historian` adds a caveat
> worth carrying into the post-roll check: the fleet is MIXED for hours after a release, and
> `-l app=agent-chassis` selects 2 pods of the many running that binary — so verify
> behaviourally, per service, not by one pod's grep.**
>
> **Coverage changes, stated rather than left to shrink quietly:** one test REMOVED
> (`TestConstructorSizesThePoolItOpensItself` — its whole subject was the deleted second pool;
> its salvageable half survives as a `DATABASE_URL`-set case in the sibling table), two vacuous
> `sqlDB` assertions dropped from 239's regression test which is otherwise kept, and
> `TestSuccessResponseStatusStillComplete` redirected from the deleted `sendWorkflowResponse`
> wrapper to `sendWorkflowResponseWithStatus` with identical arguments. That redirect was
> **mutation-verified** rather than assumed: forcing the error-only status override to fire
> unconditionally makes it fail on `IsComplete` and `IsError`, and restoring makes it pass — so
> the `bugs_open/196` success-envelope coverage is genuinely preserved, not merely still green.

**Status: ~~OPEN, not fixed~~ — but as of 2026-08-13 fully DIAGNOSED and ready to fix.** All
three sites are now assessed and all three are **deletions**, each on its own separate
proof: **C** redundant (agentbase runs the same claim on a live handle), **B** unreachable
(zero callers), **A** inert (both branches `return nil`, one log line between them).
Recommended fix is **candidate 1** — drop `p.sqlDB` and all three sites together — with care
reserved for the two `bugs_open/239` fallback readers (`:649-658`, `:1192`), which are live
and must be simplified to `p.db` rather than removed. Platform code: council gate + register
touch owed.

~~Severity is **not uniform across the three** — one is harmless redundancy, one substitutes
a placeholder for real data, one is unassessed. Do not fix them as a batch.~~ **The
"placeholder for real data" severity on B was withdrawn 2026-08-12** (the function that
builds it has no callers), and **A's "unassessed" was resolved 2026-08-13**. "Do not fix as
a batch" still holds for the *reasoning* — see "why this is not a sweep" below — but the
three conclusions have converged on delete.

## The mechanism (same one `bugs_open/239` already fixed once)

`MessageProcessor` holds two database handles:

- `p.db` — passed in by `agentbase` at `agent.go:341`. **Non-nil in production.**
- `p.sqlDB` — opened inside `NewMessageProcessor` from **`DATABASE_URL`**, which is
  **not set on chassis pods** (they set `CLIENTS_DATABASE_URL`, a different variable).
  **Always nil in production.**

Three sites gate real work on `if p.sqlDB != nil`. All three are therefore unreachable
on every chassis pod, and have been since the handle was introduced.

**This is not a new class.** `bugs_open/239` found and fixed a *fourth* instance
(`recordDispatchFailureState`) with `db := p.db; if db == nil { db = p.sqlDB }`, and
recorded the lesson in its handoff: *"`p.db` and `p.sqlDB` are NOT interchangeable —
`sqlDB` is nil on the chassis, and building a test fixture with it set tests a shape
production lacks."* **Nobody swept the remaining sites.** That sweep is this bug.

## The three sites (line numbers at `039cfce84`+; re-locate before editing)

### A — `~:351`, child-workflow completion check → ~~**[UNASSESSED]**~~ **RESOLVED 2026-08-13: DEAD *and* INERT — the early return suppresses nothing, and no treatment of A can change a response**
```go
if msgCtx.IsChildOrchestration() {
    if p.sqlDB != nil {
        repo := orchestration.NewStateRepository(p.sqlDB, msgCtx.Logger)
        state, _ := repo.GetState(ctx, msgCtx.ExecutionContext.OrchestrationID)
        if state != nil && state.Status == orchestration.StatusCompleted {
            return nil   // suppress the response to the parent
        }
    }
}
```
~~Dead ⇒ the early return never fires ⇒ a completed child's response is never suppressed.
Whether that produces a duplicate parent response, or is merely belt-and-braces over a
guard elsewhere, is **not established**. Do not assume either way.~~

> **RESOLVED 2026-08-13 (dispatch/pool lane). The answer is NEITHER, and it needed no live
> measurement — the question dissolves on reading the fifteen lines around it.**
>
> **The premise was false at source.** That `return nil` does not "suppress the response to
> the parent", because there is no response-sending code for it to skip. Look at what
> actually sits between the guard and the end of the function (`processor.go:350-364`,
> v. HEAD `8ac5e9cce`):
>
> ```go
>     if msgCtx.IsChildOrchestration() {
>         if p.sqlDB != nil {
>             ...
>             if state != nil && state.Status == orchestration.StatusCompleted {
>                 msgCtx.Logger.Info("Child workflow completed, sending response to parent")
>                 return nil          // <-- the "suppressing" return
>             }
>         }
>     }
>
>     msgCtx.Logger.Info("Workflow successfully handed off to the orchestrator")
>     return nil                      // <-- the fall-through
> }
> ```
>
> **Both paths `return nil`, and the entire skipped region is one log statement.** So the
> early return is indistinguishable from the fall-through by anything outside the function:
> - `process()` is declared `func (p *MessageProcessor) process(ctx context.Context, msgCtx *MessageContext) error` at `:141` — an **unnamed** `error` return, so no `defer` can rewrite it;
> - there is **no `defer` anywhere in `process()`** (`:141`–`:364`, checked as a range, not as a guess);
> - the two returns carry the **identical value**, so the caller cannot tell them apart.
>
> **Therefore, for each of the three possible treatments of A:**
> | treatment | behavioural effect |
> |---|---|
> | leave it (today) | branch never entered; logs `:362`, returns nil |
> | **delete the block** | identical — logs `:362`, returns nil. **A provable no-op, not a live-behaviour change.** |
> | activate it on `p.db` | one `GetState` read per child message, and *which* of two log lines is emitted. Still cannot suppress or emit any response. |
>
> **What this changes for the fix.** The ordering constraint in "Fix candidates" — *"A must be
> assessed first, because removing the handle turns A's guard into an ordinary `p.db` read,
> which is a live-behaviour change"* — **dissolves.** Deleting A alters no behaviour, so
> **all three sites are now deletes** and candidate 1 (drop `p.sqlDB` entirely) is unblocked
> on A's account. The remaining care belongs to the `db := p.db; if db == nil { db = p.sqlDB }`
> fallback readers (`:649-658`, `:1192`) left by `bugs_open/239`, which must be simplified
> rather than deleted.
>
> **The premise it rested on, re-grounded rather than inherited** — `[MEASURED 2026-08-13,
> v1.0.1295]` `DATABASE_URL` is **empty** on both live chassis pods while
> `CLIENTS_DATABASE_URL` is set, so `p.sqlDB` is nil in production as the filing said.
>
> **Also worth fixing while deleting: the log line lies.** `"Child workflow completed,
> sending response to parent"` sits on a path that sends nothing and returns nil. If A is
> ever revived instead of deleted, that string is the next reader's trap — it is the only
> reason this site reads as response-handling at all, and it is what made the question look
> like it needed a cluster measurement.
>
> **Why no `090` run and no live behavioural measurement, stated rather than skipped** (owner
> ruling 2026-07-31): the claim here is a **negative about fifteen contiguous lines with no
> indirection** — no interface dispatch, no goroutine, no `defer`, no named return, and a
> fully-enclosed skip region. The disconfirming evidence would have to be *inside* the quoted
> block, so it is exhibited above in full for the next reader to re-check in one screen
> rather than take on trust. This is the "local and self-evidencing" exemption, used
> deliberately and narrowly: it does **not** extend to the deletion itself, which is platform
> code and owes the council gate and a register touch.

### B — `~:582`, the workflow's final result → **the one with a visible consequence**
```go
var finalResult interface{}
if p.sqlDB != nil {                              // never true
    repo := orchestration.NewStateRepository(p.sqlDB, msgCtx.Logger)
    state, err := repo.GetState(ctx, msgCtx.ExecutionContext.OrchestrationID)
    if err == nil && state != nil {
        finalResult = state.CollectedData        // the REAL result
    }
}
if finalResult == nil {
    finalResult = map[string]interface{}{"status": "completed"}   // always taken
}
return p.sendWorkflowResponse(ctx, msgCtx, finalResult)
```
**Every workflow response sent from this path carries the literal placeholder
`{"status":"completed"}` instead of the orchestration's `CollectedData`.** That is a
code-reading certainty, not an inference: with the guard dead, `finalResult` is nil at
the `if`, so the placeholder branch is unconditional.

A concrete downstream effect, also from reading: `sendWorkflowResponseWithStatus` scans
`result.(map[string]interface{})` for a nested map to identify `lastActionResult`
("suppress premature responses"). The placeholder's single value is a **string**, so
that scan finds nothing to inspect on every call.
~~`[UNMEASURED]` — what a parent actually does with the stub, and whether any live
pipeline depends on the real `CollectedData` arriving by this route. **Measure before
fixing:** turning this on changes what every parent receives.~~

> ## CORRECTION 2026-08-12 (`bugfix_239` lane, on taking this bug): **B's containing function has NO CALLERS. Nothing ever receives the stub.**
>
> The `[UNMEASURED]` question above is not merely unanswered — it is **unaskable**, and the
> answer is zero. Site B lives in `sendWorkflowSuccessResponse` (`processor.go:567`), and
> that function is **called from nowhere in the repository**:
>
> ```
> $ grep -rn "sendWorkflowSuccessResponse" --include=*.go .
> platform/messaging/processor.go:567:func (p *MessageProcessor) sendWorkflowSuccessResponse(...)
> platform/messaging/processor.go:572:	...Info("In file processor.go sendWorkflowSuccessResponse",   # its own log string
> ```
> Two hits: the definition and a log literal inside it. **Control** (so the method is known
> to find callers when they exist): the same grep for the live sibling
> `sendWorkflowFailureResponse` returns its definition *plus* a real call site at
> `processor.go:~347`. Tests included in both greps; an unexported method can only satisfy
> an interface within its own package, and the name appears in no interface.
>
> **The dead cluster is larger than one function.** `sendWorkflowResponse` (the inner one)
> has exactly ONE non-test caller — line `:594`, inside the dead function — so it is dead
> too. And the placeholder literal is built in exactly one place fleet-wide:
> ```
> $ grep -rn '"status": *"completed"' --include=*.go platform/ | grep -v _test.go
> platform/messaging/processor.go:591          # inside the dead function
> ```
> **The live response path does not pass through any of it**: `sendWorkflowResponseWithStatus`
> (`:804`) is reached directly from `:625` and `:2131`, never via B.
>
> **What this changes.** (1) The statement "every workflow response sent from this path
> carries the placeholder" is true but **vacuous** — no response is ever sent from this
> path. (2) The `sendWorkflowResponseWithStatus` `lastActionResult` scan finding nothing is
> likewise never exercised. (3) B's remedy moves from *measure-then-fix* to **delete**, and
> B joins `bugs_open/247`'s family rather than being the delicate one. (4) The stated risk
> that motivated caution here — *"turning this on changes what every parent receives"* —
> **does not exist**: there is nothing to turn on, because the caller does not exist. Do
> NOT "fix" B by adding the `p.db` fallback; that would resurrect a dead path and *create*
> the live-behaviour change the file warned about.
>
> **Why the filing said otherwise, recorded because the reasoning was sound:** the filer
> read the control flow inside the function correctly and marked the downstream honestly as
> `[UNMEASURED]`. The missing step is one every reader of this class owes — **a function's
> internal certainty says nothing about whether the function runs.** Reachability is a
> separate question from correctness, and it is one grep. That check is now the first thing
> this file's "How to verify any fix" section asks for.
>
> **Unchanged by this correction:** A (live path, `process()`) and C (live path,
> `ProcessMessage()`) are both genuinely reachable — confirmed by the same method — so the
> file's central warning stands for them, and C's redundancy evidence is untouched.

### C — `~:1486`, the two-phase dedup claim → **REDUNDANT, and that is now evidenced**
The whole `bugs_open/003` F3 two-phase claim in `ProcessMessage` sits behind
`if p.sqlDB != nil`, including the comment explaining that "an un-dedupable message must
be LOUD". It has never run.

**It is not a dedup hole.** `agentbase` performs the same two-phase claim itself, on
`a.stateRepo` (built from `a.db`, non-nil), at `agent.go:1149-1173` — and it is
demonstrably working:

```sql
SELECT date_trunc('hour', processed_at) AS hr, status, count(*) AS rows,
       count(DISTINCT processed_by) AS writers
FROM processed_messages WHERE processed_at > now() - interval '4 hours'
GROUP BY 1,2 ORDER BY 1 DESC;
```
Measured 2026-08-12: **449 rows / 82 distinct writers** in the 13:00 hour, and the claim
lifecycle is visible in the data (3 rows `processing`, the rest `complete`). Rows are
being written continuously, so the live layer is claiming and completing.

> The log markers are **not** the check to use here. `Duplicate message ignored` and
> `DEDUPE_CLAIM_LOST` fire only when a duplicate actually occurs, so their absence
> (observed) means nothing. The table is the instrument; the log lines are an event.

The processor's copy is therefore **a second, dead implementation of a judgement the
platform already makes correctly elsewhere** — the `bugs_open/247` shape exactly: a
plausible-looking mechanism that a session tracing "where does the chassis dedupe?" will
find and reason about.

## Why this is NOT a sweep, and must not be fixed as one

The tempting fix — mechanically apply 239's `db := p.db; if db == nil { db = p.sqlDB }`
to all three — is **wrong**, and dangerously so, because it reads as consistency:

- On **C** it would **switch on a second dedup layer** that has never run. The code's own
  comment says the two layers were designed to coexist via "the same-pod lease exemption
  in `HasProcessedMessage`", so it may be safe — but "may be safe" is not a reason to
  enable a concurrency mechanism fleet-wide inside a tidy-up. **Deleting C is strictly
  safer than fixing it**, and matches the estate's rule of one implementation per
  judgement.
- On **A** and **B** it **changes live behaviour** — what parents receive, and whether
  responses are suppressed. Neither has a measured blast radius yet.

~~**None of these three is a no-op fix.** That is the opposite of `bugs_open/246`, whose
safety rested on the deleted calls being byte-identical to the defaults; do not carry
246's confidence across to this bug.~~

> **CORRECTED 2026-08-13: all three ARE no-op deletions, established site by site — but the
> reasoning that gets you there is nothing like 246's, so the warning's *spirit* stands.**
> C is redundant (agentbase does the same claim on a live handle, evidenced). B is
> unreachable (`sendWorkflowSuccessResponse` has zero callers). A is inert (both paths
> `return nil`; the skipped region is one log line). 246's safety came from a **symmetry
> argument** — deleted calls being byte-identical to defaults. Here it comes from **three
> different, individually-exhibited proofs**, which is why the file was right to refuse the
> mechanical sweep in candidate 3: the *conclusion* converged, the *justifications* did not,
> and a sweep would have asserted the conclusion without any of them.

## Fix candidates, ordered by what closes the door

1. **Delete `p.sqlDB` entirely, and with it C.** The handle is nil in production, every
   remaining reader already falls back to `p.db`, and the field's only effect is to make
   three paths unreachable and mislead readers. Removing it makes the bad state
   **unrepresentable** — there is no second handle to guard on. ~~A and B then become
   ordinary `p.db` reads and must be assessed on their own merits (below) before their
   guards are removed.~~ **2026-08-13: A and B are now both assessed and both are deletes,
   so this candidate is UNBLOCKED and is the recommended fix.** Delete all three sites with
   the field. The one place needing care is not A, B or C but the two fallback readers
   `bugs_open/239` left behind — `:649-658` and `:1192` (`db := p.db; if db == nil { db = p.sqlDB }`)
   — which must be **simplified to `p.db`**, not deleted, since they are the live paths.
2. ~~**Assess A and B individually, with a measurement each, and fix them separately.**
   B first: it has a certain, describable defect (a placeholder where real data belongs)
   and needs a live measurement of what consumes the response. A needs the duplicate
   question answered.~~ **SUPERSEDED 2026-08-13 — both assessments are done and neither
   needed the live measurement this candidate budgeted for.** B's consumer set is empty
   because the function has no callers; A's duplicate question is void because both of its
   paths return the same nil. Kept on the record because the candidate was correctly
   ordered at filing time: it was right to demand the assessments, and wrong only in
   expecting them to require cluster evidence.
3. Fix the guards mechanically across all three. **Rejected** — see above.
4. Document and leave. **Rejected**: a comment is not a control, and the 239 lane already
   proved this class bites for real.

## How to verify any fix

- **FIRST, for each site: is the containing function REACHABLE?** One grep, before any
  reasoning about what the code does — `grep -rn "<funcName>" --include=*.go .` — and read
  the hit count against a **control** (a sibling known to be live, so you can tell a
  working grep from a lucky one). This is now first because it was skipped on site B, whose
  internal control flow was analysed correctly and whose function turned out to have no
  callers at all: **a function's internal certainty says nothing about whether it runs.**
  Result of that check, 2026-08-12: **A reachable** (`process()`), **B DEAD**
  (`sendWorkflowSuccessResponse`, zero callers), **C reachable** (`ProcessMessage()`).
- **Never build a fixture with `sqlDB` set** — that tests a shape production does not
  have (239's recorded trap; `processor_dispatch_resolution_test.go:433-454` exists
  precisely to pin the production shape `db` set / `sqlDB` nil).
- For B, the disconfirming observation is a parent receiving real `CollectedData` where
  it previously received `{"status":"completed"}`.
- For C, after deletion, `processed_messages` must keep showing the same write rate — the
  agentbase layer is doing that work and must be unaffected.

## Relations

`bugs_open/239` (fixed the fourth instance; its handoff carries the non-interchangeable
warning) · `bugs_open/246` (the constructor that opens `sqlDB`; D5 in its handoff is
this bug, now upgraded from tidy-up to defect) · `bugs_open/247` (dead code that reads as
the live path — same misleading-signpost class) · `bugs_open/003` (the F3 two-phase claim
that C is a dead copy of).

---

## On the 090 requirement (owner ruling 2026-07-31), stated rather than skipped

This file asserts a structural, cross-cutting cause, so the standing rule is that it is
not "filed" until it has been through the diagnosis loop **or the filing session states
plainly why it substituted equivalent first-hand verification**. This is that statement.

**090 was not run for this bug, and the reason is specific rather than convenient: the
loop is structurally blind to the fact this bug turns on.** The whole mechanism rests on
`DATABASE_URL` being unset on chassis pods — a **Kubernetes Deployment environment**
fact. The loop's evidence surface is the repo plus `clients_db`; it has no `kubectl`.

That is not speculation. **It was measured on this exact area yesterday.** Run
`105970e4-dd02-4654-9536-84a2dd6a3da2` (filed by this lane for `bugs_open/246`, whose
mechanism also turned on an env var) asked precisely the right question, went looking for
the key in `agent_definitions.env_vars`, got `(0 rows)`, re-ran the identical query a
second iteration and got `(0 rows)` again, then exhausted all five iterations and
terminated `COMPLETED` with **five `bundle` artifacts, zero `council_report`, and no
verdict note**. Worse, the evidence it *could* reach pointed the wrong way: `(0 rows)`
reads as "nothing sets this, so it does not matter". Written up as a landmine
(`LANDMINES.md`, "The 090 diagnosis loop cannot read a Kubernetes env var").

**What was done instead, all first-hand and all disconfirmable:**

1. **The env fact, from the pods themselves** — `DATABASE_URL` unset, `CLIENTS_DATABASE_URL`
   set, on both chassis replicas.
2. **The dead-path claim, from the code at HEAD** — all three guards read and quoted above,
   not paraphrased. Site B's placeholder is a control-flow certainty, not an inference.
3. **The "is C a hole?" question, from live data with a demand control** —
   `processed_messages` shows 449 rows / 82 writers in one hour with the claim lifecycle
   visible, so the agentbase layer is provably doing the work. **The check that would have
   been blind was tried first and rejected**: the `Duplicate message ignored` /
   `DEDUPE_CLAIM_LOST` log markers read zero, and zero there means nothing, because those
   lines fire only when a duplicate occurs.
4. **The prior art** — `bugs_open/239` fixed a fourth instance of this exact guard and
   recorded the non-interchangeability warning; this bug is the unswept remainder.

**Where first-hand verification did NOT reach, marked rather than papered over:** the
downstream consequence of B (what a parent does with the stub) and the whole of A are
`[UNMEASURED]`, and the fix candidates say so. A 090 would not have closed those either —
they need a live behavioural measurement, which is the next session's first task.

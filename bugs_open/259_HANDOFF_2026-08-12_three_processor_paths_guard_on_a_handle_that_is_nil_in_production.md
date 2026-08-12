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

**Status: OPEN, not fixed.** Severity is **not uniform across the three** — one is
harmless redundancy, one substitutes a placeholder for real data, one is unassessed.
Do not fix them as a batch. See "why this is not a sweep" below.

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

### A — `~:351`, child-workflow completion check → **[UNASSESSED]**
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
Dead ⇒ the early return never fires ⇒ a completed child's response is never suppressed.
Whether that produces a duplicate parent response, or is merely belt-and-braces over a
guard elsewhere, is **not established**. Do not assume either way.

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

**None of these three is a no-op fix.** That is the opposite of `bugs_open/246`, whose
safety rested on the deleted calls being byte-identical to the defaults; do not carry
246's confidence across to this bug.

## Fix candidates, ordered by what closes the door

1. **Delete `p.sqlDB` entirely, and with it C.** The handle is nil in production, every
   remaining reader already falls back to `p.db`, and the field's only effect is to make
   three paths unreachable and mislead readers. Removing it makes the bad state
   **unrepresentable** — there is no second handle to guard on. A and B then become
   ordinary `p.db` reads and must be assessed on their own merits (below) before their
   guards are removed.
2. **Assess A and B individually, with a measurement each, and fix them separately.**
   B first: it has a certain, describable defect (a placeholder where real data belongs)
   and needs a live measurement of what consumes the response. A needs the duplicate
   question answered.
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

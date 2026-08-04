# 195 — a workflow rejected by `ValidateWorkflow` is classified as TRANSIENT, retried, and leaves no durable record anywhere

**Filed 2026-08-04** by the `bugs_open/173` lane, which lost three dispatches and ~15 minutes
to it while running an induced fault. **Status: OPEN, UNOWNED.**
**Severity: medium.** Nothing is corrupted; the cost is diagnosability and wasted retries. But
it is a **hole in `bugs_closed/034`'s guarantee**, and 034 exists precisely so that this class
of failure cannot be invisible.

> **VERIFICATION STATEMENT, per the owner ruling of 2026-07-31.** That ruling says a
> `bugs_open/` file asserting a cross-cutting or structural cause is not filed until it has
> been through `090_TRIGGER_needs_diagnosis_v1.sh`, or the filing session states plainly why
> it substituted equivalent first-hand verification. **I substituted first-hand verification,
> and it consisted of an induced fault plus a code read, not inference:**
>
> - **I induced it deliberately, twice.** Once by accident (my own malformed induction seed for
>   `173`), then once on purpose with a throwaway agent whose only step is
>   `{"action":"complete"}` — invalid, because `complete` is a step *name* whose action is
>   `complete_workflow`. Probe correlation `34268b8a-c93a-4e2b-8b20-91bfc6e4bf81`,
>   orchestration `ca528ff3-8992-4d18-8020-ed24d29cd812`, 2026-08-04, chassis `v1.0.1250`.
> - **I read the whole path in the running source**, not from grep hits:
>   `platform/messaging/processor.go:275-282` (the rejection), `platform/agentbase/agent.go:1181-1214`
>   (the classifier branch), `platform/messaging/*.go` (`ValidationErrorNeedles`,
>   `MatchedValidationNeedle`).
> - **Every negative claim below carries a positive control** — the "0 rows in
>   `agent_error_log`" claim is paired with **1,301 rows in the same table over 24h**, so a row
>   was plainly reachable.
> - **What makes the substitution safe:** this is a *closed* claim about one code path I read
>   end to end and then made fire on demand. There is no "the cause is somewhere else" risk,
>   because I caused it and watched the exact branch it took.
> - **What I did NOT establish** is listed under "Not established" and is not asserted anywhere
>   above it.

## Symptom

Dispatch an orchestration whose agent definition fails `ValidateWorkflow`. You get:

- **no `orchestration_states` row** — the orchestration never starts;
- **no `agent_error_log` row**;
- **no row in `orchestration_state_audit` or `diagnosis_artifacts`**;
- **no error response to the sender**;
- **one ephemeral pod log line**, which rotates. Mine was **completely gone within 8 hours**.

From the dispatcher's side this is indistinguishable from queue latency — and CLAUDE.md's own
(correct, for other cases) guidance is *"a missing orchestration row is almost always latency,
not a dropped dispatch — do not retry on that evidence"*. So the guidance actively steers you
away from the truth here. That is what cost the `173` lane three dispatches.

**And it is retried.** `Processing failed` appeared **twice** for one probe — a static config
error being re-attempted, which cannot ever succeed.

## Root cause — a case-sensitive substring classifier that does not match its own commonest input

`platform/agentbase/agent.go:1189` decides whether a processing failure is a *permanent
validation* error (drop it, record it durably, do not retry) or a *transient* one (retry it):

```go
if needle := messaging.MatchedValidationNeedle(errMsg); needle != "" { … }
```

and the needle list is:

```go
var ValidationErrorNeedles = []string{"is required", "validation", "invalid", "missing"}
// MatchedValidationNeedle uses strings.Contains — CASE-SENSITIVE
```

The error a rejected workflow actually produces is (verbatim, from
`complete_work_item_verification_test.go:29` and reproduced by the probe):

```
WORKFLOW_INVALID: Invalid workflow configuration (caused by: step 'done' with action 'complete' requires a topic)
```

Match it against the four needles:

| needle | present? | why not |
|---|---|---|
| `"is required"` | **no** | the text says *"requires a topic"* |
| `"validation"` | **no** | the word never appears |
| `"invalid"` | **no** | it appears as `WORKFLOW_INVALID` and `Invalid` — **capitalised**, and the match is case-sensitive |
| `"missing"` | **no** | — |

**So the single most common permanent configuration error on this platform matches none of the
four needles, purely on capitalisation and wording.** It therefore takes the `else` branch:
`observability.MessagesFailed` + `handleProcessingError` — i.e. treated as transient, retried,
and **`recordDroppedValidationError` is never called**.

That last point is the sharp one. `recordDroppedValidationError` is *the fix `034` built* so
that a dropped validation error would leave a queryable row. Its own comment says so:

> *"a dropped validation error otherwise leaves NO durable trace — `handleProcessingError` is
> skipped, so the parent gets no error response, and the only record is an ephemeral pod log
> plus a label-only counter that cannot name the message. Persist a queryable row so the drop
> is investigable."*

**The row is never written, because the branch that writes it is never entered.**

## Measured, with controls

Probe orchestration `ca528ff3-8992-4d18-8020-ed24d29cd812`, 2026-08-04, chassis `v1.0.1250`:

| observation | value | control |
|---|---|---|
| `Invalid workflow configuration` log lines | **1** | confirms the fault fired at all |
| `not calling handleProcessingError` warns (needle HIT) | **0** | **the classifier did not match** |
| `Processing failed` for that correlation | **2** | it was **retried** |
| rows in `agent_error_log` | **0** | **1,301 rows in the same table over 24h** — a row was reachable |
| rows in `orchestration_states` | **0** | the table is written constantly by the live fleet |
| rows in `orchestration_state_audit` / `diagnosis_artifacts` | **0** | — |

The needle-HIT count of 0 is the load-bearing measurement: it is what separates *"the durable
record is broken"* from *"the message was never classified as a validation error in the first
place"*. It is the second.

## Why `bugs_closed/034` does not already cover this

034 is closed and its four drop sites are live and induction-proven. Its site 4 is *"agent.go
substring classifier"* — but 034 fixed the case where the classifier **matches** (ensuring a
durable row is written before the drop). **This bug is the case where it fails to match.**
034 never examines whether the needle list is *correct*, and grepping the closed file for
`ValidateWorkflow`, `Invalid workflow`, or any mention of case sensitivity returns **nothing**.

So this is not a regression in 034 and not a duplicate of it: it is the assumption 034's fix
rests on — *"a validation error will be recognised as one"* — going unchecked. A fix that
guarantees a durable record **conditional on a classifier** inherits every gap in that
classifier.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Classify on the typed error CODE, not on the message text.** The error is already
   constructed as `errors.New(errors.ErrWorkflowInvalid, "Invalid workflow configuration")`
   (`processor.go:280`) — the code is right there and is exact. Matching on it cannot be
   defeated by rewording or capitalisation, and it removes the whole class rather than this
   instance. **This is the one that closes the door**: a substring classifier over
   human-readable prose is a category error, and every future error string is another chance
   to reopen it.
2. **Write the durable record on EVERY processing failure, not only on the classified ones.**
   Complements (1) rather than competing: it makes the *trace* independent of the
   classification being right at all, so a future mis-classification costs wasted retries but
   never invisibility. The retry decision can stay where it is.
3. **Add needles / make the match case-insensitive.** Cheapest, and the worst of the three:
   it is more of the same fragile mechanism, it fixes exactly today's spelling, and the next
   error phrasing slips through in silence again. If taken, it should be explicitly temporary.

**Prefer (1) + (2).** (1) makes classification exact where the information already exists;
(2) means being wrong about classification is never *invisible* again.

**Note for whoever takes it:** `MatchedValidationNeedle` returns the needle rather than a bool
specifically so that an unanchored match is auditable — a deliberate, good decision from 034's
lane. Do not remove that property; a code-based classifier should record *which* code matched
for the same reason.

## Not established — do not inherit these as findings

- **Whether a parent agent hangs.** My probes were dispatched from the CLI, so nothing was
  awaiting a reply. `recordDroppedValidationError`'s own comment says the skipped path means
  *"the parent gets no error response"*, and `handleProcessingError` does run here — but
  **[UNVERIFIED]** what it sends, and whether a `call_agent`/`spawn_agent` parent receives a
  failure or waits out its timeout. If it waits, this is more serious than diagnosability and
  it touches the hung-spawn family (`bugs_open/029`). **Worth measuring first.**
- **How many retries** a mis-classified permanent error consumes before giving up, and what
  that costs the dispatch lanes. I observed `Processing failed` **twice** for one probe and did
  not chase the retry cap. **[UNMEASURED]**
- **Which other permanent error classes miss the needles.** I checked exactly one
  (`WORKFLOW_INVALID`). A census of error constructors against the four needles would say
  whether this is one gap or a family. **[UNMEASURED]**

## How to verify a fix

**A green build proves nothing — induce it.** Seed a throwaway agent whose workflow is invalid
and dispatch it:

```sql
INSERT INTO agent_definitions (type, display_name, description, category, default_config, is_active)
VALUES ('test-invalid-workflow-probe','probe','temporary','experimental',
 '{"workflow":{"start_step":"done","processing_mode":"orchestrator","timeout_seconds":60,
   "steps":{"done":{"action":"complete","description":"invalid on purpose"}}}}'::jsonb, true);
-- DELETE FROM agent_definitions WHERE type='test-invalid-workflow-probe';   -- cleanup, owed
```

Dispatch it with the envelope in
`docs024_key_docs_latest/bugfix_173_substep_error_tolerance/RUNBOOK_substep_error_tolerance.md` §R8,
then assert **all three**:

1. a durable row naming the orchestration exists (`agent_error_log`, or wherever the fix puts it);
2. `not calling handleProcessingError` **or** the code-based equivalent fires — i.e. it is
   classified as permanent;
3. `Processing failed` appears **once**, not twice — it is not retried.

Today those read **0 rows / 0 / 2**. Any fix that does not move all three has not closed it.
And pair the row count with a positive control on the same table, or a `0` proves nothing.

## Related

- `bugs_closed/034` — the durable-record guarantee this gap undermines; read its four drop
  sites first, and note this is a fifth *mode* (mis-classification), not a fifth site.
- `bugs_open/029` — hung spawns / dispatch saturation, IF the parent-hang question above turns
  out to be real. Speculative until measured.
- `bugs_open/193` — filed the same day, same family in miniature: a config value that is
  present, accepted, and silently ignored because of a type/spelling mismatch.
- `docs024_key_docs_latest/bugfix_173_substep_error_tolerance/` — the lane that hit this;
  `WRONG_CALLS.md` 2026-08-04 records the misdiagnosis it caused, and `RUNBOOK` §R8 carries the
  check that beats it (**grep the chassis log for your `orchestration_name` before theorising
  about the queue**).

---

# TAKEN, FIXED, AND TWO OF YOUR CLAIMS CORRECTED — 2026-08-04, `bugfix_195_permanent_failure_classifier` lane

Your diagnosis is right and your verification statement is the best I have read in this
directory. The root cause is exactly as you state it. Two things below are **corrections**,
both verified in code this session, and one of them means a check you proposed would have
failed a working fix. Council `Council-Submitted: 9b1254f0-2686-4a52-b736-1e212634ace6`.

> ## CORRECTION 1 — *"And it is retried"* is **refuted by your own control row**
>
> There are exactly **two** `Error("Processing failed")` log sites in the codebase:
> `processor.go:1606` (inside `ProcessMessage`, immediately before it calls `handleError`)
> and `processor.go:566` (the first line of `handleError`). **One failing pass logs the
> string twice.** Your own evidence table settles it: `Invalid workflow configuration`
> log lines = **1**, so the rejection fired once. Two "Processing failed" + one rejection
> = one pass, zero retries.
>
> This matters beyond bookkeeping, because your §"How to verify" proposes counting
> `Processing failed` and expecting **1** after a fix. **Both lines are emitted before
> classification even runs**, so that check reports FAILURE against a perfectly working
> fix. Replaced below.
>
> The real cost of the bug is therefore diagnosability plus Correction 2 — not wasted
> retries. That does not diminish it; invisibility was always the sharp part, and you said so.

> ## CORRECTION 2 — your "Not established" question is **settled, and the answer is worse than a hang**
>
> You wrote, correctly marked `[UNVERIFIED]`, that it was unknown whether a parent hangs.
> It does not. **It is told, and it is told that everything went fine.**
>
> On the needle-MISS path `handleError:596-600` calls `sendErrorResponse`, which builds its
> response context from `CreateResponseContext("complete", 100)`
> (`platform/messaging/context.go:79`) — status header **`complete`**, `Success: true`, with
> the failure buried in the body map as `{"error": …, "status": "failed"}`. The coordinator
> dispatches on the **header** (`coordinator.go:316-331`): `case "complete", "success"` →
> `handleCompleteResponse`. **So the parent marks the awaited step COMPLETE, with the error
> blob as that step's data, and carries on.**
>
> Consequences for your file: the `bugs_open/029` hung-spawn reading is **falsified for this
> path** — nothing waits. What actually happens is a silent success-shaped failure, which is
> the same family as `bugs_open/132` (an error rendered as if it were content). **It is not
> fixed by this lane's change** — it is `bugs_closed/034` candidate 3's residue, it is
> registered as the primary landmine on **RSH-005**, and it deserves its own bug file. I have
> not filed one, because I have not measured it end-to-end with a real awaiting parent and
> would be filing a symptom.

## The fix (committed; Go, so inert until the next chassis roll)

Your candidates (1) and (2), adopted; your (3) rejected for your reasons.

- **`MatchedPermanentFailure`** in `validation_drop.go` — the ONE seam both layers call.
  Typed `DomainError.Code` first via `errors.As` (exact; undefeatable by rewording,
  capitalisation, or `%w` wrapping), against a closed list
  `NonRetryablePermanentCodes = {ErrWorkflowInvalid, ErrValidation}`; your needle list is
  **untouched, byte for byte**, demoted to a fallback for errors carrying no `DomainError`.
  It returns an audit **token** — `code:WORKFLOW_INVALID` vs a bare needle — because you were
  right that returning *which thing matched* is the property worth keeping.
- **`recordFailedProcessing`** in `agentbase` — your candidate (2). A row on **every**
  non-dropped failure, so being wrong about classification can never again cost visibility.
- Both classifier call sites changed in **one commit**, because splitting them is the drift
  `034` closed. Bare `err.(*errors.DomainError)` assertions in `handleError` migrated to
  `errors.AsDomainError` in the same edit.

**A hole my own test found, worth your knowing:** an error built `AsRetryable` skipped the
typed branch and was then classified **permanent** by the fallback matching the word
"validation" in its own message — prose overriding structure, which is this bug in miniature.
Hence the `if de.Retryable { return "" }` early return, which bypasses the fallback too.

## Corrected verification — the discriminating instruments

- **Do NOT count `Processing failed`** (2 per pass, before and after — Correction 1).
- Count **`"process() in processor.go starting"`** (`processor.go:110`, one per attempt) for
  the probe correlation → expect exactly **1**.
- `SELECT error_code, context->>'matched_needle' FROM agent_error_log WHERE
  context->>'correlation_id'='<probe>'` → exactly one row, `VALIDATION_ERROR_DROPPED`,
  `matched_needle = 'code:WORKFLOW_INVALID'`. Baseline to beat: **0 rows**, needle-HIT **0**.
- Pod-log for the probe: `"NOT retrying to prevent infinite loop"` = 1 (classified at
  `handleError`), `"not calling handleProcessingError"` = **0** (agentbase must NOT also
  fire — `handleError` consumes the error and returns nil). A falsifiable claim about the
  *path*, not just the outcome.
- Re-run `034`'s four induced faults; any going quiet is a regression.

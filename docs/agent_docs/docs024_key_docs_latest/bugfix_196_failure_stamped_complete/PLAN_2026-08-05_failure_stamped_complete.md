# PLAN — bugfix 196: a failure response reaches the parent stamped `complete`

**Lane opened 2026-08-05** (session fc6ee578). Bug: `bugs_open/196_HANDOFF_2026-08-04_…`.
Filed OPEN/UNOWNED by the bugfix-195 lane; claimed by this session after checking
`who-owns.py`, the work-item queue (0 rows), and the live transcripts (the only
session touching the symbols was 195's own, which filed 196 and has since closed 195).

## Verification statement (owner ruling 2026-07-31)

No `090` run. Substituted equivalent first-hand verification, declared here:

- **Traced the full mechanism at HEAD end to end**, not just the four excerpts the
  bug file quotes: `handleError` (processor.go:565-610, both branches call
  `sendErrorResponse`), `sendErrorResponse` (1920-1934), `sendWorkflowFailureResponse`
  (547-563, a second complete-stamped error sender the bug file does not name),
  `sendWorkflowResponse` (640-741, hard-codes `Body.Success: true` at :709),
  `CreateResponseContext` (messaging/context.go:77-80 → types/context.go:459, which
  derives `IsComplete`/`IsError` correctly from ANY status — the seam already exists,
  only the messaging wrapper pins "complete"), wire propagation
  (`ToResponseHeaders` carries `Status` → coordinator.go:510 parses `headers["status"]`),
  and the coordinator switch (coordinator.go:316-331).
- **The bug file's prediction is CONFIRMED at the code level and sharpened** (see
  NOTES §mechanism): on the non-permanent branch the child sends TWO responses —
  the processor's complete-stamped blob first, then agentbase `handleProcessingError`'s
  correctly-stamped one (`getErrorStatus`, agent.go:1737-1742). Both key on
  correlation_id (same partition, ordered), so the bogus one always arrives first,
  claims the awaited request (`ClaimAwaitedRequest`), and the correct one is
  discarded as a duplicate. On the permanent/validation branch `handleError` returns
  nil, so the correct sender never fires at all: the blob is the only answer.
- **The end-to-end induction the bug file owed is scheduled as this lane's
  acceptance check** (post-fix, with the pre-fix behaviour as the baseline
  prediction; see §Acceptance below).

## Measurements (2026-08-05, all run against the live cluster)

- Production probe: 0 step values with the error-blob shape (direct or under the
  `.response` wrapper `applyResponseToState` writes) across ALL 4,403
  `orchestration_states` rows, any status. **Bounded claim**: terminal rows are
  reaped ~24h, and the `output_mapping` path (coordinator.go:2647) ERASES the blob
  (extracts only mapped fields), so this is a lower bound on incidence, not proof
  of absence. Historical evidence the dialect flows: the 2026-07-18 sweep behind
  `handlerReportedFailure` found 54 completed work items carrying
  `response.status='failed'`.
- Consumers census (the decisive one): fleet-wide, exactly TWO active workflow
  conditions reference status/error — `fix-proposer` (`diagnosis_row.status`, a DB
  column) and `content-reviewer` (evaluator output fields). **No live workflow
  branches on a child failure arriving as complete.**
- Go readers of the `status:"failed"` blob dialect (kept working, see Risks):
  `complete_work_item_verification.go:handlerReportedFailure` (017's compensating
  guard), `loop_actions.go:267,291`, `multipage_actions.go:1671`,
  `git_deployer_actions.go:574,589`, `fixloop_digest_action.go:343`.
- `errors.Retryable` defaults false; NOTHING in production code constructs
  `Retryable: true` today (only the builder + validation_drop.go honour it), so the
  typed mapping below sends `error_unrecoverable` for every failure that exists
  today. `handleRecoverableError` stays reachable for future typed retryables.
- Both coordinator error handlers are mature, maintained, and **already exercised
  in production by adapters** (which send proper error statuses; bugs 075/003 fixed
  real behaviour in them). The fix routes onto tested code, not dead code.

## The fix (bug file's candidate 1 — close the door at the sender)

All edits in `platform/messaging/processor.go` (+ its test file). No coordinator
change; no agentbase change; no change to either `CreateResponseContext` (other
callers keep their behaviour).

1. Extract the body of `sendWorkflowResponse` into an internal
   `sendWorkflowResponseWithStatus(ctx, msgCtx, result, status string, errInfo *types.ErrorInfo)`.
   After `responseCtx := msgCtx.CreateResponseContext()`, when `errInfo != nil`
   override `responseCtx.Status = status`, `IsComplete = false`, `IsError = true`
   (this also covers the legacy nil-ExecutionContext fallback), and build
   `Body: {Success: errInfo == nil, Body: result, Error: errInfo}`.
2. `sendWorkflowResponse(ctx, msgCtx, result)` becomes the thin success wrapper:
   `sendWorkflowResponseWithStatus(ctx, msgCtx, result, "complete", nil)`.
   Success path byte-identical in behaviour.
3. `sendErrorResponse`: typed status decision, mirroring 195's remedy —
   `status := "error_unrecoverable"; if errors.IsRetryable(err) { status = "error_recoverable" }`.
   NO prose/substring matching (that is bug 197's seam, deliberately not
   duplicated here). ErrorInfo carries `Code: errors.CodeOf(err)`, `Message`,
   `Recoverable`. The legacy body blob `{"error":…, "status":"failed"}` stays in
   `Body.Body` unchanged — every dialect reader above keeps working, and
   `handleUnrecoverableError` prefers `Body.Error` (better messages) with the
   body fallback intact.
4. `sendWorkflowFailureResponse`: same treatment (workflow-start failures).
5. Tests (processor package), following 195's convention that a test asserts the
   MISS: (a) `ReproducesTheBug_…` — build the pre-fix envelope shape and assert
   the coordinator switch would route it to `handleCompleteResponse` (documents
   the defect); (b) error envelope now carries `error_unrecoverable`,
   `Success=false`, `IsError=true`, populated `ErrorInfo`; (c) a
   `Retryable=true` DomainError maps to `error_recoverable`; (d) the success path
   still emits `complete`/`Success=true`; (e) body blob dialect unchanged.

## What changes for the fleet (the guarantee, stated for consumers)

Before: a chassis child's processing failure reached its parent as a COMPLETE
step whose data was the error blob; the workflow continued on junk. Loops,
`error_step` routing and `failWorkflow` never saw child failures from this path.

After: the parent's existing switch routes the failure — `continue_on_error`
loops skip-and-record, steps with `error_step` route to it, otherwise the
workflow fails with the child's real error message. This is the ALREADY-DOCUMENTED
intended semantics (`fail_workflow_action.go`'s header: "the caller's call_agent
error_step fires"); adapters have always behaved this way.

Named consumers, told not merely measured:
- `handlerReportedFailure` guard (017 lane): its motivating case (WORKFLOW_INVALID
  child) now fails the parent at response time instead of completing-with-evidence;
  the guard REMAINS for handler-authored verdicts and pre-roll rows. Noted in the
  bug file trail.
- Work-item pipeline (149 lane's territory): a handler saga that previously
  completed-on-junk now fails; item recovery for failed sagas is the SAME path
  adapter-failure sagas already take today. No new state introduced.
- `bugs_open/029` (hung spawns): 196's file already appended the settlement; the
  fix does not change the "parent is answered promptly" property — it changes only
  WHAT the answer says.

## Risks

- A workflow that "worked" only because it ignored a child failure will now fail
  visibly. That is the point, but it may surface latent failures fleet-wide as
  FAILED orchestrations. They were failing before; the record was lying.
- `error_recoverable` mapping is dormant today (nothing constructs Retryable=true);
  if 197's lane later drives typed retryables, coordinator retry caps (max 3,
  bugs_open/075's fixes) bound the blast.
- Rollback: single-commit revert of processor.go restores the old envelope.

## Acceptance (the bug file's own settlement measurement, run post-roll)

Dispatch a parent whose only step is `call_agent` at an agent guaranteed to fail
non-permanently; read the PARENT's row. Pre-fix prediction (from the code): parent
completes/advances with the blob as step data. Post-fix: parent routes to error
handling (FAILED with the child's message, or error_step). Falsifier named before
running: if the post-fix parent still completes with blob step data, the fix is
wrong and this plan is corrected in place.

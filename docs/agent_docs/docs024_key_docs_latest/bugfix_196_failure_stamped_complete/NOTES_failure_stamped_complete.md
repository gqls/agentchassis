# NOTES — bugfix 196 (append-only, newest at the bottom)

## 2026-08-05 — claim, validity check, measurements

- Picked 196 after sweeping `bugs_open/`: every other recent bug had fix commits
  from 08-04/05; 196/197 filed OPEN/UNOWNED by the 195 lane. Checked
  `who-owns.py 196` (only 195's filing commit), `site_work_items` (0 open rows
  matching), and live transcripts: session 424d6591 = the 195 lane itself, which
  filed 196 and just closed 195. Took 196 over 197 because 196 is medium-high with
  one owed measurement; 197 explicitly wants 195's `agent_error_log` population to
  accumulate first (cheaper after the 1252 roll beds in).
- **Validity: CONFIRMED at HEAD** (post-195 line numbers): handleError
  processor.go:565-610; sendErrorResponse 1920-1934; sendWorkflowFailureResponse
  547-563 (a SECOND complete-stamped error sender the bug file does not name);
  sendWorkflowResponse 640-741 with `Success: true` hard-coded at :709;
  coordinator switch 316-331; header propagation proven (`ToResponseHeaders`
  carries Status; coordinator.go:510 reads `headers["status"]`).
- **Sharpening beyond the filed mechanism**: non-permanent branch sends TWO
  responses — processor's complete-stamped blob, then agentbase
  handleProcessingError's CORRECT one (agent.go:1421-1502, getErrorStatus
  1737-1742). Same key (correlation_id) ⇒ same partition ⇒ ordered ⇒ the bogus
  one always claims the awaited request first; the correct one is
  DUPLICATE_SKIPPED (processResponseClaimWithRetry). So the coordinator's error
  handlers were UNREACHABLE from chassis-child failures on this path, while
  remaining exercised by adapters. Permanent branch: handleError returns nil, so
  the correct sender never fires — the blob is the only answer.
- `sendErrorResponseOLD` (processor.go:1937) is the pre-refactor implementation
  and it set `Status: "error_unrecoverable"`, `IsError: true`, `Success: false`
  correctly. The complete-stamping arrived with the `deaaa56b7` refactor
  ("refactoring syntactically correct - untested"). The defect is a regression,
  not a design.
- **Misstep (probe 1)**: first production probe used
  `collected_data::text LIKE '%"status": "failed"%'` → 86/3169 rows, which I
  nearly recorded as live incidence. Sampling showed every match was deep inside
  `agent_config`/`__raw_message__` (workflow definitions mention the string), not
  the error blob. Re-encoded the probe on the shapes `applyResponseToState`
  actually writes (direct + `.response`-wrapped): 0 rows. The first number
  answered "which rows contain these bytes anywhere", not "which steps recorded a
  failure blob" — the question I encoded, not the one I asked. Recorded in
  WRONG_CALLS.md.
- 0 is a LOWER BOUND: ~24h terminal-row retention, and the output_mapping path
  (coordinator.go:2647) erases the blob shape entirely (extracts mapped fields
  only — a failure there records an empty mapped result + metadata).
- Conditions census (jsonb_path_query over active agent_definitions): exactly 2
  conditions fleet-wide reference status/error; neither reads a child failure
  blob. No workflow depends on the broken dialect.
- Go-side dialect readers found and preserved (see PLAN): handlerReportedFailure
  (017's guard — its own doc comment names this exact envelope), loop_actions
  267/291, multipage 1671, git_deployer 574/589, fixloop_digest 343.
- `errors.Retryable`: default false (errors.go:138); no production constructor
  sets true. Typed mapping ⇒ all of today's failures go error_unrecoverable.

## 2026-08-05 — council APPROVED round 1; objection discharges

- Submitted corr `d1a63089-af5b-41a2-bea1-62259aa5db52` at ~11:46Z; verdict
  **APPROVED 11:57Z, round 1** — "approved with 1 advisory objection(s), none
  high-severity", 7 seats abstained (relevance-gated). No queue latency this time
  (the 29-minute figure in CLAUDE.md remains the budget, not the norm).
- Objection discharges, each with its evidence:
  1. *"Run the precedent check before treating the shared envelope constructor as
     routine"* (architecture-flavoured, medium) → seam registered as **CTS-058**
     in `contracts-and-standards.md` + index row, in the same commit as the fix
     (ordering-exemption condition 2; the 07-29 ruling retired condition 1).
  2. *"Census only covered conditions matching status/error"* (medium) →
     enumerated ALL distinct conditions fleet-wide (jsonb_path_query, no filter):
     ~96 distinct expressions read content fields, counts, or DB columns; the only
     envelope-adjacent readers (`page_content.response.skipped`,
     `site_plan.response.needs_*`) read content keys absent from a failure blob
     and key on nothing the fix changes. No condition reads Success/is_error/
     response_status.
  3. *"State the pod-grep explicitly"* → in the RUNBOOK: positive control (a
     string the change ADDS) + negative control (a string it REMOVES, expect 0),
     same exec, every replica, per the 153 lesson.
  4. *"Retryable:true absence claim lacked a citation"* (low) → the cited command:
     `grep -rn "AsRetryable\|Retryable: *true\|Retryable = true" --include="*.go"
     platform/ internal/ pkg/ cmd/ | grep -v _test` → only
     `platform/errors/errors.go:171-172` (the builder itself). Caveat honoured:
     a grep proves absence only for the spellings searched; the field is only
     settable via the builder or a literal, both covered.

## 2026-08-05 (afternoon) — implementation, mutation proof, commit

- Implementation delegated to an Opus agent; it died TWICE on infra (a 529
  overload, then the account session limit, resets 17:50 London) — but its edits
  and tests were complete and on disk before the second death. I verified every
  edit against the approved plan myself, then ran the checks it could not finish.
- All green, uncached: gofmt clean, `go build ./platform/...` OK, `go vet` OK,
  `go test ./platform/messaging/ -count=1` OK. Six new tests (11 cases including
  subtests), including `ReproducesTheBug_completeStampedFailureRoutesToCompleteArm`.
- **Mutation check run by me** (the step the agent was on when it died): new test
  file copied into a clean `git archive HEAD` extract (HEAD = un-fixed) →
  **fails on exactly the defect assertions** (`Status = "complete", want
  "error_unrecoverable"`, `Body.Success = true on a failure response`,
  `Body.Error is nil`). The guard can fail, so it guards.
- **Committed `d16e6d23c`** — processor.go + processor_response_status_test.go +
  CTS-058 register entry, trailer `Council-Reviewed: d1a63089-…`. Pattern-check
  twins both expected and deliberate: `sendErrorResponseOLD` is the DEAD
  pre-regression copy (left untouched, narrow commit); `sendWorkflowSuccessResponse`
  routes through the fixed wrapper.
- **Same-file passenger, benign, noting per the rule:** my CTS-058 row appended to
  `000_concept_index.md` was swept into the 198 lane's commit `c48c773c1`
  (fix(198) css-patch-agent) before I committed. Nothing lost; the row is at HEAD;
  it simply arrived under their message rather than mine.
- Misstep (small, caught immediately): my "do definitions carry an initial step?"
  query tested `workflow ? 'initial_step'` → 0/175, but the real key is
  **`start_step`** (`convertToWorkflowPlan`, processor.go:417). A zero from a
  wrong-spelling probe reads exactly like a real absence — same family as the
  jsonb probe misstep above; caught because the probe seeds needed the key.
- **The chassis build the owner rolled this afternoon PREDATES `d16e6d23c`** — the
  fix is committed but NOT live. Next build picks it up (build-from-HEAD).

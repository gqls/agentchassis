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

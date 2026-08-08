# PLAN 2026-08-08 — bugs_open/217: converge `notifyParentOfFailure` through `RetryDisposition`

Continuation of the sender-convergence lane (207 → 216 → **217**, the third and last
unconverged chassis failure sender). Decision: **converge** (bug file option 1), not
decline — the hardcoded terminal is an accident of a literal, the parent's recoverable
arm is now real (216 proven live on v1.0.1266), and the population measurement below
shows a majority of these failures are transient-shaped.

## Validity re-check (2026-08-08, this session)

- `coordinator.go:3924` still `Status: "error_unrecoverable"` hardcoded; `:3937` still
  `Recoverable: false`. Bug is live at HEAD.
- Ownership: `who-owns 217` names only this lane (207/216 dirs, both finished);
  all four recent transcripts mentioning the symbol are wrapped-up sessions; the
  `site_work_items` queue has no open item touching it. Taken by this session.

## The structural finding the bug file asked for: the import direction is a CYCLE

`platform/messaging` imports `platform/orchestration` (processor.go:21,
validation_drop.go:27 — the drop recorder calls `orchestration.LogAgentError`). So the
coordinator **cannot** import `messaging.RetryDisposition`; the bug file's "needs
checking for cycles" resolves to: yes, a cycle, direct convergence is impossible.

Routes considered:

1. **Extract the classification core to `platform/errors` (chosen).** A true leaf
   (stdlib imports only), already home to `DomainError`, `ErrorCode` and the typed
   constants the classifiers read. `platform/messaging` keeps thin re-exports so every
   existing caller and every pinning test compiles and passes unchanged, and there is
   still exactly ONE implementation. The RSH-007 header's argument ("one function
   decides retry-vs-terminal for every agent in the fleet, which is why it lives at the
   shared seam") is *completed*, not reversed: a second layer now needs the same
   judgement and the seam must therefore live where both layers can reach it.
2. **Hand-copy the two-stage order into the coordinator — rejected.** Two lists for one
   judgement is the drift disease 034 → 195 → 197 existed to end; `bugs_open/224` (seven
   private annuity formulas) is this week's demonstration of where copies go.
3. **Inject a classifier func into `SagaCoordinator` at wiring time — rejected.** A nil
   field silently reverts to hardcoded-terminal; an unset default is a comment, not a
   control, on a tree this many sessions share (owner ruling 2026-08-02 §2 is about
   exactly this shape).
4. **Move the classifiers into `platform/orchestration` — rejected.** Messaging would
   re-export from the orchestration layer; the dependency arrow would point from the
   generic classification seam INTO the coordinator package, which is upside-down and
   invites the next cycle.

## Measurements (2026-08-08, live DB; queries in NOTES)

`severity='fatal'` in `agent_error_log` is written **only** by `notifyParentOfFailure`
(coordinator.go:3917, sole writer), and only after the parent-exists check — so the
fatal rows are exactly this sender's population:

- **11,970** parent notifications in 14 days.
- Replaying `RetryDisposition`'s needle logic in SQL (permanent needles case-sensitive,
  transient folded, permanent first — i.e. what the fix will do to these strings):
  **6,239 (52%) flip to `error_recoverable`** · 4,756 match a permanent needle (stay
  terminal) · 975 unclassified (stay terminal).
- The flip decomposes: 4,163 "connection" (dominant shape: browser-runner
  `ERR_TUNNEL_CONNECTION_FAILED` — a proxy fault, plausibly cured by retry), 1,989
  "deadline exceeded" (firecrawl POST timeouts — the textbook transient), 87 "timeout".
- **Nesting depth of failure chains: 0 (8,557 rows) and 1 (3,417 rows); no depth ≥ 2 in
  14 days.** This bounds the amplification risk below.
- `TimeoutMonitor.sendTimeoutResponse` (helpers.go:409), the bug file's `[UNMEASURED]`
  sibling suspect: **dead code** — `NewTimeoutMonitor`, `MonitorChildOrchestration`,
  `MonitorRequest` and the literal `TimeoutMonitor{` have zero call sites outside
  helpers.go and tests. No convergence needed there; recorded in the bug file.

## The edits (council plan, ≤8)

1. **NEW `platform/errors/permanent_failure.go`** — move from
   `messaging/validation_drop.go`: `ValidationErrorNeedles`,
   `NonRetryablePermanentCodes`, `MatchedPermanentFailure`, `MatchedValidationNeedle`.
   Headers travel with the code (the census and rejected-needle arguments are the
   documentation of record).
2. **NEW `platform/errors/transient_failure.go`** — move from
   `messaging/retryable_transient.go`: `RetryableTransientCodes`,
   `TransientErrorNeedles` (exported now; messaging keeps an unexported alias),
   `MatchedTransientFailure`, `matchedTransientNeedle`, `RetryDisposition`. Headers travel.
3. **EDIT `platform/messaging/validation_drop.go`** — delete moved declarations; add
   aliases/wrappers (`var ValidationErrorNeedles = errors.ValidationErrorNeedles`, etc.);
   `recordDroppedValidationError` and its orchestration dependency stay.
4. **EDIT `platform/messaging/retryable_transient.go`** — same: aliases
   (`var transientErrorNeedles = errors.TransientErrorNeedles`) + wrappers for
   `MatchedTransientFailure` / `RetryDisposition`.
5. **EDIT `platform/orchestration/coordinator.go` `notifyParentOfFailure`** — classify
   before stamping: `recoverable, matched := perrors.RetryDisposition(errors.New(errorMsg))`
   (`perrors` alias per agentbase convention); status derived in lockstep with
   `Recoverable`; `Code: "CHILD_ORCHESTRATION_FAILED"` and the message KEPT (its only
   consumers key on the code, and the 090-recovery landmine documents it); an Info log
   with a long greppable literal + the matched token, mirroring the processor senders.
6. **NEW `platform/orchestration/notify_parent_disposition_test.go`** — reuses the 216
   harness (`recordingProducer`, `unreachableDB`, same package). Pins: transient prose →
   `error_recoverable`/`Recoverable:true`; **sequencing pin** — prose carrying BOTH
   needles (`"pq: invalid connection"`) stays `error_unrecoverable`; unclassifiable prose
   stays terminal; code preserved; no-parent early return still produces nothing.

Docs in the same commit (not code edits): `bugs_open/217` updated (decision, measurements,
TimeoutMonitor verdict, close criteria), RSH-007 register entry corrected (relocated home,
messaging re-exports — same-commit registration per the platform-seams ruling),
`platform/errors/errors.go:320` stale prose pointer, this PLAN + NOTES.

## What changes on the wire, and for whom (consumers NAMED, not just measured)

- **Parent coordinators (same binary, fleet-wide):** a transient-shaped child-orchestration
  failure now arrives `error_recoverable` → `handleRecoverableError` → replay via the
  claimed row (216's proven seam), capped at `RetryVersion >= 3`. Terminal-by-accident
  becomes retry-then-terminal.
- **`platform/messaging` / `platform/agentbase`:** no behavioural change — re-exports;
  agentbase deliberately untouched (its source-scan test pins `messaging.Matched*` call
  sites; the call-order guard stays as is).
- **Readers of `CHILD_ORCHESTRATION_FAILED`:** unchanged — the code survives; only
  status/recoverable are computed now.
- **Adapter services** (thunder/analyser/browserrunner): out of scope, own senders, named
  in the bug file's census.

## Risks, stated

- **Amplification across nesting levels.** Today the hardcoded terminal cuts every retry
  cascade at the first coordinator boundary; after convergence each level retries its
  child (cap 3+1 per level). Measured chain depth is ≤ 1, so the realized worst case is
  (1+3)² = **16 innermost executions** for a persistently-transient failure, each spaced
  by real execution time (a deadline-exceeded child burns its deadline before failing).
  Monitors: the RSH-006 storm-watch (retry_version histogram, wall at 3) plus the
  `severity='fatal'` row rate. If amplification shows up in practice, the named
  follow-up is a terminal-exhaustion marker minted at the cap site — NOT shipped here,
  because widening the shared classifier's stages inside a bug patch is the seam-in-a-
  bug-patch shape the guardian rightly vetoed on 124.
- **The two success-path call sites** (result undeliverable at 3692/3752): a transient
  Kafka produce failure now replays the WHOLE child (repeat work, bounded at 3) — the
  result finally arrives instead of the parent failing terminally. The deterministic
  refusals stay terminal: the size-cap message carries no transient needle, and
  "message validation failed" matches the permanent needle `validation`.
- **Over-match cost** ("connection" catching TLS prose, "dial tcp" catching NXDOMAIN) is
  the classifier's own documented, accepted trade — bounded at 3 futile retries, and it
  now applies to this sender exactly as it does to the processor's.

## Verification

1. `go build ./... && go test ./platform/messaging/ ./platform/agentbase/ ./platform/orchestration/ ./platform/errors/` green against `git archive HEAD` (shared-tree rule).
2. Mutation checks: (a) revert edit 5's classification to the hardcoded literal → the
   transient-prose pin fails; (b) swap `RetryDisposition` for `MatchedTransientFailure`
   in edit 5 → the both-needles sequencing pin fails.
3. Council gate before/alongside commit; `Council-Submitted:` trailer if the verdict has
   not landed.
4. Post-roll close criteria (verify at the artefact, never the tag):
   - pod-grep every replica for the new sender log literal + a negative control
     (`scripts/pick-pod-marker.py`).
   - Induction (SEED_test_207_probe recipe): a deadline-exceeded child-orchestration
     failure must draw `error_recoverable` from THIS sender (the 16:07:45-shaped
     envelope), and the parent's `awaited_requests.retry_version >= 1`.
   - Storm watch: retry_version histogram — mass at 0–1, hard wall at 3, zero above;
     plus week-over-week `severity='fatal'` row rate for amplification.

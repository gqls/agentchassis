# Plan — continue from the 195 handoff: take `bugs_open/197` (the retryable-side substring classifier)

## Context

The 195 handoff is SUPERSEDED — that lane is closed, live, and council-approved. Its two spawned follow-ons were the continuation:

- **`bugs_open/196`** is **already OWNED and FIXED** by another lane (claimed 2026-08-05, session fc6ee578; fix `d16e6d23c` committed with 430 lines of tests; awaiting their roll verification). Taking it would be competing. **Off the table.**
- **`bugs_open/197`** is **OPEN, UNOWNED, untouched** since it was filed. This plan takes it.

**The bug:** `isRecoverableError` (`platform/agentbase/agent.go:1730-1735`) decides retry-vs-terminal for every agent's processing failure by **case-sensitive substring match over error prose** — three needles: `timeout`, `connection`, `temporary`, no nil guard. Its verdict becomes the wire status (`error_recoverable`/`error_unrecoverable`), which the coordinator routes to retry or terminal. A dead twin, `errors.IsRecoverable` (`platform/errors/errors.go:316-347`, 12 case-folded patterns, **zero callers**), disagrees with it on case — the two-lists drift that `bugs_closed/034` closed on the permanent side.

**The census (run during planning; population = `agent_error_log` rows with `error_code='PROCESSING_FAILED'` or `classification='transient'` — exactly the errors that reach this seam, courtesy of 195's unconditional recorder):**

| measure | value |
|---|---|
| population | **2,996** |
| current verdicts | 570 recoverable / 2,426 unrecoverable |
| contain `deadline exceeded`, classified **terminal** | **885** (~30% — transient by definition; work lost that a retry would save) |
| capitalised `Timeout`/`Connection`/`Temporary`, missed on case alone | **882** (195's capital-letter defect, live on the sibling) |
| rate-limit/5xx family | 37, of which 20 missed |
| suspicious "recoverable" verdicts | SSL/TLS **certificate** errors matched via `connection` — likely permanent, retried futilely |

**One scope caveat, verified in code:** post-196, the processor's `sendErrorResponse` (`processor.go:1959-1992`) classifies **typed-only** and its comment explicitly defers the substring question to 197. On the orchestrated main path the processor's response is sent first and wins the coordinator's `ClaimAwaitedRequest` race, so the agentbase verdict is live-deciding only where the processor's sender does not fire (no ResponsesTopic; errors bypassing `handleError`; spawned-child paths). The live flip is therefore smaller than the census's raw 885 — and the retry-storm risk correspondingly smaller. The full convergence of the processor's senders belongs to the **active 196 lane** and is scheduled/told, not done here as a passenger edit.

## The fix

Mirror 195 exactly: typed-first, substring fallback, audit token, one shared seam.

### New shared classifier — CREATE `platform/messaging/retryable_transient.go`

Beside its permanent twin (`validation_drop.go#MatchedPermanentFailure`). Same package so the existing AST lockstep guard extends for free; `platform/errors` is the wrong altitude for an operational pattern list (the dead `IsRecoverable` rotting there is the exhibit).

```go
func MatchedTransientFailure(err error) string {
    if err == nil { return "" }
    if de, ok := errors.AsDomainError(err); ok {
        if de.Retryable { return "retryable:" + string(de.Code) }  // author intent outranks everything
        for _, code := range RetryableTransientCodes {             // {ErrAgentTimeout, ErrAgentOverloaded, ErrRateLimited}
            if de.Code == code { return "code:" + string(code) }
        }
        // deliberate fall-through: Retryable=false must NOT force terminal —
        // InternalError() defaults Retryable=false (mirrors the permanent twin's argument)
    }
    return matchedTransientNeedle(err.Error())  // ToLower once, lowercase needles, first match wins
}
```

**Fallback needles, case-folded, each admitted against "can retrying this succeed?"** (census as arbiter): `timeout`, `deadline exceeded` (the 885-row headline fix), `connection` (kept — live behaviour; SSL-cert over-match documented, cost bounded at 3 retries), `temporary`, `dial tcp`, `too many requests`, `rate limit`, `service unavailable`, `bad gateway`, plus `unreachable` **or** `network` (pick one from the missed rows' actual prose during implementation).

**REJECTED: bare `502`/`503`/`504`** from the dead twin's list — the dominant message shape nests correlation UUIDs, which are hex, so `503` matches inside an arbitrary ID. Pinned OUT by a test.

**Why case-folding is right here but stayed forbidden on the permanent side:** asymmetric costs. Permanent over-match = retryable work dropped forever (unbounded). Transient over-match = at most 3 futile redispatches then terminal anyway (`coordinator.go:2978` `retry_version >= 3`, plus the independent adapter cap from bug 075). Folding buys the 882 case-missed rows for a bounded downside.

### The edits (≤8, council schema)

1. **CREATE `platform/messaging/retryable_transient.go`** — classifier as above; header comment carries the census figures, per-needle justifications, the rejected digits and why, and the fold-asymmetry argument.
2. **CREATE `platform/messaging/retryable_transient_test.go`** — census-derived table of REAL messages verbatim (nested `workflow failed: Request <uuid>… deadline exceeded` → its needle; capitalised rows → pins the fold; `WORKFLOW_INVALID…` → `""`; a message whose only `503` sits in a hex UUID → `""`); typed branch (`%w`-wrapped `AsRetryable` → `retryable:` token; `ErrAgentTimeout` → `code:`; `InternalError`-wrapped deadline → falls through to needle); nil → `""`. Every needle mutation-proven both directions (remove it → a case fails; widen to rejected digits → a case fails).
3. **EDIT `platform/agentbase/agent.go`** (one function family, three hunks):
   - `handleProcessingError:1433` → `transientMatch := messaging.MatchedTransientFailure(err); recoverable := transientMatch != ""`, plus a log line with a LONG literal (`"retry disposition decided by shared transient classifier"`) for the post-roll pod-grep.
   - `ErrorInfo.Code:1493` → `perrors.CodeOf(err)` with `"PROCESSING_ERROR"` fallback, mirroring `:1354` — today the DB row and the wire disagree about the same failure's code. (Re-grep `.Error.Code` consumers before committing; record in appendix.)
   - `recordFailedProcessing` context gains `retry_disposition` + `transient_matched` keys — makes the after-census a one-query read on a *different column* from what the classifier consumes (not self-confirming).
   - DELETE `isRecoverableError:1730-1735`.
4. **EDIT `platform/errors/errors.go`** — DELETE the dead `IsRecoverable:316-347` (zero callers, verified; its patterns are subsumed by the judged admissions). No shim.
5. **EDIT `platform/agentbase/agent_test.go`** — extend the post-f887ed1ad AST guard: agent.go must call `messaging.MatchedTransientFailure` and must contain no `isRecoverableError` declaration/call.
6. **EDIT `platform/messaging/processor.go` (comments only, `:559-561`, `:1967-1969`)** — the "197's seam, deliberately not duplicated" comments now name `MatchedTransientFailure` as existing and state the senders stay typed-only by the 196 lane's standing decision, convergence scheduled. No behaviour change; 196's pinned tests stay green.
7. **Register `RSH-006`** (`docs/agent_docs/docs026_concept_register/register/resilience-self-heal.md` + index row + headline recount) — same commit, per the ordering exemption. Landmines: bare-digit needles rejected (UUIDs are hex); the processor sender claims first (`DUPLICATE_SKIPPED` — don't assert this seam's verdict from `orchestration_states` without checking which response claimed); `dial tcp` admits NXDOMAIN typos by design; the `retry_version >= 3` cap is now load-bearing for a wider population.
8. **APPEND to `bugs_open/197`** — the census results (its own "owed first" step, now done), the "decides fleet-wide" sharpening (processor pre-emption), each admitted/rejected needle, verification plan. **Stays in `bugs_open/` until fixed AND live** — the move to `bugs_closed/` is a post-roll follow-up.

## Process (the standing workflow)

1. Fable plan — done (this document).
2. Implement the 8 edits; `gofmt -w` only touched files; `go build ./platform/...`; full `go test ./platform/messaging/ ./platform/agentbase/ -vet=off`.
3. **Mutation-prove** with the fail-closed recipe (backup via `mktemp` + `test -s` before mutating — the corrected 192-RUNBOOK recipe).
4. **Council submit BEFORE commit** (`097_TRIGGER…` from repo root — the path is relative; a stale `cd` costs a run), then one pathspec commit with `Council-Submitted: <corr>`. Include the §2-ruling argument: the seam already exists; the current behaviour (transient failures made terminal, capitalised needles missed) *is* the unsafe default; 195's identical argument was approved; the flip is live but bounded (3-retry cap + adapter cap + processor pre-emption).
5. Working docs: create `docs/agent_docs/docs024_key_docs_latest/bugfix_197_transient_classifier/` with PLAN/NOTES/README as work proceeds; missteps → `WRONG_CALLS.md`.
6. Act on the verdict (REVISE loops are normal — both prior bugs took one round of real objections).
7. Fix is Go-only → **inert until the owner rolls a chassis image**. After the roll: verification below, then close (move to `bugs_closed/`, closing banner with evidence, 016b §10 row, `MEMORY_closed.md`).

## Verification (each check can come out FALSE)

- **Unit + mutation** (pre-commit): the named mutations each flip a named test — revert `:1433` to a local substring → AST guard fails; delete `deadline exceeded` → census-row test fails; remove `ToLower` → capitalised cases fail; re-add `503` → hex-UUID case fails; drop the `de.Retryable` early return → `retryable:` token test fails.
- **Post-roll pod-grep** (both replicas, LONG literal + positive control): `grep -ac "retry disposition decided by shared transient classifier"` → 1; a pre-existing long literal → 1.
- **Post-roll induction** (≥300s after pod start): (a) recoverable-shaped probe (step with a 1-2s timeout against a real call → deadline-exceeded prose) → `agent_error_log.context->>'retry_disposition'='error_recoverable'` + named needle; `awaited_requests.retry_version ≥ 1`; after 3 failures → `Max retries exceeded` + terminal (proves the bound, not assumes it). Also pod-grep which response claimed (`RESPONSE_MATCHED` vs `DUPLICATE_SKIPPED`) — this can falsify the pre-emption caveat in either direction. (b) unrecoverable-shaped probe (`"unclassifiable failure xyzzy"`) → `retry_version` stays 0, one attempt, terminal — if this retries, the fallback over-widened. Baselines: pre-fix, disposition key absent and `retry_version = 0` on the same shapes. Clean up probe agents after.
- **Census re-run, before/after**: after-census reads `retry_disposition`/`transient_matched` (columns the code writes, not what the classifier reads). Watch for a storm: `SELECT retry_version, count(*) FROM awaited_requests WHERE …` — healthy = mass at 0-1, hard wall at 3; anything at 4+ falsifies the bound.

## Who must be TOLD (not merely measured)

- **196 lane** (ACTIVE — same file family; convergence of their senders onto the shared classifier is *their* call; also flag the same-file-passenger hazard on `processor.go`).
- **Coordinator/retry-policy owners** — the `error_recoverable` arm fires more; the `:2978` cap is now load-bearing for a wider population; no backoff exists on that arm (named as a known gap, not fixed here).
- **173 lane** (filed 195/197's parent) — failing substeps change from terminal-on-first to up-to-3.
- Adjacent fact for the appendix: the spawned-child parent-propagation block (`agent.go:1508-1553`) hardcodes `"status":"failed"`, so parents always take the unrecoverable arm for child failures — 196/029 family, recorded not changed.

## Critical files

- `platform/messaging/retryable_transient.go` + `_test.go` (new)
- `platform/agentbase/agent.go` (`handleProcessingError`, `recordFailedProcessing`, delete `isRecoverableError`)
- `platform/errors/errors.go` (delete dead `IsRecoverable`)
- `platform/agentbase/agent_test.go` (AST guard)
- `platform/messaging/processor.go` (comments only)
- `platform/messaging/validation_drop.go` (the twin to mirror — read, not edited)
- `bugs_open/197_…md`, register `resilience-self-heal.md` + `000_concept_index.md`

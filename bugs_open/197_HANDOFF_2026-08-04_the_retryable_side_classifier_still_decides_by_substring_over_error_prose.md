# 197 — the RETRYABLE-side classifier still decides fleet-wide disposition by substring over error prose

**Filed 2026-08-04** by the `bugfix_195_permanent_failure_classifier` lane, at the direction
of the council's `bug_historian` seat (correlation `9b1254f0-…`, verdict REVISE):
*"The classifier being fixed here (permanent-vs-transient) and the classifier being left
alone (retryable-vs-not) are the same failure family — substring matching over error prose
deciding fleet-wide disposition — so the next silent-misclassification bug is already named
in this plan's own risk section but not scheduled."*

The seat is right, and this is the scheduling. **Status: OPEN, UNOWNED.**
**Severity: medium — no known live instance yet, which is exactly the point.**

> **VERIFICATION STATEMENT (owner ruling 2026-07-31).** No `090` run and **no asserted root
> cause of any live failure**. This is a **structural claim about a mechanism**, not a
> diagnosis of a symptom: the same technique that provably failed in `bugs_open/195` is still
> in use on the sibling path. What I read: `platform/agentbase/agent.go#isRecoverableError`
> and `platform/errors/errors.go#IsRecoverable`. **What I have NOT done** is find a live
> error string that these misclassify — see "What is owed before fixing".

## Why this is filed with no failing instance

`bugs_open/195` was this exact defect on the permanent-vs-transient side, and it went
undetected for as long as it did because nobody had fed the classifier its commonest real
input. It cost the fleet's most frequent permanent configuration error total invisibility,
on capitalisation alone (`"invalid"` vs `"Invalid"`).

The remedy there was structural: match the **typed `DomainError.Code`**, which is exact and
survives rewording, capitalisation and `%w` wrapping, and demote prose to a fallback for
untyped errors. That remedy was applied to **one** of the two classifiers.

Filing without a live instance is deliberate: waiting for one is the failure mode. The
sibling has the same three vulnerabilities, all demonstrated on its twin —
**rewording**, **capitalisation**, and **over-matching** on substrings that appear inside
unrelated runtime errors.

## What is owed before fixing (do this first — it decides severity)

The check that would have caught `195` in one line, run against these two functions:

1. **Enumerate the real inputs.** Collect distinct error strings that actually reach these
   classifiers — `agent_error_log.error_message` is now a usable population precisely because
   `195` made the record unconditional (`recordFailedProcessing`, live from the next roll), so
   this is cheaper after that ships than it was before.
2. **Feed each to the classifier and read the disposition**, then ask of every disagreement:
   is retrying this genuinely capable of succeeding? A "recoverable" verdict on a static
   configuration fault is an infinite loop; an "unrecoverable" verdict on a transient network
   blip is lost work.
3. **Write the misses as tests asserting the MISS**, the way `195`'s
   `ReproducesTheBug_needleMissesButCodeMatches` does — so a later tidy-up of the substring
   list cannot silently remove the reason the seam exists.

**Do not skip step 1 and go straight to the fix.** `195`'s own remedy is only safe because
its blast radius was measured (`ErrWorkflowInvalid`: one construction site; `ErrValidation`:
one, and it sends rather than returns). The equivalent census here is unrun, and a
retryability change made blind is a fleet-wide behaviour change.

## Fix shape, when it is measured

Mirror `195`: a typed decision first (`errors.AsDomainError` + `DomainError.Retryable`, which
already exists and is already honoured by `MatchedPermanentFailure`'s early return), with the
substring list demoted to a fallback for errors carrying no type. Reuse
`errors.AsDomainError`/`errors.CodeOf` — they were added by `195` and are chain-safe.

**Do not case-fold the substring lists** as a shortcut: on the sibling seam that widens every
over-match hazard to its capitalised variants, exactly as it would have on `195`'s.

## Related

- `bugs_open/195` — the same defect, the same family, on the permanent-vs-transient side;
  fixed, council-approved-pending, registered **RSH-005**. Its `016b` §9 entry
  ("a guarantee conditional on a classifier inherits its gaps") is the transferable write-up.
- `bugs_closed/034` — unified the two substring lists and deferred the typed fix in its own
  header; `195` was that deferral coming due on one side, and this is the other side.
- `bugs_open/196` — a third finding from the same investigation, unrelated mechanism.

---

# TAKEN AND FIXED IN CODE 2026-08-06 — `bugfix_197_transient_classifier` lane. Your "owed first" census is DONE, and it upgrades your severity from "no known live instance" to MEASURED.

**Status: fixed in code, OPEN until the next chassis roll** (Go-only, inert until then).
Commit carries `Council-Submitted:`; registered **RSH-006** in the same commit.

## The census you required before any fix (your step 1/2, run 2026-08-05/06)

Population: `agent_error_log` rows with `error_code='PROCESSING_FAILED'` or
`classification='transient'` — **exactly the errors that reached this seam**, which exists as
a queryable population only because `bugs_closed/195` made the failure record unconditional,
precisely as your file predicted it would be "cheaper after that ships".

| measure | value |
|---|---|
| population | **2,996** |
| old classifier's verdicts | 570 recoverable / 2,426 unrecoverable |
| contain `deadline exceeded`, classified **TERMINAL** | **885 (~30%)** — transient by definition; work lost that a retry would save |
| capitalised `Timeout`/`Connection`/`Temporary`, missed on case alone | **882** — 195's capital-letter defect, live on this sibling |
| rate-limit/5xx prose | 37 rows, 20 missed |
| suspicious "recoverable" verdicts | SSL/TLS **certificate** errors matching `connection` inside "establish a secure connection" — retried without hope |

## One sharpening of your framing, verified in code

"Deciding fleet-wide disposition" needs a caveat: post-196, the **processor's**
`sendErrorResponse` classifies typed-only, is sent FIRST on the orchestrated path, and wins
the coordinator's `ClaimAwaitedRequest` race — the agentbase verdict decides only where that
sender does not fire (no ResponsesTopic; errors bypassing `handleError`; spawned-child
paths). So the live flip is smaller than the raw 885, and the retry-storm risk smaller with
it. Converging the processor's two senders onto the shared classifier is the **196 lane's**
guarantee change to accept — scheduled in RSH-006 and told in their notes, not done here as
a passenger edit.

## The fix (mirrors 195, as you specified)

- **`messaging.MatchedTransientFailure`** (`platform/messaging/retryable_transient.go`,
  beside its permanent twin): author's `AsRetryable` outranks everything
  (`retryable:<code>`); typed list `{AGENT_TIMEOUT, AGENT_OVERLOADED, RATE_LIMITED}`
  (`code:<code>`); case-folded needle fallback (bare needle token). `Retryable=false` with
  an unlisted code **falls through** to the needles — `InternalError()` defaults
  `Retryable=false` while wrapping transient causes, and classifying on the flag would
  re-create this bug one level up.
- **Needles, each argued in the file header with the census as arbiter:**
  `deadline exceeded` (the 885-row headline), `timeout`, `connection` (SSL over-match
  documented and pinned visible by a test), `temporary`, `dial tcp` (NXDOMAIN over-match
  accepted, bounded), `too many requests`, `rate limit`, `service unavailable`,
  `bad gateway`. **REJECTED:** bare `502/503/504` (the dominant message shape nests hex
  UUIDs — pinned OUT by a test whose message's only `503` sits inside one), and
  `unreachable`/`network` (**zero** census rows — a needle with no observed input is policy
  nobody asked for).
- **Your case-folding warning, honoured with an argument rather than ignored:** folding
  stays forbidden on the permanent side (over-match = unbounded loss) and is right here
  (over-match = at most 3 futile redispatches then terminal — `retry_version >= 3` plus the
  bug-075 adapter cap). Asymmetric costs, not a shortcut.
- **Both old classifiers deleted:** agentbase's `isRecoverableError` (its caller now goes
  through the shared seam, pinned by an AST guard that fails if it is re-declared OR
  re-called) and the dead `errors.IsRecoverable` (zero callers, disagreed on case).
- **Your step 3, done:** the misses are tests asserting the miss — the nested
  census-message case fails if `deadline exceeded` is removed; the capitalised cases fail
  if the fold is removed; five mutations each flip named tests.
- Extra: `ErrorInfo.Code` on the wire now uses the typed code (`wireErrorCode`), ending the
  DB-row/wire disagreement about the same failure; `recordFailedProcessing` writes
  `retry_disposition` + `transient_matched`, so the after-census is one query on a column
  the classifier does not consume.

## What closes this file

Post-roll: pod-grep both replicas for the LONG literal
`retry disposition decided by shared transient classifier` (+ positive control); induce a
deadline-exceeded-shaped probe → `retry_disposition='error_recoverable'`,
`awaited_requests.retry_version ≥ 1`, terminal at `Max retries exceeded` after 3 (proving
the bound); induce `unclassifiable failure xyzzy` → `retry_version=0`, one attempt. Storm
watch: retry_version histogram — mass at 0-1, hard wall at 3.

Adjacent fact, recorded not changed: the spawned-child parent-propagation block
(`agent.go` post-`sendErrorResponse`) hardcodes `"status":"failed"`, so parents take the
unrecoverable arm for child failures regardless of the recoverable flag beside it —
196/029 family.

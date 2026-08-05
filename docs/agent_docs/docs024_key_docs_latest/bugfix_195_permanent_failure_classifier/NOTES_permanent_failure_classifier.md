# NOTES — `bugs_open/195`, the prose-guessing classifier

Append-only, newest at the bottom.

---

## 2026-08-04 — picking it, and why this one

Re-surveyed `bugs_open` (52 open) against 41 live session transcripts. Three bugs were new
since my last sweep: **193** and **194** sit with the lanes that spawned them (`173` and
`087` — 30 and 50 transcript mentions respectively). **195** appeared in exactly ONE
session: the `173` lane that filed it, whose own file says **"OPEN, UNOWNED"**. Took it.

Its verification statement is the best I have read in `bugs_open/`: fault induced twice
(once by accident, once deliberately), the whole path read in source rather than grepped,
and **every negative claim paired with a positive control** — the "0 rows in
`agent_error_log`" finding sits beside "1,301 rows in the same table over 24h", so a row
was demonstrably reachable. That is the shape that makes a filing trustworthy.

## 2026-08-04 — validity re-checked before planning

Needles unchanged (`"is required","validation","invalid","missing"`, case-sensitive);
`processor.go:280` still builds a capital-I `"Invalid workflow configuration"`; both
classifier call sites unchanged. Still valid, verbatim.

## 2026-08-04 — two corrections to the filing, both verified in code

**1. "And it is retried" is REFUTED by the filer's own control.** There are exactly two
`Error("Processing failed")` sites — `processor.go:1606` (in `ProcessMessage`, before it
calls `handleError`) and `processor.go:566` (first line of `handleError`). One failing pass
logs the string twice. Their own row `Invalid workflow configuration = 1` proves a single
pass. **This is not pedantry**: their §"How to verify" proposes counting that string and
expecting 1, and both lines are emitted *before classification runs* — so their check
reports FAILURE against a perfectly working fix. Replaced with
`"process() in processor.go starting"` (one per attempt).

**2. "Whether a parent agent hangs" is SETTLED, and the answer is worse than a hang.**
`handleError:596-600` calls `sendErrorResponse` → `CreateResponseContext("complete", 100)`
(`context.go:79`) → status header **`complete`**, `Success: true`, failure only in the body
map. `coordinator.go:316-331` switches on that header: `case "complete","success"` →
`handleCompleteResponse`. **The parent marks the awaited step COMPLETE with the error blob
as its step data.** So `bugs_open/029`'s hung-parent reading is falsified for this path, and
what actually happens is a silent success-shaped failure — same family as `bugs_open/132`.
Registered as RSH-005's primary landmine. **Not fixed here, and not filed as a bug by me**:
I have not measured it end-to-end with a real awaiting parent and would be filing a symptom.

## 2026-08-04 — a hole in my own fix, found by my own test

`ExplicitlyRetryableIsNeverPermanent` failed on my first implementation. An error built
`AsRetryable` correctly skipped the typed branch (`!de.Retryable` guard) — and then **fell
through to the substring fallback**, which matched the word "validation" in its own message
and classified it permanent. **Structure overridden by prose: this bug in miniature, inside
the fix for it.**

Fixed with an early `if de.Retryable { return "" }` that bypasses the fallback entirely.
Worth recording because the failing test was one I wrote to assert a property I believed
already held — the value of writing the assertion you think is trivially true.

## 2026-08-04 — MISSTEP: I mutated a shared-tree file with a backup that never existed

Full account in `WRONG_CALLS.md`. Short version: my mutation recipe wrote its backup to a
`$SCRATCH` path that had been cleared. The `cp` failed, **`set -e` did not stop the
script**, the mutation applied, and the restore then failed too — leaving
`validation_drop.go` sitting in a deliberately-broken state on a branch other sessions
build from. **Only the trailing `diff` caught it**, and only because that step exists purely
to prove the restore happened.

The mutation itself did its job (three cases failed as designed, which is V1 satisfied). The
lesson is that a destructive recipe must **fail closed**: assert the backup exists *before*
mutating, and prefer a file whose current state is committed so `git show HEAD:<path>`
restores it with no external state to lose. **I had written that recipe into the `192`
RUNBOOK three hours earlier** — it was right about what to do and silent about what must
hold for it to be safe, which is how a correct-looking procedure becomes a hazard. Corrected
in place there.

Second, smaller: a `python3 - <<'PY'` heredoc whose *content* contained the terminator `PY`
closed early and leaked shell fragments. No damage (the stray `cp` failed on a missing
operand), but the fix is to use a delimiter that cannot appear in the payload.

## 2026-08-04 — state

Committed `28ef7a044`, council `Council-Submitted: 9b1254f0-2686-4a52-b736-1e212634ace6`.
Registered **RSH-005**; index headline re-grepped **1,721 → 1,760** (~38 rows arrived from
other lanes in the few hours since WFA-009 — never carry that number forward).

**The fix is Go-only, so it is inert until the next chassis roll.** `195` therefore stays
**OPEN**: the defect is still reproducible on the running binary. What closes it is the
post-roll induction in the PLAN's verification section.

## 2026-08-04 — verified the census I had SUBMITTED but not personally checked

The blast-radius claim in my council submission — *"`ErrValidation` is built in exactly one
non-test place, and that place sends it as a response rather than returning it through the
classifier"* — came from the planning pass, not from my own grep. It is in `grounded_in`,
so a reviewer will check it. I checked it first.

**It holds, but not for the reason the plan gave.** The plan named
`internal/agents/contentcreator/agent.go:226` as the construction site. My own grep found
that site plus something the plan missed: `platform/errors/errors.go:194` is
`return New(ErrValidation, "Validation failed")` **inside the public helper
`func ValidationError(field, issue string)`** — so every caller of that helper produces an
`ErrValidation`, and the blast radius is the helper's caller count, not one literal.

Counted: `errors.ValidationError(` has exactly **one** real caller — the contentcreator
line — and it is `a.sendErrorResponse(headers, errors.ValidationError("payload", "invalid
JSON"))`, i.e. **sent as a response, never returned up the stack into the classifier**. So
the submitted claim is true. But it was true by luck of the helper having one caller, not
because there was one literal, and the reviewer would have been entitled to the sharper
version.

> **The irony, recorded because it is the same lesson one level up:** my first grep pattern
> was `ValidationError(`, which matched **`recordDroppedValidationError(`** four times — a
> substring hit on an unrelated symbol, which made the blast radius look 5× larger than it
> is. **I was measuring a substring-matching bug with a substring match.** The fix is the
> same in both places: match the thing, not text that contains the thing —
> `grep -w`, or an exact boundary — and read the hits before counting them.

## 2026-08-04 — council REVISE, and the gating objection found a REAL gap in my fix

`9b1254f0`: **REVISE**, gated by `guardian`, 6 abstained, 0 unreadable, not truncation-gated.
`bug_historian` called the core fix *"the strongest-shaped fix I've seen against this pattern
in this batch"* and still objected — correctly, on scope.

**`editquality` (medium) — the one that mattered, and it was right.** My unconditional
recorder went into `agentbase/agent.go` only, and the seat asked whether `handleError`'s
failure path can reach a caller *other* than `agent.go`'s `processMessage` — *"checkable from
source and worth resolving before approval, since it's exactly the class of gap 195 itself
describes."*

Checked. **There IS a second caller:** `platform/agentbase/server.go:110`, in
`AgentServer.processMessage` — and it is worse than the objection supposed. That loop logs
the failure to the pod log and commits: no classification, no `handleProcessingError`, **no
`agent_error_log` row at all**. A failure arriving there is invisible in exactly the way 195
is about.

**But it is unreachable, measured:** `AgentServer` is referenced *only inside `server.go`
itself*, and `NewAgentServer` has **zero callers anywhere in the tree**. Every live agent runs
`Agent.processMessage` in `agent.go`. So no live failure bypasses the recorder, and the
honest answer is "second caller exists in source, is dead code" — not "there is only one
caller", which is what I had implicitly claimed.

Left the dead loop in place but **documented it where it fires**: a comment naming it
unreachable and stating that reviving it without porting `recordFailedProcessing` and the
`MatchedPermanentFailure` branch reintroduces 195 wholesale — invisibly, while it does so.

**`bug_historian` (medium + low) — two findings I had deferred are now TRACKED, not just
doc'd.** The seat's point is the durable one: a landmine note is one missed doc-read away
from being lost, and *"deferral is reasonable triage, but nothing creates a tracked
follow-up"*. So:
- **`bugs_open/196`** — the success-shaped failure envelope (failures reach the parent stamped
  `complete`, `Success: true`, and the coordinator dispatches on the header). Filed with its
  unmeasured half stated plainly: I have not induced it with a live awaiting parent, and the
  whole severity rests on that.
- **`bugs_open/197`** — the sibling retryable-side classifier (`isRecoverableError`,
  `IsRecoverable`) still deciding by substring over prose. Filed with **no live instance**,
  deliberately: waiting for one is the failure mode, and 195 is the proof. It names the census
  owed before any fix, which is cheaper now precisely because 195 made the failure record
  unconditional.

**Three lows, all "you asserted, you didn't cite":** `recordFailedProcessing` is described as
a thin wrapper mirroring its sibling, with no citation. Fair. The sibling is
`recordDroppedValidationError` (`agent.go:1389` and `validation_drop.go:150`), both thin
wrappers over the one writer `orchestration.LogAgentError` — the shape the reuse seat's own
prior ruling (council `180d7c68`) permits, forbidding forked INSERTs. Cited in the resubmission.

## 2026-08-05 — council round 2 APPROVED; and I misread my own verdict query first

**`9b1254f0` round 2: APPROVED**, 5 advisory objections, none high-severity.
`bug_historian`: *"unusually self-aware for this council … that is the right shape of response
to this pattern."* Round 1 REVISE (08-04 20:05Z) → round 2 APPROVED (08-05 10:13Z).

> **My own measurement error, caught immediately and worth the entry.** My first check was
> `SELECT count(*), coalesce(max(metadata->>'decision'),'pending')` — which returned
> **`revise`**, and I reported the round as REVISE. `max()` on a **text** column picks
> **alphabetically**, and `'revise' > 'approved'`. The verdict was APPROVED all along. The
> fix is `ORDER BY created_at DESC LIMIT 1`, and the general rule is the one this lane has
> now hit three times in different clothes: **an aggregate over a column whose ordering is
> not the ordering you mean answers a different question than the one you asked.**

## 2026-08-05 — the one medium objection, answered on an INDEPENDENT axis

`editquality` (medium) was right to distrust my method, and cited the landmine set against it:
my *"`AgentServer` is dead code"* claim rested on **grep-for-callers of `NewAgentServer`**, and
entry-point-shaped components *"read as an orphan on every database/callgraph axis whether
dead or in daily use"*. A separate `cmd/` binary, env-var dispatch or reflection could
construct it without a textual call.

**Checked the compiled artefact instead of the source** — a genuinely different axis, and the
one that answers the objection, because reflection or a second binary would both leave the
symbol in place:

```
running chassis v1.0.1254 (pod agent-chassis-d69d4467c-dvn8k):
  grep -ac "AgentServer"            /app/agent-chassis  ->  0   (linker stripped it)
  grep -ac "recordFailedProcessing" /app/agent-chassis  ->  3   (positive control: a LIVE method IS present)
```

Go's linker performs dead-code elimination, so a symbol absent from the shipped binary while a
live sibling is present is strong evidence of unreachability — and it is evidence source-grep
structurally cannot give. **Stated limit, so nobody over-reads it:** this proves it for the
`agent-chassis` binary, which is the one that runs agents. Other service binaries were not
checked, and if `AgentServer` is ever wired into one, the `server.go` comment is the warning
that fires.

The three low objections are all "documentation/tracking, not a mechanism fix" — correct, and
I should have named them as such rather than presenting them as parity with the code fix.
`bugs_open/196` and `197` remain live exposures, filed and unfixed, which is the honest state.

## 2026-08-05 — another session's code review found FOUR real errors in my shipped work

Commit `f887ed1ad`, *"fix(code-review F1,F2,F8,F14): four findings against the shipped 195
classifier — all behaviourally inert"*. Not mine, and I am not reverting any of it. Two I can
see directly in the files, and both are errors I made:

- **F1 — my own comment contradicted itself.** I wrote *"This change only ever ADDS permanent
  classifications; it removes none"* six lines below a paragraph arguing FOR a removal (the
  `de.Retryable` early return, which stops an `AsRetryable` error whose prose contains a needle
  from being classified permanent). The code was right; **the claim was wrong**, and it is now
  corrected in place with the reason it is latent today (nothing sets `Retryable=true`
  fleet-wide, so the first producer inherits the behaviour — which is exactly why the two had
  to agree *before* that producer exists).
- **F8 — `IsRetryable`/`GetRetryAfter` still used bare type assertions.** I added
  `AsDomainError` precisely because `err.(*DomainError)` answers false for a `%w`-wrapped
  error, migrated the two call sites in `handleError`, and **left these two untouched in the
  file I was editing**. Zero callers fleet-wide, so inert — but it is the same defect, in the
  same package, in the same commit's blast radius, and I walked past it. This is a partial
  down-payment on `bugs_open/197`.

Recorded because the pattern is mine, not the reviewer's: **I fixed a class at the sites my
symptom pointed at, and not at the sites the class actually lives.** That is the `016b` §9
shape I had just finished writing up.

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

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

# PLAN — 164: the diagnosis bundle's body cap discards in silence

**Lane opened 2026-07-31.** Bug: `bugs_open/164_HANDOFF_2026-07-31_bundle_body_cap_breaks_instead_of_skipping_and_drops_the_rest_of_scope.md`.
Filed by the `bugfix_145` lane **at the council gate's explicit request** (the
`bug_historian` seat objected, medium, corr `bce4caab`, that a disclosed-then-declined
adjacent defect must be FILED, not footnoted).

## What is broken

`platform/orchestration/actions/diagnose_assemble_bundle_action.go`, the in-scope
body loop. Two discard paths, inconsistent severity, **both invisible to the
consumer**:

| path | pre-fix behaviour | consequence |
|---|---|---|
| body cannot be READ | `continue` — siblings still render | symbol vanishes with no explanation |
| body does not FIT | **`break`** — the whole rest of scope is abandoned | an alphabetical tail is destroyed |

`truncated` was set, but **nothing was written into the bundle text**. The two
sibling caps in the same file (`:520`, `:604`) both write a marker before breaking.
This loop was the sole deviation from a convention its own file established twice.

## Why it is a loop bug, not only a rendering bug

Scope arrives `sort.Strings`-SORTED (`pkg/diagnose/loop.go:390`, `:416`). The verdict
names its `next_scope` from what it could see, and the convergence guard measures
whether scope NARROWS — so a verdict formed on a silently short bundle can converge
on the wrong region and have that recorded as **progress**.

## Decisions, and their reasons

1. **Measure before fixing.** The filing left the rate `[UNMEASURED]` deliberately
   and instructed the fixing thread to establish it first. Done — see NOTES. It has
   fired: **18 of 254 bundles (7.1%)**, worst case **14 of 18 symbols lost**, and
   **three bundles rendered the `## In-scope code` heading with nothing beneath it.**
   That last is not an inference; the artefacts were read.
2. **Candidates 1 + 2, together.** `continue` + an inline per-symbol marker, plus a
   conditional coverage line. Ranked first by *what closes the door*: after it there
   is no representable difference between "not in scope" and "did not fit".
3. **Fix the SIBLING read-failure path in the same edit.** Not scope creep — six
   lines up, same loop, same silence. Fixing only the cap half would leave exactly
   the 016b §9 / bug-021 shape ("one call site gets the rigorous fix; the sibling
   stays heuristic") that the council flagged on `165`, and would mean filing a
   second bug for an identical defect in the same loop. **Disclosed explicitly in
   the submission** rather than left for a reviewer to notice — that omission is the
   whole reason 164 exists.
4. **The two markers are worded DIFFERENTLY on purpose.** A size omission is a
   *coverage* signal; a read failure is a *defect* signal. `bd003f67a` already ruled
   on this distinction at the sibling cap in this same file: a defect signal must not
   render as "coverage was capped".
5. **DECLINED — candidate 3, the per-symbol budget** (`maxBodyChars/len(scope)`, by
   analogy with `siblingSignatures`' fair share). Stated so it can be overruled:
   `siblingSignatures` allocates among many *tiny* signature lines; bodies are few
   and large. A fair share over an 18-symbol scope is 3,333 chars, which would **omit
   most real function bodies that render fine today** — reducing coverage in the
   common case to improve fairness in the rare one. First-come-first-served plus a
   visible, individually re-requestable omission is **recoverable**, and
   recoverability beats a cleverer allocator.
6. **DECLINED — candidate 4, bound the read.** It is the knob `bugs_closed/145`
   declined for being a knob rather than a boundary; it belongs to that boundary if
   anywhere.
7. **DECLINED — a `pattern-check.py` gate for the shape.** Surveyed the premise
   before building it (NOTES): the entire repo has **three** char-budget cap sites,
   all in this one file, and two were already correct. A heuristic detector guarding
   a population of three, in one file, is poor value against its false-positive
   risk. The convention is enforced by the file's own tests instead.
8. **No config flag on the markers.** A default-OFF safety marker is the
   mechanism-rots-unexercised failure the owner ruled against on 2026-07-29, and the
   pre-fix behaviour is not a state anyone should be able to opt back into.

## Additional, beyond the filing's candidates

Persist `symbols_omitted_size` and `symbols_unreadable` separately in the artefact
metadata (jsonb, no DDL). The reason this bug was filed `[UNMEASURED]` is that the
artefact carried only a boolean, so `symbols_in_scope − symbol_count` conflated the
two paths and the cap's own rate was not separable. **Make the next measurement a
query, not a research task.**

## How this is verified

- Four tests in `diagnose_assemble_bodycap_test.go`. **Induced, not assumed**: the
  action was reverted to HEAD and the tests re-run — three FAIL against the pre-fix
  loop. Each asserts BOTH branches (137's lesson: the finding branch alone is
  satisfied by deleting the guard).
- Negative control: a fitting scope matches the exact pre-fix format string. It
  passes against *both* versions, which is the point of a control.
- Compiled and tested against a clean `git archive HEAD` with only these two files
  overlaid — the shared tree carries three other sessions' WIP in this same package.

## Council

Submitted **before** committing: `SUBMISSION_CORR 75f3cd52-316c-4cb3-a55d-1b1c3f316214`.
Committed with `Council-Submitted:` (the trailer that asserts nothing), per the
2026-07-30 rule — `Council-Reviewed:` goes only on a verdict actually read.

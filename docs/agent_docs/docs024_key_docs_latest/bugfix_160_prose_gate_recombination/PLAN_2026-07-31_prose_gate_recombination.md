# PLAN — 160: the report prose gate rejects a legitimate recombination

**Started** 2026-07-31 ~20:30 BST, picking up `bugs_open/160` (filed earlier today by the
`robot_hands_gripper_dossier` lane, which has since closed its pilot down — it filed this on
the way out and did not take it).

## The defect, in one paragraph

`verify_report_prose` is fail-closed: a violation destroys the whole report (no page composed,
URL 404s). Its SKU classifier `modelNumberRe`
(`platform/orchestration/actions/verify_report_prose_action.go:241`) treats any letter-digit
adjacency followed by a hyphen and more word characters as a model number, and clears it only
if it is **contained verbatim** in the fact block / request context, or overlaps a scored
candidate name. So `IP54-or-better` — a real rating the writer composed into an English phrase,
with `IP54` sitting in the fact block — is reported as an invented part number and the report
is destroyed.

## The design decision, and why

The bug file offers two shapes. Neither is taken as written:

- **Candidate 1 (segment provenance) as filed** clears a segment when it is *"digit-free
  English"*. That opens the hole the bug file itself names: `EGP-50-X` decomposes into `EGP`
  (traceable), `50` and `X` — and `X` is digit-free, so a rule phrased that way clears a
  fabricated sibling.
- **Candidate 2 (strip hyphens, compare both ways)** weakens the guard by exactly the amount
  substring matching always does, and reverses the containment direction, so a token can clear
  by *containing* a fact rather than *being* one.

**What is being built instead: a traceable head + qualifier tail rule.**

A SKU-shaped token clears when it splits into a **head** that traces (verbatim in the allowed
text, or overlapping a scored candidate name — the existing test, unchanged) and a **tail**
whose every hyphen segment is an ordinary English qualifier. `modelNumberRe` guarantees the
letter-digit adjacency lives in segment 0, so the head always carries the code part: **the
digit-bearing head must still trace verbatim.** Nothing that is rejected today by the numeric
gate, the vendor gate or the no-match contract is touched — this is purely additive relaxation
of check (2).

A tail segment is a qualifier when **all three** hold, and each clause has the counterexample
it exists to reject:

| clause | rejects |
|---|---|
| contains no digit | `EGP-50-X`, `2F-140` — a digit in the tail *is* a variant number |
| at least 2 characters | `2F-85-X` — single letters are SKU suffix material (`EGP 40-N-S-B`) |
| all lower-case | `2F-85-XL` — upper-case codes are SKU suffix material |

**Stated limitation, deliberately not solved:** the third clause means a title-cased qualifier
(`IP54-Rated` in a heading) is still rejected. The alternative is a closed vocabulary of
qualifier words, which is speculative machinery for a case nobody has observed. The asymmetry
decides it: an over-strict rejection produces a retry whose different phrasing usually passes
(016b records that this very report passed on retry), while an under-strict clearance publishes
a fabricated model number to a customer-facing report. If the title-case case is ever observed,
it is a one-line change with a test.

**Not in scope, and not oversights:** the fail-closed behaviour stays (016b and a prior council
HIGH both establish it is right); no attempt to make the failure louder on a dashboard; the
already-deleted robot-hands fixture pages are the gripper lane's own cleanup and are not
regenerated here.

## Phasing

1. Reproduce in a unit test against the real scoring fixture — the failing assertion first.
2. Implement head/tail decomposition beside the existing checks.
3. Prove **both** halves (016b's explicit instruction): the recombination clears **and** four
   fabricated siblings are still rejected, asserting on the **violation text**, not merely on
   "there was a violation" — the numeric gate can absorb the mutation otherwise.
4. Mutation-check: revert the helper, confirm the new tests fail for the right reason.
5. Council gate (platform code), commit narrowly, then image + roll, then verify at the pod.

## Decisions log

- **20:35** — took 160 over 123/132/158. 123 needs an owner call before anything is
  publishable; 132 is B2/Cloudflare infrastructure with its candidate 4 already shipped; 160 is
  code-local, unit-testable, and has a live consequence (a destroyed report).

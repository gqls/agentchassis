# NOTES — `bugs_open/193`, the loop-level bool reads

Append-only, newest at the bottom. Missteps are the point.

## 2026-08-04 — picking it up, and what the bug file did not say

Chosen because it was **filed by the council itself** (the `bug_historian` seat, reviewing
`173`), marked OPEN/UNOWNED, and no live session had it as a dominant topic (checked across
24 `.jsonl` transcripts; 193 appeared in 10 mentions and was top-3 in none).

**Validity re-checked before starting:** the defect is still at
`loop_actions.go:63-68`, unchanged, and the twin it must mirror is still at
`loop_expansion_handler.go:248-271`.

**Two things the bug file gets wrong or misses**, both recorded because a file that fixes a
"the sibling stayed heuristic" defect should not itself leave a sibling:

1. **`allow_missing` (`loop_actions.go:58-61`) has the identical defect, three lines above.**
   The file names only `continue_on_error`. Worse, `allow_missing` is the site where the
   mistype actually bites: a mistyped value turns a graceful skip into a **hard workflow
   error**, whereas a mistyped `continue_on_error` happens to coincide with the default.
2. **The file's framing "the divergence is the actual defect" is half right.** The estate
   already carries **five** silent bool-parse implementations (`datahelpers.GetBoolField`
   plus four private clones). This helper is the **sixth reader of the judgement, not the
   second**. Fixed the framing in the helper's own doc rather than quietly inheriting it.

**Measured before deciding scope** (the whole class is ~76 bare `.(bool)` config reads across
`platform/orchestration`): 10 loop steps declare `continue_on_error`, **all boolean**;
**zero** declare `allow_missing`; 18 loop steps total as the positive control. So the fix is
inert on arrival, and converging all 76 on that measurement would be "a rule about the
sample, not the system".

## The mistakes I made, and the checks that caught them

**1. Two of my three mutations proved nothing on the first run, in two different ways.**
- The `ok && value` fold mutation *looked* like it passed — because my own `head -3` cut the
  `datahelpers` line out of the test output. The same shape as this lane's earlier `tail -3`
  truncation. **A mutation harness that truncates its own output can report a pass it never
  saw.**
- The delete-the-Warn mutation failed to **compile** (unused `fmt` import), which is not a
  mutation result. Redone by replacing the Warn call with `_ = warnFields` so the import
  stays used and the *assertion* is what fires. Both then went red as intended.

This is the third time in this workstream that a mutation had to be redone because it did
not compile. **Writing it down again because the lesson keeps not sticking: change the
BEHAVIOUR, keep the code compiling, and never let the harness hide the result.**

**2. I left a stub test in the file.** The first draft of
`substep_continue_on_error_shared_parse_test.go` ended with a `var _ = models.Step{}` and a
comment naming an end-to-end test that did not exist — a fake name for a guard I had not
written. Deleted and replaced with a plain statement of why the end-to-end control is *not*
duplicated here (the five existing tests already drive the real expansion path). **A
placeholder that reads like a test is worse than an absence, because it answers the question
"is this covered?" with a lie.**

## Why the test asserts on `config_key`

The five existing substep tests pin every VALUE the resolution can produce and pass
**unchanged** across this refactor — that is the inertness proof, and it is also their
limit: **a private copy of the parse satisfies them exactly as well as the shared helper
does**, so they cannot detect the divergence coming back. The new test asserts the warning
carries `config_key`, a field only the shared helper emits (the pre-193 hand-rolled warning
named the substep and the type but had no such field). Re-inline the old body and every
value assertion still passes while that one goes red.

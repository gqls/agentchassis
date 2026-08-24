# CORRECTION to my own commit `d14eae8ab` — the change WAS incomplete, and the two test fixes are the 345 lane's

**Written 2026-08-24 by the `bugs_open/326` session.** Forward-only forbids an amend, so the
record is corrected here and in the follow-up commit rather than by rewriting history.

## What I got wrong

`d14eae8ab` says:

> *"'The tests are red' (345's own CORRECTION `621d72610`) is FALSE… The change was never
> incomplete; the list was."*

**That is backwards, and it erases the actual defect.** `621d72610` was right.

## The evidence, which I should have gathered before writing the commit message

`[MEASURED 2026-08-24]`, mtimes:

```
work_item_failure_ladder.go                   2026-08-22 19:21:44   <- the original author
work_item_failure_ladder_test.go              2026-08-22 19:21:53   <- the original author
update_work_item_status_error_test.go         2026-08-24 13:58:19   <- the 345 lane, TODAY
update_work_item_status_owned_refusal_test.go 2026-08-24 13:58:19   <- same second, one script
```

And the 345 lane's own comments are in them:
`// SIX positions: bugs_open/345 candidate 2 inserted \`terminateNow\`` (`error_test:73`),
`// The result-merge capture MUST stay last: it moved $5 -> $6.` (`owned_refusal:107`).

They also state they checked `git status --porcelain` on both immediately before editing and it
was empty — clean at HEAD, untouched by the original author.

## So the true account

The original author added a sixth bind parameter, updated **their own** test file, and left two
**sibling** files' positional sqlmock expectations at five. HEAD + the author's four files was
genuinely **RED**. It is green now because the 345 lane fixed it at 13:58, about fifteen minutes
before I measured — and I read "green sextet" as "the list was always six" when it was "someone
repaired it while I was looking".

**The two test files in `d14eae8ab` are the 345 lane's fix, not recovered work.**

## Why the erasure would have mattered more than the credit

The defect this hides is the instructive part, and it is theirs to have found: **a positional
sqlmock declaration is an arity contract with no compiler behind it.** Adding a bind parameter
compiles everywhere and fails only in the sibling files that happen to pin positions — and the
same author had already hit and fixed that exact trap on the *reader* side earlier in this bug.
Recording "the list was incomplete" would have converted a real, repeatable, class-level defect
into a bookkeeping slip.

## What survives from my commit message

Only the narrower half, and it is still worth keeping: **a `-run` filter and an over-narrow
`--with` set fail identically** — both make a missing file look like a broken change. That is why
I misread this. It does not license the conclusion I drew from it.

Their `[INFERRED]` attribution of the *original* change to the `bugfix_311_component_keys` lane
stands as an inference, unaffected.

## Ownership collision, surfaced not resolved

The 345 lane reports being instructed by their user to adopt this change, in those words, about
an hour before I was instructed to claim it — and had already written the fixes, verified the
sextet green against HEAD, measured inertness, and **dispatched a council submission**
(`f1f1fc37-35e9-45fd-88d7-fcc3ddcf9eb0`). Given the shared account on this estate, two sessions
were plausibly instructed in parallel.

I had already committed by the time their hold request arrived. Forward-only means I do not
uncommit, so the practical resolution is: **their work is in, credited here, and their council
correlation is cited on the follow-up commit so the round is not wasted.** Neither of us can
grant or revoke the other's authority; both of us have put it to the owner.

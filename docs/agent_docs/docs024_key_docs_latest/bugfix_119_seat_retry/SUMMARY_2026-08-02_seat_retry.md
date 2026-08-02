# SUMMARY — 2026-08-02 — the reviewer that couldn't be read, and the instruction nobody was giving

*(Milestone read-out: the fix is live in production and the bug is closed. Written to be
read aloud. Previous: `SUMMARY` was not written for this lane before now — the earlier
milestone was recorded in `README_where_we_are.md`, as the cadence rule prefers.)*

## What we're trying to do

When we propose a change to the platform, we send it to a review council: sixteen or so
independent reviewers, each asked to judge one aspect and reply with a small structured
document. The verdict is assembled from their replies.

If **one** reviewer's reply cannot be read by our code, the whole round is thrown away —
every other reviewer's completed work with it. The submitter gets back a verdict that says
nothing about their change, and pays again in credits and about half an hour to try. We
wanted that to stop costing us whole rounds.

## Where we've come from

The bug was filed on 27 July with a specific cause: a reviewer producing a document that was
complete but had a bracket in the wrong place.

The first thing worth reporting is that **the filed cause turned out to be unmeasurable, and
the filed defect turned out to be entirely true** — and telling those apart changed what we
built. We confirmed the damage was real and worsening: of 429 review rounds ever run, **23
were decided by "we couldn't read one reviewer"** rather than by anyone's judgement, 15 of
them in a single week. But when we went looking for live examples of the bracket slip so we
could fix exactly that, there were none — not few, zero. The evidence for most historical
cases had already been deleted by routine cleanup, and every survivor was a different
failure: the reviewer ran out of room and produced nothing at all.

The lesson we took from that, and wrote up fleet-wide, is that **a bug report states its
damage and its cause in one voice, but they are measured from different places** — damage
from a table kept for months, cause from evidence pruned in days. Re-confirming the symptom
feels exactly like re-confirming the diagnosis, and it isn't.

## What we've done

Two things, plus a third the reviewers asked for.

**We made a setting that the whole fleet writes actually mean something.** Each step can
declare what kind of answer it needs. The code read a setting called `output_type`; the fleet
writes `output_format` — a near-identical name that **nothing anywhere read**. Ninety-one
steps across thirty-three agents were declaring "I need JSON" into a void, every council
reviewer among them, and were being handed a generic instruction sheet that does not mention
JSON at all. The final measurement is stark: over four months and twenty-five agents,
**9,061 of 9,063 calls** got the wrong instruction sheet, and the setting the old code *could*
read has not been used by a single call in that entire window.

**We made the platform ask again.** When a step that asked for JSON gets back something
unusable, it now re-asks exactly once — not with the identical question, which just
reproduces the same failure, but with a short corrective note matched to what went wrong:
"you ran out of room, same judgement but shorter" or "that wasn't valid JSON, same answer as
one valid document". If the retry also fails, everything behaves as before.

We deliberately did **not** raise the reviewers' output budget, though that was our first
instinct. Our own code carries a note from an earlier round of this problem saying, in
effect: don't — whoever has the most to say will always reach whatever the ceiling is. That
was borne out live. Another thread has been raising the limit for reviewers as each one
fails; four are now on double, thirteen are not, and the failure that hit our own round 1
landed on one of the thirteen. Raising the ceiling relocates the problem. Asking for the same
judgement more briefly does not.

**And we stopped the failure being silent.** A reviewer on the council made a sharper point
than we had: a retry makes the failure rarer without changing what happens when it still
fails — the step goes on quietly returning prose to something that asked for structured data
and reporting success. So an unmet contract is now marked on the result. We stopped short of
making it a hard error, which would convert ninety-odd currently-limping steps into outright
failures over content they didn't author.

It went through the council twice: **REVISE**, then **APPROVED**. The first objection was
correct and we had earned it — we had claimed a reach of ninety-one steps without checking
the setting sat where the code could see it. Checking confirmed the assumption *and corrected
our own count upward*, because our census had walked only top-level steps and missed one
nested inside a loop.

## Where we are now

**Live and closed.** Chassis v1.0.1228 rolled this morning; we verified it inside both
running copies of the service rather than trusting the version number, with controls designed
so they could actually fail.

Two measurement errors were caught in the course of proving it, and both are now written up
fleet-wide because neither is specific to this bug:

- Our own verification instructions told the next person to check for a line the change
  deleted — but that line was a code comment, and comments don't survive into the finished
  program. It would have come back "not found" whether or not anything shipped, and read as
  a pass. **Check that a control can fail before trusting that it didn't.**
- More seriously: measuring whether the missing instruction was really missing, we counted
  prompts containing it and got 31 out of 9,063. All 31 were ours. A submission's explanation
  is pasted into every reviewer's prompt, and we had quoted the missing instruction verbatim
  while explaining that it was missing — so our own bug report registered as evidence the bug
  wasn't there. The giveaway was the shape, not the number: exactly two per reviewer is two
  review rounds, not traffic.

That is the second time in two days on this one bug that **writing something down changed
what we then measured** — the first being that a landmine note written about this change came
back as six reviewers objecting to it. It is a real property of a platform where
documentation feeds the systems that observe the thing documented, and it argues for better
detectors, not less writing.

**One gap, stated plainly rather than papered over:** we cannot yet show the fix working in
production. The fleet is nearly idle — four calls in the half hour after the roll, none of
the kind this change touches — and no council has convened since our own. So the retry has
fired zero times out of roughly zero opportunities, which is evidence of nothing, and the
wasted-rounds figure is unchanged at 23 of 429 because no new rounds exist to count.

We also closed against the project's stated bar (*fixed and live*) rather than the stricter
one our own handoff had proposed (*the wasted-rounds figure falls*). That figure depends on
other people submitting reviews, not on this fix; it was unsatisfiable by anything this lane
could do. We recorded the change of bar rather than quietly dropping the promise.

## Where we're going

Nothing here is blocking, and none of it needs the author.

- **When the fleet is next busy, re-run two queries** — both sit in the closed bug file: did
  the re-ask fire, and does it stay rare; and has the wasted-rounds figure moved off 23 of
  429. That is the outcome measurement we are owed, not an open defect.
- **The bracket-slip class is still not claimed as fixed.** It has zero measurable live
  instances, so the link between the missing JSON instruction and that specific failure
  remains an inference, and is labelled as one.
- **The review-harness defect deserves a decision.** Reviewers cannot distinguish a warning
  written *about* a change from one written *by* it, so following the rule that says
  "document the seam in the same commit" actively arms the council against you. It is
  written up as a transferable pattern; whether it becomes its own case is a judgement for
  whoever owns the gate.

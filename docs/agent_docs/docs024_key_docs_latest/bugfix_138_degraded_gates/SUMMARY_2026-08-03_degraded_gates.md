# SUMMARY — 2026-08-03 — the test could not be staged, because the reviewer refused to lie

## What we're trying to do

A council of automated reviewers reads every code change before it lands. Each reviewer
has a fixed budget of words; run past it and the reply is cut off, kept as a fragment,
and treated as blocking — because the serious objection might have been in the part we
lost. The bug was that the record then blamed the reviewer, as though it had judged.
The fix has been live for days. One part of it had never been seen working: the new,
honest label that says "this round was blocked by a cut-off, not by a judgement." It
had never appeared, because the rest of our work removed the pressure that produces it.

## Where we've come from

By the first of August all four planned fixes were done and measured, and you decided
to trigger the missing case deliberately rather than wait to observe something we were
actively suppressing. The plan was to shrink one rarely-used reviewer's word budget for
a few minutes, ask it to raise one mild objection, and watch the label print.

## What we've done

**The first attempt missed, twice over.** The reviewer's reply was cut so short the
system discarded it entirely — honestly, exactly as designed — and a different reviewer
vetoed the submission anyway, because its only change was a comment and that reviewer's
own rule says a comment-only plan is an empty one. Ten of eleven approved; one veto
ends the round before the label is ever considered. The disruption window was about two
and a half minutes, and we proved twice over that no one else's round was touched.

**Before the second attempt we rehearsed offline — three direct calls costing pennies —
and the rehearsal ended the whole programme.** First it showed the budget arithmetic
was wrong in a way nobody suspected: the model spends part of its word budget on
invisible deliberation before it writes anything, so a measurement of "how much room is
enough" taken at one setting simply does not carry to another. And then it produced the
decisive fact: given enough room to finish, **the reviewer refused the instruction.**
Told to object regardless of the change's merits, it approved, and wrote that a
reviewer which outputs a verdict contrary to its own judgement because embedded text
told it to would itself be violating its role. The first attempt's cut-off had merely
hidden the same refusal. You cannot stage this test dishonestly on this model — which
is, it must be said, exactly the property you would want your reviewers to have.

**The investigation paid for itself on the way out.** Reading the recovery code closely
enough to stage the test exposed a real defect in it: the routine that rescues cut-off
replies is confused by brackets inside quoted text, and can destroy a reply that was
perfectly rescuable — and review text contains code fragments and brackets all the
time. We fixed that properly, gave the routine its first-ever tests, and put it through
the council on its own merits. It was approved — by the same reviewer whose veto had
stopped round two, now satisfied — and after this morning's deployment it is verified
running in production. The verification itself taught one more lesson: the check I had
promised was impossible by construction (a compiled program contains no comments and no
local variable names, so grepping for either proves nothing), and the honest method —
dating the build by which neighbouring changes it does and doesn't contain — is now
written down where the next person will find it.

**And that dating method caught something affecting another team.** The build running
now demonstrably predates a change that another workstream recorded as verified-live
this morning under the same version number. Two different builds appear to share one
version tag, and today's deployment may have quietly rolled their change back out.
They have been given the evidence, in their own log, without speculation.

**The wrong turns are recorded as carefully as the wins.** This lane's recurring error
— reasoning about what a mechanism does from its name or a past measurement instead of
its current behaviour — bit three times in three days in three different costumes, and
each instance is now a written check at the place where the mistake gets made.

## Where we are now

Everything this bug can produce has been produced. All four fixes are live. The
instruments are running. Nobody else's work was harmed, and one hazard to another
team's work was found and handed over. The only artefact never witnessed in production
remains that one honest label — and we now know why: the fix suppresses the failure,
the model will not fake the verdict, and the only remaining honest route needs a
deliberately imperfect submission plus luck, at maybe one-in-two or one-in-three odds
per costly attempt.

## Where we're going

One decision, and it is yours: close the case on the strength of the aggregate evidence
— every ingredient of the label separately proven in production, the assembly covered
by tests — or leave it open as a passive watch for the label to appear on its own. My
recommendation is to close it. Either way, no work remains; only the decision.

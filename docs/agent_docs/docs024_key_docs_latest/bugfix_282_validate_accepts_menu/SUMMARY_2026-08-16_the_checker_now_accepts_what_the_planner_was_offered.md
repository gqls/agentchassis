# The checker now accepts what the planner was offered

*Milestone read-out, 2026-08-16. Bug 282 — designed, built, reviewed and
approved; awaiting a build to go live.*

## What we're trying to do

Stop the site-rebuild pipeline from silently throwing away parts of its own
plans. Specifically: when the agent that designs a page proposes a calculator,
the pipeline should keep the calculator.

## Where we've come from

In August we taught the planner that it may use a site's own calculators. That
was half a change. The step that checks the planner's proposal against reality
was never taught the same thing, so it did not recognise a single calculator as
a legitimate component. On the loan-calculator site the planner duly put a
calculator on each of the twelve calculator pages, and the checker deleted all
twelve on the way past.

Nothing broke visibly. There was no error, no failure, no alert — just a saved
plan describing twelve calculator pages with no calculators on them. Nothing
live was damaged, because a separate rule stops us rebuilding a tool page
unattended and the twelve calculators are individually locked. But twelve
redesign tickets have been held since Friday waiting for this, and the first
diagnosis blamed the planner's judgement, which was wrong.

## What we've done

**Fixed the specific bug.** The checker now reads the very list of components the
planner was shown, instead of consulting a separate hardcoded list of its own.
There is one list, in one place. If we widen what the planner may use again, the
checker follows automatically, because there is nothing to remember to update.

We deliberately did *not* copy the rule across, which was the obvious fix and the
one the bug report recommended. That rule is not code — it lives in the database
as configuration, and it had already been edited once since the bug was filed.
Copying it would have left three copies of one rule that must never disagree,
which is precisely how this bug happened.

**Then the review made it a bigger fix, twice.** We put it through the reviewer
council, as the rules require for platform code. It came back "revise" twice
before approving, and both objections were right:

- The first said we had stopped the calculators being deleted but left the
  underlying habit intact — *any* section thrown away, for any reason, on any
  site, still vanished with nothing but a log line that is gone within ninety
  seconds. Now every discarded section leaves a permanent record.
- The second said we had put that record in one place, while three other places
  in the system throw sections away through *exactly the same code*. Worse, those
  three are on the busier path — they run about a hundred and sixteen times a
  month against a handful for the one we had fixed. All four now keep records.

**And we found that our own checks were the weakest part.** Three tests could not
have failed: two passed with the new code removed, and one deploy check could not
tell "shipped" from "not shipped" at all. All three are fixed and written up. A
test that cannot fail is worse than no test, because it produces the confidence
without the check — which is the same thing the bug itself was.

## Where we are now

The fix is complete, reviewed and approved. The configuration half is live. The
code half is committed and needs the next build to take effect; until then
nothing has changed for any site.

The record is in the places people will actually look: the bug file, the
long-term concept register, the estate's trap list, the debugging guide, and the
notes of the team whose work was blocked. Two of my own mistakes are logged in
the shared wrong-calls file, including one where the rule I broke was already
written down in my own notes.

## Where we're going

Someone cuts a build — and must raise the version number when they do, or it
ships the old code from cache. After that, the loan-calculator team re-runs their
rebuild and checks that the twelve calculators appear in the saved plan. Only
then does the bug close, and only then do their twelve held tickets get worked.

Two things we noticed and deliberately did not fix are written down for whoever
wants them: sites whose pages use positional slot names have a related exposure
in the same step, and the checker still accepts names the planner was never
offered, which is a tightening we could make but should decide on separately.

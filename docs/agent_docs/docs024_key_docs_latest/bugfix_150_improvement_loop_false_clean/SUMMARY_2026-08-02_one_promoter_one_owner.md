# SUMMARY — 2026-08-02 — one promoter, one owner: the class is closed, not just the bug

## What we're trying to do

Stop a shared piece of machinery — the step that sweeps up everything a site has
found wrong and puts it in the queue — from being run by three different agents at
once. When three of them race for it, whichever gets there first takes everything and
the other two truthfully report "I moved nothing", which one of them then read as
"there is nothing wrong here". That is how a site with sixty-seven queued problems came
to be announced as clean.

## Where we've come from

Yesterday we fixed the *symptom*: the shared step now also reports the state of the
site, not just what that particular run did, so the decision no longer depends on who
got there first. That went live and was proven.

But the review council said, from two independent directions, that it did not close
the underlying problem — it left the next agent someone adds inheriting two similar
signals and a rule for choosing between them buried in a code comment. They were
right, and it became a written proposal with three options for the owner to pick from.

The proposal admitted it was missing one fact: nobody had checked whether anything
*other* than the improvement loop calls the two child agents, and that fact decided
the cheapest option. Asked to explain the proposal for a decision, we ran that check
first. Nothing else calls them — not "probably nothing", but a scan of every live
agent returning exactly two results, both the improvement loop. The audit that had
made the structural option look expensive was done, and it found nothing to audit.

## What we've done

**The owner chose the structural option, and it shipped the same day.** The two child
agents no longer do the sweeping up; only the improvement loop does.

**The half that matters is not the deletion.** Removing two steps is a one-off that
the next agent to gain one silently undoes. So an action can now *declare* that it is
meant to have exactly one owner, and a check reads the entire fleet and reports any
action that has picked up a second one. Run before the change it reported the problem,
naming all three agents; run after, it reported nothing. Same command, same fleet —
that pair is the evidence, rather than anybody's say-so.

**The council approved it**, with two seats raising advisory points and none blocking.
Four seats independently asked for the same check to be done before deleting the step;
it was done, and came back in the change's favour.

**And a real run proves it works**, which config alone cannot: both child agents ran to
completion with their step removed, and the loop's own count went from zero — which it
had been in every run ever observed — to twelve.

## Where we are now

Closed, live, reviewed and proven. One mistake was made and is worth stating plainly:
when the step was deleted, the normal path into it was redirected and the *error* path
was not, leaving a pointer to something that no longer existed. The migration's own
safety check found this, printed it — and the change committed anyway, because the
check was written as a question rather than as a stop. Fixed within minutes; the fix's
version genuinely halts, and was proved to halt by deliberately re-breaking things
inside a transaction that was thrown away. The same question asked across the whole
fleet found no other instance.

That is now written down in three places, because it is the kind of trap that fires
when you touch something rather than when you have a symptom.

## Where we're going

**One open question for you, and it is the sharpest thing the review said.** The new
check reports problems and returns a failure code — but nothing runs it automatically.
Neither does any of the other three audit scripts in the repository. They all need
live database access, so putting one into the commit path for every session would add
a round-trip to the cluster on every commit, across roughly thirty concurrent
sessions. That is a decision about everyone's working conditions rather than a
bug-fix call, so it was deliberately not done unilaterally.

Until then the mechanism exists and is correct, but it protects you only when somebody
runs it. The rest of the workstream is finished.

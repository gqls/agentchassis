# SUMMARY — 2026-07-31 — making a refused plan recoverable (bugs_open/099, candidate 2)

## What we're trying to do

Stop the platform throwing away good work because of a rule it never told anyone.

There is a step in the feature/fix loop whose job is to check the *shape* of a plan
before anything acts on it — sensible things, like "don't list the same file twice in
one stage" — and then save it. If the plan fails one of those checks, the step fails
and the whole run is discarded. The plan itself is produced by a language model that
has just spent real money thinking about the problem, and the design is often
perfectly sound; it falls over on bookkeeping it was never briefed on. Worse, it fails
*silently*: the run reports as completed, the error column in the database is empty,
and the real reason is buried in a field nobody reads. A dashboard shows a clean run
and a good design is gone.

The aim was to fix the class rather than the case: not to teach the model one more
rule, but to make a refusal something the platform can recover from.

## Where we've come from

Somebody had already fixed this once, in July, by adding the offending rule to the
designer's instructions. That worked and is still working. But there are about a dozen
rules in that checker and it gains more over time, so that approach has to be repeated
for every rule, in every agent that uses the step, and it comes undone the moment the
checker changes. The bug file itself said as much and named the real fix — and then
recorded that the real fix was "not done". It had been sitting untouched for three
days when this session picked it up.

## What we've done

The refusal is now recoverable. When a plan fails a shape check, three things happen
that did not happen before. The rejected plan and the exact complaints are **written
down durably**, so a design the loop gave up on can be read back by a person. The
complaints are **handed back to the designer** with a prompt that says plainly: your
plan was not reviewed and not rejected on its merits, it failed a structural check
before anyone saw it, here is precisely what to fix, change nothing else. And it does
that a **bounded** number of times — once by default — before failing exactly as it
does today.

The gate is not loosened. Nothing invalid is ever saved, and nothing reaches a
reviewer that hasn't passed the check. Agents that don't opt in behave exactly as
before, byte for byte, which is why the two other users of this shared step needed no
change at all.

Along the way the review council caught four things worth having. It found that the
rollback instructions in our migration were **wrong** — pointing at the wrong database
table, and sorting by a column that turns out to be identical across every backup, so
a restore would have failed at the worst possible moment. It found we were writing a
third hand-rolled copy of a counter that already existed twice, so that got extracted
into one shared function. It pointed out that making the failure *survivable* hadn't
made it *findable* — which was the original complaint — so refusals now also write an
operator-visible record. And it flagged that the sibling agent with the same defect
had nothing to ever remind anyone; that migration is now written and tested, sitting
one command away for the team that owns it.

## Where we are now

It is live, and we have watched it work on the real system rather than only in tests.
A deliberately-broken plan was refused, the design was preserved in full — about ten
thousand characters of it — the operator record was written, and the run was routed to
the repair step. Every stage of the mechanism did what it was supposed to.

And then that run died anyway, which is the most useful thing that happened all day.
The repair ran, produced a corrected plan, and got one field's shape wrong — an array
where a piece of text was wanted. We had classed that kind of mistake alongside
"the model ran out of room mid-sentence" and made both fatal. They are not the same:
one is a truncated response that should never be retried, the other is a complete
answer in the wrong format, which is the most mechanically fixable error there is. So
the loop preserved the design, recorded it, handed it back, and still lost the run.

That boundary was our own assumption, which is precisely why none of the fifteen tests
we had written could have caught it. It took one real repair round. It is now fixed,
with the two cases deliberately tested as a pair going in opposite directions — but
that fix is code, so it is waiting on the next rebuild.

Three separate times on this bug, a check would have reported success for a reason
unrelated to what it was checking: the bug file's own stated verification procedure,
the first cap we chose to trigger the test with, and the control string we used to
prove the deployment. Each was caught, and each is written down.

## Where we're going

Two things remain. The next rebuild carries the format-error fix, and then one more
deliberately-broken run should show a repair that completes — a design actually saved,
end to end, in production. Until that is observed the bug stays open, because "the
routing works" and "a design was saved" are different claims and only the second one
closes it.

After that, the sibling agent's migration is ready whenever its owners want it, and
there is one open question worth a decision: the checker's size limits ("this plan is
too big") ask for less scope, while our repair prompt forbids dropping scope. Those
two instructions contradict each other, so a size complaint will burn its repair round
and stop. That is arguably correct — a plan that is genuinely too big probably should
reach a human rather than be quietly shrunk by a model — but it should be a decision
somebody makes, not an accident of how the prompt is worded.

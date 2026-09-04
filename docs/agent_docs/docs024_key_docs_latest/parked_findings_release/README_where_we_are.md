# Where we are — releasing the parked findings

Plain prose, append-only, newest at the bottom. The owner maintains this file too — append, never
rewrite or reorder, and never edit his words.

---

**2026-09-04, lane opened.**

You asked for a lane to work through all the parked findings, and to default to approving them
rather than checking every one. This is that lane. Here is where it stands after the first look.

There are **3,184 parked findings across 39 sites** — verdicts the review seats wrote about pages,
recorded but deliberately not acted on. They are still arriving: seven review seats are filing new
ones as I write, and eleven of the twelve sites built since this parking started already have some.

Each parked finding carries its own instructions for how to release it. **Those instructions do not
work, for 89% of them, and they fail silently.** That is the one thing worth your attention today.

The mechanism, in plain terms. Releasing a finding puts it in a queue called "detected". A separate
job then picks things out of that queue and sends them off to be done. That job has five checks it
applies before it will pass anything along. The fifth check was added a week ago, for a good
reason: it refuses anything that came from a model's *opinion* about a page, as opposed to a
mechanical defect that was measured. Its purpose was to stop exactly the kind of unattended rewrite
that damaged a good page once before.

Those parked findings **are** model opinions. They are stamped as such, by the very same code that
parks them. So if we run the release instructions written on the rows, 2,832 of the 3,184 would move
from one holding pen to another and stop there — and it would look like it had worked. The status
changes, a handler gets named, nothing reports an error. Only 352 would actually go anywhere.

I don't think anyone did anything wrong here. The check was added by one lane, which wrote down that
it didn't affect parked findings — true, while they are parked, and false the moment you release
one. The release instructions were written by another lane before that check existed. Nothing joins
the two up, and the join is only visible if you go and read both.

**So the decision in front of you is smaller and sharper than "release them".** That fifth check is
the mechanical form of the thing you have just ruled on. It exists to stop the system approving
model opinions automatically; you have said that for now, on fresh sites, it should. The honest way
to do that is to teach the check to allow a finding that a *person has deliberately released*, while
it keeps refusing ones that nobody has looked at. That keeps the other four checks — one of which is
already usefully refusing a group of 32, because the handler meant to fix those has never once
succeeded at that job.

The alternative is to skip the queue entirely and shove things straight into "ready to run". Faster,
and it throws away all five checks including the one catching those 32. I would not.

**Two other things worth knowing.** The release does not need pacing from us — the picking job is
capped at 20 items every 15 minutes, about 80 an hour, so 3,184 findings is roughly 40 hours of
drip. And released findings keep their original date, which means they sort to the *front* of the
fleet-wide queue, ahead of today's live work, not behind it. That is worth deciding deliberately
rather than discovering.

**What I have not done:** nothing has been released, and I have not yet run the cheap experiment
that would prove the fifth check behaves the way I have read it to. It has never actually fired in
production, so what I am telling you is read off the live configuration rather than observed. I would
like to prove it before acting on it — it is one throwaway row and half an hour.

**What I need from you** is in the last section of the plan: which release mechanism, what order,
and whether "approve each as they go" means per site or per finding — because per finding is 3,184
decisions and I don't think that is what you meant.

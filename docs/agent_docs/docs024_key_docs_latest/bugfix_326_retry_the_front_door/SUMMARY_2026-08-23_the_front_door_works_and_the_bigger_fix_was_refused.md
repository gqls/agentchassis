# The front door works again — and the bigger fix was refused, correctly

*A read-out on `bugs_open/326`, 2026-08-23.*

## What we're trying to do

We sell a one-shot website build: a customer gives us a domain and the system does the rest.
When one of those builds fails halfway — and they do — the obvious thing for whoever is
watching to do is submit the domain again. That has to work. It is the difference between a
setback and a dead end.

## Where we've come from

It did not work. Submitting the same domain a second time produced a green tick and no build.
Nothing ran, nothing was queued, and nothing anywhere said so. The recovery, written down in a
handoff for the next person, was to go into the database by hand and rename seventy-eight rows.

That was filed as a bug five days ago with a clear explanation attached: the system refuses to
file the same piece of work twice, whatever state the previous one is in.

## What we've done

**The explanation was wrong, and that mattered more than it sounds.** The mechanism everyone
was about to fix is innocent — it explicitly ignores finished work, and we can show that from
the database's own definition. The real culprit was a different safety mechanism sitting just
above it: if the same piece of work finished less than three hours ago, the system quietly
declined to file it again, and returned exactly the same answer it gives when the work is
genuinely already running. So the caller could not tell "you're covered" from "I threw your
request away".

The failed build we had was re-submitted two hours and twenty-eight minutes after the first
attempt started. Twenty-eight minutes later and it would have worked. Which also means the
seventy-eight rows of hand-surgery in the handoff were unnecessary — it was surgery for a
three-hour timer.

**None of this was new.** We hit the same thing in July, understood it correctly, fixed it, and
left ourselves a switch for the kind of work where a repeat is normal. That fix reached the
places one person happened to be touching. Nobody counted how many other places needed it:
nineteen out of twenty-one still did not have it, including every step of the customer build.

So the fix has three parts. The build steps are now classified properly — that is live. There is
a new check that lists every step nobody has classified, so this cannot quietly happen a third
time. And a third, larger change would have made the mechanism *delay* work instead of
destroying it, everywhere, for everyone.

**The review council refused that third part, and it was right to.** Its argument: the customer's
problem is fixed by the classification alone; the larger change is a separate decision about how
the whole system behaves, and I had attached it to an urgent bug fix where it would be waved
through on someone else's urgency. That is exactly what I had done.

## Where we are now

**Fixed, live, and proven on a real build.** A greenfield build of a new domain died this
evening on an unrelated defect. It was re-submitted at two hours and six minutes — inside the
window that would have swallowed it that morning — and the work was queued. Checked twice, by
two people, independently, with what each possible result would mean agreed in writing
beforehand. Minutes later the new work had been picked up and was running.

The protection against two people submitting the same domain at once is untouched, and it lives
in the database rather than in configuration, so it cannot be undone by an edit.

**The larger change is written, tested, and deliberately not shipped.** On this system,
committing code is shipping it, so it sits as a proposal with the patch attached, three options
costed, and the decision left to someone else. Fourteen configuration steps and thirty-six
places in the code are still exposed to the original problem — that is what the proposal is
about, and it currently rests on a single real casualty, which the document says plainly.

## Where we're going

Three things, in order of who they need.

**Needs a decision from you:** whether an unclassified part of the system should be allowed to
destroy a request silently while we work through the list. The proposal lays out the options,
including doing nothing to the mechanism and instead driving the list to zero.

**Needs a release:** one further change that makes the front door say out loud when it genuinely
cannot start a build, rather than reporting success. It is written and held back, because
applying it before the code ships would stop the front door working entirely.

**Needs other people:** fourteen steps in other teams' areas that nobody has classified. The new
check names them, which is the whole point — it turns "somebody should look at this" into a list
with names on it.

---

*Two footnotes I would rather write down than leave out. I made three measurement errors in one
day on this, all of the same shape: a result that could not have come out any other way. Each is
recorded where it happened. And twice my own guidance to the lane doing the live test would have
produced a wrong reading — once because my fix invalidated their test without my noticing, once
because my suggested timing would have collided with an unrelated item. Both times their own
care caught it, not mine.*

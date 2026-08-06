# Where we are — the token-cap watchdog (bugs_open/183)

## 2026-08-06, morning

Picked up bug 183. The story of that bug is short: every time we ask the AI to write
something, we set a maximum length. One step — the one that works out what a newly
adopted website is *about* — had a limit that was set a long time ago and never
looked at again. The documents it writes kept getting longer. On 2 August they
finally got too long, and the step started failing. It gets three tries per site, and
because the failure was the same every time, it used all three and gave up. Sites
were left with no strategic profile at all.

Someone already raised that limit, twice, so the immediate problem is gone. The bug
stayed open because **nothing was watching**. There are 126 of these steps. Any one of
them could be quietly growing towards its own limit right now, and we would only find
out the way we found out about this one: when it starts failing and somebody
investigates the wrong thing first — because a failure like this looks like a problem
with the *site*, not with a setting.

So I built the watchman rather than adjusting one more limit.

It runs every six hours, costs nothing (it is a single database query — no AI is
involved), and it looks at every step in the fleet, asking one question: how close is
this step running to its ceiling? If a step is near its limit, or has already hit it,
it writes a note. Crucially it only writes when the *situation changes*, so it does
not become one of those alerts everybody stops reading.

There was already a version of this for the council review seats, built a week ago.
Mine is the same idea widened to everything else, deliberately sharing its thresholds
so the two halves agree with each other.

## The bit I want to flag: it found something on its very first run

The top line of the first report was a step in the **vet practice verification**
pipeline. It had failed **every single call for the last 34 hours**, and no one had
noticed or filed anything.

Two separate things are wrong there, and I have written them up as a new bug (205).

First, that step **has no length limit set at all**. Not a bad one — none. When
that happens the system falls back to a built-in default of 2048, which is the
smallest number anywhere in the estate, and nothing in the configuration tells you
this is happening. I checked how widespread that is: **8 of our 126 steps have no
limit set**, and six of those are not currently in use — which means they are traps
waiting for the first time somebody uses them.

Second, and worse: the failures are **the same two records, over and over**. A batch
job re-checks unverified records every five minutes. These two produce a longer
document than the limit allows, so they fail; because they fail they never get marked
as verified; so five minutes later the batch picks them up again. It has been going
round that loop since Tuesday morning and would have continued indefinitely.

**I have not fixed either one.** Raising a limit on somebody else's pipeline has
always been your call here, and the vet work belongs to another lane — so it is
written up with the evidence and the options rather than quietly changed. The loop is
the part I would want a decision on soonest, because it is burning credits every five
minutes to fail in exactly the same way.

## What I got wrong along the way

Worth saying plainly, because two of them were the same kind of mistake.

My first measurement of the whole fleet was **wrong in a way that made things look
calmer than they were**. When a response gets cut off, the system does not record how
long it was — so if you filter on "responses we have a length for", you silently
throw away every single failure. I know this; it is written down in two places I have
read. I wrote the query that way anyway. It reported that vet step as *fine*. It is
the worst-affected step in the estate.

And the first version of my watchman **could not have caught the bug it was built
for**. It looked at the last two weeks; that step ran only three times in the two
weeks before it failed — too few to say anything about. I only found this by pinning
the clock to the day before the failure and asking what it would have said then. With
a three-month window it flags the problem a full day *before* the first failure.
That test — "would this have caught the thing it exists for?" — is the one I would
keep if I could keep only one.

## Where this leaves bug 183

One decision away from closing, and it is yours, not an investigation. The watchman
is live. The limit is raised with about five times the headroom it needs, and growth
back towards it will now be announced rather than discovered by a burned site.

The remaining idea in the bug file is to split that step into four smaller pieces so
it *cannot* outgrow its limit. That is a bigger change to a pipeline other people are
actively using. My view: not now, and the monitoring makes waiting safe — but if you
want the structural version, the bug should stay open and say so.

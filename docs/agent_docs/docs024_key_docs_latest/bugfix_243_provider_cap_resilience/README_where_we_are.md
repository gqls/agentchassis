# Where we are — the Anthropic cap, and what the platform does when it hits

Plain prose, append-only, newest at the bottom. The owner's running log.

## Sunday 24 August 2026 — picking this up

You asked me to look at `bugs_open/243`. That number names two completely different bugs,
which is a trap the guidelines warn about. One is about tool-acceptance losing its screenshots;
that one is owned by another thread which committed to it today, and all its questions are
already answered, so I left it alone. The other is the one about the Anthropic account hitting
its usage limit and the whole fleet's LLM work failing. That one says "open, unowned" at the
top, has no lane directory of its own, and six different threads have written into it over the
last two weeks without any of them taking it on. That is the one I picked up.

### The bug is still real, and it is happening more often than the file says

The file reads like a story of three incidents. The database says otherwise. Counting days on
which we actually got the "you have reached your specified API usage limits" refusal:
**seven of the last fifteen days.** And two of them were bad — on 22 August there were 113
refusals, which was very nearly every failure of any kind we had that day, and on 23 August
another 32.

The file is not wrong, exactly. It is a chronicle of the occasions when somebody happened to be
watching. Nobody was watching on the other days.

Right now, today, there is no cap. We have had 750 successful calls and no refusals. So I am
working on a system that looks perfectly healthy, which is the right time to do this and also
means I have to be careful: any fix I make will look fine today whether or not it works.

### The thing the file has been telling everyone to wait for has already happened

In four separate places the bug file says the real prevention is the other bug, 244 — the one
about the council review sending near-identical enormous prompts with no caching. "Fix that and
this bug's trigger largely goes away."

That fix landed two weeks ago and it worked. The expensive part of our prompt bill dropped
about sevenfold. But the total volume of prompting grew several times over in the same period,
so the actual spend only fell by roughly a third — and the two worst cap days in the entire
record came *after* the fix. So the prevention landed and the problem got worse. I have written
that into the bug file's own terms, because the next person to read it would otherwise sit and
wait for something that is already done.

I also think — though this is an inference from the shape of the message rather than something
I can read off a bill — that this is a **monthly** limit. The refusal always says access
returns on the 1st. That fits the pattern: the refusals cluster towards the end of the month as
spending accumulates. If that is right, we should expect this to bite again between now and the
31st, and then go quiet in early September regardless of anything we do.

### What actually goes wrong inside our system when the cap hits

This is the part worth your attention, because the cap itself is your problem to fix (you add
credit) but what our platform *does* about it is ours, and it does two things badly.

**First: one refused call stops the entire build queue for up to an hour.**

We keep a table with one row per AI endpoint saying whether it is healthy. When any LLM call
anywhere in the fleet gets refused, that row is set to "unhealthy". And every single piece of
queued work — every site, every page build — checks that row before it will start. If the row
says unhealthy, the work is put back in the queue untouched.

The problem is the asymmetry. A single failed call can mark the endpoint down, immediately.
But **nothing that succeeds can ever mark it back up.** I checked: there is exactly one place
in the whole codebase that writes to that row from live traffic, and it only ever writes
"unhealthy". The only thing that writes "healthy" is a background probe — and for Anthropic
that probe runs **once an hour**.

So the sequence is: one call gets refused, the whole fleet's queue stops, and it stays stopped
until the top of the next hour — even though calls are succeeding again seconds later. That
is not a theory; another thread measured it on 17 August, and it was exactly 60 minutes and 25
seconds of the entire fleet doing no queued work, while 93 out of 99 real calls in the same
window succeeded perfectly. Nothing alerted. Nothing was marked as failed. The only trace was
buried inside one field of one orchestration record.

**And there is a twist I found today that nobody had noticed.** That hourly probe cannot
actually detect this problem. The refusal arrives as an HTTP 400, and the probe's code treats
any 400 as "the API is reachable, so it's fine". So the probe is not really a health check at
all for this condition — it is just a timer that clears the flag after an hour, whether or not
the provider is still refusing us. This also means a conclusion recorded in the bug file on 17
August ("the probe recovered, therefore the cap was intermittent") does not follow: that probe
would have reported healthy either way. The conclusion may still be true for other reasons, but
that particular piece of evidence proves nothing, and it is sitting in the file looking like it
does.

**Second: one refused call throws away an entire council review.**

When you put a change through the review council, it fires about ten reviewers. Each one is a
separate LLM call. Every one of those seventeen reviewer slots is configured so that if its
call *errors*, the whole round is abandoned — including the opinions of the reviewers who had
already answered and been paid for.

The daft part is the asymmetry again. If a reviewer comes back with *garbage* — unreadable
output — the system handles that gracefully: it records the seat as unreadable, carries on, and
makes sure an unreadable seat can never be the difference between "approved" and "needs
revision". That machinery already exists and works well. But if a reviewer's call simply
*fails*, none of that applies and the round dies. **Garbage is survivable; a transient fault is
fatal.** A thread measured the cost on 19 August as roughly a coin-flip per round.

The reason for this, underneath, is that by the time the error reaches the part of the code that
decides what to do, it has been turned into a plain string. The error type that says "this was
the provider refusing us, not bad work" has been thrown away. There is even a comment in our
own code describing the consumer that was supposed to read that type — and that consumer was
never written.

### Where I am now

I have the mechanism nailed down and cited, and I have put it through the diagnosis loop as the
guidelines ask before asserting a cross-cutting cause. I also have a fix shape I am fairly
confident about, and one genuine complication I want to be honest about: the obvious quick fix
for the council problem (just let a failed reviewer be skipped) would quietly turn a reviewer we
never heard from into a reviewer who *didn't object* — which could let something get approved
that shouldn't. Our own code comments warn against exactly that confusion. So the fix has to
record a failed reviewer as "we were owed this opinion and lost it", which the system already
knows how to handle, rather than as "not applicable".

Next: a plan, ranked by what actually closes the door rather than what is quickest, then the
council, then commit.

## Sunday 24 August 2026, later — reviewed, approved, and two thirds of it shipped

The council approved it first time, with four advisory comments. Two of them were right in a
way that changed the code, and I want to record that plainly because it is the argument for
bothering with the review at all.

The **guardian** seat pointed out that the place I was adding the new failure record is not the
council's code — it is the shared error path that *every* workflow in the estate goes through.
It was right: a workflow that loops and keeps failing would have grown that record without
limit, on exactly the runs that are already going badly. So it is now capped, with a marker
saying when the cap was hit, so nobody can mistake "we stopped recording" for "nothing failed".

The **reuse** seat asked why I was adding a second key beside the existing one instead of just
widening the existing one — which would be tidier — and suggested it had "few readers". So I
counted: **39** things read it, 33 in code and 6 in live agent configuration. That settles it,
and it settles it with a number rather than my preference, which is the useful part.

A third comment claimed a gap that turned out not to exist — it said work items would still
burn through their retry budget on a refused call. They do not; that was already handled
elsewhere. I checked rather than argued.

### What has actually shipped, and what has not

**Shipped:** the change that lets a successful call clear the "endpoint is unhealthy" flag, and
the change that makes the council record a reviewer whose call *failed* as an opinion we lost
rather than as one that was never wanted. Both are committed. Neither is live yet — our Go code
only takes effect when a new image rolls out.

**Not shipped, and this is the annoying one.** The third piece is about twenty lines in a file
called `coordinator.go`. It is written and it is correct. I have not committed it, because
another session is part-way through their own change in that same file — and the function their
change calls lives in a file they have not committed at all. Because of how commits work on this
shared tree, if I commit that file I take their half-finished work with it, without the piece
that makes it compile, and I break the build for everyone.

So I have left it, told them, and offered two ways to clear it. The piece I *did* ship is
written to do nothing at all until that lands, so there is no half-working state — and the
database change that goes with it is deliberately held back too.

### One thing I would like your steer on

Part of the fix is a single database setting: how often we re-check whether Anthropic is
answering. It is currently **once an hour**, and that hour is exactly how long the whole fleet's
work queue can sit stopped after one refused call. Changing it to **once a minute** is one
statement, instantly reversible, and it is the value we already use for our other endpoints.

It does not depend on the blocked piece above. I have not applied it, because it changes live
production settings across the fleet and you did not ask me to do that. Given the refusals
cluster towards the end of the month and we are a week out, **say the word and I will apply it**
— or tell me to leave it and it goes in with the rest.

### Also worth knowing

While reading the code I found that the hourly check **cannot actually detect this problem**.
The refusal comes back as an HTTP 400, and the checking code treats any 400 as "the service is
reachable, all fine". So it is not really a health check for this; it is a timer that clears the
flag after an hour whether or not we are still being refused. That matters because the fix
everyone was about to build — making that check less trigger-happy — would have aimed at a piece
of code that never fires for this problem, shipped, and looked like a fix. I nearly did it
myself. It is written up as my own mistake in the shared log, because I repeated the claim from
another thread's notes without reading the function.

## Sunday 24 August 2026, evening — the interval is in, and the main fix is proven working

You asked for the shorter probe interval, so that is done. It was bundled inside the held-back
database change, so I split it out into its own file, applied it by hand, checked it, and
recorded it properly. The re-check now happens every minute or two instead of once an hour.

**One correction, made before it could get quoted at anyone.** I had written that this brings
the worst case down to "about a minute". It does not. The check only happens when a background
task ticks *and* the endpoint's own timer has elapsed, and that task also runs on a minute — so
the two compose. I measured the actual gaps: 94 and 92 seconds. So it is one to two minutes.
Still about forty times better than an hour, which is the point, but "one minute" is the kind of
round number that gets repeated for years and it is not what the system does.

**The main fix is live and I have proven it works on real traffic.** The new build carries it —
I checked the actual running binary on both machines rather than trusting the version number,
with two control checks to make sure the test could tell the difference.

Then I proved the behaviour: I marked the endpoint unhealthy by hand, and a single real
successful call cleared it within about forty seconds. I can attribute that to our new code
rather than to the old hourly check with certainty, because our code updates one timestamp and
not the other — and the two came out sixty-one seconds apart, which nothing else in the system
can produce. It then did it again on its own a couple of minutes later, healing a genuine blip
in about eleven seconds. **So the hour-long fleet stall this bug was really about is now closed
at the cause, not just shortened.**

### I have to tell you about a bad half-hour in the middle

My first attempt to prove it reported failure. I spent a good while diagnosing my own
freshly-shipped code — reading the insertion point, checking whether those agents take a
different route, checking whether the pods were running an older image.

**The fault was in my test.** The database prints a true/false column one way on its own and a
different way when you join it onto other text, and I was comparing against the wrong one. My
check could never have reported success, for any state of the system. Two further attempts
"failed" for an unrelated reason: there simply were no calls happening in those windows, so
there was nothing for a fix that reacts to successful calls to react to.

Three failures, two different causes, none of them about the code. Three consistent negatives
felt like confirmation; they were one broken instrument and two empty rooms. **I was one commit
from writing "this does not work" into the bug file — about my own code, on the day it shipped.**
What caught it was noticing an impossible pair of timestamps in a state dump.

It is written up in the shared mistakes log, and both checks now sit inside the experiment
rather than in my memory.

### Still outstanding

The third piece is still blocked on the other thread's uncommitted file, exactly as this
morning — I re-checked, it has not landed. So the database change that depends on it stays held
back, and I have confirmed by inspecting the running binary that the piece it waits for is
genuinely not there. Nothing is half-applied.

And the two big questions remain yours, unchanged: topping up is still the only thing that
restores service under a real cap, and whether we add a second AI provider is still undecided —
every one of our 127 configured AI steps points at the same account and the same key.

## Sunday 24 August 2026, night — unblocked, and all three pieces are now committed

You approved the sweep, so that is done and the blockage is gone.

Before touching another thread's work I checked it was genuinely parked rather than in
progress: both files had sat untouched since Friday evening, two days, with nothing committed
for that thread since. And before committing code I did not write, I made sure it was safe —
it builds, its own tests all pass, and I read it properly. It cannot do anything at all on our
live system yet, which is what made it tolerable to land.

I did it as two commits so the build never passes through a broken state, and both messages say
plainly which work is theirs.

**A useful thing fell out of it.** Their notes contain a table that reads as though their fix
had been measured working against our live configuration. It has not — the setting their code
keys on does not appear anywhere in any live agent, which I checked in all three places it
could have been hiding. So their bug is not actually fixed: the code is there and nothing can
reach it. I have told them, and passed on how we handled the same situation here.

### Where the whole job stands

- The **main fix** — a successful call clearing the unhealthy flag — is live and proven twice on
  real traffic.
- The **probe interval** is applied, at a measured 92–94 seconds rather than the hour it was.
- The **council fix** is now fully committed, but it is Go, so it does nothing until the next
  build goes out. The database change that finishes it stays held back until then, and I have
  written the exact check into that file so whoever applies it cannot get it wrong.

The last thing owed is a proof I cannot force: a council round where one reviewer's call fails,
which should now cost that one reviewer instead of the whole round. That will happen on its own.

Your two decisions remain open and unchanged: topping up is still the only thing that restores
service under a real cap, and whether we add a second AI provider is still undecided.

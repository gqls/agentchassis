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

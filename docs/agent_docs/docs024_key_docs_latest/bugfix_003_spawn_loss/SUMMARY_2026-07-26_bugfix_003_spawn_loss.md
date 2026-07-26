# Bug 003 — spawned children lose their response. CLOSED, 2026-07-26

## What we were trying to do

Stop work disappearing. A parent job would hand a task to a child agent, wait
for the answer, and never get one — even when the child had actually done the
work. The parent then sat there doing nothing until a slow database sweep gave
up on it, thirty to ninety minutes later. It hit page builds, image generation,
business-intelligence ingestion, and — worst of it — the diagnosis system we use
to investigate everything else, so the tool we reach for when something is wrong
was itself being eaten by the thing that was wrong.

## Where we came from

This was filed on 15 July and took eleven days and five separate threads to
finish, because it was never one bug. It was four, and each one was only visible
after the one in front of it was fixed.

The first looked like a network fault to one specific Kafka server. That turned
out to be real but not the story — the pattern moved around and became
intermittent, so it was split off to its own case. The second was that our
message loop told Kafka "handled" *before* handling the message, so any pod
restart destroyed whatever was in flight — and the de-duplication layer had the
identical flaw one level up, meaning a fix to either alone would have done
nothing. The third was that our timeouts were kept alive by a sleeping timer
inside a single process; when that process restarted, every pending timer
vanished and nothing rebuilt them. The three compounded: a restart lost the
message, lost the timer that would have noticed, and left a job nothing would
ever clean up.

The fixes for the second and third went live on 25 July in a slightly untidy
way — our own review council had vetoed the roll twice on the grounds that a
change this size belonged in an architecture review rather than a point-fix
gate, and while that was being sorted, another team's routine build carried the
code to production anyway. Shown the evidence of it working, the owner ratified
keeping it live, and we created the architecture-review track the council had
asked for. **No part of this work carries a review-approved stamp, and the
records say so plainly** — it was never approved, it was ratified after the fact.

## What we've done

Today we set out to close the case and found it was not closeable, which is the
useful part of the story.

Checking how the retry system was actually behaving turned up a fourth fault
hiding inside our own fix. When a response arrives, the code first "claims" it,
so two servers can't process the same answer twice. Only afterwards does it do
things that can fail. If any of those failed, the claim was never given back —
and nothing anywhere knew how to release one. The row sat in a state no cleanup
job had ever been taught about, invisible to every recovery path we had. The
parent job waited for ever.

There were **181** of these, the oldest a month old. Two of them had a live job
still waiting: one was a web page that would never have been published.

The fix went into the cleanup routine that already runs every minute, so it was
live within minutes and needed no software release at all. The stuck count went
from 181 to 8, and the 8 are simply responses being handled right now. The
stranded page-publish job was picked up by a completely different service and
finally given an honest answer.

We also re-ran the health-check test that failed embarrassingly on 20 July, when
a deliberately broken server cheerfully reported itself healthy for six minutes.
This time it reported itself unhealthy and **the platform restarted it**, which
is the behaviour we had claimed for six days without ever proving. That also let
us delete a piece of planned work — we were going to add a way for a server to
kill itself when stuck, as insurance against the health check being wrong.
The health check is not wrong. Two mechanisms for one job is worse than one that
works.

## Where we are now

The symptom is gone, measured rather than asserted. In the day and a half before
the fix, jobs died waiting for an answer at **2.34 an hour**. In the day and a
half after, **0.38 an hour**. On 26 July, across 1,114 completed jobs, **none**.
Thirty jobs recovered by themselves that would previously have been silent
losses.

The case is closed and moved to the closed folder. What remains genuinely
belongs elsewhere and is written down where it belongs: the original network
flakiness has its own case; a related fault about servers refusing to take over
each other's work is being fixed by another thread right now; running more than
one copy of the main server is still blocked on a separate design question.

## Where we're going

One date to keep: **1 August**, a week after the main fix went live, to confirm
the numbers hold rather than having caught a quiet weekend. The query is in the
runbook.

Beyond that, this case leaves the platform two habits rather than just a fix.
First, whenever a status exists that a row can sit in, somebody has to own
moving it out — and the question "which sweeper owns this state?" would have
found today's fault a month ago in about a minute. Second, a check that can only
return one answer is not a check: our verification of the sister service passed
today for the entirely wrong reason, and only a deliberate positive control
exposed it. Both are written into the debugging guide, which is the only way
they outlive the people who learned them.

# Where we are — the dispatch queue that makes everything feel broken

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-20 — picking up bug 030

The complaint behind this one is familiar to anyone who has fired a trigger at the
cluster: you run the script, it prints an id and exits happily, and then nothing
happens. No orchestration row, no log line, nothing for half an hour. Every
instinct says the message was thrown away. It wasn't — it's sitting in a queue.

That was filed yesterday as bug 030. My job today was to check it still holds and
work out what to actually do about it.

**The topology facts still hold.** One Kafka topic with a single partition, and —
this is the part I think matters more than the bug file gives it credit for — a
single chassis pod, with no autoscaler. So there is exactly one thing in the whole
system reading that queue.

**I nearly got this badly wrong, and I want that on the record.** For the first few
minutes I watched the queue position sit completely still while the backlog grew,
and I found no sign of any message being picked up in a thirty-minute window of
logs. I was drafting the sentence "consumption has stopped, there's about nineteen
hours of backlog" when I remembered the bug file has a warning about precisely this
— another session made the same call earlier this week and was wrong. So I left the
measurement running instead of concluding. Ten minutes later the queue lurched
forward by forty-seven messages. Nothing was broken. It just moves in fits.

The lesson is worth more than the measurement: **on this queue, a frozen position
tells you nothing about whether anything is alive.** You have to watch it for
twenty minutes before you're allowed an opinion.

**What is actually going on.** Once I let it run properly the shape was clear: the
queue sticks at one position for about eight minutes, then jumps. Then sticks
again. So one single message is holding everything else up for eight minutes at a
time.

Reading the code explains why. When the chassis picks up a message it handles it
there and then, on the spot, and doesn't go back for another one until it's
completely finished. And "completely finished" can mean a lot — the part that runs
a workflow will keep marching through step after step in that same breath, and each
step that calls an AI model takes about half a minute. So a workflow with a dozen
such steps parks the only reader in the system for six minutes, and everyone else's
work simply waits.

**Which changes what we should do about it.** The bug file suggests splitting the
Kafka topic into more partitions. On its own that would achieve nothing, because
there's only one pod to read them — one reader would just be handed all the pieces
and still work through them one at a time. That's worth being careful about,
because splitting a topic is a **one-way door**: Kafka won't let you merge them back.

I don't want to assert that from my own reading alone, because a confident wrong
answer here would send whoever fixes it down the wrong path. So I've put the claim
through the diagnosis loop to have it checked properly. It's queued — behind this
very backlog, which is a fittingly on-the-nose demonstration.

**Where I'd go first, whatever the verdict.** The delay isn't really the problem.
Waiting half an hour is annoying; not being able to tell waiting from broken is
what actually costs us. Both of the incidents recorded in the bug file were
misdiagnoses — one session paid for a duplicate run, another abandoned an
investigation entirely — and neither would have happened if the trigger script had
simply said "you're queued behind 41 messages, expect about 25 minutes" on the way
out. That's a small change, it's not a one-way door, and it fixes the expensive
part. The throughput question can then be decided separately and carefully.

**One other thing I found while reading.** There's a comment in the Kafka consumer
saying it commits the message position "after successful processing". It doesn't —
it commits immediately on pickup, before the work is done. That means a message
being processed when the pod dies is simply lost. This isn't news, it's already
recorded as one of the causes in the spawn-loss workstream, but the bug file's own
troubleshooting note explains the frozen-queue symptom using the wrong mechanism,
so I've corrected that where it matters.

## 2026-07-20 (later) — two of us got the same measurement wrong, in opposite directions

A twist worth writing down. While I was measuring the queue, another session was
measuring it too, and published a figure into the bug file: the queue drains at
about a fifth of a message a minute, so a submission waits roughly six and a half
hours.

That number is wrong, and I could show it precisely — because their two readings
turned out to be two of *my* samples. I had a continuous log running at
thirty-second intervals, and both of their figures appear in it, sixty-nine seconds
apart. They had been labelled fourteen minutes apart. The real drain rate is about
2.4 messages a minute, which means the queue clears in around half an hour — which
is exactly what the bug said in the first place, the day before. Nothing had got
twelve times worse overnight.

What makes this worth more than a correction is that **I had made the same mistake
two hours earlier, in the opposite direction** — I nearly declared the cluster dead
when it was merely mid-pause. Two sessions, same week, same queue, opposite wrong
answers. That stops being a story about either of us and becomes a story about the
queue: it moves in fits, so any measurement shorter than one full stop-start cycle
tells you something confident and false. I've written the rule down with the exact
command to use, and logged both errors in the shared list of wrong calls.

I also want to be fair about what the other session got right, because it's the
more important half: they worked out *why* the queue behaves this way — that each
message is held until its work reaches a natural pause — purely from the timings.
I reached the same conclusion from reading the code. Two independent routes to the
same mechanism is the strongest evidence we have, and I've said so in the bug file
rather than leaving a correction that reads like a rebuke.

The honest summary of both our errors: we each had the source code available, and
we each used the stopwatch instead. Ten minutes in two files explains the whole
behaviour.

**Where things stand:** the root-cause claim is with the diagnosis loop to be
checked properly, and it's queued — behind the backlog it's about, which is either
poetic or annoying depending on the hour. You've said wait for that verdict before
building anything, so nothing is being changed in the meantime.

## 2026-07-20 (later still) — and then I made the same mistake myself

I have to correct the account I gave above, because the tidy version I wrote — where
I spotted another session's error and put the number right — is not what happened.

My replacement figure was wrong too. I said the queue drains at about 2.4 messages a
minute and clears in half an hour, and that nothing had really degraded. I had a
twenty-minute measurement running when I wrote that, and I used the first half of it.
When it finished, it said the opposite: over the full run the queue was draining at
about *0.6* messages a minute, the backlog grew from 82 to 130 while I watched, and
at one point a single message held everything up for more than fifteen minutes. The
queue wasn't holding steady. It was falling behind.

So the other session's conclusion — that this is slow, that the variation is huge,
and that there's no dependable answer to "how long will I wait" — was closer to the
truth than mine. Their *working* was genuinely faulty and they've since owned that.
But I took "their method was wrong" as licence to overturn what they'd concluded, and
those are two different things. That's the part I'd most want to avoid repeating.

The useful thing to come out of it is bigger than the correction. Three of us have now
measured this queue on the same afternoon and got 0.21, 2.4 and 0.62 messages a
minute. All three sums are correct. When three careful measurements disagree by
twelvefold, the problem isn't the measuring — **it's that the thing being measured
doesn't exist as a stable quantity.** The speed of this queue is just however long
the job currently at the front happens to take, and that ranges from instant to a
quarter of an hour. An average of that describes no actual moment and can't forecast
anything.

Which means the advice I confidently wrote down earlier — "measure it for twenty
minutes and you'll get the right answer" — was also wrong, and I've withdrawn it. A
longer measurement gives you a steadier number, not a truer one. I've replaced it
with the honest version: you can find out whether your job is queued (that's the
question that actually costs us time, and it's easy), and you can see what kind of
work is in front of it. If someone wants an ETA, the right answer is that there
isn't a reliable one, and the variability *is* the finding.

Worth noting all three of us could have read the code instead. It explains the whole
behaviour in ten minutes and would have told us up front that there was no steady
rate to go looking for. We each reached for the stopwatch with the source open.

## 2026-07-20 (evening) — it turns out the queue is almost entirely the robot's own housekeeping

After three of us had spent the afternoon arguing about how *fast* the queue empties,
it occurred to me that nobody had looked at what was actually in it. That turned out
to be the useful question.

Ninety-three per cent of the messages in the queue are put there by our own
scheduler, not by anyone working. Two routine jobs — a health check on the AI
endpoints, and a trigger that looks for pages to build — account for eighty-four per
cent of everything in the lane on their own. The council reviews and diagnosis runs
that this whole bug is *about* make up about six per cent. They are queuing behind
the housekeeping.

And the housekeeping is winning. Over seventy minutes I watched, the scheduler put
messages in at about 2.6 a minute and the system took them out at about 1.4 a
minute. The backlog grew from 82 to 164, which is exactly the difference — the
arithmetic balances to the penny. So this isn't sessions getting in each other's way,
which is how the bug was originally framed. It's two routine chores running more
often than the system can service them, and everything else waiting behind.

That's much better news than it sounds, because the schedule lives in a database
column rather than in code — it can be changed immediately, without a rebuild.

**One warning I've written up carefully, because it's the kind of thing that gets
"tidied".** Both those jobs are configured to run every 30 seconds, but they actually
run every 60. The reason is a hair's-breadth timing quirk: the scheduler wakes up
every 30 seconds too, and when it wakes at the exact moment a job becomes due, it
checks a fraction of a second early, decides the job isn't due yet, and picks it up
on the next pass. So we're getting half the configured rate — and if someone spots
that discrepancy and "corrects" it, the load doubles overnight and the backlog goes
from growing slowly to growing fast.

I also checked something the bug file asserts: that everything in the system funnels
through this one lane. It doesn't. Each spawned worker gets its own private queue and
its own pod — there are over 800 such queues. Only the top-level "start this job"
messages go through the single lane. That makes the problem narrower and the fix
smaller than the file implies.

**One awkward practical note.** The diagnosis run I submitted to check my own
reasoning has now been waiting fifty-five minutes and still hasn't started, because
it's stuck in exactly the queue it was sent to investigate. And since the queue is
currently growing rather than shrinking, I can't honestly tell you when it will run
— it'll go when the job at the front happens to be a quick one. So "wait for the
verdict before doing anything" may not be a workable plan today, and you may want to
decide without it.

## 2026-07-21 — made the small fix, and took a snapshot

You asked for two things: a milestone snapshot, and the config-only change to make the
scheduled jobs work together properly. Both done.

The snapshot is a new summary document (I never overwrite the last one — the series is
the record), written in plain prose to be read aloud: what the bug is, the rather
embarrassing measurement history, the finding that the queue is almost all our own
housekeeping, and the change I've just made.

The change itself: I slowed the two jobs that dominate the queue, and — just as
importantly — I set them to intervals that don't trip the timing trap I found
yesterday. The health check goes to 60 seconds, which is what it was actually doing
all along; that one's really about making the configuration honest and removing the
trap, so nobody later "fixes" it and doubles the load. The build trigger goes to 120
seconds, and that's the one that should actually help, because it halves how often the
big, slow build jobs get kicked off — and those are what jam up the single worker for
minutes at a stretch.

It's a database change, so it took effect the moment I made it — no rebuild, no
deploy — and it's a one-line command to put back if we don't like it. I grounded the
numbers first: I checked what the health check actually needs (the endpoints it pings
want checking every 60 seconds at most, except one that's already broken), so 60
seconds doesn't starve anything real.

I want to be careful about what I'm claiming. This reduces how much work the scheduler
pours into the queue. Whether that's enough to get the queue *shrinking* rather than
growing, I don't yet know — I've got a check running, but an honest answer needs
twenty minutes of watching, and the queue was still growing when I made the change. So
this is "a sensible first lever pulled", not "fixed". The bigger structural question —
whether to give the scheduler its own separate lane so it can never queue in front of
your work again — I've left for a separate, careful decision, and it's written up.

And the diagnosis run I sent yesterday to double-check my reasoning is still stuck in
the queue, more than an hour on. So if it never surfaces, we decide the structural
part on the evidence we have, which is already fairly strong.

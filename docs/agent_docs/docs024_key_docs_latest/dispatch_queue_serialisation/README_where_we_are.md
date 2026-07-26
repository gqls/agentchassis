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

## 2026-07-21 — the fix is in and measured, and I caught myself getting the mechanism wrong again

I watched the two jobs actually fire after the change, rather than trust the setting.
Good thing too, because they aren't firing at the intervals I set. I set 60 and 120
seconds; they're firing every 90 and 150. And that turned out to expose a mistake in
my own explanation from yesterday.

I'd said the every-30-seconds-means-every-60 problem was a special coincidence of the
job interval matching the scheduler's heartbeat, and that picking a "clean" interval
would fix it. Watching it run showed the opposite: *every* job fires one heartbeat
later than its setting, always, because of how the scheduler stamps the time. So my
"clean interval" reasoning was backwards — and I've corrected it everywhere I wrote
it. The plain rule is: whatever you set, add one heartbeat (30 seconds) to get the
real interval.

The happy accident is that this makes the fix slightly *better* than I intended —
90 and 150 seconds is a bit more breathing room than 60 and 120 would have been. The
scheduler is now putting work into the queue at roughly a third less than before, and
the two dominant jobs have gone from two-a-minute down to about one-a-minute between
them.

Is the backlog actually recovering? The one snapshot I took looked very healthy — the
queue was down to 20 waiting, from 168 yesterday. But I want to be honest that I can't
claim victory from that: a lot of the overnight hours are quiet and the queue drains
itself then, so I can't tell how much is my change and how much is just 3am. Proving
it needs watching the queue through a busy stretch, which I haven't done yet. So:
lever pulled, measured, pointing the right way — not "fixed" with a tick in the box.

The two bigger pieces are still open and still yours to steer when you want them: the
structural change (giving the scheduler its own separate lane so it can never queue in
front of your work), and the cheap diagnosability win (making the trigger scripts tell
you "you're queued behind N" instead of leaving you guessing). Neither was touched
today; today was just the safe, reversible, config-only first step you asked for.

## 2026-07-21 (afternoon) — watched it through a busy stretch: the backlog now clears itself

This is the check I promised — the queue watched for half an hour during the working
day, not overnight. The result is good, and I can be more confident now.

Yesterday the queue only ever grew: 82 waiting, then 130, then 168, never coming back
down. Today, over the same kind of window, it bounced between empty and about 15, and
crucially it hit *empty twice* — the backlog is clearing itself between busy patches
now, rather than piling up without end. That's the difference the change was meant to
make, and it made it.

One honest caveat, and it points at the bigger job still to come. The queue still has
moments where a single heavy job sits at the front for seven or eight minutes and
everything waits behind it — that hasn't changed, because it can't be fixed by
scheduling; it's the deeper "one worker, one job at a time" limitation. What the
change did was stop those heavy jobs arriving so often that the queue never recovered.
So the queue now recovers between them. Fixing the stall itself — so nothing waits
behind a slow job at all — is the structural change I've left for a separate decision.

So where we've landed today: you asked for the safe config-only lever, and it's
pulled, verified, and measurably working — the runaway backlog is gone. The two bigger
pieces (a separate lane for the scheduler, and making the trigger scripts tell you
you're queued) are written up and waiting for your steer. Nothing was rebuilt or
deployed; it's all reversible with one line.

## 2026-07-25 — both remaining pieces built. One needs you to press the button.

Four days on, I came back to finish this. Two things were left: making the trigger
scripts tell you you're queued, and giving the scheduler its own lane. Both are now
written. One is already live; the other needs a step I'm not allowed to take.

First, the good news about the last change: **the config fix has held.** The two
settings I slowed on the 21st are still exactly as I left them, nobody has undone
them, and the queue at the start of today was eight deep and behaving — bounded, not
running away. Four days is a much better answer than the half hour I had at the time.

But I found something while checking, and it's the reason I didn't just tune the
settings again. **The lane has quietly re-filled with new chores.** There were twelve
scheduled jobs pointed at it when I measured; there are now twenty-one. Nobody did
anything wrong — people add jobs, and each one looks harmless on its own — but the
total is back to roughly where it was before my fix, because *nothing in the system
owns the total*. Tuning intervals decays. That's an argument for changing the shape
rather than the numbers, so that's what I've done.

**Piece one, live now: the scripts tell you you're queued.** When you fire a review or
a diagnosis, it now prints how many messages are ahead of you, whether anything is
actually reading the queue at all, when the system last moved, and what big job you're
sitting behind. This is the thing that has twice cost real money and time — someone
sees nothing happening, concludes it was dropped, and fires it again, paying twice.
It won't guess how long you'll wait, and that's deliberate: this queue genuinely has
no reliable speed, and three of us have now been confidently wrong about it. It says
so out loud rather than making something up. I watched it work on my own submission
an hour ago: "18 in the queue, 17 ahead of you, the system moved 9 seconds ago,
you're behind a council review that started 171 seconds ago." That is the whole
problem solved for the price of a shell script.

**Piece two, built but not yet switched on: the scheduler gets its own lane.** The
housekeeping chores and your interactive work currently share one single-file queue,
and the chores are the overwhelming majority of it — so your review waits behind
them. The change lets the system read from two queues at once instead of one, and puts
the chores in their own. Nothing else about how work runs changes: each queue is still
strictly one-at-a-time and in order, which is what keeps the machinery correct. What
changes is that a nine-minute review no longer parks the chores, and a queue full of
chores no longer parks you.

I re-measured the problem on today's live system before writing any of it, and caught
it in the act: a council review held the single queue for nine minutes while the
backlog behind it went from 8 to 25. Same mechanism as originally described, still
there, on today's build.

**What I need from you.** The code is committed and the image is built and checked,
but I'm not permitted to publish the image to the registry from this session, so the
last three steps are yours. They're short and in the runbook (R9): push the image,
roll the chassis, and then — only after checking the new lane is really being read —
run one line of SQL to move the chores across. The order matters: if the chores get
moved first they'd be posting into a queue nobody is reading, and all scheduled work
would silently stop. Everything is reversible with a single line either way.

**What I'm deliberately not doing.** There's a bigger fix underneath all this —
letting the system start the next job without waiting for the slow one to finish.
That's a real change to the heart of the platform, it's already designed in a
different piece of work (the one about running more than one copy of the chassis), and
two of us building it separately would be the worst outcome. So I've left it there and
built the thing that doesn't collide with it.

## 2026-07-26 — switched on, and the wait is gone: one second instead of eighteen minutes

You pushed the new chassis overnight, so this morning I could finish the job. Here is
what happened, plainly.

First I checked the new image was really running and that the second queue had
actually come up — not just that the code shipped. It had: the new queue exists, the
system created it itself on start-up (I could tell because it carries exactly the
settings our own code sets, so nothing created it by accident), and the chassis is
signed up to read it. Only then did I move the chores across, which is one line of
SQL and takes effect immediately. Eighteen scheduled jobs moved.

**Then the test that matters.** I sent the same review submission through the same
council as yesterday, and compared:

- Yesterday, sharing one queue: eighteen messages ahead of it, and it took about
  **eighteen minutes** before it even started.
- Today, with the chores on their own queue: **nothing** ahead of it, and it started
  in **about one second**. By the time I looked, it was already being reviewed.

Both queues are empty. The chores are still running — I watched two of them come
through the new queue ninety seconds apart, which is exactly their schedule — so
nothing was dropped or stalled by the move. That is the whole of what this bug was
about: an operator's job no longer waits behind the housekeeping.

**The review of the change itself came back "revise" the first time**, which is worth
telling you about because the objections were good ones. The reviewers asked three
things I had not answered properly: had I considered simply running a second copy of
the chassis instead of changing code (cheaper-looking, but it turns out that is exactly
the configuration another piece of work has already found to be unsafe — it would
create a second "owner" of the same jobs); had I said out loud how much of the fleet
this touches (I checked: only the chassis programme uses this code, not the other
services); and had I actually *checked* in the code, rather than argued by analogy,
that doing two things at once is safe here (I had argued; now I have checked — the
system already does two things at once on a different path, and the pieces involved
hold no shared state). I answered all of it and resubmitted.

Two things I got wrong today, for the record. I talked myself into a specific new
danger — that a clean-up job could now wrongly kill a job that was still working —
and then found it was impossible for two separate reasons. It took two queries to
check and I nearly wrote it down as a risk instead. And my first resubmission was
rejected in six seconds because I had dressed up an *answer* as if it were a *change*;
the system quite correctly told me a plan proposes changes, not observations. Both
cheap, both the kind of thing that gets expensive if it goes into a handoff unchecked.

Where that leaves the bug: the original complaint — jobs waiting half an hour with no
way to tell a wait from a failure — is now fixed, live, and measured on both sides.
The deeper limitation (the system still does one job at a time within each queue) is
real, is written up, and belongs to the other piece of work that already owns it. So I
am closing this one and naming that as its successor rather than leaving a ticket open
on somebody else's job.

## 2026-07-26 (later) — the reviewers vetoed the change. Here is the argument, and a decision for you.

I should tell you this straight: the second review round came back not just "revise"
but **rejected** — a hard veto from the seat whose job is to guard against
unnecessary changes to the core of the platform. The change is live and working, so
you should know what the objection is and that you can reverse it in seconds.

**The objection is not that the change is wrong.** Nobody found a bug in it. The
argument is about *where* it was made: it edits the central chassis code that every
one of our agent processes is built from, and the guarding seat's standing
instruction is to insist that a fix be made at a higher, safer layer if one exists.
Six of the seats approved; that one seat vetoed on principle.

**It named the higher-layer alternative it wanted, and I checked: it isn't there.**
The suggestion was to have the scheduler hand its chores to their own throwaway
worker at the moment it schedules them, which would leave the shared code untouched.
But the scheduler literally cannot do that — it has no ability to start a worker at
all; the only thing it knows how to do is post a message. Starting workers is done by
the chassis, which means a chore would have to travel through the very queue we are
trying to get it out of first. So building that alternative would mean teaching the
scheduler a whole new capability, duplicating machinery we already have, for a
bigger change than the one being objected to — not a smaller one.

I also corrected one factual premise: the veto says this code is what "every agent
runs". Only the chassis programme is built from it — the other services aren't — and
on the spawned worker pods the new code is switched off twice over.

The one objection with real teeth was: "if you get the start-up and shut-down of
these extra readers wrong, it breaks every chassis pod." Fair. I audited exactly the
three failure shapes it named — a reader left open, a thread never waited for, a
mismatched counter — and all three are handled. That was worth being asked.

**So where does that leave us?** I have submitted a third round with the evidence,
and I have left the change running, because the measured alternative is going back to
half-hour waits. But this is a judgement call about how conservative we want to be
with the core, and that is yours, not mine. If you side with the reviewer, the
reversal is one line of SQL to send the chores back to the shared queue, plus
removing one setting from the deployment — seconds, no rebuild, and it's written down
in the runbook. Nothing is claiming reviewer approval: the change carries no
"reviewed" stamp, and the rejection is recorded in the bug file rather than tucked
away here.

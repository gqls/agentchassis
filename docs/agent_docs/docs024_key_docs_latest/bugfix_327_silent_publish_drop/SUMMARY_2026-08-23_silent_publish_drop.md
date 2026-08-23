# Summary — the build trigger that could start nothing and say it had

2026-08-23. Written to be read aloud.

---

## What we're trying to do

Make it impossible for someone to launch a new site build, be told it worked, and be wrong.

## Where we've come from

There is one command that starts a whole site build for a new domain — the thing we actually
sell. Five days ago a lane discovered that sometimes you run it, it prints all the right
reference numbers, it finishes cleanly, and **nothing happens at all.** No build. Nothing
queued. No error anywhere. One submission in three disappeared that way.

The cause was already known and written down. The command sends its message by starting a
throwaway container and feeding the message in through the container's input. Those two things
race. If the container starts first it sees an empty input, decides there is nothing to send,
exits successfully, and is then deleted — taking the evidence with it.

What made this worth a project rather than a one-line patch was the scale. **218 scripts in
this repo send messages this way. Twenty-five of them had the documented fix applied — they
ask the sender to shout "PUBLISH_OK" when it has genuinely sent something. Only two actually
listen for that shout.** The other twenty-three print it and carry on regardless. The remedy
had been written down for a month and was being copied without being made to work.

## What we've done

We built the remedy as something you can call rather than something you copy. One shared
publisher, used by the build trigger. It puts the message in the container's start-up command
instead of its input, so the race cannot happen at all; it refuses a message that would be
split into fragments; and — the part that matters — **it insists on hearing the confirmation
and fails loudly if it doesn't.**

It also answers a question nobody could answer before. When a build doesn't appear, there are
two possible reasons, and the right response to each is the opposite of the other: *the
message never left* (send it again immediately) or *the message left but nothing picked it up*
(wait — sending again just creates a duplicate). Those looked identical. They now produce
different exit codes and different printed instructions, and the tool checks the rejection
records too, because a rejected message leaves exactly the same silence as a lost one.

Alongside it, a check that runs on every commit and warns anyone adding a new script with the
old racing pattern. We measured it on 300 real commits before switching it on: it fired five
times, and all five were genuine.

## Where we are now

The build trigger is fixed and proven. Point it at a broken address and it exits with an
error, names what didn't go out, and gives you a working retry command — and it no longer
prints the reassuring "save this reference number" line, which used to appear *before* it had
even tried to send anything.

Three things we want to say plainly rather than bury.

**We could not reproduce the original failure.** On the day we tested, ten out of ten old-style
sends arrived. That rules out the four-in-five loss rate seen last month, but it only tells us
today's rate is under about a quarter. The fault is load-dependent. The new approach doesn't
beat the race so much as sidestep it entirely, which is the stronger position, but we should
not claim a head-to-head win we didn't observe.

**We found a second fault nobody predicted.** In that same test the old method sent one message
*twice*. On the real system a duplicate submission means two builds. That was one observation,
not a measured rate, and we've written it down as such.

**The fix itself could not be reviewed.** We put it to the review council and it declined —
correctly — because the code lives in a directory the council doesn't cover. We recorded that
rather than overriding it. In compensation the tool has its own self-test and every behaviour
was proven against live data.

One irony worth telling: the tool that checks our own landmine documentation turned out to use
the very pattern we were fixing, and reports "0 failed to publish" using the one signal that is
always absent when a message is silently lost.

## Where we're going

The bug stays open on purpose. One of about 178 runnable scripts has been migrated. The rest
are not being swept, because bulk-editing old one-off scripts risks firing live triggers, and
roughly two dozen of them are notes files that cannot run at all.

Next in the queue is the landmine checker, for the obvious reason. Beyond that there is a
better destination we have costed but not built: sending these messages from inside the cluster
in Go, where the message broker itself confirms receipt. That would make a silent loss
impossible rather than merely detectable, and it would leave a permanent record — which matters,
because we discovered that the table recording whether a message arrived keeps only about two
days, while the table recording rejections keeps a month. After forty-eight hours you can still
find out whether a message was refused, but never whether it arrived. That asymmetry is the
strongest argument for the sender proving itself at the time, and it is why the receipt exists.

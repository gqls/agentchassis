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

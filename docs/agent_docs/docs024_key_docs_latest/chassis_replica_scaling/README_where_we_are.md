# README — where we are (chassis replica scaling)

Append-only. Newest at the bottom. Plain prose.

---

**2026-07-20.** Started this off the back of the bug-003 work, at the owner's
request: state the problem plainly first, turn it into a plan afterwards.

The short version. There is one pod that handles everything the platform does
which isn't a spawned worker — the main chassis. It runs as a single copy, and
we can't just run two, because of a fault someone found back in May and never
fixed: each copy listens for replies under its own private identity, so every
reply gets delivered to *every* copy, and they both act on it. Two of them were
caught processing the same reply a fifth of a second apart. So a second copy
doesn't share the load, it duplicates the work and they trip over each other.

But sitting at one copy isn't a safe resting place either. Every time we deploy,
there is a gap where nothing at all is listening — and because of a separate
flaw (the system tells Kafka a message is handled *before* handling it),
anything arriving in that gap isn't delayed, it's destroyed. That's the real
reason behind the rule of thumb people have been following about not dispatching
work just after a restart. And if that single pod is killed for any reason — a
machine drain, running out of memory — the whole front door is shut until it
comes back, while scheduled jobs keep posting into a queue nobody is reading.

Two things I want to flag because they change what a fix should look like.

First, the fault is not as dormant as its write-up implies. That doc says it
only showed up because there were three copies running at the time. But the way
our deploys are configured, a new pod is started *before* the old one is
stopped — so every single deploy briefly runs two, and during that overlap the
duplicate-reply path is live. I've marked this as reasoning from the deployment
settings rather than something I've caught happening, because I haven't watched
a deploy closely enough to prove it. It's cheap to check.

Second, a trap. It would be natural to reach for "more copies" to fix the
separate complaint that dispatches queue up behind each other for half an hour.
It wouldn't work. The queue is a single lane, and Kafka only lets one consumer
read a lane at a time — so a second pod would sit idle. That's a different fix
(more lanes), and doing it by adding copies would import the duplication fault
for no gain at all.

There are three questions I'd like your steer on before this becomes a plan,
and they're really budget questions. Do we actually want multiple copies, or do
we just want deploys to stop losing work? — because the second is achievable on
its own, and is already half-built. Is the half-hour dispatch delay worth
re-plumbing the queue for, or is it just surprising rather than harmful? And
when we do fix the reply-listening fault, are we willing to accept a small
replay of old messages at the changeover, or should we skip them?

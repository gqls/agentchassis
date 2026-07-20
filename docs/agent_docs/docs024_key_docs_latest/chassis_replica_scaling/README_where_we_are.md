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

---

**2026-07-20, later.** You gave the steer: thousands of domains. That answers
the first two questions at once — we need real throughput, not just safe
deploys — so I've worked the problem statement up into a plan (the bottom half
of the PLAN file). The short version of what I found and what I propose:

First, a nasty surprise while checking the "cheap fix". The assumption was
that any copy can handle any reply, because everything lives in the database.
That's true at the database layer — but there's a line of code that actively
*throws away* a reply if it arrives at a copy other than the one that sent the
request. Worse, it throws it away *after* marking it taken, so nothing else
can ever pick it up. So the fix everyone had pencilled in — give the copies a
shared listening identity — would, on its own, quietly destroy most replies.
The guard has to be deleted in the same change; the database already prevents
double-handling properly on its own. I've sent my reading of this to the
diagnosis loop to check me, rather than building on my own confidence.

Second, the arithmetic. Today we get through roughly one dispatch every 25
seconds, everything in one lane, single file — about two to four thousand jobs
a day, and the queue routinely runs half an hour deep. Thousands of domains
means hundreds of thousands of jobs a day. No amount of copies fixes that
under the current design, because the queue only lets one copy read a lane,
and each lane handles one job at a time, start to finish. You'd need over a
hundred lanes, decided in advance, unchangeable downwards.

So the plan changes the shape rather than the numbers: keep Kafka as the
letterbox, and make the database the to-do list. The moment a message arrives
it becomes a visible row — so "did my dispatch vanish?" is answered with a
query, instantly, forever. A pool of workers picks jobs off the list; a slow
council run ties up one worker instead of blocking everyone behind it; and
copies of the chassis finally mean more throughput, not duplicated work.
Nothing upstream changes — every script and scheduler posts exactly as today.

Phased so each step pays for itself: first the to-do-list change (kills the
half-hour waits — no risky topology change at all), then the shared listening
identity plus deleting that guard plus running two or three copies (kills the
deploy losses), then tuning for scale. The fixes already in flight from the
bug-003 work are the foundation and stay exactly as planned.

Three things I'd like from you, none blocking the first phase: roughly what
one domain generates per day at target scale (sets how far we tune, nothing
else); a yes to doing the to-do-list phase first; and how long finished job
records need to stay queryable before we archive them.

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

---

2026-07-28. You asked this morning how we can parallelise work-item
processing — per domain, per area of code, or some other way — and the answer,
after reading everything this repo and the live system had to say, is: none of
those. The right unit is the individual job (the orchestration). Per-domain
ordering turned out to be a convenience someone coded into three database
queries, not something the system actually needs; per-area-of-code isn't a
thing the system even records. What the system genuinely requires is much
narrower — a job that declares "I depend on that other job" waits for it, two
copies of the same job can't both run, and two agents can't write the same
component at once. All of that is already enforced, and none of it needs the
one-at-a-time queue we have today.

So the build is the to-do-list design described above, and you approved the
whole programme today, stage by stage, with a written read-out to you at each
boundary. The stages: first a small safety fix (without it, two workers could
occasionally double-process one reply); then the main change — messages become
database rows the moment they arrive, and a pool of workers takes jobs off the
list, so a slow council ties up one worker instead of everyone behind it; then
replies go through the same pool; then two or three copies of the chassis;
then we hand the "release more than one site per tick" knob to the thread that
owns that query. Each stage ships dark behind a switch, gets proven live
(including deliberately breaking it to watch it recover), and can be turned
off again by unsetting one variable.

Two pieces of luck this morning: the council traffic got its own lane overnight
(another thread finished that), so the worst of the queue-jamming is already
relieved while we build; and the "spawn race" that blocked running councils in
their own pods turned out to be something else — a reply that never comes, not
one that comes too early — which is now squarely in the bug 029 owner's court,
not ours. Rolls of the chassis will wait for a quiet moment (no council
mid-run) as you ruled.

---

2026-07-28, late morning — the first stage boundary, and one discovery that
matters more than the day's building.

The building first. The small safety fix went to the review council and was
APPROVED (its first review run was literally killed by another thread's
deploy landing mid-review — the exact "a restart costs an in-flight council"
problem we have on record — so it went round twice). The main change — the
to-do-list and worker pool — is written, tested, and committed, switched off
by default and provably identical to today's behaviour until we flip it. Its
review is on round three: round one's objection was "prove your prerequisites
are done, don't assert them", which was fair, and round two's was "we can't
verify your citations" — also fair — so the audit it demanded now lives in
the database where reviewers can query it, the danger-radius is measured (the
change compiles into exactly one binary), and the safety cap I added has a
real measurement behind it instead of a guess.

The discovery. To measure "before", I fired five cheap page-refreshes at the
system at once — the same five that sail through one-at-a-time in seconds.
**All five failed. Twice. And not for the reason anyone would guess.** The
system did all the work: the page was rebuilt, the publish succeeded, the
"done" message was sent back. But the queue in front of the publisher was
running minutes deep with the fleet's own routine work, and the queue of
"done" messages coming back was minutes deep too — and the system only waits
three minutes before giving up and re-asking, and every re-ask joins the BACK
of the queue. So a success message arrived five seconds after the system had
stopped listening for it, four times in a row, and the job was declared a
failure with its own success sitting right there. It is a treadmill: once the
round trip is slower than the patience, no amount of retrying helps — the
retrying IS what keeps it failing.

Why this matters to you: it means the platform today cannot reliably run even
five of its cheapest jobs at the same time — not because any part is broken,
but because three queues and one timeout interact badly under load. It
almost certainly explains a chunk of the mystery timeout failures the bug-029
thread has been chasing for days (I've handed them the evidence and a cheap
test for it). And it confirms the build order you approved: the worker pool
fixes the first queue, sending replies through the pool (the next stage)
fixes the second, and the third — the publishers themselves being
one-at-a-time services — is now on the map as its own future stage, with a
decision for you when we get there about whether jobs should wait longer or
retry smarter.

Nothing needs a choice from you today. Next: the review's round three, then
building the new chassis and rolling it out in a quiet moment, then proving
the whole thing live with the same five-at-once test that fails today.

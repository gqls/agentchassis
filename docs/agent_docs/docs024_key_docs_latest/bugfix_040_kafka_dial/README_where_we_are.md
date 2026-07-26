# Where we are — Kafka dial timeouts (040-kafka-dial)

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-26

Picked up bug 040 — the one about Kafka connections intermittently timing out
all over the cluster. First thing: checked nobody else was on it. The number 040
is used by two different bugs, and the other one is owned by an active thread, so
that needed care. The one I want was untouched.

Then I went and measured the cluster instead of reading about it, and most of what
the bug file told me to look at turned out to be a dead end. The bug file was
written six days ago and suggested the likely culprits were connection-tracking
tables filling up, or the Kafka brokers being too busy to answer. Neither is true
now. The brokers are close to idle — about 6% of one CPU each. The connection
tracking table is at 1,021 entries out of 262,144, and its highest point in the
last four weeks was still less than half full. The kernel counter that would show
connections being dropped because a queue was full reads zero on every machine,
every day for the last eight days. So I've written all of that down as ruled out,
because otherwise the next person spends their whole time re-checking it.

I did catch the fault happening once, and it's worth describing because it's the
clearest look anyone's had at it. One service started up, tried to create its
Kafka topics sixty seconds later, and hung for **ten and a third seconds** before
giving up — and ten seconds is exactly the built-in giving-up time, so it waited
the maximum. Then it immediately tried again and succeeded in under a second. So
whatever this is, it's brief, it's self-healing, and it costs ten seconds each
time it happens.

Chasing that, I found something real about how the cluster looks up names. Every
Kafka connection has to translate a broker's name into an address, and the name
the brokers hand out is *slightly* too short. Because of a standard Kubernetes
setting, that means the lookup fails three times before it succeeds on the fourth
attempt — every single time, forever. The numbers back it up: **73% of all name
lookups in the entire cluster are failures**, 384,392 out of 525,152 in a day. The
cluster is doing about seven and a half lookups to get one useful answer.

**But I want to be straight about this, because I nearly overclaimed it.** I had
that looking like the cause, and then I tested it properly and it isn't. The name
server answers in under three milliseconds. I ran twelve hundred lookups from
three different machines and not one of them was slow. I timed the short name
against the correct long name and both finished in under a second. So the wasted
lookups are real waste — three times as much network traffic as we need, and
therefore three times as many chances for a packet to go missing — but they are
not what's causing the ten-second hangs. Fixing it makes the system tidier and
less exposed. It is not the cure, and I'm not going to write it up as one.

I also nearly made a mistake worth recording. The obvious fix for the name problem
is to change the setting that causes it. I worked through what would actually
happen and it makes things **worse** — four failed lookups instead of three. The
real fix is to have the brokers hand out their full name. I've written that into
the configuration files with the reasoning attached, but I have **not** applied it,
because doing so restarts all three Kafka brokers and that's your call.

**Then I found the thing that changes the whole picture.** I was about to add a
counter so we could finally measure how often this happens, and I thought I'd
better check the counter would actually be collected. It wouldn't have been.
**Nothing in this system has ever reported any metrics at all.** There's a whole
file of counters written and maintained, the monitoring system is configured to
collect them, every worker is labelled telling the monitor where to look — and
nobody ever opened the door. The monitor has been knocking on a closed port since
the day the fleet was built. I confirmed it directly: there are zero application
metrics in the monitoring database.

That's why this bug has been so hard to pin down. It isn't subtle. It's that we
have never been able to see it. The bug file even instructs the next person to
measure the rate by reading logs out of the worker pods — but those workers are
temporary, and their logs are deleted before anyone can look. So the one
instruction for measuring it describes something that can't be done.

So I've changed the shape of the job. Rather than guess at a cause, I've made it
measurable and cleared out the things making it worse:

- All Kafka connections now go through one piece of code that counts every attempt
  and how long it took, and separates "the name lookup stalled" from "the
  connection stalled" — which is precisely the question we can't currently answer.
- Metrics are now actually served, on the port the system has been claiming to
  serve them on all along. This is the biggest single win here and it was an
  accident.
- Four genuine defects fixed on the way past: connections were using four
  different and inconsistent timeouts (one path 3 seconds, another 10, nobody had
  noticed); one fallback address pointed at a service that doesn't exist and never
  did, so it could only ever waste ten seconds before failing; one configuration
  entry was missing its port number entirely; and one loop retried without pausing,
  so an unreachable broker would spin the processor flat out.
- Timeouts cut from 10 seconds to 5, and made adjustable without a rebuild.

**One thing I want to flag, because you asked me to close this bug and I haven't.**
Three reasons. The project's own rule is that a bug only moves to closed when it's
fixed *and* live — and everything I've written is Go code, which does nothing
until the next image is built. More importantly, I have not found the root cause;
I've found what it isn't, plus a set of things making it worse. And the bug file
itself says, in as many words, don't close this just because supporting fixes
shipped. So I've left it open and written a precise test for when it *can* close:
once the new counter is live, if it shows no timeouts across a week, it's done.

The other thing I should say plainly: **when that counter first goes live and reads
zero, that is not good news yet.** A counter nobody collects reads zero too. I've
put the check for that at the top of the runbook, and written a test that
deliberately fails a connection to prove the counter moves — because I came within
one step of shipping a metric into a void and then reading the silence as success.

Waiting on the review council now, and then on your next build. Nothing here goes
live until that build happens.

## 2026-07-26, later — the review council knocked it back, and it was right to

The review council rejected the first version. I want to write down why, because
the objection was a good one and I'd made the mistake in plain sight.

Along with the counting, I had also **changed three settings across the whole
fleet**: how long every service waits before giving up on a Kafka connection
(cut from ten seconds to five), and two settings controlling how long producers
hold connections open and how often they refresh their view of the cluster. The
council's guardian blocked it: those are behaviour changes to shared messaging
plumbing, affecting every pipeline we run, smuggled into something I'd described
as "just adding a counter".

The embarrassing part is that **I'd already written the argument against it
myself**, in the submission's own risks section — I noted the shorter timeout
would make things fail faster instead of waiting, which is only an improvement if
the waiting was pointless, and I didn't know that. And my justification for the
new numbers was a remark in the bug file and "the Java library uses a different
default". Neither is a measurement. I was picking numbers by guesswork *while
building the very instrument that would tell me the right ones*.

So all three are reverted. Every connection now keeps exactly the timeout it had;
the only thing that changed is that they're counted. The tidy-up isn't abandoned,
it's just in the right order now: ship the measurement, look at what the real
timings are, then choose. That needs a separate architecture review, which is
what the guardian asked for.

The council also caught that I'd built a new metrics server when the project
already had one that nothing was using — fair, and I hadn't explained why. When I
looked, that existing one turned out to be unusable as written: it would have
re-introduced a fake "everything's fine" health check that we deliberately removed
last week, and it registers itself in a way that crashes the process if anything
else does the same. So neither calling it nor ignoring it was right — I fixed it
and then called it, so there's one way to do this rather than two.

**And I found one more thing that would have sunk the whole fix.** Opening the
metrics port isn't enough on this cluster. The monitoring system here only looks
at things it's been explicitly told to look at, and nothing had ever told it to
look at our services — the labels on our workers use an older convention this
setup ignores entirely. So even with the port open, still nothing would have been
collected, and the resulting zero would have looked exactly like a fixed bug.
That's the same trap as before, one level up, twice in one day. The missing piece
is written and committed but not switched on; it goes live with the next build.

One honest limitation, now written into the case file: the counter covers the main
agent program and every temporary worker it spawns — which is the group this bug
actually affects — but **not** the thirteen other services. Their connections stay
invisible for now. Wiring them up is the obvious next job, and I've deliberately
not bundled it in, since bundling is exactly what got the first version rejected.

The revised version is back with the council now.

## 2026-07-26, evening — it's live, and it needed three fixes rather than one

Your build went out and the counting code is running. I checked it properly rather
than trusting the version number — and that mattered, because **the version number
didn't change**. Still v1.0.1167, same as before. The only way to know whether the
new code was actually in there was to look inside the running container, which I
did, with a control check either side to prove the method wasn't lying to me. It's
there.

Then the part worth telling you about. The metrics port was open — and the
monitoring system still collected **nothing**. There were two more layers
underneath, and each one on its own produces exactly the same symptom: a number
that reads zero.

**Second layer:** the monitoring system had never been told to look at our
services. It only watches things it's explicitly pointed at, and nothing pointed it
at us. I'd already written that piece; I switched it on. Six targets appeared
immediately — and every one of them failed.

**Third layer, and this is the one I'd frame:** the cluster's firewall rules were
blocking it. There is a rule, written some time ago, whose entire purpose is to let
the monitoring system read metrics — **it even names the correct port**. Somebody
sat down and wrote that rule intending exactly what I was trying to do. It has never
worked, for two independent reasons: it looks for the monitoring system in the wrong
place, and it identifies it by a name it doesn't go by. Either mistake alone means
it matches nothing at all.

I found it by testing rather than guessing: from inside our own services the metrics
port answers fine; from the monitoring system it times out; a different port answers
from both. That combination can only be a firewall rule. Fixed it narrowly — one
port, one source, nothing else opened — and kept a copy of the old rule.

**It works now.** Six of six targets healthy, and the monitoring database went from
**zero** of our metrics to sixteen. Not just the new one either: how many tasks each
agent processes, how many messages, how many workflows start, whether agents are
healthy. All of those were being counted and thrown away, for as long as the system
has existed. They're being recorded now.

**The first real measurement:** 240 Kafka connections, **every single one
successful**, and the slowest 1% completing in **28 milliseconds**.

Two things follow from that, and they point opposite ways, so I want to be careful.

**It does not mean the bug is fixed**, and I'd resist reading it that way. Twenty
minutes is not a week, and the original evidence had the always-on service showing
zero errors in exactly the same window where the temporary workers showed dozens. A
clean short sample from a quiet system is precisely what this fault looks like when
it isn't happening. Concluding "fixed" from it would be the same mistake this whole
investigation has been about.

**But it does settle the argument I got wrong earlier.** I'd wanted to cut the
give-up time from ten seconds to five, and the review council blocked me for
guessing. Now there's data: normal connections finish in 28 milliseconds against a
ten-second limit — about **360 times** the headroom. So the stall isn't ordinary
slowness drifting into the limit; it's a distinct, rare event. When we do change
that setting we'll pick the number from the measurements. The council's block cost
two rounds and bought a decision made on evidence instead of instinct.

What's left: let it run a week. If no failures appear, the case closes. If some do,
the counter will say whether the stall is in the name lookup or the connection
itself — and that single fact points at either DNS or the network fabric, which is
the fork nobody has been able to get past.

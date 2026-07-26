# SUMMARY — 2026-07-26b — 040-kafka-dial

Second summary of the day, written because the state genuinely inverted: the work
went from *committed but inert* to *live and collecting*, and the reason that took
three fixes rather than one is the finding. Current state only; chronology is in
`NOTES` and `README_where_we_are`.

---

## What we're trying to do

Kafka connections across the cluster time out intermittently — a service tries to
reach a broker, waits the full ten seconds, gives up, then usually succeeds
immediately on the retry. All three brokers, most machines. We want to know why and
stop it.

## Where we've come from

Filed six days ago as a network fault, with six things to investigate. Nobody
started it. Measurement ruled out five of the six — the brokers are idle, the
connection tables are nearly empty, no queues are overflowing, the name server is
fast. The real obstacle turned out not to be the network at all: **the fault had
never been measurable.** No counter existed for it, and the instruction in the case
file for measuring it — read the logs of the temporary workers — describes something
impossible, because those workers are deleted before anyone can look.

Building the counter turned up something larger: **nothing in this system had ever
reported a single metric.** Counters written and maintained for years, a monitoring
system configured to collect them, every worker labelled to say where to look — and
nobody had opened the port.

## What we've done

Made it measurable, and removed the things making it worse: connections all counted
now with the stall type recorded, four inconsistent timeouts identified, three
separate fallback addresses pointing at a service that has never existed, a config
entry missing its port, a retry loop with no pause.

The review council rejected the first attempt and was right — I had bundled
fleet-wide timeout changes into what I called instrumentation, having argued for
them from opinion rather than measurement. Reverted. Over three rounds it also
turned up two of those three phantom addresses and a fake health-check one call
away from being reintroduced.

**Then your build landed, and the fix still did nothing.** Serving the metrics port
was necessary and nowhere near sufficient. Three layers, each of which alone
produces an identical symptom — a number reading zero:

1. **Nothing served the port.** Fixed in this build.
2. **The monitoring system had never been told to watch us.** It only reads what
   it is explicitly pointed at; nothing pointed it at us. Switched on.
3. **The cluster firewall blocked it.** A rule exists whose sole purpose is to
   let monitoring read metrics, and it **names the correct port** — so the intent
   was deliberate. It had never matched a single packet, for two independent
   reasons: it looks for monitoring in the wrong place, and identifies it by a name
   it does not go by. Fixed narrowly; one port, one source.

All three are now live.

## Where we are now

**It works.** Six of six targets healthy. The monitoring database went from **zero**
of our metrics to **sixteen** — and not just this case's counter. Task counts,
message counts, workflow starts, agent health: all of it was being computed and
discarded for the life of the system, and all of it is being recorded now. That is
a larger win than the bug that uncovered it.

**First real measurement: 240 Kafka connections, every one successful, slowest 1%
completing in 28 milliseconds.**

**The case stays open, and I want to be firm about why.** Twenty minutes is not a
week. More pointedly, the original evidence showed the always-on service at zero
errors in exactly the window where the temporary workers showed dozens — so a clean
short sample from a quiet system is precisely what this fault looks like when it
isn't happening. Calling it fixed on this would be the same absence-as-evidence
mistake the whole investigation has been about.

Two things remain deliberately not done, both because they restart the Kafka cluster
and that is your call: a memory and CPU floor for the brokers (they run with none at
all today, and an unbounded memory ceiling), and the fix for the wasteful name
lookups that account for 73% of all cluster DNS traffic.

One honest limit: the counter covers the main agent program and every temporary
worker — the group this bug affects — but not the thirteen other services. Their
connections stay invisible until the same pattern is extended to them.

## Where we're going

1. **Let it run a week.** No failures → the case closes. Failures → the counter
   says whether the stall is in the name lookup or the connection itself, and that
   one fact points at either DNS or the network fabric. That is the fork nobody has
   been able to get past, and it is now answerable.
2. **The timeout decision is now evidence-backed rather than guesswork.** Normal
   connections finish in 28ms against a ten-second limit — around 360 times the
   headroom. So the stall is a distinct rare event, not ordinary slowness drifting
   into the limit. When we change that setting we will pick the number from the
   histogram. The council's block cost two rounds and bought exactly this.
3. A decision is owed on the two broker-side changes.
4. Extending metrics to the other thirteen services is the obvious follow-up,
   deliberately not bundled in — bundling is what got the first attempt rejected.

The wider point stands and is worth repeating now that it is fixed: **until this
afternoon, any claim anywhere in this project resting on one of our own metrics was
unverifiable**, because none had ever been collected. That was never specific to
this bug. This was just the bug that walked into it.

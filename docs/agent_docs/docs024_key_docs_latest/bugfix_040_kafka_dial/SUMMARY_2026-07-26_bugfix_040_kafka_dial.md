# SUMMARY — 2026-07-26 — 040-kafka-dial

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and `NOTES_040_kafka_dial.md`.

---

## What we're trying to do

Kafka connections across the cluster time out intermittently — a service tries to
reach a message broker, waits the full ten seconds, gives up, and then usually
succeeds immediately on the retry. It affects all three brokers and most machines.
It was reported six days ago. On its own each occurrence costs ten seconds, but
combined with known gaps elsewhere it can strand a piece of work for half an hour
or lose it entirely. We want to find out why it happens and stop it.

## Where we've come from

The problem was originally understood as a network routing fault: back on 15 July
it looked like one specific machine couldn't reach one specific broker. By the
20th that had changed shape entirely — it had spread to all three brokers and four
of five machines, at a low, flickering rate. That's what got it written up as its
own case, filed as a cluster infrastructure problem rather than a fault in our own
code, with a list of six things to investigate: connection-tracking tables filling
up, brokers too busy to answer, network plugin logs, and so on.

Nobody had started on it. It sat for six days.

## What we've done

We measured the cluster properly rather than reasoning from the write-up, and
**five of its six suggested causes turned out to be dead ends.** The brokers are
close to idle. The connection-tracking tables are at well under one percent, and
have never come close to full. The kernel counters that would show connections
being refused because a queue was full read zero, on every machine, every day for
the last eight days. The name server responds in under three milliseconds. Twelve
hundred deliberate lookups from three machines produced not one slow result. All of
that is now written into the case file as ruled out, with the exact commands, so
the next person doesn't spend their time there.

Along the way we found that the brokers hand out a name that's slightly too short,
which makes **73% of all name lookups in the cluster fail before the one that
works** — 384,000 wasted lookups a day. Real waste, worth fixing. But we tested
whether it was actually causing the timeouts and it isn't; with a healthy network
it costs no measurable time at all. We've been careful to write it up as waste
rather than as the answer, and we noted the trap that the obvious fix for it
actually makes it worse.

**Then we found the thing that matters more than the bug.** Preparing to add a
counter so the fault could finally be measured, we checked the counter would
actually be collected — and discovered that **nothing in this system has ever
reported any metrics at all.** There is a whole file of counters, actively
maintained; the monitoring system is configured to collect them; every worker
carries a label telling the monitor exactly where to look. Nobody ever opened the
port. We proved it from inside a live production container: the counter names
*are* compiled into the running program, and the port the system advertises
answers "connection refused". The monitoring database holds zero application
metrics.

So we changed the shape of the job: make it measurable first, then clear out
what's making it worse. All Kafka connections now run through a single piece of
code that counts every attempt, times it, and — importantly — records whether a
stall was in looking up the name or in making the connection, which is the exact
question we still can't answer. Metrics are now genuinely served. Four real
defects were fixed on the way: four different and inconsistent connection timeouts
nobody had noticed, a fallback address pointing at a service that has never
existed, a configuration entry missing its port number, and a retry loop with no
pause in it.

> **CORRECTED 2026-07-26, later the same day — two paragraphs above are now wrong,
> and the body is left standing deliberately** (a summary is what we believed at a
> milestone; overwriting it destroys the only record of how the understanding
> moved). What changed within hours of writing it:
>
> - **"four different and inconsistent connection timeouts nobody had noticed"** —
>   still true as a description of what was found, but the *fix* for it was
>   **vetoed by the review council and reverted**. Every connection keeps exactly
>   the timeout it had; only the counting is new. Unifying them, and choosing a
>   replacement value from the histogram rather than by guesswork, is deferred to a
>   separate architecture review. See NOTES and README for why the veto was right.
> - **"a fallback address pointing at a service that has never existed"** —
>   singular, and it was **three sites, not one**: `topic_manager.go`,
>   `spawn_actions.go`, and `agent_handlers.go`. The third is the worst (three
>   nonexistent brokers with no valid entry at all) and was missed twice, once by
>   me and once by an enumerating grep I had truncated with `| head`.
> - Also learned after writing: opening the metrics port is **not sufficient** —
>   this cluster's Prometheus cannot discover the pods at all without a PodMonitor,
>   which is now committed and not applied.

## Where we are now

The code is written, tested and committed, and it is **not live**. It takes effect
at the next image build. We confirmed it missed the build that went out this
afternoon.

**We have not found the root cause, and we're saying so plainly.** What we have is
a solid list of what it isn't, a set of amplifiers removed, and — for the first
time — the ability to see the fault when it happens.

The case therefore **stays open**, which is a deliberate departure from the
instruction to close it. Three reasons: the project's rule is that a bug closes
only when it's fixed *and* live; the root cause is unknown; and the case file
itself says not to close it on supporting fixes alone. Instead it now carries an
exact test for when it can close.

Two changes are written into the configuration files but deliberately **not
applied**, because both restart the Kafka cluster and that's the owner's call: a
proper memory and CPU floor for the brokers (they currently run with none at all,
and an unbounded memory ceiling), and the fix for the wasteful name lookups.

The change is with the review council now.

## Where we're going

1. The next image build makes all of this live. Nothing happens until then.
2. **First check afterwards is not the bug — it's whether the metric is actually
   being collected.** A counter nobody reads reports zero, and so does a fixed
   bug. That check is the first section of the runbook, and there's a test in the
   suite that deliberately fails a connection to prove the counter moves.
3. Then take a baseline, and let it run for a week.
4. If it shows no timeouts, the case closes. If it shows some, we'll finally know
   whether the stall is in the name lookup or the connection itself — and that
   single fact points at either the DNS layer or the network fabric, which is the
   fork the last six days couldn't get past.
5. Separately, a decision is owed on the two broker-side changes, which need a
   rolling restart.

The wider consequence is worth stating on its own: **until that build ships, any
claim anywhere in this project that rests on one of our own metrics is
unverifiable**, because none of them have ever been collected. That is not
specific to this bug; it just happened to be the bug that walked into it.

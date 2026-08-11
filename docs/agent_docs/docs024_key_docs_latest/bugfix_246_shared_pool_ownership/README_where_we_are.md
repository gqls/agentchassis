# Where we are — bug 246, the database connection setting that isn't

Append-only, newest at the bottom. Plain prose.

---

## 2026-08-11, early afternoon — what this is about

Picked up bug 246 from the open pile. Checked three different ways that nobody else
is on it: the live sessions of every other Claude working this tree (31 of them, none
mentions 246), the commit history, and — most directly — the note the lane that
*filed* it left behind, which says in writing that it is unowned and not started.

The bug is about a setting that looks live and isn't.

When one of our agent pods starts up, it opens a connection to the database and
decides how many simultaneous connections it is allowed to hold. That number is
deliberately configurable, and on the live chassis it is set to 12. Somebody raised
it to 12 on purpose, for a good reason: the chassis now runs a pool of workers
handling messages in parallel, and the code has a comment explaining that with four
workers plus a couple of background loops plus a reply path all sharing only four
connections, you get a queue — and, if anything is unlucky enough to be holding one
connection while it asks for a second, a freeze.

A few lines later, a second piece of code sets that same number back to 4.

Not a different pool — the *same* one. It's handed the pool object and quietly
resizes it. So the 12 survives for a few milliseconds and then goes away, every
time, on every pod. Anyone reading the configuration, or the comment, believes the
fleet runs 12. It runs 4.

## What I checked before believing any of that

I didn't want to take the bug file's word for it, and it's as well I didn't: the
bug file quotes the code as hard-coding 12, and it doesn't — the built-in default is
4, and the 12 only ever comes from configuration. That distinction turns out to
matter a lot, because it means the fix changes nothing at all for every agent that
hasn't set the value, and only affects the ones where somebody deliberately asked
for something different. That's a much safer change than the bug file implies.

I proved the resizing itself with a tiny throwaway program rather than by reading
documentation. It prints 12, then 4. Crucially I also made it print 9 when set to
9 — otherwise "it printed 4" would be indistinguishable from a test that always
prints 4 whatever you do.

Then I checked the live pods, because a defect that only bites when a setting is
present isn't a defect if nobody sets it. Both chassis pods do set it to 12, and
both are running the parallel-worker mode that the 12 was raised for. So we are
running the exact workload the raise was meant to support, on the connection limit
the raise was meant to replace.

## The honest part: I can't show you it's hurting us

This is the bit I want to be straight about, because it would be easy to oversell.

The lane that filed the bug also left an instrument — a log line that fires when a
database lookup fails transiently, which is the kind of thing a starved connection
pool would cause. I looked. It's zero.

Zero sounds like good news, and I don't think it means anything yet. The pods have
only been up under two hours, and over that window the chassis handled somewhere
between one and two messages a minute. At that rate a four-connection pool isn't
under any strain at all, so a zero is exactly what you'd see whether the limit is
fine or whether it would seize up the moment traffic climbs. The measurement can't
tell those two apart, so I'm not going to report it as if it can.

There's also no way to look directly. Go can tell a program how often it has had to
wait for a free connection, but nothing in our system reports that anywhere, so
today there is simply no instrument that can see this pool filling up.

So the case for fixing it isn't "this is causing an outage". It's narrower and, I
think, stronger: **a setting we deliberately configured does nothing, and nothing in
the system would ever tell us.** The filer said the same thing — "severity: unknown,
and that is the point".

## Two things I found that weren't in the bug

First, the same constructor would **crash** an agent that runs without a database at
all — it calls a method on a pool that doesn't exist. No agent is currently in that
position, so it's never fired, but removing the offending lines removes the crash
too, for free.

Second, and this one is a bit awkward: the configuration that sets the chassis to 12
**isn't in the repository at all**. It only exists on the live cluster. My first
reaction was alarm — that would mean the next deployment wipes it. I checked before
saying so, and that turns out to be wrong: the way Kubernetes merges changes, a
setting that's only on the live object and was never applied from a file gets left
alone. So it's safe. What's true, and still worth saying, is that you cannot learn
how the fleet is actually configured by reading the repository, which is how the
neighbouring pgbouncer config ended up with a comment reasoning from "3 chassis
replicas × 4 connections" when it's now 2 replicas that want 12.

I've written the alarming version down as a near-miss rather than a finding, because
it's the version I nearly wrote.

## Where this is going

I've asked for an implementation plan and I've put the mechanism through the
diagnosis loop for an independent read, rather than trusting my own confidence. Next
is the council review, then the code change itself, which I expect to be the removal
of three lines plus a test that stops anyone putting them back.

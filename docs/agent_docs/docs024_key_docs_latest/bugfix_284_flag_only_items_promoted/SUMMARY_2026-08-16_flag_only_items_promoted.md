# SUMMARY 2026-08-16 — the findings the machinery filed as its own failures

First summary for this lane. Written to be read aloud.

## What we're trying to do

Stop the platform recording its own correct observations as breakdowns.

The system inspects our sites and writes down what it finds. Most findings are
**jobs** — something is wrong and we have an agent that can fix it, so the note
names that agent and the machinery routes it. But some findings are **flags**:
nobody, human or agent, can automatically repaint a client's brand colours, restart
a customer's virtual machine, or decide which page a duplicated paragraph belongs
on. For those, the checker deliberately leaves the "who fixes this" field blank and
the note sits there for a person to read.

Those blank-handler notes were being swept into the work queue anyway, handed to a
dispatcher that found nobody to give them to, and stamped **"blocked — cannot be
routed to any agent."** The aim of this work was to make the machinery stop doing
that, in a way that holds for every checker that exists now and every one written
later.

## Where we've come from

The bug was filed yesterday by another session, which found eighteen damaged rows
of one kind and — being honest about the limits of what it had checked — said in
the file that the cause was **not** established and named the next step. It
suspected the step that hands work to agents.

That suspicion was wrong, and the file's own title says so: it accuses something of
grabbing rows that were parked. Nothing grabs a parked row; the dispatcher
physically cannot see one. The damaged rows had never been parked in the first
place. They were written in a state their authors believed was inert, and it isn't.

## What we've done

Traced the whole path first-hand and found the cause one step upstream of where the
bug was looking: the step that moves new findings into the queue never looked at
whether a finding had anyone to send it to. Six different checkers had each written
a comment at their own call site saying "no handler, on purpose" — and a rule kept
as a comment in six places is wrong in some of them.

Measured the real size: **sixty damaged rows across four kinds of finding on at
least fifteen sites**, not eighteen of one kind, with another thirty-seven queued
to go the same way. The original count was low because the search that found the
producers looked for a line of code setting the field to empty, and the two worst
offenders never mention the field at all — the language fills it in for them.

Fixed it where it closes the door: the promoting step now asks exactly the question
the dispatcher will ask a moment later, and both get that question from one shared
piece of code, so they cannot drift apart. It also now reports how many findings it
held back and of what kind, because a filter that quietly does less looks identical
to a quiet week. Three deliberate sabotage tests confirm each guard actually fails
when broken, rather than passing for the wrong reason.

Put it through the review council. It came back **REVISE** first time — and was
right twice. One seat proved that the marker I used to attribute the damage cannot
identify which checker produced a row; re-measuring on the right marker gave a
cleaner answer (nine and nine, exactly the two files edited). Another caught that a
loose phrase in my write-up was hiding a **sixth** producer I had not found. Round
two came back **APPROVED**.

## Where we are now

The fix is committed and approved. It is **not yet running** — this kind of change
only takes effect when a new chassis image is built and rolled — and the sixty
damaged rows are deliberately **not** repaired yet, because until the guard is live
they would simply be blocked again within the hour. The repair is written and
waiting, as a script that refuses to run until you tell it which verified build is
live, takes its own backup, and fails loudly if it does not do what it claims.

Two things are honestly unfinished and both are written down rather than glossed.
Two of the sixty rows were not created by the machinery at all — they were inserted
by hand, by other sessions, already in the queue with no agent named. This fix
cannot see those. And the two review seats that looked hardest at the change
disagree with each other about it: one says we unified too little, the other that
we touched a shared piece of the system more than the bug required. That is a
judgement for a person, so nothing was decided unilaterally.

## Where we're going

Three steps, strictly in this order, and the order is the whole point.

**First**, the next chassis release carries the guard, and we confirm it by asking
the running service what it was built from rather than trusting the release
happened. **Second**, and only then, the sixty rows are repaired to the states
their authors meant them to have. **Third**, a database-level rule makes the bad
combination impossible to write at all — which is the only thing that catches the
hand-inserted case, because around twenty places in the code write these rows
directly and bypass the shared front door.

That third step must come **after** the release, not before. Database changes take
effect instantly while code changes wait for a build, so adding the rule first
would make today's code fail on every site holding one of these findings, and stop
the improvement loop across the whole fleet. Getting that order wrong would turn a
tidy-up into an outage.

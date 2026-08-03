# Where we are — the work-item record that describes the wrong problem

Plain prose, newest at the bottom. Append; never rewrite.

---

## 2026-08-03, early hours

I picked up bug 091 off the open pile. It had been sitting since 26 July with
half of it done and the other half explicitly waiting for somebody.

**What the bug is, in ordinary terms.** When the platform notices something wrong
with a site, it writes a note into a shared to-do list. To stop the same note
being written a thousand times, it refuses to add a second note about the same
subject while the first one is still open. That is sensible and it works.

The problem is what happens when the *second* note is about something different.
The evidence checker files one note per **site**, but it finds problems per
**fact**. So: a number on your site drifts, a note gets filed saying so, a human
hasn't got to it yet — and then a *different* number drifts. The system tries to
file that, the duplicate-guard refuses it, and the finding is simply gone. The
note that remains says something that stopped being true days ago.

**I checked whether this is actually happening, and it is, right now.** The
evidence sweep ran at 18:36 last night. Comparing what each open note *says* with
what the sweep actually *found*: four of the five open notes are wrong. On one
site the note points at a completely different fact from the one that has moved —
so a person opening it would go and fix the wrong thing. On another the note
describes a problem that has since resolved itself. On a third, two drifting
facts are invisible because the note only mentions the three it saw first.

The bug file called this "a delay, not a loss" and rated it medium. I think that
undersold it. Nothing is lost for ever, true — the next sweep after somebody
closes the note will re-find everything. But in the meantime the only record a
human ever reads is actively misdirecting them, and that is worse than having no
record, because a wrong record looks exactly like a right one.

**What I changed.** The shared to-do writer now has a second mode: instead of
throwing the new finding away, it updates the open note so it describes what is
true *now*. It stays one note — this is not a licence to flood the review queue,
which is a live concern the owner has already ruled on. The new mode is off by
default and exactly one caller turns it on, so nothing else in the fleet behaves
differently.

**Two things I got wrong on the way, both worth knowing about.**

The first is that the obvious way to write this fix would have re-created the
very bug it fixes. The original complaint in 091 was that the system *reported*
having filed a note when it hadn't. The natural database idiom here ("insert, or
update if it's already there") makes the system unable to tell those two apart
again — it counts an update as if it were a new note. I only noticed because I
was writing the test. So the fix is deliberately a bit longer than the one-liner
the bug file suggested, and the code says why.

The second was messier. I added a small new column to the shared write, which
built fine, and then twenty tests failed in eight files I had never opened —
belonging to other people's work, some of it in progress right now. The
temptation was to go and update all twenty. What that failure was actually
telling me is that widening something shared charges everybody who uses it,
including people who get no benefit. So the new column is now only sent by the
callers that need it, and I touched three test files instead of eleven.

**One more thing, which I'd flag to anyone working in this area.** While testing,
I deliberately broke one of the guards to check the tests would notice — and they
didn't. It turned out that a particular database probe in this code ignores its
own errors on purpose, which means the test harness is simply deaf to a whole
class of change there. "No test failed" was not evidence of anything. I've
written that up as a landmine so the next person doesn't trust a green run for
something it cannot see.

**Where it stands.** The code is written, all the package tests pass, and I've
deliberately broken each guard in turn to prove the tests catch it. It has gone
to the review council (that takes about half an hour to come back). It is
committed, so the next chassis build will carry it, but it is **not live yet** —
Go changes need a new image. I'll keep the bug open until it is live and I have
watched it work on a real drift, because a fix that isn't shipped is still a bug
you can reproduce.

The one decision in here that is a judgement rather than a measurement, and which
I would welcome being overruled on: if a note is sitting in the human review
queue and somebody is halfway through reading it, this change can update it under
them. I decided that a note quietly changing is better than a note being quietly
wrong. A note that a machine has actively *claimed* is left alone.

## 2026-08-03, morning — it's live, it's proven, and the ticket is closed

The council came back **approved, 13 to 0**, with six advisory comments. Four of them
were things I could actually go and check, so I did rather than just noting them.

Three reviewers independently asked the same sharp question: one of my tests works by
asserting that a particular database call is *never made* — and they pointed out that
a nearby piece of code in the same function is known to swallow errors, which would
make a test like that pass no matter what. That is a fair challenge and it was one I
had flagged myself. I proved it by deliberately breaking the code and watching the
test fail with the right error. It isn't vacuous. But they were right to ask, and I've
written the proof into the test so the next person doesn't have to repeat it.

One reviewer made a point I liked a lot: my fix warns you when it fails to record
something — but the warning lived in the one place that calls it, not in the shared
code. So the next person to use this would have had to *remember* to add their own
warning, and if they forgot, the bug would come back silently. That is the same shape
as the bug I was fixing, one level down. The warning now lives in the shared code.

Then the good bit. **It shipped**, on chassis v1.0.1237 — another session's build
picked it up. I checked the running pods properly: not just that my new code is
there, but that the code it replaced is *gone*, which is the check that actually
tells you something.

**And the test wrote itself.** The bug file says to verify this by deliberately
corrupting a fact on a site to force a second problem. I didn't need to: four sites
were already dropping findings on every single sweep. So I just re-ran the daily task
that runs anyway, and watched.

It worked. All four reported "didn't create a note, refreshed the existing one" —
which is the honest answer, and the one thing this fix was most at risk of getting
wrong. And the notes were corrected in place, same rows, no duplicates. The one that
had been pointing at the wrong fact since 26 July now names the right one.

**Two things I want to be straight about rather than let them read as finished.**

The four improvements the council asked for are committed but haven't shipped yet.
They make the fix better; they aren't the fix, which is live and proven.

And one of the five notes is *still* wrong. It describes a problem that has since
sorted itself out, and my change was never going to catch that — it only does
something when there's a new finding to record. Retracting a note whose problem went
away is a different job, which someone else is already designing. So it's four out of
five corrected, and I'd rather say that than round it up.

**What's next.** I found the same defect in three more places while I was in there —
and one of them puts a *count* of the problems it's about to throw away in its own
summary line, which is a bleak little detail. I've filed those as bug 184 rather than
holding this one open, because a closed ticket ought to mean "this isn't biting us any
more", and it isn't.

I deliberately haven't fixed those three. The tool now exists and switching them over
is nearly free, but each one needs the same judgement made on its own evidence — is a
note that quietly *changes* better than one that's quietly *wrong*, for whoever reads
that particular queue? I answered yes here because I could measure it. Answering it
three more times at speed, on sites I haven't measured, is how a careful fix turns
into a fleet-wide surprise.

## 2026-08-03, late morning — the rest of the family, and a lesson about waiting

The remaining four improvements went live on chassis v1.0.1238, so the whole of the
091 work is now shipped and proven. Nothing is owed on it.

Then I went after the three other places with the same defect (bug 184). Before
changing anything I did what the bug file told the next person to do: measure.

**And the measurement said the opposite of what I'd assumed.** I'd filed those three
implying they were losing findings daily. They aren't. The directory freshness sweep
runs every day and examines *nothing*, because the claims it watches each have a
staleness window and none of them is due until **23 August**.

The interesting part is that this is an argument for fixing it *now*, not for leaving
it. In three weeks a batch of claims falls due all at once. The stale note from 25
July will still be sitting on the key when that happens — nothing works that queue —
so the first real finding in a month would be dropped and nobody would know. "Wait
until it actually bites" is usually good advice; it's wrong when the biting is
scheduled, singular, and unrepeatable.

One of the three I genuinely cannot measure and have said so rather than guessing.
Its note has listed the same 15 rejected items since 24 July, and whether the sweeps
since found different ones is simply not recoverable: the run history is kept for a
day, and a rejected candidate never gets written anywhere else. That's the case for
the fix rather than against it — it's what makes the next one visible.

**I also got a test badly wrong and want it on the record.** I wrote what looked
like a careful test for all three places — a table of the cases, a clear explanation
of why the bug is invisible — and it never called any of the three. It tested the
shared machinery instead, which was already proven. It passed happily while the code
was wrong.

What caught it was deliberately breaking one of the three and expecting the test to
notice. It didn't. And I nearly missed *that*, because the test suite was already
failing for an unrelated reason, so "broke it and got a failure" looked the same as
"broke it and nothing happened". A test that fails is only evidence if you know it
was passing first. Both tests are rewritten to call the real code, and I've broken
each one in turn to prove it complains.

It's committed and with the review council. It is **not live** — that needs another
build. And there's a date in the diary: 23 August is when the directory sweep will
actually have something to find, and that's the run that proves this properly.

**2026-08-03, evening.** The fix went live and the second bug is closed. The owner
deployed the new chassis the normal way — the whole fleet moves to one new image
number together, which I now know is the procedure; earlier today I had tried to
roll just the one service on its own number, which the permission system stopped,
and that turned out to be the right outcome for the wrong reason. Checked the
running pods directly: all three new log lines are in the binary, the two old ones
are still there, and the string that could never work reads zero, exactly as it
should. Before closing I re-checked my own morning's reasoning and found one place
I had been lucky rather than thorough: I had said "nothing can pick up a parked
human-review item" after reading two pieces of code someone else's comment pointed
at. This evening I enumerated every piece of code that can pick up any work item —
there are four, and none of them can touch a parked one. Same answer, but now it is
checked rather than believed. One thing remains, and only the calendar can do it:
on 23 August the first real run of the stale-claims check falls due, and someone
must look that day.

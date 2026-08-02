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

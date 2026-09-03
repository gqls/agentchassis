# README — where we are (bugfix 440)

**2026-09-02.** New lane. The bug in one sentence: when a page-rebuild instruction carries a
label the system doesn't recognise, it quietly does the cheapest thing (re-ship the old page)
and reports success — and today it is impossible to make it refuse instead, because the same
label field is also where humans legitimately write free-text notes to each other (eleven such
notes written just today). The plan: give the routing label its own field, keep the notes field
free forever, and then a label nobody understands becomes refusable — checked at the database
door itself, so even hand-written migration scripts can't slip past. Big design piece (RFC_062)
for the behaviour change; a small inert foundation ships now. The 404 team built the current
warning half and their review round just finished — we've left them a note so nothing is done
behind their back.

---

**2026-09-03.** You made all five calls and we've acted on every one. The design questions are
closed and written into the design document itself, so nobody has to re-litigate them: a
rejected instruction goes to the human review queue rather than being thrown away or blowing up
the job; the database itself will refuse a bad label, which is the layer hand-written scripts
can't sneak past; the free-text notes field stays free forever; and the 404 team co-signs the
one migration that touches their work.

With the wait lifted, the producer half is built and submitted for review. One thing in it was
worth the care: the obvious implementation would have quietly changed how one existing case
behaves — a particular kind of rebuild instruction that deliberately does nothing today would
have started doing something once we flip the switch, and it wouldn't have shown up until weeks
later, looking like the flip was broken. We found it by reading the vocabulary's own notes, and
the shipped version keeps the new label locked in step with the old one, so the flip provably
changes nothing for the cases that already work.

Remaining: the authoring safeguards for hand-written migration scripts, then the flip itself —
which needs one technical confirmation first (how the gate treats a missing label versus an
empty one), because getting that backwards would refuse everything already in flight.

---

**2026-09-03, evening.** The actual fix for this bug is now written and tested, and it is sitting
one signature away from going live.

To recap what the bug is: when something asks for a page to be re-rendered, it attaches a short
label saying why. The page-rerender machinery looks at that label, and if it recognises it, it
re-renders the page properly. If it does *not* recognise it, it quietly falls back to re-shipping
the page exactly as it already was — and reports success. So a typo, or a label someone invented
without telling anyone, produces a job that goes green having changed nothing at all.

The fix, now built, puts a door in front of that decision. A label the system recognises carries
on as before. No label at all carries on as before too — that is a legitimate case and always was.
But a label that is *present and not recognised* now stops, and the job is parked for a person to
look at, with a message that tells them exactly what the bad label was and what the legal ones are.
There is also a second, stronger guard at the other end: the database itself will now refuse to
accept a bad label in the first place, so most mistakes never get as far as the machinery.

**Two things I want to flag, because they are the interesting part of today.**

First, I nearly shipped something that would have caused a fleet-wide stall. Adding the database
guard needs a brief exclusive lock on one of the busiest tables we have. I tested it and it took
two milliseconds, which looks like proof that it is harmless. Then an earlier test of the *same*
statement turned out to have been stuck for over two minutes before I killed it. The wait is
luck-of-the-draw, and while it waits, everything else that wants that table queues up behind it —
so the bad case is not "my migration is slow", it is "the work-item pipeline stops". There is now
a five-second timeout in the file and an instruction to retry in a quiet period rather than force
it through. The lesson worth keeping: one measurement of something probabilistic certifies
whichever answer you happened to get.

Second, the fix quietly creates a smaller version of the very problem it fixes, and I only caught
it by running the check rather than reading it. We have a daily automated check that compares the
live system against what the code says it should be. That check keeps passing after this change —
which sounds good, and is actually a hole: the new half of the configuration is not covered by it
at all. So the exact kind of silent drift that caused the original bug could happen again to the
new code. The remedy is written down in the migration file itself, as work whoever applies it has
to do in the same breath, rather than left as a good intention.

**Why it is not live.** You ruled that the lane which owns the surrounding code co-signs this
change. That lane has had nobody on it since 26 August and my note asking for the co-sign is
unanswered. You chose "build it all, stop before applying", so that is where it is: written,
executed end-to-end against the real database inside a transaction that I then threw away, proven
to roll back cleanly, and waiting. Nothing is live and nothing has changed in production.

**What it is safe to know before deciding.** I measured the things that would make this risky, and
they all came back clean: no queued job would be wrongly refused on the day it goes live, and no
existing row anywhere in the table would fail the new database guard. So this is not a change that
needs a nervous window — it needs a signature.

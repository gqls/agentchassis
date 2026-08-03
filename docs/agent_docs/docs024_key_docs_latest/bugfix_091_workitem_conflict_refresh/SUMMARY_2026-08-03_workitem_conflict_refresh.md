# SUMMARY — 2026-08-03 — the work-item record that described the wrong problem

## What we're trying to do

Make the platform's shared to-do list tell the truth about what it found.

When the system spots something wrong with a site, it writes a note into a shared
queue. To stop the same note being written over and over, it refuses a second note
about the same subject while the first is still open. That is right, and it works.
The goal here was to fix the case it gets wrong: when the second note is about
something *different*.

## Where we've come from

Bug 091 was filed on 26 July by a session that found it deliberately, by inducing a
fault. It came with three suggested fixes. The cheap one — stop *reporting* a note
that was never written — shipped on 27 July. The real one, and the one this work
is about, sat for a week: nobody had picked it up, because it changes a piece of
machinery that around twenty different parts of the system call, and that is not
something to do casually at the end of a session.

The file rated it Medium and said "nothing is permanently lost". That turned out to
be the part worth checking.

## What we've done

**First, we measured it, and it was worse than filed.** The evidence sweep runs
daily. Comparing what each open note *said* against what the sweep actually *found*:
four of the five open notes were wrong. On one site the note named a completely
different fact from the one that had moved — a person opening it would have gone and
fixed the wrong thing. On another, two drifting facts were invisible. On a third, the
note described a problem that had since resolved itself.

**Then we built the fix, and deliberately did not build the one that was suggested.**
The obvious database idiom for "update it if it's already there" would have re-created
the very bug being fixed: it counts an update as if it were a new note, which is
exactly the false report 091 was originally filed about. The fix is therefore slightly
longer than the one-liner in the bug file, and reports three outcomes rather than two.
Two further deviations, both for the same reason — making the wrong thing impossible
rather than discouraged: the new behaviour is off by default and reachable only
through a new function, so a caller cannot half-use it; and the refresh refuses to
touch a note that a machine has actively claimed.

**We also closed a fork.** Three places in the code hand-wrote their own version of
this shared write, which two reviewers had objected to back in July. The reason they
forked turned out to be a missing field on the shared helper — so the shared door was
unusable to anyone who needed it. We added the field and routed all three through it.

**The council approved it 13–0**, with six advisory objections. Four were checkable,
so we checked them rather than banking the verdict: three reviewers independently
asked whether one of our tests was meaningless (it wasn't — we proved it by breaking
the code and watching the test fail); one asked us to move a warning into the shared
code rather than rely on each caller remembering it; one asked us to stop building a
query by string-pasting; and one caught a genuine contradiction in our own numbers.
All five are fixed.

**And it is live and proven.** It shipped on chassis v1.0.1237 and we verified it
against the running pods — checking not only that the new code is there, but that the
code it replaced is gone, which is the check that actually discriminates.

The acceptance test needed no artificial fault: four sites were *already* dropping
findings on every sweep, so simply re-running the standing daily task was the real
experiment. The run reported "created: false, refreshed: true" on all four, and the
notes were corrected in place — same row, no duplicates. The site whose note had been
pointing at the wrong fact since 26 July now names the right one.

## Where we are now

Bug 091 is **closed** — fixed, live, and verified against real dropped findings
rather than a laboratory one. The queue did not grow: five notes, five sites, before
and after.

Two things are honestly outstanding rather than quietly assumed:

- The four refinements the council asked for are committed but not yet in a shipped
  image. They improve the fix; they are not the fix, which is live and proven.
- One of the five notes is still wrong, and this change was never going to fix it. It
  describes a problem that has since gone away, and a refresh only fires when there
  *is* something to report. Retracting a finding that resolved is a different
  mechanism, already being designed elsewhere. So: four of five corrected, not five.

## Where we're going

The same defect exists in three more places, found by enumerating rather than
guessing. They are filed as **bug 184**. The fix they need already exists and is
live — what each one needs is a judgement, not a build: is a note that silently
*changes* better than one that is silently *wrong*, for this particular reader? We
answered yes for the evidence sweep on measured evidence. That answer does not
transfer for free, which is precisely why they were not switched over in the same
breath, and why 184 says what to measure first rather than what to type.

# SUMMARY — 2026-08-06b — the review-queue sweep can now reach the work it exists to do

*Second summary of the day. The morning one (`SUMMARY_2026-08-06_…`) closed the retraction work and
handed on one inherited limit. That limit is now fixed, live and proven, so the read-out has
genuinely moved — which is the only reason this file exists.*

---

## What we're trying to do

Stop the platform's human-review queue filling up with findings nobody will ever look at. The
queue collects "this page is missing something" items raised automatically. Many stop being true on
their own, because the page gets rebuilt. A daily sweep re-checks parked findings against what is
actually deployed now and closes the ones that no longer hold, so the queue holds real work rather
than ghosts.

## Where we've come from

The sweep was built, approved and scheduled earlier in this lane's life, and it worked. But the
session that scheduled it noticed something odd on the very first run and wrote it down as an open
item: of 500 items examined, 469 came back "I have no way to judge this". Ninety-four percent of
each day's effort was wasted. It recorded the waste, recommended a fix, and handed the lane on.

## What we've done

**Measured the thing the earlier note had not.** It had established the waste but never asked
whether any of the items the sweep never reached were items it could actually *have done something
about*. They were: **64 of them, 38% of all the work the sweep exists to do.** And not "reached
eventually" — never. The queue is read oldest-first and the front of it is packed with kinds of
item the sweep cannot judge, which therefore never leave; only about a hundred places in the queue
ever turn over. Everything behind them was permanently invisible.

The sharpest detail: the starved items were the **newest** ones. Oldest-first was chosen on the
reasoning that old findings are likeliest to be out of date — but a finding raised last week is the
one a recent rebuild has probably already fixed. The rule was starving exactly the items most
likely to be closable.

**Fixed it twice, deliberately.** A config change first — raise how many items one pass will look
at, from 500 to 1500. Live within the hour, no software release, reversible in one line, and it
un-starved all 64 immediately. Then the real fix: the sweep now only loads the kinds of item it can
actually judge, so its limit applies to useful work and an unrelated backlog can no longer crowd it
out. That list is taken directly from the list of kinds it knows how to judge, so the two cannot
drift apart as more are added.

**Found the recommended fix couldn't have worked.** The handoff proposed several scheduled jobs,
one per kind of item. The setting that would control that is read from somewhere the schedule
cannot reach — a trap the same page had documented two paragraphs earlier, after paying to learn
it. Logged fleet-wide, because the lesson generalises: a trap you have just written down is not
thereby disarmed for your next paragraph, and the recommendation half of a handoff gets far less
scrutiny than the evidence half despite being the part the next person actually executes.

**Proved it, twice over.** The review board approved it first time. Four of its advisory notes were
checkable so they were checked, not argued: two found real facts — the scheduled job genuinely is
switched on, and no existing caller can be affected by the one behaviour change. Then a build went
out carrying it, and rather than trust the version number the change was confirmed present in the
running processes against a baseline taken beforehand. Finally it was run for real.

## Where we are now

**One sweep closed 20 findings, and every one of them was an item the old code could never have
reached.** All were raised between 3rd and 5th August — the young end of the queue, the starved
part. The run before the fix closed nothing at all. The sweep now looks at all 168 judgeable items
rather than the first 500 rows of a mostly unjudgeable pile, and reports that it is finished rather
than truncated.

It also reported, for the first time, the size of the problem it *cannot* solve: **611 parked items
are of kinds nothing knows how to re-check.** That was always true; the old code was structurally
incapable of printing it, because it could only count the unjudgeable items that happened to fall
inside its batch — so the gap looked smallest exactly when it was worst.

One loose end turned out to be the same bug in disguise. An item that looked individually skipped —
no record of ever being checked, while its neighbours had one — had been suspected of a naming
inconsistency. It wasn't: it was simply too new to be reached, and the neighbours it was compared
against were two weeks older. It was checked and closed with the "inconsistency" untouched.

**And a correction, made an hour after the fact.** On finishing, this session wrote "the lane now
has nothing open" into the handoff. That was an overstatement: it closed the item the handoff was
reopened for, not the lane. Two things remain, both known and neither urgent — a de-duplication
change that is blocked behind a set of duplicate rows, and a dormant tripwire in a sibling check.
Re-measuring the first, this session then got the number wrong by inventing a filter from memory
instead of reading the live index, and nearly recorded a 56% growth that was an artefact of its own
query. Read properly, the blocker has grown from 48 collisions to 53. Both mistakes are written up
where they were made.

## Where we're going

Nothing needs a decision. Tomorrow morning's automatic run should repeat today's result unattended
— that is the only thing worth a glance.

Beyond that the lane's remaining items are the two named above, and the honest big one is the 611:
closing that gap means teaching the sweep more kinds of item, one at a time, each with its own
review. That is deliberately not this lane's next move, because the standing lesson from earlier in
this same lane is that adding a closer without first asking what already closes those items is how
you build duplicate machinery. The gap is now visible on every single run instead of invisible,
which is the precondition for doing it properly.

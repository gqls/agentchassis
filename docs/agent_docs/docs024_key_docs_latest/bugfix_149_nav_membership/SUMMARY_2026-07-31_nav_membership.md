# SUMMARY 2026-07-31 — nav membership: one declaration, one writer

Milestone read-out for `bugs_open/149` Group A. Written to be read aloud.

## What we're trying to do

Make the platform's own repair loop actually repair. `bugs_open/149` is a queue of
thirteen measured defects in the checker layer — the machinery that inspects live sites,
notices something wrong, and hands the problem to an agent that can fix it. This piece
is the **routing** half: what happens after a checker has found something real.

## Where we've come from

The queue was filed on 29 July by a session that pulled one thread — an unreachable
tool page — and kept pulling. Another session closed three of its items the next day
(a claims floor at the persistence seam, and making a silently-skipped discovery check
fail loudly). What was left in Group A was a set of five entries that all looked like
separate routing problems: a repair that cannot repair, two nav builders that
contradict each other, two tool creators that each do half a job, and a database
column whose default quietly makes the routing decision.

They were not five problems. They were one, and it had a shape.

## What we've done

**Found the loop, and proved it with a control.** A checker notices a page that has
been marked as belonging in the site's navigation and is missing from it, and hands it
to the agent whose entire job is rebuilding navigation. That agent rebuilds the nav
tables, re-renders the header and footer, and pushes them to every page — it does
everything required. Then one line inside it discards the page before reading whether
it was marked for navigation at all, because its address begins with `/tools/`. The job
completes, reports success, changes nothing, and the checker finds the same page again.

The evidence is on the record rather than argued: a job filed on 29 July for
gamesdesign.co.uk **named its own four targets** and completed that afternoon; two days
later all four were still missing. And the same checker, handler and action **did**
repair two pages on robot-hands.com — which differ from the four only in not living
under `/tools/`. That is as close to a controlled experiment as this platform offers.
The diagnosis loop confirmed the mechanism independently, first iteration, citing the
same line and the same row.

**Stated the rule the code was missing.** *A page's own flags declare whether it
belongs in navigation. Its web address may decide **where** it appears; it may never
decide **whether** it appears.* The code already half-knew this — it had a rule sending
exactly these pages to the footer, and a blunter rule shadowing it. Collapsing the two
was the fix, and it is smaller than the bug.

**Removed a second writer.** The navigation table is derived from the pages. Two things
wrote to it: the derivation, which wipes and rebuilds, and an older function that
scribbled one extra row per tool. The derivation could not reproduce the scribble, so
it deleted it — seven live links across three sites were in that state. The scribbler
is gone. Creation now records the intent on the page and **asks** for a rebuild.

**Declined the fix the bug asked for, and said why.** The queue said to write the
navigation row at creation time, because making a bad state impossible beats detecting
it. Here that is wrong: a navigation row is not a link (headers and footers are saved
files), while the checker that spots unreachable pages treats a row as proof of
reachability. Writing the row would have left the page just as unreachable and switched
off the only thing that would have noticed. The council's own seats called that "the
fail-loud-not-silent discipline this council exists to enforce".

**Shipped and proved it.** Council `4486f1a9`: **approved**, twelve reviewers, no
high-severity objection. Live on chassis `v1.0.1215`, verified inside both running
pods. On gamesdesign.co.uk the footer navigation now carries all six flagged tool
pages, correctly labelled, with the main menu untouched and the stray "Tools" group
self-healed.

## Where we are now

**Two of the three senses of "done" are true.** The fix is in the running binary, and
it demonstrably produces the right navigation and the right saved footer. The third is
not: pushing that footer out to each page needs a queue that **stopped platform-wide at
13:21 today**, before this work touched anything. Thirty-four page updates for this one
site are waiting in it. That queue belongs to another session's lane and has been handed
to them, measured, with the specific trap that hid it — the scheduler's "last fired"
timestamp keeps advancing while nothing runs.

**Two admissions on the record.** The change broke something it does not touch: routing
child pages into navigation for the first time fed full web addresses into a
label-shortening function that had only ever seen simple page names, and six footer
labels came out as "Tools/Damage Formula Designer/Index". Only looking at the finished
labels found it — the diff cannot show it. It is fixed, and the six labels are corrected
in the live database so the site is right now rather than after the next deployment.
And I answered a reviewer's objection with a week-long average that was true and
irrelevant: the queue had been dead for two hours when I quoted it. The reviewer was
more right than my rebuttal, and the correction is in all three places I had written it.

**Bug 149 stays open, and its own arithmetic needed correcting.** The file says
"of 12"; it has **thirteen** items. Counted properly: six are done — three closed
yesterday, three today (the broken repair, the half-done creators, and one answered as
"not a defect, and the reason was already written down") — one is half done, and **six
remain**. Closing the ticket would delete the record of those six. I inherited the wrong
denominator from the previous status block and repeated it before checking, which is a
small instance of the thing this whole file exists to guard against.

## Where we're going

Three of the six open items — scheduling, dispatch, and the dead checks — are the other
lane's live work, and today's dispatch stall is now the sharpest thing in front of them.
Of the rest: the parent-listing route is a genuine piece of build work, and the two
database-level questions (a column defaulting to "yes", and a footer column built by its
own private query) each want their own review round because both change behaviour across
every site. The next concrete step for this lane is narrower than any of that: chassis
`v1.0.1216` is built, pushed and waiting, and one of the twenty-six new footer links per
site will be wrong on any site whose pages have no authored navigation label until it
rolls.

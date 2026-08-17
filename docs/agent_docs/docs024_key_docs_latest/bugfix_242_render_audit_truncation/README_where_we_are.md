# README — where we are (bugfix 242)

2026-08-11 — Picked this bug up because nothing and nobody was working it. The problem in
plain words: once a week, every site gets its live pages photographed and measured by a
robot browser, and any site with more than 25 pages only gets its first 25 looked at — but
the saved report doesn't say so. It reads exactly like a complete, clean sweep. Both sites
big enough to hit the limit have been silently under-measured on every run so far, and the
same skipped pages are skipped every week.

Why the report loses the "I was cut short" note: the platform has a known design flaw
(already written up and ruled on by the owner as RFC_012) where any step that sends a
request and waits for the answer throws away its own notes-to-self when the answer
arrives — only the answer survives. The owner has already decided the remedy pattern:
important notes must either travel inside the answer itself, or be written to a permanent
log table before the request is sent.

So the fix follows that ruling exactly: the request now tells the robot browser "there
were 27 pages, I'm sending you 25", the robot repeats that back inside its answer (so the
saved report says 25-of-27, cut short), the step that files defect tickets from the report
now also says so, and a permanent log row records every time the cap bites. Plus the cap
itself is raised from 25 to 60 so no current site is actually cut short — the honesty
machinery is for the day one grows past it.

Next: implement, test, put it through the review council, commit for the next build.

2026-08-11 (later) — The fix is written, tested and committed, and the review council has
it. One nice piece of rigour: we didn't just test that the permanent log entry gets
written, we deliberately broke the code (moved the write to after the send) and watched
the test catch it, then put it back. The page limit is already raised from 25 to 60 on the
live system — that part took effect immediately, so next week's sweeps will cover every
page of every current site. The "say so when cut short" machinery rides the next software
release. Waiting on the council's verdict; if it asks for changes we'll make them.

2026-08-11 (night) — It's live and proven. The new software rolled out, and rather than
wait for a site to grow past the limit naturally, we ran the test deliberately: dropped
the limit to 5 for a few minutes, audited the 26-page loan calculator site, and checked
every place the truth should now appear. All three held — the saved report says "5 of 26,
cut short", the ticket-filing step says the same, and the permanent log has a row naming
the run. Limit put straight back to 60. The review council approved the change, and the
verification run even confirmed the one thing a reviewer had asked us to double-check
(that the log row names the right agent and step — it does). This bug is done in
substance; the file stays in bugs_open per the owner's filing rule.

2026-08-17 — The bug file has now actually moved to the closed pile. The owner changed the
filing rule back on the 12th ("if it is fixed and live it should be moved"), so today's
session re-checked everything and did the move. The re-check itself turned up two things
worth saying out loud. First: the database rows from last week's proof run have since been
cleaned away, so the lane's own notes are now the surviving record of that proof — the
grading was done against the live rows at the time, and that's written where anyone can
find it. Second, and more useful: the only render-audit run now in the system was this
morning's weekly sweep, and it tripped over something new. It picked a brand-new site
(created an hour earlier, no pages published yet), the audit step correctly said "nothing
to measure here, skipping" — and then the next step, the one that files defect tickets,
didn't recognise that answer and recorded the whole run as a FAILURE. So a polite "nothing
to do" gets written down as "something broke". That's been filed as its own bug (299),
fixed the same day (the ticket-filing step now understands the skip and says "skipped,
nothing filed" instead of erroring), tested including deliberately breaking the fix to
prove the test notices, and sent to the review council. The fix rides the next software
release. It had actually happened once before — during last week's 242 work, unnoticed,
in a run we were using to check something else. Both occurrences ever are now explained.

2026-08-17 (evening) — Two things to know, and the first one matters more than my bug.

**The fresh build that went out this afternoon contains none of today's fixes.** I checked
before trusting it, and the answer is not ambiguous: I asked the running program directly
whether it contains the new code — twice, on both copies of it, each time alongside two
control questions whose answers I already knew, so I could tell a real "no" from a broken
instrument. The controls came back exactly right and the answer was no. The reason is
simple and mechanical: the version label was not changed, so the machines served the copy
they already had rather than pulling the new one. New-looking pods, old code. Another
session working on a different bug (295) found precisely the same thing within the hour,
by the same method, independently. So at least two finished, reviewed fixes are sitting
committed and inert, waiting on one thing: bump the version label, rebuild, roll. That is
your call to make, not a session's — rolling the fleet is your action.

There is a trap in this worth naming: had I gone ahead and run my fix's live test against
that build, it would have failed, and the failure would have looked exactly like "the fix
does not work". I have written that warning into the bug file so nobody else walks into it.

**My bug (299) is otherwise finished and approved.** The review council passed it first
time round, with six advisory comments. I read all six and answered each with a
measurement rather than an argument — including going and checking whether the same defect
exists in the neighbouring piece of code (it does not: that one is built the opposite way,
permissive where mine was strict), and deliberately breaking my own test to prove it would
actually notice the problem it claims to guard against. One comment caught a genuine
procedural slip of mine, which I have logged in the wrong-calls file: I recorded the new
key in the register 15 minutes after the code commit rather than in it.

Two corrections went out beyond my own lane. Another bug file (294) told everyone the
review council was unavailable until September because of an account limit — that is
wrong, my round ran and was approved in six minutes today, so I corrected it in place with
the timings; if that note had stood, the next session would have skipped review for no
reason. And my bug number 299 collides with an unrelated 299 filed by another session this
evening; neither gets renumbered, so both files now say to identify them by name, not
number.

2026-08-17 (night) — The new build is real this time, and the bug is finished.

I ran the same check as before — asking the running program directly whether it contains
the new code, on both copies, with two control questions whose answers I already knew — and
this time the answer is yes on both. The version label was bumped properly, so the machines
actually pulled the new code.

Then I tested it for real. There was a wrinkle: the site that triggered this bug in the
morning had published its pages during the day, so the fault condition no longer existed
anywhere. Rather than invent a fake site or fiddle with a real one, I used one of the
internal "pool" placeholder rows, which genuinely has no pages — so it hits the same
branch honestly, and nothing was created or changed to make the test work. The run finished
the normal way, its record says plainly "skipped, no deployed pages, nothing filed", there
is no error attached to it, and it added nothing to the error log. Nine hours earlier the
identical input produced a recorded failure. Same input, different build, opposite result —
that is what makes it a proof rather than a demonstration.

One extra thing fell out of this that is worth knowing: the earlier bug (242, the audit that
could quietly measure only part of a site) had only ever been proven with a deliberately
rigged test. Tonight's ordinary weekly sweep of one of the live sites shows its honesty
fields working in normal traffic for the first time. So both fixes are now confirmed in the
real world rather than on a test bench.

The bug file has moved to the closed pile. Worth noting for anyone reading later: there are
two different bugs numbered 299 — mine, and an unrelated one another session filed this
evening about a call-to-action link on a web page. Neither gets renumbered; identify them by
name, not by number.

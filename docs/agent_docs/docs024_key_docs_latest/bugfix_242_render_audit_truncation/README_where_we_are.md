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

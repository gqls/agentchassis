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

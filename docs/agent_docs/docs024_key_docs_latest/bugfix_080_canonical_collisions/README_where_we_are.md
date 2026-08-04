# Where we are — bugfix 080 (duplicate pages from disagreeing names)

Append-only, newest at the bottom. Plain prose.

## 2026-08-03 — picked up, and the problem is bigger than the ticket says

The bug: two different parts of the system used to invent a page's name and web address in two
different ways, so the same logical page could exist twice — two live copies at two addresses.
The code that caused the known case was already fixed a week ago. What nobody did yet: check how
much damage exists, close the remaining ways it can happen, and build the alarm that notices it.

What I found today: robot-hands.com has TWO live duplicate pairs, not one — the news page (the
known case, both copies serving, with search engines being told the stray copy is the real one)
and the gripper catalogue page, which the ticket never mentions. Nothing in the platform can
retire the stray copies today — the "unpublish a page" mechanism was vetoed by review and is
being redesigned — so per your call we are NOT touching the live site; we make the damage
visible instead and leave the choice of which copy survives as a filed decision.

The work: close the last three code paths that still invent names by hand (the blog-post writer,
the tool deployer, a hidden fallback in the tool-page creator), pin the one that diverges on
purpose so it cannot drift, and build the alarm — a check that spots two pages claiming the same
canonical identity, filing a decision item when both are live. Measured first: none of the
changes moves any existing page.

## 2026-08-04 — written, reviewed, approved; waiting only on the next release

All of it is coded, tested and committed, and the review council approved it first time (six
minor advisories, none blocking — mostly asking me to double-check which database write each
changed path ends up using; I checked all three and wrote the answers down). One database-side
safety catch is already live: the timeout sweeper can no longer auto-complete these decision
items past their verifier — and fixing that up also closed a gap another thread had declared but
never switched on.

What's left needs the next fleet release, which is yours to run: once an image built from
today's code is live, I switch the new alarm on, watch it find exactly the two robot-hands
duplicates we know about, and then close the ticket. The "which copy of the page survives"
question lands in the review queue as two concrete decision items — the machinery to actually
retire the losing copy is still the vetoed/redesign question from the other workstream, so
nothing on the live sites changes until a human decides.

## 2026-08-04, later — done and closed

Your release landed and everything worked exactly as predicted: the new code is verified in
both running pods, the alarm is switched on, and its first live run found precisely the two
robot-hands duplicates we knew about — nothing more, nowhere else. A second run correctly
recognised it had already reported them and filed nothing new. The ticket is closed and moved
to the closed pile.

The one thing waiting on you, whenever it suits: two decision items in the review queue asking
which copy of each duplicated robot-hands page survives. Each one carries the naming convention
that was already decided, so it's a confirm-and-go, not research — but actually removing the
losing copy from the live site still waits on the "unpublish" mechanism being redesigned after
its veto.

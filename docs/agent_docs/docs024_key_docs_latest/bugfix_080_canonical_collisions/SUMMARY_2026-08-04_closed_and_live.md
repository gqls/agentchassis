# SUMMARY 2026-08-04 — bugfix 080: closed and live

**What we're trying to do.** Stop the platform from ever holding two live copies of the
same logical page. Different parts of the system used to invent a page's name and web
address independently, and because the database only rejects an exact name clash, a
*differently*-named copy slid straight in — two live pages at two addresses, search
engines being told the stray was the real one.

**Where we've come from.** The bug was filed in July when the gap-planner was caught
minting these; that one code path was fixed within a day. What sat unfinished for over a
week was everything around it: three more code paths could still do the same thing, the
known damage on robot-hands.com had never been dealt with, and nothing watched for the
problem happening again. The ticket also understated the damage — it knew about one live
duplicate pair; there were two.

**What we've done.** Three things, all reviewed and approved first time by the council.
First, closed the class: every code path that creates pages now derives the page's
identity from the one shared rule, with the two paths that *deliberately* differ (imported
sites keep their original addresses) exempted on the record and pinned by a test so they
cannot drift silently. Measured before shipping: not one existing page moves. Second,
built the alarm: a discovery check that spots two pages claiming the same canonical
identity and files a decision item for a human when both are live. Third, proved the lot
end to end on today's release: the new code is in both running pods, the check is
switched on, and a live run found exactly the two known robot-hands duplicates — and
nothing else — then a second run correctly filed nothing new.

**Where we are now.** The bug is closed and moved to the closed pile. The two duplicate
pairs on robot-hands.com are still being served — deliberately untouched, per the ruling
that this lane detects and files rather than mutating a live site. The "which copy
survives" question now sits in the review queue as two concrete items, each carrying the
already-decided naming convention so the decision is a confirmation, not research.

**Where we're going.** Two follow-ons, both owned elsewhere: acting on those two decision
items needs the "unpublish a page" mechanism, which the council vetoed in its first form
and is being redesigned under RFC 011; and the review queue those items sit in still has
no working surface (bug 033). When either lands, the robot-hands duplicates can actually
be retired. This lane is done.

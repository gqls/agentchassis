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

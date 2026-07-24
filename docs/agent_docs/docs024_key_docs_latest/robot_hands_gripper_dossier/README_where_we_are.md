# README — where we are: the gripper dossier pilot (robot-hands.com)

*Owner's running log. Plain prose, append-only, newest at the bottom.*

---

**2026-07-24 — designed, and the first piece built**

This is the first real build of the per-site AI idea (the per_site_ai
workstream): a paid-shaped, produced deliverable on robot-hands.com. A visitor
chats with a small assistant on a new page, describes their pick-and-place
application, leaves their email, and gets back a link to a proper engineering
report — every gripper in the site's index scored against their actual
application with the physics shown, every figure traced to a manufacturer
datasheet, and an honest "nothing in this index fits" when that's the truth.

The design is settled and written down (DESIGN doc in this directory). The
shape in one line: the public half lives on the little isolated island server
that the gauntlet work already built (the public never touches our cluster,
and the visitor's email never leaves the island); the report-building half
runs inside the platform and publishes the report as a page on the site; the
island notices the page appear and emails the link.

Decisions you made today: shared sender address for the emails; testing on
the live site approved (with cleanup); soft-launch without a nav link first;
a £/$50-a-month cap on the new AI key. The tunnel to the island is live —
your authorisation click happened. The one thing still only you can do:
issue that second spend-capped API key.

Built today: the scoring engine — the server-side port of MatchMatrix v2's
physics, reading the same verified figures the tool uses, with the whole
verdict logic (Match / Marginal / Insufficient data / No match), the
conflict-note maths, and the fact block that will keep the report's prose
honest (the writer may only use numbers from it). Tests all pass, including
the case where nothing fits and the case where a figure isn't published —
both must be said plainly, never papered over.

# Gripper dossier — summary, 2026-08-26

**What we're trying to do.** Prove that one of our sites can sell real engineering
help, not just pages: a visitor on robot-hands.com describes their pick-and-place
application in a short chat, and our platform sends them a proper gripper selection
dossier — deterministic scoring, honest prose around the computed numbers, a
published report page, a link by email. The visitor-facing half runs on an isolated
rented machine (the "island") so the public, and their email addresses, never touch
our main cluster.

**Where we've come from.** The cluster half — scoring, report building, publishing —
was proven end to end in late July on staged fixtures, including the honest-failure
branch. The public half was built in mid-August as a second tenant inside the
island's existing API: a chat intake, a bot gate, a pull feed the cluster polls, and
an email sender. It then sat built-but-not-shipped for ten days waiting on
deployment steps, credentials, and a database change on the island.

**What we've done.** Over the last two days we shipped it. The island's database
gained its three intake tables; the new service build went across and came up clean;
a configuration regression that would have silently undone July's rate-limiter
tuning was caught before it shipped and fixed at the source. The one genuine
surprise was that the intake's dedicated AI key turned out to have no credit — our
August "verified live" check had used a free endpoint that couldn't see an empty
balance. Credit restored, we drove two real requests through the whole system this
morning as its first genuine users.

**Where we are now.** The pilot is live, end to end, both branches proven in
production today. A precisely-specified request travelled chat → submission → pull →
build → validation → published page → email in about thirty minutes, and the page
carries the visitor's actual numbers. A vaguely-specified request was correctly
refused publication and the visitor got a polite apology by email instead — which is
the safety net working, but it also exposed the pilot's first real product defect:
the chat is satisfied by an answer the report-writer downstream isn't allowed to
hedge about, so a vague visitor is currently guaranteed a failure rather than a
follow-up question. That, plus a vocabulary trap around the word "travel", is filed
as bug 409 with fix directions. Nothing on the site links to any of this yet, so no
member of the public can reach it.

**Where we're going.** Next is the visitor-facing widget and report page on
robot-hands.com itself — and bug 409 should be fixed first, because the widget
invites exactly the vague first message that currently ends in an apology. After
that: the soft-launch decision (unlinked page versus a footer link) is the owner's,
and the longer-term plan remains generalising the scoring engine so the next site's
dossier is a configuration exercise, not a build.

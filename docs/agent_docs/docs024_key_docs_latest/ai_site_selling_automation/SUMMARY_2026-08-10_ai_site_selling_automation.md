# SUMMARY — 2026-08-10 — AI-site-selling automation

**What we're trying to do.** Sell AI-built websites end to end: a visitor
chats on webdesign.uk, a site gets built by the existing pipeline, it goes
live on a ugg2.com subdomain or a real customer domain, with an admin screen
to manage customers and their builds, and payment gating the work. The
machine does the intake and the build; the owner reviews the finished preview
and releases it.

**Where we've come from.** A research thread established (10 Aug) that most
of the pieces already exist and work — live chat intake, live 7-agent build
pipeline, live zero-config subdomain delivery, proven API domain onboarding —
and that the missing part is the wiring between them, plus one platform bug
(dispatch unreliability, bug 239) that blocks "completely automated". The
design corpus for exactly this product already existed in the concept
register, unbuilt.

**What we've done.** In one evening session: the owner ruled the three
gating questions (customer identity lives on the existing clients table;
the £1,200 done-for-you tier is the automation target, with a human
releasing; safe pieces may be built while the domain cutover awaits review).
On those rulings we shipped customer identity end to end — new database
columns, new admin API endpoints, and a Customers tab in the admin dashboard
— and all of it is now deployed and verified live on the running services.
Along the way we caught and recorded a real trap: the admin API's existing
"client" endpoints talk to an empty leftover tenant table, not to the table
that owns the websites; the research handoff was corrected in place. The
next build item — pulling chat conversations off the isolated chat machine
into the database without breaking the isolation rule — is fully designed.
The platform-code change went through the advisory review gate as soon as
the gate came back from a fleet-wide outage (the Anthropic account spending
cap, hit mid-afternoon, lifted by evening).

**Where we are now.** Customer records can be created, edited and browsed in
the admin dashboard against live data. Nothing yet creates those records
automatically, no payment gates anything, and the chat transcripts still go
nowhere. The dispatch bug is being worked in its own thread; the domain
cutover still awaits the owner's review.

**Where we're going.** Next: build the transcript ingestion (designed), then
the automated trigger — chat brief to seeded specs to a build firing with a
real customer id — once the cutover is reviewed and dispatch is trustworthy.
Payment gating follows the owner's pending choices on Stripe plumbing and
refund handling, listed in the current handoff.

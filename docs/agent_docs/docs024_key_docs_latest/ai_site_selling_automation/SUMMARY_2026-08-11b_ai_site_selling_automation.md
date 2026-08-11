# SUMMARY — 2026-08-11b — every decision made, every gate cleared: ready to build

**What we're trying to do.** Sell AI-built websites with almost no human
effort per sale. A visitor talks to the chat box on webdesign.uk, a site gets
built by the existing pipeline for £149 all-in, they approve the preview,
pay, and take the finished site away as a download to host wherever they
like. A small visible queue keeps the workload honest, voucher codes let the
owner hand out discounted or nearly-free builds, and an admin screen manages
the customers. Positioning is deliberately no-frills and honest about being
no-frills — and about the sites being AI-built.

**Where we've come from.** Two days ago this was a research handoff: the
pieces existed (chat intake, build pipeline, delivery layer) but nothing
connected them, and a dozen decisions were open. Yesterday the first working
session shipped the customer records end to end — database columns, admin
API, dashboard tab, all live and unanimously approved by the platform's
review council — and caught a real trap on the way (the admin API's old
"clients" endpoints talk to an empty leftover table, not the real one).
This morning's summary recorded the big product decisions, including
retiring the £1,200 offer entirely, with its complete copy archived in this
folder in case it ever comes back.

**What we've done since that summary.** The last open product question is
answered: the queue's wait note is an approximation with nothing binding,
and the owner can pause intake if software or scale misbehaves — so the
copy will promise nothing and "paused" is a designed state, not an error.
The Nominet side went from "owed" to "proven" in one afternoon: the owner
supplied the tag name (DESIGNCONSULT), applied for a second tag for this
venture using an application we drafted from Nominet's own registrar
guidance, and fixed the IP allowlist — and rather than trust the checklist,
we sent a real EPP login from the server cluster and got back "command
completed successfully", which is the only response that actually proves
the whole chain works. The domain-automation lane has been handed that
evidence and is no longer blocked on anything Nominet-side. Finally, two
build-ready designs were written into the plan: the queue and submission
gate (occupancy counted from live work items so there's no second source of
truth to drift, capacity and pause controlled from the admin dashboard),
and the subscription service with vouchers (single-use codes with a
recipient name and expiry, race-safe redemption, a switch for taking
payment after approval now and up front later, and the payment
notification pattern already proven with real money elsewhere on the
estate — with no refund code anywhere, since refunds stay manual and
unadvertised by design).

**Where we are now.** Every decision needed to build the product has been
made, and nothing external blocks the work except two things that were
always going to take their own time: Nominet's decision on the second tag,
and the three registrar API keys the owner will supply later. The live
website still describes the retired £1,200 offer everywhere — that
contradiction is known, archived against, and first in the work queue. The
customer admin surface is live; the dispatch bug that gates full automation
is being fixed in its own thread; the domain cutover still awaits the
owner's review.

**Where we're going.** The next session starts from the 11 August handoff
and builds, in order: the site copy and FAQ rewritten to the honest
no-frills model (including the hosting setup guide and affiliate links);
the queue; the subscription service and vouchers; the ZIP download step;
then the chat transcript ingestion and the rebuild of the chat inside the
framework. After that comes the automated trigger — chat brief to seeded
specs to a build firing with a real customer id — once the domain cutover
is reviewed and dispatch is trustworthy. The example sites built from the
owner's own domains become the sales proof that £149 is worth it.

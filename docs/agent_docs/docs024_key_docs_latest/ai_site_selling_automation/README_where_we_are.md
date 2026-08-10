# Where we are — AI-site-selling automation

Owner's plain-prose log. Append-only, newest at the bottom. Anyone may append;
nobody rewrites.

---

**10 Aug 2026, evening.** This folder is the working home for the next stage of
webdesign.uk: selling AI-built websites end to end — a visitor chats, a site
gets built, it goes live on a subdomain of ugg2.com or on a real domain, with
an admin screen to manage the customers and their builds. The big research
handoff in this folder says most of the pieces already exist and work; what's
missing is the wiring between them.

Three decisions were made today, put to you directly at the start of this
session: customer records will live in the existing `clients` table (we add
the missing columns rather than inventing a new ownership structure); the
automation we're building serves the £1,200 done-for-you tier — the machine
does the intake and the build, you review the preview and release it; and
while the webdesign.uk domain switch-over is still waiting on your review,
this thread builds only the safe pieces (the admin Customers screen, the
customer database columns, and the design for pulling chat transcripts into
the database) and keeps the auto-trigger on paper.

Two things you should know tonight, both found while checking the ground:

1. **The Anthropic account ran into its monthly spending limit at about ten to
   four this afternoon, and that has switched off every AI feature at once** —
   the chat box on webdesign.uk now politely hands out contact details instead
   of answering (that fallback working is by design), and the platform's
   internal review machinery is down too. Nothing here is broken; the fix is
   raising the limit in the Anthropic Console, which only you can do.
   Otherwise it resets on 1 September.
2. **The dispatch bug that blocks "completely automated" (bug 239) is being
   worked on right now in another chat**, so this thread is staying off it.

Next from here: build the Customers tab in the admin dashboard, and write up
how chat transcripts get from the isolated chat machine into the database
without breaking the isolation rule (the database reaches out and collects;
the chat machine never gets to push in).

**10 Aug 2026, later the same evening.** First real work landed, and one
assumption from the research didn't survive contact with the code. The
research said the admin screen for customers would be pure front-end work
because "the client endpoints already exist" — it turns out those endpoints
talk to a completely different, empty bookkeeping table left over from the
platform's multi-tenant era, not to the table that actually owns the
websites. So we built the missing piece properly: the customer database
columns you approved are now live in the database, there's a new set of
admin endpoints that read and write the real customer records, and the admin
dashboard has a working Customers tab — list, create, edit, and each
customer's sites. The two code halves sit ready and switch on with the next
routine redeploys; nothing needed doing tonight for that. We also wrote up,
in the plan, exactly how chat conversations will get pulled from the chat
machine into the database without breaking the isolation rule — that's the
next build item. The trap we found is written up so no future session wires
the customer screen to the wrong table.

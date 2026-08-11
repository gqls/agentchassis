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

**10 Aug 2026, night.** The routine redeploy went out and everything from this
evening is now switched on and checked: the Customers tab is live in the admin
dashboard, talking to the new endpoints, talking to the new database columns —
verified on the running services themselves, not just the version number. The
AI spending limit has been lifted (the chat box on webdesign.uk answers
properly again, quoting the right price), and with the review machinery back
up, the platform-code change went through for its advisory review — the
verdict will arrive in about half an hour and is recorded either way. What's
left on this stage is the next build item (pulling chat conversations into
the database) and then the trigger work, which stays on paper until the
domain switch-over gets your review and the dispatch bug is fixed. The
decisions still needing you are listed at the end of tonight's handoff:
which payment plumbing to grow, whether manual-only refunds stay acceptable,
where the chat service lives long-term, and who registers customers' domains.

**11 Aug 2026.** The advisory review of last night's platform change came back
approved unanimously on the first round, so that work is fully closed out. You
made the four outstanding decisions: we'll build out the subscription service
that's already half-present in the platform (even if it idles at first); the
price becomes £149 all-in with a small visible queue (3–4 slots, rough wait
time shown, closed when full) plus voucher codes you can hand out for £10 and
£55 sites; the chat gets rebuilt properly inside the framework so it knows
what we can actually build, and can later sell smaller things on its own; and
customers keep their own domain and DNS — they get a preview and a ZIP of
their finished site to host wherever they like, with hosting by us as a
clearly optional paid extra. No refunds, one round of changes. Worth saying
plainly: the £149 model contradicts most of what the live site currently
promises (£1,200, "you only pay if you like it", "we handle domain and
hosting"), so the site's copy, the FAQ answer about domains, and the chat
bot's price line all need changing together — that's the first job of the
next session, alongside deciding exactly when the £149 is taken. On Nominet:
we have the password you gave us, but still need the TAG name (the username
that goes with it) and the five server addresses added to the allowlist —
the registrar keys can wait, as you said.

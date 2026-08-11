# SUMMARY — 2026-08-11 — AI-site-selling automation: the product is now decided

**What we're trying to do.** Sell AI-built websites with almost no human
effort per sale: a visitor chats on webdesign.uk, a site gets built by the
existing pipeline, they approve it and pay, and they take the finished site
away to host themselves — with an admin screen for managing customers and a
small queue keeping the workload honest.

**Where we've come from.** Yesterday's summary recorded the first working
session: customer records built end to end (database columns, admin API,
dashboard tab — now live and approved unanimously by the platform's review
council), a designed-but-unbuilt path for pulling chat transcripts into the
database, and a list of open decisions. Since then the owner has answered
every major decision, and one of those answers changed the product itself.

**The decisions — what the options were, and what we chose.**

1. *Where customer records live.* Options: extend the existing clients
   table; put an owner column on each site; or a new ownership junction
   table. **Chosen: extend the clients table** — least new machinery, and
   the admin API already speaks that language. Built and live.
2. *Payment plumbing.* Options: port idea.uk's small, proven Stripe code
   into the platform, or build out the half-finished subscription service
   already in the platform. **Chosen: build out the subscription service**,
   even though it will idle at first. Vouchers will live there too.
3. *The product and its price.* Options: automate the intake of the £1,200
   done-for-you service (yesterday's answer); keep £1,200 as a premium tier
   beside a cheap tier; or drop it. **Chosen: £149 all-in, and the £1,200
   offer is off the table entirely** — the owner won't have time to make
   full, bespoke websites at the moment. The complete £1,200 site copy is
   archived in this folder so the offer can be revived if £149 causes
   trouble. Positioning is deliberately no-frills and honest about it —
   Ryanair-style: you get what you pay for, and we say so, including that
   the sites are AI-built.
4. *Demand control.* **A visible queue of three or four slots** with a rough
   wait note; when it's full, submissions close. (Exact queue mechanics are
   the one small question still open.)
5. *Vouchers.* **Single-use codes the owner hands out, nameable to the
   recipient, with an expiry** — one kind drops the price to £10, another
   to £55.
6. *When money is taken.* Options: pay up front, or pay after approving the
   preview. **Chosen: after approval while we test the system, moving to
   up-front later** — so the switch gets built, not the constant. Refunds
   stay possible behind the scenes (manual, in the Stripe dashboard) but
   are never offered publicly.
7. *The chat.* Options: keep the hand-written service on the VM, or move to
   an edge-worker design. **Chosen: neither as-is — the chat gets brought
   into the framework** (built and seeded like everything else, though it
   can still run on the VM), so it knows what we can actually build rather
   than reading a hand-edited list of facts. Later it may sell smaller
   things on its own: palette choices, layout choices, logo uploads.
8. *Domains and hosting.* Options: we manage their DNS (even our own
   nameserver), or customers keep their own. **Chosen: customers keep their
   own domain and DNS.** They get a private preview and a ZIP of the
   finished site to host wherever they like. We'll write a setup guide
   recommending a UK-based storage service plus Cloudflare or a UK
   equivalent, offer hosting/transfer as clearly optional paid extras, and
   send people who want a different kind of service to alternatives like
   Lovable or Durable via affiliate links. The own-nameserver idea is
   parked: it would make us responsible for customers' email, which the
   owner explicitly does not want.
9. *Differentiation.* **The example sites built from the owner's own
   domains are the sales proof** — the portfolio shows what £149 buys.

**Where we are now.** The customer admin surface is live. The product,
price, positioning, payment timing, voucher shape and delivery model are all
decided. The £1,200 copy is safely archived. The live website and chat bot
still describe the old offer everywhere, so the site currently contradicts
the decided product.

**Where we're going.** Next session, in order: rewrite the site copy and FAQ
to the no-frills model (through the framework, with a sweep so no stale
price claims survive), design the queue, build the subscription service with
vouchers and the payment-timing switch, add the ZIP download step, then the
transcript ingestion and the framework chat rebuild. The automated trigger
still waits on the domain cutover review and the dispatch bug. Outstanding
owner asks: the Nominet TAG name and allowlist IPs, affiliate programme
sign-ups, and in time the three registrar keys.

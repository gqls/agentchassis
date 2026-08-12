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

**11 Aug 2026, later.** You settled the rest of it, and one bigger call: the
£1,200 service is off the table completely — not hidden behind the new offer,
gone — because you won't have time to hand-finish full websites. Before
anything touches the site, we took a complete archive of the £1,200 site as
it stands today (every page, every block of copy, the rendered pages and the
fact register), stored in this folder, so the offer can be brought back if
the £149 experiment disappoints. There's no ready-made "pin a whole site"
button on the platform — the archive is the pin. The rest of your answers:
payment after approval while we test (switching to up-front later), refunds
possible behind the scenes but never advertised, vouchers single-use with a
name and an expiry, hosting handed to third parties with a setup guide we
write (UK storage plus Cloudflare or similar), affiliate links to the likes
of Lovable and Durable for people who want something else, and the honest
no-frills positioning throughout — including saying plainly the sites are
AI-built. The one small open question is exactly what the queue counts and
what the wait note promises. A summary of every option and choice to date
was written today for reading aloud: SUMMARY_2026-08-11.

---

2026-08-11, late evening. The payment machinery now exists. Tonight's session
built the £149 payment surface the rulings asked for: the voucher system
(single-use codes you can name to a person, expiring, dropping the price to
£10 or £55 — the codes look like WD-XXXXX-XXXXX and are generated for you),
the order ledger, the Stripe integration in the shape that already took real
money on idea.uk, and the switch for "payment after approval now, up-front
later". The database side is live; the code side waits for the next
auth-service deploy and for two Stripe keys only you can supply (a restricted
secret key and a webhook signing secret — same as idea.uk's setup). Until
those keys exist the endpoints politely refuse rather than pretend. There is
deliberately no refund button anywhere, as ruled.

The website copy rewrite (getting £1,200 off the live site) did NOT start
tonight, on purpose: the other Claude session that owns webdesign.uk was, at
that exact moment, live-testing the page-rebuild lock it built after the
chat-box-wipe incident. Rewriting pages through the same machinery it was
testing would have risked exactly that incident again. It's first in the
queue once their testing is quiet, and they've been left a note asking them
to signal.

One decision will be needed from you soon, but not tonight: when the Stripe
keys arrive, the webhook (the message Stripe sends when someone pays) needs a
public web address pointing into the cluster, and nothing exposes the cluster
publicly today. Options and trade-offs will be written up when it's due.

---

2026-08-11, after the evening deploy. Your fresh build carried the payment
machinery live: the service confirms it is running tonight's code, the billing
endpoints exist and answer correctly, and — as designed — everything payment-
shaped politely refuses until the Stripe keys exist. Nothing can charge anyone
by accident; nothing pretends. The four database tables (vouchers, orders,
payment events, the settings switch) are live and verified.

Three decisions now sit with you, none urgent tonight; they are written up
properly in the handoff. In one line each: (1) when you want to start selling,
create the two Stripe secrets (a restricted key and the webhook signing
secret) — same setup as idea.uk; (2) with those keys comes choosing how
Stripe's payment confirmations reach the cluster from the internet — three
options written up, my recommendation is routing through the webdesign.uk box
you already run; (3) the old half-built subscription code and the new payment
system both claim to know whether a customer has paid — I recommend retiring
the old one's write surface once a first real £149 sale has gone through.

The website copy rewrite (removing £1,200 from the live site) is still
queued behind the other session's lock-testing on webdesign.uk — they were
still at it this evening. It is the first item for the next working session.

---

2026-08-12, afternoon. I started the copy rewrite, and the first thing to
report is that the site was not saying what we thought it was saying.

My own note from last night said the live site still quotes £1,200. It does
not. The other session re-priced it on the 10th, so what is actually live is a
£75 deposit, a fourteen-day money-back window, and two rounds of revisions.
The price itself is not on the site at all any more: the home page says "one
fixed price" and invites people to ask what it costs. So the job was never
"delete £1,200" — it is "replace last week's terms, which are all wrong now",
and that turned out to be a page bigger than we thought, because a guide page
we had not counted repeats the deposit and the fourteen days in full.

Worth knowing, and it caught me out for a few minutes: **webdesign.uk itself
does not show the shop.** It has always redirected to webdesign.co.uk, which
is a different site of tools and guides. The actual shopfront lives at
preview.webdesign.uk. You confirmed the redirect was deliberate back on the
10th; I mention it only because anyone checking "is the site right yet?" on
the plain address will be looking at the wrong site entirely.

What I have done today is the half that has to come first. Every site we build
has a register of facts it is allowed to state, and a list of phrases it is
forbidden to use. The site's writer reads the first; a checker enforces the
second. Until that register says £149, asking the system to rewrite the pages
would just produce last week's offer again, very fluently. So the register now
describes the offer you ruled on the 11th: £149 all in, no VAT, paid after you
approve the site, no refunds, one set of changes, only a few sites at a time,
the site is AI-built and says so, and you get a preview link and then a ZIP
you host yourself, with hosting and your domain staying yours.

I also armed the other half, which is the part experience says gets forgotten:
the retired phrases are now banned mechanically. Before switching over, I ran
the platform's own checker against every page as it stands. Under the old
rules it found 3 problems; under the new ones it finds 36, spread across five
pages. That number is the point of the exercise — it is the proof the ban list
actually bites, rather than sitting there looking responsible. I then ran the
replacement wording through the same checker and it came back clean.

Two things I got wrong and caught before they shipped, both by running the
tool instead of trusting myself. One was a single wrong value type in the new
register, which would have quietly switched the whole checker off for this
site — no error, no warning, just a site with no guard. The other was a safety
assertion of my own that would have blocked the change for the wrong reason.
Both took seconds to find because the platform has a command-line version of
the same checker that runs on live data without deploying anything.

Five small things need a word from you, none urgent, all written up in full in
the working notes. The two worth saying out loud: the "three to four days"
promise was agreed for the £1,200 offer and nobody has re-confirmed it at
£149, so it is carried over and flagged rather than quietly re-stated. And the
payment-timing switch, which I described to you as yours to flip whenever you
like, turns out not to be free after all: the pages will say "you pay after
you approve it", so flipping it to up-front makes the site wrong until the
copy is rewritten. Worth knowing before you flip it, not after.

Next is the rewrite of the five pages themselves, through the framework, and
then checking the result against the same list. The other session is still
working on webdesign.uk today, but on the chat box's plumbing rather than the
page copy, and the only thing they have locked is the chat box, which carries
none of the old terms. So the collision risk that held this up last night has
gone.

---

2026-08-12, early evening. It is done. All five pages now sell the £149 offer,
and the old terms are gone from the live site.

I did it in two goes rather than all at once: two pages first, chosen so they
would disagree with each other. The FAQ was the hardest page on the site, with
eight separate things wrong on it; the home page had exactly one. If the
machinery could handle both ends of that range without damage, it could handle
the middle. It did, so I released the other three about twenty minutes later.

The thing I was most worried about was the writer quietly throwing work away.
When you ask this system to rewrite a page, it can hand back something shorter
and blander that technically satisfies the instruction. There is a setting that
makes it edit the page it already has rather than start again, and I used it.
The proof is in the sizes: not one of the five pages came back smaller in
substance. The FAQ actually grew by a fifth, because the new offer needs more
explaining than the old one did.

The guide page worried me for a different reason. It is a long article with
links out to four other pages, and this system has form for silently dropping
links during a rewrite: another project measured it losing five of thirteen. So
before touching it I wrote down the four links as data, and used a checker
another session had already built for exactly this. It passes now. More to the
point, I ran that checker against a deliberately impossible link first to watch
it fail, because a green light you have never seen go red tells you nothing.

Two things I want to flag rather than bury.

The first is a small mess with a clean explanation. One of the five pages is
recorded in the system as having **failed**, and the page is completely fine and
live. What actually happened is that the work finished, the page was written,
deployed and published, and then the message saying "I finished" got lost on its
way back. I have left the failure recorded rather than tidying it to "done",
because the record is true: something did fail. But I have written in three
places that the page itself is correct, so nobody re-runs it and rewrites a page
that is already right.

The second is the one that needs you. The site says, on three pages, "anything
that's our mistake, we fix at no cost". That was agreed back when the price was
£1,200. I deliberately left it out of the new list of things the site is allowed
to claim, because at £149 with no ongoing service it is an open-ended promise.
But the rewrite kept the sentence, because it was already on the page and I only
told the system to change the things that were wrong. Nothing objected, and
nothing would have: the system enforces the things you forbid, and merely hopes
for the things you leave out. So it is your call. Either you are happy to keep
fixing our own mistakes free at £149, and I add it back properly, or you would
rather it went, and I run one more short pass. It is not urgent and the site is
not lying either way, but it should not sit there with nothing behind it.

Same for "three to four days", which was also agreed at the old price and which
nobody has re-confirmed at £149. It is on the site and flagged in my notes as
carried over rather than re-agreed.

One last practical note. The site now promises two things the machinery cannot
yet do on its own: it says you get a ZIP of the finished site, and it says we
only take a few jobs at a time and close when full. Both are true and both are
currently manual. The ZIP would have to be assembled by hand today, and the
queue is a promise rather than a mechanism. Neither is a problem while nobody
can pay yet, but they move up the list the moment the Stripe keys go in.

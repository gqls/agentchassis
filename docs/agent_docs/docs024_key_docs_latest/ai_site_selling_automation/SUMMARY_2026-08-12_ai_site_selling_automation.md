# SUMMARY — 2026-08-12 — the site now sells the offer we actually have

**What we're trying to do.** Sell AI-built websites with almost no human effort
per sale. A visitor talks to the chat box on webdesign.uk, a site gets built by
the existing pipeline for £149 all-in, they approve the preview, pay, and take
the finished site away as a download to host wherever they like. A small visible
queue keeps the workload honest, voucher codes let the owner hand out discounted
or nearly-free builds, and an admin screen manages the customers. Positioning is
deliberately no-frills, and says plainly that the sites are AI-built.

**Where we've come from.** The product decisions were all made on the 11th, and
the payment machinery was built and deployed the same evening — deliberately
without keys, so nothing can charge anyone until the owner creates the Stripe
secrets. That left one embarrassing gap: **the live site was still selling the
old offer.** It had been repriced once already, on the 10th, to a £75 deposit
with a fourteen-day money-back window and two rounds of revisions — terms the
owner then retired outright. Fixing the site was the top item and had been
deferred twice, not because it was hard but because another session was
live-testing the rerender machinery on that exact site and we did not want to
collide with it.

**What we've done.** The site now sells the £149 offer, on all five pages that
mentioned commercial terms, and every retired term is gone from the live site.

It was two jobs, and the order mattered. Every site we build carries a register:
a list of facts it is allowed to state, and a list of phrases it is forbidden to
use. The writer reads the first; a checker enforces the second. Rewriting the
pages before fixing the register would have produced last week's offer again,
very fluently. So the register went first — £149 with no VAT, payment after the
customer approves, no refunds, one set of changes, a few sites at a time, built
by AI and saying so, delivery as a preview link then a ZIP the customer hosts
themselves. Then the pages were rewritten by the framework's own writer, as
ordinary queued work, with no hand-editing.

Two things about how it was checked are worth saying, because they are the
difference between believing it worked and knowing. The retired phrases were
banned mechanically, and we proved the ban list bites by running it against the
site as it stood: the old rules found three problems, the new ones found
thirty-six. And the riskiest page — a long guide with links out to four other
pages, where this system has form for silently dropping links — was protected by
a checker that was first shown failing on an impossible link, because a green
light you have never seen go red tells you nothing.

Two mistakes were caught before they shipped, both by running the tool rather
than reasoning. One would have switched the site's entire claims checker off
silently, through a single wrong value type that no error message would ever
have reported.

**Where we are now.** The site is accurate and the payment surface is still
keyless, which is the right way round: it advertises an offer we cannot yet
charge for, rather than charging for one we no longer sell. Two commercial terms
need a word from the owner — a "we fix our own mistakes free" promise that
survived the rewrite without anything behind it, and a "three or four days"
timescale carried over from the old price. Two things the site now promises are
real but manual: the ZIP is assembled by hand, and the queue is a promise rather
than a mechanism. Neither matters while nobody can pay.

One page's work item is recorded as failed while the page itself is correct and
live — the finishing message got lost after the work was done. That has been
left recorded rather than tidied away, and flagged in three places so nobody
re-runs it.

**Where we're going.** The Stripe keys and how Stripe's confirmations reach the
cluster are the owner's two decisions, and they are what stands between here and
a first sale. Behind them: the queue mechanism the copy now describes, automated
ZIP delivery, and the admin screen for issuing vouchers. The chat bot is the
other loose end — a sibling session is building the relay that will feed it the
same register the site now uses, which will end the standing problem of the bot
quoting prices from a hard-coded list that nobody remembers to update.

# BRIEF 2026-09-04 — the pricing page, rewritten around RENT versus OWN

**For the owner to read before it runs** (his standing request, and the precedent set with the
playground page). Nothing is dispatched from this file until he says so.

**Where it came from.** Owner, 2026-09-04: *"The pricing page can have a price list on it and it needs
to be rewritten e.g. 'What actually drives the cost is scope, not buzzwords.'"* Then, after looking at
the site: *"the 99 pounds gets you the model is fine, I missed it when I was looking at the hours page
where people were paying by the hour. Maybe the pricing page can make that clearer as to what you can
rent and what you can have."*

**The diagnosis in one line:** the offer is not unclear, the page is. `£99 buys a model you keep` and
`£n/hour rents a machine to try one on` are two different transactions, and the site currently presents
them in two places without ever putting them side by side. **A reader who lands on the hours page first
concludes the whole thing is rented.** The owner did, on his own site.

`[MEASURED 2026-09-04 13:35Z]` `/pricing.html` today is five sections — hero, one generic-text-block,
features, faq, call-to-action — and contains **no price list of any kind**.

## The proposed page, section by section

1. **hero** — one plain line: what things cost here, and the split named immediately. Something to the
   effect of "Two ways to pay us, and they buy different things: a model of your own, or an hour on a
   machine to try one." The writer's words, not these.
2. **generic-text-block — WHAT YOU CAN HAVE.** The £99 fine-tune. The trained model file is **yours to
   download and keep**, it runs on your own hardware, offline, for as long as you like, and nothing
   about it expires. What is included, in the plain terms the technical-details page already uses.
3. **generic-text-block — WHAT YOU CAN RENT.** The playground hour: a booked session on a machine with
   your model already loaded, so you can try it on real work before you rely on it. Priced per hour,
   by machine size, because a bigger machine answers faster and costs more to run. **The free demo on
   `/playground.html` costs nothing and needs no booking.**
4. **generic-text-block — WHAT ACTUALLY DRIVES THE COST.** The owner's own line is the subject:
   *"What actually drives the cost is scope, not buzzwords."* Then the honest list of what moves a
   price: how much of your writing there is, how many rounds of training it takes to sound right, how
   big a model the job needs, and how quickly you need it to answer. **No competitor comparisons, no
   "unlike other agencies", no jargon being knocked down.**
5. **the price list itself** — a plain table: the thing, what you get, the price. See the blocker below.
6. **faq** — the questions this page raises: do I own it, what happens when the hour ends, can I buy
   more hours, do I need a GPU (no, that is what the file is for), what if the model is not good enough.
7. **call-to-action** — book a call, or read the offer.

## ⚠ The blocker on section 5, which is the owner's to clear

This site is claims-gated: its `evidence_base` writer block forbids stating a number that is not a
registered fact, and the framework enforces it. **Today the only registered prices are `ft-price-99`
(£99, exact) and `ft-market-anchor` (~$5,000, approximate).** The hourly rates are not facts yet —
they are a proposal in a chat.

So the price list can be built **as soon as the owner confirms the hourly numbers**, at which point
they get registered exactly like the £99 did, with their basis recorded (the vendor's invoiced GPU
rates and the multiple applied). Until then, section 5 either does not exist or says "hours are priced
per machine size; ask us" — which is honest but weak, and not what he asked for.

**The numbers awaiting his word** (from the £1.50 + 6× cost decision): **£3.15/hour** on the small
machine, **£6.65/hour** on the big one. If he prefers the fixed £1.50 per booking rather than per hour,
the small two-hour session is £5.04 rather than £6.31.

## What this does NOT do

It does not touch `/your-own-model.html` or `/technical-details.html`, whose copy he has approved and
which already say the model is his to keep. It does not invent a new product: **"buy outright" turned
out to be what the £99 already is**, and the fix is that the pricing page says so.

# DRAFT — the "why not just ask an AI?" answer, and the £8 example place

**Status: DRAFT FOR THE OWNER TO EDIT. Nothing here is on the site.** The consent wording
is a commitment to a third party, not a code change. I am not a lawyer and this has not
been reviewed by one.

Owner decisions (2026-07-28): do both; price the trial at **£8**, not £5 — "5 sounds like
the service is a cheap one disguised as an expensive one".

> **REWRITTEN after owner feedback, same day.** V1 of both paragraphs was too long and
> "sounded like AI — negative framing and too many *not*s". It was: the copy kept defining
> things by what they were not ("we won't identify you", "this is not a promise", "nothing
> else changes"). V2 is roughly half the length and written in positive constructions. The
> content the owner approved is unchanged. **Keeping the diagnosis because it is the more
> useful half: stacked negatives are a reliable AI tell, and the fix is to say what the
> thing IS.**

---

## A. The objection paragraph — for `/report.html`

### Couldn't I just ask an AI myself?

Up to a point, yes — prompt one of the big tools for an evening and you will get something
worth reading.

The work is in the checking. We research every report with live web searches and put the
source under each finding, so you can go and read it yourself. A second AI, from a
different company, pulls the first one's work apart before you ever see it. Then a person
reads the lot.

We ask the same questions every time: what the problem is, who has it, who else is solving
it, where you are strong, where you are exposed, and one thing to do next.

And it will tell you when the answer is no. Watch that happen in the example.

**[See the example report in full →]** `/report/example/index.html`

---

## B. The £8 example place — consent wording

### The £8 example place

We are building a library of example reports so people can see the work before they buy.
Take one of those places and you get the full report — same research, same length — for
**£8**.

In return, you let us publish it as an example.

We publish the idea and the report. Your name, your email and your company stay out of it.

Which reports go in the library is our call — yours might appear next week, or never.
Taking it out again is yours: email us and it comes down within five working days.

Everything else matches the full report — same delivery, same refund if it turns up
nothing worth acting on.

Want your idea kept private? That is what the £29 report is for.

---

## The name-and-link offer: considered, then DROPPED (owner, 2026-07-28)

Proposed and then withdrawn the same day. Recording it because the reason generalises.

The idea was to publish the example with the submitter's name and a link to their site,
so the £8 read as exposure as well as a report. **The owner killed it on the sharpest
possible objection: we will not publish rude or poor submissions, so we cannot promise
publication — and a link is a promise whose delivery depends on what they send us.**

**A veto and a promise cannot both be honest.** We keep the right to refuse anything we
would not want in the library; the moment we advertise a link in exchange for £8 we owe it
to whoever pays, including the ones we would refuse. Better to promise only the report,
which we can always deliver, and keep publication entirely our call.

Two further reasons it was the right call, from the earlier analysis:

1. **We could not support the SEO claim anyway.** idea.uk had 26 non-bot page views in the
   ten days to 28 July, mostly cloud scanners and a Tor exit. A link from a site with no
   traffic is worth very little today, and this product's brand is not overstating things.
2. **Selling a link makes it a paid link.** Search engines expect `rel="sponsored"` or
   `nofollow` on links exchanged for money — which removes the value being sold. So the
   version that is safe to offer is the version with nothing much to offer.

**Transferable:** do not attach a benefit to a transaction when delivering that benefit
depends on the quality of what the customer supplies. Either you publish things you would
rather not, or you break a promise to someone who paid. The report itself has no such
problem — we can always deliver that, whatever arrives.

---

## The framing that makes two prices coherent

**£29 buys privacy. £8 buys the same report without it.**

A cheaper twin with no explanation tells the market the dearer one was a mark-up — which
was the owner's objection to £5. A cheaper twin that visibly costs the buyer something
does not, and it gives the £29 a stated benefit it did not visibly have.

---

## Conditions to hold it to

1. **Cap it.** "While we are building our examples library", or a fixed number. Five to
   ten specimens is plenty, and `MAX_ACTIVE_ORDERS=5` means a successful offer could fill
   the queue and make full-price buyers wait.
2. **£8 covers compute, not attention.** Model spend ~£1 per report [$1.23 measured
   2026-07-28; £ figure approximate], Stripe ~32p on £8, so ~£6.70 nets. That pays for the
   machine, not for the human confirm-and-approve step. Fine for an experiment.
3. **Same report, explicitly.** A lesser version would make the published specimen no
   evidence of what a £29 buyer receives, which defeats the point.
4. **Consent is an unticked box the buyer ticks**, stored on the order, with the wording
   visible at that moment. A pre-ticked box is not consent. One decision now, one record —
   the credit-me choice is gone with the link offer.
5. **Keep the right not to publish.** Stated in the copy, and it is the backstop for
   adverse selection.

---

## What building it takes (scoped 2026-07-28, not yet built)

`PriceGBP` is one value threaded `Config → StripeProvider.priceGBP → CreateCheckout`, so
price is fixed per *process*, not per *order*:

- `Order` gains a price and a consent flag (both persisted — a consent record must
  outlive the session that captured it).
- `Provider.CreateCheckout(orderID, email)` gains a price argument — an interface change,
  so `StripeProvider` and `FakeProvider` move together.
- `sendPayLink` passes the order's own price, or the email and the checkout disagree about
  what the buyer owes.
- Form: tier choice and consent box. Operator review email shows the tier, so approval is
  never accidental.
- Tests: no consent-flagged order without the box ticked; an £8 order never produces a £29
  checkout, or the reverse.

Roughly a morning with tests. It touches money, so it wants the same induce-the-fault
verification as `bugs_closed/089` rather than trust in a green suite.

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

**Your name is your choice.** Ask us to credit you and we will publish your name and a
link to your site alongside the report. Stay quiet and we publish the idea and the report
on their own — your name, your email and your company stay out of it.

We might publish next week, or never. The permission is yours to give and yours to
withdraw: email us and it comes down within five working days.

Everything else matches the full report — same delivery, same refund if it turns up
nothing worth acting on.

Want your idea kept private? That is what the £29 report is for.

---

## On crediting with a link — offer it, do not sell it as SEO

The owner's idea: publish the example with the submitter's name and a link to their site,
so they read the £8 as buying exposure as well as a report. **Worth doing. Worth being
careful how it is described.**

**Why to offer it.** It converts the cost of the barter into a benefit, and it improves
the library: a worked example with a real business attached is more persuasive than an
anonymous one. It also helps the adverse-selection problem — someone who wants their name
on it is likelier to be a real business with a real idea than someone offloading a
throwaway.

**Why not to market it as SEO.** Two reasons, and the second is the sharper one:

1. **We cannot support the claim.** idea.uk had 26 non-bot page views in the ten days to
   28 July, most of them cloud scanners and a Tor exit. A link from a site with no traffic
   and no history is worth very little today, and this product's entire brand is not
   overstating things. Selling "an SEO link" from here is the one claim that would read as
   dishonest to exactly the audience we want.
2. **Selling a link makes it a paid link.** If part of what £8 buys is a backlink, search
   engines treat that as a paid link and expect `rel="sponsored"` or `nofollow` — which
   removes the value being sold. Framed as *credit for the person whose idea it is*, it is
   editorial attribution and the problem does not arise. The framing is the difference,
   and the difference is real.

**So:** offer the credit and the link plainly — "we will publish your name and a link to
your site" — and let the reader decide what that is worth to them. The honest appeal is
**exposure and credibility**, not link juice: being the worked example on someone's site
is a marketing benefit that does not depend on domain authority.

Mark the link `rel="nofollow"` regardless. It costs nothing we are honestly offering, it
matches the `nofollow` discipline already agreed for commercial links elsewhere on the
estate, and it means the offer stays clean if the site's traffic ever does grow.

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
4. **Consent is an unticked box the buyer ticks**, stored on the order, wording visible at
   that moment. A pre-ticked box is not consent. Store the credit-me choice separately —
   two decisions, two records.
5. **Keep the right not to publish.** Stated in the copy, and it is the backstop for
   adverse selection.

---

## What building it takes (scoped 2026-07-28, not yet built)

`PriceGBP` is one value threaded `Config → StripeProvider.priceGBP → CreateCheckout`, so
price is fixed per *process*, not per *order*:

- `Order` gains price, a consent flag, and the credit-me choice (all persisted — a consent
  record must outlive the session that captured it).
- `Provider.CreateCheckout(orderID, email)` gains a price argument — an interface change,
  so `StripeProvider` and `FakeProvider` move together.
- `sendPayLink` passes the order's own price, or the email and the checkout disagree about
  what the buyer owes.
- Form: tier choice, consent box, optional name-and-URL fields. Operator review email
  shows the tier, so approval is never accidental.
- Tests: no consent-flagged order without the box ticked; an £8 order never produces a £29
  checkout, or the reverse.

Roughly a morning with tests. It touches money, so it wants the same induce-the-fault
verification as `bugs_closed/089` rather than trust in a green suite.

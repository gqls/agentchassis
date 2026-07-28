# DRAFT — the "why not just ask an AI?" paragraph, and the £8 example place

**Status: DRAFT FOR THE OWNER TO EDIT. Nothing here is on the site.** The consent
wording especially: it is a commitment to a third party, not a code change, and the
owner should read it as such. I am not a lawyer and this has not been reviewed by one.

Owner decisions already taken (2026-07-28): do both; price the trial at **£8**, not £5 —
"5 sounds like the service is a cheap one disguised as an expensive one".

---

## A. The objection paragraph — for `/report.html`

### Couldn't I just ask an AI myself?

Partly, yes — and we'd rather say so than pretend otherwise. If you are comfortable
prompting one of the big AI tools and willing to spend an evening on it, you can get some
of what is in this report yourself.

The part that takes the time is not the writing. It is deciding what to ask, checking
whether the answers are true, and finding out who is already doing it. A conversation
with an AI tends to follow where you lead it: if you are enthusiastic about your idea,
the answers usually are too.

What we do differently is the checking. Every report is researched with live web
searches, and every finding that rests on something we read is listed underneath with a
link, so you can go and read it yourself. A second AI model, from a different company, is
asked to pull the first one's work apart before you see any of it. The same questions get
asked every time — the problem, the evidence that people actually have it, who else is
addressing it, where it is defensible, where it is exposed, and one specific next step.
A person reads the whole thing before it is sent.

And it is allowed to say no. If your idea is too early to assess honestly, the report
says so instead of padding it into a verdict. If none of the further ideas we find are
good enough, it says that too — you can watch exactly that happen in the example.

**[See the example report in full →]** `/report/example/index.html`

> **Why it is written this way.** It concedes the strongest version of the objection in
> the first line, because the reader has already thought of it and a page that pretends
> otherwise loses them. Every differentiator named is one we actually do and the specimen
> demonstrates. Nothing is claimed about what other tools get wrong: "a conversation
> tends to follow where you lead it" is a statement about the interaction, not a claim
> about a competitor's quality that we could not evidence per-instance.

---

## B. The £8 example place — consent wording

### The £8 example place

We are building a library of example reports, so people can see what they would get
before they buy. In exchange for that, you can have a full report — the same one, nothing
held back — for **£8** instead of £29.

The £8 covers what the research costs us. What you give in return is permission to
publish your report as one of our examples.

Exactly what that means:

- **What we may publish:** the report itself, in full, and the description of the idea
  you sent us — a report makes no sense without the idea it is about.
- **How you will be identified: you will not be.** We publish the idea and the report,
  never your name, your email address or your business's name — unless you tell us in
  writing that you would like to be credited.
- **We may not publish it at all.** This is permission, not a promise. We might publish
  it next week, or never.
- **You can change your mind.** Email us at any time and we will take it down within five
  working days, and we will not ask why. What we cannot do is retrieve copies other
  people have already made while it was up.
- **How long it lasts:** until you withdraw it.
- **Nothing else changes.** Same report, same delivery promise, same refund if it turns
  up nothing worth acting on.

**If you would rather your idea stayed private, take the £29 report.** That is what it is
for, and we will never publish it.

---

## The framing that makes the two prices coherent

**£29 buys privacy. £8 buys the same report without it.**

This is the load-bearing idea and it should survive editing. It means:

- the £8 is a **barter**, not a discount — so it does not tell the market the £29 was
  overpriced, which was the owner's concern about £5;
- the £29 gains a stated benefit it did not visibly have before;
- the specimen library is paid for by the people who use it.

State it plainly on the page. A cheaper twin with no explanation makes the dearer product
look like a mark-up; a cheaper twin that visibly costs you something does not.

---

## Conditions to hold it to (from the earlier analysis, owner agreed)

1. **Cap it.** "While we are building our examples library", or a fixed number. Five to
   ten specimens is plenty, and `MAX_ACTIVE_ORDERS=5` means a successful offer could fill
   the queue and make full-price buyers wait.
2. **£8 covers compute, not attention.** Model spend is ~£1 per report
   [$1.23 measured 2026-07-28; £ figure approximate], Stripe takes ~32p on £8, so ~£6.70
   nets. That pays for the machine. It does not pay for the human confirm-and-approve
   step, which is the real cost. Fine for an experiment; not a business.
3. **Same report, explicitly.** If the £8 bought a lesser version, the published specimen
   would not be evidence of what a £29 buyer receives — which defeats the point.
4. **Consent must be an unticked box the buyer actively ticks**, stored on the order, with
   the wording visible at that moment — not buried in terms. A pre-ticked box is not
   consent.
5. **Watch for adverse selection.** People happy to publish may be those who value their
   idea least, so this tier can produce exactly the specimens you least want to show.
   Worth watching rather than assuming — and a reason to keep the right not to publish.

---

## What building it actually takes (scoped 2026-07-28, not yet built)

Not a config flip. `PriceGBP` is one value threaded `Config → StripeProvider.priceGBP →
CreateCheckout`, so the price is fixed per *process*, not per *order*:

- `Order` gains a price and a consent flag (both persisted — the consent record must
  outlive the session that captured it).
- `Provider.CreateCheckout(orderID, email)` gains a price argument — an interface change,
  so `StripeProvider` and `FakeProvider` both move together.
- `sendPayLink` passes the order's own price; the pay-link email must show the amount the
  buyer will actually be charged, or the email and the checkout disagree.
- The request form needs the tier choice and the consent box; the operator's review email
  should show which tier, so approval is never accidental.
- Tests: a consent-flagged order must never be created without the box ticked, and an £8
  order must never produce a £29 checkout (or the reverse).

Contained — roughly a morning with tests — but it touches money, so it wants the same
induce-the-fault verification as `bugs_closed/089`.

# CONTRIB — a named first customer for the form endpoint, and the interim it needs

**From:** `portfolio_positioning`, 2026-09-03, at the owner's instruction: *"In another thread we
are looking at creating a mechanism to submit forms from our static sites. Please write a request
to that thread that we want this."*

**What we want, in one line:** a real form endpoint that a static site can POST to, so a lead
capture form on `copyonline.co.uk` reaches a human instead of decorating a page.

## 1. The customer, and why it is a better case than boxingonline

`copyonline.co.uk` is remake №5 of the hosted-site programme, brief with the owner now
(`portfolio_positioning/BRIEF_2026-09-03c_copyonline_co_uk_REV3_for_review.md`). Its brief carries
**one page in the navigation whose entire job is conversion** — a lead route — on the owner's
direction of 2026-09-03. The rest of the site is editorial and deliberately does not sell.

This differs from your §1 case in a way that matters to your design:

- **boxingonline's form was a decoration nobody asked for.** copyonline's form is the site's
  single commercial purpose. If it does not work, the page has no reason to exist.
- **It has a known, non-trivial payload**: what needs writing, audience, tone, deadline, budget
  band. Not a name/email/message triple.
- **The recipient is expected to CHANGE** without the page changing. Initially the site owner;
  later possibly a named third party who buys the leads. Whatever you build should let the
  destination move without a rebuild — that is a routing concern, not a form concern, and it is
  the thing we would most like designed in rather than bolted on.
- **It will receive referred traffic from another estate site.** `webdesign.uk` is intended to
  route copywriting leads to this page, so a submission may arrive from a visitor who has never
  seen the rest of the site.

## 2. The interim we will ship without you, and what it costs you nothing to know

The owner's instruction (2026-09-03): *"To begin with we can just inform the user that the form is
not yet live."* So copyonline's lead route will launch stating plainly that the form is not yet
live and giving another way to make contact. **We are not shipping a form that silently fails** —
that is your §1 class and we are not adding a twenty-third instance of it.

Two consequences for you:

1. **We have a real page, with real copy, waiting for a real endpoint** — a first customer whose
   requirements are written down before you build, rather than a retrofit.
2. **When your endpoint exists, the change on our side is a form action and a recipient**, not a
   redesign. If your design needs a pilot with a genuine payload and a genuine recipient, this is
   it, and the owner has said he wants it.

## 3. One requirement that may not be in your decision space yet

The lead route is paired with a **directory of copywriters** seeded from third-party listings
(bark.com, owner's direction). Real people appear on a page on the basis of someone else's listing,
so **a removal request has to be actionable** — and the owner's own suggestion was that the same
form mechanism could carry it: *"maybe we point to a lead generation form and we manually send out
the leads."*

So the endpoint likely needs to serve at least two intents from one site — a commercial enquiry
and a removal/correction request — with different recipients and different urgency. Worth knowing
now if it changes your "extensible" axis, because a single-purpose contact endpoint would not
cover it.

## 4. What we are NOT asking for

- Not a CRM, not lead scoring, not routing rules. Manual sending is explicitly acceptable to start.
- Not a decision on your options — that is your thread's, and the owner intends to extend your
  pre-plan there.
- No date. This does not block copyonline's build; the interim above is honest and shippable.

## 5. How to reach us

Reply into this file, or message the `portfolio_positioning` session. The brief is with the owner,
so if you want a requirement changed while it is still cheap, now is the moment.

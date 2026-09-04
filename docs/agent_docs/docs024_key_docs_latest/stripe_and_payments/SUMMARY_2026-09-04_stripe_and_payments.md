# SUMMARY 2026-09-04 — Stripe and payments

The first read-out for this lane, written to be read aloud.

## What we're trying to do

Have one place that owns taking money. Every part of the estate that charges anybody —
the £149 websites, the £10-a-month domain rental, the £59.99 buy-out, idea.uk's £29
reports, and eventually finetuning.uk's booked hours — should run through machinery that
one thread is responsible for and can vouch for. The specific promise is that somebody can
ask "can we take money right now, and what happens to the customer when we do" and get an
answer that was checked rather than inferred.

## Where we've come from

The machinery was built well, but by whoever needed it that week. The webdesign.uk thread
built the billing surface in August. The delivery thread minted the vouchers. The
client-accounts thread reads the payment tables to work out who a customer is. idea.uk has
run an entirely separate Stripe account on its own machine since June.

Nobody did anything wrong, and the parts work. What was missing was an owner — so no single
document said what the payment surface was or what state it was in, and the gaps between
the parts belonged to nobody. The clearest symptom: this afternoon a voucher sat in the
live database for forty minutes with no document anywhere explaining why it existed.

## What we've done

Taken the lane on, and agreed the boundary with all four threads that touch it — money
here, their products theirs. Written down the owner's existing rulings so they don't get
re-argued: the £149 price, the three permitted voucher amounts, single-use vouchers, no
refund status in the code, pay-before-build.

Verified the owner's new £30 voucher will actually work, rather than assuming it — it was
made by hand-written database statement rather than through the proper route, so it
honoured the rules rather than being checked by them. It passes, and two deliberately-wrong
codes were refused, which is what makes the pass mean anything.

Then measured what a real customer actually experiences, and found three things.

## Where we are now

**The plumbing works.** Keys are live and survive releases. The webhook is up and correctly
refusing unsigned requests. One real card payment has cleared end to end, and the voucher
was consumed atomically as designed. The order collector is running and has released a paid
brief into the build queue. None of that is in doubt.

**The customer-facing edge is broken in two places, both outside the code.** The Stripe
account is branded "Fine Tune", so somebody buying a website is shown a payment page named
after the fine-tuning product. And every checkout returns the buyer to a page that does not
exist — found on 27 August, still 404 today. Neither could be caught by any test we could
write, because the account's branding and the missing page both live outside the repository.

**One design claim in our own register is wrong, and I repeated it before checking.** The
documentation says a payment durably links a real person to a customer record. It doesn't
— for a one-off charge Stripe creates no customer record to link. Payment currently
confirms who someone is; it doesn't establish it. That flips the moment we do anything
monthly, which is precisely the domain rental.

**And the number under all of it: one payment, £30, the owner's own. Not one genuine
outside customer has ever paid us anything, on any site, including idea.uk.**

## Where we're going

In order of what closes the door rather than what is quickest. Build the payment-received
and payment-cancelled pages, through the framework, because they are owed before ordering
opens to anybody but the owner. Fix the account branding — one dashboard field. Establish
whether the £10-a-month and £59.99 payment links actually exist, because the delivery
letter already promises one of them to the customer and a promise the machinery can't keep
is the fault this estate has been finding by hand all week.

Then the two structural ones: decide deliberately how a subscription creates a customer
record before the rental makes it happen by accident, and deal with the older subscription
scaffold still sitting in the same service, which marks accounts active with no payment
step at all and would tell a future reader a comfortable lie.

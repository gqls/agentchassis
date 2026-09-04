# Where we are — Stripe and payments

Plain prose, append-only, newest at the bottom. The owner's document.

---

## 2026-09-04 — this lane opens, and the first afternoon of looking

You asked this thread to take on everything to do with Stripe and taking money — vouchers
included — and to find out who else was doing bits of it.

**Four other threads were.** The delivery thread was minting your vouchers. The
webdesign.uk thread built the payment machinery in the first place. The client-accounts
thread reads the payment tables to work out who a customer is. And idea.uk has been
quietly running an entirely separate Stripe account on its own machine for months.

None of that was anyone doing anything wrong. But it does mean nobody owned the answer to
"what can take money right now, and does it work". That is what this thread is for now.
All four have agreed the boundary: money is mine, their products stay theirs.

**Your voucher is good.** The delivery thread minted `WD-KN3WU-9PZN4` for £30 this
afternoon for a second trial run. It made it by writing directly to the database rather
than through the proper route, and told me so, which was the right thing to do. I checked
it properly rather than assuming — I ran the exact statement the real payment path runs,
against a live copy, and then threw the change away. It came back correct, and two
deliberately-wrong codes came back refused, which is what makes the first answer mean
anything. **It will work when you use it.**

Then I went looking at what happens after someone pays, and found three things.

**The payment page says "Fine Tune".** Not webdesign.uk. When you paid your £30 in August,
the Stripe page you were looking at was headed with the name of the fine-tuning product.
You wouldn't have blinked, because both are yours. A customer would. It is one field in the
Stripe dashboard and costs nothing to change — but no test we will ever write could have
found it, because the account's branding isn't in our code at all.

**And after paying, you land on a page that doesn't exist.** This was found on the 27th of
August and is still true today — I checked the live site this afternoon. Every checkout
sends the buyer to `webdesign.uk/pay/success` when they're done, and there is no such page.
They pay real money and get a blank 404. Nothing is lost, because we take payment from
Stripe's own notification and not from the buyer's browser — but it is the worst possible
moment to show somebody nothing.

So the one payment page a customer will ever see is branded as the wrong company and
returns them to a dead end. Both are small fixes. Neither would ever have been caught by
anything automatic, for the same reason: they live outside the code.

**I also got something wrong, within an hour of starting.** I told the client-accounts
thread that a payment is what durably links a real person to a customer record in our
system, because that is what the design document says. They checked and it isn't true, and
when I checked their working myself, they were right.

The reason is quite specific and worth knowing: for a one-off payment, Stripe doesn't
create a customer record at all unless something forces it to. So there was never anything
to link. The payment worked perfectly and simply had nothing to write down. What actually
ties your Boxing Online order to that customer is the order reference, decided *before* the
payment — payment confirms who they are, it doesn't establish it.

That matters for one reason: it will start behaving the way the document describes the
moment we do a monthly subscription, which is exactly the £10-a-month domain rental.
So it wants deciding on purpose rather than discovering.

**The number I keep coming back to.** One payment has ever gone through the platform. It
was £30 and it was yours. Across every site we run, including idea.uk, **not one genuine
outside customer has ever paid us anything.** Everything above is a real defect and worth
fixing, but I'd rather state the thing plainly than let a list of working machinery imply
otherwise.

**One thing I could not check and need you for.** I tried to ask Stripe whether the
payment links exist — the £10-a-month domain rental and the £59.99 buy-out. My own
permission settings blocked the call, and I didn't try to go around it. So I don't know
whether those links have been made.

This matters more than it sounds: **the delivery email already tells the customer their
rental link "arrives in your delivery email"**. If the link doesn't exist, that letter is
promising something we can't do — which is the exact fault the delivery thread spent this
week finding by hand, three times over.

Two questions when you have a moment: have you made those payment links in Stripe? And
would you like me to fix the "Fine Tune" branding, or would you rather do it yourself
while you're in there?

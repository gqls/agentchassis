# SUMMARY 2026-08-27b — launch afternoon: the first order, taken and paid

*Second summary of the day, deliberately: the morning entry (SUMMARY_2026-08-27) ends at
"live, ordering closed, nothing sold", and by evening that account was out of date in the
way that matters most. Joint read-out with the delivery lane, as before.*

## What we're trying to do

Sell starter websites at webdesign.uk — £149 up front, built in three or four days,
delivered as a live link plus a ZIP the customer owns outright. The owner quality-checks
everything before it goes out, and the whole pipeline, from the sales chat to the
delivery email, is the same framework we sell.

## Where we've come from

This morning the site went live: the parking redirects came off, a never-exercised DNS
gap was found and fixed in minutes, and the full launch checklist passed from the
internet. At that point the machinery had never faced a real user. The plan was to leave
it resting until Monday.

## What we've done

The owner stayed a few hours and did the most valuable thing anyone can do to a
just-launched system: he used it like a customer. That surfaced three faults, all fixed
the same afternoon.

First, the chat appeared broken — actually the shared AI budget had hit its monthly
ceiling, and the chat's designed fallback (offering direct contact details) was doing
its job. The owner raised the limit; conversation resumed within a minute.

Second, a real bug: the chat's anti-abuse limiter was counting every message instead of
every new conversation, so a genuine five-message conversation hit a wall mid-flow. The
fix matches the code to its own stated design — new conversations stay limited, while a
conversation in progress is bounded by its twenty-turn cap and the daily spending
ceiling instead. Tested, proven by deliberately breaking it both ways and watching the
tests catch it, released to the box, and demonstrated live with a seven-message
conversation.

Third, the Website Brief Starter tool still told visitors to paste their summary "into
our contact form" — a form that no longer exists. The platform's own tool-improvement
machinery rewrote its three stale passages to hand the summary to the chat, which is the
real intake, and the served page was verified carrying the new wording. One requested
change was also cleanly reversed: the owner first asked to remove the tool's mentions of
third-party AI builders, then decided they should stay — both work orders were cancelled
before anything ran, and the ruling is recorded so nobody "tidies" those references away
later.

Then the milestone. The owner went through the whole buying journey as customer zero:
brief submitted through the chat (which held its own — including politely declining a
request for content-creation work that isn't in the product), an order reference issued,
a £30 trial voucher minted through the real admin machinery, a Stripe checkout created
against his reference, and a real card payment made. The payment landed end to end:
Stripe shows it succeeded, the webhook fired, the order is marked paid in our records,
and the voucher shows redeemed — every link in the money chain worked the first time it
was used in anger.

The purchase also exposed two things, one embarrassing and one reassuring. The
embarrassing one: after paying, the buyer is sent to a "payment received" page that was
never built — the owner landed on a bare 404. No money depends on that page, but it goes
on the must-fix list before ordering opens to the public. The reassuring one: his paid
brief did *not* start building automatically, and that is a deliberate hold — the
collector that releases paid briefs ships switched off until the wiring that carries the
customer's contact details and the honesty-guard evidence into the new site is built.
Releasing early would build a worse site; the brief waits safely.

## Where we are now

The site is live and selling-shaped: chat answering well and fairly rate-limited, the
brief tool pointing at the right door, Stripe proven with real money, the first order
paid and recorded, and the delivery machinery fully armed — the mail password reached
the pods with today's deployment, and every check passed. Nothing customer-facing can
happen unattended: the one paid brief is held by design, and every future delivery still
requires the owner's approval step.

## Where we're going

Monday, in order: build the release wiring (contact details and evidence into the new
site), switch the collector on, and let the owner's paid brief become the first
customer-shaped build; then his edit pass, the Approve press, and the delivery email to
his own inbox — which turns his £30 trial into the full rehearsal on a real paid order.
After that, opening ordering needs the payment-received pages, the Stripe payment links
for domains, and taking the "Not active yet" notice down. The voice editor remains the
next build after launch settles.

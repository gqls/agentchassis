# SUMMARY 2026-08-27 — webdesign.uk went live (joint read-out with the delivery lane)

*The series' previous entry is SUMMARY_2026-08-26 ("launch-ready"). This one is deliberately
fuller than usual, at the owner's request, and covers both halves of the product: the
shopfront that sells the sites, and the delivery machinery that hands them over.*

## What we're trying to do

Sell starter websites at webdesign.uk: a customer pays £149 up front, describes what they
want, and three or four days later gets a finished site — live at an address for a month,
and theirs to keep as a ZIP of files they can host anywhere. No revisions are included;
the pitch is honesty about exactly that. The site that sells this is itself built by the
same framework that builds the customers' sites, because a hand-made shopfront would
demonstrate nothing. Behind the scenes, the owner quality-checks and can quietly edit
every site before it goes out, customers get one email carrying everything, and the next
thing we build after launch is editing by voice.

## Where we've come from

Yesterday ended at "launch-ready, parked": the shopfront finished and verified behind a
parking redirect, the repositioned copy served on all seven pages, Stripe restored
durably after a deployment had silently wiped its keys, and both open product questions
ruled — the owner gives himself an invisible edit pass, and customers get no editing at
launch. The delivery machinery was complete, council-approved, and live as configuration
on a live deployment: the review gate, the once-only handover stamp, the customer links
door, the ZIP download route, the refresher that keeps download links alive, and an email
account proven end-to-end at Gmail itself. What remained was a short list: the mail
password onto the cluster, one configuration ride-along, one route applied on the box, a
rehearsal — and the launch switch itself.

## What we've done

The day began with a fault: the preview address answered "bad gateway". The cause was
nothing in the cluster — the small server that fronts all the webdesign addresses had
had its web server die at 6:22 that morning, when an automatic package upgrade restarted
it at the exact moment the private line to the cluster didn't answer. It had stayed dead
behind a healthy tunnel, which is why everything looked like a cluster outage and wasn't.
It was started again, a guard was installed so that failure now retries itself every
fifteen seconds instead of staying down, and the trap was written up so no future session
mistakes that error shape for a cluster problem.

That repair paid for itself immediately, because three launch items closed in passing.
The overnight deployment had already carried the mail settings onto the pods — one item
done with no work. The customer download-link route was applied on the box and verified
from the internet. And the "Not active yet" notice turned out to have been silently
stripped from the front page by an automatic rebuild the evening before — exactly the
hazard a note had predicted — so it was put back and verified showing in both places.

The owner created the mail password secret. The question of whether it should live in
terraform was answered no, with the mechanism checked rather than assumed: the Stripe
keys were wiped because terraform owns that particular secret wholesale, but it doesn't
know this new one exists, so a release cannot touch it — and moving it in would mean the
password living in a file, which we deliberately avoided.

One product change was requested and then withdrawn, and both halves went properly. The
Blueprint Compiler tool recommends pasting its output into two named third-party AI
builders; the owner asked for that to go, and the change was filed through the
framework's own tool-improvement mechanism — no hand-edited HTML, two work items, the
second queued so it couldn't collide with a repair already waiting. The owner then
changed his mind: we're a different kind of service, and may in future positively
recommend such tools where we don't suit. Both items were cancelled before the platform
had picked either up, the page verified untouched, and the ruling recorded so no future
session "tidies" those references away.

Then the launch. The owner switched off the two parking redirects in Cloudflare — and
the first visitors got "connection timed out". This was the day's genuine discovery: the
parking redirect had answered at Cloudflare's edge since early August, so the domain's
address records had never once been exercised, and they didn't point at the tunnel. The
server side had been correctly configured for weeks; nobody could have known the
addresses weren't. The records were rewritten from the box itself using the tunnel's own
credential, the main address answered within seconds, and www settled minutes later.

The full launch checklist then passed from the internet: both addresses serve the real
site with the right headline and the "Not active yet" notice twice; the one path that
must never leak a cluster route returns the correct kind of 404; the payment webhook
gives the keyed refusal that proves Stripe is armed; the chat answered a real question
sensibly; and none of the neighbouring services changed. Everything a future session
needs was written down: the trap entries updated, the verifier dispatched, both lanes'
logs and the memory index brought current.

## Where we are now

webdesign.uk is live on the internet at both addresses, selling with ordering still
closed — the notice stays up until payments open. Chat works. Stripe's webhook is keyed.
The customer links door (confirmation and download) is live end-to-end and verified from
outside. Every piece of the delivery chain is running configuration on a running
deployment, every review trail is approved, and the email needs exactly one thing: the
next routine fleet deployment, which carries the password onto the pods. Nothing
customer-facing has happened yet — zero customer links exist, zero sites handed over —
so nothing is at risk while we finish. And the estate is a little safer than yesterday:
the box heals its own web server now, and two traps that cost real time this morning are
recorded where the next person will trip them.

## Where we're going

Immediate: the owner's quiet-moment fleet deployment, then the full rehearsal on a site
of our own — file the review, owner edits and presses Approve, cut the ZIP, send the
email to ourselves, click every link from the internet, and check the message
authenticates at the receiving end. That rehearsal is the last thing between us and a
real customer.

To open ordering: the Stripe Payment Links (the £10-a-month domain rental and the
£59.99 buy-out), taking the "Not active yet" notice down, and the analytics items on the
owner's checklist. In parallel, the domain programme's first real registration stays a
deliberate, owner-gated moment. After launch settles, the next build on the delivery
lane is the customer voice editor, per the owner's ruling — and the delivery lane still
owes the weekly chase and the end-of-window retraction job before the first month of
hosting actually expires on anyone.

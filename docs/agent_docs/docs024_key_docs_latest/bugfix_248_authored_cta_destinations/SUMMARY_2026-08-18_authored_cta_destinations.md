# SUMMARY 2026-08-18 — authored CTA destinations

## What we're trying to do

Stop our own repair machinery from destroying the "Get in touch" buttons on our sites. On a
site that sells a service, the contact button is the point of the page — and something in the
platform had been quietly repointing those buttons at whatever calculator ranked first.

## Where we've come from

The fault was found on 10 August by the lane that was about to bulk-drain a repair queue, and
it stopped them one command short of running the damage across every site at once. It was
written up carefully, with a mechanical reproduction, and then it sat. Its own file argued the
danger was contained, because the scheduled jobs that trigger the repair were switched off at
the time.

There is a longer history behind it. The same mechanism was noticed in July, from the harmless
direction — a button that *should* be able to point at contact was never allowed to — and that
case was closed with the right fix named and not built. The destructive direction went
unnoticed for another three weeks.

## What we've done

We found that one list — "places a button should not send people" — was being used to answer
three different questions, and was only ever evidence for the first of them. It is right that
the system should not *invent* a link to the contact page. It does not follow that a contact
link already on the page is untrustworthy, and it certainly does not follow that a live button
pointing at a real contact page should be reported to a human as a defect.

The fix separates them, and rests on something the code already guarantees: the system is
*incapable* of choosing the contact page as a destination. So a button pointing there, on a
site where that page really exists, must have been put there by a person. That gives us the
"who wrote this" information the original bug report said would need a new database field and a
data migration to obtain.

We also found and fixed a second copy of the same fault that nobody had filed: two different
places in the code write these links — one when repairing a page, one when rebuilding it — and
only the repairing one had been reported. Fixing that alone would have left the same buttons
dying at the next rebuild.

Two things are worth recording about how we got here. An adversarial review very nearly killed
the design with a detailed, well-cited objection that turned out to be reading a historical
seed file rather than the live database; the check that settled it was one query. And while
checking a page yesterday evening we caught the defect actually happening — a button reading
"Start a Conversation" repointed to a password-strength calculator at 19:11, with the repair
job that did it named in the same second. That is the first time this damage has been
attributable rather than merely counted.

## Where we are now

The fix is committed and covered by tests that we deliberately broke three times to prove they
were doing work. It is **not yet live** — it ships with the next fleet release, and until then
the bug stays open, because a fix that has not shipped is still broken in production.

Eighteen buttons fleet-wide are still in the vulnerable state, thirteen of them reachable by
the code we changed, eight of those on webdesign.uk including its homepage. The scheduled jobs
that trigger the damage are currently switched on.

One honest limitation: if a button's text shares even a single word with some other page, the
system will still prefer that page over the contact page. We have not closed that, because the
same mechanism is what repairs genuinely wrong buttons.

## Where we're going

Once the release goes out, we prove it on a real page rather than trusting the tests — a
before-and-after snapshot of one affected page, a repair run at it, and a second link on the
same page that *should* change, so that "nothing changed" cannot be mistaken for "nothing ran".

Then there is a decision waiting. There are 149 cases where the system correctly identifies a
wrong button and correctly says it should point at the contact page — and then cannot do it,
because the repairer is not allowed to choose that page. Fixing that undoes the reasoning this
fix depends on, so it needs the database field the original report proposed. It is now the
second bug to need it, which is probably the argument for building it.

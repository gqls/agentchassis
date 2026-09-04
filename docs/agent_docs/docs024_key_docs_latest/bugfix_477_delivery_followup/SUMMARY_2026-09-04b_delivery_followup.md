# SUMMARY 2026-09-04b — the blocker is closed, the interval is ruled, and the thing is built and switched off

Supersedes `SUMMARY_2026-09-04_delivery_followup.md`, written earlier the same day. That one is kept
unedited: it describes a blocker as open and an interval as undecided, and both changed within hours.

## What we are trying to do

Make one sentence on a customer-facing page true. When we deliver a website the customer gets a link
to a page with a button, and the page said: press this when you have moved, and you will not get any
more reminders. We do not send reminders. Nothing ever has. So the button asked a customer to act,
told them what it saved them from, and the thing it saved them from did not exist.

## Where we have come from

The bug was filed the night before, after the owner asked in passing whether we might repeat the
hosting instructions in a follow-up email a week or so after delivery. The question named a mechanism
nobody had built. It was the third defect of that shape in one day — all three found by a person
reading the words and asking whether they were true, and none of them findable by anything automatic.

## What we have done

**The copy is honest on all three surfaces.** Two were pages, deleted rather than reworded, because
the reason to press the button *was* the false part. The third was the delivery email itself — the
one that actually reached a human — and the delivery lane fixed it within an hour of being told.

**The follow-up email is built, reviewed and switched off.** The engineering went almost entirely
into one property: it must be impossible to email the same customer twice. A site is claimed in a
single statement only one caller can win, and that same statement re-checks whether the customer has
already pressed the button — so somebody who presses it in the gap between selection and sending does
not get the email. That check is the first time anything in the system has ever read what the button
records.

**And the blocker that stopped it reaching anybody is closed.** We had no durable record of who a
delivered site was delivered to. Worse than that, the column a person would naturally read holds a
different, plausible, wrong address — so anyone answering a support question would have been
confidently misled, with no blank to warn them. There is now a proper record, written in the same
statement that hands the site over, so a delivery cannot happen without it. The one historic address
was recovered from the system's own log with about four and three quarter hours before that log
discarded it.

**Three council rounds, all approved**, including the architecture seat ruling that the change to the
handover was a point fix and needed no formal design review.

## Where we are now

One decision away from working, and the decision is a human one rather than a technical one.

The owner has ruled the interval: **three days**. The letter carries two things from him performing
the hosting steps himself — that it takes about forty minutes and that slowness is not a fault, so it
reads "here is what you need" rather than "you have not done this" — and a paragraph telling the
customer to open their new address in a private window. That last one matters more than it sounds: a
newly uploaded site can be visible only to the person who uploaded it while looking completely normal
to them, so a customer could confirm in good faith with a site nobody else can reach, and the confirm
button would then silence the one message that might have told them.

The schedule is seeded and **disabled**. The interval being settled is not the same as somebody
deciding to email a real person today, and the only site the sender can currently see is our own
rehearsal site — whose address is the owner's.

## Where we are going

**Switch it on, when he says so and knows the first run reaches him.** That is one command, after two
migrations are applied and a release has carried the code.

**Then put the stronger wording back** — on the page and in the email — and delete the test that
guards it, in the same commit. That test exists precisely so restoring the promise has to be a
deliberate act by someone who notices they are making one.

The wider thing this lane leaves behind is not the email. It is that the estate can now answer "who
did we send this site to?", which support, refunds and renewals all need and none of them could ask
before.

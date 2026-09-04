# SUMMARY 2026-09-04 — the confirm button, and the follow-up email that did not exist

> **SUPERSEDED the same day by `SUMMARY_2026-09-04b_delivery_followup.md`.** Kept unedited, because
> the series is the record and what we believed at this point is the thing a later reader cannot
> rederive. Two of its statements went out of date within hours: the blocker it describes as open
> ("the follow-up cannot email anybody") is closed, and the interval it lists as the owner's to decide
> has been decided. One sentence WAS corrected in place — the delivery log's retention — because it
> was factually wrong rather than merely overtaken; see the b summary for what was wrong with it.

## What we are trying to do

Make one sentence on a customer-facing page true. When we deliver a website, the customer gets a link
to a page with a button that says, in effect, *press this when you have moved your site off our
hosting, and you will not get any more reminders about it.* We do not send reminders. Nothing in the
system ever has. So the button asked a customer to do something, told them what it saved them from,
and the thing it saved them from did not exist.

Two ways to make that true: stop saying it, or build the reminder. The owner had already asked for
the reminder — in passing, as a suggestion — so the work is both: stop the untruth today, build the
thing that makes the stronger wording honest, then put the wording back.

## Where we have come from

The bug was filed the previous evening by the delivery lane, after the owner asked whether we might
repeat the hosting instructions in a follow-up email about a week after delivery. The question turned
out to name a mechanism nobody had built. It was the third defect of that same shape found in one
day, all three by a person reading the words and asking whether they were true: an email promising
instructions inside a download that has none, instructions promising a site opens by double-click
when it does not, and this. Nothing automated found any of them, and nothing automated could have.

## What we have done

**The copy is honest.** The false sentence is deleted from both pages — deleted, not reworded,
because the reason to press the button *was* the false part and inventing a replacement reason is the
same defect in different words. A test now fails if anyone puts it back while there is still nothing
to send, so restoring it has to be a deliberate act.

**A third instance turned up, and it is the only one a human has actually read.** The delivery email
itself carried the same promise. Unlike the two pages, that text lives in the database rather than in
the program, so it was fixable the same afternoon with no release — and the delivery lane, told about
it, shipped the fix within the hour.

**The follow-up email is built, and switched off.** The care went almost entirely into one property:
it must be impossible to email the same customer twice. A site is *claimed* in a single database
statement that only one caller can win, and the same statement re-checks whether the customer has
already pressed the confirm button — so somebody who presses it in the gap between us choosing to
write and the email being sent does not get it. That re-check is the first time anything in the
system has ever read what the button records. Proved against the real database, including a control
showing the guard is what refuses rather than a fixture that could never have passed.

**Both halves went through the reviewer council and both were approved.** Six advisory objections
came back; four were right and are acted on, two were wrong because the summary I gave the reviewers
showed only a fragment of one file.

## Where we are now

The follow-up cannot email anybody, and finding out why is the most useful thing this work produced.

**We have no durable record of who any site was delivered to.** The rule — correctly — is that a
customer's address comes from the order they placed. The one site we have ever delivered was our own
rehearsal, sent to an address typed in by hand, so there is no order and no address. The only other
copy of it lives in the delivery run's own log, and those are discarded a day after the run finishes;
a follow-up due days later would be looking for something already gone.

That was found by running the selector with the calendar deliberately relaxed so it *had* to return
the one site we know about. It returned nothing. Without that check, the follow-up would have run
switched on, selecting nobody, looking healthy indefinitely — because a selector with nothing to do
and a selector that can never work produce exactly the same silence.

The fix is small and belongs in the delivery step rather than in ours: record the address at the
moment we hand a site over. It has been written up and handed to the lane that owns that code.

## Where we are going

Three things, and two of them are the owner's.

**He decides the interval.** "A week or so" is not a number, and the code refuses to run without one
rather than guessing — so nothing happens until he says.

**He needs to be told before it is switched on**, because the only site it can currently see is our
rehearsal site and the address on it is his. The first real run emails him.

**Then the stronger wording goes back**, on the page and in the email, and the test that guards it is
deleted in the same commit — deliberately, which is the whole reason the test exists.

Underneath that sits the real repair: once a delivery records who it went to, this stops being a
mechanism with no population and becomes the thing the owner actually asked for.

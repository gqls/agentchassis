# README — where we are (bugs_open/308, CTA destination provenance)

Plain prose, append-only, newest at the bottom.

## 2026-08-22 — what this is, and what I checked before starting

The bug in one sentence: **a check spots a button whose words say "Contact our supply team"
but whose link goes to a break-even calculator, correctly says so, asks for it to be pointed
at the contact page — and the thing that does repairs is physically unable to point anything
at a contact page, so it reports success and changes nothing, and the check finds the same
button again next time.**

Two things that surprised me when I went looking.

**First, it has got worse, not better.** When the bug was written on 17 August there were
149 of these. Today there are 200. More telling than the growth: 112 of them sit on jobs the
platform has marked **complete**. Those are not jobs waiting to run. They are jobs that ran,
declared victory and left the button pointing at the calculator.

**Second — and this is the part I want to flag, because it changes what a good fix looks
like — the check already works out the right answer and nobody reads it.** When the check
files its report it writes down the page it thinks the button should point at. I grepped the
entire codebase for anything that reads that field. Nothing does. The repair job is handed a
one-word reason ("the CTA links are stale"), and then goes off and works the answer out
again from scratch, from a shorter list of pages than the check used. So the two halves are
not merely disagreeing by accident; the half that knows the answer is not being asked.

That is the same shape as another open bug (071), where a gate detects every broken link on
a page and then discards the finding. So I think the durable fix here is not just "let the
repairer see contact pages too" — it is closing the gap where one part of the system
computes an answer and the next part throws it away.

**A caution I have put in the notes so nobody trips over it later.** The check has not run at
all for three days — it last produced anything on 19 August. So the number 200 is a
stock-take, not a rate. If we fix this and then re-run the query, it will still say 200,
because nothing is looking. Any claim that the fix worked has to come from deliberately
making the check run and then looking at an actual live page, not from watching that number.

**On not treading on anyone.** The bug is signposted to an existing piece of work (a "CTA
target content pass"). I read that work's plan in full: it is about rewording the button
*text* so the existing machinery picks better targets, and it lists the change I need as an
open question it has not taken. So they are two different jobs, and I have opened this one as
its own lane and will write into both the bug file and their notes rather than starting a
rival. I also checked that no other session is part-way through editing the files involved —
they are clean.

**The direction is already decided, so I am not choosing it.** The owner ruled on 18 August:
build a proper record of where a link came from, rather than continuing to *infer* it from
the fact that the machinery "could never have produced a contact link" — which is exactly the
assumption that has to be given up to fix this. The owner also ruled: no new switches that let
other agents opt out of the rule. I have handed both of those to the planning step as hard
constraints rather than preferences.

Next: a plan, then the council, then code.

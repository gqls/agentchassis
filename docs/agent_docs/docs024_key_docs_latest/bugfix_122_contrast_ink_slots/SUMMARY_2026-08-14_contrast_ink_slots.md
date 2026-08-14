# SUMMARY — bug 122, contrast and ink slots. 2026-08-14.

*The second summary this workstream has written. The last was 2026-08-10, when the first
page measured clean. This one exists because the answer to "who closes these tickets?"
turned out to be a different answer than the one we were pursuing, and that is worth
saying out loud.*

## What we're trying to do

Make the text on our sites readable, and make it stay readable without anyone watching.

The original fault: the renderer builds each site's colours from a small palette, and some
components used the brand colour as a *text* colour. On a site whose brand is dark navy and
whose background is nearly black, that text is invisible — a contrast ratio of about 1.06
where the standard asks for 4.5. Every input was individually correct; only the combination
failed. That is why fifty automated checks missed it: not one of them rendered a page and
looked at it.

The second half of the job, which is where we now are, is the "without anyone watching"
part. Finding the failures automatically was solved last week. **Closing them again when
they are fixed was not**, and that is what this period was about.

## Where we've come from

Four days ago we had a working detector and a broken disposal route. The weekly browser
check was live and finding real faults — thirty-four on one site's interior pages within
seventy seconds of being switched on. But the findings flowed to a repair agent we had
separately proven could mark work "complete" without doing it. So we had a machine that
produced true tickets and no trustworthy way to retire them.

The interim answer was to **park** them. Two hundred and twenty-six findings across the
fleet were set aside deliberately, held in a state where nothing would act on them, so that
a known-faulty repairer could not chew through them and report success. That was the right
call and it was always a holding position: a parked ticket is a debt, and the pile only
grows while the detector keeps running.

The obvious next move was to fix the repairer. We did not do that, and the reason is the
most useful thing in this period. A repair agent that changes colours is a large, risky
piece of work, and it would still leave the underlying asymmetry: something has to *notice*
when a fault is gone. We were about to build a second machine to verify the first.

## What we've done

**We taught the checker to close its own tickets.** When the weekly browser check
re-measures a page and the fault it filed last week is no longer there, it now retires that
ticket itself, citing what it measured as the evidence. No repair agent, no verifier, no
second machine. The thing that is already looking at the page is the thing best placed to
say the problem has gone.

This went through the review council and was approved on substance — thirteen reviewers,
none blocked, five advisory objections all answered. Two of those objections found real
things, and one found a genuine gap in our own evidence rather than in the code, which is
exactly what the review is for. It has been live in the running system since yesterday and
was re-verified this morning against a fresh build.

**We found that a promise we made in July was not true.** When we introduced the
"guaranteed readable" colours, we said they preserve each site's brand character. Another
team measured all sixteen sites and found they do not: the mechanism always returned the
site's ordinary body-text colour. The elements we repaired were genuinely invisible before
and are genuinely readable now — that part stands — but our stated reason for it being safe
was wrong. We recorded that as a correction rather than quietly moving on. That team has
since repaired the derivation properly, in two rounds, and it is now live.

**We reviewed that repair twice and both rounds found something.** The first round would
have re-broken an element we had fixed the week before: it checked its new colours against
the background the palette *declares*, while our page checker measures the background a
visitor actually *sees*, and those differ because some panels are semi-transparent. The gap
was about ten times the safety margin the method leaves itself. That is now fixed at the
cause, with the numbers pinned in a test so they cannot drift silently.

**We found why one stubborn failure would not go away, and it is a different kind of
fault.** A button on several sites is meant to be white with the brand colour as its label.
But on some sites that "brand colour" is stored as a *gradient* — and a gradient is not a
colour. The browser discards the instruction entirely and the text falls back to inheriting
white, on a white button. The safety net written into that line never runs, because it only
helps when a value is missing, not when it is present and of the wrong sort. Sixteen of the
seventeen tickets of that shape across the estate are this exact fault, confirmed against
what the checker actually measured. Nothing we have built can fix it, and neither can the
colour repair — it needs the component to stop asking for a background colour to paint text
with.

## Where we are now

The loop is closed in principle and untested in practice, and it is worth being precise
about that distinction.

Everything is live and verified in the running system: the ticket-closing mechanism, and
both rounds of the colour repair. The two hundred and twenty-six parked findings are
untouched — none has closed, none has been acted on, and the pile has not grown.

**The mechanism has never actually run.** It cannot until Monday afternoon, because each
site is re-checked on a seven-day cycle and the whole fleet was measured on the tenth. That
is the design working, not a fault, and we have deliberately not forced it. Monday at
14:54 the robotics site comes up first, and we have written down in advance exactly what
must happen: on one page, two tickets must close and a third must stay open, because we
know a past fix addressed two of them and not the third. If all three close, the mechanism
is closing things too eagerly and we stop. That is a distinction no count of closures can
draw, which is the whole point of setting it up that way.

The colour repair is live but **asleep** — no site has taken the new colours yet, because a
site only picks them up when it rebuilds, and nothing is scheduling that. That is
deliberate on the other team's part: they intend one site first, and your decision before
widening.

## Where we're going

Monday's test comes first, and everything else waits behind it. If it behaves, the parked
findings can be released to drain on their own, and we expect fewer than two hundred and
twenty-six to remain by then — that is the mechanism working, not a discrepancy.

**Two things need you, and one of them changed this morning without anyone deciding it.**

The first is unchanged from the last summary: when a component paints its own coloured band
across the page, the renderer has no correct answer to give it, because it is not measuring
the surface the text actually stands on. About two dozen failures sit there. It is a design
question, not a bug, and it has been waiting for a human since the tenth.

The second is new. The colour repair being live means every one of fourteen sites will take
new link colours the moment it rebuilds for any reason at all. Until this morning that was
protected by two separate facts — the code was not in the system, *and* nothing had
rebuilt. Now only the second is true, and nobody controls it: another team rebuilding a
site for an unrelated reason will change its link colours without intending to or noticing.
Your decision on whether you want that visual change did not become urgent, but it did
become less protected, by a deployment rather than by a choice.

Beyond that, the gradient-in-a-colour-slot fault is written up and unowned. It is small,
well understood, and genuinely separate from everything else here — a good candidate for
whoever picks up next, precisely because it needs none of this machinery.

---

> **CORRECTION, appended 2026-08-14 late (the summary above is otherwise unedited).** The second
> "needs you" item says every one of fourteen sites takes the new colours "the moment it rebuilds
> for any reason at all". Measured the same evening: **false in the safe direction.** Only the
> design agent regenerates a stylesheet; routine page rebuilds cannot move the colours. The real
> exposure is someone deliberately dispatching the design agent at an affected site — and the only
> such request in existence is ours, held pending your look. The decision is better protected than
> this summary said. Full working in the lane notes and handoff.

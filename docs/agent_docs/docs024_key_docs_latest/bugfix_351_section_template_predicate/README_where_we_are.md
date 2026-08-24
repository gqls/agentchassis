# Where we are — the calculator-reuse bug (`bugs_open/351`)

Plain-prose running log for the owner. **Append-only, newest at the bottom.** No jargon where
plain words will do.

---

## 2026-08-23 (evening) — the fix works, and we can now prove it on a real site

Short version: **the half of this bug that we fixed on Friday is now live and we have watched it
work on a real site.** The other half is a decision, not a bug, and I have not taken it yet.

**What the bug was.** We have a library of ready-made page sections. Twenty-two of them are
calculators — mortgage repayment, stamp duty, loan overpayment, that sort of thing — built for the
loan-and-mortgage site back in mid-August. The platform had a safety check whose job is to spot a
section whose text got cut off halfway through being written, so we never put a broken half-section
on a page. That check worked by looking for a particular closing tag. A calculator is built
differently and never contains that tag. So the check looked at all twenty-two perfectly good
calculators and declared every one of them broken.

The cost of that was not a broken page — it was worse and quieter. The platform simply pretended
those calculators did not exist. Every new site that wanted a mortgage calculator paid to have one
written from scratch, while twenty-two finished ones sat on the shelf unused.

**What we changed.** Friday's fix replaced "look for that one closing tag" with an actual
examination of whether the section is complete: is the markup balanced, and does it finish properly
rather than stopping mid-word. We fixed it in the shared place rather than the one place that was
complaining — there was a second use of the same faulty test that was *refusing to save* legitimate
new work, and patching only one of the two would have left them disagreeing with each other.

**What is new today.** Two things.

First, the fix has actually shipped. Both copies of the service are running a build that contains
it. Friday's notes said it was sitting inert waiting for a release, and that had stopped being true
— so anyone reading those notes was being told to wait for something that had already happened. I
have corrected that in the three places it was written down.

Second, and this is the part that matters: **we can now point at a real site using it.** Today,
loanzy.uk built three pages that each needed a calculator, and each time it picked up a
ready-made one from the library — a credit health check twice and a damage checker once — instead
of commissioning a new one. Those three calculators were originally built for a *different*
customer's site, which is exactly the reuse we wanted and could not previously get. And crucially,
no "we need a new component" job was raised for any of them. That was the test we set ourselves for
saying this is fixed, and it passed on its own, in production, without anyone nudging it.

I checked one obvious way I could be fooling myself: whether somebody had quietly edited those
calculators today, which would mean the data changed rather than our code working. They were last
touched on the 20th. So it is our change.

**What is left, and why I have not just done it.** Twenty-one of the twenty-two calculators are
still missing a label that a *second* lookup route uses. The tempting move is to fill that label in
for all of them. I have spent some time on that today and I am fairly sure **it would achieve
nothing at all** — the first lookup route already finds them by name and runs before the second one
ever gets a turn, so the label we would be adding is one that route never gets asked about. I have
written that reasoning down and marked it clearly as reasoning rather than something I have proven
by running it, because it deserves one proper check before anyone acts on it.

**One thing I found on the way that I think is a genuine, separate problem.** The library keeps a
"how often has this been used" number, and the platform uses it to prefer sections with a track
record. It turns out that number is only ever incremented on one of the two ways a section can be
chosen. Everything picked the other way is used on a real page and never counted. Today that means
**96 of our 149 sections show a usage count of zero despite being live on pages, and about 1,800
uses are invisible to it.** So a score that is meant to mean "this one is proven" currently means
"this one happened to be found by a particular route".

That is not causing visible damage today, and I have not filed it as a bug yet. It matters here
because it is a second argument against the tempting fix above: filling in those labels would put
our twenty-two library calculators into a popularity contest against near-identical copies made for
individual customers, scored on a number that measures the route taken rather than the merit. We
would systematically prefer the one-off copy over the general-purpose original.

**Two mistakes of mine today, both caught before they went anywhere, both written up.** I nearly
used a table to work out what was running on the cluster at lunchtime — it only keeps two hours of
history, so it genuinely cannot answer that, and it would have told me our fix had failed when it
had not. And I miscounted the library by three because I searched for a word rather than reading
the right column; the only reason I noticed is that my two subtotals did not add up to the total I
had printed beside them.

**Where it goes next.** The formal review round we ran on Friday died without giving a verdict — the
account had hit its usage limit that evening — so I have resubmitted it with today's proof attached
and it is working through the reviewers now. Once that comes back I will take the remaining
decision in writing, one way or the other, rather than leaving it open a third time.

---

## 2026-08-24 — the reviewers found something I had missed, and the door is now closed

Three things happened since the last entry: the formal review came back, it caught a real gap in my
own checking, and you chose to close the birth door.

**The review approved it, and the useful part was the objection.** One reviewer asked whether I had
checked every kind of component the changed function serves, not just the ones the bug was about.
I had not. The function has two branches — "tools" and *everything else* — and I had only measured
tools and page sections. Twelve other components (site headers, footers, a page head) go through the
same branch, and it turns out **six of those were also being wrongly rejected**, and had been all
along. Nobody had noticed because the symptom is invisible: the platform just quietly acts as though
the component isn't there. Re-measured across everything: 28 rescued, nothing broken.

The lesson I have written down for myself is that I checked the population the *bug* was about
rather than the population the *changed code* serves, and those are not the same set. A catch-all
branch is the easiest place for that gap to hide.

**A second reviewer objected that I had asserted something rather than shown it** — that the library
calculators were being found by name rather than by label. Fair. I had flagged it as unproven, but
flagging is not proving, and the proof was two minutes away: a label that is empty can never match a
search for a specific label, and the platform's own usage counter (which only ticks on the other
route) is still sitting at zero for those components. Both check out, and I confirmed the counter
behaves as expected on components that *were* found the other way, so the test could have failed.

**On your decision to close the door.** It is written, tested and committed, but deliberately **not
switched on yet** — a database change like this goes live the instant it is applied and there is no
build to roll back, so it waits for its review to come back. What it does is refuse, at the moment of
creation, a section component that arrives without the label the platform searches by. It refuses
only that; it repairs nothing and touches none of the existing components.

Two details I want on record because they were nearly wrong. First, it must *not* apply to copies —
the tool-deployment code makes copies without that label on purpose, so a slightly broader rule would
have broken tool deployment in production while looking perfectly reasonable in review. I proved that
by deliberately breaking my own rule and confirming the test caught it. Second, it applies only when
a component is *created*, never when one is edited — the more obvious way of writing it would have
started rejecting ordinary repairs to the twenty-five existing components.

**What I decided not to do, and why it is the more interesting half.** The tempting fix was to fill
in the missing labels on those twenty-five. I have declined it. The reason the bug file originally
gave for declining it had actually expired, which is a trap — anyone noticing that would conclude
filling them in was now safe. It is not, for two reasons I found by reading the code rather than
assuming: another part of the platform deliberately *relies* on that label being absent, and treats
its absence as a signal to be more careful about what it tells the component writer. Filling the
labels in would switch that caution off and re-open a fault we fixed separately last week. And the
platform's tie-breaking between two near-identical components has no tie-breaker at all — it would
come down to chance.

**Where it stands.** The original problem is fixed, live, proven on a real site, and formally
approved. The door-closing change is written and awaiting its own review. Nothing is outstanding
that I am aware of and not telling you about.

---

## 2026-08-24 (later) — done, and switched on

The door-closing change passed its review and is now live. I applied it, checked it by hand on the
real database, and closed the bug.

**What "checked it by hand" means here, because "it's installed" is not the same as "it works".** I
tried, on the live table, to create a section component without its label — refused. I tried to strip
the label off a real existing component — refused. Then the two that had to *succeed*: an ordinary
repair to one of the twenty-five older components with no label went through fine, and creating a
properly labelled component went through fine. Everything was rolled back afterwards, so nothing was
actually changed. That third check is the one I cared about most: it is the difference between
closing a door and locking twenty-five people in a room.

**The reviewers earned their keep again.** One pointed out that I had only blocked the *creation* of
an unlabelled component, and nothing stopped someone later *removing* a label from a good one —
same problem, different door. It asked whether that was a real risk or a theoretical one, and said
it couldn't tell. I could: nothing in the codebase ever removes one of those labels, so it was
theoretical. I closed it anyway, because it cost one line. A second reviewer caught that re-running
the change would have failed with an error instead of doing nothing — a small thing that becomes an
unpleasant surprise at the wrong moment.

**One thing I found along the way that I have written up separately.** The platform keeps a
"how often has this been used" number against each component and uses it to prefer well-proven ones.
That number is only ever incremented on one of the two ways a component can be chosen. Ninety-six of
our hundred and forty-nine components show zero despite being live on pages, and about eighteen
hundred uses are invisible to it. So a score that is meant to mean "this is proven" currently means
"this happened to be found the counting way". It is not breaking anything today and I have not fixed
it — it is filed as its own item with the evidence and, importantly, with a list of the things I
have *not* measured, including whether it has ever actually changed which component got picked. That
last one is the question worth answering before anyone touches it.

**Where this leaves the original complaint.** Sites can reuse the calculators we already own instead
of paying to have near-identical ones written; that is proven on a real site, not just in a test. New
components can no longer be created in the state that caused it. The twenty-five older ones are
deliberately left as they are, for reasons I have written down twice because the original reason for
leaving them expired and I did not want the next person to reach the opposite conclusion from a
stale note.

One loose end that is not mine and not this bug: the `stylesheet_gutted` check this lane built back
on the 22nd is still waiting to be released.

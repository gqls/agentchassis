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

# SUMMARY — the calculators now have to prove they are right, and two of them cannot

**2026-08-08.** Written to be read aloud.

## What we are trying to do

loanandmortgagecalculator.co.uk is a site whose entire product is twenty-three
calculators. People type in a loan amount or a house price and act on the number
that comes back — what they can afford, which offer is cheaper, how much tax
they will owe on completion. The site is only worth anything if those numbers
are right.

So the goal of this piece of work was narrow and blunt: build something that can
tell us whether each calculator is actually correct, as opposed to merely
working.

## Where we have come from

Over the past ten days this lane adopted the site into the framework, rebuilt
its structure, and put a lot of checking around it. That checking got good at
one thing: noticing when we break something. We record what every calculator
answers on a given day and compare against it afterwards, so a rewrite that
changes an answer is caught immediately.

On the 6th the owner asked whether we had "a comprehensive check on the
calculators that they produce validated output", and the honest answer was no.
The distinction matters more than it sounds. Recording today's answers and
checking they don't change tells you nothing about whether today's answers were
ever right. If a calculator has been wrong since the day it was written, the
recording captured the wrong number faithfully, and every check since has been
confirming the bug with a green tick. We had a very reliable way of preserving
mistakes.

That gap was written down as an open question and left. This is the work that
closed it.

## What we have done

We built an independent second opinion. The expected answers are computed from
first principles — the standard loan repayment formula, the stamp duty tables
published by HMRC, the plain definitions of yield and loan-to-value — in code
that has never read the website's own workings.

The independence is the whole value, and it is easy to lose by accident. To type
into a page you need to know which box is which, and the quickest way to find
out is to read the page's code — one line above the sums. So the harness reads
each page the way a person does: the visible label next to each box, the text on
the button, the caption above each answer. The site's own code was not opened
until a check had already failed, at which point reading it is diagnosis rather
than copying. Both of the problems below were found before any of it was read.

It then drives the real pages in a real browser, at deliberately chosen numbers —
the exact edges of tax bands, a 0% interest rate, a one-month term — because
those are where mistakes live and where a general-purpose checker never lands.

One hundred and seventy-six checks across eighteen calculators. The remaining
five are scoring tools with no external right answer; we check what must be true
of them regardless — answering worse cannot score better, percentages stay
between 0 and 100, saved data comes back unchanged — and we say plainly that
this is weaker evidence, rather than letting the green ticks imply otherwise.

## Where we are now

**Sixteen of the eighteen are right.** That is worth saying first, and it is not
a formality: those sixteen were checked at their boundaries by an oracle that
had every opportunity to convict them.

**The stamp duty calculator is running a tax rule that expired sixteen months
ago.** First-time buyers used to get relief up to £625,000; that was temporary
and ended on 31 March 2025. Since then relief stops at £500,000, and above it you
pay ordinary rates on the whole price. Our page still uses the old cap, and for
anyone buying between £500,000 and £625,000 it quotes exactly £5,000 too little.
The page's own text, immediately above the calculator, correctly says the
temporary period ended. The words are right; only the arithmetic is out of date.
There is a smaller second version of the same thing: it charges the buy-to-let
surcharge below £40,000, where no surcharge applies.

**A 0% interest rate breaks six of the seven loan calculators.** That is not a
contrived input — interest-free credit cards, employer loans and manufacturer 0%
car finance are ordinary products. And there is a clean explanation. The site
has one shared file containing the loan formula written properly, zero case
included. Every mortgage calculator uses it and every one passed. Every loan
calculator has its own private copy, and not one of those handles zero. It is
one mistake copied six times, not six mistakes.

Three of them display "£NaN", which at least looks broken. Three quietly leave
the previous answer on screen, which does not. The clearest demonstration needs
no code at all: type the same three numbers into the same three boxes, arrive by
a different route, and get a different answer — £143.47 one way, £429.81 the
other. The one to fix first is the comparison tool: give it a 0% loan against a
5% loan and it declares the 5% one cheaper, confidently, in a green box.

**We have changed nothing.** These are tax and credit figures a real person
budgets against, and a changed answer is a changed claim. Both are written up
with their evidence and their sources, waiting on a decision.

**And the harness argued with itself, which is the part I would want known.**
It accused the rate forecaster of miscalculating; the forecaster was right and my
formula was the naive one. Its own controls found a bug in its number parser. Two
"defects" in the scoring tools turned out to be the harness clicking a "Start
Over" button and not waiting for a one-second save. Most instructive: the
mechanism I built to explain findings kindly was, on the first run, quietly
explaining away the stamp duty finding as a difference of opinion. All four are
written down where the next person will find them, because a checker nobody has
seen fail is not evidence of anything.

## Where we are going

1. **A decision from the owner on the stamp duty numbers.** Everything else can
   proceed without it.
2. **Fix the zero-rate family by deletion, not by patching.** The shared formula
   is already correct and already loaded by eleven pages; the right change is to
   remove six private copies, not to repair them.
3. **Then hand the checks to the platform.** There is a mechanism that turns
   these expectations into criteria the platform re-runs on its own schedule,
   for ever, unprompted. We deliberately have not used it yet: doing so before
   the fixes would write today's wrong answers into the acceptance record as the
   expected ones and then defend them.
4. **And the same question for the sibling sites.** loancalculator.co.uk and
   mortgagecalculator.co.uk carry overlapping tools, checked the same way, with
   the same blind spot.

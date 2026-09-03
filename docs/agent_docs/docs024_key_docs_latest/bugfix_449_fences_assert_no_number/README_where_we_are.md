# README — where we are: the calculators nobody checks the arithmetic of

The owner's running plain-prose log for bug 449. Append only, newest at the bottom.
No jargon where plain words will do.

---

## 2026-09-03, midday — picking this up, and what it actually says

Yesterday evening the mortgage-calculator lane was asked to "verify the tools". What it
found, and filed as bug 449 at 22:33, was that where the platform's verification runs at
all, **it is not checking what the tools are for.**

Here is the plain version.

When the system builds a calculator, it also writes a short test plan for it — a "fence" —
that says what has to be true for the calculator to count as working. Something writes that
fence automatically. And what it writes, every single time, is a list of checks about the
calculator's *health*: does the page load, does it throw errors in the browser console, does
it fit on a phone, and does *something appear* when you click the button.

Not one of those checks asks whether the number that appeared is right.

So a mortgage calculator that confidently tells a customer their monthly payment is £303
when it is really £430 passes every test we have, and its record says PASSED.

## The size of it, measured again today

I re-counted this morning rather than trusting yesterday's figure, and I am glad I did,
because it has moved.

Of the test plans the generator has written, there are now **186**, and **115 of them
assert no expected value of any kind** — up from 170 and 107 yesterday. The newest one was
written **today**. So this is not an old backlog waiting to be tidied up. It is a tap that
is still running: every calculator the fleet builds today gets another fence that cannot
tell a right answer from a wrong one.

The sharpest cut of the number is this one: **55 of those fences fill in the calculator's
input boxes and then check nothing about what came out.** The fence itself says "this thing
takes numbers in". It just never looks at what comes back.

## Why it happens, and it is not a hard problem

The platform *does* have a check that compares actual numbers. It is called
`computed_values`, it works, and eight of this site's operator-written fences use it. It
has simply never been mentioned to the thing that writes the fences.

I read the generator's instructions in full. They list four checks it must always include
and one it may add, and they finish with the sentence "No other check type exists". So the
correctness check is not a candidate it weighs up and rejects — it has never heard of it.

## The catch, which is why this is not a five-minute fix

The obvious move is to tell the generator about the check. That is right, and it is not
enough, and the bug says so itself.

The comparison check works by recording what the calculator printed when it was known to be
working, and then defending that number against later changes. That is a good guard against
someone breaking a working calculator. It is useless at *birth*, because a brand-new
calculator has never been known to be working — so recording what it prints on day one just
carves today's mistake into stone and then defends it.

That is not a theoretical worry. We have shipped exactly that twice. One of them was a
stamp-duty calculator applying a first-time-buyer threshold that had expired, and it sat
there certified green for **sixteen months**.

So the real question this fix has to answer is: *where is the right answer allowed to come
from, if not from the calculator itself?* The mortgage lane has already built the honest
version of that — it re-derives every number it pins from a published formula, or from this
site's own registered legal facts with their GOV.UK citations attached, and it labels how
strong each source is. That machinery exists. It just lives in one lane's folder rather
than in the framework.

## What I am doing about it

Three things, in this order of stubbornness:

1. **Make the gap visible.** Right now nobody can see which calculators are unchecked
   without running the query I ran. That should be a standing report, not an archaeology
   exercise.
2. **Stop the tap.** The system has exactly one door that every automatically written test
   plan goes through. A rule at that door can refuse to accept a fence that fills in a
   calculator's boxes and then checks nothing — which is the version that cannot be
   forgotten, unlike an instruction in a prompt.
3. **Teach the generator the check** — with a rule about where the expected number is
   allowed to come from, so we do not fix "checks nothing" by shipping "checks the wrong
   thing, forever".

And one discipline over all of it, which the bug is right to insist on: **a fix that only
adds checks which pass is indistinguishable from no fix.** Before I believe any of this, I
will take a calculator that currently passes, deliberately break one number in it, and
confirm that it now fails. Today it would still pass.

## One thing worth saying about how the record reads

Until this is fixed, a "PASSED" against a calculator does not mean what it sounds like. On
this site right now, two tools are passing on fences that assert no value at all — that
verdict means the page loads and something appears when you click. A third is passing four
real arithmetic checks and nothing else — no load check, no error check — and on mobile all
four are skipped, so its mobile pass is a pass in which nothing ran.

Not one tool on that site is verified for both correctness and health at once. That is the
sentence I would not want the owner to learn from a customer.

---

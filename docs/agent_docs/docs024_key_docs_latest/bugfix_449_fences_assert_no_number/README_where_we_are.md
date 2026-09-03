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

## 2026-09-03, early afternoon — what shipped, and what I deliberately did not fix

Three things are done, and one important thing is deliberately still undone.

**What is done.**

The first is the one I care most about. When a calculator passes its automated check, the record
now says *what kind of pass it was*. If the fence never compared a number, the verdict says so, in
the note a person reads and in the field a program reads. Nothing changes about which tools pass —
a tool that passes today still passes. What changes is that "PASSED" can no longer be quoted as "the
arithmetic is right" when nobody checked the arithmetic.

That is worth more than it might sound, because it fixes **every existing calculator at once**.
There are 186 of these test plans; 116 check no number. Repairing them one at a time is weeks of
work per site. Making the verdict honest took one change and covers all of them today, and it can't
rot: the label is worked out from the test plan each time the check runs, so a plan that gets weaker
automatically gets a weaker verdict, with nobody having to remember anything.

The second: the single door that every automatically-written test plan goes through now leaves a
note when a plan fills in a calculator's boxes and then checks nothing about what comes out. It
**records** rather than **refuses** — refusing would leave the tool with no test plan at all, and a
tool with no plan is checked by nothing, which is worse than being checked badly.

The third: a standing report, so nobody has to run my query by hand again. It deliberately leads
with *how many blind plans were written in the last week*, not the total — because the old ones
don't repair themselves, so a total would read as "no improvement" for a month after a fix that
worked, and would get quietly abandoned.

**What I deliberately did not fix, and this is the real decision of the day.**

I did not teach the system to write better test plans. That is the obvious move and it is the one
thing I have held back, for two reasons.

The check that compares numbers works by recording what a calculator printed when it was known to
be working, and defending that. At birth, nothing is known to be working — so recording today's
answer just carves today's mistake into stone and then guards it. We have done that twice, once for
sixteen months on a stamp-duty calculator using an expired threshold.

And separately: another open bug means test plans are currently pointing at page elements that have
moved. A number check that can't find its element **fails**, by design. So switching number checks
on today wouldn't make wrong calculators fail — it would make *right* ones fail, loudly, and send an
automated fixer to rewrite arithmetic that was never wrong.

So the order is: make the record honest first (done), fix the moved elements (another lane, in
flight), then teach the generator — with a firm rule that if it cannot work out the right answer
from something other than the calculator's own code, it must write **no** number check rather than
a guessed one.

**Two mistakes of mine, both written down.** One was a shell slip that silently re-pointed my
searches at the wrong folder — dangerous because the quiet version of it returns "nothing found"
rather than an error, and I had been using exactly that kind of search to prove a *negative*. The
other is more interesting: I built a self-check into the new report, to stop it reporting "all
clear" when it had actually gone blind — and then wired that self-check to a team's name, which
changed a few hours earlier. It survived by luck. It is now wired to the thing it is actually
testing instead.

**One thing found for someone else.** While checking my own work against the shared codebase I
found a test that another lane's change had left failing this morning. I have not fixed it — it sits
inside their work — but I have told them, including the part they couldn't see: the verification
step they wrote down for themselves doesn't select that test, so their own check passes while the
shared build is red.

---

## 2026-09-03, mid-afternoon — it's running, the reviewers approved it, and I want to be careful about what that means

The new chassis went out and the two code changes are in it. I checked at the binary rather than
trusting the version number, and I'm glad I did, because **an earlier build four hours before had
the same version bump, the same restarted servers, and did not contain my change at all.** Three
things that all looked like proof — the version went up, the servers restarted, my work was in the
shared codebase — and together they proved nothing. The only thing that answered it was asking the
running program directly whether it contained the new behaviour, with a deliberate wrong answer
thrown in to check the question was capable of coming back negative.

The internal reviewers also approved it, on the second round. The first round said *revise*, and it
was right to. One reviewer spotted that my new record-keeping write could have been silently
rejected by the database and I'd never have known, because I'd deliberately made that write
non-fatal so it couldn't break anything more important. I checked: it was fine — but **only because
of which value I happened to be using**, and I hadn't verified it. That's a fair catch, and the
version now on record says which value it uses and why, and warns what to look at first if the
thing ever goes quiet. Two of the six points they raised changed the code. That's a good return on
half an hour.

**Now the careful bit.** "It's running" is not the same as "it's working". Nothing has actually
exercised the new behaviour yet: no calculator has been built and no calculator has been checked in
the few minutes since the new version went out. So there's no example on record of a verdict
carrying its new honesty label. I've written down exactly what to look for and — importantly — how
to tell "nothing has happened yet" from "it's broken", because both look like an empty result and
they want opposite responses. I'd rather hand that over as an open question than tell you it works
on the strength of it being installed.

**The daily report, though, is genuinely running and has produced real numbers.** It graded 241 test
plans across the whole estate: **58 of them fill in a calculator's boxes and then check nothing at
all about what comes out.** Thirteen of those were written in the last week. So this isn't a
historical mess — it's a tap that's still running, and now there's something watching it every
morning and writing down what it saw, including on the days it sees nothing wrong, so a silent day
can't be mistaken for a clean one.

**What I have not done, and would like a decision on eventually.** I still haven't taught the system
to write better test plans. That's the actual cause. I've explained why in the last entry — briefly:
recording what a brand-new calculator prints just protects today's mistake, and separately there's an
open bug making test plans point at moved page elements, so switching number-checks on now would make
*correct* calculators fail loudly. I've asked the two teams who own those pieces and neither has
answered yet. Until they do, doing it would be guessing, and guessing here produces something worse
than what we have: a wrong number defended by a green tick.

Everything is written up for whoever picks it up next, and the handoff leads with the distinction
above rather than burying it.

---

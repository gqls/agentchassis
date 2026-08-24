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

## 2026-08-22, later — the fix is written, and the review sent it back once (which was worth it)

**What the fix actually does, in one go.** Today the system works out whether it may rewrite a
button's link by *reasoning*: "we could never have produced a link to the contact page, so if a
link to the contact page is sitting there, a person must have put it there." That reasoning is
sound right now and it is exactly what has to stop being true in order to fix this bug — because
fixing it means letting the machinery point buttons at contact pages. So instead of reasoning, we
now **write it down**: whenever the machinery sets a button's destination, it records, next to the
link, which link it set. A link counts as a person's when it is real *and* that record does not
name it.

The detail that turned out to matter is that the record stores **which** link, not just "we set
this one". Consider someone editing a button's destination by hand. The old record survives that
edit. If it only said "we set this", the machinery would read it and feel entitled to overwrite the
person's new choice — which is the previous bug in this family all over again. Because it names the
old link, it simply no longer matches, and the person's edit is correctly left alone. That one
choice is why almost every other part of the system needs no change at all.

**The review sent it back, and it was right to.** I put the plan through the reviewer council.
Fourteen reviewers; eight approved, four objected, and the verdict was "revise". Three of the
objections found real problems:

- One reviewer noticed my written description said the phone-number branch should get a record,
  while my code did not give it one. **The description was wrong and the code was right** — a
  `tel:` link is always a person's, and recording it as ours would let the machinery replace a
  genuine "call us" button with a link to a tool page. Fixed the description.
- Another spotted a guard in my plan that could never fire — dead code. Correct; removed.
- A third said, bluntly, that I had listed "I haven't checked whether the save actually keeps this
  record" in my own risks, and that *"'owed' is not a control on a mechanism whose whole purpose is
  a record reaching the database"*. That was the fair hit of the round. I had measured the outcome
  (sixteen rows in the live database already hold a similar undeclared value, so the save clearly
  keeps them) but had not read the code that does the saving. I read it. It keeps everything.

**Two mistakes of my own, both caught by deliberately breaking my own code.** The practice here is
that you don't get to say a safety check works — you sabotage it and confirm the right test
screams.

1. My first version would have **caused the very freeze it exists to prevent**. Most of these
   components have *two* buttons. The save merges records shallowly, so writing a record for the
   first button silently threw away the second button's — after which the second button would look
   like a person's work and be stuck for ever. Fixed.
2. I had put the repair for that in the surrounding loop. When I sabotaged it — deleted the call
   entirely — **every test in the repository still passed**. So I moved it inside the two functions
   themselves, where the tests actually reach it, and confirmed that deleting it now fails. I also
   found a second helper I'd written was doing nothing at all, and deleted it rather than ship a
   piece of machinery nobody exercises.

**And one thing I got wrong three times before checking once.** I asserted — in a code comment, in
the submission to the council, and in a test — that the system treats `/contact.html` and
`/contact/index.html` as the same page. It does not. The test caught it the first time I ran it.
Then, while writing the correction, I described a sabotage-test as "(verified)" without having run
it; it turned out not to work the way I said. Both are logged. The honest lesson is that being
mid-correction felt like being careful and wasn't.

**Where this leaves us.** The recording half is written, tested, committed, and will go out with the
next build — it changes no behaviour on its own, which is deliberate: it is the thing that makes the
*next* step safe. The next step (letting buttons point at contact and about pages, and making the
detector and the repairer share one list of candidate pages) is a separate submission that should
not start until this one is confirmed running on the live machines. Round two of the review is
queued now.

One caution I keep repeating because it will otherwise cause a false result: the check that finds
these broken buttons **has not run since 19 August**. So the count will sit at 200 whether we have
fixed anything or not. Proving this worked means deliberately making the check run, and then
looking at a real page in a browser.

---

## 2026-08-23, afternoon — the block was never real, and Phase B is written and shipped

**First, a correction to what I wrote this morning.** I said the estate's AI budget was exhausted
until 1 September and that Phase B could not start until then. That was wrong, and if you had read
only that you would have lost nine days for nothing. The cap was never a spend limit on our own
account: the billing page you land on by default belongs to a *different* Anthropic organisation
from the one the fleet's key lives on, so it read "0% used" while the API refused every call, and
the 1 September date it quoted was that other account's monthly reset. You sorted it this morning.
The fleet's own call log shows the last refusal at **10:10Z** and 134 successful calls since. So I
restarted the stalled review round, and it came back within the hour.

**Then the measurement, which is the real work of this session and which changed the design.**

Before writing a line of Phase B I built a throwaway harness that replays the platform's *actual*
matching code over a frozen snapshot of the live fleet — 829 pages, 667 button-bearing components,
1,266 buttons with words on them — and asked what the widening would do. Two things came out that
I would not have guessed.

**One: this change is much bigger than the bug.** Today the platform rewrites about 32 button
destinations across the whole fleet on a given pass. After the widening it would rewrite **291** —
about nine times as many — and roughly two thirds of that is nothing to do with bug 308. Every one
of those is a button whose words name one page while its link points at another, so most of them
are corrections. But it is a fleet-wide content change and I want you to have the number rather
than a reassuring adjective. It only happens to a page when that page is next built or rebuilt.

**Two: a third of the rewrites were decided by alphabetical order.** When two pages score equally,
the code picks whichever page's name comes first in the alphabet. That is not a small tail — it
decided **263 of 1,146** matches, and **137** of the rewrites. Worse, I could see it going wrong:
on finetuning.uk the button reading "how we work" *already points at* `/how-we-work.html`, and the
widening would have moved it to `/about.html` — because the About page's title happens to read
"Who We Are and How We Work", so both pages tie on the word "work" and "about" wins the alphabet.
Thirteen live findings tell the platform to make exactly that change. The same thing on
dartsonline.com would move "Read the guides" off the guides page.

So I tried to break the tie better — twice, with two different rules — and measured both over the
whole fleet. Both fixed some cases and broke others, for the same reason a third rule was tried
and thrown away back on 11 August: when two pages tie on one word, there is genuinely nothing there
to decide with, and any rule you invent is just a different arbitrary answer. **So the change I
made instead is that the platform now says "I don't know" and leaves the button alone.** The
repairs bug 308 is actually about — "Book a discovery call" pointing at a password tool instead of
the contact page — all survive it.

**One more thing that is worth knowing because it is counter-intuitive.** The obvious safer
version of this change is to add *only* contact/about/terms pages to the list of candidates, and
leave everything else as it is. I measured that too. It rewrites a third as much and gets **more**
of it wrong — it invents repairs like "Talk to us about your setup" → the About page, purely
because the word "about" is in the sentence and the page it *should* have found is not on the list
it was given. A short list does not make the machine careful; it makes it certain for the wrong
reason.

**Where we are now.** Phase B is written, tested, mutation-checked (I deliberately broke it seven
different ways and confirmed a test caught each one), committed, registered, and submitted to the
review council. It is inert until the next fleet build. I also wrote up the one part of it that
changes a shared piece of machinery as a formal architecture note (RFC 047) for you to rule on,
because it changes what that piece of machinery *promises* — and there is a live choice in it I
would rather you made than I did: whether the *detector* should go on guessing on a tie even
though the *repairer* now refuses to.

**Bug 308 stays open.** Nothing here has touched a real page yet. The bar is a button moving on a
site somebody can load in a browser, and that needs the next build.

**And a correction to my own caution from this morning:** I said the check that finds these broken
buttons had not run since 19 August. It ran on **22 August** and filed 40 items. The count is 188
today, not 200 — it moved down because items changed status, not because anything was fixed.

---

**2026-08-24, later — I went and looked at the old broken sites, and the backlog turned out to be
the wrong thing to be worrying about.**

You asked whether the sites broken back in July still need fixing. The short answer is no, and the
longer answer is that the queue of 215 stuck repair jobs is not a list of what is broken. It is a
list of what *was* broken, and most of it has since been overtaken by events.

Here is how I checked. Rather than trust the old records, I took the actual check the platform uses
to find these broken buttons, and re-ran it — read-only, changing nothing — over every page as it
stands today. To make sure my copy of the check behaved like the real one, I found a page the live
system had flagged two hours earlier and had not yet repaired, and confirmed my version produced
exactly the same result down to the suggested destination. It did.

**July.** Every one of the July jobs is the same site, vonc.com, and it comes to seven distinct
broken buttons on three pages. Four of them simply no longer exist — the copy was rewritten at some
point in August, and I loaded the live pages to confirm it: the About page now offers "Find Your
Archetype" and "Enter the Gauntlet", both pointing somewhere sensible. Two more are unchanged, but
they are the odd case where the button's wording names the very page it is sitting on ("Explore All
Archetypes", on the archetypes page). The platform now deliberately refuses to guess at those,
because it decided that is a wording problem rather than a link problem, and I agree with that. The
last one is a genuine coin-toss between two pages, which it also now declines. So: nothing from July
needs a repair job. There may be a couple of buttons there whose *wording* you would want changed,
but no machine is going to decide that.

**The backlog as a whole.** Of the 215 stuck jobs — which are really 325 distinct findings, because
the same page gets re-filed over and over — 124 have already been fixed by other means, 65 describe
buttons that no longer exist, and 64 are unchanged buttons the platform now declines to judge. That
leaves 65 that are still real, still live, and still fixable automatically.

**But 301 buttons are wrong across the estate right now, and only 65 of them are in that backlog.**
So releasing the backlog would fix about a fifth of the problem while re-rendering eleven sites'
worth of pages to do it. That is a poor trade.

**What I think is actually going on.** The last time the platform swept the whole fleet for these
was 17-18 August, which was before the fix went live. Two sites have been swept since the fix:
finetuning.uk and robot-hands.com. Those two now have essentially nothing left — nought and one
respectively. Every site last swept before the fix still has a pile. robot-hands.com is the clearest
picture of it: seventeen bad buttons found at ten to one this afternoon, sixteen of them repaired by
twenty past, nobody involved. The remaining one was still in the queue and two others sit in a kind
of component the repairer is not allowed to touch.

So the remedy is not to release the old jobs. It is to sweep each site again, now that the fix is
live, and let the same drain happen. The sweep itself is cheap — it is database queries, no AI
involved — and it can be done one site at a time, which is the per-site caution you wanted.

**One thing you should see regardless of what you decide.** gaswholesalers.com is serving, right
now, a button that says "Contact Our Sales Team" that takes the visitor to a fuel budget forecaster,
and another saying "Review Supply Terms" that goes to a break-even calculator. On the industries
page, "Contact our sales team" goes to the break-even calculator too. I checked those on the live
site, not in the database. That site has 25 wrong buttons and every one of them is the automatically
fixable kind. If you want one site done first, it should be that one.

**And a correction to the handoff I picked this up from:** it says releasing the backlog would touch
eleven client sites. It is seven, as of today. Counts like that go stale by addition and subtraction
both, which is why they now get written with the date they were taken.

**Same day, after you asked for a second opinion.** I had a second model re-derive all of it from
scratch — its own program, its own queries, without looking at mine — and told it to try to prove
me wrong. Everything held, with three small corrections (one button had moved rather than stayed
put; four sites have been swept since the fix, not two; the one "still queued" robot-hands button is
actually stuck behind a page-ownership guard and won't be machine-fixed). None of that changes the
picture.

What it did find is the thing I had not named, and it changes how to say this. Those 215 "stuck
jobs" were never jobs waiting to run. The platform has a two-strikes rule: if it has already tried
to repair a page twice in a week and the problem is still there, the next finding is filed straight
into a "gave up" pile so a human can see it. That is what the 215 are — labels saying "tried twice,
didn't stick" — and the reason they didn't stick is the one this whole bug is about: the repairer
could not reach the page the button named. So "releasing the backlog" would have been re-triaging
labels, not unblocking work. It was the wrong picture all along, and I'm glad we did not act on it.

Two practical consequences for the sweep. First, a page whose broken links are all in a kind of
component the repairer is not allowed to touch will "complete" without changing anything, and if
you sweep the same site three times inside a week it manufactures new "gave up" rows. So sweeps of
one site should be more than a week apart until that is fixed properly, which is the next piece of
work on this bug. Second, sweeping the same site again within three hours is silently ignored and
looks like a clean result. Neither applies to gaswholesalers.com today — every one of its 25 broken
buttons is the fixable kind, and it was last swept a week ago — which is one more reason to start
there.

One more thing to hold onto: about 500 buttons across the estate are ones the platform now
deliberately declines to judge because the wording could mean two pages equally. That is by design,
but it means "nothing left to fix" on a site does not mean every button is right.

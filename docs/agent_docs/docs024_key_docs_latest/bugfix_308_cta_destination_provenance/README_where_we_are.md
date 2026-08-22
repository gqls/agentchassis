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

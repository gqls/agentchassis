# Where we are — loancash.co.uk and the FCA rules

Plain-prose log, append-only, newest at the bottom.

## 2026-08-11 — the thing we were worried about isn't wrong

We'd been carrying a worry about loancash for a while, written into three separate
handoffs: that its payday-loan tools have the FCA's price-cap numbers typed straight
into the page, with nothing checking them — and that they might be out of date, the
way the stamp duty rule was on the other site. That one had been wrong for sixteen
months and was under-quoting people by five thousand pounds, so it was a fair worry
to have.

**I checked it properly today, against the FCA's own rulebook rather than against
our own page, and the numbers are right.** All three of them: the 0.8% a day cap on
interest and fees, the £15 ceiling on default charges, and the rule that total
charges can never exceed the amount borrowed. I've noted the exact rule numbers
alongside each one. They were last changed on the second of January 2015 and haven't
moved since — the FCA reviewed them in 2017 and left them alone.

I also checked the sums themselves, not just the constants, on both tools. They're
right. One detail worth recording because it's the sort of thing a tidy-up would
break: the price cap checker asks for "total interest and fees you paid", not "total
you repaid". That distinction is what makes its final test correct. If someone
reworded that label to something that sounded cleaner, the arithmetic underneath
would quietly become wrong by about a factor of two, and nothing would complain.

So the stamp-duty comparison doesn't really hold. Stamp duty thresholds move every
time there's a Budget — being out of date there is a matter of when, not if. This
cap hasn't moved in eleven years.

**What is still true is the second half of the worry: nothing is checking.** The
numbers are right today and we'd have no way of knowing if that changed. That's
worth fixing, but it's a monitoring job rather than an urgent repair, and it should
be built cheaply — nobody is being misquoted in the meantime.

**One thing I haven't checked, and it's now the one I'd worry about.** There's a
third tool on that site, a complaint deadline calculator. It works off time limits
for making a complaint — the six-year rule, and the six-month deadline once a firm
has given you its final response. Those come from a completely different part of the
law than the price cap, and unlike the price cap, those sorts of deadlines do get
changed. I'd move the worry there rather than drop it.

I've written down a plan, and the main thing it says is: don't overbuild this. On
the mortgage site we built a proper independent calculator to check the calculators,
because those do genuinely complicated sums where there's a real right answer to
work out separately. Here there are three tools doing a couple of multiplications
each. Writing our own copy of "amount times 0.008 times days" proves nothing — it
would agree with a wrong number just as cheerfully as a right one. What actually
earns its keep is something that reads the rulebook and shouts if it disagrees with
what's on our page.

## 2026-09-03 — the thing this lane asked for in August now exists, and it found three mistakes on day one

Back in August I wrote here that the price-cap numbers on loancash were right, but that **nothing was
checking** — and that what would earn its keep is "something that reads the rulebook and shouts if it
disagrees with what is on our page". You asked today for loancash to be populated. That is what this
is, and it closes exactly that gap.

**What it is, in plain terms.** The site now has a register: a list of nineteen statements of what
the law actually says, each one carrying a link to the official source and a sentence quoted word for
word from it. Every night the system re-reads those nineteen sources and checks the quoted sentences
are still there. If the FCA rewrites a rule, or moves a page, we hear about it. Until today there was
nothing to re-read, so the nightly check passed over this site in silence.

I also went back for the loose end I flagged in August. I'd said the complaint-deadline tool was the
thing to worry about, because those deadlines — unlike the price cap, which has not moved since 2015
— are the sort that do change. All three of those deadlines are now in the register, so that worry is
now covered rather than just recorded.

**Three things on the site are wrong, and I have not touched the pages.** Every one of the four
finance sites we have done this to has turned up errors, and this one turned up three. Recording them
is my call; changing published wording is yours.

1. **The £15 default fee.** Two pages say the lender can charge £15 "once per missed payment". The
   rule is stronger than that: £15 is the most a lender can charge in default fees **for the whole
   loan**, however many payments you miss. So the site is understating the protection — someone
   reading it who missed two payments would accept a second £15 charge as legitimate, and it isn't.
   On a site whose entire purpose is telling people what they are owed, that one matters most.

2. **The payment-attempt rule.** The loan-sharks page says a lender "cannot take more than one
   payment attempt of over £1" from your account. The real rule is that they get **two** attempts,
   and after two failures they must stop and contact you. There is no "£1" anywhere in that part of
   the rulebook. The site's own dedicated page on this gets it exactly right — so it is one page
   contradicting another, not a settled misunderstanding.

3. **A wrong rule number.** One page credits the affordability checks to "CONC 5A". That is the
   price-cap chapter; affordability is a different one. Three other pages cite it correctly.

**One thing I nearly got wrong myself, worth a line because it is the kind of error that never
announces itself.** When copying a sentence out of the rulebook I typed commas where the original has
brackets. It looked identical to a human. The automated checker rejected it — and had I not run that
check, the site would have reported a false alarm about that rule every night, for ever, and the
false alarm would have looked exactly like a real one. I now paste those sentences rather than
retyping them.

**And a near-miss on the banned-phrases list.** Our other loan sites carry a list of sales phrases
that must never appear — "guaranteed approval", "no credit check loans", that sort of thing. It turns
out the list exists in two slightly different versions, and the wider one would have flagged a
perfectly correct sentence on loancash explaining that an employer salary advance involves no credit
check. Copying "the standard list" without looking closely would have had the site telling us off for
giving good advice. I used the narrower version and wrote down why, so nobody tidies it back.

**Where that leaves the wider picture.** All five of the finance sites we set out to give registers
to now have one. The job I flagged this morning as the last outstanding item is done. The remaining
gap is the one I described earlier: a third of our live sites have no register at all, and nothing in
the system can notice a site that is missing one, because the nightly check builds its list from the
sites that already have one. That is not this site's problem any more, but it is still somebody's.

## 2026-09-03, later — your four decisions, and where each of them got to

You made four calls this afternoon. Three are done and running; one is waiting on another thread.

**"Fix the loancash wrong sentences" — dispatched.** I have not edited the pages by hand, because
that is not how this estate is meant to work and you have ruled on it twice. Instead the three
corrections are now four work items (the £15 one appears on two pages), each carrying the exact
sentence that is wrong, the rule that governs it, what must be **kept** as well as what must change,
and a test anyone can run against the live page afterwards. The framework rewrites the copy; I
supplied the instruction.

Two things about that are worth your knowing, because both would quietly mislead anyone who checks
later. First, a page re-render was already queued on all four pages twenty minutes before I filed the
repairs — and a re-render rebuilds the page from its stored content, so it will ship the *same wrong
words* and report success. Second, there is no automatic verifier for this kind of item at all, so a
work item saying "complete" is not evidence the sentence changed. Both facts are written into every
item and into the handoff. There is also a 48-hour clock: if nothing picks these up by Friday
afternoon, a housekeeping job will mark them "unresolved", which reads like they were processed rather
than ignored.

**"Build the missing check" — built, live, and I watched it work.** It went in at half past three and
fired within a minute, filing twelve items — one for each live site with no register. It runs daily.

It is deliberately not a piece of program code but a database job, for a reason worth stating: code
here does nothing until somebody rebuilds and redeploys the system, and we have a case on record of a
check sitting switched off for nine days after the thing blocking it was cleared. A database change is
live the moment it is applied. This also follows a rule you approved earlier for exactly this
situation.

One design choice I want to flag because it is the same mistake we spent this morning fixing
elsewhere: the job reports **three numbers every single time it runs** — how many sites are missing a
register, how many it filed, how many were already queued. The shorter version would have said nothing
when there was nothing to do, and then "nothing wrong" and "the job never ran" would look identical
from outside. That is precisely the trap we hit this morning with the other check. Now a missing line
means it did not run, and a zero is a positive statement of health.

**"A register for each site, but a lower bar for normal ones" — designed, and your instinct matched
something we had already agreed.** We ruled back on Tuesday that sites sit on one of three rungs
depending on whether their claims are about themselves or about the world. What you asked for changes
one line of that: the bottom rung used to mean "no register needed", and now it means "a register, but
the cheap kind".

I worked out what "cheap" should mean by reading the code rather than guessing, and it comes out
better than I expected. The anti-slop protection you are after comes from a check that flags any
number on a page that no registered fact supports — and that check **does not look at sources at
all**. It only compares the number. So an ordinary site's register can just be its own figures with
who vouched for them and when: no citations, no nightly fetching, no risk of false alarms, and it
takes hours rather than half a day. Only sites making claims about the outside world need the
expensive version with a link and a quoted sentence for every fact.

The honest limit, so nobody discovers it after building twelve of them: that numeric check is
switched off on guides, blog posts and tool pages, because their body text is instruction rather than
a claim about the business. It covers landing pages, ordinary content pages and section indexes. That
exclusion was measured, not guessed, and I would leave it alone.

**"A register for vetcomparison" — asked, no answer yet.** I put two questions to that thread: is it
theirs or nobody's, and is a register even the right tool for a comparison site, whose numbers may be
other people's claims rather than its own. That second question could change the whole approach, so I
would rather wait for the answer than guess. Its item is in the queue either way.

**What is left is a programme, not a decision.** Twelve sites need registers. That is days of work
rather than an afternoon, and every lane that has done one so far has found real mistakes in its own
site's live copy — four out of four. The queue now exists and nothing gets lost.

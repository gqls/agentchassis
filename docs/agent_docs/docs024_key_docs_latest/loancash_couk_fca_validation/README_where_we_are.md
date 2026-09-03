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

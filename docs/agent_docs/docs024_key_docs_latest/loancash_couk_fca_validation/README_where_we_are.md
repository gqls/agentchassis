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

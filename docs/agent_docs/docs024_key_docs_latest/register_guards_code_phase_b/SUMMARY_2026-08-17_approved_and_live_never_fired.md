# SUMMARY — 2026-08-17 — approved, live, and it has never fired

## What we're trying to do

Stop a calculator on one of our sites quietly computing a number that the law changed
years ago.

We keep, for each site, a register of facts: every tax band, rate and threshold, each
with a link to the government page it came from, re-checked automatically every morning.
It is good machinery and it works. But it only ever governed what a page could **say**.
It never governed what a calculator **works out**. So a site could hold the correct
stamp-duty threshold in its register, freshly verified that very morning, while the
calculator on the same page used a figure that expired sixteen months earlier — and
nothing anywhere would notice, because nothing was ever asked to compare the two.

The job was to connect them.

## Where we've come from

A bug filed in August, `225`: our stamp-duty calculator was charging first-time buyers
using a relief cap that stopped being law in March 2025, under-quoting a real tax bill by
£5,000. That was fixed on the 9th and is verified still fixed.

What outlived the fix was a section inside the bug file titled *"why no existing check
could ever have caught this"*. It was right, and nobody had acted on it. Three separate
checks are blind to this class of mistake, and — the awkward part — all three are
sensible decisions on their own terms. Our text checks ignore anything inside a program,
because code is not prose. Our number checks skip calculator pages, because a
calculator's help text is full of numbers that aren't claims. Our number checks ignore
money amounts, because otherwise every price on every site would trip them. Each is
defensible; together they leave a hole shaped exactly like this bug.

## What we've done

We found that this had already been thought through — a plan written on 9 August, which
the owner has seen, setting out four pieces. Piece one was already live. Pieces two and
three had been designed and then set aside when that team moved on to other work. So
rather than invent something alongside their plan, we built their pieces two and three.

In plain terms: **a calculator can now declare which registered facts it relies on, and
when one of those facts changes, the morning check names that calculator the same day.**

It went through our review council three times. **Both rejections found a real defect,
and the second one was the valuable one**: it showed that our fix for the first rejection
*could never actually fire* — we had fixed the symptom and left the cause sitting in the
code. A green light after one round would have shipped a check that looked fine and was
incapable of doing anything. The third round approved it, thirteen of fifteen reviewers,
and we acted on both remaining advisory notes anyway — one of which pointed out we had
named a handler that does not exist.

## Where we are now

The code is **live**, confirmed by asking the running program what it contains rather
than trusting a version number, on both machines. The morning check ran with it at
09:04 today across all thirteen sites that have a register: eight updates written, **zero
errors**. So it works and it has broken nothing.

**And it has never once fired**, because no calculator has made a declaration yet. Today's
clean run is a check with nothing to check, and we have written that into the tracking
file in those words — because "the sweep ran clean" is exactly the sentence someone would
later quote as proof the thing works.

One honest limit worth repeating: this tells us when a figure has **moved**. It cannot
tell us whether a figure is **right**. If the register and the calculator are both wrong
in the same direction, they agree and this says nothing. That is the fourth piece of the
plan, and it needs its own review before anyone builds it.

## Where we're going

One line of configuration, on one calculator, and the machinery starts protecting the
page the original bug was about. That line belongs to the team that owns
mortgagecalculator.co.uk, not to us — and they were working on that site yesterday. We
asked rather than doing it, because switching it on files items into their work queue,
and a queue we already know has no working screen for someone to read them on. The
request is now at the top of the file a fresh session on that lane reads first, with
three options and no preference from us.

After that: prove it fires by deliberately breaking something and watching it complain —
the only test that distinguishes a working check from a decorative one. Then the second
site, then the remaining half of the problem, which is the same blindness in ordinary
page text rather than in calculator code.

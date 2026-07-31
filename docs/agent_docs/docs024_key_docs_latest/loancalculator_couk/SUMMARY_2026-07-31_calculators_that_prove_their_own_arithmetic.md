# SUMMARY — 2026-07-31 — Calculators that prove their own arithmetic

*Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the technical log is in `NOTES_loancalculator_couk.md`.*

*This is the second summary in this lane. The first
(`SUMMARY_2026-07-30_adopting_a_hand_built_site.md`) was about getting a
hand-built site into the platform at all. This one is about a different and
larger thing that came out of it.*

---

## What we're trying to do

loancalculator.co.uk was built by hand, outside the platform, and adopted into
it. The owner's instruction was that it must become **"completely editable and
evolve and improve like the other sites will, just as long as it starts
similarly enough with working tools."**

The last five words are the whole problem. "Working tools" is easy to say and,
it turns out, was not something we could actually check. The site's twelve
calculator pages are its entire reason to exist — people arrive at them to find
out what a loan will cost — and if the platform starts editing those pages, we
need to know the moment a number stops being right.

So the job became two jobs. Make the calculators into things the platform can
own and improve. And build the thing that tells us when one of them breaks.

## Where we've come from

The site was adopted frozen: every page stored byte-for-byte, nothing editable,
which kept it safe and kept it useless. Unpicking that meant deciding which
parts could be handed to the platform's content loops, and the honest answer was
that we couldn't decide until we could verify.

Then we found out what our verification was actually worth. The platform has a
ladder of checks for tools. One asks whether the tool has any JavaScript at all.
One asks whether the elements a tool refers to exist on the page. One measures
whether they're big enough to see. The best of them can check that some text
matches a pattern — that a box contains something shaped like money.

**Not one of them knows what the answer should be.** We proved it rather than
argued it: we built a page with one input box, one output box and no code
whatsoever, and the existing tooling scored it as responding correctly. A
calculator that has quietly started producing wrong numbers looks exactly like
one that works.

That mattered immediately, because it meant an earlier claim of ours — "twelve
calculators, all working" — had never been established. It has been corrected
where it was written.

## What we've done

**We built the missing check, and put it in the platform rather than in a script
of our own.** It drives a calculator with fixed inputs and then insists on the
exact answer: not "something money-shaped", but £303.44. It runs on the
platform's normal schedule, in a real browser, alongside the checks that already
exist. Its central test is the interesting one — it takes a single page with a
wrong number on it, runs it past the old checks (which pass it) and then past
the new one (which fails it, naming both the wrong value and the right one). A
gate nobody has watched fail is not evidence of anything.

We deliberately did **not** give it a cheap static version. A weaker check
wearing this name would recreate the exact false confidence we'd just spent a day
proving was there.

**We rewrote every calculator on the site — eleven of them — and proved each one
computes identically to the original.** They're now proper platform components:
their labels, defaults, currency and rates are configuration a content agent can
edit, their styling no longer depends on a stylesheet that might be rewritten
underneath them, and they no longer trample each other's code.

Along the way it turned out to be eleven, not twelve: the homepage calculator and
the standard loan calculator are the *same* calculator, sharing one piece of code.

**We built the harness that made that tractable.** It takes the real page, cuts
the old widget out, drops the new one in, serves the site locally and compares.
So each rewrite was proven *before* it went anywhere near production — which
matters here, because on this platform committing is shipping.

## Where we are now

All eleven reproduce their recorded values exactly. Nothing is live yet.

**The gate earned its place several times over, on our own work.** Three findings
stand out.

A piece of ordinary text — the phrase *"58-day"*, in quotes — was dropped into
the calculator's code by the templating system and produced a syntax error that
killed the entire tool. It displayed £0.00 for every input while still passing
every structural check the platform has. We now refuse, at build time, to put
quotable text into code at all.

The same layout rule broke two tools in *opposite* directions, because their
original stylesheets happened to disagree. Nothing about reading either page
could have told us.

And our first attempt to *improve* the harness silently moved its test click onto
the site's navigation menu on nine of twelve pages. The gate had stopped testing
the calculators entirely, and only comparing two full measurements against each
other showed it.

**We also found real defects, and split them by a rule worth stating.** Fixes
that provably cannot change what a visitor sees went in — a "clear" button that
was wiping *all* browser storage rather than its own, two tools counting
checkboxes belonging to other parts of the page, a restore that failed silently.
Anything that changes what the page actually says did not, however worthwhile:
one calculator prints three decimal places on a money figure, another computes
nothing at all on a 0% deal, and a third signals its entire verdict by colour
alone. Each is queued as its own change with its own re-measurement. **The test
is whether a change is visible, not whether it is worth making** — and we held
ourselves to it, reverting an accessibility improvement we'd already written.

**The platform change went through the review council and came back
"revise" — correctly, on a point we'd missed.** Four independent reviewers landed
on the same hazard: a check the running system doesn't recognise yet is *skipped*,
and a set of skipped checks *passes*. Install one of these too early and it
reports green having tested nothing — the precise failure we built it to
eliminate, reproduced by the fix for it. We've since written the guard that
refuses to install one until the deployed system provably understands it. On its
first run that guard caught a fault in *itself*, and said so instead of lying.

## Where we're going

Three things, in order.

The components need to go into the database and the pages rebuilt from them —
that's when the site becomes editable, which was the original instruction.

The new check needs to ship and then be switched on for each calculator, in that
order, using the guard we've just written.

And the queued defect fixes: the three-decimal-place money bug, the 0% case, the
colour-only verdict.

Two things still need you. **The GitHub token can't see the repository holding
the site's source**, which blocks one part of the adoption; that needs someone
with GitHub admin. And the **open question from last week is still open** —
whether to do the full decomposition here or the cheaper freeze-the-calculators
split the neighbouring lane chose. The rewrite didn't depend on that answer, but
the next step does.

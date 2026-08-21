# SUMMARY — 2026-08-21. The pages are repaired

## What we're trying to do

Make sure that when the platform *finds* something wrong with a live page, something can actually
*fix* it. The lane started from one bug — a whole class of findings (`required_fields_missing`) that
no handler in the fleet owned, so they piled up at "needs a human" for ever — and it has kept
running into the same shape from different directions: detection is good, repair is where the
estate thins out.

## Where we've come from

We built a router that sends each finding to a handler that can actually act on it, and proved it
works for most of the class. Then the last stubborn group refused to yield, and understanding why
took several corrections of our own earlier conclusions. We first blamed page ownership; that was
wrong. The real reason was harder: for these pages the stored content **cannot reproduce the page
that is served**, so every repair route we had — all of which work by regenerating the page from its
stored content — was inapplicable by construction, not by policy. Nothing could fix them, and the
mechanism kept re-filing the finding and closing it as "won't fix", quietly.

The owner ruled on 20 August that these seven findings did deserve a route, even though the defect
is mild (backticks showing as literal characters around code words on developer-tool pages). So we
built the one shape the estate did not have: a small, mechanical edit to the finished page HTML
itself, with no language model in the loop.

## What we've done

Today the code was live and the repair ran, end to end, on the real site.

The one genuinely new thing we learned is a delay nobody had seen: these checks only run when a
site's turn comes up on a seven-day rotation, and the rotation had gone quiet because no site was
old enough to be due. The next examination of this site would have been the 25th. With the owner's
go-ahead we forced a single examination of the one site, in a way that did not use up its place in
the queue.

The check filed eight findings and every one was routed to the new repair. We promoted one by hand —
required, because the dispatcher will not trust a route until it has seen one success — and it
finished in three minutes. Fifteen minutes later the dispatcher released the remaining six on its
own and they repaired themselves. Seven pages, all verified, all confirmed on the page a visitor
actually gets, with the risky case holding: the backticks that belong to the pages' own JavaScript
were untouched, all forty-four of them on the busiest page.

We also cleared an older debt: one of the review council's seats had been giving authors advice that
was the exact reverse of what the code does, and a script that copies reviewer configuration had
mangled one of that seat's headings into nonsense. Both are fixed, and the script is anchored so it
cannot do it again.

## Where we are now

The first half of bug 277 is closed on evidence: a page that could not be repaired has been
repaired, mechanically, and the proof is the bytes the site serves. The second half is untouched and
honest about it — a different group of parked findings, needing a different agent, which nothing in
today's work brings closer.

Bug 083's fix is finished and proven; it is sitting out an agreed week of watching before we close
it, which ends around the 24th–25th. One thing that close must say plainly: a safety door we built
to hand stuck work to a human has never once handed anything back the other way. Nothing is wrong
with it — the condition simply hasn't occurred — but it should not be described as proven.

## Where we're going

Three things, in order. Read the council's verdict on today's reviewer-seat change and act on it.
Close 083 at the end of its week, saying what has and hasn't been demonstrated. Then the remaining
half of 277 — the parked findings with no route — which is the last thing keeping that file open.

There is also a small prediction to check on the 25th. One of today's eight findings was born dead:
a rule that stops us retrying a page we've already failed on twice counts failures per *page*, not
per *method*, so two failures by the methods we had just proved useless were charged against the new
one that works. It should heal itself when those old failures age out. If it doesn't, that tells us
something worth knowing, which is why we left it alone rather than nudging it.

# SUMMARY 2026-08-09 — bug 220: the dead-link repair that repaired the wrong page

*The first summary in this lane. Written at the milestone where the fix became
provable rather than merely deployed.*

## What we're trying to do

Our sites are built and maintained by the platform itself: checks notice problems,
file work items, and dispatch handlers to fix them. One of those checks notices an
internal link pointing at a page that has never been built — a link that leads
nowhere. It files a job saying "build this missing page".

We are trying to make that job actually do what it says: build the missing page,
publish it, and only then declare itself finished.

## Where we've come from

It did not do that. It did something worse than failing — it succeeded at the wrong
thing, quietly.

The check filed the job correctly, recording the missing page as the target. But the
dispatcher handed the handler the wrong page: the page *containing* the broken link,
not the page the link *pointed at*. So the handler rebuilt and republished the
containing page, reported success, and the platform marked the job complete. The
missing page stayed missing. The link stayed dead. The next scan found it again,
filed it again, and the cycle repeated — a green loop that converged on nothing.

Two things made it invisible. The job's own record read `complete` with no error, and
the target page did often appear eventually, built by some unrelated piece of work
days later — so any spot check of a single site looked fine. Only joining "when did
the job close" against "when did the page actually deploy" revealed the job had been
green while the page did not exist.

It also did real damage. Rebuilding the containing page could overwrite it with the
wrong content: on this site, the beginners guide was overwritten with grip-styles
material.

## What we've done

We fixed it in three parts, because the first two were not enough on their own.

The dispatcher now passes the target page's id. The handler now treats that id as
authoritative when it loads the page record. And — the leg we discovered we had
missed only after watching a live run — the step that saves the new content now takes
the page *name* from the loaded record rather than from the job's original request,
so the content lands on the target rather than the container.

We added a verifier, so the job can no longer close on the handler's say-so. It
closes only when it can state a reason: either the target page has shipped and the
link now resolves, or the link is no longer on the page at all. Those two reasons are
recorded distinctly, which matters more than it sounds — the weaker one is what the
first live run produced, and mistaking it for success would have closed this bug two
days early.

We repaired the page that had been overwritten, and checked the repair at the live
website rather than at the job status.

The change went through the council gate and was approved. The mechanism is
registered so other workstreams can find it.

## Where we are now

It works, and we have watched it work from end to end.

At twenty past two this afternoon the queue reached our jobs. Within eight minutes,
three of them converged. The grip-styles guide — planned but never built, the target
of every dead link on the site — was built, published, and now serves a real page at
its own address. The pages that contained the dead links were left alone, still
carrying their own content. Each job closed with the strong reason: the target has
shipped, the link resolves.

A fourth converged on a different kind of page entirely, a section listing rather
than an article, which we had written down as something the builder could not make.

So bug 220 is fixed, live, and proven. The file stays in the open-bugs directory per
the owner's ruling, with its status updated at the head.

Two corrections belong in this summary, because both are cases of us being confidently
wrong in writing.

The acceptance check this lane documented was faulty. One of its columns read blank
for every job on the system — the instructions pointed at the wrong depth in the
stored result. It survived review because the lane's own control expected that column
to be blank, and a blank meaning "nothing published" is indistinguishable from a blank
meaning "you are looking in the wrong place". The check was run exactly as prescribed
and could not have objected. Had a genuine success not landed in front of us while the
check still called it a failure, we would have reported the fix as incomplete. The
instructions are corrected and a second control added.

And our prediction that four jobs would fail loudly, because they targeted section
listings, is refuted — one of them succeeded first. That prediction had been
extrapolated from a different page type failing earlier the same day and was never
tested before being written down. It was serving as the argument for a further piece
of work we had deferred; that argument is now weaker and possibly gone.

## Where we're going

Three of the ten jobs from this run are still working through the queue. Their
outcomes are the honest evidence about whether section listings build reliably, and
nobody should pick up the deferred routing work without reading them.

There is one piece of old damage left, on a different site: a job that closed green
back on 5 August, before the verifier existed, leaving a live page linking to a
missing one. We are deliberately not hand-repairing it. A closed job frees its slot,
so the next routine scan of that site will file the finding again and it will now
converge honestly — giving us a second, independent proof for free.

Beyond that, this lane's work is done. What remains is watching the machinery do
unattended what we have now seen it do once under observation.

# SUMMARY 2026-08-11b — the detector is built, and the remaining gap is now visible weekly

*Second read-out today. The first one closed with "build a detector for the next
collision" as future work; that is what changed.*

## What we're trying to do

Stop a particular kind of silent failure: a job that reports success without doing
anything, and a checker that agrees with it because it was asked the wrong question.
Specifically, our work tickets carry a type, and each type can have a checker attached
that re-tests the fault before the ticket is allowed to close. The type is the only
thing joining the two. When a second part of the system starts filing tickets under a
type somebody else's checker owns, that checker re-tests **its own** fault, correctly
finds it absent, and waves the ticket through with the real defect untouched.

## Where we've come from

We fixed the one live instance: the second filer got its own ticket type, and a checker
can now declare which tickets it actually speaks for and refuse the rest. That is live
and proven in production. But the declaration is optional — deliberately, because
forcing it on every existing checker would be a much riskier change — and the review
council said plainly that optional means the same bug returns the next time two filers
converge, unless a person remembers. The owner's ruling was: build something that
notices, rather than leaving a query for someone to think to run.

## What we've done

That detector now runs daily against live data and is deployed. Once a day it asks one
question of every ticket type that has a checker attached: *are tickets plainly arriving
here from more than one source, while the checker has never said which of them it speaks
for?* If so it files a single ticket saying exactly that, with no handler attached,
because the fix is a person writing four lines of code and not a robot retrying
something. When the situation is put right, the same job closes its own ticket.

Two things about it are worth saying out loud.

**It shows its working when it finds nothing.** A daily check that only ever says "all
clear" is indistinguishable from one that has quietly stopped looking, and this estate
has been bitten by exactly that before. So every run names the case it *did* see and
chose not to file — today it says, in effect, "this type still has two sources and it is
fine, because its checker now declares what it grades" — and it has a switch that
re-runs the same census with that reassurance turned off, which reproduces the original
bug as a live finding. The zero is a zero that looked.

**Telling two sources apart is harder than it sounds, and every obvious way is wrong.**
Who created the row, which pipeline it came from, which check it names, or simply "the
data looks different" — I measured all four against the live database and every one of
them fires on ticket types that have only ever had a single source. What survives is a
fuzzy comparison of the data's shape: two shapes sharing more than half their fields are
one source with a revision; two sharing almost none are two sources. That sounds loose,
and it is pinned by real numbers — across our whole fleet, same-source pairs share at
least two-thirds of their fields, and the one genuine two-source pair shares nothing at
all.

The review council sent the first version back, which was the right call and cost about
ten minutes. Two of its objections were things I had built but had not listed, so it
could not see them; two made me *check* rather than assume (a status I was relying on
behaves as I thought — but only in combination with another field, which is worth
knowing); and one was simply correct about the code, so the code changed. It approved
the second round.

## Where we are now

The instance is fixed and live. The class now has a watchman, deployed, running daily,
and proven end to end — including that the thing running in the cluster is built from
the exact commit we reviewed, checked link by link rather than assumed.

The honest limit: **it has never yet filed a real finding**, because nothing currently
qualifies. Deployed and exercised are different claims and we should not blur them.

And one thing found while measuring something else, which matters more than the
detector. The faults this whole investigation was about are **already being re-found by
the system** — fourteen of them today alone — and thirteen of those have already closed
again **unchecked**, because the new ticket type still has no checker of its own. We
expected that to be a theoretical gap for a while. It is happening weekly, in the open,
about thirteen at a time.

## Where we're going

**Build the checker for the new ticket type.** It was already the biggest of the three
decisions; it now has a weekly bleed to point at rather than an argument. The hard part
is not the code — it is that checking these particular faults means asking the closing
step to look at a real rendered page, which we have deliberately kept it from doing.
That deserves its own review round rather than being smuggled in.

**A tidy-up we have deferred on the record.** Counting properly for this piece of work
turned up **nine** scheduled self-check jobs across the platform, each with its own copy
of the same plumbing, and nobody — including me — could name the set from memory. Three
council seats have now asked for a consolidation pass twice. It is written up as a
proposal rather than done, because doing it inside this piece of work would have been
the fourth thing at once.

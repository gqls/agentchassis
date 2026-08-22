# Summary — 2026-08-22 — retired pages were being filed as work, and eighteen checks can do it

Written to be read aloud.

## What we're trying to do

Stop the platform generating work about pages it has already retired. The immediate goal was
narrow — answer a loose end left behind in bug 298 — but the thing underneath it is a rule the
codebase already states and mostly does not follow: a page has two independent facts attached
to it, *has it ever been published* and *do we still want it live*, and any query about pages
has to decide which of the two it means.

## Where we've come from

Bug 298 was about the internal linker choosing link targets from only fifteen candidate pages,
alphabetically. That was fixed on 19 August and we confirmed it is genuinely fixed — the cap is
gone from the live configuration, the migration is recorded, and the linker has since produced
its first real link plan.

But 298 also recorded something it deliberately did not chase, and marked twice as unclaimed: a
large minority of the linker's completed jobs had finished having found no page to work on at
all. Nobody had picked it up. That is what this lane took.

## What we've done

**Found the cause, and it was not in the linker.** Every one of the seventeen empty runs was
pointed at a page we had already retired. The check that finds unreachable pages asks only the
publishing question, so it keeps finding retired pages, declaring them unreachable — which they
are, because we retired them — and filing work asking a handler to link to them.

**Established that this is one producer's defect and not three handlers'.** The work goes to
three different handlers, and all three refuse it. We read all three and each already checks the
lifecycle question properly. The producer is the only party in the system that disagrees.

**Measured the damage, in a way that could have come out the other way.** Thirty-four archived
pages are being filed right now across the three branches, and the same pages have been
re-detected every rotation since April. The expensive part is second-order: these impossible
jobs burn the two-strike retry allowance, and when a batch gets parked it parks everything
queued at the time. Fifteen of the twenty currently-parked linker jobs name perfectly live
pages — real work retired as collateral.

**Sized the class by reading all seventy-one checks.** Eighteen of them can route a retired page
at a handler. It is not eighteen separate mistakes: only three used the shared helper, everyone
else hand-wrote it four different ways, and **two of those four ways exclude nothing at all** —
one filters out pages marked "deleted" and we have no such status. That one is on the
highest-priority check of the eighteen.

**Fixed the measured case and closed the class.** Two lines on the orphan check. For the class,
every check that queries pages must now declare which stance it takes and why, enforced by the
build. We deliberately did not batch-add the filter to the other seventeen — partly because each
needs its own judgement, but mainly because **a retired page can still be publicly visible**,
and another lane left a standing warning that hiding archived pages from audits is the opposite
error. One of our checks depends on being allowed to see them.

**Put it through the council: approved first round**, fifteen reviewers. Two advisory objections
named real gaps and both are fixed — a missing deploy-verification step, and an existing helper
we had not searched for (which turned out to carry the very bug we had just fixed, so finding it
was worth more than reusing it would have been).

## Where we are now

The fix and the guard are committed and approved. The guard is a test, so it protects the tree
immediately; the two-line fix is Go and takes effect at the next chassis build. The bug stays
open, deliberately: the bar is fixed *and* live, and seventeen gaps remain declared but unfixed.

Two things we are not claiming. The independent diagnosis loop **did not ratify this** — one run
died on an infrastructure fault, the second returned "unverifiable, could not narrow scope". It
independently reached two of our central citations but named three gaps it could not close; we
closed all three by hand and have said so plainly rather than letting the reader assume a
confirmation. And we found no detector anywhere for "this page is retired but still serving to
the public" — a real hole that this change neither creates nor closes.

The most useful thing we produced may be the record of getting the safeguard wrong. We shipped
two versions that passed cleanly and asserted nothing: the first was satisfied by the comment
sitting next to the fix, the second by an unrelated filter on a different table in the same
file — which is the *same* mistake that made the original bug invisible, reproduced inside the
tool written to prevent it. Only deliberately deleting the fix and checking that the safeguard
complained told the three versions apart.

## Where we're going

The seventeen remaining gaps are the larger half of the ticket and each is a small, independent
piece of work: add the arm, move the entry, and the build starts enforcing it. The registry
names them, and the test prints the running count, so the backlog shrinks visibly rather than
being forgotten.

Beyond that, two things are recorded for whoever wants them: the missing archived-and-serving
detector, and a latent copy of the comment-stripper bug in a neighbouring lane's test — reported
rather than fixed uninvited, because it scans one file that happens not to trigger it.
